package vantage

import (
	"context"
	"reflect"
	"testing"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
