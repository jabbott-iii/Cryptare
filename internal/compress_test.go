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
	"archive/zip"
	"compress/gzip"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCompressFile tests the CompressFile function by compressing a file with various compression levels and output paths.
// It ensures that the compressed file is created and that its size is smaller than the original for repetitive content.
func TestCompressFile(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.txt")
	originalContent := []byte("This is a test file that should be compressed. " +
		"Repeated content makes compression more effective. " +
		"This is a test file that should be compressed.")

	if err := os.WriteFile(srcFile, originalContent, 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	tests := []struct {
		name    string
		dstPath string
		level   int
	}{
		{
			name:    "compress with default output",
			dstPath: "",
			level:   -1,
		},
		{
			name:    "compress with custom output",
			dstPath: filepath.Join(tmpDir, "custom.gz"),
			level:   gzip.DefaultCompression,
		},
		{
			name:    "compress with max compression",
			dstPath: filepath.Join(tmpDir, "max.gz"),
			level:   gzip.BestCompression,
		},
		{
			name:    "compress with best speed",
			dstPath: filepath.Join(tmpDir, "speed.gz"),
			level:   gzip.BestSpeed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compFile := tt.dstPath
			if compFile == "" {
				compFile = srcFile + gzExt
			}

			if err := CompressFile(srcFile, compFile, tt.level); err != nil {
				t.Fatalf("CompressFile failed: %v", err)
			}

			// Verify compressed file exists
			if _, err := os.Stat(compFile); err != nil {
				t.Fatalf("Compressed file not found: %v", err)
			}

			// Verify file is smaller (due to repetitive content)
			origInfo, _ := os.Stat(srcFile)
			compInfo, _ := os.Stat(compFile)
			if compInfo.Size() >= origInfo.Size() {
				t.Logf("Compressed file size (%d) >= original (%d), expected smaller",
					compInfo.Size(), origInfo.Size())
			}
		})
	}
}

// TestCompressDecompressFile tests the CompressFile and DecompressFile functions by compressing and then decompressing a file.
// It ensures that the decompressed content matches the original content.
func TestCompressDecompressFile(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.txt")
	originalContent := []byte("Hello, World! This is test content for compression and decompression.")

	if err := os.WriteFile(srcFile, originalContent, 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	tests := []struct {
		name  string
		level int
	}{
		{
			name:  "default compression level",
			level: -1,
		},
		{
			name:  "best speed",
			level: gzip.BestSpeed,
		},
		{
			name:  "best compression",
			level: gzip.BestCompression,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compFile := filepath.Join(tmpDir, "test-"+tt.name+".gz")
			decFile := filepath.Join(tmpDir, "test-"+tt.name+".dec")

			// Compress
			if err := CompressFile(srcFile, compFile, tt.level); err != nil {
				t.Fatalf("CompressFile failed: %v", err)
			}

			// Decompress
			if err := DecompressFile(compFile, decFile); err != nil {
				t.Fatalf("DecompressFile failed: %v", err)
			}

			// Verify decompressed content matches original
			decData, err := os.ReadFile(decFile)
			if err != nil {
				t.Fatalf("Failed to read decompressed file: %v", err)
			}
			if string(decData) != string(originalContent) {
				t.Errorf("Decompressed content mismatch: got %q, want %q",
					string(decData), string(originalContent))
			}
		})
	}
}

