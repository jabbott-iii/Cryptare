# Architecture and Maintainability Guide

This is the authoritative source for Cryptare's architecture and maintainability
rules (see `AGENTS.md`). `CONTRIBUTING.md` must stay consistent with it.

Last reviewed: 2026-09-23

## 1. System overview

Cryptare is a single-binary Go CLI/TUI for local file protection:

- password-based encryption and decryption of files and directories (AES-256-GCM);
- compression and extraction (gzip, tar.gz, zip);
- a local key store (SQLite) for randomly generated, password-protected keys.

It has no network surface: there is no HTTP server and nothing calls a remote service.

- Module: `github.com/jabbott-iii/Cryptare`; `go.mod` declares `go 1.26.0`.
- Entry point: `main.go` opens the database, builds the Cobra root command and runs it.
- All application logic lives in one flat Go package, `internal/`.

## 2. Layers and responsibilities

| Layer | Files | Responsibility | May depend on |
|---|---|---|---|
| Entry | `main.go`, `database_path.go` | Work out the DB path (`CRYPTARE_DB_PATH`, default `cryptare.db` in the current directory), open the DB, run the root command | `internal` |
| Interfaces | `internal/logic-cli.go` (Cobra); `internal/ui-dashboard.go` + `internal/logic-tui.go` (Bubble Tea) | Parse input, prompt, call core/storage, render results | core, storage |
| Core operations | `internal/crypto.go`, `internal/compress.go` | Encryption formats, key blobs, key export/import, archive creation/extraction | stdlib, `golang.org/x/crypto` |
| Storage | `internal/database.go` | GORM/SQLite schema (`KeyModel`) and key CRUD | GORM, SQLite driver |

Rules:

1. **Core stays UI-agnostic.** Functions in `crypto.go` and `compress.go` take paths,
   passwords or bytes and return `error`. They do not print, prompt, or import Cobra or
   Bubble Tea.
2. **The CLI and TUI must match.** Today the TUI's `buildActionCmd` re-implements the
   CLI's key generate/export/import flows. Until those flows move into shared core
   functions (`intel/plan.md`), a change to one interface must be mirrored in the other
   and tested in both.
3. **Data-protecting validation goes in core.** Checks such as non-empty passwords,
   overwrite protection and extraction limits belong in the core layer so both
   interfaces get them. The TUI accepting empty passwords (SEC-001) is what happens
   when this rule is missed.
4. **One persistence type.** `*Database` is the only persistence type. The `Storage`
   interface is declared but callers don't use it. Either adopt it at the interface
   layer (for test doubles) or remove it, in a change dedicated to that.
5. **Only key commands and the TUI need the database.** Today `main.go` opens (and
   creates) it for every command. Don't add new dependencies on the DB from
   file-only commands.

## 3. On-disk formats (compatibility contract)

Existing artifacts must stay readable. Changing any format below requires a versioned
header, a legacy read path, and tests covering both.

| Artifact | Layout | Notes |
|---|---|---|
| Encrypted file (`.enc`) | `salt(16) ‖ nonce(12) ‖ AES-256-GCM ciphertext+tag` | Key = PBKDF2-HMAC-SHA256(password, salt, 100,000 iterations, 32 bytes). No header, version or AAD. |
| Encrypted directory (`.enc`) | `"CRYPTARE-DIR-ENC\x00" ‖ salt ‖ nonce ‖ ciphertext+tag` | Plaintext is a tar.gz of the tree. The magic string is also the GCM AAD, so a stripped magic can't be decrypted as a single file. |
| Stored key (`key_models.encrypted_blob`) | base64(`salt ‖ nonce ‖ GCM(32-byte raw key)`) | Encrypted with the master password using the same KDF. |
| Key export (`.ckey`) | base64(`salt ‖ nonce ‖ GCM(JSON KeyExport{version:1,…})`) | The JSON carries the already-encrypted stored blob; the outer layer uses the export password. |
| Compressed output | `.gz` (single file, gzip header `Name` set), `.tar.gz` (directory), `.zip` | Archive entries use forward-slash relative paths. Symlinks and special files are rejected. |

`TestDecryptLegacyEncryptedFile` guards the single-file format; keep it passing.

## 4. Coding conventions

These are observed in the codebase and required for new code:

