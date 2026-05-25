# updater-npm

npm updater plugin for Semantic Release.

Updates npm package metadata and versions during Semantic Release.

## Documentation

- Docs (coming soon): <https://github.com/SemRels/semrel/tree/main/docs/plugins/updater-npm>
- Template source: <https://github.com/SemRels/plugin-template>

## Repository Layout

`	ext
cmd/plugin/              Plugin entry point
internal/plugin/         Business logic scaffold
internal/grpc/           gRPC transport scaffold
proto/v1                 Symlink to the SemRel protobuf contract
.github/workflows/       CI, release, and security automation
`

## Development

`ash
go build ./cmd/plugin
go test ./...
`

## Configuration Example

`yaml
plugins:
  - name: updater-npm
    type: updater
    config:
      package_file: package.json
      lock_file: package-lock.json
      registry: https://registry.npmjs.org
`

## Status

This repository is bootstrapped from SemRels/plugin-template and is ready for implementation.