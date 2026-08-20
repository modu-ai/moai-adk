# t79 Deep Audit Report — sync-auditor (t79-deep-audit), verbatim relay

> Reviewer note: relayed verbatim from the sync-auditor subagent (2026-08-17).
> The orchestrator independently re-verified F1's line-level facts (see
> review-verdict.md §Evidence) before adopting the FAIL verdict.

Review mode: read-only static analysis (git show / blob diffs / mcp-go source). No builds or tests re-run (machine-load discipline; orchestrator verification cited as-is).

## (a) Findings (by severity)

### F1 [Major — MERGE BLOCKING] `DefaultGLMTaskTimeout` (600s) is never the effective bound — the real bound is the shared audit client's 120s

- `internal/cli/glm_task.go:143-144` — sync arm calls `callGLMTask(ctx, …)` with NO timeout wrapper. The constant is not applied on this path at all.
- `internal/cli/glm_task.go:200-203` — background arm wraps `context.WithTimeout(ctx, config.DefaultGLMTaskTimeout)`.
- `internal/cli/mcp_glm.go:79,90` — `glmAuditHTTPTimeout = 120 * time.Second`; `glmHTTPClient = &http.Client{Timeout: glmAuditHTTPTimeout}`. `http.Client.Timeout` covers the ENTIRE request (connect + headers + body read), so background is effectively `min(120s, 600s) = 120s`.
- `internal/config/defaults.go:352-362` — doc claims it "bounds ONE glm_task call — sync or background" and is "deliberately distinct from the audit path's timeout". Both statements are contradicted by the implementation: the effective bound in BOTH forms is exactly that audit timeout.
- `git grep DefaultGLMTaskTimeout 9865e87ed` → exactly one reader (glm_task.go:200). Dead in sync, shadowed in background.
- The in-code comment at glm_task.go:196-198 ("the HTTP client's 120s ceiling bounds one connection, not the job as a whole") is wrong on both counts — Client.Timeout includes body read, and a GLM job is exactly one HTTP call, so 120s IS the job bound.

Failure scenario: super-advisor delegates an open-ended task (default max_tokens = DefaultGLMTaskMaxTokens = 8192); z.ai streams 8k tokens for 150s; at 120s the client timeout fires → sync returns `{status:"failed", error:"z.ai request failed: … Client.Timeout exceeded…"}` / background job records failed. The designed 600s is unreachable. codex_task applies its 600s genuinely in BOTH forms (runCodexTaskTurn), so "mirrors DefaultCodexTaskTimeout" is false — the mirror is silently weaker.

Fix: task-family client (Timeout 0 or 600s) + `WithTimeout(DefaultGLMTaskTimeout)` on the sync arm too + a test shortening DefaultGLMTaskTimeout below 120s proving the task constant governs. **Blocks merge: YES** — headline claim in the commit message and defaults.go is not delivered; dead constant + incorrect doc comments.

### F2 [Minor — non-blocking] cancel-vs-complete RMW race not excluded across registry instances (inherited verbatim from codex)

- `internal/cli/glm_job_control.go:179` — cancel uses `currentGLMJobRegistry().update(...)` (a FRESH instance per handler call).
- `internal/cli/glm_task.go:178` + handler — the background goroutine holds the DIFFERENT instance created in handleGLMTask.
- `internal/cli/glm_jobs.go:180-183` — `mu sync.Mutex` is per-instance. Two instances share the directory but not the mutex. `writeFileAtomic`'s rename prevents torn reads, not lost updates.

Scenario: between cancel's `loadGLMJobFor` (reads running) and its `update` (writes cancelled), the goroutine's `updateUnlessCancelled` completes read(running)→write(completed). Then cancel's unconditional `update` overwrites completed with cancelled (Output survives under a cancelled status) — or the reverse ordering flips the record to completed after the tool returned cancelled. Window = the load→update span (µs–ms); outcome = one-time status inconsistency, no crash or data loss. Codex has the IDENTICAL structure (`currentCodexJobRegistry()` vs the captured registry), so this is a faithfully mirrored pre-existing limitation, not a regression. `TestGLMJobCancel_NoRecordWriteAfterCancel` proves only the guarded direction. Fix later for both families: process-wide registry singleton or per-id lock. **Blocks: NO.**

