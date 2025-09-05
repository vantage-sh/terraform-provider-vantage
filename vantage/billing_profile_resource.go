package vantage

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_billing_profile"
	billingprofilesv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/billing_profiles"
)

var (
	_ resource.Resource                = (*BillingProfileResource)(nil)
	_ resource.ResourceWithConfigure   = (*BillingProfileResource)(nil)
	_ resource.ResourceWithImportState = (*BillingProfileResource)(nil)
)

type BillingProfileResource struct {
	client *Client
}

func NewBillingProfileResource() resource.Resource {
	return &BillingProfileResource{}
}

func (r *BillingProfileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_billing_profile"
}

func (r BillingProfileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_billing_profile.BillingProfileResourceSchema(ctx)
	attrs := s.GetAttributes()

	// Override the token attribute to add a PlanModifier
	s.Attributes["token"] = schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: attrs["token"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	resp.Schema = s
}

func (r BillingProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *billingProfileModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := data.toCreate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := billingprofilesv2.NewCreateBillingProfileParams().WithCreateBillingProfile(body)
	out, err := r.client.V2.BillingProfiles.CreateBillingProfile(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*billingprofilesv2.CreateBillingProfileBadRequest); ok {
			handleBadRequest("Create Billing Profile Resource", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Create Billing Profile Resource", &resp.Diagnostics, err)
		return
	}

	// Debug: Log the API response to understand what's being returned
	if responseBytes, err := json.Marshal(out.Payload); err == nil {
		tflog.Debug(ctx, "API Response", map[string]interface{}{
			"response": string(responseBytes),
		})
	}

	// Apply the API response - the API returns all the data correctly
	if diag := data.applyPayload(ctx, out.Payload); diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r BillingProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *billingProfileModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := billingprofilesv2.NewGetBillingProfileParams().WithBillingProfileToken(state.Token.ValueString())
	out, err := r.client.V2.BillingProfiles.GetBillingProfile(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*billingprofilesv2.GetBillingProfileNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Get Billing Profile Resource", &resp.Diagnostics, err)
		return
	}

	diag := state.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r BillingProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r BillingProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *billingProfileModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := data.toUpdate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := billingprofilesv2.NewUpdateBillingProfileParams().
		WithBillingProfileToken(data.Token.ValueString()).
		WithUpdateBillingProfile(body)

	out, err := r.client.V2.BillingProfiles.UpdateBillingProfile(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*billingprofilesv2.UpdateBillingProfileBadRequest); ok {
			handleBadRequest("Update Billing Profile Resource", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Update Billing Profile Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r BillingProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *billingProfileModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := billingprofilesv2.NewDeleteBillingProfileParams()
	params.SetBillingProfileToken(state.Token.ValueString())
	_, err := r.client.V2.BillingProfiles.DeleteBillingProfile(params, r.client.Auth)
	if err != nil {
		handleError("Delete Billing Profile Resource", &resp.Diagnostics, err)
	}
}

// Configure adds the provider configured client to the data source.
func (r *BillingProfileResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}
