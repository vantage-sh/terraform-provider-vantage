---
page_title: "vantage_aws_provider Resource - terraform-provider-vantage"
subcategory: "Vendor Integrations"
description: |-
  Manages an AWS Account Integration.
---

# vantage_aws_provider (Resource)

## Example Usage

```terraform
resource "vantage_aws_provider" "demo" {
  cross_account_arn = "arn:aws:iam::123456789012:role/TestRole"
  bucket_arn        = "arn:aws:s3:::test-bucket"
}
```

## Schema

### Required
- `cross_account_arn` (String) ARN for cross account access.

### Optional
- `bucket_arn` (String) Bucket ARN with CUR data.

### Read-Only
- `id` (Integer)