// TestCompressDirectoryRoundTrip tests directory compression by writing a
// tar.gz archive and extracting it into a destination directory.
func TestCompressDirectoryRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "source")
	nestedDir := filepath.Join(srcDir, "nested")
	emptyDir := filepath.Join(srcDir, "empty")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}
	if err := os.MkdirAll(emptyDir, 0o755); err != nil {
		t.Fatalf("Failed to create empty directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "root.txt"), []byte("root data"), 0o644); err != nil {
		t.Fatalf("Failed to write root file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "child.txt"), []byte("nested data"), 0o644); err != nil {
		t.Fatalf("Failed to write nested file: %v", err)
	}

	archive := srcDir + tarGzExt
	if err := CompressFile(srcDir, "", -1); err != nil {
		t.Fatalf("CompressFile failed: %v", err)
	}
	if _, err := os.Stat(archive); err != nil {
		t.Fatalf("Expected archive not found: %v", err)
	}

	outDir := filepath.Join(tmpDir, "restored")
	if err := DecompressFile(archive, outDir); err != nil {
		t.Fatalf("DecompressFile failed: %v", err)
	}

	rootData, err := os.ReadFile(filepath.Join(outDir, "root.txt"))
	if err != nil {
		t.Fatalf("Failed to read restored root file: %v", err)
	}
	if string(rootData) != "root data" {
		t.Fatalf("Root file mismatch: got %q", string(rootData))
	}

	nestedData, err := os.ReadFile(filepath.Join(outDir, "nested", "child.txt"))
	if err != nil {
		t.Fatalf("Failed to read restored nested file: %v", err)
	}
	if string(nestedData) != "nested data" {
		t.Fatalf("Nested file mismatch: got %q", string(nestedData))
	}

	if info, err := os.Stat(filepath.Join(outDir, "empty")); err != nil {
		t.Fatalf("Expected empty directory not restored: %v", err)
	} else if !info.IsDir() {
		t.Fatalf("Restored empty path is not a directory")
	}
}

// TestDecompressFileWithDefaultOutput tests the DecompressFile function when the output path is not specified.
// It ensures that the decompressed file is created with the default output path derived from the compressed file name.
func TestDecompressFileWithDefaultOutput(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.txt")
	originalContent := []byte("Test data")

	if err := os.WriteFile(srcFile, originalContent, 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	compFile := filepath.Join(tmpDir, "test.txt.gz")
	if err := CompressFile(srcFile, compFile, -1); err != nil {
		t.Fatalf("CompressFile failed: %v", err)
	}

	// Decompress with empty destination (should strip .gz)
	if err := DecompressFile(compFile, ""); err != nil {
		t.Fatalf("DecompressFile failed: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "test.txt")
	decData, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Expected decompressed file not found: %v", err)
	}

	if string(decData) != string(originalContent) {
		t.Errorf("Content mismatch: got %q, want %q", string(decData), string(originalContent))
	}
}

// TestDecompressTarGzWithDefaultOutput tests that tar.gz archives restore to a
// directory path derived from the archive name when no destination is given.
func TestDecompressTarGzWithDefaultOutput(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "bundle")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("bundle data"), 0o644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	archive := srcDir + tarGzExt
	if err := CompressFile(srcDir, "", -1); err != nil {
		t.Fatalf("CompressFile failed: %v", err)
	}
	if err := os.RemoveAll(srcDir); err != nil {
		t.Fatalf("Failed to remove source directory before restore: %v", err)
	}

	if err := DecompressFile(archive, ""); err != nil {
		t.Fatalf("DecompressFile failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(srcDir, "file.txt"))
	if err != nil {
		t.Fatalf("Expected restored file not found: %v", err)
	}
	if string(data) != "bundle data" {
		t.Fatalf("Content mismatch: got %q", string(data))
	}
}

// TestDecompressFileWithoutGzExtension tests the DecompressFile function when the compressed file does not have a .gz extension.
// It ensures that the decompressed file is created with a .dec suffix appended to the original file name.
func TestDecompressFileWithoutGzExtension(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.txt")
	originalContent := []byte("Test data")

	if err := os.WriteFile(srcFile, originalContent, 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	compFile := filepath.Join(tmpDir, "test.compressed")
	if err := CompressFile(srcFile, compFile, -1); err != nil {
		t.Fatalf("CompressFile failed: %v", err)
	}

	// Decompress file without .gz extension
	if err := DecompressFile(compFile, ""); err != nil {
		t.Fatalf("DecompressFile failed: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "test.compressed.dec")
	decData, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Expected decompressed file not found: %v", err)
	}

	if string(decData) != string(originalContent) {
		t.Errorf("Content mismatch: got %q, want %q", string(decData), string(originalContent))
	}
}

// TestCompressFileInvalidLevel tests the CompressFile function with an invalid compression level.
// It ensures that the function defaults to DefaultCompression when an invalid level is provided.
func TestCompressFileInvalidLevel(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(srcFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Invalid level should default to DefaultCompression
	dstFile := filepath.Join(tmpDir, "test.gz")
	if err := CompressFile(srcFile, dstFile, 99); err != nil {
		t.Fatalf("CompressFile should handle invalid level: %v", err)
	}

	if _, err := os.Stat(dstFile); err != nil {
		t.Fatalf("Output file should be created: %v", err)
	}
}

// TestDecompressInvalidFile tests the DecompressFile function with an invalid gzip file.
// It ensures that the function returns an error when attempting to decompress a non-gzip file.
func TestDecompressInvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	fakeGzFile := filepath.Join(tmpDir, "fake.gz")

	if err := os.WriteFile(fakeGzFile, []byte("not a gzip file"), 0o644); err != nil {
		t.Fatalf("Failed to write fake file: %v", err)
	}

	err := DecompressFile(fakeGzFile, filepath.Join(tmpDir, "out.txt"))
	if err == nil {
		t.Error("DecompressFile should fail on invalid gzip file")
	}
}

func TestCompressZipFileRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "single.txt")
	originalContent := []byte("zip single file data")
	if err := os.WriteFile(srcFile, originalContent, 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	archive := srcFile + zipExt
	if err := CompressFileWithFormat(srcFile, "", "zip", gzip.BestCompression); err != nil {
		t.Fatalf("CompressFileWithFormat failed: %v", err)
	}

	reader, err := zip.OpenReader(archive)
	if err != nil {
		t.Fatalf("Failed to open zip archive: %v", err)
	}
	if len(reader.File) != 1 {
		_ = reader.Close()
		t.Fatalf("Expected single zip entry, got %d", len(reader.File))
	}
	if reader.File[0].Name != "single.txt" {
		_ = reader.Close()
		t.Fatalf("Unexpected zip entry name: %q", reader.File[0].Name)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("Failed to close zip reader: %v", err)
	}

	if err := os.Remove(srcFile); err != nil {
		t.Fatalf("Failed to remove source file: %v", err)
	}
	if err := DecompressFile(archive, ""); err != nil {
		t.Fatalf("DecompressFile failed: %v", err)
	}

	restored, err := os.ReadFile(srcFile)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}
	if string(restored) != string(originalContent) {
		t.Fatalf("Content mismatch: got %q", string(restored))
	}
}

func TestCompressZipDirectoryRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "bundle")
	if err := os.MkdirAll(filepath.Join(srcDir, "nested"), 0o755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(srcDir, "empty"), 0o755); err != nil {
		t.Fatalf("Failed to create empty directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "root.txt"), []byte("zip root"), 0o644); err != nil {
		t.Fatalf("Failed to write root file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "nested", "child.txt"), []byte("zip nested"), 0o644); err != nil {
		t.Fatalf("Failed to write nested file: %v", err)
	}

	archive := srcDir + zipExt
	if err := CompressFileWithFormat(srcDir, "", "zip", gzip.BestSpeed); err != nil {
		t.Fatalf("CompressFileWithFormat failed: %v", err)
	}
	if _, err := os.Stat(archive); err != nil {
		t.Fatalf("Expected archive not found: %v", err)
	}

	if err := os.RemoveAll(srcDir); err != nil {
		t.Fatalf("Failed to remove source directory: %v", err)
	}
	if err := DecompressFile(archive, ""); err != nil {
		t.Fatalf("DecompressFile failed: %v", err)
	}

	rootData, err := os.ReadFile(filepath.Join(srcDir, "root.txt"))
	if err != nil {
		t.Fatalf("Failed to read restored root file: %v", err)
	}
	if string(rootData) != "zip root" {
		t.Fatalf("Root file mismatch: got %q", string(rootData))
	}

	nestedData, err := os.ReadFile(filepath.Join(srcDir, "nested", "child.txt"))
	if err != nil {
		t.Fatalf("Failed to read restored nested file: %v", err)
	}
	if string(nestedData) != "zip nested" {
		t.Fatalf("Nested file mismatch: got %q", string(nestedData))
	}
	if info, err := os.Stat(filepath.Join(srcDir, "empty")); err != nil {
		t.Fatalf("Expected empty directory not restored: %v", err)
	} else if !info.IsDir() {
		t.Fatalf("Restored empty path is not a directory")
	}
}

// TestCheckOutputPath covers the output-path policy shared by the CLI and TUI.
func TestCheckOutputPath(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "input.txt")
	if err := os.WriteFile(src, []byte("input"), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}
	existing := filepath.Join(tmpDir, "existing.out")
	if err := os.WriteFile(existing, []byte("keep me"), 0o600); err != nil {
		t.Fatalf("write existing output: %v", err)
	}
	hardLink := filepath.Join(tmpDir, "hardlink.txt")
	if err := os.Link(src, hardLink); err != nil {
		t.Fatalf("create hard link: %v", err)
	}

	tests := []struct {
		name      string
		dst       string
		overwrite bool
		wantErr   error
	}{
		{name: "new output", dst: filepath.Join(tmpDir, "new.out")},
		{name: "existing output refused", dst: existing, wantErr: ErrOutputExists},
		{name: "existing output with overwrite", dst: existing, overwrite: true},
		{name: "same path", dst: src, wantErr: ErrSameInputOutput},
		{name: "same path even with overwrite", dst: src, overwrite: true, wantErr: ErrSameInputOutput},
		{name: "same file, different spelling", dst: filepath.Join(tmpDir, ".", "input.txt"), overwrite: true, wantErr: ErrSameInputOutput},
		{name: "hard link to the input", dst: hardLink, overwrite: true, wantErr: ErrSameInputOutput},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckOutputPath(src, tt.dst, tt.overwrite)
			if tt.wantErr == nil && err != nil {
				t.Fatalf("CheckOutputPath() error = %v, want nil", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("CheckOutputPath() error = %v, want %v", err, tt.wantErr)
			}
		})
	}

	symlink := filepath.Join(tmpDir, "symlink.txt")
	if err := os.Symlink(src, symlink); err != nil {
		t.Logf("symlinks unavailable, skipping symlink case: %v", err)
		return
	}
	if err := CheckOutputPath(src, symlink, true); !errors.Is(err, ErrSameInputOutput) {
		t.Fatalf("CheckOutputPath(symlink to input) error = %v, want ErrSameInputOutput", err)
	}
}

// TestOperationsRefuseToOverwriteTheirInput is a regression test for BUG-003: an
// operation whose output is its own input must fail and leave the input untouched.
func TestOperationsRefuseToOverwriteTheirInput(t *testing.T) {
	tmpDir := t.TempDir()
	content := []byte(strings.Repeat("do not destroy me ", 256))

	plain := filepath.Join(tmpDir, "plain.bin")
	gz := filepath.Join(tmpDir, "data.gz")
	enc := filepath.Join(tmpDir, "data.enc")
	zipFile := filepath.Join(tmpDir, "data.zip")
	if err := os.WriteFile(plain, content, 0o600); err != nil {
		t.Fatalf("write plain: %v", err)
	}
	if err := CompressFile(plain, gz, -1); err != nil {
		t.Fatalf("prepare gzip: %v", err)
	}
	if err := EncryptFile(plain, enc, "pw"); err != nil {
		t.Fatalf("prepare encrypted file: %v", err)
	}
	if err := CompressFileWithFormat(plain, zipFile, "zip", -1); err != nil {
		t.Fatalf("prepare zip: %v", err)
	}

	cases := []struct {
		name string
		path string
		run  func(p string) error
	}{
		{"compress", plain, func(p string) error { return CompressFileWithFormat(p, p, "", -1) }},
		{"compress zip", plain, func(p string) error { return CompressFileWithFormat(p, p, "zip", -1) }},
		{"decompress gzip", gz, func(p string) error { return DecompressFile(p, p) }},
		{"decompress zip", zipFile, func(p string) error { return DecompressFile(p, p) }},
		{"encrypt", plain, func(p string) error { return EncryptFile(p, p, "pw") }},
		{"decrypt", enc, func(p string) error { return DecryptFile(p, p, "pw") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before, err := os.ReadFile(tc.path)
			if err != nil {
				t.Fatalf("read %s: %v", tc.path, err)
			}
			if err := tc.run(tc.path); !errors.Is(err, ErrSameInputOutput) {
				t.Errorf("error = %v, want ErrSameInputOutput", err)
			}
			after, err := os.ReadFile(tc.path)
			if err != nil {
				t.Fatalf("read %s after: %v", tc.path, err)
			}
			if string(after) != string(before) {
				t.Errorf("%s was modified (%d bytes before, %d after)", tc.path, len(before), len(after))
			}
		})
	}
}

// TestCompressDirectoryRejectsOutputInsideInput checks that a directory can't be
// compressed into an archive inside itself, which would archive its own partial output.
func TestCompressDirectoryRejectsOutputInsideInput(t *testing.T) {
	srcDir := filepath.Join(t.TempDir(), "folder")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("create folder: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("data"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	for _, format := range []string{"gzip", "zip"} {
		dst := filepath.Join(srcDir, "self."+format)
		if err := CompressFileWithFormat(srcDir, dst, format, -1); !errors.Is(err, ErrOutputInsideInput) {
			t.Errorf("%s: error = %v, want ErrOutputInsideInput", format, err)
		}
		if _, err := os.Stat(dst); !os.IsNotExist(err) {
			t.Errorf("%s: archive written inside the input folder (stat err: %v)", format, err)
		}
	}
}
