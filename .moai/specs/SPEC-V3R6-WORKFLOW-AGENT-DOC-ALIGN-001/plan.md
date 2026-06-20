# Implementation Plan — SPEC-V3R6-WORKFLOW-AGENT-DOC-ALIGN-001

## §A — Context

Tier L documentation/agent-consistency cleanup. **38 files** match the archived-name alternation (36 carry archived-proper names + 2 carry `researcher` only) with **139 raw archived-name matches**, of which **14 are legitimate `researcher` role_profile references (KEEP)** and **~125 are archived-proper occurrences** across 36 files to purge (skill files under `.claude/skills/moai/` incl. `references/{reference,anti-patterns,mx-tag}.md` + 6 agent bodies under `.claude/agents/moai/`); 4 team files use 2 non-existent subagent types; 1 rule cross-reference is broken; 3 stale ground-truth values; 1 orphaned RED test. Most fixes mirror to `internal/template/templates/.claude/` (EXCEPT the dev-only `release.md`/`github.md`/`release-update.md` which have no mirror, and the pre-existing baseline divergence in `manager-spec.md`/`manager-docs.md`/`manager-git.md` which is out of scope). This is a high-file-count SPEC (38 files match; ~36 archived-proper files touched) — SSE-stall mitigation is handled by per-Milestone delegation (each Milestone is a bounded file set), per `.claude/rules/moai/development/sprint-round-naming.md`.

> **Directory truth (D1)**: there is NO `.claude/agents/builder/`. All 6 agent bodies (incl. `builder-harness.md`) live under `.claude/agents/moai/`.
> **role_profile carve-out (D2)**: `researcher`/`analyst`/`architect`/`implementer`/`tester`/`designer`/`reviewer` are LIVE workflow.yaml `role_profiles:` — NOT archived agents. The 14 `researcher` matches are KEPT untouched. The archived-proper purge set is the 11 actually-archived agents.

The development methodology is `cycle_type=ddd` (ANALYZE-PRESERVE-IMPROVE): these are existing files whose behavior (routing intent) must be PRESERVED while the archived spawn TARGETS are replaced. There is no new functionality; the characterization is "what does each file currently route, and to whom" — and the IMPROVE step swaps the archived target for its canonical replacement without changing the routed intent.

## §B — Known Issues / Constraints from plan-phase

1. **`release-update` dispatch (REQ-WADA-008) not reproduced**: plan-phase grep found NO `release-update` token in the current `SKILL.md` working tree. The audit's "SKILL.md:258" finding does not reproduce at this tree state (line 258 is harness-namespace content). REQ-WADA-008 is therefore conditional — run-phase MUST re-grep; if the token is absent, AC-WADA-008 passes vacuously (no broken dispatch exists) and the milestone records "not reproduced, no change needed".
2. **SKILL.md has no frontmatter `version` field** — the version is a body footer (`Version: 2.6.0`). REQ-WADA-013 targets the body footer, not frontmatter.
3. **~125 archived-proper occurrences is a count, not 125 distinct edits** — many are clustered (multiple lines per file). The unit of work is the file, not the occurrence. The 14 `researcher` role_profile matches are NOT in this count (they are KEPT).
4. **Baseline template divergence (D4)**: `manager-spec.md`/`manager-docs.md`/`manager-git.md` are ALREADY diverged live↔template at baseline (pre-existing §E.5/Mx lifecycle content, owned by SPEC-V3R6-LIFECYCLE-REDESIGN-001). Whole-file byte parity is unsatisfiable for these 3 — run-phase verifies only that THIS SPEC's archived-purge edit landed in both trees (changed-line mirror, REQ-WADA-016).
5. **Dev-only no-mirror trio (D4/D5)**: `release.md`/`github.md`/`release-update.md` are in `devOnlyLocalFiles` (workflow_split_test.go:184) — no template mirror by design. `release.md` archived-purge applies to the LIVE file only; excluded from `TestTemplateMirrorParity` and AC-WADA-016.

## §C — Pre-flight (run-phase entry checks)

Before M1, run-phase MUST re-execute the §A.2 evidence baseline greps to confirm the working tree has not drifted since plan-phase (multi-session race awareness per `feedback_shared_main_orphan_race`). If counts differ materially, surface to orchestrator before proceeding.

## §D — Constraints

