package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	teamsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/teams"
)

var (
	_ datasource.DataSource              = &teamsDataSource{}
	_ datasource.DataSourceWithConfigure = &teamsDataSource{}
)

func NewTeamsDataSource() datasource.DataSource {
	return &teamsDataSource{}
}

type teamDataSourceModel struct {
	Token           types.String `tfsdk:"token"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	WorkspaceTokens types.List   `tfsdk:"workspace_tokens"`
	UserTokens      types.List   `tfsdk:"user_tokens"`
	UserEmails      types.List   `tfsdk:"user_emails"`
}

type teamsDataSourceModel struct {
	Teams []teamDataSourceModel `tfsdk:"teams"`
}

type teamsDataSource struct {
	client *Client
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *teamsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*Client)
}

// Metadata implements datasource.DataSource.
func (d *teamsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_teams"
}

// Read implements datasource.DataSource.
func (d *teamsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state teamsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	params := teamsv2.NewGetTeamsParams()
	out, err := d.client.V2.Teams.GetTeams(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Teams",
			err.Error(),
		)
		return
	}

	teams := []teamDataSourceModel{}
	for _, team := range out.Payload.Teams {
		workspaceTokens, diag := types.ListValueFrom(ctx, types.StringType, team.WorkspaceTokens)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		userTokens, diag := types.ListValueFrom(ctx, types.StringType, team.UserTokens)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		userEmails, diag := types.ListValueFrom(ctx, types.StringType, team.UserEmails)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		teams = append(teams, teamDataSourceModel{
			Token:           types.StringValue(team.Token),
			Name:            types.StringValue(team.Name),
			Description:     types.StringValue(team.Description),
			WorkspaceTokens: workspaceTokens,
			UserTokens:      userTokens,
			UserEmails:      userEmails,
		})
	}

	state.Teams = teams
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Schema implements datasource.DataSource.
func (d *teamsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"teams": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"token": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"workspace_tokens": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"user_tokens": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"user_emails": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
					},
				},
				Computed: true,
			},
		},
	}
}
