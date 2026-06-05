package crypto

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTempKeyDir redirects key storage to a temp directory for the test.
func setupTempKeyDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	// os.UserHomeDir() reads $HOME on Unix
	t.Setenv("HOME", tmpDir)
	return tmpDir
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	setupTempKeyDir(t)

	tests := []struct {
		name      string
		plaintext string
	}{
		{"simple string", "hello world"},
		{"empty string", ""},
		{"special characters", "p@$$w0rd!#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"unicode", "日本語テスト 🔑"},
		{"very long string", string(make([]byte, 1024))}, // 1024 zero bytes
		{"newline in value", "line1\nline2\nline3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			if encrypted == tt.plaintext {
				t.Error("Encrypted value should not equal plaintext")
			}

			decrypted, err := Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			if decrypted != tt.plaintext {
				t.Errorf("Round-trip failed: got %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncryptProducesDifferentCiphertexts(t *testing.T) {
	setupTempKeyDir(t)

	plaintext := "same-input"
	enc1, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	enc2, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if enc1 == enc2 {
		t.Error("Two encryptions of the same plaintext should produce different ciphertexts (random IV)")
	}
}

func TestDecryptInvalidInputs(t *testing.T) {
	setupTempKeyDir(t) // ensure valid key exists for these tests

	tests := []struct {
		name       string
		ciphertext string
		wantErr    string
	}{
		{"empty string", "", "decryption failed"},
		{"invalid base64", "not-valid-base64!!!", "decryption failed"},
		{"too short (less than block size)", "AAAA", "decryption failed"},
		{"valid base64 but corrupt", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", "decryption failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decrypt(tt.ciphertext)
			if err == nil {
				t.Error("expected error for invalid ciphertext, got nil")
			}
			if err != nil && err.Error() != tt.wantErr {
				t.Errorf("expected error %q, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestLoadOrCreateKey(t *testing.T) {
	tmpDir := setupTempKeyDir(t)
	keyPath := filepath.Join(tmpDir, ".aix", "key")

	// First call creates the key
	key1, err := loadOrCreateKey()
	if err != nil {
		t.Fatalf("first loadOrCreateKey failed: %v", err)
	}
	if len(key1) != 32 {
		t.Errorf("key length = %d, want 32", len(key1))
	}

	// Verify the key file exists
	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("key file not created at %s: %v", keyPath, err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("key file permissions = %o, want 0600", info.Mode().Perm())
	}

	// Second call loads the same key
	key2, err := loadOrCreateKey()
	if err != nil {
		t.Fatalf("second loadOrCreateKey failed: %v", err)
	}
	if string(key1) != string(key2) {
		t.Error("second load should return the same key")
	}
}

func TestLoadKeyTruncatedFile(t *testing.T) {
	tmpDir := setupTempKeyDir(t)
	keyPath := filepath.Join(tmpDir, ".aix", "key")

	// Create truncated key file (16 bytes instead of 32)
	os.MkdirAll(filepath.Dir(keyPath), 0700)
	os.WriteFile(keyPath, make([]byte, 16), 0600)

	_, err := loadOrCreateKey()
	if err == nil {
		t.Fatal("expected error for truncated key file")
	}
	if err.Error() != "invalid key file: expected 32 bytes, got 16" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadKeyOversizedFile(t *testing.T) {
	tmpDir := setupTempKeyDir(t)
	keyPath := filepath.Join(tmpDir, ".aix", "key")

	// Create oversized key file (64 bytes instead of 32)
	os.MkdirAll(filepath.Dir(keyPath), 0700)
	os.WriteFile(keyPath, make([]byte, 64), 0600)

	_, err := loadOrCreateKey()
	if err == nil {
		t.Fatal("expected error for oversized key file")
	}
	if err.Error() != "invalid key file: expected 32 bytes, got 64" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadKeyValidFile(t *testing.T) {
	tmpDir := setupTempKeyDir(t)
	keyPath := filepath.Join(tmpDir, ".aix", "key")

	// Create valid 32-byte key file
	expectedKey := make([]byte, 32)
	for i := range expectedKey {
		expectedKey[i] = byte(i)
	}
	os.MkdirAll(filepath.Dir(keyPath), 0700)
	os.WriteFile(keyPath, expectedKey, 0600)

	key, err := loadOrCreateKey()
	if err != nil {
		t.Fatalf("loadOrCreateKey failed: %v", err)
	}
	if len(key) != 32 {
		t.Errorf("key length = %d, want 32", len(key))
	}
	for i := range expectedKey {
		if key[i] != expectedKey[i] {
			t.Errorf("key[%d] = %d, want %d", i, key[i], expectedKey[i])
		}
	}
}
