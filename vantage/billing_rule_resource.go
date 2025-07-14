package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_billing_rule"
	billingrulesv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/billing_rules"
)

var (
	_ resource.Resource                   = (*billingRuleResource)(nil)
	_ resource.ResourceWithConfigure      = (*billingRuleResource)(nil)
	_ resource.ResourceWithValidateConfig = (*billingRuleResource)(nil)
	_ resource.ResourceWithImportState    = (*billingRuleResource)(nil)
)

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
		if data.Percentage.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("percentage"),
				"Missing Attribute Configuration",
				"Expected percentage to be configured with adjustment type",
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

	if data.Type.ValueString() == "custom" {
		if data.SqlQuery.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("sql_query"),
				"Missing Attribute Configuration",
				"Expected sql_query to be present with custom type",
			)
		}
	}
}

func (r *billingRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_billing_rule.BillingRuleResourceSchema(ctx)

	attrs := s.GetAttributes()

	// Override the token field with a PlanModifier
	s.Attributes["token"] = schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: attrs["token"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	resp.Schema = s
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

func (r *billingRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
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
