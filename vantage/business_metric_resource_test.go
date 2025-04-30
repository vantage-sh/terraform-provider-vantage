package vantage

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

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
    metric_name = "CPUUtilization"
    namespace = "AWS/EC2"
    region = "us-east-1"
    stat = "Average"
    label_dimension = "InstanceId"
  }
}
`, id, title)
}

func testAccVantageBusinessMetricTf_cloudwatch_updated(id string, title string) string {
	return fmt.Sprintf(`
resource "vantage_business_metric" %[1]q {
  title = %[2]q

  cloudwatch_fields = {
    metric_name = "MemoryUtilization"
    namespace = "AWS/ECS"
    region = "us-east-1"
    stat = "Maximum"
    label_dimension = "ClusterName"
  }
}
`, id, title)
}

func TestAccBusinessMetric_datadog(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageBusinessMetricTf_datadog("test-datadog", "Datadog Test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vantage_business_metric.test-datadog", "token"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-datadog", "title", "Datadog Test"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-datadog", "datadog_metric_fields.query", "avg:system.cpu.user{*}"),
				),
			},
			{
				// Test updating Datadog query
				Config: testAccVantageBusinessMetricTf_datadog_updated("test-datadog", "Datadog Test Updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vantage_business_metric.test-datadog", "token"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-datadog", "title", "Datadog Test Updated"),
					resource.TestCheckResourceAttr("vantage_business_metric.test-datadog", "datadog_metric_fields.query", "avg:system.memory.used{*}"),
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
    query = "avg:system.cpu.user{*}"
  }
}
`, id, title)
}

func testAccVantageBusinessMetricTf_datadog_updated(id string, title string) string {
	return fmt.Sprintf(`
resource "vantage_business_metric" %[1]q {
  title = %[2]q
  datadog_metric_fields = {
    query = "avg:system.memory.used{*}"
  }
}
`, id, title)
}
