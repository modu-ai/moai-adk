# Design: SPEC-TODO-UNIFY-001

The mechanisms behind the requirements of `spec.md`. This is the HOW layer; identifiers and file names appear here and in the evidence, not in the requirements (a requirement may name an observable artifact path, such as a verdict file). Tree basis `026f48658f`. Revision 2: the dependency graph that orders these mechanisms is `plan.md` §A.

## §D.1 Current-state map

| Concern | Where it lives today | Fact |
|---------|---------------------|------|
| `moai todo` tree | `internal/cli/todo.go` `newTodoCmd` (registers 26 verbs at `:342`; one persistent pre-run hook that runs the lane guard) | the only place queue verbs are registered |
| `moai gtd` | `internal/cli/gtd.go` `NewGTDCommand`: takes `newTodoCmd()`, renames it, adds six stage-verb constructors at `:79` | no handler of its own for the queue verbs |
| Stage verbs | `gtd.go` (`newGTD{Capture,Clarify,Organize,Reflect,Engage}Cmd`), `gtd_answer.go` (`newGTDAnswerCmd`) | call `factory.CaptureGTDItem`, `ClarifyGTDItem`, `OrganizeGTDItemWithRelations`, `ReflectGTDStore`, `EngageGTDItem`; `goal.go` calls two of those library functions directly |
| Lane guard | `todo.go` `todoRefuseLaneMutation`, allowlist `todoLaneReadOnlyVerbs` (7 verbs) | refuses every other verb for a lane session, parent-with-arguments included |
| Close gate | `internal/cli/leader_approval_gate.go` `holdDoneApprovalGate`; `internal/homestate/leader_approval.go` `VerifyBinding` | wraps every close of a factory-linked card; the receipt must carry the leader role marker, an issuer that is not the performing owner, the card uuid, the run, the version, and a non-empty equal evidence hash |
| Receipt issuance | `moai factory approve` (`leader_approval_gate.go`), `FactoryDB.IssueLeaderApproval` | the only issuance; lane sessions are refused at the boundary |
| Auto-done scan | `internal/cli/todo_autodone.go`, `internal/factory/autodone_scan.go` | closed evidence set of two forms; three misfire guards; log `auto-done-log.jsonl`; source label `auto-land`; a candidate whose receipt is unverified is skipped |
| Verdict reader | `internal/homestate/card_evidence_readers.go:20-21` | `verdict: (PASS\|PASS-WITH-DEBT\|FAIL)` and `audited_sha: <12-40 hex>` lines |
| Merge verifier | `card_evidence_readers.go:290` `verifyMerge` | two-parent merge, tree equal to the second parent's, reachable from `refs/heads/<integration>`, re-measure record keyed to the tree satisfies the gate |
| Transition table | `internal/homestate/card_transition.go` (`transitionTable`, 71 accepted pairs pinned at `fr_transition_test.go:166`) | T18 row at `:157`; reserved-edge branch at `:333`; guard body at `:722-747` |
| Push gate | `card_evidence.go:64` `PushGateTarget`; caller `internal/cli/factory_card.go:3331` | returns `pushed` with a remote, `done` without one |
| spec_id carriers | queue item `SpecID` (set by `todo next <n> --spec`); factory card `spec_id` (set by `factory assign --spec` while `picked`; refusal at `card_picked.go:185`) | no post-pick path |
| Slash and skills | `.claude/commands/moai/{gtd,todo}.md`, `.agents/skills/moai-{gtd,todo}/SKILL.md` (generated), `internal/template/gtd_canonical_surface_test.go`, `published_skills.go` | `todo.md` dispatches `gtd` and a test pins it; the router skill already routes `todo` canonical (`SKILL.md:81-82`) |
| Workflow document | `.claude/skills/moai/workflows/gtd.md` (template and deployed copies byte-identical at `026f48658f`); `.moai/workflows/gtd.md` (deployed-only, no template source, differs) | the path is named by five rule files' `paths:` keys, `embed_manifest_gen.go:516`, the recorded fixture, and seven test files (`research.md` §R.10) |

## §D.2 The warning alias (REQ-TU-001 to REQ-TU-003)

**Shape.** `NewGTDCommand` keeps building from `newTodoCmd()`. After M4 it adds nothing else (the stage verbs are on the todo tree), so `gtd` is the todo tree under another name, which is what REQ-TU-001 states.

