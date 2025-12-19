package vantage

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

// TestAssignCostReportTokens_OrderPreservation tests that assignCostReportTokens
// correctly reorders API response data to match the original plan order.
func TestAssignCostReportTokens_OrderPreservation(t *testing.T) {
	ctx := context.Background()

	attrTypes := map[string]attr.Type{
		"cost_report_token": types.StringType,
		"unit_scale":        types.StringType,
		"label_filter":      types.ListType{ElemType: types.StringType},
	}

	// Helper to create a cost report token object
	createToken := func(token, unitScale string) attr.Value {
		labelFilter, err := types.ListValueFrom(ctx, types.StringType, []string{})
		if err != nil {
			t.Fatalf("failed to create label_filter list: %v", err)
		}

		obj, err := types.ObjectValue(attrTypes, map[string]attr.Value{
			"cost_report_token": types.StringValue(token),
			"unit_scale":        types.StringValue(unitScale),
			"label_filter":      labelFilter,
		})
		if err != nil {
			t.Fatalf("failed to create cost report token object: %v", err)
		}
		return obj
	}

	// Plan order: token1, token2, token3
	planList, _ := types.ListValue(
		types.ObjectType{AttrTypes: attrTypes},
		[]attr.Value{
			createToken("rprt_token1", "per_unit"),
			createToken("rprt_token2", "per_thousand"),
			createToken("rprt_token3", "per_million"),
		},
	)

	// API returns in different order: token3, token1, token2
	apiList, _ := types.ListValue(
		types.ObjectType{AttrTypes: attrTypes},
		[]attr.Value{
			createToken("rprt_token3", "per_million"),
			createToken("rprt_token1", "per_unit"),
			createToken("rprt_token2", "per_thousand"),
		},
	)

	// Create model with API order
	data := &businessMetricResourceModel{
		CostReportTokensWithMetadata: apiList,
	}

	// Run the function
	var diags diag.Diagnostics
	assignCostReportTokens(ctx, data, planList, &diags)

	if diags.HasError() {
		t.Fatalf("assignCostReportTokens returned errors: %v", diags)
	}

	// Verify the order matches plan order
	tokens := make([]*businessMetricResourceModelCostReportToken, 0, 3)
	if d := data.CostReportTokensWithMetadata.ElementsAs(ctx, &tokens, false); d.HasError() {
		t.Fatalf("Failed to extract tokens: %v", d)
	}

	expectedOrder := []string{"rprt_token1", "rprt_token2", "rprt_token3"}
	for i, expected := range expectedOrder {
		if tokens[i].CostReportToken.ValueString() != expected {
			t.Errorf("Token at index %d: expected %q, got %q", i, expected, tokens[i].CostReportToken.ValueString())
		}
	}

	// Also verify computed values were preserved from API response
	expectedUnitScales := []string{"per_unit", "per_thousand", "per_million"}
	for i, expected := range expectedUnitScales {
		if tokens[i].UnitScale.ValueString() != expected {
			t.Errorf("UnitScale at index %d: expected %q, got %q", i, expected, tokens[i].UnitScale.ValueString())
		}
	}

	// Verify label_filter is an empty list, not null
	for i, token := range tokens {
		if token.LabelFilter.IsNull() {
			t.Errorf("LabelFilter at index %d: expected empty list, got null", i)
		}
		if len(token.LabelFilter.Elements()) != 0 {
			t.Errorf("LabelFilter at index %d: expected 0 elements, got %d", i, len(token.LabelFilter.Elements()))
		}
	}
}

