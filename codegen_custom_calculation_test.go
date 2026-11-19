package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

var usageUnitNames = []string{"usage_unit", "usage_unit_numerator", "usage_unit_denominator"}
var attachmentSchemas = map[string]string{
	"CustomCalculation":       "cost_report_tokens",
	"createCustomCalculation": "cost_report_tokens",
	"updateCustomCalculation": "cost_report_tokens",
	"CostReport":              "custom_calculation_tokens",
	"createCostReport":        "custom_calculation_tokens",
	"updateCostReport":        "custom_calculation_tokens",
}

func TestCustomCalculationCodegenNormalization(t *testing.T) {
	for _, variant := range []string{"hosted", "oneOf", "array"} {
		t.Run(variant, func(t *testing.T) {
			fixture := codegenFixture(t)
			properties := codegenObject(t, fixture, "components", "schemas", "CustomCalculationFilter", "properties")
			if variant == "oneOf" {
				for _, name := range usageUnitNames {
					property := codegenObject(t, properties, name)
					property["oneOf"] = property["x-oneOf"]
					delete(property, "x-oneOf")
					delete(property, "type")
				}
			}
			input := codegenJSON(t, fixture)
			if variant == "array" {
				input = normalizeCalculations(t, input)
			}
			output := normalizeCalculations(t, input)
			if twice := normalizeCalculations(t, output); !bytes.Equal(output, twice) {
				t.Fatal("normalization is not idempotent")
			}
			var normalized map[string]any
			if err := json.Unmarshal(output, &normalized); err != nil {
				t.Fatal(err)
			}
			normalizedProperties := codegenObject(t, normalized, "components", "schemas", "CustomCalculationFilter", "properties")
			if len(normalizedProperties) != 7 {
				t.Fatalf("filter has %d fields, want seven", len(normalizedProperties))
			}
			for _, name := range usageUnitNames {
				property := codegenObject(t, normalizedProperties, name)
				if property["type"] != "array" || property["nullable"] != true {
					t.Fatalf("%s must be a nullable array: %#v", name, property)
				}
				items := codegenObject(t, property, "items")
				if items["type"] != "string" || items["pattern"] != `\S` {
					t.Fatalf("%s lost its string constraints: %#v", name, items)
				}
				for _, removed := range []string{"oneOf", "x-oneOf", "properties"} {
					if _, exists := property[removed]; exists {
						t.Fatalf("%s retained %s", name, removed)
					}
				}
				if property["description"] != codegenObject(t, properties, name)["description"] {
					t.Fatalf("%s lost its outer description", name)
				}
			}
			unit := codegenObject(t, normalizedProperties, "usage_unit")
			if unit["minItems"] != float64(1) || unit["maxItems"] != float64(5) || codegenObject(t, unit, "items")["minLength"] != float64(1) {
				t.Fatalf("array constraints changed: %#v", unit)
			}
			if !reflect.DeepEqual(unit["example"], []any{"GB", "GiB"}) {
				t.Fatalf("two-unit example changed: %#v", unit["example"])
			}
			for _, name := range []string{"type", "filter", "filter_numerator", "filter_denominator"} {
				if !reflect.DeepEqual(normalizedProperties[name], properties[name]) {
					t.Fatalf("unrelated filter member %s changed", name)
				}
			}
			for schema, field := range attachmentSchemas {
				original := codegenObject(t, fixture, "components", "schemas", schema, "properties", field)
				original["format"] = "set"
				if got := codegenObject(t, normalized, "components", "schemas", schema, "properties", field); !reflect.DeepEqual(original, got) {
					t.Fatalf("%s.%s lost metadata: %#v", schema, field, got)
				}
			}
			if !reflect.DeepEqual(codegenObject(t, fixture, "components", "schemas", "Unrelated"), codegenObject(t, normalized, "components", "schemas", "Unrelated")) {
				t.Fatal("unrelated schemas changed")
			}
		})
	}
}

func TestCustomCalculationCodegenMissingTargets(t *testing.T) {
	for _, fixture := range []map[string]any{
		{"openapi": "3.0.0", "paths": map[string]any{}},
		{"components": map[string]any{"schemas": map[string]any{"Unrelated": map[string]any{"type": "string"}}}},
		{"components": map[string]any{"schemas": map[string]any{"CustomCalculationFilter": map[string]any{"type": "object"}, "CostReport": map[string]any{"properties": map[string]any{"title": map[string]any{"type": "string"}}}}}},
	} {
		var normalized map[string]any
		if err := json.Unmarshal(normalizeCalculations(t, codegenJSON(t, fixture)), &normalized); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(fixture, normalized) {
			t.Fatalf("missing targets changed input: %#v", normalized)
		}
	}
}

