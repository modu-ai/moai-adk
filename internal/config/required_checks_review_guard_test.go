package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// reviewGuardRoot is the tree whose .github/required-checks.yml is read. It is the project root; the
// CCR_GUARD_ROOT environment variable names another tree and exists only as a mutant-probe seam, so the
// guard's red can be observed on a mutated COPY of the contract without touching the real file.
func reviewGuardRoot(t *testing.T) string {
	t.Helper()
	if p := os.Getenv("CCR_GUARD_ROOT"); p != "" {
		return p
	}
	return findProjectRoot(t)
}

// readReleaseGate returns the optional top-level release_gate list of the required-checks file under root.
// It reads the YAML key directly, so the guard does not depend on a RequiredChecks struct field: the key
// is absent from a contract that has no release gate, and an absent key reads as an empty list.
func readReleaseGate(t *testing.T, root string) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, RequiredChecksFile))
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		ReleaseGate []string `yaml:"release_gate"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse %s: %v", RequiredChecksFile, err)
	}
	return doc.ReleaseGate
}

// AC-CCR-017 (REQ-CCR-016, SPEC-CI-CODEX-REVIEW-001, regression-guard) — no Codex review check is a required,
// release-gate, or auxiliary check until the required-checks confirmation gate G4 (spec section 5) is passed.
// The test is permanent: removing it is the user-confirmed change of gate G4. This SPEC never edits
// .github/required-checks.yml.
//
// Mutant: adding a "Codex Review" context to branches.main.contexts or to auxiliary turns a row red.
func TestRequiredChecksCarryNoCodexReviewCheckUntilConfirmed(t *testing.T) {
	root := reviewGuardRoot(t)
	rc, err := LoadRequiredChecks(root)
	if err != nil {
		t.Fatal(err)
	}
	scan := func(t *testing.T, where string, contexts []string) {
		t.Helper()
		for _, c := range contexts {
			if strings.Contains(strings.ToLower(c), "codex") {
				t.Fatalf("%s lists %q: a Codex review check in the required-checks contract needs the user-confirmed gate G4 (spec section 5, decision-index deferred gates); this guard blocks an unconfirmed edit", where, c)
			}
		}
	}
	branch := func(t *testing.T, key string) []string {
		t.Helper()
		b, ok := rc.Branches[key]
		if !ok {
			t.Fatalf("branches has no key %q; the guard would sweep nothing", key)
		}
		return b.Contexts
	}
	t.Run("not_in_main", func(t *testing.T) { scan(t, "branches.main.contexts", branch(t, "main")) })
	t.Run("not_in_ci_candidate_key", func(t *testing.T) { scan(t, `branches."ci/**".contexts`, branch(t, "ci/**")) })
	t.Run("not_in_release_key", func(t *testing.T) { scan(t, `branches."release/*".contexts`, branch(t, "release/*")) })
	t.Run("not_in_release_gate", func(t *testing.T) { scan(t, "release_gate", readReleaseGate(t, root)) })
	t.Run("not_in_auxiliary", func(t *testing.T) { scan(t, "auxiliary", rc.Auxiliary) })
	t.Run("control_release_gate_reader_reads_the_key", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, ".github"), 0o755); err != nil {
			t.Fatal(err)
		}
		probe := "version: 1\nrelease_gate:\n  - Lint\n  - Codex Review\n"
		if err := os.WriteFile(filepath.Join(dir, RequiredChecksFile), []byte(probe), 0o644); err != nil {
			t.Fatal(err)
		}
		got := readReleaseGate(t, dir)
		if len(got) != 2 || got[1] != "Codex Review" {
			t.Fatalf("readReleaseGate returned %v for a file that lists Codex Review under release_gate; the release-gate row above would pass on a real mutation", got)
		}
	})
	t.Run("control_main_lists_lint", func(t *testing.T) {
		for _, c := range branch(t, "main") {
			if c == "Lint" {
				return
			}
		}
		t.Fatalf("branches.main.contexts does not list Lint: the loader read an empty or wrong contract, so the rows above proved nothing; got %v", rc.Branches["main"].Contexts)
	})
}
