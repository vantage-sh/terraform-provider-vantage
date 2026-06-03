package vantage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	workspacesv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/workspaces"
)

var (
	_ datasource.DataSource              = (*workspaceByNameDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*workspaceByNameDataSource)(nil)
)

func NewWorkspaceByNameDataSource() datasource.DataSource {
	return &workspaceByNameDataSource{}
}

type workspaceByNameDataSourceModel struct {
	Name  types.String `tfsdk:"name"`
	Token types.String `tfsdk:"token"`
}

type workspaceByNameDataSource struct {
	client *Client
}

func (d *workspaceByNameDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

func (d *workspaceByNameDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_by_name"
}

func (d *workspaceByNameDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up a workspace by name and returns its token.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the workspace to find.",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique token of the matched workspace.",
			},
		},
	}
}

func (d *workspaceByNameDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state workspaceByNameDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := workspacesv2.NewGetWorkspacesParams()
	out, err := d.client.V2.Workspaces.GetWorkspaces(params, d.client.Auth)
	if err != nil {
		handleError("Read Workspace By Name", &resp.Diagnostics, err)
		return
	}

	target := state.Name.ValueString()
	for _, workspace := range out.Payload.Workspaces {
		if workspace.Name == target {
			state.Token = types.StringValue(workspace.Token)
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	resp.Diagnostics.AddError(
		"Workspace Not Found",
		fmt.Sprintf("No workspace with name %q was found.", target),
	)
}
