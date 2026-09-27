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
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/pbkdf2"
)

const (
	saltLen                  = 16
	ivLen                    = 12
	keyLen                   = 32
	pbkdf2Iter               = 100_000
	encExt                   = ".enc"
	directoryArtifactMagicV1 = "CRYPTARE-DIR-ENC\x00"
)

// MinPasswordLength is the minimum length, in Unicode code points, of a password that
// protects new data. It follows NIST SP 800-63B-4 section 3.1.1.2, which requires at
// least 15 characters for a password used on its own (single-factor).
const MinPasswordLength = 15

var (
	// ErrEmptyPassword is returned when an encryption operation is given an empty
	// password. Decryption still accepts one so that files, archives and key blobs
	// created before this check was added remain readable.
	ErrEmptyPassword = errors.New("password must not be empty")

	// ErrWeakPassword is returned when a password that would protect new data fails
	// the password policy (see CheckPasswordPolicy).
	ErrWeakPassword = errors.New("password is too weak")

	// ErrPasswordMismatch is returned by the CLI and TUI when a new password and its
	// confirmation differ.
	ErrPasswordMismatch = errors.New("passwords do not match")

	// ErrInvalidKeyExport is returned by ImportKeyFromFile when an export's contents
	// aren't what Cryptare writes (see validateKeyExport).
	ErrInvalidKeyExport = errors.New("invalid key export")
)

//--------------------------------------------------core-------------------------------------------------------------------------------------------------//

// deriveKey derives a 32-byte AES key from a password and salt using PBKDF2-SHA256.
func deriveKey(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, pbkdf2Iter, keyLen, sha256.New)
}

// CheckPasswordPolicy reports whether password may protect new data: encrypted files
// and directories, stored keys and key exports. It returns ErrEmptyPassword for an
// empty password and ErrWeakPassword for one that is shorter than MinPasswordLength
// Unicode code points or is a single character repeated. There are no composition
// rules and no maximum length.
//
// Passwords used to decrypt or import are not checked, so data protected before the
// policy existed stays readable.
func CheckPasswordPolicy(password string) error {
	if password == "" {
		return ErrEmptyPassword
	}
	if utf8.RuneCountInString(password) < MinPasswordLength {
		return fmt.Errorf("%w: use at least %d characters", ErrWeakPassword, MinPasswordLength)
	}
	// Compare bytes so that invalid UTF-8 (which decodes to RuneError) isn't mistaken
	// for a repeated character.
	_, size := utf8.DecodeRuneInString(password)
	if len(password)%size == 0 && strings.Repeat(password[:size], len(password)/size) == password {
		return fmt.Errorf("%w: it is one character repeated", ErrWeakPassword)
	}
	return nil
}

// EncryptFile encrypts src with AES-256-GCM using password, writing to dst.
// If dst is empty, the output path is src + ".enc". The password must meet the
// password policy (CheckPasswordPolicy).
func EncryptFile(src, dst, password string) error {
	if err := CheckPasswordPolicy(password); err != nil {
		return err
	}

	info, err := os.Lstat(src)
	if err != nil {
		return fmt.Errorf("lstat source path: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("encrypt input: symlinks are not supported (%s)", src)
	}

	statInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat source path: %w", err)
	}

	if dst == "" {
		dst = src + encExt
	}
	if err := checkNotSameFile(src, dst); err != nil {
		return err
	}

	if statInfo.IsDir() {
		return encryptDirectory(src, dst, password)
	}

	plaintext, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read source file: %w", err)
	}

	ciphertext, err := encryptBytes(plaintext, password)
	if err != nil {
		return err
	}

	if err := writeFileAtomic(dst, ciphertext); err != nil {
		return fmt.Errorf("write output file: %w", err)
	}
	return nil
}

// DecryptFile decrypts an AES-256-GCM encrypted file at src using password,
// writing plaintext to dst.  If dst is empty, the ".enc" suffix is stripped.
// An encrypted directory is restored under DefaultExtractLimits.
func DecryptFile(src, dst, password string) error {
	return DecryptFileWithLimits(src, dst, password, DefaultExtractLimits())
}

// DecryptFileWithLimits is DecryptFile with explicit limits for restoring an
// encrypted directory, which is extracted like an archive (see
// DecompressFileWithLimits). Single files are not affected: their plaintext is
// never larger than the encrypted file.
func DecryptFileWithLimits(src, dst, password string, limits ExtractLimits) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read source file: %w", err)
	}

	if dst == "" {
		dst = defaultDecryptOutput(src)
	}
	if err := checkNotSameFile(src, dst); err != nil {
		return err
	}

	if bytes.HasPrefix(data, []byte(directoryArtifactMagicV1)) {
		archive, err := decryptBytesWithAAD(data[len(directoryArtifactMagicV1):], password, []byte(directoryArtifactMagicV1))
		if err != nil {
			return err
		}
		return restoreDirectoryArchive(archive, src, dst, limits)
	}

	plaintext, err := decryptBytes(data, password)
	if err != nil {
		return err
	}

	if err := writeFileAtomic(dst, plaintext); err != nil {
		return fmt.Errorf("write output file: %w", err)
	}
	return nil
}

