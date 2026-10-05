# Implementation Plan

Active implementation plans and follow-on work. IDs refer to
[`cybersec.md`](cybersec.md) (SEC-…) and [`notes.md`](notes.md) (BUG-…, Q-…); the record
of how each item was done is in [`history.md`](history.md).

Last updated: 2026-10-05

## Status

- **Released:** v1.0.1 (`5286920`), v1.1.0 (`c47a94f`), v1.2.0 (`89a64e8`), v1.3.0
  (`a088f7f`) and v1.3.1 (`6773ae3`). v1.0.0 doesn't run and is left as it is (Q-007).
- **On `main`, not yet released:** round 6, committed as `c114573` and `566f96c` and
  pushed. `c114573` on its own didn't build (the new files weren't in it); `566f96c`
  adds them and builds and passes its tests. It holds Phase 6: usable stored keys, one
  password per key, golden format fixtures, `SECURITY.md`, release and image
  attestations, the Docker image on GitHub Packages and the Makefile targets.
- **Next release:** a minor version, because the format gains key source 2 (files that
  v1.3.1 and earlier can't read). Its notes come from the README's "Upgrading from
  v1.3.1 or earlier".

## Next up

1. Check CI, Docker and Security on `566f96c`: the Docker UID check (closes SEC-013), the
   golden tests on Windows (their `.gitattributes`), and the race-enabled jobs.
2. Run CD by hand on `main` (Actions → CD → Run workflow): it builds, smoke-tests and
   packages every target and the container image without publishing.
3. Tag the release (`make release VERSION=vX.Y.Z`). Then check the attestation steps
   (SEC-020), make the new GitHub Packages package public in its settings (GitHub
   creates it private), and check `docker pull ghcr.io/jabbott-iii/cryptare:<version>`
   and `gh attestation verify` on an archive and on the image.
4. Update `SECURITY.md`'s supported line to the new minor version.

## Open items

| # | Item | Refs | Next step |
|---|---|---|---|
| W7 / 6.12 | darwin/amd64 is cross-compiled and never run (CD smoke-tests every other target) | `cd.yml` | Run it once on an Intel Mac or under Rosetta (owner) |
| W8 | With `fail_ci_if_error: false`, a missing `CODECOV_TOKEN` wouldn't fail CI | `ci.yml` | Check the Codecov dashboard (owner) |
| 4.3 | Test gap: the TUI `View` (about 41% covered) | `maint.md` §5 | Tests through `View()` output |
| 6.13 | Optional: macOS notarisation and Windows code signing | SEC-020 | Owner decision, by audience |
| — | A longer run of `FuzzExtractTar` | `notes.md` §3 | `make fuzz FUZZ=FuzzExtractTar FUZZTIME=10m` |
| SEC-013 | The UID check's first passing CI run | `cybersec.md` | Step 1 above |
| SEC-020 | The first tagged release with attestations | `cybersec.md` | Step 3 above |

## Principles

- **One issue per PR**, following `CONTRIBUTING.md`: open an issue, get confirmation,
  then send a small PR.
- **Regression test first:** each fix comes with a test that fails before it.
- **Compatibility:** keep the CLI flags and on-disk formats (`maint.md` §3) working, and
  the golden fixtures opening. Anything breaking needs explicit approval.
- **Approval required:** changes to CI, the Dockerfile or dependencies need explicit
  owner approval (`AGENTS.md`).
- **`make check-all`** before a pull request.

## Completed work

"Released in" names the first release that carries the change.

### Phase 0 — Owner actions

| # | Action | Refs | Done |
|---|---|---|---|
| 0.1 | Remove `cryptare.db` from `main`; decide on rewriting history (no, Q-005) | SEC-003, Q-005 | `aa27461` (pushed 2026-09-23) |
| 0.2 | Discard the key held in the committed database | SEC-003 | Owner, 2026-09-26 |
| 0.3 | Leave v1.0.0 as it is; the README says it is broken | BUG-001, Q-007 | 2026-09-26 |
| 0.4 | Release strategy: a native CGO build per OS | Q-001 | v1.0.1 |
| 0.5 | Password policy, and stored keys to be usable | Q-002, Q-004 | `b415ffc` (v1.2.0) |

### CI/CD rework and follow-ups

| # | Change | Done |
|---|---|---|
| W1–W5 | Native CGO builds, SHA-pinned actions, gosec SARIF upload, smoke tests in CI, CD and Docker | `5286920` (v1.0.1) |
| W6 | README release table without the Windows ARM64 row | `eb330a3` (v1.2.0) |
| W9 | Smoke-test passwords long enough for the policy | `b415ffc` (v1.2.0) |
| W10 | Actions moved to their Node.js 24 releases; CodeQL v4 | `b520b97` (v1.2.0) |
| W11 | Smoke tests use `--password-file` | `3d9384e` (v1.2.0) |

### Phase 1 — Correctness and critical security

| # | Change | Refs | Done |
|---|---|---|---|
| 1.1 | Password prompt reads the whole line without echo | SEC-002 | `ed46150` (v1.1.0) |
| 1.1a | Terminal restored when the prompt is interrupted | BUG-012 | `ed46150` (v1.1.0) |
| 1.2 | Empty passwords refused on every encryption path | SEC-001 | `ed46150` (v1.1.0) |
| 1.3 | No panics on unhandled keys in TUI forms | BUG-002 | `18ea97a` (v1.1.0) |
| 1.4 | Working release builds, smoke-run in CI and CD | BUG-001, Q-001 | `5286920` (v1.0.1) |
| 1.5 | No output over its own input; existing outputs only with `--force` | BUG-003, BUG-004, BUG-013 | `c47a94f` (v1.1.0) |

### Phase 2 — Hardening

| # | Change | Refs | Done |
|---|---|---|---|
| 2.1 | Extraction size and entry limits, with clean-up | SEC-007 | `b415ffc` (v1.2.0) |
| 2.2 | Atomic, confined, owner-only extraction | SEC-008, BUG-014 | `b415ffc`, `d751967` (v1.2.0) |
| 2.3 | No plaintext archive in the temp folder (superseded by 3.2) | SEC-006 | `d683739` (v1.2.0) |
| 2.4 | SQLite `secure_delete` | SEC-009 | `d683739` (v1.2.0) |
| 2.4a | Optional one-time `VACUUM` | SEC-009 | Declined |
| 2.5 | Key store opened only when needed, created 0600 | SEC-010, BUG-005 | `d683739` (v1.2.0) |
| 2.6 | `--password-file`, and a warning for `--password` | SEC-004 | `e512844` (v1.2.0) |
| 2.7 | Imported key metadata validated | SEC-011 | `3d9384e` (v1.2.0) |
| 2.8 | Archiving through an `os.Root` | SEC-014 | `3d9384e` (v1.2.0) |
| 2.9 | Case-insensitive extensions; export reports the file written | BUG-006, BUG-007 | `3d9384e` (v1.2.0) |
| 2.10 | One TUI action at a time | BUG-008 | `3d9384e` (v1.2.0) |
| 2.11 | Atomic single-file writes | BUG-004 | `b4cd66f` (v1.2.0) |
| 2.12 | Optional TUI overwrite choice | BUG-004 | Declined |

### Phase 3 — Formats and architecture

| # | Change | Refs | Done |
|---|---|---|---|
| 3.1 | Versioned header, Argon2id, legacy reads | SEC-005 | `0c57aef` (v1.2.0) |
| 3.2 | Streaming, chunked authenticated encryption | BUG-010 | `0c57aef` (v1.2.0) |
| 3.3 | Key flows shared by the CLI and TUI (`keys.go`) | `maint.md` §2 | `c114573`, `566f96c` |
| 3.4 | Stored keys encrypt files (key source 2) | Q-002, Q-014, BUG-011 | `c114573`, `566f96c` |
| 3.5 | Per-user key store with a notice-only migration; `keys path` | Q-003, SEC-010, SEC-016 | `a088f7f` (v1.3.0) |
| 3.6 | `--version` | BUG-009 | `5286920` (v1.0.1) |

### Phase 4 — Tooling, tests, docs

| # | Change | Refs | Done |
|---|---|---|---|
| 4.1 | gosec triaged in Code Scanning; govulncheck in CI | SEC-012 | `5286920`, `4ba516a`, `6a5fcb1` (v1.2.0) |
| 4.2 | Non-root container; images pinned by digest | SEC-013 | `eb330a3`, `89a64e8` (v1.2.0) |
| 4.3 | Test gaps | `maint.md` §5 | Partly; the TUI `View` is open |
| 4.4 | `SECURITY.md` with private reporting | — | `c114573` |
| 4.5 | `NOTICE` and `THIRD_PARTY_LICENSES.txt` in release archives | Q-006 | `a088f7f` (v1.3.0) |
| 4.6 | Test fixture renamed (`keys.db`) | — | `a088f7f` (v1.3.0) |
| 4.7 | README release notes and install steps corrected | BUG-001 | `eb330a3`, `89a64e8` (v1.2.0) |
| 4.8 | `intel/golang.md` listed in `AGENTS.md` | — | 2026-10-05 (docs round) |
| 4.9 | Post-merge doc updates | W3, W6 | `eb330a3`, `89a64e8` (v1.2.0) |
| 4.10 | Linux releases linked against musl; C library notices | Q-013 | `6773ae3` (v1.3.1) |

### Phase 5 — Remediation from the 2026-10-03 analysis

| # | Change | Refs | Done |
|---|---|---|---|
| 5.1 | Go 1.26.8 toolchain, govulncheck, Dependabot | SEC-018 | `6a5fcb1`, `b3278ea` (v1.2.0) |
| 5.2 | `keys export` output safety | BUG-017 | `6a5fcb1`, `b3278ea` (v1.2.0) |
| 5.3 | Folder encryption output kept outside the folder | BUG-016, BUG-013 | `6a5fcb1`, `b3278ea` (v1.2.0) |
| 5.4 | Latin-1-safe gzip names | BUG-015 | `6a5fcb1`, `b3278ea` (v1.2.0) |
| 5.5 | Cancellation and clean-up on interrupt | SEC-015, BUG-018 | `6a5fcb1`, `b3278ea` (v1.2.0) |
| 5.6 | Key store trust check; escaped listings | SEC-016, SEC-011 | `4ba516a` (v1.2.0) |
| 5.7 | Quiet database layer; typed errors | SEC-017 | `eb330a3` (v1.2.0) |
| 5.8 | Database path handling | SEC-010 | `eb330a3` (v1.2.0) |
| 5.9 | Repository and container hygiene | SEC-003, SEC-013 | `eb330a3`, `89a64e8` (v1.2.0) |
| 5.10 | Windows permissions documented (no ACLs, Q-012) | SEC-019 | `eb330a3` (v1.2.0) |
| 5.11 | Containment checks by file identity | BUG-019 | `eb330a3` (v1.2.0) |
| 5.12 | Compression options validated | BUG-020 | `eb330a3` (v1.2.0) |
| 5.13 | Shorter temporary names | BUG-021 | `eb330a3` (v1.2.0) |
| 5.14 | `--password ""` counts as given | BUG-023 | `eb330a3` (v1.2.0) |
| 5.15 | Size caps for whole-file reads | BUG-024 | `eb330a3` (v1.2.0) |
| 5.16 | `decompress --raw` | BUG-022, Q-011 | `eb330a3` (v1.2.0) |
| 5.17 | Export protected by the key's master password, checked | BUG-011, Q-015 | `c114573`, `566f96c` |
| 5.18 | golangci-lint, gosec and govulncheck run on every push | SEC-012, SEC-018 | 2026-10-04 |
| 5.19 | `go test -race` in CI on Linux and macOS | `maint.md` §7 | `eb330a3` (v1.2.0) |
| 5.20 | README and `CONTRIBUTING.md` claims corrected | `notes.md` §3 | `eb330a3` (v1.2.0) |
| 5.21 | Fuzz targets for the untrusted-input readers | `maint.md` §5 | `eb330a3` (v1.2.0) |

### Phase 6 — Production-readiness review (2026-10-04)

| # | Change | Refs | Done |
|---|---|---|---|
| 6.1 | `SECURITY.md` replaces GitHub's template | 4.4 | `c114573` |
| 6.2 | Usable stored keys; one password per key | 3.3, 3.4, 5.17, Q-014, Q-015 | `c114573`, `566f96c` |
| 6.3 | Golden format fixtures from released binaries | `maint.md` §3 | `566f96c` |
| 6.4 | README installs the latest release | SEC-018 | `c114573` |
| 6.5 | Usage text only for command-line mistakes | — | `c114573` |
| 6.6 | Build provenance attestations on release files | SEC-020 | `c114573` |
| 6.7 | Dependabot for the Dockerfile's images | SEC-018, SEC-013 | `c114573` |
| 6.8 | `docker.yml` checks the image runs as UID 10001 | SEC-013 | `c114573` |
| 6.9 | SEC-018 closed on the v1.3.1 binaries | SEC-018 | 2026-10-04 |
| 6.10 | Why `FuzzExtractTar` seemed to stall, documented | 5.21 | `c114573` |
| 6.11 | `notes.md` and this plan brought up to date | — | `c114573` |
| 6.14 | Docker image published to GitHub Packages on release, with an attestation | SEC-020, SEC-013 | `c114573` |
| 6.15 | Makefile targets for contributors (`make check`, `make check-all`, …) | `maint.md` §7 | `566f96c` |
