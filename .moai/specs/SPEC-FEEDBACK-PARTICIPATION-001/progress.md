# Progress — SPEC-FEEDBACK-PARTICIPATION-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-10-07
- Artifacts reflect version 0.5.2 of the spec. Versions 0.3.0-0.5.1 carried the plan through plan-audit iterations 1-4 (0.62 → 0.73 → 0.79 → 0.86; the iteration reports are local gitignored files under `.moai/reports/t1498/` and are not committed). Version 0.5.2 is the iteration-4 delta repair, one clause: D40 — the lock-lifecycle age-only break alternative deleted; the stale-lock break fires only on a verified-dead owner and a live owner always blocks (an age-only break could discard a live slow owner's committed mutation); AC-018 pins the invariant in place with the live-owner arm and `TestLiveOwnerOfAgeExceededLockStillBlocks`; the acceptance preface broadens "user-scoped consent store" to the user-scoped stores. No new REQ or AC (25/25); DEC-7, DEC-8, and the identifier rename are intact; D35 remains the recorded open operator decision.
- Open clarifications: zero. All former markers are recorded decisions in plan.md section B (DEC-1 to DEC-8).
- SPEC ID rename: DONE (v0.4.0). The directory, the frontmatter `id`, and every cross-reference carry `SPEC-FEEDBACK-PARTICIPATION-001`; the former identifier `SPEC-FEEDBACK-ANON-PARTICIPATION-001` (audit finding D22) is retired and appears only in this line and in the spec HISTORY 0.4.0 row. The branch name stays `WT-feedback-optin-anon`.
- Branch state: rebased onto `origin/main` `5a9d34fbb`; the SPEC commits are `a8bf27c49` (v0.2.0), `c34e24cab` (v0.3.0), `647b5e789` (v0.4.0), `28a4a16bd` (branch-base line fix), `c957ecc9d` (v0.5.0), and `be75563eb` (v0.5.1). The v0.5.2 delta lands as the next commit on `WT-feedback-optin-anon`.
- Rebase re-pin (finding D29): E2, E2p, and E21 carry their re-measurement pins; E2p was refreshed again in this delta (D39: 22 on the committed tree at `c957ecc9d`, historical series recorded in the entry).
- Optional findings D31-D35: D31 carve-out clause added to REQ-ANON-002; D32 walk-authoritative inventory note added and the two post-rebase recover sites named in design section 2 and plan M3; D33 abbreviation rule noted in the AC matrix header; D34 left as self-disclosed (the E11 note carries its own caveat); D35 (the `REQ-ANON-` requirement token family) is presented for an operator decision and is unchanged by this revision.
- gh title-token search feasibility (finding D23) is recorded as the M5 first-test item in plan.md section C.
- Plan-audit verdict (iteration 5, delta): PASS-WITH-DEBT 0.88 — must_pass_failed 0, blocking_count 0; verdict file `.moai/reports/t1498/plan-audit-iter5.md`; codex receipt `rcpt-d56a4335c42963c8d744210f`; plan-artifact hash `cb94c1fa5e8bb4f034108930be8fb69de016c1d73baeeeb2c3ceddd2d3bdd584`; audited_sha `b00d2f7d6`. The plan phase is complete per the admission predicate; the enumerated debts are the binding run conditions below.

## Binding run conditions

- **D35 (run)** — the REQ-ANON token-family rename is an open OPERATOR decision (recorded in this file); run-phase proceeds with the existing tokens and must not pre-empt the operator; the orchestrator relays the decision request to the leader.
- **D38 (run)** — post-summary preview identity: the design sentence + the AC-017 test are run-phase work; alternatively record an explicit disposition if run-phase scope rules it out.
- **D34 (sync)** — E11 ledger cell's retired test-name selector refresh.
- **D41 (sync)** — spec-compact.md header version bump to match the current spec version.

## §E.2 Run-phase Evidence

_run in progress — M1..M7 landed through generation 4; the twelve-finding review-gate repair stack (9 from the M6 resume block + 3 lane-directed), M7 (skill bodies, docs wording, shipping guards), and the run-close verification batch are in (see the generation-4 table and §E.3)_

### Landed commits (this run, branch WT-feedback-optin-anon)

| SHA | What | Verification tail |
|---|---|---|
| 277ff11e2 | M1 taxonomy/payload/fingerprint (RED: build failure observed first) | bugreport pkg ok; AC-010 5/5 + AC-011 6/6 PASS lines; coverage 92.6% |
| 29de933e1 | M1 fix: frame filter `.go` substring misclassification (RED observed) | frames family + AC-010 named family 5/5 |
| fb0dbb2bf | M2 consent + setting surfaces (RED captured across 5 pkgs) | config 3/3, wizard 12/12 pins, cli 25 subtests, settings 8/8, web 10/10, E2 absence `:0/:0` exit 1 |
| 640859ed8 | fix: consent reader = exactly one YAML document (RED observed) | participation trio + settings/cli Participation families ok |
| cb7bc0b4e | fix: ParsePayload rejects trailing data (RED: 3 shapes accepted) | bugreport pkg ok; AC-011 family 7/7 |
| 22ec5871e | M3 detection/attribution wiring (RED: 15 unwired recover sites named by the walk) | bugreport 38 PASS; homestate/resilience/escalation/guardliveness/navigator ok; registry family 51 PASS; windows build 0 |
| 733868d60 | fix #5: hex-encode the boot identity (organic RED: lost update) | loss test 3x green; identity pins 2/2 |
| 6428f2ceb | fix #2: platform-split the boot identity (RED: GOOS=linux build failure) | native/linux/windows/darwin builds 0 |
| f4fd8b0d6 | fix #6: break verifies a fresh read; release verifies its own label (deterministic seam-RED) | 4 pins x2 green; feedback suite ok; 3 builds 0 |
| da753b219 | fix #3: the capture box covers the consent read (RED: FIFO block past 30s timeout) | FIFO pin 0.05s; pkg ok |
| 46dcc2c27 | fix #7: panicking-stub test exercises the real write path | family 4/4 |
| 90d3f1754 | fix #4: spool ceiling check+append under one Claim section (RED: >200 lines) | ceiling pin 3x green; pkg ok |
| f045c530d | M4 local pipeline/preview/log/withdrawal | outbox 10/10, cli 3/3, feedback+bugreport ok, 3 builds 0 |

### Review-gate fix stack (generation 2, lane-directed; each RED-first)

| SHA | What | Verification tail |
|---|---|---|
| f0ffd0252 | #7 purge propagates the spool removal error (RED: nil with the spool surviving) | outbox pkg ok |
| 9448ad690 | capture takes the CALLER's stack on the caller's goroutine (P1; RED: a real panic produced NO spool line) | bugreport pkg ok; probe pkg added for the frame-filter-shaped test |
| 62910fc91 | the D37 lock-owner machinery moves to internal/atomicfile (pure move; enabler for the spool lock) | feedback+atomicfile+bugreport+outbox ok; 3 builds 0 |
| 446032f5d | pid liveness treats os.ErrProcessDone as death (RED: finished reaped child judged alive — reproduced exactly) | atomicfile ok; 3 builds 0 |
| b6dc3791f | stale-lock break: rename + moved-bytes verify (superseded by c5040efd6; RED: rival's live lock disposed) | atomicfile+feedback+bugreport+outbox ok |
| 0999542b8 | the spool section lock gains the owner-verified reclaim (RED: claim budget exhausted on a dead-owner lock) | bugreport pkg ok; 3 builds 0 |
| c5040efd6 | the breaker's verdict-to-disposal span is ONE O_EXCL section (P1 rename race per the gate's B/C repro: RED "C acquired while B held") | atomicfile ok under -count=3 -race; feedback 16x50 reclaim pin ok under -race x3; 3 builds 0 |
| 60de8bfc0 | the FIFO deadline fixture compiles only on unix (P1; GOOS=windows go test -c failed) | GOOS=windows AND GOOS=linux `go vet ./internal/bugreport/...` exit 0; FIFO test passes on darwin |
| 9236c5cd6 | hook_timeout routes to its ambiguous verdict before the chain rows (P2; RED: Attribute returned environment, the signal was silently discarded) | attribution family + full internal/hook suite ok (294s, env-scrubbed — TestStaleRunNotice lanes are this session's MOAI_KANBAN_*/MOAI_FACTORY_* env, not the change) |
| 5e9d7e1d0 | dedupe check through ledger record is ONE queue-lock section (RED: 2 concurrent drains double-enqueued; caps breached) | both green -count=3 -race; outbox ok; vet clean |
| e07abc38e | the drain enforces the queue bound at enqueue (RED: queue grew to 21 over the cap) | green -count=3 -race; outbox ok |
| 728a1904b | the drain consumes only the spool batch it read (RED: compile-RED on the seams, then the late capture survived) | outbox+bugreport ok; outbox -count=3 -race ok; 3 builds 0 |
| 65380b6e8 | the spool entry carries the CAPTURE-TIME build identity (RED: a v3.2.0 capture flushed by a v9.9.9 binary queued the v9.9.9 fingerprint) | outbox+bugreport ok; 3 builds 0. Note for sync: design section 1's bugreport import sentence gains pkg/version (cycle-free, outside AC-025's guard set) and the spool schema gains version/commit — for manager-spec to fold into the design body |
| d6613a015 | M5 publication through the user's gh, deterministic text | publish 15/15; AC-003 four names PASS; outbox+feedback+cli families ok; 3 builds 0; lint 0 issues on touched packages |
| 131aeb712 | the drain honors its context through every lock wait (gen-2's final review-gate fix, landed during the gen-3 handoff window) | atomicfile+feedback+bugreport+publish+outbox ok; 3 builds 0; the 200ms-deadline repro 13.31s → 0.20s |
| 9bb474a1f | docs: record the drain-context fix + refresh the M6 resume block | docs commit |
| acfeb7259 | review-gate 1/5: the breaker-marker reclaim is a CAS guarded by its own .reclaim marker | atomicfile ok + ok -race; consumers ok; 3 builds 0 |
| 1b3735ce3 | review-gate 5/5: the marker format round-trips the multi-field detail losslessly (QuoteMarkerValue/SplitMarkerTokens) | publish+outbox ok (golden byte-identical); 3 builds 0 |
| c309681a1 | review-gate 1/5 (P1): the sender re-validates the stored body and regenerates what it publishes (revalidate.go) | publish+outbox+feedback ok; 3 builds 0; RED: tampered body reached gh |
| 58cec539c | review-gate 3/5 (P2): per-item cross-process send ownership (outbox.ClaimItemSend) | publish+outbox ok; ownership pair -count=3; 3 builds 0; RED: two flushes each sent |
| 60deba377 | review-gate 4/5 (P2): the drain consumes only the prefix it processed (PrefixLenForEntries) | outbox+bugreport ok; drain family -count=3; 3 builds 0 |
| e397ec59d | review-gate residual 1: the sender re-checks the live queue after taking an item's claim | publish ok; 3 builds 0; RED: flush B re-sent flush A's item |
| 026c7ef7d | review-gate residual 2: an orphaned ledger record re-queues instead of deduping the report away | outbox+feedback ok; RED both shapes: report deduped away and lost |
| 1159d38ea | review-gate residual 3: every marker reclaim takes its own guard — the .reclaim path included, depth-capped | atomicfile ok + ok -race; 3 builds 0; RED: rival's live guard deleted |
| 5f00c7aac | hardening 1 (P1): the recursive reclaim honors the caller's cancellation (BreakStaleLockContext) | atomicfile ok; 3 builds 0; pre-cancelled reclaim = 0 verdict reads |
| 270e91599 | hardening 2 (P2): a failed queue save no longer double-counts into the rolling caps (rollbackLedgerRecord, ctx-bounded) | outbox ok; RED: 3 phantom QueuedAt after 3 failed saves |
| 834930afd | hardening 3 (P2): a recorded send is reconciled, never re-sent (sent-first ordering + reconcile rule) | publish ok; RED: creates=0 comments=1 re-sent |
| 5dc638bbc | hardening 4 (P2): the whole flush honors its deadline — complete/drop/fail take MutateContext | publish+outbox ok; RED: 1.15s on a 200ms deadline |
| 9693fcab9 | hardening amendment: a past sent record suppresses only inside the DEC-3 window (SentHistoryHasFingerprintWithin) | publish+outbox+feedback ok; 3 builds 0; RED: 8-day-old row suppressed a new report |
| 131aeb712 | the drain honors its context through every lock wait (RED: a 200ms-deadline drain ran 13.31s against a live holder — the gate measured 13.36) | green in 0.20s, 3/3 -race; all five touched suites ok; 3 builds 0; ClaimSection/MutateContext carry ctx |

### Review-gate repair stack — generation 4 (M6 resume block 9 findings + 3 lane-directed; each RED-first)

| SHA | What | RED tail / verification |
|---|---|---|
| 368cf5609 | finding 1 (P1): the REUSED stored summary passes validateSummary at send time; failure → deterministic template, decision recorded | "the unvalidated stored summary reached the public issue body" (×2 shapes); publish ok; trio 0 |
| ffcb7029d | finding 4 (P2): the ledger rollback verifies its own reservation (recorded stamp == attempt stamp) before deleting | "the prior success's fingerprint record was deleted"; outbox ok incl. TestFailedSaves family |
| fa6029007 | finding 8 (P2): an attempt-exhausted discard is recorded TERMINALLY in the ledger (Discarded map); recovery re-queues only genuinely unfinished reservations | "queue holds 1 items, want the discarded report NOT re-enrolled" + "left no terminal discard marker"; both suites ok |
| 3a0377137 | findings 5+9 (P2): orphan recovery ADOPTS the orphaned reservation's cap slot (GlobalCapsAllowedExcluding + RecordQueuedAdopting) | "queue holds 0 items — cap-judged against its own orphan slot" + "2 QueuedAt entries for ONE report"; outbox ok |
| a0029c056 + 8a651afe9 | finding 3 (P2): the model-call budget persists per user (modelcalls.json), counted atomically cross-process (outbox.AllowAndRecordModelCall, one queue-locked mutation); in-memory budget deleted; PurgeStores includes it | "a new sender called the model again (total 7) — the budget reset per sender"; both suites ok |
| 80dfbec3f | finding 2 (P2): the consent check precedes the gh auth status network probe | "gh auth status ran 1 time(s) with participation OFF"; publish ok |
| 8a9feed5e | finding 7 (P2): model byte caps reference config.DefaultBugreportModelInput/OutputMaxBytes (4096/2048), not local 8192/4096 | "model input cap = 8192, want the central 4096"; publish ok |
| 06dbfed28 | finding 6 (P2): the summarizer env is os.Environ() through the audit scrub (claudeParticipationEnv), not scrubClaudeAuditEnv(nil) | compile RED "undefined: claudeParticipationEnv"; HOME/PATH present, ANTHROPIC_API_KEY absent, validateClaudeAuditEnv nil |
| d13574925 | lane finding A (P2): the consent reader refuses non-regular files (stat) and bounds open+read (DefaultParticipationConsentReadTimeBox 100ms / MaxBytes 4096) | "the consent reader blocked 10.002125125s on a FIFO consent file"; config ok |
| a8c251596 | lane finding B (P2): the attempt limit is re-checked on the LIVE item after the claim | "searches=1 creates=1 — an item exhausted between the snapshot and the claim was sent anyway"; publish ok |
| 1afbe0e7d | lane finding C (P2): an expired discard marker no longer suppresses (TerminallyDiscardedWithin, window-scoped like the sent-record reconcile) | "an 8-day-old discard record suppressed a genuinely unfinished reservation and the report was consumed"; outbox ok |

### M7 (generation 4, mechanical)

| SHA | What | Verification tail |
|---|---|---|
| 0d50c2713 | AC-023: three feedback skill-body copies gain the relationship-to-participation section (Template-First: authored in the template copy, synced byte-identical; make build regenerated catalog.yaml); four docs-site pages drop the "created automatically" wording (E18: four :0) and positively state the corrected per-locale wording (E19: en 1 / ko 2 / ja 2 / zh 2) | grep -c participation → 4/4/4; TestTemplateNeutralityAudit PASS; make build 0 |
| 1966ca0d0 | AC-024: TestNoAutoRepairArtifactsShipped (templates + plugins/moai walk: no auto-repair path token, no moai-bugreport content) + canary; TestParticipationQuestionRequiresSender + missing-sender canary. The first canary run caught the guard judging the ABSOLUTE path (its own temp-dir name tripped the token check) — the walk now judges the root-relative path | E22a/E22b were empty-grep red; both guards + canaries PASS; wizard suite ok |

Gate item 3 (re-verify at the new HEAD): the double-ClaimSection dead-owner repro family at HEAD 9236c5cd6 — TestBreakerExcludesRivals..., TestBreakNeverDisposes..., TestBreakAborts..., TestBreakStillFires..., TestBreakGateAborts... x -count=3 -race = 15/15 PASS; TestStaleLockReclaimDoesNotDeleteTheNewLock (16 concurrent ClaimSection callers against a dead-owner fixture, no release) x -count=3 -race = 3/3 PASS.

### Generation-4 verification batch (run close, HEAD 1afbe0e7d)

| Item | Command | Observed |
|---|---|---|
| Full suites | `go test ./internal/feedback/... ./internal/bugreport/... -count=1` | ok × 4 packages (feedback, outbox, publish, bugreport) |
| Wizard + cli families | `go test ./internal/cli/wizard/ -count=1` + targeted cli selectors | wizard ok; 6/6 PASS (TestFlushWiresClaudeSummarizer, TestParticipationSummarizerEnvCarriesTheBaseEnvironment, TestUpdateParticipationStep, TestUpdateFlagInventoryClassified, TestApplyParticipationFromWizard, TestEveryRegisteredHookHandlerHasBugreportName) |
| Coverage (union -coverpkg across the four packages, all their tests) | `go test -coverpkg=<feedback,feedback/outbox,feedback/publish,bugreport> ./internal/feedback/... ./internal/bugreport/... -coverprofile=...` then `go tool cover -func` | bugreport 84.4%, feedback 86.8%, outbox 81.8%, publish 76.8% — union total 82.4% |
| Boundary grep | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/feedback/ internal/bugreport/ internal/config/participation_user.go` (non-test, non-comment) | 0 hits |
| Lint (NEW vs baseline: baseline was 0 issues on these packages at generation start) | `golangci-lint run --timeout=2m ./internal/feedback/... ./internal/bugreport/... ./internal/config/...` | 0 issues |
| Build trio + vet | `go build ./...` / GOOS=linux / GOOS=windows + `go vet` on the touched packages | all exit 0 at HEAD 1afbe0e7d |
| Divergence at close | `git fetch origin main` + `git rev-list --count --left-right origin/main...HEAD` | `106 73` — the card-worktree baseline (branch is develop-descended; integration is the leader's develop-merge, no push from this lane) |

Publish coverage gap (honest): ghrunner.go is 0% — the production `gh` exec paths are exercised by the Runner seam's tests, not unit-executed (no real gh in tests); template.go (1 stmt) same family. revalidate.go 68%. These figures are the union across all four packages' test binaries; the internal/cli model file (feedback_participation_model.go) is outside this profile (cli suite not in the -coverpkg run — load discipline).

### Gaps (explicitly unobserved)

- internal/cli FULL suite: two local runs hit the go-test wall (601s default; 1801s at `-timeout 30m`) on a machine with three other lanes running suites; a 6-minute verbose diagnostic showed 918 tests progressing normally (no hang; the timeout fired mid `TestCodexReviewScope...`, unrelated to this SPEC). All M2/M3-affected cli families pass under targeted selectors, and the M5-affected families (TestParticipation*, flush/update wiring) pass. CI owns the repository-wide verdict.
- internal/hook FULL suite: 972s FAIL in a background run — 1324 PASS with 3 failures (TestAstgrepCorpusRunDoesNotSkip 120s timeout; TestSessionStart_DeferredScanJoinsWithinBound timing; TestStaleRunNoticeFactoryLegacyLabel "context deadline exceeded"). Env-scrubbed re-run (the lane-env lesson: MOAI_KANBAN_ID pollutes run-state-gated tests) turned the two stale-run notice tests GREEN; the remaining two are load-correlated timing tests in files M3 never touched (empty diff 640859ed8..22ec5871e for stale_run_gate.go, session_stale_run.go, internal/factory). At generation 2's HEAD the FULL env-scrubbed hook suite ran ok (294s) including those timing tests. CI owns the verdict.
- TestWeb TodoGraphView flaked once in a full-web run, passes in isolation and in a later full run (89.6s ok).
- D23 (gh title-token search): the read-only half is measured — `gh issue list --repo modu-ai/moai-adk --state all --search "<16-hex> in:title" --json number,title --limit 10` round-trips in ~0.7s and returns `[]` for an absent token (probe token ee5e69700339b582, 2026-10-07). The fresh-issue index-latency half requires a throwaway PUBLIC issue — an irreversible external-shared action, operator-held — so the acceptance-stated edge stands as the recorded decision: two concurrent first filers may create two issues; the consumer merges by fingerprint; comments are append-only so counts do not lose updates.

### M5 disposition notes (binding conditions)

- D35: the `REQ-ANON-` token family is UNCHANGED — no requirement token was renamed or removed in any run-phase commit; the open operator decision stands.
- D38 (post-summary preview identity): the M5-owned half is IMPLEMENTED — TestPreviewMatchesCreateBytes (AC-020) pins the preview bytes equal to the create bytes the sender hands to gh. The summary-augmented body is AC-017's subject, M6-owned per the plan: M6's sender path slots the validated summary AHEAD of CreateBody and the AC-017 tests pin it; the preview keeps printing the queued render (the pre-summary text) by the design sentence. Disposition: satisfied as split by plan M5/M6; no design change needed.

### Resume block (next session: the M6 review-gate stack, then M7)

```
✂──── 여기부터 복사 ────✂
ultrathink. SPEC-FEEDBACK-PARTICIPATION-001 run resuming: the M6 review-gate stack (7 findings, RECEIVED NOT FIXED), then M7.
mode: serial
applied lessons: feedback_glm_lane_env_pollutes_claude_audit (scrub MOAI_KANBAN_* before suites)

Preconditions:
1) git rev-parse --short HEAD → 20e3ea022-or-later on WT-feedback-optin-anon; tree clean; M6 LANDED at c011045ab (publish ok; 3 builds 0; lint ineffassign fixed at 20e3ea022)
2) go test ./internal/feedback/publish/ ./internal/feedback/outbox/ -count=1 → ok

