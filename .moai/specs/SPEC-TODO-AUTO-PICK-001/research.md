# SPEC-TODO-AUTO-PICK-001 — Research

Plan-phase research for card t1448. Pinned tree `4bf547bca` (branch `WT-todo-auto-pick-autonomy`).
Every claim below is marked **OBSERVED** (a command was run in this tree and its output read),
**READ** (read from a file at a cited line; no execution), or **INFERRED** (reasoned, not
measured). Nothing here is a verification claim about code that does not exist yet.

## R1. How a card is selected today

**READ** `internal/cli/factory_card.go:402-542`, `factoryNextSelectAndLease`. Four arms, tried in
order: (a) a card already assigned to this lane; (b) an operator-picked card with a record row at
`picked` and no owner; (b2) a queue-picked card with no record row yet; (c) the first `queued`
card in stored order that is not classification-`blocked`, and — when it is a serial card — only
while no other serial card is in flight. Arm (c) promotes the card to `picked` inside
`todoStoreAt(root).Mutate(...)`, then records and claims it through the version-checked F1 edges.
Stored order is priority order; selection never re-sorts (`factory_card.go:497-502`).

**READ** `newFactoryNextCommand` (`factory_card.go:629-698`): the flags are `--wait`,
`--wait-bound`, `--run`. There is no way to name a card. The CLI chooses; the session cannot.
`factoryNextNoCardExit = 3` (`factory_card.go:181`). The MCP form (`handleFactoryNext`,
`internal/cli/mcp_factory_card.go:111`) shares the implementation but takes **two** inputs, `run`
and `project_root` (`mcp_factory_card.go:42-48`; the description says "no --wait"). An earlier
draft said "the same three inputs"; the plan audit (D13) showed that to be wrong — the CLI has
three flags, the MCP form two inputs.

**READ** the lane queue boundary: `todoLaneReadOnlyVerbs` (`internal/cli/todo.go:~346`) allows a
lane only `list`, `history`, `show`, `why`, `pr`, `triage`. `todoRefuseLaneMutation`
(`todo.go:387`) lets the bare parent command through when `len(args) == 0`.

## R2. Observations from a throwaway probe (deleted, never committed)

A temporary test file `internal/cli/zz_t1448_probe_test.go` was written, run once, and removed
before any commit (`git status --short` empty afterwards, observed). It used the existing fixtures
`fcFixture` / `fcQueue` / `fcClassify` / `sdRegisterLane` and called `factoryNextLeaseOnce` and
`runTodo` directly. **The probe no longer exists, so none of these four rows is a re-executable
RED-now cell** (`verification-completeness.md` §2.1 undecidable disposition); they are history that
motivates the criteria, and the criteria's own RED-now cells are the re-executable ones in
`acceptance.md`.

| Id | Setup | Verbatim output | What it shows |
|---|---|---|---|
| O1 | two queued parallelizable cards; `t1` text opens with `[보류 operator decision pending]` | `PROBE hold-marker: leased=true card=t1 err=<nil>` | arm (c) ignores the `[보류` text marker and leases the marked card |
| O2 | two queued parallelizable cards; live `depends` finding `{subject t1, related t2}` (t1 waits on t2) | `PROBE relation-blocked: leased=true card=t1 err=<nil> (t1 waits on t2)` | arm (c) ignores sequencing findings; only the `--auto` cycle filters them (`todo_auto.go:167`) |
| O3 | `t1` in state `hold`, `t2` queued | `PROBE hold-state: leased=true card=t2 err=<nil>` | the `hold` state is excluded structurally (positive state enumeration) |
| O4 | lane env set (`MOAI_FACTORY_ROLE=lane`, `MOAI_FACTORY_WORKER=lane-1`); `runTodo(t, "--auto", "--auto-wait", "1ms")` over two queued cards | `PROBE lane --auto: err=<nil> out="jev: unavailable … selection: ranked t1 t2 … accept t1 factory card 1 … unpick t1 non-finding: worker evidence absent at deadline … accept t2 …"` | a lane session is **not** refused `moai todo --auto`; the serial cycle ran, picked and unpicked queue cards from inside a lane, outside the factory record and outside the lease |

