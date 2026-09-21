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

### E7 — resumed-session same-load A/B (2026-09-12 18:37–18:42)

The lead's 18:06 handoff WITHDREW the pre-existing-red attribution this verdict originally
relied on: the three lanes' "confirmed develop red" observations of
`TestSessionStart_DeferredScanDoesNotBlockReturn` were all taken under high load, lane-4
measured red at load 14.30 and green at 14.08 (load is not monotonic), and the failure
message is a 500ms budget overrun, not a deferral failure. That withdrawal removes ground 1
of the attribution below. This section replaces the reasoning with same-load observation.

All six measurements were taken in the resumed run, in this worktree, load 19.4–26.4
throughout (one load window), each an env-scrubbed single invocation:

| # | Tree | Selector set | Result |
|---|------|--------------|--------|
| 1 | HEAD 9233d3213 | 4 new tests + WriteFallback + SessionStart (-v) | 4 new PASS; SessionStart FAIL 572.99ms |
| 2 | HEAD 9233d3213 | SessionStart only | PASS (ok 1.099s) |
| 3 | HEAD 9233d3213 | same 5-selector battery | SessionStart FAIL 0.52s (exit 1) |
| 4 | base eabce7444 | SessionStart only | PASS 0.44s (exit 0) |
| 5 | base eabce7444 | WriteFallback + SessionStart (base-equivalent battery) | PASS 0.976s (exit 0) |
| 6 | HEAD 9233d3213 | WriteFallback + SessionStart (base-equivalent battery) | PASS 1.005s (exit 0) |

Rows 4–5 were measured with `internal/hook/pre_tool.go` and
`internal/hook/pre_tool_security_test.go` checked out at `eabce7444`; restoration after
each was verified by `git status --porcelain -- internal/hook/` printing nothing.

Reading:

- **The source change is not the trigger.** With the base-equivalent selector set the
  changed tree passes (row 6) exactly as base does (row 5); with a single selector both
  pass (rows 2, 4). The E6 attribution to a product-code regression does not survive.
- **The trigger correlates with the added tests' composition, not their assertions.**
  Every `TestCheckFileAccess_*` is `t.Parallel()`; the SessionStart test is sequential. Go
  runs each parallel test's synchronous setup (t.TempDir + symlink creation) before the
  sequential test, so the 3 added tests queue extra filesystem-setup work immediately
  ahead of the 500ms budget measurement. Including them: 3/3 runs FAIL (rows 1, 3 and the
  E6 full-package run). Excluding them: 4/4 PASS (rows 2, 5, 6).
- This is consistent with lane-4's finding and does not contradict it: the flake exists
  without t596 (red at load 14.30); the added parallel setups lower the failure threshold
  under load (base passed at 0.44s; the failures landed at 0.52–0.58s against a 500ms
  budget).

Attribution status after E7: the E6 package-level FAIL is **not** attributed to the t596
source change (rows 5→6). It is facilitated by the card's added parallel-test setups under
high load — a test-composition effect on an already-marginal 500ms budget assertion, whose
discrimination and repair belong to t662.

## Baseline-attribution

Every measurement above was taken in this run, in the worktree
`.claude/worktrees/t596`, against base commit `eabce74448e094dd1a4393044138a04b2016af99`,
with the working-tree edits to `internal/hook/pre_tool.go` and
`internal/hook/pre_tool_security_test.go` described in E3/E4. The RED measurement (E1)
was taken on that same tree with the test file added and `pre_tool.go` unmodified.

The one package failure was originally attributed as pre-existing on develop, on three
grounds. **Ground 1 (name match with the lead-confirmed pre-existing red) is WITHDRAWN** —
the lead's 18:06 handoff retracted that confirmation as a high-load artifact. Ground 3
(no call path from the SessionStart deferred-scan handler to `checkFileAccess`) remains a
reasoned argument. The attribution now rests on E7's direct same-load A/B (rows 5→6),
which observes the failure absent with the base-equivalent selector set on the changed
tree; E7 supersedes the withdrawn ground wherever the two disagree.

## Gaps

- The base tree was measured single-selector and 2-selector only (E7 rows 4–5). Base under
  the full added-test storm is structurally unmeasurable — the 3 added tests do not exist
  at base — so the composition effect's size at base cannot be quantified.
- All E7 measurements share one load window (19.4–26.4) on one machine. The
  load-independence of the composition effect is not established; t662 owns the
  discrimination.
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
- CI note for the integrator: the develop push's full-suite run may show
  `internal/hook` red on `TestSessionStart_DeferredScanDoesNotBlockReturn` (E6, E7). Per
  E7 that would be the known DeferredScan budget flake (t662) marginally facilitated by
  this card's added parallel-test setups — not a regression in the source change. CI's
  load profile differs from this machine's, so the failure may simply not occur there.
