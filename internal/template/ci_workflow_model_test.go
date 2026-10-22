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
// Modelling limits, stated so a green is read at its true width: a step is
// modelled as failing only when the test says so or when its script ends in an
// unconditional `exit 1` line; shell behaviour inside a `run:` body is not
// executed. Job-level `if:` and step-level `if:` are evaluated; `needs` is
// modelled through the `needs.<id>.result` / `.outputs.<name>` context values
// the test supplies.
package template_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"unicode"

	"gopkg.in/yaml.v3"
)

// rpcStep is one workflow step.
type rpcStep struct {
	ID             string         `yaml:"id"`
	Name           string         `yaml:"name"`
	If             string         `yaml:"if"`
	Uses           string         `yaml:"uses"`
	Run            string         `yaml:"run"`
	With           map[string]any `yaml:"with"`
	Env            map[string]any `yaml:"env"`
	TimeoutMinutes yaml.Node      `yaml:"timeout-minutes"`
}

// rpcJob is one workflow job.
type rpcJob struct {
	ID             string `yaml:"-"`
	Name           string `yaml:"name"`
	If             string `yaml:"if"`
	Needs          yaml.Node
	TimeoutMinutes yaml.Node `yaml:"timeout-minutes"`
	Outputs        map[string]string
	Strategy       struct {
		Matrix map[string]any `yaml:"matrix"`
	} `yaml:"strategy"`
	Steps []rpcStep `yaml:"steps"`
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
	Jobs map[string]*rpcJob `yaml:"jobs"`
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

// triggerBranches reports whether the workflow subscribes to an event and the
// branch filter it declares (empty = no filter).
func (w *rpcWorkflow) triggerBranches(event string) (bool, []string) {
	switch w.On.Kind {
	case yaml.ScalarNode:
		return w.On.Value == event, nil
	case yaml.SequenceNode:
		for _, c := range w.On.Content {
			if c.Value == event {
				return true, nil
			}
		}
		return false, nil
	case yaml.MappingNode:
		for i := 0; i+1 < len(w.On.Content); i += 2 {
			if w.On.Content[i].Value != event {
				continue
			}
			val := w.On.Content[i+1]
			if val.Kind != yaml.MappingNode {
				return true, nil
			}
			for k := 0; k+1 < len(val.Content); k += 2 {
				if val.Content[k].Value != "branches" {
					continue
				}
				bn := val.Content[k+1]
				if bn.Kind == yaml.ScalarNode {
					return true, []string{bn.Value}
				}
				out := make([]string, 0, len(bn.Content))
				for _, c := range bn.Content {
					out = append(out, c.Value)
				}
				return true, out
			}
			return true, nil
		}
	}
	return false, nil
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
	vars      map[string]string
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
		l = rpcTruthy(l) || rpcTruthy(r)
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
		l = rpcTruthy(l) && rpcTruthy(r)
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

func rpcEvalIf(t *testing.T, expr string, ctx rpcExprCtx) bool {
	t.Helper()
	toks, err := rpcLex(expr)
	if err != nil {
		t.Fatalf("evaluate %q: %v", expr, err)
	}
	p := &rpcParser{toks: toks, ctx: ctx}
	v, err := p.parseOr()
	if err == nil && p.peek().kind != "eof" {
		err = fmt.Errorf("trailing token %q", p.peek().text)
	}
	if err != nil {
		t.Fatalf("evaluate %q: %v", expr, err)
	}
	return rpcTruthy(v)
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
	return rpcEvalIf(t, rpcWithStatusDefault(j.If), jc)
}

var rpcHardFailRe = regexp.MustCompile(`(?m)^\s*exit 1\s*$`)

// rpcStepHardFails reports whether a step's script ends in an unconditional
// `exit 1` line — the convention the verdict and marker steps use.
func rpcStepHardFails(s rpcStep) bool { return rpcHardFailRe.MatchString(s.Run) }

// rpcFirstVerificationStep names the first plain run step of a job that is
// neither conditional nor a hard-fail step: the step the "verification fails"
// variants of a cell fail.
func rpcFirstVerificationStep(j *rpcJob) string {
	for _, s := range j.Steps {
		if s.Run != "" && s.If == "" && !rpcStepHardFails(s) {
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

// rpcSimulateJob walks a job's steps and returns its conclusion.
func rpcSimulateJob(t *testing.T, j *rpcJob, ctx rpcExprCtx, mode rpcStepMode) string {
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
			run = rpcEvalIf(t, rpcWithStatusDefault(s.If), sc)
		}
		if !run {
			continue
		}
		if (failName != "" && s.Name == failName) || rpcStepHardFails(s) {
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
		if rpcJobRuns(t, j, ctx) {
			conclusion = rpcSimulateJob(t, j, ctx, mode)
		}
		for i := 0; i < legs; i++ {
			out = append(out, conclusion)
		}
	}
	return out
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
