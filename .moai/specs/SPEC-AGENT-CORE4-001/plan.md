---
id: SPEC-AGENT-CORE4-001
title: "Plan: piece 1"
version: "0.2.0"
created: 2026-10-18
---

# plan.md — SPEC-AGENT-CORE4-001 (piece 1)

Tier M (see §F). TDD per `quality.yaml` `development_mode: tdd`; characterization first because the
neutral layer must reproduce shipped bytes. No time estimates.

## §A Decisions most likely to change (reviewed first)

1. Neutral frontmatter and capability tokens (new data model; every later piece builds on it).
2. Q4: C1 = C2 for the carried four reverses `AGENTS.local.md` §2.0 (open product decision; gates M2 only).
3. The pinned changed-line list for the carried four (wording of the path-free replacements).
4. Location of neutral sources: `internal/template/agentemit/neutral/` (outside `templates/`, not distributed).

## §B Design (compact)

- **Neutral file**: frontmatter `name`, `description`, `capabilities` (ordered neutral tokens),
  `sandbox`, `model`, `effort`, `skills`, `memory`, `hooks`; then the body. Two optional blocks,
  `<!-- claude-only -->...<!-- /claude-only -->` and `<!-- codex-only -->...`, are kept in the
  matching render and removed from the other.
- **Capability expansion**: a table in the manifest maps each neutral token to the exact ordered
  Claude tool names it expands to (for example a write token to `Write, Edit`; MCP groups to their
  `mcp__moai__*` names). The Codex side keeps the existing class mapping in `agents-codex.yaml`.
  A token with no entry fails emission naming file and token. The expansion must reproduce each
  current `tools:` line byte for byte; if one cannot be expressed, the token set grows, never a raw
  escape hatch.
- **Render**: `Render(neutral) -> C2 bytes`; the existing `emitAll` then derives C3 from those
  bytes as today, so the C3 contract is untouched. `make agents-emit` writes C2, C3 (all twelve) and
  C1, C3p (carried four); `make agents-emit-check` re-renders in memory, compares the same sets and
  never writes; it prints `specs=<n> copies=<k> diff=<d>`.
- **Carried-four wording**: skills and files are named, not path-prefixed (a skill is "the
  `moai-workflow-testing` skill, file `references/e2e-desktop-native-recipes.md`"). The 8 changed C2
  lines are listed in `testdata/carried-four-lines.txt` (old line, new line).
- **Project root**: piece 1 adds no hook or path guard; nothing resolves a project root (D2 moved to piece 2).

## §C Milestones (priority High; commit `feat(SPEC-AGENT-CORE4-001): M<n> <subject> (card t1649)` + `🗿 MoAI`)

### M0 baseline, alone in its commit
Add exactly `.moai/specs/SPEC-AGENT-CORE4-001/m0-baseline.md` (re-measured on the then-HEAD, rebase
re-pin: the ledger commands, the 12-file C2/C3 sizes, rosterguard result) and
`internal/template/agentemit/testdata/agents-head.sha256` (sha256 of the 24 C2/C3 files). Both are
tracked paths; nothing under `.moai/reports/` is staged (gitignored by operator directive,
`.gitignore:232-248`; no `git add -f`). Gate: AC9.

### M1 neutral layer, renderers, check wiring (the eight)
RED first: AC1-AC4, AC7 template subtests. Work: loader and capability expansion; Claude renderer;
Codex path unchanged downstream; neutral sources for the eight with zero body change (paths kept
verbatim, they retire in piece 2); `agents-emit(-check)` rewired with the summary line; existing
`TestGoldenCommittedArtifactsMatchEmission` retargeted to neutral input, no assertion relaxed.
Gate: AC1-AC4, AC7(template) green; `git diff --quiet b4ea2d6c93 HEAD -- <the 16 C2/C3 files of the eight>` exit 0.

### M2 carried four path-free, project copies generated (needs Q4 answered)
RED first: AC5, AC6, AC7 project subtests. Work: neutral sources for the carried four with
install-root-free bodies and the pinned line list; generate C1, C2, C3, C3p for them; summary
becomes `copies=4`. Gate: AC1 (twelve), AC5-AC7, guards.

## §D Constraints
Template-First (edit neutral source, `make agents-emit`, `make build`); PRESERVE the eight's shipped
bytes, `agents-codex.yaml` semantics, `docs-site/**`, `README*.md`, other SPECs, `AGENTS.local.md`;
no `AskUserQuestion` in `internal/template`; no SPEC ids, dates or SHAs in templates; explicit
pathspecs; HEAD re-read before each commit; no skipped test, no relaxed assertion. Process rules
(formerly REQ-011/020) live here, not in the requirement layer.

## §E Self-verification
5-section format per milestone; CLI measurements cite a `moai` built from the tree by path with its
commit next to the tree HEAD (`verification-claim-integrity.md` §2.2).

## §F Tier judgment
Tier M. Scope: about 8 Go source and test files (loader, renderer, check wiring, tests), one
Makefile edit, 12 neutral sources, 2 testdata files and regenerated copies of 4 agents x 4 copies;
estimated 600-1000 LOC excluding neutral bodies, below the 1000-LOC and 15-file marks that define
Tier L, no constitutional change, 8 REQ and 9 AC against the Tier M ceilings of 16 and 16. The
original Tier L judgment was for the unsplit scope. Artifacts: spec, plan, acceptance (design and
research of the first revision are removed; their history is `b4ea2d6c93`).

## §G Risks
Exact-byte capability expansion may need more tokens than expected (mitigated: tokens grow, test
decides); Q4 unanswered stops M2 only; the neutral layer duplicates the `.md` files until piece 2
absorbs the eight (accepted: C2 stays the distribution truth).
