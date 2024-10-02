package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	// resp.Schema = resource_business_metric.BusinessMetricResourceSchema(ctx)
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cost_report_tokens_with_metadata": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cost_report_token": schema.StringAttribute{
							Required:            true,
							Description:         "The token of the CostReport the BusinessMetric is attached to.",
							MarkdownDescription: "The token of the CostReport the BusinessMetric is attached to.",
						},
						"label_filter": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							Description:         "The labels that the BusinessMetric is filtered by within a particular CostReport.",
							MarkdownDescription: "The labels that the BusinessMetric is filtered by within a particular CostReport.",
						},
						"unit_scale": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							Description:         "Determines the scale of the BusinessMetric's values within the CostReport.",
							MarkdownDescription: "Determines the scale of the BusinessMetric's values within the CostReport.",
							Validators: []validator.String{
								stringvalidator.OneOf(
									"per_unit",
									"per_hundred",
									"per_thousand",
									"per_million",
									"per_billion",
								),
							},
							Default: stringdefault.StaticString("per_unit"),
						},
					},
					CustomType: resource_business_metric.CostReportTokensWithMetadataType{
						ObjectType: types.ObjectType{
							AttrTypes: resource_business_metric.CostReportTokensWithMetadataValue{}.AttributeTypes(ctx),
						},
					},
				},
				Optional:            true,
				Computed:            true,
				Description:         "The tokens for any CostReports that use the BusinessMetric, and the unit scale.",
				MarkdownDescription: "The tokens for any CostReports that use the BusinessMetric, and the unit scale.",
			},
			"created_by_token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token of the User who created the BusinessMetric.",
				MarkdownDescription: "The token of the User who created the BusinessMetric.",
			},
			"title": schema.StringAttribute{
				Required:            true,
				Description:         "The title of the BusinessMetrics.",
				MarkdownDescription: "The title of the BusinessMetrics.",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token of the business metric",
				MarkdownDescription: "The token of the business metric",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"values": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"amount": schema.Float64Attribute{
							Required: true,
						},
						"date": schema.StringAttribute{
							Required:            true,
							Description:         "The date of the Business Metric Value. ISO 8601 formatted.",
							MarkdownDescription: "The date of the Business Metric Value. ISO 8601 formatted.",
						},
						"label": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							Description:         "The label of the Business Metric Value.",
							MarkdownDescription: "The label of the Business Metric Value.",
						},
					},
					CustomType: resource_business_metric.ValuesType{
						ObjectType: types.ObjectType{
							AttrTypes: resource_business_metric.ValuesValue{}.AttributeTypes(ctx),
						},
					},
				},
				Optional:            true,
				Computed:            true,
				Description:         "The dates and amounts for the BusinessMetric.",
				MarkdownDescription: "The dates and amounts for the BusinessMetric.",
			},
		},
	}

}

func (r *businessMetricResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// var data resource_business_metric.BusinessMetricModel
	var data *businessMetricResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

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

	diag := data.applyPayload(ctx, out.Payload, false)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *businessMetricResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *businessMetricResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

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

	diag := data.applyPayload(ctx, out.Payload, false)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
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

	diag := data.applyPayload(ctx, out.Payload, false)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
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
