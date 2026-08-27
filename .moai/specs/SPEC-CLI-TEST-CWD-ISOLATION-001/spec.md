---
id: SPEC-CLI-TEST-CWD-ISOLATION-001
title: "internal/cli tests must not leave a .moai/ in the repository tree — cwd state-write isolation"
version: "0.2.2"
status: in-progress
created: 2026-08-28
updated: 2026-08-28
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "cli, test-isolation, cwd, residue, state-dir, tempdir"
related_specs: [SPEC-CLI-STATE-DIR-BOUND-001, SPEC-AGENT-EMIT-LINEAGE-001]
---

# SPEC-CLI-TEST-CWD-ISOLATION-001

## §A Problem / Motivation

Running the `internal/cli` test suite materializes a `.moai/` directory inside the repository
tree at the package's own working directory (`internal/cli/.moai/state/…`). Go test binaries
run with `cwd` = the package directory, so every state write whose project root resolves to an
empty or relative value lands *inside the repository tree*. The residue is untracked and
gitignored, so `git status` never shows it — but every `.moai`-marker-based upward walk that
later starts below the repository root now stops at `internal/cli/.moai/` instead of reaching
the real project root. The misjudgment surfaces as a **different answer, not an error** — that
silence is the hazard.

### Observed evidence

**(1) Measured RED — reproducer verified on THIS worktree** (lead session, 2026-08-28, base
`origin/develop` `d34a789a4`, pre-fix tree, env scrubbed per the lane env-isolated
verification form; probe record: `.moai/reports/t334/red-probe.md`):

```console
$ rm -rf internal/cli/.moai
$ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR \
      MOAI_KANBAN_SETTINGS_INJECTED CLAUDE_PROJECT_DIR \
  && go test ./internal/cli -run 'Kanban|Factory' -count=1
  → rc=0, 0.869s
$ post-run: internal/cli/.moai/state/todo/leads.json        (EXISTS)
            internal/cli/.moai/state/factory/workers.json   (EXISTS)
```

This command is the **frozen reproducer** — the ACs in acceptance.md anchor on it verbatim. It
also explains this worktree's earlier residue (2026-08-28 00:04, same two files, five minutes
after tree creation): attribution established — it was this probe.

**(2) Primary checkout residue — re-verified 2026-08-28** (historical context; pre-t306 file
set): `state/config-cache.json` (Aug 13), `state/kanban/leads.json` (Aug 20),
`state/factory/workers.json` (Aug 20); untracked, gitignored via `.gitignore:280`
(`internal/cli/.moai/`). The `kanban/` → `todo/` directory rename landed with
SPEC-TODO-SQLITE-001 (t306), which is why the current write paths differ from this list.

**(3) Negative probes — measured on this base, same session**: `internal/config` package tests
produce NO residue; `internal/kanban` package tests produce NO residue;
`go test ./internal/cli -run 'Config|Cache|Settings' -count=1` produces NO residue. The
historical `config-cache.json` writer is therefore **not a producer on the current base** — it
stays in §A as context only and is excluded from the ACs. (`internal/hook` was NOT probed —
no claim is made about it.)

**(4) Originating observation — SPEC-AGENT-EMIT-LINEAGE-001 (card t317) §G-1**
(`git show 0ad4b52ba:.moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/progress.md`): the residue
**actually failed that SPEC's D9 gate** — a `.moai`-marker upward walk stopped at
`internal/cli/.moai/` and misjudged an applicable tree as inapplicable. t317 re-anchored its
own consumer on the committed emission set (`f3e5006ce`) but neither removed the residue nor
protected other marker-walk consumers. G-1 proposed this card as option (a): "isolate
`internal/cli` tests from writing state to the package cwd (generalize what the golden test
did for itself to the whole package)".

**(5) The mechanism is already documented in the tree.**
`internal/cli/doctor_golden_test.go:62-68`:

