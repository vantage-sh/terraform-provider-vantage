package vantage

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccAwsProviderResource_basic(t *testing.T) {
	t.Skip("Requires a real AWS cross-account ARN in a Vantage-accessible AWS account")

	resourceName := "vantage_aws_provider.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsProviderConfig(
					"arn:aws:iam::123456789012:role/TestRole",
					"arn:aws:s3:::test-bucket",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "cross_account_arn", "arn:aws:iam::123456789012:role/TestRole"),
					resource.TestCheckResourceAttr(resourceName, "bucket_arn", "arn:aws:s3:::test-bucket"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

func testAccAwsProviderConfig(crossAccountArn, bucketArn string) string {
	return `
resource "vantage_aws_provider" "test" {
  cross_account_arn = "` + crossAccountArn + `"
  bucket_arn        = "` + bucketArn + `"
}
`
}
