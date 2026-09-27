package escalation

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/gitenv"
	"github.com/modu-ai/moai-adk/internal/navigator/astx"
)

// HookCheckpoint names the on-demand checkpoint, the only place class 4 runs
// (spec.md C4). It starts git, so it is never called from a tool-call hook.
const HookCheckpoint = "Checkpoint"

// Class 4 addition sub-kinds (REQ-AE-009).
const (
	AdditionPackage      = "new-package"
	AdditionExportedDecl = "exported-declaration"
	AdditionCLIVerb      = "cli-verb"
	AdditionMCPTool      = "mcp-tool"
	AdditionConfigKey    = "config-key"
)

// registrationSubKinds are the sub-kinds that need a per-language
// registration recognizer; only Go has one (design.md §C.4).
var registrationSubKinds = []string{AdditionCLIVerb, AdditionMCPTool, AdditionConfigKey}

// Go registration recognizers, matched against added diff lines.
var (
	goCLIVerbRe   = regexp.MustCompile(`\bUse:\s*"([A-Za-z][\w-]*)`)
	goMCPToolRe   = regexp.MustCompile(`\bNewTool\(\s*"([^"]+)"`)
	goConfigKeyRe = regexp.MustCompile("`[^`]*\\byaml:\"([A-Za-z0-9_]+)")
)

// Addition is one observed new architecture or API element.
type Addition struct {
	Kind string
	// Name is the qualified name: the package directory for a new package,
	// the declaration name, CLI verb, MCP tool name, or configuration key.
	Name string
	// Path is the repository-relative file or directory it was observed in.
	Path string
}

// CheckpointResult is what an on-demand checkpoint observed.
type CheckpointResult struct {
	Additions []Addition
	// NotObserved lists every class 4 detection the checkpoint could not
	// complete (REQ-AE-022).
	NotObserved []string
}

// Checkpoint runs the on-demand checkpoint for the worktree containing cwd:
// the disarm check and resolution as on every event, then class 4
// new-architecture-or-api against the card base (REQ-AE-009). It records
// only and returns what it observed; under any mode other than contract it
// does nothing.
//
// @MX:NOTE: [AUTO] no CLI verb calls this yet — the verb is plan.md §H Q2, an open operator decision; the function is the checkpoint body any surface will call
func Checkpoint(s config.AutonomySettings, cwd string) CheckpointResult {
	var res CheckpointResult
	if !Active(s) {
		return res
	}
	observeInto(s, Event{Hook: HookCheckpoint, CWD: cwd}, time.Now(), &res)
	return res
}

// classNewAPI trips class 4 once per addition between the card base and
// HEAD, and lists what it could not observe.
func (r *run) classNewAPI() {
	if r.s.NewAPIDetector == config.AutonomyNewAPIDetectorOff {
		return
	}
	res := r.checkpoint
	base, err := cardBase(r.root)
	if err != nil {
		res.NotObserved = append(res.NotObserved, "card base: "+err.Error())
		r.notObserved(ClassNewArchitectureOrAPI, "on-demand checkpoint", res.NotObserved)
		return
	}
	adds, notObs, err := newAPIAdditions(r.root, base)
	if err != nil {
		r.notChecked(ClassNewArchitectureOrAPI, err.Error())
		return
	}
	res.Additions, res.NotObserved = adds, notObs
	if len(notObs) > 0 {
		r.notObserved(ClassNewArchitectureOrAPI, "on-demand checkpoint", notObs)
	}
	cdata, _ := os.ReadFile(r.st.Armed.ContractPath)
	ref := contractRef(ContractItemLine(cdata, ClassNewArchitectureOrAPI, "escalate_on"), cdata, "escalate_on")
	for _, a := range adds {
		r.writeRecord(Record{
			Kind: KindContract, Class: ClassNewArchitectureOrAPI, EscalateOn: ClassNewArchitectureOrAPI,
			Fingerprint: Fingerprint(ClassNewArchitectureOrAPI, a.Kind, a.Name),
			ContractRef: ref, NotObserved: slices.Clone(notObs),
			Observation: fmt.Sprintf("Between card base %s and HEAD:\n\n- kind: %s\n- name: %s\n- path: %s",
				shortSHA(base), a.Kind, a.Name, a.Path),
			Options: []string{
				"Accept the addition as within the contract's approach and record why",
				"Remove the addition, or amend the contract's approach and re-sign it",
			},
		})
	}
}