### F3 [Advisory] `job_id` caller-supplied, no format validation before `pathFor()`

`glm_jobs.go:193-195` (`filepath.Join(r.dir, id+".json")`), `glm_job_control.go:212-225`. `glm_job_status` with a traversal id reads any parseable JSON under the project and returns its GLMJobRecord-shaped fields. The write direction IS blocked — cancel refuses ids with no `glmLiveJobs` entry. The `//nolint:gosec // registry-owned: <generated id>` comment is inaccurate on the job-control path (id is caller input). Mirrors codex exactly. Defense-in-depth: reject ids not matching `^job-…-[0-9a-f]{8}$`. **Blocks: NO.**

### F4 [Advisory] status/result report an orphaned record from a previous server lifetime as `running`, no staleness hint

`glm_job_control.go:235-266`. A caller polling after a restart loops on "running" + pending note forever; only the cancel refusal names the staleness. Mirrors codex; non-survival is documented in the glm_task description. A live-map cross-check in status/result would close it. **Blocks: NO.**

### F5 [Advisory] record `Output`/`Error` are model-produced and not passed through `codexJobSummary`

A model echoing a credential from the prompt lands verbatim in the 0600-mode record. `RequestSummary` IS redacted-then-truncated (glm_jobs.go:210; codex regexes: bearer / k=v / prefix tokens — verified). Mirrors codex. **Blocks: NO.**

### F6 [Advisory] background create→running failure leaves a permanent `queued` orphan

`glm_task.go:172-177` — live entry deleted + cancel + toolErr is correct, but the record stays non-terminal (cancel will refuse it). Requires a disk failure. Same shape as codex. **Blocks: NO.**

### F7 [Advisory — test gaps] three concrete gaps

(1) No test shortens DefaultGLMTaskTimeout — the test that would have caught F1. (2) The grace-expired branch (glmJobCancelGraceNote, glm_job_control.go:227-229) is never deterministically exercised. (3) No coverage for traversal ids or stale-record status reads. The other ~1,053 lines assert real behavior (value-based hint checks, updated_at as no-write witness, two-sided redaction assertion, live-map join) — not vacuous.

### F8 [Advisory — cosmetics] drive-by comment reindent in `mcp_server.go:613-620` (gofmt of an unrelated @MX:DEBT block, harmless); `glm_job_control.go:241` references `config.DefaultCodexJobCancelPoll` directly — documented as shared, the one live cross-family coupling beyond the derived grace const. **Blocks: NO.**

## Verified claims (adversarially checked, survived)

1. Credentials: key only in `x-api-key` header (glm_task.go:252); record has no credential/endpoint field by construction (glm_jobs.go:121-138); summary redacted-then-truncated; error strings carry status codes and paths only. No key reaches disk/stdout on any path traced. TRUE.
2. Atomic writes: `writeFileAtomic` = same-dir CreateTemp + Chmod + rename (settings.go:124-155). Malformed JSON → structured error result, no crash (tested). TRUE.
3. Cancel semantics: cancel status recorded BEFORE context revocation; goroutine terminal writes go through `updateUnlessCancelled`; revocation genuinely aborts the in-flight request (request built with NewRequestWithContext; blocking-doer test models a real client). TRUE (residual race = F2).
4. Stale refusal: `glmLiveJobs` sync.Map, entries deleted on every exit path; `!live → refuse` (glm_job_control.go:164-171). TRUE (status-side gap = F4).
5. Fail-open completeness: missing key (short-circuits pre-HTTP, test asserts no HTTP), 4xx/5xx, transport error, malformed/no-content — all structured failed results, never Go errors, never IsError. IsError arms = missing prompt + unwritable state dir only, matching the documented contract. TRUE.
6. Goroutine hygiene: deferred Delete+cancel on all paths; no lock held across the HTTP call; `Content[0]` len-guarded. TRUE.
7. Catalog: counted 21 entries, exactly 6 WriteCapable (goal_arm, verify_snapshot, codex_task, codex_job_cancel, glm_task, glm_job_cancel); capacity hints updated 17→21 at both sites; registration ReadOnlyHint ↔ WriteCapable consistent for all four tools. TRUE.
8a. `glm_job_cancel` WC:true mirrors codex_job_cancel's pre-existing true. TRUE.
8b. super-advisor mirror drift is PRE-EXISTING: source-vs-template drift at parent 4004a2a06 and at 9865e87ed is byte-identical content (glm-5.2/glm-5.3 line + design-authority bullet); the only delta is hunk line numbers shifted by the +11 lines added identically to both blobs. Added hunks equivalent. TRUE.
9. Model resolution: `resolveGLMTaskModel` → `resolveGLMModelForAgent("super-advisor")` — shared SSOT param-refactor of the audit resolver (mcp_glm.go); no hardcoded model; fallback = config.DefaultGLMHigh, pinned by test. TRUE.
10. Background goroutines survive handler return under stdio: mcp-go v0.57.0 `ServeStdio` creates ONE connection-scoped ctx (stdio.go:847, cancelled only on SIGTERM/SIGINT) — no per-request cancellation. TRUE for the stdio deployment (an HTTP transport would break it; codex shares the pattern).
11. Korean progress messages match both codex_task and glm_audit conventions — consistent, not a finding.
12. 600s timeout: FALSE (F1).

