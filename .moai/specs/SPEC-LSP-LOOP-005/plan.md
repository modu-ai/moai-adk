# SPEC-LSP-LOOP-005: Implementation Plan

## Phases

### Phase 1: Feedback Struct Extension
- Add `Diagnostics []lsp.Diagnostic` to `internal/loop/feedback.go`
- Update serialization (JSON tags) for cache/resume compatibility
- **LOC**: +50 modified

### Phase 2: GoFeedbackGenerator Integration
- Inject `aggregator *lsp.Aggregator` via constructor
- `Collect(ctx)` calls `aggregator.GetDiagnostics` and merges into Feedback
- Graceful fallback when aggregator is nil
- **LOC**: +100 modified

### Phase 3: Ralph ClassifyFeedback Enhancement
- Update signature to accept diagnostics via Feedback (no breaking change to signature)
- Implement severity/source classification logic
- Preserve integer fallback when diagnostics empty
- **LOC**: +200 modified

### Phase 4: PostTool Hook → LoopController Routing
- Add `RecordFeedback` channel to LoopController
- PostTool hook writes to channel after diagnostics collection
- Backpressure: bounded channel, drop oldest on overflow
- **LOC**: +150 new + modified

### Phase 5: RalphConfig Flag Wiring
- `LintAsInstruction` / `WarnAsInstruction` read in Decide()
- Unit tests for each flag combination
- **LOC**: +100 modified

### Phase 6: Stagnation Signal
- Update `detectStagnation` to check diagnostic count trend
- Upward trend over 3 iterations triggers re-planning
- **LOC**: +100 new

### Phase 7: Tests
- End-to-end test: edit file → hook → aggregator → feedback → classify → decide
- Mock Aggregator, mock PostTool input
- **LOC**: +300 new

## Estimated LOC: +750 new, +450 modified

## Dependencies

- **Hard**: SPEC-LSP-AGG-003 complete
- **Blocks**: None
