package statusline

// CanonicalSegments lists the 15 canonical statusline segment keys in display
// order. SegmentRepo is intentionally excluded — it is the 16th constant,
// outside the 15-key statusline schema (SLM-7). This is the single source of
// truth for the segment key set; both the CLI render path and the profile
// write path derive their segment maps from it.
var CanonicalSegments = []string{
	SegmentModel, SegmentContext, SegmentOutputStyle, SegmentDirectory,
	SegmentGitStatus, SegmentClaudeVersion, SegmentMoaiVersion, SegmentGitBranch,
	SegmentSessionTime, SegmentUsage5H, SegmentUsage7D,
	SegmentTask, SegmentWorktree, SegmentEffortThinking,
	SegmentPR,
}

// PresetToSegments expands a statusline preset name (plus an optional custom
// segment map) into a full 15-key segment-to-enabled map over CanonicalSegments.
// Unknown presets fall back to "full" (all enabled). This is the SSOT for preset
// expansion: the CLI render path and the profile write path both call it so a
// non-custom preset always materializes its full segment map at save time, never
// silently no-ops (SLR-3). Runtime precedence (segments map wins over preset) is
// unaffected — this only governs the WRITE-time expansion.
func PresetToSegments(preset string, custom map[string]bool) map[string]bool {
	segments := make(map[string]bool, len(CanonicalSegments))

	switch preset {
	case "compact":
		// Compact: essentials + workflow context (task + PR for SPEC visibility)
		compactEnabled := map[string]bool{
			SegmentModel: true, SegmentContext: true,
			SegmentGitStatus: true, SegmentGitBranch: true,
			SegmentTask: true, SegmentPR: true,
		}
		for _, seg := range CanonicalSegments {
			segments[seg] = compactEnabled[seg]
		}
	case "minimal":
		// Minimal: model + context only (no workflow context)
		minimalEnabled := map[string]bool{
			SegmentModel: true, SegmentContext: true,
		}
		for _, seg := range CanonicalSegments {
			segments[seg] = minimalEnabled[seg]
		}
	case "custom":
		if custom == nil {
			// No custom selections provided, default all to true
			for _, seg := range CanonicalSegments {
				segments[seg] = true
			}
		} else {
			for _, seg := range CanonicalSegments {
				val, exists := custom[seg]
				if exists {
					segments[seg] = val
				} else {
					segments[seg] = true // Default missing segments to enabled
				}
			}
		}
	default:
		// "full" and any unknown preset: all segments enabled
		for _, seg := range CanonicalSegments {
			segments[seg] = true
		}
	}

	return segments
}
