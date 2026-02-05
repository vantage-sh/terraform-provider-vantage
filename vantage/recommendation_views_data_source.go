package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	recviewsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/recommendation_views"
)

var (
	_ datasource.DataSource              = &recommendationViewsDataSource{}
	_ datasource.DataSourceWithConfigure = &recommendationViewsDataSource{}
)

func NewRecommendationViewsDataSource() datasource.DataSource {
	return &recommendationViewsDataSource{}
}

type recommendationViewsDataSource struct {
	client *Client
}

type recommendationViewDataSourceModel struct {
	Token             types.String `tfsdk:"token"`
	Title             types.String `tfsdk:"title"`
	WorkspaceToken    types.String `tfsdk:"workspace_token"`
	CreatedAt         types.String `tfsdk:"created_at"`
	CreatedBy         types.String `tfsdk:"created_by"`
	StartDate         types.String `tfsdk:"start_date"`
	EndDate           types.String `tfsdk:"end_date"`
	ProviderIds       types.List   `tfsdk:"provider_ids"`
	BillingAccountIds types.List   `tfsdk:"billing_account_ids"`
	AccountIds        types.List   `tfsdk:"account_ids"`
	Regions           types.List   `tfsdk:"regions"`
	TagKey            types.String `tfsdk:"tag_key"`
	TagValue          types.String `tfsdk:"tag_value"`
}

type recommendationViewsDataSourceModel struct {
	RecommendationViews []recommendationViewDataSourceModel `tfsdk:"recommendation_views"`
}

func (d *recommendationViewsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*Client)
}

func (d *recommendationViewsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_recommendation_views"
}

func (d *recommendationViewsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"recommendation_views": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"token": schema.StringAttribute{
							Computed:            true,
							Description:         "The token of the RecommendationView.",
							MarkdownDescription: "The token of the RecommendationView.",
						},
						"title": schema.StringAttribute{
							Computed:            true,
							Description:         "The title of the RecommendationView.",
							MarkdownDescription: "The title of the RecommendationView.",
						},
						"workspace_token": schema.StringAttribute{
							Computed:            true,
							Description:         "The token for the Workspace the RecommendationView is a part of.",
							MarkdownDescription: "The token for the Workspace the RecommendationView is a part of.",
						},
						"created_at": schema.StringAttribute{
							Computed:            true,
							Description:         "The date and time, in UTC, the view was created. ISO 8601 Formatted.",
							MarkdownDescription: "The date and time, in UTC, the view was created. ISO 8601 Formatted.",
						},
						"created_by": schema.StringAttribute{
							Computed:            true,
							Description:         "The token for the Creator of this RecommendationView.",
							MarkdownDescription: "The token for the Creator of this RecommendationView.",
						},
						"start_date": schema.StringAttribute{
							Computed:            true,
							Description:         "Filter recommendations created on/after this YYYY-MM-DD date.",
							MarkdownDescription: "Filter recommendations created on/after this YYYY-MM-DD date.",
						},
						"end_date": schema.StringAttribute{
							Computed:            true,
							Description:         "Filter recommendations created on/before this YYYY-MM-DD date.",
							MarkdownDescription: "Filter recommendations created on/before this YYYY-MM-DD date.",
						},
						"provider_ids": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							Description:         "Filter by one or more providers.",
							MarkdownDescription: "Filter by one or more providers.",
						},
						"billing_account_ids": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							Description:         "Filter by billing account identifiers.",
							MarkdownDescription: "Filter by billing account identifiers.",
						},
						"account_ids": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							Description:         "Filter by cloud account identifiers.",
							MarkdownDescription: "Filter by cloud account identifiers.",
						},
						"regions": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							Description:         "Filter by region slugs (e.g. us-east-1, eastus, asia-east1).",
							MarkdownDescription: "Filter by region slugs (e.g. us-east-1, eastus, asia-east1).",
						},
						"tag_key": schema.StringAttribute{
							Computed:            true,
							Description:         "Filter by tag key (must be used with tag_value).",
							MarkdownDescription: "Filter by tag key (must be used with tag_value).",
						},
						"tag_value": schema.StringAttribute{
							Computed:            true,
							Description:         "Filter by tag value (requires tag_key).",
							MarkdownDescription: "Filter by tag value (requires tag_key).",
						},
					},
				},
				Computed: true,
			},
		},
	}
}

func (d *recommendationViewsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data recommendationViewsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := recviewsv2.NewGetRecommendationViewsParams()
	out, err := d.client.V2.RecommendationViews.GetRecommendationViews(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get Recommendation Views",
			err.Error(),
		)
		return
	}

	views := []recommendationViewDataSourceModel{}
	for _, rv := range out.Payload.RecommendationViews {
		providerIds, diag := types.ListValueFrom(ctx, types.StringType, rv.ProviderIds)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		billingAccountIds, diag := types.ListValueFrom(ctx, types.StringType, rv.BillingAccountIds)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		accountIds, diag := types.ListValueFrom(ctx, types.StringType, rv.AccountIds)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		regions, diag := types.ListValueFrom(ctx, types.StringType, rv.Regions)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		views = append(views, recommendationViewDataSourceModel{
			Token:             types.StringValue(rv.Token),
			Title:             types.StringValue(rv.Title),
			WorkspaceToken:    types.StringValue(rv.WorkspaceToken),
			CreatedAt:         types.StringValue(rv.CreatedAt),
			CreatedBy:         types.StringValue(rv.CreatedBy),
			StartDate:         types.StringValue(rv.StartDate),
			EndDate:           types.StringValue(rv.EndDate),
			ProviderIds:       providerIds,
			BillingAccountIds: billingAccountIds,
			AccountIds:        accountIds,
			Regions:           regions,
			TagKey:            types.StringValue(rv.TagKey),
			TagValue:          types.StringValue(rv.TagValue),
		})
	}
	data.RecommendationViews = views
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
