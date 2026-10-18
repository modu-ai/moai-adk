# Design: SPEC-TODO-UNIFY-001

The mechanisms behind the requirements of `spec.md`. This is the HOW layer; identifiers and file names appear here and nowhere in the requirements. Tree basis `026f48658f`.

## §D.1 Current-state map

| Concern | Where it lives today | Fact |
|---------|---------------------|------|
| `moai todo` tree | `internal/cli/todo.go` `newTodoCmd` (registers 26 verbs; one pre-run hook that runs the lane guard) | the only place queue verbs are registered |
| `moai gtd` | `internal/cli/gtd.go` `NewGTDCommand`: takes `newTodoCmd()`, renames it, adds six stage-verb constructors | no handler of its own for the queue verbs |
| Stage verbs | `gtd.go` (`newGTD{Capture,Clarify,Organize,Reflect,Engage}Cmd`), `gtd_answer.go` (`newGTDAnswerCmd`) | call `factory.CaptureGTDItem`, `ClarifyGTDItem`, `OrganizeGTDItemWithRelations`, `ReflectGTDStore`, `EngageGTDItem`; `goal.go` calls two of those library functions directly |
| Lane guard | `todo.go` `todoRefuseLaneMutation`, allowlist `todoLaneReadOnlyVerbs` (7 verbs) | refuses every other verb for a lane session, parent-with-arguments included |
| Close gate | `internal/cli/leader_approval_gate.go` `holdDoneApprovalGate` | wraps every close of a factory-linked card; verifies the receipt under a factory write transaction |
| Auto-done scan | `internal/cli/todo_autodone.go`, `internal/factory/autodone_scan.go` | closed evidence set of two forms; three misfire guards; log `auto-done-log.jsonl`; source label `auto-land` |
| Transition table | `internal/homestate/card_transition.go` (`transitionTable`, 71 accepted pairs pinned at `fr_transition_test.go:166`) | T18 row at `:157`; reserved-edge branch at `:333` |
| Push gate | `internal/homestate/card_evidence.go` `PushGateTarget`; caller `internal/cli/factory_card.go` | returns `pushed` with a remote, `done` without one |
| spec_id carriers | queue item `SpecID` (set by `todo next <n> --spec`); factory card `spec_id` (set by `factory assign --spec` while `picked`; refusal at `card_picked.go:185`) | no post-pick path |
| Slash and skills | `.claude/commands/moai/{gtd,todo}.md`, `.agents/skills/moai-{gtd,todo}/SKILL.md` (generated), `internal/template/gtd_canonical_surface_test.go`, `published_skills.go` | `todo.md` dispatches `gtd` and a test pins it |

## §D.2 The warning alias (REQ-TU-001 to REQ-TU-003)

**Shape.** `NewGTDCommand` keeps building from `newTodoCmd()`. After M4 it adds nothing else (the stage verbs are either on the todo tree or gone), so `gtd` is the todo tree under another name, which is what REQ-TU-001 states.

**The warning.** One pre-run function, chained in front of the tree's existing pre-run hook (the lane guard), writes the warning once per process. The hook fires once per `moai ...` invocation for the parent and every subcommand alike, so it is the single place that satisfies "exactly one"; it does not fire for a help request, because cobra renders help without running pre-run, which satisfies the no-warning-on-help half of REQ-TU-003. The text goes to the command's stderr, never stdout, and follows the existing deprecation convention (`Warning:` prefix on stderr; precedent `warnDeprecatedAgentModelFlags` in `agent_model_flags_retired.go`). It names the `moai todo` spelling of what was run (the verb token is taken from the executing command) and the removal release.

**The removal release** is one exported-within-package constant. AC-TU-002 asserts the warning text contains it and that it is not empty, so an unset constant cannot ship. OQ-2 decides its value.

**Why not hide `gtd`.** The existing help-parity test requires `gtd --help` and `todo --help` bodies to be equal modulo the name; hiding the command from the root listing changes neither, but it also removes the discoverability the grace period exists to give. Not done; revisitable.

**Why the lane guard still sees the right name.** `todoSurfaceName` already reports the spelling the operator typed, so the guard's refusal text stays truthful under both names.

**Dependency-initialisation.** `gtd` is not in `trivialCommands` and must not be added: the queue verbs read dependencies, and a warning alias that skipped initialisation would diverge from the `todo` form.

## §D.3 The substitution mechanism (REQ-TU-009 to REQ-TU-011)

Hand-editing 73 files and 255 lines is the failure mode the card names. The mechanism is one script, `scripts/gtd-to-todo-substitute.sh`, POSIX `sh` plus `sed` without `-i` (write to a temporary file and move it, so it behaves the same on macOS and Linux), driven by a class table and an allowlist file.

