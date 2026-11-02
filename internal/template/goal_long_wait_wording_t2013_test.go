package template

import (
	"strings"
	"testing"
)

// TestT2013_GoalWaitPassageNamesTheNextEventAndTheBounds pins the passage that
// says an unmet goal does not block while the session's card holds an open wait
// naming a next check event, and that every bound still wins over the wait.
func TestT2013_GoalWaitPassageNamesTheNextEventAndTheBounds(t *testing.T) {
	markers := []string{
		"does not block the turn end while the session's card holds an open `factory wait` record that names a next check event",
		"its reason carries a `next=` event other than `unknown`",
		"a wait without a next event, a woken wait, an unreadable record",
		"every bound still applies",
		"so a wait never extends a goal beyond its ceiling",
		"step 6 of the leader check pass covers that",
	}
	for copyName, root := range t2013HandoffRoots() {
		t.Run(copyName, func(t *testing.T) {
			section := t2013HandoffSection(t, root, "goal-directive-detail.md", "## Native `/goal` and long waits")
			for _, m := range markers {
				if !strings.Contains(section, m) {
					t.Errorf("goal wait passage lacks %q", m)
				}
			}
		})
	}
}

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
