# SPEC-TODO-AUTO-PICK-001 — Acceptance

Tier M. Every release-blocking criterion carries a **RED-now cell** (a read-only single-invocation
command, its verbatim output, its exit code, the pinned tree) and a **green-path cell** (the
milestone that flips it and what the passing output becomes), per
`verification-completeness.md` §2 and §2.1. Given-When-Then is the verification layer's format;
the requirements are GEARS in `spec.md`.

## Evidence ledger (RED-now observations)

Whole-ledger pin: tree `4bf547bca` (`4bf547bcad7c155b1e91485921569db709ec3ac2`), branch
`WT-todo-auto-pick-autonomy`, observed 2026-10-02 in this card's worktree. After the run these rows
are history by design — each describes the pinned tree and is expected to print something else on a
tree that carries the linked milestone. Row ids are `L<n>` (RED-now / context) and `C<n>` (controls).

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
| L11 | `wc -c .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | `   26959 .claude/rules/moai/workflow/kanban-dispatch.md` · `   26637 internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` · `   53596 total` | 0 | baseline sizes the non-growth bound is measured against (see AC-TAU-011) |
| L12 | `git diff --no-index --numstat -- .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | `1	1	{.claude => internal/template/templates/.claude}/rules/moai/workflow/kanban-dispatch.md` | 1 | baseline: exactly one differing line (the pre-existing line-177 drift) |
| C3 | `cmp .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | `.claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md differ: char 23771, line 177` | 1 | positive control: `cmp` detects the one drifting pair |
| L13 | `cmp .claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/skills/moai/workflows/gtd.md` | (empty) | 0 | baseline: identical (parity guard — green by design; see AC-TAU-010) |
| L14 | `git diff --name-only 4bf547bca -- internal/kanban internal/graph` | (empty) | 0 | baseline: no change under the two packages REQ-TAU-003 freezes (guard — green by design) |

**Context rows** (observed, not RED-now cells — their commands are compound or no longer exist):

- **G1.** The baseline of the existing pinned tests, lane environment scrubbed in one compound
  invocation: `unset … && go test ./internal/cli -run '^(TestFactoryNextSerialMutualExclusivity|…|TestAutoRankAgentDoctrine)$' -count=1 -v`
  → exit 0, 8 of 8 named tests `--- PASS` (full command and list: `research.md` R3). The swept set is
  non-empty (eight named PASS lines), so the green is not an empty sweep.
- **P1-P4.** The four throwaway-probe observations (`research.md` R2, O1-O4). The probe file was
  deleted before commit; they are history and motivate AC-TAU-004, -005, -008, but are **not**
  re-executable, so no criterion's release-blocking status rests on them alone.

## AC-TAU-001 — A lane takes the card it nominates (REQ-TAU-002, -004, -005)

**Covers**: maps REQ-TAU-002, REQ-TAU-004, REQ-TAU-005

Release-blocking.

**Given** a factory run with lane `lane-1`, and a queue of three queued parallelizable cards
`t1`, `t2`, `t3`
**When** `lane-1` runs `moai factory next --card t2`
**Then** `t2` — not `t1` — is leased to `lane-1` in the factory record, its queue state is
`picked`, its per-card worktree is ensured exactly as for an unnominated lease, and the output
line has the unnominated shape (`<id> stage=… worktree=…`).

**Given** the same queue
**When** `lane-1` runs `moai factory next --card t9` (an id that is in no queue)
**Then** the command exits with the refusal code (distinct from 0, 1 and the no-card status 3),
prints one stderr line naming the token `unknown-card`, prints nothing on stdout, and neither the
queue file nor the factory record changes (byte-identical before and after).

- **RED-now:** L1 (the flag is unknown) and L2 (no nomination test exists). Red because the
  nomination form does not exist.
- **Green path:** M2 flips L1 — the command no longer prints `Unknown flag`; M1's
  `TestFactoryNextNominateLeasesNominee` and the `unknown-card` subtest pass. The swept set is
  checked: the run's `-v` output must list at least the named tests as `--- PASS`.
- **Mutant probe:** (MU-1) a `--card` that is parsed and ignored (falls through to priority order)
  leases `t1` and fails the "leases `t2`" assertion; (MU-2) a nomination that leases without the
  version-checked edge changes a row the refusal case asserts unchanged.

## AC-TAU-002 — Concurrent lanes never double-pick, through the nominated path (REQ-TAU-006, -002)

**Covers**: maps REQ-TAU-006, REQ-TAU-002

Release-blocking.

**Given** two registered lanes `lane-1` and `lane-2` and two queued parallelizable cards `t1`, `t2`
**When** the two lanes run `factory next --card t1` and `factory next --card t2` concurrently
**Then** each lane holds its own card and no card has two holders.

**Given** the same two lanes and one queued card `t1`
**When** both run `factory next --card t1` concurrently
**Then** exactly one lane holds `t1`; the other exits with the refusal code and the token `raced`
or `owned`, and re-nominating a different candidate (`t2`) succeeds.

- **RED-now:** L2 (the swept tests do not exist), with control C1.
- **Green path:** M2 — `TestFactoryNextNominateConcurrentLanes` and
  `TestFactoryNextNominateSameCardExactlyOne` pass under goroutine contention (the shape of the
  existing `TestFactoryNextParallelizableConcurrentLeases`); the five pre-existing lease tests of
  G1 stay green.
- **Mutant probe:** (MU-3) a nomination implemented as read-then-write without the version check
  lets both lanes hold `t1` and fails the exactly-one assertion.

## AC-TAU-003 — Lane `--auto` self-service never asks which card to take (REQ-TAU-001, -003)

**Covers**: maps REQ-TAU-001, REQ-TAU-003

Release-blocking (doctrine + behavior pair).

**Given** a lane session under an `--auto` authorization and a queue with ranked candidates
**When** the lane's rules are read and the lane takes its next card
**Then** the rule text (AC-TAU-007's literals) authorizes the lane to choose on its own judgment
through `moai factory next [--card <id>]`, and the lane's decision record (AC-TAU-009) — not an
operator question — carries the choice. No `moai gtd add/drop/edit/done/relate` runs in a lane
(the existing refusals still hold: `TestTodoLaneRefusesAutoCycle` and the existing lane-refusal
tests green).

- **RED-now:** L5 and L9 — the docs authorize serial queue-order consumption only and never name
  the nominated form. Red because the doctrine that authorizes the judgment is absent.
- **Green path:** M5 (docs) and M3 (refusal) — L5 reads empty, L9 lists the five files.
- **Note:** the "does not ask" half is a model behavior; it is bounded by the doctrine text and the
  decision record, not by a mechanical check (spec §G).

## AC-TAU-004 — Keep-set cards are never leased, including when nominated (REQ-TAU-009, -007, -011)

**Covers**: maps REQ-TAU-009, REQ-TAU-007, REQ-TAU-011, REQ-TAU-010

Release-blocking.

**Given** a queue holding: `tH` in state `hold`; `tM` queued with text opening `[보류`; `tB`
queued with classification `blocked`; `tS2` a serial card while serial card `tS1` is in flight; and
`tOK` an ordinary queued card
**When** a lane runs `moai factory next --card <id>` for each of `tH`, `tM`, `tB`, `tS2`
**Then** each exits with the refusal code and its own token (`held`, `hold-marker`, `blocked`,
`serial-slot`), changes no queue or record state, and `moai factory next --card tOK` then leases
`tOK`.

**Given** the same queue
**When** a lane runs the **bare** `moai factory next`
**Then** it never leases `tH`, `tM` or `tB`: the `hold` state and the classification were already
skipped, and the `[보류` marker is skipped by REQ-TAU-007 (the single default-path delta).

**Given** a queue holding a card marked by neither `hold` nor `[보류` whose text says the work
needs a payment confirmation
**When** the lane applies REQ-TAU-010
**Then** the card is absent from the nominations and present in the record's `skipped=` entry with
its reason; no mechanical check claims to have caught it.

- **RED-now:** L2 (tests absent) with C1. Supporting history, not a cell: probe O1 — arm (c)
  leased the `[보류` card (`research.md` R2).
- **Green path:** M2 — `TestFactoryNextNominateRefusesKeepSet` (subtests `hold-state`,
  `hold-marker`, `blocked`, `serial-slot`) and `TestFactoryNextArmCSkipsHoldMarker` pass.
- **Mutant probe:** (MU-4) a keep-set predicate applied to the bare path only, not to nomination,
  leases `tM` when nominated and fails the `hold-marker` subtest; (MU-5) a predicate on nomination
  only leaves the bare arm leasing `tM` and fails `TestFactoryNextArmCSkipsHoldMarker`; (MU-6) a
  predicate that tests the marker anywhere in the text (not only at the start) refuses a card that
  merely mentions `[보류` mid-text and fails the "ordinary card still leases" assertion.
- **Isolation:** the arm-(c) half (REQ-TAU-007) is its own subtest and can be dropped with the
  requirement without touching the nominated-path rows.

## AC-TAU-005 — The lease is the only lane pick path: `moai todo --auto` is refused in a lane (REQ-TAU-008)

**Covers**: maps REQ-TAU-008

Release-blocking; isolated (droppable with REQ-TAU-008).

**Given** a queue of two queued cards and the lane environment (`MOAI_FACTORY_ROLE=lane`,
`MOAI_FACTORY_WORKER=lane-1`)
**When** the lane runs `moai todo --auto --auto-wait 1ms`
**Then** the command is refused with the lane-boundary refusal text (it contains `lane boundary`
and `moai factory next`), exits non-zero, and the queue file is byte-identical before and after
(`sdQueueBytes`).

**Given** the same environment
**When** the lane runs bare `moai todo` (no flags) or `moai todo list`
**Then** both still run (the read-only allowlist and the bare parent render are unchanged).

- **RED-now:** L3 (test absent) with C1. Supporting history, not a cell: probe O4 — the cycle ran
  from a lane and printed `accept t1 …` / `unpick t1 …`.
- **Green path:** M3 — `TestTodoLaneRefusesAutoCycle` passes; the existing lane-refusal tests
  (REQ-SD-015 family) stay green.
- **Mutant probe:** (MU-7) a guard that refuses every bare-parent invocation breaks `moai todo`
  with no args and fails the second Given; (MU-8) a guard keyed on the flag but placed after the
  cycle starts leaves the queue changed and fails the byte-identity assertion.

## AC-TAU-006 — The bare `factory next` is unchanged (REQ-TAU-007)

**Covers**: maps REQ-TAU-007

Release-blocking (regression guard with its own RED).

**Given** a marker-free queue with ranked, classified, serial and parallelizable cards and
quota/Codex conditions as the pinned tests build them
**When** bare `moai factory next` is run repeatedly by several lanes
**Then** the leases, their order, the printed lines, and the exit codes equal the golden recorded
before the change (`TestFactoryNextBareUnchanged`), and the five existing pinned tests of G1 stay
green.

- **RED-now:** L2 (the golden test does not exist). The green-now existing pins (G1) are the
  regression baseline; this criterion's own RED is the missing golden, so a mutant that perturbs
  the default path has nothing to fail today.
- **Green path:** M1 writes the golden against the **unmodified** tree (it must be GREEN there —
  it pins current behavior), M2 keeps it green.
- **Mutant probe:** (MU-9) re-sorting candidates in the bare arm, or skipping a card the old arm
  leased, fails the golden.

## AC-TAU-007 — The doctrine sentences are replaced, not appended (REQ-TAU-001, -002, -003, -011, -014)

**Covers**: maps REQ-TAU-001, REQ-TAU-002, REQ-TAU-003, REQ-TAU-011, REQ-TAU-014

Release-blocking.

**Given** the five edited docs (live and mirror) of `spec.md` § C.1
**When** `TestAutoPickDocDoctrine` reads them
**Then** every "must contain" literal is present, every "must no longer contain" literal is absent
from every surface listed for it, and the existing pins moved in the same commit still pass.

- **RED-now:** L5, L6, L7, L8 (old sentences present: every row exits 0 with a hit) and L9 (new
  literal absent). Red because the old wording is still in place and the new is not.
- **Green path:** M5 — L5-L8 read empty/exit 1, L9 lists the five files.
- **Mutant probe (the card's item d):** (MU-10) a doc edit that **adds** the new wording but
  **leaves** `consumption of the queue and nothing else` fails the L5-form assertion; (MU-11) one
  that leaves `Promotion is the operator's act, always.` or the unscoped `The leader never picks
  for the operator` fails the L6-form assertion; (MU-12) one that edits the live copy only fails the
  mirror-parity assertion (AC-TAU-010); (MU-13) one that rewrites the pinned sentence in
  `kanban-dispatch.md` but not the pin in `todo_auto_doc_test.go` fails the existing
  `TestAutoRankDoctrineAmendment` — which is why REQ-TAU-014 moves them together.

## AC-TAU-008 — The decision record and the open input set (REQ-TAU-012, -013)

**Covers**: maps REQ-TAU-012, REQ-TAU-013

Release-blocking.

**Given** `auto-semantics.md` §9.3 and the `gtd.md` `--auto` section after the edit
**When** `TestAutoPickDocDoctrine` reads them
**Then** they contain the record form with `ladder_path=gate-row card pick`, name the inputs
(card class, relation records, pull-request/landing state, worktree presence, file overlap with
in-flight lanes, skipped candidates), require `unmeasured` for an input that could not be read,
state the file-overlap fallback, and state the open-set sentence
`adds an input without amending the keep-set or the lease path`; and no sentence names the
queue-findings store or `gtd_relations` as *the* relation source.

- **RED-now:** L10 (the `ladder_path=gate-row card pick` literal is absent) with control C2.
- **Green path:** M5 — L10 lists `auto-semantics.md`.
- **Mutant probe:** (MU-14) a record form that omits `unmeasured` (so an unread input reads as
  absent-and-fine) fails the `unmeasured` literal assertion; (MU-15) a rule that names
  `gtd_relations` as the relation source fails the not-store-bound assertion.

## AC-TAU-009 — The record, as the lane writes it (REQ-TAU-012)

**Covers**: maps REQ-TAU-012

Regression-guard on the run-phase evidence (process criterion; not a RED-now item — the record is
authored by a lane, not by code).

**Given** the first run-phase lane lease taken under this doctrine
**When** the sync audit re-reads the card's progress record
**Then** one `decision record:` line carries `decided_by=`, `evidence_refs=` with every named
input (each `key=value`, `unmeasured` where unread, `skipped=` for passed-over candidates) and
`ladder_path=gate-row card pick (AUTONOMOUS, auto-semantics §9)`, and the line is explicitly marked
self-attested (no board writer exists — `research.md` R4-iii).

- **Disposition:** the criterion's starting observation cannot be re-executed on this tree (no
  lease has been taken under the new doctrine), so per `verification-completeness.md` §2.1 it is a
  **regression-guard** and is not recorded as a pass until the sync audit reads the line.

## AC-TAU-010 — Mirrors agree, and the one pre-existing drift is preserved (REQ-TAU-015)

**Covers**: maps REQ-TAU-015

Release-blocking (guard with contrast control).

**Given** the edited rule/skill/agent files and their `internal/template/templates` mirrors
**When** each pair is compared
**Then** `kanban-dispatch-detail.md`, `auto-semantics.md`, `gtd.md`, `manager-todo.md`,
`moai-kanban-foreman/SKILL.md` and `moai-mcp-tools-catalogue.md` are byte-identical (`cmp` exit 0),
and `kanban-dispatch.md` differs by exactly the one line it differed by before
(`git diff --no-index --numstat` reads `1	1`, and the differing line is still line 177).

- **RED-now:** none by design — L13 (`cmp` exit 0) and L12 (`1 1`) are green today; this criterion
  guards against a one-sided or drift-absorbing edit. **Contrast control:** C3 shows the same `cmp`
  returning exit 1 on the one pair that does differ, so a `cmp` that cannot fail is ruled out; the
  mutants below are the non-vacuity argument.
- **Green path:** M6 — every `cmp` exit 0 except the one pair, `numstat` still `1 1`.
- **Mutant probe:** (MU-12) an edit applied to the live copy only fails `cmp`; (MU-16) an edit that
  syncs line 177 either way changes the numstat to `0 0` or `2 2` and fails the preserved-drift
  assertion.

## AC-TAU-011 — The always-loaded stub does not grow (REQ-TAU-016)

**Covers**: maps REQ-TAU-016

Release-blocking; **conjunctive** so it is not vacuous.

**Given** the pinned baselines (live 26,959 B, mirror 26,637 B — L11)
**When** both copies of `kanban-dispatch.md` are measured after the edit
**Then** (a) the old sentence `consumption of the queue and nothing else` is absent from both
(the L5 form), **and** (b) `wc -c` of each copy is ≤ its baseline.

- **RED-now:** L5 (conjunct a is red: the old sentence is present on both). Conjunct b is green
  today by construction (growth 0 ≤ 0); the right reason for the red is conjunct a.
- **Green path:** M5 + M6 — L5 empty for the two files and `wc -c` ≤ 26,959 / 26,637.
- **Mutant probe:** (MU-17) an append-only edit that satisfies (a) by replacing the sentence but
  adds new paragraphs without removing others satisfies (a) and fails (b) — the case the
  feasibility draft (+11 B before trimming) shows is reachable if the author does not trim.
- **Report duty:** the commit body states the measured before/after bytes of both copies and the
  non-invoking-cost sentence (`rule-authoring.md` (c)).

## AC-TAU-012 — The boundaries hold (REQ-TAU-003)

**Covers**: maps REQ-TAU-003

Release-blocking (guard; green by design).

**Given** the merge-base `CARD_BASE=$(git merge-base develop HEAD)` taken at read time
**When** the card's diff is listed
**Then** it names no file under `internal/kanban/**` or `internal/graph/**`, no queue/gtd schema
file, and the only `internal/cli` non-test changes are `factory_card.go`, `mcp_factory_card.go`,
`todo.go`; the control `git diff --name-only "$CARD_BASE"..HEAD` lists at least one path (a
count of zero would be "unmeasurable", not "no change").

- **RED-now:** none by design (L14 is empty today); the criterion is a guard. Non-vacuity: the
  control row above must list ≥ 1 path after the first run commit, and a mutant that touches
  `internal/kanban/backlog_store.go` makes the pathspec probe non-empty.
- **Green path:** M7 — the pathspec probe stays empty and the control is non-empty.

## Mutant probe summary (the card's item d)

| Mutant | Violates | Caught by |
|---|---|---|
| MU-1 `--card` parsed and ignored | REQ-TAU-004 | AC-TAU-001 |
| MU-2/MU-3 lease without the version-checked edge | REQ-TAU-006 | AC-TAU-001/-002 |
| MU-4/MU-5 keep-set on one path only | REQ-TAU-009/-007 | AC-TAU-004 |
| MU-6 marker matched mid-text | REQ-TAU-009 | AC-TAU-004 |
| MU-7/MU-8 over-broad or late lane guard | REQ-TAU-008 | AC-TAU-005 |
| MU-9 default arm perturbed | REQ-TAU-007 | AC-TAU-006 |
| MU-10/MU-11 old sentence left in place | REQ-TAU-014 | AC-TAU-007 |
| MU-12/MU-16 one-sided or drift-absorbing edit | REQ-TAU-015 | AC-TAU-010 |
| MU-13 pin not moved with the docs | REQ-TAU-014 | existing `TestAutoRankDoctrineAmendment` |
| MU-14/MU-15 record without `unmeasured`; store-bound relation source | REQ-TAU-012/-013 | AC-TAU-008 |
| MU-17 append-only edit | REQ-TAU-016 | AC-TAU-011 |

## Edge cases

- **Quota hold:** while the quota gate holds new leases, `--card <id>` is refused with
  `quota-hold` (exit via the existing hold line semantics) and the lane's already-assigned card
  still leases (arm (a) unchanged).
- **Codex lane:** a nominee at or past `merge-ready` is refused for a Codex lane exactly as the
  unnominated arms skip it.
- **Nominee already `picked` by the operator with no owner:** leasable (arms (b)/(b2) semantics);
  `picked` and owned by another lane: `owned`.
- **`--wait` with `--card`:** the nomination is evaluated each pass; a refusal other than
  `raced`/`serial-slot` ends the wait immediately (a permanent refusal is not retried).
- **MCP form:** `factory_next` with `card` returns the same refusal line as an error result;
  without `card` it behaves as before (`TestFactoryNextNominateMCPParity`).
- **Empty queue / all candidates in the keep-set:** the session records the empty-candidate state
  and ends its turn with an explicit wait; it does not invent work and does not ask which card.
- **Lane environment in tests:** every fixture builds through `runTodo add`, so the lane variables
  are scrubbed first (`sdClearLaneEnv`); a test that skips this reads a false red.

## Quality gates

- TRUST 5: tests written first and observed RED (M1); coverage ≥ 85% of the changed functions;
  `golangci-lint` at the CI version clean; `GOOS=windows GOARCH=amd64 go build ./...` passes.
- Doc gates: `moai spec lint SPEC-TODO-AUTO-PICK-001` clean (plan phase); the existing
  `TestAutoRank*` pins and `internal/template` mirror tests green after M5/M6.
- MX: the nomination function and the shared predicate are new exported-or-high-fan-in code —
  `@MX:NOTE` / `@MX:ANCHOR` as the protocol requires (`factory_card.go` is already at its 3-anchor
  limit; use `@MX:NOTE`).

## Definition of Done

1. AC-TAU-001..008, -010, -011, -012 PASS with the RED-now cells observed and the green-path
   outputs recorded in `progress.md` §E.2.
2. AC-TAU-009 read by the sync audit from the first lane lease taken under the new doctrine.
3. The commit graph shows the baseline measurements (this plan commit) before any run commit.
4. The always-loaded byte measurement and the non-invoking-cost statement are in the commit body.
5. The completion report names the preserved line-177 drift, the two deliberate deltas (D-DEF,
   D-LANE) and the unmarked operator-decision cards, each as a gap or follow-up.
