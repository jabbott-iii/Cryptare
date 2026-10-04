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
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

// Version 2 of Cryptare's encrypted format (plan 3.1 and 3.2, SEC-005, BUG-010).
//
// Every artifact written by this version (encrypted files and folders, stored keys and
// key exports) starts with a 46-byte header:
//
//	offset size field
//	0      9    magic "CRYPTARE\x00" (a legacy folder artifact has '-' at offset 8)
//	9      1    format version, 2
//	10     1    content type: 1 file, 2 folder (tar.gz), 3 stored key, 4 key export
//	11     1    key source: 1 password (2 is reserved for stored keys, plan 3.4)
//	12     1    KDF: 1 Argon2id over the password, 2 Argon2id over its NFKC form (Q-009)
//	13     4    Argon2id memory in KiB (big-endian)
//	17     4    Argon2id passes (big-endian)
//	21     1    Argon2id lanes
//	22     16   salt
//	38     1    chunk size as a power of two (16 = 64 KiB)
//	39     7    random nonce prefix
//
// The plaintext follows in chunks sealed with AES-256-GCM under the derived key: every
// chunk holds chunk-size bytes except the last, which may be shorter or empty. A chunk's
// nonce is the prefix, a 4-byte big-endian chunk counter and a byte that is 1 for the
// last chunk and 0 otherwise, and its additional data is the whole header. Reordering,
// dropping, truncating or appending chunks, or changing the header, therefore fails
// authentication (the STREAM construction).

const (
	v2Magic        = "CRYPTARE\x00"
	formatV2       = 2
	v2HeaderLen    = 46
	noncePrefixLen = 7
	gcmTagLen      = 16

	contentFile      byte = 1
	contentFolder    byte = 2
	contentStoredKey byte = 3
	contentKeyExport byte = 4

	keySourcePassword byte = 1

	// Key derivation identifiers. Both are Argon2id; they differ in the bytes of the
	// password it is applied to (Q-009). This version writes kdfArgon2id when the
	// normalised password (normalizePassword) is the password as given, so the header
	// is the one earlier versions write and they can read the data with the same
	// password. Otherwise it writes kdfArgon2idNFKC, which earlier versions refuse as
	// an unknown key derivation instead of reporting a wrong password. Data recorded as
	// kdfArgon2id is read with each form passwordCandidates returns, kdfArgon2idNFKC
	// data with the normalised form only.
	kdfArgon2id     byte = 1
	kdfArgon2idNFKC byte = 2

	defaultChunkShift = 16 // 64 KiB
	minChunkShift     = 10
	maxChunkShift     = 24

	// Limits on the Argon2id settings a file may ask for, so that decrypting a crafted
	// file can't take unbounded memory or time.
	maxArgonMemoryKiB  = 1 << 20 // 1 GiB
	maxArgonIterations = 10
	maxArgonThreads    = 16
)

// argon2Params are the Argon2id settings recorded in a version 2 header.
type argon2Params struct {
	memoryKiB  uint32
	iterations uint32
	threads    uint8
}

// defaultPasswordKDF is RFC 9106's second recommended Argon2id setting: 64 MiB of
// memory, 3 passes and 4 lanes (owner decision for plan 3.1).
var defaultPasswordKDF = argon2Params{memoryKiB: 64 * 1024, iterations: 3, threads: 4}

// passwordKDF is the setting used for new data. Tests lower it to keep them fast; the
// setting is recorded in each header, so reading doesn't depend on it.
var passwordKDF = defaultPasswordKDF

// errUnsupportedFormat is returned for version 2 data this build can't read.
var errUnsupportedFormat = errors.New("unsupported encrypted data")

// errDecrypt is returned when authentication fails: a wrong password or damaged data.
var errDecrypt = errors.New("decryption failed: wrong password or corrupted file")

// v2Header is a parsed version 2 header.
type v2Header struct {
	content     byte
	kdfID       byte
	kdf         argon2Params
	salt        [saltLen]byte
	chunkShift  uint8
	noncePrefix [noncePrefixLen]byte
}

// isV2 reports whether data starts like a version 2 artifact.
func isV2(data []byte) bool {
	return bytes.HasPrefix(data, []byte(v2Magic))
}

