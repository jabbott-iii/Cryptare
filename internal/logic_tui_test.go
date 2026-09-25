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
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// typeString simulates typing each rune of s into the currently focused field.
func typeString(m DashboardModel, s string) DashboardModel {
	for _, r := range s {
		next, _ := m.updateForm(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = next.(DashboardModel)
	}
	return m
}

func newVimDashboardModel(db *Database) DashboardModel {
	return NewDashboardModelWithOptions(db, dashboardOptions{vimEnabled: true})
}

// TestDashboardNavigation verifies that the cursor moves between the main
// menu items and that "enter" on "Manage keys" switches to the key screen.
func TestDashboardNavigation(t *testing.T) {
	db := newTestDatabase(t, false)

	m := NewDashboardModel(db)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(DashboardModel)
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.cursor)
	}

	for m.cursor < 4 {
		next, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = next.(DashboardModel)
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(DashboardModel)
	if m.screen != screenKeys {
		t.Fatalf("screen = %v, want screenKeys", m.screen)
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(DashboardModel)
	if m.screen != screenMain {
		t.Fatalf("screen = %v, want screenMain after Esc", m.screen)
	}
}

func TestDashboardVimMenuNavigation(t *testing.T) {
	db := newTestDatabase(t, false)

	m := newVimDashboardModel(db)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = next.(DashboardModel)
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1 after j", m.cursor)
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = next.(DashboardModel)
	if m.cursor != 0 {
		t.Fatalf("cursor = %d, want 0 after k", m.cursor)
	}

	for m.cursor < 4 {
		next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m = next.(DashboardModel)
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	m = next.(DashboardModel)
	if m.screen != screenKeys {
		t.Fatalf("screen = %v, want screenKeys after l", m.screen)
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	m = next.(DashboardModel)
	if m.screen != screenMain {
		t.Fatalf("screen = %v, want screenMain after h", m.screen)
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = next.(DashboardModel)
	for m.cursor < 4 {
		next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m = next.(DashboardModel)
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	m = next.(DashboardModel)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	m = next.(DashboardModel)
	if m.screen != screenMain {
		t.Fatalf("screen = %v, want screenMain after b", m.screen)
	}
}

func TestDashboardVimFormModeTransitions(t *testing.T) {
	db := newTestDatabase(t, false)

	m := newVimDashboardModel(db)
	m.startForm(actionEncrypt, screenMain)

	if m.formMode != formModeInsert {
		t.Fatalf("formMode = %v, want insert", m.formMode)
	}

	m = typeString(m, "ab")
	if got := m.fieldValue(labelFilePath); got != "ab" {
		t.Fatalf("fieldValue = %q, want %q", got, "ab")
	}

	next, _ := m.updateForm(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(DashboardModel)
	if m.screen != screenForm {
		t.Fatalf("screen = %v, want screenForm after leaving insert mode", m.screen)
	}
	if m.formMode != formModeNormal {
		t.Fatalf("formMode = %v, want normal after Esc", m.formMode)
	}

	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = next.(DashboardModel)
	if m.fieldIdx != 1 {
		t.Fatalf("fieldIdx = %d, want 1 after j", m.fieldIdx)
	}
	if got := m.fieldValue(labelFilePath); got != "ab" {
		t.Fatalf("fieldValue changed in normal mode: got %q, want %q", got, "ab")
	}

	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	m = next.(DashboardModel)
	if m.fieldIdx != 2 {
		t.Fatalf("fieldIdx = %d, want 2 after o", m.fieldIdx)
	}
	if m.formMode != formModeInsert {
		t.Fatalf("formMode = %v, want insert after o", m.formMode)
	}

	m = typeString(m, "pw")
	if got := m.fieldValue(labelPassword); got != "pw" {
		t.Fatalf("password field = %q, want %q", got, "pw")
	}
}

func TestDashboardVimNormalModeDoesNotEditFields(t *testing.T) {
	db := newTestDatabase(t, false)

	m := newVimDashboardModel(db)
	m.startForm(actionEncrypt, screenMain)
	m = typeString(m, "path")

	next, _ := m.updateForm(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(DashboardModel)

	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	m = next.(DashboardModel)
	if got := m.fieldValue(labelFilePath); got != "path" {
		t.Fatalf("fieldValue = %q, want %q after x in normal mode", got, "path")
	}

	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeySpace})
	m = next.(DashboardModel)
	if got := m.fieldValue(labelFilePath); got != "path" {
		t.Fatalf("fieldValue = %q, want %q after space in normal mode", got, "path")
	}

	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeyBackspace})
	m = next.(DashboardModel)
	if got := m.fieldValue(labelFilePath); got != "path" {
		t.Fatalf("fieldValue = %q, want %q after backspace in normal mode", got, "path")
	}

	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	m = next.(DashboardModel)
	m = typeString(m, "!")
	if got := m.fieldValue(labelFilePath); got != "path!" {
		t.Fatalf("fieldValue = %q, want %q after returning to insert mode", got, "path!")
	}
}

func TestDashboardVimNormalModeLSubmitsLastField(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "submit.txt")
	if err := os.WriteFile(srcFile, []byte("vim submit"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	db := newTestDatabase(t, false)

	m := newVimDashboardModel(db)
	m.startForm(actionEncrypt, screenMain)
	m = typeString(m, srcFile)

	next, _ := m.updateForm(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(DashboardModel)
	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(DashboardModel)
	m = typeString(m, "hunter2")

	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(DashboardModel)
	if m.formMode != formModeNormal {
		t.Fatalf("formMode = %v, want normal", m.formMode)
	}

	next, cmd := m.updateForm(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	m = next.(DashboardModel)
	if cmd == nil {
		t.Fatal("expected a command to be returned for vim l submission")
	}
	if m.screen != screenMain {
		t.Fatalf("screen = %v, want screenMain after submission", m.screen)
	}
	if !m.busy {
		t.Fatal("expected busy=true after submission")
	}

	msg := cmd()
	result, ok := msg.(actionResultMsg)
	if !ok {
		t.Fatalf("expected actionResultMsg, got %T", msg)
	}
	if result.err != nil {
		t.Fatalf("encrypt action failed: %v", result.err)
	}

	if _, err := os.Stat(srcFile + encExt); err != nil {
		t.Fatalf("encrypted file not created: %v", err)
	}
}

func TestDashboardStandardFormBindingsUnaffectedWhenVimDisabled(t *testing.T) {
	db := newTestDatabase(t, false)

	m := NewDashboardModel(db)
	m.startForm(actionEncrypt, screenMain)

	next, _ := m.updateForm(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = next.(DashboardModel)
	if got := m.fieldValue(labelFilePath); got != "j" {
		t.Fatalf("fieldValue = %q, want %q with vim disabled", got, "j")
	}

	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(DashboardModel)
	if m.screen != screenMain {
		t.Fatalf("screen = %v, want screenMain after Esc with vim disabled", m.screen)
	}
}

// TestDashboardEncryptDecryptRoundTrip drives the encrypting, then decrypt forms
// exactly as a user typing into the TUI would, and verifies the resulting
// files match the CLI's behavior.
func TestDashboardEncryptDecryptRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("secret tui data")
	if err := os.WriteFile(srcFile, content, 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	db := newTestDatabase(t, false)

	m := NewDashboardModel(db)
	m.startForm(actionEncrypt, screenMain)

	m = typeString(m, srcFile)
	next, _ := m.updateForm(tea.KeyMsg{Type: tea.KeyEnter}) // move to output
	m = next.(DashboardModel)
	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeyEnter}) // leave output empty, move to password
	m = next.(DashboardModel)
	m = typeString(m, "hunter2")

	next, cmd := m.updateForm(tea.KeyMsg{Type: tea.KeyEnter}) // submit
	m = next.(DashboardModel)
	if cmd == nil {
		t.Fatal("expected a command to be returned for encrypt submission")
	}

	msg := cmd()
	result, ok := msg.(actionResultMsg)
	if !ok {
		t.Fatalf("expected actionResultMsg, got %T", msg)
	}
	if result.err != nil {
		t.Fatalf("encrypt action failed: %v", result.err)
	}

	encFile := srcFile + encExt
	if _, err := os.Stat(encFile); err != nil {
		t.Fatalf("encrypted file not created: %v", err)
	}

	next, _ = m.Update(result)
	m = next.(DashboardModel)
	if m.isError || m.status == "" {
		t.Fatalf("expected success status after encrypt, got isError=%v status=%q", m.isError, m.status)
	}

	// Now decrypt using the form.
	m.startForm(actionDecrypt, screenMain)
	m = typeString(m, encFile)
	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeyEnter}) // move to output
	m = next.(DashboardModel)
	decFile := filepath.Join(tmpDir, "decrypted.txt")
	m = typeString(m, decFile)
	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeyEnter}) // move to password
	m = next.(DashboardModel)
	m = typeString(m, "hunter2")

	next, cmd = m.updateForm(tea.KeyMsg{Type: tea.KeyEnter}) // submit
	m = next.(DashboardModel)
	if cmd == nil {
		t.Fatal("expected a command to be returned for decrypt submission")
	}

	msg = cmd()
	result, ok = msg.(actionResultMsg)
	if !ok {
		t.Fatalf("expected actionResultMsg, got %T", msg)
	}
	if result.err != nil {
		t.Fatalf("decrypt action failed: %v", result.err)
	}

	decData, err := os.ReadFile(decFile)
	if err != nil {
		t.Fatalf("read decrypted file: %v", err)
	}
	if string(decData) != string(content) {
		t.Fatalf("content mismatch: got %q, want %q", decData, content)
	}
}

// TestDashboardEncryptDecryptDirectoryRoundTrip verifies the TUI accepts a
// directory path for encrypt/decrypt actions and restores the directory tree.
func TestDashboardEncryptDecryptDirectoryRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "bundle")
	if err := os.MkdirAll(filepath.Join(srcDir, "nested"), 0o755); err != nil {
		t.Fatalf("create source directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(srcDir, "empty"), 0o755); err != nil {
		t.Fatalf("create empty source directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "nested", "file.txt"), []byte("secret tui directory data"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	db := newTestDatabase(t, false)

	m := NewDashboardModel(db)
	m.startForm(actionEncrypt, screenMain)
	m = typeString(m, srcDir)
	next, _ := m.updateForm(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(DashboardModel)
	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(DashboardModel)
	m = typeString(m, "hunter2")

	next, cmd := m.updateForm(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(DashboardModel)
	if cmd == nil {
		t.Fatal("expected a command to be returned for encrypt submission")
	}

	msg := cmd()
	result, ok := msg.(actionResultMsg)
	if !ok {
		t.Fatalf("expected actionResultMsg, got %T", msg)
	}
	if result.err != nil {
		t.Fatalf("encrypt action failed: %v", result.err)
	}

	encFile := srcDir + encExt
	if _, err := os.Stat(encFile); err != nil {
		t.Fatalf("encrypted artifact not created: %v", err)
	}

	m.startForm(actionDecrypt, screenMain)
	m = typeString(m, encFile)
	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(DashboardModel)
	restoreDir := filepath.Join(tmpDir, "restored")
	m = typeString(m, restoreDir)
	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(DashboardModel)
	m = typeString(m, "hunter2")

	_, cmd = m.updateForm(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a command to be returned for decrypt submission")
	}

	msg = cmd()
	result, ok = msg.(actionResultMsg)
	if !ok {
		t.Fatalf("expected actionResultMsg, got %T", msg)
	}
	if result.err != nil {
		t.Fatalf("decrypt action failed: %v", result.err)
	}

	data, err := os.ReadFile(filepath.Join(restoreDir, "nested", "file.txt"))
	if err != nil {
		t.Fatalf("read restored file: %v", err)
	}
	if string(data) != "secret tui directory data" {
		t.Fatalf("content mismatch: got %q", string(data))
	}
	if info, err := os.Stat(filepath.Join(restoreDir, "empty")); err != nil {
		t.Fatalf("expected empty directory not restored: %v", err)
	} else if !info.IsDir() {
		t.Fatal("restored empty path is not a directory")
	}
}

func TestDashboardCompressDecompressZipRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "zip-bundle")
	if err := os.MkdirAll(filepath.Join(srcDir, "nested"), 0o755); err != nil {
		t.Fatalf("create source directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(srcDir, "empty"), 0o755); err != nil {
		t.Fatalf("create empty source directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "nested", "file.txt"), []byte("zip tui data"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	db := newTestDatabase(t, false)

	m := NewDashboardModel(db)
	m.startForm(actionCompress, screenMain)
	m = typeString(m, srcDir)
	next, _ := m.updateForm(tea.KeyMsg{Type: tea.KeyEnter}) // output
	m = next.(DashboardModel)
	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeyEnter}) // format
	m = next.(DashboardModel)
	m = typeString(m, "zip")
	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeyEnter}) // level
	m = next.(DashboardModel)
	next, cmd := m.updateForm(tea.KeyMsg{Type: tea.KeyEnter}) // submit
	m = next.(DashboardModel)
	if cmd == nil {
		t.Fatal("expected a command to be returned for compress submission")
	}

	msg := cmd()
	result, ok := msg.(actionResultMsg)
	if !ok {
		t.Fatalf("expected actionResultMsg, got %T", msg)
	}
	if result.err != nil {
		t.Fatalf("compress action failed: %v", result.err)
	}

	archive := srcDir + zipExt
	if _, err := os.Stat(archive); err != nil {
		t.Fatalf("zip archive not created: %v", err)
	}
	if err := os.RemoveAll(srcDir); err != nil {
		t.Fatalf("remove source directory: %v", err)
	}

	m.startForm(actionDecompress, screenMain)
	m = typeString(m, archive)
	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeyEnter}) // output
	m = next.(DashboardModel)
	_, cmd = m.updateForm(tea.KeyMsg{Type: tea.KeyEnter}) // submit
	if cmd == nil {
		t.Fatal("expected a command to be returned for decompress submission")
	}

	msg = cmd()
	result, ok = msg.(actionResultMsg)
	if !ok {
		t.Fatalf("expected actionResultMsg, got %T", msg)
	}
	if result.err != nil {
		t.Fatalf("decompress action failed: %v", result.err)
	}

	data, err := os.ReadFile(filepath.Join(srcDir, "nested", "file.txt"))
	if err != nil {
		t.Fatalf("read restored file: %v", err)
	}
	if string(data) != "zip tui data" {
		t.Fatalf("content mismatch: got %q", string(data))
	}
	if info, err := os.Stat(filepath.Join(srcDir, "empty")); err != nil {
		t.Fatalf("expected empty directory not restored: %v", err)
	} else if !info.IsDir() {
		t.Fatal("restored empty path is not a directory")
	}
}

