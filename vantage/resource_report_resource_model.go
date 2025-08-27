package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_resource_report"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type resourceReportModel resource_resource_report.ResourceReportModel

func (r *resourceReportModel) applyPayload(ctx context.Context, payload *modelsv2.ResourceReport, isDataSource bool) diag.Diagnostics {
	r.CreatedAt = types.StringValue(payload.CreatedAt)
	r.CreatedByToken = types.StringValue(payload.CreatedByToken)
	r.Filter = types.StringValue(payload.Filter)
	r.Title = types.StringValue(payload.Title)
	r.Token = types.StringValue(payload.Token)
	r.Id = types.StringValue(payload.Token)
	r.UserToken = types.StringValue(payload.UserToken)
	r.WorkspaceToken = types.StringValue(payload.WorkspaceToken)

	if r.Columns.IsNull() || r.Columns.IsUnknown() {
		r.Columns = types.ListNull(types.StringType)
	}

	return nil
}

func (r *resourceReportModel) toCreateModel() *modelsv2.CreateResourceReport {
	dst := &modelsv2.CreateResourceReport{
		Title:          r.Title.ValueString(),
		Filter:         r.Filter.ValueString(),
		WorkspaceToken: r.WorkspaceToken.ValueStringPointer(),
	}

	if !r.Columns.IsNull() && !r.Columns.IsUnknown() {
		columns := make([]string, 0, len(r.Columns.Elements()))
		for _, element := range r.Columns.Elements() {
			if str, ok := element.(types.String); ok {
				columns = append(columns, str.ValueString())
			}
		}
		dst.Columns = columns
	}

	return dst
}

func (r *resourceReportModel) toUpdateModel() *modelsv2.UpdateResourceReport {
	dst := &modelsv2.UpdateResourceReport{
		Title:  r.Title.ValueString(),
		Filter: r.Filter.ValueString(),
	}

	if !r.Columns.IsNull() && !r.Columns.IsUnknown() {
		columns := make([]string, 0, len(r.Columns.Elements()))
		for _, element := range r.Columns.Elements() {
			if str, ok := element.(types.String); ok {
				columns = append(columns, str.ValueString())
			}
		}
		dst.Columns = columns
	}

	return dst
}
