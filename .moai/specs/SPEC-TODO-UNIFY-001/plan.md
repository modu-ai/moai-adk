# Plan: SPEC-TODO-UNIFY-001

Tree basis: `026f48658f` (branch `WT-todo-unify`; plan artifacts and the audit report sit on top, the rest of the tree is identical, see `research.md` §R.1). Methodology: TDD (`quality.yaml` `development_mode: tdd`). Priorities and ordering only; no time estimates. Revision 3 (audit iteration 2 FAIL 0.80; revision 2 answered iteration 1 FAIL 0.62); the decisions of `decision-index.md` Q1 to Q9 are written into `spec.md`. Revision 3 changes milestones M1 (T29's deciders), M2a (facts read from merged objects, issuance inside T29's transaction), M2b (log issuer, `pushed` row), M3a/M3b (state set, completing repeat), M5 (warning at the run step), and M7/M8 (where `TestNoResidualGTDSpelling` lands); the graph below gains the arrow M1 → M2a. Revision 4 (post-ceiling correction under standing rule C; audit iteration 3 FAIL 0.86, D29 to D31) changes M1 (T29's landing test without a remote, the raw reader `gitReadRaw`, a CLI-level row; `factory_card.go` is read, not edited) and M2a (the reserved issuer `auto`, raw reads, the graft and shallow refusal) and adds the known debt KD-1 to §H; the graph and the milestone set are unchanged.

## §A Decision-reversibility ranking and the dependency graph

Sections and milestones below are ordered by how likely the decision they rest on is to change, highest first. A data-model or interface decision comes before a mechanical step, so review attention lands where a reversal would cost most.

| Rank | Decision | Why it may change | Decision row | Milestone |
|------|----------|-------------------|--------------|-----------|
| 1 | Remove T18 and the T19 reservation; declare the evidence-guarded edge T29 | Removes the human-decided no-remote completion; two draft SPECs touch the same rows | Q6 | M1 |
| 2 | Who may close a card without an operator, and the automatic receipt (three facts, issuer `auto`, revocable) | Changes the approval gate's trust model; leader-confirmed 10-18, conditional | Q4, Q5, Q7 | M2a, M2b |
| 3 | Shape and caller of the post-pick spec_id setter | A new verb on a shipped command | Q8, Q9 | M3a, M3b |
| 4 | The six stage verbs registered on `moai todo` | A user-facing registration | Q1 | M4 |
| 5 | Warning alias behavior and the removal statement | User-visible stderr contract | Q2 | M5 |
| 6 | Slash, skill, and workflow-document wording | Edits of shipped surfaces (paths kept) | Q3 | M6 |
| 7 | Substitution classes and the script | Mechanical; reversible by a revert | Q3 | M7 |
| 8 | Applying the script | Mechanical | none | M8 |
| 9 | Always-loaded text (`AGENTS.md`) and rule statements | Edited last (cache directive 3); size-bounded | none | M9 |

**Dependency graph (replaces the iteration-1 independence claim, which `design.md` §D.5 and §D.6 contradicted).** An arrow means "needs the earlier milestone's code or test to exist".

```
M1 ──► M2a ──► M2b ────────────────────────► M9
M3a ──► M3b                       (no other milestone depends on it, and it needs none)
M4 ──► M5 ──► M6 ──► M7 ──► M8 ──► M9
```

Reasons, one per arrow: M2a hooks the automatic issuance into the planning step of edge T29 (the receipt is written inside the closing edge's transaction), so it needs the edge M1 declares; M2b requests T29 through the close helper and relies on the issuance M2a defines (M2b needs M1 through M2a); M3b registers a verb on the tree M3a's function backs; M5's warning must cover the verbs M4 registers, or the six would run under `gtd` without it; M6 inverts the slash files before M7's script runs, so the script does not rewrite the wrapper's own notice; M7 supplies the script and allowlist M8 runs; M8 converts the corpus before M9 adds the allowlisted `AGENTS.md` row; M9 states the lane own-card done and the T18/T19 removal that M1 and M2b deliver.

Same-file order (not a logical dependency): `internal/cli/todo.go` is edited by M2b (lane guard), M3b (verb registration), and M4 (verb registration), so the run order is M2b, then M3b, then M4. `internal/cli/factory_card.go` is edited only by M7.

**Run order.** M1, M2a, M2b, M3a, M3b, M4, M5, M6, M7, M8, M9. A milestone that discovers its premise false returns a blocker report; the graph says which later milestones it stops.

## §B Decisions

The nine questions iteration 1 recorded are decided by the lane's decision ladder (`auto-semantics.md` §6); each row of `decision-index.md` carries the evidence, the alternate rejected, and a one-line decision record. Summary, with the requirement each answer lands in:

| Q | Decision | Lands in |
|---|----------|----------|
| Q1 | Register all six stage verbs on `moai todo`; delete none | REQ-TU-004, REQ-TU-006 |
| Q2 | The warning states removal "one release after the release that first carries the warning"; no literal release number | REQ-TU-002 |
| Q3 | Rename the user-facing spelling only; internals, persisted formats, and the workflow document path stay | REQ-TU-007, REQ-TU-012 |
| Q4 | This SPEC delivers the lane own-card path, the automatic receipt inside the two close paths, and edge T29; no observers | REQ-TU-017, REQ-TU-020, Out of Scope |
| Q5 | The receipt is issued automatically only when three machine facts hold (read from the merge commit and evidence commit the row names, issued inside the closing edge's transaction, revision 3), issuer `auto`, revocable; otherwise the leader's manual approval stays (fail-closed) | REQ-TU-017, REQ-TU-025 |
| Q6 | Remove T18 and the T19 reservation now; declare T29 (`merged-local → done`, evidence and receipt guarded) | REQ-TU-018, REQ-TU-019 |
| Q7 | Both records end `done`; "landed" is reachability from the landing ref; factory edge first, queue archive second, with a stated recovery | REQ-TU-020 |
| Q8 | A new verb `moai todo spec <n> <SPEC-ID>` writing both carriers, set-once per carrier; first fill from `picked` through `kickoff`, a completing repeat in any non-terminal state (revision 3) | REQ-TU-022, REQ-TU-023 |
| Q9 | Leader only | REQ-TU-023 |

## §C Pre-flight

Run before M1 (read-only; commands are plain, one per call):

1. `git rev-parse --short HEAD` and `git branch --show-current` (expect a descendant of `026f48658f` and `WT-todo-unify`).
2. Re-measure the substitution baseline: `git grep -I -l 'moai gtd' -- . ':!.moai/specs' ':!.moai/reports' ':!CHANGELOG.md'` (expect 73 paths at `026f48658f`; the list is in `research.md` §R.4). Re-measure after absorbing `main`; never re-cite 73 without re-measuring.
3. Confirm the open sibling facts have not moved: `SPEC-GITHUB-FLOW-DEFAULT-001` status, `git_strategy.manual.workflow`, `SPEC-FACTORY-COMPLETION-RECOVERY-001` status (its REQ-FCR-010 would implement T19, and its REQ-FCR-014 "no lane surface shall mint a receipt" is the text REQ-TU-017 narrows for one path, `design.md` §D.5; if that draft is no longer `draft`, M1 and M2a return a blocker report naming REQ-FCR-010 and REQ-FCR-014), `SPEC-FACTORY-RECORD-REPAIR-001` status, and whether card t1639 has started (its `pushed → done` edge and observers).
4. Build the tool under measurement from this tree and invoke it by path; record the judging build's commit next to the tree HEAD (verification-claim-integrity §2.2). The installed `moai` at plan time is `ee981a932-dirty`, a strict ancestor 516 commits behind.
5. Windows build check: `GOOS=windows GOARCH=amd64 go build ./...`.
6. Verification scoping: changed packages and direct importers (§F). No full-suite run; no background load; take the heavy-run slot only for the `internal/cli` name-exact selection, per the lane protocol.
7. Classify the 67 paths M8 will convert (the 73 of E-01 minus the 4 recorded fixtures minus `AGENTS.md` and its template) into template-sourced, template-mirror, and deployed-only with `git ls-files`; the classification is the M8 input (D12).
8. (M2a, read-only) Count the rows that already carry the issuer `auto` in the project's `leader_approvals` store; a non-zero count is a blocker report, because such a row cannot be told from an automatic one once the string is reserved (`design.md` §D.5, case ii). Not measured at plan time (Gap).

## §D Constraints

- PRESERVE: `internal/factory/gtd_*.go` and `backlog_gtd_schema.go` content, the `gtd_*` table names, `gtdIDPattern`, the `.moai/gtd/answers` path, the `workflows/gtd.md` path and the strings that name it (five rule files' `paths:` keys, the generated embed line, the recorded fixture), the 71/370 pair counts in `fr_transition_test.go`, the auto-done scan's closed evidence set and three misfire guards, the lane guard's REQ-SD-015 refusal text, `moai factory approve` as the leader's manual issuance path (lane-refused), and the receipt gate's verification (`VerifyBinding`) unchanged.
- No test skips (`t.Skip*`) added; no timeout value raised, in code or in any test invocation.
- Templates are edited in `internal/template/templates/`, never first in `.claude/`.
- The always-loaded files in the conversion set are `AGENTS.md` and `internal/template/templates/AGENTS.md.tmpl` only (the four rule files that carry the spelling have `paths:` keys, measured by `git grep -m1 '^paths:'`, `research.md` §R.10). Their edit waits for M9 (cache-aware directive 3) and carries the byte-size and non-invoking-cost statement in the commit body when it grows the file by more than 1,000 bytes.
- Commits stage by explicit pathspec; every commit message carries `card t1641`.

## §E Milestones

Each milestone: RED first (the new test observed failing for the stated reason, classified `EXPECTED_RED` per the TDD result contract, verbatim output recorded in `progress.md` §E.2), then GREEN (new test plus the regression set), then refactor. Size figures are illustration-level (files / changed lines); a milestone that would exceed roughly 10 files or 400 lines at run time splits at the seam it names before it starts. Every `-run` selector below names tests that exist or that the milestone's RED step creates; the selected count is read before the verdict.

### M1 — Remove T18 and the T19 reservation, declare T29 (Priority High)

- **RED (in this order).** (1) The T29 matrix tests, written first and observed red: in a no-remote fixture a merged-local card with a reachable merge commit and a leader-issued receipt reaches `done`; with the branch reset so the merge is unreachable, or with no receipt, it is refused and the row is byte-identical; in a remote fixture T29 admits only when the merge commit is reachable from the remote-tracking landed ref; the decider rows (`TestLandedEdgeT29Deciders`): `human` and `audit` admitted, an empty or any other decider refused with the decider refusal. (2) The removal tests: no edge guarded by the no-remote completion and no reserved-edge refusal remain; `pushed → ci-green` fails with the undeclared-transition error. (3) The pair-count pin stays green (in the fixture the T29 request lacks a receipt, so the pair stays refused). (4) The CLI-level rows (`TestPushGateNoRemoteNoConfig`, in `internal/cli`, through `factoryDecideCards`, the body of `decide --gate push`): no remote, no `git-strategy.yaml`, a valid receipt, a merge commit contained in a local branch other than the card's own: the card reaches `done`; the same with the merge commit contained only in the card's own branch: refused, row byte-identical; a configured branch that does not contain the merge commit while another local branch does: refused (AC-TU-018).
- **GREEN.** Remove row T18, the `guardNoRemote` branch, the reserved-edge predicate with its error value and call; add T29 for the pair `merged-local → done` with the evidence-and-receipt guard of `design.md` §D.6, checking its decider first (`human` or `audit`; the `DeciderAudit` doc comment, "edge T8a only", is updated to name T29); the push gate keeps requesting `done` for a no-remote repository, which T29 now declares. T29's no-remote landing test is exact for a non-empty `IntegrationBranch` and, for an empty one, `git for-each-ref --contains` minus the card's own branch, read through a new raw reader `gitReadRaw` (`design.md` §D.5 reader, §D.6).
- **Files (about 9 / 400 lines; if M1 would exceed 10 files or 400 lines at run time, the CLI-level rows and the containing-branch helper move to an M1b that follows it).** `internal/homestate/card_transition.go`, `card_record.go` (the `DeciderAudit` comment), `card_pr_states.go` (comment), `card_evidence.go` (`gitReadRaw`, beside `gitRead`), `card_evidence_readers.go` (the containing-branch helper), tests `fr_transition_test.go`, a new `fr_landed_edge_test.go`, `leader_approval_test.go`, `f7_gate_owner_admission_test.go`, and a new `internal/cli/factory_push_gate_noconfig_test.go`. `internal/cli/factory_card.go` is read and **not edited**: the push gate caller already forwards the configured target, possibly empty (`factory_card.go:3238`, `:3394`), and T29's no-remote test is defined for the empty value (`design.md` §D.6; the audit's alternative of resolving the branch with `factoryResolveIntegrationBranch` does not work, §D.6).
- **Packages.** `internal/homestate`; `internal/cli` (one new test file; its production code is unchanged).
- **Name-exact selection.** `internal/homestate`: `-run '^(TestFR_AC005_TransitionTableEdgeCount|TestPushGateTargetByRemote|TestLandedEdgeT29.*|TestReservedEdgeRemoved)$'` (the last two are created at RED; the first two exist); `internal/cli`: `-run '^TestPushGateNoRemoteNoConfig$'` (created at RED). Also the leader-approval and evidence tests the removal touches, listed from `git grep -n '^func Test'` in the files named above at RED time.

### M2a — The automatic receipt and its revocation (Priority High; needs M1)

- **RED (AC-TU-016, AC-TU-019; fixtures are real temporary git repositories).** (1) The three-fact matrix: of the eight combinations of (merged-local state, a verdict file tracked in the merge commit's tree whose `audited_sha` is a prefix of the recorded evidence commit and where every path in which the card tip differs from that commit and from the merge's first parent lies under `.moai/reports/<card-id>/`, a re-measure record that names the merged tree and the merge's first parent and is valid), exactly one issues a receipt and each two-of-three combination issues nothing; (2) the flows lanes run today pass (card tip = evidence commit plus an evidence-only commit; card tip = evidence commit; and the landing-train shape, evidence commit, then main absorbed, then an evidence-only commit); (3) stale evidence is refused: a lane code commit after the audit, an absorption conflict resolved to a third content, and a card edit of a file main also changed; (4) the verdict is read from the merge's tree, not the worktree: a working-copy-only PASS, and a FAIL in the tree with a PASS in the worktree; (5) the binding rows: FAIL, `audited_sha` not the recorded evidence commit, two verdict files of one iteration number; (6) the record binds the merged tree: a record under the right name naming another tree, a record whose base is not the merge's first parent, a record with a non-zero exit; (7) one transaction: a fault after the receipt insert and before the commit leaves neither receipt nor edge; (8) existing receipts: a valid leader-issued receipt is used and not replaced, a stale one is not overwritten; a receipt for another card or run is refused by the unchanged gate; (9) the issued receipt carries issuer `auto` and the leader role marker; (10) revocation: after the leader revokes, the repeated queue archive is refused as receipt-absent and no automatic issuance recurs for that card and run; a lane session is refused the revocation; a leader-issued receipt is unaffected; (11) the reads are raw (AC-TU-016 rows i to k): a post-audit path that begins with a space and whose trimmed form lies under the reports directory issues nothing; a replacement ref for the evidence commit that hides a post-audit code change, and one for the merge commit that forges the tree the verdict is read from, each issue nothing; a graft file that rewrites the merge's first parent, and a shallow file, each issue nothing; and the evaluator's source names `gitReadRaw(` and no `gitRead(` (`TestAutoReceiptRawReads`); (12) the issuer `auto` is reserved (AC-TU-016 row h, AC-TU-019): `moai factory approve --issuer auto` and `--issuer AUTO` are refused with the issuer refusal and the store is unchanged, a leader receipt that replaces an automatic one carries the leader's issuer, and a leader-issued receipt for the same card and run survives a revocation (`TestAutoReceiptReservedIssuer`).
- **GREEN.** In `internal/homestate`, the planning step of edge T29 gains the automatic-issuance path: when the request carries the flag, the store holds no receipt for the (card, run) and no revocation exists, it evaluates the three facts against git objects named by the loaded row (`design.md` §D.5) and inserts the receipt through the transaction, then the existing `verifyTransitionApproval` runs; the re-measure facts come through a new callback on the request (`internal/homestate` cannot import `internal/factory`, `TestHomestateDoesNotImportFactory`, so the callback rides the request the way the existing `VerifyRemeasure` does), `factory.VerifyRemeasureBinding` in `internal/factory/integration_remeasure.go` (it reads the record, runs `ValidateRemeasureRecord`, then compares `tree` and `base` as written); a revocation record in the receipt store (the schema addition is this milestone's, additive); the leader-only revocation surface on `moai factory approve`. The three code comments that carry REQ-FCR-014 (`leader_approval_gate.go:282`, `leader_approval.go:9-13` and `:144-148`) are updated to name the REQ-TU-017 path. `IssueLeaderApproval` refuses the reserved issuer `auto` (trimmed, compared without regard to case) with `ErrApprovalIssuer` before any write, and the automatic insert is a separate unexported transaction-level function, so only the automatic path can write `auto`. The evaluator reads only through `gitReadRaw` (introduced in M1) and first refuses when `git rev-parse --git-path info/grafts` or `--git-path shallow` names an existing file (`design.md` §D.5).
- **Files (about 8 / 520 lines; splits at the issuance / revocation seam if it exceeds 400 at run time: issuance and its rows (1) to (9) and (11) first, revocation with the reserved issuer and row (12) as its own milestone M2c before M2b starts).** `internal/homestate/card_transition.go` (the T29 planning step), a new `leader_approval_auto.go`, `leader_approval.go` (store addition, comment), `internal/factory/integration_remeasure.go` (`VerifyRemeasureBinding`), `internal/cli/leader_approval_gate.go` (revocation flag, anchor text), tests in the three packages.
- **Packages.** `internal/homestate`, `internal/factory`, `internal/cli`.
- **Name-exact selection.** `TestAutoReceiptThreeFacts|TestAutoReceiptIssuer|TestAutoReceiptBindsMergedObjects|TestAutoReceiptRemeasureBinding|TestAutoReceiptAtomicWithEdge|TestAutoReceiptExistingReceipt|TestAutoReceiptRawReads|TestAutoReceiptReservedIssuer|TestAutoReceiptRevocation` (created at RED) plus the existing `TestLeaderApproval.*` family named in `leader_approval_test.go` and `leader_approval_gate_test.go` at RED time.

### M2b — Lane own-card done and the both-records close (Priority High; needs M1 and M2a)

- **RED.** (1) Lane env, card held by that lane, landing evidence of the closed form, the three facts present: `done` is admitted, the factory row ends `done` through edge T29 (the receipt, issuer `auto`, is issued inside that edge's transaction) and the queue row is archived, in that order, and the log row names the closing path and the issuer read back from the store; (2) the refusal matrix: another lane's card, no landing evidence, every other non-allowlisted verb, each leaves the queue file byte-identical; (3) one fact absent: nothing is issued, the close is refused with the gate's receipt-absent text, both records are unchanged; (4) a fault injected between the factory edge and the archive leaves the factory row `done` and the queue row open, exits non-zero naming both, and a repeat completes the archive without a second factory edge; (5) the leader's auto-done scan closes a landed `merged-local` card the same way; a `pushed` row with no leader-issued receipt is not closed (nothing auto-issued; both records byte-identical), and with a leader-issued receipt its queue row closes while the factory row stays byte-identical `pushed`, named in the log with the leader issuer (`TestTodoClosePushedRowUntouched`); a lane `done` that uses a leader-issued receipt logs the leader's label, not `auto` (`TestTodoCloseLogIssuer`); (6) the twelve-card fixture (ten with landing evidence, two without): exactly ten end `done` in both records; (7) mutant: a guard that admits `done` for any card fails row (2).
- **GREEN.** One admission function in the lane guard; the verb layer asks the existing landing predicate and refuses without it; the close helper shared by the lane path and the scan requests edge T29 (decider `audit`, the backlog uuid, the automatic-issuance flag, the `VerifyRemeasureBinding` callback) and then archives the queue row, in that order; the scan stops skipping a `merged-local` candidate whose receipt is not yet verifiable and requests the edge instead; the log row carries the new source label and the issuer read back from the receipt store; `todo undone` pairs its reversal with that label.
- **Files (about 7 / 520 lines; splits at the lane-path / scan-integration seam).** `internal/cli/todo.go` (guard), a new `todo_lane_done.go`, `todo_autodone.go`, `leader_approval_gate.go`, tests `todo_lane_done_test.go` and additions to the existing lane-guard and auto-done tests.
- **Packages.** `internal/cli`; `internal/factory` only if the evidence helper moves.
- **Name-exact selection.** `TestTodoLane.*|TestTodoCloseBothRecords|TestTodoCloseLogIssuer|TestTodoClosePushedRowUntouched|TestCloseFailureBetweenRecords|TestTenLandedCardsClose` (created at RED) plus the existing `TestAutoDone.*` and `TestLeaderApprovalGate.*` families (names read by `git grep -n '^func Test'` in `todo_autodone_test.go` and `leader_approval_gate_test.go` at RED time).

### M3a — The spec_id setter function (Priority High)

- **RED.** (1) The state table (`TestCardSpecSetStateSet`): fixtures in each of `picked`, `assigned`, `leased`, `plan`, `plan-audit`, `kickoff` with empty carriers: the function fills both carriers, one `card.fields` event, version plus one; fixtures in `run`, `sync`, `sync-audit`, `merge-ready`, `merging`, `merged-local`, `pushed`, and the terminal states: refused, neither carrier changes; (2) refusal matrix: a different id in either carrier, malformed id: neither carrier changes; (3) identical values in both carriers: no-op success in the six states and in `run`, refused in `done`; (4) a fault injected after the queue write and before the factory commit leaves the queue carrier set and the factory carrier empty, and a repeat with the same id completes the pair; a fault in the queue write leaves both carriers unchanged; (5) the completing repeat after the card moved on (`TestTodoSpecSetterRepeatAfterLease` at the verb level, the function-level rows here): to `leased` and `plan` (first-fill states), to `run` (outside the first-fill set, completing repeat admitted, different id refused), and to a terminal state (refused).
- **GREEN.** A set-once-per-carrier function in `internal/homestate` whose factory write transaction runs the queue write in its pre-commit step (the seam transitions already use for the dispatch record, `design.md` §D.7); its state rule is the first-fill set (`picked` through `kickoff`) for a first fill and any non-terminal state for a completing repeat.
- **Files (about 4 / 260 lines).** `internal/homestate/card_picked.go` (or a new `card_spec_set.go`), a small helper in `internal/factory`, tests in both packages.
- **Packages.** `internal/homestate`, `internal/factory`.
- **Name-exact selection.** `TestCardSpecSetOnce|TestCardSpecSetStateSet|TestCardSpecSetFailureBetweenCarriers` (created at RED).

### M3b — The setter verb (Priority High; needs M3a)

- **RED.** The verb writes through M3a's function for the leader; refuses lane sessions with the standing refusal text; malformed ids are refused before any write; `next --spec`, `factory assign --spec`, and the unpick clear keep their present tests green (names in `acceptance.md` AC-TU-022); kickoff's SPEC lookup passes after the setter, including the wedge fixture (a card leased with empty carriers, SPEC authored afterwards, setter while `leased`, then `plan-audit` admission and the audit kickoff resolve the directory; the control without the setter is refused by both); the lease-between-attempts rows at the verb level (`TestTodoSpecSetterRepeatAfterLease`).
- **GREEN.** Register `spec` on the `todo` tree (the 27th verb: the todo tree reads 27 until M4 registers the six stage verbs, then 33); the lane guard keeps it refused by not listing it.
- **Files (about 3 / 200 lines).** a new `internal/cli/todo_spec.go`, **`internal/cli/todo.go` (the single registration of the verb tree)**, tests `todo_spec_test.go`.
- **Packages.** `internal/cli`.
- **Name-exact selection.** `TestTodoSpecSetter|TestTodoSpecSetterLaneRefused|TestTodoSpecSetterPartialRepeat|TestTodoSpecSetterRepeatAfterLease|TestKickoffAfterSpecSetter` (created at RED) plus the six existing tests named in AC-TU-022.

### M4 — The six stage verbs on `moai todo` (Priority Medium)

- **Step 0 (already measured at plan time).** `reflect --rebuild-projection` writes private projection files (`research.md` §R.12), so `reflect` is admitted for lanes only without that flag; the other five write and stay refused.
- **RED.** (1) `TestGuidanceVerbsResolve`: every `moai todo <verb>` and `moai gtd <verb>` token in the guidance corpus resolves in its command tree (observed red: six lines of the workflow skill name verbs `moai todo` lacks; the token and file floors are in AC-TU-005); (2) the verb-surface parity test rewritten to "the `gtd` verb set equals the `todo` verb set, six stage verbs included", and `TestStageVerbsOnTodoTree` pinning the literal 33 names on each tree (the 26 present, `spec` from M3b, the six; the baseline 26 and 32 was measured at HEAD `3dc1052f6f`, `design.md` §D.4); (3) the five-stage path of the existing `TestGTDFiveStageCLIUsesSameSQLite` run again through the `todo` spelling with an equal store; (4) the lane classification test of REQ-TU-006.
- **GREEN.** Register the six constructors on the `todo` tree, delete the `gtd`-only registration line, classify each for the lane guard; the mistyped-verb guard reads the tree, so no separate list changes.
- **Files (about 6 / 260 lines).** `internal/cli/todo.go`, `gtd.go`, `gtd_answer.go`, tests `gtd_compat_test.go`, a new `todo_guidance_resolves_test.go`, `todo_surface_test.go`.
- **Packages.** `internal/cli`.
- **Name-exact selection.** `TestGTDAllTodoVerbsParity|TestGTDFiveStageCLIUsesSameSQLite|TestGTDCLIInputRefusalsDoNotFallThrough|TestStageVerbsOnTodoTree|TestGuidanceVerbsResolve|TestLaneClassifiesStageVerbs` (the first three exist at `gtd_compat_test.go:49`, `:124`, `:164`; the rest are created at RED).

### M5 — The warning alias (Priority Medium; needs M4)

- **RED.** `TestGTDAliasWarnsThenDelegates` over every verb (warning once on stderr, first line `Warning: moai gtd has been renamed to moai todo`, naming the `moai todo` form; stdout and exit equal to the `todo` form); `TestGTDAliasWarningStreams` (`--json` stdout parses, help requests carry no warning, and the six refusals of AC-TU-002, argument, flag, **required flag, flag group, and lane guard**, print the `todo` form's error with no warning; the positive control `gtd capture x --event e1` warns exactly once); the old parity assertion `gtdErr != todoErr` (`gtd_compat_test.go:83`) is rewritten to "the `todo` stderr equals the `gtd` stderr with the one warning block removed" (the pre-run hook's own notices come before the warning, so the comparison removes the block rather than assuming its position).
- **GREEN.** `NewGTDCommand` builds from the todo tree and, after adding nothing else, walks the tree once and wraps the `RunE` of every runnable command (the parent and every verb) with a function that writes the warning and then calls the original handler, so the warning is printed from cobra's run step, after argument, required-flag, and flag-group validation and after both pre-run hooks (`design.md` §D.2); a per-tree `warned` flag in the constructor's closure keeps it to one per invocation; the existing deprecation convention applies (stderr, `Warning:` prefix; `warnDeprecatedAgentModelFlags` is the precedent). No command in the two trees uses `Run`, so the wrap covers every handler.
- **Files (about 4 / 200 lines).** `internal/cli/gtd.go`, a small warning helper, tests `gtd_compat_test.go`, a new `gtd_alias_warn_test.go`.
- **Packages.** `internal/cli`.

### M6 — Slash command, skills, and the workflow document wording (Priority Medium; needs M5)

- **RED.** `TestTodoCanonicalSurfaceGolden` (inverts `TestGTDCanonicalSurfaceGolden`): `/moai:todo` dispatches `todo`; `/moai:gtd` dispatches `todo` and carries the rename notice; published skills equal the publisher output (count stays 17); the router skill routes `todo` canonical and `gtd` alias; the queue workflow document's header names `moai todo` canonical and its path is unchanged; the existing doc tests keep passing against the unchanged path.
- **GREEN.** Edit template sources (the two command files, the router skill rows, the workflow document header and body wording), regenerate the published skills with the project build.
- **Files (about 9 / 110 lines).** template commands (2) and their deployed copies, template and deployed router skill, template and deployed workflow document, `.agents/skills/moai-todo` and `moai-gtd` (generated), `internal/template/gtd_canonical_surface_test.go`.
- **Packages.** `internal/template`, `internal/cli`.
- **Name-exact selection.** `internal/template`: `TestTodoCanonicalSurfaceGolden`; `internal/cli`: `TestTodoSkillDocumentsHistoryVerb|TestTodoSkillDocumentsJevSource` (they exist at `todo_skill_doc_test.go:17` and `todo_skill_doc_parity_test.go:19`).

### M7 — The substitution script and the executable consumers (Priority Low; mechanical; needs M6)

- **RED.** (1) A fixture test of the script against a temp tree holding one file per substitution class and one per excluded class: expected output exact, excluded files byte-identical; (2) a second run changes nothing (AC-TU-009 covers both; `TestNoResidualGTDSpelling` is **not** created here: it would stay red over the unconverted corpus until M8 and AC-TU-024 forbids a skip, so it belongs to M8's RED, audit iteration 2 D25); (3) the consumer edits: argument vectors in the board mod, the cutover script default, the push-gate help line in `factory_card.go`, the `moai todo` help text.
- **GREEN.** Land `scripts/gtd-to-todo-substitute.sh` (`design.md` §D.3 holds the class table and the dry-run contract), its test, and the reviewed consumer edits.
- **Files (about 6 / 380 lines).** the script, its test, `mods/moai-board/hooks/data.ts`, `mods/moai-board/tests/support.ts`, `scripts/cutover-precheck.sh`, `internal/cli/factory_card.go`.
- **Gap carried.** The board mod's JavaScript tests are not run by any Go selection; the run phase records the runner used, or lists them as an unverified Gap.

### M8 — Apply the script (Priority Low; mechanical; needs M7)

- **Unit and size.** The unit is changed lines, not authored code: about 244 lines across 67 files (the 73 of E-01, minus 4 recorded fixtures, minus the two always-loaded `AGENTS.md` files), measured against E-01 and re-measured at pre-flight. The 10-file split rule does not apply; the review is dry-run path-set equality.
- **Batches by origin class (from pre-flight item 7).** M8a: template sources under `internal/template/templates/` and the `internal/cli` and `internal/template` test strings that are prose; M8b: deployed copies that have a template source follow the regenerated output, and the deployed-only copies (`.moai/workflows/gtd.md` has no template source, `git ls-files` lists only `README.md` on the template side) are converted by the script directly and named in the batch record; M8c: `docs-site/content` in four locales and the four READMEs (the `moai-gtd` pages stay; no redirect or menu change).
- **RED.** `TestNoResidualGTDSpelling` is created here (AC-TU-010): its planted-mutant fixture proves the check can go red, its corpus half is observed red against the unconverted tree (the paths of E-01 are residue until the apply), and its swept-count floor is asserted before the verdict.
- **Mechanism.** Dry-run first; the printed path set must equal the expected set (pre-flight list minus excluded classes); then apply; then re-run, expecting zero replacements and a green `TestNoResidualGTDSpelling`.
- **Docs-site obligations.** Four locales change in the same commit batch; the locale-parity baseline is not extended; the docs-site verification recipe (warning-free build, parity, URL scan) runs before the batch is accepted.

### M9 — AGENTS.md, rule statements, and the closing guards (Priority Low; last; needs M8, M2b, M1)

- **Content.** The `AGENTS.md` §11 row and its template mirror (`moai todo` canonical, `moai gtd` a warning alias, no "fully supported"); the hand-authored statements in the factory-dispatch rule family (the lane own-card `done` exception, the automatic receipt with issuer `auto`, the removal of T18 and the T19 reservation, edge T29, the leader's remaining role) in the template source and the deployed copy and the `.moai/policies` projection.
- **Guards.** `TestAgentsMDTodoRowWording` (live file and template mirror) and `TestCodexContractByteCeiling` stay green (AGENTS.md is 23,780 bytes against 24,576); the always-loaded growth statement is written for any edit above 1,000 bytes; the rule-loading budget measurement (count and bytes of always-loaded files before and after) is recorded.
- **RED.** The wording test fails on the current row (E-16); the residual-ledger test lists the `AGENTS.md` row as an allowed alias statement and fails before the edit.

## §F Affected packages and landing re-measure scope

Changed packages: `internal/cli`, `internal/homestate`, `internal/factory`, `internal/template`; `internal/config` through the AGENTS.md ceiling guard. Non-Go surfaces: `scripts/`, `mods/moai-board`, `docs-site`, templates and deployed mirrors.

Direct importers to include in the re-measure (non-test importers found by `git grep` of the import path at `026f48658f`; a `go list` reverse-dependency query was refused by the worktree guard, see `research.md` §R.1):

- of `internal/homestate`: `internal/cli`, `internal/codexwiring`, `internal/decision`, `internal/discovery`, `internal/escalation`, `internal/factory`, `internal/factorymsg`, `internal/hook` and `internal/hook/handoff`, `internal/web`.
- of `internal/factory`: `internal/cli`, `internal/discovery`, `internal/factorymsg`, `internal/graph`, `internal/hook`, `internal/statusline`, `internal/web`, and `cmd/t657-merge`.
- of `internal/template`: `internal/cli`, `internal/core/project`, `internal/hook`, `internal/mirrornotice`, `internal/settings`, `internal/userassets`, `internal/web`.

`internal/cli` is run by name-exact `-run` selections of the tests that name the changed files or identifiers (each milestone above lists its selection); the rest of `internal/cli` is carried to batch CI as an explicit Gap. The landing note states that Gap and the selections run.

## §G Template-First obligations

1. Source of truth: `internal/template/templates/**`. A deployed `.claude/`, `.codex/`, `.agents/`, or `.moai/` copy that has a template source follows it; none is edited first. A deployed-only file (`.moai/workflows/gtd.md`) has no template source and is edited directly by the script, named as such in the M8b batch record.
2. Commands render differently in the deployed tree than in the template (the template body carries the dual-harness dispatch block); mirror checks therefore compare rendered output, not raw bytes. The router skill is a byte-identical mirror guarded by the existing parity tests; the template and deployed `.claude/skills/moai/workflows/gtd.md` were byte-identical at `026f48658f` (`diff -q` printed nothing, `research.md` §R.10).
3. `.agents/skills/moai-todo` and `moai-gtd` carry "Generated by the MoAI command-to-skill publisher; regenerate, do not edit": regenerate, then confirm the drift check is green.
4. After the build, `git status --short` shows only the intended change set; a regenerated embedded catalog is part of that set and is staged by explicit pathspec.
5. Instruction files stay English; docs-site and README keep four locales in step.

## §H Risks

| Risk | Effect | Mitigation |
|------|--------|-----------|
| Substitution touches a persisted name or a recorded fixture | Data or test breakage | Class table with explicit exclusions; excluded files byte-identical in the script test; `TestNoResidualGTDSpelling` and the REQ-TU-012 pin |
| Argument-vector consumers use `'moai', 'gtd'`, which a `moai gtd` text pattern does not match | Alias warning reaches a parser, or the consumer breaks at removal | M7 reviewed edits plus a grep over the argument-vector form |
| The automatic receipt moves the issuance of a leader-role receipt into a path a lane can trigger | A card completes without the leader's read of the evidence | Three machine facts, all required, re-read at issuance; fail-closed; issuer `auto`; leader revocation; the gate's verification unchanged; leader-confirmed 10-18, conditional |
| T29 replaces T18's pair with a different guard | A reviewer reads it as a renamed row | AC-TU-018 asserts the behavior that distinguishes them (admits with a remote when the merge is on the landed ref; refuses on unreachable merge; no human-decider requirement) |
| Overlap with t1621 and t1639 on the transition-table files and the `pushed → done` edge (t1663's overlap is not measured) | Merge conflicts or contradictory edges | Pre-flight item 3; M1 rebases on whichever lands first; this SPEC declares no `pushed → done` edge |
| The receipt store gains a revocation record | A schema addition in a store other SPECs read | Additive and read-compatible; `TestAutoReceiptRevocation` plus the existing receipt tests |
| The automatic receipt reverses REQ-FCR-014 of the draft `SPEC-FACTORY-COMPLETION-RECOVERY-001` and three code comments that carry it (D23) | A reader of the draft or the comments sees a lane surface that mints | Named in `design.md` §D.5 and `research.md` §R.8; this SPEC wins for the one path, the draft is not edited, M2a updates the three comments and its pre-flight re-reads the draft's status and blocks on a change |
| The verdict and the re-measure record the facts read are authored by the lane's own tooling (the verdict file is a lane-committed file; the record is a file in the state directory) | The facts bind to the merged content but do not prove an independent auditor ran | The leader-confirmed trust shape (Q5); the compensating control is REQ-TU-025's revocation; a fourth fact (an audit receipt id) was considered and not taken (`design.md` §D.9) |
| Cards whose branches commit no verdict file (five of eight merges sampled, `research.md` §R.14) | They never reach the automatic path, so the ten-card criterion reads lower than the fixture suggests | Fail-closed by design; the live listing at sync names the first absent fact per card; the lane protocol's evidence-commit duty is not changed here |
| A completing setter repeat in a late card state bumps the factory version | A receipt bound to the old version goes stale | Rare (a half-written pair overtaken past `kickoff`); the leader re-issues, recorded in `design.md` §D.7 |
| The script test needs `sh` on every CI runner (not measured) | A Windows runner fails the test, or tempts a skip | The test fails loudly if `sh` is absent (no skip, AC-TU-024); M7 RED decides whether the script becomes a Go helper |
| AGENTS.md grows by a few bytes | Ceiling or budget breach | M9 measurement before and after; byte statement |
| The installed `moai` lags the tree | RED/GREEN observed on the wrong build | Build from the tree and invoke by path; record both commits |
| A card branch rewrites what git reads (a replacement ref, a graft file, or a shallow file in the shared git directory) | The three facts bind to forged content | `gitReadRaw` disables replacement objects; a graft or shallow file makes the evaluator issue nothing (AC-TU-016 rows j, k); residual: a write inside the evaluator's few milliseconds of reads, or a rewrite of the object store (`design.md` §D.5) |
| T29 without a remote and without configuration tests containment in a non-card local branch | A merge dropped from the integration branch but still held by another local branch reads as landed | Accepted for that case only; the exact test applies wherever a branch is configured; the receipt gate still binds the close (`design.md` §D.6, AC-TU-018) |
| The issuer `auto` is reserved, and a row written before this change may already carry it | An old manual `auto` row reads as automatic and is revocable | M2a pre-flight item 8 counts such rows and returns a blocker report on a non-zero count |

### Known debt (revision 4)

**KD-1, Known debt: the automatic-receipt facts depend on lanes committing their evidence under a gitignored path.** Fact 2 of REQ-TU-017 reads a tracked `sync-audit` verdict file from `.moai/reports/<card-id>/` in the merge commit's tree, and rule R2 compares paths under that directory. `.gitignore:235` ignores `.moai/reports/*` (measured: `git check-ignore -v --no-index .moai/reports/t1641/progress.md` printed `.gitignore:235:.moai/reports/*	.moai/reports/t1641/progress.md`); that this is the operator's policy of 2026-09-14 is the leader's statement, the order itself was not read here. Evidence reaches the tracked tree only where a lane force-adds it (`git add -f`), as the cards t1856, t1837, and t1812 did; five of eight merges sampled in `research.md` §R.14 committed none, and the iteration-3 audit counted 65 of 83 recent integration merges with no committed sync-audit file (its measurement, not repeated here). Consequence: where the evidence is not committed, fact 2 is absent, the automatic path simply does not issue (fail-closed), and the leader approves by hand; the live rate of the ten-card criterion is bounded by this. Decision: the SPEC deliberately does **not** fix this by requiring lanes to commit their reports (operator policy), does not remove fact 2, and no criterion claims that the automatic path works for uncommitted evidence (AC-TU-020 states the opposite). Any fix belongs to the operator's reports policy or to a successor SPEC that places the evidence somewhere tracked. Trigger to revisit: the sync-time live listing (REQ-TU-021) shows most cards stopping at fact 2. The same entry stands in `decision-index.md`.

## §I Self-verification deliverables (run phase)

Each item carries the command, the verbatim output, and the tree HEAD of this run: the AC matrix with RED-now and GREEN cells; the change-scope measurement from the absorbed-ref merge-base; the cross-platform build; changed-package coverage; lint; the substitution script's dry-run path set and second-run zero; the always-loaded budget measurement; and the Gaps list (the unrun `internal/cli` remainder, the JavaScript tests, the docs-site build if unavailable).

## §J Anti-patterns for this change

- Hand-editing the 67 files instead of running the script.
- Substituting `gtd` as a bare word (it appears in identifiers, table names, hex-prefixed ids, recorded fixtures, and the GTD methodology name).
- Writing the warning to stdout or printing it for help or for an argument-refused invocation.
- Adding a second handler under the `gtd` spelling.
- Renaming `workflows/gtd.md` (five rule files' `paths:` keys, a generated embed line, and a recorded fixture name it).
- Issuing the automatic receipt on two of the three facts, when the verdict file is only in the working tree, when the audited commit is merely an ancestor of the integration branch, or from a transaction other than the closing edge's.
- Printing the alias warning from a pre-run hook (it would precede required-flag and flag-group refusals).
- Reading the automatic receipt's facts through the trimming `gitRead`, with replacement objects enabled, or in a repository that carries a graft or shallow file.
- Resolving the push gate's integration branch from the card's own worktree (it never contains the merge commit, and `factory complete` refuses it as an integration branch).
- Letting the leader's manual issuance write the issuer `auto`.
- Editing the deployed `.claude/` copy before the template source.
- Treating a green run with an empty selection as a pass (a name-exact selector that matches no test).

## §K Cross-references

`spec.md` (requirements), `acceptance.md` (criteria), `design.md` (mechanisms), `research.md` (measurements and the evidence ledger), `decision-index.md` (Q1 to Q9). Related completed SPECs: `SPEC-TODO-LAND-AUTO-DONE-001`, `SPEC-GTD-CANON-BODY-001`, `SPEC-FACTORY-RECORD-001`, `SPEC-LOCAL-MAIN-FLOW-001`.
