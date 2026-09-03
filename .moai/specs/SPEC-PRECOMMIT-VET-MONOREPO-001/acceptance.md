# SPEC-PRECOMMIT-VET-MONOREPO-001 — Acceptance Criteria

> 12 ACs (AC-PVM-001..012), each independently verifiable by a command or file check. "Twin pair" = the Go constant `preCommitHookContent` (`internal/cli/hook_install_precommit.go`) and the template file `internal/template/templates/.git_hooks/pre-commit` — every AC about hook content names BOTH surfaces because they are byte-identical by contract (REQ-PVM-006). Commands run from the worktree root (`.claude/worktrees/t237`); every PASS cites the command, its verbatim observed output, and the tree SHA of the run (VCI §2 attribution).

## §1 Twin-pair and core-fix scenarios

**AC-PVM-001 — Twin pair byte-identical (guard; green today, must stay green on every run-phase commit)**
Given both surfaces of the twin pair after any run-phase commit, When `go test -count=1 ./internal/cli/ -run 'TestPreCommitTemplateMatchesConstant'` runs, Then it exits 0 — the constant and the template file are byte-identical (REQ-PVM-006).
Verification: the command above, exit code observed unpiped, on the commit under test. Any RED is a twin-edit discipline breach, not progress.

**AC-PVM-002 — Submodule clean pass (RED on the old constant; flips GREEN at M1)**
Given a fixture repo (inside `t.TempDir()`) whose ONLY `go.mod` lives in `submod/` and a staged, gofmt-clean, vet-clean `.go` file at `submod/clean.go`, When the installed hook runs (`runPreCommitHook`, shim PATH with go present and moai absent), Then the hook exits 0 and stderr contains no `FAILED` (REQ-PVM-007).
Verification: `go test -count=1 ./internal/cli/ -run 'TestPreCommitHook_SubmodulePassesClean'` → `ok`. RED-now cell: against the pre-edit constant pair this exits 1 with `FAILED: go vet reported issues` — the fixture-based reproduction of the reported defect (module-resolution failure misreported as a vet finding). RED is captured at M1's opening act BEFORE the edit lands; adoption gate per verification-completeness §2 (failing command + verbatim output + exit code + fixture + tree SHA).

**AC-PVM-003 — Submodule vet block with module-root naming (vacuous-green guard; RED on old via the stderr discriminator)**
Given the same fixture with a staged, gofmt-clean file carrying a real vet diagnostic (Printf verb/arg mismatch) at `submod/bad.go`, When the installed hook runs, Then the hook exits 1 AND stderr contains `module root: submod` (REQ-PVM-008, REQ-PVM-003).
Verification: `go test -count=1 ./internal/cli/ -run 'TestPreCommitHook_SubmoduleVetBlocks'` → `ok`. RED-now note: the OLD constant ALSO exits 1 on this fixture (for the wrong reason — module-resolution failure), so the exit-code assertion alone is green-on-old and would be vacuous; the missing `module root: submod` stderr line is the discriminating assertion. The RED cell must state this reason.

**AC-PVM-004 — Repo-root fallback unchanged (guard; green today, must stay green)**
Given a repo whose `go.mod` IS at the repo root and a staged vet-dirty file, When the hook runs, Then the block behaves exactly as before: exit 1 with the go vet hint. The repo-root path is the M=`.` degenerate case of the module-root walk (REQ-PVM-002).
Verification: `go test -count=1 ./internal/cli/ -run 'TestPreCommitHook_GoVetBlocks'` → `ok`, test body unmodified (`internal/cli/hook_install_precommit_test.go:504`).

**AC-PVM-005 — Build tags read from the repo root before any cd (content-ordering check, both surfaces)**
Given the re-authored vet block in BOTH surfaces of the twin pair, When the line positions of the `BT_TAGS` read and the first per-module directory change are compared, Then the build-tags block (`.moai/config/build-tags` read + `BT_TAGS` assignment) precedes the first `cd` into a module root in BOTH files (REQ-PVM-004).
Verification (shape check, identifiers per the adopted reference patch):
```sh
for f in internal/cli/hook_install_precommit.go internal/template/templates/.git_hooks/pre-commit; do
  bt=$(grep -n 'BT_TAGS=""' "$f" | head -1 | cut -d: -f1)
  mr=$(grep -n 'cd "\$_mr"' "$f" | head -1 | cut -d: -f1)
  echo "$f bt=$bt first_cd=$mr"; [ -n "$bt" ] && [ -n "$mr" ] && [ "$bt" -lt "$mr" ] || echo "ORDER-FAIL $f"
done
```
Expected: both files print `bt=` < `first_cd=` and no `ORDER-FAIL`. (If run-phase re-authors identifier names, the check is re-derived against the shipped names; the invariant — tags read before the first directory change, in both surfaces — does not move.)

