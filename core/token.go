package core

import (
	"bufio"
	"fmt"
	"os"
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
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		if parts[0] == account {
			return crypto.Decrypt(parts[1])
		}
	}

	return "", utils.NewTokenError(fmt.Sprintf("token not found for %s", account))
}

// setInFile stores a token in the encrypted file
func (tm *TokenManager) setInFile(account, token string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Remove existing entry first
	_ = tm.deleteFromFile(account)

	encrypted, err := crypto.Encrypt(token)
	if err != nil {
		return err
	}

	path := constants.TokenEncPath()
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "%s:%s\n", account, encrypted)
	return err
}

// deleteFromFile removes a token from the encrypted file
func (tm *TokenManager) deleteFromFile(account string) error {
	path := constants.TokenEncPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	lines := strings.Split(string(data), "\n")
	var kept []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 && parts[0] == account {
			continue
		}
		kept = append(kept, line)
	}

	if len(kept) == 0 {
		return os.Remove(path)
	}

	return os.WriteFile(path, []byte(strings.Join(kept, "\n")+"\n"), 0600)
}
