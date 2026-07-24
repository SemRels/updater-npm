// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The semrel Authors

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	updaterplugin "github.com/SemRels/updater-npm/internal/plugin"
)

const pluginSchemaVersion = 1

type versionUpdater interface {
	Update(path, version string) error
}

type commandRunner func(context.Context, string, string, ...string) error

func main() {
	os.Exit(run(context.Background(), os.Stdout, os.Stderr, os.Getenv, updaterplugin.NewUpdater(), exec.LookPath, runCommand))
}

func run(
	ctx context.Context,
	stdout, stderr io.Writer,
	getenv func(string) string,
	updater versionUpdater,
	lookPath func(string) (string, error),
	runCommand commandRunner,
) int {
	_, _ = fmt.Fprintf(stderr, "plugin_schema_version=%d\n", pluginSchemaVersion)
	version := getenv("SEMREL_VERSION")
	if version == "" {
		version = getenv("SEMREL_NEXT_VERSION")
	}
	if version == "" {
		_, _ = fmt.Fprintln(stderr, "updater-npm: SEMREL_VERSION is required")
		return 1
	}
	version = strings.TrimPrefix(version, "v")

	file := getenv("SEMREL_PLUGIN_FILE")
	if file == "" {
		file = "package.json"
	}

	if getenv("SEMREL_DRY_RUN") == "true" {
		_, _ = fmt.Fprintf(stdout, "updater-npm: [dry-run] would update %s to version %s\n", file, version)
		return 0
	}

	if err := updater.Update(file, version); err != nil {
		_, _ = fmt.Fprintf(stderr, "updater-npm: %v\n", err)
		return 1
	}

	if getenv("SEMREL_PLUGIN_UPDATE_LOCKFILE") == "true" {
		updateLockfile(ctx, stderr, file, lookPath, runCommand)
	}

	_, _ = fmt.Fprintf(stdout, "updater-npm: updated %s to version %s\n", file, version)
	return 0
}

func updateLockfile(
	ctx context.Context,
	stderr io.Writer,
	file string,
	lookPath func(string) (string, error),
	runCommand commandRunner,
) {
	if _, err := lookPath("npm"); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			_, _ = fmt.Fprintf(stderr, "updater-npm: npm not found on PATH — skipping lockfile update\n")
			return
		}
		_, _ = fmt.Fprintf(stderr, "updater-npm: unable to find npm on PATH: %v — skipping lockfile update\n", err)
		return
	}

	if err := runCommand(ctx, filepath.Dir(file), "npm", "install", "--package-lock-only"); err != nil {
		_, _ = fmt.Fprintf(stderr, "updater-npm: npm install --package-lock-only failed: %v — skipping lockfile update\n", err)
	}
}

func runCommand(ctx context.Context, dir, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	if len(output) == 0 {
		return err
	}
	return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
}
