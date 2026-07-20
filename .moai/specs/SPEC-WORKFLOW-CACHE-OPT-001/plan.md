# plan.md — SPEC-WORKFLOW-CACHE-OPT-001

> Tier **L**. Go code (new snapshot engine + `internal/goal` integration + CLI verb) + 19 template-mirrored doc files (17 workflow + 2 agent) + 5 axes. Shared findings: `./research.md`; M1 architecture: `./design.md`.
>
> **Tier rationale**: file_count = 19 doc surfaces + new Go package + `cmd/moai` registration + `internal/goal` change → multi-domain (Go source, workflow skills, agent files, template mirrors) with `thorough` harness per the Complexity Estimator (`file_count >= 10 AND domain_count >= 2 → thorough`). A tightly-scoped M was considered (M1-only Go + M2+ doc edits) but the gate-merging axis touches Kickoff-adjacent surfaces whose regression cost is high — Tier L's full A-E delegation template and thorough audit envelope are warranted. Tier L artifact set (5 files incl. design.md) satisfied per plan-audit iter-1 D9.

Milestones are ordered by **decision reversibility** — the decisions most likely to change (data model, evaluator integration, user-facing gate flows) lead; mechanical doc edits trail.

## §A — Context

- **Work location**: repo root `/Users/goos/MoAI/moai-adk-go`, branch `main` (Hybrid Trunk direct-push per git-workflow doctrine).
- **SPEC artifacts**: `.moai/specs/SPEC-WORKFLOW-CACHE-OPT-001/{spec,plan,acceptance,design,research,progress}.md` (Tier L 5-artifact set + progress skeleton).
- **Existing infrastructure (EXTEND, do not rebuild)**:
  - `.moai/state/verify/<session>/` — gitignored per-session evidence log dirs already in active use (agent-common-protocol § Evidence persistence obligation). The snapshot artifact joins this namespace.
  - `internal/goal/` — goal engine (schema.go / evaluate.go / state.go / prune.go). `evaluate.go` `CmdRunner` re-executes Tier-1 mechanical condition commands each turn-end; `state.go` already implements atomic per-session JSON persistence (reuse the pattern for the snapshot store).
  - loop-verdict JSON (`loop.md` § Remaining-Issue Persistence) — seed shape for the snapshot `conditions` block.
  - `.moai/cache/loop-snapshots/` — best-effort, no mechanical writer; REQ-SNAP-008 gives loop diagnostics a mechanical writer in the shared schema (the cache dir stays as-is for resume snapshots).
- **Target doc surfaces — 19-file edit-target inventory** (17 workflow + 2 agent; live-measured at plan time: 18 MIRROR-BYTE-EQ, `sync-auditor.md` sanitized-divergent by one line per D2/REQ-GUARD-004; every edit is a local+template pair):
  - Axis 1: `workflows/gate.md` (+ `--fresh` mode per REQ-SNAP-005/009), `workflows/run/task-decomposition.md` (Phase 2.75), `workflows/sync/quality-gates-context.md` (Phase 0), `workflows/loop.md` (Steps 1/3/1.5)
  - Axis 2: `workflows/moai.md` (Steps 11.3+11.5 + full-pipeline completion close), `workflows/sync/delivery.md` (Phase 4 "Completion and Next Steps" — REQ-GATE-004 second surface), `workflows/project/codebase-analysis.md` (Stage B), `workflows/harness-build-entry.md`, `workflows/feedback.md`
  - Axis 3: `workflows/fix.md`, `workflows/loop.md` (Step 6), `workflows/codemaps.md`, `workflows/clean.md` (incl. Phase 5.5 orchestrator-direct pin), `workflows/mx.md`
  - Axis 4: `workflows/run.md` + `workflows/run/phase-execution.md` (Tier S audit default), `workflows/project/doc-generation.md` (Phase 3.1 retry), auditor output contract surfaces (`.claude/agents/moai/plan-auditor.md`, `.claude/agents/moai/sync-auditor.md` — defect-list format; sanitized-parity class)
  - Axis 5: `workflows/loop.md` (Steps 5/6/7/7.5), `workflows/review.md` (Phase 2 + secrets scan)
  - Unique-file roll-up (19): root workflows `gate, moai, loop, fix, clean, codemaps, review, mx, feedback, harness-build-entry, run` (11) + sub-dir `run/task-decomposition, run/phase-execution, sync/quality-gates-context, sync/delivery, project/codebase-analysis, project/doc-generation` (6) + agents `plan-auditor, sync-auditor` (2). Cite-only (NOT edit targets): `SKILL.md`, `rules/moai/workflow/cache-aware-execution.md`.
