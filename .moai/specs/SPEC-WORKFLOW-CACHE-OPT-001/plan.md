# plan.md — SPEC-WORKFLOW-CACHE-OPT-001

> Tier **L**. Go code (new snapshot engine + `internal/goal` integration + CLI verb) + 16 template-mirrored workflow/doctrine files + 5 axes. Shared findings: `./research.md`.
>
> **Tier rationale**: file_count ≥ 16 doc surfaces + new Go package + `cmd/moai` registration + `internal/goal` change → multi-domain (Go source, workflow skills, rules, template mirrors) with `thorough` harness per the Complexity Estimator (`file_count >= 10 AND domain_count >= 2 → thorough`). A tightly-scoped M was considered (M1-only Go + M2+ doc edits) but the gate-merging axis touches Kickoff-adjacent surfaces whose regression cost is high — Tier L's full A-E delegation template and thorough audit envelope are warranted.

Milestones are ordered by **decision reversibility** — the decisions most likely to change (data model, evaluator integration, user-facing gate flows) lead; mechanical doc edits trail.

## §A — Context

- **Work location**: repo root `/Users/goos/MoAI/moai-adk-go`, branch `main` (Hybrid Trunk direct-push per git-workflow doctrine).
- **SPEC artifacts**: `.moai/specs/SPEC-WORKFLOW-CACHE-OPT-001/{spec,plan,acceptance,progress,research}.md`.
- **Existing infrastructure (EXTEND, do not rebuild)**:
  - `.moai/state/verify/<session>/` — gitignored per-session evidence log dirs already in active use (agent-common-protocol § Evidence persistence obligation). The snapshot artifact joins this namespace.
  - `internal/goal/` — goal engine (schema.go / evaluate.go / state.go / prune.go). `evaluate.go` `CmdRunner` re-executes Tier-1 mechanical condition commands each turn-end; `state.go` already implements atomic per-session JSON persistence (reuse the pattern for the snapshot store).
  - loop-verdict JSON (`loop.md` § Remaining-Issue Persistence) — seed shape for the snapshot `conditions` block.
  - `.moai/cache/loop-snapshots/` — best-effort, no mechanical writer; REQ-SNAP-008 gives loop diagnostics a mechanical writer in the shared schema (the cache dir stays as-is for resume snapshots).
- **Target doc surfaces** (all verified MIRROR-BYTE-EQ against `internal/template/templates/` at plan time — every edit is a local+template pair):
  - Axis 1: `workflows/gate.md`, `workflows/run/task-decomposition.md` (Phase 2.75), `workflows/sync/quality-gates-context.md` (Phase 0), `workflows/loop.md` (Steps 1/3/1.5)
  - Axis 2: `workflows/moai.md` (Steps 11.3+11.5), `workflows/project/codebase-analysis.md` (Stage B), `workflows/harness-build-entry.md`, `workflows/feedback.md`
  - Axis 3: `workflows/fix.md`, `workflows/loop.md` (Step 6), `workflows/codemaps.md`, `workflows/clean.md`, `workflows/mx.md`
  - Axis 4: `workflows/run.md` + `workflows/run/phase-execution.md` (Tier S audit default), `workflows/project/doc-generation.md` (Phase 3.1 retry), auditor output contract surfaces (`.claude/agents/moai/plan-auditor.md`, `.claude/agents/moai/sync-auditor.md` — defect-list format)
  - Axis 5: `workflows/loop.md` (Steps 5/6/7/7.5), `workflows/review.md` (Phase 2 + secrets scan)
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

# 2. Template parity baseline for the 16 target files (must be byte-equal BEFORE edits)
for f in gate.md moai.md loop.md fix.md clean.md codemaps.md review.md mx.md feedback.md harness-build-entry.md; do \
  cmp -s .claude/skills/moai/workflows/$f internal/template/templates/.claude/skills/moai/workflows/$f && echo "EQ $f" || echo "DIFF $f"; done

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

