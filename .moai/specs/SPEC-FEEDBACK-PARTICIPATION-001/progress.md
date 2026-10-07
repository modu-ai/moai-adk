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

_run in progress — M1..M5 + the full review-gate fix stack (14 commits) landed; M6 is the resume point (see the resume block at the end of this section)_

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

Gate item 3 (re-verify at the new HEAD): the double-ClaimSection dead-owner repro family at HEAD 9236c5cd6 — TestBreakerExcludesRivals..., TestBreakNeverDisposes..., TestBreakAborts..., TestBreakStillFires..., TestBreakGateAborts... x -count=3 -race = 15/15 PASS; TestStaleLockReclaimDoesNotDeleteTheNewLock (16 concurrent ClaimSection callers against a dead-owner fixture, no release) x -count=3 -race = 3/3 PASS.

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

Run: RED-first, own commit each — (1) P1 sender.go: itemSummary REUSES a stored Summary WITHOUT validateSummary (a saved summary carrying a token goes public; validate the reused summary, template-fallback on failure); (3) P2 budget resets per Sender instance (persist reservations in the user-scoped store, atomic cross-process); (4) P2 rollback fires for lock-acquisition errors that created NO reservation → verify-before-rollback (a prior success's record gets deleted → double enqueue); (5) P2 orphan recovery leaves the FAILED reservation in QueuedAt → the recovery target gets capped+consumed (reuse or exclude the reservation); (2) P2 gh auth status (network) runs before the Send consent check; (6) P2 scrubClaudeAuditEnv(nil) builds an empty env — use os.Environ(); (7) P2 model byte caps are local 8192/4096, diverging from config/defaults.go — centralize. THEN M7 per the earlier resume block (skill bodies x3 Template-First + make build; docs-site 4 locales E18/E19; auto_repair guard + wizard coexistence guard; AC-023/024). The M6 live-CLI record (claudeAuditArgs-minus-schema accepted; unauthenticated → template) was NOT taken — carry it into M7's §E evidence or record a disposition. Then §E.3 + the E1-E8 batch.

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

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_

## §F Phase 4 Mode Selection

- Input parameters: tier L; scope >15 files (new internal/bugreport; internal/feedback, internal/cli, internal/config, internal/settings, internal/web; three skill-body copies; four docs pages); domains: Go source + embedded templates + docs (multi-domain); language mix Go-dominant; concurrency benefit LOW (coding-heavy, new code); agent-team prereqs: not requested.
- Mode evaluation: direct — no (semantic new-code work). fanout — no (coding-heavy per Anthropic's coding-task parallelism caveat). sweep — no (semantic multi-rule work, not mechanical-uniform). agent-team — no (explicit operator request absent). serial — selected.
- Decision: serial
- Justification: coding-heavy new-code implementation is the canonical serial case; milestones M1→M7 are dependency-ordered by decision reversibility; the plan's M2/M3 disjoint-file parallel carve-out is not exercised (one writer per tree; the shared integration point internal/config/defaults.go serializes it anyway).

## §G Kickoff Decision Record

decision record: decided_by=lane-4(glm)+orchestrator evidence_refs=.moai/reports/t1498/plan-audit-iter5.md (AUDIT-VERDICT: PASS-WITH-DEBT spec=SPEC-FEEDBACK-PARTICIPATION-001, score 0.88, must_pass_failed=0, blocking_count=0, receipts=rcpt-d56a4335c42963c8d744210f, plan_artifact_hash=cb94c1fa5e8bb4f034108930be8fb69de016c1d73baeeeb2c3ceddd2d3bdd584, audited_sha=b00d2f7d6) + convergence_check ok (moai verify audit-plan, unmet=[]) + progress.md §E.1 audit-ready + §F Mode Selection serial ladder_path=plan→run Kickoff, §9.1 autonomous form — keep-set categories absent (implementation is worktree-isolated; the terminal push+PR is the dispatch-designated path of card t1498, 2026-10-07). The decision-board mirror is a leader handoff (the board's record verb is leader-only).
