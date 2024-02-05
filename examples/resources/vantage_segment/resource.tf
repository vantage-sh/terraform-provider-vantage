resource "vantage_segment" "demo_segment" {
  title = "Demo Segment"
  description = "This is still a demo segment"
  priority = 50
  track_unallocated = false
  filter = "(costs.provider = 'aws' AND tags.name = NULL)"

  # either provide workspace_token or parent_segment_token
  workspace_token = "wrkspc_47c3254c790e9351"
  # parent_segment_token = "fltr_sgmt_1e866feb74f0b1b4"
}
