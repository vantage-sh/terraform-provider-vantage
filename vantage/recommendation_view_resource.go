package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_recommendation_view"
	recviewsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/recommendation_views"
)

var (
	_ resource.Resource                = (*recommendationViewResource)(nil)
	_ resource.ResourceWithConfigure   = (*recommendationViewResource)(nil)
	_ resource.ResourceWithImportState = (*recommendationViewResource)(nil)
)

type recommendationViewResource struct {
	client *Client
}

func NewRecommendationViewResource() resource.Resource {
	return &recommendationViewResource{}
}

func (r *recommendationViewResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r *recommendationViewResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *recommendationViewResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_recommendation_view"
}

func (r *recommendationViewResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_recommendation_view.RecommendationViewResourceSchema(ctx)
	attrs := s.GetAttributes()
	s.Attributes["token"] = schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: attrs["token"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
	resp.Schema = s
}

func (r *recommendationViewResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data recommendationViewResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	model := data.toCreateModel(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := recviewsv2.NewCreateRecommendationViewParams().WithCreateRecommendationView(model)
	out, err := r.client.V2.RecommendationViews.CreateRecommendationView(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*recviewsv2.CreateRecommendationViewBadRequest); ok {
			handleBadRequest("Create RecommendationView Resource", &resp.Diagnostics, e.GetPayload())
			return
		}

		handleError("Create RecommendationView Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *recommendationViewResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data recommendationViewResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := recviewsv2.NewGetRecommendationViewParams().WithRecommendationViewToken(data.Token.ValueString())
	out, err := r.client.V2.RecommendationViews.GetRecommendationView(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*recviewsv2.GetRecommendationViewNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}
		handleError("Get RecommendationView Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *recommendationViewResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data recommendationViewResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model := data.toUpdateModel(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := recviewsv2.NewUpdateRecommendationViewParams().
		WithRecommendationViewToken(data.Token.ValueString()).
		WithUpdateRecommendationView(model)
	out, err := r.client.V2.RecommendationViews.UpdateRecommendationView(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*recviewsv2.UpdateRecommendationViewBadRequest); ok {
			handleBadRequest("Update RecommendationView Resource", &resp.Diagnostics, e.GetPayload())
			return
		}

		handleError("Update RecommendationView Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *recommendationViewResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data recommendationViewResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := recviewsv2.NewDeleteRecommendationViewParams().WithRecommendationViewToken(data.Token.ValueString())
	_, err := r.client.V2.RecommendationViews.DeleteRecommendationView(params, r.client.Auth)
	if err != nil {
		handleError("Delete RecommendationView Resource", &resp.Diagnostics, err)
	}
}
