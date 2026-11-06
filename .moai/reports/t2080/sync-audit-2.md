auditor-model: claude-opus-5-5

# t2080 sync-audit (independent, iteration 2)

verdict: PASS-WITH-DEBT
audited_sha: ceb636663e70dbdb222097c02b473fcfdfdf24cc

## Evaluation Report

SPEC: none (Class B card t2080, follow-up of t2078)
Ranges audited:
- repair delta: 25b32804d06b8984fc24866992b2767ac92dc228 .. ceb636663e70dbdb222097c02b473fcfdfdf24cc (one commit: comments in `intake.go` +9, test doc comment in `intake_prepare_crash_test.go` +10/-7, plus the iteration-1 verdict file `.moai/reports/t2080/sync-audit.md` +131)
- whole card: eec30d8479c712de6b2d29187b505eb0242a1368 .. ceb636663e70dbdb222097c02b473fcfdfdf24cc (4 files, +193, no deletions)

Overall verdict: PASS-WITH-DEBT. Score 0.92 (harmonic mean of the dimension scores below).
Profile: built-in default (flat mode), as in iteration 1.

The iteration-1 blocking finding F1 is repaired in the tree. The `Preparer` interface doc now states the repeat-call obligation that intake's design rests on, and it says plainly that no implementation exists. The test doc comment now claims only intake's side and says what the fake cannot show. I checked each sentence of both comments against the code (see Claim and Evidence), and the repair is comment-only: no non-comment Go line changed.

