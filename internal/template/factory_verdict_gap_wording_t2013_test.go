package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestT2013_VerdictGapWordingDropsUncommitted pins the verdict-file gap sentence
// in the factory-dispatch rule files (local copy and template mirror). Verdict
// files live under the gitignored .moai/reports/ and are read from disk, so
// "uncommitted" cannot be a gap condition; the binding condition is the
// verdict's audited_sha against the card's HEAD.
func TestT2013_VerdictGapWordingDropsUncommitted(t *testing.T) {
	files := []string{
		"factory-dispatch.md",
		"factory-dispatch-core.md",
	}
	roots := map[string]string{
		"template": filepath.Join("templates", ".claude", "rules", "moai", "workflow"),
		"local":    filepath.Join("..", "..", ".claude", "rules", "moai", "workflow"),
	}
	const stale = "unreadable, or uncommitted verdict file"
	const lead = "the leader reads the verdict **file**"
	const tail = "is a gap exactly like a missing progress record."

	for copyName, root := range roots {
		for _, name := range files {
			t.Run(copyName+"/"+name, func(t *testing.T) {
				b, err := os.ReadFile(filepath.Join(root, name))
				if err != nil {
					t.Fatalf("read: %v", err)
				}
				text := string(b)
				if strings.Contains(text, stale) {
					t.Errorf("stale phrase still present: %q", stale)
				}
				i := strings.Index(text, lead)
				if i < 0 {
					t.Fatalf("verdict sentence lead %q missing", lead)
				}
				j := strings.Index(text[i:], tail)
				if j < 0 {
					t.Fatalf("verdict sentence tail %q missing", tail)
				}
				sentence := text[i : i+j+len(tail)]
				for _, m := range []string{"audited_sha", "HEAD"} {
					if !strings.Contains(sentence, m) {
						t.Errorf("verdict sentence lacks marker %q: %s", m, sentence)
					}
				}
			})
		}
	}
}
