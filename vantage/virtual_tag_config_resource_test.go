package vantage

import (
	"fmt"
	"testing"
	"time"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

type testAccVantageVirtualTagConfig_basicContext struct {
	overridable   bool
	backfillUntil string
	keyPre        string
	keyV0         string
	keyV1         string
	resourceName  string
}

func TestAccVantageVirtualTagConfig_basic(t *testing.T) {

	keyV0 := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	now := time.Now()

	ctx := testAccVantageVirtualTagConfig_basicContext{
		overridable: true,
		// 3 months ago, beginning of month
		backfillUntil: now.AddDate(0, -3, -now.Day()+1).Format("2006-01-02"),
		keyPre:        keyV0 + "-pre",
		keyV0:         keyV0,
		keyV1:         keyV0 + "-updated",
		resourceName:  "vantage_virtual_tag_config.test",
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
				Config: testAccVantageVirtualTagConfig_basicTf("test-no-values", ctx.keyPre, ctx.overridable, ctx.backfillUntil, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_virtual_tag_config.test-no-values", "key", ctx.keyPre),
					resource.TestCheckResourceAttr("vantage_virtual_tag_config.test-no-values", "overridable", "true"),
					resource.TestCheckResourceAttr("vantage_virtual_tag_config.test-no-values", "backfill_until", ctx.backfillUntil),
					resource.TestCheckResourceAttrSet("vantage_virtual_tag_config.test-no-values", "token"),
					resource.TestCheckResourceAttr("vantage_virtual_tag_config.test-no-values", "values.#", "0"),
				),
			},
			// Create: with values
			{
				Config: testAccVantageVirtualTagConfig_basicTf("test", ctx.keyV0, ctx.overridable, ctx.backfillUntil, `
				values = [
					{
						name = "value-0"
						filter = "(costs.provider = 'aws' AND costs.service = 'AmazonEC2') OR (costs.provider = 'gcp' AND costs.service = 'ComputeEngine')"
					},
					{
						filter = "(costs.provider = 'aws' AND costs.service = 'AwsApiGateway')"
						cost_metric = {
							aggregation = {
								tag = "environment"
							}
							filter = "(costs.provider = 'aws' AND costs.service = 'AmazonECS')"
						}
					}
				]
				`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ctx.resourceName, "key", ctx.keyV0),
					resource.TestCheckResourceAttr(ctx.resourceName, "overridable", "true"),
					resource.TestCheckResourceAttr(ctx.resourceName, "backfill_until", ctx.backfillUntil),
					resource.TestCheckResourceAttrSet(ctx.resourceName, "token"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.#", "2"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.0.name", "value-0"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.0.filter", "(costs.provider = 'aws' AND costs.service = 'AmazonEC2') OR (costs.provider = 'gcp' AND costs.service = 'ComputeEngine')"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.1.filter", "(costs.provider = 'aws' AND costs.service = 'AwsApiGateway')"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.1.cost_metric.aggregation.tag", "environment"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.1.cost_metric.filter", "(costs.provider = 'aws' AND costs.service = 'AmazonECS')"),
				),
			},
			// Update: not specifying values
			{
				Config: testAccVantageVirtualTagConfig_basicTf(
					"test",
					ctx.keyV1,
					!ctx.overridable,
					ctx.backfillUntil,
					"",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(ctx.resourceName, "key", ctx.keyV1),
					resource.TestCheckResourceAttr(ctx.resourceName, "overridable", "false"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.#", "0"),
				),
			},
			// Update: set multiple values with filters
			{
				Config: testAccVantageVirtualTagConfig_basicTf("test", ctx.keyV1, !ctx.overridable, ctx.backfillUntil, `
				values = [
					{
						name = "value-0"
						filter = "(costs.provider = 'aws' AND costs.service = 'AmazonEC2') OR (costs.provider = 'gcp' AND costs.service = 'ComputeEngine')"
					},
					{
						name = "value-1"
						filter = "(costs.provider = 'gcp' AND costs.service != 'ComputeEngine')"
					}
				]
				`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(ctx.resourceName, "key", ctx.keyV1),
					resource.TestCheckResourceAttr(ctx.resourceName, "overridable", "false"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.#", "2"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.0.name", "value-0"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.1.name", "value-1"),
				),
			},
			// Update: preserving existing values
			{
				Config: testAccVantageVirtualTagConfig_basicTf(
					"test",
					ctx.keyV1,
					!ctx.overridable,
					ctx.backfillUntil,
					fmt.Sprintf(`values = %[1]s`, fromState(ctx.keyV1, "values")),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(ctx.resourceName, "key", ctx.keyV1),
					resource.TestCheckResourceAttr(ctx.resourceName, "overridable", "false"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.#", "2"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.0.name", "value-0"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.1.name", "value-1"),
				),
			},
			// // Update: adding to existing values
			// {
			// 	Config: testAccVantageVirtualTagConfig_basicTf(
			// 		"test",
			// 		ctx.keyV1,
			// 		!ctx.overridable,
			// 		ctx.backfillUntil,
			// 		fmt.Sprintf(`values = concat(%[1]s, [{ name = "value-2" }])`, fromState(ctx.keyV1, "values")),
			// 	),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttr(ctx.resourceName, "key", ctx.keyV1),
			// 		resource.TestCheckResourceAttr(ctx.resourceName, "overridable", "false"),
			// 		resource.TestCheckResourceAttr(ctx.resourceName, "values.#", "3"),
			// 		resource.TestCheckResourceAttr(ctx.resourceName, "values.0.name", "value-0"),
			// 		resource.TestCheckResourceAttr(ctx.resourceName, "values.1.name", "value-1"),
			// 		resource.TestCheckResourceAttr(ctx.resourceName, "values.2.name", "value-2"),
			// 	),
			// },
			// Update: removing values
			{
				Config: testAccVantageVirtualTagConfig_basicTf("test", ctx.keyV1, !ctx.overridable, ctx.backfillUntil, "values = []"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(ctx.resourceName, "key", ctx.keyV1),
					resource.TestCheckResourceAttr(ctx.resourceName, "overridable", "false"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.#", "0"),
				),
			},
		},
	})
}

func testAccVantageVirtualTagConfig_basicTf(id string, key string, overridable bool, backfillUntil string, valuesStr string) string {
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
