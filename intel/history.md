# Repository History

An append-only record of significant repository changes. Add new entries at the end,
and never edit or remove earlier entries.

Entry format: `## YYYY-MM-DD — Title`, followed by what changed, why, and references
(commits, PRs, SEC/BUG/Q IDs).

## 2026-07-25 → 2026-09-23 — Earlier history (reconstructed from git log)

Reconstructed on 2026-09-23 from `git log`: 102 commits on local `main`, 101 on
`origin/main`. Dates are commit dates.

- 2026-07-25 — Initial commit (`62f9ae8`).
- 2026-08-27 — Base CLI structure established (`95a1aea`).
- 2026-08-31 — Full terminal UI merged (PR #8, `f16ec22`).
- 2026-09-01 — `cryptare.db` first committed (`d185a94`). The version in `06bda0b`
  (2026-09-06) holds a stored key row; see SEC-003.
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

- Created the `intel/` documents required by `AGENTS.md`: `maint.md`, `map.md`,
  `cybersec.md`, `history.md`, `notes.md`, `plan.md`.
- Recorded the baseline analysis:
  - security items SEC-001…SEC-014;
  - defects BUG-001…BUG-011;
  - owner questions Q-001…Q-007;
  - validation results (tests, vet, lint and gosec all run on Go 1.26.8).
- Rebuilt `CONTRIBUTING.md`, which had been emptied in the working tree. Its earlier
  contribution rules were kept, and setup, validation and PR expectations were added,
  consistent with `maint.md`.
- Updated `README.md` to follow the section order in `AGENTS.md`:
  - added a description, use cases, prerequisites, build-from-source steps,
    configuration, testing and project structure;
  - corrected the release asset names to match `cd.yml`;
  - added a known-issue note about the non-working release binaries (BUG-001) and the
    interactive password prompt (SEC-002).
- No code, tests, configuration, CI, dependencies or git history were changed.

## 2026-09-24 — `--version` flag (plan 3.6, BUG-009)

- Added `var version = "dev"` and a `newRootCmd` helper in package `main`. The helper
  sets Cobra's `Version` field, so `cryptare --version` / `-v` prints
  `cryptare version <value>`.
- This makes the existing `-ldflags "-X main.version=<tag>"` in `cd.yml` take effect.
  It resolves W4 of the CI/CD rework, so the staged smoke tests can call `--version`.
- New test: `version_test.go`.
- Docs updated: `README.md`, `intel/maint.md` §6, `intel/map.md`, `intel/notes.md`
  (BUG-009), `intel/plan.md` (3.6, W4).
- `internal.NewRootCmd` is unchanged, and no dependency changed (Cobra was already a
  direct requirement).

## 2026-09-24 — CI/CD rework fixes W1–W3 and W5 (delivered as a patch)

- **Why a patch:** the owner approved the CI changes, but `.github/workflows/` is
  write-protected for the assistant's remote tools on this machine. The fixes were
  therefore delivered as `cryptare-workflow-fixes.patch`, and the owner applies it.
- **W1:** the smoke tests in `ci.yml`, `cd.yml` and `docker.yml` now run Cryptare
  commands:
  - `--version` (in CD it must print the release tag);
  - `keys generate` followed by `keys list | grep AES-256-GCM`;
  - an encrypt → decrypt → `cmp` round-trip (CI and CD).
- **W2:** `MUNUS_DB_PATH` → `CRYPTARE_DB_PATH`.
- **W3:** `munus` names → `cryptare`: release binaries and archives, the CI binary, the
  Docker image tag and the smoke volume.
- **W5:** the `docker.yml` comment no longer claims the image runs as non-root. The
  image itself is unchanged (SEC-013, plan 4.2).
- **Validation** (analysis environment, Go 1.26.8, linux/amd64):
  - actionlint 1.7.12 is clean; no `munus` references remain.
  - The steps were extracted from the YAML and run against real builds:
    - the CI smoke step passes with CGO and fails with CGO disabled (BUG-001's
      failure mode);
    - the CD build step produces a statically linked `dist/cryptare_linux_amd64`;
    - the CD smoke step passes with tag `v1.0.1` and fails when the version stamp
      doesn't match;
    - release packaging produces `cryptare_<os>_<arch>` archives and `checksums.txt`.
  - The patch applies cleanly to the owner's current workflow files.
- **Not run:** Windows and macOS runners, the Docker workflow (no Docker daemon was
  available), and GitHub-hosted runs.

## 2026-09-24 — v1.0.1 released with working binaries (BUG-001 fixed)

- The owner applied the workflow patch and committed the rework as `5286920`, tagged
  `v1.0.1`. The owner reports that the CI, CD, Docker and Security workflows all
  succeeded.
- The release assets were verified independently:
  - all five assets match `checksums.txt`;
  - each binary's build info shows `CGO_ENABLED=1` and `main.version=v1.0.1`, with
    the expected OS/architecture;
  - none contains go-sqlite3's CGO-less stub message, which the v1.0.0 binary does;
  - `cryptare_linux_amd64` passed `--version`, `keys generate`/`keys list`, and an
    encrypt/decrypt round-trip.
- Closed: BUG-001 and Q-001. Plan 0.4, 1.4 and 3.6 are done.
- Still open:
  - v1.0.0's broken assets remain published (Q-007);
  - Windows arm64 is no longer built (W6);
  - darwin/amd64 hasn't been run (W7);
  - gosec alerts are not yet triaged (SEC-012).
- Docs updated for the new pipeline (plan 4.9): `map.md` CI/CD table, `maint.md` §6,
  and `CONTRIBUTING.md`. The README release table and the known-issue callout wait until
  the owner pulls `3c80050` (a README edit made on GitHub).

## 2026-09-24 — TUI no longer crashes on unhandled keys (plan 1.3, BUG-002)

- `internal/logic-tui.go`: removed the three `panic("unhandled default case")`
  branches.
  - `updateForm` now ignores keys a form doesn't use (arrows, Delete, Home/End, Page
    Up/Down, Ctrl+U, function keys).
  - `Update` no longer has an unreachable panicking default for the Enter key.
  - `buildActionCmd` returns an `unsupported action` error message instead of panicking.
- New tests in `internal/logic_tui_test.go`:
  - `TestDashboardFormIgnoresUnhandledKeys` covers 9 keys in standard and vim modes
    (18 cases);
  - `TestDashboardUnknownActionReportsError`;
  - `TestDashboardEnterOnUnknownScreenIsIgnored`.
- Validation (Go 1.26.8, linux/amd64):
  - the new test panicked before the fix and passes after it;
  - `gofmt -s`, `go mod tidy` (no diff), `go vet` and golangci-lint v2.13.2 are clean;
    `go test -race ./...` passes;
  - end to end, the old binary in a pseudo-terminal crashed with
    `unhandled default case` after Left/Delete/Home/End in the Encrypt form; the fixed
    binary stayed up and exited cleanly on Ctrl+C.

## 2026-09-24 — Empty passwords rejected for encryption (plan 1.2, SEC-001)

- `internal/crypto.go`: new sentinel `ErrEmptyPassword`, following the `ErrKeyNotFound`
  pattern. `EncryptFile` (files and directories) and `EncryptKeyBlob` (key generation,
  and `ExportKeyToFile` through it) return it before doing any work. The check is in the
  core layer, so the CLI and TUI both inherit it (`maint.md` §2).
- Decryption is unchanged and still accepts an empty password, so artifacts created
  before the fix stay readable. Through the TUI's Decrypt form, a blank password
  recovers them.
- New tests:
  - `TestEncryptRejectsEmptyPassword` and `TestDecryptAcceptsLegacyEmptyPassword` in
    `crypto_test.go`;
  - `TestDashboardRejectsEmptyPassword` in `logic_tui_test.go`;
  - `TestEncryptCmdRejectsEmptyInteractivePassword` in `logic_cli_test.go`, which
    guards the prompt rewrite planned in 1.1.
- Validation (Go 1.26.8, linux/amd64):
  - the core and TUI tests failed before the fix and pass after it;
  - `gofmt -s`, `go mod tidy` (no diff), `go vet` and golangci-lint v2.13.2 are clean;
    `go test -race ./...` passes;
  - end to end in a pseudo-terminal: the old binary encrypted a file with a blank
    password; the fixed binary shows "password must not be empty" and writes nothing.
- SEC-001 stays In Progress until the minimum-length policy (Q-004) is decided.

## 2026-09-24 — Password prompt reads the full line, hidden (plan 1.1, SEC-002)

- `internal/logic-cli.go` `readPassword`:
  - it reads one whole line and strips only `\r`/`\n`, so spaces are kept;
  - on a terminal it reads without echo through `github.com/charmbracelet/x/term`;
  - with piped input it reads the first line;
  - an empty line now reaches the core, which rejects it for encryption
    (`ErrEmptyPassword`) and accepts it for decryption, so legacy empty-password files
    can be decrypted from the CLI.
- `go.mod`: `github.com/charmbracelet/x/term v0.2.2` moved from indirect to direct,
  approved by the owner. The version and `go.sum` are unchanged.
- New tests in `logic_cli_test.go`: `TestReadPassword` (7 cases),
  `TestEncryptDecryptCmdMultiWordPromptPassword` and
  `TestDecryptCmdPromptAcceptsEmptyPasswordForLegacyFiles`. They failed before the fix.
- Validation (Go 1.26.8, linux/amd64):
  - `gofmt -s`, `go vet` (also cross-compiled for windows/amd64, darwin/arm64 and
    linux/arm64) and golangci-lint v2.13.2 are clean;
  - `go mod tidy` is stable; `go test -race ./...` passes;
  - in a pseudo-terminal, the fixed binary hides a four-word passphrase, handles a
    Backspace-corrected typo, and the file decrypts with the full passphrase only.
- README: the known-issue note is replaced by the new behaviour, a migration note
  (decrypt old multi-word files with the first word) and the Ctrl+C caveat.
- Known regression: Ctrl+C at the hidden prompt leaves terminal echo off (BUG-012,
  plan 1.1a). SEC-002 stays In Progress.

## 2026-09-24 — Terminal restored on Ctrl+C at the password prompt (plan 1.1a, BUG-012)

- `internal/logic-cli.go`: new `readTerminalPassword`, used by `readPassword` when
  input is a terminal.
  - It saves the terminal state and listens for `os.Interrupt` while the hidden read is
    in progress. On Ctrl+C it restores the terminal, ends the prompt line, and exits
    with status 130 (128 + SIGINT).
  - The helper goroutine is owned by the call and ends when the read returns. This
    meets `golang.md`'s rules for goroutines, which are allowed here because signal
    handling requires one.
- This fixes a problem introduced by plan 1.1 before it shipped: Ctrl+C at the hidden
  prompt killed the process with echo still off.
- New tests in `logic_cli_test.go`: `TestReadPasswordFromNonTerminalFile` (a real,
  non-terminal file takes the line-reading path) and
  `TestReadTerminalPasswordRejectsNonTerminal`.
