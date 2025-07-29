package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	resourcereportsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/resource_reports"
)

var (
	_ datasource.DataSource              = &resourceReportColumnsDataSource{}
	_ datasource.DataSourceWithConfigure = &resourceReportColumnsDataSource{}
)

func NewResourceReportColumnsDataSource() datasource.DataSource {
	return &resourceReportColumnsDataSource{}
}

type resourceReportColumnsDataSourceModel struct {
	ResourceType types.String   `tfsdk:"resource_type"`
	Columns      []types.String `tfsdk:"columns"`
}

type resourceReportColumnsDataSource struct {
	client *Client
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *resourceReportColumnsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*Client)
}

// Metadata implements datasource.DataSource.
func (d *resourceReportColumnsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_report_columns"
}

// Read implements datasource.DataSource.
func (d *resourceReportColumnsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config resourceReportColumnsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceType := config.ResourceType.ValueString()
	
	params := resourcereportsv2.NewGetResourceReportColumnsParams().WithResourceType(resourceType)
	out, err := d.client.V2.ResourceReports.GetResourceReportColumns(params, d.client.Auth)
	if err != nil {
		if e, ok := err.(*resourcereportsv2.GetResourceReportColumnsBadRequest); ok {
			handleBadRequest("Get Resource Report Columns", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Get Resource Report Columns", &resp.Diagnostics, err)
		return
	}

	var state resourceReportColumnsDataSourceModel
	state.ResourceType = types.StringValue(resourceType)
	
	// Convert response columns to terraform types
	state.Columns = make([]types.String, len(out.Payload.Columns))
	for i, column := range out.Payload.Columns {
		state.Columns[i] = types.StringValue(column)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Schema implements datasource.DataSource.
func (d *resourceReportColumnsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Data source for retrieving available columns for a specific resource type in resource reports.",
		MarkdownDescription: "Data source for retrieving available columns for a specific resource type in resource reports.",
		Attributes: map[string]schema.Attribute{
			"resource_type": schema.StringAttribute{
				Required:            true,
				Description:         "VQL resource type name (e.g., 'aws_instance', 'aws_ebs_volume'). See https://docs.vantage.sh/vql_resource_report#resource-type for available types.",
				MarkdownDescription: "VQL resource type name (e.g., 'aws_instance', 'aws_ebs_volume'). See https://docs.vantage.sh/vql_resource_report#resource-type for available types.",
			},
			"columns": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "Array of available column names for the specified resource type.",
				MarkdownDescription: "Array of available column names for the specified resource type.",
			},
		},
	}
}