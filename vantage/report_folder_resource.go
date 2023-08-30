package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	costsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/costs"
)

type FolderResource struct {
	client *Client
}

func NewFolderResource() resource.Resource {
	return &FolderResource{}
}

type FolderResourceModel struct {
	Title             types.String `tfsdk:"title"`
	ParentFolderToken types.String `tfsdk:"parent_folder_token"`
	Token             types.String `tfsdk:"token"`
	WorkspaceToken    types.String `tfsdk:"workspace_token"`
}

func (r *FolderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_report_folder"
}

func (r FolderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"title": schema.StringAttribute{
				MarkdownDescription: "Title of the folder",
				Required:            true,
			},
			"parent_folder_token": schema.StringAttribute{
				MarkdownDescription: "Token of the folder's parent folder",
				Optional:            true,
				Computed:            true,
			},
			"workspace_token": schema.StringAttribute{
				MarkdownDescription: "Workspace token to add the cost report to.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"token": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique folder identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		MarkdownDescription: "Manages a Folder.",
	}
}

func (r FolderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *FolderResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := costsv2.NewCreateFolderParams()
	rf := &modelsv2.PostFolders{
		Title:             data.Title.ValueStringPointer(),
		ParentFolderToken: data.ParentFolderToken.ValueString(),
		WorkspaceToken:    data.WorkspaceToken.ValueString(),
	}
	params.WithFolders(rf)
	out, err := r.client.V2.Costs.CreateFolder(params, r.client.Auth)
	if err != nil {
		handleError("Create Report Folder Resource", &resp.Diagnostics, err)
		return
	}

	data.Token = types.StringValue(out.Payload.Token)
	data.Title = types.StringValue(out.Payload.Title)
	data.ParentFolderToken = types.StringValue(out.Payload.ParentFolderToken)
	data.WorkspaceToken = types.StringValue(out.Payload.WorkspaceToken)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r FolderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *FolderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := costsv2.NewGetFolderParams()
	params.SetFolderToken(state.Token.ValueString())
	out, err := r.client.V2.Costs.GetFolder(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*costsv2.GetFolderNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Get Report Folder Resource", &resp.Diagnostics, err)
		return
	}

	state.Token = types.StringValue(out.Payload.Token)
	state.ParentFolderToken = types.StringValue(out.Payload.ParentFolderToken)
	state.WorkspaceToken = types.StringValue(out.Payload.WorkspaceToken)
	state.Title = types.StringValue(out.Payload.Title)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r FolderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *FolderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := costsv2.NewUpdateFolderParams()
	params.WithFolderToken(data.Token.ValueString())
	model := &modelsv2.PutFolders{
		ParentFolderToken: data.ParentFolderToken.ValueString(),
		Title:             data.Title.ValueString(),
		// WorkspaceToken:    data.Title.WorkspaceToken(),
	}
	params.WithFolders(model)
	out, err := r.client.V2.Costs.UpdateFolder(params, r.client.Auth)
	if err != nil {
		handleError("Update Report Folder Resource", &resp.Diagnostics, err)
		return
	}

	data.ParentFolderToken = types.StringValue(out.Payload.ParentFolderToken)
	data.Title = types.StringValue(out.Payload.Title)
	data.WorkspaceToken = types.StringValue(out.Payload.WorkspaceToken)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r FolderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	panic("not implemented")
}

// Configure adds the provider configured client to the data source.
func (r *FolderResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}
