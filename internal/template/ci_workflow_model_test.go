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
// modelled as failing only when the test says so or when its script reaches an
// unconditional top-level `exit 1` (rpcScriptHardFails: not inside a condition,
// loop or group, not behind `&&` / `||`, not after an earlier top-level `exit`);
// shell behaviour inside a `run:` body is otherwise not executed. Job-level `if:` and step-level `if:` are evaluated; `needs` is
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

// triggerPathFilters returns the `paths:` and `paths-ignore:` filters an event
// declares (present=false when the workflow does not subscribe to the event).
// GitHub applies them only to push and pull_request events.
func (w *rpcWorkflow) triggerPathFilters(event string) (present bool, paths, ignore []string) {
	if w.On.Kind != yaml.MappingNode {
		present, _ = w.triggerBranches(event)
		return present, nil, nil
	}
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
	for i := 0; i+1 < len(w.On.Content); i += 2 {
		if w.On.Content[i].Value != event {
			continue
		}
		val := w.On.Content[i+1]
		if val.Kind != yaml.MappingNode {
			return true, nil, nil
		}
		for k := 0; k+1 < len(val.Content); k += 2 {
			switch val.Content[k].Value {
			case "paths":
				paths = list(val.Content[k+1])
			case "paths-ignore":
				ignore = list(val.Content[k+1])
			}
		}
		return true, paths, ignore
	}
	return false, nil, nil
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

func rpcEvalValue(t *testing.T, expr string, ctx rpcExprCtx) any {
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
	return v
}

func rpcEvalIf(t *testing.T, expr string, ctx rpcExprCtx) bool {
	t.Helper()
	return rpcTruthy(rpcEvalValue(t, expr, ctx))
}

var rpcInterpRe = regexp.MustCompile(`\$\{\{\s*(.+?)\s*\}\}`)

// rpcInterpolate substitutes every ${{ expr }} in text with its string value.
func rpcInterpolate(t *testing.T, text string, ctx rpcExprCtx) string {
	t.Helper()
	return rpcInterpRe.ReplaceAllStringFunc(text, func(m string) string {
		inner := rpcInterpRe.FindStringSubmatch(m)[1]
		return rpcAsString(rpcEvalValue(t, inner, ctx))
	})
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

// rpcShellStatements splits a `run:` script into statements, each a list of
// words. It is a reader for the straight-line scripts the verdict and marker
// steps use, not a shell: it understands quoting, comments, `;` and newline as
// statement ends, a trailing `&&` / `||` / `|` / backslash as a continuation,
// `$( ... )` as one word and heredoc bodies as skipped text. `&&`, `||`, `|` and
// `&` are kept as words so a conditional `cmd || exit 1` stays visibly one
// statement.
func rpcShellStatements(script string) [][]string {
	var stmts [][]string
	var cur []string
	var word strings.Builder
	inWord := false
	heredoc, heredocNext := "", false
	flushWord := func() {
		if !inWord {
			return
		}
		w := word.String()
		word.Reset()
		inWord = false
		switch {
		case heredocNext:
			heredoc, heredocNext = strings.Trim(w, `'"`), false
		case strings.HasPrefix(w, "<<") && !strings.HasPrefix(w, "<<<"):
			d := strings.Trim(strings.TrimPrefix(strings.TrimPrefix(w, "<<"), "-"), `'"`)
			if d == "" {
				heredocNext = true
			} else {
				heredoc = d
			}
		}
		cur = append(cur, w)
	}
	endStmt := func() {
		flushWord()
		if len(cur) > 0 {
			stmts = append(stmts, cur)
			cur = nil
		}
	}
	rs := []rune(script)
	sub := 0
	for i := 0; i < len(rs); i++ {
		c := rs[i]
		switch {
		case sub > 0:
			word.WriteRune(c)
			inWord = true
			switch c {
			case '(':
				sub++
			case ')':
				sub--
			}
		case c == '\\':
			if i+1 < len(rs) {
				i++
				if rs[i] == '\n' {
					flushWord()
				} else {
					word.WriteRune(rs[i])
					inWord = true
				}
			}
		case c == '\'':
			inWord = true
			for i++; i < len(rs) && rs[i] != '\''; i++ {
				word.WriteRune(rs[i])
			}
		case c == '"':
			inWord = true
			for i++; i < len(rs) && rs[i] != '"'; i++ {
				if rs[i] == '\\' && i+1 < len(rs) {
					i++
				}
				word.WriteRune(rs[i])
			}
		case c == '$' && i+1 < len(rs) && rs[i+1] == '(':
			word.WriteString("$(")
			inWord = true
			i++
			sub = 1
		case c == '#' && !inWord:
			for i+1 < len(rs) && rs[i+1] != '\n' {
				i++
			}
		case c == '\n':
			flushWord()
			if heredoc != "" {
				j := i + 1
				for j < len(rs) {
					k := j
					for k < len(rs) && rs[k] != '\n' {
						k++
					}
					line := strings.TrimSpace(string(rs[j:k]))
					j = k + 1
					if line == heredoc {
						break
					}
				}
				i, heredoc = j-1, ""
			}
			if n := len(cur); n > 0 && (cur[n-1] == "&&" || cur[n-1] == "||" || cur[n-1] == "|") {
				continue
			}
			endStmt()
		case c == ';':
			endStmt()
		case c == ' ' || c == '\t' || c == '\r':
			flushWord()
		case c == '&' || c == '|':
			flushWord()
			op := string(c)
			if i+1 < len(rs) && rs[i+1] == c {
				op += string(c)
				i++
			}
			cur = append(cur, op)
		default:
			word.WriteRune(c)
			inWord = true
		}
	}
	endStmt()
	return stmts
}

// rpcScriptHardFails reports whether running the script reaches an
// unconditional `exit 1`: a top-level statement that is exactly `exit 1`, not
// preceded by another top-level `exit`, and not inside an if / for / while /
// until / case body, a `{ }` group or a `( )` subshell, nor chained behind
// `&&` / `||`. An `exit 1` that sits inside a condition is NOT unconditional
// (this reader does not evaluate the condition, so it never credits one — an
// `if false; then exit 1; fi` and an `exit 0` followed by `exit 1` both read as
// "does not hard-fail"). That is the conservative direction for a guard: a
// verdict step that is wrapped in a condition turns the guard red.
func rpcScriptHardFails(script string) bool {
	depth := 0
	for _, st := range rpcShellStatements(script) {
		for len(st) > 0 && (st[0] == "then" || st[0] == "do" || st[0] == "else") {
			st = st[1:]
		}
		if len(st) == 0 {
			continue
		}
		switch st[0] {
		case "if", "for", "while", "until", "case", "select":
			depth++
		case "fi", "done", "esac":
			if depth > 0 {
				depth--
			}
		case "exit":
			if depth == 0 && len(st) <= 2 {
				return len(st) == 2 && st[1] == "1"
			}
		}
		for _, w := range st {
			switch w {
			case "{", "(":
				depth++
			case "}", ")":
				if depth > 0 {
					depth--
				}
			}
		}
	}
	return false
}

// rpcStepHardFails reports whether a step's script reaches an unconditional
// `exit 1` — the convention the verdict and marker steps use.
func rpcStepHardFails(s rpcStep) bool { return rpcScriptHardFails(s.Run) }

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
