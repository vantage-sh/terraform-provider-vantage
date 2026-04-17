package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = (*ElasticProviderResource)(nil)
var _ resource.ResourceWithConfigure = (*ElasticProviderResource)(nil)

type ElasticProviderResource struct{ client *Client }

func NewElasticProviderResource() resource.Resource { return &ElasticProviderResource{} }

type ElasticProviderResourceModel struct {
	APIKey types.String `tfsdk:"api_key"`
	Id     types.String `tfsdk:"id"`
}

func (r *ElasticProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *ElasticProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elastic_provider"
}

func (r *ElasticProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Elastic Account Integration.\n\n~> **Note:** This resource is not yet fully supported. Creating or updating this resource will return an error until SDK support is added.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "The Elastic Cloud API key.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the Elastic integration.",
			},
		},
	}
}

func (r *ElasticProviderResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Not Supported", "The Elastic integration is not yet supported by the current vantage-go SDK version.")
}

func (r *ElasticProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ElasticProviderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ElasticProviderResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Not Supported", "The Elastic integration is not yet supported by the current vantage-go SDK version.")
}

func (r *ElasticProviderResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}