> "Without isolation, a prior test's config load writes `<cwd>/.moai/state/config-cache.json`
> (the cache's MkdirAll side effect) … t.Chdir + the cache-disable env keep the cwd empty for
> the run."

**(6) The base tree carries 4 COMMITTED `.moai` directories** (verified 2026-08-28 on
`d34a789a4`; `git ls-files` confirms tracked content in all four):
`internal/template/templates/.moai` and
`internal/harness/router/testdata/{normal,keyword-force,spec-overrides}/.moai`. Consequence:
any "tree contains no `.moai`" predicate is unsatisfiable on this base by construction — the
ACs therefore judge a **pre-run-baseline delta** (no NEW `.moai` entries), never emptiness
(audit iter1 D1; the instrument is defined in acceptance.md §D).

### Producer surface (measured + code-verified)

| # | Path | Status on this base | Mechanism |
|---|------|--------------------|-----------|
| P1 | `internal/config/cache.go:242` `writeCache` MkdirAll, via `ConfigManager.Load(projectRoot)` (`internal/config/manager.go:66-74`) | **NOT a producer** (probe (3)); historical only | empty/relative `projectRoot` → relative `.moai` configDir → `<cwd>/.moai/state/` |
| P2 | `internal/kanban/factory_slots.go:48,68` (`FactoryRegistryPath`, `SaveFactoryRegistry`) | **producer** (reproducer (1)) | `filepath.Join(root, ".moai", "state", "factory", "workers.json")` + `MkdirAll` — empty/relative `root` → cwd-relative |
| P3 | `internal/cli/kanban.go:330,368` + `internal/kanban/state_dir.go:47` (`leadRegistryPath`, `companionRegistryPath`, `StateDirForRoot`) | **producer** (reproducer (1): `todo/leads.json`) | same join shape — `<root>/.moai/state/todo` |

Which **tests** inside the `Kanban|Factory` selector drive P2/P3 with a cwd-relative root is
the one identification step left to M1 (the obvious registry unit-test files —
`kanban_lead_name_test.go`, `kanban_companion_name_test.go`, `factory_test.go` — all pass
`t.TempDir()` roots already; the producers are the cobra end-to-end paths). The fix is
expected to be confined to `internal/cli` test files, using the same medicine card t161
applied to 3 tests in SPEC-CLI-STATE-DIR-BOUND-001: explicit state-dir/root injection or a
temp cwd.

## §B Scope

**In Scope**:
- cwd write isolation of `internal/cli` tests: no `.moai/` directory created anywhere under the
  repository working tree as a side effect of running those tests (REQ-1, REQ-2, REQ-4).
- A durable residue guard inside the `internal/cli` test suite that fails the suite when the
  residue reappears (REQ-3).
- The fix mechanics: explicit state-dir/root injection, `t.Chdir(t.TempDir())`, `t.TempDir()`
  roots (REQ-2; ladder in plan.md §E D4).

**Out of Scope**: the items below.

### Out of Scope — consumer-side marker-walk hardening
- `findProjectRoot()` (`internal/cli/glm.go:1054-1058`), `findStateDir` consumers, hook-side
  walkers, and any other `.moai`-marker upward-walk consumers are NOT modified. Lane-9's
  "exhaustive consumer sweep" is an estimate from 2 observations, not a survey; adopting it
  requires first measuring the consumer count and filing a separate card (t317 G-1 option (b)).
  SPEC-CLI-STATE-DIR-BOUND-001 owns the consumer-side convention.

### Out of Scope — other packages' test residue
- `internal/config` and `internal/kanban` tests measured clean on this base (probe (3));
  `internal/hook` was **not probed** and no claim is made about it. No analogous fix is
  attempted for any of them — separate card if a future base regresses.

### Out of Scope — `.gitignore:280` entry removal
- The `internal/cli/.moai/` ignore line STAYS. Removing it would surface the primary checkout's
  stale residue as `git status` noise for every parallel session on this machine; regression
  detection is the guard test's job (REQ-3), not git's. Rationale in plan.md §E D2.

### Out of Scope — template mirrors
- Test-only change under `internal/`; no `internal/template/templates/` counterpart exists for
  the touched files. The Template-First mirror obligation does not apply.

### Out of Scope — CI residue-check workflow
- No `.github/workflows/` changes. The guard test rides the existing `go test` CI step
  automatically.

### Out of Scope — production state-resolution semantics
- P1-P3 production behavior (resolution order, relocation policy, fail-open posture) is
  unchanged. SPEC-CLI-STATE-DIR-BOUND-001 and SPEC-TODO-SQLITE-001 own those surfaces.

## §C Requirements (GEARS)

Acceptance criteria (Tier M) are enumerated canonically in **acceptance.md §D** — including
the frozen reproducer instrument and the baseline-delta tree scan (§A (6)). This section
carries the requirement layer only.

**Tier classification — M, with the S-shaped scope stated honestly** (for independent audit
judgment). The *measured implementation scope* alone is S-shaped: a test-only fix with no
production-code change, both producers pinned by measurement (P2/P3, §A (1)), a sub-second
frozen reproducer, and an expected diff of a handful of `internal/cli` test files plus one
guard. M stands for three reasons that outweigh that: (1) **the canonical artifact set is 3
files** — acceptance.md carries the AC matrix, the baseline-delta instrument (D1), and the
two-cell adoption discipline; folding it away to fit S deletes load-bearing verification
surface (audit iter1 D2: retain-over-delete). (2) **Audit-cycle state**: this SPEC has an
iter-1 FAIL on record; under the Tier S plan-audit retry ceiling (1) that FAIL reads as
loop-exhausting and forces user escalation, cutting off exactly the iter-2 revision slot the
SPEC is now using. (3) **M-weight verification ceremony**: a durable suite-level guard
(REQ-3) plus per-test isolation (REQ-4) are regression infrastructure, not a one-file patch.
Downgrade path, should the iter2 auditor judge the S-shape dominant: fold acceptance.md back
inline per the v0.2.0 pattern — in ONE atomic artifact write (the v0.2.0 fold executed across
two messages and audit iter1 read the torn intermediate; that failure mode is documented in
§G HISTORY) — and restate this block at S.

### REQ-1 — no `.moai/` creation under the repository tree during test runs

**While** the `internal/cli` test suite runs from the repository working tree, the suite's run
shall not create a `.moai/` directory anywhere under the repository working tree.

The judgment target is the residue itself — the directory's existence on disk — never a marker
proxy or a mock (design argument inherited from t317).

### REQ-2 — test-driven state writes land in the test-owned sandbox

**When** an `internal/cli` test exercises a code path that persists moai state files under a
project root (name-claim registries such as todo leads / factory workers, and any successor
state files), every such write the test causes shall land inside a test-owned temporary
directory, not inside the repository working tree.

### REQ-3 — durable residue guard fails the suite on regression

**Where** the `internal/cli` package test binary completes its run, a residue guard included in
the package's test suite shall fail the run when a `.moai/` directory exists at the package's
process working directory and did not exist when the run started.

The guard's locus is the package directory (the measured residue site, O(1) check). Tree-wide
detection beyond the package directory is the AC's baseline-delta scan, not the Go guard's —
the division is deliberate.

### REQ-4 — per-test isolation under individual selectors

**While** a test identified by plan.md M1 as residue-producing runs in isolation via its
individual `-run` selector, the repository working tree shall remain free of newly created
`.moai/` directories.

This closes the "passes only under whole-suite ordering" escape.

### REQ-5 — production default-path behavior is preserved

**Where** moai runs outside tests with no test-injected isolation variables set, production
state-resolution and state-write behavior shall remain identical to the pre-SPEC behavior.

## §D Options Considered

**Recommendation: Option 1 — per-test state-dir injection / temp cwd (the t161 medicine),
plus one residue guard.**

| Option | Content | Assessment |
|--------|---------|------------|
| **1 (recommended)** | Fix each M1-identified test: explicit absolute `t.TempDir()`/state-dir injection where the test controls the root; `t.Chdir(t.TempDir())` (+ `.moai` scaffold) where the code under test resolves from cwd. Plus one residue guard (REQ-3) | Implements t317 G-1 option (a); the same medicine SPEC-CLI-STATE-DIR-BOUND-001's card t161 applied to 3 tests (explicit `StateDir` injection); precedents in-tree (`doctor_golden_test.go:73-82`, `state_m2_test.go` `m2SetupState`); test-only diff expected — no production-code change |
| 2 | `TestMain`-global `os.Chdir(temp)` for the whole package | Rejected as primary: package-wide cwd change silently re-bases every test that relies on repo-relative fixtures; unmeasured blast radius. The guard, not a global chdir, is the TestMain-level intervention |
| 3 | Producer-side rejection of relative roots | Out of scope: changes production default-path behavior (REQ-5). Revisit as a separate SPEC only if the class recurs |
| 4 | Delete or skip the producing tests | Anti-pattern (plan.md AP-1); test deletion is not isolation |

## §E Constraints / Non-Goals / Residual Risks

### Constraints

- **RED-first (binding)**: no isolation fix lands before AC-001's RED (command + verbatim
  output + tree SHA) is re-established and recorded in `progress.md` §E.2 by the run-phase
  session. A test that never demonstrated RED is not accepted (t317 design argument;
  verification-completeness §1.1/§2).
