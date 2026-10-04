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

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/jabbott-iii/Cryptare/internal"
	"github.com/spf13/cobra"
)

// version is reported by --version. Release builds set it with
// -ldflags "-X main.version=<tag>" (see .github/workflows/cd.yml).
var version = "dev"

func main() {
	rootCmd := newRootCmd(databaseOpener())
	if err := rootCmd.Execute(); err != nil {
		os.Exit(exitCode(err))
	}
}

// exitCode is the exit status for a command that failed with err: 128 plus the
// signal's number for one stopped by a signal (130 for Ctrl+C), as shells report, and
// 1 otherwise.
func exitCode(err error) int {
	var interrupted *internal.InterruptedError
	if errors.As(err, &interrupted) {
		return interrupted.ExitCode()
	}
	return 1
}

// databaseOpener opens the key database at CRYPTARE_DB_PATH (default ./cryptare.db).
// The CLI calls it only for the keys commands and the TUI, so other commands never
// create the database file. A database refused as untrusted (SEC-016) gets a hint
// about CRYPTARE_DB_PATH.
func databaseOpener() internal.DatabaseOpener {
	path := databasePathFromEnv()
	return func() (*internal.Database, error) {
		db, err := internal.NewDatabase(path)
		if errors.Is(err, internal.ErrUntrustedDatabase) {
			return nil, fmt.Errorf("%w; set %s to use a key database of your own", err, databasePathEnv)
		}
		return db, err
	}
}

// newRootCmd returns the CLI root command with the build version attached;
// a non-empty Version makes Cobra register the --version flag.
func newRootCmd(open internal.DatabaseOpener) *cobra.Command {
	cmd := internal.NewRootCmdLazy(open)
	cmd.Version = version
	return cmd
}
