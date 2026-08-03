// Package planhtml renders the plan-phase HTML report (SPEC-GOAL-HTML-FLOW-001
// REQ-GHF-007/008): it parses a plan-auditor review markdown file into a struct,
// derives the 8-field autonomy contract deterministically from SPEC artifacts,
// and renders both into a single self-contained HTML document. The output is
// openable offline (inline CSS, zero external JS/CSS framework dependency).
//
// The renderer is a pure function over its inputs — it performs read-only file
// I/O on the specDir + reviewFile and returns bytes; no subprocess, no network.
// When the review file is absent or unparseable, the renderer fails OPEN: it
// emits the report with an "audit verdict unavailable" placeholder rather than
// failing the plan-phase pipeline (REQ-GHF-007 fail-open).
package planhtml

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// planHTMLModel is the render model. Every field is a plain string (NEVER
// template.HTML — AP-1). html/template's context-aware auto-escape is the sole
// escape path for every untrusted field.
type planHTMLModel struct {
	SpecID   string
	Title    string
	Goal     string
	Verdict  string
	Score    string
	HasAudit bool
	MustPass []string
	Defects  []string
	// 8-field autonomy contract — each non-empty (or "undetermined — <reason>").
	Contract []contractField
	// Milestone list from plan.md §F.
	Milestones []string
}

type contractField struct {
	Label string
	Value string
}

// RenderPlanHTML renders the plan-phase HTML report. It reads spec.md, plan.md,
// acceptance.md, and settings.json from specDir, and the plan-auditor review
// markdown from reviewFile. It derives the 8-field autonomy contract
// deterministically (REQ-GHF-008) and renders a self-contained HTML document.
// When reviewFile is absent or unparseable, the audit slot carries an "audit
// verdict unavailable" placeholder (fail-open, REQ-GHF-007).
func RenderPlanHTML(specDir, reviewFile string) ([]byte, error) {
	specMD, _ := readText(filepath.Join(specDir, "spec.md"))
	planMD, _ := readText(filepath.Join(specDir, "plan.md"))
	acceptanceMD, _ := readText(filepath.Join(specDir, "acceptance.md"))
	settingsBytes, settingsErr := os.ReadFile(filepath.Join(specDir, "settings.json"))

	m := planHTMLModel{
		SpecID: extractFrontmatter(specMD, "id"),
		Title:  extractFrontmatter(specMD, "title"),
		Goal:   deriveGoal(specMD),
	}
	m.Contract = derive8Fields(specMD, planMD, acceptanceMD, settingsBytes, settingsErr)
	m.Milestones = deriveMilestones(planMD)

	// Parse the review markdown (fail-open on absence / parse miss).
	audit := parseReview(reviewFile)
	if audit != nil {
		m.HasAudit = true
		m.Verdict = audit.verdict
		m.Score = audit.score
		m.MustPass = audit.mustPass
		m.Defects = audit.defects
	}

	tmpl, err := template.New("planhtml").Parse(planHTMLTemplate)
	if err != nil {
		return nil, fmt.Errorf("planhtml parse: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, m); err != nil {
		return nil, fmt.Errorf("planhtml execute: %w", err)
	}
	return buf.Bytes(), nil
}

func readText(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// extractFrontmatter pulls a single YAML frontmatter field value from markdown.
// Best-effort; returns "" on miss.
func extractFrontmatter(md, field string) string {
	lines := strings.Split(md, "\n")
	inFM := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			if !inFM {
				inFM = true
				continue
			}
			break // end of frontmatter
		}
		if inFM {
			if strings.HasPrefix(trimmed, field+":") {
				val := strings.TrimSpace(strings.TrimPrefix(trimmed, field+":"))
				val = strings.Trim(val, "\"'")
				return val
			}
		}
	}
	return ""
}

// deriveGoal extracts the "so that" clause from SPEC §A (REQ-GHF-008 field 1).
var goalClauseRe = regexp.MustCompile(`(?i)so that\s+(.+?)(?:\.|\n|$)`)

func deriveGoal(specMD string) string {
	if m := goalClauseRe.FindStringSubmatch(specMD); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return "undetermined — no 'so that' goal clause found in §A"
}

// derive8Fields produces the 8-field contract deterministically (REQ-GHF-008).
// Each field is non-empty; an underivable field renders as
// "undetermined — <reason>".
func derive8Fields(specMD, planMD, acceptanceMD string, settingsBytes []byte, settingsErr error) []contractField {
	return []contractField{
		{Label: "goal", Value: deriveGoal(specMD)},
		{Label: "scope", Value: deriveScope(specMD)},
		{Label: "non-goals", Value: deriveNonGoals(specMD)},
		{Label: "tools-permissions", Value: derivePermissions(settingsBytes, settingsErr)},
		{Label: "stopping-condition", Value: deriveStoppingCondition(acceptanceMD)},
		{Label: "evidence", Value: deriveEvidence(acceptanceMD, planMD)},
		{Label: "escalation", Value: deriveEscalation(planMD)},
		{Label: "budget", Value: deriveBudget(specMD, planMD)},
	}
}

