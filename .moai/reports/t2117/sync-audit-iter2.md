auditor-model: glm-5.3
verdict: PASS
audited_sha: d9442b09f3f9716d7de2b661fe2e266cc13f4c6a

# t2117 sync-audit iter2 — repair round 2 (F3+ / F7)

Card t2117 · branch `WT-main-c1-instructions` · audited tree = worktree
`/Users/goos/.moai/worktrees/projects/moai-adk-go/t2117` · repair delta =
`git diff 16e4bc90ff..d9442b09f3` (2 files, +117/−42:
`internal/config/token_budget_guard.go` + its test). Repair commits share one
commit (`d9442b09f3`); the RED claims were re-established by this auditor on
an extracted pre-repair tree (below). Operator ruling d-20261106T134133Z-aa98
mandated both repairs. iter1 (FAIL on F3+) is untouched at
`sync-audit-iter1.md`.

Overall Verdict: **PASS** — the single blocking finding F3+ and the optional
F7 are both **fixed-and-verified** (RED reproduced on the pre-repair tree,
GREEN at the repair HEAD, mechanism reviewed line-by-line); the repair-born
blocking count is 0; codex's repair-delta re-audit returned `pass` with zero
findings (receipt below).

## Dimension Scores (harmonic 93.5)

| Dimension | Score | Verdict | Evidence (verbatim) |
|-----------|-------|---------|---------------------|
| Functionality (40%) | 96 | PASS | Repair HEAD, attributed env (`GOTOOLCHAIN=go1.27.1 GOCACHE=/tmp/t2117-gocache`): `ok github.com/modu-ai/moai-adk/internal/config 1.019s` with `go test -list 'TestAlwaysLoaded'` sweeping 6 tests including both new ones (`TestAlwaysLoadedSurfaceRefusesSymlinkEscapeImport`, `TestAlwaysLoadedSurfaceCountsBothLocalSlots`). Regression groups unchanged-code: `ok .../internal/hook 1.548s`, `ok .../internal/cli 1.934s`. RED on pre-repair tree (`git archive 16e4bc90ff` + the two new tests as a separate file): `--- FAIL: TestAlwaysLoadedSurfaceRefusesSymlinkEscapeImport ... surface counts .../linked.md — a symlink whose target escapes repoRoot` and `--- FAIL: TestAlwaysLoadedSurfaceCountsBothLocalSlots ... omits .../CLAUDE.local.md ... omits .../legacy-import.md`. Tripwire stable: `always-loaded surface = 65156 tokens (budget 77600, headroom 12444, 16 entries)` (this tree carries only AGENTS.local.md; the repair adds enumeration only when both slots coexist). The iter1 stated-criterion failure ("escaping refused") now demonstrably holds on REAL paths. |
| Security (25%) | 92 | PASS | F3+ fix mechanism verified in source: `real, rerr := filepath.EvalSymlinks(resolved)` → dangling skip → `filepath.Rel(realRoot, real)` escape rejection, with `realRoot` itself normalized via `filepath.EvalSymlinks(repoRoot)` (macOS `/var`→`/private/var` handled — judged children against an unresolved root would reject everything). Symlinked-ancestor escapes resolve through EvalSymlinks and are rejected; the behavior test pins target-AND-spelling absence from the surface. Hook side unchanged and unaffected (Lstat refuses symlinks). Residual: N1 TOCTOU between EvalSymlinks and Stat/ReadFile — same accepted-subset class as iter1 F1/F2. |
| Craft (20%) | 94 | PASS | RED-first re-established mechanically (auditor-run RED above; the repair's RED observations were working-tree, since tests+fix share commit `d9442b09f3` — the reproduction closes that gap). Both new tests are behavior pins, not count-gaming: the escape test asserts absence of BOTH the link spelling and the outside target; the slots test asserts both bodies and both import closures. Rename complete: `grep -rn "localInstructionSlot\b" internal/ --include="*.go"` → no matches (exit 1). Enumeration re-derivation derives `wantTotal` from `len(localInstructionSlots(root))` — mildly circular for the count alone, bounded by the closure pin and the dedicated behavior tests. |
| Consistency (15%) | 92 | PASS | `go vet ./internal/config/ ./internal/hook/ ./internal/cli/` → `VET-EXIT=0`; `gofmt -l` on the three packages → empty list (both via background task, exit 0). Plural-rename call sites all updated (`alwaysLoadedSurface` extras, `instructionImportMembers` seeds, tests). Comments updated with F3+/F7 grounding (AGENTS.md §8 cited in-source); MX notes carry `[AUTO]` + card ref. |

Must-pass firewall: Functionality and Security both independently above
threshold. All new findings below are **optional** — none blocks.

## Dispositions of iter1 findings

| id | disposition | evidence |
|----|-------------|----------|
| F3+ (was BLOCKING) | **fixed-and-verified** | Real-path boundary judgment (EvalSymlinks on both target and root; dangling skip; in-root spelling preserved for surface paths). RED reproduced pre-repair; GREEN at HEAD; codex repair-delta re-audit `pass`, zero findings. |
| F7 (Medium/optional) | **fixed-and-verified** (folded into this repair per the operator ruling) | `localInstructionSlots` enumerates ALL present regular-file slots; closure walk seeds from all of them; `alwaysLoadedSurface` unions them before dedup. RED reproduced pre-repair; GREEN at HEAD. |
| F1, F2, F4, F5, F6 | unchanged, optional | Not in this repair's scope (hook/CLI subsets and cosmetics); recorded in iter1. |

Repair stop rule: blocking findings born inside the repair diff since the
iter1 audited SHA: **0** (codex zero findings; this auditor's review of the
repair found only note-level items below).

## New findings (all optional)

| id | loc | severity | disposition | file:line | description |
|----|-----|----------|-------------|-----------|-------------|
| N1 | in-diff | Low | optional | internal/config/token_budget_guard.go:405-412 | TOCTOU remains between `EvalSymlinks(resolved)` and the later `os.Stat`/`os.ReadFile` on the enqueued spelling: an attacker swapping the link in that window still reaches the old judgment. Same accepted-subset class as iter1 F1/F2 (measurement guard; hook prompt surface unaffected). |
| N2 | in-diff | Info | optional | internal/config/token_budget_guard.go:415-418 | Dedup is by in-root spelling, not resolved identity: two distinct in-root spellings resolving to one real file both enter the surface and double-count its tokens. Conservative direction for a budget guard (early trip, never a miss); note only. |
| N3 | in-diff | Info | optional | internal/config/token_budget_guard_test.go:316-321 | `TestAlwaysLoadedSurfaceEnumeration` derives `wantTotal` via the production `localInstructionSlots` — circular for the count alone; bounded by the closure pin (no unexpected non-rule members) and the dedicated behavior tests that pin the function itself. |

## Codex repair re-audit (final allowance spent)

- Call: `mcp__moai__codex_audit`, `target="baseBranch"`,
  `delta_base=16e4bc90ff`, `project_root=<own toplevel>` — the LAST call of
  the de2d cap-2 allowance (iteration-1's call + this one; the four
  `stale_server` errors and the iter1 receipt
  `rcpt-570cb4d72e7e4f0aaab856cb` are recorded in iter1).
- Result: **verdict `pass`, zero findings**; receipt
  **`rcpt-660e3171b7294d5d3462c03a`**; server provenance
  `"backend":"codex","resolved_model":"gpt-6.1-sol"`,
  `"build_commit":"04d8c27a7e"`; review scope
  `"review_base":"delta_base (merge base 16e4bc90ff2b61c1cb43fe0063d754336c25d08d)"`.
  The result carries no explicit head-SHA field; the reviewed delta is
  merge-base `16e4bc90ff…` to call-time HEAD
  `d9442b09f3f9716d7de2b661fe2e266cc13f4c6a` (verified unchanged immediately
  before the call). Codex notes it reviewed statically without running
  tests; the mechanical evidence above is this auditor's own runs.

## Verification Record (5-section)

**Claim** — (1) F3+ fixed: the import boundary judges REAL paths
(EvalSymlinks-resolved target against an equally-resolved root), rejects
target-or-ancestor escape, skips dangling links, preserves in-root spellings
for the surface. (2) F7 fixed: all present local instruction slots are
enumerated and seed the import closure. (3) No regression in the three
packages' targeted groups. (4) The two RED claims hold on the pre-repair
tree. (5) No new blocking defect in the repair delta.

**Evidence** — HEAD `d9442b09f3…`, attributed env:
```
ok  	github.com/modu-ai/moai-adk/internal/config	1.019s   (-run 'TestAlwaysLoaded' -count=1; list-swept 6 tests incl. both new)
ok  	github.com/modu-ai/moai-adk/internal/hook	1.548s     (-run 'Instruction' -count=1)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.934s      (-run 'TestMigrateLocalInstructions' -count=1)
token_budget_guard_test.go:70: always-loaded surface = 65156 tokens (budget 77600, headroom 12444, 16 entries)
VET-EXIT=0 / GOFMT-EXIT=0 (empty list)   (background task, exit 0)
grep -rn "localInstructionSlot\b" internal/ → no matches (exit 1)
```
RED tree (`git archive 16e4bc90ff` to /tmp/t2117-red2 + the two new tests appended as a separate file so the tree compiles):
```
--- FAIL: TestAlwaysLoadedSurfaceRefusesSymlinkEscapeImport (0.01s)
    zz_t2117_red2_test.go:42: surface counts .../linked.md — a symlink whose target escapes repoRoot; the boundary check must resolve links before judging the escape
--- FAIL: TestAlwaysLoadedSurfaceCountsBothLocalSlots (0.00s)
    zz_t2117_red2_test.go:76: always-loaded surface omits .../CLAUDE.local.md — both local slots and both import closures must be counted while the pair coexists
    zz_t2117_red2_test.go:76: always-loaded surface omits .../legacy-import.md — both local slots and both import closures must be counted while the pair coexists
```

**Baseline-attribution** — every GREEN figure: command + verbatim output
observed this run, this tree, HEAD `d9442b09f3f9716d7de2b661fe2e266cc13f4c6a`,
attributed environment `GOTOOLCHAIN=go1.27.1 GOCACHE=/tmp/t2117-gocache`.
Every RED figure: observed this run on the extracted tree of pre-repair
commit `16e4bc90ff` with only the two new test functions added (auditor-
authored file, verbatim from the repair's tests). Codex receipt minted by
build 04d8c27a7e against merge-base `16e4bc90ff…`.

**Gaps** — (a) Full suite not run (lane discipline; CI owns the repository
verdict). (b) The repair's own RED observations were working-tree (tests and
fix share one commit) — closed by this auditor's independent pre-repair
reproduction, but the commit graph alone cannot witness that ordering
(verification-claim-integrity §2.3 applies to future repairs: land the RED
commit separately). (c) Codex reviewed statically (no test/build execution,
per its instructions) — mechanical evidence here is this auditor's runs.
(d) Environment: the concurrent cache damage persisted (a vet+gofmt batch
was backgrounded past 180s before completing exit 0); no environmental
failure recorded as a card defect. (e) Symlink test skips gracefully where
`os.Symlink` is unavailable (unprivileged Windows); not compiled on Windows
here. (f) Coverage not re-measured for the repair subset (iter1's
subset-figure caveat carries over).

**Residual-risk** — (1) N1's TOCTOU window (post-EvalSymlinks link swap)
stays open — measurement-only impact. (2) N2's duplicate-spelling
over-count is possible in symlink-heavy trees. (3) The hook-side BFS keeps
the Lstat discipline (symlinks refused outright) while the config side now
follows in-root symlinks — the two surfaces still enumerate symlinked
members differently (iter1 F3 reduced to: config follows, hook refuses;
both bounded, noted for a future unification). (4) Codex's static review
did not execute the new tests; the GREEN evidence rests on this auditor's
runs alone.
