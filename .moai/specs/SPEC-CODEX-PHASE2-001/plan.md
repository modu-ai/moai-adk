# Plan — SPEC-CODEX-PHASE2-001

> Implementation plan for the Codex Phase 2 tool surface. Milestones are ordered by **decision-reversibility**: the protocol probe and the two open design forks come first, because everything downstream is shaped by their answers; the mechanical registration work is last.

## §A. Context

- **Baseline tree**: `origin/main` at `e55b48def`. Note the local checkout was 14 commits behind at authoring time — the delivered-state analysis in `spec.md` §A was made against `origin/main`, not the local `HEAD`.
- **Primary file**: `internal/cli/mcp_codex.go` (730 lines on `origin/main`).
- **Registration site**: `internal/cli/mcp_server.go` `registerMoaiMCPTools` (L105-242); codex tools currently at L191-208.
- **Second consumer of the client**: `internal/cli/codex_review_gate.go` `HandleCodexReviewGate` (L66-110) — it calls `runCodexReviewRPC` with a 900 s context and must keep working unchanged.
- **Existing test surface** (extend, do not replace): `mcp_codex_test.go`, `codex_review_rpc_test.go`, `codex_rpc_error_test.go`, `codex_review_gate_test.go`, `codex_review_gate_live_test.go`, `mcp_audit_test.go`.

### §A.1 PRESERVE list

- `runCodexReviewRPC`'s existing behavior for its two current callers (review-gate ALLOW/BLOCK, `codex_audit`).
- `inconclusiveReview` / `VerdictInconclusive` fail-open semantics.
- `readCodexReviewGateEnabled`'s nested `workflow.codex.review_gate.enabled` key path and its fail-closed truth table (`mcp_codex.go:708-729`) — an earlier revision read this at the top level and the toggle could never read true; do not "simplify" it back.
- `TestReviewGateReaders_AgreeWithConfigLoader` and `TestMCPAudit_NoDirectFrontmatterRead`.
- Every file outside `internal/cli/` and `internal/config/`.

## §B. Known issues carried into this plan

- **B1 — the predecessor's deferral list is stale.** Anyone re-reading `SPEC-MOAI-MCP-SERVER-001/spec.md:47` will see "native Go JSON-RPC client" listed as deferred. It is substantially delivered. Work from `spec.md` §A of *this* SPEC.
- **B2 — the `model` parameter is a live bug, not a missing feature.** `codex_audit` advertises `model` and silently drops it (`buildCodexReviewParams`, `mcp_codex.go:433`). M1 fixes it; the fix changes what is sent to codex, so the review-gate tests are in the blast radius.
- **B3 — the SSOT guard passes vacuously.** `TestMCPAudit_NoDirectFrontmatterRead` greps `mcp_codex.go` for frontmatter reads and finds none because nothing is resolved at all. After M1 the guard must still pass *and* a positive test must assert the resolver is actually called.
- **B4 — undocumented protocol.** `turn/interrupt` and any write-mode flag are assumed, not observed. M0 exists to convert them into observations or to force a scope change.
- **B5 — cross-platform.** Process termination in cancellation is platform-sensitive; `GOOS=windows GOARCH=amd64 go build ./...` must pass.

## §C. Pre-flight

```bash
git fetch origin && git rev-parse --short origin/main
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
go test ./internal/cli/... 2>&1 | tail -20
golangci-lint run --timeout=2m 2>&1 | tail -5
codex --version || echo "codex absent — live probes skip, fail-open paths still testable"
```

## §D. Milestones

### M0 — Protocol probe (lead; most likely to change)

The two design forks that formerly sat in M0 are **closed** — both were resolved by user decision on 2026-08-10 and are recorded below as decisions. M0's remaining work is the protocol probe alone. Everything below M0 still assumes the answers the probe produces; do not start M1 before M0 closes.

