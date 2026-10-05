# Engineering Notes

Durable engineering notes and unresolved technical questions. Security issues are
tracked in [`cybersec.md`](cybersec.md), work sequencing in [`plan.md`](plan.md), and
the record of past work, including resolved defects and answered questions, in
[`history.md`](history.md).

Last updated: 2026-10-04 (round 6)

## 1. Current snapshot (2026-10-04, round 6)

- **Branch state.** `main` is at `6773ae3`, level with `origin/main` and tagged v1.3.1,
  which CD released on 2026-10-04 with round 5 (Q-013's musl Linux builds and PR #26's
  `golang.org/x/crypto` 0.57.0). The working tree holds this round's uncommitted
  changes: usable stored keys (plans 3.3, 3.4, 5.17; BUG-011), golden format fixtures,
  `SECURITY.md`, the README install fix, usage text only for command-line mistakes,
  release attestations, Dependabot for Docker images, the image UID check, and the
  Docker image published to GitHub Packages by CD (`ghcr.io/jabbott-iii/cryptare`).
- **Toolchain.** `go.mod` declares `go 1.26.0` and `toolchain go1.26.8` (`6a5fcb1`), so
  CI, CD and the Docker builder use Go 1.26.8. Every v1.3.1 release binary reports
  go1.26.8 (`go version -m`, SEC-018, closed).
- **Validation** of round 6 (a scratch copy, linux/amd64, Go 1.26.8 built from source,
  `x/crypto` 0.57.0, `x/text` 0.42.0, `x/sys` 0.48.0; the module proxy was unreachable,
  so those three and the two GORM modules came from their GitHub mirrors at the same
  tags, and every other module was checked against `go.sum`):

  | Check | Result |
  |---|---|
  | `gofmt -s -l .` | no files listed |
  | `go vet ./...` | clean |
  | `go test -race -count=1 ./...` | pass; coverage 89.2% (`main`), 80.9% (`internal`) |
  | golangci-lint v2.13.2 | 0 issues |
  | gosec v2.29.0 | 0 issues, 10 `#nosec` (one new G304, SEC-012) |
  | actionlint 1.7.12 | nothing reported for the four workflows, with shellcheck 0.11.0 checking their scripts |
  | govulncheck | not run here (database unreachable); Security runs it on every push |
  | Golden fixtures | open; a mutated nonce layout fails all four golden tests while the round-trip tests still pass |
  | Real binaries | v1.3.1's `cryptare_linux_amd64` and the new build read each other's password files; v1.3.1 refuses a stored-key file ("key source 2") |

- **Not verified:** the new workflow steps (the attestation steps and the GitHub
  Packages push need a tag; the Docker UID check and the Windows runs need CI; Docker
  isn't available here, so the container job's build, smoke test and push haven't run;
  its tag logic was run against simulated release lists), darwin/amd64 on real hardware (W7),
  and the Windows DLL check's log (v1.3.1 was published by CD, which runs it before
  packaging).
- **Lowest coverage:** the TUI's `View` and `handleVimFormKey`, `readTerminalPassword`
  and `readPassword` (terminal-only paths), `replacePath` and `writeZipFile`.

## 2. Open defects (non-security)

None open.

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
| BUG-011 | `keys export` didn't check the key's master password and `keys import` didn't check the key inside, so an export could need two passwords; the prompts didn't say which password was meant | round 6, not yet committed (one password per key, Q-015) |
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

- **Stored keys encrypt files** (plan 3.4, round 6, not yet committed). `encrypt --key`
  and the TUI's "Stored key ID" field write a key source 2 header naming the key
  (Q-014); `decrypt` reads the header of a regular file (`EncryptedWithStoredKey`) and
  asks for that key's master password. A stored-key file read through a pipe isn't
  looked up, because peeking would consume the stream; it fails with
  `ErrStoredKeyRequired` and a hint. Losing a key, or the key store, without an export
  makes the files encrypted with it unrecoverable; the README says to keep exports.
