// codex_review_ownership_m4_test.go pins the M4 slice of the review-ownership
// work: which agents hold the on-demand self-review tools, where the lane's
// card-review stage sits in the distributed kanban doctrine, and that the
// distributed workflow template mentions the tree-scope policy key only as a
// commented example.
//
// Every checker takes document text and returns the list of problems it
// found, so the same function judges the real files (expecting none) and a
// table of deliberately broken variants (expecting a named problem). A check
// that has never been seen to fail on a known bad input proves nothing
// (verification-completeness §1.1); the mutant tables are that observation.
package template_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const (
	coCodexReview = "mcp__moai__codex_review"
	coGLMReview   = "mcp__moai__glm_review"
	coCodexAudit  = "mcp__moai__codex_audit"
	coGLMAudit    = "mcp__moai__glm_audit"

	coCardReviewPath = ".moai/reports/<card-id>/card-review.md"
)

var (
	coReviewHolders = []string{"manager-develop", "manager-docs", "manager-lead"}
	coAuditHolders  = []string{"plan-auditor", "sync-auditor"}
)

// --- agent tool lists ---

// coReadAgents returns agent name -> file content for every *.md in dir.
func coReadAgents(t *testing.T, dir string) map[string]string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	out := map[string]string{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		out[strings.TrimSuffix(e.Name(), ".md")] = string(raw)
	}
	if len(out) == 0 {
		t.Fatalf("no agent definitions found in %s — empty sweep", dir)
	}
	return out
}

// coToolSet parses the `tools:` line of an agent definition's frontmatter.
func coToolSet(md string) map[string]bool {
	set := map[string]bool{}
	for _, line := range strings.Split(md, "\n") {
		if strings.HasPrefix(line, "tools:") {
			for _, tok := range strings.Split(strings.TrimPrefix(line, "tools:"), ",") {
				if tok = strings.TrimSpace(tok); tok != "" {
					set[tok] = true
				}
			}
			break
		}
	}
	return set
}

func coHolders(agents map[string]string, tool string) []string {
	var holders []string
	for name, md := range agents {
		if coToolSet(md)[tool] {
			holders = append(holders, name)
		}
	}
	sort.Strings(holders)
	return holders
}

// coHolderProblems checks the two holder sets: the review tools sit with the
// three lane-side agents only, the audit tools with the two auditors only.
func coHolderProblems(agents map[string]string) []string {
	var problems []string
	check := func(tool string, want []string) {
		got := coHolders(agents, tool)
		if strings.Join(got, ",") != strings.Join(want, ",") {
			problems = append(problems, fmt.Sprintf("%s holders = %v, want exactly %v", tool, got, want))
		}
	}
	check(coCodexReview, coReviewHolders)
	check(coGLMReview, coReviewHolders)
	check(coCodexAudit, coAuditHolders)
	check(coGLMAudit, coAuditHolders)
	return problems
}

// coParityProblems requires the local and distributed copies of each agent to
// declare the same tool set.
func coParityProblems(c1, c2 map[string]string) []string {
	var problems []string
	names := map[string]bool{}
	for n := range c1 {
		names[n] = true
	}
	for n := range c2 {
		names[n] = true
	}
	for n := range names {
		a, aok := c1[n]
		b, bok := c2[n]
		if !aok || !bok {
			problems = append(problems, fmt.Sprintf("%s exists in only one of the local and distributed copies", n))
			continue
		}
		sa, sb := coToolSet(a), coToolSet(b)
		for tool := range sa {
			if !sb[tool] {
				problems = append(problems, fmt.Sprintf("%s: %s is listed locally but not in the distributed copy", n, tool))
			}
		}
		for tool := range sb {
			if !sa[tool] {
				problems = append(problems, fmt.Sprintf("%s: %s is listed in the distributed copy but not locally", n, tool))
			}
		}
	}
	sort.Strings(problems)
	return problems
}

