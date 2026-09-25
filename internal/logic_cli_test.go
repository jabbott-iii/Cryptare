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
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// TestNewRootCmd tests the NewRootCmd function to ensure it returns a valid root command.
// It verifies that the command is not nil and has the expected use string.
func TestNewRootCmd(t *testing.T) {
	db := newTestDatabase(t, false)

	cmd := NewRootCmd(db)
	if cmd.Use != "cryptare" {
		t.Errorf("Root command Use = %q, want cryptare", cmd.Use)
	}

	flag := cmd.Flags().Lookup("vim")
	if flag == nil {
		t.Fatal("expected --vim flag to be registered on root command")
	}
	if flag.DefValue != "false" {
		t.Fatalf("--vim default = %q, want false", flag.DefValue)
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
		{
			name: "with .zip extension",
			src:  "folder.zip",
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
		name   string
		src    string
		format string
		want   string
	}{
		{
			name:   "file path",
			src:    filePath,
			format: "",
			want:   filePath + gzExt,
		},
		{
			name:   "directory path",
			src:    dirPath,
			format: "",
			want:   dirPath + tarGzExt,
		},
		{
			name:   "zip file path",
			src:    filePath,
			format: "zip",
			want:   filePath + zipExt,
		},
		{
			name:   "zip directory path",
			src:    dirPath,
			format: "zip",
			want:   dirPath + zipExt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deriveCompressOutput(tt.src, tt.format); got != tt.want {
				t.Errorf("deriveCompressOutput(%q, %q) = %q, want %q", tt.src, tt.format, got, tt.want)
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

	db := newTestDatabase(t, false)

	rootCmd := NewRootCmd(db)
	rootCmd.SetArgs([]string{"encrypt", srcFile, "-p", "testpass"})

	var out bytes.Buffer
	rootCmd.SetOut(&out)

	err := rootCmd.Execute()
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

	db := newTestDatabase(t, false)

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
	if err := os.MkdirAll(filepath.Join(srcDir, "empty"), 0o755); err != nil {
		t.Fatalf("Failed to create empty source directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "root.txt"), []byte("root secret"), 0o644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "nested", "child.txt"), []byte("nested secret"), 0o644); err != nil {
		t.Fatalf("Failed to write nested source file: %v", err)
	}

	db := newTestDatabase(t, false)

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
	if info, err := os.Stat(filepath.Join(restoreDir, "empty")); err != nil {
		t.Fatalf("Expected empty directory not restored: %v", err)
	} else if !info.IsDir() {
		t.Fatal("Restored empty path is not a directory")
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

	db := newTestDatabase(t, false)

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

	db := newTestDatabase(t, false)

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

func TestCompressDecompressZipCmdRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "artifact.txt")
	originalContent := []byte("zip file content")
	if err := os.WriteFile(srcFile, originalContent, 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	db := newTestDatabase(t, false)

	rootCmd := NewRootCmd(db)
	rootCmd.SetArgs([]string{"compress", srcFile, "--format", "zip"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Compress failed: %v", err)
	}

	archive := srcFile + zipExt
	if _, err := os.Stat(archive); err != nil {
		t.Fatalf("Zip archive not created: %v", err)
	}

	if err := os.Remove(srcFile); err != nil {
		t.Fatalf("Failed to remove source file: %v", err)
	}

	rootCmd = NewRootCmd(db)
	rootCmd.SetArgs([]string{"decompress", archive})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Decompress failed: %v", err)
	}

	restored, err := os.ReadFile(srcFile)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}
	if string(restored) != string(originalContent) {
		t.Fatalf("Content mismatch: got %q", string(restored))
	}
}

func TestCompressDecompressZipDirectoryCmdRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "folder")
	if err := os.MkdirAll(filepath.Join(srcDir, "nested"), 0o755); err != nil {
		t.Fatalf("Failed to create source directories: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "nested", "file.txt"), []byte("zip directory data"), 0o644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(srcDir, "empty"), 0o755); err != nil {
		t.Fatalf("Failed to create empty source directory: %v", err)
	}

	db := newTestDatabase(t, false)

	rootCmd := NewRootCmd(db)
	rootCmd.SetArgs([]string{"compress", srcDir, "--format", "zip"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Compress failed: %v", err)
	}

	archive := srcDir + zipExt
	if _, err := os.Stat(archive); err != nil {
		t.Fatalf("Zip archive not created: %v", err)
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
	if string(data) != "zip directory data" {
		t.Fatalf("Content mismatch: got %q", string(data))
	}
	if info, err := os.Stat(filepath.Join(srcDir, "empty")); err != nil {
		t.Fatalf("Expected empty directory not restored: %v", err)
	} else if !info.IsDir() {
		t.Fatal("Restored empty path is not a directory")
	}
}

// TestKeysListCmd tests the "keys list" command to ensure it correctly lists all keys in the database.
// It verifies that the output contains the expected key IDs.
func TestKeysListCmd(t *testing.T) {
	db := newTestDatabase(t, false)

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

func TestKeysDeleteCmdConfirmed(t *testing.T) {
	db := newTestDatabase(t, false)

	km := &KeyModel{
		KeyID:         "delete-me",
		Algorithm:     "AES-256-GCM",
		EncryptedBlob: "blob",
		CreatedAt_:    time.Now().Unix(),
	}
	if err := db.SaveKey(km); err != nil {
		t.Fatalf("SaveKey failed: %v", err)
	}

	rootCmd := NewRootCmd(db)
	rootCmd.SetArgs([]string{"keys", "delete", km.KeyID})
	rootCmd.SetIn(bytes.NewBufferString("y\n"))

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	if _, err := db.GetKey(km.KeyID); err == nil {
		t.Fatal("expected key to be deleted")
	}
	if !strings.Contains(out.String(), "Deleted key: "+km.KeyID) {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestKeysDeleteCmdRequiresConfirmation(t *testing.T) {
	db := newTestDatabase(t, false)

	km := &KeyModel{
		KeyID:         "do-not-delete",
		Algorithm:     "AES-256-GCM",
		EncryptedBlob: "blob",
		CreatedAt_:    time.Now().Unix(),
	}
	if err := db.SaveKey(km); err != nil {
		t.Fatalf("SaveKey failed: %v", err)
	}

	rootCmd := NewRootCmd(db)
	rootCmd.SetArgs([]string{"keys", "delete", km.KeyID})
	rootCmd.SetIn(bytes.NewBufferString("n\n"))

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation rejection error")
	}
	if !strings.Contains(err.Error(), "aborted") {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := db.GetKey(km.KeyID); err != nil {
		t.Fatalf("expected key to remain after abort, got: %v", err)
	}
}

func TestKeysDeleteCmdForceBypass(t *testing.T) {
	db := newTestDatabase(t, false)

	km := &KeyModel{
		KeyID:         "force-delete",
		Algorithm:     "AES-256-GCM",
		EncryptedBlob: "blob",
		CreatedAt_:    time.Now().Unix(),
	}
	if err := db.SaveKey(km); err != nil {
		t.Fatalf("SaveKey failed: %v", err)
	}

	rootCmd := NewRootCmd(db)
	rootCmd.SetArgs([]string{"keys", "delete", km.KeyID, "--force"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	if _, err := db.GetKey(km.KeyID); err == nil {
		t.Fatal("expected key to be deleted")
	}
}

func TestKeysDeleteCmdYesBypass(t *testing.T) {
	db := newTestDatabase(t, false)

	km := &KeyModel{
		KeyID:         "yes-delete",
		Algorithm:     "AES-256-GCM",
		EncryptedBlob: "blob",
		CreatedAt_:    time.Now().Unix(),
	}
	if err := db.SaveKey(km); err != nil {
		t.Fatalf("SaveKey failed: %v", err)
	}

	rootCmd := NewRootCmd(db)
	rootCmd.SetArgs([]string{"keys", "delete", km.KeyID, "--yes"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	if _, err := db.GetKey(km.KeyID); err == nil {
		t.Fatal("expected key to be deleted")
	}
}

func TestKeysDeleteCmdNotFound(t *testing.T) {
	db := newTestDatabase(t, false)

	rootCmd := NewRootCmd(db)
	rootCmd.SetArgs([]string{"keys", "delete", "missing-key", "--yes"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected not found error")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestEncryptCmdRejectsEmptyInteractivePassword guards SEC-001 on the CLI: an empty
// answer at the password prompt must not produce an encrypted file.
func TestEncryptCmdRejectsEmptyInteractivePassword(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(srcFile, []byte("test content"), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	db := newTestDatabase(t, false)
	rootCmd := NewRootCmd(db)
	rootCmd.SetArgs([]string{"encrypt", srcFile})
	rootCmd.SetIn(strings.NewReader("\n"))
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)

	if err := rootCmd.Execute(); err == nil {
		t.Fatal("encrypt with an empty password succeeded, want an error")
	}
	if _, err := os.Stat(srcFile + encExt); !os.IsNotExist(err) {
		t.Fatalf("encrypted file created despite the empty password (stat err: %v)", err)
	}
}

// TestReadPassword is a regression test for SEC-002: when input is not a terminal,
// the prompt must keep the whole line, including spaces, and strip only the line ending.
func TestReadPassword(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "multi-word passphrase", input: "correct horse battery staple\n", want: "correct horse battery staple"},
		{name: "windows line ending", input: "pass word\r\n", want: "pass word"},
		{name: "leading and trailing spaces kept", input: "  spaced out  \n", want: "  spaced out  "},
		{name: "no trailing newline", input: "secret", want: "secret"},
		{name: "only the first line is read", input: "first line\nsecond line\n", want: "first line"},
		{name: "empty line", input: "\n", want: ""},
		{name: "no input", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.SetIn(strings.NewReader(tt.input))
			var prompt bytes.Buffer
			cmd.SetErr(&prompt)

			got, err := readPassword(cmd, "Enter password: ")
			if tt.wantErr {
				if err == nil {
					t.Fatalf("readPassword() = %q, want an error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("readPassword() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("readPassword() = %q, want %q", got, tt.want)
			}
			if prompt.String() != "Enter password: " {
				t.Fatalf("prompt written = %q, want %q", prompt.String(), "Enter password: ")
			}
		})
	}
}

// TestEncryptDecryptCmdMultiWordPromptPassword is a regression test for SEC-002: a
// passphrase typed at the prompt must be used in full, not cut at the first space.
func TestEncryptDecryptCmdMultiWordPromptPassword(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("multi-word passphrase content")
	if err := os.WriteFile(srcFile, content, 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	const passphrase = "correct horse battery staple"
	db := newTestDatabase(t, false)

	run := func(stdin string, args ...string) error {
		rootCmd := NewRootCmd(db)
		rootCmd.SetArgs(args)
		rootCmd.SetIn(strings.NewReader(stdin))
		var out bytes.Buffer
		rootCmd.SetOut(&out)
		rootCmd.SetErr(&out)
		return rootCmd.Execute()
	}

	encFile := srcFile + encExt
	if err := run(passphrase+"\n", "encrypt", srcFile); err != nil {
		t.Fatalf("encrypt via prompt failed: %v", err)
	}
	if err := run("", "decrypt", encFile, "--output", filepath.Join(tmpDir, "first-word.txt"), "--password", "correct"); err == nil {
		t.Fatal("decrypting with only the first word succeeded; the prompt truncated the passphrase")
	}
	decFile := filepath.Join(tmpDir, "decrypted.txt")
	if err := run(passphrase+"\n", "decrypt", encFile, "--output", decFile); err != nil {
		t.Fatalf("decrypt via prompt with the full passphrase failed: %v", err)
	}
	if got, err := os.ReadFile(decFile); err != nil || string(got) != string(content) {
		t.Fatalf("decrypted content = %q (err %v), want %q", got, err, content)
	}
}

// TestDecryptCmdPromptAcceptsEmptyPasswordForLegacyFiles checks that a file encrypted
// with an empty password before SEC-001 was fixed can be decrypted from the CLI prompt.
func TestDecryptCmdPromptAcceptsEmptyPasswordForLegacyFiles(t *testing.T) {
	tmpDir := t.TempDir()
	content := []byte("legacy content")
	ciphertext, err := encryptBytes(content, "")
	if err != nil {
		t.Fatalf("encryptBytes: %v", err)
	}
	encFile := filepath.Join(tmpDir, "legacy.txt.enc")
	if err := os.WriteFile(encFile, ciphertext, 0o600); err != nil {
		t.Fatalf("write legacy file: %v", err)
	}

	decFile := filepath.Join(tmpDir, "legacy.txt")
	rootCmd := NewRootCmd(newTestDatabase(t, false))
	rootCmd.SetArgs([]string{"decrypt", encFile, "--output", decFile})
	rootCmd.SetIn(strings.NewReader("\n"))
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("decrypt with an empty prompt password failed: %v", err)
	}
	if got, err := os.ReadFile(decFile); err != nil || string(got) != string(content) {
		t.Fatalf("decrypted content = %q (err %v), want %q", got, err, content)
	}
}

// TestReadPasswordFromNonTerminalFile checks that input from a real file that isn't a
// terminal (as with shell redirection) is read as a plain line, not through the
// hidden terminal prompt.
func TestReadPasswordFromNonTerminalFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "password.txt")
	if err := os.WriteFile(path, []byte("piped pass phrase\n"), 0o600); err != nil {
		t.Fatalf("write password file: %v", err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open password file: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })

	cmd := &cobra.Command{}
	cmd.SetIn(f)
	cmd.SetErr(&bytes.Buffer{})
	got, err := readPassword(cmd, "Enter password: ")
	if err != nil {
		t.Fatalf("readPassword() error = %v", err)
	}
	if got != "piped pass phrase" {
		t.Fatalf("readPassword() = %q, want %q", got, "piped pass phrase")
	}
}

// TestReadTerminalPasswordRejectsNonTerminal checks that the hidden-input helper
// fails cleanly, before installing its interrupt handler, when fd is not a terminal.
func TestReadTerminalPasswordRejectsNonTerminal(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "not-a-terminal")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })

	if _, err := readTerminalPassword(f.Fd(), &bytes.Buffer{}); err == nil {
		t.Fatal("readTerminalPassword() on a regular file succeeded, want an error")
	}
}

// TestFileCmdsRefuseExistingOutput is a regression test for BUG-004: encrypt, decrypt,
// compress and decompress must not replace an existing output unless --force is given.
func TestFileCmdsRefuseExistingOutput(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "data.txt")
	if err := os.WriteFile(src, []byte("payload"), 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}
	enc := src + encExt
	if err := EncryptFile(src, enc, "pw"); err != nil {
		t.Fatalf("prepare encrypted file: %v", err)
	}
	gz := src + gzExt
	if err := CompressFile(src, gz, -1); err != nil {
		t.Fatalf("prepare gzip file: %v", err)
	}
	db := newTestDatabase(t, false)

	run := func(args ...string) error {
		rootCmd := NewRootCmd(db)
		rootCmd.SetArgs(args)
		rootCmd.SetIn(strings.NewReader(""))
		var out bytes.Buffer
		rootCmd.SetOut(&out)
		rootCmd.SetErr(&out)
		return rootCmd.Execute()
	}

	cases := []struct {
		name string
		args func(out string) []string
	}{
		{"encrypt", func(out string) []string { return []string{"encrypt", src, "--output", out, "--password", "pw"} }},
		{"decrypt", func(out string) []string { return []string{"decrypt", enc, "--output", out, "--password", "pw"} }},
		{"compress", func(out string) []string { return []string{"compress", src, "--output", out} }},
		{"decompress", func(out string) []string { return []string{"decompress", gz, "--output", out} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := filepath.Join(tmpDir, tc.name+".out")
			if err := os.WriteFile(out, []byte("existing"), 0o600); err != nil {
				t.Fatalf("write existing output: %v", err)
			}

			err := run(tc.args(out)...)
			if !errors.Is(err, ErrOutputExists) || !strings.Contains(err.Error(), "--force") {
				t.Fatalf("error = %v, want ErrOutputExists with a --force hint", err)
			}
			if got, _ := os.ReadFile(out); string(got) != "existing" {
				t.Fatalf("existing output was modified to %q", got)
			}

			if err := run(append(tc.args(out), "--force")...); err != nil {
				t.Fatalf("with --force: %v", err)
			}
			if got, _ := os.ReadFile(out); string(got) == "existing" {
				t.Fatal("with --force the output was not replaced")
			}
		})
	}

	// Default output: decrypting data.txt.enc would write data.txt, which still exists.
	if err := run("decrypt", enc, "--password", "pw"); !errors.Is(err, ErrOutputExists) {
		t.Fatalf("decrypt to existing default output: error = %v, want ErrOutputExists", err)
	}
	if got, _ := os.ReadFile(src); string(got) != "payload" {
		t.Fatalf("default output was modified to %q", got)
	}
}

// TestCompressCmdRefusesSameInputOutputEvenWithForce is a regression test for
// BUG-003: --force must not allow an output that is the input itself.
func TestCompressCmdRefusesSameInputOutputEvenWithForce(t *testing.T) {
	src := filepath.Join(t.TempDir(), "data.bin")
	if err := os.WriteFile(src, []byte("irreplaceable"), 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}
	rootCmd := NewRootCmd(newTestDatabase(t, false))
	rootCmd.SetArgs([]string{"compress", src, "--output", src, "--force"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)

	if err := rootCmd.Execute(); !errors.Is(err, ErrSameInputOutput) {
		t.Fatalf("error = %v, want ErrSameInputOutput", err)
	}
	if got, _ := os.ReadFile(src); string(got) != "irreplaceable" {
		t.Fatalf("source was modified to %q", got)
	}
}
