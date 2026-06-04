package vantage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
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

	allFolders, err := fetchAllFolders(d.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Folders",
			err.Error(),
		)
		return
	}

	folders := []folderDataSourceModel{}

	for _, f := range allFolders {
		savedFilterTokens, diag := types.ListValueFrom(ctx, types.StringType, f.SavedFilterTokens)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
		folders = append(folders, folderDataSourceModel{
			Title:             types.StringPointerValue(f.Title),
			Token:             types.StringValue(f.Token),
			ParentFolderToken: types.StringPointerValue(f.ParentFolderToken),
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

func fetchAllFolders(client *Client) ([]*modelsv2.Folder, error) {
	limit := int32(1000)
	var all []*modelsv2.Folder
	var page *int32

	for {
		params := foldersv2.NewGetFoldersParams()
		params.SetLimit(&limit)
		if page != nil {
			params.SetPage(page)
		}

		out, err := client.V2.Folders.GetFolders(params, client.Auth)
		if err != nil {
			return nil, err
		}

		all = append(all, out.Payload.Folders...)

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