// coWithTool returns agents with the tool appended to / removed from one
// agent's `tools:` line. ok is false when the edit changed nothing.
func coWithTool(agents map[string]string, agent, tool string, add bool) (map[string]string, bool) {
	out := map[string]string{}
	for k, v := range agents {
		out[k] = v
	}
	md, found := out[agent]
	if !found {
		return out, false
	}
	lines := strings.Split(md, "\n")
	for i, line := range lines {
		if !strings.HasPrefix(line, "tools:") {
			continue
		}
		switch {
		case add && !coToolSet(md)[tool]:
			lines[i] = line + ", " + tool
		case !add && coToolSet(md)[tool]:
			lines[i] = strings.Replace(line, ", "+tool, "", 1)
		default:
			return out, false
		}
		out[agent] = strings.Join(lines, "\n")
		return out, out[agent] != md
	}
	return out, false
}

// --- doctrine ---

// coSection returns the text from the line that equals heading to the next
// heading of the same or a higher level (or the end of the document).
func coSection(doc, heading string) string {
	level := len(heading) - len(strings.TrimLeft(heading, "#"))
	lines := strings.Split(doc, "\n")
	start := -1
	for i, l := range lines {
		if strings.TrimRight(l, " ") == heading {
			start = i
			break
		}
	}
	if start < 0 {
		return ""
	}
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		l := lines[i]
		if !strings.HasPrefix(l, "#") {
			continue
		}
		if n := len(l) - len(strings.TrimLeft(l, "#")); n <= level && strings.HasPrefix(l[n:], " ") {
			end = i
			break
		}
	}
	return strings.Join(lines[start:end], "\n")
}

// coSentence returns the sentence around the first occurrence of token: from
// the previous ". " (or the start of the paragraph) to the next ". " (or the
// end of the paragraph).
func coSentence(section, token string) string {
	i := strings.Index(section, token)
	if i < 0 {
		return ""
	}
	start := 0
	if p := strings.LastIndex(section[:i], "\n\n"); p >= 0 {
		start = p + 2
	}
	if p := strings.LastIndex(section[start:i], ". "); p >= 0 {
		start += p + 2
	}
	end := len(section)
	if p := strings.Index(section[i:], "\n\n"); p >= 0 {
		end = i + p
	}
	if p := strings.Index(section[i:end], ". "); p >= 0 {
		end = i + p + 1
	}
	return strings.TrimSpace(section[start:end])
}

const (
	coDetailStageHeading = "## The card-review stage"
	coStubStageHeading   = "### The lane's task list carries the card's stages"
	coStubReadHeading    = "## Completion is read, never trusted"
)

var (
	coStageItem = regexp.MustCompile("(?m)^[0-9]+\\. `\\[([a-z-]+)\\]`")
	coNegation  = regexp.MustCompile("(?i)\\bno card-review\\b|not required|`\\[card-review\\]`[^\n]*\\b(skipped|optional)\\b|may skip the card-review")
	coCeiling   = regexp.MustCompile(`(?i)\bat most (2|two)\b`)
)

// coDetailProblems judges factory-dispatch-detail.md: the ordered stage list
// puts card-review after run-exit verification and before integration, the
// re-review ceiling and the evidence path are written down, and no sentence in
// the section negates the stage.
func coDetailProblems(detail string) []string {
	sec := coSection(detail, coDetailStageHeading)
	if sec == "" {
		return []string{"detail: section " + coDetailStageHeading + " is missing"}
	}
	var problems []string
	var order []string
	for _, m := range coStageItem.FindAllStringSubmatch(sec, -1) {
		order = append(order, m[1])
	}
	if got, want := strings.Join(order, ">"), "run-exit>card-review>integration>report"; got != want {
		problems = append(problems, fmt.Sprintf("detail: ordered stage list = %q, want %q", got, want))
	}
	if !strings.Contains(sec, coCardReviewPath) {
		problems = append(problems, "detail: the card-review evidence path "+coCardReviewPath+" is not stated")
	}
	if !coCeiling.MatchString(sec) {
		problems = append(problems, "detail: the re-review ceiling (at most 2) is not stated")
	}
	if !strings.Contains(strings.ToLower(sec), "advisory") {
		problems = append(problems, "detail: the stage is not described as advisory")
	}
	if loc := coNegation.FindString(sec); loc != "" {
		problems = append(problems, fmt.Sprintf("detail: the stage section carries a negating phrase %q", loc))
	}
	return problems
}

