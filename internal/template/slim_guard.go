package template

import (
	"fmt"
	"io/fs"
)

// @MX:NOTE: [AUTO] CATALOG-002 REQ-021 friendly error — sentinel CATALOG_SLIM_HARNESS_MISSING + 4-substring hint for moai doctor diagnostic routing.
//
// AssertBuilderHarnessAvailable returns a CATALOG_SLIM_HARNESS_MISSING-tagged
// error when the builder-harness agent is absent from the given project FS.
// When builder-harness is present (e.g., after the user opted into full deploy
// via --all or MOAI_DISTRIBUTE_ALL=1, or after SPEC-V3R4-CATALOG-005 bootstrap),
// returns nil.
//
// The returned error message contains four well-known substrings that
// downstream tooling (moai doctor) can match for diagnostic routing:
//   - "CATALOG_SLIM_HARNESS_MISSING" (sentinel)
//   - "MOAI_DISTRIBUTE_ALL=1" (env opt-out hint)
//   - "moai init --all" (flag opt-out hint)
//   - "SPEC-V3R4-CATALOG-005" (forward reference to auto-bootstrap)
//
// Source: SPEC-V3R4-CATALOG-002 REQ-021.
func AssertBuilderHarnessAvailable(projectFS fs.FS) error {
	if projectFS == nil {
		return nil
	}
	_, err := fs.Stat(projectFS, ".claude/agents/moai/builder-harness.md")
	if err == nil {
		return nil
	}
	return fmt.Errorf(
		"CATALOG_SLIM_HARNESS_MISSING: builder-harness omitted in slim mode. "+
			"Run `moai init --all` or set MOAI_DISTRIBUTE_ALL=1, or wait for "+
			"SPEC-V3R4-CATALOG-005 auto-bootstrap. (underlying: %w)", err)
}
