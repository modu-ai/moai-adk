package hook

// t1729 — embedded-template fallback tests for the SessionStart role-rule
// injection.
//
// Defect: a worktree branched from a main whose committed rule files predate
// the installed binary carries unmarked (or stale) deployed copies
// permanently — `moai update` refreshed only the tree where it ran, and the
// old remediation ("run `moai update`") cannot fix a worktree. The binary
// itself carries a fresh copy of both rule files, so an unmarked deployed
// file — or a detected version skew — must serve the embedded copy with a
// one-line notice instead of failing the session.
//
// The embedded source rides the roleRuleEmbeddedRule seam: tests stub it with
// a fixture core whose text cannot come from anywhere else, so a passing
// assertion proves the core was served from the embedded copy. Seam-stubbing
// tests must not run t.Parallel().

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// embeddedFallbackCoreText is the fixture embedded core — its text is unique
// to this file, so Contains proves the embedded source was served.
const embeddedFallbackCoreText = "FALLBACK-EMBEDDED-CORE: the lane waits for disk evidence before advancing a card."

// markedEmbeddedRule renders a marked rule file whose core carries
// embeddedFallbackCoreText (the shape the real embedded template copies have).
func markedEmbeddedRule() string {
	return "# Fixture rule (embedded copy)\n\n" +
		config.RoleCoreMarkerStart + "\n" + embeddedFallbackCoreText + "\n" +
		config.RoleCoreMarkerEnd + "\n\nplain tail\n"
}

// stubEmbeddedRule replaces the embedded-template seam for one test: both
// role rule files resolve to the given content (ok=true) or to a miss
// (ok=false). Restored via t.Cleanup.
func stubEmbeddedRule(t *testing.T, content string, ok bool) {
	t.Helper()
	prev := roleRuleEmbeddedRule
	roleRuleEmbeddedRule = func(rel string) (string, bool) {
		if rel == roleRuleFiles[0].Rel || rel == roleRuleFiles[1].Rel {
			return content, ok
		}
		return "", false
	}
	t.Cleanup(func() { roleRuleEmbeddedRule = prev })
}

// writeSystemYAMLFixture writes a system.yaml carrying the given body under
// the deployed config path (the stamp fixture the skew tests use).
func writeSystemYAMLFixture(t *testing.T, root, systemYAML string) {
	t.Helper()
	cfgDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "system.yaml"), []byte(systemYAML), 0o644); err != nil {
		t.Fatalf("write system.yaml fixture: %v", err)
	}
}

// TestRoleRulesUnmarkedDeploysEmbeddedFallback is the card's primary RED: a
// deployed rule file WITHOUT role-core markers (the stale-committed-main
// worktree shape) must inject the embedded copy's core with the one-line
// notice — not take the REQ-ALB-009 failure path.
func TestRoleRulesUnmarkedDeploysEmbeddedFallback(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	stubEmbeddedRule(t, markedEmbeddedRule(), true)

	root := t.TempDir()
	writeRoleRuleFixture(t, root, roleRuleFiles[0], "# stale deployed rule\n\nno role-core markers here\n")
	writeRoleRuleFixture(t, root, roleRuleFiles[1], smallMarkedRule("messaging"))

	inj := roleRuleInjectionFor(root, "startup", langEnglish)
	if inj.OperatorNotice != "" {
		t.Fatalf("unmarked deployed file must fall back to the embedded copy, not fail; notice: %s", inj.OperatorNotice)
	}
	if !strings.Contains(inj.Context, embeddedFallbackCoreText) {
		t.Fatalf("injection does not carry the embedded copy's core:\n%s", inj.Context)
	}
	if strings.Contains(inj.Context, "no role-core markers here") {
		t.Fatalf("injection leaked the unmarked deployed body instead of the embedded core:\n%s", inj.Context)
	}
	if !strings.Contains(inj.Context, "embedded template copy") {
		t.Fatalf("injection does not say the core came from the embedded copy (one-line notice missing):\n%s", inj.Context)
	}
	t.Logf("PASS: unmarked deployed file served the embedded core with the fallback notice")
}

