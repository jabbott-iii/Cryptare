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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const gzExt = ".gz"

//--------------------------------------------------core-------------------------------------------------------------------------------------------------//

// CompressFile compresses src with gzip at the given level (1–9), writing to dst.
// If dst is empty, the output path is src + ".gz".
func CompressFile(src, dst string, level int) error {
	if level < gzip.BestSpeed || level > gzip.BestCompression {
		level = gzip.DefaultCompression
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer in.Close()

	if dst == "" {
		dst = src + gzExt
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer out.Close()

	gz, err := gzip.NewWriterLevel(out, level)
	if err != nil {
		return fmt.Errorf("create gzip writer: %w", err)
	}
	gz.Name = filepath.Base(src)

	if _, err := io.Copy(gz, in); err != nil {
		return fmt.Errorf("compress data: %w", err)
	}

	if err := gz.Close(); err != nil {
		return fmt.Errorf("finalise gzip: %w", err)
	}
	return nil
}

// DecompressFile decompresses a gzip file at src, writing to dst.
// If dst is empty, the ".gz" suffix is stripped; otherwise ".dec" is appended.
func DecompressFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer in.Close()

	gz, err := gzip.NewReader(in)
	if err != nil {
		return fmt.Errorf("create gzip reader: %w", err)
	}
	defer gz.Close()

	if dst == "" {
		if strings.HasSuffix(src, gzExt) {
			dst = src[:len(src)-len(gzExt)]
		} else {
			dst = src + ".dec"
		}
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, gz); err != nil {
		return fmt.Errorf("decompress data: %w", err)
	}
	return nil
}
