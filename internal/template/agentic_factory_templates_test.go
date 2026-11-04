// agentic_factory_templates_test.go: structural guards for the Agentic Factory
// GitHub and GitLab product templates (issue forms, issue/PR/MR templates,
// reusable comment sources, index files).
//
// Scope and honesty notes, because these tests are conservative checks and
// not renderers:
//   - No Mermaid renderer exists in this environment, so the diagram check is
//     a structural one (header, direction, statement grammar, balanced
//     labels). It does NOT prove a diagram renders.
//   - No Pele renderer exists either. The tests pin the visible placeholder
//     line and the absence of any generated SVG, so a fake asset cannot slip
//     in; they never assert a rendered result.
//   - Everything is read from the on-disk template tree. The embedded-FS and
//     deploy checks live in agentic_factory_embed_deploy_test.go.
//
// Sentinels on failure: AF_TEMPLATE_INVENTORY / AF_TEMPLATE_FORM /
// AF_TEMPLATE_CONTENT / AF_TEMPLATE_MERMAID / AF_TEMPLATE_PELE /
// AF_TEMPLATE_LEAK / AF_TEMPLATE_HARDCODE.
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

// afTemplatesRoot is the template source root relative to this package dir.
const afTemplatesRoot = "templates"

// Canonical sentences. Every template that accepts free text carries the
// redaction rule and the comment-style rule verbatim (whitespace-normalised).
const (
	afRedaction = "Redaction rule: never write tokens, API keys, credentials or cookies; personal data (email addresses, personal names, user-identifying local paths); absolute home-directory paths (use repository-relative paths); raw logs, environment dumps or stack traces (write a short summary and link an accessible artifact); or content copied from private remotes or private repositories. If sensitive content was already posted, stop and report it to the operator instead of silently editing history."

	afCommentStyle = "Comment style: summary and delta only. Say what changed since the previous revision, link the evidence, and name the next action. Do not repeat the full Issue body, earlier comments or logs, and post nothing when nothing substantive changed."

	afRateLimit = "Rate-limit stance: when a provider budget is exhausted or a request is refused with 401, 403 or 429, wait for the provider-stated reset or retry-after time and record the blocker. Never work around it with a new key, token or account, and never post extra comments to compensate."

	afPeleLine = "Pele SVG: not generated (renderer unavailable) — see GAP-PELE-SVG"
)

// afTemplate is one of the 17 DELTA2 templates and its per-provider homes.
type afTemplate struct {
	ID      string
	Slug    string
	Diagram bool // DELTA2 requires a Mermaid source and a Pele SVG
	Ledger  bool // carries the context ledger section
	Rate    bool // carries the rate-limit stance
	// Kind is "issue" (issue form / issue template), "handoff" (PR / MR
	// template) or "comment" (no repository-file surface on either provider;
	// the file is a reusable comment source).
	Kind string
}

func (a afTemplate) key() string { return a.ID + "-" + a.Slug }

func (a afTemplate) githubPath() string {
	switch a.Kind {
	case "issue":
		return ".github/ISSUE_TEMPLATE/agentic-factory-" + a.key() + ".yml"
	case "handoff":
		return ".github/PULL_REQUEST_TEMPLATE/agentic-factory-" + a.key() + ".md"
	}
	return ".github/agentic-factory/comment-templates/" + a.key() + ".md"
}

func (a afTemplate) gitlabPath() string {
	switch a.Kind {
	case "issue":
		return ".gitlab/issue_templates/agentic-factory-" + a.key() + ".md"
	case "handoff":
		return ".gitlab/merge_request_templates/agentic-factory-" + a.key() + ".md"
	}
	return ".gitlab/agentic-factory/comment-templates/" + a.key() + ".md"
}

// afTemplates is the closed list of the 17 templates (DELTA2 section 1).
var afTemplates = []afTemplate{
	{ID: "01", Slug: "parent-issue", Diagram: true, Ledger: true, Rate: true, Kind: "issue"},
	{ID: "02", Slug: "child-issue", Diagram: true, Ledger: true, Kind: "issue"},
	{ID: "03", Slug: "spec-change", Diagram: true, Ledger: true, Kind: "issue"},
	{ID: "04", Slug: "dispatch", Rate: true, Kind: "comment"},
	{ID: "05", Slug: "progress", Ledger: true, Kind: "comment"},
	{ID: "06", Slug: "check-run", Kind: "comment"},
	{ID: "07", Slug: "review", Kind: "comment"},
	{ID: "08", Slug: "audit", Ledger: true, Kind: "comment"},
	{ID: "09", Slug: "repair", Diagram: true, Kind: "comment"},
	{ID: "10", Slug: "block-approval", Rate: true, Kind: "comment"},
	{ID: "11", Slug: "handoff", Ledger: true, Kind: "handoff"},
	{ID: "12", Slug: "merge", Kind: "comment"},
	{ID: "13", Slug: "done", Ledger: true, Kind: "comment"},
	{ID: "14", Slug: "recovery", Diagram: true, Rate: true, Kind: "comment"},
	{ID: "15", Slug: "final-report", Diagram: true, Ledger: true, Kind: "comment"},
	{ID: "16", Slug: "delta-sync", Rate: true, Kind: "comment"},
	{ID: "17", Slug: "ack", Kind: "comment"},
}

