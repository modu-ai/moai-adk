# Plan — SPEC-CLI-TEST-CWD-ISOLATION-001

## §A Context

- **Worktree**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t334`, branch `WT-cli-test-cwd`,
  base = `origin/develop` `d34a789a4` (verified at plan-phase start; re-read before any commit).
- **Card**: t334. Route: git-flow lane protocol — no card PR; integration into local `develop`
  via the integration worktree; CI verdict on `origin/develop`.
- **Artifacts**: spec.md + plan.md + acceptance.md — **Tier M (3 artifacts)** + `progress.md`.
  `related_specs`: SPEC-CLI-STATE-DIR-BOUND-001, SPEC-AGENT-EMIT-LINEAGE-001.
- **Development mode**: `tdd` (quality.yaml) — RED-GREEN-REFACTOR; this SPEC's RED-first
  constraint is the same discipline at AC granularity.
- **Measured RED (already observed at plan phase, spec.md §A (1); probe record
  `.moai/reports/t334/red-probe.md`)**: env-scrubbed `go test ./internal/cli -run
  'Kanban|Factory' -count=1` → rc=0, 0.869 s, residue `state/todo/leads.json` +
  `state/factory/workers.json` recreated. The run-phase session re-establishes this with its
  own attribution triple before any fix commit.
- **Expected fix shape**: confined to `internal/cli` test files — explicit state-dir/root
  injection or temp cwd, the same medicine card t161 applied to 3 tests in
  SPEC-CLI-STATE-DIR-BOUND-001 — plus one residue guard.
- **Existing infrastructure to EXTEND**: `internal/cli/main_test.go` TestMain (guard placement),
  `doctor_golden_test.go` isolation idiom, `state_m2_test.go` `m2SetupState` scaffold idiom.
- **PRESERVE targets**: `internal/kanban/state_dir.go` semantics, `internal/config/cache.go`
  default path, `.gitignore:280`, all currently-green test assertions.
- **Pre-existing worktree residue**: `internal/cli/.moai/state/{factory/workers.json,
  todo/leads.json}` — attributed to the lead's plan-phase reproducer probe (2026-08-28 00:04;
  red-probe.md). Untracked + gitignored. Pre-flight records then removes it — it is the
  baseline contaminant for every RED/GREEN judgment.
- **Base-tree fact (audit iter1 D1)**: 4 COMMITTED `.moai` dirs exist on `d34a789a4`
  (`internal/template/templates/.moai` + 3 under `internal/harness/router/testdata/`). All
  tree judgments are **baseline deltas**, never emptiness checks (D8).

## §B Known Issues

- **B1 — two runtimes (measured)**: the frozen reproducer runs in ~1 s (0.869 s observed) —
  iterate with it. A FULL `internal/cli` package run measures ~336 s; full-package runs (M4 /
  AC-004) MUST set a Bash timeout ≥ 600 s and run serially.
- **B2 — `t.Setenv`/`t.Chdir` vs `t.Parallel`**: tests that gain either must drop
  `t.Parallel()`; audit the touched tests for parallel markers before editing.
- **B3 — chdir is process-wide for the test binary's duration of that test**: a chdir'd test
  must not leak assumptions to subsequent tests; where the code under test resolves the root
  from cwd and the test must tolerate resolution failure, scaffold a temp `.moai`
  (`m2SetupState` idiom).
- **B4 — residue is invisible to git**: `.gitignore:280` hides it; NEVER use `git status` as
  the detection command — use `find` / `test -e` / the baseline delta.
- **B5 — file-set drift**: residue names moved `kanban/` → `todo/` (SPEC-TODO-SQLITE-001);
  never assert an exact file list — AC-001 judges "≥1 file under `internal/cli/.moai`".
- **B6 — test caching invalidates evidence**: every evidence-bearing run uses `-count=1`.
- **B7 — cross-platform**: the guard and any path logic must be `filepath`-based and pass
  `GOOS=windows GOARCH=amd64 go build ./...`.
- **B8 — scope discipline on a shared machine**: stage by explicit pathspec only; never
  `git add -A`; re-read `git status --short` immediately before staging.
- **B9 — env scrub is part of the frozen command**: the `unset MOAI_KANBAN* …
  CLAUDE_PROJECT_DIR` prefix travels with the reproducer verbatim; a judgment run without it
  is a different command (D7).

## §C Pre-Flight (run these before any change; record outputs)

```bash
git rev-parse --short HEAD && git branch --show-current
# expect: d34a789a4 (or descendant) / WT-cli-test-cwd

# Record + remove the plan-phase probe residue:
ls -laR internal/cli/.moai/ && rm -rf internal/cli/.moai

