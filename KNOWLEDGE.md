# Dell Terraform Providers — Knowledge Base

<!-- yaml-metadata-start -->
scope_paths: ["./"]
capture_git_sha: "57f6e2aa6f2e6b513ec62b77b088c7eb8535e3e4"
status: "current"
auto_update: false
preview_before_apply: true
scaffold_version: "1.0"
# session_state: { is_complete: true }
<!-- yaml-metadata-end -->

Tribal knowledge, patterns, and gotchas for the Dell Terraform provider family.

---

## Repository Structure

```
terraform-providers/
├── terraform-provider-powerstore/   # PowerStore arrays
├── terraform-provider-powerflex/    # PowerFlex (VxFlex OS)
├── terraform-provider-powerscale/   # PowerScale / Isilon
├── terraform-provider-powermax/     # PowerMax (VMAX)
├── terraform-provider-objectscale/  # ObjectScale (S3-compatible)
├── terraform-provider-redfish/      # iDRAC Redfish API
└── terraform-provider-ome/          # OpenManage Enterprise
```

Each provider is a standalone Go module with its own `go.mod`, `Makefile`,
and release pipeline.

---

## Environment Variables

Every provider follows the same credential injection pattern. Replace
`<PROVIDER>` with the uppercase provider name (e.g., `POWERSTORE`, `POWERFLEX`).

| Variable | Purpose |
|----------|---------|
| `<PROVIDER>_ENDPOINT` | Array/server management IP or FQDN |
| `<PROVIDER>_USERNAME` | API username |
| `<PROVIDER>_PASSWORD` | API password |
| `<PROVIDER>_INSECURE` | `true` to skip TLS verification (lab only) |
| `<PROVIDER>_TIMEOUT` | Request timeout in seconds (default: 120) |

**Example (PowerStore):**
```bash
export POWERSTORE_ENDPOINT="https://10.0.0.1/api/rest"
export POWERSTORE_USERNAME="admin"
export POWERSTORE_PASSWORD="secret"
export POWERSTORE_INSECURE="true"
```

Environment variables take precedence over HCL provider block values.

---

## Makefile Targets

All providers share these standard targets:

| Target | Purpose | Hardware Required |
|--------|---------|-------------------|
| `make build` | Compile provider binary | No |
| `make install` | Install to `~/.terraform.d/plugins/` | No |
| `make test` | Run unit tests | No |
| `make testacc` | Run acceptance tests | **Yes** |
| `make check` | Format, lint, vet | No |
| `make gosec` | Security scan | No |
| `make cover` | Generate coverage report | No |
| `make generate` | Generate documentation | No |

**Never run `make testacc` without live hardware credentials.** Acceptance
tests perform real CRUD operations against arrays/servers.

---

## SDK Strategies

### Public SDKs (gopowerstore, goscaleio)

```go
// go.mod
require github.com/dell/gopowerstore v1.18.0
```

- Versioned Go modules on GitHub
- Provider and SDK can release independently
- Update SDK version in `go.mod`, run `go mod tidy`

### Vendored SDKs (powerscale-go-client, powermax-go-client)

```go
// go.mod
require dell/powerscale-go-client v0.0.0

replace dell/powerscale-go-client => ./powerscale-go-client
```

- SDK source lives inside the provider repo
- SDK and provider release together
- Changes to SDK require changes in the same repo

### Internal Clients (objectscale, ome)

- No external SDK dependency
- REST calls implemented directly in provider code
- Full control but more maintenance burden

### Third-Party (gofish for Redfish)

```go
require github.com/stmcginnis/gofish v0.20.0
```

- Vendor-neutral Redfish library (not Dell-specific)
- Community-maintained
- May lag behind iDRAC firmware features

---

## Provider Pattern

Every provider follows this structure:

```go
// provider.go
type Provider struct {
    client  *client.Client
    version string
}

func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest,
    resp *provider.ConfigureResponse) {
    // 1. Read config from HCL or environment variables
    // 2. Initialize SDK client
    // 3. Validate authentication (dummy API call)
    // 4. Set resp.ResourceData and resp.DataSourceData
}

func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        newVolumeResource,
        newSnapshotRuleResource,
        // ...
    }
}

func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
    return []func() datasource.DataSource{
        newVolumeDataSource,
        // ...
    }
}
```

---

## Resource Pattern

