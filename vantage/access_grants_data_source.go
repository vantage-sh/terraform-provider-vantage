package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	accessgrantsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/access_grants"
)

var (
	_ datasource.DataSource              = &accessGrantsDataSource{}
	_ datasource.DataSourceWithConfigure = &accessGrantsDataSource{}
)

func NewAccessGrantsDataSource() datasource.DataSource {
	return &accessGrantsDataSource{}
}

type accessGrantDataSourceModel struct {
	Token         types.String `tfsdk:"token"`
	TeamToken     types.String `tfsdk:"team_token"`
	ResourceToken types.String `tfsdk:"resource_token"`
	Access        types.String `tfsdk:"access"`
}

type accessGrantsDataSourceModel struct {
	AccessGrants []accessGrantDataSourceModel `tfsdk:"access_grants"`
}

type accessGrantsDataSource struct {
	client *Client
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *accessGrantsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*Client)
}

// Metadata implements datasource.DataSource.
func (d *accessGrantsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_grants"
}

// Read implements datasource.DataSource.
func (d *accessGrantsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state accessGrantsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	params := accessgrantsv2.NewGetAccessGrantsParams()
	out, err := d.client.V2.AccessGrants.GetAccessGrants(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Access Grants",
			err.Error(),
		)
		return
	}

	accessGrants := []accessGrantDataSourceModel{}

	for _, ag := range out.Payload.AccessGrants {
		accessGrants = append(accessGrants, accessGrantDataSourceModel{
			Token:         types.StringValue(ag.Token),
			TeamToken:     types.StringValue(ag.TeamToken),
			ResourceToken: types.StringValue(ag.ResourceToken),
			Access:        types.StringValue(ag.Access),
		})
	}
	state.AccessGrants = accessGrants

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Schema implements datasource.DataSource.
func (d *accessGrantsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_grants": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"token": schema.StringAttribute{
							Computed: true,
						},
						"team_token": schema.StringAttribute{
							Computed: true,
						},
						"resource_token": schema.StringAttribute{
							Computed: true,
						},
						"access": schema.StringAttribute{
							Computed: true,
						},
					},
				},
				Computed: true,
			},
		},
	}
}
