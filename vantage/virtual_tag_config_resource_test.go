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
	keyCollapsed  string
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
		keyCollapsed:  keyV0 + "-collapsed",
		resourceName:  "vantage_virtual_tag_config.test",
	}

	collapsedResourceId := "collapsed-tag-keys"

	resourceName := func(key string) string {
		return fmt.Sprintf("vantage_virtual_tag_config.%s", key)
	}

	fromState := func(resourceId, key, field string) string {
		return fmt.Sprintf(
			`{ for vtag in data.vantage_virtual_tag_configs.%[1]s.virtual_tag_configs : vtag.key => vtag }[%[2]q].%[3]s`,
			resourceId,
			key,
			field,
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
					},
					{
						filter = "(costs.provider = 'gcp' AND costs.service != 'ComputeEngine')"
						percentages = [
							{
								pct = 25
								value = "Marketing"
							},
							{
								pct = 65
								value = "Engineering"
							},
							{
								pct = 10
								value = "Support"
							},
						]
					}
				]
				`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ctx.resourceName, "key", ctx.keyV0),
					resource.TestCheckResourceAttr(ctx.resourceName, "overridable", "true"),
					resource.TestCheckResourceAttr(ctx.resourceName, "backfill_until", ctx.backfillUntil),
					resource.TestCheckResourceAttrSet(ctx.resourceName, "token"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.#", "3"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.0.name", "value-0"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.0.filter", "(costs.provider = 'aws' AND costs.service = 'AmazonEC2') OR (costs.provider = 'gcp' AND costs.service = 'ComputeEngine')"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.1.filter", "(costs.provider = 'aws' AND costs.service = 'AwsApiGateway')"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.1.cost_metric.aggregation.tag", "environment"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.1.cost_metric.filter", "(costs.provider = 'aws' AND costs.service = 'AmazonECS')"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.2.filter", "(costs.provider = 'gcp' AND costs.service != 'ComputeEngine')"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.2.percentages.#", "3"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.2.percentages.0.pct", "25"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.2.percentages.0.value", "Marketing"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.2.percentages.1.pct", "65"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.2.percentages.1.value", "Engineering"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.2.percentages.2.pct", "10"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.2.percentages.2.value", "Support"),
				),
			},
			// Update: not specifying values
			// TODO(cp): I believe this is not working as intended - I'd expect the provider to support modifying a pre-existing virtual tag config
			// without specifying values and have the values unchanged (the TF provider submits values as null).
			//
			// In the interest of changing as little as possible for now + supporting the test behavior, values are cleared in this scenario.
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
						},
						{
							filter = "(costs.provider = 'aws' AND costs.service = 'AwsApiGateway')"
							cost_metric = {
								aggregation = {
									tag = "environment"
								}
								filter = "(costs.provider = 'aws' AND costs.service = 'AmazonECS')"
							}
						},
						{
							filter = "(costs.provider = 'gcp' AND costs.service = 'ComputeEngine')"
							percentages = [
								{
									pct = 30
									value = "Marketing"
								},
								{
									pct = 55
									value = "Engineering"
								},
								{
									pct = 15
									value = "Support"
								},
							]
						}
					]`,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(ctx.resourceName, "key", ctx.keyV1),
					resource.TestCheckResourceAttr(ctx.resourceName, "overridable", "false"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.#", "4"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.0.filter", "(costs.provider = 'aws' AND costs.service = 'AmazonEC2') OR (costs.provider = 'gcp' AND costs.service = 'ComputeEngine')"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.0.name", "value-0"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.1.filter", "(costs.provider = 'gcp' AND costs.service != 'ComputeEngine')"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.1.name", "value-1"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.2.filter", "(costs.provider = 'aws' AND costs.service = 'AwsApiGateway')"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.2.cost_metric.filter", "(costs.provider = 'aws' AND costs.service = 'AmazonECS')"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.2.cost_metric.aggregation.tag", "environment"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.3.filter", "(costs.provider = 'gcp' AND costs.service = 'ComputeEngine')"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.3.percentages.#", "3"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.3.percentages.0.pct", "30"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.3.percentages.0.value", "Marketing"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.3.percentages.1.pct", "55"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.3.percentages.1.value", "Engineering"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.3.percentages.2.pct", "15"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.3.percentages.2.value", "Support"),
				),
			},
			// Update: preserving existing values
			// TODO: I don't know what I thought I was testing here. re-using state via `fromState` feels out of scope of this resource test.
			{
				Config: testAccVantageVirtualTagConfig_basicTf(
					"test",
					ctx.keyV1,
					!ctx.overridable,
					ctx.backfillUntil,
					fmt.Sprintf(`values = %[1]s`, fromState("test", ctx.keyV1, "values")),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(ctx.resourceName, "key", ctx.keyV1),
					resource.TestCheckResourceAttr(ctx.resourceName, "overridable", "false"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.#", "4"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.0.name", "value-0"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.1.name", "value-1"),
					resource.TestCheckResourceAttrSet(ctx.resourceName, "values.2.cost_metric.filter"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.3.percentages.#", "3"),
				),
			},
			// Update: removing values
			{
				Config: testAccVantageVirtualTagConfig_basicTf("test", ctx.keyV1, !ctx.overridable, ctx.backfillUntil, "values = []"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(ctx.resourceName, "key", ctx.keyV1),
					resource.TestCheckResourceAttr(ctx.resourceName, "overridable", "false"),
					resource.TestCheckResourceAttr(ctx.resourceName, "values.#", "0"),
				),
			},
			// -- collapsed tag keys --
			// Create: with collapsed tag keys
			{
				Config: testAccVantageVirtualTagConfig_basicTf(
					collapsedResourceId,
					ctx.keyCollapsed,
					ctx.overridable,
					ctx.backfillUntil,
					`
					collapsed_tag_keys = [
						{
							key = "environment"
						},
						{
							key = "service"
						},
						{
							key = "project"
							providers = ["aws", "gcp"]
						}
					]`,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName(collapsedResourceId), "collapsed_tag_keys.#", "3"),
					resource.TestCheckResourceAttr(resourceName(collapsedResourceId), "collapsed_tag_keys.0.key", "environment"),
					resource.TestCheckResourceAttr(resourceName(collapsedResourceId), "collapsed_tag_keys.1.key", "service"),
					resource.TestCheckResourceAttr(resourceName(collapsedResourceId), "collapsed_tag_keys.2.key", "project"),
					resource.TestCheckResourceAttr(resourceName(collapsedResourceId), "collapsed_tag_keys.2.providers.#", "2"),
					resource.TestCheckResourceAttr(resourceName(collapsedResourceId), "collapsed_tag_keys.2.providers.0", "aws"),
					resource.TestCheckResourceAttr(resourceName(collapsedResourceId), "collapsed_tag_keys.2.providers.1", "gcp"),
				),
			},
			// -- collapsed tag keys --
			// Update: replacing collapsed tag keys
			{
				Config: testAccVantageVirtualTagConfig_basicTf(
					collapsedResourceId,
					ctx.keyCollapsed,
					ctx.overridable,
					ctx.backfillUntil,
					`
					collapsed_tag_keys = [
						{
							key = "some-new-key"
						},
					]`,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName(collapsedResourceId), "collapsed_tag_keys.#", "1"),
					resource.TestCheckResourceAttr(resourceName(collapsedResourceId), "collapsed_tag_keys.0.key", "some-new-key"),
				),
			},
		},
	})
}

func testAccVantageVirtualTagConfig_basicTf(id string, key string, overridable bool, backfillUntil string, rest string) string {
	return fmt.Sprintf(
		`data "vantage_virtual_tag_configs" %[1]q {}

		 resource "vantage_virtual_tag_config" %[1]q {
		   key = %[2]q
		   overridable = %[3]t
		   backfill_until = %[4]q
		   %[5]s
		 }
		`, id, key, overridable, backfillUntil, rest,
	)
}
