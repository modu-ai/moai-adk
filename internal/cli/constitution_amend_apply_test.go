package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// SPEC-CON-AMEND-APPLY-001 — CLI-level acceptance tests (AC-CAA-015, the CLI
// case of AC-CAA-022, the CLI case of AC-CAA-024). Every test sets
// MOAI_CONSTITUTION_REGISTRY and CLAUDE_PROJECT_DIR itself (REQ-CAA-015),
// calls runConstitutionAmend with --dry-run only (the non-dry-run CLI path is
// the approved Gap G5), and keeps every fixture under t.TempDir().

const (
	amendRuleID = "CONST-V3R6-789"
	amendBefore = "Keep every amendment small and reviewable."
	amendAfter  = "Keep every amendment small, reviewable, and reversible."
)

// amendRegistry renders a two-entry registry whose target entry carries clause.
func amendRegistry(clause string) string {
	return "# Zone Registry (fixture)\n\n```yaml\n" +
		"- id: CONST-V3R6-787\n  zone: Frozen\n  file: rules/other.md\n  anchor: \"#fixture\"\n" +
		"  clause: \"Frozen text stays as written.\"\n  canary_gate: false\n\n" +
		"- id: " + amendRuleID + "\n  zone: Evolvable\n  file: rules/target.md\n  anchor: \"#fixture\"\n" +
		"  clause: \"" + clause + "\"\n  canary_gate: false\n" +
		"```\n"
}

// writeAmendFile writes content to path, creating parent directories.
func writeAmendFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeAmendProject writes a fixture project under dir with the given target
// rule body and returns the registry path.
func writeAmendProject(t *testing.T, dir, ruleBody string) string {
	t.Helper()
	registry := filepath.Join(dir, ".claude", "rules", "moai", "core", "zone-registry.md")
	writeAmendFile(t, registry, amendRegistry(amendBefore))
	writeAmendFile(t, filepath.Join(dir, "rules", "other.md"), "# Other\n\nFrozen text stays as written.\n")
	writeAmendFile(t, filepath.Join(dir, "rules", "target.md"), ruleBody)
	return registry
}

