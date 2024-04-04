// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package datasource_financial_commitment_reports

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func FinancialCommitmentReportsDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"financial_commitment_reports": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"created_at": schema.StringAttribute{
							Computed:            true,
							Description:         "The date and time, in UTC, the report was created. ISO 8601 Formatted.",
							MarkdownDescription: "The date and time, in UTC, the report was created. ISO 8601 Formatted.",
						},
						"date_bucket": schema.StringAttribute{
							Computed:            true,
							Description:         "How costs are grouped and displayed in the FinancialCommitmentReport. Possible values: day, week, month.",
							MarkdownDescription: "How costs are grouped and displayed in the FinancialCommitmentReport. Possible values: day, week, month.",
						},
						"date_interval": schema.StringAttribute{
							Computed:            true,
							Description:         "The date range for the FinancialCommitmentReport. Only present if a custom date range is not specified.",
							MarkdownDescription: "The date range for the FinancialCommitmentReport. Only present if a custom date range is not specified.",
						},
						"default": schema.BoolAttribute{
							Computed:            true,
							Description:         "Indicates whether the FinancialCommitmentReport is the default report.",
							MarkdownDescription: "Indicates whether the FinancialCommitmentReport is the default report.",
						},
						"end_date": schema.StringAttribute{
							Computed:            true,
							Description:         "The end date for the FinancialCommitmentReport. Only set for custom date ranges. ISO 8601 Formatted.",
							MarkdownDescription: "The end date for the FinancialCommitmentReport. Only set for custom date ranges. ISO 8601 Formatted.",
						},
						"groupings": schema.StringAttribute{
							Computed:            true,
							Description:         "The grouping aggregations applied to the filtered data.",
							MarkdownDescription: "The grouping aggregations applied to the filtered data.",
						},
						"on_demand_costs_scope": schema.StringAttribute{
							Computed:            true,
							Description:         "The scope for the costs. Possible values: discountable, all.",
							MarkdownDescription: "The scope for the costs. Possible values: discountable, all.",
						},
						"start_date": schema.StringAttribute{
							Computed:            true,
							Description:         "The start date for the FinancialCommitmentReport. Only set for custom date ranges. ISO 8601 Formatted.",
							MarkdownDescription: "The start date for the FinancialCommitmentReport. Only set for custom date ranges. ISO 8601 Formatted.",
						},
						"title": schema.StringAttribute{
							Computed:            true,
							Description:         "The title of the FinancialCommitmentReport.",
							MarkdownDescription: "The title of the FinancialCommitmentReport.",
						},
						"token": schema.StringAttribute{
							Computed: true,
						},
						"user_token": schema.StringAttribute{
							Computed:            true,
							Description:         "The token for the User who created this FinancialCommitmentReport.",
							MarkdownDescription: "The token for the User who created this FinancialCommitmentReport.",
						},
						"workspace_token": schema.StringAttribute{
							Computed:            true,
							Description:         "The token for the Workspace the FinancialCommitmentReport is a part of.",
							MarkdownDescription: "The token for the Workspace the FinancialCommitmentReport is a part of.",
						},
					},
					CustomType: FinancialCommitmentReportsType{
						ObjectType: types.ObjectType{
							AttrTypes: FinancialCommitmentReportsValue{}.AttributeTypes(ctx),
						},
					},
				},
				Computed: true,
			},
		},
	}
}

type FinancialCommitmentReportsModel struct {
	FinancialCommitmentReports types.List `tfsdk:"financial_commitment_reports"`
}

var _ basetypes.ObjectTypable = FinancialCommitmentReportsType{}

type FinancialCommitmentReportsType struct {
	basetypes.ObjectType
}

