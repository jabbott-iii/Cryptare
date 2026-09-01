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

// TestDashboardNavigation verifies that the cursor moves between the main
// menu items and that "enter" on "Manage keys" switches to the keys screen.
func TestDashboardNavigation(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := NewDatabase(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}

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

// TestDashboardEncryptDecryptRoundTrip drives the encrypt then decrypt forms
// exactly as a user typing into the TUI would, and verifies the resulting
// files match the CLI's behaviour.
func TestDashboardEncryptDecryptRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("secret tui data")
	if err := os.WriteFile(srcFile, content, 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	db, err := NewDatabase(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}

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

// TestDashboardKeysGenerateAndExport drives the "Generate a new key" and
// "Export a key" forms and verifies the key is stored and exported.
func TestDashboardKeysGenerateAndExport(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := NewDatabase(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}

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
	tmpDir := t.TempDir()
	db, err := NewDatabase(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}

	m := NewDashboardModel(db)
	m.startForm(actionEncrypt, screenMain)
	m = typeString(m, "abc")

	next, _ := m.updateForm(tea.KeyMsg{Type: tea.KeyBackspace})
	m = next.(DashboardModel)

	if got := m.fieldValue(labelFilePath); got != "ab" {
		t.Fatalf("fieldValue = %q, want %q", got, "ab")
	}
}
