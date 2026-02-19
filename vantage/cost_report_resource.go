package vantage

import (
	"context"
	"fmt"

	goaruntime "github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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

type CostReportSettingsModel struct {
	IncludeCredits     types.Bool   `tfsdk:"include_credits"`
	IncludeRefunds     types.Bool   `tfsdk:"include_refunds"`
	IncludeDiscounts   types.Bool   `tfsdk:"include_discounts"`
	IncludeTax         types.Bool   `tfsdk:"include_tax"`
	Amortize           types.Bool   `tfsdk:"amortize"`
	Unallocated        types.Bool   `tfsdk:"unallocated"`
	AggregateBy        types.String `tfsdk:"aggregate_by"`
	ShowPreviousPeriod types.Bool   `tfsdk:"show_previous_period"`
}

type CostReportResourceModel struct {
	Token                   types.String `tfsdk:"token"`
	Id                      types.String `tfsdk:"id"`
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
	Settings                types.Object `tfsdk:"settings"`
}

// costReportSettingsAttrTypes defines the attribute types for the settings object.
func costReportSettingsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"include_credits":      types.BoolType,
		"include_refunds":      types.BoolType,
		"include_discounts":    types.BoolType,
		"include_tax":          types.BoolType,
		"amortize":             types.BoolType,
		"unallocated":          types.BoolType,
		"aggregate_by":         types.StringType,
		"show_previous_period": types.BoolType,
	}
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
				// https://discuss.hashicorp.com/t/framework-migration-test-produces-non-empty-plan/54523/8
				Default: stringdefault.StaticString(""),
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
			"settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Settings for the Cost Report.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"include_credits": schema.BoolAttribute{
						MarkdownDescription: "Report will include credits.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
					},
					"include_refunds": schema.BoolAttribute{
						MarkdownDescription: "Report will include refunds.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
					},
					"include_discounts": schema.BoolAttribute{
						MarkdownDescription: "Report will include discounts.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
					},
					"include_tax": schema.BoolAttribute{
						MarkdownDescription: "Report will include tax.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
					},
					"amortize": schema.BoolAttribute{
						MarkdownDescription: "Report will amortize.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"unallocated": schema.BoolAttribute{
						MarkdownDescription: "Report will show unallocated costs.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"aggregate_by": schema.StringAttribute{
						MarkdownDescription: "Report will aggregate by cost or usage.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("cost"),
						Validators: []validator.String{
							stringvalidator.OneOf("cost", "usage"),
						},
					},
					"show_previous_period": schema.BoolAttribute{
						MarkdownDescription: "Report will show previous period costs or usage comparison.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
				},
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
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique cost report identifier (aliases to token)",
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
	}

	if !data.ChartType.IsUnknown() && !data.ChartType.IsNull() {
		body.ChartType = data.ChartType.ValueStringPointer()
	}

	if !data.DateBin.IsUnknown() && !data.DateBin.IsNull() {
		body.DateBin = data.DateBin.ValueStringPointer()
	}

	params.WithCreateCostReport(body)
	out, err := r.client.V2.Costs.CreateCostReport(params, r.client.Auth)
	if err != nil {
		handleError("Create Cost Report Resource", &resp.Diagnostics, err)
		return
	}

	token := out.Payload.Token

	// The Create API does not apply settings, so we follow up with a raw
	// update call to ensure settings are persisted.
	if !data.Settings.IsNull() && !data.Settings.IsUnknown() {
		var settings CostReportSettingsModel
		resp.Diagnostics.Append(data.Settings.As(ctx, &settings, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		updated, updateErr := r.updateCostReportSettingsRaw(token, settings)
		if updateErr != nil {
			handleError("Create Cost Report Resource (settings update)", &resp.Diagnostics, updateErr)
			return
		}
		out.Payload = updated
	}

	data.Token = types.StringValue(out.Payload.Token)
	data.Id = types.StringValue(out.Payload.Token)
	data.Filter = types.StringPointerValue(out.Payload.Filter)
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
	settingsObj, settingsDiags := costReportSettingsObjectFromPayload(ctx, out.Payload.Settings)
	resp.Diagnostics.Append(settingsDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Settings = settingsObj

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
	state.Id = types.StringValue(out.Payload.Token)
	state.Filter = types.StringPointerValue(out.Payload.Filter)
	state.Title = types.StringValue(out.Payload.Title)
	state.Groupings = types.StringValue(out.Payload.Groupings)
	state.StartDate = types.StringValue(out.Payload.StartDate)
	state.EndDate = types.StringValue(out.Payload.EndDate)
	state.PreviousPeriodStartDate = types.StringValue(out.Payload.PreviousPeriodStartDate)
	state.PreviousPeriodEndDate = types.StringValue(out.Payload.PreviousPeriodEndDate)
	state.DateInterval = types.StringValue(out.Payload.DateInterval)
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
	settingsObj, settingsDiags := costReportSettingsObjectFromPayload(ctx, out.Payload.Settings)
	resp.Diagnostics.Append(settingsDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Settings = settingsObj

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r CostReportResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Set BOTH id and token from the provided ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("token"), types.StringValue(req.ID))...)

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
	}

	if !data.ChartType.IsUnknown() && !data.ChartType.IsNull() {
		model.ChartType = data.ChartType.ValueStringPointer()
	}

	if !data.DateBin.IsUnknown() && !data.DateBin.IsNull() {
		model.DateBin = data.DateBin.ValueStringPointer()
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

	// Settings are updated via a separate raw call to avoid the omitempty
	// issue in the generated UpdateCostReportSettings struct.
	if !data.Settings.IsNull() && !data.Settings.IsUnknown() {
		var settings CostReportSettingsModel
		resp.Diagnostics.Append(data.Settings.As(ctx, &settings, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
		updated, settingsErr := r.updateCostReportSettingsRaw(data.Token.ValueString(), settings)
		if settingsErr != nil {
			handleError("Update Cost Report Settings", &resp.Diagnostics, settingsErr)
			return
		}
		out.Payload = updated
	}

	data.Title = types.StringValue(out.Payload.Title)
	data.FolderToken = types.StringValue(out.Payload.FolderToken)
	data.Filter = types.StringPointerValue(out.Payload.Filter)
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
	settingsObj, settingsDiags := costReportSettingsObjectFromPayload(ctx, out.Payload.Settings)
	resp.Diagnostics.Append(settingsDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Settings = settingsObj

	data.Id = types.StringValue(out.Payload.Token)
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

func costReportSettingsObjectFromPayload(ctx context.Context, s *modelsv2.CostReportSettings) (types.Object, diag.Diagnostics) {
	if s == nil {
		return types.ObjectNull(costReportSettingsAttrTypes()), nil
	}
	m := CostReportSettingsModel{
		IncludeCredits:     types.BoolPointerValue(s.IncludeCredits),
		IncludeRefunds:     types.BoolPointerValue(s.IncludeRefunds),
		IncludeDiscounts:   types.BoolPointerValue(s.IncludeDiscounts),
		IncludeTax:         types.BoolPointerValue(s.IncludeTax),
		Amortize:           types.BoolPointerValue(s.Amortize),
		Unallocated:        types.BoolPointerValue(s.Unallocated),
		AggregateBy:        types.StringPointerValue(s.AggregateBy),
		ShowPreviousPeriod: types.BoolPointerValue(s.ShowPreviousPeriod),
	}
	return types.ObjectValueFrom(ctx, costReportSettingsAttrTypes(), m)
}

// updateCostReportSettingsRaw uses the go-openapi transport directly to PUT
// cost report settings. This bypasses the generated UpdateCostReportSettings
// struct which uses `bool` + `omitempty`, causing `false` values to be dropped
// from the JSON payload.
func (r *CostReportResource) updateCostReportSettingsRaw(token string, settings CostReportSettingsModel) (*modelsv2.CostReport, error) {
	body := map[string]interface{}{
		"settings": map[string]interface{}{
			"include_credits":      settings.IncludeCredits.ValueBool(),
			"include_refunds":      settings.IncludeRefunds.ValueBool(),
			"include_discounts":    settings.IncludeDiscounts.ValueBool(),
			"include_tax":          settings.IncludeTax.ValueBool(),
			"amortize":             settings.Amortize.ValueBool(),
			"unallocated":          settings.Unallocated.ValueBool(),
			"show_previous_period": settings.ShowPreviousPeriod.ValueBool(),
		},
	}
	if !settings.AggregateBy.IsNull() && !settings.AggregateBy.IsUnknown() {
		body["settings"].(map[string]interface{})["aggregate_by"] = settings.AggregateBy.ValueString()
	}

	op := &goaruntime.ClientOperation{
		ID:                 "updateCostReport",
		Method:             "PUT",
		PathPattern:        "/cost_reports/{cost_report_token}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"https"},
		Params: goaruntime.ClientRequestWriterFunc(func(req goaruntime.ClientRequest, reg strfmt.Registry) error {
			if err := req.SetPathParam("cost_report_token", token); err != nil {
				return err
			}
			return req.SetBodyParam(body)
		}),
		Reader: goaruntime.ClientResponseReaderFunc(func(response goaruntime.ClientResponse, consumer goaruntime.Consumer) (interface{}, error) {
			if response.Code() == 200 {
				result := &modelsv2.CostReport{}
				if err := consumer.Consume(response.Body(), result); err != nil {
					return nil, err
				}
				return result, nil
			}
			return nil, fmt.Errorf("unexpected status code %d updating cost report settings", response.Code())
		}),
		AuthInfo: r.client.Auth,
	}

	result, err := r.client.V2.Transport.Submit(op)
	if err != nil {
		return nil, err
	}

	payload, ok := result.(*modelsv2.CostReport)
	if !ok {
		return nil, fmt.Errorf("unexpected response type updating cost report settings")
	}
	return payload, nil
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
