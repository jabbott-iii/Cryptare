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
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

//--------------------------------------------------styles---------------------------------------------------------------------------------------//

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	itemStyle     = lipgloss.NewStyle().PaddingLeft(2)
	selectedStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("212")).Bold(true)
	statusStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

//--------------------------------------------------messages-------------------------------------------------------------------------------------//

type keysLoadedMsg struct{ keys []KeyModel }
type errMsg struct{ err error }

//--------------------------------------------------model----------------------------------------------------------------------------------------//

type dashboardScreen int

const (
	screenMain dashboardScreen = iota
	screenKeys
)

// DashboardModel is the root TUI model.
type DashboardModel struct {
	db      *Database
	screen  dashboardScreen
	cursor  int
	keys    []KeyModel
	status  string
	isError bool
	width   int
	height  int
}

var mainMenuItems = []string{
	"Encrypt a file",
	"Decrypt a file",
	"Compress a file",
	"Decompress a file",
	"Manage keys",
	"Quit",
}

// NewDashboardModel creates the initial dashboard model.
func NewDashboardModel(db *Database) DashboardModel {
	return DashboardModel{db: db}
}

//--------------------------------------------------bubbletea interface--------------------------------------------------------------------------//

func (m DashboardModel) Init() tea.Cmd {
	return loadKeysCmd(m.db)
}

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

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.screen == screenMain && m.cursor < len(mainMenuItems)-1 {
				m.cursor++
			} else if m.screen == screenKeys && m.cursor < len(m.keys)-1 {
				m.cursor++
			}

		case "enter":
			if m.screen == screenMain {
				switch m.cursor {
				case 4: // Manage keys
					m.screen = screenKeys
					m.cursor = 0
				case 5: // Quit
					return m, tea.Quit
				default:
					m.status = "Use the CLI for file operations (see --help)"
					m.isError = false
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

func (m DashboardModel) View() string {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render("✦ Cryptare"))
	sb.WriteString("\n\n")

	switch m.screen {
	case screenMain:
		sb.WriteString("  Main Menu\n\n")
		for i, item := range mainMenuItems {
			if i == m.cursor {
				sb.WriteString(selectedStyle.Render("▶ " + item))
				sb.WriteString("\n")
			} else {
				sb.WriteString(itemStyle.Render("  " + item))
				sb.WriteString("\n")
			}
		}

	case screenKeys:
		sb.WriteString("  Stored Keys  (press Esc to go back)\n\n")
		if len(m.keys) == 0 {
			sb.WriteString(itemStyle.Render("No keys stored. Use `cryptare keys generate` to create one."))
			sb.WriteString("\n")
		} else {
			sb.WriteString(itemStyle.Render(fmt.Sprintf("%-20s  %-12s  %s", "KEY ID", "ALGORITHM", "CREATED")))
			sb.WriteString("\n")
			for i, k := range m.keys {
				created := time.Unix(k.CreatedAt_, 0).Format("2006-01-02 15:04")
				line := fmt.Sprintf("%-20s  %-12s  %s", k.KeyID, k.Algorithm, created)
				if i == m.cursor {
					sb.WriteString(selectedStyle.Render("▶ " + line))
					sb.WriteString("\n")
				} else {
					sb.WriteString(itemStyle.Render("  " + line))
					sb.WriteString("\n")
				}
			}
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

	sb.WriteString(statusStyle.Render("\n↑/↓ navigate • enter select • q quit"))
	sb.WriteString("\n")
	return sb.String()
}

//--------------------------------------------------commands-------------------------------------------------------------------------------------//

func loadKeysCmd(db *Database) tea.Cmd {
	return func() tea.Msg {
		keys, err := db.ListKeys()
		if err != nil {
			return errMsg{err}
		}
		return keysLoadedMsg{keys}
	}
}
