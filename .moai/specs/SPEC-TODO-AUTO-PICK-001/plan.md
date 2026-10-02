# SPEC-TODO-AUTO-PICK-001 — Plan

Tier M. Card t1448, plan-start HEAD `4bf547bca`, repair base `b3646de10`, worktree
`.moai/worktrees/t1448`, branch `WT-todo-auto-pick-autonomy`. Milestones are ordered by **decision
reversibility** — the decisions most likely to change come first (the nominated-lease interface a
lane sees, then the doctrine the operator reads), mechanical edits last. No time estimates;
priority labels and ordering only.

## 1. Approach in one paragraph

Add one opt-in nomination form to `moai factory next`, validated before any write and leased
atomically, with one shared keep-set predicate; refuse `moai todo --auto` inside a lane (by an
explicit lane predicate, with a dedicated refusal text and a corrected fallback instruction) so the
lease is the only lane pick path; then amend the doctrine sentences the card names — **replacing**
them, never appending, **live and mirror in one change** — and move the doc pins in that same
change. All work is TDD-first. Net bytes and characters on the always-loaded `kanban-dispatch.md`
must not grow.

## 2. Milestones

### M1 — The nominated-lease interface and the lane predicate, tests first (Priority High)

The decisions most likely to change: the flag name, the refusal tokens, the exit code, the lane
predicate.

- Write the new tests in `internal/cli/factory_nominate_test.go` against the existing fixtures
  (`fcFixture`, `fcQueue`, `fcClassify`, `sdRegisterLane`, `sdLaneEnv`, `runFactory`), **scrubbing
  the lane environment first** (`sdClearLaneEnv`). Test set — each is the swept set of the
  criterion named:
  - AC-TAU-001: `TestFactoryNextNominateLeasesNominee`, `TestFactoryNextNominateUnknownCard`,
    `TestFactoryNextFlagSet` (the `--help` flag set equals the baseline plus `--card`),
    `TestFactoryNextNominateMCPParity` (the MCP form's `card` input; inputs are `run`,
    `project_root`, `card`).
  - AC-TAU-002: `TestFactoryNextNominateConcurrentLanes`, `TestFactoryNextNominateSameCardExactlyOne`.
  - AC-TAU-004: `TestFactoryNextNominateRefusesKeepSet` (subtests `held`, `hold-marker`, `blocked`,
    `serial-slot`, `dropped`, `owned`), `TestFactoryNextArmCSkipsHoldMarker`.
  - AC-TAU-014: `TestFactoryNextNominateQuotaHold`, `TestFactoryNextNominateBackendSkip`,
    `TestFactoryNextNominateRefusalLeavesStateUnchanged` (byte-compares queue and record),
    `TestFactoryNextNominatePromoteThenLose`, `TestFactoryNextNominateClaimRefusedRollsBack`.
  - AC-TAU-005: `TestTodoLaneRefusesAutoCycle` (subtests `label-only`, `role-only`,
    `role-and-label`, each asserting the queue byte-identical), `TestTodoLaneAutoRefusalText`,
    `TestTodoNonLaneGPTSessionNotRefused`, `TestFactoryFallbackDeclarePrintsLeasePath`.
  - AC-TAU-006: `TestFactoryNextBareUnchanged` (the golden) and
    `TestFactoryNextAllMarkerQueueExitsNoCard`.
- **M1 exit (D2).** (a) Every new test **except** the golden and the non-lane GPT guard is observed
  RED on the unmodified tree for its stated reason (the flag does not exist, the string is still
  there, the lane runs the cycle, the marker card is leased) — a runtime failure, not a compile
  error. (b) `TestFactoryNextBareUnchanged` and `TestTodoNonLaneGPTSessionNotRefused` are GREEN on
  the unmodified tree (they pin current behavior and must stay green). (c) Their non-vacuity is
  shown by a **seeded perturbation** recorded in `progress.md` §E.2 with command, verbatim stdout,
  exit code and tree SHA: for the golden, a one-line mutation of `factory_card.go` (for example
  moving the `noNewCards` early return, or swapping the arm (b)/(c) order); for the GPT guard, the
  over-broad predicate (`factoryLaneRefusal()` in the new guard). Each mutation is then reverted
  (`git diff` empty) before M2. (d) The scoped baseline of R3/R11 re-run and still green.
