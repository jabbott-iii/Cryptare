# Security Requirements and Issue Register

This file holds Cryptare's security requirements, identified issues, remediation items
and fix status (see `AGENTS.md` → Security Issue Tracking). Never delete items. Close
an item only after its remediation is implemented and its validation is complete.

Last updated: 2026-10-04 (round 5: v1.3.0 released; SEC-003, SEC-010 and SEC-016 closed)

## 1. Security requirements

These apply to all changes.

1. **Confidentiality.** File contents and stored keys are protected with authenticated
   encryption (currently AES-256-GCM) under keys derived from user secrets. The KDF
   cost must meet current guidance (SEC-005).
2. **Password handling.**
   - Passwords are never logged, echoed, or written to disk.
   - They are read in full, including spaces.
   - Empty passwords are rejected on every encryption path.
   - Passwords that protect new data (encrypt, stored keys, key exports) meet the
     password policy: at least 15 Unicode code points and not one repeated character,
     with no composition rules (`CheckPasswordPolicy`, following NIST SP 800-63B-4
     §3.1.1.2). Typed passwords are confirmed. Decryption and import accept any
     password so that existing data stays readable.
   - Non-interactive password input should not require the secret to appear in the
     process arguments.
3. **Integrity.** Tampered or wrong-password ciphertext fails closed with a generic
   error. Directory artifacts stay bound to their type through AAD.
4. **Filesystem safety.**
   - Symlinks and special files are rejected when archiving.
   - Extraction never writes outside the destination and bounds its output size
     (default 10 GiB and 100,000 entries, SEC-007).
   - Outputs are created with mode `0o600`.
   - Plaintext is not staged outside the user-chosen locations.
   - A command stopped by a signal, or a TUI action cancelled by quitting, removes its
     unfinished output (SEC-015).
5. **Key storage.**
   - Key material is stored only in encrypted form.
   - The database is not world-readable.
   - On Unix, a key database or SQLite side file that another user owns or that
     group or others can write is refused (SEC-016).
   - Values read from the database are escaped before they are shown in a terminal
     (SEC-016).
   - Deleted keys are not recoverable from the database file.
   - A stored key that encrypts files is unlocked with its master password for that
     operation only, and each file gets its own key derived from it (plan 3.4); a
     key's master password must meet the password policy to encrypt new files or
     protect an export (SEC-001, BUG-011).
6. **Supply chain and CI.** Security scanning results are visible and actionable.
   Third-party actions and base images are pinned. Builds use a pinned, current Go
   toolchain, and govulncheck runs in CI (SEC-018). Published release files carry
   signed build provenance (SEC-020), and vulnerabilities are reported privately
   (`SECURITY.md`).
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

### 3a. Re-analysis (2026-10-03)

- **Scope:** all Go sources at `0c57aef` (including the version 2 format), the tests
  that guard the format, workflows, `Dockerfile`, `.gitignore`, `.devcontainer/` and
  the README's security claims, checked against the register below so that only new
  issues or gaps in earlier fixes are added.
- **Probes:** scratch tests and real binaries built from a copy of the tree, outside
  the repository. Go 1.26.0 (the toolchain `go.mod` resolves to) on linux/amd64, as
  a non-root user.
- **Tools:** `gofmt -s -l` and `go vet ./...` clean; `go test -race -count=1 ./...`
  passes (75.0% coverage for `main`, 75.4% for `internal`); `go mod verify` reports
  all modules verified (offline, against the module cache).
- **Scanners**, run once the owner approved the downloads (installed in a scratch
  folder, built with Go 1.26.8):
  - golangci-lint v2.13.2 (CI's version, default linters): 0 issues;
  - gosec v2.29.0: the same 10 findings as on 2026-09-27, triaged under SEC-012;
  - govulncheck v1.8.0 (database of 2026-10-01): with the go1.26.0 toolchain that
    `go.mod` selects, 3 standard-library vulnerabilities are reachable from Cryptare's
    code, 6 more sit in imported packages but aren't called, and 25 apply only at
    module level. With go1.26.8, none are reachable or imported (SEC-018).
- **Not run:** Windows and macOS; the Docker build; fuzzing.
- **Result:** SEC-015 to SEC-019 added; gaps recorded in SEC-003, SEC-005, SEC-010,
  SEC-011 and SEC-013. No flaw was found in the version 2 construction itself (STREAM
  nonces, last-chunk flag, header as additional data, header limits checked before key
  derivation, folder streams drained to the final chunk).

## 4. Summary

| ID | Title | Severity | Status |
|---|---|---|---|
| SEC-001 | Empty passwords accepted for encryption and key protection | High | Closed |
| SEC-002 | Interactive password prompt truncates at whitespace and echoes input | High | Closed |
| SEC-003 | Encrypted key material committed to the public repository | Medium | Closed |
| SEC-004 | Passwords accepted as command-line arguments | Medium | Closed |
| SEC-005 | KDF work factor below current guidance; formats unversioned | Medium | Closed |
| SEC-006 | Directory encryption stages plaintext in the system temp directory | Medium | Closed |
| SEC-007 | Unbounded decompression and extraction (decompression bomb) | Medium | Closed |
| SEC-008 | Extraction follows existing symlinks in the destination and overwrites files | Low | Closed |
| SEC-009 | Deleted keys remain recoverable from the database file | Low | Closed |
| SEC-010 | Database created world-readable in the current directory on every run | Low | Closed |
| SEC-011 | Imported key metadata not validated before storage and display | Low | Closed |
| SEC-012 | CI security-scan results discarded; actions not pinned | Low | Closed |
| SEC-013 | Container runs as root; base images not pinned | Low | In Progress |
| SEC-014 | Symlink race (TOCTOU) when archiving a directory tree | Low | Closed |
| SEC-015 | Interrupted decrypt or extraction leaves partial plaintext in hidden temporary files | Medium | Closed |
| SEC-016 | Key database files from untrusted locations are trusted | Low | Closed |
| SEC-017 | GORM's default logger prints SQL with bound values to stdout | Low | Closed |
| SEC-018 | Builds use the Go 1.26.0 toolchain, with reachable standard-library vulnerabilities | Medium | Closed |
| SEC-019 | Owner-only permission guarantees don't hold on Windows | Low | Closed |
| SEC-020 | Release files have no verifiable proof of origin | Low | In Progress |

## 5. Issue register

### SEC-001 — Empty passwords accepted for encryption and key protection

- **Status:** Closed (2026-09-27)
- **Progress (2026-09-24, committed in `ed46150`):** Remediation steps 1 and 2 are implemented
  and validated.
  - `EncryptFile` and `EncryptKeyBlob` return `ErrEmptyPassword`, which covers files,
    directory archives, key generation and `ExportKeyToFile`. The CLI and TUI show
    the error, and nothing is written.
  - Decryption still accepts an empty password, so legacy artifacts stay readable.
  - Tests: `TestEncryptRejectsEmptyPassword`, `TestDecryptAcceptsLegacyEmptyPassword`,
    `TestDashboardRejectsEmptyPassword`, `TestEncryptCmdRejectsEmptyInteractivePassword`.
    The core and TUI tests failed before the fix. End to end, the fixed TUI shows
    "password must not be empty" and writes no file, and its Decrypt form still
    recovers a file the old build encrypted with a blank password.
  - Remaining: step 3, the minimum-length policy (Q-004). This item stays open
    until that is decided and implemented.
- **Progress (2026-09-27, `b415ffc`):** step 3 is implemented and validated. The owner
  asked for a default policy (Q-004).
  - `CheckPasswordPolicy` in `crypto.go` requires at least `MinPasswordLength` (15)
    Unicode code points and rejects a single repeated character (`ErrWeakPassword`).
    It has no composition rules and no maximum length, following NIST SP 800-63B-4
    §3.1.1.2 for single-factor passwords.
  - `EncryptFile` and `EncryptKeyBlob`, and through it `ExportKeyToFile`, enforce the
    policy in place of the empty-only check, so the CLI (`--password` and prompt) and
    the TUI all get it. The CLI prompt and the TUI also check it before asking for
    confirmation (SEC-002 step 3).
  - Decryption and import don't check it. `TestDecryptAcceptsLegacyShortPassword` and
    `TestLegacyShortPasswordCmds` show files, key blobs and exports protected with a
    short password still open, and a key stored under a short master password can
    be exported with a compliant export password.
  - New tests, which failed against the previous code: `TestCheckPasswordPolicy` (14
    cases, including code-point counting and invalid UTF-8),
    `TestEncryptRejectsWeakPassword`, `TestNewPasswordCmdsRejectWeakPassword` (flag
    and prompt for encrypt, keys generate and keys export) and
    `TestDashboardRejectsWeakOrMismatchedPassword`.
  - Not covered: a blocklist of common or breached passwords, which NIST also asks
    verifiers for. Cryptare has no offline list today; a 15-character minimum still
    admits weak phrases such as a repeated word.
  - The CI, CD and Docker smoke tests used passwords shorter than 15 characters. The
    workflow patch `cryptare-password-policy-workflows.patch` lengthens them (plan W9);
    the owner has applied it.
  - CI passed on `b415ffc` with the patched workflows; closed 2026-09-27.
- **Update (2026-10-04, round 6, BUG-011 and plan 3.4; not yet committed):** the policy
  now also covers passwords that protect new data indirectly. `keys export` protects
  the export with the key's own master password (owner decision, one password per
  key), so `ExportKeyToFile` checks that the password unlocks the key and then applies
  `CheckPasswordPolicy` to it; `StoredKeyCredential` applies it when a stored key
  encrypts new files, and `EncryptFileWithCredentialContext` refuses a stored key that
  wasn't unlocked for new data, so the check is in core (`TestDecryptOnlyCredentialDoesntEncrypt`). The capability noted above, exporting a key stored under a short
  master password with a longer export password, is gone: such a key still decrypts,
  but can't encrypt new files or be exported (the user generates a new key). Nothing
  protects new data with a password that fails the policy, so this remediation is
  unchanged. Tests: `TestEncryptRejectsEmptyPassword` and
  `TestEncryptRejectsWeakPassword` (export of a legacy key under an empty or short
  master password), `TestLegacyShortPasswordCmds`, `TestStoredKeyCredentialChecks`.
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
- **Resolution:** Fixed in `ed46150` (empty passwords, v1.1.0) and `b415ffc` (password policy, 2026-09-27). Validated by the tests and end-to-end checks above, and by green CI on `b415ffc`: CI #129 (Ubuntu, macOS and Windows, including the smoke tests with the new passwords), Docker #11 and Security #134. Closed 2026-09-27.

### SEC-002 — Interactive password prompt truncates at whitespace and echoes input

