package vantage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_workspace"
	workspacesv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/workspaces"
)

var (
	_ resource.Resource                = (*workspaceResource)(nil)
	_ resource.ResourceWithConfigure   = (*workspaceResource)(nil)
	_ resource.ResourceWithImportState = (*workspaceResource)(nil)
)

func NewWorkspaceResource() resource.Resource {
	return &workspaceResource{}
}

type workspaceResource struct {
	client *Client
}

func (r *workspaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (r *workspaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_workspace.WorkspaceResourceSchema(ctx)
}

func (r *workspaceResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}

func (r *workspaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r *workspaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *resource_workspace.WorkspaceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API call logic
	params := workspacesv2.NewCreateWorkspaceParams()
	params.SetDefaults()
	params.SetName(data.Name.ValueString())

	// Only set optional parameters if they have explicit values
	if !data.Currency.IsNull() && !data.Currency.IsUnknown() && data.Currency.ValueString() != "" {
		params.SetCurrency(data.Currency.ValueStringPointer())
	}
	if !data.EnableCurrencyConversion.IsNull() && !data.EnableCurrencyConversion.IsUnknown() {
		params.SetEnableCurrencyConversion(data.EnableCurrencyConversion.ValueBoolPointer())
	}
	if !data.ExchangeRateDate.IsNull() && !data.ExchangeRateDate.IsUnknown() && data.ExchangeRateDate.ValueString() != "" {
		params.SetExchangeRateDate(data.ExchangeRateDate.ValueStringPointer())
	}

	out, err := r.client.V2.Workspaces.CreateWorkspace(params, r.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Vantage Workspace",
			fmt.Sprintf("API Error: %v", err.Error()),
		)
		return
	}

	workspace := out.Payload
	data.Id = types.StringValue(workspace.Token) // Use Token as ID
	data.Token = types.StringValue(workspace.Token)
	data.Name = types.StringValue(workspace.Name)
	data.CreatedAt = types.StringValue(workspace.CreatedAt)
	data.Currency = types.StringValue(workspace.Currency)
	data.EnableCurrencyConversion = types.BoolValue(workspace.EnableCurrencyConversion)
	data.ExchangeRateDate = types.StringValue(workspace.ExchangeRateDate)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *workspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *resource_workspace.WorkspaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	params := workspacesv2.NewGetWorkspaceParams()
	params.SetWorkspaceToken(state.Token.ValueString())

	out, err := r.client.V2.Workspaces.GetWorkspace(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*workspacesv2.GetWorkspaceNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Vantage Workspace",
			err.Error(),
		)
		return
	}

	workspace := out.Payload
	state.Id = types.StringValue(workspace.Token)
	state.Token = types.StringValue(workspace.Token)
	state.Name = types.StringValue(workspace.Name)
	state.CreatedAt = types.StringValue(workspace.CreatedAt)
	state.Currency = types.StringValue(workspace.Currency)
	state.EnableCurrencyConversion = types.BoolValue(workspace.EnableCurrencyConversion)
	state.ExchangeRateDate = types.StringValue(workspace.ExchangeRateDate)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *workspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *resource_workspace.WorkspaceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update API call logic
	params := workspacesv2.NewUpdateWorkspaceParams()
	params.SetWorkspaceToken(data.Token.ValueString())

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		params.SetName(data.Name.ValueStringPointer())
	}
	if !data.Currency.IsNull() && !data.Currency.IsUnknown() {
		params.SetCurrency(data.Currency.ValueStringPointer())
	}
	if !data.EnableCurrencyConversion.IsNull() && !data.EnableCurrencyConversion.IsUnknown() {
		params.SetEnableCurrencyConversion(data.EnableCurrencyConversion.ValueBoolPointer())
	}
	if !data.ExchangeRateDate.IsNull() && !data.ExchangeRateDate.IsUnknown() {
		params.SetExchangeRateDate(data.ExchangeRateDate.ValueStringPointer())
	}

	out, err := r.client.V2.Workspaces.UpdateWorkspace(params, r.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Vantage Workspace",
			err.Error(),
		)
		return
	}

	workspace := out.Payload
	data.Id = types.StringValue(workspace.Token)
	data.Token = types.StringValue(workspace.Token)
	data.Name = types.StringValue(workspace.Name)
	data.CreatedAt = types.StringValue(workspace.CreatedAt)
	data.Currency = types.StringValue(workspace.Currency)
	data.EnableCurrencyConversion = types.BoolValue(workspace.EnableCurrencyConversion)
	data.ExchangeRateDate = types.StringValue(workspace.ExchangeRateDate)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *workspaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *resource_workspace.WorkspaceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete API call logic
	params := workspacesv2.NewDeleteWorkspaceParams()
	params.SetWorkspaceToken(state.Token.ValueString())

	_, err := r.client.V2.Workspaces.DeleteWorkspace(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*workspacesv2.DeleteWorkspaceNotFound); ok {
			// Resource already deleted
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Delete Vantage Workspace",
			err.Error(),
		)
		return
	}
}