// Non-template files of the feature, relative to the templates root.
const (
	afGitHubIndex    = ".github/agentic-factory/README.md"
	afGitLabIndex    = ".gitlab/agentic-factory/README.md"
	afSupportMatrix  = ".moai/docs/agentic-factory-provider-support.md"
	afProposedLabels = ".github/agentic-factory/labels.proposed.yml"
)

// afExpectedFiles is the closed set of files the feature adds.
func afExpectedFiles() []string {
	files := []string{afGitHubIndex, afGitLabIndex, afSupportMatrix, afProposedLabels}
	for _, a := range afTemplates {
		files = append(files, a.githubPath(), a.gitlabPath())
	}
	sort.Strings(files)
	return files
}

// afNormalise collapses all whitespace runs so wrapped or indented copies of a
// canonical sentence compare equal.
func afNormalise(s string) string { return strings.Join(strings.Fields(s), " ") }

func afReadTemplate(t *testing.T, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(afTemplatesRoot, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("AF_TEMPLATE_INVENTORY: read %s: %v", rel, err)
	}
	return string(data)
}

// TestAgenticFactoryInventoryComplete: every one of the 17 templates has a
// non-empty file on both providers, and the on-disk feature tree matches the
// closed expected set exactly (nothing missing, nothing extra).
func TestAgenticFactoryInventoryComplete(t *testing.T) {
	t.Parallel()
	if len(afTemplates) != 17 {
		t.Fatalf("AF_TEMPLATE_INVENTORY: the contract defines 17 templates, table has %d", len(afTemplates))
	}
	for _, rel := range afExpectedFiles() {
		data, err := os.ReadFile(filepath.Join(afTemplatesRoot, filepath.FromSlash(rel)))
		if err != nil {
			t.Errorf("AF_TEMPLATE_INVENTORY: missing %s: %v", rel, err)
			continue
		}
		if strings.TrimSpace(string(data)) == "" {
			t.Errorf("AF_TEMPLATE_INVENTORY: %s is empty", rel)
		}
	}

	want := map[string]bool{}
	for _, rel := range afExpectedFiles() {
		want[rel] = true
	}
	for _, root := range []string{".github/ISSUE_TEMPLATE", ".github/PULL_REQUEST_TEMPLATE", ".github/agentic-factory", ".gitlab"} {
		base := filepath.Join(afTemplatesRoot, filepath.FromSlash(root))
		_ = filepath.WalkDir(base, func(p string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(afTemplatesRoot, p)
			rel = filepath.ToSlash(rel)
			if !want[rel] {
				t.Errorf("AF_TEMPLATE_INVENTORY: unexpected file %s (add it to the expected set deliberately)", rel)
			}
			return nil
		})
	}
}

