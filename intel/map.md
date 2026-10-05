# Repository Map

Last updated: 2026-10-04 (round 6: stored keys, plans 3.3/3.4, golden fixtures, attestations; earlier: Q-003, Q-005, Q-006, Q-009, Q-013). Architecture rules live in [`maint.md`](maint.md).

## Structure

```text
Cryptare/
├── main.go                  # entry: build Cobra root command with a lazy DB opener, execute
├── database_path.go         # key store path: CRYPTARE_DB_PATH or the user data folder; legacy notice; `keys path`
├── database_path_test.go    # default path per OS, private folder, legacy-database notice
├── version_test.go          # --version flag test
├── lazy_database_test.go    # only keys commands open (and create) the DB
├── interrupt_test.go        # exit codes; TestRunMain helper for subprocess tests
├── interrupt_unix_test.go   # SIGINT during a decrypt: exit 130, nothing left behind
├── keys_output_test.go      # keys commands keep stdout clean (subprocess, SEC-017)
├── internal/                # single Go package `internal`
│   ├── crypto.go            # file/dir encryption, legacy (v1) reads, key blobs, key export/import
│   ├── format_v2.go         # version 2 format: header, Argon2id or stored key (HKDF), chunked AES-GCM stream
│   ├── keys.go              # key-store flows shared by CLI and TUI: generate, unlock (StoredKeyCredential), import
│   ├── compress.go          # gzip / tar.gz / zip create + extract; closeWithError helper
│   ├── database.go          # GORM + SQLite, KeyModel, key CRUD, database file trust check
│   ├── fileowner_unix.go    # fileOwner: a file's owner on Unix (build-tagged)
│   ├── fileowner_other.go   # fileOwner stub for other platforms
│   ├── logic-cli.go         # Cobra commands, password/confirm prompts, output-name helpers
│   ├── ui-dashboard.go      # Bubble Tea model types, messages, constructors, Init
│   ├── logic-tui.go         # Bubble Tea Update/View, forms, vim mode, action commands
│   ├── testdata/golden/     # files written by v1.0.1, v1.3.1 and the first stored-key build, with passwords (README.md)
│   └── *_test.go            # unit, CLI and TUI tests; legacy_fixtures_test.go writes v1 layouts; cancel_test.go stops operations partway; fuzz_test.go fuzzes the untrusted-input readers; password_norm_test.go covers password normalisation; golden_test.go opens the golden fixtures; keys_test.go and stored_keys_test.go cover stored keys
├── .github/workflows/       # ci.yml, cd.yml, docker.yml, security.yml
├── .github/dependabot.yml   # weekly gomod, github-actions and docker update PRs
├── scripts/third-party-licenses.sh  # writes THIRD_PARTY_LICENSES.txt for release archives (cd.yml)
├── scripts/licenses/        # notices of the C libraries in the release binaries: musl (Linux), MinGW-w64 runtime (Windows)
├── .devcontainer/           # Ubuntu + Go + Neovim dev container
├── Dockerfile               # CGO build (golang:1.26.8-alpine) → alpine:3.22 runtime, both pinned by digest; runs as UID 10001
├── .dockerignore            # keeps the build context to the sources
├── Makefile                 # release tagging only: tag, push-tag, release
├── intel/                   # engineering docs (this directory)
├── AGENTS.md, README.md, CONTRIBUTING.md, CODE_OF_CONDUCT.md, SECURITY.md (private vulnerability reporting)
├── LICENSE (Apache-2.0), NOTICE (project attribution), CODEOWNERS (@jabbott-iii)
└── .idea/ (tracked), .junie/ (untracked, empty)   # IDE / agent workspace files
```

## Components

