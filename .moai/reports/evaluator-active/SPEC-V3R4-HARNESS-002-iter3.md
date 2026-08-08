# Evaluation Report — SPEC-V3R4-HARNESS-002 Iter 3

**SPEC**: SPEC-V3R4-HARNESS-002 Multi-Event Observer Expansion
**Evaluator**: evaluator-active (fresh spawn, ephemeral judgment — §11.4.1)
**Iteration**: 3 (post iter-2 P1 fix: PromptPreview 64-byte boundary)
**Timestamp**: 2026-05-15T~08:00Z
**Worktree**: `/Users/goos/.moai/worktrees/MoAI-ADK/SPEC-V3R4-HARNESS-002`

---

## Overall Verdict: PASS

---

## Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 88/100 | PASS | 12/13 ACs fully verified; AC-HRN-OBS-006 "enabled + append → 3-line byte-identical" scenario partially covered (base-field + append tested separately, combined fixture absent) |
| Security (25%) | 92/100 | PASS | PII strategies enforced; no raw prompt in log; hash-only default; SHA-256 truncation correct; no AskUserQuestion in harness code; no secrets |
| Craft (20%) | 90/100 | PASS | `internal/harness` coverage 87.9% (> 85% threshold); `internal/harness/safety` 94.3%; 27 new tests all GREEN; `go vet` clean; UTF-8 boundary fix correct |
| Consistency (15%) | 90/100 | PASS | Template-first discipline upheld (3 new `.sh.tmpl` files + additive `settings.json.tmpl`); LogSchemaVersion v1 unchanged; omitempty pattern consistent with baseline; camelCase stdin fields align with Claude Code hook schema |

---

## Must-Pass Firewall

| Must-Pass Dimension | Threshold | Score | Status |
|---|---|---|---|
| Functionality | ≥ 60 | 88 | PASS |
| Security | ≥ 60 | 92 | PASS |

Neither must-pass dimension is failing. Overall PASS is not blocked.

---

## AC Verification Matrix

| AC ID | Description | Verdict | Evidence |
|---|---|---|---|
| AC-HRN-OBS-001 | Cobra subcommand registration (help text) | PASS | `go run ./cmd/moai hook --help` shows 3 new subcommands |
| AC-HRN-OBS-002 | Stop handler — session_stop event, hash+len fields | PASS | `TestRunHarnessObserveStop_RecordsWhenEnabled` PASS; SHA-256[:16] + byte length verified |
| AC-HRN-OBS-002.a | Empty last_assistant_message → hash/len omitted | PASS | `TestRunHarnessObserveStop_EmptyMessageNoHashFields` PASS |
| AC-HRN-OBS-003 | SubagentStop handler — subagent_stop event, agent fields | PASS | `TestRunHarnessObserveSubagentStop_RecordsAllFields` PASS; camelCase agentName/agentType/agent_id + parentSessionID verified |
| AC-HRN-OBS-004 | UserPromptSubmit — Strategy A default, no raw prompt | PASS | `TestRunHarnessObserveUserPromptSubmit_StrategyA_HashLenLang` PASS |
| AC-HRN-OBS-004.a | Strategy None — no-op | PASS | `TestRunHarnessObserveUserPromptSubmit_StrategyNone_NoOp` PASS |
| AC-HRN-OBS-004.b | Strategy C (full) | PASS | `TestRunHarnessObserveUserPromptSubmit_StrategyC_Full` PASS |
| AC-HRN-OBS-005 | Gate uniformity — all 4 handlers no-op when disabled | PASS | `TestGateUniformity_AllHandlersNoOpWhenDisabled` 4/4 PASS |
| AC-HRN-OBS-006 | Schema additivity — old entries byte-identical after append | PARTIAL | omitempty serialization tested in `types_extension_test.go:44`; PreservesExisting (disabled) in stop+subagent tests; **NO dedicated "enabled + 2 pre-seeded baseline entries + Stop append → 3 lines, originals byte-identical" test** |
| AC-HRN-OBS-007 | PII default (Strategy A) — no preview/content fields | PASS | `TestRunHarnessObserveUserPromptSubmit_StrategyA_HashLenLang` asserts prompt_preview absent |
| AC-HRN-OBS-008 | Gate disabled — file byte-identical preservation | PASS | Stop + SubagentStop + GateUniformity tests all cover disabled preservation |
| AC-HRN-OBS-008.a | PromptPreview 64-byte UTF-8 boundary (iter 2 P1 fix) | PASS | `TestRunHarnessObserveUserPromptSubmit_StrategyB_Preview` 3/3 PASS: short_ascii_under_64, long_ascii_over_64 (64 'a'), multibyte_korean_over_64 (21 '가'=63 bytes ≤ 64, utf8.ValidString=true) |
| AC-HRN-OBS-009 | Fail-open on invalid strategy | PASS | `TestRunHarnessObserveUserPromptSubmit_FailOpenOnInvalidStrategy` PASS |
| AC-HRN-OBS-010 | EventType enum extension (3 new constants) | PASS | `TestEventType_Extension`: session_stop="session_stop", subagent_stop="subagent_stop", user_prompt="user_prompt" verified |
| AC-HRN-OBS-011.a | Template files — 4 wrapper .sh.tmpl present | PASS | `internal/template/templates/.claude/hooks/moai/` has handle-harness-observe.sh.tmpl + 3 new files confirmed |
| AC-HRN-OBS-011.b | settings.json.tmpl additive registration | PASS | Lines 98/100 (Stop), 122/124 (SubagentStop), 191/193 (UserPromptSubmit) confirmed; original slots preserved |
| AC-HRN-OBS-012 | Contract preservation (FROZEN zone, 5-Layer Safety) | PASS | `.claude/rules/moai/design/constitution.md` not in git diff; LogSchemaVersion="v1" unchanged |
| AC-HRN-OBS-013 | Latency benchmark informational (< 100ms P95, non-blocking) | N/A | acceptance.md states "reports values without failing the build if exceeded (informational)"; `hook_harness_latency_bench_test.go` absent — per AC definition this is informational only, not a CI gate |

