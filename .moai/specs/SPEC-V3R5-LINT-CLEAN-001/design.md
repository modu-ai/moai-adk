# Design — SPEC-V3R5-LINT-CLEAN-001 — 5-Layer Safety Pipeline Compatibility

This document specifies the safety architecture for executing the four LCLN-Phase cleanup phases without risking constitutional violation, runtime regression, or Mega-Sprint W-series cross-SPEC entanglement. The structure mirrors the 5-Layer Safety Pipeline from `.claude/rules/moai/design/constitution.md` §5, adapted for agent-definition cleanup (rather than design-system evolution).

**Terminology**: Per `spec.md` §2.0 Glossary, "LCLN-Phase N" refers to this SPEC's internal cleanup phases (Phase 1..4); "Mega-Sprint Wave W0..W4" refers to v3.5.0 release-roadmap SPECs.

---

## Layer 1 — Frozen Guard (Pre-Edit Gate)

### 1.1 Guarded Paths (Superset of zone-registry Frozen List)

**Per plan-auditor iter-1 Finding 5**: This SPEC's guarded-path set is intentionally a **superset** of the files enumerated in `.claude/rules/moai/core/zone-registry.md` as `zone: Frozen`. The zone-registry currently labels exactly 7 files with `zone: Frozen` (CONST-V3R2 entries span 001..152 in total). This SPEC additionally guards operational invariants — files that are not formally `zone: Frozen` but whose modification would defeat the cleanup's safety goals.

The superset rationale is documented per-path in **Appendix A** at the bottom of this section.

**Effective guarded-paths glob** (this SPEC MUST NOT touch):

```
.claude/rules/moai/core/moai-constitution.md      # zone-registry Frozen + Agent Core Behaviors invariant
.claude/rules/moai/core/agent-common-protocol.md  # zone-registry Frozen + AskUserQuestion contract
.claude/rules/moai/core/askuser-protocol.md       # operational invariant (Appendix A item 3)
.claude/rules/moai/core/zone-registry.md          # operational invariant (Appendix A item 4)
.claude/rules/moai/design/constitution.md         # zone-registry Frozen (design FROZEN zone mirror)
.claude/rules/moai/development/coding-standards.md  # zone-registry Frozen + 16-lang neutrality
.claude/rules/moai/development/agent-authoring.md   # operational invariant (Appendix A item 7)
.claude/rules/moai/workflow/spec-workflow.md      # zone-registry Frozen + SPEC+EARS contract
.claude/rules/moai/workflow/mx-tag-protocol.md    # zone-registry Frozen + @MX TAG contract
internal/cli/agent_lint.go                        # operational invariant (Appendix A item 10) per REQ-LCLN-012
internal/cli/agent_lint_test.go                   # operational invariant (Appendix A item 11)
CLAUDE.md                                         # zone-registry Frozen (Claude Code substrate)
```

### 1.1 Appendix A — Per-Path Rationale

