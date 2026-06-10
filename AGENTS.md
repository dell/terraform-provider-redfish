# AGENTS.md - Dell Terraform Provider for Redfish

## Project Overview

This is the Terraform provider for Dell iDRAC server management via the Redfish API. It implements resources using HashiCorp's Terraform Plugin Framework, enabling infrastructure-as-code management of Dell server hardware through the industry-standard Redfish interface.

- **Language:** Go 1.25
- **Module path:** `terraform-provider-redfish`
- **Terraform Plugin Framework:** v1.19.0
- **SDK:** `github.com/stmcginnis/gofish` v0.20.0 (third-party, vendor-neutral Redfish library)
- **Registry address:** `registry.terraform.io/dell/redfish`
- **License:** Mozilla Public License 2.0

## Architecture

The provider follows the standard Terraform Plugin Framework architecture. It runs as a gRPC server that Terraform Core communicates with to manage server hardware resources via Redfish.

### Provider Configuration

The provider authenticates to iDRAC endpoints using endpoint (with `/redfish/v1` path), username, and password. Configuration can be supplied via HCL provider block or environment variables (`REDFISH_ENDPOINT`, `REDFISH_USERNAME`, `REDFISH_PASSWORD`, `REDFISH_INSECURE`, `REDFISH_TIMEOUT`).

### SDK Strategy

Uses `gofish` — a **third-party, vendor-neutral** Redfish library (`github.com/stmcginnis/gofish`). It is community-maintained and not Dell-specific. Dell-specific Redfish OEM extensions are handled in the `gofish/dell/` package within the provider repo. The library may lag behind iDRAC firmware features.

### Resources and Data Sources

The provider exposes approximately 21 resources covering Redfish entities such as BIOS settings, boot order, firmware updates, iDRAC attributes, network adapters, power management, storage volumes, user accounts, virtual media, and certificates.

## Directory Structure

```
main.go                           Entry point (providerserver.Serve)
redfish/
  provider/
    provider.go                   Provider configuration, resource/datasource registration
    *_resource.go                 Resource implementations
    *_test.go                     Unit and acceptance tests
  helper/                         Shared helper functions
  models/                         Terraform state model structs
gofish/
  dell/                           Dell-specific OEM Redfish extensions
common/                           Shared utilities
mutexkv/                          Mutex key-value store for serial access
scripts/                          Utility scripts
test-data/                        Test fixture data
examples/                         Example HCL configurations
docs/                             Generated documentation
templates/                        Documentation templates
about/                            Provider metadata
```

## Build Commands

The provider uses `GNUmakefile` instead of `Makefile`.

| Command | Description |
|---------|-------------|
| `make build` | Compile the provider binary |
| `make install` | Build and install to `~/.terraform.d/plugins/` |
| `make test` | Run unit tests |
| `make testacc` | Run acceptance tests (`TF_ACC=1`, requires live hardware) |
| `make check` | Run `gofmt`, `golangci-lint`, `go vet` |
| `make gosec` | Run security scan with `gosec` |

## Testing

### Unit Tests (mockey)

- Test files follow `*_test.go` convention in `redfish/provider/`.
- Frameworks: `github.com/bytedance/mockey` (function-level mocking).
- Run with `make test`.
- No hardware required.

### Acceptance Tests (terraform-plugin-testing)

- **Requires live iDRAC hardware** with credentials set via environment variables.
- Creates real resources — clean up after failures.
- Run with `make testacc`.

### Running Tests

```bash
# Unit tests (no hardware)
make test

# Acceptance tests (requires live iDRAC)
export REDFISH_ENDPOINT="https://idrac-ip/redfish/v1"
export REDFISH_USERNAME="admin"
export REDFISH_PASSWORD="secret"
export REDFISH_INSECURE="true"
make testacc
```

## Code Style and Conventions

### Code Organization Patterns

- **Resource pattern:** Each resource is implemented in `redfish/provider/` with resource, model, and helper files.
- **Dell OEM extensions:** Dell-specific Redfish operations in `gofish/dell/`.
- **Mutex KV:** `mutexkv/` provides serialized access for concurrent operations.

### File Header

All source files must include the Dell copyright and MPL 2.0 license header.

## Common Development Tasks

### Adding a New Resource

1. Create resource and model files in `redfish/provider/` and `redfish/models/`.
2. If the resource requires Dell OEM extensions, add them to `gofish/dell/`.
3. Register in `redfish/provider/provider.go`.
4. Add unit and acceptance tests.
5. Create example HCL in `examples/resources/redfish_<name>/`.

### Updating the Redfish SDK

```bash
go get github.com/stmcginnis/gofish@<version>
go mod tidy
```

## CI/CD

GitHub Actions workflows in `.github/workflows/`. GoReleaser configuration in `.goreleaser.yml` builds cross-platform binaries.

## Code Ownership

All files are owned by the maintainers defined in `.github/CODEOWNERS`.