// TestDashboardKeysGenerateAndExport drives the "Generate a new key" and
// "Export a key" forms and verifies the key is stored and exported.
func TestDashboardKeysGenerateAndExport(t *testing.T) {
	tmpDir := t.TempDir()
	db := newTestDatabase(t, false)

	m := NewDashboardModel(db)
	m.startForm(actionKeysGenerate, screenKeys)
	m = typeString(m, "masterpass")

	_, cmd := m.updateForm(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a command for keys generate submission")
	}
	msg := cmd()
	result, ok := msg.(actionResultMsg)
	if !ok {
		t.Fatalf("expected actionResultMsg, got %T", msg)
	}
	if result.err != nil {
		t.Fatalf("generate key failed: %v", result.err)
	}
	if !result.reload {
		t.Fatal("expected reload=true after generating a key")
	}

	keys, err := db.ListKeys()
	if err != nil || len(keys) != 1 {
		t.Fatalf("expected 1 stored key, got %d (err=%v)", len(keys), err)
	}

	// Export the generated key.
	m.startForm(actionKeysExport, screenKeys)
	m = typeString(m, keys[0].KeyID)
	next, _ := m.updateForm(tea.KeyMsg{Type: tea.KeyEnter}) // move to output
	m = next.(DashboardModel)
	outFile := filepath.Join(tmpDir, "exported.ckey")
	m = typeString(m, outFile)
	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeyEnter}) // move to password
	m = next.(DashboardModel)
	m = typeString(m, "masterpass")

	_, cmd = m.updateForm(tea.KeyMsg{Type: tea.KeyEnter}) // submit
	if cmd == nil {
		t.Fatal("expected a command for keys export submission")
	}
	msg = cmd()
	result, ok = msg.(actionResultMsg)
	if !ok {
		t.Fatalf("expected actionResultMsg, got %T", msg)
	}
	if result.err != nil {
		t.Fatalf("export key failed: %v", result.err)
	}
	if _, err := os.Stat(outFile); err != nil {
		t.Fatalf("exported key file not created: %v", err)
	}
}