// TestAssignCostReportTokens_NullLabelFilter tests that null label_filter from API
// is converted to an empty list to match Terraform's expectations.
func TestAssignCostReportTokens_NullLabelFilter(t *testing.T) {
	ctx := context.Background()

	attrTypes := map[string]attr.Type{
		"cost_report_token": types.StringType,
		"unit_scale":        types.StringType,
		"label_filter":      types.ListType{ElemType: types.StringType},
	}

	// Helper to create a cost report token with empty label_filter (plan)
	createTokenWithEmptyFilter := func(token, unitScale string) attr.Value {
		labelFilter, diags := types.ListValueFrom(ctx, types.StringType, []string{})
		if diags.HasError() {
			t.Fatalf("failed to create labelFilter list value: %v", diags)
		}

		obj, diags := types.ObjectValue(attrTypes, map[string]attr.Value{
			"cost_report_token": types.StringValue(token),
			"unit_scale":        types.StringValue(unitScale),
			"label_filter":      labelFilter,
		})
		if diags.HasError() {
			t.Fatalf("failed to create cost report token object (empty label_filter): %v", diags)
		}

		return obj
	}

	// Helper to create a cost report token with null label_filter (API response)
	createTokenWithNullFilter := func(token, unitScale string) attr.Value {
		obj, diags := types.ObjectValue(attrTypes, map[string]attr.Value{
			"cost_report_token": types.StringValue(token),
			"unit_scale":        types.StringValue(unitScale),
			"label_filter":      types.ListNull(types.StringType),
		})
		if diags.HasError() {
			t.Fatalf("failed to create cost report token object (null label_filter): %v", diags)
		}

		return obj
	}

	// Plan has empty list for label_filter
	planList, diags := types.ListValue(
		types.ObjectType{AttrTypes: attrTypes},
		[]attr.Value{
			createTokenWithEmptyFilter("rprt_token1", "per_unit"),
		},
	)
	if diags.HasError() {
		t.Fatalf("failed to create planList value: %v", diags)
	}

	// API returns null for label_filter
	apiList, diags := types.ListValue(
		types.ObjectType{AttrTypes: attrTypes},
		[]attr.Value{
			createTokenWithNullFilter("rprt_token1", "per_unit"),
		},
	)
	if diags.HasError() {
		t.Fatalf("failed to create apiList value: %v", diags)
	}

	// Create model with API response (null label_filter)
	data := &businessMetricResourceModel{
		CostReportTokensWithMetadata: apiList,
	}

	// Run the function
	assignCostReportTokens(ctx, data, planList, &diags)

	if diags.HasError() {
		t.Fatalf("assignCostReportTokens returned errors: %v", diags)
	}

	// Verify label_filter is an empty list, not null
	tokens := make([]*businessMetricResourceModelCostReportToken, 0, 1)
	if d := data.CostReportTokensWithMetadata.ElementsAs(ctx, &tokens, false); d.HasError() {
		t.Fatalf("Failed to extract tokens: %v", d)
	}

	if tokens[0].LabelFilter.IsNull() {
		t.Error("LabelFilter should be an empty list, not null")
	}
	if len(tokens[0].LabelFilter.Elements()) != 0 {
		t.Errorf("LabelFilter should have 0 elements, got %d", len(tokens[0].LabelFilter.Elements()))
	}
}

