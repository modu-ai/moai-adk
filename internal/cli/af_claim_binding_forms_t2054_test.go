package cli

// Card t2054 finding 2 (RC40 review, af.go:61): the t2027 pin covers ledger.db, af/ledger.db and ./ledger.db. This
// pins the forms it does not: a parent-relative path, and a relative path resolved from a symlinked working
// directory. Each must still stop inside ledger.Open and write no heartbeat binding.

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestAFClaim_ParentRelativeAndSymlinkCwdDBReachNoBinding(t *testing.T) {
	project := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", project)

	real := t.TempDir()
	if err := os.MkdirAll(filepath.Join(real, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "cwd-link")
	if err := os.Symlink(filepath.Join(real, "sub"), link); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name, cwd, db string
	}{
		{"parent-relative", filepath.Join(real, "sub"), filepath.Join("..", "x", "ledger.db")},
		{"parent-relative-bare", filepath.Join(real, "sub"), ".." + string(filepath.Separator) + "ledger.db"},
		{"symlink-cwd", link, "ledger.db"},
		{"symlink-cwd-parent", link, filepath.Join("..", "ledger.db")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Chdir(c.cwd)
			cmd := newAFCmd()
			var out, errOut bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&errOut)
			cmd.SetArgs([]string{"claim", "--db", c.db, "--session", "s-forms", "--worktree", project})
			err := cmd.Execute()
			t.Logf("--db %q from %q: err=%v stdout=%q", c.db, c.cwd, err, out.String())
			if err == nil {
				t.Errorf("claim succeeded with a relative --db; the binding path must be checked")
			}
			binding := filepath.Join(project, ".moai", "state", "ledger-heartbeat", "s-forms.json")
			if raw, readErr := os.ReadFile(binding); readErr == nil {
				t.Errorf("a binding was written: %s", raw)
			}
		})
	}
}
