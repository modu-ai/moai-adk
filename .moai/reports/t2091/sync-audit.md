auditor-model: claude-opus-5-5

# Sync audit — card t2091 (follow-ups to the t2071 audit)

verdict: PASS
audited_sha: 0ad78fd95ec7421e58d3c9bd5d24136ab823e339

Overall: PASS, score 0.92 (Functionality 94 / Security 96 / Craft 88 / Consistency 92).
Range audited: `a63295d159..0ad78fd95e`. The base is the merge that brought in t2071's tip `00e8737b32`, which was already audited. HEAD is the t2091 code commit.
Iteration: 1. This is the first audit of this card, so no finding can be repair-born.
Audit mode: Claude-only and in-session, as the brief required. I made no codex, glm, audit_multi or claude_audit call and ran no paid or cross-model tool.

Must-pass firewall (Functionality, Security): both dimensions pass. There are no blocking findings. All findings below are optional.

## Dimension scores

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 94 | PASS | All 9 focused tests and 7 subtests pass (verbatim output below). F2, F3 (mutants a and b), F4, F7 and the earlier-issuance residual risk are each addressed by the diff. The one production change only alters the refusal suffix text. |
| Security (25%) | 96 | PASS | No gate, receipt or stored-data logic changed. The new note interpolates `card.CardID`, which comes from a row loaded after `homestate.ValidCardID(cardID)` (factory_card_rebind.go:165). There is no new input surface and no write path. |
| Craft (20%) | 88 | PASS | `go vet` is clean and `gofmt -l` is empty. The new guards kill the mutants the card named. Small survivors remain (N2, N4), and the new note's behaviour on the verdict path is untested (N1). |
| Consistency (15%) | 92 | PASS | The change follows the file's `note`-suffix idiom and its card-tag comments, and the header comment was updated. One doc bullet is narrower than the code (N3). |

## Findings

- N1 [low] [optional] [in-diff] `internal/cli/factory_card_rebind.go:124-128`, applied at line 236: the new sha-less note is returned by `resolveRecordedLanding` whatever the caller's path, so it also reaches the verdict-path refusal (`card %s: no %s verdict to re-bind onto: %w%s`). Plain `moai todo landed <card>` without `--sha` is a normal record (todo_landed.go:189-202 sets `SHA` only when one is supplied). So a staged card whose verdict is missing or unreadable, and which carries such a record, now gets the suffix "the card's landing record names no delivering commit — add it with `moai todo landed …`". The suffix is accurate, and following it does enable the designed squash-landed fallback (line 238). It is consistent with the other notes that already ride the verdict path (earlier issuance, commit refused). But no test pins this combination, and the name `TestAFFlowRebindHintIsNotOnTheVerdictPath` can be misread as "no landing suggestion ever rides the verdict path". Suggested fix: add a subtest with a staged card, a missing verdict and a sha-less record that pins the intended suffix, or narrow the test name or comment to "the absent-record hint". Confidence: high.
- N2 [info] [optional] [in-diff, tests] `TestAFFlowRebindHintWhenTheLandingRecordNamesNoCommit` asserts "names no delivering commit" and the absence of "no landing is recorded". It does not assert that the note names the remedy (`moai todo landed t1 --sha`). A mutant that drops the remedy command from the new note passes every new test. Confidence: high.
- N3 [info] [optional] [in-diff, doc] `.moai/docs/af-flow-landed-card-close.md:45-46` attributes `recorded landing commit refused: ...` only to unreachability. `homestate.ResolveLandedEvidence` (card_metadata.go:383-395) also returns errors behind that prefix for an unusable ref (`landed ref %q is not usable`) and for a commit that does not resolve (`resolveCommit`). The remedy the doc gives is right for the af-flow case it describes. Confidence: high.
- N4 [info] [optional] [in-diff, tests] Two mutants still pass the landed-hint guard: (i) the hint emitted on the `unrunnable` branch, which was mutant (c) of the t2071 audit and is outside this card's named scope; (ii) a hint hard-coded to the literal ref `origin/main`, which would change behaviour only for `origin/<other branch>`. Neither is reachable through the tests' git fixtures without extra setup. Confidence: medium.

