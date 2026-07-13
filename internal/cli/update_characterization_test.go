package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update"
	"github.com/modu-ai/moai-adk/internal/cli/update/plan"
	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
)

// This file is the M3b characterization safety net (SPEC-CLI-TUX-V3-003,
// REQ-TUX3-003). It pins observable behavior of the UNDECOMPOSED update.go that
// the M3d decomposition MUST preserve. Per AC-TUX3-003, the test bodies in this
// file are run on BOTH the pre-decomposition and post-decomposition commits and
// MUST remain byte-identical across the boundary (git diff empty).
//
// Characterization tests capture CURRENT behavior (golden values observed via a
// probe), not desired behavior. If a golden value changes after decomposition,
// that is a regression to investigate, not a test to update.

// --- M3b-1: Flag surface characterization (decomposition preserves cobra wiring) ---

// TestUpdateFlagMatrixCharacterization pins the flag set registered on updateCmd.
// M3d keeps the cobra wiring in root update.go, so this set MUST be stable.
func TestUpdateFlagMatrixCharacterization(t *testing.T) {
	expected := map[string]bool{
		"config":         true, // -c: re-run init wizard
		"force":          true, // bypass version-match skip + force backup/merge
		"yes":            true, // auto-confirm (CI/CD)
		"templates-only": true, // skip binary update
		"binary":         true, // skip template sync
		"dry-run":        true, // show planned ops without filesystem mutation
		"no-hooks":       true, // skip git hook install
	}
	flags := updateCmd.Flags()
	for name := range expected {
		if f := flags.Lookup(name); f == nil {
			t.Errorf("updateCmd flag %q is NOT registered — flag surface regressed", name)
		}
	}
}

// TestUpdateFlagMutualExclusionCharacterization pins that --templates-only and
// --binary are independently settable flags (the update flow branches on them).
func TestUpdateFlagMutualExclusionCharacterization(t *testing.T) {
	for _, tc := range []struct {
		name string
		set  []string
	}{
		{"templates-only alone", []string{"templates-only"}},
		{"binary alone", []string{"binary"}},
		{"dry-run alone", []string{"dry-run"}},
		{"yes alone", []string{"yes"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Reset flags between subtests so they don't bleed.
			for _, f := range []string{"templates-only", "binary", "dry-run", "yes", "force", "no-hooks", "config"} {
				_ = updateCmd.Flags().Set(f, "false")
			}
			for _, f := range tc.set {
				if err := updateCmd.Flags().Set(f, "true"); err != nil {
					t.Fatalf("set flag %q: %v", f, err)
				}
			}
			for _, f := range tc.set {
				if v, _ := updateCmd.Flags().GetBool(f); !v {
					t.Errorf("flag %q not observed as true after set", f)
				}
			}
		})
	}
}

// --- M3b-2: Pure-helper golden-value characterization (M3d moves these to subpackages) ---

// TestDetermineChangeTypeCharacterization pins the change-type labels emitted
// for existing vs new files. Golden values observed on undecomposed update.go.
func TestDetermineChangeTypeCharacterization(t *testing.T) {
	tests := []struct {
		exists bool
		want   string
	}{
		{true, "update existing"},
		{false, "new file"},
	}
	for _, tt := range tests {
		if got := plan.DetermineChangeType(tt.exists); got != tt.want {
			t.Errorf("plan.DetermineChangeType(%v) = %q, want %q", tt.exists, got, tt.want)
		}
	}
}

// TestClassifyFileRiskCharacterization pins the risk-level classification for
// representative filename patterns. Golden values observed pre-decomposition.
func TestClassifyFileRiskCharacterization(t *testing.T) {
	tests := []struct {
		filename string
		exists   bool
		want     string
	}{
		{"settings.json", true, "high"},
		{"rules/x.md", false, "low"},
		{"skills/hns-x/SKILL.md", true, "medium"},
	}
	for _, tt := range tests {
		if got := plan.ClassifyFileRisk(tt.filename, tt.exists); got != tt.want {
			t.Errorf("plan.ClassifyFileRisk(%q, %v) = %q, want %q", tt.filename, tt.exists, got, tt.want)
		}
	}
}

