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

package pkg

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// Run runs cryptare with the arguments in os.Args, reporting version for --version, and
// returns the exit status: 0 on success, otherwise exitCode's status for the error.
// main calls it, so the tests can run cryptare in a subprocess the same way (TestRunMain).
func Run(version string) int {
	rootCmd := newRootCmd(databaseOpener(os.Stderr), version)
	if err := rootCmd.Execute(); err != nil {
		return exitCode(err)
	}
	return 0
}

// exitCode is the exit status for a command that failed with err: 128 plus the
// signal's number for one stopped by a signal (130 for Ctrl+C), as shells report, and
// 1 otherwise.
func exitCode(err error) int {
	var interrupted *InterruptedError
	if errors.As(err, &interrupted) {
		return interrupted.ExitCode()
	}
	return 1
}

// databaseOpener opens the key database at CRYPTARE_DB_PATH or, when that isn't set,
// at the default path in the user's data folder (databasePath), creating that folder
// first and writing noticeLegacyDatabase's notice, if any, to stderr. The CLI calls it
// only for the keys commands, the TUI, encrypt --key and the decryption of a file
// encrypted with a stored key, so other commands never create anything. A database
// refused as untrusted (SEC-016) gets a hint about CRYPTARE_DB_PATH.
func databaseOpener(stderr io.Writer) DatabaseOpener {
	return func() (*Database, error) {
		path, fromEnv, err := databasePath()
		if err != nil {
			return nil, err
		}
		if !fromEnv {
			noticeLegacyDatabase(stderr, path)
			if err := prepareDefaultDatabaseFolder(path); err != nil {
				return nil, err
			}
		}
		db, err := NewDatabase(path)
		if errors.Is(err, ErrUntrustedDatabase) {
			return nil, fmt.Errorf("%w; set %s to use a key database of your own", err, databasePathEnv)
		}
		return db, err
	}
}

// newRootCmd returns the CLI root command with the build version attached (a
// non-empty Version makes Cobra register the --version flag) and "keys path"
// (newKeysPathCmd).
func newRootCmd(open DatabaseOpener, version string) *cobra.Command {
	cmd := NewRootCmdLazy(open)
	cmd.Version = version
	if keys, _, err := cmd.Find([]string{"keys"}); err == nil && keys != cmd {
		keys.AddCommand(newKeysPathCmd())
	}
	return cmd
}
