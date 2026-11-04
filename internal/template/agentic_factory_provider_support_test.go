// agentic_factory_provider_support_test.go: the provider support matrix and
// the labels decision for the Agentic Factory templates.
//
// The matrix is shipped to user projects, so it cites public provider
// documentation only. The citations to the design contract text (the canon
// each GitHub cell rests on) live HERE, in afMatrixCanon, so the shipped file
// carries no internal design trail.
//
// Sentinels on failure: AF_PROVIDER_MATRIX / AF_LABELS_DECISION.
package template_test

import (
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const (
	afLabelsYML     = ".github/labels.yml"
	afLabelSyncYML  = ".github/workflows/label-sync.yml"
	afMatrixHeading = "| Concept | GitHub | GitLab | Used by these templates |"
)

// afMatrixStatuses is the closed status vocabulary of a matrix cell.
var afMatrixStatuses = []string{"supported differently", "not supported", "supported", "unverified"}

// afMatrixWant pins the status of both provider cells per concept. GitHub
// cells rest on the design contract (afMatrixCanon); GitLab cells are
// "unverified" unless provider documentation was read in this task, because
// the contract states that no GitHub feature may be assumed for GitLab.
var afMatrixWant = []struct {
	Concept, GitHub, GitLab string
}{
	{"Milestone", "supported", "unverified"},
	{"Epic", "not supported", "supported differently"},
	{"Sub-issue", "supported", "unverified"},
	{"Issue dependencies", "supported", "unverified"},
	{"Issue types", "supported differently", "unverified"},
	{"Plan-tier gating", "supported differently", "supported differently"},
	{"Labels", "supported", "unverified"},
	{"Projects", "supported differently", "unverified"},
	{"Issue forms", "supported", "not supported"},
	{"Issue templates (Markdown)", "supported", "supported"},
	{"Pull request / merge request templates", "supported", "supported"},
	{"Comment templates as repository files", "not supported", "not supported"},
	{"Light/dark image selection", "supported", "unverified"},
	{"Mermaid source rendering", "supported", "unverified"},
}

// afMatrixCanon records, per concept, the design-contract text the GitHub cell
// rests on. It is the citation the shipped matrix deliberately omits.
var afMatrixCanon = map[string]string{
	"Milestone":                  "DELTA3 s1: Milestone = repository-scope release/acceptance goal; not an organisation-wide milestone.",
	"Epic":                       "DELTA3 s1: Epic = parent Issue; assuming a separate Epic object on GitHub is forbidden; organisation Issue Type Epic only after confirmation and approval, else an approved type:epic label fallback.",
	"Sub-issue":                  "DELTA3 s1/s4: native sub-issues, triage and above, at most 100 per parent, at most 8 levels (about 3 recommended); cross-repository links need both sides' permission and visibility.",
	"Issue dependencies":         "DELTA3 s4: dependencies available on Free/Pro/Team/Enterprise Cloud; a relation change is not itself a runtime gate.",
	"Issue types":                "DELTA3 s4: organisation scope, edited by the organisation owner, at most 25; defaults may have been changed; repository admin is not organisation owner.",
	"Plan-tier gating":           "DELTA3 s4: do not assume a feature exists on every plan or deployment; inventory actual capability/endpoint/API version in preflight.",
	"Labels":                     "DELTA3 s3/s4: repository scope; create/edit/delete need write, apply/remove need triage; existing labels are never deleted, renamed or recoloured.",
	"Projects":                   "DELTA3 s4: user/organisation scope, not repository-owned; separate permission; private-repository items need repository read; at most 50 fields.",
	"Light/dark image selection": "DELTA2 s6: picture element with prefers-color-scheme is the documented GitHub mechanism; other platforms must be confirmed separately.",
	"Mermaid source rendering":   "DELTA2 s6: a platform-rendered Mermaid fence is not the same engine as the pinned Pele output.",
}

type afMatrixRow struct {
	Concept, GitHub, GitLab, Used string
}

func afParseMatrix(t *testing.T, text string) map[string]afMatrixRow {
	t.Helper()
	if !strings.Contains(text, afMatrixHeading) {
		t.Fatalf("AF_PROVIDER_MATRIX: matrix heading %q not found", afMatrixHeading)
	}
	rows := map[string]afMatrixRow{}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") || strings.HasPrefix(line, "| Concept") || strings.HasPrefix(line, "|---") || strings.HasPrefix(line, "| ---") {
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		if len(cells) != 4 {
			continue
		}
		for i := range cells {
			cells[i] = strings.TrimSpace(cells[i])
		}
		rows[cells[0]] = afMatrixRow{cells[0], cells[1], cells[2], cells[3]}
	}
	return rows
}

