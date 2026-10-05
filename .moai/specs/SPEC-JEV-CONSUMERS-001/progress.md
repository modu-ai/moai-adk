# SPEC-JEV-CONSUMERS-001 — Progress

Card: t1020 · Tier L · plan-phase artifacts authored 2026-09-20; revised 2026-09-20 (revision 0.2.0). Split from `SPEC-JEV-INTEGRATION-001` on the M4+M5+M6 seam.

## §E.1 Plan-phase Audit-Ready Signal

| Item | State |
|---|---|
| SPEC ID regex check | `PASS` (executed) |
| SPEC ID collision | none |
| Tier | L — REQ 15 / ceiling 25; AC 15 / ceiling 25 |
| Artifact set | spec.md · plan.md · acceptance.md · design.md · research.md · progress.md |
| Requirements | 15 (`REQ-JEVN-001` … `REQ-JEVN-015`) |
| Acceptance criteria | 15 (`AC-JEVN-001` … `AC-JEVN-015`) |
| Revision | 0.2.0 — answers the iteration-1 plan-audit verdict |
| Audit verdict answered | `.moai/reports/t1020/plan-audit-SPEC-JEV-CONSUMERS-001.md` (iteration 1/3, FAIL 0.64 against the Tier L 0.85 threshold) |
| Coordinate baseline | every coordinate cited across the six artifacts re-verified at HEAD `7e1ed63b9` |
| Predecessor | `SPEC-JEV-OPTIN-MEASURE-001` |
| Successor | `SPEC-JEV-GOAL-DIST-001` |
| Status transition | (none) → draft (unchanged by this revision) |

### What revision 0.2.0 changed

- **N1 (dedup precedence) is SETTLED**, confirmed by the operator exactly as proposed, and is recorded as a decision in `plan.md` §B1 and `design.md` §2 rather than as an open question. It leaves the Kickoff-gate carry list.
- `REQ-JEVN-006` restated: the rule's two halves are separated, half (a) is shown to require a **new source-agnostic unordered predicate** (`HasFindingForPairAnySource`) that M4 now budgets, and the claim that `AppendFindingOnce` / `SamePairAs` enforce it is withdrawn as false.
- `REQ-JEVN-015` added: the disposition for a consumer whose gate **cannot be run**, distinct from one whose gate ran and failed. `AC-JEVN-015` asserts the distinction.
- `REQ-JEVN-001` scoped explicitly to card admission, excluding the `moai todo analyze` re-sweep; `AC-JEVN-013` asserts it.
- `AC-JEVN-012` added (write-path `Source: agent` absence with a positive control), `AC-JEVN-014` added (the ranked-signal presentation half of `REQ-JEVN-013`).
- `AC-JEVN-003` widened to all four source combinations in both arrival directions; `AC-JEVN-008` / `AC-JEVN-011` given call-path-exists preconditions; `AC-JEVN-007` / `AC-JEVN-010` gated on the fitted threshold's existence with a stated blocked disposition; the Definition of Done scoped per consumer; every AC now names its REQ ids.
- Three write-path coordinates corrected as **miscitations, not drift** — they resolved at neither HEAD `7e1ed63b9` nor the declared baseline `fd75cf692`, and no commit touched the files between the two trees.

### Open questions carried to the Implementation Kickoff Approval gate

- **N2 — which code path hosts Consumer A, and what verification surface covers Consumer B.** OPEN and **blocking M5 and M6**, which are declared blocked in `plan.md` §F. Not covered by the 2026-09-20 operator decisions. This revision deliberately does not name a host for either consumer: an unmeasured host would be exactly the unverified-premise defect this chain is bounded against. N2 replaces N1 on this list.
- **Q3 — labelled-set size per consumer.** OPEN, owned by `SPEC-JEV-OPTIN-MEASURE-001`.

Not an open question, recorded here because it shapes what run-phase will produce: with live measurement excluded from this batch, `Report.Verdict()` (`internal/jevmeasure/measure.go:193`) withholds every consumer, so M4's expected terminal state is the `REQ-JEVN-015` recorded decision rather than a shipped consumer.

## §E.2 Run-phase Evidence