Run: RED-first, own commit each — (1) P1 sender.go: itemSummary REUSES a stored Summary WITHOUT validateSummary (a saved summary carrying a token goes public; validate the reused summary, template-fallback on failure); (3) P2 budget resets per Sender instance (persist reservations in the user-scoped store, atomic cross-process); (4) P2 rollback fires for lock-acquisition errors that created NO reservation → verify-before-rollback (a prior success's record gets deleted → double enqueue); (5) P2 orphan recovery leaves the FAILED reservation in QueuedAt → the recovery target gets capped+consumed (reuse or exclude the reservation); (2) P2 gh auth status (network) runs before the Send consent check; (6) P2 scrubClaudeAuditEnv(nil) builds an empty env — use os.Environ(); (7) P2 model byte caps are local 8192/4096, diverging from config/defaults.go — centralize. **+2 refinement (2026-10-08, review gate)**: (8) P2 drain.go:538 — a send-attempt-EXHAUSTED discard leaves no queue/sent record, so a later drain mistakes it for an UNFINISHED reservation and re-enrolls it inside the 7-day window, resetting attempts AND the model budget → record a TERMINAL discard marker in the ledger; recovery re-queues only genuinely unfinished reservations (RED: discarded-report fixture); (9) P2 drain.go:545 — orphan recovery must ADOPT the existing reservation's cap usage or cancel it first, never count it AND discard the item (RED: third-reservation fixture — the recovery target was cap-judged and DELETED from the spool; refines finding 5, same neighborhood). Sequence: P1 first, then the reservation/recovery set (3/4/5/8/9), then 2/6/7, then M7.

After merge: /moai sync SPEC-FEEDBACK-PARTICIPATION-001
✂──── 여기까지 복사 ────✂
```


```
✂──── 여기부터 복사 ────✂
ultrathink. SPEC-FEEDBACK-PARTICIPATION-001 run M6 resuming.
mode: serial
applied lessons: feedback_glm_lane_env_pollutes_claude_audit (scrub MOAI_KANBAN_*/MOAI_FACTORY_* before hook/cli suites), feedback_codex_task_turn_bound_use_raw_exec

Preconditions:
1) git rev-parse --short HEAD → tip 131aeb712 (M5 d6613a015 + the ctx fix) on WT-feedback-optin-anon; tree clean
2) go build ./... && GOOS=linux GOARCH=amd64 go build ./... && GOOS=windows GOARCH=amd64 go build ./... → all exit 0
3) go test ./internal/feedback/publish/ ./internal/feedback/outbox/ -count=1 → ok (15 + 14 tests)