- **The golden's minimum case set** (also in AC-TAU-006): default arm order (a)→(b)→(b2)→(c) over
  assigned / operator-picked / queue-picked / queued cards; `--wait` with a bounded wait through
  the `factoryNextWaitSleep` seam; the quota hold (`noNewCards`) with an assigned card still
  leasing; the Codex skip (`factoryNextSkipForBackend`) of a card at or past merge-ready; the serial
  slot (a second serial card not leased while one is in flight); and the no-card exit 3 with its
  stdout line.
- The refusal exit code is **4** (`factoryNextRefusedExit`). Pre-flight: measured, within
  `factory*.go` only 1 and 3 are used (`factoryNextNoCardExit = 3`, `factory_lane_relaunch.go:51`);
  `moai slot` also uses 4 (`slotExitBusy = 4`, `slot.go:38`) — a different verb family, so exit
  codes do not collide; the run phase re-greps before choosing.

### M2 — Implement nomination and the shared keep-set predicate (Priority High)

- `internal/cli/factory_card.go`: add the `--card` flag to `newFactoryNextCommand`; add a
  nomination path beside `factoryNextSelectAndLease` that reuses `factoryNextClaim` and
  `factoryNextRecordAndClaim` (the version-checked edges — no second lease route); extract one
  keep-set predicate `factoryKeepSetRefusal` returning the § C.2 token, applied by the nomination
  path **and** by arm (c) for the `[보류` skip only (REQ-TAU-007).
- **State semantics (D7), the mechanism.** The nominated path runs four steps in order:
  1. *Validate, read-only.* One pure queue read plus one factory-record read decide every § C.2
     token — `unknown-card`, `dropped`, `held`, `owned`, `hold-marker`, `blocked`, `serial-slot`,
     `quota-hold`, `backend-skip` — and run the foreign-worktree precheck that `factoryNextClaim`
     would otherwise run only after a promotion (`factoryRefuseForeignWorktree`). No write has
     happened when any of these refuses.
  2. *Promote inside the queue `Mutate`, re-validating.* A `queued` nominee is promoted to `picked`
     only if, inside the lock, its state is still `queued` and the predicate still passes; if it is
     no longer `queued` another lane moved it first and the invocation refuses `raced` without
     writing. A nominee that is already `picked` and unowned (operator pick, arms (b)/(b2)) skips
     this step; a nominee assigned to this lane is arm (a)'s.
  3. *Claim* through `factoryNextRecordAndClaim` / `factoryNextClaim`.
  4. *Compensate.* If the claim is refused or lost and the factory record shows **no other holder**
     for the card, the invocation restores the queue item from `picked` to `queued` in one `Mutate`
     (it undoes only the promotion it made); if another holder exists, the queue state is that
     holder's and is left alone, and the invocation refuses `raced`.
  A package-level function variable in the style of `factoryCardNow` marks the point between
  steps 2 and 3 so `TestFactoryNextNominatePromoteThenLose` and
  `TestFactoryNextNominateClaimRefusedRollsBack` can interleave a competing lease and an injected
  claim refusal.
- The arm (c) skip counts the marker card as seen (`sawQueued`), so a queue holding only
  marker-bearing cards ends on the no-card exit 3 (AC-TAU-006).
- Refusal: stderr one line `factory next: refused <token>: <detail>`, nothing on stdout, exit 4.
- Exit: M1's nominate/bare/arm-c/flag-set/state tests GREEN (MCP parity waits for M4); the five
  existing pinned lease tests of R3 still green; the golden unchanged.

### M3 — Refuse `moai todo --auto` in a lane, route the lane to the lease (Priority High; isolated — REQ-TAU-008)