**Class table.**

| Class | Token forms | Action |
|-------|-------------|--------|
| S1 prose command | `moai gtd <word>` where `<word>` is a queue verb or end of token | `moai todo <word>` |
| S2 slash | `/moai gtd`, `/moai:gtd` | `/moai todo`, `/moai:todo` |
| S3 path | `workflows/gtd.md`; docs-site page path `moai-gtd` and its menu and `_meta.yaml` rows | renamed path; page rows merged into the `moai-todo` entry with a redirect row per locale |
| S4 stage verbs | `moai gtd capture\|clarify\|organize\|reflect\|engage\|answer` | the script takes the OQ-1 disposition as `--absorbed <list>`; absorbed verbs follow S1; any line naming a deleted verb is listed for hand editing and is never rewritten |
| X1 reviewed edits | argument vectors (`'moai', 'gtd'`), environment variable names (`CUTOVER_GTD_LIST_CMD`), in-tree error and help strings | not touched by the script; edited in M7 with their own tests |
| E1 history | `.moai/specs/**`, `.moai/reports/**`, `CHANGELOG.md` | excluded |
| E2 binary | any path `git grep -I` treats as binary | excluded |
| E3 recorded fixtures | `internal/cli/testdata/codex-rollouts-t1171/**`, `internal/template/testdata/binding_ledger.json` | excluded |
| E4 internals | Go identifiers `GTD*`, `gtd_*` names, `gtd-<hex>` ids, `.moai/gtd/`, the `internal/factory/gtd_*.go` and `backlog_gtd_schema.go` contents | excluded (OQ-3 default) |
| E5 alias statements | lines listed in the allowlist file | excluded; they are the hand-authored deprecation statements |

**The input set** is `git ls-files` intersected with the S1 to S3 patterns minus E1 to E5, never a filesystem walk, so untracked scratch cannot be touched.

**Dry-run contract.** `--dry-run` prints one line per file (`would-change <path> <count>`) and a total, changes nothing, and exits 0. The set it prints must equal the pre-flight set (E-01) minus the excluded classes; M8 compares the two sets and stops on any difference.

**Idempotence (REQ-TU-010).** After `--apply`, a second run prints a total of 0 and `--apply` changes no file (`git status --short` empty for the touched paths).

**Residual ledger (REQ-TU-011).** `TestNoResidualGTDSpelling` walks the tracked guidance corpus (S1 to S3 scope, same exclusions), reads the allowlist file the script reads, and fails on any `moai gtd` or `/moai gtd` outside it. Two safeguards keep it honest: it asserts the swept file count is at least a floor measured at M7 (an empty sweep is not a pass), and a fixture containing a planted `moai gtd add` line proves the test can go red. It runs in the ordinary suite, so a later edit that reintroduces the spelling is caught by CI without anyone asking.

**Docs-site obligations in the same batch.** Four locales change together; a redirect row for the retired `moai-gtd` page in each locale; the menu row and the `_meta.yaml` row removed or merged; the locale-parity baseline is not extended; the docs-site verification recipe (warning-free build, parity, URL and emoji scans) runs before the batch is accepted.

## §D.4 The six-verb disposition (REQ-TU-004 to REQ-TU-006)

Step 0 of M4 measures, for each verb: code callers (`git grep` of the constructor and of the library function), documentation mentions (skill, rules, READMEs, docs-site), hook or script callers, and tests. The output is one table in `.moai/reports/t1641/verb-inventory.md`:

| Verb | Library function | CLI callers | Doc mentions | Disposition |
|------|------------------|-------------|--------------|-------------|
| capture, clarify, organize, reflect, engage, answer | as in §D.1 | measured at M4 | measured at M4 | ABSORB or DELETE per OQ-1 |

Under reading (a) every row is ABSORB. Under (b) the rows with no measured use are DELETE and the verb, its tests, and its guidance lines go; the library functions stay (E4, and `goal.go` uses two). Under (c) the table is empty and REQ-TU-004 is reported unmet for the six.

**Registration.** An absorbed verb is registered by adding its constructor to `newTodoCmd`; the `gtd`-only registration line is deleted, so `gtd` inherits the verb through the shared tree. The mistyped-verb guard builds its known-verb list from the tree, so no separate list changes.

