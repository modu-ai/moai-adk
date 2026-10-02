# SPEC-TODO-AUTO-PICK-001 — Acceptance

Tier M. Every release-blocking criterion carries a **RED-now cell** (a read-only single-invocation
command, its verbatim output, its exit code, the pinned tree) and a **green-path cell** (the
milestone that flips it and what the passing output becomes), per
`verification-completeness.md` §2 and §2.1. A criterion that cannot be red on arrival is a
**regression-guard**, is labelled so, and is never recorded as release-blocking. Given-When-Then is
the verification layer's format; the requirements are GEARS in `spec.md`. Where a clause is a model
act that no mechanical check reaches, the criterion says *(doctrine-only)*.

## Evidence ledger (RED-now observations)

Pins. Rows `L1`-`L14` and `C1`-`C3` were observed on tree `4bf547bca`
(`4bf547bcad7c155b1e91485921569db709ec3ac2`) and re-executed by the plan audit on `b3646de10`,
which differs from `4bf547bca` only by the SPEC's own files. Rows `L15`-`L22`, `S1`-`S2`, `C4` were
measured in the iteration-1 repair on `b3646de10`
(`b3646de107ca38c3c0d66e776ec4d8341b0e5ead`), branch `WT-todo-auto-pick-autonomy`. After the run
these rows are history by design — each describes the pinned tree and is expected to print
something else on a tree that carries the linked milestone. Row ids are `L<n>` (RED-now / context),
`C<n>` (controls), `S<n>` (setup).

