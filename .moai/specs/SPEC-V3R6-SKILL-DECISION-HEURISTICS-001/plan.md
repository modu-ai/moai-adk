# Plan — SPEC-V3R6-SKILL-DECISION-HEURISTICS-001

> Implementation plan for the Decision Heuristics + anti-pattern provenance binding across 4 high-traffic moai-adk skills.
> Companion to `spec.md` (§A-§J) and `acceptance.md`. Tier S, V3R6 era, Sprint 15.

---

## §A. Context

This SPEC applies two borrowable skill-craft devices from `github.com/wquguru/harness-books` (`.codex/skills/harness-book-best-practice/SKILL.md`) to the 4 highest-traffic moai-adk skills:

1. A closing `## Decision Heuristics` section — 3-5 compact "if X, default to Y" rules giving an agent fast default-routing without loading the full skill body.
2. Inline past-failure provenance one-liners on the skills' existing evolvable rationalization / red-flag rows, bound to the memory system's documented incident history (date + SPEC-ID).

### Honest scope reframing (load-bearing)

The SPEC source clause "(b) For each existing AP-* anti-pattern code already present in that skill" was empirically falsified: **none of the 4 target skills carry inline `AP-*` codes today** (`grep -c "AP-" SKILL.md` = 0 for all four). The canonical `AP-*` codes live in the rule files. Deliverable (b) is therefore reframed onto the skills' **existing evolvable rationalization/red-flag rows** — each relevant row gains a one-line provenance one-liner. This honesty is required by the no-unobserved-claim invariant.

### The 4 target skills

| # | Skill | Current end-of-body anchor | Heuristic source sections |
|---|---|---|---|
| 1 | `.claude/skills/moai-foundation-core/SKILL.md` | `## Token Budget` (absorbed) closing block | Quick Decision Guide, TRUST 5, Delegation Patterns, Token Optimization |
| 2 | `.claude/skills/moai-workflow-spec/SKILL.md` | `## Verification` evolvable block | GEARS Format, Requirement Clarification, SPEC Lifecycle, SPEC Scope/Classification |
| 3 | `.claude/skills/moai-foundation-cc/SKILL.md` | (to be located in M2) | Skill/Agent/Hook authoring sections |
| 4 | `.claude/skills/moai-meta-harness/SKILL.md` | `## Works Well With` / `## Out of Scope` | 7-Phase Workflow, Namespace Separation, Trigger Mechanics |

## §B. Known Issues

- **K-1**: The 4 skills have different end-of-body structures (some end with a Works Well With section, some with an evolvable verification block, some with an absorbed-module block). The Decision Heuristics section placement must be consistent (immediately before the final `## Works Well With` or, where absent, at the literal end of body).
- **K-2**: moai-foundation-cc/SKILL.md content was not fully read at plan-phase; M2 begins with a structural read to locate the insertion anchor. Low risk (additive append).
- **K-3**: Some memory provenance is date-rich (AP-V-004, AP-SRN-004, AP-D-001..005); some is date-poor. The pending-provenance form covers the latter. No date is invented.

## §C. Pre-flight Checks (verifiable, before M1)

1. `ls .claude/skills/moai-{foundation-core,workflow-spec,foundation-cc,meta-harness}/SKILL.md` → 4 files exist.
2. `grep -c "## Decision Heuristics" .claude/skills/moai-{foundation-core,workflow-spec,foundation-cc,meta-harness}/SKILL.md` → 0 / 0 / 0 / 0 (section not yet present; baseline).
3. `grep -c "AP-" .claude/skills/moai-{foundation-core,workflow-spec,foundation-cc,meta-harness}/SKILL.md` → 0 / 0 / 0 / 0 (no inline AP-* codes today — confirms honest-scope reframing).
4. Memory provenance spot-check: `grep -lE "AP-(V-004|SRN-004|D-00[1-5]|V-00[1-3])" ~/.claude/projects/*/memory/*.md` → ≥3 files return (provenance sources present).

## §D. Constraints (recap from spec.md §D)

- C-1 Localization: English (agent-facing).
- C-2 Scroll limit: binary — ≤ 13 lines PASS, ≥ 14 lines VIOLATION (matches spec.md §D C-2 and acceptance.md AC-SDH-003).
- C-3 Provenance honesty: pending-provenance form is the ONLY fallback; no fabricated dates.
- C-4 Template scope: LOCAL `.claude/skills/` only; template mirror sync is follow-up.
- C-5 Additive only: no deletions, no rewrites, no reorderings.
- C-6 Tier-S envelope: 4 files, single domain, no Go code, no rule-file change.

## §E. Self-Verification (plan-phase, manager-spec owned)

This section is the plan-phase audit-ready skeleton marker. The full §E evidence population belongs to manager-develop (run-phase, §E.2/§E.3) and manager-docs (sync/Mx-phase, §E.4/§E.5) per the artifact ownership matrix. See `progress.md` for the canonical §E placeholder skeleton.