**Lane classification (REQ-TU-006).** `todoLaneReadOnlyVerbs` gains the read-only absorbed verbs by name. Measured: `reflect` reads through a read store on its default path (`ReflectGTDStore` over `newTodoReadStore`) and is currently refused for a lane; whether its `--rebuild-projection` flag writes anything is confirmed at M4 step 0 before the classification is fixed. `capture`, `clarify`, `organize`, `engage`, and `answer` write (queue rows, GTD tables, or the answers directory) and stay refused. AC-TU-006 enumerates the table, so a verb added later without a classification fails the test instead of inheriting the refusal by default.

**Guidance executes (REQ-TU-005).** `TestGuidanceVerbsResolve` extracts every `moai todo <word>` and `moai gtd <word>` token from the queue guidance corpus (the workflow skill in both trees, the router skill, the READMEs, the docs-site queue pages, and the help `Long` texts) and resolves each against the command tree. A token that resolves to no verb fails with its file and line. This is the t1590 clause as a standing test, not a one-time check.

## §D.5 The lane's own-card done (REQ-TU-015 to REQ-TU-017, REQ-TU-020 to REQ-TU-021)

**Two layers, deliberately.** The guard decides whether a lane may *attempt* `done`; the verb decides whether the close is *evidenced*.

1. *Guard layer.* In `todoRefuseLaneMutation`, a lane session's `done` is admitted when the single argument names the card the lane currently holds. "Holds" is read from the same record the dispatch path writes (the queue's current-dispatch record and the factory assignment whose owner label equals the lane's normalised label, using the existing label normalisation). Any other argument, any other verb, or an unreadable record is refused with the existing refusal text, so REQ-TU-016's matrix is the old behaviour plus one admitted shape.
2. *Verb layer.* When the caller is a lane session, `done` behaves as if `--require-landed` were set: it asks the existing landing predicate (the closed two-form evidence set of the auto-done scan) and refuses without evidence. This reuses `AutoDoneDecide`'s evidence forms and its misfire guards (reissued-id collision, SPEC not completed, non-landing declaration) rather than restating them.

**The receipt seam (REQ-TU-017).** The close still goes through `holdDoneApprovalGate`, so a factory-linked card is verified at the archive moment under the factory write transaction. What satisfies the receipt condition is one function, `laneCloseReceiptSatisfied`, whose body depends on OQ-5: (a) true when the landing evidence is present; (b) true when a receipt issued from the verdict, landing evidence, and remeasure hash verifies; (c) true only when a leader-issued receipt already verifies. The admission matrix tests call the function through the verb, so they are identical under (a) and (b), and under (c) the "automatic" criterion is reported unmet rather than faked.

**The log row.** The close appends a row to the existing execution log with `source` set to a new label for the lane path (the scan uses `auto-land`), so the two automatic paths are distinguishable. `todo undone` currently pairs its reversal row with the scan's log rows; M2 extends that pairing to the new label, and AC-TU-016 checks the reversal.

**Both records end `done` (REQ-TU-020).** Under OQ-7 (a), the verb closes the queue row (archive) and requests the factory transition for the same card in the same operation; neither record is left open. Which factory edge applies after M1 depends on OQ-6; the request goes through the table, never around it.

**The ten-card criterion (REQ-TU-021).** A fixture builds a queue and factory record with twelve cards, ten of which have landing evidence of the closed form and two of which do not, runs the closing acts, and asserts exactly the ten end `done` in both records with no operator input and the two stay open (the negative control). At sync, a live read-only command lists the ten most recent cards with landing evidence and their states, and the sync record carries its output.

## §D.6 Removing T18 and the T19 reservation (REQ-TU-018, REQ-TU-019)

**What changes.** In `card_transition.go`: the `T18` row, the `guardNoRemote` branch of `planTransition`, the `isReservedEdge` predicate with its call and the `ErrReservedEdge` value, and the two comments that name T19 as reserved. In `card_evidence.go`: `PushGateTarget`'s `done` result. In `factory_card.go`: the caller's handling of that result. In `card_pr_states.go`: the comment that cites T18.

**Why the pinned counts do not move.** The pair-count test runs with a remote configured. In that fixture `merged-local → done` and `pushed → ci-green` are already refused, so the accepted set stays 71 and the refused set 370. What changes is the error class: both requests now fail as undeclared (`ErrIllegalTransition`) instead of one failing as reserved. The final assertion at `fr_transition_test.go:179-181` keeps its meaning.

**Replacement path (OQ-6).** Under (a), a merged-local card in a no-remote repository completes through a named path; the minimum that keeps the machine closed is a declared edge guarded by the same human decider and receipt gate T18 used, or the lane own-card path of M2. Under (b), only the reservation goes. Under (c), M1 is deferred and the SPEC records the card's removal item as undelivered until `SPEC-GITHUB-FLOW-DEFAULT-001` closes.

