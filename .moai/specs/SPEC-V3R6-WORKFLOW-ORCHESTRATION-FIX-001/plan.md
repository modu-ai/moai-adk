---
id: SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001
artifact: plan
version: "0.1.0"
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
plan_commit_sha: "<pending>"
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-25 | manager-spec | Initial Tier L plan-phase artifact authored from research.md §G suggested split; 6 milestones (M1-M6); 22 in-scope files + 10 mirror parity = ~32 files total scope |

---

## §A — Lifecycle Table

| Step | Phase | Owner | Status | Commit SHA | Notes |
|------|-------|-------|--------|------------|-------|
| 0 | Pre-flight (research.md authored) | orchestrator | DONE | `592b752e1` | 3 parallel research agents — Anthropic + Context7 + GitHub |
| 0 | Plan-phase M0 (4 SPEC artifacts) | manager-spec | IN-PROGRESS | `<pending>` | spec.md + plan.md + acceptance.md + design.md |
| 0.5 | plan-auditor (Tier L PASS ≥ 0.85, skip ≥ 0.90) | plan-auditor | NOT-STARTED | n/a | Single-pass + max-3 retry contract; design.md mandatory at this tier |
| 0.95 | Mode Selection (NEW) | orchestrator | NOT-STARTED | n/a | 5-mode autonomous selection per REQ-WOF-004; expected `parallel` mode for Tier L multi-domain |
| 1 | Run-phase M1 — HUMAN GATE 5종 + Mode Selection logic | manager-develop | NOT-STARTED | n/a | REQ-WOF-001 + REQ-WOF-004 + REQ-WOF-014; writes new `human-gates.md` + `orchestration-mode-selection.md` rules + plan.md/run.md/sync.md router updates |
| 1 | Run-phase M2 — sync-phase 3 specialists restoration | manager-develop | NOT-STARTED | n/a | REQ-WOF-002 + REQ-WOF-007; updates sync.md + sub-skills + agent definitions for manager-quality / expert-security / manager-develop coverage invocation |
| 1 | Run-phase M3 — run-phase manager-strategy chain restoration | manager-develop | NOT-STARTED | n/a | REQ-WOF-003 + REQ-WOF-006; updates run.md + sub-skills + manager-develop-prompt-template.md to enforce 3-spawn chain |
| 1 | Run-phase M4 — Skill router discipline + Producer-Reviewer cycle | manager-develop | NOT-STARTED | n/a | REQ-WOF-005 + REQ-WOF-010; session-handoff.md clarification + run.md Phase 2.0 Sprint Contract + evaluator-active invocation |
| 1 | Run-phase M5 — plan-phase Explore + research.md + GitHub Issue + BODP audit | manager-develop | NOT-STARTED | n/a | plan.md restore Explore parallel spawn + Decision Point 1; Phase 2.5 GitHub Issue auto-creation reference; BODP audit trail reference |
| 1 | Run-phase M6 — manager-git PR doctrine + 4-artifact mirror parity | manager-develop | NOT-STARTED | n/a | REQ-WOF-013; resolve Hybrid Trunk vs PR-based contradiction; mirror all changes to `internal/template/templates/.claude/skills/moai/workflows/` + rule paths |
| 2 | Sync-phase (M5 in Tier L) | manager-docs | NOT-STARTED | n/a | CHANGELOG.md + frontmatter `status: in-progress → implemented` for all 4 artifacts + §E.4 Sync-phase Audit-Ready Signal in progress.md |
| 3 | Mx-phase (M6 in Tier L) | orchestrator | NOT-STARTED | n/a | Step C SKIP-judge per mx-tag-protocol.md §a (markdown-only, 0 .go files, 0 goroutines) expected; §E.5 Mx-phase Audit-Ready Signal in progress.md |
| 4 | 4-phase close | orchestrator | NOT-STARTED | n/a | Status `implemented → completed` + L60 atomic backfill if needed |

Key dates: created 2026-05-25, target plan-phase complete 2026-05-25, run-phase target 2026-05-26 to 2026-05-28 (3-day Tier L window per Sprint 10 cadence).

