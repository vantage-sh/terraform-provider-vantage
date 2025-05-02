package vantage

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"strings"
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
	Settings                SettingsValue `tfsdk:"settings"`
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
				Attributes: map[string]schema.Attribute{
					"aggregate_by": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "Report will aggregate by cost or usage.",
						MarkdownDescription: "Report will aggregate by cost or usage.",
						Default:             stringdefault.StaticString("cost"),
					},
					"amortize": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "Report will amortize.",
						MarkdownDescription: "Report will amortize.",
						Default:             booldefault.StaticBool(true),
					},
					"include_credits": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "Report will include credits.",
						MarkdownDescription: "Report will include credits.",
						Default:             booldefault.StaticBool(false),
					},
					"include_discounts": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "Report will include discounts.",
						MarkdownDescription: "Report will include discounts.",
						Default:             booldefault.StaticBool(true),
					},
					"include_refunds": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "Report will include refunds.",
						MarkdownDescription: "Report will include refunds.",
						Default:             booldefault.StaticBool(false),
					},
					"include_tax": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "Report will include tax.",
						MarkdownDescription: "Report will include tax.",
						Default:             booldefault.StaticBool(true),
					},
					"unallocated": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "Report will show unallocated costs.",
						MarkdownDescription: "Report will show unallocated costs.",
						Default:             booldefault.StaticBool(false),
					},
				},
				CustomType: SettingsType{
					ObjectType: types.ObjectType{
						AttrTypes: SettingsValue{}.AttributeTypes(ctx),
					},
				},
				Optional:            true,
				Computed:            true,
				Description:         "Report settings.",
				MarkdownDescription: "Report settings.",
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
	var data *costReportModel 
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := data.toCreate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := costsv2.NewCreateCostReportParams().WithCreateCostReport(body)
	out, err := r.client.V2.Costs.CreateCostReport(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*costsv2.CreateCostReportBadRequest); ok {
			handleBadRequest("Create Cost Report Resource", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Create Cost Report Resource", &resp.Diagnostics, err)
		return
	}

	if diag := data.applyPayload(ctx, out.Payload); diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r CostReportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *costReportModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := costsv2.NewGetCostReportParams().WithCostReportToken(state.Token.ValueString())
	out, err := r.client.V2.Costs.GetCostReport(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*costsv2.GetCostReportNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Get Cost Report Resource", &resp.Diagnostics, err)
		return
	}

	diag := state.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r CostReportResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r CostReportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *costReportModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := data.toUpdate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	costsv2.NewUpdateCostReportParams()
	params := costsv2.NewUpdateCostReportParams().
		WithCostReportToken(data.Token.ValueString()).
		WithUpdateCostReport(body)

	out, err := r.client.V2.Costs.UpdateCostReport(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*costsv2.UpdateCostReportBadRequest); ok {
			handleBadRequest("Update Cost Report Resource", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Create Cost Report Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r CostReportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *costReportModel
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

var _ basetypes.ObjectTypable = SettingsType{}

type SettingsType struct {
	basetypes.ObjectType
}

func (t SettingsType) Equal(o attr.Type) bool {
	other, ok := o.(SettingsType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t SettingsType) String() string {
	return "SettingsType"
}

func (t SettingsType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	aggregateByAttribute, ok := attributes["aggregate_by"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`aggregate_by is missing from object`)

		return nil, diags
	}

	aggregateByVal, ok := aggregateByAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`aggregate_by expected to be basetypes.StringValue, was: %T`, aggregateByAttribute))
	}

	amortizeAttribute, ok := attributes["amortize"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`amortize is missing from object`)

		return nil, diags
	}

	amortizeVal, ok := amortizeAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`amortize expected to be basetypes.BoolValue, was: %T`, amortizeAttribute))
	}

	includeCreditsAttribute, ok := attributes["include_credits"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`include_credits is missing from object`)

		return nil, diags
	}

	includeCreditsVal, ok := includeCreditsAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`include_credits expected to be basetypes.BoolValue, was: %T`, includeCreditsAttribute))
	}

	includeDiscountsAttribute, ok := attributes["include_discounts"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`include_discounts is missing from object`)

		return nil, diags
	}

	includeDiscountsVal, ok := includeDiscountsAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`include_discounts expected to be basetypes.BoolValue, was: %T`, includeDiscountsAttribute))
	}

	includeRefundsAttribute, ok := attributes["include_refunds"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`include_refunds is missing from object`)

		return nil, diags
	}

	includeRefundsVal, ok := includeRefundsAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`include_refunds expected to be basetypes.BoolValue, was: %T`, includeRefundsAttribute))
	}

	includeTaxAttribute, ok := attributes["include_tax"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`include_tax is missing from object`)

		return nil, diags
	}

	includeTaxVal, ok := includeTaxAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`include_tax expected to be basetypes.BoolValue, was: %T`, includeTaxAttribute))
	}

	unallocatedAttribute, ok := attributes["unallocated"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`unallocated is missing from object`)

		return nil, diags
	}

	unallocatedVal, ok := unallocatedAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`unallocated expected to be basetypes.BoolValue, was: %T`, unallocatedAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return SettingsValue{
		AggregateBy:      aggregateByVal,
		Amortize:         amortizeVal,
		IncludeCredits:   includeCreditsVal,
		IncludeDiscounts: includeDiscountsVal,
		IncludeRefunds:   includeRefundsVal,
		IncludeTax:       includeTaxVal,
		Unallocated:      unallocatedVal,
		state:            attr.ValueStateKnown,
	}, diags
}

