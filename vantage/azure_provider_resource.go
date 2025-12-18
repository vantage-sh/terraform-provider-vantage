package vantage

import (
    "context"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
    modelsv1 "github.com/vantage-sh/vantage-go/vantagev1/models"
    integrationsv1 "github.com/vantage-sh/vantage-go/vantagev1/vantage/integrations"
)

type AzureProviderResource struct{ client *Client }
func NewAzureProviderResource() resource.Resource { return &AzureProviderResource{} }
type AzureProviderResourceModel struct {
    TenantID       types.String `tfsdk:"tenant_id"`
    SubscriptionID types.String `tfsdk:"subscription_id"`
    ClientID       types.String `tfsdk:"client_id"`
    ClientSecret   types.String `tfsdk:"client_secret"`
    Id             types.Int64  `tfsdk:"id"`
}

func (r *AzureProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_azure_provider"
}

func (r AzureProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "tenant_id":       schema.StringAttribute{Required: true},
            "subscription_id": schema.StringAttribute{Required: true},
            "client_id":       schema.StringAttribute{Required: true},
            "client_secret":   schema.StringAttribute{Required: true, Sensitive: true},
            "id":              schema.Int64Attribute{Computed: true},
        },
        MarkdownDescription: "Manages an Azure Account Integration.",
    }
}

func (r AzureProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var data AzureProviderResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() { return }
    params := integrationsv1.NewCreateIntegrationsAzureParams()
    payload := &modelsv1.CreateIntegrationsAzure{
        TenantId:       data.TenantID.ValueStringPointer(),
        SubscriptionId: data.SubscriptionID.ValueStringPointer(),
        ClientId:       data.ClientID.ValueStringPointer(),
        ClientSecret:   data.ClientSecret.ValueStringPointer(),
    }
    params.WithCreateIntegrationsAzure(payload)
    out, err := r.client.V1.Integrations.CreateIntegrationsAzure(params, r.client.Auth)
    if err != nil { handleError("Create Azure Integration", &resp.Diagnostics, err); return }
    data.Id = types.Int64Value(int64(out.Payload.ID))
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r AzureProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state AzureProviderResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() { return }
    params := integrationsv1.NewGetIntegrationsAzureParams()
    params.SetAccessCredentialID(int32(state.Id.ValueInt64()))
    out, err := r.client.V1.Integrations.GetIntegrationsAzure(params, r.client.Auth)
    if err != nil { handleError("Read Azure Integration", &resp.Diagnostics, err); return }
    state.TenantID = types.StringValue(out.Payload.TenantId)
    state.SubscriptionID = types.StringValue(out.Payload.SubscriptionId)
    state.ClientID = types.StringValue(out.Payload.ClientId)
    state.ClientSecret = types.StringValue(out.Payload.ClientSecret)
    state.Id = types.Int64Value(int64(out.Payload.ID))
    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r AzureProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan AzureProviderResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() { return }
    params := integrationsv1.NewUpdateIntegrationsAzureParams()
    params.SetAccessCredentialID(int32(plan.Id.ValueInt64()))
    payload := &modelsv1.UpdateIntegrationsAzure{
        TenantId:       plan.TenantID.ValueStringPointer(),
        SubscriptionId: plan.SubscriptionID.ValueStringPointer(),
        ClientId:       plan.ClientID.ValueStringPointer(),
        ClientSecret:   plan.ClientSecret.ValueStringPointer(),
    }
    params.WithUpdateIntegrationsAzure(payload)
    out, err := r.client.V1.Integrations.UpdateIntegrationsAzure(params, r.client.Auth)
    if err != nil { handleError("Update Azure Integration", &resp.Diagnostics, err); return }
    plan.Id = types.Int64Value(int64(out.Payload.ID))
    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r AzureProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state AzureProviderResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() { return }
    params := integrationsv1.NewDeleteIntegrationsAzureParams()
    params.SetAccessCredentialID(int32(state.Id.ValueInt64()))
    _, err := r.client.V1.Integrations.DeleteIntegrationsAzure(params, r.client.Auth)
    if err != nil { handleError("Delete Azure Integration", &resp.Diagnostics, err); return }
}