func encryptBytes(plaintext []byte, password string) ([]byte, error) {
	return encryptBytesWithAAD(plaintext, password, nil)
}

func encryptBytesWithAAD(plaintext []byte, password string, aad []byte) ([]byte, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("generate salt: %w", err)
	}

	iv := make([]byte, ivLen)
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("generate iv: %w", err)
	}

	key := deriveKey(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	ciphertext := gcm.Seal(nil, iv, plaintext, aad)

	// Layout: [salt (16)] [iv (12)] [ciphertext]
	out := make([]byte, 0, saltLen+ivLen+len(ciphertext))
	out = append(out, salt...)
	out = append(out, iv...)
	out = append(out, ciphertext...)
	return out, nil
}

func decryptBytes(data []byte, password string) ([]byte, error) {
	return decryptBytesWithAAD(data, password, nil)
}

func decryptBytesWithAAD(data []byte, password string, aad []byte) ([]byte, error) {
	if len(data) < saltLen+ivLen {
		return nil, errors.New("file too small to be a valid encrypted file")
	}

	salt := data[:saltLen]
	iv := data[saltLen : saltLen+ivLen]
	ciphertext := data[saltLen+ivLen:]

	key := deriveKey(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, iv, ciphertext, aad)
	if err != nil {
		return nil, errors.New("decryption failed: wrong password or corrupted file")
	}
	return plaintext, nil
}

func encryptDirectory(src, dst, password string) error {
	plaintext, err := buildDirectoryArchive(src)
	if err != nil {
		return err
	}

	ciphertext, err := encryptBytesWithAAD(plaintext, password, []byte(directoryArtifactMagicV1))
	if err != nil {
		return err
	}

	payload := make([]byte, 0, len(directoryArtifactMagicV1)+len(ciphertext))
	payload = append(payload, directoryArtifactMagicV1...)
	payload = append(payload, ciphertext...)
	if err := writeFileAtomic(dst, payload); err != nil {
		return fmt.Errorf("write output file: %w", err)
	}
	return nil
}

// buildDirectoryArchive returns a tar.gz of the directory tree at src. It is built in
// memory, so no plaintext copy of the directory is written to disk (SEC-006); the
// whole archive has to be in memory for encryption anyway.
func buildDirectoryArchive(src string) ([]byte, error) {
	info, err := os.Lstat(src)
	if err != nil {
		return nil, fmt.Errorf("lstat source path: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("encrypt directory: symlinks are not supported (%s)", src)
	}

	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.DefaultCompression)
	if err != nil {
		return nil, fmt.Errorf("create gzip writer: %w", err)
	}
	gz.Name = filepath.Base(filepath.Clean(src)) + ".tar"

	if _, err := writeTarGz(gz, src); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("finalise gzip: %w", err)
	}
	return buf.Bytes(), nil
}

// restoreDirectoryArchive extracts the decrypted tar.gz of an encrypted directory
// (read from src) into dst, through extractToDir and within limits.
func restoreDirectoryArchive(archive []byte, src, dst string, limits ExtractLimits) (err error) {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return fmt.Errorf("read decrypted directory archive: %w", err)
	}
	defer closeWithError(&err, gz, "close directory archive reader")

	budget := &extractBudget{limits: limits}
	return extractToDir(src, dst, func(dir string) error {
		return extractTarGz(gz, dir, budget)
	})
}

//--------------------------------------------------key management---------------------------------------------------------------------------------------//

// GenerateKey generates a random 32-byte AES key.
func GenerateKey() ([]byte, error) {
	key := make([]byte, keyLen)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	return key, nil
}

// EncryptKeyBlob encrypts rawKey with masterPassword and returns a base64 blob.
// masterPassword must meet the password policy (CheckPasswordPolicy).
func EncryptKeyBlob(rawKey []byte, masterPassword string) (string, error) {
	if err := CheckPasswordPolicy(masterPassword); err != nil {
		return "", err
	}

	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	iv := make([]byte, ivLen)
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("generate iv: %w", err)
	}

	key := deriveKey(masterPassword, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	ct := gcm.Seal(nil, iv, rawKey, nil)

	out := make([]byte, 0, saltLen+ivLen+len(ct))
	out = append(out, salt...)
	out = append(out, iv...)
	out = append(out, ct...)

	return base64.StdEncoding.EncodeToString(out), nil
}