func afCellStatus(cell string) string {
	for _, s := range afMatrixStatuses {
		if cell == s || strings.HasPrefix(cell, s+" —") || strings.HasPrefix(cell, s+" -") {
			return s
		}
	}
	return ""
}

// TestAgenticFactorySupportMatrix: tier, epic and sub-issue (and the rest)
// are stated per provider with the contract's status; no equivalence is
// faked, and unverified GitLab cells say so.
func TestAgenticFactorySupportMatrix(t *testing.T) {
	t.Parallel()
	text := afReadTemplate(t, afSupportMatrix)
	rows := afParseMatrix(t, text)

	for _, w := range afMatrixWant {
		row, ok := rows[w.Concept]
		if !ok {
			t.Errorf("AF_PROVIDER_MATRIX: concept %q missing from the matrix", w.Concept)
			continue
		}
		if got := afCellStatus(row.GitHub); got != w.GitHub {
			t.Errorf("AF_PROVIDER_MATRIX: %q GitHub status = %q (cell %q), want %q", w.Concept, got, row.GitHub, w.GitHub)
		}
		if got := afCellStatus(row.GitLab); got != w.GitLab {
			t.Errorf("AF_PROVIDER_MATRIX: %q GitLab status = %q (cell %q), want %q", w.Concept, got, row.GitLab, w.GitLab)
		}
		if row.Used != "yes" && row.Used != "no" {
			t.Errorf("AF_PROVIDER_MATRIX: %q 'Used by these templates' must be yes or no, got %q", w.Concept, row.Used)
		}
	}
	if len(rows) != len(afMatrixWant) {
		t.Errorf("AF_PROVIDER_MATRIX: matrix has %d rows, expected %d (update afMatrixWant deliberately)", len(rows), len(afMatrixWant))
	}
	for concept := range afMatrixCanon {
		if _, ok := rows[concept]; !ok {
			t.Errorf("AF_PROVIDER_MATRIX: canon citation for %q has no matrix row", concept)
		}
	}

	// the three concepts the brief singles out, both providers, no blanks
	for _, c := range []string{"Plan-tier gating", "Epic", "Sub-issue"} {
		r := rows[c]
		if r.GitHub == "" || r.GitLab == "" {
			t.Errorf("AF_PROVIDER_MATRIX: %q must have a cell for both providers", c)
		}
	}
	// GitHub has no native Epic object; the cell must say what stands in for it
	if c := rows["Epic"].GitHub; !strings.Contains(c, "parent Issue") {
		t.Errorf("AF_PROVIDER_MATRIX: the GitHub Epic cell must name the parent Issue as the stand-in, got %q", c)
	}
	for _, concept := range []string{"Milestone", "Sub-issue", "Labels"} {
		if !strings.Contains(rows[concept].GitLab, "not asserted") {
			t.Errorf("AF_PROVIDER_MATRIX: unverified GitLab cell for %q must say \"not asserted\", got %q", concept, rows[concept].GitLab)
		}
	}

	for _, must := range []string{
		"No equivalence between the providers is assumed",
		"GAP-PELE-SVG", "GAP-MERMAID-RENDER", "GAP-LOCALE-VARIANTS", "GAP-REMOTE-SETUP", "GAP-PROVIDER-GATED-DEPLOY",
		"https://docs.github.com/", "https://docs.gitlab.com/",
	} {
		if !strings.Contains(text, must) {
			t.Errorf("AF_PROVIDER_MATRIX: matrix document must contain %q", must)
		}
	}
	// the shipped file must not leak the internal design trail
	for _, banned := range []string{"DELTA", "S28", "S46", "S51", "S52", "#1948", "AF-DESIGN"} {
		if strings.Contains(text, banned) {
			t.Errorf("AF_PROVIDER_MATRIX: shipped matrix cites internal design identifier %q; keep citations in the test", banned)
		}
	}
}

// ---------------------------------------------------------------------------
// Labels decision
// ---------------------------------------------------------------------------

type afLabel struct {
	Name        string `yaml:"name"`
	Color       string `yaml:"color"`
	Description string `yaml:"description"`
}

