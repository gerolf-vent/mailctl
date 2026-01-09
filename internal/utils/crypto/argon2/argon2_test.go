package argon2

import (
	"strings"
	"testing"
)

func TestGenerateFromPassword(t *testing.T) {
	password := []byte("password123")
	time := uint32(1)
	memory := uint32(64 * 1024)
	threads := uint8(4)
	keyLen := uint32(32)

	hash, err := GenerateFromPassword(password, time, memory, threads, keyLen)
	if err != nil {
		t.Fatalf("GenerateFromPassword failed: %v", err)
	}

	if len(hash) == 0 {
		t.Fatal("GenerateFromPassword returned empty hash")
	}

	hashStr := string(hash)
	if !strings.HasPrefix(hashStr, "$argon2id$") {
		t.Errorf("Hash does not start with correct prefix: %s", hashStr)
	}

	// Verify that generating a second time produces a different hash (salt check)
	hash2, err := GenerateFromPassword(password, time, memory, threads, keyLen)
	if err != nil {
		t.Fatalf("GenerateFromPassword failed: %v", err)
	}
	if string(hash) == string(hash2) {
		t.Error("GenerateFromPassword produced identical hashes for same password (salt missing?)")
	}
}

func TestCompareHashAndPassword(t *testing.T) {
	password := []byte("secret_password")
	time := uint32(1)
	memory := uint32(64 * 1024)
	threads := uint8(4)
	keyLen := uint32(32)

	hash, err := GenerateFromPassword(password, time, memory, threads, keyLen)
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	tests := []struct {
		name        string
		hash        []byte
		password    []byte
		expectError bool
		errType     error
	}{
		{
			name:        "Correct password",
			hash:        hash,
			password:    password,
			expectError: false,
		},
		{
			name:        "Incorrect password",
			hash:        hash,
			password:    []byte("wrong_password"),
			expectError: true,
			errType:     ErrMismatchedHashAndPassword,
		},
		{
			name:        "Empty hash",
			hash:        []byte(""),
			password:    password,
			expectError: true,
		},
		{
			name:        "Invalid prefix",
			hash:        []byte("!argon2id$v=19$m=65536,t=1,p=4$salt$hash"),
			password:    password,
			expectError: true,
		},
		{
			name:        "Invalid format (missing parts)",
			hash:        []byte("$argon2id$v=19$m=65536,t=1,p=4$salt"),
			password:    password,
			expectError: true,
		},
		{
			name:        "Invalid identifier",
			hash:        []byte("$argon2i$v=19$m=65536,t=1,p=4$salt$hash"),
			password:    password,
			expectError: true,
		},
		{
			name:        "Invalid version format",
			hash:        []byte("$argon2id$v=abc$m=65536,t=1,p=4$salt$hash"),
			password:    password,
			expectError: true,
		},
		{
			name:        "Unsupported version",
			hash:        []byte("$argon2id$v=99$m=65536,t=1,p=4$salt$hash"),
			password:    password,
			expectError: true,
		},
		{
			name:        "Invalid parameters format",
			hash:        []byte("$argon2id$v=19$m=65536,t=1$salt$hash"),
			password:    password,
			expectError: true,
		},
		{
			name:        "Invalid memory parameter",
			hash:        []byte("$argon2id$v=19$m=abc,t=1,p=4$salt$hash"),
			password:    password,
			expectError: true,
		},
		{
			name:        "Invalid salt encoding",
			hash:        []byte("$argon2id$v=19$m=65536,t=1,p=4$invalid_base64!$hash"),
			password:    password,
			expectError: true,
		},
		{
			name:        "Invalid hash encoding",
			hash:        []byte("$argon2id$v=19$m=65536,t=1,p=4$c2FsdA$invalid_base64!"),
			password:    password,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CompareHashAndPassword(tt.hash, tt.password)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				if tt.errType != nil && err != tt.errType {
					t.Errorf("Expected error %v, got %v", tt.errType, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}
