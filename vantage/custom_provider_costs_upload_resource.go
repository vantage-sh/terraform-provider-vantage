package vantage

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv1 "github.com/vantage-sh/vantage-go/vantagev1/models"
	costsv1 "github.com/vantage-sh/vantage-go/vantagev1/vantage/custom_provider_costs"
)

type CustomProviderCostsUploadResource struct{ client *Client }
func NewCustomProviderCostsUploadResource() resource.Resource { return &CustomProviderCostsUploadResource{} }

type CustomProviderCostsUploadResourceModel struct {
	ProviderId types.Int64  `tfsdk:"provider_id"`
	Period     types.String `tfsdk:"period"`
	Content    types.String `tfsdk:"content"`
	Status     types.String `tfsdk:"status"`
	Id         types.Int64  `tfsdk:"id"`
}

func (r *CustomProviderCostsUploadResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_provider_costs_upload"
}

func (r CustomProviderCostsUploadResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"provider_id": schema.Int64Attribute{Required: true, MarkdownDescription: "ID of the custom provider"},
			"period":      schema.StringAttribute{Required: true, MarkdownDescription: "Period (e.g. '2023-12')"},
			"content":     schema.StringAttribute{Required: true, Sensitive: true, MarkdownDescription: "CSV/JSON costs content as string"},
			"status":      schema.StringAttribute{Computed: true, MarkdownDescription: "Import status"},
			"id":          schema.Int64Attribute{Computed: true},
		},
		MarkdownDescription: "Uploads costs for a custom provider.",
	}
}

func (r CustomProviderCostsUploadResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CustomProviderCostsUploadResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() { return }
	params := costsv1.NewUploadCustomProviderCostsParams()
	payload := &modelsv1.UploadCustomProviderCosts{
		ProviderID: int32(data.ProviderId.ValueInt64()),
		Period:     data.Period.ValueStringPointer(),
		Content:    data.Content.ValueStringPointer(), // CSV or JSON format as documented
	}
	params.WithUploadCustomProviderCosts(payload)
	out, err := r.client.V1.CustomProviderCosts.UploadCustomProviderCosts(params, r.client.Auth)
	if err != nil { handleError("Upload Custom Provider Costs", &resp.Diagnostics, err); return }
	data.Status = types.StringValue(out.Payload.Status)
	data.Id = types.Int64Value(int64(out.Payload.ID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Note: read/update/delete for upload jobs is usually not supported, but you could store upload status.
func (r CustomProviderCostsUploadResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CustomProviderCostsUploadResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	// If upload job/status is retrievable, implement here.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r CustomProviderCostsUploadResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Usually not supported for uploads; you might want to error.
}

func (r CustomProviderCostsUploadResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Usually not supported for uploads; consider no-op or error.
}