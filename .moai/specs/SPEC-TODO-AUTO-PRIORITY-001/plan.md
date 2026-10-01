# SPEC-TODO-AUTO-PRIORITY-001 — Implementation Plan

cycle_type: **tdd** (`quality.yaml` `constitution.development_mode: tdd`). Every
behavior below has a RED state observable before the work; the named tests are
authored RED-first in M1/M2 and their verbatim failing output is the E8
deliverable of the run. Plan-phase only: nothing here is committed or run-owned
by manager-spec.

Tier: **M** (spec.md + plan.md + acceptance.md). Core scope = 13 files (§E);
LOC guidance 300-1000. Requirement count 14 / acceptance-criterion count 15,
both within the Tier M ceiling of 16.

## §Findings (read-only investigation, tree HEAD `7d8a9bdbc`, branch `WT-auto-priority-pick`)

Every line number below was read in this run against this tree.

**(a) A priority field really exists; the comment at `todo_edit_move.go:99` is stale.**

- `internal/kanban/classification.go:34-36` — `ClassPriorityHigh|Normal|Low`;
  `:74-81` — `CardClassification{Priority, Blocked, Mode, Decider, ClassifiedAt,
  Reason}`; `:137-142` — `EffectiveCardClassification` derives priority `normal`,
  blocked `false` for an absent field (`DefaultCardClassification` `:87-94`).
- `internal/kanban/backlog_store.go:107-116` — the additive nullable
  `Classification *CardClassification` field on `BacklogItem`.
- `internal/cli/todo_classify.go:31` — shipped decider is `DefaultCardDecider`
  (priority `normal`); `:36` — the fail-safe fallback notice (`priority=normal,
  blocked=false, mode=serial`); `:76-84` — `todoClassifyInLock` records the
  judgment inside the locked add. So a card's recorded priority is `normal`
  unless a judgment file (`--classification-file`) supplied another value.
- The stale text: `internal/cli/todo_edit_move.go:99` ("there are no priority
  fields") and `:5-10` (the header claims the queue is never reordered by inferred
  priority).
- Value-set discipline: `classification.go:11` states the closed sets are
  "validated here, once: no other file re-spells them"; `classPriorities`
  (`:55-59`) is unexported, so the fallback needs an exported rank accessor
  rather than a second spelling in `internal/cli` (M1).

**(b) The queue's sort key is the classification order, applied at add time only.**

- `classification.go:198-209` — `SortByClassification`: non-blocked before
  blocked, then priority `high > normal > low`, `sort.SliceStable` so insertion
  order holds within a rank. Called from the add's locked write
  (`backlog_store.go:966-972`); `classification.go:190-193` states it never runs
  on read.
- `newTodoMoveCmd` (`todo_edit_move.go:106`, doc comment `:99-105`) is documented
  to permute the item slice and nothing else (its `Long` text); no re-sort call is
  made in that file (read, not exercised). A legacy card with no classification
  sorts as `normal`/non-blocked.
- Consequence for this SPEC: stored order already approximates priority, but
  (i) it never excludes a `blocked` card from `--auto`, (ii) it carries no
  readiness signal, and (iii) it is operator-`move`-sensitive. The new ranking
  therefore adds eligibility and demotion; it does not duplicate the sort.

**(c) The pickup predicate and the insertion point.**

- `internal/cli/todo_auto.go:142-189` `autoPickTargets`: arm 1 `:143-161` selects
  `picked` cards whose owner measures dead (`lv.ownerAlive` `:147`); arm 2
  `:162-187` selects `queued` cards, skipping relation-blocked ones through
  `rec.FindingsBlocking` (`:171-184`, labelled non-finding `:179-182`).
- `runAutoCycle` `:204-331` order today: jev display line `:226` → notes
  `:236-238` → no-eligible report `:239-242` → accept loop `:244`.
- Insertion point: a ranking function called in `runAutoCycle` after `:232`,
  applied to the queued suffix of `targets` (arm-1 targets have `State ==
  picked`, arm-2 targets `State == queued`), before `:239`. This leaves the
  `autoPickTargets` signature, the rescue arm, and the relation filter byte-for-byte
  intact, and `todo_auto_test.go:107` / `todo_relation_filter_test.go` keep calling
  it unchanged.

**(d) The Jev surface and the meaning of "available".**

- Display-only script line: `consultJev` `todo_auto.go:339-349` runs
  `scripts/jev/route.sh` (`:340`); a missing script returns "jev: unavailable (no
  local scripts) …" (`:341-343`); invoked at `:226` through the seam `opts.jev`
  (`:198`, `:217-219`). The only consumer of its output is `Fprintln`.