func NewSettingsValueNull() SettingsValue {
	return SettingsValue{
		state: attr.ValueStateNull,
	}
}

func NewSettingsValueUnknown() SettingsValue {
	return SettingsValue{
		state: attr.ValueStateUnknown,
	}
}

func NewSettingsValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (SettingsValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing SettingsValue Attribute Value",
				"While creating a SettingsValue value, a missing attribute value was detected. "+
					"A SettingsValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("SettingsValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid SettingsValue Attribute Type",
				"While creating a SettingsValue value, an invalid attribute value was detected. "+
					"A SettingsValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("SettingsValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("SettingsValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra SettingsValue Attribute Value",
				"While creating a SettingsValue value, an extra attribute value was detected. "+
					"A SettingsValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra SettingsValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewSettingsValueUnknown(), diags
	}

	aggregateByAttribute, ok := attributes["aggregate_by"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`aggregate_by is missing from object`)

		return NewSettingsValueUnknown(), diags
	}

	aggregateByVal, ok := aggregateByAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`aggregate_by expected to be basetypes.StringValue, was: %T`, aggregateByAttribute))
	}

	amortizeAttribute, ok := attributes["amortize"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`amortize is missing from object`)

		return NewSettingsValueUnknown(), diags
	}

	amortizeVal, ok := amortizeAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`amortize expected to be basetypes.BoolValue, was: %T`, amortizeAttribute))
	}

	includeCreditsAttribute, ok := attributes["include_credits"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`include_credits is missing from object`)

		return NewSettingsValueUnknown(), diags
	}

	includeCreditsVal, ok := includeCreditsAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`include_credits expected to be basetypes.BoolValue, was: %T`, includeCreditsAttribute))
	}

	includeDiscountsAttribute, ok := attributes["include_discounts"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`include_discounts is missing from object`)

		return NewSettingsValueUnknown(), diags
	}

	includeDiscountsVal, ok := includeDiscountsAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`include_discounts expected to be basetypes.BoolValue, was: %T`, includeDiscountsAttribute))
	}

	includeRefundsAttribute, ok := attributes["include_refunds"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`include_refunds is missing from object`)

		return NewSettingsValueUnknown(), diags
	}

	includeRefundsVal, ok := includeRefundsAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`include_refunds expected to be basetypes.BoolValue, was: %T`, includeRefundsAttribute))
	}

	includeTaxAttribute, ok := attributes["include_tax"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`include_tax is missing from object`)

		return NewSettingsValueUnknown(), diags
	}

	includeTaxVal, ok := includeTaxAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`include_tax expected to be basetypes.BoolValue, was: %T`, includeTaxAttribute))
	}

	unallocatedAttribute, ok := attributes["unallocated"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`unallocated is missing from object`)

		return NewSettingsValueUnknown(), diags
	}

	unallocatedVal, ok := unallocatedAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`unallocated expected to be basetypes.BoolValue, was: %T`, unallocatedAttribute))
	}

	if diags.HasError() {
		return NewSettingsValueUnknown(), diags
	}

	return SettingsValue{
		AggregateBy:      aggregateByVal,
		Amortize:         amortizeVal,
		IncludeCredits:   includeCreditsVal,
		IncludeDiscounts: includeDiscountsVal,
		IncludeRefunds:   includeRefundsVal,
		IncludeTax:       includeTaxVal,
		Unallocated:      unallocatedVal,
		state:            attr.ValueStateKnown,
	}, diags
}

func NewSettingsValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) SettingsValue {
	object, diags := NewSettingsValue(attributeTypes, attributes)

	if diags.HasError() {
		// This could potentially be added to the diag package.
		diagsStrings := make([]string, 0, len(diags))

		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}

		panic("NewSettingsValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t SettingsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewSettingsValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewSettingsValueUnknown(), nil
	}

	if in.IsNull() {
		return NewSettingsValueNull(), nil
	}

	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}

	err := in.As(&val)

	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)

		if err != nil {
			return nil, err
		}

		attributes[k] = a
	}

	return NewSettingsValueMust(SettingsValue{}.AttributeTypes(ctx), attributes), nil
}

func (t SettingsType) ValueType(ctx context.Context) attr.Value {
	return SettingsValue{}
}

var _ basetypes.ObjectValuable = SettingsValue{}

type SettingsValue struct {
	AggregateBy      basetypes.StringValue `tfsdk:"aggregate_by"`
	Amortize         basetypes.BoolValue   `tfsdk:"amortize"`
	IncludeCredits   basetypes.BoolValue   `tfsdk:"include_credits"`
	IncludeDiscounts basetypes.BoolValue   `tfsdk:"include_discounts"`
	IncludeRefunds   basetypes.BoolValue   `tfsdk:"include_refunds"`
	IncludeTax       basetypes.BoolValue   `tfsdk:"include_tax"`
	Unallocated      basetypes.BoolValue   `tfsdk:"unallocated"`
	state            attr.ValueState
}

func (v SettingsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 7)

	var val tftypes.Value
	var err error

	attrTypes["aggregate_by"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["amortize"] = basetypes.BoolType{}.TerraformType(ctx)
	attrTypes["include_credits"] = basetypes.BoolType{}.TerraformType(ctx)
	attrTypes["include_discounts"] = basetypes.BoolType{}.TerraformType(ctx)
	attrTypes["include_refunds"] = basetypes.BoolType{}.TerraformType(ctx)
	attrTypes["include_tax"] = basetypes.BoolType{}.TerraformType(ctx)
	attrTypes["unallocated"] = basetypes.BoolType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 7)

		val, err = v.AggregateBy.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["aggregate_by"] = val

		val, err = v.Amortize.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["amortize"] = val

		val, err = v.IncludeCredits.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["include_credits"] = val

		val, err = v.IncludeDiscounts.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["include_discounts"] = val

		val, err = v.IncludeRefunds.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["include_refunds"] = val

		val, err = v.IncludeTax.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["include_tax"] = val

		val, err = v.Unallocated.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["unallocated"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v SettingsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v SettingsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v SettingsValue) String() string {
	return "SettingsValue"
}

func (v SettingsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributeTypes := map[string]attr.Type{
		"aggregate_by":      basetypes.StringType{},
		"amortize":          basetypes.BoolType{},
		"include_credits":   basetypes.BoolType{},
		"include_discounts": basetypes.BoolType{},
		"include_refunds":   basetypes.BoolType{},
		"include_tax":       basetypes.BoolType{},
		"unallocated":       basetypes.BoolType{},
	}

	if v.IsNull() {
		return types.ObjectNull(attributeTypes), diags
	}

	if v.IsUnknown() {
		return types.ObjectUnknown(attributeTypes), diags
	}

	objVal, diags := types.ObjectValue(
		attributeTypes,
		map[string]attr.Value{
			"aggregate_by":      v.AggregateBy,
			"amortize":          v.Amortize,
			"include_credits":   v.IncludeCredits,
			"include_discounts": v.IncludeDiscounts,
			"include_refunds":   v.IncludeRefunds,
			"include_tax":       v.IncludeTax,
			"unallocated":       v.Unallocated,
		})

	return objVal, diags
}

func (v SettingsValue) Equal(o attr.Value) bool {
	other, ok := o.(SettingsValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.AggregateBy.Equal(other.AggregateBy) {
		return false
	}

	if !v.Amortize.Equal(other.Amortize) {
		return false
	}

	if !v.IncludeCredits.Equal(other.IncludeCredits) {
		return false
	}

	if !v.IncludeDiscounts.Equal(other.IncludeDiscounts) {
		return false
	}

	if !v.IncludeRefunds.Equal(other.IncludeRefunds) {
		return false
	}

	if !v.IncludeTax.Equal(other.IncludeTax) {
		return false
	}

	if !v.Unallocated.Equal(other.Unallocated) {
		return false
	}

	return true
}

func (v SettingsValue) Type(ctx context.Context) attr.Type {
	return SettingsType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v SettingsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"aggregate_by":      basetypes.StringType{},
		"amortize":          basetypes.BoolType{},
		"include_credits":   basetypes.BoolType{},
		"include_discounts": basetypes.BoolType{},
		"include_refunds":   basetypes.BoolType{},
		"include_tax":       basetypes.BoolType{},
		"unallocated":       basetypes.BoolType{},
	}
}