// TestAgenticFactoryIndexMapsAll17: each index file lists every template id
// with its provider file, and states the "no native home" case explicitly for
// the comment templates.
func TestAgenticFactoryIndexMapsAll17(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		index string
		path  func(afTemplate) string
	}{
		{afGitHubIndex, afTemplate.githubPath},
		{afGitLabIndex, afTemplate.gitlabPath},
	} {
		text := afReadTemplate(t, tc.index)
		for _, a := range afTemplates {
			if !strings.Contains(text, a.key()) {
				t.Errorf("AF_TEMPLATE_INVENTORY: %s does not list template %s", tc.index, a.key())
			}
			base := filepath.Base(tc.path(a))
			if !strings.Contains(text, base) {
				t.Errorf("AF_TEMPLATE_INVENTORY: %s does not map %s to its file %s", tc.index, a.key(), base)
			}
			if a.Kind == "comment" {
				row := regexp.MustCompile(`(?m)^\|[^\n]*` + regexp.QuoteMeta(base) + `[^\n]*no native home[^\n]*$`)
				if !row.MatchString(text) {
					t.Errorf("AF_TEMPLATE_INVENTORY: %s row for %s must say \"no native home\"", tc.index, a.key())
				}
			}
		}
		for _, must := range []string{"GAP-PELE-SVG", "Mermaid", "labels"} {
			if !strings.Contains(text, must) {
				t.Errorf("AF_TEMPLATE_INVENTORY: %s must mention %q", tc.index, must)
			}
		}
	}
	gh := afReadTemplate(t, afGitHubIndex)
	// the config.yml decision is recorded where a maintainer will look
	if !strings.Contains(gh, "config.yml") || !strings.Contains(gh, "not shipped") {
		t.Errorf("AF_TEMPLATE_INVENTORY: %s must record the decision to not ship an issue-chooser config.yml", afGitHubIndex)
	}
	if _, err := os.Stat(filepath.Join(afTemplatesRoot, ".github", "ISSUE_TEMPLATE", "config.yml")); err == nil {
		t.Errorf("AF_TEMPLATE_INVENTORY: .github/ISSUE_TEMPLATE/config.yml must not be shipped (decision recorded in the index)")
	}
}

// ---------------------------------------------------------------------------
// GitHub issue forms
// ---------------------------------------------------------------------------

var afFormIDRe = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

type afFormElement struct {
	Type        string         `yaml:"type"`
	ID          string         `yaml:"id"`
	Attributes  map[string]any `yaml:"attributes"`
	Validations map[string]any `yaml:"validations"`
}

type afForm struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	Title       string          `yaml:"title"`
	Body        []afFormElement `yaml:"body"`
}

// TestAgenticFactoryIssueFormsValid checks the three issue forms against
// GitHub's documented issue-form schema (top-level keys and element rules).
// Labels, assignees, projects and type are deliberately absent: a form that
// applied them would mutate remote state outside the approval flow.
func TestAgenticFactoryIssueFormsValid(t *testing.T) {
	t.Parallel()
	names := map[string]string{}
	for _, a := range afTemplates {
		if a.Kind != "issue" {
			continue
		}
		rel := a.githubPath()
		raw := afReadTemplate(t, rel)

		var top map[string]any
		if err := yaml.Unmarshal([]byte(raw), &top); err != nil {
			t.Errorf("AF_TEMPLATE_FORM: %s does not parse as YAML: %v", rel, err)
			continue
		}
		for _, banned := range []string{"labels", "assignees", "projects", "type"} {
			if _, ok := top[banned]; ok {
				t.Errorf("AF_TEMPLATE_FORM: %s sets top-level %q; forms must not apply remote state", rel, banned)
			}
		}
		for _, k := range []string{"name", "description", "body"} {
			if _, ok := top[k]; !ok {
				t.Errorf("AF_TEMPLATE_FORM: %s missing required top-level key %q", rel, k)
			}
		}

		var form afForm
		if err := yaml.Unmarshal([]byte(raw), &form); err != nil {
			t.Errorf("AF_TEMPLATE_FORM: %s does not decode: %v", rel, err)
			continue
		}
		if prev, dup := names[form.Name]; dup {
			t.Errorf("AF_TEMPLATE_FORM: form name %q is used by %s and %s; names must be unique", form.Name, prev, rel)
		}
		names[form.Name] = rel

		if len(form.Body) == 0 {
			t.Errorf("AF_TEMPLATE_FORM: %s has an empty body", rel)
		}
		ids := map[string]bool{}
		inputs := 0
		for i, el := range form.Body {
			where := fmt.Sprintf("%s body[%d] (%s)", rel, i, el.Type)
			switch el.Type {
			case "markdown":
				if v, _ := el.Attributes["value"].(string); strings.TrimSpace(v) == "" {
					t.Errorf("AF_TEMPLATE_FORM: %s needs a non-empty attributes.value", where)
				}
				if el.ID != "" {
					t.Errorf("AF_TEMPLATE_FORM: %s must not carry an id (markdown is not submitted)", where)
				}
			case "textarea", "input", "dropdown", "checkboxes":
				inputs++
				if l, _ := el.Attributes["label"].(string); strings.TrimSpace(l) == "" {
					t.Errorf("AF_TEMPLATE_FORM: %s needs attributes.label", where)
				}
				if el.ID != "" {
					if !afFormIDRe.MatchString(el.ID) {
						t.Errorf("AF_TEMPLATE_FORM: %s id %q has characters outside [A-Za-z0-9_-]", where, el.ID)
					}
					if ids[el.ID] {
						t.Errorf("AF_TEMPLATE_FORM: %s duplicate id %q", where, el.ID)
					}
					ids[el.ID] = true
				}
				if el.Type == "dropdown" || el.Type == "checkboxes" {
					opts, _ := el.Attributes["options"].([]any)
					if len(opts) == 0 {
						t.Errorf("AF_TEMPLATE_FORM: %s needs a non-empty options list", where)
					}
					if el.Type == "checkboxes" {
						for j, o := range opts {
							m, ok := o.(map[string]any)
							if !ok || m["label"] == nil || strings.TrimSpace(fmt.Sprint(m["label"])) == "" {
								t.Errorf("AF_TEMPLATE_FORM: %s option %d needs a label", where, j)
							}
						}
					}
				}
			default:
				t.Errorf("AF_TEMPLATE_FORM: %s has unsupported element type %q", where, el.Type)
			}
		}
		if inputs == 0 {
			t.Errorf("AF_TEMPLATE_FORM: %s needs at least one non-markdown element", rel)
		}
	}
	if len(names) != 3 {
		t.Errorf("AF_TEMPLATE_FORM: expected 3 issue forms, found %d", len(names))
	}
}