// TestDashboardFormBackspace verifies that backspace removes the last
// character typed into the focused field.
func TestDashboardFormBackspace(t *testing.T) {
	db := newTestDatabase(t, false)

	m := NewDashboardModel(db)
	m.startForm(actionEncrypt, screenMain)
	m = typeString(m, "abc")

	next, _ := m.updateForm(tea.KeyMsg{Type: tea.KeyBackspace})
	m = next.(DashboardModel)

	if got := m.fieldValue(labelFilePath); got != "ab" {
		t.Fatalf("fieldValue = %q, want %q", got, "ab")
	}
}

func TestDashboardKeysDelete(t *testing.T) {
	db := newTestDatabase(t, false)

	km := &KeyModel{
		KeyID:         "tui-delete",
		Algorithm:     "AES-256-GCM",
		EncryptedBlob: "blob",
		CreatedAt_:    time.Now().Unix(),
	}
	if err := db.SaveKey(km); err != nil {
		t.Fatalf("SaveKey failed: %v", err)
	}

	m := NewDashboardModel(db)
	m.startForm(actionKeysDelete, screenKeys)
	m = typeString(m, km.KeyID)
	next, _ := m.updateForm(tea.KeyMsg{Type: tea.KeyEnter}) // move to confirmation field
	m = next.(DashboardModel)
	m = typeString(m, "DELETE")

	_, cmd := m.updateForm(tea.KeyMsg{Type: tea.KeyEnter}) // submit
	if cmd == nil {
		t.Fatal("expected a command for keys delete submission")
	}
	msg := cmd()
	result, ok := msg.(actionResultMsg)
	if !ok {
		t.Fatalf("expected actionResultMsg, got %T", msg)
	}
	if result.err != nil {
		t.Fatalf("delete key failed: %v", result.err)
	}
	if !result.reload {
		t.Fatal("expected reload=true after deleting a key")
	}

	if _, err := db.GetKey(km.KeyID); err == nil {
		t.Fatal("expected key to be deleted")
	}
}