Measured on this tree, `WT-jev-consumers`, unless a row names another baseline. M4 (Consumer C) landed at `e3efef5c3`; the row outputs for Consumer C's criteria below were re-measured in THIS run against the absorbed tree, so every row is a this-run observation rather than a carry-over.

### M6 — Consumer B (skill suggestion): the gate-unrun record

**State: `REQ-JEVN-016` gate-unrun.** The measurement gate is runnable and has **not** been run. Consumer B's implementation is present in the tree and unreachable at the shipped default. **What the gate still needs, and who owns it:** a labelled set for the skill-suggestion task, sized per the base rate the predecessor's harness measures — open question Q3, owned by `SPEC-JEV-OPTIN-MEASURE-001` — and a live measurement pass, excluded from this batch as for M4 (no TypeSafe credential in this tree). **This record cites no measurement, because none has been taken.** It is not a gate-not-runnable record: nothing here claims a withholding mechanism closed the gate, because nothing did — the owed work survives this milestone.

**Verification surface (the N2 answer this milestone carries).** Consumer B's host is the `/moai` intent router — prose the Go quality gate cannot reach. Following the foreman precedent (`internal/kanban/foreman_queue_watch_test.go`), the verification surface extracts the suggestion-protocol block VERBATIM from both skill copies (`internal/cli/jev_skill_suggest_test.go` `suggestionSkillPaths`), asserts the two byte-identical, validates the block against the protocol contract, executes the mechanical anchor the block names (`moai jev-suggest`, `internal/cli/jev_skill_suggest.go`), and carries an in-suite falsification control: four pinned mutants (gate framing removed, authority clause inverted, suppression rule removed, anchor command removed) are each asserted REJECTED.

**E1 — AC matrix.**

