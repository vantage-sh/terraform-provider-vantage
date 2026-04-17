package vantage

import (
"context"

"github.com/hashicorp/terraform-plugin-framework/resource"
"github.com/hashicorp/terraform-plugin-framework/resource/schema"
"github.com/hashicorp/terraform-plugin-framework/types"
)

type MongodbProviderResource struct{ client *Client }

func NewMongodbProviderResource() resource.Resource { return &MongodbProviderResource{} }

type MongodbProviderResourceModel struct {
ClusterUri types.String `tfsdk:"cluster_uri"`
ApiKey     types.String `tfsdk:"api_key"`
Id         types.String `tfsdk:"id"`
}

func (r *MongodbProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
if req.ProviderData == nil {
return
}
r.client = req.ProviderData.(*Client)
}

func (r *MongodbProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
resp.TypeName = req.ProviderTypeName + "_mongodb_provider"
}

func (r *MongodbProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
resp.Schema = schema.Schema{
Attributes: map[string]schema.Attribute{
"cluster_uri": schema.StringAttribute{Required: true},
"api_key":     schema.StringAttribute{Required: true, Sensitive: true},
"id":          schema.StringAttribute{Computed: true},
},
MarkdownDescription: "Manages a MongoDB Account Integration.",
}
}

func (r *MongodbProviderResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
resp.Diagnostics.AddError("Not Supported", "The MongoDB integration is not yet supported by the current vantage-go SDK version.")
}

func (r *MongodbProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
var state MongodbProviderResourceModel
resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *MongodbProviderResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
resp.Diagnostics.AddError("Not Supported", "The MongoDB integration is not yet supported by the current vantage-go SDK version.")
}

func (r *MongodbProviderResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {}
