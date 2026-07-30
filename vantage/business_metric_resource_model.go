package vantage

import (
	"context"
	"fmt"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_business_metrics"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_business_metric"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type BusinessMetricPayloadApplier interface {
	SetTitle(title types.String)
	SetId(id types.String)
	SetToken(token types.String)
	SetCreatedByToken(createdByToken types.String)
	SetCostReportTokensWithMetadata(costReportTokens types.List)
	SetImportType(importType types.String)
	SetIntegrationToken(integrationToken types.String)
	SetCloudwatchFields(cloudwatchFields resource_business_metric.CloudwatchFieldsValue)
	SetDatadogMetricFields(datadogMetricFields resource_business_metric.DatadogMetricFieldsValue)
	SetSnowflakeMetricFields(snowflakeMetricFields resource_business_metric.SnowflakeMetricFieldsValue)
}

type businessMetricResourceModel resource_business_metric.BusinessMetricModel

type businessMetricDataSourceValue datasource_business_metrics.BusinessMetricsValue
type businessMetricResourceModelValue struct {
	Amount types.Float64 `tfsdk:"amount"`
	Date   types.String  `tfsdk:"date"`
	Label  types.String  `tfsdk:"label"`
}
type businessMetricResourceModelCostReportToken struct {
	CostReportToken types.String `tfsdk:"cost_report_token"`
	UnitScale       types.String `tfsdk:"unit_scale"`
	CalculationType types.String `tfsdk:"calculation_type"`
	Label           types.String `tfsdk:"label"`
	LabelFilter     types.List   `tfsdk:"label_filter"`
	LabelFilters    types.Map    `tfsdk:"label_filters"`
}

func resourceCostReportTokenAttrTypes(ctx context.Context) map[string]attr.Type {
	return resource_business_metric.CostReportTokensWithMetadataValue{}.AttributeTypes(ctx)
}

func resourceCostReportTokenListType(ctx context.Context) attr.Type {
	return resource_business_metric.CostReportTokensWithMetadataValue{}.Type(ctx)
}

func dataSourceCostReportTokenListType(ctx context.Context) attr.Type {
	return datasource_business_metrics.CostReportTokensWithMetadataValue{}.Type(ctx)
}

func emptyLabelFiltersMap() types.Map {
	return types.MapNull(types.ListType{ElemType: types.StringType})
}

func labelFiltersToAPI(ctx context.Context, v types.Map, diags *diag.Diagnostics) map[string][]string {
	if v.IsNull() || v.IsUnknown() {
		return map[string][]string{}
	}

	out := map[string][]string{}
	diags.Append(v.ElementsAs(ctx, &out, false)...)
	if diags.HasError() {
		return nil
	}
	return out
}

func labelFiltersMapFromAPI(ctx context.Context, raw interface{}) (types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ListType{ElemType: types.StringType}
	if raw == nil {
		return types.MapNull(elemType), diags
	}

	normalized, ok := normalizeLabelFilters(raw)
	if !ok {
		diags.AddWarning(
			"Unable to parse label_filters from API",
			fmt.Sprintf("Unexpected label_filters type %T; leaving attribute null.", raw),
		)
		return types.MapNull(elemType), diags
	}
	if len(normalized) == 0 {
		return types.MapNull(elemType), diags
	}

	m, d := types.MapValueFrom(ctx, elemType, normalized)
	diags.Append(d...)
	return m, diags
}

func normalizeLabelFilters(raw interface{}) (map[string][]string, bool) {
	switch v := raw.(type) {
	case map[string][]string:
		return v, true
	case map[string]interface{}:
		out := make(map[string][]string, len(v))
		for key, val := range v {
			switch values := val.(type) {
			case []string:
				out[key] = values
			case []interface{}:
				strs := make([]string, 0, len(values))
				for _, item := range values {
					s, ok := item.(string)
					if !ok {
						return nil, false
					}
					strs = append(strs, s)
				}
				out[key] = strs
			default:
				return nil, false
			}
		}
		return out, true
	default:
		return nil, false
	}
}

func (m *businessMetricResourceModel) SetTitle(title types.String) {
	m.Title = title
}

func (m *businessMetricResourceModel) SetToken(token types.String) {
	m.Token = token
}

func (m *businessMetricResourceModel) SetId(id types.String) {
	m.Id = id
}

func (m *businessMetricResourceModel) SetCreatedByToken(createdByToken types.String) {
	m.CreatedByToken = createdByToken
}

func (m *businessMetricResourceModel) SetCostReportTokensWithMetadata(costReportTokens types.List) {
	m.CostReportTokensWithMetadata = costReportTokens
}

func (m *businessMetricResourceModel) SetImportType(importType types.String) {
	m.ImportType = importType
}

func (m *businessMetricResourceModel) SetIntegrationToken(integrationToken types.String) {
	m.IntegrationToken = integrationToken
}

func (m *businessMetricResourceModel) SetCloudwatchFields(cloudwatchFields resource_business_metric.CloudwatchFieldsValue) {
	m.CloudwatchFields = cloudwatchFields
}

func (m *businessMetricResourceModel) SetDatadogMetricFields(datadogMetricFields resource_business_metric.DatadogMetricFieldsValue) {
	m.DatadogMetricFields = datadogMetricFields
}

func (m *businessMetricResourceModel) SetSnowflakeMetricFields(snowflakeMetricFields resource_business_metric.SnowflakeMetricFieldsValue) {
	m.SnowflakeMetricFields = snowflakeMetricFields
}

func (m *businessMetricDataSourceValue) SetTitle(title types.String) {
	m.Title = title
}

func (m *businessMetricDataSourceValue) SetToken(token types.String) {
	m.Token = token
}

func (m *businessMetricDataSourceValue) SetId(id types.String) {
	// m.Id = id
}

func (m *businessMetricDataSourceValue) SetCreatedByToken(createdByToken types.String) {
	m.CreatedByToken = createdByToken
}

func (m *businessMetricDataSourceValue) SetCostReportTokensWithMetadata(costReportTokens types.List) {
	m.CostReportTokensWithMetadata = costReportTokens
}

func (m *businessMetricDataSourceValue) SetImportType(importType types.String) {
	m.ImportType = importType
}

func (m *businessMetricDataSourceValue) SetIntegrationToken(integrationToken types.String) {
	m.IntegrationToken = integrationToken
}

func (m *businessMetricDataSourceValue) SetCloudwatchFields(cloudwatchFields resource_business_metric.CloudwatchFieldsValue) {
	ctx := context.Background()

	emptyDimensions, _ := types.ListValue(
		types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"name":  types.StringType,
				"value": types.StringType,
			},
		},
		[]attr.Value{},
	)

	labelDimension := cloudwatchFields.LabelDimension
	if labelDimension.IsNull() || labelDimension.IsUnknown() {
		labelDimension = types.StringValue("")
	}

	metricName := cloudwatchFields.MetricName
	if metricName.IsNull() || metricName.IsUnknown() {
		metricName = types.StringValue("")
	}

	namespace := cloudwatchFields.Namespace
	if namespace.IsNull() || namespace.IsUnknown() {
		namespace = types.StringValue("")
	}

	region := cloudwatchFields.Region
	if region.IsNull() || region.IsUnknown() {
		region = types.StringValue("")
	}

	stat := cloudwatchFields.Stat
	if stat.IsNull() || stat.IsUnknown() {
		stat = types.StringValue("")
	}

	// Use the configured dimensions if available, or empty list if not
	dimensions := emptyDimensions
	if !cloudwatchFields.Dimensions.IsNull() && !cloudwatchFields.Dimensions.IsUnknown() {
		dims, dimsErr := types.ListValueFrom(
			ctx,
			types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"name":  types.StringType,
					"value": types.StringType,
				},
			},
			cloudwatchFields.Dimensions,
		)
		if dimsErr == nil {
			dimensions = dims
		}
	}

	objVal, _ := types.ObjectValue(
		map[string]attr.Type{
			"label_dimension": types.StringType,
			"metric_name":     types.StringType,
			"namespace":       types.StringType,
			"region":          types.StringType,
			"stat":            types.StringType,
			"dimensions": types.ListType{ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"name":  types.StringType,
					"value": types.StringType,
				},
			}},
		},
		map[string]attr.Value{
			"label_dimension": labelDimension,
			"metric_name":     metricName,
			"namespace":       namespace,
			"region":          region,
			"stat":            stat,
			"dimensions":      dimensions,
		},
	)
	m.CloudwatchFields = objVal
}

