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
	"io"
	"os"

	"github.com/jabbott-iii/Cryptare/internal"
	"github.com/spf13/cobra"
)

// version is reported by --version. Release builds set it with
// -ldflags "-X main.version=<tag>" (see .github/workflows/cd.yml).
var version = "dev"

func main() {
	rootCmd := newRootCmd(databaseOpener(os.Stderr))
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

// databaseOpener opens the key database at CRYPTARE_DB_PATH or, when that isn't set,
// at the default path in the user's data folder (databasePath), creating that folder
// first and writing noticeLegacyDatabase's notice, if any, to stderr. The CLI calls it
// only for the keys commands, the TUI, encrypt --key and the decryption of a file
// encrypted with a stored key, so other commands never create anything. A database
// refused as untrusted (SEC-016) gets a hint about CRYPTARE_DB_PATH.
func databaseOpener(stderr io.Writer) internal.DatabaseOpener {
	return func() (*internal.Database, error) {
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
		db, err := internal.NewDatabase(path)
		if errors.Is(err, internal.ErrUntrustedDatabase) {
			return nil, fmt.Errorf("%w; set %s to use a key database of your own", err, databasePathEnv)
		}
		return db, err
	}
}

// newRootCmd returns the CLI root command with the build version attached (a
// non-empty Version makes Cobra register the --version flag) and "keys path", which
// needs databasePath from this package.
func newRootCmd(open internal.DatabaseOpener) *cobra.Command {
	cmd := internal.NewRootCmdLazy(open)
	cmd.Version = version
	if keys, _, err := cmd.Find([]string{"keys"}); err == nil && keys != cmd {
		keys.AddCommand(newKeysPathCmd())
	}
	return cmd
}
