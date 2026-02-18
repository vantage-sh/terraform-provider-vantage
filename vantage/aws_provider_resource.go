package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv1 "github.com/vantage-sh/vantage-go/vantagev1/models"
	integrationsv1 "github.com/vantage-sh/vantage-go/vantagev1/vantage/integrations"
)

type AwsProviderResource struct {
	client *Client
}

func NewAwsProviderResource() resource.Resource {
	return &AwsProviderResource{}
}

// AwsProviderResourceModel describes the Terraform resource data model to
// match the resource schema.
type AwsProviderResourceModel struct {
	CrossAccountARN types.String `tfsdk:"cross_account_arn"`
	BucketARN       types.String `tfsdk:"bucket_arn"`
	Id              types.Int64  `tfsdk:"id"`
}

func (r *AwsProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_provider"
}

func (r AwsProviderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cross_account_arn": schema.StringAttribute{
				MarkdownDescription: "ARN to use for cross account access.",
				Required:            true,
			},
			"bucket_arn": schema.StringAttribute{
				MarkdownDescription: "Bucket ARN for where CUR data is being stored.",
				Optional:            true,
			},
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Service generated identifier for the the account access.",
				//PlanModifiers: planmodifier.String{
				//stringplanmodifier.UseStateForUnknown(),
				//},
			},
		},
		MarkdownDescription: "Manages an AWS Account Integration.",
	}
}

func (r AwsProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *AwsProviderResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := integrationsv1.NewCreateIntegrationsAWSParams()
	model := &modelsv1.CreateIntegrationsAWS{
		CrossAccountArn: data.CrossAccountARN.ValueStringPointer(),
		BucketArn:       data.BucketARN.ValueString(),
	}
	params.WithCreateIntegrationsAWS(model)
	out, err := r.client.V1.Integrations.CreateIntegrationsAWS(params, r.client.Auth)
	if err != nil {
		handleError("Create AWS Integration", &resp.Diagnostics, err)
		return
	}

	data.Id = types.Int64Value(int64(out.Payload.ID))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r AwsProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AwsProviderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := integrationsv1.NewDeleteIntegrationsAWSParams()
	params.SetAccessCredentialID(int32(state.Id.ValueInt64()))
	_, err := r.client.V1.Integrations.DeleteIntegrationsAWS(params, r.client.Auth)
	if err != nil {
		handleError("Delete AWS Integration", &resp.Diagnostics, err)
		return
	}
}

func (r AwsProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AwsProviderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := integrationsv1.NewGetIntegrationsAWSParams()
	params.SetAccessCredentialID(int32(state.Id.ValueInt64()))
	out, err := r.client.V1.Integrations.GetIntegrationsAWS(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*integrationsv1.GetIntegrationsAWSNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}
		handleError("Get AWS Integration", &resp.Diagnostics, err)
		return
	}

	// Overwrite items with refreshed state
	if out.Payload.BucketArn != nil && *out.Payload.BucketArn != "" {
		state.BucketARN = types.StringPointerValue(out.Payload.BucketArn)
	} else {
		state.BucketARN = types.StringNull()
	}
	state.CrossAccountARN = types.StringValue(out.Payload.CrossAccountArn)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r AwsProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *AwsProviderResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := integrationsv1.NewPutIntegrationsAWSParams()
	params.SetAccessCredentialID(int32(data.Id.ValueInt64()))
	m := &modelsv1.PutIntegrationsAWS{
		CrossAccountArn: data.CrossAccountARN.ValueStringPointer(),
		BucketArn:       *data.BucketARN.ValueStringPointer(),
	}
	params.WithPutIntegrationsAWS(m)
	out, err := r.client.V1.Integrations.PutIntegrationsAWS(params, r.client.Auth)
	if err != nil {
		handleError("Update AWS Integration", &resp.Diagnostics, err)
		return
	}
	data.Id = types.Int64Value(int64(out.Payload.ID))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

// Configure adds the provider configured client to the data source.
func (r *AwsProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}
