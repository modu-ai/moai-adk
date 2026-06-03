---
id: SPEC-CCSYNC-CLAUDEMD-001
title: "Claude Code instruction-layer doc sync (CLAUDE.md + rules templates)"
version: "0.1.1"
status: completed
created: 2026-06-02
updated: 2026-06-03
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/template/templates"
lifecycle: spec-anchored
tags: "ccsync, claude-md, template, doc-drift, agent-catalog"
---

# SPEC-CCSYNC-CLAUDEMD-001 — Claude Code instruction-layer doc sync

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-02 | manager-spec | Initial plan-phase authoring. Derived from a 26-document Claude Code official-docs gap audit. Three verified findings (H1 template version drift + archived-agent refs, H2 context-window threshold drift, H8 Agent Teams min-version + spawn-capability contradiction) plus one low-priority dead-reference bundle. |
| 0.1.1 | 2026-06-03 | manager-spec | plan-auditor PASS-WITH-DEBT defect remediation (AC command portability + mirror-test filename + REQ-011 AC coverage): corrected the byte-parity mirror-test filename to the real `rule_template_mirror_test.go` (verified present); reshaped the multi-file `grep -c` ACs (008-015) to mechanically correct true-sum `grep -rh … | wc -l` or per-file forms; added AC-CCSYNC-017 to gate REQ-CCSYNC-011 (co-located `v2.1.50` removal in agent-authoring.md + mirror); aligned plan.md §H to describe the shared line ~212 consistently with the sibling. |

## A. Context and Motivation

The instruction layer that every `moai init` / `moai update` user receives is the
**template** copy of `CLAUDE.md` and the `.claude/rules/moai/**` rule files under
`internal/template/templates/`. A 26-document audit of the Claude Code official
documentation against this instruction layer surfaced three categories of drift
where the deployed template either (a) lags the maintainer's authoritative dev-root
copy, (b) contradicts a canonical single-source-of-truth (SSOT) rule file, or
(c) contradicts the official Claude Code documentation.

This SPEC scopes the WHAT and WHY of the remediation. The HOW (the actual edits,
`make build`, and embedded-file regeneration) is run-phase work owned by
manager-develop. This SPEC describes observable behaviors and mechanically
checkable acceptance criteria so the run phase has an unambiguous target.

Baseline: HEAD `5042e309c`. A parallel session has already renamed the
sync-phase auditor agent `evaluator-active` → `sync-auditor` across the agent
catalog. All requirements below reference the current name **`sync-auditor`**.

### Why this matters

- **Correctness for downstream users**: The stale template ships self-contradictory
  content — it lists agents (`expert-backend`, `manager-strategy`, `manager-quality`,
  `expert-devops`, `expert-frontend`) as the recommended SPEC-execution chain while
  the same template marks those exact agents as rejected-at-spawn. A new user
  following the stale chain would attempt to spawn an archived agent.
- **Operational safety**: The context-window threshold "1M = 75%" contradicts the
  canonical SSOT `.claude/rules/moai/workflow/context-window-management.md`
  (1M = 50%). A user (or orchestrator) acting on 75% would push past the 50%
  handoff boundary and risk an SSE stream stall during a long agent run.
