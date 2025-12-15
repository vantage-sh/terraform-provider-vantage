package vantage

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_virtual_tag_config"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type virtualTagConfigModel resource_virtual_tag_config.VirtualTagConfigModel

type virtualTagConfigValueModel struct {
	BusinessMetricToken types.String                                `tfsdk:"business_metric_token"`
	CostMetric          resource_virtual_tag_config.CostMetricValue `tfsdk:"cost_metric"`
	Filter              types.String                                `tfsdk:"filter"`
	Name                types.String                                `tfsdk:"name"`
	Percentages         types.List                                  `tfsdk:"percentages"`
}

// Intermediate types for shared conversion logic between Create and Update operations.
// The API generates separate types for each operation, but the data extraction from
// Terraform state is identical.

type collapsedTagKeyData struct {
	Key       *string
	Providers []string
}

type percentageData struct {
	Pct   float32
	Value *string
}

type aggregationData struct {
	Tag *string
}

type costMetricData struct {
	Filter      *string
	Aggregation *aggregationData
}

type valueData struct {
	Name                string
	Filter              *string
	BusinessMetricToken string
	CostMetric          *costMetricData
	Percentages         []percentageData
}

func (m *virtualTagConfigModel) applyPayload(ctx context.Context, payload *modelsv2.VirtualTagConfig) diag.Diagnostics {
	m.Token = types.StringValue(payload.Token)
	m.Id = types.StringValue(payload.Token)
	m.Key = types.StringValue(payload.Key)
	m.Overridable = types.BoolValue(payload.Overridable)
	m.BackfillUntil = types.StringValue(payload.BackfillUntil)
	m.CreatedByToken = types.StringValue(payload.CreatedByToken)

	tfCollapsedTagKeys := make([]resource_virtual_tag_config.CollapsedTagKeysValue, 0, len(payload.CollapsedTagKeys))
	for _, c := range payload.CollapsedTagKeys {
		tfProviders, diag := types.ListValueFrom(ctx, types.StringType, c.Providers)
		if diag.HasError() {
			return diag
		}
		collapsedTagKey, diag := resource_virtual_tag_config.NewCollapsedTagKeysValue(resource_virtual_tag_config.CollapsedTagKeysValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"key":       types.StringValue(c.Key),
			"providers": tfProviders,
		})
		if diag.HasError() {
			return diag
		}

		tfCollapsedTagKeys = append(tfCollapsedTagKeys, collapsedTagKey)
	}

	tfCollapsedTagKeysValue, diag := types.ListValueFrom(ctx, resource_virtual_tag_config.CollapsedTagKeysValue{}.Type(ctx), tfCollapsedTagKeys)
	if diag.HasError() {
		return diag
	}
	m.CollapsedTagKeys = tfCollapsedTagKeysValue

	if payload.Values != nil {
		tfValues := make([]basetypes.ObjectValue, 0, len(m.Values.Elements()))
		for _, v := range payload.Values {
			value := resource_virtual_tag_config.ValuesValue{
				Name:                types.StringValue(v.Name),
				Filter:              types.StringValue(v.Filter),
				BusinessMetricToken: types.StringValue(v.BusinessMetricToken),
			}

			if v.CostMetric != nil {
				costMetric := resource_virtual_tag_config.CostMetricValue{
					Filter: types.StringValue(v.CostMetric.Filter),
				}

				if v.CostMetric.Aggregation != nil {
					aggregation := resource_virtual_tag_config.AggregationValue{
						Tag: types.StringValue(v.CostMetric.Aggregation.Tag),
					}

					tfAggregation, diag := aggregation.ToObjectValue(ctx)
					if diag.HasError() {
						return diag
					}

					costMetric.Aggregation = tfAggregation
				}

				tfCostMetric, diag := costMetric.ToObjectValue(ctx)
				if diag.HasError() {
					return diag
				}

				value.CostMetric = tfCostMetric
			}

			if v.Percentages != nil {
				tfPercentages := make([]resource_virtual_tag_config.PercentagesValue, 0, len(v.Percentages))
				for _, p := range v.Percentages {
					pv, diag := resource_virtual_tag_config.NewPercentagesValue(
						resource_virtual_tag_config.PercentagesValue{}.AttributeTypes(ctx),
						map[string]attr.Value{
							"pct":   types.Float64Value(*p.Pct),
							"value": types.StringValue(*p.Value),
						},
					)
					if diag.HasError() {
						return diag
					}
					tfPercentages = append(tfPercentages, pv)
				}
				tfPercentagesValue, diag := types.ListValueFrom(
					ctx,
					resource_virtual_tag_config.PercentagesValue{}.Type(ctx),
					tfPercentages,
				)
				if diag.HasError() {
					return diag
				}
				value.Percentages = tfPercentagesValue
			}

			tfValue, diag := value.ToObjectValue(ctx)
			if diag.HasError() {
				return diag
			}
			tfValues = append(tfValues, tfValue)
		}

		values, diag := types.ListValueFrom(
			ctx,
			types.ObjectType{AttrTypes: resource_virtual_tag_config.ValuesValue{}.AttributeTypes(ctx)},
			tfValues,
		)
		if diag.HasError() {
			return diag
		}
		m.Values = values
	}

	return nil
}

