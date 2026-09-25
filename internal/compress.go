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
	"archive/zip"
	"compress/flate"
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
	zipExt   = ".zip"
)

type compressFormat string

const (
	formatGzip compressFormat = "gzip"
	formatZip  compressFormat = "zip"
)

var (
	// ErrOutputExists is returned by CheckOutputPath when the output already exists
	// and overwriting was not requested.
	ErrOutputExists = errors.New("output already exists")
	// ErrSameInputOutput is returned when an operation's output is its own input;
	// writing it would destroy the input before it has been read.
	ErrSameInputOutput = errors.New("output is the same file as the input")
	// ErrOutputInsideInput is returned when a directory would be archived into a file
	// inside itself, which would archive its own partial output.
	ErrOutputInsideInput = errors.New("output is inside the input directory")
)

//--------------------------------------------------core-------------------------------------------------------------------------------------------------//

// CompressFile compresses src at the given level (1–9), writing to dst.
// It defaults to gzip/tar.gz unless dst implies zip.
func CompressFile(src, dst string, level int) (err error) {
	return CompressFileWithFormat(src, dst, "", level)
}

// CompressFileWithFormat compresses src with the selected format (gzip or zip)
// at the given level (1–9), writing to dst.
func CompressFileWithFormat(src, dst, format string, level int) (err error) {
	selectedFormat, err := resolveCompressFormat(format, dst)
	if err != nil {
		return err
	}

	if level < gzip.BestSpeed || level > gzip.BestCompression {
		level = gzip.DefaultCompression
	}

	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat source path: %w", err)
	}

	if dst == "" {
		dst = defaultCompressOutput(src, info.IsDir(), selectedFormat)
	}
	if err := checkNotSameFile(src, dst); err != nil {
		return err
	}
	if info.IsDir() {
		if err := checkOutputOutsideDir(src, dst); err != nil {
			return err
		}
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer closeWithError(&err, out, "close output file")

	if selectedFormat == formatZip {
		if err := writeZip(out, src, info, level); err != nil {
			return err
		}
		return nil
	}

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

// DecompressFile decompresses a gzip/zip file at src, writing to dst.
// Tar-based gzip archives and zip archives are extracted into directories.
func DecompressFile(src, dst string) (err error) {
	if strings.HasSuffix(strings.ToLower(src), zipExt) {
		if dst == "" {
			dst = defaultDecompressOutput(src)
		}
		if err := checkNotSameFile(src, dst); err != nil {
			return err
		}
		return extractZip(src, dst)
	}

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
	if err := checkNotSameFile(src, dst); err != nil {
		return err
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

func defaultCompressOutput(src string, isDir bool, format compressFormat) string {
	if format == formatZip {
		return filepath.Clean(src) + zipExt
	}
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
	case strings.HasSuffix(src, zipExt):
		return src[:len(src)-len(zipExt)]
	default:
		return src + ".dec"
	}
}

func isTarGzArchive(src, gzipName string) bool {
	return strings.HasSuffix(src, tarGzExt) ||
		strings.HasSuffix(src, tgzExt) ||
		strings.HasSuffix(gzipName, ".tar")
}

func resolveCompressFormat(format, dst string) (compressFormat, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", string(formatGzip):
		if strings.HasSuffix(strings.ToLower(dst), zipExt) {
			return formatZip, nil
		}
		return formatGzip, nil
	case string(formatZip):
		return formatZip, nil
	default:
		return "", fmt.Errorf("unsupported compression format %q (supported: gzip, zip)", format)
	}
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

func writeZip(w io.Writer, src string, info fs.FileInfo, level int) (err error) {
	zw := zip.NewWriter(w)
	defer closeWithError(&err, zw, "finalise zip")

	zw.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, level)
	})

	if info.IsDir() {
		return writeZipDirectory(zw, src)
	}
	return writeZipFile(zw, src, filepath.Base(src))
}