func TestCustomCalculationCodegenPreservesStrongerBounds(t *testing.T) {
	fixture := codegenFixture(t)
	if err := json.Unmarshal(normalizeCalculations(t, codegenJSON(t, fixture)), &fixture); err != nil {
		t.Fatal(err)
	}
	property := codegenObject(t, fixture, "components", "schemas", "CustomCalculationFilter", "properties", "usage_unit")
	property["minItems"] = float64(3)
	property["maxItems"] = float64(4)
	var normalized map[string]any
	if err := json.Unmarshal(normalizeCalculations(t, codegenJSON(t, fixture)), &normalized); err != nil {
		t.Fatal(err)
	}
	property = codegenObject(t, normalized, "components", "schemas", "CustomCalculationFilter", "properties", "usage_unit")
	if property["minItems"] != float64(3) || property["maxItems"] != float64(4) {
		t.Fatalf("stronger bounds were lost: %#v", property)
	}
}

func TestCustomCalculationCodegenRejectsUnexpectedSchemas(t *testing.T) {
	for _, invalid := range []string{"items", "outer_items", "hidden_items", "weakened_pattern", "outer_type", "scalar_minimum", "union", "attachment"} {
		t.Run(invalid, func(t *testing.T) {
			fixture := codegenFixture(t)
			property := codegenObject(t, fixture, "components", "schemas", "CustomCalculationFilter", "properties", "usage_unit")
			switch invalid {
			case "items":
				codegenObject(t, property["x-oneOf"].([]any)[1].(map[string]any), "items")["type"] = "integer"
			case "outer_items":
				property["items"] = map[string]any{"type": "integer"}
			case "hidden_items":
				codegenObject(t, property["x-oneOf"].([]any)[1].(map[string]any), "items")["type"] = "integer"
				property["items"] = map[string]any{"type": "string"}
			case "weakened_pattern":
				property["items"] = map[string]any{"pattern": ".*"}
			case "outer_type":
				property["type"] = "integer"
			case "scalar_minimum":
				property["x-oneOf"].([]any)[1].(map[string]any)["minItems"] = float64(2)
			case "union":
				property["x-oneOf"] = append(property["x-oneOf"].([]any), map[string]any{"type": "object"})
			case "attachment":
				codegenObject(t, fixture, "components", "schemas", "CostReport", "properties", "custom_calculation_tokens")["type"] = "string"
			}
			command := exec.Command(codegenTool(t, "jq"), "-f", "fix-custom-calculation-schema.jq")
			command.Stdin = bytes.NewReader(codegenJSON(t, fixture))
			if output, err := command.CombinedOutput(); err == nil {
				t.Fatalf("unexpected schema was silently accepted: %s", output)
			}
		})
	}
}

func TestCustomCalculationCodegenGeneratedSchemas(t *testing.T) {
	openapi := codegenTool(t, "tfplugingen-openapi")
	framework := codegenTool(t, "tfplugingen-framework")
	directory := t.TempDir()
	input := filepath.Join(directory, "oas3.json")
	config := filepath.Join(directory, "generator.yaml")
	spec := filepath.Join(directory, "spec.json")
	codegenWrite(t, input, normalizeCalculations(t, codegenJSON(t, codegenFixture(t))))
	codegenWrite(t, config, []byte(calculationGeneratorConfig))
	codegenRun(t, "", openapi, "generate", "--config", config, "--output", spec, input)
	generated := codegenRead(t, spec)
	calculation := codegenObject(t, codegenNamed(t, generated["resources"], "custom_calculation"), "schema")
	filter := codegenObject(t, codegenNamed(t, calculation["attributes"], "filter"), "single_nested")
	if len(filter["attributes"].([]any)) != 7 {
		t.Fatal("generated filter must have all seven members")
	}
	for _, name := range usageUnitNames {
		list := codegenObject(t, codegenNamed(t, filter["attributes"], name), "list")
		codegenObject(t, list, "element_type", "string")
	}
	for resource, field := range map[string]string{"custom_calculation": "cost_report_tokens", "cost_report": "custom_calculation_tokens"} {
		attributes := codegenObject(t, codegenNamed(t, generated["resources"], resource), "schema")["attributes"]
		set := codegenObject(t, codegenNamed(t, attributes, field), "set")
		if set["computed_optional_required"] != "computed_optional" {
			t.Fatalf("%s must be configurable", field)
		}
	}
	codegenNamed(t, generated["datasources"], "custom_calculations")
	for _, kind := range []string{"resources", "data-sources"} {
		codegenRun(t, "", framework, "generate", kind, "--input", spec, "--output", directory)
	}
	// Compile the actual generated types and round-trip two units through the generated value helper.
	for _, name := range []string{"go.mod", "go.sum"} {
		contents, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		codegenWrite(t, filepath.Join(directory, name), contents)
	}
	codegenWrite(t, filepath.Join(directory, "generated_test.go"), []byte(calculationGeneratedTest))
	codegenRun(t, directory, codegenTool(t, "go"), "test", "-mod=readonly", "./...")
}

func normalizeCalculations(t *testing.T, input []byte) []byte {
	t.Helper()
	command := exec.Command(codegenTool(t, "jq"), "-f", "fix-custom-calculation-schema.jq")
	command.Stdin = bytes.NewReader(input)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("normalize schema: %v\n%s", err, output)
	}
	return output
}

