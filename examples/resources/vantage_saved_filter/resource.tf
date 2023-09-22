resource "vantage_saved_filter" "demo_filter" {
  title  = "Demo Saved Filter"
  filter = "(costs.provider = 'aws')"
}