- **Go surfaces**: new `internal/verify/` package (name provisional — see Settled Decisions), `cmd/moai` / `internal/cli` verb registration, `internal/goal/evaluate.go` snapshot-aware path + tests.

## §B — Known Issues (auto-injected, filtered to relevant categories)

- **B3 Subagent boundary**: `internal/goal` and the new snapshot package are hook-domain code — zero `AskUserQuestion` references (`grep -rn 'AskUserQuestion\|mcp__askuser' internal/verify internal/goal | grep -v _test.go | grep -v '^\s*//'` → 0 matches; keep/extend the CI boundary guard test).
- **B4 Frontmatter**: any SPEC-artifact edits use `created`/`updated`/`tags` canonical keys.
- **B6 spec-lint heading**: `## §D — Exclusions` alone is insufficient; the `### Out of Scope — <topic>` H3 sub-headings in spec.md are load-bearing for `OutOfScopeRule`.
- **B8 Working-tree hygiene**: do not commit `.moai/state/**` (gitignored; verify no force-add), `git add` specific paths only.
- **B9 Commit discipline**: per-milestone Conventional Commits `feat(SPEC-WORKFLOW-CACHE-OPT-001): M{N} <subject>`; never `--no-verify`.
- **B10 PRESERVE scope**: see §D Constraints.
- **Custom-1 — Advisory-check discipline (HARD, coding-standards.md)**: the `stop-goal` Stop hook runs each turn-end. The snapshot key computation (git subprocesses) on that path MUST be constant-cost and time-boxed (context deadline); on deadline exceed, degrade to command re-execution (never block the turn, never a linear-cost scan). This is the single highest-risk integration point.
- **Custom-2 — `TestAgentlessUtilityNoLLMControlFlow`**: fix.md Phase 3 carries an @MX:WARN that the Level→agent dispatch table must remain a static mapping. REQ-DELEG-001 rewrites `Level 1 → agent` to `Level 1 → orchestrator-direct formatter` — still a static mapping, but the run-phase MUST read this test first and keep it green (or update its expectation deliberately, never accidentally).
- **Custom-3 — run.md hard sentence conflict**: `run.md` Quick Reference states "Phase 0.5 (Plan Audit Gate): 모든 harness level에서 SKIP 불가". REQ-AUDIT-001 (Tier S single-pass default) does NOT skip the audit — it removes the iterative re-execution loop for Tier S. The edit must amend the wording precisely so "audit always runs once" is preserved and only the re-run loop is tiered. Sloppy wording here creates a doctrine contradiction the plan-auditor will flag.
- **Custom-4 — Step 1.5 independence (REQ-SNAP-009)**: the most tempting wrong optimization is letting the loop's success-exit consume its own snapshot. The carve-out sentence must land in loop.md Step 1.5 itself, not only in the snapshot doctrine.
- **Custom-5 — Sentinel preservation**: CI audits grep for literal sentinels in workflow bodies (`MODE_UNKNOWN`, `MODE_TEAM_UNAVAILABLE`, `MODE_PIPELINE_ONLY_UTILITY`, `MODE_FLAG_IGNORED_FOR_UTILITY`, `/moai run --mode loop` cross-reference text). Doc edits must not displace them.
- **Custom-6 — OTEL test env**: no `t.Setenv` with OTEL vars in new Go tests; use `t.TempDir()` for all snapshot-store tests.

## §C — Pre-flight (run-phase entry checks)

```bash
# 1. Baseline
git branch --show-current && git rev-parse HEAD
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=2m 2>&1 | tail -5

# 2. Template parity baseline for ALL 19 edit-target files
#    (expected: 18 EQ; agents/moai/sync-auditor.md DIFF by exactly the documented §25 sanitized line)
W=.claude/skills/moai/workflows; T=internal/template/templates/.claude
for f in gate.md moai.md loop.md fix.md clean.md codemaps.md review.md mx.md feedback.md harness-build-entry.md run.md \
         run/task-decomposition.md run/phase-execution.md sync/quality-gates-context.md sync/delivery.md \
         project/codebase-analysis.md project/doc-generation.md; do \
  cmp -s $W/$f $T/skills/moai/workflows/$f && echo "EQ $f" || echo "DIFF $f"; done
for f in plan-auditor.md sync-auditor.md; do \
  cmp -s .claude/agents/moai/$f $T/agents/moai/$f && echo "EQ agents/$f" || echo "DIFF agents/$f"; done
diff .claude/agents/moai/sync-auditor.md $T/agents/moai/sync-auditor.md   # expect exactly the (SPEC-V3R2-HRN-003)↔(HRN-003) line

# 3. Guard tests that constrain this SPEC's edits
grep -rn "TestAgentlessUtilityNoLLMControlFlow" internal/ --include="*.go" -l
go test ./internal/goal/... ./internal/template/...

# 4. stop-goal evaluator integration point
sed -n '1,60p' internal/goal/evaluate.go   # CmdRunner interface + Tier-1 execution path

# 5. Retired/superseded conflict scan on affected packages
grep -rn "Retired\|superseded" internal/goal | head -5 || echo "no conflicts"
```

