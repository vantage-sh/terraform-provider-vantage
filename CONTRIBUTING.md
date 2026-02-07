# Contributing to terraform-provider-vantage

Thank you for your interest in contributing to the Vantage Terraform Provider!

## Getting Started

1. Fork the repository
2. Clone your fork locally
3. Set up your development environment (see [README.md](README.md) for details)

## Development

### Local Development Setup

Create a `dev_overrides` configuration in `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/vantage-sh/vantage" = "<PATH TO GO BIN>"
  }
  direct {}
}
```

Build and install the provider:

```bash
go install
```

### Running Tests Locally

Run unit tests:

```bash
go test ./...
```

Run acceptance tests locally (optional — CI handles this for you):

> **Warning:** Acceptance tests create, modify, and delete real resources in your Vantage account. Not recommended.

```bash
export VANTAGE_API_TOKEN="your-api-token"
TF_ACC=1 make test
```

## CI for External Contributors

Our acceptance tests run against internal infrastructure, so they require maintainer approval for external contributors.

### What to Expect

1. **Open your PR** - You'll see a "pending approval" status on the `Terraform Provider Tests` check
2. **A maintainer will review** - Once reviewed, a maintainer will comment `/test` to trigger the acceptance tests
3. **Test results appear as a PR comment** - You'll see a summary of passed/failed tests with details

### Workflow

```
PR Opened → Unit Tests Run → Acceptance Tests Pending
                                    ↓
                         Maintainer reviews PR
                                    ↓
                         Maintainer comments /test
                                    ↓
                         Acceptance Tests Run
                                    ↓
                         Results posted to PR
```

### If Tests Fail

- Check the test output in the PR comment
- Fix any issues in your code
- Push new commits to your PR
- A maintainer can re-run tests by commenting `/test` again

## Submitting Changes

1. Create a branch for your changes
2. Make your changes with clear, descriptive commits
3. Ensure all unit tests pass locally
4. Open a pull request with a clear description of the changes
5. Wait for CI and maintainer review

## Code Style

- Follow Go conventions and idioms
- Run `go fmt` before committing
- Regenerate documentation if you change schemas: `go generate ./...`

## Questions?

If you have questions about contributing, feel free to open an issue for discussion.
