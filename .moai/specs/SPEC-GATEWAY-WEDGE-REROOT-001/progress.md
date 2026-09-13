# Progress — SPEC-GATEWAY-WEDGE-REROOT-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-13
note: Plan artifacts authored (spec.md + plan.md + acceptance.md, Tier M). Plan-audit iteration-2 delta PASS 1.00 (Tier M threshold 0.80; iteration 1 COND-FAIL 0.6875, D1-D7 fix pass applied). Evidence: .moai/reports/t700/plan-audit.md (iter-1), .moai/reports/t700/plan-audit-iter2.md (iter-2). Symbol pins re-verified on the absorbed tree (a86ff2e3c); t707 consistency checked — no contradiction.

## §E.2 Run-phase Evidence

### M1 pre-flight record (plan.md §C — executed at absorb HEAD `d025b463c`, 2026-09-14)

- Absorb: `git merge develop` (develop = origin/develop = `643abfb8c`) → merge HEAD `d025b463c`, conflicts 0. Carries `.moai/reports/t838/` (live-instrumentation verdict).
- Ancestry gate: `git merge-base --is-ancestor f45c2dddf HEAD` → PASS (t697 merge is ancestor).
- Exported pins: `go doc ./internal/gateway/receipt Manifest.Check` → resolves; `go doc ./internal/gateway/conversation Manager.Fork` → resolves.
- Unexported pins (declaration grep): `NewReceiptHistory` :86, `NewGPTSubscriptionReceiptHistory` :93, `Check` :197, `replayCause` :223, `Publish` :239, `checkObserved` :262 (receipt_history.go); `authorizeGatewayNativeReceipt` :59, `newGatewayHandlerFactory` :80 (gateway_factory.go); `families.Fork(` :204 (gateway_session.go) — ALL resolve.
- Characterization baseline: `go test ./internal/gateway/translate/ -run TestReceiptHistory -count=1` → `ok github.com/modu-ai/moai-adk/internal/gateway/translate 1.225s`.
- Conflict pre-scan: `grep -rn "Retired\|superseded" internal/gateway/translate/ internal/gateway/receipt/` → no conflicts.
- Scoped lint baseline: `golangci-lint run --new-from-rev=HEAD internal/gateway/... internal/cli/...` → `0 issues.`
- Kickoff authorization: operator approval relayed by lead (dispatch citing plan-audit-iter2.md PASS 1.00 @ `10f6be792`).
- AC-WRR-013 disposition: live probe SKIPPED by orchestrator decision — t672 matrix C4b already proves the acceptance shape live; the new surface (client-side transcript surgery) is unit-testable; upstream cost avoided. Gap recorded; revisit only if run-phase evidence leaves the wedge→re-root→retry path in doubt.

### M2 — validator characterization lock (commit at HEAD recorded in the M2 commit; 2026-09-14)

- Command: `go test ./internal/gateway/translate/ -run 'TestReceiptHistory|TestHistoryReplayError' -count=1` → `ok github.com/modu-ai/moai-adk/internal/gateway/translate 0.884s` — 14/14 PASS, including the three new characterization tests:
  - `TestReceiptHistoryAcceptsTailRerootedWedgeReplay` (AC-WRR-002): wedge shape (published turns + user turn + never-published assistant boundary with tool_use + dependent tool_result + plain user turn) rejected pre-recovery; tail-re-rooted remainder accepted by the unchanged check.
  - `TestReceiptHistoryRejectsMidDropAfterTailReroot` (AC-WRR-003): mid-history boundary drop stays rejected, chain-classified, after a tail re-root.
  - `TestHistoryReplayErrorGoldenStrings` (AC-WRR-007): all three `HistoryReplayError` messages byte-identical to full literals (guidance prefix + clause), independent of the production constants.
