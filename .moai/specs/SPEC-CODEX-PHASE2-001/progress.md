# Progress — SPEC-CODEX-PHASE2-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- tier: M
- artifacts: spec.md + plan.md + acceptance.md (Tier M set)
- spec version: v0.3.0 (plan-audit iteration 2 revision — D1/D2/D3/D4 closed, S2 taken; see `spec.md` HISTORY)
- run-phase entry gate: M0 must close before M1 starts. M0's probe record lands in §E.2 below, under a `### M0 probe — codex-cli <pinned version>` heading, per `plan.md` §D M0 items 2-3.
- open blockers: none. The two `plan.md` §D M0 design forks (background job execution model; write-mode opt-in surface) were resolved by user decision on 2026-08-10 and are recorded as decisions in `plan.md` §D M0 — (i) in-process goroutine, and the `workflow.codex.task.allow_write` config key. spec.md v0.3.0 carries the propagated consistency amendments plus the plan-audit iteration 2 revision. Plan-audit PASSED at 0.84 against the Tier M threshold 0.80 (iteration 2 of 3; report at `.moai/reports/plan-audit/SPEC-CODEX-PHASE2-001-review-2.md`). The SPEC is clear for run-phase, subject to Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

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

**Gaps (M4 correction).**

- **`-count=20` bounds the flake, it does not prove its absence.** The race window is narrow; 20 iterations passing is strong evidence the join closed it, not a proof. What makes the fix trustworthy is the mechanism (the goroutine is joined at a point after its last write), not the iteration count.
- **The production regression test observes `updated_at`, not the write syscall.** It proves no record write landed after cancel returned. It does not prove the goroutine performs no other side effect after that point — it still reads the record and closes the session, both of which are unobserved by this test.
- **The gap this correction closes is in my own verification discipline, not only in the code.** A single green run was reported as evidence for a probabilistic property. Repetition (`-count`) is required for any test whose failure mode is a race, and no run in the original M4 batch had it.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