- **Local verification scope**: affected packages only (`internal/cli`, expected sole touched
  package). `go test ./...` locally is FORBIDDEN (2026-08-15 machine-stall incident);
  full-suite judgment belongs to CI on the integration branch.
- **Suite runtimes (measured)**: the frozen reproducer runs in ~1 s (0.869 s observed); a FULL
  `internal/cli` package run measures ~336 s — full-package runs use a Bash timeout ≥ 600 s and
  run serially.
- **`t.TempDir()`** for all test temp directories (CLAUDE.local.md §6). `t.Setenv`/`t.Chdir`
  tests must not call `t.Parallel()`.
- **PRESERVE**: `internal/kanban/state_dir.go` resolution/relocation semantics;
  `internal/config/cache.go` default (env-unset) path; `.gitignore:280`; all currently-green
  test assertions.

### Residual Risks

- **R1** *(premise confirmed)*: the defect reproduces on this base (evidence (1)). Remaining
  contingency: if the run-phase re-verification unexpectedly fails to reproduce on a clean
  baseline, return a blocker report — do not force a fix.
- **R2** *(narrowed)*: the producing tests sit inside one selector (`Kanban|Factory`) but are
  not yet individually named; if the set turns out large, per-test fixes grow the diff. The
  mechanism ladder (plan.md §E D4) keeps each fix minimal; a producer seam remains last resort
  under REQ-5.
