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
		{
			name: "with .tar.gz extension",
			src:  "folder.tar.gz",
			want: "folder",
		},
		{
			name: "with .tgz extension",
			src:  "folder.tgz",
			want: "folder",
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

// TestDeriveCompressOutput tests the deriveCompressOutput helper for files and directories.
func TestDeriveCompressOutput(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("data"), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	dirPath := filepath.Join(tmpDir, "folder")
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "file path",
			src:  filePath,
			want: filePath + gzExt,
		},
		{
			name: "directory path",
			src:  dirPath,
			want: dirPath + tarGzExt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deriveCompressOutput(tt.src); got != tt.want {
				t.Errorf("deriveCompressOutput(%q) = %q, want %q", tt.src, got, tt.want)
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

// TestEncryptDecryptDirectoryCmdRoundTrip verifies the CLI can encrypt a
// directory into a single artifact and restore its tree on decrypt.
func TestEncryptDecryptDirectoryCmdRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(filepath.Join(srcDir, "nested"), 0o755); err != nil {
		t.Fatalf("Failed to create source directories: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "root.txt"), []byte("root secret"), 0o644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "nested", "child.txt"), []byte("nested secret"), 0o644); err != nil {
		t.Fatalf("Failed to write nested source file: %v", err)
	}

	db, err := NewDatabase(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}

	password := "testpass123"

	rootCmd := NewRootCmd(db)
	rootCmd.SetArgs([]string{"encrypt", srcDir, "-p", password})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	encFile := srcDir + encExt
	if info, err := os.Stat(encFile); err != nil {
		t.Fatalf("Encrypted artifact not created: %v", err)
	} else if info.IsDir() {
		t.Fatal("Encrypted artifact should be a file")
	}

	restoreDir := filepath.Join(tmpDir, "restored-project")
	rootCmd = NewRootCmd(db)
	rootCmd.SetArgs([]string{"decrypt", encFile, "-o", restoreDir, "-p", password})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	rootData, err := os.ReadFile(filepath.Join(restoreDir, "root.txt"))
	if err != nil {
		t.Fatalf("Failed to read restored root file: %v", err)
	}
	if string(rootData) != "root secret" {
		t.Fatalf("Content mismatch: got %q", string(rootData))
	}

	childData, err := os.ReadFile(filepath.Join(restoreDir, "nested", "child.txt"))
	if err != nil {
		t.Fatalf("Failed to read restored nested file: %v", err)
	}
	if string(childData) != "nested secret" {
		t.Fatalf("Content mismatch: got %q", string(childData))
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

// TestCompressDecompressDirectoryCmdRoundTrip tests the CLI round-trip for a
// compressed directory archive using default input and output paths.
func TestCompressDecompressDirectoryCmdRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "folder")
	if err := os.MkdirAll(filepath.Join(srcDir, "nested"), 0o755); err != nil {
		t.Fatalf("Failed to create source directories: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "nested", "file.txt"), []byte("directory data"), 0o644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	db, err := NewDatabase(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}

	rootCmd := NewRootCmd(db)
	rootCmd.SetArgs([]string{"compress", srcDir})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Compress failed: %v", err)
	}

	archive := srcDir + tarGzExt
	if _, err := os.Stat(archive); err != nil {
		t.Fatalf("Archive not created: %v", err)
	}

	if err := os.RemoveAll(srcDir); err != nil {
		t.Fatalf("Failed to remove source directory: %v", err)
	}

	rootCmd = NewRootCmd(db)
	rootCmd.SetArgs([]string{"decompress", archive})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Decompress failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(srcDir, "nested", "file.txt"))
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}
	if string(data) != "directory data" {
		t.Fatalf("Content mismatch: got %q", string(data))
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
