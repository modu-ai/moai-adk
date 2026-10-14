auditor-model: glm-5.3-flash

# Cold Sync-Audit Re-Judgment (r2) — card t1630 (SPEC-USER-SETTINGS-PROTECT-001), post-transfer scope

verdict: PASS-WITH-DEBT
audited_sha: 6f7ddb13078b5432270df03966058caa5afb02ed

Scope-limited final re-judgment per operator decision **d-20261014T085956Z-bee2** (option ii — the ENTIRE inter-process lock surface F9-F12 transfers to follow-up card **t1730** (after=t1630) as named debts; the card CLOSES in its current state: in-process `recordProtocolMu` serialization + first-generation lockfile) and the leader's landing-path instruction. Prior verdict file `sync-audit.md` (re-entry #2, FAIL) is retained as history and is NOT this card's binding verdict. Audited tree unchanged since re-entry #2: HEAD 6f7ddb1307, working tree clean (re-verified immediately before this judgment).

## GLM-lane model-mapping note (leader-ruled acceptance)

This audit was requested at opus/high per the operator role table (d-20261014T081435Z-64ee) and actually served on **glm-5.3-flash**. Leader ruling (board records **d-20261014T090106Z-30e9**, standing): on GLM lanes the 'opus' spawn request resolving to glm-5.3-flash is **measured reality** (spawn model=opus, verdict line `auditor-model: glm-5.3-flash`; GLM 'opus' alias routes to `llm.glm.models.high` = glm-5.3-flash); the mapping repair is card t1728's, and the **BINDING verdict comes from the codex required gate**. This file's `auditor-model:` line records the serving model verbatim per that ruling.

## Convergence digest — round 3, post-transfer scope (for the orchestrator's check)