func writeZipDirectory(zw *zip.Writer, root string) error {
	root = filepath.Clean(root)
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk source directory: %w", walkErr)
		}
		if path == root {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("read entry info: %w", err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("compress directory: symlinks are not supported (%s)", path)
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return fmt.Errorf("compress directory: unsupported file type at %s", path)
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("derive archive path: %w", err)
		}
		rel = filepath.ToSlash(rel)

		if info.IsDir() {
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return fmt.Errorf("create zip header: %w", err)
			}
			header.Name = rel + "/"
			if _, err := zw.CreateHeader(header); err != nil {
				return fmt.Errorf("write zip header: %w", err)
			}
			return nil
		}

		return writeZipFile(zw, path, rel)
	})
}

func writeZipFile(zw *zip.Writer, srcPath, zipName string) error {
	info, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("stat source file: %w", err)
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("create zip header: %w", err)
	}
	header.Name = filepath.ToSlash(zipName)
	header.Method = zip.Deflate

	writer, err := zw.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("write zip header: %w", err)
	}

	file, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	if _, err := io.Copy(writer, file); err != nil {
		closeErr := file.Close()
		if closeErr != nil {
			return errors.Join(fmt.Errorf("write zip contents: %w", err), fmt.Errorf("close source file: %w", closeErr))
		}
		return fmt.Errorf("write zip contents: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close source file: %w", err)
	}
	return nil
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

func extractZip(src, dst string) (err error) {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("open zip archive: %w", err)
	}
	defer closeWithError(&err, r, "close zip reader")

	singleFile := len(r.File) == 1 && !r.File[0].FileInfo().IsDir() && r.File[0].Mode().IsRegular()
	cleanDst := filepath.Clean(dst)
	if singleFile {
		entryName := filepath.Clean(filepath.FromSlash(r.File[0].Name))
		if filepath.Base(entryName) == filepath.Base(cleanDst) {
			return extractZipSingleFile(r.File[0], cleanDst)
		}
	}

	cleanDstWithSep := cleanDst + string(os.PathSeparator)
	if err := os.MkdirAll(cleanDst, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	for _, file := range r.File {
		archivePath := filepath.Clean(filepath.FromSlash(file.Name))
		target := filepath.Join(cleanDst, archivePath)
		if target == cleanDst {
			continue
		}
		if !strings.HasPrefix(target, cleanDstWithSep) {
			return fmt.Errorf("extract archive: invalid path %q", file.Name)
		}
		if archivePath == ".." || strings.HasPrefix(archivePath, ".."+string(filepath.Separator)) {
			return fmt.Errorf("extract archive: invalid path %q", file.Name)
		}

		mode := file.Mode()
		if mode&os.ModeSymlink != 0 {
			return fmt.Errorf("extract archive: unsupported entry type %q", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, mode.Perm()); err != nil {
				return fmt.Errorf("create directory: %w", err)
			}
			continue
		}
		if !mode.IsRegular() {
			return fmt.Errorf("extract archive: unsupported entry type %q", file.Name)
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("create parent directory: %w", err)
		}

		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("open zip entry: %w", err)
		}

		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode.Perm())
		if err != nil {
			closeErr := rc.Close()
			if closeErr != nil {
				return errors.Join(fmt.Errorf("create output file: %w", err), fmt.Errorf("close zip entry: %w", closeErr))
			}
			return fmt.Errorf("create output file: %w", err)
		}
		if _, err := io.Copy(out, rc); err != nil {
			closeOutErr := out.Close()
			closeReadErr := rc.Close()
			if closeOutErr != nil && closeReadErr != nil {
				return errors.Join(fmt.Errorf("extract file contents: %w", err), fmt.Errorf("close output file: %w", closeOutErr), fmt.Errorf("close zip entry: %w", closeReadErr))
			}
			if closeOutErr != nil {
				return errors.Join(fmt.Errorf("extract file contents: %w", err), fmt.Errorf("close output file: %w", closeOutErr))
			}
			if closeReadErr != nil {
				return errors.Join(fmt.Errorf("extract file contents: %w", err), fmt.Errorf("close zip entry: %w", closeReadErr))
			}
			return fmt.Errorf("extract file contents: %w", err)
		}
		if err := out.Close(); err != nil {
			closeErr := rc.Close()
			if closeErr != nil {
				return errors.Join(fmt.Errorf("close output file: %w", err), fmt.Errorf("close zip entry: %w", closeErr))
			}
			return fmt.Errorf("close output file: %w", err)
		}
		if err := rc.Close(); err != nil {
			return fmt.Errorf("close zip entry: %w", err)
		}
	}

	return nil
}