### Open decisions — resolve before Implementation Kickoff Approval

1. **[NEEDS CLARIFICATION: snapshot key — untracked-file participation]** Should untracked (non-ignored) files participate in the working-tree content hash? They affect test outcomes (a new untracked `_test.go` changes `go test` results), but hashing untracked content adds cost and edge cases (large untracked dirs). Options: (a) include untracked non-ignored paths+mtimes via `git status --porcelain=v2` digest (cheap, catches presence/rename, misses in-place content edits of untracked files), (b) full content hash including untracked (safest, slowest), (c) tracked-only (fastest, known false-fresh window). Recommendation: (a).
2. **[NEEDS CLARIFICATION: per-layer freshness TTL]** Is key-equality alone sufficient (same tree ⇒ reuse regardless of age), or should a wall-clock TTL additionally bound reuse (flaky-test staleness, environment drift)? Options: (a) no TTL — pure key equality, (b) global TTL (e.g., 30 min), (c) per-layer TTL table (gate lenient / sync Phase 0 strict). The chosen rule becomes the REQ-SNAP-003 "layer's freshness-acceptance rule".
3. **[NEEDS CLARIFICATION: stop-goal command-matching granularity]** REQ-SNAP-010 matches a Tier-1 condition to a snapshot entry — by exact command-string equality, or via a canonical check-id mapping (e.g., normalize `go test ./...` variants)? Exact-match is safe but low-hit-rate; normalization raises hit rate but risks false matches. Recommendation: exact-match in M1, normalization as follow-up.

### Settled decisions

- **Package name**: `internal/verify` (mirrors the `.moai/state/verify/` namespace). If a name collision emerges at run-phase, `internal/diagsnap` is the fallback — record the change in progress.md §E.2.
- **CLI surface**: `moai verify record` (stdin JSON or flag-driven single-check record) + `moai verify check --key-current [--check <id>]` (freshness query, exit 0 fresh / exit 1 stale). Registered in the root command tree — registration is a separately-pinned AC (cross-file reachability lesson).
- **Snapshot file layout**: one JSON per key under `.moai/state/verify/snapshots/<key-prefix>.json`, atomic write via the `internal/goal/state.go` temp+rename pattern; single-checkout scope (multi-session sharing out of scope).
- **Tier S inversion semantics**: audit ALWAYS runs once for every tier; Tier S removes only the iterative re-execution loop (PASS final on first pass; FAIL/INCONCLUSIVE halts + escalates as today).

## §E — Self-Verification (run-phase completion deliverables)

Per verification-claim-integrity §3 (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk), the run-phase completion report includes:

