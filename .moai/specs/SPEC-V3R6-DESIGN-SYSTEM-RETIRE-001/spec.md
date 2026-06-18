---
id: SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001
title: "Retire the moai-design-system skill — full lifecycle removal"
version: "0.1.0"
status: implemented
created: 2026-06-19
updated: 2026-06-19
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/skills + internal/cli + internal/template + internal/design + docs-site"
lifecycle: spec-anchored
tags: "skill-retire, design-system, phantom-skill, template-neutrality, namespace-cleanup, overlap-consolidation"
era: V3R6
tier: L
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-19 | manager-spec | Initial plan-phase authoring (5-artifact Tier L). Resumes the H5 design-system retirement that the 2026-06-18 skill-audit-followup session staged (stash@{0}) and aborted on a parallel-session working-tree race; orchestrator performed `git stash pop stash@{0}` (dropped `d7fc56d849`) so the 2-file deletion seed is now unstaged in the working tree. This SPEC formalizes the FULL retirement surface. |

## §A. Context (WHY)

The `moai-design-system` skill was created in v3.0.0 (CHANGELOG line 1075) as a
**merge target** — it absorbed the retired `moai-design-craft` and
`moai-domain-uiux` skills plus the Pencil portion of `moai-design-tools`. The
R4 skill audit (`.moai/design/v3-redesign/research/r4-skill-audit.md` line 421)
directed the merge as the "cleanest path" to resolve the P-S13 design-cluster
4-way overlap.

Three independent observations now make the merged skill itself a retirement
candidate:

1. **Phantom status (zero preloads).** A grep of `.claude/agents/**` shows
   `moai-design-system` appears in the `skills:` frontmatter of **zero** agents.
   The sole cross-skill reference is a negative pointer in
   `moai-workflow-design/SKILL.md` line 8 — `"NOT for general design system
   documentation (see moai-design-system)"`. The lint-baseline snapshots in
   `.moai/state/lint-baseline-*.json` repeatedly emit
   `"Skill preload drift in category 'expert': moai-design-system is not
   preloaded by all agents"` — a recurring drift signal never resolved. See
   `research.md` §R1 for the full evidence trail.

2. **Functional overlap with surviving design-adjacent skills.** The four
   surviving design-adjacent skills cover the design-system skill's advertised
   scope:
   - `moai-domain-brand-design` — "Brand-aligned visual design system
     specialist ... WCAG 2.1 AA accessibility ... design tokens, color
     palettes, typography, spacing, component specifications" (direct
     overlap with the design-system skill's "design tokens, theming, WCAG/ARIA
     accessibility, component libraries" scope).
   - `moai-domain-frontend` — "React 19, Next.js 16, Vue 3.5 ... component
     architecture" (overlap with the design-system skill's "shadcn, Radix UI,
     component libraries" scope).
   - `moai-domain-design-handoff` — design handoff workflow.
   - `moai-workflow-design` — `/moai design` workflow orchestration.
   The brand-design skill is documented (in the `skill_audit_followup_wip`
   memory) as over-covering the design-system scope such that retiring the
   design-system skill loses zero unique coverage. See `research.md` §R2.

3. **Prior session staging.** The 2026-06-18 skill-audit-followup session
   staged the 2-file SKILL.md deletion into stash@{0} before the
   NAMESPACE-V2 SPEC sync, explicitly scope-out so retirement gets its own
   SPEC (verbatim provenance in NAMESPACE-V2 progress.md line 87). The
   orchestrator has since `git stash pop`-ed the stash. The retirement
   decision was made; this SPEC formalizes it across the full surface.

### §A.1 Honest mixed-evidence note

The retirement rationale rests on phantom status + overlap, NOT on a fresh
telemetry-based zero-invocation proof. R4 audit's telemetry-based RETIRE
candidates (e.g. `moai-tool-svg`, problem P-S10) required a 30-day
invocation window; this SPEC has not run that window. The phantom-preload
signal is the strongest available evidence short of a telemetry pass, and
the overlap analysis shows zero unique-coverage loss. `research.md` §R3
documents this gap honestly and flags a residual risk: if a user has been
manually invoking `moai-design-system` by name via `Skill("moai-design-system")`,
that invocation breaks on retirement. No mitigation in this SPEC — see
`plan.md` §G Anti-Patterns for the forward-link to a follow-up.

## §B. Scope (WHAT)

