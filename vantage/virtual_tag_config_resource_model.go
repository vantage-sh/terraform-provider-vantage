package vantage

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type VirtualTagConfigResourceModel struct {
	Token          types.String `tfsdk:"token"`
	Key            types.String `tfsdk:"key"`
	Overridable    types.Bool   `tfsdk:"overridable"`
	CreatedByToken types.String `tfsdk:"created_by_token"`
	BackfillUntil  types.String `tfsdk:"backfill_until"`
	Values         types.List   `tfsdk:"values"`
}

type virtualTagConfigResourceModelValue struct {
	Name   types.String `tfsdk:"name"`
	Filter types.String `tfsdk:"filter"`
}

func (m *VirtualTagConfigResourceModel) applyPayload(ctx context.Context, payload *modelsv2.VirtualTagConfig) diag.Diagnostics {
	m.Token = types.StringValue(payload.Token)
	m.Key = types.StringValue(payload.Key)
	m.Overridable = types.BoolValue(payload.Overridable)
	m.BackfillUntil = types.StringValue(payload.BackfillUntil)
	m.CreatedByToken = types.StringValue(payload.CreatedByToken)

	if payload.Values != nil {
		tfValues := make([]virtualTagConfigResourceModelValue, 0, len(m.Values.Elements()))
		for _, v := range payload.Values {
			tfValues = append(tfValues, virtualTagConfigResourceModelValue{
				Name:   types.StringValue(v.Name),
				Filter: types.StringValue(v.Filter),
			})
		}

		values, diag := types.ListValueFrom(
			ctx,
			types.ObjectType{AttrTypes: map[string]attr.Type{
				"name":   types.StringType,
				"filter": types.StringType,
			}},
			tfValues,
		)
		if diag.HasError() {
			return diag
		}
		m.Values = values
	}

	return nil
}

func (m *VirtualTagConfigResourceModel) toCreate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.CreateVirtualTagConfig {
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
				Name:   v.Name.ValueStringPointer(),
				Filter: v.Filter.ValueString(),
			}
			values = append(values, value)
		}
		model.Values = values
	}

	return model
}

func (m *VirtualTagConfigResourceModel) toUpdate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.UpdateVirtualTagConfig {
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
				Name:   v.Name.ValueStringPointer(),
				Filter: v.Filter.ValueString(),
			}
			values = append(values, value)
		}
		model.Values = values
	}

	return model
}

func (m *VirtualTagConfigResourceModel) backfillUntilFromTf(diags *diag.Diagnostics) *strfmt.Date {
	date := strfmt.Date{}
	if err := date.UnmarshalText([]byte(m.BackfillUntil.ValueString())); err != nil {
		diags.AddError("Unable to parse backfill_until", err.Error())
	}
	return &date
}

func (m *VirtualTagConfigResourceModel) valuesFromTf(ctx context.Context, diags *diag.Diagnostics) []*virtualTagConfigResourceModelValue {
	values := make([]*virtualTagConfigResourceModelValue, 0, len(m.Values.Elements()))
	if diag := m.Values.ElementsAs(ctx, &values, false); diag.HasError() {
		diags.Append(diag...)
		return nil
	}
	return values
}
