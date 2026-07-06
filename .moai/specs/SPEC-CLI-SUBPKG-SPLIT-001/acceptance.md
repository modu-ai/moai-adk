# Acceptance — SPEC-CLI-SUBPKG-SPLIT-001

Given-When-Then acceptance criteria for the phased `internal/cli` subpackage split. Because this
is a behavior-preserving refactor of working code, the acceptance model is CHARACTERIZATION: the
existing test suite + cross-platform build + `moai --help` output are the behavior contract every
milestone must preserve.

## §A. Per-Milestone Behavior-Preservation Gate (applies to EVERY milestone)

### AC-CSS-001 — Full test suite green after each milestone (REQ-CSS-002)
- **Given** a completed cluster extraction milestone (files + tests moved, symbols re-scoped),
- **When** `go test ./...` is run,
- **Then** it exits 0 with zero failures — the moved white-box tests pass in their new package.
- **Verification**: `go test ./... 2>&1 | tail -3` → no `FAIL`; compare pass set to `/tmp/cli-before.txt`.

### AC-CSS-002 — Cross-platform build green after each milestone (REQ-CSS-003)
- **Given** a completed milestone (including moved platform-tagged siblings),
- **When** the cross-platform build matrix runs,
- **Then** both invocations exit 0.
- **Verification**: `go build ./...` → exit 0 AND `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.

### AC-CSS-003 — `moai --help` subcommand list unchanged (REQ-CSS-001)
- **Given** a completed milestone,
- **When** `moai --help` is captured,
- **Then** the subcommand names, groups, and order are identical to the pre-milestone snapshot.
- **Verification**: `go run ./cmd/moai --help > /tmp/help-after.txt; diff /tmp/help-before.txt /tmp/help-after.txt` → empty diff.

### AC-CSS-004 — Exactly-once command registration, no cobra panic (REQ-CSS-005)
- **Given** an extracted command wired via `rootCmd.AddCommand(<domain>.<Cmd>)`,
- **When** `moai <subcommand> --help` and `moai --help` run,
- **Then** the command resolves once (no duplicate-`Use` panic, no missing command).
- **Verification**: `go run ./cmd/moai <subcommand> --help` exit 0; `grep -c 'AddCommand(<domain>' internal/cli/root.go` = 1.

## §B. Structural Correctness Gates

### AC-CSS-005 — Extracted cluster is a proper leaf package (REQ-CSS-004)
- **Given** an extracted cluster at `internal/cli/<domain>/`,
- **When** the package is inspected,
- **Then** every file declares `package <domain>`, exactly the registered cobra command is
  exported, and cluster-internal symbols remain unexported.
- **Verification**: `head -1 internal/cli/<domain>/*.go | grep -c 'package <domain>'` = file count;
  `go doc ./internal/cli/<domain>` shows the exported command only.

### AC-CSS-006 — No import cycle; kernel helpers via `uikit` (REQ-CSS-006)
- **Given** a kernel-dependent cluster extraction,
- **When** the build runs,
- **Then** the subpackage imports `internal/cli/uikit` (NOT `internal/cli`), and no import cycle
  exists.
- **Verification**: `go build ./...` exit 0 (Go reports import cycles at build time);
  `grep -rn '"github.com/modu-ai/moai-adk/internal/cli"' internal/cli/<domain>/` → no match
  (subpackage never imports package cli).

### AC-CSS-007 — deps via provider injection (REQ-CSS-007)
- **Given** a deps-coupled cluster extraction (e.g. `update`),
- **When** the command runs,
- **Then** dependencies arrive via a package-level provider set in `root.go` (not via a global
  `deps` reference inside the subpackage), and `InitDependencies` stays in `package cli`.
- **Verification**: `grep -rn '\bdeps\b' internal/cli/<domain>/` → no reference to the `package cli`
  global; `grep -n 'Provider\|Deps' internal/cli/root.go` shows the injection; command runs correctly.

### AC-CSS-008 — Public entry point unchanged (REQ-CSS-008)
- **Given** the full refactor,
- **When** `cmd/moai/main.go` is inspected,
- **Then** it still calls `cli.Execute()` unchanged and builds.
- **Verification**: `grep -n 'cli.Execute' cmd/moai/main.go` present; `go build ./cmd/moai` exit 0.

### AC-CSS-009 — No AskUserQuestion introduced (subagent boundary, C-HRA-008)
- **Given** any extracted subpackage,
- **When** grepped,
- **Then** no `AskUserQuestion`/`mcp__askuser` call exists.
- **Verification**: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/<domain>/ | grep -v '_test.go' | grep -v '^[^:]*:[0-9]*:[ \t]*//'` → no output.

## §C. Phasing Discipline Gates

### AC-CSS-010 — One cluster per milestone, atomic commit (REQ-CSS-009)
- **Given** a milestone commit,
- **When** its diff is inspected,
- **Then** it touches exactly one cluster's files (+ their tests) + the `root.go` wiring for that
  cluster + (if kernel-dependent) `uikit` — no second cluster.
- **Verification**: `git show --stat <milestone-sha>` shows one `internal/cli/<domain>/` dir + root.go.

### AC-CSS-011 — No functional change / no test weakening (REQ-CSS-011, REQ-CSS-012)
- **Given** a milestone diff,
- **When** reviewed,
- **Then** no logic edit beyond symbol re-scoping + import rewiring; no test deleted or skipped;
  test count is preserved (moved, not removed).
- **Verification**: `git show <sha>` — non-test diffs are `package`/export/import lines only;
  test-file count before == after (moved); `grep -rn 't.Skip' internal/cli/<domain>/` unchanged.

### AC-CSS-012 — Checkpoint stop-condition honored (REQ-CSS-010)
- **Given** the post-M4 checkpoint,
- **When** marginal value is judged insufficient,
- **Then** the work stops with M1-M4 shipped and the decision is recorded — no forced completion
  of M5-M7.
- **Verification**: checkpoint decision recorded in `progress.md`; if STOP chosen, no M5+ commits.

## §D. Quality Gates (per milestone)

- `go test ./...` exit 0 (AC-CSS-001) — the binding gate.
- `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (AC-CSS-002).
- `go vet ./...` exit 0.
- `golangci-lint run --timeout=2m` — no NEW issues vs the pre-flight baseline (pre-existing baseline
  reported separately).
- Coverage: per-package coverage of the moved cluster ≥ the pre-move coverage (tests moved intact,
  so coverage is preserved by construction — verify not regressed).

## §E. Definition of Done

A milestone is DONE when:
1. AC-CSS-001..004 pass (behavior preserved: tests green, cross-platform build green, `moai --help`
   diff empty, command registered exactly once).
2. AC-CSS-005..009 pass (proper leaf package, no import cycle, deps injected, entry point stable,
   no AskUserQuestion).
3. AC-CSS-010..011 pass (atomic one-cluster commit, no functional change, no test weakening).
4. `go vet` + `golangci-lint` clean of NEW issues.
5. The milestone is committed with a Conventional-Commit subject
   `refactor(SPEC-CLI-SUBPKG-SPLIT-001): M<N> extract <cluster> cluster`.

The SPEC as a whole is DONE when the chosen milestone set (per the §F.9 recommendation and the
post-M4 checkpoint decision) is shipped — NOT necessarily when all of M1-M7 complete (AC-CSS-012).
