package vantage

import (
    "context"

    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"

    modelsv1 "github.com/vantage-sh/vantage-go/vantagev1/models"
    integrationsv1 "github.com/vantage-sh/vantage-go/vantagev1/vantage/integrations"
)

type AwsProviderResource struct{ client *Client }

func NewAwsProviderResource() resource.Resource { return &AwsProviderResource{} }

type AwsProviderResourceModel struct {
    CrossAccountARN types.String `tfsdk:"cross_account_arn"`
    BucketARN       types.String `tfsdk:"bucket_arn"`
    Id              types.Int64  `tfsdk:"id"`
}

func (r *AwsProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_aws_provider"
}

func (r AwsProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "cross_account_arn": schema.StringAttribute{Required: true},
            "bucket_arn":        schema.StringAttribute{Optional: true},
            "id":                schema.Int64Attribute{Computed: true},
        },
        MarkdownDescription: "Manages an AWS Account Integration.",
    }
}

func (r AwsProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var data AwsProviderResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() { return }
    params := integrationsv1.NewCreateIntegrationsAWSParams()
    payload := &modelsv1.CreateIntegrationsAWS{
        CrossAccountArn: data.CrossAccountARN.ValueStringPointer(),
        BucketArn:       data.BucketARN.ValueString(),
    }
    params.WithCreateIntegrationsAWS(payload)
    out, err := r.client.V1.Integrations.CreateIntegrationsAWS(params, r.client.Auth)
    if err != nil { handleError("Create AWS Integration", &resp.Diagnostics, err); return }
    data.Id = types.Int64Value(int64(out.Payload.ID))
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r AwsProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state AwsProviderResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() { return }
    params := integrationsv1.NewGetIntegrationsAWSParams()
    params.SetAccessCredentialID(int32(state.Id.ValueInt64()))
    out, err := r.client.V1.Integrations.GetIntegrationsAWS(params, r.client.Auth)
    if err != nil { handleError("Read AWS Integration", &resp.Diagnostics, err); return }
    state.CrossAccountARN = types.StringValue(out.Payload.CrossAccountArn)
    state.BucketARN = types.StringValue(out.Payload.BucketArn)
    state.Id = types.Int64Value(int64(out.Payload.ID))
    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r AwsProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan AwsProviderResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() { return }
    params := integrationsv1.NewUpdateIntegrationsAWSParams()
    params.SetAccessCredentialID(int32(plan.Id.ValueInt64()))
    payload := &modelsv1.UpdateIntegrationsAWS{
        CrossAccountArn: plan.CrossAccountARN.ValueStringPointer(),
        BucketArn:       plan.BucketARN.ValueString(),
    }
    params.WithUpdateIntegrationsAWS(payload)
    out, err := r.client.V1.Integrations.UpdateIntegrationsAWS(params, r.client.Auth)
    if err != nil { handleError("Update AWS Integration", &resp.Diagnostics, err); return }
    plan.Id = types.Int64Value(int64(out.Payload.ID))
    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r AwsProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state AwsProviderResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() { return }
    params := integrationsv1.NewDeleteIntegrationsAWSParams()
    params.SetAccessCredentialID(int32(state.Id.ValueInt64()))
    _, err := r.client.V1.Integrations.DeleteIntegrationsAWS(params, r.client.Auth)
    if err != nil { handleError("Delete AWS Integration", &resp.Diagnostics, err); return }
}