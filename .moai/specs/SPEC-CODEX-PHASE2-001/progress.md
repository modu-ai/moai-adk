# Progress — SPEC-CODEX-PHASE2-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- tier: M
- artifacts: spec.md + plan.md + acceptance.md (Tier M set)
- spec version: v0.3.0 (plan-audit iteration 2 revision — D1/D2/D3/D4 closed, S2 taken; see `spec.md` HISTORY)
- run-phase entry gate: M0 must close before M1 starts. M0's probe record lands in §E.2 below, under a `### M0 probe — codex-cli <pinned version>` heading, per `plan.md` §D M0 items 2-3.
- open blockers: none. The two `plan.md` §D M0 design forks (background job execution model; write-mode opt-in surface) were resolved by user decision on 2026-08-10 and are recorded as decisions in `plan.md` §D M0 — (i) in-process goroutine, and the `workflow.codex.task.allow_write` config key. spec.md v0.3.0 carries the propagated consistency amendments plus the plan-audit iteration 2 revision. Plan-audit PASSED at 0.84 against the Tier M threshold 0.80 (iteration 2 of 3; report at `.moai/reports/plan-audit/SPEC-CODEX-PHASE2-001-review-2.md`). The SPEC is clear for run-phase, subject to Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

> **How to resolve the SHAs below (added at sync).** Every commit SHA cited in this section — the per-milestone commits, the correction commits, and the HEADs the live probes were run against — lives on the `feat/SPEC-CODEX-PHASE2-001-run` branch ref, which is still on `origin` and was not deleted. PR #1440 was **squash**-merged, so none of them is an ancestor of `main`; on `main` the whole run phase is the single squash commit `ac37c4aea`. The SHAs are left exactly as measured rather than rewritten to `ac37c4aea`, because each one attributes a specific observation to the specific tree it was taken against — overwriting them would falsify that attribution. To inspect one, resolve it against the branch ref (`git fetch origin feat/SPEC-CODEX-PHASE2-001-run`), not against `main`.

### M0 probe — codex-cli 0.146.1

Observed 2026-08-10 against `codex-cli 0.146.1` (`codex --version`), on the worktree at HEAD `edb519b35`.

**Method.** The probe did not have to guess at the wire protocol. `codex app-server` ships a `generate-json-schema --out <DIR>` subcommand that emits the complete protocol as JSON Schema. All four items below are answered from that generated schema — the authoritative machine-readable protocol definition — rather than from an inferred handshake. Command:

```
codex app-server generate-json-schema --out <dir>
```

It produced 41 files including `ClientRequest.json` (the full client→server request union), `ServerNotification.json`, and `ServerRequest.json`. Method names were extracted with:

```
jq -r '.oneOf[] | .properties.method.enum[0] // .properties.method.const // empty' ClientRequest.json | sort -u
```

**(a) Is there a `turn/interrupt`-equivalent method, and what are its params? — PRESENT.**

`turn/interrupt` is a real method in the client→server union. Its params:

```json
{ "type": "object",
  "required": ["threadId", "turnId"],
  "properties": { "threadId": {"type":"string"}, "turnId": {"type":"string"} } }
```

REQ-CX2-011 therefore does **not** degrade to process-termination-only, and the `plan.md` §D M0 item-4 degrade path for (a) is not exercised.

**Consequence that does require a SPEC amendment.** `turn/interrupt` requires **`turnId`**, not `threadId` alone. `REQ-CX2-003`'s job-record field list carries job id, status, timestamps, `threadId`, pid, requested mode, and a request summary — but **no `turnId`**; the token appears zero times across all four artifacts (`grep -c "turnId\|turn_id"` → 0/0/0/0). Without it the job record cannot supply the second required argument, so `REQ-CX2-011` is unimplementable as currently written.

The value is obtainable. `turn/started` (server→client notification) carries:

```json
{ "required": ["threadId","turn"], "properties": { "threadId": {...}, "turn": {"$ref":"#/definitions/Turn"} } }
```

and `Turn` has required `id` — that `id` is the `turnId` `turn/interrupt` expects. The implementation chain is therefore: `turn/start` → read `turn/started` → persist `turn.id` into the job record → `turn/interrupt {threadId, turnId}`.

**(b) Is a second `turn/start` on an existing `threadId` accepted after a turn completes? — STRUCTURALLY SUPPORTED.**

`TurnStartParams` lists `threadId` as **required**, so addressing an existing thread is the method's normal shape rather than a special case; a turn is not modelled as owning its thread. `thread/resume` also exists as a distinct method, alongside `thread/fork`, `thread/rollback`, and `turn/steer`. This is a schema-level observation, not an executed two-turn session — see Gaps below.

**(c) Does the turn request carry a model/effort field, and under what name? — PRESENT, as two separate fields.**

`TurnStartParams` carries both:

- `model` — `["string","null"]`
- `effort` — `ReasoningEffort | null`, described verbatim as *"Override the reasoning effort for this turn and subsequent turns."*

REQ-CX2-002's SSOT wiring therefore has a real destination, and G3 (the resolved model being discarded by `buildCodexReviewParams`) is a genuine live defect rather than a hypothetical one.

**(d) Is a write/sandbox mode expressible per-turn? — PRESENT.**

`TurnStartParams` carries `sandboxPolicy` (`SandboxPolicy | null`) and `approvalPolicy`. `SandboxPolicy` is a 4-variant union:

```
readOnly | workspaceWrite | dangerFullAccess | externalSandbox
```

REQ-CX2-007's write mode is therefore expressible on the turn, and the `plan.md` §D M0 item-4 degrade path for (d) is not exercised.

**Second consequence worth carrying into M3/M4.** Both `effort` and `sandboxPolicy` are documented as applying *"for this turn **and subsequent turns**"* — they are sticky on the thread, not scoped to one turn. Under the recorded fork decision 1 (in-process goroutine holding a reusable session) plus `resume_last` thread reuse (REQ-CX2-008), a write-enabled task would leave its thread write-enabled for a later task that did not opt in. Any implementation of REQ-CX2-007 must therefore reset `sandboxPolicy` explicitly on each turn rather than relying on a per-turn default.

**M0 closure status.** All four probe items carry a recorded observation; none is a recorded absence, so no degrade path fires. Closure is nonetheless **not yet reached**: the `turnId` finding under (a) requires a `REQ-CX2-003` record-shape amendment, and `plan.md` §D M0's closure criterion requires every amendment the probe forces to have landed before M1 starts.

**M0 closure — reached (SPEC v0.5.0).** Both forced amendments have landed in the SPEC artifacts, so the remaining closure condition above is satisfied and M1 may start. (1) `turnId` is now a REQ-CX2-003 record field, sourced from the `turn/started` notification's `turn.id`; REQ-CX2-011 names it as the second required `turn/interrupt` argument; AC-CX2-005 decodes it and AC-CX2-014 asserts it on the sent interrupt; `plan.md` M2/M4 carry the capture and send points. (2) REQ-CX2-007 now requires `sandboxPolicy` on every turn (`readOnly` when not opted in), closing the sticky-policy route around the `allow_write` gate; AC-CX2-010 gained the two-turn reused-thread arm and `plan.md` M3 carries the hazard. `effort`'s identical stickiness was judged a cost/quality drift rather than a safety-boundary crossing and left untreated — see the v0.5.0 HISTORY entry. The Gaps recorded above are unchanged by this closure: they remain open and are carried into M1/M4 as live risks, in particular that no `turn/interrupt` was actually issued and the `Turn.id` → `turnId` correspondence is a schema inference.

**Gaps.**

- The probe read the generated protocol schema; it did **not** execute a live session. (b) in particular is a schema-shape inference — `threadId` being required proves the method addresses an existing thread, not that the server accepts a second turn after the first completes. A two-turn live session would settle it.
- The `Turn.id` → `turnId` correspondence is inferred from naming and type agreement across two schema files, not from an observed `turn/started` payload followed by an accepted `turn/interrupt`.
- No `turn/interrupt` call was actually issued, so its runtime behaviour (in particular whether it is honoured mid-tool-call) is unobserved.
- `codex_app_server_protocol.v2.schemas.json` was emitted alongside the v1 schema and was not examined; if the client negotiates v2, some shapes above may differ.

### M1 — Reusable session handle + model/effort SSOT wiring

REQ-CX2-001, REQ-CX2-002. Evidence captured against the working tree on top of HEAD `355250a01` — the tree the M1 commit lands verbatim. Every row names the command run and the output observed in this run, against this tree.

**Protocol re-read (extends the M0 probe; no SPEC amendment forced).** M0 recorded `model` + `effort` on `TurnStartParams` and stopped there. M1 needed the destination for the *review* path too, so the same generated schema was re-read for the other two methods (`codex app-server generate-json-schema --out <dir>`, `codex-cli 0.146.1`, `jq '.definitions.<T>' ClientRequest.json`):

- `ReviewStartParams` — `{"required":["target","threadId"],"properties":{"delivery":…,"target":…,"threadId":…}}`. It carries **no `model` and no `effort` field at all**. Injecting either would put unknown fields on the Stop-hook gate's own request path, so `review/start` carries neither.
- `ThreadStartParams` — carries `model` (`["string","null"]`), plus `cwd`, `sandbox`, `approvalPolicy`, and others. No `effort`.
- `ReasoningEffort` — `{"description":"A non-empty reasoning effort value advertised by the model.","type":"string","minLength":1}`. Not a closed enum, but an unadvertised value is not guaranteed to be accepted.

Consequence for REQ-CX2-002: the resolved model reaches codex at **two** destinations rather than one — `thread/start` (session-level, the only reachable destination on the `review/start` path) and `turn/start` (per-turn, carrying `model` + `effort`). REQ-CX2-002's clause is "carry the resolved value into the params actually transmitted to codex", which both satisfy; no requirement or acceptance criterion needed amending.

**C7 non-regression guard.** The default profile matrix resolves `sync-auditor` (the audit agent key, shared with the GLM sibling) to `{Model: "opus", Effort: high|medium|low}` — a Claude id the codex app-server cannot serve. Transmitting it would have broken the review gate for every project that never opted in. A resolved model outside the codex-servable families is therefore dropped together with its paired effort, leaving the transmitted request byte-identical to the pre-M1 shape unless the project explicitly configures a codex model. This mirrors the GLM sibling, which filters its own SSOT result through `IsGLMBackend` before using it.

| AC | Status | Command | Observed output |
|----|--------|---------|-----------------|
| AC-CX2-001 (REQ-CX2-001) | PASS | `go test ./internal/cli/ -run TestCodexSession_SecondTurnReusesThread` | `ok github.com/modu-ai/moai-adk/internal/cli 0.871s` — two turns on one handle; exactly 1 `initialize`, 1 `thread/start`, 2 `review/start`, both turns carrying `threadId` `tid-fake`, both returning their result (pass then fail) |
| AC-CX2-002 (REQ-CX2-001) | PASS | `go test ./internal/cli/ -run 'Codex\|ReviewGate' -v` | every pre-existing codex and review-gate test PASS with assertions unmodified, incl. the PRESERVE-listed `TestReviewGateReaders_AgreeWithConfigLoader` and `TestRunCodexReviewRPC_SurfacesServerError` |
| AC-CX2-003 (REQ-CX2-002) | PASS | `go test ./internal/cli/ -run 'TestCodexSession_ResolvedModelReachesTransmittedParams\|TestCodexSession_ExplicitModelOverridesResolved'` | `ok … 0.871s` — no explicit model ⇒ transmitted `thread/start.model` and `turn/start.model` = `gpt-5-codex` with `turn/start.effort` = `high`, both resolved via `template.ResolveAgentModelEffort`; explicit `model` argument ⇒ `o4-mini` transmitted verbatim. The drop at `buildCodexReviewParams` is closed. |
| AC-CX2-004 (REQ-CX2-002) | PASS | `go test ./internal/cli/ -run 'TestMCPAudit_NoDirectFrontmatterRead\|TestCodexSession_ResolvedModelReachesTransmittedParams'` | `ok …` — the negative guard still passes, and it can no longer pass vacuously: the companion positive test asserts the resolver is invoked on the codex path and that its resolved value reaches the transmitted params |

Supporting non-regression rows (not AC-bound): `TestCodexSession_NonCodexModelNotTransmitted` (a Claude id from the SSOT is dropped along with its effort), `TestCodexSession_ReviewStartCarriesNoModelOrEffort` (the review path stays schema-clean while the session model still rides `thread/start`), `TestCodexServableModel` (the servability predicate).

**§E verification batch.**

- E2 cross-platform build — `go build ./... && GOOS=windows GOARCH=amd64 go build ./...` → `HOST_BUILD_EXIT=0`, `WIN_BUILD_EXIT=0`.
- E3 coverage — `go test -cover ./internal/cli/` → `ok github.com/modu-ai/moai-adk/internal/cli 173.287s coverage: 76.6% of statements`. Pre-change baseline measured on a detached worktree at HEAD `355250a01` (`go -C <base> test -cover ./internal/cli/`) → `coverage: 76.5% of statements`. Coverage is above its pre-change level.
- E4 subagent boundary — `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/mcp_codex*.go internal/cli/codex_*.go | grep -v '_test.go' | grep -v '//'` → no output (exit 1).
- E5 lint — `golangci-lint run --timeout=2m` → `0 issues.`
- Full suite — `go test ./...` → every package `ok`; zero `FAIL` lines. Run in full rather than affected-packages-only, so cross-cutting template-mirror guards were exercised.
- Template surface — `git status --porcelain internal/template/templates/` → empty.

**Gaps (M1).** No live codex session was executed; the two-turn reuse (AC-CX2-001) is proven against the canned session runner, so the M0 Gap "(b) is a schema-shape inference — a two-turn live session would settle it" remains open and is carried into M2/M3. The `thread/start.model` and `turn/start.model` destinations are likewise schema-verified and canned-session-verified, not observed against a live codex accepting them.

### M2 — Job registry

REQ-CX2-003, REQ-CX2-004, REQ-CX2-005. Evidence captured against the working tree on top of HEAD `146baf37f` (branch `feat/SPEC-CODEX-PHASE2-001-run`) — the tree the M2 commit lands verbatim. Every row names the command run and the output observed in this run, against this tree.

