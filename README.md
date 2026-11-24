# updater-npm

[![Latest Release](https://img.shields.io/github/v/release/SemRels/updater-npm?label=version&color=blue)](https://github.com/SemRels/updater-npm/releases/latest)
[![CI](https://github.com/SemRels/updater-npm/actions/workflows/ci.yml/badge.svg)](https://github.com/SemRels/updater-npm/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/SemRels/updater-npm)](LICENSE)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/SemRels/updater-npm/badge)](https://scorecard.dev/viewer/?uri=github.com/SemRels/updater-npm)
[![REUSE status](https://api.reuse.software/badge/github.com/SemRels/updater-npm)](https://api.reuse.software/info/github.com/SemRels/updater-npm)

Updates the package version in `package.json`.

This plugin is distributed as the standalone Go binary `semrel-plugin-updater-npm`. Semrel executes the binary as a subprocess, provides plugin configuration through `SEMREL_PLUGIN_*` environment variables, provides release context through `SEMREL_*` environment variables, reads standard output, and treats exit code `0` as success and any non-zero exit code as failure. Install the binary in `~/.semrel/plugins/` or anywhere on your `$PATH`.

## Installation

### Binary

```bash
go install github.com/SemRels/updater-npm/cmd/plugin@latest
```

### Docker

Pre-built, multi-platform images (linux/amd64, linux/arm64) are published to the GitHub Container Registry on every release:

```bash
docker pull ghcr.io/semrels/updater-npm:latest
```

Images are signed with [cosign](https://github.com/sigstore/cosign) and include a full SBOM attestation. Verify the signature:

```bash
cosign verify ghcr.io/semrels/updater-npm:latest \
  --certificate-identity-regexp 'https://github.com/SemRels/updater-npm/.github/workflows/container.yml.*' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```


## Configuration

```yaml
plugins:
  - uses: "@semrel/updater-npm"
    phase: pre-tag
    args:
      file: "package.json"
      update_lockfile: "true"
```

## Options

Every option can be set as an entry under `args` in `.semrel.yaml`. semrel passes each entry to the plugin as the environment variable shown in the table, so you can also set that variable directly (for example in your CI environment) instead of using `args`.

| Argument | Environment variable | Required | Description | Default |
| --- | --- | --- | --- | --- |
| `file` | `SEMREL_PLUGIN_FILE` | Optional | Path to the package manifest to update. | package.json |
| `update_lockfile` | `SEMREL_PLUGIN_UPDATE_LOCKFILE` | Optional | When `true`, runs `npm install --package-lock-only` after updating `package.json` to keep `package-lock.json` in sync. Missing `npm` or command failures are logged as warnings and do not fail the plugin. | false |

## `SEMREL_*` release context used

| Variable | Description |
| --- | --- |
| `SEMREL_VERSION` | Resolved release version for the current run. |
| `SEMREL_NEXT_VERSION` | Next version computed by semrel for the release. |
| `SEMREL_DRY_RUN` | Whether semrel is running in dry-run mode. |

## Example behavior

The plugin updates the `version` field in `package.json` to the next release version. When `SEMREL_PLUGIN_UPDATE_LOCKFILE=true`, it also runs `npm install --package-lock-only` in the manifest directory so `package-lock.json` stays aligned with the new version.

## License

Apache-2.0
