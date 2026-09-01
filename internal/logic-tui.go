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
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

//--------------------------------------------------styles---------------------------------------------------------------------------------------//

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f88e02")).
			Bold(true).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9333EA")).
			Bold(true).
			PaddingBottom(1)

	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f88e02")).
			PaddingBottom(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f10c0c")).
			MarginTop(1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f88e02"))
)

//--------------------------------------------------field labels---------------------------------------------------------------------------------//

const (
	labelFilePath = "File path"
	labelOutput   = "Output path (optional)"
	labelPassword = "Password"
	labelLevel    = "Compression level 1-9 (optional)"
	labelKeyID    = "Key ID"
)

//--------------------------------------------------form field sets------------------------------------------------------------------------------//

func fieldsFor(action actionKind) []formField {
	switch action {
	case actionEncrypt, actionDecrypt:
		return []formField{
			{label: labelFilePath},
			{label: labelOutput},
			{label: labelPassword, password: true},
		}
	case actionCompress:
		return []formField{
			{label: labelFilePath},
			{label: labelOutput},
			{label: labelLevel},
		}
	case actionDecompress:
		return []formField{
			{label: labelFilePath},
			{label: labelOutput},
		}
	case actionKeysGenerate:
		return []formField{
			{label: labelPassword, password: true},
		}
	case actionKeysExport:
		return []formField{
			{label: labelKeyID},
			{label: labelOutput},
			{label: labelPassword, password: true},
		}
	case actionKeysImport:
		return []formField{
			{label: labelFilePath},
			{label: labelPassword, password: true},
		}
	default:
		return nil
	}
}

func actionTitle(action actionKind) string {
	switch action {
	case actionEncrypt:
		return "Encrypt a file"
	case actionDecrypt:
		return "Decrypt a file"
	case actionCompress:
		return "Compress a file"
	case actionDecompress:
		return "Decompress a file"
	case actionKeysGenerate:
		return "Generate a new key"
	case actionKeysExport:
		return "Export a key"
	case actionKeysImport:
		return "Import a key"
	default:
		return ""
	}
}

// fieldValue returns the current text of the field with the given label, or
// "" if no such field exists on the form.
func (m DashboardModel) fieldValue(label string) string {
	for _, f := range m.fields {
		if f.label == label {
			return string(f.value)
		}
	}
	return ""
}

//--------------------------------------------------bubbletea update-----------------------------------------------------------------------------//

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case keysLoadedMsg:
		m.keys = msg.keys
		return m, nil

	case errMsg:
		m.status = msg.err.Error()
		m.isError = true
		return m, nil

	case actionResultMsg:
		m.busy = false
		if msg.err != nil {
			m.status = msg.err.Error()
			m.isError = true
			return m, nil
		}
		m.status = msg.message
		m.isError = false
		if msg.reload {
			return m, loadKeysCmd(m.db)
		}
		return m, nil

	case tea.KeyMsg:
		if m.screen == screenForm {
			return m.updateForm(msg)
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "up", "shift+tab":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "tab":
			if m.screen == screenMain && m.cursor < len(mainMenuItems)-1 {
				m.cursor++
			} else if m.screen == screenKeys && m.cursor < len(keysMenuItems)-1 {
				m.cursor++
			}

		case "enter":
			switch m.screen {
			case screenMain:
				switch m.cursor {
				case 0:
					m.startForm(actionEncrypt, screenMain)
				case 1:
					m.startForm(actionDecrypt, screenMain)
				case 2:
					m.startForm(actionCompress, screenMain)
				case 3:
					m.startForm(actionDecompress, screenMain)
				case 4: // Manage keys
					m.screen = screenKeys
					m.cursor = 0
				}
			case screenKeys:
				switch m.cursor {
				case 0:
					m.startForm(actionKeysGenerate, screenKeys)
				case 1:
					m.startForm(actionKeysExport, screenKeys)
				case 2:
					m.startForm(actionKeysImport, screenKeys)
				}
			}

		case "esc", "b":
			if m.screen != screenMain {
				m.screen = screenMain
				m.cursor = 0
			}
		}
	}

	return m, nil
}