// coStubProblems judges factory-dispatch.md: the stage is named in the lane
// stage section, the leader's conditional sentence sits there, and the leader's
// completion-read section lists the evidence path and the gap rule.
func coStubProblems(stub string) []string {
	var problems []string

	stage := coSection(stub, coStubStageHeading)
	if stage == "" {
		return []string{"stub: section " + coStubStageHeading + " is missing"}
	}
	for _, need := range []string{"`card-review`", coCardReviewPath, "advisory"} {
		if !strings.Contains(stage, need) {
			problems = append(problems, "stub: the lane stage section does not carry "+need)
		}
	}
	if loc := coNegation.FindString(stage); loc != "" {
		problems = append(problems, fmt.Sprintf("stub: the lane stage section carries a negating phrase %q", loc))
	}

	// The leader's sentence is conditional: the tree_scope: skip condition and
	// the missing turn-end gate share one sentence, condition first.
	sent := coSentence(stage, "turn-end codex review gate")
	condIdx := strings.Index(sent, "`tree_scope: skip`")
	gateIdx := strings.Index(sent, "turn-end codex review gate")
	switch {
	case sent == "":
		problems = append(problems, "stub: no sentence in the lane stage section names the turn-end codex review gate")
	case condIdx < 0:
		problems = append(problems, "stub: the leader sentence is unconditional — it does not name `tree_scope: skip`")
	case condIdx > gateIdx:
		problems = append(problems, "stub: the leader sentence names `tree_scope: skip` after the gate — not a condition")
	case !regexp.MustCompile(`^(With|If|Where|While|When) `).MatchString(sent):
		problems = append(problems, "stub: the leader sentence does not open with a condition")
	case !strings.Contains(sent, "directly with the same tools"):
		problems = append(problems, "stub: the leader sentence does not say the leader reviews directly with the same tools")
	}

	read := coSection(stub, coStubReadHeading)
	if read == "" {
		return append(problems, "stub: section "+coStubReadHeading+" is missing")
	}
	rule := coSentence(read, "card-review.md")
	if rule == "" {
		return append(problems, "stub: the completion-read section does not name card-review.md")
	}
	for _, need := range []string{coCardReviewPath, "neither cites", "nor records a reason", "gap", "stays in its stage"} {
		if !strings.Contains(rule, need) {
			problems = append(problems, "stub: the completion-read gap rule lacks "+need)
		}
	}
	if regexp.MustCompile(`(?i)not a gap|need not|is not required`).MatchString(rule) {
		problems = append(problems, "stub: the completion-read gap rule is negated")
	}
	return problems
}

func coRead(t *testing.T, root, rel string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(raw)
}

const (
	coStubRel   = ".claude/rules/moai/workflow/factory-dispatch.md"
	coDetailRel = ".claude/rules/moai/workflow/factory-dispatch-detail.md"
	coTplPrefix = "internal/template/templates/"
)

// --- template workflow.yaml ---