// TestAgenticFactoryMarkdownTemplatesShape: every template carries the
// machine-metadata marker; Markdown templates also carry the common field
// table and a title line. The form carries the marker as a YAML comment.
func TestAgenticFactoryMarkdownTemplatesShape(t *testing.T) {
	t.Parallel()
	for _, a := range afTemplates {
		marker := "moai-template:" + a.key() + "@1.0.0"
		for _, rel := range []string{a.githubPath(), a.gitlabPath()} {
			text := afReadTemplate(t, rel)
			if !strings.Contains(text, marker) {
				t.Errorf("AF_TEMPLATE_CONTENT: %s lacks marker %q", rel, marker)
			}
			if !strings.Contains(text, "never authorization") {
				t.Errorf("AF_TEMPLATE_CONTENT: %s marker must state that machine metadata is correlation only, never authorization", rel)
			}
			if strings.HasSuffix(rel, ".md") {
				for _, need := range []string{"| Common field | Value |", "actor / role / session_id", "head_sha / base_sha / spec_hash", "current state / next action / owner"} {
					if !strings.Contains(text, need) {
						t.Errorf("AF_TEMPLATE_CONTENT: %s lacks common-field row %q", rel, need)
					}
				}
				if !regexp.MustCompile(`(?m)^# `).MatchString(text) {
					t.Errorf("AF_TEMPLATE_CONTENT: %s has no top-level title", rel)
				}
			}
		}
	}
}

// TestAgenticFactoryRequiredStatements pins the contract statements every
// template must carry: redaction rule, comment style; and the subset-specific
// statements (rate-limit stance, context ledger).
func TestAgenticFactoryRequiredStatements(t *testing.T) {
	t.Parallel()
	redaction := afNormalise(afRedaction)
	style := afNormalise(afCommentStyle)
	rate := afNormalise(afRateLimit)
	ledgerTerms := []string{
		"commit-pinned link", "artifact commit", "artifact sha256", "target head",
		"owner", "verification state", "verified_at",
		"STALE", "inaccessible-local", "inaccessible-unpushed", "LINK_BLOCKED",
	}
	for _, a := range afTemplates {
		for _, rel := range []string{a.githubPath(), a.gitlabPath()} {
			text := afNormalise(afReadTemplate(t, rel))
			if !strings.Contains(text, redaction) {
				t.Errorf("AF_TEMPLATE_CONTENT: %s lacks the canonical redaction rule", rel)
			}
			if !strings.Contains(text, style) {
				t.Errorf("AF_TEMPLATE_CONTENT: %s lacks the canonical comment-style rule", rel)
			}
			if a.Rate && !strings.Contains(text, rate) {
				t.Errorf("AF_TEMPLATE_CONTENT: %s lacks the canonical rate-limit stance", rel)
			}
			if a.Ledger {
				for _, term := range ledgerTerms {
					if !strings.Contains(text, term) {
						t.Errorf("AF_TEMPLATE_CONTENT: %s ledger section lacks %q", rel, term)
					}
				}
			}
		}
	}
}