- **R3**: the guard checks the package-directory locus only; a hypothetical producer writing
  to an intermediate directory between package and root would be caught by the AC's
  baseline-delta scan, not the Go guard. No such producer observed (test cwd is always the
  package dir).
- **R4**: the stale residue in the primary checkout persists after this SPEC lands (different
  tree). Post-merge cleanup is lead-owned (plan.md M4 report).
- **R5**: residue file names drift (measured `kanban/` → `todo/`); AC-001 deliberately judges
  "≥1 file under `internal/cli/.moai`", never an exact name list.
- **R6**: the baseline-delta scan assumes the 4 committed `.moai` roots are stable during the
  run; a mid-run `git checkout`/merge on this worktree would shift the baseline and invalidate
  the diff (the serial, no-other-writer lane discipline covers this; re-record the baseline if
  the tree legitimately changes mid-task).

## §F Cross-References

- `internal/cli/doctor_golden_test.go:62-82` — self-isolation precedent and the documented
  residue mechanism quote.
- `internal/kanban/factory_slots.go:42-76` — P2: FactoryRegistryPath + SaveFactoryRegistry.
- `internal/cli/kanban.go:324-369` — P3: companionRegistryPath / leadRegistryPath.
- `internal/kanban/state_dir.go:36-48` — stateDirName `todo` (the measured kanban→todo drift).
- `internal/config/cache.go:57-59,71-81,229-247`, `internal/config/manager.go:62-74` — P1
  (historical, non-producer on this base): cache path, escape hatch, writeCache MkdirAll,
  relative-root join.
