# Engineering Notes

Durable engineering notes and unresolved technical questions. Security issues are
tracked in [`cybersec.md`](cybersec.md), work sequencing in [`plan.md`](plan.md), and
the record of past work, including resolved defects and answered questions, in
[`history.md`](history.md).

Last updated: 2026-10-04

## 1. Current snapshot (2026-10-03; toolchain updated 2026-10-04)

- **Branch state.** `main` is at `b3278ea`, level with `origin/main` (2026-10-04).
- **Toolchain.** `go.mod` declares `go 1.26.0` and `toolchain go1.26.8` (`6a5fcb1`), so
  CI, CD and the Docker builder use Go 1.26.8, and an older local `go` downloads it
  (SEC-018). Releases up to v1.1.0 were built with Go 1.26.0.
- **Validation** on the owner's machine (linux/amd64, non-root):

  | Check | Result |
  |---|---|
  | `gofmt -s -l .` | no files listed |
  | `go vet ./...` | clean |
  | `golangci-lint run` (v2.13.2) | 0 issues |
  | `go test -race -count=1 ./...` | pass; coverage 75.0% (`main`), 75.4% (`internal`) |
  | `go mod verify` | all modules verified |
  | gosec v2.29.0 | 10 findings, triaged under SEC-012 |
  | govulncheck v1.8.0 | go1.26.0: 3 reachable standard-library vulnerabilities; go1.26.8: none (SEC-018) |
  | Probe tests and real-binary checks | recorded under SEC-015–SEC-019 and BUG-015–BUG-024 |

- **Not verified:** Windows and macOS behaviour, the Docker build, fuzzing, and the Go
  version inside the published release binaries.
- **Lowest coverage:** `View` and `actionTitle` (0%), `readTerminalPassword` (20%),
  `readPassword` (46%), `Update` (54%), `replacePath` (56%), `writeZipFile` (59%).

## 2. Open defects (non-security)

| ID | Defect | Evidence | Location |
|---|---|---|---|
| BUG-011 | `keys export` doesn't check that the export password matches the key's master password, and `keys import` doesn't check that the inner blob decrypts. A `.ckey` can therefore need two different passwords to be usable. The CLI prompt for the export password reads "Enter master password:" and the flag help says "master password for export encryption", although the value only protects the export file, so users can't tell the two passwords apart. | Code review | `crypto.go` `ExportKeyToFile`, `ImportKeyFromFile`; `logic-cli.go` `newKeysExportCmd` |
| BUG-019 | **(Low; fixed 2026-10-04, not yet committed: plan 5.11)** The output-containment checks compare paths as text. `checkOutputOutsideDir` (BUG-013) misses an output reached through a symlink, or one spelled in different letter case on a case-insensitive file system (the macOS and Windows defaults). `checkInputOutsideOutput` resolves symlinks but compares case-sensitively, so on those systems `--force` could replace a folder that holds the archive, deleting the archive with it. | Probe: `compress tree --output link/self.tar.gz` with `link → tree` wasn't refused and failed with "archive/tar: write too long"; a small tree can instead embed its own partial output. The letter-case cases are from code review; not run on macOS or Windows. | `compress.go` `pathWithin`, `checkOutputOutsideDir`, `checkInputOutsideOutput` |
| BUG-020 | **(Low; fixed 2026-10-04, not yet committed: plan 5.12)** Compression options are silently overridden. `--format gzip --output x.zip` writes a zip, because `resolveCompressFormat` lets the `.zip` extension win over an explicit `gzip`. `--level` values outside 1–9 (0, 10, 42, −7) are accepted and replaced by the default. The TUI behaves the same. | Probe: zip magic `504b` from `--format gzip`; all four levels exited 0. | `compress.go` `resolveCompressFormat`, `CompressFileWithFormat`; `logic-cli.go` `newCompressCmd`; `logic-tui.go` compress action |
| BUG-021 | **(Low; fixed 2026-10-04, not yet committed: plan 5.13)** Names within about 20 bytes of the 255-byte limit can't be written, because the temporary name `.<name>.<random>.tmp` is too long ("file name too long"). Affects every output: encrypt, decrypt, compress, single-file decompress, extraction folders and key export. | Probe: a 240-character name (244 with `.enc`) failed both encrypt and compress; 236 worked. | `compress.go` `createAtomicFile`, `extractToDir` |
| BUG-022 | **(Low; fixed 2026-10-04, not yet committed: plan 5.16)** A single file whose name ends in `.tar` doesn't round-trip. `compress backup.tar` writes `backup.tar.gz`, which `decompress` treats as a tarball and extracts into `backup/` instead of restoring `backup.tar`. A tarball holding symlinks (common) fails with "unsupported entry type", so Cryptare can't give the file back at all. | Probe: a tar with a symlink entry, compressed and then decompressed: `extract archive: unsupported entry type "latest"`. | `compress.go` `isTarGzArchive`, `DecompressFileWithLimits` |
| BUG-023 | **(Low; fixed 2026-10-04, not yet committed: plan 5.14)** `--password ""` counts as "not given" and falls back to the prompt, which reads stdin. A script therefore can't pass an explicitly empty password, which decrypting a legacy empty-password file without a terminal needs; with no stdin the command fails with "read password: EOF". | Probe: `decrypt x.enc --password ""` → `read password: EOF`. | `logic-cli.go` `passwordFlags.get` |
| BUG-024 | **(Low; fixed 2026-10-04, not yet committed: plan 5.15)** Legacy decrypt and key import read their whole input with no size cap (`io.ReadAll`, `os.ReadFile`). Pointing `decrypt` at a large file that isn't a version 2 artifact, or `keys import` at a large file, exhausts memory instead of failing fast; real `.ckey` files are under 1 KiB. | Code review | `crypto.go` `DecryptFileWithLimits` (legacy branch), `ImportKeyFromFile` |

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
- **Leftover from another project.** `database_path_test.go` uses a `tasks.db` fixture
  (cosmetic; plan 4.6).
