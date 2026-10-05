# Security Requirements and Issue Register

This file holds Cryptare's security requirements and a summary record of every security
issue found, with its remediation and status (see `AGENTS.md` → Security Issue
Tracking). The repository is public, so records stay at summary level: no evidence,
reproduction steps, proof-of-concept inputs or attack descriptions, and no pointers to
commits that still hold old key material. Report a suspected vulnerability privately, as
[`SECURITY.md`](../SECURITY.md) describes; details of an open issue stay in a private
GitHub security advisory. Never delete a record. Close one only after its remediation is
implemented and its validation is complete.

Last updated: 2026-10-05 (records condensed to summaries at the owner's request; earlier
versions of this file are in the repository's history)

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

## 2. Controls in place (2026-10-05)

- **Encryption.** New data uses Cryptare's versioned format: AES-256-GCM in authenticated
  64 KiB chunks, with the whole header as additional data and a key per file derived
  with Argon2id from the password (64 MiB, 3 passes, 4 lanes) or with HKDF-SHA256 from a
  stored key. Older formats stay readable (`maint.md` §3).
- **Failures** give a generic error, write no output, and leave an existing output as it
  was; interrupted commands remove their unfinished output.
- **Passwords:** read in full and hidden at a terminal, normalised (NFKC), held to the
  password policy when they protect new data, and kept off the command line
  (`--password-file`).
- **Files:** archiving rejects symlinks and special files and reads through an
  `os.Root`; extraction is confined to a new folder, bounded in size and entries, and
  owner-only; outputs never overwrite their input or, without `--force`, an existing
  file. On Windows, outputs take the folder's permissions (SEC-019).
- **Key store:** per user, created 0600 in a 0700 folder, refused on Unix when another
  user owns it or others can write it, deleted keys overwritten, stored values escaped
  when shown.
- **Supply chain:** actions and base images pinned to SHAs and digests, Go 1.26.8 from
  `go.mod`, govulncheck, gosec and CodeQL in CI, Dependabot for modules, actions and
  images, tracked key material refused by CI, and signed build provenance on release
  files and the container image.

## 3. Assessments

| Date | Scope | Tools | Outcome |
|---|---|---|---|
| 2026-09-23 | All sources, tests, workflows, Dockerfile, git history, v1.0.0 assets | manual review, probe tests, gosec, golangci-lint, `go vet`, `go test -race` | SEC-001 to SEC-014 recorded |
| 2026-10-03 | Sources at `0c57aef` (version 2 format), workflows, Dockerfile, README claims | manual review, probes, golangci-lint, gosec, govulncheck | SEC-015 to SEC-019 recorded; no flaw found in the version 2 construction |
| 2026-10-04 | v1.3.1 and its release assets; the stored-key change (round 6) | manual review, release-asset checks, gosec, golangci-lint, fuzzing, an independent review | SEC-020 recorded; no flaw found in the stored-key format or its cryptography |

## 4. Summary

| ID | Title | Severity | Status |
|---|---|---|---|
| SEC-001 | Empty passwords accepted for encryption and key protection | High | Closed |
| SEC-002 | Interactive password prompt truncates at whitespace and echoes input | High | Closed |
| SEC-003 | Key store committed to the repository | Medium | Closed |
| SEC-004 | Passwords accepted as command-line arguments | Medium | Closed |
| SEC-005 | KDF work factor below current guidance; formats unversioned | Medium | Closed |
| SEC-006 | Directory encryption staged plaintext in the temp folder | Medium | Closed |
| SEC-007 | Unbounded decompression and extraction | Medium | Closed |
| SEC-008 | Extraction could write through existing symlinks in the destination | Low | Closed |
| SEC-009 | Deleted keys recoverable from the database file | Low | Closed |
| SEC-010 | Key database created world-readable in the current folder | Low | Closed |
| SEC-011 | Imported key metadata not validated | Low | Closed |
| SEC-012 | CI security-scan results discarded; actions not pinned | Low | Closed |
| SEC-013 | Container ran as root; base images not pinned | Low | In Progress |
| SEC-014 | Race when archiving a directory tree | Low | Closed |
| SEC-015 | Interrupted operations left partial plaintext behind | Medium | Closed |
| SEC-016 | Key database files from untrusted locations were trusted | Low | Closed |
| SEC-017 | Database logger printed SQL with bound values to stdout | Low | Closed |
| SEC-018 | Builds used an outdated Go toolchain | Medium | Closed |
| SEC-019 | Owner-only permission guarantees don't hold on Windows | Low | Closed |
| SEC-020 | Release files had no verifiable proof of origin | Low | In Progress |

## 5. Issue register

Record format (`AGENTS.md`): severity and status; affected component; remediation; fixed
in (commits and the first release with the fix); validation; and, while open, what
remains.

### SEC-001 — Empty passwords accepted for encryption and key protection

- **Severity:** High · **Status:** Closed (2026-09-27)
- **Affected component:** password checks in `pkg/crypto.go`; the CLI and TUI forms.
- **Remediation:** every path that protects new data with a password applies
  `CheckPasswordPolicy` (15 or more code points, not one repeated character); typed new
  passwords are confirmed; decryption and import still accept older, shorter passwords.
  Since round 6 a stored key's master password must also meet the policy to encrypt new
  files or protect an export, enforced in core.
- **Fixed in:** `ed46150` (v1.1.0), `b415ffc` (v1.2.0); round 6 extension `c114573`.
- **Validation:** core, CLI and TUI tests for empty and weak passwords on every path,
  legacy-password tests, and CI on all three systems.

### SEC-002 — Interactive password prompt truncates at whitespace and echoes input

- **Severity:** High · **Status:** Closed (2026-10-04)
- **Affected component:** the CLI password prompt (`readPassword`).
- **Remediation:** the prompt reads the whole line without echo at a terminal, restores
  the terminal on interrupt, and confirms new passwords; the README tells v1.0.1 users
  how to open files encrypted with the old prompt.
- **Fixed in:** `ed46150` (v1.1.0), `b415ffc`; migration notes in v1.2.0.
- **Validation:** multi-word, non-terminal and pseudo-terminal tests; CI.

### SEC-003 — Key store committed to the repository

- **Severity:** Medium · **Status:** Closed (2026-10-04)
- **Affected component:** a `cryptare.db` once committed to the repository.
- **Remediation:** the file was removed from `main` and the key it held discarded;
  `.gitignore` covers key databases, their side files and key exports; CI refuses any
  that are tracked. The owner chose not to rewrite history (Q-005). The golden test
  fixtures (`pkg/testdata/golden/`) are public test vectors with published
  passwords, kept as `.txt` so the CI guard keeps its meaning for real key material.
- **Fixed in:** `aa27461`, `a088f7f` (v1.3.0).
- **Validation:** the CI guard on Ubuntu, Windows and macOS; no key database or export
  tracked on `main`.

### SEC-004 — Passwords accepted as command-line arguments

- **Severity:** Medium · **Status:** Closed (2026-10-04)
- **Affected component:** `--password` on the encrypt, decrypt and key commands.
- **Remediation:** `--password-file` and piped input; `--password` still works but always
  prints a warning; README and smoke tests use `--password-file`.
- **Fixed in:** `e512844` (v1.2.0).
- **Validation:** tests for the new input and the warning; CI smoke tests.

### SEC-005 — KDF work factor below current guidance; formats unversioned

- **Severity:** Medium · **Status:** Closed (2026-10-04)
- **Affected component:** key derivation and every on-disk format.
- **Remediation:** a versioned header, Argon2id (64 MiB, 3 passes, 4 lanes) with read
  limits, and streaming chunked encryption; legacy formats stay readable.
- **Fixed in:** `0c57aef` (v1.2.0).
- **Validation:** format, tampering and legacy tests; the golden fixtures; CI.

### SEC-006 — Directory encryption staged plaintext in the temp folder

- **Severity:** Medium · **Status:** Closed (2026-09-27)
- **Affected component:** directory encryption.
- **Remediation:** the folder's archive is never written in plaintext; it streams into
  the encrypted output.
- **Fixed in:** `d683739`, then streaming in `0c57aef` (v1.2.0).
- **Validation:** `TestEncryptDirectoryWritesNoTempPlaintext`; CI.

### SEC-007 — Unbounded decompression and extraction

- **Severity:** Medium · **Status:** Closed (2026-09-27)
- **Affected component:** decompression and extraction.
- **Remediation:** limits on output size and entries (default 10 GiB and 100,000),
  `--max-size` and `--max-entries`, and clean-up when a limit is hit.
- **Fixed in:** `b415ffc` (v1.2.0).
- **Validation:** limit tests; gosec G110 resolved; CI.

### SEC-008 — Extraction could write through existing symlinks in the destination

- **Severity:** Low · **Status:** Closed (2026-09-27)
- **Affected component:** archive extraction.
- **Remediation:** extraction into a new temporary folder through an `os.Root`, renamed
  into place on success, with owner-only permissions.
- **Fixed in:** `b415ffc`, `d751967` (v1.2.0).
- **Validation:** regression tests; gosec G703 on extraction resolved; CI.

### SEC-009 — Deleted keys recoverable from the database file

- **Severity:** Low · **Status:** Closed (2026-09-27)
- **Affected component:** the key store.
- **Remediation:** SQLite `secure_delete` on every connection. Keys deleted by v1.1.0 or
  earlier, and earlier copies of the file, aren't covered (optional plan 2.4a declined).
- **Fixed in:** `d683739` (v1.2.0).
- **Validation:** `TestDeleteKeyWipesBlobFromFile`; CI.

### SEC-010 — Key database created world-readable in the current folder

- **Severity:** Low · **Status:** Closed (2026-10-04)
- **Affected component:** the key store's location and creation.
- **Remediation:** the database is opened only by commands that use it, created 0600,
  and kept in a per-user data folder; unsupported database paths are refused.
- **Fixed in:** `d683739`, `eb330a3` (v1.2.0), `a088f7f` (v1.3.0).
- **Validation:** lazy-open, permission and path tests; CI on all three systems.

### SEC-011 — Imported key metadata not validated

- **Severity:** Low · **Status:** Closed (2026-10-04)
- **Affected component:** key import.
- **Remediation:** `validateKeyExport` checks version, key ID, algorithm and blob shape;
  stored values are escaped when shown (SEC-016).
- **Fixed in:** `3d9384e`, `4ba516a` (v1.2.0).
- **Validation:** import tests; fuzzing of export validation; CI.

### SEC-012 — CI security-scan results discarded; actions not pinned

- **Severity:** Low · **Status:** Closed (2026-10-04)
- **Affected component:** the GitHub workflows.
- **Remediation:** every action pinned to a commit SHA; gosec results uploaded to Code
  Scanning and triaged (accepted findings annotated `#nosec <rule> -- <reason>`; 10
  G304 annotations on user-chosen paths); govulncheck in CI.
- **Fixed in:** `5286920` (v1.0.1), `b520b97`, `4ba516a`, `6a5fcb1` (v1.2.0).
- **Validation:** 0 open Code Scanning alerts; gosec 0 issues; CI.

### SEC-013 — Container ran as root; base images not pinned

- **Severity:** Low · **Status:** In Progress
- **Affected component:** `Dockerfile`; the published image.
- **Remediation:** the image runs as UID 10001, which owns `/app/data`; both base images
  are pinned by digest; a `.dockerignore` keeps the build context to the sources;
  `docker.yml` and CD's container job check the UID before an image is used or pushed.
- **Fixed in:** `eb330a3`, `89a64e8` (v1.2.0); UID checks in `c114573`.
- **Validation:** Docker workflow smoke tests passed; the UID check prints 10001.
- **Open:** the UID check's first passing run in CI closes this item.

### SEC-014 — Race when archiving a directory tree

- **Severity:** Low · **Status:** Closed (2026-10-04)
- **Affected component:** archiving (also used by directory encryption).
- **Remediation:** the tree is walked and opened through an `os.Root`.
- **Fixed in:** `3d9384e` (v1.2.0).
- **Validation:** regression test; gosec G122 resolved; CI.

### SEC-015 — Interrupted operations left partial plaintext behind

- **Severity:** Medium · **Status:** Closed (2026-10-04)
- **Affected component:** the CLI and TUI file operations.
- **Remediation:** file operations take a context; signals in the CLI and quitting in
  the TUI cancel them, and their unfinished output is removed before exit.
- **Fixed in:** `6a5fcb1`, `b3278ea` (v1.2.0).
- **Validation:** cancellation tests, a signal test on Linux and macOS; CI.

### SEC-016 — Key database files from untrusted locations were trusted

- **Severity:** Low · **Status:** Closed (2026-10-04)
- **Affected component:** opening the key store; listing keys.
- **Remediation:** on Unix, a database or side file owned by another user or writable by
  others is refused; stored values are escaped when shown; the default store no longer
  depends on the current folder.
- **Fixed in:** `4ba516a`, `eb330a3` (v1.2.0), `a088f7f` (v1.3.0).
- **Validation:** trust-check and escaping tests; CI.

### SEC-017 — Database logger printed SQL with bound values to stdout

- **Severity:** Low · **Status:** Closed (2026-10-04)
- **Affected component:** the key store's database layer.
- **Remediation:** a silent logger and typed errors that name the key.
- **Fixed in:** `eb330a3` (v1.2.0).
- **Validation:** `TestKeysCommandsKeepStdoutClean`, `TestDatabaseErrorsAreTyped`; CI.

### SEC-018 — Builds used an outdated Go toolchain

- **Severity:** Medium · **Status:** Closed (2026-10-04)
- **Affected component:** `go.mod`, the workflows and the `Dockerfile`.
- **Remediation:** `toolchain go1.26.8` in `go.mod`, the same Go in CI, CD and the
  digest-pinned Docker builder, a govulncheck job, and Dependabot.
- **Fixed in:** `6a5fcb1`, `b3278ea`, `eb330a3` (v1.2.0).
- **Validation:** govulncheck on every push; `go version -m` on the v1.3.1 binaries shows
  go1.26.8.

### SEC-019 — Owner-only permission guarantees don't hold on Windows

- **Severity:** Low · **Status:** Closed (2026-10-04)
- **Affected component:** files and the key store written on Windows.
- **Remediation:** documented in the README and `maint.md` §4: outputs take the folder's
  permissions. Owner-only ACLs were declined (Q-012).
- **Fixed in:** `eb330a3` (v1.2.0).
- **Validation:** README review.

### SEC-020 — Release files had no verifiable proof of origin

- **Severity:** Low · **Status:** In Progress
- **Affected component:** `cd.yml` (release and container jobs); releases up to v1.3.1.
- **Remediation:** signed build provenance (`actions/attest`) for every release archive,
  `checksums.txt` and the container image on GitHub Packages, with verification steps
  in the README. Optional, by audience: macOS notarisation and Windows code signing.
- **Fixed in:** `c114573`.
- **Validation:** the next tag's CD run attests its files and image, and
  `gh attestation verify` passes for them and fails for a modified copy.
- **Open:** the first tagged release with attestations.
