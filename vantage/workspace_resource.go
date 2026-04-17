package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_workspace"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	workspacesv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/workspaces"
)

var (
	_ resource.Resource                   = (*WorkspaceResource)(nil)
	_ resource.ResourceWithConfigure      = (*WorkspaceResource)(nil)
	_ resource.ResourceWithImportState    = (*WorkspaceResource)(nil)
	_ resource.ResourceWithValidateConfig = (*WorkspaceResource)(nil)
)

type WorkspaceResource struct {
	client *Client
}

func NewWorkspaceResource() resource.Resource {
	return &WorkspaceResource{}
}

func (r *WorkspaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (r WorkspaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_workspace.WorkspaceResourceSchema(ctx)
	s.MarkdownDescription = "Manages a Workspace."

	s.Attributes["token"] = schema.StringAttribute{
		Computed:            true,
		Description:         "The token of the workspace",
		MarkdownDescription: "The token of the workspace",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
	s.Attributes["id"] = schema.StringAttribute{
		Computed:            true,
		Description:         "Alias of `token`.",
		MarkdownDescription: "Alias of `token`.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	resp.Schema = s
}

func (r WorkspaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *resource_workspace.WorkspaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := &modelsv2.CreateWorkspace{
		Name: data.Name.ValueStringPointer(),
	}
	if !data.Currency.IsNull() && !data.Currency.IsUnknown() && data.Currency.ValueString() != "" {
		body.Currency = data.Currency.ValueString()
	}
	if !data.EnableCurrencyConversion.IsNull() && !data.EnableCurrencyConversion.IsUnknown() {
		v := data.EnableCurrencyConversion.ValueBool()
		body.EnableCurrencyConversion = &v
	}
	if !data.ExchangeRateDate.IsNull() && !data.ExchangeRateDate.IsUnknown() && data.ExchangeRateDate.ValueString() != "" {
		s := data.ExchangeRateDate.ValueString()
		body.ExchangeRateDate = &s
	}

	params := workspacesv2.NewCreateWorkspaceParams().WithCreateWorkspace(body)
	out, err := r.client.V2.Workspaces.CreateWorkspace(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*workspacesv2.CreateWorkspaceBadRequest); ok {
			handleBadRequest("Create Workspace Resource", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Create Workspace Resource", &resp.Diagnostics, err)
		return
	}

	applyWorkspacePayload(out.Payload, data)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r WorkspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *resource_workspace.WorkspaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := workspacesv2.NewGetWorkspaceParams()
	params.SetWorkspaceToken(state.Token.ValueString())
	out, err := r.client.V2.Workspaces.GetWorkspace(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*workspacesv2.GetWorkspaceNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}
		handleError("Read Workspace Resource", &resp.Diagnostics, err)
		return
	}

	applyWorkspacePayload(out.Payload, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r WorkspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *resource_workspace.WorkspaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model := &modelsv2.UpdateWorkspace{
		Name: data.Name.ValueString(),
	}
	if !data.Currency.IsNull() && !data.Currency.IsUnknown() && data.Currency.ValueString() != "" {
		model.Currency = data.Currency.ValueString()
	}
	if !data.EnableCurrencyConversion.IsNull() && !data.EnableCurrencyConversion.IsUnknown() {
		v := data.EnableCurrencyConversion.ValueBool()
		model.EnableCurrencyConversion = &v
	}
	if !data.ExchangeRateDate.IsNull() && !data.ExchangeRateDate.IsUnknown() && data.ExchangeRateDate.ValueString() != "" {
		s := data.ExchangeRateDate.ValueString()
		model.ExchangeRateDate = &s
	}

	params := workspacesv2.NewUpdateWorkspaceParams()
	params.SetWorkspaceToken(data.Token.ValueString())
	params.WithUpdateWorkspace(model)
	out, err := r.client.V2.Workspaces.UpdateWorkspace(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*workspacesv2.UpdateWorkspaceBadRequest); ok {
			handleBadRequest("Update Workspace Resource", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Update Workspace Resource", &resp.Diagnostics, err)
		return
	}

	applyWorkspacePayload(out.Payload, data)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r WorkspaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *resource_workspace.WorkspaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := workspacesv2.NewDeleteWorkspaceParams()
	params.SetWorkspaceToken(state.Token.ValueString())
	_, err := r.client.V2.Workspaces.DeleteWorkspace(params, r.client.Auth)
	if err != nil {
		handleError("Delete Workspace Resource", &resp.Diagnostics, err)
	}
}

func (r WorkspaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("token"), types.StringValue(req.ID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
}

func (r *WorkspaceResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data resource_workspace.WorkspaceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// currency is only meaningful when enable_currency_conversion is true.
	// When conversion is disabled the API returns display_currency (USD) regardless.
	if !data.EnableCurrencyConversion.IsNull() && !data.EnableCurrencyConversion.IsUnknown() &&
		!data.EnableCurrencyConversion.ValueBool() &&
		!data.Currency.IsNull() && !data.Currency.IsUnknown() &&
		data.Currency.ValueString() != "" && data.Currency.ValueString() != "USD" {
		resp.Diagnostics.AddAttributeError(
			path.Root("currency"),
			"Invalid Attribute Combination",
			"currency can only be set to a non-USD value when enable_currency_conversion is true. "+
				"When currency conversion is disabled, costs are displayed in USD.",
		)
	}
}

func (r *WorkspaceResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func applyWorkspacePayload(w *modelsv2.Workspace, m *resource_workspace.WorkspaceModel) {
	if w == nil || m == nil {
		return
	}
	m.Token = types.StringValue(w.Token)
	m.Id = types.StringValue(w.Token)
	m.Name = types.StringValue(w.Name)
	m.Currency = types.StringValue(w.Currency)
	m.EnableCurrencyConversion = types.BoolValue(w.EnableCurrencyConversion)
	m.ExchangeRateDate = types.StringValue(w.ExchangeRateDate)
	m.CreatedAt = types.StringValue(w.CreatedAt)
}