- **Shared key flows** (plan 3.3). `keys.go` holds `GenerateStoredKey`,
  `StoredKeyCredential` and `ImportStoredKey`, used by both interfaces, and exporting
  goes through `ExportKeyToFile`, which checks the master password. What remains in
  each interface is prompting, the output-path check and the TUI's field handling.
- **The `Storage` interface** is now the parameter type of the shared key flows.
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
- **C libraries in the release binaries (Q-013).** Up to v1.3.0 the Linux binaries were
  linked statically against the Ubuntu runner's GNU C Library (glibc,
  LGPL-2.1-or-later). Since round 5 (not yet committed) they are linked statically
  against musl (MIT) in the `Dockerfile`'s pinned Alpine builder, and
  `THIRD_PARTY_LICENSES.txt` adds musl's notice for Linux and the MinGW-w64 runtime's
  for Windows (whose runtime code Go's linker always pulls in with `-lmingwex
  -lmingw32`). libgcc's licence exception needs no notice. The owner left the
  published v1.0.1–v1.3.0 releases as they are.
- **Password normalisation depends on `golang.org/x/text`.** v0.42.0 fixed NFC/NFKC
  composition bugs in v0.41.0: an accent could compose with a letter across an Indic
  vowel sign or length mark that should block it (Tamil, Malayalam, Bengali, Oriya,
  Kannada, Sinhala, Myanmar, Balinese, Grantha and similar). In 300,000 random
  mixed-script strings, 104 normalised differently, and v0.42.0 matched Python's
  `unicodedata` in every case. A v1.3.0 file protected with such a password (KDF 2)
  doesn't open in builds with v0.42.0; it opens with v1.3.0, which can re-encrypt it.
  `TestNormalizePasswordKnownAnswers` now pins the normalised forms, so any later
  change, including the Unicode 17 tables that `x/text` selects from Go 1.27 on, fails
  CI before a release.
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
- **Release integrity.** Releases up to v1.3.1 publish only `checksums.txt` beside the
  archives, which catches accidental corruption but not a substituted file. From the
  next release, CD attaches signed build provenance to every published file (SEC-020).
- **Fuzzing `FuzzExtractTar`** isn't hung when it seems to stall: the fuzzer spends up
  to `-fuzzminimizetime` (60 s by default) minimising each new input, and each run of
  this target creates folders, so the exec count stops moving meanwhile. With
  `-fuzzminimizetime 3s` it ran 60 s, about 24,800 inputs and 25 new ones, with no
  failure (2026-10-04, round 6).
- **Usage text.** The CLI printed its whole usage after every error, a wrong password
  included. `silenceUsageOnRun` now keeps it for command-line mistakes, which Cobra
  reports before a command runs.

## 4. Open questions (owner decisions)

None open.

Answered in round 6 (2026-10-04): Q-014 (a file encrypted with a stored key records
the key's ID in its header: key source 2, KDF 3 for HKDF-SHA256 from the key, a 54-byte
header; v1.3.1 and earlier refuse such files) and Q-015 (a key has one password: an
export is protected by the key's own master password, which is checked, and import
checks the key inside, asking for its master password when an older export has a
password of its own). The owner also chose to build plans 3.3 and 3.4 now, approved the
attestation, Dependabot Docker and Docker UID changes, and set the security policy:
acknowledgement within 7 days, latest minor line (1.3.x) supported.

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
changed), and Q-013 (Linux release binaries are built against musl; C library notices
ship in `THIRD_PARTY_LICENSES.txt`; past releases are left as they are). Their
discussion is in `history.md`.

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
go test -run '^$' -fuzz '^FuzzExtractTar$' -fuzztime 5m -fuzzminimizetime 3s ./internal
```

Run probes that can use a lot of memory, such as decompression or allocation bombs,
inside a memory-capped cgroup, for example
`systemd-run --user --scope -p MemoryMax=768M -p MemorySwapMax=0 <command>`. An
address-space limit (`ulimit -v`) is no substitute: a low one breaks Go binaries (cgo
thread creation fails) before memory runs out, and on 2026-10-03 a probe limited only
by `ulimit -v` exhausted the whole machine's memory.
