# progress.md — SPEC-NAVIGATOR-SYNC-005 (BAS Epic M3 — Falconer Fix)

> Run-phase + sync-phase evidence accumulates here. Plan-phase: §E.1 is populated;
> §E.2–§E.4 are placeholder headings (populated by manager-develop / manager-docs in
> later phases). The literal `§E.2` / `§E.3` / `§E.4` heading tokens are parser-load-bearing
> for `internal/spec/era.go` era classification — do NOT rename them.

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: pending-audit-iter2
- plan_complete_at: 2026-08-12
- audit_iter1_verdict: FAIL 0.74 (3 blocking defects D1/D2/D3 — architecture praised as SOUND)
- audit_iter1_fixes: D1 diff-scope UNION semantics unified across 3 artifacts; D2 REQ-NS5-013 + AC-NS5-013 scope-conformance guard; D3 REQ-NS5-008(c4) approval_token + AC-NS5-008d + DBT-2 apply idempotence
- tier: L
- artifact_count: 5 (spec.md + plan.md + acceptance.md + design.md + research.md)
- req_count: 13 (REQ-NS5-001 through REQ-NS5-013)
- ac_count: 20 (AC-NS5-001a through AC-NS5-013, per acceptance.md §B traceability matrix)
- depends_on: [SPEC-NAVIGATOR-SYNC-002 (M1 Detect, completed), SPEC-NAVIGATOR-SYNC-004 (M2 Route, completed)]
- key_decision: split-architecture (Go diff-scope engine + orchestrator-mediated manager-develop AI draft + AskUserQuestion approval gate + approval_token) — defended in design.md §A
- success_metric: Fix 자동화율 ≥ 50% (design report §9 line 443)

## §E.2 Run-phase Evidence

### M3.1 — Split-architecture scaffold + diff-scope contract

- AC-NS5-003 (diff-scope identifies stale subtrees, UNION semantics): PASS — `go test ./internal/navigator/fix/ -run TestComputeScope -v -count=1` exit 0
- AC-NS5-004a/004b (draft-request schema + provenance + idempotence): PASS — `go test ./internal/navigator/fix/ -run TestRun -v -count=1` exit 0

### M3.2 — Draft-request emission + Hidden CLI subcommand

- AC-NS5-001a/001b (on-demand CLI, no PostToolUse): PASS — `go test ./internal/cli/ -run TestNavigatorFix -v -count=1` exit 0
- AC-NS5-011 (Hidden cobra subcommand): PASS — `go test ./internal/cli/ -run TestNavigatorFixHidden -v -count=1` exit 0

### M3.3 — AI draft delegation wiring (orchestrator → manager-develop)

- AC-NS5-007a (zero LLM-client imports): PASS — `go test ./internal/navigator/fix/ -run TestFixNoLLMClient -count=1` exit 0 (the split-architecture grep guard)
- AC-NS5-007b (stdout handoff signal): PASS — `go test ./internal/navigator/fix/ -run TestRun_StdoutHandoff -v -count=1` exit 0

### M3.4 — Approval surface (scope-conformance + patch preview)

**Claim**: AC-NS5-013 (draft scope-conformance: out-of-scope subtree excluded) + AC-NS5-008b Go-half (*.patch preview generation) PASS.

**Evidence**:
- AC-NS5-013 — `go test ./internal/navigator/fix/ -run TestDraftScopeConformance -v -count=1`:
  - `--- PASS: TestDraftScopeConformance` — 3 draft subtree IDs (2 in-scope + 1 out-of-scope) → inScope=[in-scope-A,in-scope-B], excluded=[out-of-scope-X]
  - `--- PASS: TestLogScopeExclusion_Warning` — `.moai/logs/navigator-sync.log` carries the excluded ID + the diff_scope representation + the `scope-conformance` marker
  - `--- PASS: TestConformDraftToScope_AllInScope` + `TestConformDraftToScope_EmptyInputs` — degenerate cases
- AC-NS5-008b Go-half — `go test ./internal/navigator/fix/ -run 'TestUnifiedDiff|TestGeneratePreviews' -v -count=1`:
  - `--- PASS: TestUnifiedDiff_Structure` — `--- a/`, `+++ b/`, `@@` hunk header, `-`/`+`/context lines present
  - `--- PASS: TestUnifiedDiff_IdenticalReturnsEmpty` — no-op case
  - `--- PASS: TestGeneratePreviews_WritesPatchFiles` — 2 `<doc-surface>.patch` files written with valid unified-diff structure
  - `--- PASS: TestGeneratePreviews_MissingLiveDocSkipped` — fail-open skip on missing live doc

