package vantage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	workspacesv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/workspaces"
)

var (
	_ datasource.DataSource              = (*workspaceLookupDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*workspaceLookupDataSource)(nil)
)

func NewWorkspaceDataSource() datasource.DataSource {
	return &workspaceLookupDataSource{}
}

type workspaceLookupModel struct {
	Name  types.String `tfsdk:"name"`
	Token types.String `tfsdk:"token"`
}

type workspaceLookupDataSource struct {
	client *Client
}

func (d *workspaceLookupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

func (d *workspaceLookupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (d *workspaceLookupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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

func (d *workspaceLookupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state workspaceLookupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	allWorkspaces, err := fetchAllWorkspaces(d.client)
	if err != nil {
		handleError("Read Workspace", &resp.Diagnostics, err)
		return
	}

	target := state.Name.ValueString()
	for _, workspace := range allWorkspaces {
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

// fetchAllWorkspaces pages through the Get All Workspaces endpoint until
// links.next is nil, collecting every workspace across all pages.
func fetchAllWorkspaces(client *Client) ([]*modelsv2.Workspace, error) {
	limit := int32(1000)
	var all []*modelsv2.Workspace
	var page *int32

	for {
		params := workspacesv2.NewGetWorkspacesParams()
		params.SetLimit(&limit)
		if page != nil {
			params.SetPage(page)
		}

		out, err := client.V2.Workspaces.GetWorkspaces(params, client.Auth)
		if err != nil {
			return nil, err
		}

		all = append(all, out.Payload.Workspaces...)

		if out.Payload.Links == nil || out.Payload.Links.Next == nil {
			break
		}

		nextPage, err := pageFromURL(*out.Payload.Links.Next)
		if err != nil {
			return nil, fmt.Errorf("parsing next page from links.next %q: %w", *out.Payload.Links.Next, err)
		}
		page = &nextPage
	}

	return all, nil
}
