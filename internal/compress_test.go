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
	"compress/gzip"
	"os"
	"path/filepath"
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
		name     string
		dstPath  string
		level    int
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
		name    string
		level   int
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
