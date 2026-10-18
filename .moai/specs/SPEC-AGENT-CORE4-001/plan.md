---
id: SPEC-AGENT-CORE4-001
title: "Plan: piece 1"
version: "0.3.0"
created: 2026-10-18
---

# plan.md — SPEC-AGENT-CORE4-001 (piece 1)

Tier L (§F). TDD per `quality.yaml` `development_mode: tdd`; characterization first because the
agent source layer must reproduce shipped bytes. No time estimates. Design: `design.md`.

## §A Decisions most likely to change (reviewed first)

1. Q17 — from M1 on C2 is generated, which reverses `AGENTS.local.md` §2.0 ("C2 hand-edited"); gates M1 onward.
2. Q4 — C1 = C2 for the carried four; gates M2.
3. The agent source data model (frontmatter, capability tokens, block semantics) — `design.md` §1.
4. Q19 — the manager-design line 206 citation (chosen: the existing `delegation.yaml`).

## §B Gating by open decisions

| Milestone | Needs | Implementable without it |
|---|---|---|
| M0 | nothing | yes — pure measurement and pin, no behavior change |
| M1 | Q17 verdict recorded | code can be written and tested on the branch; it must not land (merge to local main) before Q17, because landing flips every concurrent lane's edit workflow |
| M2 | Q4 (and Q17) | no |

Verdicts on open Q rows are recorded in `progress.md` `## §F Operator verdicts` (a non-hashed
section, new letter per the section-letter rule), never by editing `decision-index.md` after audit
(which is in the plan-artifact hash). A recorded verdict is a precondition line of the M1/M2 gate.
Transition for in-flight cards that edit a C2 file while M1 lands: the lane absorbs main, moves its
edit into the matching agent source and runs `make agents-emit`; until then `make build` is red by
design.

## §C Milestones (priority High; commit `feat(SPEC-AGENT-CORE4-001): M<n> <subject> (card t1649)` + `🗿 MoAI`)

### M0 baseline and pin, alone in its commit
Exactly two tracked files: `.moai/specs/SPEC-AGENT-CORE4-001/m0-baseline.md` (the ledger commands
re-run on the then-HEAD) and `internal/template/agentemit/testdata/agents-head.sha256` (sha256 of the
24 C2/C3 files of the twelve at the parent tree, the pin; if main changed those files since
`b4ea2d6c93` the pin records the then-current bytes and `m0-baseline.md` says so). Nothing under
`.moai/reports/` is staged (`.gitignore:232-248`; no `git add -f`). Gate: AC-AC4-010(c).

### M1 agent sources for all twelve, renderers, check wiring (needs Q17)
RED first: AC-AC4-001..005, 008 (template subtests), 009, 011, 012. Work: `agentemit/source/*.md`
for all twelve as byte-for-byte copies of C2 (paths verbatim; the pin file proves zero change); the
loader, capability expansion, and two renderers (`design.md` §2); `agents-emit(-check)` rewired with
the summary line; numeral registry rows for `source/manager-design.md` mirroring the three existing
rows (`registry.go` manager-design citation rows); comments in `Makefile`, `emit.go`, `loader.go`
updated; the existing golden test retargeted (assertions kept, AC-AC4-012). Gate: AC-AC4-001..005,
008(template), 009 (`specs=12 copies=2`), 011, 012, guards. Because M1 changes no shipped byte,
the gate compares to the pin file, not to `b4ea2d6c93`.

### M2 carried four: the 8 line changes, C1/C3p generation (needs Q4)
RED first: AC-AC4-006, 007, 008 (project subtests). Work: apply the 8 pairs of `acceptance.md` §3 to
the four sources; generate C1, C2, C3, C3p for the carried four; the post pins and the pairs file
`testdata/carried-four-post.sha256` are copied from `acceptance.md` §3 (authored at plan time, not
by the implementer). Summary becomes `specs=12 copies=4`. Gate: AC-AC4-006..010, 011 and guards;
the numeral layer stays green because line 206 keeps the numeral and the new rows exist.

## §D Constraints (process rules, formerly in the requirement layer)
Template-First (edit the agent source, `make agents-emit`, `make build`, sync local copies); PRESERVE
the eight's shipped bytes, `agents-codex.yaml` semantics, `docs-site/**`, `README*.md`, other SPECs,
`AGENTS.local.md`; no `AskUserQuestion` in `internal/template`; no SPEC ids, dates or SHAs inside
template files; explicit pathspecs; re-read HEAD before each commit; no skipped test, no relaxed
assertion; the pin file is never edited after M0. Always-loaded rule files are not touched.

## §E Self-verification
5-section format per milestone; any CLI measurement cites a `moai` built from the tree by path with
its commit next to the tree HEAD (`verification-claim-integrity.md` §2.2).

## §F Tier judgment (lane judgment, decision Q18)
Counted honestly. Authored files: 12 agent sources, about 8 Go source and test files, `Makefile`,
`registry.go` (+ numeral test), 2 testdata files, `m0-baseline.md` — about 24 (the audit measured 24
excluding generated copies and 39 including the 16 generated copies). Lines: the 12 sources are
2,969 lines (252,623 bytes) of new tracked agent text, plus roughly 600-1000 lines of Go. By files
affected (more than 15) and by the size of the new data model that every later piece builds on,
this is **Tier L** (threshold 0.85), not Tier M. Ceilings: 12 REQ and 12 AC against 25 and 25.
Artifact set: spec, plan, acceptance, design, research (5) plus progress and decision-index.

## §G Risks
Exact-byte capability expansion may need more tokens than expected (tokens grow, tests decide); Q17
and Q4 unanswered stop M1 landing and M2; the source layer duplicates the `.md` files for the eight
until piece 2 absorbs them (accepted; C2 stays the distribution truth); the pin file is a permanent
frozen comparison until piece 2 retires or retargets AC-AC4-005 (hand-off table).
