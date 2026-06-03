package core

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/h4ck4life/aix-go/constants"
	"github.com/h4ck4life/aix-go/crypto"
	"github.com/h4ck4life/aix-go/keychain"
	"github.com/h4ck4life/aix-go/utils"
)

// TokenManager handles token storage
type TokenManager struct {
	mu          sync.RWMutex
	useKeychain bool
}

// NewTokenManager creates a new token manager
func NewTokenManager() *TokenManager {
	return &TokenManager{
		useKeychain: keychain.IsAvailable(),
	}
}

// GetToken retrieves a token for a provider
func (tm *TokenManager) GetToken(providerName string) (string, error) {
	account := keychain.AccountName(providerName)

	// Try keychain first
	if tm.useKeychain {
		token, err := keychain.Get(constants.KeychainService, account)
		if err == nil {
			return token, nil
		}
	}

	// Fall back to encrypted file
	return tm.getFromFile(account)
}

// SetToken stores a token for a provider
func (tm *TokenManager) SetToken(providerName, token string) error {
	account := keychain.AccountName(providerName)

	if tm.useKeychain {
		return keychain.Set(constants.KeychainService, account, token)
	}

	return tm.setInFile(account, token)
}

// DeleteToken removes a token for a provider
func (tm *TokenManager) DeleteToken(providerName string) error {
	account := keychain.AccountName(providerName)

	var keychainErr error

	if tm.useKeychain {
		keychainErr = keychain.Delete(constants.KeychainService, account)
	}

	tm.mu.Lock()
	fileErr := tm.deleteFromFile(account)
	tm.mu.Unlock()

	if keychainErr != nil && fileErr != nil {
		return utils.NewTokenError(fmt.Sprintf("failed to delete token: keychain=%v, file=%v", keychainErr, fileErr))
	}

	return nil
}

// MoveToken moves a token from one provider to another
func (tm *TokenManager) MoveToken(oldProvider, newProvider string) error {
	token, err := tm.GetToken(oldProvider)
	if err != nil {
		// No token to move; rename should still succeed
		return nil
	}

	if err := tm.SetToken(newProvider, token); err != nil {
		return err
	}

	return tm.DeleteToken(oldProvider)
}

// HasToken checks if a token exists for a provider
func (tm *TokenManager) HasToken(providerName string) bool {
	_, err := tm.GetToken(providerName)
	return err == nil
}

// GetStorageInfo returns the active storage backend
func (tm *TokenManager) GetStorageInfo() string {
	if tm.useKeychain {
		return "keychain"
	}
	return "encrypted file"
}

// getFromFile retrieves a token from the encrypted file
func (tm *TokenManager) getFromFile(account string) (string, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	path := constants.TokenEncPath()
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Account names contain colons (format: "username:provider"),
		// so split at the LAST colon to separate account from encrypted value.
		idx := strings.LastIndex(line, ":")
		if idx < 0 {
			continue
		}
		if line[:idx] == account {
			return crypto.Decrypt(line[idx+1:])
		}
	}

	return "", utils.NewTokenError(fmt.Sprintf("token not found for %s", account))
}

// setInFile stores a token in the encrypted file using a single atomic read-modify-write.
func (tm *TokenManager) setInFile(account, token string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	encrypted, err := crypto.Encrypt(token)
	if err != nil {
		return err
	}

	path := constants.TokenEncPath()
	newEntry := fmt.Sprintf("%s:%s", account, encrypted)

	// Read existing lines
	lines, err := tm.readFileLines(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Replace existing entry or append
	replaced := false
	for i, line := range lines {
		idx := strings.LastIndex(line, ":")
		if idx >= 0 && line[:idx] == account {
			lines[i] = newEntry
			replaced = true
			break
		}
	}
	if !replaced {
		lines = append(lines, newEntry)
	}

	return tm.writeFileAtomically(path, lines)
}

// deleteFromFile removes a token from the encrypted file using atomic writes.
func (tm *TokenManager) deleteFromFile(account string) error {
	path := constants.TokenEncPath()

	lines, err := tm.readFileLines(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var kept []string
	for _, line := range lines {
		idx := strings.LastIndex(line, ":")
		if idx >= 0 && line[:idx] == account {
			continue
		}
		kept = append(kept, line)
	}

	if len(kept) == 0 {
		return os.Remove(path)
	}

	return tm.writeFileAtomically(path, kept)
}

// readFileLines reads the token file and returns non-empty lines.
func (tm *TokenManager) readFileLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	raw := strings.Split(string(data), "\n")
	var lines []string
	for _, line := range raw {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines, nil
}

// writeFileAtomically writes lines to a file using temp-file-then-rename.
func (tm *TokenManager) writeFileAtomically(path string, lines []string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	content := strings.Join(lines, "\n") + "\n"
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(content), 0600); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
