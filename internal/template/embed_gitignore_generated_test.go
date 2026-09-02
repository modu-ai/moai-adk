// Card t373 template embed test.
//
// `moai graph build` writes .moai/project/graph/edges.jsonl (a ~180k-line
// generated artifact) into the user's project. The template .gitignore did not
// cover it, so a single `git add -A` sweeps it into a commit. That has already
// happened once in this repository (commit 59a622c5a, +180,178 lines; it never
// reached develop). Every project deployed from the template carries the same
// hole.
//
// The template source MUST be edited first and re-embedded via `make build`
// (Template-First), which is why this asserts against the EMBEDDED FS rather
// than the file on disk: a template edit without `make build` leaves the
// shipped binary serving the old .gitignore, and that gap is exactly what this
// guard catches.
package template

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// generatedArtifactIgnoreRules are the .gitignore lines that must reach users.
// Each names a path some `moai` subcommand WRITES into the project tree and
// that nothing is expected to commit.
var generatedArtifactIgnoreRules = []string{
	// `moai graph build` (internal/cli/graph.go) — edges.jsonl + edges.meta.json.
	".moai/project/graph/",
}

func TestEmbeddedGitignoreCoversGeneratedArtifacts(t *testing.T) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	data, err := fs.ReadFile(fsys, ".gitignore")
	if err != nil {
		t.Fatalf("read embedded .gitignore: %v", err)
	}

	present := gitignoreRuleLines(data)

	for _, rule := range generatedArtifactIgnoreRules {
		if !present[rule] {
			t.Errorf("embedded .gitignore missing generated-artifact rule %q — "+
				"a `git add -A` in a deployed project would commit it", rule)
		}
	}
}

// gitignoreRuleLines splits a .gitignore body into the set of discrete,
// whitespace-trimmed lines it declares. Discrete-line membership, never a
// substring match: a substring is also satisfied by the path appearing inside
// a comment, which ignores nothing.
func gitignoreRuleLines(data []byte) map[string]bool {
	present := make(map[string]bool)
	for _, ln := range strings.Split(string(data), "\n") {
		present[strings.TrimSpace(ln)] = true
	}
	return present
}

// repositoryRoot walks up from the package directory (go test's working
// directory) until it reaches the directory holding go.mod. The root
// .gitignore lives there. The walk is used rather than a fixed "../.." so the
// guard survives the package moving, and rather than an environment variable
// so it does not depend on how the test was launched.
func repositoryRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd(): %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("no go.mod found walking up from the package directory")
		}
		dir = parent
	}
}

// TestGitignoreDeclaredRulesOnBothSurfaces asserts every declared
// generated-artifact rule on BOTH surfaces: the embedded template filesystem
// (what a deployed project receives) AND this repository's own root
// .gitignore. TestEmbeddedGitignoreCoversGeneratedArtifacts reads only the
// first, so removing a rule from the root file alone leaves it green while the
// hole that rule closed is reopened here.
//
// Only the DECLARED rules are asserted. The two files legitimately diverge —
// dozens of repository-specific rules exist on one side only, and the comment
// blocks differ on purpose — so neither byte equality nor full rule-set
// equality is available.
func TestGitignoreDeclaredRulesOnBothSurfaces(t *testing.T) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	embeddedData, err := fs.ReadFile(fsys, ".gitignore")
	if err != nil {
		t.Fatalf("read embedded .gitignore: %v", err)
	}

	rootPath := filepath.Join(repositoryRoot(t), ".gitignore")
	rootData, err := os.ReadFile(rootPath)
	if err != nil {
		t.Fatalf("read root .gitignore at %s: %v", rootPath, err)
	}

	surfaces := []struct {
		name  string // the surface token the failure message names
		where string
		lines map[string]bool
	}{
		{name: "root", where: rootPath, lines: gitignoreRuleLines(rootData)},
		{name: "embedded", where: "embedded template FS .gitignore", lines: gitignoreRuleLines(embeddedData)},
	}

	for _, rule := range generatedArtifactIgnoreRules {
		for _, surface := range surfaces {
			if !surface.lines[rule] {
				t.Errorf("surface %q (%s) is missing generated-artifact rule %q — "+
					"a `git add -A` on that surface would commit the artifact",
					surface.name, surface.where, rule)
			}
		}
	}
}