**The warning.** One pre-run function, chained in front of the tree's persistent pre-run hook (the lane guard), writes the warning once per process. Cobra v1.10.2 orders `execute` as: parse flags (`command.go:919-922`), return on a help flag (`:926-935`), `ValidateArgs` (`:968-970`), then the persistent pre-run loop (`:984-998`). The hook therefore fires once per executed invocation, for the parent and every subcommand alike, which satisfies "exactly one"; it never fires for a help request, a flag-parse error, or an argument-validation refusal (a mistyped verb, a missing argument), because each returns first. REQ-TU-003 states exactly that boundary instead of promising a warning before those refusals (iteration 1 promised it for the mistyped verb; the order above shows the mistyped-verb guard lives in the parent's `Args` function, `todo.go:260-279`, and so runs before the hook). The text goes to the command's stderr, never stdout, and follows the existing deprecation convention (`Warning:` prefix; precedent `warnDeprecatedAgentModelFlags` in `agent_model_flags_retired.go`). Its first line is `Warning: moai gtd has been renamed to moai todo`; it names the `moai todo` spelling of what was run (the verb token is taken from the executing command) and states removal one release after the release that first carries the warning.

**The removal statement** is one constant. AC-TU-001 asserts the warning contains it and that it is not empty, so an unset constant cannot ship. No release number is hard-coded (Q2).

**Why not hide `gtd`.** The existing help-parity test requires `gtd --help` and `todo --help` bodies to be equal modulo the name; hiding the command from the root listing removes the discoverability the grace period exists to give. Not done; revisitable.

**Why the lane guard still sees the right name.** `todoSurfaceName` already reports the spelling the operator typed, so the guard's refusal text stays truthful under both names.

**Dependency-initialisation.** `gtd` is not in `trivialCommands` and must not be added: the queue verbs read dependencies, and a warning alias that skipped initialisation would diverge from the `todo` form.

## §D.3 The substitution mechanism (REQ-TU-009 to REQ-TU-011)

Hand-editing 67 files and about 244 lines is the failure mode the card names. The mechanism is one script, `scripts/gtd-to-todo-substitute.sh`, POSIX `sh` plus `sed` without `-i` (write to a temporary file and move it, so it behaves the same on macOS and Linux), driven by a class table and an allowlist file. Whether `sh` is available to the test on every CI runner is not measured; the test fails loudly rather than skipping if it is absent, and the script moves to a Go helper in that case (decided at M7 RED).

**Class table.**

| Class | Token forms | Action |
|-------|-------------|--------|
| S1 prose command | `moai gtd <word>` where `<word>` is a queue verb (the 26 plus the six stage verbs, all registered on `todo` after M4) or end of token | `moai todo <word>` |
| S2 slash | `/moai gtd`, `/moai:gtd` | `/moai todo`, `/moai:todo` |
| X1 reviewed edits | argument vectors (`'moai', 'gtd'`), environment variable names (`CUTOVER_GTD_LIST_CMD`), in-tree error and help strings | not touched by the script; edited in M7 with their own tests |
| E1 history | `.moai/specs/**`, `.moai/reports/**`, `CHANGELOG.md` | excluded |
| E2 binary | any path `git grep -I` treats as binary | excluded |
| E3 recorded fixtures | `internal/cli/testdata/codex-rollouts-t1171/**`, `internal/template/testdata/binding_ledger.json` | excluded |
| E4 internals | Go identifiers `GTD*`, `gtd_*` names, `gtd-<hex>` ids, `.moai/gtd/`, the `internal/factory/gtd_*.go` and `backlog_gtd_schema.go` contents, and the path `workflows/gtd.md` wherever it is named | excluded (Q3) |
| E5 alias statements | lines listed in the allowlist file | excluded; they are the hand-authored deprecation statements (the `/moai:gtd` wrapper notice, the alias lines of the router skill and the help text) |
| E6 always-loaded | `AGENTS.md`, `internal/template/templates/AGENTS.md.tmpl` | excluded from M8; edited by hand in M9 |

The earlier class that renamed `workflows/gtd.md` and merged the `moai-gtd` docs-site pages is dropped: the path is an identifier (`research.md` §R.10) and the pages are URLs; both wait for the removal release.

**The input set** is `git ls-files` intersected with the S1 and S2 patterns minus E1 to E6, never a filesystem walk, so untracked scratch cannot be touched.