// TestAgenticFactoryLifecycleVocabulary: one Issue / multiple tasks / plan,
// run, sync; agent handoff after the PR; leader closure; asynchronous
// preparation stages.
func TestAgenticFactoryLifecycleVocabulary(t *testing.T) {
	t.Parallel()
	byID := map[string]afTemplate{}
	for _, a := range afTemplates {
		byID[a.ID] = a
	}
	need := map[string][]string{
		"01": {"one active agent", "child Issue"},
		"02": {"one Issue", "tasks", "plan", "run", "sync", "ACCEPTED", "ISSUE_CONFIRMED", "BRANCH_REGISTERED", "AGENT_RESERVED", "WT_PREPARED", "READY_ACKED"},
		"04": {"ACCEPTED", "ISSUE_CONFIRMED", "BRANCH_REGISTERED", "AGENT_RESERVED", "WT_PREPARED", "READY_ACKED", "no development launch before READY_ACKED"},
		"11": {"handoff", "leader", "closes", "worktree", "preserved", "quiescence", "closing keyword"},
		"12": {"leader", "merge"},
		"13": {"leader", "closes the Issue"},
		"15": {"leader"},
	}
	for id, terms := range need {
		a := byID[id]
		for _, rel := range []string{a.githubPath(), a.gitlabPath()} {
			lower := strings.ToLower(afReadTemplate(t, rel))
			for _, term := range terms {
				if !strings.Contains(lower, strings.ToLower(term)) {
					t.Errorf("AF_TEMPLATE_CONTENT: %s lacks lifecycle term %q", rel, term)
				}
			}
		}
	}
	// the preparation stage order is a contract: ACK -> Issue published ->
	// branch+worktree READY -> start. Each index must show the order.
	for _, idx := range []string{afGitHubIndex, afGitLabIndex} {
		text := afReadTemplate(t, idx)
		i1 := strings.Index(text, "ACCEPTED")
		i2 := strings.Index(text, "ISSUE_CONFIRMED")
		i3 := strings.Index(text, "READY_ACKED")
		if i1 < 0 || i2 < i1 || i3 < i2 {
			t.Errorf("AF_TEMPLATE_CONTENT: %s must show the preparation stages in order ACCEPTED, ISSUE_CONFIRMED, ..., READY_ACKED", idx)
		}
	}
}

// ---------------------------------------------------------------------------
// Mermaid (structural only) and the Pele placeholder
// ---------------------------------------------------------------------------

const afMermaidNode = `[A-Za-z][A-Za-z0-9_]*(?:\[[^\[\]\(\)\{\}"|<>]+\]|\{[^\[\]\(\)\{\}"|<>]+\}|\([^\[\]\(\)\{\}"|<>]+\))?`

var (
	afMermaidHeaderRe = regexp.MustCompile(`^(flowchart|graph) TD$`)
	afMermaidStmtRe   = regexp.MustCompile(`^` + afMermaidNode + `(?:\s*(?:-->|---|-\.->|==>)(?:\|[^|]+\|)?\s*` + afMermaidNode + `)*$`)
	afMermaidIDRe     = regexp.MustCompile(`[A-Za-z][A-Za-z0-9_]*`)
	afMermaidDeclRe   = regexp.MustCompile(`([A-Za-z][A-Za-z0-9_]*)[\[\{\(][^\]\}\)]*[\]\}\)]`)
	afMermaidLabelRe  = regexp.MustCompile(`[\[\{\(][^\]\}\)]*[\]\}\)]`)
	afMermaidPipeRe   = regexp.MustCompile(`\|[^|]*\|`)
)

// afExtractMermaid returns the body lines of every ```mermaid fence in text
// and whether every fence was closed. Leading indentation (YAML block
// scalars) is ignored.
func afExtractMermaid(text string) (blocks [][]string, closed bool) {
	closed = true
	var cur []string
	in := false
	for _, line := range strings.Split(text, "\n") {
		tl := strings.TrimSpace(line)
		switch {
		case !in && strings.HasPrefix(tl, "```mermaid"):
			in, cur = true, nil
		case in && tl == "```":
			blocks = append(blocks, cur)
			in = false
		case in:
			cur = append(cur, tl)
		}
	}
	if in {
		closed = false
	}
	return blocks, closed
}

