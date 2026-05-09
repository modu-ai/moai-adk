---
spec_id: SPEC-V3R2-ORC-001
phase: "1B — Implementation Plan"
created_at: 2026-05-09
author: manager-spec
status: audit-ready
base_commit: "464366583"
branch: feature/SPEC-V3R2-ORC-001-roster
---

# Plan: Agent roster consolidation (22 → 17)

Implementation plan for SPEC-V3R2-ORC-001. Defines milestones M1-M5 with
file:line anchors, mx_plan tags, owner roles, and verification gates. All
work happens **template-first** under
`internal/template/templates/.claude/agents/moai/`, then `make build`
regenerates the embedded copy and the local `.claude/agents/moai/` tree
mirrors byte-identical (CLAUDE.local.md §2 Template-First Rule).

---

## 1. Strategy Overview

This SPEC inherits a **partially-completed roster** from
SPEC-V3R3-RETIRED-AGENT-001 (PR #776) and SPEC-V3R3-RETIRED-DDD-001 (PR #781):

- `manager-cycle.md` already exists with full body and `cycle_type` parameter.
- `manager-ddd.md` and `manager-tdd.md` are already retired stubs.

Therefore, the implementation work is concentrated on the **5 outstanding
retirements** (3 builders, expert-debug, expert-testing) and the **6 refactor
passes** described in research.md §3.

The plan is sequenced to minimize git-history rewrite risk: each milestone
produces an atomic, reviewable diff with its own verification gate.

| Milestone | Scope | Owner | Gate |
|-----------|-------|-------|------|
| M1 | Verification of carry-over + delivery contract | manager-spec | research.md citations re-verified, OQ-1..5 resolved |
| M2 | Retire 3 builders + 2 experts (5 stubs); create builder-platform | expert-backend | `diff -r` template↔local clean; trigger-union test PASS |
| M3 | Refactor passes (6) on existing agents | expert-backend | All 6 frontmatter/body changes applied; `grep -c` verifications pass |
| M4 | Downstream reference sync (rules, skills, CLAUDE.md) | manager-docs | No remaining literal `manager-ddd\|manager-tdd\|builder-{agent,skill,plugin}\|expert-debug\|expert-testing` outside stub bodies and migration tables |
| M5 | REFACTOR + MX tags + completion gate | manager-quality | `make build` byte-identical; all ACs PASS; PR ready |

---

## 2. Milestone Detail

### 2.1 M1 — Verification + Delivery Contract Definition

**Owner**: manager-spec (audit), expert-backend (re-verification)
**Priority**: P0 (blocks everything)

**Tasks**:

1. M1.1 — Re-verify carry-over state:
   - Read `internal/template/templates/.claude/agents/moai/manager-cycle.md` and confirm L17 tools list, L54 cycle_type declaration, L60-66 migration table, L8-15 trigger union (research.md §3.1 citations).
   - Read `manager-ddd.md` and `manager-tdd.md` stubs; confirm `retired: true`, `retired_replacement: manager-cycle`, body has only one-line redirect.
   - Run `diff -r internal/template/templates/.claude/agents/moai/ .claude/agents/moai/` and capture the baseline divergence list (5 files known to differ — research.md §1).

2. M1.2 — Resolve open questions OQ-1..OQ-5 (research.md §4):
   - OQ-1: Confirm 17-active interpretation (R5-baseline) vs 20-actual (with post-R5 additions).
   - OQ-2: Confirm stub frontmatter schema (use SPEC-V3R3-RETIRED-DDD-001 pattern).
   - OQ-3: Exclude `examples.md` from M4 sweep.
   - OQ-4: Adopt ORC-002 lint contract preemptively.
   - OQ-5: KEEP `claude-code-guide.md`; update spec.md §10.1 destiny table.

3. M1.3 — Define delivery contract (artefacts):
   - 5 new stub files (3 builders, expert-debug, expert-testing).
   - 1 new active file (builder-platform.md).
   - 6 modified active files (manager-quality, manager-project, expert-backend, expert-performance, manager-git, plan-auditor).
   - Updated CLAUDE.md Agent Catalog (L106-126).
   - Updated `.moai/specs/SPEC-V3R2-ORC-001/spec.md` §10.1 destiny table (post-R5 additions row added).

**Files touched**: NONE (read-only verification + specification)

**mx_plan tags**:
- `@MX:NOTE M1 contract` on the spec.md §10.1 update line (intent: documenting the post-R5 reconciliation).

**Verification gate**:
- All research.md §3 citations re-verified.
- `diff -r` baseline captured to `.moai/specs/SPEC-V3R2-ORC-001/diff-baseline.txt` (M1 artefact, optional).
- OQ-1..5 documented as resolved in HISTORY of spec.md.

---

### 2.2 M2 — Retire 5 Agents + Create builder-platform

**Owner**: expert-backend (file authoring), manager-spec (review)
**Priority**: P0 (largest diff; blocks M3-M5)

**Tasks**:

1. M2.1 — Author `internal/template/templates/.claude/agents/moai/builder-platform.md`:
   - Frontmatter: `name: builder-platform`, `model: sonnet`, `permissionMode: bypassPermissions`, `memory: user`, `effort:` deferred to ORC-003.
   - Required parameter: `artifact_type: agent|skill|plugin|command|hook|mcp-server|lsp-server`.
   - Trigger keywords (REQ-ORC-001-009 trigger-union): inherit from `builder-agent.md` L6-9 + `builder-skill.md` L6-9 + `builder-plugin.md` L6-9, deduped per artifact_type.
   - Tools list: same shape as the 3 source agents (Read, Write, Edit, Grep, Glob, WebFetch, WebSearch, Bash, TodoWrite, Agent, Skill, mcp__sequential-thinking, mcp__context7__*).
   - Skills: union (`moai-foundation-cc, moai-foundation-core, moai-workflow-project`).
   - Body: 5-phase workflow (Requirements → Research → Architecture → Implementation → Validation), with artifact_type dispatch table.
   - Migration Notes table (mirroring manager-cycle.md L60-66 pattern).

2. M2.2 — Replace `internal/template/templates/.claude/agents/moai/builder-agent.md` body with retired-stub schema (matching `manager-ddd.md` L1-39):
   - `retired: true`, `retired_replacement: builder-platform`, `retired_param_hint: "artifact_type=agent"`, `tools: []`, `skills: []`.
   - Body: 1-line redirect + Migration Guide table.

3. M2.3 — Same as M2.2 for `builder-skill.md` (`retired_param_hint: "artifact_type=skill"`).

4. M2.4 — Same as M2.2 for `builder-plugin.md` (`retired_param_hint: "artifact_type=plugin"`).

5. M2.5 — Replace `expert-debug.md` body with retired-stub schema:
   - `retired_replacement: manager-quality`, `retired_param_hint: "diagnostic-mode"`.
   - Body: 1-line redirect, references manager-quality.md Diagnostic Sub-Mode (added in M3).

6. M2.6 — Replace `expert-testing.md` body with retired-stub schema:
   - `retired_replacement: manager-cycle | expert-performance` (dual-target).
   - Body: 1-line redirect with split table (strategy → manager-cycle cycle_type=tdd; load → expert-performance --deepthink load-test).

7. M2.7 — Run `make build` to regenerate `internal/template/embedded.go`.

8. M2.8 — Run `diff -r internal/template/templates/.claude/agents/moai/ .claude/agents/moai/`; verify byte-identical (the 5 known divergences from M1 will resolve here once stubs replace bodies).

**Files touched** (template-first, mirrored to local via `make build`):
- `internal/template/templates/.claude/agents/moai/builder-platform.md` (NEW)
- `internal/template/templates/.claude/agents/moai/builder-agent.md` (BODY REPLACED with stub)
- `internal/template/templates/.claude/agents/moai/builder-skill.md` (BODY REPLACED with stub)
- `internal/template/templates/.claude/agents/moai/builder-plugin.md` (BODY REPLACED with stub)
- `internal/template/templates/.claude/agents/moai/expert-debug.md` (BODY REPLACED with stub)
- `internal/template/templates/.claude/agents/moai/expert-testing.md` (BODY REPLACED with stub)
- `internal/template/embedded.go` (REGENERATED by `make build`)

**mx_plan tags**:
- `@MX:ANCHOR builder-platform.cycle_type` on builder-platform.md `**artifact_type**:` line (high fan_in: every artifact creation routes through it).
- `@MX:NOTE retirement-pattern` on each new stub file (intent: matches SPEC-V3R3-RETIRED-DDD-001).
- `@MX:WARN trigger-union` on builder-platform.md trigger row (REASON: REQ-ORC-001-017 forbids trigger drops; CI test enforces).

**Verification gate**:
- `find internal/template/templates/.claude/agents/moai/ -name "*.md" -type f | wc -l` = 26 (no file count change).
- `grep -l "retired: true" internal/template/templates/.claude/agents/moai/*.md | wc -l` = 7 (manager-ddd, manager-tdd, builder-agent, builder-skill, builder-plugin, expert-debug, expert-testing).
- Trigger-union test (custom Bash script comparing trigger rows of new stub-source bodies vs builder-platform frontmatter) returns 0 dropped keywords.
- `diff -r` template↔local empty.

---

### 2.3 M3 — Refactor Passes on Existing Agents

**Owner**: expert-backend (file edits), manager-quality (review)
**Priority**: P1 (depends on M2 retirement of expert-debug for 3.1)

**Tasks**:

1. M3.1 — Add Diagnostic Sub-Mode to `manager-quality.md`:
   - Insert new "## Diagnostic Sub-Mode" section after the Primary Mission block.
   - Content: lift the delegation routing table from retired `expert-debug.md` (L59-200 range, semantic equivalent), preserving the 70%-router shape.
   - Drop `mcp__context7__resolve-library-id, mcp__context7__get-library-docs` from tools list (research.md §3.8 — REQ-ORC-001-011).
   - File: `internal/template/templates/.claude/agents/moai/manager-quality.md`
   - Anchor: L13 (tools list), L30+ (after Primary Mission).

2. M3.2 — Scope-shrink `manager-project.md`:
   - Remove L57-65 routing block (`settings_modification`, `glm_configuration`, `template_update_optimization` modes).
   - Replace with a "Scope Boundary" section: only `.moai/project/{product,structure,tech}.md` writes allowed.
   - Add blocker-report template referenced by REQ-ORC-001-008 (point at `moai update`, `moai glm`, `moai cc`).
   - Anchor: L57-65 (existing block to remove), L66+ (insert blocker-report template).

3. M3.3 — Trigger dedup for `expert-backend.md`:
   - Reduce L6 EN row from 22 tokens to 12-15 high-precision tokens (drop `data modeling`, merge SQL/NoSQL forms; keep one Oracle form per row).
   - Drop standalone `Oracle` from KO/JA/ZH rows where the localized form is also present (e.g., L7 keep `오라클` only).
   - Anchor: L6 (EN), L7 (KO), L8 (JA), L9 (ZH).
   - Verification: `grep -c "Oracle" expert-backend.md` returns ≤ 4 (one per language row, not 4 per row).

4. M3.4 — Trigger trim for `manager-git.md`:
   - Reduce L6-9 trigger rows to ≈8 high-precision EN tokens per P-A16; same proportions for KO/JA/ZH.
   - Anchor: L6-9.

5. M3.5 — Add `Write` tool to `expert-performance.md` (REQ-ORC-001-014):
   - Insert `Write` between `Read` and `Edit` in tools list.
   - Add a body-level "Write Scope" subsection limiting writes to `.moai/docs/performance-analysis-*.md`.
   - Anchor: tools list line (L13 of current file).

6. M3.6 — Add `memory: project` to `plan-auditor.md` (REQ-ORC-001-013):
   - Insert `memory: project` line after `permissionMode: default` (current L15 frontmatter).
   - Anchor: L15-17.

7. M3.7 — Run `make build` again to regenerate embedded copy.

**Files touched**:
- `internal/template/templates/.claude/agents/moai/manager-quality.md`
- `internal/template/templates/.claude/agents/moai/manager-project.md`
- `internal/template/templates/.claude/agents/moai/expert-backend.md`
- `internal/template/templates/.claude/agents/moai/manager-git.md`
- `internal/template/templates/.claude/agents/moai/expert-performance.md`
- `internal/template/templates/.claude/agents/moai/plan-auditor.md`
- `internal/template/embedded.go` (REGENERATED)

**mx_plan tags**:
- `@MX:ANCHOR diagnostic-routing` on manager-quality.md Diagnostic Sub-Mode header.
- `@MX:NOTE scope-shrink` on manager-project.md Scope Boundary block.
- `@MX:WARN trigger-cap` on expert-backend.md L6 (REASON: P-A17 cap 12-15 EN tokens enforced by CI in ORC-002).
- `@MX:NOTE memory-add` on plan-auditor.md frontmatter.

**Verification gate**:
- `grep -c "Oracle" expert-backend.md` ≤ 4.
- `grep -c "context7" manager-quality.md` = 0.
- `grep -c "memory: project" plan-auditor.md` ≥ 1.
- `grep -c "settings_modification\|glm_configuration\|template_update_optimization" manager-project.md` = 0.
- `grep -c "Diagnostic Sub-Mode" manager-quality.md` ≥ 1.
- `grep "Write" expert-performance.md | head -1` shows tools list contains Write.

---

### 2.4 M4 — Downstream Reference Sync

**Owner**: manager-docs (sweep + edits), manager-spec (review)
**Priority**: P1 (depends on M3 for the destination shapes)

**Tasks**:

1. M4.1 — Sweep skill files for retired-agent references:
   - Pattern: `manager-ddd|manager-tdd|builder-agent|builder-skill|builder-plugin|expert-debug|expert-testing`
   - Scope: `internal/template/templates/.claude/skills/` and `.claude/skills/` (mirrored sets, ~57 hits per research.md §3.11).
   - Exclude: `examples.md` files (per OQ-3 resolution).
   - Replacement table per REQ-ORC-001-007.

2. M4.2 — Sweep rule files for retired-agent references:
   - Scope: `internal/template/templates/.claude/rules/` and `.claude/rules/` (~28 hits).
   - Same replacement table.

3. M4.3 — Update root `CLAUDE.md` Agent Catalog (L106-126):
   - L108: Manager Agents reduce from 8 to 7 (remove `ddd`, `tdd`; add `cycle`).
   - L112: Expert Agents reduce from 8 to 6 (remove `debug`, `testing`; the count adjusts).
   - L116: Builder Agents reduce from 3 to 1 (replace `agent, skill, plugin` with `platform`).
   - L62, L147, L378: replace specific agent invocation strings per REQ-ORC-001-007 table.

4. M4.4 — Update template `internal/template/templates/CLAUDE.md` mirror:
   - L136: replace `builder-agent subagent` with `builder-platform subagent`.
   - L378: replace `expert-debug subagent` with `manager-quality (diagnostic mode) subagent`.

5. M4.5 — Run `make build` if any template-tree CLAUDE.md changed; final `diff -r` must be clean.

**Files touched** (template-first; local mirror updated via M4.5 build step):
- ~57 skill files (subset; `examples.md` excluded)
- ~28 rule files
- `CLAUDE.md` (root)
- `internal/template/templates/CLAUDE.md`
- `internal/template/embedded.go` (if applicable)

**mx_plan tags**:
- `@MX:NOTE migration-table` on the root CLAUDE.md Agent Catalog block (intent: links REQ-ORC-001-007 migration table).
- `@MX:TODO catalog-count` on L108/L112/L116 lines (resolved in M5 GREEN gate after counts are confirmed).

**Verification gate**:
- `grep -rln "manager-ddd\|manager-tdd\|builder-agent\|builder-skill\|builder-plugin\|expert-debug\|expert-testing" .claude/skills/ .claude/rules/ CLAUDE.md internal/template/templates/.claude/skills/ internal/template/templates/.claude/rules/ internal/template/templates/CLAUDE.md | grep -v examples.md` returns ONLY:
  - retired stub files themselves (`*-agent.md`, `*-skill.md`, etc.)
  - migration tables explicitly documenting the rewrite
- All other matches replaced.

---

### 2.5 M5 — REFACTOR + MX Tags + Completion Gate

**Owner**: manager-quality (verification), manager-spec (sign-off)
**Priority**: P0 (final gate before PR)

**Tasks**:

1. M5.1 — Apply `@MX` tags planned in M2-M4 to actual code lines:
   - Resolve all `@MX:TODO` planted in M4.
   - Verify `@MX:ANCHOR`, `@MX:WARN`, `@MX:NOTE` placements per mx_plan.

2. M5.2 — Update `spec.md` §10.1 destiny table (per OQ-1, OQ-5 resolutions):
   - Add row: `claude-code-guide` — meta — KEEP — `Post-R5 addition; Claude Code-related upstream investigator`.
   - Add row: `expert-mobile` — expert — KEEP — `Post-R5 addition; mobile native + cross-platform specialist`.
   - Add row: `manager-brain` — manager — KEEP — `Post-R5 addition; brain workflow orchestrator`.
   - Add row: `builder-platform` — builder — NEW — `New consolidated builder per REQ-ORC-001-003`.
   - Update HISTORY: add 0.1.1 entry noting these reconciliations.

3. M5.3 — Run all AC verifications from `acceptance.md`:
   - AC-ORC-001-01 to AC-ORC-001-12 (manual + scripted).

4. M5.4 — Run `make build && make test` (Go test suite must remain green; this SPEC does not touch Go code, so no regression expected).

5. M5.5 — Run `diff -r internal/template/templates/.claude/agents/moai/ .claude/agents/moai/` final check (must be empty).

6. M5.6 — Run `golangci-lint run` (must remain clean).

7. M5.7 — Stage commit and push to `feature/SPEC-V3R2-ORC-001-roster` branch.

**Files touched**:
- `.moai/specs/SPEC-V3R2-ORC-001/spec.md` (HISTORY + §10.1 amendments only)
- All M2-M4 files (final mx_tag pass)
- `progress.md` (final timestamp + audit-ready marker)

**mx_plan tags**:
- M5 resolves all prior `@MX:TODO` tags planted in M2-M4 (none remain after this milestone).

**Verification gate**:
- All 12 ACs PASS per acceptance.md (manual or scripted).
- `make build` exit code 0; `make test` exit code 0; `golangci-lint run` exit code 0.
- `diff -r` clean.
- No `@MX:TODO` remains anywhere in the modified file set.
- Final `git status` clean except for tracked changes.

---

## 3. Technical Approach Notes

### 3.1 Template-First Discipline

[HARD] Every file change starts in `internal/template/templates/`. After the
template-tree edit, `make build` regenerates
`internal/template/embedded.go`. Local `.claude/agents/moai/` is touched
ONLY by the build step (or a follow-up `cp` if the build skips agent dirs —
verify in M2.7). Never edit the local copy directly. Reason: CLAUDE.local.md
§2 Template-First Rule.

### 3.2 Stub Schema Consistency

All 5 new stubs use the SPEC-V3R3-RETIRED-DDD-001 stub schema verbatim.
Pattern (research.md §2 evidence):

```yaml
---
name: <retired-name>
description: |
  Retired (SPEC-V3R2-ORC-001) — use <replacement> with <param-hint>.
  This agent has been consolidated into <replacement>.
  See <replacement>.md for the active replacement.
retired: true
retired_replacement: <replacement>
retired_param_hint: "<param>=<value>"
tools: []
skills: []
---

# <retired-name> — Retired Agent

<one-line redirect + migration table>
```

### 3.3 Trigger-Union Preservation Test

REQ-ORC-001-009 + REQ-ORC-001-017 require the union of trigger keywords from
all retired source bodies to appear in the merged target. M2.8 verification
runs a Bash script:

```bash
# Pseudocode for trigger-union test (final form authored in M5)
for src in builder-agent builder-skill builder-plugin; do
  for lang in EN KO JA ZH; do
    extract_trigger_row "$src.md" "$lang" >> /tmp/expected.txt
  done
done
extract_trigger_row builder-platform.md ALL > /tmp/actual.txt
sort -u /tmp/expected.txt > /tmp/expected-deduped.txt
diff /tmp/expected-deduped.txt /tmp/actual.txt  # exit 0 = pass
```

The script is owned by SPEC-V3R2-ORC-002 (CI lint); ORC-001 references it
but does not implement.

### 3.4 Effort Field Deferral

This SPEC explicitly does NOT add `effort:` fields to any of the 17 final
agents. SPEC-V3R2-ORC-003 owns that responsibility. Plan-auditor will check
that `effort:` is consistently absent from this SPEC's diffs (or, if present
on agents predating this SPEC, unchanged).

