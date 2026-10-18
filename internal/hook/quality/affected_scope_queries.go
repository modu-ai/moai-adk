package quality

// affected_scope_queries.go — the two thin adapters that feed the affected-set
// core: the commit-accurate changed-file list (git) and the module package
// graph (go list). SPEC-PRECOMMIT-AFFECTED-SCOPE-001 REQ-PAS-005/-008/-009.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// commitChanges is what the commit being assembled will record.
type commitChanges struct {
	// Files are the changed paths, repository-root-relative, slash form.
	Files []string
	// ModulePrefix is the queried directory's offset inside the repository
	// ("" when it is the repository root).
	ModulePrefix string
}

// runScopeQuery runs one read-only query command in dir with env and returns its
// stdout. It carries the same process-group and wait-delay hygiene as runStep,
// so a hung query's descendants do not outlive it.
func runScopeQuery(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	isolateProcessGroup(cmd)
	cmd.WaitDelay = stepWaitGrace
	cmd.Cancel = func() error {
		terminateProcessGroup(cmd.Process.Pid)
		return cmd.Process.Kill()
	}
	cmd.Dir = dir
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	defer func() {
		if cmd.Process != nil {
			terminateProcessGroup(cmd.Process.Pid)
		}
	}()
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}

// commitQueryEnv is the environment of the changed-set query: every
// repository-location variable is scrubbed (card t560 — a leaked GIT_DIR names
// another repository) EXCEPT the index location the pre-commit hook was handed.
// Under `git commit -a` and `git commit <paths>` that variable names the
// temporary index holding the commit's content, which the real index does not
// (SPEC-PRECOMMIT-AFFECTED-SCOPE-001 REQ-PAS-009). A relative value is relative
// to the hook's working directory, which is this process's.
func commitQueryEnv() []string {
	env := stepEnv()
	if idx := os.Getenv("GIT_INDEX_FILE"); idx != "" {
		if abs, err := filepath.Abs(idx); err == nil {
			idx = abs
		}
		env = append(env, "GIT_INDEX_FILE="+idx)
	}
	return env
}

// queryCommitChanges returns the paths the commit being assembled will record.
// Rename detection is off, so a rename reports both its old and its new path
// (REQ-PAS-005), and names are NUL-separated so unusual file names survive. An
// empty list with a nil error means nothing is staged; a non-nil error means the
// query failed — the caller must tell the two apart (REQ-PAS-008).
//
// For `git commit --amend` the result is what is newly staged against the HEAD
// being amended, not the files the amended commit already carried.
func queryCommitChanges(ctx context.Context, dir string) (commitChanges, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return commitChanges{}, fmt.Errorf("git is not available: %w", err)
	}
	env := commitQueryEnv()
	prefix, err := runScopeQuery(ctx, dir, env, "git", "rev-parse", "--show-prefix")
	if err != nil {
		return commitChanges{}, err
	}
	out, err := runScopeQuery(ctx, dir, env, "git", "diff", "--cached", "--name-only", "--no-renames", "-z")
	if err != nil {
		return commitChanges{}, err
	}
	ch := commitChanges{ModulePrefix: strings.TrimSuffix(strings.TrimSpace(string(prefix)), "/")}
	for _, name := range strings.Split(string(out), "\x00") {
		if name != "" {
			ch.Files = append(ch.Files, name)
		}
	}
	return ch, nil
}

// goListPackage is the slice of `go list -json` the graph needs.
type goListPackage struct {
	ImportPath      string
	Module          *struct{ Path string }
	TestGoFiles     []string
	XTestGoFiles    []string
	Imports         []string
	TestImports     []string
	XTestImports    []string
	EmbedFiles      []string
	TestEmbedFiles  []string
	XTestEmbedFiles []string
	Error           *struct{ Err string }
	Incomplete      bool
	DepsErrors      []struct{ Err string }
}

// problem describes why go list marked the package errored or incomplete, or
// returns "" for a sound package. `go list -e` exits 0 for both, so the exit
// status proves nothing (REQ-PAS-008).
func (p *goListPackage) problem() string {
	switch {
	case p.Error != nil:
		return "errored: " + firstLine(p.Error.Err)
	case p.Incomplete && len(p.DepsErrors) > 0:
		return "incomplete: " + firstLine(p.DepsErrors[0].Err)
	case p.Incomplete:
		return "incomplete"
	}
	return ""
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(s), "\n")
	return line
}

// listModuleGraph returns the package graph of the Go module rooted at dir,
// from one `go list -e -json` query over ./... . A package go list reports as
// errored or incomplete fails the whole query: a selection computed from a
// partial graph could under-scope the commit.
func listModuleGraph(ctx context.Context, dir, tags string) (moduleGraph, error) {
	args := []string{"list", "-e",
		"-json=ImportPath,Module,TestGoFiles,XTestGoFiles,Imports,TestImports,XTestImports," +
			"EmbedFiles,TestEmbedFiles,XTestEmbedFiles,Error,Incomplete,DepsErrors"}
	if tags != "" {
		args = append(args, "-tags="+tags)
	}
	args = append(args, "./...")
	out, err := runScopeQuery(ctx, dir, stepEnv(), "go", args...)
	if err != nil {
		return moduleGraph{}, err
	}

	g := moduleGraph{Packages: map[string]*scopePackage{}}
	dec := json.NewDecoder(bytes.NewReader(out))
	for {
		var p goListPackage
		if err := dec.Decode(&p); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return moduleGraph{}, fmt.Errorf("decode go list output: %w", err)
		}
		if why := p.problem(); why != "" {
			return moduleGraph{}, fmt.Errorf("package %s is %s", p.ImportPath, why)
		}
		if p.Module == nil || p.Module.Path == "" {
			return moduleGraph{}, fmt.Errorf("package %s reports no module", p.ImportPath)
		}
		if g.ModulePath == "" {
			g.ModulePath = p.Module.Path
		} else if g.ModulePath != p.Module.Path {
			return moduleGraph{}, fmt.Errorf("package %s belongs to module %s, not %s", p.ImportPath, p.Module.Path, g.ModulePath)
		}
		g.Packages[p.ImportPath] = p.toScopePackage()
	}
	if len(g.Packages) == 0 {
		return moduleGraph{}, errors.New("go list reported no packages")
	}
	return g, nil
}

// toScopePackage converts a go list record, mapping its embedded files (relative
// to the package directory) to module-root-relative slash paths.
func (p *goListPackage) toScopePackage() *scopePackage {
	relDir := strings.TrimPrefix(strings.TrimPrefix(p.ImportPath, p.Module.Path), "/")
	sp := &scopePackage{
		ImportPath:   p.ImportPath,
		HasTests:     len(p.TestGoFiles)+len(p.XTestGoFiles) > 0,
		Imports:      p.Imports,
		TestImports:  p.TestImports,
		XTestImports: p.XTestImports,
	}
	for _, list := range [][]string{p.EmbedFiles, p.TestEmbedFiles, p.XTestEmbedFiles} {
		for _, f := range list {
			sp.EmbedFiles = append(sp.EmbedFiles, path.Join(relDir, filepath.ToSlash(f)))
		}
	}
	return sp
}
