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
	"github.com/modu-ai/moai-adk/pkg/version"
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

// TestRoleRulesSkewedTreeKeepsDeployedCoreWithNotice pins the post-round-3
// shape: version stamps cannot establish build order (an explicit rc
// VERSION reads HIGHER — version-management.md:97), so a skewed tree whose
// deployed files still parse KEEPS the deployed core; the skew is surfaced
// as a one-line notice naming both versions instead of a swap.
func TestRoleRulesSkewedTreeKeepsDeployedCoreWithNotice(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	stubEmbeddedRule(t, markedEmbeddedRule(), true)

	root := t.TempDir()
	writeRoleRuleFixture(t, root, roleRuleFiles[0], smallMarkedRule("STALE-OR-NEWER-DEPLOYED"))
	writeRoleRuleFixture(t, root, roleRuleFiles[1], smallMarkedRule("messaging"))
	writeSystemYAMLFixture(t, root, "moai:\n  template_version: v0.0.0-fixture-stamp\n")

	inj := roleRuleInjectionFor(root, "startup", langEnglish)
	if inj.OperatorNotice != "" {
		t.Fatalf("skew alone must not fail the session; notice: %s", inj.OperatorNotice)
	}
	if !strings.Contains(inj.Context, "STALE-OR-NEWER-DEPLOYED") {
		t.Fatalf("deployed core was replaced on an unprovable-premise skew:\n%s", inj.Context)
	}
	if strings.Contains(inj.Context, embeddedFallbackCoreText) {
		t.Fatalf("embedded copy replaced a parseable deployment:\n%s", inj.Context)
	}
	if !strings.Contains(inj.Context, "v0.0.0-fixture-stamp") || !strings.Contains(inj.Context, version.GetVersion()) {
		t.Fatalf("skew notice does not name both versions:\n%s", inj.Context)
	}
	t.Logf("PASS: skewed tree keeps the deployed core and surfaces the skew in one line")
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

// TestRoleRulesNewerOrUnparseableStampKeepsDeployed pins the directional
// skew (audit round-2 P2): only a stamp genuinely OLDER than the binary
// swaps to the embedded copy. A NEWER stamp (a newer binary deployed the
// project; an older binary runs now) or an unorderable free-form stamp
// keeps the deployed core — the swap must never regress a newer deployment.
func TestRoleRulesNewerOrUnparseableStampKeepsDeployed(t *testing.T) {
	for _, tc := range []struct{ name, stamp string }{
		{"newer_stamp", "v9.9.9"},
		{"unorderable_stamp", "cutover/github-flow"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clearFactoryEnv(t)
			t.Setenv(config.EnvMoaiFactoryWorkers, "1")
			stubEmbeddedRule(t, markedEmbeddedRule(), true)

			root := t.TempDir()
			writeRoleRuleFixture(t, root, roleRuleFiles[0], smallMarkedRule("DEPLOYED-CORE"))
			writeRoleRuleFixture(t, root, roleRuleFiles[1], smallMarkedRule("messaging"))
			writeSystemYAMLFixture(t, root, "moai:\n  template_version: "+tc.stamp+"\n")

			inj := roleRuleInjectionFor(root, "startup", langEnglish)
			if inj.OperatorNotice != "" {
				t.Fatalf("non-older stamp must not fail the session; notice: %s", inj.OperatorNotice)
			}
			if !strings.Contains(inj.Context, "DEPLOYED-CORE") {
				t.Fatalf("deployed core was replaced under a %s stamp:\n%s", tc.stamp, inj.Context)
			}
			if strings.Contains(inj.Context, embeddedFallbackCoreText) {
				t.Fatalf("embedded copy regressed a %s deployment:\n%s", tc.stamp, inj.Context)
			}
		})
	}
}

// TestRoleRulesCodexOnlyFallbackErrorsNameDeployedPath pins the round-4
// path-normalization fix: on a Codex-only install the rule deploys under
// .moai/policies — every operator-facing error from the embedded fallback
// names THAT path, never the Claude-layout template path the embed carries.
func TestRoleRulesCodexOnlyFallbackErrorsNameDeployedPath(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	// The embedded copy itself is unusable, so the fallback fails and the
	// REQ-ALB-009 detail renders the label.
	stubEmbeddedRule(t, "# embedded copy with no markers\n", true)

	root := t.TempDir()
	altDispatch := ".moai/policies/workflow/factory-dispatch.md"
	altMessaging := ".moai/policies/workflow/cross-session-messaging.md"
	writeRoleRuleFixture(t, root, roleRuleFile{Rel: altDispatch, Name: "factory-dispatch.md"},
		"# stale codex-only dispatch rule\n\nno role-core markers here\n")
	writeRoleRuleFixture(t, root, roleRuleFile{Rel: altMessaging, Name: "cross-session-messaging.md"},
		smallMarkedRule("messaging"))

	inj := roleRuleInjectionFor(root, "startup", langEnglish)
	if inj.OperatorNotice == "" {
		t.Fatal("unusable embedded copy must keep the failure path")
	}
	if !strings.Contains(inj.OperatorNotice, altDispatch) {
		t.Fatalf("failure detail does not name the deployed Codex-only path: %s", inj.OperatorNotice)
	}
	if strings.Contains(inj.OperatorNotice, roleRuleFiles[0].Rel) {
		t.Fatalf("failure detail leaked the Claude-layout template path on a Codex-only install: %s", inj.OperatorNotice)
	}
}

