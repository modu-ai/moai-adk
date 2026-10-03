# SPEC-TODO-AUTO-PICK-001 — Plan

Tier M. Card t1448, plan-start HEAD `4bf547bca`, iteration-1 repair base `b3646de10`, iteration-2
repair base `63daaf6a7`, iteration-3 repair base `625f01718`, worktree `.moai/worktrees/t1448`, branch `WT-todo-auto-pick-autonomy`.
Milestones are ordered by **decision reversibility** — the decisions most likely to change come
first (the nominated-lease interface a lane sees, then the doctrine the operator reads), mechanical
edits last. No time estimates; priority labels and ordering only.

## 1. Approach in one paragraph

Add one opt-in nomination form to `moai factory next`, validated before any write and leased
atomically, with one shared keep-set predicate; refuse `moai todo --auto` inside a lane (by an
explicit lane predicate, with a dedicated refusal text and a corrected fallback instruction) so the
lease is the only lane pick path; then amend the doctrine sentences the card names — **replacing**
them, never appending, **live and mirror in one change, with the generated Codex agent artifact
regenerated in the same change** — and move the doc pins in that same change. All work is
TDD-first. Net bytes and characters on the always-loaded `kanban-dispatch.md` must not grow.

## 2. Milestones

### M1 — The nominated-lease interface and the lane predicate, tests first (Priority High)

The decisions most likely to change: the flag name, the refusal tokens, the exit code, the lane
predicate.

- **The only production edit M1 may make (N1).** One package-level function variable in
  `internal/cli/factory_card.go`, declared with its default and **not called by any production
  path**: `var factoryNominateBeforeRecord = func(cardID string) error { return nil }`, in the
  style of the existing `factoryCardNow` seam. It changes no behavior, so the tree after M1 is the
  unmodified tree plus one inert declaration. It exists so that the seam tests **compile** at M1
  (an assignment to an undeclared identifier is a compile error, not a RED) and fail at **runtime**
  for their stated reason (the nomination flag does not exist). M2 gives the variable its single
  call site, **before** `RecordPicked` (N2).
- Write the new tests in `internal/cli/factory_nominate_test.go` against the existing fixtures
  (`fcFixture`, `fcQueue`, `fcClassify`, `sdRegisterLane`, `sdLaneEnv`, `runFactory`), **scrubbing
  the lane environment first** (`sdClearLaneEnv`). Test set — each is the swept set of the
  criterion named:
  - AC-TAU-001: `TestFactoryNextNominateLeasesNominee`, `TestFactoryNextNominateUnknownCard`,
    `TestFactoryNextFlagSet` (the `--help` flag set equals the baseline plus `--card`),
    `TestFactoryNextNominateMCPParity` (the MCP form's `card` input; inputs are `run`,
    `project_root`, `card`).
  - AC-TAU-002: `TestFactoryNextNominateConcurrentLanes`, `TestFactoryNextNominateSameCardExactlyOne`.
  - AC-TAU-004: `TestFactoryNextNominateRefusesKeepSet` (subtests `held`, `hold-marker`,
    `marker-leading-space`, `marker-mid-text` (leases), `blocked`, `serial-slot`, `dropped`,
    `owned`), `TestFactoryNextArmCSkipsHoldMarker` (the same three marker cards on the bare path).
  - AC-TAU-014: `TestFactoryNextNominateQuotaHold`, `TestFactoryNextNominateBackendSkip`,
    `TestFactoryNextNominateRecordStateTokens` (one subtest per § C.2 record-state class: `owned`,
    `recorded`, `foreign-worktree`, and the three leasable shapes),
    `TestFactoryNextNominateRefusalLeavesStateUnchanged` (byte-compares queue and record),
    `TestFactoryNextNominatePromoteThenLose`, `TestFactoryNextNominateClaimRefusedRollsBack`,
    `TestFactoryNextNominateCompensationFailure` (subtest `item-moved`).
  - AC-TAU-005: `TestTodoLaneRefusesAutoCycle` (subtests `label-only`, `role-only`,
    `role-and-label`, each asserting the queue byte-identical), `TestTodoLaneAutoRefusalText`,
    `TestTodoNonLaneGPTSessionNotRefused`, `TestFactoryFallbackDeclarePrintsLeasePath`.
  - AC-TAU-006: `TestFactoryNextBareUnchanged` (the golden) and
    `TestFactoryNextAllMarkerQueueExitsNoCard`.
- **M1 exit (D2, N1).** The tree measured is **the unmodified tree plus the seam declaration**.
  (a) Every new test **except** the golden and the non-lane GPT guard **compiles** and is observed
  RED at runtime for its stated reason (the flag does not exist, the string is still there, the
  lane runs the cycle, the marker card is leased) — never a compile error: `go vet ./internal/cli`
  passes and the new tests build under an anchored selector over their names. (b) `TestFactoryNextBareUnchanged`
  and `TestTodoNonLaneGPTSessionNotRefused` are GREEN on that tree (they pin current behavior and
  must stay green). (c) Their non-vacuity is shown by a **seeded perturbation** recorded in
  `progress.md` §E.2 with command, verbatim stdout, exit code and tree SHA: for the golden, a
  one-line mutation of `factory_card.go` (for example moving the `noNewCards` early return, or
  swapping the arm (b)/(c) order); for the GPT guard, the over-broad predicate
  (`factoryLaneRefusal()` in the new guard). Each mutation is then reverted (`git diff` shows only
  the seam declaration) before M2. (d) The scoped baselines of R3, R11 and G2 re-run and still
  green.
- **The golden's minimum case set** (also in AC-TAU-006): default arm order (a)→(b)→(b2)→(c) over
  assigned / operator-picked / queue-picked / queued cards; `--wait` with a bounded wait through
  the `factoryNextWaitSleep` seam; the quota hold (`noNewCards`) with an assigned card still
  leasing; the Codex skip (`factoryNextSkipForBackend`) of a card at or past merge-ready; the serial
  slot (a second serial card not leased while one is in flight); and the no-card exit 3 with its
  stdout line.
- The refusal exit code is **4**. M1's tests assert the **literal** `4`; no constant by the name
  `factoryNextRefusedExit` (or any name) exists at the plan tree (`git grep -n "factoryNextRefusedExit"
  -- internal/cli` is empty), so the run phase **introduces** the named constant in M2 and no M1 test
  may reference it (S1). Pre-flight: measured, within
  `factory*.go` only 1 and 3 are used (`factoryNextNoCardExit = 3`, `factory_lane_relaunch.go:51`);
  `moai slot` also uses 4 (`slotExitBusy = 4`, `slot.go:38`) — a different verb family, so exit
  codes do not collide; the run phase re-greps before choosing.

