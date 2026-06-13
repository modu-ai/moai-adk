# SPEC-SEC-HARDEN-002 — Acceptance Criteria

All criteria are reproduction-first: each RED scenario asserts the **CURRENT** (pre-fix) code resolves a malicious input to allow / traversal / dir-created-outside-root; the fix flips it to DENY / reject; the NO-REG scenario asserts legitimate inputs are unchanged. cycle_type = tdd.

Legend: **RED** = must FAIL on pre-fix code (defect present). **GREEN** = must PASS after fix. **NO-REG** = must STAY green before and after.

---

## M1 — `ValidateSpecID` sanitizer helper (leaf package `internal/cli/specid`)

> **Import-cycle resolution (run-phase amendment, 2026-06-14)**: The canonical helper is the EXPORTED `ValidateSpecID` in the NEW leaf package `internal/cli/specid` (package `specid`), NOT an unexported `validateSpecID` in package `cli`. Rationale: `internal/cli` already imports `internal/cli/worktree`, so `internal/cli/worktree` cannot import `internal/cli` (import cycle) — a package-`cli` helper is unreachable from `worktree/new.go` (M2a). The leaf package `internal/cli/specid` is importable by BOTH `internal/cli` (M3 call sites) AND `internal/cli/worktree` (M2a call site), and imports neither → no cycle. The validation logic (`..` / path-separator / absolute-path rejection) is defined exactly once, in the leaf package.

### AC-SEC2-M1-001 — Rejects `..` traversal
- **Given** the shared `specid.ValidateSpecID` helper exists in the leaf package `internal/cli/specid`
- **When** it is called with `"../../../../tmp/evil"`
- **Then** it returns a non-nil structured validation error (does not return nil/accept).
- **RED (pre-fix)**: helper does not exist in the leaf package → test references undefined symbol → RED.
- **GREEN verify**: `go test -run 'TestValidateSpecID' ./internal/cli/specid` → `PASS`

### AC-SEC2-M1-002 — Accepts legitimate canonical SPEC-ID
- **Given** the helper exists
- **When** it is called with `"SPEC-SEC-HARDEN-002"`
- **Then** it returns nil (accepts).
- **NO-REG verify**: `go test -run 'TestValidateSpecID' ./internal/cli/specid` includes a legitimate-ID sub-case asserting `err == nil`.