**What M2 delivers.** `internal/cli/codex_jobs.go` (the record + the registry) and three seams it consumes, added to `internal/cli/mcp_codex.go`: `codexProcessConn` / `codexConnPID` (the pid of the process the session runner spawned), `codexSessionHandle.turnID` + `onTurnStarted` (the mid-flight turnId hook), and `codexTurnIDOf` (reads `turn.id` out of the `turn/started` notification). The record's summary bound is `config.DefaultCodexJobSummaryMaxLen` in `internal/config/defaults.go` (REQ-CX2-015). No tool is registered — `codex_task` is M3 and the `codex_job_*` tools are M4.

**RED evidence (E8, captured before any implementation existed).** `go test ./internal/cli/ -run 'CodexJob|CodexConnPID'`:

```
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/codex_jobs_test.go:48:38: undefined: codexJobRegistry
internal/cli/codex_jobs_test.go:67:9: undefined: newCodexJobRegistry
internal/cli/codex_jobs_test.go:76:25: undefined: codexJobSpec
internal/cli/codex_jobs_test.go:78:26: handle.pid undefined (type *codexSessionHandle has no field or method pid)
internal/cli/codex_jobs_test.go:88:9: handle.onTurnStarted undefined (type *codexSessionHandle has no field or method onTurnStarted)
internal/cli/codex_jobs_test.go:103:10: undefined: CodexJobRecord
internal/cli/codex_jobs_test.go:111:19: undefined: codexJobStatusQueued
internal/cli/codex_jobs_test.go:112:60: undefined: codexJobStatusQueued
internal/cli/codex_jobs_test.go:142:9: undefined: newCodexJobRegistry
internal/cli/codex_jobs_test.go:143:25: undefined: codexJobSpec
internal/cli/codex_jobs_test.go:143:25: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
```

**Two design points worth carrying forward.**

- *The turnId is written mid-flight, not after the turn.* `turnIDRecorder` is installed on the session handle before the turn starts and fires from inside the notification loop, so the id lands in the record while the turn is still running — which is the only window in which the id is useful, since M4 cancels a turn that has not finished. The write there is best-effort by construction (the notification loop has no caller to return an error to), and the consequence is stated rather than hidden: the record keeps an empty `turn_id`, and M4's cancel path must treat an empty `turn_id` as "turn not addressable" instead of sending a malformed `turn/interrupt`.
- *A processless session records pid 0.* `codexProcessConn` is deliberately the optional half of `codexConn`, so a canned session reports no pid at all rather than borrowing one. This keeps REQ-CX2-012's ownership check honest at M4: 0 is "no process of ours", never a pid to signal.

