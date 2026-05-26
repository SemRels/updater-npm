# updater-npm

Updates the package version in `package.json`.

This plugin is distributed as the standalone Go binary `semrel-plugin-updater-npm`. Semrel executes the binary as a subprocess, provides plugin configuration through `SEMREL_PLUGIN_*` environment variables, provides release context through `SEMREL_*` environment variables, reads standard output, and treats exit code `0` as success and any non-zero exit code as failure. Install the binary in `~/.semrel/plugins/` or anywhere on your `$PATH`.

## Installation

```bash
go install github.com/SemRels/updater-npm/cmd/plugin@latest
```

## Configuration

```yaml
plugins:
  - name: updater-npm
    path: ~/.semrel/plugins/semrel-plugin-updater-npm
    env:
      SEMREL_PLUGIN_FILE: "package.json"
```

## `SEMREL_PLUGIN_*` variables

| Name | Required | Description | Default |
| --- | --- | --- | --- |
| `SEMREL_PLUGIN_FILE` | Optional | Path to the package manifest to update. | package.json |

## `SEMREL_*` release context used

| Variable | Description |
| --- | --- |
| `SEMREL_VERSION` | Resolved release version for the current run. |
| `SEMREL_NEXT_VERSION` | Next version computed by semrel for the release. |
| `SEMREL_DRY_RUN` | Whether semrel is running in dry-run mode. |

## Example behavior

The plugin updates the `version` field in `package.json` to the next release version.

## License

Apache-2.0