func (t FinancialCommitmentReportsType) Equal(o attr.Type) bool {
	other, ok := o.(FinancialCommitmentReportsType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t FinancialCommitmentReportsType) String() string {
	return "FinancialCommitmentReportsType"
}

func (t FinancialCommitmentReportsType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	createdAtAttribute, ok := attributes["created_at"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`created_at is missing from object`)

		return nil, diags
	}

	createdAtVal, ok := createdAtAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`created_at expected to be basetypes.StringValue, was: %T`, createdAtAttribute))
	}

	dateBucketAttribute, ok := attributes["date_bucket"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`date_bucket is missing from object`)

		return nil, diags
	}

	dateBucketVal, ok := dateBucketAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`date_bucket expected to be basetypes.StringValue, was: %T`, dateBucketAttribute))
	}

	dateIntervalAttribute, ok := attributes["date_interval"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`date_interval is missing from object`)

		return nil, diags
	}

	dateIntervalVal, ok := dateIntervalAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`date_interval expected to be basetypes.StringValue, was: %T`, dateIntervalAttribute))
	}

	defaultAttribute, ok := attributes["default"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`default is missing from object`)

		return nil, diags
	}

	defaultVal, ok := defaultAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`default expected to be basetypes.BoolValue, was: %T`, defaultAttribute))
	}

	endDateAttribute, ok := attributes["end_date"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`end_date is missing from object`)

		return nil, diags
	}

	endDateVal, ok := endDateAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`end_date expected to be basetypes.StringValue, was: %T`, endDateAttribute))
	}

	groupingsAttribute, ok := attributes["groupings"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`groupings is missing from object`)

		return nil, diags
	}

	groupingsVal, ok := groupingsAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`groupings expected to be basetypes.StringValue, was: %T`, groupingsAttribute))
	}

	onDemandCostsScopeAttribute, ok := attributes["on_demand_costs_scope"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`on_demand_costs_scope is missing from object`)

		return nil, diags
	}

	onDemandCostsScopeVal, ok := onDemandCostsScopeAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`on_demand_costs_scope expected to be basetypes.StringValue, was: %T`, onDemandCostsScopeAttribute))
	}

	startDateAttribute, ok := attributes["start_date"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`start_date is missing from object`)

		return nil, diags
	}

	startDateVal, ok := startDateAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`start_date expected to be basetypes.StringValue, was: %T`, startDateAttribute))
	}

	titleAttribute, ok := attributes["title"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`title is missing from object`)

		return nil, diags
	}

	titleVal, ok := titleAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`title expected to be basetypes.StringValue, was: %T`, titleAttribute))
	}

	tokenAttribute, ok := attributes["token"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`token is missing from object`)

		return nil, diags
	}

	tokenVal, ok := tokenAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`token expected to be basetypes.StringValue, was: %T`, tokenAttribute))
	}

	userTokenAttribute, ok := attributes["user_token"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`user_token is missing from object`)

		return nil, diags
	}

	userTokenVal, ok := userTokenAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`user_token expected to be basetypes.StringValue, was: %T`, userTokenAttribute))
	}

	workspaceTokenAttribute, ok := attributes["workspace_token"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`workspace_token is missing from object`)

		return nil, diags
	}

	workspaceTokenVal, ok := workspaceTokenAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`workspace_token expected to be basetypes.StringValue, was: %T`, workspaceTokenAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return FinancialCommitmentReportsValue{
		CreatedAt:          createdAtVal,
		DateBucket:         dateBucketVal,
		DateInterval:       dateIntervalVal,
		Default:            defaultVal,
		EndDate:            endDateVal,
		Groupings:          groupingsVal,
		OnDemandCostsScope: onDemandCostsScopeVal,
		StartDate:          startDateVal,
		Title:              titleVal,
		Token:              tokenVal,
		UserToken:          userTokenVal,
		WorkspaceToken:     workspaceTokenVal,
		state:              attr.ValueStateKnown,
	}, diags
}

func NewFinancialCommitmentReportsValueNull() FinancialCommitmentReportsValue {
	return FinancialCommitmentReportsValue{
		state: attr.ValueStateNull,
	}
}