Retire the `moai-design-system` skill across its FULL lifecycle surface. The
retirement applies symmetrically to the local `.claude/` tree and the
template `internal/template/templates/.claude/` tree per the Template
Neutrality doctrine (CLAUDE.local.md §15, §25). The skill is a `moai-*`
namespace skill = template-managed (CLAUDE.local.md §24), so deletion is the
correct treatment (not user-owned-preservation).

### §B.1 File-surface table (8 categories, 10 edit sites)

The orchestrator pre-investigated 8 file categories via repo-wide grep. The
plan-phase author verified each via direct Read. The table below is the
authoritative retirement surface. Any file not listed here is Out of Scope
(see §C).

| # | File (relative to repo root) | Category | Current state | Intended change |
|---|------------------------------|----------|---------------|-----------------|
| 1 | `.claude/skills/moai-design-system/SKILL.md` | local skill | DELETED (unstaged, stash-pop seed) | Confirm deletion in commit |
| 2 | `internal/template/templates/.claude/skills/moai-design-system/SKILL.md` | template skill | DELETED (unstaged, stash-pop seed) | Confirm deletion in commit |
| 2b | `.claude/skills/moai-design-system/` (directory) | local skill dir | empty after file deletion | `rmdir` (empty dir) |
| 2c | `internal/template/templates/.claude/skills/moai-design-system/` (directory) | template skill dir | empty after file deletion | `rmdir` (empty dir) |
| 3 | `internal/cli/doctor_skills.go` line 24 | Go allowlist | `"moai-design-system",` literal in static core-skill list (comment `// design (1)`) | Remove the literal AND the `// design (1)` comment line; the design category becomes empty |
| 4 | `internal/cli/doctor_skills_test.go` lines 51-55 | Go test | test case `name: "valid static core skill moai-design-system returns PASS", skillName: "moai-design-system"` | Remove the 5-line test-case struct block (lines 51-55 inclusive: opening `{`@51, `name:`@52, `skillName:`@53, `wantClass:`@54, `},`@55) |
| 5 | `internal/template/catalog.yaml` lines 164-168 | catalog entry | 5-line block: `- name: moai-design-system / tier / path / hash / version` | Remove the 5-line block from the `design` pack (the pack retains 6 robustly-populated entries — see research.md §R3.3) |
| 6 | `internal/design/dtcg/frozen_guard_test.go` line 50 | Go frozen-path test | `".claude/skills/moai-design-system/SKILL.md",` literal in allowedPaths | Remove the literal from allowedPaths slice #1 |
| 6b | `internal/design/dtcg/frozen_guard_test.go` line 104 | Go frozen-path test | `".claude/skills/moai-design-system/SKILL.md",` literal in allowedPaths slice #2 | Remove the literal from allowedPaths slice #2 |
| 7 | `internal/template/templates/.claude/skills/moai-workflow-design/SKILL.md` line 8 | cross-skill reference | frontmatter description prose `"... (see moai-design-system)."` and/or body prose | Remove the parenthetical `(see moai-design-system)` from the description; leave the rest of the description intact |
| 8 | `docs-site/content/en/advanced/skill-guide.md` line 126 | docs-site EN | table row `` `moai-design-system` | Intent-first design ... `` | Remove the table row AND decrement the section header `### Design (Design System) - 1 skill` → either remove the section entirely or rewrite the header to `### Design (Design System) - 0 skills` with an explanatory note |
| 8b | `docs-site/content/ko/advanced/skill-guide.md` line 123 | docs-site KO | table row `` `moai-design-system` | ... `` | Same treatment as 8, Korean prose |
| 8c | `docs-site/content/ja/advanced/skill-guide.md` line 120 | docs-site JA | table row | Same treatment, Japanese prose |
| 8d | `docs-site/content/zh/advanced/skill-guide.md` line 122 | docs-site ZH | table row | Same treatment, Chinese prose |

### §C. Exclusions (What NOT to Build)

This section preserves the canonical "What NOT to Build" intent. Each
sub-heading is an `### Out of Scope — <topic>` H3 per the
`OutOfScopeRule` lint convention (codified by sibling SPEC
`SPEC-V3R6-OUTOFSCOPE-GUIDANCE-ALIGN-001`).

### Out of Scope — Historical references (archival records)

- `CHANGELOG.md` (line 1072, 1075) — the v3.0.0 consolidation entry naming
  `moai-design-system` as a NEW skill is a record-of-the-past. Modifying
  historical CHANGELOG entries violates immutability.
- `.moai/specs/SPEC-V3R2-WF-001/**`, `.moai/specs/SPEC-V3R{3,4,5}-*/**` —
  past SPECs that reference `moai-design-system` in their plan/spec/research
  bodies. These are immutable SPEC archives.
