package quality

// affected_scope_queries.go — the two thin adapters that feed the affected-set
// core: the commit-accurate changed-file list (git) and the module package
// graph (go list). SPEC-PRECOMMIT-AFFECTED-SCOPE-001 REQ-PAS-005/-008/-009.

import (
	"context"
	"errors"
)

// commitChanges is what the commit being assembled will record.
type commitChanges struct {
	// Files are the changed paths, repository-root-relative, slash form.
	Files []string
	// ModulePrefix is the queried directory's offset inside the repository
	// ("" when it is the repository root).
	ModulePrefix string
}

// queryCommitChanges returns the changes the commit will record.
func queryCommitChanges(ctx context.Context, dir string) (commitChanges, error) {
	return commitChanges{}, errors.New("affected-scope: commit-changes adapter not implemented")
}

// listModuleGraph returns the package graph of the Go module rooted at dir.
func listModuleGraph(ctx context.Context, dir, tags string) (moduleGraph, error) {
	return moduleGraph{}, errors.New("affected-scope: package-graph adapter not implemented")
}
