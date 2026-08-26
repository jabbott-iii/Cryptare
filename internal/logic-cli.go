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
	"github.com/spf13/cobra"
)

//-----------------------------------------core---------------------------------------------------------//

// NewRootCmd is the rete application entry point
func NewRootCmd(db *Database) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cryptare",
		Short: "A file encryption and management tool",
		RunE: func(cmd *cobra.Command, args []string) error {
			p := tea.NewProgram(NewDashboardModel(db), tea.WithAltScreen())
			_, err := p.Run()
			return err
		},
	}

	return cmd
}
