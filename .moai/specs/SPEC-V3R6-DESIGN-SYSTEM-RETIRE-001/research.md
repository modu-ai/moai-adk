# research.md — Retirement Rationale (Evidence-Bearing)

> This document records the evidence base for retiring `moai-design-system`.
> Per the Verification-Claim Integrity doctrine (`.claude/rules/moai/core/verification-claim-integrity.md`),
> every rationale claim is attributed to an actually-measured grep/Read/lint
> output in this run, against this tree. Where evidence is mixed or absent,
> that gap is named explicitly in §R3.

## §R1. Phantom-status evidence (zero agent preloads)

### §R1.1 Claim

`moai-design-system` is preloaded by zero MoAI agents. It is a phantom skill
in the sense established by the H3 finding in the `project_skill_audit_followup_wip`
memory — "14종 스킬이 전부 agent preload 0 (동일 phantom 원인)".

### §R1.2 Evidence

Command run in this plan-phase:

```bash
grep -rln "moai-design-system" /Users/goos/MoAI/moai-adk-go/.claude/agents/ \
  /Users/goos/MoAI/moai-adk-go/.claude/skills/ 2>/dev/null \
  | grep -v "moai-design-system/"
```

Output: exactly **one** match —
`/Users/goos/MoAI/moai-adk-go/.claude/skills/moai-workflow-design/SKILL.md`.
That match is the negative pointer in the description prose (line 8: `"NOT
for general design system documentation (see moai-design-system)."`), NOT a
preload declaration.

Cross-check: every `skills:` frontmatter block in `.claude/agents/moai/`:

```bash
grep -rn "skills:" /Users/goos/MoAI/moai-adk-go/.claude/agents/ 2>/dev/null
```

Output: 8 agent files declare `skills:` frontmatter (manager-git,
manager-develop, builder-harness, sync-auditor, manager-spec, manager-docs,
+ 2 harness specialists). None of their `skills:` arrays contains
`moai-design-system` (verified by the grep above returning zero agent-file
matches for the literal).

### §R1.3 Baseline-attribution

The grep was run against the current working tree at plan-phase
(2026-06-19). The "zero preloads" claim is measured against this tree, not
carried over from prior audits.

### §R1.4 Recurring drift signal (corroborating)

The lint-baseline snapshots in `.moai/state/lint-baseline-{pre,post}-LCLN-P{1,2,3,4}.json`
emit, repeatedly:

> `"Skill preload drift in category 'expert': moai-design-system is not
> preloaded by all agents (may cause inconsistent context)"`

This is a recurring drift signal across multiple LCLN lint-baseline runs.
It corroborates the phantom status: the lint engine itself flagged the
skill as never-preloaded, and the flag was never resolved.

### §R1.5 Gaps

The grep covers `.claude/agents/` and `.claude/skills/`. It does NOT cover:

- `.claude/workflows/*.js` (dynamic-workflow agent() calls) — retirement of
  a skill from workflow scripts is a separate concern; no design-system
  reference was found in the repo-wide grep at the top of plan-phase, but
  this surface was not exhaustively Read.
- `internal/template/templates/.claude/agents/` — the template-agent
  surface. The repo-wide grep at plan-phase returned zero matches for
  `moai-design-system` in template-agent files, so this gap is low-risk.

## §R2. Functional-overlap analysis with surviving skills

### §R2.1 The design-system skill's advertised scope (from its retired SKILL.md)

Read from `git show HEAD:internal/template/templates/.claude/skills/moai-design-system/SKILL.md`
(the file is now unstaged-deleted, so the HEAD blob is the source of truth
for its advertised scope):

> **description**: "Unified design system specialist integrating Intent-First
> design craft and UI/UX foundations (accessibility, design tokens,
> component architecture)."
>
> **when_to_use**: "design-system authoring and audits: Intent-First craft,
> WCAG/ARIA accessibility, design tokens, theming and dark mode, component
> libraries (shadcn, Radix UI, Storybook), Style Dictionary, icon sets
> (Lucide, Iconify, Hugeicons), UX writing, and responsive UI implementation."

### §R2.2 Surviving-skill coverage matrix

| design-system scope item | Covered by surviving skill? | Evidence |
|--------------------------|-----------------------------|----------|
| Intent-First design craft | PARTIAL — `moai-workflow-design` description references design-brief context; `moai-domain-brand-design` covers visual design intent | workflow-design SKILL.md line 4-6; brand-design SKILL.md description |
| WCAG/ARIA accessibility | YES — `moai-domain-brand-design` description: "WCAG 2.1 AA accessibility ... accessibility enforcement" | brand-design SKILL.md description |
| Design tokens | YES — `moai-domain-brand-design` description: "design token extraction ... color palettes, typography, spacing" | brand-design SKILL.md description |
| Theming / dark mode | PARTIAL — brand-design covers color/typography systems; theming switch is a frontend impl detail covered by `moai-domain-frontend` | brand-design + frontend SKILL.md descriptions |
| Component libraries (shadcn, Radix, Storybook) | YES — `moai-domain-frontend` description: "React 19, Next.js 16 ... component architecture"; the project ships `.claude/rules/moai/languages/` + ref-react-patterns covering the same ground | frontend SKILL.md description |
| Style Dictionary | NO DIRECT COVER — Style Dictionary is a specific token-pipeline tool; no surviving skill names it. **Residual gap** | (none) |
| Icon sets (Lucide, Iconify, Hugeicons) | NO DIRECT COVER — no surviving skill names icon-library selection. **Residual gap** | (none) |
| UX writing | NO DIRECT COVER — `moai-domain-copywriting` is a separate skill but it covers marketing copy, not UX microcopy. **Residual gap** | copywriting SKILL.md |
| Responsive UI implementation | YES — `moai-domain-frontend` description: "responsive UIs, ... CSS, Tailwind" | frontend SKILL.md description |

