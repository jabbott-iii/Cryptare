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
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//--------------------------------------------------core-------------------------------------------------------------------------------------------------//

// Database owns the gorm connection for internal data access.
type Database struct {
	conn *gorm.DB
}

// NewDatabase opens (or creates) the sqlite file and runs schema migrations.
func NewDatabase(path string) (*Database, error) {
	if path == "" {
		path = "rete.db"
	}

	conn, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
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

// SaveKey persists a key record.
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
	return d.conn.Where("key_id = ?", keyID).Delete(&KeyModel{}).Error
}
