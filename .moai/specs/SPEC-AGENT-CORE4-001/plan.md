---
id: SPEC-AGENT-CORE4-001
title: "Plan: piece 1"
version: "0.4.0"
created: 2026-10-18
---

# plan.md — SPEC-AGENT-CORE4-001 (piece 1)

Tier L (§F). TDD per `quality.yaml` `development_mode: tdd`; characterization first because the
agent source layer must reproduce shipped bytes. No time estimates. Design: `design.md`.

## §A Decisions most likely to change (reviewed first)

1. The agent source format: verbatim shipped files plus block markers (`design.md` §1); the neutral capability vocabulary is deferred to piece 2.
2. Q19 — the manager-design line 206 citation (open; the plan's choice stands, two baseline-derived results depend on it).
3. Q20 — C1 of the eight stays hand-managed until piece 2 (default applied, implementation-level).
4. Block-marker syntax and fail-closed rules (`design.md` §1).

Closed by the operator (board `d-20261018T062320Z-5952`, decided_by operator, 2026-10-18, scope card:t1649): Q4 and Q17.
Their record is in `decision-index.md` (hashed; changed in this revision) and in `progress.md` `## §G
Operator verdicts`. Nothing is gated on an open decision any more: M0, M1 and M2 are all
implementable and landable once planned and audited.

## §B Milestones (priority High; commit `feat(SPEC-AGENT-CORE4-001): M<n> <subject> (card t1649)` + `🗿 MoAI`)

### M0 baseline, alone in its commit
Exactly one tracked file: `.moai/specs/SPEC-AGENT-CORE4-001/m0-baseline.md`: the parent SHA `<M0>^`,
the ledger commands re-run on it, `git ls-tree -r <M0>^` blob ids of the 24 C2/C3 files, the 30
test names of `go test ./internal/template/agentemit -list '^Test'`, the rosterguard result, and, if
the carried-four files differ from `8260bbaa0b`, the recomputed expected hashes. The commit message
subject is `feat(SPEC-AGENT-CORE4-001): M0 baseline (card t1649)`; the follow-up commit adding
`m0_commit: <sha>` to `progress.md` is not a change to any guarded path. Nothing under
`.moai/reports/` is staged (`.gitignore:232-248`; no `git add -f`). If any of the 8 old lines of spec
§C.1 no longer matches, M0 aborts (blocker report) and the pairs are recomputed and re-audited. A
foreign commit that changed other agent files is simply part of the baseline.

### M1 agent sources for all twelve, renderers, check wiring, §2.0 rewrite
RED first: AC-AC4-001..005, 008 (template subtests), 009, 011, 012, 013. Work:
`agentemit/source/*.md` for all twelve as verbatim copies of the baseline C2 files (paths and
frontmatter as shipped); the two renderers (`design.md` §2) with block markers; fail-closed rules;
`agents-emit(-check)` rewired with the summary line and `copies=2`; numeral registry rows for
`source/manager-design.md` mirroring the three existing rows; comments in `Makefile`, `emit.go`,
`agentemit.go`, `loader.go`, `golden_test.go` updated (the "Option A: the emitter never re-renders"
header is reversed for the Markdown face, so these files are in the change set); the golden test
retargeted per REQ-AC4-012; and the narrow `AGENTS.local.md` §2.0 rewrite (REQ-AC4-013) **in this
same commit**, whose message cites board `d-20261018T062320Z-5952`. Gate: AC-AC4-001..005, 008(template), 009, 011..013,
guards. M1 changes no shipped byte, so it is gated against the baseline read from git, not a literal SHA.

### M2 carried four: the 8 line changes, C1/C3p generation
RED first: AC-AC4-006, 007, 008 (project subtests). Work: apply the 8 pairs of spec §C.1 to the four
sources; generate C1, C2, C3, C3p for the carried four; summary becomes `specs=12 copies=4`. Gate:
AC-AC4-006..010 and guards; the numeral layer stays green because line 206 keeps its numeral.

## §C Constraints (process rules)
Template-First (edit the agent source, `make agents-emit`, `make build`, sync local copies); PRESERVE
the eight's shipped bytes at M1, `agents-codex.yaml` semantics, `docs-site/**`, `README*.md`, other
SPECs; `AGENTS.local.md` changes only as REQ-AC4-013 says; no `AskUserQuestion` in
`internal/template`; no SPEC ids, dates or SHAs inside template files; explicit pathspecs; re-read
HEAD before each commit; no skipped test, no relaxed assertion. The render path is **build-time
only** (`EmitAll` has no non-test caller outside the package; the deployed `moai update` path
copies embedded template files and is untouched), and `TestRealSetMarkdownTreeUnmodified` keeps
passing. Cards that edit an agent file directly and land after piece 1 move the edit into the
source (operator decision `d-20261018T062320Z-5952`).

## §D Self-verification
5-section format per milestone; any CLI measurement cites a `moai` built from the tree by path with
its commit next to the tree HEAD (`verification-claim-integrity.md` §2.2).

## §F Tier judgment (lane judgment, decision Q18)
Authored files: 12 agent sources, about 8 Go source and test files (+ the AGENTS.local edit test),
`Makefile`, `registry.go` (+ numeral test), `m0-baseline.md`, `AGENTS.local.md` — about 25 (the
iter2 audit measured 24 excluding and 39 including the 16 generated copies). Lines: the 12 sources
are 2,969 lines (252,623 bytes) of new tracked agent text plus roughly 600-1100 lines of Go. By files
affected (more than 15) this is **Tier L** (threshold 0.85). 13 REQ and 13 AC against ceilings 25 and
25; artifact set: spec, plan, acceptance, design, research (5) plus progress and decision-index.

## §G Risks
Exact-byte reproduction of 12 files by a block-aware renderer (the pin is the baseline; tests
decide); the shared-tree edit workflow changes for every concurrent card at M1 landing (accepted by
the operator, board `d-20261018T062320Z-5952`); the source layer duplicates the `.md` files of the eight until piece 2;
a foreign edit to one of the 8 old lines before M0 aborts M0.
