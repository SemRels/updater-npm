package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	updaterplugin "github.com/SemRels/updater-npm/internal/plugin"
)

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func TestRunSucceedsAfterMutationWhenOutputFails(t *testing.T) {
	file := filepath.Join(t.TempDir(), "package.json")
	if err := os.WriteFile(file, []byte("{\n  \"version\": \"1.0.0\"\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	env := map[string]string{"SEMREL_VERSION": "v1.1.0", "SEMREL_PLUGIN_FILE": file, "SEMREL_PLUGIN_UPDATE_LOCKFILE": "true"}
	code := run(context.Background(), failingWriter{}, failingWriter{}, func(key string) string { return env[key] }, updaterplugin.NewUpdater(), func(string) (string, error) {
		return "", errors.New("npm unavailable")
	}, func(context.Context, string, string, ...string) error {
		return errors.New("run command should not be called")
	})
	if code != 0 {
		t.Fatalf("run() code = %d, want 0", code)
	}
}