| Id | Command | Verbatim stdout | Exit | Why it is red |
|---|---|---|---|---|
| L1 | `go run ./cmd/moai factory next --card t1` | (empty stdout); the process error output reads `ERROR`, `Unknown flag: --card.`, `Try --help for usage.`, `exit status 1` | 1 | the nomination flag does not exist |
| L2 | `git grep -c -F "TestFactoryNextNominat" -- internal/cli` | (empty) | 1 | no nomination test exists |
| C1 | `git grep -c -F "TestFactoryNextParallelizableConcurrentLeases" -- internal/cli` | `internal/cli/factory_classify_test.go:2` | 0 | control for L2/L3/L4: the same pathspec and probe hit a test that does exist, so the empty rows are measured absences |
| L3 | `git grep -c -F "TestTodoLaneRefusesAutoCycle" -- internal/cli` | (empty) | 1 | no lane `--auto` refusal test exists |
| L4 | `git grep -c -F "TestAutoPickDocDoctrine" -- internal/cli` | (empty) | 1 | the new doc-contract guard does not exist |
| L5 | `git grep -c -F "consumption of the queue and nothing else" -- .claude/rules/moai/workflow/kanban-dispatch.md .claude/skills/moai/workflows/gtd.md .claude/agents/moai/manager-todo.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/agents/moai/manager-todo.md` | `.claude/agents/moai/manager-todo.md:1` · `.claude/rules/moai/workflow/kanban-dispatch.md:1` · `.claude/skills/moai/workflows/gtd.md:1` · `internal/template/templates/.claude/agents/moai/manager-todo.md:1` · `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md:1` · `internal/template/templates/.claude/skills/moai/workflows/gtd.md:1` | 0 | the old serial-only sentence is present on all six surfaces (a criterion demanding zero hits is red) |
| L6 | `git grep -c -F "The leader never picks for the operator" -- .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | `.claude/rules/moai/workflow/kanban-dispatch.md:1` · `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md:1` | 0 | the unscoped prohibition is still in both copies |
| L7 | `git grep -c -F "consumption in queue order is authorized" -- .claude/skills/moai-kanban-foreman/SKILL.md internal/template/templates/.claude/skills/moai-kanban-foreman/SKILL.md` | `.claude/skills/moai-kanban-foreman/SKILL.md:1` · `internal/template/templates/.claude/skills/moai-kanban-foreman/SKILL.md:1` | 0 | the foreman still authorizes queue-order consumption only |
| L8 | `git grep -c -F "process cards in queue order" -- .claude/agents/moai/manager-todo.md internal/template/templates/.claude/agents/moai/manager-todo.md` | `.claude/agents/moai/manager-todo.md:1` · `internal/template/templates/.claude/agents/moai/manager-todo.md:1` | 0 | the agent mission still says queue order |
| L9 | `git grep -c -F "factory next --card" -- .claude/rules/moai/workflow/kanban-dispatch.md .claude/rules/moai/workflow/auto-semantics.md .claude/skills/moai/workflows/gtd.md .claude/agents/moai/manager-todo.md .claude/skills/moai-kanban-foreman/SKILL.md` | (empty) | 1 | none of the five docs names the nominated form |
| L10 | `git grep -c -F "ladder_path=gate-row card pick" -- .claude/rules/moai/workflow/auto-semantics.md .claude/skills/moai/workflows/gtd.md` | (empty) | 1 | no card-pick decision-record form exists |
| C2 | `git grep -c -F "ladder_path=" -- .claude/rules/moai/workflow/auto-semantics.md` | `.claude/rules/moai/workflow/auto-semantics.md:1` | 0 | control for L10: the sibling literal is live on the same pathspec |
| L11 | `wc -c .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | `   26959 .claude/rules/moai/workflow/kanban-dispatch.md` · `   26637 internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` · `   53596 total` | 0 | baseline byte sizes the non-growth bound is measured against |
| L12 | `git diff --no-index --numstat -- .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | `1	1	{.claude => internal/template/templates/.claude}/rules/moai/workflow/kanban-dispatch.md` | 1 | baseline: exactly one differing line (the pre-existing line-177 drift) |
| C3 | `cmp .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | `.claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md differ: char 23771, line 177` | 1 | positive control: `cmp` detects the one drifting pair |
| L13 | `cmp .claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/skills/moai/workflows/gtd.md` | (empty) | 0 | baseline: identical (parity guard — green by design; AC-TAU-010) |
| L14 | `git diff --name-only 4bf547bca -- internal/kanban internal/graph` | (empty) | 0 | baseline: no change under the two packages REQ-TAU-003 freezes (guard — green by design; AC-TAU-012) |
| L15 | `go run ./cmd/moai factory next --help` | `Lease the lane's next card through the factory record (lane session, parent checkout)` · `USAGE` · `moai factory next [--flags]` · `FLAGS` · `-h --help     Help for next` · `--run         Factory run id (default: the single active run)` · `--wait        Keep re-checking at a fixed interval until a card is leased or the wait bound elapses` · `--wait-bound  How long --wait re-checks before reporting no card (15m0s)` | 0 | the flag set has no `--card`; the baseline the flag-set-equality criterion compares to (AC-TAU-001) |
| L16 | `git grep -n -F "switch to /moai:todo --auto self-service pickup" -- internal/cli/factory_messaging.go` | `internal/cli/factory_messaging.go:255:			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "fallback declared: trigger=%s card=%s at %s\nswitch to /moai:todo --auto self-service pickup (REQ-FLA-001)\n",` | 0 | the shipped fallback verb still routes a lane to the serial cycle (AC-TAU-005) |
| L17 | `git grep -c -F "authorization is exercised through" -- internal/cli/todo.go` | (empty) | 1 | no dedicated `--auto` refusal text exists (AC-TAU-005) |
| S1 | `go build -o <S>/moai ./cmd/moai` (S = the plan session's scratch directory, outside the tree) | (empty) | 0 | setup for L18/C4: this tree's binary, built at `b3646de10` |
| S2 | `MOAI_HOME=<S>/home CLAUDE_PROJECT_DIR=<S>/proj <S>/moai todo add "d4 scratch card one"` | `t1 1` (stderr carries one `config sections directory not found, using defaults` WARN) | 0 | setup: a one-card scratch queue under a redirected `MOAI_HOME`; the database landed at `<S>/home/db/proj-b4a38042/todo/backlog.db`, never in the real home |
| L18 | `env MOAI_FACTORY_WORKER=lane-1 MOAI_HOME=<S>/home CLAUDE_PROJECT_DIR=<S>/proj <S>/moai todo --auto --auto-wait 1ms` (run from a shell with no `MOAI_FACTORY_*` variables; this lane's shell needed an `unset` first, done as a separate precondition) | `jev: unavailable (no local scripts) — labelled non-finding; proceeding on the operator session's own judgment` · `selection: source=fallback reason=jev-disabled` · `selection: ranked t1` · `selection: note landed signal unmeasured for t1 (no answer or unknown)` · `accept t1 d4 scratch card one` · `dispatch (one isolated in-session Agent() worker, isolation: worktree):` · `card: t1` · `evidence: <S>/proj/.moai/reports/t1/evidence.md` · `worker orders: …` · `unpick t1 non-finding: worker evidence absent at deadline … card returned to queued, never done` | 0 | a session whose only lane variable is the lane label **runs** the serial cycle and mutates the queue; the criterion demands a refusal (AC-TAU-005) |
| C4 | `env MOAI_KANBAN_BACKEND=gpt MOAI_HOME=<S>/home CLAUDE_PROJECT_DIR=<S>/proj <S>/moai todo --auto --auto-wait 1ms` (same precondition as L18) | the same ten lines as L18 (`accept t1 …` through `unpick t1 …`) | 0 | **control, must stay green:** a non-lane session carrying only the Codex backend marker runs `--auto` today and must keep running it (AC-TAU-005 non-lane arm); a guard built on `factoryLaneRefusal()` would turn this row red |
| L19 | `git grep -c -F "exercises the --auto authorization through" -- .claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/skills/moai/workflows/gtd.md` | (empty) | 1 | gtd.md has no lane routing sentence (AC-TAU-007) |
| L20 | `git grep -c -F "report-only for an operator-picked card" -- .claude/rules/moai/workflow/kanban-dispatch-detail.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-detail.md` | (empty) | 1 | the detail companion has no operator-picked report-only sentence (AC-TAU-007) |
| L21 | `git diff --numstat 4bf547bca -- .claude/rules/moai/workflow/kanban-dispatch.md .claude/rules/moai/workflow/kanban-dispatch-detail.md .claude/rules/moai/workflow/auto-semantics.md .claude/skills/moai/workflows/gtd.md .claude/agents/moai/manager-todo.md .claude/skills/moai-kanban-foreman/SKILL.md` | (empty) | 0 | the floor of AC-TAU-013 (every named doc has at least one changed line) is not met: nothing is edited yet |
| L22 | `wc -m .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | `   26754 .claude/rules/moai/workflow/kanban-dispatch.md` · `   26433 internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` · `   53187 total` | 0 | baseline character counts (AC-TAU-011) |
| L23 | `git grep -c -F "consumption of the queue and nothing else" -- internal/template/templates/.codex/agents/moai/manager-todo.toml` | `internal/template/templates/.codex/agents/moai/manager-todo.toml:1` | 0 | the **generated** Codex artifact carries the old serial-only sentence (AC-TAU-007, -010) |
| L24 | `git grep -c -F "process cards in queue order" -- internal/template/templates/.codex/agents/moai/manager-todo.toml` | `internal/template/templates/.codex/agents/moai/manager-todo.toml:1` | 0 | and the old queue-order mission sentence |
| L25 | `git grep -c -F "authorizes serial queue consumption and nothing else" -- .claude/rules/moai/workflow/auto-semantics.md internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md` | `.claude/rules/moai/workflow/auto-semantics.md:1` · `internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md:1` | 0 | the second place the old authority is stated (§9.2, line 186), which the L5 literal does not find (AC-TAU-007) |
| L26 | `git grep -c -F "factoryNominateBeforeRecord" -- internal/cli` | (empty) | 1 | the M1 seam does not exist (AC-TAU-014) |
| C5 | `git grep -c -F "factoryCardNow" -- internal/cli/factory_card.go` | `internal/cli/factory_card.go:15` | 0 | control for L26: the seam style the plan names exists, so L26's empty output is a measured absence |
| L27 | `git grep -c -F "TestFactoryNextBareUnchanged" -- internal/cli` | (empty) | 1 | the golden does not exist (AC-TAU-006) |
| L28 | `git grep -c -F "TestFactoryNextAllMarkerQueueExitsNoCard" -- internal/cli` | (empty) | 1 | the all-marker test does not exist (AC-TAU-006) |
| L29 | `git grep -c -F "TestFactoryNextArmCSkipsHoldMarker" -- internal/cli` | (empty) | 1 | the arm-(c) marker test does not exist (AC-TAU-004) |
| L30 | `git grep -c -F "TestFactoryNextNominateLeasesNominee" -- internal/cli` | (empty) | 1 | AC-TAU-001 |
| L31 | `git grep -c -F "TestFactoryNextNominateUnknownCard" -- internal/cli` | (empty) | 1 | AC-TAU-001 |
| L32 | `git grep -c -F "TestFactoryNextFlagSet" -- internal/cli` | (empty) | 1 | AC-TAU-001 |
| L33 | `git grep -c -F "TestFactoryNextNominateMCPParity" -- internal/cli` | (empty) | 1 | AC-TAU-001 |
| L34 | `git grep -c -F "TestFactoryNextNominateConcurrentLanes" -- internal/cli` | (empty) | 1 | AC-TAU-002 |
| L35 | `git grep -c -F "TestFactoryNextNominateSameCardExactlyOne" -- internal/cli` | (empty) | 1 | AC-TAU-002 |
| L36 | `git grep -c -F "TestFactoryNextNominateRefusesKeepSet" -- internal/cli` | (empty) | 1 | AC-TAU-004 |
| L37 | `git grep -c -F "TestFactoryNextNominateQuotaHold" -- internal/cli` | (empty) | 1 | AC-TAU-014 |
| L38 | `git grep -c -F "TestFactoryNextNominateBackendSkip" -- internal/cli` | (empty) | 1 | AC-TAU-014 |
| L39 | `git grep -c -F "TestFactoryNextNominateRecordStateTokens" -- internal/cli` | (empty) | 1 | AC-TAU-014 |
| L40 | `git grep -c -F "TestFactoryNextNominateRefusalLeavesStateUnchanged" -- internal/cli` | (empty) | 1 | AC-TAU-014 |
| L41 | `git grep -c -F "TestFactoryNextNominatePromoteThenLose" -- internal/cli` | (empty) | 1 | AC-TAU-014 |
| L42 | `git grep -c -F "TestFactoryNextNominateClaimRefusedRollsBack" -- internal/cli` | (empty) | 1 | AC-TAU-014 |
| L43 | `git grep -c -F "TestFactoryNextNominateCompensationFailure" -- internal/cli` | (empty) | 1 | AC-TAU-014 |
| L44 | `git grep -c -F "TestTodoLaneAutoRefusalText" -- internal/cli` | (empty) | 1 | AC-TAU-005 |
| L45 | `git grep -c -F "TestTodoNonLaneGPTSessionNotRefused" -- internal/cli` | (empty) | 1 | AC-TAU-005 (a guard: GREEN on the tree once written, see AC-TAU-005) |
| L46 | `git grep -c -F "TestFactoryFallbackDeclarePrintsLeasePath" -- internal/cli` | (empty) | 1 | AC-TAU-005 |
| L47 | `git grep -c -F "TestAutoPickMirrorParity" -- internal/cli` | (empty) | 1 | AC-TAU-007 |

Rows L23-L47 and C5 were measured in the iteration-2 repair on `63daaf6a7`
(`63daaf6a7652624be630996af57ef633e63f7359`); that commit differs from `b3646de10` only by the
iteration-1 repair's SPEC files, so every row's tree is the same code tree as the rows above.
Rows L27-L47 are the **right-selector** cells (iteration-2 N4): L2 greps the substring
`TestFactoryNextNominat` and flips green as soon as any nomination test exists, so it can no longer
stand as the only RED cell of a criterion whose tests have other names; every release-blocking test
name now has its own cell, and C1 is the shared positive control (the same pathspec and probe hit a
test that exists).

**Context rows** (observed, not RED-now cells — their commands are compound or no longer exist):

- **G1.** The baseline of the existing lease pins, lane environment scrubbed in one compound
  invocation: `unset … && go test ./internal/cli -run '^(TestFactoryNextSerialMutualExclusivity|…|TestAutoRankAgentDoctrine)$' -count=1 -v`
  → exit 0, 8 of 8 named tests `--- PASS` (`research.md` R3).
- **G2.** The baseline of the doc pins that read the amended files: `unset … && go test ./internal/cli -run '^(TestAutoRankDoctrineAmendment|TestAutoRankMirrorParity|TestAutoRankMarkerDisclosure|TestAutoRankAgentDoctrine|TestAutoHelpAndRefusalDoNotAssertPickOrder|TestTodoSkillDocumentsClassification)$' -count=1 -v`
  → exit 0, 6 of 6 named tests `--- PASS`, `ok  github.com/modu-ai/moai-adk/internal/cli  2.035s` (`b3646de10`). A wider prefix selector
  over `TestAutoRank`, `TestTodoAuto`, `TestAutoHelp` ran 54 passing tests (`ok … 108.063s`, exit 0).
- **G3.** The baseline of the generated-artifact parity test on the unmodified tree, lane variables
  scrubbed: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/template/agentemit/... -run '^TestGoldenCommittedArtifactsMatchEmission$' -count=1 -v`
  → exit 0, `--- PASS: TestGoldenCommittedArtifactsMatchEmission (0.00s)`, `ok  …/agentemit  0.373s`
  (`63daaf6a7`; the swept set is the one named test, and its `--- PASS` line is printed; an earlier
  unanchored run of the same test printed `0.350s`, exit 0).
- **G4.** The acceptance-counter baseline guard, observed **before** this repair edited
  `acceptance.md`: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/spec -run '^TestACCounterFullCorpusMatchesBaseline$' -count=1 -v`
  → exit 0, `--- PASS: TestACCounterFullCorpusMatchesBaseline (6.84s)` (an earlier unanchored run:
  `ok … 6.230s`; `63daaf6a7`; already green, and **again green after** the repair edited
  `acceptance.md`: this SPEC's `acceptance.md` is listed as `absent-from-snapshot … COUNT 14`, which
  the test reports and does not fail — so **no baseline regeneration was needed**). The
  result after the repair is in `progress.md` §E.1.
- **G5.** The three operator-decision cards, read with the read-only `moai gtd show` through the
  installed binary (attributed to that binary's queue store, not to this tree's build): `t810 live
  picked`, `t1294 live queued`, `t1383 live queued`, none with text that begins with `[보류`
  (exit 0 each). The queue was not touched.
- **P1-P4.** The four throwaway-probe observations (`research.md` R2, O1-O4). The probe file was
  deleted before commit; they motivate AC-TAU-004, -005 and -006 and are **not** re-executable, so
  no criterion's release-blocking status rests on them alone (L18 re-measures the lane arm).

## AC-TAU-001 — A lane takes the card it nominates, and the interface is exactly one flag wider (REQ-TAU-002, -003, -004, -005)

**Covers**: maps REQ-TAU-002, REQ-TAU-003, REQ-TAU-004, REQ-TAU-005

Release-blocking.

**Given** a factory run with lane `lane-1` and three queued parallelizable cards `t1`, `t2`, `t3`
**When** `lane-1` runs `moai factory next --card t2`
**Then** `t2` — not `t1` — is leased to `lane-1`, its queue state is `picked`, its per-card
worktree is ensured exactly as for an unnominated lease, and the output line has the unnominated
shape (`<id> stage=… worktree=…`).

**Given** the same queue
**When** `lane-1` runs `moai factory next --card t9` (an id in no queue)
**Then** the command exits 4, prints one stderr line `factory next: refused unknown-card: …`,
prints nothing on stdout, and the queue and the factory record are byte-identical before and after.

**Given** the command's flag set
**When** `moai factory next --help` is read
**Then** the flags are exactly `--card`, `--run`, `--wait`, `--wait-bound` (and `-h/--help`): the
baseline L15 plus `--card`, nothing else (`TestFactoryNextFlagSet`).

**Given** the MCP tool `factory_next`
**When** it is called with `card` and without it
**Then** with `card` it leases the nominee or returns the same refusal line as an error result;
without `card` it behaves as before; its inputs are exactly `run`, `project_root`, `card`
(`TestFactoryNextNominateMCPParity`).

- **RED-now:** L1 (the flag is unknown), L15 (the flag set has no `--card`) and L30-L33 (each
  named test of this criterion is absent, control C1). Red because the nomination form does not
  exist.
- **Green path:** M2 flips the CLI rows — L1 no longer prints `Unknown flag`, L15 lists `--card`;
  M4 flips the MCP row. The run's `-v` output must list the named tests as `--- PASS` (the swept
  set is checked, not assumed).
- **Mutant probe:** (MU-1) `--card` parsed and ignored (falls through to priority order) leases
  `t1` and fails "leases `t2`"; (MU-2) a lease that bypasses the version-checked edge changes a row
  the refusal case asserts unchanged; (MU-18) a second flag such as `--force-card` fails the
  flag-set equality; (MU-19) an MCP `card` that is parsed and ignored leases `t1` and fails parity.

## AC-TAU-002 — Concurrent lanes never double-pick, through the nominated path (REQ-TAU-002, -006)

**Covers**: maps REQ-TAU-006, REQ-TAU-002

Release-blocking. **Card item 6 is satisfied here:** "two concurrent `--auto` lanes lease
different cards" is exercised through the nominated path, because a lane no longer runs
`moai todo --auto` (AC-TAU-005).

**Given** two registered lanes `lane-1`, `lane-2` and two queued parallelizable cards `t1`, `t2`
**When** the lanes run `factory next --card t1` and `factory next --card t2` concurrently
**Then** each lane holds its own card and no card has two holders.

**Given** the same two lanes and one queued card `t1`
**When** both run `factory next --card t1` concurrently
**Then** exactly one lane holds `t1`; the other exits 4 with the token `raced` or `owned`, and
re-nominating a different candidate (`t2`) succeeds.

- **RED-now:** L34 and L35 (the two named tests do not exist), with control C1.
- **Green path:** M2 — `TestFactoryNextNominateConcurrentLanes` and
  `TestFactoryNextNominateSameCardExactlyOne` pass under goroutine contention (the shape of the
  existing `TestFactoryNextParallelizableConcurrentLeases`); the five pre-existing lease tests of
  G1 stay green. The lane's act of re-selecting is *doctrine-only*; the test performs it.
- **Mutant probe:** (MU-3) a read-then-write nomination without the version check lets both lanes
  hold `t1` and fails the exactly-one assertion.

## AC-TAU-003 — The doctrine lets the session choose, and routes a lane to the lease (REQ-TAU-001)

**Covers**: maps REQ-TAU-001

Release-blocking; the "does not ask" half is *doctrine-only* — it is bounded by the doctrine text
and the decision record, not by a mechanical check.

**Given** the amended docs (live and mirror)
**When** a session reads `kanban-dispatch.md`, the `gtd.md` `--auto` section, `manager-todo.md`
and the foreman skill
**Then** each authorizes the invoked session to take cards on its own judgment, and the `gtd.md`
section states that a lane session exercises the `--auto` authorization through `moai factory
next` (the § C.1 literals), so a lane that reads it has a working path and no reason to ask.

- **RED-now:** L5 and L9 (the docs authorize serial queue-order consumption only and never name the
  nominated form) and L19 (no lane routing sentence). Red because the doctrine that authorizes the
  judgment is absent.
- **Green path:** M5 — L5 reads empty, L9 lists the five files, L19 lists the two `gtd.md` copies.

## AC-TAU-004 — Keep-set cards are never leased, including when nominated (REQ-TAU-007, -009, -010, -011)

**Covers**: maps REQ-TAU-009, REQ-TAU-007, REQ-TAU-011, REQ-TAU-010

Release-blocking.

**Given** a queue holding: `tH` in state `hold`; `tD` in state `dropped`; `tM` queued with text
opening `[보류`; `tB` queued with classification `blocked`; `tS2` a serial card while serial card
`tS1` is in flight; `tO` a card leased by another lane; `tW` queued with text that opens with
whitespace and then the marker (`"   [보류 …"`); `tX` queued with text that merely mentions the
marker after other words (`"…notes [보류 mid-text…"`); and `tOK` an ordinary queued card. The marker
predicate is the one the serial cycle already uses — the trimmed text **begins with** `[보류`
(`strings.HasPrefix(strings.TrimSpace(text), "[보류")`, `todo_auto_rank.go:~150`) — so whitespace
before the marker still opens with it and a mid-text mention does not
**When** a lane runs `moai factory next --card <id>` for each of `tH`, `tD`, `tM`, `tW`, `tB`, `tS2`,
`tO`, and then for `tX` and `tOK`
**Then** `tH`, `tD`, `tM`, `tW`, `tB`, `tS2`, `tO` each exit 4 with their own token (`held`,
`dropped`, `hold-marker`, `hold-marker`, `blocked`, `serial-slot`, `owned`) and change no queue or
record state, while `tX` — the mid-text mention — **is leased** (it is not a marker card), and
`moai factory next --card tOK` leases `tOK`.

**Given** the same queue
**When** a lane runs the **bare** `moai factory next`
**Then** it never leases `tH`, `tD`, `tM`, `tW` or `tB`: the `hold` state, the `dropped` state and the
classification were already skipped, and the leading-marker cards `tM` and `tW` are skipped by
REQ-TAU-007's second clause, while `tX` stays leasable.

**Given** a card marked by neither `hold` nor `[보류` whose text says the work needs a payment
confirmation, and a `queued` card the session read as having an open pull request
**When** the session screens candidates (REQ-TAU-010)
**Then** both are absent from its nominations and present in the record's `skipped=` entry with
their reasons, while an operator-`picked` card with an open pull request is **not** skipped and is
reported only — *doctrine-only*; no mechanical check claims to have caught any of these.

- **RED-now:** L36 (`TestFactoryNextNominateRefusesKeepSet`) and L29
  (`TestFactoryNextArmCSkipsHoldMarker`) — each absent, control C1. Supporting history, not a cell:
  probe O1 — arm (c) leased the `[보류` card (`research.md` R2).
- **Green path:** M2 — `TestFactoryNextNominateRefusesKeepSet` (subtests `held`, `hold-marker`,
  `marker-leading-space`, `marker-mid-text`, `blocked`, `serial-slot`, `dropped`, `owned`) and
  `TestFactoryNextArmCSkipsHoldMarker` pass.
- **Mutant probe:** (MU-4) a keep-set predicate on the bare path only leases `tM` when nominated;
  (MU-5) a predicate on nomination only leaves the bare arm leasing `tM`; (MU-6) a predicate that
  matches the marker anywhere in the text (`strings.Contains`) refuses `tX` and fails the
  `marker-mid-text` subtest; (MU-41) a predicate that does not trim leading whitespace leases `tW`
  and fails `marker-leading-space`; (MU-20) a nomination that leases a `dropped` card fails the
  `dropped` subtest.
- **Isolation:** the arm-(c) half is its own subtest and can be dropped with REQ-TAU-007's second
  clause without touching the nominated-path rows.

## AC-TAU-005 — The lease is the only lane pick path: a lane is refused `moai todo --auto`, a non-lane Codex session is not (REQ-TAU-008)

**Covers**: maps REQ-TAU-008

Release-blocking; isolated (droppable with REQ-TAU-008).

**Given** a queue of two queued cards and a session whose only lane variable is the lane label
(`MOAI_FACTORY_WORKER=lane-1`), then one whose role marker is `lane`, then one with both
**When** the session runs `moai todo --auto --auto-wait 1ms`
**Then** each is refused with the dedicated text — it contains `the --auto authorization is
exercised through` and `moai factory next --card <id>` — exits non-zero, and the queue file is
byte-identical before and after.

**Given** a **non-lane** session whose only marker is `MOAI_KANBAN_BACKEND=gpt` (no role marker, no
lane label), and the same queue
**When** it runs `moai todo --auto --auto-wait 1ms`
**Then** it is **not** refused: the cycle runs exactly as today (row C4), because a Codex-backend
leader session has no lease alternative.

**Given** a session with the lane environment
**When** it runs bare `moai todo` or `moai todo list`
**Then** both still run (the read-only allowlist and the bare parent render are unchanged).

**Given** `moai factory fallback declare --trigger channel-unavailable`
**When** it prints its confirmation
**Then** the instruction line names `moai factory next [--card <id>]` and no longer names
`/moai:todo --auto` (`TestFactoryFallbackDeclarePrintsLeasePath`).

**Given** the amended `gtd.md` `--auto` section
**When** it is read
**Then** it routes a lane to the lease path in the pinned sentence (AC-TAU-007).

- **RED-now:** L18 (a label-only session **runs** the cycle: the output reads `accept t1 d4 scratch
  card one … unpick t1 …`, exit 0), L16 (the shipped fallback string still routes to `/moai:todo
  --auto`), L17 (no dedicated refusal text), and L3, L44, L46 (the refusal, text and string tests are
  absent, control C1; L45 is the guard test and is absent too). Red because the lane refusal, its
  text and the corrected instruction do not exist. **Control C4** must stay green: it
  is the non-lane arm.
- **Green path:** M3 flips L18, L16, L17 — the Go rows do not read the docs; the routing-sentence
  row flips at M5 and is stated against M5. `TestTodoNonLaneGPTSessionNotRefused` is GREEN on the
  unmodified tree (plus the inert seam declaration) and stays green; its non-vacuity is the seeded
  over-broad predicate of M1 (c).
- **Mutant probe:** (MU-7) a guard refusing every bare-parent invocation fails the third Given;
  (MU-8) a guard placed after the cycle starts leaves the queue changed; (MU-21) a refusal that
  reuses the queue-mutation text fails the text assertion; (MU-22) a guard built on
  `factoryLaneRefusal()` refuses the GPT-only session and fails the non-lane Given; (MU-23) an
  unchanged fallback string fails the string test.

## AC-TAU-006 — The bare `factory next` is unchanged, and an all-marker queue ends on the no-card exit (REQ-TAU-007)

**Covers**: maps REQ-TAU-007

Release-blocking (regression guard with its own seeded perturbation).

**Given** a marker-free queue exercising, at minimum: the default arm order (a)→(b)→(b2)→(c) over
assigned / operator-picked / queue-picked / queued cards; `--wait` with a bounded wait; the quota
hold (an assigned card still leasing); the Codex skip of a card at or past merge-ready; the serial
slot (a second serial card not leased while one is in flight); and the no-card exit 3 with its
stdout line
**When** bare `moai factory next` is run by the lanes the golden records
**Then** the leases, their order, the printed lines and the exit codes equal the golden recorded
**before** the change (`TestFactoryNextBareUnchanged`), and the five existing pinned lease tests of
G1 stay green.

**Given** a queue whose only queued cards all open with `[보류`
**When** bare `moai factory next` is run
**Then** it exits 3 with the no-card line and **not** a retry or race exit: the skipped card counts
as seen (`TestFactoryNextAllMarkerQueueExitsNoCard`).

- **RED-now:** L27 (`TestFactoryNextBareUnchanged` does not exist) and L28
  (`TestFactoryNextAllMarkerQueueExitsNoCard` does not exist), each with control C1 — the **right
  selectors**; the earlier L2 greps a different substring and is not cited here. The golden is
  **GREEN on the tree measured at M1** — the unmodified tree plus the one inert seam declaration
  (plan M1) — by design (it pins current behavior), so its red is the **seeded perturbation**
  recorded at M1 (progress §E.2): a one-line mutation of `factory_card.go` — command, verbatim
  stdout, exit code and tree SHA of the failing run, then reverted. The all-marker test is RED on
  that tree (the marker card is leased today).
- **Green path:** M1 writes the golden against the unmodified tree plus the seam declaration
  (GREEN there; the test file compiles because the seam variable exists); M2 keeps it GREEN and
  flips the all-marker test.
- **Mutant probe:** (MU-9) re-sorting candidates in the bare arm, or skipping a card the old arm
  leased, fails the golden; (MU-24) a marker skip that does not count the skipped card turns the
  all-marker queue into a retry and fails the exit-3 assertion; (MU-25) a perturbation of `--wait`,
  the quota hold, the Codex skip or the serial slot fails its golden case.

## AC-TAU-007 — The doctrine sentences are replaced, not appended, live and mirror together (REQ-TAU-001, -002, -003, -011, -014)

**Covers**: maps REQ-TAU-001, REQ-TAU-002, REQ-TAU-003, REQ-TAU-011, REQ-TAU-014

Release-blocking.

**Given** the edited docs (live and mirror) of `spec.md` § C.1
**When** `TestAutoPickDocDoctrine` reads them
**Then** every "must contain" literal is present, every "must no longer contain" literal is absent
from every surface listed for it — live and mirror, in the **same commit** — and the moved
existing pins of plan § M5 (the `kanban-dispatch.md` prohibition sentence and the
`TestAutoRankMirrorParity` start marker) pass in that commit.

- **RED-now:** L5, L6, L7, L8 (old sentences present: each exits 0 with hits), L23 and L24 (the
  same old sentences in the **generated** Codex artifact), L25 (the second old-authority sentence in
  `auto-semantics.md`), L9 (the nominated form absent), L19 and L20 (the lane routing and
  report-only sentences absent), and L4 and L47 (the doc guard tests are absent, control C1). Red
  because the old wording is still in place and the new is not.
- **Green path:** M5 — L5-L8, L23-L25 read empty/exit 1, L9 lists the five files, L19 and L20 list
  their copies; one commit (docs, mirrors, the regenerated TOML and the pins), so no state exists in
  which a live pin is green and its mirror pin red or the TOML is stale.
- **Mutant probe (the card's item d):** (MU-10) a doc edit that **adds** the new wording but
  **leaves** `consumption of the queue and nothing else` fails the L5-form assertion; (MU-11) one
  that leaves `Promotion is the operator's act, always.` or the unscoped `The leader never picks
  for the operator` fails the L6-form assertion; (MU-12) one that edits the live copy only fails
  the mirror comparison (AC-TAU-010); (MU-13) one that rewrites the pinned sentence but not the
  pin fails the existing `TestAutoRankDoctrineAmendment`; (MU-26) one that renames the `gtd.md`
  `--auto` heading fails `TestAutoRankMirrorParity`'s start marker; (MU-27) one that puts a SPEC
  id, ISO date or REQ token in a mirror passage fails the neutrality subtests.

## AC-TAU-008 — The decision record, its location, and the open input set (REQ-TAU-012, -013)

**Covers**: maps REQ-TAU-012, REQ-TAU-013

Release-blocking.

**Given** `auto-semantics.md` §9.3 and the `gtd.md` `--auto` section after the edit
**When** `TestAutoPickDocDoctrine` reads them
**Then** they carry the record form with `ladder_path=gate-row card pick`, name the inputs (card
class, relation records, pull-request/landing state, worktree presence, file overlap with in-flight
lanes, skipped candidates), require `unmeasured` for an input that could not be read, locate the
line in the card's progress record as evidence and not the decision board, state the open-set
sentence `adds an input without amending the keep-set or the lease path`, and state that the
file-overlap fallback is unmeasured and the record does not require it, and (in `auto-semantics.md`
§9.3) state `no party re-reads the card-pick record`; no sentence names the queue-findings store or
`gtd_relations` as *the* relation source.

- **RED-now:** L10 (the `ladder_path=gate-row card pick` literal is absent), with control C2.
- **Green path:** M5 — L10 lists `auto-semantics.md` and `gtd.md`.
- **Mutant probe:** (MU-14) a record form that omits `unmeasured` fails the literal assertion;
  (MU-15) a rule that names `gtd_relations` as the relation source fails the not-store-bound
  assertion; (MU-28) a record described as the §11 board fails the `evidence, not the decision
  board` literal.

## AC-TAU-009 — The record, as the lane writes it (REQ-TAU-012)

**Covers**: maps REQ-TAU-012

**Regression-guard** (not release-blocking): a process criterion over run-phase evidence — no code
writes the line and **no party is assumed to execute a re-read** (spec §B.5, §G).

**Given** the first run-phase lane lease taken under this doctrine
**When** any reader opens the card's progress record
**Then** it holds one `decision record:` line with `decided_by=`, `evidence_refs=` carrying every
named input as `key=value` (`unmeasured` where unread; `skipped=` for passed-over candidates) and
`ladder_path=gate-row card pick (AUTONOMOUS, auto-semantics §9)`, marked self-attested and marked as
evidence, not the board.

- **Disposition:** the criterion's starting observation cannot be re-executed on this tree (no
  lease has been taken under the new doctrine), so per `verification-completeness.md` §2.1 it is a
  regression-guard and is not recorded as a pass until a reader has opened the record.

## AC-TAU-010 — Mirrors agree, and the one pre-existing drift is preserved (REQ-TAU-015)

**Covers**: maps REQ-TAU-015

**Regression-guard** (not release-blocking): it cannot be red on arrival — the pairs are identical
today (L13) and the drift is exactly one line (L12).

**Given** the edited rule/skill/agent files and their `internal/template/templates` mirrors
**When** each pair is compared
**Then** `kanban-dispatch-detail.md`, `auto-semantics.md`, `gtd.md`, `manager-todo.md`,
`moai-kanban-foreman/SKILL.md` and `moai-mcp-tools-catalogue.md` are byte-identical (`cmp` exit 0),
and `kanban-dispatch.md` differs by exactly the one line it differed by before
(`git diff --no-index --numstat` reads `1	1`, and the differing line is still line 177).

**Given** the edited template `manager-todo.md` and the **generated** Codex artifact
`internal/template/templates/.codex/agents/moai/manager-todo.toml`
**When** `AGENTEMIT_UPDATE= go test ./internal/template/agentemit/... -run '^TestGoldenCommittedArtifactsMatchEmission$' -count=1 -v`
is run (lane variables scrubbed in one compound invocation)
**Then** it exits 0 and prints `--- PASS: TestGoldenCommittedArtifactsMatchEmission (0.00s)` — the committed
TOML equals the emission of the edited `.md` (baseline G3: PASS on the unmodified tree; the TOML was
regenerated with `make agents-emit`, never hand-edited). A hand-edited or stale TOML turns this test
red, and `make build` with it (`agents-emit-check`).

- **Contrast control:** C3 shows the same `cmp` returning exit 1 on the one pair that does differ,
  so a `cmp` that cannot fail is ruled out. The generated-parity row is a guard green on arrival
  (G3) and red the moment `manager-todo.md` is edited without regeneration; the seeded probe is to
  perform M5 step 2's edit and run the test **before** `make agents-emit` — the run phase records
  that red once (progress §E.2) and the green after regeneration.
- **Mutant probe:** (MU-12) an edit applied to the live copy only fails `cmp`; (MU-16) an edit that
  syncs line 177 either way changes the numstat to `0 0` or `2 2`.

## AC-TAU-011 — The always-loaded stub does not grow, in bytes or characters (REQ-TAU-016)

**Covers**: maps REQ-TAU-016

Release-blocking; **conjunctive** so it is not vacuous.

**Given** the pinned baselines — bytes live 26,959 / mirror 26,637 (L11), characters live 26,754 /
mirror 26,433 (L22)
**When** both copies of `kanban-dispatch.md` are measured after the edit
**Then** (a) the old sentence `consumption of the queue and nothing else` is absent from both (the
L5 form), **and** (b) `wc -c` of each copy is ≤ its byte baseline **and** `wc -m` of each copy is
≤ its character baseline.

- **RED-now:** L5 (conjunct a is red). Conjunct b is green today by construction (growth 0 ≤ 0);
  the right reason for the red is conjunct a.
- **Green path:** M5 — L5 empty for the two files, and the `wc` commands read ≤ the baselines. The
  reproducible feasibility draft (plan §3) reads −12 B and −16 chars.
- **Mutant probe:** (MU-17) an append-only edit that replaces the sentence but adds paragraphs
  without trimming satisfies (a) and fails (b); (MU-29) an edit that trades multi-byte punctuation
  for ASCII passes the byte bound and fails the character bound.
- **Report duty:** the M5 commit body states the measured before/after bytes and characters of
  both copies and the non-invoking-cost sentence (`rule-authoring.md` (c)).

## AC-TAU-012 — The boundaries hold (REQ-TAU-003)

**Covers**: maps REQ-TAU-003

**Regression-guard** (not release-blocking): green today by design (L14).

**Given** the merge-base `CARD_BASE=$(git merge-base develop HEAD)` taken at read time
**When** the card's diff is listed
**Then** it names no file under `internal/kanban/**` or `internal/graph/**`, no queue/gtd schema
file, and the only non-test `internal/cli` changes are `factory_card.go`, `mcp_factory_card.go`,
`todo.go`, `factory_messaging.go`; the control `git diff --name-only "$CARD_BASE"..HEAD` lists at
least one path (a count of zero is "unmeasurable", not "no change"). The flag set (AC-TAU-001) and
the refusal-token set (spec § C.2) bound the new CLI surface.

- **Non-vacuity:** the control must list ≥ 1 path after the first run commit; a mutant that touches
  `internal/kanban/backlog_store.go` makes the pathspec probe non-empty.

## AC-TAU-013 — The edits are surgical: each file's diff stays inside a measured bound (REQ-TAU-014)

**Covers**: maps REQ-TAU-014

Release-blocking; a floor and a ceiling, so a reflow fails and an empty edit fails.

**Given** the merge-base `CARD_BASE=$(git merge-base develop HEAD)` taken at read time
**When** `git diff --numstat "$CARD_BASE"..HEAD -- <path>` is read for each live file and its mirror
**Then** each file has at least one added line (the floor) and no more than its cap (the ceiling),
and the pre-existing line 177 of `kanban-dispatch.md` is untouched:

| File (each copy) | max added | max deleted |
|---|---|---|
| `kanban-dispatch.md` | 3 | 3 (the three single-line paragraphs L29, L31, L33) |
| `kanban-dispatch-detail.md` | 12 | 0 |
| `auto-semantics.md` | 45 | 2 (the card-pick row L169 and the §9.2 sentence L186; method: a scratch draft of §9.3 is 27 lines, plus one added line for each of the two replaced single-line paragraphs, plus roughly half again for rewording — `plan.md` §3) |
| `gtd.md` | 80 | 24 (the `--auto` clause paragraph, L334-L357, may be re-wrapped) |
| `manager-todo.md` | 12 | 12 (the two regions L22-L25 and L33-L42) |
| `.codex/agents/moai/manager-todo.toml` (template tree only, generated) | 12 | 12 (the same bound as `manager-todo.md`, whose body it emits; checked on the committed diff, produced only by `make agents-emit`) |
| `moai-kanban-foreman/SKILL.md` | 12 | 12 (Boundary 1 L61-L70, step 4) |
| `moai-mcp-tools-catalogue.md` | 1 | 1 (one table row) |

The caps are chosen from the named regions' line counts and the feasibility draft (three changed
lines in `kanban-dispatch.md`); exceeding one needs a recorded reason in the commit body and the
delta re-audit's agreement, not a silent raise. The numstat of the mirror pair must equal the live
file's.

