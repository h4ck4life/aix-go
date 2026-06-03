package keychain

import (
	"fmt"
	"os"
	"os/user"

	"github.com/h4ck4life/aix-go/constants"
	"github.com/zalando/go-keyring"
)

// Get retrieves a token from the OS keychain
func Get(service, account string) (string, error) {
	return keyring.Get(service, account)
}

// Set stores a token in the OS keychain
func Set(service, account, password string) error {
	return keyring.Set(service, account, password)
}

// Delete removes a token from the OS keychain
func Delete(service, account string) error {
	return keyring.Delete(service, account)
}

// AccountName formats the keychain account name
func AccountName(providerName string) string {
	u, err := user.Current()
	if err != nil {
		return fmt.Sprintf("unknown:%s", providerName)
	}
	return fmt.Sprintf("%s:%s", u.Username, providerName)
}

// IsAvailable checks if the OS keychain is available
func IsAvailable() bool {
	testAccount := fmt.Sprintf("aix-test-%d", os.Getpid())
	_ = Delete(constants.KeychainService, testAccount)
	err := Set(constants.KeychainService, testAccount, "test")
	if err != nil {
		return false
	}
	_, err = Get(constants.KeychainService, testAccount)
	_ = Delete(constants.KeychainService, testAccount)
	return err == nil
}