### 3.5 Worktree Isolation Deferral

This SPEC does NOT add `isolation: worktree` to manager-cycle or
builder-platform. SPEC-V3R2-ORC-004 owns that. The new builder-platform
frontmatter has no `isolation` field; ORC-004 will add it.

---

## 4. Risks Reflected from spec.md §8

| Risk (from spec.md §8) | M-mapping | Mitigation in plan |
|------------------------|-----------|---------------------|
| Merging manager-ddd + manager-tdd drops a phase-specific rule | M1.1 verification step | Already done in prior SPEC; M1 re-verifies citations |
| builder-platform parameter dispatch loses an artifact variant | M2.1 + M2.8 | Explicit `artifact_type` enum in body; trigger-union test |
| Retired-agent references break legacy SPEC routing | M4 sweep | Stub redirects retain compatibility per Master §8 BC catalog |
| manager-project blocker-report over-triggers | M3.2 | Explicit allowlist; CLI references named in blocker template |
| expert-frontend Pencil split deferred → tech debt | (out of scope per §1.2) | Documented; revisit in WF-003 |
| Trigger-union dedup miscounts near-duplicates | M3.3 | Deterministic token-normalization script (M2.8 / M5.3) |

---

## 5. Out-of-Plan Concerns

These are documented but NOT implemented in this SPEC:

- ORC-002 CI lint rule for `AskUserQuestion` text in agent bodies
- ORC-003 effort matrix population
- ORC-004 worktree MUST upgrade
- MIG-001 v2→v3 user migrator (rewrites legacy SPEC bodies)
- WF-003 expert-frontend Pencil split (Phase 6)

---

End of plan.
