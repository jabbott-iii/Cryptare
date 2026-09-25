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
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
)

//-----------------------------------------core---------------------------------------------------------//

// NewRootCmd is the cryptare application entry point.
func NewRootCmd(db *Database) *cobra.Command {
	var vim bool

	cmd := &cobra.Command{
		Use:   "cryptare",
		Short: "A file encryption and management tool",
		RunE: func(cmd *cobra.Command, args []string) error {
			p := tea.NewProgram(NewDashboardModelWithOptions(db, dashboardOptions{vimEnabled: vim}), tea.WithAltScreen())
			_, err := p.Run()
			return err
		},
	}

	cmd.AddCommand(
		newEncryptCmd(),
		newDecryptCmd(),
		newCompressCmd(),
		newDecompressCmd(),
		newKeysCmd(db),
	)
	cmd.Flags().BoolVar(&vim, "vim", false, "enable vim keybindings in the TUI")

	return cmd
}

//-----------------------------------------encrypt------------------------------------------------------//

func newEncryptCmd() *cobra.Command {
	var output, password string

	cmd := &cobra.Command{
		Use:   "encrypt [path]",
		Short: "Encrypt a file or directory with AES-256-GCM",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			src := args[0]
			if password == "" {
				var err error
				password, err = readPassword(cmd, "Enter password: ")
				if err != nil {
					return err
				}
			}
			if err := EncryptFile(src, output, password); err != nil {
				return err
			}
			dst := output
			if dst == "" {
				dst = src + encExt
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Encrypted: %s → %s\n", src, dst); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output file path (default: <path>.enc)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "encryption password")
	return cmd
}

//-----------------------------------------decrypt------------------------------------------------------//

func newDecryptCmd() *cobra.Command {
	var output, password string

	cmd := &cobra.Command{
		Use:   "decrypt [path]",
		Short: "Decrypt an AES-256-GCM encrypted file or directory archive",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			src := args[0]
			if password == "" {
				var err error
				password, err = readPassword(cmd, "Enter password: ")
				if err != nil {
					return err
				}
			}
			if err := DecryptFile(src, output, password); err != nil {
				return err
			}
			dst := output
			if dst == "" {
				dst = deriveDecryptOutput(src)
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Decrypted: %s → %s\n", src, dst); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output file or directory path")
	cmd.Flags().StringVarP(&password, "password", "p", "", "decryption password")
	return cmd
}

//-----------------------------------------compress-----------------------------------------------------//

func newCompressCmd() *cobra.Command {
	var output string
	var level int
	var format string

	cmd := &cobra.Command{
		Use:   "compress [path]",
		Short: "Compress a file or directory with gzip or zip",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			src := args[0]
			if err := CompressFileWithFormat(src, output, format, level); err != nil {
				return err
			}
			dst := output
			if dst == "" {
				dst = deriveCompressOutput(src, format)
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Compressed: %s → %s\n", src, dst); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output path (default: <file>.gz or <dir>.tar.gz for gzip, <path>.zip for zip)")
	cmd.Flags().StringVarP(&format, "format", "f", "", "compression format: gzip or zip (default: gzip)")
	cmd.Flags().IntVarP(&level, "level", "l", -1, "compression level 1-9 (default: -1 = default)")
	return cmd
}

//-----------------------------------------decompress---------------------------------------------------//

func newDecompressCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "decompress [archive]",
		Short: "Decompress gzip/tar.gz files or extract zip archives",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			src := args[0]
			if err := DecompressFile(src, output); err != nil {
				return err
			}
			dst := output
			if dst == "" {
				dst = deriveDecompressOutput(src)
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Decompressed: %s → %s\n", src, dst); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output path")
	return cmd
}

//-----------------------------------------keys---------------------------------------------------------//

func newKeysCmd(db *Database) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Manage encryption keys",
	}

	cmd.AddCommand(
		newKeysListCmd(db),
		newKeysGenerateCmd(db),
		newKeysExportCmd(db),
		newKeysImportCmd(db),
		newKeysDeleteCmd(db),
	)

	return cmd
}

func newKeysListCmd(db *Database) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List stored encryption keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			keys, err := db.ListKeys()
			if err != nil {
				return fmt.Errorf("list keys: %w", err)
			}
			if len(keys) == 0 {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), "No keys stored."); err != nil {
					return fmt.Errorf("write command output: %w", err)
				}
				return nil
			}
			w := cmd.OutOrStdout()
			if _, err := fmt.Fprintf(w, "%-20s  %-12s  %s\n", "KEY ID", "ALGORITHM", "CREATED"); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}
			for _, k := range keys {
				created := time.Unix(k.CreatedAt_, 0).Format("2006-01-02 15:04")
				if _, err := fmt.Fprintf(w, "%-20s  %-12s  %s\n", k.KeyID, k.Algorithm, created); err != nil {
					return fmt.Errorf("write command output: %w", err)
				}
			}
			return nil
		},
	}
}

