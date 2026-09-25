# Cryptare

Cryptare is a terminal tool for encrypting, decrypting, compressing and extracting files and directories. It combines a scriptable command-line interface with an interactive terminal UI and a small local store for password-protected keys. It is meant for people who want password-based file protection and archiving from the command line.

## Features

- **File Encryption & Decryption**
  - Encrypt files with AES-256-GCM
  - Encrypt directories into a single encrypted archive
  - Decrypt previously encrypted files
  - Decrypt encrypted directory archives back into their original tree
  - Optional custom output paths
  - Password-based protection for secure workflows

- **Compression & Decompression**
  - Compress files with gzip
  - Compress files or directories as zip archives
  - Compress directories as tar.gz archives
  - Decompress gzip files, tar.gz archives, and zip archives
  - Optional compression levels
  - Automatic output name derivation

- **Key Management**
  - Generate and store encryption keys
  - List stored keys
  - Export keys to encrypted files
  - Import keys from encrypted files
  - Delete stored keys with confirmation safeguards

- **Interactive TUI**
  - Terminal user interface for file and key management
  - Launches automatically when run without subcommands

## Use cases

- Protect a sensitive document before copying it to a USB drive or cloud storage: `cryptare encrypt ./tax-return.pdf`
- Back up a project folder as one encrypted file, then restore it later: `cryptare encrypt ./project-dir --output ./project-dir-backup.enc`, then `cryptare decrypt ./project-dir-backup.enc --output ./restored-project-dir`
- Package build output as a tar.gz or zip archive for sharing, and extract archives you receive: `cryptare compress ./build --format zip`, `cryptare decompress ./build-backup.tar.gz`
- Use menus instead of flags: run `cryptare` (or `cryptare --vim` for vim-style keys)

## Prerequisites

- To build from source: Go 1.26 or newer, plus a C compiler (gcc or clang) with CGO enabled. The SQLite driver used for the key store requires CGO. On Windows this means a CGO-capable toolchain such as MinGW-w64.
- Optional: Docker, to build and run the container image.

## Installation

### Build from source

```bash
git clone https://github.com/jabbott-iii/Cryptare.git
cd Cryptare
CGO_ENABLED=1 go build -o cryptare .
sudo mv cryptare /usr/local/bin/cryptare   # optional: put it on your PATH
```

### Release archives

Tagged releases publish these assets, built by `.github/workflows/cd.yml`, together with a `checksums.txt` file (SHA-256):

| Platform | Asset |
|---|---|
| Linux x86-64 | `cryptare_linux_amd64.tar.gz` |
| Linux ARM64 | `cryptare_linux_arm64.tar.gz` |
| macOS Intel | `cryptare_darwin_amd64.tar.gz` |
| macOS Apple silicon | `cryptare_darwin_arm64.tar.gz` |
| Windows x86-64 | `cryptare_windows_amd64.zip` |
| Windows ARM64 | `cryptare_windows_arm64.zip` |

> ⚠️ **Known issue:** every release published so far (up to and including v1.0.0) was built with CGO disabled. Those binaries exit on every command with `go-sqlite3 requires cgo to work`. Build from source until a fixed release is available. See BUG-001 in [intel/notes.md](intel/notes.md).

Once a working release is available, install it on Linux like this:

```bash
sha256sum --ignore-missing -c checksums.txt
tar -xzf cryptare_linux_amd64.tar.gz
chmod +x cryptare_linux_amd64
sudo mv cryptare_linux_amd64 /usr/local/bin/cryptare
```

- macOS: use the matching `darwin` archive in the same way, and check its hash with `shasum -a 256`.
- Windows: extract the zip, rename `cryptare_windows_amd64.exe` to `cryptare.exe`, and add it to your PATH.

## Quick start

```bash
cryptare encrypt ./secret.txt        # prompts for a password; prints "Encrypted: ./secret.txt → ./secret.txt.enc"
cryptare decrypt ./secret.txt.enc    # prompts for the password; prints "Decrypted: ./secret.txt.enc → ./secret.txt"
cryptare                             # opens the interactive TUI
```

- Decryption writes to the output path without asking, replacing any existing file there.
- The password prompt reads the whole line, spaces included, and hides what you type when run in a terminal. When input is piped in, the first line is used.
- ⚠️ **Upgrading from v1.0.1 or earlier:** the old prompt kept only the text before the first space. If you encrypted a file at the prompt with a multi-word passphrase, decrypt it with just the first word. See SEC-002 in [intel/cybersec.md](intel/cybersec.md).

## Core CLI capabilities

Cryptare is organized into focused command groups:

- cryptare encrypt — encrypt a file or directory
- cryptare decrypt — decrypt a file or encrypted directory archive
- cryptare compress — compress a file or directory
- cryptare decompress — decompress a gzip file or extract a tar.gz/zip archive
- cryptare keys — manage encryption keys
- cryptare --version (or -v) — print the version: release builds report their tag, local builds report `dev`

### encrypt

- cryptare encrypt [path] — encrypt a file with AES-256-GCM or package a directory into a single encrypted archive
- cryptare encrypt [path] --output [path] — write to a custom output file
- cryptare encrypt [path] --password [value] — provide the encryption password non-interactively

Examples:
- cryptare encrypt ./secret.txt
- cryptare encrypt ./secret.txt --output ./secret.txt.enc
- cryptare encrypt ./project-dir --output ./project-dir-backup.enc
- cryptare encrypt ./secret.txt --password "correct horse battery staple"

