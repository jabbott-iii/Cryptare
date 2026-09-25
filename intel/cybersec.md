# Security Requirements and Issue Register

This file holds Cryptare's security requirements, identified issues, remediation items
and fix status (see `AGENTS.md` → Security Issue Tracking). Never delete items. Close
an item only after its remediation is implemented and its validation is complete.

Last updated: 2026-09-24

## 1. Security requirements

These apply to all changes.

1. **Confidentiality.** File contents and stored keys are protected with authenticated
   encryption (currently AES-256-GCM) under keys derived from user secrets. The KDF
   cost must meet current guidance (SEC-005).
2. **Password handling.**
   - Passwords are never logged, echoed, or written to disk.
   - They are read in full, including spaces.
   - Empty passwords are rejected on every encryption path.
   - Non-interactive password input should not require the secret to appear in the
     process arguments.
3. **Integrity.** Tampered or wrong-password ciphertext fails closed with a generic
   error. Directory artifacts stay bound to their type through AAD.
4. **Filesystem safety.**
   - Symlinks and special files are rejected when archiving.
   - Extraction never writes outside the destination and bounds its output size.
   - Outputs are created with mode `0o600`.
   - Plaintext is not staged outside the user-chosen locations.
5. **Key storage.**
   - Key material is stored only in encrypted form.
   - The database is not world-readable.
   - Deleted keys are not recoverable from the database file.
6. **Supply chain and CI.** Security scanning results are visible and actionable.
   Third-party actions and base images are pinned.
7. **No regression.** A change must not weaken a remediation recorded below.

## 2. Existing controls (verified 2026-09-23)

- **Encryption scheme.** AES-256-GCM with a fresh random 16-byte salt and 12-byte nonce
  per encryption, so each file gets a new key and nonce reuse is not a practical
  concern.
- **Failure handling.** Wrong password or tampering gives the generic error
  `decryption failed: wrong password or corrupted file`, and no output is written.
- **Directory artifacts.** They carry the magic `CRYPTARE-DIR-ENC\x00` and bind it as GCM
  AAD.
- **Output modes.** Encrypted, decrypted and compressed files are written with mode
  `0o600`.
- **Archiving.** Creating an archive rejects symlinks and non-regular files. Encryption
  rejects symlinked input paths.
- **Extraction.** `..` entries and archive-carried symlinks, hard links and special
  entries are rejected. Absolute entry names are re-rooted inside the destination
  (verified with probe tests).
- **Stored keys.** Keys are encrypted with the master password before storage. The
  `keys delete` command confirms by default and uses an atomic
  `DELETE … RETURNING` in a transaction.
- **CI scanning.** CodeQL (`security-extended`) runs on push, on PRs, and weekly.

## 3. Assessment method (2026-09-23 baseline)

- **Manual review:** all Go sources, tests, workflows, Dockerfile and git history.
- **Probe tests:** scratch tests run against an unmodified copy of the working tree,
  with Go 1.26.8 on linux/amd64. Dependencies were checked against the repository's
  `go.sum`.
- **Tools:**
  - gosec v2.29.0: 30 findings in non-test code.
  - golangci-lint v2.13.2: 0 issues.
  - `go vet`: clean.
  - `go test -race`: passing.
- **Artifacts:** git objects for `cryptare.db` (row counts only; no key data was
  extracted); the public v1.0.0 release assets.
- **Not run:** govulncheck (vulnerability database unreachable from the analysis
  environment), fuzzing, Windows and macOS behaviour, Docker build.

## 4. Summary

| ID | Title | Severity | Status |
|---|---|---|---|
| SEC-001 | Empty passwords accepted for encryption and key protection | High | Open |
| SEC-002 | Interactive password prompt truncates at whitespace and echoes input | High | Open |
| SEC-003 | Encrypted key material committed to the public repository | Medium | Open |
| SEC-004 | Passwords accepted as command-line arguments | Medium | Open |
| SEC-005 | KDF work factor below current guidance; formats unversioned | Medium | Open |
| SEC-006 | Directory encryption stages plaintext in the system temp directory | Medium | Open |
| SEC-007 | Unbounded decompression and extraction (decompression bomb) | Medium | Open |
| SEC-008 | Extraction follows existing symlinks in the destination and overwrites files | Low | Open |
| SEC-009 | Deleted keys remain recoverable from the database file | Low | Open |
| SEC-010 | Database created world-readable in the current directory on every run | Low | Open |
| SEC-011 | Imported key metadata not validated before storage and display | Low | Open |
| SEC-012 | CI security-scan results discarded; actions not pinned | Low | In Progress |
| SEC-013 | Container runs as root; base images not pinned | Low | Open |
| SEC-014 | Symlink race (TOCTOU) when archiving a directory tree | Low | Open |