**AC-PVM-006 — Per-module vet isolation: subshell form present in both surfaces (REQ-PVM-001, REQ-PVM-005)**
Given the re-authored vet block, When both surfaces are grepped for the per-module invocation shape, Then each contains a subshell-scoped directory change of the form `( cd "<module-root-var>" && go vet …` — the cd does not leak past the invocation, so the heavy-gate block still runs from the repo root.
Verification: `grep -c '( cd "\$_mr" && go vet' internal/cli/hook_install_precommit.go internal/template/templates/.git_hooks/pre-commit` → 1 match in each file (identifier per the adopted reference patch; re-derived if renamed).

## §2 Preserved behaviors (guards; green today, must stay green)

**AC-PVM-007 — gofmt block untouched**
`go test -count=1 ./internal/cli/ -run 'TestPreCommitHook_GofmtBlocks'` → `ok` (REQ-PVM-009).

**AC-PVM-008 — Skip-bypass untouched**
`go test -count=1 ./internal/cli/ -run 'TestPreCommitHook_SkipBypass'` → `ok` (REQ-PVM-009).

**AC-PVM-009 — Toolchain-absent and no-staged-Go skips untouched (16-language audience guarantee)**
`go test -count=1 ./internal/cli/ -run 'TestPreCommitHook_ToolchainAbsent|TestPreCommitHook_NoStagedGo'` → `ok`, both test bodies unmodified (REQ-PVM-009).

## §3 Sweep and coordination

**AC-PVM-010 — Full package sweep green**
Given the completed run phase, When `go test -count=1 ./internal/cli/...` runs with a timeout of at least 600s and the exit code observed directly (unpiped), Then it exits 0 with no test failures.
Verification: the command above (Bash timeout ≥ 600000ms; `go test -timeout` ≥ 600s if the flag is passed); exit code recorded verbatim in progress.md §E.3 with the tree SHA.

**AC-PVM-011 — Cross-platform build green**
`go build ./...` exit 0 AND `GOOS=windows GOARCH=amd64 go build ./...` exit 0 on the run-phase tree. (The contrast tests skip on Windows by design — POSIX-only shim construction, the `TestPreCommitHook_ToolchainAbsent` precedent — so the windows CI leg sees skips, not failures.)

**AC-PVM-012 — Release-coordination line carried (file check, not a test)**
Given the SPEC's closing (run-phase completion report and progress.md §E.3), When the lead reads them, Then both carry this one line: **t230's landing precondition is satisfied (`32d2221fa` + `539349c5b` are develop ancestors); the remaining "at least one release must pass after t230's landing before this ships" is a deployment-time concern owned by release card t204 and does not block this SPEC's phases.**
Verification: `sed -n '/^## §E.3/,/^## §E.4/p' .moai/specs/SPEC-PRECOMMIT-VET-MONOREPO-001/progress.md | grep -c 'release card t204'` ≥ 1 — scoped to the §E.3 close record only. The §E.1 plan-phase pre-mention and the §E.3 placeholder's own "naming t204" scaffolding phrase do not satisfy this check (measured 0 on the plan-phase tree); only the mandated sentence form `release card t204`, as quoted in this AC, flips it.

## §4 Quality gates

- `go test -count=1 ./internal/cli/...` exit 0 (AC-PVM-010); `internal/cli` coverage reported against the 90% critical-package target.
- `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (AC-PVM-011).
- golangci-lint: zero NEW findings vs the M1 pre-flight baseline.
- RED evidence for AC-PVM-002 and AC-PVM-003 captured verbatim against the PRE-EDIT constant pair at M1's opening act, before any edit lands (verification-completeness two-cell adoption; M2 records the GREEN cells and the flip).
- Every PASS row in the final matrix names its command, verbatim output, and tree SHA (VCI §2) — no attributed-to-nothing greens.