- **Dependency, stated:** `git ls-files scripts/jev` in this tree lists
  `scripts/jev/test_triage.py` and `scripts/jev/triage.py`; `ls scripts/jev/route.sh`
  exits 1 (absent). `triage.sh` and `ask.sh` are likewise untracked here — the
  scripts are local-only tooling (`.moai/docs/jev-local-operations.md:16,39-45`;
  `SPEC-MANAGER-TODO-001/research.md:20`). `route.sh` emits one of five decision
  grades, not a per-card order (`jev-local-operations.md:45`). The SPEC therefore
  takes no dependency on them (spec.md §B.2).
- Shipping capability: `internal/jev/jev.go` — `New(enabled)` `:289`, the
  `Availability` enum `:104-128` (`Available`, `Disabled`, `NoCredential`,
  `Unauthorized`, `RateLimited`, `Overloaded`, `Unreachable`, `Oversize`,
  `SecretDetected`, `Malformed`), `Result` `:231`, `NoticeLine` `:245`, one
  `Request` of one state plus a list of questions `:172-183`, `Answer` with
  `Choice|Score|Probability` `:192-203`. Package imports are standard-library only
  (`display_only_test.go` `TestPackageImports_AreStandardLibraryOnly`).
- Gate: key `workflow.jev.enabled` — `internal/config/types.go:831-833`, code
  default `false` at `internal/config/defaults.go:1180-1182`; this repo's own
  `workflow.yaml:243-244` sets it `true` for dogfood, the template ships `false`;
  resolved by `jevEnabled(root)` (`internal/cli/doctor_jev.go:126`). Credential:
  `~/.moai/.env.typesafe`, read through `jevcred.Load`
  (`workflow.yaml:222-225`).
- Precedent consumer to mirror: `liveJevNearDuplicateProbe`
  (`internal/cli/todo_jev_finding.go:145`, doc comment `:136-144`) — gate check, `jev.New(true)`,
  `client.LoadCredential = jevcred.Load`, `context.WithTimeout`, `res.OK()` else a
  notice and no effect; call bound constant `jevAdmissionTimeout = 5 * time.Second`
  (`:66`); candidate bound `jevNearDuplicateCandidateLimit = 40` (`:72`); a
  timeout is documented to land on `Unreachable` (`:60-65`, read from the comment,
  not exercised in this run).
- Output format of the new signal: typed answers (`Choice`/`Score`/`Probability`),
  never prose. Failure modes: the ten `Availability` values above plus the
  incomplete-answer case the SPEC adds.

**(e) Clauses to amend (exact lines).**

