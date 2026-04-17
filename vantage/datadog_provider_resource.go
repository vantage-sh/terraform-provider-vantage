package vantage

import (
"context"

"github.com/hashicorp/terraform-plugin-framework/resource"
"github.com/hashicorp/terraform-plugin-framework/resource/schema"
"github.com/hashicorp/terraform-plugin-framework/types"
)

type DatadogProviderResource struct{ client *Client }

func NewDatadogProviderResource() resource.Resource { return &DatadogProviderResource{} }

type DatadogProviderResourceModel struct {
ApiKey types.String `tfsdk:"api_key"`
AppKey types.String `tfsdk:"app_key"`
Id     types.String `tfsdk:"id"`
}

func (r *DatadogProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
if req.ProviderData == nil {
return
}
r.client = req.ProviderData.(*Client)
}

func (r *DatadogProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
resp.TypeName = req.ProviderTypeName + "_datadog_provider"
}

func (r *DatadogProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
resp.Schema = schema.Schema{
Attributes: map[string]schema.Attribute{
"api_key": schema.StringAttribute{Required: true, Sensitive: true},
"app_key": schema.StringAttribute{Required: true, Sensitive: true},
"id":      schema.StringAttribute{Computed: true},
},
MarkdownDescription: "Manages a Datadog Account Integration.",
}
}

func (r *DatadogProviderResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
resp.Diagnostics.AddError("Not Supported", "The Datadog integration is not yet supported by the current vantage-go SDK version.")
}

func (r *DatadogProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
var state DatadogProviderResourceModel
resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DatadogProviderResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
resp.Diagnostics.AddError("Not Supported", "The Datadog integration is not yet supported by the current vantage-go SDK version.")
}

func (r *DatadogProviderResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {}
