# Dell Terraform Providers — Architecture

## Metadata

<!-- yaml-metadata-start -->
scope_paths: ["./"]
capture_git_sha: "57f6e2aa6f2e6b513ec62b77b088c7eb8535e3e4"
status: "current"
auto_update: false
preview_before_apply: true
scaffold_version: "1.0"
<!-- yaml-metadata-end -->

---

## Overview

This repository contains seven Terraform providers that bring infrastructure-
as-code to Dell storage arrays and server management platforms. Each provider
is a standalone Go binary communicating with Terraform Core over gRPC,
exposing Dell hardware resources through HashiCorp's Terraform Plugin
Framework.

---

## Repository Layout

```
terraform-providers/
├── terraform-provider-powerstore/   # PowerStore block storage
├── terraform-provider-powerflex/    # PowerFlex (VxFlex OS) software-defined
├── terraform-provider-powerscale/   # PowerScale / Isilon scale-out NAS
├── terraform-provider-powermax/     # PowerMax enterprise arrays
├── terraform-provider-objectscale/  # ObjectScale S3-compatible object storage
├── terraform-provider-redfish/      # iDRAC server management (Redfish API)
└── terraform-provider-ome/          # OpenManage Enterprise fleet management
```

Each provider directory follows identical structure:

```
terraform-provider-<name>/
├── main.go                    # Entry point
├── go.mod / go.sum            # Go module definition
├── Makefile                   # Build, test, install targets
├── .goreleaser.yaml           # Cross-platform release config
├── <name>/                    # Provider implementation
│   ├── provider.go            # Provider schema and Configure()
│   ├── resource_*.go          # Managed resources (CRUD)
│   ├── datasource_*.go        # Data sources (read-only)
│   └── *_test.go              # Unit tests
├── client/                    # SDK client wrapper (if applicable)
├── models/                    # Terraform ↔ SDK type mappings
├── docs/                      # Generated documentation
├── examples/                  # Example HCL configurations
└── about/                     # Metadata files
```

---

## System Context

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         User Workflow                                    │
│                                                                         │
│   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                │
│   │ main.tf     │    │ main.tf     │    │ main.tf     │                │
│   │ (Storage)   │    │ (Servers)   │    │ (Fleet)     │                │
│   └──────┬──────┘    └──────┬──────┘    └──────┬──────┘                │
│          │                  │                  │                        │
│          └──────────────────┼──────────────────┘                        │
│                             │                                           │
│                      ┌──────▼──────┐                                    │
│                      │  Terraform  │                                    │
│                      │    Core     │                                    │
│                      └──────┬──────┘                                    │
│                             │ gRPC (go-plugin)                          │
└─────────────────────────────┼───────────────────────────────────────────┘
                              │
┌─────────────────────────────▼───────────────────────────────────────────┐
│                    Dell Terraform Providers                              │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ Storage Providers                                                │   │
│  │  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────┐        │   │
│  │  │powerstore │ │powerflex  │ │powerscale │ │powermax   │        │   │
│  │  │           │ │           │ │           │ │           │        │   │
│  │  │gopowersto.│ │goscaleio  │ │vendored   │ │vendored   │        │   │
│  │  └─────┬─────┘ └─────┬─────┘ └─────┬─────┘ └─────┬─────┘        │   │
│  └────────┼─────────────┼─────────────┼─────────────┼──────────────┘   │
│           │             │             │             │                   │
│  ┌────────┼─────────────┼─────────────┼─────────────┼──────────────┐   │
│  │ Infrastructure Providers           │             │               │   │
│  │  ┌───────────┐ ┌───────────┐ ┌─────┴─────┐       │               │   │
│  │  │objectscale│ │redfish    │ │ome        │       │               │   │
│  │  │           │ │           │ │           │       │               │   │
│  │  │internal   │ │gofish     │ │internal   │       │               │   │
│  │  └─────┬─────┘ └─────┬─────┘ └─────┬─────┘       │               │   │
│  └────────┼─────────────┼─────────────┼─────────────┼──────────────┘   │
└───────────┼─────────────┼─────────────┼─────────────┼───────────────────┘
            │             │             │             │
            ▼             ▼             ▼             ▼
