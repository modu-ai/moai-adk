// ci_workflow_model_test.go: shared workflow model for the
// SPEC-RELEASE-PIPELINE-001 guard tests (card t1536).
//
// The guard tests in ci_cost_gating_test.go, verification_budget_test.go and
// release_drafter_cleanup_test.go read the workflow YAML from the repository
// tree (never a copied fixture, plan.md B3) and need three things the standard
// library does not give them:
//
//   - a typed view of a workflow (jobs, steps, matrix, triggers);
//   - an evaluator for the subset of the GitHub Actions expression language the
//     `if:` conditions use, so a guard can ask "does this job run, and what does
//     it conclude" under a stated event / classification outcome instead of
//     pattern-matching condition text;
//   - a way to mutate an in-memory copy so each guard carries a negative control
//     (verification-completeness.md 1.1: a check is unfinished until its
//     failure has been observed).
//
// Modelling limits, stated so a green is read at its true width. Job-level and
// step-level `if:` are evaluated; `needs` is modelled through the
// `needs.<id>.result` / `.outputs.<name>` context values the test supplies.
// A step whose condition reads a needed job's result (the classification verdict
// of a real job and both steps of a skip-marker, rpcIsClassificationStep) is
// EXECUTED: its `run:` body, with the cell's expressions substituted and the
// workflow / job / step env applied, runs under bash (shells other than bash,
// sh and the default are refused), and a non-zero exit is the step failing. A
// `continue-on-error` step cannot fail its job. Every other run step (go build,
// go vet, ...) is a verification step: it passes, or in the "verification
// fails" mode the first plain one fails. Not modelled: `matrix.*` values inside
// executed scripts (they read as empty), runner-provided commands beyond the
// host PATH, the runner's own bash version, step `timeout-minutes`,
// `working-directory`, and a job-level `continue-on-error` (reported as a
// violation by the verdict guard instead).
package template_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode"

	"gopkg.in/yaml.v3"
)

// rpcStep is one workflow step.
type rpcStep struct {
	ID             string         `yaml:"id"`
	Name           string         `yaml:"name"`
	If             string         `yaml:"if"`
	Uses           string         `yaml:"uses"`
	Shell          string         `yaml:"shell"`
	Run            string         `yaml:"run"`
	With           map[string]any `yaml:"with"`
	Env            map[string]any `yaml:"env"`
	TimeoutMinutes yaml.Node      `yaml:"timeout-minutes"`
	// ContinueOnError is the raw `continue-on-error:` value (bool, or an
	// expression string); see rpcContinuesOnError.
	ContinueOnError any `yaml:"continue-on-error"`
}

// rpcDefaults is a `defaults:` block (workflow or job level); only run.shell is
// modelled.
type rpcDefaults struct {
	Run struct {
		Shell string `yaml:"shell"`
	} `yaml:"run"`
}

// rpcJob is one workflow job.
type rpcJob struct {
	ID             string `yaml:"-"`
	Name           string `yaml:"name"`
	If             string `yaml:"if"`
	Needs          yaml.Node
	TimeoutMinutes yaml.Node `yaml:"timeout-minutes"`
	Outputs        map[string]string
	Env            map[string]any `yaml:"env"`
	Defaults       rpcDefaults    `yaml:"defaults"`
	// ContinueOnError is the raw job-level `continue-on-error:` value.
	ContinueOnError any       `yaml:"continue-on-error"`
	Concurrency     yaml.Node `yaml:"concurrency"`
	Container       yaml.Node `yaml:"container"`
	Services        yaml.Node `yaml:"services"`
	Strategy        struct {
		Matrix rpcMatrix `yaml:"matrix"`
	} `yaml:"strategy"`
	Steps []rpcStep `yaml:"steps"`
}

// rpcMatrixScalarKey is the key under which a `strategy.matrix` written as a
// scalar (`matrix: ${{ fromJSON(needs.x.outputs.m) }}`) is held, so the model
// parses that form instead of failing on it and a guard can see and refuse it.
const rpcMatrixScalarKey = "$scalar"

// rpcMatrix is a job's `strategy.matrix`: normally a mapping, tolerated as a scalar.
type rpcMatrix map[string]any

func (m *rpcMatrix) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode {
		*m = rpcMatrix{rpcMatrixScalarKey: n.Value}
		return nil
	}
	var raw map[string]any
	if err := n.Decode(&raw); err != nil {
		return err
	}
	*m = raw
	return nil
}

// rpcWorkflow is a parsed workflow file.
type rpcWorkflow struct {
	Path        string
	Name        string `yaml:"name"`
	On          yaml.Node
	Concurrency struct {
		Group            string `yaml:"group"`
		CancelInProgress any    `yaml:"cancel-in-progress"`
	} `yaml:"concurrency"`
	Env      map[string]any     `yaml:"env"`
	Defaults rpcDefaults        `yaml:"defaults"`
	Jobs     map[string]*rpcJob `yaml:"jobs"`
}

