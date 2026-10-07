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

_run in progress — M1..M4 + eight review-gate fixes landed; M5 is the resume point (see the resume block at the end of this section)_

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

### Gaps (explicitly unobserved)

- internal/cli FULL suite: two local runs hit the go-test wall (601s default; 1801s at `-timeout 30m`) on a machine with three other lanes running suites; a 6-minute verbose diagnostic showed 918 tests progressing normally (no hang; the timeout fired mid `TestCodexReviewScope...`, unrelated to this SPEC). All M2/M3-affected cli families pass under targeted selectors. CI owns the repository-wide verdict.
- internal/hook FULL suite: 972s FAIL in a background run — 1324 PASS with 3 failures (TestAstgrepCorpusRunDoesNotSkip 120s timeout; TestSessionStart_DeferredScanJoinsWithinBound timing; TestStaleRunNoticeFactoryLegacyLabel "context deadline exceeded"). Env-scrubbed re-run (the lane-env lesson: MOAI_KANBAN_ID pollutes run-state-gated tests) turned the two stale-run notice tests GREEN; the remaining two are load-correlated timing tests in files M3 never touched (empty diff 640859ed8..22ec5871e for stale_run_gate.go, session_stale_run.go, internal/factory). CI owns the verdict.
- TestWeb TodoGraphView flaked once in a full-web run, passes in isolation and in a later full run (89.6s ok).

### Resume block (next session, M5)

```
✂──── 여기부터 복사 ────✂
ultrathink. SPEC-FEEDBACK-PARTICIPATION-001 run M5 resuming.
mode: serial
applied lessons: feedback_glm_lane_env_pollutes_claude_audit, verification-claim-integrity §3.1 refusal recording

Preconditions:
1) git rev-parse --short HEAD → M4 tip f045c530d on WT-feedback-optin-anon; tree clean
2) go build ./... && GOOS=linux GOARCH=amd64 go build ./... && GOOS=windows GOARCH=amd64 go build ./... → all exit 0
3) go test ./internal/feedback/outbox/ -count=1 → ok (10 tests)

Run: RED-first M5 per acceptance.md — create internal/feedback/publish tests first
(TestSenderChecksConsentPerItem, TestSenderQuietWithoutGh, TestSenderNeverRunsOnHookPath,
TestSenderTimeBox, TestExistingFingerprintGetsOccurrenceComment, TestExactTitleKeyOnly,
TestClosedIssueStillCounts, TestOccurrenceCapSkipsComment, TestIssueContractRoundTrip,
TestSameFingerprintSameTitleKey, TestNoLabelsNoBodyEdit, TestOccurrenceMarkersAreUntrustedInput,
TestPreviewMatchesCreateBytes, TestSenderNoopWhenParticipationOff,
TestSenderIgnoresTrackedFileConsentAndRepository), capture the verbatim RED, then implement
ghrunner.go (os/exec ONLY there), sender.go, lookup.go, contract.go, template.go,
testdata/bugreport_issue_v1.golden; wire publish.Flush into the flush call site in
internal/cli/feedback_participation.go runParticipationFlush (drain exists); M6: model.go,
budget.go, cli/feedback_participation_model.go + AC-017/018/019/025 tests; M7: skill bodies
x3, docs-site 4 locales, auto_repair + coexistence guards, make build.

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
