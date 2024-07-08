package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	billingrulesv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/billing_rules"
)

var _ resource.Resource = (*billingRuleResource)(nil)
var _ resource.ResourceWithConfigure = (*billingRuleResource)(nil)
var _ resource.ResourceWithValidateConfig = (*billingRuleResource)(nil)

func NewBillingRuleResource() resource.Resource {
	return &billingRuleResource{}
}

type billingRuleResource struct {
	client *Client
}

func (r *billingRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *billingRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_billing_rule"
}

func (r *billingRuleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {

	var data billingRuleModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Type.ValueString() == "exclusion" && data.ChargeType.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("charge_type"),
			"Missing Attribute Configuration",
			"Expected charge_type to be configured with exclusion type",
		)
	}

	if data.Type.ValueString() == "adjustment" {
		if data.AdjustedRate.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("adjusted_rate"),
				"Missing Attribute Configuration",
				"Expected adjusted_rate to be configured with adjustment type",
			)
		}

		if data.Service.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("service"),
				"Missing Attribute Configuration",
				"Expected service to be configured with adjustment type",
			)
		}

		if data.Category.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("category"),
				"Missing Attribute Configuration",
				"Expected category to be configured with adjustment type",
			)
		}
	}

	if data.Type.ValueString() == "credit" || data.Type.ValueString() == "charge" {
		if data.Service.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("service"),
				"Missing Attribute Configuration",
				"Expected service to be configured with credit or charge type",
			)
		}

		if data.Category.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("category"),
				"Missing Attribute Configuration",
				"Expected category to be configured with credit or charge type",
			)
		}

		if data.SubCategory.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("sub_category"),
				"Missing Attribute Configuration",
				"Expected sub_category to be configured with credit or charge type",
			)
		}

		if data.StartPeriod.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("start_period"),
				"Missing Attribute Configuration",
				"Expected start_period to be configured with credit or charge type",
			)
		}

		if data.Amount.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("amount"),
				"Missing Attribute Configuration",
				"Expected amount to be configured with credit or charge type",
			)
		}
	}

}

func (r *billingRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"adjusted_rate": schema.Float64Attribute{
				Optional:            true,
				Description:         "The adjusted rate of the Billing Rule. Example value: 0.85",
				MarkdownDescription: "The adjusted rate of the Billing Rule. Example value: 0.85",
			},
			"amount": schema.Float64Attribute{
				Optional:            true,
				Description:         "The credit amount for the Billing Rule. Example value: 300",
				MarkdownDescription: "The credit amount for the Billing Rule. Example value: 300",
			},
			"category": schema.StringAttribute{
				Optional:            true,
				Description:         "The category of the Billing Rule.",
				MarkdownDescription: "The category of the Billing Rule.",
			},
			"charge_type": schema.StringAttribute{
				Optional:            true,
				Description:         "The charge type of the Billing Rule.",
				MarkdownDescription: "The charge type of the Billing Rule.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The date and time, in UTC, the Billing Rule was created. ISO 8601 Formatted.",
				MarkdownDescription: "The date and time, in UTC, the Billing Rule was created. ISO 8601 Formatted.",
			},
			"created_by_token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token of the User who created the Billing Rule.",
				MarkdownDescription: "The token of the User who created the Billing Rule.",
			},
			"service": schema.StringAttribute{
				Optional:            true,
				Description:         "The service of the Billing Rule.",
				MarkdownDescription: "The service of the Billing Rule.",
			},
			"start_period": schema.StringAttribute{
				Optional:            true,
				Description:         "The start period of the Billing Rule.",
				MarkdownDescription: "The start period of the Billing Rule.",
			},
			"sub_category": schema.StringAttribute{
				Optional:            true,
				Description:         "The subcategory of the Billing Rule.",
				MarkdownDescription: "The subcategory of the Billing Rule.",
			},
			"title": schema.StringAttribute{
				Required:            true,
				Description:         "The title of the Billing Rule.",
				MarkdownDescription: "The title of the Billing Rule.",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token of the billing rule",
				MarkdownDescription: "The token of the billing rule",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Required:            true,
				Description:         "The type of the Billing Rule. Note: the values are case insensitive.",
				MarkdownDescription: "The type of the Billing Rule. Note: the values are case insensitive.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"exclusion",
						"adjustment",
						"credit",
						"charge",
					),
				},
			},
		},
	}
}

func (r *billingRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data billingRuleModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API call logic
	model := data.toCreateModel(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// Create API call logic
	params := billingrulesv2.NewCreateBillingRuleParams().WithCreateBillingRule(model)
	out, err := r.client.V2.BillingRules.CreateBillingRule(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*billingrulesv2.CreateBillingRuleBadRequest); ok {
			handleBadRequest("Create BillingRule Resource", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Create BillingRule Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *billingRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data billingRuleModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := billingrulesv2.NewGetBillingRuleParams().WithBillingRuleToken(data.Token.ValueString())
	out, err := r.client.V2.BillingRules.GetBillingRule(params, r.client.Auth)

	if err != nil {
		if _, ok := err.(*billingrulesv2.GetBillingRuleNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}
		handleError("Read BillingRule Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *billingRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data billingRuleModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	model := data.toUpdateModel(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := billingrulesv2.NewUpdateBillingRuleParams().WithUpdateBillingRule(model).WithBillingRuleToken(data.Token.ValueString())
	out, err := r.client.V2.BillingRules.UpdateBillingRule(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*billingrulesv2.UpdateBillingRuleBadRequest); ok {
			handleBadRequest("Update BillingRule Resource", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Update BillingRule Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *billingRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data billingRuleModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := billingrulesv2.NewDeleteBillingRuleParams().WithBillingRuleToken(data.Token.ValueString())
	_, err := r.client.V2.BillingRules.DeleteBillingRule(params, r.client.Auth)
	if err != nil {
		handleError("Delete BillingRule Resource", &resp.Diagnostics, err)
		return
	}

}