// coWorkflowYAMLProblems judges the distributed workflow template: the
// codex.review_gate block parses to the single live key `enabled`, and the
// tree-scope key appears only inside comment lines (at least once).
func coWorkflowYAMLProblems(text string) []string {
	var problems []string
	var doc struct {
		Workflow struct {
			Codex struct {
				ReviewGate map[string]any `yaml:"review_gate"`
			} `yaml:"codex"`
		} `yaml:"workflow"`
	}
	if err := yaml.Unmarshal([]byte(text), &doc); err != nil {
		return []string{"workflow.yaml does not parse: " + err.Error()}
	}
	keys := make([]string, 0, len(doc.Workflow.Codex.ReviewGate))
	for k := range doc.Workflow.Codex.ReviewGate {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if strings.Join(keys, ",") != "enabled" {
		problems = append(problems, fmt.Sprintf("workflow.codex.review_gate keys = %v, want exactly [enabled]", keys))
	}
	if v, ok := doc.Workflow.Codex.ReviewGate["enabled"]; !ok || v != false {
		problems = append(problems, fmt.Sprintf("workflow.codex.review_gate.enabled = %v, want false", v))
	}
	commented := 0
	for i, line := range strings.Split(text, "\n") {
		if !strings.Contains(line, "tree_scope") {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			commented++
			continue
		}
		problems = append(problems, fmt.Sprintf("line %d carries tree_scope outside a comment: %q", i+1, line))
	}
	if commented == 0 {
		problems = append(problems, "no commented example of tree_scope in the template")
	}
	return problems
}

// --- tests ---

// TestReviewOwnership_ToolHolders — AC-012 (a)(b): the self-review tools sit
// with manager-develop, manager-docs and manager-lead only; the audit tools
// stay with plan-auditor and sync-auditor only; the local and distributed
// copies declare the same tools.
func TestReviewOwnership_ToolHolders(t *testing.T) {
	root := findProjectRootForMirrorTest(t)
	c1 := coReadAgents(t, filepath.Join(root, ".claude/agents/moai"))
	c2 := coReadAgents(t, filepath.Join(root, coTplPrefix+".claude/agents/moai"))
	for label, agents := range map[string]map[string]string{"local (.claude/agents/moai)": c1, "distributed (templates)": c2} {
		for _, p := range coHolderProblems(agents) {
			t.Errorf("%s: %s", label, p)
		}
	}
	for _, p := range coParityProblems(c1, c2) {
		t.Errorf("local/distributed tools drift: %s", p)
	}
}

// TestReviewOwnership_ToolHoldersCodexEmission — AC-012 (c), the emitted layer:
// the audit tools are named in the two auditors' emitted definitions only, and
// no emitted definition outside the three lane-side agents names a review tool.
// The byte-level regeneration check is the agentemit golden test; this pins who
// is allowed to mention what.
func TestReviewOwnership_ToolHoldersCodexEmission(t *testing.T) {
	root := findProjectRootForMirrorTest(t)
	dir := filepath.Join(root, coTplPrefix+".codex/agents/moai")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	holders := map[string][]string{}
	n := 0
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".toml") {
			continue
		}
		n++
		body := coRead(t, root, coTplPrefix+".codex/agents/moai/"+e.Name())
		name := strings.TrimSuffix(e.Name(), ".toml")
		for _, tool := range []string{coCodexAudit, coGLMAudit, coCodexReview, coGLMReview} {
			if strings.Contains(body, tool) {
				holders[tool] = append(holders[tool], name)
			}
		}
	}
	if n == 0 {
		t.Fatal("no emitted agent definitions found — empty sweep")
	}
	for _, tool := range []string{coCodexAudit, coGLMAudit} {
		if got := strings.Join(holders[tool], ","); got != strings.Join(coAuditHolders, ",") {
			t.Errorf("emitted definitions naming %s = %v, want exactly %v", tool, holders[tool], coAuditHolders)
		}
	}
	allowed := map[string]bool{}
	for _, h := range coReviewHolders {
		allowed[h] = true
	}
	for _, tool := range []string{coCodexReview, coGLMReview} {
		for _, h := range holders[tool] {
			if !allowed[h] {
				t.Errorf("emitted definition %s names %s but is not a lane-side holder", h, tool)
			}
		}
	}
}

