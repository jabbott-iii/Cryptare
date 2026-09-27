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
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"runtime"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var ErrKeyNotFound = errors.New("key not found")

//--------------------------------------------------core-------------------------------------------------------------------------------------------------//

// Database owns the gorm connection for internal data access.
type Database struct {
	conn *gorm.DB
}

// NewDatabase opens (or creates) the sqlite file and runs schema migrations.
// Every connection has SQLite's secure_delete on, so a deleted key's encrypted blob
// is overwritten in the file instead of being left in free space (SEC-009).
func NewDatabase(path string) (*Database, error) {
	if path == "" {
		path = "cryptare.db"
	}
	if err := prepareDatabaseFile(path); err != nil {
		return nil, err
	}

	conn, err := gorm.Open(sqlite.Open(withSecureDelete(path)), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	if err := conn.AutoMigrate(
		&KeyModel{},
	); err != nil {
		return nil, fmt.Errorf("auto-migrate schema: %w", err)
	}

	return &Database{conn: conn}, nil
}

// prepareDatabaseFile makes the key database private to its owner (SEC-010). A new
// file is created with mode 0600 before SQLite opens it; SQLite gives its journal
// files the same mode. An existing database, or a journal left next to it, that others
// can read is set to 0600. Files owned by someone else are left as they are.
// In-memory databases and SQLite URI paths are left to SQLite.
func prepareDatabaseFile(path string) error {
	if path == ":memory:" || strings.HasPrefix(path, "file:") {
		return nil
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err == nil {
		if err := f.Close(); err != nil {
			return fmt.Errorf("create database file: %w", err)
		}
		return nil
	}
	if !errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("create database file: %w", err)
	}
	if runtime.GOOS == "windows" {
		return nil // Unix permission bits don't apply
	}

	for _, p := range []string{path, path + "-journal", path + "-wal", path + "-shm"} {
		info, err := os.Stat(p)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("check database file: %w", err)
		}
		if info.Mode().Perm()&0o077 == 0 {
			continue
		}
		if err := os.Chmod(p, 0o600); err != nil && !errors.Is(err, fs.ErrPermission) {
			return fmt.Errorf("restrict database file permissions: %w", err)
		}
	}
	return nil
}

// withSecureDelete adds go-sqlite3's _secure_delete=on parameter to a database path.
// The driver applies it to each new connection, which a one-off PRAGMA would not do
// for a connection pool.
func withSecureDelete(path string) string {
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	return path + sep + "_secure_delete=on"
}

// Conn exposes the raw gorm handle for advanced queries/transactions.
func (d *Database) Conn() *gorm.DB {
	return d.conn
}

//-----------------------------------------------------------models and types------------------------------------------------------------------------------------------------//

// KeyModel persists an encryption key (always stored encrypted) together with its metadata.
type KeyModel struct {
	gorm.Model
	KeyID         string `gorm:"uniqueIndex;not null"`
	Algorithm     string `gorm:"not null"`
	EncryptedBlob string `gorm:"not null"` // base64-encoded AES-256-GCM ciphertext
	CreatedAt_    int64  `gorm:"column:created_epoch"`
}

//-----------------------------------------------------------database operations--------------------------------------------------------------------------------------------//

type Storage interface {
	SaveKey(k *KeyModel) error
	ListKeys() ([]KeyModel, error)
	GetKey(keyID string) (*KeyModel, error)
	DeleteKey(keyID string) error
}

// SaveKey persists as a key record.
func (d *Database) SaveKey(k *KeyModel) error {
	return d.conn.Save(k).Error
}

// ListKeys returns all stored key records.
func (d *Database) ListKeys() ([]KeyModel, error) {
	var keys []KeyModel
	if err := d.conn.Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

// GetKey returns a key record by its KeyID.
func (d *Database) GetKey(keyID string) (*KeyModel, error) {
	var k KeyModel
	if err := d.conn.Where("key_id = ?", keyID).First(&k).Error; err != nil {
		return nil, err
	}
	return &k, nil
}

// DeleteKey removes a key record by its KeyID.
func (d *Database) DeleteKey(keyID string) error {
	sqlDB, err := d.conn.DB()
	if err != nil {
		return err
	}
	tx, err := sqlDB.Begin()
	if err != nil {
		return err
	}

	var blob []byte
	if err := tx.QueryRow("DELETE FROM key_models WHERE key_id = ? RETURNING encrypted_blob", keyID).Scan(&blob); err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return ErrKeyNotFound
		}
		return err
	}
	defer zeroBytes(blob)

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return err
	}
	return nil
}

func zeroBytes(buf []byte) {
	for i := range buf {
		buf[i] = 0
	}
}
