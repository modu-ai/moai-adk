# SPEC-V3R5-WORKFLOW-LEAN-001 — Implementation Plan

> **LEAN dogfooding**: This plan is Tier S. Section A-E template NOT applied to milestone delegations. Run-phase MAY use minimal delegation prompts (task + constraints + AC only). plan-auditor PASS threshold for this SPEC: 0.75.

## 1. Overview

Four sequential milestones, each modifies one rule/skill/agent file. No Go code, no test additions, no CI changes. Total estimated diff: ~150-250 LOC across 5 markdown files.

Late-branch discipline: commits accumulate on main (per SPEC-V3R5-LATE-BRANCH-001 v0.3.0). PR branch `feat/SPEC-V3R5-WORKFLOW-LEAN-001` created at PR-time.

## 2. Milestones

### M1 — spec-workflow.md: Tier S/M/L Definition

**File**: `.claude/rules/moai/workflow/spec-workflow.md`

**Changes**:

- Add new section `## SPEC Complexity Tier` (or insert into existing phase-overview section) with:
  - Tier S: <300 LOC scope OR <5 files → 2 artifacts (spec.md + plan.md), AC inline in spec.md §3
  - Tier M: 300-1000 LOC OR 5-15 files → 3 artifacts (+ acceptance.md)
  - Tier L: 1000+ LOC OR 15+ files OR multi-package → 5 artifacts (current default, + design.md + research.md)
  - Tier decision rule: rounded UP on borderline (when in doubt, choose higher tier)
- Add cross-reference to `.claude/rules/moai/development/spec-frontmatter-schema.md` for the optional `tier` field.

**Supplement to spec-frontmatter-schema.md** (same milestone):

- Add `tier` to Optional Fields table: `tier: S | M | L` (enum, optional, absence = L for backward compat)
- One-line description: "SPEC complexity tier; controls artifact count and plan-auditor threshold."

**Satisfies**: REQ-WL-001, REQ-WL-002, REQ-WL-003, REQ-WL-004, REQ-WL-009

**AC binary check**: AC-WL-001, AC-WL-009

**Delegation hint**: Minimal prompt. No Section A-E. Pass spec.md path + this milestone scope + AC commands. Subagent reads spec.md §3.1 and §3.6 directly.

---

### M2 — manager-develop-prompt-template.md: Section A-E Optional Clause

**File**: `.claude/rules/moai/development/manager-develop-prompt-template.md`

**Changes**:

- Insert a new section at the top of "1. 표준 위임 Prompt 5-Section 구조" (before Section A description) titled `## When to Apply` or `## Tier Applicability`:
  - For Tier S SPECs: Section A-E MAY be skipped. Minimal delegation prompt = task + constraints + AC only.
  - For Tier M SPECs: Section A-E SHOULD be applied. Section B (known issues B1-B8) MAY filter by domain relevance.
  - For Tier L SPECs: Section A-E MUST be applied. Full B1-B8 inclusion recommended.
- Add example minimal Tier S delegation prompt (10-15 lines) showing the pattern.
- Add anti-pattern note: "Applying Section A-E to a Tier S SPEC inflates delegation prompt baseline to ~4500 tokens. Do not apply when SPEC scope is <300 LOC."

**Satisfies**: REQ-WL-005, REQ-WL-011

**AC binary check**: AC-WL-002

**Delegation hint**: Minimal prompt. Read manager-develop-prompt-template.md current state, insert the tier-applicability section, preserve existing Section A-E content unchanged.

---

### M3 — plan-auditor.md: STOP Escalation + Tier Thresholds + Iter Cap

**File**: `.claude/agents/moai/plan-auditor.md`

**Changes** (three related additions in the agent prompt body):

1. **Tier-differentiated PASS threshold** (one new subsection or table):
   - Tier S: aggregate ≥ 0.75 → PASS
   - Tier M: aggregate ≥ 0.80 → PASS
   - Tier L: aggregate ≥ 0.85 → PASS
   - Read tier from `spec.md` frontmatter `tier` field; default L when absent.

2. **STOP-on-score-regression** (one new behavior section):
   - If iter(N+1) aggregate < iter(N) aggregate (any magnitude), emit a STOP signal.
   - STOP signal format: `STATUS: SCOPE_REDUCTION_RECOMMENDED` with a 3-5 line proposal describing what scope to drop.
   - User receives the proposal via the orchestrator (via AskUserQuestion); user MAY override and continue to next iter.
   - This is NOT a hard abort — it is a structured pause for human judgment.

3. **Max 3 iterations cap**:
   - After iter3 completes (regardless of pass/fail), the agent SHALL emit one of three outcomes:
     - `STATUS: PASS` (if threshold met)
     - `STATUS: PASS_WITH_DEBT` (if 0.05 below threshold; orchestrator records as known debt)
     - `STATUS: ESCALATE` (if >0.05 below threshold; orchestrator surfaces to user with three options: scope reduction / extend cap / abandon SPEC)

**Satisfies**: REQ-WL-006, REQ-WL-007, REQ-WL-008

**AC binary check**: AC-WL-003, AC-WL-007, AC-WL-008

**Delegation hint**: Minimal prompt. Read plan-auditor.md current state, append three subsections under existing scoring rubric. Preserve all existing audit dimensions unchanged.

---

### M4 — spec-assembly.md: Tier Judgment Socratic Question

**File**: `.claude/skills/moai/workflows/plan/spec-assembly.md`

**Changes**:

