# Architecture and Maintainability Guide

This is the authoritative source for Cryptare's architecture and maintainability
rules (see `AGENTS.md`). `CONTRIBUTING.md` must stay consistent with it.

Last reviewed: 2026-10-05 (round 6: stored keys, golden fixtures; docs round)

## 1. System overview

Cryptare is a single-binary Go CLI/TUI for local file protection:

- password-based encryption and decryption of files and directories (AES-256-GCM);
- compression and extraction (gzip, tar.gz, zip);
- a local key store (SQLite) for randomly generated, password-protected keys, which can
  encrypt files in place of a password (plan 3.4).

It has no network surface: there is no HTTP server and nothing calls a remote service.

- Module: `github.com/jabbott-iii/Cryptare`; `go.mod` declares `go 1.26.0` and
  `toolchain go1.26.8` (SEC-018).
- Entry point: `main.go` opens the database, builds the Cobra root command and runs it.
- All application logic lives in one flat Go package, `../pkg`.

## 2. Layers and responsibilities

| Layer | Files | Responsibility | May depend on |
|---|---|---|---|
| Entry | `main.go`, `database_path.go` | Work out the DB path (`CRYPTARE_DB_PATH`, default `cryptare/cryptare.db` in the user data folder), build the root command with a lazy database opener and `keys path`, run it | `internal` |
| Interfaces | `../pkg` (Cobra); `../pkg` + `internal/logic-tui.go` (Bubble Tea) | Parse input, prompt, call core/storage, render results | core, storage |
| Core operations | `internal/crypto.go`, `internal/format_v2.go`, `internal/compress.go`, `internal/keys.go` | Encryption formats, key blobs, key export/import, archive creation/extraction; `keys.go` holds the key-store flows both interfaces share (plan 3.3) | stdlib, `golang.org/x/crypto`, `golang.org/x/text`; `keys.go` uses storage only through the `Storage` interface |
| Storage | `../pkg` | GORM/SQLite schema (`KeyModel`) and key CRUD | GORM, SQLite driver |

Rules:

1. **Core stays UI-agnostic.** Functions in `crypto.go` and `compress.go` take paths,
   passwords or bytes and return `error`. They do not print, prompt, or import Cobra or
   Bubble Tea.