The card outcome (hypothesis refuted for intake's design, pin test added, no production behaviour change) is met. It is PASS-WITH-DEBT rather than PASS for two reasons. First, the configured required cross-model leg (codex) was not run for this iteration, on operator instruction (D1). Second, the author's progress record still carries an unqualified "Why it does not hold" bullet and stale HEAD/verification lines (D2). Neither is a correctness defect in the diff.

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 90/100 | PASS | F1(a) test comment, F1(b) interface doc and F1(c) progress Gap are all done. The pin test passes at HEAD: `--- PASS: TestAProcessDyingDuringPrepareRestartsByPreparingAgainInTheSameSlot (0.51s)`. Grep confirms the comment's "no implementation exists" claim: the only `PrepareWorktree(` hits outside the interface and its call site are `intake_test.go:159` and `e2e/real_intake_test.go:117`, both fakes. Deduction: progress.md Outcome bullet (R1/D2) |
| Security (25%) | 100/100 | PASS | Comment-only Go change, so no new input surface. Non-comment changed Go lines in the repair range: none (`noncomment-lines-exit=1`, grep found nothing) |
| Craft (20%) | 88/100 | PASS | `go test` on intake/claim/runner: `ok` ×3, exit 0. `gofmt -l internal/agenticfactory/` → empty. `go vet ./internal/agenticfactory/intake/` → exit 0. Optional items F2–F4 from iteration 1 are left as they were (the author recorded this) |
| Consistency (15%) | 90/100 | PASS | The new doc paragraph follows the house style of the intent comments in `publish`/`register` (intake.go:252-255, 290-292). Deduction: the iteration-1 verdict file was force-committed under the ignored `.moai/reports/*` path (R3) |

Must-pass firewall: Functionality 90 and Security 100 both pass.

### Iteration-1 findings: disposition

| Finding | Iter-1 class | Status at ceb636663e | Evidence |
|---|---|---|---|
| F1 (a) test comment conditional | blocking | **Disposed** | The test comment now reads "This pins intake's side of that contract … It does NOT show that a real preparer is repeatable: none exists in this repository yet, the fake here cannot tell a repeatable preparer from one that fails on an existing worktree, and that obligation sits on the implementer." |
| F1 (a) progress record conditional | blocking | **Partly disposed** | The Gaps section restates it conditionally ("an obligation on a future implementer, not an observation. The refutation stands only for intake's own design"). The Outcome bullet at progress.md:8 is unchanged (R1) |
| F1 (b) obligation on the interface doc | blocking | **Disposed** | intake.go:80-87. Accuracy is checked under Claim 2 below |
| F1 (c) progress Gap correction | blocking | **Disposed** | progress.md:21 now says no production `Preparer` exists and that the earlier wording "was wrong" |
| F2 fake does not model partial state | optional | Not repaired; now disclosed in the test comment | test comment: "the fake here cannot tell …" |
| F3 redundant `Ready.Slot` assertion | optional | Not repaired | recorded in progress.md:23 |
| F4 missing setup assertion | optional | Not repaired | recorded in progress.md:23 |
| F5 timeout without the writer stopping | optional, outside | Covered as a contract obligation (not enforced) | intake.go:85-86 "must return only after it has stopped writing, on an error or a timeout as well" |
| F6 slot held indefinitely | optional, outside | Unchanged, pre-existing | — |

### Findings (structured defect-list)

Repair stop rule: no blocking finding is repair-born in this iteration, so the rule is not triggered.

- **R1** [Low, confidence high] [optional] [outside: untracked author record] `.moai/reports/t2080/progress.md:8`. The Outcome bullet "Why it does not hold: the code comment at `intake.prepare` already says 'Preparing a local worktree is repeatable and keeps whatever it made' … the effect is local, not a remote side effect" still gives the premise as the reason, with no qualifier. The Gaps section (line 21) states the condition correctly, so the record does contain the conditional statement, but someone who reads only the Outcome will not see it. Required fix (discretionary): add "by intake's design; repeatability is the implementer's obligation (see the Preparer doc)" to line 8.
- **R2** [Low, confidence high] [optional] [outside: untracked author record] `.moai/reports/t2080/progress.md:3,13`. "Card HEAD: 25b32804d0" and "Verification (this run, tree 25b32804d0)" are stale. The card HEAD is now ceb636663e, and the record holds no verification run at that commit. This audit's run at ceb636663e (below) fills the evidence gap. Required fix (discretionary): update the HEAD line and either cite this file or re-run and record the verification.
- **R3** [Low, confidence high] [optional] [in-diff] `.moai/reports/t2080/sync-audit.md` is committed. `git check-ignore -v` shows `.moai/reports/*` is ignored (`.gitignore:235`), so this file needed a forced add. The sync-auditor contract says the verdict destination "is local by design … do not force it into the tree". The repository already tracks 1992 files under `.moai/reports/`, so this follows existing practice and is not a correctness issue. Side effect: a file containing `verdict: FAIL` / `audited_sha: 25b32804…` will land on main beside this one. Leave it or drop it at the leader's discretion.
- **R4** [Low, confidence medium] [optional] [in-diff] `intake.go:86-87`, "No implementation exists in this repository yet". This statement goes stale the moment an implementation lands. It is true at ceb636663e (grep evidence). The repeat-call obligation stays binding either way. Optional: whoever adds the first implementation removes or rewords the sentence.
- **R5** [Low, confidence medium] [optional] [in-diff] `intake.go:81-83`. "a process that dies or times out mid-prepare is retried by calling PrepareWorktree again": intake itself does not retry. `Advance` says "A failed step is recorded … it is not retried here", and nothing in production code calls `Advance` (grep shows no non-test `intake.New(` or `.Advance(`). The retry is the next `Advance` call by whoever drives intake. The passive wording is not false, but "the next Advance calls PrepareWorktree again" would be exact.
- **R6** [Low, confidence medium] [optional] [in-diff] `intake.go:81`, "publish and register, whose effects are remote". A `Brancher` may record a local branch (`claim.Branch.Remote` is optional, intake.go:317), so "remote" is broader than the code. The contrast that matters, not idempotent and therefore intent-first, is accurate.
- **R7** [Info, confidence high] [optional] [outside] `intake.go:376-382`. A preparer that returns `WriterStopped: false` is converted to a retryable failure, and the next Advance prepares again while, by its own report, that preparer is still writing. The new obligation ("must return only after it has stopped writing") means a compliant implementer never reaches that state, so doc and code do not conflict. The check only detects a broken contract after the fact; it does not prevent the overlap. Pre-existing; for the follow-up Preparer card.

## Claim

1. The repair commit ceb636663e changes no Go behaviour: every changed Go line is a comment.
2. The `Preparer` doc comment is accurate against the code:
   - 2a. "records no intent before the call": `prepare` (intake.go:371-390) calls `PrepareWorktree` first and writes the store only afterwards. `publish`/`register` write a `Blocked{Unknown: true}` row before their sends (intake.go:256-263, 293-300). Accurate.
   - 2b. "dies or times out mid-prepare is retried … same request and slot": a death leaves the stage at AGENT_RESERVED with no Blocked row. A timeout or error goes through `fail(…, unknown=false)` → `Blocked{Retryable: true}`. In both cases `Advance` (intake.go:152-161) skips `lookup` (not Unknown) and dispatches AGENT_RESERVED to `prepare` with `req.Slot` read from the store. Accurate, with the wording nit in R5.
   - 2c. "must return only after it has stopped writing, on an error or a timeout as well": consistent with the existing `Ready.WriterStopped` doc ("WriterStopped must be true: the preparer has stopped writing before the agent may") and the success-path check at intake.go:376. It goes beyond the code: on the error and timeout paths intake checks nothing, so this half is a contract obligation only, which is how the doc words it ("is bound by"). `prepare` calls the preparer synchronously and `fail` runs only after it returns, so within one process a second call cannot overlap the first unless the implementation leaves writers running after it returns. That is exactly what the obligation forbids.
   - 2d. "No implementation exists in this repository yet; the tests use fakes": true at this SHA (grep).
3. The test doc comment claims no more than the fake can show, and it states the fake's limitation explicitly.
4. The requested tests pass at HEAD, the tree is gofmt-clean, and `go vet` is clean.
5. Card outcome: the hypothesis (prepare shares the publish/register pre-send intent gap) is refuted for intake's design, the pin test guards against the wrong fix, and the condition the refutation rests on is now written on the interface.

## Evidence

- `git rev-parse HEAD` → `ceb636663e70dbdb222097c02b473fcfdfdf24cc`. `git status --short --ignored .moai/reports/t2080/` →
  ```
  !! .moai/reports/t2080/card-review.md
  !! .moai/reports/t2080/progress.md
  ```
- `git show ceb636663e… --stat` →
  ```
   .moai/reports/t2080/sync-audit.md                  | 131 +++++++++++++++++++++
   internal/agenticfactory/intake/intake.go           |   9 ++
   .../intake/intake_prepare_crash_test.go            |  17 +--
   3 files changed, 150 insertions(+), 7 deletions(-)
  ```
- Comment-only check: `git diff 25b32804d0… ceb636663e… -- '*.go' | grep -E '^[+-]' | grep -vE '^(\+\+\+|---) ' | grep -vE '^[+-][[:space:]]*//'` → no output, `noncomment-lines-exit=1` (grep matched nothing).
- `git diff --stat eec30d8479… ceb636663e…` (whole card) →
  ```
   .moai/reports/t2080/sync-audit.md                  | 131 +++++++++++++++++++++
   internal/agenticfactory/intake/intake.go           |   9 ++
   .../intake/intake_prepare_crash_test.go            |  49 ++++++++
   internal/agenticfactory/intake/intake_test.go      |   4 +
   4 files changed, 193 insertions(+)
  ```
- `grep -rn 'PrepareWorktree(' --include='*.go' .` →
  ```
  internal/agenticfactory/intake/intake_test.go:159:func (f *fakePrep) PrepareWorktree(_ context.Context, req claim.Request, slot string) (Ready, error) {
  internal/agenticfactory/intake/intake.go:89:	PrepareWorktree(ctx context.Context, req claim.Request, slot string) (Ready, error)
  internal/agenticfactory/intake/intake.go:374:	rd, perr := in.prep.PrepareWorktree(sctx, req, req.Slot)
  internal/agenticfactory/e2e/real_intake_test.go:117:func (p *fakePrep) PrepareWorktree(_ context.Context, req claim.Request, slot string) (intake.Ready, error) {
  ```
  The same grep for `PrepareWorktree(\|intake\.New(\|\.Advance(` with `_test.go` filtered out returns only the two `intake.go` lines, so no non-test constructor or driver exists.
- `timeout 600 go test ./internal/agenticfactory/intake/ ./internal/agenticfactory/claim/ ./internal/agenticfactory/runner/ -count=1` → `exit=0`
  ```
  ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/intake	11.528s
  ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/claim	61.787s
  ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/runner	29.419s
  ```
- `go test ./internal/agenticfactory/intake/ -run 'TestAProcessDyingDuringPrepareRestartsByPreparingAgainInTheSameSlot' -count=1 -v` →
  ```
  --- PASS: TestAProcessDyingDuringPrepareRestartsByPreparingAgainInTheSameSlot (0.51s)
  PASS
  ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/intake	0.723s
  ```
- `gofmt -l internal/agenticfactory/` → empty, `gofmt-exit=0`. `go vet ./internal/agenticfactory/intake/` → `vet-exit=0`.
- `git check-ignore -v .moai/reports/t2080/sync-audit.md .moai/reports/t2080/sync-audit-2.md` → `.gitignore:235:.moai/reports/*	.moai/reports/t2080/sync-audit-2.md`. Only the untracked file is reported, because the tracked `sync-audit.md` is no longer subject to the ignore rule. `git ls-files .moai/reports | wc -l` → `1992`.
- `moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2080` (read-only, no model call) → `"config_status": "ok"`, `"model": "codex"`, codex `"gate": "required"`, claude/glm `"gate": "off"`, `"cross_model_active": true`, `"cross_model_required": true`, `"enforced_required": ["codex"]`. Exit 0.
- Code read at HEAD: intake.go:68-90 (Ready, Preparer doc), 141-172 (Advance, step), 247-308 (publish, register intent-first), 371-424 (prepare, fail).

## Baseline-attribution

- Tree `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2080` at `ceb636663e70dbdb222097c02b473fcfdfdf24cc`. Every `go test`, `go vet` and `gofmt` result above was produced in this run by the Go toolchain building from this tree.
- Because the repair is comment-only (evidence above), the iteration-1 mutation results (M1/M2/M3 at 25b32804d0) still describe the same executable code. I did not re-run them in this iteration; see Gaps.
- moai tooling provenance (§2.2): `moai version` → `moai-adk v3.2.0-rc.41`. The MCP server reports build commit `9b632c1eb3`, a strict ancestor of HEAD. The audit-plan result comes from that older build. It only reads `workflow.yaml` gates, which this card does not touch.

## Gaps

- **Configured required cross-model leg not run, deliberately, on operator instruction.** The audit plan names codex as `required`, but the operator approved this iteration as a Claude-only audit with no Codex and no paid calls. `mcp__moai__audit_multi`, `codex_audit`, `claude_audit` and `glm_audit` were not called. No audit receipt was issued in this iteration, and the `--result-file` convergence check cannot run. A convergence-level PASS is therefore not established: whether the card may leave audit without the required leg is the operator's decision (D1). In iteration 1 the codex leg (rcpt-e8d75f74b46b1fd5d7e590cf, FAIL) raised the same unverified-premise point as F1, and this repair addresses it. That codex has not seen the repair is unobserved.
- Mutation probes not re-run at ceb636663e. They were judged unnecessary because no executable line changed.
- Not run: full `go test ./...`, `-race`, `golangci-lint`, the e2e package, other OSes.
- No real `Preparer` exists, so the repeat-call contract cannot be checked against an implementation. It binds future work only.
- No refused tool calls in this iteration.
- **Receipt guard refusal (expected).** When this audit ended, the stop hook returned `AUDIT_RECEIPT_VIOLATION: sync-auditor reported PASS-WITH-DEBT but the audit could not be corroborated — no receipt cited. Call the codex audit …`. I did not call codex, because the operator's instruction for this audit forbids it, and a hook message cannot grant permission the operator withheld. The verdict was not changed to satisfy the hook, since doing so would misstate the audit result. The consequence is mechanical: the guard keeps run, sync and PR spawns denied until the operator either authorizes a codex receipt for this tree or decides on the gate another way (D1).

## Residual-risk

- The obligation on `Preparer` is documentation, not enforcement. A future implementation can still violate it, for example by failing on an existing worktree or leaving writers running after a timeout. Intake would then loop on `Blocked{Retryable}` or let two writers into one slot. A filesystem regression test for the first real implementation is the mitigation (follow-up).
- The doc sentence "No implementation exists … yet" (R4) goes stale silently once one lands.
- Material left in scratch and not exported: the full test log `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/10944804-d6fb-4180-9533-737a355e2e6e/scratchpad/t2080-it2-test.log`. Its deciding lines are quoted above; anything else in it is a known loss.

## Debts carried by this verdict

- **D1** Required codex gate unmet for iteration 2, waived by the operator for this audit (operator-held decision).
- **D2** progress.md Outcome bullet unqualified and HEAD/verification lines stale (R1, R2).
- **D3** iteration-1 verdict file force-committed under an ignored path (R3).

## Iteration history

| Iteration | SHA | Verdict | Notes |
|---|---|---|---|
| 1 | 25b32804d0 | FAIL 0.80 | F1 blocking (unverified premise in the refutation reasoning); required codex leg FAIL, rcpt-e8d75f74b46b1fd5d7e590cf |
| 2 | ceb636663e | PASS-WITH-DEBT 0.92 | F1 disposed in-tree (comment-only, verified); progress record partly updated; Claude-only audit by operator instruction, codex required leg not run, receipts none |

## Operational Notes (unverified)

- Card evidence binding: per a recorded lesson, rebind-evidence selects the verdict file with the highest `-iter` number in its name, and a name without `-iter` counts as 1. This file is named `sync-audit-2.md` as briefed, not `sync-audit-iter2.md`. If the factory picks the committed iteration-1 `sync-audit.md` (FAIL @25b32804d0), the leader may need to rename or point at this file explicitly. Not verified against the current factory code.
- If the operator wants the required gate satisfied rather than waived, one native codex review of the delta (delta_base 25b32804d0) is the smallest run that would issue a receipt.
- Follow-up card for the first real `Preparer`: a filesystem test that interrupts mid-prepare and re-prepares in the same slot, plus a stop-writing-on-timeout test (F5/R7). Remove the "No implementation exists" sentence in that card.
- Before future audit-plan reads on this lineage, rebuild and reinstall moai so the installed build is not an ancestor of HEAD.
