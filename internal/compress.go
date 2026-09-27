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
	"math"
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
	// ErrInputInsideOutput is returned when extraction would replace an existing
	// output directory that contains the archive being extracted.
	ErrInputInsideOutput = errors.New("input is inside the output directory")
	// ErrExtractLimit is returned when decompression or extraction would exceed its
	// ExtractLimits. The partial output is removed.
	ErrExtractLimit = errors.New("extraction limit exceeded")
)

// Default extraction limits (SEC-007): the most one decompression or extraction may
// write, as a guard against decompression bombs.
const (
	DefaultMaxExtractBytes   int64 = 10 << 30 // 10 GiB
	DefaultMaxExtractEntries       = 100_000
)

// ExtractLimits bounds what one decompression or extraction may write. A zero field
// means no limit.
type ExtractLimits struct {
	MaxBytes   int64 // total bytes of file content written
	MaxEntries int   // archive entries (files and directories) extracted
}

// DefaultExtractLimits returns the limits used when the caller doesn't set its own.
func DefaultExtractLimits() ExtractLimits {
	return ExtractLimits{MaxBytes: DefaultMaxExtractBytes, MaxEntries: DefaultMaxExtractEntries}
}

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

	out, err := createAtomicFile(dst)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer out.Abort()

	if selectedFormat == formatZip {
		err = writeZip(out, src, info, level)
	} else {
		err = writeGzip(out, src, info, level)
	}
	if err != nil {
		return err
	}
	if err := out.Commit(); err != nil {
		return fmt.Errorf("finalise output file: %w", err)
	}
	return nil
}

// writeGzip writes src to w as gzip: a tar.gz stream for a directory, plain gzip for
// a file. The gzip stream is finalised before it returns.
func writeGzip(w io.Writer, src string, info fs.FileInfo, level int) (err error) {
	gz, err := gzip.NewWriterLevel(w, level)
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
func DecompressFile(src, dst string) error {
	return DecompressFileWithLimits(src, dst, DefaultExtractLimits())
}

// DecompressFileWithLimits is DecompressFile with explicit extraction limits. When a
// limit is hit it returns ErrExtractLimit and leaves no output behind. Archives are
// extracted into a new hidden directory next to dst, which is then renamed to dst,
// so an existing dst is replaced as a whole rather than merged into.
func DecompressFileWithLimits(src, dst string, limits ExtractLimits) (err error) {
	budget := &extractBudget{limits: limits}
	if strings.HasSuffix(strings.ToLower(src), zipExt) {
		if dst == "" {
			dst = defaultDecompressOutput(src)
		}
		if err := checkNotSameFile(src, dst); err != nil {
			return err
		}
		return extractZip(src, dst, budget)
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
		return extractToDir(src, dst, func(dir string) error {
			return extractTarGz(gz, dir, budget)
		})
	}

	out, err := createAtomicFile(dst)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer out.Abort()

	if err := budget.copy(out, gz); err != nil {
		if errors.Is(err, ErrExtractLimit) {
			return err
		}
		return fmt.Errorf("decompress data: %w", err)
	}
	if err := out.Commit(); err != nil {
		return fmt.Errorf("finalise output file: %w", err)
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

// extractTarGz extracts a tar stream into the existing directory dst, counting its
// entries and bytes against budget. Callers pass a new, empty directory (see
// extractToDir). Writes go through an os.Root on dst, so no entry can land outside
// it, and permissions are set by extractDirMode and extractFileMode.
func extractTarGz(r io.Reader, dst string, budget *extractBudget) (err error) {
	cleanDst := filepath.Clean(dst)
	cleanDstWithSep := cleanDst + string(os.PathSeparator)
	root, err := os.OpenRoot(cleanDst)
	if err != nil {
		return fmt.Errorf("open output directory: %w", err)
	}
	defer closeWithError(&err, root, "close output directory")

	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read tar header: %w", err)
		}
		if err := budget.addEntry(); err != nil {
			return err
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
		// Relative to the output; a leading "/" in the entry name was dropped by Join.
		name := strings.TrimPrefix(target, cleanDstWithSep)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := root.MkdirAll(name, extractDirMode); err != nil {
				return fmt.Errorf("create directory: %w", err)
			}
		case tar.TypeReg, 0:
			if err := root.MkdirAll(filepath.Dir(name), extractDirMode); err != nil {
				return fmt.Errorf("create parent directory: %w", err)
			}
			file, err := root.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, extractFileMode(header.FileInfo().Mode()))
			if err != nil {
				return fmt.Errorf("create output file: %w", err)
			}
			if err := budget.copy(file, tr); err != nil {
				closeErr := file.Close()
				if closeErr != nil {
					return errors.Join(wrapExtractError(err), fmt.Errorf("close output file: %w", closeErr))
				}
				return wrapExtractError(err)
			}
			if err := file.Close(); err != nil {
				return fmt.Errorf("close output file: %w", err)
			}
		default:
			return fmt.Errorf("extract archive: unsupported entry type %q", header.Name)
		}
	}
}

