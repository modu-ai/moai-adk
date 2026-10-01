package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// pin_literal_sweep_test.go — the no-hardcode sweep for the audit-pin literals
// (SPEC-AGENT-TIER-001 REQ-TIER-013 / AC-TIER-005).
//
// The audit-pin model ids may appear as inline literals ONLY at their
// declared single-source locations. Per REQ-TIER-013 those locations are the
// Go constant/default declarations in internal/config (the whole package is
// the declared home — constants, defaults, and their doc comments) and the
// distributed template mirror (internal/template/templates/**). Every other
// location is a restatement and drifts.
//
// Matching: .go files are matched on the QUOTED form ("id") — a Go string
// literal; comments carry the ids as prose legitimately and do not match.
// .yaml/.templ files are matched on the bare token (their value form). In
// .yaml, values are what would restate a pin.
//
// Effort tokens ("high", "max", ...) are vocabulary values used across every
// surface legitimately and are NOT swept — the model id is the identifying
// token of a pin restatement. "glm-5.3" is matched as a substring, so it also
// catches "glm-5.3-flash" spellings: over-matching is the conservative
// direction for a detector.
//
// Files excluded BY NAME beyond the single-source locations, each with its
// documented reason (a non-pin axis that legitimately quotes the same id):
//   - template/model_policy.go — main-session model policy (ModelIDOpus55),
//     a different axis from the audit pins;
//   - statusline/memory.go — the statusline context-window table, a
//     different axis from the audit pins;
//   - cli/glm.go — GLM main-session help text (a display string naming the
//     context-window models), not an audit pin.

// pinSweepLiterals are the audit-pin model ids this sweep guards.
var pinSweepLiterals = []string{"gpt-6.1-sol", "claude-opus-5-5", "glm-5.3"}

// pinSweepExcludedFiles are non-pin-axis files excluded by name (rel path
// against the walk root), each with its documented reason.
var pinSweepExcludedFiles = []struct {
	rel    string
	reason string
}{
	{"template/model_policy.go", "main-session model policy axis (ModelIDOpus55), not an audit pin"},
	{"statusline/memory.go", "statusline context-window table axis, not an audit pin"},
	{"cli/glm.go", "GLM main-session help-text display string, not an audit pin"},
}

// sweepPinLiterals walks root for .go/.yaml/.templ files and reports every
// file whose content carries a pin literal outside the declared
// single-source locations. It returns the number of files actually swept
// (an empty sweep asserts nothing — the caller must check it) and one
// violation entry per offending file.
func sweepPinLiterals(root string) (swept int, violations []string) {
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		relSlash := filepath.ToSlash(rel)
		if d.IsDir() {
			// The template mirror — declared single-source location. Matched
			// on the rel prefix so the walk root spelling does not matter.
			if relSlash == "template/templates" || strings.HasPrefix(relSlash, "template/templates/") {
				return filepath.SkipAll
			}
			return nil
		}
		name := d.Name()
		if strings.HasSuffix(name, "_test.go") {
			return nil // verification instruments restate expected values
		}
		ext := filepath.Ext(name)
		if ext != ".go" && ext != ".yaml" && ext != ".templ" {
			return nil
		}
		for _, ex := range pinSweepExcludedFiles {
			if relSlash == ex.rel {
				return nil // documented non-pin axis
			}
		}
		if strings.HasPrefix(relSlash, "config/") {
			return nil // internal/config — the declared single-source package
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		swept++
		content := string(data)
		for _, lit := range pinSweepLiterals {
			needle := lit
			if ext == ".go" {
				needle = `"` + lit + `"` // Go string-literal form; comments do not match
			}
			if strings.Contains(content, needle) {
				violations = append(violations, relSlash+" carries pin literal "+lit)
				break
			}
		}
		return nil
	})
	if err != nil {
		return swept, append(violations, "walk error: "+err.Error())
	}
	return swept, violations
}

// TestPinLiteralSweep_PositiveControl demonstrates the sweep RED on a planted
// violation (verification-completeness.md §1.1 observed-failure completion):
// a scratch file carrying a pin literal inside the swept scope must be
// reported. The planted file lives in a temp tree and is removed with it;
// the two decoys (a test file and a mirror-path file) must NOT be reported.
func TestPinLiteralSweep_PositiveControl(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "pkg"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "pkg", "planted.go"),
		[]byte("package pkg\n\nconst x = \"gpt-6.1-sol\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "pkg", "planted_test.go"),
		[]byte("package pkg\n\nconst y = \"gpt-6.1-sol\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "template", "templates", "s"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "template", "templates", "s", "m.yaml"),
		[]byte("model: glm-5.3\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	swept, violations := sweepPinLiterals(root)
	if len(violations) > 0 && violations[0] == "" {
		t.Fatal("unreachable")
	}
	if swept == 0 {
		t.Fatalf("positive control swept zero files (walk diagnostics: %v)", violations)
	}
	if len(violations) != 1 {
		t.Fatalf("positive control: want exactly 1 violation (the planted non-test file), got %d: %v", len(violations), violations)
	}
	if !strings.Contains(violations[0], "planted.go") || !strings.Contains(violations[0], "gpt-6.1-sol") {
		t.Errorf("positive control violation misattributed: %q", violations[0])
	}
}

// TestPinLiteralSweep_RealTreeClean runs the sweep on the repository's
// internal/ tree: every pin literal must live at a declared single-source
// location. The swept file count is reported so an empty sweep can never read
// as a pass (verification-completeness.md §1.1).
func TestPinLiteralSweep_RealTreeClean(t *testing.T) {
	root := ".." // internal/ — the test binary runs in internal/config
	swept, violations := sweepPinLiterals(root)
	t.Logf("pin literal sweep: swept %d files under internal/ (excl. tests, internal/config, template mirror, documented non-pin axes)", swept)
	if swept < 100 {
		t.Fatalf("swept file count %d is implausibly small — the sweep root is wrong (diagnostics: %v)", swept, violations)
	}
	if len(violations) != 0 {
		t.Errorf("pin literals found outside the declared single-source locations (%d):", len(violations))
		for _, v := range violations {
			t.Errorf("  %s", v)
		}
	}
}
