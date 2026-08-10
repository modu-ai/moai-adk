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

### M0 — Protocol probe and design-fork resolution (lead; most likely to change)

Everything below M0 assumes answers M0 produces. Do not start M1 before M0 closes.

1. Probe `codex app-server` against a pinned codex-cli version for: (a) whether a `turn/interrupt`-equivalent method exists and its params; (b) whether a second `turn/start` on an existing `threadId` is accepted after a turn completes; (c) whether the turn request carries a model/effort field and under what name; (d) whether a write/sandbox mode is expressible per-turn.
2. Record each observed request/response verbatim. An unobserved method does not become a requirement — if (a) is absent, REQ-CX2-011 degrades to process-termination-only and `acceptance.md` AC-CX2-011 is amended before run-phase.
3. Resolve the two open forks below.

**[NEEDS CLARIFICATION: background job execution model]** — a background job can be (i) a goroutine inside the long-lived `moai mcp-server` process holding the codex subprocess, or (ii) a detached codex subprocess that survives an mcp-server restart. (i) makes cancellation and pid ownership trivial but loses every in-flight job when the server exits; (ii) survives restarts but needs pid reattachment and makes REQ-CX2-012's ownership check the load-bearing safety property. The record shape in REQ-CX2-003 differs between them. Recommendation: (i) — it matches the single-process lifetime the server already has and keeps the first delivery honest; (ii) is a later upgrade if job durability is actually wanted.

**[NEEDS CLARIFICATION: write-mode opt-in surface]** — REQ-CX2-007 needs a named opt-in. Candidates: a new `workflow.codex.task.allow_write` key following the existing `workflow.codex.review_gate.enabled` shape (`mcp_codex.go:708`), or an environment variable in `envkeys.go`. Recommendation: the config key, because it mirrors a pattern already in the tree and is inspectable by `codex_setup`. Either way the distributed default is false.

### M1 — Reusable session handle + model/effort SSOT wiring

REQ-CX2-001, REQ-CX2-002. Split `runCodexReviewRPC` so the handshake (`initialize` → `thread/start` → `threadId`) is reachable as a reusable session, with `runCodexReviewRPC` retained as a thin caller so both existing consumers are untouched. Wire `template.ResolveAgentModelEffort` into the codex path and stop dropping the resolved value in `buildCodexReviewParams`. Add a positive test asserting the resolver is called (B3).

### M2 — Job registry

REQ-CX2-003, REQ-CX2-004, REQ-CX2-005. Per-job JSON files under `.moai/state/codex-jobs/`, atomic write per transition, structured error on an unwritable state directory, no secrets in the record. Follows the `.moai/state/audit-multi/<session>.json` precedent (`internal/cli/mcp_convergence.go:73`).

### M3 — `codex_task`

REQ-CX2-006, REQ-CX2-007, REQ-CX2-008. Foreground and background forms, the write opt-in gate, and `resume_last` thread reuse on top of M1's session handle.

### M4 — Job control tools

REQ-CX2-009, REQ-CX2-010, REQ-CX2-011, REQ-CX2-012. Status, result, and cancel; cancellation sends the M0-confirmed interrupt method, then terminates the process this server spawned for that job, and only that process.

### M5 — Registration, boundary, and hardcoding sweep

REQ-CX2-013, REQ-CX2-014, REQ-CX2-015. Tool registration with JSON Schema and read-only hints, the AskUserQuestion boundary grep, constants placement, and confirmation that `internal/template/templates/` is untouched.

**Critical path**: M0 → M1 → M2 → M3 → M4 → M5. M2 may proceed in parallel with M1 once M0's execution-model fork is resolved, since the record shape does not depend on the session refactor.

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
