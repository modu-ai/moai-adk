# t596 — new files could be written outside the project through a symlinked parent

Source finding: hooks audit 2026-09-11, H01 (P1). Target: `internal/hook/pre_tool.go`.

- Card: t596
- Worktree: `.claude/worktrees/t596`
- Branch: `WT-symlink-write-escape`
- Base commit (tree measured): `eabce74448e094dd1a4393044138a04b2016af99` (origin/develop)
- Applicable rules: `.claude/rules/moai/core/verification-claim-integrity.md` §1.1, §2; `.claude/rules/moai/development/verification-completeness.md` §1.1, §2

## Claim

1. The audit finding H01 REPRODUCES on the current develop tip — it was not already repaired.
2. The repair makes the escape visible while preserving the new-file allow path.
3. The repair required a second, adjacent correction (allowlist normalization symmetry) to avoid introducing a false-deny regression.

## Evidence

### E1 — reproduction on the unrepaired tree (RED)

Tree: `eabce74448e094dd1a4393044138a04b2016af99`, before any edit to `pre_tool.go`.

```
go test ./internal/hook/ -run 'TestCheckFileAccess_NewFileUnderSymlinkedParentBlocked|TestCheckFileAccess_NewFileUnderInProjectSymlinkedParentAllowed' -v -count=1
```

```
--- FAIL: TestCheckFileAccess_NewFileUnderSymlinkedParentBlocked (0.00s)
    pre_tool_security_test.go:169: new file under escaping symlinked parent: decision="" reason="", want "deny"
--- PASS: TestCheckFileAccess_NewFileUnderInProjectSymlinkedParentAllowed (0.01s)
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.968s
```

Exit code: 1. Swept set: 2 selectors → 2 tests run (1 FAIL, 1 PASS) — not an empty sweep.

The observed `decision="" reason=""` matches the audit's recorded observation
(`new_symlink_file decision="" reason="" / outside_write=true`) exactly.

### E2 — mechanism

`checkFileAccess` resolves symlinks with `filepath.EvalSymlinks(resolvedPath)`. For a
not-yet-existing leaf that call fails, and the guard falls back to the FULL unresolved
path with `resolvedSymlink=false`. The boundary check then compares lexically:
`<project>/linked/new.txt` yields `rel = "linked/new.txt"`, no `..` prefix, allow — while
the Write tool follows `linked -> /outside` and lands outside the project. The CWE-61
escape sits on the PARENT, not on the leaf, which is why the existing leaf-resolving
guard does not see it.

### E3 — repair, GREEN

`resolveThroughExistingParent` resolves the nearest EXISTING ancestor and rejoins the
not-yet-existing remainder, used only on the branch where whole-path `EvalSymlinks` fails.

```
go test ./internal/hook/ -run 'TestCheckFileAccess_NewFileUnderSymlinkedParentBlocked|TestCheckFileAccess_NewFileUnderInProjectSymlinkedParentAllowed|TestCheckFileAccess_NewFileUnderSymlinkedAllowedExternalDirAllowed|TestCheckFileAccess_NewFileWriteFallback' -v -count=1
```

```
--- PASS: TestCheckFileAccess_NewFileWriteFallback (0.00s)
--- PASS: TestCheckFileAccess_NewFileUnderInProjectSymlinkedParentAllowed (0.00s)
--- PASS: TestCheckFileAccess_NewFileUnderSymlinkedParentBlocked (0.00s)
--- PASS: TestCheckFileAccess_NewFileUnderSymlinkedAllowedExternalDirAllowed (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/hook	0.981s
```

Exit code: 0. Swept set: 4 selectors → 4 tests run, 4 PASS.

### E4 — adjacent regression found and closed

Resolving the parent changes `resolvedPath` into its resolved form, but
`isAllowedExternalPath` compared it against allowlist entries normalized with
`filepath.Abs` only. An allowlist entry is routinely itself a symlink (macOS
`/tmp -> /private/tmp`), so a new file genuinely inside an allowed directory would have
read as an escape and been denied. The comparison now applies the same normalization to
both sides (lexical AND `EvalSymlinks`-resolved forms of each entry).

This also repairs the same asymmetry on the pre-existing existing-file path, where
`resolvedPath` was already resolved; it is not scope creep but the minimum needed to keep
the primary repair from regressing behavior.

