package security

import (
	"os"
	"path/filepath"
	"testing"
)

// SPEC-SEC-SCAN-SURFACE-001 M1 — the config-derived covered-language set.
//
// These tests observe the derivation itself, in the package that owns it. The
// end-to-end scan-count consequences of the same derivation are observed in
// internal/hook (see pre_tool_scan_config_test.go).

// coverageRepoRoot walks up to the repository root (the directory carrying go.mod).
func coverageRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found above %s", dir)
		}
		dir = parent
	}
}

// coverageCopyTree copies a directory tree, creating parents as needed.
func coverageCopyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(src, path)
		if relErr != nil {
			return relErr
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if mkErr := os.MkdirAll(filepath.Dir(target), 0o755); mkErr != nil {
			return mkErr
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatalf("copy tree %s -> %s: %v", src, dst, err)
	}
}

// coverageShippedRoot builds a temp project root carrying a copy of the ruleset
// the template distributes. Tests never point at the developer's own
// .moai/config/astgrep-rules, which is local-only and dogfood-grade.
func coverageShippedRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	src := filepath.Join(coverageRepoRoot(t), "internal", "template", "templates", ".moai", "config", "astgrep-rules")
	coverageCopyTree(t, src, filepath.Join(root, ".moai", "config", "astgrep-rules"))
	return root
}

// writeSGConfig overwrites the sgconfig.yml of a project root's shipped ruleset.
func writeSGConfig(t *testing.T, root, body string) {
	t.Helper()
	path := filepath.Join(root, ".moai", "config", "astgrep-rules", "sgconfig.yml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write sgconfig: %v", err)
	}
}

// TestCoveredLanguagesFollowConfig closes AC-SSS-006: the covered-language set
// is derived from the resolved configuration, not from a list in the code. A
// hardcoded language list could not produce the covered/uncovered split below.
func TestCoveredLanguagesFollowConfig(t *testing.T) {
	t.Run("shipped ruleset covers go", func(t *testing.T) {
		cov := NewRuleManager().ResolveCoverage(coverageShippedRoot(t))
		if cov.ConfigPath == "" {
			t.Fatal("expected the shipped ruleset's sgconfig.yml to resolve")
		}
		if !cov.Known {
			t.Fatal("expected a known covered-language set, got unknown")
		}
		if !cov.Covers("go") {
			t.Errorf("expected go to be covered, set = %v", cov.Languages)
		}
	})

	t.Run("ruleDirs without the go rules does not cover go", func(t *testing.T) {
		root := coverageShippedRoot(t)
		// Add a directory carrying a single python rule, then point ruleDirs at
		// it alone: no directory naming a `language: go` rule remains.
		pyDir := filepath.Join(root, ".moai", "config", "astgrep-rules", "pyonly")
		if err := os.MkdirAll(pyDir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		rule := "id: py-only-probe\nlanguage: python\nseverity: error\nmessage: probe\nrule:\n  pattern: os.system($CMD)\n"
		if err := os.WriteFile(filepath.Join(pyDir, "probe.yml"), []byte(rule), 0o644); err != nil {
			t.Fatalf("write rule: %v", err)
		}
		writeSGConfig(t, root, "ruleDirs:\n  - pyonly\n")

		cov := NewRuleManager().ResolveCoverage(root)
		if !cov.Known {
			t.Fatal("expected a known covered-language set, got unknown")
		}
		if cov.Covers("go") {
			t.Errorf("expected go NOT to be covered, set = %v", cov.Languages)
		}
		if !cov.Covers("python") {
			t.Errorf("expected python to be covered, set = %v", cov.Languages)
		}
	})
}

// TestUnreadableOrEmptyConfigEscalates closes AC-SSS-007: an unreadable,
// unwalkable, or empty derivation is reported as UNKNOWN so the gate escalates
// to `sg` rather than skipping. Absence of evidence is not evidence of absence.
func TestUnreadableOrEmptyConfigEscalates(t *testing.T) {
	t.Run("malformed yaml", func(t *testing.T) {
		root := coverageShippedRoot(t)
		writeSGConfig(t, root, "ruleDirs: [go\n  : : not yaml\n")
		cov := NewRuleManager().ResolveCoverage(root)
		if cov.ConfigPath == "" {
			t.Fatal("expected the config file to still resolve")
		}
		if cov.Known {
			t.Errorf("malformed config must be unknown, got known set %v", cov.Languages)
		}
	})

	t.Run("ruleDirs names a missing directory", func(t *testing.T) {
		root := coverageShippedRoot(t)
		writeSGConfig(t, root, "ruleDirs:\n  - does-not-exist\n")
		cov := NewRuleManager().ResolveCoverage(root)
		if cov.Known {
			t.Errorf("missing ruleDir must be unknown, got known set %v", cov.Languages)
		}
	})

	t.Run("rule files declare no language", func(t *testing.T) {
		root := coverageShippedRoot(t)
		dir := filepath.Join(root, ".moai", "config", "astgrep-rules", "nolang")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		rule := "id: no-language-probe\nseverity: error\nmessage: probe\nrule:\n  pattern: probe($X)\n"
		if err := os.WriteFile(filepath.Join(dir, "probe.yml"), []byte(rule), 0o644); err != nil {
			t.Fatalf("write rule: %v", err)
		}
		writeSGConfig(t, root, "ruleDirs:\n  - nolang\n")
		cov := NewRuleManager().ResolveCoverage(root)
		if cov.Known {
			t.Errorf("empty derived set must be unknown, got known set %v", cov.Languages)
		}
	})
}