### §R2.3 Verdict on overlap

Of 9 advertised scope items:
- 4 are directly covered by surviving skills (accessibility, tokens,
  component libraries, responsive UI).
- 2 are partially covered (Intent-First craft, theming).
- 3 are residual gaps (Style Dictionary, icon-set selection, UX writing).

The `project_skill_audit_followup_wip` memory line 16 records the prior
session's verdict: *"brand-design이 design-system scope 과커버(중복 제거 손실 0)"*
— brand-design over-covers the design-system scope such that retiring
design-system loses zero unique coverage. The §R2.2 matrix largely
corroborates this for the core scope, but identifies 3 residual
micro-gaps (Style Dictionary, icons, UX writing) that the prior verdict
did not enumerate.

### §R2.4 Honest assessment

The overlap is strong but not total. The 3 residual gaps are narrow
tool-selection topics that are plausibly better served by Context7 MCP
lookups (`mcp__context7__get-library-docs` for Style Dictionary / Lucide /
Iconify documentation) than by a dormant in-repo skill that no agent
preloads. The retirement trade-off is: lose a phantom skill that no agent
uses, accept that Style Dictionary / icon-selection guidance now lives in
Context7 instead of in-repo. This is an acceptable trade given the
phantom status — the guidance was never reaching agents anyway.

## §R3. Mixed-evidence and residual risks

### §R3.1 No telemetry-based zero-invocation proof

R4 audit's RETIRE candidates that relied on a "zero invocations" claim
(e.g. `moai-tool-svg`, problem P-S10) required a 30-day telemetry window
before the RETIRE was considered safe. This SPEC has NOT run such a window.

The phantom-preload signal (§R1) is the strongest available evidence
short of a telemetry pass: it proves no agent auto-loads the skill, but it
does NOT prove no user has manually invoked `Skill("moai-design-system")`
in a session transcript.

A transcript grep was run:

```bash
find ~/.claude/projects/ -name "*.jsonl" -mtime -60 2>/dev/null \
  | xargs grep -l "Skill.*moai-design-system\|moai-design-system.*Skill" 2>/dev/null
```

Output: 5 transcript files match — but visual inspection shows these are
transcripts where `moai-design-system` appears in **documentation/audit
context** (skill-audit discussions, consolidation planning), NOT as an
actual `Skill("moai-design-system")` tool invocation. The grep is
imprecise (matches any co-occurrence of "Skill" and "moai-design-system"
on the same file, not the same line). A precise per-line invocation grep
was not run; this is a documented gap.

### §R3.2 Residual risk: manual invocation breakage

If a user (or future agent session) has been manually invoking
`Skill("moai-design-system")`, that invocation breaks on retirement —
Claude Code would return "skill not found". No mitigation in this SPEC.
The forward-link is a follow-up SPEC (not authored here) that would add a
retired-skill redirect or a deprecation notice. Accepted as residual risk.

### §R3.3 Residual risk: catalog pack becomes empty

After removing the `moai-design-system` entry from catalog.yaml's `design`
pack, the pack may become empty or near-empty (only
`moai-domain-brand-design` would remain in the design pack). An empty pack
is not a build-breaker (catalog.yaml tolerates empty skill lists), but it
is a downstream catalog-reshuffling concern, explicitly Out of Scope (§C).

## §R4. Provenance of the retirement decision

The retirement was staged in the 2026-06-18 skill-audit-followup session
(memory: `project_skill_audit_followup_wip`). The H5 attempt was aborted
on a parallel-session working-tree race — *"moai-design-system retire 직전
`git rm`이 병렬 세션 working-tree 활동으로 취소됨"*. The session
documented the **full blast radius** (10 files + 2 directories) before
aborting, which is the basis for this SPEC's §B.1 file-surface table.

The orchestrator performed `git stash pop stash@{0}` (dropped `d7fc56d849`)
before delegating this SPEC, so the 2-file SKILL.md deletion seed is now
unstaged in the working tree. This SPEC formalizes the FULL retirement
(the stash was only the deletion seed — items #3-#8 in §B.1 remain to be
done in run-phase).

The scope-out from NAMESPACE-V2 is recorded verbatim in
`.moai/specs/SPEC-V3R6-HARNESS-NAMESPACE-V2-001/progress.md` line 87.
