package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/planmodifiers"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_budget_alert"
	budgetalertsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/budget_alerts"
)

var (
	_ resource.Resource                = (*budgetAlertResource)(nil)
	_ resource.ResourceWithConfigure   = (*budgetAlertResource)(nil)
	_ resource.ResourceWithImportState = (*budgetAlertResource)(nil)
)

func NewBudgetAlertResource() resource.Resource {
	return &budgetAlertResource{}
}

type budgetAlertResource struct {
	client *Client
}

func (r *budgetAlertResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *budgetAlertResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_budget_alert"
}

// The API requires at least one recipient, but only at create time. A config
// validator would also run on update, where the recipient fields are allowed
// to be absent because they then resolve from prior state.
func validateBudgetAlertRecipients(ctx context.Context, config tfsdk.Config, diags *diag.Diagnostics) {
	var cfg budgetAlertModel
	diags.Append(config.Get(ctx, &cfg)...)
	if diags.HasError() {
		return
	}

	for _, list := range []types.List{cfg.UserTokens, cfg.RecipientEmails, cfg.RecipientChannels} {
		// The value comes from elsewhere in the config and cannot be judged yet.
		if list.IsUnknown() {
			return
		}
		// An empty list is no better than an absent one here: the payload sends
		// it as null and the API rejects the create.
		if !list.IsNull() && len(list.Elements()) > 0 {
			return
		}
	}

	diags.AddError(
		"Missing Budget Alert Recipients",
		"At least one of user_tokens, recipient_emails, or recipient_channels must be set to a non-empty list when creating a budget alert.",
	)
}

func (r *budgetAlertResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.schema(ctx)
}

func (r *budgetAlertResource) schema(ctx context.Context) schema.Schema {
	s := resource_budget_alert.BudgetAlertResourceSchema(ctx)
	attrs := s.GetAttributes()

	s.Attributes["token"] = schema.StringAttribute{
		Computed:            true,
		Description:         attrs["token"].GetDescription(),
		MarkdownDescription: attrs["token"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	s.Attributes["workspace_token"] = schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		Description:         "The token of the Workspace to add the BudgetAlert to. Required if the API token is associated with multiple Workspaces.",
		MarkdownDescription: "The token of the Workspace to add the BudgetAlert to. Required if the API token is associated with multiple Workspaces.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	s.Attributes["period_to_track"] = schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		Description:         attrs["period_to_track"].GetDescription(),
		MarkdownDescription: attrs["period_to_track"].GetMarkdownDescription(),
		Validators: []validator.String{
			stringvalidator.OneOf("start_of_the_month", "end_of_the_month"),
		},
	}
	s.Attributes["duration_in_days"] = schema.StringAttribute{
		Required:            true,
		Description:         "The number of days from the start or end of the month to trigger the alert if the threshold is reached. Use an empty string for the full month. This write attribute is a string because the API uses an empty-string sentinel; the budget_alerts data source returns the API's nullable integer response.",
		MarkdownDescription: "The number of days from the start or end of the month to trigger the alert if the threshold is reached. Use `\"\"` for the full month. This write attribute is a string because the API uses an empty-string sentinel; the `vantage_budget_alerts` data source returns the API's nullable integer response.",
	}

	// All three recipient fields are Optional+Computed, and an update must keep
	// the omitted ones at their prior values: the payload cannot omit a field,
	// and the API reads a null recipient key as an instruction to drop those
	// recipients.
	//
	// user_tokens and recipient_emails are derived from each other, so a change
	// to one means the API recomputes the other and its prior value has to be
	// given up. recipient_channels is independent of both, so it keeps its prior
	// value regardless of what the others do.
	recipientModifiers := map[string][]planmodifier.List{
		"user_tokens":        {planmodifiers.ListUseStateUnlessSiblingsChange(path.Root("recipient_emails"))},
		"recipient_emails":   {planmodifiers.ListUseStateUnlessSiblingsChange(path.Root("user_tokens"))},
		"recipient_channels": {listplanmodifier.UseStateForUnknown()},
	}

	for name, modifiers := range recipientModifiers {
		description := attrs[name].GetDescription()
		switch name {
		case "user_tokens":
			description = "The tokens of organization users that receive the alert. The API also exposes their addresses in recipient_emails; freeform verified-domain addresses appear only in recipient_emails."
		case "recipient_emails":
			description = "The complete list of email addresses that receive the alert, including addresses derived from user_tokens and freeform addresses on verified domains."
		}
		description += " At least one recipient list must resolve to a non-empty value when the alert is created."

		s.Attributes[name] = schema.ListAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			Computed:            true,
			Description:         description,
			MarkdownDescription: description,
			PlanModifiers:       modifiers,
		}
	}

	return s
}

func (r *budgetAlertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	validateBudgetAlertRecipients(ctx, req.Config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var data *budgetAlertModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := data.toCreate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := budgetalertsv2.NewCreateBudgetAlertParams().WithCreateBudgetAlert(input)
	out, err := r.client.V2.BudgetAlerts.CreateBudgetAlert(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*budgetalertsv2.CreateBudgetAlertBadRequest); ok {
			handleBadRequest("Create Budget Alert", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Create Budget Alert", &resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(data.applyPayload(ctx, out.Payload)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *budgetAlertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *budgetAlertModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := budgetalertsv2.NewGetBudgetAlertParams().WithBudgetAlertToken(data.Token.ValueString())
	out, err := r.client.V2.BudgetAlerts.GetBudgetAlert(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*budgetalertsv2.GetBudgetAlertNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}
		handleError("Read Budget Alert", &resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(data.applyPayload(ctx, out.Payload)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *budgetAlertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *budgetAlertModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := data.toUpdate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := budgetalertsv2.NewUpdateBudgetAlertParams().
		WithBudgetAlertToken(data.Token.ValueString()).
		WithUpdateBudgetAlert(input)

	out, err := r.client.V2.BudgetAlerts.UpdateBudgetAlert(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*budgetalertsv2.UpdateBudgetAlertBadRequest); ok {
			handleBadRequest("Update Budget Alert", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Update Budget Alert", &resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(data.applyPayload(ctx, out.Payload)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *budgetAlertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *budgetAlertModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := budgetalertsv2.NewDeleteBudgetAlertParams().WithBudgetAlertToken(data.Token.ValueString())
	if _, err := r.client.V2.BudgetAlerts.DeleteBudgetAlert(params, r.client.Auth); err != nil {
		handleError("Delete Budget Alert", &resp.Diagnostics, err)
	}
}

func (r *budgetAlertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}