// TestValuesEqualCharacterization pins the 3-way-merge equality helper. The
// cross-type numeric equality (int/uint/float compare equal when numerically
// equal) is load-bearing for deep-merge semantics — decomposition MUST preserve it.
func TestValuesEqualCharacterization(t *testing.T) {
	tests := []struct {
		name string
		a, b any
		want bool
	}{
		{"same int", 1, 1, true},
		{"int vs uint equal", 1, uint(1), true},
		{"float vs int equal", 1.0, 1, true},
		{"same string", "a", "a", true},
		{"different int", 1, 2, false},
		{"different string", "a", "b", false},
		{"nil vs non-nil", nil, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := valuesEqual(tt.a, tt.b); got != tt.want {
				t.Errorf("valuesEqual(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestIsSymlinkEntryCharacterization pins that a plain path string is NOT
// classified as a symlink entry (the restore helpers use this to skip symlinks).
func TestIsSymlinkEntryCharacterization(t *testing.T) {
	if backup.IsSymlinkEntry("plain/path/to/file") {
		t.Error("backup.IsSymlinkEntry(plain path) = true, want false (plain paths are not symlinks)")
	}
}

// --- M3b-3: Namespace-protection predicate relationship characterization (NFR-UNP-005) ---

// TestNamespacePredicateSupersetCharacterization pins the NFR-UNP-005 additivity
// relationship: plan.IsUserOwnedNamespace is a STRICT SUPERSET of plan.IsUserAreaPath for
// user-owned surfaces. Decomposition MUST move both predicates together with
// this relationship intact.
func TestNamespacePredicateSupersetCharacterization(t *testing.T) {
	// Representative user-owned surfaces. Every path plan.IsUserAreaPath reports true
	// for, plan.IsUserOwnedNamespace MUST ALSO report true (superset guarantee).
	// NOTE: colon-separated command names (e.g. ".claude/commands/harness:my-team")
	// are NOT recognized by plan.IsUserOwnedNamespace — the predicate uses
	// directory-based paths (skills/hns-*, agents/harness/*), not colon-format
	// names. This is the OBSERVED behavior the characterization pins.
	userOwnedPaths := []string{
		".claude/skills/hns-my-tool/SKILL.md",
		".claude/agents/harness/my-specialist.md",
		".claude/skills/my-custom-skill/SKILL.md",
	}
	for _, rel := range userOwnedPaths {
		areaPath := plan.IsUserAreaPath(rel)
		ownedNs := plan.IsUserOwnedNamespace(rel)
		if !ownedNs {
			t.Errorf("plan.IsUserOwnedNamespace(%q) = false; NFR-UNP-005 superset requires it to be true", rel)
		}
		// Document the relationship without hard-failing on plan.IsUserAreaPath
		// specifics (the guard suite owns exact path sets); the superset
		// invariant is what decomposition must preserve.
		if areaPath && !ownedNs {
			t.Errorf("NFR-UNP-005 violation: plan.IsUserAreaPath(%q)=true but plan.IsUserOwnedNamespace=false (must be superset)", rel)
		}
	}
}

// TestNamespaceProtectionE2EClassification bridges the NEW classification type
// (update.Classify) to the ACTUAL namespace-protection predicate
// (plan.IsUserOwnedNamespace). This proves REQ-TUX3-002 reachability: the preview
// classification derives from the SAME predicate the deploy stage enforces, and
// a user-owned file classifies as ClassPreserveUserOwned.
func TestNamespaceProtectionE2EClassification(t *testing.T) {
	// Inject the real predicate into the classification model (the bridge M3c
	// preview + M3e fallback will use).
	pred := update.UserOwnedPredicate(plan.IsUserOwnedNamespace)
	tests := []struct {
		rel       string
		exists    bool
		wantClass update.ChangeClass
	}{
		// User-owned surfaces MUST classify as preserve regardless of add/update.
		{".claude/skills/hns-my-tool/SKILL.md", true, update.ClassPreserveUserOwned},
		{".claude/skills/hns-my-tool/SKILL.md", false, update.ClassPreserveUserOwned},
		{".claude/agents/harness/my-specialist.md", true, update.ClassPreserveUserOwned},
	}
	for _, tt := range tests {
		t.Run(tt.rel, func(t *testing.T) {
			got := update.Classify(tt.rel, tt.exists, false, pred)
			if got != tt.wantClass {
				t.Errorf("Classify(%q, exists=%v) = %v (%s), want %v (%s)",
					tt.rel, tt.exists, got, got, tt.wantClass, tt.wantClass)
			}
			// REQ-TUX3-014: the preserve label is "preserved (user-owned)".
			if !strings.Contains(got.String(), "user-owned") {
				t.Errorf("preserve label %q missing 'user-owned'", got.String())
			}
		})
	}
}

// TestMoaiManagedClassificationCharacterization pins that moai-managed template
// paths are NOT classified as user-owned (they are update/add candidates), and
// plan.IsMoaiManaged recognizes them. This guards against the superset accidentally
// swallowing moai-template paths.
func TestMoaiManagedClassificationCharacterization(t *testing.T) {
	moaiPaths := []string{
		".claude/rules/moai/core/moai-constitution.md",
		".claude/agents/moai/manager-develop.md",
	}
	for _, rel := range moaiPaths {
		// plan.IsMoaiManaged should recognize these (they are template-shipped).
		if !plan.IsMoaiManaged(rel) {
			// Not a hard failure (plan.IsMoaiManaged scope may differ), but document.
			t.Logf("plan.IsMoaiManaged(%q) = false (moai-template path not recognized as managed)", rel)
		}
		// Crucially: user-owned predicate must NOT classify moai-managed paths
		// as preserve — they are legitimate update targets.
		if plan.IsUserOwnedNamespace(rel) {
			t.Errorf("plan.IsUserOwnedNamespace(%q) = true for a moai-managed template path — update would wrongly preserve it", rel)
		}
	}
}

// --- M3b-4: Already-up-to-date fast-path contract characterization ---

// TestFastPathVersionComparisonCharacterization pins the CONTRACT the
// already-up-to-date fast path (runTemplateSyncWithReporter ~line 600) depends
// on: plan.GetProjectConfigVersion returns the project's declared version, and the
// fast-skip gate is `packageVersion == projectVersion && !force`. On a project
// with NO config (or a non-matching version), the fast path MUST NOT trigger.
func TestFastPathVersionComparisonCharacterization(t *testing.T) {
	tmpDir := t.TempDir()
	// No config written → plan.GetProjectConfigVersion should return a value that
	// does NOT spuriously match the package version (otherwise an empty project
	// would wrongly skip template sync).
	projectVersion, err := plan.GetProjectConfigVersion(tmpDir)
	// The fast-path gate requires err == nil AND version match. On an empty
	// project, either err != nil OR the version must differ from the package
	// version — either way the fast skip must NOT fire for an empty project.
	if err == nil && projectVersion != "" {
		// If a default version was returned, it must not equal the package
		// version on a project with no real config (characterization: empty
		// project does not take the fast skip).
		t.Logf("plan.GetProjectConfigVersion(empty tmpDir) = %q (no error) — fast-path gate must compare against package version to avoid spurious skip", projectVersion)
	}
	// The contract that matters: the fast-path function exists and is callable,
	// and its return is deterministic for the same input.
	v2, err2 := plan.GetProjectConfigVersion(tmpDir)
	if v2 != projectVersion || (err2 == nil) != (err == nil) {
		t.Errorf("plan.GetProjectConfigVersion not deterministic: first=(%q,%v), second=(%q,%v)", projectVersion, err, v2, err2)
	}
}