func afParseLabels(t *testing.T, rel string) []afLabel {
	t.Helper()
	var labels []afLabel
	if err := yaml.Unmarshal([]byte(afReadTemplate(t, rel)), &labels); err != nil {
		t.Fatalf("AF_LABELS_DECISION: %s does not parse as a label list: %v", rel, err)
	}
	return labels
}

// afPinnedLabelNames are the labels the template labels.yml carried before
// this feature. label-sync must keep seeing every one of them.
var afPinnedLabelNames = []string{
	"stale", "pinned", "help wanted", "good first issue", "wontfix", "duplicate", "security", "dependencies",
	"type:feature", "type:fix", "type:docs", "type:chore", "type:ci", "type:refactor", "type:security", "type:test", "type:performance",
	"priority:P0", "priority:P1", "priority:P2", "priority:P3", "priority:P4",
	"status:in-progress", "status:review", "status:blocked", "status:needs-info",
}

// TestAgenticFactoryLabelsDecision: labels.yml is NOT extended; the proposed
// labels live in a separate, inert manifest that label-sync never reads.
func TestAgenticFactoryLabelsDecision(t *testing.T) {
	t.Parallel()
	existing := afParseLabels(t, afLabelsYML)
	have := map[string]bool{}
	for _, l := range existing {
		have[l.Name] = true
	}
	for _, name := range afPinnedLabelNames {
		if !have[name] {
			t.Errorf("AF_LABELS_DECISION: labels.yml lost label %q; existing labels may never be removed", name)
		}
	}
	if len(existing) != len(afPinnedLabelNames) {
		t.Errorf("AF_LABELS_DECISION: labels.yml has %d labels, the pinned baseline has %d; the feature must not extend labels.yml (labels.yml is applied automatically by label-sync)", len(existing), len(afPinnedLabelNames))
	}

	sync := afReadTemplate(t, afLabelSyncYML)
	for _, must := range []string{"config-file: .github/labels.yml", `- ".github/labels.yml"`, `- ".github/workflows/label-sync.yml"`} {
		if !strings.Contains(sync, must) {
			t.Errorf("AF_LABELS_DECISION: label-sync.yml no longer contains %q; workflow behaviour must be unchanged", must)
		}
	}
	if strings.Contains(sync, "agentic-factory") || strings.Contains(sync, "labels.proposed") {
		t.Errorf("AF_LABELS_DECISION: label-sync.yml must not reference the proposed manifest")
	}

	proposedRaw := afReadTemplate(t, afProposedLabels)
	for _, must := range []string{"PROPOSED", "separate manifest", "label-sync", "explicit approval", "never deleted, renamed or recoloured"} {
		if !strings.Contains(proposedRaw, must) {
			t.Errorf("AF_LABELS_DECISION: %s header must record %q", afProposedLabels, must)
		}
	}
	proposed := afParseLabels(t, afProposedLabels)
	if len(proposed) == 0 {
		t.Fatalf("AF_LABELS_DECISION: %s lists no labels", afProposedLabels)
	}
	seen := map[string]bool{}
	colour := regexp.MustCompile(`^[0-9a-fA-F]{6}$`)
	for _, l := range proposed {
		if strings.TrimSpace(l.Name) == "" || strings.TrimSpace(l.Description) == "" {
			t.Errorf("AF_LABELS_DECISION: proposed label %+v needs a name and description", l)
		}
		if !colour.MatchString(l.Color) {
			t.Errorf("AF_LABELS_DECISION: proposed label %q colour %q is not 6 hex digits", l.Name, l.Color)
		}
		if seen[l.Name] {
			t.Errorf("AF_LABELS_DECISION: duplicate proposed label %q", l.Name)
		}
		seen[l.Name] = true
		if have[l.Name] {
			t.Errorf("AF_LABELS_DECISION: proposed label %q duplicates an existing label; reuse it instead", l.Name)
		}
		if strings.HasPrefix(l.Name, "status:") {
			t.Errorf("AF_LABELS_DECISION: no new status label may be proposed (runtime state is not a label), got %q", l.Name)
		}
	}
	for _, want := range []string{"type:epic", "area:agentic-factory", "agent:managed"} {
		if !seen[want] {
			t.Errorf("AF_LABELS_DECISION: expected proposed label %q", want)
		}
	}
	// the manifest must not sit at a path label-sync reads
	if afProposedLabels == afLabelsYML {
		t.Errorf("AF_LABELS_DECISION: proposed manifest must be a separate file")
	}
}
