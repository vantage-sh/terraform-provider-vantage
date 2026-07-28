package vantage

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_scenario_model"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

// scenarioModelModel mirrors the generated schema, with API "provider" exposed as
// cloud_provider because "provider" is a reserved root attribute name in Terraform.
type scenarioModelModel struct {
	CreatedAt      types.String `tfsdk:"created_at"`
	CreatedByToken types.String `tfsdk:"created_by_token"`
	Id             types.String `tfsdk:"id"`
	Periods        types.List   `tfsdk:"periods"`
	Priority       types.Int64  `tfsdk:"priority"`
	CloudProvider  types.String `tfsdk:"cloud_provider"`
	Service        types.String `tfsdk:"service"`
	Title          types.String `tfsdk:"title"`
	Token          types.String `tfsdk:"token"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	WorkspaceToken types.String `tfsdk:"workspace_token"`
}

type scenarioModelPeriodModel struct {
	Amount     types.Float64 `tfsdk:"amount"`
	AmountType types.String  `tfsdk:"amount_type"`
	EndAt      types.String  `tfsdk:"end_at"`
	StartAt    types.String  `tfsdk:"start_at"`
}

func (m *scenarioModelModel) applyPayload(ctx context.Context, src *modelsv2.ScenarioModel) diag.Diagnostics {
	m.Token = types.StringValue(src.Token)
	m.Id = types.StringValue(src.Token)
	m.Title = types.StringValue(src.Title)
	m.CreatedAt = types.StringValue(src.CreatedAt)
	m.UpdatedAt = types.StringValue(src.UpdatedAt)
	m.CreatedByToken = types.StringPointerValue(src.CreatedByToken)
	m.WorkspaceToken = types.StringPointerValue(src.WorkspaceToken)
	m.CloudProvider = types.StringPointerValue(src.Provider)
	m.Service = types.StringPointerValue(src.Service)

	if src.Priority != nil {
		m.Priority = types.Int64Value(int64(*src.Priority))
	} else {
		m.Priority = types.Int64Null()
	}

	periods, diags := scenarioModelPeriodsFromAPI(ctx, src.Periods)
	if diags.HasError() {
		return diags
	}
	m.Periods = periods
	return nil
}

func scenarioModelPeriodsFromAPI(ctx context.Context, src []*modelsv2.ScenarioModelPeriod) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	periodType := resource_scenario_model.PeriodsValue{}.Type(ctx)

	if src == nil {
		return types.ListNull(periodType), diags
	}

	elements := make([]attr.Value, 0, len(src))
	attrTypes := resource_scenario_model.PeriodsValue{}.AttributeTypes(ctx)
	for _, p := range src {
		if p == nil {
			continue
		}

		amount, err := strconv.ParseFloat(p.Amount, 64)
		if err != nil {
			diags.AddError("Unable to parse scenario model period amount", err.Error())
			return types.ListNull(periodType), diags
		}

		endAt := types.StringNull()
		if p.EndAt != nil {
			endAt = types.StringValue(*p.EndAt)
		}

		periodValue, d := resource_scenario_model.NewPeriodsValue(
			attrTypes,
			map[string]attr.Value{
				"amount":      types.Float64Value(amount),
				"amount_type": types.StringValue(p.AmountType),
				"end_at":      endAt,
				"start_at":    types.StringValue(p.StartAt),
			},
		)
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(periodType), diags
		}
		elements = append(elements, periodValue)
	}

	list, d := types.ListValue(periodType, elements)
	diags.Append(d...)
	return list, diags
}

func (m *scenarioModelModel) toCreate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.CreateScenarioModel {
	dst := &modelsv2.CreateScenarioModel{
		Title: m.Title.ValueStringPointer(),
	}

	if !m.WorkspaceToken.IsNull() && !m.WorkspaceToken.IsUnknown() && m.WorkspaceToken.ValueString() != "" {
		dst.WorkspaceToken = m.WorkspaceToken.ValueString()
	}
	if !m.Priority.IsNull() && !m.Priority.IsUnknown() {
		priority := int32(m.Priority.ValueInt64())
		dst.Priority = &priority
	}
	if !m.CloudProvider.IsNull() && !m.CloudProvider.IsUnknown() {
		dst.Provider = m.CloudProvider.ValueStringPointer()
	}
	if !m.Service.IsNull() && !m.Service.IsUnknown() {
		dst.Service = m.Service.ValueStringPointer()
	}

	dst.Periods = m.periodsToCreate(ctx, diags)
	if diags.HasError() {
		return nil
	}
	return dst
}

func (m *scenarioModelModel) toUpdate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.UpdateScenarioModel {
	dst := &modelsv2.UpdateScenarioModel{}

	if !m.Title.IsNull() && !m.Title.IsUnknown() {
		dst.Title = m.Title.ValueString()
	}
	if !m.WorkspaceToken.IsNull() && !m.WorkspaceToken.IsUnknown() && m.WorkspaceToken.ValueString() != "" {
		dst.WorkspaceToken = m.WorkspaceToken.ValueString()
	}
	if !m.Priority.IsNull() && !m.Priority.IsUnknown() {
		priority := int32(m.Priority.ValueInt64())
		dst.Priority = &priority
	}
	if !m.CloudProvider.IsNull() && !m.CloudProvider.IsUnknown() {
		dst.Provider = m.CloudProvider.ValueStringPointer()
	}
	if !m.Service.IsNull() && !m.Service.IsUnknown() {
		dst.Service = m.Service.ValueStringPointer()
	}

	if !m.Periods.IsNull() && !m.Periods.IsUnknown() {
		dst.Periods = m.periodsToUpdate(ctx, diags)
		if diags.HasError() {
			return nil
		}
	}
	return dst
}

func (m *scenarioModelModel) periodsFromTf(ctx context.Context, diags *diag.Diagnostics) []*scenarioModelPeriodModel {
	if m.Periods.IsNull() || m.Periods.IsUnknown() {
		return nil
	}
	periods := make([]*scenarioModelPeriodModel, 0, len(m.Periods.Elements()))
	diags.Append(m.Periods.ElementsAs(ctx, &periods, false)...)
	if diags.HasError() {
		return nil
	}
	return periods
}

func (m *scenarioModelModel) periodsToCreate(ctx context.Context, diags *diag.Diagnostics) []*modelsv2.CreateScenarioModelPeriodsItems0 {
	periods := m.periodsFromTf(ctx, diags)
	if diags.HasError() || periods == nil {
		return nil
	}

	dst := make([]*modelsv2.CreateScenarioModelPeriodsItems0, 0, len(periods))
	for _, p := range periods {
		item, ok := periodToCreateItem(p, diags)
		if !ok {
			return nil
		}
		dst = append(dst, item)
	}
	return dst
}

func (m *scenarioModelModel) periodsToUpdate(ctx context.Context, diags *diag.Diagnostics) []*modelsv2.UpdateScenarioModelPeriodsItems0 {
	periods := m.periodsFromTf(ctx, diags)
	if diags.HasError() {
		return nil
	}

	dst := make([]*modelsv2.UpdateScenarioModelPeriodsItems0, 0, len(periods))
	for _, p := range periods {
		item, ok := periodToUpdateItem(p, diags)
		if !ok {
			return nil
		}
		dst = append(dst, item)
	}
	return dst
}

func periodToCreateItem(p *scenarioModelPeriodModel, diags *diag.Diagnostics) (*modelsv2.CreateScenarioModelPeriodsItems0, bool) {
	startAt, ok := parsePeriodDate(p.StartAt, "start_at", diags)
	if !ok {
		return nil, false
	}

	item := &modelsv2.CreateScenarioModelPeriodsItems0{
		Amount:     p.Amount.ValueFloat64Pointer(),
		AmountType: p.AmountType.ValueStringPointer(),
		StartAt:    startAt,
	}

	if !p.EndAt.IsNull() && !p.EndAt.IsUnknown() && p.EndAt.ValueString() != "" {
		endAt, ok := parsePeriodDate(p.EndAt, "end_at", diags)
		if !ok {
			return nil, false
		}
		item.EndAt = endAt
	}
	return item, true
}

func periodToUpdateItem(p *scenarioModelPeriodModel, diags *diag.Diagnostics) (*modelsv2.UpdateScenarioModelPeriodsItems0, bool) {
	startAt, ok := parsePeriodDate(p.StartAt, "start_at", diags)
	if !ok {
		return nil, false
	}

	item := &modelsv2.UpdateScenarioModelPeriodsItems0{
		Amount:     p.Amount.ValueFloat64Pointer(),
		AmountType: p.AmountType.ValueStringPointer(),
		StartAt:    startAt,
	}

	if !p.EndAt.IsNull() && !p.EndAt.IsUnknown() && p.EndAt.ValueString() != "" {
		endAt, ok := parsePeriodDate(p.EndAt, "end_at", diags)
		if !ok {
			return nil, false
		}
		item.EndAt = endAt
	}
	return item, true
}

func parsePeriodDate(value types.String, field string, diags *diag.Diagnostics) (*strfmt.Date, bool) {
	raw := strings.Trim(value.ValueString(), "\"")
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		diags.AddError("parsing error", fmt.Sprintf("failed to parse %s: %s", field, err))
		return nil, false
	}
	date := strfmt.Date(parsed)
	return &date, true
}
