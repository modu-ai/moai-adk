package config

import (
	"os"
	"path/filepath"
	"testing"
)

// contractHeadroomBytes is the margin the Codex contract documents must keep
// under CodexContractByteCeiling (card t1944, split B of t1923). The ceiling
// guard only fails at the limit; this test fails while the margin is thinner
// than the headroom, so growth cannot quietly spend it.
const contractHeadroomBytes = 2048

func TestContractDocumentsKeepHeadroom(t *testing.T) {
	root := filepath.Join("..", "..")
	docs := []string{
		filepath.Join(root, "CLAUDE.md"),
		filepath.Join(root, "internal", "template", "templates", "CLAUDE.md.tmpl"),
	}
	limit := CodexContractByteCeiling - contractHeadroomBytes
	for _, p := range docs {
		fi, err := os.Stat(p)
		if err != nil {
			t.Fatalf("stat %s: %v", p, err)
		}
		if n := int(fi.Size()); n > limit {
			t.Errorf("%s is %d bytes; headroom limit is %d (ceiling %d - %d), over by %d",
				p, n, limit, CodexContractByteCeiling, contractHeadroomBytes, n-limit)
		}
	}
}