| File (live ↔ template mirror) | Lines | Clause | Disposition |
|---|---|---|---|
| `.claude/rules/moai/workflow/kanban-dispatch.md` ↔ `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | `:29` | `[HARD]` Promotion is the operator's act… "never reorders by inferred priority" | amend: name the `--auto`-scoped ranking exception |
| same | `:31` | the `--auto` reconciliation clause (one reconciliation, "queue order and nothing else") | amend: reference the exception |
| same | `:37` | leader may attach a finding, may not act on one | keep (analyser, not `--auto`) |
| `.claude/skills/moai/workflows/gtd.md` ↔ template (byte-identical today: `diff` exit 0) | `:328-332` | pickup-order sentence ("the queue order") | amend |
| same | `:334-348` | `[HARD]` `--auto` clause: "queue order" `:334-336`, "no ordering logic of its own" `:336-339`, "never self-promotes, reorders…" `:339-341`, "display-only signal for dispatch order and priority" `:345-347` | amend |
| same | `:293-296` | `[HARD]` the pick is the operator's (leader / `gtd next`) | keep |
| same | `:114` | analyser never reorders | keep |
| `.claude/agents/moai/manager-todo.md` ↔ template (byte-identical today) | `:6-8`, `:22`, `:32-34`, `:47-53` | frontmatter description, "strict queue order", serial-cycle contract, Jev Decision Boundary | amend (REQ-TAP-014) |
| `internal/cli/todo_auto.go` | `:17-27`, `:221-225`, `:333-338` | header "never reorders", Jev display-only comments | amend in M3/M4 |
| `internal/cli/todo_edit_move.go` | `:5-10`, `:99` | stale "no priority fields" / "never reordered by inferred priority" | one-line correction in M4 |
| `internal/cli/todo.go` | `:273`, `:311-314` | the refusal string and the `--auto` flag help both say "in queue order" (read in this revision) | reword in M3 so neither asserts the cycle's pick order; the marker disclosure joins the flag help (REQ-TAP-013) |

- **Frozen-clause check:** a search for `ZONE` in `gtd.md` and `manager-todo.md`,
  and for `ZONE:Frozen` in `kanban-dispatch.md`, returns no tag; `zone-registry.md` carries
  no entry naming either file. **No Frozen clause is amended; no constitution
  amend gate is triggered.**
- **Mirror state:** `diff` of the two `kanban-dispatch.md` copies differs at line
  `:177` only (the live copy carries one extra `moai worktree sweep` sentence).
  That drift pre-exists and is NOT touched. The template `gtd.md` is held neutral
  by `TestTodoSkillDocumentsClassification` (regex rejects SPEC ids, requirement
  tokens, ISO dates, 9+ hex runs; `todo_classify_doc_parity_test.go`).
- **Always-loaded cost:** `kanban-dispatch.md` is an always-loaded rule
  (`ls -l`: live 26,352 bytes, template 26,030 bytes); `rule-authoring.md` duty (b)
  requires a byte statement for growth above 1,000 bytes.
- **Other restatements NOT amended here (see §B S-1 and Deferred follow-up):** `SKILL.md:180` pick
  sentence, `agent-authoring.md:147`, `workflow.yaml` (live `:227-229`, template `:229`),
  `internal/jev/jev.go:24`, `internal/cli/mcp_jev.go:8`, `SPEC-MANAGER-TODO-001`
  REQ-MT-014 (`spec.md:69`) and REQ-MT-015.

**(f) Existing tests to characterize before changing.**

- `internal/cli/todo_auto_test.go` (557 lines): `TestTodoAutoPickupSelection`
  `:76` (pins `want = [t4 t9 t3 t8]` `:116`), liveness `:148/:185/:213`,
  `TestTodoAutoSerialCycle` `:246`, clear guidance `:338`, no-lease `:394`,
  empty queue `:439`, Jev tests `:468` (poisoned value, queue unchanged), `:503`
  (degraded non-finding), `:512` (script present), pid probe `:530`, entry flag
  `:542`.
- `internal/cli/todo_relation_filter_test.go` (346 lines):
  `TestAutoPickTargetsRelationBlocked`, `…ReturnsAfterDone`,
  `TestRunAutoCycleSkipsBlockedCards`, `…RescueArmUnfiltered`,
  `…NonSequencingRelation`.
- Doc/mirror guards: `todo_classify_doc_parity_test.go`, `todo_hold_doc_test.go`,
  `todo_landed_doc_test.go`, `internal/template/gtd_canonical_surface_test.go`,
  `internal/template/sanitized_pair_parity_test.go` `TestSanitizedPairParity`
  (`:161`), `internal/contract/kickoff/activation_test.go` `TestJevAmendmentLinkage`
  (`:208`; requires the token `contract-mode Kickoff` to stay in both
  `workflow.yaml` copies).
- **Baseline observed at `7d8a9bdbc`:** `go test -count=1 -v -run
  'TestTodoAuto|TestAutoPickTargets|TestRunAutoCycle' ./internal/cli/` under a
  partial env scrub (the 5 `MOAI_KANBAN*` variables only) printed PASS for the 13
  `TestTodoAuto[A-Z]*` cycle tests and the 5 relation-filter tests and FAIL for 21
  `TestTodoAutoDone_*` / `TestTodoAutoDoneSkipsHeldCard` tests with `refused —
  lane boundary: a lane session cannot mutate the queue`. Re-running one of them
  (`TestTodoAutoDoneSkipsHeldCard`) with all eleven variables scrubbed — adding
  `MOAI_FACTORY_WORKER`, `MOAI_FACTORY_ROLE`, `MOAI_FACTORY_WORKERS`,
  `MOAI_FACTORY_CLEAR_POLICY`, `MOAI_FACTORY_AUTO_DISPATCH`, `MOAI_KANBAN_BACKEND`
  — printed PASS. The other 20 `TestTodoAutoDone_*` failures were not individually
  re-run (Gap G-2). Lesson for every command in this SPEC: scrub all eleven.

## §A Context

Card t1400: `--auto` has no ranking step. The operator's four decisions are
binding (spec.md §B.1). This plan realizes them with one new ranking stage, one
Jev-ordering consumer behind the existing default-off gate, a printed decision
record, and a scoped doctrine amendment. Nothing in the queue's storage changes.

## §B Decisions and open items

**Design decisions (D):**

- **D-1 — Ranking stage lives in a new file.** `internal/cli/todo_auto_rank.go`
  (plus its test file) holds the stage; `todo_auto.go` gains one call site and the
  comment amendments. Keeps the characterized file's diff small. **Pinned:** the
  `selection:` record is rendered by a function in `todo_auto_rank.go` and that
  file is the only non-test source that writes the `selection:` literals;
  `todo_auto.go` calls the renderer and holds none. The RED-now probe E1
  (acceptance.md) is scoped to non-test sources under `internal/cli` so that it
  flips on this file and cannot be flipped by a test file alone.
- **D-2 — Rank only the queued suffix.** Rescue-arm targets keep their position at
  the head (REQ-TAP-009) by construction, independent of the ranking source.
- **D-3 — Jev request shape.** One request, one state, a list of questions
  (`internal/jev.Request` already forces this, `:172-183`): state = the bounded
  candidate list rendered by code (id, text prefix, recorded priority, and the
  three computed readiness signals); one typed score question (`KindScore`) per
  candidate keyed by the card id, with five named levels (a named level-count
  constant) — valid scores are the finite numbers 0 to 4, zero-based, a higher
  score meaning the card is more ready to be picked now (A-6); order key = answer
  `Score` descending, then `Probability` (the answer's confidence) descending,
  then fallback order (spec.md REQ-TAP-002). Validation (REQ-TAP-011): an answer
  naming an id that was not sent, a sent id with no answer, or a score outside the
  range degrades the WHOLE result to `jev-incomplete-answer` — never a partial
  application. Question wording and the level texts are authored in M2 against
  `moai-ref-jev-question-design` (compute in code, always admit a no-match — level
  0 is the lowest, not-ready level —, small state, expressible both polarities);
  the level count and the order direction are fixed here, the wording is not.
- **D-4 — Bounds as named constants.** Candidate bound (precedent 40) and call
  timeout are named constants, not inline literals (hardcoding-prevention); the
  timeout runs outside any queue lock because the cycle ranks over an in-memory
  `LoadPure` record. Surplus beyond the bound follows in fallback order and the
  record names it (REQ-TAP-002).
- **D-5 — External inputs are seams; the other two readiness checks are pure.**
  The landed check wraps the existing `computeTodoPRRows` computation
  (`todo_pr.go:214`, one `gh` query, local-git landed check, fail-open `unknown`);
  the near-duplicate check reads `rec.FindingsNaming` for
  `kanban.BacklogRelationNearDuplicate` (`backlog_store.go:139`) and the marker
  check is a trimmed-prefix test against a named constant — both pure over the
  loaded record, so neither is a seam. The two external inputs — the landed lookup
  and the Jev ranker — are new fields of `autoOptions`; tests inject both.
  **Nil-seam default (inert):** `runAutoCycle` does NOT default a nil seam to its
  live implementation (unlike `opts.jev` today, `todo_auto.go:217-219`). A nil Jev
  ranker means Jev unavailable with reason `jev-disabled` — no credential read, no
  request, no network; a nil landed lookup means the landed signal is unmeasured
  for every card — one `selection: note` names it, no `gh` or git subprocess runs.
  Production wiring sets both live seams where the `--auto` flag builds its
  `autoOptions` (`todo.go:275-278`). The 10 existing characterization tests build
  `autoOptions` without the new fields, so they run on the inert default: the
  only visible difference is extra `selection:` lines before the first `accept`
  (Gap G-7).
- **D-6 — Output order.** jev display line (unchanged, `:226`) → existing notes
  (`:236-238`) → selection record → accept loop. The record is skipped only when
  the candidate set is empty.

**Assumptions (A) — correct before kickoff if wrong:**

- **A-1** Exclusion of `blocked` cards applies on the Jev source as well as the
  fallback (operator decision 1 lists it under the fallback only). Reason: a
  `blocked` card is an eligibility fact, matching the factory lease predicate
  (`SPEC-TODO-CLASSIFY-DISPATCH-001` REQ-TCD-007); Jev orders the eligible set.
- **A-2** "A near-duplicate finding on the card" means a live finding of relation
  `near-duplicate` that names the card on either side, from any source. Symmetric
  demotion needs no orientation assumption and keeps the relative order of the
  pair; a stricter "newer card only" reading is a one-line change.
- **A-3** The `[보류` test runs on the trimmed card text (leading whitespace
  tolerated, the marker must open the text).
- **A-4** Readiness signals are computed once at cycle start, so `--auto` makes at
  most one `gh` query per invocation, as `moai todo pr` does.
- **A-5** "Jev available" is the Go capability (`internal/jev`, gate
  `workflow.jev.enabled`), not the local scripts. Operator decision 1's fallback
  condition "scripts absent" is read as covered by that definition: the scripts
  are untracked and absent from this tree and from the template (§Findings (d)),
  and emit a decision grade with no per-card order, so they are not an ordering
  source at all. This is an interpretation, not an operator statement. Correct
  before kickoff if the operator meant `scripts/jev/route.sh` as the primary
  source: spec.md §B.2, REQ-TAP-002/-003 and AC-TAP-002/-003 change, and the
  ranking acquires a dependency on an untracked file (§G, first anti-pattern).
- **A-6** The score scale is five named levels with zero-based indices (valid
  scores 0 to 4, a higher score meaning a card Jev judges more ready to be picked
  now). The package documents a score question's `Levels` as "ordered … lowest
  first" and the answer `Legend` as keyed by "a score's level index"
  (`internal/jev/jev.go:162-164`, `:201-202`); whether that index is zero-based
  was not observed against the live capability (credential-gated, Gap G-6).
  Correct before kickoff if the base is one: the REQ-TAP-011 range becomes 1 to 5
  and AC-TAP-010(iii) changes with it; the run keeps both bounds as named
  constants in one place.

**Settled decisions (operator, 2026-10-02 — they replace the two open items of
the first draft; nothing in this section is open):**

- **S-1 — Doctrine scope.** Amend ONLY `kanban-dispatch.md` (`:29`, `:31`),
  `workflows/gtd.md` (`:328-348`) and `manager-todo.md`, live files and template
  mirrors where they exist. The Jev-side surfaces are not touched by this card.
  Wording constraint on the amended clauses: scope the exception to selection
  order only. Each amended passage carries both literals, `auto-scoped ranking
  exception` and `selection order only`, in one paragraph (spec.md §B.5,
  REQ-TAP-012, REQ-TAP-014), because until the follow-up lands the Jev-side
  documents still say "display-only".
- **S-2 — Jev ordering accuracy.** The Jev ordering consumer ships only behind
  the default-off gate (`workflow.jev.enabled`, code default `false`,
  `internal/config/defaults.go:1180-1182`). Enabling it is the operator's act
  under `SPEC-JEV-OPTIN-MEASURE-001` REQ-JEVO-009. This SPEC's requirements and
  acceptance criteria claim no ordering accuracy (spec.md §D). That no labelled
  set exists for an ordering question is the plan-time statement of the first
  draft; it was not re-measured in this revision.

**Deferred follow-up — a card the leader issues; NOT part of this card.** It
carries the linked amendment of the surfaces that still state the Jev
display-only principle, listed exactly:

- `internal/jev/jev.go:24` — the package comment ("The capability is
  display-only").
- `.moai/config/sections/workflow.yaml` (live `:227-229`) and
  `internal/template/templates/.moai/config/sections/workflow.yaml` (`:229`) —
  the `workflow.jev` comment. The token `contract-mode Kickoff` must stay in both
  copies (`TestJevAmendmentLinkage`, `internal/contract/kickoff/activation_test.go:208`).
- `internal/cli/mcp_jev.go:8` — the tool's doc comment.
- The MCP tools catalogue `.claude/rules/moai/core/moai-mcp-tools-catalogue.md`
  § Judgment (`:135`, `:139`) and its template mirror.
- `SPEC-JEV-CORE-001`.
- `SPEC-MANAGER-TODO-001` REQ-MT-014 and REQ-MT-015.

Observed in §Findings (e) and outside the operator's list, equally untouched
here: the `SKILL.md:180` pick sentence and `agent-authoring.md:147`; the leader
decides whether they join the follow-up card.

Interim state until it lands: those surfaces still say "display-only" (spec.md
§B.5, §G R-6).

## §C Pre-flight (the run inherits)

Every Go command in this SPEC runs with all eleven lane variables scrubbed, in one
compound invocation (a separate `unset` does not carry):

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_WORKER MOAI_FACTORY_ROLE MOAI_FACTORY_WORKERS MOAI_FACTORY_CLEAR_POLICY MOAI_FACTORY_AUTO_DISPATCH MOAI_KANBAN_BACKEND && go test -count=1 -v -run '^<TestName>$' ./internal/cli/
```

All `-run` patterns are anchored (`^…$`): an unanchored pattern also selects longer
names, which is a vacuous-pass shape (the spec lint rule `VacuousTestAssertion`
flags it).

1. `git rev-parse --short HEAD` and `git branch --show-current` — expect the card
   branch; re-read before any commit.
2. Baseline: the §Findings (f) command; record the pass/fail set before editing.
3. `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` (no
   OS-specific call is planned, so no build-tag split is needed; the Windows build
   still runs, B1).
4. Test selector sweep control: `go test -list 'TestAutoRank' ./internal/cli/`.
   Observed at `7d8a9bdbc`: prints only `ok  	github.com/modu-ai/moai-adk/internal/cli	0.859s`
   — zero names, exit 0. A green `-run 'TestAutoRank'` with an empty swept set is
   therefore a vacuous pass; the swept count is read from `-list` first.

## §D Constraints

- Template-First: edit under `internal/template/templates/` first, `make build`,
  then bring the live copy to the same wording. The template copies carry no SPEC
  id, requirement token, ISO date, or 9+ hex run (the mirror guard enforces it);
  citations such as `SPEC-TODO-AUTO-PRIORITY-001` live only in spec artifacts and
  commit messages.
- `kanban-dispatch.md` is always-loaded: keep the amendment short; any single
  edit that grows it by more than 1,000 bytes needs the byte/cost statement in the
  commit body (`rule-authoring.md` duty (b)/(c)).
- Keep the token `contract-mode Kickoff` intact in both `workflow.yaml` copies if
  they are ever touched (`TestJevAmendmentLinkage`). This plan does not touch
  them, nor any other surface listed under §B Deferred follow-up.
- Preserve: `autoPickTargets` signature and behavior, the rescue arm, the relation
  filter, hold exclusion, the jev display line at `todo_auto.go:226`, the
  `internal/jev` stdlib-only import set, every existing test in §Findings (f).
- No queue writes from the ranking stage; no `Mutate`, `ArchiveCard`, or queue
  verb in `todo_auto_rank.go`. `AskUserQuestion` stays orchestrator-only (CLI code
  never calls it).
- No time estimates; milestones are priority-ordered.
- manager-develop owns `draft → in-progress`; this agent writes no `progress.md`.

## §E Milestones (ordered by decision-reversibility — likeliest-to-change first)

Files changed (core scope, 13): `internal/kanban/classification.go`,
`internal/kanban/classification_test.go`, `internal/cli/todo_auto_rank.go` (new),
`internal/cli/todo_auto_rank_test.go` (new), `internal/cli/todo_auto.go`,
`internal/cli/todo.go` (flag help), `internal/cli/todo_auto_doc_test.go` (new),
`.claude/rules/moai/workflow/kanban-dispatch.md` and its template mirror,
`.claude/skills/moai/workflows/gtd.md` and its template mirror,
`.claude/agents/moai/manager-todo.md` and its template mirror. Comment-only
touches in `todo_edit_move.go` count toward the same set at M4 (14th file; still
Tier M). Regenerated embedded template output from `make build` is a build
artifact, not a hand edit.

Files the run changed beyond that list (reconciled after the run phase; four
files, each required by a committed guard or by an audit finding, none a scope
addition):

- `internal/cli/doctor_jev_test.go` — consumer declaration for the `internal/jev`
  import guard. `TestJevCallPath_HasExactlyTheDeclaredConsumers` was red on the M1
  tree at the start of M2 because `todo_auto_rank.go` imports `internal/jev`; the
  guard's own comment names its allow-list as the way a new consumer arrives, so
  the fix is one declared consumer.
- `internal/cli/todo_auto_test.go` — the audit finding D-N1 change: the hermetic
  seam replacement and call-count assertions in `TestTodoAutoEntryPointFlag`, plus
  one `internal/jev` import. Without it the `--auto` flag path reached the real
  `gh` once the live seams were wired. The ten §Findings (f) characterization
  tests in that file are unchanged.
- `internal/template/catalog.yaml` — hash lines regenerated by `make build` after
  the template edits (the `moai` skill directory at M3, the `manager-todo` entry
  at M4); a build artifact, not a hand edit.
- `internal/template/templates/.codex/agents/moai/manager-todo.toml` — generated
  from the amended agent body by `make agents-emit`; the committed-TOML drift
  guard requires it whenever an agent body changes.

With these four, the card changed 22 files in total (`git diff --name-only
7d8a9bdbc..HEAD`), counting the four SPEC artifacts.

**M1 — Selection contract: record format, ranking types, fallback keys (Priority High).**
Decisions most likely to change: the output tokens and the closed reason
vocabulary. Add `PriorityRank(string) int` to `kanban/classification.go` (one
spelling of the value set) with the table test `TestPriorityRank`
(`classification_test.go`). Add `todo_auto_rank.go`: the
candidate type, the landed-lookup seam and the pure readiness checks, the fallback
comparator (clean before poor, priority, stable queue order), blocked exclusion,
the selection-record renderer (the only place the `selection:` literals are
written, D-1), and the `jev-*` reason mapping from `jev.Availability`. M1 is
stage-level: its tests call the stage function and the renderer directly with
injected seams, and nothing is wired into `runAutoCycle` yet. Green at M1:
`TestAutoRankFallbackOrder`, `TestAutoRankDemotion`,
`TestAutoRankUnmeasuredSignal`, the `kanban` rank-accessor test
`TestPriorityRank`, and the fallback-source subtests of
`TestAutoRankBlockedExcluded`. Gate: those green, §Findings (f) tests still green.
The criteria whose Then-clauses read the cycle's output (AC-TAP-001, -002, -003,
-007, -008, -009, -010) are NOT flipped at M1; they flip at M2, where the wiring
lands.

