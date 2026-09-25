# Implementation Plan

Active implementation plans and follow-on work. IDs refer to
[`cybersec.md`](cybersec.md) (SEC-…) and [`notes.md`](notes.md) (BUG-…, Q-…).

Last updated: 2026-09-24

Status: **In progress.**
- 0.1 and 3.6 (`--version`, done early to resolve W4) are done. 3.6 is not yet committed.
- A CI/CD rework covering 1.4 and most of 4.1 is staged in the working tree, not yet
  committed. The 2026-09-24 review found blocking problems in it; see
  [Current work](#current-work--cicd-rework-staged-reviewed-2026-09-24).
- The fixes for W1–W3 and W5 are written and validated, but not yet in the tree. They
  are delivered as `cryptare-workflow-fixes.patch` for the owner to apply, because
  `.github/workflows/` is write-protected for the assistant's remote tools on this
  machine.
- No other items have started.

## Principles

- **One issue per PR**, following `CONTRIBUTING.md`: open an issue, get confirmation,
  then send a small PR.
- **Regression test first:** each fix comes with a test that fails before it.
- **Compatibility:** keep the CLI flags and on-disk formats (`maint.md` §3) working.
  Anything breaking needs explicit approval.
- **Approval required:** changes to CI, the Dockerfile or dependencies need explicit
  owner approval (`AGENTS.md`).

## Phase 0 — Owner actions (no code)

| # | Action | Refs | Done when | Status |
|---|---|---|---|---|
| 0.1 | Push `aa27461` (removes `cryptare.db`); decide on purging history | SEC-003, Q-005 | `origin/main` no longer tracks `cryptare.db` | **Done** (pushed 2026-09-23; verified 2026-09-24). Purging history (Q-005) is still undecided. |
| 0.2 | Stop using the master password that was used with the committed DB; discard that key | SEC-003 | Owner confirms | Open |
| 0.3 | Mark the v1.0.0 release as broken, or pull its assets | BUG-001, Q-007 | Release page updated | Open |
| 0.4 | Decide the CGO/release strategy | Q-001 | Decision recorded in `notes.md` | **Direction chosen:** a native CGO build per OS (staged `cd.yml`). Done once the rework passes validation. |
| 0.5 | Decide the password policy and whether stored keys should be usable | Q-002, Q-004 | Decisions recorded | Open |

## Current work — CI/CD rework (staged, reviewed 2026-09-24)

**Scope.** Staged, uncommitted changes to `ci.yml`, `cd.yml`, `docker.yml` and
`security.yml`. They implement 1.4 and most of 4.1.

### Verified as correct

- **Action pins.** All 9 third-party actions are pinned to full commit SHAs, and each
  SHA matches its release tag on GitHub:
  - `actions/checkout` v4.4.0
  - `actions/setup-go` v5.6.0
  - `golangci/golangci-lint-action` v9.3.0
  - `codecov/codecov-action` v5.5.5
  - `actions/upload-artifact` v4.6.2
  - `actions/download-artifact` v4.3.0
  - `softprops/action-gh-release` v2.6.2
  - `github/codeql-action` v3.38.1
  - `securego/gosec` v2.29.0
- **gosec.** Pinned to v2.29.0, and its SARIF is now uploaded to Code Scanning
  (category `gosec`).
- **Permissions.** CD defaults to `contents: read`; only the release job gets
  `contents: write`.
- **CGO builds.** CD builds each target with CGO on a runner of the matching OS. The
  static Linux flags work for Cryptare:
  - tags `sqlite_omit_load_extension,osusergo,netgo`;
  - `-linkmode external -extldflags -static`.

  A local linux/amd64 build with these flags was reported "statically linked" by
  `file` and passed a `keys generate` / `keys list` check.
- **Lint.** actionlint 1.7.12 reports no errors. shellcheck wasn't available, so the
  shell snippets were reviewed by hand.

### Blocking fixes (required before merging)

Status 2026-09-24: W4 is resolved. W1–W3 and W5 are fixed in
`cryptare-workflow-fixes.patch`, which is validated and waiting for the owner to apply
it (`git apply cryptare-workflow-fixes.patch`).

The smoke tests and names were carried over from the Munus project.

| # | Fix | Files | Evidence |
|---|---|---|---|
| W1 | **Fixed in patch.** Replace the Munus smoke commands (`--version`, `add --title …`, `list`) with Cryptare checks: `keys generate --password <throwaway>` then `keys list \| grep -q AES-256-GCM`, plus an encrypt → decrypt → `cmp` round-trip | `ci.yml`, `cd.yml`, `docker.yml` | Each Munus command exits 1 against a Cryptare build. The proposed checks pass on a CGO build and fail on a CGO-less one, so they catch BUG-001. |
| W2 | **Fixed in patch.** Set `CRYPTARE_DB_PATH` instead of `MUNUS_DB_PATH` | `ci.yml`, `cd.yml` | Cryptare ignores `MUNUS_DB_PATH`, so the smoke run would create `cryptare.db` in the checkout. |
| W3 | **Fixed in patch.** Rename the Munus artifacts to Cryptare ones: `dist/munus_*` → `dist/cryptare_*` (build output and the packaging glob), the `munus-ci` binary, the `munus:` image tag, and the `munus-smoke` volume | `cd.yml`, `ci.yml`, `docker.yml` | The README documents `cryptare_<os>_<arch>` release assets. |
| W4 | ~~`--version` doesn't exist yet.~~ **Resolved 2026-09-24:** the owner chose to add the flag (3.6 done), so the smoke steps can keep calling `--version`. | `ci.yml`, `cd.yml` | A build with `-X main.version=v1.0.1` prints `cryptare version v1.0.1`; a local build prints `cryptare version dev`. |
| W5 | **Fixed in patch.** The `docker.yml` comment says the image runs as a non-root user, but the `Dockerfile` still runs as root. Fix the comment, or do 4.2 (SEC-013) in the same change. | `docker.yml`, `Dockerfile` | The `Dockerfile` is unchanged. |

### Follow-ups (not blocking)

| # | Item | Files |
|---|---|---|
| W6 | Windows arm64 is no longer built. Either restore it (for example on a Windows Arm runner) or remove `cryptare_windows_arm64.zip` from the README release table. | `cd.yml`, `README.md` |
| W7 | darwin/amd64 is cross-compiled on Apple silicon with its smoke test turned off. Manually check that asset from the first release run. | `cd.yml` |
| W8 | The Codecov upload needs a `CODECOV_TOKEN` secret (or tokenless uploads enabled for the repo). Because `fail_ci_if_error: false` is set, a failed upload won't fail CI, so check the Codecov dashboard. | `ci.yml`, repo settings |

### Validation before tagging a release

1. Run actionlint on the fixed workflows.
2. Push a branch. CI must pass on Ubuntu, Windows and macOS, including the new smoke
   step.
3. Run CD manually (`workflow_dispatch`). All build jobs and smoke tests must pass.
   The per-target binaries can be downloaded from the run; the GitHub Release step
   runs only for tags.
4. Tag a new version (for example v1.0.1). Download the Linux, macOS and Windows
   assets and run `keys list` with each. This closes BUG-001 and completes 0.4 and
   1.4.

## Phase 1 — Correctness and critical security (small PRs)

| # | Change | Refs | Acceptance |
|---|---|---|---|
| 1.1 | Rewrite `readPassword`: read a full line, no echo on a TTY (`charmbracelet/x/term`) | SEC-002 | A multi-word passphrase round-trips via the prompt; a non-TTY test passes |
| 1.2 | Reject empty passwords in the core encrypt, key-blob and export paths; show the error in CLI and TUI | SEC-001 | Core, CLI and TUI tests; legacy decrypt still works |
| 1.3 | Replace the `panic("unhandled default case")` branches with no-ops or errors | BUG-002 | Tests send Left, Right, Delete, Home, End and Ctrl+U to forms without panicking |
| 1.4 | Fix release builds with a native CGO build per OS, and smoke-run each built binary in CD and CI. **In progress:** the staged rework plus the W1–W3/W5 patch passes local checks. Next: apply the patch, then run the validation steps above. | BUG-001, Q-001 | CI runs each built binary; a new release tag produces working assets |
| 1.5 | Guard against `src == dst` and refuse to overwrite existing outputs unless forced (new flag) | BUG-003, BUG-004 | Tests for the same-path and existing-output cases |

## Phase 2 — Hardening

| # | Change | Refs |
|---|---|---|
| 2.1 | Extraction size and entry limits, with clean-up on abort | SEC-007 |
| 2.2 | Extract via `os.Root`; mask archive modes | SEC-008 |
| 2.3 | Build the directory-encryption tar.gz in memory (no temp plaintext) | SEC-006 |
| 2.4 | Enable SQLite `secure_delete`; align README wording | SEC-009 |
| 2.5 | Open the DB lazily (keys commands and TUI only); create it with mode 0600 | SEC-010, BUG-005 |
| 2.6 | Add `--password-file` / `--password-stdin`; warn when `--password` is used; update README examples | SEC-004 |
| 2.7 | Validate imported key metadata | SEC-011 |
| 2.8 | Root-scoped opens when archiving | SEC-014 |
| 2.9 | Fix case-insensitive extension handling and the export filename reporting | BUG-007, BUG-006 |
| 2.10 | Gate TUI actions on `busy` | BUG-008 |

## Phase 3 — Formats and architecture

| # | Change | Refs |
|---|---|---|
| 3.1 | Versioned file header, Argon2id (or PBKDF2 ≥ 600k), legacy read path; switch to stdlib `crypto/pbkdf2` | SEC-005 |
| 3.2 | Streaming, chunked authenticated encryption for large files (depends on 3.1) | BUG-010 |
| 3.3 | Move the key generate/export/import flows into shared core functions used by both CLI and TUI | `maint.md` §2 |
| 3.4 | Wire stored keys into encrypt/decrypt, if Q-002 = yes | Q-002, BUG-011 |
| 3.5 | Per-user default DB path plus migration, if Q-003 = yes | Q-003 |
| 3.6 | `main.version` variable and a `--version` flag. **Done 2026-09-24 (uncommitted)**, moved ahead of 1.4 for W4. | BUG-009, W4 |

## Phase 4 — Tooling, tests, docs

| # | Change | Refs |
|---|---|---|
| 4.1 | Pin actions to SHAs; pin gosec and upload its SARIF; triage the findings. **In progress:** pins and SARIF upload are staged and verified. Remaining: merge, confirm alerts appear in Code Scanning, triage them. | SEC-012 |
| 4.2 | Non-root container user; pin images by digest (see W5) | SEC-013 (needs approval) |
| 4.3 | Fill test gaps: `readPassword`, `ImportKeyFromFile`, CLI export/import round-trip, extraction traversal rejection, TUI `View` | `maint.md` §5 |
| 4.4 | Add `SECURITY.md` with a private reporting channel | `CONTRIBUTING.md` |
| 4.5 | Settle the contents of `NOTICE` | Q-006 |
| 4.6 | Rename the `tasks.db` fixture in `database_path_test.go` (the Munus names in the workflows are covered by W3) | `notes.md` §3 |
| 4.7 | Once releases work, remove the README's "known issue" callout and re-verify the install steps | BUG-001 |
| 4.8 | Doc follow-ups from the 2026-09-24 review: fix the `--` typo on the `intel/golang.md` line in `AGENTS.md`, and add `golang.md` to its "Repository Intelligence Documents" list; add `golang.md` to the structure in `map.md`; optionally change `golang.md`'s `gofmt -w` to the repo's `gofmt -s -w` | `AGENTS.md`, `intel/` |
| 4.9 | After the CI/CD rework merges, update the CI/CD table in `map.md`, the release notes in `maint.md` §6, the CI description in `CONTRIBUTING.md`, and the README release table | W3, W6 |
