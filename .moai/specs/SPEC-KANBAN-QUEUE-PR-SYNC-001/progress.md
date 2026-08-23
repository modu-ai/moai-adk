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
13
```

Arithmetic: 19 − 4 (the doctrine requirements leaving for the sibling) = 15;
+2 for the landed carrier (REQ-1.9, REQ-1.10) = 17; −1 by consolidating the
former REQ-2.6 (human render) and REQ-2.7 (JSON form) into a single REQ-2.6,
since they are one requirement about one surface = **16**. At the Tier M
ceiling, not over it, and reached by consolidating rather than relaxing.

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