Run: RED-first M6 per acceptance.md — in internal/feedback/publish create
model.go (the Summarizer interface owned by publish) + budget.go (daily cap 6
from DEC-3) and internal/cli/feedback_participation_model.go (the production
implementation over the existing headless claude runner — claudeAuditArgs flag
set minus the audit --json-schema; injected at the flush call in
internal/cli/feedback_participation.go runParticipationFlushWork). Tests:
TestPublishCallsModelOnceAndValidates, TestPublishFallsBackToTemplate,
TestModelInputIsPayloadFieldsOnly (AC-017; model input golden = payload fields
only), TestSummaryPersistedBeforeCreateAndReusedOnRetry,
TestModelCallBoundPerQueueItem, TestMarkerBoundsCrashWindowRecall,
TestKillMidMutateLeavesQueueWritable, TestLiveOwnerOfAgeExceededLockStillBlocks
(AC-018; the stale-lock pieces of the fixture family are already green —
verify which names exist and add only the missing ones),
TestLLMBudgetZeroCalls, TestLLMBudgetPositiveControl, TestDailyModelCallCap
(AC-019), TestBugreportAndOutboxImportAllowlist,
TestPublishImportAllowlist, TestModelSeamIsInjectedNotImported +
internal/cli TestFlushWiresClaudeSummarizer (AC-025). M6 first test: record
that the claudeAuditArgs flag set is accepted by the installed claude for a
summary prompt, and that an unauthenticated session yields the template
fallback. The summary_requested marker lands here (D28). The channel is
decided (DEC-6); ambiguous stays LOCAL-ONLY (DEC-7 — no model call, no send;
the path must not exist). Then M7 (mechanical): skill bodies x3
(Template-First: author internal/template/templates/.claude/skills/moai/
workflows/feedback.md, sync .claude/skills/ + plugins/moai/ copies, make
build; make agents-emit/commands-emit ONLY if command/agent definitions
change — none should), docs-site 4 locales
(docs-site/content/{en,ko,ja,zh}/utility-commands/moai-feedback.md: remove
the "created automatically" wording per E18, add the per-locale positive
pattern per E19), internal/template/auto_repair_guard_test.go
(TestNoAutoRepairArtifactsShipped + canary Test), internal/cli/wizard/
participation_guard_test.go (TestParticipationQuestionRequiresSender +
synthetic-tree canary), make build, AC-023/024 verify lines. Then §E.2 rows
for M6/M7 + §E.3 Run-phase Audit-Ready Signal, E1-E8 verification batch.