- `internal/cli/todo.go` `todoRefuseLaneMutation`: refuse when `todoAutoFlag` is set and the
  session is a lane session — `factoryLaneAdmission()` or a non-empty `EnvMoaiFactoryWorker`; the
  Codex-backend clause of `factoryLaneRefusal()` is **not** used here (a Codex-backend leader keeps
  its batch approval — R11), and `factoryLaneRefusal()` itself is **not** changed. The refusal text
  is a dedicated function, not `todoLaneMutationRefusalText`: it names `moai factory next --card
  <id>` and says the `--auto` authorization is exercised through it.
- `internal/cli/factory_messaging.go` (~L234 comment, L255 string): `moai factory fallback declare`
  prints the lease path `moai factory next [--card <id>]` instead of `/moai:todo --auto
  self-service pickup (REQ-FLA-001)`; `TestFactoryFallbackDeclarePrintsLeasePath` pins the new
  string (no test pins any string there today).
- Exit: the M1 lane tests GREEN; the existing lane-refusal tests green. Dropping M3 removes only
  REQ-TAU-008 and AC-TAU-005.

### M4 — The MCP parameter (Priority Medium)

- `internal/cli/mcp_factory_card.go`: `factory_next` gains an optional `card` string and calls the
  same nomination implementation ("same implementation, same record changes, same refusals").
  `internal/mcp/catalog.go` carries only name and write-capability — **no change expected**
  (`catalog.go:112`); the run phase re-confirms with the catalog equality guard.
- Exit: `TestFactoryNextNominateMCPParity` GREEN (the MCP inputs are `run`, `project_root`, `card`;
  the CLI's `--wait`/`--wait-bound` have no MCP counterpart, so the MCP form takes **two** inputs
  today, not three).

### M5 — Doctrine amendment, live and mirror in one change, pins first (Priority High)

Everything below is **one commit**: no state exists in which a live file is edited and its mirror is
not, or in which a doc is edited and its pin is not.

Order inside the milestone: (1) write `TestAutoPickDocDoctrine` and `TestAutoPickMirrorParity` in
`internal/cli/todo_auto_pick_doc_test.go` — observe RED (AC-TAU-007/-008 RED-now cells); (2) edit
every doc and every mirror; (3) move the existing pins in `internal/cli/todo_auto_doc_test.go`;
(4) before moving the pins, run the scoped command once to **observe the two expected breakages**
and record them.

**Every existing marker that must move (D6, enumerated):**

| Test (file) | Marker | Disposition |
|---|---|---|
| `TestAutoRankDoctrineAmendment` (`todo_auto_doc_test.go`) | `The leader never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item.` for live and mirror `kanban-dispatch.md` | **moves** to `Outside an --auto authorization the leader never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item.` |
| `TestAutoRankMirrorParity` (same file, ~L188) | passage start marker `[HARD] **Promotion is the operator's act, always.**` for `kanban-dispatch.md` | **moves** to the new heading start of that paragraph; the end marker `[HARD] **The self-dispatch lane exception.**` is kept |
| `TestAutoRankMirrorParity` | `gtd.md` start `### --auto — the serial batch consumption`, end `## Standing sources` | **kept** — heading unchanged |
| `TestAutoRankMirrorParity` | template passage neutrality (no SPEC id, REQ token, ISO date, 9+ hex run) over both passages | **kept** — the new gtd/kanban text must pass it |
| `TestAutoRankDoctrineAmendment` | `An empty queue is a state to report, not a prompt to invent work.`; `never from queue emptiness, card readiness, or a peer's request`; `queue ADMISSION (production) stays the operator's.`; the five `autoDocGTDProhibitions` | **kept** verbatim |
| `TestAutoRankMarkerDisclosure` | four clauses (`[보류`, `only a card whose text begins with the [보류 marker is demoted`, `a hold stated in prose without the marker is not`, `the structural hold is moai todo hold`) on live gtd.md, mirror gtd.md, `--auto` flag help | **kept** — the D12 sentence is added beside them, not instead |
| `TestAutoRankAgentDoctrine` | both pinned literals in the `Serial-cycle contract (` and `## Jev Decision Boundary` regions; four stale phrases absent; kept Jev prohibitions | **kept** — new manager-todo wording must not reintroduce a stale phrase |
| `TestTodoSkillDocumentsClassification` (`todo_classify_doc_parity_test.go`) | whole-file neutrality of mirror `gtd.md` | **kept** — no internal token in the new gtd text |
| `internal/template/jev_auto_exception_test.go` | presence anchors `jaeAnchors` in `kanban-dispatch.md`, `gtd.md`, `manager-todo.md` | **kept** (read: markers are Jev tokens the edit leaves) |
| `internal/template/contract_mode_guided_test.go` | Kickoff-bearing documents classified (`kanban-dispatch.md` is `R`) | **kept** — no Kickoff text is touched |
| `internal/cli/init_headroom_export_test.go` | a path list for a measurement harness (skips without its flag) | **kept** — no assertion on content |
| `internal/template/rule_template_mirror_test.go` | read in this repair: no `kanban-dispatch`/`gtd`/`manager-todo` marker | **kept** |

