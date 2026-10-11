package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestInjectedCoreCarriesCodexReachableObligations guards delta3 finding F4
// (SPEC-ROLE-INJECTION-BUDGET-001): the M2 role-core rewrite relocated two
// always-required lane obligations — the fixed-field dispatch block and the
// stuck-lane disk wait — outside the moai:role-core markers, and no runtime Go
// code reads the always-loaded stub, so a Codex-only install (rules deployed
// under .moai/policies) has no path to them. The injected role-core body is
// the only delivery path that reaches every role session; it must carry both
// obligations' key phrases in both deployed trees.
func TestInjectedCoreCarriesCodexReachableObligations(t *testing.T) {
	surfaces := []string{
		filepath.Join("..", "..", ".claude", "rules", "moai", "workflow", "factory-dispatch.md"),
		filepath.Join("..", "template", "templates", ".claude", "rules", "moai", "workflow", "factory-dispatch.md"),
	}
	phrases := []string{
		"fixed-field", // the dispatch block obligation
		"disk wait",   // the stuck-lane obligation
	}
	for _, rel := range surfaces {
		t.Run(rel, func(t *testing.T) {
			data, err := os.ReadFile(rel)
			if err != nil {
				t.Fatalf("read %s: %v", rel, err)
			}
			regions, marked := config.ExtractRoleCoreRegions(string(data))
			if !marked {
				t.Fatalf("%s carries no role-core markers", rel)
			}
			injected := strings.Join(regions, "\n\n")
			for _, phrase := range phrases {
				if !strings.Contains(injected, phrase) {
					t.Errorf("%s: the injected role-core body does not carry %q — a Codex-only install has no path to this obligation", rel, phrase)
				}
			}
		})
	}
}