```go
// resource_volume.go
type volumeResource struct {
    client *client.Client
}

func (r *volumeResource) Create(ctx context.Context,
    req resource.CreateRequest, resp *resource.CreateResponse) {
    // 1. Read plan into model struct
    // 2. Call SDK to create resource
    // 3. Map response to state
    // 4. Set resp.State
}

func (r *volumeResource) Read(ctx context.Context,
    req resource.ReadRequest, resp *resource.ReadResponse) {
    // 1. Read state into model struct
    // 2. Call SDK to get current state
    // 3. Map response to state (detect drift)
    // 4. Set resp.State
}

func (r *volumeResource) Update(ctx context.Context,
    req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // 1. Read plan and state
    // 2. Compute diff
    // 3. Call SDK to update
    // 4. Set resp.State
}

func (r *volumeResource) Delete(ctx context.Context,
    req resource.DeleteRequest, resp *resource.DeleteResponse) {
    // 1. Read state
    // 2. Call SDK to delete
    // 3. State is automatically removed
}

func (r *volumeResource) ImportState(ctx context.Context,
    req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    // 1. Parse import ID
    // 2. Call SDK to get resource
    // 3. Populate state
}
```

---

## Testing Patterns

### Unit Tests (mockey)

```go
func TestVolumeResource_Create(t *testing.T) {
    // Mock SDK calls using bytedance/mockey
    mockey.PatchConvey("Create volume", t, func() {
        mockey.Mock((*client.Client).CreateVolume).Return(&Volume{ID: "123"}, nil).Build()
        // ... test logic
    })
}
```

- No hardware required
- Fast execution
- Run with `make test`

### Acceptance Tests (terraform-plugin-testing)

```go
func TestAccVolumeResource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccVolumeConfig("test-vol", 10),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("powerstore_volume.test", "name", "test-vol"),
                    resource.TestCheckResourceAttr("powerstore_volume.test", "size", "10737418240"),
                ),
            },
        },
    })
}
```

- **Requires live hardware**
- Creates real resources (clean up after!)
- Run with `TF_ACC=1 make testacc`

---

## GoReleaser Configuration

All providers use identical GoReleaser settings:

```yaml
builds:
  - env:
      - CGO_ENABLED=0  # Static binary, no C dependencies
    goos:
      - freebsd
      - windows
      - linux
      - darwin
    goarch:
      - amd64
      - '386'
      - arm
      - arm64
    ignore:
      - goos: darwin
        goarch: '386'  # No 32-bit macOS
```

---

## Common Gotchas

### 1. Endpoint URL format varies by provider

- **PowerStore:** Must end with `/api/rest`
- **PowerFlex:** Gateway URL (not MDM directly)
- **PowerScale:** Platform API port (typically 8080)
- **Redfish:** iDRAC IP with `/redfish/v1` path

### 2. Sensitive attributes must be marked

```go
"password": schema.StringAttribute{
    Sensitive: true,  // Required for credentials
}
```

Without this, passwords appear in `terraform plan` output and state files.

### 3. Vendored SDK updates

For powerscale/powermax, SDK changes require:
1. Edit files in `./powerscale-go-client/` or `./powermax-go-client-100/`
2. No `go mod tidy` needed (local replace directive)
3. Commit SDK and provider changes together

### 4. Acceptance test cleanup

If acceptance tests fail mid-run, resources may be left on the array.
Clean up manually before re-running tests.

### 5. State file contains secrets

Terraform state files contain full resource representations including
credentials. Always use encrypted remote backends (S3+KMS, Terraform Cloud)
in production.

---

## Version Compatibility

| Provider | Min Terraform | Plugin Framework |
|----------|---------------|------------------|
| powerstore | 1.4+ | v1.13.0 |
| powerflex | 1.4+ | v1.13.0 |
| powerscale | 1.4+ | v1.15.1 |
| powermax | 1.4+ | v1.19.0 |
| objectscale | 1.4+ | v1.15.1 |
| redfish | 1.4+ | v1.19.0 |
| ome | 1.4+ | v1.19.0 |

**Note:** powerstore still has partial `terraform-plugin-sdk/v2` dependency
alongside the framework (migration in progress).

---

## References

- [Terraform Plugin Framework Docs](https://developer.hashicorp.com/terraform/plugin/framework)
- [GoReleaser Docs](https://goreleaser.com/intro/)
- [bytedance/mockey](https://github.com/bytedance/mockey)
- [Dell Terraform Registry](https://registry.terraform.io/namespaces/dell)
