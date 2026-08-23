# Progress — SPEC-KANBAN-QUEUE-PR-SYNC-001

## §E.1 Plan-phase Audit-Ready Signal

### Artifacts

Tier M set: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`.

### Audit iteration 1 — FAIL, and what was done about it

`.moai/reports/t210/verdict.md` returned **FAIL** (0.60 harmonic, Tier M
threshold 0.80) on must-pass MP-3. The score was never the binding part.

| Finding | Disposition |
|---|---|
| D1 — `tags` was a YAML sequence; `moai spec lint` returned `ParseFailure` and **the document had never parsed** | Fixed. `tags` retyped to a comma-separated string. Every other lint rule had been silently unevaluated until this landed; the verbatim re-run is below. |
| D4 — AC-002 and AC-003 fixtures refuted by the carrier data | Fixed. AC-002 rebuilt on `t188`/#1601 (single-bodied, title-absent); AC-003 rebuilt on `t201`/{#1611,#1612} (the genuine ambiguous case). REQ-1.3, REQ-1.4, REQ-1.6 now have working criteria. |
| D2 + D16 + D17 — t199 unreachable by every specified carrier; REQ-1.5 over-generalized | Fixed per the lead's Decision A. §C now separates Q1 (attribution) from Q2 (landed); REQ-1.5 is scoped to Q1; REQ-1.9 adds the landed check with a mandated `--perl-regexp`; REQ-1.10 makes it boolean-only; REQ-1.7 is now four-valued. |
| D14 — 19 leaf REQs against a Tier M ceiling of 16 | Fixed per the lead's Decision B. The former REQ-3 became `SPEC-KANBAN-PR-CARD-TRACEABILITY-001` (Tier S). |
| D3 | Moved with the doctrine to the sibling SPEC, as REQ-002's report-and-re-decide wording. |
| D5 | Decided: no `--pr` flag is in scope. Recorded in §H with the AC-009 interaction stated. |
| D6 | Fixed. §B.1 now cites the third [HARD] clause of the same section (line 31), the same-file precedent whose byte-identity language REQ-2.1 adopts. |
| D7 | Fixed. AC-004 is now a recursive directory digest over `.moai/state/kanban/` plus a path-set assertion, covering REQ-2.2's wider surface. |
| D9, D10, D11, D12, D13, D15 | Fixed: `Where`→`While` on data conditions; trailers promoted to `shall` form; AC-005 disjunction resolved to "blank"; AC-013 split into mechanical and reviewer-judgement halves; AC-006's unreproducible count dropped; mirror parity extended to `todo.md`. |
| D8 | Accepted as cosmetic. The `REQ-N.M` scheme is internally consistent; noted for grep-based tooling assuming `REQ-[0-9]{3}`. |
| D18 | No action — a recorded positive verification of REQ-1.8. |

### Budget

```
$ grep -cE '^\*\*REQ-[0-9]+\.[0-9]+\*\*' .moai/specs/SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md
16
$ grep -cE '^### AC-[0-9]{3}' .moai/specs/SPEC-KANBAN-QUEUE-PR-SYNC-001/acceptance.md
14
```

Arithmetic: 19 − 4 (the doctrine requirements leaving for the sibling) = 15;
+2 for the landed carrier (REQ-1.9, REQ-1.10) = 17; −1 by consolidating the
former REQ-2.6 (human render) and REQ-2.7 (JSON form) into a single REQ-2.6,
since they are one requirement about one surface = **16**. At the Tier M
ceiling, not over it, and reached by consolidating rather than relaxing.

Iteration 2 added AC-014 (14 criteria, ceiling 16) and **changed no requirement
count** — the SPEC remains at 16/16. `plan.md` §C.1 records that there is no
headroom, and that a seventeenth requirement means a tier-up or a further split,
never a relaxed budget.

### Lint — verbatim (D1 closure)

```
$ moai spec lint .moai/specs/SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md --json
[]
EXIT=0
```

An empty array. This is the first run in which the file parsed, so it is also
the first run in which any other lint rule was actually evaluated on this SPEC.

### Measurements observed in this worktree (not relayed)

Landed-check controls (REQ-1.9, AC-011):

```
$ git log origin/main --perl-regexp --grep='\bt199\b' --oneline
b4b8bdfbe docs: update CHANGELOG for v3.1.3
711bfdbba merge(t199): internal/web 자기-SIGTERM TOCTOU — 시그널 등록을 바인드 앞으로
d9899f437 fix(web): register signal handling before binding the listener (t199)

$ git log origin/main -E --grep='\bt199\b' --oneline
                                                    (empty — the ERE trap)

$ git log origin/main --perl-regexp --grep='\bt205\b' --oneline
                                                    (empty — not landed)
```

Carrier data backing the rebuilt AC-002 / AC-003 fixtures:

```
$ gh pr list --state open --limit 40 --json number,title,body \
    -q '.[] | "\(.number)|TITLE:…|BODY:…"'