Two probe-environment facts worth keeping: the first run failed on every case with
`moai add: refused — lane boundary …` because this very session carries lane environment
variables (`MOAI_FACTORY_ROLE`, `MOAI_FACTORY_WORKER`, …); the fix was `sdClearLaneEnv(t)` at the
top of each probe. Any run-phase test that builds a fixture through `runTodo add` must scrub the
same variables (`feedback_lane_env_falsifies_env_reading_guard_tests_locally`).

## R3. Baseline of the tests that already pin concurrent lease behavior

**OBSERVED**, lane env scrubbed in one compound invocation, tree `4bf547bca`:

```text
unset MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_KANBAN_BACKEND MOAI_KANBAN_ID MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli -run '^(TestFactoryNextSerialMutualExclusivity|TestFactoryNextSkipsClassificationBlocked|TestFactoryNextParallelizableConcurrentLeases|TestFactoryNextRecordAndClaimRaceOnLeasedRow|TestFactoryNextDuplicateDispatchGuard|TestAutoRankDoctrineAmendment|TestAutoRankMirrorParity|TestAutoRankAgentDoctrine)$' -count=1 -v
exit=0   (8 of 8 named tests PASS; ok github.com/modu-ai/moai-adk/internal/cli 31.555s)
```

What is **already pinned** (so this SPEC does not re-specify it as new behavior):

| Behavior | Pinned by | Status |
|---|---|---|
| two lanes racing `next` over two distinct parallelizable cards hold different cards | `TestFactoryNextParallelizableConcurrentLeases` (`factory_classify_test.go:226`) | pinned |
| two lanes racing the same queued card end with exactly one holder | `TestFactoryNextDuplicateDispatchGuard` | pinned |
| a record-and-claim race on an already-leased row re-selects | `TestFactoryNextRecordAndClaimRaceOnLeasedRow` (`:280`) | pinned |
| a classification-`blocked` card is never auto-selected | `TestFactoryNextSkipsClassificationBlocked` (`:160`) | pinned |
| a second serial card is never leased while one is in flight | `TestFactoryNextSerialMutualExclusivity` (`:67`) | pinned |
| the `--auto` doctrine paragraphs, mirror parity, agent doctrine wording | `TestAutoRankDoctrineAmendment`, `TestAutoRankMirrorParity`, `TestAutoRankAgentDoctrine` | pinned — **and they pin the sentences this card must amend** (R6) |

What is **NEW** in this SPEC: choosing a *specific* card (nomination), the keep-set refusal on a
nominated card, the `[보류` skip on the default arm, refusing `moai todo --auto` inside a lane, the
decision-record obligation, and the doctrine sentences. The concurrency core (distinct leases,
exactly-one on a shared card) is existing machinery; the new tests re-assert it **through the
nominated path** so that a nomination implementation that bypasses the version-checked edges is
caught.

## R4. Unverified items from the dispatch — resolved by reading

**(i) Where `[보류` and hold/blocked are defined.** **READ**
`internal/cli/todo_auto_rank.go:48` — `const autoRankHoldMarker = "[보류"`, "Only a card whose
trimmed text OPENS with it is demoted". It is consumed by the `--auto` ranking stage only
(`grep` over `internal/` non-test Go finds the literal in exactly `todo_auto_rank.go:48` and the
help string `todo.go:315`). `factory_card.go` never references it (O1 confirms). The structural
hold is `BacklogStateHold` (`internal/kanban/backlog_store.go:73`), "no lease path gains a verb
that sets or clears it". `Blocked` is a field of `CardClassification`
(`internal/kanban/classification.go:87`), read through `EffectiveCardClassification`.

