// doctor_todo_store_test.go — SPEC-TODO-STALE-STORE-001 (card t1307) M2:
// the doctor divergence check over the stale project-local queue store.
//
// Three states are the requirement (REQ-TSS-010): divergence fails with
// both last_seq values in the message, a missing home database fails, and
// a missing legacy store passes. An unreadable (zero-byte or non-SQLite)
// store is its own state — reported, never failed as a divergence
// (acceptance §D.2). The check is read-only on every branch (REQ-TSS-013,
// AC-TSS-013).
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

func TestDoctorTodoStoreDivergence_ThreeStates(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(t *testing.T, root, homeDB, legacyDB string)
		want    uikit.CheckStatus
		wantMsg []string // substrings the message or detail must carry
	}{
		{
			name:   "divergent",
			mutate: func(t *testing.T, root, homeDB, legacyDB string) {},
			want:   uikit.CheckFail,
			wantMsg: []string{
				".moai/state/todo/backlog.db",
				"1305",
				"661",
			},
		},
		{
			name: "home-absent",
			mutate: func(t *testing.T, root, homeDB, legacyDB string) {
				if err := os.Remove(homeDB); err != nil {
					t.Fatalf("remove home db: %v", err)
				}
			},
			want:    uikit.CheckFail,
			wantMsg: []string{"home queue database missing"},
		},
		{
			name: "legacy-absent",
			mutate: func(t *testing.T, root, homeDB, legacyDB string) {
				if err := os.RemoveAll(filepath.Dir(legacyDB)); err != nil {
					t.Fatalf("remove legacy dir: %v", err)
				}
			},
			want: uikit.CheckOK,
		},
		{
			name: "unreadable-store",
			mutate: func(t *testing.T, root, homeDB, legacyDB string) {
				if err := os.Truncate(legacyDB, 0); err != nil {
					t.Fatalf("truncate legacy db: %v", err)
				}
			},
			want:    uikit.CheckWarn,
			wantMsg: []string{".moai/state/todo/backlog.db", "unreadable"},
		},
		{
			name: "equal-seqs",
			mutate: func(t *testing.T, root, homeDB, legacyDB string) {
				seedTodoStoreAt(t, legacyDB, 1305)
			},
			want: uikit.CheckOK,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, homeDB, legacyDB := staleStoreFixture(t)
			tc.mutate(t, root, homeDB, legacyDB)

			check := checkTodoStoreDivergence(root, false)

			if check.Status != tc.want {
				t.Errorf("status = %v, want %v (message %q)", check.Status, tc.want, check.Message)
			}
			joined := check.Message + "\n" + check.Detail
			for _, want := range tc.wantMsg {
				if !strings.Contains(joined, want) {
					t.Errorf("message does not carry %q:\n%s\n%s", want, check.Message, check.Detail)
				}
			}
			if check.Name != todoStoreDivergenceCheckName {
				t.Errorf("name = %q, want %q", check.Name, todoStoreDivergenceCheckName)
			}
		})
	}
}

// TestDoctorTodoStoreDivergence_ReadOnly — AC-TSS-013: running the check
// over a divergent tree changes no database byte or mtime and leaves no
// new file (no lock file) behind in either store directory.
func TestDoctorTodoStoreDivergence_ReadOnly(t *testing.T) {
	root, homeDB, legacyDB := staleStoreFixture(t)

	before := map[string]cliDBFingerprint{
		homeDB:   cliFingerprint(t, homeDB),
		legacyDB: cliFingerprint(t, legacyDB),
	}

	_ = checkTodoStoreDivergence(root, false)

	for path, was := range before {
		now := cliFingerprint(t, path)
		if now.sum != was.sum || now.mtime != was.mtime {
			t.Errorf("%s changed across the doctor check: sha %s->%s mtime %d->%d (REQ-TSS-013)",
				path, was.sum, now.sum, was.mtime, now.mtime)
		}
		if strings.Join(now.names, ",") != strings.Join(was.names, ",") {
			t.Errorf("%s directory gained or lost entries across the doctor check: %v -> %v",
				path, was.names, now.names)
		}
	}
}