- Add a new step early in the assembly workflow (before SPEC artifact creation begins) titled `## Step: Tier Judgment`:
  - Orchestrator invokes AskUserQuestion with options:
    - Tier S (Recommended for <300 LOC scope, <5 files) — 2 artifacts, AC inline
    - Tier M (300-1000 LOC, 5-15 files) — 3 artifacts
    - Tier L (1000+ LOC or multi-package) — 5 artifacts (current default behavior)
    - Other (free-form: user specifies tier with justification)
  - If user request explicitly contains "Tier S/M/L" in the original phrasing, MAY skip the question and use the explicit tier.
  - Record the chosen tier in spec.md frontmatter `tier:` field.

**Satisfies**: REQ-WL-010, REQ-WL-012

**AC binary check**: AC-WL-004

**Delegation hint**: Minimal prompt. Read spec-assembly.md current state, insert the Tier Judgment step before the existing artifact creation steps.

---

## 3. Verification (run-phase end gate)

After all M1-M4 commits, the orchestrator runs the following batch (single-turn parallel) to verify ACs:

```bash
# AC-WL-001: spec-workflow.md tier definition
grep -n "Tier S/M/L\|complexity tier" .claude/rules/moai/workflow/spec-workflow.md

# AC-WL-002: manager-develop-prompt-template.md optional clause
grep -n "Tier S\|optional\|Tier S/M/L" .claude/rules/moai/development/manager-develop-prompt-template.md

# AC-WL-003: plan-auditor.md STOP escalation
grep -nE "STOP|score regression|escalation|iter[23]\b" .claude/agents/moai/plan-auditor.md

# AC-WL-004: spec-assembly.md tier judgment
grep -nE "Tier|complexity tier" .claude/skills/moai/workflows/plan/spec-assembly.md

# AC-WL-007: plan-auditor.md tier thresholds
grep -nE "0\.75|0\.80|tier threshold|tier-differentiated" .claude/agents/moai/plan-auditor.md

# AC-WL-008: plan-auditor.md iteration cap
grep -nE "max 3|iteration cap|3 iterations|three iterations" .claude/agents/moai/plan-auditor.md

# AC-WL-009: spec-frontmatter-schema.md tier field
grep -n "tier:" .claude/rules/moai/development/spec-frontmatter-schema.md

# AC-WL-006: this SPEC dir contains exactly 2 files
ls .moai/specs/SPEC-V3R5-WORKFLOW-LEAN-001/ | wc -l  # must equal 2

# AC-WL-010: total LOC ≤ 800
wc -l .moai/specs/SPEC-V3R5-WORKFLOW-LEAN-001/spec.md .moai/specs/SPEC-V3R5-WORKFLOW-LEAN-001/plan.md

# AC-WL-011: spec lint clean
go run ./cmd/moai spec lint --strict .moai/specs/SPEC-V3R5-WORKFLOW-LEAN-001/spec.md
```

All commands MUST be invoked in a single orchestrator turn as parallel Bash calls.

AC-WL-005 (end-to-end Tier S validation) is deferred to the next simple SPEC's run-phase; this SPEC's run-phase does not measure it.

## 4. Delegation Policy

This SPEC dogfoods LEAN. Delegation rules:

- **No Section A-E template applied** to any of M1-M4 delegations. Each delegation is a minimal prompt:
  - Task description (1-3 sentences)
  - File path
  - Specific changes (bullet list from §2)
  - AC commands to self-verify
  - Constraint: preserve existing content, additive changes only

- **Single manager-develop delegation** for all 4 milestones is RECOMMENDED (over 4 separate delegations). Each milestone touches one file with additive changes; no inter-milestone dependencies beyond file ordering.

- **plan-auditor on this SPEC**: PASS threshold = 0.75 (Tier S). Max 3 iterations. If iter2 < iter1 → STOP and surface scope reduction proposal.

- **No subagent baseline grep** required: the AC commands in §3 are self-contained shell commands; subagent reads them directly.

## 5. Risks (Implementation-specific)

- R-WL-IMPL-001 (Low): grep patterns in §3 may match pre-existing lines unrelated to this SPEC's changes. Mitigation: M1-M4 commits each include a deliberate marker phrase (e.g., "SPEC complexity tier", "Tier S/M/L", "STOP-on-regression", "Tier judgment") that subsequent grep can disambiguate.

- R-WL-IMPL-002 (Low): The orchestrator may forget to invoke verification commands in parallel. Mitigation: §3 lists them as a single bash block to copy-paste; the run-phase orchestrator follows verification-batch-pattern.md.

- R-WL-IMPL-003 (Medium): Adding `tier` field documentation to spec-frontmatter-schema.md must NOT change `FrontmatterSchemaRule` lint behavior (C-WL-001). Mitigation: M1 supplement edits only the "Optional Fields" table; the 12 required fields remain untouched.

## 6. Out of Scope (Run-Phase Boundaries)

- No Go code changes (C-WL-003).
- No test additions or modifications.
- No CI workflow changes (no `.github/workflows/` edits).
- No `internal/spec/lint.go` changes (C-WL-001).
- No `internal/spec-auditor/` changes (C-WL-002).
- No retroactive Tier assignment to existing SPECs.

## 7. Post-Implementation

After PR merge:

1. Apply LEAN to the next simple SPEC chosen by the user (candidates: SPEC-V3R5-LANG-COMPLIANCE-002 re-entry, or another Tier S SPEC).
2. Measure AC-WL-005 metrics (tool calls, wall-time, plan-auditor iter count, artifact count) and record in that SPEC's run-phase progress.md.
3. If measurement confirms ≤30 tool calls and ≤15 min wall-time, mark this SPEC `completed`. If not, file a follow-up SPEC for refinement.