func TestAccBusinessMetric_basic(t *testing.T) {
	now := time.Now()
	date1 := fmt.Sprintf("%d-03-01", now.Year())
	date2 := fmt.Sprintf("%d-02-01", now.Year())
	date3 := fmt.Sprintf("%d-01-01", now.Year())
	tfValues := func(values []map[string]string) string {
		if values == nil {
			return ""
		}
		var valuesList []string
		for _, value := range values {
			var fields []string
			for k, v := range value {
				fields = append(fields, fmt.Sprintf(`%[1]q = %[2]q`, k, v))
			}
			valuesList = append(valuesList, fmt.Sprintf(`{ %s }`, strings.Join(fields, ",")))
		}

		return fmt.Sprintf(`values = [%[1]s]`, strings.Join(valuesList, ","))
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create without values
				Config: testAccVantageBusinessMetricTf_basic("test-no-values", "test", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vantage_business_metric.test-no-values", "token"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-no-values", "title", "test"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-no-values", "values.#", "0"),
				),
			},
			{ // update without values
				Config: testAccVantageBusinessMetricTf_basic("test-no-values", "updated-test", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vantage_business_metric.test-no-values", "token"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-no-values", "title", "updated-test"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-no-values", "values.#", "0"),
				),
			},
			{ // create with values
				Config: testAccVantageBusinessMetricTf_basic("test", "test", tfValues([]map[string]string{
					{"date": date1, "amount": "345.12"},
					{"date": date2, "amount": "123.45", "label": "a-label"},
				})),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vantage_business_metric.test", "token"),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "title", "test"),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.0.date", date1),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.0.amount", "345.12"),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.0.label", ""),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.1.date", date2),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.1.amount", "123.45"),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.1.label", "a-label"),
				),
			},
			{ // update values
				Config: testAccVantageBusinessMetricTf_basic("test", "test", tfValues([]map[string]string{
					{"date": date1, "amount": "345.12"},
					{"date": date2, "amount": "123.45", "label": "a-label"},
					{"date": date3, "amount": "123.45", "label": "a-label"},
				})),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vantage_business_metric.test", "token"),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "title", "test"),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.0.date", date1),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.0.amount", "345.12"),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.0.label", ""),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.1.date", date2),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.1.amount", "123.45"),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.1.label", "a-label"),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.2.date", date3),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.2.amount", "123.45"),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.2.label", "a-label"),
				),
			},
			{ // update the resource, but dont touch the values
				Config: testAccVantageBusinessMetricTf_basic("test", "updated-test", tfValues([]map[string]string{
					{"date": date1, "amount": "345.12"},
					{"date": date2, "amount": "123.45", "label": "a-label"},
					{"date": date3, "amount": "123.45", "label": "a-label"},
				})),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vantage_business_metric.test", "token"),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "title", "updated-test"),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.0.date", date1),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.0.amount", "345.12"),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.0.label", ""),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.1.date", date2),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.1.amount", "123.45"),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.1.label", "a-label"),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.2.date", date3),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.2.amount", "123.45"),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.2.label", "a-label"),
				),
			},
		},
	})
}

func testAccVantageBusinessMetricTf_basic(id string, title string, valuesStr string) string {
	return fmt.Sprintf(
		`data "vantage_business_metrics" %[1]q {}

		 resource "vantage_business_metric" %[1]q {
		   title = %[2]q
		   %[3]s
		 }
		`, id, title, valuesStr,
	)
}

func TestAccBusinessMetric_cloudwatch(t *testing.T) {
	t.Skip("Skipping until we have support for Integrations/AccessCredentials")
	// NOTE: You will also have to replace the hard coded access credential token in the test data.
	//
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageBusinessMetricTf_cloudwatch("test-cloudwatch", "CloudWatch Test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vantage_business_metric.test-cloudwatch", "token"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-cloudwatch", "title", "CloudWatch Test"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-cloudwatch", "cloudwatch_fields.metric_name", "CPUUtilization"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-cloudwatch", "cloudwatch_fields.namespace", "AWS/EC2"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-cloudwatch", "cloudwatch_fields.region", "us-east-1"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-cloudwatch", "cloudwatch_fields.stat", "Average"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-cloudwatch", "cloudwatch_fields.label_dimension", "InstanceId"),
				),
			},
			{
				// Test updating CloudWatch fields
				Config: testAccVantageBusinessMetricTf_cloudwatch_updated("test-cloudwatch", "CloudWatch Test Updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vantage_business_metric.test-cloudwatch", "token"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-cloudwatch", "title", "CloudWatch Test Updated"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-cloudwatch", "cloudwatch_fields.metric_name", "MemoryUtilization"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-cloudwatch", "cloudwatch_fields.namespace", "AWS/ECS"),
				),
			},
		},
	})
}

func testAccVantageBusinessMetricTf_cloudwatch(id string, title string) string {
	return fmt.Sprintf(`
resource "vantage_business_metric" %[1]q {
  title = %[2]q

  cloudwatch_fields = {
    dimensions = [
      {
        name = "InstanceId"
        value = "12345"
      }
    ]
		metric_name = "CPUUtilization"
    namespace = "AWS/EC2"
    region = "us-east-1"
    stat = "Average"
    label_dimension = "InstanceId"
    integration_token = "accss_crdntl_4e4a878a8f885856"
  }
}
`, id, title)
}