func (m *virtualTagConfigModel) toCreate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.CreateVirtualTagConfig {
	model := &modelsv2.CreateVirtualTagConfig{
		Key:         m.Key.ValueStringPointer(),
		Overridable: m.Overridable.ValueBoolPointer(),
	}

	backfillUntil := m.backfillUntilFromTf(diags)
	if diags.HasError() {
		return nil
	}
	model.BackfillUntil = *backfillUntil

	if collapsedTagKeys := m.collapsedTagKeysFromTf(ctx, diags); collapsedTagKeys != nil {
		if diags.HasError() {
			return nil
		}
		model.CollapsedTagKeys = make([]*modelsv2.CreateVirtualTagConfigCollapsedTagKeysItems0, 0, len(collapsedTagKeys))
		for _, c := range collapsedTagKeys {
			model.CollapsedTagKeys = append(model.CollapsedTagKeys, &modelsv2.CreateVirtualTagConfigCollapsedTagKeysItems0{
				Key:       c.Key,
				Providers: c.Providers,
			})
		}
	}

	if !m.Values.IsNull() && !m.Values.IsUnknown() {
		tfValues := m.valuesFromTf(ctx, diags)
		if diags.HasError() {
			return nil
		}
		model.Values = make([]*modelsv2.CreateVirtualTagConfigValuesItems0, 0, len(tfValues))
		for _, v := range tfValues {
			data := v.toValueData(ctx, diags)
			if diags.HasError() {
				return nil
			}

			value := &modelsv2.CreateVirtualTagConfigValuesItems0{
				Name:                data.Name,
				Filter:              data.Filter,
				BusinessMetricToken: data.BusinessMetricToken,
			}

			if data.CostMetric != nil {
				value.CostMetric = &modelsv2.CreateVirtualTagConfigValuesItems0CostMetric{
					Filter: data.CostMetric.Filter,
				}
				if data.CostMetric.Aggregation != nil {
					value.CostMetric.Aggregation = &modelsv2.CreateVirtualTagConfigValuesItems0CostMetricAggregation{
						Tag: data.CostMetric.Aggregation.Tag,
					}
				}
			}

			if len(data.Percentages) > 0 {
				value.Percentages = make([]*modelsv2.CreateVirtualTagConfigValuesItems0PercentagesItems0, 0, len(data.Percentages))
				for _, p := range data.Percentages {
					pct := p.Pct
					value.Percentages = append(value.Percentages, &modelsv2.CreateVirtualTagConfigValuesItems0PercentagesItems0{
						Pct:   &pct,
						Value: p.Value,
					})
				}
			}
			model.Values = append(model.Values, value)
		}
	}

	return model
}

func (m *virtualTagConfigModel) toUpdate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.UpdateVirtualTagConfig {
	if m.Token.IsNull() || m.Token.IsUnknown() {
		diags.AddError("virtual_tag_config_token is required", "")
		return nil
	}

	model := &modelsv2.UpdateVirtualTagConfig{}

	if !m.Key.IsNull() {
		model.Key = m.Key.ValueString()
	}

	if !m.Overridable.IsNull() {
		model.Overridable = m.Overridable.ValueBoolPointer()
	}

	if !m.BackfillUntil.IsNull() {
		model.BackfillUntil = m.backfillUntilFromTf(diags)
		if diags.HasError() {
			return nil
		}
	}

	if collapsedTagKeys := m.collapsedTagKeysFromTf(ctx, diags); collapsedTagKeys != nil {
		if diags.HasError() {
			return nil
		}
		model.CollapsedTagKeys = make([]*modelsv2.UpdateVirtualTagConfigCollapsedTagKeysItems0, 0, len(collapsedTagKeys))
		for _, c := range collapsedTagKeys {
			model.CollapsedTagKeys = append(model.CollapsedTagKeys, &modelsv2.UpdateVirtualTagConfigCollapsedTagKeysItems0{
				Key:       c.Key,
				Providers: c.Providers,
			})
		}
	}

	if !m.Values.IsNull() && !m.Values.IsUnknown() {
		tfValues := m.valuesFromTf(ctx, diags)
		if diags.HasError() {
			return nil
		}

		model.Values = make([]*modelsv2.UpdateVirtualTagConfigValuesItems0, 0, len(tfValues))
		for _, v := range tfValues {
			data := v.toValueData(ctx, diags)
			if diags.HasError() {
				return nil
			}

			value := &modelsv2.UpdateVirtualTagConfigValuesItems0{
				Name:                data.Name,
				Filter:              data.Filter,
				BusinessMetricToken: data.BusinessMetricToken,
			}

			if data.CostMetric != nil {
				value.CostMetric = &modelsv2.UpdateVirtualTagConfigValuesItems0CostMetric{
					Filter: data.CostMetric.Filter,
				}
				if data.CostMetric.Aggregation != nil {
					value.CostMetric.Aggregation = &modelsv2.UpdateVirtualTagConfigValuesItems0CostMetricAggregation{
						Tag: data.CostMetric.Aggregation.Tag,
					}
				}
			}

			if len(data.Percentages) > 0 {
				value.Percentages = make([]*modelsv2.UpdateVirtualTagConfigValuesItems0PercentagesItems0, 0, len(data.Percentages))
				for _, p := range data.Percentages {
					pct := p.Pct
					value.Percentages = append(value.Percentages, &modelsv2.UpdateVirtualTagConfigValuesItems0PercentagesItems0{
						Pct:   &pct,
						Value: p.Value,
					})
				}
			}
			model.Values = append(model.Values, value)
		}
	}

	return model
}

