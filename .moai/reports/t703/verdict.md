# Card t703 — Run-phase Verdict: "invalid conversation receipt" on mid-conversation model switch (sol → luna)

- Branch: `WT-luna-turn-receipt` (worktree `.claude/worktrees/t703`, base `44e56d017`)
- Fix commit: `9bc624a39` `fix(gateway): classify observations-level receipt rejections with recovery guidance (t703)`

## Changed files (overlap check for the lead)

| Path | Status | t707 overlap |
|---|---|---|
| `internal/gateway/translate/receipt_history.go` | modified (+35/−6) | **HIGH — this file is t707's home.** My diff touches: new `classifiedError` type + `classify` helper (after the `historyReplayGuidance` const block), 4 wrap sites inside `observations()`, and the `Check` entry error mapping. It does NOT touch `checkObserved`, `replayCause`, or the `ReplayCause` enum values. Merge order either way should be a trivial textual conflict, but the same file will conflict textually. |
| `internal/gateway/translate/receipt_history_classify_test.go` | new | none |
| `.moai/reports/t703/verdict.md` | new | none |

## Claim

1. The production 400 body `invalid conversation receipt` is the bare `receipt.ErrInvalid` text, which at the deployed build (identical to this tree) can only originate from the raw `ErrInvalid` returns inside `receiptHistory.observations()` — the manifest-level mismatches were already wrapped into `HistoryReplayError` guidance text by t672 (`715b6b3ea`, verified ancestor of the deployed build `9bfe424ee`). The firing site is the opaque-tool-marker-without-envelope check (`receipt_history.go`, the `strings.HasPrefix(id, opaque.ToolPrefix)` branch in `observations()`), reached because the client re-encoded the replayed history on the switch turn with the gateway-issued `redacted_thinking` envelopes stripped while the bound tool markers survived.
2. The card's hypothesis (per-model-bound receipt state rejecting the first luna turn) is FALSE at this tree: sol and luna map to the same receipt domain (`historyFamily` → `"gpt-5.6"`), and an unmodified luna replay of sol-published history passes Check (new guard test, passes before and after the fix).
3. The gateway cannot make the stripped turn succeed: without the envelope the request is untranslatable (the forward path requires the envelope to restore tool ids and re-inject upstream reasoning items), and accepting a stripped replay would bypass the t672 tamper invariant. The in-repo fix is therefore diagnosability + classification, not acceptance.

## Evidence

### Baseline attribution (all measured in this run, this tree)

- Deployed gateway binary at incident time: `v3.2.0-rc.10`, build commit `9bfe424ee`, built `2026-09-13T11:28:34Z` (incident 12:03:36Z). `~/go/bin/moai version` output recorded in session; ancestry: `git merge-base --is-ancestor 715b6b3ea 9bfe424ee` → yes (t672 classification included); `ce79ef7ca` (t649, luna in `historyFamily`) → yes.
- Deployed gateway code == this tree: `git diff 9bfe424ee 44e56d017 --stat -- internal/gateway/` → empty.
- Evidence log: `~/.moai/state/gateway-conversations/families/da7062b1-7b2b-4f3c-8f08-049654a8e378/native/debug/da7062b1-7b2b-4f3c-8f08-049654a8e378.txt` — :6854 `set_model model=gpt-5.6-luna` (12:03:22Z), :7348 `[claude-code:unrecognized_model]` (12:03:36.033Z), :7355 dispatch, :7357 `400 {"error":{"message":"invalid conversation receipt",...}}` (12:03:36.205Z, ~120 ms), :7375 `set_model model=gpt-5.6-sol` (12:04:07Z) after which requests succeed.
- The `unrecognized_model` warning also fired for **sol** at :405 (11:26:56Z) and 30+ sol turns succeeded — the warning line itself is recurring noise, not the trigger; the trigger is the turn whose model differs from the model that produced the replayed turns.

### RED (reproduction, exact command + output)

Command (in this worktree, at base `44e56d017` before the fix commit):

```
go test ./internal/gateway/translate/ -run 'TestModelSwitch' -count=1 -v
```

Output:

```
=== RUN   TestModelSwitchWithinFamilyAcceptsUnmodifiedHistory
--- PASS: TestModelSwitchWithinFamilyAcceptsUnmodifiedHistory (0.05s)
=== RUN   TestModelSwitchStrippedEnvelopeRejectionIdentifiesReason
    receipt_history_classify_test.go:121: gpt-5.6-luna stripped-envelope rejection must carry recovery guidance, got invalid conversation receipt
--- FAIL: TestModelSwitchStrippedEnvelopeRejectionIdentifiesReason (0.04s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway/translate	0.519s
FAIL
```

