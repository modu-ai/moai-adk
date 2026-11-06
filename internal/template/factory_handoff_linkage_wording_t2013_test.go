package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// t2013HandoffSection returns the text from heading up to the next level-2
// heading (or the end of the file).
func t2013HandoffSection(t *testing.T, root, file, heading string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, file))
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}
	text := string(b)
	i := strings.Index(text, heading)
	if i < 0 {
		t.Fatalf("%s: heading %q missing", file, heading)
	}
	section := text[i:]
	if j := strings.Index(section[len(heading):], "\n## "); j >= 0 {
		section = section[:len(heading)+j]
	}
	return section
}

func t2013HandoffRoots() map[string]string {
	return map[string]string{
		"template": filepath.Join("templates", ".claude", "rules", "moai", "workflow"),
		"local":    filepath.Join("..", "..", ".claude", "rules", "moai", "workflow"),
	}
}

// TestT2013_LeaseRecoveryStepsAreStated pins the lease-recovery sentences under
// the lease renewal passage of the factory-dispatch mechanics rule.
func TestT2013_LeaseRecoveryStepsAreStated(t *testing.T) {
	markers := []string{
		"the registered owner of the card runs",
		"factory next --card <id> --run <run-id>",
		"applies the expiry and leases the card to the owner again in a single step",
		"keeping stage,\nworktree, evidence and uncommitted work",
		"still refuses without writing",
		"retired and must not be used",
		"does\nnot re-lease a card it does not own this way",
		"never retried inside one watchdog pass",
		"an expired lease alone never wakes or reassigns anyone",
	}
	// The earlier recovery (a refused stage call returning the card to assigned)
	// is retired; its wording must not survive in the passage.
	retired := []string{
		"the stage verb run against that card is itself refused",
		"returns the card to assigned with stage, worktree and evidence kept",
	}
	for copyName, root := range t2013HandoffRoots() {
		t.Run(copyName, func(t *testing.T) {
			section := t2013HandoffSection(t, root, "factory-dispatch-mechanics.md", "## Factory Mode mechanics")
			for _, m := range markers {
				if !strings.Contains(section, m) {
					t.Errorf("lease renewal passage lacks %q", m)
				}
			}
			for _, m := range retired {
				if strings.Contains(section, m) {
					t.Errorf("lease renewal passage still carries the retired recovery %q", m)
				}
			}
		})
	}
}

// TestT2013_DispatchRuleStatesUnmeasuredReadAndGoalRelease pins the lane wait
// paragraph (a failed lookup is not a hold) and the lane card transition
// paragraph (an armed goal is released at completion) in factory-dispatch.md.
func TestT2013_DispatchRuleStatesUnmeasuredReadAndGoalRelease(t *testing.T) {
	cases := []struct {
		heading string
		markers []string
	}{
		{"## The dispatch cycle", []string{
			"is `unmeasured`, not a hold or a withdrawal",
			"record a wait only on an observed condition",
		}},
		{"## Lane card transition", []string{
			"`/moai goal clear`",
			"an unfinished card keeps its goal",
			"no code clears it automatically",
		}},
	}
	for copyName, root := range t2013HandoffRoots() {
		for _, c := range cases {
			t.Run(copyName+"/"+c.heading, func(t *testing.T) {
				section := t2013HandoffSection(t, root, "factory-dispatch.md", c.heading)
				for _, m := range c.markers {
					if !strings.Contains(section, m) {
						t.Errorf("%s lacks %q", c.heading, m)
					}
				}
			})
		}
	}
}