func (m *virtualTagConfigModel) backfillUntilFromTf(diags *diag.Diagnostics) *strfmt.Date {
	date := strfmt.Date{}
	if err := date.UnmarshalText([]byte(m.BackfillUntil.ValueString())); err != nil {
		diags.AddError("Unable to parse backfill_until", err.Error())
	}
	return &date
}

func (m *virtualTagConfigModel) valuesFromTf(ctx context.Context, diags *diag.Diagnostics) []*virtualTagConfigValueModel {
	values := make([]*virtualTagConfigValueModel, 0, len(m.Values.Elements()))
	if diag := m.Values.ElementsAs(ctx, &values, false); diag.HasError() {
		diags.Append(diag...)
		return nil
	}
	return values
}

// collapsedTagKeysFromTf extracts collapsed tag keys from Terraform state into an intermediate format.
func (m *virtualTagConfigModel) collapsedTagKeysFromTf(ctx context.Context, diags *diag.Diagnostics) []collapsedTagKeyData {
	if m.CollapsedTagKeys.IsNull() || m.CollapsedTagKeys.IsUnknown() {
		return nil
	}

	tfCollapsedTagKeys := make([]resource_virtual_tag_config.CollapsedTagKeysValue, 0, len(m.CollapsedTagKeys.Elements()))
	if d := m.CollapsedTagKeys.ElementsAs(ctx, &tfCollapsedTagKeys, false); d.HasError() {
		diags.Append(d...)
		return nil
	}

	result := make([]collapsedTagKeyData, 0, len(tfCollapsedTagKeys))
	for _, c := range tfCollapsedTagKeys {
		providers := make([]string, 0, len(c.Providers.Elements()))
		if d := c.Providers.ElementsAs(ctx, &providers, false); d.HasError() {
			diags.Append(d...)
			return nil
		}
		result = append(result, collapsedTagKeyData{
			Key:       c.Key.ValueStringPointer(),
			Providers: providers,
		})
	}
	return result
}

// valueDataFromTf extracts a single value's data from Terraform state into an intermediate format.
func (v *virtualTagConfigValueModel) toValueData(ctx context.Context, diags *diag.Diagnostics) *valueData {
	data := &valueData{
		Name:                v.Name.ValueString(),
		Filter:              v.Filter.ValueStringPointer(),
		BusinessMetricToken: v.BusinessMetricToken.ValueString(),
	}

	if !v.CostMetric.IsNull() && !v.CostMetric.IsUnknown() {
		data.CostMetric = &costMetricData{
			Filter: v.CostMetric.Filter.ValueStringPointer(),
		}

		if !v.CostMetric.Aggregation.IsNull() && !v.CostMetric.Aggregation.IsUnknown() {
			aggregation, d := resource_virtual_tag_config.NewAggregationValue(
				v.CostMetric.Aggregation.AttributeTypes(ctx),
				v.CostMetric.Aggregation.Attributes(),
			)
			if d.HasError() {
				diags.Append(d...)
				return nil
			}
			data.CostMetric.Aggregation = &aggregationData{
				Tag: aggregation.Tag.ValueStringPointer(),
			}
		}
	}

	if !v.Percentages.IsNull() && !v.Percentages.IsUnknown() {
		tfPercentages := make([]resource_virtual_tag_config.PercentagesValue, 0, len(v.Percentages.Elements()))
		if d := v.Percentages.ElementsAs(ctx, &tfPercentages, false); d.HasError() {
			diags.Append(d...)
			return nil
		}
		data.Percentages = make([]percentageData, 0, len(tfPercentages))
		for _, p := range tfPercentages {
			data.Percentages = append(data.Percentages, percentageData{
				Pct:   float32(p.Pct.ValueFloat64()),
				Value: p.Value.ValueStringPointer(),
			})
		}
	}

	return data
}
