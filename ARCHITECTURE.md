# Architecture: terraform-provider-redfish

## Metadata

<!-- yaml-metadata-start -->
scope_paths: ["./"]
capture_git_sha: "a16a761cfb2b88bbeec310b505e046ea0a94aa5b"
status: "current"
auto_update: false
preview_before_apply: true
scaffold_version: "1.0"
<!-- yaml-metadata-end -->

---

## Purpose and Structure

Terraform provider for Dell iDRAC server management via Redfish API.
Implements 21 managed resources and 10 data sources
using HashiCorp's Terraform Plugin Framework, enabling
infrastructure-as-code management via REST API.

The provider is a standalone Go binary that communicates with Terraform
Core over gRPC (go-plugin protocol).

**SDK strategy:** Third-party vendor-neutral Redfish library (community-maintained). May lag behind iDRAC firmware features. Dell OEM extensions are implemented via `gofish/` local extension package.

---

## Components

| Component | Path | Responsibility |
|-----------|------|---------------|
| Entry point | `main.go` | `providerserver.Serve` — starts gRPC server |
| Provider | `redfish/provider.go` | Schema, Configure, resource/datasource registration |
| Resources | `redfish/*_resource.go` | CRUD lifecycle for 21 managed resources |
| Data sources | `redfish/*_datasource.go` | Read-only queries for 10 data sources |
| Dell OEM extensions | `gofish/` | Dell-specific Redfish extensions |
| Common utilities | `common/` | Shared helper functions |
| Mutex KV | `mutexkv/` | Mutex-based key-value for concurrent operations |
| Models | `redfish/models/` | Terraform state model structs |
| Test data | `test-data/` | Test fixtures and mock data |
| Scripts | `scripts/` | Helper scripts |
| Examples | `examples/` | HCL configurations for resources and data sources |
| Docs | `docs/` | Generated provider documentation |

---

## Key Behaviors

### Authentication

**GIVEN** a user configures the provider with endpoint, username,
and password (via HCL block or environment variables)
**WHEN** `Configure()` runs
**THEN** (1) env vars `REDFISH_ENDPOINT`, `REDFISH_USERNAME`,
`REDFISH_PASSWORD`, `REDFISH_INSECURE`, `REDFISH_TIMEOUT`
override HCL values, (2) SDK client is initialized, (3) authentication
is validated before any resource operations proceed

### Resource CRUD Lifecycle

**GIVEN** a resource definition in HCL
**WHEN** `terraform apply` runs
**THEN** the resource's `Create()` reads the plan into a model struct,
calls the SDK/client to create the resource, maps the API response
back to Terraform state, and sets `resp.State`

### Drift Detection

**GIVEN** a resource exists in Terraform state
**WHEN** `terraform plan` or `terraform refresh` runs
**THEN** `Read()` calls the SDK/client to fetch current state,
compares it with stored state, and updates the state if drifted

### Import

**GIVEN** a resource exists on the hardware but not in Terraform state
**WHEN** `terraform import` runs
**THEN** `ImportState()` fetches the resource by ID and populates state

---

## Interfaces

### Provider Configuration Schema

| Attribute | Type | Env Var | Description |
|-----------|------|---------|-------------|
| `endpoint` | string | `REDFISH_ENDPOINT` | iDRAC IP with `/redfish/v1` path |
| `username` | string | `REDFISH_USERNAME` | API username |
| `password` | string (sensitive) | `REDFISH_PASSWORD` | API password |
| `insecure` | bool | `REDFISH_INSECURE` | Skip TLS verification (lab only) |
| `timeout` | int64 | `REDFISH_TIMEOUT` | Request timeout in seconds |

---

## Dependencies

| Depends On | For |
|------------|-----|
| `github.com/stmcginnis/gofish` v0.20.0 (third-party) | Platform API SDK/client |
| `hashicorp/terraform-plugin-framework` v1.19.0 | Core provider interfaces |
| `hashicorp/terraform-plugin-framework-validators` | Attribute validation |
| `hashicorp/terraform-plugin-log` | Structured logging |
| `hashicorp/terraform-plugin-testing` | Acceptance test harness |
| `bytedance/mockey` | Unit test function-level mocking |
| `stretchr/testify` | Test assertions |

---

## Known Constraints

1. **Terraform Plugin Framework only** — no SDK v2 code.
2. **CGO_ENABLED=0** — static binaries for all platforms.
3. **Sensitive attributes marked** — credentials never in plan output.
4. **ImportState required** — all resources support `terraform import`.
5. **Environment variable fallback** — all credentials support env vars.
6. **Acceptance tests gated** — never run without `TF_ACC=1`.
7. **Endpoint format** — iDRAC IP with `/redfish/v1` path.
8. **Dell OEM extensions** — `gofish/` extends the
   third-party library with Dell-specific Redfish operations.

---

## Change History

| Date | Feature | What Changed | Author |
|------|---------|-------------|--------|
| 2026-06-10 | Initial architecture | Provider-specific architecture extracted from generic multi-provider doc | architecture-agent |