func testAccVantageBusinessMetricTf_cloudwatch_updated(id string, title string) string {
	return fmt.Sprintf(`
resource "vantage_business_metric" %[1]q {
  title = %[2]q

  cloudwatch_fields = {
	  dimensions = [
	    {
	      name = "InstanceId"
	      value = "12345"
	    }
	  ]
    metric_name = "MemoryUtilization"
    namespace = "AWS/ECS"
    region = "us-east-1"
    stat = "Maximum"
    label_dimension = "ClusterName"
    integration_token = "accss_crdntl_4e4a878a8f885856"
  }
}
`, id, title)
}

func TestAccBusinessMetric_datadog(t *testing.T) {
	t.Skip("Skipping until we have support for Integrations/AccessCredentials")
	// NOTE: You will also have to replace the hard coded access credential token in the test data.

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageBusinessMetricTf_datadog("test-datadog", "Datadog Test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vantage_business_metric.test-datadog", "token"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-datadog", "title", "Datadog Test"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-datadog", "datadog_metric_fields.query", "avg:system.cpu.user{*}.rollup(avg, daily)"),
				),
			},
			{
				// Test updating Datadog query
				Config: testAccVantageBusinessMetricTf_datadog_updated("test-datadog", "Datadog Test Updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vantage_business_metric.test-datadog", "token"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-datadog", "title", "Datadog Test Updated"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-datadog", "datadog_metric_fields.query", "avg:system.memory.used{*}.rollup(avg, daily)"),
				),
			},
		},
	})
}

func testAccVantageBusinessMetricTf_datadog(id string, title string) string {
	return fmt.Sprintf(`
resource "vantage_business_metric" %[1]q {
  title = %[2]q
  datadog_metric_fields = {
    query = "avg:system.cpu.user{*}.rollup(avg, daily)"
    integration_token = "accss_crdntl_27a4dff7012ecce3"
  }
}
`, id, title)
}

func testAccVantageBusinessMetricTf_datadog_updated(id string, title string) string {
	return fmt.Sprintf(`
resource "vantage_business_metric" %[1]q {
  title = %[2]q
  datadog_metric_fields = {
    query = "avg:system.memory.used{*}.rollup(avg, daily)"
    integration_token = "accss_crdntl_27a4dff7012ecce3"
  }
}
`, id, title)
}

func TestAccBusinessMetric_costReportTokensOrder(t *testing.T) {
	// This test verifies that cost_report_tokens_with_metadata maintains the order
	// specified in the Terraform config, even if the API returns them in a different order.
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create business metric with multiple cost report tokens
				Config: testAccVantageBusinessMetricTf_withCostReportTokens("test-tokens", "test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vantage_business_metric.test-tokens", "token"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-tokens", "title", "test"),
					// Verify the order matches what we specified in config
					resource.TestCheckResourceAttrPair(
						"vantage_business_metric.test-tokens", "cost_report_tokens_with_metadata.0.cost_report_token",
						"vantage_cost_report.report1", "token",
					),
					resource.TestCheckResourceAttrPair(
						"vantage_business_metric.test-tokens", "cost_report_tokens_with_metadata.1.cost_report_token",
						"vantage_cost_report.report2", "token",
					),
					resource.TestCheckResourceAttrPair(
						"vantage_business_metric.test-tokens", "cost_report_tokens_with_metadata.2.cost_report_token",
						"vantage_cost_report.report3", "token",
					),
					resource.TestCheckResourceAttr("vantage_business_metric.test-tokens", "cost_report_tokens_with_metadata.0.unit_scale", "per_unit"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-tokens", "cost_report_tokens_with_metadata.1.unit_scale", "per_thousand"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-tokens", "cost_report_tokens_with_metadata.2.unit_scale", "per_million"),
				),
			},
			{ // update title but keep same cost report tokens order
				Config: testAccVantageBusinessMetricTf_withCostReportTokens("test-tokens", "updated-test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_business_metric.test-tokens", "title", "updated-test"),
					// Verify order is still preserved after update
					resource.TestCheckResourceAttrPair(
						"vantage_business_metric.test-tokens", "cost_report_tokens_with_metadata.0.cost_report_token",
						"vantage_cost_report.report1", "token",
					),
					resource.TestCheckResourceAttrPair(
						"vantage_business_metric.test-tokens", "cost_report_tokens_with_metadata.1.cost_report_token",
						"vantage_cost_report.report2", "token",
					),
					resource.TestCheckResourceAttrPair(
						"vantage_business_metric.test-tokens", "cost_report_tokens_with_metadata.2.cost_report_token",
						"vantage_cost_report.report3", "token",
					),
				),
			},
		},
	})
}