- Validation (Go 1.26.8, linux/amd64):
  - `gofmt -s`, `go mod tidy` (no diff), `go vet` (native, plus cross-compiled for
    windows/amd64, darwin/arm64 and linux/arm64) and golangci-lint v2.13.2 are clean;
    `go test -race ./...` passes;
  - in a pseudo-terminal, Ctrl+C at the prompt: the build without the fix was killed by
    the signal with echo left off; the fixed build exited 130 with echo restored and
    wrote no file. Normal hidden entry still works.
  - Not verified: Windows console behaviour; CI runs only the non-terminal tests there.
- README: the Ctrl+C known-issue note was removed.

## 2026-09-24 — Outputs never overwrite their input or an existing file by default (plan 1.5, BUG-003, BUG-004, BUG-013)

- **Core (`compress.go`, `crypto.go`):**
  - New sentinels: `ErrOutputExists`, `ErrSameInputOutput` and `ErrOutputInsideInput`.
  - New `CheckOutputPath(src, dst, overwrite)`.
  - `EncryptFile`, `DecryptFile`, `CompressFileWithFormat` and `DecompressFile` always
    refuse an output that is their input. This compares the files themselves, so a
    different spelling, a symlink or a hard link is caught too.
  - Compressing a folder into an archive inside itself is refused (BUG-013, found while
    doing this item).
- **CLI:** `encrypt`, `decrypt`, `compress` and `decompress` resolve the output path
  first. They refuse an existing output unless the new `--force` flag is given (long
  form only, because `-f` is `compress --format`); the error suggests `--force`. The
  check runs before the password prompt.
- **TUI:** the same check without an overwrite option; the error suggests choosing a
  different output path.
- **Behaviour change:** decrypting `x.enc` while `x` still exists now stops instead of
  silently replacing `x`. The README documents this and `--force`.
- **New tests:**
  - core: `TestCheckOutputPath` (7 cases plus symlink),
    `TestOperationsRefuseToOverwriteTheirInput` (6 operations),
    `TestCompressDirectoryRejectsOutputInsideInput`;
  - CLI: `TestFileCmdsRefuseExistingOutput` (4 commands, with and without `--force`,
    plus a default-output case), `TestCompressCmdRefusesSameInputOutputEvenWithForce`;
  - TUI: `TestDashboardRefusesExistingOutput`.

  All failed against the previous behaviour. There, compressing a 4,608-byte file onto
  itself left 33 bytes.
- **Validation** (Go 1.26.8, linux/amd64):
  - `gofmt -s`, `go mod tidy` (no diff), `go vet` (native, plus windows/amd64 and
    darwin/arm64) and golangci-lint v2.13.2 are clean; `go test -race ./...` passes,
    including all existing tests; coverage 63.1%;
  - with real binaries, the old build destroyed its input, replaced an edited
    `note.txt` on decrypt and wrote a self-containing archive, all with exit 0; the new
    build refuses each, and `--force` overwrites when asked.
- **Remaining:** atomic writes (plan 2.11), an optional TUI overwrite choice (2.12), and
  per-entry extraction safety under `--force` (SEC-008).

## 2026-09-26 — v1.1.0 released; owner decisions recorded

- **v1.1.0** was tagged on 2026-09-25 on `c47a94f`. It contains plan items 1.1, 1.1a,
  1.2, 1.3 and 1.5, which completes Phase 1. Verified: the `checksums.txt` entry for
  `cryptare_linux_amd64` matches; the binary reports `cryptare version v1.1.0`, refuses
  an input-as-output, and offers `--force`.
- **Owner decisions:**
  - the key protected by the master password from the committed database is
    discarded (plan 0.2; SEC-003 step 2);
  - v1.0.0 stays as it is now that v1.1.0 is out (plan 0.3; Q-007 closed);
  - stored keys should be usable for file encryption (Q-002 = yes; plan 3.4
    approved);
  - the password policy (Q-004) is still undecided.

## 2026-09-26 — Atomic writes for single-file outputs (plan 2.11, BUG-004)

- `internal/compress.go`: new `atomicFile` helpers — `createAtomicFile`, `Commit`,
  `Abort` — and `writeFileAtomic`.
  - Output is written to a hidden `.<name>.*.tmp` file (mode 0600) in the destination
    folder, synced, then renamed over the destination.
  - Compression now writes through this. The gzip body moved into a new `writeGzip`
    helper so that the stream is finalised before the rename.
  - Single-file gzip decompression uses it too.
- `internal/crypto.go`: `EncryptFile` (file and directory artifacts), `DecryptFile`
  (single file) and `ExportKeyToFile` use `writeFileAtomic` instead of `os.WriteFile`.
- Archive extraction (tar.gz and zip, including a single-file zip) still writes in
  place, keeping the archive's permissions. It moves to plan 2.2.
- **New tests:** `TestFailedDecompressLeavesNoPartialOutput`,
  `TestFailedCompressLeavesNoPartialOutput` (gzip and zip) and `TestWriteFileAtomic`.
  The first two failed before the change, leaving partial outputs behind.
- **Validation** (Go 1.26.8, linux/amd64):
  - `gofmt -s`, `go mod tidy` (no diff), `go vet` (native, plus windows/amd64 and
    darwin/arm64) and golangci-lint v2.13.2 are clean; `go test -race ./...` passes;
    coverage 63.8%;
  - with real binaries, a failed `decompress --force` of a truncated gzip replaced an
    existing file with 49,961 bytes of partial data under v1.1.0; the new build leaves
    it intact. Round trips still work, outputs are mode 0600, and no temporary files
    are left.

## 2026-09-27 — Default password policy (plan 0.5, Q-004; SEC-001, SEC-002)

- **Owner decision:** use a default password policy (Q-004). The policy follows NIST
  SP 800-63B-4 §3.1.1.2 for single-factor passwords:
  - at least 15 characters, counted as Unicode code points;
  - no composition rules and no maximum length;
  - a single repeated character is refused.
- **Scope:** only passwords that protect new data, that is `encrypt`, the
  `keys generate` master password and the `keys export` password. Decrypt and
  `keys import` accept any password, so older files, key blobs and exports stay
  readable.
- `internal/crypto.go`: new `MinPasswordLength`, `ErrWeakPassword`,
  `ErrPasswordMismatch` and `CheckPasswordPolicy`. `EncryptFile` and `EncryptKeyBlob`
  (and so `ExportKeyToFile`) call it in place of the empty-password check.
  `ErrEmptyPassword` is still returned for an empty password.
- `internal/logic-cli.go`: `readNewPassword`/`readNewPasswordWith` check the policy,
  then ask "Confirm password: " when the input is a terminal. Piped input is read
  once. The new `terminalInput` helper is shared with `readPassword`. `encrypt`,
  `keys generate` and `keys export` use it; `--password` values are checked by core.
- `internal/logic-tui.go`: the Encrypt, Generate key and Export key forms have a masked
  "Confirm password" field. `checkTUINewPassword` checks the policy and then the
  match. The Decrypt and Import forms are unchanged.
- **Tests:**
  - new: `TestCheckPasswordPolicy`, `TestEncryptRejectsWeakPassword`,
    `TestDecryptAcceptsLegacyShortPassword`, `TestReadNewPasswordWith`,
    `TestNewPasswordCmdsRejectWeakPassword`, `TestLegacyShortPasswordCmds`,
    `TestDashboardNewPasswordFormsHaveConfirmation` and
    `TestDashboardRejectsWeakOrMismatchedPassword`. Apart from the two legacy tests,
    which guard compatibility, all failed against the previous code;
  - existing tests that encrypted with short passwords now use a compliant one, and
    TUI tests fill in the confirmation field.
- **Validation** (Go 1.26.8, linux/amd64):
  - `gofmt -s`, `go mod tidy` (no diff), `go vet` (native, plus linux/arm64,
    darwin/arm64, darwin/amd64 and windows/amd64) and golangci-lint v2.13.2 are
    clean;
  - `go test -race ./...` passes, with 67.8% total coverage;
  - gosec v2.29.0 reports the same 25 findings as before the change;
  - in a pseudo-terminal the new binary passed 7 of 7 end-to-end checks (confirmation,
    mismatch, weak password, Ctrl+C at confirmation, piped input, TUI mismatch then
    match). A build of `b4cd66f` passed only the piped-input check.
- **Workflows (W9):** the CI, CD and Docker smoke tests passed 8–13 character
  passwords, which the policy rejects. `cryptare-password-policy-workflows.patch`
  lengthens them. With the new code, the CI smoke step passed with the patched
  workflow and failed with the original. The CD and Docker steps weren't run here.
- **Behaviour change:** scripts that pass a password shorter than 15 characters to
  `encrypt`, `keys generate` or `keys export` now fail. The README documents this.

## 2026-09-27 — Extraction limits and atomic extraction (plan 2.1, part of 2.2; SEC-007, SEC-008)

- **Owner decisions:**
  - default limits of 10 GiB of output and 100,000 entries per run;
  - `--max-size` and `--max-entries` flags, where 0 means no limit;
  - partial output is cleaned up by extracting into a temporary folder and renaming
    it into place;
  - the limits apply to encrypted folders on `decrypt` as well.
- `internal/compress.go`:
  - new `ExtractLimits`, `DefaultExtractLimits`, `DefaultMaxExtractBytes`,
    `DefaultMaxExtractEntries`, `ErrExtractLimit`, `ErrInputInsideOutput` and
    `DecompressFileWithLimits`. `DecompressFile` uses the defaults;
  - `extractBudget` counts entries and copies through `io.CopyN`. A zip with too many
    entries in its central directory is refused up front;
  - `extractToDir` extracts into a new hidden 0700 folder next to the output and
    renames it into place. An existing output (`--force`) is moved aside, replaced
    and removed, so it is replaced rather than merged into. Replacing a folder that
    holds the archive is refused;
  - a single-file zip is extracted through `createAtomicFile` (0600);
  - `pathWithin` is shared with `checkOutputOutsideDir`.
- `internal/crypto.go`: new `DecryptFileWithLimits`. `restoreDirectoryArchive`
  extracts through `extractToDir` under the limits.
- `internal/logic-cli.go`: `--max-size` (parsed by `parseSize`, which accepts B, KB–TB
  and KiB–TiB) and `--max-entries` on `decompress` and `decrypt`, validated before
  anything else runs. Limit errors say which flag to use (`withLimitHint`).
- `internal/logic-tui.go`: decrypt and decompress errors from a limit explain that the
  TUI uses the defaults (`withTUILimitHint`).
- **Behaviour changes:**
  - with `--force`, an existing output folder is replaced, not merged into;
  - the top folder of an extraction has mode 0700, where it was 0755 minus the umask;
  - a single file extracted from a zip has mode 0600, where it had the archive's mode;
  - archives over the default limits need the flags.