1. Probe `codex app-server` against the pinned version **`codex-cli 0.146.1`** (the version observed by `codex --version` at authoring time; if the installed version differs at run-phase, record the observed `codex --version` output and pin that instead) for: (a) whether a `turn/interrupt`-equivalent method exists and its params; (b) whether a second `turn/start` on an existing `threadId` is accepted after a turn completes; (c) whether the turn request carries a model/effort field and under what name; (d) whether a write/sandbox mode is expressible per-turn.
2. **Probe record — destination and form.** Each of (a)-(d) is recorded in `progress.md` **§E.2 Run-phase Evidence**, under a `### M0 probe — codex-cli <pinned version>` heading, as either (i) the request and response NDJSON lines verbatim, or (ii) an explicit recorded absence naming the command run and what was observed instead. No `research.md` is created — the Tier M artifact set is spec + plan + acceptance, and `progress.md` §E.2 is the run-phase evidence surface the manager-develop §E attribution triple already writes to. The pinned version accompanies every entry.
3. **M0 closure criterion.** M0 closes when all four of (a)-(d) carry a recorded observation or a recorded absence in `progress.md` §E.2 **and**, for every item recorded absent, the corresponding SPEC amendment has landed. M1 does not start before M0 closes.
4. **Degrade paths (an unobserved method does not become a requirement).** If (a) is absent, REQ-CX2-011 degrades to process-termination-only and `acceptance.md` **AC-CX2-014** — the AC bound to REQ-CX2-011 per the `acceptance.md` §D table, *not* AC-CX2-011, which verifies REQ-CX2-008 `resume_last` — is amended before run-phase to drop its interrupt-send clause, so it asserts only that `codex_job_cancel` terminates the process this server spawned within the bounded grace window, that the call returns within a bounded time, and that the job's recorded status becomes `cancelled`. If (b) is absent, REQ-CX2-001/REQ-CX2-008 and AC-CX2-001/AC-CX2-011 are amended before M1 starts (AC-CX2-002, REQ-CX2-001's other AC per the §D table, is a regression guard on the pre-existing codex and review-gate tests and survives whether or not thread reuse is possible — its omission here is deliberate, not the id arithmetic warned against above). If (c) is absent, REQ-CX2-002's "carry the resolved value into the params actually transmitted" clause and AC-CX2-003 are amended. If (d) is absent, REQ-CX2-007's writing-mode arm and AC-CX2-010 are amended to the non-writing path alone. Each degrade path is checked against the `acceptance.md` §D traceability table rather than against REQ-N ↔ AC-N id arithmetic, which does not hold past AC-CX2-005.

**Decision — background job execution model: (i) in-process goroutine.** Resolved 2026-08-10 by user decision. A background `codex_task` job runs as a goroutine inside the long-lived `moai mcp-server` process that holds the codex subprocess. This makes the pid a job records always one *this* process spawned in *this* lifetime, so REQ-CX2-012's ownership check is a same-process comparison rather than a reattachment problem, and REQ-CX2-011's cancellation is a direct send on the session the goroutine still holds. Accepted consequence: every in-flight job is lost when the server exits — job durability across an mcp-server restart is explicitly not delivered here, and no resume path is specified by this SPEC. Rejected: (ii) a detached codex subprocess surviving an mcp-server restart — it buys durability at the cost of pid reattachment, which would make REQ-CX2-012's ownership check load-bearing against pids this process never spawned; it is a later upgrade if job durability is actually wanted, not a requirement of this SPEC.

**Decision — write-mode opt-in surface: the config key.** Resolved 2026-08-10 by user decision. REQ-CX2-007's write opt-in is a new nested config key `workflow.codex.task.allow_write`, following the existing `workflow.codex.review_gate.enabled` shape (`mcp_codex.go:708`; typed at `internal/config/types.go` `CodexConfig` / `CodexReviewGateConfig`), read fail-closed so a missing file, a parse error, or an absent block all yield false. The distributed default is `false`. It is inspectable by `codex_setup`, which already surfaces the sibling gate's state. Rejected: an environment variable in `envkeys.go` — it would not be visible to `codex_setup`, and it would add a second opt-in shape alongside a pattern the tree already carries.

### M1 — Reusable session handle + model/effort SSOT wiring

REQ-CX2-001, REQ-CX2-002. Split `runCodexReviewRPC` so the handshake (`initialize` → `thread/start` → `threadId`) is reachable as a reusable session, with `runCodexReviewRPC` retained as a thin caller so both existing consumers are untouched. Wire `template.ResolveAgentModelEffort` into the codex path and stop dropping the resolved value in `buildCodexReviewParams`. Add a positive test asserting the resolver is called (B3).

### M2 — Job registry

REQ-CX2-003, REQ-CX2-004, REQ-CX2-005. Per-job JSON files under `.moai/state/codex-jobs/`, atomic write per transition, structured error on an unwritable state directory, no secrets in the record. Follows the `.moai/state/audit-multi/<session>.json` precedent (`internal/cli/mcp_convergence.go:73`).