func extractZip(src, dst string, budget *extractBudget) (err error) {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("open zip archive: %w", err)
	}
	defer closeWithError(&err, r, "close zip reader")

	// The central directory lists every entry, so an archive with too many is refused
	// before anything is written.
	if limit := budget.limits.MaxEntries; limit > 0 && len(r.File) > limit {
		return entryLimitError(limit)
	}

	singleFile := len(r.File) == 1 && !r.File[0].FileInfo().IsDir() && r.File[0].Mode().IsRegular()
	cleanDst := filepath.Clean(dst)
	if singleFile {
		entryName := filepath.Clean(filepath.FromSlash(r.File[0].Name))
		if filepath.Base(entryName) == filepath.Base(cleanDst) {
			return extractZipSingleFile(r.File[0], cleanDst, budget)
		}
	}

	return extractToDir(src, cleanDst, func(dir string) error {
		return extractZipEntries(r.File, dir, budget)
	})
}

// extractZipEntries extracts zip entries into the existing directory dst, like
// extractTarGz: through an os.Root, with permissions from extractDirMode and
// extractFileMode, counting entries and bytes against budget.
func extractZipEntries(files []*zip.File, dst string, budget *extractBudget) (err error) {
	cleanDst := filepath.Clean(dst)
	cleanDstWithSep := cleanDst + string(os.PathSeparator)
	root, err := os.OpenRoot(cleanDst)
	if err != nil {
		return fmt.Errorf("open output directory: %w", err)
	}
	defer closeWithError(&err, root, "close output directory")

	for _, file := range files {
		if err := budget.addEntry(); err != nil {
			return err
		}
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
		// Relative to the output; a leading "/" in the entry name was dropped by Join.
		name := strings.TrimPrefix(target, cleanDstWithSep)

		mode := file.Mode()
		if mode&os.ModeSymlink != 0 {
			return fmt.Errorf("extract archive: unsupported entry type %q", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := root.MkdirAll(name, extractDirMode); err != nil {
				return fmt.Errorf("create directory: %w", err)
			}
			continue
		}
		if !mode.IsRegular() {
			return fmt.Errorf("extract archive: unsupported entry type %q", file.Name)
		}

		if err := root.MkdirAll(filepath.Dir(name), extractDirMode); err != nil {
			return fmt.Errorf("create parent directory: %w", err)
		}

		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("open zip entry: %w", err)
		}

		out, err := root.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, extractFileMode(mode))
		if err != nil {
			closeErr := rc.Close()
			if closeErr != nil {
				return errors.Join(fmt.Errorf("create output file: %w", err), fmt.Errorf("close zip entry: %w", closeErr))
			}
			return fmt.Errorf("create output file: %w", err)
		}
		if err := budget.copy(out, rc); err != nil {
			closeOutErr := out.Close()
			closeReadErr := rc.Close()
			if closeOutErr != nil && closeReadErr != nil {
				return errors.Join(wrapExtractError(err), fmt.Errorf("close output file: %w", closeOutErr), fmt.Errorf("close zip entry: %w", closeReadErr))
			}
			if closeOutErr != nil {
				return errors.Join(wrapExtractError(err), fmt.Errorf("close output file: %w", closeOutErr))
			}
			if closeReadErr != nil {
				return errors.Join(wrapExtractError(err), fmt.Errorf("close zip entry: %w", closeReadErr))
			}
			return wrapExtractError(err)
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

// extractZipSingleFile extracts a zip holding one file straight to the file dst. Like
// other single-file outputs it is written through a temporary file and a rename, so
// it gets mode 0600 and a failure leaves no partial file.
func extractZipSingleFile(file *zip.File, dst string, budget *extractBudget) (err error) {
	mode := file.Mode()
	if mode&os.ModeSymlink != 0 || !mode.IsRegular() {
		return fmt.Errorf("extract archive: unsupported entry type %q", file.Name)
	}
	if err := budget.addEntry(); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create parent directory: %w", err)
	}

	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("open zip entry: %w", err)
	}
	defer closeWithError(&err, rc, "close zip entry")

	out, err := createAtomicFile(dst)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer out.Abort()

	if err := budget.copy(out, rc); err != nil {
		return wrapExtractError(err)
	}
	if err := out.Commit(); err != nil {
		return fmt.Errorf("finalise output file: %w", err)
	}
	return nil
}

