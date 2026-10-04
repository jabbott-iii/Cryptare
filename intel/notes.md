# Engineering Notes

Durable engineering notes and unresolved technical questions. Security issues are
tracked in [`cybersec.md`](cybersec.md), work sequencing in [`plan.md`](plan.md), and
the record of past work, including resolved defects and answered questions, in
[`history.md`](history.md).

Last updated: 2026-10-04 (round 4)

## 1. Current snapshot (2026-10-04)

- **Branch state.** `main` is at `89a64e8`, level with `origin/main`, and tagged
  v1.2.0, which CD #4 released on 2026-10-04 with the drafted notes; CI #142, Docker #23
  and Security #148 passed on it. The working tree holds this round's uncommitted
  changes (Q-003, Q-005's CI guard, Q-006 and Q-009). Dependabot's PR #26
  (`golang.org/x/crypto` 0.56.0 → 0.57.0, CI green) is open; it edits the same `go.mod`
  require block as Q-009's `golang.org/x/text` line, so whichever lands second may need
  a trivial rebase.
- **Toolchain.** `go.mod` declares `go 1.26.0` and `toolchain go1.26.8` (`6a5fcb1`), so
  CI, CD and the Docker builder use Go 1.26.8, and an older local `go` downloads it
  (SEC-018). Releases up to v1.1.0 were built with Go 1.26.0; v1.2.0 is the first built
  with 1.26.8.
- **Validation** of the 2026-10-04 round on the owner's machine (linux/amd64, non-root,
  Go 1.26.8):

  | Check | Result |
  |---|---|
  | `gofmt -s -l .` | no files listed |
  | `go vet ./...` | clean, also for windows, darwin and freebsd (CGO off), with test builds |
  | `go test -race -count=1 ./...` | pass |
  | `go mod tidy` | can't run here (some module hosts are unreachable); `go mod tidy -e` leaves `go.mod` as edited. CI checks it |
  | gosec v2.29.0 | 0 issues, 9 `#nosec` |
  | actionlint 1.7.12 | nothing reported for the four workflows |
  | golangci-lint, govulncheck | not run here; CI and Security run them on every push |
  | Real-binary checks | the v1.2.0 source against this round: Q-003 (SEC-010) and Q-009 compatibility, recorded in `history.md` |

- **Not verified:** Windows and macOS behaviour beyond CI, the Docker image's UID
  (SEC-013), the Go version inside the v1.2.0 binaries (SEC-018), and the release job's
  new packaging steps, which first run on the next tag.
- **Lowest coverage (2026-10-03):** `View` and `actionTitle` (0%), `readTerminalPassword` (20%),
  `readPassword` (46%), `Update` (54%), `replacePath` (56%), `writeZipFile` (59%).

## 2. Open defects (non-security)

| ID | Defect | Evidence | Location |
|---|---|---|---|
| BUG-011 | `keys export` doesn't check that the export password matches the key's master password, and `keys import` doesn't check that the inner blob decrypts. A `.ckey` can therefore need two different passwords to be usable. The CLI prompt for the export password reads "Enter master password:" and the flag help says "master password for export encryption", although the value only protects the export file, so users can't tell the two passwords apart. | Code review | `crypto.go` `ExportKeyToFile`, `ImportKeyFromFile`; `logic-cli.go` `newKeysExportCmd` |

### Resolved defects

Kept as an index because other documents refer to these IDs. The evidence and
validation for each fix are in `history.md`.

