# Evaluation Report — SPEC-V3R4-HARNESS-002 Iteration 2

**Evaluator**: evaluator-active (ephemeral per design-constitution §11.4.1)
**SPEC**: SPEC-V3R4-HARNESS-002 Multi-Event Observer Expansion
**Iteration**: 2 (post iter-1 fix commit 97e2b129e)
**Date**: 2026-05-15
**Verdict**: **FAIL**

---

## Executive Summary

Iteration 1 fixed 3 defects (prompt_hash truncation, prompt_len byte vs rune, prompt_full→prompt_content rename). All 3 are confirmed resolved. However, this iteration independently found 2 new blocking defects:

1. **PromptPreview uses 200 runes, SPEC mandates first 64 bytes** (AC-HRN-OBS-008.a FAIL — Functionality dimension)
2. **internal/cli package coverage 67.1% < 85% threshold** (Craft dimension FAIL)

Must-pass firewall: Functionality FAIL → Overall FAIL regardless of other dimension scores.

---

## Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 55/100 | FAIL | AC-HRN-OBS-008.a FAIL: PromptPreview 200-rune vs 64-byte; 12/13 AC PASS |
| Security (25%) | 85/100 | PASS | PII gate verified — Strategy A default confirmed; no raw prompt stored in Strategy A/None; hash[:16]+bytelength only; fail-open preserves privacy posture |
| Craft (20%) | 42/100 | FAIL | internal/cli coverage 67.1% vs 85% threshold; internal/harness 87.9% PASS; go vet clean |
| Consistency (15%) | 88/100 | PASS | Template-first discipline maintained (4 .sh.tmpl wrappers); settings.json.tmpl additive; EventType string values match SPEC exactly; schema version "v1" preserved; omitempty pattern consistent |

**Must-Pass Firewall Check**: Functionality = FAIL → Overall = **FAIL**

---

## AC Verdict Table

| AC | Description | Verdict | Evidence |
|----|-------------|---------|----------|
| AC-HRN-OBS-001 | Three EventType constants added | PASS | types.go:36-42: `session_stop`, `subagent_stop`, `user_prompt` verified |
| AC-HRN-OBS-002 | Three CLI subcommands registered | PASS | hook.go:104-122: `harness-observe-stop`, `harness-observe-subagent-stop`, `harness-observe-user-prompt-submit` |
| AC-HRN-OBS-003 | `learning.enabled` fail-open gate in all 3 handlers | PASS | T-C1/C2/C3 all have NoOp+Preserve tests passing |
| AC-HRN-OBS-004 | Stop handler: session_id, msg_hash[:16], msg_len_bytes, exit_code | PASS | T-C1 `TestRunHarnessObserveStop_RecordsWhenEnabled` PASS; hash 16 hex chars, len bytes |
| AC-HRN-OBS-005 | SubagentStop handler: agent_type, agent_name, parent_session_id, result_summary | PASS | T-C2 `TestRunHarnessObserveSubagentStop_AllFields` PASS |
| AC-HRN-OBS-006 | prompt_hash = SHA-256(prompt)[:16] | PASS | `assistantMessageFields` (hook.go:608-613) confirmed; T-C1 hash 16-char assertion verified |
| AC-HRN-OBS-007 | prompt_len = len([]byte(msg)) | PASS | iter-1 fix confirmed; T-C1 byte-count assertion passes |
| AC-HRN-OBS-008 | UserPromptSubmit PII strategies | PARTIAL — see below |
| AC-HRN-OBS-008.a | Strategy B: prompt_preview = first 64 bytes | **FAIL** | impl uses 200 runes (hook.go:833-840); test expects 200 runes; SPEC REQ-HRN-OBS-013 + acceptance.md §008.a mandate 64 bytes |
| AC-HRN-OBS-008.b | Field name is prompt_content (not prompt_full) | PASS | iter-1 fix confirmed; types.go:118 comment + json tag verified |
| AC-HRN-OBS-009 | SPEC-ID extracted from UserPromptSubmit | PASS | T-C3 `TestRunHarnessObserveUserPromptSubmit_ExtractSpecID` PASS |
| AC-HRN-OBS-010 | Language heuristic in UserPromptSubmit | PASS | T-C3 `TestRunHarnessObserveUserPromptSubmit_LangHeuristic` PASS |
| AC-HRN-OBS-011 | JSONL append-only log | PASS | All 3 handlers use same append pattern; schema additivity verified via omitempty |
| AC-HRN-OBS-012 | LogSchemaVersion "v1" preserved | PASS | types.go: `LogSchemaVersion = "v1"` |
| AC-HRN-OBS-013 | 4 wrapper .sh.tmpl files | PASS | All 4 confirmed in `internal/template/templates/.claude/hooks/moai/` |

---

## Iter-1 Defect Verification

| Defect | Status | Evidence |
|--------|--------|----------|
| prompt_hash not truncated to [:16] | RESOLVED | hook.go:608: `hex.EncodeToString(h[:])[:16]`; T-C1 16-char assertion PASS |
| prompt_len using rune count | RESOLVED | hook.go:613: `len([]byte(msg))`; T-C1 byte-count assertion PASS |
| Field name prompt_full (wrong) | RESOLVED | types.go:118 json tag `prompt_content`; referenced in types.go comment |

