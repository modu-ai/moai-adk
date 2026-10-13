package report

// preserved_lines_test.go pins REQ-RC30-015 (U-16, AC-RC30-17): the default
// reconciliation summary names the preserved files by one count line, and the
// full path list appears only under --verbose. Archived removals and conflicts
// keep their path lists, because REQ-UPM-031 keeps deletions visible.

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreservedLinesCountOnly(t *testing.T) {
	counts := ReconciliationCounts{Refreshed: 1, Preserved: 3, ArchivedRemoved: 1}
	preserved := []string{".claude/rules/moai/a.md", ".claude/rules/moai/b.md", ".claude/rules/moai/c.md"}
	archived := []string{".claude/rules/moai/old-rule.md"}

	// Default: one count line for the preserved category, no per-path line.
	dflt := RenderReconciliation(counts, nil, preserved, archived)
	assert.Equal(t, 1, strings.Count(dflt, "\n  preserved: "))
	assert.Contains(t, dflt, "\n  preserved: 3 file(s) (run with --verbose to list)")
	for _, p := range preserved {
		assert.NotContains(t, dflt, p)
	}
	assert.Contains(t, dflt, "\n  archived-removed: .claude/rules/moai/old-rule.md")

	// --verbose: every preserved path is listed, and no count hint is printed.
	verbose := RenderReconciliation(counts, nil, preserved, archived)
	for _, p := range preserved {
		assert.Contains(t, verbose, "\n  preserved: "+p)
	}
	assert.NotContains(t, verbose, "run with --verbose to list")

	// Nothing preserved: no preserved line in either mode.
	none := RenderReconciliation(ReconciliationCounts{Refreshed: 1}, nil, nil, nil)
	assert.NotContains(t, none, "preserved:")
}
