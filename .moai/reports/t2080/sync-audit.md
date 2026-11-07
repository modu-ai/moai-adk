auditor-model: claude-opus-5-5

# t2080 sync-audit (independent, iteration 1)

verdict: FAIL
audited_sha: 25b32804d06b8984fc24866992b2767ac92dc228

## Evaluation Report

SPEC: none (Class B card t2080, follow-up of t2078)
Range audited: eec30d8479c712de6b2d29187b505eb0242a1368 .. 25b32804d06b8984fc24866992b2767ac92dc228 (one commit, test-only: `intake_prepare_crash_test.go` +46, `intake_test.go` +4)
Overall Verdict: FAIL. Score 0.80 (harmonic mean of the dimension scores below)
Profile: `.moai/config/evaluator-profiles/default.md` (harness.yaml `default_profile: "default"`), flat mode

The verdict is FAIL for two reasons. First, the card's acceptance basis ("if the hypothesis is refuted, add a pin test **and the reasoning**") is only partly met. The pin test is a real pin of intake's re-call behaviour. The reasoning, however, presents an unverified premise as fact: that PrepareWorktree is repeatable and keeps what it made. No production implementation of `intake.Preparer` exists in the repository, so that premise cannot have been observed. It is a contract obligation on a future implementer, and that obligation is written nowhere on the interface. Second, the configured required cross-model gate (codex) returned `fail` on the same point.

