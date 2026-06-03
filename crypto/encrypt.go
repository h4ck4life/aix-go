package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/h4ck4life/aix-go/constants"
)

// errDecryptionFailed is the generic error returned for all decryption failures.
var errDecryptionFailed = errors.New("decryption failed")

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
		return "", errDecryptionFailed
	}

	if len(data) < aes.BlockSize {
		return "", errDecryptionFailed
	}

	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]

	if len(data)%aes.BlockSize != 0 {
		return "", errDecryptionFailed
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errDecryptionFailed
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(data, data)

	// Constant-time PKCS7 padding validation
	padding := int(data[len(data)-1])

	// Check all potential padding bytes unconditionally.
	// For positions within the padding range, verify they equal the padding value.
	// Accumulate mismatches so we never short-circuit.
	var mismatch int
	for i := 0; i < aes.BlockSize; i++ {
		b := int(data[len(data)-1-i])
		// 1 if this byte is within the padding range, 0 otherwise
		inRange := subtle.ConstantTimeLessOrEq(i, padding-1)
		// XOR: 0 if b == padding, non-zero otherwise
		xor := b ^ padding
		// Only count mismatches for bytes in the padding range
		mismatch |= subtle.ConstantTimeSelect(inRange, xor, 0)
	}

	// padding must be 1..BlockSize
	validRange := subtle.ConstantTimeLessOrEq(1, padding) & subtle.ConstantTimeLessOrEq(padding, aes.BlockSize)

	// validRange==1 and mismatch==0 means valid padding
	if validRange&int(subtle.ConstantTimeEq(int32(0), int32(mismatch))) != 1 {
		return "", errDecryptionFailed
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
