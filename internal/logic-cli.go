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
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
)

//-----------------------------------------core---------------------------------------------------------//

// DatabaseOpener opens the key database. The CLI calls it only for the commands that
// use the key store, so other commands never create the database file (SEC-010).
type DatabaseOpener func() (*Database, error)

// NewRootCmd is the cryptare application entry point, using an already-open database.
func NewRootCmd(db *Database) *cobra.Command {
	return NewRootCmdLazy(func() (*Database, error) { return db, nil })
}

// NewRootCmdLazy builds the root command around open, which is called at most once,
// and only by the keys commands and the TUI.
func NewRootCmdLazy(open DatabaseOpener) *cobra.Command {
	var vim bool
	openDB := openOnce(open)

	cmd := &cobra.Command{
		Use:   "cryptare",
		Short: "A file encryption and management tool",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			p := tea.NewProgram(NewDashboardModelWithOptions(db, dashboardOptions{vimEnabled: vim}), tea.WithAltScreen())
			_, err = p.Run()
			return err
		},
	}

	cmd.AddCommand(
		newEncryptCmd(),
		newDecryptCmd(),
		newCompressCmd(),
		newDecompressCmd(),
		newKeysCmd(openDB),
	)
	cmd.Flags().BoolVar(&vim, "vim", false, "enable vim keybindings in the TUI")

	return cmd
}

//-----------------------------------------encrypt------------------------------------------------------//

func newEncryptCmd() *cobra.Command {
	var output string
	var pw passwordFlags
	var force bool

	cmd := &cobra.Command{
		Use:   "encrypt [path]",
		Short: "Encrypt a file or directory with AES-256-GCM",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			src := args[0]
			dst := output
			if dst == "" {
				dst = src + encExt
			}
			if err := CheckOutputPath(src, dst, force); err != nil {
				return withForceHint(err)
			}
			password, given, err := pw.get(cmd)
			if err != nil {
				return err
			}
			if !given {
				password, err = readNewPassword(cmd, "Enter password: ")
				if err != nil {
					return err
				}
			}
			if err := EncryptFile(src, dst, password); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Encrypted: %s → %s\n", src, dst); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output file path (default: <path>.enc)")
	pw.register(cmd, "encryption password")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite the output if it already exists")
	return cmd
}

//-----------------------------------------decrypt------------------------------------------------------//

