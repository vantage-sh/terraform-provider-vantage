package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	integrationsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/integrations"
)

type AzureProviderResource struct{ client *Client }

func NewAzureProviderResource() resource.Resource { return &AzureProviderResource{} }

type AzureProviderResourceModel struct {
	Tenant types.String `tfsdk:"tenant"`
	AppID  types.String `tfsdk:"app_id"`
	// Password is write-only; not returned by the API.
	Password types.String `tfsdk:"password"`
	Token    types.String `tfsdk:"token"`
	Id       types.String `tfsdk:"id"`
	Status   types.String `tfsdk:"status"`
}

func (r *AzureProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *AzureProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_azure_provider"
}

func (r *AzureProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"tenant": schema.StringAttribute{
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "Azure AD Tenant ID.",
			},
			"app_id": schema.StringAttribute{
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "Service Principal Application ID.",
			},
			"password": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "Service Principal Password.",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Unique token of the Azure integration.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Same as token.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the integration.",
			},
		},
		MarkdownDescription: "Manages an Azure Account Integration.",
	}
}

func (r *AzureProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AzureProviderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := integrationsv2.NewCreateAzureIntegrationParams()
	params.WithCreateAzureIntegration(&modelsv2.CreateAzureIntegration{
		Tenant:   data.Tenant.ValueStringPointer(),
		AppID:    data.AppID.ValueStringPointer(),
		Password: data.Password.ValueStringPointer(),
	})

	out, err := r.client.V2.Integrations.CreateAzureIntegration(params, r.client.Auth)
	if err != nil {
		handleError("Create Azure Integration", &resp.Diagnostics, err)
		return
	}

	data.Token = types.StringValue(out.Payload.Token)
	data.Id = types.StringValue(out.Payload.Token)
	data.Status = types.StringValue(out.Payload.Status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AzureProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AzureProviderResourceModel
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
		handleError("Read Azure Integration", &resp.Diagnostics, err)
		return
	}

	state.Token = types.StringValue(out.Payload.Token)
	state.Id = types.StringValue(out.Payload.Token)
	state.Status = types.StringValue(out.Payload.Status)
	// tenant, app_id, and password are not returned by the API; preserve state values.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AzureProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All credential fields are RequiresReplace, so Update is never reached in practice.
	var plan AzureProviderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AzureProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AzureProviderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := integrationsv2.NewDeleteIntegrationParams()
	params.SetIntegrationToken(state.Token.ValueString())

	_, err := r.client.V2.Integrations.DeleteIntegration(params, r.client.Auth)
	if err != nil {
		handleError("Delete Azure Integration", &resp.Diagnostics, err)
	}
}