## 5. Issue register

### SEC-001 — Empty passwords accepted for encryption and key protection

- **Status:** Open
- **Affected component:**
  - `internal/logic-tui.go` `buildActionCmd` (encrypt, keys generate, keys export)
  - `internal/crypto.go` `EncryptFile`, `EncryptKeyBlob`, `ExportKeyToFile` (none of
    these validate the password)
- **Risk:** Submitting a TUI form with the password field left blank produces an
  `.enc` file (or a stored key or export) that anyone can decrypt with an empty
  password. The TUI still reports success. The CLI is only protected by accident:
  `fmt.Fscan` waits for a non-empty token.
- **Evidence:** A probe test encrypted a file through the TUI with an empty password;
  `DecryptFile(…, "")` succeeded. TUI key generation with an empty password also
  succeeded.
- **Required remediation:**
  1. Reject empty passwords in the core encryption paths (`EncryptFile`,
     `EncryptKeyBlob`, `ExportKeyToFile`) with a typed error, and surface it in the CLI
     and TUI.
  2. Keep decryption of existing empty-password artifacts working.
  3. Decide the minimum-length policy (Q-004 in `notes.md`).
- **Validation:** Core, CLI and TUI tests that assert rejection on each encryption
  path, plus a test that legacy empty-password artifacts still decrypt.
- **Resolution:** —

### SEC-002 — Interactive password prompt truncates at whitespace and echoes input

- **Status:** Open
- **Affected component:** `internal/logic-cli.go` `readPassword` (uses `fmt.Fscan`);
  used by `encrypt`, `decrypt` and the `keys` subcommands.
- **Risk:**
  - A passphrase typed at the prompt (for example `correct horse battery staple`) is
    silently cut down to its first word (`correct`). That sharply reduces its strength.
  - The same passphrase later passed via `--password` does not decrypt the file.
  - Typed characters are echoed to the terminal, where they are exposed to onlookers,
    scrollback and session recordings.
  - The unread remainder stays in the stdin buffer.
- **Evidence:** A probe encrypted a file through the prompt with a four-word
  passphrase. `DecryptFile` succeeded with `correct` and failed with the full
  passphrase.
- **Required remediation:**
  1. Read a whole line, stripping only the trailing `\r\n`.
  2. When stdin is a terminal, read without echo using
     `github.com/charmbracelet/x/term` `ReadPassword`. It is already in the module
     graph and would be promoted to a direct dependency. When stdin is not a terminal,
     read one line.
  3. Consider confirming the password on encrypt.
  4. Migration note for release notes and README: files encrypted so far through the
     prompt with a multi-word passphrase can only be decrypted with the first word.
- **Validation:**
  - A CLI test with a multi-word passphrase passed via `SetIn` round-trips using the
    full passphrase.
  - A non-TTY test.
  - A manual TTY check that input is not echoed.
- **Resolution:** —

### SEC-003 — Encrypted key material committed to the public repository

- **Status:** Open
- **Affected component:** `cryptare.db` in git history.
  - First added in `d185a94` (2026-09-01) with no rows.
  - Holds one live `key_models` row in `06bda0b`.
  - Still tracked at `origin/main` (`f57e633`). There the row is deleted, but its
    encrypted blob remains in the file (see SEC-009).
  - Local commit `aa27461` removes the file but was not pushed as of 2026-09-23.
  - The repository was publicly accessible on 2026-09-23.
- **Risk:** Anyone can download the database and brute-force the master password
  offline (PBKDF2 at 100,000 iterations; see SEC-005). No code path currently uses
  stored keys to encrypt files, which limits direct impact. The master password itself,
  however, is exposed to guessing and may be reused elsewhere.
