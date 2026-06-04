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
	_ datasource.DataSource              = &workspacesDataSource{}
	_ datasource.DataSourceWithConfigure = &workspacesDataSource{}
)

func NewWorkspacesDataSource() datasource.DataSource {
	return &workspacesDataSource{}
}

type workspaceDataSourceModel struct {
	Token types.String `tfsdk:"token"`
	Name  types.String `tfsdk:"name"`
}

type workspacesDataSourceModel struct {
	Workspaces []workspaceDataSourceModel `tfsdk:"workspaces"`
}

type workspacesDataSource struct {
	client *Client
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *workspacesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*Client)
}

// Metadata implements datasource.DataSource.
func (d *workspacesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspaces"
}

// Read implements datasource.DataSource.
func (d *workspacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state workspacesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	workspaces, err := fetchAllWorkspaces(d.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Workspaces",
			err.Error(),
		)
		return
	}

	for _, workspace := range workspaces {
		state.Workspaces = append(state.Workspaces, workspaceDataSourceModel{
			Token: types.StringValue(workspace.Token),
			Name:  types.StringValue(workspace.Name),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Schema implements datasource.DataSource.
func (d *workspacesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"workspaces": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"token": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
					},
				},
				Computed: true,
			},
		},
	}
}

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
