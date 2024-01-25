package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	params := workspacesv2.NewGetWorkspacesParams()
	out, err := d.client.V2.Workspaces.GetWorkspaces(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Workspaces",
			err.Error(),
		)
		return
	}

	for _, workspace := range out.Payload.Workspaces {
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
