package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1integrations "github.com/vantage-sh/vantage-go/vantagev1/vantage/integrations"
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

	params := v1integrations.NewCreateIntegrationsAWSParams()
	params.SetCrossAccountArn(data.CrossAccountARN.ValueString())
	params.SetBucketArn(strPtr(data.BucketARN.ValueString()))
	out, err := r.client.V1.Integrations.CreateIntegrationsAWS(params, r.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while attempting to create the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"API Error: "+err.Error(),
		)
		return
	}

	if !out.IsSuccess() {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while attempting to create the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"API Error: "+out.Error(),
		)
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

	params := v1integrations.NewDeleteIntegrationsAWSParams()
	params.SetAccessCredentialID(int32(state.Id.ValueInt64()))
	out, err := r.client.V1.Integrations.DeleteIntegrationsAWS(params, r.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			"An unexpected error occurred while attempting to delete the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"API Error: "+err.Error(),
		)
		return
	}

	if !out.IsSuccess() {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			"An unexpected error occurred while attempting to delete the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"API Error: "+out.Error(),
		)
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

	params := v1integrations.NewGetIntegrationsAWSParams()
	params.SetAccessCredentialID(int32(state.Id.ValueInt64()))
	out, err := r.client.V1.Integrations.GetIntegrationsAWS(params, r.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Resource",
			"An unexpected error occurred while attempting to get the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"API Error: "+err.Error(),
		)
		return
	}

	if !out.IsSuccess() {
		resp.Diagnostics.AddError(
			"Unable to Get Resource",
			"An unexpected error occurred while attempting to get the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"API Error: "+out.Error(),
		)
		return
	}

	if out.Payload == nil {
		diags = resp.State.Set(ctx, &state)
		resp.Diagnostics.Append(diags...)
		return
	}

	// Overwrite items with refreshed state
	if out.Payload.BucketArn != "" {
		state.BucketARN = types.StringValue(out.Payload.BucketArn)
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

	params := v1integrations.NewPutIntegrationsAWSParams()
	params.SetAccessCredentialID(int32(data.Id.ValueInt64()))
	params.SetCrossAccountArn(data.CrossAccountARN.ValueString())
	params.SetBucketArn(strPtr(data.BucketARN.ValueString()))
	out, err := r.client.V1.Integrations.PutIntegrationsAWS(params, r.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while attempting to update the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"API Error: "+err.Error(),
		)
		return
	}

	if !out.IsSuccess() {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while attempting to update the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"API Error: "+out.Error(),
		)
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
