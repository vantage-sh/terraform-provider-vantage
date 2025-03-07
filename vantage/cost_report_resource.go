package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	costsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/costs"
)

var (
	_ resource.Resource                = (*CostReportResource)(nil)
	_ resource.ResourceWithConfigure   = (*CostReportResource)(nil)
	_ resource.ResourceWithImportState = (*CostReportResource)(nil)
)

type CostReportResource struct {
	client *Client
}

func NewCostReportResource() resource.Resource {
	return &CostReportResource{}
}

type CostReportResourceModel struct {
	Token                   types.String `tfsdk:"token"`
	Title                   types.String `tfsdk:"title"`
	FolderToken             types.String `tfsdk:"folder_token"`
	Filter                  types.String `tfsdk:"filter"`
	SavedFilterTokens       types.List   `tfsdk:"saved_filter_tokens"`
	WorkspaceToken          types.String `tfsdk:"workspace_token"`
	Groupings               types.String `tfsdk:"groupings"`
	StartDate               types.String `tfsdk:"start_date"`
	EndDate                 types.String `tfsdk:"end_date"`
	PreviousPeriodStartDate types.String `tfsdk:"previous_period_start_date"`
	PreviousPeriodEndDate   types.String `tfsdk:"previous_period_end_date"`
	DateInterval            types.String `tfsdk:"date_interval"`
	ChartType               types.String `tfsdk:"chart_type"`
	DateBin                 types.String `tfsdk:"date_bin"`
}

func (r *CostReportResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cost_report"
}

func (r CostReportResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"title": schema.StringAttribute{
				MarkdownDescription: "Title of the Cost Report",
				Required:            true,
			},
			"folder_token": schema.StringAttribute{
				MarkdownDescription: "Token of the folder this Cost Report resides in.",
				Optional:            true,
				Computed:            true,
			},
			"filter": schema.StringAttribute{
				MarkdownDescription: "Filter query to apply to the Cost Report",
				Optional:            true,
				Computed:            true,
			},
			"groupings": schema.StringAttribute{
				MarkdownDescription: "Grouping aggregations applied to the filtered data.",
				Optional:            true,
				Computed:            true,
			},
			"start_date": schema.StringAttribute{
				MarkdownDescription: "Start date to apply to the Cost Report.",
				Optional:            true,
				Computed:            true,
			},
			"end_date": schema.StringAttribute{
				MarkdownDescription: "End date to apply to the Cost Report.",
				Optional:            true,
				Computed:            true,
			},
			"previous_period_start_date": schema.StringAttribute{
				MarkdownDescription: "Start date to apply to the Cost Report.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"previous_period_end_date": schema.StringAttribute{
				MarkdownDescription: "End date to apply to the Cost Report.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"date_interval": schema.StringAttribute{
				MarkdownDescription: "Date interval to apply to the Cost Report.",
				Optional:            true,
				Computed:            true,
			},
			"chart_type": schema.StringAttribute{
				MarkdownDescription: "Chart type to apply to the Cost Report.",
				Optional:            true,
				Computed:            true,
			},
			"date_bin": schema.StringAttribute{
				MarkdownDescription: "Date bin to apply to the Cost Report.",
				Optional:            true,
				Computed:            true,
			},
			"saved_filter_tokens": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Saved filter tokens to be applied to the Cost Report.",
				Optional:            true,
				Computed:            true,
			},
			"workspace_token": schema.StringAttribute{
				MarkdownDescription: "Workspace token to add the Cost Report to.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"token": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique cost report identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		MarkdownDescription: "Manages a CostReport.",
	}
}

