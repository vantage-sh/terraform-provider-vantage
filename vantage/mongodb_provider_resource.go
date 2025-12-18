package vantage

import (
    "context"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
    modelsv1 "github.com/vantage-sh/vantage-go/vantagev1/models"
    integrationsv1 "github.com/vantage-sh/vantage-go/vantagev1/vantage/integrations"
)

type MongodbProviderResource struct{ client *Client }
func NewMongodbProviderResource() resource.Resource { return &MongodbProviderResource{} }
type MongodbProviderResourceModel struct {
    ClusterUri types.String `tfsdk:"cluster_uri"`
    ApiKey     types.String `tfsdk:"api_key"`
    Id         types.Int64  `tfsdk:"id"`
}

func (r *MongodbProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_mongodb_provider"
}

func (r MongodbProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "cluster_uri": schema.StringAttribute{Required: true},
            "api_key":     schema.StringAttribute{Required: true, Sensitive: true},
            "id":          schema.Int64Attribute{Computed: true},
        },
        MarkdownDescription: "Manages a MongoDB Account Integration.",
    }
}

func (r MongodbProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var data MongodbProviderResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() { return }
    params := integrationsv1.NewCreateIntegrationsMongoDBParams()
    payload := &modelsv1.CreateIntegrationsMongoDB{
        ClusterUri: data.ClusterUri.ValueStringPointer(),
        ApiKey:     data.ApiKey.ValueStringPointer(),
    }
    params.WithCreateIntegrationsMongoDB(payload)
    out, err := r.client.V1.Integrations.CreateIntegrationsMongoDB(params, r.client.Auth)
    if err != nil { handleError("Create MongoDB Integration", &resp.Diagnostics, err); return }
    data.Id = types.Int64Value(int64(out.Payload.ID))
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r MongodbProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state MongodbProviderResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() { return }
    params := integrationsv1.NewGetIntegrationsMongoDBParams()
    params.SetAccessCredentialID(int32(state.Id.ValueInt64()))
    out, err := r.client.V1.Integrations.GetIntegrationsMongoDB(params, r.client.Auth)
    if err != nil { handleError("Read MongoDB Integration", &resp.Diagnostics, err); return }
    state.ClusterUri = types.StringValue(out.Payload.ClusterUri)
    state.ApiKey = types.StringValue(out.Payload.ApiKey)
    state.Id = types.Int64Value(int64(out.Payload.ID))
    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r MongodbProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan MongodbProviderResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() { return }
    params := integrationsv1.NewUpdateIntegrationsMongoDBParams()
    params.SetAccessCredentialID(int32(plan.Id.ValueInt64()))
    payload := &modelsv1.UpdateIntegrationsMongoDB{
        ClusterUri: plan.ClusterUri.ValueStringPointer(),
        ApiKey:     plan.ApiKey.ValueStringPointer(),
    }
    params.WithUpdateIntegrationsMongoDB(payload)
    out, err := r.client.V1.Integrations.UpdateIntegrationsMongoDB(params, r.client.Auth)
    if err != nil { handleError("Update MongoDB Integration", &resp.Diagnostics, err); return }
    plan.Id = types.Int64Value(int64(out.Payload.ID))
    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r MongodbProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state MongodbProviderResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() { return }
    params := integrationsv1.NewDeleteIntegrationsMongoDBParams()
    params.SetAccessCredentialID(int32(state.Id.ValueInt64()))
    _, err := r.client.V1.Integrations.DeleteIntegrationsMongoDB(params, r.client.Auth)
    if err != nil { handleError("Delete MongoDB Integration", &resp.Diagnostics, err); return }
}