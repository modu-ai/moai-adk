# decision-index.md — SPEC-TODO-UNIFY-001

`interview.decision_gate: on`. Nine decisions surfaced while assembling the plan for card t1641. This file carries no status field; the SPEC's lifecycle lives in `spec.md` alone. Label vocabulary: DECIDED, POLICY-COVERED, EVIDENCE-NEEDED, FOUNDER.

**Revision 2 (plan-audit iteration 1 FAIL, MP-7).** Iteration 1 left six product-level rows empty and three rows `DEFAULT-APPLIED` with their plan markers still open. In revision 2 the lane resolved every row through the decision ladder of `.claude/rules/moai/workflow/auto-semantics.md` §6 (① disk evidence, ② the decision board, ③ the audit cross, ⑤ self-judgment; ④ `jev_ask` is off here) and recorded each as a one-line decision record per §10. Nothing here is keep-set (environment-impossible, operator-held, irreversible external-shared operation), so none required the operator. Each row below states the decision, the evidence behind it with the command or file and line, and the alternate that was rejected. The decision is also written into the requirement it affects, as a testable condition (`spec.md` §3, `plan.md` §B). The board read this run (`moai decision read --scope card:t1641`, `board=ok unparseable=0 records=108`) holds no record that rules on Q1 to Q9; the one standing record that bears on Q9 is cited there.

No row is labelled DECIDED or POLICY-COVERED: the places that decide these questions are the live queue card (not in the committed tree) and the untracked 3.2 plan and redesign reports under `.moai/reports/`, which are never authority. The label stays FOUNDER and the verdict line records that the lane, not the operator, decided. Committed material named under `Authority anchor` is related context, not authority.

Verdict forms used: `LANE-DECIDED <UTC> manager-spec (...)` for Q1 to Q4 and Q6 to Q9; Q5 carries the one trust-model decision and is marked as the leader confirmed it.

**Revision 3 (plan-audit iteration 2 FAIL, receipt `rcpt-717ae7d0bd489fdeced99ba8`).** No question is reopened and no row's label or verdict changes. Q5's shape (three machine facts, issuer `auto`, revocable, fail-closed) is exactly what the leader confirmed and its marker text is unchanged; the audit tightened how facts 2 and 3 are bound (D16, D17), which the Q5 decision line now states, without adding or removing a fact. Q8's decision gained the state set and the completing-repeat rule (D20, D21), and Q9 stays leader-only. The board read this round (`moai decision read --scope card:t1641`, `board=ok unparseable=0 records=108`) holds no record that rules on Q1 to Q9; the record `d-20261010T001732Z-d4e8` bears on Q8 and Q9 only and is the evidence for the setter's state set (`research.md` §R.15).

