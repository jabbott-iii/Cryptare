# Contributing to Cryptare

Thanks for your interest in improving Cryptare. This guide covers setup, workflow,
validation, coding expectations and pull-request expectations.

Architecture and maintainability rules live in [`intel/maint.md`](intel/maint.md),
which is authoritative; this guide must stay consistent with it. Also read
[`AGENTS.md`](AGENTS.md) and the other [`intel/`](intel/) documents.

## Ground rules

- Create an issue to pitch an addition or change. Pull requests with no corresponding
  issue will be denied.
- If the issue already exists, comment on it before making a pull request to address
  it.
- Confirmation of a pitched concept is required on the issue before making a pull
  request that modifies the code base.
- The license header must be kept on every source file.
- Run `gofmt -s -w .` at the repository root before making a pull request.
- Follow the [Code of Conduct](CODE_OF_CONDUCT.md).

## Development setup

### Prerequisites

- **Go 1.26 or newer** (`go.mod` declares `go 1.26.0`).
- **A C compiler (gcc or clang) with CGO enabled.** The SQLite driver
  (`github.com/mattn/go-sqlite3`, pulled in by `gorm.io/driver/sqlite`) requires CGO.
  Builds with `CGO_ENABLED=0` compile but fail on every command at runtime.
- **Git.**
- Optional:
  - [golangci-lint](https://golangci-lint.run/) v2.13.2 (the version CI uses);
  - Docker, for the container build;
  - the dev container in [`.devcontainer/`](.devcontainer/).

### Build and run

```bash
git clone https://github.com/jabbott-iii/Cryptare.git
cd Cryptare
go build -o cryptare .
CRYPTARE_DB_PATH="$HOME/.cryptare-dev.db" ./cryptare --help
```

Every run opens (and creates, if missing) the SQLite key store at `CRYPTARE_DB_PATH`,
or at `./cryptare.db` when the variable is unset. `*.db` is git-ignored. **Never commit
a database file.**

## Workflow

1. Open or find the issue, and get confirmation before starting (see Ground rules).
2. Create a branch from `main`.
3. Read `AGENTS.md` and the `intel/` documents relevant to your change.
4. Make the smallest complete change. Avoid unrelated refactoring, formatting churn,
   renames and dependency upgrades.
5. Add or update tests (see below).
6. Update documentation:
   - `README.md` for user-visible behaviour;
   - `intel/` documents as `AGENTS.md` requires (`map.md` for structure,
     `cybersec.md` for security work, an entry appended to `history.md` for
     significant changes).
7. Run the validation commands, then open a pull request that references the issue.

## Validation

These mirror CI (`.github/workflows/ci.yml`). Run them from the repository root:

```bash
gofmt -s -l .                                        # must print nothing
go mod tidy && git diff --exit-code go.mod go.sum    # no diff unless you changed dependencies
go vet ./...
golangci-lint run                                    # v2.13.2
go test ./...                                        # add -race when a C toolchain is available
```

CI runs on Ubuntu, Windows and macOS. After the checks above, it builds the binary with
CGO and smoke-tests it: `--version`, `keys generate`/`keys list`, and an encrypt/decrypt
round-trip. On every push and pull request, CodeQL and gosec also run
(`security.yml`). Pull requests to `main` get a Docker build smoke test
(`docker.yml`).

## Coding expectations

The full rules are in [`intel/maint.md`](intel/maint.md). In short:

- **Layering.** `internal/crypto.go` and `internal/compress.go` stay UI-agnostic: no
  printing, prompting, Cobra or Bubble Tea.
- **CLI and TUI stay in step.** A behaviour change must land in both `logic-cli.go`
  and `logic-tui.go`, with tests for each.
- **Validation that protects data goes in the core layer**, so both interfaces
  inherit it.
- **Compatibility.** Existing encrypted files, directory artifacts, stored keys and
  `.ckey` exports must stay readable. Format changes need a versioned header and a
  legacy read path.
- **Error handling.** Wrap errors with context (`fmt.Errorf("…: %w", err)`) and check
  errors from output writes. Use `closeWithError` for deferred closes on writers.
- **Filesystem.** Write outputs with mode `0o600`. Keep rejecting symlinks, special
  files and path traversal.
- **Dependencies.** Don't add one when the standard library or an existing dependency
  already does the job.
- **Tests.** Use `t.TempDir()`; close database handles (Windows CI depends on it);
  follow the existing CLI (`SetArgs`/`SetIn`/`SetOut`) and TUI (`tea.KeyMsg`) test
  patterns.

## Security

- Never commit secrets, passwords, keys, `.env` files or database files.
- Treat changes to encryption, key storage, password handling or archive extraction as
  security-sensitive. Say so in the pull request and include regression tests.
- Don't weaken or bypass a remediation recorded in
  [`intel/cybersec.md`](intel/cybersec.md), and don't disable tests, linters or
  security scans to get a build to pass.
- Don't report suspected vulnerabilities in public issues. Contact the maintainer
  (@jabbott-iii) privately.

## Pull request expectations

- It links an issue on which the change was confirmed.
- Its scope is focused on that issue.
- The description says what changed and why, and lists the validation commands you
  ran with their results.
- CI passes on all three operating systems.
- CLI flags and on-disk formats stay backward-compatible, unless the issue explicitly
  approves a breaking change.
- Documentation is updated (`README.md`, `intel/`).
- Every change is reviewed by the code owner, @jabbott-iii (see `CODEOWNERS`).