Baseline of the scoped doc-pin command (observed, lane env scrubbed, one compound invocation,
`b3646de10`): `go test ./internal/cli -run '^(TestAutoRankDoctrineAmendment|TestAutoRankMirrorParity|TestAutoRankMarkerDisclosure|TestAutoRankAgentDoctrine|TestAutoHelpAndRefusalDoNotAssertPickOrder|TestTodoSkillDocumentsClassification)$' -count=1 -v` → exit 0, 6 of 6 named tests `--- PASS`, `ok … 2.035s`. (A wider prefix selector over `TestAutoRank`, `TestTodoAuto`, `TestAutoHelp` ran 54 passing tests in 108.063s, exit 0; the anchored list is the pin set that reads the amended files.)

Edits, each a **replacement** of a named sentence, **applied to the live file and its mirror together**
(surgical — lane-5's card t1453 absorbs these, so no reflow, no heading moves, no list
restructuring; AC-TAU-013 bounds each file's diff):

| File | Replace | With (the § C.1 literals; other wording free) |
|---|---|---|
| `kanban-dispatch.md` L29 | `**Promotion is the operator's act, always.**` and the unscoped `The leader never picks for the operator, …` | promotion is the operator's act, in person or in advance through `--auto`; `Outside an --auto authorization the leader never picks for the operator, …` |
| `kanban-dispatch.md` L31 | `The one reconciliation is named, not excepted: … authorized serial consumption of the queue and nothing else.` | authorizes the invoked session to take cards from the queue on its own judgment, each only through a lease and never a keep-set card, and nothing else |
| `kanban-dispatch.md` L33 | `only through \`moai factory next\`, whose lease lands in the factory record the way a leader's dispatch lands in the queue.` | `only through \`moai factory next\` — bare, or \`--card <id>\` for the lane's own judged pick.` |
| `gtd.md` `--auto` section (L334-L361) | the `[HARD] … authorizes serial consumption of the queue and nothing else` clause | the same clause with `on its own judgment`; the lane routing sentence (`a lane session exercises the --auto authorization through moai factory next …`); the **keep-set list** (mechanical: `hold` state, `[보류` marker, `blocked`, serial slot, owned; text-judgement: payments, secrets, or irreversible external-shared work); the **record form** (`decision record:` with `ladder_path=gate-row card pick (AUTONOMOUS, auto-semantics §9)` and `unmeasured`, located in the card's progress record as evidence); the D12 sentence naming both treatments of the marker; the "exactly one card is in flight" sentence scoped to the operator-session cycle |
| `auto-semantics.md` §9 card-pick row | the row text | keeps the AUTONOMOUS disposition, adds the pointer `§9.3` |
| `auto-semantics.md` new `### 9.3` | — | the keep-set, the input set, the record form with `ladder_path=gate-row card pick (AUTONOMOUS, auto-semantics §9)`, `evidence, not the decision board`, `unmeasured`, the open-set sentence `adds an input without amending the keep-set or the lease path` |
| `manager-todo.md` L23, L34-L35 | `process cards in queue order`; `serial consumption of the queue and nothing else` | `on its own judgment` / keep-set wording; the serial-cycle region keeps the pinned ranking literals |
| `moai-kanban-foreman/SKILL.md` Boundary 1, step 4 | `serial consumption in queue order is authorized`; `\`queued\` items are not yours to pick.` | `… on the iteration's own judgment, within the keep-set`; `… not yours to pick outside a batch authorization.` |
| `kanban-dispatch-detail.md` § The pre-dispatch cross-check | — | one added paragraph carrying `a pull request or landed state is a skip input for a queued candidate the session chose and is report-only for an operator-picked card` and the B.2 reconciliation. Lazy file — growth is free |
| `moai-mcp-tools-catalogue.md` factory_next row | the one cell | mentions the optional `card` argument |

New detail goes in `auto-semantics.md` §9.3 (lazy, `paths:`-scoped to the watchdog skill), `gtd.md`
(a lane reads it) and `kanban-dispatch-detail.md` — **never** into the always-loaded stub beyond the
replaced sentences. Preserve each copy's `kanban-dispatch.md` line 177 exactly (see §4).

- Exit: the new doc pins GREEN **for live and mirror at once**; the moved existing pins GREEN; every
  AC-TAU-007/-008/-010/-011/-013 reading holds on the committed tree; the mutants of `acceptance.md`
  each fail their criterion.

### M6 — Verification and measurement only (Priority High; no file edits)

Run only what the change can affect, then let CI run the rest (`AGENTS.md` §4); record every command,
exit code and verbatim tail in `progress.md` §E.2:

- `internal/cli`: the new tests plus `^(TestFactoryNext|TestAutoRank|TestAutoPick|TestAutoHelp|TestTodoAuto|TestTodoLane|TestTodoNonLane|TestTodoSkill|TestFactoryFallback)`; take the `moai slot` lease for the
  package run (`.claude/rules/local/gitflow-lane-protocol.md` §8) and scrub the lane environment in
  one compound invocation.
- `internal/template`: the tests naming the edited files (`jev_auto_exception_test.go`,
  `contract_mode_guided_test.go`, `rule_template_mirror_test.go`) by `-run` selector, with a
  swept-count check.
- Measurements (AC-TAU-010, -011, -012, -013): `cmp` on each pair, `wc -c` and `wc -m` on both
  `kanban-dispatch.md` copies against the baselines, `git diff --numstat "$CARD_BASE"..HEAD` against
  the caps, the boundary pathspec probe, and the draft-versus-actual byte comparison (§3).
- `GOOS=windows GOARCH=amd64 go build ./...`; `golangci-lint` at the CI version.
- The commit body of M5 carries the measured before/after bytes **and characters** and the
  non-invoking-cost sentence (`rule-authoring.md` (c)).

### Milestone ↔ criterion order check (D1)

Each criterion's green path names a milestone, and every obligation that criterion reads is
completed in that milestone or an earlier one — checked by hand for **all** fourteen:

| AC | Green at | Reads work from | Order holds because |
|---|---|---|---|
| AC-TAU-001 | M2 (CLI rows), M4 (MCP row) | M2, M4 | the MCP row is stated against M4 only; M2 does not claim it |
| AC-TAU-002 | M2 | M2 | uses the nominated path only |
| AC-TAU-003 | M5 | M5 | doctrine literals, live and mirror, one commit |
| AC-TAU-004 | M2 | M2 | the arm (c) skip and the nominated refusals are both M2 |
| AC-TAU-005 | M3 (Go), M5 (routing sentence) | M3, M5 | the Go rows do not read the docs; the routing-sentence row is stated against M5 |
| AC-TAU-006 | golden GREEN at M1 (unmodified), stays GREEN at M2 | M1, M2 | the golden is GREEN before any change, so no RED/GREEN conflict (D2) |
| AC-TAU-007 | M5 | M5 | live, mirror, pins in one commit (D1) |
| AC-TAU-008 | M5 | M5 | same commit |
| AC-TAU-009 | run evidence (M6 report) | the first lane lease under the doctrine | no executing re-reader is assumed |
| AC-TAU-010 | M5, measured M6 | M5 | both sides edited together |
| AC-TAU-011 | M5, measured M6 | M5 | measurement only at M6 |
| AC-TAU-012 | M6 | all | boundary read over the finished diff |
| AC-TAU-013 | M5, measured M6 | M5 | floor and ceiling both read the M5 commit |
| AC-TAU-014 | M2 | M2 | validation, promotion, compensation in one milestone |

## 3. Always-loaded growth: the constraint, the method, the draft

`kanban-dispatch.md` is always-loaded, the repo is already over its instruction budget (t1318:
214,155 chars against 210k), and `rule-authoring.md` requires a statement for growth over 1,000 B.
This SPEC requires **net growth ≤ 0 on both copies, in bytes and in characters**. Baselines
(observed, tree `b3646de10`, the file is unchanged since `4bf547bca`): live 26,959 B / 26,754
chars; mirror 26,637 B / 26,433 chars.

Measuring commands, both copies: `wc -c` and `wc -m` of `.claude/rules/moai/workflow/kanban-dispatch.md` and
`internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md`.

**Feasibility draft — reproducible.** The scratch script below replaces the three paragraphs with
wording that carries the § C.1 literals and prints the per-line and total delta in bytes and
characters plus the literal checks. It was run at `b3646de10` against the live copy (the three
replaced lines are identical in the mirror, so the deltas apply to both):

```text
python3 draft2.py .claude/rules/moai/workflow/kanban-dispatch.md
  line 29: +72 B / +70 chars     line 31: -54 B / -54 chars     line 33: -30 B / -32 chars
  total delta: -12 B, -16 chars; every retained pin literal present; the three old literals absent
```

The script (about 40 lines) replaces, by exact string, the heading phrase and the
`The leader never picks for the operator,` clause on L29; the `The one reconciliation is named, not
excepted: … authorized serial consumption of the queue and nothing else.` sentence, the
`, so this clause and the cycle state the same rule from two sides` clause, the `Serial-contract
detail:` label and ` in the gate inventory` on L31; and the `, whose lease lands in the factory
record …` clause on L33 — and then measures `len(s.encode())` and `len(s)`. It lives in the
scratch directory of the plan session and is **not committed** (a script outside the tree cannot be
re-run by a stranger); the run phase re-derives the measurement from the committed edit with the
`wc` commands above, which are the binding measurement. The draft is evidence that ≤ 0 is
reachable, not the binding figure.

## 4. Mirror handling (preserve, do not absorb)

The live and mirror `kanban-dispatch.md` differ by one pre-existing line (live line 177, the `moai
worktree sweep …` sentence, present in the live copy only; research R7). Preserve each copy's line
177 exactly; do not sync it either way. Reason: absorbing it is another card's act, it would put an
unreviewed sentence into every user project's always-loaded rule, and the existing parity test is
deliberately scoped to the amended passage. After the edits `git diff --no-index --numstat` of the
pair must still read `1  1` (AC-TAU-010). The completion report names the drift for a follow-up card.
If the Makefile defines a template embed refresh, run it after the edits (pre-flight; research R10).

## 5. Files to modify

| Area | File | Change |
|---|---|---|
| Go | `internal/cli/factory_card.go` | `--card` flag, nomination path (validate / promote / claim / compensate), `factoryKeepSetRefusal`, arm (c) `[보류` skip counted as seen, refusal exit/diagnostic, promote-then-claim test seam |
| Go | `internal/cli/mcp_factory_card.go` | optional `card` parameter, shared implementation |
| Go | `internal/cli/todo.go` | lane refusal of `--auto` in `todoRefuseLaneMutation` (lane-session predicate, dedicated text) |
| Go | `internal/cli/factory_messaging.go` | the printed fallback instruction (and its comment) point at the lease path |
| Go test (new) | `internal/cli/factory_nominate_test.go` | nomination, keep-set, state semantics, concurrency, bare golden, arm (c), flag set, MCP parity, lane `--auto`, GPT guard, fallback string |
| Go test (new) | `internal/cli/todo_auto_pick_doc_test.go` | doc contract literals, stale-absent, mirror parity, diff bound, mutants |
| Go test (edit) | `internal/cli/todo_auto_doc_test.go` | move the two markers of the § M5 table |
| Rule (live + mirror) | `kanban-dispatch.md` | replace three paragraphs (net ≤ 0 B and chars; line 177 preserved) |
| Rule (live + mirror) | `kanban-dispatch-detail.md` | one paragraph (D10 sentence) |
| Rule (live + mirror) | `auto-semantics.md` | §9 row pointer + new §9.3 |
| Skill (live + mirror) | `.claude/skills/moai/workflows/gtd.md` | `--auto` section: replacement, keep-set list, record form, lane routing, D12 sentence |
| Agent (live + mirror) | `.claude/agents/moai/manager-todo.md` | two sentence replacements |
| Skill (live + mirror) | `.claude/skills/moai-kanban-foreman/SKILL.md` | Boundary 1 and step 4 |
| Rule (live + mirror) | `.claude/rules/moai/core/moai-mcp-tools-catalogue.md` | one catalogue cell |

No file under `internal/kanban/**`, `internal/graph/**`, or any schema file changes (REQ-TAU-003);
no `sync-auditor` / `sync-audit-4dim` change (REQ-TAU-012 states the record is unread, not
controlled).

## 6. Forward-compatibility with t1454

What this SPEC **leaves open**: the input set is open — the record's `evidence_refs` and the
selection rule name the inputs t1448 needs and say a later card adds an input without amending the
keep-set or the lease path (REQ-TAU-013). Relation records are read through whatever the lane-read
surfaces expose (today `moai todo why`); no code here names the queue-findings store or the
`gtd_relations` table as *the* source. File overlap is a pluggable input: expected-file data when
present; otherwise a fallback that is INFERRED and unmeasured (a lane reading another lane's branch
by name from its own tree); `unmeasured` is the expected value until measured, and the record does
not require it.

What this SPEC **must not pre-empt**: it adds no schema field, no relation kind, no `moai graph`
change, no expected-file, parent/spawned-by or size field, and no mechanical refusal keyed on a
relation store (the relation-blocked refusal is deliberately *not* in the keep-set predicate;
probe O2). The mechanical guard reads only `state`, the `[보류` marker, `classification.blocked`,
the serial slot, and ownership — none of which t1454's inputs change.

## 7. Operational follow-ups (not part of this SPEC's code)

- **Before lanes exercise this doctrine** the operator or leader must `moai gtd hold` (or
  `[보류`-mark) the cards kept on an operator decision queue — t810, t1294, t1383 per the memory
  index and read as selectable by the plan audit. Lanes cannot mutate the queue, so this is the
  operator's or leader's act; the completion report must carry the note (Definition of Done 6).
- The line-177 drift in `kanban-dispatch.md` needs a follow-up card to decide which copy is right.
- The lane bootstrap notice still says bare `moai factory next`; a follow-up candidate is to mention
  the judged `--card` form in the four locales of `session_start_factory_i18n.go`.
- No party re-reads the decision record (spec §G); a follow-up card may add that step to the sync
  audit if the record is wanted as a control.

## 8. Risks and their milestones

| Risk | Where handled |
|---|---|
| Doc pins red when the docs change | M5: docs, mirrors and pins in one commit; the two expected breakages observed first |
| Default-path `[보류` skip surprises an operator | isolated clause of REQ-TAU-007; stated in B.4 |
| Lane `--auto` refusal reverses an FLA doctrine move, or refuses a Codex leader | isolated REQ-TAU-008 with an explicit predicate, a non-lane GPT guard, and the corrected fallback string |
| Promoted-then-lost nominee leaves a `picked` card | M2 steps 1-4, two interleave tests |
| Self-attested, unread decision record | spec §G (unimplemented residual risk), DoD item 6 |
| Parallel edit collision with lane-5's t1453 | surgical replacements; AC-TAU-013's diff bound; this card's rule-doc edits land first |
| Always-loaded growth | AC-TAU-011, bytes and characters, both copies |

## 9. Commit structure

Plan phase: the plan commit and this repair commit carry the artifact set. Run phase: tests-first
commit(s) per milestone; M5 is one commit for docs, mirrors and pins. The baseline measurements of
`acceptance.md` land in the plan commits, **before** any run commit, so the commit graph — not a
message — witnesses that the baselines precede the change (`verification-claim-integrity.md` §2.3).

## 10. Audit-1 resolution map

| Item | Resolution (where / how) |
|---|---|
| D1 | M5 now edits live, mirror and pins in one commit; M6 is measurement only; the milestone↔criterion order table above checks all fourteen criteria; AC-TAU-007/-008/-010 green paths restated |
| D2 | M1 exit split: all new tests RED except the golden and the non-lane GPT guard (GREEN on the unmodified tree) with seeded-perturbation non-vacuity cells; golden minimum case set in M1 and AC-TAU-006 |
| D3 | `factory_messaging.go` added to § 5 and M3; printed string moves to the lease path; `TestFactoryFallbackDeclarePrintsLeasePath` pins it; REQ-TAU-008 and spec B.4 carry the supersession sentence for REQ-FLA-001 |
| D4 | REQ-TAU-008 defines the predicate (role marker `lane` or non-empty label; Codex marker alone not sufficing); dedicated refusal text; gtd.md routing sentence pinned (C.1); AC-TAU-005 includes a non-lane GPT-backend session measured with a real command (R11 M1) |
| D5 | Option (b): spec B.5 and §G retract the compensating-control claim (unimplemented residual risk); REQ-TAU-012 states the location and "evidence, not the §11 board"; AC-TAU-009 rewritten with no executing party; no audit-surface edit (Out of Scope) |
| D6 | AC-TAU-013 bounds each edited file's diff (floor and ceiling by `git diff --numstat`); § M5 table enumerates every existing marker including `TestAutoRankMirrorParity`'s start marker; the jev/contract/rule-mirror/headroom tests were read |
| D7 | REQ-TAU-005 restated as the state left; spec § C.2 token table normative; M2 mechanism (validate read-only, re-validate inside `Mutate`, claim, compensate); AC-TAU-014 covers `dropped`, `owned`, `quota-hold`, `backend-skip`, promote-then-lose, claim-refused rollback |
| D8 | AC-TAU-001 (flag-set equality, MCP parity), AC-TAU-014 (quota hold and Codex skip honored by the nominee), AC-TAU-004 (`dropped`), AC-TAU-006 (all-marker queue exits 3, golden minimum case set) |
| D9 | AC-TAU-010 and AC-TAU-012 reclassified as regression-guards (neither is release-blocking); AC-TAU-011 stays conjunctive; AC-TAU-013 carries a real RED-now floor |
| D10 | spec B.2, REQ-TAU-010, § C.1 literal for `kanban-dispatch-detail.md`, and the M5 paragraph: PR/landed state is a skip input only for `queued` candidates the session chose, report-only for operator-`picked` cards |
| D11 | Definition of Done item 6 and § 7: the operator or leader holds or marks t810, t1294, t1383 before lanes exercise the doctrine; the completion report carries the note |
| D12 | REQ-TAU-011 and a § C.1 literal: the marker demotes in the serial cycle and excludes on the lease path, one sentence in the gtd.md `--auto` section |
| D13 | research R1 corrected: the MCP form takes `run` and `project_root` (two inputs), M4 states it |
| D14 | M1: `moai slot` uses exit 4 (`slot.go:38`) in a different verb family; the run phase re-greps |
| D15 | REQ count held at 16: REQ-TAU-003 tightened, REQ-TAU-004 is `When`, doctrine-only parts marked in the requirements and named in the criteria |
| D16 | REQ-TAU-013, spec B.7 and § 6: the file-overlap fallback is INFERRED, `unmeasured` is the expected value, the record does not require it (research R5/R11 mark it) |
| D17 | REQ-TAU-014 and § C.1: the gtd.md `--auto` section carries the keep-set list, the record form and the lane routing sentence themselves |
| D18 | AC-TAU-011 measures bytes (`wc -c`) and characters (`wc -m`) with baselines; § 3 states the draft's reproducible method and that the committed `wc` measurement is binding |
