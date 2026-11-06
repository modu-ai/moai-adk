package cli

// N20 (audit 22a0afd247 onwards, residual classification): TestCleanReportsArchive_DryRunMovesNothing asks whether
// archive/2026-06 exists, a month that stopped being the month of the fixture's candidate when the clock moved on, so a dry run
// that created the real shard (archive/<month of the candidate>) passed it. This test names no month: after a dry run the
// archive directory holds what the fixture put there and nothing else.

import (
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/config"
)

func TestCleanReportsArchive_ResidualN20ADryRunLeavesTheArchiveDirectoryAsItWas(t *testing.T) {
	root := setupReportsScaffold(t)
	stubTrackedProbe(t, func(rel string) bool { return false })

	p := printer.New(printer.WithWriters(&strings.Builder{}, &strings.Builder{}))
	if err := runCleanReportsArchiveWithRoot(p, false, config.DefaultReportsArchiveRetentionDays, root); err != nil {
		t.Fatalf("reports-archive dry-run: %v", err)
	}
	entries, err := os.ReadDir(reportsPath(root, "archive"))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if len(names) != 1 || names[0] != "2000-01" {
		t.Errorf("archive/ holds %v after a dry run, want only the fixture's 2000-01", names)
	}
}