- `.moai/research/**` (incl. `harness-autonomy-vision-2026-05-18.md`,
  `workflow-audit-2026-05-16.md`) — research notes are point-in-time records.
- `.moai/archive/**` (incl. `moai-design-craft/RETIRED.md`,
  `moai-domain-uiux/RETIRED.md` which name `moai-design-system` as the
  Substitute) — RETIRED.md files are the canonical "where did this skill go"
  records and remain valid as historical pointers even after the substitute
  itself retires.
- `.moai/release/**`, `.moai/decisions/skill-rename-map.yaml`,
  `.moai/brain/**`, `.moai/design/v3-redesign/**`, `.moai/state/lint-baseline-*.json`,
  `.moai/backups/**` — archival / state / backup records.
- `.claude/worktrees/**` and `.moai/specs/SPEC-V3R6-DOCS-COVERAGE-001/research.md`
  (which lists `moai-design-system` in a doc-coverage inventory) — worktree
  copies and prior-SPEC research bodies are out of scope.
- `docs/design/**` (incl. `docs/design/major-v3-master.md` lines 543 + 822,
  which name `moai-design-system` as the design-cluster merge target and in
  the v3.0.0 migration table) — the v3.0.0 master design doc is a
  design-record-of-the-past authored 2026-04-23, the same immutability class
  as `.moai/design/v3-redesign/**` (above). Without this exclusion, the §F
  success-criterion grep (which scopes `.claude/`, `internal/**`, and
  `docs-site/content/`) would technically PASS while leaving the
  `docs/design/` historical reference live.

### Out of Scope — Surviving skill content migration

- Migrating the design-system skill's functional content (modules/*.md, if
  any) into `moai-domain-brand-design` or `moai-domain-frontend`. The
  overlap analysis (research.md §R2) shows the surviving skills already
  cover the scope; migration would duplicate content. The design-system
  skill is deleted, not migrated.
- Editing `moai-domain-brand-design`, `moai-domain-frontend`,
  `moai-domain-design-handoff`, or `moai-workflow-design` body content to
  absorb design-system material. Only the `(see moai-design-system)`
  cross-reference in `moai-workflow-design` line 8 is in scope (item #7
  above).

### Out of Scope — Telemetry-based zero-invocation proof

- Running a 30-day invocation-window telemetry pass to prove zero manual
  `Skill("moai-design-system")` invocations. The phantom-preload signal +
  overlap analysis is the strongest available evidence short of a telemetry
  pass. The residual risk (manual invocation breakage) is documented in
  research.md §R3 and accepted.

### Out of Scope — Catalog-pack reshuffling

- Reorganizing the `design` pack in catalog.yaml beyond removing the single
  `moai-design-system` entry. After this single removal, the `design` pack
  retains **6 robustly-populated entries** (`moai-domain-brand-design`,
  `moai-domain-copywriting`, `moai-domain-humanize`,
  `moai-domain-design-handoff`, `moai-workflow-design`,
  `moai-workflow-gan-loop` — see research.md §R3.3), so NO empty-pack risk
  arises. Reordering those 6 surviving entries remains out of scope (a
  downstream concern for a separate catalog-reorganization SPEC, not this
  retirement).

## §D. Requirements (GEARS notation)

### REQ-DSR-001 — Template-symmetry of the deletion (Ubiquitous)

The retirement process **shall** apply the `moai-design-system/SKILL.md`
deletion symmetrically to BOTH the local `.claude/skills/` tree AND the
template `internal/template/templates/.claude/skills/` tree, preserving the
template-neutrality invariant (CLAUDE.local.md §15, §25).

### REQ-DSR-002 — Go allowlist cleanup (Ubiquitous)

The `internal/cli/doctor_skills.go` static core-skill list **shall not**
contain the `"moai-design-system"` literal after retirement; the now-empty
`// design (1)` comment line **shall** be removed alongside it.

### REQ-DSR-003 — Go test consistency (Event-detected)

**When** the `"moai-design-system"` literal is removed from
`internal/cli/doctor_skills.go`, the corresponding test case in
`internal/cli/doctor_skills_test.go` (lines 52-53) **shall** be removed in
the same change so the test suite does not reference a non-existent static
core skill.

### REQ-DSR-004 — Catalog entry removal (Ubiquitous)

The `internal/template/catalog.yaml` `design` pack **shall not** contain the
5-line `moai-design-system` entry block after retirement.