// rpcReadRepoFile reads a repo-relative file from the module root.
func rpcReadRepoFile(t *testing.T, rel string) string {
	t.Helper()
	root := findProjectRootForMirrorTest(t)
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(raw)
}

// rpcParseWorkflow parses workflow text. The text may be a mutated copy.
func rpcParseWorkflow(t *testing.T, rel, text string) *rpcWorkflow {
	t.Helper()
	var wf rpcWorkflow
	if err := yaml.Unmarshal([]byte(text), &wf); err != nil {
		t.Fatalf("parse %s: %v", rel, err)
	}
	wf.Path = rel
	for id, j := range wf.Jobs {
		j.ID = id
	}
	return &wf
}

// rpcLoadWorkflow reads and parses a workflow from the repository tree.
func rpcLoadWorkflow(t *testing.T, rel string) *rpcWorkflow {
	t.Helper()
	return rpcParseWorkflow(t, rel, rpcReadRepoFile(t, rel))
}

// jobIDs returns the job ids in sorted order, for deterministic output.
func (w *rpcWorkflow) jobIDs() []string {
	ids := make([]string, 0, len(w.Jobs))
	for id := range w.Jobs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// needsList returns the job ids named by `needs:` (scalar or sequence).
func (j *rpcJob) needsList() []string {
	switch j.Needs.Kind {
	case yaml.ScalarNode:
		return []string{j.Needs.Value}
	case yaml.SequenceNode:
		out := make([]string, 0, len(j.Needs.Content))
		for _, c := range j.Needs.Content {
			out = append(out, c.Value)
		}
		return out
	default:
		return nil
	}
}

// rpcTrigger is the filter set a workflow declares for one event.
type rpcTrigger struct {
	present                                                               bool
	types, branches, branchesIgnore, tags, tagsIgnore, paths, pathsIgnore []string
}

// trigger returns the filters the workflow declares for an event (present=false
// when it does not subscribe to the event).
func (w *rpcWorkflow) trigger(event string) rpcTrigger {
	list := func(n *yaml.Node) []string {
		if n.Kind == yaml.ScalarNode {
			return []string{n.Value}
		}
		out := make([]string, 0, len(n.Content))
		for _, c := range n.Content {
			out = append(out, c.Value)
		}
		return out
	}
	switch w.On.Kind {
	case yaml.ScalarNode:
		return rpcTrigger{present: w.On.Value == event}
	case yaml.SequenceNode:
		for _, c := range w.On.Content {
			if c.Value == event {
				return rpcTrigger{present: true}
			}
		}
	case yaml.MappingNode:
		for i := 0; i+1 < len(w.On.Content); i += 2 {
			if w.On.Content[i].Value != event {
				continue
			}
			tr := rpcTrigger{present: true}
			val := w.On.Content[i+1]
			if val.Kind != yaml.MappingNode {
				return tr
			}
			for k := 0; k+1 < len(val.Content); k += 2 {
				v := list(val.Content[k+1])
				switch val.Content[k].Value {
				case "types":
					tr.types = v
				case "branches":
					tr.branches = v
				case "branches-ignore":
					tr.branchesIgnore = v
				case "tags":
					tr.tags = v
				case "tags-ignore":
					tr.tagsIgnore = v
				case "paths":
					tr.paths = v
				case "paths-ignore":
					tr.pathsIgnore = v
				}
			}
			return tr
		}
	}
	return rpcTrigger{}
}

// rpcPullRequestMustFireOn are the activity types a required check needs on a
// pull_request: the check must be published when the PR is opened and again for
// every push to it. (`reopened` is not needed: the head commit of a reopened PR
// already carries its check runs; `ready_for_review` only matters for drafts,
// which the default types already cover.) An absent `types:` list means GitHub's
// defaults (opened, synchronize, reopened), which include both.
var rpcPullRequestMustFireOn = []string{"opened", "synchronize"}

// rpcWorkflowAdmits reports whether an event of the representative shape starts
// the workflow: its `types:` (pull_request), `branches:` / `branches-ignore:`
// (the base branch of a pull_request, the pushed branch of a push), `tags:`-only
// push filters, and `paths:` / `paths-ignore:` (evaluated against a diff
// consisting of exactly {file}). why names the first filter that refuses the event;
// a pattern the glob model cannot read is returned as an error, never guessed at.
func rpcWorkflowAdmits(w *rpcWorkflow, event, branch, file string) (admitted bool, why string, err error) {
	tr := w.trigger(event)
	if !tr.present {
		return false, "it does not subscribe to " + event, nil
	}
	if event == "pull_request" && len(tr.types) > 0 {
		for _, need := range rpcPullRequestMustFireOn {
			if !rpcContains(tr.types, need) {
				return false, fmt.Sprintf("its pull_request types %v omit %q, so the check is not published when a PR is opened or pushed to", tr.types, need), nil
			}
		}
	}
	if event == "push" && len(tr.tags)+len(tr.tagsIgnore) > 0 && len(tr.branches)+len(tr.branchesIgnore) == 0 {
		return false, "its push trigger filters on tags only, so a branch push does not start it", nil
	}
	match := func(patterns []string, s string) (bool, error) {
		for _, p := range patterns {
			if p == s {
				return true, nil
			}
			ok, err := rpcGlobMatch(p, s)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	}
	if len(tr.branches) > 0 {
		ok, err := match(tr.branches, branch)
		if err != nil {
			return false, "", err
		}
		if !ok {
			return false, fmt.Sprintf("its %s branches %v do not include %q", event, tr.branches, branch), nil
		}
	}
	if len(tr.branchesIgnore) > 0 {
		ok, err := match(tr.branchesIgnore, branch)
		if err != nil {
			return false, "", err
		}
		if ok {
			return false, fmt.Sprintf("its %s branches-ignore %v exclude %q", event, tr.branchesIgnore, branch), nil
		}
	}
	ok, err := rpcPathsAdmit(tr.paths, tr.pathsIgnore, file)
	if err != nil {
		return false, "", err
	}
	if !ok {
		return false, fmt.Sprintf("its %s paths / paths-ignore filter skips the diff {%s}", event, file), nil
	}
	return true, "", nil
}

// rpcGlobMatch matches a repo-relative file path against a workflow / paths-filter
// glob: `*` within a path segment, `**` across segments (`**/` also matches no
// directory), `?` one character, dotfiles matched (picomatch dot:true). Negation,
// brace and character-class syntax is refused rather than guessed at.
func rpcGlobMatch(pattern, file string) (bool, error) {
	if strings.ContainsAny(pattern, "{}[]()!") {
		return false, fmt.Errorf("glob syntax in %q is not modelled", pattern)
	}
	var re strings.Builder
	re.WriteString("^")
	for i := 0; i < len(pattern); i++ {
		switch c := pattern[i]; c {
		case '*':
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				for i+1 < len(pattern) && pattern[i+1] == '*' {
					i++
				}
				if i+1 < len(pattern) && pattern[i+1] == '/' {
					i++
					re.WriteString("(?:.*/)?")
				} else {
					re.WriteString(".*")
				}
			} else {
				re.WriteString("[^/]*")
			}
		case '?':
			re.WriteString("[^/]")
		default:
			re.WriteString(regexp.QuoteMeta(string(c)))
		}
	}
	re.WriteString("$")
	return regexp.MatchString(re.String(), file)
}

// rpcPathsAdmit reports whether a workflow's paths / paths-ignore filter lets a
// pull_request or push whose changed files are exactly {file} start.
func rpcPathsAdmit(paths, ignore []string, file string) (bool, error) {
	if len(paths) > 0 {
		for _, p := range paths {
			ok, err := rpcGlobMatch(p, file)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	}
	for _, p := range ignore {
		ok, err := rpcGlobMatch(p, file)
		if err != nil {
			return false, err
		}
		if ok {
			return false, nil
		}
	}
	return true, nil
}

// rpcMatrixCombos expands a strategy.matrix into its legs following GitHub's
// rules: the cartesian product of the array axes, each `include` entry merged
// into every combination it does not contradict (adding keys, never
// overwriting an original axis value) or appended as a new leg when it merges
// into none.
func rpcMatrixCombos(m map[string]any) []map[string]string {
	if len(m) == 0 {
		return nil
	}
	axisKeys := make([]string, 0, len(m))
	for k, v := range m {
		if k == "include" || k == "exclude" {
			continue
		}
		if _, ok := v.([]any); ok {
			axisKeys = append(axisKeys, k)
		}
	}
	sort.Strings(axisKeys)
	var combos []map[string]string
	if len(axisKeys) > 0 {
		combos = []map[string]string{{}}
		for _, k := range axisKeys {
			var next []map[string]string
			for _, base := range combos {
				for _, val := range m[k].([]any) {
					c := make(map[string]string, len(base)+1)
					for bk, bv := range base {
						c[bk] = bv
					}
					c[k] = fmt.Sprint(val)
					next = append(next, c)
				}
			}
			combos = next
		}
		// `exclude` removes the axis combinations it matches (every key/value of an
		// exclude entry equal), before any `include` is applied; sync-audit iter4
		// scan: a matrix that excludes its only leg publishes no check at all.
		excl, _ := m["exclude"].([]any)
		for _, raw := range excl {
			ex, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			var kept []map[string]string
			for _, c := range combos {
				match := true
				for k, v := range ex {
					if c[k] != fmt.Sprint(v) {
						match = false
						break
					}
				}
				if !match {
					kept = append(kept, c)
				}
			}
			combos = kept
		}
	}
	incs, _ := m["include"].([]any)
	axisSet := map[string]bool{}
	for _, k := range axisKeys {
		axisSet[k] = true
	}
	// Only the original (axis) combinations are merge candidates; a leg an
	// earlier include appended is not.
	orig := combos
	for _, raw := range incs {
		inc, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		merged := false
		for _, c := range orig {
			conflict := false
			for k, v := range inc {
				if axisSet[k] && c[k] != fmt.Sprint(v) {
					conflict = true
					break
				}
			}
			if conflict {
				continue
			}
			for k, v := range inc {
				if !axisSet[k] {
					c[k] = fmt.Sprint(v)
				}
			}
			merged = true
		}
		if !merged {
			c := make(map[string]string, len(inc))
			for k, v := range inc {
				c[k] = fmt.Sprint(v)
			}
			combos = append(combos, c)
		}
	}
	return combos
}

var rpcMatrixRefRe = regexp.MustCompile(`\$\{\{\s*matrix\.([A-Za-z0-9_-]+)\s*\}\}`)

// displayNames returns the check-run names the job publishes: the expanded
// `name:` per matrix leg, or `name (v1, v2)` for a matrix job whose name uses
// no matrix expression (GitHub appends the leg values), or the job id.
func (j *rpcJob) displayNames() []string {
	name := j.Name
	if name == "" {
		name = j.ID
	}
	combos := rpcMatrixCombos(j.Strategy.Matrix)
	if len(combos) == 0 {
		if _, scalar := j.Strategy.Matrix[rpcMatrixScalarKey]; len(j.Strategy.Matrix) > 0 && !scalar {
			return nil // a declared matrix whose every leg is excluded publishes nothing
		}
		return []string{name}
	}
	usesExpr := rpcMatrixRefRe.MatchString(name)
	out := make([]string, 0, len(combos))
	for _, c := range combos {
		if usesExpr {
			out = append(out, rpcMatrixRefRe.ReplaceAllStringFunc(name, func(s string) string {
				return c[rpcMatrixRefRe.FindStringSubmatch(s)[1]]
			}))
			continue
		}
		keys := make([]string, 0, len(c))
		for k := range c {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		vals := make([]string, 0, len(keys))
		for _, k := range keys {
			vals = append(vals, c[k])
		}
		out = append(out, name+" ("+strings.Join(vals, ", ")+")")
	}
	return out
}

// ---- expression evaluation -------------------------------------------------

// rpcExprCtx is the context an `if:` expression is evaluated in.
type rpcExprCtx struct {
	vars map[string]string
	// bools are typed boolean context values (workflow_dispatch boolean inputs
	// are booleans in expressions, not strings); a missing key reads as null.
	bools     map[string]bool
	cancelled bool
	success   bool
	failure   bool
}

type rpcTok struct {
	kind string // str ident op lp rp comma eof
	text string
}

func rpcLex(expr string) ([]rpcTok, error) {
	rs := []rune(expr)
	var toks []rpcTok
	for i := 0; i < len(rs); {
		c := rs[i]
		switch {
		case unicode.IsSpace(c):
			i++
		case c == '\'':
			i++
			var b strings.Builder
			closed := false
			for i < len(rs) {
				if rs[i] == '\'' {
					if i+1 < len(rs) && rs[i+1] == '\'' {
						b.WriteRune('\'')
						i += 2
						continue
					}
					i++
					closed = true
					break
				}
				b.WriteRune(rs[i])
				i++
			}
			if !closed {
				return nil, fmt.Errorf("unterminated string literal")
			}
			toks = append(toks, rpcTok{"str", b.String()})
		case c == '(':
			toks = append(toks, rpcTok{"lp", "("})
			i++
		case c == ')':
			toks = append(toks, rpcTok{"rp", ")"})
			i++
		case c == ',':
			toks = append(toks, rpcTok{"comma", ","})
			i++
		case c == '!':
			if i+1 < len(rs) && rs[i+1] == '=' {
				toks = append(toks, rpcTok{"op", "!="})
				i += 2
			} else {
				toks = append(toks, rpcTok{"op", "!"})
				i++
			}
		case c == '=' && i+1 < len(rs) && rs[i+1] == '=':
			toks = append(toks, rpcTok{"op", "=="})
			i += 2
		case c == '&' && i+1 < len(rs) && rs[i+1] == '&':
			toks = append(toks, rpcTok{"op", "&&"})
			i += 2
		case c == '|' && i+1 < len(rs) && rs[i+1] == '|':
			toks = append(toks, rpcTok{"op", "||"})
			i += 2
		case unicode.IsLetter(c) || c == '_':
			j := i
			for j < len(rs) && (unicode.IsLetter(rs[j]) || unicode.IsDigit(rs[j]) || rs[j] == '_' || rs[j] == '.' || rs[j] == '-') {
				j++
			}
			toks = append(toks, rpcTok{"ident", string(rs[i:j])})
			i = j
		default:
			return nil, fmt.Errorf("unsupported character %q", c)
		}
	}
	return append(toks, rpcTok{"eof", ""}), nil
}

type rpcParser struct {
	toks []rpcTok
	pos  int
	ctx  rpcExprCtx
}

func (p *rpcParser) peek() rpcTok { return p.toks[p.pos] }
func (p *rpcParser) next() rpcTok { t := p.toks[p.pos]; p.pos++; return t }

func rpcTruthy(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case string:
		return x != ""
	}
	return false
}

func rpcAsString(v any) string {
	switch x := v.(type) {
	case bool:
		return strconv.FormatBool(x)
	case string:
		return x
	}
	return ""
}

func (p *rpcParser) parseOr() (any, error) {
	l, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.peek().kind == "op" && p.peek().text == "||" {
		p.next()
		r, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		// GitHub semantics: `a || b` yields a when a is truthy, else b.
		if !rpcTruthy(l) {
			l = r
		}
	}
	return l, nil
}

func (p *rpcParser) parseAnd() (any, error) {
	l, err := p.parseEq()
	if err != nil {
		return nil, err
	}
	for p.peek().kind == "op" && p.peek().text == "&&" {
		p.next()
		r, err := p.parseEq()
		if err != nil {
			return nil, err
		}
		// GitHub semantics: `a && b` yields a when a is falsy, else b.
		if rpcTruthy(l) {
			l = r
		}
	}
	return l, nil
}

func (p *rpcParser) parseEq() (any, error) {
	l, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	if p.peek().kind == "op" && (p.peek().text == "==" || p.peek().text == "!=") {
		op := p.next().text
		r, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		eq := strings.EqualFold(rpcAsString(l), rpcAsString(r))
		if op == "!=" {
			return !eq, nil
		}
		return eq, nil
	}
	return l, nil
}

func (p *rpcParser) parseUnary() (any, error) {
	if p.peek().kind == "op" && p.peek().text == "!" {
		p.next()
		v, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return !rpcTruthy(v), nil
	}
	return p.parsePrimary()
}

func (p *rpcParser) parsePrimary() (any, error) {
	tok := p.next()
	switch tok.kind {
	case "lp":
		v, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if p.next().kind != "rp" {
			return nil, fmt.Errorf("missing closing parenthesis")
		}
		return v, nil
	case "str":
		return tok.text, nil
	case "ident":
		if p.peek().kind == "lp" {
			p.next()
			var args []any
			for p.peek().kind != "rp" {
				a, err := p.parseOr()
				if err != nil {
					return nil, err
				}
				args = append(args, a)
				if p.peek().kind == "comma" {
					p.next()
				}
			}
			p.next()
			return p.call(tok.text, args)
		}
		switch tok.text {
		case "true":
			return true, nil
		case "false":
			return false, nil
		}
		if b, ok := p.ctx.bools[tok.text]; ok {
			return b, nil
		}
		return p.ctx.vars[tok.text], nil
	}
	return nil, fmt.Errorf("unexpected token %q", tok.text)
}

func (p *rpcParser) call(name string, args []any) (any, error) {
	switch name {
	case "always":
		return true, nil
	case "cancelled":
		return p.ctx.cancelled, nil
	case "success":
		return p.ctx.success, nil
	case "failure":
		return p.ctx.failure, nil
	case "startsWith":
		if len(args) == 2 {
			return strings.HasPrefix(strings.ToLower(rpcAsString(args[0])), strings.ToLower(rpcAsString(args[1]))), nil
		}
	}
	return nil, fmt.Errorf("unsupported function %s(%d args)", name, len(args))
}

func rpcStripExpr(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "${{") && strings.HasSuffix(s, "}}") {
		s = strings.TrimSpace(s[3 : len(s)-2])
	}
	return s
}

var rpcStatusFuncRe = regexp.MustCompile(`\b(always|cancelled|success|failure)\s*\(`)

func rpcHasStatusFunc(expr string) bool { return rpcStatusFuncRe.MatchString(expr) }

// rpcEvalValueE evaluates an expression and returns an error for syntax or a
// function the model does not implement.
func rpcEvalValueE(expr string, ctx rpcExprCtx) (any, error) {
	toks, err := rpcLex(expr)
	if err != nil {
		return nil, fmt.Errorf("evaluate %q: %w", expr, err)
	}
	p := &rpcParser{toks: toks, ctx: ctx}
	v, err := p.parseOr()
	if err == nil && p.peek().kind != "eof" {
		err = fmt.Errorf("trailing token %q", p.peek().text)
	}
	if err != nil {
		return nil, fmt.Errorf("evaluate %q: %w", expr, err)
	}
	return v, nil
}

var rpcInterpRe = regexp.MustCompile(`\$\{\{\s*(.+?)\s*\}\}`)

// rpcInterpolateE substitutes every ${{ expr }} in text with its string value,
// reporting an expression the model cannot evaluate instead of aborting.
func rpcInterpolateE(text string, ctx rpcExprCtx) (string, error) {
	var firstErr error
	out := rpcInterpRe.ReplaceAllStringFunc(text, func(m string) string {
		inner := rpcInterpRe.FindStringSubmatch(m)[1]
		v, err := rpcEvalValueE(inner, ctx)
		if err != nil && firstErr == nil {
			firstErr = err
		}
		return rpcAsString(v)
	})
	return out, firstErr
}

// withStatusDefault applies GitHub's rule: an `if:` carrying no status function
// is implicitly wrapped in success(); an absent `if:` is success().
func rpcWithStatusDefault(cond string) string {
	cond = rpcStripExpr(cond)
	if cond == "" {
		return "success()"
	}
	if !rpcHasStatusFunc(cond) {
		return "success() && (" + cond + ")"
	}
	return cond
}

// rpcJobRuns reports whether a job starts: its `if:` evaluated with success()
// meaning every needed job concluded success.
func rpcJobRuns(t *testing.T, j *rpcJob, ctx rpcExprCtx) bool {
	t.Helper()
	ok, err := rpcJobRunsE(j, ctx)
	if err != nil {
		t.Fatal(err)
	}
	return ok
}

// rpcJobRunsE is rpcJobRuns reporting an unevaluable condition as an error.
func rpcJobRunsE(j *rpcJob, ctx rpcExprCtx) (bool, error) {
	jc := ctx
	// A cancelled run starts no further job unless its condition says always():
	// success() is false once the run is cancelled.
	jc.success, jc.failure = !ctx.cancelled, false
	for _, n := range j.needsList() {
		res, ok := ctx.vars["needs."+n+".result"]
		if !ok {
			continue
		}
		if res != "success" {
			jc.success = false
		}
		if res == "failure" {
			jc.failure = true
		}
	}
	v, err := rpcEvalValueE(rpcWithStatusDefault(j.If), jc)
	return rpcTruthy(v), err
}

// rpcInterpolate substitutes every ${{ expr }} in text with its string value.
func rpcInterpolate(t *testing.T, text string, ctx rpcExprCtx) string {
	t.Helper()
	out, err := rpcInterpolateE(text, ctx)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// ---- executing a step's script ------------------------------------------------
//
// The verdict and marker steps are not read, they are RUN (sync-audit iter4 B1).
// An earlier revision tried to decide "does this script fail" by reading the shell
// text; every reader is out-parsed by a spelling bash accepts (`trap 'exit 0' EXIT`,
// an indented heredoc terminator, an `exit()` function, `exec true`). The authority
// is now the exit status of the real script under the shell GitHub would use.

// rpcRequireBash skips on windows (the ubuntu legs run the executed guards, the
// same precedent as the cleanup guards) and fails, not skips, when bash is absent
// elsewhere: an absent tool must never turn a guard into a silent pass.
func rpcRequireBash(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the verdict guards execute the workflow's bash; not exercised on windows (the verdict is platform-independent and the ubuntu legs run it)")
	}
	for _, tool := range []string{"bash", "sh"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Fatalf("%s is required by the verdict guard and is not on PATH (absence fails, it does not skip)", tool)
		}
	}
}

