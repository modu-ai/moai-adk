# SPEC-TODO-AUTO-PICK-001 — Plan

Tier M. Card t1448, plan-start HEAD `4bf547bca`, worktree `.moai/worktrees/t1448`, branch
`WT-todo-auto-pick-autonomy`. Milestones are ordered by **decision reversibility** — the decisions
most likely to change come first (the nominated-lease interface a lane sees, then the doctrine the
operator reads), mechanical edits last. No time estimates; priority labels and ordering only.

## 1. Approach in one paragraph

Add one opt-in nomination form to `moai factory next`, validated and leased atomically, with one
shared keep-set predicate; refuse `moai todo --auto` inside a lane so the lease is the only lane
pick path; then amend the doctrine sentences the card names — **replacing** them, never appending
— and move the doc-pin tests that assert the old wording in the same milestone. All work is
TDD-first: each milestone opens with tests observed RED on this tree. Net bytes on the always-loaded
`kanban-dispatch.md` must not grow.

## 2. Milestones

### M1 — The nominated-lease interface, tests first (Priority High)

The decision most likely to change: the flag name, the refusal tokens, the exit code.

- Write the new tests in `internal/cli/factory_nominate_test.go` against the existing fixtures
  (`fcFixture`, `fcQueue`, `fcClassify`, `sdRegisterLane`, `sdLaneEnv`, `runFactory`), **scrubbing
  the lane environment first** (`sdClearLaneEnv`), and observe each RED for the stated reason (the
  flag does not exist: `runFactory(t, "next", "--card", …)` fails with `Unknown flag: --card`, not
  a compile error). Test set (names are the swept set of `acceptance.md` AC-TAU-001, -002, -004, -005, -006):
  `TestFactoryNextNominateLeasesNominee`, `TestFactoryNextNominateRefusesKeepSet` (subtests
  `hold-state`, `hold-marker`, `blocked`, `serial-slot`, `owned`, `unknown-card`, `not-queued`),
  `TestFactoryNextNominateQuotaHold`, `TestFactoryNextNominateConcurrentLanes`,
  `TestFactoryNextNominateSameCardExactlyOne`, `TestFactoryNextBareUnchanged` (golden over a
  marker-free queue), `TestFactoryNextArmCSkipsHoldMarker`, `TestFactoryNextNominateMCPParity`,
  `TestTodoLaneRefusesAutoCycle`.
- Fix the closed reason-token set and the refusal exit code in the test expectations: tokens
  `unknown-card`, `not-queued`, `held`, `hold-marker`, `blocked`, `serial-slot`, `owned`,
  `quota-hold`, `raced`; the exit code is **4** (`factoryNextRefusedExit`), distinct from the
  no-card `3` and from `1`. **Pre-flight:** measure that 4 is free in the `factory` verb family
  (research R10); choose the next free code if not.
- Exit: every new test observed RED on the unmodified tree for the stated reason; the baseline of
  R3 re-run and still green.

### M2 — Implement nomination and the shared keep-set predicate (Priority High)

- `internal/cli/factory_card.go`: add the `--card` flag to `newFactoryNextCommand`; add a
  nomination path beside `factoryNextSelectAndLease` that reuses `factoryNextClaim` and
  `factoryNextRecordAndClaim` (the version-checked edges — it must not open a second lease route);
  extract one keep-set predicate `factoryKeepSetRefusal` that returns the refusal token, applied by
  the nomination path **and** by arm (c) for the `[보류` skip only (REQ-TAU-007).
- Nominee states: a `queued` card is promoted to `picked` inside the same `Mutate` (as arm (c)
  does) after the predicate passes; a `picked` card with no owner (operator pick, arms (b)/(b2)) is
  leasable; a card that is `hold`, `dropped`, owned by another lane, or absent is refused. The
  quota gate (`noNewCards`) and the Codex backend skip (`factoryNextSkipForBackend`) bind the
  nominee exactly as they bind the unnominated new-card arms.
- Refusal: stderr one line `factory next: refused <token>: <detail>`, nothing on stdout, exit 4,
  no queue or record change.
