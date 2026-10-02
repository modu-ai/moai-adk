# SPEC-TODO-AUTO-PICK-001 — compact (run-phase load)

Autonomous card selection under `--auto`: the invoked session judges, the lease is the only pick
path, the keep-set is skipped and reported. Tier M, card t1448, plan-start HEAD `4bf547bca`,
version 0.2.0 (plan-audit iteration 1 repair on `b3646de10`). Full text: `spec.md`; evidence:
`research.md`; criteria with RED-now cells: `acceptance.md`; resolution map: `plan.md` §10.

## Requirements (GEARS) — 16

**A — selection authority**
- REQ-TAU-001 When `--auto` is authorized, the invoked session shall choose the next card on its
  own judgment and not ask the operator *(doctrine-only)*.
- REQ-TAU-002 The `moai factory next` lease (CLI/MCP) shall be the only lane pick path; bare =
  CLI priority order, nominated = the session's judged choice.
- REQ-TAU-003 Admission stays the operator's, no schema/relation-kind/`moai graph`/verb change;
  only CLI additions: `--card`, MCP `card`, the lane refusal of `--auto`.

**B — the nominated lease**
- REQ-TAU-004 When `--card <id>` (MCP `card`) is given: validate before any write, lease atomically
  through the same version-checked edges, honor quota hold and Codex skip; refuse any other card.
- REQ-TAU-005 When a nominee is refused or the race lost: exit 4, one stderr line
  `factory next: refused <token>: <detail>` (closed set, spec § C.2), queue and record as before
  the invocation (a promotion the invocation made is undone if the claim fails with no other holder).
- REQ-TAU-006 Different cards → one holder each; the same card → exactly one holder, one refusal.
- REQ-TAU-007 Without `--card`: unchanged for queues with no `[보류`-opening queued card; otherwise
  the promoting arm skips such a card and counts it as seen (all-marker queue → exit 3) — D-DEF.
- REQ-TAU-008 While the session is a lane session (role marker `lane` or non-empty lane label; the
  Codex marker alone not sufficing): `moai todo --auto` is refused with a dedicated text naming
  `moai factory next --card <id>`, queue byte-identical; `factory fallback declare` prints the
  lease path; REQ-FLA-001's `/moai:todo --auto` is superseded in lane sessions — D-LANE.

**C — the keep-set**
- REQ-TAU-009 A nominated lease refuses: absent, `hold`, `dropped`, owned-by-other, `[보류`-opening,
  `blocked`, held-slot serial — each with its token.
- REQ-TAU-010 The session does not nominate a card whose text hinges on payments/secrets/
  irreversible external-shared work, nor a `queued` card with an open PR / landed fix; it reports
  each as skipped; for an operator-`picked` card PR/landed state is report-only *(doctrine-only)*.
- REQ-TAU-011 The operator-decision queue is the existing `hold` state and `[보류` marker only; no
  new marker/state/field/verb; the marker demotes in the serial cycle and excludes on the lease path.

**D — record and open inputs**
- REQ-TAU-012 When a lane leases under `--auto`: one `decision record:` line in the card's progress
  record (evidence, not the §11 board), `ladder_path=gate-row card pick (AUTONOMOUS,
  auto-semantics §9)`, `evidence_refs` naming the inputs; unreadable → `unmeasured`.
- REQ-TAU-013 The input set is open (adding an input amends neither keep-set nor lease path);
  relation records not store-bound; file overlap pluggable, fallback inferred, `unmeasured` expected.

**E — surfaces**
- REQ-TAU-014 Replace (not append) the named sentences in the five docs plus one detail paragraph,
  live and mirror in one change, carry the § C.1 literals (the gtd.md section holds the keep-set
  list, record form and lane routing themselves), bound each diff (AC-TAU-013), move the two pin
  markers (`TestAutoRankDoctrineAmendment` sentence; `TestAutoRankMirrorParity` start marker).
- REQ-TAU-015 Mirrors byte-identical except the pre-existing line-177 drift (preserved, reported).
- REQ-TAU-016 No growth of the always-loaded `kanban-dispatch.md` in bytes or characters (live
  26,959 B / 26,754 chars; mirror 26,637 B / 26,433 chars).

## Refusal tokens (closed set)

`unknown-card`, `dropped`, `held`, `owned`, `hold-marker`, `blocked`, `serial-slot`, `quota-hold`,
`backend-skip`, `raced`.

## Acceptance (Given / When / Then, condensed) — 14

- AC-TAU-001 `--card t2` leases t2; `--card t9` → exit 4 `unknown-card`, nothing changed; `--help`
  flag set = baseline + `--card`; MCP `card` parity. RED-now L1/L2/L15. *(release-blocking)*
- AC-TAU-002 Two lanes, different cards → one each; same card → exactly one holder; satisfies card
  item 6 through the nominated path. RED-now L2.
- AC-TAU-003 Docs authorize the session's own judgment and route a lane to the lease. RED-now
  L5/L9/L19.
- AC-TAU-004 hold/dropped/marker/blocked/serial-slot/owned nominations refused with their token;
  bare path never leases hold/dropped/marker/blocked; text-judgement and PR-skip doctrine-only.
- AC-TAU-005 A label-only or role-marker lane running `moai todo --auto` is refused (dedicated text,
  queue byte-identical); a non-lane GPT-backend session is NOT refused (C4); bare `moai todo`/`list`
  still run; fallback string and gtd routing sentence corrected. RED-now L18/L16/L17.
- AC-TAU-006 Bare `factory next` golden-unchanged (min case set: arm order, `--wait`, quota hold,
  Codex skip, serial slot, exit 3) with a seeded-perturbation cell; all-marker queue exits 3.
- AC-TAU-007 Contract literals present / old literals absent, live and mirror in one commit, pins
  moved. RED-now L5-L9, L19, L20.
- AC-TAU-008 Record form, location (evidence, not the board), `unmeasured`, open-set sentence.
- AC-TAU-009 *(regression-guard)* the first lane lease writes the record, self-attested, unread.
- AC-TAU-010 *(regression-guard)* mirrors byte-identical except the one preserved drift (`1 1`).
- AC-TAU-011 Old sentence absent AND `wc -c` and `wc -m` ≤ baselines, both copies.
- AC-TAU-012 *(regression-guard)* the diff touches nothing under `internal/kanban/**`,
  `internal/graph/**`, or any schema file.
- AC-TAU-013 Each edited file's `git diff --numstat` within [1, cap]; whole-file reflow fails.
  RED-now L21.
- AC-TAU-014 Nominee state: `quota-hold`, `backend-skip`, refusals leave state byte-identical,
  promote-then-lose leaves the winner's state, injected claim refusal rolls the promotion back.

## Files to modify

Go: `internal/cli/factory_card.go`, `internal/cli/mcp_factory_card.go`, `internal/cli/todo.go`,
`internal/cli/factory_messaging.go`; tests (new): `internal/cli/factory_nominate_test.go`,
`internal/cli/todo_auto_pick_doc_test.go`; tests (edit): `internal/cli/todo_auto_doc_test.go`.
Docs, each live + template mirror in one change: `kanban-dispatch.md`, `kanban-dispatch-detail.md`,
`auto-semantics.md`, `gtd.md`, `manager-todo.md`, `moai-kanban-foreman/SKILL.md`,
`moai-mcp-tools-catalogue.md`.

## Exclusions

Out of Scope — queue admission and card production; schemas, relation kinds, `moai graph`, and the
t1454 inputs; Jev, ranking, and the audit surfaces (no sync-audit change); the lane bootstrap notice
(four locales); a completed SPEC's body; absorbing the line-177 drift; a decision-board writer.
