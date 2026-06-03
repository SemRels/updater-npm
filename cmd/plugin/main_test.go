// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The semrel Authors

package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	updaterplugin "github.com/SemRels/updater-npm/internal/plugin"
)

func TestRunUpdatesPackageJSON(t *testing.T) {
	t.Parallel()

	file := filepath.Join(t.TempDir(), "package.json")
	if err := os.WriteFile(file, []byte("{\n    \"name\": \"demo\",\n    \"version\": \"1.0.0\"\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	env := map[string]string{"SEMREL_VERSION": "v1.1.0", "SEMREL_PLUGIN_FILE": file}
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), &stdout, &stderr, func(key string) string { return env[key] }, updaterplugin.NewUpdater(), func(string) (string, error) {
		return "", errors.New("lookPath should not be called")
	}, func(context.Context, string, string, ...string) error {
		return errors.New("runCommand should not be called")
	}); code != 0 {
		t.Fatalf("run() code = %d stderr = %s", code, stderr.String())
	}

	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), `    "version": "1.1.0"`) {
		t.Fatalf("updated file = %s", got)
	}
}

func TestRunLockfileUpdateWarnsWhenNPMMissing(t *testing.T) {
	t.Parallel()

	file := filepath.Join(t.TempDir(), "package.json")
	if err := os.WriteFile(file, []byte("{\n  \"name\": \"demo\",\n  \"version\": \"1.0.0\"\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	env := map[string]string{
		"SEMREL_VERSION":                "1.1.0",
		"SEMREL_PLUGIN_FILE":            file,
		"SEMREL_PLUGIN_UPDATE_LOCKFILE": "true",
	}
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), &stdout, &stderr, func(key string) string { return env[key] }, updaterplugin.NewUpdater(), func(string) (string, error) {
		return "", exec.ErrNotFound
	}, func(context.Context, string, string, ...string) error {
		return errors.New("runCommand should not be called")
	}); code != 0 {
		t.Fatalf("run() code = %d stderr = %s", code, stderr.String())
	}

	if !strings.Contains(stderr.String(), "npm not found on PATH — skipping lockfile update") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunLockfileUpdateWarnsWhenNPMFails(t *testing.T) {
	t.Parallel()

	file := filepath.Join(t.TempDir(), "package.json")
	if err := os.WriteFile(file, []byte("{\n  \"name\": \"demo\",\n  \"version\": \"1.0.0\"\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	env := map[string]string{
		"SEMREL_VERSION":                "1.1.0",
		"SEMREL_PLUGIN_FILE":            file,
		"SEMREL_PLUGIN_UPDATE_LOCKFILE": "true",
	}
	var stdout, stderr bytes.Buffer
	called := false
	if code := run(context.Background(), &stdout, &stderr, func(key string) string { return env[key] }, updaterplugin.NewUpdater(), func(string) (string, error) {
		return "/fake/npm", nil
	}, func(_ context.Context, dir, name string, args ...string) error {
		called = true
		if dir != filepath.Dir(file) {
			t.Fatalf("dir = %q, want %q", dir, filepath.Dir(file))
		}
		if name != "npm" {
			t.Fatalf("name = %q", name)
		}
		if strings.Join(args, " ") != "install --package-lock-only" {
			t.Fatalf("args = %q", args)
		}
		return errors.New("exit status 1")
	}); code != 0 {
		t.Fatalf("run() code = %d stderr = %s", code, stderr.String())
	}

	if !called {
		t.Fatal("runCommand was not called")
	}
	if !strings.Contains(stderr.String(), "npm install --package-lock-only failed: exit status 1") {
		t.Fatalf("stderr = %q", stderr.String())
	}
	if !strings.Contains(stdout.String(), "updated") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunDoesNotUpdateLockfileByDefault(t *testing.T) {
	t.Parallel()

	file := filepath.Join(t.TempDir(), "package.json")
	if err := os.WriteFile(file, []byte("{\n  \"name\": \"demo\",\n  \"version\": \"1.0.0\"\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	env := map[string]string{"SEMREL_VERSION": "1.1.0", "SEMREL_PLUGIN_FILE": file}
	var stdout, stderr bytes.Buffer
	lookedUp := false
	runCalled := false
	if code := run(context.Background(), &stdout, &stderr, func(key string) string { return env[key] }, updaterplugin.NewUpdater(), func(string) (string, error) {
		lookedUp = true
		return "/fake/npm", nil
	}, func(context.Context, string, string, ...string) error {
		runCalled = true
		return nil
	}); code != 0 {
		t.Fatalf("run() code = %d stderr = %s", code, stderr.String())
	}

	if lookedUp {
		t.Fatal("lookPath was called")
	}
	if runCalled {
		t.Fatal("runCommand was called")
	}
}

func TestRunDryRun(t *testing.T) {
	t.Parallel()

	env := map[string]string{"SEMREL_VERSION": "1.1.0", "SEMREL_DRY_RUN": "true"}
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), &stdout, &stderr, func(key string) string { return env[key] }, updaterplugin.NewUpdater(), func(string) (string, error) {
		return "", errors.New("lookPath should not be called")
	}, func(context.Context, string, string, ...string) error {
		return errors.New("runCommand should not be called")
	}); code != 0 {
		t.Fatalf("run() code = %d", code)
	}
	if !strings.Contains(stdout.String(), "[dry-run]") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunRequiresVersion(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), &stdout, &stderr, func(string) string { return "" }, updaterplugin.NewUpdater(), func(string) (string, error) {
		return "", errors.New("lookPath should not be called")
	}, func(context.Context, string, string, ...string) error {
		return errors.New("runCommand should not be called")
	}); code != 1 {
		t.Fatalf("run() code = %d", code)
	}
	if !strings.Contains(stderr.String(), "SEMREL_VERSION is required") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
