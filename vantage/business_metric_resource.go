package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_business_metric"
	businessmetricsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/business_metrics"
)

var (
	_ resource.Resource                = (*businessMetricResource)(nil)
	_ resource.ResourceWithConfigure   = (*businessMetricResource)(nil)
	_ resource.ResourceWithImportState = (*businessMetricResource)(nil)
)

func NewBusinessMetricResource() resource.Resource {
	return &businessMetricResource{}
}

type businessMetricResource struct {
	client *Client
}

// Configure implements resource.ResourceWithConfigure.
func (r *businessMetricResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)

}

func (r *businessMetricResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_business_metric"
}

func (r *businessMetricResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_business_metric.BusinessMetricResourceSchema(ctx)
	attrs := s.GetAttributes()
	s.Attributes["token"] = schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: attrs["token"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
	applyEmptyLabelDefault(s.Attributes, "values")
	applyEmptyLabelDefault(s.Attributes, "forecasted_values")
	applyCostReportTokenMetadataDefaults(s.Attributes)

	resp.Schema = s
}

func applyEmptyLabelDefault(attrs map[string]schema.Attribute, attrName string) {
	attr, ok := attrs[attrName].(schema.ListNestedAttribute)
	if !ok {
		return
	}

	labelAttr, ok := attr.NestedObject.Attributes["label"].(schema.StringAttribute)
	if !ok {
		return
	}

	labelAttr.Default = stringdefault.StaticString("")
	attr.NestedObject.Attributes["label"] = labelAttr
	attrs[attrName] = attr
}

func applyCostReportTokenMetadataDefaults(attrs map[string]schema.Attribute) {
	attr, ok := attrs["cost_report_tokens_with_metadata"].(schema.ListNestedAttribute)
	if !ok {
		return
	}

	// Mutate in place so generated Validators (and other schema metadata) are preserved.
	if calcAttr, ok := attr.NestedObject.Attributes["calculation_type"].(schema.StringAttribute); ok {
		calcAttr.Default = stringdefault.StaticString("unit_cost")
		attr.NestedObject.Attributes["calculation_type"] = calcAttr
	}

	if labelAttr, ok := attr.NestedObject.Attributes["label"].(schema.StringAttribute); ok {
		labelAttr.PlanModifiers = append(labelAttr.PlanModifiers, stringplanmodifier.UseStateForUnknown())
		attr.NestedObject.Attributes["label"] = labelAttr
	}

	if labelFiltersAttr, ok := attr.NestedObject.Attributes["label_filters"].(schema.MapAttribute); ok {
		labelFiltersAttr.PlanModifiers = append(labelFiltersAttr.PlanModifiers, mapplanmodifier.UseStateForUnknown())
		attr.NestedObject.Attributes["label_filters"] = labelFiltersAttr
	}

	attrs["cost_report_tokens_with_metadata"] = attr
}

// clearOmittedCostReportTokenLabels nulls plan labels that were not set in config so
// UseStateForUnknown does not re-send a stale derived label when calculation_type changes.
func clearOmittedCostReportTokenLabels(ctx context.Context, plan, config *businessMetricResourceModel, diags *diag.Diagnostics) {
	if plan == nil || config == nil {
		return
	}
	if plan.CostReportTokensWithMetadata.IsNull() || plan.CostReportTokensWithMetadata.IsUnknown() {
		return
	}

	planTokens := make([]*businessMetricResourceModelCostReportToken, 0, len(plan.CostReportTokensWithMetadata.Elements()))
	diags.Append(plan.CostReportTokensWithMetadata.ElementsAs(ctx, &planTokens, false)...)
	if diags.HasError() {
		return
	}

	configByToken := map[string]*businessMetricResourceModelCostReportToken{}
	if !config.CostReportTokensWithMetadata.IsNull() && !config.CostReportTokensWithMetadata.IsUnknown() {
		configTokens := make([]*businessMetricResourceModelCostReportToken, 0, len(config.CostReportTokensWithMetadata.Elements()))
		diags.Append(config.CostReportTokensWithMetadata.ElementsAs(ctx, &configTokens, false)...)
		if diags.HasError() {
			return
		}
		for _, token := range configTokens {
			configByToken[token.CostReportToken.ValueString()] = token
		}
	}

	changed := false
	for _, planToken := range planTokens {
		configToken, ok := configByToken[planToken.CostReportToken.ValueString()]
		if ok && !configToken.Label.IsNull() && !configToken.Label.IsUnknown() {
			continue
		}
		if !planToken.Label.IsNull() && !planToken.Label.IsUnknown() {
			planToken.Label = types.StringNull()
			changed = true
		}
	}
	if !changed {
		return
	}

	elements := make([]attr.Value, 0, len(planTokens))
	attrTypes := resourceCostReportTokenAttrTypes(ctx)
	for _, token := range planTokens {
		tokenValue, d := resource_business_metric.NewCostReportTokensWithMetadataValue(
			attrTypes,
			map[string]attr.Value{
				"cost_report_token": token.CostReportToken,
				"unit_scale":        token.UnitScale,
				"calculation_type":  token.CalculationType,
				"label":             token.Label,
				"label_filter":      token.LabelFilter,
				"label_filters":     token.LabelFilters,
			},
		)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
		elements = append(elements, tokenValue)
	}

	newList, d := types.ListValue(resourceCostReportTokenListType(ctx), elements)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	plan.CostReportTokensWithMetadata = newList
}

func (r *businessMetricResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *businessMetricResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	oldValues := data.Values
	oldForecastedValues := data.ForecastedValues
	oldCostReportTokens := data.CostReportTokensWithMetadata
	model := data.toCreate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := businessmetricsv2.NewCreateBusinessMetricParams().WithCreateBusinessMetric(model)
	out, err := r.client.V2.BusinessMetrics.CreateBusinessMetric(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*businessmetricsv2.CreateBusinessMetricBadRequest); ok {
			handleBadRequest("Create Business Metric", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Create Business Metric", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	attrTypes := map[string]attr.Type{
		"amount": types.Float64Type,
		"date":   types.StringType,
		"label":  types.StringType,
	}

	if oldValues.IsUnknown() {
		data.Values = types.ListNull(types.ObjectType{AttrTypes: attrTypes})
	} else {
		assignValues(ctx, data, oldValues, &resp.Diagnostics)
	}

	if oldForecastedValues.IsUnknown() {
		data.ForecastedValues = types.ListNull(types.ObjectType{AttrTypes: attrTypes})
	} else {
		assignForecastedValues(ctx, data, oldForecastedValues, &resp.Diagnostics)
	}

	// Preserve the original order of cost report tokens from the plan
	if !oldCostReportTokens.IsNull() && !oldCostReportTokens.IsUnknown() {
		assignCostReportTokens(ctx, data, oldCostReportTokens, &resp.Diagnostics)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// if labels are unknown in values, sets them to empty string
func assignValues(ctx context.Context, data *businessMetricResourceModel, tfValues types.List, diags *diag.Diagnostics) {
	values := make([]*businessMetricResourceModelValue, 0, len(tfValues.Elements()))
	if diag := tfValues.ElementsAs(ctx, &values, false); diag.HasError() {
		diags.Append(diag...)
		return
	}

	newTfValues := []businessMetricResourceModelValue{}
	for _, value := range values {
		var labelValue types.String
		if value.Label == types.StringUnknown() {
			labelValue = types.StringValue("")
		} else {
			labelValue = value.Label
		}
		newTfValues = append(newTfValues, businessMetricResourceModelValue{
			Amount: value.Amount,
			Date:   value.Date,
			Label:  labelValue,
		})
	}

	attrTypes := map[string]attr.Type{
		"amount": types.Float64Type,
		"date":   types.StringType,
		"label":  types.StringType,
	}

	newList, diag := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: attrTypes}, newTfValues)
	if diag.HasError() {
		diags.Append(diag...)
		return
	}

	data.Values = newList
}

// if labels are unknown in forecasted values, sets them to empty string
func assignForecastedValues(ctx context.Context, data *businessMetricResourceModel, tfValues types.List, diags *diag.Diagnostics) {
	values := make([]*businessMetricResourceModelValue, 0, len(tfValues.Elements()))
	if diag := tfValues.ElementsAs(ctx, &values, false); diag.HasError() {
		diags.Append(diag...)
		return
	}

	newTfValues := []businessMetricResourceModelValue{}
	for _, value := range values {
		var labelValue types.String
		if value.Label == types.StringUnknown() {
			labelValue = types.StringValue("")
		} else {
			labelValue = value.Label
		}
		newTfValues = append(newTfValues, businessMetricResourceModelValue{
			Amount: value.Amount,
			Date:   value.Date,
			Label:  labelValue,
		})
	}

	attrTypes := map[string]attr.Type{
		"amount": types.Float64Type,
		"date":   types.StringType,
		"label":  types.StringType,
	}

	newList, diag := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: attrTypes}, newTfValues)
	if diag.HasError() {
		diags.Append(diag...)
		return
	}

	data.ForecastedValues = newList
}

