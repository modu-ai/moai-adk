# SPEC-LSP-LOOP-005: Acceptance Criteria

## Functional

### AC1: Feedback Carries Diagnostics
- [ ] `Feedback` struct has `Diagnostics []lsp.Diagnostic` field (REQ-LL-001)
- [ ] JSON round-trip preserves Diagnostics (REQ-LL-001)
- [ ] When aggregator is nil (disabled), Diagnostics is empty slice (REQ-LL-008)

### AC2: GoFeedbackGenerator Uses Aggregator
- [ ] `GoFeedbackGenerator.Collect` invokes Aggregator when available (REQ-LL-002)
- [ ] Diagnostics merged into Feedback alongside go test/vet output (REQ-LL-002)
- [ ] Backwards compatibility: existing tests (no aggregator) pass (REQ-LL-008)
- [ ] Only Go-language diagnostics are merged; non-Go results are filtered out (REQ-LL-002)

### AC3: Ralph Severity-Aware Classification
- [ ] Error + compiler → Blocker (REQ-LL-004, REQ-LL-005)
- [ ] Error + staticcheck SA* → Approval (REQ-LL-005)
- [ ] Warning + staticcheck → AutoFix (REQ-LL-005)
- [ ] Information → Skip (REQ-LL-005)
- [ ] Integer fallback when Diagnostics empty (REQ-LL-008)

### AC4: PostTool → LoopController Routing
- [ ] PostTool hook writes feedback event to channel (REQ-LL-003)
- [ ] Channel is bounded; overflow drops oldest with warn log (REQ-LL-009)
- [ ] LoopController consumes events in Decide loop (REQ-LL-003)

### AC5: RalphConfig Flags Wired
- [ ] `LintAsInstruction: true` routes lint findings as instruction, not block (REQ-LL-006)
- [ ] `WarnAsInstruction: true` routes warnings as instruction (REQ-LL-007)
- [ ] Combined flags compose correctly (REQ-LL-006, REQ-LL-007)

### AC6: Stagnation Detection
- [ ] Rising diagnostic count over 3 iterations triggers re-planning signal (REQ-LL-010)
- [ ] Falling count resets stagnation counter (REQ-LL-010)
- [ ] Test with synthetic feedback stream (REQ-LL-010)

## Quality (TRUST 5)

### Tested
- [ ] ≥ 85% coverage for `internal/loop/` and `internal/ralph/`
- [ ] End-to-end scenario test (edit → feedback → decision)
- [ ] Concurrent PostTool invocations pass race detector

### Readable
- [ ] Classification rules documented in `engine.go` godoc
- [ ] Feedback flow diagram in architecture rules

### Unified
- [ ] Single `Feedback` type used across loop, ralph, hooks
- [ ] No duplicate diagnostic types

### Secured
- [ ] Channel overflow handled gracefully (no panic)
- [ ] Diagnostic content sanitized before systemMessage injection

### Trackable
- [ ] `@MX:ANCHOR` on `Ralph.ClassifyFeedback`
- [ ] SPEC-LSP-LOOP-005 in commit scopes

## Deliverables

- [ ] 3+ modified files (feedback.go, go_feedback.go, engine.go)
- [ ] 2+ new files (channel routing, stagnation detection)
- [ ] Comprehensive end-to-end test
- [ ] CHANGELOG entry