- [ ] SPEC ID regex decomposition printed with `→ PASS` marker (done in authoring turn).
- [ ] 12 canonical frontmatter fields present, `era: V3R6` set, `status: draft`.
- [ ] No snake_case aliases (`created` not `created_at`, `tags` not `labels`).
- [ ] 3 files created (spec.md, plan.md, acceptance.md) + progress.md §E skeleton.
- [ ] Exclusions section (§H) present with ≥1 entry (7 entries).
- [ ] GEARS notation used for all REQs (Ubiquitous / State-driven / Event-detected / Capability gate / Unwanted behavior).
- [ ] Honest scope note (§A) documents the AP-* absence finding.

## §F. Milestones (priority-ordered, no time estimates)

### M1 — moai-foundation-core + moai-workflow-spec edits (Priority High)

- **M1.1** — Append `## Decision Heuristics` section to `moai-foundation-core/SKILL.md` (3-5 heuristics distilled from Quick Decision Guide + TRUST 5 + Token Budget). ≤ 12 lines.
- **M1.2** — Append provenance one-liners to moai-foundation-core's existing evolvable rationalization/red-flag rows where a memory-documented anti-pattern family maps (e.g. delegation-bypass rows may cite past permission-bypass incidents if memory carries them; else pending-provenance form).
- **M1.3** — Append `## Decision Heuristics` section to `moai-workflow-spec/SKILL.md` (3-5 heuristics distilled from GEARS Format + SPEC Lifecycle + SPEC Scope/Classification). ≤ 12 lines.
- **M1.4** — Append provenance one-liners to moai-workflow-spec's existing evolvable rows where AP-SRN-004 / SPEC-scope-confusion history maps. Use verified dates from memory (e.g. AP-SRN-004 → 2026-05-25 Sprint 10 chore `64310df3f`).
- **M1.5** — Self-verify M1.1-M1.4: `grep -c "## Decision Heuristics"` → 2 hits across the 2 files; line-count per section ≤ 13.

### M2 — moai-foundation-cc + moai-meta-harness edits (Priority Medium)

- **M2.1** — Structural read of `moai-foundation-cc/SKILL.md` to locate insertion anchor (resolves K-2).
- **M2.2** — Append `## Decision Heuristics` section to `moai-foundation-cc/SKILL.md` (3-5 heuristics distilled from Skill/Agent/Hook authoring sections). ≤ 13 lines.
- **M2.3** — Append provenance one-liners to moai-foundation-cc's evolvable rows where memory-documented anti-pattern families map (e.g. AskUserQuestion-bypass, background-write-restriction families have memory entries).
- **M2.4** — Append `## Decision Heuristics` section to `moai-meta-harness/SKILL.md` (3-5 heuristics distilled from 7-Phase Workflow + Namespace Separation + Trigger Mechanics). ≤ 13 lines. **Deliverable (a) only** — this skill carries ZERO `moai:evolvable-start` markers (§F A-3 baseline verified 2026-06-18), so deliverable (b) has no evolvable row to bind to.
- **M2.5** — **N/A for moai-meta-harness.** This milestone is a no-op for moai-meta-harness because that skill has no evolvable rationalization/red-flag/verification blocks (verified `grep -c 'moai:evolvable-start' .claude/skills/moai-meta-harness/SKILL.md` = 0). No provenance one-liners are appended; no evolvable content is fabricated. This honest N/A is documented in spec.md §F A-3 and verified by M3.6 below. (M2.5 remains a named step so the run-phase log records the explicit N/A decision rather than silently skipping it.)
- **M2.6** — Self-verify M2.1-M2.5: `grep -c "## Decision Heuristics"` → cumulative 4 hits across all 4 files (moai-meta-harness contributes 1 hit from M2.4 deliverable (a) only).

### M3 — Frontmatter lint + grep verification (Priority High, gate)

- **M3.1** — Run skill frontmatter lint on all 4 edited skills (or equivalent YAML parse + schema check). Expected: 0 findings.
- **M3.2** — `grep -rc "## Decision Heuristics" .claude/skills/moai-{foundation-core,workflow-spec,foundation-cc,meta-harness}/SKILL.md` → 4 hits total (1 per file).
- **M3.3** — Line-count each Decision Heuristics section: confirm each ≤ 13 lines PASSES; ≥ 14 lines is a VIOLATION (binary threshold per spec.md §D C-2).
- **M3.4** — Provenance honesty audit: `grep -E "AP-[A-Z]" .claude/skills/moai-{foundation-core,workflow-spec,foundation-cc}/SKILL.md` — confirm every cited date/SHA also appears in the memory system OR uses the pending-provenance form. No fabricated dates. (moai-meta-harness is excluded from this grep — it has no evolvable rows to carry provenance, per M2.5 N/A.)
- **M3.5** — Body-preservation audit: `git diff` on the 4 files shows additive changes only (no `-` lines except whitespace). Confirms C-5.
- **M3.6** — Honest-evolvable-N/A verification for moai-meta-harness: `grep -c 'moai:evolvable-start' .claude/skills/moai-meta-harness/SKILL.md` MUST return 0 (unchanged from §F A-3 baseline). Confirms deliverable (b) was honestly skipped for this skill, not fabricated. Also: `grep -c 'moai:evolvable-start' .claude/skills/moai-{foundation-core,workflow-spec,foundation-cc}/SKILL.md` MUST each return 3 (unchanged baseline — the 3 evolvable-bearing skills keep their markers; provenance was appended INSIDE existing blocks, not by adding new ones).
- **M3.7** — Template follow-up flag verification (REQ-SDH-009 / AC-SDH-008): `grep -nE 'FU-1|Template mirror sync|template-source mirror' plan.md` MUST return ≥1 hit in plan.md §I, confirming the follow-up is flagged as out-of-run-phase-scope. This is a documentation-existence check, NOT a run-phase deliverable.

