package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	foldersv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/folders"
)

var (
	_ datasource.DataSource              = &foldersDataSource{}
	_ datasource.DataSourceWithConfigure = &foldersDataSource{}
)

func NewFoldersDataSource() datasource.DataSource {
	return &foldersDataSource{}
}

type foldersDataSource struct {
	client *Client
}

type folderDataSourceModel struct {
	Token             types.String `tfsdk:"token"`
	Title             types.String `tfsdk:"title"`
	ParentFolderToken types.String `tfsdk:"parent_folder_token"`
	SavedFilterTokens types.List   `tfsdk:"saved_filter_tokens"`
	WorkspaceToken    types.String `tfsdk:"workspace_token"`
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
						"parent_folder_token": schema.StringAttribute{
							Computed: true,
						},
						"saved_filter_tokens": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"workspace_token": schema.StringAttribute{
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

	for _, f := range out.Payload.Folders {
		savedFilterTokens, diag := types.ListValueFrom(ctx, types.StringType, f.SavedFilterTokens)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
		folders = append(folders, folderDataSourceModel{
			Title:             types.StringValue(f.Title),
			Token:             types.StringValue(f.Token),
			ParentFolderToken: types.StringValue(f.ParentFolderToken),
			SavedFilterTokens: savedFilterTokens,
			WorkspaceToken:    types.StringValue(f.WorkspaceToken),
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
