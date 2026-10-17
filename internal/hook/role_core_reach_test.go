package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestInjectedCoreCarriesCodexReachableObligations guards the Codex-only
// delivery path after the Opus IN/OUT selection (leader gate, card t1617):
// IN rows' key phrases must appear in the injected role-core body; OUT rows'
// key phrases must appear in the read-directive target file's BODY (the
// injection opens with a [HARD] read directive naming that file — the
// leader's Opus adjudication 「읽기 지시가 가리키는 파일 안이면 수용」).
func TestInjectedCoreCarriesCodexReachableObligations(t *testing.T) {
	surfaces := []string{
		filepath.Join("..", "..", ".claude", "rules", "moai", "workflow", "factory-dispatch.md"),
		filepath.Join("..", "template", "templates", ".claude", "rules", "moai", "workflow", "factory-dispatch.md"),
	}
	// phrase, inInjectedBody
	checks := []struct {
		phrase string
		inBody bool
	}{
		{"disk wait", true},                                          // ALB-0276 (IN): stuck-lane obligation
		{"gate boundary", true},                                      // ALB-0276: leader reach limit
		{"git-flow only", true},                                      // ALB-0303 (IN): landing scope condition
		{"never force", true},                                        // ALB-0303: push discipline
		{"normalized text is identical", true},                       // ALB-0260 (IN): rejection criterion
		{"never folds away, reorders around, drops, or edits", true}, // ALB-0260
		{"tree_scope: skip", true},                                   // ALB-0275 (IN): leader self-review
		{"no failure signal is not a pass", true},                    // ALB-0280 (IN)
		{"fixed-field", false},                                       // ALB-0272 (OUT): body delivery
		{"not Class B", false},                                       // ALB-0272: Class B spec omission
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
			for _, c := range checks {
				hay := injected
				scope := "injected role-core body"
				if !c.inBody {
					hay = string(data)
					scope = "read-directive target file body"
				}
				if !strings.Contains(hay, c.phrase) {
					t.Errorf("%s: %s lacks %q", rel, scope, c.phrase)
				}
			}
		})
	}
}
