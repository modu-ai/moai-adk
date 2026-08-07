package goal

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
)

// dashboardModel is the render model passed to the html/template. Every field
// is a plain string (or a slice of structs whose inner fields are plain
// strings) — NEVER template.HTML (AP-1 / K-2). html/template's context-aware
// auto-escape is the sole escape path for every untrusted field.
//
// @MX:ANCHOR: [AUTO] goal dashboard render model — XSS auto-escape binding surface
// @MX:REASON: this model is the DOM-visible claim surface for the goal Verdict
// (verification-claim-integrity §1.1). Every untrusted field flows through
// html/template's {{.Field}} action in HTML context; no field is typed
// template.HTML (the escape-hatch footgun, AP-1). AC-GHF-002 binds this.
type dashboardModel struct {
	SessionID  string
	Goal       string
	Status     string
	TurnsUsed  int
	MaxTurns   int
	Mode       string
	CreatedAt  string
	Conditions []conditionModel
	// Verdict section — nil when no verdict has been produced (AC-GHF-011).
	HasVerdict        bool
	Turn              int
	Ceiling           int
	FailedConditions  []FailedCond
	CeilingExit       bool
	WallClockExit     bool
	Stagnation        bool
	Claim             string
	Evidence          string
	BaselineAttrib    string
	Gaps              string
	ResidualRisk      string
	SnapshotAttrib    []string
	ReArmIndicator    string // non-empty when pending.json embedded_goal present (M5)
	ReArmedView       string // non-empty when post-/clear new-session goal exists (M5)
	UnboundedBanner   string // non-empty when IsUnbounded() (M5)
}

type conditionModel struct {
	Index     int
	Type      string
	Cmd       string
	ExpectExit int
	Claim     string
}


// ReArmContext carries the already-landed mechanical re-arm pipeline's state
// for the dashboard's render-only re-arm UI (SPEC-GOAL-HTML-FLOW-001
// REQ-GHF-010 / AC-GHF-007). It mirrors the handoff.EmbeddedGoal fields WITHOUT
// importing internal/hook/handoff (the dashboard renderer stays a pure
// renderer; the CLI/orchestrator layer translates). A nil EmbeddedCondition
// with EmbeddedMaxTurns==0 AND no NewSessionID means "no re-arm state" — the
// dashboard renders the base view.
//
// The three UI states (AC-GHF-007):
//   - EmbeddedCondition non-empty + !EmbeddedUnbounded → "re-arm on /clear" indicator
//   - NewSessionID non-empty → "re-armed under <id>" view (post-/clear)
//   - EmbeddedUnbounded true → D8-rejection banner
type ReArmContext struct {
	EmbeddedCondition   string
	EmbeddedMaxTurns    int
	EmbeddedMaxDuration int
	EmbeddedCostCap     int
	EmbeddedUnbounded   bool
	NewSessionID        string // post-/clear session whose goal file exists
}

// RenderDashboard produces a self-contained HTML document rendering the goal
// metadata and (when non-nil) the Verdict's failed conditions, turn/ceiling,
// and the 5-section CeilingVerdict (REQ-GHF-001..003). It is a pure function —
// no I/O, no subprocess. When v is nil, it renders goal metadata + a "no
// verdict yet" placeholder and omits the verdict section entirely (AC-GHF-011).
func RenderDashboard(g *Goal, v *Verdict) ([]byte, error) {
	return RenderDashboardReArm(g, v, nil)
}

