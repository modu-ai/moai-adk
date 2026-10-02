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
`internal/cli/mcp_factory_card.go:111`) shares the implementation and has the same three inputs
(`mcp_factory_card.go:42-48`).

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
§10: "detection, not prevention"); it lands in the card's progress record. SPEC consequence: the
requirement binds the line's *form and content*, and the sync audit's re-read is the compensating
control — the SPEC does not claim a board write it cannot specify.

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
sequential group"), which is arm (c)'s serial-slot rule. This SPEC therefore reads FLA's
"self-service pickup" as the lease path and does not edit FLA's body (completed SPEC).

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
- `TestAutoRankAgentDoctrine` requires both pinned literals in the manager-todo serial-cycle
  region and keeps `strict queue order` etc. absent — compatible with the amendment.
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

**Feasibility measurement** (scratch script, not committed): replacing exactly the sentences below
in `kanban-dispatch.md` lines 29, 31, 33 gives per-line deltas +72, −31, −30 = **+11 B** net
before any trimming, with every retained pin literal still present and the old fragment
`serial consumption` gone. So the budget is reachable by trimming ≥ 11 B more (e.g. the phrase
"in the gate inventory"); the run phase owns the final wording.

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
4. **Self-attested decision records.** No board writer exists (R4-iii). Residual risk is stated in
   `spec.md` §G; the sync audit's re-read is the control.
5. **`unmeasured` inputs read as "none".** The record form forces an explicit `unmeasured` token,
   and the criteria mutant-probe a record that omits it.
6. **Unmarked operator-decision cards (t810/t1294/t1383-class) stay selectable** until the operator
   `hold`s them. Out of this SPEC's code; stated so the operator can act (plan.md §Operational
   follow-ups).
7. **MEMORY index drift** (INFERRED): this SPEC cites cards from the memory index
   (t810/t1294/t1383) only as the motivating phrase; their live states were not queried.

## R10. Not verified

- The live state of cards t810/t1294/t1383 (not queried; the queue is the operator's).
- Whether the repo's template embed needs a regeneration step after editing
  `internal/template/templates/**` (Makefile not read); plan.md carries it as a run-phase
  pre-flight check.
- Whether `MOAI_FACTORY_*` scrubbing is required for the doc-only tests (they read files; the
  lane-env hazard is for fixtures built through `runTodo`).
- The claim that no lane-writable decision-board verb exists (second-hand from t1400 via t1403).
- Exit code 4 availability in the `factory` verb family (only 1 and 3 were seen in
  `factory_card.go`; the run phase measures before choosing).
