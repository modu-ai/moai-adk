package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestT2013_WaitVerbReachesLanesThroughRule pins that the lane-wait section of
// the factory-dispatch rule body (local copy and template mirror) names the
// `moai factory wait` verb and the `wake:` line `moai factory complete` prints,
// so a lane has a documented path to record a wait and a completing session
// knows whom to nudge.
func TestT2013_WaitVerbReachesLanesThroughRule(t *testing.T) {
	roots := map[string]string{
		"template": filepath.Join("templates", ".claude", "rules", "moai", "workflow"),
		"local":    filepath.Join("..", "..", ".claude", "rules", "moai", "workflow"),
	}
	for copyName, root := range roots {
		t.Run(copyName, func(t *testing.T) {
			b, err := os.ReadFile(filepath.Join(root, "factory-dispatch.md"))
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			text := string(b)
			i := strings.Index(text, "### Lane waits are explicit, and stalls are watched")
			if i < 0 {
				t.Fatal("lane-wait section heading missing")
			}
			section := text[i:]
			if j := strings.Index(section[4:], "\n## "); j >= 0 {
				section = section[:j+4]
			}
			for _, m := range []string{"moai factory wait", "`wake:` line", "--waker", "sends nothing"} {
				if !strings.Contains(section, m) {
					t.Errorf("lane-wait section lacks marker %q", m)
				}
			}
		})
	}
}