// cardBase is the merge-base of HEAD with the integration branch the git
// strategy configuration names: the development branch where one is
// configured, otherwise the default branch (REQ-AE-009).
func cardBase(root string) (string, error) {
	var candidates []string
	if dev := config.LoadGitFlowDevelopBranch(root); dev != "" {
		candidates = []string{dev}
	} else {
		if ref, err := gitOut(root, "symbolic-ref", "-q", "--short", "refs/remotes/origin/HEAD"); err == nil && ref != "" {
			candidates = append(candidates, ref)
		}
		candidates = append(candidates, "main", "master")
	}
	for _, branch := range candidates {
		if _, err := gitOut(root, "rev-parse", "--verify", "-q", branch+"^{commit}"); err != nil {
			continue
		}
		base, err := gitOut(root, "merge-base", "HEAD", branch)
		if err != nil || base == "" {
			return "", fmt.Errorf("no merge-base of HEAD with %s", branch)
		}
		return base, nil
	}
	return "", fmt.Errorf("no integration branch resolves (tried %s)", strings.Join(candidates, ", "))
}

// newAPIAdditions compares base with HEAD.
func newAPIAdditions(root, base string) ([]Addition, []string, error) {
	changed, err := gitOut(root, "diff", "--name-status", "--no-renames", base, "HEAD")
	if err != nil {
		return nil, nil, fmt.Errorf("list changes: %w", err)
	}
	baseFiles, err := gitOut(root, "ls-tree", "-r", "--name-only", base)
	if err != nil {
		return nil, nil, fmt.Errorf("list base tree: %w", err)
	}
	baseDirs := map[string]bool{}
	for _, f := range strings.Split(baseFiles, "\n") {
		if f != "" {
			baseDirs[path.Dir(f)] = true
		}
	}

	var adds []Addition
	var notObs []string
	seenPkg := map[string]bool{}
	noRecognizer := map[string]bool{}
	for _, line := range strings.Split(changed, "\n") {
		status, file, ok := strings.Cut(line, "\t")
		if !ok || (status != "A" && status != "M") {
			continue
		}
		lang := astx.DetectLanguage(file)
		if lang == "" {
			continue // not source code
		}
		if dir := path.Dir(file); status == "A" && !baseDirs[dir] && !seenPkg[dir] {
			seenPkg[dir] = true
			adds = append(adds, Addition{Kind: AdditionPackage, Name: dir, Path: dir})
		}
		decls, supported, err := newDeclarations(root, base, file, lang, status == "A")
		if err != nil {
			return nil, nil, err
		}
		if !supported {
			notObs = append(notObs, fmt.Sprintf("%s %s (%s: extraction unsupported)", AdditionExportedDecl, file, lang))
		}
		for _, d := range decls {
			adds = append(adds, Addition{Kind: AdditionExportedDecl, Name: d, Path: file})
		}
		if lang != "go" {
			noRecognizer[lang] = true
			continue
		}
		regs, err := goRegistrations(root, base, file)
		if err != nil {
			return nil, nil, err
		}
		adds = append(adds, regs...)
	}
	langs := make([]string, 0, len(noRecognizer))
	for l := range noRecognizer {
		langs = append(langs, l)
	}
	slices.Sort(langs)
	for _, l := range langs {
		for _, sub := range registrationSubKinds {
			notObs = append(notObs, sub+" ("+l+")")
		}
	}
	return adds, notObs, nil
}

// newDeclarations returns the exported declarations of file at HEAD that it
// did not have at base. Go is parsed with the standard library; other
// languages go through the navigator extractor, whose grammars need CGO.
func newDeclarations(root, base, file, lang string, added bool) ([]string, bool, error) {
	head, err := gitShow(root, "HEAD", file)
	if err != nil {
		return nil, false, err
	}
	var before []byte
	if !added {
		if before, err = gitShow(root, base, file); err != nil {
			return nil, false, err
		}
	}
	extract := goExported
	if lang != "go" {
		extract = func(src []byte) ([]string, bool) { return astxDecls(lang, file, src) }
	}
	now, ok := extract(head)
	if !ok {
		return nil, false, nil
	}
	var prev []string
	if !added {
		if prev, ok = extract(before); !ok {
			return nil, false, nil
		}
	}
	var out []string
	for _, d := range now {
		if !slices.Contains(prev, d) && !slices.Contains(out, d) {
			out = append(out, d)
		}
	}
	return out, true, nil
}