// afCheckMermaid is the conservative structural check. It returns problems.
func afCheckMermaid(lines []string) []string {
	var problems []string
	var body []string
	for _, l := range lines {
		if l == "" || strings.HasPrefix(l, "%%") {
			continue
		}
		body = append(body, l)
	}
	if len(body) < 2 {
		return []string{"fewer than two lines"}
	}
	if !afMermaidHeaderRe.MatchString(body[0]) {
		problems = append(problems, fmt.Sprintf("header %q is not \"flowchart TD\" or \"graph TD\" (TD is the only allowed direction)", body[0]))
	}
	declared := map[string]bool{}
	referenced := map[string]bool{}
	edges := 0
	for _, l := range body[1:] {
		if !afMermaidStmtRe.MatchString(l) {
			problems = append(problems, fmt.Sprintf("statement %q does not match the conservative node/edge grammar", l))
			continue
		}
		if strings.Contains(l, "-->") || strings.Contains(l, "---") || strings.Contains(l, "-.->") || strings.Contains(l, "==>") {
			edges++
		}
		for _, m := range afMermaidDeclRe.FindAllStringSubmatch(l, -1) {
			declared[m[1]] = true
		}
		stripped := afMermaidLabelRe.ReplaceAllString(l, "")
		stripped = afMermaidPipeRe.ReplaceAllString(stripped, "")
		for _, id := range afMermaidIDRe.FindAllString(stripped, -1) {
			referenced[id] = true
		}
	}
	if edges == 0 {
		problems = append(problems, "no edge")
	}
	if len(declared) < 2 {
		problems = append(problems, "fewer than two labelled nodes")
	}
	for id := range referenced {
		if !declared[id] {
			problems = append(problems, fmt.Sprintf("node %q is referenced but never given a label", id))
		}
	}
	return problems
}

// TestAgenticFactoryMermaidBlocks: the six diagram templates carry at least
// one structurally plausible TD-only Mermaid source on both providers; no
// template anywhere carries a non-TD direction.
func TestAgenticFactoryMermaidBlocks(t *testing.T) {
	t.Parallel()
	directionRe := regexp.MustCompile(`(?m)^\s*(flowchart|graph)\s+(LR|RL|BT|TB)\b`)
	for _, a := range afTemplates {
		for _, rel := range []string{a.githubPath(), a.gitlabPath()} {
			text := afReadTemplate(t, rel)
			if directionRe.MatchString(text) {
				t.Errorf("AF_TEMPLATE_MERMAID: %s uses a direction other than TD", rel)
			}
			blocks, closed := afExtractMermaid(text)
			if !closed {
				t.Errorf("AF_TEMPLATE_MERMAID: %s has an unclosed mermaid fence", rel)
			}
			if a.Diagram && len(blocks) == 0 {
				t.Errorf("AF_TEMPLATE_MERMAID: %s must carry a Mermaid source block", rel)
			}
			for i, b := range blocks {
				for _, p := range afCheckMermaid(b) {
					t.Errorf("AF_TEMPLATE_MERMAID: %s block %d: %s", rel, i, p)
				}
			}
		}
	}
}

// TestAgenticFactoryMermaidCheckerRejects keeps the structural checker honest:
// it must reject the shapes it claims to reject.
func TestAgenticFactoryMermaidCheckerRejects(t *testing.T) {
	t.Parallel()
	cases := map[string][]string{
		"LR direction":     {"flowchart LR", "A[x] --> B[y]"},
		"unlabelled node":  {"flowchart TD", "A[x] --> B"},
		"unbalanced label": {"flowchart TD", "A[x --> B[y]"},
		"no edge":          {"flowchart TD", "A[x]", "B[y]"},
		"bad statement":    {"flowchart TD", "A[x] ==> ==> B[y]"},
	}
	for name, lines := range cases {
		if len(afCheckMermaid(lines)) == 0 {
			t.Errorf("AF_TEMPLATE_MERMAID: checker accepted %s", name)
		}
	}
	if p := afCheckMermaid([]string{"flowchart TD", "N0[Start]", "N1[Next]", "N0 --> N1"}); len(p) != 0 {
		t.Errorf("AF_TEMPLATE_MERMAID: checker rejected a valid diagram: %v", p)
	}
}