**Collisions.** Draft `SPEC-FACTORY-COMPLETION-RECOVERY-001` REQ-FCR-010 and `SPEC-FACTORY-RECORD-REPAIR-001` item (3) edit the same rows in opposite directions. M1's pre-flight re-reads both statuses, and the milestone rebases on whichever lands first rather than resolving by guess.

## §D.7 The spec_id setter (REQ-TU-022 to REQ-TU-024)

**Why two carriers.** The queue row's `SpecID` is what `contract_report` reads; the factory card's `spec_id` is what kickoff reads (`<worktree>/.moai/specs/<spec_id>`). A setter that wrote one would leave the other empty, which is the observed double repair.

**Flow.** Under the queue lock: read the item; require queue state `picked`; if a factory card row exists, require its state in `{picked, assigned}`; require a well-formed identifier (the homestate pattern `^SPEC-[A-Z0-9]+(?:-[A-Z0-9]+)*$`, so the setter and the existing field-update gate cannot disagree); require the current spec_id empty, or equal to the requested one (no-op success); write the queue row and, inside the factory write transaction, the card row with a `card.fields` event; release. A failure at any step writes neither.

**Set-once is the safety property.** The existing freeze ("hints and pointers change only while picked") protects a card whose lane has already started from a pointer swap. The setter never swaps; it fills an empty value, and only until the lease state. A mutant that overwrites a non-empty value fails AC-TU-022.

**Interactions.**

| Existing path | Contract after this change |
|---------------|----------------------------|
| `todo next <n> --spec X` | unchanged: sets the queue spec_id at pick |
| `factory assign --spec X` | unchanged: sets the factory spec_id while `picked`; the freeze message stays |
| `todo unpick <n>` | unchanged: clears the queue spec_id; the setter can fill it again |
| kickoff SPEC lookup | unchanged code; passes because the factory spec_id is now non-empty |
| `contract_report` | unchanged code; reads the queue spec_id the setter wrote |

**Name and caller (OQ-8, OQ-9).** The default is a new `moai todo spec <n> <SPEC-ID>` verb refused for lane sessions like every queue mutation. If OQ-9 answers (b), the guard gains a second admitted shape (own card, empty spec_id) beside M2's.

## §D.8 Verification design

**Order.** Per milestone: write the test; run it; observe `EXPECTED_RED` for the stated reason (the assertion message names the missing behaviour, not a compile or fixture error, which would be `TOOL_FAILURE` and does not count); record the verbatim output and the tree HEAD; implement; run the new test plus the regression set; refactor with both green.

**Mutants considered (each one is a case the criteria must catch).**

| Criterion | Mutant that satisfies a weak version | Caught by |
|-----------|--------------------------------------|-----------|
| warning printed | warning on stdout | `--json` stdout must parse as JSON (AC-TU-002) |
| warning printed once | warning in the parent hook and again per verb | exact-count assertion across a subverb invocation |
| `gtd` delegates | a copied handler that drifts | differential fixture: same input through both spellings, same store effect (AC-TU-003) |
| guidance executes | verbs registered but guidance not scanned | scan counts tokens and asserts the count is at least the measured floor |
| lane own-card done | admit `done` for any card | other-lane row of the refusal matrix (AC-TU-015) |
| lane own-card done | skip the evidence check | no-evidence row (AC-TU-015) |
| T18 removed | rename the row instead of removing it | table test enumerates the declared edge set (AC-TU-017) |
| spec_id set-once | overwrite a non-empty value | different-id row (AC-TU-022) |
| residual ledger | an empty sweep | swept-count floor and the planted-mutant fixture (AC-TU-010) |

**Continued firing.** Each guard test lives in the ordinary suite, so non-execution is visible as a CI result. The substitution test additionally fails when its corpus shrinks below the measured floor, so a changed selector that sweeps nothing cannot read as green.

**Scoping.** Name-exact `-run` selections per milestone (`plan.md` §E); never an empty selection; the selector's matched-test count is read before the verdict.

## §D.9 Alternatives considered and not taken

- **Keep `gtd` as a fully supported alias.** Rejected by the card; it also keeps every guidance surface bilingual.
- **Delete `gtd` now.** Contradicts the one-release grace.
- **Rename internals with the user-facing name.** Changes persisted formats; separate SPEC (OQ-3).
- **A SessionEnd observer for automatic done in this SPEC.** Adds a hook under the advisory-check discipline and overlaps t1639; deferred by the OQ-4 default.
- **Overwrite-capable spec_id edit.** Would reopen the pointer-swap hazard the freeze exists for.
- **A second known-verb list for the mistyped-verb guard.** The tree already is that list.