---

## §B — Run-phase Strategy

### §B.1 Target: Single-pass (Tier L 1-pass success)

Tier L SPECs target a 1-pass run-phase completion per the plan-auditor max-3 retry contract. Given that this SPEC is markdown-only (workflow rules + skills + agent definitions), the single-pass probability is HIGH provided:

1. plan-auditor verdict on this plan.md achieves PASS ≥ 0.85 (Tier L threshold)
2. Skip-eligible threshold 0.90 NOT required (markdown-heavy SPECs naturally MARGINAL on Tier L; PASS 0.85 acceptable)
3. M1-M6 milestone scope respects the §C in-scope file inventory exactly (path-specific staging discipline L46)

### §B.2 Known Issues Section B — Pre-flight injection

Per `.claude/rules/moai/development/manager-develop-prompt-template.md` §1.B, the run-phase delegation prompt MUST inject the B1-B12 known-issues catalog. For this Tier L markdown-only SPEC, the following B-categories are particularly relevant:

- **B3 Subagent Boundary Discipline** — REQ-WOF-011 makes this central; new rule files must enforce no AskUserQuestion in subagent prompts
- **B4 Frontmatter Canonical Schema** — all NEW rule file frontmatter must use canonical fields
- **B6 spec-lint Heading Regulation** — `## Out of Scope` h2 alone is insufficient; use h3 sub-section
- **B8 Working Tree Hygiene** — do NOT touch runtime files (`.moai/state/`, `.moai/harness/`, `.moai/cache/`)
- **B9 Git Commit + Push (Hybrid Trunk)** — manager-develop performs main-direct push per Tier L `--pr` flag exception or main-direct default per REQ-WOF-013 routing
- **B10 PRESERVE List Invariant** — only the §C.1 22-file scope inventory + 10-file mirror parity may be modified; no drive-by edits to predecessor SPECs
- **B12 Sync-phase CHANGELOG discipline** (manager-docs only) — duplicate-entry detection via `grep -c '<SPEC-ID>' CHANGELOG.md` before append

### §B.3 Pre-flight Self-verification (orchestrator before M1 spawn)

```bash
# 1. Working tree clean baseline
git rev-parse HEAD && git diff --stat
# Expected: 592b752e1, 0 changes (after plan-phase commit)

# 2. Divergence check (parallel session race detection)
git fetch origin main && git rev-list --count --left-right origin/main...HEAD
# Expected: "0 0" (clean)

# 3. SPEC artifact inventory verification
ls -la .moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/
# Expected: 4 files (spec.md, plan.md, acceptance.md, design.md) + research.md

# 4. Predecessor SPEC SHA verification (read-only inputs)
grep '^updated:\|^status:' .moai/specs/SPEC-V3R6-GEARS-MIGRATION-001/spec.md
grep '^updated:\|^status:' .moai/specs/SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001/spec.md
# Expected: status: implemented or completed for all 3 depends_on predecessors

# 5. GEARS lint baseline on existing workflow rules
grep -rn 'IF.*THEN' .claude/skills/moai/workflows/ | wc -l
# Expected: 0 (existing skill bodies already GEARS-migrated; new edits must remain GEARS-compliant)
```

---

## §C — Scope

### §C.1 In-scope file inventory (run-phase scope)

This subsection enumerates the run-phase modification scope. See spec.md §C.1 for the corresponding domain narrative.

**Tier 1 — Workflow router skills (3 files)**:
- `.claude/skills/moai/workflows/plan.md`
- `.claude/skills/moai/workflows/run.md`
- `.claude/skills/moai/workflows/sync.md`

**Tier 2 — Sub-skill modules (7 files)**:
- `.claude/skills/moai/workflows/plan/context-discovery.md`
- `.claude/skills/moai/workflows/plan/clarity-interview.md`
- `.claude/skills/moai/workflows/plan/spec-assembly.md`
- `.claude/skills/moai/workflows/run/context-loading.md`
- `.claude/skills/moai/workflows/run/phase-execution.md`
- `.claude/skills/moai/workflows/run/task-decomposition.md`
- `.claude/skills/moai/workflows/run/mode-orchestration.md`