**Baseline-attribution**: `(this run, this tree)` against HEAD pre-commit on branch `worktree-bas-m2-route`. Full package `go test ./internal/navigator/fix/ -count=1` → `ok ... 5.169s`. Coverage `go test -cover ./internal/navigator/fix/` → `coverage: 88.0%` (≥85% target). Cross-platform `GOOS=windows GOARCH=amd64 go build ./...` → exit 0. Race `go test -race ./internal/navigator/fix/` → `ok ... 6.086s`. Lint `golangci-lint run ./internal/navigator/fix/... ./internal/cli/...` → `0 issues.`

**Gaps**: AC-NS5-008b orchestrator-side (the AskUserQuestion 4-option gate rendering) is NOT in this package (orchestrator-integration level per acceptance.md) — verified absent via subagent-boundary grep: `grep -rn 'AskUserQuestion' internal/navigator/fix/*.go internal/cli/navigator_fix.go | grep -v _test.go | grep -v "// "` → 0 matches.

**Residual-risk**: the LCS-based UnifiedDiff is O(n*m) — acceptable for the doc surfaces M3 diffs (hundreds to low-thousands of lines); a future Myers optimization is out of scope for M3.4.

### M3.5 — Apply-on-approval + non-overlap enforcement

**Claim**: AC-NS5-008a (no live-doc mutation without approval) + AC-NS5-008c (apply atomic-rename + applied.json ledger) + AC-NS5-008d (approval_token refusal — both branches) + AC-NS5-005a/005b (consumer-only) + AC-NS5-006 (non-overlap) PASS.

**Evidence**:
- AC-NS5-008d — `go test ./internal/navigator/fix/ -run TestApplyTokenRefusal -v -count=1`:
  - `--- PASS: TestApplyTokenRefusal_NoToken` — no approval.json → Apply returns ErrApprovalTokenMissing (wrapped), live doc SHAs unchanged, applied.json NOT written
  - `--- PASS: TestApplyTokenRefusal_InvalidToken` — tampered token → ErrApprovalTokenInvalid, live doc unchanged
  - `--- PASS: TestApplyTokenRefusal_ValidToken` — valid token → live docs mutated to draft content, applied.json written with Approver + git-committer-date ApprovalTimestamp + 2 AppliedSubtreeIDs + ResultingLiveDocSHA
  - `--- PASS: TestApplyTokenRefusal` (combined: NoToken + InvalidToken + ValidToken subtests)
- AC-NS5-008c — `go test ./internal/navigator/fix/ -run TestApply_AtomicRenameAndLedger -v -count=1`:
  - `--- PASS: TestApply_AtomicRenameAndLedger` — atomic-rename via `.tmp` + os.Rename (no stale .tmp left behind); applied.json ledger carries ResultingLiveDocSHA
- AC-NS5-008a + REQ-NS5-013 — `go test ./internal/navigator/fix/ -run TestApply_OnlyApprovedSubtreesTouched -v -count=1`:
  - `--- PASS: TestApply_OnlyApprovedSubtreesTouched` — out-of-scope draft subtree (over-produced, REQ-NS5-013) excluded from apply; live capability-symbols.json SHA unchanged
- DBT-2 idempotence — `go test ./internal/navigator/fix/ -run TestApply_Idempotence -v -count=1`:
  - `--- PASS: TestApply_Idempotence` — second Apply after re-written draft content is a no-op (live doc SHA unchanged), ledger does not grow (no duplicate IDs)
- AC-NS5-005a (PRESERVE) — `git diff --name-only origin/main...HEAD | grep -E '^internal/(navigator/(sync|detect|route|tiers)|hook/navigator_detect|mx)/'` → grep exit 1 (no matches = PASS — M0/M1/M2/M4 + mx byte-unchanged)
- AC-NS5-005b (consumer-only writes) — `grep -rn 'os.WriteFile\|os.Rename' internal/navigator/fix/*.go` excluding apply.go → only the atomicWriteFile helper in request.go (generic `<path>.tmp`, call site = request.json staging write); apply.go post-approval live-doc writes are the AC-NS5-008c intended exception
- AC-NS5-006 (non-overlap) — `go test ./internal/navigator/fix/ -run TestNonOverlap -v -count=1`:
  - `--- PASS: TestNonOverlap_ReadSurfacesAreReadOnly` — read-surface literals only in READ context (apply.go live-doc writes are the AC-NS5-008c exception via isLiveDocSurface)
  - `--- PASS: TestNonOverlap_NoForbiddenWriteTargets` — zero writes to tiers.json / blueprint/ / decisions/ / work-items.md
  - `--- PASS: TestNonOverlap_ProducerChainScriptsUntouched` — Fix layer does not exec the 3 predecessor chain scripts (reads outputs directly)
