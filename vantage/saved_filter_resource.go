package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	filtersv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/saved_filters"
)

type SavedFilterResource struct {
	client *Client
}

func NewSavedFilterResource() resource.Resource {
	return &SavedFilterResource{}
}

type SavedFilterResourceModel struct {
	Token          types.String `tfsdk:"token"`
	Title          types.String `tfsdk:"title"`
	Filter         types.String `tfsdk:"filter"`
	WorkspaceToken types.String `tfsdk:"workspace_token"`
}

func (r *SavedFilterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_saved_filter"
}

func (r SavedFilterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"title": schema.StringAttribute{
				MarkdownDescription: "Title of the Saved Filter",
				Required:            true,
			},
			"filter": schema.StringAttribute{
				MarkdownDescription: "VQL Query used for this saved filter.",
				Optional:            true,
				Computed:            true,
			},
			"workspace_token": schema.StringAttribute{
				MarkdownDescription: "Workspace token to add the saved filter into.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"token": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique saved filter identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		MarkdownDescription: "Manages a SavedFilter.",
	}
}

func (r SavedFilterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *SavedFilterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := filtersv2.NewCreateSavedFilterParams()
	body := &modelsv2.CreateSavedFilter{
		Title:          data.Title.ValueStringPointer(),
		Filter:         data.Filter.ValueString(),
		WorkspaceToken: data.WorkspaceToken.ValueString(),
	}
	params.WithCreateSavedFilter(body)
	out, err := r.client.V2.SavedFilters.CreateSavedFilter(params, r.client.Auth)
	if err != nil {
		//TODO(macb): Surface 400 errors more clearly.
		handleError("Create Saved Filter Resource", &resp.Diagnostics, err)
		return
	}

	data.Token = types.StringValue(out.Payload.Token)
	//TODO(macb): This value can be different than user input even though the
	// output is the same.
	//data.Filter = types.StringValue(out.Payload.Filter)
	data.WorkspaceToken = types.StringValue(out.Payload.WorkspaceToken)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r SavedFilterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *SavedFilterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := filtersv2.NewGetSavedFilterParams()
	params.SetSavedFilterToken(state.Token.ValueString())
	out, err := r.client.V2.SavedFilters.GetSavedFilter(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*filtersv2.GetSavedFilterNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Get Saved Filter Resource", &resp.Diagnostics, err)
		return
	}

	state.Token = types.StringValue(out.Payload.Token)
	state.Filter = types.StringValue(out.Payload.Filter)
	state.Title = types.StringValue(out.Payload.Title)
	state.WorkspaceToken = types.StringValue(out.Payload.WorkspaceToken)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r SavedFilterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *SavedFilterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := filtersv2.NewUpdateSavedFilterParams()
	params.WithSavedFilterToken(data.Token.ValueString())
	model := &modelsv2.UpdateSavedFilter{
		Title:  data.Title.ValueString(),
		Filter: data.Filter.ValueString(),
	}
	params.WithUpdateSavedFilter(model)
	out, err := r.client.V2.SavedFilters.UpdateSavedFilter(params, r.client.Auth)
	if err != nil {
		handleError("Update Saved Filter Resource", &resp.Diagnostics, err)
		return
	}

	// TODO(macb): filter is weird.
	//data.Filter = types.StringValue(out.Payload.Filter)
	data.Title = types.StringValue(out.Payload.Title)
	data.Token = types.StringValue(out.Payload.Token)
	data.WorkspaceToken = types.StringValue(out.Payload.WorkspaceToken)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r SavedFilterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *SavedFilterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := filtersv2.NewDeleteSavedFilterParams()
	params.SetSavedFilterToken(state.Token.ValueString())
	_, err := r.client.V2.SavedFilters.DeleteSavedFilter(params, r.client.Auth)
	if err != nil {
		handleError("Delete Saved Filter Resource", &resp.Diagnostics, err)
	}
}

// Configure adds the provider configured client to the data source.
func (r *SavedFilterResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}