func NewFinancialCommitmentReportsValueUnknown() FinancialCommitmentReportsValue {
	return FinancialCommitmentReportsValue{
		state: attr.ValueStateUnknown,
	}
}

func NewFinancialCommitmentReportsValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (FinancialCommitmentReportsValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing FinancialCommitmentReportsValue Attribute Value",
				"While creating a FinancialCommitmentReportsValue value, a missing attribute value was detected. "+
					"A FinancialCommitmentReportsValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("FinancialCommitmentReportsValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid FinancialCommitmentReportsValue Attribute Type",
				"While creating a FinancialCommitmentReportsValue value, an invalid attribute value was detected. "+
					"A FinancialCommitmentReportsValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("FinancialCommitmentReportsValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("FinancialCommitmentReportsValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra FinancialCommitmentReportsValue Attribute Value",
				"While creating a FinancialCommitmentReportsValue value, an extra attribute value was detected. "+
					"A FinancialCommitmentReportsValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra FinancialCommitmentReportsValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewFinancialCommitmentReportsValueUnknown(), diags
	}

	createdAtAttribute, ok := attributes["created_at"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`created_at is missing from object`)

		return NewFinancialCommitmentReportsValueUnknown(), diags
	}

	createdAtVal, ok := createdAtAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`created_at expected to be basetypes.StringValue, was: %T`, createdAtAttribute))
	}

	dateBucketAttribute, ok := attributes["date_bucket"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`date_bucket is missing from object`)

		return NewFinancialCommitmentReportsValueUnknown(), diags
	}

	dateBucketVal, ok := dateBucketAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`date_bucket expected to be basetypes.StringValue, was: %T`, dateBucketAttribute))
	}

	dateIntervalAttribute, ok := attributes["date_interval"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`date_interval is missing from object`)

		return NewFinancialCommitmentReportsValueUnknown(), diags
	}

	dateIntervalVal, ok := dateIntervalAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`date_interval expected to be basetypes.StringValue, was: %T`, dateIntervalAttribute))
	}

	defaultAttribute, ok := attributes["default"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`default is missing from object`)

		return NewFinancialCommitmentReportsValueUnknown(), diags
	}

	defaultVal, ok := defaultAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`default expected to be basetypes.BoolValue, was: %T`, defaultAttribute))
	}

	endDateAttribute, ok := attributes["end_date"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`end_date is missing from object`)

		return NewFinancialCommitmentReportsValueUnknown(), diags
	}

	endDateVal, ok := endDateAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`end_date expected to be basetypes.StringValue, was: %T`, endDateAttribute))
	}

	groupingsAttribute, ok := attributes["groupings"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`groupings is missing from object`)

		return NewFinancialCommitmentReportsValueUnknown(), diags
	}

	groupingsVal, ok := groupingsAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`groupings expected to be basetypes.StringValue, was: %T`, groupingsAttribute))
	}

	onDemandCostsScopeAttribute, ok := attributes["on_demand_costs_scope"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`on_demand_costs_scope is missing from object`)

		return NewFinancialCommitmentReportsValueUnknown(), diags
	}

	onDemandCostsScopeVal, ok := onDemandCostsScopeAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`on_demand_costs_scope expected to be basetypes.StringValue, was: %T`, onDemandCostsScopeAttribute))
	}

	startDateAttribute, ok := attributes["start_date"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`start_date is missing from object`)

		return NewFinancialCommitmentReportsValueUnknown(), diags
	}

	startDateVal, ok := startDateAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`start_date expected to be basetypes.StringValue, was: %T`, startDateAttribute))
	}

	titleAttribute, ok := attributes["title"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`title is missing from object`)

		return NewFinancialCommitmentReportsValueUnknown(), diags
	}

	titleVal, ok := titleAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`title expected to be basetypes.StringValue, was: %T`, titleAttribute))
	}

	tokenAttribute, ok := attributes["token"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`token is missing from object`)

		return NewFinancialCommitmentReportsValueUnknown(), diags
	}

	tokenVal, ok := tokenAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`token expected to be basetypes.StringValue, was: %T`, tokenAttribute))
	}

	userTokenAttribute, ok := attributes["user_token"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`user_token is missing from object`)

		return NewFinancialCommitmentReportsValueUnknown(), diags
	}

	userTokenVal, ok := userTokenAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`user_token expected to be basetypes.StringValue, was: %T`, userTokenAttribute))
	}

	workspaceTokenAttribute, ok := attributes["workspace_token"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`workspace_token is missing from object`)

		return NewFinancialCommitmentReportsValueUnknown(), diags
	}

	workspaceTokenVal, ok := workspaceTokenAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`workspace_token expected to be basetypes.StringValue, was: %T`, workspaceTokenAttribute))
	}

	if diags.HasError() {
		return NewFinancialCommitmentReportsValueUnknown(), diags
	}

	return FinancialCommitmentReportsValue{
		CreatedAt:          createdAtVal,
		DateBucket:         dateBucketVal,
		DateInterval:       dateIntervalVal,
		Default:            defaultVal,
		EndDate:            endDateVal,
		Groupings:          groupingsVal,
		OnDemandCostsScope: onDemandCostsScopeVal,
		StartDate:          startDateVal,
		Title:              titleVal,
		Token:              tokenVal,
		UserToken:          userTokenVal,
		WorkspaceToken:     workspaceTokenVal,
		state:              attr.ValueStateKnown,
	}, diags
}

func NewFinancialCommitmentReportsValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) FinancialCommitmentReportsValue {
	object, diags := NewFinancialCommitmentReportsValue(attributeTypes, attributes)

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

		panic("NewFinancialCommitmentReportsValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t FinancialCommitmentReportsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewFinancialCommitmentReportsValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewFinancialCommitmentReportsValueUnknown(), nil
	}

	if in.IsNull() {
		return NewFinancialCommitmentReportsValueNull(), nil
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

	return NewFinancialCommitmentReportsValueMust(FinancialCommitmentReportsValue{}.AttributeTypes(ctx), attributes), nil
}

func (t FinancialCommitmentReportsType) ValueType(ctx context.Context) attr.Value {
	return FinancialCommitmentReportsValue{}
}

var _ basetypes.ObjectValuable = FinancialCommitmentReportsValue{}

type FinancialCommitmentReportsValue struct {
	CreatedAt          basetypes.StringValue `tfsdk:"created_at"`
	DateBucket         basetypes.StringValue `tfsdk:"date_bucket"`
	DateInterval       basetypes.StringValue `tfsdk:"date_interval"`
	Default            basetypes.BoolValue   `tfsdk:"default"`
	EndDate            basetypes.StringValue `tfsdk:"end_date"`
	Groupings          basetypes.StringValue `tfsdk:"groupings"`
	OnDemandCostsScope basetypes.StringValue `tfsdk:"on_demand_costs_scope"`
	StartDate          basetypes.StringValue `tfsdk:"start_date"`
	Title              basetypes.StringValue `tfsdk:"title"`
	Token              basetypes.StringValue `tfsdk:"token"`
	UserToken          basetypes.StringValue `tfsdk:"user_token"`
	WorkspaceToken     basetypes.StringValue `tfsdk:"workspace_token"`
	state              attr.ValueState
}

func (v FinancialCommitmentReportsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 12)

	var val tftypes.Value
	var err error

	attrTypes["created_at"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["date_bucket"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["date_interval"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["default"] = basetypes.BoolType{}.TerraformType(ctx)
	attrTypes["end_date"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["groupings"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["on_demand_costs_scope"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["start_date"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["title"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["token"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["user_token"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["workspace_token"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 12)

		val, err = v.CreatedAt.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["created_at"] = val

		val, err = v.DateBucket.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["date_bucket"] = val

		val, err = v.DateInterval.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["date_interval"] = val

		val, err = v.Default.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["default"] = val

		val, err = v.EndDate.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["end_date"] = val

		val, err = v.Groupings.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["groupings"] = val

		val, err = v.OnDemandCostsScope.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["on_demand_costs_scope"] = val

		val, err = v.StartDate.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["start_date"] = val

		val, err = v.Title.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["title"] = val

		val, err = v.Token.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["token"] = val

		val, err = v.UserToken.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["user_token"] = val

		val, err = v.WorkspaceToken.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["workspace_token"] = val

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

func (v FinancialCommitmentReportsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v FinancialCommitmentReportsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v FinancialCommitmentReportsValue) String() string {
	return "FinancialCommitmentReportsValue"
}

func (v FinancialCommitmentReportsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	objVal, diags := types.ObjectValue(
		map[string]attr.Type{
			"created_at":            basetypes.StringType{},
			"date_bucket":           basetypes.StringType{},
			"date_interval":         basetypes.StringType{},
			"default":               basetypes.BoolType{},
			"end_date":              basetypes.StringType{},
			"groupings":             basetypes.StringType{},
			"on_demand_costs_scope": basetypes.StringType{},
			"start_date":            basetypes.StringType{},
			"title":                 basetypes.StringType{},
			"token":                 basetypes.StringType{},
			"user_token":            basetypes.StringType{},
			"workspace_token":       basetypes.StringType{},
		},
		map[string]attr.Value{
			"created_at":            v.CreatedAt,
			"date_bucket":           v.DateBucket,
			"date_interval":         v.DateInterval,
			"default":               v.Default,
			"end_date":              v.EndDate,
			"groupings":             v.Groupings,
			"on_demand_costs_scope": v.OnDemandCostsScope,
			"start_date":            v.StartDate,
			"title":                 v.Title,
			"token":                 v.Token,
			"user_token":            v.UserToken,
			"workspace_token":       v.WorkspaceToken,
		})

	return objVal, diags
}

func (v FinancialCommitmentReportsValue) Equal(o attr.Value) bool {
	other, ok := o.(FinancialCommitmentReportsValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.CreatedAt.Equal(other.CreatedAt) {
		return false
	}

	if !v.DateBucket.Equal(other.DateBucket) {
		return false
	}

	if !v.DateInterval.Equal(other.DateInterval) {
		return false
	}

	if !v.Default.Equal(other.Default) {
		return false
	}

	if !v.EndDate.Equal(other.EndDate) {
		return false
	}

	if !v.Groupings.Equal(other.Groupings) {
		return false
	}

	if !v.OnDemandCostsScope.Equal(other.OnDemandCostsScope) {
		return false
	}

	if !v.StartDate.Equal(other.StartDate) {
		return false
	}

	if !v.Title.Equal(other.Title) {
		return false
	}

	if !v.Token.Equal(other.Token) {
		return false
	}

	if !v.UserToken.Equal(other.UserToken) {
		return false
	}

	if !v.WorkspaceToken.Equal(other.WorkspaceToken) {
		return false
	}

	return true
}

func (v FinancialCommitmentReportsValue) Type(ctx context.Context) attr.Type {
	return FinancialCommitmentReportsType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v FinancialCommitmentReportsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"created_at":            basetypes.StringType{},
		"date_bucket":           basetypes.StringType{},
		"date_interval":         basetypes.StringType{},
		"default":               basetypes.BoolType{},
		"end_date":              basetypes.StringType{},
		"groupings":             basetypes.StringType{},
		"on_demand_costs_scope": basetypes.StringType{},
		"start_date":            basetypes.StringType{},
		"title":                 basetypes.StringType{},
		"token":                 basetypes.StringType{},
		"user_token":            basetypes.StringType{},
		"workspace_token":       basetypes.StringType{},
	}
}