- **Required remediation:**
  1. The owner pushes the removal commit.
  2. Treat the master password used with that database as compromised: don't reuse it,
     and discard the key.
  3. Keep `*.db` in `.gitignore` (already present).
  4. Owner decision: purge the file from history (for example with `git filter-repo`)
     and force-push. This is irreversible and, under `AGENTS.md`, agents must not do
     it.
  5. Optional: a CI check that fails if a `*.db` file is tracked.
- **Validation:** `git ls-tree -r origin/main --name-only` does not list
  `cryptare.db`. If history is purged, `git log --all -- cryptare.db` on a fresh clone
  is empty.
- **Resolution:** —

### SEC-004 — Passwords accepted as command-line arguments

- **Status:** Open
- **Affected component:** `--password/-p` on `encrypt`, `decrypt`, `keys generate`,
  `keys export` and `keys import` (`internal/logic-cli.go`); the README examples.
- **Risk:** Command-line arguments are visible to other local users (`ps`,
  `/proc/<pid>/cmdline`) and are kept in shell history and CI logs.
- **Required remediation:**
  1. Keep the flag for compatibility.
  2. Add a non-interactive channel that keeps the secret out of the arguments
     (`--password-file <path>` or `--password-stdin`).
  3. Print a warning to stderr when `--password` is used.
  4. Change the README examples to prefer the prompt or stdin.
- **Validation:** Tests for the new input path and for the warning; README review.
- **Resolution:** —

### SEC-005 — KDF work factor below current guidance; formats unversioned

- **Status:** Open
- **Affected component:** `internal/crypto.go` `deriveKey` (PBKDF2-HMAC-SHA256,
  `pbkdf2Iter = 100_000`) and every format in `maint.md` §3.
- **Risk:** The OWASP Password Storage Cheat Sheet recommends at least 600,000
  iterations for PBKDF2-HMAC-SHA256, or Argon2id with at least m=19 MiB, t=2, p=1. At
  100k iterations, offline guessing against a stolen `.enc`, `.ckey` or database is
  about 6× cheaper than that guidance. The formats have no version or KDF-parameter
  header, so the cost can't be raised without breaking existing files.
- **Required remediation:**
  1. Introduce a versioned header (magic, version, KDF identifier and parameters) for
     newly written files, key blobs, exports and directory artifacts.
  2. Use Argon2id (`golang.org/x/crypto/argon2`, already in the module) or PBKDF2 at
     ≥600k iterations.
  3. Keep the legacy headerless decrypt path.
  4. Replace the frozen `x/crypto/pbkdf2` wrapper with the standard library's
     `crypto/pbkdf2`. The output is identical.
- **Validation:** Known-answer tests for the new format, the existing legacy tests
  passing, and a benchmark of the KDF cost on target hardware.
- **Resolution:** —

### SEC-006 — Directory encryption stages plaintext in the system temp directory

- **Status:** Open
- **Affected component:** `internal/crypto.go` `createDirectoryArchiveTempFile`
  (`os.CreateTemp("", "cryptare-dir-*.tar.gz")`) and `encryptDirectory`.
- **Risk:**
  - A full plaintext tar.gz of the directory is written to `$TMPDIR` (the file itself
    is mode 0600). That location is often a shared `/tmp` and may be persistent,
    backed up, or on another device.
  - The file is removed with `os.Remove`, which does not wipe it, and it is left behind
    if the process is killed.
- **Evidence:** A probe with an unusable `TMPDIR` failed with
  `create temporary directory archive: open $TMPDIR/cryptare-dir-….tar.gz`.
- **Required remediation:** Build the tar.gz in memory. The whole archive is already
  read into memory before encryption, so this adds no new memory ceiling. Longer term,
  stream into an authenticated chunked format (with SEC-005).
- **Validation:** A test asserting that no files are created in `TMPDIR` during
  directory encryption, with the existing round-trip tests still passing.
- **Resolution:** —

### SEC-007 — Unbounded decompression and extraction (decompression bomb)