**Dry-run contract.** `--dry-run` prints one line per file (`would-change <path> <count>`) and a total, changes nothing, and exits 0. The set it prints must equal the pre-flight set (E-01) minus the excluded classes; M8 compares the two sets and stops on any difference.

**Idempotence (REQ-TU-010).** After `--apply`, a second run prints a total of 0 and `--apply` changes no file (`git status --short` empty for the touched paths).

**Residual ledger (REQ-TU-011).** `TestNoResidualGTDSpelling` walks the tracked guidance corpus (S1 and S2 scope, same exclusions), reads the allowlist file the script reads, and fails on any `moai gtd` or `/moai gtd` outside it. Safeguards: it asserts the swept file count is at least 5,000 (an empty sweep is not a pass); it asserts every path E-01 printed that is not in E3 or E6 is among the swept set; and a fixture containing a planted `moai gtd add` line proves the test can go red. It runs in the ordinary suite, so a later edit that reintroduces the spelling is caught by CI without anyone asking.

**Docs-site obligations in the same batch.** Four locales change together; the locale-parity baseline is not extended; the docs-site verification recipe (warning-free build, parity, URL and emoji scans) runs before the batch is accepted.

## §D.4 The six stage verbs (REQ-TU-004 to REQ-TU-006)

All six are absorbed (Q1): the constructors stay where they are and their registration line moves from `NewGTDCommand` to `newTodoCmd`; the `gtd`-only registration line is deleted, so `gtd` inherits the verbs through the shared tree. The mistyped-verb guard builds its known-verb list from the tree, so no separate list changes. No inventory table is produced: nothing is deleted, so there is no disposition to measure beyond the one flag below.

**Lane classification (REQ-TU-006).** `reflect` reads through a read store (`ReflectGTDStore` over `newTodoReadStore`); its `--rebuild-projection` flag calls `BuildPrivateGTDProjectionFromStore`, which creates temporary files and renames them into the projection directory (`internal/graph/gtd_private.go:179`, `:264`, `:273`), so the flag form writes. The guard therefore admits `reflect` only when that flag is absent and refuses the flag form with the standing text; `capture`, `clarify`, `organize`, `engage`, and `answer` write (queue rows, GTD tables, or the answers directory) and stay refused. AC-TU-006 enumerates the table, so a verb added later without a classification fails the test instead of inheriting the refusal by default.

**Guidance executes (REQ-TU-005).** `TestGuidanceVerbsResolve` extracts every `moai todo <word>` and `moai gtd <word>` token from the queue guidance corpus (the workflow skill in both trees, the router skill, the READMEs, the docs-site queue pages, and the help `Long` texts) and resolves each against the command tree. A token that resolves to no verb fails with its file and line. This is the t1590 clause as a standing test, not a one-time check.

## §D.5 The lane's own-card done (REQ-TU-015 to REQ-TU-017, REQ-TU-020, REQ-TU-021, REQ-TU-025)

**Two layers, deliberately.** The guard decides whether a lane may attempt `done`; the verb decides whether the close is evidenced.

1. *Guard layer.* In `todoRefuseLaneMutation`, a lane session's `done` is admitted when the single argument names the card the lane currently holds. "Holds" is read from the same record the dispatch path writes (the queue's current-dispatch record and the factory assignment whose owner label equals the lane's normalised label). Any other argument, any other verb, or an unreadable record is refused with the existing refusal text, so REQ-TU-016's matrix is the old behaviour plus one admitted shape.
2. *Verb layer.* When the caller is a lane session, `done` behaves as if `--require-landed` were set: it asks the existing landing predicate (the closed two-form evidence set of the auto-done scan) and refuses without evidence. This reuses `AutoDoneDecide`'s evidence forms and misfire guards rather than restating them.

**The automatic receipt (REQ-TU-017; Q5, leader-confirmed 10-18 conditional).** The gate is unchanged: `VerifyBinding` still demands the leader role marker, an issuer that is not the performing owner, the card uuid, the run, the version, and an equal non-empty evidence hash. What changes is who may mint, and only at one moment: the shared close helper, immediately before the gate verifies, calls one issuing function that re-reads three facts and writes nothing unless all three hold.