//--------------------------------------------------extraction limits and output-------------------------------------------------------------------------//

// extractDirMode is the mode of every folder created by extraction. Extracted files
// and folders are private to the owner whatever the archive stores (SEC-008), and a
// read-only folder in the archive can't block extracting its contents.
const extractDirMode fs.FileMode = 0o700

// extractFileMode returns the mode of an extracted file: 0700 when the archive marks
// it executable for anyone, otherwise 0600.
func extractFileMode(archived fs.FileMode) fs.FileMode {
	if archived&0o111 != 0 {
		return 0o700
	}
	return 0o600
}

// extractBudget tracks one decompression or extraction against its ExtractLimits.
type extractBudget struct {
	limits  ExtractLimits
	written int64
	entries int
}

// addEntry counts one archive entry against the entry limit.
func (b *extractBudget) addEntry() error {
	b.entries++
	if b.limits.MaxEntries > 0 && b.entries > b.limits.MaxEntries {
		return entryLimitError(b.limits.MaxEntries)
	}
	return nil
}

// copy copies r to w and counts the bytes against the size limit. It reads at most
// one byte beyond the limit, which is how an overrun is detected.
func (b *extractBudget) copy(w io.Writer, r io.Reader) error {
	n := int64(math.MaxInt64)
	if b.limits.MaxBytes > 0 {
		n = b.limits.MaxBytes - b.written + 1
	}
	copied, err := io.CopyN(w, r, n)
	b.written += copied
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	if b.limits.MaxBytes > 0 && b.written > b.limits.MaxBytes {
		return fmt.Errorf("%w: the output is larger than %s", ErrExtractLimit, formatSize(b.limits.MaxBytes))
	}
	return nil
}

func entryLimitError(limit int) error {
	return fmt.Errorf("%w: the archive has more than %d entries", ErrExtractLimit, limit)
}

// wrapExtractError adds context to an error from copying an entry's contents. Limit
// errors are returned unchanged so their message stays readable.
func wrapExtractError(err error) error {
	if errors.Is(err, ErrExtractLimit) {
		return err
	}
	return fmt.Errorf("extract file contents: %w", err)
}

// formatSize formats n bytes with the largest unit (binary or decimal) that divides it
// exactly, such as "10 GiB" or "500 MB", and otherwise in bytes.
func formatSize(n int64) string {
	units := []struct {
		name string
		size int64
	}{
		{"TiB", 1 << 40}, {"TB", 1e12}, {"GiB", 1 << 30}, {"GB", 1e9},
		{"MiB", 1 << 20}, {"MB", 1e6}, {"KiB", 1 << 10}, {"KB", 1e3},
	}
	for _, u := range units {
		if n >= u.size && n%u.size == 0 {
			return fmt.Sprintf("%d %s", n/u.size, u.name)
		}
	}
	return fmt.Sprintf("%d bytes", n)
}

// extractToDir runs extract on a new, empty hidden directory next to dst (mode 0700)
// and renames it to dst when extract succeeds. A failure, including a limit being
// hit, leaves no partial output. Because extraction always starts from an empty
// directory, it can't follow symlinks planted in an existing dst (SEC-008). An
// existing dst, which the CLI allows only with --force, is replaced as a whole, not
// merged into; that is refused when dst holds the archive src itself.
func extractToDir(src, dst string, extract func(dir string) error) (err error) {
	dst = filepath.Clean(dst)
	if err := checkInputOutsideOutput(src, dst); err != nil {
		return err
	}
	parent := filepath.Dir(dst)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("create parent directory: %w", err)
	}
	tmp, err := os.MkdirTemp(parent, "."+filepath.Base(dst)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(tmp)
		}
	}()

	if err := extract(tmp); err != nil {
		return err
	}
	return replacePath(tmp, dst)
}

