# plan.md — SPEC-ANALYZE-FIRST-ROUTING-001

> Tier M. Docs/config only (no Go). Epic AGENTIC-CORE, SPEC 1 of 3.
> Shared findings: `./research.md`.

## §A — Context

- **Work location**: repo root `/Users/goos/MoAI/moai-adk-go`.
- **Affected surfaces** (all docs/config):
  - `CLAUDE.md` §2 (Request Processing Pipeline).
  - `.claude/skills/moai/SKILL.md` (Intent Router P3 + `team/*.md` refs).
  - `.claude/agents/moai/*.md` (all 9 agent bodies).
  - `.claude/skills/moai-foundation-core/SKILL.md` (`related-skills` drift).
  - `.claude/rules/moai/development/skill-authoring.md` (`triggers:` section).
  - `.claude/rules/moai/development/agent-authoring.md` (example alignment, if needed).
  - `internal/template/templates/.claude/**` mirrors for every changed `.claude/` file.
- **PRESERVE (do NOT touch)**: CLAUDE.md §1 HARD Rules, §8 AskUserQuestion
  architecture, Implementation Kickoff Approval; all Go source; the
  `run.md § Run-phase Autonomy (/goal ac_converge)` section (owned by
  `SPEC-AUTONOMY-RUN-GOAL-001`); the CLAUDE.md §2 pipeline enumeration STRUCTURE
  (compact it, do not delete — REQ-AFR-004).

## §A.5 PRESERVE / EXTEND list

| Path | Disposition |
|------|-------------|
| `CLAUDE.md` §1, §8 | PRESERVE |
| `CLAUDE.md` §2 | REWRITE (compact, keep ordered enumeration) |
| `.claude/skills/moai/SKILL.md` P1 | PRESERVE |
| `.claude/skills/moai/SKILL.md` P3 | EXTEND (language clause + `lint` collision fix) |
| `.claude/agents/moai/*.md` (×9) | REWRITE (diet trigger blocks) |
| `internal/**`, `pkg/**`, `cmd/**` | PRESERVE (no Go changes) |
| `internal/template/templates/.claude/**` | MIRROR (D5) |

## §B — Known Issues (filtered for docs/config Tier M)

- **B4 — Frontmatter canonical schema**: this SPEC's own frontmatter uses
  `created:`/`updated:`/`tags:` (verified). N/A to the edited docs (they are not SPECs).
- **B6 — spec-lint heading convention**: this spec.md uses `### §D.1 Out of Scope
  — <topic>` H3 sub-headings (satisfies `OutOfScopeRule`).
- **B8/B10 — working-tree hygiene / scope discipline**: agent-body edits are
  bulk-touch across 9 files; commit only the intended `.claude/` + mirror paths,
  never runtime-managed files.
- **B2 — cross-SPEC conflict**: do NOT touch the `run.md` autonomy section
  (`SPEC-AUTONOMY-RUN-GOAL-001`); do NOT delete the CLAUDE.md §2 enumeration
  (`SPEC-HARNESS-EVOLVE-001` attaches to it).
- **Template neutrality (§25)**: agent-body mirrors MUST NOT carry internal SPEC
  IDs. The agent bodies themselves reference no SPEC ID in trigger prose, so the
  diet is neutrality-safe by construction — verify with a grep before mirroring.

## §C — Pre-flight (run before editing)

```bash
# 1. Baseline: measure the trigger-block char total (REQ-AFR-011 delta needs this)
grep -c "MUST INVOKE" .claude/agents/moai/*.md
wc -c .claude/agents/moai/*.md | tail -1

# 2. Confirm the P3 lint collision (read the P3 body)
grep -n "lint" .claude/skills/moai/SKILL.md

# 3. Confirm team/*.md references
grep -n "team/" .claude/skills/moai/SKILL.md || echo "no team/ refs"

# 4. Confirm foundation-core drift
grep -n "related-skills" .claude/skills/moai-foundation-core/SKILL.md
grep -n "related-skills" internal/template/templates/.claude/skills/moai-foundation-core/SKILL.md

# 5. Enumerate the 9 agent files
ls .claude/agents/moai/*.md
```