// TestReviewOwnership_ToolHolderMutants — the holder checker must go red on
// each cheapest way of getting the grant wrong.
func TestReviewOwnership_ToolHolderMutants(t *testing.T) {
	root := findProjectRootForMirrorTest(t)
	base := coReadAgents(t, filepath.Join(root, coTplPrefix+".claude/agents/moai"))
	fourth := "manager-git"
	if _, ok := base[fourth]; !ok {
		t.Fatalf("fixture agent %s is not in the roster", fourth)
	}
	cases := []struct {
		name  string
		agent string
		tool  string
		add   bool
		want  string // substring the problem list must contain
	}{
		{"review tool granted to a fourth agent", fourth, coCodexReview, true, fourth},
		{"glm review granted to an auditor", "sync-auditor", coGLMReview, true, "sync-auditor"},
		{"audit tool granted to a lane agent", "manager-develop", coGLMAudit, true, "manager-develop"},
		{"review tool withheld from a lane agent", "manager-lead", coGLMReview, false, "manager-lead"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// A removal mutant needs the grant to exist; make the base carry it first.
			start := base
			if !c.add {
				full := base
				for _, h := range coReviewHolders {
					for _, tool := range []string{coCodexReview, coGLMReview} {
						full, _ = coWithTool(full, h, tool, true)
					}
				}
				start = full
			}
			mut, changed := coWithTool(start, c.agent, c.tool, c.add)
			if !changed {
				t.Fatalf("mutant did not apply: %s %v %s", c.agent, c.add, c.tool)
			}
			problems := strings.Join(coHolderProblems(mut), "\n")
			if !strings.Contains(problems, c.want) {
				t.Errorf("the holder checker did not flag %q; problems:\n%s", c.want, problems)
			}
		})
	}
	t.Run("distributed copy edited, local copy not", func(t *testing.T) {
		c1 := coReadAgents(t, filepath.Join(root, ".claude/agents/moai"))
		// Flip the grant in the distributed copy only: add it when absent, remove it when present.
		c2, changed := coWithTool(base, "manager-docs", coCodexReview, !coToolSet(base["manager-docs"])[coCodexReview])
		if !changed {
			t.Fatal("mutant did not apply")
		}
		if len(coParityProblems(c1, c2)) == 0 {
			t.Error("a one-sided edit passed the local/distributed parity check")
		}
	})
	t.Run("parity checker flags either direction", func(t *testing.T) {
		a := map[string]string{"x": "tools: Read, mcp__moai__codex_review"}
		b := map[string]string{"x": "tools: Read"}
		if len(coParityProblems(a, b)) == 0 || len(coParityProblems(b, a)) == 0 {
			t.Error("parity checker is blind to a one-sided tools edit")
		}
	})
}

// TestReviewOwnership_CardReviewDoctrine — AC-013 and AC-014 on the local
// copies and the distributed mirrors, plus mirror neutrality.
func TestReviewOwnership_CardReviewDoctrine(t *testing.T) {
	root := findProjectRootForMirrorTest(t)
	for _, rel := range []string{coStubRel, coTplPrefix + coStubRel} {
		for _, p := range coStubProblems(coRead(t, root, rel)) {
			t.Errorf("%s: %s", rel, p)
		}
	}
	for _, rel := range []string{coDetailRel, coTplPrefix + coDetailRel} {
		for _, p := range coDetailProblems(coRead(t, root, rel)) {
			t.Errorf("%s: %s", rel, p)
		}
	}
	// The detail companion is byte-identical between the two trees today and
	// stays so; the stub may differ elsewhere, so only the card-review text is
	// compared there.
	if coRead(t, root, coDetailRel) != coRead(t, root, coTplPrefix+coDetailRel) {
		t.Error("factory-dispatch-detail.md differs between the local and distributed trees")
	}
	for _, h := range []string{coStubStageHeading, coStubReadHeading} {
		if a, b := coSection(coRead(t, root, coStubRel), h), coSection(coRead(t, root, coTplPrefix+coStubRel), h); a != b {
			t.Errorf("stub section %q differs between the local and distributed trees", h)
		}
	}
	// The distributed card-review text carries no SPEC id, card number or date.
	// Only the paragraphs that talk about card-review are judged: the stage
	// section also holds an older paragraph this work does not own.
	for _, c := range []struct{ rel, heading string }{
		{coTplPrefix + coStubRel, coStubStageHeading},
		{coTplPrefix + coStubRel, coStubReadHeading},
		{coTplPrefix + coDetailRel, coDetailStageHeading},
	} {
		sec := coSection(coRead(t, root, c.rel), c.heading)
		if sec == "" {
			continue // reported by the problem checks above
		}
		for _, para := range strings.Split(sec, "\n\n") {
			if !strings.Contains(para, "card-review") && c.rel != coTplPrefix+coDetailRel {
				continue
			}
			if loc := mirrorForkLeakPattern.FindString(para); loc != "" {
				t.Errorf("%s § %s carries an internal token %q in a card-review paragraph", c.rel, c.heading, loc)
			}
		}
	}
}

