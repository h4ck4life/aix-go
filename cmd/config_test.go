package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/h4ck4life/aix-go/constants"
	"github.com/h4ck4life/aix-go/core"
	"github.com/spf13/cobra"
)

func TestConfigImportInvalidJSON(t *testing.T) {
	setupTempCmd(t)

	// Create invalid JSON file
	invalidFile := filepath.Join(t.TempDir(), "invalid.json")
	os.WriteFile(invalidFile, []byte("not json"), 0600)

	cmd := &cobra.Command{}
	args := []string{invalidFile}

	err := runConfigImport(cmd, args)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestConfigImportNonexistentFile(t *testing.T) {
	setupTempCmd(t)

	cmd := &cobra.Command{}
	args := []string{"/nonexistent/file.json"}

	err := runConfigImport(cmd, args)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestConfigExportRoundtrip(t *testing.T) {
	tmpDir := setupTempCmd(t)

	// Add a provider
	registry := core.NewRegistry()
	registry.Load()
	registry.SetOne("testexport", constants.ProviderConfig{
		BaseURL:  "https://api.example.com",
		TokenVar: constants.TokenTypeAPIKey,
	})

	// Export
	exportFile := filepath.Join(tmpDir, "export.json")
	configExportOutput = exportFile
	defer func() { configExportOutput = "" }()

	cmd := &cobra.Command{}
	err := runConfigExport(cmd, []string{})
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	// Verify export file exists and contains our provider
	data, err := os.ReadFile(exportFile)
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}

	var exported map[string]constants.ProviderConfig
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("failed to parse export: %v", err)
	}

	cfg, ok := exported["testexport"]
	if !ok {
		t.Fatal("expected 'testexport' in exported data")
	}
	if cfg.BaseURL != "https://api.example.com" {
		t.Errorf("expected BaseURL 'https://api.example.com', got '%s'", cfg.BaseURL)
	}
}

func TestConfigImportMerge(t *testing.T) {
	tmpDir := setupTempCmd(t)

	// Add existing provider
	registry := core.NewRegistry()
	registry.Load()
	registry.SetOne("existing", constants.ProviderConfig{
		BaseURL:   "https://api.example.com",
		TokenVar:  constants.TokenTypeAPIKey,
		ModelName: "old-model",
	})

	// Create import file with updated provider and new provider
	imported := map[string]constants.ProviderConfig{
		"existing": {
			BaseURL:   "https://api.example.com",
			ModelName: "new-model",
		},
		"newprovider": {
			BaseURL:  "https://api.new.com",
			TokenVar: constants.TokenTypeAuthToken,
		},
	}
	importFile := filepath.Join(tmpDir, "import.json")
	data, _ := json.Marshal(imported)
	os.WriteFile(importFile, data, 0600)

	configImportMerge = true
	defer func() { configImportMerge = false }()

	cmd := &cobra.Command{}
	err := runConfigImport(cmd, []string{importFile})
	if err != nil {
		t.Fatalf("import --merge failed: %v", err)
	}

	// Verify merge behavior
	registry2 := core.NewRegistry()
	registry2.Load()
	providers, _ := registry2.GetAll()

	// Existing provider should have merged model name
	cfg := providers["existing"]
	if cfg.ModelName != "new-model" {
		t.Errorf("expected merged model 'new-model', got '%s'", cfg.ModelName)
	}
	if cfg.TokenVar != constants.TokenTypeAPIKey {
		t.Errorf("expected original tokenVar preserved, got '%s'", cfg.TokenVar)
	}

	// New provider should be added
	_, ok := providers["newprovider"]
	if !ok {
		t.Fatal("expected 'newprovider' to be added")
	}
}