## (b) Verdict

**FAIL** — one blocking Major (F1). The new code's structure, hygiene, and test quality are solid, and 11 of 12 claim groups survived adversarial checking, but the headline constant `DefaultGLMTaskTimeout` (600s) is dead in sync and shadowed by the 120s audit client in background — contradicting the commit message, defaults.go, and the codex-mirror claim. The fix is small (task-family client + sync-arm wrap + a shortening test). F2–F8 are non-blocking (mostly symmetric limitations inherited from codex).

## (c) 5-Section Close

**Claim**: statically verified the security/integrity/symmetry claims of the glm_task family. Credential handling, atomicity, cancel semantics, stale refusal, fail-open, catalog counts, mirror drift, and model SSOT verified TRUE; the 600s timeout claim refuted (F1).

**Evidence**: `git show 9865e87ed --stat` (16 files, +1,886/−31); full read of /tmp/t79_full.diff (2,222 lines); `git show 9865e87ed:internal/cli/{codex_task,codex_jobs,codex_job_control,mcp_glm}.go` counterpart comparison — key observations: `glmHTTPClient = &http.Client{Timeout: glmAuditHTTPTimeout}` (mcp_glm.go:90), `glmAuditHTTPTimeout = 120s` (:79); codex sync applies WithTimeout via runCodexTaskTurn, GLM sync does not (glm_task.go:143-144 vs :200); `git grep DefaultGLMTaskTimeout` → one reader; catalog.go full read + hand count (21/6); 4-blob super-advisor diff (parent/commit × source/template) with 3-stage diff comparison; helper reads (settings.go writeFileAtomic, mcp_server.go toolJSON/toolErr, mcp_progress.go nil-token no-op, mcp_glm_test.go seam restore); mcp-go v0.57.0 module cache (stdio.go:847 connection-scoped ctx, server.go:1893 handleToolCall).

**Baseline-attribution**: all observations obtained this review by reading commit 9865e87ed blobs (parent 4004a2a06) and the mcp-go v0.57.0 module cache directly. No tests re-run by auditor; orchestrator's deterministic verification (build OK; `go test ./internal/cli/ -run 'GLM|TestMoaiMCPServer|TestAC_C'` ok 1.357s; `./internal/mcp/` ok 0.250s; `-race` ok 3.491s; rule mirror byte-identical) quoted as-is.

**Gaps**: (1) no dynamic verification (mission constraint); (2) frequency with which real z.ai generations exceed 120s (F1's real-world hit rate) unmeasured; (3) mcp-go ctx lifetime verified for stdio only — other transports unexercised (server is stdio-deployed today); (4) F2 race window size is reasoned, not measured; (5) full-suite status rests on the orchestrator's re-run, not the auditor's.

**Residual-risk**: if F1 is fixed by raising/removing the client Timeout alone, the sync arm regains NO bound — the WithTimeout wrap must land together. Background jobs die silently on server exit mid-generation (by design); post-restart orphans poll as "running" forever (F4). Two concurrent `moai mcp-server` processes on one checkout have no cross-process mutex — rename prevents torn reads only (shared with codex).
