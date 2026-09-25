# Implementation Plan

Active implementation plans and follow-on work. IDs refer to
[`cybersec.md`](cybersec.md) (SEC-…) and [`notes.md`](notes.md) (BUG-…, Q-…).

Last updated: 2026-09-24

Status: **In progress.**
- Done: 0.1, 0.4, 1.4 and 3.6. The CI/CD rework shipped in v1.0.1 (`5286920`,
  2026-09-24); see [Completed: CI/CD rework](#completed--cicd-rework-v101-2026-09-24).
- In progress:
  - 4.1: gosec alerts still need to be confirmed in Code Scanning and triaged.
  - 4.9: the README part waits until the owner pulls `3c80050`, a README edit made
    on GitHub.
- Next up: Phase 1 items 1.1, 1.2, 1.3 and 1.5, and owner actions 0.2, 0.3 and 0.5.

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
| 0.3 | Mark the v1.0.0 release as broken, or pull its assets | BUG-001, Q-007 | Release page updated | Open. v1.0.1 is now available as a working replacement. |
| 0.4 | Decide the CGO/release strategy | Q-001 | Decision recorded in `notes.md` | **Done:** a native CGO build per OS, verified by the v1.0.1 release (2026-09-24). |
| 0.5 | Decide the password policy and whether stored keys should be usable | Q-002, Q-004 | Decisions recorded | Open |

## Completed — CI/CD rework (v1.0.1, 2026-09-24)

The staged rework plus fixes W1–W5 was committed as `5286920` and tagged `v1.0.1`. It
covers:
- a native CGO build per OS;
- all third-party actions pinned to commit SHAs;
- gosec SARIF uploaded to Code Scanning;
- Cryptare smoke tests in CI, CD and Docker.

Step-by-step details and the pre-merge checks are in `history.md`.

**Validation:**
- **Workflow runs:** the owner reports that the CI, CD, Docker and Security workflows
  all succeeded (2026-09-24).
- **Release assets,** checked independently on 2026-09-24:
  - all five assets match `checksums.txt`;
  - every binary's embedded build info shows `CGO_ENABLED=1`,
    `main.version=v1.0.1` and the expected OS/architecture;
  - none contains go-sqlite3's CGO-less "stub" message, which the v1.0.0 binary does
    contain;
  - `cryptare_linux_amd64` was run: `--version` prints `cryptare version v1.0.1`, and
    `keys generate`/`keys list` and an encrypt/decrypt round-trip work.
- **Coverage gaps:** only linux/amd64 was run in the analysis environment. CD
  smoke-tested linux/arm64, darwin/arm64 and windows/amd64 on their own runners;
  darwin/amd64 has not been run (W7).

### Remaining follow-ups

| # | Item | Files |
|---|---|---|
| W6 | v1.0.1 ships 5 assets, with no Windows arm64 build. Remove `cryptare_windows_arm64.zip` from the README release table (part of 4.9), or restore the target on a Windows Arm runner. | `README.md` or `cd.yml` |
| W7 | darwin/amd64 is cross-compiled and not smoke-tested by CD. Its build info is correct (x86-64 Mach-O, CGO on, v1.0.1), but it hasn't been run on an Intel Mac or under Rosetta. | `cd.yml` |
| W8 | Check the Codecov dashboard: with `fail_ci_if_error: false`, a missing `CODECOV_TOKEN` wouldn't fail CI. | `ci.yml`, repo settings |

## Phase 1 — Correctness and critical security (small PRs)

| # | Change | Refs | Acceptance |
|---|---|---|---|
| 1.1 | Rewrite `readPassword`: read a full line, no echo on a TTY (`charmbracelet/x/term`) | SEC-002 | A multi-word passphrase round-trips via the prompt; a non-TTY test passes |
| 1.2 | Reject empty passwords in the core encrypt, key-blob and export paths; show the error in CLI and TUI | SEC-001 | Core, CLI and TUI tests; legacy decrypt still works |
| 1.3 | Replace the `panic("unhandled default case")` branches with no-ops or errors | BUG-002 | Tests send Left, Right, Delete, Home, End and Ctrl+U to forms without panicking |
| 1.4 | Fix release builds with a native CGO build per OS, and smoke-run each built binary in CD and CI. **Done** in v1.0.1 (2026-09-24). | BUG-001, Q-001 | CI runs each built binary; a new release tag produces working assets |
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
| 3.6 | `main.version` variable and a `--version` flag. **Done** (2026-09-24, shipped in v1.0.1). It was moved ahead of 1.4 for W4. | BUG-009, W4 |

## Phase 4 — Tooling, tests, docs

| # | Change | Refs |
|---|---|---|
| 4.1 | Pin actions to SHAs; pin gosec and upload its SARIF; triage the findings. **In progress:** pins and SARIF upload shipped in v1.0.1, and the Security workflow succeeded. Remaining: confirm gosec alerts appear in Code Scanning, triage them, and decide on govulncheck. | SEC-012 |
| 4.2 | Non-root container user; pin images by digest | SEC-013 (needs approval) |
| 4.3 | Fill test gaps: `readPassword`, `ImportKeyFromFile`, CLI export/import round-trip, extraction traversal rejection, TUI `View` | `maint.md` §5 |
| 4.4 | Add `SECURITY.md` with a private reporting channel | `CONTRIBUTING.md` |
| 4.5 | Settle the contents of `NOTICE` | Q-006 |
| 4.6 | Rename the `tasks.db` fixture in `database_path_test.go` (the Munus names in the workflows are covered by W3) | `notes.md` §3 |
| 4.7 | Replace the README's release "known issue" callout with a note that v1.0.0 is broken and v1.0.1+ works, and re-verify the install steps. **Unblocked:** v1.0.1 works. Waiting for the owner to pull `3c80050` so the README edit doesn't conflict. | BUG-001 |
| 4.8 | Doc follow-ups from the 2026-09-24 review. The owner has fixed the `AGENTS.md` typo and `golang.md`'s `gofmt -s`, and `map.md` already covers `intel/` as a directory. **Remaining (owner):** add `intel/golang.md` to the "Repository Intelligence Documents" list in `AGENTS.md`. | `AGENTS.md`, `intel/` |
| 4.9 | Post-merge doc updates. **Done 2026-09-24:** the CI/CD table in `map.md`, `maint.md` §6, and the CI description in `CONTRIBUTING.md`. **Remaining:** the README release table (W6), after the owner pulls `3c80050`. | W3, W6 |
