package quality

// affected_scope.go — the pure core that decides which Go packages a commit
// can affect (SPEC-PRECOMMIT-AFFECTED-SCOPE-001, REQ-PAS-002..-007).
//
// The core takes the changed-file list and an in-memory package graph and
// returns a selection plus the report lines. It spawns no process and reads
// neither the clock nor the environment, so its tests are deterministic; the
// git and `go list` adapters that feed it live beside the gate wiring.

import (
	"fmt"
	"path"
	"slices"
	"strings"
)

// scopeNotePrefix opens every report line the affected-scope logic adds to the
// gate summary, so a reader can tell them from the step inventory.
const scopeNotePrefix = "affected-scope: "

// scopeNoTestFilesReason is the reason a selected package is dropped (REQ-PAS-006).
const scopeNoTestFilesReason = "no test files"

// widenBuildTagsPath is the build-tags file whose change alters the package
// graph the gate computes, so it widens the scope like go.mod does.
const widenBuildTagsPath = ".moai/config/build-tags"

// widenBaseNames are the file names that change the module graph at any
// directory level (REQ-PAS-007).
var widenBaseNames = []string{"go.mod", "go.sum", "go.work", "go.work.sum"}

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
// itself, which widens the scope to the full module (REQ-PAS-007). It looks at
// every path of the commit, so it runs before any module-root filtering.
func widenTrigger(changed []string) (string, bool) {
	for _, p := range changed {
		if slices.Contains(widenBaseNames, path.Base(p)) ||
			p == widenBuildTagsPath || strings.HasSuffix(p, "/"+widenBuildTagsPath) {
			return p, true
		}
	}
	return "", false
}

// @MX:NOTE: [AUTO] pure core of the pre-commit Go test scope — no process, clock or environment access; the git and go list adapters feed it
// @MX:SPEC: SPEC-PRECOMMIT-AFFECTED-SCOPE-001
// selectAffected computes the affected package set: the packages that own a
// changed file plus the packages that directly import a package with a source
// change (REQ-PAS-002..-006).
func selectAffected(in affectedInput) affectedResult {
	if trigger, ok := widenTrigger(in.Changed); ok {
		return affectedResult{
			Full:    true,
			Trigger: trigger,
			Notes: []string{fmt.Sprintf("%swidened to the full module (go test ./...): %s changed",
				scopeNotePrefix, trigger)},
		}
	}

	g := in.Graph
	owners := newScopeOwners(g)
	touched := map[string]bool{} // import paths owning a changed file
	source := map[string]bool{}  // import paths owning a source (non-test-only) change
	var unowned []string
	inModule := 0
	for _, p := range in.Changed {
		rel, ok := moduleRelative(p, in.ModuleRel)
		if !ok {
			continue // outside this toolchain's module root
		}
		inModule++
		ip, ok := owners.owner(rel)
		if !ok {
			unowned = append(unowned, p)
			continue
		}
		touched[ip] = true
		if !isTestOnlyPath(rel) {
			source[ip] = true
		}
	}

	selected := map[string]bool{}
	for ip := range touched {
		if _, exists := g.Packages[ip]; exists {
			selected[ip] = true
		}
	}
	importers := g.importers()
	for ip := range source {
		for _, dependent := range importers[ip] {
			selected[dependent] = true
		}
	}

	res := affectedResult{Selected: sortedKeys(selected), Unowned: unowned}
	slices.Sort(res.Unowned)
	for _, ip := range res.Selected {
		if g.Packages[ip].HasTests {
			res.Run = append(res.Run, g.dirPattern(ip))
		} else {
			res.Dropped = append(res.Dropped, ip)
		}
	}
	slices.Sort(res.Run)

	if len(res.Selected) > 0 {
		res.Notes = append(res.Notes, fmt.Sprintf("%sselected %d, dropped %d (running %d)",
			scopeNotePrefix, len(res.Selected), len(res.Dropped), len(res.Run)))
	}
	for _, ip := range res.Dropped {
		res.Notes = append(res.Notes, fmt.Sprintf("%sdropped %s: %s", scopeNotePrefix, ip, scopeNoTestFilesReason))
	}
	for _, p := range res.Unowned {
		res.Notes = append(res.Notes, fmt.Sprintf("%s%s is owned by no package", scopeNotePrefix, p))
	}

	if len(res.Run) == 0 {
		switch {
		case len(in.Changed) == 0:
			res.SkipReason = "the commit records no changed file"
		case inModule == 0:
			res.SkipReason = "no changed file lies inside this module root"
		case len(res.Selected) == 0:
			res.SkipReason = "no package owns a changed file"
		default:
			res.SkipReason = "every selected package has " + scopeNoTestFilesReason
		}
	}
	return res
}

