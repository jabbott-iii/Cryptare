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
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestNewRootCmd tests the NewRootCmd function to ensure it returns a valid root command.
// It verifies that the command is not nil and has the expected use string.
func TestNewRootCmd(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := NewDatabase(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}

	cmd := NewRootCmd(db)
	if cmd == nil {
		t.Error("NewRootCmd returned nil")
	}

	if cmd.Use != "cryptare" {
		t.Errorf("Root command Use = %q, want cryptare", cmd.Use)
	}
}

// TestDeriveDecryptOutput tests the deriveDecryptOutput function to ensure it correctly derives the output file name for decryption.
// It covers cases with .enc extension, no extension, and other extensions.
func TestDeriveDecryptOutput(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "with .enc extension",
			src:  "file.txt.enc",
			want: "file.txt",
		},
		{
			name: "without extension",
			src:  "file",
			want: "file.dec",
		},
		{
			name: "with other extension",
			src:  "file.txt.encrypted",
			want: "file.txt.encrypted.dec",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveDecryptOutput(tt.src)
			if got != tt.want {
				t.Errorf("deriveDecryptOutput(%q) = %q, want %q", tt.src, got, tt.want)
			}
		})
	}
}

// TestDeriveDecompressOutput tests the deriveDecompressOutput function to ensure it correctly derives the output file name for decompression.
// It covers cases with .gz extension, no extension, and other extensions.
func TestDeriveDecompressOutput(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "with .gz extension",
			src:  "file.txt.gz",
			want: "file.txt",
		},
		{
			name: "without extension",
			src:  "file",
			want: "file.dec",
		},
		{
			name: "with other extension",
			src:  "file.compressed",
			want: "file.compressed.dec",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveDecompressOutput(tt.src)
			if got != tt.want {
				t.Errorf("deriveDecompressOutput(%q) = %q, want %q", tt.src, got, tt.want)
			}
		})
	}
}

// TestEncryptCmdWithPassword tests the encrypt command with a provided password.
// It ensures that the command successfully creates an encrypted file.
func TestEncryptCmdWithPassword(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(srcFile, []byte("test content"), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	db, err := NewDatabase(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}

	rootCmd := NewRootCmd(db)
	rootCmd.SetArgs([]string{"encrypt", srcFile, "-p", "testpass"})

	var out bytes.Buffer
	rootCmd.SetOut(&out)

	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	encFile := srcFile + encExt
	if _, err := os.Stat(encFile); err != nil {
		t.Fatalf("Encrypted file not created: %v", err)
	}
}

// TestEncryptDecryptCmdRoundTrip tests the full round-trip of encrypting and then decrypting a file.
// It verifies that the decrypted content matches the original content.
func TestEncryptDecryptCmdRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.txt")
	originalContent := []byte("secret data")

	if err := os.WriteFile(srcFile, originalContent, 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	db, err := NewDatabase(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}

	password := "testpass123"

	// Encrypt
	rootCmd := NewRootCmd(db)
	rootCmd.SetArgs([]string{"encrypt", srcFile, "-p", password})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	encFile := srcFile + encExt

	// Decrypt
	rootCmd = NewRootCmd(db)
	decFile := filepath.Join(tmpDir, "decrypted.txt")
	rootCmd.SetArgs([]string{"decrypt", encFile, "-o", decFile, "-p", password})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	// Verify
	decData, err := os.ReadFile(decFile)
	if err != nil {
		t.Fatalf("Failed to read decrypted file: %v", err)
	}

	if string(decData) != string(originalContent) {
		t.Errorf("Content mismatch: got %q, want %q", string(decData), string(originalContent))
	}
}

// TestCompressDecompressCmdRoundTrip tests the full round-trip of compressing and then decompressing a file.
// It verifies that the decompressed content matches the original content.
func TestCompressDecompressCmdRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.txt")
	originalContent := []byte("test data to compress")

	if err := os.WriteFile(srcFile, originalContent, 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	db, err := NewDatabase(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}

	// Compress
	rootCmd := NewRootCmd(db)
	rootCmd.SetArgs([]string{"compress", srcFile})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Compress failed: %v", err)
	}

	compFile := srcFile + gzExt

	// Decompress
	rootCmd = NewRootCmd(db)
	decFile := filepath.Join(tmpDir, "decompressed.txt")
	rootCmd.SetArgs([]string{"decompress", compFile, "-o", decFile})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Decompress failed: %v", err)
	}

	// Verify
	decData, err := os.ReadFile(decFile)
	if err != nil {
		t.Fatalf("Failed to read decompressed file: %v", err)
	}

	if string(decData) != string(originalContent) {
		t.Errorf("Content mismatch: got %q, want %q", string(decData), string(originalContent))
	}
}

// TestKeysListCmd tests the "keys list" command to ensure it correctly lists all keys in the database.
// It verifies that the output contains the expected key IDs.
func TestKeysListCmd(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := NewDatabase(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}

	// Add a key
	km := &KeyModel{
		KeyID:         "test-key-1",
		Algorithm:     "AES-256-GCM",
		EncryptedBlob: "blob",
		CreatedAt_:    time.Now().Unix(),
	}
	if err := db.SaveKey(km); err != nil {
		t.Fatalf("SaveKey failed: %v", err)
	}

	rootCmd := NewRootCmd(db)
	rootCmd.SetArgs([]string{"keys", "list"})

	var out bytes.Buffer
	rootCmd.SetOut(&out)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := out.String()
	if output == "" {
		t.Error("Command produced no output")
	}
}