- **Tests** that failed against the previous code:
  - `TestDecompressEnforcesSizeLimit` (gzip, tar.gz, zip, single-file zip);
  - `TestExtractEnforcesEntryLimit`, `TestDefaultExtractLimits`;
  - `TestFailedExtractionLeavesNoPartialOutput`, `TestExtractReplacesExistingOutput`
    (also the SEC-008 symlink case), `TestExtractRefusesToReplaceFolderHoldingInput`;
  - `TestDecryptDirectoryEnforcesExtractLimits`, `TestFormatSize`;
  - CLI: `TestParseSize` and `TestExtractLimitFlags`;
  - TUI: `TestTUILimitHint`.
- **Validation** (Go 1.26.8, linux/amd64):
  - `gofmt -s`, `go mod tidy` (no diff), `go vet` (native, plus linux/arm64,
    darwin/arm64, darwin/amd64 and windows/amd64) and golangci-lint v2.13.2 are clean;
  - `go test -race ./...` passes, with 71.5% total coverage;
  - gosec v2.29.0: 20 findings, down from 25. All four G110 (decompression bomb)
    findings are gone, and no new rule fires;
  - real binaries, 10 of 10 checks passed. A 305 KB gzip bomb and a 620 KB tar.gz
    with 100,001 entries were stopped and cleaned up, and extracted fully with the
    limits raised. A 300 MiB gzip passed under the defaults. `--force` replaced an
    existing folder and left it untouched when the limit was hit. The extracted
    folder was 0700, and an invalid `--max-size` was refused. The TUI reported the
    entry limit with the hint. The previous build wrote all 300 MiB and all 100,001
    files.
- **W9:** the owner applied `cryptare-password-policy-workflows.patch`; the three
  workflows on disk match the patched versions. It is uncommitted, alongside 0.5.

## 2026-09-27 — Plan 2.2 completed: `os.Root` extraction and owner-only permissions (SEC-008, BUG-014)

- The owner committed 0.5 and 2.1 as `b415ffc`, with the W9 workflow patch. CI was
  still running when last checked.
- **Owner decision:** extracted files and folders are owner-only. Folders are 0700
  and files 0600, or 0700 when the archive marks them executable. This applies to
  `decompress` and to decrypted folders.
- `internal/compress.go`:
  - `extractTarGz` and `extractZipEntries` open an `os.Root` on the new extraction
    folder and create folders and files through it, using the entry's path relative
    to the output. The lexical `..` check stays, and an absolute entry name is still
    extracted inside the output;
  - new `extractDirMode` (0700) and `extractFileMode`. Archive permissions are no
    longer applied;
  - the two functions no longer create the output folder, since `extractToDir`
    always passes one that exists.
- **BUG-014 (new, fixed in the same change):** as a non-root user, an archive with a
  read-only folder that has contents failed with "permission denied". Running as
  root hides this, which is why it hadn't been seen.
- **Tests:**
  - `TestExtractMasksArchivePermissions` (tar.gz and zip, hand-built archives with
    0777, 0666, 0775, 0500 and 0444 entries) failed against `b415ffc`, both as root
    (wrong modes) and as a non-root user (permission denied);
  - `TestEncryptDecryptDirectory` now expects a restored 0640 file to be 0600;
  - new `TestExtractRejectsPathTraversal` and `TestExtractAbsoluteEntryStaysInside`
    cover path handling, which had no tests (part of 4.3). They pass against both
    versions.
- **Validation** (Go 1.26.8, linux/amd64):
  - `gofmt -s`, `go mod tidy` (no diff), `go vet` (native plus four other targets)
    and golangci-lint v2.13.2 are clean;
  - `go test -race ./...` passes, with 71.8% total coverage. The internal test suite
    also passes when run as a non-root user;
  - gosec v2.29.0: 11 findings, down from 20. The G703 path-traversal findings on
    extraction are gone, as are four of six G301 and two G304;
  - real binaries as a non-root user with umask 000: `b415ffc` extracted a 0777
    folder, a 0666 file and a 0775 script unchanged and failed on a 0500 folder. The
    new build produced 0700, 0600 and 0700, the script still ran, and the read-only
    folder was extracted.

## 2026-09-27 — CI green on `b415ffc`; SEC-001 and SEC-007 closed

- The owner committed plan 2.2 as `d751967`.
- GitHub Actions on `b415ffc` all succeeded: CI #129 (Ubuntu, macOS and Windows,
  including the smoke tests with the lengthened passwords), Docker #11 and
  Security #134 (CodeQL and gosec).
- SEC-001 (password policy) and SEC-007 (extraction limits) are closed. SEC-002 stays
  open for its release-notes item. SEC-008 closes once CI passes on `d751967`.
- The run annotations show upcoming CI maintenance, recorded as plan W10:
  - actions that target Node.js 20 are being forced onto Node.js 24;
  - `github/codeql-action` v3 is deprecated in December 2026;
  - `ubuntu-latest` moves to Ubuntu 26 from 2026-10-19.

## 2026-09-27 — Directory encryption without a plaintext temp file (plan 2.3, SEC-006)

- `internal/crypto.go`: `buildDirectoryArchive` builds the directory's tar.gz in a
  `bytes.Buffer`. It replaces `createDirectoryArchiveTempFile`, which wrote a
  plaintext `cryptare-dir-*.tar.gz` to `$TMPDIR` and read it back. The archive and the
  encrypted-folder format are unchanged, and memory use is about the same, because
  the old code also read the whole archive into memory.
- **Test:** `TestEncryptDirectoryWritesNoTempPlaintext` sets `TMPDIR`, `TMP` and `TEMP`
  to a missing folder. It failed against `d751967` and passes now.
- **Validation** (Go 1.26.8, linux/amd64):
  - `gofmt -s`, `go mod tidy` (no diff), `go vet` (native plus four other targets)
    and golangci-lint v2.13.2 are clean;
  - `go test -race ./...` passes, with 71.6% total coverage;
  - gosec: 10 findings, down from 11;
  - `strace` on real binaries: `d751967` opened `$TMPDIR/cryptare-dir-….tar.gz` for
    writing and the new build didn't. Each build decrypted the other's artifact to an
    identical tree. With a missing `TMPDIR` the old build failed and the new one
    succeeded.

## 2026-09-27 — SQLite secure delete for the key store (plan 2.4, SEC-009)

- `internal/database.go`: `NewDatabase` opens the database through `withSecureDelete`,
  which adds go-sqlite3's `_secure_delete=on` parameter (with `&` when the path
  already has URI parameters). The driver applies it to every pooled connection.
- `README.md`: the key-deletion warning now says what is overwritten and what isn't
  (earlier copies, and keys deleted by v1.1.0 or earlier).
- **Tests:** `TestDeleteKeyWipesBlobFromFile` failed against the previous code (the
  deleted blob was still in the file and the setting was off) and passes now. New:
  `TestWithSecureDelete`.
- **Validation** (Go 1.26.8, linux/amd64):
  - `gofmt -s`, `go mod tidy` (no diff), `go vet` (native plus four other targets)
    and golangci-lint v2.13.2 are clean;
  - `go test -race ./...` passes, with 71.7% total coverage; the internal suite also
    passes as a non-root user;
  - gosec: 10 findings, unchanged;
  - with real binaries, after `keys delete` the previous build left the blob in
    `cryptare.db` and the new build didn't.
- `VACUUM` was not adopted. A one-time clean-up of keys deleted by older versions is
  recorded as optional plan item 2.4a.

## 2026-09-27 — Key database opened only when needed, and created 0600 (plan 2.5, SEC-010, BUG-005)

- **Owner decision:** an existing database that others can read is set to 0600 when
  Cryptare opens it.
- `main.go`: no longer opens the database up front. `databaseOpener()` is passed to
  the new `newRootCmd(open)`, and the `log` import is gone.
- `internal/logic-cli.go`:
  - new `DatabaseOpener` type and `NewRootCmdLazy`. `NewRootCmd(db)` wraps it for
    tests;
  - `openOnce` (using `sync.OnceValues`) runs the opener at most once and adds "open
    key database" to its errors;
  - the `keys` subcommands and the TUI open the database first thing in `RunE`.
- `internal/database.go`: `prepareDatabaseFile` creates a new database file with mode
  0600 before SQLite opens it. It tightens an existing database, or its
  `-journal`/`-wal`/`-shm` file, to 0600 when others can read it, and ignores files
  owned by someone else. In-memory and `file:` URI paths, and Windows, are skipped.
- **Docs:** README (Configuration), CONTRIBUTING (build and run), `maint.md` (rule 5
  and §6).
  - Note for §6: a CGO-less build now fails only on `keys` commands and the TUI, so
    the CI/CD smoke tests' `keys generate`/`keys list` steps are what catch BUG-001
    regressions. Keep them.
- **Tests:**
  - failed against the previous code: `TestRootCmdOpensDatabaseOnlyForKeys`,
    `TestNewDatabaseCreatesPrivateFile`, `TestNewDatabaseTightensExistingFile`;
  - new: `TestFileCommandsDoNotCreateDatabase` (`main` package) and
    `TestKeysCmdReportsDatabaseOpenError`;
  - the version test now fails if `--version` opens the database.
- **Validation** (Go 1.26.8, linux/amd64):
  - `gofmt -s`, `go mod tidy` (no diff), `go vet` (native plus four other targets)
    and golangci-lint v2.13.2 are clean;
  - `go test -race ./...` passes, with 72.5% total coverage. Both test packages also
    pass as a non-root user;
  - gosec: 11 findings (one new G304 on the configured database path);
  - real binaries: the previous build created a 0644 `cryptare.db` on `--help`,
    `encrypt` and `compress`. The new build created none for those commands, created
    the database 0600 on `keys list`, and tightened an existing 0644 database. The
    TUI opened the key screen and created the database 0600.
- SEC-010 stays In Progress for step 3, a per-user default path (Q-003, plan 3.5).

## 2026-09-27 — Windows CI failure on `d683739` fixed (test only)

- CI #130 on `d683739` failed on `windows-latest` in "Run tests with coverage"; Ubuntu
  and macOS passed. The log needs a GitHub sign-in, so the cause was found locally.
- **Cause:** the new `TestFileCommandsDoNotCreateDatabase` (plan 2.5) ran `keys list`,
  which opened `cryptare.db` in the test's temporary folder, and never closed it.
  Windows can't delete an open file, so removing the temporary folder at the end of
  the test fails there. `maint.md` §5 already requires test databases to be closed
  for this reason. On Linux a probe of `/proc/self/fd` after the test showed the
  handle still open on the deleted `cryptare.db`.
- **Fix:** the test wraps `databaseOpener()` and closes every database it opened in a
  `t.Cleanup` that runs before the folder is removed. No product code changed.
- **Validation:** the same probe, run after both test packages, found no handle left
  open in any test temporary folder. `gofmt -s`, `go vet` (including windows/amd64),
  golangci-lint and `go test -race ./...` pass. Not verified on Windows itself;
  confirm with the next CI run.