func (m *businessMetricDataSourceValue) SetDatadogMetricFields(datadogMetricFields resource_business_metric.DatadogMetricFieldsValue) {
	query := datadogMetricFields.Query
	if query.IsNull() || query.IsUnknown() {
		query = types.StringValue("")
	}

	objVal, _ := types.ObjectValue(
		map[string]attr.Type{
			"query": types.StringType,
		},
		map[string]attr.Value{
			"query": query,
		})

	m.DatadogMetricFields = objVal
}

func (m *businessMetricDataSourceValue) SetSnowflakeMetricFields(snowflakeMetricFields resource_business_metric.SnowflakeMetricFieldsValue) {
	attrTypes := datasource_business_metrics.SnowflakeMetricFieldsValue{}.AttributeTypes(context.Background())

	if snowflakeMetricFields.IsNull() {
		m.SnowflakeMetricFields = types.ObjectNull(attrTypes)
		return
	}

	sqlQuery := snowflakeMetricFields.SqlQuery
	if sqlQuery.IsNull() || sqlQuery.IsUnknown() {
		sqlQuery = types.StringValue("")
	}

	objVal, _ := types.ObjectValue(
		attrTypes,
		map[string]attr.Value{
			"sql_query": sqlQuery,
		},
	)
	m.SnowflakeMetricFields = objVal
}

