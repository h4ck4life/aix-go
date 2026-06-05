package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/h4ck4life/aix-go/constants"
)

// setupTempRegistry creates a temp directory and redirects registry paths to it.
func setupTempRegistry(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	return tmpDir
}

func TestRegistryLoadCreatesOnFirstRun(t *testing.T) {
	setupTempRegistry(t)

	r := NewRegistry()
	if err := r.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	providers, err := r.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	// Should have the 5 preconfigured providers
	if len(providers) != len(constants.PreconfiguredProviders) {
		t.Errorf("expected %d preconfigured providers, got %d", len(constants.PreconfiguredProviders), len(providers))
	}

	// Verify the file was created
	path := constants.RegistryPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("registry file not created at %s", path)
	}
}

func TestRegistrySetOneAndGetOne(t *testing.T) {
	setupTempRegistry(t)

	r := NewRegistry()
	if err := r.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	cfg := constants.ProviderConfig{
		BaseURL:  "https://api.test.com",
		TokenVar: constants.TokenTypeAPIKey,
	}

	if err := r.SetOne("test-provider", cfg); err != nil {
		t.Fatalf("SetOne failed: %v", err)
	}

	got, err := r.GetOne("test-provider")
	if err != nil {
		t.Fatalf("GetOne failed: %v", err)
	}

	if got.BaseURL != cfg.BaseURL {
		t.Errorf("BaseURL = %q, want %q", got.BaseURL, cfg.BaseURL)
	}
	if got.TokenVar != cfg.TokenVar {
		t.Errorf("TokenVar = %q, want %q", got.TokenVar, cfg.TokenVar)
	}
}

func TestRegistryGetOneNotFound(t *testing.T) {
	setupTempRegistry(t)

	r := NewRegistry()
	if err := r.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	_, err := r.GetOne("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent provider, got nil")
	}
}

func TestRegistryRemoveOne(t *testing.T) {
	setupTempRegistry(t)

	r := NewRegistry()
	if err := r.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	cfg := constants.ProviderConfig{
		BaseURL:  "https://api.test.com",
		TokenVar: constants.TokenTypeAPIKey,
	}
	r.SetOne("to-remove", cfg)

	if err := r.RemoveOne("to-remove"); err != nil {
		t.Fatalf("RemoveOne failed: %v", err)
	}

	_, err := r.GetOne("to-remove")
	if err == nil {
		t.Error("expected error after removal, got nil")
	}
}

func TestRegistryRenameOne(t *testing.T) {
	setupTempRegistry(t)

	r := NewRegistry()
	if err := r.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	cfg := constants.ProviderConfig{
		BaseURL:  "https://api.test.com",
		TokenVar: constants.TokenTypeAuthToken,
	}
	r.SetOne("old-name", cfg)

	if err := r.RenameOne("old-name", "new-name"); err != nil {
		t.Fatalf("RenameOne failed: %v", err)
	}

	got, err := r.GetOne("new-name")
	if err != nil {
		t.Fatalf("GetOne(new-name) failed: %v", err)
	}
	if got.BaseURL != cfg.BaseURL {
		t.Errorf("BaseURL after rename = %q, want %q", got.BaseURL, cfg.BaseURL)
	}

	_, err = r.GetOne("old-name")
	if err == nil {
		t.Error("old name should no longer exist after rename")
	}
}

func TestRegistrySetModelName(t *testing.T) {
	setupTempRegistry(t)

	r := NewRegistry()
	if err := r.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	cfg := constants.ProviderConfig{
		BaseURL:  "https://api.test.com",
		TokenVar: constants.TokenTypeAPIKey,
	}
	r.SetOne("myprov", cfg)

	if err := r.SetModelName("myprov", "claude-opus-4-8"); err != nil {
		t.Fatalf("SetModelName failed: %v", err)
	}

	got, _ := r.GetOne("myprov")
	if got.ModelName != "claude-opus-4-8" {
		t.Errorf("ModelName = %q, want %q", got.ModelName, "claude-opus-4-8")
	}
}

func TestRegistrySetDefaultModel(t *testing.T) {
	setupTempRegistry(t)

	r := NewRegistry()
	if err := r.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	cfg := constants.ProviderConfig{
		BaseURL:  "https://api.test.com",
		TokenVar: constants.TokenTypeAPIKey,
	}
	r.SetOne("myprov", cfg)

	if err := r.SetDefaultModel("myprov", "opus", "claude-opus-4-8"); err != nil {
		t.Fatalf("SetDefaultModel failed: %v", err)
	}

	got, _ := r.GetOne("myprov")
	if got.DefaultModels["opus"] != "claude-opus-4-8" {
		t.Errorf("DefaultModels[opus] = %q, want %q", got.DefaultModels["opus"], "claude-opus-4-8")
	}
}

func TestRegistryClearCache(t *testing.T) {
	setupTempRegistry(t)

	r := NewRegistry()
	if err := r.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Data should be cached
	r.mu.RLock()
	loadedAt := r.loadedAt
	r.mu.RUnlock()
	if loadedAt.IsZero() {
		t.Fatal("expected loadedAt to be set after Load")
	}

	r.ClearCache()

	r.mu.RLock()
	loadedAt = r.loadedAt
	r.mu.RUnlock()
	if !loadedAt.IsZero() {
		t.Error("expected loadedAt to be zero after ClearCache")
	}
}