- Exit: M1's nominate/bare/arm-c tests GREEN; the five existing pinned tests of R3 still green; the
  bare-path golden unchanged for marker-free queues.

### M3 — Refuse `moai todo --auto` in a lane (Priority High; isolated — REQ-TAU-008)

- `internal/cli/todo.go` `todoRefuseLaneMutation`: refuse when `todoAutoFlag` is set while
  `factoryLaneRefusal()` holds, with the existing `todoLaneMutationRefusalText` (it already names
  `moai factory next`). The bare parent render (`len(args) == 0`, no `--auto`) stays allowed.
- Exit: `TestTodoLaneRefusesAutoCycle` GREEN (queue bytes identical before and after, per
  `sdQueueBytes`); the existing lane-refusal tests green. Dropping M3 removes only REQ-TAU-008 and
  AC-TAU-005.

### M4 — The MCP parameter (Priority Medium)

- `internal/cli/mcp_factory_card.go`: `factory_next` gains an optional `card` string and calls the
  same nomination implementation ("same implementation, same record changes, same refusals").
  `internal/mcp/catalog.go` carries only name and write-capability — **no change expected**
  (measured, `catalog.go:112`); the run phase re-confirms with the catalog equality guard.
- Update the one catalogue row in `moai-mcp-tools-catalogue.md` (live and mirror) to mention the
  optional `card` argument — one cell, parity-checked.
- Exit: `TestFactoryNextNominateMCPParity` GREEN.

### M5 — The doctrine amendment, pins first (Priority High)

Order inside the milestone: (1) write `TestAutoPickDocDoctrine` and `TestAutoPickMirrorParity` in
`internal/cli/todo_auto_pick_doc_test.go` — observe RED (AC-TAU-007/-008 RED-now cells); (2) edit
the docs; (3) move the existing pins in `internal/cli/todo_auto_doc_test.go` to the new wording in
the **same commit** as the docs (R6): the `kanban-dispatch.md` prohibition sentence changes from
`The leader never picks for the operator, …` to its scoped form; every other pinned sentence
stays.

Edits, each a **replacement** of a named sentence (surgical — lane-5's card t1453 absorbs these, so
no reflow, no heading moves, no list restructuring):

| File | Replace | With (the § C.1 literals; other wording free) |
|---|---|---|
| `kanban-dispatch.md` L29 | `**Promotion is the operator's act, always.**` and the unscoped `The leader never picks for the operator, …` | promotion is the operator's act, in person or in advance through `--auto`; `Outside an --auto authorization the leader never picks for the operator, …` |
| `kanban-dispatch.md` L31 | `authorized serial consumption of the queue and nothing else` (and the lead-in `The one reconciliation is named, not excepted:`) | authorizes the invoked session to take cards from the queue on its own judgment, each only through a lease and never a keep-set card, and nothing else |
| `kanban-dispatch.md` L33 | `only through \`moai factory next\`, whose lease lands in the factory record the way a leader's dispatch lands in the queue.` | `only through \`moai factory next\` — bare, or \`--card <id>\` for the lane's own judged pick.` |
| `gtd.md` `--auto` section (L334-L336) | the `[HARD] … authorizes serial consumption of the queue and nothing else` clause | the same clause with `on its own judgment`, the lane/lease path (`factory next --card`), and a pointer to `auto-semantics.md` §9.3 (keep-set, record); the "exactly one card is in flight" sentence is scoped to the operator-session cycle |
| `auto-semantics.md` §9 card-pick row | the row text | keeps the AUTONOMOUS disposition, adds the pointer `§9.3` |
| `auto-semantics.md` new `### 9.3` | — | the keep-set (mechanical: `hold` state, `[보류` marker, `blocked`, serial slot, owned; text-judgement: payments, secrets, irreversible external-shared work), the input set, the record form with `ladder_path=gate-row card pick (AUTONOMOUS, auto-semantics §9)`, the open-input sentence |
| `manager-todo.md` L23, L34-L35 | `process cards in queue order`; `serial consumption of the queue and nothing else` | `on its own judgment` / keep-set wording; the serial-cycle region keeps the pinned ranking literals |
| `moai-kanban-foreman/SKILL.md` Boundary 1, step 4 | `serial consumption in queue order is authorized`; `\`queued\` items are not yours to pick.` | `… on the iteration's own judgment, within the keep-set`; `… not yours to pick outside a batch authorization.` |
| `kanban-dispatch-detail.md` § The pre-dispatch cross-check | — | one added paragraph: the reconciliation of B.2 (operator-picked card: reports, never vetoes; autonomously chosen card: input, skipped and reported). Lazy file — growth is free |