After merge: /moai sync SPEC-FEEDBACK-PARTICIPATION-001
✂──── 여기까지 복사 ────✂
```

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-08
run_commit_sha: "1afbe0e7d"   # run tip, branch WT-feedback-optin-anon; nothing pushed
run_status: complete            # run-phase complete; sync pending (manager-docs owns the close)
ac_pass_count: 25               # recorded in §E.2 across M1..M7 generations; generation 4 re-measured AC-017/018/019/023/024/025 + all twelve repair families (40 named `--- PASS` lines in the close batch, table above)
ac_fail_count: 0
preserve_list_post_run_count: 0 # five plan-artifact bodies byte-untouched: git diff b00d2f7d6..HEAD -- .moai/specs/ shows only progress.md (evidence) and spec.md's single sanctioned status line draft→in-progress (M1, Status Transition Ownership Matrix)
l44_pre_commit_fetch: "git fetch origin main; git rev-list --count --left-right origin/main...HEAD → 106 73 at HEAD 1afbe0e7d — the card-worktree baseline (develop-descended branch; integration is the leader's develop-merge; this lane does not push)"
l44_post_push_fetch: "n/a — no push from this lane (B9); remote landing is the leader's batch push + CI verdict"
new_warnings_or_lints_introduced: 0  # golangci-lint 0 issues on internal/feedback/... internal/bugreport/... internal/config/... (baseline at generation start was also 0); go vet clean on all touched packages
cross_platform_build:
  native: exit 0
  linux: exit 0   # GOOS=linux GOARCH=amd64 go build ./...
  windows: exit 0 # GOOS=windows GOARCH=amd64 go build ./...
total_run_phase_files: 154  # git diff --name-only b00d2f7d6..HEAD (154 files, +13919/-78) — M1..M7 generations combined
m1_to_mN_commit_strategy: one commit per RED-GREEN unit (RED observed first, verbatim output in the commit body), docs/resume updates as separate docs commits; generation 4 = 14 commits (12 repair + 2 M7)
```