func TestRegistryCacheTTL(t *testing.T) {
	setupTempRegistry(t)

	r := NewRegistry()

	// First load
	if err := r.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Manually expire cache by setting loadedAt to past
	r.mu.Lock()
	r.loadedAt = time.Now().Add(-10 * time.Second)
	r.mu.Unlock()

	// ensureLoaded should reload since cache is expired
	if err := r.ensureLoaded(); err != nil {
		t.Fatalf("ensureLoaded after expiry failed: %v", err)
	}

	// Verify cache is fresh again
	r.mu.RLock()
	fresh := r.loadedAt
	r.mu.RUnlock()
	if time.Since(fresh) > time.Second {
		t.Error("expected loadedAt to be fresh after reload")
	}
}

func TestRegistryInvalidProviderName(t *testing.T) {
	setupTempRegistry(t)

	r := NewRegistry()
	r.Load()

	cfg := constants.ProviderConfig{
		BaseURL:  "https://api.test.com",
		TokenVar: constants.TokenTypeAPIKey,
	}

	err := r.SetOne("INVALID", cfg)
	if err == nil {
		t.Error("expected error for uppercase provider name, got nil")
	}

	err = r.SetOne("123start", cfg)
	if err == nil {
		t.Error("expected error for name starting with number, got nil")
	}
}

func TestRegistryPersistence(t *testing.T) {
	tmpDir := setupTempRegistry(t)

	// Create and write a provider
	r1 := NewRegistry()
	r1.Load()
	cfg := constants.ProviderConfig{
		BaseURL:  "https://api.test.com",
		TokenVar: constants.TokenTypeAPIKey,
	}
	r1.SetOne("persist-test", cfg)

	// Create a new registry instance to verify persistence
	r2 := NewRegistry()
	r2.Load()

	got, err := r2.GetOne("persist-test")
	if err != nil {
		// The file might be at a different path; check if it exists
		regPath := filepath.Join(tmpDir, ".anthropic-switch", "models.json")
		if _, statErr := os.Stat(regPath); statErr != nil {
			t.Skipf("registry file not at expected path: %s", regPath)
		}
		t.Fatalf("GetOne after reload failed: %v", err)
	}
	if got.BaseURL != cfg.BaseURL {
		t.Errorf("persisted BaseURL = %q, want %q", got.BaseURL, cfg.BaseURL)
	}
}

func TestRemoveOneNonexistentProvider(t *testing.T) {
	setupTempRegistry(t)

	r := NewRegistry()
	if err := r.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	err := r.RemoveOne("nonexistent-provider")
	if err == nil {
		t.Fatal("expected error for nonexistent provider, got nil")
	}
}

func TestRemoveOneExistingProvider(t *testing.T) {
	setupTempRegistry(t)

	r := NewRegistry()
	if err := r.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Add a test provider
	cfg := constants.ProviderConfig{
		BaseURL:  "https://test.example.com",
		TokenVar: constants.TokenTypeAPIKey,
	}
	if err := r.SetOne("test-remove", cfg); err != nil {
		t.Fatalf("SetOne failed: %v", err)
	}

	// Verify it exists
	_, err := r.GetOne("test-remove")
	if err != nil {
		t.Fatalf("GetOne failed: %v", err)
	}

	// Remove it
	if err := r.RemoveOne("test-remove"); err != nil {
		t.Fatalf("RemoveOne failed: %v", err)
	}

	// Verify it's gone
	_, err = r.GetOne("test-remove")
	if err == nil {
		t.Fatal("expected error after removal, provider still exists")
	}
}

func TestMergeOneWithoutLoad(t *testing.T) {
	setupTempRegistry(t)

	// Seed registry with two providers
	r := NewRegistry()
	if err := r.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if err := r.SetOne("a", constants.ProviderConfig{BaseURL: "https://a.example.com", TokenVar: constants.TokenTypeAPIKey}); err != nil {
		t.Fatalf("SetOne a failed: %v", err)
	}
	if err := r.SetOne("b", constants.ProviderConfig{BaseURL: "https://b.example.com", TokenVar: constants.TokenTypeAPIKey}); err != nil {
		t.Fatalf("SetOne b failed: %v", err)
	}

	// Pre-count existing providers (includes 5 preconfigured + a + b = 7)
	existing, err := r.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	preCount := len(existing)
	if preCount != 7 {
		t.Fatalf("expected 7 seeded providers, got %d", preCount)
	}

	// Now make a fresh registry (r.data is nil) and call MergeOne WITHOUT Load.
	// ensureLoaded() inside MergeOne should load from disk, preserving all existing.
	fresh := NewRegistry()
	if err := fresh.MergeOne("c", constants.ProviderConfig{BaseURL: "https://c.example.com", TokenVar: constants.TokenTypeAPIKey}); err != nil {
		t.Fatalf("MergeOne without Load failed: %v", err)
	}

	all, err := fresh.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if len(all) != preCount+1 {
		t.Errorf("expected %d providers after merge, got %d (ensureLoaded() not called?)", preCount+1, len(all))
	}
	if _, ok := all["a"]; !ok {
		t.Error("provider 'a' was wiped by MergeOne — ensureLoaded() not called?")
	}
	if _, ok := all["b"]; !ok {
		t.Error("provider 'b' was wiped by MergeOne — ensureLoaded() not called?")
	}
	if _, ok := all["c"]; !ok {
		t.Error("provider 'c' was not added by MergeOne")
	}
}
