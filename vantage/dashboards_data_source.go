package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_dashboards"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	dashboardsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/dashboards"
)

var (
	_ datasource.DataSource              = &dashboardsDataSource{}
	_ datasource.DataSourceWithConfigure = &dashboardsDataSource{}
)

type dashboardsDataSource struct {
	client *Client
}

func NewDashboardsDataSource() datasource.DataSource {
	return &dashboardsDataSource{}
}

func (d *dashboardsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_dashboards.DashboardsDataSourceSchema(ctx)
}

func (d *dashboardsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*Client)
}

func (d *dashboardsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboards"
}

func (d *dashboardsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_dashboards.DashboardsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := dashboardsv2.NewGetDashboardsParams()
	if !data.Q.IsNull() && !data.Q.IsUnknown() {
		params.SetQ(data.Q.ValueStringPointer())
	}
	if !data.WorkspaceToken.IsNull() && !data.WorkspaceToken.IsUnknown() {
		params.SetWorkspaceToken(data.WorkspaceToken.ValueStringPointer())
	}

	out, err := d.client.V2.Dashboards.GetDashboards(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Dashboards",
			err.Error(),
		)
		return
	}

	dashboards := make([]datasource_dashboards.DashboardsValue, 0, len(out.Payload.Dashboards))
	for _, dashboard := range out.Payload.Dashboards {
		value, diags := dashboardDataSourceValueFromPayload(ctx, dashboard)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		dashboards = append(dashboards, value)
	}

	list, diags := types.ListValueFrom(ctx, datasource_dashboards.DashboardsValue{}.Type(ctx), dashboards)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Dashboards = list

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func dashboardDataSourceValueFromPayload(ctx context.Context, payload *modelsv2.Dashboard) (datasource_dashboards.DashboardsValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	savedFilters, d := types.ListValueFrom(ctx, types.StringType, payload.SavedFilterTokens)
	diags.Append(d...)
	if diags.HasError() {
		return datasource_dashboards.NewDashboardsValueNull(), diags
	}

	widgets, d := dashboardDataSourceWidgetsFromPayload(ctx, payload.Widgets)
	diags.Append(d...)
	if diags.HasError() {
		return datasource_dashboards.NewDashboardsValueNull(), diags
	}

	value, d := datasource_dashboards.NewDashboardsValue(
		datasource_dashboards.DashboardsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"created_at":          types.StringValue(payload.CreatedAt),
			"date_bin":            types.StringPointerValue(payload.DateBin),
			"date_interval":       types.StringPointerValue(payload.DateInterval),
			"end_date":            types.StringPointerValue(payload.EndDate),
			"id":                  types.StringValue(payload.Token),
			"saved_filter_tokens": savedFilters,
			"start_date":          types.StringPointerValue(payload.StartDate),
			"title":               types.StringValue(payload.Title),
			"token":               types.StringValue(payload.Token),
			"updated_at":          types.StringValue(payload.UpdatedAt),
			"widgets":             widgets,
			"workspace_token":     types.StringValue(payload.WorkspaceToken),
		},
	)
	diags.Append(d...)
	return value, diags
}

func dashboardDataSourceWidgetsFromPayload(ctx context.Context, payload []*modelsv2.DashboardWidget) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	settingsAttrTypes := datasource_dashboards.SettingsValue{}.AttributeTypes(ctx)
	widgetAttrTypes := datasource_dashboards.WidgetsValue{}.AttributeTypes(ctx)

	tfWidgets := make([]datasource_dashboards.WidgetsValue, 0, len(payload))
	for _, widget := range payload {
		var settingsObj types.Object
		if widget.Settings != nil {
			settingsVal, d := datasource_dashboards.NewSettingsValue(settingsAttrTypes, map[string]attr.Value{
				"display_type":    types.StringValue(widget.Settings.DisplayType),
				"grid":            dashboardDataSourceGridObject(ctx, widget.Settings.Grid),
				"kpi_calculation": types.StringPointerValue(widget.Settings.KpiCalculation),
				"kpi_type":        types.StringPointerValue(widget.Settings.KpiType),
				"kpi_usage_unit":  types.StringPointerValue(widget.Settings.KpiUsageUnit),
			})
			diags.Append(d...)
			if diags.HasError() {
				return types.ListNull(datasource_dashboards.WidgetsValue{}.Type(ctx)), diags
			}
			settingsObj, d = settingsVal.ToObjectValue(ctx)
			diags.Append(d...)
			if diags.HasError() {
				return types.ListNull(datasource_dashboards.WidgetsValue{}.Type(ctx)), diags
			}
		} else {
			settingsObj = types.ObjectNull(settingsAttrTypes)
		}

		widgetVal, d := datasource_dashboards.NewWidgetsValue(widgetAttrTypes, map[string]attr.Value{
			"settings":         settingsObj,
			"title":            types.StringValue(widget.Title),
			"token":            types.StringValue(widget.Token),
			"widgetable_token": stringValueOrNull(widget.WidgetableToken),
		})
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(datasource_dashboards.WidgetsValue{}.Type(ctx)), diags
		}
		tfWidgets = append(tfWidgets, widgetVal)
	}

	list, d := types.ListValueFrom(ctx, datasource_dashboards.WidgetsValue{}.Type(ctx), tfWidgets)
	diags.Append(d...)
	return list, diags
}

func dashboardDataSourceGridObject(ctx context.Context, grid *modelsv2.DashboardWidgetGridLayout) types.Object {
	gridAttrTypes := datasource_dashboards.GridValue{}.AttributeTypes(ctx)
	if grid == nil {
		return types.ObjectNull(gridAttrTypes)
	}

	gridVal, diags := datasource_dashboards.NewGridValue(gridAttrTypes, map[string]attr.Value{
		"x": types.Int64Value(int64(grid.X)),
		"y": types.Int64Value(int64(grid.Y)),
		"w": types.Int64Value(int64(grid.W)),
		"h": types.Int64Value(int64(grid.H)),
	})
	if diags.HasError() {
		return types.ObjectNull(gridAttrTypes)
	}
	obj, diags := gridVal.ToObjectValue(ctx)
	if diags.HasError() {
		return types.ObjectNull(gridAttrTypes)
	}
	return obj
}