- AC-WRR-001 (wedge rejected, chain-classified): `TestReceiptHistoryClassifiesDesyncWedgeAfterUnpublishedTurn` PASS (pre-existing C5 lock, re-confirmed this run).
- AC-WRR-004/005/006: `TestReceiptHistoryRejectsForeignItemReplay` / `TestReceiptHistoryKeepsStrippedReasoningRejection` PASS unchanged (lineage assertion inside the foreign-item test's CauseLineage section).

### M3 — client-side recovery path (TDD; 2026-09-14)

- **E8 RED evidence (verbatim, pre-GREEN)** — command `go test ./internal/cli/ -run 'TestGatewayReroot' -count=1`:
  ```
  # github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
  internal/cli/gateway_reroot_test.go:88:14: undefined: rerootGatewayTranscript
  internal/cli/gateway_reroot_test.go:130:14: undefined: rerootGatewayTranscript
  internal/cli/gateway_reroot_test.go:155:16: undefined: rerootGatewayTranscript
  internal/cli/gateway_reroot_test.go:167:16: undefined: rerootGatewayTranscript
  internal/cli/gateway_reroot_test.go:177:16: undefined: rerootGatewayTranscript
  internal/cli/gateway_reroot_test.go:195:14: undefined: rerootGatewayTranscript
  internal/cli/gateway_reroot_test.go:267:14: undefined: rerootGatewayTranscript
  internal/cli/gateway_reroot_test.go:426:14: undefined: rerootGatewayTranscript
  FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
  FAIL
  ```
- **GREEN** — `go test ./internal/cli/ -run 'TestGatewayReroot' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 1.486s` (9/9 PASS: AC-WRR-009 removal semantics + boundary-only edge, AC-WRR-008 single-shot same/fresh-process + stable guidance, clean refusal edge, AC-WRR-010 structural + store-digest complement, AC-WRR-011 negative + explicit-invocation wiring + passthrough strip, fork-combination refusal, AC-WRR-016 forged tail).
- Change-scoped regression: `go test ./internal/cli/ -run 'TestGatewaySession|TestGatewayConversation|TestGatewayReroot' -count=1` → `ok ... 1.331s`, swept 17 top-level PASS (non-empty sweep); `go test ./internal/gateway/conversation/ -count=1` → `ok ... 1.859s`.
- **Surface decisions (HOW, per plan.md M3 license)**: recovery core `internal/cli/gateway_reroot.go` (`rerootGatewayTranscript`); explicit invocation = `--reroot` paired with `--resume <uuid>` in `prepareGatewayConversation` (`internal/cli/gateway_session.go` — the EXTEND target; both frozen files untouched); record access via new `Manager.TranscriptPath` in a NEW file `internal/gateway/conversation/reroot.go` (un-gated transcript pointer resolution — family.go byte-frozen); durable marker + aside = sidecars `<transcript>.reroot.json` / `<transcript>.reroot-aside.jsonl` (marker written first, exclusive-create, so the single-shot bound survives crashes and process restarts; aside-before-replace).
- **Design notes**: (1) API-error display rows (`isApiErrorMessage`) are NOT treated as the unpublished boundary — the TDD RED-GREEN loop caught the first implementation mis-taking them for the boundary (test `TestGatewayRerootRequiresExplicitInvocation` failed, implementation corrected). (2) A wedge transcript carrying a plain user turn after the unpublished boundary recovers to a shape the native completion gate (transcriptModel) refuses to resume; removal semantics per REQ-WRR-003-1 are unchanged and the aside preserves everything — the launcher-reachable wedge+retry shape (single trailing user row + terminal API-error row) resumes cleanly; the general case's sanctioned fallback is the fork path (documented in the M6 operator doc).
- Launcher-reachable wedge shape recorded: the phantom boundary carries client-side `end_turn` (stream looked complete, receipt never published) followed by the user's post-failure turn and the API-error display row — this is the only wedge shape that resumes (transcriptModel complete) and therefore the one the 400 chain rejection surfaces through.

### M4 — gateway non-invasiveness lock (AC-WRR-012 / AC-WRR-014; 2026-09-14)

- **AC-WRR-012 diff-scope assertion (verbatim)** — command `git diff --name-only 643abfb8cc536b1152efbaf3091efe43dce69222..HEAD` (merge-base re-derived at measurement time on this tree; develop had not moved between derivation and diff):
  ```
  .moai/reports/t700/plan-audit-iter2.md
  .moai/reports/t700/plan-audit.md
  .moai/specs/SPEC-GATEWAY-WEDGE-REROOT-001/acceptance.md
  .moai/specs/SPEC-GATEWAY-WEDGE-REROOT-001/plan.md
  .moai/specs/SPEC-GATEWAY-WEDGE-REROOT-001/progress.md
  .moai/specs/SPEC-GATEWAY-WEDGE-REROOT-001/spec.md
  internal/cli/gateway_reroot.go
  internal/cli/gateway_reroot_test.go
  internal/cli/gateway_session.go
  internal/gateway/conversation/reroot.go
  internal/gateway/translate/receipt_history_cause_test.go
  ```
  None of the five preserved files appears (`core.go`, `receipt_history.go`, `request.go`, `family.go`, `gateway_factory.go` — file-level zero-diff); no path under `internal/gateway/receipt/` appears. `gateway_session.go` is the plan-declared EXTEND target; `receipt_history_cause_test.go` is a test file (the freeze binds `receipt_history.go`); `conversation/reroot.go` is a new file (family.go untouched). PASS.
- **AC-WRR-014 Fork call-site count** — pre-card baseline at merge-base `643abfb8c`: exactly one production (non-test) call site, `internal/cli/gateway_session.go:204` (`families.Fork(`; all other matches are `*_test.go`). At HEAD: exactly one production call site, the same call at `gateway_session.go:232` (line shift only, from the `--reroot` wiring in the same function). Count unchanged 1 → 1; no new gateway-state-changing recovery path. PASS.

### M7 — run-phase verification (E1-E8; 2026-09-14)

- **E1 AC matrix**: AC-WRR-001 PASS (`TestReceiptHistoryClassifiesDesyncWedgeAfterUnpublishedTurn`), 002 PASS (`TestReceiptHistoryAcceptsTailRerootedWedgeReplay`), 003 PASS (`TestReceiptHistoryRejectsMidDropAfterTailReroot`), 004/005/006 PASS (existing locks unchanged), 007 PASS (`TestHistoryReplayErrorGoldenStrings`), 008/009/010/011/016 PASS (`TestGatewayReroot*` 11/11 after coverage completion), 012 PASS (diff-scope assertion above), 013 SKIPPED (M1 disposition — gap stands), 014 PASS (Fork count 1→1), 015 PASS (6/6 element greps ≥1). Totals: 15 PASS / 0 FAIL / 1 SKIPPED.
- **E2 cross-platform build**: `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (this run, this tree).
- **E3 coverage (this run)**: `go test -cover ./internal/gateway/translate/` → `coverage: 91.9% of statements` (package gate ≥85% PASS); `./internal/gateway/conversation/` → `80.7%` package (pre-existing baseline 77.0% — package-wide figure dominated by pre-existing untested branches, not this card's regression; this card's new code `TranscriptPath` measured `89.5%` after the reroot_test.go additions); cli scoped to the card's tests: `rerootGatewayTranscript 85.4%`, `rerootRefusalError 66.7%` (defensive fallback branch), `gatewayConversationPassthrough 100%`, `prepareGatewayConversation 65.1%` (pre-existing function; its `--reroot` branch covered by the wiring tests).
- **E4 subagent boundary**: `grep -rn 'AskUserQuestion\|mcp__askuser'` over the touched files (gateway_reroot.go, gateway_session.go, conversation/reroot.go), non-test non-comment → 0 matches.
- **E5 lint**: `golangci-lint run --new-from-rev=643abfb8cc536b1152efbaf3091efe43dce69222 ./internal/cli/... ./internal/gateway/...` → `0 issues.` exit 0 (no NEW issues vs the card base).
- **E6 commits + push state**: M2 `54a82200f`, M3 `a2f46eda3`, M4 `abbbb4eb9`, M6 `714aa35c6`, M7 (this commit). Lane does NOT push — the lead batch-pushes origin/develop; unpushed count reported in the completion report.
- **Full-suite honesty note**: a background full `go test ./internal/cli/` run (M3 tree; Go tree unchanged by M4/M6 doc commits) returned `FAIL github.com/modu-ai/moai-adk/internal/cli 601.239s` — the duration matches the 600s default go-test timeout abort shape on this loaded machine (the package historically measures ~1583s), but per-run detail was lost to output truncation, so the run classifies as UNRESOLVED locally, not as a pass and not as a verified defect. The change-scoped selector (`TestGatewaySession|TestGatewayConversation|TestGatewayReroot`, 17 top-level tests) is green in three separate runs, translate and conversation packages are green, and B5's pre-existing environmental failure (`TestAppServerSubprocessHTTPToolContinuation`) is named to-ignore. The full-suite verdict belongs to CI on the lead's origin/develop push (lane doctrine).
- **E8**: RED evidence recorded in the M3 section above.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-09-14
run_commit_sha: pending-backfill-run
run_status: complete
ac_pass_count: 15
ac_fail_count: 0
ac_skipped_count: 1 (AC-WRR-013 live probe — orchestrator skip decision, gap recorded in §E.2 M1)
preserve_list_post_run_count: 5 (all five PRESERVE files zero-diff; AC-WRR-012 assertion recorded verbatim)
new_warnings_or_lints_introduced: 0
cross_platform_build.native: exit 0
cross_platform_build.windows: exit 0
total_run_phase_files: 13
m1_to_mN_commit_strategy: one commit per milestone (M1 pre-flight by orchestrator; M2 test; M3 feat; M4 docs; M6 docs; M7 test+evidence)
full_suite_disposition: local full internal/cli run UNRESOLVED (600s default-timeout shape, detail lost to truncation); change-scoped selectors green ×3; full-suite verdict is CI's on the lead's origin/develop push

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