func (r *businessMetricResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *businessMetricResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save the original order of cost report tokens from state
	oldCostReportTokens := data.CostReportTokensWithMetadata

	params := businessmetricsv2.NewGetBusinessMetricParams().WithBusinessMetricToken(data.Token.ValueString())
	out, err := r.client.V2.BusinessMetrics.GetBusinessMetric(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*businessmetricsv2.GetBusinessMetricNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}
		handleError("Get Business Metric", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// Preserve the original order of cost report tokens from state
	if !oldCostReportTokens.IsNull() && !oldCostReportTokens.IsUnknown() {
		assignCostReportTokens(ctx, data, oldCostReportTokens, &resp.Diagnostics)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *businessMetricResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r *businessMetricResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *businessMetricResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config *businessMetricResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// UseStateForUnknown can keep a previously derived label in plan when config omits
	// it. Clear those so a calculation_type change can re-derive the API default.
	clearOmittedCostReportTokenLabels(ctx, data, config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	oldValues := data.Values
	oldForecastedValues := data.ForecastedValues
	oldCostReportTokens := data.CostReportTokensWithMetadata

	model := data.toUpdate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := businessmetricsv2.NewUpdateBusinessMetricParams().WithBusinessMetricToken(data.Token.ValueString()).WithUpdateBusinessMetric(model)

	out, err := r.client.V2.BusinessMetrics.UpdateBusinessMetric(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*businessmetricsv2.UpdateBusinessMetricBadRequest); ok {
			handleBadRequest("Update Business Metric", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Update Business Metric", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	attrTypes := map[string]attr.Type{
		"amount": types.Float64Type,
		"date":   types.StringType,
		"label":  types.StringType,
	}

	if oldValues.IsUnknown() {
		data.Values = types.ListNull(types.ObjectType{AttrTypes: attrTypes})
	} else {
		assignValues(ctx, data, oldValues, &resp.Diagnostics)
	}

	if oldForecastedValues.IsUnknown() {
		data.ForecastedValues = types.ListNull(types.ObjectType{AttrTypes: attrTypes})
	} else {
		assignForecastedValues(ctx, data, oldForecastedValues, &resp.Diagnostics)
	}

	// Preserve the original order of cost report tokens from the plan
	if !oldCostReportTokens.IsNull() && !oldCostReportTokens.IsUnknown() {
		assignCostReportTokens(ctx, data, oldCostReportTokens, &resp.Diagnostics)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *businessMetricResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *businessMetricResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := businessmetricsv2.NewDeleteBusinessMetricParams()
	params.SetBusinessMetricToken(data.Token.ValueString())

	_, err := r.client.V2.BusinessMetrics.DeleteBusinessMetric(params, r.client.Auth)
	if err != nil {
		handleError("Delete Business Metric", &resp.Diagnostics, err)
		return
	}

}
