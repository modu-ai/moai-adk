package mx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAC004_Characterization_PathBodyUnchanged is a characterization test
// (AC-MX-ASSOC-004). It captures TODAY's path-based and body-based association
// outputs on a fixed fixture tag set and asserts byte-for-byte equality (same
// length, same order, same IDs). The golden values are LITERAL []string — they
// are NOT computed at runtime, so any regression in path/body association
// surfaces as a test failure rather than a self-consistent re-derivation.
//
// Every fixture tag carries SpecRef == "" (no @MX:SPEC sub-line), so the
// additive sub-line source introduced by this SPEC MUST NOT alter these
// outputs (REQ-MX-ASSOC-004: path/body outputs preserved byte-for-byte for
// every tag that does not carry an @MX:SPEC sub-line).
func TestAC004_Characterization_PathBodyUnchanged(t *testing.T) {
	specModules := map[string][]string{
		"SPEC-AUTH-001": {"internal/auth/"},
		"SPEC-DB-002":   {"internal/db/", "internal/cache/"},
	}

	fixtures := []struct {
		name string
		tag  Tag
		// golden is the pre-change expected association slice (LITERAL, not
		// runtime-derived). Source order: path → body.
		golden []string
	}{
		{
			name: "path-only association",
			tag: Tag{
				Kind: MXNote,
				File: "internal/auth/handler.go",
				Line: 5,
				Body: "general note without a SPEC token",
			},
			golden: []string{"SPEC-AUTH-001"},
		},
		{
			name: "body-only association",
			tag: Tag{
				Kind: MXAnchor,
				File: "internal/elsewhere/handler.go",
				Line: 10,
				Body: "ANCHOR for SPEC-AUTH-001 handler",
			},
			golden: []string{"SPEC-AUTH-001"},
		},
		{
			name: "path+body overlap (de-dup)",
			tag: Tag{
				Kind: MXAnchor,
				File: "internal/auth/handler.go",
				Line: 10,
				Body: "ANCHOR for SPEC-AUTH-001 handler",
			},
			golden: []string{"SPEC-AUTH-001"},
		},
		{
			name: "no association",
			tag: Tag{
				Kind: MXNote,
				File: "internal/elsewhere/helper.go",
				Line: 3,
				Body: "a note with no SPEC token and no matching path",
			},
			golden: []string{},
		},
		{
			name: "WARN-with-reason (path-based)",
			tag: Tag{
				Kind:   MXWarn,
				File:   "internal/db/query.go",
				Line:   12,
				Body:   "guarded mutation",
				Reason: "concurrency hazard",
			},
			golden: []string{"SPEC-DB-002"},
		},
		{
			name: "ANCHOR (path-based, second module path)",
			tag: Tag{
				Kind:     MXAnchor,
				File:     "internal/cache/store.go",
				Line:     20,
				Body:     "cache invariant",
				AnchorID: "cache-anchor",
			},
			golden: []string{"SPEC-DB-002"},
		},
		{
			name: "DEBT (path-based)",
			tag: Tag{
				Kind:    MXDebt,
				File:    "internal/db/query.go",
				Line:    30,
				Body:    "in-memory map cache",
				RotRisk: "no-trigger",
			},
			golden: []string{"SPEC-DB-002"},
		},
		{
			name: "TODO (body-based, two SPEC IDs)",
			tag: Tag{
				Kind: MXTodo,
				File: "internal/elsewhere/work.go",
				Line: 8,
				Body: "finish SPEC-AUTH-001 and SPEC-DB-002 wiring",
			},
			golden: []string{"SPEC-AUTH-001", "SPEC-DB-002"},
		},
	}

	for _, fx := range fixtures {
		t.Run(fx.name, func(t *testing.T) {
			if fx.tag.SpecRef != "" {
				t.Fatalf("characterization fixture %q must have empty SpecRef; got %q", fx.name, fx.tag.SpecRef)
			}
			associator := NewSpecAssociator(specModules)
			got := associator.Associate(fx.tag)
			if !sameStrSlice(got, fx.golden) {
				t.Errorf("byte-for-byte mismatch:\n  got    = %v\n  golden = %v", got, fx.golden)
			}
		})
	}
}