// DecryptKeyBlob decrypts a base64 blob previously produced by EncryptKeyBlob.
func DecryptKeyBlob(blob, masterPassword string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(blob)
	if err != nil {
		return nil, fmt.Errorf("decode blob: %w", err)
	}

	if len(data) < saltLen+ivLen {
		return nil, errors.New("blob too small")
	}

	salt := data[:saltLen]
	iv := data[saltLen : saltLen+ivLen]
	ct := data[saltLen+ivLen:]

	key := deriveKey(masterPassword, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	rawKey, err := gcm.Open(nil, iv, ct, nil)
	if err != nil {
		return nil, errors.New("decryption failed: wrong master password or corrupted blob")
	}

	return rawKey, nil
}

//--------------------------------------------------key export/import------------------------------------------------------------------------------------//

// KeyExport is the JSON-serializable export envelope written to disk.
type KeyExport struct {
	Version       int    `json:"version"`
	KeyID         string `json:"key_id"`
	Algorithm     string `json:"algorithm"`
	CreatedAt     int64  `json:"created_at"`
	EncryptedBlob string `json:"encrypted_blob"` // base64 AES-256-GCM ciphertext
}

// ExportKeyToFile writes an encrypted key export to the path using masterPassword,
// which must meet the password policy (CheckPasswordPolicy).
func ExportKeyToFile(km *KeyModel, masterPassword, path string) error {
	export := KeyExport{
		Version:       1,
		KeyID:         km.KeyID,
		Algorithm:     km.Algorithm,
		CreatedAt:     km.CreatedAt_,
		EncryptedBlob: km.EncryptedBlob,
	}

	raw, err := json.Marshal(export)
	if err != nil {
		return fmt.Errorf("marshal export: %w", err)
	}

	blob, err := EncryptKeyBlob(raw, masterPassword)
	if err != nil {
		return fmt.Errorf("encrypt export: %w", err)
	}

	if path == "" {
		path = defaultExportPath(km.KeyID)
	}

	if err := writeFileAtomic(path, []byte(blob)); err != nil {
		return fmt.Errorf("write export file: %w", err)
	}
	return nil
}

// timeNow is the clock used for default export file names; tests replace it.
var timeNow = time.Now

// defaultExportPath returns the default export file name for keyID,
// <key-id>-<unix time>.ckey in the current directory. Callers compute it once and pass
// it on, so the name they report is the file written (BUG-006).
func defaultExportPath(keyID string) string {
	return fmt.Sprintf("%s-%d.ckey", keyID, timeNow().Unix())
}

// defaultDecryptOutput strips the ".enc" extension from src, in any letter case
// (BUG-007), or appends ".dec" when there is none.
func defaultDecryptOutput(src string) string {
	if len(src) > len(encExt) && hasSuffixFold(src, encExt) {
		return src[:len(src)-len(encExt)]
	}
	return src + ".dec"
}

// ImportKeyFromFile reads an export file and returns a KeyModel (not yet persisted).
func ImportKeyFromFile(path, masterPassword string) (*KeyModel, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read export file: %w", err)
	}

	plaintext, err := DecryptKeyBlob(string(raw), masterPassword)
	if err != nil {
		return nil, fmt.Errorf("decrypt export file: %w", err)
	}

	// Parse minimal JSON manually to avoid import cycle (encoding/json is fine here).
	var export KeyExport
	if err := unmarshalKeyExport(plaintext, &export); err != nil {
		return nil, fmt.Errorf("parse export: %w", err)
	}
	if err := validateKeyExport(export); err != nil {
		return nil, err
	}

	return &KeyModel{
		KeyID:         export.KeyID,
		Algorithm:     export.Algorithm,
		EncryptedBlob: export.EncryptedBlob,
		CreatedAt_:    export.CreatedAt,
	}, nil
}

const (
	keyExportVersion = 1
	keyAlgorithm     = "AES-256-GCM"
	keyIDLen         = 16 // hex characters, from newKeyID
	gcmTagLen        = 16
	// storedKeyBlobLen is the decoded size of a stored key blob from EncryptKeyBlob:
	// salt, nonce, and the 32-byte key sealed with a GCM tag.
	storedKeyBlobLen = saltLen + ivLen + keyLen + gcmTagLen
)

// validateKeyExport checks an imported export's fields before they are stored and
// shown by keys list and the TUI (SEC-011): the version must be 1, the key ID 16
// lower-case hex characters as newKeyID makes, the algorithm AES-256-GCM, and the
// encrypted key a base64 blob of the size EncryptKeyBlob produces. Rejected values
// are quoted with %q, so control characters in them are escaped, not printed.
func validateKeyExport(e KeyExport) error {
	if e.Version != keyExportVersion {
		return fmt.Errorf("%w: unsupported version %d", ErrInvalidKeyExport, e.Version)
	}
	if !isKeyID(e.KeyID) {
		return fmt.Errorf("%w: key ID %q is not %d lower-case hexadecimal characters", ErrInvalidKeyExport, e.KeyID, keyIDLen)
	}
	if e.Algorithm != keyAlgorithm {
		return fmt.Errorf("%w: unsupported algorithm %q", ErrInvalidKeyExport, e.Algorithm)
	}
	blob, err := base64.StdEncoding.DecodeString(e.EncryptedBlob)
	if err != nil || len(blob) != storedKeyBlobLen {
		return fmt.Errorf("%w: the encrypted key is malformed", ErrInvalidKeyExport)
	}
	return nil
}

// isKeyID reports whether s has the format of a generated key ID.
func isKeyID(s string) bool {
	if len(s) != keyIDLen {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

// unmarshalKeyExport parses a JSON export envelope.
func unmarshalKeyExport(data []byte, out *KeyExport) error {
	return json.Unmarshal(data, out)
}

// newKeyID generates a random hex key ID.
func newKeyID() (string, error) {
	b := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}