**M2 — Wire the stage and the Jev ordering consumer (Priority High).**
Call the ranking from `runAutoCycle` after `autoPickTargets` (order per D-6) and
set the live seams where `--auto` builds its `autoOptions` (`todo.go:275-278`,
D-5; `runAutoCycle` itself leaves a nil seam inert). Build the Jev consumer
(D-3/D-4) behind `jevEnabled`, mirroring `liveJevNearDuplicateProbe`; validate
answers (REQ-TAP-011) and degrade the whole result on any defect. Every
cycle-level criterion flips here. Green at M2: `TestAutoRankSelectionRecord`,
`TestAutoRankFallbackReasons`, `TestAutoRankJevOrdering`,
`TestAutoRankJevMalformedAnswer` (the name is kept from the sweep set; it
exercises `jev-incomplete-answer`, not `jev-malformed`), `TestAutoRankRescueFirst`,
`TestAutoRankQueueUnchanged`, `TestAutoRankNoQueueWriteGuard`, and the full
`TestAutoRankBlockedExcluded` (Jev-source subtests added). Gate: M1 tests still
green; §Findings (f) tests green on the inert nil-seam default (D-5); no network
call in tests (stubbed seams).

**M3 — Doctrine amendment and disclosure (Priority Medium).**
Amend `kanban-dispatch.md` `:29`/`:31` and `gtd.md` `:328-348` (template first),
using the two pinned literals `auto-scoped ranking exception` and
`selection order only`, in the same paragraph, in both documents (spec.md §B.5);
extend the `--auto` flag help (`todo.go:311-314`) and the `gtd.md` section with
the marker-limitation disclosure (REQ-TAP-013), and reword the "in queue order"
phrasing of the refusal string (`todo.go:273`) and of the flag help so neither
asserts the cycle's pick order. Add `todo_auto_doc_test.go` asserting the two
literals and their same-paragraph co-occurrence, the disclosure, live↔mirror
agreement of the amended passage, and the unchanged prohibitions. Gate: `TestTodoSkillDocumentsClassification`,
`TestSanitizedPairParity`, the new doc test.

