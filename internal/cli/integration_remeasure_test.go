package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegrationRemeasureCommand(t *testing.T) {
	for _, tc := range []struct {
		name, command, want string
		dirty               bool
		wantErr             bool
	}{
		{name: "records clean candidate", command: "true", want: "command: true (exit 0)"},
		{name: "requires command", want: "pass the command", wantErr: true},
		{name: "refuses dirty candidate", command: "true", want: "not clean at start", dirty: true, wantErr: true},
		{name: "reports failed measurement", command: "exit 7", want: "exited 7", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, _, _ := sdMergeFixture(t, true, true, false, 1)
			// Runtime fixture state and linked worktrees are outside the measured source.
			exclude := filepath.Join(root, ".git", "info", "exclude")
			existing, err := os.ReadFile(exclude)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(exclude, append(existing, []byte("\n.moai/\n.claude/\n")...), 0o600); err != nil {
				t.Fatal(err)
			}
			t.Chdir(root)
			if tc.dirty {
				if err := os.WriteFile(filepath.Join(root, "dirty.txt"), []byte("dirty\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			out, err := runIntegration(t, root, "remeasure", "--", tc.command)
			if (err != nil) != tc.wantErr {
				t.Fatalf("output %q, error %v, want error %v", out, err, tc.wantErr)
			}
			observed := out
			if err != nil {
				observed += err.Error()
			}
			if !strings.Contains(observed, tc.want) {
				t.Fatalf("got %q, want %q", observed, tc.want)
			}
		})
	}
}
