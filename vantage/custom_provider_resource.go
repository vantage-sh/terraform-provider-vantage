package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	integrationsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/integrations"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/planmodifiers"
)

var (
	_ resource.Resource                = (*CustomProviderResource)(nil)
	_ resource.ResourceWithConfigure   = (*CustomProviderResource)(nil)
	_ resource.ResourceWithImportState = (*CustomProviderResource)(nil)
)

type CustomProviderResource struct{ client *Client }

func NewCustomProviderResource() resource.Resource { return &CustomProviderResource{} }

type CustomProviderResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Token       types.String `tfsdk:"token"`
	Id          types.String `tfsdk:"id"`
	Status      types.String `tfsdk:"status"`
}

func (r *CustomProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *CustomProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_provider"
}

func (r *CustomProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				// Uncomment when update is available in the API.
				// PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				// Remove when API supports updating.
				PlanModifiers:       []planmodifier.String{planmodifiers.ImmutableAfterCreate("name")},
				MarkdownDescription: "The display name for the custom provider. Cannot be changed after creation.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				// Uncomment when update is available in the API.
				// PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				// Remove when API supports updating.
				PlanModifiers:       []planmodifier.String{planmodifiers.ImmutableAfterCreate("description")},
				MarkdownDescription: "A description for the custom provider. Cannot be changed after creation.",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Unique token of the custom provider integration.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Same as token.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "The status of the integration. This is system-managed and cannot be configured.",
			},
		},
		MarkdownDescription: "Manages a Custom Provider integration.",
	}
}

func (r *CustomProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r *CustomProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CustomProviderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := &modelsv2.CreateCustomProviderIntegration{
		Name: data.Name.ValueStringPointer(),
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		payload.Description = data.Description.ValueString()
	}

	params := integrationsv2.NewCreateCustomProviderIntegrationParams()
	params.WithCreateCustomProviderIntegration(payload)

	out, err := r.client.V2.Integrations.CreateCustomProviderIntegration(params, r.client.Auth)
	if err != nil {
		handleError("Create Custom Provider", &resp.Diagnostics, err)
		return
	}

	data.Token = types.StringValue(out.Payload.Token)
	data.Id = types.StringValue(out.Payload.Token)
	data.Status = types.StringValue(out.Payload.Status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CustomProviderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := integrationsv2.NewGetIntegrationParams()
	params.SetIntegrationToken(state.Token.ValueString())

	out, err := r.client.V2.Integrations.GetIntegration(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*integrationsv2.GetIntegrationNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}
		handleError("Read Custom Provider", &resp.Diagnostics, err)
		return
	}

	state.Token = types.StringValue(out.Payload.Token)
	state.Id = types.StringValue(out.Payload.Token)
	state.Status = types.StringValue(out.Payload.Status)
	// name and description are not returned by the generic GET endpoint in a
	// reliable way, so preserve whatever is already in state. On a fresh import
	// (state.Name is empty/unknown) we seed the value from AccountIdentifier,
	// which the API populates with the user-supplied name on creation.
	if state.Name.IsNull() || state.Name.IsUnknown() || state.Name.ValueString() == "" {
		if out.Payload.AccountIdentifier != nil {
			state.Name = types.StringValue(*out.Payload.AccountIdentifier)
		}
	}
	// description is never returned by the API; always preserve state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

//
// To update name/description, the UI POST to /settings/custom_providers/${CUSTOM_PROVIDER_TOKEN}/update_details
//
// _method=patch
// integrations_custom_provider_access[name]=${PROVIDER_NAME}
// integrations_custom_provider_access[description]=${PROVIDER_DESCRIPTION}
// commit=Update+Details
//
func (r *CustomProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// name and description are guarded by ImmutableAfterCreate plan modifiers,
	// so they are always reverted to their state values before this method runs.
	// Carry all fields forward from the plan (which already holds state values
	// for name/description) without making an API call.
	var plan CustomProviderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state CustomProviderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Token = state.Token
	plan.Id = state.Id
	plan.Status = state.Status
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CustomProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CustomProviderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := integrationsv2.NewDeleteIntegrationParams()
	params.SetIntegrationToken(state.Token.ValueString())

	_, err := r.client.V2.Integrations.DeleteIntegration(params, r.client.Auth)
	if err != nil {
		handleError("Delete Custom Provider", &resp.Diagnostics, err)
	}
}
