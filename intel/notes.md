# Engineering Notes

Durable engineering notes and unresolved technical questions. Security issues are
tracked in [`cybersec.md`](cybersec.md), and work sequencing in [`plan.md`](plan.md).

Last updated: 2026-09-24

## 1. Baseline snapshot (2026-09-23)

- **Branch state.** Local `main` was one commit ahead of `origin/main` (`aa27461`
  "stash", which deletes `cryptare.db`). The working tree had uncommitted changes:
  - `CONTRIBUTING.md` emptied (since rebuilt; see `history.md`);
  - the third-party list in `NOTICE` removed;
  - `.idea/Cryptare.iml` modified;
  - `.idea/inspectionProfiles/` untracked.
- **Tags.** `v1.0.0` (annotated, on `768f4cd`, 2026-09-13) is pushed, and a GitHub
  Release with CD-built assets exists.
- **Validation performed.** This was done in an isolated analysis environment, not on
  the owner's machine. Go 1.26.8 was built from source. All modules were fetched and
  checked against `go.sum`, and `go mod verify` reported "all modules verified".

  | Check | Result |
  |---|---|
  | `gofmt -s -l .` | no files listed |
  | `go mod tidy` | no diff |
  | `go vet ./...` | clean |
  | `golangci-lint run` (v2.13.2) | 0 issues |
  | `go test -race -count=1 ./...` | pass; total coverage 59.8% |
  | `go build` (CGO on) | ok |
  | `CGO_ENABLED=0 go build` | builds, but fails at runtime (BUG-001) |
  | gosec v2.29.0 | 30 findings, triaged into `cybersec.md` |

- **Not verified.**
  - Windows and macOS behaviour.
  - The Docker image build.
  - Recent GitHub Actions run results.
  - govulncheck (its database was unreachable).

## 2. Confirmed defects (non-security)

| ID | Defect | Evidence | Location |
|---|---|---|---|
| BUG-001 | Release binaries don't work: every command, including `--help`, exits with `go-sqlite3 requires cgo to work. This is a stub`. **Fixed in v1.0.1 (2026-09-24):** every release target is built natively with CGO; the v1.0.1 assets were checked (build info, checksums, and a run of the linux/amd64 binary). The broken v1.0.0 assets are still published (Q-007). | Downloaded the v1.0.0 `cryptare_linux_amd64.tar.gz` (checksum matched `checksums.txt`) and ran it; a local `CGO_ENABLED=0` build behaves the same way. | `.github/workflows/cd.yml` (`CGO_ENABLED=0`); `ci.yml` cross-builds the same way and never runs the binaries |
| BUG-002 | The TUI crashes (panics) when an unhandled key is pressed in a form: Left, Right, Delete, Home, End, Ctrl+U, and other key types not listed in the switch. **Fixed 2026-09-24 (uncommitted):** unhandled form keys are now ignored; see plan 1.3. | A probe called `updateForm` with each key type; each one panicked with `unhandled default case`. | `internal/logic-tui.go` `updateForm` default branch; similar `panic` defaults in `Update` and `buildActionCmd` |
| BUG-003 | `compress <file> --output <same file>` destroys the source: the output is opened with `O_TRUNC` before the source is read. | Probe: a 4096-byte file became a gzip stream that decompresses to 0 bytes; no error was returned. | `internal/compress.go` `CompressFileWithFormat` |
| BUG-004 | Existing outputs are overwritten silently, and writes are not atomic. Decrypting `x.enc` replaces an existing `x`; compress, decompress and extract truncate existing files; a failure mid-write leaves partial output. | Code review | `crypto.go` `DecryptFile`, `EncryptFile`; `compress.go` writers |
| BUG-005 | Every command opens (and creates) the SQLite database, including commands that don't use it. | `cryptare --help` created `cryptare.db` | `main.go` (also SEC-010) |
| BUG-006 | `keys export` without `--output` reports a filename built from a second `time.Now()` call, which can differ from the file actually written. | Code review | `logic-cli.go` `newKeysExportCmd`; `logic-tui.go` export action; `crypto.go` `ExportKeyToFile` |
| BUG-007 | Extension handling is inconsistent. `.zip` detection ignores case, but `.gz`/`.tgz`/`.tar.gz` detection and default output naming are case-sensitive. For example, `FOO.ZIP` extracts to `FOO.ZIP.dec/`, and an uppercase `.TAR.GZ` without a `.tar` gzip header name is written out as a raw tar file. | Code review | `compress.go` `DecompressFile`, `defaultDecompressOutput`, `isTarGzArchive` |
| BUG-008 | `DashboardModel.busy` is set but never checked, so a second TUI action can be started while one is still running. | Code review | `ui-dashboard.go`, `logic-tui.go` |
| BUG-009 | `-ldflags -X main.version=…` in CD has no effect (no such variable), and there is no `--version` flag. **Fixed 2026-09-24 (uncommitted):** `main.version` (default `dev`) and `--version`/`-v`; see plan 3.6. | `go version -m` on a probe build; `cryptare --version` → `unknown flag` | `main.go`, `cd.yml` |
| BUG-010 | Encryption and decryption read whole files into memory, so memory use grows with file size and large inputs can exhaust RAM. | Code review (`os.ReadFile`) | `crypto.go` |
| BUG-011 | `keys export` doesn't check that the export password matches the key's master password, and `keys import` doesn't check that the inner blob decrypts. A `.ckey` can therefore need two different passwords to be usable. | Code review | `crypto.go` `ExportKeyToFile`, `ImportKeyFromFile` |