- overall_verdict: **fail** (raw fold; scope-classified below per bee2)
- residual_risk_note: `required-backend FAIL: codex`; gate_unmet absent; plan_source: **config**
- per_backend_verdicts: codex / gate **required** / verdict **fail** (sole participant; participant_count 1, disagreement_flag null, fail_open_backends [])
- audit_receipt (round 3, deciding): **rcpt-56c5c4ac1f91207f1d2eaa71** (round 2: rcpt-2f0145b43a32bd96b19103d7; round 1: rcpt-9a7cb95694999d93a622b157; round 0: rcpt-ce0152d51eac3787c2b2a0d7)
- Codex's single finding this round: **[P1] "FIFO lock file disables the time bound and halts global serialization"** at tier_render.go:237 — a FIFO planted at the lock path makes `os.ReadFile` block indefinitely (no writer) while `recordProtocolMu` is held; codex reproduced with a child killed after 7s. Codex also re-ran `moai verify codex-review` → inconclusive (`codex stdout closed before response to id=1`, same child-EOF shape as rounds 1-2 — named non-evidence).
- **Scope classification (this auditor's judgment): OUT of the post-transfer card scope — an instance of the transferred t1730 debt class.** The finding targets `takeoverStaleRecordLock` (tier_render.go:231-244), the exact line range card-review.md:48-49 enumerates as F9 (:236-244) and F10 (:231-244 + :203-210) named debts. It is a hostile-input hardening VARIANT of that class: the production code creates only regular files (`O_CREATE|O_EXCL|O_WRONLY`, stamp-failure remove at :194-201) — no moai path can produce a FIFO; the scenario requires a local actor already owning the project's `.moai/state/` tree (the same trust domain as settings.json itself). The in-process serialization (`recordProtocolMu`) behaves as designed; the blocking read is the first-gen lockfile corner the operator transferred. Recorded as **F13** below, t1730 debt class, with codex's fix direction (non-blocking open + regular-file verification + bounded metadata read).
- Fold rule application: a required FAIL folds to overall fail ONLY for in-scope findings. Under bee2 the lock-corner surface is out of this card's scope; **no remaining in-scope blocker was returned** → the verdict is PASS-WITH-DEBT per the operator's stated rule.
- This auditor is read-only: the orchestrator writes `.moai/state/audit-plan-result.json` from these digest members and runs `moai verify audit-plan --project-root <toplevel> --result-file <that path>`. **This verdict is not final until that check passes.**

## Per-finding final status

- **F1 [P1] — CLOSED.** Board ruling d-20261014T052723Z-3eef (verified via `moai decision read --scope card:t1630`, scope=card:t1630): full-bundle path routed to t1666; SPEC records non-normative REQ-003 scoping (consistent with 10fe/56ff/6892).
- **F2 [P1] — FIXED**, verified under `-race` at the audited HEAD (this audit session, Claim 1 below).
- **F3 [P2] — FIXED**, same evidence.
- **F2' primary — FIXED.** The primary scenario (two live concurrent healthy renders) is arbitrated correctly by the first-gen lock: verified by this auditor's own two-process `-race` run (reexec child = genuine second OS process) + code read (mutex-first ordering, O_EXCL create, bounded retry, EPERM-treated-alive pid probe, windows age-bound).
- **F8 [minor] — FIXED.** `managedAllowRecord` comment states the rollback semantics (tier_render.go:374-383).
- **F9, F10, F11, F12 — → t1730 NAMED DEBTS** per operator decision bee2, recorded at `.moai/reports/t1630/card-review.md:46-51` (verified this round): F9 age-only takeover of live holders; F10 takeover TOCTOU + ownership-free release; F11 BuildInto body read before the lock; F12 USER-scope spans unserialized.
- **F13 [P1, t1730 debt class — NEW this round]** tier_render.go:237 FIFO-planted lockfile blocks the takeover read indefinitely under `recordProtocolMu`. Mechanism real but requires a foreign FIFO at the lock path (production creates regular files only). Scope-classified out per bee2; fix direction carried to t1730.
- **F4 [minor], F5, F6, F7 — optional, carried, do not gate.**

## Deciding evidence (five-section format)

**Claim 1 — the in-scope record protocol (F2/F3/F2'-primary) holds under the race detector at the audited HEAD.**
- Evidence (this audit session, verbatim): `go test -race -count=1 -v -run '^(TestInterProcessRecordLock_TwoProcessRepro|TestConcurrentRendersKeepRecordConsistentWithSettings|TestFailedSettingsWriteRollsRecordBack_UserAddedEntrySurvivesRevocation)$' ./internal/config/toolpolicy/` → `--- PASS: TestInterProcessRecordLock_TwoProcessRepro (13.65s)` (all three subtests RUN), `--- PASS: TestConcurrentRendersKeepRecordConsistentWithSettings (5.95s)`, `--- PASS: TestFailedSettingsWriteRollsRecordBack_UserAddedEntrySurvivesRevocation (0.02s)`, `PASS`, `ok ... 20.928s`, exit 0, no DATA RACE.
- Baseline-attribution: this run, HEAD 6f7ddb13078b5432270df03966058caa5afb02ed, worktree `.moai/worktrees/t1630`.
- Gaps: F13-class interleavings not covered by these tests (t1730); the RED run at 486fc3747d cannot be re-witnessed from the commit graph (test+fix share 9b9fc68069 — §2.3 ordering; recorded as a Gap, GREEN re-observed).
- Residual-risk: the t1730 debt class (F9-F13).

**Claim 2 — Craft gates hold on the card's tree.**
- Evidence (this audit session): full `-race -count=1` → `ok internal/config/toolpolicy 18.781s`, `ok internal/core/project 12.728s`; darwin build exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go test -count=1 -cover` → `coverage: 88.0%` / `coverage: 89.0%` (≥ 85%); `gofmt -l` empty; `go vet` clean; `golangci-lint run` → `0 issues.`
- Baseline-attribution: this run, HEAD 6f7ddb1307.
- Gaps: AC selectors not individually re-run this session (delta scope; full package suites ran twice, once under `-race`).
- Residual-risk: none in-scope beyond the carried items.

**Claim 3 — the transfer premise is on disk and board.**
- Evidence: `card-review.md:46-51` carries the t1730 debt record naming F9-F12 with file:line; `moai decision read --scope card:t1630` returns the bee2-scoped ruling and d-20261014T090106Z-30e9 (model mapping). HEAD re-read at 6f7ddb1307 immediately before judgment — unchanged.
- Gaps: none for this claim.
- Residual-risk: none.

## Carried Gaps (not failures — owner named)

- **F9-F13 inter-process lock corner → t1730** (named debts; fix directions: process-exit-released OS locking, ownership-bound release, body read inside the span, USER-span serialization, non-blocking open + regular-file check + bounded metadata read for F13).
- init byte-diff ≠ 0 (`ensureGlobalSettingsEnv`) → **t1703**; `TestRuleDateProvenance` date leak → **t1704**; staging-window persistence-direction residual → **t1703**.
- **AC-013 gate red** pending t1619 + t1591 landing — merge-ready HOLD, leader-owned; not re-measured this round (delta scope), not failed on.
- Auditor-model mapping (opus→glm-5.3-flash on GLM lanes) → **t1728** (leader-ruled measured reality, see note above).
- build_lag: installed moai binary from ec193e477 (ancestor); no moai-CLI tree measurement is cited as evidence in this verdict.

## Iteration history (this card's sync-audit chain)

1. Round 0 (25df1da08c): FAIL, folded codex (rcpt-ce0152d51eac3787c2b2a0d7); F1-F3 blocking.
2. Re-entry #1 (486fc3747d): FAIL — F2/F3 fixed; F1 ruling pending; F2' open (rcpt-9a7cb95694999d93a622b157).
3. Re-entry #2 (6f7ddb1307): FAIL — F1 closed, F8 fixed, F2'-primary verified; codex FAIL with F9-F12 (rcpt-2f0145b43a32bd96b19103d7). Iteration ceiling reached; operator elected transfer (bee2).
4. **r2 re-judgment (6f7ddb1307, this file): PASS-WITH-DEBT** — post-transfer scope, codex's sole finding (F13) scope-classified into the t1730 debt class; receipt rcpt-56c5c4ac1f91207f1d2eaa71.