func applyPayload[T BusinessMetricPayloadApplier](ctx context.Context, m T, payload *modelsv2.BusinessMetric) diag.Diagnostics {
	m.SetTitle(types.StringValue(payload.Title))
	m.SetToken(types.StringValue(payload.Token))
	m.SetId(types.StringValue(payload.Token))
	m.SetCreatedByToken(types.StringPointerValue(payload.CreatedByToken))
	m.SetImportType(types.StringPointerValue(payload.ImportType))
	m.SetIntegrationToken(types.StringPointerValue(payload.IntegrationToken))

	tfCloudwatchFields, d := cloudwatchFieldsFromApiModel(ctx, payload.CloudwatchFields, payload.IntegrationToken)
	if d.HasError() {
		return d
	}
	m.SetCloudwatchFields(tfCloudwatchFields)

	tfDatadogMetricFields, d := datadogMetricFieldsFromApiModel(ctx, payload.DatadogMetricFields, payload.IntegrationToken)
	if d.HasError() {
		return d
	}
	m.SetDatadogMetricFields(tfDatadogMetricFields)

	tfSnowflakeMetricFields, d := snowflakeMetricFieldsFromApiModel(ctx, payload.SnowflakeMetricFields, payload.IntegrationToken)
	if d.HasError() {
		return d
	}
	m.SetSnowflakeMetricFields(tfSnowflakeMetricFields)

	if payload.CostReportTokensWithMetadata != nil {
		var costReportTokens types.List
		var tokenDiags diag.Diagnostics

		switch any(m).(type) {
		case *businessMetricDataSourceValue:
			costReportTokens, tokenDiags = costReportTokensListForDataSource(ctx, payload.CostReportTokensWithMetadata)
		default:
			costReportTokens, tokenDiags = costReportTokensListForResource(ctx, payload.CostReportTokensWithMetadata)
		}

		if tokenDiags.HasError() {
			return tokenDiags
		}

		m.SetCostReportTokensWithMetadata(costReportTokens)
	}

	return nil
}

func costReportTokensListForResource(ctx context.Context, payloadTokens []*modelsv2.AttachedCostReportForBusinessMetric) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrTypes := resourceCostReportTokenAttrTypes(ctx)
	elements := make([]attr.Value, 0, len(payloadTokens))

	for _, costReportToken := range payloadTokens {
		labelFilter, d := types.ListValueFrom(ctx, types.StringType, costReportToken.LabelFilter)
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(resourceCostReportTokenListType(ctx)), diags
		}

		labelFilters, d := labelFiltersMapFromAPI(ctx, costReportToken.LabelFilters)
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(resourceCostReportTokenListType(ctx)), diags
		}

		tokenValue, d := resource_business_metric.NewCostReportTokensWithMetadataValue(
			attrTypes,
			map[string]attr.Value{
				"cost_report_token": types.StringPointerValue(costReportToken.CostReportToken),
				"unit_scale":        types.StringValue(costReportToken.UnitScale),
				"calculation_type":  types.StringValue(costReportToken.CalculationType),
				"label":             types.StringPointerValue(costReportToken.Label),
				"label_filter":      labelFilter,
				"label_filters":     labelFilters,
			},
		)
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(resourceCostReportTokenListType(ctx)), diags
		}
		elements = append(elements, tokenValue)
	}

	list, d := types.ListValue(resourceCostReportTokenListType(ctx), elements)
	diags.Append(d...)
	return list, diags
}

func costReportTokensListForDataSource(ctx context.Context, payloadTokens []*modelsv2.AttachedCostReportForBusinessMetric) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrTypes := datasource_business_metrics.CostReportTokensWithMetadataValue{}.AttributeTypes(ctx)
	elements := make([]attr.Value, 0, len(payloadTokens))

	for _, costReportToken := range payloadTokens {
		labelFilter, d := types.ListValueFrom(ctx, types.StringType, costReportToken.LabelFilter)
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(dataSourceCostReportTokenListType(ctx)), diags
		}

		labelFilters, d := labelFiltersMapFromAPI(ctx, costReportToken.LabelFilters)
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(dataSourceCostReportTokenListType(ctx)), diags
		}

		tokenValue, d := datasource_business_metrics.NewCostReportTokensWithMetadataValue(
			attrTypes,
			map[string]attr.Value{
				"cost_report_token": types.StringPointerValue(costReportToken.CostReportToken),
				"unit_scale":        types.StringValue(costReportToken.UnitScale),
				"calculation_type":  types.StringValue(costReportToken.CalculationType),
				"label":             types.StringPointerValue(costReportToken.Label),
				"label_filter":      labelFilter,
				"label_filters":     labelFilters,
			},
		)
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(dataSourceCostReportTokenListType(ctx)), diags
		}
		elements = append(elements, tokenValue)
	}

	list, d := types.ListValue(dataSourceCostReportTokenListType(ctx), elements)
	diags.Append(d...)
	return list, diags
}

