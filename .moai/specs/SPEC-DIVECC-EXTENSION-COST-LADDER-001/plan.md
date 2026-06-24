# Implementation Plan — SPEC-DIVECC-EXTENSION-COST-LADDER-001

> Tier S plan. Pairs with plan.md of SPEC-DIVECC-DELEGATION-TOKEN-COST-001 (N3).

## §A. Context

- **Tier**: S (doc/doctrine-only; single rule file edit + optional template mirror + optional policy pointer; < 300 LOC of prose; < 5 files).
- **Run-phase target**: `.claude/rules/moai/development/agent-authoring.md` (canonical agent-authoring rule, `paths:`-scoped to `**/.claude/agents/**`).
- **Template mirror**: `agent-authoring.md` lives under `.claude/rules/` → it IS template-distributed. The run-phase MUST mirror the edit to `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` and run `make build`, then verify template neutrality (AC-ECL-006). The ladder content (generic mechanism cost classification + public paper citation + permanent-rule cross-references) is in the acceptable content class.
- **Optional secondary target**: a builder-harness policy section MAY carry a one-line cross-reference pointer to the ladder (REQ-ECL-005). The primary home is `agent-authoring.md`.
- **PRESERVE**: all existing `agent-authoring.md` sections (frontmatter, namespace separation, frontmatter fields, per-spawn specialization, decision tree). The ladder is an ADDITIVE new section, not a rewrite.

## §B. Known Issues (filtered to Tier S relevance)

- **B6 — spec-lint Out of Scope heading**: this plan's sibling spec.md uses a `### Out of Scope — <topic>` H3 sub-heading set under a `## §F. Out of Scope` H2 — satisfies `OutOfScopeRule`. (Plan-phase concern, already handled in spec.md §F.)
- **Template neutrality (CLAUDE.local.md §15/§25)**: when mirroring to the template tree, the ladder must carry NO forbidden internal-content class. The paper citation (arXiv:2604.14228) is a public-source citation (acceptable); permanent-rule cross-references are acceptable; no internal SPEC ID / REQ token / commit SHA / internal date may appear in the mirrored prose. AC-ECL-006 gates this.
- **B10 — scope discipline**: touch ONLY `agent-authoring.md` (+ its template mirror + optional builder-harness pointer). Do NOT touch hook scripts, skill bodies, plugin manifests, MCP config, or the N3 delegation surfaces (`moai.md` §4 / `CLAUDE.local.md` §16).

## §C. Pre-flight (run-phase)

```bash
# 1. Confirm target structure still matches plan grounding
grep -n "Static Agent File vs Per-Spawn\|## " .claude/rules/moai/development/agent-authoring.md | tail -10
# 2. Confirm cross-reference anchors still exist
grep -n "Progressive Disclosure" .claude/rules/moai/development/skill-authoring.md | head -2
grep -n "Purpose-driven model+effort selection" .claude/rules/moai/workflow/dynamic-workflows.md | head -2
# 3. Confirm template mirror exists
ls internal/template/templates/.claude/rules/moai/development/agent-authoring.md
```

## §D. Constraints (DO NOT VIOLATE)

- Doc-only. No Go code change, no hook/skill/plugin/MCP change (REQ-ECL-007).
- The four-tier cost figures are the paper's claim — attribute them, never present as moai-tree measurement (REQ-ECL-006).
- Mirror edit to the template tree + `make build`; verify template neutrality.
- Conventional Commits; `🗿 MoAI` trailer; NO `--no-verify` / `--amend` / force-push.

## §E. Self-Verification (run-phase deliverable)

Run-phase manager-develop reports the AC-ECL-001..006 PASS/FAIL matrix with verbatim grep output, plus:
- `git show --stat <run-commit>` confirming only the rule file (+ mirror + optional pointer) changed (AC-ECL-005).
- `go test ./internal/template/... -run TestTemplateNeutralityAudit` PASS (AC-ECL-006).

## §F. Milestones (priority-ordered, no time estimates)

- **M1 — Author the ladder section** in `.claude/rules/moai/development/agent-authoring.md`: a new section (e.g. `## Extension-Mechanism Context-Cost Ladder`) with the 4-tier table, the decision criterion, the paper attribution, and the two parallel-taxonomy cross-references. Satisfies REQ-ECL-001..004, REQ-ECL-006.
- **M2 — Mirror + build**: copy the section into the template tree, `make build`, verify `TestTemplateNeutralityAudit` (AC-ECL-006).
- **M3 — Optional builder-harness pointer** (REQ-ECL-005): if a builder-harness policy section is a better home for the *decision criterion* invocation, add a one-line cross-reference there. Skip if `agent-authoring.md` suffices.
- **M4 — Self-verify + commit + push**: run the AC grep matrix, commit (`docs(SPEC-DIVECC-EXTENSION-COST-LADDER-001): M1 extension context-cost ladder`), push to main (Hybrid Trunk Tier S).

## §G. Anti-Patterns to avoid

- Presenting the four-tier figures as moai-measured (violates REQ-ECL-006 + verification-claim-integrity).
- Rewriting adjacent `agent-authoring.md` sections (violates scope discipline — additive section only).
- Editing the N3 delegation surfaces (out of scope — N3 owns them).
- Leaking the SPEC ID / REQ tokens into the template mirror (violates template neutrality).

## §H. Cross-References

- spec.md §C (REQ-ECL-001..007), §D (AC matrix), §G (cross-references).
- SPEC-DIVECC-DELEGATION-TOKEN-COST-001 (N3) — thematic pair.
- `.moai/docs/template-internal-isolation-doctrine.md` — template neutrality gate for M2.