- [HARD] Template-First: every live `.claude/` edit to a MIRRORED file is mirrored to `internal/template/templates/.claude/` (changed-line; EXCEPT dev-only no-mirror trio + baseline-diverged trio per §B).
- [HARD] Replacement prose follows `archived-agent-rejection.md §C` migration table verbatim (no improvised replacements).
- [HARD] Legitimate references TO `archived-agent-rejection.md` are NOT removed.
- [HARD] `researcher`/`analyst`/`architect`/`implementer`/`tester`/`designer`/`reviewer` role_profile references are NOT removed (live workflow.yaml profiles — over-purge guard AC-WADA-001a). The archived-proper purge set is exactly the 11 actually-archived names; `researcher` is NOT among them.
- [HARD] No Go production code changes; `TestEntryRouterLOCCeiling` itself is NOT modified.
- [HARD] Template neutrality (§25): no internal SPEC IDs / REQ tokens / audit citations / dates / SHAs leak into template replacement text.

## §E — Self-Verification

Run-phase self-verification deliverables (E1-E7 per manager-develop template) recorded in `progress.md §E.2`/`§E.3`. The canonical verification batch:

```bash
# G1 zero archived-PROPER spawn targets (11 names — researcher EXCLUDED), both trees, excl rejection rule.
# Bidirectional: loop BOTH trees so a template-only residual is caught (D4).
ARCHIVED11='manager-strategy|manager-quality|manager-brain|manager-project|claude-code-guide|expert-backend|expert-frontend|expert-security|expert-devops|expert-performance|expert-refactoring'
for TREE in .claude internal/template/templates/.claude; do
  grep -rnE "$ARCHIVED11" "$TREE/skills/moai/" "$TREE/agents/moai/" 2>/dev/null \
    | grep -v 'archived-agent-rejection' | grep -v '^[^:]*:[0-9]*:[ \t]*#'
done
# (No .claude/agents/builder/ — builder-harness.md is under .claude/agents/moai/.)
# researcher KEEP guard: baseline count preserved
grep -rcn 'researcher' .claude/skills/moai/ .claude/agents/moai/ | awk -F: '{s+=$2} END {print s}'  # expect 14
# G4 GREEN + mirror set-comparison
go test ./internal/skills/ -run TestEntryRouterLOCCeiling
go test ./internal/skills/ -run TestTemplateMirrorParity
# full suite
go test ./...
# template neutrality
go test ./internal/template/... -run TestTemplateNeutralityAudit
```

## §F — Milestones (priority-based, no time estimates)

### M1 — ANALYZE: occurrence inventory + replacement map (priority: High)

- Re-run §A.2 evidence greps (pre-flight §C) — confirm 38 files match / 139 raw / 14 researcher-KEEP / ~125 archived-proper across 36 files.
- For each archived-PROPER occurrence (11 names, NOT researcher), classify the archived agent → canonical replacement per migration table §C, and classify the routing intent (spawn target vs documentation pointer vs frontmatter list vs **researcher role_profile KEEP**).
- Mark every `researcher` match as KEEP (role_profile) — these are NOT in the replacement map.
- Produce the replacement map (which becomes the PRESERVE characterization).
- Confirm template mirror status for every file: MIRRORED (apply both trees), DEV-ONLY no-mirror (`release.md`/`github.md`/`release-update.md`, live only), or BASELINE-DIVERGED (`manager-spec/docs/git.md`, changed-line mirror only).
- Output: inventory table in `progress.md` (file → archived agent(s) → replacement → intent class → mirror status).

### M2 — Group 1a: workflow skill files archived-agent purge (priority: High)

- Remediate the workflow files (`workflows/*.md` + `workflows/run/*.md` + `workflows/sync/*.md` + `workflows/plan/*.md` + `workflows/project/*.md` + **`references/*.md`**): feedback, fix, loop, clean, review, security, e2e, design, moai, release, brain, gate, mx, project, run/{phase-execution, mode-orchestration, task-decomposition}, sync/{quality-gates-quality, quality-gates-context, doc-execution, delivery}, plan/clarity-interview, project/doc-generation, **references/reference.md, references/anti-patterns.md, references/mx-tag.md** (D7 — these 3 reference files carry archived refs incl. frontmatter `agents:` lists).
- Apply migration-table replacements (REQ-WADA-002/003/004/005). `feedback.md` → orchestrator-direct `gh` CLI (REQ-WADA-005).
- Mirror each to template EXCEPT `release.md` (dev-only no-mirror — live edit only).
- Bounded file set (SSE-stall mitigation).

### M3 — Group 1b: team skill files + agent bodies + frontmatter (priority: High)

- Remediate `team/run.md` (verify already-correct general-purpose pattern preserved + KEEP its `researcher` role_profile refs), and the 6 agent bodies under `.claude/agents/moai/` (REQ-WADA-006: replace spawn-instruction refs, keep migration-table pointers). NOTE: for `manager-spec/docs/git.md`, apply changed-line mirror only (baseline-diverged — do NOT attempt whole-file reconciliation).
- Remove/replace archived entries in skill frontmatter `agents:` lists (REQ-WADA-007), e.g. `feedback.md:24 agents: ["manager-quality"]`. KEEP any `researcher` entry (role_profile — AC-WADA-001a).
- Mirror each to template (the agent bodies all have mirrors; apply changed-line mirror to the 3 baseline-diverged ones).

