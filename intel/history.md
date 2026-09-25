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