func testAccVantageBusinessMetricTf_withCostReportTokens(id string, title string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_report" "report1" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = "Report 1 for Business Metric Test"
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_month"
}

resource "vantage_cost_report" "report2" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = "Report 2 for Business Metric Test"
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_month"
}

resource "vantage_cost_report" "report3" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = "Report 3 for Business Metric Test"
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_month"
}

resource "vantage_business_metric" %[1]q {
  title = %[2]q
  
  cost_report_tokens_with_metadata = [
    {
      cost_report_token = vantage_cost_report.report1.token
      unit_scale        = "per_unit"
    },
    {
      cost_report_token = vantage_cost_report.report2.token
      unit_scale        = "per_thousand"
    },
    {
      cost_report_token = vantage_cost_report.report3.token
      unit_scale        = "per_million"
    }
  ]
}
`, id, title)
}

// TestAccBusinessMetric_withValuesAndEmptyLabelFilter tests the exact scenario
// from customer issue: business metric with CSV values and label_filter = []
func TestAccBusinessMetric_withValuesAndEmptyLabelFilter(t *testing.T) {
	now := time.Now()
	date1 := fmt.Sprintf("%d-01-01", now.Year())
	date2 := fmt.Sprintf("%d-02-01", now.Year())
	date3 := fmt.Sprintf("%d-03-01", now.Year())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageBusinessMetricTf_withValuesAndEmptyLabelFilter(
					"test-customer-scenario",
					"OV - Test",
					date1, date2, date3,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vantage_business_metric.test-customer-scenario", "token"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-customer-scenario", "title", "OV - Test"),
					// Verify values are set
					resource.TestCheckResourceAttr("vantage_business_metric.test-customer-scenario", "values.#", "3"),
					// Verify cost_report_tokens_with_metadata
					resource.TestCheckResourceAttr("vantage_business_metric.test-customer-scenario", "cost_report_tokens_with_metadata.#", "1"),
					// Verify label_filter is empty list, not null
					resource.TestCheckResourceAttr("vantage_business_metric.test-customer-scenario", "cost_report_tokens_with_metadata.0.label_filter.#", "0"),
				),
			},
			{ // Update title to verify no drift on label_filter
				Config: testAccVantageBusinessMetricTf_withValuesAndEmptyLabelFilter(
					"test-customer-scenario",
					"OV - Test Updated",
					date1, date2, date3,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_business_metric.test-customer-scenario", "title", "OV - Test Updated"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-customer-scenario", "cost_report_tokens_with_metadata.0.label_filter.#", "0"),
				),
			},
		},
	})
}

func testAccVantageBusinessMetricTf_withValuesAndEmptyLabelFilter(id, title, date1, date2, date3 string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_report" "product_line" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = "Product Line Report"
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_month"
}

# Simulating CSV data that would come from csvdecode()
locals {
  order_volume_data = [
    { date = %[3]q, amount = "1000.50" },
    { date = %[4]q, amount = "1500.75" },
    { date = %[5]q, amount = "2000.25" },
  ]
}

resource "vantage_business_metric" %[1]q {
  title  = %[2]q
  values = local.order_volume_data

  cost_report_tokens_with_metadata = [
    {
      cost_report_token = vantage_cost_report.product_line.token
      label_filter      = []
    }
  ]
}
`, id, title, date1, date2, date3)
}