No defect was found in the earlier-issuance test, the negative control or the dead-import removal.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim 1 — the delta is the four named files, and the only production change is the note in `resolveRecordedLanding` plus a comment

Evidence:

```
$ git -C <wt> diff a63295d159..HEAD --stat
 .moai/docs/af-flow-landed-card-close.md          |  13 +++
 internal/cli/factory_af_flow_close_test.go       |   9 +-
 internal/cli/factory_af_flow_hints_t2091_test.go | 120 +++++++++++++++++++++++
 internal/cli/factory_card_rebind.go              |   9 +-
 4 files changed, 147 insertions(+), 4 deletions(-)
```

The production hunk, as read in the full diff:

```go
	var sha, ref, observed string
+	hasRecord := false
	for i := range rec.Items {
		if rec.Items[i].ID == card.CardID && rec.Items[i].Landing != nil {
+			hasRecord = true
...
	if sha == "" {
+		if hasRecord {
+			// card t2091: a record without a delivering commit is not an absent record.
+			return "", fmt.Sprintf("; the card's landing record names no delivering commit — add it with `moai todo landed %s --sha <delivering commit> [--ref <branch>]`", card.CardID)
+		}
		return "", ""
	}
```

The comment change adds one sentence to the header: "A record that names no delivering commit (card t2091) is not absent: its note says so." No gate, receipt, store or migration code is in the diff.

Callers, from `grep -rn "resolveRecordedLanding" internal`:

```
internal/cli/factory_card_rebind.go:95:// resolveRecordedLanding reads ...
internal/cli/factory_card_rebind.go:109:func resolveRecordedLanding(...)
internal/cli/factory_card_rebind.go:226:		landed, note := resolveRecordedLanding(ctx, root, db, runID, card)
```

There is exactly one caller (line 226). It uses `note` only as a suffix, on two paths:
- Stage-empty path (`phaseErr != nil`, lines 228-234): a non-empty note suppresses the t2071 "no landing is recorded" hint (`if note == ""`). This is the intended F2 behaviour: a sha-less record now yields the accurate note instead of the "absent" hint. The returned error still wraps `phaseErr` with `%w`, so `errors.Is` callers are unaffected.
- Verdict path (`phaseErr == nil`, line 236): the note is appended to the verdict refusal. This is the behaviour change described in N1. It is accurate, and no caller interprets the note as a control signal.

`landed == ""` is still the only control signal. The sha-less branch returns `""` exactly as before, so the success and refusal branching is unchanged. Only the error text differs.

Baseline-attribution: tree at `0ad78fd95ec7421e58d3c9bd5d24136ab823e339` (`git rev-parse HEAD` observed this value in this audit). Diff base `a63295d159`.

### Claim 2 — the focused tests pass

Evidence. I ran this command, wrote its output to a scratch file, and filtered the result lines:

```
$ go test ./internal/cli -run 'TestAFFlow|TestFactoryRebindEvidenceStageEmptyRow|TestCompletionChainAcrossRunRows' -count=1 -v   (exit=0)
--- PASS: TestAFFlowLocalMainCardClosesThroughFormalVerbs (3.03s)
--- PASS: TestAFFlowRefusalsNameTheLandingStep (4.63s)
    --- PASS: TestAFFlowRefusalsNameTheLandingStep/rebind_without_a_landing_record_names_todo_landed_with_ref (1.88s)
    --- PASS: TestAFFlowRefusalsNameTheLandingStep/landed_on_the_default_ref_names_the_local_branch_form (2.75s)
--- PASS: TestAFFlowRebindHintWhenTheLandingRecordNamesNoCommit (1.80s)
--- PASS: TestAFFlowRebindHintIsNotOnTheVerdictPath (1.58s)
--- PASS: TestAFFlowEarlierIssuanceRecordIsRefusedAndRecordingAgainHeals (2.80s)
--- PASS: TestAFFlowLandedHintOnlyForAnUnreachableCommitOnAnOriginRef (3.06s)
    --- PASS: TestAFFlowLandedHintOnlyForAnUnreachableCommitOnAnOriginRef/a_local_ref_gets_no_suggestion (0.36s)
    --- PASS: TestAFFlowLandedHintOnlyForAnUnreachableCommitOnAnOriginRef/a_missing_commit_gets_no_suggestion (0.34s)
--- PASS: TestFactoryRebindEvidenceStageEmptyRow (4.76s)
    --- PASS: TestFactoryRebindEvidenceStageEmptyRow/rebinds_from_the_recorded_landing_without_inferring_a_phase (1.61s)
    --- PASS: TestFactoryRebindEvidenceStageEmptyRow/still_refuses_without_a_landing_record (1.49s)
    --- PASS: TestFactoryRebindEvidenceStageEmptyRow/still_refuses_a_recorded_landing_that_is_not_on_the_landed_ref (1.66s)
--- PASS: TestCompletionChainAcrossRunRows (4.27s)
    --- PASS: TestCompletionChainAcrossRunRows/rebind_and_approve_on_the_current_run_row_satisfy_the_done_gate (1.49s)
    --- PASS: TestCompletionChainAcrossRunRows/a_receipt_for_an_older_run_is_still_refused_and_approve_says_so (1.42s)
    --- PASS: TestCompletionChainAcrossRunRows/an_unlanded_stage_empty_card_still_cannot_reach_the_gate (1.36s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	27.532s
```

```
$ go vet ./internal/cli        -> exit=0, no output
$ gofmt -l <the three changed .go files>   -> no output, exit=0
```

Baseline-attribution: `go test` compiled the binary from this worktree at HEAD `0ad78fd95e`, which was clean at session start. No installed binary was involved, so §2.2 tool provenance does not apply.

### Claim 3 — mutant analysis of the lane's guard tests (from the test text; nothing executed with mutants)

- `TestAFFlowRebindHintWhenTheLandingRecordNamesNoCommit` writes a `Landing` with no `SHA` on a stage-empty row. It asserts the audit-phase refusal, the substring "names no delivering commit", and the absence of "no landing is recorded". It kills the pre-card behaviour (`("", "")` for a sha-less record, which the lane reported as the observed RED) and any wording that still calls the record absent. It does not kill a mutant that drops the remedy command from the note (N2).
- `TestAFFlowRebindHintIsNotOnTheVerdictPath` deletes `sync-audit-iter1.md` on a staged card that has no landing record. It asserts "verdict to re-bind onto" and bans "no landing is recorded" and "moai todo landed". It kills mutant (a), the t2071 hint moved out of `if phaseErr != nil`. It also kills dropping the `hasRecord` check, because the new note would then appear with no record and carry "moai todo landed". It does not cover the verdict path combined with a sha-less record (N1).
- `TestAFFlowLandedHintOnlyForAnUnreachableCommitOnAnOriginRef`:
  - Subtest (a), with an orphan commit and `--ref main`, kills mutant (b): a hard-coded `--ref main` hint, or `TrimPrefix` without the `remote` check, would emit "--ref main" for ref `main`.
  - Subtest (b), with a non-existent sha on the default ref, kills a hint added to the existence branch.
  - Survivors: N4.
- `TestAFFlowEarlierIssuanceRecordIsRefusedAndRecordingAgainHeals` kills two mutants: removal of the engagement-start comparison (the first assertion fails), and a `todo landed` that keeps the old `ObservedAt` (the second rebind fails). Code checked: todo_landed.go:131 `rec.Items[i].Landing = record` and :192 `ObservedAt: time.Now().UTC().Format(time.RFC3339)`.

Gap: the lane's two temporary-mutant runs (progress.md) were not re-executed, because the brief forbids editing tracked files. The judgments above come from reading the test text.

### Claim 4 — the e2e negative control (`todo done` before `approve`) proves gate engagement

Evidence (code read). On plain `todo done t1` (todo.go:1387-1445), the only refusals before archive are these: no item (the item exists), `--expect` (not passed), `--require-landed` (not passed), gate open errors, and `gate.verifyForClose`.

