# terraform-provider-vantage

# Use

For prebuilt modules utilizing this and other cloud providers to configure your Vantage account, see the [Vantage Maintained Integration Modules](https://github.com/vantage-sh/terraform-vantage-integrations).

### Development

To develop:

Create a `dev_overrides` for this provider:

In `~/.terraformrc`
```
provider_installation {

  dev_overrides {
      "registry.terraform.io/vantage-sh/vantage" = "<PATH TO GO BIN>"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

Then `go install` or `go build`, ensuring the binary produced is placed in the above override directory. This built binary will be used in place of the versioned binary available from the registry.

Once overridden, proceed to the `examples/cost_report` directory and do the usual:

You'll need `VANTAGE_API_TOKEN` exposed either via flags to the apply or via environment with a tool like `direnv`.

```
terraform apply
```

You should see a warning about the override:
```
╷
│ Warning: Provider development overrides are in effect
│
│ The following provider development overrides are set in the CLI configuration:
│  - vantage-sh/vantage in <PATH FROM ABOVE>
│
│ The behavior may therefore not match any released version of the provider and applying changes may cause the state to become
│ incompatible with published releases.
```

### Docs

Regenerate documentation with
```
go generate ./...
```