// moduleRelative maps a repository-relative path to the module root, reporting
// false for a path outside it.
func moduleRelative(p, moduleRel string) (string, bool) {
	if moduleRel == "" {
		return p, true
	}
	if rel, ok := strings.CutPrefix(p, moduleRel+"/"); ok {
		return rel, true
	}
	return "", false
}

// hasTestdataDir returns the index of the first "testdata" directory segment of
// a module-relative path, or -1. A file that is itself named testdata is not
// under a testdata directory.
func hasTestdataDir(segs []string) int {
	for i := 0; i < len(segs)-1; i++ {
		if segs[i] == "testdata" {
			return i
		}
	}
	return -1
}

// isTestOnlyPath reports whether a changed file is a test-only change: a
// _test.go file, or any file under a testdata directory (REQ-PAS-004).
func isTestOnlyPath(rel string) bool {
	return strings.HasSuffix(rel, "_test.go") || hasTestdataDir(strings.Split(rel, "/")) >= 0
}

// scopeOwners resolves the owning package of a changed file (REQ-PAS-003).
type scopeOwners struct {
	g        moduleGraph
	embedded map[string]string // module-relative embedded file -> owning import path
}

func newScopeOwners(g moduleGraph) scopeOwners {
	o := scopeOwners{g: g, embedded: map[string]string{}}
	for ip, p := range g.Packages {
		for _, f := range p.EmbedFiles {
			o.embedded[f] = ip
		}
	}
	return o
}

// importPathOf maps a module-relative directory to its import path; it does not
// require the package to exist, so a deleted package still has one.
func (o scopeOwners) importPathOf(dir string) string {
	if dir == "" || dir == "." {
		return o.g.ModulePath
	}
	return o.g.ModulePath + "/" + dir
}

// owner decides the owning import path of a module-relative changed file, in
// the REQ-PAS-003 order: (1) a path under a testdata directory belongs to the
// package that directory sits in; (2) a Go source file of any build
// configuration belongs to its directory's package; (3) any other file belongs
// to the package whose embedded set holds it, else to the nearest ancestor
// directory that is a package.
func (o scopeOwners) owner(rel string) (string, bool) {
	segs := strings.Split(rel, "/")
	if i := hasTestdataDir(segs); i >= 0 {
		return o.importPathOf(strings.Join(segs[:i], "/")), true
	}
	if strings.HasSuffix(rel, ".go") {
		return o.importPathOf(path.Dir(rel)), true
	}
	if ip, ok := o.embedded[rel]; ok {
		return ip, true
	}
	for d := path.Dir(rel); ; d = path.Dir(d) {
		ip := o.importPathOf(d)
		if _, ok := o.g.Packages[ip]; ok {
			return ip, true
		}
		if d == "." {
			return "", false
		}
	}
}

// importers maps an import path to the packages that directly import it through
// their regular, in-package-test or external-test imports (REQ-PAS-002). A
// package is never its own importer: its external-test package names it.
func (g moduleGraph) importers() map[string][]string {
	out := map[string][]string{}
	for ip, p := range g.Packages {
		for _, list := range [][]string{p.Imports, p.TestImports, p.XTestImports} {
			for _, imp := range list {
				if imp != ip && !slices.Contains(out[imp], ip) {
					out[imp] = append(out[imp], ip)
				}
			}
		}
	}
	return out
}

// dirPattern renders an import path as the module-root-relative directory
// pattern handed to go test ("." for the root package, "./a/b" otherwise).
func (g moduleGraph) dirPattern(importPath string) string {
	rel := strings.TrimPrefix(strings.TrimPrefix(importPath, g.ModulePath), "/")
	if rel == "" {
		return "."
	}
	return "./" + rel
}

// sortedKeys returns the keys of a set in ascending order.
func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