### REQ-DSR-005 — Frozen-guard test fixture update (Event-detected)

**When** the `.claude/skills/moai-design-system/SKILL.md` path no longer
exists on disk, the `internal/design/dtcg/frozen_guard_test.go` allowedPaths
slices (line 50 AND line 104) **shall** be updated to remove the path
literal, so the test does not whitelist a non-existent file.

### REQ-DSR-006 — Cross-skill reference cleanup (Ubiquitous)

The `moai-workflow-design/SKILL.md` description **shall not** contain the
`(see moai-design-system)` parenthetical pointer after retirement; the
remainder of the description **shall** be preserved verbatim.

### REQ-DSR-007 — docs-site 4-locale parity (Ubiquitous)

The docs-site skill-guide **shall** remove the `moai-design-system` table
row across ALL FOUR locales (`en`, `ko`, `ja`, `zh`) AND update the section
header consistently across the four locales, so the 4-locale parity
invariant is preserved.

### REQ-DSR-008 — Empty directory cleanup (Ubiquitous)

The retirement **shall** remove the now-empty
`.claude/skills/moai-design-system/` directory and the now-empty
`internal/template/templates/.claude/skills/moai-design-system/` directory,
so the working tree carries no orphan empty skill directories.

### REQ-DSR-009 — Historical-reference preservation (Unwanted)

The retirement process **shall not** modify any file listed in
§C "Out of Scope — Historical references", even when such files reference
`moai-design-system`; archival records are records-of-the-past and remain
valid as historical pointers.

### REQ-DSR-010 — Build, test, and lint green (State-driven)

**While** the retirement changes are staged for commit, the run-phase
verification batch (`go build ./...`, `go test ./internal/cli/... ./internal/design/dtcg/...`,
`go run ./cmd/moai spec lint`, `go run ./cmd/moai doctor --skills` smoke)
**shall** pass before the run-phase commit is created.

## §E. Constraints

- **era: V3R6** — the SPEC follows the modern 4-phase lifecycle (plan / run
  / sync / Mx).
- **GEARS notation** — requirements use GEARS (current); no `IF/THEN`.
- **Template Neutrality (§15, §25)** — local + template symmetry is a HARD
  invariant. The stash-pop seed already satisfies this for the 2 SKILL.md
  files; the SPEC preserves it across all 8 file categories.
- **Harness Namespace (§24)** — `moai-design-system` is template-managed;
  retirement (deletion) is the correct treatment, not user-owned-preservation.
- **spec-lint clean** — spec.md MUST pass `moai spec lint` with 0 findings.
- **No implementation in plan-phase** — Go code changes are enumerated in
  plan.md but executed in run-phase by manager-develop.

## §F. Success Criteria

- All 10 edit sites in §B.1 are applied.
- `go build ./...` passes.
- `go test ./internal/cli/... ./internal/design/dtcg/...` passes.
- `grep -rn "moai-design-system" .claude/ internal/template/templates/ internal/cli/ internal/template/catalog.yaml internal/design/dtcg/ docs-site/content/` returns ZERO matches (active-code zero-tolerance; historical references in CHANGELOG/.moai/** AND `docs/design/**` excluded by design per §C — the grep scope is intentionally narrower than the full repo to respect archival immutability).
- spec-lint on this SPEC returns 0 findings.
- 4-locale docs-site parity verified (same row-removal + header-update + global 32→31 count update in en/ko/ja/zh; no locale references `32` as the current total post-retirement).

## §G. Cross-References

- Sibling: `SPEC-V3R6-HARNESS-NAMESPACE-V2-001` (progress.md line 87 — the scope-out provenance).
- Sibling: `SPEC-V3R6-OUTOFSCOPE-GUIDANCE-ALIGN-001` (OutOfScopeRule h3 convention this SPEC's §C follows).
- Prior: `SPEC-V3R6-SKILL-CONSOLIDATE-001`, `SPEC-V3R6-SKILL-COMPRESS-001` (v3.0.0 skill consolidation lineage).
- Prior: `.moai/archive/skills/v3.0/moai-design-craft/RETIRED.md` (Substitute = moai-design-system — the original merge provenance, now itself retiring).
- Doctrine: CLAUDE.local.md §15 (Template Language Neutrality), §24 (Harness Namespace), §25 (Template Internal-Content Isolation).
- Research: `.moai/design/v3-redesign/research/r4-skill-audit.md` (R4 audit, merge directive).
- Memory: `project_skill_audit_followup_wip` (H5 staging provenance + blast-radius investigation).
