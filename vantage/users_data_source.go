package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	usersv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/users"
)

var (
	_ datasource.DataSource              = &usersDataSource{}
	_ datasource.DataSourceWithConfigure = &usersDataSource{}
)

func NewUsersDataSource() datasource.DataSource {
	return &usersDataSource{}
}

type usersDataSource struct {
	client *Client
}

type userDataSourceModel struct {
	Email types.String `tfsdk:"email"`
	Token types.String `tfsdk:"token"`
	Name  types.String `tfsdk:"name"`
	Role  types.String `tfsdk:"role"`
}

type usersDataSourceModel struct {
	Users []userDataSourceModel `tfsdk:"users"`
}

func (d *usersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *usersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"email": schema.StringAttribute{
							Computed: true,
						},
						"token": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"role": schema.StringAttribute{
							Computed: true,
						},
					},
				},
				Computed: true,
			},
		},
	}
}

func (d *usersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state usersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	params := usersv2.NewGetUsersParams()
	out, err := d.client.V2.Users.GetUsers(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Users",
			err.Error(),
		)
		return
	}

	users := []userDataSourceModel{}

	for _, u := range out.Payload.Users {
		users = append(users, userDataSourceModel{
			Email: types.StringValue(u.Email),
			Token: types.StringValue(u.Token),
			Name:  types.StringValue(u.Name),
			Role:  types.StringValue(u.Role),
		})
	}
	state.Users = users
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *usersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*Client)
}
