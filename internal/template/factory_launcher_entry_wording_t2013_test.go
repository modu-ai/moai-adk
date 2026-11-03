package template

import (
	"strings"
	"testing"
)

// TestT2013_LauncherTableStatesAbsolutePathAndLaneRejoin pins the launcher table
// of the factory-dispatch mechanics rule (local copy and template mirror): an
// absolute -w path is entered only when it is a registered worktree, an
// unregistered one is refused before anything is written, and a dead lane is
// rejoined with its own label.
func TestT2013_LauncherTableStatesAbsolutePathAndLaneRejoin(t *testing.T) {
	markers := []string{
		"the launcher enters the path only when it is a registered worktree of this repository",
		"`.moai/worktrees/*` trees included",
		"an unregistered absolute path is refused before anything is written",
		"`moai cc -a --name <lane-label>`",
		"granted only when no live session holds that label, refused otherwise",
		"the launcher moves no lease",
	}
	retired := []string{
		"Claude may ask to approve an external worktree path",
	}
	for copyName, root := range t2013HandoffRoots() {
		t.Run(copyName, func(t *testing.T) {
			section := t2013HandoffSection(t, root, "factory-dispatch-mechanics.md", "## Isolation — launcher")
			for _, m := range markers {
				if !strings.Contains(section, m) {
					t.Errorf("launcher table lacks %q", m)
				}
			}
			for _, m := range retired {
				if strings.Contains(section, m) {
					t.Errorf("launcher table still carries the retired wording %q", m)
				}
			}
		})
	}
}