// TestRoleRulesVersionSkewSwapsToEmbedded pins the skew arm: marked deployed
// files under a system.yaml stamp differing from the running binary are STALE
// by definition — the injection serves the binary's embedded copy, not the
// stale deployed core.
func TestRoleRulesVersionSkewSwapsToEmbedded(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	stubEmbeddedRule(t, markedEmbeddedRule(), true)

	root := t.TempDir()
	writeRoleRuleFixture(t, root, roleRuleFiles[0], smallMarkedRule("STALE-DEPLOYED-CORE"))
	writeRoleRuleFixture(t, root, roleRuleFiles[1], smallMarkedRule("messaging"))
	// The stamp differs from pkg/version's build default (v3.1.3), the same
	// skew shape TestRoleRulesVersionSkewFailureDetailNamesRemedy uses.
	writeSystemYAMLFixture(t, root, "moai:\n  template_version: v0.0.0-fixture-stamp\n")

	inj := roleRuleInjectionFor(root, "startup", langEnglish)
	if inj.OperatorNotice != "" {
		t.Fatalf("skew alone must not fail the session; notice: %s", inj.OperatorNotice)
	}
	if !strings.Contains(inj.Context, embeddedFallbackCoreText) {
		t.Fatalf("skewed tree does not serve the embedded copy's core:\n%s", inj.Context)
	}
	if strings.Contains(inj.Context, "STALE-DEPLOYED-CORE") {
		t.Fatalf("skewed tree served the stale deployed core:\n%s", inj.Context)
	}
	if !strings.Contains(inj.Context, "embedded template copy") {
		t.Fatalf("skewed-tree injection does not say the core came from the embedded copy:\n%s", inj.Context)
	}
	t.Logf("PASS: version skew swapped the injection source to the embedded copy")
}

// TestRoleRulesFallbackIsPerRule pins the mixed shape: one unmarked rule
// falls back while its sibling keeps the deployed core, and the notice line
// appears exactly once.
func TestRoleRulesFallbackIsPerRule(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	stubEmbeddedRule(t, markedEmbeddedRule(), true)

	root := t.TempDir()
	writeRoleRuleFixture(t, root, roleRuleFiles[0], "# stale\n\nno role-core markers here\n")
	writeRoleRuleFixture(t, root, roleRuleFiles[1], smallMarkedRule("messaging"))

	inj := roleRuleInjectionFor(root, "startup", langEnglish)
	if inj.OperatorNotice != "" {
		t.Fatalf("mixed tree must not fail; notice: %s", inj.OperatorNotice)
	}
	if !strings.Contains(inj.Context, embeddedFallbackCoreText) {
		t.Fatalf("unmarked rule did not fall back to the embedded core:\n%s", inj.Context)
	}
	if !strings.Contains(inj.Context, "Small binding block for messaging") {
		t.Fatalf("marked sibling lost its deployed core:\n%s", inj.Context)
	}
	if got := strings.Count(inj.Context, "embedded template copy"); got != 1 {
		t.Fatalf("fallback notice must appear exactly once, got %d:\n%s", got, inj.Context)
	}
	t.Logf("PASS: fallback is per-rule with a single notice line")
}

// TestRoleRulesUnmarkedWithoutEmbeddedFailsVisible pins the preserved
// REQ-ALB-009 retreat: when the embedded copy is unavailable (a packaging
// bug), an unmarked deployed file still fails visible — warning AND
// directive, never a silent start.
func TestRoleRulesUnmarkedWithoutEmbeddedFailsVisible(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	stubEmbeddedRule(t, "", false)

	root := t.TempDir()
	writeRoleRuleFixture(t, root, roleRuleFiles[0], "# stale\n\nno role-core markers here\n")
	writeRoleRuleFixture(t, root, roleRuleFiles[1], smallMarkedRule("messaging"))

	inj := roleRuleInjectionFor(root, "startup", langEnglish)
	assertFailVisible(t, inj, "unmarked-without-embedded")
}

// TestRoleRulesRemediationNamesWorktreeCause pins the corrected remediation
// text (card fix 2): the skew guidance names the worktree cause — an
// uncommitted `moai update` on the base branch — and the refresh verb
// (`moai worktree sync`), not a bare "run `moai update` inside the worktree".
func TestRoleRulesRemediationNamesWorktreeCause(t *testing.T) {
	root := t.TempDir()
	writeSystemYAMLFixture(t, root, "moai:\n  template_version: v0.0.0-fixture-stamp\n")

	detail := roleRuleSkewDetail(root)
	if detail == "" {
		t.Fatal("skewed fixture produced no skew detail")
	}
	if !strings.Contains(detail, "moai worktree sync") {
		t.Errorf("skew detail does not name the worktree refresh verb: %s", detail)
	}
	if !strings.Contains(detail, "uncommitted") {
		t.Errorf("skew detail does not name the uncommitted-update cause: %s", detail)
	}

	for _, lang := range []string{"ko", "ja", "zh", "en"} {
		warn := roleRuleLocaleFor(lang).InjectionFailed("factory-lane", "role rule file absent: x")
		if !strings.Contains(warn, "moai worktree sync") {
			t.Errorf("locale %q InjectionFailed does not name the worktree refresh verb: %s", lang, warn)
		}
		if !strings.Contains(warn, "moai update") {
			t.Errorf("locale %q InjectionFailed lost the primary-checkout update guidance: %s", lang, warn)
		}
	}
}
