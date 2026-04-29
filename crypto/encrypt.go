package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/h4ck4life/aix-go/constants"
)

// Encrypt encrypts plaintext with AES-256-CBC using the key from ~/.aix/key
func Encrypt(plaintext string) (string, error) {
	key, err := loadOrCreateKey()
	if err != nil {
		return "", fmt.Errorf("failed to load key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// PKCS7 padding
	padding := aes.BlockSize - len(plaintext)%aes.BlockSize
	paddedText := plaintext + strings.Repeat(string(rune(padding)), padding)

	ciphertext := make([]byte, aes.BlockSize+len(paddedText))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], []byte(paddedText))

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext with AES-256-CBC using the key from ~/.aix/key
func Decrypt(ciphertext string) (string, error) {
	key, err := loadOrCreateKey()
	if err != nil {
		return "", fmt.Errorf("failed to load key: %w", err)
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	if len(data) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(data, data)

	// Remove PKCS7 padding
	padding := int(data[len(data)-1])
	if padding > aes.BlockSize || padding == 0 {
		return "", errors.New("invalid padding")
	}
	for i := len(data) - padding; i < len(data); i++ {
		if int(data[i]) != padding {
			return "", errors.New("invalid padding")
		}
	}

	return string(data[:len(data)-padding]), nil
}

// loadOrCreateKey loads the encryption key from ~/.aix/key, creating it if missing
func loadOrCreateKey() ([]byte, error) {
	keyPath := constants.TokenKeyPath()

	// Ensure directory exists
	if err := os.MkdirAll(constants.TokenDir(), 0700); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(keyPath)
	if err == nil {
		return data, nil
	}

	if !os.IsNotExist(err) {
		return nil, err
	}

	// Generate new key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	if err := os.WriteFile(keyPath, key, 0600); err != nil {
		return nil, err
	}

	return key, nil
}
