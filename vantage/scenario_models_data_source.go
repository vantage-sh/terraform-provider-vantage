package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_scenario_models"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	scenariomodelsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/scenario_models"
)

var (
	_ datasource.DataSource              = (*scenarioModelsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*scenarioModelsDataSource)(nil)
)

func NewScenarioModelsDataSource() datasource.DataSource {
	return &scenarioModelsDataSource{}
}

type scenarioModelsDataSource struct {
	client *Client
}

type scenarioModelsDataSourceModel struct {
	ScenarioModels []scenarioModelDataSourceModel `tfsdk:"scenario_models"`
}

type scenarioModelDataSourceModel struct {
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

type scenarioModelDataSourcePeriodModel struct {
	Amount     types.String `tfsdk:"amount"`
	AmountType types.String `tfsdk:"amount_type"`
	EndAt      types.String `tfsdk:"end_at"`
	StartAt    types.String `tfsdk:"start_at"`
}

func (d *scenarioModelsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

func (d *scenarioModelsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scenario_models"
}

func (d *scenarioModelsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	s := datasource_scenario_models.ScenarioModelsDataSourceSchema(ctx)

	listAttr, ok := s.Attributes["scenario_models"].(schema.ListNestedAttribute)
	if ok {
		// API field is "provider"; Terraform reserves that root attribute name.
		listAttr.NestedObject.Attributes["cloud_provider"] = schema.StringAttribute{
			Computed:            true,
			Description:         "The cloud provider filter for the ScenarioModel.",
			MarkdownDescription: "The cloud provider filter for the ScenarioModel.",
		}
		// Drop the generated CustomType so the extra attribute is accepted.
		listAttr.NestedObject.CustomType = nil
		if periodsAttr, ok := listAttr.NestedObject.Attributes["periods"].(schema.ListNestedAttribute); ok {
			periodsAttr.NestedObject.CustomType = nil
			listAttr.NestedObject.Attributes["periods"] = periodsAttr
		}
		s.Attributes["scenario_models"] = listAttr
	}

	resp.Schema = s
}

func (d *scenarioModelsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data scenarioModelsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := scenariomodelsv2.NewGetScenarioModelsParams()
	out, err := d.client.V2.ScenarioModels.GetScenarioModels(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Scenario Models",
			err.Error(),
		)
		return
	}

	models := out.Payload.ScenarioModels
	if models == nil {
		models = []*modelsv2.ScenarioModel{}
	}

	values := make([]scenarioModelDataSourceModel, 0, len(models))
	for _, model := range models {
		value, diags := scenarioModelDataSourceModelFromAPI(ctx, model)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		values = append(values, value)
	}

	data.ScenarioModels = values
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func scenarioModelDataSourceModelFromAPI(ctx context.Context, src *modelsv2.ScenarioModel) (scenarioModelDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if src == nil {
		return scenarioModelDataSourceModel{}, diags
	}

	periods, d := scenarioModelDataSourcePeriodsFromAPI(ctx, src.Periods)
	diags.Append(d...)
	if diags.HasError() {
		return scenarioModelDataSourceModel{}, diags
	}

	priority := types.Int64Null()
	if src.Priority != nil {
		priority = types.Int64Value(int64(*src.Priority))
	}

	return scenarioModelDataSourceModel{
		CreatedAt:      types.StringValue(src.CreatedAt),
		CreatedByToken: types.StringPointerValue(src.CreatedByToken),
		Id:             types.StringValue(src.Token),
		Periods:        periods,
		Priority:       priority,
		CloudProvider:  types.StringPointerValue(src.Provider),
		Service:        types.StringPointerValue(src.Service),
		Title:          types.StringValue(src.Title),
		Token:          types.StringValue(src.Token),
		UpdatedAt:      types.StringValue(src.UpdatedAt),
		WorkspaceToken: types.StringPointerValue(src.WorkspaceToken),
	}, diags
}

func scenarioModelDataSourcePeriodsFromAPI(ctx context.Context, src []*modelsv2.ScenarioModelPeriod) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrTypes := map[string]attr.Type{
		"amount":      types.StringType,
		"amount_type": types.StringType,
		"end_at":      types.StringType,
		"start_at":    types.StringType,
	}
	objectType := types.ObjectType{AttrTypes: attrTypes}

	if src == nil {
		return types.ListNull(objectType), diags
	}

	periods := make([]scenarioModelDataSourcePeriodModel, 0, len(src))
	for _, p := range src {
		if p == nil {
			continue
		}
		endAt := types.StringNull()
		if p.EndAt != nil {
			endAt = types.StringValue(*p.EndAt)
		}
		periods = append(periods, scenarioModelDataSourcePeriodModel{
			Amount:     types.StringValue(p.Amount),
			AmountType: types.StringValue(p.AmountType),
			EndAt:      endAt,
			StartAt:    types.StringValue(p.StartAt),
		})
	}

	list, d := types.ListValueFrom(ctx, objectType, periods)
	diags.Append(d...)
	return list, diags
}
