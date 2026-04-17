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

type GcpProviderResource struct{ client *Client }

func NewGcpProviderResource() resource.Resource { return &GcpProviderResource{} }

type GcpProviderResourceModel struct {
	ProjectID      types.String `tfsdk:"project_id"`
	BillingAccount types.String `tfsdk:"billing_account"`
	DatasetName    types.String `tfsdk:"dataset_name"`
	Token          types.String `tfsdk:"token"`
	Id             types.String `tfsdk:"id"`
	Status         types.String `tfsdk:"status"`
}

func (r *GcpProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *GcpProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gcp_provider"
}

func (r *GcpProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "GCP project ID.",
			},
			"billing_account": schema.StringAttribute{
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "GCP billing account ID.",
			},
			"dataset_name": schema.StringAttribute{
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "BigQuery dataset name.",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Unique token of the GCP integration.",
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
		MarkdownDescription: "Manages a GCP Account Integration.",
	}
}

func (r *GcpProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GcpProviderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := integrationsv2.NewCreateGCPIntegrationParams()
	params.WithCreateGCPIntegration(&modelsv2.CreateGCPIntegration{
		ProjectID:        data.ProjectID.ValueStringPointer(),
		BillingAccountID: data.BillingAccount.ValueStringPointer(),
		DatasetName:      data.DatasetName.ValueStringPointer(),
	})

	out, err := r.client.V2.Integrations.CreateGCPIntegration(params, r.client.Auth)
	if err != nil {
		handleError("Create GCP Integration", &resp.Diagnostics, err)
		return
	}

	data.Token = types.StringValue(out.Payload.Token)
	data.Id = types.StringValue(out.Payload.Token)
	data.Status = types.StringValue(out.Payload.Status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GcpProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GcpProviderResourceModel
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
		handleError("Read GCP Integration", &resp.Diagnostics, err)
		return
	}

	state.Token = types.StringValue(out.Payload.Token)
	state.Id = types.StringValue(out.Payload.Token)
	state.Status = types.StringValue(out.Payload.Status)
	// project_id, billing_account, and dataset_name are not returned by the API; preserve state values.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *GcpProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All fields are RequiresReplace; Update is never reached in practice.
	var plan GcpProviderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GcpProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GcpProviderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := integrationsv2.NewDeleteIntegrationParams()
	params.SetIntegrationToken(state.Token.ValueString())

	_, err := r.client.V2.Integrations.DeleteIntegration(params, r.client.Auth)
	if err != nil {
		handleError("Delete GCP Integration", &resp.Diagnostics, err)
	}
}