// TestAC004_SubLineAdditiveDoesNotShrinkCharacterization verifies that adding a
// SpecRef to a characterization tag NEVER shrinks its association set relative
// to the no-SpecRef baseline (the sub-line source is additive; REQ-MX-ASSOC-002
// / REQ-MX-ASSOC-004).
func TestAC004_SubLineAdditiveDoesNotShrinkCharacterization(t *testing.T) {
	specModules := map[string][]string{
		"SPEC-AUTH-001": {"internal/auth/"},
	}
	associator := NewSpecAssociator(specModules)

	tag := Tag{
		Kind: MXNote,
		File: "internal/auth/handler.go",
		Line: 5,
		Body: "general note",
	}
	baseline := associator.Associate(tag)

	// Attach an unrelated sub-line SPEC; the path-based association MUST remain.
	tag.SpecRef = "SPEC-UNRELATED-999"
	withSubLine := associator.Associate(tag)

	if !containsStr(withSubLine, "SPEC-AUTH-001") {
		t.Errorf("additive sub-line removed path association: baseline=%v withSubLine=%v", baseline, withSubLine)
	}
	if !containsStr(withSubLine, "SPEC-UNRELATED-999") {
		t.Errorf("additive sub-line did not contribute its own ID: %v", withSubLine)
	}
}

// TestAC005_ProductionValidatorNoNewErrorVocabulary is the regression guard for
// AC-MX-ASSOC-005. The production "validator green" property is, at the Go-test
// level, the assertion that this change introduces NO new error-prefix
// vocabulary into Scanner.GetErrors. The pre-existing error vocabulary is
// {invalid tag format, unknown tag kind, DuplicateAnchorID}; the NEW sentinel
// strings this SPEC adds (DanglingSpecRef, UnresolvedSpecRef) belong in
// GetWarnings / associator diagnostics, NEVER in GetErrors. Scanning the real
// internal/mx source tree (which carries @MX prose in comments the scanner
// parses) is the stress surface; if any GetErrors entry carries a NEW prefix,
// the change is non-additive at the error channel.
//
// Note: the pre-existing prose-induced parse errors (e.g. comment lines that
// mention "@MX:UPGRADE" without a colon) are the baseline and are excluded by
// the prefix allowlist below — they are not regressions this SPEC introduces.
func TestAC005_ProductionValidatorNoNewErrorVocabulary(t *testing.T) {
	mxDir := findRepoSubdir(t, "internal/mx")
	if mxDir == "" {
		t.Skip("not running from the moai-adk-go checkout (internal/mx not found by walking up from CWD)")
	}

	scanner := NewScanner()
	if _, err := scanner.ScanDir(mxDir); err != nil {
		t.Fatalf("ScanDir internal/mx: %v", err)
	}

	allowedPrefixes := []string{
		"invalid tag format", // prose-without-colon baseline
		"unknown tag kind",   // unrecognized standalone kind baseline
		"DuplicateAnchorID",  // anchor collision baseline
	}
	// Sentinels this SPEC introduces — they MUST NOT appear in GetErrors.
	forbiddenFragments := []string{"DanglingSpecRef", "UnresolvedSpecRef"}

	for _, errStr := range scanner.GetErrors() {
		for _, frag := range forbiddenFragments {
			if strings.Contains(errStr, frag) {
				t.Errorf("NEW error vocabulary leaked into GetErrors (should be warning/diagnostic): %q", errStr)
			}
		}
		matched := false
		for _, p := range allowedPrefixes {
			if strings.HasPrefix(errStr, p) || strings.Contains(errStr, p) {
				matched = true
				break
			}
		}
		if !matched {
			t.Errorf("unrecognized (possibly NEW) error vocabulary in GetErrors: %q", errStr)
		}
	}
}

// sameStrSlice reports whether two string slices are equal, treating nil and
// the empty slice as equivalent (Associate returns nil for the no-match case).
func sameStrSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// findRepoSubdir walks up from the test's CWD looking for a directory that
// contains the given repo-relative subdir. It returns the absolute path to that
// subdir, or "" if not found within a reasonable number of parent steps.
func findRepoSubdir(t *testing.T, rel string) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := cwd
	for i := 0; i < 8; i++ {
		candidate := filepath.Join(dir, rel)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