### decrypt

- cryptare decrypt [path] — decrypt an AES-256-GCM encrypted file or restore an encrypted directory archive
- cryptare decrypt [path] --output [path] — write to a custom output file or restore into a target directory
- cryptare decrypt [path] --password [value] — provide the decryption password non-interactively

Examples:
- cryptare decrypt ./secret.txt.enc
- cryptare decrypt ./secret.txt.enc --output ./secret.txt
- cryptare decrypt ./project-dir-backup.enc --output ./restored-project-dir
- cryptare decrypt ./secret.txt.enc --password "correct horse battery staple"

### compress

- cryptare compress [path] — compress with gzip (default) or zip
- cryptare compress [path] --output [path] — write to a custom output file or archive
- cryptare compress [path] --format [gzip|zip] — select compression format
- cryptare compress [path] --level [1-9] — set the compression level (applies to gzip and zip)
- zip is also selected automatically when --output ends in .zip

Examples:
- cryptare compress ./artifact.bin
- cryptare compress ./artifact.bin --output ./artifact.bin.gz
- cryptare compress ./build --output ./build-backup.tar.gz
- cryptare compress ./build --format zip --output ./build-backup.zip
- cryptare compress ./artifact.bin --level 9

### decompress

- cryptare decompress [archive] — decompress a gzip file or extract a tar.gz/zip archive
- cryptare decompress [archive] --output [path] — write to a custom output file or extract to a directory

Examples:
- cryptare decompress ./artifact.bin.gz
- cryptare decompress ./artifact.bin.gz --output ./artifact.bin
- cryptare decompress ./build-backup.tar.gz --output ./restored-build
- cryptare decompress ./build-backup.zip --output ./restored-build

### keys

- cryptare keys list — list stored encryption keys
- cryptare keys generate — generate and store a new random encryption key
- cryptare keys export [key-id] — export an encrypted key to a file
- cryptare keys import [file] — import an encrypted key from a file
- cryptare keys delete [key-id] — delete a stored encryption key (irreversible)
- cryptare keys delete [key-id] --yes — delete non-interactively (automation)
- cryptare keys delete [key-id] --force — delete non-interactively (automation)

Examples:
- cryptare keys list
- cryptare keys generate --password "master password"
- cryptare keys export key-123 --output key-123.ckey
- cryptare keys import ./key-123.ckey
- cryptare keys delete key-123
- cryptare keys delete key-123 --yes

⚠️ Key deletion is permanent. Once deleted, the stored key cannot be recovered.

### Interactive TUI

- Running cryptare with no subcommand launches the terminal UI dashboard for interactive file and key management.
- Run `cryptare --vim` to enable vim-style TUI bindings.
- With vim bindings enabled:
  - Menus accept `j`/`k` to move, `l` to select, and `Esc`, `h`, or `b` to go back when a back action exists.
  - Forms start in insert mode so file paths and passwords can be typed normally.
  - Press `Esc` in a form to switch to normal mode, then use `j`/`k` to change fields, `h` to cancel, `l` or `Enter` to advance, and `i`/`a`/`o` to return to insert mode. `o` moves to the next field before re-entering insert mode.

## Configuration

| Setting | Default | Description |
|---|---|---|
| `CRYPTARE_DB_PATH` (environment variable) | `cryptare.db` in the current working directory | Location of the SQLite key store. It is opened, and created if missing, on every run, including `--help`. |
| `--vim` (flag) | off | Turns on vim-style key bindings in the TUI. |

Command flags (`--output`, `--password`, `--format`, `--level`, `--yes`/`--force`) are described under [Core CLI capabilities](#core-cli-capabilities).

## Docker

The image builds Cryptare with CGO enabled for `linux/amd64`. It sets `CRYPTARE_DB_PATH=/app/data/cryptare.db` and declares `/app/data` as a volume.

### Build
```bash
docker build -t cryptare:latest .
```

### Run (interactive)
```bash
docker run --rm -it cryptare:latest
```

### Persist data
```bash
mkdir -p ~/.cryptare
docker run --rm -it \
  -v ~/.cryptare:/app/data \
  -e CRYPTARE_DB_PATH=/app/data/cryptare.db \
  cryptare:latest
```

### CLI usage
```bash
docker run --rm -it cryptare:latest --help
docker run --rm -it -v ~/.cryptare:/app/data cryptare:latest keys list
```

## Testing and quality checks

Run these from the repository root (they mirror CI):

```bash
gofmt -s -l .          # lists files that need formatting; should print nothing
go vet ./...
golangci-lint run      # CI uses v2.13.2
go test ./...
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full validation workflow.

## Project structure

```text
main.go, database_path.go   entry point and CRYPTARE_DB_PATH handling
internal/crypto.go          AES-256-GCM encryption, key blobs, key export/import
internal/compress.go        gzip, tar.gz and zip creation and extraction
internal/database.go        SQLite key store (GORM)
internal/logic-cli.go       Cobra commands
internal/ui-dashboard.go,
internal/logic-tui.go       Bubble Tea terminal UI
intel/                      architecture, security, plans and repository map
.github/workflows/          CI, CD, Docker and security workflows
```

[intel/map.md](intel/map.md) has the detailed map and diagrams.

## Contributing and license

See [CONTRIBUTING.md](CONTRIBUTING.md). Cryptare is licensed under the Apache License 2.0; see [LICENSE](LICENSE) and [NOTICE](NOTICE).