### AC-SEC2-M1-003 — Rejects path separators
- **Given** the helper exists
- **When** it is called with `"foo/bar"` (and `"foo\\bar"`)
- **Then** it returns a non-nil error for each.
- **GREEN verify**: `go test -run 'TestValidateSpecID' ./internal/cli/specid` → `PASS` (sub-cases for `/` and `\`).

### AC-SEC2-M1-004 — Single canonical impl in a leaf package, invoked (exported) at every call site
- **Given** the M2 and M3 call sites apply sanitization
- **Then** the validation logic is defined exactly ONCE in the leaf package `internal/cli/specid`, and the EXPORTED `ValidateSpecID` is invoked at every guarded call site (no per-call-site bespoke re-implementation — REQ-SEC2-M1-004 intent preserved).
- **Positive grep — single canonical definition (recursive over the leaf package)**: `grep -rc 'func ValidateSpecID' internal/cli/specid/` → exactly `1`. (Recursive over the leaf package, NOT a non-recursive `internal/cli/*.go` glob — the helper no longer lives in the `cli` package root.)
- **Positive grep — exported helper invoked at all FOUR call sites** (M2a worktree-new + M3's three spec subcommands): each of the following returns ≥ 1 match (adjust the import-alias prefix to whatever alias the implementation uses; the assertion is "the exported leaf-package helper is invoked at this site"):
  - `grep -n 'specid.ValidateSpecID(' internal/cli/worktree/new.go`
  - `grep -n 'specid.ValidateSpecID(' internal/cli/spec_view.go`
  - `grep -n 'specid.ValidateSpecID(' internal/cli/spec_status.go`
  - `grep -n 'specid.ValidateSpecID(' internal/cli/spec_close.go`
- **No bespoke re-implementation elsewhere**: `grep -rc 'func ValidateSpecID\|func validateSpecID' internal/cli/` → `1` (the leaf-package definition is the only one; no duplicate in the `cli` package root or `worktree`).
- This is a robust positive assertion (the desired post-fix state — single leaf-package helper present + invoked at all 4 sites) and is now SATISFIABLE under Go's import rules (the prior package-`cli` definition pin contradicted the package-`worktree` call requirement via an import cycle).

---

## M2 — `worktree new` boundary + `--` argv separator

### AC-SEC2-M2-001 — `worktree new` rejects traversal SPEC-ID before any dir creation
- **Given** the worktree-new path-construction site (new.go:108→141)
- **When** invoked as `moai worktree new '../../../../tmp/evil'`
- **Then** the command returns a structured rejection error BEFORE calling `os.MkdirAll` / `WorktreeProvider.Add`, and NO directory is created outside `~/.moai/worktrees/<project>/`.
- **RED command (pre-fix, demonstrates defect — DO NOT run against a real homedir without isolation; use the test harness)**: `moai worktree new '../../../../tmp/evil'` currently runs `filepath.Join(homeDir, ".moai", "worktrees", projectName, "../../../../tmp/evil")` + `os.MkdirAll`/`Add` → creates a worktree dir OUTSIDE the root.
- **GREEN verify**: `go test -run 'TestRunNew.*Traversal' ./internal/cli/worktree` → `PASS` (asserts rejection + no out-of-root dir via injected `userHomeDirFunc`/`getProjectNameFunc` pointing at `t.TempDir()`).

### AC-SEC2-M2-002 — Legitimate SPEC-ID still constructs the canonical worktree path (NO-REG)
- **Given** the guarded `worktree new`
- **When** invoked with a legitimate `SPEC-SEC-HARDEN-002`-style ID
- **Then** it constructs `~/.moai/worktrees/<project>/<SPEC-ID>` and proceeds unchanged.
- **NO-REG verify**: `go test ./internal/cli/worktree/...` → `ok` (existing worktree-new tests stay green).

### AC-SEC2-M2-003 — git worktree-add argv uses `--` before user-derived operands
- **Given** the git `worktree add` argv assembly (worktree.go:46 + 52)
- **When** a worktree path or branch beginning with `-` is supplied (`'--upload-pack=x'`)
- **Then** the argv contains a `--` end-of-options separator before the first user-derived operand, so the `-`-leading value is treated as a positional argument, not a git option.
- **RED command (pre-fix)**: `moai worktree new '--upload-pack=x'` currently assembles `git worktree add ... --upload-pack=x` with no `--` → git interprets it as an option (option smuggling).
- **GREEN verify**: `go test -run 'TestAdd.*DashDash|TestWorktreeAdd.*Separator' ./internal/core/git` → `PASS` (argv-capture asserts a `--` token immediately before the user-derived operand).
- **grep**: `grep -n '"--"' internal/core/git/worktree.go` → ≥ 1 match in the `Add` argv.

### AC-SEC2-M2-004 — `--path` escape hatch rejects `..` after Clean (minimal, not over-constrained)
- **Given** the documented `--path` user-custom flag
- **When** `--path` is supplied with a value containing `..`
- **Then** the command rejects it after `filepath.Clean` still shows a `..`-traversal; otherwise the documented `--path` escape hatch is NOT further constrained.
- **GREEN verify**: `go test -run 'TestRunNew.*PathFlag.*Traversal' ./internal/cli/worktree` → `PASS`.
- **NO-REG**: a legitimate absolute `--path` (no `..`) STILL works → existing `--path` tests stay green.

---

## M3 — spec view/status/close read boundaries (THREE files; `spec drift` EXCLUDED)

### AC-SEC2-M3-001 — `spec view` rejects traversal SPEC-ID before read-path construction
- **Given** `viewAcceptanceCriteria` (CLI handler `spec_view.go:30` `specID := args[0]` → join at `spec_view.go:49`)
- **When** invoked as `moai spec view '../../../../etc'`
- **Then** it rejects the input via `specid.ValidateSpecID` at the `args[0]` boundary BEFORE `filepath.Join(projectRoot, ".moai", "specs", specID)`, so no file outside `.moai/specs/` is read.
- **RED command (pre-fix)**: `moai spec view '../../../../etc'` currently joins to a path outside `.moai/specs/` and stats/reads `spec.md` there.
- **GREEN verify**: `go test -run 'TestSpecView.*Traversal|TestViewAcceptanceCriteria.*Traversal' ./internal/cli` → `PASS`.

### AC-SEC2-M3-002 — All THREE positional-SPEC-ID spec subcommands guarded at the CLI boundary
- **Given** exactly three spec subcommands accept a positional SPEC-ID: `spec view` (`spec_view.go:30`), `spec status` (`spec_status.go:52`), `spec close` (`spec_close.go:98`). `spec drift` is EXCLUDED — it has no positional SPEC-ID (RunE/PostRunE only, repo-wide drift command).
- **Then** each of the three applies the `specid.ValidateSpecID` guard at its CLI `args[0]` boundary. For `spec close`, the guard is at the CLI handler (`spec_close.go:98`, before `spec.Close(specID, opts)`); the path-join sink is the deeper transitive `internal/spec/closer.go:173`, which is NOT modified — guarding at the CLI boundary stops the traversal before it reaches that sink.
- **Positive grep — guard invoked in each of the THREE handlers** (each returns ≥ 1 match):
  - `grep -n 'specid.ValidateSpecID(' internal/cli/spec_view.go`
  - `grep -n 'specid.ValidateSpecID(' internal/cli/spec_status.go`
  - `grep -n 'specid.ValidateSpecID(' internal/cli/spec_close.go`
- **Exclusion assertion — `spec_drift.go` is NOT guarded and NOT modified**: `grep -c 'ValidateSpecID' internal/cli/spec_drift.go` → `0` (no positional SPEC-ID to guard).
- **Exclusion assertion — `closer.go` not touched**: `git diff --name-only` does NOT list `internal/spec/closer.go`.

### AC-SEC2-M3-003 — Legitimate SPEC-ID resolves unchanged (NO-REG)
- **Given** the three guarded spec subcommands (`spec view`, `spec status`, `spec close`)
- **When** invoked with a legitimate SPEC-ID
- **Then** each resolves to `.moai/specs/<SPEC-ID>/spec.md` (or performs the `spec.Close` operation for `spec close`) and behaves unchanged.
- **NO-REG verify**: `go test ./internal/cli/...` → `ok` (existing spec-subcommand tests stay green, including existing `spec close` / `spec drift` tests).

---

## M4 — Permission redirect-operator scanner extension

### AC-SEC2-M4-001 — Redirect operators past a `:*` prefix now DENY (RED → GREEN)
- **Given** an allow-listed `:*` prefix rule (e.g. `go test:*`)
- **When** the resolver matches against an input whose remainder contains an unquoted redirect operator
- **Then** the resolver reports NO match (the redirect-bearing command is not silently allowed).
- **RED cases (pre-fix resolve to ALLOW — must FAIL pre-fix)**:
  - `go test > /etc/cron.d/payload`
  - `go test >> ~/.bashrc`
  - `go test 2> /tmp/x`
  - `go test < /etc/shadow`
- **GREEN verify**: `go test -run 'TestMatches.*Redirect|TestHasUnquotedShellSeparator.*Redirect' ./internal/permission` → `PASS` (each of the 4 cases → deny / no-match).
- **grep**: `grep -n "c == '>'" internal/permission/stack.go` AND `grep -n "c == '<'" internal/permission/stack.go` → ≥ 1 match each (additive to the existing separator `case`).

### AC-SEC2-M4-002 — Additive extension of the existing quote-aware scanner (no rewrite)
- **Given** `hasUnquotedShellSeparator` already exists with the D1 unterminated-quote guard
- **Then** `>` and `<` are added to the SAME unquoted-separator `case` (line ~189), and the D1 guard (return `inSingle || inDouble`) is preserved verbatim.
- **grep**: `grep -n 'return inSingle || inDouble' internal/permission/stack.go` → 1 match (D1 guard preserved).
- **NO-REG verify**: `go test -run 'TestMatches_UnterminatedQuoteBypass' ./internal/permission` → `PASS` (D1 8-case suite stays green).

### AC-SEC2-M4-003 — Quoted redirect chars do NOT cause false rejection (NO-REG)
- **Given** the extended scanner
- **When** the input carries `>` or `<` only inside a quoted segment
- **Then** the resolver does NOT treat the quoted redirect as a command boundary (still matches the prefix rule).
- **NO-REG cases (must STAY ALLOW)**:
  - `go test -run 'TestGreater>'` (single-quoted `>`)
  - `go test "a > b"` (double-quoted `>`)
- **GREEN verify**: `go test -run 'TestMatches.*QuotedRedirect|TestHasUnquotedShellSeparator.*Quoted' ./internal/permission` → `PASS` (both stay matched/allow).

### AC-SEC2-M4-004 — `&>` / `>&` / `2>&1` remain conservatively denied (NO-REG, err-safe)
- **Given** these operators are already denied today via the `&` branch
- **Then** this SPEC does NOT special-case them back to allow; they remain denied.
- **NO-REG verify**: `go test ./internal/permission/...` → `ok` (full permission suite green; `&`-bearing inputs stay denied).

---

## §D.x — Quality Gate / Definition of Done

- All AC-SEC2-M{1,2,3,4}-* GREEN-verify commands return `PASS` / `ok`.
- All NO-REG suites stay green: `go test ./internal/cli/... ./internal/cli/worktree/... ./internal/core/git/... ./internal/permission/...` → `ok`.
- Cross-platform: `go build ./...` AND `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- Subagent boundary: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli internal/cli/worktree internal/permission internal/core/git | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//"` → 0 matches.
- Lint: `golangci-lint run --timeout=2m` → no NEW issues vs baseline.
- Coverage: new `internal/cli/specid` leaf package (`ValidateSpecID`) ≥ 85%; touched packages no regression.
- Reproduction-first proof: each milestone's RED test is demonstrably present and FAILED on the pre-fix tree (captured in run-phase evidence).
- Scope discipline: `git diff --name-only` touches ONLY M1-M4 files + this SPEC's artifacts; PRESERVE list untouched. M3 touches exactly THREE CLI files (`spec_view.go`, `spec_status.go`, `spec_close.go`) — NOT `spec_drift.go` and NOT `internal/spec/closer.go`.
