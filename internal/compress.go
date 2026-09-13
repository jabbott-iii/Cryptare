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
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	gzExt    = ".gz"
	tarGzExt = ".tar.gz"
	tgzExt   = ".tgz"
)

//--------------------------------------------------core-------------------------------------------------------------------------------------------------//

// CompressFile compresses src with gzip at the given level (1–9), writing to dst.
// If src is a directory, it is archived as tar.gz.
func CompressFile(src, dst string, level int) (err error) {
	if level < gzip.BestSpeed || level > gzip.BestCompression {
		level = gzip.DefaultCompression
	}

	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat source path: %w", err)
	}

	if dst == "" {
		dst = defaultCompressOutput(src, info.IsDir())
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer closeWithError(&err, out, "close output file")

	gz, err := gzip.NewWriterLevel(out, level)
	if err != nil {
		return fmt.Errorf("create gzip writer: %w", err)
	}
	defer closeWithError(&err, gz, "finalise gzip")

	if info.IsDir() {
		gz.Name = filepath.Base(filepath.Clean(src)) + ".tar"
		if _, err := writeTarGz(gz, src); err != nil {
			return err
		}
	} else {
		in, err := os.Open(src)
		if err != nil {
			return fmt.Errorf("open source file: %w", err)
		}
		defer closeWithError(&err, in, "close source file")

		gz.Name = filepath.Base(src)

		if _, err := io.Copy(gz, in); err != nil {
			return fmt.Errorf("compress data: %w", err)
		}
	}
	return nil
}

// DecompressFile decompresses a gzip file at src, writing to dst.
// Tar-based gzip archives are extracted into a directory.
func DecompressFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer closeWithError(&err, in, "close source file")

	gz, err := gzip.NewReader(in)
	if err != nil {
		return fmt.Errorf("create gzip reader: %w", err)
	}
	defer closeWithError(&err, gz, "close gzip reader")

	if dst == "" {
		dst = defaultDecompressOutput(src)
	}

	if isTarGzArchive(src, gz.Name) {
		if err := extractTarGz(gz, dst); err != nil {
			return err
		}
		return nil
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer closeWithError(&err, out, "close output file")

	if _, err := io.Copy(out, gz); err != nil {
		return fmt.Errorf("decompress data: %w", err)
	}
	return nil
}

func defaultCompressOutput(src string, isDir bool) string {
	if isDir {
		return filepath.Clean(src) + tarGzExt
	}
	return src + gzExt
}

func defaultDecompressOutput(src string) string {
	switch {
	case strings.HasSuffix(src, tarGzExt):
		return src[:len(src)-len(tarGzExt)]
	case strings.HasSuffix(src, tgzExt):
		return src[:len(src)-len(tgzExt)]
	case strings.HasSuffix(src, gzExt):
		return src[:len(src)-len(gzExt)]
	default:
		return src + ".dec"
	}
}

func isTarGzArchive(src, gzipName string) bool {
	return strings.HasSuffix(src, tarGzExt) ||
		strings.HasSuffix(src, tgzExt) ||
		strings.HasSuffix(gzipName, ".tar")
}

func writeTarGz(w io.Writer, root string) (int, error) {
	tw := tar.NewWriter(w)
	entries := 0

	root = filepath.Clean(root)
	if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk source directory: %w", walkErr)
		}
		if path == root {
			return nil
		}
		entries++

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("read entry info: %w", err)
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("derive archive path: %w", err)
		}
		rel = filepath.ToSlash(rel)

		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("compress directory: symlinks are not supported (%s)", path)
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return fmt.Errorf("compress directory: unsupported file type at %s", path)
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("create tar header: %w", err)
		}
		header.Name = rel
		if info.IsDir() {
			header.Name += "/"
		}

		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("write tar header: %w", err)
		}
		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open source file: %w", err)
		}

		if _, err := io.Copy(tw, file); err != nil {
			closeErr := file.Close()
			if closeErr != nil {
				return errors.Join(fmt.Errorf("write tar contents: %w", err), fmt.Errorf("close source file: %w", closeErr))
			}
			return fmt.Errorf("write tar contents: %w", err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("close source file: %w", err)
		}
		return nil
	}); err != nil {
		closeErr := tw.Close()
		if closeErr != nil {
			return entries, errors.Join(err, fmt.Errorf("finalise tar archive: %w", closeErr))
		}
		return entries, err
	}
	if err := tw.Close(); err != nil {
		return entries, fmt.Errorf("finalise tar archive: %w", err)
	}
	return entries, nil
}

func extractTarGz(r io.Reader, dst string) error {
	cleanDst := filepath.Clean(dst)
	cleanDstWithSep := cleanDst + string(os.PathSeparator)
	if err := os.MkdirAll(cleanDst, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read tar header: %w", err)
		}

		archivePath := filepath.Clean(filepath.FromSlash(header.Name))
		target := filepath.Join(cleanDst, archivePath)
		if target == cleanDst {
			continue
		}
		if !strings.HasPrefix(target, cleanDstWithSep) {
			return fmt.Errorf("extract archive: invalid path %q", header.Name)
		}
		if archivePath == ".." || strings.HasPrefix(archivePath, ".."+string(filepath.Separator)) {
			return fmt.Errorf("extract archive: invalid path %q", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, header.FileInfo().Mode().Perm()); err != nil {
				return fmt.Errorf("create directory: %w", err)
			}
		case tar.TypeReg, 0:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return fmt.Errorf("create parent directory: %w", err)
			}
			file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, header.FileInfo().Mode().Perm())
			if err != nil {
				return fmt.Errorf("create output file: %w", err)
			}
			if _, err := io.Copy(file, tr); err != nil {
				closeErr := file.Close()
				if closeErr != nil {
					return errors.Join(fmt.Errorf("extract file contents: %w", err), fmt.Errorf("close output file: %w", closeErr))
				}
				return fmt.Errorf("extract file contents: %w", err)
			}
			if err := file.Close(); err != nil {
				return fmt.Errorf("close output file: %w", err)
			}
		default:
			return fmt.Errorf("extract archive: unsupported entry type %q", header.Name)
		}
	}
}

func closeWithError(target *error, closer io.Closer, message string) {
	if closer == nil {
		return
	}
	if err := closer.Close(); err != nil && *target == nil {
		*target = fmt.Errorf("%s: %w", message, err)
	}
}