- **RED-now:** L21 — `git diff --numstat 4bf547bca -- <the six live docs>` is empty (exit 0): the
  floor is not met because nothing is edited yet. Red for the right reason.
- **Green path:** M5 — every file's numstat is within [1, cap]; the M6 measurement records it.
- **Mutant probe:** (MU-30) reflowing or reordering the whole of `kanban-dispatch.md` while keeping
  the literals and shedding bytes elsewhere satisfies AC-TAU-007 and AC-TAU-011 and fails the
  ceiling; (MU-31) editing only a mirror fails the floor on the live copy.

## AC-TAU-014 — A refused nominee leaves no write of its own; the promotion is undone only where the stores allow; the quota, Codex and record-state rules bind it (REQ-TAU-004, -005, -009)

**Covers**: maps REQ-TAU-004, REQ-TAU-005, REQ-TAU-009

Release-blocking. What this criterion promises is exactly what the stores can undo (spec §B.8): the
factory record API has no delete, so the criterion asserts the restore and the absence of a row
only for the failure point the single M1 seam reaches — **before `RecordPicked`** — and states what
a later failure leaves.

**Given** a queued nominee and the quota gate holding new leases
**When** a lane runs `moai factory next --card <id>`
**Then** it exits 4 with the token `quota-hold` and changes nothing, while a card already assigned
to the lane still leases (arm (a)); a `--wait` run keeps waiting through `quota-hold` exactly as the
bare form does.

