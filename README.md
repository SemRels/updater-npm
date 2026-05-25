# updater-npm

`updater-npm` is the SemRel FilesUpdater plugin for npm projects. It updates the root `package.json` version, optionally updates `package-lock.json`, and can update workspace `package.json` files during a release.

## What it updates

- `package.json` in the configured `root_dir`
- `package-lock.json` in the configured `root_dir` when `update_lock_file` is not set to `false`
- workspace `package.json` files from either:
  - plugin config key `workspaces`
  - the root `package.json` `workspaces` field

Workspace paths are treated as literal directories; glob expansion is not performed by the plugin.

## Configuration

| Key | Default | Description |
| --- | --- | --- |
| `root_dir` | `.` | Working directory that contains the root `package.json`. |
| `update_lock_file` | `true` | Whether to update `package-lock.json` if it exists. |
| `workspaces` | empty | Comma-separated workspace directories to update in addition to any workspaces declared in the root `package.json`. |

## Example

```yaml
plugins:
  - name: updater-npm
    type: updater
    config:
      root_dir: .
      update_lock_file: true
      workspaces: packages/app-one,packages/app-two
```

If the root `package.json` contains:

```json
{
  "name": "example-monorepo",
  "version": "1.4.2",
  "workspaces": ["packages/web"]
}
```

and SemRel chooses `1.5.0`, the plugin updates:

- `package.json`
- `package-lock.json` if present and enabled
- `packages/web/package.json`
- `packages/app-one/package.json`
- `packages/app-two/package.json`

## Dry runs

When `ctx.dry_run` is `true`, the plugin validates inputs and reports which files would change, but it does not write anything to disk.

## Error scenarios

The plugin returns an `error_message` in the `UpdateFilesResponse` for common failures, including:

- missing root `package.json`
- missing configured workspace `package.json`
- invalid JSON in `package.json` or `package-lock.json`
- invalid `update_lock_file` values
- missing `next_version` in the release context

## Transport

This plugin is served as a HashiCorp `go-plugin` gRPC plugin. stdout is reserved for the plugin handshake; all operational logging must go to stderr.

Default handshake settings:

- protocol version: `1`
- plugin map key: `files-updater`
- magic cookie key: `SEMREL_PLUGIN`
- magic cookie value: `updater-npm`

The magic cookie key and value can be overridden with these environment variables:

- `SEMREL_PLUGIN_MAGIC_COOKIE_KEY`
- `SEMREL_PLUGIN_MAGIC_COOKIE_VALUE`

## Development

```bash
go test ./...
go build ./cmd/plugin
```

## Repository layout

```text
cmd/plugin/              go-plugin entrypoint
internal/gen/v1/         generated protobuf and gRPC bindings
internal/grpc/           FilesUpdater gRPC adapter
internal/plugin/         npm file update logic and go-plugin wiring
proto/v1/                local copy of the SemRel protobuf contract
```