// goExported lists the exported top-level declarations of a Go file: funcs,
// methods (Type.Method), types, consts, and vars.
func goExported(src []byte) ([]string, bool) {
	f, err := parser.ParseFile(token.NewFileSet(), "", src, parser.SkipObjectResolution)
	if err != nil {
		return nil, false
	}
	var out []string
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			name := d.Name.Name
			if d.Recv != nil && len(d.Recv.List) > 0 {
				name = receiverName(d.Recv.List[0].Type) + "." + name
			}
			if exportedGo(d.Name.Name) {
				out = append(out, name)
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch sp := spec.(type) {
				case *ast.TypeSpec:
					if exportedGo(sp.Name.Name) {
						out = append(out, sp.Name.Name)
					}
				case *ast.ValueSpec:
					for _, n := range sp.Names {
						if exportedGo(n.Name) {
							out = append(out, n.Name)
						}
					}
				}
			}
		}
	}
	return out, true
}

// receiverName returns a method receiver's base type name.
func receiverName(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.StarExpr:
		return receiverName(t.X)
	case *ast.IndexExpr:
		return receiverName(t.X)
	case *ast.IndexListExpr:
		return receiverName(t.X)
	case *ast.Ident:
		return t.Name
	}
	return "?"
}

func exportedGo(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}

// astxDecls extracts the declaration names of a non-Go source through the
// navigator extractor (graph_file_api's extractor), from a temp copy.
func astxDecls(lang, file string, src []byte) ([]string, bool) {
	dir, err := os.MkdirTemp("", "moai-escalation-*")
	if err != nil {
		return nil, false
	}
	defer func() { _ = os.RemoveAll(dir) }()
	p := filepath.Join(dir, filepath.Base(file))
	if err := os.WriteFile(p, src, 0o600); err != nil {
		return nil, false
	}
	set, err := astx.Extract(lang, p)
	if err != nil || !set.Supported {
		return nil, false
	}
	var out []string
	for _, syms := range set.Symbols {
		for _, s := range syms {
			if !slices.Contains(out, s.Name) {
				out = append(out, s.Name)
			}
		}
	}
	slices.Sort(out)
	return out, true
}

// goRegistrations applies the Go registration recognizers to the lines a Go
// file gained between base and HEAD.
func goRegistrations(root, base, file string) ([]Addition, error) {
	diff, err := gitOut(root, "diff", "-U0", base, "HEAD", "--", file)
	if err != nil {
		return nil, fmt.Errorf("diff %s: %w", file, err)
	}
	var out []Addition
	for _, line := range strings.Split(diff, "\n") {
		if !strings.HasPrefix(line, "+") || strings.HasPrefix(line, "+++") {
			continue
		}
		for _, rc := range []struct {
			kind string
			re   *regexp.Regexp
		}{{AdditionCLIVerb, goCLIVerbRe}, {AdditionMCPTool, goMCPToolRe}, {AdditionConfigKey, goConfigKeyRe}} {
			for _, m := range rc.re.FindAllStringSubmatch(line, -1) {
				out = append(out, Addition{Kind: rc.kind, Name: m[1], Path: file})
			}
		}
	}
	return out, nil
}

// gitShow returns the bytes of path at rev.
func gitShow(root, rev, file string) ([]byte, error) {
	out, err := gitRaw(root, "show", rev+":"+file)
	if err != nil {
		return nil, fmt.Errorf("read %s at %s: %w", file, shortSHA(rev), err)
	}
	return out, nil
}

// gitOut runs git in root and returns trimmed stdout.
func gitOut(root string, args ...string) (string, error) {
	out, err := gitRaw(root, args...)
	return strings.TrimSpace(string(out)), err
}

// gitRaw runs git in root with the caller's repository-locating variables
// scrubbed.
func gitRaw(root string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	cmd.Env = gitenv.Env()
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errb.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, errors.New(msg)
	}
	return out.Bytes(), nil
}

func shortSHA(s string) string {
	if len(s) > 12 {
		return s[:12]
	}
	return s
}
