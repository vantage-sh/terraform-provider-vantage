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
	_ datasource.DataSource              = (*folderLookupDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*folderLookupDataSource)(nil)
)

func NewFolderDataSource() datasource.DataSource {
	return &folderLookupDataSource{}
}

type folderLookupModel struct {
	Title             types.String `tfsdk:"title"`
	Token             types.String `tfsdk:"token"`
	ParentFolderToken types.String `tfsdk:"parent_folder_token"`
	WorkspaceToken    types.String `tfsdk:"workspace_token"`
}

type folderLookupDataSource struct {
	client *Client
}

func (d *folderLookupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

func (d *folderLookupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_folder"
}

func (d *folderLookupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up a folder by title and returns its token. Use `workspace_token` to narrow the search to a specific workspace.",
		Attributes: map[string]schema.Attribute{
			"title": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The title of the folder to find.",
			},
			"workspace_token": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Filter folders by workspace token. If not specified, the first folder matching the title is returned. Also populated as an output with the workspace token of the matched folder.",
			},
			"parent_folder_token": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Filter folders by parent folder token. Set to an empty string (`\"\"`) to match only root-level folders (those with no parent). Also populated as an output with the parent folder token of the matched folder.",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique token of the matched folder.",
			},
		},
	}
}

func (d *folderLookupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state folderLookupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	allFolders, err := fetchAllFolders(d.client)
	if err != nil {
		handleError("Read Folder", &resp.Diagnostics, err)
		return
	}

	target := state.Title.ValueString()
	var workspaceFilter string
	if !state.WorkspaceToken.IsNull() && !state.WorkspaceToken.IsUnknown() {
		workspaceFilter = state.WorkspaceToken.ValueString()
	}
	filterByParent := !state.ParentFolderToken.IsNull() && !state.ParentFolderToken.IsUnknown()
	var parentFolderFilter string
	if filterByParent {
		parentFolderFilter = state.ParentFolderToken.ValueString()
	}

	for _, folder := range allFolders {
		if folder.Title == nil || *folder.Title != target {
			continue
		}
		if workspaceFilter != "" && folder.WorkspaceToken != workspaceFilter {
			continue
		}
		folderParent := ""
		if folder.ParentFolderToken != nil {
			folderParent = *folder.ParentFolderToken
		}
		if filterByParent {
			if folderParent != parentFolderFilter {
				continue
			}
		}

		state.Token = types.StringValue(folder.Token)
		state.WorkspaceToken = types.StringValue(folder.WorkspaceToken)
		state.ParentFolderToken = types.StringPointerValue(folderParent)
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	resp.Diagnostics.AddError(
		"Folder Not Found",
		fmt.Sprintf("No folder with title %q was found.", target),
	)
}

// fetchAllFolders pages through the Get All Folders endpoint until
// links.next is nil, collecting every folder across all pages.
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