**M4 — Agent text, stale comments, build and parity sweep (Priority Low, mechanical).**
Amend `manager-todo.md` ×2 (REQ-TAP-014; both literals in one paragraph of the
serial-cycle contract and of the Jev Decision Boundary); correct the comment passages named in
§Findings (e) (`todo_auto.go:17-27`, `:221-225`, `:333-338`; `todo_edit_move.go`
`:5-10`, `:99`); run `make build`; run the verification set below.

## §F Verification commands (all with the §C scrub prefix where they run Go)

```bash
go test -list 'TestAutoRank' ./internal/cli/
go test -count=1 -v -run '^(TestAutoRankSelectionRecord|TestAutoRankJevOrdering|TestAutoRankFallbackReasons|TestAutoRankFallbackOrder|TestAutoRankDemotion|TestAutoRankUnmeasuredSignal|TestAutoRankBlockedExcluded|TestAutoRankRescueFirst|TestAutoRankQueueUnchanged|TestAutoRankNoQueueWriteGuard|TestAutoRankJevMalformedAnswer|TestAutoRankDoctrineAmendment|TestAutoRankMirrorParity|TestAutoRankMarkerDisclosure|TestAutoRankAgentDoctrine)$' ./internal/cli/
go test -count=1 -v -run '^(TestTodoAutoPickupSelection|TestAutoPickTargetsRelationBlocked|TestAutoPickTargetsReturnsAfterDone|TestRunAutoCycleSkipsBlockedCards|TestAutoPickTargetsRescueArmUnfiltered|TestAutoPickTargetsNonSequencingRelation|TestTodoAutoSerialCycle|TestTodoAutoJevPoisonedValueCausesNoMutation|TestTodoAutoJevDegradedNonFinding|TestTodoAutoJevScriptPresentSignal)$' ./internal/cli/
go test -count=1 -run '^(TestTodoSkillDocumentsClassification|TestTodoHoldDocumentedOnEverySurface)$' ./internal/cli/
go test -count=1 -run '^TestSanitizedPairParity$' ./internal/template/
go test -count=1 -run '^TestJevAmendmentLinkage$' ./internal/contract/kickoff/
go test -count=1 -v -run '^TestPriorityRank$' ./internal/kanban/
go vet ./internal/cli/ ./internal/kanban/
```

