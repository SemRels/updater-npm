// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The semrel Authors

// Package plugin updates package.json files in-place.
package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

var versionFieldPattern = regexp.MustCompile(`(?m)^(\s*"version"\s*:\s*)"[^"]*"(\s*,?\s*)$`)
var indentationPattern = regexp.MustCompile(`(?m)^(\s+)"[^"]+"\s*:`)

// Updater updates package.json version fields.
type Updater struct{}

// NewUpdater creates an updater.
func NewUpdater() *Updater {
	return &Updater{}
}

// Update rewrites the version field in package.json.
func (u *Updater) Update(path, version string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	if _, ok := decoded["version"]; !ok {
		return fmt.Errorf("version field not found in %s", path)
	}

	if !versionFieldPattern.Match(data) {
		return fmt.Errorf("version field not found in %s", path)
	}

	updated := versionFieldPattern.ReplaceAllString(string(data), `${1}"`+version+`"${2}`)
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// Indentation returns the indentation detected in the package.json file.
func Indentation(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	if match := indentationPattern.FindSubmatch(data); match != nil {
		return string(match[1]), nil
	}
	return "", nil
}
