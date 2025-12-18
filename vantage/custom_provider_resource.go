package vantage

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv1 "github.com/vantage-sh/vantage-go/vantagev1/models"
	integrationsv1 "github.com/vantage-sh/vantage-go/vantagev1/vantage/integrations"
)

type CustomProviderResource struct{ client *Client }
func NewCustomProviderResource() resource.Resource { return &CustomProviderResource{} }

type CustomProviderResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Identifier  types.String `tfsdk:"identifier"`
	Id          types.Int64  `tfsdk:"id"`
}

func (r *CustomProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_provider"
}

func (r CustomProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name":       schema.StringAttribute{Required: true, MarkdownDescription: "The display name for your custom provider"},
			"identifier": schema.StringAttribute{Required: true, MarkdownDescription: "A unique identifier for this custom provider"},
			"id":         schema.Int64Attribute{Computed: true, MarkdownDescription: "Provider identifier"},
		},
		MarkdownDescription: "Manages a Custom Provider integration.",
	}
}

func (r CustomProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CustomProviderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() { return }
	params := integrationsv1.NewCreateCustomProviderParams()
	payload := &modelsv1.CreateCustomProvider{
		Name:       data.Name.ValueStringPointer(),
		Identifier: data.Identifier.ValueStringPointer(),
	}
	params.WithCreateCustomProvider(payload)
	out, err := r.client.V1.Integrations.CreateCustomProvider(params, r.client.Auth)
	if err != nil { handleError("Create Custom Provider", &resp.Diagnostics, err); return }
	data.Id = types.Int64Value(int64(out.Payload.ID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r CustomProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CustomProviderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }
	params := integrationsv1.NewGetCustomProviderParams()
	params.SetProviderID(int32(state.Id.ValueInt64()))
	out, err := r.client.V1.Integrations.GetCustomProvider(params, r.client.Auth)
	if err != nil { handleError("Read Custom Provider", &resp.Diagnostics, err); return }
	state.Name = types.StringValue(out.Payload.Name)
	state.Identifier = types.StringValue(out.Payload.Identifier)
	state.Id = types.Int64Value(int64(out.Payload.ID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r CustomProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CustomProviderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() { return }
	params := integrationsv1.NewUpdateCustomProviderParams()
	params.SetProviderID(int32(plan.Id.ValueInt64()))
	payload := &modelsv1.UpdateCustomProvider{
		Name:       plan.Name.ValueStringPointer(),
		Identifier: plan.Identifier.ValueStringPointer(),
	}
	params.WithUpdateCustomProvider(payload)
	out, err := r.client.V1.Integrations.UpdateCustomProvider(params, r.client.Auth)
	if err != nil { handleError("Update Custom Provider", &resp.Diagnostics, err); return }
	plan.Id = types.Int64Value(int64(out.Payload.ID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r CustomProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CustomProviderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }
	params := integrationsv1.NewDeleteCustomProviderParams()
	params.SetProviderID(int32(state.Id.ValueInt64()))
	_, err := r.client.V1.Integrations.DeleteCustomProvider(params, r.client.Auth)
	if err != nil { handleError("Delete Custom Provider", &resp.Diagnostics, err); return }
}