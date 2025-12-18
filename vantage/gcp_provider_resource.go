package vantage

import (
    "context"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
    modelsv1 "github.com/vantage-sh/vantage-go/vantagev1/models"
    integrationsv1 "github.com/vantage-sh/vantage-go/vantagev1/vantage/integrations"
)

type GcpProviderResource struct{ client *Client }
func NewGcpProviderResource() resource.Resource { return &GcpProviderResource{} }
type GcpProviderResourceModel struct {
    ProjectID      types.String `tfsdk:"project_id"`
    BillingAccount types.String `tfsdk:"billing_account"`
    ServiceAccount types.String `tfsdk:"service_account"`
    Id             types.Int64  `tfsdk:"id"`
}

func (r *GcpProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_gcp_provider"
}

func (r GcpProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "project_id":      schema.StringAttribute{Required: true},
            "billing_account": schema.StringAttribute{Required: true},
            "service_account": schema.StringAttribute{Required: true, Sensitive: true},
            "id":              schema.Int64Attribute{Computed: true},
        },
        MarkdownDescription: "Manages a GCP Account Integration.",
    }
}

func (r GcpProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var data GcpProviderResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() { return }
    params := integrationsv1.NewCreateIntegrationsGCPParams()
    payload := &modelsv1.CreateIntegrationsGCP{
        ProjectId:      data.ProjectID.ValueStringPointer(),
        BillingAccount: data.BillingAccount.ValueStringPointer(),
        ServiceAccount: data.ServiceAccount.ValueStringPointer(),
    }
    params.WithCreateIntegrationsGCP(payload)
    out, err := r.client.V1.Integrations.CreateIntegrationsGCP(params, r.client.Auth)
    if err != nil { handleError("Create GCP Integration", &resp.Diagnostics, err); return }
    data.Id = types.Int64Value(int64(out.Payload.ID))
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r GcpProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state GcpProviderResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() { return }
    params := integrationsv1.NewGetIntegrationsGCPParams()
    params.SetAccessCredentialID(int32(state.Id.ValueInt64()))
    out, err := r.client.V1.Integrations.GetIntegrationsGCP(params, r.client.Auth)
    if err != nil { handleError("Read GCP Integration", &resp.Diagnostics, err); return }
    state.ProjectID = types.StringValue(out.Payload.ProjectId)
    state.BillingAccount = types.StringValue(out.Payload.BillingAccount)
    state.ServiceAccount = types.StringValue(out.Payload.ServiceAccount)
    state.Id = types.Int64Value(int64(out.Payload.ID))
    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r GcpProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan GcpProviderResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() { return }
    params := integrationsv1.NewUpdateIntegrationsGCPParams()
    params.SetAccessCredentialID(int32(plan.Id.ValueInt64()))
    payload := &modelsv1.UpdateIntegrationsGCP{
        ProjectId:      plan.ProjectID.ValueStringPointer(),
        BillingAccount: plan.BillingAccount.ValueStringPointer(),
        ServiceAccount: plan.ServiceAccount.ValueStringPointer(),
    }
    params.WithUpdateIntegrationsGCP(payload)
    out, err := r.client.V1.Integrations.UpdateIntegrationsGCP(params, r.client.Auth)
    if err != nil { handleError("Update GCP Integration", &resp.Diagnostics, err); return }
    plan.Id = types.Int64Value(int64(out.Payload.ID))
    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r GcpProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state GcpProviderResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() { return }
    params := integrationsv1.NewDeleteIntegrationsGCPParams()
    params.SetAccessCredentialID(int32(state.Id.ValueInt64()))
    _, err := r.client.V1.Integrations.DeleteIntegrationsGCP(params, r.client.Auth)
    if err != nil { handleError("Delete GCP Integration", &resp.Diagnostics, err); return }
}