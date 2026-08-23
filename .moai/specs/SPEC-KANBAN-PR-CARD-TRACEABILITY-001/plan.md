# Implementation Plan — SPEC-KANBAN-PR-CARD-TRACEABILITY-001

Tier S: 2 artifacts (spec.md + plan.md); acceptance criteria are inline in
`spec.md` §D.

## A. Context

Doctrine-only SPEC, split out of `SPEC-KANBAN-QUEUE-PR-SYNC-001` per audit D14.
Two files change and no Go code ships:

- `.claude/rules/moai/workflow/kanban-dispatch.md`
- `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md`

Evidence base: `.moai/reports/t210/measurement.md` (M1, M2, M6) and
`.moai/reports/t210/verdict.md` (D3, D12, D14, D15). Do not re-derive either.

**Lands before the sibling.** `SPEC-KANBAN-QUEUE-PR-SYNC-001` ships the tooling;
this SPEC ships the obligation and the naming convention that makes the tooling
exact.

## B. Known issues

- **The AC greps are only non-vacuous against their zero baselines.** Both
  baselines were confirmed at 0 during the t210 audit, but they must be re-run
  and cited immediately before the edit — a baseline measured yesterday proves
  nothing about the tree being edited today.
- **The [HARD] marker grep is order-sensitive.** `grep -c '\[HARD\].*pre-dispatch
  PR cross-check'` matches only when the marker precedes the phrase on the same
  line. Write the clause that way, or the criterion fails on a correct clause.
- **The mirror is not a verbatim copy.** `internal/template/templates/` is
  deliberately neutralized — internal SPEC IDs, report paths, dates, and commit
  SHAs are stripped per the template neutrality catalogue. Copying the local
  file over the mirror will fail the neutrality guard. Write the mirror clause
  in neutral prose that names no SPEC ID and no `.moai/reports/` path.

## C. Pre-flight

```bash
# Zero baselines — cite these verbatim before editing.
grep -c 'pre-dispatch PR cross-check' .claude/rules/moai/workflow/kanban-dispatch.md
grep -c 'confirms or withdraws' .claude/rules/moai/workflow/kanban-dispatch.md
grep -c 'PR title MUST carry the delivering card id' .claude/rules/moai/workflow/kanban-dispatch.md

# Confirm both edit targets exist.
ls .claude/rules/moai/workflow/kanban-dispatch.md
ls internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md

# Read the section being amended, and the branch-naming rule REQ-004 reconciles.
grep -n 'sole producer\|Promotion is the operator\|may attach a finding' .claude/rules/moai/workflow/kanban-dispatch.md
grep -n 'WT-\|descriptive slug\|three other carriers' .claude/rules/moai/workflow/kanban-dispatch.md
```

## D. Constraints

- **Doctrine only.** No Go file is touched. If the work appears to need code,
  that is the sibling SPEC's scope — return a blocker rather than widening this
  one.
- **Do not weaken the three existing [HARD] clauses.** REQ-002's wording sits
  next to *"Promotion is the operator's act, always"* and must read as
  compatible with it, not as an exception to it.
- **Template-First.** Local edit and mirror edit land in the same commit,
  followed by `make build`.
- **Template neutrality.** The mirrored clause names no SPEC ID, no
  `.moai/reports/` path, no date, and no commit SHA.

## E. Self-verification

Each milestone's exit cites command output verbatim per
`.claude/rules/moai/core/verification-claim-integrity.md`. The
reviewer-judgement halves of AC-001, AC-002, and AC-003 are reported **as
judgements** — recorded, attributed, and never presented as mechanical passes.

## F. Milestones

### M1 — The pre-dispatch cross-check clause (REQ-001, REQ-002)

Least reversible of the two clauses: it changes what the lead must do on every
dispatch, and its report-and-re-decide wording is the control against a
de-facto-veto reading (audit D3).

- Site the clause in `kanban-dispatch.md` § Entry into the board is an operator
  act, after the three existing [HARD] clauses — that section already owns the
  operator-authority boundary.
- Write `[HARD]` before the phrase `pre-dispatch PR cross-check` on the same
  line (see §B).
- Include the literal `confirms or withdraws` (AC-002).
- Exit: the three AC-001 / AC-002 greps return ≥ 1 against cited zero baselines.

### M2 — The PR-title clause and the non-contradiction note (REQ-003, REQ-004)

- Site it in § Isolation, next to the branch-naming rule it reconciles with.
- Include the literal `PR title MUST carry the delivering card id`.
- Scope it to card-delivering pull requests; name release, batch, and
  release-update PRs as carrying no obligation.
- Name all four traceability carriers — dispatch `card:` field, commit message,
  evidence path, PR title — and state which is machine-readable.
- Exit: AC-003 grep returns ≥ 1 against its cited zero baseline.

### M3 — Template mirror

- Mirror both clauses into
  `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md`,
  in neutral prose per §D.
- `make build`.
- Exit: AC-004 — `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/... -v`
  passes; both clauses present in the mirror.

## G. Anti-patterns to avoid

- **Writing REQ-002 as a veto.** "The lead shall not dispatch a card carrying an
  open PR" reads as lead-side authority over a picked card and contradicts line
  29. The clause reports; the operator decides.
- **Copying the local file over the template mirror.** Fails the neutrality
  guard; the mirror is deliberately not a verbatim copy.
- **Claiming the reviewer-judgement halves as mechanical passes.** The greps
  prove a string was typed. That the clause *means* what REQ-001 through REQ-004
  require is a judgement, and is recorded as one.
- **Citing the t210 audit's zero baselines instead of re-measuring.** They were
  true at audit time; re-run them against the tree being edited.
- **Widening scope into tooling.** Any `gh`, `git`, or Go work belongs to
  `SPEC-KANBAN-QUEUE-PR-SYNC-001`.

## H. Cross-references

- `spec.md` §B (the four requirements), §C (why they are cheap), §D (the inline
  acceptance criteria)
- `SPEC-KANBAN-QUEUE-PR-SYNC-001` — the sibling tooling SPEC; lands second
- `.moai/reports/t210/measurement.md`, `.moai/reports/t210/verdict.md`