2. **The CLI and TUI must match.** Both call the same core functions: `GenerateStoredKey`,
   `StoredKeyCredential` and `ImportStoredKey` in `keys.go`, `ExportKeyToFile`, and the
   `…WithCredentialContext` file operations (plan 3.3). What each interface still does
   itself (prompting, output-path checks, the TUI's fields and confirmation) must stay
   equivalent, and changes to it are tested in both.
3. **Data-protecting validation goes in core.** Checks such as the password policy,
   overwrite protection and extraction limits belong in the core layer so both
   interfaces get them. For example, `ErrEmptyPassword` is enforced in `crypto.go`, not
   in the CLI or TUI (SEC-001: the TUI once accepted empty passwords because this rule
   was missed).
4. **One persistence type.** `*Database` is the only persistence type. The shared key
   flows in `keys.go` take it as a `Storage`, the interface it implements, so core code
   doesn't depend on GORM.
5. **Only commands that use stored keys open the database.** `main.go` passes a
   `DatabaseOpener` to `NewRootCmdLazy`. Only the `keys` commands, the TUI, `encrypt
   --key`, and `decrypt` of a regular file whose header names a stored key
   (`EncryptedWithStoredKey`) call it, at most once per run (SEC-010). Don't add other
   dependencies on the DB from file commands. A stored key credential carries whether
   it was unlocked for new data (`canEncrypt`), and `EncryptFileWithCredentialContext`
   refuses one that wasn't, so the policy on its master password is enforced in core. `NewDatabase` creates the file with mode 0600 and tightens an
   existing one. On Unix it refuses a database or SQLite side file that another user
   owns or that group or others can write (`ErrUntrustedDatabase`, SEC-016). A plain
   path containing `?` is refused (`ErrUnsupportedDatabasePath`); a `file:` URI is
   prepared like a path (SEC-010).
6. **The default key store is per user** (Q-003). Without `CRYPTARE_DB_PATH`,
   `databasePath` uses `cryptare/cryptare.db` in `userDataDir`: `%LocalAppData%` on
   Windows, `~/Library/Application Support` on macOS, and `$XDG_DATA_HOME` (absolute
   only) or `~/.local/share` elsewhere. A data folder that isn't an absolute path is an
   error that names `CRYPTARE_DB_PATH`; never fall back to the current folder. The
   opener creates the folder 0700 and is still the only thing that creates anything.
   A `cryptare.db` left in the current folder by an earlier version only gets a notice
   on stderr (`noticeLegacyDatabase`); it is never opened, moved or copied, because it
   may not be the user's (SEC-016). `keys path` is defined in package `main`, because
   it needs `databasePath`, and opens nothing.

## 3. On-disk formats (compatibility contract)

Existing artifacts must stay readable. Everything new is written in the version 2 format
below (plan 3.1/3.2). The version 1 (legacy) layouts are read-only and stay readable
with no time limit; there is no migration command (owner decision, 2026-09-27). Any
further change needs a new format version or content type, a read path for everything
already written, and tests covering both.

**Version 2** (`internal/format_v2.go`). Every artifact starts with a 46-byte header:

| Offset | Size | Field |
|---|---|---|
| 0 | 9 | magic `"CRYPTARE\x00"` (a legacy folder artifact has `-` at offset 8) |
| 9 | 1 | format version, `2` |
| 10 | 1 | content type: `1` file, `2` folder (tar.gz), `3` stored key, `4` key export |
| 11 | 1 | key source: `1` password, `2` stored key (plan 3.4, Q-014) |
| 12 | 1 | KDF: `1` Argon2id over the password, `2` Argon2id over its normalised form (Q-009), `3` HKDF-SHA256 from a stored key (key source 2 only) |
| 13 | 4 | Argon2id memory in KiB, big-endian (default 65,536 = 64 MiB; `0` with a stored key) |
| 17 | 4 | Argon2id passes, big-endian (default 3; `0` with a stored key) |
| 21 | 1 | Argon2id lanes (default 4; `0` with a stored key) |
| 22 | 16 | random salt |
| 38 | 1 | chunk size as a power of two (default 16 = 64 KiB) |
| 39 | 7 | random nonce prefix |
| 46 | 8 | the stored key's ID, its 16 hexadecimal characters decoded (key source 2 only, which makes the header 54 bytes) |

- **Key:** Argon2id(password, salt, the header's settings), 32 bytes, for AES-256-GCM.
  With a stored key (key source 2): HKDF-SHA256(stored key, salt, info
  `"cryptare v2 stored-key data key"`), 32 bytes, so each file gets a key of its own.
  Only files and folders use key source 2; stored keys and key exports are always
  protected with a password. Readers find the stored key from the header's key ID
  (`EncryptedWithStoredKey`, which reads regular files only, so a pipe isn't consumed)
  and refuse the wrong kind of credential before reading the body
  (`ErrStoredKeyRequired`, `ErrPasswordRequired`, `ErrWrongStoredKey`). Builds up to
  v1.3.1 refuse key source 2 ("unsupported encrypted data: key source 2").
- **Password bytes (Q-009):** new data uses the password without a leading byte order
  mark, in Unicode NFKC form (`normalizePassword`). The header records KDF `2` when
  that differs from the password as given, and `1` otherwise, so data protected with
  any other password is exactly what earlier versions write and read. Readers try, for
  KDF `1` and for the legacy formats, the password without a leading byte order mark,
  then as given if it had one, then its NFKC form (`passwordCandidates`, at most three
  key derivations); for KDF `2`, the NFKC form only. Builds up to v1.2.0 refuse KDF `2`
  as unknown ("unsupported encrypted data: key derivation 2"). The first chunk is
  authenticated when the reader is created (`newDecryptingReader`), so a wrong password
  is reported before any output exists.
- **Normalisation is part of the format.** KDF `2` data opens only through the NFKC
  form, so the normalised form of a password must never change between versions.
  `TestNormalizePasswordKnownAnswers` pins it (checked against Python's
  `unicodedata`). If an update of `golang.org/x/text`, or a Go toolchain that selects
  newer Unicode tables (`x/text` switches to Unicode 17 from Go 1.27), makes it fail,
  treat it as a format change: work out which passwords change, then either keep the
  old form readable (another candidate or KDF identifier) or document the break in the
  release notes. Don't just update the expected values. `x/text` v0.42.0 was such a
  change: it fixed composition bugs in v0.41.0, the version v1.3.0 was built with
  (`notes.md` §3).
- **Body:** the plaintext in chunks of the chunk size, each sealed with its own 16-byte
  tag. The last chunk may be shorter or empty; data that ends on a chunk boundary ends
  with a full last chunk. A chunk's nonce is the prefix, a 4-byte big-endian chunk
  counter and a byte that is `1` for the last chunk and `0` otherwise. Its additional
  data is the whole header. So reordering, dropping, truncating or appending chunks,
  or changing any header byte (including the content type), fails authentication.
- **Limits on reading:** a header asking for more than 1 GiB of memory, 1–10 passes,
  1–16 lanes (and at least 8 KiB per lane) or a chunk size outside 2^10–2^24 is refused
  before any key derivation (`errUnsupportedFormat`), as are unknown versions, content
  types, key sources and KDFs, and combinations this version doesn't write: KDF 3 with
  a password, KDF 1 or 2 with a stored key, Argon2id settings with a stored key, a
  stored key or key export under key source 2, and a key source 2 header cut short.
- **Output:** a file or folder is released only after its last chunk authenticates:
  decrypted files go through a temporary file and folders through `extractToDir`.

| Artifact | Version 2 (written) | Version 1 (legacy, read-only) |
|---|---|---|
| Encrypted file (`.enc`) | header (type 1) ‖ chunks of the file | `salt(16) ‖ nonce(12) ‖ AES-256-GCM ciphertext+tag`. Key = PBKDF2-HMAC-SHA256(password, salt, 100,000 iterations, 32 bytes). No header, version or AAD. |
| Encrypted file or directory with a stored key (`.enc`) | 54-byte header (type 1 or 2, key source 2) ‖ chunks, as above, under the HKDF key | — (new in round 6) |
| Encrypted directory (`.enc`) | header (type 2) ‖ chunks of a tar.gz of the tree, streamed as it is built | `"CRYPTARE-DIR-ENC\x00" ‖ salt ‖ nonce ‖ ciphertext+tag` with the PBKDF2 key. Plaintext is a tar.gz of the tree. The magic string is also the GCM AAD, so a stripped magic can't be decrypted as a single file. |
| Stored key (`key_models.encrypted_blob`) | base64(header (type 3) ‖ one chunk holding the 32-byte key): 94 bytes decoded | base64(`salt ‖ nonce ‖ GCM(32-byte raw key)`): 76 bytes, PBKDF2 key from the master password. Existing rows keep this format. |
| Key export (`.ckey`) | base64(header (type 4) ‖ chunks of the JSON `KeyExport{version:1,…}`) | base64(`salt ‖ nonce ‖ GCM(JSON KeyExport{version:1,…})`), PBKDF2 key. |
| Compressed output | `.gz` (single file, gzip header `Name` set), `.tar.gz` (directory), `.zip` | (unversioned; unchanged) |

- The export's JSON carries the stored blob as it is in the database (either format).
  The outer layer is protected by the key's own master password (BUG-011, Q-015):
  `ExportKeyToFile` checks that the password unlocks the key and meets the password
  policy. Exports from v1.3.1 and earlier may use a separate password; `ImportStoredKey`
  takes both (`otherPasswords`, from `--export-password-file`, a second prompt at a
  terminal, or the TUI's second field), tries each on the export and on the key inside
  so their order doesn't matter, reports `ErrSeparateKeyPassword` when only the
  export's password was given, and checks the key inside every export before storing
  it. Import accepts only version 1, a
  16-character lower-case hex key ID, `AES-256-GCM`, and a stored blob that is either
  76 bytes (legacy) or a 94-byte version 2 stored key whose header passes the checks
  above (`validateKeyExport`, `validStoredKeyBlob`; SEC-011).
- Readers tell the versions apart by the magic: data starting with `"CRYPTARE\x00"` is
  version 2, anything else is read as version 1. A legacy file whose random salt starts
  with those 9 bytes (probability 2^-72) would be misread.
- Compressed output archive entries use forward-slash relative paths. Symlinks and
  special files are rejected.
- Builds from before this change, v1.1.0 included, can't read version 2 data: they take
  it for a legacy file and report the usual "wrong password or corrupted" error.

Tests that guard this contract; keep them passing:
- golden fixtures (`golden_test.go`, `testdata/golden/`): files, folders, stored keys and
  exports written by v1.0.1 and v1.3.1, and stored-key files written by round 6's
  build. They catch a change made to the writer and the reader alike, which the
  round-trip tests can't. Never regenerate or edit a fixture to make a test pass; a new
  format version adds new fixtures;
- stored keys: `TestStoredKeyEncryptDecrypt`, `TestStoredKeyHeaderIsAuthenticated` and
  `TestStoredKeyHeaderChecks` (`keys_test.go`);
- legacy reads: `TestDecryptLegacyEncryptedFile`, `TestDecryptAcceptsLegacyEmptyPassword`,
  `TestDecryptAcceptsLegacyShortPassword` and `TestLegacyShortPasswordCmds`
  (`legacy_fixtures_test.go` holds the test-only writers for the legacy layouts);
- version 2: `TestEncryptWritesVersion2Format`, `TestDefaultPasswordKDFIsWritten`,
  `TestStreamRoundTripSizes`, `TestStreamRejectsTampering`,
  `TestDecryptRejectsUnsupportedHeaders` and `TestDecryptRejectsWrongContentType`;
- password forms: `TestNormalizedPasswordHeader`, `TestEarlierDataOpensWithPasswordForms`
  and `TestUnicodePasswordFormsOpenTheSameData` (`password_norm_test.go`);
- the legacy KDF: `TestDeriveKey` (PBKDF2 known answers).

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
- Files the tool writes use mode `0o600`; directories it creates, including missing
  parent folders of an output, use `0o700`. On Windows these bits only set or clear the
  read-only attribute, so outputs and the key database inherit the folder's access
  control list; the README says so, and owner-only ACLs are not set (Q-012, owner
  decision 2026-10-04: document only; SEC-019).
- Values read from the key database are shown through `displayText`, so stored control
  characters print escaped instead of reaching the terminal (SEC-016).
- A gosec finding accepted by design is annotated on its line as
  `// #nosec <rule> -- <reason>`, naming only that rule, and recorded in
  `cybersec.md` (SEC-012). Fix a finding rather than suppress it when a fix exists.
- **Output safety.** Core operations refuse an output that is their own input
  (`ErrSameInputOutput`), and compressing or encrypting a folder refuses an output
  inside that folder (`ErrOutputInsideInput`; `checkOutputOutsideFolder`, BUG-016).
  A folder's default encrypted output sits next to it (`defaultEncryptOutput`).
  Interfaces call `CheckOutputPath` before writing: the CLI overwrites an existing
  output only with `--force`, and the TUI never does. `keys export` uses
  `checkExportOutput`, which also never writes over the key database
  (`ErrOutputIsKeyDatabase`, BUG-017). New commands that write files must follow the
  same rules.
- **Atomic single-file writes.** Write single-file outputs with `writeFileAtomic` or
  `createAtomicFile` + `Commit` (in `compress.go`). They write a hidden temporary file
  next to the destination and then rename it, so a failed run leaves the destination
  untouched. Consequences:
  - a replaced output becomes a new 0600 file;
  - a symlink at the output path is replaced, not written through;
  - only a process that can't run its clean-up (power loss, `kill -9`) can leave a
    hidden `.<name>.*.tmp` file (mode 0600) behind; see Cancellation below.

  Archive extraction is atomic too: `extractToDir` extracts into a new hidden
  `.<name>.*.tmp` folder (mode 0700) and renames it into place only on success. An
  existing output (allowed with `--force`) is moved aside, replaced, then removed,
  so it is replaced as a whole rather than merged into. Replacing a folder that holds
  the archive being extracted is refused (`ErrInputInsideOutput`).
- **Cancellation (SEC-015).** The long-running file operations have `…Context`
  variants (`EncryptFileContext`, `DecryptFileWithLimitsContext`,
  `CompressFileWithFormatContext`, `DecompressFileWithLimitsContext`); the plain
  functions wrap them with `context.Background()`. They check the context before each
  read (`copyContext`), at each archive entry (`extractBudget.addEntry`, the source
  walk) and before the final rename, so cancelling ends them through the usual
  `Abort`/`RemoveAll` clean-up. The CLI runs them under `runCancellable`: SIGINT,
  SIGTERM and SIGHUP cancel the context, and the command exits with 128 plus the
  signal's number (`InterruptedError`); a second signal ends the process at once. The
  TUI runs actions through `actionRunner`: quitting while one runs cancels it and quits
  when it reports back, and the launcher waits for it however the program ends. New
  operations that read or write data in a loop must take a context the same way. The
  hidden password prompt restores the terminal on SIGINT, SIGTERM, SIGQUIT and SIGHUP
  (BUG-018).
- **Gzip header names.** A gzip header holds Latin-1 only, so names go through
  `gzipHeaderName` (a folder's falls back to `archive.tar`, a file's to none;
  BUG-015). Readers must not depend on the stored name.
- **Extraction permissions and confinement.** Entries are written through an
  `os.Root` opened on the new extraction folder, so nothing can be created outside
  it. Permissions stored in an archive are ignored: folders get `extractDirMode`
  (0700) and files `extractFileMode` (0600, or 0700 when marked executable).
- **Extraction limits.** Every decompression or extraction runs under
  `ExtractLimits` (default 10 GiB of output and 100,000 entries, SEC-007) and copies
  through `extractBudget`, never a bare `io.Copy` from a decompressor. New code that
  extracts or decompresses must do the same. The CLI exposes the limits as
  `--max-size` and `--max-entries`; the TUI uses the defaults. Input that is read whole
  is bounded too (BUG-024): a legacy-format file by `MaxBytes`, checked before reading,
  and a key export by `maxKeyExportSize` (1 MiB), through `readAllLimit`.
- **Compression options.** `resolveCompressFormat` refuses an explicit format that
  contradicts the output's extension (`ErrFormatMismatch`), and the CLI and TUI refuse a
  level other than 1–9 or -1 with `checkCompressLevel` (BUG-020). `decompress --raw`
  (`GunzipFileContext`) gunzips without extracting a tarball (BUG-022); it is a CLI
  option only, like the extraction-limit flags.
- **Temporary names.** Temporary files and folders next to an output use
  `tempNamePart`, at most 64 bytes of the output's name (BUG-021).
- **Containment checks** (`pathWithin`) compare paths as text and then folders by
  identity (`os.SameFile`), so symlinked or differently cased spellings are caught
  (BUG-019).
- **Password policy.** Any operation that protects new data with a password
  (`EncryptFile`, `EncryptKeyBlob`, `ExportKeyToFile`, and `StoredKeyCredential` for
  encrypting, which checks the key's master password) calls `CheckPasswordPolicy`: at least `MinPasswordLength` (15) Unicode code points, not
  one repeated character, and no composition rules or maximum length, counted on the
  normalised password that actually protects the data (Q-009). Operations that
  read existing data (decrypt, `DecryptKeyBlob`, import) must not check it, so older
  files and keys stay readable. Interfaces also confirm a typed new password: the CLI
  with `readNewPassword` (terminal input only) and the TUI with a "Confirm password"
  field and `checkTUINewPassword`. New commands or forms that set a password must do
  the same. A key's existing master password (`encrypt --key`, `keys export`) is checked
  against the key instead of confirmed.
- **Password input.** Commands that take a password register it with
  `passwordFlags.register` and read it with `passwordFlags.get`, which handles
  `--password-file` and the `--password` warning. They fall back to `readPassword`, or
  to `readNewPassword` when the password protects new data (SEC-004). A stored key's
  master password goes through `unlockStoredKey`, which prompts with the key's ID.
- **Usage text.** `NewRootCmdLazy` wraps every command's `RunE` with
  `silenceUsageOnRun`, so the usage text follows only command-line mistakes that Cobra
  reports before a command runs. A command added outside it (such as `keys path` in
  `main`) sets `SilenceUsage` itself.
- Symlinks and non-regular files are rejected when reading directory trees, which
  are always read through `walkSourceTree` (an `os.Root` on the source folder,
  SEC-014). Path traversal (`..`) is rejected when extracting, before the `os.Root`
  check. Keep these
  checks. An absolute entry name is extracted inside the output folder.
- Code is grouped with the existing `//----- section -----//` banner comments.
- File naming is mixed (`logic-cli.go` vs `logic_cli_test.go`). Don't rename existing
  files as part of unrelated changes. Name new files in `snake_case.go` (Go convention).
- Keep dependencies minimal and prefer the standard library. Current direct
  dependencies:
  - Cobra (CLI)
  - Bubble Tea and Lip Gloss (TUI)
  - GORM with `gorm.io/driver/sqlite`, which uses `github.com/mattn/go-sqlite3` (CGO)
  - `golang.org/x/crypto`, used for `argon2` (the version 2 format). The legacy PBKDF2
    key uses the standard library's `crypto/pbkdf2` (since plan 3.1).
  - `golang.org/x/text`, used for `unicode/norm` (password normalisation, Q-009). It
    was already linked in through GORM; it is direct since Q-009.

  `github.com/charmbracelet/x/term` (direct since 2026-09-24) provides `ReadPassword` and
  `IsTerminal` for the CLI's hidden password prompt (`readPassword` in `logic-cli.go`).

## 5. Testing

- Tests sit beside the code (`../pkg`, `database_path_test.go`). They use
  `t.TempDir()` and must not write anywhere else.
- Database tests use `newTestDatabase(t, useFile)`. Close the SQL handles so Windows
  CI can delete the temporary files.
- Fuzz tests (`../pkg`) cover the code that reads untrusted input:
  the version 2 header and decrypting reader, key exports, `parseSize`, and tar and zip
  extraction (plan 5.21). `go test` runs their seed corpora; fuzz one with
  `go test -run '^$' -fuzz FuzzParseV2Header -fuzztime 1m ./internal`. The decryption
  fuzzers skip headers that ask for an expensive Argon2id setting, and
  `FuzzDecryptingReader` tries each input with a password and with a stored key. For
  the extraction fuzzers, add `-fuzzminimizetime 3s`: by default the fuzzer spends up to
  60 s minimising each new input, which looks like a hang (`notes.md` §3).
- Golden fixtures (`golden_test.go`, `testdata/golden/`, with its `README.md`) are files
  written by released binaries, with their passwords; §3 says how to treat them. Their
  `.gitattributes` keeps Windows checkouts from converting their line endings.
- Subprocess tests run cryptare through the `TestRunMain` helper (`interrupt_test.go`),
  for behaviour an in-process test can't see: exit statuses after a signal, and
  stdout written by GORM's logger (`keys_output_test.go`, SEC-017).
- Cancellation tests stop operations partway through with a context that reports
  itself cancelled after a set number of checks (`cancelAfter` in
  `internal/cancel_test.go`), not with timing. `interrupt_unix_test.go` runs cryptare
  as a subprocess (through the `TestRunMain` helper) and signals it mid-decrypt, with
  the input fed through a named pipe.
- TUI tests drive `DashboardModel.Update` and `updateForm` with synthetic
  `tea.KeyMsg` values; follow `../pkg`.
- CLI tests run `NewRootCmd(db)` with `SetArgs`, `SetIn` and `SetOut`; follow
  `../pkg`.
- The `internal` tests write new data with a cheap Argon2id setting (64 KiB, 1 pass,
  1 lane), set in `TestMain` (`legacy_fixtures_test.go`), so the suite stays fast. The
  setting is recorded in each header, so reading is unaffected;
  `TestDefaultPasswordKDFIsWritten` checks the real default.
- Every security fix needs a regression test that fails before the fix. Record it in
  `intel/cybersec.md`.
- Baseline on 2026-10-04 (Go 1.26.8, linux/amd64): `go test -race ./...` passes with
  89.2% (`main`) and 80.9% (`internal`) statement coverage (`make cover`). The least
  covered code is the TUI's `View`, the terminal-only password reading, and a few error
  paths (`notes.md` §1).

## 6. Build and release

- **CGO is required.** The SQLite driver (`gorm.io/driver/sqlite`, which uses
  `github.com/mattn/go-sqlite3`) needs CGO. A `CGO_ENABLED=0` build compiles but can't
  open the database (BUG-001, fixed in v1.0.1). Since the database is opened lazily
  (plan 2.5), such a build still runs `--help` and the file commands and fails only on
  what uses the key store (`keys` commands, the TUI, `encrypt --key` and decrypting a
  stored-key file), so the smoke tests' `keys generate`/`keys list` steps are what
  catch it. Releases therefore build each target natively with
  CGO (Q-001):
  - Linux binaries are statically linked against musl (tags
    `sqlite_omit_load_extension,osusergo,netgo`; `-linkmode external -extldflags -static`),
    built inside the `Dockerfile`'s pinned `golang:1.26.8-alpine` image
    (`LINUX_BUILDER` in `cd.yml`) on a runner of the target's architecture. Up to v1.3.0
    they were linked against the Ubuntu runner's glibc (Q-013).
  - The Windows smoke test fails if the binary imports a DLL from the MinGW-w64
    toolchain (`libwinpthread`, `libgcc_s`, …): such a DLL would be missing on users'
    machines, and the runtime's notices assume it is linked in.
  - darwin/amd64 is cross-compiled on an Apple silicon runner.
  - CI and CD smoke-run each built binary against a real database, so a CGO-less build
    fails the pipeline instead of shipping. Keep those smoke steps.
- **Releases** are cut by pushing a `vX.Y.Z` tag (`make release VERSION=vX.Y.Z`), which
  triggers `.github/workflows/cd.yml`. It publishes
  `cryptare_<os>_<arch>.tar.gz`/`.zip` archives for linux/amd64, linux/arm64,
  darwin/amd64, darwin/arm64 and windows/amd64, plus `checksums.txt`. On a tag, the
  release job also attaches signed build provenance (`actions/attest`) to each of those
  files (SEC-020), which users check with `gh attestation verify <file> --repo
  jabbott-iii/Cryptare`; that job alone has `id-token: write` and `attestations: write`.
  After it, the `container` job builds the `Dockerfile` (stamped with the tag, with OCI
  labels whose `source` links the package to the repository), smoke-tests the image
  (`--version`, UID 10001, `keys generate`/`keys list`), and pushes it to GitHub
  Packages as `ghcr.io/<owner>/cryptare` with the tags `X.Y.Z`, plus `X.Y` and `latest`
  when this is the newest release of its line and overall (checked against the
  repository's tags, so an older patch release doesn't move them back; pre-releases get
  only their own tag). It attests the image digest and pushes the attestation to the
  registry (`create-storage-record: false`: storage records are for organisation
  repositories). It uses plain `docker` commands, not third-party actions, and only it
  has `packages: write`. A manual CD run builds and smoke-tests the image without
  pushing.
- **Licences in releases (Q-006):** each archive holds the binary, `LICENSE`, `NOTICE`
  (the project's own attribution only) and `THIRD_PARTY_LICENSES.txt`. The release job
  generates the last with `scripts/third-party-licenses.sh`, which lists the modules
  that `go list -deps` reports for each release target (keep its target list in step
  with the `cd.yml` matrix) and copies each module's licence files and the Go standard
  library's `LICENSE`. A module with no licence file stops the release until it is
  added to the script's `stated_licence` list, so every new dependency gets its
  licence checked. C code: SQLite (public domain) is noted with go-sqlite3, and the C
  libraries the release toolchains link in are listed per OS in the script's
  `c_libraries`, with their notices under `scripts/licenses/` (Q-013): musl for Linux,
  the MinGW-w64 runtime for Windows. libgcc's licence exception needs no notice, and
  macOS binaries only link Apple's system libraries. Update the list and the files when
  a toolchain changes.
- **Version stamping:** `main.version` defaults to `dev`. Release builds set it with
  `-ldflags "-X main.version=<tag>"` (as `cd.yml` does), and `cryptare --version` (or
  `-v`) prints `cryptare version <value>`. Keep the variable's name and package stable,
  because the release workflow depends on it.
- **Docker:** the `Dockerfile` builds with CGO for `linux/amd64` only (`GOARCH=amd64`),
  from images pinned by digest, and the runtime runs as user 10001, which owns
  `/app/data` (SEC-013). It adds no runtime packages. `.dockerignore` keeps the build
  context to the sources. When the toolchain line in `go.mod` changes, update the
  builder's tag and digest with it.
- **Go toolchain (SEC-018):** CI and CD install the `toolchain` version from `go.mod`
  (`setup-go` with `go-version-file`), and the `Dockerfile` builder image names the
  same release (`golang:1.26.8-alpine`), as does `cd.yml`'s `LINUX_BUILDER`, which
  pins the same image by digest. Change all three together; the Linux release build
  runs with `GOTOOLCHAIN=local`, so a mismatch fails it. Dependabot proposes
  module, action and Docker base image updates weekly (golang patch versions only; apply
  a golang image update to all three places together), and the Security workflow's govulncheck job fails
  when Cryptare's code reaches a known vulnerability.

## 7. Change checklist

Before opening a pull request (`make check-all` runs the automated steps 1–4, plus gosec
and govulncheck; `make check` runs exactly what `ci.yml` runs):

1. `gofmt -s -l .` prints nothing. No key database or key export (`*.db`, its SQLite
   side files, `*.ckey`) is tracked; CI fails if one is (SEC-003).
2. `go mod tidy` leaves `go.mod`/`go.sum` unchanged, unless the change adds a
   dependency on purpose.
3. `go vet ./...` and `golangci-lint run` are clean.
4. `go test ./...` passes. Use `-race` when a C toolchain is available.
5. The CLI and TUI paths are both updated and tested.
6. On-disk formats stay backward-compatible (§3), and the golden fixture tests pass
   without touching the fixtures.
7. `intel/` docs are updated as `AGENTS.md` requires:
   - `map.md` for structure changes;
   - `cybersec.md` for security work;
   - `history.md` for significant changes;
   - `plan.md` and `notes.md` as needed.