## 2026-09-27 — CI green on `a5edc91`; SEC-006, SEC-008 and SEC-009 closed

- The owner committed the Windows test fix as `a5edc91` and reports every workflow
  green. That run is the first CI for `d751967` (plan 2.2) and `d683739` (plans
  2.3–2.5) on all three operating systems.
- Closed: SEC-006 (no plaintext temp file), SEC-008 (extraction hardening) and
  SEC-009 (secure delete).
- Still open:
  - SEC-002: the release-notes item;
  - SEC-010: a per-user default path (Q-003);
  - the rest of the register.

## 2026-09-27 — `--password-file` and a warning for `--password` (plan 2.6, SEC-004)

- **Owner decisions:**
  - add `--password-file` only; piped input already covers stdin;
  - `--password` always prints a warning, with no way to switch it off.
- `internal/logic-cli.go`:
  - new `passwordFlags` (`register`, `get`), used by `encrypt`, `decrypt`,
    `keys generate`, `keys export` and `keys import`. It adds `--password-file`,
    mutually exclusive with `--password`, which is kept for compatibility;
  - new `readPasswordFile`: the first line of the file, at most 64 KiB, spaces kept
    and only the line ending removed;
  - `--password` prints `passwordFlagWarning` on stderr;
  - the prompts are unchanged when neither flag is given.
- `README.md`: new flag lines; the examples use `--password-file`; a scripting note
  on keeping passwords off the command line; an upgrade note about the warning.
- **Tests:** `TestReadPasswordFile`, `TestPasswordFileFlag` and `TestPasswordFlagWarns`
  failed against the previous code and pass now.
- **Validation** (Go 1.26.8, linux/amd64):
  - `gofmt -s`, `go mod tidy` (no diff), `go vet` (native plus four other targets)
    and golangci-lint v2.13.2 are clean;
  - `go test -race ./...` passes, with 72.9% total coverage;
  - gosec: 12 findings (one new G304 on the password-file path);
  - a real binary: file and piped input work without a warning, `--password` warns,
    and combining the two flags is refused.
- **Follow-up (optional, W11):** move the CI, CD and Docker smoke tests to
  `--password-file`.

## 2026-09-27 — Smoke tests use `--password-file` (W11)

- The owner approved W11. `cryptare-w11-password-file-smoke.patch` changes the smoke
  steps:
  - `ci.yml` and `cd.yml` write a throwaway passphrase to `smoke-pw.txt` and pass it
    with `--password-file` to `keys generate`, `encrypt` and `decrypt`;
  - `docker.yml` mounts the file read-only into the container for `keys generate`.
- The workflows are protected from direct edits, so this is a patch. It needs the 2.6
  code, which the owner committed as `e512844` in the meantime.
- **Validation:**
  - actionlint 1.7.12 is clean, and the patch applies to the workflow files on the
    owner's machine;
  - run locally against the 2.6 code, the CI smoke step and the CD smoke step both
    passed (the CD static-link check needs a Linux release build, so it was skipped)
    with no warning in their output;
  - the Docker step wasn't run: no Docker daemon was available.

## 2026-09-27 — Imported key metadata validated (plan 2.7, SEC-011)

- `internal/crypto.go`: new `ErrInvalidKeyExport`, `validateKeyExport` and `isKeyID`,
  and constants for the export version, algorithm, key-ID length and stored-blob
  size. `ImportKeyFromFile` validates before returning, so the CLI and TUI both get
  it.
- **Tests:** `TestImportKeyRejectsInvalidMetadata`,
  `TestKeysImportCmdRejectsInvalidExport` and `TestDashboardImportRejectsInvalidExport`
  failed against the previous code and pass now.
- **Validation** (Go 1.26.8, linux/amd64):
  - `gofmt -s`, `go mod tidy` (no diff), `go vet` (native plus four other targets)
    and golangci-lint v2.13.2 are clean;
  - `go test -race ./...` passes, with 73.6% total coverage; both packages also pass
    as a non-root user;
  - gosec: 12 findings, unchanged;
  - real binaries: a `.ckey` with terminal escapes in its key ID was imported and
    echoed raw by the previous build. The new build refused it, with the ID escaped
    in the error.
- `README.md`: `keys import` notes that only Cryptare's export format is accepted.

## 2026-09-27 — Folders are archived through `os.Root` (plan 2.8, SEC-014)

- `internal/compress.go`: new `walkSourceTree` and `visitOpenFile`, and a nil-by-default
  test hook `testHookBeforeArchiveOpen`. `writeTarGz` and `writeZipDirectory` now
  walk and read the source folder through an `os.Root`, and build each entry's
  header from the opened file's own metadata.
- **Test:** `TestArchivingIgnoresFileSwappedForSymlink` covers tar.gz, zip and
  directory encryption. With the same hook added to the previous code, the swapped-in
  outside file was archived with no error in all three. Now each is refused and
  leaves no output.
- **Validation** (Go 1.26.8, linux/amd64):
  - `gofmt -s`, `go mod tidy` (no diff), `go vet` (native plus four other targets)
    and golangci-lint v2.13.2 are clean;
  - `go test -race ./...` passes, with 73.9% total coverage; the internal suite also
    passes as a non-root user;
  - gosec: 10 findings, down from 12 (G122 and one G304 are gone);
  - tar.gz and zip archives of the same tree from the previous and new builds are
    byte-identical.

## 2026-09-27 — Case-insensitive extensions and the export file name (plan 2.9, BUG-007, BUG-006)

- **BUG-007:**
  - `compress.go`: new `hasSuffixFold`. `defaultDecompressOutput` and
    `isTarGzArchive` (the file extension and the gzip header name) ignore letter
    case;
  - `crypto.go`: new `defaultDecryptOutput` (`.enc` in any case) is used by
    `DecryptFile` and the CLI's `deriveDecryptOutput`;
  - so `FOO.ZIP` now extracts to `FOO`, and an upper-case `.TAR.GZ` made by `tar czf`
    is extracted as a folder instead of being written out as a raw tar file named
    `….dec`.
- **BUG-006:**
  - `crypto.go`: new `defaultExportPath`, which reads the clock through `timeNow`, a
    package variable tests can replace;
  - the CLI and TUI compute the default export name once and pass it to
    `ExportKeyToFile`, so the name they report is the file written.
- **Tests:**
  - `TestExtensionsIgnoreCase` failed against the previous code and passes now;
  - `TestKeysExportReportsWrittenPath` (CLI) and `TestDashboardExportReportsWrittenPath`
    (TUI) use a clock that advances on every read. The previous code read the clock
    twice: with the same clock routed into it, both tests failed, reporting a file
    one second later than the one written.
- **Real binary:** `decompress DATA.TAR.GZ` (made by `tar -czf`) produced a raw
  `DATA.TAR.GZ.dec` file under the previous build and a `DATA/` folder under the new
  one.

## 2026-09-27 — One TUI action at a time (plan 2.10, BUG-008)

- `internal/logic-tui.go`: `advanceOrSubmitForm` refuses to submit while `busy`. It
  shows "Another action is still running; …" and keeps the form open, so it can be
  submitted once the running action reports back.
- **Test:** `TestDashboardRefusesSecondActionWhileBusy` failed against the previous
  code, which started a second action.
- **Test fixes:** three existing TUI tests (`TestDashboardEncryptDecryptDirectoryRoundTrip`,
  `TestDashboardCompressDecompressZipRoundTrip` and the `submit` helper in
  `TestDashboardRejectsEmptyPassword`) submitted a second form without handing the
  first result back to the model, which Bubble Tea always does. They now deliver it.
  The `submit` helper also stops after one pass through the form instead of looping
  forever; before this fix it hung until the test timeout.
- **Validation for 2.7–2.10** (Go 1.26.8, linux/amd64):
  - `gofmt -s`, `go mod tidy` (no diff), `go vet` (native plus four other targets)
    and golangci-lint v2.13.2 are clean;
  - `go test -race ./...` passes, with 74.3% total coverage; both packages also pass
    as a non-root user;
  - gosec: 10 findings;
  - the pseudo-terminal password checks still pass 7 of 7, including the TUI
    mismatch-then-match flow.
- With 2.10, Phase 2 is complete apart from the optional items 2.4a and 2.12.

## 2026-09-27 — Owner decisions; W10 action updates

- The owner committed 2.7–2.10 with the W11 workflow patch as `3d9384e`.
- **Decisions:** W10 approved. Optional items 2.4a (a one-time `VACUUM` for keys
  deleted by older versions) and 2.12 (a TUI overwrite option) are declined.
- **W10** (`cryptare-w10-action-updates.patch`): every action that ran on Node.js 20
  moves to its current release on Node.js 24, pinned to a full commit SHA:
  - `actions/checkout` v4.4.0 → v7.0.1. v6 keeps persisted credentials in a separate
    file; v7 refuses fork checkouts under `pull_request_target`/`workflow_run`, which
    these workflows don't use;
  - `actions/setup-go` v5.6.0 → v7.0.0. The toolchain directive is honoured and the
    cache key is based on `go.mod` unless `cache-dependency-path` is set (`ci.yml`
    sets `go.sum`);
  - `actions/upload-artifact` v4.6.2 → v7.0.1 (archiving stays on by default);
  - `actions/download-artifact` v4.3.0 → v8.0.1 (`pattern` plus `merge-multiple` are
    unchanged; a digest mismatch now fails the download);
  - `codecov/codecov-action` v5.5.5 → v7.1.1 (it now calls `actions/github-script`
    v8, on Node.js 24);
  - `github/codeql-action` v3.38.1 → v4.38.2 (init, autobuild, analyze and
    upload-sarif);
  - `softprops/action-gh-release` v2.6.2 → v3.0.3;
  - golangci-lint-action v9.3.0 already runs on Node.js 24, and gosec is a Docker
    action, so both are unchanged.
- **Validation:**
  - each new SHA is the commit behind its release tag (`git ls-remote`), and its
    `action.yml` declares Node.js 24 (or is a composite action);
  - every input the workflows pass exists in the new `action.yml`;
  - actionlint is clean, and the patch applies to the workflows on the owner's
    machine;
  - the workflows themselves weren't run: that needs a push.

## 2026-09-27 — Version 2 format: Argon2id and streaming encryption (plans 3.1 and 3.2)

- The owner committed W10 as `b520b97`.
- **Decisions** for plan 3.1:
  - Argon2id with 64 MiB, 3 passes and 4 lanes;
  - the versioned header and chunked streaming (3.2) built together;
  - every kind of artifact written in the new format;
  - old formats readable with no time limit, and no migration command.