| ID | Defect | Fixed in |
|---|---|---|
| BUG-001 | Release binaries built without CGO didn't run | v1.0.1 (`5286920`) |
| BUG-002 | The TUI panicked on unhandled keys in forms | `18ea97a` |
| BUG-003 | `compress` onto its own input destroyed the input | `c47a94f` (v1.1.0) |
| BUG-004 | Existing outputs were overwritten silently, and writes weren't atomic | `c47a94f`, `b4cd66f`, `b415ffc` |
| BUG-005 | Every command opened, and created, the key database | `d683739` |
| BUG-006 | `keys export` reported a different file name from the one it wrote | `3d9384e` |
| BUG-007 | Archive and `.enc` extensions were matched case-sensitively | `3d9384e` |
| BUG-008 | A second TUI action could start while one was running | `3d9384e` |
| BUG-009 | No `--version`, and `-X main.version` had no effect | v1.0.1 (`5286920`) |
| BUG-010 | Encryption and decryption held whole files in memory | `0c57aef` (legacy reads: BUG-024) |
| BUG-012 | Ctrl+C at the hidden prompt left echo off | `ed46150` (other signals: BUG-018) |
| BUG-013 | A folder could be compressed into an archive inside itself | `c47a94f` (gaps: BUG-016, BUG-019) |
| BUG-014 | A read-only folder in an archive blocked extraction for non-root users | `d751967` |
| BUG-015 | Names outside Latin-1 broke gzip compression and folder encryption | `6a5fcb1` (tests in `b3278ea`) |
| BUG-016 | Encrypting a folder could put the output inside the folder | `6a5fcb1` (tests in `b3278ea`) |
| BUG-017 | `keys export` overwrote any file, including the key database | `6a5fcb1` (tests in `b3278ea`) |
| BUG-018 | The hidden prompt left echo off after SIGTERM, SIGQUIT or SIGHUP | `6a5fcb1` (tests in `b3278ea`) |
| BUG-019 | Containment checks compared paths as text (symlinks, letter case) | `eb330a3` (v1.2.0) |
| BUG-020 | Contradictory `--format` and out-of-range `--level` were silently overridden | `eb330a3` (v1.2.0) |
| BUG-021 | Names near the 255-byte limit failed because of long temporary names | `eb330a3` (v1.2.0) |
| BUG-022 | A single `.tar` file didn't round-trip through compress and decompress | `eb330a3` (v1.2.0; `decompress --raw`) |
| BUG-023 | `--password ""` fell back to the prompt | `eb330a3` (v1.2.0) |
| BUG-024 | Legacy decrypt and key import read whole inputs without a cap | `eb330a3` (v1.2.0) |

## 3. Design observations

- **Stored keys aren't used for file encryption.** `encrypt` and `decrypt` derive keys
  from passwords only, and no code path consumes `key_models`. The owner wants stored
  keys usable (plan 3.4).
- **CLI and TUI duplicate flows.** Key generate, export and import are implemented
  twice (`logic-cli.go` and `logic-tui.go` `buildActionCmd`), so a change to one must be
  mirrored in the other until plan 3.3 shares them. Checks that protect data, such as
  the password policy, live in core for this reason; only the confirmation step is
  implemented in each interface.
- **Dead code.** The `Storage` interface is declared but unused.
- **Stale comment.** `ImportKeyFromFile` says "Parse minimal JSON manually to avoid
  import cycle", but it uses `encoding/json`.
- **Leftover from another project.** `database_path_test.go` used a `tasks.db` fixture
  (cosmetic; plan 4.6). Renamed to `keys.db` on 2026-10-04 with the Q-003 tests (not yet
  committed).
- **IDE files.** `.idea/` is tracked (`.gitignore` has `# .idea/` commented out).
  `.junie/plans/` is an empty, untracked agent workspace.
- **NOTICE** ended at "This product includes third-party software:" with an empty list.
  Since Q-006 (2026-10-04, not yet committed) it holds the project's attribution and
  points to `THIRD_PARTY_LICENSES.txt`, which the release job generates with
  `scripts/third-party-licenses.sh` and puts in every archive.
- **C libraries in the release binaries (Q-013).** The Linux binaries are linked
  statically (`-linkmode external -extldflags -static` on an Ubuntu runner), so they
  contain parts of the runner's GNU C Library (glibc, LGPL-2.1-or-later), and the
  Windows build may include parts of the MinGW-w64 runtime and libgcc.
  `THIRD_PARTY_LICENSES.txt` covers the Go code and SQLite only.
- **TUI password field.** It masks input with one `*` per character, which reveals the
  password's length.
