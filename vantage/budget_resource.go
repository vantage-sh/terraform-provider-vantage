package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_budget"
	budgetsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/budgets"
)

var (
	_ resource.Resource                = (*budgetResource)(nil)
	_ resource.ResourceWithConfigure   = (*budgetResource)(nil)
	_ resource.ResourceWithImportState = (*budgetResource)(nil)
)

func NewBudgetResource() resource.Resource {
	return &budgetResource{}
}

type budgetResource struct {
	client *Client
}

func (r *budgetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}
func (r *budgetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_budget"
}

func (r *budgetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"budget_alert_tokens": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "The tokens of the BudgetAlerts associated with the Budget.",
				MarkdownDescription: "The tokens of the BudgetAlerts associated with the Budget.",
			},
			"cost_report_token": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The CostReport token.",
				MarkdownDescription: "The CostReport token.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The date and time, in UTC, the Budget was created. ISO 8601 Formatted.",
				MarkdownDescription: "The date and time, in UTC, the Budget was created. ISO 8601 Formatted.",
			},
			"created_by_token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token of the Creator of the Budget.",
				MarkdownDescription: "The token of the Creator of the Budget.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the Budget.",
				MarkdownDescription: "The name of the Budget.",
			},
			"performance": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"actual": schema.StringAttribute{
							Computed:            true,
							Description:         "The date and time, in UTC, the Budget was created. ISO 8601 Formatted.",
							MarkdownDescription: "The date and time, in UTC, the Budget was created. ISO 8601 Formatted.",
						},
						"amount": schema.StringAttribute{
							Computed:            true,
							Description:         "The amount of the Budget Period as a string to ensure precision.",
							MarkdownDescription: "The amount of the Budget Period as a string to ensure precision.",
						},
						"date": schema.StringAttribute{
							Computed:            true,
							Description:         "The date and time, in UTC, the Budget was created. ISO 8601 Formatted.",
							MarkdownDescription: "The date and time, in UTC, the Budget was created. ISO 8601 Formatted.",
						},
					},
					CustomType: resource_budget.PerformanceType{
						ObjectType: types.ObjectType{
							AttrTypes: resource_budget.PerformanceValue{}.AttributeTypes(ctx),
						},
					},
				},
				Computed:            true,
				Description:         "The historical performance of the Budget.",
				MarkdownDescription: "The historical performance of the Budget.",
			},
			"periods": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"amount": schema.Float64Attribute{
							Required:            true,
							Description:         "The amount of the period.",
							MarkdownDescription: "The amount of the period.",
						},
						"end_at": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							Description:         "The end date of the period.",
							MarkdownDescription: "The end date of the period.",
						},
						"start_at": schema.StringAttribute{
							Required:            true,
							Description:         "The start date of the period.",
							MarkdownDescription: "The start date of the period.",
						},
					},
					CustomType: resource_budget.PeriodsType{
						ObjectType: types.ObjectType{
							AttrTypes: resource_budget.PeriodsValue{}.AttributeTypes(ctx),
						},
					},
				},
				Optional:            true,
				Computed:            true,
				Description:         "The periods for the Budget. The start_at and end_at must be iso8601 formatted e.g. YYYY-MM-DD.",
				MarkdownDescription: "The periods for the Budget. The start_at and end_at must be iso8601 formatted e.g. YYYY-MM-DD.",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token of the budget",
				MarkdownDescription: "The token of the budget",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token for the User who created this Budget.",
				MarkdownDescription: "The token for the User who created this Budget.",
			},

			"workspace_token": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The token of the Workspace to add the Budget to.",
				MarkdownDescription: "The token of the Workspace to add the Budget to.",
			},
		},
	}

}

func (r *budgetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data budgetModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := budgetsv2.NewCreateBudgetParams().WithCreateBudget(toCreateModel(ctx, &resp.Diagnostics, data))
	out, err := r.client.V2.Budgets.CreateBudget(params, r.client.Auth)

	if err != nil {
		if e, ok := err.(*budgetsv2.CreateBudgetBadRequest); ok {
			handleBadRequest("Create Budget", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Create Budget", &resp.Diagnostics, err)
		return
	}

	tflog.Debug(ctx, "applyBudgetPayload create")
	diag := applyBudgetPayload(ctx, false, out.Payload, &data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *budgetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data budgetModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := budgetsv2.NewGetBudgetParams().WithBudgetToken(data.Token.ValueString())
	out, err := r.client.V2.Budgets.GetBudget(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*budgetsv2.GetBudgetNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}
		handleError("Get Budget", &resp.Diagnostics, err)
		return
	}
	tflog.Debug(ctx, "applyBudgetPayload read")
	diag := applyBudgetPayload(ctx, false, out.Payload, &data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *budgetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r *budgetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data budgetModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := budgetsv2.NewUpdateBudgetParams().WithUpdateBudget(toUpdateModel(ctx, &resp.Diagnostics, data)).WithBudgetToken(data.Token.ValueString())
	out, err := r.client.V2.Budgets.UpdateBudget(params, r.client.Auth)

	if err != nil {
		if e, ok := err.(*budgetsv2.UpdateBudgetBadRequest); ok {
			handleBadRequest("Update Budget", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Update Budget", &resp.Diagnostics, err)
		return
	}
	tflog.Debug(ctx, "applyBudgetPayload update")
	diag := applyBudgetPayload(ctx, false, out.Payload, &data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *budgetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data budgetModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := budgetsv2.NewDeleteBudgetParams().WithBudgetToken(data.Token.ValueString())
	_, err := r.client.V2.Budgets.DeleteBudget(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*budgetsv2.DeleteBudgetNotFound); ok {
			handleBadRequest("Delete Budget", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Delete Budget", &resp.Diagnostics, err)
		return
	}

}