- **IDE files.** `.idea/` is tracked (`.gitignore` has `# .idea/` commented out).
  `.junie/plans/` is an empty, untracked agent workspace.
- **NOTICE is incomplete.** It ends at "This product includes third-party software:"
  with an empty list (Q-006).
- **TUI password field.** It masks input with one `*` per character, which reveals the
  password's length.
- **Documentation drift.**
  - The README still says every published release is broken ("Known issue", BUG-001),
    although v1.0.1 and later work (plan 4.7), and its release table lists
    `cryptare_windows_arm64.zip`, which isn't built (W6).
  - `CONTRIBUTING.md` says CGO-less builds "fail on every command at runtime"; only the
    `keys` commands and the TUI fail (`maint.md` §6).
  - Some README guarantees don't hold: "everything the tool writes is private to you"
    (SEC-019 on Windows). Missing parent folders were created 0755 until the
    2026-10-03 G301 fix (SEC-012). "Commands don't overwrite anything by default"
    (BUG-017) and "a command that fails part-way leaves no partial output" (SEC-015)
    hold since `6a5fcb1`.
- **CI runs `go test` with `-race` on Linux and macOS** since plan 5.19 (2026-10-04, not
  yet committed); Windows runs it without, as the race detector needs CGO there too.
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
| Q-003 | Should the default database move from `./cryptare.db` to a per-user location (e.g. `os.UserConfigDir()/cryptare/cryptare.db`)? How should existing `./cryptare.db` files be migrated? | SEC-010 and SEC-016: the working-folder default is what lets a planted database be used. A behaviour change for existing users. |
| Q-005 | Should `cryptare.db` be purged from git history and force-pushed? | SEC-003; irreversible, so it is the owner's call. |
| Q-006 | Was the `NOTICE` third-party list removed on purpose, to be regenerated? The file now ends mid-sentence. Where dependency attributions and license texts should live (NOTICE, or a bundled licenses file with the binaries) is a licensing decision for the owner. | Licensing hygiene. |
| Q-008 | Should the Argon2id settings accepted when reading (up to 1 GiB of memory, 10 passes, 16 lanes) be lowered, for example to 256 MiB (four times the 64 MiB default), or made configurable? | A crafted 62-byte file costs about 1 GiB of memory and 2 s per attempt (SEC-005), enough to kill a small VM or container. A lower limit restricts stronger settings in the future. |
| Q-009 | Should passwords for new data be Unicode-normalised (NFC, or NFKC as NIST SP 800-63B suggests), and should a UTF-8 byte-order mark at the start of a `--password-file` be dropped? | The same-looking password typed on another system or input method can produce different bytes, so a file may not decrypt there. Normalising needs a header or KDF flag so existing data stays readable, and possibly `golang.org/x/text` (already an indirect dependency) as a direct one. |

Answered or closed: Q-001 (release builds use native CGO per OS), Q-002 (stored keys
should be usable; plan 3.4), Q-004 (the default password policy, `maint.md` §4) and
Q-007 (v1.0.0 is left as it is), Q-010 (an untrusted key database is refused, not
warned about; SEC-016), Q-011 (a gunzip-only `decompress --raw`; BUG-022) and Q-012
(Windows permissions are documented, not enforced with ACLs; SEC-019). Their
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
```

Run probes that can use a lot of memory, such as decompression or allocation bombs,
inside a memory-capped cgroup, for example
`systemd-run --user --scope -p MemoryMax=768M -p MemorySwapMax=0 <command>`. An
address-space limit (`ulimit -v`) is no substitute: a low one breaks Go binaries (cgo
thread creation fails) before memory runs out, and on 2026-10-03 a probe limited only
by `ulimit -v` exhausted the whole machine's memory.