Scope discipline: the full `internal/cli` suite is not run locally (CI owns it);
the sets above are the change-scoped selection.

## §G Anti-patterns (named for the run to refuse)

- Pinning the Jev branch on `route.sh` or any file under `scripts/jev/` — untracked
  and absent; the ranking source is the Go capability only.
- Applying demotion on the Jev source, or letting a Jev answer touch a card's
  state, text, position, or any queue verb.
- Re-spelling `high|normal|low` in `internal/cli` instead of using the exported
  rank accessor.
- A zero-hit `-run` selector read as a pass (§C item 4).
- Writing a SPEC id, requirement token, date, or hash into either template mirror.
- Removing the `[보류` test's trimmed-prefix semantics so that a mid-text mention
  demotes a card.

## §H Cross-references

`SPEC-MANAGER-TODO-001` (the `--auto` cycle, REQ-MT-012/014/015),
`SPEC-RELATION-PICKUP-FILTER-001` (the relation filter this SPEC leaves intact),
`SPEC-TODO-CLASSIFY-DISPATCH-001` (priority/blocked data model and sort),
`SPEC-TODO-HOLD-STATE-001` (hold state; positive state enumeration),
`SPEC-JEV-CORE-001` and `SPEC-JEV-OPTIN-MEASURE-001` (capability, gate, shipping
gate), `.claude/rules/moai/workflow/kanban-dispatch.md` § Entry into the board is
an operator act, `.claude/skills/moai/workflows/gtd.md` § `--auto`,
`.claude/rules/moai/development/verification-completeness.md` §1-§2.

