package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_virtual_tag_configs"
	vtagv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/virtual_tags"
)

var (
	_ datasource.DataSource              = (*virtualTagConfigsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = &virtualTagConfigsDataSource{}
)

func NewVirtualTagConfigsDataSource() datasource.DataSource {
	return &virtualTagConfigsDataSource{}
}

type virtualTagConfigsDataSource struct {
	client *Client
}

type virtualTagConfigsDataSourceModel struct {
	VirtualTagConfigs []VirtualTagConfigResourceModel `tfsdk:"virtual_tag_configs"`
}

func (d *virtualTagConfigsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*Client)
}

func (d *virtualTagConfigsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_tag_configs"
}

func (d *virtualTagConfigsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_virtual_tag_configs.VirtualTagConfigsDataSourceSchema(ctx)
}

func (d *virtualTagConfigsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data virtualTagConfigsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	params := vtagv2.NewGetVirtualTagConfigsParams()
	apiRes, err := d.client.V2.VirtualTags.GetVirtualTagConfigs(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get Vantage Virtual Tag Configs",
			err.Error(),
		)
		return
	}

	vtags := make([]VirtualTagConfigResourceModel, 0, len(apiRes.Payload.VirtualTagConfigs))
	for _, element := range apiRes.Payload.VirtualTagConfigs {
		model := VirtualTagConfigResourceModel{}
		diag := model.applyPayload(ctx, element)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
		vtags = append(vtags, model)
	}

	data.VirtualTagConfigs = vtags

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
