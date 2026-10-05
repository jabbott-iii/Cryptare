# Repository History

An append-only record of significant repository changes. Add new entries at the end.
Earlier entries are never edited or removed, except for owner-approved redactions
recorded in this file.

Entry format: `## YYYY-MM-DD — Title`, followed by what changed, why, and references
(commits, PRs, SEC/BUG/Q IDs).

**Redaction (2026-10-05):** at the owner's request, vulnerability details were removed
from the entries below: evidence, reproduction and probe steps, attack descriptions, and
pointers to commits that still hold old key material. Each entry keeps its date, IDs,
outcome and decisions. Earlier versions of this file are in the repository's history.

## 2026-07-25 → 2026-09-23 — Earlier history (reconstructed from git log)

Reconstructed on 2026-09-23 from `git log`: 102 commits on local `main`, 101 on
`origin/main`. Dates are commit dates.

- 2026-07-25 — Initial commit (`62f9ae8`).
- 2026-08-27 — Base CLI structure established (`95a1aea`).
- 2026-08-31 — Full terminal UI merged (PR #8, `f16ec22`).
- 2026-09-01 — `cryptare.db` first committed; see SEC-003.
- 2026-09-12/13 — Single-file directory encryption through authenticated tar.gz
  archives (`af2baaf` and follow-ups). Confirmed key deletion with an atomic
  `DELETE … RETURNING` (`9636e91` and follow-ups).
- 2026-09-13 — Tag `v1.0.0` on `768f4cd`. CD published release assets built with
  `CGO_ENABLED=0`; see BUG-001.
- 2026-09-16 — ZIP added as a first-class format across core, CLI and TUI (`6bb7244`).
- 2026-09-17 — Opt-in vim key bindings for the TUI (PR #16, merged in `00ab703`).
- 2026-09-23 — Local commit `aa27461` ("stash") removes `cryptare.db` from the tree.
  It had not been pushed at the time of writing.

## 2026-09-23 — Repository intelligence baseline

- Created the `intel/` documents required by `AGENTS.md` (`maint.md`, `map.md`,
  `cybersec.md`, `history.md`, `notes.md`, `plan.md`) and recorded the baseline
  (SEC-001…SEC-014, BUG-001…BUG-011, Q-001…Q-007; tests, vet, lint, gosec on 1.26.8).
- Rebuilt the emptied `CONTRIBUTING.md`; `README.md` follows `AGENTS.md`'s section
  order, with corrected release asset names and known-issue notes (BUG-001, SEC-002).
- No code, tests, configuration, CI, dependencies or git history were changed.

## 2026-09-24 — `--version` flag (plan 3.6, BUG-009)

- `var version = "dev"` and a `newRootCmd` helper in package `main`: `cryptare
  --version` / `-v` prints `cryptare version <value>`, so `cd.yml`'s `-ldflags "-X
  main.version=<tag>"` takes effect (resolves W4). New `version_test.go`; no
  dependency changed.
- Docs updated: `README.md`, `intel/maint.md` §6, `map.md`, `notes.md`, `plan.md`.

## 2026-09-24 — CI/CD rework fixes W1–W3 and W5 (delivered as a patch)

- `.github/workflows/` is write-protected for the assistant's remote tools, so the
  owner-approved fixes went out as `cryptare-workflow-fixes.patch`.
- **W1:** the CI, CD and Docker smoke tests run Cryptare: `--version` (in CD it must
  print the tag), `keys generate`/`keys list`, an encrypt → decrypt → `cmp` round trip.
  **W2:** `MUNUS_DB_PATH` → `CRYPTARE_DB_PATH`. **W3:** `munus` names → `cryptare`.
  **W5:** `docker.yml` no longer claims a non-root image (SEC-013, plan 4.2).
- **Validation:** actionlint 1.7.12 clean; the extracted steps pass against real builds
  and catch BUG-001's failure mode; the patch applies. Not run: Windows and macOS
  runners, the Docker workflow, GitHub-hosted runs.

## 2026-09-24 — v1.0.1 released with working binaries (BUG-001 fixed)

- The owner committed the patch as `5286920`, tagged `v1.0.1`; CI, CD, Docker and
  Security succeeded. All five assets match `checksums.txt`, show `CGO_ENABLED=1` and
  `main.version=v1.0.1`, and `cryptare_linux_amd64` passed the smoke checks.
- Closed: BUG-001 and Q-001; plan 0.4, 1.4 and 3.6 done. Still open: Q-007 (v1.0.0's
  assets), W6 (Windows arm64), W7 (darwin/amd64), SEC-012 (gosec triage).
- Docs updated for the new pipeline (plan 4.9): `map.md`, `maint.md` §6,
  `CONTRIBUTING.md`; the README release table waits for `3c80050`.

## 2026-09-24 — TUI no longer crashes on unhandled keys (plan 1.3, BUG-002)

- `internal/logic-tui.go`: the three `panic("unhandled default case")` branches are
  gone; forms ignore unused keys and `buildActionCmd` returns an `unsupported action`
  error. New tests in `logic_tui_test.go` failed before the fix and pass after it.
- Validation (Go 1.26.8): `gofmt -s`, `go mod tidy`, `go vet` and golangci-lint
  v2.13.2 clean; `go test -race ./...` passes; confirmed in a pseudo-terminal.

## 2026-09-24 — Empty passwords rejected for encryption (plan 1.2, SEC-001)

- `internal/crypto.go`: new `ErrEmptyPassword`, returned by `EncryptFile` and
  `EncryptKeyBlob` before any work, so the CLI and TUI inherit it. Decryption still
  accepts an empty password, so older artifacts stay readable.
- New core, TUI and CLI tests; lint and vet clean; `go test -race ./...` passes.
- SEC-001 stays In Progress until the minimum-length policy (Q-004) is decided.

## 2026-09-24 — Password prompt reads the full line, hidden (plan 1.1, SEC-002)

- `readPassword` (`internal/logic-cli.go`) reads one whole line, keeping spaces,
  without echo on a terminal; piped input gives the first line. An empty line is
  rejected for encryption and accepted for decrypting legacy files.
- `github.com/charmbracelet/x/term v0.2.2` moved from indirect to direct (owner
  approved; `go.sum` unchanged). New CLI tests; lint and vet clean; `go test -race
  ./...` passes. README: the new behaviour and a migration note.
- Known regression: Ctrl+C at the hidden prompt leaves echo off (BUG-012, plan 1.1a).
  SEC-002 stays In Progress.

## 2026-09-24 — Terminal restored on Ctrl+C at the password prompt (plan 1.1a, BUG-012)

- New `readTerminalPassword` (`internal/logic-cli.go`): on Ctrl+C during the hidden
  read it restores the terminal and exits 130 (128 + SIGINT); its goroutine ends with
  the read (`golang.md`'s rules). Fixes a problem from plan 1.1 before it shipped.
- Two new tests; lint and vet clean; `go test -race ./...` passes; checked in a
  pseudo-terminal. Not verified: Windows console behaviour. README note removed.

## 2026-09-24 — Outputs never overwrite their input or an existing file by default (plan 1.5, BUG-003, BUG-004, BUG-013)

- Core: `ErrOutputExists`, `ErrSameInputOutput`, `ErrOutputInsideInput` and
  `CheckOutputPath`. The four file operations refuse an output that is their input,
  compared by file identity; compressing a folder into itself is refused (BUG-013).
- CLI: an existing output needs the new `--force` (long form only), checked before the
  prompt. TUI: the same check, with no overwrite option. **Behaviour change:**
  decrypting `x.enc` while `x` exists now stops (README documents it and `--force`).
- New core, CLI and TUI tests failed before; `go test -race ./...` passes (63.1%).
- **Remaining:** atomic writes (2.11), a TUI overwrite choice (2.12), and per-entry
  extraction safety under `--force` (SEC-008).

## 2026-09-26 — v1.1.0 released; owner decisions recorded

- **v1.1.0** was tagged on 2026-09-25 on `c47a94f`, with plan items 1.1, 1.1a, 1.2,
  1.3 and 1.5 (Phase 1 complete). Verified: checksum, `cryptare version v1.1.0`,
  input-as-output refused, `--force` offered.
- **Owner decisions:** plan 0.2 (SEC-003 step 2) decided; v1.0.0 stays as it is
  (plan 0.3; Q-007 closed); stored keys should be usable for file encryption (Q-002 =
  yes; plan 3.4 approved); the password policy (Q-004) is still undecided.

## 2026-09-26 — Atomic writes for single-file outputs (plan 2.11, BUG-004)

- `internal/compress.go`: `atomicFile` (`createAtomicFile`, `Commit`, `Abort`) and
  `writeFileAtomic` write to a hidden 0600 `.<name>.*.tmp` file, sync, then rename.
  Used by compression, single-file gzip decompression, `EncryptFile`, `DecryptFile`
  and `ExportKeyToFile`. Archive extraction moves to plan 2.2.
- Three new tests; `go test -race ./...` passes (63.8%); real binaries leave an
  existing file intact on a failed decompress and leave no temporary files.

## 2026-09-27 — Default password policy (plan 0.5, Q-004; SEC-001, SEC-002)

- **Owner decision (Q-004):** a default policy following NIST SP 800-63B-4 §3.1.1.2:
  at least 15 characters (Unicode code points), no composition rules or maximum
  length, a single repeated character refused. It covers only passwords that protect
  new data (`encrypt`, `keys generate`, `keys export`); decrypt and `keys import`
  accept any password, so older data stays readable.
- New `MinPasswordLength`, `ErrWeakPassword`, `ErrPasswordMismatch` and
  `CheckPasswordPolicy` in core; the CLI asks "Confirm password: " on a terminal; the
  TUI's Encrypt, Generate key and Export key forms gain a confirmation field.
- Eight new tests; `go test -race ./...` passes (67.8%); gosec unchanged at 25
  findings; 7 of 7 pseudo-terminal checks passed (a `b4cd66f` build passed one).
- **W9:** `cryptare-password-policy-workflows.patch` lengthens the smoke-test
  passwords. **Behaviour change:** passwords under 15 characters now fail for
  `encrypt`, `keys generate` and `keys export` (README).

## 2026-09-27 — Extraction limits and atomic extraction (plan 2.1, part of 2.2; SEC-007, SEC-008)

- **Owner decisions:** defaults of 10 GiB of output and 100,000 entries per run;
  `--max-size` and `--max-entries` (0 = no limit); extraction into a temporary folder
  renamed into place; the limits apply to encrypted folders on `decrypt` too.
- New `ExtractLimits`, `DecompressFileWithLimits`, `DecryptFileWithLimits` and
  `ErrExtractLimit`; limit errors name the flag; the TUI uses the defaults.
- **Behaviour changes:** `--force` replaces an existing output folder instead of
  merging; the extracted top folder is 0700 and a single file from a zip 0600; archives
  over the defaults need the flags.
- New tests failed before; `go test -race ./...` passes (71.5%); gosec: 20 findings,
  down from 25; real binaries passed 10 of 10 checks.
- **W9:** applied by the owner, uncommitted alongside 0.5.

## 2026-09-27 — Plan 2.2 completed: `os.Root` extraction and owner-only permissions (SEC-008, BUG-014)

- The owner committed 0.5 and 2.1 as `b415ffc`, with the W9 workflow patch.
- **Owner decision:** extracted output is owner-only: folders 0700, files 0600 (0700
  if marked executable), for `decompress` and decrypted folders.
- Extraction creates entries through an `os.Root`; archive permissions are no longer
  applied. **BUG-014** (new, fixed here): a read-only folder with contents failed to
  extract for non-root users.
- New tests; `go test -race ./...` passes (71.8%), also as non-root; gosec: 11
  findings, down from 20.

## 2026-09-27 — CI green on `b415ffc`; SEC-001 and SEC-007 closed

- The owner committed plan 2.2 as `d751967`. On `b415ffc`, CI #129 (Ubuntu, macOS,
  Windows), Docker #11 and Security #134 (CodeQL, gosec) succeeded.
- Closed: SEC-001 (password policy), SEC-007 (extraction limits). SEC-002 stays open
  for release notes; SEC-008 closes once CI passes on `d751967`.
- Recorded as W10: Node.js 20 actions forced onto Node.js 24; `github/codeql-action`
  v3 deprecated in December 2026; `ubuntu-latest` moves to Ubuntu 26 from 2026-10-19.

## 2026-09-27 — Directory encryption without a plaintext temp file (plan 2.3, SEC-006)

- `buildDirectoryArchive` builds the tar.gz in memory, replacing
  `createDirectoryArchiveTempFile`; the format and memory use are unchanged.
- `TestEncryptDirectoryWritesNoTempPlaintext` failed against `d751967` and passes now;
  `go test -race ./...` passes (71.6%); gosec: 10 findings, down from 11.

## 2026-09-27 — SQLite secure delete for the key store (plan 2.4, SEC-009)

- `NewDatabase` opens the store with go-sqlite3's `_secure_delete=on`
  (`withSecureDelete`) on every pooled connection; the README's key-deletion warning
  is updated.
- Two new tests; `go test -race ./...` passes (71.7%); gosec: 10 findings, unchanged.
- `VACUUM` not adopted; a one-time clean-up for older versions is optional item 2.4a.

## 2026-09-27 — Key database opened only when needed, and created 0600 (plan 2.5, SEC-010, BUG-005)

- **Owner decision:** an existing database others can read is set to 0600 on open.
- `main.go` passes `databaseOpener()` to `newRootCmd(open)`; new `DatabaseOpener`,
  `NewRootCmdLazy` and `openOnce`; only the `keys` commands and the TUI open the store.
  `prepareDatabaseFile` creates it 0600 and tightens it and its side files.
- A CGO-less build now fails only on `keys` and the TUI, so keep the smoke tests'
  `keys generate`/`keys list` steps (`maint.md` §6).
- New tests; `go test -race ./...` passes (72.5%); gosec: 11 findings.
- SEC-010 stays In Progress for step 3, a per-user default path (Q-003, plan 3.5).

## 2026-09-27 — Windows CI failure on `d683739` fixed (test only)

- CI #130 failed on `windows-latest`: `TestFileCommandsDoNotCreateDatabase` left
  `cryptare.db` open, so Windows couldn't remove the temporary folder (`maint.md` §5).
- The test now closes its databases in `t.Cleanup`; no product code changed. Checks
  pass locally; not verified on Windows itself.

## 2026-09-27 — CI green on `a5edc91`; SEC-006, SEC-008 and SEC-009 closed

- The owner committed the fix as `a5edc91`; every workflow is green, the first CI for
  `d751967` (plan 2.2) and `d683739` (plans 2.3–2.5) on all three systems.
- Closed: SEC-006, SEC-008 and SEC-009. Still open: SEC-002 (release notes), SEC-010
  (Q-003) and the rest of the register.

## 2026-09-27 — `--password-file` and a warning for `--password` (plan 2.6, SEC-004)

- **Owner decisions:** add `--password-file` only; `--password` always warns.
- `passwordFlags` adds `--password-file` (first line, at most 64 KiB, spaces kept) to
  `encrypt`, `decrypt` and the `keys generate`, `export` and `import` commands,
  exclusive with `--password`, kept for compatibility. README examples use the file.
- Three new tests; `go test -race ./...` passes (72.9%); gosec: 12 findings.
- **Follow-up (W11):** move the smoke tests to `--password-file`.

## 2026-09-27 — Smoke tests use `--password-file` (W11)

- Owner approved. `cryptare-w11-password-file-smoke.patch` passes a throwaway
  `smoke-pw.txt` to the CI, CD and Docker smoke steps; it needs the 2.6 code, committed
  as `e512844`. actionlint clean; CI and CD steps pass locally; Docker step not run.

## 2026-09-27 — Imported key metadata validated (plan 2.7, SEC-011)

- `ImportKeyFromFile` validates the export (`validateKeyExport`, `isKeyID`,
  `ErrInvalidKeyExport`) before returning. Three new tests; `go test -race ./...`
  passes (73.6%); gosec: 12 findings. README: only Cryptare's export format accepted.

## 2026-09-27 — Folders are archived through `os.Root` (plan 2.8, SEC-014)

- `writeTarGz` and `writeZipDirectory` walk and read the source through an `os.Root`
  (`walkSourceTree`, `visitOpenFile`), with headers from the opened file. New test;
  `go test -race ./...` passes (73.9%); gosec: 10 findings, down from 12; archives are
  byte-identical to the previous build's.

## 2026-09-27 — Case-insensitive extensions and the export file name (plan 2.9, BUG-007, BUG-006)

- **BUG-007:** default outputs and tar.gz detection ignore letter case
  (`hasSuffixFold`, `defaultDecryptOutput`): `FOO.ZIP` → `FOO`, `.TAR.GZ` → a folder.
- **BUG-006:** the default export name is computed once (`defaultExportPath`,
  `timeNow`), so the reported name is the file written. New tests failed before.

## 2026-09-27 — One TUI action at a time (plan 2.10, BUG-008)

- `advanceOrSubmitForm` refuses to submit while `busy` and keeps the form open. New
  test; three TUI tests and the `submit` helper fixed. Validation for 2.7–2.10:
  `go test -race ./...` passes (74.3%); gosec: 10 findings; 7 of 7 terminal checks.
- Phase 2 is complete apart from the optional items 2.4a and 2.12.

## 2026-09-27 — Owner decisions; W10 action updates

- The owner committed 2.7–2.10 and W11 as `3d9384e`. **Decisions:** W10 approved;
  optional items 2.4a (one-time `VACUUM`) and 2.12 (TUI overwrite option) declined.
- **W10** (`cryptare-w10-action-updates.patch`): Node.js 20 actions move to Node.js 24
  releases pinned by full SHA: `actions/checkout` v4.4.0 → v7.0.1, `actions/setup-go`
  v5.6.0 → v7.0.0, `actions/upload-artifact` v4.6.2 → v7.0.1,
  `actions/download-artifact` v4.3.0 → v8.0.1, `codecov/codecov-action` v5.5.5 →
  v7.1.1, `github/codeql-action` v3.38.1 → v4.38.2, `softprops/action-gh-release`
  v2.6.2 → v3.0.3; golangci-lint-action v9.3.0 and gosec unchanged.
- Validated: SHAs match release tags, inputs exist, actionlint clean, patch applies.

## 2026-09-27 — Version 2 format: Argon2id and streaming encryption (plans 3.1 and 3.2)

- The owner committed W10 as `b520b97`. **Decisions (3.1):** Argon2id with 64 MiB, 3
  passes and 4 lanes; built together with the versioned header and chunked streaming
  (3.2); every artifact kind written in the new format; old formats readable with no
  time limit, and no migration command.
- **Format** (`internal/format_v2.go`; `maint.md` §3): a 46-byte header (magic
  `CRYPTARE\0`, version, content type, key source, KDF, Argon2id settings, salt, chunk
  size, nonce prefix), then 64 KiB AES-256-GCM chunks (STREAM, header as additional
  data). Readers take both versions and release output only after the final chunk
  authenticates; settings above 1 GiB, 10 passes or 16 lanes are refused.
- Legacy `deriveKey` moves to `crypto/pbkdf2`; `go.mod` unchanged. Why: SEC-005 and
  BUG-010 (whole files held in memory).
- New format, stream, tampering and header tests; `go test -race ./...` passes (75.4%
  `internal`); gosec unchanged at 10. A 1 GiB file peaked at 77 MiB, against 3,088 MiB
  (encrypt) and 2,063 MiB (decrypt) before; older artifacts read correctly.
- Left uncommitted for the owner's review.

## 2026-10-03 — Repository re-analysis: SEC-015–SEC-019, BUG-015–BUG-024, plan Phase 5

- **Why:** the owner asked for an analysis of bugs and security issues and a plan for
  all of them. **Scope:** every Go source at `0c57aef`, the workflows, `Dockerfile`,
  `.gitignore`, `.devcontainer/` and the README's security claims.
- Plans 3.1 and 3.2 were found committed in `0c57aef`; `plan.md` and SEC-005 corrected.
- **New security items:** SEC-015 (Medium, interrupted decrypt), SEC-016 (Low, key
  database trust and listings), SEC-017 (Low, GORM logging), SEC-018 (Medium, Go
  toolchain; govulncheck), SEC-019 (Low, Windows permissions). Gaps recorded in
  SEC-003, SEC-005, SEC-010, SEC-011 (escaping moves to SEC-016) and SEC-013.
- **New defects:** BUG-015 (non-Latin-1 names), BUG-016 (folder output inside the
  folder), BUG-017 (`keys export` overwrites files), BUG-018 to BUG-024 (lower
  severity); a note on BUG-011. **Questions:** Q-008 to Q-012; Q-003 more urgent.
- **Plan:** Phase 5, four tiers; Tier A (5.1–5.5) first, then a patch release.
- Validation (Go 1.26.0): `gofmt`, `go vet`, `go test -race` and `go mod verify` pass.
- **Changed:** `intel/cybersec.md`, `intel/notes.md`, `intel/plan.md`, this file only.

## 2026-10-03 — gosec Code Scanning alerts resolved (SEC-012, plan 4.1)

- The 10 open gosec alerts (#14, #17, #18, #25, #31–#33, #35–#37) were SEC-012's 10
  findings. G304 ×8: accepted by design, annotated `// #nosec G304 -- <reason>`.
  G301 ×2: owner decision, missing parent folders of an output are created 0700.
- New `TestExtractCreatesPrivateParentFolders`; `maint.md` §4 documents 0700 and
  `#nosec`. gosec v2.29.0: 10 findings before, 0 after (8 `#nosec`); tests pass
  (scratch copy, Go 1.26.0, dependencies from clones of the `go.sum` tags).
- Found: CodeQL warns "Actions workflow file not found" for configurations left by
  `codeql.yml` (added `58d80ae`, deleted `2691714`).
- **Changed:** `internal/` sources, `compress_test.go`, `intel/` docs. Not committed.

## 2026-10-03 — Key database trust and escaped key listings (SEC-016 steps 1–2, plan 5.6)

- **Owner decisions:** Q-010: an untrusted key database is refused, not warned about.
  Moving the default path (Q-003, plan 3.5, step 3) is separate; SEC-016 stays open.
- **Step 1:** `checkDatabaseFileTrust` refuses (`ErrUntrustedDatabase`) a store or
  SQLite side file owned by another user or with a group or other write bit; new
  `internal/fileowner_unix.go` and `fileowner_other.go`; the opener hints at
  `CRYPTARE_DB_PATH`. **Step 2:** `displayText` Go-quotes non-printable key IDs and
  algorithms in `keys list` and the TUI.
- **Behaviour change:** on Linux and macOS the `keys` commands and TUI stop on such a
  database (README upgrade note).
- Six new tests; `go test -race ./...` passes (80.0% `main`, 77.3% `internal`); gosec:
  0 issues (8 `#nosec`). **Changed:** `main.go`, `internal/` sources and tests,
  `README.md`, `intel/` docs, this file. Not committed.

## 2026-10-03 — Phase 5 Tier A: toolchain, export and encrypt outputs, Unicode names, cancellation (plans 5.1–5.5)

- **Owner decisions:** 5.1 approved; the behaviour changes in 5.2 and 5.3 confirmed;
  downloads of Go 1.26.8 and actionlint approved.
- **5.4 / BUG-015:** `gzipHeaderName` keeps Latin-1 names and otherwise stores a
  fallback, so non-Latin-1 names compress and encrypt.
- **5.2 / BUG-017:** `keys export` refuses an existing file without the new `--force`
  (the TUI never overwrites) and never writes over the key database or its SQLite
  files (`ErrOutputIsKeyDatabase`), checked before the password prompt.
- **5.3 / BUG-016:** a folder's default output is `dir.enc` beside it; an output inside
  the folder is refused.
- **5.5 / SEC-015, BUG-018:** cancellable `…Context` operations; an interrupted CLI run
  exits 128 + signal and quitting the TUI cancels the action, leaving nothing behind;
  the prompt restores the terminal on SIGINT, SIGTERM, SIGQUIT and SIGHUP.
- **5.1 / SEC-018:** `toolchain go1.26.8`, `golang:1.26.8-alpine`, a govulncheck
  v1.8.0 job in `security.yml`, `.github/dependabot.yml`.
- Validation (Go 1.26.8): `go test -race ./...` passes (also on 1.26.0); gosec: 0
  issues; actionlint: nothing; real-binary checks passed. **Changed:** `go.mod`,
  `Dockerfile`, workflows, `main.go`, `internal/`, docs. Not committed.

## 2026-10-04 — CI verified; SEC-004/005/011/012/014/015 closed; Tier B, Tier C, 4.2/4.7/4.9, 5.19–5.21

- **Verified on GitHub:** the 2026-10-03 work was committed as `4ba516a`, `6a5fcb1`
  and `b3278ea`; CI #136–#138, Docker #18–#20 and Security #142–#144 passed. On
  `b3278ea`, `setup-go` installed go1.26.8, Docker resolved
  `golang:1.26.8-alpine@sha256:8ac98ca5…` and govulncheck found nothing reachable;
  Code Scanning shows 0 open alerts. Dependabot opened PR #26 (`golang.org/x/crypto`
  0.56.0 → 0.57.0).
- **Closed:** SEC-004 (`e512844`), SEC-005 (`0c57aef`), SEC-011 and SEC-014
  (`3d9384e`), SEC-012, SEC-015 (`6a5fcb1`/`b3278ea`); BUG-015 to BUG-018 resolved.
- `6a5fcb1` turned `security.yml`'s CRLF line endings into LF; left as it is.
- **Owner decisions:** 4.2/5.9 and 5.19 approved; 5.12's behaviour change confirmed;
  Q-011 = `decompress --raw`; Q-012 = document only; 5.15's legacy cap = `--max-size`.
- **Changes:** 5.7 / SEC-017: GORM's logger silenced, `ErrKeyExists`. 5.8 / SEC-010:
  database path handling tightened. 4.2, 5.9 / SEC-013, SEC-003: image runs as UID
  10001, digest-pinned, no runtime packages, `-trimpath`, `.dockerignore`;
  `.gitignore` gains `*.ckey` and SQLite side files. 5.10 / SEC-019: Windows
  permissions documented. 5.11 / BUG-019: containment by file identity. 5.12 /
  BUG-020: contradictory `--format` and out-of-range `--level` refused. 5.13 /
  BUG-021: shorter temporary names. 5.14 / BUG-023: `--password ""` counts as given.
  5.15 / BUG-024: key exports capped at 1 MiB, legacy inputs at `--max-size`. 5.16 /
  BUG-022: `decompress --raw`. 5.19: `-race` in CI. 5.20: docs. 5.21: six fuzz
  targets. 4.7, 4.9: README release section and install steps.
- Validation (Go 1.26.8): `go test -race ./...` passes (100% `main`, 79.4%
  `internal`); gosec: 0 issues (9 `#nosec`); short fuzz runs found nothing; real-binary
  checks passed.
- **Changed:** `.dockerignore`, `.gitignore`, `Dockerfile`, `ci.yml`, `internal/`
  sources and tests, `README.md`, `CONTRIBUTING.md`, `intel/` docs. Not committed.

## 2026-10-04 — v1.2.0 released; Q-003, Q-005, Q-006, Q-008 and Q-009 answered and implemented

- **Verified:** round 3 committed as `eb330a3` and `89a64e8`, tagged v1.2.0 on
  `89a64e8`; CI #141/#142, Docker #22/#23, Security #147/#148 and CD #4 passed.
  Closed: SEC-002, SEC-017, SEC-019; BUG-019 to BUG-024 resolved.
- **Owner decisions:** Q-003: per-user key store, notice-only migration. Q-005: no
  history purge; delete the Copilot branch; add a CI guard. Q-006: rewrite `NOTICE`,
  ship a generated licence file. Q-008: keep the Argon2id read limits (1 GiB, 10
  passes, 16 lanes). Q-009: NFKC, byte order mark handling, `golang.org/x/text` direct.
- **Q-005 / SEC-003:** the `copilot/research-compression-implementation` branch was
  already gone from GitHub; `ci.yml` gains a "Refuse committed key material" step.
- **Q-006 / plan 4.5:** `NOTICE` points to `THIRD_PARTY_LICENSES.txt`, written by new
  `scripts/third-party-licenses.sh` for each release target and packed into every
  release archive with `LICENSE` and `NOTICE`. Found: the static Linux binaries
  contain glibc (LGPL), not covered (Q-013, plan 4.10).
- **Q-009:** new data uses the password without a leading byte order mark, in NFKC
  form; the header's KDF byte is `2` when that differs from the password as given,
  else `1`. Reads try several forms (`passwordCandidates`); a wrong password fails
  before any output. The policy counts the normalised password.
- **Q-003 / plan 3.5:** the default store is `cryptare/cryptare.db` in the user data
  folder, created 0700; a `cryptare.db` in the current folder gets a notice and is
  never opened; new `cryptare keys path`. Plan 4.6: fixture `tasks.db` → `keys.db`.
- Validation (Go 1.26.8): `go test -race ./...` passes; gosec: 0 issues (9 `#nosec`);
  actionlint clean; the licence script wrote 30 modules; the CI guard passes on this
  tree and fails on trees that track key files; data moves between v1.2.0 and the new
  build as designed.
- **Changed:** `ci.yml`, `cd.yml`, `.gitignore`, `NOTICE`, `go.mod`, the licence script,
  `database_path.go`, `main.go`, `internal/` sources and tests, docs. Not committed.

## 2026-10-04 — v1.3.0 released; Q-013 (musl Linux builds, C library notices); `golang.org/x/crypto` 0.57.0

- **Verified:** round 4 committed as `a088f7f`, released as v1.3.0 (CD #5); CI #143,
  Docker #24 and Security #149 passed. Closed: SEC-003 (the CI guard ran on all three
  systems), SEC-010 and SEC-016 (the per-user key store, tested on all three).
- **PR #26:** `golang.org/x/crypto` 0.57.0 needs `golang.org/x/text` 0.42.0 (and
  `x/sys` 0.48.0), which changes NFC/NFKC composition: 104 of 300,000 random strings
  normalised differently, each with an Indic vowel sign or length mark before a
  combining accent, and v0.42.0 matched Python's `unicodedata` in all 300,000.
- **Owner decisions:** Q-013: build Linux releases against musl and add C library
  notices; leave v1.0.1–v1.3.0 as they are. PR #26: take it, with a known-answer test
  and a release note.
- **Q-013 / plan 4.10:** `cd.yml` builds Linux targets statically against musl in the
  pinned `golang:1.26.8-alpine` image; the Windows smoke test fails on a MinGW-w64
  toolchain DLL import; the licence script adds musl's `COPYRIGHT` (1.2.5) and the
  MinGW-w64 runtime notice (v12.0.0), vendored under `scripts/licenses/`.
- **PR #26 / Q-009:** `go.mod`/`go.sum` match PR #26 (blobs `2a8dcd7`, `30a1907`); new
  `TestNormalizePasswordKnownAnswers` pins 16 forms checked against Python.
- Validation (Go 1.26.8): `go test -race ./...` passes; gosec: 0 issues; actionlint
  clean; 30 modules and 2 C libraries listed. Not run: the musl build and DLL check
  (CD only).
- **Changed:** `cd.yml`, `go.mod`, `go.sum`, the licence script and
  `scripts/licenses/`, `internal/password_norm_test.go`, docs. Not committed.

## 2026-10-04 — Round 6: production-readiness review; usable stored keys (plans 3.3, 3.4, 5.17); golden fixtures; SECURITY.md; attestations

- **Verified on GitHub and on the release assets:** round 5 was committed as `6773ae3`
  and released as v1.3.1 by CD. All five v1.3.1 assets match `checksums.txt`, and
  `go version -m` reports go1.26.8, `CGO_ENABLED=1`, `vcs.revision=6773ae3…` and
  `vcs.modified=false` for each, so SEC-018 is closed. `cryptare_linux_amd64` (static,
  musl) runs: `--version`, `keys generate`/`keys list`, and files move both ways
  between it and a build from source.
- **Review findings** (2026-10-04, against v1.3.1): `SECURITY.md` was GitHub's
  template, listing versions 5.1.x and 4.0.x and no reporting channel; stored keys
  could be generated, exported and imported but never used (plan 3.4), and an export
  could need two passwords (BUG-011); no test opened files written by a released
  binary; the README's install example pinned v1.1.0; every runtime error printed the
  whole usage text; release files had no proof of origin (SEC-020); Dependabot didn't
  cover the digest-pinned Docker images; the Docker UID check (SEC-013) wasn't
  automated; `notes.md` and `plan.md` still described round 5 as uncommitted. The
  `FuzzExtractTar` "stall" (plan 5.21) is the fuzzer minimising each new input (60 s
  by default), not a hang; with `-fuzzminimizetime 3s` it ran 60 s, about 24,800
  inputs, without a failure.
- **Owner decisions:** build 3.3 and 3.4 now; Q-014 (the header records the stored
  key's ID); Q-015 (one password per key); approval of the attestation step, the
  Dependabot `docker` entry and the Docker UID check; `SECURITY.md` with a 7-day
  acknowledgement and the latest minor line (1.3.x) supported.
- **Format (plan 3.4, Q-014):** key source 2 with KDF 3: the Argon2id fields are zero,
  the stored key's 8-byte ID follows the 46 common bytes (a 54-byte header, all of it
  each chunk's additional data), and the data key is HKDF-SHA256(stored key, salt,
  "cryptare v2 stored-key data key"). Only files and folders use it. `parseV2Header`
  refuses every combination this version doesn't write. v1.3.1 refuses such files with
  "unsupported encrypted data: key source 2".
- **Code:** `Credential` (password or unlocked stored key) through the encrypting
  writer, decrypting reader and the `…WithCredentialContext` file operations;
  `EncryptedWithStoredKey` reads a regular file's header only, so a piped input isn't
  consumed. New `keys.go` (plan 3.3): `GenerateStoredKey`, `StoredKeyCredential` and
  `ImportStoredKey` (checks the key inside; `ErrSeparateKeyPassword` for an older
  export with its own password), over the `Storage` interface. `ExportKeyToFile`
  checks the master password, then the policy (5.17, BUG-011). CLI: `encrypt
  --key`/`-k`; decrypt finds the key from the header; export asks once for the key's
  master password; import takes an older export's own password through
  `--export-password-file` or a second prompt; delete warns about files encrypted with
  the key; `silenceUsageOnRun` keeps the usage text for command-line mistakes. The TUI
  gains the matching fields.
- **Golden fixtures:** `internal/testdata/golden/` holds files, folders, stored keys and
  exports written by v1.0.1 and v1.3.1 (`linux_amd64` assets checked against their
  `checksums.txt`; archive hashes in its `README.md`), including a KDF 2 file and an
  export with a separate password, plus a stored-key file and folder written by this
  build with v1.3.1's key. Exports and stored keys are `.txt`, not `.ckey`/`.db`
  (SEC-003 note); a `.gitattributes` keeps Windows from converting them.
  `golden_test.go` opens them all. Mutating the nonce's last-chunk flag in both writer
  and reader fails all four golden tests while the round-trip tests still pass.
- **Changed behaviour,** in the README's "Upgrading from v1.3.1 or earlier" note:
  `keys export` takes the key's master password (a separate export password now fails,
  and a key with a short legacy master password can't be exported); import checks the
  key inside; errors no longer print the usage text; `encrypt --key` and decrypting a
  stored-key file open the key store.
- **Workflows (owner approved):** `cd.yml` attests every published file on tags with
  `actions/attest` v4.2.2 (SHA-pinned; `id-token: write` and `attestations: write` on
  the release job only); `docker.yml` fails unless the image runs as UID 10001;
  `dependabot.yml` adds the `docker` ecosystem (golang patch updates only).
- **Docs:** `SECURITY.md` rewritten (plan 4.4); README (features, use case, install from
  `releases/latest`, attestation check, stored keys, upgrade note, configuration);
  `CONTRIBUTING.md`; `intel/cybersec.md` (SEC-001 update, SEC-003 note, SEC-013
  progress, SEC-018 closed, new SEC-020), `intel/maint.md`, `intel/map.md`,
  `intel/notes.md` (BUG-011 resolved; Q-014, Q-015), `intel/plan.md` (Phase 6).
- **Validation** (scratch copy, Go 1.26.8 built from source; `x/crypto`, `x/text`,
  `x/sys` and the two GORM modules from GitHub mirrors; `go.mod`/`go.sum` unchanged):
  `gofmt -s -l .` and `go vet ./...` clean; `go test -race -count=1 ./...` passes
  (89.2% `main`, 80.9% `internal`); golangci-lint v2.13.2 and gosec v2.29.0: 0 issues
  (10 `#nosec`); actionlint 1.7.12: nothing. Seven password tests' export cases and
  TUI tests that step through the encrypt form changed expectations by owner decision.
- **Independent review** (a separate agent, with 30–40 s fuzz runs of
  `FuzzParseV2Header` and `FuzzDecryptingReader`, about 450,000 inputs each): no defect
  in the format or the cryptography. The flow issues it raised were fixed:
  `encrypt --key ""` is refused; `ImportStoredKey` takes an older export's passwords in
  either order, with `--export-password-file` for scripts; the policy on a stored
  key's master password is enforced in core (`canEncrypt`). Also fixed: stale comments
  and docs, and gosec G602 on `readV2Header` (rewritten, not suppressed).
- **Not run:** govulncheck (database unreachable), the Docker build, Windows and macOS,
  and the new workflow steps (they run on the next push and tag).
- **Changed:** `.github/dependabot.yml`, `cd.yml`, `docker.yml`, `SECURITY.md`,
  `README.md`, `CONTRIBUTING.md`, `main.go` (comment), `database_path.go`, `internal/`
  sources (new `internal/keys.go`) and tests (new `golden_test.go`, `keys_test.go`,
  `stored_keys_test.go`, `testdata/golden/`), `intel/` docs and this file. Not
  committed.

## 2026-10-04 — Round 6 (continued): Docker image published to GitHub Packages

- **Owner request:** publish a package on GitHub from CD. The owner chose a container
  image on GitHub Packages over `.deb`/`.rpm` release assets (GitHub Packages has no
  registry for plain binaries).
- **`cd.yml`:** a new `container` job after `release` (so no image is published for a
  failed release). It builds the `Dockerfile` with `VERSION` set to the tag and OCI
  labels (`org.opencontainers.image.source` links the package to the repository;
  revision, version, licence, title, description), smoke-tests the image (`--version`
  carries the tag, `id -u` is 10001, `keys generate`/`keys list` on a volume), then on
  tags logs in with the job's token on stdin, pushes `ghcr.io/<owner>/cryptare:X.Y.Z`,
  plus `X.Y` and `latest` when the tag is the newest release of its line and overall
  (from `git ls-remote --tags`), reads the digest from the push, and attests it with
  `actions/attest` v4.2.2 (`push-to-registry: true`, `create-storage-record: false`).
  Permissions on that job only: `contents: read`, `packages: write`, `id-token: write`,
  `attestations: write`. No new third-party action. A manual CD run builds and
  smoke-tests the image without pushing.
- **Docs:** README (Docker: pulling the published image and checking its attestation),
  `CONTRIBUTING.md`, `intel/maint.md` §6, `intel/map.md`, `intel/cybersec.md` (SEC-020
  progress, SEC-013 note), `intel/plan.md` (6.14 and the one-time step of making the new
  package public), `intel/notes.md`.
- **Validation:** actionlint 1.7.12 with shellcheck 0.11.0: nothing reported for the
  four workflows. The tag logic, run against simulated release lists, gives
  `1.4.0 1.4 latest` for a new release, `1.3.2 1.3` for a patch to an older line after
  1.4.0, only `1.3.1` when 1.3.2 exists, `1.10.0 1.10 latest` after 1.9.0, only the
  pre-release's own tag, `manual-N` for a manual run, and refuses a tag that isn't
  `vX.Y.Z[-PRE]`; the owner name is lower-cased. **Not run:** the job itself (Docker
  isn't available here; it first runs on a manual CD run or the next tag).
- **Changed:** `.github/workflows/cd.yml` (delivered in `cryptare-round6-workflows.patch`
  with the round's other workflow changes), `README.md`, `CONTRIBUTING.md`,
  `intel/cybersec.md`, `intel/maint.md`, `intel/map.md`, `intel/notes.md`,
  `intel/plan.md` and this file. Not committed.

## 2026-10-05 — Round 6 (continued): Makefile targets for contributors

- **Owner request:** a Makefile for full contributor building and testing. It held only
  the release-tagging targets (`tag`, `push-tag`, `release`), which are kept as they
  were.
- **Targets:** `make check` runs what `ci.yml` runs, in its order and failing fast:
  `fmt-check`, `tidy-check` (`go mod tidy -diff`, which changes nothing), `keys-check`
  (the SEC-003 guard), `vet`, `lint`, `test-race`, `smoke` (CI's smoke test on
  `bin/cryptare` in a temporary folder with its own key store, plus a stored-key round
  trip). `make check-all` adds `security` (`sec`: gosec, `vuln`: govulncheck) and
  `actionlint`. Also `build` (`bin/cryptare`, CGO, `-trimpath`, stamped with
  `git describe` through `BUILD_VERSION`, so the release targets keep `VERSION`),
  `test`, `cover`, `cover-html`, `golden`, `fuzz` (`FUZZ`, `FUZZTIME`,
  `-fuzzminimizetime 3s`), `fmt`, `docker-build`, `docker-smoke` (docker.yml's checks:
  UID 10001, version, key store on a throwaway volume), `clean`, and a `help` generated
  from the targets' comments (the default goal).
- **Tools:** golangci-lint v2.13.2, gosec v2.29.0, govulncheck v1.8.0 and actionlint
  v1.7.12 run through `go run pkg@version` (the workflows' versions, checked by the Go
  checksum database, nothing to install); each is a variable that can point at an
  installed binary. GNU Make 3.81 (macOS) features only; bash recipes.
- **Also:** `/bin/` in `.gitignore`, `bin` in `.dockerignore`; `CONTRIBUTING.md`
  (prerequisites, build, a validation table), README "Testing and quality checks",
  `intel/maint.md` §7, `intel/map.md`, `intel/notes.md` §5, `intel/plan.md` (6.15).
- **Validation** (GNU Make 4.3, Go 1.26.8, linux/amd64): `make help`, `build`, `smoke`,
  `test`, `golden`, `fmt-check`, `vet`, `keys-check`, `cover` (81.2%), `cover-html`,
  `fuzz FUZZ=FuzzExtractTar FUZZTIME=5s` and `clean` pass, and the whole `make check`
  chain passes in about a minute (with `tidy-check` skipped, below). `lint`, `sec` and
  `actionlint` pass with the tool variables pointed at the same versions' binaries.
  `fmt-check` fails on an unformatted file and `keys-check` on a tracked `.db`, each
  naming the file. `go mod tidy -diff` was checked on a scratch module: it prints the
  diff, exits 1 and leaves `go.mod` alone when untidy, and exits 0 when tidy.
  **Not run here:** `tidy-check` on this repository and the default `go run` tool
  downloads (the module proxy is unreachable in the analysis environment), `vuln`
  (database unreachable), the `docker-*` targets (no registry access), and GNU Make
  3.81.
- **Changed:** `Makefile`, `.gitignore`, `.dockerignore`, `CONTRIBUTING.md`,
  `README.md`, `intel/maint.md`, `intel/map.md`, `intel/notes.md`, `intel/plan.md` and
  this file. Not committed.

## 2026-10-05 — Security detail removed from the docs; documentation drift fixed

- **Commits recorded:** the owner committed round 6 as `c114573` (which didn't build on
  its own: `internal/keys.go`, the new tests and `internal/testdata/` weren't added) and
  `566f96c` (those files and the Makefile targets), both pushed. `566f96c` builds and
  passes `go test ./...`. Nothing is released yet; v1.3.1 is the latest release.
- **Owner decisions:** keep a summary record for each security item and strip the
  detail (evidence, attack descriptions, reproduction steps and pointers to commits that
  hold old key material) from the public docs; redact past entries of this file the
  same way; change `AGENTS.md` to match.
- **Changes:**
  - `intel/cybersec.md`: from about 1,500 lines to about 320. The requirements stay;
    the stale "existing controls" section becomes the current controls; the assessments
    are a table; each of SEC-001…SEC-020 is a short record (severity, status,
    component, remediation, fixed in, validation, and what remains for SEC-013 and
    SEC-020). Some titles are reworded neutrally; IDs, severities and statuses are
    unchanged.
  - `intel/history.md`: past entries redacted (39 entries kept, with their dates, IDs,
    outcomes and decisions), with a redaction note at the top.
  - `AGENTS.md`: security issues are recorded as summaries; open-issue detail stays in
    a private GitHub security advisory; the record template matches; owner-approved
    redactions of this file are allowed; `intel/golang.md` is listed (plan 4.8).
  - `CONTRIBUTING.md`: security fixes stay at summary level in PRs, commits and docs.
  - Drift: `intel/plan.md` rewritten around the current status, next steps and open
    items, with one line per completed item; `intel/notes.md` §1 and three notes that
    still said "not yet committed"; `intel/maint.md` §5's 2026-09-23 test baseline.
- **Validation:** no key-material commit pointers remain in the docs; every hash,
  version and ID in the redacted history appears in the original; all `## ` headings are
  unchanged; relative links resolve.
- **Not changed:** git history keeps the earlier versions of these files and the
  committed database; rewriting history would need a force-push (the owner decided
  against it, Q-005).
- **Changed:** `AGENTS.md`, `CONTRIBUTING.md`, `intel/cybersec.md`, `intel/history.md`,
  `intel/maint.md`, `intel/notes.md`, `intel/plan.md`. Not committed.