func TestDashboardKeysDeleteRequiresConfirmationPhrase(t *testing.T) {
	db := newTestDatabase(t, false)

	km := &KeyModel{
		KeyID:         "tui-delete-confirmation",
		Algorithm:     "AES-256-GCM",
		EncryptedBlob: "blob",
		CreatedAt_:    time.Now().Unix(),
	}
	if err := db.SaveKey(km); err != nil {
		t.Fatalf("SaveKey failed: %v", err)
	}

	m := NewDashboardModel(db)
	m.startForm(actionKeysDelete, screenKeys)
	m = typeString(m, km.KeyID)
	next, _ := m.updateForm(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(DashboardModel)
	m = typeString(m, "no")

	_, cmd := m.updateForm(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a command for keys delete submission")
	}
	msg := cmd()
	result, ok := msg.(actionResultMsg)
	if !ok {
		t.Fatalf("expected actionResultMsg, got %T", msg)
	}
	if result.err == nil {
		t.Fatal("expected confirmation error")
	}
	if !strings.Contains(result.err.Error(), "type \"DELETE\"") {
		t.Fatalf("unexpected error: %v", result.err)
	}
	if _, err := db.GetKey(km.KeyID); err != nil {
		t.Fatalf("expected key to remain after failed confirmation, got: %v", err)
	}
}

// TestDashboardFormIgnoresUnhandledKeys is a regression test for BUG-002: keys a
// form doesn't use (arrows, Delete, Home/End, …) must be ignored, not crash the TUI.
func TestDashboardFormIgnoresUnhandledKeys(t *testing.T) {
	keys := []tea.KeyType{
		tea.KeyLeft, tea.KeyRight, tea.KeyDelete, tea.KeyHome, tea.KeyEnd,
		tea.KeyPgUp, tea.KeyPgDown, tea.KeyCtrlU, tea.KeyF1,
	}

	for _, vim := range []bool{false, true} {
		for _, key := range keys {
			name := tea.Key{Type: key}.String()
			if vim {
				name = "vim/" + name
			}
			t.Run(name, func(t *testing.T) {
				db := newTestDatabase(t, false)
				m := NewDashboardModelWithOptions(db, dashboardOptions{vimEnabled: vim})
				m.startForm(actionEncrypt, screenMain)
				m = typeString(m, "abc")

				next, cmd := m.Update(tea.KeyMsg{Type: key})
				m = next.(DashboardModel)

				if cmd != nil {
					t.Fatalf("Update(%s) returned a command, want none", name)
				}
				if m.screen != screenForm {
					t.Fatalf("screen = %v, want screenForm", m.screen)
				}
				if got := m.fieldValue(labelFilePath); got != "abc" {
					t.Fatalf("fieldValue = %q, want %q", got, "abc")
				}
			})
		}
	}
}

// TestDashboardUnknownActionReportsError checks that submitting a form with no
// valid action reports an error instead of panicking.
func TestDashboardUnknownActionReportsError(t *testing.T) {
	db := newTestDatabase(t, false)
	m := NewDashboardModel(db)
	m.action = actionNone

	msg, ok := m.buildActionCmd()().(actionResultMsg)
	if !ok {
		t.Fatalf("buildActionCmd() message type = %T, want actionResultMsg", msg)
	}
	if msg.err == nil {
		t.Fatal("buildActionCmd() error = nil, want an unsupported-action error")
	}
}

// TestDashboardEnterOnUnknownScreenIsIgnored checks that Enter on a screen with no
// menu is a no-op instead of a panic.
func TestDashboardEnterOnUnknownScreenIsIgnored(t *testing.T) {
	db := newTestDatabase(t, false)
	m := NewDashboardModel(db)
	m.screen = dashboardScreen(99)

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("Update(enter) returned a command, want none")
	}
	if got := next.(DashboardModel).screen; got != dashboardScreen(99) {
		t.Fatalf("screen = %v, want unchanged", got)
	}
}