func (r CostReportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *CostReportResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	sft := []types.String{}
	if !data.SavedFilterTokens.IsNull() && !data.SavedFilterTokens.IsUnknown() {
		sft = make([]types.String, 0, len(data.SavedFilterTokens.Elements()))
		resp.Diagnostics.Append(data.SavedFilterTokens.ElementsAs(ctx, &sft, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	params := costsv2.NewCreateCostReportParams()
	body := &modelsv2.CreateCostReport{
		Title:                   data.Title.ValueStringPointer(),
		FolderToken:             data.FolderToken.ValueString(),
		Filter:                  data.Filter.ValueString(),
		Groupings:               data.Groupings.ValueString(),
		SavedFilterTokens:       fromStringsValue(sft),
		WorkspaceToken:          data.WorkspaceToken.ValueString(),
		StartDate:               data.StartDate.ValueString(),
		EndDate:                 data.EndDate.ValueStringPointer(),
		DateInterval:            data.DateInterval.ValueString(),
		PreviousPeriodStartDate: data.PreviousPeriodStartDate.ValueString(),
		PreviousPeriodEndDate:   data.PreviousPeriodEndDate.ValueStringPointer(),
		ChartType:               data.ChartType.ValueStringPointer(),
		DateBin:                 data.DateBin.ValueStringPointer(),
	}
	params.WithCreateCostReport(body)
	out, err := r.client.V2.Costs.CreateCostReport(params, r.client.Auth)
	if err != nil {
		//TODO(macb): Surface 400 errors more clearly.
		handleError("Create Cost Report Resource", &resp.Diagnostics, err)
		return
	}

	data.Token = types.StringValue(out.Payload.Token)
	data.Filter = types.StringValue(out.Payload.Filter)
	data.Groupings = types.StringValue(out.Payload.Groupings)
	data.StartDate = types.StringValue(out.Payload.StartDate)
	data.EndDate = types.StringValue(out.Payload.EndDate)
	data.PreviousPeriodStartDate = types.StringValue(out.Payload.PreviousPeriodStartDate)
	data.PreviousPeriodEndDate = types.StringValue(out.Payload.PreviousPeriodEndDate)
	data.DateInterval = types.StringValue(out.Payload.DateInterval)
	data.ChartType = types.StringValue(out.Payload.ChartType)
	data.DateBin = types.StringValue(out.Payload.DateBin)
	data.FolderToken = types.StringValue(out.Payload.FolderToken)
	data.WorkspaceToken = types.StringValue(out.Payload.WorkspaceToken)
	savedFilterTokensValue, diag := types.ListValueFrom(ctx, types.StringType, out.Payload.SavedFilterTokens)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	data.SavedFilterTokens = savedFilterTokensValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r CostReportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *CostReportResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := costsv2.NewGetCostReportParams()
	params.SetCostReportToken(state.Token.ValueString())
	out, err := r.client.V2.Costs.GetCostReport(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*costsv2.GetCostReportNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Get Cost Report Resource", &resp.Diagnostics, err)
		return
	}

	state.Token = types.StringValue(out.Payload.Token)
	state.Filter = types.StringValue(out.Payload.Filter)
	state.Title = types.StringValue(out.Payload.Title)
	state.Filter = types.StringValue(out.Payload.Filter)
	state.Groupings = types.StringValue(out.Payload.Groupings)
	state.PreviousPeriodStartDate = types.StringValue(out.Payload.PreviousPeriodStartDate)
	state.PreviousPeriodEndDate = types.StringValue(out.Payload.PreviousPeriodEndDate)
	state.DateInterval = types.StringValue(out.Payload.DateInterval)
	// if out.Payload.DateInterval == "custom" {
	state.StartDate = types.StringValue(out.Payload.StartDate)
	state.EndDate = types.StringValue(out.Payload.EndDate)
	// }
	state.ChartType = types.StringValue(out.Payload.ChartType)
	state.DateBin = types.StringValue(out.Payload.DateBin)
	state.WorkspaceToken = types.StringValue(out.Payload.WorkspaceToken)
	state.FolderToken = types.StringValue(out.Payload.FolderToken)
	savedFilterTokensValue, diag := types.ListValueFrom(ctx, types.StringType, out.Payload.SavedFilterTokens)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	state.SavedFilterTokens = savedFilterTokensValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r CostReportResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r CostReportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *CostReportResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sft := []types.String{}
	if !data.SavedFilterTokens.IsNull() && !data.SavedFilterTokens.IsUnknown() {
		sft = make([]types.String, 0, len(data.SavedFilterTokens.Elements()))
		resp.Diagnostics.Append(data.SavedFilterTokens.ElementsAs(ctx, &sft, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	params := costsv2.NewUpdateCostReportParams()
	params.WithCostReportToken(data.Token.ValueString())
	model := &modelsv2.UpdateCostReport{
		FolderToken:             data.FolderToken.ValueString(),
		Title:                   data.Title.ValueString(),
		Filter:                  data.Filter.ValueString(),
		SavedFilterTokens:       fromStringsValue(sft),
		Groupings:               data.Groupings.ValueString(),
		PreviousPeriodStartDate: data.PreviousPeriodStartDate.ValueString(),
		PreviousPeriodEndDate:   data.PreviousPeriodEndDate.ValueString(),
		ChartType:               data.ChartType.ValueStringPointer(),
		DateBin:                 data.DateBin.ValueStringPointer(),
	}

	if data.DateInterval.ValueString() == "custom" {
		model.StartDate = data.StartDate.ValueString()
		model.EndDate = data.EndDate.ValueString()
		model.DateInterval = "custom"
	} else {
		model.DateInterval = data.DateInterval.ValueString()
	}

	params.WithUpdateCostReport(model)
	out, err := r.client.V2.Costs.UpdateCostReport(params, r.client.Auth)
	if err != nil {
		handleError("Update Cost Report Resource", &resp.Diagnostics, err)
		return
	}

	data.Title = types.StringValue(out.Payload.Title)
	data.FolderToken = types.StringValue(out.Payload.FolderToken)
	data.Filter = types.StringValue(out.Payload.Filter)
	data.Groupings = types.StringValue(out.Payload.Groupings)
	data.WorkspaceToken = types.StringValue(out.Payload.WorkspaceToken)
	data.StartDate = types.StringValue(out.Payload.StartDate)
	data.EndDate = types.StringValue(out.Payload.EndDate)
	data.PreviousPeriodStartDate = types.StringValue(out.Payload.PreviousPeriodStartDate)
	data.PreviousPeriodEndDate = types.StringValue(out.Payload.PreviousPeriodEndDate)
	data.DateInterval = types.StringValue(out.Payload.DateInterval)
	data.ChartType = types.StringValue(out.Payload.ChartType)
	data.DateBin = types.StringValue(out.Payload.DateBin)
	savedFilterTokensValue, diag := types.ListValueFrom(ctx, types.StringType, out.Payload.SavedFilterTokens)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	data.SavedFilterTokens = savedFilterTokensValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r CostReportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *CostReportResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := costsv2.NewDeleteCostReportParams()
	params.SetCostReportToken(state.Token.ValueString())
	_, err := r.client.V2.Costs.DeleteCostReport(params, r.client.Auth)
	if err != nil {
		handleError("Delete Cost Report Resource", &resp.Diagnostics, err)
	}
}

// Configure adds the provider configured client to the data source.
func (r *CostReportResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}

func (r *CostReportResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("folder_token"),
			path.MatchRoot("workspace_token"),
		),
	}
}