The fixture publishes 3 sol turns (each assistant message: `redacted_thinking` envelope + `toolu_moai_v1_*` bound tool marker + text), then Checks a luna request (a) unmodified — passes, and (b) with the envelopes stripped and markers intact — reproduces the exact production body `invalid conversation receipt` at the marker-without-envelope site. The unmodified-pass result simultaneously disproves hypothesis (2) above.

### GREEN (fix + verification, exact command + output)

Fix: the four raw `ErrInvalid`-class returns in `observations()` are wrapped with a `classifiedError` carrying their `ReplayCause` (`CauseReasoning` for the dropped-envelope marker site, `CauseChain` for shape/restore-conflict sites), and `Check` maps classified errors to `HistoryReplayError` — the same client-visible guidance the manifest-level classes already use. Validation is unchanged; nothing that was rejected is now accepted.

Command:

```
go test ./internal/gateway/translate/ -run 'TestModelSwitch' -count=1 -v
go test ./internal/gateway/translate/ ./internal/gateway/receipt/ -count=1
go vet ./internal/gateway/... ./internal/cli/
go build ./internal/cli/
golangci-lint run internal/gateway/translate/...
```

Output:

```
--- PASS: TestModelSwitchWithinFamilyAcceptsUnmodifiedHistory (0.08s)
--- PASS: TestModelSwitchStrippedEnvelopeRejectionIdentifiesReason (0.04s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	0.971s
ok  	github.com/modu-ai/moai-adk/internal/gateway/receipt	1.109s
```

`go vet` exit 0; `go build ./internal/cli/` exit 0; golangci-lint: 15 pre-existing errcheck findings, all in test files not touched by this card (`receipt_history_test.go`, `tool_reference_test.go`, …); zero findings in the two files this card changed.

### Invariants

- Preserved: every modified, stripped, foreign-session, foreign-owner, or tampered history is still rejected (full existing translate + receipt suites pass, including t672's "stripped cipher accepted"-must-fail assertions). No authorization or validation is bypassed; the change is error classification only.
- Relaxed: none in acceptance. The only behavioral change is the 400 body for the affected rejection class: previously the bare `invalid conversation receipt`, now the t672 guidance (`"... (reason: replayed history omits gateway-issued reasoning recorded at this position; replay the history unmodified or start a new conversation)"`).

### Task-4 adjudications

- (a) Reverse switch luna→sol: same mechanism, symmetric. `historyFamily` maps both to `"gpt-5.6"` (identical domain), and the client-side re-encoding that strips envelopes is a function of the current-turn model differing from the producing model, not of direction. Covered by the same test (both directions assert `CauseReasoning`).
- (b) `[claude-code:unrecognized_model]` at :7348: the warning line is **incidental** — it also fired for sol at :405 (11:26:56Z) before 30+ successful turns. The underlying *condition* (model not in the client's catalog) is plausibly the enabler of the switch-turn history re-encoding that orphans the envelopes, but the warning itself is neither necessary (sol showed it while succeeding) nor sufficient.

### c6dd8d8d adjudication (separate defect — not folded in)

`families/d12913aa-…/native/debug/c6dd8d8d-963a-43d8-b6e3-78f8401a5d33.txt` :235 shows `400 {"error":{"message":"Bad Request",...}}` at 11:26:33Z — a generic body, not `invalid conversation receipt`, in a different conversation family, with zero luna strings (grep verified). The family's own transcript (`…d12913aa-f044-….txt` :2339) separately shows a t672-classified rejection (`conversation history changed, lacks reasoning, …`) at 08:13:17Z. Both are distinct failure classes from t703; no fix attempted here.

## Gaps

- The client-side payload of the failing luna request was not captured (the gateway debug log does not record request bodies), so "the client stripped the envelopes" is a reconstruction from: bare-ErrInvalid message class (narrows to the 4 `observations()` sites, none of which read the model string ⇒ the payload must have changed on the switch turn), the sol-revert success (stripping is outgoing-request-only, not transcript mutation), and the fixture-level reproduction of that shape. A request-body capture on the next occurrence will confirm which of the 4 sites fires — the classification fix now makes that visible in the 400 body itself, which is the diagnosability closure for exactly this gap.
- Not run: full local suite (machine-load rule; CI owns the full verdict). Affected packages `./internal/gateway/translate/` and `./internal/gateway/receipt/` plus `go vet ./internal/gateway/... ./internal/cli/` and `go build ./internal/cli/` were run.

## Residual-risk

- If the next occurrence shows a different classified cause (e.g. `CauseChain` on the canonical-prefix site), the client transformation differs from the modeled stripping (e.g. extra message fields instead of envelope removal); the classification will say so directly.
- The genuine client-side continuation fix (making Claude Code not strip gateway-issued reasoning on a model switch, or presenting the family models to the client as one recognizable model) is outside this repository; until it lands, a switch turn with tool-bearing history will still fail on the first request of the new model — now with a self-identifying, actionable 400 body.
