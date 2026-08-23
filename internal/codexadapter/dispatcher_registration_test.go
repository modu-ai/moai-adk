package codexadapter

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// dispatcherSourceRel is the file that registers `moai hook <arg>` subcommands.
const dispatcherSourceRel = "internal/cli/hook.go"

// subcommandLit matches the `use` field of each registered hook subcommand,
// e.g. `{"pre-tool", "Handle pre-tool-use event", hook.EventPreToolUse},`.
var subcommandLit = regexp.MustCompile(`\{"([a-z0-9-]+)",\s*"Handle `)

// TestDispatcherArgsExist asserts every dispatcher argument in EventTable is
// actually registered in internal/cli/hook.go (AC-REQ-1a).
//
// This check exists because the SPEC's first draft asserted four Codex events
// had no MoAI counterpart, which the registration list contradicts. The table's
// right-hand side has to be mechanically checkable rather than re-assertable.
func TestDispatcherArgsExist(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	src, err := os.ReadFile(filepath.Join(root, dispatcherSourceRel))
	if err != nil {
		t.Fatalf("read %s: %v", dispatcherSourceRel, err)
	}

	registered := make(map[string]bool)
	for _, m := range subcommandLit.FindAllStringSubmatch(string(src), -1) {
		registered[m[1]] = true
	}
	if len(registered) == 0 {
		t.Fatalf("parsed zero subcommands from %s — the matcher no longer fits the source", dispatcherSourceRel)
	}

	for _, row := range EventTable {
		if !registered[row.DispatcherArg] {
			t.Errorf("%s maps to dispatcher arg %q, which is not registered in %s",
				row.CodexEvent, row.DispatcherArg, dispatcherSourceRel)
		}
	}
}

// repoRoot walks up from the package directory to the module root.
func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found walking up from the package directory")
		}
		dir = parent
	}
}
