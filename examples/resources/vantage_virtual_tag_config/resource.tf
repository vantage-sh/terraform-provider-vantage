data "vantage_virtual_tag_configs" "demo" {}
resource "vantage_virtual_tag_config" "demo_virtual_tag_config" {
  key            = "Demo Tag"
  backfill_until = "2024-01-01"
  overridable    = true
  values = [
    {
      name   = "Demo Value 0"
      filter = "(costs.provider = 'aws' AND costs.region = 'us-east-1') OR (costs.provider = 'gcp' AND costs.region = 'us-central1')"
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
    # {
    #   filter = "(costs.provider = 'aws' AND costs.service = 'AmazonECS')"
    #   business_metric_token = ""
    # }
    # Example: apply label_transforms to a business-metric-backed value to
    # split a "&&&"-delimited project label and reformat it into a team tag.
    # {
    #   filter                = "costs.provider = 'aws'"
    #   business_metric_token = "bsnss_mtrc_XXXXXXXX"
    #   label_transforms = [
    #     { type = "split", delimiter = "&&&", index = 0 },
    #     { type = "format", template = "team-{0}" },
    #   ]
    # }
  ]
}
