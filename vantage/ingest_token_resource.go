package vantage

import (
	"context"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	ingesttokensv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/ingest_tokens"
)

var _ resource.ResourceWithModifyPlan = (*ingestTokenResource)(nil)

type ingestTokenResource struct{ client *Client }
type ingestTokenModel struct {
	ID             types.String `tfsdk:"id"`
	Token          types.String `tfsdk:"token"`
	FeedToken      types.String `tfsdk:"feed_token"`
	Name           types.String `tfsdk:"name"`
	ExpiresAt      types.String `tfsdk:"expires_at"`
	RetentionDays  types.Int64  `tfsdk:"retention_days"`
	PlaintextToken types.String `tfsdk:"plaintext_token"`
	Provider       types.String `tfsdk:"telemetry_provider"`
	Last4          types.String `tfsdk:"last4"`
	IssuedAt       types.String `tfsdk:"issued_at"`
	CreatedBy      types.String `tfsdk:"created_by"`
}

func NewIngestTokenResource() resource.Resource { return &ingestTokenResource{} }
func (r *ingestTokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ingest_token"
}
func (r *ingestTokenResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		r.client = req.ProviderData.(*Client)
	}
}
func (r *ingestTokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{MarkdownDescription: "Manages an ingest token. The secret is returned only at creation and is stored in Terraform state. Destroying the resource revokes the token. Changes to configuration replace it.", Attributes: map[string]schema.Attribute{
		"feed_token":      schema.StringAttribute{Required: true, MarkdownDescription: "Public token of the telemetry feed.", PlanModifiers: replace},
		"name":            schema.StringAttribute{Optional: true, MarkdownDescription: "Non-empty name with no leading or trailing whitespace.", PlanModifiers: replace, Validators: []validator.String{stringvalidator.LengthAtMost(100)}},
		"expires_at":      schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Expiration in RFC3339 format with whole-second precision. Must be in the future and within 18 calendar months of creation. Defaults to 18 months after creation.", PlanModifiers: replace},
		"retention_days":  schema.Int64Attribute{Optional: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()}, Validators: []validator.Int64{int64validator.Between(1, 2147483647)}},
		"plaintext_token": schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Secret credential, returned only at creation and stored in Terraform state."},
		"token":           schema.StringAttribute{Computed: true}, "id": schema.StringAttribute{Computed: true},
		"telemetry_provider": schema.StringAttribute{Computed: true}, "last4": schema.StringAttribute{Computed: true},
		"issued_at": schema.StringAttribute{Computed: true}, "created_by": schema.StringAttribute{Computed: true},
	}}
}
func validateIngestTokenInputs(data ingestTokenModel, diagnostics *diag.Diagnostics) {
	if !data.Name.IsNull() && !data.Name.IsUnknown() && (data.Name.ValueString() == "" || strings.TrimSpace(strings.Trim(data.Name.ValueString(), "\x00")) != data.Name.ValueString()) {
		diagnostics.AddError("Invalid name", "name must be non-empty and have no leading or trailing whitespace.")
	}
	if !data.ExpiresAt.IsNull() && !data.ExpiresAt.IsUnknown() {
		t, err := time.Parse(time.RFC3339, data.ExpiresAt.ValueString())
		if err != nil {
			diagnostics.AddError("Invalid expiration", "expires_at must be an RFC3339 timestamp: "+err.Error())
		} else if t.Nanosecond() != 0 {
			diagnostics.AddError("Invalid expiration", "expires_at must have whole-second precision.")
		}
	}
}

func (r *ingestTokenResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ingestTokenModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if !resp.Diagnostics.HasError() {
		validateIngestTokenInputs(data, &resp.Diagnostics)
	}
}

func ingestTokenMaxExpiry(issued time.Time) time.Time {
	first := time.Date(issued.Year(), issued.Month()+18, 1, issued.Hour(), issued.Minute(), issued.Second(), 0, issued.Location())
	lastDay := first.AddDate(0, 1, -1).Day()
	day := issued.Day()
	if day > lastDay {
		day = lastDay
	}
	return first.AddDate(0, 0, day-1)
}

