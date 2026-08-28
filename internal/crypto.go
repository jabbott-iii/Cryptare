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
	"time"

	"golang.org/x/crypto/pbkdf2"
)

const (
	saltLen   = 16
	ivLen     = 12
	keyLen    = 32
	pbkdf2Iter = 100_000
	encExt    = ".enc"
)

//--------------------------------------------------core-------------------------------------------------------------------------------------------------//

// deriveKey derives a 32-byte AES key from a password and salt using PBKDF2-SHA256.
func deriveKey(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, pbkdf2Iter, keyLen, sha256.New)
}

// EncryptFile encrypts src with AES-256-GCM using password, writing to dst.
// If dst is empty, the output path is src + ".enc".
func EncryptFile(src, dst, password string) error {
	plaintext, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read source file: %w", err)
	}

	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("generate salt: %w", err)
	}

	iv := make([]byte, ivLen)
	if _, err := rand.Read(iv); err != nil {
		return fmt.Errorf("generate iv: %w", err)
	}

	key := deriveKey(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("create GCM: %w", err)
	}

	ciphertext := gcm.Seal(nil, iv, plaintext, nil)

	// Layout: [salt (16)] [iv (12)] [ciphertext]
	out := make([]byte, 0, saltLen+ivLen+len(ciphertext))
	out = append(out, salt...)
	out = append(out, iv...)
	out = append(out, ciphertext...)

	if dst == "" {
		dst = src + encExt
	}

	if err := os.WriteFile(dst, out, 0o600); err != nil {
		return fmt.Errorf("write output file: %w", err)
	}
	return nil
}

// DecryptFile decrypts an AES-256-GCM encrypted file at src using password,
// writing plaintext to dst.  If dst is empty the ".enc" suffix is stripped.
func DecryptFile(src, dst, password string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read source file: %w", err)
	}

	if len(data) < saltLen+ivLen {
		return errors.New("file too small to be a valid encrypted file")
	}

	salt := data[:saltLen]
	iv := data[saltLen : saltLen+ivLen]
	ciphertext := data[saltLen+ivLen:]

	key := deriveKey(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return errors.New("decryption failed: wrong password or corrupted file")
	}

	if dst == "" {
		dst = src
		if filepath.Ext(dst) == encExt {
			dst = dst[:len(dst)-len(encExt)]
		} else {
			dst = dst + ".dec"
		}
	}

	if err := os.WriteFile(dst, plaintext, 0o600); err != nil {
		return fmt.Errorf("write output file: %w", err)
	}
	return nil
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
func EncryptKeyBlob(rawKey []byte, masterPassword string) (string, error) {
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

// KeyExport is the JSON-serialisable export envelope written to disk.
type KeyExport struct {
	Version       int    `json:"version"`
	KeyID         string `json:"key_id"`
	Algorithm     string `json:"algorithm"`
	CreatedAt     int64  `json:"created_at"`
	EncryptedBlob string `json:"encrypted_blob"` // base64 AES-256-GCM ciphertext
}

// ExportKeyToFile writes an encrypted key export to path using masterPassword.
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
		path = fmt.Sprintf("%s-%d.ckey", km.KeyID, time.Now().Unix())
	}

	return os.WriteFile(path, []byte(blob), 0o600)
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

	return &KeyModel{
		KeyID:         export.KeyID,
		Algorithm:     export.Algorithm,
		EncryptedBlob: export.EncryptedBlob,
		CreatedAt_:    export.CreatedAt,
	}, nil
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