var reqBulletRe = regexp.MustCompile(`(?m)^-\s+(REQ-[A-Z0-9-]+)\s+[—-]\s+(.+)$`)

func deriveScope(specMD string) string {
	var rows []string
	for _, m := range reqBulletRe.FindAllStringSubmatch(specMD, -1) {
		rows = append(rows, fmt.Sprintf("%s — %s", m[1], strings.TrimSpace(m[2])))
	}
	if len(rows) == 0 {
		return "undetermined — no REQ bullets found in §B scope"
	}
	return strings.Join(rows, "; ")
}

var outOfScopeRe = regexp.MustCompile(`(?m)^###\s+Out of Scope\s*[—-]\s*(.+)$`)

func deriveNonGoals(specMD string) string {
	var topics []string
	for _, m := range outOfScopeRe.FindAllStringSubmatch(specMD, -1) {
		topics = append(topics, strings.TrimSpace(m[1]))
	}
	if len(topics) == 0 {
		return "undetermined — no '### Out of Scope —' H3 headings in §B"
	}
	return strings.Join(topics, "; ")
}

func derivePermissions(settingsBytes []byte, settingsErr error) string {
	if settingsErr != nil || len(settingsBytes) == 0 {
		return "undetermined — settings.json absent or unreadable"
	}
	var doc struct {
		Permissions struct {
			Allow []string `json:"allow"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal(settingsBytes, &doc); err != nil {
		return "undetermined — settings.json unparseable"
	}
	if len(doc.Permissions.Allow) == 0 {
		return "undetermined — settings.json has no permissions.allow"
	}
	return strings.Join(doc.Permissions.Allow, ", ")
}

var acRowRe = regexp.MustCompile(`(?m)^\|\s*(AC-[A-Z0-9-]+)\s*\|`)

func deriveStoppingCondition(acceptanceMD string) string {
	var ids []string
	seen := map[string]bool{}
	for _, m := range acRowRe.FindAllStringSubmatch(acceptanceMD, -1) {
		id := m[1]
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return "undetermined — no AC rows found in acceptance.md §D"
	}
	return strings.Join(ids, ", ") + " (MUST severity)"
}

func deriveEvidence(acceptanceMD, planMD string) string {
	var parts []string
	// "Then" clauses from acceptance.md GWT.
	thenCount := strings.Count(acceptanceMD, "**Then**")
	if thenCount > 0 {
		parts = append(parts, fmt.Sprintf("%d Given-When-Then 'Then' clauses", thenCount))
	}
	// plan.md §E self-verification items (bullets under §E).
	evCount := countPlanSelfVerification(planMD)
	if evCount > 0 {
		parts = append(parts, fmt.Sprintf("%d plan.md §E self-verification items", evCount))
	}
	if len(parts) == 0 {
		return "undetermined — no evidence rows found"
	}
	return strings.Join(parts, "; ")
}

func countPlanSelfVerification(planMD string) int {
	// Count "- " bullets that appear after a "## §E" or "Self-Verification" heading.
	idx := strings.Index(strings.ToLower(planMD), "self-verification")
	if idx < 0 {
		return 0
	}
	section := planMD[idx:]
	// Take until the next "## " heading.
	if next := strings.Index(section[1:], "\n## "); next >= 0 {
		section = section[:next]
	}
	count := 0
	for _, line := range strings.Split(section, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "- ") {
			count++
		}
	}
	return count
}

func deriveEscalation(planMD string) string {
	// plan.md open-questions / blocker cross-references.
	oqCount := strings.Count(planMD, "OQ-")
	if oqCount > 0 {
		return fmt.Sprintf("%d open-question cross-references in plan.md", oqCount)
	}
	return "undetermined — no open-questions / blocker cross-references in plan.md"
}

func deriveBudget(specMD, planMD string) string {
	tier := extractFrontmatter(specMD, "tier")
	phase := extractFrontmatter(specMD, "phase")
	milestoneCount := len(deriveMilestones(planMD))
	return fmt.Sprintf("tier=%s; phase=%s; milestone_count=%d", tier, phase, milestoneCount)
}

var milestoneRe = regexp.MustCompile(`(?m)^###\s+Milestone\s+(M\d+)`)

func deriveMilestones(planMD string) []string {
	var out []string
	seen := map[string]bool{}
	for _, m := range milestoneRe.FindAllStringSubmatch(planMD, -1) {
		id := m[1]
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

// reviewSummary is the parsed plan-auditor review markdown (verdict / score /
// must-pass / defects). Parsing is strict about the section headings but
// fail-opens (returns nil) on any miss.
type reviewSummary struct {
	verdict  string
	score    string
	mustPass []string
	defects  []string
}

var verdictRe = regexp.MustCompile(`(?m)^Verdict:\s*(PASS|FAIL|INCONCLUSIVE|BYPASSED)\b`)
var scoreRe = regexp.MustCompile(`(?m)^Overall Score:\s*([0-9.]+)`)
var mustPassRe = regexp.MustCompile(`(?m)^-\s*\[(PASS|FAIL|N/A)\]\s+(MP-\d+\s+.+)$`)
var defectRe = regexp.MustCompile(`(?m)^(D\d+\.\s+.+)$`)

func parseReview(reviewFile string) *reviewSummary {
	data, err := os.ReadFile(reviewFile)
	if err != nil {
		return nil // fail-open
	}
	md := string(data)
	vm := verdictRe.FindStringSubmatch(md)
	sm := scoreRe.FindStringSubmatch(md)
	if len(vm) < 2 || len(sm) < 2 {
		return nil // fail-open: missing verdict or score
	}
	r := &reviewSummary{verdict: vm[1], score: sm[1]}
	for _, m := range mustPassRe.FindAllStringSubmatch(md, -1) {
		r.mustPass = append(r.mustPass, strings.TrimSpace(m[2]))
	}
	for _, m := range defectRe.FindAllStringSubmatch(md, -1) {
		r.defects = append(r.defects, strings.TrimSpace(m[1]))
	}
	return r
}

const planHTMLTemplate = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Plan Report — {{.SpecID}}</title>
<style>
  :root {
    --bg: #FAF9F5; --paper: #FFFFFF; --ink: #141413; --clay: #D97757;
    --muted: #87867F; --line: #D1CFC5; --ok: #788C5D; --warn: #B85C3E;
  }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { background: var(--bg); color: var(--ink); font-family: system-ui, -apple-system, "Segoe UI", sans-serif; font-size: 15px; line-height: 1.6; padding: 40px 20px 80px; }
  .wrap { max-width: 960px; margin: 0 auto; }
  header { background: var(--paper); border: 1.5px solid var(--line); border-radius: 12px; padding: 24px 28px; margin-bottom: 20px; }
  header h1 { font-size: 20px; font-weight: 700; }
  header .sub { color: var(--muted); font-size: 13px; margin-top: 4px; }
  section { background: var(--paper); border: 1.5px solid var(--line); border-radius: 12px; padding: 20px 24px; margin-bottom: 20px; }
  section h2 { font-size: 15px; text-transform: uppercase; letter-spacing: 0.05em; color: var(--muted); margin-bottom: 12px; }
  .goal { border-left: 4px solid var(--clay); padding: 12px 16px; background: #FBF6F2; border-radius: 8px; font-size: 16px; }
  .audit-score { display: inline-block; padding: 4px 12px; border-radius: 4px; font-weight: 700; font-size: 14px; }
  .audit-score.PASS { background: #E8EEDC; color: var(--ok); }
  .audit-score.FAIL { background: #FBE9E3; color: var(--warn); }
  table { width: 100%; border-collapse: collapse; font-size: 14px; }
  th, td { text-align: left; padding: 8px 10px; border-bottom: 1px solid var(--line); vertical-align: top; }
  th { font-weight: 600; color: var(--muted); font-size: 12px; text-transform: uppercase; }
  .placeholder { color: var(--muted); font-style: italic; }
  ul.checklist { list-style: none; }
  ul.checklist li { padding: 4px 0; }
  .milestones { display: flex; flex-wrap: wrap; gap: 8px; }
  .milestone { padding: 4px 10px; border: 1px solid var(--line); border-radius: 4px; font-size: 13px; font-family: ui-monospace, "SF Mono", Menlo, Consolas, monospace; }
  footer { color: var(--muted); font-size: 12px; text-align: center; margin-top: 24px; }
</style>
</head>
<body>
<div class="wrap">
  <header>
    <h1>{{.Title}}</h1>
    <div class="sub">{{.SpecID}} · Plan Report</div>
  </header>

  <section>
    <h2>Goal</h2>
    <div class="goal">{{.Goal}}</div>
  </section>

  <section>
    <h2>Plan-Auditor Verdict</h2>
    {{if .HasAudit}}
    <p><span class="audit-score {{.Verdict}}">{{.Verdict}}</span> overall score <strong>{{.Score}}</strong></p>
    {{if .MustPass}}
    <h2 style="margin-top:16px;">Must-Pass Results</h2>
    <ul class="checklist">
      {{range .MustPass}}<li>{{.}}</li>{{end}}
    </ul>
    {{end}}
    {{if .Defects}}
    <h2 style="margin-top:16px;">Defects</h2>
    <ul class="checklist">
      {{range .Defects}}<li>{{.}}</li>{{end}}
    </ul>
    {{end}}
    {{else}}
    <p class="placeholder">audit verdict unavailable</p>
    {{end}}
  </section>

  <section>
    <h2>Autonomy Contract (8 fields)</h2>
    <table>
      <tr><th>Field</th><th>Derived Value</th></tr>
      {{range .Contract}}
      <tr>
        <td><strong>{{.Label}}</strong></td>
        <td>{{.Value}}</td>
      </tr>
      {{end}}
    </table>
  </section>

  <section>
    <h2>Milestones</h2>
    {{if .Milestones}}
    <div class="milestones">
      {{range .Milestones}}<span class="milestone">{{.}}</span>{{end}}
    </div>
    {{else}}
    <p class="placeholder">no milestones found in plan.md §F</p>
    {{end}}
  </section>

  <footer>Rendered by MoAI plan report · open this file in a browser to review offline</footer>
</div>
</body>
</html>`