| AC | Status | Command | Observed output |
|----|--------|---------|-----------------|
| AC-CX2-005 (REQ-CX2-003) | PASS | `go test ./internal/cli/ -run TestCodexJobRegistry_RecordShapeAndTurnIDCapture -v` | `--- PASS` — exactly one file under the registry dir; its JSON decodes with a non-empty id, `status: queued`, non-zero `created_at`/`updated_at`, `thread_id` `tid-fake`, `turn_id` `trn-77` (the canned `turn/started` notification's `turn.id`), `pid` 424242 (the pid the spawning connection reports), `mode` `adversarial`, and a non-empty `request_summary` |
| AC-CX2-006 (REQ-CX2-004) | PASS | `go test ./internal/cli/ -race -run TestCodexJobRegistry_TransitionsAreAtomic -v` | `--- PASS` — a concurrent reader loops over the record across five transitions (`running`→`running`→`completed`→`failed`→`cancelled`); every completed read decoded, carried an in-enum status and the right id, and the read count was non-zero (the test fails rather than passing vacuously when no read lands). Race detector clean. A companion row, `TestCodexJobRegistry_RejectsUnknownStatus`, shows an out-of-enum transition is refused with the on-disk record unchanged |
| AC-CX2-007 (REQ-CX2-004) | **PASS-WITH-DEBT** | `go test ./internal/cli/ -run TestCodexJobRegistry_UnwritableStateDirStructuredError -v` | `--- PASS` — with `.moai/state` occupied by a regular file, `create` returns a `*codexJobStateError` carrying op + path + wrapped cause; the test process does not panic; removing the obstruction leaves the same registry able to serve the next call. **Debt**: the AC's clause "*the tool* returns a structured error result" cannot be closed at M2 — `codex_task` does not exist until M3. M2 delivers the structured error value and the `toolErr` rendering path it will travel; the tool-level arm is carried into M3 |
| AC-CX2-008 (REQ-CX2-005) | PASS | `go test ./internal/cli/ -run TestCodexJobRecord_CarriesNoSecrets -v` | `--- PASS` — with credential sentinels seeded into the environment (`CODEX_API_KEY`, `OPENAI_API_KEY`) and into the request text (an `api_key=sk-proj-…` pair, an `Authorization: Bearer …` header, a `password: …` pair), none of the four sentinels appears in the record file, and the record carries no environment block. `TestCodexJobRecord_NoReattachmentMetadata` additionally asserts the serialized record carries no `env` / `argv` / `binary_path` / `server_pid` / `resumable` / `reattach` key |

Supporting non-AC rows: `TestCodexJobRegistry_ConcurrentJobsDoNotCollide` (8 concurrent creates → 8 distinct ids and 8 files, acceptance.md §B), `TestCodexJobRegistry_MalformedRecordIsReported` and `TestCodexJobRegistry_UnknownIDIsNotFound` (§B — a corrupt record and an unknown id are distinguishable structured errors, neither a crash), `TestCodexJobRecord_RequestSummaryBounded`, `TestCodexConnPID_ZeroWithoutProcess`.

**§E verification batch.**

- E2 cross-platform build — `go build ./...` → `HOST_BUILD_EXIT=0`; `GOOS=windows GOARCH=amd64 go build ./...` → `WIN_BUILD_EXIT=0`. `go vet ./internal/cli/` → `VET_EXIT=0`.
- E3 coverage — `go test -cover ./internal/cli/` → `ok github.com/modu-ai/moai-adk/internal/cli 233.121s coverage: 76.7% of statements` (log: `.moai/state/verify/m2/cover-v2.txt`), against the M1 level of 76.6% recorded above. Coverage is at or above its pre-change level. **Flake disclosure**: this command was run four times on this tree — 193.668s PASS (76.7%), then `FAIL` at 227.820s, then 217.810s PASS (76.7%), then the 233.121s PASS quoted above. The failing run's per-test detail is **not attributable**: that invocation kept only the last two output lines, so the log naming the failing test was discarded and no claim is made about which test failed or why. The known local `internal/cli` flake predates this milestone; the immediately-following `go test ./...` run in the same batch reported `internal/cli` `ok` at 226.756s, so the failure did not reproduce against the same tree. A subsequent tightening of `TestCodexJobRegistry_TransitionsAreAtomic` (its reader loop now yields between reads instead of hot-spinning) removed one plausible contributor, but no causal claim is made — the two later PASS runs are consistent with the flake simply not firing.
- E4 subagent boundary — `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/mcp_codex*.go internal/cli/codex_*.go | grep -v '_test.go' | grep -v '//'` → no output (exit 1).
- E5 lint — `golangci-lint run --timeout=3m` → `0 issues.` An earlier run of the same command reported `internal/cli/codex_jobs.go:109:6: func codexJobTerminal is unused`; that helper was M4 forward-scaffolding with no caller and was deleted rather than suppressed, and the re-run above is the result.
- Full suite — `go test ./...` → exit 0, 112 packages `ok`, zero `FAIL` / `panic` / `--- FAIL` lines (`internal/cli` `ok` at 233.121s in that run). Run in full rather than affected-packages-only, so the cross-cutting template-mirror guards were exercised. Output at `.moai/state/verify/m2/full-suite-v2.txt`; an earlier identical run against the same tree (before the test-loop yield) is at `.moai/state/verify/m2/full-suite-final.txt`, also exit 0 / 112 `ok` / 0 failures.
- Template surface — `git status --porcelain internal/template/templates/` → empty.
- `gofmt -l` on the four touched files (`internal/cli/codex_jobs.go`, `internal/cli/codex_jobs_test.go`, `internal/cli/mcp_codex.go`, `internal/config/defaults.go`) → no output. A repo-wide `gofmt -l .` lists ~35 files, all of them untouched by this SPEC — a pre-existing baseline, not an M2 regression.

**Gaps (M2).**

- No live codex session was executed. The `turn/started` → `turn.id` capture is proven against a canned transcript **that this milestone authored from the M0 schema reading**, so it inherits the M0 gap verbatim: the `Turn.id` → `turnId` correspondence remains a schema inference, and no `turn/interrupt` has ever been issued. A canned transcript cannot falsify a wrong reading of the protocol — only a live two-turn session with a real interrupt can, and that is still owed.
- Whether codex emits `turn/started` at all on a `review/start` turn (as opposed to a `turn/start` turn) is unobserved. If it does not, the record's `turn_id` stays empty on the review path and M4's cancel degrades to process termination for those jobs.
- AC-CX2-007's tool-level arm is deferred to M3 (see the PASS-WITH-DEBT row above). The registry half is closed.
- The atomicity evidence is a same-process concurrent read against the POSIX rename guarantee. The Windows rename-retry path (`renameWithRetry`) is exercised only by the cross-compile build, not by a running Windows reader.
- `codexJobSummary` redacts credential-*shaped* content. A secret with no recognizable shape and no credential-shaped key beside it would survive into the summary; the record's structural guarantee (no environment block, no argv, no binary path) is the stronger half of REQ-CX2-005, and the redaction is defense in depth on top of it.

#### M2 correction — the atomicity test flaked ~1-10% (found by orchestrator re-verification of `9aa60a014`)

**The original M2 §E claim did not hold under repetition, and this correction records why.** The M2 evidence above reports `TestCodexJobRegistry_TransitionsAreAtomic` PASS against `-race` and against several full-suite runs. Every one of those runs was `-count=1`. Under repetition the test fails: the orchestrator's `-count=60` sweep produced 6 failures across 60 iterations, and an independent reproduction here produced 1 failure across 200 iterations on an idle machine — the rate is load-dependent, which is exactly why single runs never surfaced it.

This is the same sampling error the M4 correction named, committed one milestone earlier and not caught at the time. The M2 §E batch even ran the coverage command four times and disclosed a flake, but attributed it to the pre-existing `internal/cli` flake and explicitly declined to name a cause — while a test *this milestone authored* was flaking in the same package. The disclosure was honest and the inference was wrong.

**Reproduction (pre-fix), against HEAD `9aa60a014`:**

```
$ go test -count=200 -run 'TestCodexJobRegistry_TransitionsAreAtomic' ./internal/cli/
--- FAIL: TestCodexJobRegistry_TransitionsAreAtomic (0.00s)
    codex_jobs_test.go:236: the concurrent reader never completed a read — the assertion proved nothing
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	5.134s
FAIL
```

**The guard was right; the test around it was wrong.** The message is the test's own `reads == 0` guard, and it fired truthfully: the writer's five transitions can complete before the reader goroutine's first `os.ReadFile` lands, and the guard then correctly refuses to report a pass for an assertion that observed nothing. Removing the guard, softening it to a warning, gating it on a retry count, or raising the iteration count until green would each have converted a visible flake into a test that passes without asserting — strictly worse. The defect was that the reader's arrival inside the transition window was left to **timing**.

**Fix — the observation is now a rendezvous, not a race.** The writer performs the first transition (to `running`) and then blocks on a channel the reader closes once it has decoded a `running` record; only then does it proceed to the remaining four transitions. While the writer is blocked, the on-disk record *stays* `running`, so the state the reader is waiting for cannot expire before the reader sees it — there is no window to miss. `reads == 0` is therefore unreachable by construction rather than merely unlikely, and the guard is retained unchanged as a defense against a future edit reintroducing the hole.

Two properties keep the rendezvous from trading a flake for a hang: the reader also closes a `readerStopped` channel on exit, and the writer selects on both — so a reader that dies early (a decode failure, an out-of-enum status) fails the test with that reason instead of blocking forever. **There is no deadline anywhere in the path**, so no timing input reaches the pass decision. The reader's former `time.Sleep(50µs)` yield is now `runtime.Gosched()`; with the writer blocked on the reader, a wall-clock sleep has no purpose.

Explicitly not done: no iteration-count tuning, no sleep to widen the window, no `t.Skip`, no retry wrapper, no change to the guard at what is now line 278.

| Claim | Command | Observed output (verbatim) |
|---|---|---|
| The flake is gone under heavy repetition | `go test -count=200 -run 'TestCodexJobRegistry_TransitionsAreAtomic' ./internal/cli/` | `ok  	github.com/modu-ai/moai-adk/internal/cli	4.522s` |
| The wider codex/job surface is stable under repetition | `go test -count=60 -run 'Codex\|Job\|Task\|Cancel' ./internal/cli/` | `ok  	github.com/modu-ai/moai-adk/internal/cli	40.692s` |
| No data race in the new rendezvous | `go test -race -run 'Codex\|Job\|Task\|Cancel' ./internal/cli/` | `ok  	github.com/modu-ai/moai-adk/internal/cli	10.514s` |
| No regression package-wide | `go test ./...` | exit 0; 112 packages `ok`; zero `FAIL` / `panic` / `--- FAIL` lines; `ok  	github.com/modu-ai/moai-adk/internal/cli	254.330s` |
| Both builds | `go build ./... && GOOS=windows GOARCH=amd64 go build ./...` | no output, exit 0 |
| Lint | `golangci-lint run --timeout=3m` | `0 issues.` |

Baseline attribution: every row above was captured against the working tree on top of HEAD `9aa60a014` — the tree the correction commit lands verbatim. The pre-fix reproduction was captured against `9aa60a014` itself, before any edit.

**Gaps (M2 correction).**

- **The rendezvous proves the reader is inside the sequence; it does not prove it is inside a `rename` call.** What the test now guarantees deterministically is that a concurrent reader observes a mid-sequence record, and that every record it observes is complete and in-enum. Catching a reader *during* the kernel's rename would require injecting a seam into the write path, which this SPEC's scope does not justify — POSIX rename atomicity is the property being relied on, and it is a platform guarantee rather than something this test can demonstrate.
- **The correction says nothing about the pre-existing `internal/cli` flake.** The M2 §E coverage-run failure disclosed above remains unattributed: its per-test output was discarded at the time and cannot be recovered. It is plausible that it was this test, and it is equally plausible that it was the unrelated `TestNavigatorEnrich_AtomicWriteBarrier`. Both are consistent with the evidence; neither is established, and this correction does not claim to have fixed it.
- **`-count=200` is a sampling bound, not a proof.** A rendezvous with no deadline makes the failure mode structurally unreachable, which is the actual argument; 200 clean iterations are corroboration, not the basis of the claim.

### M3 — `codex_task`

REQ-CX2-006, REQ-CX2-007, REQ-CX2-008. Evidence captured against the working tree on top of HEAD `628422d8c` (branch `feat/SPEC-CODEX-PHASE2-001-run`) — the tree the M3 commit lands verbatim. Every row names the command run and the output observed in this run, against this tree.

**What M3 delivers.** `internal/cli/codex_task.go` (`handleCodexTask`, the background-job goroutine, and the live-session map M4 will read), plus four seams added to `internal/cli/mcp_codex.go`: `codexSandboxPolicy`, `readCodexTaskAllowWrite`, `codexMethodThreadResume` + `openCodexSessionOn`, and the `allow_write` field on `handleCodexSetup`'s result. `codexJobRegistry.latestThreadID` is added to `internal/cli/codex_jobs.go`. The config key is typed at `internal/config/types.go` (`CodexTaskConfig.AllowWrite`) with its distributed default `false` at `internal/config/defaults.go` (REQ-CX2-015). No tool is registered — registration with JSON Schema and read-only hints is M5.

**Protocol re-read (extends the M0 probe; no SPEC amendment forced).** M0 recorded `sandboxPolicy` as a 4-variant union and stopped at the variant names. M3 needed the wire ENVELOPE, so the generated schema was re-read (`codex app-server generate-json-schema --out <dir>`, `codex-cli 0.146.1`, `jq '.definitions.SandboxPolicy' ClientRequest.json`):

- `SandboxPolicy` is a `oneOf` of **four objects**, each with `type` in its `required` list — it is an **internally-tagged object** (`{"type":"readOnly"}`), NOT a bare string. This is the same shape `target` has, and PR #1430 learned that shape by being rejected with JSON-RPC -32600. Transmitting `"sandboxPolicy":"readOnly"` would have been rejected the same way.
- `AC-CX2-010`'s sticky arm says turn 1's params "carry `sandboxPolicy` `workspaceWrite`" and turn 2's "carry `sandboxPolicy` `readOnly`". The AC names the **variant**, which is what the M0 probe observed; it does not specify the JSON envelope. The implementation sends the variant inside the tagged object the protocol declares, and the test asserts the `type` discriminant equals the literal variant. **No requirement or acceptance criterion needed amending** — a literal bare-string reading would have specified a request codex rejects.
- `ThreadResumeParams` requires only `{threadId}` and its response carries the same `{thread:{id}}` shape `thread/start` returns, so `extractThreadID` reads both unchanged. `resume_last` therefore sends the real `thread/resume` method rather than silently skipping the handshake step.

**RED evidence (E8, captured before any implementation existed).** `go test ./internal/cli/ -run 'CodexTask|CodexSetup_ReportsAllowWrite'`:

```
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/codex_task_test.go:73:14: undefined: handleCodexTask
internal/cli/codex_task_test.go:213:30: undefined: codexSandboxReadOnly
internal/cli/codex_task_test.go:214:89: undefined: codexSandboxReadOnly
internal/cli/codex_task_test.go:215:68: undefined: codexSandboxReadOnly
internal/cli/codex_task_test.go:216:54: undefined: codexSandboxReadOnly
internal/cli/codex_task_test.go:217:51: undefined: codexSandboxWorkspaceWrite
internal/cli/codex_task_test.go:264:67: undefined: codexTaskMode
internal/cli/codex_task_test.go:293:21: undefined: codexSandboxWorkspaceWrite
internal/cli/codex_task_test.go:294:74: undefined: codexSandboxWorkspaceWrite
internal/cli/codex_task_test.go:300:15: undefined: codexSandboxReadOnly
internal/cli/codex_task_test.go:300:15: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
```

**Three design points worth carrying forward.**

- *The non-writing turn is the one that sends the field.* `codexSandboxPolicy` is called on every turn `codex_task` starts, and `buildCodexReviewParams` forwards the value only when the caller supplied one. `codex_task` always supplies it; `codex_audit` and the review gate never do, so their request stays byte-identical to its pre-M3 shape (C7).
- *`resume_last` resumes the last **recorded** thread, and records are created for background jobs.* A project that has only ever run foreground tasks has no recorded thread, so `resume_last` opens a new one and says so in the result (REQ-CX2-008's second arm). This is the literal reading of "the most recently recorded threadId" — records are the only place a thread is recorded, and REQ-CX2-003 creates them for background jobs. No new state surface was added to widen it.
- *The background record is created BEFORE the turn is handed off.* An unwritable state directory is therefore reported to the caller as a structured error rather than surfacing later as a job nobody can observe, and the session is torn down on that path so a job that cannot be recorded never leaves a codex process running.

| AC | Status | Command | Observed output |
|----|--------|---------|-----------------|
| AC-CX2-009 (REQ-CX2-006) | PASS | `go test ./internal/cli/ -run 'TestCodexTask_ForegroundReturnsOutput\|TestCodexTask_BackgroundReturnsJobIDResolvingToRecord' -v` | `--- PASS` ×2 — `background:false` ⇒ the result carries `output` "the refactor is complete" and no job id; `background:true` ⇒ the result carries a job id that `registry.load` resolves, with no `output` on the immediate result, and the record reaches `completed` carrying the task output and `turn_id` `trn-bg` (captured mid-flight) |
| AC-CX2-010 (REQ-CX2-007) | PASS | `go test ./internal/cli/ -run 'TestCodexTask_WriteGateFailClosed\|TestCodexTask_SandboxPolicyResetOnReusedThread\|TestCodexSetup_ReportsAllowWriteState' -v` | `--- PASS` ×3. **Gate arm**: 5 sub-cases (absent-file / absent-key / malformed-YAML / explicit-`false` / explicit-`true`); the first four read as not opted in and transmit `sandboxPolicy.type` `readOnly` with a non-empty note stating the write was not honored, the fifth transmits `workspaceWrite`. **Sticky arm**: two turns on ONE reused thread (`thread/start` sent 0 times, `thread/resume` twice, both turns addressing `tid-fake`) — turn 1 (write, opted in) transmits `workspaceWrite`, turn 2 (no write request) transmits `readOnly` and the field is PRESENT, so the thread's inherited policy cannot outlive the request that opted into it. **`codex_setup` arm**: the decoded result map carries `allow_write` equal to the literal expected value per state — `false` / `false` / `false` / `true` — alongside `enable_review_gate` |
| AC-CX2-011 (REQ-CX2-008) | PASS | `go test ./internal/cli/ -run 'TestCodexTask_ResumeLastReusesRecordedThread\|TestCodexTask_ResumeLastWithNoRecordedThread' -v` | `--- PASS` ×2 — with a recorded thread, `thread/start` is sent 0 times, `thread/resume` once carrying `threadId` `tid-fake`, the turn addresses `tid-fake`, and the result reports `resumed_thread:true` with `thread_id` `tid-fake`; with none recorded, `thread/start` is sent once and the result reports `resumed_thread:false` with a note stating no prior thread was resumed |
| AC-CX2-007 (REQ-CX2-004) — **inherited debt, now CLOSED** | PASS | `go test ./internal/cli/ -run TestCodexTask_UnwritableStateDirStructuredError -v` | `--- PASS` — with `.moai/state` occupied by a regular file, a background `codex_task` returns `IsError` (a structured error result, not a Go error and not a panic), and the immediately-following foreground call is still served. M2 recorded this AC as PASS-WITH-DEBT because its "*the tool* returns a structured error result" clause needed `codex_task`; the tool-level arm is now closed and the AC is PASS outright |

Supporting non-AC rows: `TestCodexTaskAllowWriteReader_AgreesWithConfigLoader` (the hand-rolled reader agrees with `config.NewLoader().Load` on on / off / absent-block / absent-file — the parallel of the PRESERVE-listed `TestReviewGateReaders_AgreeWithConfigLoader`, guarding the exact top-level-vs-nested key-path bug that once made the sibling toggle unreadable), `TestCodexTaskAllowWrite_DistributedDefaultIsFalse` (AP-5), `TestCodexTask_FailOpenOnMissingCodex` and `TestCodexTask_MissingPromptIsStructuredError` (acceptance.md §B).

**§E verification batch.**

- E2 cross-platform build — `go build ./...` → `HOST_BUILD_EXIT=0`; `GOOS=windows GOARCH=amd64 go build ./...` → `WIN_BUILD_EXIT=0`. `go vet ./internal/cli/ ./internal/config/` → `VET_EXIT=0`.
- E3 coverage — `go test -cover ./internal/cli/` → `ok github.com/modu-ai/moai-adk/internal/cli 177.271s coverage: 76.7% of statements` (log: `.moai/state/verify/m3/cover-v2.txt`), against the M2 level of 76.7%. Coverage is at its pre-change level. **Flake disclosure**: the first coverage run on this tree FAILED at 172.023s, and unlike M2 this failure IS attributable — `--- FAIL: TestNavigatorEnrich_AtomicWriteBarrier (0.05s)` / `navigator_enrich_test.go:91: barrier file not created (goroutine did not reach barrier)` (log: `.moai/state/verify/m3/cover.txt` line 494). That test is a goroutine-timing barrier belonging to `SPEC-PROJECT-NAVIGATOR-003` (`6b0f124d0`); it touches no codex code and this SPEC touches no navigator code. `go test ./internal/cli/ -run TestNavigatorEnrich_AtomicWriteBarrier -count=10` → `ok … 1.176s` (10/10 pass in isolation), and both full-suite runs on this tree reported `internal/cli` `ok`. The failure is a timing flake in an unrelated test, not an M3 regression — but it did fail once on this tree, and that is stated rather than dropped.
- E4 subagent boundary — `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/mcp_codex*.go internal/cli/codex_*.go | grep -v '_test.go' | grep -v '//'` → no output (exit 1).
- E5 lint — `golangci-lint run --timeout=3m` → `0 issues.`
- Full suite — `go test ./...` → exit 0, 112 packages `ok`, zero `FAIL` / `--- FAIL` / `panic` lines (`internal/cli` `ok` at 201.576s). Run in full rather than affected-packages-only, so the cross-cutting template-mirror guards were exercised. Output at `.moai/state/verify/m3/full-suite-final.txt`; an earlier identical run before the two config-agreement tests were added is at `.moai/state/verify/m3/full-suite.txt`, also exit 0 / 112 `ok` / 0 failures.
- Race — `go test -race -run 'Codex|Job|Task' ./internal/cli/` → `ok … 2.290s`. The background-job goroutine is covered; the test polls the record file and never reads the shared fake session's `sent` slice while the goroutine is appending to it.
- Template surface — `git status --porcelain internal/template/templates/` → empty. `grep -rln 'allow_write' internal/template/templates/` → no match (exit 1): the new config key does not leak into the distributed surface (§25 / REQ-CX2-015).
- Constant placement (REQ-CX2-015, an AC-CX2-016 sub-check landing early) — `grep -n 'AllowWrite' internal/config/types.go` → `586`/`592` (the typed field); `grep -n 'AllowWrite' internal/config/defaults.go` → `670` (the distributed default `false`).
- `gofmt -l` on the six touched files → no output. A repo-wide `gofmt -l .` still lists the same ~35 pre-existing files, none touched by this SPEC.

**Gaps (M3).**

- **No live codex session was executed, and this is now the third milestone in a row saying so.** Every M3 assertion about the wire — that `sandboxPolicy` is accepted as a tagged object, that `thread/resume` re-opens a thread this process did not start, that a turn on a resumed thread is accepted — is verified against the generated protocol schema plus a canned transcript **this milestone authored from that schema**. A canned transcript cannot falsify a wrong reading of the protocol. The M0/M1/M2 gap is inherited verbatim and compounded.
- **The stickiness itself is unobserved.** The hazard REQ-CX2-007 exists to close is documented in the schema ("for this turn and subsequent turns"), not observed: no live session has demonstrated that a `workspaceWrite` turn actually leaves its thread writable for a subsequent turn. The mitigation is correct whether or not the hazard is real (sending `readOnly` explicitly is harmless if the field turns out to be turn-scoped), so the unobserved premise costs nothing — but the claim "this closes a real route around the gate" rests on the schema's own wording alone.
- **`resume_last` resumes only a thread recorded by a background job.** A foreground-only project has nothing to resume. This is the literal reading of REQ-CX2-008 against REQ-CX2-003, and the result states the outcome rather than hiding it — but a user who expects `resume_last` to continue their last *foreground* task will get a new thread and a note.
- **The write gate bounds what moai REQUESTS, not what codex enforces.** `codex_task` transmits `readOnly` when not opted in; whether the codex app-server honors the policy is codex's guarantee, unobserved here. The gate is a request-side control, not a sandbox.
- **`turn_id` may be empty.** M2's open gap — whether codex emits `turn/started` on a `review/start` turn — is untouched. `codex_task` drives `turn/start`, where the notification is the method's own, so the task path is the better-founded case; the M4 cancel path must still treat an empty `turn_id` as "not addressable".
- **The live-session map is in-process only.** `codexLiveJobSessions` holds a handle for the lifetime of the server process. A record found `running` with no entry is stale by construction — exactly the case REQ-CX2-012 requires M4 to refuse — but nothing in M3 verifies that M4 does so.

### M4 — Job control

REQ-CX2-009, REQ-CX2-010, REQ-CX2-011, REQ-CX2-012. Evidence captured against the working tree on top of HEAD `3419349c7` (branch `feat/SPEC-CODEX-PHASE2-001-run`) — the tree the M4 commit lands verbatim. Every row names the command run and the output observed in this run, against this tree.

**What M4 delivers.** `internal/cli/codex_job_control.go` (`handleCodexJobStatus`, `handleCodexJobResult`, `handleCodexJobCancel`, the process-termination seam, and the bounded grace wait), plus three seams: `codexMethodTurnInterrupt` + `codexSessionHandle.sendTurnInterrupt` + a mutex on the handle's request-id counter in `internal/cli/mcp_codex.go`, and `codexJobTerminal` in `internal/cli/codex_jobs.go`. The grace window and its poll interval are `config.DefaultCodexJobCancelGrace` / `config.DefaultCodexJobCancelPoll` in `internal/config/defaults.go` (REQ-CX2-015). No tool is registered — registration with JSON Schema and read-only hints is M5.

**RED evidence (E8, captured before any implementation existed).** `go test ./internal/cli/ -run 'CodexJobStatus|CodexJobResult|CodexJobCancel'`:

```
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/codex_job_control_test.go:136:10: undefined: codexTerminateProcess
internal/cli/codex_job_control_test.go:137:2: undefined: codexTerminateProcess
internal/cli/codex_job_control_test.go:143:21: undefined: codexTerminateProcess
internal/cli/codex_job_control_test.go:155:10: undefined: codexJobCancelGrace
internal/cli/codex_job_control_test.go:156:2: undefined: codexJobCancelGrace
internal/cli/codex_job_control_test.go:157:21: undefined: codexJobCancelGrace
internal/cli/codex_job_control_test.go:228:26: undefined: codexMethodTurnInterrupt
internal/cli/codex_job_control_test.go:256:29: undefined: handleCodexJobStatus
internal/cli/codex_job_control_test.go:276:33: undefined: handleCodexJobStatus
internal/cli/codex_job_control_test.go:284:33: undefined: handleCodexJobStatus
internal/cli/codex_job_control_test.go:284:33: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
```

**Four design points worth carrying forward.**

- *Ownership is decided by live-session membership, not by the record's pid.* `codexLiveJobSessions` holds an entry exactly for the jobs this server lifetime started, and `runCodexBackgroundJob` deletes it on every exit path — so its absence on a non-terminal record means the record outlived the process that owned it (REQ-CX2-012). The pid that termination targets is read from the live session's own connection, never from the record file, so a stale or tampered record cannot direct a signal anywhere at all. This is strictly stronger than comparing the recorded pid against a set of spawned pids, and it is why AC-CX2-015's refusal needs no pid comparison.
- *Termination is one `os.Process.Kill` on one pid, and no escalation.* The graceful stage already happened — `turn/interrupt` was sent and the grace window elapsed — so a SIGTERM-then-SIGKILL ladder would only extend an already-bounded call, and `syscall.SIGTERM` does not exist on Windows, which would have forced a build-tag split for a stage the flow does not need (`plan.md` §B5). No name-matched or pattern-matched kill exists anywhere in the tree (AP-4 sweep below).
- *`turn/interrupt` is sent and NOT awaited.* The job's own goroutine is inside `awaitCodexTurnReview` draining the connection; a second reader would race it for lines and could swallow the `turn/completed` the turn loop is waiting for. The outcome is observed as the job leaving the live-session map instead, which is why the grace wait polls the map rather than the record file — the entry is removed by the goroutine itself, so its absence means the turn is genuinely done rather than merely marked done.
- *The cancelled status is written BEFORE the interrupt is sent.* `runCodexBackgroundJob` keeps a `cancelled` status untouched when its turn returns; this ordering is what puts the status there in time for that guard to see it. The test asserts the status still reads `cancelled` after the blocked turn is released.

| AC | Status | Command | Observed output |
|----|--------|---------|-----------------|
| AC-CX2-012 (REQ-CX2-009) | PASS | `go test ./internal/cli/ -run 'TestCodexJobStatus_ReturnsRecordAndNotFound\|TestCodexJobStatus_MalformedRecordIsReported' -v` | `--- PASS` ×2 — a known id returns the record with `id` / `status` / `thread_id` `tid-fake` / `turn_id` `trn-status` / `mode` `task` / `pid` 424242; an unknown id returns `IsError` with the id named in the text, not a Go error; a missing `job_id` returns `IsError`; a present-but-malformed record file is reported as an error result rather than decoded into an empty record (acceptance.md §B) |
| AC-CX2-013 (REQ-CX2-010) | PASS | `go test ./internal/cli/ -run TestCodexJobResult_TerminalReturnsOutputRunningReturnsStatus -v` | `--- PASS` — a `completed` job returns `output` "the refactor is complete" with `terminal:true`; a `running` job returns `status: running`, `terminal:false`, no output, and the call returned in well under the 1s bound the test asserts (measured, not assumed); an unknown id returns `IsError` |
| AC-CX2-014 (REQ-CX2-011) | PASS | `go test ./internal/cli/ -run TestCodexJobCancel_SendsInterruptAndTerminates -v` | `--- PASS` (0.09s) — against a background job whose turn never completes, the sent lines carry a `turn/interrupt` request whose params include BOTH required arguments, `"threadId":"tid-fake"` and `"turnId":"trn-cancel"`, read from that job's record; the recorded status becomes `cancelled`; the turn does not end within the (shortened) grace window, so termination is attempted on exactly `[424242]` — the pid of the process the live session spawned — and the call returns well inside the asserted 5s bound. Releasing the blocked turn afterwards leaves the status `cancelled`, so a late turn cannot overwrite the cancel |
| AC-CX2-015 (REQ-CX2-012) | PASS | `go test ./internal/cli/ -run TestCodexJobCancel_RefusesRecordThisServerDidNotSpawn -v` | `--- PASS` — a record in `running` naming thread `tid-stale`, turn `trn-stale`, and pid 999999, with no live-session entry (the shape a previous server lifetime leaves), is refused with a structured `IsError` result naming the job; the recorded termination seam observed **zero** attempts. Asserted by recording signal attempts, never by observing a live process |

Supporting non-AC rows: `TestCodexJobCancel_EmptyTurnIDDegradesToTermination` (the branch M2 and M3 both left open — a record whose `turn/started` was never observed carries an empty `turn_id`, so no `turn/interrupt` is sent at all, a note states why, and the cancel degrades to terminating the spawned process; asserted by the absence of any `turn/interrupt` line), `TestCodexJobCancel_TerminalJobSendsNothing` (acceptance.md §B — an already-terminal job returns its terminal status, sends nothing, signals nothing, and is not an error).

**§E verification batch.**

- E2 cross-platform build — `go build ./...` → `HOST_BUILD_EXIT=0`; `GOOS=windows GOARCH=amd64 go build ./...` → `WIN_BUILD_EXIT=0`. `go vet ./internal/cli/ ./internal/config/` → `VET_OK`.
- E3 coverage — `go test -cover ./internal/cli/` → `ok github.com/modu-ai/moai-adk/internal/cli 171.461s coverage: 76.8% of statements` (log: `.moai/state/verify/m4/cover.txt`), against the M3 level of 76.7%. Coverage is above its pre-change level. No flake fired on this run.
- E4 subagent boundary — `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/mcp_codex*.go internal/cli/codex_*.go | grep -v '_test.go' | grep -v '//'` → no output (exit 1).
- E5 lint — `golangci-lint run --timeout=3m` → `0 issues.` An earlier run of the same command reported `internal/cli/codex_job_control.go:234:2: QF1002: could use tagged switch on rec.TurnID (staticcheck)`; the tagless `switch` was rewritten as an `if`/`else if` chain and the re-run above is the result.
- Full suite — `go test ./...` → exit 0, 112 packages `ok`, zero `FAIL` / `panic` lines (`internal/cli` `ok` at 176.555s). Run in full rather than affected-packages-only, so the cross-cutting template-mirror guards were exercised. Output at `.moai/state/verify/m4/full-suite.txt`. The known local flake `TestNavigatorEnrich_AtomicWriteBarrier` (an unrelated SPEC's goroutine-timing barrier, disclosed under M3) did not fire on either the coverage run or the full-suite run of this tree.
- Race — `go test -race -run 'Codex|Job|Task|Cancel' ./internal/cli/` → `ok … 2.521s`. This milestone introduces genuine cross-goroutine sharing of one session — the job goroutine drives the turn while the cancel path sends an interrupt on the same connection — so the handle's request-id counter gained a mutex and the test's fake connection records sends under one.
- AP-4 sweep (no process-kill primitive entered the tree) — `grep -rn 'pkill\|killall\|exec.Command("kill"' internal/cli/codex_*.go internal/cli/mcp_codex*.go` → exactly one match, `codex_job_control.go:22`, inside the doc comment naming the anti-pattern it forbids. No name-matched or pattern-matched kill exists in code. `grep -rn 'Process.Kill()\|proc.Kill()'` → `codex_job_control.go:102` (the precise kill on the live session's pid) and the pre-existing `mcp_codex.go:401` (`realCodexConn.close`'s bounded-wait-then-kill of its own subprocess, untouched by this milestone).
- Template surface — `git status --porcelain internal/template/templates/` → empty.
- `gofmt -l` on the five touched files (`internal/cli/codex_job_control.go`, `internal/cli/codex_job_control_test.go`, `internal/cli/codex_jobs.go`, `internal/cli/mcp_codex.go`, `internal/config/defaults.go`) → no output. A repo-wide `gofmt -l .` still lists the same ~35 pre-existing files, none touched by this SPEC.

**Gaps (M4).**

- **No live codex session was executed, and this is now the fourth milestone in a row saying so.** The interrupt is asserted against a canned transcript this milestone authored from the M0 schema reading. **No `turn/interrupt` has ever been issued against a real codex**, so whether the server honours it — and in particular whether it honours it mid-turn — remains exactly as unobserved as M0 recorded it. The M0 gap is inherited verbatim and is now load-bearing for a shipped code path rather than for a plan.
- **The grace window's expiry, not the interrupt's effect, is what the cancel test exercises.** The canned turn never yields, so every cancel test lands on the termination branch. The branch where a real codex honours `turn/interrupt` and the turn ends inside the grace window is covered only by the code path's own structure (`codexAwaitJobExit` returning true), not by an observation.
- **Termination is recorded, never performed.** `codexTerminateProcess` is swapped for a recorder in every test, so `os.FindProcess` + `Kill` is exercised by no test at all — deliberately, since a test that kills a real process is the test AP-4 warns about. Its cross-platform behaviour rests on the `os` package's contract and the Windows cross-compile, not on a running Windows process.
- **A killed process is not waited on here.** After termination the job's goroutine observes EOF, updates the record, and closes the session; the cancel call returns before that happens. A caller that reads the record immediately after a cancel sees `cancelled` (written before the interrupt), which is correct — but the underlying process reaping is asynchronous and unobserved by the tool.
- **The live-session map is the ownership oracle, and it is in-process only.** That is what makes the REQ-CX2-012 refusal sound, and it is also its limit: a record left behind by a crashed server can never be cancelled through this tool, only observed. Nothing cleans such records up; `codex_job_status` will keep reporting them as `running` forever.
- **`codex_job_result` reports a `failed` job's error text as recorded.** It does not distinguish a codex-side failure from a transport failure, because the record does not — `runCodexBackgroundJob` writes `out.Summary` into `Error` for both.

#### M4 correction — a post-cancel record write (found by orchestrator re-verification of `2c851678e`)

**The earlier §E claim did not hold.** The M4 batch above reports `go test ./...` → exit 0, 112 `ok`, zero failures. On the committed tree at `2c851678e` the orchestrator's independent re-run returned exit 1:

```
--- FAIL: TestCodexJobCancel_EmptyTurnIDDegradesToTermination (0.08s)
    testing.go:1464: TempDir RemoveAll cleanup: unlinkat /var/folders/.../TestCodexJobCancel_EmptyTurnIDDegradesToTermination4144156776/001/.moai/state/codex-jobs: directory not empty
```

**Why the original run missed it.** The claim was true of the run that produced it and false of the tree. The defect is a race between the test's `t.TempDir()` cleanup and a goroutine still writing, so it fires only when the goroutine's write lands inside the `RemoveAll` window — a single `go test ./...` pass samples that window once per test. The evidence was single-sample where the property was probabilistic, and nothing in the batch repeated the affected test. `-count=20` on the single test reproduces it within a handful of iterations; the full suite did not. A green single run was therefore never sufficient evidence for "this test passes", and reporting it as such was the error — not the run itself.

**Root cause — one test defect and one production defect, distinguishable.**

*Proximate (test).* `startHangingBackgroundJob` starts a background job whose turn never completes, and the cleanup released the blocked turn without waiting for its goroutine. `t.TempDir()`'s `RemoveAll` is registered first and so runs last, i.e. after the release: the goroutine's terminal transition calls `write`, whose `os.MkdirAll` recreates `.moai/state/codex-jobs/` while `RemoveAll` is walking it. The distinguishing evidence is the asymmetry between the two cancel tests, which differ in exactly this respect and nothing else:

```
$ go test -run 'TestCodexJobCancel_EmptyTurnIDDegradesToTermination' -count=20 ./internal/cli/   # does NOT join its goroutine
--- FAIL: TestCodexJobCancel_EmptyTurnIDDegradesToTermination (0.08s)
    testing.go:1464: TempDir RemoveAll cleanup: unlinkat ... directory not empty   (2 of 20)
FAIL
$ go test -run 'TestCodexJobCancel_SendsInterruptAndTerminates' -count=20 ./internal/cli/        # joins: unblocks + waits in the body
ok  	github.com/modu-ai/moai-adk/internal/cli	4.164s
```

*Underlying (production, and NOT in M4).* The race exposed a real write that should not exist. `runCodexBackgroundJob` guarded the cancelled status inside its **mutator**, and a mutator can decline to CHANGE the record but cannot decline the WRITE: `registry.update` persists whatever the mutator leaves behind and refreshes `UpdatedAt` unconditionally. So a job's goroutine finishing after `codex_job_cancel` returned still rewrote the record. The status survived — that guard worked — but the file changed after the tool reported the job cancelled. Under the M0 in-process decision that goroutine sits inside the long-lived `moai mcp-server`, so this is not a test artifact: a caller that cancels and then reads sees a record whose `updated_at` moves under it. Pinned by a new deterministic regression test using `updated_at` as the witness (it is refreshed on every write, so an unchanged value proves no write occurred). RED, 3 of 3 runs before the fix:

```
--- FAIL: TestCodexJobCancel_NoRecordWriteAfterCancel (0.09s)
    codex_job_control_test.go:561: updated_at moved from 2026-08-10T16:02:23.851602Z to 2026-08-10T16:02:23.930217Z: a record write landed AFTER codex_job_cancel returned; a job already recorded cancelled must produce no further write
```

**This is M2 + M3 code, not M4 alone.** The unconditional write is `codexJobRegistry.update` (M2); the guard placed where it could not prevent it is `runCodexBackgroundJob` (M3). M4's cancel path did not introduce either — it was the first caller to make the ordering observable. The fix lands in `internal/cli/codex_jobs.go` and `internal/cli/codex_task.go` accordingly: `updateUnlessCancelled` moves the guard into the registry, the only layer that can skip a write, and the background runner uses it. AC-CX2-006's atomicity evidence is unaffected (the write path itself is unchanged); the M2/M3 §E rows above stand, with this note as the amendment.

**Rejected fix — making cancel wait for the goroutine.** Considered and refused: `codex_job_cancel` would then block until the turn ends, which is precisely what REQ-CX2-011 designs against (the interrupt is deliberately not awaited) and what AC-CX2-014 forbids ("the call still returns within a bounded time"). Waiting on the goroutine is waiting on the turn. The bound is kept; the stray write is removed instead.

Also rejected, per the re-verification instruction and because each hides the signal rather than removing it: a `t.Cleanup` that force-empties the directory, moving the registry out of `t.TempDir()`, a `time.Sleep` to widen the window, and silencing the cleanup error. The test now joins the goroutine it started — the `codexLiveJobSessions` entry is deleted by the goroutine AFTER its terminal write, so its absence is a sound join point — which is what acceptance.md §B ("no goroutine leaks past the test") asked for all along.

**Re-verification (fix tree, on top of HEAD `2c851678e`).**

```
$ go test -run 'TestCodexJobCancel' -count=20 ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	6.050s
exit=0

$ go test -race -run 'Codex|Job|Task|Cancel' ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	2.608s
exit=0

$ go build ./... && GOOS=windows GOARCH=amd64 go build ./...
build_exit=0

$ golangci-lint run --timeout=3m
0 issues.
lint_exit=0

$ go test ./...
full_exit=0        # 112 packages ok, zero --- FAIL / FAIL / panic lines; internal/cli ok at 185.220s
                   # log: .moai/state/verify/m4/full-suite-fix.txt
```

The `-count=20` run is the load-bearing one: it is what the original single-pass batch lacked, and it is the evidence the earlier claim should have rested on.

**Flake disclosure (unrelated package, observed on this tree).** A later confirmation run of `go test -count=1 ./...` on the committed fix tree returned exit 1 on a package this SPEC does not touch:

```
--- FAIL: TestBranchGuard_Latency (7.38s)
    pre_tool_branch_guard_integration_test.go:166: iteration 92: checkBranchState took 754.249958ms, ceiling 500ms (REQ-WBG-010 requires <= the per-OS ceiling per invocation)
```

It is a wall-clock latency assertion in `internal/hook` belonging to `SPEC-WORKTREE-BRANCH-GUARD-001`, failing on one iteration out of 100 by exceeding a 500 ms ceiling under machine load (the host was running several worktrees concurrently). Attribution: `git diff --stat 3419349c7..cea1f9ce5 -- internal/hook/` is empty — neither M4 commit touches that package, which is confined to `internal/cli` and `internal/config`. `go test -count=3 -run TestBranchGuard_Latency ./internal/hook/` → `ok … 13.928s` (3/3 in isolation), and the immediately-following full-suite run returned exit 0 with 112 `ok` including `internal/hook` `ok` at 73.124s (log: `.moai/state/verify/m4/full-suite-nudge2.txt`; the failing run is preserved at `.moai/state/verify/m4/full-suite-nudge.txt`). Recorded rather than dropped, and deliberately NOT presented as fixed — it is someone else's timing-sensitive ceiling, and a load-dependent latency assertion of that shape will fail again on a busy host.

**Gaps (M4 correction).**

- **`-count=20` bounds the flake, it does not prove its absence.** The race window is narrow; 20 iterations passing is strong evidence the join closed it, not a proof. What makes the fix trustworthy is the mechanism (the goroutine is joined at a point after its last write), not the iteration count.
- **The production regression test observes `updated_at`, not the write syscall.** It proves no record write landed after cancel returned. It does not prove the goroutine performs no other side effect after that point — it still reads the record and closes the session, both of which are unobserved by this test.
- **The gap this correction closes is in my own verification discipline, not only in the code.** A single green run was reported as evidence for a probabilistic property. Repetition (`-count`) is required for any test whose failure mode is a race, and no run in the original M4 batch had it.

### M5 — Registration, boundary, and hardcoding sweep

REQ-CX2-013, REQ-CX2-014, REQ-CX2-015. Evidence captured against the working tree on top of HEAD `cea1f9ce5` (branch `feat/SPEC-CODEX-PHASE2-001-run`) — the tree the two M5 commits land verbatim; the final batch below was re-run on the tree that is now HEAD `bae3e8616`. Every row names the command run and the output observed in this run, against this tree.

**What M5 delivers.** Registration of all four tools in `registerMoaiMCPTools` (`internal/cli/mcp_server.go`), the new `internal/cli/codex_registration_test.go` (the registration-shape check, the four independent per-tool existence assertions, the REQ-CX2-014 static boundary guard, and the §B mid-handshake edge case), and one sweep item in `internal/cli/codex_jobs.go` (`codexJobDirMode`). No handler was modified: M1-M4 wrote the behavior, M5 only declares it to the host.

**Three registration decisions worth stating.**

- *Tool names are quoted literals at the registration site, and that is the AC's own instruction, not a lapse.* `acceptance.md` §A names the "quoted-literal registration shape of `"codex_audit"` / `"codex_setup"`" and its existence check greps for it. The handlers carry the same names as constants (`codexTaskToolName`, `codexJob*ToolName`), so the string exists twice. The two cannot drift: `TestCodexJobTools_RegistrationShape` looks each registered tool up **by the constant**, so a literal that stopped matching its constant fails there rather than silently registering a tool nobody can reach.
- *The `false` read-only hints are stated explicitly, not inherited.* `mcp.NewTool` already seeds `ReadOnlyHint` to `false`, so `codex_task` and `codex_job_cancel` would carry the right value with no call at all. They are set anyway, because a hint that is correct only by inheriting a library default is indistinguishable from one nobody considered — and REQ-CX2-013 is a statement about what each tool DOES, not about what the constructor happened to leave behind.
- *The AC's warning about assertion shape was followed literally, and it was load-bearing in both directions.* A nil-check on `ReadOnlyHint` would have passed vacuously (the pointer is never nil), and a "schema is present" check could not fail at all (`Properties` is always a non-nil map). Both are therefore asserted by VALUE: `len(InputSchema.Properties) > 0` per tool, `Properties[job_id]` present on the three job tools, and `*Annotations.ReadOnlyHint == <expected>` per tool. The per-tool existence check is likewise four independent assertions rather than one counting grep, for the reason the AC gives: `grep -c` counts LINES, so four lines naming only `codex_task` would clear a `>= 4` gate with three tools missing.

**RED evidence (E8, captured before the registration existed).** `go test ./internal/cli/ -run 'TestCodexJobTools_RegistrationShape|TestCodexPhase2Tools_RegisteredIndependently|TestCodexPhase2_NoAskUserQuestion' -v`:

```
=== RUN   TestCodexJobTools_RegistrationShape
    codex_registration_test.go:53: tool "codex_task" is not registered (REQ-CX2-013)
    codex_registration_test.go:53: tool "codex_job_status" is not registered (REQ-CX2-013)
    codex_registration_test.go:53: tool "codex_job_result" is not registered (REQ-CX2-013)
    codex_registration_test.go:53: tool "codex_job_cancel" is not registered (REQ-CX2-013)
--- FAIL: TestCodexJobTools_RegistrationShape (0.00s)
=== RUN   TestCodexPhase2Tools_RegisteredIndependently
    codex_registration_test.go:101: MISSING codex_task — not registered as a quoted literal in mcp_server.go (REQ-CX2-013)
    codex_registration_test.go:101: MISSING codex_job_status — not registered as a quoted literal in mcp_server.go (REQ-CX2-013)
    codex_registration_test.go:101: MISSING codex_job_result — not registered as a quoted literal in mcp_server.go (REQ-CX2-013)
    codex_registration_test.go:101: MISSING codex_job_cancel — not registered as a quoted literal in mcp_server.go (REQ-CX2-013)
--- FAIL: TestCodexPhase2Tools_RegisteredIndependently (0.00s)
=== RUN   TestCodexPhase2_NoAskUserQuestion
--- PASS: TestCodexPhase2_NoAskUserQuestion (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.756s
```

The third test PASSED in the RED run, and that is reported rather than hidden: it is a static boundary guard over sources M1-M4 already wrote clean, so there was no failing state for it to start from. Only the two registration tests were genuinely RED. The §B edge-case test added afterwards likewise passed as written — it is coverage closure over an M3 path, not a TDD cycle, and no production change accompanies it.

| AC | Status | Command | Observed output |
|----|--------|---------|-----------------|
| AC-CX2-016 (REQ-CX2-013 registration) | PASS | `for t in codex_task codex_job_status codex_job_result codex_job_cancel; do grep -c "\"$t\"" internal/cli/mcp_server.go; done` | `1` / `1` / `1` / `1` — each name asserted independently, in its own invocation; exactly one quoted-literal occurrence per tool, and no `MISSING` line |
| AC-CX2-016 (REQ-CX2-013 schema + hints) | PASS | `go test ./internal/cli/ -run TestCodexJobTools_RegistrationShape -v` | `--- PASS` — all four tools present in `srv.ListTools()`; each declares a non-empty `InputSchema.Properties`; `codex_job_status` / `codex_job_result` / `codex_job_cancel` each declare the `job_id` property; `*Annotations.ReadOnlyHint` is `true` for `codex_job_status` and `codex_job_result` and `false` for `codex_task` and `codex_job_cancel` |
| AC-CX2-016 (REQ-CX2-014 boundary) | PASS | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/mcp_codex*.go internal/cli/codex_*.go \| grep -v '_test.go' \| grep -v '//'` | no output (exit 1). Backed by the static guard `go test -run TestCodexPhase2_NoAskUserQuestion` → `--- PASS` over `codex_task.go`, `codex_job_control.go`, `codex_jobs.go`, `mcp_codex.go` — the `TestNew_NoAskUserQuestion` shape `internal/cli/CLAUDE.md` requires, so the property is enforced on every `go test` rather than only when someone remembers the grep |
| AC-CX2-016 (REQ-CX2-015 placement) | PASS | `grep -n 'AllowWrite' internal/config/types.go` ; `grep -n 'AllowWrite' internal/config/defaults.go` | `586`/`592` (the doc comment + the typed field `AllowWrite bool \`yaml:"allow_write"\``) ; `684` (`AllowWrite: false`, the distributed default). Thresholds likewise: `DefaultCodexJobSummaryMaxLen` (285), `DefaultCodexJobCancelGrace` (293), `DefaultCodexJobCancelPoll` (299), all in `internal/config/defaults.go` |
| AC-CX2-016 (template surface) | PASS | `git status --porcelain internal/template/templates/` ; `grep -rn 'allow_write\|codex_job' internal/template/templates/` | empty ; no match (exit 1) — no distributed-surface expansion, and neither the config key nor a job identifier leaked into the template tree |
| AC-CX2-016 (cross-platform build) | PASS | `go build ./...` ; `GOOS=windows GOARCH=amd64 go build ./...` | `HOST_BUILD_EXIT=0` ; `WIN_BUILD_EXIT=0` |

**Hardcoding sweep (REQ-CX2-015), over M1-M4 as well as M5.** The sweep read every numeric literal, duration, env-var access, and path constant in the four files this SPEC added or touched.

- **Moved**: `os.MkdirAll(r.dir, 0o755)` in `codex_jobs.go` was the one inline literal belonging in a named constant — its 0600 file-mode sibling `codexJobFileMode` was already named two lines away. It is now `codexJobDirMode`, declared beside it so the pair is read and changed together.
- **Env-var names**: `grep -n 'os.Getenv\|Setenv\|LookupEnv'` over `codex_jobs.go`, `codex_task.go`, `codex_job_control.go`, `mcp_codex.go` → no output. This SPEC introduces no environment variable, so `internal/config/envkeys.go` needed no entry (REQ-CX2-015's env clause is vacuously satisfied, not skipped).
- **Left in place, deliberately**: the `3 * time.Second` close-wait, the 8 MB scanner buffer, and the 128-line channel in `mcp_codex.go` are pre-existing (PR #1430, `spec.md` §A.1) and outside what this SPEC introduces; touching them would be a drive-by change to PRESERVE-listed code. `filepath.Join(projectDir, ".moai", "state", …)` is the package-wide idiom (`harness_ledger.go`, `doctor.go`, `deps.go` all inline it) — matching the surrounding style is the convention here, and a lone constant would be the odd one out.

**§E verification batch** (final run, against the tree at HEAD `bae3e8616`).

- E2 cross-platform build — `go build ./...` → `HOST_BUILD_EXIT=0`; `GOOS=windows GOARCH=amd64 go build ./...` → `WIN_BUILD_EXIT=0`. `go vet ./internal/cli/ ./internal/config/` → `VET_EXIT=0`.
- E3 coverage — `go test -cover ./internal/cli/` → `ok github.com/modu-ai/moai-adk/internal/cli 200.203s coverage: 76.8% of statements` (log: `.moai/state/verify/m5/cover-final.txt`), against the M4 level of 76.8%. Coverage is at its pre-change level. (The pre-SPEC baseline recorded at M1 was 76.5%.)
- E4 subagent boundary — the grep above, plus the static guard test.
- E5 lint — `golangci-lint run --timeout=3m` → `0 issues.`
- Full suite — `go test ./...` → exit 0, 112 packages `ok`, zero `FAIL` / `--- FAIL` / `panic` lines (`internal/cli` `ok` at 211.102s). Run in full rather than affected-packages-only, so the cross-cutting template-mirror guards were exercised. Output at `.moai/state/verify/m5/full-suite-final.txt`; an earlier identical run before the §B edge-case test was added is at `.moai/state/verify/m5/full-suite.txt`, also exit 0 / 112 `ok` / 0 failures.
- Repetition — `go test -count=20 -run 'Codex|Job|Task|Cancel' ./internal/cli/` → `ok … 13.447s`. Run because the M4 correction established that a single green pass is not evidence for a test whose failure mode is a race; the post-cancel-write regression this SPEC fixed is exactly such a test, and it stays green under repetition on this tree.
- Race — `go test -race -run 'Codex|Job|Task|Cancel|Register' ./internal/cli/` → `ok … 3.781s`.
- Targeted — `go test ./internal/cli/... -run 'Codex|MCP|Job|Task|Cancel|Register'` → `ok` across all 17 `internal/cli` packages, zero failures.
- `gofmt -l` on the three touched files (`internal/cli/mcp_server.go`, `internal/cli/codex_jobs.go`, `internal/cli/codex_registration_test.go`) → no output. A repo-wide `gofmt -l .` still lists the same pre-existing ~33 files, none of them touched by this SPEC.
- Known local flakes — neither of the two unrelated timing-sensitive tests disclosed earlier fired on any run in this batch (not on either full-suite run, not on the coverage run, not on the `-count=20` run): `TestNavigatorEnrich_AtomicWriteBarrier` (`internal/cli`, an unrelated SPEC's goroutine-timing barrier, disclosed under M3) and `TestBranchGuard_Latency` (`internal/hook`, a load-dependent 500 ms latency ceiling belonging to `SPEC-WORKTREE-BRANCH-GUARD-001`, disclosed under the M4 correction). Two clean full-suite runs are evidence that neither fired here — not evidence that either is fixed; both remain load-dependent and neither was touched.

**Gaps (M5).**

- **No live codex session was executed, and this is the fifth milestone in a row saying so.** M5 adds no wire interaction of its own — it declares tools whose handlers M1-M4 wrote — so it neither closes nor worsens the inherited gap. But registration is what makes those unverified paths *reachable by an MCP host*: until this milestone the code existed and could not be called, and now it can. The gap is unchanged in substance and larger in consequence. It is restated in full in the run-phase closing report rather than left implicit here.

  **RESOLVED 2026-08-11** by § Live protocol verification below. A live `codex-cli 0.146.1` session was executed against the production session code path. Now OBSERVED: thread reuse (a second `turn/start` on an existing thread, item 1); `turn/started` arrival and its `turn.id` accepted as `turn/interrupt`'s `turnId`, with the turn ending `"status":"interrupted"` (item 2); the `{"type":"readOnly"}` `sandboxPolicy` envelope accepted on every turn, and policy stickiness reproduced against a fresh-thread baseline (item 3); `turn/started` emitted on the `review/start` path (item 4). Still NOT observed: `thread/resume` re-opening a thread in a new process (the path `resume_last` actually takes), the stall recovery at context expiry, background-job context lifetime through an MCP host, and any codex version or host other than 0.146.1 on darwin/arm64. One NEW finding — an unanswered `item/fileChange/requestApproval` stalls the turn — is recorded there and returned to the orchestrator rather than fixed under this SPEC.
- **Registration is verified against the in-process tool table, not against a host.** `srv.ListTools()` reads the server's own registry. The full stdio round-trip (`initialize` → `tools/list` → `tools/call`) is exercised for the pre-existing core surface by `TestMoaiMCPServer_ToolsListDeclaresSchema` and `TestMCPServer_StdioRoundTripSubprocess`, and the four new tools are included in that `tools/list` by construction — but no test calls one of the four new tools *through* the transport. The schema a host actually receives is therefore inferred from the registered `mcp.Tool` value, not observed on the wire.
- **The read-only hints are declarations, not enforcement.** `ReadOnlyHint` is advisory metadata an MCP host may use to decide what to auto-approve. Nothing in this server refuses a write because a tool declared itself read-only; `codex_job_status` is read-only because of what its handler does, and the annotation merely says so.
- **§B "two concurrent background tasks" is verified at the registry, not at the tool.** `TestCodexJobRegistry_ConcurrentJobsDoNotCollide` runs 8 concurrent creates and asserts 8 distinct ids and 8 files — the surface where a collision could actually occur. Two concurrent `handleCodexTask` calls are not tested, because the canned session double is a package-level singleton whose recorded-sends slice both calls would share: such a test would race on the test double rather than on production code, and would prove nothing about the registry that the 8-way test does not already prove.
- **§B "codex binary absent" is verified for `codex_task` only.** `TestCodexTask_FailOpenOnMissingCodex` covers the one new tool that invokes codex. The three job-control tools never spawn it — they read a record file, and cancel operates on a session this process already holds — so codex absence cannot reach them. This is a structural argument, not an observation; no test asserts it.

> **The first gap above is now CLOSED for four of its five wire-level claims.** See § Live protocol verification below for what was executed, what it confirmed, and what remains open.

### Live protocol verification — codex-cli 0.146.1

Observed 2026-08-11 against `codex-cli 0.146.1`, on the worktree at HEAD `4a059f8b1` (branch `feat/SPEC-CODEX-PHASE2-001-run`, working tree clean apart from the probe file this section documents). This is the FIRST live codex session executed for this SPEC. Every claim below names the command run and quotes the observed NDJSON verbatim.

**Method.** The probe drives the PRODUCTION session code path — `openCodexSession` → `runTurn` → `sendTurnInterrupt`, the same functions `codex_task` and the review gate call — rather than hand-rolling requests, so what is verified is what ships. The only inserted seam is `probeTap`, a pass-through `codexConn` that records each line in both directions and forwards it unchanged. Harness: `internal/cli/codex_live_protocol_probe_test.go`, opt-in via `MOAI_CODEX_LIVE_PROBE=1` and skipped by default, so it never runs on `go test ./...` and never spends quota unasked.

Transcripts were written to `.moai/state/verify/live-probe/*.ndjson`. **That path is gitignored** (`.gitignore:275` — `.moai/state/`), so the load-bearing lines are inlined here verbatim rather than cited by a path that will not resolve for a later reader.

**Binary resolution — a finding before the protocol was even reached.** `exec.LookPath("codex")` — which is what production uses (`codexLookPath`) — resolves on this host to `/Users/goos/.bun/bin/codex`, a shim whose vendored binary is missing:

```
Error: spawn /Users/goos/.bun/install/global/node_modules/@openai/codex/vendor/aarch64-apple-darwin/codex/codex ENOENT
```

The functional 0.146.1 install is `/Users/goos/.nvm/versions/node/v22.14.0/bin/codex`, reachable interactively only through a shell function — which is why `codex --version` succeeds at a prompt while `exec.LookPath` finds a broken binary. This is an environment fact, not an implementation defect (the fail-open path handles it correctly, and it is why the pre-existing `TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey` skips rather than runs during `go test ./...`). It is recorded because a reader who runs `codex --version` at a shell and concludes "the environment has a working codex for the Go code" would be drawing the wrong inference. The probe takes an explicit `MOAI_CODEX_LIVE_BIN` override.

#### Item 1 — a second `turn/start` on an existing `threadId`: CONFIRMED

M0 recorded this as "STRUCTURALLY SUPPORTED" — a schema-shape inference from `threadId` being required, which proves the method addresses an existing thread, not that a second turn is accepted after the first completes. It is now executed. Command: `MOAI_CODEX_LIVE_PROBE=1 MOAI_CODEX_LIVE_BIN=… go test ./internal/cli/ -run TestCodexLive_ThreadReuseAndTurnInterrupt -v`.

Turn 1 completed, then turn 2 was sent on the same thread and accepted:

```
--> {"id":4,"jsonrpc":"2.0","method":"turn/start","params":{"input":[{"text":"say ok","type":"text"}],"sandboxPolicy":{"type":"readOnly"},"threadId":"019fecb9-e537-7df0-aa4e-ccfa60eab514"}}
<-- {"id":4,"result":{"turn":{"id":"019fecba-2671-74f0-b67d-de48510731f6",…,"status":"inProgress",…}}}
<-- {"method":"turn/completed","params":{"threadId":"019fecb9-e537-7df0-aa4e-ccfa60eab514","turn":{"id":"019fecba-2671-74f0-b67d-de48510731f6",…,"status":"completed",…,"durationMs":1895}}}
```

Three turns ran on one `initialize` + `thread/start` handshake, each with a distinct `turn.id`. REQ-CX2-001's reusable handle and REQ-CX2-008's `resume_last` premise both hold at the wire, and AC-CX2-001's two-turns-on-one-handshake assertion is now backed by an executed session rather than a canned one. Reproduced on two independent runs.

#### Item 2 — `turn/started` arrival and `turnId` correspondence: CONFIRMED

The sharpest M0 inference: `Turn.id` → `turnId` was read across two schema files by naming and type agreement, and no `turn/interrupt` had ever been issued. Both halves are now observed.

`turn/started` arrives and carries `turn.id`:

```
<-- {"method":"turn/started","params":{"threadId":"019fecb9-e537-7df0-aa4e-ccfa60eab514","turn":{"id":"019fecba-2ea5-78c2-af90-96670bc587a8",…,"status":"inProgress","startedAt":1786383052,…}}}
```

That id was then sent back as `turn/interrupt`'s `turnId`, from a second goroutine while the turn's own goroutine was inside `awaitCodexTurnReview` — the production M4 shape — and was ACCEPTED:

```
--> {"id":6,"jsonrpc":"2.0","method":"turn/interrupt","params":{"threadId":"019fecb9-e537-7df0-aa4e-ccfa60eab514","turnId":"019fecba-2ea5-78c2-af90-96670bc587a8"}}
<-- {"id":6,"result":{}}
<-- {"method":"turn/completed","params":{…,"turn":{"id":"019fecba-2ea5-78c2-af90-96670bc587a8",…,"status":"interrupted",…,"durationMs":152}}}
```

The response is a bare success `result` with no error arm, and the turn ended with `"status":"interrupted"` 152 ms after the request — so the interrupt was honoured mid-flight, not merely accepted and ignored. This closes the M0 gap "no `turn/interrupt` call was actually issued, so its runtime behaviour (in particular whether it is honoured mid-tool-call) is unobserved", and it means M4's cancel path reaches its `turn/interrupt` branch in reality rather than only its grace-expiry branch, which is the one every unit test lands on. Reproduced on two independent runs.

#### Item 3 — `sandboxPolicy`: envelope CONFIRMED, stickiness CONFIRMED

**(a) The envelope M3 sends is accepted.** `{"type":"readOnly"}` — the internally-tagged object, not a bare string — rode every turn above and no turn was rejected. The `-32600` failure mode that taught the `target` shape did not occur.

**(b) Stickiness is real, and REQ-CX2-007's per-turn transmission is load-bearing.** This required disambiguation, because a turn that writes after OMITTING the field has two possible explanations that are not the same claim: inheritance from an earlier turn, or an omitted field simply defaulting to write-capable. Both were tested.

`TestCodexLive_SandboxPolicyStickiness` — turn 1 sent `workspaceWrite` and wrote; turn 2 on the same thread sent NO `sandboxPolicy` and also wrote:

```
--> {"id":3,…,"method":"turn/start","params":{"input":[{"text":"create a file named first.txt containing the word hi",…}],"sandboxPolicy":{"type":"workspaceWrite"},"threadId":"019fecbb-59b9-7761-805b-24729d97f414"}}
<-- {"method":"turn/completed",…,"text":"Created [first.txt](…/first.txt) containing `hi`.",…,"status":"completed",…}
<-- {"method":"turn/completed",…,"text":"Created [second.txt](…/second.txt) containing `hi`.",…,"status":"completed",…}
```

Zero `requestApproval` lines in that transcript — the policy-omitting turn wrote unchallenged.

`TestCodexLive_OmittedSandboxPolicyBaseline` removes the predecessor: a FIRST turn on a FRESH thread, sending no `sandboxPolicy`. The session default is read-only, and the write attempt was blocked pending approval instead of succeeding:

```
<-- {"id":2,"result":{"thread":{…},"approvalPolicy":"on-request","sandbox":{"type":"readOnly","networkAccess":false},"activePermissionProfile":{"id":":read-only"},…}}
<-- {"method":"thread/status/changed","params":{…,"status":{"type":"active","activeFlags":["waitingOnApproval"]}}}
<-- {"method":"item/fileChange/requestApproval","id":0,"params":{"threadId":"019fecc4-9437-7161-90e0-5afeaff7f843","turnId":"019fecc4-a648-7650-a60c-d1cfe102011f","itemId":"exec-bf72b916-…","startedAtMs":1786383749290,"reason":null,"grantRoot":null}}
```

Omission alone is therefore NOT write-capable — the session default is `readOnly`. The stickiness probe's turn 2 wrote because it **inherited** turn 1's `workspaceWrite`. This is exactly the hazard REQ-CX2-007 describes, now observed rather than read: a thread reused under `resume_last` WOULD inherit a write-enabled policy from an earlier opted-in turn, and sending `readOnly` explicitly on every turn is what closes it. The M0 schema reading ("for this turn and subsequent turns") was correct, and the amendment it forced was warranted.

#### Item 4 — `turn/started` on a `review/start` turn: CONFIRMED

M2's open question, and the premise behind M4's empty-`turn_id` degradation branch. The probe sends `review/start` and tears the session down the moment the question is answered, so no full review turn was billed (the test ran in 6.66 s):

```
--> {"id":3,"jsonrpc":"2.0","method":"review/start","params":{"target":{"type":"uncommittedChanges"},"threadId":"019fecbb-12bb-7bd1-87d0-68e37c4e0725"}}
<-- {"method":"turn/started","params":{"threadId":"019fecbb-12bb-7bd1-87d0-68e37c4e0725","turn":{"id":"019fecbb-266a-75a0-8030-009db340cd67",…,"status":"inProgress","startedAt":1786383115,…}}}
```

`review/start` DOES emit `turn/started`, so a review-path job records a real `turn_id` and its cancel does NOT degrade to termination-only. The degradation branch M4 implemented remains correct code for the case where `turn/started` is missed (a cancel arriving before the notification), but its premise is not the steady-state behaviour of the review path.

#### NEW FINDING — an unanswered server→client request stalls the turn

Not one of the four items, and not a contradiction of any REQ or AC: nothing in the SPEC claims approval requests are handled. It is an unhandled protocol case that only a live session could surface, and it is reachable through the envelope `codex_task` actually sends.

`item/fileChange/requestApproval` is a server→client **request** — it carries an `"id"` and expects a response. The session driver never answers it: `awaitCodexTurnReview` dispatches on `msg.Method` and has cases only for `turn/started`, `item/completed`, and `turn/completed`, so the line is read, matched to the thread, and dropped. Codex holds the turn at `activeFlags:["waitingOnApproval"]` and neither side moves.

`TestCodexLive_ExplicitReadOnlyApprovalStall` exercises the exact non-write envelope `codex_task` builds (`codexSandboxPolicy(false)` → `{"type":"readOnly"}`) with a prompt that provokes a file write. The turn did not return within 120 s, and the approval request is the last line of the transcript:

```
<-- {"method":"thread/status/changed","params":{…,"status":{"type":"active","activeFlags":["waitingOnApproval"]}}}
<-- {"method":"item/fileChange/requestApproval","id":0,"params":{"threadId":"019fecc8-4b47-73e1-8cdd-5eefa94b8ee5","turnId":"019fecc8-5925-7db0-83c0-837a551592ce","itemId":"exec-21c320c3-…","startedAtMs":1786384004385,"reason":null,"grantRoot":null}}
```

Reachability is high rather than exotic: EVERY `codex_task` turn without a write grant runs `readOnly`, and any such turn where the model decides to edit a file lands here. The stall is bounded only by the context — `exec.CommandContext` kills the subprocess at ctx expiry, which closes stdout and unblocks the reader — and `handleCodexTask` sets no deadline of its own, so the bound is whatever the MCP caller imposes. No fix is applied here: answering a server→client request is new behaviour no requirement covers, so it is returned to the orchestrator as a proposed amendment rather than written in under this SPEC's scope.

#### What this probe did NOT verify

- **The `-count=60` / `-race` batches below cover the canned paths only.** Nothing about the live findings is asserted by a test that runs in CI; the probe is opt-in by design, so a protocol regression in a future codex release will not be caught automatically.
- **The stall's outer bound is reasoned, not measured.** That `exec.CommandContext` unblocks the reader at ctx expiry follows from the code; the probe cut its own wait at 120 s and never ran a turn to context expiry, so the recovery path was not observed.
- **Background-job context lifetime is untouched.** `runCodexBackgroundJob` receives the MCP request context and `exec.CommandContext` is keyed to it, so whether a background job survives its tool call returning depends on when the host cancels that context. No live background job was run through an MCP host, and nothing here establishes the answer either way.
- **One host, one version.** Everything above is `codex-cli 0.146.1` on darwin/arm64 with this account's config (which carries its own MCP servers and hooks — visible as noise in the transcripts). `codex_app_server_protocol.v2.schemas.json` remains unexamined, as M0 recorded.
- **`resume_last` was not exercised end-to-end.** Item 1 proves a second turn on a held thread; it does not exercise `thread/resume` re-opening a thread from codex's on-disk store in a NEW process, which is the path `resume_last` actually takes.

**§E verification batch** (after adding the probe file, against the tree at HEAD `4a059f8b1`).

- Full suite — `go test ./...` → `FULL_SUITE_EXIT=0`, 112 packages `ok`, zero `FAIL` / `--- FAIL` / `panic` lines. Neither known flake (`TestNavigatorEnrich_AtomicWriteBarrier`, `TestBranchGuard_Latency`) fired.
- Repetition — `go test -count=60 -run 'Codex|Job|Task|Cancel' ./internal/cli/` → `ok … 47.492s`.
- Race — `go test -race -run 'Codex|Job|Task|Cancel' ./internal/cli/` → `ok … 3.965s`.
- Cross-platform build — `go build ./...` → `HOST_BUILD_EXIT=0`; `GOOS=windows GOARCH=amd64 go build ./...` → `WIN_BUILD_EXIT=0`; `go vet ./internal/cli/` → `VET_EXIT=0`.
- Lint — `golangci-lint run --timeout=3m` → `0 issues.`
- Default-skip proof — `go test ./internal/cli/ -run TestCodexLive -v` with no opt-in env → all five probe tests `--- SKIP` with `MOAI_CODEX_LIVE_PROBE != 1`.

### M6 — Protocol liveness (amendment, v0.6.0)

REQ-CX2-016 / AC-CX2-017 and REQ-CX2-017 / AC-CX2-018. Implemented against the tree at HEAD `be9eb7e40` (branch `feat/SPEC-CODEX-PHASE2-001-run`, working tree clean at start). This milestone closes the gap the live probe found (§ Live protocol verification, NEW FINDING) and nothing else: no existing requirement was amended, and the M1-M5 suites were re-run unmodified as the regression check.

#### The response wire shape — read from the schema, not inferred

`plan.md` §F AP-9 forbids inferring the answer's shape from the request transcript, because a wrong field name reads as *no decision* while a wrong value could read as **approval**. The shape was therefore read from the generated protocol schema, the same source M0 used:

```
$ codex --version
codex-cli 0.146.1
$ codex app-server generate-json-schema --out <dir>
```

`FileChangeRequestApprovalResponse.json`, verbatim and complete in its load-bearing part:

```json
{
  "title": "FileChangeRequestApprovalResponse",
  "type": "object",
  "required": ["decision"],
  "properties": { "decision": { "$ref": "#/definitions/FileChangeApprovalDecision" } },
  "definitions": {
    "FileChangeApprovalDecision": {
      "oneOf": [
        { "description": "User approved the file changes.", "enum": ["accept"] },
        { "description": "User approved the file changes and future changes to the same files should run without prompting.", "enum": ["acceptForSession"] },
        { "description": "User denied the file changes. The agent will continue the turn.", "enum": ["decline"] },
        { "description": "User denied the file changes. The turn will also be immediately interrupted.", "enum": ["cancel"] }
      ]
    }
  }
}
```

The schema settles the shape outright: the response is `{"decision": "<variant>"}`, and it enumerates exactly two denying variants. **`decline` was chosen over `cancel`** because the schema's own descriptions distinguish them — `decline` denies and lets the agent continue the turn, `cancel` denies and interrupts it. Denial that still lets the turn finish is what returns the caller a turn output stating what could not be done; `cancel` would return an interrupted turn. Both deny; the weaker consequence is the one that preserves the caller's result.

Two sibling requests share the shape and are recorded because they bound the claim: `CommandExecutionRequestApprovalResponse` and `PermissionsRequestApprovalResponse` also require `decision`, with `decline` / `cancel` among their variants. Neither is given a recognized case here — see the Gaps below.

The binary used for the schema is the functional 0.146.1 install (`/Users/goos/.nvm/versions/node/v22.14.0/bin/codex`), not the broken PATH shim the probe documented.

#### What was implemented

- **`internal/cli/mcp_codex.go`** — `awaitCodexTurnReview` gained one branch: a decoded line carrying BOTH a method and a non-null id is a server→client request and is answered via `answerCodexClientRequest`. `item/fileChange/requestApproval` gets `{"decision":"decline"}`; anything else gets a JSON-RPC error arm with code `-32601` (Method not found). Two small writers were added (`writeCodexResponse` / `writeCodexErrorResponse`) rather than reusing `writeCodexRequest`, which builds a request envelope. The branch sits BEFORE the thread filter on purpose: an unrecognized request need not carry a `threadId`, and a request dropped for want of one stalls the turn exactly as surely as one dropped for want of a case.
- **`internal/config/defaults.go`** — `DefaultCodexTaskTimeout = 600 * time.Second`, a `var` so a test can shorten it, and deliberately distinct in both name and value from `DefaultCodexReviewGateTimeout` (900 s).
- **`internal/cli/codex_task.go`** — `runCodexTaskTurn` races the turn against that bound and is used by BOTH the foreground handler and `runCodexBackgroundJob`. A context deadline alone would not have sufficed: the driver blocks in `conn.recv()`, which returns when the connection ends, not when a context expires.
- **The denial is unconditional and no grant path exists** (`plan.md` §F AP-7). `workflow.codex.task.allow_write` remains the sole write opt-in: a turn that opted in already runs `workspaceWrite`, so an approval request that reaches this driver is by construction one that must not be granted.

#### RED evidence

The tests were written first. Verbatim first run, before any production change (the constant the assertions name did not exist):

```
$ go test ./internal/cli/ -run 'TestCodexTask_Denies|TestCodexTask_AnswersUnrecognized|TestCodexReviewGate_Denies|TestCodexTask_ForegroundTurnBounded|TestCodexTask_BackgroundTurnTimeout|TestCodexTaskTimeout_IsDistinct'
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/codex_protocol_liveness_test.go:268:17: undefined: codexApprovalDecisionDecline
internal/cli/codex_protocol_liveness_test.go:270:14: undefined: codexApprovalDecisionDecline
internal/cli/codex_protocol_liveness_test.go:339:61: undefined: codexApprovalDecisionDecline
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
```

A compile failure is a weak RED — it proves the constant was absent, not that the withholding device works. So the device was **falsified directly**: with the implementation complete and passing, the single dispatch line `answerCodexClientRequest(conn, msg)` was replaced by a bare `continue` and the arms re-run. Verbatim:

```
--- FAIL: TestCodexTask_DeniesFileChangeApprovalRequest (5.00s)
    codex_protocol_liveness_test.go:259: the turn did not return cleanly: status=failed result=map[... error:codex_task turn timed out after 5s (the bound codex_task imposes on its own turns); the turn was abandoned and the session torn down ...]
--- FAIL: TestCodexTask_AnswersUnrecognizedClientRequest (5.00s)
    codex_protocol_liveness_test.go:298: the turn did not return cleanly: status=failed result=map[... error:codex_task turn timed out after 5s ...]
```

Both arms fail when the answer is dropped, and they fail by exhausting the bound — so the falsification demonstrates BOTH criteria at once: AC-CX2-017's withholding server really does withhold, and AC-CX2-018's bound really does fire on a turn that stops advancing. The line was restored immediately after.

That falsification also **found a defect in the test this milestone authored**: `TestCodexReviewGate_DeniesFileChangeApprovalRequest` originally called `runCodexReviewRPC` with `context.Background()` and no bound of its own. REQ-CX2-017 binds `codex_task` only, so with the answer removed that test HUNG rather than failing (the run was killed at 126 s). It now drives the call in a goroutine under a 10 s ceiling and closes the conn to release the reader — a regression test that hangs instead of failing is the one outcome such a test must not have.

#### A production data race the milestone introduced, and fixed

`go test -race` failed on the first attempt with a genuine race this change created:

```
WARNING: DATA RACE
Read at 0x00c000799750 by goroutine 25:
  ...cli.handleCodexTask() codex_task.go:240
Previous write at 0x00c000799750 by goroutine 26:
  ...cli.(*codexSessionHandle).noteTurnStarted() mcp_codex.go:607
  ...cli.awaitCodexTurnReview() mcp_codex.go:887
  ...cli.runCodexTaskTurn.func1() codex_task.go:97
```

Before M6, `runTurn` was synchronous, so `session.turnID` was always written before it was read. Bounding the turn made the read concurrent with an abandoned turn goroutine still writing it. Fixed at the source rather than papered over in the test: `noteTurnStarted` now writes under the handle's existing mutex (firing the observer OUTSIDE the lock, since it writes a job record), a `currentTurnID()` accessor is the only outside read, and `setTurnStartedObserver` gives the observer field the same discipline. `go test -race` and `-race -count=20` are clean afterwards.

This is the third time on this branch that a race-shaped defect survived something weaker than a repeated run (M2, M4, now M6) — and the first time the repeated-run discipline caught one *at the milestone that introduced it* rather than a milestone later.

#### AC verdicts

| AC | REQ | Verdict | Evidence |
|---|---|---|---|
| AC-CX2-017 main arm — approval answered with a denial | REQ-CX2-016 | **PASS** | `TestCodexTask_DeniesFileChangeApprovalRequest` — asserts the transmitted line carrying the request's id decodes to `result.decision == "decline"` and that the turn returned; the canned server withholds `turn/completed` until an id-matched line is observed |
| AC-CX2-017 unknown-method arm | REQ-CX2-016 | **PASS** | `TestCodexTask_AnswersUnrecognizedClientRequest` — the request carries NO `threadId`, proving the answer precedes the thread filter; asserts a JSON-RPC `error` arm with code and message |
| AC-CX2-018 foreground arm | REQ-CX2-017 | **PASS** | `TestCodexTask_ForegroundTurnBoundedByOwnDeadline` — `context.Background()` (no caller deadline), bound overridden to 300 ms, two-sided assertion (≥ 300 ms, ≤ 30 s), status `failed`, error names the timeout |
| AC-CX2-018 background arm | REQ-CX2-017 | **PASS** | `TestCodexTask_BackgroundTurnTimeoutReachesTerminalStatus` — job record reaches `failed` after ≥ the bound, error names the timeout |
| AC-CX2-018 distinctness arm | REQ-CX2-017 | **PASS** | `grep -n 'DefaultCodexReviewGateTimeout' internal/cli/codex_task.go` → no output (exit 1). Also held as `TestCodexTaskTimeout_IsDistinctFromReviewGateBudget`, so a later edit that collapses the two budgets fails in CI rather than only in a one-off grep |
| Blast radius — the review gate answers too, contract unchanged (C7) | REQ-CX2-016 | **PASS** | `TestCodexReviewGate_DeniesFileChangeApprovalRequest` — `runCodexReviewRPC` keeps its `(ReviewOutput, error)` shape and returns a verdict; the same denial is transmitted |

**§E verification batch** (against the tree at HEAD `be9eb7e40` plus the M6 change, before commit).

- Targeted — `go test -count=1 ./internal/cli/ -run 'Codex|MCP|Job|Task|Cancel|Approval|Timeout'` → `ok … 1.880s`.
- Repetition — `go test -count=60 -run 'Codex|Job|Task|Cancel|Approval|Timeout' ./internal/cli/` → `ok … 206.745s`.
- Race — `go test -race -run 'Codex|Job|Task|Cancel|Approval|Timeout' ./internal/cli/` → `ok … 3.451s`; and repeated, `-race -count=20` → `ok … 26.948s`. A single race run is not evidence for a timing defect, which is why both are recorded.
- Full suite — `go test ./...` → 112 packages `ok`, zero `FAIL` / `panic` lines. Neither known unrelated flake (`TestNavigatorEnrich_AtomicWriteBarrier`, `TestBranchGuard_Latency`) fired.
- Cross-platform build — `go build ./...` → `HOST_BUILD_EXIT=0`; `GOOS=windows GOARCH=amd64 go build ./...` → `WIN_BUILD_EXIT=0`; `go vet ./internal/cli/ ./internal/config/` → `VET_EXIT=0`.
- Coverage — `go test -cover ./internal/cli/` → `ok … 201.245s coverage: 76.9% of statements`, against the M5 level of 76.8%. Above its pre-change level.
- Lint — `golangci-lint run --timeout=3m` → `0 issues.`
- Boundary — `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/mcp_codex*.go internal/cli/codex_*.go | grep -v '_test.go' | grep -v '//'` → no matches.
- Template surface — `git status --porcelain internal/template/templates/` → empty.
- Formatting — `gofmt -l` over the four touched files → empty. (`gofmt -l .` across the whole repo lists ~140 pre-existing files untouched by this milestone; that condition predates M6 and is not addressed here.)
- Default-skip proof — `go test ./internal/cli/ -run 'TestCodexLive' -v` → all five live probes `--- SKIP`. The opt-in probe still never runs on `go test ./...` and still spends no quota unasked.

#### Live confirmation of the fix — codex-cli 0.146.1, HEAD `e5295dee6`

Added after the milestone's own §E batch, and attributed to the ORCHESTRATOR's run rather than this agent's: the orchestrator re-ran the probe that originally observed the stall, against the committed fix, and returned the transcript below. It is recorded here because it closes the residual risk this milestone had left standing (see the correction in the next subsection).

```
$ MOAI_CODEX_LIVE_PROBE=1 MOAI_CODEX_LIVE_BIN=/Users/goos/.nvm/versions/node/v22.14.0/bin/codex \
    go test ./internal/cli/ -run 'TestCodexLive_ExplicitReadOnlyApprovalStall' -v -timeout 6m
...
<-- {"method":"item/completed","params":{"item":{"type":"agentMessage",...,"text":"I couldn’t create `blocked.txt` because write approval was denied.","phase":"final_answer"},"threadId":"019fed4d-ae34-7b80-868f-57be97a290fa","turnId":"019fed4d-b19f-71e2-aeb7-4f5d23fcb8bb",...}}
<-- {"method":"turn/completed","params":{"threadId":"019fed4d-ae34-7b80-868f-57be97a290fa","turn":{"id":"019fed4d-b19f-71e2-aeb7-4f5d23fcb8bb",...,"status":"completed","error":null,"startedAt":1786392719,"completedAt":1786392730,"durationMs":10605}}}
--- PASS: TestCodexLive_ExplicitReadOnlyApprovalStall (12.40s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	13.775s
```

The binary is the functional 0.146.1 install; `MOAI_CODEX_LIVE_BIN` is set because `exec.LookPath` on this host still resolves to the broken bun shim the original probe documented.

Three separate claims land in that one transcript, and it is worth naming them separately because they are not the same claim:

1. **The envelope was ACCEPTED on the wire.** No `-32600` / `-32602` rejection came back for the response line. The schema reading above is now corroborated by the server's own behavior rather than standing alone.
2. **Codex acted on it as a denial.** The write did not happen, and the agent's final message says why: `"I couldn’t create blocked.txt because write approval was denied."` — so `decline` denied rather than merely being tolerated.
3. **The turn ENDED.** `turn/completed` with `"status":"completed"`, `"error":null`, `durationMs: 10605`.

The before/after contrast is the whole point of the milestone, and both numbers come from the same test against the same envelope: the pre-fix run of this probe did not return **within 120 s** (§ Live protocol verification, NEW FINDING — the approval request was the last line of that transcript); the post-fix run completed in **10.6 s**. The stall is gone, and `decline` is the reason.

#### What M6 did NOT verify

- ~~**No live session was run against the fix.** The whole milestone is verified against canned conns. The stall was OBSERVED live; the cure is not. A live re-run of `TestCodexLive_ExplicitReadOnlyApprovalStall` would be the direct confirmation that codex accepts `{"decision":"decline"}` and proceeds — it was not performed, so **codex's acceptance of the response envelope is read from the schema, not observed on the wire**. This is the same evidence class M0 operated in, and it is the single largest residual risk of this milestone.~~ **CLOSED** by § Live confirmation of the fix above — the orchestrator ran exactly that test at HEAD `e5295dee6`, and it returned `turn/completed` / `status:"completed"` in 10.6 s with the write denied. Codex's acceptance of the response envelope is now OBSERVED on the wire, not read from the schema. The original wording is struck rather than deleted, matching how the M2 and M4 corrections were recorded: the gap was real when it was written, and a reader tracing how it closed should be able to see what it said.
- **What that live run does NOT extend to.** One live run of one test closes one inference. It says nothing about the other gaps below, every one of which remains open exactly as written — in particular it does **not** establish anything about `item/commandExecution/requestApproval` or `item/permissions/requestApproval`, which the probe never exercised.
- **Only `item/fileChange/requestApproval` has a recognized case.** `item/commandExecution/requestApproval` and `item/permissions/requestApproval` exist in the schema and share the `{"decision":...}` shape, but they fall to the `-32601` error arm here. That unblocks the turn (which is what REQ-CX2-016 requires of an unrecognized method) but it is a coarser answer than the `decline` those schemas would accept. Handling them was left out as scope the amendment does not name; whether codex treats a `-32601` on those requests as gracefully as a `decline` is **not established**.
- **The `-32601` code is a choice, not an observation.** The requirement asks for "a JSON-RPC error response" and cites codex's own `-32600` rejection as precedent. `-32601` (Method not found) was chosen as the semantically accurate code. No probe confirms which codes codex tolerates in a client→server error response.
- **The abandoned turn goroutine's exit is reasoned, not asserted.** On timeout the reader is left blocked in `recv()` and is released by the caller's session tear-down, which both call sites perform on every exit path. No test asserts the goroutine actually exited (no leak detector is in use here); the argument is structural.
- **The 600 s value is a judgment, not a measurement.** Nothing was measured to choose it. It is distinct from the gate's 900 s, which is what the requirement demands; whether it is the right length for real task turns is unknown.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-11
run_commit_sha: ac37c4aea          # SHA re-attributed at sync: PR #1440 squash-merged the run phase onto main as ac37c4aea, so this is the commit that carries the implementation on main. The pre-squash branch value was bae3e8616 (the last IMPLEMENTATION commit on feat/SPEC-CODEX-PHASE2-001-run), which is NOT an ancestor of main — see the §E.2 header note for how to resolve it.
run_status: complete
ac_pass_count: 16
ac_fail_count: 0
preserve_list_post_run_count: 5
l44_pre_commit_fetch: not-applicable   # run-phase commits stay local on feat/SPEC-CODEX-PHASE2-001-run; no push performed this phase
l44_post_push_fetch: not-applicable    # no push performed this phase
new_warnings_or_lints_introduced: 0    # golangci-lint run --timeout=3m → "0 issues."
cross_platform_build:
  host: 0                              # go build ./... → exit 0
  windows_amd64: 0                     # GOOS=windows GOARCH=amd64 go build ./... → exit 0
total_run_phase_files: 9
m1_to_mN_commit_strategy: one commit per milestone (M1 146baf37f, M2 628422d8c, M3 3419349c7, M4 2c851678e), plus two follow-ups — cea1f9ce5 (the M4 post-cancel-write correction found by orchestrator re-verification) and bae3e8616 (the §B mid-handshake coverage closure). M5 itself is bf7183101. No amend, no force-push, no squash.
```

Notes on the fields above, so a later reader does not have to re-derive them:

- `ac_pass_count: 16` covers AC-CX2-001..016, every one MUST. AC-CX2-007 was recorded PASS-WITH-DEBT at M2 (its "*the tool* returns a structured error result" clause needed `codex_task`, which did not exist yet) and closed outright at M3; it is counted once, as PASS.
- `preserve_list_post_run_count: 5` is the `plan.md` §A.1 PRESERVE list, all five entries intact: `runCodexReviewRPC`'s `(ReviewOutput, error)` contract for both existing callers, the `inconclusiveReview` / `VerdictInconclusive` fail-open semantics, `readCodexReviewGateEnabled`'s nested key path and fail-closed truth table, `TestReviewGateReaders_AgreeWithConfigLoader` + `TestMCPAudit_NoDirectFrontmatterRead`, and every file outside `internal/cli/` and `internal/config/`.
- `total_run_phase_files: 9` — `internal/cli/mcp_codex.go`, `internal/cli/mcp_server.go`, `internal/cli/codex_jobs.go`, `internal/cli/codex_task.go`, `internal/cli/codex_job_control.go`, `internal/cli/codex_jobs_test.go`, `internal/cli/codex_task_test.go`, `internal/cli/codex_job_control_test.go`, `internal/cli/codex_registration_test.go`; plus `internal/config/types.go` and `internal/config/defaults.go` for the typed key and its default. `internal/template/templates/` is untouched.
- The `l44_*` fetch fields are `not-applicable` rather than `0`: the run phase commits to a feature branch and pushes nothing, so there was no push boundary at which to fetch. Recording `0` would assert a check that was never run.
- **M6 (v0.6.0 amendment) restates three of the fields above, and the block is again left as measured rather than overwritten.** With M6 landed: `ac_pass_count` reads **18** (AC-CX2-001..018, AC-CX2-017 and AC-CX2-018 added by the amendment, still `ac_fail_count: 0`); `total_run_phase_files` reads **11** on the branch as a whole, the eleventh being `internal/cli/codex_protocol_liveness_test.go` (M6 modified three files already counted — `mcp_codex.go`, `codex_task.go`, `internal/config/defaults.go` — and added no other); and `new_warnings_or_lints_introduced` stays `0` (`golangci-lint run --timeout=3m` → `0 issues.` on the M6 tree). `preserve_list_post_run_count` stays **5**: M6 touches `awaitCodexTurnReview`, which the review gate shares, but `runCodexReviewRPC`'s `(ReviewOutput, error)` contract, the `inconclusiveReview` / `VerdictInconclusive` fail-open path, `readCodexReviewGateEnabled`, the two named guard tests, and the file boundary are all intact — and the gate's caller-supplied 900 s bound at `codex_review_gate.go:87` was deliberately left where it is (REQ-CX2-017 binds `codex_task` only).
- The block above describes the run phase as it stood at `bae3e8616` and is left as measured. The later live-protocol probe (§E.2 § Live protocol verification, HEAD `4a059f8b1`) adds a tenth file, `internal/cli/codex_live_protocol_probe_test.go` — test-only, opt-in, no production change — so `total_run_phase_files` reads 10 on the branch as a whole. The figure is noted here rather than overwritten above, so the original measurement stays attributable to the tree it was taken against.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-11
sync_commit_sha: e6f7e72b1               # the sync commit carrying the 3-phase close; backfilled in the immediately following commit, per spec-frontmatter-schema.md § SHA placeholder backfill exemption (a commit cannot cite its own hash)
sync_status: complete
run_phase_merged_as: ac37c4aea           # PR #1440, squash-merged to main 2026-08-10; verified `git merge-base --is-ancestor ac37c4aea origin/main` → exit 0
b12_self_test_a: pass                    # grep -c 'SPEC-CODEX-PHASE2-001' CHANGELOG.md → 0 (no prior entry; emission proceeds)
b12_self_test_b: pass                    # acceptance.md distinct AC count → 18 (AC-CX2-001..018); CHANGELOG entry states 18
b12_self_test_c: pass                    # every path named in the CHANGELOG entry verified present via `ls`
changelog_entry_position: "[Unreleased] → Added, first bullet (ahead of the SPEC-HARNESS-LEARNING-EVO-001/002 entry)"
frontmatter_status_transitions:
  spec_md: in-progress → completed        # merged in-progress → implemented → completed on this single sync commit
  plan_md: n/a                            # plan.md carries no YAML frontmatter (Tier M artifact authored without one)
  acceptance_md: n/a                      # acceptance.md carries no YAML frontmatter
  updated_field: 2026-08-11 (spec.md; already current, unchanged)
canary_compliance_check: n/a              # this SPEC defines no forward-looking policy that its own sync would test
docs_surface_changed: none                # see the note below
```

Notes on the fields above:

- **SHA attribution.** Every SHA this SPEC recorded during the run phase was orphaned by the squash merge. The two verified against `origin/main` at sync time: `ac37c4aea` → ancestor (exit 0); `bae3e8616` and `e5295dee6` → NOT ancestors (exit 1). §E.3's `run_commit_sha` was re-attributed to `ac37c4aea` with its pre-squash value retained inline; §E.2's per-milestone SHAs were deliberately left as measured, with a resolution note added at the section head rather than an overwrite.
- **No docs-site or README change.** The four tools and the `workflow.codex.task.allow_write` key are a new user-facing surface, but the codex MCP tool surface has no existing documentation page in any locale — `codex_audit` and `codex_setup` shipped without one, and `docs-site/content/*/advanced/autonomous-loops.md` mentions `workflow.codex.review_gate` only in passing as a config-key sibling. Opening a first codex-tool documentation section is a scoped docs task in its own right, and it would carry the 4-locale parity obligation across ko/en/ja/zh. It was not started in this pass rather than started and left partial. The CHANGELOG entry is the user-facing record for this release.
- **Write gate ships off.** `workflow.codex.task.allow_write` defaults to `false` in `internal/config/defaults.go` and was deliberately NOT added to `internal/template/templates/` (spec.md §C Out of Scope), so a distributed user receives no template key to flip accidentally.
- **What sync did NOT verify.** No live codex session was run during the sync phase. The four protocol behaviours confirmed live against `codex-cli 0.146.1` during the run phase (thread reuse; `turn.id` → `turnId` with a real `turn/interrupt`; `sandboxPolicy` acceptance and stickiness; `turn/started` on `review/start`) plus the M6 fix confirmation stand as recorded in §E.2; everything else in this SPEC's wire-level reasoning remains a reading of the generated schema, and the §E.2 "did NOT verify" lists are unchanged by sync.