---

## Findings

### [LOW] AC-HRN-OBS-006 missing "enabled + append → 3-line, byte-identical" test fixture

`internal/cli/hook_harness_observe_stop_test.go` — no test number

**Description**: AC-HRN-OBS-006 acceptance.md Verification block explicitly requires: "Test fixture writes the two baseline entries, runs a new observer with `learning.enabled: true`, then reads the file and asserts: total line count 3, first two lines unchanged, third line is the new event with optional fields populated."

**Current state**: Two partial tests exist that together partially cover this scenario:
- `TestRunHarnessObserveStop_RecordsWhenEnabled` (enabled=true, fresh log, 1 line result) — covers append correctness
- `TestRunHarnessObserveStop_PreservesExistingLogWhenDisabled` (disabled=false, 1 pre-seeded line) — covers preservation (disabled path only)

Neither test: (a) pre-seeds 2 baseline entries, (b) runs with enabled=true, (c) asserts exactly 3 lines, (d) asserts first 2 lines are byte-identical to original. The full AC-HRN-OBS-006 combined scenario (read backward-compatible + append + verify) is untested.

**Severity**: LOW. The underlying behavior (append-only, no truncation) is exercised implicitly by `TestRunHarnessObserveStop_RecordsWhenEnabled`, and the JSONL append mechanic is the same code path. This is a test coverage gap, not a behavior gap. The actual risk of regression here is low.

**Recommendation**: Add `TestRunHarnessObserveStop_SchemaAdditivity_ThreeLines` matching the acceptance.md Verification block verbatim.

### [INFO] `types_extension_test.go` line 87 field name discrepancy

`internal/harness/types_extension_test.go:87`

**Description**: Line 87 tests for absence of field `prompt_full`, but the actual JSON field name in the `Event` struct is `prompt_content` (via the `json:"prompt_content,omitempty"` tag). The test passes vacuously because neither name appears when zero-valued.

**Impact**: No functional defect. The test is testing the right behavior (absence of optional field when zero) but testing the wrong field name. If `prompt_full` were ever accidentally introduced as an alias, this test would catch it — but it misses direct verification of `prompt_content` absence specifically.

**Recommendation**: Change line 87 to `"prompt_content"` to match the actual struct tag.

### [INFO] `hook_harness_latency_bench_test.go` absent

No dedicated file for `BenchmarkHookHarnessObserveLatency`.

**Description**: AC-HRN-OBS-013 states the benchmark is "informational only" and "does not fail the build if exceeded." Per the acceptance.md definition, this is not a CI gate. The missing benchmark file is noted here as information but does not affect the verdict.

---

## Iteration 2 P1 Fix Verification (AC-HRN-OBS-008.a)

**Command**: `go test -count=1 -v -run "TestRunHarnessObserveUserPromptSubmit_StrategyB_Preview" ./internal/cli/...`

**Result**:
```
--- PASS: TestRunHarnessObserveUserPromptSubmit_StrategyB_Preview (0.00s)
    --- PASS: TestRunHarnessObserveUserPromptSubmit_StrategyB_Preview/short_ascii_under_64 (0.00s)
    --- PASS: TestRunHarnessObserveUserPromptSubmit_StrategyB_Preview/long_ascii_over_64 (0.00s)
    --- PASS: TestRunHarnessObserveUserPromptSubmit_StrategyB_Preview/multibyte_korean_over_64 (0.00s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli  0.371s
```

**Fix location**: `internal/cli/hook.go:835-843`
```go
b := []byte(prompt)
end := min(len(b), 64)
for end > 0 && !utf8.Valid(b[:end]) {
    end--
}
evt.PromptPreview = string(b[:end])
```

All 3 sub-tests verify `len([]byte(preview)) <= 64` AND `utf8.ValidString(preview)` AND exact value match. The multibyte Korean case correctly truncates 22 '가' chars (66 bytes) to 21 '가' chars (63 bytes ≤ 64, valid UTF-8). PASS confirmed.

---

## Coverage Summary

| Package | Coverage | Threshold | Status |
|---|---|---|---|
| `internal/harness` | 87.9% | 85% | PASS (+2.9pp margin) |
| `internal/harness/safety` | 94.3% | 85% | PASS (+9.3pp margin) |
| `internal/harness` combined | 88.8% | 85% | PASS |

New handler tests (`internal/cli/hook_harness_observe_*.go`): 27 tests, all PASS. Coverage uplift in scope per evaluator constraint (pre-existing uncovered `internal/cli` functions excluded).

---

## Recommendations

1. **[P2] Add `TestRunHarnessObserveStop_SchemaAdditivity_ThreeLines`** (AC-HRN-OBS-006 gap): Pre-seed 2 baseline entries, run Stop with enabled=true, assert 3-line result, assert original lines byte-identical. This closes the one partial AC in the matrix.

2. **[P3] Fix `types_extension_test.go:87` field name**: Change `"prompt_full"` to `"prompt_content"` to match the actual `json:` tag and make the absence assertion meaningful.

3. **[P4] Add `BenchmarkHookHarnessObserveLatency`**: While non-blocking (AC-HRN-OBS-013 informational), adding the benchmark would complete the acceptance.md verification block and provide future regression detection for hook latency.