// amendSnapshot maps every path under root to its sha256, "dir", or its
// symbolic-link target.
func amendSnapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	m := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		switch {
		case d.Type()&fs.ModeSymlink != 0:
			target, lerr := os.Readlink(p)
			if lerr != nil {
				return lerr
			}
			m[rel] = "link:" + target
		case d.IsDir():
			m[rel] = "dir"
		default:
			data, rerr := os.ReadFile(p)
			if rerr != nil {
				return rerr
			}
			sum := sha256.Sum256(data)
			m[rel] = hex.EncodeToString(sum[:])
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// assertAmendSnapshot fails when root's snapshot differs from before.
func assertAmendSnapshot(t *testing.T, root string, before map[string]string) {
	t.Helper()
	after := amendSnapshot(t, root)
	if len(after) != len(before) {
		t.Errorf("tree %s: %d paths before, %d after", root, len(before), len(after))
	}
	for k, v := range before {
		if after[k] != v {
			t.Errorf("tree %s: path %s changed", root, k)
		}
	}
}

// AC-CAA-015 — the CLI dry-run surfaces the validation failure.
func TestConstitutionAmend_DryRun_SurfacesValidation(t *testing.T) {
	t.Run("two_occurrences", func(t *testing.T) {
		t.Setenv("MOAI_CONSTITUTION_REGISTRY", "")
		t.Setenv("CLAUDE_PROJECT_DIR", "")
		dir := t.TempDir()
		writeAmendProject(t, dir, "# Target\n\n"+amendBefore+"\n\nAgain: "+amendBefore+"\n")
		before := amendSnapshot(t, dir)
		var stdout, stderr bytes.Buffer

		err := runConstitutionAmend(&stdout, &stderr, dir, amendRuleID, amendBefore, amendAfter, "", true)
		if err == nil {
			t.Fatal("runConstitutionAmend: want an error on a two-occurrence fixture, got nil")
		}
		if strings.Contains(stdout.String(), "Dry-run success") {
			t.Errorf("stdout carries the dry-run success line:\n%s", stdout.String())
		}
		assertAmendSnapshot(t, dir, before)
	})
	t.Run("valid", func(t *testing.T) {
		t.Setenv("MOAI_CONSTITUTION_REGISTRY", "")
		t.Setenv("CLAUDE_PROJECT_DIR", "")
		dir := t.TempDir()
		writeAmendProject(t, dir, "# Target\n\n"+amendBefore+"\n")
		before := amendSnapshot(t, dir)
		var stdout, stderr bytes.Buffer

		if err := runConstitutionAmend(&stdout, &stderr, dir, amendRuleID, amendBefore, amendAfter, "", true); err != nil {
			t.Fatalf("runConstitutionAmend on the valid fixture: %v", err)
		}
		if !strings.Contains(stdout.String(), "Dry-run success") {
			t.Errorf("stdout lacks the dry-run success line:\n%s", stdout.String())
		}
		assertAmendSnapshot(t, dir, before)
	})
}

// AC-CAA-022 (CLI case) — the CLI resolver returns the shared resolver's path.
func TestResolveRegistryPath_MatchesExecute(t *testing.T) {
	P, Q := t.TempDir(), t.TempDir()
	alt := filepath.Join(P, "alt", "zone-registry.md")
	t.Setenv("MOAI_CONSTITUTION_REGISTRY", alt)
	t.Setenv("CLAUDE_PROJECT_DIR", Q)

	got := resolveRegistryPath(P)
	if want := constitution.ResolveRegistryPath(P); got != want {
		t.Errorf("resolveRegistryPath(P) = %q, shared resolver = %q", got, want)
	}
	if got != alt {
		t.Errorf("resolveRegistryPath(P) = %q, want the override %q", got, alt)
	}
}

// AC-CAA-024 (CLI case) — the CLI's registry validation refuses a relative
// CLAUDE_PROJECT_DIR escape with the same check Execute runs, before the
// pipeline is reached.
func TestConstitutionAmend_ContainmentCheck_RelativeEnvEscape(t *testing.T) {
	B, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	P := filepath.Join(B, "root")
	writeAmendProject(t, P, "# Target\n\n"+amendBefore+"\n")
	offending := filepath.Join(B, "other", ".claude", "rules", "moai", "core", "zone-registry.md")
	// The copy gives the target a different clause, so a CLI that read it
	// without the check would stop at its own --before comparison instead.
	writeAmendFile(t, offending, amendRegistry("A different clause outside the root."))
	t.Chdir(P)
	t.Setenv("MOAI_CONSTITUTION_REGISTRY", "")
	t.Setenv("CLAUDE_PROJECT_DIR", filepath.Join("..", "other"))
	before := amendSnapshot(t, B)
	var stdout, stderr bytes.Buffer

	err = runConstitutionAmend(&stdout, &stderr, P, amendRuleID, amendBefore, amendAfter, "", true)
	if err == nil {
		t.Fatal("runConstitutionAmend: want a containment refusal, got nil")
	}
	msg := err.Error()
	resolved, _ := filepath.EvalSymlinks(offending)
	if !strings.Contains(msg, offending) && (resolved == "" || !strings.Contains(msg, resolved)) {
		t.Errorf("error %q does not name the offending registry %s", msg, offending)
	}
	for _, bad := range []string{"amendment failed", "clause mismatch"} {
		if strings.Contains(msg, bad) {
			t.Errorf("error %q carries %q — the refusal did not come from the CLI's own registry validation", msg, bad)
		}
	}
	if strings.Contains(stdout.String(), "Dry-run success") {
		t.Errorf("stdout carries the dry-run success line:\n%s", stdout.String())
	}
	assertAmendSnapshot(t, B, before)
}
