package template

import (
	"strings"
	"testing"
)

// TestT2013_GoalSectionNamesNativeGoalAndWaiting pins the long-wait goal
// guidance in the goal-directive detail companion.
func TestT2013_GoalSectionNamesNativeGoalAndWaiting(t *testing.T) {
	markers := []string{
		"the native `/goal` has no stop guard and no turn ceiling",
		"re-entered with empty turns",
		"a long-wait goal uses `/moai goal`",
		"three consecutive stagnant turns",
		"run `/goal clear`",
	}
	for copyName, root := range t2013HandoffRoots() {
		t.Run(copyName, func(t *testing.T) {
			section := t2013HandoffSection(t, root, "goal-directive-detail.md", "## Native `/goal` and long waits")
			for _, m := range markers {
				if !strings.Contains(section, m) {
					t.Errorf("goal long-wait section lacks %q", m)
				}
			}
		})
	}
}