## §D — Constraints (DO NOT VIOLATE)

- **PRESERVE (never modify)**: Implementation Kickoff Approval mandatory/score-independent wording (moai.md Pipeline Gates #2, run.md § Run-phase Autonomy #1, orchestration-mode-selection.md header); AskUserQuestion channel monopoly surfaces (`askuser-protocol.md` is cite-only); `verification-claim-integrity.md` (cite-only); `.claude/hooks/moai/sync-phase-quality-gate.sh` (out of scope); `cadence-bridge.md` (out of scope); loop-verdict base schema field names (additive compat only); Tier M/L audit iteration ceilings; `run.md § Run-phase Autonomy (/goal ac_converge)` section body (owned by SPEC-AUTONOMY-RUN-GOAL-001) except where a cross-reference line is added.
- **Forbidden**: `git add -A` / `git add .`; `--no-verify`; force-push; committing `.moai/state/**` or `.moai/cache/**`; editing archived agent files; removing CI sentinels.
- **Required**: Template-First pairing per milestone (local edit + `internal/template/templates/` mirror + `make build`); Conventional Commits with `🗿 MoAI` trailer; new Go package coverage ≥ 85%.

### Settled decisions

All clarification topics are resolved — zero `NEEDS CLARIFICATION` markers remain in plan.md/research.md.

- **Snapshot key composition** (user decision, plan-audit iter-1 D1 #1; strengthened additively per iter-2 D13): the key = HEAD commit SHA + digest of `git status --porcelain=v2` output + content hash of `git diff HEAD` output. Porcelain v2 covers the file-set shape (untracked non-ignored paths, staged/unstaged delta listing) but its output is byte-identical across successive edits to an already-dirty tracked file (experimentally refuted in iter-2 D13 — v2 carries HEAD/index object names, no worktree content hash); the `git diff HEAD` content hash is the leg that catches those re-edits. One extra constant-cost git subprocess; the Advisory-Check time-box (Custom-1) still binds. Remaining accepted limitation: an in-place content edit to an already-listed untracked file is outside all three inputs — recorded as Residual-risk in consumer reports, mitigated by the TTL bound below.
- **Freshness rule** (user decision, plan-audit iter-1 D1 #2): reuse requires key-equality AND wall-clock TTL — `recorded_at` within TTL, default **10 minutes**, configurable. Both conditions necessary; either failing ⇒ stale ⇒ re-execute.
- **stop-goal command matching** (user decision, plan-audit iter-1 D1 #3): exact byte-string match of the condition command in M1; on miss, fall back to the existing re-execution path unchanged. Normalized check-id matching is explicitly Out of Scope (spec.md §D; M2+ follow-up candidate).
- **Package name**: `internal/verify` (mirrors the `.moai/state/verify/` namespace). If a name collision emerges at run-phase, `internal/diagsnap` is the fallback — record the change in progress.md §E.2.
- **CLI surface**: `moai verify record` (stdin JSON or flag-driven single-check record) + `moai verify check --key-current [--check <id>]` (freshness query, exit 0 fresh / exit 1 stale). Registered in the root command tree — registration is a separately-pinned AC (cross-file reachability lesson).
- **Gate force-fresh mode** (plan-audit iter-1 D7): `/moai gate --fresh` disables ALL snapshot consumption for that invocation (fresh executions still recorded). loop.md Step 1.5 MUST invoke the gate with `--fresh` — this closes the gate-mediated self-consumption path (REQ-SNAP-005/009).
- **Snapshot file layout**: one JSON per key under `.moai/state/verify/snapshots/<key-prefix>.json`, atomic write via the `internal/goal/state.go` temp+rename pattern; single-checkout scope (multi-session sharing out of scope).
- **Tier S inversion semantics**: audit ALWAYS runs once for every tier; Tier S removes only the iterative re-execution loop (PASS final on first pass; FAIL/INCONCLUSIVE halts + escalates as today).
- **Agent-file parity model** (plan-audit iter-1 D2): workflow files = byte-parity; `.claude/agents/moai/*.md` = sanitized-parity (byte-equal after normalizing internal long-form SPEC-ID refs to sanitized short form; known instance: sync-auditor.md `(SPEC-V3R2-HRN-003)` ↔ `(HRN-003)`). Never copy internal SPEC IDs into the template (§25 neutrality CI guard).

## §E — Self-Verification (run-phase completion deliverables)

Per verification-claim-integrity §3 (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk), the run-phase completion report includes:

- **E1**: AC-WCO-001..036 binary PASS/FAIL matrix with per-AC verification command + verbatim output.
- **E2**: `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- **E3**: `go test -cover ./internal/verify/... ./internal/goal/...` ≥ 85% for the new package; goal package coverage non-regressing.
- **E4**: subagent-boundary grep (B3) → 0 matches.
- **E5**: `golangci-lint run` — NEW issues 0 (baseline distinguished).
- **E6**: per-milestone commit SHAs + push state.
- **E7**: template parity sweep after `make build` — all touched workflow files byte-equal (`cmp -s`); touched agent files sanitized-parity (diff confined to the documented §25 SPEC-ID sanitization lines; verbatim diff output cited).

## §F — Milestones (decision-reversibility order)

### M1 — Shared diagnostic snapshot contract (axis 1; Go + schema + consumers)

The only milestone with Go code; highest change-likelihood decisions (schema, key, freshness, evaluator integration) land here for earliest review.

1. `internal/verify/` package: schema (REQ-SNAP-001, loop-verdict-compatible `conditions` block), key computation (REQ-SNAP-002 — HEAD SHA + porcelain-v2 digest + `git diff HEAD` content hash per D13), freshness check (REQ-SNAP-003 — key-equality AND 10-min-default configurable TTL), atomic store. Table-driven tests incl. an E2E freshness test: record → same-tree in-TTL reuse PASS → (a) touch tracked file → stale, (b) RE-EDIT an already-dirty tracked file → stale (D13 boundary case), (c) add untracked file → stale, (d) advance clock past TTL → stale. Full architecture: design.md.
2. `moai verify` CLI verbs + root-tree registration + `--help` smoke.
3. `internal/goal/evaluate.go`: snapshot-aware Tier-1 path (REQ-SNAP-010 — exact byte-string command match; miss ⇒ existing re-execution), time-boxed key computation with re-execution fallback (Custom-1), verdict payload attribution field; tests with fake runner proving (a) exact-match fresh-hit reuse (call-count 0), (b) stale/miss re-execution, (c) deadline fallback.
4. Doctrine injection: gate.md (REQ-SNAP-005 produce+consume + `--fresh` force-fresh mode per D7), run/task-decomposition.md Phase 2.75 (REQ-SNAP-006), sync/quality-gates-context.md Phase 0 (REQ-SNAP-007), loop.md Steps 1/3 (REQ-SNAP-008) + Step 1.5 independence carve-out incl. the `gate --fresh` invocation (REQ-SNAP-009) + VCI attribution wording (REQ-SNAP-011).
5. Template mirror + `make build` + full test suite.

### M2 — Gate merging (axis 2; user-facing flow decisions)

1. moai.md: merge Step 11.3 Kickoff question + Step 11.5 execution-shape question into one AskUserQuestion call (REQ-GATE-001), preserving the Pipeline Gates #2 mandatory/score-independent text verbatim (REQ-GUARD-001).
2. project/codebase-analysis.md Stage B: one multi-question call for remaining axes (REQ-GATE-002); Stage B always-runs semantics untouched.
3. harness-build-entry.md: proposal + approval single round (REQ-GATE-003).
4. feedback.md: single 3-question round (REQ-GATE-005).
5. Full-pipeline completion close on BOTH surfaces (REQ-GATE-004): moai.md completion step AND `workflows/sync/delivery.md` Phase 4 "Completion and Next Steps" — no manufactured next-step question on full-pipeline success; single-phase "(Recommended)" chain retained on both.
6. Template mirror + build.

### M3 — Audit improvement (axis 4; protocol-shape decisions)

1. run.md Quick Reference + run/phase-execution.md: Tier S single-pass default with precise wording (Custom-3; REQ-AUDIT-001).
2. plan-auditor.md + sync-auditor.md output contract: structured defect-list on FAIL (REQ-AUDIT-002); orchestrator delta re-check flow documented in the owning workflow bodies (`workflows/run/phase-execution.md` for plan-audit, `workflows/moai.md` Pipeline Gates for sync-audit); verdict authority sentence (REQ-AUDIT-004). Agent-file edits follow sanitized-parity (Settled Decisions).
3. project/doc-generation.md Phase 3.1: retry cap 3 → 1 (REQ-AUDIT-003).
4. Template mirror + build.

### M4 — Delegation relaxation (axis 3; mechanical doc edits)

1. fix.md Phase 3 + Execution Summary: Level-1 orchestrator-direct formatter path; delegation mandate re-scoped to Level 2+ (REQ-DELEG-001); keep static dispatch table shape (Custom-2, REQ-DELEG-006).
2. loop.md Step 6: same exception (REQ-DELEG-002).
3. codemaps.md: 3-spawn → ≤1-spawn restructure (REQ-DELEG-003).
4. clean.md: 4-spawn → 2-spawn restructure (REQ-DELEG-004); Phase 5.5 pinned orchestrator-direct (specialist alternative removed — worst case stays ≤2); approval + @MX:ANCHOR safety text preserved.
5. mx.md: <5-item orchestrator-direct Pass 3 (REQ-DELEG-005).
6. Template mirror + build.

### M5 — Bookkeeping batching (axis 5; mechanical doc edits)

1. loop.md Steps 5/6/7: per-iteration single bookkeeping batch (REQ-BOOK-001); Step 7.5 → aggregate MX_TAG_REPORT at exit (REQ-BOOK-002).
2. review.md Phase 2: Mode-4 parallel 4-lens fan-out, sync-auditor synthesis retained (REQ-BOOK-003).
3. review.md secrets scan: incremental last-SHA checkpoint under `.moai/state/` + full-scan fallback/flag (REQ-BOOK-004).
4. Template mirror + build.

### M6 — Verification sweep (closure)

1. Full parity sweep (E7), `make build`, `go test ./...`, `golangci-lint`, `moai spec lint` clean on this SPEC.
2. E1 AC matrix executed end-to-end; progress.md §E.2/§E.3 populated by manager-develop.

## §G — Anti-Patterns (named, run-phase MUST avoid)

- **Self-referential success-exit**: loop Step 1.5 consuming the same run's snapshot (violates REQ-SNAP-009 — the exact failure mode the independent pass exists to prevent).
- **Gate deletion disguised as merging**: removing the Kickoff question or making it conditional on plan-auditor score. Merging = one AskUserQuestion call carrying both questions; nothing else.
- **Partial-freshness reuse**: accepting a snapshot on partial key match ("HEAD same, tree probably same") or past the TTL ("only 12 minutes old") — the freshness predicate is binary on BOTH legs (key equality AND in-TTL), per REQ-SNAP-003.
- **Gate-mediated self-consumption**: loop Step 1.5 invoking `/moai gate` WITHOUT `--fresh`, letting the same-run snapshot flow back through the gate layer (D7; violates REQ-SNAP-009 transitively).
- **Snapshot-as-permission**: citing a snapshot to skip a HUMAN GATE or hook. The snapshot replaces re-execution of a check, never an approval.
- **Vacuous-grep ACs**: token-presence checks without reachability (CLI verb text present but unregistered; doctrine names an engine that doesn't build). Every consumer AC pairs doctrine text with the mechanical surface it invokes.
- **Kitchen-sink commits**: `git add -A` absorbing parallel-session artifacts.

## §H — Cross-References

- design.md (this dir) — M1 snapshot-contract architecture (schema, key, freshness, store, consumer wiring, force-fresh mechanism).
- research.md (this dir) — duplication inventory with file:line anchors, contradiction ledger, existing-infra survey.
- `.claude/rules/moai/workflow/cache-aware-execution.md` — Phase 1 doctrine (motivation; cite-only).
- `.claude/rules/moai/core/verification-claim-integrity.md` §2/§3 — attribution rules REQ-SNAP-011 binds to.
- `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution / § Evidence persistence — `.moai/state/verify/` namespace precedent.
- `internal/goal/` — evaluator integration point; `state.go` atomic-write pattern to reuse.
- `.claude/rules/moai/development/coding-standards.md` § Advisory-Check Discipline — Stop-hook cost constraint (Custom-1).
