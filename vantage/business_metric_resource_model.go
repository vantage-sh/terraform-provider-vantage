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
	SetToken(token types.String)
	SetCreatedByToken(createdByToken types.String)
	SetCostReportTokensWithMetadata(costReportTokens types.List)
	SetImportType(importType types.String)
	SetIntegrationToken(integrationToken types.String)
	SetCloudwatchFields(cloudwatchFields resource_business_metric.CloudwatchFieldsValue)
	SetDatadogMetricFields(datadogMetricFields resource_business_metric.DatadogMetricFieldsValue)
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
	LabelFilter     types.List   `tfsdk:"label_filter"`
}

func (m *businessMetricResourceModel) SetTitle(title types.String) {
	m.Title = title
}

func (m *businessMetricResourceModel) SetToken(token types.String) {
	m.Token = token
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

func (m *businessMetricDataSourceValue) SetTitle(title types.String) {
	m.Title = title
}

func (m *businessMetricDataSourceValue) SetToken(token types.String) {
	m.Token = token
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

func applyPayload[T BusinessMetricPayloadApplier](ctx context.Context, m T, payload *modelsv2.BusinessMetric) diag.Diagnostics {
	m.SetTitle(types.StringValue(payload.Title))
	m.SetToken(types.StringValue(payload.Token))
	m.SetCreatedByToken(types.StringValue(payload.CreatedByToken))
	m.SetImportType(types.StringValue(payload.ImportType))
	m.SetIntegrationToken(types.StringValue(payload.IntegrationToken))

	tfCloudwatchFields, diag := cloudwatchFieldsFromApiModel(ctx, payload.CloudwatchFields, payload.IntegrationToken)
	if diag.HasError() {
		return diag
	}
	m.SetCloudwatchFields(tfCloudwatchFields)

	tfDatadogMetricFields, diag := datadogMetricFieldsFromApiModel(ctx, payload.DatadogMetricFields, payload.IntegrationToken)
	if diag.HasError() {
		return diag
	}
	m.SetDatadogMetricFields(tfDatadogMetricFields)

	if payload.CostReportTokensWithMetadata != nil {
		tfCostReportTokens := []businessMetricResourceModelCostReportToken{}
		for _, costReportToken := range payload.CostReportTokensWithMetadata {
			labelFilter, diag := types.ListValueFrom(ctx, types.StringType, costReportToken.LabelFilter)
			if diag.HasError() {
				return diag
			}
			tfCostReportTokens = append(tfCostReportTokens, businessMetricResourceModelCostReportToken{
				CostReportToken: types.StringValue(costReportToken.CostReportToken),
				UnitScale:       types.StringValue(costReportToken.UnitScale),
				LabelFilter:     labelFilter,
			})
		}

		costReportTokens, diag := types.ListValueFrom(
			ctx,
			types.ObjectType{AttrTypes: map[string]attr.Type{
				"cost_report_token": types.StringType,
				"unit_scale":        types.StringType,
				"label_filter":      types.ListType{ElemType: types.StringType},
			}},
			tfCostReportTokens,
		)

		if diag.HasError() {
			return diag
		}

		m.SetCostReportTokensWithMetadata(costReportTokens)
	}

	return nil
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
				LabelFilter:     tfLabelFilter,
			}
			costReportTokens = append(costReportTokens, costReportToken)
		}
		model.CostReportTokensWithMetadata = costReportTokens

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
				LabelFilter:     tfLabelFilter,
			}
			costReportTokens = append(costReportTokens, costReportToken)
		}
		model.CostReportTokensWithMetadata = costReportTokens
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

func (m *businessMetricResourceModel) costReportTokensFromTf(ctx context.Context, diags *diag.Diagnostics) []*businessMetricResourceModelCostReportToken {
	costReportTokens := make([]*businessMetricResourceModelCostReportToken, 0, len(m.CostReportTokensWithMetadata.Elements()))
	if diag := m.CostReportTokensWithMetadata.ElementsAs(ctx, &costReportTokens, false); diag.HasError() {
		diags.Append(diag...)
		return nil
	}
	return costReportTokens
}

func datadogMetricFieldsFromApiModel(ctx context.Context, apiFields *modelsv2.DatadogMetricFields, integrationToken string) (resource_business_metric.DatadogMetricFieldsValue, diag.Diagnostics) {
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
			"integration_token": types.StringValue(integrationToken),
		},
	)
	diags.Append(d...)

	return tfValue, diags
}

func cloudwatchFieldsFromApiModel(ctx context.Context, apiFields *modelsv2.CloudwatchFields, integrationToken string) (resource_business_metric.CloudwatchFieldsValue, diag.Diagnostics) {
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
			"label_dimension":   types.StringValue(apiFields.LabelDimension),
			"dimensions":        dimensionsListValue,
			"integration_token": types.StringValue(integrationToken),
		},
	)
	diags.Append(d...)

	if diags.HasError() {
		return resource_business_metric.NewCloudwatchFieldsValueUnknown(), diags
	}

	return tfValue, diags
}
