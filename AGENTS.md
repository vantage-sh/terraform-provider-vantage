# Terraform Provider Vantage - Agent Guidelines

This is a Terraform provider for [Vantage](https://vantage.sh), a cloud cost management platform. It uses the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework).

## Project Structure

```
vantage/
├── provider.go              # Provider definition and configuration
├── client.go                # Vantage API client wrapper
├── *_resource.go            # Resource implementations (handwritten)
├── *_resource_model.go      # Resource model helpers (handwritten)
├── *_resource_test.go       # Acceptance tests
├── *_data_source.go         # Data source implementations
├── resource_*/              # Generated schema code (DO NOT EDIT)
│   └── *_resource_gen.go
├── datasource_*/            # Generated schema code (DO NOT EDIT)
│   └── *_data_source_gen.go
```

## Key Conventions

### Resources Follow This Pattern

1. **Generated schema** in `resource_<name>/<name>_resource_gen.go` - auto-generated, do not edit
2. **Resource implementation** in `<name>_resource.go` - implements CRUD operations
3. **Model helpers** in `<name>_resource_model.go` - conversion between API and Terraform types
4. **Tests** in `<name>_resource_test.go` - acceptance tests

### Resource Interface Implementation

Resources must implement these interfaces:
```go
var (
    _ resource.Resource                = (*MyResource)(nil)
    _ resource.ResourceWithConfigure   = (*MyResource)(nil)
    _ resource.ResourceWithImportState = (*MyResource)(nil)  // if importable
)
```

### Token vs ID Pattern

All resources use a `token` field as the primary identifier, with `id` aliased to `token`:
```go
data.Token = types.StringValue(out.Payload.Token)
data.Id = types.StringValue(out.Payload.Token)
```

### Error Handling

Use the `handleError` helper from `client.go`:
```go
handleError("Create Resource Name", &resp.Diagnostics, err)
```

### API Client

The provider uses `vantage-go` client library with V1 and V2 API versions:
- Most resources use `r.client.V2` for API calls
- Authentication is handled via `r.client.Auth`

## Development

### Local Development Setup

1. Add dev override in `~/.terraformrc`:
```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/vantage-sh/vantage" = "<PATH TO GO BIN>"
  }
  direct {}
}
```

2. Build and install: `go install`
3. Set `VANTAGE_API_TOKEN` environment variable

### Running Tests

```bash
TF_ACC=1 make test
```

Tests require a valid `VANTAGE_API_TOKEN` for acceptance tests.

**IMPORTANT: Always run tests after making changes.** Tests catch bugs like schema mismatches between generated code and implementation.

### Code Generation

**Always regenerate after pulling changes or before submitting PRs:**

```bash
make generate    # Regenerate schemas from swagger
go generate ./...  # Regenerate documentation
```

The generation pipeline:
1. Downloads swagger from `https://api.vantage.sh/v2/swagger.json`
2. Converts to OpenAPI 3.0
3. Generates `spec.json` via `tfplugingen-openapi`
4. Generates Go code in `resource_*/` and `datasource_*/` directories

**Warning:** Generated files (`*_gen.go`) can become stale if not regenerated after swagger updates. This has caused bugs where the generated schema didn't match the API (e.g., a list field generated as a single object). Always regenerate and run tests to catch these issues.

### Regenerating Documentation

```bash
go generate ./...
```

## Testing Conventions

- Test function naming: `TestAccVantage<Resource>_<scenario>`
- Use `sdkacctest.RandStringFromCharSet` for random test data
- Use `acctest.PreCheck(t)` in `PreCheck` function
- Use `testAccProtoV6ProviderFactories` for provider factories
- Test configs are helper functions returning HCL strings
- **Write tests for list/array fields with multiple items** to catch schema bugs where lists are incorrectly generated as single objects
- Use `t.Skip()` for tests requiring specific features (e.g., MSP invoicing)

## Adding New Resources

1. Generate schema in `resource_<name>/` directory
2. Create `<name>_resource.go` with CRUD methods
3. Create `<name>_resource_model.go` if complex type conversions needed
4. Register in `provider.go` under `Resources()` function
5. Create `<name>_resource_test.go` with acceptance tests
6. Add example in `examples/resources/vantage_<name>/`

## Code Style

- Use `types.StringValue()`, `types.StringPointer()` for Terraform types
- Check `IsNull()` and `IsUnknown()` before accessing optional values
- Use `resp.Diagnostics.Append()` for error accumulation
- Use `PlanModifiers` like `stringplanmodifier.UseStateForUnknown()` for computed fields