- **Format** (new `internal/format_v2.go`; layout in `maint.md` §3):
  - a 46-byte header: magic `CRYPTARE\0`, version 2, content type, key source (2 is
    reserved for plan 3.4), KDF, Argon2id settings, salt, chunk size and nonce prefix;
  - the data follows in 64 KiB AES-256-GCM chunks (STREAM construction: counter and
    last-chunk flag in the nonce, the header as additional data).
- **What writes it:** `EncryptFile` for files and folders (a folder's tar.gz streams
  straight in), `EncryptKeyBlob` and `ExportKeyToFile`.
- **What reads it:** `DecryptFile`, `DecryptKeyBlob` and `ImportKeyFromFile` read both
  versions, and release output only after the final chunk authenticates.
- **Limits:** header settings above 1 GiB, 10 passes or 16 lanes are refused before
  any key derivation.
- **Library:** `deriveKey`, now used only for legacy data, moves to the standard
  library's `crypto/pbkdf2`. `x/crypto` is kept for `argon2`, so `go.mod` is unchanged.
- **Code removed:** `encryptBytes`, `encryptBytesWithAAD` and `buildDirectoryArchive`
  leave production code. Test-only copies of the legacy writers are in
  `legacy_fixtures_test.go`, whose `TestMain` lowers Argon2id for speed.
- **Why:** SEC-005 (the KDF was below guidance and the formats had no version) and
  BUG-010 (whole files were held in memory).
- **Tests added:**
  - `TestEncryptWritesVersion2Format`, which failed against the previous code for all
    four artifact kinds;
  - `TestDefaultPasswordKDFIsWritten`, `TestStreamRoundTripSizes` and
    `TestStreamRejectsTampering` (12 kinds of tampering, file and folder);
  - `TestDecryptRejectsUnsupportedHeaders` and `TestDecryptRejectsWrongContentType`;
  - `TestDeriveKey` now checks PBKDF2 known answers.
- **Validation** (Go 1.26.8, linux/amd64):
  - `gofmt -s`, `go mod tidy` (no diff), `go vet` (native plus four other targets)
    and golangci-lint v2.13.2 are clean;
  - `go test -race ./...` passes, with 75.4% coverage of `internal`; both packages
    also pass as a non-root user;
  - gosec: the same 10 findings as before;
  - the pseudo-terminal password checks pass 7 of 7, and the extraction-limit checks
    10 of 10.
  - On real binaries:
    - a 1 GiB file peaked at 77 MiB with the new build, against 3,088 MiB (encrypt)
      and 2,063 MiB (decrypt) with the previous one;
    - the previous build's file, folder, stored key and `.ckey` read correctly with
      the new build;
    - the previous build refuses version 2 data with the generic error;
    - the static Linux release build passes the CD smoke steps.
- Left uncommitted for the owner's review.

## 2026-10-03 — Repository re-analysis: SEC-015–SEC-019, BUG-015–BUG-024, plan Phase 5

- **Why:** the owner asked for an analysis of the repository's bugs and security
  issues, and a plan to handle all of them.
- **Scope:** every Go source at `0c57aef` (including the version 2 format), the tests
  that guard the format, the workflows, `Dockerfile`, `.gitignore`, `.devcontainer/`
  and the README's security claims. Each finding was checked against the existing
  register, so only new issues and gaps in earlier fixes were added.
- **Status found:** plans 3.1 and 3.2 had been committed in `0c57aef`; entries written
  before then called them uncommitted. `plan.md` and SEC-005 are corrected.
- **New security items** (`cybersec.md`):
  - SEC-015 (Medium): an interrupted decrypt leaves partial plaintext in hidden
    temporary files or folders, after Ctrl+C or another signal in the CLI, or after
    quitting the TUI while an action runs;
  - SEC-016 (Low): the key database in the working folder is used even when another
    user owns it, and its rows are printed raw, which allows terminal escape
    injection;
  - SEC-017 (Low): GORM's default logger prints SQL with bound values, including
    encrypted key blobs, to stdout;
  - SEC-018 (Medium): CI, release and local builds use Go 1.26.0, so later security
    releases aren't picked up, and govulncheck isn't run;
  - SEC-019 (Low): the owner-only permissions the README promises don't apply on
    Windows.
- **Gaps in existing items:**
  - SEC-003: `*.ckey` exports and SQLite side files aren't git-ignored;
  - SEC-005: committed; the worst-case read cost was measured (about 1 GiB and 1.9 s
    per attempt from a 62-byte file);
  - SEC-010: a `?` in the database path or a `file:` URI bypasses the 0600 file;
  - SEC-011: display-side escaping moves to SEC-016;
  - SEC-013: no `.dockerignore`, and unused runtime packages.
- **New defects** (`notes.md`):
  - BUG-015: names outside Latin-1 break gzip compression and folder encryption;
  - BUG-016: folder encryption can write its output inside the folder (`encrypt dir/`
    gives `dir/.enc`);
  - BUG-017: `keys export` overwrites any file, including the key database;
  - lower severity, BUG-018 to BUG-024: prompt signals other than SIGINT,
    text-based containment checks, silently overridden compression options, long
    names, single `.tar` files, `--password ""`, and uncapped whole-file reads;
  - BUG-011 gains a note on the misleading export prompt.
- **Questions:** Q-008 to Q-012 added; Q-003 is more urgent.
- **Plan:** Phase 5 in `plan.md`, in four tiers. Tier A (5.1–5.5) comes first,
  followed by a patch release.