┌───────────────────────────────────────────────────────────────────────┐
│                         Dell Hardware                                  │
│                                                                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │
│  │ PowerStore  │  │ PowerFlex   │  │ PowerScale  │  │ PowerMax    │  │
│  │ REST API    │  │ Gateway API │  │ Platform API│  │ Unisphere   │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘  │
│                                                                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                    │
│  │ ObjectScale │  │ iDRAC       │  │ OpenManage  │                    │
│  │ IAM/S3 API  │  │ Redfish API │  │ Enterprise  │                    │
│  └─────────────┘  └─────────────┘  └─────────────┘                    │
└───────────────────────────────────────────────────────────────────────┘
```

---

## Provider Inventory

| Provider | Go | SDK | Version | Strategy |
|----------|-----|-----|---------|----------|
| powerstore | 1.25.0 | `github.com/dell/gopowerstore` | v1.18.0 | Public |
| powerflex | 1.24 | `github.com/dell/goscaleio` | v1.19.0 | Public |
| powerscale | 1.25.4 | `dell/powerscale-go-client` | local | Vendored |
| powermax | 1.25.8 | `dell/powermax-go-client` | local | Vendored |
| objectscale | 1.25.4 | Internal client | — | None |
| redfish | 1.25.8 | `github.com/stmcginnis/gofish` | v0.20.0 | Third-party |
| ome | 1.25.8 | Internal client | — | None |

---

## Component Architecture

### Provider Structure

Every provider implements the Terraform Plugin Framework interfaces:

```go
type Provider struct {
    client  *client.Client    // SDK client instance
    version string            // Provider version
}

// Schema — defines provider configuration (endpoint, credentials)
func (p *Provider) Schema(ctx, req, resp)

// Configure — initializes SDK client, validates auth
func (p *Provider) Configure(ctx, req, resp)

// Resources — returns list of managed resource constructors
func (p *Provider) Resources(ctx) []func() resource.Resource

// DataSources — returns list of data source constructors
func (p *Provider) DataSources(ctx) []func() datasource.DataSource
```

### Resource Lifecycle

```
┌─────────────────────────────────────────────────────────────────────┐
│                    Terraform Resource Lifecycle                      │
│                                                                     │
│  terraform plan          terraform apply         terraform destroy  │
│        │                       │                       │            │
│        ▼                       ▼                       ▼            │
│  ┌──────────┐            ┌──────────┐            ┌──────────┐      │
│  │  Read()  │            │ Create() │            │ Delete() │      │
│  │          │            │ Update() │            │          │      │
│  └────┬─────┘            └────┬─────┘            └────┬─────┘      │
│       │                       │                       │            │
│       ▼                       ▼                       ▼            │
│  ┌──────────────────────────────────────────────────────────┐      │
│  │                      SDK Layer                            │      │
│  │  GET /resource        POST /resource        DELETE /resource   │
│  │                       PUT /resource                        │      │
│  └──────────────────────────────────────────────────────────┘      │
│       │                       │                       │            │
│       ▼                       ▼                       ▼            │
│  ┌──────────────────────────────────────────────────────────┐      │
│  │                    Dell REST API                          │      │
│  └──────────────────────────────────────────────────────────┘      │
└─────────────────────────────────────────────────────────────────────┘
```

### SDK Strategies

**Public SDKs** — Versioned Go modules on GitHub:
```go
require github.com/dell/gopowerstore v1.18.0
```

**Vendored SDKs** — Local directory with replace directive:
```go
require dell/powerscale-go-client v0.0.0
replace dell/powerscale-go-client => ./powerscale-go-client
```

**Internal Clients** — REST calls implemented in provider code (no SDK).

**Third-Party** — Community library (gofish for Redfish).

---

## Data Flow

### Create Resource

```
HCL Config
    │
    ▼
Terraform Core (parse, validate, plan)
    │
    ▼
Provider.Create(ctx, req, resp)
    │
    ├─► Read plan into Go struct (types.String → string)
    │
    ├─► SDK.CreateResource(params)
    │       │
    │       ▼
    │   POST /api/rest/resource
    │       │
    │       ▼
    │   Dell Array (create resource, return ID)
    │
    ├─► Map response to Terraform state
    │
    └─► resp.State.Set(ctx, model)
            │
            ▼
        State file updated
```

### Read Resource (Drift Detection)

```
State file
    │
    ▼
Terraform Core (refresh)
    │
    ▼
Provider.Read(ctx, req, resp)
    │
    ├─► Read state into Go struct
    │
    ├─► SDK.GetResource(id)
    │       │
    │       ▼
    │   GET /api/rest/resource/{id}
    │       │
    │       ▼
    │   Dell Array (return current state)
    │
    ├─► Compare API response with state
    │
    └─► resp.State.Set(ctx, model)  // Updates if drifted