// TestAgenticFactoryPelePlaceholder: where a Pele SVG is required the exact
// visible placeholder is present, no generated SVG/PNG asset exists in the
// feature tree, no template pretends to embed one, and the docs define the
// gap as unmet.
func TestAgenticFactoryPelePlaceholder(t *testing.T) {
	t.Parallel()
	for _, a := range afTemplates {
		for _, rel := range []string{a.githubPath(), a.gitlabPath()} {
			text := afReadTemplate(t, rel)
			has := strings.Contains(text, afPeleLine)
			if a.Diagram && !has {
				t.Errorf("AF_TEMPLATE_PELE: %s must carry the exact placeholder line %q", rel, afPeleLine)
			}
			if !a.Diagram && has {
				t.Errorf("AF_TEMPLATE_PELE: %s carries a Pele placeholder but no diagram is required here", rel)
			}
			for _, fake := range []string{"<svg", "<picture", ".svg)", ".svg\"", ".svg ", "![", "data:image"} {
				if strings.Contains(text, fake) {
					t.Errorf("AF_TEMPLATE_PELE: %s embeds an image/SVG marker %q; no SVG has been generated", rel, fake)
				}
			}
		}
	}
	for _, root := range []string{".github", ".gitlab", ".moai/docs"} {
		_ = filepath.WalkDir(filepath.Join(afTemplatesRoot, filepath.FromSlash(root)), func(p string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(p))
			if ext == ".svg" || ext == ".png" {
				t.Errorf("AF_TEMPLATE_PELE: %s is a rendered image asset; the renderer is unavailable, so none may exist", p)
			}
			return nil
		})
	}
	for _, idx := range []string{afGitHubIndex, afGitLabIndex, afSupportMatrix} {
		if !strings.Contains(afReadTemplate(t, idx), "GAP-PELE-SVG") {
			t.Errorf("AF_TEMPLATE_PELE: %s must mention GAP-PELE-SVG", idx)
		}
	}
	gap := afReadTemplate(t, afSupportMatrix)
	if !regexp.MustCompile(`(?s)GAP-PELE-SVG.{0,400}(unmet|not met)`).MatchString(gap) {
		t.Errorf("AF_TEMPLATE_PELE: %s must record GAP-PELE-SVG as unmet", afSupportMatrix)
	}
}

// ---------------------------------------------------------------------------
// Leak scan, language, hardcode, quick actions
// ---------------------------------------------------------------------------

// afAllFeatureTexts is every file the feature ships.
func afAllFeatureTexts(t *testing.T) map[string]string {
	t.Helper()
	out := map[string]string{}
	for _, rel := range afExpectedFiles() {
		out[rel] = afReadTemplate(t, rel)
	}
	return out
}

// afLeakPatterns builds the scanner patterns from fragments so this test
// file itself holds no credential-shaped literal.
func afLeakPatterns() map[string]*regexp.Regexp {
	dash5 := strings.Repeat("-", 5)
	homeUnix := "/" + "Users" + "/"
	homeLinux := "/" + "home" + "/"
	winUsers := "C:" + `\` + "Users"
	return map[string]*regexp.Regexp{
		"github token":     regexp.MustCompile(`\bgh[pousr]_[A-Za-z0-9]{20,}\b|github_pat_[A-Za-z0-9_]{20,}`),
		"gitlab token":     regexp.MustCompile(`\bglpat-[A-Za-z0-9_-]{10,}`),
		"aws key":          regexp.MustCompile(`\bAK` + `IA[0-9A-Z]{16}\b`),
		"slack token":      regexp.MustCompile(`\bxox[abprs]-[A-Za-z0-9-]{10,}`),
		"private key":      regexp.MustCompile(dash5 + `BEGIN [A-Z ]*PRIVATE KEY` + dash5),
		"bearer value":     regexp.MustCompile(`(?i)\bbearer\s+[A-Za-z0-9._~+/=-]{20,}`),
		"credential value": regexp.MustCompile(`(?i)\b(api[_-]?key|secret|token|password)\s*[:=]\s*["']?[A-Za-z0-9_\-]{16,}`),
		"email address":    regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`),
		"unix home path":   regexp.MustCompile(regexp.QuoteMeta(homeUnix) + `[A-Za-z0-9._-]+`),
		"linux home path":  regexp.MustCompile(regexp.QuoteMeta(homeLinux) + `[A-Za-z0-9._<-]+`),
		"windows user dir": regexp.MustCompile(regexp.QuoteMeta(winUsers)),
		"file url":         regexp.MustCompile(`file://`),
	}
}

// TestAgenticFactoryNoLeaks: no token shapes, private keys, email addresses,
// absolute home paths, or Windows user paths in any shipped file.
func TestAgenticFactoryNoLeaks(t *testing.T) {
	t.Parallel()
	patterns := afLeakPatterns()
	for rel, text := range afAllFeatureTexts(t) {
		for name, re := range patterns {
			if loc := re.FindStringIndex(text); loc != nil {
				t.Errorf("AF_TEMPLATE_LEAK: %s contains a %s: %q", rel, name, text[loc[0]:min(loc[1], loc[0]+40)])
			}
		}
	}
}

