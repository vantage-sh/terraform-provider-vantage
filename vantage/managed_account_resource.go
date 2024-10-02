package vantage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	managedaccountsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/managed_accounts"
)

var (
	_ resource.Resource                = (*managedAccountResource)(nil)
	_ resource.ResourceWithConfigure   = (*managedAccountResource)(nil)
	_ resource.ResourceWithImportState = (*managedAccountResource)(nil)
)

func NewManagedAccountResource() resource.Resource {
	return &managedAccountResource{}
}

type managedAccountResource struct {
	client *Client
}

func (r *managedAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *managedAccountResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_account"
}

func (r *managedAccountResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_credential_tokens": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "Access Credential (aka Integrations) tokens to assign to the Managed Account.",
				MarkdownDescription: "Access Credential (aka Integrations) tokens to assign to the Managed Account.",
			},
			"billing_rule_tokens": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "Billing Rule tokens to assign to the Managed Account.",
				MarkdownDescription: "Billing Rule tokens to assign to the Managed Account.",
			},
			"contact_email": schema.StringAttribute{
				Required:            true,
				Description:         "The contact email address for the Managed Account.",
				MarkdownDescription: "The contact email address for the Managed Account.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the Managed Account.",
				MarkdownDescription: "The name of the Managed Account.",
			},
			"parent_account_token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token for the parent Account.",
				MarkdownDescription: "The token for the parent Account.",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token of the managed account",
				MarkdownDescription: "The token of the managed account",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}

}

func (r *managedAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *managedAccountModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API call logic
	model := data.toCreateModel(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := managedaccountsv2.NewCreateManagedAccountParams().WithCreateManagedAccount(model)
	out, err := r.client.V2.ManagedAccounts.CreateManagedAccount(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*managedaccountsv2.CreateManagedAccountBadRequest); ok {
			handleBadRequest("Create Managed Account Resource", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Create Managed Account Resource", &resp.Diagnostics, err)
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

func (r *managedAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *managedAccountModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := managedaccountsv2.NewGetManagedAccountParams().WithManagedAccountToken(data.Token.ValueString())
	out, err := r.client.V2.ManagedAccounts.GetManagedAccount(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*managedaccountsv2.GetManagedAccountNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Get Managed Account Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload, false)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r *managedAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *managedAccountModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model := data.toUpdateModel(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := managedaccountsv2.NewUpdateManagedAccountParams().
		WithManagedAccountToken(data.Token.ValueString()).
		WithUpdateManagedAccount(model)

	out, err := r.client.V2.ManagedAccounts.UpdateManagedAccount(params, r.client.Auth)

	if err != nil {
		handleError("Update Managed Account Resource", &resp.Diagnostics, err)
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("update payload: %v", out.Payload))
	diag := data.applyPayload(ctx, out.Payload, false)

	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *managedAccountModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := managedaccountsv2.NewDeleteManagedAccountParams().
		WithManagedAccountToken(data.Token.ValueString())

	_, err := r.client.V2.ManagedAccounts.DeleteManagedAccount(params, r.client.Auth)
	if err != nil {
		handleError("Delete Managed Account Resource", &resp.Diagnostics, err)
		return
	}
}