// newV2Header returns a header for new content, with a fresh salt and nonce prefix and
// the current password KDF setting, recorded as kdfArgon2id.
func newV2Header(content byte) (v2Header, error) {
	h := v2Header{content: content, kdfID: kdfArgon2id, kdf: passwordKDF, chunkShift: defaultChunkShift}
	if _, err := rand.Read(h.salt[:]); err != nil {
		return v2Header{}, fmt.Errorf("generate salt: %w", err)
	}
	if _, err := rand.Read(h.noncePrefix[:]); err != nil {
		return v2Header{}, fmt.Errorf("generate nonce: %w", err)
	}
	return h, nil
}

func (h v2Header) marshal() []byte {
	b := make([]byte, 0, v2HeaderLen)
	b = append(b, v2Magic...)
	b = append(b, formatV2, h.content, keySourcePassword, h.kdfID)
	b = binary.BigEndian.AppendUint32(b, h.kdf.memoryKiB)
	b = binary.BigEndian.AppendUint32(b, h.kdf.iterations)
	b = append(b, h.kdf.threads)
	b = append(b, h.salt[:]...)
	b = append(b, h.chunkShift)
	b = append(b, h.noncePrefix[:]...)
	return b
}

// parseV2Header parses and checks a version 2 header. It refuses versions, content
// types, key sources and KDFs this build doesn't know, and settings beyond its limits.
func parseV2Header(b []byte) (v2Header, error) {
	if len(b) < v2HeaderLen || !isV2(b) {
		return v2Header{}, fmt.Errorf("%w: header too short or not a Cryptare header", errUnsupportedFormat)
	}
	if b[9] != formatV2 {
		return v2Header{}, fmt.Errorf("%w: format version %d", errUnsupportedFormat, b[9])
	}
	h := v2Header{content: b[10]}
	if h.content < contentFile || h.content > contentKeyExport {
		return v2Header{}, fmt.Errorf("%w: content type %d", errUnsupportedFormat, h.content)
	}
	if b[11] != keySourcePassword {
		return v2Header{}, fmt.Errorf("%w: key source %d", errUnsupportedFormat, b[11])
	}
	h.kdfID = b[12]
	if h.kdfID != kdfArgon2id && h.kdfID != kdfArgon2idNFKC {
		return v2Header{}, fmt.Errorf("%w: key derivation %d", errUnsupportedFormat, h.kdfID)
	}
	h.kdf = argon2Params{
		memoryKiB:  binary.BigEndian.Uint32(b[13:17]),
		iterations: binary.BigEndian.Uint32(b[17:21]),
		threads:    b[21],
	}
	if err := h.kdf.check(); err != nil {
		return v2Header{}, err
	}
	copy(h.salt[:], b[22:38])
	h.chunkShift = b[38]
	if h.chunkShift < minChunkShift || h.chunkShift > maxChunkShift {
		return v2Header{}, fmt.Errorf("%w: chunk size 2^%d", errUnsupportedFormat, h.chunkShift)
	}
	copy(h.noncePrefix[:], b[39:46])
	return h, nil
}

// check enforces the limits on Argon2id settings read from a file.
func (p argon2Params) check() error {
	switch {
	case p.threads < 1 || p.threads > maxArgonThreads:
		return fmt.Errorf("%w: Argon2id lanes %d (allowed 1–%d)", errUnsupportedFormat, p.threads, maxArgonThreads)
	case p.iterations < 1 || p.iterations > maxArgonIterations:
		return fmt.Errorf("%w: Argon2id passes %d (allowed 1–%d)", errUnsupportedFormat, p.iterations, maxArgonIterations)
	case p.memoryKiB < 8*uint32(p.threads) || p.memoryKiB > maxArgonMemoryKiB:
		return fmt.Errorf("%w: Argon2id memory %d KiB (allowed %d KiB–%s)", errUnsupportedFormat, p.memoryKiB, 8*uint32(p.threads), formatSize(maxArgonMemoryKiB*1024))
	}
	return nil
}

// aead derives the header's key from password with Argon2id and returns the cipher.
func (h v2Header) aead(password string) (cipher.AEAD, error) {
	key := argon2.IDKey([]byte(password), h.salt[:], h.kdf.iterations, h.kdf.memoryKiB, h.kdf.threads, keyLen)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}
	return gcm, nil
}

