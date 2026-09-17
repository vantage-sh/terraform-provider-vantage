package vantage

import (
	"context"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	accesspoliciesv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/access_policies"
)

var (
	_ resource.Resource                = (*AccessPolicyResource)(nil)
	_ resource.ResourceWithConfigure   = (*AccessPolicyResource)(nil)
	_ resource.ResourceWithImportState = (*AccessPolicyResource)(nil)
)

var accessPolicyPolicyAttrTypes = map[string]attr.Type{
	"api_version": types.StringType,
	"filter":      types.StringType,
}

type AccessPolicyResource struct {
	client *Client
}

func NewAccessPolicyResource() resource.Resource {
	return &AccessPolicyResource{}
}

type accessPolicyModel struct {
	Id          types.String `tfsdk:"id"`
	Token       types.String `tfsdk:"token"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Policy      types.Object `tfsdk:"policy"`
	TeamTokens  types.List   `tfsdk:"team_tokens"`
}

type accessPolicyPolicyModel struct {
	APIVersion types.String `tfsdk:"api_version"`
	Filter     types.String `tfsdk:"filter"`
}

func (r *AccessPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_policy"
}

func (r AccessPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Access Policy.",
		Attributes: map[string]schema.Attribute{
			"title": schema.StringAttribute{
				MarkdownDescription: "Title of the Access Policy.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the Access Policy.",
				Optional:            true,
			},
			"policy": schema.SingleNestedAttribute{
				MarkdownDescription: "Access Policy definition using vantage.* VQL.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"api_version": schema.StringAttribute{
						MarkdownDescription: "Access Policy document version. Currently only `v1` is supported.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf(modelsv2.AccessPolicyDocumentAPIVersionV1),
						},
					},
					"filter": schema.StringAttribute{
						MarkdownDescription: "Vantage Query Language (VQL) that controls which costs this Access Policy allows.",
						Required:            true,
					},
				},
			},
			"team_tokens": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Tokens for Teams this Access Policy is assigned to.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"token": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique Access Policy identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Same as token, for Terraform import compatibility.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r AccessPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *accessPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := data.toCreate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plannedPolicy := data.Policy

	params := accesspoliciesv2.NewCreateAccessPolicyParams().WithCreateAccessPolicy(body)
	out, err := r.client.V2.AccessPolicies.CreateAccessPolicy(params, r.client.Auth)
	if err != nil {
		handleError("Create Access Policy Resource", &resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(data.applyPayload(ctx, out.Payload)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// API normalizes VQL (e.g. strips redundant outer parentheses). Keep the
	// configured policy so apply matches the plan, same as saved_filter.
	data.Policy = plannedPolicy

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r AccessPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *accessPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, found, err := findAccessPolicy(r.client, state.Token.ValueString())
	if err != nil {
		handleError("Read Access Policy Resource", &resp.Diagnostics, err)
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	// Preserve configured policy across refresh so API VQL normalization
	// (e.g. stripping outer parentheses) does not cause perpetual drift.
	// On import, policy is null and we take the API value instead.
	existingPolicy := state.Policy

	resp.Diagnostics.Append(state.applyPayload(ctx, payload)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !existingPolicy.IsNull() && !existingPolicy.IsUnknown() {
		state.Policy = existingPolicy
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r AccessPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r AccessPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *accessPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := data.toUpdate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	plannedPolicy := data.Policy

	// Generated UpdateAccessPolicy uses `team_tokens,omitempty`, which drops empty
	// slices. Bypass that with a body that always serializes team_tokens so
	// `team_tokens = []` can clear assignments.
	params := accesspoliciesv2.NewUpdateAccessPolicyParams().
		WithAccessPolicyToken(data.Token.ValueString())
	out, err := r.client.V2.AccessPolicies.UpdateAccessPolicy(
		params,
		r.client.Auth,
		func(op *runtime.ClientOperation) {
			op.Params = &accessPolicyUpdateParams{
				token: data.Token.ValueString(),
				body:  *body,
			}
		},
	)
	if err != nil {
		handleError("Update Access Policy Resource", &resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(data.applyPayload(ctx, out.Payload)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Policy = plannedPolicy

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r AccessPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *accessPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := accesspoliciesv2.NewDeleteAccessPolicyParams().WithAccessPolicyToken(state.Token.ValueString())
	_, err := r.client.V2.AccessPolicies.DeleteAccessPolicy(params, r.client.Auth)
	if err != nil {
		handleError("Delete Access Policy Resource", &resp.Diagnostics, err)
	}
}

func (r *AccessPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}

func (m *accessPolicyModel) applyPayload(ctx context.Context, payload *modelsv2.AccessPolicy) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Token = types.StringValue(payload.Token)
	m.Id = types.StringValue(payload.Token)
	m.Title = types.StringValue(payload.Title)
	// Treat blank API descriptions as null so omitting description in config converges.
	if payload.Description == nil || *payload.Description == "" {
		m.Description = types.StringNull()
	} else {
		m.Description = types.StringValue(*payload.Description)
	}

	policy, d := accessPolicyPolicyFromPayload(payload.Policy)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	m.Policy = policy

	teamTokens, d := types.ListValueFrom(ctx, types.StringType, payload.TeamTokens)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	m.TeamTokens = teamTokens

	return diags
}

func (m *accessPolicyModel) toCreate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.CreateAccessPolicy {
	policy := m.policyForCreate(ctx, diags)
	if diags.HasError() {
		return nil
	}

	body := &modelsv2.CreateAccessPolicy{
		Title:      m.Title.ValueStringPointer(),
		Policy:     policy,
		TeamTokens: m.teamTokens(ctx, diags),
	}
	if diags.HasError() {
		return nil
	}

	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		body.Description = m.Description.ValueStringPointer()
	}

	return body
}

// accessPolicyUpdateBody mirrors modelsv2.UpdateAccessPolicy but keeps
// team_tokens present when empty so clearing assignments reaches the API.
type accessPolicyUpdateBody struct {
	Description *string                            `json:"description,omitempty"`
	Policy      *modelsv2.UpdateAccessPolicyPolicy `json:"policy,omitempty"`
	TeamTokens  []string                           `json:"team_tokens"`
	Title       string                             `json:"title,omitempty"`
}

type accessPolicyUpdateParams struct {
	token string
	body  accessPolicyUpdateBody
}

func (p *accessPolicyUpdateParams) WriteToRequest(r runtime.ClientRequest, _ strfmt.Registry) error {
	if err := r.SetPathParam("access_policy_token", p.token); err != nil {
		return err
	}
	return r.SetBodyParam(p.body)
}

func (m *accessPolicyModel) toUpdate(ctx context.Context, diags *diag.Diagnostics) *accessPolicyUpdateBody {
	policy := m.policyForUpdate(ctx, diags)
	if diags.HasError() {
		return nil
	}

	body := &accessPolicyUpdateBody{
		Title:      m.Title.ValueString(),
		Policy:     policy,
		TeamTokens: m.teamTokens(ctx, diags),
	}
	if diags.HasError() {
		return nil
	}

	// Always send description on update so removing it from config clears the
	// API value. The generated client omits nil (`omitempty`), so use "" instead
	// of a nil pointer when the attribute is null.
	if !m.Description.IsUnknown() {
		desc := m.Description.ValueString()
		body.Description = &desc
	}

	return body
}

func (m *accessPolicyModel) teamTokens(ctx context.Context, diags *diag.Diagnostics) []string {
	if m.TeamTokens.IsNull() || m.TeamTokens.IsUnknown() {
		return []string{}
	}

	tokens := []string{}
	diags.Append(m.TeamTokens.ElementsAs(ctx, &tokens, false)...)
	if diags.HasError() {
		return nil
	}
	return tokens
}

func (m *accessPolicyModel) policyValues(ctx context.Context, diags *diag.Diagnostics) (string, string) {
	var policy accessPolicyPolicyModel
	diags.Append(m.Policy.As(ctx, &policy, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return "", ""
	}
	return policy.APIVersion.ValueString(), policy.Filter.ValueString()
}

func (m *accessPolicyModel) policyForCreate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.CreateAccessPolicyPolicy {
	apiVersion, filter := m.policyValues(ctx, diags)
	if diags.HasError() {
		return nil
	}

	return &modelsv2.CreateAccessPolicyPolicy{
		APIVersion: &apiVersion,
		Policy: &modelsv2.CreateAccessPolicyPolicyPolicy{
			Filter: &filter,
		},
	}
}

func (m *accessPolicyModel) policyForUpdate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.UpdateAccessPolicyPolicy {
	apiVersion, filter := m.policyValues(ctx, diags)
	if diags.HasError() {
		return nil
	}

	return &modelsv2.UpdateAccessPolicyPolicy{
		APIVersion: &apiVersion,
		Policy: &modelsv2.UpdateAccessPolicyPolicyPolicy{
			Filter: &filter,
		},
	}
}

func accessPolicyPolicyFromPayload(policy *modelsv2.AccessPolicyDocument) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if policy == nil || policy.Policy == nil {
		diags.AddError("Invalid Access Policy", "API response did not include a policy document.")
		return types.ObjectNull(accessPolicyPolicyAttrTypes), diags
	}

	obj, d := types.ObjectValue(accessPolicyPolicyAttrTypes, map[string]attr.Value{
		"api_version": types.StringValue(policy.APIVersion),
		"filter":      types.StringValue(policy.Policy.Filter),
	})
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(accessPolicyPolicyAttrTypes), diags
	}

	return obj, diags
}

// findAccessPolicy looks up a policy by token via the list endpoint because the
// public API does not currently expose GET /access_policies/{token}.
func findAccessPolicy(client *Client, token string) (*modelsv2.AccessPolicy, bool, error) {
	page := int32(1)
	limit := int32(1000)

	for {
		params := accesspoliciesv2.NewGetAccessPoliciesParams().WithPage(&page).WithLimit(&limit)
		out, err := client.V2.AccessPolicies.GetAccessPolicies(params, client.Auth)
		if err != nil {
			return nil, false, err
		}

		for _, policy := range out.Payload.AccessPolicies {
			if policy != nil && policy.Token == token {
				return policy, true, nil
			}
		}

		if out.Payload.Links == nil || out.Payload.Links.Next == nil || *out.Payload.Links.Next == "" {
			return nil, false, nil
		}

		page++
	}
}