func (m *businessMetricDataSourceValue) applyPayload(ctx context.Context, payload *modelsv2.BusinessMetric) diag.Diagnostics {
	return applyPayload(ctx, m, payload)
}

func (m *businessMetricResourceModel) applyPayload(ctx context.Context, payload *modelsv2.BusinessMetric) diag.Diagnostics {
	return applyPayload(ctx, m, payload)
}

func (m *businessMetricResourceModel) toCreate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.CreateBusinessMetric {
	model := &modelsv2.CreateBusinessMetric{
		Title: m.Title.ValueStringPointer(),
	}

	if !m.CloudwatchFields.IsNull() && !m.CloudwatchFields.IsUnknown() {
		cloudwatchFields := &modelsv2.CreateBusinessMetricCloudwatchFields{
			IntegrationToken: m.CloudwatchFields.IntegrationToken.ValueString(),
			MetricName:       m.CloudwatchFields.MetricName.ValueString(),
			Namespace:        m.CloudwatchFields.Namespace.ValueString(),
			Region:           m.CloudwatchFields.Region.ValueString(),
			Stat:             m.CloudwatchFields.Stat.ValueString(),
			LabelDimension:   m.CloudwatchFields.LabelDimension.ValueString(),
		}

		if !m.CloudwatchFields.Dimensions.IsNull() && !m.CloudwatchFields.Dimensions.IsUnknown() {
			dimsLen := len(m.CloudwatchFields.Dimensions.Elements())
			if dimsLen > 0 {
				dimensions := make([]*modelsv2.CreateBusinessMetricCloudwatchFieldsDimensionsItems0, 0, dimsLen)
				var tfDimensions []resource_business_metric.DimensionsValue
				diags.Append(m.CloudwatchFields.Dimensions.ElementsAs(ctx, &tfDimensions, false)...)
				if diags.HasError() {
					return nil
				}

				for _, dim := range tfDimensions {
					dimensions = append(dimensions, &modelsv2.CreateBusinessMetricCloudwatchFieldsDimensionsItems0{
						Name:  dim.Name.ValueString(),
						Value: dim.Value.ValueString(),
					})
				}
				cloudwatchFields.Dimensions = dimensions
			}
		}

		model.CloudwatchFields = cloudwatchFields
	}

	if !m.DatadogMetricFields.IsNull() && !m.DatadogMetricFields.IsUnknown() {
		datadogMetricFields := &modelsv2.CreateBusinessMetricDatadogMetricFields{
			IntegrationToken: m.DatadogMetricFields.IntegrationToken.ValueString(),
			Query:            m.DatadogMetricFields.Query.ValueString(),
		}
		model.DatadogMetricFields = datadogMetricFields
	}

	if !m.SnowflakeMetricFields.IsNull() && !m.SnowflakeMetricFields.IsUnknown() {
		model.SnowflakeMetricFields = &modelsv2.CreateBusinessMetricSnowflakeMetricFields{
			IntegrationToken: m.SnowflakeMetricFields.IntegrationToken.ValueString(),
			SQLQuery:         m.SnowflakeMetricFields.SqlQuery.ValueString(),
		}
	}

	if !m.Values.IsNull() && !m.Values.IsUnknown() {
		tfValues := m.valuesFromTf(ctx, diags)
		if diags.HasError() {
			return nil
		}

		values := make([]*modelsv2.CreateBusinessMetricValuesItems0, 0, len(tfValues))
		for _, v := range tfValues {
			amt := v.Amount.ValueFloat64()
			t, err := time.Parse("2006-01-02", v.Date.ValueString())
			if err != nil {
				diags.AddError(fmt.Sprintf("Failed to parse date: %s", v.Date.ValueString()), err.Error())
				return nil
			}
			ts := strfmt.DateTime(t)
			label := v.Label.ValueStringPointer()

			value := modelsv2.CreateBusinessMetricValuesItems0{
				Amount: &amt,
				Date:   &ts,
				Label:  label,
			}

			values = append(values, &value)
		}

		model.Values = values
	}

	if !m.CostReportTokensWithMetadata.IsNull() && !m.CostReportTokensWithMetadata.IsUnknown() {
		tfCostReportTokens := m.costReportTokensFromTf(ctx, diags)
		if diags.HasError() {
			return nil
		}

		costReportTokens := make([]*modelsv2.CreateBusinessMetricCostReportTokensWithMetadataItems0, 0, len(tfCostReportTokens))
		for _, v := range tfCostReportTokens {
			tfLabelFilter := []string{}
			if !v.LabelFilter.IsNull() && !v.LabelFilter.IsUnknown() {
				tfLabelFilter = make([]string, 0, len(v.LabelFilter.Elements()))
				diags.Append(v.LabelFilter.ElementsAs(ctx, &tfLabelFilter, false)...)
				if diags.HasError() {
					return nil
				}
			}

			costReportToken := &modelsv2.CreateBusinessMetricCostReportTokensWithMetadataItems0{
				CostReportToken: v.CostReportToken.ValueStringPointer(),
				UnitScale:       v.UnitScale.ValueStringPointer(),
				CalculationType: calculationTypePointer(v.CalculationType),
				Label:           costReportAttachmentLabelForAPI(v.Label),
				LabelFilter:     tfLabelFilter,
				LabelFilters:    labelFiltersToAPI(ctx, v.LabelFilters, diags),
			}
			if diags.HasError() {
				return nil
			}
			costReportTokens = append(costReportTokens, costReportToken)
		}
		model.CostReportTokensWithMetadata = costReportTokens

	}

	if !m.ForecastedValues.IsNull() && !m.ForecastedValues.IsUnknown() {
		tfForecastedValues := m.forecastedValuesFromTf(ctx, diags)
		if diags.HasError() {
			return nil
		}

		forecastedValues := make([]*modelsv2.CreateBusinessMetricForecastedValuesItems0, 0, len(tfForecastedValues))
		for _, v := range tfForecastedValues {
			amt := v.Amount.ValueFloat64()
			t, err := time.Parse("2006-01-02", v.Date.ValueString())
			if err != nil {
				diags.AddError(fmt.Sprintf("Failed to parse forecasted value date: %s", v.Date.ValueString()), err.Error())
				return nil
			}
			ts := strfmt.DateTime(t)
			label := v.Label.ValueStringPointer()

			value := modelsv2.CreateBusinessMetricForecastedValuesItems0{
				Amount: &amt,
				Date:   &ts,
				Label:  label,
			}

			forecastedValues = append(forecastedValues, &value)
		}

		model.ForecastedValues = forecastedValues
	}

	return model
}

