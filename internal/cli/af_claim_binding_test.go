package cli

// Card t2027 finding 3 (P2, static review of af.go:219): a relative --db would be stored as given in the heartbeat
// binding. Today every relative --db fails inside ledger.Open before any binding is written, so this test passes.
// It pins that: if ledger.Open ever accepts a relative path, this test fails and flags the latent af.go:219 issue.

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestAFClaim_RelativeDBReachesNoBinding(t *testing.T) {
	project := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", project)
	t.Chdir(t.TempDir())

	for _, rel := range []string{"ledger.db", filepath.Join("af", "ledger.db"), "." + string(filepath.Separator) + "ledger.db"} {
		cmd := newAFCmd()
		var out, errOut bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&errOut)
		cmd.SetArgs([]string{"claim", "--db", rel, "--session", "s-rel", "--worktree", project})
		err := cmd.Execute()
		t.Logf("--db %q: err=%v stdout=%q", rel, err, out.String())
		if err == nil {
			t.Errorf("--db %q: claim succeeded; the binding is checked below", rel)
		}
		_, statErr := os.Stat(filepath.Join(project, ".moai", "state", "ledger-heartbeat", "s-rel.json"))
		if statErr == nil {
			raw, _ := os.ReadFile(filepath.Join(project, ".moai", "state", "ledger-heartbeat", "s-rel.json"))
			t.Errorf("--db %q: a binding was written: %s", rel, raw)
		}
	}
}