- The Apache-2.0 license header appears at the top of every `.go` file.
- Code is formatted with `gofmt -s` and passes `go vet` and `golangci-lint`. CI pins
  golangci-lint v2.13.2 with its default linters; there is no config file.
- Errors are wrapped with context, e.g. `fmt.Errorf("read source file: %w", err)`, with
  lower-case messages. Errors from output writes (`fmt.Fprintf`) are checked, as in the
  existing commands.
- Use `closeWithError` (in `compress.go`) for deferred `Close` on writers so close
  failures aren't lost.
- Files the tool writes use mode `0o600`; directories it creates currently use `0o755`.
- Symlinks and non-regular files are rejected when reading directory trees, and path
  traversal (`..`) is rejected when extracting. Keep both behaviours.
- Code is grouped with the existing `//----- section -----//` banner comments.
- File naming is mixed (`logic-cli.go` vs `logic_cli_test.go`). Don't rename existing
  files as part of unrelated changes. Name new files in `snake_case.go` (Go convention).
- Keep dependencies minimal and prefer the standard library. Current direct
  dependencies:
  - Cobra (CLI)
  - Bubble Tea and Lip Gloss (TUI)
  - GORM with `gorm.io/driver/sqlite`, which uses `github.com/mattn/go-sqlite3` (CGO)
  - `golang.org/x/crypto`, used for `pbkdf2`. This package is now a frozen wrapper
    around the standard library's `crypto/pbkdf2`.

  The existing indirect dependency `github.com/charmbracelet/x/term` already provides
  `ReadPassword` and `IsTerminal`.

## 5. Testing

- Tests sit beside the code (`internal/*_test.go`, `database_path_test.go`). They use
  `t.TempDir()` and must not write anywhere else.
- Database tests use `newTestDatabase(t, useFile)`. Close the SQL handles so Windows
  CI can delete the temporary files.
- TUI tests drive `DashboardModel.Update` and `updateForm` with synthetic
  `tea.KeyMsg` values; follow `internal/logic_tui_test.go`.
- CLI tests run `NewRootCmd(db)` with `SetArgs`, `SetIn` and `SetOut`; follow
  `internal/logic_cli_test.go`.
- Every security fix needs a regression test that fails before the fix. Record it in
  `intel/cybersec.md`.
- Baseline on 2026-09-23 (Go 1.26.8, linux/amd64): `go test -race ./...` passes with
  59.8% statement coverage. These have no tests: `readPassword`, `ImportKeyFromFile`,
  the TUI `View`, and archive path-traversal rejection.

## 6. Build and release

- **CGO is required.** The SQLite driver (`gorm.io/driver/sqlite`, which uses
  `github.com/mattn/go-sqlite3`) needs CGO. A `CGO_ENABLED=0` build compiles but fails
  at startup on every command, including `--help`, because `main.go` opens the DB
  first. Every build or release pipeline must either enable CGO with a C toolchain for
  each target or switch to a pure-Go SQLite driver. That decision is Q-001 in
  `intel/notes.md`; the defect is BUG-001.
- **Releases** are cut by pushing a `vX.Y.Z` tag (`make release VERSION=vX.Y.Z`), which
  triggers `.github/workflows/cd.yml`. It publishes
  `cryptare_<os>_<arch>.tar.gz`/`.zip` archives plus `checksums.txt`.
- **Version stamping:** `cd.yml` passes `-X main.version=…` to the linker, but no
  `version` variable exists in package `main`, so the flag has no effect. Add the
  variable (and a `--version` flag) before relying on it.
- **Docker:** the `Dockerfile` builds with CGO for `linux/amd64` only (`GOARCH=amd64`).

## 7. Change checklist

Before opening a pull request:

1. `gofmt -s -l .` prints nothing.
2. `go mod tidy` leaves `go.mod`/`go.sum` unchanged, unless the change adds a
   dependency on purpose.
3. `go vet ./...` and `golangci-lint run` are clean.
4. `go test ./...` passes. Use `-race` when a C toolchain is available.
5. The CLI and TUI paths are both updated and tested.
6. On-disk formats stay backward-compatible (§3).
7. `intel/` docs are updated as `AGENTS.md` requires:
   - `map.md` for structure changes;
   - `cybersec.md` for security work;
   - `history.md` for significant changes;
   - `plan.md` and `notes.md` as needed.