| Fact | How it is read | Source |
|------|----------------|--------|
| (1) state `merged-local` | the factory row of the card's recorded run | `loadCard` |
| (2) committed PASS-family verdict | the file `.moai/reports/<card-id>/sync-audit.md` exists in the card worktree's `HEAD` tree (read from the commit, not the working copy), carries `verdict: PASS` or `PASS-WITH-DEBT` and one consistent `audited_sha`, and that sha is an ancestor of `refs/heads/<integration>` | the reader of `card_evidence_readers.go:20-21`, `isAncestor` |
| (3) valid re-measure record | the record keyed to the merge commit's tree satisfies the same gate callback T16 used | the `verifyMerge` callback |

If any fact is absent the function returns without writing; the gate then refuses with its receipt-absent error and the card waits for `moai factory approve` (the leader's manual path, still lane-refused). The receipt is written with `Issuer: "auto"` and `IssuerRole: leader` (the role marker is what the gate accepts; the issuer string is what distinguishes it). `CanonicalApprovalOwnerLabel("auto")` is `auto`, which never equals a lane label, so the performer-is-not-approver clause holds without change. **Revocation (REQ-TU-025).** A revocation record keyed by (card uuid, run id) lives beside the receipt store (an additive table owned by M2a). The verifier excludes an `auto` receipt that has a revocation row, which reads as receipt-absent; the issuing function refuses while a revocation row exists for the pair; a leader-issued receipt is unaffected because the exclusion is keyed on the `auto` issuer. The revocation surface is a leader-only flag of `moai factory approve`; the lane refusal that verb already applies covers it.

**The close helper and its order (REQ-TU-020).** Shared by the lane path and the scan: (1) mint if the gate needs a receipt and the facts hold; (2) if the factory row is `merged-local`, request edge T29 with the row's expected version (the transition's own receipt gate then runs and re-stamps the receipt to the post-transition version inside its transaction, the chain already used for T18 and T20); (3) archive the queue row under the held gate, whose verification reads the re-stamped receipt. No factory transaction is held while waiting for the queue lock. **Recovery.** A failure after step 2 and before step 3 leaves the factory row `done` and the queue row open: the helper exits non-zero naming both records. A repeat of the same act finds the factory row already `done`, skips step 2, and completes step 3 under the re-stamped receipt; a failure the other way (step 2 refused) leaves both unchanged. The factory edge goes first on purpose: the reverse order archives the queue row and leaves a factory row no live card references, which no retry can reach. A queue row left open after step 2 is also closable by the scan, since the re-stamped receipt verifies. A factory row in another state (for example `pushed`) skips step 2 and is named in the log; with no automatic receipt for it (fact 1 fails), the queue row closes only under a leader-issued receipt.

**The log row.** The close appends a row to the existing execution log with `source` set to a new label for the lane path (the scan keeps `auto-land`) and the receipt issuer, so the automatic paths are distinguishable. `todo undone` currently pairs its reversal row with the scan's log rows; M2b extends that pairing to the new label.

**The ten-card criterion (REQ-TU-021).** A fixture builds a queue and factory record with twelve cards, ten with landing evidence, the three facts, and a `merged-local` row, and two without landing evidence; it runs the closing acts and asserts exactly the ten end `done` in both records with no operator input and the two stay open (the negative control). At sync, a live read-only command lists the ten most recent cards with landing evidence and their states, and the sync record carries its output; any card whose factory row is not `merged-local` is named there as outside this SPEC's edge.

## §D.6 Removing T18 and the T19 reservation, declaring T29 (REQ-TU-018, REQ-TU-019)

**What is removed.** In `card_transition.go`: the `T18` row, the `guardNoRemote` branch of `planTransition` (`:722-747`), the `isReservedEdge` predicate with its call (`:333`) and the `ErrReservedEdge` value, and the two comments that name T19 as reserved. In `card_pr_states.go`: the comment that cites T18.

**What replaces it: edge T29.** One row for the pair `merged-local → done` with a new guard, `guardLanded`. It differs from T18 on every axis a test can observe: T18 required the human decider, was refused while a remote exists, and checked only the receipt; T29 admits decider `human` (the push gate path) or `audit` (the evidence-derived close path), has no remote condition, and requires (a) the card's recorded merge commit to be reachable from the landing ref (with a remote, the remote-tracking ref of the integration branch through the T17 verifier `verifyPushed`; without one, `refs/heads/<integration>` through the reachability test `verifyMerge` applies at `card_evidence_readers.go:316`), re-verified at request time, and (b) the receipt gate (`verifyTransitionApproval`, with the receipt re-stamp). The caller passes the integration branch it resolved from the landed-ref chain, so with a remote both rules name the same ref. The push gate keeps requesting `done` for a no-remote repository, which T29 now declares, so `PushGateTarget`, `decideOne`, and their tests keep their meaning and `decide --gate push` is the no-remote completion path: leader-issued receipt in, `done` out, no push step.

