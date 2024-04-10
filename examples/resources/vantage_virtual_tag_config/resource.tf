data "vantage_virtual_tag_configs" "demo" {}
resource "vantage_virtual_tag_config" "demo_virtual_tag_config" {
  key = "Demo Tag"
  backfill_until = "2024-01-01"
  overridable = true
  values = [
    {
      name = "Demo Value 0"
      filter = "(costs.provider = 'aws' AND costs.region = 'us-east-1') OR (costs.provider = 'gcp' AND costs.region = 'us-central1')"
    }
  ]
}
