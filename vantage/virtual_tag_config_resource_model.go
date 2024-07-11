package vantage

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_virtual_tag_config"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type virtualTagConfigModel resource_virtual_tag_config.VirtualTagConfigModel

type virtualTagConfigResourceModelValue struct {
	BusinessMetricToken types.String                                `tfsdk:"business_metric_token"`
	CostMetric          resource_virtual_tag_config.CostMetricValue `tfsdk:"cost_metric"`
	Filter              types.String                                `tfsdk:"filter"`
	Name                types.String                                `tfsdk:"name"`
}

func (m *virtualTagConfigModel) applyPayload(ctx context.Context, payload *modelsv2.VirtualTagConfig) diag.Diagnostics {
	m.Token = types.StringValue(payload.Token)
	m.Key = types.StringValue(payload.Key)
	m.Overridable = types.BoolValue(payload.Overridable)
	m.BackfillUntil = types.StringValue(payload.BackfillUntil)
	m.CreatedByToken = types.StringValue(payload.CreatedByToken)

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

	if !m.Values.IsNull() && !m.Values.IsUnknown() {
		tfValues := m.valuesFromTf(ctx, diags)
		if diags.HasError() {
			return nil
		}
		values := make([]*modelsv2.CreateVirtualTagConfigValuesItems0, 0, len(tfValues))
		for _, v := range tfValues {
			value := &modelsv2.CreateVirtualTagConfigValuesItems0{
				Name:                v.Name.ValueString(),
				Filter:              v.Filter.ValueStringPointer(),
				BusinessMetricToken: v.BusinessMetricToken.ValueString(),
			}
			if !v.CostMetric.IsNull() && !v.CostMetric.IsUnknown() {
				value.CostMetric = &modelsv2.CreateVirtualTagConfigValuesItems0CostMetric{
					Filter: v.CostMetric.Filter.ValueStringPointer(),
				}

				if !v.CostMetric.Aggregation.IsNull() && !v.CostMetric.Aggregation.IsUnknown() {
					aggregation, diag := resource_virtual_tag_config.NewAggregationValue(v.CostMetric.Aggregation.AttributeTypes(ctx), v.CostMetric.Aggregation.Attributes())
					if diag.HasError() {
						diags.Append(diag...)
						return nil
					}
					value.CostMetric.Aggregation = &modelsv2.CreateVirtualTagConfigValuesItems0CostMetricAggregation{
						Tag: aggregation.Tag.ValueStringPointer(),
					}
				}
			}
			values = append(values, value)
		}
		model.Values = values
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

	if !m.Values.IsNull() && !m.Values.IsUnknown() {
		tfValues := m.valuesFromTf(ctx, diags)
		if diags.HasError() {
			return nil
		}

		values := make([]*modelsv2.UpdateVirtualTagConfigValuesItems0, 0, len(tfValues))
		for _, v := range tfValues {
			value := &modelsv2.UpdateVirtualTagConfigValuesItems0{
				Name:                v.Name.ValueString(),
				Filter:              v.Filter.ValueStringPointer(),
				BusinessMetricToken: v.BusinessMetricToken.ValueString(),
			}

			if !v.CostMetric.IsNull() && !v.CostMetric.IsUnknown() {
				value.CostMetric = &modelsv2.UpdateVirtualTagConfigValuesItems0CostMetric{
					Filter: v.CostMetric.Filter.ValueStringPointer(),
				}

				if !v.CostMetric.Aggregation.IsNull() && !v.CostMetric.Aggregation.IsUnknown() {
					aggregation, diag := resource_virtual_tag_config.NewAggregationValue(v.CostMetric.Aggregation.AttributeTypes(ctx), v.CostMetric.Aggregation.Attributes())
					if diag.HasError() {
						diags.Append(diag...)
						return nil
					}
					value.CostMetric.Aggregation = &modelsv2.UpdateVirtualTagConfigValuesItems0CostMetricAggregation{
						Tag: aggregation.Tag.ValueStringPointer(),
					}
				}
			}
			values = append(values, value)
		}
		model.Values = values
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

func (m *virtualTagConfigModel) valuesFromTf(ctx context.Context, diags *diag.Diagnostics) []*virtualTagConfigResourceModelValue {
	values := make([]*virtualTagConfigResourceModelValue, 0, len(m.Values.Elements()))
	if diag := m.Values.ElementsAs(ctx, &values, false); diag.HasError() {
		diags.Append(diag...)
		return nil
	}
	return values
}
