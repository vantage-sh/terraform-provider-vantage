package vantage

import (
    "context"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
    modelsv1 "github.com/vantage-sh/vantage-go/vantagev1/models"
    integrationsv1 "github.com/vantage-sh/vantage-go/vantagev1/vantage/integrations"
)

type SnowflakeProviderResource struct{ client *Client }
func NewSnowflakeProviderResource() resource.Resource { return &SnowflakeProviderResource{} }
type SnowflakeProviderResourceModel struct {
    AccountName types.String `tfsdk:"account_name"`
    UserName    types.String `tfsdk:"user_name"`
    Password    types.String `tfsdk:"password"`
    Role        types.String `tfsdk:"role"`
    Id          types.Int64  `tfsdk:"id"`
}

func (r *SnowflakeProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_snowflake_provider"
}

func (r SnowflakeProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "account_name": schema.StringAttribute{Required: true},
            "user_name":    schema.StringAttribute{Required: true},
            "password":     schema.StringAttribute{Required: true, Sensitive: true},
            "role":         schema.StringAttribute{Optional: true},
            "id":           schema.Int64Attribute{Computed: true},
        },
        MarkdownDescription: "Manages a Snowflake Account Integration.",
    }
}

func (r SnowflakeProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var data SnowflakeProviderResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() { return }
    params := integrationsv1.NewCreateIntegrationsSnowflakeParams()
    payload := &modelsv1.CreateIntegrationsSnowflake{
        AccountName: data.AccountName.ValueStringPointer(),
        UserName:    data.UserName.ValueStringPointer(),
        Password:    data.Password.ValueStringPointer(),
        Role:        data.Role.ValueStringPointer(),
    }
    params.WithCreateIntegrationsSnowflake(payload)
    out, err := r.client.V1.Integrations.CreateIntegrationsSnowflake(params, r.client.Auth)
    if err != nil { handleError("Create Snowflake Integration", &resp.Diagnostics, err); return }
    data.Id = types.Int64Value(int64(out.Payload.ID))
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r SnowflakeProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state SnowflakeProviderResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() { return }
    params := integrationsv1.NewGetIntegrationsSnowflakeParams()
    params.SetAccessCredentialID(int32(state.Id.ValueInt64()))
    out, err := r.client.V1.Integrations.GetIntegrationsSnowflake(params, r.client.Auth)
    if err != nil { handleError("Read Snowflake Integration", &resp.Diagnostics, err); return }
    state.AccountName = types.StringValue(out.Payload.AccountName)
    state.UserName = types.StringValue(out.Payload.UserName)
    state.Password = types.StringValue(out.Payload.Password)
    state.Role = types.StringValue(out.Payload.Role)
    state.Id = types.Int64Value(int64(out.Payload.ID))
    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r SnowflakeProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan SnowflakeProviderResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() { return }
    params := integrationsv1.NewUpdateIntegrationsSnowflakeParams()
    params.SetAccessCredentialID(int32(plan.Id.ValueInt64()))
    payload := &modelsv1.UpdateIntegrationsSnowflake{
        AccountName: plan.AccountName.ValueStringPointer(),
        UserName:    plan.UserName.ValueStringPointer(),
        Password:    plan.Password.ValueStringPointer(),
        Role:        plan.Role.ValueStringPointer(),
    }
    params.WithUpdateIntegrationsSnowflake(payload)
    out, err := r.client.V1.Integrations.UpdateIntegrationsSnowflake(params, r.client.Auth)
    if err != nil { handleError("Update Snowflake Integration", &resp.Diagnostics, err); return }
    plan.Id = types.Int64Value(int64(out.Payload.ID))
    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r SnowflakeProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state SnowflakeProviderResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() { return }
    params := integrationsv1.NewDeleteIntegrationsSnowflakeParams()
    params.SetAccessCredentialID(int32(state.Id.ValueInt64()))
    _, err := r.client.V1.Integrations.DeleteIntegrationsSnowflake(params, r.client.Auth)
    if err != nil { handleError("Delete Snowflake Integration", &resp.Diagnostics, err); return }
}