**Revision 4 (post-ceiling correction under the leader's standing rule C; plan-audit iteration 3 FAIL 0.86, receipt `rcpt-9557497ba4bcc4b92912703c`).** No question is reopened and no row's label, verdict, or marker changes; Q5's shape (three machine facts, issuer `auto`, revocable, fail-closed) and its operator-verdict marker are exactly what the leader confirmed and are not edited. The audit's D30 only makes "issuer `auto`, distinguishable" true by construction (the leader's manual issuance refuses that value), and D31 only makes "read from git objects" true by construction (raw reads, replacement objects off, graft and shallow refused); neither adds or removes a fact. D29 changes Q6's replacement edge T29 only in what it tests without a remote (`design.md` §D.6). The one new item is a known debt, below, which is not a decision row and carries no Q number.

## Q1: What does "clean up the 6 gtd subverbs" mean — absorb all six into `moai todo`, absorb the used ones and delete the rest, or leave them under `moai gtd`?

- Label: FOUNDER
- Class: product-level
- Authority anchor: none. Related committed context: `AGENTS.md` §11; `.claude/skills/moai/workflows/gtd.md` lines 150, 161, 173, 185, 194, 211 (the guidance names the six as `moai todo <verb>`).
- Why unresolved at authoring: card t1590 fixes "register the six on `moai todo`"; card t1602 says the GTD-stage verbs "stay under `moai gtd`"; the redesign's D3 sentence says absorb the used ones and delete the rest.
- Decision: **(a) absorb all six; delete none.** Written into REQ-TU-004, REQ-TU-006, AC-TU-004.
- Evidence: reading (c) is already excluded by the card's completion criterion "gtd warns, then delegates" and by removal after one release. Between (a) and (b): the deletion set of (b) would have to come from measured use, and the tree cannot measure it. `git grep -n -E 'moai (gtd|todo) (capture|clarify|organize|reflect|engage|answer)'` outside docs and templates printed only a comment in `gtd_answer.go:15`, a test header `gtd_answer_test.go:10`, and a comment `todo_transition_stamps_test.go:195`: no code caller, but no in-tree caller is not evidence of no use by operators or lanes, and the tree holds no usage telemetry. The CLI path is pinned by tests: `TestGTDFiveStageCLIUsesSameSQLite` (`gtd_compat_test.go:124`), `TestGTDCLIInputRefusalsDoNotFallThrough` (`:164`), three `TestGTDAnswer*` tests (`gtd_answer_test.go:13, 43, 52`), and `newGTDEngageCmd` is used at `todo_issuance_test.go:600`. The guidance already tells readers to run all six as `moai todo <verb>` (E-03). `goal.go` calls two of the library functions directly, which is unaffected. The published-default rule ranks "preserves current behavior" first, and (a) deletes no user-facing behavior.
- Alternate rejected: (b) absorb the used ones and delete the rest: removes a documented user-facing feature (four READMEs, four docs-site locales) on an assumption the tree cannot confirm (an unobserved defect-style claim, `verification-claim-integrity.md` §1.1 surface 3).
- Default (policy rule): (a) (rule: preserves current behavior)
- decision record: decided_by=manager-spec (plan owner, card t1641) evidence_refs=research.md §R.2 E-03 E-04, internal/cli/gtd_compat_test.go:124,164, gtd_answer_test.go:13-52, todo_issuance_test.go:600, gtd.md:150-211 ladder_path=①⑤
- Operator verdict: LANE-DECIDED 2026-10-18T07:11:14Z manager-spec (plan owner, card t1641, decision ladder ①⑤)

## Q2: Does the warning name a literal removal release or say "the next release"?

- Label: FOUNDER
- Class: implementation-level
- Authority anchor: none.
- Why unresolved at authoring: the card says "one release of grace" without a number.
- Decision: **the warning states removal "one release after the release that first carries the warning"; no literal release number is held anywhere.** Written into REQ-TU-002, AC-TU-001.
- Evidence: the card text (`research.md` §R.3: "경고 별칭(한 릴리스 유예)") fixes the grace as one release and names none; no committed artifact read for this plan names the removal release (the plan read the card, the redesign extract, and the SPEC set; no exhaustive search was run). A literal number would have to be kept current and would be false the moment the release plan moves. The chosen phrase is true whatever the number, and AC-TU-001 can test it (non-empty constant contained in the warning).
- Alternate rejected: (a) a literal release taken from the release plan: needs a number no committed artifact supplies, and goes stale. (The earlier "the next release" wording was also dropped: it is false if the warning ships one release earlier than the removal.)
- Default (policy rule): smaller user-visible surface
- decision record: decided_by=manager-spec (plan owner, card t1641) evidence_refs=research.md §R.3 card text, acceptance.md AC-TU-001 ladder_path=①⑤
- Operator verdict: LANE-DECIDED 2026-10-18T07:11:14Z manager-spec (plan owner, card t1641, decision ladder ①⑤; replaces DEFAULT-APPLIED of iteration 1)

## Q3: Is "rename gtd to todo everywhere" the user-facing spelling only, or also the internals?

- Label: FOUNDER
- Class: product-level
- Authority anchor: none. Related committed context: the doc comment on `NewGTDCommand` (`internal/cli/gtd.go:65-66`): "Internal package names and the on-disk todo/backlog.db location deliberately remain unchanged."
- Why unresolved at authoring: the operator's order reads "rename gtd to todo everywhere"; renaming persisted formats needs a migration and touches about 47 files.
- Decision: **user-facing spelling only; Go identifiers, `gtd_*` tables, the `gtd-<16 hex>` item id, `.moai/gtd/answers`, and the path `workflows/gtd.md` stay.** Written into REQ-TU-007, REQ-TU-012, AC-TU-011.
- Evidence: the card's scope list names "문서 치환·슬래시 정본" (documents and the slash entry), and none of t1641, t1602, or t1590 asks for a table, id, or file rename (`research.md` §R.3). The code comment above states the internals deliberately remain. The workflow-document path is an identifier wired into five rule files' `paths:` keys, `embed_manifest_gen.go:516`, a recorded fixture, and seven test files (`research.md` §R.10), so renaming it is an internal rename; iteration 1's plan to rename it is withdrawn.
- Alternate rejected: (b) everything: changes persisted formats (migration, dual-read of the old item id, about 47 files) and is a separate SPEC; nothing in M1 to M9 blocks it later.
- Default (policy rule): preserves current behavior
- decision record: decided_by=manager-spec (plan owner, card t1641) evidence_refs=internal/cli/gtd.go:65-66, research.md §R.3 §R.6 §R.10 ladder_path=①⑤
- Operator verdict: LANE-DECIDED 2026-10-18T07:11:14Z manager-spec (plan owner, card t1641, decision ladder ①⑤)

## Q4: Is the landing-to-done machine transition with three observers, and automatic receipt issuing, in this card's scope or card t1639's?

- Label: FOUNDER
- Class: product-level
- Authority anchor: none. Related committed context: `SPEC-FACTORY-RECORD-REPAIR-001` spec.md §G and lines 134-136 (the `pushed → done` item moved to 3.2-2-2, t1639).
- Why unresolved at authoring: the master plan lists the machine transition and automatic receipts under t1639 and only "lane own-card done" and "T18/T19 removal" under t1641, while t1641's completion criterion reads "done is automatic for the most recent 10 landed cards".
- Decision: **this SPEC delivers the lane own-card path, the automatic receipt issued only inside the two close paths (the lane's `done` and the leader's auto-done scan), and the replacement edge T29; it delivers no observer, no hook, no general issuance at observation time, and no `pushed → done` edge.** Written into REQ-TU-017, REQ-TU-020, the Out of Scope list, AC-TU-014, AC-TU-020.
- Evidence: the card's range list names exactly "레인 자기 카드 done 예외, T18/T19 제거" (`research.md` §R.3); the observers are t1639's (`research.md` §R.8; RECORD-REPAIR spec.md:134-136). The criterion is reachable without observers because the two close paths already exist (the lane's `done` and the leader's scan) and the only thing that blocked them for a factory-linked card was the missing receipt, which Q5 supplies; the factory edge is T29 (Q6). The overlap with t1639 is stated, not hidden: t1639's general mechanism remains out.
- Alternate rejected: (b) also deliver the SessionEnd and `factory complete` observers: duplicates t1639 and adds a hook under the advisory-check discipline.
- Default (policy rule): smaller user-visible surface
- decision record: decided_by=manager-spec (plan owner, card t1641) evidence_refs=research.md §R.3 §R.8, RECORD-REPAIR spec.md:134-136, internal/cli/leader_approval_gate.go ladder_path=①⑤
- Operator verdict: LANE-DECIDED 2026-10-18T07:11:14Z manager-spec (plan owner, card t1641, decision ladder ①⑤)

## Q5: What satisfies the leader-approval receipt gate when a lane closes its own card?

- Label: FOUNDER
- Class: product-level
- Authority anchor: none. Related committed context: `internal/homestate/leader_approval.go:116-142` (what the gate verifies), `internal/cli/leader_approval_gate.go:275-375` (issuance is leader-only); draft `SPEC-FACTORY-COMPLETION-RECOVERY-001` REQ-FCR-002.
- Why unresolved at authoring: the gate binds every close of a factory-linked card and only the leader can mint a receipt. Iteration 1 ranked option (c), a leader-issued receipt, first; the audit showed (c) makes the card's criterion "done is automatic for the most recent ten landed cards" unreachable and AC-TU-020 unmet by construction.
- Decision: **the gate stays and is proxied by machine verification: a receipt is issued automatically only when all three machine facts exist — (1) the card is in `merged-local`; (2) a committed `sync-audit.md` verdict of the PASS family, read from the recorded merge commit's tree, whose `audited_sha` is the recorded evidence commit and after which the card adds nothing to the merge except its own `.moai/reports/<card-id>/` files, main absorbed after the audit excepted (revision 3 replaces "an ancestor of the integration branch", which let a stale PASS license later unaudited code); (3) a valid landing re-measure record that names the merged tree and the merge's first parent in its own fields (revision 3: the older callback never compared them). All three are read from git objects the card's factory row names and are checked, the receipt written, and the edge committed in one factory transaction. If any one is missing nothing is issued and the card stays on the leader's manual approval (fail-closed). The receipt records issuer `auto`, a value the leader's manual issuance refuses (revision 4), so it is distinguishable from a leader-issued one, and the leader can revoke it.** Written into REQ-TU-017, REQ-TU-025, AC-TU-016, AC-TU-019 (the mutant that issues on two of the three fails AC-TU-016).
- Evidence: the gate's verification (`VerifyBinding`: leader role marker, issuer is not the performing owner, card uuid, run, version, equal non-empty evidence hash) is unchanged, so the gate's form is intact; only the issuer of the receipt differs, and `auto` never equals a lane's owner label (`CanonicalApprovalOwnerLabel`, `leader_approval.go:91-105`). The three facts are exactly what the leader reads before approving and each is already machine-checkable in the tree: state from the factory row, the verdict reader (`card_evidence_readers.go:20-21`, `PASS|PASS-WITH-DEBT|FAIL` and `audited_sha`), and the merge gate's re-measure callback (`card_evidence_readers.go:290-355`). The coordinator's addendum of 10-18 fixed this shape after the leader confirmed the direction.
- Alternate rejected: (a) landing evidence alone substitutes for the receipt: removes the gate's form; (c) a leader-issued receipt only: leaves the criterion unreachable without a leader act.
- Default (policy rule): none applied; the leader's answer governs.
- decision record: decided_by=manager-spec (plan owner, card t1641) on the leader's confirmation relayed by the coordinator 10-18 evidence_refs=research.md §R.13, internal/homestate/leader_approval.go:116-142, card_evidence_readers.go:20-21,290-355, plan-audit.md D2 D3 ladder_path=①③⑤ plus leader confirmation
- Operator verdict: lane-judgment, leader-confirmed 10-18 (conditional)

## Q6: Are T18 and the T19 reservation removed now, given the GitHub Flow transition is not complete in this repository?

- Label: FOUNDER
- Class: product-level
- Authority anchor: none. Related committed context: `.moai/config/sections/git-strategy.yaml` (`mode: manual`, `manual.workflow: git-flow`); `SPEC-GITHUB-FLOW-DEFAULT-001` `status: in-progress`; draft `SPEC-FACTORY-COMPLETION-RECOVERY-001` REQ-FCR-010; `internal/homestate/card_evidence.go:64` `PushGateTarget`.
- Why unresolved at authoring: the decision text removes the git-flow edges "with the completion of the GitHub Flow transition", which has not happened here; T18 is today the only completion edge for a card whose repository has no remote.
- Decision: **(a) remove T18 and the T19 reservation now, and declare the replacement in the same change: edge T29, `merged-local → done`, guarded by re-verified merge reachability from the landing ref plus the receipt gate; the push gate keeps requesting `done` for a no-remote repository, which T29 declares.** Written into REQ-TU-018, REQ-TU-019, AC-TU-017, AC-TU-018; M1's RED order puts the T29 tests first.
- Evidence: the card's scope says "T18/T19 제거" and the operator instruction of 10-17 ("3.3 이월 해제·지금 진행") releases it; iteration 1's default (c) left the card's removal item undelivered (plan-audit D2). The replacement is needed because after the removal there is otherwise no declared edge into `done` for a merged-local card (T20 needs `ci-green`, which nothing reaches; `isReservedEdge` at `card_transition.go:207`, T20 row at `:158`). The pair-count pin does not move: the pin runs against a fixture with a remote and no receipt, where both pairs are refused before and after (`fr_transition_test.go:129`, `:166`; static reading). T29 is not T18 renamed: T18 required the human decider, refused while a remote exists, and never re-checked the merge (`card_transition.go:722-747`); T29 has no remote condition, re-verifies reachability, and keeps the receipt gate, and AC-TU-018 asserts those differences.
- Alternate rejected: (b) remove only the T19 reservation: contradicts the decision text for T18; (c) defer until `SPEC-GITHUB-FLOW-DEFAULT-001` closes: leaves the card's removal item undelivered.
- Default (policy rule): none applied; the card text governs.
- decision record: decided_by=manager-spec (plan owner, card t1641) evidence_refs=internal/homestate/card_transition.go:157,203-209,333,722-747, fr_transition_test.go:127-183, card_evidence.go:64-73, plan-audit.md D2 D4, research.md §R.3 ladder_path=①③⑤
- Operator verdict: LANE-DECIDED 2026-10-18T07:11:14Z manager-spec (plan owner, card t1641, decision ladder ①③⑤)

## Q7: Which record carries "done" in the ten-card criterion, and what counts as "landed"?

- Label: FOUNDER
- Class: product-level
- Authority anchor: none. Related committed context: `SPEC-TODO-LAND-AUTO-DONE-001` (the auto-done scan reads the resolved landed ref and closes the queue row); `SPEC-LOCAL-MAIN-FLOW-001` (in the local-main flow the merge is on local `main` while `origin/main` advances by batch).
- Why unresolved at authoring: the queue (archive) and the factory card row each have a `done`, closed by different paths today; iteration 1 ranked "queue only" first, which contradicted AC-TU-014 and AC-TU-020.
- Decision: **both records end `done`; "landed" is reachability of the card's merge commit from the resolved landed ref (closed evidence form i); the factory edge is written first and the queue archive second, a failure between them exits non-zero naming both records, and a repeat completes the archive without a second factory edge.** Written into REQ-TU-020, AC-TU-014, AC-TU-023.
- Evidence: a card done in one record and open in the other is the observed wedge (`research.md` §R.3, RECORD-REPAIR spec.md:47: 75 of 76 done rows were written outside the transition API, the card body's claim, not measured here); "queue only" leaves it possible. The two stores commit separately (queue write through `mutateLocked`, `backlog_store.go:1001-1029`; factory row in its own transaction), so an order is required: factory first, because the reverse leaves an archived queue row and a factory row no live card reaches, which no retry can repair, while factory-first leaves an open queue row that the repeat or the scan closes under the receipt T29 re-stamps (`card_transition.go:744`, `:786`). A factory row in `pushed` is left untouched (that edge is t1639's, Q4).
- Alternate rejected: queue only (b) or factory row only (c): leave the one-record-done wedge possible; landed (ii) on local `main` only: weaker than the auto-done scan's closed evidence set, which this SPEC preserves.
- Default (policy rule): none applied; the criterion text and AC-TU-014/020 govern.
- decision record: decided_by=manager-spec (plan owner, card t1641) evidence_refs=internal/factory/backlog_store.go:1001-1029, internal/homestate/card_transition.go:744,786, plan-audit.md D2 D5, RECORD-REPAIR spec.md:47 ladder_path=①③⑤
- Operator verdict: LANE-DECIDED 2026-10-18T07:11:14Z manager-spec (plan owner, card t1641, decision ladder ①③⑤)

## Q8: What shape does the post-pick spec_id setter take?

- Label: FOUNDER
- Class: implementation-level
- Authority anchor: none. Related committed context: `internal/homestate/card_picked.go:185` (the refusal "its hints and pointers change only while picked"); `internal/cli/contract_report.go` `cardToSpecID` (reads the queue spec_id).
- Why unresolved at authoring: three shapes exist: a new verb writing both carriers (a); widening `factory assign --spec` to set-if-empty, which fixes the factory row only (b); letting `next <n> --spec` re-enter a picked card, which removes the positive pick gate (c).
- Decision: **(a) a new `moai todo spec <n> <SPEC-ID>` verb writing both carriers, set-once per carrier, under the queue lock, with the queue write inside the factory write transaction's pre-commit step; a first fill from `picked` through `kickoff` (the states of the observed leased wedge, up to the last guard that reads the identifier), and a failure leaves a state the same verb completes on a repeat in any non-terminal state.** Written into REQ-TU-022, REQ-TU-023, AC-TU-021, AC-TU-025.
- Evidence: kickoff reads the factory spec_id and `contract_report` reads the queue spec_id (`design.md` §D.7), so only a two-carrier writer removes the double repair; (b) fixes kickoff only; (c) removes a safety gate. Atomicity (plan-audit D5): the two stores commit separately (`backlog_store.go:1001-1029` for the queue), so the SPEC states the order and recovery (REQ-TU-022) and tests both fault points (AC-TU-025); the seam it uses is the pre-commit step transitions already use (`card_transition.go:286`).
- Alternate rejected: (b) widen `factory assign --spec`: leaves the queue carrier empty; (c) re-enter a picked card: removes a gate.
- Default (policy rule): preserves current behavior (every existing verb is unchanged)
- decision record: decided_by=manager-spec (plan owner, card t1641) evidence_refs=internal/homestate/card_picked.go:185, internal/factory/backlog_store.go:1001-1029, card_transition.go:286, plan-audit.md D5 ladder_path=①③⑤
- Operator verdict: LANE-DECIDED 2026-10-18T07:11:14Z manager-spec (plan owner, card t1641, decision ladder ①③⑤; replaces DEFAULT-APPLIED of iteration 1)

## Q9: Who may call the spec_id setter — the leader only, or a lane for its own card?

- Label: FOUNDER
- Class: implementation-level
- Authority anchor: none. Related committed context: `todoLaneReadOnlyVerbs` in `internal/cli/todo.go:405-413` (a lane may run seven read-only verbs; every other verb is refused).
- Why unresolved at authoring: a lane learns the SPEC id at plan phase, so a leader-only setter forces a message to the leader; admitting a lane widens the lane allowlist a second time.
- Decision: **(a) leader only; a lane session is refused with the standing text.** Written into REQ-TU-022, REQ-TU-023, AC-TU-021.
- Evidence: standing board record `d-20261010T001732Z-d4e8` (read this run, `moai decision read --scope card:t1641`): "The leader stamps spec_id on a leased card row once the lane names the SPEC directory it authored ... A lane that authors a SPEC after leasing sends the leader the SPEC id in one line before it needs the kickoff edge." It also says "The missing lane-side surface is repair card t1621 scope"; t1621's narrowed SPEC lists that item out of scope (`research.md` §R.8), so no lane-side surface is planned anywhere. This SPEC does not widen the lane allowlist twice for a path the standing ruling already routes through the leader. The leader-side verb replaces the leader's direct database edit.
- Alternate rejected: (b) a lane, for its own card, only while empty: widens the lane authority a second time beside M2's own-card done; revisit as part of whichever card takes the lane-side surface.
- Default (policy rule): preserves current behavior
- decision record: decided_by=manager-spec (plan owner, card t1641) evidence_refs=board d-20261010T001732Z-d4e8, internal/cli/todo.go:405-413, research.md §R.8 ladder_path=①②⑤
- Operator verdict: LANE-DECIDED 2026-10-18T07:11:14Z manager-spec (plan owner, card t1641, decision ladder ①②⑤; replaces DEFAULT-APPLIED of iteration 1)

## Known debt (revision 4; not a decision row)

**KD-1, Known debt: the automatic-receipt facts depend on lanes committing their evidence under a gitignored path.** Fact 2 of REQ-TU-017 reads a tracked `sync-audit` verdict file from `.moai/reports/<card-id>/` in the merge commit's tree, and rule R2 compares the paths under that directory. `.gitignore:235` ignores `.moai/reports/*` (`git check-ignore -v --no-index .moai/reports/t1641/progress.md` printed `.gitignore:235:.moai/reports/*`); the operator's policy of 2026-09-14 behind it is the leader's statement, the order itself was not read. The premise that a committed verdict exists in the merge tree therefore holds only for lanes that force-add their evidence, as t1856, t1837, and t1812 did (`research.md` §R.14: five of eight sampled merges committed none; the iteration-3 audit counted 65 of 83 recent integration merges with no committed sync-audit file, its own measurement).

- Consequence: where the evidence is not committed, fact 2 is absent, the automatic path simply does not issue (fail-closed), and the leader approves by hand. The ten-card criterion of REQ-TU-021 is a mechanism proof plus a live report, not a live pass, until lanes commit evidence.
- Decision: the SPEC deliberately does **not** fix this by requiring lanes to commit their reports (operator policy), does not remove fact 2, and no requirement or criterion claims that the automatic path works for uncommitted evidence (AC-TU-020 says the opposite).
- Class: product-level (it shapes what the shipped close path does in practice), recorded as a debt and not as a question because no option is open: the operator's reports policy is fixed input.
- Owner and trigger: the operator's reports policy, or a successor SPEC that places the evidence somewhere tracked; revisit when the sync-time listing (REQ-TU-021) shows most cards stopping at fact 2.
- The same entry stands in `plan.md` §H.