For each guarded path, this appendix records whether it is `zone-registry Frozen` or an `operational invariant` (this SPEC's superset), and why touching it would compromise the cleanup's safety goals.

| # | Path | Source | Rationale |
|---|------|--------|-----------|
| 1 | `.claude/rules/moai/core/moai-constitution.md` | **zone-registry Frozen** (CONST-V3R2-002, -028..035) | TRUST 5 framework + Agent Core Behaviors. Modification could weaken core orchestrator invariants that this SPEC's cleanup assumes. |
| 2 | `.claude/rules/moai/core/agent-common-protocol.md` | **zone-registry Frozen** (CONST-V3R2-036..046) | AskUserQuestion subagent prohibition + Skeptical-Evaluator Mandate. LR-01 and LR-07 rules depend on this protocol; the SPEC fixes agents to conform to it, not the protocol itself. |
| 3 | `.claude/rules/moai/core/askuser-protocol.md` | **operational invariant** | Canonical reference cited by LR-01 fix Pattern A (paraphrase + cross-ref). If this file's anchor structure changes during cleanup, Phase 3's edits would dangle. Therefore this SPEC pins it as guarded. |
| 4 | `.claude/rules/moai/core/zone-registry.md` | **operational invariant** | The registry itself is the source-of-truth for zone classification. Modifying it during cleanup would re-classify what is "FROZEN" mid-flight, undermining the SPEC's own scope discipline. |
| 5 | `.claude/rules/moai/design/constitution.md` | **zone-registry Frozen** (CONST-V3R2-051..072) | Design FROZEN zone mirror. Cleanup targets agent definitions, not design protocol; unrelated. |
| 6 | `.claude/rules/moai/development/coding-standards.md` | **zone-registry Frozen** (CONST-V3R2-004, -005) | 16-language neutrality + Template-First discipline. Both invariants are operative throughout the cleanup. |
| 7 | `.claude/rules/moai/development/agent-authoring.md` | **operational invariant** | Defines the canonical agent frontmatter schema this SPEC's edits must conform to. Modifying it would shift the target spec mid-cleanup. |
| 8 | `.claude/rules/moai/workflow/spec-workflow.md` | **zone-registry Frozen** (CONST-V3R2-001) | SPEC+EARS format contract. The SPEC itself depends on it; this cleanup does not modify the workflow it operates under. |
| 9 | `.claude/rules/moai/workflow/mx-tag-protocol.md` | **zone-registry Frozen** (CONST-V3R2-003) | @MX TAG protocol. Unrelated to agent-lint cleanup; included for completeness. |
| 10 | `internal/cli/agent_lint.go` | **operational invariant** (per REQ-LCLN-012) | The linter source. Cleanup fixes agents to conform to LR-XX rules; modifying the linter would defeat the purpose. Explicit REQ-LCLN-012 prohibition. |
| 11 | `internal/cli/agent_lint_test.go` | **operational invariant** | The linter's test surface. Modifying it would require either (a) testing weaker rules (rejected per REQ-LCLN-012), or (b) modifying the linter implementation (rejected). Cleanup remains agents-only. |
| 12 | `CLAUDE.md` | **zone-registry Frozen** (CONST-V3R2-007) | Claude Code substrate identity. Out of scope per spec.md §6 Exclusions. |

**zone-registry-classified Frozen**: 7 paths (items 1, 2, 5, 6, 8, 9, 12 above + zone-registry's small set that excludes some operational items).
**Operational invariant (this SPEC's superset)**: 5 paths (items 3, 4, 7, 10, 11).
**Total guarded paths**: 12.

The total in zone-registry CONST-V3R2 entry numbering spans **001..152** (not 001..072 as iter-1 design.md claimed); this SPEC explicitly does not depend on every CONST-V3R2 entry being relevant — it depends on the **path superset** being respected. The Frozen Guard script (§1.2) checks paths, not CONST-V3R2 IDs.

### 1.2 Pre-edit script (run before each Edit/Write tool invocation)

Per CLAUDE.local.md [HARD] Approach-First Development, before executing any LCLN-Phase's edit batch:

```bash
#!/bin/bash
# .moai/scripts/frozen-guard.sh — invoked at start of each LCLN-Phase's run-phase

GUARDED_PATHS=(
  ".claude/rules/moai/core/moai-constitution.md"
  ".claude/rules/moai/core/agent-common-protocol.md"
  ".claude/rules/moai/core/askuser-protocol.md"
  ".claude/rules/moai/core/zone-registry.md"
  ".claude/rules/moai/design/constitution.md"
  ".claude/rules/moai/development/coding-standards.md"
  ".claude/rules/moai/development/agent-authoring.md"
  ".claude/rules/moai/workflow/spec-workflow.md"
  ".claude/rules/moai/workflow/mx-tag-protocol.md"
  "internal/cli/agent_lint.go"
  "internal/cli/agent_lint_test.go"
  "CLAUDE.md"
)

CHANGED=$(git diff --name-only origin/main..HEAD)

for pattern in "${GUARDED_PATHS[@]}"; do
  if echo "$CHANGED" | grep -F "$pattern" > /dev/null; then
    echo "BLOCK: Guarded-path violation: $pattern was modified in this branch"
    exit 1
  fi
done

echo "OK: Frozen Guard passed — no guarded-path files modified"
exit 0
```

This script runs as a pre-merge gate in each LCLN-Phase PR (referenced from AC-LCLN-003.4). It does not need to be checked in to the repo for this SPEC — it can be inlined per AC-LCLN-003.4 test command — but the orchestrator may choose to commit it for reusability.

### 1.3 Edit-tool surface

All edits use the `Edit` tool with exact string matching. The Frozen Guard applies BEFORE any Edit is executed; once an LCLN-Phase begins editing, the targeted files are validated against the guarded-paths list and rejected if matching.

---

## Layer 2 — Canary Check (Representative Subset Validation)

Before applying any cleanup pattern to the full agent file set, the LCLN-Phase validates the pattern on a representative subset.

### 2.1 Canary subset definition

For each LCLN-Phase, the canary subset is a small (3-5 file) sample drawn from the phase's target finding distribution:

| LCLN-Phase | Canary subset | Rationale |
|------------|---------------|-----------|
| Phase 1 | `expert-backend.md`, `manager-spec.md`, `evaluator-active.md` (template + live) | Covers all four Phase 1 rules (LR-03, LR-12, LR-06, LR-05) without overlap |
| Phase 2 | `.claude/agents/moai/expert-mobile.md` (single file) | Single-file scope; full population is the canary |
| Phase 3 | `expert-backend.md`, `builder-harness.md`, `claude-code-guide.md` (template + live) | Covers LR-01 in both pattern A (paraphrase) and pattern B (subagent escalation) contexts, plus LR-02 |
| Phase 4 | `manager-spec.md`, `manager-strategy.md`, `expert-devops.md` (template + live) | Covers manager-category LR-08 + expert-category LR-08, including the highest-frequency drift skill (`moai-foundation-core`) |

### 2.2 Canary validation procedure

For each canary file:

```
1. Apply the LCLN-Phase's edit pattern.
2. Run `./bin/moai agent lint --strict --format=json` on the affected file only.
3. Verify the phase's target findings on that file are reduced (and no NEW findings introduced).
4. Run `go test ./internal/template/...` to verify embedded.go regeneration is clean.
5. If any canary fails, abort phase and re-design the edit pattern.
6. If all canaries pass, proceed to full population edit.
```

### 2.3 Canary failure handling

If a canary edit introduces a NEW finding (e.g., LR-08 drift triggered by adding a skill), the LCLN-Phase halts and the orchestrator reports the canary observation. The user (via AskUserQuestion if delegated, or via plan revision) decides whether to:

- Revise the edit pattern (e.g., add the skill to ALL manager agents, not just the outlier)
- Defer the specific finding to a separate sub-phase
- Add a new finding pattern that the phase will resolve in addition

The canary check is the cheapest place to catch pattern-design errors. It exists per `.claude/rules/moai/design/constitution.md` §5 Layer 2 Canary semantics — apply in memory, evaluate before commit.

---

## Layer 3 — Contradiction Detector (Cross-Rule Audit)

Each cleanup edit must not violate other LR-XX rules. The Contradiction Detector identifies potential cross-rule conflicts before commit.

### 3.1 Known cross-rule interactions

| Edit pattern | Potential cross-rule violation | Mitigation |
|--------------|-------------------------------|-----------|
| Add `effort:` field (LR-03 fix) | Could introduce LR-12 effort drift if wrong value chosen | Use canonical matrix from SPEC-V3R2-ORC-003 (documented in plan.md §4.1 T1.1 table) |
| Remove `--deepthink flag:` text (LR-06 fix) | Could break description's grammatical structure, triggering future rules | Edit preserves sentence structure; only the boilerplate substring is removed |
| Rewrite `AskUserQuestion` references (LR-01 fix) | Could remove cross-reference to askuser-protocol.md (CONST-V3R2-006..038 cite this protocol) | Pattern A explicitly adds cross-reference to `[askuser-protocol.md](.claude/rules/moai/core/askuser-protocol.md)` to preserve traceability |
| Remove `Agent` from `tools:` CSV (LR-02 fix) | Could leave agent unable to perform documented capability | Verify body does not actually invoke `Agent` token before removing; if it does, restructure via blocker-report pattern |
| Add `isolation: worktree` (LR-05 fix) | Could conflict with LR-09 (read-only agents must NOT have `isolation: worktree`) | Only `manager-develop.md` (write-heavy) needs this; verify before adding |
| Add skill to `skills:` array (LR-08 fix) | Could introduce LR-08 drift in another agent (asymmetry) | Add to ALL agents in the category, not just the outlier (preserves category symmetry) |

### 3.2 Cross-rule audit step (per LCLN-Phase) — strengthened per plan-auditor iter-1 Finding 12

After all LCLN-Phase edits complete, before opening the PR, run a **per-rule count regression check** (NOT the weaker "unique rule type count" check from iter-1):

```bash
PHASE=P1
PRE=.moai/state/lint-baseline-pre-LCLN-${PHASE}.json
POST=.moai/state/lint-baseline-post-LCLN-${PHASE}.json

# Strengthened per-rule check (Finding 12):
# Assert NO individual rule's count increased post-phase.
# This is stronger than the iter-1 "unique rule type count ≤ previous" invariant,
# because a rule's count can rise even if no NEW rule type appears.
jq -s '
  (.[0].violations | group_by(.rule) | map({key: .[0].rule, value: length}) | from_entries) as $pre |
  (.[1].violations | group_by(.rule) | map({key: .[0].rule, value: length}) | from_entries) as $post |
  ($pre + $post | keys) | map(
    {rule: ., pre: ($pre[.] // 0), post: ($post[.] // 0)}
  ) | map(select(.post > .pre))
' $PRE $POST > /tmp/per-rule-regressions-${PHASE}.json

REGRESSIONS=$(jq 'length' /tmp/per-rule-regressions-${PHASE}.json)

if [ "$REGRESSIONS" -ne 0 ]; then
  echo "BLOCK: ${REGRESSIONS} per-rule regression(s) in LCLN-${PHASE}:"
  jq '.[]' /tmp/per-rule-regressions-${PHASE}.json
  exit 1
fi
echo "OK: no per-rule count increases"
```

This implements AC-LCLN-003.5 (per Finding 12). The check is strictly stronger than the unique-rule-type check it replaces.

### 3.3 Implicit contradiction with FROZEN zone

If any LCLN-Phase edit attempts to fix a finding by modifying a guarded-path file (e.g., `agent-common-protocol.md` already paraphrases `AskUserQuestion` heavily; LR-01 findings on that file are spurious because it IS the orchestrator's protocol doc), the Frozen Guard (Layer 1) catches it. The Contradiction Detector serves as a second-line check at the edit-pattern level.

---

## Layer 4 — Rate Limiter (LCLN-Phase Velocity Control)

### 4.1 Per-day LCLN-Phase cap

Per CLAUDE.local.md §18.4 release cadence and §6 development workflow: this SPEC adopts **one LCLN-Phase per day maximum**. Each LCLN-Phase PR opens, runs CI (including all four lints), passes Frozen Guard audit, and merges before the next phase opens.

Rationale:
- Cooldown allows regression observation (any unintended runtime effect of metadata changes surfaces in normal usage between phases).
- Reviewer cognitive load remains bounded (each PR has clear scope).
- CI signal per phase preserved (cause-and-effect tracing).

### 4.2 Per-LCLN-Phase PR cap

Each LCLN-Phase is **exactly one PR**. Splitting a phase into multiple PRs is prohibited because:
- It complicates the delta verification (multiple pre/post baselines per phase)
- It creates pseudo-parallelism that the strategy explicitly rejects (plan.md §1)

If a phase's task list grows during run-phase (e.g., canary reveals additional patterns), the SPEC should be re-planned (status `draft` → revised) rather than splitting the phase.

### 4.3 Maximum LCLN-Phases per SPEC

This SPEC has exactly 4 LCLN-Phases (Phase 1, 2, 3, 4). Adding a 5th phase (e.g., for newly-discovered rule violations) requires SPEC version bump and plan revision per CLAUDE.local.md §6 development workflow.

### 4.4 Concurrency with Mega-Sprint W-series critical-path SPECs

LCLN-Phases 1, 3, 4 can proceed in parallel with Mega-Sprint W-series critical-path SPECs (W1 CONSTITUTION-DUAL, W3 HARNESS-AUTONOMY, W4 PROJECT-MEGA) because:
- File-scope orthogonality: this SPEC touches agent definitions only; Mega-Sprint W1/W3/W4 SPECs touch rule documents and runtime code
- Per AC-LCLN-005.1: rebase-onto-main before each LCLN-Phase PR catches any cross-PR conflict early

LCLN-Phase 2 must NOT proceed while Mega-Sprint W2 critical-path (SPEC-V3R5-CORE-SLIM-001) is in `run` phase because both might delete or modify `.claude/agents/moai/expert-mobile.md` simultaneously. Coordinate via:

- AC-LCLN-005.1 rebase step
- LCLN-Phase 2 PR description includes `Coordinated with: SPEC-V3R5-CORE-SLIM-001 (no conflict because LCLN-Phase 2 executes before Mega-Sprint W2 enters run-phase)`

---

## Layer 5 — Human Oversight (LCLN-Phase Approval Protocol)

### 5.1 Per-LCLN-Phase human approval — admin override scoping (Finding 10)

Each LCLN-Phase PR requires **explicit human merge approval** (no auto-merge without review).

**Admin override (per CLAUDE.local.md §18.7) is allowed in EXACTLY ONE scenario**:

- **LCLN-Phase 2 single-file deletion** (`git rm .claude/agents/moai/expert-mobile.md` only). The PR's diff MUST contain exactly one file deletion and zero non-deletion changes. Even `embedded.go` regeneration is NOT bundled in this PR. If the PR's diff fails this constraint, admin override is REJECTED and standard review applies.

All other admin override scenarios are explicitly DISALLOWED:

- LCLN-Phase 1 LR-12-only fast-track — REMOVED in iter-2 (iter-1 allowed this; the carve-out widened the back-door)
- LCLN-Phase 3 or 4 — never admin-mergeable; standard review required

The single allowed admin override is justified by:
- Diff size is exactly one line (file deletion)
- Live `expert-mobile.md` has no consumers (W0 retired the template; user projects do not have this agent)
- CI all-green confirms `moai update --dry-run` shows no registry regression (T2.1 post-condition)

Cross-reference: CLAUDE.local.md §18.7 sets `enforce_admins: false` on branch protection; this design.md explicitly scopes when that capability may be invoked, preventing the §18.7 back-door from widening over time.

### 5.2 Review checklist (per LCLN-Phase PR)

The PR description MUST include:

- [ ] Pre-phase baseline JSON committed (`.moai/state/lint-baseline-pre-LCLN-P<N>.json`)
- [ ] Post-phase baseline JSON committed (`.moai/state/lint-baseline-post-LCLN-P<N>.json`)
- [ ] Delta verification output: `NEW findings introduced by LCLN-P<N>: 0`
- [ ] Total reduction matches sum-check target: Phase 1 = 60, Phase 2 = 4, Phase 3 = 30, Phase 4 = 70
- [ ] Frozen Guard output: `OK: no guarded-path files modified`
- [ ] Per-rule regression check (AC-LCLN-003.5) output: `OK: no per-rule count increases`
- [ ] Orthogonal lint snapshot: `moai spec lint = 0, golangci-lint = 0, moai workflow lint = 0`
- [ ] `go test ./...` exit 0 confirmed in CI
- [ ] Template-first invariant (AC-LCLN-007.1 for Phases 1/3/4, AC-LCLN-007.2 for Phase 2)
- [ ] (Phase 2 only) `moai update --dry-run` confirmation (no non-mobile registry changes)
- [ ] (Phase 4 only) `.moai/state/lint-w2-deferred.json` committed; bound check `[11, 16]` passes

### 5.3 Escalation procedure

If an LCLN-Phase PR fails CI or review:

1. The orchestrator reports the failure as a "blocker report" (per agent-common-protocol §Blocker Report Format)
2. The user (via AskUserQuestion) decides whether to:
   - Re-attempt with revised edit pattern (cycle back to Canary)
   - Defer the failing finding to a future SPEC
   - Abort the phase and revise the SPEC
3. If a phase fails repeatedly (>2 attempts), the SPEC is marked status `blocked` and escalated to user review

### 5.4 Special case: LCLN-Phase 4 W2-deferred subset

LCLN-Phase 4 deliberately leaves ~12 findings unfixed (W2-deferred). The human reviewer MUST verify:

- The W2-deferred subset matches the documented pattern (skill or agent in Mega-Sprint W2-retirement list, per `research.md` §2.2)
- `.moai/state/lint-w2-deferred.json` is committed and includes a forward reference to SPEC-V3R5-CORE-SLIM-001
- The post-Phase-4 baseline JSON shows the residual is bounded (AC-LCLN-005.2: `[11, 16]`)

---

## Cross-Layer Integration

Each LCLN-Phase's run-phase loop:

```
1. [Layer 4] Verify rate limit (one phase/day, one PR/phase)
2. [Layer 1] Run Frozen Guard pre-edit script
3. [Layer 2] Apply edit pattern to canary subset
4. [Layer 3] Run Contradiction Detector on canary (per-rule check)
5. (If canary passes) Apply edit pattern to full population
6. [Layer 3] Run Contradiction Detector on full edits (per-rule check, AC-LCLN-003.5)
7. Run `make build` to regenerate embedded.go (Phases 1, 3, 4 only — Phase 2 is delete-only)
8. Capture post-phase baseline JSON
9. Run all four orthogonal lints (AC-LCLN-003.x)
10. Compute delta vs pre-phase baseline (AC-LCLN-002.x)
11. Run go test (AC-LCLN-004.x)
12. Verify template-first invariant (AC-LCLN-007.1 or 007.2)
13. Open PR with verification artifacts in description
14. [Layer 5] Wait for human review and approval (or admin override only for Phase 2 single-file delete)
15. Merge (squash) and observe main CI
16. Update memory with phase completion entry
```

This loop is delegated to `manager-develop` via `/moai run SPEC-V3R5-LINT-CLEAN-001` (per quality.yaml `development_mode`). The orchestrator coordinates AskUserQuestion at Layer 5 boundaries (decision points: edit pattern approval before Canary, escalation on canary failure, final phase approval).

---

## Operational Notes

- The 5-Layer Safety Pipeline as documented in `.claude/rules/moai/design/constitution.md` §5 was designed for the design-system evolution context. For this SPEC, Layers 1, 4, 5 apply verbatim; Layers 2, 3 are adapted from design-system semantics (Canary shadow-eval on last 3 projects → per-file canary subset; Contradiction Detector on rule files → per-rule LR-XX count regression).
- The cleanup does NOT touch the Sprint Contract Protocol (`.claude/rules/moai/design/constitution.md` §11) because no GAN Loop is involved; this is a non-iterative refactor.
- Memory hygiene: this SPEC will add one memory entry on completion (`project_v3r5_lint_clean_001_complete`), supersede none, and reference Mega-Sprint W0/W2 as cross-SPEC neighbors.