## §I Gaps (unobserved in this run)

- **G-1** The live queue's contents were not read; the motivating cards (a
  `[보류`-prefixed card at the front, a lead-signal-blocked card) are taken from
  the card text, not observed.
- **G-2** 20 of the 21 `TestTodoAutoDone_*` failures seen under the partial scrub
  were not individually re-run under the full scrub; one was (PASS).
- **G-3** The judging build for `moai spec audit` / lint on these artifacts is the
  running MCP server (build `v3.2.0-rc.23`, commit `d194083fb`), which is older
  than HEAD `7d8a9bdbc`; a lint rule added after that commit is not exercised.
- **G-4** The `Unreachable`-on-timeout mapping and the Jev request wire shape were
  read from source comments and types, not exercised.
- **G-5** No mutant probe was executed at plan time; the mutant lines in
  `acceptance.md` are the run's M1/M2 gate, not completed observations.
- **G-6** The score level-index base (A-6) and the live capability's answer shape
  for a score question were not observed: the capability is credential-gated and
  no request was sent in this revision.
- **G-7** The full bodies of the 10 existing characterization tests were not read
  in this revision; the nil-seam default (D-5) rests on a grep of their assertion
  forms (`accept ` prefix scans, `done `/`unpick ` counts, substring contains),
  not on a run. The confirming observation is AC-TAP-012's command at M2.
- **G-8** The revised acceptance ledger rows (E1, E1c, E9-E11) were measured in
  this revision at HEAD `7d8a9bdbc`, and E3-E6 were re-run to the same output.
  `moai spec lint` ran on the installed build `v3.2.0-rc.23` (commit
  `d194083fb`, `moai version`), which was not rebuilt from this tree; a lint
  rule added after that commit is not exercised (as in G-3).
