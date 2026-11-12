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
	Type              types.String `tfsdk:"type"`
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
		MarkdownDescription: "Looks up a folder by title and returns its token. Use `workspace_token` and `parent_folder_token` to narrow the search.",
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Filter by folder type: CostFolder, DashboardFolder, or ProviderResourceFolder. Also returns the matched folder's type.",
			},
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

	allFolders, err := fetchFoldersByType(d.client, state.Type.ValueStringPointer())
	if err != nil {
		handleError("Read Folder", &resp.Diagnostics, err)
		return
	}

	target := state.Title.ValueString()
	filterByWorkspace := !state.WorkspaceToken.IsNull() && !state.WorkspaceToken.IsUnknown()
	var workspaceFilter string
	if filterByWorkspace {
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
		if filterByWorkspace && folder.WorkspaceToken != workspaceFilter {
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

		state.Type = types.StringValue(folder.Type)
		state.Token = types.StringValue(folder.Token)
		state.WorkspaceToken = types.StringValue(folder.WorkspaceToken)
		state.ParentFolderToken = folderTokenState(state.ParentFolderToken, folder.ParentFolderToken)
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	filters := ""
	if !state.Type.IsNull() && !state.Type.IsUnknown() {
		filters += fmt.Sprintf(" of type %q", state.Type.ValueString())
	}
	if filterByWorkspace {
		filters += fmt.Sprintf(" in workspace %q", workspaceFilter)
	}
	if filterByParent {
		filters += fmt.Sprintf(" with parent folder %q", parentFolderFilter)
	}
	resp.Diagnostics.AddError(
		"Folder Not Found",
		fmt.Sprintf("No folder with title %q%s was found.", target, filters),
	)
}

// fetchAllFolders pages through the Get All Folders endpoint until
// links.next is nil, collecting every folder across all pages.
func fetchAllFolders(client *Client) ([]*modelsv2.Folder, error) {
	return fetchFoldersByType(client, nil)
}

func fetchFoldersByType(client *Client, folderType *string) ([]*modelsv2.Folder, error) {
	limit := int32(1000)
	var all []*modelsv2.Folder
	var page *int32

	for {
		params := foldersv2.NewGetFoldersParams()
		params.SetLimit(&limit)
		params.SetType(folderType)
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
