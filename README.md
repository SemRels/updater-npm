# updater-npm

NPM package updater plugin for SemRel.

Updates npm package versions and related metadata files during a SemRel release.

## Documentation

- SemRel docs (planned): <https://github.com/SemRels/semrel/tree/main/docs/plugins/updater-npm>
- Plugin template: <https://github.com/SemRels/plugin-template>
- Registry: <https://registry.semrel.io>

## Repository Layout

~~~text
cmd/plugin/              Plugin entry point
internal/plugin/         Business logic scaffold
internal/grpc/           gRPC transport scaffold
proto/v1                 Symlink to the SemRel protobuf contract
.github/workflows/       CI, release, and security automation
~~~

## Development

~~~bash
go build ./cmd/plugin
go test ./...
~~~

## Configuration Example

~~~yaml
plugins:
  - name: updater-npm
    type: updater
    config:
      package_file: package.json
      lock_file: package-lock.json
      registry: https://registry.npmjs.org
~~~

## Status

This repository is bootstrapped from SemRels/plugin-template and is ready for implementation.