func (h v2Header) chunkSize() int { return 1 << h.chunkShift }

// chunkNonce returns the nonce for chunk number counter.
func (h v2Header) chunkNonce(nonce []byte, counter uint32, last bool) []byte {
	nonce = append(nonce[:0], h.noncePrefix[:]...)
	nonce = binary.BigEndian.AppendUint32(nonce, counter)
	if last {
		return append(nonce, 1)
	}
	return append(nonce, 0)
}

//--------------------------------------------------encrypting writer------------------------------------------------------------------------------------//

// encryptingWriter encrypts what is written to it into version 2 chunks. It holds back
// one chunk of plaintext so that Close can seal the last chunk as the final one.
type encryptingWriter struct {
	w       io.Writer
	h       v2Header
	aad     []byte
	aead    cipher.AEAD
	buf     []byte
	out     []byte
	nonce   []byte
	counter uint32
	closed  bool
}

// newEncryptingWriter writes a new version 2 header for content to w and returns a
// writer that encrypts into w with a key derived from the normalised password (Q-009).
// Close must be called to write the final chunk; it doesn't close w.
func newEncryptingWriter(w io.Writer, password string, content byte) (*encryptingWriter, error) {
	h, err := newV2Header(content)
	if err != nil {
		return nil, err
	}
	normalized := normalizePassword(password)
	if normalized != password {
		h.kdfID = kdfArgon2idNFKC
	}
	aead, err := h.aead(normalized)
	if err != nil {
		return nil, err
	}
	aad := h.marshal()
	if _, err := w.Write(aad); err != nil {
		return nil, fmt.Errorf("write header: %w", err)
	}
	size := h.chunkSize()
	return &encryptingWriter{
		w: w, h: h, aad: aad, aead: aead,
		buf: make([]byte, 0, size), out: make([]byte, 0, size+gcmTagLen), nonce: make([]byte, 0, 12),
	}, nil
}

func (e *encryptingWriter) Write(p []byte) (int, error) {
	if e.closed {
		return 0, errors.New("write to closed encrypting writer")
	}
	written := 0
	for len(p) > 0 {
		if len(e.buf) == cap(e.buf) {
			// More data follows a full chunk, so it isn't the last one.
			if err := e.seal(false); err != nil {
				return written, err
			}
		}
		n := copy(e.buf[len(e.buf):cap(e.buf)], p)
		e.buf = e.buf[:len(e.buf)+n]
		p = p[n:]
		written += n
	}
	return written, nil
}

// Close seals the remaining plaintext, possibly none, as the final chunk.
func (e *encryptingWriter) Close() error {
	if e.closed {
		return nil
	}
	e.closed = true
	return e.seal(true)
}

func (e *encryptingWriter) seal(last bool) error {
	if e.counter == ^uint32(0) {
		return errors.New("encrypt: too many chunks")
	}
	e.out = e.aead.Seal(e.out[:0], e.h.chunkNonce(e.nonce, e.counter, last), e.buf, e.aad)
	if _, err := e.w.Write(e.out); err != nil {
		return fmt.Errorf("write encrypted data: %w", err)
	}
	e.counter++
	e.buf = e.buf[:0]
	return nil
}

//--------------------------------------------------decrypting reader------------------------------------------------------------------------------------//

// decryptingReader returns the plaintext of version 2 chunks, releasing each chunk only
// after it has been authenticated. The end of the data is reported only after a chunk
// sealed as the last one; data that stops earlier is reported as corrupted.
type decryptingReader struct {
	r       *bufio.Reader
	h       v2Header
	aad     []byte
	aead    cipher.AEAD
	cbuf    []byte
	plain   []byte
	pending []byte
	nonce   []byte
	counter uint32
	done    bool
	err     error
}

