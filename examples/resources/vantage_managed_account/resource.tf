terraform {
  required_providers {
    vantage = {
      source = "vantage-sh/vantage"
    }
  }
}

resource "vantage_managed_account" "terraform-managed-account" {
  contact_email = "support+terraform@vantage.sh"
  name          = "Terraform managed account"
  access_credential_tokens = [
    "accss_crdntl_145aa8924bdc55a9"
  ]
  billing_rule_tokens = [
    "bllng_rule_bc95e52f2af7bac6",
  ]
}