**(ii) Whether any "operator-decision-queue" marker exists.** **OBSERVED** by search
(`grep -rniE 'decision.queue|operator.decision|decision_queue|operator-decision'` over `internal`,
`.claude`, `.moai/docs`, `*.go *.md *.yaml`): every hit is an unrelated phrase ("operator decision
2" in a config comment, `Decider … operator-decision edges` in `homestate/card_transition.go:42`).
**No marker, state, field, or text convention named an operator decision queue.** The memory index
entries for cards t810/t1294/t1383 say they are "kept on an operator decision queue"; whatever
that means operationally, **it is not expressed in the queue record by anything the code reads**
(INFERRED: it is a verbal arrangement, possibly realized as plain `queued` cards the operator
does not want worked). Consequence: those three cards are not mechanically protected today. The
queue schema is frozen additive-only ("no per-item field may ever be added",
`backlog_store.go:~52`), so a new per-item marker is not an option; the available structural
expressions are the `hold` state and the `[보류` text marker (decision D2 in `spec.md` §B).

**(iii) What writes the §11 decision board.** **OBSERVED** by search (`grep` for `decision record`,
`decisions.jsonl`, `decision-board` over non-test Go): the only hits are the unrelated
`renderAutoSelectionRecord` (`todo_auto_rank.go:500`, the `selection:` lines). **No Go code
appends to a decision board.** `moai factory decide` is gate-only (`--gate kickoff|push`,
`--choice approve|reject|resume|block|unblock|abandon`, `--decider human`; `factory_card.go:1559-1562`).
The lane-watchdog skill instructs a lane to write the record line "on the disk board" with no
verb named. **READ** `SPEC-JEV-AUTO-EXCEPTION-001/progress.md` §F.1: the previous card recorded
that the board "was NOT written", citing that its predecessor t1400 "recorded `factory_decide` as
refused for a lane session and found no lane-writable board verb", and noted the finding was not
re-measured. That second-hand statement is **not re-measured here either**. So the card-pick decision
record is, in this tree, a policy-layer obligation with a self-attested line (`auto-semantics.md`
§10: "detection, not prevention"); it lands in the card's progress record. **Retraction (plan
audit D5):** an earlier draft called "the sync audit's re-read" a compensating control. The audit
measured `grep -n -i decision .claude/agents/moai/sync-auditor.md` (only a model-escalation note)
and found no decision-record step in `.claude/workflows/sync-audit-4dim.js`, so **no party executes
that re-read today**. SPEC consequence: the requirement binds the line's *form, content and
location* (evidence in the card's progress record, explicitly not the §11 board, which calls a
tree-local copy "a ghost — never a board"), and the missing re-read is an **unimplemented
residual risk**, not a control. This SPEC does not add the re-read (it would edit audit surfaces the
t1453/t1454 cards may touch).

## R5. The three lane pick paths today (the double-pick surface)