## §G. Anti-Patterns (process)

- **AP-PLAN-001 — Inventing heuristics not present in the body.** Each heuristic MUST cite the section it compresses (e.g. `(<- §Quick Decision Guide)`). Inventing new policy violates REQ-SDH-004.
- **AP-PLAN-002 — Fabricating a provenance date.** The memory system is the only date/SHA source. Where memory lacks the fact, use `observed recurrence, provenance pending in memory` verbatim. Fabrication violates REQ-SDH-006 and the no-unobserved-claim invariant.
- **AP-PLAN-003 — Editing the template source during this SPEC.** LOCAL only. Template mirror is a flagged follow-up. Editing `internal/template/templates/.claude/skills/` mid-SPEC over-scopes Tier S and violates C-4.
- **AP-PLAN-004 — Rewriting an existing evolvable row when appending provenance.** The provenance one-liner is APPENDED to the existing row (inside the evolvable block). Rewriting the row's text violates C-5 + REQ-SDH-008.
- **AP-PLAN-005 — Exceeding the 12-line scroll bound.** A 5-heuristic section must stay compact. If a heuristic cannot fit one line, it is too detailed — compress further or drop to 4 heuristics.

## §H. Cross-References

- **spec.md** — §A Honest Scope Note, §C REQs (REQ-SDH-001..009), §D Constraints, §H Exclusions.
- **acceptance.md** — AC-SDH-001..008 (Given/When/Then), edge cases, Definition of Done. AC-SDH-008 added iter-2 for REQ-SDH-009 (template follow-up flag verification).
- **progress.md** — §E skeleton (placeholder headings; run/sync/Mx evidence populated by downstream agents).
- `.claude/rules/moai/core/verification-claim-integrity.md` — no-unobserved-claim invariant (binds REQ-SDH-006).
- `.claude/rules/moai/development/sprint-round-naming.md` — AP-SRN-004 provenance source.
- `.claude/rules/moai/workflow/session-handoff.md` — AP-V-001..004 + AP-D-001..005 provenance source.
- CLAUDE.local.md §2 — Template-First Rule (governs C-4 follow-up flag).
- CLAUDE.local.md §19.1 — Implementation Kickoff Approval (plan-to-implement human gate; this SPEC is plan-phase only).

## §I. Follow-Up Flags (out of this SPEC's scope, must not be forgotten)

- **FU-1 (Template mirror sync)** — After this SPEC's LOCAL edits land, the template source at `internal/template/templates/.claude/skills/moai-{foundation-core,workflow-spec,foundation-cc,meta-harness}/SKILL.md` must be synced with the same Decision Heuristics + provenance additions, then `make build` run to regenerate `internal/template/embedded.go`. This is a SEPARATE step (Tier-S scope protection per C-4). A chore commit or follow-up SPEC should handle it. Tracked here so it is not forgotten.
- **FU-2 (Extend to remaining skills)** — If the 4-skill pilot validates the Decision Heuristics device, a later SPEC may extend the pattern to the remaining moai-adk skills. Out of scope for this SPEC.
- **FU-3 (Provenance refresh)** — The memory system accumulates new incidents over time. A periodic refresh SPEC may backfill newer dates into the provenance one-liners. Out of scope for this SPEC.

## §J. Plan-Phase Audit-Ready Signal

Plan-phase complete (iter-2). 4 artifacts authored (spec.md / plan.md / acceptance.md / progress.md §E skeleton). SPEC ID self-check PASS. Honest scope reframing documented. iter-2 resolved 5 plan-auditor defects: D1 (§H Out of Scope h3), D2 (§F A-3 honest 3-of-4 evolvable baseline + REQ-SDH-005 N/A for moai-meta-harness + M2.5/M3.6 honest N/A verification), D3 (AC-SDH-008 added for REQ-SDH-009), D4 (binary ≤13/≥14 threshold across C-2/M3.3/AC-SDH-003), D5 (REQ-SDH-009 system-actor subject). spec-lint clean. Ready for Implementation Kickoff Approval (§19.1) before any run-phase delegation.
