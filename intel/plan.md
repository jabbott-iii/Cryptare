# Implementation Plan

Active implementation plans and follow-on work. IDs refer to
[`cybersec.md`](cybersec.md) (SEC-…) and [`notes.md`](notes.md) (BUG-…, Q-…).

Last updated: 2026-10-04

Status: **In progress.**
- **Released:**
  - v1.0.1 (`5286920`): the CI/CD rework; see
    [Completed: CI/CD rework](#completed--cicd-rework-v101-2026-09-24).
  - v1.1.0 (`c47a94f`, tagged 2026-09-25): 1.1, 1.1a, 1.2, 1.3 and 1.5. Phase 1 is
    complete.
  - v1.2.0 (`89a64e8`, released by CD #4 on 2026-10-04, with the drafted notes): Phases
    2 and 3.1/3.2, and Phase 5 except 5.17, with 4.2, 4.7 and 4.9.
  - v1.3.0 (`a088f7f`, released by CD #5 on 2026-10-04): 3.5, 4.5, 4.6, the SEC-003 CI
    guard and Q-009's password normalisation. SEC-003, SEC-010 and SEC-016 are closed.
  - v1.3.1 (`6773ae3`, released by CD on 2026-10-04): round 5, 4.10 (musl Linux builds,
    the Windows DLL check, C library notices) and PR #26's `x/crypto` 0.57.0 update.
    Its binaries report go1.26.8, which closed SEC-018.
- **Owner decisions (2026-09-26):**
  - 0.2 and 0.3 are done.
  - Stored keys should be usable (Q-002 = yes), so 3.4 is approved.
  - The password policy (Q-004) was still open. On 2026-09-27 the owner asked for a
    default policy (0.5).
- **Committed:** 2.11, atomic writes for single-file outputs (`b4cd66f`, 2026-09-26).
- **Committed 2026-09-27:**
  - `b415ffc`: 0.5 (the default password policy, with the W9 workflow patch) and 2.1
    (extraction limits, with the atomic-extraction part of 2.2). CI, Docker and
    Security all passed, so SEC-001 and SEC-007 are closed;
  - `d751967`: the rest of 2.2 (`os.Root` extraction and owner-only permissions);
  - `d683739`: 2.3 (directory-encryption archive built in memory), 2.4 (SQLite secure
    delete) and 2.5 (database opened only when needed, created 0600);
  - `a5edc91`: a test-only fix for Windows CI. CI, Docker and Security are green, so
    SEC-006, SEC-008 and SEC-009 are closed;
  - `e512844`: 2.6 (`--password-file` and the `--password` warning);
  - `3d9384e`: 2.7–2.10 and the W11 workflow patch. Phase 2 is complete.
  - `b520b97`: W10, which moves every pinned action to its Node.js 24 release.
- **Owner decisions (2026-09-27):**
  - W10 is approved.
  - The optional items 2.4a and 2.12 are declined.
  - For 3.1, the owner chose:
    - Argon2id with 64 MiB, 3 passes and 4 lanes;
    - to build the versioned header and the chunked streaming (3.2) together;
    - to write every kind of artifact in the new format: files, folders, stored keys
      and key exports;
    - to keep old formats readable with no time limit, and no migration command.
- **Committed 2026-09-27 (`0c57aef`):** 3.1 and 3.2, the version 2 format
  (`format_v2.go`). Its CI results haven't been reported yet.
- **Committed 2026-10-03/04:** `4ba516a` (the gosec dispositions and 5.6 steps 1–2),
  `6a5fcb1` and `b3278ea` (Tier A, 5.1–5.5). CI #136–#138, Docker #18–#20 and
  Security #142–#144 passed (checked on GitHub 2026-10-04); Code Scanning has 0 open
  alerts. SEC-004, SEC-005, SEC-011, SEC-012, SEC-014 and SEC-015 are closed.
- **Committed 2026-10-04:** `eb330a3` and `89a64e8` (round 3: Tier B, Tier C except
  5.17, Tier D, 4.2, 4.7, 4.9). CI #141/#142, Docker #22/#23 and Security #147/#148
  passed, and v1.2.0 was released from `89a64e8`. SEC-002, SEC-017 and SEC-019 are
  closed.
- **Owner decisions (2026-10-04):** Q-003 = move the default key store to a per-user
  data folder with a notice-only migration (3.5); Q-005 = no history purge, delete the
  Copilot branch (already gone) and add a CI guard; Q-006 = `NOTICE` rewrite plus a
  generated licence file in the release archives (4.5); Q-008 = keep the read limits;
  Q-009 = NFKC normalisation, byte order mark handling and `golang.org/x/text` as a
  direct dependency.
- **Owner decisions (2026-10-04, round 5):** Q-013 = build the Linux releases against
  musl and ship the C library notices; leave the published releases as they are. PR
  #26 = take the `golang.org/x/crypto` 0.57.0 update (with `x/text` 0.42.0), add a
  normalisation known-answer test and note the change in v1.3.1's release notes.
- **Committed 2026-10-04 (round 5, `6773ae3`, v1.3.1):** 4.10 (musl Linux builds, the
  Windows DLL check, C library notices) and PR #26's update with
  `TestNormalizePasswordKnownAnswers`.
- **Owner decisions (2026-10-04, round 6):** build 3.3 and 3.4 now; Q-014 = a file
  encrypted with a stored key records the key's ID in its header; Q-015 = one password
  per key (exports use the key's master password); approved: attestations in `cd.yml`,
  the Dependabot `docker` entry and the UID check in `docker.yml`; `SECURITY.md`:
  acknowledge within 7 days, latest minor line (1.3.x) supported.
- **Done 2026-10-04, not yet committed (round 6):** Phase 6 below, with 3.3, 3.4, 4.4
  and 5.17.
- **2026-10-03 analysis:** five new security items (SEC-015–SEC-019), ten new defects
  (BUG-015–BUG-024), gaps in five existing items and five owner questions
  (Q-008–Q-012). They are planned in
  [Phase 5](#phase-5--remediation-from-the-2026-10-03-analysis).
- **Next up:**
  1. Review round 6. It changes the file format (key source 2) and `keys export`'s
     behaviour, so read the README's "Upgrading from v1.3.1 or earlier" note first, and
     confirm the golden-fixture naming noted under SEC-003. Then commit it.
  2. Watch the first CI, Docker and Security runs: the Docker UID check (closes
     SEC-013), the golden tests on Windows (the fixtures' `.gitattributes`), and the
     race-enabled jobs.
  3. Run CD by hand on `main` before tagging, then release the next version (a minor
     version, since the format gains key source 2) with notes drawn from the README's
     v1.3.1 upgrade note. Check its attestation step and run `gh attestation verify` on
     an asset (SEC-020).
  4. After that release's CD run: the image is on GitHub Packages
     (`ghcr.io/jabbott-iii/cryptare`); GitHub creates a new package as private, so make
     it public in the package's settings, then check `docker pull` and
     `gh attestation verify oci://ghcr.io/jabbott-iii/cryptare:<version> --repo jabbott-iii/Cryptare`.
  5. When the new minor version ships, update `SECURITY.md`'s supported line.
  6. Still open: W7 (darwin/amd64 never run), 4.3's TUI `View` tests, 4.8, and a longer
     fuzz run of `FuzzExtractTar` with `-fuzzminimizetime` set (`notes.md` §3).

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
| 0.1 | Push `aa27461` (removes `cryptare.db`); decide on purging history | SEC-003, Q-005 | `origin/main` no longer tracks `cryptare.db` | **Done** (pushed 2026-09-23; verified 2026-09-24). Q-005 answered 2026-10-04: no purge; a CI step now refuses tracked key material (SEC-003). |
| 0.2 | Stop using the master password that was used with the committed DB; discard that key | SEC-003 | Owner confirms | **Done** (2026-09-26): the owner discarded the key. |
| 0.3 | Mark the v1.0.0 release as broken, or pull its assets | BUG-001, Q-007 | Release page updated | **Done** (2026-09-26): v1.1.0 is released and verified. The owner decided to leave v1.0.0 as it is. |
| 0.4 | Decide the CGO/release strategy | Q-001 | Decision recorded in `notes.md` | **Done:** a native CGO build per OS, verified by the v1.0.1 release (2026-09-24). |
| 0.5 | Decide the password policy and whether stored keys should be usable | Q-002, Q-004 | Decisions recorded | **Done** (2026-09-27, `b415ffc`): stored keys should be usable (Q-002 = yes, 2026-09-26). For Q-004 the owner asked for a default policy, now implemented: new passwords need at least 15 characters (Unicode code points) and must not be one repeated character, with no composition rules; typed passwords are confirmed; decrypt and import accept any password. See SEC-001 and SEC-002. |

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
| W9 | The smoke tests pass `ci-smoke`, `release-smoke` and `docker-smoke`, which the password policy (0.5) rejects. `cryptare-password-policy-workflows.patch` lengthens them to `ci-smoke-passphrase`, `release-smoke-passphrase` and `docker-smoke-passphrase` (test-only values, not secrets). **Applied by the owner and committed with the policy change (`b415ffc`).** | `ci.yml`, `cd.yml`, `docker.yml` |
| W10 | CI annotations on 2026-09-27: several pinned actions target Node.js 20, which GitHub now forces onto Node.js 24; `github/codeql-action` v3 must move to v4 before its December 2026 deprecation; `ubuntu-latest` moves to Ubuntu 26 from 2026-10-19. Update the pinned SHAs (needs owner approval as a CI change). **Approved and done 2026-09-27:** delivered as `cryptare-w10-action-updates.patch`: checkout v7.0.1, setup-go v7.0.0, upload-artifact v7.0.1, download-artifact v8.0.1, codecov-action v7.1.1, codeql-action v4.38.2 and action-gh-release v3.0.3 (golangci-lint-action v9.3.0 is already on Node.js 24; gosec is a Docker action). Runner labels are unchanged, so `ubuntu-latest` moves to Ubuntu 26 on 2026-10-19 on its own. Committed by the owner as `b520b97`; CI, Docker and Security passed on it and on every later commit (checked 2026-10-04). | `ci.yml`, `cd.yml`, `docker.yml`, `security.yml` |
| W11 | Switch the CI, CD and Docker smoke tests from `--password` to `--password-file`, so they exercise it and their logs lose the new warning. **Approved and done 2026-09-27:** delivered as `cryptare-w11-password-file-smoke.patch` for the owner to apply (2.6 is in `e512844`). | `ci.yml`, `cd.yml`, `docker.yml` |

## Phase 1 — Correctness and critical security (small PRs)

| # | Change | Refs | Acceptance |
|---|---|---|---|
| 1.1 | Rewrite `readPassword`: read a full line, no echo on a TTY (`charmbracelet/x/term`) **Done 2026-09-24** (`ed46150`), with owner approval for the `go.mod` change. | SEC-002 | A multi-word passphrase round-trips via the prompt; a non-TTY test passes |
| 1.1a | Restore the terminal when Ctrl+C interrupts the hidden password prompt. This needs a signal handler and a short-lived goroutine, which `golang.md` allows only when required. **Done 2026-09-24** (`ed46150`). | BUG-012 | In a pseudo-terminal, echo is on again after Ctrl+C at the prompt; exit status 130 |
| 1.2 | Reject empty passwords in the core encrypt, key-blob and export paths; show the error in CLI and TUI **Done 2026-09-24** (`ed46150`). The minimum-length policy and confirmation followed on 2026-09-27 (0.5). | SEC-001 | Core, CLI and TUI tests; legacy decrypt still works |
| 1.3 | Replace the `panic("unhandled default case")` branches with no-ops or errors. **Done 2026-09-24** (`18ea97a`). | BUG-002 | Tests send Left, Right, Delete, Home, End and Ctrl+U to forms without panicking |
| 1.4 | Fix release builds with a native CGO build per OS, and smoke-run each built binary in CD and CI. **Done** in v1.0.1 (2026-09-24). | BUG-001, Q-001 | CI runs each built binary; a new release tag produces working assets |
| 1.5 | Guard against `src == dst` and refuse to overwrite existing outputs unless forced (new flag) **Done 2026-09-24** (`c47a94f`, v1.1.0). Also refuses compressing a folder into an archive inside itself (BUG-013). Atomic writes, the rest of BUG-004, moved to 2.11. | BUG-003, BUG-004 | Tests for the same-path and existing-output cases |

## Phase 2 — Hardening

| # | Change | Refs |
|---|---|---|
| 2.1 | Extraction size and entry limits, with clean-up on abort. **Done 2026-09-27 (`b415ffc`).** Owner decisions: defaults of 10 GiB of output and 100,000 entries per run; `--max-size` and `--max-entries` on `decompress` and `decrypt` (0 = no limit); the TUI uses the defaults; the limits also apply to encrypted folders; clean-up through a temporary folder (below). | SEC-007 |
| 2.2 | Extract via `os.Root`; mask archive modes. Also make extraction atomic (extract to a temporary sibling, then rename), which 2.11 left out because extracted files keep the archive's permissions. **Atomic extraction done 2026-09-27 (`b415ffc`, with 2.1):** `extractToDir` extracts into a new hidden 0700 folder and renames it into place, replacing an existing output with `--force` instead of merging into it; a single-file zip goes through a temporary file (0600). This also removes the planted-symlink risk. **`os.Root` and permissions done 2026-09-27 (`d751967`):** entries are written through an `os.Root` on the extraction folder; the owner chose owner-only permissions, so folders are 0700 and files 0600, or 0700 when marked executable. | SEC-008, BUG-014 |
| 2.3 | Build the directory-encryption tar.gz in memory (no temp plaintext). **Done 2026-09-27 (`d683739`):** `buildDirectoryArchive` replaces the temp file; the format is unchanged. Superseded by 3.2 (2026-09-27): the tar.gz now streams into the encrypted output instead of being held in memory. | SEC-006 |
| 2.4 | Enable SQLite `secure_delete`; align README wording. **Done 2026-09-27 (`d683739`):** `_secure_delete=on` on every connection; README updated. | SEC-009 |
| 2.4a | Optional: clear keys deleted by older versions with a one-time `VACUUM` when the database has free pages, run with `temp_store=MEMORY` so no copy lands in the temp folder. **Declined by the owner 2026-09-27.** | SEC-009 |
| 2.5 | Open the DB lazily (keys commands and TUI only); create it with mode 0600. **Done 2026-09-27 (`d683739`):** `NewRootCmdLazy` with a once-only `DatabaseOpener`; `prepareDatabaseFile` creates it 0600, and (owner decision) tightens an existing database to 0600. | SEC-010, BUG-005 |
| 2.6 | Add `--password-file` / `--password-stdin`; warn when `--password` is used; update README examples. **Done 2026-09-27 (`e512844`).** Owner decisions: `--password-file` only (piped input already covers stdin), and the warning always shows. | SEC-004 |
| 2.7 | Validate imported key metadata. **Done 2026-09-27 (`3d9384e`):** `validateKeyExport` checks version, key ID, algorithm and blob shape. | SEC-011 |
| 2.8 | Root-scoped opens when archiving. **Done 2026-09-27 (`3d9384e`):** `walkSourceTree` walks and opens through an `os.Root`; archives are byte-identical to before. | SEC-014 |
| 2.9 | Fix case-insensitive extension handling and the export filename reporting. **Done 2026-09-27 (`3d9384e`).** | BUG-007, BUG-006 |
| 2.10 | Gate TUI actions on `busy`. **Done 2026-09-27 (`3d9384e`).** | BUG-008 |
| 2.11 | Atomic writes: write to a temporary file in the destination folder, then rename, so a failure leaves no partial output. **Done 2026-09-26 (`b4cd66f`)** for single-file outputs (encrypt, decrypt, compress, gzip decompress, key export). Archive extraction moves to 2.2. | BUG-004 |
| 2.12 | Optional: an overwrite choice in the TUI forms, which refuse existing outputs today. **Declined by the owner 2026-09-27;** the TUI keeps refusing existing outputs. | BUG-004 |

## Phase 3 — Formats and architecture

| # | Change | Refs |
|---|---|---|
| 3.1 | Versioned file header, Argon2id (or PBKDF2 ≥ 600k), legacy read path; switch to stdlib `crypto/pbkdf2`. **Done 2026-09-27, with 3.2; committed in `0c57aef`.** Owner choices: Argon2id at 64 MiB, 3 passes, 4 lanes; header and streaming together; every artifact written in the new format; old formats readable with no time limit, no migration command. What was built: a 46-byte header (`format_v2.go`, layout in `maint.md` §3) on files, folders, stored keys and key exports; read limits on the header's Argon2id settings; legacy reads kept; `deriveKey` on `crypto/pbkdf2`. | SEC-005 |
| 3.2 | Streaming, chunked authenticated encryption for large files (depends on 3.1). **Done 2026-09-27, with 3.1; committed in `0c57aef`:** 64 KiB AES-256-GCM chunks in the STREAM construction; folders stream their tar.gz; output is kept only after the final chunk authenticates. A 1 GiB file peaks at 77 MiB instead of 3 GiB. | BUG-010 |
| 3.3 | Move the key generate/export/import flows into shared core functions used by both CLI and TUI. **Done 2026-10-04 (round 6, not yet committed):** `keys.go` (`GenerateStoredKey`, `StoredKeyCredential`, `ImportStoredKey`), with exports through `ExportKeyToFile`; the `Storage` interface is their parameter type. | `maint.md` §2 |
| 3.4 | Wire stored keys into encrypt/decrypt. **Approved 2026-09-26 (Q-002 = yes).** Needs 3.1's versioned header to record which key encrypted a file. **Done 2026-10-04 (round 6, not yet committed; Q-014):** key source 2 with KDF 3 (HKDF-SHA256 over the header's salt) and the key's ID in a 54-byte header; `encrypt --key`, a TUI field, and `decrypt` finding the key from the header. | Q-002, BUG-011 |
| 3.5 | Per-user default DB path plus migration, if Q-003 = yes. **Done 2026-10-04, committed in `a088f7f`, released in v1.3.0 (Q-003 = yes, notice-only migration):** `cryptare/cryptare.db` in the user data folder (`databasePath`, `userDataDir`), its folder created 0700 by the opener; a `cryptare.db` in the current folder gets a notice and is never opened; new `keys path` (SEC-010, SEC-016). | Q-003 |
| 3.6 | `main.version` variable and a `--version` flag. **Done** (2026-09-24, shipped in v1.0.1). It was moved ahead of 1.4 for W4. | BUG-009, W4 |

## Phase 4 — Tooling, tests, docs

| # | Change | Refs |
|---|---|---|
| 4.1 | Pin actions to SHAs; pin gosec and upload its SARIF; triage the findings. **In progress:** pins and SARIF upload shipped in v1.0.1, and the Security workflow succeeded. Remaining: confirm gosec alerts appear in Code Scanning, triage them, and decide on govulncheck. **2026-10-03:** the alerts appear in Code Scanning (10 open), and their dispositions are applied in code: G304 annotated `#nosec`, G301 parent folders now 0700. Remaining: push and confirm the alerts close, delete the stale CodeQL configurations left by the removed `codeql.yml`, and decide on govulncheck. **Done 2026-10-04:** all alerts closed after `4ba516a`; govulncheck runs in Security since `6a5fcb1`; SEC-012 closed. | SEC-012 |
| 4.2 | Non-root container user; pin images by digest **Done 2026-10-04, committed in `eb330a3`/`89a64e8`, released in v1.2.0 (owner approved):** non-root user 10001 owning `/app/data`, both images pinned by digest; validated by the Docker workflow on the next push (Docker isn't available in the analysis environment). | SEC-013 (needs approval) |
| 4.3 | Fill test gaps: `readPassword`, `ImportKeyFromFile`, CLI export/import round-trip, extraction traversal rejection, TUI `View`. **Partly done:** `readPassword` (1.1), `ImportKeyFromFile` and a CLI import/export of a legacy key (0.5, `TestLegacyShortPasswordCmds`), and traversal rejection (2.2, `TestExtractRejectsPathTraversal`). **Remaining:** a full CLI export/import round trip and the TUI `View`. | `maint.md` §5 |
| 4.4 | Add `SECURITY.md` with a private reporting channel. **Done 2026-10-04 (round 6, not yet committed):** the GitHub template it held is replaced: supported 1.3.x, private vulnerability reporting, acknowledgement within 7 days. | `CONTRIBUTING.md` |
| 4.5 | Settle the contents of `NOTICE`. **Done 2026-10-04, committed in `a088f7f`, released in v1.3.0 (Q-006):** `NOTICE` holds the project's attribution and points to `THIRD_PARTY_LICENSES.txt`, which `scripts/third-party-licenses.sh` writes in the release job; every archive holds it with `LICENSE` and `NOTICE`. | Q-006 |
| 4.6 | Rename the `tasks.db` fixture in `database_path_test.go` (the Munus names in the workflows are covered by W3). **Done 2026-10-04, committed in `a088f7f`, released in v1.3.0:** `keys.db`, with the Q-003 tests. | `notes.md` §3 |
| 4.10 | Settle the licensing of the C libraries in the release binaries: static glibc on Linux, MinGW-w64 runtime and libgcc on Windows (owner decision Q-013). **Done 2026-10-04, not yet committed (Q-013: musl; past releases left as they are):** `cd.yml` builds Linux in the `Dockerfile`'s pinned `golang:1.26.8-alpine` image, statically against musl; the Windows smoke test fails if the binary imports a MinGW-w64 toolchain DLL; `scripts/third-party-licenses.sh` adds the musl and MinGW-w64 runtime notices from `scripts/licenses/`. | Q-013 |
| 4.7 | Replace the README's release "known issue" callout with a note that v1.0.0 is broken and v1.0.1+ works, and re-verify the install steps. **Unblocked:** v1.0.1 works. Waiting for the owner to pull `3c80050` so the README edit doesn't conflict. **Done 2026-10-04, committed in `eb330a3`/`89a64e8`, released in v1.2.0:** the callout now says only v1.0.0 is broken; the install steps add the download commands and were run against v1.1.0. | BUG-001 |
| 4.8 | Doc follow-ups from the 2026-09-24 review. The owner has fixed the `AGENTS.md` typo and `golang.md`'s `gofmt -s`, and `map.md` already covers `intel/` as a directory. **Remaining (owner):** add `intel/golang.md` to the "Repository Intelligence Documents" list in `AGENTS.md`. | `AGENTS.md`, `intel/` |
| 4.9 | Post-merge doc updates. **Done 2026-09-24:** the CI/CD table in `map.md`, `maint.md` §6, and the CI description in `CONTRIBUTING.md`. **Remaining:** the README release table (W6), after the owner pulls `3c80050`. **Done 2026-10-04, committed in `eb330a3`/`89a64e8`, released in v1.2.0:** the Windows ARM64 row is gone (W6). | W3, W6 |

## Phase 5 — Remediation from the 2026-10-03 analysis

Covers SEC-015 to SEC-019 ([`cybersec.md`](cybersec.md)), BUG-015 to BUG-024
([`notes.md`](notes.md)), the gaps recorded in SEC-003, SEC-005, SEC-010, SEC-011 and
SEC-013, and questions Q-008 to Q-012. The [Principles](#principles) apply: one issue
per PR, a regression test that fails before the fix, formats and flags kept compatible,
and owner approval for CI, `go.mod`, `Dockerfile` and dependency changes.

**Order.** Tier A first: it covers data loss, a broken feature, possibly vulnerable
release binaries and plaintext left behind on interruption. Cut a patch release once
Tier A is merged. Then Tier B (hardening) and Tier C (minor bugs). Tier D (tooling and
docs) can run alongside.

### Tier A — fix first

| # | Change | Refs | Acceptance | Approval |
|---|---|---|---|---|
| 5.1 | **Toolchain and vulnerability scanning.** Confirm first: `go version -m` on the v1.1.0 release binaries; `govulncheck ./...` and `govulncheck -mode=binary`. Then: `toolchain go1.26.<latest>` in `go.mod` (or `go-version: '1.26.x'` with `check-latest: true`); a govulncheck job in `security.yml` that fails on reachable findings; the same Go version for the Docker builder; Dependabot for `gomod` and `github-actions`. Optional: `-trimpath` and release attestations. **Done 2026-10-03, committed in `6a5fcb1`/`b3278ea`, CI green (owner approved):** v1.1.0's linux/amd64 binary confirmed built with go1.26.0; `toolchain go1.26.8` in `go.mod`; `golang:1.26.8-alpine` builder; a govulncheck v1.8.0 job in `security.yml`; `.github/dependabot.yml` for `gomod` and `github-actions`. Not done: `govulncheck -mode=binary` (database unreachable here), the digest pin (with 4.2), and the optional `-trimpath` and attestations (SEC-018). | SEC-018, SEC-012 step 5 | New release binaries report the latest 1.26.x; govulncheck passes in CI; CI, CD and Docker use one Go version | Yes: `go.mod`, CI, Dockerfile |
| 5.2 | **`keys export` output safety.** Call `CheckOutputPath` in the CLI (overwrite only with a new `--force`) and in the TUI (always refuse, like the other forms). Refuse the key database file itself, even with `--force` (`os.SameFile`). **Done 2026-10-03, committed in `6a5fcb1`/`b3278ea`, CI green (behaviour change confirmed by the owner):** `checkExportOutput` in the CLI (`--force`) and the TUI (never overwrites); the key database and its SQLite files are always refused (`ErrOutputIsKeyDatabase`). | BUG-017 | CLI and TUI tests: an existing file is refused without `--force` and left intact; the database path is always refused; the default name still works | Behaviour change: confirm |
| 5.3 | **Folder-encryption output containment.** One default-output helper for the core, CLI and TUI that writes next to the folder (`dir/` → `dir.enc`; `.` → `<parent>/<name>.enc`). Refuse an output inside the folder with `ErrOutputInsideInput`, as `compress` does. **Done 2026-10-03, committed in `6a5fcb1`/`b3278ea`, CI green (confirmed by the owner):** `defaultEncryptOutput` and `checkOutputOutsideFolder`, used by the core, the CLI and the TUI. | BUG-016, BUG-013 | Core, CLI and TUI tests for `dir/`, `.` and an explicit output inside the folder; existing round trips pass | Behaviour change: confirm |
| 5.4 | **Latin-1-safe gzip names.** Store the name in the gzip header only when Latin-1 can hold it. Otherwise use a fallback: an ASCII name ending in `.tar` for a folder (so `isTarGzArchive` still recognises it), or no name for a file. No format change, because restoring a folder doesn't use the name. **Done 2026-10-03, committed in `6a5fcb1`/`b3278ea`, CI green:** `gzipHeaderName` / `gzipFolderName` (fallback `archive.tar`). | BUG-015 | Round-trip tests with Chinese, Cyrillic and emoji names for gzip, tar.gz and folder encryption; artifacts from older builds still read | No |
| 5.5 | **Cancellation and clean-up.** Pass a `context.Context` through the core file operations, checked between chunks and entries (add `…Context` variants and keep today's functions as wrappers). Run the CLI file commands under `signal.NotifyContext` (SIGINT, SIGTERM, SIGHUP) and exit with 128 + the signal after clean-up. In the TUI, quitting cancels a running action and exits once it reports back. Extend the password prompt's handler to SIGTERM, SIGQUIT and SIGHUP. Update the README and `maint.md` §4. **Done 2026-10-03, committed in `6a5fcb1`/`b3278ea`, CI green:** as described, with a small signal handler of its own (`runCancellable`) instead of `signal.NotifyContext`, so the exit status can name the signal. See SEC-015. | SEC-015, BUG-018 | The tests listed under SEC-015's validation; a pseudo-terminal check that echo comes back after each signal | No (internal API only) |

### Tier B — hardening

| # | Change | Refs | Acceptance | Approval |
|---|---|---|---|---|
| 5.6 | **Database trust.** On Unix, refuse (or warn about, per Q-010) a database or side file that the user doesn't own or that group or others can write. Escape control characters in every stored field that `keys list` and the TUI show. Bring 3.5 (Q-003) forward. **2026-10-03, committed in `4ba516a`, CI green:** the refusal (Q-010: refuse) and the escaping are done and validated (SEC-016); 3.5 stays separate, as the owner chose not to move the default path in this change. | SEC-016, SEC-011 | The tests listed under SEC-016's validation | Q-010 (answered: refuse) |
| 5.7 | **Quiet database layer.** Silence GORM's logger (`logger.Silent`, or parameterised and on stderr only when enabled). Turn "record not found" and UNIQUE-constraint errors into clear messages. **Done 2026-10-04, committed in `eb330a3`/`89a64e8`, released in v1.2.0:** silent GORM logger with error translation; `ErrKeyNotFound`/`ErrKeyExists` explained with the key ID (SEC-017). | SEC-017 | Failed key commands print nothing on stdout; errors name the key ID | No |
| 5.8 | **Database path handling.** Build the DSN so the driver opens exactly the file `prepareDatabaseFile` prepared (for example a `file:` URI with the path escaped, plus `_secure_delete=on`), or refuse paths containing `?`. Prepare the file behind a `file:` URI as well. **Done 2026-10-04, committed in `eb330a3`/`89a64e8`, released in v1.2.0:** `databaseFilePath`; `?` in a plain path refused; `file:` URIs prepared and trust-checked (SEC-010). | SEC-010 gap | Tests with `?` in a folder name and with a `file:` URI: the database holding the keys is 0600 and no stray file is left | No |
| 5.9 | **Repository and container hygiene.** Add `*.ckey`, `*.db-journal`, `*.db-wal` and `*.db-shm` to `.gitignore`. Add a `.dockerignore`. Drop `sqlite-libs` and `ca-certificates` from the runtime image, and stamp its version. Do it together with 4.2 (SEC-013: non-root user, digest pins). **Done 2026-10-04, committed in `eb330a3`/`89a64e8`, released in v1.2.0 (owner approved):** `.gitignore` patterns, `.dockerignore`, runtime packages dropped, version stamped (`ARG VERSION`), `-trimpath`; with 4.2 (SEC-003, SEC-013). | SEC-003, SEC-013 | `git check-ignore` matches the new patterns; the Docker smoke test passes; `docker run --entrypoint id <image> -u` isn't 0 | Yes: Dockerfile |
| 5.10 | **Windows permissions.** State in the README and `maint.md` §4 that on Windows the output inherits the permissions of the folder it is written to. Optionally set owner-only ACLs (Q-012). **Done 2026-10-04, committed in `eb330a3`/`89a64e8`, released in v1.2.0 (Q-012: document only):** README and `maint.md` §4 (SEC-019). | SEC-019 | README review; with ACLs, a Windows CI test reads an output's DACL | Q-012; `go.mod` if ACLs |

### Tier C — lower-severity bugs

| # | Change | Refs | Acceptance |
|---|---|---|---|
| 5.11 | **Containment by file identity.** In `checkOutputOutsideDir` and `checkInputOutsideOutput`, walk up from the output's or input's parent comparing folders with `os.SameFile`, instead of comparing spellings. Reuse it in 5.3. **Done 2026-10-04, committed in `eb330a3`/`89a64e8`, released in v1.2.0:** `pathWithin` also compares folders by identity (`os.SameFile`) while walking up from the path. | BUG-019 | Tests with a symlinked spelling; a letter-case test on the macOS and Windows runners |
| 5.12 | **Validate compression options.** Refuse an explicit `--format` that contradicts the output's extension, and a `--level` outside 1–9 (−1 stays the default), in the CLI and TUI. Behaviour change for scripts: owner to confirm, README note. **Done 2026-10-04, committed in `eb330a3`/`89a64e8`, released in v1.2.0 (behaviour change confirmed):** `resolveCompressFormat` refuses contradictions (`ErrFormatMismatch`); `checkCompressLevel` in the CLI and TUI (`ErrInvalidLevel`). | BUG-020 | CLI and TUI tests for each case |
| 5.13 | **Shorter temporary names.** Cap the part of the output name copied into temporary file and folder names (for example 64 bytes, cut at a UTF-8 boundary). **Done 2026-10-04, committed in `eb330a3`/`89a64e8`, released in v1.2.0:** `tempNamePart` (64 bytes at a UTF-8 boundary) for temporary files and folders. | BUG-021 | Encrypt, decrypt, compress and extract all succeed with a 251-byte name |
| 5.14 | **Explicit empty password.** Treat `--password` as given when `Flags().Changed("password")`. **Done 2026-10-04, committed in `eb330a3`/`89a64e8`, released in v1.2.0:** `passwordFlags.get` uses `Flags().Changed("password")`. | BUG-023 | CLI test: `--password ""` decrypts a legacy empty-password file without reading stdin |
| 5.15 | **Size caps for whole-file reads.** Refuse a `.ckey` above a small limit (for example 1 MiB). Refuse a legacy-format input above a cap the owner chooses, before reading it, with a clear error. **Done 2026-10-04, committed in `eb330a3`/`89a64e8`, released in v1.2.0:** key exports capped at 1 MiB (`ErrInputTooLarge`); legacy-format inputs capped by `--max-size` (owner decision), checked before reading. | BUG-024 | Oversized inputs fail fast with a clear message |
| 5.16 | **Single `.tar` files.** Per Q-011: a gunzip-only option, recording the input type, or documentation. **Done 2026-10-04, committed in `eb330a3`/`89a64e8`, released in v1.2.0 (Q-011: `--raw`):** `GunzipFileContext`, `decompress --raw` (CLI only; the TUI always extracts). | BUG-022, Q-011 | `compress x.tar` followed by `decompress` gives `x.tar` back, or the documented route does |
| 5.17 | **Export password clarity.** Name the prompt and flag help for what they set (the password that protects the export file), and check the key's master password before exporting. Do it with 3.3. **Done 2026-10-04 (round 6, not yet committed; Q-015):** the export is protected by the key's master password, asked for once by name and checked; import checks the key inside. | BUG-011 | CLI and TUI tests: a wrong master password is refused; the prompts name the right password |

### Tier D — tooling and documentation

| # | Change | Refs |
|---|---|---|
| 5.18 | Run what the 2026-10-03 analysis couldn't: golangci-lint v2.13.2, gosec v2.29.0 and govulncheck. Triage gosec in Code Scanning (4.1). **Done 2026-10-04:** CI runs golangci-lint v2.13.2 and Security runs gosec v2.29.0 and govulncheck v1.8.0 on every push, all passing on `b3278ea`; gosec and actionlint were also run locally on 2026-10-03. | SEC-012, SEC-018 |
| 5.19 | Run `go test -race` in CI on Linux and macOS. Needs owner approval (CI change). **Done 2026-10-04, committed in `eb330a3`/`89a64e8`, released in v1.2.0 (owner approved):** `-race` on the Linux and macOS jobs of `ci.yml`. | `maint.md` §7 |
| 5.20 | Docs: the README claims that don't hold yet (`notes.md` §3) and `CONTRIBUTING.md`'s CGO statement. Until 5.2–5.5 ship, add README notes on Unicode folder names, `keys export` overwriting and interrupted runs. **Done 2026-10-04, committed in `eb330a3`/`89a64e8`, released in v1.2.0:** the README claims hold since Tier A, Windows is documented (5.10), and `CONTRIBUTING.md`'s CGO statement is corrected. | `notes.md` §3 |
| 5.21 | Fuzz the code that reads untrusted input with Go's built-in fuzzing: `parseV2Header` and `decryptingReader`, `validateKeyExport`, `parseSize`, `extractTarGz` and `extractZipEntries`. **Done 2026-10-04, committed in `eb330a3`/`89a64e8`, released in v1.2.0:** `internal/fuzz_test.go` with six targets (header, decrypting reader, key export, `parseSize`, tar and zip extraction); CI runs their seeds. Each ran 25–35 s of coverage-guided fuzzing without a failure, except `FuzzExtractTar`, whose worker stalls in this analysis environment inside Go runtime code; 80,000 uninstrumented fuzz inputs and a 200,000-input mutation stress of tar extraction found no hang, slow input or write outside the output folder. Run it on a normal machine to confirm. | `maint.md` §5 |

### Decisions and approvals Phase 5 needs

- **Approvals:** 5.1 (approved 2026-10-03), 5.9 with 4.2 and 5.19 (approved 2026-10-04);
  5.10's ACLs weren't wanted (Q-012).
- **Behaviour changes to confirm:** 5.2 and 5.3 were confirmed on 2026-10-03, and 5.12
  on 2026-10-04.
- **Questions:** all answered. Q-010 (2026-10-03: refuse), Q-011 (2026-10-04: `--raw`),
  Q-012 (2026-10-04: document only), and on 2026-10-04 Q-003 (per-user store, notice-only
  migration), Q-008 (keep the limits) and Q-009 (NFKC, byte order mark, `x/text`).
- **Release:** v1.2.0 shipped Tiers A–D (except 5.17) on 2026-10-04, with SEC-002's
  migration notes.

## Phase 6 — Production-readiness review (2026-10-04)

From the 2026-10-04 review of v1.3.1. Approvals and decisions are under "Owner
decisions (2026-10-04, round 6)" above. All done in round 6 and not yet committed,
except where the status says otherwise.

| # | Change | Refs | Status |
|---|---|---|---|
| 6.1 | `SECURITY.md`: replace GitHub's template (it listed versions 5.1.x and 4.0.x and no way to report) | 4.4 | Done |
| 6.2 | Usable stored keys and one password per key | 3.3, 3.4, 5.17, BUG-011, Q-014, Q-015 | Done |
| 6.3 | Golden format fixtures written by v1.0.1 and v1.3.1 (and stored-key files by this build), opened by `golden_test.go`; a round-trip test can't catch a format change made in the writer and the reader alike | `maint.md` §3, SEC-003 note | Done |
| 6.4 | README install: it pinned `VERSION=v1.1.0`, which can't read version 2 files and was built with Go 1.26.0; now `releases/latest/download` | SEC-018 | Done |
| 6.5 | Usage text only for command-line mistakes (`silenceUsageOnRun`) | — | Done |
| 6.6 | Build provenance attestations on release files, and verification in the README | SEC-020 | Done; validated by the next tag |
| 6.7 | Dependabot for the `Dockerfile`'s digest-pinned images (golang patch updates only) | SEC-018, SEC-013 | Done |
| 6.8 | `docker.yml` checks that the image runs as UID 10001 | SEC-013 | Done; closes SEC-013 when it passes in CI |
| 6.9 | Close SEC-018 with `go version -m` on the v1.3.1 binaries | SEC-018 | Done (closed) |
| 6.10 | Document why `FuzzExtractTar` seems to stall (minimisation) and the flag that avoids it | 5.21 | Done (`notes.md` §3, `maint.md` §5) |
| 6.11 | Bring `notes.md` §1 and this plan up to date with v1.3.1 | — | Done |
| 6.12 | Run darwin/amd64 once on an Intel Mac or under Rosetta | W7 | Open (owner) |
| 6.13 | Optional: macOS notarisation and Windows Authenticode signing, depending on audience | SEC-020 step 2 | Open (owner decision) |
| 6.14 | Publish the Docker image to GitHub Packages on each release (owner request, 2026-10-04): `cd.yml`'s `container` job, after the release; `ghcr.io/<owner>/cryptare` tagged `X.Y.Z`, plus `X.Y` and `latest` for the newest release; smoke test before pushing; attestation on the digest. After the first tagged run, the owner makes the package public in its settings (GitHub creates it private) | SEC-020, SEC-013 | Done; validated by the next tag |
