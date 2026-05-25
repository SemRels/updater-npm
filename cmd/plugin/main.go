// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The semrel Authors

package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	plugin "github.com/SemRels/updater-npm/internal/plugin"
)

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Getenv))
}

func run(stdout, stderr io.Writer, getenv func(string) string) int {
	version := getenv("SEMREL_VERSION")
	if version == "" {
		version = getenv("SEMREL_NEXT_VERSION")
	}
	if version == "" {
		fmt.Fprintln(stderr, "updater-npm: SEMREL_VERSION is required")
		return 1
	}
	version = strings.TrimPrefix(version, "v")

	file := getenv("SEMREL_PLUGIN_FILE")
	if file == "" {
		file = "package.json"
	}

	if getenv("SEMREL_DRY_RUN") == "true" {
		fmt.Fprintf(stdout, "updater-npm: [dry-run] would update %s to version %s\n", file, version)
		return 0
	}

	if err := plugin.NewUpdater().Update(file, version); err != nil {
		fmt.Fprintln(stderr, "updater-npm:", err)
		return 1
	}

	fmt.Fprintf(stdout, "updater-npm: updated %s to version %s\n", file, version)
	return 0
}
