package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestInjectedCoreCarriesObligationElements asserts obligation ELEMENTS (not
// mere string presence of the old compression): each check names a detail the
// base obligation carried and the delta3/Opus audits found missing or
// weakened in the injected role-core body. Element-level per the operator
// decision d-20261011T081356Z-6c5e.
func TestInjectedCoreCarriesObligationElements(t *testing.T) {
	surfaces := []string{
		filepath.Join("..", "..", ".claude", "rules", "moai", "workflow", "factory-dispatch.md"),
		filepath.Join("..", "template", "templates", ".claude", "rules", "moai", "workflow", "factory-dispatch.md"),
	}
	// row -> element phrases that must all appear in the injected body.
	checks := []struct {
		row      string
		elements []string
	}{
		{"ALB-0276 (disk wait: the three required items)", []string{"reason, whom, recheck point"}},
		{"ALB-0276 (watchdog first)", []string{"moai-lane-watchdog"}},
		{"ALB-0276 (recurring cron)", []string{"CronCreate"}},
		{"ALB-0272 (Class B spec omission condition)", []string{"not Class B"}},
		{"ALB-0272 (fixed-field, 10-line ceiling)", []string{"10-line fixed-field"}},
		{"ALB-0272 (send-result three shapes)", []string{"routing", "Cross-session delivery notice"}},
		{"ALB-0303 (git-flow-only condition)", []string{"git-flow only"}},
		{"ALB-0303 (acquire before entry)", []string{"acquire"}},
		{"ALB-0303 (no-force push)", []string{"never force"}},
		{"ALB-0260 (normalized-text rejection criterion)", []string{"normalized text is identical"}},
		{"ALB-0260 (the four individual prohibitions)", []string{"never folds away, reorders around, drops, or edits"}},
		{"ALB-0255 (conditional, not closed)", []string{"conditional, not closed"}},
		{"ALB-0255 (operator-approved card requests only)", []string{"operator-approved card requests"}},
		{"ALB-0254 (translation, not invention)", []string{"translation, not invention"}},
		{"ALB-0275 (tree_scope: skip self-review)", []string{"tree_scope: skip"}},
		{"ALB-0275 (advisory, never replacing)", []string{"never replacing the leader evidence read"}},
		{"ALB-0279 (merge-approval token)", []string{"LEAD-MERGE-APPROVED"}},
		{"ALB-0279 (measurer attribution)", []string{"naming its measurer"}},
		{"ALB-0280 (absence of failure signal)", []string{"no failure signal is not a pass"}},
		{"ALB-0284 (Merge Risk line)", []string{"Merge Risk:"}},
		{"ALB-0284 (combined endpoint)", []string{"/commits/{sha}/status"}},
		{"ALB-0294 (/clear after, never before)", []string{"never before"}},
		{"ALB-0295 (configured base, never reused)", []string{"never reused"}},
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
				for _, phrase := range c.elements {
					if !strings.Contains(injected, phrase) {
						t.Errorf("%s: injected role-core lacks %s element %q", rel, c.row, phrase)
					}
				}
			}
		})
	}
}
