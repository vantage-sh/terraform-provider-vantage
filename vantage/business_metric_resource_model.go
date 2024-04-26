package vantage

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type businessMetricResourceModel struct {
	CostReportTokensWithMetadata types.List   `tfsdk:"cost_report_tokens_with_metadata"`
	CreatedByToken               types.String `tfsdk:"created_by_token"`
	Title                        types.String `tfsdk:"title"`
	Token                        types.String `tfsdk:"token"`
	Values                       types.List   `tfsdk:"values"`
}

type businessMetricResourceModelValue struct {
	Amount types.Float64 `tfsdk:"amount"`
	Date   types.String  `tfsdk:"date"`
	Label  types.String  `tfsdk:"label"`
}

type businessMetricDataSourceModelValue struct {
	Amount types.String `tfsdk:"amount"`
	Date   types.String `tfsdk:"date"`
	Label  types.String `tfsdk:"label"`
}
type businessMetricResourceModelCostReportToken struct {
	CostReportToken types.String `tfsdk:"cost_report_token"`
	UnitScale       types.String `tfsdk:"unit_scale"`
	LabelFilter     types.List   `tfsdk:"label_filter"`
}

func (m *businessMetricResourceModel) applyPayload(ctx context.Context, payload *modelsv2.BusinessMetric, isDataSource bool) diag.Diagnostics {

	m.Title = types.StringValue(payload.Title)
	m.Token = types.StringValue(payload.Token)
	m.CreatedByToken = types.StringValue(payload.CreatedByToken)
	if payload.Values != nil {
		diag := m.parseValues(ctx, isDataSource, payload.Values)
		if diag.HasError() {
			return diag
		}
	}

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

		m.CostReportTokensWithMetadata = costReportTokens
	}

	return nil
}

func (m *businessMetricResourceModel) toCreate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.CreateBusinessMetric {
	model := &modelsv2.CreateBusinessMetric{
		Title: m.Title.ValueStringPointer(),
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

func (m *businessMetricResourceModel) parseValues(ctx context.Context, isDataSource bool, values []*modelsv2.BusinessMetricValue) diag.Diagnostics {
	if isDataSource {
		tfValues := []businessMetricDataSourceModelValue{}
		for _, value := range values {
			t, _ := time.Parse(time.RFC3339, value.Date)
			date := t.Format("2006-01-02")
			tfValues = append(tfValues, businessMetricDataSourceModelValue{
				Amount: types.StringValue(value.Amount),
				Date:   types.StringValue(date),
				Label:  types.StringValue(value.Label),
			})
		}

		attrTypes := map[string]attr.Type{
			"amount": types.StringType,
			"date":   types.StringType,
			"label":  types.StringType,
		}

		values, diagnostic := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: attrTypes}, tfValues)

		if diagnostic.HasError() {
			return diagnostic
		}
		m.Values = values

	} else {
		tfValues := []businessMetricResourceModelValue{}
		for _, value := range values {
			amt, _ := strconv.ParseFloat(value.Amount, 64)

			t, _ := time.Parse(time.RFC3339, value.Date)
			date := t.Format("2006-01-02")

			tfValues = append(tfValues, businessMetricResourceModelValue{
				Amount: types.Float64Value(amt),
				Date:   types.StringValue(date),
				Label:  types.StringValue(value.Label),
			})
		}
		attrTypes := map[string]attr.Type{
			"amount": types.Float64Type,
			"date":   types.StringType,
			"label":  types.StringType,
		}
		values, diagnostic := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: attrTypes}, tfValues)
		if diagnostic.HasError() {
			return diagnostic
		}
		m.Values = values
	}

	return nil
}