// RenderDashboardReArm extends RenderDashboard with the render-only re-arm UI
// (REQ-GHF-010 / AC-GHF-007). A nil reArm produces output byte-identical to
// RenderDashboard (the re-arm path is purely additive).
func RenderDashboardReArm(g *Goal, v *Verdict, reArm *ReArmContext) ([]byte, error) {
	if g == nil {
		return nil, fmt.Errorf("RenderDashboard: nil goal")
	}
	m := buildDashboardModel(g, v)
	applyReArm(&m, reArm)
	tmpl, err := template.New("dashboard").Parse(dashboardTemplate)
	if err != nil {
		return nil, fmt.Errorf("RenderDashboard parse: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, m); err != nil {
		return nil, fmt.Errorf("RenderDashboard execute: %w", err)
	}
	return buf.Bytes(), nil
}

// applyReArm populates the model's ReArmIndicator / ReArmedView / UnboundedBanner
// fields from the already-landed mechanical re-arm state. The D8 banner takes
// precedence (an unbounded embedded goal is rejected at rearm; surfacing the
// rejection is more important than the indicator).
func applyReArm(m *dashboardModel, reArm *ReArmContext) {
	if reArm == nil {
		return
	}
	if reArm.EmbeddedUnbounded {
		m.UnboundedBanner = fmt.Sprintf(
			"D8 rejection: the embedded goal %q is unbounded (max_turns=0 with no real bound) "+
				"and was rejected at /clear re-arm. Re-arm a bounded goal (max_turns>0, or "+
				"--max-duration / --cost-cap) to resume the loop.",
			reArm.EmbeddedCondition)
		return
	}
	if reArm.EmbeddedCondition != "" {
		m.ReArmIndicator = fmt.Sprintf(
			"This goal will re-arm on /clear — embedded condition: %q (ceiling: max_turns=%d, "+
				"max_duration=%d, cost_cap=%d).",
			reArm.EmbeddedCondition, reArm.EmbeddedMaxTurns,
			reArm.EmbeddedMaxDuration, reArm.EmbeddedCostCap)
	}
	if reArm.NewSessionID != "" {
		m.ReArmedView = fmt.Sprintf(
			"Re-armed under session %s — the new goal state lives at "+
				".moai/state/goal/%s.json.",
			reArm.NewSessionID, reArm.NewSessionID)
	}
}

func buildDashboardModel(g *Goal, v *Verdict) dashboardModel {
	m := dashboardModel{
		SessionID: g.SessionID,
		Goal:      g.Goal,
		Status:    string(g.Status),
		TurnsUsed: g.TurnsUsed,
		MaxTurns:  g.Ceiling.MaxTurns,
		Mode:      string(g.ProgressionMode),
		CreatedAt: g.CreatedAt,
	}
	for i, c := range g.Conditions {
		cm := conditionModel{
			Index:      i,
			Type:       string(c.Type),
			Cmd:        c.Cmd,
			ExpectExit: c.ExpectExit,
			Claim:      c.Claim,
		}
		m.Conditions = append(m.Conditions, cm)
	}
	if v != nil {
		m.HasVerdict = true
		m.Turn = v.Turn
		m.Ceiling = v.Ceiling
		m.CeilingExit = v.CeilingExit
		m.WallClockExit = v.WallClockExit
		m.Stagnation = v.Stagnation
		m.SnapshotAttrib = v.SnapshotAttribution
		m.FailedConditions = v.FailedConditions
		if v.Verdict != nil {
			m.Claim = v.Verdict.Claim
			m.Evidence = v.Verdict.Evidence
			m.BaselineAttrib = v.Verdict.BaselineAttribution
			m.Gaps = v.Verdict.Gaps
			m.ResidualRisk = v.Verdict.ResidualRisk
		}
	}
	return m
}

const dashboardTemplate = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>MoAI Goal Dashboard — {{.SessionID}}</title>
<style>
  :root {
    --bg: #FAF9F5; --paper: #FFFFFF; --ink: #141413; --clay: #D97757;
    --muted: #87867F; --line: #D1CFC5; --ok: #788C5D; --warn: #B85C3E;
  }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { background: var(--bg); color: var(--ink); font-family: system-ui, -apple-system, "Segoe UI", sans-serif; font-size: 15px; line-height: 1.6; padding: 40px 20px 80px; }
  .wrap { max-width: 920px; margin: 0 auto; }
  header { background: var(--paper); border: 1.5px solid var(--line); border-radius: 12px; padding: 24px 28px; margin-bottom: 20px; }
  header h1 { font-size: 20px; font-weight: 700; margin-bottom: 4px; }
  header .meta { color: var(--muted); font-size: 13px; font-family: ui-monospace, "SF Mono", Menlo, Consolas, monospace; }
  .goal-text { background: var(--paper); border: 1.5px solid var(--line); border-left: 4px solid var(--clay); border-radius: 8px; padding: 16px 20px; margin-bottom: 20px; }
  .goal-text .label { font-size: 12px; text-transform: uppercase; letter-spacing: 0.05em; color: var(--muted); margin-bottom: 6px; }
  .goal-text .text { font-size: 16px; }
  section { background: var(--paper); border: 1.5px solid var(--line); border-radius: 12px; padding: 20px 24px; margin-bottom: 20px; }
  section h2 { font-size: 15px; text-transform: uppercase; letter-spacing: 0.05em; color: var(--muted); margin-bottom: 12px; }
  table { width: 100%; border-collapse: collapse; font-size: 14px; }
  th, td { text-align: left; padding: 8px 10px; border-bottom: 1px solid var(--line); vertical-align: top; }
  th { font-weight: 600; color: var(--muted); font-size: 12px; text-transform: uppercase; }
  td.mono, .mono { font-family: ui-monospace, "SF Mono", Menlo, Consolas, monospace; font-size: 13px; }
  .badge { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 12px; font-weight: 600; }
  .badge.exit { background: #FBE9E3; color: var(--warn); }
  .badge.ok { background: #E8EEDC; color: var(--ok); }
  .kv { display: grid; grid-template-columns: 140px 1fr; gap: 4px 12px; font-size: 14px; }
  .kv dt { color: var(--muted); }
  .placeholder { color: var(--muted); font-style: italic; padding: 12px 0; }
  .banner { border-left: 4px solid var(--warn); background: #FBE9E3; padding: 12px 16px; border-radius: 8px; margin-bottom: 16px; }
  .verdict-section { margin-bottom: 14px; }
  .verdict-section .h { font-weight: 700; font-size: 14px; color: var(--clay); margin-bottom: 4px; }
  .verdict-section .body { font-size: 14px; }
  .rearm { border-left: 4px solid var(--ok); background: #E8EEDC; padding: 12px 16px; border-radius: 8px; margin-bottom: 16px; }
  footer { color: var(--muted); font-size: 12px; text-align: center; margin-top: 24px; }
</style>
</head>
<body>
<div class="wrap">
  <header>
    <h1>Goal Dashboard</h1>
    <div class="meta">session: {{.SessionID}} · status: {{.Status}} · turns: {{.TurnsUsed}}/{{.MaxTurns}} · mode: {{.Mode}}{{if .CreatedAt}} · created: {{.CreatedAt}}{{end}}</div>
  </header>

  <div class="goal-text">
    <div class="label">Goal Condition</div>
    <div class="text">{{.Goal}}</div>
  </div>

  {{if .UnboundedBanner}}
  <div class="banner">{{.UnboundedBanner}}</div>
  {{end}}

  {{if .ReArmIndicator}}
  <div class="rearm">{{.ReArmIndicator}}</div>
  {{end}}

  {{if .ReArmedView}}
  <div class="rearm">{{.ReArmedView}}</div>
  {{end}}

  <section>
    <h2>Declared Conditions</h2>
    {{if .Conditions}}
    <table>
      <tr><th>#</th><th>Type</th><th>Detail</th></tr>
      {{range .Conditions}}
      <tr>
        <td>{{.Index}}</td>
        <td>{{.Type}}</td>
        <td>{{if eq .Type "mechanical"}}<span class="mono">{{.Cmd}}</span> (expect exit {{.ExpectExit}}){{else}}{{.Claim}}{{end}}</td>
      </tr>
      {{end}}
    </table>
    {{else}}
    <div class="placeholder">no conditions declared</div>
    {{end}}
  </section>

  {{if .HasVerdict}}
  <section>
    <h2>Verdict — Turn {{.Turn}} / Ceiling {{.Ceiling}}</h2>
    {{if .CeilingExit}}<p class="placeholder">ceiling reached{{if .WallClockExit}} (wall-clock bound){{else if .Stagnation}} (stagnation){{end}}</p>{{end}}
    {{if .FailedConditions}}
    <table>
      <tr><th>Command</th><th>Exit</th><th>Output Tail</th></tr>
      {{range .FailedConditions}}
      <tr>
        <td class="mono">{{.Cmd}}</td>
        <td><span class="badge exit">{{.Exit}}</span></td>
        <td class="mono">{{.Tail}}</td>
      </tr>
      {{end}}
    </table>
    {{else}}
    <div class="placeholder">no failed mechanical conditions this turn</div>
    {{end}}

    {{if or .Claim .Evidence .BaselineAttrib .Gaps .ResidualRisk}}
    <div style="margin-top:16px;">
      <div class="verdict-section"><div class="h">Claim</div><div class="body">{{.Claim}}</div></div>
      <div class="verdict-section"><div class="h">Evidence</div><div class="body">{{.Evidence}}</div></div>
      <div class="verdict-section"><div class="h">Baseline-attribution</div><div class="body">{{.BaselineAttrib}}</div></div>
      <div class="verdict-section"><div class="h">Gaps</div><div class="body">{{.Gaps}}</div></div>
      <div class="verdict-section"><div class="h">Residual-risk</div><div class="body">{{.ResidualRisk}}</div></div>
    </div>
    {{end}}

    {{if .SnapshotAttrib}}
    <div style="margin-top:14px;">
      <div class="label" style="font-size:12px;text-transform:uppercase;color:var(--muted);margin-bottom:4px;">Snapshot Attribution</div>
      <ul class="mono" style="font-size:12px;color:var(--muted);">
      {{range .SnapshotAttrib}}<li>{{.}}</li>{{end}}
      </ul>
    </div>
    {{end}}
  </section>
  {{else}}
  <section>
    <h2>Verdict</h2>
    <div class="placeholder">no verdict yet</div>
  </section>
  {{end}}

  <footer>Rendered by MoAI goal dashboard · open this file in a browser to monitor the loop</footer>
</div>
</body>
</html>`

// htmlPathDerivable is a sentinel-free helper note: callers derive the .html
// sibling from StatePath via strings.TrimSuffix(path, ".json") + ".html". Kept
// here only as a doc anchor; the derivation itself lives in the CLI/state layer.
var _ = strings.TrimSuffix