func (m *businessMetricResourceModel) toUpdate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.UpdateBusinessMetric {
	if m.Token.IsNull() || m.Token.IsUnknown() {
		diags.AddError("Token is required for update", "")
		return nil
	}

	model := &modelsv2.UpdateBusinessMetric{}

	// TODO need IsUnknown check here?
	if !m.Title.IsNull() {
		model.Title = m.Title.ValueString()
	}

	if !m.CloudwatchFields.IsNull() && !m.CloudwatchFields.IsUnknown() {
		cloudwatchFields := &modelsv2.UpdateBusinessMetricCloudwatchFields{
			IntegrationToken: m.CloudwatchFields.IntegrationToken.ValueString(),
			MetricName:       m.CloudwatchFields.MetricName.ValueString(),
			Namespace:        m.CloudwatchFields.Namespace.ValueString(),
			Region:           m.CloudwatchFields.Region.ValueString(),
			Stat:             m.CloudwatchFields.Stat.ValueString(),
			LabelDimension:   m.CloudwatchFields.LabelDimension.ValueString(),
		}

		if !m.CloudwatchFields.Dimensions.IsNull() && !m.CloudwatchFields.Dimensions.IsUnknown() {
			dimsLen := len(m.CloudwatchFields.Dimensions.Elements())
			if dimsLen > 0 {
				dimensions := make([]*modelsv2.UpdateBusinessMetricCloudwatchFieldsDimensionsItems0, 0, dimsLen)
				var tfDimensions []resource_business_metric.DimensionsValue
				diags.Append(m.CloudwatchFields.Dimensions.ElementsAs(ctx, &tfDimensions, false)...)
				if diags.HasError() {
					return nil
				}

				for _, dim := range tfDimensions {
					dimensions = append(dimensions, &modelsv2.UpdateBusinessMetricCloudwatchFieldsDimensionsItems0{
						Name:  dim.Name.ValueString(),
						Value: dim.Value.ValueString(),
					})
				}
				cloudwatchFields.Dimensions = dimensions
			}
		}

		model.CloudwatchFields = cloudwatchFields
	}

	if !m.DatadogMetricFields.IsNull() && !m.DatadogMetricFields.IsUnknown() {
		datadogMetricFields := &modelsv2.UpdateBusinessMetricDatadogMetricFields{
			IntegrationToken: m.DatadogMetricFields.IntegrationToken.ValueString(),
			Query:            m.DatadogMetricFields.Query.ValueString(),
		}
		model.DatadogMetricFields = datadogMetricFields
	}

	if !m.SnowflakeMetricFields.IsNull() && !m.SnowflakeMetricFields.IsUnknown() {
		model.SnowflakeMetricFields = &modelsv2.UpdateBusinessMetricSnowflakeMetricFields{
			IntegrationToken: m.SnowflakeMetricFields.IntegrationToken.ValueString(),
			SQLQuery:         m.SnowflakeMetricFields.SqlQuery.ValueString(),
		}
	}

	if !m.Values.IsNull() && !m.Values.IsUnknown() {
		tfValues := m.valuesFromTf(ctx, diags)
		if diags.HasError() {
			return nil
		}

		values := make([]*modelsv2.UpdateBusinessMetricValuesItems0, 0, len(tfValues))
		for _, v := range tfValues {
			amt := v.Amount.ValueFloat64()
			t, err := time.Parse("2006-01-02", v.Date.ValueString())
			if err != nil {
				diags.AddError(fmt.Sprintf("Failed to parse date: %s", v.Date.ValueString()), err.Error())
				return nil
			}
			ts := strfmt.DateTime(t)
			label := v.Label.ValueStringPointer()

			value := modelsv2.UpdateBusinessMetricValuesItems0{
				Amount: &amt,
				Date:   &ts,
				Label:  label,
			}

			values = append(values, &value)
		}

		model.Values = values
	}

	if !m.CostReportTokensWithMetadata.IsNull() && !m.CostReportTokensWithMetadata.IsUnknown() {
		tfCostReportTokens := m.costReportTokensFromTf(ctx, diags)
		if diags.HasError() {
			return nil
		}

		costReportTokens := make([]*modelsv2.UpdateBusinessMetricCostReportTokensWithMetadataItems0, 0, len(tfCostReportTokens))
		for _, v := range tfCostReportTokens {
			tfLabelFilter := []string{}
			if !v.LabelFilter.IsNull() && !v.LabelFilter.IsUnknown() {
				tfLabelFilter = make([]string, 0, len(v.LabelFilter.Elements()))
				diags.Append(v.LabelFilter.ElementsAs(ctx, &tfLabelFilter, false)...)
				if diags.HasError() {
					return nil
				}
			}
			costReportToken := &modelsv2.UpdateBusinessMetricCostReportTokensWithMetadataItems0{
				CostReportToken: v.CostReportToken.ValueStringPointer(),
				UnitScale:       v.UnitScale.ValueStringPointer(),
				CalculationType: calculationTypePointer(v.CalculationType),
				Label:           costReportAttachmentLabelForAPI(v.Label),
				LabelFilter:     tfLabelFilter,
				LabelFilters:    labelFiltersToAPI(ctx, v.LabelFilters, diags),
			}
			if diags.HasError() {
				return nil
			}
			costReportTokens = append(costReportTokens, costReportToken)
		}
		model.CostReportTokensWithMetadata = costReportTokens
	}

	if !m.ForecastedValues.IsNull() && !m.ForecastedValues.IsUnknown() {
		tfForecastedValues := m.forecastedValuesFromTf(ctx, diags)
		if diags.HasError() {
			return nil
		}

		forecastedValues := make([]*modelsv2.UpdateBusinessMetricForecastedValuesItems0, 0, len(tfForecastedValues))
		for _, v := range tfForecastedValues {
			amt := v.Amount.ValueFloat64()
			t, err := time.Parse("2006-01-02", v.Date.ValueString())
			if err != nil {
				diags.AddError(fmt.Sprintf("Failed to parse forecasted value date: %s", v.Date.ValueString()), err.Error())
				return nil
			}
			ts := strfmt.DateTime(t)
			label := v.Label.ValueStringPointer()

			value := modelsv2.UpdateBusinessMetricForecastedValuesItems0{
				Amount: &amt,
				Date:   &ts,
				Label:  label,
			}

			forecastedValues = append(forecastedValues, &value)
		}

		model.ForecastedValues = forecastedValues
	}

	return model
}