## 3. Design observations

- **Stored keys are never used for file encryption.** `encrypt` and `decrypt` derive
  keys from passwords only, and no code path consumes `key_models`. Key management is
  currently standalone (see Q-002).
- **CLI and TUI duplicate flows.** Key generate, export and import are implemented
  twice (`logic-cli.go` and `logic-tui.go` `buildActionCmd`). They have already
  drifted: the TUI accepts empty passwords and the CLI doesn't.
- **Dead code.** The `Storage` interface is declared but unused.
- **Stale comment.** `ImportKeyFromFile` says "Parse minimal JSON manually to avoid
  import cycle", but it uses `encoding/json`.
- **Leftovers from another project.** `database_path_test.go` uses `tasks.db`
  (cosmetic). The Munus names in the workflows, including the `munus:<sha>` image in
  `docker.yml`, are fixed in the 2026-09-24 workflow patch (plan W3).
- **IDE files.** `.idea/` is tracked (`.gitignore` has `# .idea/` commented out).
  `.junie/plans/` is an empty, untracked agent workspace.
- **README drift.** The previous install section named `cryptare-<os>-<arch>`
  binaries, but CD publishes `cryptare_<os>_<arch>.tar.gz`/`.zip` plus
  `checksums.txt`, including `windows_arm64`. Also, zip support (2026-09-16) and vim
  bindings (2026-09-17) were added after `v1.0.0` (2026-09-13).
- **NOTICE is incomplete.** The working tree ends at "This product includes
  third-party software:" with an empty list (Q-006).
- **TUI password field.** It masks input with one `*` per character, which reveals the
  password's length.

## 4. Open questions (owner decisions)

| ID | Question | Why it matters |
|---|---|---|
| Q-001 | How should releases handle CGO? (a) Build each target natively with CGO, using an OS matrix or a cross C toolchain (e.g. zig cc). (b) Switch to a pure-Go SQLite driver, which adds a new dependency. (c) Drop SQLite. **2026-09-24: option (a) chosen.** The staged `cd.yml` builds each target natively with CGO (static Linux binaries; darwin/amd64 cross-compiled on Apple silicon). **Closed 2026-09-24:** verified by the working v1.0.1 release. | Blocks BUG-001, which means every published binary is unusable. |
| Q-002 | Should stored keys be usable for file encryption (e.g. `encrypt --key <id>`), or is the key store meant to stay standalone? | Decides the value of the key-management feature and shapes the format work (SEC-005). |
| Q-003 | Should the default database move from `./cryptare.db` to a per-user location (e.g. `os.UserConfigDir()/cryptare/cryptare.db`)? How should existing `./cryptare.db` files be migrated? | SEC-010; a behaviour change for existing users. |
| Q-004 | What password policy applies on encrypt: a minimum length, and confirmation (typing it twice)? | SEC-001/SEC-002; a typo on encrypt currently makes data unrecoverable. |
| Q-005 | Should `cryptare.db` be purged from git history and force-pushed? | SEC-003; irreversible, so it is the owner's call. |
| Q-006 | Was the `NOTICE` third-party list removed on purpose, to be regenerated? The file now ends mid-sentence. Where dependency attributions and license texts should live (NOTICE, or a bundled licenses file with the binaries) is a licensing decision for the owner. | Licensing hygiene. |
| Q-007 | What should happen to the broken v1.0.0 release: mark it as broken, remove its assets, or supersede it with v1.0.1? v1.0.1 (working) is now available as the replacement. | Users downloading v1.0.0 get non-working binaries. |

## 5. Reproducing the validation locally

From the repository root, on a machine with Go 1.26 and a C compiler:

```bash
gofmt -s -l .
go mod tidy && git diff --exit-code go.mod go.sum
go vet ./...
golangci-lint run          # v2.13.2
go test -race -count=1 ./...
CGO_ENABLED=0 go build -o /tmp/cryptare-nocgo . && (cd "$(mktemp -d)" && /tmp/cryptare-nocgo --help)   # demonstrates BUG-001
```
