## ADDED Requirements

### Requirement: PKCS7 padding validation is constant-time
The `Decrypt` function in `crypto/encrypt.go` SHALL validate PKCS7 padding using a constant-time comparison that does not short-circuit on the first invalid byte. The function SHALL always compare all padding bytes regardless of individual mismatches.

#### Scenario: Correctly padded ciphertext decrypts successfully
- **WHEN** a ciphertext with valid PKCS7 padding is decrypted
- **THEN** the function returns the original plaintext without error

#### Scenario: Invalid padding byte in last position
- **WHEN** a ciphertext with an invalid padding byte (value 0 or value > 16) is decrypted
- **THEN** the function returns a generic "decryption failed" error without revealing which byte was invalid

#### Scenario: Valid padding length but incorrect padding content
- **WHEN** a ciphertext has a padding byte indicating N bytes of padding, but the preceding N-1 bytes are not all equal to N
- **THEN** the function returns the same generic "decryption failed" error as any other padding failure, with no timing difference

### Requirement: Cryptographic errors are generic
All error returns from `Decrypt` for invalid ciphertext, bad padding, or incorrect key SHALL use the same error message text. The function SHALL NOT distinguish between "ciphertext too short", "invalid padding", or other specific failure modes in error output.

#### Scenario: Ciphertext shorter than AES block size
- **WHEN** a ciphertext with fewer than 16 bytes is passed to `Decrypt`
- **THEN** the error message is "decryption failed" (same as padding errors)

#### Scenario: Base64 decode failure
- **WHEN** a non-base64 string is passed to `Decrypt`
- **THEN** the base64 error is wrapped with a generic message, not returned verbatim