The hypothesis itself (intake.prepare shares the publish/register pre-send intent gap) is correctly NOT reproduced as a defect of the intake design. A restart re-prepares in the same exclusively held slot with no operator step, and nothing remote is registered twice. The production code needs no intent marker. The defect is in how the refutation is stated and where its condition is recorded.

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 60/100 | FAIL | The pin passes and catches the wrong fix (mutant M1 FAIL: `step AGENT_RESERVED has no lookup`). The card's "reasoning" deliverable asserts an unverified premise (F1). `grep -rn 'PrepareWorktree(' --include='*.go' .` → only `intake.go:80` (interface), `intake.go:365` (call), `intake_test.go:159` (fake), `e2e/real_intake_test.go:117` (fake) |
| Security (25%) | 100/100 | PASS | Test-only diff. `git show 25b3… \| grep -inE 'token\|secret\|password\|api[_-]?key'` → no output. No input surface, no production change |
| Craft (20%) | 85/100 | PASS | `go test -cover ./internal/agenticfactory/intake/` → `coverage: 91.8% of statements`; `gofmt -l internal/agenticfactory/` → empty; `go vet ./internal/agenticfactory/intake/` → exit 0; race run of prepare/crash tests → `ok`. Minor test-precision items F3–F5 |
| Consistency (15%) | 90/100 | PASS | Mirrors the sibling crash tests (`intake_crash_test.go`, `intake_register_crash_test.go`): same `crash` flag on the fake, same panic/recover restart idiom. The sibling's setup assertion is missing (F4) |

Must-pass firewall: Functionality (must-pass) FAILS at 60 < its "all acceptance criteria met" threshold, so the overall verdict is FAIL whatever the other scores.

### Cross-model gate (audit plan)

`moai verify audit-plan --project-root <toplevel>` → `config_status: ok`, `model: codex`, `cross_model_active: true`, `cross_model_required: true`, `enforced_required: ["codex"]`, claude gate `off`, glm gate `off`.

`mcp__moai__audit_multi` (project_root = this tree, target baseBranch, delta_base eec30d8479…) digest:

- `overall_verdict`: `fail`
- `gate_unmet`: not present in the returned object
- `plan_source`: `config`
- `per_backend_verdicts`: `{backend: codex, gate: required, verdict: fail}`
- `audit_receipt`: `rcpt-e8d75f74b46b1fd5d7e590cf`
- codex finding (P2, `intake_prepare_crash_test.go:12`), paraphrased: `fakePrep` records only call count and slot, then panics, so it leaves no worktree or partial artifact. The test does not check that existing files, branch or worktree survive a re-call; it pins only "re-call in the same slot". The repository's only `PrepareWorktree` implementations are two test fakes, and the `Preparer` interface states no preservation obligation, so slot exclusivity alone does not prove repeat safety. The conclusion that intent-first is unnecessary therefore lacks grounds.

This verdict is not final until the orchestrator's `moai verify audit-plan --result-file` convergence check runs over this digest. The digest is a required-backend FAIL in any case.

### Findings (structured defect-list)

- **F1** [Medium, confidence high] [blocking] [in-diff] `internal/agenticfactory/intake/intake_prepare_crash_test.go:9-16` and the commit message: these state as fact that "PrepareWorktree builds a LOCAL worktree … and preparing again keeps what the first attempt made". No production `intake.Preparer` exists (grep above), so the premise is unobserved (verification-claim-integrity §1.1 surface 4). The only statement of the obligation is an inline comment at `intake.go:372`. The interface doc at `intake.go:77-81` says nothing about repeat calls. The author's `progress.md` Gap ("the real preparer's behavior on a second call was not read") implies a real preparer exists; none does. The codex required leg independently raised the same point.
  **Required fix:** (a) In the test's doc comment and the progress record, restate the refutation as conditional: "intake treats prepare as repeatable by design. No production Preparer exists yet. Repeatability is an obligation on any implementer. This test pins intake's side (re-call in the same slot, no operator step), not the implementer's." (b) Write the obligation onto the `Preparer` interface doc comment (`intake.go:77-81`; a comment-only change): a second call for the same request and slot must succeed against whatever an interrupted earlier call left in that slot, and must touch nothing outside the slot. A future implementer is then bound by the contract the refutation rests on. (c) Correct the progress.md Gap so it says no real preparer exists, instead of saying its behaviour was "not read".

- **F2** [Low, confidence high] [optional] [in-diff] `intake_test.go:156-164`: the fake's `crash` path panics without modelling the partial state the comment claims ("the worktree IS prepared"). The test therefore cannot tell a repeatable preparer from one that fails on an existing worktree. Once F1(a) narrows the claim, this is inherent to a fake and acceptable. Optional: record a "made" flag on the first call and have the second call observe it, which documents the intended contract in the fake.

- **F3** [Low, confidence medium] [optional] [in-diff] `intake_prepare_crash_test.go:35`: `got.Ready.Slot != reserved.Slot` cannot fail, because `prepare` copies `req.Slot` into `Ready` (`intake.go:377`) whichever slot was prepared. Mutant M2 (prepare a different slot) was caught only by the `r.pr.slots` loop at :41-45. The assertion is redundant, not wrong. Optional: also assert `r.in.Store.SlotOwner(reserved.Slot) == "crash-prep"` after the restart.

- **F4** [Low, confidence high] [optional] [in-diff] `intake_prepare_crash_test.go:21-26`: there is no setup assertion that the crash attempt actually called the preparer (`calls == 1` before the restart), unlike `intake_register_crash_test.go:26-30`. It is covered implicitly: without the panic, the restart would return `ErrTransitionInvalid` (`waits for the reserved agent to acknowledge`). Optional, for symmetry.

- **F5** [Medium, confidence medium] [optional] [outside] `intake.go:362-374`: a step timeout (`context.DeadlineExceeded`) is recorded as retryable, and the next Advance re-prepares. A real preparer that does not stop on context cancellation could leave two writers in one slot. "Repeatable" as a contract therefore also needs: "returns only after it has stopped writing, on error or timeout too". This is pre-existing and outside the diff; it is listed for the operator and does not affect the verdict.

- **F6** [Low, confidence high] [optional] [outside] `claim/request.go:432-443` / `ReleaseWorktree`: a request left at AGENT_RESERVED keeps its slot indefinitely, because there is no expiry and release requires `Started`. This does not block a restart (`AcquireSlot` is idempotent for the same owner, `request.go:653-655`) and double-registers nothing. Informational only; pre-existing.

Repair stop rule: this is iteration 1 of this card's audit, so no finding is repair-born.

## Claim

1. The diff is test-only and changes no production code.
2. After a crash mid-prepare, a restart prepares again in the slot the request holds, needs no operator step, and registers nothing twice. Verified for the intake design, with a fake preparer.
3. The pin test is a real pin: it fails for an intent-first Unknown marker (the "fix" it guards against) and for a wrong-slot prepare, and passes for a harmless retryable marker.
4. The premise "PrepareWorktree is repeatable and keeps what it made" is NOT verified, because no production implementation exists (F1).
5. The tests requested in the brief pass, and the tree is gofmt-clean.

## Evidence

- `git -C <tree> rev-parse HEAD` → `25b32804d06b8984fc24866992b2767ac92dc228`; `git -C <tree> status --short` → empty (clean).
- `git show 25b32804d0 --stat` → `.../intake/intake_prepare_crash_test.go | 46 ++++`, `internal/agenticfactory/intake/intake_test.go | 4 ++`, `2 files changed, 50 insertions(+)`.
- `grep -rn 'PrepareWorktree(' --include='*.go' .` →
  ```
  internal/agenticfactory/intake/intake.go:80:	PrepareWorktree(ctx context.Context, req claim.Request, slot string) (Ready, error)
  internal/agenticfactory/intake/intake.go:365:	rd, perr := in.prep.PrepareWorktree(sctx, req, req.Slot)
  internal/agenticfactory/intake/intake_test.go:159:func (f *fakePrep) PrepareWorktree(_ context.Context, req claim.Request, slot string) (Ready, error) {
  internal/agenticfactory/e2e/real_intake_test.go:117:func (p *fakePrep) PrepareWorktree(_ context.Context, req claim.Request, slot string) (intake.Ready, error) {
  ```
  `grep -rn 'intake\.New(' --include='*.go' .` → only `internal/agenticfactory/e2e/real_intake_test.go:148`. Nothing in production code constructs an Intake.
- `timeout 600 go test ./internal/agenticfactory/intake/ ./internal/agenticfactory/claim/ ./internal/agenticfactory/runner/ -count=1` → `exit=0`
  ```
  ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/intake	14.880s
  ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/claim	70.335s
  ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/runner	34.713s
  ```
- `gofmt -l internal/agenticfactory/` → empty (`gofmt:[]`). `go vet ./internal/agenticfactory/intake/` → exit 0.
- `go test -race ./internal/agenticfactory/intake/ -run 'Prepar|Crash|Dying' -v` → `--- PASS: TestAProcessDyingDuringPrepareRestartsByPreparingAgainInTheSameSlot (0.40s)`, plus the publish/register crash tests and the two Preparer tests PASS; `ok … 3.760s`.
- `go test -cover ./internal/agenticfactory/intake/` → `coverage: 91.8% of statements`.
- Mutation probes via `go test -overlay` (mutant copies of `intake.go` in session scratch; no tracked file touched), `-run TestAProcessDyingDuringPrepareRestartsByPreparingAgainInTheSameSlot`:
  - M1, an intent-first `Blocked{Unknown: true}` written before PrepareWorktree →
    `--- FAIL: … intake_prepare_crash_test.go:30: restart after a crash mid-prepare needs no operator step: agenticfactory: an effect may have happened and must be looked up, not repeated: step AGENT_RESERVED has no lookup`
  - M2, PrepareWorktree called with `in.cfg.Slots[len-1]` instead of `req.Slot` →
    `--- FAIL: … intake_prepare_crash_test.go:43: a prepare touched slot "wt-2", the request holds "wt-1"` (×2)
  - M3, a retryable, non-unknown in-flight marker → `ok … 0.719s` (harmless; correctly allowed)
- Code read: `intake.go:362-381` (`prepare`: no pre-send marker; failure goes to `fail(…, unknown=false)` and becomes retryable; success is `UpdateRequest` to WTPrepared). `claim/request.go:500-556` (`ReserveRequest`: the slot is acquired under the requests lock; a slot the request already holds is reused). `claim/request.go:622-659` (`AcquireSlot`: the same owner is idempotent, a foreign owner gets `ErrHeld`). `claim/request.go:678-693` (`ReleaseSlot`: owner only). `claim/claim.go:324-338` and `request.go:156-161` (the Store is file-backed with a lock file, so the in-process panic/recover "restart" reads the same durable state a new process would).
- `moai verify audit-plan --project-root <tree>` and `mcp__moai__audit_multi`: see the § Cross-model gate digest above; receipt `rcpt-e8d75f74b46b1fd5d7e590cf`.

## Baseline-attribution

- Tree: `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2080` at `25b32804d06b8984fc24866992b2767ac92dc228`, clean. Every `go test`, `go vet` and `gofmt` result above was produced in this run by the Go toolchain building from this tree.
- `intake.go` is identical at base `eec30d8479` and head; the diff stat lists only test files. The mutation baseline is therefore the head file itself.
- moai tooling provenance (§2.2): the installed `moai` (`/Users/goos/go/bin/moai`, v3.2.0-rc.41) and the MCP server were built from `9b632c1eb3`, a strict ancestor of the tree HEAD, as `audit_multi` itself reported (`build_lag`). The audit-plan and audit_multi results come from that older build. The commits after it (t2072/t2078/t2080) touch `internal/agenticfactory`, but I did not verify whether they change audit machinery (see Gaps).

## Gaps

- **Refused commands (§3.1):** the worktree-isolation guard refused (1) a compound `git rev-parse && git status && go test` line, and (2) a `for` loop running `go test -overlay` with a shell variable. Both were re-run as plain separate commands (`git -C <tree> …`, one `go test -overlay` per mutant), with nothing substituted for them. If the failure-event hook logged these refusals, they appear as `tool_failure:Bash:*` rows in `.moai/lessons-inbox.jsonl`.
- No production `Preparer` exists, so the repeat-call behaviour of a real worktree preparer (for example `git worktree add` on an existing path) was not observed and cannot be in this tree.
- Not run: full `go test ./...`, `golangci-lint`, the e2e package, other OSes.
- Cross-model coverage: the claude and glm legs are `off` by config. Codex's delta scope (`delta_base`) was given in the prompt only and not mechanically bounded (per the tool's residual note).
- The moai build lag (9b632c1eb3 vs HEAD) was not rebuilt away. Audit-plan semantics are assumed unchanged by the three later agenticfactory commits.

