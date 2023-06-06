package vantage

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AwsProviderResource struct {
	client *vantageClient
}

func NewAwsProviderResource() resource.Resource {
	return &AwsProviderResource{}
}

// AwsProviderResourceModel describes the Terraform resource data model to
// match the resource schema.
type AwsProviderResourceModel struct {
	CrossAccountARN types.String `tfsdk:"cross_account_arn"`
	BucketARN       types.String `tfsdk:"bucket_arn"`
	Id              types.String `tfsdk:"id"`
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
			"id": schema.StringAttribute{
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

	// Convert from Terraform data model into API data model
	createReq := AwsProviderResourceAPIModel{
		CrossAccountARN: data.CrossAccountARN.ValueString(),
		BucketARN:       data.BucketARN.ValueString(),
	}

	provider, err := r.client.AddAwsProvider(createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while attempting to create the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"API Error: "+err.Error(),
		)

		return
	}

	// Convert from the API data model to the Terraform data model
	// and set any unknown attribute values.
	data.Id = types.StringValue(strconv.Itoa(provider.Id))

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

	id, err := strconv.Atoi(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			"An unexpected error occurred while attempting to delete the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"API Error: "+err.Error(),
		)

		return
	}

	err = r.client.DeleteAwsProvider(id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading AWS Provider",
			fmt.Sprintf("Could not delete AWS Provider ID %s: %v", state.Id.ValueString(), err.Error()),
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

	id, err := strconv.Atoi(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while attempting to update the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"API Error: "+err.Error(),
		)

		return
	}

	out, err := r.client.GetAwsProvider(id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading AWS Provider",
			fmt.Sprintf("Could not read AWS Provider ID %s: %v", state.Id.ValueString(), err.Error()),
		)
		return
	}

	if out == nil {
		diags = resp.State.Set(ctx, &state)
		resp.Diagnostics.Append(diags...)
		return
	}

	// Overwrite items with refreshed state
	if out.BucketARN != "" {
		state.BucketARN = types.StringValue(out.BucketARN)
	}
	state.CrossAccountARN = types.StringValue(out.CrossAccountARN)

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

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while attempting to update the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"API Error: "+err.Error(),
		)

		return
	}

	// Convert from Terraform data model into API data model
	updateRequest := AwsProviderResourceAPIModel{
		Id:              id,
		CrossAccountARN: data.CrossAccountARN.ValueString(),
		BucketARN:       data.BucketARN.ValueString(),
	}

	provider, err := r.client.UpdateAwsProvider(updateRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while attempting to update the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"API Error: "+err.Error(),
		)

		return
	}

	// Convert from the API data model to the Terraform data model
	// and set any unknown attribute values.
	data.Id = types.StringValue(strconv.Itoa(provider.Id))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

// Configure adds the provider configured client to the data source.
func (r *AwsProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*vantageClient)
}
