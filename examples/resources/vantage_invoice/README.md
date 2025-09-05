# Invoice Resource Examples

This directory contains examples for using the `vantage_invoice` resource.

## Files

- `resource.tf` - Basic invoice creation example

## Usage

The invoice resource creates invoices for managed accounts. Invoices are **immutable** once created and cannot be updated or deleted through the API.

### Required Parameters

- `account_token` - Token of the managed account to invoice
- `billing_period_start` - Start date of billing period (YYYY-MM-DD format)
- `billing_period_end` - End date of billing period (YYYY-MM-DD format)

### Computed Attributes

- `token` - The unique token of the created invoice
- `invoice_number` - Sequential invoice number for the MSP account
- `total` - Total amount for the invoice period
- `status` - Current status of the invoice
- `account_name` - Name of the managed account
- `created_at` - When the invoice was created
- `updated_at` - When the invoice was last updated

## Important Notes

- Invoices cannot be updated after creation
- Invoices cannot be deleted - they are permanent records
- Removing an invoice from Terraform state will not delete it from Vantage
- You need a valid managed account token to create invoices
