# Changelog

## Unreleased

### Features

* Add the `vantage_virtual_tag_config_value` resource for managing individual virtual tag config values and the `vantage_virtual_tag_config_values` data source for listing and finding values within a config.

### Breaking Changes

* `vantage_virtual_tag_config`: omitting `values` now preserves existing values on parent updates instead of clearing them, allowing for migration to `vantage_virtual_tag_config_value`.
  * Configurations that relied on omission to clear allocations must explicitly set `values = []` to retain that behavior.
  * Manage values either inline through `values` OR through standalone `vantage_virtual_tag_config_value` resources, never both for the same config. When using standalone resources, omit `values` from the parent.

* `vantage_team`: `user_emails`, `user_tokens`, and `workspace_tokens` changed from **List** to **Set** of String (schema version `0` → `1`).
  * Existing state is upgraded automatically via a v0→v1 state upgrader; no manual state surgery is required for type migration.
  * Configs that index these attributes (for example `user_emails[0]`) must be updated to treat them as unordered sets.
  * Duplicate elements in these attributes are no longer allowed.
  * Membership order from the API is no longer compared during plan/apply, which avoids false diffs when join order differs from config order.