# Record the tree baseline (D1/D8 — this is the GREEN judgment's reference):
find . -mindepth 2 -name .moai -type d -not -path './.git/*' | sort > /tmp/t334-moai-pre.txt
cat /tmp/t334-moai-pre.txt
# expect EXACTLY these 4 committed dirs (tracked; verified plan-phase + audit iter1):
#   ./internal/harness/router/testdata/keyword-force/.moai
#   ./internal/harness/router/testdata/normal/.moai
#   ./internal/harness/router/testdata/spec-overrides/.moai
#   ./internal/template/templates/.moai
# (if ./internal/cli/.moai appears, the removal above did not take — stop and fix)

# Re-establish RED with the frozen reproducer (verbatim, incl. env scrub):
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR \
      MOAI_KANBAN_SETTINGS_INJECTED CLAUDE_PROJECT_DIR \
  && go test ./internal/cli -run 'Kanban|Factory' -count=1
# expect: rc=0; find internal/cli/.moai -type f → todo/leads.json + factory/workers.json
# record command + verbatim output + SHA in progress.md §E.2, then rm -rf internal/cli/.moai
# and confirm the scan equals the recorded 4-line baseline again

go build ./... && GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=2m 2>&1 | tail -5
# record lint baseline (distinguish NEW vs pre-existing findings later)
```

## §D Constraints (Hard)

1. **RED-first**: AC-001 re-established and recorded in `progress.md` §E.2 (command + verbatim
   output + HEAD SHA) before any isolation-fix commit. No RED → blocker report, not a fix.
2. **Local scope**: affected packages only; `go test ./...` locally FORBIDDEN; full-suite
   verdict = CI on the integration branch.
3. **`t.TempDir()`** for all new temp dirs (CLAUDE.local.md §6).
4. **No consumer-side changes**: `findProjectRoot`/`findStateDir`/hook walkers untouched.
5. **`.gitignore:280` stays** (§E D2).
6. **No template changes; no CHANGELOG** (sync-phase owns it).
7. **English code comments**; env names via `internal/config/envkeys.go` constants.
8. **Conventional commits** with card id `t334` in every message; 🗿 MoAI trailer per repo
   convention.
9. **Producer seams are last resort** and MUST be env-guarded, default-off (REQ-5). The
   expected diff contains zero non-test changes.

## §E Self-Verification (design decisions)

- **D1 — guard anchors on the judgment target**. The guard observes the residue directory's
  existence delta across the run (absent at TestMain entry, present after `m.Run()` → fail,
  failure message names the detected path). No marker proxy, no mock — the binding design
  argument inherited from t317 (commit `f3e5006ce`).
- **D2 — keep `.gitignore:280`**. Removing the entry would surface the primary checkout's
  stale residue (a different tree this card cannot clean) as `git status` noise for every
  parallel session. Detection duty moves to the guard test, which is louder and CI-enforced.
- **D3 — guard placement**. Extend the existing `TestMain` in `internal/cli/main_test.go`
  (record `<cwd>/.moai` absence before `m.Run()`, assert absence after) or add a dedicated
  TestMain-adjacent file if separation reads cleaner — either way it MUST NOT disturb the cobra
  warm-up guard already living there, and MUST NOT be `t.Parallel`. The guard runs in every
  CI OS matrix for free.
- **D4 — isolation mechanism ladder** (cheapest first, per producing test):
  1. explicit absolute `t.TempDir()` root / state-dir injection where the test controls the
     root parameter (the t161 medicine);
  2. `t.Chdir(t.TempDir())` (+ minimal `.moai` scaffold via the `m2SetupState` idiom) where
     the code under test resolves the root from cwd — mind that `findProjectRootFn()` from an
     empty temp dir errors, which is itself behavior the test must tolerate or scaffold around;
  3. `t.Setenv(config.EnvConfigCacheDisabled, "1")` — historical lever for the cache producer;
     NOT needed on this base (config-cache measured non-producer) and kept only as fallback;
  4. env-guarded producer seam, default-off — only if 1-2 demonstrably cannot isolate.
- **D5 — requirements target the class, not a file list**. Measured drift `kanban/` → `todo/`
  inside one week.
- **D6 — attribution closed at plan phase**: the worktree 00:04 residue was the lead's
  reproducer probe (recorded at `.moai/reports/t334/red-probe.md`). The run-phase session
  still owns its own attribution triple: plan-phase evidence anchors the requirement;
  run-phase evidence anchors the fix.
- **D7 — selector freeze**: the frozen reproducer (command + env scrub + selector) is
  evidence. M1's per-test refinement ADDS child selectors under `Kanban|Factory`; changing
  the frozen command itself re-establishes RED first.
- **D8 — tree judgments are baseline deltas (audit iter1 D1)**. The base carries 4 committed
  `.moai` roots; every tree-wide judgment records the pre-run scan (expected: exactly those
  4) and asserts post-run == pre-run. Emptiness predicates are forbidden — they are
  unsatisfiable on this base. If the tree legitimately changes mid-task (rebase, checkout),
  re-record the baseline before the next judgment.

## §F Milestones (M1 gates the rest; expected small diffs throughout)

### M1 — Re-establish RED + per-test identification — Priority High
1. Pre-flight (§C): record + remove the probe residue; record the 4-line baseline; re-run the
   frozen reproducer; record the attribution triple in `progress.md` §E.2; remove the residue
   again and confirm the scan equals the baseline.
2. Bisect the selector: run each test matched by `Kanban|Factory` individually
   (`go test ./internal/cli -run '<TestName>' -count=1`, ~1 s each) to name the producing
   tests and the root-resolution each hits (empty/relative root vs cwd-derived).
3. Freeze the per-test selector list alongside the frozen reproducer (D7).
4. Contingency: if the reproducer unexpectedly fails to reproduce on a clean baseline, return
   a blocker report — do not force a fix.

### M2 — Residue guard (durable RED) — Priority High
1. Add the guard per D3 (TestMain-anchored, package-cwd locus, English comments).
2. Observe guard RED: `go test ./internal/cli -run 'Kanban|Factory|<guard>' -count=1` (same
   env scrub) on the unfixed tree → run FAILS naming the detected path; record verbatim.

### M3 — Isolation fixes (GREEN) — Priority High
1. Apply the D4 ladder per producing test; minimal diff; drop `t.Parallel()` where
   `t.Setenv`/`t.Chdir` enter (B2).
2. Re-run the frozen reproducer → rc=0, `internal/cli/.moai` absent, post-run scan diffs
   empty vs the recorded baseline (AC-003); guard green in the M2 pairing; per-test selectors
   each leave the delta clean (REQ-4).
3. Record GREEN evidence in `progress.md` §E.2 (same command shape as RED).

### M4 — Package-wide verification + wrap-up — Priority Medium
1. `go test ./internal/cli -count=1` (timeout ≥ 600 s, serial) → exit 0; post-run scan diffs
   empty vs the baseline (AC-004).
2. Diff enumeration vs base: list every touched non-`*_test.go` file with rationale — expected
   empty; touched-package tests green (AC-005).
3. `go vet ./internal/cli/...` + lint delta vs baseline.
4. Populate `progress.md` §E.2/§E.3; report to lead with the PRIMARY-checkout cleanup
   recommendation (stale `internal/cli/.moai/state/` there).

## §G Anti-Patterns

- **AP-1**: deleting, skipping, or `t.Skip`-ing the producing tests. Test deletion is not
  isolation.
- **AP-2**: asserting a fixed residue file list (drifts; B5).
- **AP-3**: detecting residue via `git status` (gitignored → invisible; B4).
- **AP-4**: a fix that passes only under whole-suite ordering (REQ-4 exists for this).
- **AP-5**: changing producer default-path behavior to make tests pass (REQ-5 violation).
- **AP-6**: local `go test ./...` "to be thorough" (machine-stall incident).
- **AP-7**: a guard that checks a proxy/mock instead of the directory's existence (D1).
- **AP-8**: evidence from cached test results — no `-count=1` (B6).
- **AP-9**: running the reproducer without the env-scrub prefix and citing it as the frozen
  command (B9/D7) — it is a different command.
- **AP-10**: writing the fix first and "confirming" RED afterward with the fix present
  (RED-first violation; the RED tree must be the unfixed tree).
- **AP-11**: asserting scan EMPTINESS as a GREEN condition — unsatisfiable on this base
  (4 committed `.moai` roots; D8). Always diff against the recorded baseline.

## §H Cross-References

- `spec.md` §A evidence (1)-(6) + producer table, §C REQ-1..5, §E risks R1-R6.
- `acceptance.md` §D — canonical AC matrix, frozen reproducer + baseline-delta instrument,
  §D.2 adoption discipline.
- `.claude/rules/moai/development/verification-completeness.md` §1.1/§2/§4 — observed-failure
  completion, two-cell adoption (incl. the "impossible direction" D1 embodied), evidence
  pinning.
- `.claude/rules/local/gitflow-lane-protocol.md` — integration + push discipline; §8 the
  env-isolated verification form the frozen reproducer follows.
- `.moai/reports/t334/plan-audit-iter1.md` — audit iter1 findings this plan's D8/AP-11 answer.
- `CLAUDE.local.md` §4/§6 — affected-packages-only verification, t.TempDir() rule.
