# SPEC-TODO-AUTO-PICK-001 — compact (run-phase load)

Autonomous card selection under `--auto`: the invoked session judges, the lease is the only pick
path, the keep-set is skipped and reported. Tier M, card t1448, plan-start HEAD `4bf547bca`.
Full text: `spec.md`; evidence: `research.md`; criteria with RED-now cells: `acceptance.md`.

## Requirements (GEARS)

**Module A — selection authority**
- REQ-TAU-001 When `--auto` is authorized (typed, or a lane's self-service pickup), the invoked
  session shall choose the next card on its own judgment and not ask the operator.
- REQ-TAU-002 The `moai factory next` lease (CLI/MCP) shall be the only lane pick path; bare =
  CLI priority order, nominated = the session's judged choice.
- REQ-TAU-003 Admission stays the operator's, schemas untouched: no lane queue mutation, no
  admission by `--auto`, Jev never chooses, the PR cross-check reports-never-vetoes an operator
  pick and is an input for a self-chosen card; no schema/relation-kind/`moai graph`/verb change.

**Module B — the nominated lease**
- REQ-TAU-004 Where `--card <id>` (MCP `card`) is given: validate and lease atomically through the
  same version-checked edges, honoring quota hold and Codex skip; refuse any other card.
- REQ-TAU-005 When a nominee is refused or the race is lost: distinct exit (not 0/1/3), one stderr
  line with a closed reason token, no state change; the session re-selects.
- REQ-TAU-006 Two lanes nominating different cards each hold their own; the same card, exactly one
  holder and one refusal.
- REQ-TAU-007 Without `--card`: unchanged for every queue with no `[보류`-opening queued card; for a
  queue with one, the promoting arm skips it (the single default-path delta, D-DEF).
- REQ-TAU-008 While the lane-refusal predicate holds, `moai todo --auto` is refused with the
  lane-boundary text naming `moai factory next`; queue byte-identical (D-LANE).

**Module C — the keep-set**
- REQ-TAU-009 A nominated lease refuses: not-in-queue, `hold`, `dropped`, owned-by-other, a queued
  card whose text opens with `[보류`, `blocked` classification, a serial card while the serial slot
  is held — each with its own token.
- REQ-TAU-010 When a candidate's text hinges on an operator confirmation (payments, secrets,
  irreversible external-shared work), the session does not nominate it, reports it skipped with the
  reason, and changes no queue state.
- REQ-TAU-011 The operator-decision queue is expressed by the existing `hold` state and `[보류`
  marker only; no new marker/state/field/verb; a card marked by neither is judged by REQ-TAU-010.

**Module D — record and open inputs**
- REQ-TAU-012 When a lane leases under `--auto`, it writes one `decision record:` line (§10 form)
  with `ladder_path=gate-row card pick (AUTONOMOUS, auto-semantics §9)` and `evidence_refs`
  naming card, class, relation records, PR/landing, worktree presence, file overlap, skipped
  candidates; unreadable inputs read `unmeasured`, never omitted, never `none`.
- REQ-TAU-013 The input set is open (a later card adds an input without amending the keep-set or
  the lease path); relation records are not bound to one store; file overlap is pluggable with a
  fallback (in-flight lane branch paths) and then `unmeasured`.

**Module E — surfaces**
- REQ-TAU-014 Replace (not append) the named sentences in `kanban-dispatch.md`, `gtd.md`,
  `auto-semantics.md`, `manager-todo.md`, the foreman skill; carry the § C.1 contract literals;
  move the doc-pin tests in the same milestone.
- REQ-TAU-015 Edited files stay byte-identical to mirrors, except the pre-existing line-177 drift in
  `kanban-dispatch.md`, preserved exactly and reported.
- REQ-TAU-016 No growth of the always-loaded `kanban-dispatch.md` (live 26,959 B / mirror 26,637 B).

## Acceptance (Given / When / Then, condensed)

- AC-TAU-001 Given queued t1..t3, when `factory next --card t2`, then t2 (not t1) is leased with the
  unnominated output shape; `--card t9` exits with the refusal code, token `unknown-card`, nothing
  changed. RED-now L1/L2.
- AC-TAU-002 Two lanes nominating t1/t2 hold one each; both nominating t1 end with exactly one
  holder, the other refused (`raced`/`owned`), re-nomination of t2 succeeds. RED-now L2.
- AC-TAU-003 Under `--auto` a lane chooses through the lease and records a decision record; no
  queue mutation verb runs. RED-now L5/L9.
- AC-TAU-004 Nominating `hold` / `[보류` / `blocked` / held-slot serial cards is refused with its
  token and changes nothing; the bare path never leases hold/marker/blocked cards. RED-now L2.
- AC-TAU-005 A lane's `moai todo --auto` is refused with the lane-boundary text, queue
  byte-identical; bare `moai todo`/`list` still run. RED-now L3.
- AC-TAU-006 Bare `factory next` is golden-unchanged on a marker-free queue. RED-now L2 (golden
  absent); existing five lease pins stay green.
- AC-TAU-007 Contract literals present / old literals absent on all five docs, live and mirror
  (mutants: old sentence left, pin not moved). RED-now L5-L9.
- AC-TAU-008 `ladder_path=gate-row card pick`, the input list, `unmeasured`, the file-overlap
  fallback and the open-set sentence present; no store named as THE relation source. RED-now L10.
- AC-TAU-009 (regression-guard) the first lane lease writes the record line, marked self-attested.
- AC-TAU-010 Mirrors byte-identical except exactly the one preserved line-177 drift (`1 1`).
- AC-TAU-011 Old sentence absent from both `kanban-dispatch.md` copies AND `wc -c` ≤ baseline.
- AC-TAU-012 The diff touches nothing under `internal/kanban/**`, `internal/graph/**`, or any
  schema file.

## Files to modify

Go: `internal/cli/factory_card.go`, `internal/cli/mcp_factory_card.go`, `internal/cli/todo.go`;
tests (new): `internal/cli/factory_nominate_test.go`, `internal/cli/todo_auto_pick_doc_test.go`;
tests (edit): `internal/cli/todo_auto_doc_test.go`. Docs, each live + template mirror:
`kanban-dispatch.md`, `kanban-dispatch-detail.md`, `auto-semantics.md`, `gtd.md`, `manager-todo.md`,
`moai-kanban-foreman/SKILL.md`, `moai-mcp-tools-catalogue.md`.

## Exclusions

Out of Scope — queue admission and card production; schemas, relation kinds, `moai graph`, and
the t1454 inputs; Jev and the ranking exception; the lane bootstrap notice (four locales); a
completed SPEC's body; absorbing the line-177 drift; a decision-board writer.
