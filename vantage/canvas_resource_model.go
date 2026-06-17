package vantage

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

var canvasDataAttrTypes = map[string]attr.Type{
	"error": types.StringType,
	"table": types.StringType,
}

type canvasModel struct {
	CreatedAt      types.String `tfsdk:"created_at"`
	Data           types.Object `tfsdk:"data"`
	Id             types.String `tfsdk:"id"`
	Prompt         types.String `tfsdk:"prompt"`
	Title          types.String `tfsdk:"title"`
	Token          types.String `tfsdk:"token"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	WorkspaceToken types.String `tfsdk:"workspace_token"`
}

func (m *canvasModel) applyPayload(ctx context.Context, payload *modelsv2.Canvas) diag.Diagnostics {
	m.CreatedAt = types.StringValue(payload.CreatedAt)
	m.Prompt = types.StringValue(payload.Prompt)
	m.Title = types.StringValue(payload.Title)
	m.Token = types.StringValue(payload.Token)
	m.Id = types.StringValue(payload.Token)
	m.UpdatedAt = types.StringValue(payload.UpdatedAt)
	m.WorkspaceToken = types.StringValue(payload.WorkspaceToken)

	data, d := canvasDataObject(ctx, payload.Data)
	if d.HasError() {
		return d
	}
	m.Data = data

	return nil
}

func canvasDataObject(ctx context.Context, data *modelsv2.CanvasData) (types.Object, diag.Diagnostics) {
	if data == nil {
		return types.ObjectNull(canvasDataAttrTypes), nil
	}

	tableValue, d := canvasTableJSON(data.Table)
	if d.HasError() {
		return types.ObjectNull(canvasDataAttrTypes), d
	}

	obj, d := types.ObjectValue(canvasDataAttrTypes, map[string]attr.Value{
		"error": types.StringPointerValue(data.Error),
		"table": tableValue,
	})
	if d.HasError() {
		return types.ObjectNull(canvasDataAttrTypes), d
	}

	return obj, nil
}

func canvasTableJSON(table modelsv2.CanvasTable) (types.String, diag.Diagnostics) {
	if table == nil {
		return types.StringNull(), nil
	}

	encoded, err := json.Marshal(table)
	if err != nil {
		var d diag.Diagnostics
		d.AddError("Unable to encode canvas table data", err.Error())
		return types.StringNull(), d
	}

	return types.StringValue(string(encoded)), nil
}

func (m *canvasModel) toCreate() *modelsv2.CreateCanvas {
	body := &modelsv2.CreateCanvas{
		Title:  m.Title.ValueStringPointer(),
		Prompt: m.Prompt.ValueStringPointer(),
	}

	if !m.WorkspaceToken.IsNull() && !m.WorkspaceToken.IsUnknown() {
		body.WorkspaceToken = m.WorkspaceToken.ValueString()
	}

	return body
}

func (m *canvasModel) toUpdate() *modelsv2.UpdateCanvas {
	return &modelsv2.UpdateCanvas{
		Title:  m.Title.ValueString(),
		Prompt: m.Prompt.ValueString(),
	}
}