| AC | Status | Verification command (test) | Actual output (this run, this tree) |
|---|---|---|---|
| AC-JEVN-009 | PASS | `go test ./internal/cli/ -run TestJevSkillSuggest_TwoRequests` | `--- PASS` — exactly 2 wire requests; req 1 = Noul + 5 wide-rank scores over short descriptions only; req 2 = 3 rerank scores over exactly the top-3 under fuller text |
| AC-JEVN-010 | **BLOCKED-ON-THRESHOLD** | `TestJevSkillSuggest_SuppressionMechanism` (mechanism only) | **No fitted threshold exists** (open question Q3, owned by `SPEC-JEV-OPTIN-MEASURE-001`) — the criterion is not evaluable and is **not** recorded as failed or passed. The suppression branch itself is mechanism-tested with a synthetic threshold (negative-above suppresses, negative-below and positive never suppress, nil threshold never suppresses); that is a branch test, not the criterion |
| AC-JEVN-011 | PASS | `TestJevSkillSuggest_RouterAuthorityUntouched` | `--- PASS` — paired enabled/disabled runs over the same input; precondition asserted (enabled stub logged ≥1); disabled run = zero requests + zero-value signal; presented value byte-identical across enabled runs and contract-clean (no selection/dispatch field). Prose half: the "keeps its selection authority" clause is a validated contract token |
| AC-JEVN-014 | PASS | `TestJevSkillSuggest_RankSignalShape` + `TestJevSuggestCmd_ArgumentValidationAndHappyPath` | `--- PASS` — presented object carries only contract keys, candidates carry consecutive 1-based ranks; positive control: a value carrying a `selection` field is REJECTED (plus malformed-shape rejections in `TestInspectJevRankSignal_RejectsMalformedShapes`, and wire-level key checks on the command's stdout) |
| AC-JEVN-016 (5 conditions) | PASS | `TestJevGateUnrun_DefaultOffEverywhere` · `TestJevSkillSuggestEntry_GateOffConstructsNothing` · `TestJevSuggestCmd_GateOffOutputIdenticalApartFromNotice` · `TestJevGateUnrun_NotPresentedAsAvailable` · this record | all five `--- PASS` — (1) compiled default false + template sweep non-empty, every `enabled:` under `jev:` reads `false`; (2) gate off → zero requests AND zero client constructions, positive control gate-on → ≥1; (3) gate-off output byte-identical to the absent-call-site run apart from exactly the one notice line; (4) `jev-suggest` absent from CHANGELOG/READMEs with the `moai doctor` positive control firing; (5) this record |

**AC-JEVN-016 condition 1 positive control — recorded adaptation.** The criterion's written control (the same pattern over the LOCAL `.moai/config/sections/workflow.yaml`) cannot fire: this tree's local config carries no `jev:` block (measured, `grep -c "jev:"` → 0), and adding one to make a control fire would edit operator-owned state to suit a test. The control is adapted to fixture configs: the scanner is shown resolving a real `enabled: false` block AND reading an `enabled: true` value, so a template violation would be detected rather than silently unread. The adaptation is recorded here per the delegation instruction.

**Consumer C rows (re-measured in this run, not carried over):**

| AC | Status | Verification command (test) | Actual output |
|---|---|---|---|
| AC-JEVN-001 | PASS | `go test ./internal/kanban/ -run TestHasAgentFindingForPair_FalseOnJevOnlyPair` | suite `ok` (`internal/kanban` 168.148s, 0 FAIL) |
| AC-JEVN-002 | PASS | `TestBacklogSourceJev_IsAThirdConstant` | suite `ok` |
| AC-JEVN-003 | PASS | `TestHasFindingForPairAnySource` + `TestJevPrecedence_HalfB_LaterFindingsLandAlongside` + `TestJevFinding_PrecedenceHalfA_Suppressed` | suite `ok` (both packages) |
| AC-JEVN-004 | PASS | `TestJevFindingLine_DistinctFromMechanicalScore` | suite `ok` (`internal/cli` 983.528s, 0 FAIL) |
| AC-JEVN-005 | PASS | `TestJevFinding_WritesNoCardField` | suite `ok` |
| AC-JEVN-012 | PASS | `TestJevProducers_NeverWriteSourceAgent` | suite `ok` |
| AC-JEVN-013 | PASS | `TestJevFinding_AdmissionOnly_ReSweepNeverCallsIt` | suite `ok` |

**Consumer A rows: NOT RUN.** AC-JEVN-006/007/008 belong to M5, which remains blocked (forbidden this session — `§F` Milestone unblock provenance). AC-JEVN-015 is **not applicable this run**: no consumer is in the un-runnable state, and the M6 record above maintains the distinction that criterion polices (an unrun gate is owed and names what is missing; it does not claim a withholding mechanism).

**RED evidence (captured before GREEN, this run).** RED-A — the verification-surface tests against the real tree before the prose block existed (`go test ./internal/cli/ -run TestSuggestionProtocol`, exit 1): `no "### Skill Suggestion (gated — default off)" section found — the suggestion-protocol block is absent` under both `--- FAIL: TestSuggestionProtocol_CopiesAgreeAndCarryContract` and `--- FAIL: TestSuggestionProtocol_FalsificationControl`. RED-B — the anchor tests before the anchor existed (`go test ./internal/cli/ -run 'TestJevSkillSuggest|TestJevGateUnrun|TestJevSuggestCmd'`, exit 1, build failed): `undefined: jevSkillCandidate … undefined: jevSkillSuggestionFlow … undefined: presentJevSkillRankSignal … too many errors` — the compile failure is the Go TDD red for a not-yet-existing subject; no test-after substitution occurred. The in-suite falsification control (`TestSuggestionProtocol_FalsificationControl`) re-observes the mutant red on every run.

**Mid-run defect, fixed.** The full `internal/cli` suite run (first pass, 1012.461s) failed one test: `TestJevCallPath_HasExactlyTheDeclaredConsumers` — the guard's declared-consumer set did not know the new file. The guard's own doc comment names extension as "the declared way a consumer arrives", so `jev_skill_suggest.go` was added to the allowed map citing `SPEC-JEV-CONSUMERS-001 M6 — Consumer B (gate-unrun)`; the full suite was then re-run green (983.528s). Three transient lint findings (ineffassign, staticcheck QF1001, unused helper) were introduced and fixed during the same run; net new lint warnings: 0.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-22
run_commit_sha: "34fc5ac63"   # the M6 feat commit; this evidence-record commit follows it
run_status: "gate-unrun (M6): Consumer B present, default-off, unreachable at the shipped default; measurement gate not yet run — REQ-JEVN-016"
ac_pass_count: 11
ac_fail_count: 0
ac_blocked_count: 1          # AC-JEVN-010 blocked-on-threshold (no fitted threshold; Q3, SPEC-JEV-OPTIN-MEASURE-001)
ac_not_run_count: 3          # AC-JEVN-006/007/008 — M5 blocked, forbidden this session
ac_not_applicable_count: 1   # AC-JEVN-015 — no consumer in the un-runnable state this run
preserve_list_post_run_count: 4   # todo_jev_finding.go, todo_analysis.go, backlog_store.go, scripts/jev/** — none in the M6 diff
l44_pre_commit_fetch: "absorbed pre-dispatch by the lead (origin/develop merged at 12ecb2d4e, 34 2 divergence)"
l44_post_push_fetch: "n/a — lane does not push; the lead batch-pushes develop"
new_warnings_or_lints_introduced: 0   # net; 3 transient findings introduced and fixed in-run (see §E.2)
cross_platform_build.darwin: pass     # go build ./... exit 0; make build exit 0
cross_platform_build.windows: pass    # GOOS=windows GOARCH=amd64 go build ./... exit 0
total_run_phase_files: 9              # M6: 7 code/test/template files + catalog.yaml + spec.md; +1 progress.md in the evidence commit
m1_to_mN_commit_strategy: "per-milestone feature commits; M6 = 1 feat commit (implementation + skill prose + spec frontmatter transition) + 1 evidence-record commit"
```


## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-22
sync_commit_sha: "0e2a61c81"   # backfilled after the sync commit landed (schema doctrine D3); placeholder in the sync commit itself
sync_status: "closed per the SPEC's own Definition of Done — no consumer ships; M4/M6 recorded gate-unrun (REQ-JEVN-016), M5 block recorded"
b12_self_test_a: "PASS — grep -c 'SPEC-JEV-CONSUMERS-001' CHANGELOG.md → 0 (no duplicate-emission risk)"
b12_self_test_b: "PASS-WITH-NOTE — grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l → 20 = AC-JEVN-001..016 (16 live criteria of this SPEC) + 4 cross-referenced sibling tokens (AC-JEVC-003, AC-JEVO-010/013/014); no emission was made against this count"
b12_self_test_c: "N/A — no CHANGELOG entry emitted, so no file path was claimed in one"
changelog_entry_position: "none — withheld deliberately; see below"
frontmatter_status_transitions:
  in_progress_to_implemented: "merged into the sync commit (3-phase close — no separate implemented chore)"
  implemented_to_completed: "same single sync commit"
  updated_field: "2026-09-22 (already read the sync date; no edit needed)"
canary_compliance_check:
  ac_jevn_016_4_absence: "PASS — grep -c 'jev-suggest' CHANGELOG.md README.md README.ko.md → 0 0 0; positive control grep -c 'doctor' CHANGELOG.md → ≥1 (greps fire)"
  consumer_lines_added: 0
  readme_docs_site_touched: false
```

**Sync-phase records (three decisions, each with its reason):**

1. **CHANGELOG conflict recorded — no entry emitted.** The repo's sync template emits one `[Unreleased]` entry per sync close, but `REQ-JEVN-016` condition (iv) and `AC-JEVN-016.4` forbid presenting `jev-suggest` or any consumer as available, and the tree's own tests (`TestJevGateUnrun_NotPresentedAsAvailable`) assert the absence with a `moai doctor` positive control. Per the delegation instruction the conflict is recorded here and **no entry was added**: the expected CHANGELOG diff for this sync is zero consumer lines. An entry describing the close without naming `jev-suggest` was also declined — the delegation names a zero-line diff as the expected state, and a close entry whose subject cannot be named would advertise a state the SPEC explicitly refuses to advertise. The close is instead recorded here, in the frontmatter transition, and in the sync commit message.
2. **MX tags added (sync sub-step, per the 3-phase close).** Two `@MX:NOTE` tags were added over the run-modified Go files: `internal/cli/todo_jev_finding.go` (`appendJevNearDuplicateFinding` — Consumer C admission hook) and `internal/cli/jev_skill_suggest.go` (`jevSkillSuggestionFlow` — the Consumer B mechanical anchor), each naming the `REQ-JEVN-016` gate-unrun state and the owed measurement gate (Q3, `SPEC-JEV-OPTIN-MEASURE-001`). No `@MX:ANCHOR` was owed: both functions have exactly 1 non-test caller (measured, `grep -rn` over `internal/` excluding tests). Neither file exports any identifier, so no exported-function rule fired. The Hidden CLI (`newJevSuggestCmd`) received no user-facing docs by design — code annotation only.
3. **Codemaps partial refresh (minimal, moved rows only).** `.moai/project/codemaps/modules.md` covers `internal/cli`; the run phase added 2 root files there, so the moved rows were re-measured with the table's own commands and updated (`internal/cli` total 318 → 325, root 244 → 249, plus the `doctor*`/`todo*`/`init*`/`mcp*`/`integration*` rows and a new `jev*` row). Of the +5 root files, 2 are this card's (`todo_jev_finding.go`, `jev_skill_suggest.go`); the other 3 (`doctor_jev.go`, `init_jev_wizard.go` — t1020's CORE-001 sync; `integration_codemaps_card.go` — t1018) landed via absorbed develop commits whose own syncs did not refresh the codemap, and are reconciled here as measured drift. `provenance.json` is intentionally untouched: it stamps a `codemaps-gen` regeneration, and this is a hand-applied row refresh — the next full regeneration re-stamps it.

**Provenance note (sync-audit F1, SHOULD-FIX).** Commit `fce262818` — the SPEC-body reconciliation across `SPEC-JEV-CORE-001` + this SPEC — landed without an `Authored-By-Agent:` trailer, so the actor cannot be established from git alone. Per this card's own run records the reconciliation was the t1066 card's work (session of 2026-09-21, before trailer discipline was applied on this branch); the content itself was audited correct (`.moai/reports/t1066/sync-audit-SPEC-JEV-CONSUMERS-001.md`, PASS 92/100, 0 BLOCKING). Trailer discipline holds on every actor commit from `34fc5ac63` onward.

## §F Phase 4 Mode Selection

Logged 2026-09-22 (card t1066, session `2c27ccc1`) before this session's first run-phase `Agent()` spawn. Run phase re-entered mid-flight: M4 landed at `e3efef5c3`, SPEC reconciliation at `fce262818`, `origin/develop` absorbed at `12ecb2d4e` (`34 2` divergence, re-measure owed on the merged tree).

**Input parameters.** Tier L. M6 scope: one prose skill surface (+ its template mirror), one Go verification-surface test following the foreman precedent, the gate-unrun record — under 10 files, effectively one domain (Go test + skill markdown), coding-heavy, concurrency benefit LOW (one milestone, one implementer).

**Mode evaluation.**

| Mode | Selected | Rationale |
|---|---|---|
| direct | not selected | implementation work, not orchestrator-trivial |
| serial | **selected** | coding-heavy single-milestone delegation; Anthropic coding-task parallelism caveat |
| fanout | not selected | under 3 domains, no independent research lens to parallelize |
| sweep | not selected | not ≥ ~30 files, not a mechanical-uniform transform |

**Decision:** `serial` — one `manager-develop` spawn for M6, sequential.

**Justification.** M6 is one cohesive implementation: Consumer B's prose protocol, its verification-surface test, and the gate-unrun record. The verification shape is already fixed by the foreman precedent (`internal/kanban/foreman_queue_watch_test.go:73`), so fan-out has nothing independent to parallelize; coding-heavy work defaults to serial.

**Milestone unblock provenance.** M6 was declared blocked on N2 (`plan.md` §F). The Implementation Kickoff Approval answer — the operator-approved paste-ready handoff resumed this session — names the verification surface (the `extractForemanWatchScript` pattern: prose verbatim extraction + execution + falsification control) and directs M6 into the `REQ-JEVN-016` gate-unrun state. M5 and `SPEC-JEV-GOAL-DIST-001` seat (i) remain blocked and are forbidden this session.
