package main

import (
	"context"
	"terraform-provider-vantage/vantage"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Generate the Terraform provider documentation using `tfplugindocs`:
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	providerserver.Serve(context.Background(), vantage.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/vantage-sh/vantage",
	})
}
