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

	// Discrete-line match, not a substring match: a substring would also be
	// satisfied by the path appearing inside a comment, which ignores nothing.
	present := make(map[string]bool)
	for _, ln := range strings.Split(string(data), "\n") {
		present[strings.TrimSpace(ln)] = true
	}

	for _, rule := range generatedArtifactIgnoreRules {
		if !present[rule] {
			t.Errorf("embedded .gitignore missing generated-artifact rule %q — "+
				"a `git add -A` in a deployed project would commit it", rule)
		}
	}
}
