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
