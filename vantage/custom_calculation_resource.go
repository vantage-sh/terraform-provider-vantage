package vantage

import (
	"context"
	"math"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_custom_calculation"
	calculationsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/custom_calculations"
)

var (
	_ resource.Resource                = (*customCalculationResource)(nil)
	_ resource.ResourceWithConfigure   = (*customCalculationResource)(nil)
	_ resource.ResourceWithImportState = (*customCalculationResource)(nil)
)

type customCalculationResource struct {
	client *Client
}

func NewCustomCalculationResource() resource.Resource {
	return &customCalculationResource{}
}

func (r *customCalculationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r *customCalculationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *customCalculationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_calculation"
}

func (r *customCalculationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_custom_calculation.CustomCalculationResourceSchema(ctx)
	attrs := s.GetAttributes()
	s.Attributes["token"] = schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: attrs["token"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
	description := attrs["description"].(schema.StringAttribute)
	description.Computed = false
	description.Validators = []validator.String{stringvalidator.LengthBetween(1, 10000)}
	description.MarkdownDescription = "An optional description. Remove it from the configuration to clear it."
	s.Attributes["description"] = description

	filter := attrs["filter"].(schema.SingleNestedAttribute)
	for name, attribute := range filter.Attributes {
		switch attribute := attribute.(type) {
		case schema.StringAttribute:
			attribute.Computed = false
			filter.Attributes[name] = attribute
		case schema.ListAttribute:
			attribute.Computed = false
			attribute.MarkdownDescription = "Non-empty list of equivalent usage unit names. Omit to use costs. Units are summed without conversion."
			filter.Attributes[name] = attribute
		}
	}
	s.Attributes["filter"] = filter

	unitScale := attrs["unit_scale"].(schema.Int64Attribute)
	unitScale.Validators = []validator.Int64{int64validator.Between(math.MinInt32, math.MaxInt32)}
	unitScale.MarkdownDescription = "The integer unit scale of the calculation. Defaults to 1."
	s.Attributes["unit_scale"] = unitScale

	reports := attrs["cost_report_tokens"].(schema.SetAttribute)
	reports.MarkdownDescription = "The attached Cost Report tokens. Omit to leave attachments unmanaged or set an empty set to clear them. Manage a given attachment from only one resource."
	s.Attributes["cost_report_tokens"] = reports
	s.MarkdownDescription = "Manages a Custom Calculation."
	resp.Schema = s
}

func (r *customCalculationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data customCalculationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	model := data.toCreateModel(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := calculationsv2.NewCreateCustomCalculationParams().WithCreateCustomCalculation(model)
	out, err := r.client.V2.CustomCalculations.CreateCustomCalculation(params, r.client.Auth)
	if err != nil {
		switch e := err.(type) {
		case *calculationsv2.CreateCustomCalculationBadRequest:
			handleBadRequest("Create Custom Calculation", &resp.Diagnostics, e.GetPayload())
			return
		case *calculationsv2.CreateCustomCalculationUnprocessableEntity:
			handleBadRequest("Create Custom Calculation", &resp.Diagnostics, e.GetPayload())
			return
		}

		handleError("Create CustomCalculation Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *customCalculationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data customCalculationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := calculationsv2.NewGetCustomCalculationParams().WithCustomCalculationToken(data.Token.ValueString())
	out, err := r.client.V2.CustomCalculations.GetCustomCalculation(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*calculationsv2.GetCustomCalculationNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}
		handleError("Get CustomCalculation Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *customCalculationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, prior customCalculationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var configuredReports types.Set
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("cost_report_tokens"), &configuredReports)...)
	model := data.toUpdateModel(ctx, prior, configuredReports, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if model.Title == "" && model.Description == nil && model.DisplayAs == "" && model.Operation == "" &&
		model.UnitScale == nil && model.Filter == nil && model.CostReportTokens == nil {
		resp.State = req.State
		return
	}

	params := calculationsv2.NewUpdateCustomCalculationParams().
		WithCustomCalculationToken(data.Token.ValueString()).
		WithUpdateCustomCalculation(model)
	out, err := r.client.V2.CustomCalculations.UpdateCustomCalculation(params, r.client.Auth)
	if err != nil {
		switch e := err.(type) {
		case *calculationsv2.UpdateCustomCalculationBadRequest:
			handleBadRequest("Update Custom Calculation", &resp.Diagnostics, e.GetPayload())
			return
		case *calculationsv2.UpdateCustomCalculationUnprocessableEntity:
			handleBadRequest("Update Custom Calculation", &resp.Diagnostics, e.GetPayload())
			return
		}

		handleError("Update CustomCalculation Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *customCalculationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data customCalculationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := calculationsv2.NewDeleteCustomCalculationParams().WithCustomCalculationToken(data.Token.ValueString())
	_, err := r.client.V2.CustomCalculations.DeleteCustomCalculation(params, r.client.Auth)
	if _, missing := err.(*calculationsv2.DeleteCustomCalculationNotFound); err != nil && !missing {
		handleError("Delete CustomCalculation Resource", &resp.Diagnostics, err)
	}
}
