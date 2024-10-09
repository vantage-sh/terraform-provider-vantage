package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_resource_report"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type resourceReportModel resource_resource_report.ResourceReportModel

func (r *resourceReportModel) applyPayload(_ context.Context, payload *modelsv2.ResourceReport, isDataSource bool) diag.Diagnostics {
	r.CreatedAt = types.StringValue(payload.CreatedAt)
	r.CreatedByToken = types.StringValue(payload.CreatedByToken)
	r.Filter = types.StringValue(payload.Filter)
	r.Title = types.StringValue(payload.Title)
	r.Token = types.StringValue(payload.Token)
	r.UserToken = types.StringValue(payload.UserToken)
	r.WorkspaceToken = types.StringValue(payload.WorkspaceToken)

	return nil
}

func (r *resourceReportModel) toCreateModel() *modelsv2.CreateResourceReport {
	dst := &modelsv2.CreateResourceReport{
		Title:          r.Title.ValueString(),
		Filter:         r.Filter.ValueString(),
		WorkspaceToken: r.WorkspaceToken.ValueStringPointer(),
	}

	return dst
}

func (r *resourceReportModel) toUpdateModel() *modelsv2.UpdateResourceReport {
	dst := &modelsv2.UpdateResourceReport{
		Title:  r.Title.ValueString(),
		Filter: r.Filter.ValueString(),
	}

	return dst
}