func (m *businessMetricResourceModel) valuesFromTf(ctx context.Context, diags *diag.Diagnostics) []*businessMetricResourceModelValue {
	values := make([]*businessMetricResourceModelValue, 0, len(m.Values.Elements()))
	if diag := m.Values.ElementsAs(ctx, &values, false); diag.HasError() {
		diags.Append(diag...)
		return nil
	}
	return values
}

func (m *businessMetricResourceModel) forecastedValuesFromTf(ctx context.Context, diags *diag.Diagnostics) []*businessMetricResourceModelValue {
	values := make([]*businessMetricResourceModelValue, 0, len(m.ForecastedValues.Elements()))
	if diag := m.ForecastedValues.ElementsAs(ctx, &values, false); diag.HasError() {
		diags.Append(diag...)
		return nil
	}
	return values
}

func (m *businessMetricResourceModel) costReportTokensFromTf(ctx context.Context, diags *diag.Diagnostics) []*businessMetricResourceModelCostReportToken {
	costReportTokens := make([]*businessMetricResourceModelCostReportToken, 0, len(m.CostReportTokensWithMetadata.Elements()))
	if diag := m.CostReportTokensWithMetadata.ElementsAs(ctx, &costReportTokens, false); diag.HasError() {
		diags.Append(diag...)
		return nil
	}
	return costReportTokens
}

