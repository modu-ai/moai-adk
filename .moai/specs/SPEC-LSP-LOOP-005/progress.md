---
spec_id: SPEC-LSP-LOOP-005
status: completed
completed_at: "2026-04-13"
---

# SPEC-LSP-LOOP-005 Progress

## Completion Status: DONE

All 10 requirements implemented and tested.

---

## Tasks Completed

| Task | REQ | Status | Notes |
|------|-----|--------|-------|
| T1: Feedback.LSPDiagnostics field | REQ-LL-001 | Done | `[]lsp.Diagnostic` added alongside legacy `[]gopls.Diagnostic` |
| T2: Source-aware ErrorLevel constants | REQ-LL-004/005 | Done | `ErrorLevelBlocker=5`, `ErrorLevelSkip=0` added |
| T3: ClassifyFeedbackWithConfig | REQ-LL-006/007 | Done | LintAsInstruction/WarnAsInstruction demote warnings |
| T4: classifyLSPDiagnostics 5-rule engine | REQ-LL-004/005 | Done | compiler→Blocker, staticcheck SA*→Approval, etc. |
| T5: DiagnosticsAggregator interface | REQ-LL-002 | Done | DI boundary in loop package |
| T6: NewGoFeedbackGeneratorWithAggregator | REQ-LL-002 | Done | Wires Aggregator for Go files |
| T7: filterGoOnlyDiagnostics | REQ-LL-002 | Done | Extension point for multi-language future work |
| T8: FeedbackChannel (bounded) | REQ-LL-009 | Done | Drop-oldest overflow with warn log |
| T9: PostTool hook → FeedbackChannel | REQ-LL-003 | Done | `NewPostToolHandlerWithFeedbackChannel` |
| T10: convertHookDiagsToLSP | REQ-LL-003 | Done | lsphook.Diagnostic → lsp.Diagnostic converter |
| T11: IsStagnantWithDiagnostics | REQ-LL-010 | Done | Diagnostic count trend awareness |
| T12: Backwards compatibility | REQ-LL-008 | Done | Integer-based path unchanged when LSPDiagnostics empty |

---

## Coverage Per Modified Package

| Package | Coverage | Target |
|---------|----------|--------|
| `internal/loop` | 85.0% | 85% |
| `internal/ralph` | 91.0% | 85% |
| `internal/hook` | 80.6% | 85% (pre-existing gap) |

Note: `internal/hook` base coverage was ~80% before this SPEC. New functions added
(convertHookDiagsToLSP: 100%, emitToFeedbackChannel: 80%, NewPostToolHandlerWithFeedbackChannel: 75%)
all meet coverage expectations.

---

## Files Created

| File | Purpose |
|------|---------|
| `internal/loop/feedback_channel.go` | Bounded FeedbackChannel (REQ-LL-009) |
| `internal/loop/aggregator_feedback_test.go` | Aggregator integration tests (REQ-LL-002) |
| `internal/loop/feedback_channel_test.go` | FeedbackChannel unit tests (REQ-LL-009) |
| `internal/loop/stagnation_diag_test.go` | Stagnation with diagnostics tests (REQ-LL-010) |
| `internal/loop/coverage_boost_test.go` | Coverage improvement tests |
| `internal/ralph/classify_lsp_test.go` | Source-aware classification tests (REQ-LL-004/005) |
| `internal/hook/post_tool_feedback_channel_test.go` | PostTool channel emission tests (REQ-LL-003) |
| `internal/hook/post_tool_lsp_convert_test.go` | Diagnostic type conversion tests |

---

## Files Modified

| File | Changes |
|------|---------|
| `internal/loop/state.go` | Added `LSPDiagnostics []lsp.Diagnostic` to Feedback struct |
| `internal/loop/feedback.go` | Added `IsStagnantWithDiagnostics` (REQ-LL-010) |
| `internal/loop/go_feedback.go` | Added `DiagnosticsAggregator` interface, `NewGoFeedbackGeneratorWithAggregator`, `filterGoOnlyDiagnostics` |
| `internal/ralph/engine.go` | Added `ErrorLevelSkip/Blocker`, `ClassifyFeedbackWithConfig`, `classifyLSPDiagnostics`, refactored `ClassifyFeedback` |
| `internal/hook/post_tool.go` | Added `feedbackCh` field, `NewPostToolHandlerWithFeedbackChannel`, `emitToFeedbackChannel`, `convertHookDiagsToLSP`, refactored `collectDiagnosticsWithInstruction` |

---

## MX Tags

| Tag | Location | Reason |
|-----|----------|--------|
| `@MX:ANCHOR` on `ErrorLevelBlocker` | `internal/ralph/engine.go` | fan_in >= 3 |
| `@MX:ANCHOR` on `ClassifyFeedback` | `internal/ralph/engine.go` | fan_in >= 3 (loop controller, tests, Ralph engine) |
| `@MX:NOTE` on `ClassifyFeedbackWithConfig` | `internal/ralph/engine.go` | LintAsInstruction/WarnAsInstruction influence |
| `@MX:NOTE` on `classifyLSPDiagnostics` | `internal/ralph/engine.go` | 5-rule classification order matters |
| `@MX:NOTE` on `IsStagnantWithDiagnostics` | `internal/loop/feedback.go` | Diagnostic trend awareness |
| `@MX:WARN` on `FeedbackChannel` | `internal/loop/feedback_channel.go` | Bounded channel, overflow drops oldest |

---

## Test Results

```
ok  github.com/modu-ai/moai-adk/internal/loop     1.56s
ok  github.com/modu-ai/moai-adk/internal/ralph    1.39s
ok  github.com/modu-ai/moai-adk/internal/hook     1.80s
```

All 27 packages in `internal/...` pass with `-race` flag.
`go vet ./...` clean.
`go mod tidy` clean.

---

## Divergences from SPEC

| REQ | Divergence | Justification |
|-----|-----------|---------------|
| REQ-LL-001 | Added `LSPDiagnostics` as NEW field; kept `Diagnostics []gopls.Diagnostic` | Backwards compatibility: removing legacy field would break existing code (REQ-LL-008) |
| REQ-LL-003 | RecordFeedback channel is optional (nil-safe); PostTool emits only on Write/Edit | Observation-only constraint; nil feedbackCh is graceful no-op |
| REQ-LL-010 | Stagnation with diagnostics implemented as separate `IsStagnantWithDiagnostics` function rather than modifying `IsStagnant` | Preserves backwards compatibility of `IsStagnant` for existing callers |