Run-phase gaps carried from earlier generations (unchanged, CI owns the verdicts): internal/cli FULL suite and internal/hook FULL suite were not re-run by generation 4 (load discipline; gen-1/gen-2 records in §E.2 Gaps stand); the publish union-coverage figure excludes the cli model file; D23's fresh-issue index-latency half stands as the recorded operator decision.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-10-08
sync_commit_sha: "52e1b2fdb"   # the sync-phase commit (backfilled per D3; subject: chore(SPEC-FEEDBACK-PARTICIPATION-001): sync-phase artifacts + 3-phase close (card t1498))
sync_status: complete
ac_pass_count: 25                          # unchanged from §E.3 — the sync phase adds no criteria; five plan-artifact bodies stayed frozen (only the spec.md frontmatter status/updated transition rode this commit)
ac_fail_count: 0
b12_self_test_a: "pre-emission grep: grep -c 'SPEC-FEEDBACK-PARTICIPATION-001' CHANGELOG.md → 0 (exit 1) — no duplicate entry existed before emission"
b12_self_test_b: "AC count: MOAI-AC-COUNTER on acceptance.md → live=25 excluded=0 ambiguous=0; the CHANGELOG entry references the same 25 criteria (AC-001..025)"
b12_self_test_c: "file-path verification: every path named in the entry verified by ls — internal/bugreport/, internal/feedback/, internal/feedback/outbox/, internal/feedback/publish/, internal/feedback/publish/model.go, internal/cli/feedback_participation.go, internal/cli/feedback_participation_model.go, 3 skill-body copies (internal/template/templates/.claude/skills/moai/workflows/feedback.md + .claude/skills/ + plugins/moai/), docs-site/content/{en,ko}/utility-commands/moai-feedback.md"
changelog_entry_position: "CHANGELOG.md [Unreleased] → Added, first entry (above SPEC-MEMORY-FOLD-BUDGET-001)"
frontmatter_status_transitions:
  spec_md_status: "in-progress → completed"   # merged 3-phase close on this single sync commit (no separate implemented-staging commit; draft→in-progress was M1's sanctioned run-phase edit)
  spec_md_updated: "2026-10-07 → 2026-10-08"
  plan_acceptance_updated: "n/a — plan.md and acceptance.md carry no frontmatter block at all (Artifact Statelessness permits omission); no updated: field exists to refresh"