- `internal/cli/cc.go:14-17` — `findProjectRootFn` seam; `internal/cli/glm.go:1054-1058`
  `findProjectRoot` (out-of-scope consumer, context only).
- `internal/cli/main_test.go` — existing package TestMain (cobra warm-up guard); placement
  surface for the residue guard (plan.md §E D3).
- `internal/cli/state_m2_test.go:37` — `m2SetupState`: temp `.moai/state` scaffold + chdir
  precedent (cited by SPEC-CLI-STATE-DIR-BOUND-001 §A).
- SPEC-CLI-STATE-DIR-BOUND-001 — consumer-side convention + the t161 explicit-StateDir
  injection medicine this card re-applies on the producer-test side.
- `.gitignore:280` — `internal/cli/.moai/` ignore entry (kept; plan.md §E D2).
- t317 §G-1 (`git show 0ad4b52ba:.moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/progress.md`) —
  originating observation + the D9 gate misjudgment this residue already caused.
- `.moai/reports/t334/red-probe.md` — the plan-phase RED probe record (frozen reproducer
  source).
- `.moai/reports/t334/plan-audit-iter1.md` — audit iter1 (FAIL 0.775, D1-D7); this v0.2.1
  revision is the D1-D5/D7 response.
- `CLAUDE.local.md` §6 — t.TempDir() rule; §4 — affected-packages-only local verification.

## §G HISTORY

| Date | Author | Change |
|------|--------|-------|
| 2026-08-28 | manager-spec | v0.1.0 plan-phase authoring (card t334, initial Tier M). Evidence: t317 §G-1 (`0ad4b52ba`), primary-checkout residue re-verified, worktree residue 00:04 (then unattributed). |
| 2026-08-28 | manager-spec | v0.2.0 revision from the lead's measured evidence: RED reproducer verified on this worktree (`-run 'Kanban\|Factory'`, env-scrubbed, 0.869 s, produces `todo/leads.json` + `factory/workers.json`) — premise confirmed, 00:04 residue attributed to that probe; negative probes recorded (config-cache reclassified historical, excluded from ACs); ACs re-anchored on the frozen reproducer; attempted Tier S reclassification with acceptance.md folded inline (the fold executed across two messages and audit iter1 read the torn intermediate — see v0.2.1). |
| 2026-08-28 | manager-spec | v0.2.1 audit-response revision (plan-audit iter1 FAIL 0.775, D1-D5 + D7). **D2**: Tier restored to M — acceptance.md retained as the canonical AC surface (v0.2.0 fold reversed; auditor recommended retain-over-delete; Tier M keeps the iter-2 retry slot). **D1**: tree-clean scan redefined as a pre-run-baseline DELTA — 4 committed `.moai` dirs verified on base `d34a789a4` (templates + 3 router/testdata roots), so "scan empty" was unsatisfiable; instrument moved to acceptance.md §D + §A (6); `./internal/cli` path spelling normalized. **D3**: attribution aligned everywhere + `.moai/reports/t334/red-probe.md` cited. **D4**: `internal/hook` struck from the "measured clean" claim (never probed). **D5**: REQ-1 While-clause reworded to suite-level framing. **D7**: REQ-2 `(t.TempDir()-derived)` parenthetical dropped (mechanism lives in plan.md §E D4). REQ/AC budget unchanged: 5/5. |
| 2026-08-28 | manager-spec | v0.2.2 verification pass on the lead's round-2 re-delegation: both named defects were confirmed ALREADY FIXED in v0.2.1 (round-2 was written against the stale v0.2.0 view — mechanical proof: zero matches for "PASS = empty output" / "tree-clean scan empty" / "internal/hook tests measured clean" across spec/plan/acceptance; delta instrument present at acceptance.md §D, plan.md §C/D8/AP-11). New in this pass: §C Tier-classification rationale block added at the lead's request — M retained with the S-shaped measured scope stated explicitly and the downgrade path (single atomic fold) documented for the iter2 auditor's independent judgment. No REQ/AC changes. |