- **Version accuracy**: The Agent Teams minimum-version string (`v2.1.50`) is wrong
  per the official Claude Code agent-teams documentation (minimum `v2.1.32`), and
  `agent-authoring.md` asserts a spawn capability ("teammates CAN spawn other
  teammates") that directly contradicts the official limitation and a correct
  statement elsewhere in the same file.

## B. Verified Findings (run-phase remediation scope)

### B.1 Finding H1 — template CLAUDE.md is a minor version behind and ships archived-agent references

- `internal/template/templates/CLAUDE.md` is `Version: 14.1.0`; the dev-root
  `CLAUDE.md` is `Version: 14.2.0`. The template is the file every `moai init` /
  `moai update` user receives.
- The stale template contains self-contradictory content (verified line numbers at
  HEAD `5042e309c`):
  - §2 Phase 3 Execute (line 61): `- "Use the expert-backend subagent to develop the API"` (archived agent).
  - §5 Agent Chain for SPEC Execution (lines 155-162): a 6-phase chain naming
    archived `manager-strategy` (line 158), `expert-backend` (line 159),
    `expert-frontend` (line 160), `manager-quality` (line 161).
  - §11 Error Recovery (lines 383, 386): routes to archived `manager-quality`
    (line 383) and `expert-devops` (line 386).
- All five named agents are listed in the SAME template (line 127) as archived and
  MUST-NOT-be-spawned. The template is therefore internally self-contradictory.
- The dev-root copy already carries the migrated, correct content: a 4-phase
  retained-agent chain (manager-spec → plan-auditor → manager-develop →
  manager-docs/sync-auditor → optional manager-git) with per-spawn
  `Agent(general-purpose)` migration cross-references.

### B.2 Finding H2 — stale context-window threshold "1M = 75%" contradicts the canonical SSOT (1M = 50%)

Four locations, verified at HEAD `5042e309c`:

- `CLAUDE.md` §16 (line 533): "context window thresholds (1M = 75%, 200K = 90%)".
- `internal/template/templates/CLAUDE.md` §16 (line 531): identical "1M = 75%".
- `.claude/rules/moai/core/settings-management.md` (line 214): "SSE streams stall
  when prompts approach 75% of the window".
- `internal/template/templates/.claude/rules/moai/core/settings-management.md`
  (line 214): identical "75% of the window".

The canonical SSOT `.claude/rules/moai/workflow/context-window-management.md` table
specifies 1M = **50%**, 200K = 90% (verified), and `session-handoff.md` Trigger #1
co-anchors to 1M = 50%. The four drifted references contradict the SSOT.

### B.3 Finding H8 — Agent Teams minimum version wrong (v2.1.50 → v2.1.32) + agent-authoring spawn-capability contradiction

- The official Claude Code agent-teams documentation states the minimum version is
  v2.1.32. `CLAUDE.md` §15 (line 463) and `internal/template/templates/CLAUDE.md`
  §15 (line 461) both say `Claude Code v2.1.50 or later`.
- `.claude/rules/moai/development/agent-authoring.md` line 212 (and its template
  mirror at the same line) asserts: "Agent Teams teammates CAN spawn other
  teammates using Agent() with the team_name parameter, v2.1.50+". This:
  - Contradicts the official Claude Code limitation that teammates cannot spawn
    their own teammates.
  - Contradicts line 92 of the SAME file: "subagents cannot spawn other subagents".
  - Carries the same wrong `v2.1.50+` version string flagged above.
- The run-phase remediation reconciles line 212 to match the official constraint
  (no teammate-spawns-teammate capability claim) and corrects the version string.

### B.4 Low-priority bundle — dead changelog reference

The CLAUDE.md changelog note (dev-root line 626, template line 626) references
`.claude/rules/moai/workflow/progressive-disclosure.md`. That file does not exist
(verified: `ls` returns "No such file or directory"). The canonical file for
progressive-disclosure guidance is `.claude/rules/moai/development/skill-authoring.md`
(verified present). The dead reference appears in BOTH the dev-root and template
CLAUDE.md copies. This bundle is in-scope as a minor item: fix or remove the dead
reference in both copies.

## C. Requirements (GEARS)

### Finding H1 requirements

- **REQ-CCSYNC-001** (Ubiquitous): The template CLAUDE.md `Version:` line shall be
  at version 14.2.0 or higher, matching or exceeding the dev-root CLAUDE.md version.
- **REQ-CCSYNC-002** (Ubiquitous): The template CLAUDE.md §2 Phase 3 Execute
  examples shall reference only retained agents (no `expert-backend` or any of the
  12 archived agent names) as invocation examples.
- **REQ-CCSYNC-003** (Ubiquitous): The template CLAUDE.md §5 Agent Chain for SPEC
  Execution shall describe the retained-agent chain (manager-spec → plan-auditor →
  manager-develop → manager-docs/sync-auditor → optional manager-git) and shall
  contain none of the 12 archived agent names as chain phases.
- **REQ-CCSYNC-004** (Ubiquitous): The template CLAUDE.md §11 Error Recovery shall
  route errors to retained agents or per-spawn `Agent(general-purpose)` patterns and
  shall contain none of the 12 archived agent names.
- **REQ-CCSYNC-005** (Where template-neutrality applies): Where ported content is
  written into the template, the orchestrator shall ensure the ported text contains
  no moai-adk-internal SPEC IDs, REQ tokens, AC tokens, audit citations, internal
  dates, commit SHAs, archive paths, or memory-hash references (per CLAUDE.local.md
  §25 Template Internal-Content Isolation).
- **REQ-CCSYNC-006** (When the template source changes): When any file under
  `internal/template/templates/` is modified, the build shall regenerate
  `internal/template/embedded.go` via `make build` so the embedded copy matches the
  source.

### Finding H2 requirements

- **REQ-CCSYNC-007** (Ubiquitous): The dev-root CLAUDE.md §16 and template CLAUDE.md
  §16 shall state the context-window threshold as 1M = 50% (not 1M = 75%), matching
  the canonical SSOT context-window-management.md table.
- **REQ-CCSYNC-008** (Ubiquitous): The dev-root settings-management.md and its
  template mirror shall reference the model-specific context-window threshold
  (1M = 50%, 200K = 90%) rather than a flat 75%.

### Finding H8 requirements

- **REQ-CCSYNC-009** (Ubiquitous): The dev-root CLAUDE.md §15 and template CLAUDE.md
  §15 shall state the Agent Teams minimum version as v2.1.32 (not v2.1.50).
- **REQ-CCSYNC-010** (Ubiquitous): The dev-root agent-authoring.md and its template
  mirror shall not assert that Agent Teams teammates can spawn other teammates; the
  reconciled wording shall match the official Claude Code limitation and be
  internally consistent with the same file's "subagents cannot spawn other
  subagents" statement.
- **REQ-CCSYNC-011** (When agent-authoring.md is edited for REQ-CCSYNC-010): When
  the spawn-capability wording is reconciled, the orchestrator shall also remove or
  correct the co-located `v2.1.50+` version string in that same sentence.

### Low-priority bundle requirement

- **REQ-CCSYNC-012** (Ubiquitous): The dev-root CLAUDE.md and template CLAUDE.md
  changelog notes shall not reference the nonexistent
  `.claude/rules/moai/workflow/progressive-disclosure.md`; the reference shall be
  corrected to `.claude/rules/moai/development/skill-authoring.md` or removed.

### Mirror-parity requirement (cross-cutting)

- **REQ-CCSYNC-013** (Where a rule file exists in both `.claude/rules/.../*.md` and
  its `internal/template/templates/` mirror): Where a remediated file has a template
  mirror (settings-management.md, agent-authoring.md), both copies shall be edited
  in the same commit so byte-parity invariants
  (`internal/template/rule_template_mirror_test.go`) hold.

## D. Exclusions (What NOT to Build)

- **EXC-1 — docs-site agent-guide version fix is OUT OF SCOPE.** The docs-site
  `agent-guide.md` Agent Teams version string (en/ko/zh locales) belongs to the
  sibling docs-site SPEC, not this SPEC. This SPEC touches only CLAUDE.md (both
  copies) and `.claude/rules/moai/**` (dev-root + template mirror).
- **EXC-2 — No restructuring of the agent catalog itself.** This SPEC does not add,
  remove, rename, or re-scope any agent. The `evaluator-active` → `sync-auditor`
  rename was completed by a parallel session at HEAD `5042e309c`; this SPEC consumes
  that name, it does not perform the rename.
- **EXC-3 — No changes to the canonical SSOT threshold values.** This SPEC aligns
  drifted copies TO the canonical 1M = 50% / 200K = 90% values; it does not modify
  `context-window-management.md` or `session-handoff.md` (which are already correct).
- **EXC-4 — No broad content rewrite.** Only the specific drifted lines/sections
  named in Findings H1/H2/H8 + the dead-reference bundle are in scope. No drive-by
  edits to unrelated CLAUDE.md sections or unrelated rule files.
- **EXC-5 — No version-bump release process.** Bumping the template CLAUDE.md
  `Version:` line is in scope (REQ-CCSYNC-001); the broader release workflow
  (git tags, CHANGELOG release entry, system.yaml) is NOT.
- **EXC-6 — This SPEC does not perform the edits.** Plan-phase authoring only.
  All file modifications are run-phase work owned by manager-develop.

## E. Constraints

- **CON-1**: Template-First Rule (CLAUDE.local.md §2). Template source under
  `internal/template/templates/` is edited; `make build` regenerates `embedded.go`.
  `embedded.go` is never hand-edited.
- **CON-2**: Template Internal-Content Isolation (CLAUDE.local.md §25). Ported
  template content must stay generic — strip internal SPEC IDs / REQ tokens / audit
  citations / internal dates / commit SHAs.
- **CON-3**: Mirror parity (`internal/template/rule_template_mirror_test.go`). Files with
  both a `.claude/rules` copy and a `templates/` mirror must be edited together.
- **CON-4**: Hybrid Trunk Tier M (CLAUDE.local.md §23.9). main-direct, no PR, unless
  the user explicitly escalates with `--pr`.
- **CON-5**: Shared-file sequencing. `agent-authoring.md` is also a run-phase target
  of the sibling SPEC `SPEC-CCSYNC-TOOLCAT-001` (TodoWrite→Task* recommendation).
  This SPEC and the sibling MUST be sequenced (this SPEC first; sibling rebases) to
  avoid an intra-batch edit collision. See plan.md §F and §H.

## F. Verification Strategy

Every requirement maps to a mechanically checkable acceptance criterion in
acceptance.md: version-string greps, archived-agent-name absence greps, threshold
greps returning 0 for "75%", dead-reference greps returning 0, `make build` clean,
and `go test ./internal/template/...` green (neutrality audit + internal-content
leak guard + mirror-drift). See acceptance.md for the full Given-When-Then matrix.

## G. Cross-References

- Baseline HEAD: `5042e309c` (sync-auditor rename).
- Sibling SPEC sharing `agent-authoring.md`: `SPEC-CCSYNC-TOOLCAT-001` (see CON-5).
- Canonical threshold SSOT: `.claude/rules/moai/workflow/context-window-management.md`.
- Template isolation doctrine: CLAUDE.local.md §25.
- Template-neutrality doctrine: CLAUDE.local.md §15.
- Tier-based PR routing: CLAUDE.local.md §23.9.
