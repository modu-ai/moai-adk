package worktreeroot

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadPinUsesYAMLScalarSemantics(t *testing.T) {
	for _, tc := range []struct {
		name, yaml, want string
		wantErr          bool
	}{
		{"quoted-comment", "git_strategy:\n  worktree_project_key: 'chosen' # operator pin\n", "chosen", false},
		{"inline-map", "git_strategy: {worktree_project_key: chosen}\n", "chosen", false},
		{"block-scalar", "git_strategy:\n  worktree_project_key: |-\n    chosen\n", "chosen", false},
		{"nested-decoy", "git_strategy:\n  manual:\n    worktree_project_key: decoy\n", "", false},
		{"malformed", "git_strategy: [unterminated\n", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, gitStrategyYAML)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(tc.yaml), 0o600); err != nil {
				t.Fatal(err)
			}
			got, err := readPin(root)
			if got != tc.want || (err != nil) != tc.wantErr {
				t.Fatalf("readPin = %q, %v; want %q, error=%v", got, err, tc.want, tc.wantErr)
			}
		})
	}
}
