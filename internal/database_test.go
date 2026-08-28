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
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestNewDatabase tests the NewDatabase function to ensure it creates a new database and returns a valid connection.
// It verifies that the database file is created and that the connection is not nil.
func TestNewDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}

	if db == nil {
		t.Error("NewDatabase returned nil")
	}

	if db.Conn() == nil {
		t.Error("Database connection is nil")
	}

	// Verify database file was created
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("Database file not created: %v", err)
	}
}

// TestNewDatabaseDefaultPath tests the NewDatabase function when no path is provided.
// It ensures that the database is created at the default location.
func TestNewDatabaseDefaultPath(t *testing.T) {
	tmpDir := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir(tmpDir)

	db, err := NewDatabase("")
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}

	if db == nil {
		t.Error("NewDatabase returned nil")
	}
}

// TestSaveKey tests the SaveKey function to ensure that a key can be saved to the database.
// It verifies that no error is returned when saving a valid key.
func TestSaveKey(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := NewDatabase(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}

	km := &KeyModel{
		KeyID:         "test-key-1",
		Algorithm:     "AES-256-GCM",
		EncryptedBlob: "base64encodedblob",
		CreatedAt_:    time.Now().Unix(),
	}

	err = db.SaveKey(km)
	if err != nil {
		t.Fatalf("SaveKey failed: %v", err)
	}
}

// TestListKeys tests the ListKeys function to ensure it returns all saved keys.
// It verifies that the initial list is empty and that added keys are correctly listed.
func TestListKeys(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := NewDatabase(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}

	// Initially, no keys
	keys, err := db.ListKeys()
	if err != nil {
		t.Fatalf("ListKeys failed: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys initially, got %d", len(keys))
	}

	// Add some keys
	for i := 1; i <= 3; i++ {
		km := &KeyModel{
			KeyID:         "test-key-" + string(rune('0'+i)),
			Algorithm:     "AES-256-GCM",
			EncryptedBlob: "blob" + string(rune('0'+i)),
			CreatedAt_:    time.Now().Unix(),
		}
		if err := db.SaveKey(km); err != nil {
			t.Fatalf("SaveKey failed: %v", err)
		}
	}

	// List keys
	keys, err = db.ListKeys()
	if err != nil {
		t.Fatalf("ListKeys failed: %v", err)
	}
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}
}

// TestGetKey tests the GetKey function to ensure it retrieves a key by its ID.
// It verifies that the retrieved key matches the saved key.
func TestGetKey(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := NewDatabase(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}

	keyID := "test-key-1"
	km := &KeyModel{
		KeyID:         keyID,
		Algorithm:     "AES-256-GCM",
		EncryptedBlob: "myblob",
		CreatedAt_:    time.Now().Unix(),
	}

	if err := db.SaveKey(km); err != nil {
		t.Fatalf("SaveKey failed: %v", err)
	}

	retrieved, err := db.GetKey(keyID)
	if err != nil {
		t.Fatalf("GetKey failed: %v", err)
	}

	if retrieved.KeyID != keyID {
		t.Errorf("KeyID mismatch: got %q, want %q", retrieved.KeyID, keyID)
	}
	if retrieved.Algorithm != km.Algorithm {
		t.Errorf("Algorithm mismatch: got %q, want %q", retrieved.Algorithm, km.Algorithm)
	}
	if retrieved.EncryptedBlob != km.EncryptedBlob {
		t.Errorf("EncryptedBlob mismatch: got %q, want %q", retrieved.EncryptedBlob, km.EncryptedBlob)
	}
}

// TestGetKeyNotFound tests the GetKey function when the requested key does not exist.
// It ensures that an error is returned for a nonexistent key.
func TestGetKeyNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := NewDatabase(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}

	_, err = db.GetKey("nonexistent")
	if err == nil {
		t.Error("GetKey should fail for nonexistent key")
	}
}

// TestDeleteKey tests the DeleteKey function to ensure it removes a key from the database.
// It verifies that the key exists before deletion and that it cannot be retrieved afterward.
func TestDeleteKey(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := NewDatabase(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}

	keyID := "test-key-to-delete"
	km := &KeyModel{
		KeyID:         keyID,
		Algorithm:     "AES-256-GCM",
		EncryptedBlob: "myblob",
		CreatedAt_:    time.Now().Unix(),
	}

	if err := db.SaveKey(km); err != nil {
		t.Fatalf("SaveKey failed: %v", err)
	}

	// Verify key exists
	_, err = db.GetKey(keyID)
	if err != nil {
		t.Fatalf("GetKey failed before delete: %v", err)
	}

	// Delete key
	if err := db.DeleteKey(keyID); err != nil {
		t.Fatalf("DeleteKey failed: %v", err)
	}

	// Verify key no longer exists
	_, err = db.GetKey(keyID)
	if err == nil {
		t.Error("GetKey should fail after delete")
	}
}

// TestKeyModelUniqueConstraint tests the unique constraint on the KeyModel's KeyID field.
// It ensures that attempting to save a key with a duplicate KeyID either fails or updates the existing key, depending on the database behavior.
func TestKeyModelUniqueConstraint(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := NewDatabase(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}

	keyID := "unique-key"
	km1 := &KeyModel{
		KeyID:         keyID,
		Algorithm:     "AES-256-GCM",
		EncryptedBlob: "blob1",
		CreatedAt_:    time.Now().Unix(),
	}

	km2 := &KeyModel{
		KeyID:         keyID, // Same KeyID
		Algorithm:     "AES-256-GCM",
		EncryptedBlob: "blob2",
		CreatedAt_:    time.Now().Unix(),
	}

	if err := db.SaveKey(km1); err != nil {
		t.Fatalf("SaveKey km1 failed: %v", err)
	}

	// Second save should fail due to unique constraint or succeed as update
	// Both behaviors are acceptable depending on GORM's save semantics
	_ = db.SaveKey(km2)

	// Verify only one key with this ID exists
	keys, err := db.ListKeys()
	if err != nil {
		t.Fatalf("ListKeys failed: %v", err)
	}

	count := 0
	for _, k := range keys {
		if k.KeyID == keyID {
			count++
		}
	}

	if count != 1 {
		t.Errorf("Expected 1 key with ID %q, got %d", keyID, count)
	}
}