// newDecryptingReader reads a version 2 header from r, checks that its content type is
// one of wanted, and authenticates the first chunk with a key derived from password,
// trying each form of it that may have protected the data (passwordCandidates). It
// returns the header with a reader of the plaintext. A wrong password is therefore
// reported here, before the caller creates any output.
func newDecryptingReader(r *bufio.Reader, password string, wanted ...byte) (v2Header, io.Reader, error) {
	hb := make([]byte, v2HeaderLen)
	if _, err := io.ReadFull(r, hb); err != nil {
		return v2Header{}, nil, fmt.Errorf("%w: header: %v", errUnsupportedFormat, err)
	}
	h, err := parseV2Header(hb)
	if err != nil {
		return v2Header{}, nil, err
	}
	if !bytes.Contains(wanted, []byte{h.content}) {
		return v2Header{}, nil, fmt.Errorf("%w: it holds %s, not %s", errUnsupportedFormat, contentName(h.content), contentName(wanted[0]))
	}
	size := h.chunkSize()
	d := &decryptingReader{
		r: r, h: h, aad: hb,
		cbuf: make([]byte, size+gcmTagLen), plain: make([]byte, 0, size), nonce: make([]byte, 0, 12),
	}
	n, last, err := d.read()
	if err != nil {
		return v2Header{}, nil, err
	}
	for _, candidate := range passwordCandidates(password, h.kdfID == kdfArgon2idNFKC) {
		if d.aead, err = h.aead(candidate); err != nil {
			return v2Header{}, nil, err
		}
		if err = d.open(n, last); err == nil {
			return h, d, nil
		}
	}
	return v2Header{}, nil, err
}

func (d *decryptingReader) Read(p []byte) (int, error) {
	for len(d.pending) == 0 {
		if d.err != nil {
			return 0, d.err
		}
		if d.done {
			return 0, io.EOF
		}
		d.err = d.next()
	}
	n := copy(p, d.pending)
	d.pending = d.pending[n:]
	return n, nil
}

// next reads, authenticates and decrypts the next chunk.
func (d *decryptingReader) next() error {
	n, last, err := d.read()
	if err != nil {
		return err
	}
	return d.open(n, last)
}

// read reads the next chunk into cbuf and returns its length and whether it is the
// final one.
func (d *decryptingReader) read() (n int, last bool, err error) {
	n, err = io.ReadFull(d.r, d.cbuf)
	switch {
	case errors.Is(err, io.EOF):
		return 0, false, fmt.Errorf("%w (data ends before its final chunk)", errDecrypt)
	case errors.Is(err, io.ErrUnexpectedEOF):
		last = true // a short chunk can only be the final one
	case err != nil:
		return 0, false, fmt.Errorf("read encrypted data: %w", err)
	default:
		if _, err := d.r.Peek(1); errors.Is(err, io.EOF) {
			last = true
		} else if err != nil {
			return 0, false, fmt.Errorf("read encrypted data: %w", err)
		}
	}
	if n < gcmTagLen {
		return 0, false, errDecrypt
	}
	return n, last, nil
}

// open authenticates and decrypts the chunk that read left in cbuf. cbuf is left as it
// was, so a failed open can be retried with another key.
func (d *decryptingReader) open(n int, last bool) error {
	plain, err := d.aead.Open(d.plain[:0], d.h.chunkNonce(d.nonce, d.counter, last), d.cbuf[:n], d.aad)
	if err != nil {
		return errDecrypt
	}
	if d.counter == ^uint32(0) {
		return errDecrypt
	}
	d.counter++
	d.pending = plain
	d.done = last
	return nil
}

//--------------------------------------------------small data------------------------------------------------------------------------------------//

// sealV2 encrypts a small plaintext held in memory, such as a key, as version 2 data.
func sealV2(plaintext []byte, password string, content byte) ([]byte, error) {
	var buf bytes.Buffer
	w, err := newEncryptingWriter(&buf, password, content)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(plaintext); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// openV2 decrypts small version 2 data of the given content type held in memory.
func openV2(data []byte, password string, content byte) ([]byte, error) {
	_, r, err := newDecryptingReader(bufio.NewReader(bytes.NewReader(data)), password, content)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(r)
}

// contentName describes a content type in error messages.
func contentName(c byte) string {
	switch c {
	case contentFile:
		return "an encrypted file"
	case contentFolder:
		return "an encrypted folder"
	case contentStoredKey:
		return "a stored key"
	case contentKeyExport:
		return "a key export"
	default:
		return fmt.Sprintf("content type %d", c)
	}
}
