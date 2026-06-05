package cmd

import (
	"testing"

	"github.com/h4ck4life/aix-go/constants"
	"github.com/h4ck4life/aix-go/core"
	"github.com/h4ck4life/aix-go/utils"
	"github.com/spf13/cobra"
)

// setupTempCmd creates a temp directory and redirects paths to it for cmd tests
func setupTempCmd(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	return tmpDir
}

func TestRunProviderAddHappyPath(t *testing.T) {
	setupTempCmd(t)

	cmd := &cobra.Command{}
	args := []string{"myprovider", "https://api.example.com"}

	err := runProviderAdd(cmd, args)
	if err != nil {
		t.Fatalf("runProviderAdd failed: %v", err)
	}

	// Verify provider was stored
	registry := core.NewRegistry()
	if err := registry.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	providers, _ := registry.GetAll()
	cfg, ok := providers["myprovider"]
	if !ok {
		t.Fatal("expected provider 'myprovider' to be in registry")
	}
	if cfg.BaseURL != "https://api.example.com" {
		t.Errorf("expected BaseURL 'https://api.example.com', got '%s'", cfg.BaseURL)
	}
}

func TestRunProviderAddMissingArgs(t *testing.T) {
	setupTempCmd(t)

	cmd := &cobra.Command{}
	args := []string{"myprovider"} // missing URL

	err := runProviderAdd(cmd, args)
	if err == nil {
		t.Fatal("expected error for missing args")
	}

	ve, ok := err.(*utils.ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}
	if ve.Code != 2 {
		t.Errorf("expected exit code 2, got %d", ve.Code)
	}
}

func TestRunProviderAddInvalidName(t *testing.T) {
	setupTempCmd(t)

	cmd := &cobra.Command{}
	args := []string{"MyProvider", "https://api.example.com"} // uppercase

	err := runProviderAdd(cmd, args)
	if err == nil {
		t.Fatal("expected error for invalid name")
	}

	_, ok := err.(*utils.ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestRunProviderRemoveNonexistent(t *testing.T) {
	setupTempCmd(t)

	// Load registry first so it initializes
	registry := core.NewRegistry()
	registry.Load()

	cmd := &cobra.Command{}
	providerRemoveYes = true // skip confirmation prompt
	defer func() { providerRemoveYes = false }()

	args := []string{"nonexistent"}
	err := runProviderRemove(cmd, args)
	if err == nil {
		t.Fatal("expected error for nonexistent provider")
	}

	_, ok := err.(*utils.ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestRunProviderRenameNonexistent(t *testing.T) {
	setupTempCmd(t)

	registry := core.NewRegistry()
	registry.Load()

	cmd := &cobra.Command{}
	providerRenameYes = true
	defer func() { providerRenameYes = false }()

	args := []string{"nonexistent", "newname"}
	err := runProviderRename(cmd, args)
	if err == nil {
		t.Fatal("expected error for nonexistent provider")
	}

	_, ok := err.(*utils.ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestRunProviderUseNonexistent(t *testing.T) {
	setupTempCmd(t)

	registry := core.NewRegistry()
	registry.Load()

	cmd := &cobra.Command{}
	args := []string{"nonexistent"}

	err := runProviderUse(cmd, args)
	if err == nil {
		t.Fatal("expected error for nonexistent provider")
	}
}

func TestRunProviderListEmpty(t *testing.T) {
	setupTempCmd(t)

	// Create empty registry file
	registry := core.NewRegistry()
	registry.Load()

	// Remove all providers to make it empty
	providers, _ := registry.GetAll()
	for name := range providers {
		registry.RemoveOne(name)
	}

	cmd := &cobra.Command{}
	args := []string{}

	err := runProviderList(cmd, args)
	if err != nil {
		t.Fatalf("runProviderList failed: %v", err)
	}
}

func TestRunProviderEditNoFlags(t *testing.T) {
	setupTempCmd(t)

	registry := core.NewRegistry()
	registry.Load()

	// Add a provider first
	registry.SetOne("testprovider", constants.ProviderConfig{
		BaseURL:  "https://api.example.com",
		TokenVar: constants.TokenTypeAPIKey,
	})

	// Reset flags
	providerEditURL = ""
	providerEditTokenType = ""
	providerEditModel = ""
	providerEditDefaultModel = ""

	cmd := &cobra.Command{}
	args := []string{"testprovider"}

	err := runProviderEdit(cmd, args)
	if err == nil {
		t.Fatal("expected error for no changes")
	}

	_, ok := err.(*utils.ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}
}
