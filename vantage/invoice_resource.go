package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_invoice"
	invoicesv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/invoices"
)

var (
	_ resource.Resource                = (*InvoiceResource)(nil)
	_ resource.ResourceWithConfigure   = (*InvoiceResource)(nil)
	_ resource.ResourceWithImportState = (*InvoiceResource)(nil)
)

type InvoiceResource struct {
	client *Client
}

func NewInvoiceResource() resource.Resource {
	return &InvoiceResource{}
}

func (r *InvoiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_invoice"
}

func (r InvoiceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_invoice.InvoiceResourceSchema(ctx)
	attrs := s.GetAttributes()

	// Override the token attribute to add a PlanModifier
	s.Attributes["token"] = schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: attrs["token"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	resp.Schema = s
}

func (r InvoiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *invoiceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := data.toCreate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := invoicesv2.NewCreateInvoiceParams().WithCreateInvoice(body)
	out, err := r.client.V2.Invoices.CreateInvoice(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*invoicesv2.CreateInvoiceBadRequest); ok {
			handleBadRequest("Create Invoice Resource", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Create Invoice Resource", &resp.Diagnostics, err)
		return
	}

	if diag := data.applyPayload(ctx, out.Payload); diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r InvoiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *invoiceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := invoicesv2.NewGetInvoiceParams().WithInvoiceToken(state.Token.ValueString())
	out, err := r.client.V2.Invoices.GetInvoice(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*invoicesv2.GetInvoiceNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Get Invoice Resource", &resp.Diagnostics, err)
		return
	}

	diag := state.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r InvoiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r InvoiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Invoices cannot be updated via the API - they are immutable once created
	resp.Diagnostics.AddError("Update Not Supported", "Invoices cannot be updated once created")
}

func (r InvoiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Invoices cannot be deleted via the API - they are permanent records
	// We'll just remove them from the Terraform state
}

// Configure adds the provider configured client to the data source.
func (r *InvoiceResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}
