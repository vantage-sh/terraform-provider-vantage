package vantage

import (
	"context"
	"reflect"
	"testing"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_budget_alerts"
)

// TestBudgetAlertResourceSchemaMatchesModel guards the handwritten resource
// model. Terraform fails at runtime when the model and the schema disagree on
// which attributes exist.
func TestBudgetAlertResourceSchemaMatchesModel(t *testing.T) {
	attributes := budgetAlertResourceSchema(t).Attributes

	schemaNames := make([]string, 0, len(attributes))
	for name := range attributes {
		schemaNames = append(schemaNames, name)
	}

	assertSameNames(t, "resource schema", schemaNames, tfsdkFieldNames(t, budgetAlertModel{}))
}

// TestBudgetAlertDataSourceSchemaMatchesModel guards the handwritten data source
// model against the generated nested object.
func TestBudgetAlertDataSourceSchemaMatchesModel(t *testing.T) {
	ctx := context.Background()

	alerts, ok := budgetAlertsDataSourceSchemaAttribute(ctx).(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatalf("budget_alerts is %T, want a list of nested objects", budgetAlertsDataSourceSchemaAttribute(ctx))
	}

	attributes := alerts.NestedObject.Attributes
	schemaNames := make([]string, 0, len(attributes))
	for name := range attributes {
		schemaNames = append(schemaNames, name)
	}

	assertSameNames(t, "data source schema", schemaNames, tfsdkFieldNames(t, budgetAlertDataSourceModel{}))
}

// TestBudgetAlertDurationInDaysIsOptionalInteger pins the duration_in_days
// design. The API types the field as a string on write and as an integer on
// read. The resource declares one optional integer, and omitting it tracks the
// full month, so nobody has to pass an empty string for the common case.
func TestBudgetAlertDurationInDaysIsOptionalInteger(t *testing.T) {
	attribute, ok := budgetAlertResourceSchema(t).Attributes["duration_in_days"]
	if !ok {
		t.Fatal("duration_in_days is missing from the resource schema")
	}

	if _, ok := attribute.(schema.Int64Attribute); !ok {
		t.Fatalf("duration_in_days is %T, want schema.Int64Attribute", attribute)
	}
	if attribute.IsRequired() {
		t.Fatal("duration_in_days is required, want optional so that omitting it tracks the full month")
	}
	if !attribute.IsOptional() {
		t.Fatal("duration_in_days is not optional")
	}
}

func budgetAlertResourceSchema(t *testing.T) schema.Schema {
	t.Helper()

	resp := &fwresource.SchemaResponse{}
	NewBudgetAlertResource().Schema(context.Background(), fwresource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics building the resource schema: %v", resp.Diagnostics)
	}
	return resp.Schema
}

func budgetAlertsDataSourceSchemaAttribute(ctx context.Context) dsschema.Attribute {
	return datasource_budget_alerts.BudgetAlertsDataSourceSchema(ctx).Attributes["budget_alerts"]
}

// tfsdkFieldNames returns the tfsdk tag of every field on a model struct.
func tfsdkFieldNames(t *testing.T, model any) []string {
	t.Helper()

	modelType := reflect.TypeOf(model)
	names := make([]string, 0, modelType.NumField())
	for i := 0; i < modelType.NumField(); i++ {
		tag, ok := modelType.Field(i).Tag.Lookup("tfsdk")
		if !ok {
			t.Fatalf("field %s has no tfsdk tag", modelType.Field(i).Name)
		}
		names = append(names, tag)
	}
	return names
}

func assertSameNames(t *testing.T, subject string, schemaNames, modelNames []string) {
	t.Helper()

	inModel := make(map[string]bool, len(modelNames))
	for _, name := range modelNames {
		inModel[name] = true
	}
	inSchema := make(map[string]bool, len(schemaNames))
	for _, name := range schemaNames {
		inSchema[name] = true
	}

	for _, name := range schemaNames {
		if !inModel[name] {
			t.Errorf("%s has attribute %q, which the model is missing", subject, name)
		}
	}
	for _, name := range modelNames {
		if !inSchema[name] {
			t.Errorf("the model has field %q, which %s is missing", name, subject)
		}
	}
}

// TestBudgetAlertDurationInDaysRange checks the validator that keeps zero out of
// the configuration. The API reports a full month as an absent duration, which
// reads back as null, so a configured zero would never match the applied value.
func TestBudgetAlertDurationInDaysRange(t *testing.T) {
	ctx := context.Background()

	attribute, ok := budgetAlertResourceSchema(t).Attributes["duration_in_days"].(schema.Int64Attribute)
	if !ok {
		t.Fatal("duration_in_days is not an integer attribute")
	}
	if len(attribute.Validators) == 0 {
		t.Fatal("duration_in_days has no validator")
	}

	tests := []struct {
		value     int64
		wantError bool
	}{
		{value: 0, wantError: true},
		{value: -1, wantError: true},
		{value: 32, wantError: true},
		{value: 1, wantError: false},
		{value: 7, wantError: false},
		{value: 31, wantError: false},
	}

	for _, tt := range tests {
		resp := &validator.Int64Response{}
		for _, v := range attribute.Validators {
			v.ValidateInt64(ctx, validator.Int64Request{
				Path:        path.Root("duration_in_days"),
				ConfigValue: types.Int64Value(tt.value),
			}, resp)
		}

		if got := resp.Diagnostics.HasError(); got != tt.wantError {
			t.Errorf("duration_in_days = %d: error = %v, want %v", tt.value, got, tt.wantError)
		}
	}
}

// TestBudgetAlertUserTokensCleared covers the transition that replaces the alert.
// The API answers "user_tokens is empty" to an empty array and to a null alike,
// so an update can never take the last user off an existing alert.
func TestBudgetAlertUserTokensCleared(t *testing.T) {
	set := func(values ...string) types.Set {
		return testBudgetAlertSet(t, values...)
	}

	tests := []struct {
		name        string
		configValue types.Set
		stateValue  types.Set
		want        bool
	}{
		{"emptying a set that held users replaces", set(), set("usr_1"), true},
		{"a set left out of the configuration keeps the users", types.SetNull(types.StringType), set("usr_1"), false},
		{"an unknown set decides nothing", types.SetUnknown(types.StringType), set("usr_1"), false},
		{"an already empty set is no change", set(), set(), false},
		{"changing the users is an update", set("usr_2"), set("usr_1"), false},
		{"adding users is an update", set("usr_1", "usr_2"), set("usr_1"), false},
		{"no prior state is a create", set(), types.SetNull(types.StringType), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := budgetAlertUserTokensCleared(tt.configValue, tt.stateValue); got != tt.want {
				t.Fatalf("budgetAlertUserTokensCleared() = %v, want %v", got, tt.want)
			}
		})
	}
}
