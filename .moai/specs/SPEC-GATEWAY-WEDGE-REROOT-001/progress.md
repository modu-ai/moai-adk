# Progress — SPEC-GATEWAY-WEDGE-REROOT-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-pending
plan_complete_at: 2026-09-13
note: Plan artifacts authored (spec.md + plan.md + acceptance.md, Tier M). Flips to `audit-ready` on plan-auditor PASS (card t700 lead flow, task after SPEC authoring).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

---

## Plan-phase Research Record (card t700)

Baseline-attribution: worktree `.claude/worktrees/t700`, branch `WT-wedge-reroot-policy`, base local develop `44e56d017` (this run, this tree). All reads inside the worktree; no code, test, or template files modified (plan-phase scope boundary).

### Evidence base (read in full)

- `.moai/reports/t672/verdict.md` — wedge cause classification shipped; recovery named as this card; gaps 3 ("웨지의 자동 복구(안전 재뿌리 내기)는 구현하지 않았다") and residual-risk 1 name this card.
- `.moai/reports/t672/investigation.md` — mechanism proof (one receipt root per gateway child; metadata-session gate precedes history; 4-way Check binding); design decision record (candidates A and B REJECTED with reasons); isolation discipline (§0) inherited as plan.md §D.
- `.moai/reports/t672/matrix.md` — 9-cell real-request matrix; C4b (tail-truncated fork replay → 200 before and after) and C5 (wedge replay → 400, chain-classified after) are the two load-bearing cells for the security determination.

### Key code reads (symbols pinned, not line numbers)

- `internal/gateway/translate/receipt_history.go` — observed: `ReplayCause` (`CauseChain`/`CauseLineage`/`CauseReasoning`), `HistoryReplayError.Error()` with `historyReplayGuidance` as byte-identical prefix + three fixed reason clauses, `NewReceiptHistory` / `NewGPTSubscriptionReceiptHistory` ("Request metadata never selects a receipt root or authorizes a conversation"), `Check` → `observations` → `checkObserved` → `Manifest.Check`, `replayCause` classification, `Publish` (validates all but the last observation, publishes the last candidate).
- `internal/gateway/receipt/core.go` — observed: `Manifest.Check` per-boundary four-way match (Prefix, Previous, Opaque, Items) with anti-downgrade for empty-envelope boundaries (`found = empty && !required`); `Manifest.Fork` copies the parent candidate set verbatim into a child manifest keyed to a launcher-issued UUID.
- `internal/gateway/conversation/family.go` — observed: `Manager.Fork` opens the parent store, snapshots candidates, creates the child root under `families/<family>/forks/<id>/receipt`, publishes every parent candidate into it, registers the child record, returns a descriptor with `--resume --fork-session --session-id <id>` — the sanctioned launcher-driven path-(b) instance.
- `internal/cli/gateway_factory.go` — observed: `newGatewayHandlerFactory` opens exactly one `receipt.Store` from the private payload's `conversation.receipt_dir` under the launcher-authorized `conversation.session_id`, and wires that single store into both `limits.History` (`NewGPTSubscriptionReceiptHistory`) and `limits.NativeReceiptAuthorize` (`authorizeGatewayNativeReceipt`, which enforces `policy.UserID` session equality against the root before history validation).
- `internal/gateway/translate/request.go` — observed: the native-receipt authorization gate (`NativeReceiptAuthorize`) executes before `limits.History.Check`; history replay validation entry passes `c.messages` (the replayed history) to `Check`.

### Research commands and observed outputs

1. Characterization suite inventory:
   `grep -n "func Test" internal/gateway/translate/receipt_history_cause_test.go internal/gateway/translate/receipt_history_test.go`
   → observed 11 tests, including `TestReceiptHistoryAcceptsTruncatedForkReplay` (the C4b lock), `TestReceiptHistoryClassifiesDesyncWedgeAfterUnpublishedTurn` (the C5 lock), `TestReceiptHistoryRejectsForeignItemReplay`, `TestReceiptHistoryKeepsStrippedReasoningRejection`, `TestReceiptHistorySpawnShapedFreshHistoryNeedsNoLineageSeeding`, plus 4 older rejects/tests in `receipt_history_test.go`.
2. Publish call sites (where an unpublished turn can originate):
   `grep -rn "\.Publish(" internal/gateway/ --include="*.go" | grep -v "_test.go"`
   → observed: `receipt_history.go` (candidate publish), `translate/response.go` (History.Publish on a completed turn), `receipt/store.go`, `conversation/family.go` (fork seeding). Confirms: a failed/aborted turn publishes nothing — the wedge's origin.
3. Launcher-driven fork caller (path-(b) instance):
   `grep -rn "\.Fork(" internal/cli/ internal/gateway/conversation/ --include="*.go" | grep -v "_test"`
   → observed: `internal/cli/gateway_session.go:204` (`families.Fork`) — the launcher conversation flow is the existing client-side surface this card's recovery path extends.
4. SPEC ID pre-write self-check (HARD protocol):
   `ID="SPEC-GATEWAY-WEDGE-REROOT-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL`
   → observed: `PASS` (verbatim).
5. SPEC ID uniqueness:
   `ls -d .moai/specs/SPEC-GATEWAY-WEDGE-REROOT-001` → no such directory; `grep -rl "SPEC-GATEWAY-WEDGE-REROOT" .moai/specs/` → no references. Unique.
6. Related-SPEC scan:
   `ls .moai/specs/ | grep -i -E "RECEIPT|WEDGE|REPLAY|GATEWAY"` → `SPEC-MOAI-GATEWAY-001`, `SPEC-MOAI-GATEWAY-PICKER-001`, `SPEC-MOAI-GATEWAY-TEAMMATE-001`; `grep -l "ReplayCause\|receipt chain" .moai/specs/*/spec.md` → no matches (t672 shipped as a card with reports, no SPEC — this SPEC carries the policy lineage instead).
7. Base verification:
   `git log --oneline -3 develop` → `44e56d017 Merge branch 'WT-console-tab-names' into develop (t677)` … — matches the dispatch's declared base.

### Gaps (not observed at plan phase)

- t697's actual diff was not read (worktree-session isolation; reading a sibling card's tree is outside this card's scope). The absorb requirement (plan.md §A/§C, M1) is the mitigation: every symbol pin is re-verified against the absorbed tree before the first code commit.
- No live gateway probe was run at plan phase (none needed — t672's matrix supplies the measured cells).
- The client transcript's exact file format at the recovery surface is a run-phase M3 detail (the conversation record's transcript pointer is the launcher-level anchor observed at `internal/cli/gateway_session.go`).