New detail goes in `auto-semantics.md` §9.3 (lazy, `paths:`-scoped to the watchdog skill),
`gtd.md`, and `kanban-dispatch-detail.md` — **never** into the always-loaded stub beyond the
replaced sentences.

- Exit: the new doc pins GREEN; the moved existing pins GREEN; the mutants of `acceptance.md` §M
  each fail their criterion.

### M6 — Mirrors, parity, budget (Priority High; mechanical)

- Apply the identical edits to every `internal/template/templates/.claude/...` mirror.
- **Mirror handling (preserve, do not absorb).** The live and mirror `kanban-dispatch.md` differ by
  one pre-existing line (live line 177, the `moai worktree sweep …` sentence, added to the live
  copy only by `62197bb38`). Preserve each copy's line 177 exactly; do not sync it either way.
  Reason: absorbing it is another card's act, it would put an unreviewed sentence into every user
  project's always-loaded rule, and the existing parity test is deliberately scoped to the amended
  passage. After the edits `git diff --no-index --numstat` of the pair must still read `1  1`
  (AC-TAU-010). The completion report names the drift for a follow-up card.
- **Always-loaded growth constraint (REQ-TAU-016).** `kanban-dispatch.md` is always-loaded, the repo
  is already over its instruction budget (t1318: 214,155 chars against 210k), and `rule-authoring.md`
  requires a statement for growth over 1,000 B. This SPEC requires **net growth ≤ 0 B on both
  copies**: baseline live 26,959 B, mirror 26,637 B (pinned tree). Measuring command, both
  copies: `wc -c .claude/rules/moai/workflow/kanban-dispatch.md
  internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md`. Feasibility
  (research R7): the three replaced paragraphs measure +72, −31, −30 = +11 B before trimming;
  trimming at least 11 B (e.g. `in the gate inventory`, `Serial-contract detail` wording) reaches ≤ 0.
  The commit body carries the measured before/after bytes and the non-invoking-cost statement
  (`rule-authoring.md` (c)): a session that never uses `--auto` pays zero added bytes.
- Run the repo's template embed refresh if the Makefile defines one (pre-flight; research R10).
- Exit: AC-TAU-010 and AC-TAU-011 GREEN with their contrast controls and mutants.

### M7 — Verification, scoped (Priority High)

Run only what the change can affect, then let CI run the rest (`AGENTS.md` §4):

- `internal/cli`: the new tests plus `^(TestFactoryNext|TestAutoRank|TestAutoPick|TestTodoLane)`.
  Take the `moai slot` lease for the `internal/cli` package run
  (`.claude/rules/local/gitflow-lane-protocol.md` §8), and scrub the lane environment in **one
  compound invocation** (`unset MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS
  MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND MOAI_KANBAN_ID
  MOAI_KANBAN_SETTINGS_INJECTED && go test …`).
- `internal/template`: the tests that name the edited files — `jev_auto_exception_test.go`,
  `contract_mode_guided_test.go`, `rule_template_mirror_test.go` — by `-run` selector with a
  swept-count check.
- `GOOS=windows GOARCH=amd64 go build ./...` (cross-platform), `golangci-lint` at the CI version
  (`feedback_lane_lint_must_use_ci_golangci_version`).

## 3. Files to modify