- **E1**: AC-WCO-001..036 binary PASS/FAIL matrix with per-AC verification command + verbatim output.
- **E2**: `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- **E3**: `go test -cover ./internal/verify/... ./internal/goal/...` ≥ 85% for the new package; goal package coverage non-regressing.
- **E4**: subagent-boundary grep (B3) → 0 matches.
- **E5**: `golangci-lint run` — NEW issues 0 (baseline distinguished).
- **E6**: per-milestone commit SHAs + push state.
- **E7**: template parity sweep — all touched files byte-equal local vs template after `make build`.

## §F — Milestones (decision-reversibility order)

### M1 — Shared diagnostic snapshot contract (axis 1; Go + schema + consumers)

The only milestone with Go code; highest change-likelihood decisions (schema, key, freshness, evaluator integration) land here for earliest review.

1. `internal/verify/` package: schema (REQ-SNAP-001, loop-verdict-compatible `conditions` block), key computation (REQ-SNAP-002 + clarification #1 resolution), freshness check (REQ-SNAP-003 + clarification #2), atomic store. Table-driven tests incl. an E2E freshness test: record → same-tree reuse PASS → touch tracked file → stale detected.
2. `moai verify` CLI verbs + root-tree registration + `--help` smoke.
3. `internal/goal/evaluate.go`: snapshot-aware Tier-1 path (REQ-SNAP-010, clarification #3), time-boxed key computation with re-execution fallback (Custom-1), verdict payload attribution field; tests with fake runner proving (a) fresh-hit reuse, (b) stale re-execution, (c) deadline fallback.
4. Doctrine injection: gate.md (REQ-SNAP-005 produce+consume), run/task-decomposition.md Phase 2.75 (REQ-SNAP-006), sync/quality-gates-context.md Phase 0 (REQ-SNAP-007), loop.md Steps 1/3 (REQ-SNAP-008) + Step 1.5 independence carve-out sentence (REQ-SNAP-009) + VCI attribution wording (REQ-SNAP-011).
5. Template mirror + `make build` + full test suite.

### M2 — Gate merging (axis 2; user-facing flow decisions)

1. moai.md: merge Step 11.3 Kickoff question + Step 11.5 execution-shape question into one AskUserQuestion call (REQ-GATE-001), preserving the Pipeline Gates #2 mandatory/score-independent text verbatim (REQ-GUARD-001).
2. project/codebase-analysis.md Stage B: one multi-question call for remaining axes (REQ-GATE-002); Stage B always-runs semantics untouched.
3. harness-build-entry.md: proposal + approval single round (REQ-GATE-003).
4. feedback.md: single 3-question round (REQ-GATE-005).
5. moai.md + sync delivery surface: full-pipeline success closes with no manufactured next-step question (REQ-GATE-004); single-phase "(Recommended)" chain retained.
6. Template mirror + build.

### M3 — Audit improvement (axis 4; protocol-shape decisions)

1. run.md Quick Reference + run/phase-execution.md: Tier S single-pass default with precise wording (Custom-3; REQ-AUDIT-001).
2. plan-auditor.md + sync-auditor.md output contract: structured defect-list on FAIL (REQ-AUDIT-002); orchestrator delta re-check flow documented in the owning workflow bodies; verdict authority sentence (REQ-AUDIT-004).
3. project/doc-generation.md Phase 3.1: retry cap 3 → 1 (REQ-AUDIT-003).
4. Template mirror + build.

### M4 — Delegation relaxation (axis 3; mechanical doc edits)

1. fix.md Phase 3 + Execution Summary: Level-1 orchestrator-direct formatter path; delegation mandate re-scoped to Level 2+ (REQ-DELEG-001); keep static dispatch table shape (Custom-2, REQ-DELEG-006).
2. loop.md Step 6: same exception (REQ-DELEG-002).
3. codemaps.md: 3-spawn → ≤1-spawn restructure (REQ-DELEG-003).
4. clean.md: 4-spawn → 2-spawn restructure (REQ-DELEG-004); approval + @MX:ANCHOR safety text preserved.
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
- **TTL-less blind reuse across trees**: accepting a snapshot on partial key match ("HEAD same, tree probably same") — key equality is binary.
- **Snapshot-as-permission**: citing a snapshot to skip a HUMAN GATE or hook. The snapshot replaces re-execution of a check, never an approval.
- **Vacuous-grep ACs**: token-presence checks without reachability (CLI verb text present but unregistered; doctrine names an engine that doesn't build). Every consumer AC pairs doctrine text with the mechanical surface it invokes.
- **Kitchen-sink commits**: `git add -A` absorbing parallel-session artifacts.

## §H — Cross-References

- research.md (this dir) — duplication inventory with file:line anchors, contradiction ledger, existing-infra survey.
- `.claude/rules/moai/workflow/cache-aware-execution.md` — Phase 1 doctrine (motivation; cite-only).
- `.claude/rules/moai/core/verification-claim-integrity.md` §2/§3 — attribution rules REQ-SNAP-011 binds to.
- `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution / § Evidence persistence — `.moai/state/verify/` namespace precedent.
- `internal/goal/` — evaluator integration point; `state.go` atomic-write pattern to reuse.
- `.claude/rules/moai/development/coding-standards.md` § Advisory-Check Discipline — Stop-hook cost constraint (Custom-1).
