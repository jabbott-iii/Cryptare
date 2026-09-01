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
	tea "github.com/charmbracelet/bubbletea"
)

//--------------------------------------------------messages-------------------------------------------------------------------------------------//

type keysLoadedMsg struct{ keys []KeyModel }
type errMsg struct{ err error }

// actionResultMsg carries the outcome of running a background file/key
// operation triggered from a form screen. See logic-tui.go.
type actionResultMsg struct {
	message string
	err     error
	reload  bool // true if the stored key list should be reloaded
}

//--------------------------------------------------model----------------------------------------------------------------------------------------//

type dashboardScreen int

const (
	screenMain dashboardScreen = iota
	screenKeys
	screenForm
)

// actionKind identifies which file/key operation a form screen submits.
type actionKind int

const (
	actionNone actionKind = iota
	actionEncrypt
	actionDecrypt
	actionCompress
	actionDecompress
	actionKeysGenerate
	actionKeysExport
	actionKeysImport
)

// formField is a single editable text field rendered on the form screen.
type formField struct {
	label    string
	value    []rune
	password bool
}

// DashboardModel is the root TUI model.
type DashboardModel struct {
	db      *Database
	screen  dashboardScreen
	cursor  int
	keys    []KeyModel
	status  string
	isError bool
	busy    bool
	width   int
	height  int

	// form state, used when screen == screenForm
	action     actionKind
	fields     []formField
	fieldIdx   int
	formOrigin dashboardScreen
}

var mainMenuItems = []string{
	"Encrypt a file",
	"Decrypt a file",
	"Compress a file",
	"Decompress a file",
	"Manage keys",
}

var keysMenuItems = []string{
	"Generate a new key",
	"Export a key",
	"Import a key",
}

// NewDashboardModel creates the initial dashboard model.
func NewDashboardModel(db *Database) DashboardModel {
	return DashboardModel{db: db}
}

//--------------------------------------------------bubbletea interface--------------------------------------------------------------------------//

func (m DashboardModel) Init() tea.Cmd {
	return loadKeysCmd(m.db)
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