// replacePath renames newPath to dst. An existing dst is first moved aside and is
// removed only once newPath is in place, so a failed rename leaves it where it was.
func replacePath(newPath, dst string) error {
	if _, err := os.Lstat(dst); errors.Is(err, fs.ErrNotExist) {
		if err := os.Rename(newPath, dst); err != nil {
			return fmt.Errorf("move output into place: %w", err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("check output path: %w", err)
	}

	old := newPath + ".old"
	if err := os.Rename(dst, old); err != nil {
		return fmt.Errorf("move existing output aside: %w", err)
	}
	if err := os.Rename(newPath, dst); err != nil {
		if restoreErr := os.Rename(old, dst); restoreErr != nil {
			return errors.Join(fmt.Errorf("move output into place: %w", err),
				fmt.Errorf("restore previous output from %s: %w", old, restoreErr))
		}
		return fmt.Errorf("move output into place: %w", err)
	}
	if err := os.RemoveAll(old); err != nil {
		return fmt.Errorf("output written to %s, but removing the previous version at %s failed: %w", dst, old, err)
	}
	return nil
}

// checkInputOutsideOutput returns ErrInputInsideOutput when dst is an existing
// directory (not a symlink, which would be replaced rather than followed) that
// contains src, because replacing dst would delete src.
func checkInputOutsideOutput(src, dst string) error {
	info, err := os.Lstat(dst)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("check output path: %w", err)
	}
	if !info.IsDir() {
		return nil
	}
	realDst, err := filepath.EvalSymlinks(dst)
	if err != nil {
		return fmt.Errorf("resolve output path: %w", err)
	}
	realSrc, err := filepath.EvalSymlinks(src)
	if err != nil {
		return fmt.Errorf("resolve input path: %w", err)
	}
	inside, err := pathWithin(realDst, realSrc)
	if err != nil {
		return err
	}
	if inside {
		return fmt.Errorf("%w: %s", ErrInputInsideOutput, dst)
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
	inside, err := pathWithin(srcDir, dst)
	if err != nil {
		return err
	}
	if inside {
		return fmt.Errorf("%w: %s", ErrOutputInsideInput, dst)
	}
	return nil
}

// pathWithin reports whether path is dir itself or lies inside it, comparing
// absolute paths lexically.
func pathWithin(dir, path string) (bool, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false, fmt.Errorf("resolve path %s: %w", dir, err)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, fmt.Errorf("resolve path %s: %w", path, err)
	}
	rel, err := filepath.Rel(absDir, absPath)
	if err != nil {
		return false, nil // e.g. different Windows volumes: path cannot be inside dir
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)), nil
}

// atomicFile is an output written to a hidden temporary file next to its destination
// and renamed into place by Commit, so the destination is either left as it was or
// fully written, never partially.
type atomicFile struct {
	*os.File
	dst       string
	committed bool
}

// createAtomicFile starts an atomic write of dst. The temporary file has mode 0600,
// the mode the tool uses for all single-file outputs.
func createAtomicFile(dst string) (*atomicFile, error) {
	f, err := os.CreateTemp(filepath.Dir(dst), "."+filepath.Base(dst)+".*.tmp")
	if err != nil {
		return nil, err
	}
	return &atomicFile{File: f, dst: dst}, nil
}

// Commit flushes the temporary file to disk and renames it to the destination,
// replacing any file already there.
func (a *atomicFile) Commit() error {
	if err := a.Sync(); err != nil {
		return err
	}
	if err := a.Close(); err != nil {
		return err
	}
	if err := os.Rename(a.Name(), a.dst); err != nil {
		return err
	}
	a.committed = true
	return nil
}

// Abort removes the temporary file unless Commit succeeded. It is safe to defer.
func (a *atomicFile) Abort() {
	if a.committed {
		return
	}
	_ = a.Close()
	_ = os.Remove(a.Name())
}

// writeFileAtomic writes data to dst through a temporary file and a rename.
func writeFileAtomic(dst string, data []byte) error {
	f, err := createAtomicFile(dst)
	if err != nil {
		return err
	}
	defer f.Abort()
	if _, err := f.Write(data); err != nil {
		return err
	}
	return f.Commit()
}

func closeWithError(target *error, closer io.Closer, message string) {
	if closer == nil {
		return
	}
	if err := closer.Close(); err != nil && *target == nil {
		*target = fmt.Errorf("%s: %w", message, err)
	}
}
