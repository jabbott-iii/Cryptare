# Implementation Plan

Active implementation plans and follow-on work. IDs refer to
[`cybersec.md`](cybersec.md) (SEC-…) and [`notes.md`](notes.md) (BUG-…, Q-…).

Last updated: 2026-09-23
Status: **Proposed.** This is the baseline plan from the 2026-09-23 analysis. The
owner has not prioritized it yet, and no items are in progress.

## Principles

- **One issue per PR**, following `CONTRIBUTING.md`: open an issue, get confirmation,
  then send a small PR.
- **Regression test first:** each fix comes with a test that fails before it.
- **Compatibility:** keep the CLI flags and on-disk formats (`maint.md` §3) working.
  Anything breaking needs explicit approval.
- **Approval required:** changes to CI, the Dockerfile or dependencies need explicit
  owner approval (`AGENTS.md`).

## Phase 0 — Owner actions (no code)

| # | Action | Refs | Done when |
|---|---|---|---|
| 0.1 | Push `aa27461` (removes `cryptare.db`); decide on purging history | SEC-003, Q-005 | `origin/main` no longer tracks `cryptare.db` |
| 0.2 | Stop using the master password that was used with the committed DB; discard that key | SEC-003 | Owner confirms |
| 0.3 | Mark the v1.0.0 release as broken, or pull its assets | BUG-001, Q-007 | Release page updated |
| 0.4 | Decide the CGO/release strategy | Q-001 | Decision recorded in `notes.md` |
| 0.5 | Decide the password policy and whether stored keys should be usable | Q-002, Q-004 | Decisions recorded |

## Phase 1 — Correctness and critical security (small PRs)

| # | Change | Refs | Acceptance |
|---|---|---|---|
| 1.1 | Rewrite `readPassword`: read a full line, no echo on a TTY (`charmbracelet/x/term`) | SEC-002 | A multi-word passphrase round-trips via the prompt; a non-TTY test passes |
| 1.2 | Reject empty passwords in the core encrypt, key-blob and export paths; show the error in CLI and TUI | SEC-001 | Core, CLI and TUI tests; legacy decrypt still works |
| 1.3 | Replace the `panic("unhandled default case")` branches with no-ops or errors | BUG-002 | Tests send Left, Right, Delete, Home, End and Ctrl+U to forms without panicking |
| 1.4 | Fix release builds according to Q-001, and add a post-build smoke run (`--help`) in CD and CI | BUG-001 | CI runs each built binary; a new release tag produces working assets |
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
| 3.6 | `main.version` variable and a `--version` flag | BUG-009 |

## Phase 4 — Tooling, tests, docs

| # | Change | Refs |
|---|---|---|
| 4.1 | Pin actions to SHAs; pin gosec and upload its SARIF; triage the findings | SEC-012 (needs CI approval) |
| 4.2 | Non-root container user; pin images by digest | SEC-013 (needs approval) |
| 4.3 | Fill test gaps: `readPassword`, `ImportKeyFromFile`, CLI export/import round-trip, extraction traversal rejection, TUI `View` | `maint.md` §5 |
| 4.4 | Add `SECURITY.md` with a private reporting channel | `CONTRIBUTING.md` |
| 4.5 | Settle the contents of `NOTICE` | Q-006 |
| 4.6 | Fix the `munus` image name in `docker.yml` and the `tasks.db` name in the test | `notes.md` §3 |
| 4.7 | Once releases work, remove the README's "known issue" callout and re-verify the install steps | BUG-001 |
