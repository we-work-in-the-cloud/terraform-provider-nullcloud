# Agent Guidelines

## Repository layout

This is the NullCloud Terraform provider. It wraps the NullCloud backend API and has a sibling repository that must be kept in sync:

- `terraform-provider-nullcloud/` — this repo (Terraform provider)
- `backend-nullcloud/` — the backend API server this provider talks to

Both live under the same parent directory. When making changes to one, always check whether the other needs a corresponding update.

## Internal structure

```
internal/
  client/                      # HTTP client — mirrors backend model types and owns CRUD calls
  provider/
    vpc_resource.go            # Resource: nullcloud_vpc
    subnet_resource.go         # Resource: nullcloud_subnet
    instance_resource.go       # Resource: nullcloud_instance
    loadbalancer_resource.go   # Resource: nullcloud_loadbalancer
    bucket_resource.go         # Resource: nullcloud_bucket
    database_resource.go       # Resource: nullcloud_database
    cluster_resource.go        # Resource: nullcloud_cluster
    vpc_data_source.go         # Data source: nullcloud_vpc
    subnet_data_source.go      # Data source: nullcloud_subnet
    instance_data_source.go    # Data source: nullcloud_instance
    loadbalancer_data_source.go # Data source: nullcloud_loadbalancer
    bucket_data_source.go      # Data source: nullcloud_bucket
    database_data_source.go    # Data source: nullcloud_database
    cluster_data_source.go     # Data source: nullcloud_cluster
    instance_action.go         # Action: nullcloud_instance_action (start/stop/restart)
    provider.go                # Registers all resources, data sources, and actions
docs/resources/   # Generated docs (do not hand-edit; regenerate with tfplugindocs)
examples/         # Example .tf files referenced by docs
```

## Keeping provider and backend in sync

The client types in `internal/client/client.go` are hand-maintained duplicates of the backend's `internal/model/model.go`. They must stay identical in field names and `json` tags.

| Provider file | Backend equivalent |
|---|---|
| `internal/client/client.go` (types) | `backend-nullcloud/internal/model/model.go` |
| `internal/provider/vpc_resource.go` | `backend-nullcloud/internal/api/vpc.go` |
| `internal/provider/subnet_resource.go` | `backend-nullcloud/internal/api/subnet.go` |
| `internal/provider/instance_resource.go` | `backend-nullcloud/internal/api/vsi.go` |
| `internal/provider/loadbalancer_resource.go` | `backend-nullcloud/internal/api/loadbalancer.go` |
| `internal/provider/bucket_resource.go` | `backend-nullcloud/internal/api/bucket.go` |
| `internal/provider/database_resource.go` | `backend-nullcloud/internal/api/database.go` |
| `internal/provider/cluster_resource.go` | `backend-nullcloud/internal/api/cluster.go` |

### Adding a field from the backend

1. Add the field to the client struct in `internal/client/client.go` with the same `json` tag as the backend.
2. Add the field to the `*Model` struct in the relevant `internal/provider/*_resource.go`.
3. Add the attribute to the resource `Schema`. Server-generated fields use `Computed: true` + `stringplanmodifier.UseStateForUnknown()`.
4. Set the field in both `Create` and `Read`.
5. Regenerate docs: `make docs` (or `go generate ./...`).

### Adding a new resource type

1. Add CRUD methods to `internal/client/client.go`.
2. Create `internal/provider/<name>_resource.go` following the pattern of existing resources.
3. Register the new resource in `internal/provider/provider.go` → `Resources()`.
4. Add an example under `examples/resources/nullcloud_<name>/` and regenerate docs.
5. Update `README.md` — add a row to the Resources table.
6. Update the internal structure diagram and sync table in this `AGENTS.md`.
7. Update `backend-nullcloud/AGENTS.md` — add the new provider files to the sync table.

### Adding a data source

1. Add a `Get<Name>` method to `internal/client/client.go` if not already present.
2. Create `internal/provider/<name>_data_source.go` following the pattern of `vpc_data_source.go`.
3. Register the new data source in `internal/provider/provider.go` → `DataSources()`.
4. Add an example under `examples/data-sources/nullcloud_<name>/` and regenerate docs.
5. Update `README.md` — add a row to the Data Sources table.
6. Update the internal structure diagram in this `AGENTS.md`.

### Adding an action

1. Create `internal/provider/<name>_action.go` following the pattern of `instance_action.go`.
2. Register the new action in `internal/provider/provider.go` → `Actions()`.
3. Add an example under `examples/actions/nullcloud_<name>/` and regenerate docs.
4. Update `README.md` — add a row to the Actions table.
5. Update the internal structure diagram in this `AGENTS.md`.

## Keeping docs current

Whenever you make a structural change, update these files before closing the task:

| Change | Files to update |
|---|---|
| New resource | `README.md` (Resources table), `AGENTS.md` internal structure + sync table, `backend-nullcloud/AGENTS.md` sync table |
| New data source | `README.md` (Data Sources table), `AGENTS.md` internal structure |
| New action | `README.md` (Actions table), `AGENTS.md` internal structure |
| New field on a type | No README change needed; regenerate docs with `make docs` |
| Renamed / removed resource, data source, or action | Same files as the corresponding "new" row — remove or rename the old entries |

The goal: a reader of either `README.md` or `AGENTS.md` should be able to understand the current state of the provider without reading the code.

## Resource schema conventions

- **User-supplied at create, immutable:** `Required: true` + `stringplanmodifier.RequiresReplace()`
- **User-supplied at create, optional with server default:** `Optional: true, Computed: true` + `RequiresReplace` + `UseStateForUnknown`
- **Server-generated, never changes:** `Computed: true` + `UseStateForUnknown`
- All resources are currently immutable (no in-place update); `Update` is a no-op.

## CRN convention

Every resource has a `crn` attribute (`Computed: true` + `UseStateForUnknown`). The backend generates it as:

```
crn:nullcloud:<resource-type>:<id>
```

Resource type tokens: `vpc`, `subnet`, `instance`, `loadbalancer`, `bucket`, `database`, `cluster`.

## Build and test

```sh
make test       # go test -v -cover ./...
make testacc    # TF_ACC=1 acceptance tests (requires a running backend)
make build      # cross-compiles for all platforms into dist/
make install    # installs into ~/.terraform.d/plugins for local testing
```

Always run `make test` in both this repo and `backend-nullcloud/` before considering a change complete.

## Docs

Docs under `docs/` are generated by [tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs) from the schema `Description` fields and the example files in `examples/`. Do not hand-edit `docs/` — regenerate instead.

`tfplugindocs` is pinned as a Go tool dependency in `tools/tools.go` (using the `tools.go` model) so no separate installation is needed:

```sh
make docs          # runs: cd tools && go generate ./...
```

The `go:generate` directive in `tools/tools.go` invokes:
```
go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-dir ..
```
