package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/h4ck4life/aix-go/core"
)

func TestCheckRegistryValid(t *testing.T) {
	setupTempCmd(t)

	// Ensure registry is initialized
	registry := core.NewRegistry()
	registry.Load()

	ok, msg := checkRegistry()
	if !ok {
		t.Errorf("expected checkRegistry to pass, got false: %s", msg)
	}
	if msg == "" {
		t.Error("expected non-empty message from checkRegistry")
	}
}

func TestCheckPermissionsWritable(t *testing.T) {
	tmpDir := setupTempCmd(t)

	// Create the directories that checkPermissions will test
	registry := core.NewRegistry()
	registry.Load()

	// Ensure .claude directory exists for settings path
	os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0755)

	ok, msg := checkPermissions()
	if !ok {
		t.Errorf("expected checkPermissions to pass, got false: %s", msg)
	}
	if msg == "" {
		t.Error("expected non-empty message from checkPermissions")
	}
}

func TestCheckTokenStorage(t *testing.T) {
	setupTempCmd(t)

	ok, msg := checkTokenStorage()
	if !ok {
		t.Errorf("expected checkTokenStorage to pass, got false: %s", msg)
	}
	if msg == "" {
		t.Error("expected non-empty message from checkTokenStorage")
	}
}