func codegenTool(t *testing.T, name string) string {
	t.Helper()
	tool, err := exec.LookPath(name)
	if err != nil {
		t.Skipf("%s is required for schema generation tests", name)
	}
	return tool
}

func codegenFixture(t *testing.T) map[string]any {
	t.Helper()
	return codegenRead(t, filepath.Join("testdata", "custom_calculation_codegen.json"))
}

func codegenRead(t *testing.T, path string) map[string]any {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err := json.Unmarshal(contents, &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func codegenJSON(t *testing.T, value any) []byte {
	t.Helper()
	contents, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return contents
}

func codegenObject(t *testing.T, value map[string]any, path ...string) map[string]any {
	t.Helper()
	for _, name := range path {
		nested, ok := value[name].(map[string]any)
		if !ok {
			t.Fatalf("expected object at %s, got %#v", strings.Join(path, "."), value[name])
		}
		value = nested
	}
	return value
}

func codegenNamed(t *testing.T, values any, name string) map[string]any {
	t.Helper()
	objects, ok := values.([]any)
	if !ok {
		t.Fatalf("expected named objects, got %#v", values)
	}
	for _, value := range objects {
		object, ok := value.(map[string]any)
		if !ok {
			t.Fatalf("expected named object, got %#v", value)
		}
		if object["name"] == name {
			return object
		}
	}
	t.Fatalf("missing %s", name)
	return nil
}

func codegenWrite(t *testing.T, path string, contents []byte) {
	t.Helper()
	if err := os.WriteFile(path, contents, 0600); err != nil {
		t.Fatal(err)
	}
}

func codegenRun(t *testing.T, directory, tool string, args ...string) {
	t.Helper()
	command := exec.Command(tool, args...)
	command.Dir = directory
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", tool, args, err, output)
	}
}

const calculationGeneratorConfig = `provider:
  name: vantage
resources:
  custom_calculation:
    create: {path: /custom_calculations, method: POST}
    read: {path: "/custom_calculations/{custom_calculation_token}", method: GET}
    update: {path: "/custom_calculations/{custom_calculation_token}", method: PUT}
    delete: {path: "/custom_calculations/{custom_calculation_token}", method: DELETE}
    schema:
      attributes:
        aliases: {custom_calculation_token: token}
  cost_report:
    create: {path: /cost_reports, method: POST}
    read: {path: "/cost_reports/{cost_report_token}", method: GET}
    update: {path: "/cost_reports/{cost_report_token}", method: PUT}
    delete: {path: "/cost_reports/{cost_report_token}", method: DELETE}
    schema:
      attributes:
        aliases: {cost_report_token: token}
data_sources:
  custom_calculations:
    read: {path: /custom_calculations, method: GET}
`

const calculationGeneratedTest = `package generated_test

import (
  "context"
  "reflect"
  "testing"
  "github.com/hashicorp/terraform-plugin-framework/attr"
  resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
  schemaValidator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
  "github.com/hashicorp/terraform-plugin-framework/types"
  calculation "github.com/vantage-sh/terraform-provider-vantage/resource_custom_calculation"
  report "github.com/vantage-sh/terraform-provider-vantage/resource_cost_report"
)

var _ types.Set = calculation.CustomCalculationModel{}.CostReportTokens
var _ types.Set = report.CostReportModel{}.CustomCalculationTokens

func TestUsageUnitLists(t *testing.T) {
  ctx := context.Background()
  filterSchema := calculation.CustomCalculationResourceSchema(ctx).Attributes["filter"].(resourceSchema.SingleNestedAttribute)
  validators := filterSchema.Attributes["usage_unit"].(resourceSchema.ListAttribute).Validators
  for _, expected := range [][]string{{"GB"}, {"GB", "GiB"}} {
  units, diags := types.ListValueFrom(ctx, types.StringType, expected)
  if diags.HasError() { t.Fatal(diags) }
  for _, validation := range validators {
    var response schemaValidator.ListResponse
    validation.ValidateList(ctx, schemaValidator.ListRequest{ConfigValue: units}, &response)
    if response.Diagnostics.HasError() { t.Fatalf("valid units %v rejected: %v", expected, response.Diagnostics) }
  }
  attributes := map[string]attr.Value{
    "type": types.StringNull(), "filter": types.StringValue("costs"),
    "filter_numerator": types.StringNull(), "filter_denominator": types.StringNull(),
    "usage_unit": units, "usage_unit_numerator": types.ListNull(types.StringType),
    "usage_unit_denominator": types.ListNull(types.StringType),
  }
  value, diags := calculation.NewFilterValue(calculation.FilterValue{}.AttributeTypes(ctx), attributes)
  if diags.HasError() { t.Fatal(diags) }
  var actual []string
  diags = value.UsageUnit.ElementsAs(ctx, &actual, false)
  if diags.HasError() || !reflect.DeepEqual(actual, expected) { t.Fatalf("units=%v diagnostics=%v", actual, diags) }
  if _, err := value.ToTerraformValue(ctx); err != nil { t.Fatal(err) }
  }
}
`
