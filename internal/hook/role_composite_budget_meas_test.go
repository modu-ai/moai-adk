package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestCompositeBudgetMeasurement measures the assembled composite for
// leader/lane roles x ko/en locales against the 10,000 delivery cap
// (operator decision d-20261011T081356Z-6c5e / d-20261011T070331Z-1670).
// Producer sizes are the measured/estimated fixed-maximum values from the
// Opus selection audit (in-out-selection-opus.md): ko lane 4,797 (measured),
// en lane ~6,262 (estimate), leader ko ~2,777, leader en ~3,516.
// This test is a measuring instrument: failures here ARE the report.
func TestCompositeBudgetMeasurement(t *testing.T) {
	clearFactoryEnv(t)
	// Deploy the REAL tree rule files into the fixture root.
	root := t.TempDir()
	treeRoot := "../.."
	for _, rel := range roleRuleFiles {
		relPath := filepath.Join(".", rel.Rel)
		data, err := os.ReadFile(filepath.Join(treeRoot, relPath))
		if err != nil {
			t.Fatalf("read tree rule %s: %v", relPath, err)
		}
		writeRoleRuleFixture(t, root, rel, string(data))
	}

	cases := []struct {
		name     string
		roleEnv  []string
		roleName string
		producer int // utf16 units of the fixed-maximum producer content
		lang     string
	}{
		{"lane_ko_glm", []string{config.EnvMoaiFactoryWorker}, "factory-lane", 4797, "ko"},
		{"lane_en_estimated", []string{config.EnvMoaiFactoryWorker}, "factory-lane", 6262, "en"},
		{"leader_ko_glm", []string{config.EnvMoaiFactoryWorkers}, "factory-leader", 2777, "ko"},
		{"leader_en_estimated", []string{config.EnvMoaiFactoryWorkers}, "factory-leader", 3516, "en"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, ev := range tc.roleEnv {
				t.Setenv(ev, "1")
			}
			core, _, _ := buildRoleCoreForTest(t, root)
			if strings.TrimSpace(core) == "" {
				t.Fatal("core is empty")
			}
			if strings.Contains(core, "<!-- moai:role-core-start -->") {
				t.Error("core contains a nested-open marker")
			}
			producer := strings.Repeat("p", tc.producer)
			assembled := producer + "\n\n" + core
			composite, operator := roleRuleSizeGate(assembled, core, root, tc.roleName, tc.lang)
			if operator != "" {
				t.Logf("operator warning present: %.80s", operator)
			}
			t.Logf("RESULT %s: composite utf16=%d (cap 10000, over=%d)", tc.name, utf16Len(composite), utf16Len(composite)-10000)
			if utf16Len(composite) > 10000 {
				t.Logf("OVER CAP: %s composite %d > 10000", tc.name, utf16Len(composite))
			}
		})
	}
}
