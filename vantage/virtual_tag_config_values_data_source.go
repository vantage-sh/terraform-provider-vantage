package vantage

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_virtual_tag_config_values"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	tagsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/virtual_tags"
)

var (
	_ datasource.DataSource              = (*virtualTagConfigValuesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*virtualTagConfigValuesDataSource)(nil)
)

type virtualTagConfigValuesDataSource struct {
	client *Client
}

type virtualTagConfigValuesDataSourceModel datasource_virtual_tag_config_values.VirtualTagConfigValuesModel

func NewVirtualTagConfigValuesDataSource() datasource.DataSource {
	return &virtualTagConfigValuesDataSource{}
}

func (d *virtualTagConfigValuesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_tag_config_values"
}

func (d *virtualTagConfigValuesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_virtual_tag_config_values.VirtualTagConfigValuesDataSourceSchema(ctx)
	resp.Schema.MarkdownDescription = "Reads all values for a virtual tag config, in match-precedence order. Use Terraform expressions to find values by name, token, or other attributes."
	resp.Schema.Attributes["virtual_tag_config_token"] = schema.StringAttribute{
		Required:            true,
		MarkdownDescription: "The token of the virtual tag config whose values should be returned.",
	}
}

func (d *virtualTagConfigValuesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *Client, got: %T.", req.ProviderData))
		return
	}
	d.client = client
}

func (d *virtualTagConfigValuesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data virtualTagConfigValuesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	params := tagsv2.NewGetVirtualTagConfigParams().WithToken(data.VirtualTagConfigToken.ValueString()).WithContext(ctx)
	out, err := d.client.V2.VirtualTags.GetVirtualTagConfig(params, d.client.Auth)
	if err != nil {
		handleError("Read Virtual Tag Config Values Data Source", &resp.Diagnostics, err)
		return
	}
	values := make([]datasource_virtual_tag_config_values.ValuesValue, 0, len(out.Payload.Values))
	for _, payload := range out.Payload.Values {
		value, diags := buildVirtualTagConfigDataSourceValue(ctx, payload)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		values = append(values, value)
	}
	data.Token = types.StringValue(out.Payload.Token)
	var diags diag.Diagnostics
	data.Values, diags = types.ListValueFrom(ctx, datasource_virtual_tag_config_values.ValuesValue{}.Type(ctx), values)
	resp.Diagnostics.Append(diags...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	}
}

func buildVirtualTagConfigDataSourceValue(ctx context.Context, payload *modelsv2.VirtualTagConfigValue) (datasource_virtual_tag_config_values.ValuesValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	attributes := map[string]attr.Value{
		"token": types.StringValue(payload.Token), "id": types.StringValue(payload.Token),
		"filter": types.StringPointerValue(payload.Filter), "name": types.StringPointerValue(payload.Name),
		"business_metric_token": types.StringPointerValue(payload.BusinessMetricToken),
		"display_name":          types.StringPointerValue(payload.DisplayName), "label_key": types.StringPointerValue(payload.LabelKey),
		"cost_metric": types.ObjectNull(datasource_virtual_tag_config_values.CostMetricValue{}.AttributeTypes(ctx)),
	}
	labelValues, conversionDiags := types.ListValueFrom(ctx, types.StringType, payload.LabelValues)
	diags.Append(conversionDiags...)
	attributes["label_values"] = labelValues
	labelFilters, conversionDiags := types.MapValueFrom(ctx, types.ListType{ElemType: types.StringType}, payload.LabelFilters)
	diags.Append(conversionDiags...)
	attributes["label_filters"] = labelFilters
	if payload.CostMetric != nil {
		aggregation := types.ObjectNull(datasource_virtual_tag_config_values.AggregationValue{}.AttributeTypes(ctx))
		if payload.CostMetric.Aggregation != nil {
			value, conversionDiags := datasource_virtual_tag_config_values.NewAggregationValue(
				datasource_virtual_tag_config_values.AggregationValue{}.AttributeTypes(ctx),
				map[string]attr.Value{"tag": types.StringPointerValue(payload.CostMetric.Aggregation.Tag)},
			)
			diags.Append(conversionDiags...)
			aggregation, conversionDiags = value.ToObjectValue(ctx)
			diags.Append(conversionDiags...)
		}
		metric, conversionDiags := datasource_virtual_tag_config_values.NewCostMetricValue(
			datasource_virtual_tag_config_values.CostMetricValue{}.AttributeTypes(ctx),
			map[string]attr.Value{"filter": types.StringPointerValue(payload.CostMetric.Filter), "aggregation": aggregation},
		)
		diags.Append(conversionDiags...)
		attributes["cost_metric"], conversionDiags = metric.ToObjectValue(ctx)
		diags.Append(conversionDiags...)
	}
	percentages := make([]datasource_virtual_tag_config_values.PercentagesValue, 0, len(payload.Percentages))
	for _, percentage := range payload.Percentages {
		value, conversionDiags := datasource_virtual_tag_config_values.NewPercentagesValue(
			datasource_virtual_tag_config_values.PercentagesValue{}.AttributeTypes(ctx),
			map[string]attr.Value{"pct": types.NumberValue(big.NewFloat(percentage.Pct)), "value": types.StringValue(percentage.Value)},
		)
		diags.Append(conversionDiags...)
		percentages = append(percentages, value)
	}
	dateRanges := make([]datasource_virtual_tag_config_values.DateRangesValue, 0, len(payload.DateRanges))
	for _, dateRange := range payload.DateRanges {
		value, conversionDiags := datasource_virtual_tag_config_values.NewDateRangesValue(
			datasource_virtual_tag_config_values.DateRangesValue{}.AttributeTypes(ctx),
			map[string]attr.Value{"start_date": types.StringPointerValue(dateRange.StartDate), "end_date": types.StringPointerValue(dateRange.EndDate)},
		)
		diags.Append(conversionDiags...)
		dateRanges = append(dateRanges, value)
	}
	transforms := make([]datasource_virtual_tag_config_values.LabelTransformsValue, 0, len(payload.LabelTransforms))
	for _, transform := range payload.LabelTransforms {
		index := types.Int64Null()
		if transform.Index != nil {
			index = types.Int64Value(int64(*transform.Index))
		}
		value, conversionDiags := datasource_virtual_tag_config_values.NewLabelTransformsValue(
			datasource_virtual_tag_config_values.LabelTransformsValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"type": types.StringValue(transform.Type), "delimiter": types.StringPointerValue(transform.Delimiter),
				"index": index, "template": types.StringPointerValue(transform.Template),
			},
		)
		diags.Append(conversionDiags...)
		transforms = append(transforms, value)
	}
	if diags.HasError() {
		return datasource_virtual_tag_config_values.NewValuesValueUnknown(), diags
	}
	attributes["percentages"], conversionDiags = types.ListValueFrom(ctx, datasource_virtual_tag_config_values.PercentagesValue{}.Type(ctx), percentages)
	diags.Append(conversionDiags...)
	attributes["date_ranges"], conversionDiags = types.ListValueFrom(ctx, datasource_virtual_tag_config_values.DateRangesValue{}.Type(ctx), dateRanges)
	diags.Append(conversionDiags...)
	attributes["label_transforms"], conversionDiags = types.ListValueFrom(ctx, datasource_virtual_tag_config_values.LabelTransformsValue{}.Type(ctx), transforms)
	diags.Append(conversionDiags...)
	value, conversionDiags := datasource_virtual_tag_config_values.NewValuesValue(datasource_virtual_tag_config_values.ValuesValue{}.AttributeTypes(ctx), attributes)
	diags.Append(conversionDiags...)
	return value, diags
}