## §D — Constraints

- Docs/config only — NO Go changes (REQ-AFR / §D.1).
- Template-First: mirror every changed `.claude/` file, then `make build` (D5).
- §25 neutrality: no internal SPEC ID / REQ token / audit citation in template mirrors.
- Language Policy: agent/skill/CLAUDE bodies stay English (do NOT add KO/JA/ZH
  prose — the diet REMOVES per-language keyword blocks, it does not translate).
- Keep each agent's out-of-scope (NOT-for) clause.

## §E — Self-Verification (plan-phase audit-ready signal)

Run-phase completion is verified by the `acceptance.md` AC matrix. Plan-phase
audit-ready state is recorded in `progress.md` §E.1.

## §F — Milestones (priority-ordered, no time estimates)

- **M1 — CLAUDE.md §2 rewrite (D1)**: compact Analyze-First pipeline; preserve the
  5-stage enumeration; remove stale keyword phrasing. Priority High.
- **M2 — SKILL.md Intent Router (D2)**: P3 language clause + `lint`→`fix` collision
  fix; P1 unchanged; `team/*.md` dead-ref cleanup (D4 item folded here). Priority High.
- **M3 — Agent diet (D3)**: rewrite 9 agent bodies (remove keyword blocks, add the
  language-independence line, keep NOT-for clauses, add trigger prose to
  manager-design). Priority High — largest surface.
- **M4 — Hygiene (D4)**: foundation-core `related-skills` fix; skill-authoring
  `triggers:` section update; agent-authoring example alignment. Priority Medium.
- **M5 — Template mirror + build (D5)**: mirror all changed `.claude/` files;
  `make build`; grep mirrors for neutrality. Priority High (gate).

Ordering: M1 → M2 → M3 → M4 → M5 (M5 last — mirror after all source edits settle).

## §G — Anti-Patterns to avoid

- Translating agent bodies into KO/JA/ZH (the diet REMOVES per-language blocks;
  it does not add prose — Language Policy keeps bodies English).
- Deleting the CLAUDE.md §2 ordered enumeration (breaks HARNESS-EVOLVE attach points).
- Editing the `run.md` autonomy section (owned by AUTONOMY-RUN-GOAL-001).
- Vacuous mirror check (a bare "mirrors updated" claim) — use per-file diff.

## §H — Cross-References

- Shared `research.md` §B (routing findings), §D.2 (HARNESS-EVOLVE compatibility).
- `.claude/rules/moai/development/agent-authoring.md` `:49`/`:205`/`:215`.
- `.claude/rules/moai/development/skill-authoring.md` (`triggers:` block).
- `CLAUDE.local.md` §2 (Template-First), §25 (neutrality).

## § Deferred (NOT run-phase scope)

- **docs-site 4-locale translation** of the reformed routing description
  (`content/{en,ja,zh,ko}/`). Follow-up SPEC.
- **Deprecated `moai-meta-harness` retirement** and **`triggers:` backfill** across
  the ~30 skills that lack it. Follow-up SPECs.

## § Open Decisions ([NEEDS CLARIFICATION])

- [NEEDS CLARIFICATION: agent-diet char-budget target] — the brief proposes a
  total trigger-block target of ≤ 7K chars (from an estimated ~16.7K). Proposed
  default: treat ≤ 7K as a SHOULD target, and the HARD AC as "keyword blocks
  removed from all 9 agents + measured post-diet total is strictly less than the
  re-measured pre-diet baseline" (REQ-AFR-011). Confirm whether the 7K figure is a
  hard gate or an advisory target.
- [NEEDS CLARIFICATION: P3 `lint` collision — confirm current membership] — the
  `lint`-in-two-buckets claim is a prior-session assumption (research.md §B.3).
  Proposed default: run-phase author reads the P3 body first; if `lint` is NOT
  actually dual-membership, REQ-AFR-007 degrades to a no-op confirmation (still
  passes by asserting single membership). Confirm the collision exists before
  treating its removal as a deliverable.
