package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_virtual_tag_config"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func (m *virtualTagConfigValueResourceModel) applyValuePayload(ctx context.Context, payload *modelsv2.VirtualTagConfigValue) diag.Diagnostics {
	var diags diag.Diagnostics
	var priorFilters map[string][]string
	if !m.LabelFilters.IsNull() && !m.LabelFilters.IsUnknown() {
		diags.Append(m.LabelFilters.ElementsAs(ctx, &priorFilters, false)...)
		if diags.HasError() {
			return diags
		}
	}
	value, conversionDiags := buildValueFromPayload(ctx, payload, priorFilters)
	diags.Append(conversionDiags...)
	if diags.HasError() {
		return diags
	}
	attributes := value.Attributes()
	m.BusinessMetricToken = attributes["business_metric_token"].(types.String)
	m.DateRanges = attributes["date_ranges"].(types.List)
	m.DisplayName = attributes["display_name"].(types.String)
	m.Filter = attributes["filter"].(types.String)
	m.LabelFilters = attributes["label_filters"].(types.Map)
	m.LabelKey = attributes["label_key"].(types.String)
	m.LabelTransforms = attributes["label_transforms"].(types.List)
	m.LabelValues = attributes["label_values"].(types.List)
	m.Name = attributes["name"].(types.String)
	m.Percentages = attributes["percentages"].(types.List)
	m.Token = attributes["token"].(types.String)
	metric := attributes["cost_metric"].(types.Object)
	m.CostMetric = resource_virtual_tag_config.NewCostMetricValueNull()
	if !metric.IsNull() {
		converted, conversionDiags := resource_virtual_tag_config.NewCostMetricValue(metric.AttributeTypes(ctx), metric.Attributes())
		diags.Append(conversionDiags...)
		m.CostMetric = converted
	}
	m.Id = types.StringValue(payload.Token)
	return diags
}
