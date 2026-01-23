package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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

	resp.Schema = s
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

// processValuesList processes a list of values and converts unknown labels to empty strings.
// Returns the processed list or nil if an error occurred (diagnostics are appended).
func processValuesList(ctx context.Context, tfValues types.List, diags *diag.Diagnostics) *types.List {
	values := make([]*businessMetricResourceModelValue, 0, len(tfValues.Elements()))
	if diag := tfValues.ElementsAs(ctx, &values, false); diag.HasError() {
		diags.Append(diag...)
		return nil
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
		return nil
	}

	return &newList
}

// if labels are unknown in values, sets them to empty string
func assignValues(ctx context.Context, data *businessMetricResourceModel, tfValues types.List, diags *diag.Diagnostics) {
	if newList := processValuesList(ctx, tfValues, diags); newList != nil {
		data.Values = *newList
	}
}

// if labels are unknown in forecasted values, sets them to empty string
func assignForecastedValues(ctx context.Context, data *businessMetricResourceModel, tfValues types.List, diags *diag.Diagnostics) {
	if newList := processValuesList(ctx, tfValues, diags); newList != nil {
		data.ForecastedValues = *newList
	}
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
