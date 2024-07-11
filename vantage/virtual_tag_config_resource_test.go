package vantage

import (
	"fmt"
	"strings"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccVantageVirtualTagConfig_basic(t *testing.T) {
	keyV0 := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	overridable := true
	backfillUntil := "2024-03-01"
	keyPre := keyV0 + "-pre"
	keyV1 := keyV0 + "-updated"
	resourceName := "vantage_virtual_tag_config.test"

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

	fromState := func(key, field string) string {
		return fmt.Sprintf(
			`{ for vtag in data.vantage_virtual_tag_configs.test.virtual_tag_configs : vtag.key => vtag }[%[1]q].%[2]s`,
			key, field,
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create: without values
			{
				Config: testAccVantageVirtualTagConfigTf_basic("test-no-values", keyPre, overridable, backfillUntil, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_virtual_tag_config.test-no-values", "key", keyPre),
					resource.TestCheckResourceAttr("vantage_virtual_tag_config.test-no-values", "overridable", "true"),
					resource.TestCheckResourceAttr("vantage_virtual_tag_config.test-no-values", "backfill_until", "2024-03-01"),
					resource.TestCheckResourceAttrSet("vantage_virtual_tag_config.test-no-values", "token"),
					resource.TestCheckResourceAttr("vantage_virtual_tag_config.test-no-values", "values.#", "0"),
				),
			},
			// Create: with values
			{
				Config: testAccVantageVirtualTagConfigTf_basic("test", keyV0, overridable, backfillUntil, tfValues([]map[string]string{
					{
						"name":   "value-0",
						"filter": "(costs.provider = 'aws' AND costs.service = 'AmazonEC2') OR (costs.provider = 'gcp' AND costs.service = 'ComputeEngine')",
					},
				})),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", keyV0),
					resource.TestCheckResourceAttr(resourceName, "overridable", "true"),
					resource.TestCheckResourceAttr(resourceName, "backfill_until", "2024-03-01"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttr(resourceName, "values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "values.0.name", "value-0"),
				),
			},
			// Update: not specifying values
			{
				Config: testAccVantageVirtualTagConfigTf_basic(
					"test",
					keyV1,
					!overridable,
					backfillUntil,
					"",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", keyV1),
					resource.TestCheckResourceAttr(resourceName, "overridable", "false"),
					resource.TestCheckResourceAttr(resourceName, "values.#", "0"),
				),
			},
			// Update: set multiple values with filters
			{
				Config: testAccVantageVirtualTagConfigTf_basic(
					"test",
					keyV1,
					!overridable,
					backfillUntil,
					tfValues(
						[]map[string]string{
							{
								"name":   "value-0",
								"filter": "(costs.provider = 'aws' AND costs.service = 'AmazonEC2') OR (costs.provider = 'gcp' AND costs.service = 'ComputeEngine')",
							},
							{
								"name":   "value-1",
								"filter": "(costs.provider = 'gcp' AND costs.service != 'ComputeEngine')",
							},
						},
					),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", keyV1),
					resource.TestCheckResourceAttr(resourceName, "overridable", "false"),
					resource.TestCheckResourceAttr(resourceName, "values.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "values.0.name", "value-0"),
					resource.TestCheckResourceAttr(resourceName, "values.1.name", "value-1"),
				),
			},
			// Update: preserving existing values
			{
				Config: testAccVantageVirtualTagConfigTf_basic(
					"test",
					keyV1,
					!overridable,
					backfillUntil,
					fmt.Sprintf(`values = %[1]s`, fromState(keyV1, "values")),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", keyV1),
					resource.TestCheckResourceAttr(resourceName, "overridable", "false"),
					resource.TestCheckResourceAttr(resourceName, "values.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "values.0.name", "value-0"),
					resource.TestCheckResourceAttr(resourceName, "values.1.name", "value-1"),
				),
			},
			// // Update: adding to existing values
			// {
			// 	Config: testAccVantageVirtualTagConfigTf_basic(
			// 		"test",
			// 		keyV1,
			// 		!overridable,
			// 		backfillUntil,
			// 		fmt.Sprintf(`values = concat(%[1]s, [{ name = "value-2" }])`, fromState(keyV1, "values")),
			// 	),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttr(resourceName, "key", keyV1),
			// 		resource.TestCheckResourceAttr(resourceName, "overridable", "false"),
			// 		resource.TestCheckResourceAttr(resourceName, "values.#", "3"),
			// 		resource.TestCheckResourceAttr(resourceName, "values.0.name", "value-0"),
			// 		resource.TestCheckResourceAttr(resourceName, "values.1.name", "value-1"),
			// 		resource.TestCheckResourceAttr(resourceName, "values.2.name", "value-2"),
			// 	),
			// },
			// Update: removing values
			{
				Config: testAccVantageVirtualTagConfigTf_basic("test", keyV1, !overridable, backfillUntil, "values = []"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", keyV1),
					resource.TestCheckResourceAttr(resourceName, "overridable", "false"),
					resource.TestCheckResourceAttr(resourceName, "values.#", "0"),
				),
			},
		},
	})
}

func testAccVantageVirtualTagConfigTf_basic(id string, key string, overridable bool, backfillUntil string, valuesStr string) string {
	return fmt.Sprintf(
		`data "vantage_virtual_tag_configs" %[1]q {}

		 resource "vantage_virtual_tag_config" %[1]q {
		   key = %[2]q
		   overridable = %[3]t
		   backfill_until = %[4]q
		   %[5]s
		 }
		`, id, key, overridable, backfillUntil, valuesStr,
	)
}