// rpcContinuesOnError reports whether a `continue-on-error:` value lets the unit
// fail without failing its parent. Anything that is not provably false counts as
// true, because an expression cannot be evaluated here and the safe direction for
// a guard is to assume the failure is swallowed.
func rpcContinuesOnError(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case string:
		return strings.ToLower(strings.TrimSpace(rpcStripExpr(x))) != "false"
	}
	return true
}

// rpcShellSemanticEnv are the environment variables that change how bash
// executes a script. A step that sets one is refused rather than guessed at.
var rpcShellSemanticEnv = map[string]bool{
	"BASH_ENV": true, "ENV": true, "SHELLOPTS": true, "BASHOPTS": true, "PATH": true,
	"IFS": true, "PS4": true, "PROMPT_COMMAND": true, "LD_PRELOAD": true, "LD_LIBRARY_PATH": true,
}

// rpcEffectiveShell is the shell a step runs under: its own `shell:`, else the
// job's `defaults.run.shell`, else the workflow's.
func rpcEffectiveShell(wf *rpcWorkflow, j *rpcJob, s rpcStep) string {
	for _, sh := range []string{s.Shell, j.Defaults.Run.Shell, wf.Defaults.Run.Shell} {
		if strings.TrimSpace(sh) != "" {
			return strings.TrimSpace(sh)
		}
	}
	return ""
}

