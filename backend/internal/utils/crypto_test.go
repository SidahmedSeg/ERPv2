package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "Simple password",
			password: "password123",
		},
		{
			name:     "Complex password",
			password: "C0mpl3x!P@ssw0rd#2024",
		},
		{
			name:     "Long password",
			password: "ThisIsAVeryLongPasswordWithManyCharacters1234567890!@#$%^&*()",
		},
		{
			name:     "Password with spaces",
			password: "my secure password 123",
		},
		{
			name:     "Unicode password",
			password: "пароль密码🔒",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Hash password
			hash, err := HashPassword(tt.password)
			require.NoError(t, err, "HashPassword should not return error")
			assert.NotEmpty(t, hash, "Hash should not be empty")
			assert.NotEqual(t, tt.password, hash, "Hash should not equal plaintext password")
			assert.Greater(t, len(hash), 50, "bcrypt hash should be at least 50 characters")

			// Hash again to verify different salt
			hash2, err := HashPassword(tt.password)
			require.NoError(t, err)
			assert.NotEqual(t, hash, hash2, "Each hash should be unique (different salt)")

			// Both hashes should verify
			assert.True(t, CheckPasswordHash(tt.password, hash), "First hash should verify")
			assert.True(t, CheckPasswordHash(tt.password, hash2), "Second hash should verify")
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "MySecurePassword123!"
	hash, _ := HashPassword(password)

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "Correct password",
			password: password,
			hash:     hash,
			want:     true,
		},
		{
			name:     "Wrong password",
			password: "WrongPassword",
			hash:     hash,
			want:     false,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     hash,
			want:     false,
		},
		{
			name:     "Case sensitive check",
			password: "mysecurepassword123!",
			hash:     hash,
			want:     false,
		},
		{
			name:     "Password with extra space",
			password: password + " ",
			hash:     hash,
			want:     false,
		},
		{
			name:     "Invalid hash",
			password: password,
			hash:     "invalid_hash",
			want:     false,
		},
		{
			name:     "Empty hash",
			password: password,
			hash:     "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckPasswordHash(tt.password, tt.hash)
			assert.Equal(t, tt.want, result, "CheckPasswordHash result should match expected")
		})
	}
}

func TestGenerateSecureToken(t *testing.T) {
	// Test default length (32 bytes = 64 hex characters)
	token1, err := GenerateSecureToken(32)
	require.NoError(t, err, "GenerateSecureToken should not return error")
	assert.Equal(t, 64, len(token1), "32 bytes should produce 64 hex characters")

	// Test uniqueness
	token2, err := GenerateSecureToken(32)
	require.NoError(t, err)
	assert.NotEqual(t, token1, token2, "Generated tokens should be unique")

	// Test different lengths
	lengths := []int{16, 32, 64, 128}
	for _, length := range lengths {
		t.Run(fmt.Sprintf("Length_%d", length), func(t *testing.T) {
			token, err := GenerateSecureToken(length)
			require.NoError(t, err)
			assert.Equal(t, length*2, len(token), "Token length should be double the byte count")

			// Verify hex encoding (only contains 0-9, a-f)
			assert.Regexp(t, "^[0-9a-f]+$", token, "Token should be valid hex")
		})
	}

	// Test minimum entropy
	token, _ := GenerateSecureToken(16)
	assert.Greater(t, len(token), 20, "Token should have sufficient length for entropy")

	// Test that tokens are cryptographically random (basic check)
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		t, _ := GenerateSecureToken(16)
		assert.False(t, tokens[t], "Should not generate duplicate tokens")
		tokens[t] = true
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key := []byte("test-key-32-bytes-long-for-aes") // 30 bytes + padding
	plaintext := "Sensitive data to encrypt"

	// Encrypt
	encrypted, err := Encrypt(plaintext, key)
	require.NoError(t, err, "Encrypt should not return error")
	assert.NotEqual(t, plaintext, encrypted, "Encrypted text should differ from plaintext")

	// Decrypt
	decrypted, err := Decrypt(encrypted, key)
	require.NoError(t, err, "Decrypt should not return error")
	assert.Equal(t, plaintext, decrypted, "Decrypted text should match original plaintext")

	// Test with different plaintexts
	plaintexts := []string{
		"Short",
		"A longer string with multiple words",
		"Special chars: !@#$%^&*()",
		"Unicode: 你好世界🌍",
		"",
	}

	for _, pt := range plaintexts {
		t.Run(fmt.Sprintf("Text_%s", pt), func(t *testing.T) {
			enc, err := Encrypt(pt, key)
			require.NoError(t, err)

			dec, err := Decrypt(enc, key)
			require.NoError(t, err)
			assert.Equal(t, pt, dec)
		})
	}

	// Test encryption produces unique ciphertexts (due to random IV)
	enc1, _ := Encrypt(plaintext, key)
	enc2, _ := Encrypt(plaintext, key)
	assert.NotEqual(t, enc1, enc2, "Same plaintext should produce different ciphertexts")

	// Both should decrypt to same plaintext
	dec1, _ := Decrypt(enc1, key)
	dec2, _ := Decrypt(enc2, key)
	assert.Equal(t, plaintext, dec1)
	assert.Equal(t, plaintext, dec2)
}

func TestDecryptWithWrongKey(t *testing.T) {
	key1 := []byte("key1-32-bytes-long-for-aes-256")
	key2 := []byte("key2-32-bytes-long-for-aes-256")
	plaintext := "Secret message"

	// Encrypt with key1
	encrypted, err := Encrypt(plaintext, key1)
	require.NoError(t, err)

	// Try to decrypt with key2 (should fail)
	_, err = Decrypt(encrypted, key2)
	assert.Error(t, err, "Decrypt with wrong key should fail")
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	key := []byte("test-key-32-bytes-long-for-aes")

	tests := []struct {
		name       string
		ciphertext string
	}{
		{
			name:       "Empty string",
			ciphertext: "",
		},
		{
			name:       "Invalid base64",
			ciphertext: "not-valid-base64!@#",
		},
		{
			name:       "Too short ciphertext",
			ciphertext: "YWJj",
		},
		{
			name:       "Random garbage",
			ciphertext: "SGVsbG8gV29ybGQ=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decrypt(tt.ciphertext, key)
			assert.Error(t, err, "Decrypt should fail for invalid ciphertext")
		})
	}
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	// Empty password should still hash (though not recommended in practice)
	hash, err := HashPassword("")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Should verify
	assert.True(t, CheckPasswordHash("", hash))
}

func TestPasswordHashConsistency(t *testing.T) {
	password := "TestPassword123!"
	hash, _ := HashPassword(password)

	// Verify multiple times
	for i := 0; i < 10; i++ {
		assert.True(t, CheckPasswordHash(password, hash),
			"Hash should consistently verify the correct password")
	}
}
