package quality

// affected_scope.go — the pure core that decides which Go packages a commit
// can affect (SPEC-PRECOMMIT-AFFECTED-SCOPE-001, REQ-PAS-002..-007).
//
// The core takes the changed-file list and an in-memory package graph and
// returns a selection plus the report lines. It spawns no process and reads
// neither the clock nor the environment, so its tests are deterministic; the
// git and `go list` adapters that feed it live beside the gate wiring.

// scopePackage is one package of the module graph.
type scopePackage struct {
	ImportPath string
	// HasTests is true when the package has an in-package or external test file.
	HasTests bool
	// Imports, TestImports and XTestImports are the import paths named by the
	// package's regular files, its in-package tests and its external-test
	// package. All three create a dependency edge (REQ-PAS-002).
	Imports      []string
	TestImports  []string
	XTestImports []string
	// EmbedFiles lists every file the package embeds (regular and test embeds)
	// as a module-root-relative slash path.
	EmbedFiles []string
}

// moduleGraph is the package graph of one Go module root.
type moduleGraph struct {
	ModulePath string
	Packages   map[string]*scopePackage
}

// affectedInput is the whole input of the selection.
type affectedInput struct {
	// Changed lists the changed paths, repository-root-relative, slash form.
	Changed []string
	// ModuleRel is the module root's offset inside the repository ("" when the
	// module root is the repository root).
	ModuleRel string
	Graph     moduleGraph
}

// affectedResult is the selection and its report.
type affectedResult struct {
	// Full is true when the full-module scope applies (a widening trigger fired).
	Full bool
	// Trigger names the changed path that widened the scope.
	Trigger string
	// Selected holds the selected import paths before no-test packages are dropped.
	Selected []string
	// Run holds the module-root-relative directory patterns to hand to go test.
	Run []string
	// Dropped holds the selected import paths dropped for having no test files.
	Dropped []string
	// Unowned holds the changed paths no package owns.
	Unowned []string
	// SkipReason is non-empty when nothing is left to run and the scope is not full.
	SkipReason string
	// Notes are the report lines for the gate summary.
	Notes []string
}

// widenTrigger reports the first changed path that changes the dependency graph
// itself, which widens the scope to the full module (REQ-PAS-007).
func widenTrigger(changed []string) (string, bool) {
	return "", false
}

// selectAffected computes the affected package set (REQ-PAS-002..-006).
func selectAffected(in affectedInput) affectedResult {
	return affectedResult{}
}
