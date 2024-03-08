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
	var state teams
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

	teams := []team{}
	for _, t := range out.Payload.Teams {
		workspaceTokens, diag := types.SetValueFrom(ctx, types.StringType, t.WorkspaceTokens)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		userTokens, diag := types.SetValueFrom(ctx, types.StringType, t.UserTokens)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		userEmails, diag := types.SetValueFrom(ctx, types.StringType, t.UserEmails)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		teams = append(teams, team{
			Token:           types.StringValue(t.Token),
			Name:            types.StringValue(t.Name),
			Description:     types.StringValue(t.Description),
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
						"workspace_tokens": schema.SetAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"user_tokens": schema.SetAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"user_emails": schema.SetAttribute{
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