func (r *ingestTokenResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	var plan ingestTokenModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || plan.ExpiresAt.IsNull() || plan.ExpiresAt.IsUnknown() {
		return
	}
	if !req.State.Raw.IsNull() {
		var state ingestTokenModel
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		changed := (!plan.FeedToken.IsUnknown() && !plan.FeedToken.Equal(state.FeedToken)) ||
			(!plan.Name.IsUnknown() && !plan.Name.Equal(state.Name)) ||
			(!plan.ExpiresAt.IsUnknown() && !plan.ExpiresAt.Equal(state.ExpiresAt)) ||
			(!plan.RetentionDays.IsUnknown() && !plan.RetentionDays.Equal(state.RetentionDays))
		if !changed {
			return
		}
		var configuredExpiry types.String
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("expires_at"), &configuredExpiry)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if configuredExpiry.IsNull() {
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("expires_at"), types.StringUnknown())...)
			return
		}
		if configuredExpiry.IsUnknown() {
			return
		}
	}
	expiry, err := time.Parse(time.RFC3339, plan.ExpiresAt.ValueString())
	if err != nil {
		return
	} // ValidateConfig reports invalid syntax.
	now := time.Now().UTC().Truncate(time.Second)
	if !expiry.After(now) || expiry.After(ingestTokenMaxExpiry(now)) {
		resp.Diagnostics.AddError("Invalid expiration", "expires_at must be in the future and no more than 18 calendar months after creation.")
	}
}

func (r *ingestTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ingestTokenModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	validateIngestTokenInputs(data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	body := &modelsv2.CreateIngestToken{FeedToken: data.FeedToken.ValueStringPointer(), Name: data.Name.ValueString(), RetentionDays: int32(data.RetentionDays.ValueInt64())}
	if !data.ExpiresAt.IsNull() && !data.ExpiresAt.IsUnknown() {
		t, err := time.Parse(time.RFC3339, data.ExpiresAt.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid expiration", "expires_at must be an RFC3339 timestamp: "+err.Error())
			return
		}
		body.ExpiresAt = strfmt.DateTime(t)
	}
	out, err := r.client.V2.IngestTokens.CreateIngestToken(ingesttokensv2.NewCreateIngestTokenParamsWithContext(ctx).WithCreateIngestToken(body), r.client.Auth)
	if err != nil {
		handleError("Create Ingest Token", &resp.Diagnostics, err)
		return
	}
	p := out.Payload
	data.ID = types.StringValue(p.Token)
	data.Token = types.StringValue(p.Token)
	data.PlaintextToken = types.StringValue(p.PlaintextToken)
	data.Provider = types.StringValue(p.Provider)
	data.Last4 = types.StringValue(p.Last4)
	data.IssuedAt = types.StringValue(p.IssuedAt)
	data.CreatedBy = types.StringPointerValue(p.CreatedBy)
	if data.ExpiresAt.IsUnknown() || data.ExpiresAt.IsNull() {
		data.ExpiresAt = types.StringValue(p.ExpiresAt)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
func (r *ingestTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ingestTokenModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	for page := int32(1); ; page++ {
		params := ingesttokensv2.NewGetIngestTokensParamsWithContext(ctx).WithFeedToken(data.FeedToken.ValueString()).WithPage(&page)
		out, err := r.client.V2.IngestTokens.GetIngestTokens(params, r.client.Auth)
		if err != nil {
			if _, ok := err.(*ingesttokensv2.GetIngestTokensNotFound); ok {
				resp.State.RemoveResource(ctx)
				return
			}
			handleError("Get Ingest Tokens", &resp.Diagnostics, err)
			return
		}
		for _, p := range out.Payload.Tokens {
			if p != nil && p.Token == data.Token.ValueString() {
				if p.RevokedAt != nil {
					resp.State.RemoveResource(ctx)
					return
				}
				data.Provider = types.StringValue(p.Provider)
				data.Last4 = types.StringValue(p.Last4)
				data.IssuedAt = types.StringValue(p.IssuedAt)
				data.CreatedBy = types.StringPointerValue(p.CreatedBy)
				data.Name = types.StringPointerValue(p.Name)
				if p.RetentionDays == nil {
					data.RetentionDays = types.Int64Null()
				} else {
					data.RetentionDays = types.Int64Value(int64(*p.RetentionDays))
				}
				existingExpiry, existingErr := time.Parse(time.RFC3339, data.ExpiresAt.ValueString())
				apiExpiry, apiErr := time.Parse(time.RFC3339, p.ExpiresAt)
				if existingErr != nil || apiErr != nil || !existingExpiry.Equal(apiExpiry) {
					data.ExpiresAt = types.StringValue(p.ExpiresAt)
				}
				resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
				return
			}
		}
		if out.Payload.Links == nil || out.Payload.Links.Next == nil || *out.Payload.Links.Next == "" {
			break
		}
	}
	resp.State.RemoveResource(ctx)
}
func (r *ingestTokenResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update unavailable", "Ingest tokens cannot be updated; changes require replacement.")
}
func (r *ingestTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ingestTokenModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.V2.IngestTokens.RevokeIngestToken(ingesttokensv2.NewRevokeIngestTokenParamsWithContext(ctx).WithTokenID(data.Token.ValueString()), r.client.Auth)
	if err != nil {
		if _, ok := err.(*ingesttokensv2.RevokeIngestTokenNotFound); !ok {
			handleError("Revoke Ingest Token", &resp.Diagnostics, err)
		}
	}
}