**Given** a Codex lane and a nominee at or past merge-ready
**When** the lane nominates it
**Then** it exits 4 with `backend-skip` and changes nothing.

**Given** nominees whose factory-record rows are in each class of spec § C.2 — a row `assigned` to
another lane or in any of the twelve in-flight states (`owned`), a row in a terminal or parked state
(`recorded`), a card with no recorded worktree whose landing directory belongs to no card
(`foreign-worktree`) — and, as the leasable shapes, a card with no row, a `picked` row with no
owner, and a row `assigned` to this lane
**When** the lane nominates each
**Then** the first three refuse with exit 4 and their own token and the last three lease
(`TestFactoryNextNominateRecordStateTokens`, one subtest per class; every one of the nineteen record
states is exercised or covered by its class).

**Given** any refused nomination decided by the read-only validation (every token of spec § C.2
except `raced`)
**When** the queue and factory record are compared byte-for-byte before and after
**Then** they are identical (`TestFactoryNextNominateRefusalLeavesStateUnchanged`).

**Given** a queued nominee that the invocation promoted to `picked`, and a competing lane that
leases it at the seam — after the promotion, before `RecordPicked`
**When** the claim is lost
**Then** the invocation exits 4 with `raced`, the competing lane holds the card, the queue shows the
competitor's `picked` state, and **no second record row exists**
(`TestFactoryNextNominatePromoteThenLose`).