---

## Findings

- **[P1-CRITICAL]** `internal/cli/hook.go:833-840` + `internal/harness/types.go:112` — PromptPreview uses 200 runes but AC-HRN-OBS-008.a (maps to REQ-HRN-OBS-013) mandates "first 64 bytes". The types.go comment (line ~112) also says "first 200 chars" which contradicts the SPEC. Test `TestRunHarnessObserveUserPromptSubmit_StrategyB_Preview` is internally consistent with the implementation but both diverge from the SPEC. This is a functional correctness defect against an explicit SPEC requirement.

- **[P2-HIGH]** `internal/cli` package: coverage 67.1% vs 85% threshold. Pre-existing uncovered functions — `runDBSchemaSync`, `runSpecStatus`, `loadMigrationPatterns`, `splitLines`, `trimSpace` — drag the package-level coverage below threshold. These are not SPEC-V3R4-HARNESS-002 functions, but the threshold applies to the package, not per-SPEC additions. TRUST 5 "Tested" criterion requires 85%+ per package.

- **[P3-INFO]** `internal/cli/hook_harness_observe_user_prompt_submit_test.go` — Strategy B test name `StrategyB_Preview` passes with 200-rune expectation (lines match the 200-rune implementation), confirming the spec/impl gap is consistent through both layers but not fixed.

- **[P3-INFO]** AC-HRN-OBS-005 specifies a test named `TestHookHarnessGateUniformity` as a unified table across all 4 handlers. Actual tests are in 3 separate files (`hook_harness_observe_stop_test.go`, `hook_harness_observe_subagent_stop_test.go`, `hook_harness_observe_user_prompt_submit_test.go`). Functionally equivalent coverage exists but test naming and file structure deviate from the AC spec. Non-blocking given functional equivalence.

---

## Remediation

### Fix 1 — PromptPreview: 200 runes → 64 bytes (P1, REQUIRED for PASS)

**File**: `internal/cli/hook.go` lines ~833-840

Change from:
```go
if strategy == UserPromptStrategyPreview && len(prompt) > 0 {
    runes := []rune(prompt)
    end := 200
    if len(runes) < end {
        end = len(runes)
    }
    evt.PromptPreview = string(runes[:end])
}
```

Change to:
```go
if strategy == UserPromptStrategyPreview && len(prompt) > 0 {
    b := []byte(prompt)
    if len(b) > 64 {
        b = b[:64]
    }
    evt.PromptPreview = string(b)
}
```

**File**: `internal/harness/types.go` — update the comment on `PromptPreview` field from "first 200 chars" to "first 64 bytes".

**File**: `internal/cli/hook_harness_observe_user_prompt_submit_test.go` — update `TestRunHarnessObserveUserPromptSubmit_StrategyB_Preview` to assert 64-byte slicing, not 200-rune slicing. For multi-byte characters test the boundary behavior at byte 64.

### Fix 2 — Coverage gap (P2, REQUIRED for TRUST 5 compliance)

Option A (minimal): Add targeted tests for `runDBSchemaSync`, `runSpecStatus`, `loadMigrationPatterns`, `splitLines`, `trimSpace` to bring `internal/cli` above 85%. Estimated 5-8 new test functions.

Option B (structural): Move these pre-existing functions out of scope if they are dead code in the current binary, or mark them explicitly with build tags that exclude from coverage measurement. Requires architectural justification.

Option A is recommended. The functions exist in the package; they should be tested.

---

## Test Run Evidence

```
go test -count=1 -coverprofile=/tmp/cov-cli.out ./internal/cli/...
ok  github.com/modu-ai/moai-adk/internal/cli       9.664s  coverage: 67.1% of statements
ok  github.com/modu-ai/moai-adk/internal/cli/pr    0.360s  coverage: 91.7% of statements
ok  github.com/modu-ai/moai-adk/internal/cli/wizard 1.493s  coverage: 91.3% of statements
ok  github.com/modu-ai/moai-adk/internal/cli/worktree 2.726s  coverage: 84.2% of statements

go test -count=1 -coverprofile=/tmp/cov-harness.out ./internal/harness/...
ok  github.com/modu-ai/moai-adk/internal/harness   coverage: 87.9% of statements

go vet ./internal/cli/... ./internal/harness/...
(clean — no output)
```

Wave C test matrix (24 tests):
- T-C1 (stop): 5/5 PASS
- T-C2 (subagent-stop): 5/5 PASS
- T-C3 (user-prompt-submit): 7/7 PASS (Strategy B test passes against 200-rune expectation — consistent with impl, NOT with SPEC)
- T-C4 (settings.json.tmpl): PASS
- T-C5 (embedded template): PASS

Build: `go build -o /tmp/moai-test ./cmd/moai/` — SUCCESS

---

## Conditions for PASS in Iteration 3

1. `PromptPreview` corrected to byte-slicing at 64 bytes (both impl and test)
2. `types.go` comment updated to "first 64 bytes"
3. `internal/cli` package coverage >= 85% (verified via `go test -cover ./internal/cli/`)
4. All 24 Wave C tests continue to pass
5. `go vet` remains clean