### M2 — Implement nomination and the shared keep-set predicate (Priority High)

- `internal/cli/factory_card.go`: add the `--card` flag to `newFactoryNextCommand`; add a
  nomination path beside `factoryNextSelectAndLease` that reuses `factoryNextClaim` and
  `factoryNextRecordAndClaim` (the version-checked edges — no second lease route); extract one
  keep-set predicate `factoryKeepSetRefusal` returning the § C.2 token, applied by the nomination
  path **and** by arm (c) for the `[보류` skip only (REQ-TAU-007).
- **State semantics (D7, N2), the mechanism.** The nominated path runs four steps in order; spec
  §B.8 states what each can and cannot undo:
  1. *Validate, read-only.* One pure queue read plus one factory-record read decide every § C.2
     token — `unknown-card`, `dropped`, `held`, `owned`, `recorded`, `hold-marker`, `blocked`,
     `serial-slot`, `quota-hold`, `backend-skip` — and run the foreign-worktree precheck
     (`factoryRefuseForeignWorktree`) that `factoryNextClaim` would otherwise run only after a
     promotion, mapping it to `foreign-worktree`. The record-state mapping is the § C.2 table (all
     nineteen states). No write has happened when any of these refuses.
  2. *Promote inside the queue `Mutate`, re-validating.* A `queued` nominee is promoted to `picked`
     only if, inside the lock, its state is still `queued` and the predicate still passes; if it is
     no longer `queued` another lane moved it first and the invocation refuses `raced` without
     writing. A nominee already `picked` and unowned (operator pick, arms (b)/(b2)) skips this
     step; a nominee assigned to this lane is arm (a)'s.
  3. *Call the seam, then claim.* `factoryNominateBeforeRecord(cardID)` is called **after the
     promotion and before `RecordPicked`**; a non-nil error is treated as a claim failure. Then the
     claim runs through `factoryNextRecordAndClaim` / `factoryNextClaim`.
  4. *Compensate.* On a claim failure (a seam error, a lost race, or a store error) the invocation
     reads the record: if there is **no row, or a row at `picked` with no owner**, it restores the
     queue item from `picked` to `queued` in one `Mutate` that acts **only if the item is still
     `picked`**; if another holder exists the queue state is the holder's and is left alone and
     the invocation refuses `raced`. The restore condition is exactly this one — **no row, or a row
     at `picked` with no owner** — and spec REQ-TAU-005 and §B.8 carry the same words (S2); a row
     `assigned` to this lane (a first claim edge landed, a later edge failed for a non-race reason)
     meets neither alternative, so nothing is restored: it is this lane's own arm (a) lease target
     and is untested by design. If the item is no longer `picked` the compensation does
     nothing and the original outcome is reported. If the restoring write itself fails the
     invocation exits with status 1 and `factory next: compensation failed: <cause>` — no token —
     and the card stays `picked` and unowned (adoptable by an unnominated arm or a later
     nomination).
  **What the record API cannot undo:** `RecordPicked` INSERTs a row and appends an event and the
  package has no delete, so a failure *after* `RecordPicked` leaves that row (`picked`, unowned) and
  its event; the queue item is still restored per step 4. The seam sits before `RecordPicked` so
  the tests can prove the restore and the absence of a row; the post-record case is specified in
  §B.8, accepted, visible as a `picked` row with no owner in `moai factory status`, and **not
  tested**. The residue differs from what arm (c) itself leaves: arm (c) leaves the queue item
  `picked` **and** the row `picked`, while a failed nominated claim leaves the queue item `queued`
  with a `picked` row; arm (b) skips it (the item is `queued`), and it re-adopts through arm (c)
  (S2). Side effect, accepted and stated in spec §G (S7): `factorySerialSlotFree("picked")` is
  false, so the stranded `picked` row of a serial card holds the serial slot against other serial
  cards until it is re-adopted.
- The arm (c) skip counts the marker card as seen (`sawQueued`), so a queue holding only
  marker-bearing cards ends on the no-card exit 3 (AC-TAU-006). The marker test is
  `strings.HasPrefix(strings.TrimSpace(text), "[보류")` — the same predicate as
  `autoRankHoldMarked` (`todo_auto_rank.go:~150`): leading whitespace still opens with the marker;
  a mid-text mention does not.
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

### M5 — Doctrine amendment, live and mirror and generated artifact in one change, pins first (Priority High)

Everything below is **one commit**: no state exists in which a live file is edited and its mirror is
not, in which a doc is edited and its pin is not, in which `manager-todo.md` is edited and the
generated Codex artifact is not, or in which a template artifact is edited and its stored catalog
hash is not (F1).

**The catalog hash obligation (F1), measured at `625f01718`.** `internal/template/catalog.yaml`
stores a sha256 for three artifacts this milestone edits in the template tree: the **whole skill
directory** `templates/.claude/skills/moai/` (it contains `workflows/gtd.md`), the **whole skill
directory** `templates/.claude/skills/moai-kanban-foreman/` (it contains `SKILL.md`), and the **file**
`templates/.claude/agents/moai/manager-todo.md`. `TestManifestHashFormat` recomputes every entry's hash
from the embedded templates and fails `CATALOG_HASH_UNSTABLE` on a mismatch;
`TestCatalogHashCoversSkillSubfiles` fails `CATALOG_HASH_SKINNY` for a directory entry
(`internal/template/catalog_tier_audit_test.go`, `:401` and `:475`). The hashes are regenerated by
`go run ./internal/template/scripts/gen-catalog-hashes.go --all` — the exact step `make build` runs as
a side effect (`Makefile` `build` target), which leaves a **modified, unstaged** `catalog.yaml` for a
lane that stages by pathspec, so the run phase runs the generator itself and stages the file. The
script header states: `--all` updates **every** entry's `hash:` field in place (default catalog path
`internal/template/catalog.yaml`, default templates path `internal/template/templates`, both
relative — run it from the worktree root); skill directories hash as a whole tree
(`ComputeDirTreeHash`), agent entries hash their `.md`; the file is re-marshalled with yaml.v3 (the
script notes comments are not preserved — the committed file has none). Measured on a scratch copy of
the catalog with the three template artifacts perturbed: the generator rewrote **exactly three
lines**, `git diff --no-index --numstat` read `3  3` (the `moai`, `moai-kanban-foreman` and
`manager-todo` `hash:` lines), nothing else moved. Precedent for the same kind of edit: commit
`232cd8d41` (a `manager-todo.md` edit that also changed `catalog.yaml`).