**Why the pinned counts do not move.** The pair-count test runs with a remote configured and no receipt or landed ref in the fixture, so `merged-local → done` and `pushed → ci-green` are refused there before and after. What changes is the error class for the second: it now fails as undeclared (`ErrIllegalTransition`), not as reserved. The final assertion at `fr_transition_test.go:179-183` keeps its meaning.

**Why this is not the circle of iteration 1.** The completion path is declared before the removal (M1's RED order puts the T29 tests first) and is exercised on a no-remote repository by AC-TU-018; it does not route through the lane path, so removing the old row cannot leave the machine without an edge into `done`.

**Collisions.** Draft `SPEC-FACTORY-COMPLETION-RECOVERY-001` REQ-FCR-010 would implement T19 (the opposite of the removal) and `SPEC-FACTORY-RECORD-REPAIR-001` moved its `pushed → done` item to t1639. M1's pre-flight re-reads both statuses; on a collision the milestone returns a blocker report naming the rows instead of resolving by guess. This SPEC declares no `pushed → done` edge.

## §D.7 The spec_id setter (REQ-TU-022 to REQ-TU-024)

**Why two carriers.** The queue row's `SpecID` is what `contract_report` reads; the factory card's `spec_id` is what kickoff reads (`<worktree>/.moai/specs/<spec_id>`). A setter that wrote one would leave the other empty, which is the observed double repair.

**Flow (set-once per carrier).** Under the queue lock the verb: reads the queue item and requires queue state `picked`; validates the identifier with the homestate pattern `^SPEC-[A-Z0-9]+(?:-[A-Z0-9]+)*$` (so the setter and the existing field-update gate cannot disagree); reads the factory row if one exists and requires its state in `{picked, assigned}`; and applies the per-carrier rule before any write: empty is filled, equal is skipped, different refuses the whole call. It then opens one factory write transaction through a new homestate function. Inside that transaction the function loads the card, sets `spec_id`, bumps the version, appends a `card.fields` event, and runs the queue write in the transaction's pre-commit step (the seam `TransitionRequest.BeforeCommit` already provides for transitions, `card_transition.go:286`, used so a caller can keep the queue's dispatch record in step), then commits. The queue lock is held across all of it, the same order the claim path uses (queue lock, then factory write).

**Failure behaviour.** A failure of the queue write inside the pre-commit step rolls the factory transaction back: both carriers unchanged. A failure of the factory commit after the queue write succeeded leaves the queue carrier set and the factory carrier empty: the verb exits non-zero naming both carriers, and a repeat with the same identifier finds the queue carrier equal (skip) and the factory carrier empty (fill) and succeeds. In that half-written state a different identifier is refused by the per-carrier rule. No sequence leaves a carrier holding a value other than the requested one. The window is one commit wide and the repeat is the recovery; no automatic recovery runs. Fault seams for the two cases follow the existing `cardTransitionFault` pattern (a package variable the test sets).

**Set-once is the safety property.** The existing freeze ("hints and pointers change only while picked") protects a card whose lane has already started from a pointer swap. The setter never swaps; it fills an empty value, and only until the lease state. A mutant that overwrites a non-empty value fails AC-TU-021.

**Interactions.**

| Existing path | Contract after this change |
|---------------|----------------------------|
| `todo next <n> --spec X` | unchanged: sets the queue spec_id at pick |
| `factory assign --spec X` | unchanged: sets the factory spec_id while `picked`; the freeze message stays |
| `todo unpick <n>` | unchanged: clears the queue spec_id; the setter can fill it again |
| kickoff SPEC lookup | unchanged code; passes because the factory spec_id is now non-empty |
| `contract_report` | unchanged code; reads the queue spec_id the setter wrote |