## Residual-risk

- When a real Preparer lands, it may not be repeatable (it may fail on an existing worktree or branch, or leave a half-made worktree that the second call rejects). The intake design would then turn every crash or timeout mid-prepare into a retry loop of `Blocked{Retryable}` that never clears, recoverable only by hand. Nothing in the repository currently binds that implementer to the contract (F1b).
- Timeout without the writer stopping (F5): two prepares writing in one slot.
- The in-process panic/recover simulates a crash on durable file state but not orphaned child processes a real preparer might spawn.
- Scratch-only material not exported here: the full test log (`…/scratchpad/t2080-test.log`) and the mutant files (`…/scratchpad/mut/m{1,2,3}.go`). Their deciding lines are quoted above; anything else in them is a known loss.

## Iteration history

| Iteration | SHA | Verdict | Notes |
|---|---|---|---|
| 1 | 25b32804d0 | FAIL 0.80 | F1 blocking (unverified premise in the refutation reasoning; required codex leg FAIL, rcpt-e8d75f74b46b1fd5d7e590cf) |

## Operational Notes (unverified)

- The smallest repair that would likely clear F1 is comment-only: narrow the test doc comment, add a repeatability and stop-writing obligation to the `Preparer` interface doc, and correct progress.md. That keeps the card's "no production behaviour change" outcome. A confirming re-audit can be scoped to that delta.
- Consider a follow-up card for the real Preparer: it should carry a filesystem regression test that interrupts mid-prepare and re-prepares in the same slot (codex's recommendation), plus the F5 stop-writing-on-timeout obligation.
- The installed moai binary lags HEAD (`make build && make install` per the tool note) before future audit-plan reads on this lineage.
