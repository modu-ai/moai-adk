package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
)

// A file the restore refused is reported, not dropped: the report line names the
// refused path and states that nothing was written.
func TestUnmanagedRestoreMessages_ReportsTargetRefusal(t *testing.T) {
	refs := []backup.UnmanagedRestoreRef{{Path: ".moai/config/evaluator-profiles/sub/x.md", Action: "target-refused"}}
	got := unmanagedRestoreMessages(refs, "/backups/20260101_000000")
	if len(got) != 1 || !strings.Contains(got[0], ".moai/config/evaluator-profiles/sub/x.md") || !strings.Contains(got[0], "nothing written") {
		t.Fatalf("messages = %q, want one line naming the refused path and stating nothing was written", got)
	}
}