- **Status:** Closed (2026-10-04)
- **Progress (2026-10-04, verified on GitHub):** v1.2.0 was released by CD #4 on
  `89a64e8`, and its release notes carry both migration notes (the v1.0.1 prompt
  change and the 15-character minimum for scripts). That was the last remaining step.
- **Progress (2026-09-24, committed in `ed46150`):** Remediation steps 1, 2 and 4 are
  implemented and validated.
  - `readPassword` reads the whole line and strips only `\r`/`\n`. On a terminal it
    reads without echo through `github.com/charmbracelet/x/term`, which is now a
    direct dependency; `go.mod` moved one line and `go.sum` is unchanged.
  - The README states the new behaviour and has the migration note.
  - Tests: `TestReadPassword` (7 cases),
    `TestEncryptDecryptCmdMultiWordPromptPassword` and
    `TestDecryptCmdPromptAcceptsEmptyPasswordForLegacyFiles`. They failed before the
    fix.
  - Manual TTY check in a pseudo-terminal, with a four-word passphrase containing a
    Backspace-corrected typo: the old binary echoed it and kept only the first word;
    the fixed binary doesn't echo it, and the file decrypts with the full passphrase
    but not with the first word.
  - BUG-012 is fixed (plan 1.1a). This fix had introduced a problem: Ctrl+C at the
    hidden prompt left terminal echo off. The prompt now restores the terminal and
    exits with status 130; verified in a pseudo-terminal.
  - Remaining: step 3 (confirmation on encrypt, tied to Q-004), and the migration
    note in the next release's notes.
- **Progress (2026-09-27, `b415ffc`):** step 3 is implemented and validated.
  - `readNewPassword` in `logic-cli.go` asks for the password a second time when the
    input is a terminal and refuses a mismatch (`ErrPasswordMismatch`). Piped input
    is read once, so scripts keep working. It is used by `encrypt`, `keys generate`
    and `keys export`.
  - The TUI's Encrypt, Generate key and Export key forms have a masked "Confirm
    password" field, checked after the policy.
  - Tests: `TestReadNewPasswordWith` (7 cases) and
    `TestDashboardRejectsWeakOrMismatchedPassword` failed against the previous code;
    `TestDashboardNewPasswordFormsHaveConfirmation` checks the fields.
  - In a pseudo-terminal the new binary passed 7 of 7 checks: it asks to confirm, a
    mismatch writes nothing, a weak password is refused before confirmation, Ctrl+C
    at the confirmation exits 130 with echo restored, piped input is read once, and
    the TUI refuses a mismatch and then accepts a match. A build of the previous
    commit (`b4cd66f`) passed 1 of 7, the piped-input check.
  - Remaining: the migration notes in the next release's notes (the v1.0.1 prompt
    change, and the new 15-character minimum for scripts).
  - 2026-10-04: the notes are drafted for the next release (see `history.md`); this
    item closes when that release is published with them.
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
- **Resolution:** Fixed in `ed46150` (steps 1, 2 and 4) and `b415ffc` (step 3, the
  confirmation), validated by the tests and pseudo-terminal checks above; the migration
  notes shipped with the v1.2.0 release on 2026-10-04. Closed 2026-10-04.

### SEC-003 — Encrypted key material committed to the public repository

- **Status:** Closed (2026-10-04)
- **Note (2026-10-04, round 6; not yet committed):** the golden format fixtures in
  `internal/testdata/golden/` include stored keys and key exports written by v1.0.1 and
  v1.3.1. They are deliberate, public test vectors: every password is published in
  that folder's `README.md`, and the keys protect nothing but the fixtures. They are
  kept as `.txt` (their base64 text), not `.ckey` or `.db`, so the guard keeps its
  meaning for real key material and doesn't need an exception. This is a judgement for
  the owner to confirm; the alternative is an explicit path exception in the guard.
- **Progress (2026-10-04, verified on GitHub):** the CI guard is committed in `a088f7f`
  and passed on Ubuntu, Windows and macOS in CI #143; it also ran on Dependabot's
  rebased PR #26 (CI #144, #145). `git ls-tree -r a088f7f` lists no key database or key
  export. v1.3.0 was released from `a088f7f`.
- **Progress (2026-10-04, Q-005 answered; not yet committed):**
  - Step 4: the owner decided not to purge history (Q-005 = no). `cryptare.db` stays in
    the history from `d185a94` to its removal in `aa27461` and in tag `v1.0.0`
    (`768f4cd`); the key it held was discarded (step 2). The
    `copilot/research-compression-implementation` branch, whose tip still tracked the
    file, was no longer on GitHub when checked on 2026-10-04 (`git ls-remote`); its tip
    and unmerged commits are recorded in `history.md`.
  - Step 5: `ci.yml` has a "Refuse committed key material" step after checkout, which
    fails when `git ls-files` lists a `*.db`, `*.db-journal`, `*.db-wal`, `*.db-shm` or
    `*.ckey` file (case-insensitive). Validation: it passes on the current tree, fails
    on the `v1.0.0` tree (it lists `cryptare.db`) and on a scratch repository with
    `keys.CKEY` and `x.db-wal` tracked; actionlint 1.7.12 reports nothing.
  - The step 3 patterns (`*.ckey` and the SQLite side files) were committed in
    `eb330a3`, with CI green.
  - Closes once the CI step is committed and CI passes.
- **Progress (2026-10-04, plan 5.9; not yet committed):** the gap in step 3 is closed:
  `.gitignore` adds `*.ckey`, `*.db-journal`, `*.db-wal` and `*.db-shm`, and `git
  check-ignore` matches each; no tracked file matches them. Remaining: step 4 (Q-005)
  and the optional CI check (step 5).
- **Gap found (2026-10-03 analysis):** `.gitignore` covers `*.db` but not SQLite's
  side files (`*.db-journal`, `*.db-wal`, `*.db-shm`), which can hold copies of
  database pages, or key exports (`*.ckey`), which `keys export` writes to the current
  directory by default. Working inside a repository checkout, as happened here, can
  still commit key material. Recommended addition to step 3: ignore those patterns,
  and cover them in step 5's CI check (plan 5.9).
- **Progress (2026-09-26):**
  - Step 1 is done: the removal was pushed on 2026-09-23, and `origin/main` no longer
    tracks the file.
  - Step 2 is done: the owner discarded the key.
  - Step 3 is in place (`*.db` in `.gitignore`).
  - Remaining: step 4, the history-purge decision (Q-005), and the optional CI check
    (step 5).
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
- **Resolution:** The file was removed from `main` (`aa27461`, pushed 2026-09-23) and
  the key discarded (step 2). `.gitignore` covers databases, their SQLite side files
  and key exports (steps 3 and 5.9), and CI refuses any that are tracked (step 5,
  `a088f7f`). The owner chose not to purge history (Q-005), so the file stays in old
  commits and tag `v1.0.0`; the key it held is no longer used. Validated as above, with
  CI green. Closed 2026-10-04.

### SEC-004 — Passwords accepted as command-line arguments

