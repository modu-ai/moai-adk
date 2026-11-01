package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestT2013_StandardLandingSaysDraftDoesNotStartReviewAndNonRequiredFailureDoesNotBlockReady
// pins that the standard-landing section of the factory-dispatch mechanics rule
// (local copy and template mirror) carries both sentences: a draft PR does not
// start the native review, and a failing check outside the protected required
// set does not block the ready conversion.
func TestT2013_StandardLandingSaysDraftDoesNotStartReviewAndNonRequiredFailureDoesNotBlockReady(t *testing.T) {
	roots := map[string]string{
		"template": filepath.Join("templates", ".claude", "rules", "moai", "workflow"),
		"local":    filepath.Join("..", "..", ".claude", "rules", "moai", "workflow"),
	}
	markers := []string{
		"does not start the native review",
		"leaving the PR in draft while waiting is not a wait reason",
		"outside the protected required set",
		"does not block the ready conversion",
		"records its wait with `moai factory wait` and ends its turn",
		"repaired by the same lane, in the same card worktree, with forward commits",
		"one-shot recheck",
		"the wake: line covers predecessor landings only",
	}
	for copyName, root := range roots {
		t.Run(copyName, func(t *testing.T) {
			b, err := os.ReadFile(filepath.Join(root, "factory-dispatch-mechanics.md"))
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			text := string(b)
			i := strings.Index(text, "## The lane's standard landing")
			if i < 0 {
				t.Fatal("standard landing heading missing")
			}
			section := text[i:]
			if j := strings.Index(section[4:], "\n## "); j >= 0 {
				section = section[:j+4]
			}
			for _, m := range markers {
				if !strings.Contains(section, m) {
					t.Errorf("standard landing section lacks %q", m)
				}
			}
		})
	}
}
