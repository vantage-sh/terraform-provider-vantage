package vantage

import (
	"testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAwsProviderResource_basic(t *testing.T) {
	resourceName := "vantage_aws_provider.demo"
	config := `
resource "vantage_aws_provider" "demo" {
  cross_account_arn = "arn:aws:iam::123456789012:role/TestRole"
  bucket_arn = "arn:aws:s3:::test-bucket"
}
`
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "cross_account_arn", "arn:aws:iam::123456789012:role/TestRole"),
					resource.TestCheckResourceAttr(resourceName, "bucket_arn", "arn:aws:s3:::test-bucket"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}