func extractZipSingleFile(file *zip.File, dst string) error {
	mode := file.Mode()
	if mode&os.ModeSymlink != 0 || !mode.IsRegular() {
		return fmt.Errorf("extract archive: unsupported entry type %q", file.Name)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create parent directory: %w", err)
	}

	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("open zip entry: %w", err)
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode.Perm())
	if err != nil {
		closeErr := rc.Close()
		if closeErr != nil {
			return errors.Join(fmt.Errorf("create output file: %w", err), fmt.Errorf("close zip entry: %w", closeErr))
		}
		return fmt.Errorf("create output file: %w", err)
	}

	if _, err := io.Copy(out, rc); err != nil {
		closeOutErr := out.Close()
		closeReadErr := rc.Close()
		if closeOutErr != nil && closeReadErr != nil {
			return errors.Join(fmt.Errorf("extract file contents: %w", err), fmt.Errorf("close output file: %w", closeOutErr), fmt.Errorf("close zip entry: %w", closeReadErr))
		}
		if closeOutErr != nil {
			return errors.Join(fmt.Errorf("extract file contents: %w", err), fmt.Errorf("close output file: %w", closeOutErr))
		}
		if closeReadErr != nil {
			return errors.Join(fmt.Errorf("extract file contents: %w", err), fmt.Errorf("close zip entry: %w", closeReadErr))
		}
		return fmt.Errorf("extract file contents: %w", err)
	}
	if err := out.Close(); err != nil {
		closeErr := rc.Close()
		if closeErr != nil {
			return errors.Join(fmt.Errorf("close output file: %w", err), fmt.Errorf("close zip entry: %w", closeErr))
		}
		return fmt.Errorf("close output file: %w", err)
	}
	if err := rc.Close(); err != nil {
		return fmt.Errorf("close zip entry: %w", err)
	}
	return nil
}

//--------------------------------------------------output paths-----------------------------------------------------------------------------------------//

// CheckOutputPath reports whether an operation reading src may write dst. It fails
// with ErrSameInputOutput when dst is src itself, even when overwrite is true, and
// with ErrOutputExists when dst already exists and overwrite is false. The CLI and
// TUI call it before encrypting, decrypting, compressing or decompressing.
func CheckOutputPath(src, dst string, overwrite bool) error {
	if err := checkNotSameFile(src, dst); err != nil {
		return err
	}
	if _, err := os.Lstat(dst); err == nil {
		if !overwrite {
			return fmt.Errorf("%w: %s", ErrOutputExists, dst)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("check output path: %w", err)
	}
	return nil
}

// checkNotSameFile returns ErrSameInputOutput when dst names the same file as src,
// including through a different spelling, a symlink or a hard link.
func checkNotSameFile(src, dst string) error {
	dstInfo, err := os.Stat(dst)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("check output path: %w", err)
	}
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("check input path: %w", err)
	}
	if os.SameFile(srcInfo, dstInfo) {
		return fmt.Errorf("%w: %s", ErrSameInputOutput, dst)
	}
	return nil
}

// checkOutputOutsideDir returns ErrOutputInsideInput when dst lies inside srcDir.
func checkOutputOutsideDir(srcDir, dst string) error {
	absSrc, err := filepath.Abs(srcDir)
	if err != nil {
		return fmt.Errorf("resolve input path: %w", err)
	}
	absDst, err := filepath.Abs(dst)
	if err != nil {
		return fmt.Errorf("resolve output path: %w", err)
	}
	rel, err := filepath.Rel(absSrc, absDst)
	if err != nil {
		return nil // e.g. different Windows volumes: dst cannot be inside srcDir
	}
	if rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("%w: %s", ErrOutputInsideInput, dst)
	}
	return nil
}

func closeWithError(target *error, closer io.Closer, message string) {
	if closer == nil {
		return
	}
	if err := closer.Close(); err != nil && *target == nil {
		*target = fmt.Errorf("%s: %w", message, err)
	}
}
