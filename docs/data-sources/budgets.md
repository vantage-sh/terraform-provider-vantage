---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "vantage_budgets Data Source - terraform-provider-vantage"
subcategory: ""
description: |-
  
---

# vantage_budgets (Data Source)





<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `budgets` (Attributes List) (see [below for nested schema](#nestedatt--budgets))

<a id="nestedatt--budgets"></a>
### Nested Schema for `budgets`

Read-Only:

- `budget_alert_tokens` (List of String) The tokens of the BudgetAlerts associated with the Budget.
- `cost_report_token` (String) The token of the Report associated with the Budget.
- `created_at` (String) The date and time, in UTC, the Budget was created. ISO 8601 Formatted.
- `name` (String) The name of the Budget.
- `performance` (Attributes List) The historical performance of the Budget. (see [below for nested schema](#nestedatt--budgets--performance))
- `periods` (Attributes List) The budget periods associated with the Budget. (see [below for nested schema](#nestedatt--budgets--periods))
- `token` (String)
- `user_token` (String) The token for the User who created this Budget.
- `workspace_token` (String) The token for the Workspace the Budget is a part of.

<a id="nestedatt--budgets--performance"></a>
### Nested Schema for `budgets.performance`

Read-Only:

- `actual` (String) The date and time, in UTC, the Budget was created. ISO 8601 Formatted.
- `amount` (String) The amount of the Budget Period as a string to ensure precision.
- `date` (String) The date and time, in UTC, the Budget was created. ISO 8601 Formatted.


<a id="nestedatt--budgets--periods"></a>
### Nested Schema for `budgets.periods`

Read-Only:

- `amount` (String) The amount of the Budget Period as a string to ensure precision.
- `end_at` (String) The date and time, in UTC, the Budget was created. ISO 8601 Formatted.
- `start_at` (String) The date and time, in UTC, the Budget was created. ISO 8601 Formatted.