func TestAccBusinessMetric_costReportTokensWithReferences(t *testing.T) {
	// This test verifies the exact pattern shown in the GitHub PR image:
	// Using Terraform references to other cost reports in cost_report_tokens_with_metadata
	// with multiple reports and empty label_filter arrays
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create business metric with multiple cost report token references
				Config: testAccVantageBusinessMetricTf_withMultipleCostReportReferences("test-fills-trades", "Fills (Trades)"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vantage_business_metric.test-fills-trades", "token"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-fills-trades", "title", "Fills (Trades)"),
					// Verify we have 4 cost report tokens
					resource.TestCheckResourceAttr("vantage_business_metric.test-fills-trades", "cost_report_tokens_with_metadata.#", "4"),
					// Verify the tokens are properly linked to the cost reports
					resource.TestCheckResourceAttrPair(
						"vantage_business_metric.test-fills-trades", "cost_report_tokens_with_metadata.0.cost_report_token",
						"vantage_cost_report.all_resources", "token",
					),
					resource.TestCheckResourceAttrPair(
						"vantage_business_metric.test-fills-trades", "cost_report_tokens_with_metadata.1.cost_report_token",
						"vantage_cost_report.domains", "token",
					),
					resource.TestCheckResourceAttrPair(
						"vantage_business_metric.test-fills-trades", "cost_report_tokens_with_metadata.2.cost_report_token",
						"vantage_cost_report.main_view", "token",
					),
					resource.TestCheckResourceAttrPair(
						"vantage_business_metric.test-fills-trades", "cost_report_tokens_with_metadata.3.cost_report_token",
						"vantage_cost_report.providers", "token",
					),
					// Verify all have per_unit scale
					resource.TestCheckResourceAttr("vantage_business_metric.test-fills-trades", "cost_report_tokens_with_metadata.0.unit_scale", "per_unit"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-fills-trades", "cost_report_tokens_with_metadata.1.unit_scale", "per_unit"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-fills-trades", "cost_report_tokens_with_metadata.2.unit_scale", "per_unit"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-fills-trades", "cost_report_tokens_with_metadata.3.unit_scale", "per_unit"),
					// Verify label_filter is empty list (not null)
					resource.TestCheckResourceAttr("vantage_business_metric.test-fills-trades", "cost_report_tokens_with_metadata.0.label_filter.#", "0"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-fills-trades", "cost_report_tokens_with_metadata.1.label_filter.#", "0"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-fills-trades", "cost_report_tokens_with_metadata.2.label_filter.#", "0"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-fills-trades", "cost_report_tokens_with_metadata.3.label_filter.#", "0"),
				),
			},
			{ // update to verify no drift occurs
				Config: testAccVantageBusinessMetricTf_withMultipleCostReportReferences("test-fills-trades", "Fills (Trades)"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_business_metric.test-fills-trades", "title", "Fills (Trades)"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-fills-trades", "cost_report_tokens_with_metadata.#", "4"),
				),
			},
		},
	})
}

func testAccVantageBusinessMetricTf_withMultipleCostReportReferences(id string, title string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_report" "all_resources" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = "All Resources"
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_month"
}

resource "vantage_cost_report" "domains" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = "Domains"
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_month"
}

resource "vantage_cost_report" "main_view" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = "Main View"
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_month"
}

resource "vantage_cost_report" "providers" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = "Providers"
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_month"
}

resource "vantage_business_metric" %[1]q {
  title = %[2]q
  
  cost_report_tokens_with_metadata = [
    {
      cost_report_token = vantage_cost_report.all_resources.token
      unit_scale        = "per_unit"
      label_filter      = []
    },
    {
      cost_report_token = vantage_cost_report.domains.token
      unit_scale        = "per_unit"
      label_filter      = []
    },
    {
      cost_report_token = vantage_cost_report.main_view.token
      unit_scale        = "per_unit"
      label_filter      = []
    },
    {
      cost_report_token = vantage_cost_report.providers.token
      unit_scale        = "per_unit"
      label_filter      = []
    }
  ]
}
`, id, title)
}