### E5 — mutant probes (both kill cleanly)

Mutant A — parent resolution disabled (`&& false` on the new branch):

```
--- FAIL: TestCheckFileAccess_NewFileUnderSymlinkedParentBlocked (0.00s)
--- PASS: TestCheckFileAccess_NewFileWriteFallback
--- PASS: TestCheckFileAccess_NewFileUnderInProjectSymlinkedParentAllowed
--- PASS: TestCheckFileAccess_NewFileUnderSymlinkedAllowedExternalDirAllowed
```

Mutant B — allowlist symmetry disabled:

```
--- FAIL: TestCheckFileAccess_NewFileUnderSymlinkedAllowedExternalDirAllowed (0.00s)
--- PASS: TestCheckFileAccess_NewFileWriteFallback
--- PASS: TestCheckFileAccess_NewFileUnderSymlinkedParentBlocked
--- PASS: TestCheckFileAccess_NewFileUnderInProjectSymlinkedParentAllowed
```

Each mutant reddens exactly the one test that asserts it, and no other. Neither test is
vacuous, and neither passes for an unrelated reason.

Residue check after restoring both: `grep -c '&& false' internal/hook/pre_tool.go` → `0`.

### E6 — package-level verification

```
go vet ./internal/hook/                       → exit 0
golangci-lint run ./internal/hook/...         → 0 issues
go test ./internal/hook/ -count=1             → exit 1, 279.092s
```

Total `--- FAIL` lines: 1.

```
--- FAIL: TestSessionStart_DeferredScanDoesNotBlockReturn (0.59s)
    session_start_parallel_test.go:97: Handle blocked 584.413667ms waiting for advisory scan; expected deferred (non-blocking) return
```

## Baseline-attribution

Every measurement above was taken in this run, in the worktree
`.claude/worktrees/t596`, against base commit `eabce74448e094dd1a4393044138a04b2016af99`,
with the working-tree edits to `internal/hook/pre_tool.go` and
`internal/hook/pre_tool_security_test.go` described in E3/E4. The RED measurement (E1)
was taken on that same tree with the test file added and `pre_tool.go` unmodified.

The one package failure is attributed as pre-existing on develop, on three grounds:
its name matches the pre-existing red the lead named as independently confirmed by three
lanes; its assertion is a wall-clock timing bound (584ms against a non-blocking
expectation) measured on a machine whose load average was 26-31 throughout; and the
change under test touches only `checkFileAccess` in the PreToolUse path, which the
SessionStart deferred-scan handler does not call. This attribution rests on those three
grounds and NOT on an independent re-measurement at the unmodified base commit — see Gaps.

## Gaps

- The pre-existing status of `TestSessionStart_DeferredScanDoesNotBlockReturn` was NOT
  re-measured on a clean checkout of `eabce7444` in this run. It is reasoned, not observed
  here.
- Cross-platform behavior is unobserved. `resolveThroughExistingParent` uses only
  `filepath` and `EvalSymlinks`, but Windows path semantics (drive roots, UNC paths) were
  not exercised. CI's windows matrix is the measuring instrument.
- The full test suite was not run locally (per CLAUDE.local.md §4/§6). Only
  `./internal/hook/` was measured; CI on the develop push is the full-suite verdict.
- No end-to-end test drives the real Write tool through the hook binary; the repair is
  verified at the `checkFileAccess` unit boundary only.

## Residual-risk

- `resolveThroughExistingParent` sets `resolvedSymlink=true` for many new-file paths that
  previously kept the unresolved form. Downstream that means `DenyPatterns` and
  `AskPatterns` now also match the resolved form for new files — a strengthening, but one
  that could newly deny a new file whose resolved path crosses a deny pattern the
  unresolved path did not. E3's control tests cover the common shapes; an unusual
  allowlist or deny-pattern configuration is not exhaustively covered.
- The allowlist now calls `EvalSymlinks` once per entry per check. The entry count is
  small (two built-ins plus user `additionalDirectories`) and the call is on the
  boundary-failure path only, but PreToolUse runs on every tool call under a 10s budget.
  Not measured under a pathological allowlist.
- The walk in `resolveThroughExistingParent` terminates at the filesystem root by
  `filepath.Dir` fixpoint. A path with an unusually deep non-existent tail costs one
  `EvalSymlinks` per level; unbounded in principle, bounded in practice by path length.
