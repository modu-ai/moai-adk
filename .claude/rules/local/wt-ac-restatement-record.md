---
description: "Dev-only dogfood record (card t1067) displaced from the managed copy of worktree-integration.md — measurement provenance and working-example pointers for the acceptance-criteria command boundary subsection"
paths: "**/.claude/rules/moai/workflow/worktree-integration.md"
---

# Acceptance-Criteria Command Boundary — Dogfood Record (card t1067)

This file carries the moai-adk-go internal record that annotates
`.claude/rules/moai/workflow/worktree-integration.md` § "Acceptance-criteria commands — the
measured boundary and the authoring rule". It lives here, not in the rule file, because the
managed root `.claude/rules/moai/` is wiped and redeployed by every `moai update`
(`CleanMoaiManagedPaths`) and must stay byte-identical to its neutral template mirror
(`TestRuleTemplateMirrorDrift`). `.claude/rules/local/` is tracked, unmanaged, and
mirror-free — the durable home for dev-only records. Relocated by card t1086
(SPEC-MIRROR-DOGFOOD-001); the rule file now carries the template's neutral wording.

## Measurement provenance

- The corpus-wide, block-level census of acceptance-criteria verification commands (111 files
  carrying complex git forms across `**/acceptance.md`) was probed at Claude Code **2.1.278**
  on **2026-09-22** — measured 2026-09-22, card **t1067** (SPEC-AC-GUARD-001 M2, commit
  `dbe1a6941`).

Displaced wording (verbatim, from the managed copy before relocation):

> of its block composition in a worktree-isolated session at Claude Code **2.1.278** (2026-09-22;
> every disposition traces to a recorded probe; measured 2026-09-22 — card t1067). Like every
> table in this section it is a record of observations, not a specification of the parser.

## Working-example pointers

- **AC-AEC-013 restatement** — `SPEC-AUDIT-EXPORT-CLAUSE-001/acceptance.md:622-680` (plain verbs
  + separate `echo "exit=$?"` lines + an explicit note avoiding `$(git merge-base …)`).
- **Corpus census** — `.moai/reports/t1067/census-20260922.md` (card t1067, measured
  2026-09-22).
  **Disposition: HISTORICAL, not live evidence.** The census file was never committed — absent
  from the filesystem, the git index, and full history (`git log --all` on the path: 0 rows,
  measured 2026-09-23). `.moai/reports/` is local-only and untracked; the path resolved only
  inside card t1067's session. Do not cite it as evidence that can be re-read.

Displaced wording (verbatim, from the managed copy before relocation):

> Working example: the AC-AEC-013 restatement in `SPEC-AUDIT-EXPORT-CLAUSE-001/acceptance.md:622-680`
> (plain verbs + separate `echo "exit=$?"` lines + an explicit note avoiding `$(git merge-base …)`),
> and the corpus census backing this subsection at `.moai/reports/t1067/census-20260922.md` (card
> t1067, measured 2026-09-22).