// startForm switches the model into the form screen for the given action.
func (m *DashboardModel) startForm(action actionKind, origin dashboardScreen) {
	m.action = action
	m.fields = fieldsFor(action)
	m.fieldIdx = 0
	m.formOrigin = origin
	m.screen = screenForm
	m.status = ""
}

// updateForm handles key input while the form screen is active.
func (m DashboardModel) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit

	case tea.KeyEsc:
		m.screen = m.formOrigin
		m.status = ""
		return m, nil

	case tea.KeyTab, tea.KeyDown:
		if m.fieldIdx < len(m.fields)-1 {
			m.fieldIdx++
		}
		return m, nil

	case tea.KeyShiftTab, tea.KeyUp:
		if m.fieldIdx > 0 {
			m.fieldIdx--
		}
		return m, nil

	case tea.KeyBackspace:
		if len(m.fields) > 0 && len(m.fields[m.fieldIdx].value) > 0 {
			f := &m.fields[m.fieldIdx]
			f.value = f.value[:len(f.value)-1]
		}
		return m, nil

	case tea.KeyEnter:
		if m.fieldIdx < len(m.fields)-1 {
			m.fieldIdx++
			return m, nil
		}
		cmd := m.buildActionCmd()
		m.busy = true
		m.screen = m.formOrigin
		m.status = "Working…"
		m.isError = false
		return m, cmd

	case tea.KeySpace:
		if len(m.fields) > 0 {
			f := &m.fields[m.fieldIdx]
			f.value = append(f.value, ' ')
		}
		return m, nil

	case tea.KeyRunes:
		if len(m.fields) > 0 {
			f := &m.fields[m.fieldIdx]
			f.value = append(f.value, msg.Runes...)
		}
		return m, nil
	}

	return m, nil
}

//--------------------------------------------------bubbletea view-------------------------------------------------------------------------------//

