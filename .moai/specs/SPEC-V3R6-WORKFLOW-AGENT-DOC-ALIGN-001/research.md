# Research — SPEC-V3R6-WORKFLOW-AGENT-DOC-ALIGN-001

## §A — Investigation Method

Three parallel read-only audits of the `/moai` skill + workflow + agent assets (2026-06-20), cross-validated against the canonical migration table and the working-tree state. This research.md records the evidence-grounded findings (grep/test output observed at plan-phase) and the canonical sources that constrain the remediation. Per `.claude/rules/moai/core/verification-claim-integrity.md`, every defect claim below is backed by an observed command output, not a text-pattern inference.

## §B — Verified Evidence (plan-phase, against working tree at HEAD)

### B.1 — Archived-agent references (Group 1) — re-baselined (D3)

**Re-baseline (iter-1 D3 + iter-2 D-NEW-1 corrections)**: the raw 12-name alternation (`ARCHIVED-12`, including `researcher`) returns **139 lines across 38 files** (NOT the earlier 27/125, and NOT the iter-1 "32 files" — the file count was re-verified at iter-2 D-NEW-1). Of the 38 matching files, 36 carry an archived-proper name and 2 carry `researcher` only. The earlier 125 figure had silently excluded the 14 `researcher` false-positives while REQ-WADA-001 as first written INCLUDED `researcher` in the purge set — a contradiction. Re-derived live:

- `grep -rnE 'ARCHIVED-12' .claude/skills/moai/ .claude/agents/moai/` → **139 lines**; `grep -rlnE 'ARCHIVED-12' ... | wc -l` → **38 files** (36 archived-proper + 2 researcher-only).
- `grep -rcn 'researcher' .claude/skills/moai/ .claude/agents/moai/` (summed) → **14 lines** = live role_profile references (KEEP, D2 — see B.8).
- `grep -rnE 'ARCHIVED-11' ...` (researcher EXCLUDED) → **139 − 14 = 125 archived-PROPER lines** across **36 files** = the true purge count.
- `grep -rln 'ARCHIVED-11' .claude/agents/moai/` → **6 files**: `builder-harness.md`, `manager-develop.md`, `manager-docs.md`, `manager-git.md`, `manager-spec.md`, `sync-auditor.md`. **(D1: these are all under `.claude/agents/moai/`; there is NO `.claude/agents/builder/` — `ls .claude/agents/` → `harness/ local/ moai/`.)**
- **`references/{reference,anti-patterns,mx-tag}.md` (D3 newly added)**: `grep -rln 'ARCHIVED-11' .claude/skills/moai/references/` → all 3 carry archived refs (incl. frontmatter `agents:` lists). Added to M2 scope.
- The audited skill files extend beyond the prompt's enumerated list — the live tree additionally contains: `brain.md`, `gate.md`, `mx.md`, `project.md`, `plan/clarity-interview.md`, `project/doc-generation.md`, and the 3 `references/*.md` above. Folded into M2 scope.
- `feedback.md` confirmed wholly broken: sole executor is `manager-quality` (lines 4, 24, 30, 68, 70, 104, 141, 152). The frontmatter `agents: ["manager-quality"]` (line 24) and the `gh issue create` body (line 104) show the actual operation is GitHub-issue creation via `gh` — which the migration table maps to orchestrator-direct (no retained agent owns issue creation).

### B.2 — Invalid team subagent types (Group 2)

- `grep -rn 'team-reader\|team-validator' .claude/skills/moai/team/`:
  - `glm.md:171-172` (table rows)
  - `debug.md:56,59,62` (×3 `team-reader`)
  - `review.md:24,60,72,84,96` (`team-validator` ×4 + agents-list)
  - `plan.md:24,55,59,75,88,226` (`team-reader` ×3 + agents-list + version footer)
- `team/run.md` is the CORRECT reference pattern: `subagent_type: "general-purpose"` at lines 31, 42, 173, 182, 206, 227.

### B.3 — Broken cross-reference (Group 2)

- `worktree-state-guard.md:19` and `spec-workflow.md:23` both cross-reference `worktree-integration.md § Terminology Glossary` for L1/L2/L3 definitions.
- `grep -n 'Terminology Glossary' worktree-integration.md` → **0 matches**. The section is absent → broken cross-reference confirmed.

