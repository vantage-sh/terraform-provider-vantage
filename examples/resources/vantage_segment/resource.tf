resource "vantage_segment" "demo_segment" {
  title = "Demo Segment"
  description = "This is a demo segment"
  priority = 100
  track_unallocated = false
  filter = "(costs.provider = 'aws' AND tags.name = NULL)"
}