### M4 — Group 2: broken cross-refs + invalid dispatch (priority: High)

- REQ-WADA-009: `team/plan.md`, `team/review.md`, `team/debug.md`, `team/glm.md` → `general-purpose` subagent_type.
- REQ-WADA-010: per design.md decision, EITHER add `§ Terminology Glossary` (L1/L2/L3 definitions) to `worktree-integration.md` OR repoint the 2 cross-references.
- REQ-WADA-008: re-grep `release-update`; fix dispatch if present, else record "not reproduced".
- Mirror each to template.

### M5 — Group 3: stale ground-truth (priority: Medium)

- REQ-WADA-011: `team/glm.md` GLM model names → `glm-5.2[1m]`.
- REQ-WADA-012: retired terms (Round split, sprint contract) → canonical Epic/Milestone/harness phrasing.
- REQ-WADA-013: `SKILL.md` body version footer bump.
- Mirror each to template.

### M6 — Group 4: run.md LOC trim + verification + template parity (priority: High)

- REQ-WADA-014: trim `run.md` below 200 LOC preserving entry-router routing logic (compress prose / move detail to a sub-file `workflows/run/` if needed; mirror the sub-file too). NOTE: M2 already edits run.md for archived purge — coordinate so the trim is the final run.md edit.
- REQ-WADA-015: `go test ./internal/skills/ -run TestEntryRouterLOCCeiling` → exit 0.
- REQ-WADA-016: changed-line mirror verification (NOT whole-file diff) — run the bidirectional §E batch + `go test ./internal/skills/ -run TestTemplateMirrorParity`. Whole-file `diff` is EXPECTED to differ for `manager-spec/docs/git.md` (baseline divergence, out of scope); dev-only trio excluded.
- Full `go test ./...` GREEN; `golangci-lint run` 0 issues (no Go changes expected, but verify no regression); `go test ./internal/template/... -run TestTemplateNeutralityAudit` PASS.

## §G — Anti-Patterns / Open Questions

- **OQ-1**: `release-update` dispatch not reproduced at plan-phase. Resolution: run-phase re-grep; AC-WADA-008 vacuously passes if absent.
- **OQ-2 (design.md decides)**: REQ-WADA-010 direction — add glossary section vs repoint refs. design.md recommends ADD (the cross-references expect a canonical L1/L2/L3 definition location, and `worktree-integration.md` is the natural home).
- **OQ-3**: `run.md` trim strategy — compress in-place vs extract to sub-file. design.md recommends compress-in-place first (the 46-LOC overage is prose-heavy); extract only if compression loses routing clarity.
- **AP**: Do NOT batch-edit with sed/regex across the ~125 archived-proper occurrences — each replacement is context-sensitive (spawn vs pointer vs frontmatter vs researcher-KEEP). Use Edit per occurrence-cluster with Read-before-Edit.
- **AP (D2)**: Do NOT include `researcher` in any blanket find-replace — it is a live role_profile. A blanket `s/researcher//` would break team-mode plan-phase spawn. Confirm baseline count 14 is preserved (AC-WADA-001a).
- **AP (D1)**: Do NOT reference `.claude/agents/builder/` — it does not exist. Agent bodies are under `.claude/agents/moai/`.
- **AP (D4)**: Do NOT attempt whole-file `diff` reconciliation of `manager-spec/docs/git.md` against their template mirror — the baseline divergence is owned by another SPEC. Mirror only THIS SPEC's changed lines.
- **AP**: Do NOT modify `TestEntryRouterLOCCeiling` to raise the ceiling — the fix is the asset, not the gate.

## §H — Cross-References

- `.claude/rules/moai/workflow/archived-agent-rejection.md §C` — migration table (replacement SSOT)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — frontmatter schema + ownership matrix
- `.claude/rules/moai/development/sprint-round-naming.md` — Epic/Milestone taxonomy (REQ-WADA-012)
- `CLAUDE.local.md §2 / §24 / §25` — Template-First, harness namespace, template neutrality
- `.moai/config/sections/workflow.yaml` `role_profiles:` (line ~68) — the 7 live role_profiles incl. `researcher` (D2 KEEP carve-out source)
- `internal/skills/workflow_split_test.go:154` — `TestEntryRouterLOCCeiling` (REQ-WADA-014/015)
- `internal/skills/workflow_split_test.go:159/184` — `TestTemplateMirrorParity` + `devOnlyLocalFiles` (release.md/github.md/release-update.md exclusion; D4/D5 source)