**Tier 3 — Rule files (5 files, 2 NEW)**:
- `.claude/rules/moai/development/manager-develop-prompt-template.md`
- `.claude/rules/moai/workflow/session-handoff.md`
- `.claude/rules/moai/workflow/spec-workflow.md`
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` (NEW)
- `.claude/rules/moai/workflow/human-gates.md` (NEW)

**Tier 4 — Agent definitions (3 files)**:
- `.claude/agents/core/manager-strategy.md`
- `.claude/agents/core/manager-quality.md`
- `.claude/agents/expert/expert-security.md`

**Tier 5 — Mirror parity (template synchronization, ~10 files)**:
- `internal/template/templates/.claude/skills/moai/workflows/plan.md`
- `internal/template/templates/.claude/skills/moai/workflows/run.md`
- `internal/template/templates/.claude/skills/moai/workflows/sync.md`
- `internal/template/templates/.claude/skills/moai/workflows/plan/*.md` (3 files)
- `internal/template/templates/.claude/skills/moai/workflows/run/*.md` (4 files)
- `internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md` (NEW mirror)
- `internal/template/templates/.claude/rules/moai/workflow/human-gates.md` (NEW mirror)
- `internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md`
- `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md`
- `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md`
- `internal/template/templates/.claude/agents/{core,expert}/*.md` (3 files)

Estimated total run-phase scope: **22 local + 10 mirror = 32 files**. LOC equivalent ~1500 lines markdown additions/edits.

### §C.2 PRESERVE list (DO NOT MODIFY)

Per `manager-develop-prompt-template.md` §1.D Constraints, the following are explicit PRESERVE-list entries (out-of-scope):

- All files in `.moai/specs/SPEC-V3R6-*` predecessor directories (depends_on or sibling SPECs)
- All files in `internal/spec/` (Go-side spec-lint engine — out-of-scope per spec.md §F.4)
- All files in `.claude/rules/moai/languages/` (16-language neutrality contract per spec.md §F.7)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` (canonical SSOT — no body modification, only cross-references)
- All files in `.moai/state/`, `.moai/cache/`, `.moai/logs/`, `.moai/harness/usage-log.jsonl` (runtime-managed)
- All files in `.moai/research/` outside the current SPEC directory
- `CHANGELOG.md` body modifications outside the current SPEC's sync-phase entry

### §C.3 Out of Scope

#### §C.3.1 Out of Scope — Code-level changes
No Go source file modifications (`internal/*.go`, `cmd/*.go`, `pkg/*.go`) are within scope. Hook implementation (R10, R12) is deferred to follow-up SPECs.

#### §C.3.2 Out of Scope — Predecessor SPEC bodies
Strict L48 SSOT discipline — predecessor SPECs `SPEC-V3R6-{GEARS-MIGRATION,SKILL-GEARS-ALIGN,WORKFLOW-PLAN-GEARS-ALIGN,FOUNDATION-CORE-GEARS-ALIGN,PLAN-AUDITOR-GEARS-ALIGN}-001` are read-only inputs.

#### §C.3.3 Out of Scope — Language rule files
The 16 language rule files under `.claude/rules/moai/languages/` are out of orchestration scope. Path-specific loading (`paths:` frontmatter) is the canonical mechanism for language-scoped guidance per skill-authoring.md.

---

## §D — Milestone Decomposition

### §D.1 Plan-phase M0 — SPEC artifact authoring (CURRENT)

**Owner**: manager-spec
**Scope**: 4 files (spec.md + plan.md + acceptance.md + design.md)
**REQs covered**: all 15 REQ-WOF-XXX (declarative)
**Self-verification**:
- 12 canonical frontmatter fields present in spec.md
- GEARS notation ≥80% in §D Requirements
- ≥15 AC-WOF-XXX in acceptance.md
- Traceability 100% (every REQ-WOF has ≥1 AC-WOF)
- design.md §B.1 Delegation Graph + §B.3 Mode Selection Decision Tree present
**Commit**: `feat(SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001): plan-phase artifacts (Tier L Section A-G, 4 artifacts)`

### §D.2 Phase 0.5 — plan-auditor verdict

**Owner**: plan-auditor (subagent invocation by orchestrator)
**Threshold**: Tier L PASS ≥ 0.85 (mandatory), skip-eligible ≥ 0.90 (preferred but not required)
**MP-1 Scope clarity target**: ≥ 0.85
**MP-2 GEARS/EARS rigor target**: ≥ 0.85 (15 REQs with mandated GEARS pattern distribution)
**MP-3 Traceability target**: ≥ 0.85 (REQ-WOF↔AC-WOF mapping at 100%)
**MP-4 Risk-mitigation pairing target**: ≥ 0.85 (6 risks in §G each paired with mitigation strategy)
**Self-audit estimate**: ~0.87 PASS (Tier L MARGINAL, NOT skip-eligible — re-execution acceptable)
**Retry contract**: max 3 iterations; iter(N+1) < iter(N) triggers STOP signal

### §D.3 Phase 0.95 — Mode Selection (NEW)

**Owner**: orchestrator (autonomous decision)
**Inputs**: §C.1 scope (22 local + 10 mirror = 32 files), 5 domain categories (skills, sub-skills, rules, agents, mirrors)
**Expected mode**: `parallel` per REQ-WOF-009 State-driven `While` clause (≥10 files AND ≥3 domains both satisfied)
**Decision logged at**: `.moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/progress.md` § Mode Selection
**Rationale to record**: "Tier L + multi-domain (5 categories) + scope ≥10 files → parallel multi-spawn candidate, but since this is markdown-only (no Go code) the orchestrator MAY downgrade to sequential single-spawn manager-develop per Tier L Section A-E delegation template default. Final decision: sequential single-spawn manager-develop with Tier L Section A-E template + 6 milestone breakdown."

### §D.4 Run-phase M1 — HUMAN GATE 5종 + Mode Selection logic

**Owner**: manager-develop (single spawn, Tier L Section A-E template)
**Scope**: 6 files
- `.claude/rules/moai/workflow/human-gates.md` (NEW, ~600 LOC) — canonical inventory of 5 HUMAN GATEs with AskUserQuestion patterns
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` (NEW, ~500 LOC) — canonical 5-mode decision tree per REQ-WOF-004
- `.claude/skills/moai/workflows/plan.md` (UPDATE) — restore Decision Point 1 invocation reference
- `.claude/skills/moai/workflows/run.md` (UPDATE) — add Phase 0.95 Mode Selection between Phase 0.5 and Phase 1
- `.claude/skills/moai/workflows/sync.md` (UPDATE) — restore GATE 1 + GATE 2 + Decision Point 2 invocation references
- `internal/template/templates/...` mirror parity for the 5 files above
**REQs covered**: REQ-WOF-001, REQ-WOF-004, REQ-WOF-011, REQ-WOF-014
**ACs covered**: AC-WOF-001 through AC-WOF-004 (4 HUMAN GATE ACs) + AC-WOF-010 (Mode Selection log) + AC-WOF-013 (max-3 retry contract)
**Self-verification**: `grep -c "AskUserQuestion" .claude/rules/moai/workflow/human-gates.md` ≥ 5 (5 GATEs × ≥1 invocation pattern)

### §D.5 Run-phase M2 — sync-phase 3 specialists restoration

**Owner**: manager-develop (separate spawn)
**Scope**: 4 files
- `.claude/skills/moai/workflows/sync.md` (UPDATE) — Phase 0.5.4 manager-quality + Phase 0.55 expert-security + Phase 0.7 manager-develop coverage invocation surfaces
- `.claude/agents/core/manager-quality.md` (UPDATE) — sync-phase Phase 0.5.4 entrypoint documented
- `.claude/agents/expert/expert-security.md` (UPDATE) — sync-phase Phase 0.55 dependency manifest audit entrypoint
- mirror parity for the 3 files above
**REQs covered**: REQ-WOF-002, REQ-WOF-007
**ACs covered**: AC-WOF-005, AC-WOF-006, AC-WOF-007
**Self-verification**: `grep -n "Phase 0.5.4\|Phase 0.55\|Phase 0.7" .claude/skills/moai/workflows/sync.md` returns ≥3 lines

### §D.6 Run-phase M3 — manager-strategy chain restoration

**Owner**: manager-develop (separate spawn)
**Scope**: 5 files
- `.claude/skills/moai/workflows/run.md` (UPDATE) — Phase 1 manager-strategy → Phase 2 manager-develop hierarchical chain restoration
- `.claude/skills/moai/workflows/run/phase-execution.md` (UPDATE) — explicit `Agent(manager-strategy)` call before `Agent(manager-develop)` spawn sequence
- `.claude/skills/moai/workflows/run/task-decomposition.md` (UPDATE) — manager-strategy produces tasks.md artifact
- `.claude/agents/core/manager-strategy.md` (UPDATE) — code-prohibited HARD assertion + tasks.md artifact specification
- `.claude/rules/moai/development/manager-develop-prompt-template.md` (UPDATE) — §1.A Context section references manager-strategy tasks.md as upstream input
- mirror parity
**REQs covered**: REQ-WOF-003, REQ-WOF-006, REQ-WOF-008
**ACs covered**: AC-WOF-008 (manager-strategy invocation log + tasks.md artifact) + AC-WOF-012 (code-prohibited assertion)
**Self-verification**: `grep -c "manager-strategy" .claude/skills/moai/workflows/run.md` ≥ 3 (Phase 1 reference + sequence diagram + cross-reference)

### §D.7 Run-phase M4 — Skill router discipline + Producer-Reviewer cycle

**Owner**: manager-develop (separate spawn)
**Scope**: 4 files
- `.claude/rules/moai/workflow/session-handoff.md` (UPDATE) — Block 5 `실행:` clarification: `/moai <subcommand>` auto-triggers Skill() router; orchestrator MUST NOT manually Read SKILL.md body
- `.claude/skills/moai/workflows/run.md` (UPDATE) — Phase 2.0 Sprint Contract negotiation entry point for `harness: thorough`
- `.claude/skills/moai/workflows/run/mode-orchestration.md` (UPDATE) — evaluator-active invocation pattern with `max_iterations: 3`
- mirror parity
**REQs covered**: REQ-WOF-005, REQ-WOF-010
**ACs covered**: AC-WOF-009 (Skill router invocation evidence) + (extended in acceptance.md if Sprint Contract AC is added)
**Self-verification**: `grep -n "Skill(\"moai\"" .claude/rules/moai/workflow/session-handoff.md` matches ≥1; `grep -n "evaluator-active" .claude/skills/moai/workflows/run.md` matches ≥1

### §D.8 Run-phase M5 — plan-phase Explore + research.md + GitHub Issue + BODP audit

**Owner**: manager-develop (separate spawn)
**Scope**: 5 files
- `.claude/skills/moai/workflows/plan.md` (UPDATE) — Phase 0 Explore parallel-spawn pattern (3-5 read-only subagents per Anthropic multi-agent research blog)
- `.claude/skills/moai/workflows/plan/context-discovery.md` (UPDATE) — codify Explore parallel-spawn implementation
- `.claude/skills/moai/workflows/plan/spec-assembly.md` (UPDATE) — research.md emission requirement + Decision Point 1 reference
- `.claude/skills/moai/workflows/run.md` (UPDATE) — Phase 2.5 GitHub Issue auto-creation + Phase 2.8 BODP audit trail `.moai/branches/decisions/<branch>.md`
- mirror parity
**REQs covered**: (cross-referenced) supports REQ-WOF-001 (Decision Point 1) and acceptance.md ACs 14-15
**ACs covered**: AC-WOF-014 (GitHub Issue creation) + AC-WOF-015 (BODP audit trail)
**Self-verification**: `grep -c "Explore" .claude/skills/moai/workflows/plan.md` ≥ 2 (phase reference + parallel-spawn pattern reference); `grep -c "gh issue create" .claude/skills/moai/workflows/run.md` ≥ 1

### §D.9 Run-phase M6 — manager-git PR doctrine + 4-artifact mirror parity completion

**Owner**: manager-develop (separate spawn) + manager-docs (sync-phase entry)
**Scope**: 4 files
- `.claude/skills/moai/workflows/sync.md` (UPDATE) — manager-git PR creation routing logic per REQ-WOF-013 (Tier L OR `--pr` flag → PR; else Hybrid Trunk main-direct)
- `.claude/rules/moai/workflow/spec-workflow.md` (UPDATE) — Phase 0.95 Mode Selection documented between Phase 0.5 and Phase 1 (NEW subsection)
- `.claude/skills/moai/workflows/plan/clarity-interview.md` (UPDATE) — verify Socratic interview adheres to askuser-protocol.md SSOT
- mirror parity
**REQs covered**: REQ-WOF-013, REQ-WOF-015 (compound clause routing demonstration)
**ACs covered**: AC-WOF-011 (multi-spawn parallel evidence when scope ≥10 files) + traceability completion
**Self-verification**: All 22 local + 10 mirror files modified; spec-lint clean on all touched markdown; `git diff --stat` shows path-specific changes only

### §D.10 Sync-phase (Tier L M7)

**Owner**: manager-docs
**Scope**: CHANGELOG.md unique entry under `[Unreleased]` + frontmatter `status: in-progress → implemented` for all 4 SPEC artifacts + §E.4 Sync-phase Audit-Ready Signal in progress.md + commit `chore(SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001): sync-phase artifacts`
**Self-verification per manager-develop-prompt-template.md §1.B B12**:
- `grep -c 'SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001' CHANGELOG.md` exactly 1 (no duplicate)
- All 4 SPEC artifact frontmatter `sync_commit_sha:` filled with the sync commit SHA atomically
- AC count in CHANGELOG matches `acceptance.md` SSOT exactly

### §D.11 Mx-phase (Tier L M8)

**Owner**: orchestrator
**Scope**: Mx Step C SKIP-judge per mx-tag-protocol.md §a — markdown-only, 0 .go files modified, 0 goroutines added, 0 fan_in delta. Expected SKIP-eligible.
**Output**: §E.5 Mx-phase Audit-Ready Signal in progress.md + commit `chore(SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001): Mx-phase audit-ready signal + 4-phase close`

### §D.12 4-phase close

**Owner**: orchestrator
**Action**: status `implemented → completed` in all 4 artifact frontmatter; L60 atomic backfill if any sync_commit_sha "pending" remains; verify push range matches all attributed commits.

---

## §E — Verification Strategy

### §E.1 Per-milestone self-verification (manager-develop)

After each M1-M6 completion, the manager-develop subagent MUST emit a self-verification report per `manager-develop-prompt-template.md` §1.E covering:

- **E1** AC binary PASS/FAIL matrix (the milestone-specific ACs from §D.4..§D.9 above)
- **E2** Cross-platform build N/A (markdown-only SPEC; `go build ./...` baseline check is acceptable)
- **E3** Coverage N/A (no Go code modified)
- **E4** Subagent boundary grep:
  ```bash
  grep -rn 'AskUserQuestion' .claude/skills/moai/workflows/ .claude/rules/moai/workflow/ \
    | grep -v "(권장)" | grep -v "# Example" | head -20
  ```
  Expected: matches reference the orchestrator-side patterns only; no subagent prompt contains direct `AskUserQuestion` invocation syntax.
- **E5** Lint status: `markdownlint` clean OR `moai spec lint .moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/spec.md` returns 0 errors
- **E6** Branch HEAD + push status: new commit SHA listed + `git push origin main` exit 0
- **E7** Blocker report (if any): structured report — no AskUserQuestion from subagent

### §E.2 Orchestrator 7-item Trust-but-verify parallel batch

Per `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution + `.claude/rules/moai/workflow/verification-batch-pattern.md`, after each manager-develop return the orchestrator MUST execute the following 7 verifications in parallel (single response turn, multi-Bash):

```bash
# V1. Test suite (full Go build — sanity check no accidental Go regression)
go test ./internal/spec/... 2>&1 | tail -3

# V2. Divergence (parallel session race detection)
git fetch origin main && git rev-list --count --left-right origin/main...HEAD
# Expected: "0 0"

# V3. Subagent boundary grep
grep -rn 'AskUserQuestion\|mcp__askuser' .claude/rules/moai/workflow/orchestration-mode-selection.md .claude/rules/moai/workflow/human-gates.md 2>/dev/null \
  | grep -v "# " | head -10

# V4. Sentinel-key audit (no FROZEN_SENTINEL drift)
grep -rn 'FROZEN_SENTINEL\|HARNESS_FROZEN' internal/ 2>/dev/null | head -5

# V5. CLI smoke (markdown-only SPEC; verify moai still compiles)
go run ./cmd/moai --version

# V6. Lint baseline (golangci-lint baseline still clean)
golangci-lint run --timeout=2m 2>&1 | tail -3

# V7. spec-lint on the 4 SPEC artifacts
go run ./cmd/moai spec lint .moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/ 2>&1 | tail -5
```

Expected: V1 PASS / V2 "0 0" / V3 0 matches / V4 0 matches / V5 version printed / V6 baseline matches pre-spawn count / V7 0 errors on the SPEC's own artifacts.

### §E.3 Final acceptance verification (end of M6, before sync-phase entry)

Run the full AC-WOF matrix per acceptance.md §A Mandatory Acceptance Criteria. Each AC must show PASS verdict with evidence command output captured in `progress.md § E.2 Run-phase Evidence`.

### §E.4 Plan-auditor re-execution policy

Plan-auditor re-execution is gated by the policy in `.claude/rules/moai/workflow/spec-workflow.md` § Phase 0.5 Plan Audit Gate § Plan to Run § Phase 0.5 skip-eligibility: when the most recent plan-auditor verdict on this SPEC was PASS with overall score ≥ 0.90 AND no plan-PR commit has landed since that verdict, the orchestrator MAY skip Phase 0.5 re-execution. Otherwise, Phase 0.5 runs at the start of `/moai run`.

For this SPEC (Tier L), the self-audit estimate is ~0.87 (MARGINAL). Phase 0.5 will run.

---

## §F — Cross-References

### §F.1 Plan-phase canonical references

- `.claude/rules/moai/workflow/spec-workflow.md` — Plan/Run/Sync phase + Tier S/M/L
- `.claude/skills/moai/workflows/plan.md` — current plan workflow router (target of M1 + M5 updates)
- `.claude/agents/meta/plan-auditor.md` — plan-auditor verdict schema + retry contract

### §F.2 Run-phase canonical references

- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Section A-E 5-section template (Tier L mandatory)
- `.claude/skills/moai/workflows/run.md` — current run workflow router (target of M1 + M3 + M4 updates)
- `.claude/rules/moai/workflow/context-window-management.md` — model-specific threshold (1M = 50% / 200K = 90%) for context-aware mode selection

### §F.3 Sync-phase + Mx canonical references

- `.claude/skills/moai/workflows/sync.md` — current sync workflow router (target of M1 + M2 + M6 updates)
- `.claude/rules/moai/workflow/mx-tag-protocol.md` — Mx Step C SKIP-judge protocol §a (markdown-only escape clause)

### §F.4 Research synthesis (this SPEC)

- `.moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/research.md` — orchestrator-authored synthesis of 3 parallel research agents

### §F.5 Acceptance + design siblings

- `.moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/acceptance.md` — 15+ AC-WOF-XXX with traceability matrix
- `.moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/design.md` — Tier L exclusive architecture rationale + delegation graph + 5-mode decision tree

---

Version: 0.1.0
Status: plan-phase initial authoring (M0 in progress)
Tier: L