**Given** a queued nominee that the invocation promoted, and a failure injected through the seam
(its non-nil return) with no other holder and **before any record row exists**
**When** the claim fails
**Then** the invocation exits with a non-token error status (the injected error is not a refusal of
the card), the queue item is back to `queued`, and **the factory record has no row for the card**
(`TestFactoryNextNominateClaimRefusedRollsBack`).

**Given** the same injection, but the seam first moves the queue item out of `picked` (an operator
drop between the promotion and the compensation)
**When** the compensation runs
**Then** it writes nothing, the item stays as the operator left it, and the original failure is
reported (`TestFactoryNextNominateCompensationFailure`, subtest `item-moved`). The branch where the
restoring write itself errors — exit 1 with `factory next: compensation failed: <cause>`, the card
left `picked` and unowned — is **specified in the requirement and not tested** (spec §G).

**What is deliberately not asserted:** a failure *after* `RecordPicked` leaves the card's row at
`picked`, unowned, with its `card.transition` event (the record cannot be rolled back); the queue
item is restored per the compensation rule, the row is re-adopted by an unnominated arm or a later
nomination, and it is visible in `moai factory status` as a `picked` row with no owner. That case
is accepted (it is the state arm (c)'s own failed claim leaves) and untested — the single M1 seam
sits before `RecordPicked`.

- **RED-now:** L26 (the seam variable does not exist, control C5) and L37-L43 (each of the seven
  named tests is absent, control C1) — the **right selectors**. The tests **compile at M1**
  because M1 declares the seam (plan M1) and are RED at runtime for the stated reason (the
  nomination flag does not exist). Supporting history, not a cell: arm (c) itself promotes before
  it claims (`factory_card.go:505-558`), the shape the compensation must not inherit.
- **Green path:** M2 — all seven tests pass; the nominated path validates read-only before the
  first write, calls the seam after the promotion and before `RecordPicked`, and compensates only
  the promotion it made.
- **Mutant probe:** (MU-32) a nomination that promotes before validating leaves a `picked` card
  after a refusal and fails the byte comparison; (MU-33) a compensation that reverts the queue state
  even when another holder owns the card steals the winner's state and fails `PromoteThenLose`;
  (MU-34) a nomination that ignores the quota hold leases and fails the `quota-hold` Given;
  (MU-35) one that ignores the Codex skip leases a merge-ready card and fails `backend-skip`;
  (MU-37) a nomination that leases a `done`/`abandoned` card fails the `recorded` subtest;
  (MU-38) a compensation that restores even when the item is no longer `picked` overwrites the
  operator's change and fails `item-moved`; (MU-42) a precheck that omits the foreign-worktree
  refusal promotes and then fails inside the claim, leaving a `picked` card, and fails the
  `foreign-worktree` subtest.

## Mutant probe summary (the card's item d)

| Mutant | Violates | Caught by |
|---|---|---|
| MU-1, MU-18, MU-19 `--card` ignored; extra flag; MCP `card` ignored | REQ-TAU-003, -004 | AC-TAU-001 |
| MU-2/MU-3 lease without the version-checked edge | REQ-TAU-006 | AC-TAU-001/-002 |
| MU-4/MU-5/MU-6/MU-20 keep-set on one path only; mid-text marker; `dropped` leased | REQ-TAU-009/-007 | AC-TAU-004 |
| MU-7/MU-8/MU-21/MU-22/MU-23 over-broad or late lane guard; wrong text; Codex leader refused; fallback string unchanged | REQ-TAU-008 | AC-TAU-005 |
| MU-9/MU-24/MU-25 default arm perturbed; marker card not counted | REQ-TAU-007 | AC-TAU-006 |
| MU-10/MU-11/MU-13/MU-26/MU-27 old sentence left; pin not moved; heading renamed; internal token in a mirror | REQ-TAU-014 | AC-TAU-007 |
| MU-12/MU-16 one-sided or drift-absorbing edit | REQ-TAU-015 | AC-TAU-010 |
| MU-14/MU-15/MU-28 record without `unmeasured`; store-bound relation; record posing as the board | REQ-TAU-012/-013 | AC-TAU-008 |
| MU-17/MU-29 append-only edit; multi-byte for ASCII trade | REQ-TAU-016 | AC-TAU-011 |
| MU-30/MU-31 whole-file reflow; mirror-only edit | REQ-TAU-014 | AC-TAU-013 |
| MU-32..MU-35 promote before validate; over-eager rollback; quota and Codex ignored | REQ-TAU-004/-005 | AC-TAU-014 |
| MU-37/MU-38/MU-42 `recorded` row leased; compensation overwrites an operator change; foreign-worktree not prechecked | REQ-TAU-005/-009 | AC-TAU-014 |
| MU-39 hand-edited or stale generated Codex artifact | REQ-TAU-014/-015 | AC-TAU-010 (`TestGoldenCommittedArtifactsMatchEmission`), AC-TAU-007 (L23/L24 form) |
| MU-40 the §9.2 L186 sentence left in `auto-semantics.md` | REQ-TAU-014 | AC-TAU-007 (L25 form) |
| MU-41 marker predicate that does not trim leading whitespace | REQ-TAU-009 | AC-TAU-004 |

## Edge cases

- **`--wait` with `--card`:** the nomination is evaluated each pass; `raced`, `serial-slot` and
  `quota-hold` — the three states the bare `--wait` also waits through — keep the wait going, and any
  other refusal ends it immediately (a permanent refusal is not retried); covered by a `--wait`
  subtest of `TestFactoryNextNominateQuotaHold`.
- **Nominee already `picked` by the operator with no owner:** leasable (arms (b)/(b2) semantics);
  `picked` and owned by another lane: `owned`.
- **Empty queue / all candidates in the keep-set:** the session records the empty-candidate state
  and ends its turn with an explicit wait; it does not invent work and does not ask which card.
- **Lane environment in tests:** every fixture builds through `runTodo add`, so the lane variables
  are scrubbed first (`sdClearLaneEnv`); a test that skips this reads a false red.

## Quality gates

- TRUST 5: tests written first and observed RED (M1) except the two guards; coverage ≥ 85% of the
  changed functions; `golangci-lint` at the CI version clean; `GOOS=windows GOARCH=amd64 go build
  ./...` passes.
- Doc gates: `moai spec lint SPEC-TODO-AUTO-PICK-001` and `--strict` clean (plan phase); the
  existing `TestAutoRank*` pins and the `internal/template` mirror tests green after M5.
- MX: the nomination function and the shared predicate are new high-fan-in code — `@MX:NOTE` /
  `@MX:ANCHOR` as the protocol requires (`factory_card.go` is already at its 3-anchor limit; use
  `@MX:NOTE`).

## Definition of Done

1. AC-TAU-001..008, -011, -013, -014 PASS with the RED-now cells observed and the green-path
   outputs recorded in `progress.md` §E.2; AC-TAU-009, -010, -012 hold as regression-guards.
2. The seeded-perturbation cells of M1 (golden, GPT guard) are recorded with command, verbatim
   stdout, exit code and tree SHA, and each mutation is reverted; the M1 tree is the unmodified tree
   plus the one inert seam declaration, every new test **compiles** there, and every new test except
   the golden and the GPT guard is RED at runtime for its stated reason.
3. The commit graph shows the baseline measurements (the plan commits) before any run commit.
4. The always-loaded byte and character measurement and the non-invoking-cost statement are in the
   M5 commit body.
5. The completion report names the preserved line-177 drift, the two deliberate deltas (D-DEF,
   D-LANE) and the unresolved residual risks (spec §G).
6. **Handoff note** (the sentence is identical in spec §B.3, REQ-TAU-011, spec §G, plan §7 and
   here). **Today t810 (`picked`), t1294 and t1383 (`queued`, ordinary text) carry neither the
   structural `hold` state nor a leading `[보류` marker, so the keep-set does not mechanically
   identify them. The SPEC supports exactly two identification forms — (a) the structural `hold`
   state, written only by `moai gtd hold`, and (b) card text that begins with the `[보류` marker —
   and a lane may write neither. The operator or leader must apply one of them to each such card
   before lanes exercise this doctrine.** The completion report carries this note, or states that
   the operator accepted the exposure.
7. `TestGoldenCommittedArtifactsMatchEmission` passes on the M5 commit (the generated Codex artifact
   was regenerated with `make agents-emit`), and `TestACCounterFullCorpusMatchesBaseline` passes —
   the baseline file regenerated in the same commit only if the count moved.