// TestReviewOwnership_DoctrineMutants — the doctrine checkers must go red on
// the cheapest wrong shapes of the stage text.
func TestReviewOwnership_DoctrineMutants(t *testing.T) {
	root := findProjectRootForMirrorTest(t)
	detail := coRead(t, root, coTplPrefix+coDetailRel)
	stub := coRead(t, root, coTplPrefix+coStubRel)

	apply := func(t *testing.T, doc string, edit func(string) string) string {
		t.Helper()
		out := edit(doc)
		if out == doc {
			t.Fatal("mutant did not apply — the text it edits is absent (RED until the doctrine lands)")
		}
		return out
	}
	swap := func(a, b string) func(string) string {
		return func(s string) string {
			ia := regexp.MustCompile("(?m)^[0-9]+\\. `\\[" + a + "\\]`.*$").FindString(s)
			ib := regexp.MustCompile("(?m)^[0-9]+\\. `\\[" + b + "\\]`.*$").FindString(s)
			if ia == "" || ib == "" {
				return s
			}
			s = strings.Replace(s, ia, "\x00A", 1)
			s = strings.Replace(s, ib, "\x00B", 1)
			s = strings.Replace(s, "\x00A", ib, 1)
			return strings.Replace(s, "\x00B", ia, 1)
		}
	}
	detailMutants := []struct {
		name string
		edit func(string) string
		want string
	}{
		{"stage placed after integration", swap("card-review", "integration"), "ordered stage list"},
		{"stage placed after the report", swap("card-review", "report"), "ordered stage list"},
		{"stage item removed", func(s string) string {
			return regexp.MustCompile("(?m)^[0-9]+\\. `\\[card-review\\]`.*\n").ReplaceAllString(s, "")
		}, "ordered stage list"},
		{"negating sentence in the section", func(s string) string {
			return strings.Replace(s, "Running `[card-review]`:", "A lane with no card-review is acceptable.\n\nRunning `[card-review]`:", 1)
		}, "negating phrase"},
		{"re-review ceiling dropped", func(s string) string { return regexp.MustCompile(`(?i)at most 2`).ReplaceAllString(s, "any number of") }, "ceiling"},
		{"evidence path dropped", func(s string) string { return strings.ReplaceAll(s, coCardReviewPath, ".moai/reports/card-review.md") }, "evidence path"},
	}
	for _, m := range detailMutants {
		t.Run("detail/"+m.name, func(t *testing.T) {
			got := strings.Join(coDetailProblems(apply(t, detail, m.edit)), "\n")
			if !strings.Contains(got, m.want) {
				t.Errorf("not flagged (%q); problems:\n%s", m.want, got)
			}
		})
	}

	stubMutants := []struct {
		name string
		edit func(string) string
		want string
	}{
		{"leader sentence made unconditional", func(s string) string {
			return strings.Replace(s, "With `tree_scope: skip` configured for the leader's checkout, the leader session", "The leader session", 1)
		}, "unconditional"},
		{"condition placed after the gate", func(s string) string {
			return strings.Replace(s, "With `tree_scope: skip` configured for the leader's checkout, the leader session carries no turn-end codex review gate", "The leader session carries no turn-end codex review gate when `tree_scope: skip` is configured for its checkout, and it", 1)
		}, ""},
		{"gap rule negated", func(s string) string {
			return strings.Replace(s, "is a gap and stays in its stage", "is not a gap and need not stay in its stage", 1)
		}, "gap rule"},
		{"gap rule moved out of the completion-read section", func(s string) string {
			i := strings.Index(s, coStubReadHeading)
			if i < 0 {
				return s
			}
			sec := coSection(s, coStubReadHeading)
			rule := coSentence(sec, "card-review.md")
			if rule == "" {
				return s
			}
			moved := strings.Replace(sec, rule, "", 1)
			s = strings.Replace(s, sec, moved, 1)
			return strings.Replace(s, "## Boundaries", rule+"\n\n## Boundaries", 1)
		}, "completion-read"},
		{"evidence path missing from the stage section", func(s string) string {
			sec := coSection(s, coStubStageHeading)
			return strings.Replace(s, sec, strings.ReplaceAll(sec, coCardReviewPath, "the review file"), 1)
		}, coCardReviewPath},
	}
	for _, m := range stubMutants {
		t.Run("stub/"+m.name, func(t *testing.T) {
			got := coStubProblems(apply(t, stub, m.edit))
			if len(got) == 0 {
				t.Fatalf("mutant passed the stub checker")
			}
			if m.want != "" && !strings.Contains(strings.Join(got, "\n"), m.want) {
				t.Errorf("flagged, but not for %q; problems:\n%s", m.want, strings.Join(got, "\n"))
			}
		})
	}
}