// TestDashboardRejectsEmptyPassword is a regression test for SEC-001: submitting a
// TUI form with the password left blank must fail with an error and change nothing.
func TestDashboardRejectsEmptyPassword(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "secret.txt")
	if err := os.WriteFile(srcFile, []byte("secret"), 0o600); err != nil {
		t.Fatalf("write source file: %v", err)
	}
	db := newTestDatabase(t, false)

	// submit presses Enter through every field and runs the resulting command.
	submit := func(m DashboardModel) (DashboardModel, actionResultMsg) {
		t.Helper()
		for {
			next, cmd := m.updateForm(tea.KeyMsg{Type: tea.KeyEnter})
			m = next.(DashboardModel)
			if cmd != nil {
				result, ok := cmd().(actionResultMsg)
				if !ok {
					t.Fatalf("expected actionResultMsg")
				}
				return m, result
			}
		}
	}

	// Encrypt: file path typed, output and password left blank.
	m := NewDashboardModel(db)
	m.startForm(actionEncrypt, screenMain)
	m = typeString(m, srcFile)
	m, result := submit(m)
	if !errors.Is(result.err, ErrEmptyPassword) {
		t.Fatalf("encrypt error = %v, want ErrEmptyPassword", result.err)
	}
	if _, err := os.Stat(srcFile + encExt); !os.IsNotExist(err) {
		t.Fatalf("encrypted file created despite the empty password (stat err: %v)", err)
	}
	next, _ := m.Update(result)
	if got := next.(DashboardModel); !got.isError || !strings.Contains(got.status, "password must not be empty") {
		t.Fatalf("status = %q (isError %v), want the empty-password error", got.status, got.isError)
	}

	// Generate a key with a blank master password.
	m.startForm(actionKeysGenerate, screenKeys)
	_, result = submit(m)
	if !errors.Is(result.err, ErrEmptyPassword) {
		t.Fatalf("keys generate error = %v, want ErrEmptyPassword", result.err)
	}
	if keys, err := db.ListKeys(); err != nil || len(keys) != 0 {
		t.Fatalf("stored keys = %d (err %v), want 0", len(keys), err)
	}

	// Export an existing key with a blank password.
	rawKey, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	blob, err := EncryptKeyBlob(rawKey, "masterpass")
	if err != nil {
		t.Fatalf("EncryptKeyBlob: %v", err)
	}
	if err := db.SaveKey(&KeyModel{KeyID: "0123456789abcdef", Algorithm: "AES-256-GCM", EncryptedBlob: blob, CreatedAt_: 1}); err != nil {
		t.Fatalf("SaveKey: %v", err)
	}
	exportPath := filepath.Join(tmpDir, "exported.ckey")
	m.startForm(actionKeysExport, screenKeys)
	m = typeString(m, "0123456789abcdef")
	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeyEnter}) // move to output
	m = typeString(next.(DashboardModel), exportPath)
	_, result = submit(m)
	if !errors.Is(result.err, ErrEmptyPassword) {
		t.Fatalf("keys export error = %v, want ErrEmptyPassword", result.err)
	}
	if _, err := os.Stat(exportPath); !os.IsNotExist(err) {
		t.Fatalf("export file created despite the empty password (stat err: %v)", err)
	}
}

