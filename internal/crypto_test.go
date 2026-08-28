/*
Copyright 2026 Joseph Anthony Abbott III

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeriveKey(t *testing.T) {
	tests := []struct {
		name     string
		password string
		salt     []byte
		wantLen  int
	}{
		{
			name:     "basic derivation",
			password: "testpassword",
			salt:     []byte("1234567890123456"),
			wantLen:  32,
		},
		{
			name:     "empty password",
			password: "",
			salt:     []byte("1234567890123456"),
			wantLen:  32,
		},
		{
			name:     "unicode password",
			password: "pässwörd🔐",
			salt:     []byte("1234567890123456"),
			wantLen:  32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := deriveKey(tt.password, tt.salt)
			if len(key) != tt.wantLen {
				t.Errorf("deriveKey() length = %d, want %d", len(key), tt.wantLen)
			}
			// Keys should be deterministic
			key2 := deriveKey(tt.password, tt.salt)
			for i, b := range key {
				if key2[i] != b {
					t.Errorf("deriveKey() not deterministic at byte %d", i)
					return
				}
			}
		})
	}
}

func TestEncryptDecryptFile(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.txt")
	originalContent := []byte("Hello, World! This is a test file.")

	if err := os.WriteFile(srcFile, originalContent, 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	tests := []struct {
		name     string
		password string
		dstPath  string
	}{
		{
			name:     "encrypt with default output",
			password: "testpassword",
			dstPath:  "",
		},
		{
			name:     "encrypt with custom output",
			password: "testpassword123",
			dstPath:  filepath.Join(tmpDir, "custom.enc"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encFile := tt.dstPath
			if encFile == "" {
				encFile = srcFile + encExt
			}

			// Encrypt
			if err := EncryptFile(srcFile, encFile, tt.password); err != nil {
				t.Fatalf("EncryptFile failed: %v", err)
			}

			// Verify encrypted file exists and is different
			encData, err := os.ReadFile(encFile)
			if err != nil {
				t.Fatalf("Failed to read encrypted file: %v", err)
			}
			if string(encData) == string(originalContent) {
				t.Error("Encrypted file matches original content")
			}

			// Decrypt
			decFile := filepath.Join(tmpDir, "decrypted.txt")
			if err := DecryptFile(encFile, decFile, tt.password); err != nil {
				t.Fatalf("DecryptFile failed: %v", err)
			}

			// Verify decrypted content matches original
			decData, err := os.ReadFile(decFile)
			if err != nil {
				t.Fatalf("Failed to read decrypted file: %v", err)
			}
			if string(decData) != string(originalContent) {
				t.Errorf("Decrypted content mismatch: got %q, want %q", string(decData), string(originalContent))
			}
		})
	}
}

func TestDecryptWithWrongPassword(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.txt")
	originalContent := []byte("Secret data")

	if err := os.WriteFile(srcFile, originalContent, 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	encFile := filepath.Join(tmpDir, "test.enc")
	if err := EncryptFile(srcFile, encFile, "correctpassword"); err != nil {
		t.Fatalf("EncryptFile failed: %v", err)
	}

	decFile := filepath.Join(tmpDir, "decrypted.txt")
	err := DecryptFile(encFile, decFile, "wrongpassword")
	if err == nil {
		t.Error("DecryptFile should have failed with wrong password")
	}
}

func TestGenerateKey(t *testing.T) {
	tests := []struct {
		name    string
		wantLen int
	}{
		{
			name:    "generate valid key",
			wantLen: 32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := GenerateKey()
			if err != nil {
				t.Fatalf("GenerateKey failed: %v", err)
			}
			if len(key) != tt.wantLen {
				t.Errorf("GenerateKey() length = %d, want %d", len(key), tt.wantLen)
			}

			// Keys should be random
			key2, err := GenerateKey()
			if err != nil {
				t.Fatalf("GenerateKey failed: %v", err)
			}
			if string(key) == string(key2) {
				t.Error("GenerateKey() produced identical keys")
			}
		})
	}
}

func TestEncryptDecryptKeyBlob(t *testing.T) {
	tests := []struct {
		name         string
		rawKey       []byte
		masterPass   string
	}{
		{
			name:       "basic key encryption",
			rawKey:     []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			masterPass: "masterpass123",
		},
		{
			name:       "32-byte key",
			rawKey:     make([]byte, 32),
			masterPass: "anotherpass",
		},
		{
			name:       "unicode master password",
			rawKey:     []byte("some_key_data"),
			masterPass: "pässwörd🔐",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			blob, err := EncryptKeyBlob(tt.rawKey, tt.masterPass)
			if err != nil {
				t.Fatalf("EncryptKeyBlob failed: %v", err)
			}
			if blob == "" {
				t.Error("EncryptKeyBlob returned empty string")
			}

			// Decrypt
			decrypted, err := DecryptKeyBlob(blob, tt.masterPass)
			if err != nil {
				t.Fatalf("DecryptKeyBlob failed: %v", err)
			}
			if string(decrypted) != string(tt.rawKey) {
				t.Errorf("Decrypted key mismatch: got %v, want %v", decrypted, tt.rawKey)
			}
		})
	}
}

func TestDecryptKeyBlobWithWrongPassword(t *testing.T) {
	rawKey := []byte{1, 2, 3, 4, 5}
	masterPass := "correctpass"

	blob, err := EncryptKeyBlob(rawKey, masterPass)
	if err != nil {
		t.Fatalf("EncryptKeyBlob failed: %v", err)
	}

	_, err = DecryptKeyBlob(blob, "wrongpass")
	if err == nil {
		t.Error("DecryptKeyBlob should have failed with wrong password")
	}
}

func TestEncryptDecryptKeyBlobRandomness(t *testing.T) {
	rawKey := []byte{1, 2, 3}
	masterPass := "pass"

	blob1, err := EncryptKeyBlob(rawKey, masterPass)
	if err != nil {
		t.Fatalf("EncryptKeyBlob failed: %v", err)
	}

	blob2, err := EncryptKeyBlob(rawKey, masterPass)
	if err != nil {
		t.Fatalf("EncryptKeyBlob failed: %v", err)
	}

	// Blobs should be different due to random IV and salt
	if blob1 == blob2 {
		t.Error("EncryptKeyBlob produced identical blobs for same input")
	}
}

func TestNewKeyID(t *testing.T) {
	id1, err := newKeyID()
	if err != nil {
		t.Fatalf("newKeyID failed: %v", err)
	}
	if len(id1) != 16 { // 8 bytes in hex
		t.Errorf("newKeyID length = %d, want 16", len(id1))
	}

	id2, err := newKeyID()
	if err != nil {
		t.Fatalf("newKeyID failed: %v", err)
	}

	// IDs should be unique
	if id1 == id2 {
		t.Error("newKeyID generated identical IDs")
	}
}
