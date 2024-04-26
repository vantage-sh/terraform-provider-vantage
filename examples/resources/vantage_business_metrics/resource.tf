data "vantage_business_metrics" "demo" {}
resource "vantage_business_metric" "demo_business_metric" {
  title = "Demo Business Metric"
  values = [
    {
      date = "2024-05-03"
      amount = 300
      label = "Demo Label 2"
    },
    {
      date = "2024-05-03"
      amount = 300
      label = "Demo Label"
    },
    {
      date = "2024-05-02"
      amount = 200
      label = "Demo Label"
    },
    {
      date = "2024-05-01"
      amount = 100
    },
  ]
}