**Name and caller (Q8, Q9).** A new `moai todo spec <n> <SPEC-ID>` verb, refused for lane sessions like every queue mutation (it is not on the read-only allowlist). The board record `d-20261010T001732Z-d4e8` assigns a lane-side surface to card t1621; that sibling's narrowed SPEC lists it out of scope, so no lane-side surface is planned anywhere (`research.md` §R.8).

## §D.8 Verification design

**Order.** Per milestone: write the test; run it; observe `EXPECTED_RED` for the stated reason (the assertion message names the missing behaviour, not a compile or fixture error, which would be `TOOL_FAILURE` and does not count); record the verbatim output and the tree HEAD; implement; run the new test plus the regression set; refactor with both green.

**Mutants considered (each one is a case the criteria must catch).**

| Criterion | Mutant that satisfies a weak version | Caught by |
|-----------|--------------------------------------|-----------|
| warning printed | warning on stdout | `--json` stdout must parse as JSON (AC-TU-002) |
| warning printed once | warning in the parent hook and again per verb | exact-count assertion across a subverb invocation |
| warning scope | warning also printed for argument-refused invocations | refusal rows of AC-TU-002 |
| `gtd` delegates | a copied handler that drifts | differential fixture (AC-TU-003) |
| six verbs absorbed | a verb deleted, left on `gtd`, or registered with a no-op handler | exact 32-verb sets and the store-equality run (AC-TU-004) |
| guidance executes | verbs registered but guidance not scanned | scan floors 300 tokens / 17 files (AC-TU-005) |
| lane own-card done | admit `done` for any card | other-lane row (AC-TU-015) |
| lane own-card done | skip the evidence check | no-evidence row (AC-TU-015) |
| automatic receipt | issue on two of the three facts | three two-of-three rows (AC-TU-016) |
| automatic receipt | issue with an uncommitted or FAIL verdict, or a non-ancestor `audited_sha` | the three verdict rows (AC-TU-016) |
| automatic receipt | record a leader label as issuer | issuer assertion (AC-TU-016) |
| revocation | revoke without blocking re-issuance | second-issuance row (AC-TU-019) |
| T18 removed | rename the row, keep the human decider and the remote refusal | remote row and unreachable-merge row (AC-TU-018) |
| T29 declared | omit the receipt gate | no-receipt row (AC-TU-018) |
| both records close | archive the queue first | order assertion (AC-TU-014) |
| failure between records | repeat requests the factory edge again | stale-version row (AC-TU-023) |
| spec_id set-once | overwrite a non-empty value | different-id row (AC-TU-021) |
| spec_id carriers | write only the queue carrier | factory-row assertion (AC-TU-021) |
| spec_id recovery | leave a divergent value after a fault | repeat rows (AC-TU-025) |
| residual ledger | an empty sweep | swept-count floor, the E-01 coverage assertion, and the planted-mutant fixture (AC-TU-010) |

**Continued firing.** Each guard test lives in the ordinary suite, so non-execution is visible as a CI result. The substitution and residual tests additionally fail when their corpus shrinks below the measured floors, so a changed selector that sweeps nothing cannot read as green.

**Scoping.** Name-exact `-run` selections per milestone (`plan.md` §E); never an empty selection; the selector's matched-test count is read before the verdict.

## §D.9 Alternatives considered and not taken

- **Keep `gtd` as a fully supported alias.** Rejected by the card; it also keeps every guidance surface bilingual.
- **Delete `gtd` now.** Contradicts the one-release grace.
- **Delete the unused stage verbs.** No in-tree caller is not evidence of no use; the tree holds no usage telemetry, five tests pin the CLI path, and deleting removes a documented user-facing feature (Q1).
- **Rename internals and the workflow document with the user-facing name.** Changes persisted formats and identifiers named by rule `paths:` keys, an embed line, and a recorded fixture; separate change (Q3).
- **A SessionEnd observer for automatic done in this SPEC.** Adds a hook under the advisory-check discipline and overlaps t1639 (Q4).
- **A receipt policy that waives the receipt on landing evidence alone.** Removes the gate; the chosen form keeps the gate and proxies it by machine verification (Q5).
- **A lane own-card done that stays on the leader's receipt.** Leaves the ten-card criterion unreachable without a leader act (Q5).
- **Overwrite-capable spec_id edit.** Would reopen the pointer-swap hazard the freeze exists for.
- **A second known-verb list for the mistyped-verb guard.** The tree already is that list.