// Synthetic fake values, assembled at runtime to avoid false positives from external secret scanners; not real credentials.
// TestAgenticFactoryLeakScanCatchesSamples proves the scanner is not vacuous.
// This is the only function in the change that builds credential-shaped
// strings.
func TestAgenticFactoryLeakScanCatchesSamples(t *testing.T) {
	t.Parallel()
	patterns := afLeakPatterns()
	samples := map[string]string{
		"github token":     "ghp_" + strings.Repeat("a", 30),
		"email address":    "person" + "@" + "example.com",
		"unix home path":   "/" + "Users" + "/someone/project",
		"linux home path":  "/" + "home" + "/someone/project",
		"private key":      strings.Repeat("-", 5) + "BEGIN RSA PRIVATE KEY" + strings.Repeat("-", 5),
		"windows user dir": "C:" + `\` + "Users" + `\someone`,
		"file url":         "file:" + "///tmp/x",
		"credential value": "api_key = " + strings.Repeat("a1", 12),
	}
	for name, s := range samples {
		if !patterns[name].MatchString(s) {
			t.Errorf("AF_TEMPLATE_LEAK: scanner pattern %q missed its sample", name)
		}
	}
}

// TestAgenticFactoryEnglishOnly: shipped templates are English; locale
// variants are a recorded gap, not a silent mix.
func TestAgenticFactoryEnglishOnly(t *testing.T) {
	t.Parallel()
	re := regexp.MustCompile(`[\x{AC00}-\x{D7A3}\x{3040}-\x{30FF}\x{4E00}-\x{9FFF}]`)
	for rel, text := range afAllFeatureTexts(t) {
		if loc := re.FindStringIndex(text); loc != nil {
			t.Errorf("AF_TEMPLATE_CONTENT: %s contains non-English script text near %q (locale variants are a recorded gap)", rel, text[loc[0]:min(loc[1]+10, len(text))])
		}
	}
}

// TestAgenticFactoryNoRemoteSideEffects: GitLab quick actions in a template
// description would be executed by the provider when an Issue or MR is
// created. None may appear.
func TestAgenticFactoryNoRemoteSideEffects(t *testing.T) {
	t.Parallel()
	quick := regexp.MustCompile(`(?m)^\s*/(label|unlabel|relabel|assign|unassign|reassign|milestone|remove_milestone|epic|remove_epic|parent_epic|child_epic|relate|unlink|close|reopen|due|remove_due_date|estimate|spend|copy_metadata|move|clone|merge|rebase|approve|unapprove|target_branch|draft|ready|reviewer|iteration|weight|promote|blocks|blocked_by|duplicate|lock|unlock|confidential|title|submit_review|subscribe|unsubscribe|todo|done|publish|create_merge_request|set_parent|add_child)\b`)
	for rel, text := range afAllFeatureTexts(t) {
		if quick.MatchString(text) {
			t.Errorf("AF_TEMPLATE_CONTENT: %s contains a provider quick action; templates must apply nothing remotely", rel)
		}
	}
}

// afRateNumberRes find rate-limit style numbers in prose: a number bound to
// a request/point/CPU unit, a per-time-unit rate, or an "at least N unit"
// wait. The provider schema draft is the only home for such numbers.
var afRateNumberRes = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b\d[\d,]*\s*(requests?|points?|req)\b`),
	regexp.MustCompile(`(?i)\b\d[\d,]*\s*(each|per|/|every)\s*(\d+\s*)?(hour|minute|second|min|sec)s?\b`),
	regexp.MustCompile(`(?i)\bat least (one|1|\d+)\s*(second|minute)s?\b`),
	regexp.MustCompile(`(?i)\b\d[\d,]*\s*seconds? of cpu\b`),
	regexp.MustCompile(`(?i)\b\d[\d,]*\b[^.\n]{0,24}\b(per|each|every) (hour|minute|second)\b`),
}

// TestAgenticFactoryNoRateLimitNumbersInProse: documented limit values live
// in the provider schema draft, never in template prose.
func TestAgenticFactoryNoRateLimitNumbersInProse(t *testing.T) {
	t.Parallel()
	for rel, text := range afAllFeatureTexts(t) {
		for _, re := range afRateNumberRes {
			if loc := re.FindStringIndex(text); loc != nil {
				t.Errorf("AF_TEMPLATE_HARDCODE: %s states a rate-limit number in prose: %q (limits belong to the provider schema)", rel, text[loc[0]:loc[1]])
			}
		}
	}
	// the detector itself must fire on known shapes
	for _, s := range []string{"5,000 requests", "900 points", "10 per minute", "80 content requests per minute", "wait at least one second", "90 seconds of CPU"} {
		hit := false
		for _, re := range afRateNumberRes {
			if re.MatchString(s) {
				hit = true
			}
		}
		if !hit {
			t.Errorf("AF_TEMPLATE_HARDCODE: detector missed %q", s)
		}
	}
}
