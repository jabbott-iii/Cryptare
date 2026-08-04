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

)

type Database struct {
    conn *gorm.DB
}

type Actions interface {
    GetFile()
    EncryptFile()
    DecryptFile()
    ExportFileKey()
    ImportFileKey()
}

// ---------------------------------------- MODELS ----------------------------------------- //
type ItemModel struct {
    ID        uint      `gorm:"primaryKey"`
    FileName  string    `gorm:"size:255;not null"`
    Encrypted bool      `gorm:"default:false;not null"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
// --------------------------------- Database Functionality -------------------------------- //

// NewDatabase opens (or creates) the sqlite file and runs migrations.
func NewDatabase(path string) (*Database, error) {
	if path == "" {
		path = "cryptare.db"
	}

	conn, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	// Persistence structs
	if err := conn.AutoMigrate(&ItemModel{}); err != nil {
		return nil, fmt.Errorf("auto-migrate schema: %w", err)
	}

	return &Database{conn: conn}, nil
}
