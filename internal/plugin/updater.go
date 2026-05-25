// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The updater-npm Authors

package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	semrelv1 "github.com/SemRels/updater-npm/internal/gen/v1"
)

const (
	packageJSONName     = "package.json"
	packageLockJSONName = "package-lock.json"
)

// Updater updates npm metadata files for a SemRel release.
type Updater struct{}

func NewUpdater() *Updater {
	return &Updater{}
}

func (u *Updater) UpdateFiles(_ context.Context, ctx *semrelv1.ReleaseContext) ([]string, error) {
	if ctx == nil {
		return nil, errors.New("release context is required")
	}
	if ctx.GetNextVersion() == nil {
		return nil, errors.New("next_version is required")
	}

	rootDir, err := resolveRootDir(ctx.GetConfig())
	if err != nil {
		return nil, err
	}

	version := formatVersion(ctx.GetNextVersion())
	rootPackagePath := filepath.Join(rootDir, packageJSONName)
	rootPackageContent, rootWorkspaces, rootChanged, err := preparePackageJSON(rootPackagePath, version)
	if err != nil {
		return nil, err
	}

	workspaceDirs := collectWorkspaceDirs(ctx.GetConfig(), rootWorkspaces)
	workspacePaths := make([]string, 0, len(workspaceDirs))
	seenWorkspacePaths := map[string]struct{}{}
	for _, workspaceDir := range workspaceDirs {
		workspacePath := workspacePackagePath(rootDir, workspaceDir)
		if sameFilePath(rootPackagePath, workspacePath) {
			continue
		}
		if _, seen := seenWorkspacePaths[workspacePath]; seen {
			continue
		}
		seenWorkspacePaths[workspacePath] = struct{}{}
		workspacePaths = append(workspacePaths, workspacePath)
	}
	sort.Strings(workspacePaths)

	type fileUpdate struct {
		path    string
		content []byte
	}

	pendingWrites := make([]fileUpdate, 0, 2+len(workspacePaths))
	updatedFiles := make([]string, 0, 2+len(workspacePaths))

	if rootChanged {
		pendingWrites = append(pendingWrites, fileUpdate{path: rootPackagePath, content: rootPackageContent})
		updatedFiles = append(updatedFiles, relativePath(rootDir, rootPackagePath))
	}

	updateLockFile, err := shouldUpdateLockFile(ctx.GetConfig())
	if err != nil {
		return nil, err
	}
	if updateLockFile {
		lockPath := filepath.Join(rootDir, packageLockJSONName)
		lockContent, lockChanged, err := preparePackageLockJSON(lockPath, version)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
		} else if lockChanged {
			pendingWrites = append(pendingWrites, fileUpdate{path: lockPath, content: lockContent})
			updatedFiles = append(updatedFiles, relativePath(rootDir, lockPath))
		}
	}

	for _, workspacePath := range workspacePaths {
		workspaceContent, _, workspaceChanged, err := preparePackageJSON(workspacePath, version)
		if err != nil {
			return nil, err
		}
		if !workspaceChanged {
			continue
		}
		pendingWrites = append(pendingWrites, fileUpdate{path: workspacePath, content: workspaceContent})
		updatedFiles = append(updatedFiles, relativePath(rootDir, workspacePath))
	}

	if ctx.GetDryRun() {
		return updatedFiles, nil
	}

	for _, update := range pendingWrites {
		if err := os.WriteFile(update.path, update.content, 0o644); err != nil {
			return nil, fmt.Errorf("write %s: %w", update.path, err)
		}
	}

	return updatedFiles, nil
}

func preparePackageJSON(path string, version string) ([]byte, []string, bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, false, fmt.Errorf("read %s: %w", path, err)
	}

	var document map[string]any
	if err := json.Unmarshal(content, &document); err != nil {
		return nil, nil, false, fmt.Errorf("parse %s: %w", path, err)
	}

	workspaces, err := extractWorkspaces(document["workspaces"])
	if err != nil {
		return nil, nil, false, fmt.Errorf("parse workspaces from %s: %w", path, err)
	}

	currentVersion, _ := document["version"].(string)
	changed := currentVersion != version
	if !changed {
		return nil, workspaces, false, nil
	}

	document["version"] = version
	formatted, err := marshalFormattedJSON(document)
	if err != nil {
		return nil, nil, false, fmt.Errorf("format %s: %w", path, err)
	}

	return formatted, workspaces, true, nil
}

