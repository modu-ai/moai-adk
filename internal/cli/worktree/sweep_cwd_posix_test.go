//go:build !windows

// sweep_cwd_posix_test.go — the lsof NUL-field parser test, kept beside the
// !windows source it exercises (sweep_cwd_posix.go). parseLsofCWDs does not
// exist on the Windows side of the build, so a test in the untagged
// sweep_test.go would fail `GOOS=windows go vet` with an undefined symbol.

package worktree

import "testing"

// TestParseLsofCWDs pins the NUL-field parser extracted from the lsof
// wrapper: n/-prefixed fields become paths, the (deleted) suffix is trimmed,
// and non-path fields are dropped.
func TestParseLsofCWDs(t *testing.T) {
	out := []byte("p123\x00n/some/tree\x00n/other/tree (deleted)\x00\x00nrelative\x00")
	got := parseLsofCWDs(out)
	want := []string{"/some/tree", "/other/tree"}
	if len(got) != len(want) {
		t.Fatalf("parsed %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entry %d = %q, want %q", i, got[i], want[i])
		}
	}
}
