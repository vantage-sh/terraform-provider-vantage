package vantage

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccBusinessMetric_basic(t *testing.T) {
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
			{ // create with values
				Config: testAccVantageBusinessMetricTf_basic("test", "test", tfValues([]map[string]string{
					{"date": "2024-01-10", "amount": "345.12"},
					{"date": "2024-01-01", "amount": "123.45", "label": "a-label"},
				})),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vantage_business_metric.test", "token"),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "title", "test"),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.0.date", "2024-01-10"),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.0.amount", "345.12"),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.0.label", ""),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.1.date", "2024-01-01"),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.1.amount", "123.45"),
					resource.TestCheckResourceAttr("vantage_business_metric.test", "values.1.label", "a-label"),
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
