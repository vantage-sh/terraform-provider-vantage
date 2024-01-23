package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	foldersv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/folders"
)

var (
	_ datasource.DataSource              = &usersDataSource{}
	_ datasource.DataSourceWithConfigure = &usersDataSource{}
)

func NewFoldersDataSource() datasource.DataSource {
	return &foldersDataSource{}
}

type foldersDataSource struct {
	client *Client
}

type folderDataSourceModel struct {
	Token types.String `tfsdk:"token"`
	Title types.String `tfsdk:"title"`
}

type foldersDataSourceModel struct {
	Folders []folderDataSourceModel `tfsdk:"folders"`
}

func (d *foldersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_folders"
}

func (d *foldersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"folders": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"token": schema.StringAttribute{
							Computed: true,
						},
						"title": schema.StringAttribute{
							Computed: true,
						},
					},
				},
				Computed: true,
			},
		},
	}
}

func (d *foldersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state foldersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	params := foldersv2.NewGetFoldersParams()
	out, err := d.client.V2.Folders.GetFolders(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Folders",
			err.Error(),
		)
		return
	}

	folders := []folderDataSourceModel{}

	for _, u := range out.Payload.Folders {
		folders = append(folders, folderDataSourceModel{
			Title: types.StringValue(u.Title),
			Token: types.StringValue(u.Token),
		})
	}
	state.Folders = folders
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *foldersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*Client)
}