1614|TITLE:t203|BODY:t1 t151 t203 t69 t9
1612|TITLE:t200|BODY:t200 t201
1611|TITLE:|BODY:t201
1601|TITLE:|BODY:t188
1600|TITLE:|BODY:t184
```

`t201` appears in two bodies (ambiguous); `t188` in exactly one (inferred);
`t151` in exactly one, which is why the prior AC-003 could never produce
`ambiguous`.

AC-013 zero baseline (recorded before implementation):

```
$ grep -c 'moai todo pr' .claude/skills/moai/workflows/todo.md
0
$ grep -c 'moai todo pr' internal/template/templates/.claude/skills/moai/workflows/todo.md
0
```

### Audit iteration 2 — PASS, and the four lead-verified repairs

`.moai/reports/t210/verdict-2.md` returned **PASS** on both SPECs: 0.801 against
the Tier M threshold of 0.80 here, 0.775 against 0.75 on the sibling. Neither
margin is comfortable and the Tier M iteration ceiling (2) is reached, so the
four blocking findings were repaired and **verified by the lead** rather than
sent through a third audit round.

| Finding | Disposition |
|---|---|
| N2 — no NFR had any criterion; NFR-1 is the sole justification for REQ-2.5's dedicated-verb ruling | Fixed. **AC-014** added: exactly one `gh` process regardless of queue length, asserted at two queue lengths (≥3 and 10) so the census observes invariance rather than a coincidence. NFR-2 rides the same assertion (landed check spawns `git`, never `gh`). Without it, a per-card implementation passed all 13 other criteria while costing 0.878s × queue length — passing the SPEC while defeating what the SPEC protects. |
| N1 — `acceptance.md` §D.2 claimed "every criterion exercises at least one requirement" while AC-013 appeared nowhere in the table | Fixed via route (b). The claim is restated honestly, AC-013 has a row mapping it to the Template-First constraint in `plan.md` §D, and the text records that the obligation is carried by a plan constraint rather than a requirement. Route (a) — a normative mirror requirement — is structurally better but breaches 16/16; it is logged as a follow-up in `spec.md` §H and in `plan.md` §C.1. |
| N3 — the split dropped the template-mirror requirement, orphaning the sibling's AC-004 | Fixed in `SPEC-KANBAN-PR-CARD-TRACEABILITY-001`: REQ-005 restored, AC-004 mapped to it, and a traceability table added. Tier S is at 5 of 8, so no budget obstacle. |
| N4 — a judgement sat inside AC-003's `**Mechanical.**` block in the doctrine SPEC | Fixed. The eight words were deleted; the four-carrier claim was already stated correctly in the judgement half, so nothing was lost. Flagged blocking because it was a relapse of D12 inside D12's own remedy. |

### Run-phase debt (N5-N9 — recorded, not fixed)

Carried forward deliberately. None is blocking; each is a real observation.

- **N5 — the fixture block in `acceptance.md` is an unmarked excerpt.** The
  header says the fixtures were re-measured with a `gh … -q` command that emits
  one line per open PR unconditionally; the block carries 5 lines where the
  12-PR set would give 12. No AC fixture is refuted — t188, t201, and t200 were
  re-verified against current data — and 5 PRs suffice for every criterion, so
  this is a labelling defect rather than a correctness one.
- **N6 — AC-011 and AC-012 impose opposite observability demands on the same
  value.** AC-011 needs the underlying query's non-empty commit set observed;
  AC-012 forbids a public accessor returning one. Both are satisfiable — the
  test observes the git-query helper one layer below the resolver's public
  surface — but neither the SPEC nor `plan.md` M2 says so, and M2's "returning a
  boolean and nothing else" reads as closing the door AC-011 needs. **Run-phase
  should resolve this by testing the helper layer, not by weakening AC-011 to a
  boolean check or adding the accessor AC-012 forbids.**
- **N7 — a card that both carries an open PR and is already landed reports
  `linked` only.** REQ-1.9 runs the landed query only while no open PR carries
  the token, and REQ-1.7 permits exactly one outcome kind, so the landed fact is
  masked. This may be the right design (one question, one answer), but §F does
  not name it as a boundary. **This is one of the two candidates that would
  become a seventeenth requirement — see `plan.md` §C.1.**
- **N8 — AC-012's API-surface half carries no verb and is not labelled a
  judgement.** It is mechanizable by reflection over the exported record's
  fields; it should either name that mechanism or match the judgement-labelling
  pattern AC-013 and the doctrine SPEC's AC-001/AC-002 use.
- **N9 — two requirement-id schemes ship on one card.** This SPEC uses
  `REQ-1.1`-style ids; the sibling uses canonical `REQ-001`. Grep tooling keying
  on `REQ-[0-9]{3}` matches one and not the other. `moai spec lint` returns `[]`
  on both, so nothing mechanical breaks. Renumbering touches 16 headings, 16
  traceability rows, and every back-reference — deferrable, and deferred.

### Gaps

- The AC fixtures are pinned from a PR set that has already changed since M2 was
  taken (the open set has grown to 12). `plan.md` §B makes pinning mandatory;
  a run-phase that re-fetches would invalidate them again.
- The reviewer-judgement half of AC-013 has no command, and is recorded as a
  judgement rather than reported as a mechanical pass.
- Merged-PR attribution precision remains unscored (§F). Only the landed
  question is covered.

### Signal

`status: draft`. Awaiting audit iteration 2 of 2, then Implementation Kickoff
Approval. Sequenced after `SPEC-KANBAN-PR-CARD-TRACEABILITY-001`.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