- **Status:** Closed (2026-10-04)
- **Progress (2026-09-27, `e512844`; plan 2.6):** all four remediation steps are
  implemented and validated. The owner chose `--password-file`, with piped input as
  the stdin route (it already worked), and a warning that can't be switched off.
  - `passwordFlags` in `logic-cli.go` registers `--password`/`-p` (kept, step 1) and
    the new `--password-file` on `encrypt`, `decrypt`, `keys generate`, `keys export`
    and `keys import`. The two flags are mutually exclusive.
  - `readPasswordFile` reads the first line (at most 64 KiB), keeps spaces and
    removes only the line ending, like piped input. An empty or missing file is an
    error, not a fallback to the prompt. The password policy still applies.
  - `--password` prints a one-line warning on stderr (step 3). The warning doesn't
    repeat the password.
  - The README examples use `--password-file` and the prompt, and explain piping
    (step 4).
  - Tests: `TestReadPasswordFile` (8 cases), `TestPasswordFileFlag` (all five commands
    round-trip, no warning; weak password refused; both flags refused; missing file
    refused) and `TestPasswordFlagWarns`. All failed against the previous code.
  - With a real binary: `--password-file` encrypted with nothing on stderr, piped
    input decrypted, `--password` worked and printed the warning, and combining the
    flags was refused.
  - gosec reports one new G304 (`os.Open` on the user's password-file path), the same
    class as the other file-path findings.
  - The CI, CD and Docker smoke tests still used `--password`. The owner approved
    switching them to `--password-file` (W11); the patch is delivered.
  - Close after the change is committed and CI passes.
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
- **Resolution:** Fixed in `e512844` (plan 2.6). Validated by the tests and real-binary checks above, the W11 smoke tests that now pass the password with `--password-file`, and green CI on `e512844` (CI #132, Docker #14, Security #137), confirmed on 2026-10-04. Closed 2026-10-04.

### SEC-005 — KDF work factor below current guidance; formats unversioned

- **Status:** Closed (2026-10-04)
- **Update (2026-10-03 analysis):** plans 3.1 and 3.2 were committed in `0c57aef`
  (2026-09-27, "fix: reformated outputs"); its CI results haven't been reported, so
  this stays open until they are green. The re-analysis found no flaw in the
  construction. Measured worst case of the read limits: a crafted 62-byte `.enc` or
  `.ckey` asking for the maximum (1 GiB, 10 passes, 16 lanes) made `decrypt` and
  `keys import` use about 1 GiB of memory and 1.9 s per attempt on a 16-core machine.
  That is the intended bound, but it is 16 times the default setting and can exhaust
  small machines or containers (owner question Q-008).
- **Progress (2026-09-27, uncommitted; plans 3.1 and 3.2):** all four remediation steps
  are implemented and validated. Owner decisions: Argon2id with 64 MiB, 3 passes and
  4 lanes (RFC 9106's second recommended setting, above OWASP's minimum of 19 MiB, 2
  passes, 1 lane); the header and chunked streaming built together; every kind of
  artifact moves to the new format; old formats stay readable with no migration
  command.
  - The new `internal/format_v2.go` writes a 46-byte header (magic, version, content
    type, key source, KDF, Argon2id memory, passes and lanes, salt, chunk size, nonce
    prefix), then AES-256-GCM chunks of 64 KiB in the STREAM construction. The whole
    header is each chunk's additional data. The layout is in `maint.md` §3.
  - `EncryptFile` (files and folders), `EncryptKeyBlob` and `ExportKeyToFile` write
    only version 2. `DecryptFile`, `DecryptKeyBlob` and `ImportKeyFromFile` read
    version 2 and, for anything without the version 2 magic, the legacy layouts.
  - Hostile headers: settings above 1 GiB of memory, 10 passes or 16 lanes, unknown
    versions, content types, key sources and KDFs, and odd chunk sizes are refused
    before any key derivation. A file asking for 4 GiB fails at once.
  - The content type is authenticated, so a key export can't be decrypted as a file
    or a stored key imported as an export.
  - `deriveKey` (legacy only) uses the standard library's `crypto/pbkdf2`, which
    returns an error where the frozen `x/crypto/pbkdf2` wrapper panicked.
  - Tests:
    - `TestEncryptWritesVersion2Format` fails against the previous code: the file,
      folder, stored key (76 bytes) and export (250 bytes) all lacked the header.
    - `TestDefaultPasswordKDFIsWritten` checks the 64 MiB / 3 / 4 default.
    - `TestStreamRejectsTampering`: 12 kinds of tampering on a file and a folder, all
      refused with nothing written.
    - `TestDecryptRejectsUnsupportedHeaders` (12 header cases) and
      `TestDecryptRejectsWrongContentType`.
    - `TestDeriveKey` now checks PBKDF2 known answers computed with Python's
      `hashlib`.
    - Every legacy test still passes.
  - KDF cost, measured on the 2-vCPU validation VM:
    - Argon2id at 64 MiB, 3 passes, 4 lanes: about 0.12 s per derivation, 64 MiB of
      memory;
    - the legacy PBKDF2 at 100,000 iterations: about 0.02 s.

    A small `encrypt` now takes 0.24 s instead of 0.03 s.
  - Real binaries:
    - the previous build's file, folder, stored key and `.ckey` all read correctly
      with the new build;
    - version 2 output from the new build is refused by the previous build with the
      generic error, and nothing is written.
  - Not covered:
    - existing `.enc` files, database rows and `.ckey` files keep their 100,000-
      iteration PBKDF2 protection until they are re-encrypted or re-exported. By the
      owner's choice there is no migration command;
    - a legacy file whose random salt happens to begin with the 9-byte magic
      (probability 2^-72) would be read as version 2 and fail.
  - Close after the change is committed and CI passes.
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
- **Resolution:** Fixed in `0c57aef` (plans 3.1 and 3.2). Validated by the known-answer tests, the legacy tests and the KDF cost measured above, and by green CI on `0c57aef` (CI #134, Docker #16, Security #139), confirmed on 2026-10-04. Q-008 (whether to lower the Argon2id settings accepted when reading) is an open owner question tracked in `notes.md`, not part of this remediation. Closed 2026-10-04. Update (2026-10-04): Q-008 is answered; the owner keeps the read limits (1 GiB, 10 passes, 16 lanes) as they are.

### SEC-006 — Directory encryption stages plaintext in the system temp directory

- **Status:** Closed (2026-09-27)
- **Progress (2026-09-27, `d683739`; plan 2.3):** the remediation is implemented and
  validated.
  - `buildDirectoryArchive` in `crypto.go` builds the tar.gz in a `bytes.Buffer` and
    replaces `createDirectoryArchiveTempFile`. The archive bytes and the
    `.enc` format are unchanged.
  - `TestEncryptDirectoryWritesNoTempPlaintext` points `TMPDIR`, `TMP` and `TEMP` at a
    missing folder. The previous code failed with "create temporary directory
    archive"; the new code encrypts and round-trips.
  - With `strace`, the previous build (`d751967`) opened
    `$TMPDIR/cryptare-dir-….tar.gz` for writing. The new build writes only the
    output's own temporary file (ciphertext) and the key database. Artifacts made by
    either build decrypt with the other to identical trees.
  - gosec reports one fewer G304 finding.
  - Still true: the plaintext archive is held in process memory until encryption
    finishes, as before. Go can't reliably wipe it. Streaming, chunked encryption
    (plan 3.2, with SEC-005) would limit this.
  - Committed in `d683739`; CI passed on `a5edc91` (after a test-only Windows fix); closed 2026-09-27.
  - Follow-up (2026-09-27, uncommitted; plan 3.2): directory encryption now streams
    the tar.gz straight into the chunked version 2 format (SEC-005), so the whole
    archive is no longer held in memory; only the current 64 KiB chunk is.
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
- **Resolution:** Fixed in `d683739` (plan 2.3): directory archives are built in memory, so no plaintext is written to the temp folder. Validated by `TestEncryptDirectoryWritesNoTempPlaintext`, the `strace` check and green CI on `a5edc91` (Ubuntu, macOS and Windows). The in-memory copy remains until streaming encryption (plan 3.2). Closed 2026-09-27.

### SEC-007 — Unbounded decompression and extraction (decompression bomb)

- **Status:** Closed (2026-09-27)
- **Note (2026-10-03):** builds made with Go older than 1.26.2, which includes what CI
  produces from this tree today, can be driven to exhaust memory by a crafted `.tar.gz`
  inside `tar.Reader.Next`, before this item's limits apply (GO-2026-4869, verified).
  The code-level remediation stands. The toolchain fix is tracked as SEC-018, so the
  status is unchanged.
- **Progress (2026-09-27, `b415ffc`; plan 2.1):** all three remediation steps are
  implemented and validated. The owner approved the defaults (10 GiB and 100,000
  entries), the flags, cleanup through a temporary folder, and applying the limits
  to encrypted folders too.
  - `ExtractLimits` (`MaxBytes`, `MaxEntries`; 0 means no limit) and
    `DefaultExtractLimits()` in `compress.go`. `DecompressFileWithLimits` and
    `DecryptFileWithLimits` take them; `DecompressFile` and `DecryptFile` use the
    defaults.
  - An `extractBudget` counts entries and copies each entry through `io.CopyN`,
    reading at most one byte past the size limit. Hitting a limit returns
    `ErrExtractLimit`. A zip whose central directory lists too many entries is
    refused before anything is written.
  - Extraction goes into a new hidden folder next to the output (`extractToDir`) and
    is renamed into place only on success, so a limit or any other failure leaves no
    output behind (step 3). Single-file outputs already used temporary files (2.11).
  - The CLI's `decompress` and `decrypt` have `--max-size` (for example `500MB` or
    `20GiB`) and `--max-entries`, and their errors name the flags. The TUI uses the
    defaults and says how to change them from the CLI.
  - Tests that failed against the previous code: `TestDecompressEnforcesSizeLimit`
    (gzip, tar.gz, zip and single-file zip), `TestExtractEnforcesEntryLimit`,
    `TestDecryptDirectoryEnforcesExtractLimits`, `TestFailedExtractionLeavesNoPartialOutput`,
    `TestExtractLimitFlags` and `TestTUILimitHint`, plus `TestDefaultExtractLimits`,
    `TestParseSize` and `TestFormatSize`.
  - With real binaries: a 305 KB gzip wrote 300 MiB and a 620 KB tar.gz created
    100,001 files under the previous build. The new build stopped both (with
    `--max-size 100MB`, and with the default entry limit), left nothing behind, and
    extracted them fully with the limits raised. The TUI showed the hint.
  - gosec no longer reports G110 (it did at four places).
  - CI passed on `b415ffc`; closed 2026-09-27.
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
- **Resolution:** Fixed in `b415ffc` (2026-09-27). Validated by the tests and real-binary checks above, gosec no longer reporting G110, and green CI on `b415ffc` (CI #129, Docker #11, Security #134). Closed 2026-09-27.

### SEC-008 — Extraction follows existing symlinks in the destination and overwrites files

- **Status:** Closed (2026-09-27)
- **Note (2026-10-03):** the `os.Root` confinement depends on the Go version. Builds
  older than 1.26.5 carry GO-2026-4970: a path ending in `/` can follow a symlink out of
  a root. Cryptare's extraction never produces such paths and its extraction roots hold
  no symlinks, so this isn't exploitable today. Release builds must still use a fixed
  Go version (SEC-018). Status unchanged.
- **Progress (2026-09-27, `d751967`; plan 2.2):** steps 1 and 3 are implemented and
  validated, so all remediation steps are done. The owner chose owner-only
  permissions.
  - `extractTarGz` and `extractZipEntries` write through an `os.Root` opened on the
    new extraction folder, so no entry can be created outside it, even through a
    symlink. The lexical `..` check stays, for a clear error.
  - Permissions from the archive are no longer applied: folders are 0700 and files
    0600, or 0700 when the archive marks them executable (`extractDirMode`,
    `extractFileMode`). A read-only folder in an archive therefore no longer stops
    its contents from being extracted (BUG-014).
  - Tests: `TestExtractMasksArchivePermissions` (tar.gz and zip) failed against the
    previous code. It reported the archive's modes as root, and as a non-root user
    extraction failed with "permission denied". `TestEncryptDecryptDirectory` now
    expects 0600 for a restored 0640 file. `TestExtractRejectsPathTraversal` and
    `TestExtractAbsoluteEntryStaysInside` cover path handling, which had no tests.
  - With real binaries as a non-root user with umask 000, the previous build
    extracted a 0777 folder, a 0666 file and a 0775 script unchanged, and failed on a
    0500 folder with contents. The new build produced 0700, 0600 and 0700 (the
    script still runs) and extracted the read-only folder.
  - gosec no longer reports G703 on extraction (three findings).
  - The whole internal test suite also passes as a non-root user.
  - CI passed on `a5edc91`, which includes `d751967`; closed 2026-09-27.
- **Progress (2026-09-27, `b415ffc`; plans 2.1 and 2.2):** extraction now always
  writes into a new, empty folder with mode 0700 (`extractToDir`), which is renamed
  to the output only when extraction succeeds. Archives can't contain symlinks, and
  other users can't write into a 0700 folder, so there is no symlink to follow, even
  with `--force`. An existing output folder is replaced as a whole instead of being
  merged into, which covers step 2. `TestExtractReplacesExistingOutput` (tar.gz and
  zip, symlink in the existing output) failed against the previous code, which wrote
  through the symlink. Remaining: `os.Root` extraction as defence in depth (step 1)
  and masking archive modes (step 3).
- **Progress (2026-09-24, `c47a94f`, v1.1.0):** Part of step 2 is done.
  - `decompress` and `decrypt` now refuse an output path that already exists unless
    `--force` is given, and the TUI always refuses (plan 1.5, `CheckOutputPath`).
  - By default, extraction therefore goes only into a new folder, which can't already
    contain a planted symlink.
  - Remaining: with `--force`, extraction into an existing folder still follows
    symlinks and overwrites files. `os.Root` extraction (step 1), per-entry
    overwrite checks and mode masking (step 3) are still needed.
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
- **Resolution:** Fixed in `b415ffc` (extraction into a new temporary folder, plan 2.1) and `d751967` (`os.Root` and owner-only permissions, plan 2.2). All three remediation steps are done and validated, including the symlink regression test; gosec no longer reports G703 on extraction; CI is green on `a5edc91`. Closed 2026-09-27.

### SEC-009 — Deleted keys remain recoverable from the database file

- **Status:** Closed (2026-09-27)
- **Progress (2026-09-27, `d683739`; plan 2.4):** remediation steps 1, 3 and 4 are
  implemented and validated. Step 2 (`VACUUM`) was considered and not adopted.
  - `NewDatabase` opens SQLite with go-sqlite3's `_secure_delete=on` DSN parameter
    (`withSecureDelete`). The driver applies it to every pooled connection; a single
    `PRAGMA` would only reach one.
  - `TestDeleteKeyWipesBlobFromFile` saves two keys, deletes one, and checks that its
    blob is gone from the database file and from any journal or WAL file. It also
    checks `PRAGMA secure_delete` on three pooled connections at once. It failed
    against the previous code (blob still in the file; setting off).
    `TestWithSecureDelete` covers paths that already carry URI parameters.
  - With real binaries, after `keys delete` the previous build left the blob in
    `cryptare.db` and the new build didn't. No journal or WAL file is left either way:
    the default rollback journal is removed at commit.
  - The README now describes what deletion does and doesn't cover.
  - Why no `VACUUM`: with secure delete on, new deletions are already zeroed. A
    `VACUUM` would only clear keys deleted by older versions, and it builds a
    temporary copy of the whole database, by default in the temp folder (compare
    SEC-006). Offered as an optional follow-up (plan 2.4a).
  - Not covered:
    - keys deleted with v1.1.0 or earlier stay in free pages until SQLite reuses
      them;
    - during the delete, the rollback journal briefly holds the original page. It is
      unlinked, not wiped, at commit, so a filesystem-level forensic tool might still
      find it, as with any deleted file (for example on SSDs);
    - earlier copies (backups, git history, SEC-003) are unaffected.
  - Committed in `d683739`; CI passed on `a5edc91`; closed 2026-09-27.
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
- **Resolution:** Fixed in `d683739` (plan 2.4): SQLite `secure_delete` on every connection, validated by `TestDeleteKeyWipesBlobFromFile`, a real-binary check and green CI on `a5edc91`. Not covered, as recorded above: keys deleted by v1.1.0 or earlier (optional plan 2.4a), filesystem-level remnants of the unlinked journal, and earlier copies. Closed 2026-09-27.

### SEC-010 — Database created world-readable in the current directory on every run

- **Status:** Closed (2026-10-04)
- **Progress (2026-10-04, verified on GitHub):** step 3 is committed in `a088f7f`. CI
  #143 passed on Ubuntu, Windows and macOS, so the default-path tests
  (`TestUserDataDir`, `TestDatabaseOpenerUsesPrivateDataFolder`,
  `TestLegacyDatabaseNotice`) ran on all three; Docker #24, Security #149 and CD #5
  passed, and v1.3.0 shipped the change.
- **Progress (2026-10-04, step 3, plan 3.5; Q-003 answered; not yet committed):** the
  default key store is per user. Without `CRYPTARE_DB_PATH`, `databasePath` uses
  `cryptare/cryptare.db` in the user data folder (`%LocalAppData%` on Windows,
  `~/Library/Application Support` on macOS, `$XDG_DATA_HOME` or `~/.local/share`
  elsewhere); a folder that isn't absolute is an error naming `CRYPTARE_DB_PATH`, never
  the current folder. The opener creates the folder 0700, still only for the `keys`
  commands and the TUI. Owner's migration choice: notice only. A `cryptare.db` in the
  current folder gets a notice on stderr suggesting `mv` (or `move`) when the new store
  doesn't exist, or `CRYPTARE_DB_PATH` when it does; it is never opened. New
  `cryptare keys path` prints the path without creating anything.
  - Tests: `TestDatabasePathDefaultsToDataFolder`, `TestUserDataDir` (10 cases across
    Linux, FreeBSD, macOS and Windows), `TestDatabaseOpenerUsesPrivateDataFolder`
    (folder 0700, file 0600, nothing in the current folder), `TestLegacyDatabaseNotice`
    and `TestQuotePath`. The existing `main` package tests pass with the new opener.
  - Real binaries, umask 022: `keys generate` from the v1.2.0 source (`89a64e8`)
    created `./cryptare.db`; the new build's `--help` and `encrypt` created nothing,
    `keys path` printed the new path and the notice and created nothing, and `keys
    list` created `~/.local/share` and `cryptare/` at 0700 and the store at 0600, left
    `./cryptare.db` byte-identical, and listed the old key once moved as suggested or
    with `CRYPTARE_DB_PATH=./cryptare.db`. With `HOME` unset the commands fail with
    the `CRYPTARE_DB_PATH` hint.
  - Closes once committed and CI passes on Linux, macOS and Windows.
- **Progress (2026-10-04, plan 5.8; not yet committed):** the path gap is fixed.
  `databaseFilePath` works out the file SQLite will open: a plain path, or the
  percent-decoded path of a `file:` URI (read-only `mode=ro` URIs aren't created;
  in-memory ones have no file; Windows `file:` URIs are left to SQLite, as permissions
  aren't checked there). `prepareDatabaseFile` and the export check use it, so a
  database behind a `file:` URI is now created 0600 and trust-checked (SEC-016). A
  plain path containing `?`, which go-sqlite3 cuts there, is refused with
  `ErrUnsupportedDatabasePath` and a hint to use a `file:` URI with `%3F`.
  - Validation: `TestDatabasePathsOpenThePreparedFile` (a `?` folder and a `?_journal_mode`
    suffix refused with nothing created; a `file:` URI to a `?` folder holds the keys
    in a 0600 file with no stray file, and a 0666 file behind a URI is refused; an
    in-memory URI creates nothing) and `TestDatabaseFilePath` (11 cases). Against
    `b3278ea`, the `?` path was accepted and the URI's database was 0644. Real binaries:
    `CRYPTARE_DB_PATH=a?b/k.db` left a stray 0644 file `a` before and is refused after.
  - Remaining: step 3, a per-user default path (Q-003).
- **Gap found (2026-10-03 analysis):** step 2 is incomplete for two kinds of path.
  - A path containing `?`: `prepareDatabaseFile` creates and checks the literal path,
    but go-sqlite3 cuts a non-URI DSN at its first `?` and opens a different file,
    which SQLite creates with the umask's mode. Probe, umask 022:
    `CRYPTARE_DB_PATH=a?b/keys.db` left an empty 0600 `a?b/keys.db` while the keys
    went into a new 0644 file `a`; `keys2.db?_journal_mode=WAL` left an empty 0600
    file of that name and a 0644 `keys2.db` holding the keys.
  - `file:` URIs are skipped by design, so they are also created with the umask's
    mode.

  Fix in plan 5.8: build the DSN so the driver opens exactly the file that was
  prepared, or refuse such paths. A database file owned by someone else is a separate
  issue, SEC-016.
- **Progress (2026-09-27, `d683739`; plan 2.5):** steps 1 and 2 are implemented and
  validated. Step 3, a per-user default path, is still an owner decision (Q-003,
  plan 3.5), so this item stays open.
  - `main.go` passes `databaseOpener()` to `NewRootCmdLazy`. The opener runs at most
    once (`openOnce`, with `sync.OnceValues`) and only from the `keys` commands and
    the TUI. Open errors read "open key database: …".
  - `prepareDatabaseFile` creates a new database with mode 0600 (`O_EXCL`) before
    SQLite opens it, and SQLite gives its journal files the same mode. Owner
    decision: an existing database, or a leftover journal, that others can read is
    set to 0600 when opened. Files owned by someone else are left alone; Windows is
    skipped.
  - Tests that failed against the previous code: `TestRootCmdOpensDatabaseOnlyForKeys`
    (the file commands opened it 6 times), `TestNewDatabaseCreatesPrivateFile` (0644)
    and `TestNewDatabaseTightensExistingFile` (0644 for the database and its
    journal). New: `TestFileCommandsDoNotCreateDatabase` (`main` package, built the way
    `main` builds it) and `TestKeysCmdReportsDatabaseOpenError`. The version test now
    fails if `--version` opens the database.
  - With real binaries under umask 022, the previous build created a 0644
    `cryptare.db` on `--help`, `encrypt` and `compress`, and left an existing 0644
    database unchanged. The new build created nothing for those commands, created the
    database 0600 on `keys list`, and tightened an existing 0644 file to 0600. The TUI
    still opens it (0600), and `compress` works even with an unusable database path.
  - gosec reports one new G304 (`os.OpenFile` on the configured database path in
    `prepareDatabaseFile`). The path comes from the user's own `CRYPTARE_DB_PATH`,
    the same class as the other G304 findings.
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
- **Resolution:** Fixed in `d683739` (steps 1 and 2), `eb330a3` (the `?` and `file:`
  path gap, plan 5.8) and `a088f7f` (step 3, the per-user default path, plan 3.5).
  Validated by the tests and real-binary checks above and by green CI on all three
  systems. Closed 2026-10-04.

### SEC-011 — Imported key metadata not validated before storage and display

- **Status:** Closed (2026-10-04)
- **Note (2026-10-03 analysis):** the "not covered" case below, rows that weren't
  checked on import being printed raw, is reachable without any import: the default
  database is whatever `cryptare.db` the current folder holds, so a planted database
  can inject terminal sequences through `keys list` and the TUI (verified). Escaping
  stored fields on display is tracked as SEC-016 step 2. This item's own remediation
  is unchanged.
- **Note (2026-10-03):** SEC-016 step 2 now escapes stored key IDs and algorithms in
  `keys list` and the TUI, so the "not covered" case below no longer prints raw.
- **Progress (2026-09-27, `3d9384e`; plan 2.7):** all four remediation steps are
  implemented and validated.
  - `ImportKeyFromFile` calls the new `validateKeyExport` before returning a key, so
    the CLI and TUI both get it. It requires:
    - export version 1;
    - a key ID of exactly 16 lower-case hex characters (`isKeyID`, the format
      `newKeyID` makes);
    - algorithm `AES-256-GCM`;
    - an encrypted key that decodes from base64 to exactly 76 bytes (salt, nonce, a
      32-byte key and the GCM tag). Since plan 3.1 (2026-09-27, uncommitted) a
      version 2 stored key is accepted too: 94 bytes, whose header must pass the
      format checks and name a stored key (`validStoredKeyBlob`).

    Anything else returns `ErrInvalidKeyExport`, and rejected values are quoted with
    `%q`, so control characters are escaped rather than printed.
  - Tests: `TestImportKeyRejectsInvalidMetadata` (11 malicious or malformed exports,
    plus a valid one), `TestKeysImportCmdRejectsInvalidExport` (CLI) and
    `TestDashboardImportRejectsInvalidExport` (TUI). All failed against the previous
    code; for example the terminal-escape key ID was accepted as-is. The legacy-export
    tests still pass.
  - With real binaries: importing a `.ckey` whose key ID held `ESC ]0;PWNED BEL ESC
    [31m` succeeded under the previous build, and both the import message and `keys
    list` wrote the raw escape sequences to the terminal. The new build refused it
    with the ID escaped, and stored nothing.
  - Not covered: rows imported before this change are not re-checked. `keys list`
    and the TUI still print what is stored, so remove any suspicious key with `keys
    delete`. The import still can't check that the inner blob decrypts, because it
    is protected by the key's master password (BUG-011).
  - Close after the change is committed and CI passes.
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
- **Resolution:** Fixed in `3d9384e` (plan 2.7). Validated by the import tests and real-binary checks above and by green CI on `3d9384e` (CI #133, Docker #15, Security #138), confirmed on 2026-10-04. The display of rows stored before the fix is covered by SEC-016 step 2 (`4ba516a`). Closed 2026-10-04.

### SEC-012 — CI security-scan results discarded; actions not pinned

- **Status:** Closed (2026-10-04)
- **Note (2026-10-04, round 6; not yet committed):** one more G304 annotation, the same
  disposition as the others: `EncryptedWithStoredKey` in `crypto.go` opens the file the
  user chose to decrypt, to read its header (plan 3.4). gosec v2.29.0 then reports 0
  issues with 10 `#nosec`. Two other findings during the round were fixed instead of
  annotated: G602 on `readV2Header` (rewritten so the slice only shrinks) and G101 on a
  TUI field label whose identifier contained "Pass" (renamed).
- **Progress (2026-10-03, step 5; not yet committed):** the govulncheck job is in
  `security.yml` (SEC-018 step 3).
- **Progress (2026-10-03, dispositions applied in code; not yet committed):** the 10
  open gosec alerts in Code Scanning (#14, #17, #18, #25, #31–#33, #35–#37) are the 10
  findings triaged below.
  - G304 ×8: each call is annotated `// #nosec G304 -- <the user-chosen path>`, as
    step 3 asks. The annotation names only G304, so other rules still apply there.
  - G301 ×2: owner's choice, missing parent folders of an output are now created 0700
    (`extractZipSingleFile`, `extractToDir`), so the README's "everything the tool
    writes is private to you" holds for them too (on Unix; SEC-019 covers Windows).
    New test `TestExtractCreatesPrivateParentFolders` fails before the change and
    passes after it.
  - Validated on a scratch copy (Go 1.26.0, linux/amd64): gosec v2.29.0 with CI's
    arguments reports 0 issues (8 `#nosec`) and writes a SARIF with no results;
    `gofmt -s -l .` and `go vet ./...` are clean; `go test -race -count=1 ./...`
    passes.
  - Remaining: push, and confirm the Security workflow closes the 10 alerts. Code
    Scanning's CodeQL status also warns "Actions workflow file not found" for the
    configurations left by `codeql.yml` (added `58d80ae`, deleted `2691714`); deleting
    them on the tool status page clears the warning. Step 5 (govulncheck) is open.
- **Progress (2026-10-03):** step 3's triage was done locally, and step 5's govulncheck
  was run (results under SEC-018). gosec v2.29.0 reports the same 10 findings as on
  2026-09-27; golangci-lint v2.13.2 reports 0 issues.

  | Rule | Where | Disposition |
  |---|---|---|
  | G304 ×8: file path taken from a variable | `compress.go` `writeGzip`, `DecompressFileWithLimits`, `writeZipFile`; `crypto.go` `encryptSingleFile`, `DecryptFileWithLimits`, `ImportKeyFromFile`; `database.go` `prepareDatabaseFile`; `logic-cli.go` `readPasswordFile` | Accepted by design: each path is one the user chose (a command argument, `--password-file`, `CRYPTARE_DB_PATH` or a TUI field), and a file tool has to open it. Dismiss in Code Scanning as "won't fix", or annotate with `#nosec G304 -- user-selected path`. |
  | G301 ×2: `MkdirAll` with 0755 | `compress.go` `extractZipSingleFile`, `extractToDir` | These create missing *parent* folders of the chosen output path, not archive content (which is 0700/0600). That matches `maint.md` §4, but the README says everything written is private. Either create parents 0700, or narrow the README's claim (owner's choice; plan 5.20). |

  Remaining: apply these dispositions in Code Scanning, and add the govulncheck job
  (SEC-018 step 3).
- **Progress (2026-09-27, W10, delivered as a patch):** every pinned action that ran
  on Node.js 20 moves to its current release on Node.js 24, still pinned to a full
  commit SHA.
  - The SHAs were read from each project's release tags (`git ls-remote`, following
    annotated tags to their commits), and each commit's `action.yml` was checked for
    its runtime and for the inputs the workflows pass.
  - The upgrade takes CodeQL to v4, ahead of v3's deprecation in December 2026.
  - Remaining, as before: step 3 (triage in Code Scanning) and step 5 (govulncheck).
  - The owner committed the patch as `b520b97`; its CI results are not yet reported.
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
- **Resolution:** Fixed in `5286920` (v1.0.1: steps 1, 2 and 4), `b520b97` (W10 action updates), `4ba516a` (step 3: the G304/G301 dispositions) and `6a5fcb1` (step 5: the govulncheck job). Validated on 2026-10-04: Code Scanning shows 0 open alerts (37 closed); every `uses:` line references a commit SHA; Security #144 on `b3278ea` ran CodeQL, gosec and govulncheck successfully (govulncheck: no vulnerabilities reached by Cryptare's code). The "Actions workflow file not found" warning comes from the CodeQL configurations of the deleted `codeql.yml` and needs deleting on the tool status page; it doesn't affect this item. Closed 2026-10-04.

### SEC-013 — Container runs as root; base images not pinned

- **Status:** In Progress
- **Note (2026-10-04, round 6):** the image is now also published to GitHub Packages by
  `cd.yml`'s `container` job, which runs the same UID check (and `--version` and a key
  store round trip) before pushing, so a published image can't run as root.
- **Progress (2026-10-04, round 6; owner approved the workflow change; not yet
  committed):** `docker.yml` gains "Check the image's user", which runs
  `docker run --rm --entrypoint id "$IMAGE_NAME" -u` and fails unless it prints 10001.
  actionlint 1.7.12 reports nothing; Docker isn't available in the analysis environment,
  so the step's first run is the Docker workflow on the next push. This item closes when
  that run passes.
- **Progress (2026-10-04, verified on GitHub):** the changes are committed (`eb330a3`,
  `.dockerignore` in `89a64e8`). Docker #22 and #23 passed: the build resolved both
  pinned digests, and `--help` plus `keys generate`/`keys list` on a new named volume
  worked as the image's user. Remaining validation: `docker run --rm --entrypoint id
  <image> -u` printing 10001, which the workflow log doesn't show. Run it once locally,
  or add `docker run --rm --entrypoint id "$IMAGE_NAME" -u | grep -qx 10001` to
  `docker.yml`'s smoke test (a CI change for the owner to approve).
- **Progress (2026-10-04, plans 4.2 and 5.9; owner approved the `Dockerfile` changes;
  not yet committed):**
  - Step 1: the runtime image creates user and group `cryptare` (10001), gives it
    `/app/data` (0700) and runs as `USER 10001:10001`. A new named volume takes that
    ownership. The README's bind-mount examples add `--user "$(id -u):$(id -g)"`,
    which also satisfies SEC-016's ownership check.
  - Step 2: both images are pinned by the digests the Docker workflow resolved on
    2026-10-04 (`golang:1.26.8-alpine@sha256:8ac98ca5…`,
    `alpine:3.22@sha256:5291449c…`).
  - The additional hardening: a `.dockerignore` (no `.git`, CI and editor folders,
    `intel/`, databases, `.ckey`, encrypted or compressed files, password files); no
    runtime packages (`sqlite-libs` and `ca-certificates` dropped); `ARG VERSION`
    stamps `--version`; `-trimpath`.
  - Not run here: Docker isn't available in the analysis environment. The Docker
    workflow's smoke test (build, `--help`, `keys generate`/`keys list` on a named
    volume) is the first build. Remaining validation: that run passing, and
    `docker run --rm --entrypoint id <image> -u` printing 10001.
- **Additional hardening found (2026-10-03 analysis), for the same approved change:**
  - there is no `.dockerignore`, so `COPY . .` sends the whole working tree into the
    builder stage, including `.git` and any local `*.db`, `*.ckey` or password files.
    Only the binary reaches the final image, but the builder layers and the build
    context hold the rest;
  - the runtime image installs `sqlite-libs` and `ca-certificates`, which Cryptare
    doesn't use: go-sqlite3 compiles SQLite into the binary, and the tool makes no
    network connections;
  - the image isn't version-stamped (`--version` prints `dev`) and the builder's Go
    version floats with `golang:1.26-alpine` (SEC-018).
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

- **Status:** Closed (2026-10-04)
- **Note (2026-10-03):** `walkSourceTree` reaches GO-2026-4970 (fixed in Go 1.26.5) and
  GO-2026-4602 (fixed in 1.26.1). Through `fs.FS` paths, which can't end in `/`, only
  GO-2026-4602's leak of file metadata applies. Building with a current Go (SEC-018)
  removes both.
- **Progress (2026-09-27, `3d9384e`; plan 2.8):** the remediation is implemented and
  validated.
  - The new `walkSourceTree` in `compress.go` walks the tree with `fs.WalkDir` over an
    `os.Root` opened on the source directory, and opens each file through that root.
    A file swapped for a symlink that leads out of the tree can't be opened: the
    root refuses it.
  - Each file is checked again after opening (`visitOpenFile`: still a regular file),
    and its tar or zip header is built from the opened file's own metadata, so the
    header always matches the data read.
  - `writeTarGz` (compress and directory encryption) and `writeZipDirectory` both use
    it. Walk-time rejection of symlinks and special files is unchanged.
  - `TestArchivingIgnoresFileSwappedForSymlink` uses a test hook
    (`testHookBeforeArchiveOpen`) to swap a file for a symlink to an outside file
    just before it is opened. The previous code, with the same hook added, archived
    the outside file for tar.gz, zip and directory encryption with no error. The new
    code refuses and writes no output.
  - Archives from the previous and new builds of the same tree are byte-identical,
    for both tar.gz and zip.
  - gosec no longer reports G122 (or the G304 on that open).
  - Not covered:
    - a swap to a symlink that stays inside the tree is followed, which only
      archives another file from the same tree;
    - a single-file `compress` or `encrypt` still opens its input by path;
    - a file swapped for a FIFO could block the open.
  - Close after the change is committed and CI passes.
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
- **Resolution:** Fixed in `3d9384e` (plan 2.8). Validated by `TestArchivingIgnoresFileSwappedForSymlink`, the byte-identical archive check and gosec no longer reporting G122, and by green CI on `3d9384e` (CI #133, Docker #15, Security #138), confirmed on 2026-10-04. The two Go vulnerabilities noted above are fixed by the go1.26.8 toolchain CI uses since `6a5fcb1` (SEC-018). Closed 2026-10-04.

### SEC-015 — Interrupted decrypt or extraction leaves partial plaintext in hidden temporary files

- **Status:** Closed (2026-10-04)
- **Progress (2026-10-03, plan 5.5; uncommitted):**
  - Step 1: `EncryptFileContext`, `DecryptFileWithLimitsContext`,
    `CompressFileWithFormatContext` and `DecompressFileWithLimitsContext`, with the
    old functions as wrappers. The context is checked before each read
    (`copyContext`, 32 KiB at most), at each archive entry (`extractBudget.addEntry`
    and the source walk), and before the final rename or `Commit`, so cancelling runs
    the existing `Abort`/`RemoveAll` clean-up and leaves an existing output as it was.
    The context is passed as a parameter, never stored (`golang.md`).
  - Step 2: the four file commands run under `runCancellable`: SIGINT, SIGTERM and
    SIGHUP cancel the context, the command prints `interrupted (<signal>); unfinished
    output was removed` without the usage text, and `main` exits with 128 plus the
    signal's number (`InterruptedError`). After the first signal the default handling
    is restored, so a second one ends the process at once. Signals are caught only
    while a file operation runs, so prompts and `keys` commands behave as before.
  - Step 3: in the TUI, `q` and Ctrl+C while an action runs cancel it, show
    "Cancelling…" and quit when it reports back (`actionRunner`, `quit`). However the
    program ends (including Bubble Tea's own SIGINT/SIGTERM handling, and SIGHUP
    through `runCancellable`), the launcher cancels and waits for the running action
    before the process exits.
  - Step 4: README ("Stopping a command") and `maint.md` §4 (Cancellation).
  - With it, BUG-018: the hidden prompt also restores the terminal on SIGTERM,
    SIGQUIT and SIGHUP, exiting with 128 plus the signal's number.
  - Validation (scratch copy, Go 1.26.8 and 1.26.0, linux/amd64, non-root):
    - `TestCancelledOperationsLeaveNothingBehind`: encrypt (file and folder),
      decrypt (file and folder), gzip, tar.gz, zip, gunzip, and tar.gz, zip and
      single-file zip extraction, each stopped partway through by a context that
      reports cancellation after five checks: `context.Canceled`, no output, no
      temporary file or folder, and an existing output left unchanged (22 cases);
    - `TestDecryptInterruptedBySignal` (Unix): cryptare as a subprocess decrypting
      from a named pipe gets SIGINT mid-stream and exits 130 with nothing left;
    - `TestRunCancellableStopsOnSignal`, `TestRunCancellableWithoutSignal`,
      `TestExitCode`, `TestDashboardQuitWhileBusyCancelsAction` (menu `q`, menu and
      form Ctrl+C), `TestDashboardQuitWhenIdle`, `TestActionRunnerShutdownWaitsForAction`;
    - real binaries, input through a named pipe and the signal sent while the hidden
      temporary output existed: the old build was killed by SIGINT, SIGTERM and SIGHUP
      and left `.out.<n>.tmp` (about 2.1 MB of plaintext for a file; a folder for an
      encrypted folder) each time; the new build exited 130, 143 and 129 and left
      nothing;
    - BUG-018 in a pseudo-terminal: the old build restored echo only on SIGINT; the
      new one restored it on SIGINT, SIGTERM, SIGQUIT and SIGHUP (exits 130, 143, 131,
      129).
  - Not run: Windows and macOS (signals there are covered by the build checks only),
    a TUI session driven in a real terminal.
- **Affected component:**
  - `internal/logic-cli.go`: no signal handling outside the password prompt;
  - `internal/logic-tui.go` `Update`: `q` and Ctrl+C quit even while `busy`;
  - `internal/compress.go` `createAtomicFile`/`Abort` and `extractToDir`, and
    `internal/crypto.go` `writeStreamAtomic`/`restoreDirectoryArchive`: clean-up
    runs only through deferred calls.
- **Risk:** Ctrl+C (SIGINT), `kill` (SIGTERM), closing the terminal (SIGHUP), or
  quitting the TUI while an action runs ends the process without the deferred clean-up.
  A cancelled decrypt leaves the plaintext decrypted so far in a hidden
  `.<name>.<random>.tmp` file next to the output. For an encrypted folder it leaves a
  hidden `.<name>.<random>.tmp/` folder.
  - The user saw the command cancelled and believes no plaintext was written. The
    hidden copy stays in synced, backed-up or shared folders until someone finds it.
  - It contradicts the README ("a command that fails part-way leaves no partial
    output"); `maint.md` §4 only expects leftovers after power loss or `kill -9`.
  - Owner-only modes (0600/0700) limit exposure to other local users on Unix, but not
    on Windows (SEC-019).
  - Interrupted encrypt, compress and decompress runs leave partial ciphertext or
    archive data: wasted space, but no plaintext.
- **Evidence (2026-10-03):** real binary built from `0c57aef`.
  - A 3 GiB file was encrypted, then `timeout -s INT 1 cryptare decrypt big.bin.enc`.
    The process ended on SIGINT and `.big.bin.2834428048.tmp` remained: mode 0600,
    652,148,736 bytes of plaintext.
  - A folder holding a 27-byte marker file and 1.5 GB of data was encrypted, then
    `timeout -s INT 0.5 cryptare decrypt tree.enc --output restored`.
    `.restored.164086144.tmp/` (0700) remained, holding the marker file in plaintext
    and 590 MB of the large file.
  - TUI, by code inspection: `Update` returns `tea.Quit` for `q` and Ctrl+C whatever
    `busy` is, and Bubble Tea v1.3.10 doesn't wait for running commands ("Don't wait
    on these goroutines", `tea.go` `handleCommands`), so the process exits mid-action.
- **Required remediation:**
  1. Make file operations cancellable. Pass a `context.Context` (as `golang.md`
     requires for blocking I/O) from the CLI and TUI into `EncryptFile`,
     `DecryptFileWithLimits`, `CompressFileWithFormat` and `DecompressFileWithLimits`,
     and check it between chunks and archive entries. Cancellation then returns an
     error and the existing `Abort`/`RemoveAll` clean-up runs.
  2. CLI: run the file commands under `signal.NotifyContext` for SIGINT, SIGTERM and
     SIGHUP (`os.Interrupt` and `syscall.SIGTERM` on Windows). After clean-up, exit
     with 128 + the signal number (130 for SIGINT), as `readTerminalPassword` does.
  3. TUI: while `busy`, `q` and Ctrl+C cancel the running action and quit only after
     it has reported back, showing "Cancelling…" meanwhile; or ask for confirmation.
  4. Update the README and `maint.md` §4: only `kill -9` and power loss can still
     leave temporary files.
- **Validation:**
  - A core test cancels the context in the middle of a file decrypt and a folder
    decrypt and finds no temporary files (`tempLeftovers`).
  - A CLI test (Unix) sends SIGINT to a subprocess mid-decrypt and checks exit status
    130 and no leftovers.
  - A TUI test quits while busy and checks that the action was cancelled and cleaned
    up.
  - The real-binary checks above leave nothing behind.
- **Resolution:** Fixed in `6a5fcb1`, with its tests in `b3278ea` (plan 5.5). Validated by the tests and real-binary checks above and by green CI on `b3278ea` (CI #138, Docker #20, Security #144), confirmed on 2026-10-04; `TestDecryptInterruptedBySignal` ran on the Linux and macOS jobs. Closed 2026-10-04.

### SEC-016 — Key database files from untrusted locations are trusted

- **Status:** Closed (2026-10-04)
- **Progress (2026-10-04, verified on GitHub):** step 3 is committed in `a088f7f`, with
  CI green on all three systems (see SEC-010) and released in v1.3.0.
- **Progress (2026-10-04, step 3; Q-003 answered; not yet committed):** the key store
  no longer depends on the current folder: without `CRYPTARE_DB_PATH` it is in the
  user's own data folder, created 0700, and a `cryptare.db` in the current folder is
  never opened, only reported (details and validation under SEC-010). Closes with
  SEC-010 once committed and CI passes.
- **Progress (2026-10-04, plan 5.8):** the ownership and permission check now also
  covers a database given as a `file:` URI (it used to be skipped).
- **Progress (2026-10-04):** steps 1 and 2 are committed in `4ba516a`, and CI #136,
  Docker #18 and Security #142 passed on it. Remaining: step 3 (Q-003, plan 3.5).
- **Progress (2026-10-03, steps 1 and 2; not yet committed):** the owner chose to
  refuse rather than warn (Q-010), and to leave step 3, the per-user default path, to
  plan 3.5 (Q-003 stays open).
  - Step 1: `prepareDatabaseFile` calls the new `checkDatabaseFileTrust` for the
    database and any `-journal`, `-wal` or `-shm` file next to it, also when the
    database itself was just created, since SQLite would replay a planted journal into
    it. A file owned by another user (from `fileOwner`, which reads the owner on Unix
    and reports none elsewhere) or with any group or other write bit is refused with
    `ErrUntrustedDatabase` and left unchanged; a private file the user owns that others
    can read is still set to 0600 (SEC-010). `main.go` adds a hint to set
    `CRYPTARE_DB_PATH`. Windows is skipped, as before (SEC-019).
  - Step 2: `displayText` shows a key ID or algorithm that holds a control or other
    non-printable character, or invalid UTF-8, Go-quoted (for example
    `"\x1b[2J"`); other values print unchanged. `keys list` and the TUI key table use
    it.
  - Validation, on a scratch copy (Go 1.26.0, linux/amd64, non-root):
    - new tests `TestCheckDatabaseFileTrust` (synthetic owners and modes: another
      user's file and 0666 and 0620 files refused, the user's private file accepted),
      `TestNewDatabaseRefusesFilesOthersCanWrite` (a 0666 database, and a 0666
      journal next to a new database), `TestDatabaseOpenerRefusesUntrustedDatabase`,
      `TestDisplayText`, `TestKeysListEscapesControlCharacters` and
      `TestKeyScreenEscapesControlCharacters`. With stubs that keep the old
      behaviour, every behavioural one of these fails;
    - real binaries: with the probe row from the evidence below, the old build prints
      3 raw ESC bytes from `keys list` and the new one prints none, showing the values
      quoted. A 0666 database and a 0666 planted journal are refused with exit
      status 1, and the database stays 0666;
    - `gofmt -s -l .` and `go vet ./...` clean; `go vet` and test builds also pass
      for windows, darwin and freebsd (CGO off); `go test -race -count=1 ./...`
      passes, existing database tests included; gosec v2.29.0 reports 0 issues.
  - Not run: the refusal of a file another user owns on a real system (needs a second
    account; covered by the synthetic test), macOS, golangci-lint.
  - Remaining: step 3 (Q-003, plan 3.5).
- **Affected component:** `internal/database.go` `prepareDatabaseFile` and
  `NewDatabase`; `database_path.go` (default `./cryptare.db`); `internal/logic-cli.go`
  `newKeysListCmd`; `internal/logic-tui.go` `View` (key table).
- **Risk:** The default key store is `cryptare.db` in the current folder (SEC-010,
  Q-003), so which file is used depends on where the command runs. A shared or
  attacker-writable folder (such as `/tmp`), a cloned repository (compare SEC-003) or
  an extracted archive can supply it.
  - `prepareDatabaseFile` tightens permissions only on files the user owns. A
    database owned by someone else is left as it is, then opened and written. `keys
    generate` and `keys import` then store the user's encrypted keys in a file that
    another user owns and can read, giving them the blobs to attack offline (SEC-005).
  - Stored rows are printed as they are. `keys list` and the TUI write `KeyID` and
    `Algorithm` raw, so a planted database, or rows from before SEC-011's import
    checks, can inject terminal control sequences: a changed window title, a cleared
    or spoofed screen, and on terminals that honour OSC 52, a write to the clipboard.
- **Evidence (2026-10-03):**
  - Probe test: a row with key ID `ESC ]0;PWNED BEL ESC [31m…` and algorithm
    `ESC [2J` was printed raw by `keys list` (the output contains `0x1b`) and appeared
    raw in the TUI key screen (`View`).
  - The ownership case is from code review: `os.Chmod` failing with
    `fs.ErrPermission` is ignored and the file is opened. It wasn't run, because
    creating a file owned by another user needs a second account.
- **Required remediation:**
  1. On Unix, refuse a database, or its `-journal`, `-wal` or `-shm` file, that the
     current user doesn't own or that is group- or world-writable, as OpenSSH's
     `StrictModes` does. The error gives the reason and suggests `CRYPTARE_DB_PATH`.
     Whether to refuse or only warn is an owner decision (Q-010).
  2. Escape control characters in every stored field that `keys list` and the TUI
     show (for example with `strconv.Quote`, or by replacing non-printable runes).
     This is in addition to SEC-011's import checks.
  3. Prioritise Q-003 / plan 3.5 (a per-user default path), which stops the key store
     depending on the current folder.
- **Validation:**
  - Tests: a row with control characters is listed escaped by the CLI and the TUI.
  - A unit test of the ownership and permission check with synthetic file
    information: another owner's file and a 0666 file are refused, and a private file
    owned by the user is accepted.
  - The existing database tests still pass.
- **Resolution:** Steps 1 and 2 fixed in `4ba516a` (ownership and permission check,
  escaped listings), extended to `file:` URIs in `eb330a3`; step 3 in `a088f7f` (the key
  store no longer depends on the current folder). Validated by the tests listed above
  and green CI. Closed 2026-10-04.

### SEC-017 — GORM's default logger prints SQL with bound values to stdout

- **Status:** Closed (2026-10-04)
- **Progress (2026-10-04, verified on GitHub):** committed in `eb330a3`, with
  `TestKeysCommandsKeepStdoutClean` in `89a64e8`; CI #141 and #142 passed on Ubuntu,
  Windows and macOS, and the change shipped in v1.2.0.
- **Progress (2026-10-04, plan 5.7; not yet committed):** both remediation steps are
  implemented.
  - `NewDatabase` opens GORM with `logger.Discard.LogMode(logger.Silent)`, so nothing
    is logged anywhere, and with `TranslateError`, so a duplicate key ID arrives as
    `gorm.ErrDuplicatedKey`.
  - `GetKey` returns `ErrKeyNotFound` and `SaveKey` returns the new `ErrKeyExists`;
    the CLI and TUI (`keyLookupError`, `keySaveError`) report `key not found:
    "<id>"` and `a key with this ID is already stored: "<id>"`, the ID quoted.
  - Validation: `TestKeysCommandsKeepStdoutClean` runs cryptare as a subprocess (GORM's
    logger holds the process's real stdout, which an in-process test can't swap) and
    checks that `keys export <unknown id>` and a duplicate `keys import` write nothing
    to stdout, name the key and never print the blob; `TestDatabaseErrorsAreTyped`.
    Against `b3278ea`, the first wrote GORM's coloured SQL log (302 bytes) to stdout
    and the second failed with "record not found". Real binaries: the same 302 bytes
    before, 0 after.
  - Close after the change is committed and CI passes.
- **Affected component:** `internal/database.go` `NewDatabase`: `gorm.Open(…,
  &gorm.Config{})` uses GORM's `logger.Default`.
- **Risk:** GORM's default logger writes to **stdout**: failed statements, "record not
  found" lookups, and statements slower than 200 ms. It uses ANSI colours and includes
  the bound values and the build machine's source path.
  - A failed `keys import` (the key ID is already stored) prints the whole `INSERT`,
    with the encrypted key blob. `keys export <unknown id>` prints the `SELECT` with
    the key ID.
  - Encrypted key material therefore ends up in terminal scrollback, CI logs, or files
    that stdout is redirected to.
  - Scripts that parse stdout get unexpected lines, and in the TUI the output
    corrupts the screen.
  - On a slow disk, an ordinary `keys generate` can hit the slow-statement log, which
    prints the blob too.
- **Evidence (2026-10-03):** real binary.
  - `keys export ffffffffffffffff` printed `…/internal/database.go:161 record not found
    … SELECT * FROM key_models WHERE key_id = "ffffffffffffffff" …` on stdout; the
    command's own error went to stderr.
  - Importing a key that was already stored printed `INSERT INTO key_models (…,
    encrypted_blob, …) VALUES (…, "Q1JZUFRBUkUA…")` on stdout.
- **Required remediation:**
  1. Open the database with `gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}`,
     or with a logger that sets `ParameterizedQueries: true` and writes to stderr only
     when explicitly enabled.
  2. Turn the errors that the CLI and TUI show into clear messages, since the log no
     longer explains them. For example: "key … not found" for a missing record, and
     "a key with ID … is already stored" for the UNIQUE constraint.
- **Validation:** CLI tests: `keys export <unknown id>` and a duplicate `keys import`
  write nothing to stdout; the error names the key ID; no output contains the blob.
- **Resolution:** Fixed in `eb330a3` (plan 5.7): a silent GORM logger and typed
  errors. Validated by `TestKeysCommandsKeepStdoutClean` and
  `TestDatabaseErrorsAreTyped` (both failing before the fix) and the real-binary
  comparison above, and by green CI on `89a64e8`. Closed 2026-10-04.

### SEC-018 — Builds use the Go 1.26.0 toolchain, with reachable standard-library vulnerabilities

- **Status:** Closed (2026-10-04)
- **Progress (2026-10-04, round 6):** the remaining validation is done on the published
  v1.3.1 release (tag on `6773ae3`, built by CD). All five assets match `checksums.txt`,
  and `go version -m` on each binary reports `go1.26.8`, `CGO_ENABLED=1`,
  `vcs.revision=6773ae3d5be6…`, `vcs.modified=false` and `golang.org/x/crypto` v0.57.0
  and `golang.org/x/text` v0.42.0 with the hashes in `go.sum`. The Linux binaries are
  static (musl, Q-013); `cryptare_linux_amd64` was run: `--version` prints v1.3.1,
  `keys generate`/`keys list` work, and files encrypted by it and by a build from
  source open in the other. Step 5: `-trimpath` is in the `Dockerfile` but not in the
  CD builds (their binaries embed the build paths; not sensitive), and release
  attestations moved to SEC-020.
- **Progress (2026-10-04, v1.2.0):** CD #4 built and published v1.2.0 from `89a64e8`
  with the `go.mod` toolchain (1.26.8, through `setup-go`, as in CI), and Docker #23
  built with `golang:1.26.8-alpine` pinned by digest, so step 4 is done with SEC-013.
  Remaining: `go version -m` on a v1.2.0 release binary (it needs the asset downloaded,
  which wasn't done here), then this item closes.
- **Progress (2026-10-04, verified on GitHub):** on `b3278ea`, `setup-go` installed
  `go version go1.26.8 linux/amd64` (Security #144), the Docker build resolved
  `golang:1.26.8-alpine@sha256:8ac98ca534ac3f51e1f420a1dd2c15e74c75cfa0f23f3ad27eb5d7236c349a0c`
  (Docker #20), and the govulncheck job reported no vulnerabilities reached by
  Cryptare's code, none in imported packages, and one in a required module that the
  code doesn't call. CI #138, Docker #20 and Security #144 passed. Dependabot ran and
  opened PR #26 (`golang.org/x/crypto` 0.56.0 → 0.57.0), a dependency change for the
  owner to review. Remaining: the validation on release binaries (`go version -m`
  from a CD build), which needs the next release.
- **Progress (2026-10-03, plan 5.1; owner approved the `go.mod`, CI and `Dockerfile`
  changes; uncommitted):**
  - Step 1: `go version -m` on the published v1.1.0 `cryptare_linux_amd64` (archive
    checksum matches `checksums.txt`) reports `go1.26.0`, confirming the risk.
    `govulncheck -mode=binary` wasn't run: the vulnerability database
    (`vuln.go.dev`) isn't reachable from the analysis environment.
  - Step 2: `go.mod` has `toolchain go1.26.8`, the latest 1.26 release (the
    `golang/go` tags and `actions/go-versions` agree). `setup-go` in CI and CD reads
    it through `go-version-file`. `.github/dependabot.yml` proposes `gomod` and
    `github-actions` updates weekly.
  - Step 3: `security.yml` has a `govulncheck` job (`go run
    golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./...` on the `go.mod` toolchain, job
    permissions `contents: read`), which fails when Cryptare's code reaches a known
    vulnerability.
  - Step 4, partly: the `Dockerfile` builder is `golang:1.26.8-alpine`, the same
    release; the tag's existence wasn't checked from here (Docker Hub isn't
    reachable), and pinning by digest stays with SEC-013 (plan 4.2).
  - Validation: the full test suite passes with `-race` on Go 1.26.8; `go vet` and
    test builds pass for linux, windows, darwin and freebsd; actionlint 1.7.12
    reports nothing in the four workflows. govulncheck itself wasn't run here (see
    step 1); the earlier 2026-10-03 run found nothing reachable with go1.26.8.
  - Remaining: push and confirm CI, CD and Docker build with 1.26.8 and the
    govulncheck job passes; `go version -m` on the next release's binaries; the digest
    pin with SEC-013; optionally `-trimpath` and release attestations.
- **Confirmed (2026-10-03, after the owner approved the scanner downloads):**
  govulncheck v1.8.0 (database of 2026-10-01), run with the go1.26.0 toolchain, finds
  three vulnerabilities that Cryptare's code reaches.
  - **GO-2026-4869** (CVE-2026-32288), `archive/tar`, fixed in 1.26.2: unbounded
    allocation for "old GNU sparse" maps, reached through `tar.Reader.Next` in
    `extractTarGz`. **Exploitable:** a crafted `.tar.gz` makes `decompress` allocate
    memory in proportion to its decompressed size before Cryptare sees the entry, so
    SEC-007's `--max-size` and `--max-entries` never apply.
    - Bounded probe, run inside a memory-capped cgroup with
      `--max-size 1MB --max-entries 1`: a 117 KB archive took the go1.26.0 build to
      69 MiB and a 468 KB one to 245 MiB, about 530 times the compressed size.
    - The go1.26.8 build refused both at once ("archive/tar: sparse map too long",
      14 MiB).
    - An earlier 1.9 MB version of the probe, limited only by `ulimit -v`,
      exhausted the analysis machine's memory, and the kernel's OOM killer ended an
      unrelated process.
  - **GO-2026-4970** (CVE-2026-39822), `os`, fixed in 1.26.5: `os.Root` follows a
    symlink out of the root when a path ends in `/`. Reached from `extractTarGz`
    (`os.Root.OpenFile`) and `walkSourceTree`. Not exploitable through today's code:
    entry names are cleaned (no trailing `/`), `fs.FS` paths can't end in `/`, and
    extraction roots contain no symlinks. It is still a latent bypass of SEC-008's and
    SEC-014's confinement.
  - **GO-2026-4602** (CVE-2026-27139), `os`, fixed in 1.26.1: `ReadDir` on a root can
    return metadata (lstat) of files outside it. Reached from `walkSourceTree`;
    metadata only.
  - The other 31 are in packages Cryptare imports but doesn't call (6), or apply only
    at module level (25): x509, TLS, HTTP, `html/template` and similar, plus an
    advisory against `golang.org/x/crypto/openpgp`, which Cryptare doesn't import. The
    highest fix version among all 34 is 1.26.6.
  - With go1.26.8, already cached on the owner's machine, govulncheck reports no
    reachable or imported vulnerabilities.
  - Still not verified: the Go version inside the published release binaries
    (`go version -m` needs a release asset to be downloaded). `setup-go` gives CI and
    CD exactly 1.26.0 for this `go.mod`.
- **Affected component:**
  - `go.mod`: `go 1.26.0` and no `toolchain` directive;
  - `.github/workflows/ci.yml` and `cd.yml`: `actions/setup-go` with
    `go-version-file: go.mod`;
  - `Dockerfile`: `golang:1.26-alpine`, a tag that moves;
  - `security.yml`: no govulncheck.
- **Risk:** setup-go takes the version from the `go` directive, and `1.26.0` names an
  exact release. CI and the release builds therefore compile with Go 1.26.0 and its
  standard library.
  - Go's patch releases regularly fix security issues in packages Cryptare relies on:
    `archive/tar`, `archive/zip`, `compress/*`, `os` (including `os.Root`),
    `crypto/*` and `path/filepath`. The project's own validation used 1.26.8 in
    September 2026.
  - Published binaries keep any such issue until they are rebuilt with a newer
    toolchain.
  - Docker builds use whichever 1.26.x is current, so the pipelines don't agree on the
    toolchain.
  - govulncheck, which would report this, isn't run (SEC-012 step 5).
- **Evidence (2026-10-03):** `go version` in the repository resolves to
  `golang.org/toolchain@v0.0.1-go1.26.0`; `go.mod` has no `toolchain` line; both
  workflows pass `go-version-file: go.mod`. Which vulnerabilities apply was
  established later the same day; see **Confirmed** above.
- **Required remediation:**
  1. Confirm the version with `go version -m` on a published release binary. Run
     `govulncheck ./...` and `govulncheck -mode=binary` on that binary.
  2. Add `toolchain go1.26.<latest>` to `go.mod`, or set `go-version: '1.26.x'` with
     `check-latest: true` in the workflows. Keep it current, for example with
     Dependabot for `gomod` and `github-actions`.
  3. Add a govulncheck job to `security.yml` that fails on reachable vulnerabilities.
  4. Build the Docker image with the same Go version, pinned by digest (with SEC-013).
  5. Optional, in the same change: build with `-trimpath` (binaries embed the build
     machine's paths, as SEC-017's output shows), and add GitHub artifact
     attestations for the release archives.

  Changing CI, `go.mod` or the `Dockerfile` needs owner approval (`AGENTS.md`).
- **Validation:** `go version -m` on new release binaries shows the latest 1.26.x
  release; the govulncheck job runs and passes; CI, CD and Docker report the same Go
  version.
- **Resolution:** Steps 1–4 implemented (`6a5fcb1`, `b3278ea`, `eb330a3`): `toolchain
  go1.26.8`, the govulncheck job, Dependabot, and the digest-pinned `golang:1.26.8-alpine`
  builder. Validated by govulncheck in Security on every push and by `go version -m` on
  the v1.3.1 release binaries (go1.26.8, 2026-10-04). Closed 2026-10-04.

### SEC-019 — Owner-only permission guarantees don't hold on Windows

- **Status:** Closed (2026-10-04)
- **Progress (2026-10-04):** the owner committed the README and `maint.md` wording in
  `eb330a3` and published the same caveat in the v1.2.0 release notes, which completes
  the documented validation (README review).
- **Progress (2026-10-04, plan 5.10):** the owner chose to document the limitation
  rather than set ACLs (Q-012). The README (the output-safety list and the
  `CRYPTARE_DB_PATH` row) and `maint.md` §4 now say that on Windows outputs and the
  key store get the permissions of the folder they are written to, and to keep them
  in a folder only the user can read. Not yet committed; closes once the owner has
  reviewed the wording (the documented validation).
- **Affected component:** every output path (`createAtomicFile`, `extractToDir`,
  `extractDirMode`, `extractFileMode`); `prepareDatabaseFile`, which skips Windows; the
  README ("everything the tool writes is private to you"; the key store is "readable
  only by you").
- **Risk:** On Windows, Go's permission bits only set or clear the read-only
  attribute. Files and folders that Cryptare writes inherit the access-control list of
  the folder they are written to.
  - In the user's own profile that is usually private.
  - Decrypted plaintext written to a shared location (`C:\Users\Public`, a shared
    drive, or a folder an administrator created) is readable by anyone who can read
    that folder. The same applies to the key database.
  - The README promises owner-only output with no caveat, and Windows binaries are
    published.
- **Evidence (2026-10-03):** code review and Go's `os` documentation: on Windows,
  `Chmod` uses only the 0200 bit, and `prepareDatabaseFile` returns early ("Unix
  permission bits don't apply"). Not run on Windows.
- **Required remediation:**
  1. Document the limitation in the README and in `maint.md` §4.
  2. Owner decision (Q-012): set an owner-only DACL on created files and folders on
     Windows with `golang.org/x/sys/windows`. That module is already an indirect
     dependency; making it direct is a `go.mod` change that needs approval.
- **Validation:** README review. If step 2 is done, a Windows CI test reads the DACL of
  an output file and of an extracted folder.
- **Resolution:** Step 1 done (README and `maint.md` §4, `eb330a3`); step 2 declined by
  the owner (Q-012), so Windows outputs keep the folder's permissions by design, as
  documented. Closed 2026-10-04. Reopen if owner-only ACLs are wanted later.

### SEC-020 — Release files have no verifiable proof of origin

- **Status:** In Progress
- **Progress (2026-10-04, round 6; owner approved the workflow change; not yet
  committed):** step 1 is in `cd.yml`: on a tag, the release job runs `actions/attest`
  v4.2.2 (pinned to `1e69f48a…`) on every archive and `checksums.txt`, with
  `id-token: write` and `attestations: write` added to that job only. The README's
  install section shows `gh attestation verify <file> --repo jabbott-iii/Cryptare` and
  says only releases after v1.3.1 have attestations. actionlint 1.7.12 reports nothing.
  Not run: the step itself, which needs a tag push.
- **Progress (2026-10-04, round 6, owner request; not yet committed):** the new
  `container` job in `cd.yml` publishes the Docker image to GitHub Packages on tags
  and attests its digest the same way (`subject-name` and `subject-digest`, with the
  attestation pushed to the registry). The README shows
  `gh attestation verify oci://ghcr.io/jabbott-iii/cryptare:<tag> --repo jabbott-iii/Cryptare`.
  The job holds `packages: write` and uses only plain `docker` commands and the
  already-pinned `actions/attest`, passing the token to `docker login` on stdin.
  Validation adds: the first tagged run pushes the image, and `gh attestation verify`
  passes for its tag.
- **Affected component:** `.github/workflows/cd.yml` (release and container jobs); the README's
  install instructions; every published release up to v1.3.1.
- **Risk:** Releases publish `checksums.txt` next to the archives, from the same
  place. It catches a damaged download, but anyone who can replace an asset (a
  compromised account or token, or a tampered mirror) can replace the checksum file
  too, and users have no independent way to tell a genuine binary from a substituted
  one. For an encryption tool that handles users' passwords and plaintext, a
  substituted binary is a direct compromise. Found in the 2026-10-04 review (also
  noted under SEC-018 step 5 and in `notes.md` §3).
- **Required remediation:**
  1. Attach signed build provenance to every published file (GitHub artifact
     attestations, Sigstore-backed, tied to the repository and the workflow), and
     document how to verify it.
  2. Optional, by audience: Developer ID signing and notarisation of the macOS
     binaries and Authenticode signing of the Windows binary, which operating-system
     warnings rely on (both need paid certificates).
- **Validation:** the CD run for the next tag shows the attestation step succeeding;
  `gh attestation verify` passes for each of that release's archives and
  `checksums.txt`, and fails for a modified copy.
- **Resolution:** —