- **Documentation drift** (all fixed in `eb330a3`, released in v1.2.0).
  - The README said every published release was broken ("Known issue", BUG-001),
    although v1.0.1 and later work (plan 4.7), and its release table listed
    `cryptare_windows_arm64.zip`, which isn't built (W6).
  - `CONTRIBUTING.md` said CGO-less builds "fail on every command at runtime"; only the
    `keys` commands and the TUI fail (`maint.md` §6).
  - Some README guarantees didn't hold: "everything the tool writes is private to you"
    now carries the Windows caveat (SEC-019). Missing parent folders were created 0755
    until the 2026-10-03 G301 fix (SEC-012). "Commands don't overwrite anything by
    default" (BUG-017) and "a command that fails part-way leaves no partial output"
    (SEC-015) hold since `6a5fcb1`.
- **CI runs `go test` with `-race` on Linux and macOS** since plan 5.19 (committed in
  `eb330a3`); Windows runs it without, as the race detector needs CGO there too.
- **Dev container.** The `docker-outside-of-docker` feature hands the container the
  host's Docker socket, so anything run inside it (tests, tools, dependencies) can
  control the host's Docker daemon. A convenience trade-off worth knowing about before
  running untrusted code there.
- **Release integrity.** Releases publish `checksums.txt` beside the archives, with no
  signature or provenance attestation, so the checksums only catch accidental
  corruption (SEC-018 step 5).

## 4. Open questions (owner decisions)

| ID | Question | Why it matters |
|---|---|---|
| Q-013 | How should the C libraries linked into the release binaries be licensed? The Linux binaries statically include glibc (LGPL-2.1-or-later). Options: build them against musl instead (for example in an Alpine container, as the `Dockerfile` does; MIT, which only needs its notice added), link glibc dynamically (giving up the portable static binary), or keep the static glibc build and meet the LGPL's conditions for it (offering glibc's source and a way to relink). The Windows build's MinGW-w64 runtime and libgcc parts need checking too. | Licensing of published binaries (found while doing Q-006). A licensing decision for the owner; this note isn't legal advice. |

Answered or closed: Q-001 (release builds use native CGO per OS), Q-002 (stored keys
should be usable; plan 3.4), Q-004 (the default password policy, `maint.md` §4) and
Q-007 (v1.0.0 is left as it is), Q-010 (an untrusted key database is refused, not
warned about; SEC-016), Q-011 (a gunzip-only `decompress --raw`; BUG-022) and Q-012
(Windows permissions are documented, not enforced with ACLs; SEC-019). On 2026-10-04:
Q-003 (the default key store moves to a per-user data folder, with a notice-only
migration; plan 3.5), Q-005 (no history purge; the Copilot branch is gone and CI
refuses tracked key material; SEC-003), Q-006 (`NOTICE` keeps only the project's
attribution, and release archives carry a generated `THIRD_PARTY_LICENSES.txt`), Q-008
(the Argon2id read limits stay at 1 GiB, 10 passes and 16 lanes) and Q-009 (passwords
are NFKC-normalised with a leading byte order mark dropped, `golang.org/x/text` is a
direct dependency, and a new KDF identifier marks data whose password normalisation
changed). Their discussion is in `history.md`.

## 5. Reproducing the validation locally

From the repository root, on a machine with Go 1.26 and a C compiler:

```bash
gofmt -s -l .
go mod tidy && git diff --exit-code go.mod go.sum
go vet ./...
golangci-lint run          # v2.13.2
go test -race -count=1 ./...
gosec ./...                # v2.29.0
govulncheck ./...          # checks the toolchain that go.mod selects
```

Run probes that can use a lot of memory, such as decompression or allocation bombs,
inside a memory-capped cgroup, for example
`systemd-run --user --scope -p MemoryMax=768M -p MemorySwapMax=0 <command>`. An
address-space limit (`ulimit -v`) is no substitute: a low one breaks Go binaries (cgo
thread creation fails) before memory runs out, and on 2026-10-03 a probe limited only
by `ulimit -v` exhausted the whole machine's memory.
