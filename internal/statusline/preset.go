package statusline

// CanonicalSegments lists the 15 canonical statusline segment keys in display
// order. SegmentRepo is intentionally excluded — it is the 16th constant,
// outside the 15-key statusline schema (SLM-7). This is the single source of
// truth for the segment key set; both the CLI render path and the profile
// write path derive their segment maps from it.
//
// The PresetToSegments function that previously lived here was retired by
// SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001: named presets (full/compact/minimal)
// no longer exist as a configuration surface. Segments are configured directly.
var CanonicalSegments = []string{
	SegmentModel, SegmentContext, SegmentOutputStyle, SegmentDirectory,
	SegmentGitStatus, SegmentClaudeVersion, SegmentMoaiVersion, SegmentGitBranch,
	SegmentSessionTime, SegmentUsage5H, SegmentUsage7D,
	SegmentTask, SegmentWorktree, SegmentEffortThinking,
	SegmentPR,
}