// rpcExecResult is the outcome of running one step script.
type rpcExecResult struct {
	failed bool
	// why is non-empty when the script could not be executed faithfully; the
	// caller reports the step as unmodelled, which no verdict cell accepts.
	why string
}

var rpcExecCache sync.Map // key string -> rpcExecResult

// rpcExecScript runs a step script under the named workflow shell and reports
// whether it exits non-zero. Only the shells `bash`, `sh` and the unset default
// are executed; any other shell string (a custom `bash -c "..."` template,
// pwsh, python) is not modelled and is refused.
func rpcExecScript(shell, script string, env map[string]string) rpcExecResult {
	var argv []string
	switch shell {
	case "":
		argv = []string{"bash", "-e", ""} // GitHub's default on Linux and macOS
	case "bash":
		argv = []string{"bash", "--noprofile", "--norc", "-eo", "pipefail", ""}
	case "sh":
		argv = []string{"sh", "-e", ""}
	default:
		return rpcExecResult{why: fmt.Sprintf("shell %q is not executed by this guard", shell)}
	}
	keys := make([]string, 0, len(env))
	for k := range env {
		if rpcShellSemanticEnv[strings.ToUpper(k)] {
			return rpcExecResult{why: fmt.Sprintf("env %s changes how the shell executes the script", k)}
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var kb strings.Builder
	for _, k := range keys {
		kb.WriteString(k + "=" + env[k] + "\x00")
	}
	cacheKey := shell + "\x01" + script + "\x01" + kb.String()
	if v, ok := rpcExecCache.Load(cacheKey); ok {
		return v.(rpcExecResult)
	}
	dir, err := os.MkdirTemp("", "rpc-step-")
	if err != nil {
		return rpcExecResult{why: "cannot create a scratch directory: " + err.Error()}
	}
	defer func() { _ = os.RemoveAll(dir) }()
	path := filepath.Join(dir, "step.sh")
	if err := os.WriteFile(path, []byte(script), 0o644); err != nil {
		return rpcExecResult{why: "cannot write the step script: " + err.Error()}
	}
	argv[len(argv)-1] = path
	tctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(tctx, argv[0], argv[1:]...)
	cmd.Dir = dir
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + dir,
		"GITHUB_STEP_SUMMARY=" + filepath.Join(dir, "summary"),
		"GITHUB_OUTPUT=" + filepath.Join(dir, "output"),
		"GITHUB_ENV=" + filepath.Join(dir, "env"),
		"GITHUB_PATH=" + filepath.Join(dir, "path"),
	}
	for _, k := range keys {
		cmd.Env = append(cmd.Env, k+"="+env[k])
	}
	res := rpcExecResult{}
	if runErr := cmd.Run(); runErr != nil {
		var ee *exec.ExitError
		switch {
		case tctx.Err() != nil:
			res.why = "the step script did not finish within 30s"
		case errors.As(runErr, &ee):
			res.failed = true
		default:
			res.why = "cannot start " + argv[0] + ": " + runErr.Error()
		}
	}
	rpcExecCache.Store(cacheKey, res)
	return res
}

var rpcNeedsResultRe = regexp.MustCompile(`(?i)needs\.[a-z0-9_-]+\.result`)

// rpcIsClassificationStep names the steps whose condition reads a needed job's
// result: the classification verdict of a real job and both steps of a marker.
// These are EXECUTED; every other run step of a gated job is a verification
// step (go build, go vet, ...) the guard cannot run and simulates instead. A step
// that drops its needs.*.result condition therefore stops being executed and
// reads as an always-passing verification step, which the cells then refuse.
func rpcIsClassificationStep(s rpcStep) bool {
	return s.Run != "" && rpcNeedsResultRe.MatchString(s.If)
}

// rpcRunClassificationStep executes a classification step with the expressions it
// carries substituted from the cell context, exactly as the runner would, and
// with the workflow, job and step env applied.
func rpcRunClassificationStep(wf *rpcWorkflow, j *rpcJob, s rpcStep, ctx rpcExprCtx) rpcExecResult {
	script, err := rpcInterpolateE(s.Run, ctx)
	if err != nil {
		return rpcExecResult{why: err.Error()}
	}
	env := map[string]string{"GITHUB_EVENT_NAME": ctx.vars["github.event_name"]}
	for _, src := range []map[string]any{wf.Env, j.Env, s.Env} {
		for k, v := range src {
			val, err := rpcInterpolateE(fmt.Sprint(v), ctx)
			if err != nil {
				return rpcExecResult{why: err.Error()}
			}
			env[k] = val
		}
	}
	return rpcExecScript(rpcEffectiveShell(wf, j, s), script, env)
}

// rpcFirstVerificationStep names the first plain run step of a job that is
// neither conditional nor a classification step: the step the "verification
// fails" variants of a cell fail.
func rpcFirstVerificationStep(j *rpcJob) string {
	for _, s := range j.Steps {
		if s.Run != "" && s.If == "" && !rpcIsClassificationStep(s) {
			return s.Name
		}
	}
	return ""
}

// rpcStepMode selects which verification step a simulation fails.
type rpcStepMode int

const (
	rpcAllStepsPass rpcStepMode = iota
	rpcVerificationFails
)

// rpcUnmodelled is the conclusion of a job the guard could not simulate
// faithfully. It equals no real conclusion, so every cell that expects
// success or failure reads it as a violation and prints the reason.
func rpcUnmodelled(why string) string { return "unmodelled(" + why + ")" }

// rpcSimulateJob walks a job's steps and returns its conclusion. Step `if:` is
// evaluated, classification steps are executed, a step with `continue-on-error`
// cannot fail the job, and a verification step fails only when the mode says so.
func rpcSimulateJob(t *testing.T, wf *rpcWorkflow, j *rpcJob, ctx rpcExprCtx, mode rpcStepMode) string {
	t.Helper()
	failName := ""
	if mode == rpcVerificationFails {
		failName = rpcFirstVerificationStep(j)
	}
	prevFailed := false
	for _, s := range j.Steps {
		sc := ctx
		sc.success, sc.failure = !prevFailed && !ctx.cancelled, prevFailed
		run := sc.success
		if strings.TrimSpace(s.If) != "" {
			v, err := rpcEvalValueE(rpcWithStatusDefault(s.If), sc)
			if err != nil {
				return rpcUnmodelled(err.Error())
			}
			run = rpcTruthy(v)
		}
		if !run {
			continue
		}
		failed := false
		switch {
		case rpcIsClassificationStep(s):
			res := rpcRunClassificationStep(wf, j, s, ctx)
			if res.why != "" {
				return rpcUnmodelled(fmt.Sprintf("step %q: %s", s.Name, res.why))
			}
			failed = res.failed
		case failName != "" && s.Name == failName:
			failed = true
		}
		if failed && !rpcContinuesOnError(s.ContinueOnError) {
			prevFailed = true
		}
	}
	if prevFailed {
		return "failure"
	}
	return "success"
}

// rpcConclusions returns, for one required check name, the conclusion of every
// job leg publishing it ("skipped" when the job does not start).
func rpcConclusions(t *testing.T, wf *rpcWorkflow, name string, ctx rpcExprCtx, mode rpcStepMode) []string {
	t.Helper()
	var out []string
	for _, id := range wf.jobIDs() {
		j := wf.Jobs[id]
		legs := 0
		for _, n := range j.displayNames() {
			if n == name {
				legs++
			}
		}
		if legs == 0 {
			continue
		}
		conclusion := "skipped"
		runs, err := rpcJobRunsE(j, ctx)
		switch {
		case err != nil:
			conclusion = rpcUnmodelled(fmt.Sprintf("job %s condition: %v", id, err))
		case runs:
			conclusion = rpcSimulateJob(t, wf, j, ctx, mode)
		}
		for i := 0; i < legs; i++ {
			out = append(out, conclusion)
		}
	}
	return out
}

// rpcStartedSuppliers returns the ids of the jobs of wf that publish the check
// name and whose own condition starts them in ctx. It reads no step: it answers
// "is a check run published under this name", not "does it pass".
func rpcStartedSuppliers(wf *rpcWorkflow, name string, ctx rpcExprCtx) (started []string, errs []string) {
	for _, id := range wf.jobIDs() {
		j := wf.Jobs[id]
		if !rpcContains(j.displayNames(), name) {
			continue
		}
		runs, err := rpcJobRunsE(j, ctx)
		switch {
		case err != nil:
			errs = append(errs, fmt.Sprintf("job %s of %s: its condition cannot be evaluated (%v)", id, wf.Path, err))
		case runs:
			started = append(started, id)
		}
	}
	return started, errs
}

// rpcNonSkipped drops the "skipped" entries: a skipped job publishes no
// satisfying check run (plan.md B2), so it does not count toward a name.
func rpcNonSkipped(in []string) []string {
	var out []string
	for _, c := range in {
		if c != "skipped" {
			out = append(out, c)
		}
	}
	return out
}

// rpcAllEqual reports whether got is non-empty, all entries equal want, and (when
// wantN > 0) has exactly wantN entries.
func rpcAllEqual(got []string, want string, wantN int) bool {
	if len(got) == 0 || (wantN > 0 && len(got) != wantN) {
		return false
	}
	for _, g := range got {
		if g != want {
			return false
		}
	}
	return true
}

// ---- classification cells ---------------------------------------------------

// rpcClassKeys are the outputs of ci.yml's `detect` job the cost jobs read.
var rpcClassKeys = []string{"go_code", "binary_inputs", "browser_inputs", "harness_inputs"}

// rpcCell is one (event, classification outcome, run state) context.
type rpcCell struct {
	event     string
	detect    string // success | failure | cancelled
	classes   map[string]bool
	cancelled bool
}

func (c rpcCell) ctx() rpcExprCtx {
	vars := map[string]string{
		"github.event_name":   c.event,
		"needs.detect.result": c.detect,
	}
	for _, k := range rpcClassKeys {
		v := ""
		if c.detect == "success" {
			v = strconv.FormatBool(c.classes[k])
		}
		vars["needs.detect.outputs."+k] = v
	}
	return rpcExprCtx{vars: vars, cancelled: c.cancelled}
}

// rpcOutcome returns the class outputs for a named classification outcome.
func rpcOutcome(kind string) map[string]bool {
	out := map[string]bool{}
	switch kind {
	case "docs-only":
	case "go-code":
		for _, k := range rpcClassKeys {
			out[k] = true
		}
	case "harness-only":
		// harness roots live inside go_code (REQ-RPC-004), so go_code is true too.
		out["go_code"] = true
		out["harness_inputs"] = true
	default:
		panic("unknown outcome " + kind)
	}
	return out
}

// rpcOnly returns an outcome where exactly one class output is true.
func rpcOnly(class string) map[string]bool { return map[string]bool{class: true} }