// TestReviewOwnership_TemplateWorkflowKeyIsCommentOnly — AC-015 (iii)(iv): the
// distributed workflow template mentions tree_scope only in a comment, and the
// surfaces that enumerate live keys do not know it.
func TestReviewOwnership_TemplateWorkflowKeyIsCommentOnly(t *testing.T) {
	root := findProjectRootForMirrorTest(t)
	text := coRead(t, root, coTplPrefix+".moai/config/sections/workflow.yaml")
	for _, p := range coWorkflowYAMLProblems(text) {
		t.Error(p)
	}

	inv := coRead(t, root, "internal/config/testdata/shipped_key_inventory.yaml")
	// Positive control: the inventory is read and still enumerates the
	// neighbouring review_gate keys, so the absence check below sweeps a
	// non-empty set.
	if !strings.Contains(inv, "review_gate") {
		t.Errorf("shipped_key_inventory.yaml carries no review_gate key — the absence check below would sweep nothing")
	}
	for _, rel := range []string{
		"internal/config/testdata/shipped_key_inventory.yaml",
		"internal/settings/schema_sections.go",
		"internal/web/fieldsets.templ",
		"internal/web/assets/i18n.js",
	} {
		if strings.Contains(coRead(t, root, rel), "tree_scope") {
			t.Errorf("%s mentions tree_scope — the key ships as a template comment only", rel)
		}
	}
}

// TestReviewOwnership_TemplateWorkflowMutants — the template checker must go
// red on each way of shipping the key live.
func TestReviewOwnership_TemplateWorkflowMutants(t *testing.T) {
	root := findProjectRootForMirrorTest(t)
	text := coRead(t, root, coTplPrefix+".moai/config/sections/workflow.yaml")
	anchor := "        review_gate:\n            enabled: false\n"
	if !strings.Contains(text, anchor) {
		t.Fatal("template review_gate block not found at the expected shape")
	}
	cases := []struct {
		name string
		edit func(string) string
		want string
	}{
		{"live key added", func(s string) string {
			return strings.Replace(s, anchor, anchor+"            tree_scope: review\n", 1)
		}, "keys ="},
		{"live key beside a comment example", func(s string) string {
			return strings.Replace(s, anchor, anchor+"            # tree_scope: review\n            tree_scope: skip\n", 1)
		}, "outside a comment"},
		{"example removed", func(s string) string {
			var keep []string
			for _, l := range strings.Split(s, "\n") {
				if !strings.Contains(l, "tree_scope") {
					keep = append(keep, l)
				}
			}
			return strings.Join(keep, "\n")
		}, "no commented example"},
		{"enabled flipped on", func(s string) string {
			return strings.Replace(s, anchor, "        review_gate:\n            enabled: true\n", 1)
		}, "enabled"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mut := c.edit(text)
			if mut == text && c.name != "example removed" {
				t.Fatal("mutant did not apply")
			}
			got := strings.Join(coWorkflowYAMLProblems(mut), "\n")
			if !strings.Contains(got, c.want) {
				t.Errorf("not flagged (%q); problems:\n%s", c.want, got)
			}
		})
	}
}
