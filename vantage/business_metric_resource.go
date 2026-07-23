package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	_ resource.ResourceWithModifyPlan  = (*businessMetricResource)(nil)
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

	// Do not UseStateForUnknown for label / label_filters: omission must clear the
	// prior value (API default / empty filters). Index-based state carryover also
	// mis-attributes values when attachments are reordered.
	attrs["cost_report_tokens_with_metadata"] = attr
}

// ModifyPlan clears omitted attachment label / label_filters that are still known
// in the plan (sticky carryover, reorder-by-index, or calculation_type changes) so
// apply can accept the API result without an inconsistent-result error.
func (r *businessMetricResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var plan, state, config *businessMetricResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || plan == nil || state == nil || config == nil {
		return
	}

	toInvalidate := costReportTokenComputedAttrsToInvalidate(ctx, plan, state, config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, p := range toInvalidate.LabelPaths {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, p, types.StringUnknown())...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	for _, p := range toInvalidate.LabelFiltersPaths {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, p, types.MapUnknown(types.ListType{ElemType: types.StringType}))...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}

type costReportTokenComputedAttrsToInvalidateResult struct {
	LabelPaths        []path.Path
	LabelFiltersPaths []path.Path
}

// costReportTokenComputedAttrsToInvalidate finds nested label / label_filters that
// must be planned unknown because config omitted them while plan/state still hold
// a prior value (including wrong values copied across reordered attachments).
func costReportTokenComputedAttrsToInvalidate(ctx context.Context, plan, state, config *businessMetricResourceModel, diags *diag.Diagnostics) costReportTokenComputedAttrsToInvalidateResult {
	var result costReportTokenComputedAttrsToInvalidateResult
	if plan.CostReportTokensWithMetadata.IsNull() || plan.CostReportTokensWithMetadata.IsUnknown() {
		return result
	}

	planTokens := make([]*businessMetricResourceModelCostReportToken, 0, len(plan.CostReportTokensWithMetadata.Elements()))
	diags.Append(plan.CostReportTokensWithMetadata.ElementsAs(ctx, &planTokens, false)...)
	if diags.HasError() {
		return result
	}

	stateByToken := map[string]*businessMetricResourceModelCostReportToken{}
	if !state.CostReportTokensWithMetadata.IsNull() && !state.CostReportTokensWithMetadata.IsUnknown() {
		stateTokens := make([]*businessMetricResourceModelCostReportToken, 0, len(state.CostReportTokensWithMetadata.Elements()))
		diags.Append(state.CostReportTokensWithMetadata.ElementsAs(ctx, &stateTokens, false)...)
		if diags.HasError() {
			return result
		}
		for _, token := range stateTokens {
			stateByToken[token.CostReportToken.ValueString()] = token
		}
	}

	configByToken := map[string]*businessMetricResourceModelCostReportToken{}
	if !config.CostReportTokensWithMetadata.IsNull() && !config.CostReportTokensWithMetadata.IsUnknown() {
		configTokens := make([]*businessMetricResourceModelCostReportToken, 0, len(config.CostReportTokensWithMetadata.Elements()))
		diags.Append(config.CostReportTokensWithMetadata.ElementsAs(ctx, &configTokens, false)...)
		if diags.HasError() {
			return result
		}
		for _, token := range configTokens {
			configByToken[token.CostReportToken.ValueString()] = token
		}
	}

	for i, planToken := range planTokens {
		tokenKey := planToken.CostReportToken.ValueString()
		configToken := configByToken[tokenKey]
		stateToken := stateByToken[tokenKey]
		basePath := path.Root("cost_report_tokens_with_metadata").AtListIndex(i)

		if configOmitsString(configToken, func(t *businessMetricResourceModelCostReportToken) types.String { return t.Label }) {
			if shouldInvalidateOmittedComputed(planToken.Label, stateTokenLabel(stateToken), stateTokenCalculationType(stateToken), planToken.CalculationType) {
				result.LabelPaths = append(result.LabelPaths, basePath.AtName("label"))
			}
		}

		if configOmitsMap(configToken, func(t *businessMetricResourceModelCostReportToken) types.Map { return t.LabelFilters }) {
			if shouldInvalidateOmittedComputedMap(planToken.LabelFilters, stateTokenLabelFilters(stateToken)) {
				result.LabelFiltersPaths = append(result.LabelFiltersPaths, basePath.AtName("label_filters"))
			}
		}
	}
	return result
}

func configOmitsString(configToken *businessMetricResourceModelCostReportToken, getter func(*businessMetricResourceModelCostReportToken) types.String) bool {
	if configToken == nil {
		return true
	}
	v := getter(configToken)
	return v.IsNull() || v.IsUnknown()
}

func configOmitsMap(configToken *businessMetricResourceModelCostReportToken, getter func(*businessMetricResourceModelCostReportToken) types.Map) bool {
	if configToken == nil {
		return true
	}
	v := getter(configToken)
	return v.IsNull() || v.IsUnknown()
}

func stateTokenLabel(stateToken *businessMetricResourceModelCostReportToken) types.String {
	if stateToken == nil {
		return types.StringNull()
	}
	return stateToken.Label
}

func stateTokenLabelFilters(stateToken *businessMetricResourceModelCostReportToken) types.Map {
	if stateToken == nil {
		return types.MapNull(types.ListType{ElemType: types.StringType})
	}
	return stateToken.LabelFilters
}

func stateTokenCalculationType(stateToken *businessMetricResourceModelCostReportToken) types.String {
	if stateToken == nil {
		return types.StringNull()
	}
	return stateToken.CalculationType
}

// shouldInvalidateOmittedComputed is true when an omitted Optional+Computed string
// still has a known plan/state value to clear, including reorder mismatches and
// calculation_type changes that should re-derive the API default.
func shouldInvalidateOmittedComputed(planValue, stateValue, stateCalcType, planCalcType types.String) bool {
	if planValue.IsUnknown() {
		return false
	}
	// Known planned value with config omitted — sticky carryover or reorder.
	if !planValue.IsNull() {
		return true
	}
	// Plan null but prior state still has a value — force unknown so apply clears it.
	if !stateValue.IsNull() && !stateValue.IsUnknown() {
		return true
	}
	if !stateCalcType.IsNull() && !stateCalcType.IsUnknown() && !planCalcType.Equal(stateCalcType) {
		return true
	}
	return false
}

func shouldInvalidateOmittedComputedMap(planValue, stateValue types.Map) bool {
	if planValue.IsUnknown() {
		return false
	}
	if !planValue.IsNull() {
		return true
	}
	return !stateValue.IsNull() && !stateValue.IsUnknown()
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
