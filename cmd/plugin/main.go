// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The semrel Authors

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
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

	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintf(stderr, "updater-npm: read %s: %v\n", file, err)
		return 1
	}

	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		fmt.Fprintf(stderr, "updater-npm: parse %s: %v\n", file, err)
		return 1
	}

	doc["version"] = version

	updated, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "updater-npm: marshal %s: %v\n", file, err)
		return 1
	}
	updated = append(updated, '\n')

	if err := os.WriteFile(file, updated, 0o644); err != nil {
		fmt.Fprintf(stderr, "updater-npm: write %s: %v\n", file, err)
		return 1
	}

	fmt.Fprintf(stdout, "updater-npm: updated %s to version %s\n", file, version)
	return 0
}
