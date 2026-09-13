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
	"strings"
	"testing"
)

// TestDeriveKey tests the deriveKey function for various scenarios including basic derivation, empty password, and unicode password.
// It verifies that the derived key has the expected length and is deterministic for the same inputs.
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

// TestEncryptDecryptFile tests the EncryptFile and DecryptFile functions by encrypting and decrypting a file with various passwords and output paths.
// It ensures that the encrypted file differs from the original and that decryption restores the original content accurately.
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

// TestEncryptDecryptDirectory verifies directory encryption produces one
// encrypted artifact and decrypts back into the original directory structure.
func TestEncryptDecryptDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "bundle")
	if err := os.MkdirAll(filepath.Join(srcDir, "nested"), 0o755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(srcDir, "empty"), 0o755); err != nil {
		t.Fatalf("Failed to create empty directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "root.txt"), []byte("root data"), 0o644); err != nil {
		t.Fatalf("Failed to write root file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "nested", "child.txt"), []byte("nested data"), 0o640); err != nil {
		t.Fatalf("Failed to write nested file: %v", err)
	}

	encFile := srcDir + encExt
	if err := EncryptFile(srcDir, "", "testpassword"); err != nil {
		t.Fatalf("EncryptFile failed: %v", err)
	}

	info, err := os.Stat(encFile)
	if err != nil {
		t.Fatalf("Encrypted artifact not found: %v", err)
	}
	if info.IsDir() {
		t.Fatal("Encrypted artifact should be a single file")
	}

	restoreDir := filepath.Join(tmpDir, "restored")
	if err := DecryptFile(encFile, restoreDir, "testpassword"); err != nil {
		t.Fatalf("DecryptFile failed: %v", err)
	}

	rootData, err := os.ReadFile(filepath.Join(restoreDir, "root.txt"))
	if err != nil {
		t.Fatalf("Failed to read restored root file: %v", err)
	}
	if string(rootData) != "root data" {
		t.Fatalf("Root file mismatch: got %q", string(rootData))
	}

	nestedData, err := os.ReadFile(filepath.Join(restoreDir, "nested", "child.txt"))
	if err != nil {
		t.Fatalf("Failed to read restored nested file: %v", err)
	}
	if string(nestedData) != "nested data" {
		t.Fatalf("Nested file mismatch: got %q", string(nestedData))
	}
	if info, err := os.Stat(filepath.Join(restoreDir, "nested", "child.txt")); err != nil {
		t.Fatalf("Failed to stat restored nested file: %v", err)
	} else if info.Mode().Perm() != 0o640 {
		t.Fatalf("Nested file mode mismatch: got %o, want %o", info.Mode().Perm(), 0o640)
	}

	if info, err := os.Stat(filepath.Join(restoreDir, "empty")); err != nil {
		t.Fatalf("Expected empty directory not restored: %v", err)
	} else if !info.IsDir() {
		t.Fatal("Restored empty path is not a directory")
	}
}

func TestEncryptEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	emptyDir := filepath.Join(tmpDir, "empty")
	if err := os.MkdirAll(emptyDir, 0o755); err != nil {
		t.Fatalf("Failed to create empty directory: %v", err)
	}

	err := EncryptFile(emptyDir, "", "testpassword")
	if err == nil {
		t.Fatal("EncryptFile should fail for an empty directory")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Fatalf("Expected empty directory error, got %v", err)
	}
}

func TestEncryptDirectoryWithSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "bundle")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "root.txt"), []byte("data"), 0o644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	linkPath := filepath.Join(srcDir, "linked.txt")
	if err := os.Symlink(filepath.Join(srcDir, "root.txt"), linkPath); err != nil {
		t.Skipf("Symlinks are unavailable in this environment: %v", err)
	}

	err := EncryptFile(srcDir, "", "testpassword")
	if err == nil {
		t.Fatal("EncryptFile should fail when the directory contains a symlink")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "symlink") {
		t.Fatalf("Expected symlink error, got %v", err)
	}
}

func TestEncryptSymlinkedDirectoryPath(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "target")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "file.txt"), []byte("data"), 0o644); err != nil {
		t.Fatalf("Failed to write target file: %v", err)
	}

	linkDir := filepath.Join(tmpDir, "linked-target")
	if err := os.Symlink(targetDir, linkDir); err != nil {
		t.Skipf("Symlinks are unavailable in this environment: %v", err)
	}

	err := EncryptFile(linkDir, "", "testpassword")
	if err == nil {
		t.Fatal("EncryptFile should fail when the source directory path is a symlink")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "symlink") {
		t.Fatalf("Expected symlink error, got %v", err)
	}
}

func TestDecryptLegacyEncryptedFile(t *testing.T) {
	tmpDir := t.TempDir()
	originalContent := []byte("legacy encrypted file data")

	encFile := filepath.Join(tmpDir, "legacy.txt.enc")
	legacyCiphertext, err := encryptBytes(originalContent, "testpassword")
	if err != nil {
		t.Fatalf("encryptBytes failed: %v", err)
	}
	if err := os.WriteFile(encFile, legacyCiphertext, 0o600); err != nil {
		t.Fatalf("Failed to write legacy ciphertext: %v", err)
	}

	decFile := filepath.Join(tmpDir, "legacy.txt.dec")
	if err := DecryptFile(encFile, decFile, "testpassword"); err != nil {
		t.Fatalf("DecryptFile failed: %v", err)
	}

	decData, err := os.ReadFile(decFile)
	if err != nil {
		t.Fatalf("Failed to read decrypted file: %v", err)
	}
	if string(decData) != string(originalContent) {
		t.Fatalf("Decrypted content mismatch: got %q, want %q", string(decData), string(originalContent))
	}
}

// TestDecryptWithWrongPassword tests that attempting to decrypt a file with an incorrect password fails as expected.
// It ensures that the decryption process returns an error and does not produce the original content.
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

// TestGenerateKey tests the GenerateKey function to ensure it produces keys of the expected length and that consecutive calls produce different keys.
// It verifies that the generated keys have the correct length and that multiple calls produce unique keys.
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
		name       string
		rawKey     []byte
		masterPass string
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

// TestDecryptKeyBlobWithWrongPassword tests that attempting to decrypt an encrypted key blob with an incorrect password fails as expected.
// It ensures that the decryption process returns an error and does not produce the original key.
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

// TestEncryptDecryptKeyBlobRandomness tests that encrypting the same key blob multiple times with the same password produces different encrypted outputs.
// This ensures that the encryption process uses random IVs and salts to enhance security.
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

// TestNewKeyID tests the newKeyID function to ensure it generates unique IDs of the expected length.
// It verifies that consecutive calls produce different IDs and that the length of the generated ID matches the expected value.
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