| Component | Key symbols | Notes |
|---|---|---|
| Entry | `main`, `newRootCmd`, `databaseOpener`, `version`, `databasePath`, `userDataDir`, `noticeLegacyDatabase`, `newKeysPathCmd` | Passes a lazy opener; the DB is opened only by the `keys` commands, the TUI, `encrypt --key` and `decrypt` of a file encrypted with a stored key. Without `CRYPTARE_DB_PATH` it is `cryptare/cryptare.db` in the user data folder, whose folder the opener creates 0700; a `cryptare.db` in the current folder only gets a notice (Q-003). `keys path` prints the path and opens nothing. `version` defaults to `dev`; release builds set it with `-X main.version=<tag>`. |
| CLI | `NewRootCmd`, `NewRootCmdLazy`, `DatabaseOpener`, `openOnce`, `new*Cmd`, `readPassword`, `readNewPassword`, `confirmAction`, `derive*Output`, `displayText`, `runCancellable`, `InterruptedError`, `checkExportOutput`, `unlockStoredKey`, `silenceUsageOnRun`, `withMissingKeyHint`, `withStoredKeyHint` | `--vim` is a root flag; `encrypt --key`/`-k` encrypts with a stored key, and `decrypt` finds a stored key from a regular file's header (plan 3.4); usage text only follows command-line mistakes; `--password/-p` (prints a warning) and `--password-file` (`passwordFlags`) on crypto and key commands. `--force` on `encrypt`, `decrypt`, `compress` and `decompress` allows overwriting an existing output. `--max-size` and `--max-entries` on `decompress` and `decrypt` set the extraction limits. File commands run under `runCancellable`: a signal cancels the operation, which removes its unfinished output, and the exit status is 128 + the signal (SEC-015). `keys export` takes `--force` and never writes over the key database (BUG-017). `--password ""` counts as given (BUG-023). |
| TUI | `DashboardModel`, `fieldsFor`, `updateForm`, `handleVimFormKey`, `buildActionCmd`, `checkTUINewPassword`, `actionRunner`, `quit` | Forms mirror the CLI operations; actions run as `tea.Cmd`s. Forms that set a password have a "Confirm password" field. The encrypt form's last field takes a stored key ID; export takes the key's master password once, and import has an optional second password field for an older export with a password of its own (either password in either field). The key table shows stored values through `displayText`. Actions run through `actionRunner`; quitting while one runs cancels it and quits once it reports back (SEC-015). |
| Key flows | `GenerateStoredKey`, `StoredKeyCredential`, `ImportStoredKey`, `ErrWrongMasterPassword`, `ErrWrongExportPassword`, `ErrSeparateKeyPassword`, `keyLookupError`, `keySaveError` | Shared by the CLI and the TUI (plan 3.3), over the `Storage` interface. `StoredKeyCredential` unlocks a key to encrypt (policy-checked master password) or decrypt with it (plan 3.4); `ImportStoredKey` checks the key inside an export, asking for a separate key password for older exports (BUG-011). |
| Crypto | `EncryptFile`, `DecryptFile`, `DecryptFileWithLimits`, `encryptSingleFile`, `encryptDirectory`, `writeDirectoryArchive`, `restoreDirectoryArchive`, `decryptBytesWithAAD` (legacy), `GenerateKey`, `EncryptKeyBlob`, `DecryptKeyBlob`, `openKeyData`, `ExportKeyToFile`, `ImportKeyFromFile`, `validateKeyExport`, `validStoredKeyBlob`, `CheckPasswordPolicy`, `normalizePassword`, `passwordCandidates`, `EncryptFileContext`, `DecryptFileWithLimitsContext`, `EncryptFileWithCredentialContext`, `DecryptFileWithCredentialContext`, `EncryptedWithStoredKey`, `defaultEncryptOutput`, `checkOutputOutsideFolder` | Writes only the version 2 format and streams files and folders; legacy (version 1) data is still read, whole. Directory mode reuses `writeTarGz` and `extractTarGz`. `CheckPasswordPolicy` guards every path that sets a new password (`maint.md` §4). New data uses the NFKC-normalised password; reads try the forms `passwordCandidates` lists (Q-009). A folder's default output sits next to it, and an output inside it is refused (BUG-016). |
| Format v2 | `v2Header`, `newV2Header`, `parseV2Header`, `readV2Header`, `useStoredKey`, `argon2Params`, `passwordKDF`, `kdfArgon2id`, `kdfArgon2idNFKC`, `kdfHKDFStoredKey`, `Credential`, `PasswordCredential`, `passwordAEAD`, `storedKeyAEAD`, `encryptingWriter`, `decryptingReader`, `sealV2`, `openV2`, `errDecrypt`, `errUnsupportedFormat`, `ErrStoredKeyRequired`, `ErrPasswordRequired`, `ErrWrongStoredKey` | 46-byte header (54 with a stored key's ID), Argon2id key or HKDF-SHA256 from a stored key, 64 KiB AES-256-GCM chunks (STREAM). `parseV2Header` refuses unknown values and Argon2id settings above its limits before deriving a key. KDF `2` marks a key from the normalised password (Q-009). `newDecryptingReader` authenticates the first chunk with each candidate password. Layout in `maint.md` §3. |
| Compression | `CompressFileWithFormat`, `DecompressFile`, `DecompressFileWithLimits`, `ExtractLimits`, `writeTarGz`, `writeZip`, `extractTarGz`, `extractZip`, `extractToDir`, `walkSourceTree`, `CheckOutputPath`, `writeFileAtomic`, `CompressFileWithFormatContext`, `DecompressFileWithLimitsContext`, `copyContext`, `gzipHeaderName`, `GunzipFileContext`, `checkCompressLevel`, `tempNamePart`, `readAllLimit` | Rejects symlinks, special files and `..` traversal. Folders are read through an `os.Root` (`walkSourceTree`). `CheckOutputPath` enforces the output-safety rules (`maint.md` §4). Extraction runs under `ExtractLimits` into a temporary folder (`extractToDir`), writes through an `os.Root`, and sets owner-only permissions (`extractDirMode`, `extractFileMode`). The `…Context` variants check for cancellation between reads and entries. Gzip header names are Latin-1 only (`gzipHeaderName`, BUG-015). `decompress --raw` gunzips without extracting (BUG-022); contradictory `--format` and out-of-range `--level` are refused (BUG-020). |
| Storage | `NewDatabase`, `prepareDatabaseFile`, `checkDatabaseFileTrust`, `fileOwner`, `ErrUntrustedDatabase`, `withSecureDelete`, `KeyModel`, `SaveKey`, `ListKeys`, `GetKey`, `DeleteKey`, `databaseFilePath`, `ErrKeyExists`, `ErrUnsupportedDatabasePath` | `DeleteKey` uses raw SQL `DELETE … RETURNING` (a hard delete). Connections open with SQLite `secure_delete` on, so deleted rows are overwritten. The file is created 0600, and an existing one is tightened to 0600. On Unix, a database or side file another user owns or others can write is refused (SEC-016). GORM's logger is silent and errors are translated (SEC-017). `databaseFilePath` finds the file behind a plain path or `file:` URI; a plain path with `?` is refused (SEC-010). |

## Dependencies

| Module | Version | Used for |
|---|---|---|
| `github.com/spf13/cobra` | v1.10.2 | CLI |
| `github.com/charmbracelet/bubbletea` / `lipgloss` | v1.3.10 / v1.1.0 | TUI |
| `github.com/charmbracelet/x/term` | v0.2.2 | Hidden password input at the CLI prompt |
| `gorm.io/gorm` + `gorm.io/driver/sqlite` | v1.31.2 / v1.6.0 | Storage, via `github.com/mattn/go-sqlite3` v1.14.52 (**CGO**) |
| `golang.org/x/crypto` | v0.57.0 | `argon2` only (the legacy PBKDF2 key uses the standard library's `crypto/pbkdf2`) |
| `golang.org/x/text` | v0.42.0 | `unicode/norm`: NFKC normalisation of passwords (Q-009); also used by GORM |

## Component dependencies

```mermaid
flowchart TD
  main["main.go<br/>databasePath"] -->|"opened only by keys, the TUI and stored-key encrypt/decrypt"| dbfile[("SQLite file<br/>CRYPTARE_DB_PATH or<br/>user data folder/cryptare/cryptare.db")]
  main --> root["NewRootCmd<br/>logic-cli.go"]
  root -->|no subcommand| tui["TUI<br/>ui-dashboard.go + logic-tui.go"]
  root --> cli["encrypt · decrypt · compress · decompress · keys"]
  cli --> crypto["crypto.go"]
  cli --> compress["compress.go"]
  cli --> keys["keys.go<br/>shared key flows"]
  cli --> database["database.go"]
  tui --> crypto
  tui --> compress
  tui --> keys
  tui --> database
  keys -->|"Storage interface"| database
  keys --> crypto
  crypto -->|"writeTarGz / extractTarGz"| compress
  crypto --> format["format_v2.go"]
  database --> dbfile
```

## Encryption data flow

New data (version 2):

```mermaid
flowchart LR
  pw["password"] --> nfkc["normalizePassword<br/>drop leading BOM, NFKC"] --> kdf["Argon2id<br/>64 MiB, 3 passes, 4 lanes<br/>random 16-byte salt"] --> key["256-bit key"]
  sk["stored key<br/>(unlocked with its master password)"] --> hkdf["HKDF-SHA256<br/>random 16-byte salt<br/>header names the key"] --> key
  file["file"] --> chunks
  dir["directory"] --> tgz["tar.gz stream<br/>(writeDirectoryArchive)"] --> chunks
  chunks["64 KiB chunks"] --> gcm["AES-256-GCM per chunk<br/>nonce = prefix ‖ counter ‖ last flag<br/>AAD = header"]
  key --> gcm --> out["header ‖ sealed chunks → *.enc"]
```

Legacy data (version 1) is still read: `salt ‖ nonce ‖ ciphertext` for files, and the
same after the `CRYPTARE-DIR-ENC\x00` magic (used as AAD) for folders, with a PBKDF2-
HMAC-SHA256 key (100,000 iterations). Stored keys and key exports use the same two
formats with their own content types (`maint.md` §3).

Decryption reverses the flow. The CLI and the TUI first read a regular file's header
(`EncryptedWithStoredKey`): if it names a stored key, they unlock that key instead of
asking for a password. `DecryptFileWithCredentialContext` then peeks at the first bytes:
- version 2: it reads the header, checks its limits and content type, then decrypts
  chunk by chunk. A folder's tar.gz is restored with `extractTarGz` in a temporary
  folder, and a file goes to a temporary file. Either is kept only after the final
  chunk authenticates.
- legacy: it reads the whole file. The directory magic means a folder; anything else
  is one plaintext file.

## Key management flow

```mermaid
flowchart LR
  gen["keys generate"] --> rk["GenerateKey<br/>32 random bytes"] --> eb["EncryptKeyBlob<br/>(master password, content type 3)"] --> row[("key_models row")]
  row --> exp["keys export<br/>(master password checked)"] --> env["JSON KeyExport v1<br/>(holds encrypted blob)"] --> eb2["sealV2<br/>(the key's master password, content type 4)"] --> ckey["*.ckey"]
  ckey --> imp["keys import"] --> dec["openKeyData → JSON<br/>validateKeyExport<br/>inner key checked"] --> row
  row --> use["encrypt --key / decrypt<br/>StoredKeyCredential"] --> enc2["key source 2 file"]
```

Stored keys encrypt files since round 6 (plan 3.4, Q-014): the file's header names the
key, and each file's key is derived from it with HKDF.

## CI/CD

All third-party actions are pinned to full commit SHAs.

| Workflow | Trigger | What it does |
|---|---|---|
| `ci.yml` | push/PR (all branches) | Ubuntu, Windows and macOS matrix: fails if a key database or `.ckey` is tracked (SEC-003), `go mod tidy` diff, `go vet`, golangci-lint v2.13.2, `go test` with coverage (with `-race` on Ubuntu and macOS) (Codecov). Then a native CGO build, smoke-tested with `--version`, `keys generate`/`keys list` and an encrypt/decrypt round-trip. |
| `security.yml` | push/PR, weekly | CodeQL (Go, security-extended). gosec v2.29.0, with SARIF uploaded to Code Scanning (category `gosec`). govulncheck v1.8.0 on the `go.mod` toolchain, failing on reachable vulnerabilities (SEC-018). |
| `docker.yml` | push/PR to `main` | `docker build`, `--help`, a check that the image runs as UID 10001 (SEC-013), then `keys generate`/`keys list` on a named volume |
| `cd.yml` | `v*` tags, manual | Native CGO build per target on a matching runner: linux amd64/arm64 (static against musl, inside the pinned `golang:1.26.8-alpine` image), darwin arm64, darwin amd64 (cross-compiled), windows amd64. Stamps `-X main.version=<tag>`, smoke-tests every target except darwin/amd64 (and checks the Windows binary imports no toolchain DLL), packages `cryptare_<os>_<arch>` archives (binary, `LICENSE`, `NOTICE` and `THIRD_PARTY_LICENSES.txt` from `scripts/third-party-licenses.sh`) plus `checksums.txt`, and on tags attaches signed build provenance to each published file (`actions/attest`, SEC-020) and creates a GitHub Release. Then the `container` job builds and smoke-tests the Docker image and, on tags, pushes it to GitHub Packages (`ghcr.io/<owner>/cryptare`: `X.Y.Z`, and `X.Y`/`latest` for the newest release) with an attestation on its digest. |