| Area | File | Change |
|---|---|---|
| Go | `internal/cli/factory_card.go` | `--card` flag, nomination path, `factoryKeepSetRefusal`, arm (c) `[보류` skip, refusal exit/diagnostic |
| Go | `internal/cli/mcp_factory_card.go` | optional `card` parameter, shared implementation |
| Go | `internal/cli/todo.go` | lane refusal of `--auto` in `todoRefuseLaneMutation` |
| Go test (new) | `internal/cli/factory_nominate_test.go` | nomination, keep-set, concurrency, bare golden, arm (c), MCP parity, lane `--auto` |
| Go test (new) | `internal/cli/todo_auto_pick_doc_test.go` | doc contract literals, stale-absent, mirror parity, mutants |
| Go test (edit) | `internal/cli/todo_auto_doc_test.go` | move the pinned `kanban-dispatch.md` prohibition to the scoped sentence |
| Rule (live + mirror) | `kanban-dispatch.md` | replace the three paragraphs (net ≤ 0 B) |
| Rule (live + mirror) | `kanban-dispatch-detail.md` | one reconciliation paragraph |
| Rule (live + mirror) | `auto-semantics.md` | §9 row pointer + new §9.3 |
| Skill (live + mirror) | `.claude/skills/moai/workflows/gtd.md` | `--auto` section sentence replacement |
| Agent (live + mirror) | `.claude/agents/moai/manager-todo.md` | two sentence replacements |
| Skill (live + mirror) | `.claude/skills/moai-kanban-foreman/SKILL.md` | Boundary 1 and step 4 |
| Rule (live + mirror) | `.claude/rules/moai/core/moai-mcp-tools-catalogue.md` | one catalogue cell (optional `card`) |

No file under `internal/kanban/**`, `internal/graph/**`, or any schema file changes (REQ-TAU-003).

## 4. Forward-compatibility with t1454

What this SPEC **leaves open**: the input set is open — the record's `evidence_refs` and the
selection rule name the inputs t1448 needs and say a later card adds an input without amending the
keep-set or the lease path (REQ-TAU-013). Relation records are read through whatever the lane-read
surfaces expose (today `moai todo why`); no code here names the queue-findings store or the
`gtd_relations` table as *the* source. File overlap is a pluggable input: expected-file data when
present; otherwise the paths changed by each in-flight lane's card branch; otherwise `unmeasured`.

What this SPEC **must not pre-empt**: it adds no schema field, no relation kind, no `moai graph`
change, no expected-file, parent/spawned-by or size field, and no mechanical refusal keyed on a
relation store (the relation-blocked refusal is deliberately *not* in the keep-set predicate;
probe O2). The mechanical guard reads only `state`, the `[보류` marker, `classification.blocked`,
the serial slot, and ownership — none of which t1454's inputs change.

## 5. Operational follow-ups (not part of this SPEC's code)

- The operator or leader should `moai gtd hold` (or `[보류`-mark) any card kept on an operator
  decision queue by arrangement (t810, t1294, t1383 per the memory index); until then they are
  selectable.
- The line-177 drift in `kanban-dispatch.md` needs a follow-up card to decide which copy is right.
- The lane bootstrap notice still says bare `moai factory next`; a follow-up candidate is to mention
  the judged `--card` form in the four locales of `session_start_factory_i18n.go`.

## 6. Risks and their milestones

| Risk | Where handled |
|---|---|
| Doc pins red when the docs change | M5 moves docs and pins together; RED-first |
| Default-path `[보류` skip surprises an operator | isolated REQ-TAU-007; stated in B.4 and the return report |
| Lane `--auto` refusal reverses an FLA doctrine move | isolated REQ-TAU-008; lease path is the replacement |
| Self-attested decision record | `spec.md` §G; sync-audit re-read |
| Parallel edit collision with lane-5's t1453 | surgical replacements only; this card's rule-doc edits land first |
| Always-loaded growth | M6 measurement, both copies, ≤ 0 B |

## 7. Commit structure

Plan phase: one commit carrying the artifact set (this commit). Run phase: tests-first commit(s)
per milestone; the baseline measurements of `acceptance.md` land in the plan commit, **before** any
run commit, so the commit graph — not a message — witnesses that the baselines precede the change
(`verification-claim-integrity.md` §2.3).