// TestRoleRulesCodexOnlyFallbackProjectsBodyPaths pins the round-5 body
// normalization: the embed is the Claude-layout text — on a Codex-only
// install the fallback core's PATH REFERENCES are projected onto the
// deployed .moai/policies layout, the same way the deployer's harness
// normalization rewrites them at deploy time. A core pointing agents at
// `.claude/rules/...` on an install that never deploys that layout is a
// broken required read.
func TestRoleRulesCodexOnlyFallbackProjectsBodyPaths(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	stubEmbeddedRule(t,
		"# Fixture rule (embedded copy)\n\n"+
			config.RoleCoreMarkerStart+"\n"+embeddedFallbackCoreText+"\n"+
			"Read the full body at `.claude/rules/moai/workflow/cross-session-messaging.md` before acting.\n"+
			config.RoleCoreMarkerEnd+"\n\nplain tail\n",
		true)

	root := t.TempDir()
	writeRoleRuleFixture(t, root, roleRuleFile{Rel: ".moai/policies/workflow/factory-dispatch.md", Name: "factory-dispatch.md"},
		"# stale codex-only dispatch rule\n\nno role-core markers here\n")
	writeRoleRuleFixture(t, root, roleRuleFile{Rel: ".moai/policies/workflow/cross-session-messaging.md", Name: "cross-session-messaging.md"},
		smallMarkedRule("messaging"))

	inj := roleRuleInjectionFor(root, "startup", langEnglish)
	if inj.OperatorNotice != "" {
		t.Fatalf("fallback must succeed; notice: %s", inj.OperatorNotice)
	}
	if !strings.Contains(inj.Context, ".moai/policies/workflow/cross-session-messaging.md") {
		t.Fatalf("fallback core body was not projected onto the deployed Codex-only layout:\n%s", inj.Context)
	}
	if strings.Contains(inj.Context, ".claude/rules/moai/") {
		t.Fatalf("fallback core body carries a Claude-layout path the install never deploys:\n%s", inj.Context)
	}
}

// TestRoleRulesEndMarkerOnlyStaysFailVisible pins the audit round-1 P2 fix:
// a file carrying ONLY an end marker is corruption, not a stale generation —
// the embedded fallback must NOT rescue it; the REQ-ALB-009 pair fires.
func TestRoleRulesEndMarkerOnlyStaysFailVisible(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	stubEmbeddedRule(t, markedEmbeddedRule(), true)

	root := t.TempDir()
	writeRoleRuleFixture(t, root, roleRuleFiles[0],
		"<!-- moai:role-core-end -->\n\ncorrupted: an end marker with no start\n")
	writeRoleRuleFixture(t, root, roleRuleFiles[1], smallMarkedRule("messaging"))

	inj := roleRuleInjectionFor(root, "startup", langEnglish)
	assertFailVisible(t, inj, "end-marker-only")
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
	if !strings.Contains(detail, "git merge") {
		t.Errorf("skew detail does not name the working refresh (merge the base branch into the worktree): %s", detail)
	}
	if !strings.Contains(detail, "uncommitted") {
		t.Errorf("skew detail does not name the uncommitted-update cause: %s", detail)
	}
	if strings.Contains(detail, "commit it there") {
		t.Errorf("skew detail still prescribes a base-branch commit that commit-protected repos refuse (audit round-4 P2): %s", detail)
	}

	for _, lang := range []string{"ko", "ja", "zh", "en"} {
		warn := roleRuleLocaleFor(lang).InjectionFailed("factory-lane", "role rule file absent: x", "main")
		if !strings.Contains(warn, "git merge") {
			t.Errorf("locale %q InjectionFailed does not name the worktree refresh (merge the base branch): %s", lang, warn)
		}
		if !strings.Contains(warn, "moai update") {
			t.Errorf("locale %q InjectionFailed lost the primary-checkout update guidance: %s", lang, warn)
		}
	}
}

// TestRoleRulesRemediationNamesConfiguredBaseBranch pins the round-3 base
// fix: a project running a non-main base (git_strategy.worktree_base_branch)
// gets refresh advice naming ITS branch, not a hardcoded main.
func TestRoleRulesRemediationNamesConfiguredBaseBranch(t *testing.T) {
	root := t.TempDir()
	writeSystemYAMLFixture(t, root, "moai:\n  template_version: v0.0.0-fixture-stamp\n")
	cfgDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "git-strategy.yaml"), []byte("git_strategy:\n  worktree_base_branch: develop\n"), 0o644); err != nil {
		t.Fatalf("write git-strategy fixture: %v", err)
	}

	detail := roleRuleSkewDetail(root)
	if !strings.Contains(detail, "git merge develop") {
		t.Errorf("skew detail does not name the configured base branch: %s", detail)
	}
	warn := roleRuleLocaleFor("en").InjectionFailed("factory-lane", "detail", "develop")
	if !strings.Contains(warn, "git merge develop") {
		t.Errorf("locale guidance does not render the configured base branch: %s", warn)
	}
}
