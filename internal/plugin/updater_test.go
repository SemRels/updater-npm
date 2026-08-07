// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The semrel Authors

package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdaterUpdate(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "updater-npm-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("remove temporary directory: %v", err)
		}
	})

	file := filepath.Join(dir, "package.json")
	original := "{\n    \"name\": \"demo\",\n    \"version\": \"1.2.3\",\n    \"private\": true\n}\n"
	if err := os.WriteFile(file, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := NewUpdater().Update(file, "1.3.0"); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), `    "version": "1.3.0",`) {
		t.Fatalf("updated file = %s", got)
	}
}

func TestUpdaterMissingFile(t *testing.T) {
	t.Parallel()

	err := NewUpdater().Update(filepath.Join(t.TempDir(), "package.json"), "1.3.0")
	if err == nil || !strings.Contains(err.Error(), "read") {
		t.Fatalf("expected read error, got %v", err)
	}
}

func TestUpdaterMissingVersionField(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "updater-npm-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("remove temporary directory: %v", err)
		}
	})

	file := filepath.Join(dir, "package.json")
	if err := os.WriteFile(file, []byte("{\n  \"name\": \"demo\"\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err = NewUpdater().Update(file, "1.3.0")
	if err == nil || !strings.Contains(err.Error(), "version field not found") {
		t.Fatalf("expected version error, got %v", err)
	}
}

func TestIndentation(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "updater-npm-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("remove temporary directory: %v", err)
		}
	})

	file := filepath.Join(dir, "package.json")
	if err := os.WriteFile(file, []byte("{\n    \"name\": \"demo\"\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	indent, err := Indentation(file)
	if err != nil {
		t.Fatalf("Indentation() error = %v", err)
	}
	if indent != "    " {
		t.Fatalf("indent = %q", indent)
	}
}

func TestUpdaterInvalidJSON(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "updater-npm-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("remove temporary directory: %v", err)
		}
	})

	file := filepath.Join(dir, "package.json")
	if err := os.WriteFile(file, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}

	err = NewUpdater().Update(file, "1.3.0")
	if err == nil || !strings.Contains(err.Error(), "parse") {
		t.Fatalf("expected parse error, got %v", err)
	}
}

func TestIndentationMissingFile(t *testing.T) {
	t.Parallel()

	_, err := Indentation(filepath.Join(t.TempDir(), "package.json"))
	if err == nil || !strings.Contains(err.Error(), "read") {
		t.Fatalf("expected read error, got %v", err)
	}
}

func TestIndentationNoIndent(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "updater-npm-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("remove temporary directory: %v", err)
		}
	})

	file := filepath.Join(dir, "package.json")
	if err := os.WriteFile(file, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	indent, err := Indentation(file)
	if err != nil {
		t.Fatalf("Indentation() error = %v", err)
	}
	if indent != "" {
		t.Fatalf("indent = %q", indent)
	}
}