**READ + OBSERVED (O4)**: (1) `moai factory next` — leases through the factory record; (2) `moai
todo --auto` — the serial CLI cycle, which a lane session may run (O4) and which picks in the
*queue* (`picked` state) with liveness-measured ownership; (3) `moai gtd claim` — refused to a lane
(`todo_claim.go`). Arm (b2) of `factory next` claims *any* queue-`picked` card with no record row,
by design (it exists for operator picks). **INFERRED, not measured:** a card picked by a lane's
`--auto` cycle and not yet recorded is therefore claimable by another lane's arm (b2), because
nothing in arm (b2) distinguishes an operator pick from a cycle pick. The inference is why the
card's item (2) — "the lease is the ONLY pick path" — needs a mechanical guard on the cycle path
and not only a sentence. SPEC-FACTORY-LANE-AUTONOMY-001 REQ-FLA-001 says a lane in messaging
fallback "switch[es] to `/moai:todo --auto` self-service pickup"; its REQ-FLA-006 describes that
pickup in lease-model terms (a sequential card "only while no other lane holds … a card of the same
sequential group"), which is arm (c)'s serial-slot rule. **Plan audit D3 corrected an earlier
"reading" reconciliation:** REQ-FLA-001's text is literal, and a shipped verb prints it —
`moai factory fallback declare` ends with `switch to /moai:todo --auto self-service pickup
(REQ-FLA-001)` (`factory_messaging.go:255`, comment L234; **OBSERVED** by `git grep`, cell L16);
AC-FLA-001 of the completed SPEC says the same. No test pins that string. So this SPEC does not
"read" the phrase differently: it **supersedes it in lane sessions** by the lease path, changes the
printed string and adds a test, and leaves that completed SPEC's body unedited.
**INFERRED (the file-overlap fallback, D16):** that a lane can read another lane's branch by name
from its own worktree is plausible but unmeasured — the worktree guard refuses `git -C` into other
trees (`kanban-dispatch.md` § Isolation) and the branch slug is not derivable from a card id —
so `unmeasured` is the expected value of that input until measured.

## R6. Existing doc pins that the amendment must move in lockstep

**READ** `internal/cli/todo_auto_doc_test.go`:

- `TestAutoRankDoctrineAmendment` requires, in live and mirror `kanban-dispatch.md`, the whole
  sentence `The leader never picks for the operator, never reorders by inferred priority, and never
  silently promotes a backlog item.` — the sentence card item (5) says to amend. The amendment
  therefore **cannot be a pure doc edit**: the pin test changes in the same milestone, to the new
  scoped sentence, or the repo goes red.
- the same test keeps `An empty queue is a state to report, not a prompt to invent work.`,
  `never from queue emptiness, card readiness, or a peer's request`, `queue ADMISSION (production)
  stays the operator's.` — the amendment keeps all three verbatim.
- `autoDocGTDProhibitions` keeps `The pick is the operator's. Do not preselect, …` (the `gtd next`
  leader path, not `--auto`) — **kept unchanged**.
- **`TestAutoRankMirrorParity` uses `[HARD] **Promotion is the operator's act, always.**` as the
  passage start marker for `kanban-dispatch.md`** (`todo_auto_doc_test.go` ~L188) and
  `[HARD] **The self-dispatch lane exception.**` as its end marker; for `gtd.md` the start marker is
  the heading `### --auto — the serial batch consumption` and the end marker `## Standing sources`.
  The kanban start marker is a literal the amendment must remove (plan audit D6): the marker moves
  with the heading. The same test applies the template-neutrality expression (no SPEC id, REQ
  token, ISO date or 9+ hex run) to both passages, and `TestTodoSkillDocumentsClassification`
  applies it to the whole mirror `gtd.md`.
- `TestAutoRankMarkerDisclosure` pins four clauses about the `[보류` demotion on live gtd.md,
  mirror gtd.md and the `--auto` flag help — kept; the two-treatments sentence (D12) is added
  beside them.
- `TestAutoRankAgentDoctrine` requires both pinned literals in the manager-todo serial-cycle
  region and keeps `strict queue order` etc. absent — compatible with the amendment.
- Read in the repair (plan audit gap G8): `internal/template/jev_auto_exception_test.go` uses
  `kanban-dispatch.md`, `gtd.md`, `manager-todo.md` only as presence anchors for the Jev token
  (`jaeAnchors`); `contract_mode_guided_test.go` classifies `kanban-dispatch.md` as a
  Kickoff-bearing document (`R`) — the edit touches no Kickoff text;
  `init_headroom_export_test.go` lists `kanban-dispatch.md` in a measurement-harness path list (it
  skips without its flag); `rule_template_mirror_test.go` names none of the amended files.
- `autoDocMirrorPassage` already tolerates the one pre-existing drift ("the two kanban-dispatch.md
  copies differ elsewhere, by a pre-existing sentence that is not part of this amendment").

Other tests naming `kanban-dispatch.md`: `internal/template/contract_mode_guided_test.go`,
`internal/template/jev_auto_exception_test.go`, `internal/cli/init_headroom_export_test.go` — the
run phase re-runs them (plan.md §Verification scope). Their Jev prohibitions are untouched.

## R7. Template-mirror state

**OBSERVED** (`cmp` / `git diff --no-index --numstat`, tree `4bf547bca`): `kanban-dispatch-detail.md`,
`-mechanics.md`, `auto-semantics.md`, `gtd.md`, `manager-todo.md`, `moai-kanban-foreman/SKILL.md`
are byte-identical to their `internal/template/templates/.claude/...` mirrors. `kanban-dispatch.md`
differs by exactly one line: `cmp` → `differ: char 23771, line 177`; numstat `1  1`. The live file
carries an extra trailing sentence on that line (`moai worktree sweep …`), added by commit
`62197bb38 feat(SPEC-WORKTREE-SWEEP-001): M5 doctrine …` to the live copy only. **Decision (plan.md
§Mirror handling): preserve the drift, do not absorb it** — absorbing is another card's act and would
put an unreviewed sentence into every user project's always-loaded rule; the existing parity test
is deliberately scoped to the amended passage. The run phase flags the drift in its completion
report for a follow-up card.

Byte sizes (**OBSERVED**, `wc -c`): live `kanban-dispatch.md` 26,959 B; mirror 26,637 B (diff 322 B =
the line-177 sentence). `kanban-dispatch.md` is the only target that is always-loaded
(`kanban-dispatch.md:5` "Intentionally always-loaded"); `-detail.md` and `-mechanics.md` carry
`paths:` (lazy), `auto-semantics.md` carries `paths: ".claude/skills/moai-lane-watchdog/**"` (lazy);
`gtd.md`, `manager-todo.md`, the foreman skill are skill/agent files. The mirror copy ships to user
projects as an always-loaded rule, so the non-growth bound applies to **both** copies.

**Feasibility measurement** (scratch script, not committed — see plan §3 for the reproducible
method; the binding measurement is the committed `wc -c` / `wc -m`): a first draft replacing the
sentences of `kanban-dispatch.md` lines 29, 31, 33 gave +72, −31, −30 = +11 B. After trimming two
phrases (`so this clause and the cycle state the same rule from two sides`, ` in the gate
inventory`) the draft measures **−12 B and −16 characters** (+72/−54/−30 B; +70/−54/−32 chars),
every retained pin literal present and the three old literals absent. Baselines (observed):
live 26,959 B / 26,754 chars, mirror 26,637 B / 26,433 chars. The run phase owns the final wording.
The `62197bb38` provenance of the line-177 sentence is from a `git log -S` run earlier in the plan
session; the plan audit's own re-run was refused by the worktree guard (gap G2), so it is **not
independently re-verified** — the drift itself is observed (`cmp`, numstat).

## R8. Design alternatives measured against the evidence

**Q1 mechanism.**

| Alternative | New surface | Keeps `factory next` the only path | Keep-set mechanical | Cost / risk |
|---|---|---|---|---|
| **A. `factory next --card <id>` nomination** (chosen) | one flag + one MCP param | yes — the lease still goes through arms' version-checked edges | yes, one shared predicate, refusal tokens | smallest; no flag → unchanged |
| B. LLM only reorders a CLI-offered candidate list | a verb that *prints* candidates, plus a way to hand the order back (an env var or file) | yes | partly — the CLI would still choose the final lease | two-step protocol; the hand-back channel is a new state store; a stale list is a race |
| C. read-only `peek` verb + nomination | a new read verb **and** the flag | yes | yes | the read side already exists: `moai todo list --json`, `why`, `pr`, `show` are lane-readable (`todoLaneReadOnlyVerbs`); a `peek` would duplicate them |

A is chosen because the read half of "judge, then pick" already exists as lane-read-only verbs, so
only the write half (a nominated, validated, atomic lease) is missing. B moves state to a
hand-back channel; C adds a verb that duplicates existing reads.

**Q5 foreman.** The foreman skill removes `AskUserQuestion` mechanically
(`disallowed-tools: AskUserQuestion`, `SKILL.md:18`), serializes to one worker, and never runs
`moai gtd next <n>`. Its Boundary 1 states "serial consumption in queue order is authorized" under a
batch authorization, and step 4 says "`queued` items are not yours to pick". Both sentences become
untrue statements of the new rule only in the `--auto` case; outside an authorization they remain
right. Decision: amend both minimally and scope them to "batch authorization"; keep the tool
removal and the one-worker serialization.

## R9. Risks (ordered by likelihood × consequence)

1. **Doc pin tests red in the same commit that amends the docs.** Mitigation: M5 moves docs and the
   pins together; the RED-first order in plan.md puts the new pin test before the doc edit.
2. **The `[보류` skip on arm (c) is a default-path behavior change** (R2/O1: it leases a marked card
   today). Mitigation: isolated requirement (REQ-TAU-007) with its own criterion, droppable with no
   ripple; the byte-identical claim is stated over queues with no marker-bearing queued card.
3. **Refusing `moai todo --auto` in a lane** reverses a path SPEC-FACTORY-LANE-AUTONOMY-001's
   fallback doctrine named. Mitigation: isolated requirement (REQ-TAU-008); the lane's
   self-service pickup keeps working through the lease path; nothing else in FLA changes.
4. **Self-attested, unread decision records.** No board writer exists and no party re-reads the
   line (R4-iii). Unimplemented residual risk, stated in `spec.md` §G; not a control.
5. **`unmeasured` inputs read as "none".** The record form forces an explicit `unmeasured` token,
   and the criteria mutant-probe a record that omits it.
6. **Unmarked operator-decision cards (t810/t1294/t1383-class) stay selectable** until the operator
   or leader applies one of the two identification forms — the structural `hold` state or a leading
   `[보류` marker in the card text. Out of this SPEC's code; stated so the operator can act (plan.md
   §Operational follow-ups; the sentence of spec §B.3).
7. **MEMORY index drift** (INFERRED at first authoring): this SPEC cites cards from the memory index
   (t810/t1294/t1383) as the motivating phrase; their live states were first not queried and were
   read, read-only, at the iteration-2 repair (R12, ledger G5).

## R10. Not verified

- The live state of cards t810/t1294/t1383 was **not queried at first authoring**; the plan audit
  read it through the installed rc.25 binary (t1294 and t1383 `queued` with ordinary text, t810
  `picked`) and the iteration-2 repair re-read it (R12, ledger G5) — attributed to that binary's
  queue store, not to this tree's build.
- Whether the repo's template embed needs a regeneration step after editing
  `internal/template/templates/**`: **answered at iteration 3 (R13)** — two generated artifacts
  follow a template edit, the Codex TOML (`make agents-emit`) and the catalog hashes
  (`gen-catalog-hashes.go --all`, which `make build` also runs).
- Whether `MOAI_FACTORY_*` scrubbing is required for the doc-only tests (they read files; the
  lane-env hazard is for fixtures built through `runTodo`).
- The claim that no lane-writable decision-board verb exists (second-hand from t1400 via t1403).
- Exit code 4 availability in the `factory` verb family: within `factory*.go` only 1 and 3 are used
  (`factoryNextNoCardExit = 3`; `factory_lane_relaunch.go:51`); `moai slot` uses 4
  (`slotExitBusy = 4`, `slot.go:38`) in a different verb family, so exit codes do not collide; the
  run phase re-greps before choosing.
- Whether a Codex-backend kanban leader session carries `MOAI_KANBAN_BACKEND=gpt` in production
  (`kanban.go:535-548` says the backend is re-exported on the leader path; the plan audit inferred
  a Codex leader would carry it). R11 measures what the **guard** does with that variable alone,
  not whether a real leader carries it.

## R11. The lane predicate, measured (plan audit D3/D4 repair)

The predicate `factoryLaneRefusal()` (`factory_card.go:60-64`) is true when the lane role marker
equals `lane`, **or** the lane-label variable `MOAI_FACTORY_WORKER` is non-empty, **or**
`MOAI_KANBAN_BACKEND == gpt`. `factoryLaneAdmission()` (`:50-52`) is the role marker alone, and
`factory next` itself requires it (`factoryNotALaneError`, `factory_card.go:638`). A launcher that
stamps only the label (a recorded lesson says so) would not satisfy admission, so the guard for
`--auto` must key on **role marker or label**, not on the role marker alone.

Measured on `b3646de10`, with a binary built from this tree (setup S1) against a scratch queue
under a redirected `MOAI_HOME` (setup S2), each run from a shell with the lane variables unset;
`todoRefuseLaneMutation` (`todo.go:387`) lets a bare parent invocation through when `len(args) ==
0`, and `--auto` is a flag, not an argument:

| Run | Session environment | Result today |
|---|---|---|
| M1 / cell C4 | `MOAI_KANBAN_BACKEND=gpt` only (a non-lane session) | the serial cycle runs: `accept t1 …`, `unpick t1 …`, exit 0 |
| M2 / cell L18 | `MOAI_FACTORY_WORKER=lane-1` only | the serial cycle runs, exit 0 — a lane with only its label is **not** refused |

So (a) a guard on `factoryLaneRefusal()` would refuse M1, a Codex-backend session that is not a
lane and has no lease alternative (`factory next` refuses outside a lane) — the over-broad form the
audit warned about; (b) a guard on the role marker alone would not refuse M2. The SPEC's predicate
is the union of role marker and label, with the backend marker deliberately excluded. Setup rows
S1/S2 and cells L18/C4 are in `acceptance.md`; the scratch directory is machine-local.

## R12. Iteration-2 observations (plan audit N2, N3, N5, N6)

All measured on `63daaf6a7`.

- **The record API cannot undo a write (N2).** **READ** `internal/homestate/card_picked.go:114-145`:
  `RecordPicked` INSERTs a `cards` row at `picked` and appends a `card.transition` event (from empty,
  to `picked`); when the row exists and no fields are asked it returns the existing row unchanged.
  A search for delete verbs over non-test `internal/homestate` finds none (the audit's
  `DELETE`/`DeleteCard`/`RemoveCard` grep). So a failure after `RecordPicked` leaves a row and an
  event; `moai factory status` (`factory_card.go:1392`) lists the rows. The compensation of spec
  §B.8 is therefore bounded to the queue item.
- **The foreign-worktree refusal runs after the record write (N6).** **READ** `factoryNextClaim`
  (`factory_card.go:566-572`): `factoryRefuseForeignWorktree` runs inside the claim, after
  `RecordPicked` in `factoryNextRecordAndClaim`; only a read-only precheck in the nominated path
  keeps it from leaving a promoted card and a record row (spec token `foreign-worktree`).
- **The record has nineteen states (N6).** **READ** `homestate/card_record.go:20-45`: `picked`,
  `assigned`, `leased`, `plan`, `plan-audit`, `kickoff`, `run`, `sync`, `sync-audit`, `merge-ready`,
  `merging`, `merged-local`, `pushed`, `ci-green`, `done`, `needs-decision`, `blocked`, `failed`,
  `abandoned` — twelve in-flight (`leased` to `ci-green`), two leasable shapes (`picked`,
  `assigned`), five terminal or parked.
- **The marker predicate trims (N5).** **READ** `todo_auto_rank.go:~150`: `autoRankHoldMarked` is
  `strings.HasPrefix(strings.TrimSpace(text), "[보류")`.
- **Old-authority sweep (N3).** **OBSERVED** by `git grep -n -F -i` over `.claude`,
  `internal/template/templates`, `docs-site`, `CLAUDE.md`, `AGENTS.md`: besides the surfaces
  iteration 1 named, the old wording lives in the **generated** Codex artifact
  `internal/template/templates/.codex/agents/moai/manager-todo.toml` (L22, L34, L39) and in
  `auto-semantics.md` L186 (`authorizes serial queue consumption and nothing else`); every hit and
  its disposition is `plan.md` §4a. **Observed** baseline of the artifact parity test on the
  unmodified tree: `TestGoldenCommittedArtifactsMatchEmission` PASS (ledger G3). Not observed red
  (the auditor's gap G2): that it turns red when `manager-todo.md` is edited without regeneration is
  inferred from the test reading the template `.md` and the committed TOML and from `make build`
  depending on `agents-emit-check` (`Makefile:33-50`); AC-TAU-010 has the run phase observe it.
- **The acceptance-counter baseline was green before this repair edited `acceptance.md`** (ledger
  G4): `TestACCounterFullCorpusMatchesBaseline` `ok … 6.230s`.
- **The three operator-decision cards were re-read, read-only** (ledger G5): t810 `picked`, t1294
  and t1383 `queued`, none starting with the marker.

## R13. Iteration-3 observations (plan audit F1, F2)

- **The catalog hash obligation (F1).** **READ** `internal/template/scripts/gen-catalog-hashes.go`:
  `--all` updates every entry's `hash:` in place from the embedded templates (skill directories as
  whole trees via `ComputeDirTreeHash`, agent entries as the `.md`), default paths relative to the
  worktree root, yaml.v3 re-marshal. **READ** `internal/template/catalog_tier_audit_test.go:401` and
  `:475` (the two guards) and `Makefile` `build` (runs the generator as a side effect). **OBSERVED**
  (ledger G6, G7, G8): both guards PASS on the unmodified tree (49 entries, 37 directory entries);
  both FAIL on a tree with the three template artifacts perturbed (`CATALOG_HASH_UNSTABLE` for `moai`,
  `moai-kanban-foreman`, `manager-todo`; `CATALOG_HASH_SKINNY` for the two directories); the
  generator run on a scratch catalog rewrote exactly three lines (`3  3`); the tree was restored and
  proven clean (`git status --short` empty, `cmp` against the backups empty). The red is observed.
  The perturbation was one **interior** double space — a trailing space would have been normalized
  away by `NormalizeForHash` and the guards would have stayed green.
- **User-facing pages that restate the operator-only pick (F2).** **OBSERVED**, `git grep -n -i -E
  "always the operator|never picks|operator.s act|picks a card|operator-only|operator's acts|never pick"
  -- README.md README.ko.md README.ja.md README.zh.md docs-site/content`: `README.md:161`,
  `docs-site/content/en/advanced/factory-mode.md:65` and `:117`, and
  `docs-site/content/en/advanced/kanban-mode.md:287`. The ko/ja/zh pages and the three locale
  READMEs were not matched by the English pattern and are not enumerated (a gap; sync enumerates
  them). `docs-site/content/{en,ko,ja,zh}/advanced/{factory-mode,kanban-mode}.md` all exist.