### B.4 — Stale ground-truth (Group 3)

- `team/glm.md` lines 148-150 (`glm-5.1` / `glm-4.7` / `glm-4.5-air`) and 168-172 (`glm-5.1 / glm-4.7` table + `glm-4.7-flashx`). Current ground truth per memory `project_glm52_update_completed`: `glm-5.2[1m]`.
- **Retired-term path disambiguation (D8 correction)**: the two retired terms are in **two DIFFERENT files** —
  - "Round split" → `.claude/skills/moai/team/run.md:69` (`grep -n 'Round split' team/run.md` → line 69)
  - "sprint contract" → `.claude/skills/moai/workflows/run.md:80` (`grep -n 'sprint contract' workflows/run.md` → line 80)
  Both retired per `sprint-round-naming.md` (Round folded into Milestone; Sprint→Epic). The earlier draft conflated both onto "run.md" — they are `team/run.md` (Round split) and `workflows/run.md` (sprint contract) respectively.
- `SKILL.md` body footer (lines 29-30): `Version: 2.6.0` / `Last Updated: 2026-02-25` — predates the 8-agent consolidation (SPEC-V3R6-AGENT-TEAM-REBUILD-001, 2026-05-25). NOTE: `SKILL.md` frontmatter has no `version` field; the version is the body footer.

### B.5 — Orphaned RED test (Group 4)

- `wc -l .claude/skills/moai/workflows/run.md` → **246** (both live and template). Ceiling 200.
- `go test ./internal/skills/ -run TestEntryRouterLOCCeiling` → FAIL (`run.md has 246 LOC (ceiling: 200)`), per audit. Added by SPEC-V3R6-ORCH-IGGDA-001 entry-router (now closed), left RED.

### B.6 — Template mirror (cross-cutting) — refined (D4/D5)

- Most in-scope skill files + 6 agent bodies + `worktree-integration.md` have an `internal/template/templates/.claude/` mirror.
- **Baseline divergence (D4)**: `diff .claude/agents/moai/<f> internal/template/templates/.claude/agents/moai/<f>` for `manager-spec.md` / `manager-docs.md` / `manager-git.md` → **ALL 3 DIVERGED at baseline**. The divergence is the 3-phase-vs-legacy `§E.5`/Mx lifecycle content (owned by SPEC-V3R6-LIFECYCLE-REDESIGN-001). Whole-file byte parity is UNSATISFIABLE for these 3 regardless of this SPEC's work → REQ-WADA-016 is a CHANGED-LINE mirror requirement, not whole-file (see design.md §E).
- **Dev-only no-mirror (D4/D5)**: `internal/template/templates/.claude/skills/moai/workflows/release.md` is ABSENT. `release.md` (+ `github.md`, `release-update.md`) are in `devOnlyLocalFiles` (`internal/skills/workflow_split_test.go:184`) and are excluded from `TestTemplateMirrorParity` by the test itself. `release.md` archived-purge applies to the LIVE file only.
- `.claude/agents/local/github-specialist.md` exists (dev-only, NOT mirrored) → confirmed OUT OF SCOPE.

### B.8 — `researcher` is a LIVE role_profile, NOT an archived agent (D2)

- `grep -n 'role_profiles\|researcher' .moai/config/sections/workflow.yaml` → `researcher` is defined at **workflow.yaml:68** under `role_profiles:`, alongside `analyst`, `architect`, `implementer`, `tester`, `designer`, `reviewer` (the 7 live dynamic-team spawn profiles).
- Per `.claude/rules/moai/NOTICE.md` (Anthropic 2026 Alignment, Archive Summary): the archived `researcher.md` MoAI agent file was **"originally absent — never present as a MoAI file in this repo"**. There is NO `researcher` agent file to purge.
- Therefore ALL 14 `researcher` matches are role_profile references in `team/plan.md`, `team/run.md`, and the spec-workflow Team-Mode tables — they MUST be KEPT. Purging them would break team-mode plan-phase spawn. This is the D2 correction: `researcher` is removed from the archived-agent set; the over-purge guard AC-WADA-001a asserts the baseline count of 14 is preserved.

### B.9 — `claude-code-guide` has 0 live occurrences (D6)

