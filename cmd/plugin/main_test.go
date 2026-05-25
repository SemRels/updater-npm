package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunUpdatesPackageJSON(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "updater-npm-main-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	file := filepath.Join(dir, "package.json")
	if err := os.WriteFile(file, []byte("{\n  \"name\": \"demo\",\n  \"version\": \"1.0.0\"\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	env := map[string]string{"SEMREL_VERSION": "v1.1.0", "SEMREL_PLUGIN_FILE": file}
	var stdout, stderr bytes.Buffer
	if code := run(&stdout, &stderr, func(key string) string { return env[key] }); code != 0 {
		t.Fatalf("run() code = %d stderr = %s", code, stderr.String())
	}

	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), `"version": "1.1.0"`) {
		t.Fatalf("updated file = %s", got)
	}
}

func TestRunDryRun(t *testing.T) {
	t.Parallel()

	env := map[string]string{"SEMREL_VERSION": "1.1.0", "SEMREL_DRY_RUN": "true"}
	var stdout, stderr bytes.Buffer
	if code := run(&stdout, &stderr, func(key string) string { return env[key] }); code != 0 {
		t.Fatalf("run() code = %d", code)
	}
	if !strings.Contains(stdout.String(), "[dry-run]") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunRequiresVersion(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	if code := run(&stdout, &stderr, func(string) string { return "" }); code != 1 {
		t.Fatalf("run() code = %d", code)
	}
	if !strings.Contains(stderr.String(), "SEMREL_VERSION is required") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
