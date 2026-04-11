# SPEC-LSP-LOOP-005: Acceptance Criteria

## Functional

### AC1: Feedback Carries Diagnostics
- [ ] `Feedback` struct has `Diagnostics []lsp.Diagnostic` field
- [ ] JSON round-trip preserves Diagnostics
- [ ] When aggregator is nil (disabled), Diagnostics is empty slice

### AC2: GoFeedbackGenerator Uses Aggregator
- [ ] `GoFeedbackGenerator.Collect` invokes Aggregator when available
- [ ] Diagnostics merged into Feedback alongside go test/vet output
- [ ] Backwards compatibility: existing tests (no aggregator) pass

### AC3: Ralph Severity-Aware Classification
- [ ] Error + compiler → Blocker
- [ ] Error + staticcheck SA* → Approval
- [ ] Warning + staticcheck → AutoFix
- [ ] Information → Skip
- [ ] Integer fallback when Diagnostics empty

### AC4: PostTool → LoopController Routing
- [ ] PostTool hook writes feedback event to channel
- [ ] Channel is bounded; overflow drops oldest with warn log
- [ ] LoopController consumes events in Decide loop

### AC5: RalphConfig Flags Wired
- [ ] `LintAsInstruction: true` routes lint findings as instruction, not block
- [ ] `WarnAsInstruction: true` routes warnings as instruction
- [ ] Combined flags compose correctly

### AC6: Stagnation Detection
- [ ] Rising diagnostic count over 3 iterations triggers re-planning signal
- [ ] Falling count resets stagnation counter
- [ ] Test with synthetic feedback stream

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