// assignCostReportTokens reorders the cost report tokens to match the original plan order
// and fills in any computed values from the API response
func assignCostReportTokens(ctx context.Context, data *businessMetricResourceModel, tfCostReportTokens types.List, diags *diag.Diagnostics) {
	// Get the original plan order
	planTokens := make([]*businessMetricResourceModelCostReportToken, 0, len(tfCostReportTokens.Elements()))
	if d := tfCostReportTokens.ElementsAs(ctx, &planTokens, false); d.HasError() {
		diags.Append(d...)
		return
	}

	// Get the API response values
	apiTokens := make([]*businessMetricResourceModelCostReportToken, 0, len(data.CostReportTokensWithMetadata.Elements()))
	if d := data.CostReportTokensWithMetadata.ElementsAs(ctx, &apiTokens, false); d.HasError() {
		diags.Append(d...)
		return
	}

	// Build a map of API tokens by cost_report_token for quick lookup
	apiTokenMap := make(map[string]*businessMetricResourceModelCostReportToken)
	for _, t := range apiTokens {
		apiTokenMap[t.CostReportToken.ValueString()] = t
	}

	// Reorder to match plan order, using API values for computed fields
	orderedTokens := make([]businessMetricResourceModelCostReportToken, 0, len(planTokens))
	for _, planToken := range planTokens {
		tokenKey := planToken.CostReportToken.ValueString()
		if apiToken, ok := apiTokenMap[tokenKey]; ok {
			// Handle label_filter: if API returns null/unknown but plan had a known value, use plan's value
			// This ensures empty lists stay as empty lists rather than becoming null
			labelFilter := apiToken.LabelFilter
			if (labelFilter.IsNull() || labelFilter.IsUnknown()) && !planToken.LabelFilter.IsNull() && !planToken.LabelFilter.IsUnknown() {
				labelFilter = planToken.LabelFilter
			}
			// If still null or unknown, ensure we use an empty list for consistency
			if labelFilter.IsNull() || labelFilter.IsUnknown() {
				labelFilter, _ = types.ListValueFrom(ctx, types.StringType, []string{})
			}

			calculationType := apiToken.CalculationType
			if (calculationType.IsNull() || calculationType.IsUnknown()) && !planToken.CalculationType.IsNull() && !planToken.CalculationType.IsUnknown() {
				calculationType = planToken.CalculationType
			}
			if calculationType.IsNull() || calculationType.IsUnknown() {
				calculationType = types.StringValue("unit_cost")
			}

			label := apiToken.Label

			// Trust the API for label_filters so omitting them from config clears
			// prior filters instead of re-sticking planned values into state.
			labelFilters := apiToken.LabelFilters
			if labelFilters.IsNull() || labelFilters.IsUnknown() {
				labelFilters = emptyLabelFiltersMap()
			}

			// Use the plan's cost_report_token but take computed values from API
			orderedTokens = append(orderedTokens, businessMetricResourceModelCostReportToken{
				CostReportToken: planToken.CostReportToken,
				UnitScale:       apiToken.UnitScale,
				CalculationType: calculationType,
				Label:           label,
				LabelFilter:     labelFilter,
				LabelFilters:    labelFilters,
			})
			delete(apiTokenMap, tokenKey)
		} else {
			// Plan contains a cost_report_token that is missing from the API response.
			// Emit a warning and preserve the planned token to avoid silent state drift.
			diags.AddWarning(
				"Missing cost_report_token in API response",
				fmt.Sprintf("A cost_report_token present in the plan (%q) was not returned by the API; preserving the planned value.", tokenKey),
			)
			orderedTokens = append(orderedTokens, *planToken)
		}
	}

	// Append any tokens from API that weren't in the plan (shouldn't happen normally)
	for _, apiToken := range apiTokenMap {
		orderedTokens = append(orderedTokens, *apiToken)
	}

	elements := make([]attr.Value, 0, len(orderedTokens))
	attrTypes := resourceCostReportTokenAttrTypes(ctx)
	for _, token := range orderedTokens {
		tokenValue, d := resource_business_metric.NewCostReportTokensWithMetadataValue(
			attrTypes,
			map[string]attr.Value{
				"cost_report_token": token.CostReportToken,
				"unit_scale":        token.UnitScale,
				"calculation_type":  token.CalculationType,
				"label":             token.Label,
				"label_filter":      token.LabelFilter,
				"label_filters":     token.LabelFilters,
			},
		)
		if d.HasError() {
			diags.Append(d...)
			return
		}
		elements = append(elements, tokenValue)
	}

	newList, d := types.ListValue(resourceCostReportTokenListType(ctx), elements)
	if d.HasError() {
		diags.Append(d...)
		return
	}

	data.CostReportTokensWithMetadata = newList
}