```

---

## Testing Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Test Pyramid                                 │
│                                                                     │
│                        ┌─────────────┐                              │
│                        │ Acceptance  │  make testacc                │
│                        │   Tests     │  (live hardware)             │
│                        └──────┬──────┘                              │
│                               │                                     │
│                    ┌──────────▼──────────┐                          │
│                    │    Unit Tests       │  make test               │
│                    │   (mockey mocks)    │  (no hardware)           │
│                    └──────────┬──────────┘                          │
│                               │                                     │
│              ┌────────────────▼────────────────┐                    │
│              │      Static Analysis            │  make check        │
│              │  (gofmt, golangci-lint, go vet) │  make gosec        │
│              └─────────────────────────────────┘                    │
└─────────────────────────────────────────────────────────────────────┘
```

**Unit Tests** — Use `bytedance/mockey` for runtime function patching:
```go
mockey.Mock((*client.Client).CreateVolume).Return(&Volume{ID: "123"}, nil).Build()
```

**Acceptance Tests** — Use `terraform-plugin-testing` against live hardware:
```go
resource.Test(t, resource.TestCase{
    Steps: []resource.TestStep{{Config: testConfig, Check: checkFunc}},
})
```

---

## Build & Release

### Makefile Targets

| Target | Purpose |
|--------|---------|
| `make build` | Compile provider binary |
| `make install` | Install to `~/.terraform.d/plugins/` |
| `make test` | Run unit tests |
| `make testacc` | Run acceptance tests (requires hardware) |
| `make check` | Format, lint, vet |
| `make gosec` | Security scan |
| `make cover` | Generate coverage report |
| `make generate` | Generate documentation |

### GoReleaser

All providers use identical release configuration:

- **CGO_ENABLED=0** — Static binaries, no C dependencies
- **Platforms:** freebsd, windows, linux, darwin
- **Architectures:** amd64, 386, arm, arm64
- **Output:** `terraform-provider-<name>_v<version>_<os>_<arch>.zip`

---

## Security Model

| Aspect | Implementation |
|--------|----------------|
| Credential injection | HCL provider block or environment variables |
| Sensitive attributes | `Sensitive: true` in schema (redacted from output) |
| Transport security | HTTPS with TLS verification (default) |
| Insecure mode | `insecure = true` disables TLS verification (lab only) |
| State file | Contains secrets — use encrypted remote backends |
| Binary integrity | Registry binaries signed with HashiCorp GPG key |

---

## Dependencies

### Common (all providers)

| Package | Purpose |
|---------|---------|
| `hashicorp/terraform-plugin-framework` | Core provider interfaces |
| `hashicorp/terraform-plugin-framework-validators` | Attribute validation |
| `hashicorp/terraform-plugin-go` | Low-level protocol types |
| `hashicorp/terraform-plugin-log` | Structured logging |
| `hashicorp/terraform-plugin-testing` | Acceptance test harness |
| `bytedance/mockey` | Unit test mocking |
| `stretchr/testify` | Test assertions |

### Provider-Specific

| Provider | SDK Dependency |
|----------|----------------|
| powerstore | `github.com/dell/gopowerstore` |
| powerflex | `github.com/dell/goscaleio` |
| powerscale | `./powerscale-go-client` (vendored) |
| powermax | `./powermax-go-client-100` (vendored) |
| redfish | `github.com/stmcginnis/gofish` |

---

## Constraints

1. **All providers use Terraform Plugin Framework** — No new SDK v2 code.

2. **CGO disabled** — Static binaries for all platforms.

3. **Sensitive attributes marked** — Credentials never in plan output.

4. **ImportState required** — All resources support `terraform import`.

5. **Environment variable fallback** — All credentials support env vars.

6. **Acceptance tests gated** — Never run without `TF_ACC=1`.

7. **Vendored SDKs co-versioned** — SDK and provider release together.

---

## Navigation

| I need to... | Go to |
|--------------|-------|
| Understand a specific provider | `terraform-provider-<name>/` |
| See provider configuration | `<provider>/<name>/provider.go` |
| See resource implementation | `<provider>/<name>/resource_*.go` |
| See data source implementation | `<provider>/<name>/datasource_*.go` |
| Run tests | `make test` or `make testacc` |
| Build locally | `make install` |
| See examples | `<provider>/examples/` |
| Read generated docs | `<provider>/docs/` |