// TestDashboardRefusesExistingOutput checks that TUI file actions refuse to replace
// an existing output; the TUI has no overwrite option (BUG-004).
func TestDashboardRefusesExistingOutput(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "secret.txt")
	if err := os.WriteFile(src, []byte("secret"), 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}
	taken := filepath.Join(tmpDir, "taken.enc")
	if err := os.WriteFile(taken, []byte("existing"), 0o600); err != nil {
		t.Fatalf("write existing output: %v", err)
	}

	m := NewDashboardModel(newTestDatabase(t, false))
	m.startForm(actionEncrypt, screenMain)
	m = typeString(m, src)
	next, _ := m.updateForm(tea.KeyMsg{Type: tea.KeyEnter}) // output
	m = typeString(next.(DashboardModel), taken)
	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeyEnter}) // password
	m = typeString(next.(DashboardModel), "pw")
	_, cmd := m.updateForm(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a command for the encrypt submission")
	}

	result, ok := cmd().(actionResultMsg)
	if !ok {
		t.Fatal("expected actionResultMsg")
	}
	if !errors.Is(result.err, ErrOutputExists) {
		t.Fatalf("error = %v, want ErrOutputExists", result.err)
	}
	if got, _ := os.ReadFile(taken); string(got) != "existing" {
		t.Fatalf("existing output was modified to %q", got)
	}
}
