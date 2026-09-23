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