Per the M0 in-process decision, the record's process reference is the pid of the codex subprocess this server spawned in the current process lifetime. No reattachment metadata is recorded, and the record carries nothing intended to let a later server lifetime adopt the job — a record found in a non-terminal status after a restart is stale by construction, not resumable.

### M3 — `codex_task`

REQ-CX2-006, REQ-CX2-007, REQ-CX2-008. Foreground and background forms, the write opt-in gate, and `resume_last` thread reuse on top of M1's session handle. The background form is the in-process goroutine of the M0 decision; the write gate reads `workflow.codex.task.allow_write` fail-closed (distributed default `false`).

**M3 deliverable — `codex_setup` inspectability.** M3 also adds an `allow_write` field to `handleCodexSetup`'s result map (`mcp_codex.go:631`, which today emits `installed` / `auth_provider` / `enable_review_gate` / `node_bridge` plus `binary` / `version`), reporting the same fail-closed read as the gate itself. This is a named M3 deliverable, not incidental prose: the write-mode fork chose the config key over an env var precisely *because* the key is visible to `codex_setup` (see the M0 decision's rejected alternative), so the inspectability is load-bearing to that decision rather than a convenience. It is verified by AC-CX2-010's `codex_setup` arm.

### M4 — Job control tools

REQ-CX2-009, REQ-CX2-010, REQ-CX2-011, REQ-CX2-012. Status, result, and cancel; cancellation sends the M0-confirmed interrupt method, then terminates the process this server spawned for that job, and only that process. Under the M0 in-process decision that pid is always one spawned in the current process lifetime, so the REQ-CX2-012 ownership check is a same-lifetime comparison — a record naming a pid outside that set (a leftover from a previous server lifetime) is refused rather than signalled. No pid reattachment logic is written.

### M5 — Registration, boundary, and hardcoding sweep

REQ-CX2-013, REQ-CX2-014, REQ-CX2-015. Tool registration with JSON Schema and read-only hints, the AskUserQuestion boundary grep, constants placement, and confirmation that `internal/template/templates/` is untouched.

**Critical path**: M0 → M1 → M2 → M3 → M4 → M5. M2 may proceed in parallel with M1: the execution-model decision is already recorded, so the record shape is fixed and does not depend on the session refactor.

## §E. Self-verification

Per milestone, report the §E attribution triple — the command, its verbatim output, and the HEAD SHA the evidence was captured against — per `.claude/rules/moai/development/manager-develop-prompt-template.md` § Section E.

```bash
go test ./internal/cli/... -run 'Codex|MCP'
go test ./... && go build ./... && GOOS=windows GOARCH=amd64 go build ./...
go test -cover ./internal/cli/...
grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/mcp_codex*.go internal/cli/codex_*.go | grep -v '_test.go' | grep -v '//'
git status --porcelain internal/template/templates/    # expect empty
golangci-lint run --timeout=2m
```

## §F. Anti-patterns

- **AP-1 — re-implementing the transport.** The NDJSON conn, the id matching, the error arm, and the turn loop exist (`spec.md` §A.1). Extend them; do not write a second client.
- **AP-2 — changing `runCodexReviewRPC`'s contract for its existing callers.** The Stop gate's fail-open depends on the `(ReviewOutput, error)` pair where the output is always usable.
- **AP-3 — specifying `codex_transfer` opportunistically.** It is excluded with a stated re-entry condition (`spec.md` § Out of Scope — `codex_transfer`). Adding it mid-run reopens a decision that was made deliberately.
- **AP-4 — turning cancellation into a process sweep.** Kill only the pid this server spawned for that job (REQ-CX2-012). A pattern-matched `pkill codex` would kill a developer's interactive session.
- **AP-5 — making the write gate default-on for convenience during development.** The distributed default is false; a local opt-in belongs in local config, not in the code default.
- **AP-6 — quietly fixing `synthesizeReviewOutput` while nearby.** Findings extraction is out of scope; a verdict-parsing change here would ride into the review gate untested against its own criteria.

## §G. Cross-references

- `spec.md` §A — the verified delivered/remaining split (read before touching `mcp_codex.go`).
- `acceptance.md` — AC-CX2-001..016.
- `internal/cli/CLAUDE.md` — CLI module conventions (subagent boundary, exit codes, absolute paths).