- **Validation** (owner's machine, Go 1.26.0, linux/amd64, non-root):
  - `gofmt -s -l .` and `go vet ./...` are clean;
  - `go test -race -count=1 ./...` passes, with 75.0% (`main`) and 75.4%
    (`internal`) coverage;
  - `go mod verify` reports all modules verified;
  - each finding was reproduced with probe tests or real binaries built from a scratch
    copy outside the repository, except where the item says it comes from code review.
- **Not run:** golangci-lint, gosec and govulncheck (they need downloads that weren't
  approved), Windows, macOS, the Docker build, fuzzing.
- **Changed:** `intel/cybersec.md`, `intel/notes.md`, `intel/plan.md` and this file.
  No code, tests, configuration, CI, dependencies or git history were changed.
  `map.md` and `maint.md` are unchanged: nothing structural changed, and their rule
  updates belong with the fixes.

## 2026-10-03 — gosec Code Scanning alerts resolved (SEC-012, plan 4.1)

- The 10 open gosec alerts in Code Scanning (#14, #17, #18, #25, #31–#33, #35–#37)
  were the 10 findings triaged under SEC-012. Their dispositions are now in the code:
  - G304 ×8 (file path from a variable): accepted by design. Each call opens a path
    the user chose and is annotated `// #nosec G304 -- <reason>`: `compress.go`
    `writeGzip`, `DecompressFileWithLimits` and `writeZipFile`; `crypto.go`
    `encryptSingleFile`, `DecryptFileWithLimits` and `ImportKeyFromFile`;
    `database.go` `prepareDatabaseFile`; `logic-cli.go` `readPasswordFile`.
  - G301 ×2 (`MkdirAll` with 0755): owner decision, missing parent folders of an
    output are created 0700 (`extractZipSingleFile`, `extractToDir`), like everything
    else the tool writes.
- New test: `TestExtractCreatesPrivateParentFolders` (tar.gz folder, zip folder and
  single-file zip outputs under missing parents).
- `maint.md` §4 now gives the 0700 directory mode and the `#nosec` convention.
- **Validation** on a scratch copy outside the repository (Go 1.26.0, linux/amd64,
  non-root). The module proxy was unreachable, so gorm, gorm's SQLite driver,
  x/crypto, x/sys and x/text were replaced with clones of the release tags `go.sum`
  pins; the repository's `go.mod` is unchanged.
  - gosec v2.29.0 (release checksum verified) with CI's arguments: 10 findings before
    the change, 0 after (8 `#nosec`), and the SARIF output has no results;
  - the new test fails on the unchanged code (parents 0755) and passes after;
  - `gofmt -s -l .` and `go vet ./...` are clean; `go test -race -count=1 ./...`
    passes, with 75.0% (`main`) and 75.4% (`internal`) coverage.
- **Not run:** golangci-lint, govulncheck, Windows and macOS. The alerts close only
  after the change is pushed and the Security workflow runs.
- **Found:** Code Scanning's CodeQL status warns "Actions workflow file not found" for
  configurations left by `codeql.yml` (added `58d80ae`, deleted `2691714`).
- **Changed:** `internal/compress.go`, `internal/crypto.go`, `internal/database.go`,
  `internal/logic-cli.go`, `internal/compress_test.go`, `intel/cybersec.md`,
  `intel/maint.md`, `intel/notes.md`, `intel/plan.md` and this file. Not committed.

## 2026-10-03 — Key database trust and escaped key listings (SEC-016 steps 1–2, plan 5.6)

- **Owner decisions:** Q-010 is answered: an untrusted key database is refused, not
  warned about. Moving the default path (Q-003, plan 3.5, SEC-016 step 3) is left for
  a separate change, so SEC-016 stays In Progress.
- **Step 1, refusal:** `prepareDatabaseFile` runs the new `checkDatabaseFileTrust` on
  the database and any `-journal`, `-wal` or `-shm` file next to it, including when
  the database was just created (a planted journal would otherwise be replayed into
  it). A file another user owns, or with a group or other write bit, is refused with
  the new `ErrUntrustedDatabase` and left unchanged. A private file the user owns that
  others can read is still set to 0600 (SEC-010). The owner comes from `fileOwner`,
  in the new build-tagged `internal/fileowner_unix.go` and
  `internal/fileowner_other.go`; Windows is skipped as before (SEC-019).
  `databaseOpener` in `main.go` adds a hint to set `CRYPTARE_DB_PATH`.
- **Step 2, escaping:** the new `displayText` shows a stored key ID or algorithm that
  holds a control or other non-printable character, or invalid UTF-8, Go-quoted;
  `keys list` and the TUI key table use it. Ordinary values print unchanged.
- **Behaviour change:** on Linux and macOS, the `keys` commands and the TUI now stop on
  such a database instead of using it (README upgrade note and `CRYPTARE_DB_PATH`
  row).
- **Tests:** `TestCheckDatabaseFileTrust`, `TestNewDatabaseRefusesFilesOthersCanWrite`,
  `TestDatabaseOpenerRefusesUntrustedDatabase`, `TestDisplayText`,
  `TestKeysListEscapesControlCharacters` and `TestKeyScreenEscapesControlCharacters`.
  With stubs that keep the old behaviour, the four behavioural tests fail.
- **Validation** on a scratch copy outside the repository (Go 1.26.0, linux/amd64,
  non-root; the same module setup as the gosec entry above):
  - `gofmt -s -l .` and `go vet ./...` clean; `go vet` and test builds pass for
    windows, darwin and freebsd (CGO off);
  - `go test -race -count=1 ./...` passes, with 80.0% (`main`) and 77.3%
    (`internal`) coverage;
  - gosec v2.29.0: 0 issues (8 `#nosec`);
  - real binaries: the probe row from SEC-016's evidence gave 3 raw ESC bytes from
    `keys list` before and none after; a 0666 database and a 0666 journal planted next
    to a new database were refused with exit status 1, and the database stayed 0666.
- **Not run:** golangci-lint, macOS, Windows, and a database owned by another real
  account (the synthetic test covers the rule).
- **Changed:** `main.go`, `internal/database.go`, `internal/logic-cli.go`,
  `internal/logic-tui.go`, new `internal/fileowner_unix.go` and
  `internal/fileowner_other.go`; tests in `internal/database_test.go`,
  `internal/logic_cli_test.go`, `internal/logic_tui_test.go` and
  `lazy_database_test.go`; `README.md`, `intel/cybersec.md`, `intel/maint.md`,
  `intel/map.md`, `intel/notes.md`, `intel/plan.md` and this file. Not committed.

## 2026-10-03 — Phase 5 Tier A: toolchain, export and encrypt outputs, Unicode names, cancellation (plans 5.1–5.5)

- **Owner decisions:** 5.1 approved (`go.mod`, CI and `Dockerfile` changes); the
  behaviour changes in 5.2 and 5.3 confirmed; downloads of Go 1.26.8 and actionlint
  for validation approved.
- **5.4 / BUG-015:** gzip header names go through `gzipHeaderName`, which keeps a
  Latin-1 name and otherwise stores a fallback: `archive.tar` for a folder's tar.gz
  (so it is still recognised as a tarball) and no name for a file. Compressing and
  encrypting files and folders named in Chinese, Cyrillic or with emoji now works;
  reading is unchanged.
- **5.2 / BUG-017:** `keys export` checks its output with the new `checkExportOutput`
  before asking for the password: an existing file only with the new `--force` (the
  TUI never overwrites), and never the key database or its SQLite files
  (`ErrOutputIsKeyDatabase`; `Database` now remembers its path). `CheckOutputPath`'s
  existing-output check moved into `checkOutputFree`.
- **5.3 / BUG-016:** `defaultEncryptOutput` writes a folder's default output next to
  it (`dir/` → `dir.enc`, `.` → `<parent>/<name>.enc`), and `checkOutputOutsideFolder`
  refuses an output inside the folder (`ErrOutputInsideInput`), in the core, the CLI
  (before the password prompt) and the TUI.
- **5.5 / SEC-015, BUG-018:** `…Context` variants of the four file operations, checked
  before each read (`copyContext`), at each archive entry and before the final rename;
  `runCancellable` and `InterruptedError` in the CLI (exit 128 + signal); the TUI's
  `actionRunner` and `quit` (cancel, "Cancelling…", quit after the action reports
  back; the launcher waits for it however the program ends); the password prompt
  restores the terminal on SIGINT, SIGTERM, SIGQUIT and SIGHUP.
- **5.1 / SEC-018:** confirmed the v1.1.0 linux/amd64 release binary was built with
  go1.26.0; `toolchain go1.26.8` in `go.mod`; `golang:1.26.8-alpine` in the
  `Dockerfile`; a govulncheck v1.8.0 job in `security.yml`; `.github/dependabot.yml`.
- **Tests:** `TestNonLatin1NamesRoundTrip`, `TestGzipHeaderName`,
  `TestKeysExportRefusesExistingOutput`, `TestDashboardExportRefusesExistingOutput`,
  `TestEncryptFolderOutputStaysOutside`, `TestEncryptCmdFolderOutputStaysOutside`,
  `TestDashboardEncryptFolderOutputStaysOutside`, `TestCancelledOperationsLeaveNothingBehind`
  (22 cases), `TestUncancelledContextOperationsComplete`,
  `TestDashboardQuitWhileBusyCancelsAction`, `TestDashboardQuitWhenIdle`,
  `TestActionRunnerShutdownWaitsForAction`, `TestRunCancellableStopsOnSignal`,
  `TestRunCancellableWithoutSignal` (Unix), `TestExitCode` and
  `TestDecryptInterruptedBySignal` (Unix, a subprocess fed through a named pipe).
  The six 5.2–5.4 tests fail on the tree before these changes.
- **Validation** on scratch copies outside the repository (linux/amd64, non-root;
  modules as in the gosec entry above):
  - Go 1.26.8: `gofmt -s -l .` and `go vet ./...` clean; `go vet` and test builds pass
    for windows, darwin and freebsd (CGO off); `go test -race -count=1 ./...` passes;
    gosec v2.29.0 reports 0 issues (8 `#nosec`); actionlint 1.7.12 reports nothing in
    the four workflows. The suite also passes on Go 1.26.0.
  - Real binaries, old (the tree before these changes) against new:
    - names in Japanese and Russian: the old build failed to gzip, tar.gz or encrypt
      them; the new one round-trips them;
    - `encrypt secret/`: old wrote `secret/.enc` inside the folder; new writes
      `secret.enc` beside it;
    - `keys export --output precious.txt`: old replaced the file; new refuses it
      (exit 1). `--output keys.db`: old destroyed the key database; new refuses;
    - SIGINT, SIGTERM and SIGHUP mid-decrypt (file and folder, input through a named
      pipe): old was killed and left a hidden temporary file or folder of plaintext
      every time; new exited 130, 143 and 129 and left nothing;
    - SIGINT, SIGTERM, SIGQUIT and SIGHUP at the hidden prompt, in a pseudo-terminal:
      old restored echo only on SIGINT; new restored it every time.
- **Not run:** golangci-lint, govulncheck (its database isn't reachable from here),
  `go mod tidy` (needs the module proxy), Windows and macOS, the Docker build, a TUI
  session in a real terminal.
- **Changed:** `go.mod`, `Dockerfile`, `.github/workflows/security.yml`, new
  `.github/dependabot.yml`; `main.go`, `internal/compress.go`, `internal/crypto.go`,
  `internal/database.go`, `internal/logic-cli.go`, `internal/logic-tui.go`,
  `internal/ui-dashboard.go`; tests in `internal/compress_test.go`,
  `internal/crypto_test.go`, `internal/logic_cli_test.go`, `internal/logic_tui_test.go`,
  `internal/legacy_fixtures_test.go`, new `internal/cancel_test.go`,
  `internal/logic_cli_unix_test.go`, `interrupt_test.go` and `interrupt_unix_test.go`;
  `README.md`, `CONTRIBUTING.md`, `intel/cybersec.md`, `intel/maint.md`,
  `intel/map.md`, `intel/notes.md`, `intel/plan.md` and this file. Not committed.

## 2026-10-04 — CI verified; SEC-004/005/011/012/014/015 closed; Tier B, Tier C, 4.2/4.7/4.9, 5.19–5.21

- **Pushed and verified on GitHub:** the owner committed the 2026-10-03 work as
  `4ba516a`, `6a5fcb1` and `b3278ea`. CI #136–#138, Docker #18–#20 and Security
  #142–#144 passed; on `b3278ea`, `setup-go` installed go1.26.8, the Docker builder
  resolved `golang:1.26.8-alpine@sha256:8ac98ca5…`, and the new govulncheck job found
  nothing reachable. Code Scanning shows 0 open alerts. Runs on `e512844`, `3d9384e`
  and `0c57aef` had also passed. Dependabot ran and opened PR #26 (`golang.org/x/crypto`
  0.56.0 → 0.57.0) for the owner.
- **Closed** after checking each item's validation: SEC-004 (`e512844`), SEC-005
  (`0c57aef`), SEC-011 and SEC-014 (`3d9384e`), SEC-012 (all steps; 0 open alerts),
  SEC-015 (`6a5fcb1`/`b3278ea`). BUG-015 to BUG-018 moved to the resolved index.
- **Line endings:** the 2026-10-03 edit of `.github/workflows/security.yml` turned its
  CRLF line endings into LF, so `6a5fcb1` shows the whole file as changed (content
  unchanged apart from the govulncheck job; no other file was affected). Left as it
  is to avoid another whitespace-only change; edits since preserve each file's line
  endings (`ci.yml`, CRLF, changed by 2 lines below).
- **Owner decisions:** approved 4.2/5.9 (Dockerfile) and 5.19 (CI `-race`); confirmed
  5.12's behaviour change; Q-011 = `decompress --raw`; Q-012 = document only; the
  legacy-input cap for 5.15 = `--max-size`.
- **Changes (not yet committed):**
  - 5.7 / SEC-017: GORM's logger silenced, errors translated; `ErrKeyExists`; the CLI
    and TUI name the key in "not found" and "already stored" errors.
  - 5.8 / SEC-010: `databaseFilePath`; a plain path with `?` refused; `file:` URIs
    created 0600 and trust-checked.
  - 4.2, 5.9 / SEC-013, SEC-003: non-root image user 10001 owning `/app/data`, images
    pinned by digest, no runtime packages, `ARG VERSION`, `-trimpath`, `.dockerignore`;
    `.gitignore` gains `*.ckey` and SQLite side files.
  - 5.10 / SEC-019: Windows permissions documented.
  - 5.11 / BUG-019: containment by file identity. 5.12 / BUG-020: contradictory
    `--format` and out-of-range `--level` refused. 5.13 / BUG-021: shortened temporary
    names. 5.14 / BUG-023: `--password ""` counts as given. 5.15 / BUG-024: key exports
    capped at 1 MiB, legacy inputs at `--max-size`. 5.16 / BUG-022: `decompress --raw`.
  - 5.19: `-race` in CI on Linux and macOS. 5.20: README and `CONTRIBUTING.md`.
    5.21: six fuzz targets. 4.7 and 4.9: README release section and install steps.
- **Tests:** `TestDatabaseErrorsAreTyped`, `TestKeysCommandsKeepStdoutClean`
  (subprocess), `TestDatabasePathsOpenThePreparedFile`, `TestDatabaseFilePath`,
  `TestContainmentChecksUseFileIdentity` (its letter-case part runs on the macOS and
  Windows CI jobs), `TestLongNamesFitTemporaryFiles`, `TestTempNamePart`,
  `TestLegacyReadsAreBounded`, `TestGunzipRawRoundTrip`, `TestDefaultRawOutput`,
  `TestCompressCmdRefusesContradictoryOptions`, `TestDecryptCmdExplicitEmptyPassword`,
  `TestDecompressCmdRaw`, `TestDashboardCompressRefusesContradictoryOptions`, and the
  fuzz targets. The nine behavioural ones fail against `b3278ea`.
- **Validation** (scratch copies, Go 1.26.8, linux/amd64, non-root):
  - `gofmt -s -l .` and `go vet ./...` clean; `go vet` and test builds pass for
    windows, darwin and freebsd; `go test -race -count=1 ./...` passes (100% `main`,
    79.4% `internal`); gosec v2.29.0: 0 issues (9 `#nosec`); actionlint: no findings.
  - Fuzzing: 25–35 s coverage-guided runs of five targets without a failure.
    `FuzzExtractTar`'s coverage-guided worker stalls in this environment, spinning in
    Go runtime code (sampled with ptrace), not in Cryptare's; 80,000 uninstrumented
    inputs and a 200,000-input mutation stress of tar extraction (slowest 0.6 ms, no
    write outside the output folder) and 116,512 parseable mutated zips (slowest
    56 ms) found nothing.
  - Real binaries, `b3278ea` against this round: `keys export <unknown>` wrote 302
    bytes of SQL log to stdout, now 0; `CRYPTARE_DB_PATH=a?b/k.db` left a stray 0644
    file, now refused; `--format gzip --output x.zip` wrote a zip and `--level 42` was
    accepted, now both refused; a 3 MB non-Cryptare file under `--max-size 1MiB` was
    read whole (72 ms, "wrong password"), now refused before reading (13 ms) with the
    limit hint; a 247-byte name failed to encrypt, now works; `decompress --raw` gives
    back a tarball holding a symlink byte for byte.
  - The README's Linux install steps, run as written against v1.1.0, verify the
    checksum and print `cryptare version v1.1.0`.
- **Not run:** the Docker build (Docker isn't available here; the Docker workflow is
  its first build), golangci-lint, govulncheck locally, Windows and macOS at runtime,
  coverage-guided `FuzzExtractTar` on a normal machine.
- **Changed:** `.dockerignore` (new), `.gitignore`, `Dockerfile`,
  `.github/workflows/ci.yml`, `internal/compress.go`, `internal/crypto.go`,
  `internal/database.go`, `internal/logic-cli.go`, `internal/logic-tui.go`; tests in
  `internal/compress_test.go`, `internal/database_test.go`,
  `internal/logic_cli_test.go`, `internal/logic_tui_test.go`, new
  `internal/fuzz_test.go` and `keys_output_test.go`; `README.md`, `CONTRIBUTING.md`,
  `intel/cybersec.md`, `intel/maint.md`, `intel/map.md`, `intel/notes.md`,
  `intel/plan.md` and this file. Not committed.

## 2026-10-04 — v1.2.0 released; Q-003, Q-005, Q-006, Q-008 and Q-009 answered and implemented

- **Verified on GitHub:** the owner committed round 3 as `eb330a3` and `89a64e8` and
  tagged `89a64e8` as v1.2.0. CI #141/#142, Docker #22/#23, Security #147/#148 and CD #4
  passed, and the release carries the drafted notes. Docker #23 resolved both pinned
  image digests. Closed: SEC-002 (its migration notes are released), SEC-017 (CI green)
  and SEC-019 (wording committed and published). BUG-019 to BUG-024 moved to the
  resolved index.
- **Owner decisions:** Q-003: move the default key store to a per-user data folder,
  with a notice-only migration. Q-005: no history purge; delete the Copilot branch; add
  a CI guard. Q-006: rewrite `NOTICE` and ship a generated licence file in the release
  archives. Q-008: keep the Argon2id read limits (1 GiB, 10 passes, 16 lanes). Q-009:
  NFKC normalisation, byte order mark handling, `golang.org/x/text` as a direct
  dependency.
- **Q-005 / SEC-003:** the `copilot/research-compression-implementation` branch was no
  longer on GitHub when checked (`git ls-remote origin`) and had no pull request. Its
  tip was `e910ff1de2dd757d6bc2278058825b45ddc9affd` (2026-09-10, "Tighten archive path
  validation"), with three commits not in `main`: `0441cd3` "Add directory compression
  support", `e78ba4d` "Harden tar archive extraction" and `e910ff1` (merge base
  `61fd049`; 8 files, +405/−64). The owner's clone still has them as
  `origin/copilot/research-compression-implementation` until a `git fetch --prune`, and
  `git push origin e910ff1de2dd757d6bc2278058825b45ddc9affd:refs/heads/copilot/research-compression-implementation`
  would restore the branch from there. `cryptare.db` stays in the history and in tag
  `v1.0.0`. `ci.yml` gains a "Refuse committed key material" step after checkout.
- **Q-006 / plan 4.5:** `NOTICE` now holds the project's attribution and a pointer to
  `THIRD_PARTY_LICENSES.txt`. New `scripts/third-party-licenses.sh` lists the modules
  `go list -deps` reports for each release target (CGO on), and writes their licence
  files, the Go standard library's `LICENSE` and a note on SQLite (public domain). A
  module without a licence file stops it unless listed in `stated_licence`
  (`github.com/mattn/go-localereader`, MIT per its README). The `cd.yml` release job
  checks out the repository (`persist-credentials: false`), sets up Go from `go.mod`
  without a cache, runs the script, and puts `LICENSE`, `NOTICE` and
  `THIRD_PARTY_LICENSES.txt` in every `.tar.gz` and `.zip` next to the binary.
  `/THIRD_PARTY_LICENSES.txt` is git-ignored. Found while doing it: the static Linux
  binaries contain glibc (LGPL), which the file doesn't cover (Q-013, plan 4.10).
- **Q-009:** new data is protected by the password without a leading byte order mark,
  in NFKC form (`normalizePassword`). The header's KDF byte is `2` when that differs
  from the password as given and `1` otherwise, so v1.2.0 reads any data whose password
  normalisation leaves alone, and refuses the rest as "key derivation 2" instead of
  reporting a wrong password. Reads of KDF `1` and of the legacy formats try the
  password without a leading byte order mark, then as given if it had one, then its
  NFKC form (`passwordCandidates`); KDF `2` uses the NFKC form only.
  `newDecryptingReader` now authenticates the first chunk itself, trying each
  candidate, so a wrong password fails before any output is created.
  `CheckPasswordPolicy` counts the normalised password. `golang.org/x/text` moved to
  the direct requires (no `go.sum` change; it was already linked in through GORM).
- **Q-003 / plan 3.5:** without `CRYPTARE_DB_PATH`, the key store is
  `cryptare/cryptare.db` in the user data folder (`databasePath`, `userDataDir`):
  `%LocalAppData%` on Windows, `~/Library/Application Support` on macOS,
  `$XDG_DATA_HOME` (absolute only) or `~/.local/share` elsewhere. A missing or relative
  folder is an error naming `CRYPTARE_DB_PATH`. The opener creates the folder 0700. A
  `cryptare.db` in the current folder gets a notice on stderr (`noticeLegacyDatabase`)
  suggesting `mv`/`move` or `CRYPTARE_DB_PATH`, and is never opened. New `cryptare keys
  path` prints the path and creates nothing. Plan 4.6: the `tasks.db` fixture is now
  `keys.db`.
- **Tests:** `TestNormalizePassword`, `TestPasswordCandidates`,
  `TestCheckPasswordPolicyCountsNormalizedPassword`,
  `TestUnicodePasswordFormsOpenTheSameData` (files, folders, stored keys and key
  exports), `TestNormalizedPasswordHeader`, `TestEarlierDataOpensWithPasswordForms`, a
  "key derivation identifier changed" tampering case, an id-2 fuzz seed;
  `TestDatabasePathDefaultsToDataFolder`, `TestUserDataDir`,
  `TestDatabaseOpenerUsesPrivateDataFolder`, `TestLegacyDatabaseNotice`,
  `TestQuotePath`. Against `89a64e8`, `TestUnicodePasswordFormsOpenTheSameData` (4 of 4
  cases) and the policy test (5 of 5) fail.
- **Validation** (scratch copies, Go 1.26.8, linux/amd64, non-root):
  - `gofmt -s -l .` and `go vet ./...` clean, plus `go vet` and test builds for
    windows, darwin and freebsd; `go test -race -count=1 ./...` passes; gosec v2.29.0:
    0 issues (9 `#nosec`); actionlint 1.7.12: nothing. `go mod tidy` can't run here;
    `go mod tidy -e` leaves `go.mod` as edited.
  - The licence script wrote 30 modules for the five targets (27 for linux/amd64
    alone; `coninput`, `mousetrap` and `go-localereader` only on Windows) and failed,
    writing nothing, with `go-localereader` taken out of `stated_licence`. The release
    packaging steps, run on stand-in binaries, gave archives holding the binary,
    `LICENSE`, `NOTICE` and `THIRD_PARTY_LICENSES.txt`.
  - The CI guard passes on this tree, and fails on `v1.0.0`'s (`cryptare.db`) and on a
    scratch repository tracking `keys.CKEY` and `x.db-wal`.
  - Real binaries, v1.2.0 source (`89a64e8`) against this round, through password
    files: v1.2.0 data opens in the new build with the same password, and with
    combining accents when it was made with precomposed letters; new data with an
    ASCII or precomposed password (KDF 1) opens in v1.2.0; new data with combining
    accents or a byte-order-mark password file (KDF 2) is refused by v1.2.0 as
    "unsupported encrypted data: key derivation 2" and opens in the new build however
    the password is entered; v1.2.0 data made through a byte-order-mark file opens with
    the same file. As designed, v1.2.0 data made with combining accents doesn't open
    with precomposed letters, and a wrong password writes nothing.
  - Q-003, umask 022: v1.2.0's `keys generate` created `./cryptare.db`. The new build's
    `--help` and `encrypt` created nothing; `keys path` printed the new path and the
    notice and created nothing; `keys list` created `~/.local/share/cryptare` (0700)
    and the store (0600), and left `./cryptare.db` byte-identical. The old key was
    listed after the suggested `mv`, or with `CRYPTARE_DB_PATH=./cryptare.db`. With
    `HOME` unset the commands fail with the `CRYPTARE_DB_PATH` hint.
- **Not run:** golangci-lint and govulncheck locally, Windows and macOS at runtime
  (their CI jobs run the new tests), the release job itself (first run on the next
  tag), `go version -m` on the v1.2.0 binaries (SEC-018) and the image's UID check
  (SEC-013).
- **Changed:** `.github/workflows/ci.yml`, `.github/workflows/cd.yml`, `.gitignore`,
  `NOTICE`, `go.mod`, new `scripts/third-party-licenses.sh`, `database_path.go`,
  `main.go`, `internal/crypto.go`, `internal/format_v2.go`; tests in
  `database_path_test.go`, `lazy_database_test.go`, `internal/format_v2_test.go`,
  `internal/fuzz_test.go` and new `internal/password_norm_test.go`; `README.md`,
  `CONTRIBUTING.md`, `intel/cybersec.md`, `intel/maint.md`, `intel/map.md`,
  `intel/notes.md`, `intel/plan.md` and this file. Not committed.

## 2026-10-04 — v1.3.0 released; Q-013 (musl Linux builds, C library notices); `golang.org/x/crypto` 0.57.0

- **Verified on GitHub:** the owner committed round 4 as `a088f7f` and released it as
  v1.3.0 (CD #5) with the drafted notes; CI #143, Docker #24 and Security #149 passed.
  The release archives grew by about 30–40 KB each, in line with the new licence files,
  and the packaging step ran without error. Closed: SEC-003 (the CI guard ran on all
  three systems), SEC-010 and SEC-016 (the per-user key store, with its tests run on
  Ubuntu, Windows and macOS).
- **PR #26:** Dependabot rebased it onto `a088f7f` by itself, so the expected `go.mod`
  conflict never needed resolving. `golang.org/x/crypto` 0.57.0 requires
  `golang.org/x/text` 0.42.0 (and `x/sys` 0.48.0), and `x/text` 0.42.0 changes NFC/NFKC
  composition: v0.41.0 composed an accent with a letter across a blocking character,
  and truncated recomposition keys to 16 bits. Comparing the two on 300,000 random
  mixed-script strings, 104 normalised differently; every one involved an Indic vowel
  sign or length mark (Tamil, Malayalam, Bengali, Oriya, Kannada, Sinhala, Myanmar,
  Balinese, Grantha, …) before a combining accent, and v0.42.0 matched Python's
  `unicodedata` (Unicode 13.0) in all 300,000. Both versions select the Unicode 15.0
  tables under Go 1.26; v0.42.0's Unicode 17 tables are built only from Go 1.27.
- **Owner decisions:** Q-013: build the Linux releases against musl and add the C
  library notices; leave the published releases (v1.0.1–v1.3.0, statically linked with
  glibc) as they are. PR #26: take the update, with a known-answer test and a release
  note.
- **Q-013 / plan 4.10:** `cd.yml` builds Linux targets inside the `Dockerfile`'s pinned
  `golang:1.26.8-alpine` image (`LINUX_BUILDER`, `apk add build-base git`), statically
  against musl, with the same tags and flags as before; the output is handed back to
  the runner's user and `GIT_OPTIONAL_LOCKS=0` keeps VCS stamping from rewriting
  `.git/index` as root. The Windows smoke test lists the binary's DLL imports
  (`objdump -p`) and fails on a MinGW-w64 toolchain DLL (warning only if `objdump` is
  missing). Go's linker always adds `-lmingwex -lmingw32` on Windows (`cmd/link`
  `lib.go`), so the MinGW-w64 runtime is in the binary. `scripts/third-party-licenses.sh`
  gains `c_libraries`: musl's `COPYRIGHT` (release 1.2.5, from the `kraj/musl` mirror of
  git.musl-libc.org, which the analysis machine couldn't reach) for Linux, and
  `COPYING.MinGW-w64-runtime.txt` (mingw-w64 v12.0.0) for Windows, both vendored under
  `scripts/licenses/`; libgcc (GCC Runtime Library Exception) needs no notice.
- **PR #26 / Q-009:** `go.mod` and `go.sum` carry PR #26's update, byte-identical to its
  files (blobs `2a8dcd7` and `30a1907`). New `TestNormalizePasswordKnownAnswers` pins 16
  normalised forms, all checked against Python; with `x/text` 0.41.0 its "accent after a
  vowel sign" case fails (`"o\u0bbe\u0317\u0301"` became `"\u00f3\u0bbe\u0317"`).
- **Validation** (scratch copy, Go 1.26.8, `x/crypto` 0.57.0, `x/text` 0.42.0, `x/sys`
  0.48.0): `gofmt -s -l .` and `go vet ./...` clean, plus `go vet` and test builds for
  windows, darwin and freebsd; `go test -race -count=1 ./...` passes; gosec v2.29.0:
  0 issues (9 `#nosec`); actionlint 1.7.12: nothing. The licence script lists 30
  modules and 2 C libraries for the release targets (musl for both Linux targets, the
  MinGW-w64 runtime for Windows, none for darwin) and fails, writing nothing, when a
  listed notice file is missing. The new Linux build step passes `bash -n`, its
  container script `sh -n`, and a dry run with a stand-in `docker` passes the expected
  arguments.
- **Real binaries,** the v1.3.0 source against this round, through password files: an
  ASCII password and one with combining accents work in both directions; the affected
  kind of password (`o` + U+0BBE + U+0317 + U+0301 inside a passphrase) written by
  v1.3.0 (KDF 2) doesn't open in the new build, as expected, while the new build writes
  KDF 1 for it, which v1.3.0 reads.
- **Not run:** the musl build and the Windows DLL check themselves (CD only; Docker
  isn't available here, so run CD by hand before tagging), golangci-lint and govulncheck
  locally.
- **Changed:** `.github/workflows/cd.yml`, `go.mod`, `go.sum`,
  `scripts/third-party-licenses.sh`, new `scripts/licenses/musl-COPYRIGHT.txt` and
  `scripts/licenses/mingw-w64-runtime.txt`, `internal/password_norm_test.go`;
  `README.md`, `CONTRIBUTING.md`, `intel/cybersec.md`, `intel/maint.md`,
  `intel/map.md`, `intel/notes.md`, `intel/plan.md` and this file. Not committed.

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
  binary, so a format change made in the writer and the reader alike would pass every
  test; the README's install example pinned v1.1.0; every runtime error printed the
  whole usage text; release files had no proof of origin (SEC-020); Dependabot didn't
  cover the digest-pinned Docker images; the Docker UID check (SEC-013) wasn't
  automated; `notes.md` and `plan.md` still described round 5 as uncommitted. The
  `FuzzExtractTar` "stall" (plan 5.21) was reproduced: it is the fuzzer minimising
  each new input (60 s by default), not a hang; with `-fuzzminimizetime 3s` it ran
  60 s, about 24,800 inputs, without a failure.
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
- **Code:** `Credential` (password or unlocked stored key) through `newEncryptingWriter`,
  `newDecryptingReader`, `EncryptFileWithCredentialContext` and
  `DecryptFileWithCredentialContext`; `EncryptedWithStoredKey` reads a regular file's
  header only, because reading a pipe there consumed the stream (caught by
  `TestDecryptInterruptedBySignal`, which feeds decrypt through a named pipe). New
  `keys.go` (plan 3.3): `GenerateStoredKey`, `StoredKeyCredential` (the policy applies
  to a key's master password when it encrypts) and `ImportStoredKey` (checks the key
  inside; `ErrSeparateKeyPassword` for an older export with its own password), over the
  `Storage` interface. `ExportKeyToFile` checks the master password, then the policy
  (5.17, BUG-011). CLI: `encrypt --key`/`-k`, decrypt finds the key from the header,
  export asks once for the key's master password by name, import takes an older
  export's own password through `--export-password-file` or a second prompt at a
  terminal (either password in either place), delete warns about files encrypted with the key,
  and `silenceUsageOnRun` keeps the usage text for command-line mistakes. TUI: a stored
  key field at the end of the encrypt form, decrypt finding the key, export with one
  password field, import with an optional field for an older export's key password.
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
- **Validation** (scratch copy, Go 1.26.8 built from source; the module proxy was
  unreachable, so `x/crypto`, `x/text`, `x/sys` and the two GORM modules came from
  their GitHub mirrors at the `go.mod` versions, the rest checked against `go.sum`;
  `go.mod` and `go.sum` unchanged): `gofmt -s -l .` and `go vet ./...` clean;
  `go test -race -count=1 ./...` passes (89.2% `main`, 80.9% `internal`); golangci-lint
  v2.13.2 and gosec v2.29.0: 0 issues (10 `#nosec`); actionlint 1.7.12: nothing. Tests whose
  expectations changed by owner decision: the export cases in
  `TestEncryptRejectsEmptyPassword`, `TestEncryptRejectsWeakPassword`,
  `TestNewPasswordCmdsRejectWeakPassword`, `TestLegacyShortPasswordCmds`,
  `TestDashboardRejectsEmptyPassword`, `TestDashboardRejectsWeakOrMismatchedPassword`
  and `TestDashboardNewPasswordFormsHaveConfirmation`, and TUI tests that step through
  the encrypt form's new last field.
- **Independent review** (a separate agent, with probes and 30–40 s fuzz runs of
  `FuzzParseV2Header` and `FuzzDecryptingReader`, about 450,000 inputs each): no defect
  in the format or the cryptography. Fixed from its findings: `encrypt --key ""` fell
  back to a password (now refused); `ImportStoredKey` took an empty key password as
  "not given" and needed the two passwords of an older export in one order, and scripts
  had no way to give the second (now order-independent, with `--export-password-file`);
  the policy on a stored key's master password was enforced only in the interfaces (now
  `canEncrypt` in core); stale comments and docs about which commands open the key store;
  gosec G602 on `readV2Header` (rewritten, not suppressed).
- **Not run:** govulncheck (database unreachable), the Docker build, Windows and macOS,
  and the new workflow steps (they run on the next push and tag).
- **Changed:** `.github/dependabot.yml`, `.github/workflows/cd.yml`,
  `.github/workflows/docker.yml`, `SECURITY.md`, `README.md`, `CONTRIBUTING.md`,
  `main.go` (comment), `database_path.go`, `internal/crypto.go`, `internal/format_v2.go`,
  `internal/logic-cli.go`, `internal/logic-tui.go`, new `internal/keys.go`; tests in
  `internal/crypto_test.go`, `internal/format_v2_test.go`, `internal/fuzz_test.go`,
  `internal/logic_cli_test.go`, `internal/logic_tui_test.go`,
  `internal/password_norm_test.go`, new `internal/golden_test.go`,
  `internal/keys_test.go`, `internal/stored_keys_test.go` and
  `internal/testdata/golden/`; `intel/cybersec.md`, `intel/maint.md`, `intel/map.md`,
  `intel/notes.md`, `intel/plan.md` and this file. Not committed.

## 2026-10-04 — Round 6 (continued): Docker image published to GitHub Packages

- **Owner request:** publish a package on GitHub from CD. The owner chose a container
  image on GitHub Packages over `.deb`/`.rpm` release assets (GitHub Packages has no
  registry for plain binaries).
- **`cd.yml`:** a new `container` job after `release` (so no image is published for a
  failed release). It builds the `Dockerfile` with `VERSION` set to the tag and OCI labels
  (`org.opencontainers.image.source` links the package to the repository; revision,
  version, licence, title, description), smoke-tests the image (`--version` carries the
  tag, `id -u` is 10001, `keys generate`/`keys list` on a volume), then on tags logs in
  with the job's token on stdin, pushes `ghcr.io/<owner>/cryptare:X.Y.Z`, plus `X.Y` and
  `latest` when the tag is the newest release of its line and overall (from
  `git ls-remote --tags`), reads the digest from the push, and attests it with
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
  `vX.Y.Z[-PRE]`; the owner name is lower-cased. **Not run:** the job itself (Docker isn't
  available here; it first runs on a manual CD run or the next tag).
- **Changed:** `.github/workflows/cd.yml` (delivered in `cryptare-round6-workflows.patch`
  with the round's other workflow changes), `README.md`, `CONTRIBUTING.md`,
  `intel/cybersec.md`, `intel/maint.md`, `intel/map.md`, `intel/notes.md`,
  `intel/plan.md` and this file. Not committed.
