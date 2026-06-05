package core

import (
	"errors"
	"os"
	"testing"

	"github.com/h4ck4life/aix-go/constants"
)

func setupTempToken(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
}

func newTestTokenManager() *TokenManager {
	return &TokenManager{useKeychain: false}
}

func TestTokenSetAndGet(t *testing.T) {
	setupTempToken(t)
	tm := newTestTokenManager()

	if err := tm.SetToken("myprov", "secret-token-123"); err != nil {
		t.Fatalf("SetToken failed: %v", err)
	}

	got, err := tm.GetToken("myprov")
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}
	if got != "secret-token-123" {
		t.Errorf("GetToken = %q, want %q", got, "secret-token-123")
	}
}

func TestTokenGetNotFound(t *testing.T) {
	setupTempToken(t)
	tm := newTestTokenManager()

	_, err := tm.GetToken("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent token, got nil")
	}
}

func TestTokenUpdate(t *testing.T) {
	setupTempToken(t)
	tm := newTestTokenManager()

	tm.SetToken("myprov", "old-token")
	tm.SetToken("myprov", "new-token")

	got, err := tm.GetToken("myprov")
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}
	if got != "new-token" {
		t.Errorf("GetToken after update = %q, want %q", got, "new-token")
	}
}

func TestTokenDelete(t *testing.T) {
	setupTempToken(t)
	tm := newTestTokenManager()

	tm.SetToken("myprov", "to-delete")
	if err := tm.DeleteToken("myprov"); err != nil {
		t.Fatalf("DeleteToken failed: %v", err)
	}

	_, err := tm.GetToken("myprov")
	if err == nil {
		t.Error("expected error after deletion, got nil")
	}
}

func TestTokenDeleteNonexistent(t *testing.T) {
	setupTempToken(t)
	tm := newTestTokenManager()

	// Deleting a nonexistent token should not error
	if err := tm.DeleteToken("ghost"); err != nil {
		t.Errorf("DeleteToken for nonexistent token should not error: %v", err)
	}
}

func TestTokenMove(t *testing.T) {
	setupTempToken(t)
	tm := newTestTokenManager()

	tm.SetToken("old-prov", "movable-token")
	if err := tm.MoveToken("old-prov", "new-prov"); err != nil {
		t.Fatalf("MoveToken failed: %v", err)
	}

	got, err := tm.GetToken("new-prov")
	if err != nil {
		t.Fatalf("GetToken(new-prov) failed: %v", err)
	}
	if got != "movable-token" {
		t.Errorf("GetToken after move = %q, want %q", got, "movable-token")
	}

	// Old provider should no longer have the token
	_, err = tm.GetToken("old-prov")
	if err == nil {
		t.Error("old provider should not have token after move")
	}
}

func TestTokenMoveNoToken(t *testing.T) {
	setupTempToken(t)
	tm := newTestTokenManager()

	// Moving when no token exists should succeed silently
	if err := tm.MoveToken("empty-prov", "new-prov"); err != nil {
		t.Errorf("MoveToken with no token should not error: %v", err)
	}
}

func TestTokenGetReturnsErrNotFound(t *testing.T) {
	setupTempToken(t)
	tm := newTestTokenManager()

	// GetToken on nonexistent provider should return ErrTokenNotFound, not a generic error
	_, err := tm.GetToken("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent token")
	}
	if !errors.Is(err, ErrTokenNotFound) {
		t.Errorf("expected ErrTokenNotFound, got: %v", err)
	}
}

func TestTokenHasToken(t *testing.T) {
	setupTempToken(t)
	tm := newTestTokenManager()

	if tm.HasToken("myprov") {
		t.Error("HasToken should be false before setting")
	}

	tm.SetToken("myprov", "some-token")
	if !tm.HasToken("myprov") {
		t.Error("HasToken should be true after setting")
	}
}

func TestTokenGetStorageInfo(t *testing.T) {
	tm := &TokenManager{useKeychain: false}
	if got := tm.GetStorageInfo(); got != "encrypted file" {
		t.Errorf("GetStorageInfo(useKeychain=false) = %q, want %q", got, "encrypted file")
	}

	tm = &TokenManager{useKeychain: true}
	if got := tm.GetStorageInfo(); got != "keychain" {
		t.Errorf("GetStorageInfo(useKeychain=true) = %q, want %q", got, "keychain")
	}
}

func TestTokenMultipleProviders(t *testing.T) {
	setupTempToken(t)
	tm := newTestTokenManager()

	tm.SetToken("prov-a", "token-a")
	tm.SetToken("prov-b", "token-b")
	tm.SetToken("prov-c", "token-c")

	got, _ := tm.GetToken("prov-a")
	if got != "token-a" {
		t.Errorf("prov-a token = %q, want %q", got, "token-a")
	}
	got, _ = tm.GetToken("prov-b")
	if got != "token-b" {
		t.Errorf("prov-b token = %q, want %q", got, "token-b")
	}
	got, _ = tm.GetToken("prov-c")
	if got != "token-c" {
		t.Errorf("prov-c token = %q, want %q", got, "token-c")
	}
}

func TestTokenFileNotCorruptedOnCrash(t *testing.T) {
	setupTempToken(t)
	tm := newTestTokenManager()

	// Set a token, verify file exists
	tm.SetToken("prov1", "first-token")

	// Set another token (should use atomic write, not delete-then-append)
	tm.SetToken("prov2", "second-token")

	// Both tokens should survive
	got1, err := tm.GetToken("prov1")
	if err != nil {
		t.Fatalf("prov1 lost after prov2 set: %v", err)
	}
	if got1 != "first-token" {
		t.Errorf("prov1 token = %q, want %q", got1, "first-token")
	}

	got2, err := tm.GetToken("prov2")
	if err != nil {
		t.Fatalf("prov2 not found: %v", err)
	}
	if got2 != "second-token" {
		t.Errorf("prov2 token = %q, want %q", got2, "second-token")
	}

	// Verify the file only has 2 entries (no duplicates)
	path := constants.TokenEncPath()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read token file failed: %v", err)
	}
	lines := 0
	for _, b := range data {
		if b == '\n' {
			lines++
		}
	}
	if lines != 2 {
		t.Errorf("token file has %d lines, want 2 (no duplicates)", lines)
	}
}