func newDecryptCmd() *cobra.Command {
	var output string
	var pw passwordFlags
	var force bool
	var limitFlags extractLimitFlags

	cmd := &cobra.Command{
		Use:   "decrypt [path]",
		Short: "Decrypt an AES-256-GCM encrypted file or directory archive",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			limits, err := limitFlags.limits()
			if err != nil {
				return err
			}
			src := args[0]
			dst := output
			if dst == "" {
				dst = deriveDecryptOutput(src)
			}
			if err := CheckOutputPath(src, dst, force); err != nil {
				return withForceHint(err)
			}
			password, given, err := pw.get(cmd)
			if err != nil {
				return err
			}
			if !given {
				password, err = readPassword(cmd, "Enter password: ")
				if err != nil {
					return err
				}
			}
			if err := DecryptFileWithLimits(src, dst, password, limits); err != nil {
				return withLimitHint(err)
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Decrypted: %s → %s\n", src, dst); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output file or directory path")
	pw.register(cmd, "decryption password")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite the output if it already exists")
	limitFlags.register(cmd)
	return cmd
}

//-----------------------------------------compress-----------------------------------------------------//

func newCompressCmd() *cobra.Command {
	var output string
	var level int
	var format string
	var force bool

	cmd := &cobra.Command{
		Use:   "compress [path]",
		Short: "Compress a file or directory with gzip or zip",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			src := args[0]
			if _, err := resolveCompressFormat(format, output); err != nil {
				return err
			}
			dst := output
			if dst == "" {
				dst = deriveCompressOutput(src, format)
			}
			if err := CheckOutputPath(src, dst, force); err != nil {
				return withForceHint(err)
			}
			if err := CompressFileWithFormat(src, dst, format, level); err != nil {
				return err
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
	cmd.Flags().BoolVar(&force, "force", false, "overwrite the output if it already exists")
	return cmd
}

//-----------------------------------------decompress---------------------------------------------------//

func newDecompressCmd() *cobra.Command {
	var output string
	var force bool
	var limitFlags extractLimitFlags

	cmd := &cobra.Command{
		Use:   "decompress [archive]",
		Short: "Decompress gzip/tar.gz files or extract zip archives",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			limits, err := limitFlags.limits()
			if err != nil {
				return err
			}
			src := args[0]
			dst := output
			if dst == "" {
				dst = deriveDecompressOutput(src)
			}
			if err := CheckOutputPath(src, dst, force); err != nil {
				return withForceHint(err)
			}
			if err := DecompressFileWithLimits(src, dst, limits); err != nil {
				return withLimitHint(err)
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Decompressed: %s → %s\n", src, dst); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output path")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite the output if it already exists")
	limitFlags.register(cmd)
	return cmd
}

//-----------------------------------------keys---------------------------------------------------------//

func newKeysCmd(open DatabaseOpener) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Manage encryption keys",
	}

	cmd.AddCommand(
		newKeysListCmd(open),
		newKeysGenerateCmd(open),
		newKeysExportCmd(open),
		newKeysImportCmd(open),
		newKeysDeleteCmd(open),
	)

	return cmd
}

func newKeysListCmd(open DatabaseOpener) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List stored encryption keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := open()
			if err != nil {
				return err
			}
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

func newKeysGenerateCmd(open DatabaseOpener) *cobra.Command {
	var pw passwordFlags

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate and store a new random encryption key",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := open()
			if err != nil {
				return err
			}
			password, given, err := pw.get(cmd)
			if err != nil {
				return err
			}
			if !given {
				password, err = readNewPassword(cmd, "Enter master password to protect key: ")
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

	pw.register(cmd, "master password for key protection")
	return cmd
}

func newKeysExportCmd(open DatabaseOpener) *cobra.Command {
	var output string
	var pw passwordFlags

	cmd := &cobra.Command{
		Use:   "export [key-id]",
		Short: "Export an encrypted key to a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := open()
			if err != nil {
				return err
			}
			keyID := args[0]

			km, err := db.GetKey(keyID)
			if err != nil {
				return fmt.Errorf("key not found: %w", err)
			}

			password, given, err := pw.get(cmd)
			if err != nil {
				return err
			}
			if !given {
				password, err = readNewPassword(cmd, "Enter master password: ")
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
	pw.register(cmd, "master password for export encryption")
	return cmd
}

func newKeysImportCmd(open DatabaseOpener) *cobra.Command {
	var pw passwordFlags

	cmd := &cobra.Command{
		Use:   "import [file]",
		Short: "Import an encrypted key from a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := open()
			if err != nil {
				return err
			}
			path := args[0]

			password, given, err := pw.get(cmd)
			if err != nil {
				return err
			}
			if !given {
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

	pw.register(cmd, "master password used when key was exported")
	return cmd
}

func newKeysDeleteCmd(open DatabaseOpener) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete [key-id]",
		Short: "Delete a stored encryption key (irreversible)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := open()
			if err != nil {
				return err
			}
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

// extractLimitFlags holds the --max-size and --max-entries flags of the commands that
// extract archives (decompress, and decrypt for encrypted directories).
type extractLimitFlags struct {
	maxSize    string
	maxEntries int
}

func (f *extractLimitFlags) register(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.maxSize, "max-size", formatSize(DefaultMaxExtractBytes),
		"stop extracting once the output passes this size, e.g. 500MB or 20GiB (0 = no limit)")
	cmd.Flags().IntVar(&f.maxEntries, "max-entries", DefaultMaxExtractEntries,
		"stop extracting an archive with more entries than this (0 = no limit)")
}

// limits validates the flag values and returns them as ExtractLimits.
func (f *extractLimitFlags) limits() (ExtractLimits, error) {
	size, err := parseSize(f.maxSize)
	if err != nil {
		return ExtractLimits{}, fmt.Errorf("invalid --max-size: %w", err)
	}
	if f.maxEntries < 0 {
		return ExtractLimits{}, fmt.Errorf("invalid --max-entries %d: use 0 or more", f.maxEntries)
	}
	return ExtractLimits{MaxBytes: size, MaxEntries: f.maxEntries}, nil
}

// parseSize parses a byte count such as "1024", "500MB" or "10 GiB". The number must
// be whole. Units are B, KB, MB, GB and TB (powers of 1000) or KiB, MiB, GiB and TiB
// (powers of 1024), in any letter case.
func parseSize(s string) (int64, error) {
	t := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(s), " ", ""))
	digits := 0
	for digits < len(t) && t[digits] >= '0' && t[digits] <= '9' {
		digits++
	}
	if digits == 0 {
		return 0, fmt.Errorf("size %q must start with a whole number", s)
	}
	n, err := strconv.ParseInt(t[:digits], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("size %q is too large", s)
	}
	units := map[string]int64{
		"": 1, "b": 1,
		"kb": 1e3, "mb": 1e6, "gb": 1e9, "tb": 1e12,
		"kib": 1 << 10, "mib": 1 << 20, "gib": 1 << 30, "tib": 1 << 40,
	}
	unit, ok := units[t[digits:]]
	if !ok {
		return 0, fmt.Errorf("size %q has an unknown unit (use B, KB, MB, GB, TB, KiB, MiB, GiB or TiB)", s)
	}
	if n > math.MaxInt64/unit {
		return 0, fmt.Errorf("size %q is too large", s)
	}
	return n * unit, nil
}

// openOnce wraps open so that it runs at most once, adding context to its error.
func openOnce(open DatabaseOpener) DatabaseOpener {
	return sync.OnceValues(func() (*Database, error) {
		db, err := open()
		if err != nil {
			return nil, fmt.Errorf("open key database: %w", err)
		}
		return db, nil
	})
}

// passwordFlags holds a command's --password and --password-file flags (SEC-004).
type passwordFlags struct {
	value string
	file  string
}

// passwordFlagWarning is printed whenever --password is used.
const passwordFlagWarning = "warning: --password can be seen by other users and is saved in shell history; " +
	"use the prompt, piped input or --password-file instead"

func (p *passwordFlags) register(cmd *cobra.Command, usage string) {
	cmd.Flags().StringVarP(&p.value, "password", "p", "", usage+" (visible to other users; prefer --password-file)")
	cmd.Flags().StringVar(&p.file, "password-file", "", "read the password from the first line of this file")
	cmd.MarkFlagsMutuallyExclusive("password", "password-file")
}

// get returns the password given by --password-file or --password. given is false when
// neither flag was set, and the caller then prompts. --password prints a warning to
// stderr, because command-line arguments are visible to other users.
func (p *passwordFlags) get(cmd *cobra.Command) (password string, given bool, err error) {
	if p.file != "" {
		password, err := readPasswordFile(p.file)
		if err != nil {
			return "", false, err
		}
		return password, true, nil
	}
	if p.value == "" {
		return "", false, nil
	}
	if _, err := fmt.Fprintln(cmd.ErrOrStderr(), passwordFlagWarning); err != nil {
		return "", false, fmt.Errorf("write warning: %w", err)
	}
	return p.value, true, nil
}

// maxPasswordFileSize is the most that is read from a password file.
const maxPasswordFileSize = 64 << 10

// readPasswordFile returns the first line of the file at path, keeping spaces and
// removing only the line ending, the same way a password piped to the prompt is read.
func readPasswordFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open password file: %w", err)
	}
	defer func() { _ = f.Close() }() // read-only: a close error can't lose data

	line, err := bufio.NewReader(io.LimitReader(f, maxPasswordFileSize)).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read password file: %w", err)
	}
	switch {
	case line == "":
		return "", fmt.Errorf("password file %s is empty", path)
	case !strings.HasSuffix(line, "\n") && len(line) == maxPasswordFileSize:
		return "", fmt.Errorf("password file %s: first line is too long (over %s)", path, formatSize(maxPasswordFileSize))
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// withLimitHint tells CLI users how to change the extraction limits.
func withLimitHint(err error) error {
	if errors.Is(err, ErrExtractLimit) {
		return fmt.Errorf("%w (use --max-size or --max-entries to change the limit; 0 means no limit)", err)
	}
	return err
}

// withForceHint tells CLI users how to overwrite an existing output on purpose.
func withForceHint(err error) error {
	if errors.Is(err, ErrOutputExists) {
		return fmt.Errorf("%w (use --force to overwrite)", err)
	}
	return err
}

// readPassword writes prompt to stderr and reads one line from the command's input,
// keeping spaces and removing only the line ending. When the input is a terminal,
// the line is read without echoing it.
func readPassword(cmd *cobra.Command, prompt string) (string, error) {
	if _, err := fmt.Fprint(cmd.ErrOrStderr(), prompt); err != nil {
		return "", fmt.Errorf("write password prompt: %w", err)
	}

	if f, ok := terminalInput(cmd); ok {
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

	line, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil && (!errors.Is(err, io.EOF) || line == "") {
		return "", fmt.Errorf("read password: %w", err)
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// readNewPassword prompts for a password that will protect new data. The password
// must meet the password policy (CheckPasswordPolicy). When it is typed at a terminal
// it is asked for a second time, so a typo can't make the data unrecoverable; piped
// input is read once, so scripts keep working.
func readNewPassword(cmd *cobra.Command, prompt string) (string, error) {
	_, interactive := terminalInput(cmd)
	read := func(p string) (string, error) { return readPassword(cmd, p) }
	return readNewPasswordWith(read, prompt, interactive)
}

// readNewPasswordWith implements readNewPassword. read shows a prompt and returns
// the line typed, and confirm says whether to ask for the password a second time.
// The policy is checked first, so a rejected password isn't typed twice.
func readNewPasswordWith(read func(prompt string) (string, error), prompt string, confirm bool) (string, error) {
	password, err := read(prompt)
	if err != nil {
		return "", err
	}
	if err := CheckPasswordPolicy(password); err != nil {
		return "", err
	}
	if !confirm {
		return password, nil
	}

	again, err := read("Confirm password: ")
	if err != nil {
		return "", err
	}
	if again != password {
		return "", ErrPasswordMismatch
	}
	return password, nil
}

// terminalInput returns the command's input file when it is a terminal.
func terminalInput(cmd *cobra.Command) (*os.File, bool) {
	f, ok := cmd.InOrStdin().(*os.File)
	if !ok || !term.IsTerminal(f.Fd()) {
		return nil, false
	}
	return f, true
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