func (m DashboardModel) View() string {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render("🔒 Cryptare"))
	sb.WriteString("\n\n")

	switch m.screen {
	case screenMain:
		for i, item := range mainMenuItems {
			if i == m.cursor {
				sb.WriteString(selectedStyle.Render("▶ " + item))
			} else {
				sb.WriteString(itemStyle.Render("  " + item))
			}
			sb.WriteString("\n")
		}

	case screenKeys:
		sb.WriteString(statusStyle.Render("  Manage Keys  (press Esc to go back)\n\n"))
		for i, item := range keysMenuItems {
			if i == m.cursor {
				sb.WriteString(selectedStyle.Render("▶ " + item))
			} else {
				sb.WriteString(itemStyle.Render("  " + item))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
		if len(m.keys) == 0 {
			sb.WriteString(statusStyle.Render("No keys stored. Use \"Generate a new key\" to create one."))
			sb.WriteString("\n")
		} else {
			sb.WriteString(itemStyle.Render(fmt.Sprintf("%-20s  %-12s  %s", "KEY ID", "ALGORITHM", "CREATED")))
			sb.WriteString("\n")
			for _, k := range m.keys {
				created := time.Unix(k.CreatedAt_, 0).Format("2006-01-02 15:04")
				line := fmt.Sprintf("%-20s  %-12s  %s", k.KeyID, k.Algorithm, created)
				sb.WriteString(itemStyle.Render("  " + line))
				sb.WriteString("\n")
			}
		}

	case screenForm:
		sb.WriteString(statusStyle.Render(actionTitle(m.action) + "  (Tab/Enter: next field • Shift+Tab: prev • Esc: cancel)\n\n"))
		for i, f := range m.fields {
			display := string(f.value)
			if f.password {
				display = strings.Repeat("*", len(f.value))
			}
			cursor := "  "
			style := itemStyle
			if i == m.fieldIdx {
				cursor = "▶ "
				style = selectedStyle
				display += "█"
			}
			sb.WriteString(style.Render(fmt.Sprintf("%s%s: %s", cursor, f.label, display)))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
	if m.status != "" {
		if m.isError {
			sb.WriteString(errorStyle.Render("✗ " + m.status))
			sb.WriteString("\n")
		} else {
			sb.WriteString(statusStyle.Render("• " + m.status))
			sb.WriteString("\n")
		}
	}

	sb.WriteString(statusStyle.Render("↑/shift+tab | ↓/tab: navigate • Enter: select • q: quit"))
	sb.WriteString("\n")
	return sb.String()
}

//--------------------------------------------------action commands------------------------------------------------------------------------------//

// buildActionCmd returns a tea.Cmd that performs the currently selected
// action using the values entered into the form fields, mirroring the
// behaviour of the equivalent commands in logic-cli.go.
func (m DashboardModel) buildActionCmd() tea.Cmd {
	db := m.db
	action := m.action

	file := m.fieldValue(labelFilePath)
	output := m.fieldValue(labelOutput)
	password := m.fieldValue(labelPassword)
	keyID := m.fieldValue(labelKeyID)
	levelStr := m.fieldValue(labelLevel)

	switch action {
	case actionEncrypt:
		return func() tea.Msg {
			if err := EncryptFile(file, output, password); err != nil {
				return actionResultMsg{err: err}
			}
			dst := output
			if dst == "" {
				dst = file + encExt
			}
			return actionResultMsg{message: fmt.Sprintf("Encrypted: %s → %s", file, dst)}
		}

	case actionDecrypt:
		return func() tea.Msg {
			if err := DecryptFile(file, output, password); err != nil {
				return actionResultMsg{err: err}
			}
			dst := output
			if dst == "" {
				dst = deriveDecryptOutput(file)
			}
			return actionResultMsg{message: fmt.Sprintf("Decrypted: %s → %s", file, dst)}
		}

	case actionCompress:
		return func() tea.Msg {
			level := -1
			if levelStr != "" {
				lv, err := strconv.Atoi(levelStr)
				if err != nil {
					return actionResultMsg{err: fmt.Errorf("invalid compression level %q: %w", levelStr, err)}
				}
				level = lv
			}
			if err := CompressFile(file, output, level); err != nil {
				return actionResultMsg{err: err}
			}
			dst := output
			if dst == "" {
				dst = file + gzExt
			}
			return actionResultMsg{message: fmt.Sprintf("Compressed: %s → %s", file, dst)}
		}

	case actionDecompress:
		return func() tea.Msg {
			if err := DecompressFile(file, output); err != nil {
				return actionResultMsg{err: err}
			}
			dst := output
			if dst == "" {
				dst = deriveDecompressOutput(file)
			}
			return actionResultMsg{message: fmt.Sprintf("Decompressed: %s → %s", file, dst)}
		}

	case actionKeysGenerate:
		return func() tea.Msg {
			rawKey, err := GenerateKey()
			if err != nil {
				return actionResultMsg{err: err}
			}

			keyID, err := newKeyID()
			if err != nil {
				return actionResultMsg{err: err}
			}

			blob, err := EncryptKeyBlob(rawKey, password)
			if err != nil {
				return actionResultMsg{err: err}
			}

			km := &KeyModel{
				KeyID:         keyID,
				Algorithm:     "AES-256-GCM",
				EncryptedBlob: blob,
				CreatedAt_:    time.Now().Unix(),
			}

			if err := db.SaveKey(km); err != nil {
				return actionResultMsg{err: fmt.Errorf("save key: %w", err)}
			}

			return actionResultMsg{message: fmt.Sprintf("Generated key: %s", keyID), reload: true}
		}

	case actionKeysExport:
		return func() tea.Msg {
			km, err := db.GetKey(keyID)
			if err != nil {
				return actionResultMsg{err: fmt.Errorf("key not found: %w", err)}
			}

			if err := ExportKeyToFile(km, password, output); err != nil {
				return actionResultMsg{err: err}
			}

			dst := output
			if dst == "" {
				dst = fmt.Sprintf("%s-%d.ckey", keyID, time.Now().Unix())
			}
			return actionResultMsg{message: fmt.Sprintf("Exported key %s → %s", keyID, dst)}
		}

	case actionKeysImport:
		return func() tea.Msg {
			km, err := ImportKeyFromFile(file, password)
			if err != nil {
				return actionResultMsg{err: err}
			}

			if err := db.SaveKey(km); err != nil {
				return actionResultMsg{err: fmt.Errorf("save imported key: %w", err)}
			}

			return actionResultMsg{message: fmt.Sprintf("Imported key: %s", km.KeyID), reload: true}
		}
	}

	return nil
}
