package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestT2013_RuleWordingMatchesCompletionOrder pins the three factory-dispatch
// rule files (local copy and template mirror) to the completion order the
// `moai factory complete` guidance line prints: ExitWorktree(keep), default
// branch check, operator /clear, next-card request, preserve-first. The old
// "no clear" statements must be gone from both copies.
func TestT2013_RuleWordingMatchesCompletionOrder(t *testing.T) {
	files := []string{
		"factory-dispatch.md",
		"factory-dispatch-core.md",
		"factory-dispatch-cards.md",
	}
	roots := map[string]string{
		"template": filepath.Join("templates", ".claude", "rules", "moai", "workflow"),
		"local":    filepath.Join("..", "..", ".claude", "rules", "moai", "workflow"),
	}
	stale := []string{
		"asks the operator to clear nothing",
		"does not ask the operator to clear the lane session",
		"The leader never asks the operator to clear a lane",
		"in-session, no clear",
	}
	// Order markers, each required in every file.
	order := []string{"ExitWorktree(keep)", "default-branch", "/clear", "next card"}

	for copyName, root := range roots {
		for _, name := range files {
			t.Run(copyName+"/"+name, func(t *testing.T) {
				b, err := os.ReadFile(filepath.Join(root, name))
				if err != nil {
					t.Fatalf("read: %v", err)
				}
				text := string(b)
				for _, s := range stale {
					if strings.Contains(text, s) {
						t.Errorf("stale phrase still present: %q", s)
					}
				}
				at := 0
				for _, m := range order {
					i := strings.Index(text[at:], m)
					if i < 0 {
						t.Errorf("order marker %q missing or out of order after offset %d", m, at)
						break
					}
					at += i + len(m)
				}
			})
		}
	}
}
