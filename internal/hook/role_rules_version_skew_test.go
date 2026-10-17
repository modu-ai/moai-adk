// role_rules_version_skew_test.go — SPEC-ROLE-INJECTION-BUDGET-001
// AC-RIB-004/005 (REQ-RIB-006/007): the version-skew family.
//
// Symptom B (spec.md §A.2): a deployed rules tree from an older generation
// (e.g. rc.28 rule files without role-core markers) meeting a newer binary
// (rc.29). t1729 reworked the handling: an unmarked (or skewed) deployed
// tree no longer fails the session — the injection serves the binary's
// embedded template copy with a one-line notice (role_rules_fallback_test.go
// pins the success arms). What stays on the failure path is a rule file the
// fallback cannot cover (absent/unreadable/empty/malformed), and its
// InjectionFailed detail must still name the skew and the remediation
// (commit on the base branch → `moai worktree sync`; primary checkout →
// `moai update`) in every operator locale.
//
// REQ-RIB-007 binding: the version-skew predicate is unit-tested with
// FIXTURES ONLY — no live-repo self-assertion (this repository's own
// v3.1.3-stamp/rc.29-binary state is the motivating skew; a live assertion
// would be permanently red).
package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// skewFixtureRoot builds a fixture project whose deployed dispatch rule is
// UNMARKED (the rc.28-generation shape that triggers the missing-markers
// failure) and whose system.yaml carries the given template_version.
func skewFixtureRoot(t *testing.T, systemYAML string) string {
	t.Helper()
	root := t.TempDir()
	writeRoleRuleFixture(t, root, roleRuleFiles[0], "# old-generation dispatch rule without role-core markers\n")
	writeRoleRuleFixture(t, root, roleRuleFiles[1], "# old-generation messaging rule without role-core markers\n")
	cfgDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "system.yaml"), []byte(systemYAML), 0o644); err != nil {
		t.Fatalf("write system.yaml fixture: %v", err)
	}
	return root
}

// TestRoleRulesVersionSkewUpdateGuidance asserts the remediation command is
// present in every operator locale's InjectionFailed warning (REQ-RIB-006).
func TestRoleRulesVersionSkewUpdateGuidance(t *testing.T) {
	for _, lang := range []string{"ko", "ja", "zh", "en"} {
		t.Run(lang, func(t *testing.T) {
			warn := roleRuleLocaleFor(lang).InjectionFailed("factory-lane", "role rule file carries no markers: x", "main")
			if !strings.Contains(warn, "moai update") {
				t.Errorf("locale %q InjectionFailed carries no moai update guidance: %s", lang, warn)
			}
		})
	}
}

// TestRoleRulesVersionSkewFailureDetailNamesRemedy pins the two skew shapes
// after t1729: an unmarked skewed tree SUCCEEDS through the embedded
// fallback (real embedded seam — the shipped templates carry markers), while
// a rule file the fallback cannot cover (the absent dispatch rule) keeps the
// REQ-ALB-009 failure path whose detail names the stamp and the remediation.
func TestRoleRulesVersionSkewFailureDetailNamesRemedy(t *testing.T) {
	// The fixture stamp deliberately differs from the build-default version
	// (pkg/version defaults to v3.1.3) so the skew is detectable and both
	// versions are nameable in the detail (REQ-RIB-006).
	// The stamp sits under the moai section, the shape the shipped system.yaml uses
	// (card t1619, codex review P2: the stamp is read as YAML, not searched).
	t.Run("unmarked_skewed_tree_falls_back", func(t *testing.T) {
		root := skewFixtureRoot(t, "moai:\n  template_version: v0.0.0-fixture-stamp\n")
		t.Setenv(config.EnvMoaiFactoryWorker, "lane-1")

		inj := roleRuleInjectionFor(root, "startup", "en")
		if inj.OperatorNotice != "" {
			t.Fatalf("unmarked skewed tree must fall back to the embedded copy, not fail; notice: %s", inj.OperatorNotice)
		}
		if !strings.Contains(inj.Context, "embedded template copy") {
			t.Errorf("fallback injection does not say the core came from the embedded copy:\n%s", inj.Context)
		}
	})

	t.Run("absent_rule_keeps_failure_naming_skew", func(t *testing.T) {
		root := skewFixtureRoot(t, "moai:\n  template_version: v0.0.0-fixture-stamp\n")
		t.Setenv(config.EnvMoaiFactoryWorker, "lane-1")
		removeRoleRule(t, root, roleRuleFiles[0])

		inj := roleRuleInjectionFor(root, "startup", "en")
		if inj.OperatorNotice == "" {
			t.Fatal("absent rule file must keep the InjectionFailed path")
		}
		if !strings.Contains(inj.OperatorNotice, "moai update") {
			t.Errorf("failure detail carries no moai update guidance: %s", inj.OperatorNotice)
		}
		if !strings.Contains(inj.OperatorNotice, "v0.0.0-fixture-stamp") {
			t.Errorf("failure detail does not name the rule-side stamp: %s", inj.OperatorNotice)
		}
	})
}