Order inside the milestone: (1) write `TestAutoPickDocDoctrine` and `TestAutoPickMirrorParity` in
`internal/cli/todo_auto_pick_doc_test.go` — observe RED (AC-TAU-007/-008 RED-now cells);
(2) edit every doc and every mirror — then, **before** regenerating anything, run the generated-artifact
parity test once and the two catalog guards once, and record each red once (the AC-TAU-010 seeded
probe, S3: the TOML is stale against the edited `.md`; `moai`, `moai-kanban-foreman` and
`manager-todo` read `CATALOG_HASH_UNSTABLE`); **then run `make agents-emit`** to regenerate
`internal/template/templates/.codex/agents/moai/manager-todo.toml` from the edited template
`manager-todo.md` (never hand-edit the TOML); **then run the catalog generator**
(`go run ./internal/template/scripts/gen-catalog-hashes.go --all`), only after every template
artifact is final (a later template edit makes the hashes stale again); (3) run the scoped doc-pin
command once **before** moving any pin, and record the two expected breakages; (4) move the existing
pins in `internal/cli/todo_auto_doc_test.go`; (5) run `AGENTEMIT_UPDATE= go test ./internal/template/agentemit/...
-run '^TestGoldenCommittedArtifactsMatchEmission$' -count=1` and see it PASS (it compares the committed
TOML to the template `.md`; `make build` depends on it through `agents-emit-check`), and run
`go test ./internal/template -run '^(TestManifestHashFormat|TestCatalogHashCoversSkillSubfiles)$' -count=1 -v`
and require both `--- PASS` lines (the swept set is those two named tests; both are PASS on the
unmodified tree — observed ledger G6 — so they are regression-guards that go red only when the
catalog is left stale).

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
| `TestGoldenCommittedArtifactsMatchEmission` (`internal/template/agentemit`) | the committed `.codex/agents/moai/manager-todo.toml` equals the emission of the template `manager-todo.md` | **must be regenerated** with `make agents-emit` in the same commit (baseline PASS observed, ledger G3) |
| `TestManifestHashFormat` and `TestCatalogHashCoversSkillSubfiles` (`internal/template/catalog_tier_audit_test.go`) | the stored sha256 of `moai`, `moai-kanban-foreman` (whole skill trees) and `manager-todo` (file) equals the hash recomputed from the embedded templates | **must be regenerated** with `go run ./internal/template/scripts/gen-catalog-hashes.go --all` in the same commit (baseline PASS observed, ledger G6; RED observed on a perturbed tree, ledger G7) |
| `TestContractModeEmitterSites` (`contract_mode_guided_test.go`) | every document containing the word `Kickoff` is classified; `kanban-dispatch.md` is `R` | **kept** — the new `auto-semantics.md` §9.3 and `gtd.md` text must not add the word `Kickoff` to a file the registry does not classify (use "plan→run gate" wording) |
| `internal/template/jev_auto_exception_test.go` | presence anchors `jaeAnchors` in `kanban-dispatch.md`, `gtd.md`, `manager-todo.md` | **kept** (read: markers are Jev tokens the edit leaves) |
| `internal/cli/init_headroom_export_test.go` | a path list for a measurement harness (skips without its flag) | **kept** — no assertion on content |
| `internal/template/rule_template_mirror_test.go` | read in iteration 1: no `kanban-dispatch`/`gtd`/`manager-todo` marker | **kept** |

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
| `gtd.md` `--auto` section (L334-L361) | the `[HARD] … authorizes serial consumption of the queue and nothing else` clause | the same clause with `on its own judgment`; the lane routing sentence (`a lane session exercises the --auto authorization through moai factory next …`, no backticks around the literal); the **keep-set list**; the **record form** (`decision record:` with `ladder_path=gate-row card pick (AUTONOMOUS, auto-semantics §9)` and `unmeasured`, located in the card's progress record — a fresh `§J` section, not `§F` — as evidence); the D12 sentence naming both treatments of the marker; the "exactly one card is in flight" sentence scoped to the operator-session cycle |
| `auto-semantics.md` §9 card-pick row (L169 at plan start, L199 at `b03619b29`) | the row text | keeps the AUTONOMOUS disposition, adds the pointer `§9.3` |
| `auto-semantics.md` §9.2 sentence (L186 at plan start, L216 at `b03619b29`, **N3b**) | `It is distinct from the \`--auto\` batch authorization of the card pick row, which authorizes serial queue consumption and nothing else.` | the same sentence with the authority stated as in §9.3 (the invoked session takes cards on its own judgment, each only through a lease); the old literal `authorizes serial queue consumption and nothing else` is absent afterwards |
| `auto-semantics.md` new `### 9.3` | — | the keep-set, the input set, the record form with `ladder_path=gate-row card pick (AUTONOMOUS, auto-semantics §9)`, `evidence, not the decision board`, `unmeasured`, the open-set sentence `adds an input without amending the keep-set or the lease path`, and `no party re-reads the card-pick record` (the §9.1/§10 "compensating control" wording stays and is **not edited**; §9.3 says it does not hold for this record) |
| `manager-todo.md` L23, L34-L35 (template copy and live copy) | `process cards in queue order`; `serial consumption of the queue and nothing else` | `on its own judgment` / keep-set wording; the serial-cycle region keeps the pinned ranking literals |
| `.codex/agents/moai/manager-todo.toml` (template tree only) | — **generated**, not edited | regenerated by `make agents-emit` from the edited template `manager-todo.md`; the three old phrases are absent afterwards |
| `internal/template/catalog.yaml` (template tree only) | — **generated**, not edited | the `hash:` lines of `moai`, `moai-kanban-foreman`, `manager-todo` regenerated by `go run ./internal/template/scripts/gen-catalog-hashes.go --all` after the template edits are final; never hand-edited (F1) |
| `moai-kanban-foreman/SKILL.md` Boundary 1, step 4 | `serial consumption in queue order is authorized`; `\`queued\` items are not yours to pick.` | `… on the iteration's own judgment, within the keep-set`; `… not yours to pick outside a batch authorization.` |
| `kanban-dispatch-detail.md` § The pre-dispatch cross-check | — | one added paragraph carrying `a pull request or landed state is a skip input for a queued candidate the session chose and is report-only for an operator-picked card` and the B.2 reconciliation. Lazy file — growth is free |
| `moai-mcp-tools-catalogue.md` factory_next row | the one cell | mentions the optional `card` argument |

New detail goes in `auto-semantics.md` §9.3 (lazy, `paths:`-scoped to the watchdog skill), `gtd.md`
(a lane reads it) and `kanban-dispatch-detail.md` — **never** into the always-loaded stub beyond the
replaced sentences. Preserve each copy's `kanban-dispatch.md` live-only `moai worktree sweep …`
sentence exactly (line 177 at plan start, line 181 at `b03619b29`; see §4).

- Exit: the new doc pins GREEN **for live and mirror at once**; the moved existing pins GREEN;
  `TestGoldenCommittedArtifactsMatchEmission` PASS; `TestManifestHashFormat` and
  `TestCatalogHashCoversSkillSubfiles` PASS; every AC-TAU-007/-008/-010/-011/-013 reading
  holds on the committed tree; the mutants of `acceptance.md` each fail their criterion.

### M6 — Verification and measurement only (Priority High; no file edits)

Run only what the change can affect, then let CI run the rest (`AGENTS.md` §4); record every command,
exit code and verbatim tail in `progress.md` §E.2:

- `internal/cli`: the new tests plus `^(TestFactoryNext|TestAutoRank|TestAutoPick|TestAutoHelp|TestTodoAuto|TestTodoLane|TestTodoNonLane|TestTodoSkill|TestFactoryFallback)`; take the `moai slot` lease for the
  package run (`.claude/rules/local/gitflow-lane-protocol.md` §8) and scrub the lane environment in
  one compound invocation.
- `internal/template`: the tests naming the edited files (`jev_auto_exception_test.go`,
  `contract_mode_guided_test.go`, `rule_template_mirror_test.go`) **and the two catalog guards
  `TestManifestHashFormat` and `TestCatalogHashCoversSkillSubfiles`** by one anchored selector, with a
  swept-count check (the catalog guards' own log lines report 49 entries and 37 directory entries
  swept at `625f01718`; a zero or a lower count is a gap); and `AGENTEMIT_UPDATE= go test
  ./internal/template/agentemit/... -run '^TestGoldenCommittedArtifactsMatchEmission$' -count=1 -v`
  (the swept set is the single named test: its `--- PASS` line must appear).
- `internal/spec`: `go test ./internal/spec -run '^TestACCounterFullCorpusMatchesBaseline$' -count=1`
  after the criteria are final (the acceptance counter baseline; the count is final in the plan
  commits and the guard is green, so this run is a confirmation — a regeneration, if it were ever
  needed, belongs to a plan commit, not to M6; S6).
- Measurements (AC-TAU-010, -011, -012, -013): `cmp` on each pair, `wc -c` and `wc -m` on both
  `kanban-dispatch.md` copies against the merge-base blobs (`CARD_BASE`, §3), `git diff --numstat "$CARD_BASE"..HEAD` against
  the caps, the boundary pathspec probe, and the draft-versus-actual byte comparison (§3).
- `GOOS=windows GOARCH=amd64 go build ./...`; `golangci-lint` at the CI version.
- The commit body of M5 carries the measured before/after bytes **and characters** and the
  non-invoking-cost sentence (`rule-authoring.md` (c)).

### Milestone ↔ criterion order check (D1, N1)

Each criterion's green path names a milestone, and every obligation that criterion reads is
completed in that milestone or an earlier one; **every test an exit names compiles at that
milestone** (the only identifier a test needs that production code does not yet have is the M1 seam
declaration, which M1 itself adds). Checked by hand for **all** fourteen, again after the N1 repair:

| AC | Green at | Reads work from | Order holds because |
|---|---|---|---|
| AC-TAU-001 | M2 (CLI rows), M4 (MCP row) | M2, M4 | the MCP row is stated against M4 only; M2 does not claim it |
| AC-TAU-002 | M2 | M2 | uses the nominated path only |
| AC-TAU-003 | M5 | M5 | doctrine literals, live and mirror, one commit |
| AC-TAU-004 | M2 | M2 | the arm (c) skip and the nominated refusals are both M2 |
| AC-TAU-005 | M3 (Go), M5 (routing sentence) | M3, M5 | the Go rows do not read the docs; the routing-sentence row is stated against M5 |
| AC-TAU-006 | golden GREEN at M1 (unmodified tree plus the seam declaration); stays GREEN at M2; all-marker test GREEN at M2 | M1, M2 | the golden is GREEN before any behavior change, the seam declaration is inert, so no RED/GREEN conflict; its tests compile at M1 |
| AC-TAU-007 | M5 | M5 | live, mirror, generated artifact and pins in one commit |
| AC-TAU-008 | M5 | M5 | same commit |
| AC-TAU-009 | no milestone flips it: the first lane lease taken under the doctrine, after the run phase | the lease | a regression-guard with no executing party; recorded as a pass only once a reader has opened the record |
| AC-TAU-010 | M5 (measured M6) | M5 | both sides, the generated Codex artifact and the catalog hashes edited together; the parity test and the two catalog guards run at M5 step 5, the seeded probes at step 2 |
| AC-TAU-011 | M5 (measured M6) | M5 | measurement only at M6 |
| AC-TAU-012 | M6 | all | boundary read over the finished diff |
| AC-TAU-013 | M5 (measured M6) | M5 | floor and ceiling both read the M5 commit |
| AC-TAU-014 | M2 | M1 (seam declaration), M2 | the seam tests compile at M1 and are RED at runtime there; M2 gives the seam its call site before `RecordPicked`, and all seven tests flip at M2 |

## 3. Always-loaded growth: the constraint, the method, the draft

`kanban-dispatch.md` is always-loaded, the repo is already over its instruction budget (t1318:
214,155 chars against 210k), and `rule-authoring.md` requires a statement for growth over 1,000 B.
This SPEC requires **net growth ≤ 0 on both copies, in bytes and in characters**, measured
**relative to the card's merge base with develop**: the bound of each copy is the size of that file's
blob at `CARD_BASE`, the output of `git merge-base develop HEAD`, **re-derived at reading time and
never pinned**. (Post-audit re-pin: the plan was audited against fixed figures taken at `b3646de10`
— live 26,959 B / 26,754 chars, mirror 26,637 B / 26,433 chars, the file unchanged since
`4bf547bca` — and develop `7109e0900` was absorbed afterwards. Develop's own edits to the file,
numstat `5 1`, added 1,349 B per copy, which this card's edit cannot and should not remove, so the
fixed figures became unsatisfiable while the intent, that this card does not grow the stub, did not
change.) For information only, at HEAD `b03619b29`: live 28,308 B / 28,099 chars, mirror 27,986 B /
27,778 chars, equal to the merge-base blobs (acceptance rows L48-L52). Before the first M5 edit the
two sides are equal by construction; the reading is valid before the card merges into develop and
reads "not measurable" after it (the merge base is then the card tip).

Measuring commands, both copies, in two steps (the form the worktree guard accepts): the working
file with `wc -c` and `wc -m` of `.claude/rules/moai/workflow/kanban-dispatch.md` and
`internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md`; the bound by
`git merge-base develop HEAD` (prints a SHA), then `git cat-file -s <that-sha>:<path>` (bytes) and
`git show <that-sha>:<path>` piped to `wc -m` (characters).

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
reachable, not the binding figure. (The iteration-2 audit drafted its own replacements and measured
−135 B / −135 chars, which agrees on direction.) The deltas describe the three replaced lines, not the
file size, and those lines are unchanged by the develop absorption (L29, L31, L33: same line numbers,
same text, re-read at `b03619b29`; develop's hunks sit at lines 96-118), so they still apply; the draft
script was not re-run here (it is not committed).

**The `auto-semantics.md` cap (N3b) — measured the same way.** A scratch draft of the new `### 9.3`
section, carrying every § C.1 literal, is **27 lines** (`wc -l` of the draft); the §9 row edit and
the §9.2 sentence edit are one line each. So the edit adds about 29 lines and deletes 2. AC-TAU-013
caps `auto-semantics.md` at **45 added / 2 deleted** per copy: the draft plus roughly half again for
rewording, and exactly the two replaced single-line paragraphs. The draft is not committed; the
binding measurement is `git diff --numstat` on the committed edit.

## 4. Mirror handling (preserve, do not absorb)

The live and mirror `kanban-dispatch.md` differ by one pre-existing line (the `moai
worktree sweep …` sentence, present in the live copy only; research R7; live line 177 at plan start,
live line 181 at `b03619b29` after develop's insertions above it — identify it by its text and read
its line from `cmp` at reading time). Preserve each copy's sentence
exactly; do not sync it either way. Reason: absorbing it is another card's act, it would put an
unreviewed sentence into every user project's always-loaded rule, and the existing parity test is
deliberately scoped to the amended passage. After the edits `git diff --no-index --numstat` of the
pair must still read `1  1` (AC-TAU-010). The completion report names the drift for a follow-up card.
The Makefile's embed refresh, if it defines one beyond `agents-emit`, runs after the edits
(pre-flight; research R10). **The Codex agent artifact is a second kind of mirror:** generated, not
copied, regenerated in the same M5 commit (§2 M5).

## 4a. The old-authority sweep (N3) — every surface that states the old rule

Measured at `63daaf6a7` by `git grep -n -F -i` for `serial consumption`, `serial queue consumption`,
`queue order`, `never picks for the operator`, and `nothing else` (with `auto`/`batch`/`serial`
context) across `.claude`, `internal/template/templates`, `docs-site`, `CLAUDE.md`, `AGENTS.md`;
the first sweep's `docs-site` hits were vendored minified JavaScript (a false positive of the
`nothing else` pattern); iteration 3 found that those four patterns did not match the user-facing
wording "always the operator" / "operator's acts" / "never picks", so the last two rows were
corrected (F2). Each hit is classified:

| Hit | Disposition |
|---|---|
| `.claude/rules/moai/workflow/kanban-dispatch.md` L29 (`never picks for the operator`), L31 (`serial consumption … nothing else`) — live and mirror | **edited** (M5) |
| `.claude/skills/moai/workflows/gtd.md` L335 (`serial consumption of the queue and nothing else`) — live and mirror | **edited** (M5) |
| `.claude/agents/moai/manager-todo.md` L23 (`process cards in queue order`), L35 — live and template | **edited** (M5) |
| `internal/template/templates/.codex/agents/moai/manager-todo.toml` L22, L34 (generated from the template `.md`) | **regenerated** (M5, `make agents-emit`) |
| `.claude/skills/moai-kanban-foreman/SKILL.md` L69 (`consumption in queue order is authorized`), L168 — live and mirror | **edited** (M5) |
| `.claude/rules/moai/workflow/auto-semantics.md` L186 at plan start, L216 at `b03619b29` (`authorizes serial queue consumption and nothing else`) — live and mirror | **edited** (M5; the second place the old authority is stated) |
| `.claude/rules/moai/workflow/auto-semantics.md` L169 at plan start, L199 at `b03619b29` (card pick row) — live and mirror | **edited** (pointer to §9.3) |
| `gtd.md` L47, L55 (`SORTED queue order`, `the queue order under the lock`) | **kept** — describes how the queue is sorted and repositioned, not who may pick |
| `gtd.md` L340 and `manager-todo.md` L40 (`queue order within a priority`) | **kept** — the ranking source of the serial cycle, pinned by `TestAutoRankDoctrineAmendment`/`TestAutoRankAgentDoctrine` and unchanged for the operator-session cycle |
| `.codex/agents/moai/manager-todo.toml` L39 (the same ranking sentence) | **kept** — generated from the kept `manager-todo.md` sentence |
| `.claude/rules/moai/workflow/contract-autonomy.md` L92 (`Card selection, which stays the operator's act.`) | **kept, flagged** — it lists what a signed *contract* cannot permit (`workflow.autonomy.mode: contract`); the `--auto` batch authorization is the card-pick gate's own autonomous form (`auto-semantics.md` §9), a separate authorization path. A follow-up card may cross-reference the two |
| `internal/cli/todo_auto.go` L26 (Go comment: the cycle "authorizes serial consumption of the queue and nothing") and the `--auto` flag help (`todo.go` L315) | **kept** — accurate for the operator-session serial cycle, which this SPEC leaves unchanged; the flag help is named in spec §G |
| `internal/cli/todo_auto_doc_test.go` L344-L346 (a Go comment and a stale-phrase list quoting the old agent wording) | **kept** — a historical quotation pinned as an absent phrase |
| `.claude/rules/moai/core/askuser-protocol.md` L111 (`withholds a recommendation and nothing else`) | **kept** — unrelated sense of "and nothing else" |
| `README.md` L161; `docs-site/content/en/advanced/factory-mode.md` L65; `docs-site/content/en/advanced/kanban-mode.md` L287 (F2) | **sync-phase scope, not run-phase** — user-facing pages that state the operator-only pick: "The actor that picks a card is always the operator (`moai todo next <n>`) … never picks one itself", "picking the next one remain the operator's acts", "both putting cards in the queue and picking them stay the operator's job — the foreman never picks". Found by `git grep -n -i -E "always the operator|never picks|operator.s act|picks a card|operator-only|operator's acts|never pick" -- README.md README.ko.md README.ja.md README.zh.md docs-site/content` (3 English hits; the four README locales and the ko/ja/zh docs-site pages were not matched by this English pattern and are **not enumerated** — sync enumerates them). The bare-`/loop` foreman statements stay true (outside a batch authorization the foreman still picks nothing); the damage is the unqualified "always the operator". The sync phase (`manager-docs`, 4-locale docs-site: en, ko, ja, zh) rewrites each such sentence to the new wording — the pick is the operator's, in person or in advance through `--auto` — and applies the same edit to every locale's `advanced/factory-mode.md` and `advanced/kanban-mode.md` and to the READMEs; **the run phase does not edit them** (Out of Scope, spec §D) |
| `docs-site/**` vendored JavaScript hits | **kept** — the `nothing else` pattern's false positives |

## 5. Files to modify

| Area | File | Change |
|---|---|---|
| Go | `internal/cli/factory_card.go` | the seam declaration (M1); `--card` flag, nomination path (validate / promote / seam / claim / compensate), `factoryKeepSetRefusal`, arm (c) `[보류` skip counted as seen, refusal exit/diagnostic (M2) |
| Go | `internal/cli/mcp_factory_card.go` | optional `card` parameter, shared implementation |
| Go | `internal/cli/todo.go` | lane refusal of `--auto` in `todoRefuseLaneMutation` (lane-session predicate, dedicated text) |
| Go | `internal/cli/factory_messaging.go` | the printed fallback instruction (and its comment) point at the lease path |
| Go test (new) | `internal/cli/factory_nominate_test.go` | nomination, keep-set, state semantics, concurrency, bare golden, arm (c), flag set, MCP parity, lane `--auto`, GPT guard, fallback string |
| Go test (new) | `internal/cli/todo_auto_pick_doc_test.go` | doc contract literals, stale-absent, mirror parity, diff bound, mutants |
| Go test (edit) | `internal/cli/todo_auto_doc_test.go` | move the two markers of the § M5 table |
| Rule (live + mirror) | `kanban-dispatch.md` | replace three paragraphs (net ≤ 0 B and chars against the merge-base blob; the live-only sentence at line 177, now 181, preserved) |
| Rule (live + mirror) | `kanban-dispatch-detail.md` | one paragraph (D10 sentence) |
| Rule (live + mirror) | `auto-semantics.md` | §9 row pointer + the §9.2 sentence (L186, now L216) + new §9.3 |
| Skill (live + mirror) | `.claude/skills/moai/workflows/gtd.md` | `--auto` section: replacement, keep-set list, record form, lane routing, D12 sentence |
| Agent (live + mirror) | `.claude/agents/moai/manager-todo.md` | two sentence replacements |
| **Generated** (template tree only) | `internal/template/templates/.codex/agents/moai/manager-todo.toml` | **regenerated** with `make agents-emit` in the same M5 commit; never hand-edited; `TestGoldenCommittedArtifactsMatchEmission` verifies |
| **Generated** (template tree only) | `internal/template/catalog.yaml` | the `hash:` lines of `moai`, `moai-kanban-foreman`, `manager-todo` **regenerated** with `go run ./internal/template/scripts/gen-catalog-hashes.go --all` in the same M5 commit after the template edits are final; never hand-edited; `TestManifestHashFormat` and `TestCatalogHashCoversSkillSubfiles` verify (F1; measured diff `3  3`) |
| Skill (live + mirror) | `.claude/skills/moai-kanban-foreman/SKILL.md` | Boundary 1 and step 4 |
| Rule (live + mirror) | `.claude/rules/moai/core/moai-mcp-tools-catalogue.md` | one catalogue cell |
| Baseline (tracked) | `.moai/reports/t338/ac-count-baseline.txt` | not expected to change: `TestACCounterFullCorpusMatchesBaseline` is green with this SPEC reported `absent-from-snapshot … COUNT 14` (observed; see progress §E.1); a regeneration, if ever needed, belongs to a plan commit (S6) |
| Docs (**not** run-phase) | `README.md`, `docs-site/content/{en,ko,ja,zh}/advanced/{factory-mode,kanban-mode}.md` | **sync-phase** rewrite of the "always the operator" sentences (§4a, F2); listed so the sync phase has its scope, **not edited by run** |

No file under `internal/kanban/**`, `internal/graph/**`, or any schema file changes (REQ-TAU-003);
no `sync-auditor` / `sync-audit-4dim` change (REQ-TAU-012 states the record is unread, not
controlled).

## 6. Forward-compatibility with t1454

What this SPEC **leaves open**: the input set is open — the record's `evidence_refs` and the
selection rule name the inputs t1448 needs and say a later card adds an input without amending the
keep-set or the lease path (REQ-TAU-013, **unchanged by either audit repair**). Relation records are
read through whatever the lane-read surfaces expose (today `moai todo why`); no code here names the
queue-findings store or the `gtd_relations` table as *the* source. File overlap is a pluggable input:
expected-file data when present; otherwise a fallback that is INFERRED and unmeasured (a lane reading
another lane's branch by name from its own tree); `unmeasured` is the expected value until measured,
and the record does not require it.

What this SPEC **must not pre-empt**: it adds no schema field, no relation kind, no `moai graph`
change, no expected-file, parent/spawned-by or size field, and no mechanical refusal keyed on a
relation store (the relation-blocked refusal is deliberately *not* in the keep-set predicate;
probe O2). The mechanical guard reads only `state`, the `[보류` marker, `classification.blocked`,
the serial slot, ownership and the record state — none of which t1454's inputs change.

## 7. Operational follow-ups (not part of this SPEC's code)

- **Before lanes exercise this doctrine** (the handoff sentence, identical in spec §B.3, REQ-TAU-011,
  spec §G and Definition of Done 6): **Today t810 (`picked`), t1294 and t1383 (`queued`, ordinary
  text) carry neither the structural `hold` state nor a leading `[보류` marker, so the keep-set does
  not mechanically identify them. The SPEC supports exactly two identification forms — (a) the
  structural `hold` state, written only by `moai gtd hold`, and (b) card text that begins with the
  `[보류` marker — and a lane may write neither. The operator or leader must apply one of them to
  each such card before lanes exercise this doctrine.** Read with the read-only `moai gtd show`;
  the queue is not touched by this plan.
- The live-only `moai worktree sweep …` sentence drift in `kanban-dispatch.md` (line 177 at plan start,
  line 181 now) needs a follow-up card to decide which copy is right.
- The lane bootstrap notice still says bare `moai factory next`; a follow-up candidate is to mention
  the judged `--card` form in the four locales of `session_start_factory_i18n.go`.
- No party re-reads the decision record (spec §G); a follow-up card may add that step to the sync
  audit if the record is wanted as a control — and may then reconcile the §9.1/§10 wording of
  `auto-semantics.md`, which this SPEC leaves untouched.
- `contract-autonomy.md` L92 says card selection stays the operator's act for a signed contract; a
  follow-up may cross-reference the `--auto` authorization there.
- **Sync phase (`manager-docs`) owns the user-facing pages** that state the operator-only pick
  (§4a, F2): `README.md` L161 and the 4-locale docs-site pages `advanced/factory-mode.md` and
  `advanced/kanban-mode.md`. The run phase leaves them untouched; the sync phase enumerates the
  ko/ja/zh counterparts (not enumerated here) and rewrites each sentence to say the pick is the
  operator's, in person or in advance through `--auto`.

## 8. Risks and their milestones

| Risk | Where handled |
|---|---|
| Doc pins red when the docs change | M5: docs, mirrors, generated artifact and pins in one commit; the two expected breakages observed first |
| `make build` red because the generated Codex artifact is stale | M5 step 2/5: `make agents-emit` in the same commit, `TestGoldenCommittedArtifactsMatchEmission` PASS |
| `internal/template` red because the catalog hashes are stale, or `make build` leaves an unstaged `catalog.yaml` (F1) | M5 steps 2 and 5: the generator run by the lane itself after the template edits are final, `catalog.yaml` staged by pathspec in the same commit, `TestManifestHashFormat` + `TestCatalogHashCoversSkillSubfiles` PASS; both are in the M6 `internal/template` selector |
| Default-path `[보류` skip surprises an operator | isolated clause of REQ-TAU-007; stated in B.4 |
| Lane `--auto` refusal reverses an FLA doctrine move, or refuses a Codex leader | isolated REQ-TAU-008 with an explicit predicate, a non-lane GPT guard, and the corrected fallback string |
| Promoted-then-lost nominee leaves a `picked` card or a record row | M2 steps 1-4, the seam before `RecordPicked`, spec §B.8; the post-record residue is accepted and untested |
| Self-attested, unread decision record | spec §G (unimplemented residual risk), DoD item 6 |
| Parallel edit collision with lane-5's t1453 | surgical replacements; AC-TAU-013's diff bound; this card's rule-doc edits land first |
| Always-loaded growth | AC-TAU-011, bytes and characters, both copies |
| Acceptance counter baseline moves with the criteria | M6: `TestACCounterFullCorpusMatchesBaseline`, regenerate in the same commit if red |

## 9. Commit structure

Plan phase: the plan commit and the two repair commits carry the artifact set. Run phase: tests-first
commit(s) per milestone; M5 is one commit for docs, mirrors, the generated artifact and pins. The
baseline measurements of `acceptance.md` land in the plan commits, **before** any run commit, so the
commit graph — not a message — witnesses that the baselines precede the change
(`verification-claim-integrity.md` §2.3).

## 10. Audit resolution map

### Post-audit re-pin (develop absorption; mechanical, applied after the iteration-4 delta PASS, **no re-audit run**; the leader should be told)

The plan was audited PASS in iteration 4 on audited hash `724841520` (base `4bf547bca`). Develop
`7109e0900` was absorbed afterwards (merge `095ac6c3e`) and changed files this SPEC cites. This entry
re-pins every measurement that went stale; **it changes no requirement and no criterion's meaning**
(REQ stays 16, AC stays 14, the Tier M ceiling is unaffected, the mutants MU-17/MU-29 and the report
duty are kept, `status:` is untouched). Governing rules: `verification-completeness.md` §4 (on
rebase, re-measure and re-pin) and `gitflow-lane-protocol.md` §8 (what this card changed is measured
from the merge base with the absorbed ref, re-derived at reading time, never a literal pinned SHA).

| Item | Resolution (where / how) |
|---|---|
| P1 | AC-TAU-011 conjunct (b) named fixed figures (26,959 / 26,637 B; 26,754 / 26,433 chars) that develop's own edits (+1,349 B per copy) made unsatisfiable by this card; it now reads **relative to the merge-base blob** of each copy (`git merge-base develop HEAD`, then `git cat-file -s` / `git show … \| wc -m`), conjunctive and release-blocking as before, with the current values (28,308 / 27,986 B; 28,099 / 27,778 chars at `b03619b29`) as an informational note. Before the first M5 edit the two sides are equal by construction, so the RED cell stays conjunct (a) (L5). Edited in `acceptance.md` (AC-TAU-011, L11, L22, new L48-L52), `plan.md` §3, `spec.md` REQ-TAU-016, `spec-compact.md` REQ-TAU-016 |
| P2 | Cells that named `4bf547bca` as their reference — L14 and L21 — are re-expressed against `CARD_BASE`: against the old pin they read develop's own files (`internal/kanban` ×2; three docs, `38 8` / `20 0` / `5 1`), so L21 would have been non-empty for the wrong reason. Both read empty, exit 0, against the merge base |
| P3 | Line numbers moved by develop's insertions are given both ways: the live-only `moai worktree sweep …` sentence of `kanban-dispatch.md` is line 181 (177 at plan start; C3 now `differ: char 25120, line 181`, L12 unchanged `1 1`), the `auto-semantics.md` card-pick row is L199 (was L169) and the §9.2 sentence L216 (was L186). Readers locate each by its text; the replaced lines themselves are untouched by develop |
| P4 | Cells re-measured and found unchanged (L5-L10, L13, L19, L20, L23-L25, C2, G6, and `cmp` of the other five AC-TAU-010 pairs) are tabulated in `acceptance.md` after L47 with their exit codes; no new old-authority hit appeared in the files develop changed (sweep re-run, §4a stands) |
| P5 | AC-TAU-013 caps: no cap changed. The regions each cap is derived from are untouched by develop (same text, shifted line numbers where noted); the `catalog.yaml` cap rests on G6 (re-measured PASS, 49 / 37). Not re-run, and said so in `acceptance.md`: the uncommitted scratch drafts and the G8 generator experiment; the committed edit's numstat at M5/M6 is the binding measurement |
| Untouched | `research.md` (its measurements are pinned history of tree `4bf547bca`), the ledger rows that the run's own M1-M4 commits flipped by design, and `progress.md` §E.2 (run-phase owned) |

Files changed since the audited hash `724841520`, and why: `spec.md` (the run phase's sanctioned
`status:` transition, plus this re-pin: `version`, `updated`, HISTORY, REQ-TAU-016, and the REQ-TAU-015
line citation), `plan.md`, `acceptance.md`, `spec-compact.md` (this re-pin only) and `progress.md`
(run-phase evidence, plus the §E.1 note of this re-pin; not a plan-artifact hash member). The
plan-artifact hash of the verdict is therefore no longer equal to the current files; if the Phase 1
skip contract is evaluated again, it needs a fresh audit or the leader's explicit disposition.

### Iteration 3 (audited commit `625f01718`; FAIL 0.80, MP-1..9 PASS, one blocker F1 — corrected here **without a re-audit**; the leader decides the next step)

| Item | Resolution (where / how) |
|---|---|
| F1 | `internal/template/catalog.yaml` is now a named artifact: §5 row, M5 table row, M5 order (seeded probe, `--all` generator after the template edits, step 5 PASS of both catalog guards), the M5 pin-table row, the M6 `internal/template` selector, §8 risk row, AC-TAU-010 Given and RED-now/green cells, AC-TAU-012 changed-file allowance, AC-TAU-013 row with the **measured** cap `3` added / `3` deleted (method: generator run on a scratch copy of the catalog against a tree with the three artifacts perturbed, `git diff --no-index --numstat` = `3  3`), DoD 7, ledger G6-G8. The red was **observed**, not inferred (ledger G7) |
| F2 | §4a's false "no documentation page" row replaced by the real finding (README.md L161; en `advanced/factory-mode.md` L65 and `advanced/kanban-mode.md` L287); classified **sync-phase scope**, listed in §5 and §7, and Out of Scope in spec §D for the run phase |
| S1 | M1: tests assert the literal `4`; the constant is introduced in M2 (`git grep` of `factoryNextRefusedExit` is empty) |
| S2 | REQ-TAU-005 and §B.8 reworded to the plan's "no row, or a row at `picked` with no owner"; the post-record residue is said to re-adopt through arm (c); the `assigned`-to-this-lane row is named as untested by design (M2 step 4) |
| S3 | AC-TAU-010's seeded probe scheduled in M5 step 2 |
| S4 | REQ-TAU-011 no longer says "verbatim" — it points at §B.3 for the sentence; spec-compact.md identical; research R9/R10 aligned to the two-form wording |
| S5 | the two uncaught mutants are listed in the mutant table as **not caught, accepted** with the reason (queue-`picked` marker card; promote-then-restore) |
| S6 | M6 states a baseline regeneration, if ever needed, belongs to a plan commit |
| S7 | spec §G names the serial-slot side effect of the stranded `picked` row; plan M2 repeats it as accepted |
| S8 | ledger row L26 annotated "flips at M1, the criterion at M2" |

### Iteration 2 (audited commit `63daaf6a7`; the leader approved one extra delta audit, scope N1-N6 plus a full ordering re-read)

| Item | Resolution (where / how) |
|---|---|
| N1 | M1 declares the one seam variable (`factoryNominateBeforeRecord`, inert, uncalled) as its only production edit; M1 exit is stated against "the unmodified tree plus the seam declaration" and requires every new test to **compile** and be RED at runtime except the golden and the GPT guard; AC-TAU-006, AC-TAU-014 and DoD item 2 edited to match; the full order check is redone below |
| N2 | REQ-TAU-005 and AC-TAU-014 narrowed to what the record API can undo; spec §B.8 specifies the seam point (before `RecordPicked`), what a post-record failure leaves (a `picked`, unowned row plus its event — accepted, visible in `moai factory status`, untested), the compensation guard (acts only if the item is still `picked`; a failed write exits 1 with `compensation failed`, no token), and the stated non-atomicity window |
| N3 | (a) `manager-todo.toml` is in §5 and M5, regenerated by `make agents-emit`, verified by `TestGoldenCommittedArtifactsMatchEmission` (AC-TAU-010 row, M6 command); cells L23-L24 find the old phrases in the TOML. (b) The `auto-semantics.md` L186 sentence is in the M5 table and C.1, cell L25 finds it, AC-TAU-013's cap is re-measured (45 added / 2 deleted, method in §3). The whole-tree sweep is §4a |
| N4 | One `git grep -c -F "<exact test name>" -- internal/cli` RED-now cell per release-blocking test name (L3, L4 and L27-L47, with control C1), each run for real |
| N5 | AC-TAU-004 gains the mid-text and leading-whitespace fixture cards (`marker-mid-text` leases; `marker-leading-space` refused `hold-marker`), defined against `strings.HasPrefix(strings.TrimSpace(text), "[보류")` |
| N6 | The token set is closed over every producible refusal: twelve tokens, `foreign-worktree` and `recorded` added; `owned` and `recorded` defined against all nineteen record states (spec § C.2); REQ-TAU-009 carries the `queued` qualifier; infrastructure errors exit 1 with no token, stated |
| N7 | spec §G states the §9.1/§10 "compensating control" wording stands for the other gates and is not edited; the new §9.3 sentence `no party re-reads the card-pick record` says it does not hold here; the record's section is `§J` not `§F`; the `--auto` flag help is named in §G; `--wait` with `--card` now waits through `quota-hold` like the bare form; the L19 literal is written without backticks; `TestContractModeEmitterSites` is in the pin table; step (3)/(4) order in M5 fixed; AC-TAU-009's label corrected |
| Leader items | AC baseline snapshot observed first (green) — progress §E.1; the three operator-decision cards are named by one identical handoff sentence in spec §B.3, REQ-TAU-011, §G, DoD 6 and §7; REQ-TAU-013 is intact (§6) |

### Iteration 1 (audited commit `b3646de10`)

| Item | Resolution (where / how) |
|---|---|
| D1 | M5 edits live, mirror and pins in one commit; M6 is measurement only; the milestone↔criterion order table checks all fourteen criteria; AC-TAU-007/-008/-010 green paths restated |
| D2 | M1 exit split: all new tests RED except the golden and the non-lane GPT guard (GREEN on the tree) with seeded-perturbation non-vacuity cells; golden minimum case set in M1 and AC-TAU-006 |
| D3 | `factory_messaging.go` added to § 5 and M3; printed string moves to the lease path; `TestFactoryFallbackDeclarePrintsLeasePath` pins it; REQ-TAU-008 and spec B.4 carry the supersession sentence for REQ-FLA-001 |
| D4 | REQ-TAU-008 defines the predicate (role marker `lane` or non-empty label; Codex marker alone not sufficing); dedicated refusal text; gtd.md routing sentence pinned (C.1); AC-TAU-005 includes a non-lane GPT-backend session measured with a real command (R11 M1) |
| D5 | Option (b): spec B.5 and §G retract the compensating-control claim (unimplemented residual risk); REQ-TAU-012 states the location and "evidence, not the §11 board"; AC-TAU-009 rewritten with no executing party; no audit-surface edit (Out of Scope) |
| D6 | AC-TAU-013 bounds each edited file's diff (floor and ceiling by `git diff --numstat`); § M5 table enumerates every existing marker including `TestAutoRankMirrorParity`'s start marker; the jev/contract/rule-mirror tests were read |
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