- `grep -rn 'claude-code-guide' .claude/skills/moai/ .claude/agents/moai/` → **0 matches**. No purge action required. The §C.3 built-in disambiguation (valid Anthropic built-in vs archived MoAI file) is preserved as a forward-guard only. Demoted from a per-occurrence remediation to a 0-occurrence note in spec.md REQ-WADA-001b and design.md §B.

### B.7 — Non-reproductions (honest gaps)

- `grep -rn 'release-update' .claude/skills/moai/` → **0 matches**. The prompt's "SKILL.md:258 dispatches non-existent workflows/release-update.md" does NOT reproduce at this tree state; line 258 is harness-namespace content. Recorded as OQ-1; AC-WADA-008 is conditional/vacuous.
- No command frontmatter (`triggers.agents:`) in `.claude/commands/` references archived agents → the Group 1 "command frontmatter" sub-item is not present; only skill frontmatter `agents:` lists carry them (e.g. feedback.md:24).

## §C — Canonical Sources Constraining the Remediation

| Source | Constraint |
|--------|------------|
| `archived-agent-rejection.md §C` | migration table = the ONLY authority for replacements (11 archived-proper agents; `researcher` row note: never an agent file) |
| `archived-agent-rejection.md §C.3` | `claude-code-guide` built-in vs archived MoAI file disambiguation (0 live occurrences — forward-guard only) |
| `.moai/config/sections/workflow.yaml` `role_profiles:` (line 68) | `researcher`/`analyst`/`architect`/`implementer`/`tester`/`designer`/`reviewer` are LIVE profiles — KEEP (D2) |
| `.claude/rules/moai/NOTICE.md` | researcher "originally absent" — never an agent file (D2 evidence) |
| `spec-frontmatter-schema.md` § Ownership Matrix | retained-agent owner names (no archived names as owners) |
| `sprint-round-naming.md` | Epic/Milestone taxonomy (Round/Sprint retired) |
| `internal/skills/workflow_split_test.go:154/159/184` | `TestEntryRouterLOCCeiling` (G4) + `TestTemplateMirrorParity` + `devOnlyLocalFiles` (D4/D5) |
| `CLAUDE.local.md §2/§24/§25` | Template-First, namespace boundary, template neutrality |
| `verification-claim-integrity.md` | every defect claim needs observed evidence (this research.md complies) |

## §D — Why these defects matter (impact)

- **Group 1**: The orchestrator REJECTS archived spawn targets (`ARCHIVED_AGENT_REJECTED`). Files instructing such spawns are dead or misrouting — `feedback.md` is functionally broken (its only executor is archived). ~125 archived-proper stale references (of 139 raw, minus 14 researcher role_profile KEEPs) also mislead future authors and the catalog-consistency audits.
- **Group 2**: `team-reader`/`team-validator` spawns fail (non-existent subagent types). Broken `§ Terminology Glossary` cross-ref means readers seeking L1/L2/L3 definitions hit a dead link.
- **Group 3**: Stale GLM model names mislead CG-mode users; retired taxonomy terms drift from the canonical naming SSOT; stale SKILL.md version misrepresents the asset's currency.
- **Group 4**: The RED test blocks `go test ./...` GREEN on main — a CI-visible regression left by a closed SPEC.

## §E — Approach Validation (DDD fit)

This is brownfield consistency work on existing assets. `cycle_type=ddd` (ANALYZE-PRESERVE-IMPROVE) fits: ANALYZE = the M1 occurrence inventory; PRESERVE = routing-intent map (swap the agent, keep the intent); IMPROVE = the dual-tree edits. No new behavior; the "test" is the grep/test-verifiable end state in the AC matrix. TDD (RED-GREEN-REFACTOR) does not fit because there is no new functionality to specify test-first — except Group 4, where the test already exists (RED) and the fix makes it GREEN, which is itself a GREEN-restoration consistent with DDD's continuous-validation loop.

## §F — Open Questions (carried to plan/design)

- OQ-1: `release-update` not reproduced — run-phase re-grep decides AC-WADA-008.
- OQ-2: glossary ADD vs repoint — design.md §C.1 recommends ADD.
- OQ-3: run.md trim compress vs extract — design.md §D.2 recommends compress-in-place.
