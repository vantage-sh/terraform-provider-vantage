package vantage

import (
	"context"
	"strings"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_kubernetes_efficiency_report"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type kubernetesEfficiencyReportModel resource_kubernetes_efficiency_report.KubernetesEfficiencyReportModel

func (r *kubernetesEfficiencyReportModel) applyPayload(ctx context.Context, payload *modelsv2.KubernetesEfficiencyReport, isDataSource bool) diag.Diagnostics {
	r.CreatedAt = types.StringValue(payload.CreatedAt)
	r.Filter = types.StringPointerValue(payload.Filter)
	r.Title = types.StringValue(payload.Title)
	r.Token = types.StringValue(payload.Token)
	r.Id = types.StringValue(payload.Token)
	r.UserToken = types.StringValue(payload.UserToken)
	r.WorkspaceToken = types.StringValue(payload.WorkspaceToken)
	r.AggregatedBy = types.StringValue(payload.AggregatedBy)
	r.DateBucket = types.StringValue(payload.DateBucket)
	r.DateInterval = types.StringPointerValue(payload.DateInterval)
	r.Default = types.BoolValue(payload.Default)
	r.StartDate = types.StringPointerValue(payload.StartDate)
	r.EndDate = types.StringPointerValue(payload.EndDate)

	// Handle groupings - strings.Split on empty string returns [""], not []
	// so we need to handle that case explicitly
	var groupings []string
	if payload.Groupings != nil && *payload.Groupings != "" {
		groupings = strings.Split(*payload.Groupings, ",")
	} else {
		groupings = []string{}
	}

	var d diag.Diagnostics
	r.Groupings, d = types.ListValueFrom(ctx, types.StringType, groupings)

	if d.HasError() {
		return d
	}

	return nil
}

func (r *kubernetesEfficiencyReportModel) toCreateModel(ctx context.Context) *modelsv2.CreateKubernetesEfficiencyReport {

	var groupings []string

	if !r.Groupings.IsNull() && !r.Groupings.IsUnknown() {
		r.Groupings.ElementsAs(ctx, &groupings, false)
	}

	r.Groupings.Elements()
	dst := &modelsv2.CreateKubernetesEfficiencyReport{
		Title:          r.Title.ValueStringPointer(),
		Filter:         r.Filter.ValueString(),
		WorkspaceToken: r.WorkspaceToken.ValueStringPointer(),
		AggregatedBy:   r.AggregatedBy.ValueString(),
		DateBucket:     r.DateBucket.ValueString(),
		DateInterval:   r.DateInterval.ValueString(),
		Groupings:      groupings,
	}

	if r.DateInterval.ValueString() == "custom" {
		endDate := strfmt.Date{}
		if r.EndDate.ValueString() != "" {
			endDate.UnmarshalText([]byte(r.EndDate.ValueString()))
		}
		startDate := strfmt.Date{}
		if r.StartDate.ValueString() != "" {
			startDate.UnmarshalText([]byte(r.StartDate.ValueString()))
		}
		dst.EndDate = endDate
		dst.StartDate = startDate
	}
	return dst
}

func (r *kubernetesEfficiencyReportModel) toUpdateModel(ctx context.Context) *modelsv2.UpdateKubernetesEfficiencyReport {
	var groupings []string

	if !r.Groupings.IsNull() && !r.Groupings.IsUnknown() {
		r.Groupings.ElementsAs(ctx, &groupings, false)
	}

	dst := &modelsv2.UpdateKubernetesEfficiencyReport{
		Title:        r.Title.ValueString(),
		Filter:       r.Filter.ValueString(),
		AggregatedBy: r.AggregatedBy.ValueString(),
		DateBucket:   r.DateBucket.ValueString(),
		DateInterval: r.DateInterval.ValueString(),
		Groupings:    groupings,
	}

	if r.DateInterval.ValueString() == "custom" {
		endDate := strfmt.Date{}
		if r.EndDate.ValueString() != "" {
			endDate.UnmarshalText([]byte(r.EndDate.ValueString()))
		}
		startDate := strfmt.Date{}
		if r.StartDate.ValueString() != "" {
			startDate.UnmarshalText([]byte(r.StartDate.ValueString()))
		}
		dst.EndDate = endDate
		dst.StartDate = startDate
	}
	return dst
}