- **Status:** Open
- **Affected component:** `internal/compress.go` `DecompressFile`, `extractTarGz`,
  `extractZip`, `extractZipSingleFile`. gosec G110 fires at `compress.go` lines 155,
  416, 498 and 551.
- **Risk:** A small crafted `.gz`, `.tar.gz` or `.zip` can expand until the disk is
  full, or create enough entries to exhaust inodes. This matters when opening archives
  from untrusted sources.
- **Required remediation:**
  1. Cap total extracted bytes and entry count, with sane defaults and a flag to
     override them.
  2. Copy through `io.LimitReader` or `io.CopyN`.
  3. Abort and remove partial output when a limit is hit.
- **Validation:** Tests with synthetic high-ratio archives. G110 is resolved or
  explicitly justified.
- **Resolution:** —

### SEC-008 — Extraction follows existing symlinks in the destination and overwrites files

- **Status:** Open
- **Affected component:** `internal/compress.go` `extractTarGz`, `extractZip`,
  `extractZipSingleFile`. The destination check is lexical only; `os.MkdirAll` and
  `os.OpenFile(…O_TRUNC…)` follow symlinks. gosec G703 flags the same code.
- **Risk:**
  - If the destination already contains a symlink (for example when extracting into a
    shared or attacker-writable directory), entries under that name are written
    outside the destination.
  - Existing files are truncated and overwritten without warning.
  - Permission bits from the archive are applied, subject to the umask.
- **Evidence:** A probe with `dst/link → ../outside` and an archive entry
  `link/escaped.txt` created `outside/escaped.txt` with no error. Plain `..` entries
  were correctly rejected.
- **Required remediation:**
  1. Extract through an `os.Root` opened on the destination (Go ≥ 1.24; this module
     targets 1.26). It refuses paths that escape the root, including via symlinks.
  2. Refuse to overwrite existing files unless explicitly requested.
  3. Mask archive-supplied modes.
- **Validation:** A regression test (symlink in the destination) that fails before the
  fix and passes after it; G703 findings on extraction reviewed.
- **Resolution:** —

### SEC-009 — Deleted keys remain recoverable from the database file

- **Status:** Open
- **Affected component:** `internal/database.go` `NewDatabase` (SQLite `secure_delete`
  is off) and `DeleteKey`; the README statement that deleted keys "cannot be
  recovered".
- **Risk:** After `keys delete`, the encrypted blob stays in SQLite free pages until
  they are reused or the database is vacuumed. Anyone with the file (backups, or git
  as in SEC-003) can recover it and attack it offline (SEC-005). `zeroBytes` only wipes
  the in-process copy.
- **Evidence:** A probe showed the raw database bytes still contain the blob after
  `DeleteKey`. Committed versions of `cryptare.db` with zero live rows still contain
  blob-shaped data.
- **Required remediation:**
  1. Enable secure delete, either with the `_secure_delete=on` DSN parameter (supported
     by go-sqlite3 v1.14.52) or with `PRAGMA secure_delete=ON`.
  2. Consider running `VACUUM` after a delete.
  3. Check the journal/WAL files too.
  4. Align the README wording with the result.
- **Validation:** A test asserting that the database file bytes do not contain a
  deleted blob.
- **Resolution:** —

### SEC-010 — Database created world-readable in the current directory on every run

- **Status:** Open
- **Affected component:** `main.go`, `database_path.go` (default `cryptare.db`) and
  `internal/database.go` `NewDatabase`.
- **Risk:** Every invocation, including `--help`, `encrypt` and `compress`, creates
  `cryptare.db` in the current directory with mode 0644 (under umask 022). Key IDs and
  encrypted key blobs therefore land in shared or synced directories and git
  worktrees; that is how SEC-003 happened.
- **Evidence:** Running `cryptare --help` in an empty directory created `cryptare.db`
  with mode `-rw-r--r--`.
- **Required remediation:**
  1. Open the database only for `keys` commands and the TUI.
  2. Create the file with mode 0600.
  3. Owner decision: a per-user default path (for example under `os.UserConfigDir()`),
     keeping `CRYPTARE_DB_PATH` as an override (Q-003).
