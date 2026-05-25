// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The updater-npm Authors

package plugin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	semrelv1 "github.com/SemRels/updater-npm/internal/gen/v1"
	"github.com/stretchr/testify/require"
)

func TestUpdaterUpdatesSinglePackageJSON(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeJSONFile(t, filepath.Join(rootDir, packageJSONName), map[string]any{
		"name":    "root-app",
		"version": "1.0.0",
	})

	updatedFiles, err := NewUpdater().UpdateFiles(context.Background(), releaseContext(rootDir, nil))
	require.NoError(t, err)
	require.Equal(t, []string{packageJSONName}, updatedFiles)

	packageJSON := readJSONFile(t, filepath.Join(rootDir, packageJSONName))
	require.Equal(t, "1.2.3", packageJSON["version"])
}

func TestUpdaterUpdatesPackageLockJSON(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeJSONFile(t, filepath.Join(rootDir, packageJSONName), map[string]any{
		"name":    "root-app",
		"version": "1.0.0",
	})
	writeJSONFile(t, filepath.Join(rootDir, packageLockJSONName), map[string]any{
		"name":    "root-app",
		"version": "1.0.0",
		"packages": map[string]any{
			"": map[string]any{
				"name":    "root-app",
				"version": "1.0.0",
			},
		},
	})

	updatedFiles, err := NewUpdater().UpdateFiles(context.Background(), releaseContext(rootDir, nil))
	require.NoError(t, err)
	require.Equal(t, []string{packageJSONName, packageLockJSONName}, updatedFiles)

	lockFile := readJSONFile(t, filepath.Join(rootDir, packageLockJSONName))
	require.Equal(t, "1.2.3", lockFile["version"])
	packages := lockFile["packages"].(map[string]any)
	rootPackage := packages[""].(map[string]any)
	require.Equal(t, "1.2.3", rootPackage["version"])
}

func TestUpdaterUpdatesWorkspacePackageJSONFiles(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeJSONFile(t, filepath.Join(rootDir, packageJSONName), map[string]any{
		"name":       "root-app",
		"version":    "1.0.0",
		"workspaces": []string{"packages/app-one"},
	})
	writeJSONFile(t, filepath.Join(rootDir, "packages", "app-one", packageJSONName), map[string]any{
		"name":    "app-one",
		"version": "1.0.0",
	})
	writeJSONFile(t, filepath.Join(rootDir, "packages", "app-two", packageJSONName), map[string]any{
		"name":    "app-two",
		"version": "1.0.0",
	})

	ctx := releaseContext(rootDir, map[string]string{"workspaces": "packages/app-two"})
	updatedFiles, err := NewUpdater().UpdateFiles(context.Background(), ctx)
	require.NoError(t, err)
	require.Equal(t, []string{
		packageJSONName,
		filepath.Join("packages", "app-one", packageJSONName),
		filepath.Join("packages", "app-two", packageJSONName),
	}, updatedFiles)

	workspaceOne := readJSONFile(t, filepath.Join(rootDir, "packages", "app-one", packageJSONName))
	workspaceTwo := readJSONFile(t, filepath.Join(rootDir, "packages", "app-two", packageJSONName))
	require.Equal(t, "1.2.3", workspaceOne["version"])
	require.Equal(t, "1.2.3", workspaceTwo["version"])
}

func TestUpdaterReturnsErrorForInvalidJSON(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(rootDir, packageJSONName), []byte("{not valid json}"), 0o644))

	updatedFiles, err := NewUpdater().UpdateFiles(context.Background(), releaseContext(rootDir, nil))
	require.Error(t, err)
	require.Nil(t, updatedFiles)
	require.ErrorContains(t, err, packageJSONName)
}

func TestUpdaterReturnsGracefulErrorForMissingWorkspaceFile(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeJSONFile(t, filepath.Join(rootDir, packageJSONName), map[string]any{
		"name":    "root-app",
		"version": "1.0.0",
	})

	ctx := releaseContext(rootDir, map[string]string{"workspaces": "packages/missing-app"})
	updatedFiles, err := NewUpdater().UpdateFiles(context.Background(), ctx)
	require.Error(t, err)
	require.Nil(t, updatedFiles)
	require.ErrorContains(t, err, filepath.Join("packages", "missing-app", packageJSONName))
}

func TestUpdaterRespectsDryRun(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeJSONFile(t, filepath.Join(rootDir, packageJSONName), map[string]any{
		"name":    "root-app",
		"version": "1.0.0",
	})

	ctx := releaseContext(rootDir, nil)
	ctx.DryRun = true

	updatedFiles, err := NewUpdater().UpdateFiles(context.Background(), ctx)
	require.NoError(t, err)
	require.Equal(t, []string{packageJSONName}, updatedFiles)

	packageJSON := readJSONFile(t, filepath.Join(rootDir, packageJSONName))
	require.Equal(t, "1.0.0", packageJSON["version"])
}

func TestUpdaterFormatsPreReleaseAndSkipsLockFileWhenDisabled(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	writeJSONFile(t, filepath.Join(rootDir, packageJSONName), map[string]any{
		"name":    "root-app",
		"version": "1.0.0",
	})
	writeJSONFile(t, filepath.Join(rootDir, packageLockJSONName), map[string]any{
		"name":    "root-app",
		"version": "1.0.0",
	})

	ctx := releaseContext(rootDir, map[string]string{"update_lock_file": "false"})
	ctx.NextVersion = &semrelv1.SemanticVersion{
		Major:         2,
		Minor:         0,
		Patch:         0,
		PreRelease:    "rc.1",
		BuildMetadata: "build.5",
	}

	updatedFiles, err := NewUpdater().UpdateFiles(context.Background(), ctx)
	require.NoError(t, err)
	require.Equal(t, []string{packageJSONName}, updatedFiles)

	packageJSON := readJSONFile(t, filepath.Join(rootDir, packageJSONName))
	require.Equal(t, "2.0.0-rc.1+build.5", packageJSON["version"])

	lockFile := readJSONFile(t, filepath.Join(rootDir, packageLockJSONName))
	require.Equal(t, "1.0.0", lockFile["version"])
}

func releaseContext(rootDir string, config map[string]string) *semrelv1.ReleaseContext {
	if config == nil {
		config = map[string]string{}
	}
	config["root_dir"] = rootDir

	return &semrelv1.ReleaseContext{
		Config: config,
		NextVersion: &semrelv1.SemanticVersion{
			Major: 1,
			Minor: 2,
			Patch: 3,
		},
	}
}

func writeJSONFile(t *testing.T, path string, content any) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	formatted, err := json.MarshalIndent(content, "", "  ")
	require.NoError(t, err)
	formatted = append(formatted, '\n')
	require.NoError(t, os.WriteFile(path, formatted, 0o644))
}

func readJSONFile(t *testing.T, path string) map[string]any {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(content, &decoded))
	return decoded
}
