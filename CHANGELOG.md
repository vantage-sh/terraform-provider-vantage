# Changelog

## Unreleased

### Features

* Add the `vantage_virtual_tag_config_value` resource for managing individual virtual tag config values and the `vantage_virtual_tag_config_values` data source for listing and finding values within a config.
* `vantage_virtual_tag_config`: preserve existing values when `values` is omitted. Manage values inline OR through standalone value resources, never both; `values = []` explicitly removes all values.

### Breaking Changes

* `vantage_team`: `user_emails`, `user_tokens`, and `workspace_tokens` changed from **List** to **Set** of String (schema version `0` → `1`).
  * Existing state is upgraded automatically via a v0→v1 state upgrader; no manual state surgery is required for type migration.
  * Configs that index these attributes (for example `user_emails[0]`) must be updated to treat them as unordered sets.
  * Duplicate elements in these attributes are no longer allowed.
  * Membership order from the API is no longer compared during plan/apply, which avoids false diffs when join order differs from config order.
