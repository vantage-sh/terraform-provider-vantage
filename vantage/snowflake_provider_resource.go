package vantage

import (
"context"

"github.com/hashicorp/terraform-plugin-framework/resource"
"github.com/hashicorp/terraform-plugin-framework/resource/schema"
"github.com/hashicorp/terraform-plugin-framework/types"
)

type SnowflakeProviderResource struct{ client *Client }

func NewSnowflakeProviderResource() resource.Resource { return &SnowflakeProviderResource{} }

type SnowflakeProviderResourceModel struct {
AccountName types.String `tfsdk:"account_name"`
UserName    types.String `tfsdk:"user_name"`
Password    types.String `tfsdk:"password"`
Role        types.String `tfsdk:"role"`
Id          types.String `tfsdk:"id"`
}

func (r *SnowflakeProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
if req.ProviderData == nil {
return
}
r.client = req.ProviderData.(*Client)
}

func (r *SnowflakeProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
resp.TypeName = req.ProviderTypeName + "_snowflake_provider"
}

func (r *SnowflakeProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
resp.Schema = schema.Schema{
Attributes: map[string]schema.Attribute{
"account_name": schema.StringAttribute{Required: true},
"user_name":    schema.StringAttribute{Required: true},
"password":     schema.StringAttribute{Required: true, Sensitive: true},
"role":         schema.StringAttribute{Optional: true},
"id":           schema.StringAttribute{Computed: true},
},
MarkdownDescription: "Manages a Snowflake Account Integration.",
}
}

func (r *SnowflakeProviderResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
resp.Diagnostics.AddError("Not Supported", "The Snowflake integration is not yet supported by the current vantage-go SDK version.")
}

func (r *SnowflakeProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
var state SnowflakeProviderResourceModel
resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SnowflakeProviderResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
resp.Diagnostics.AddError("Not Supported", "The Snowflake integration is not yet supported by the current vantage-go SDK version.")
}

func (r *SnowflakeProviderResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {}