`verifyForClose` (leader_approval_gate.go:91-103) returns nil when `!linked`. In that case `todo done` would succeed and the test's `err == nil` check would fail it, so the control does distinguish "gate silently treats the row as unlinked".

For a linked row, `ApprovalGate.Verify` (homestate/leader_approval.go:500-504) checks the receipt lookup first:

```go
a, err := findLeaderApprovalPreferringRun(ctx, g.tx, cardUUID, card.RunID)
if errors.Is(err, ErrApprovalMissing) {
    return fmt.Errorf("%w: card %s needs a leader approval receipt to complete", ErrApprovalMissing, card.CardID)
}
```

Before `approve` no receipt exists, so the refusal is the missing-receipt sentinel, and no card-state check precedes it. A refused `todo done` writes nothing: Mutate discards the write and the gate's transaction is rolled back. The later approve-then-done success therefore runs on an unchanged state, and the test passing confirms it. The assertion matches the substring "receipt" rather than using `errors.Is(err, homestate.ErrApprovalMissing)`. That is looser than it could be, but it is sufficient because the only reachable refusal on this path is the receipt one.

### Claim 5 — the new doc section matches the code

- "`recorded landing refused: observed <time>, before this run engaged the card (<time>) — it belongs to an earlier issuance`" matches factory_card_rebind.go:144 `"; recorded landing refused: observed %s, before this run engaged the card (%s) — it belongs to an earlier issuance"`.
- "stored per card id, not per run" matches the header comment at :104-105 and the loop keyed only on `rec.Items[i].ID == card.CardID` (:117).
- "`todo landed` replaces the record with a fresh observation time" matches todo_landed.go:131 (whole-field replacement) and :192 (`time.Now()`).
- "pinned by `TestAFFlowEarlierIssuanceRecordIsRefusedAndRecordingAgainHeals`": the test exists in the new file and passed in Claim 2.
- "`the card's landing record names no delivering commit`" matches :127.
- "`recorded landing commit refused: ...`": the prefix matches :159. The cause given is narrower than the code (N3).
- "(step 3)" refers to `moai todo landed <card> --sha <sha> --ref main`, which is step 3 in the doc's list (line 28). That is correct.

## Gaps

- I did not run `moai verify audit-plan`, and I did not call any cross-model MCP tool (audit_multi, codex, glm, claude_audit), because the operator mandated a Claude-only audit for this chain. If a receipt guard asks for a codex receipt at stop, this audit does not supply one: receipts=none.
- The lane's temporary-mutant runs were not reproduced (no tracked-file edits allowed). Claim 3 is reasoning from the test text.
- The verdict path combined with a sha-less record (N1) was not executed. The behaviour is inferred from the single caller at line 226.
- No coverage figure was measured and golangci-lint was not run. The brief limited execution to the focused test command and `go vet`. The lane reports `golangci-lint run --new-from-rev=a63295d159 ./internal/cli/...` with 0 issues; I did not re-observe that.
- The full `internal/cli` suite was not run, per the lane-local rule. Other rebind test files (`factory_card_rebind_landed_test.go`, `factory_card_t1930_test.go`, `factory_card_rebind_test.go`) were outside the filter. By grep, none of them constructs a landing record without a SHA (`squashLandedFixture` always sets `SHA` when it creates a record), so none of them can observe the new note. This is inferred, not executed.
- No tool refusals occurred. One grep failed on zsh glob expansion of `--include=*.go`, and I re-ran it without that flag.

## Residual-risk

- The sha-less note now also appears on the verdict refusal (N1). That is harmless and accurate, but unpinned, so a later edit could change it in either direction without a test noticing.
- Wording-level mutants survive (N2, N4). A later edit could drop the remedy command from the new note, or add the landed hint to the unrunnable branch, and no test would fail.
- The earlier-issuance heal relies on the re-recorded `ObservedAt` (wall clock, second precision) being no earlier than the run's engagement start. With a skewed or rewound clock, recording again would not heal it. The doc does not mention this edge.
