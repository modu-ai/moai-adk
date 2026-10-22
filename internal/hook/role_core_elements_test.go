package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestInjectedCoreCarriesObligationElements asserts obligation ELEMENTS per
// the Opus IN/OUT selection (leader gate): IN rows' elements must appear in
// the injected role-core body; OUT rows' elements must appear in the
// read-directive target file's BODY (rule-body delivery — the leader's Opus
// adjudication 「읽기 지시가 가리키는 파일 안이면 수용」).
func TestInjectedCoreCarriesObligationElements(t *testing.T) {
	surfaces := []string{
		filepath.Join("..", "..", ".claude", "rules", "moai", "workflow", "factory-dispatch.md"),
		filepath.Join("..", "template", "templates", ".claude", "rules", "moai", "workflow", "factory-dispatch.md"),
	}
	// row -> element phrases; scope: injected body (IN) or file body (OUT).
	checks := []struct {
		row      string
		inBody   bool
		elements []string
	}{
		{"ALB-0276 (disk wait: the three required items)", true, []string{"reason, whom, recheck point"}},
		{"ALB-0276 (watchdog first)", true, []string{"moai-lane-watchdog"}},
		{"ALB-0276 (recurring cron)", true, []string{"CronCreate"}},
		{"ALB-0276 (leader only at gate boundary)", true, []string{"gate boundary"}},
		{"ALB-0276 (disk evidence before replying)", true, []string{"disk evidence"}},
		{"ALB-0272 (Class B spec omission condition)", false, []string{"not Class B"}},
		{"ALB-0272 (fixed-field, 10-line ceiling)", false, []string{"10-line fixed-field"}},
		{"ALB-0272 (send-result three shapes)", false, []string{"routing", "Cross-session delivery notice"}},
		{"ALB-0303 (git-flow-only condition)", true, []string{"git-flow only"}},
		{"ALB-0303 (acquire before entry)", true, []string{"acquire"}},
		{"ALB-0303 (no-force push)", true, []string{"never force"}},
		{"ALB-0303 (Codex entry means)", true, []string{"Codex: `git -C <absolute release worktree path>`"}},
		{"ALB-0260 (normalized-text rejection criterion)", true, []string{"normalized text is identical"}},
		{"ALB-0260 (the four individual prohibitions)", true, []string{"never folds away, reorders around, drops, or edits"}},
		{"ALB-0255 (conditional, not closed)", false, []string{"conditional, not closed"}},
		{"ALB-0255 (operator-approved card requests only)", false, []string{"operator-approved card requests"}},
		{"ALB-0254 (translation, not invention)", false, []string{"translation, not invention"}},
		{"ALB-0275 (tree_scope: skip self-review)", true, []string{"tree_scope: skip"}},
		{"ALB-0275 (advisory, never replacing)", true, []string{"never replacing the leader evidence read"}},
		{"ALB-0279 (merge-approval token)", false, []string{"LEAD-MERGE-APPROVED"}},
		{"ALB-0279 (measurer attribution)", false, []string{"naming its measurer"}},
		{"ALB-0280 (absence of failure signal)", true, []string{"no failure signal is not a pass"}},
		{"ALB-0280 (declared verification path)", true, []string{"declared verification path"}},
		{"ALB-0284 (full-head binding)", false, []string{"full head SHA"}},
		{"ALB-0284 (no findings)", false, []string{"no findings or unresolved findings"}},
		{"ALB-0285 (stale or incomplete evidence blocks)", false, []string{"Missing, failed, running, stale, malformed, or contradictory evidence"}},
		{"LNC (in-session transition: exit keeping the tree)", true, []string{"ExitWorktree(keep)"}},
		{"LNC (in-session transition: enter the leased tree)", true, []string{"EnterWorktree("}},
		{"LNC (prior-card context is non-authoritative)", true, []string{"conversation context"}},
		{"ALB-0294 (integration-tree re-entry)", false, []string{"returns to the card worktree"}},
		{"ALB-0295 (configured base, never reused)", false, []string{"never reused"}},
		{"read directive (first core sentence)", true, []string{"read `.claude/rules/moai/workflow/factory-dispatch.md` to the end"}},
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
				if !c.inBody {
					hay = string(data)
				}
				for _, phrase := range c.elements {
					if !strings.Contains(hay, phrase) {
						scope := "injected role-core body"
						if !c.inBody {
							scope = "rule body (read-directive target)"
						}
						t.Errorf("%s: %s lacks %s element %q", rel, scope, c.row, phrase)
					}
				}
			}
		})
	}
}
