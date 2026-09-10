//go:build windows

package kanban

import (
	"strings"
	"testing"
)

// TestBacklogDSNWindowsDrivePath pins the DSN shape for a real Windows
// backslash path: filepath.ToSlash must fold it into the root-absolute URI
// form (file:///C:/…), leaving no drive colon for the driver to mistake for
// a URI authority (t426 windows census axis 1).
func TestBacklogDSNWindowsDrivePath(t *testing.T) {
	got := backlogDSN(`C:\Users\dev\proj\.moai\state\todo\backlog.db`)
	if !strings.HasPrefix(got, "file:///C:/Users/dev/proj/.moai/state/todo/backlog.db?") {
		t.Errorf("backlogDSN(windows path) = %q, want a file:///C:/ root-absolute URI", got)
	}
}