- ComputeApprovalToken determinism — `go test ./internal/navigator/fix/ -run TestComputeApprovalToken_Deterministic -v -count=1`:
  - `--- PASS: TestComputeApprovalToken_Deterministic` — token = 64-char hex SHA-256(draft-id + option + provenance), matches a direct sha256 re-derivation (orchestrator can re-derive)

**Baseline-attribution**: `(this run, this tree)` against HEAD pre-commit on branch `worktree-bas-m2-route`. Full package `go test ./internal/navigator/fix/ -count=1` → `ok ... 7.706s`. Coverage `go test -cover ./internal/navigator/fix/` → `coverage: 87.1% of statements` (≥85% target). Cross-platform: `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0. Race `go test -race ./internal/navigator/fix/` → `ok ... 6.953s`. Lint `golangci-lint run ./internal/navigator/fix/... ./internal/cli/...` → `0 issues.`. Vet `go vet ./internal/navigator/fix/... ./internal/cli/...` → exit 0. Subagent boundary `grep -rn 'AskUserQuestion' internal/navigator/fix/*.go internal/cli/navigator_fix.go | grep -v _test.go | grep -v "// "` → 0 matches.

**Gaps**: none for M3.5 scope. The fail-open degraded paths (REQ-NS5-009, AC-NS5-009, exit-0) land in M3.6 — they MUST NOT share the token-refusal exit-non-zero contract (plan.md §F M3.5 last bullet).

**Residual-risk**: the approval.json token is a deterministic hash, not a cryptographically-signed artifact — it binds the apply to a prior orchestrator gate approval recorded in approval.json, NOT to an external secret. A compromised staging dir (attacker can write approval.json) could forge a token; this is acceptable because the staging dir lives under `.moai/project/` (project-local, not distributed) and the approval is a human-gate artifact, not a privilege boundary. The orchestrator owns the only legitimate approval.json writer.

## §E.3 Run-phase Audit-Ready Signal

_(pending run-phase — manager-develop populates this section on run-phase completion)_

## §E.4 Sync-phase Audit-Ready Signal

_(pending sync-phase — manager-docs populates `sync_commit_sha` on the single sync commit carrying the `implemented → completed` transition)_

sync_commit_sha: pending-backfill-sync

## §F Phase 4 Mode Selection

- tier: L
- scope: 10+ new files (internal/navigator/fix/{types,scope,request,apply}.go + internal/cli/navigator_fix.go + 3 test files + testdata/fix-corpus/ fixtures)
- domain count: 1 (Go navigator engine + CLI registration — single-domain Go, NOT cross-domain fan-out)
- file language mix: 100% Go (+ Go test fixtures)
- concurrency benefit: LOW (coding-heavy — Anthropic coding-task parallelism caveat)
- Agent Teams prereqs: N/A (Mode 3 retired)

Mode evaluation:
- Mode 1 (trivial): not selected — Tier L, 500+ LOC new code
- Mode 2 (background): not selected — write-capable implementation
- Mode 3 (agent-team): RETIRED — tombstone
- Mode 4 (parallel): not selected — coding-heavy + single-domain (violates parallelism caveat)
- Mode 5 (sub-agent): SELECTED — sequential manager-develop per milestone is the Anthropic-recommended default for coding-heavy work
- Mode 6 (workflow): not selected — new-code semantic work, not a mechanical uniform transform

Decision: sub-agent (Mode 5)
Justification: Tier L coding-heavy implementation. Per Anthropic's coding-task parallelism caveat ("most coding tasks involve fewer truly parallelizable tasks than research"), sequential sub-agent delegation (manager-develop, one milestone at a time) is the safe default. Single Go domain (navigator/fix + CLI) — no cross-domain fan-out warranting manager-lead. Progression mode: semi-autonomous — orchestrator checks in per milestone (M3.1~M3.6), no /moai goal arming (GLM-backend goal-evaluator reliability concern per prior lessons). Phase 1 audit was self-conducted (plan-auditor agent failed to run on GLM backend) — verdict PASS-WITH-DEBT ~0.88, independent re-audit recommended from a Claude (non-GLM) session.