func newKeysGenerateCmd(db *Database) *cobra.Command {
	var password string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate and store a new random encryption key",
		RunE: func(cmd *cobra.Command, args []string) error {
			if password == "" {
				var err error
				password, err = readPassword(cmd, "Enter master password to protect key: ")
				if err != nil {
					return err
				}
			}

			rawKey, err := GenerateKey()
			if err != nil {
				return err
			}

			keyID, err := newKeyID()
			if err != nil {
				return err
			}

			blob, err := EncryptKeyBlob(rawKey, password)
			if err != nil {
				return err
			}

			km := &KeyModel{
				KeyID:         keyID,
				Algorithm:     "AES-256-GCM",
				EncryptedBlob: blob,
				CreatedAt_:    time.Now().Unix(),
			}

			if err := db.SaveKey(km); err != nil {
				return fmt.Errorf("save key: %w", err)
			}

			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Generated key: %s\n", keyID); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&password, "password", "p", "", "master password for key protection")
	return cmd
}

func newKeysExportCmd(db *Database) *cobra.Command {
	var output, password string

	cmd := &cobra.Command{
		Use:   "export [key-id]",
		Short: "Export an encrypted key to a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			keyID := args[0]

			km, err := db.GetKey(keyID)
			if err != nil {
				return fmt.Errorf("key not found: %w", err)
			}

			if password == "" {
				password, err = readPassword(cmd, "Enter master password: ")
				if err != nil {
					return err
				}
			}

			if err := ExportKeyToFile(km, password, output); err != nil {
				return err
			}

			if output == "" {
				output = fmt.Sprintf("%s-%d.ckey", keyID, time.Now().Unix())
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Exported key %s → %s\n", keyID, output); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output file path (default: <key-id>-<timestamp>.ckey)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "master password for export encryption")
	return cmd
}

func newKeysImportCmd(db *Database) *cobra.Command {
	var password string

	cmd := &cobra.Command{
		Use:   "import [file]",
		Short: "Import an encrypted key from a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			if password == "" {
				var err error
				password, err = readPassword(cmd, "Enter master password: ")
				if err != nil {
					return err
				}
			}

			km, err := ImportKeyFromFile(path, password)
			if err != nil {
				return err
			}

			if err := db.SaveKey(km); err != nil {
				return fmt.Errorf("save imported key: %w", err)
			}

			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Imported key: %s\n", km.KeyID); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&password, "password", "p", "", "master password used when key was exported")
	return cmd
}

func newKeysDeleteCmd(db *Database) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete [key-id]",
		Short: "Delete a stored encryption key (irreversible)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			keyID := args[0]

			if !yes {
				ok, err := confirmAction(cmd, fmt.Sprintf("Delete key %q? This cannot be undone. [y/N]: ", keyID))
				if err != nil {
					return err
				}
				if !ok {
					return errors.New("key deletion aborted by user")
				}
			}

			if err := db.DeleteKey(keyID); err != nil {
				if errors.Is(err, ErrKeyNotFound) {
					return fmt.Errorf("key %q not found", keyID)
				}
				return fmt.Errorf("delete key: %w", err)
			}

			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Deleted key: %s\n", keyID); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "delete without confirmation prompt")
	cmd.Flags().BoolVar(&yes, "force", false, "delete without confirmation prompt")
	return cmd
}

//-----------------------------------------helpers------------------------------------------------------//

// readPassword writes prompt to stderr and reads one line from the command's input,
// keeping spaces and removing only the line ending. When the input is a terminal,
// the line is read without echoing it.
func readPassword(cmd *cobra.Command, prompt string) (string, error) {
	if _, err := fmt.Fprint(cmd.ErrOrStderr(), prompt); err != nil {
		return "", fmt.Errorf("write password prompt: %w", err)
	}

	in := cmd.InOrStdin()
	if f, ok := in.(*os.File); ok && term.IsTerminal(f.Fd()) {
		pwd, err := readTerminalPassword(f.Fd(), cmd.ErrOrStderr())
		// Enter isn't echoed either, so end the prompt line before any further output.
		if _, werr := fmt.Fprintln(cmd.ErrOrStderr()); err == nil && werr != nil {
			err = werr
		}
		if err != nil {
			return "", fmt.Errorf("read password: %w", err)
		}
		return string(pwd), nil
	}

	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && (!errors.Is(err, io.EOF) || line == "") {
		return "", fmt.Errorf("read password: %w", err)
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// readTerminalPassword reads one line from the terminal fd without echo. Ctrl+C
// would otherwise kill the process with echo still off, so while the read is in
// progress an interrupt restores the terminal and exits with status 130
// (128 + SIGINT), matching what the shell reports for an interrupted command.
func readTerminalPassword(fd uintptr, out io.Writer) ([]byte, error) {
	state, err := term.GetState(fd)
	if err != nil {
		return nil, fmt.Errorf("save terminal state: %w", err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	done := make(chan struct{})
	defer func() {
		signal.Stop(interrupt)
		close(done)
	}()

	// Owned by this call: it ends when the read returns (done) or after handling Ctrl+C.
	go func() {
		select {
		case <-interrupt:
			_ = term.Restore(fd, state)
			_, _ = fmt.Fprintln(out)
			os.Exit(130)
		case <-done:
		}
	}()

	return term.ReadPassword(fd)
}

func confirmAction(cmd *cobra.Command, prompt string) (bool, error) {
	if _, err := fmt.Fprint(cmd.ErrOrStderr(), prompt); err != nil {
		return false, fmt.Errorf("write confirmation prompt: %w", err)
	}

	line, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("read confirmation: %w", err)
	}

	resp := strings.TrimSpace(line)
	return strings.EqualFold(resp, "y") || strings.EqualFold(resp, "yes"), nil
}

func deriveDecryptOutput(src string) string {
	if len(src) > len(encExt) && src[len(src)-len(encExt):] == encExt {
		return src[:len(src)-len(encExt)]
	}
	return src + ".dec"
}

func deriveCompressOutput(src, format string) string {
	info, err := os.Stat(src)
	selectedFormat, formatErr := resolveCompressFormat(format, "")
	if formatErr != nil {
		selectedFormat = formatGzip
	}
	if err == nil {
		return defaultCompressOutput(src, info.IsDir(), selectedFormat)
	}
	if selectedFormat == formatZip {
		return src + zipExt
	}
	return src + gzExt
}

func deriveDecompressOutput(src string) string {
	return defaultDecompressOutput(src)
}
