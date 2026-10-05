package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/template"
)

// TestUpdateClaudeToBothSucceeds reproduces the claude -> both transition:
// a claude deployment leaves .agents/skills/<skill> skill-mirror links, and the
// update that follows ApplyHarness("both") must complete its template sync.
// On a tree whose both profile ships no file beneath a mirror-link path this
// passes with or without the fix; it guards the transition once catalog skills
// are re-homed under .agents/skills (the dual-harness recovery work, t1100).
func TestUpdateClaudeToBothSucceeds(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

	root := filepath.Join(t.TempDir(), "proj")
	initCmd := newInitTestCmd()
	for name, val := range map[string]string{"llm": "claude", "non-interactive": "true", "name": "transition", "language": "go", "mode": "tdd"} {
		if err := initCmd.Flags().Set(name, val); err != nil {
			t.Fatalf("set --%s: %v", name, err)
		}
	}
	var initOut bytes.Buffer
	initCmd.SetOut(&initOut)
	initCmd.SetErr(&initOut)
	if err := runInit(initCmd, []string{root}); err != nil {
		t.Fatalf("init --llm claude: %v\n%s", err, initOut.String())
	}
	if err := template.ApplyHarness(root, "both"); err != nil {
		t.Fatal(err)
	}

	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("force", true, "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Bool("no-hooks", true, "")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(context.Background())
	if err := runTemplateSyncWithReporter(cmd, nil, true); err != nil {
		t.Fatalf("template sync after claude -> both: %v\n%s", err, tailLines(buf.String(), 20))
	}
}