- **Validation:** Tests asserting that `encrypt`, `compress` and `--help` create no
  database, and that a newly created database has mode 0600.
- **Resolution:** —

### SEC-011 — Imported key metadata not validated before storage and display

- **Status:** Open
- **Affected component:** `internal/crypto.go` `ImportKeyFromFile` (`KeyID`,
  `Algorithm` and `CreatedAt` taken as-is); `keys list` and the TUI key table.
- **Risk:** A crafted `.ckey` file, whose password the victim knows, can inject
  terminal control sequences through `KeyID` or `Algorithm`, and can store a blob that
  never decrypts.
- **Required remediation:**
  1. Validate the key ID format (generated IDs are 16 lower-case hex characters).
  2. Allowlist the algorithm (`AES-256-GCM`).
  3. Validate the blob's base64 structure and length.
  4. Reject anything else with a clear error.
- **Validation:** Import tests with malicious fields.
- **Resolution:** —

### SEC-012 — CI security-scan results discarded; actions not pinned

- **Status:** In Progress
- **Progress (2026-09-24):** Staged, uncommitted workflow changes cover remediation
  steps 1, 2 and 4, and were verified:
  - all 9 third-party actions are pinned to full commit SHAs, each matching its release
    tag on GitHub;
  - gosec is pinned to v2.29.0;
  - `results.sarif` is uploaded with category `gosec`.

  These changes shipped in `5286920` (tag v1.0.1), and the owner reports that the
  Security workflow succeeded. Remaining: confirm that gosec alerts appear in Code
  Scanning, triage them (step 3), and decide on govulncheck (step 5).
- **Affected component:**
  - `.github/workflows/security.yml`: `securego/gosec@master` is unpinned, runs with
    `-no-fail`, and its SARIF output is never uploaded.
  - All workflows: actions are pinned by tag only.
  - `ci.yml`: uses `codecov/codecov-action@v3`.
- **Risk:**
  - gosec findings (30 with v2.29.0 on 2026-09-23) are invisible.
  - An unpinned `@master` action runs whatever upstream code is current, in a job with
    `security-events: write`.
  - Tag-pinned third-party actions can be re-pointed by their owners.
- **Required remediation:**
  1. Pin gosec to a release or commit SHA.
  2. Upload `results.sarif` with `github/codeql-action/upload-sarif`.
  3. Triage the findings. G304/G703 on user-chosen paths are expected for a file tool
     and should be annotated with a justification.
  4. Pin third-party actions to commit SHAs.
  5. Consider adding govulncheck.

  Changing CI requires explicit owner approval (`AGENTS.md`).
- **Validation:** gosec alerts appear in Code Scanning, and workflow `uses:` lines
  reference SHAs.
- **Resolution:** —

### SEC-013 — Container runs as root; base images not pinned

- **Status:** Open
- **Affected component:** `Dockerfile`.
- **Risk:** The container process runs as root, so files it writes into mounted
  volumes are root-owned, and a compromise of the process has root in the container.
  `golang:1.26-alpine` and `alpine:3.22` are mutable tags.
- **Required remediation:**
  1. Add a non-root `USER` that owns `/app/data`.
  2. Pin both images by digest.

  This requires owner approval (it changes deployment configuration).
- **Validation:** `docker run --rm --entrypoint id <image> -u` prints a non-zero UID,
  and the docker smoke test passes.
- **Resolution:** —

### SEC-014 — Symlink race (TOCTOU) when archiving a directory tree

- **Status:** Open
- **Affected component:** `internal/compress.go` `writeTarGz` (gosec G122 at line 254)
  and `writeZipDirectory`/`writeZipFile`; these are also used by directory encryption.
- **Risk:** Entry types are checked from `WalkDir` metadata, and then the path is
  reopened by name. Someone who can modify the source tree while it is being archived
  can swap a file for a symlink in between. The contents of the link's target (outside
  the tree) then end up in the archive or encrypted output.
- **Required remediation:** Open files relative to an `os.Root` on the source
  directory, or open with no-follow semantics and compare the result with the walked
  entry's metadata.
- **Validation:** A regression test using a swap hook, or a code review demonstrating
  root-scoped opens; G122 resolved.
- **Resolution:** —