func datadogMetricFieldsFromApiModel(ctx context.Context, apiFields *modelsv2.DatadogMetricFields, integrationToken *string) (resource_business_metric.DatadogMetricFieldsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	if apiFields == nil {
		return resource_business_metric.NewDatadogMetricFieldsValueNull(), diags
	}

	tfValue, d := resource_business_metric.NewDatadogMetricFieldsValue(
		map[string]attr.Type{
			"query":             types.StringType,
			"integration_token": types.StringType,
		},
		map[string]attr.Value{
			"query":             types.StringValue(apiFields.Query),
			"integration_token": types.StringPointerValue(integrationToken),
		},
	)
	diags.Append(d...)

	return tfValue, diags
}

func snowflakeMetricFieldsFromApiModel(ctx context.Context, apiFields *modelsv2.SnowflakeMetricFields, integrationToken *string) (resource_business_metric.SnowflakeMetricFieldsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	if apiFields == nil {
		return resource_business_metric.NewSnowflakeMetricFieldsValueNull(), diags
	}

	tfValue, d := resource_business_metric.NewSnowflakeMetricFieldsValue(
		resource_business_metric.SnowflakeMetricFieldsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"sql_query":         types.StringValue(apiFields.SQLQuery),
			"integration_token": types.StringPointerValue(integrationToken),
		},
	)
	diags.Append(d...)

	return tfValue, diags
}

func cloudwatchFieldsFromApiModel(ctx context.Context, apiFields *modelsv2.CloudwatchFields, integrationToken *string) (resource_business_metric.CloudwatchFieldsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	if apiFields == nil {
		return resource_business_metric.NewCloudwatchFieldsValueNull(), diags
	}

	dimensionAttrTypes := map[string]attr.Type{
		"name":  types.StringType,
		"value": types.StringType,
	}

	tfDimensionObjects := []attr.Value{}

	for _, apiDimension := range apiFields.Dimensions {
		if apiDimension == nil {
			continue
		}
		dimensionObjectValue, d := types.ObjectValue(
			dimensionAttrTypes,
			map[string]attr.Value{
				"name":  types.StringValue(apiDimension.Name),
				"value": types.StringValue(apiDimension.Value),
			},
		)
		diags.Append(d...)
		if diags.HasError() {
			return resource_business_metric.NewCloudwatchFieldsValueUnknown(), diags
		}
		tfDimensionObjects = append(tfDimensionObjects, dimensionObjectValue)
	}

	dimensionsListValue, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: dimensionAttrTypes}, tfDimensionObjects)
	diags.Append(d...)
	if diags.HasError() {
		return resource_business_metric.NewCloudwatchFieldsValueUnknown(), diags
	}

	cloudwatchAttrTypes := map[string]attr.Type{
		"stat":              types.StringType,
		"metric_name":       types.StringType,
		"namespace":         types.StringType,
		"region":            types.StringType,
		"label_dimension":   types.StringType,
		"dimensions":        types.ListType{ElemType: types.ObjectType{AttrTypes: dimensionAttrTypes}},
		"integration_token": types.StringType,
	}

	tfValue, d := resource_business_metric.NewCloudwatchFieldsValue(
		cloudwatchAttrTypes,
		map[string]attr.Value{
			"stat":              types.StringValue(apiFields.Stat),
			"metric_name":       types.StringValue(apiFields.MetricName),
			"namespace":         types.StringValue(apiFields.Namespace),
			"region":            types.StringValue(apiFields.Region),
			"label_dimension":   types.StringPointerValue(apiFields.LabelDimension),
			"dimensions":        dimensionsListValue,
			"integration_token": types.StringPointerValue(integrationToken),
		},
	)
	diags.Append(d...)

	if diags.HasError() {
		return resource_business_metric.NewCloudwatchFieldsValueUnknown(), diags
	}

	return tfValue, diags
}

func calculationTypePointer(v types.String) *string {
	if v.IsNull() || v.IsUnknown() || v.ValueString() == "" {
		unitCost := "unit_cost"
		return &unitCost
	}
	return v.ValueStringPointer()
}

// costReportAttachmentLabelForAPI returns the attachment label for create/update
// payloads. Null or unknown (config omitted the attribute) become "" so the
// SDK's omitempty tag drops the JSON key entirely.
func costReportAttachmentLabelForAPI(v types.String) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return v.ValueString()
}