canary_compliance_check:
  shipped: true   # this SPEC defines forward-looking policies its own sync-shipped tests guard: the no-auto-repair-artifact guard and TestParticipationQuestionRequiresSender (AC-024) hold the release-ordering invariant; both green in §E.2 M7 and untouched by the sync phase
```

## §F Phase 4 Mode Selection

- Input parameters: tier L; scope >15 files (new internal/bugreport; internal/feedback, internal/cli, internal/config, internal/settings, internal/web; three skill-body copies; four docs pages); domains: Go source + embedded templates + docs (multi-domain); language mix Go-dominant; concurrency benefit LOW (coding-heavy, new code); agent-team prereqs: not requested.
- Mode evaluation: direct — no (semantic new-code work). fanout — no (coding-heavy per Anthropic's coding-task parallelism caveat). sweep — no (semantic multi-rule work, not mechanical-uniform). agent-team — no (explicit operator request absent). serial — selected.
- Decision: serial
- Justification: coding-heavy new-code implementation is the canonical serial case; milestones M1→M7 are dependency-ordered by decision reversibility; the plan's M2/M3 disjoint-file parallel carve-out is not exercised (one writer per tree; the shared integration point internal/config/defaults.go serializes it anyway).

## §G Kickoff Decision Record

decision record: decided_by=lane-4(glm)+orchestrator evidence_refs=.moai/reports/t1498/plan-audit-iter5.md (AUDIT-VERDICT: PASS-WITH-DEBT spec=SPEC-FEEDBACK-PARTICIPATION-001, score 0.88, must_pass_failed=0, blocking_count=0, receipts=rcpt-d56a4335c42963c8d744210f, plan_artifact_hash=cb94c1fa5e8bb4f034108930be8fb69de016c1d73baeeeb2c3ceddd2d3bdd584, audited_sha=b00d2f7d6) + convergence_check ok (moai verify audit-plan, unmet=[]) + progress.md §E.1 audit-ready + §F Mode Selection serial ladder_path=plan→run Kickoff, §9.1 autonomous form — keep-set categories absent (implementation is worktree-isolated; the terminal push+PR is the dispatch-designated path of card t1498, 2026-10-07). The decision-board mirror is a leader handoff (the board's record verb is leader-only).
## main reabsorption + gate P2 repairs (gen-4, 2026-10-08)
merge origin/main 71852d7f2: 4 conflicts resolved — CHANGELOG interleave, update.go dual-intent (main t1527 defer comments + card REQ-ANON-004 wiring), catalog.yaml hashes regenerated via gen-catalog-hashes --all, plugins/.../feedback.md deleted per main carrier retirement.
gate P2 x3 (turn-end codex review on merged tree): outbox.log history read bounded (drain.go FIFO refuse + time-box + cap), corrupted queue no longer blocks purge (purge.go), ledger read limited to cap+1 before allocation (ledger.go LimitReader). RED measured: FIFO blocked 10.002s; corrupt queue abort error observed. GREEN measured: both new tests pass 0.79s; card suites feedback/outbox/publish/bugreport/atomicfile all ok.
build ./... exit 0, vet on cli+template+feedback+bugreport+atomicfile+config exit 0.
## gate round 2: outbox bounded-I/O class sweep (2026-10-08)
P2 x4 on the card code (outbox) — 3rd instance of the bounded-read class; repaired as a CLASS this time: purge lock-acquire failures now propagate (only json parse failures are ignored, isQueueCorruptionError), AppendOutbox refuses non-regular + time-boxed open, sent-history read keeps the LOG TAIL past the cap (Seek+LimitReader, one cap-sized allocation, recent dedupe rows survive; cut rows fail JSON like any malformed line). RED measured: purge returned nil under a held lock (1.39s); append blocked at drain.go:99 (test timeout panic stack pins the line); oversized log lost the most recent row. GREEN: 5 targeted tests ok 1.629s incl. both prior-round tests; -race suite feedback/outbox/publish/bugreport/atomicfile all ok. gofmt clean.
NOT in card scope: init.go:947 user-asset install retry path (main-side t1509 area, came in via the merge) — relayed to the leader for disposition.
## gate round 3 + CI r6 repairs (2026-10-08)
Gate P2 x5 (card code): queue save-side cap (EnqueueMasked rejects an over-cap enqueue with ErrQueueFull; ErrQueueUnreadable sentinel now wraps the load-step parse+cap failures and the purge matches it, so an oversized queue purges while a held lock still propagates), spool generation counter (bugreport.BumpSpoolGeneration/SpoolGeneration — purge bumps FIRST; the drain re-checks before each item and stops on a mid-batch purge), sectionRereadFn bounded (FIFO refused without opening + 4KiB cap), spool consume reread bounded (boundedSpoolReread — same refusal, cap, time box as the first read). RED measured x5 (resurrect 1 item / over-cap purge aborted / owner read blocked 10.0s / consume reread blocked 10.0s / over-cap enqueue accepted); GREEN: all five pass, -race suite ok, gofmt clean. NOT in card scope: doctor_harness.go:57 user-vs-project skill path (main side) — relayed.
CI r6 (leader): 3 of 4 reproduced LOCALLY as real defects — feedback.participation schema field had no TUI bridge entry (schemaFieldBridge + profileSetupText ParticipationTitle/Desc, four locales) — repaired, all three green. TestUpdateParticipationStep does NOT reproduce locally (green) — CI-only divergence, evidence recorded, left to CI re-judgment on the new head.