func preparePackageLockJSON(path string, version string) ([]byte, bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, false, fmt.Errorf("read %s: %w", path, err)
	}

	var document map[string]any
	if err := json.Unmarshal(content, &document); err != nil {
		return nil, false, fmt.Errorf("parse %s: %w", path, err)
	}

	changed := false
	if currentVersion, _ := document["version"].(string); currentVersion != version {
		document["version"] = version
		changed = true
	}

	if packages, ok := document["packages"].(map[string]any); ok {
		if rootPackage, ok := packages[""].(map[string]any); ok {
			if currentVersion, _ := rootPackage["version"].(string); currentVersion != version {
				rootPackage["version"] = version
				changed = true
			}
		}
	}

	if !changed {
		return nil, false, nil
	}

	formatted, err := marshalFormattedJSON(document)
	if err != nil {
		return nil, false, fmt.Errorf("format %s: %w", path, err)
	}

	return formatted, true, nil
}

func resolveRootDir(config map[string]string) (string, error) {
	rootDir := "."
	if config != nil {
		if configured := strings.TrimSpace(config["root_dir"]); configured != "" {
			rootDir = configured
		}
	}

	resolved, err := filepath.Abs(rootDir)
	if err != nil {
		return "", fmt.Errorf("resolve root_dir %q: %w", rootDir, err)
	}

	return resolved, nil
}

func shouldUpdateLockFile(config map[string]string) (bool, error) {
	if config == nil {
		return true, nil
	}

	value := strings.TrimSpace(config["update_lock_file"])
	if value == "" {
		return true, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid update_lock_file value %q: %w", value, err)
	}

	return parsed, nil
}

func collectWorkspaceDirs(config map[string]string, packageJSONWorkspaces []string) []string {
	seen := map[string]struct{}{}
	workspaces := make([]string, 0, len(packageJSONWorkspaces))

	for _, workspace := range splitWorkspaceConfig(config) {
		cleaned := strings.TrimSpace(workspace)
		if cleaned == "" {
			continue
		}
		cleaned = filepath.Clean(cleaned)
		if _, ok := seen[cleaned]; ok {
			continue
		}
		seen[cleaned] = struct{}{}
		workspaces = append(workspaces, cleaned)
	}

	for _, workspace := range packageJSONWorkspaces {
		cleaned := strings.TrimSpace(workspace)
		if cleaned == "" {
			continue
		}
		cleaned = filepath.Clean(cleaned)
		if _, ok := seen[cleaned]; ok {
			continue
		}
		seen[cleaned] = struct{}{}
		workspaces = append(workspaces, cleaned)
	}

	return workspaces
}

func splitWorkspaceConfig(config map[string]string) []string {
	if config == nil {
		return nil
	}

	value := strings.TrimSpace(config["workspaces"])
	if value == "" {
		return nil
	}

	splitter := func(r rune) bool {
		switch r {
		case ',', ';', '\n', '\r':
			return true
		default:
			return false
		}
	}

	return strings.FieldsFunc(value, splitter)
}

func extractWorkspaces(raw any) ([]string, error) {
	switch value := raw.(type) {
	case nil:
		return nil, nil
	case []any:
		workspaces := make([]string, 0, len(value))
		for _, item := range value {
			workspace, ok := item.(string)
			if !ok {
				return nil, errors.New("workspaces array must contain only strings")
			}
			workspaces = append(workspaces, workspace)
		}
		return workspaces, nil
	case map[string]any:
		packages, ok := value["packages"]
		if !ok {
			return nil, errors.New("workspaces object must contain a packages array")
		}
		return extractWorkspaces(packages)
	default:
		return nil, errors.New("workspaces must be an array or an object with packages")
	}
}

func workspacePackagePath(rootDir string, workspace string) string {
	workspace = filepath.Clean(strings.TrimSpace(workspace))
	if strings.EqualFold(filepath.Base(workspace), packageJSONName) {
		if filepath.IsAbs(workspace) {
			return workspace
		}
		return filepath.Join(rootDir, workspace)
	}
	if filepath.IsAbs(workspace) {
		return filepath.Join(workspace, packageJSONName)
	}
	return filepath.Join(rootDir, workspace, packageJSONName)
}

func sameFilePath(left string, right string) bool {
	return filepath.Clean(left) == filepath.Clean(right)
}

func relativePath(rootDir string, target string) string {
	relative, err := filepath.Rel(rootDir, target)
	if err != nil {
		return filepath.Clean(target)
	}
	return relative
}

func marshalFormattedJSON(document any) ([]byte, error) {
	formatted, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return nil, err
	}
	formatted = append(formatted, '\n')
	if !json.Valid(formatted) {
		return nil, errors.New("generated JSON is invalid")
	}
	return formatted, nil
}

func formatVersion(version *semrelv1.SemanticVersion) string {
	result := fmt.Sprintf("%d.%d.%d", version.GetMajor(), version.GetMinor(), version.GetPatch())
	if preRelease := strings.TrimSpace(version.GetPreRelease()); preRelease != "" {
		result += "-" + preRelease
	}
	if buildMetadata := strings.TrimSpace(version.GetBuildMetadata()); buildMetadata != "" {
		result += "+" + buildMetadata
	}
	return result
}
