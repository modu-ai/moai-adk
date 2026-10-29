// pr_tracking.go — PR tracking and merge backlog for the monitor
// (Issue #1948 DELTA4). It decodes the monitor's {rows, backlog} JSON
// (internal/agenticfactory/monitor MarshalView) through the monitorRows seam
// and only reads. A field the monitor did not confirm renders as a gap, never
// as a value; the PR phase is shown as the monitor reports it, grouped into
// the five stages the operator tracks (checks, review, repair, merge, closed).
package web

import (
	"context"
	"encoding/json"
)

const prGap = "—"

// PRTrackRowVM is one issue's PR-facing status.
type PRTrackRowVM struct {
	Issue       int
	Project     string
	State       string
	Stage       string
	Head        string // 8-char short sha
	Checks      string // "N passed, M failed, K pending" or "not read"
	ChecksStale bool
	PR          string // "#N", or a gap when no PR is confirmed
	Phase       string // the monitor's phase, verbatim
	Track       string // checks | review | repair | merge | closed | gap
	Completed   bool
	Context     ContextVM
	Agent       string // display name (agent-N); never an identity key
	SessionID   string // full immutable id, or a gap when unconfirmed
	// S46 position detail, shown as the monitor reports it. A key the monitor
	// did not send or confirm is a gap; a confirmed null (nothing recorded) is "none".
	SubStage string
	Park     string
	Gate     string
	Refusal  string
	// Ledger is the ctxmodel view of this Issue's context ledger.
	Ledger ContextModelVM
	// Wait is the agent's wait record, shown beside the position and never
	// changing it.
	Wait WaitVM
}

// PRBacklogVM is the open-PR count against its limit.
type PRBacklogVM struct {
	Known bool
	Open  int
	Limit int
	State string // within-limit | at-limit | over-limit | no-limit-set
	// QueueKnown mirrors the monitor's queue_known. When false the three figures
	// below are zero placeholders and render as "not provided", never as 0.
	QueueKnown        bool
	AwaitingHandoff   int
	RepairBacklog     int
	OldestWaitSeconds int
}

// PRTrackingVM is the PR panel and /api/pr-tracking payload.
type PRTrackingVM struct {
	Available bool
	Note      string
	Rows      []PRTrackRowVM
	Backlog   PRBacklogVM
	Intake    IntakeRequestsVM
}

type monField[T any] struct {
	Value     T    `json:"value"`
	Confirmed bool `json:"confirmed"`
}

type monRow struct {
	Project     string
	Issue       int
	State       monField[string]
	Stage       monField[string]
	Agent       monField[string] `json:"agent"`
	SessionID   monField[string] `json:"session_id"`
	HeadSHA     monField[string]
	Checks      monField[[]struct{ Name, Conclusion, HeadSHA string }]
	ChecksStale bool
	PR          monField[struct {
		Number int    `json:"number"`
		Phase  string `json:"phase"`
	}] `json:"pr"`
	Completed bool                   `json:"completed"`
	Context   monField[[]monCtxItem] `json:"context"`
	SubStage  monField[string]       `json:"sub_stage"`
	Park      monField[*struct {
		Reason   string `json:"reason"`
		Evidence string `json:"evidence"`
		Resume   string `json:"resume_condition"`
	}] `json:"park"`
	Gate monField[*struct {
		Reason string `json:"reason"`
		Detail string `json:"detail"`
		Resume string `json:"resume_condition"`
		Head   string `json:"head"`
	}] `json:"gate"`
	Refusal monField[*struct {
		Reason string `json:"reason"`
		Detail string `json:"detail"`
	}] `json:"refusal"`
	Wait monField[*monWait] `json:"wait"`
}

type monView struct {
	Rows    []monRow `json:"rows"`
	Backlog monField[struct {
		Open              int    `json:"open"`
		Limit             int    `json:"limit"`
		State             string `json:"state"`
		AwaitingHandoff   int    `json:"awaiting_handoff"`
		RepairBacklog     int    `json:"repair_backlog"`
		OldestWaitSeconds int    `json:"oldest_wait_seconds"`
		QueueKnown        bool   `json:"queue_known"`
	}] `json:"backlog"`
	Intake monField[[]monIntakeItem] `json:"intake"`
}

func (a *app) buildPRTracking(ctx context.Context) PRTrackingVM {
	if a.monitorRows == nil {
		return PRTrackingVM{Note: "monitor data is not connected"}
	}
	raw, err := a.monitorRows(ctx)
	if err != nil {
		return PRTrackingVM{Note: "monitor data unreadable: " + err.Error()}
	}
	var doc monView
	if err := json.Unmarshal(raw, &doc); err != nil {
		return PRTrackingVM{Note: "monitor data malformed: " + err.Error()}
	}
	vm := PRTrackingVM{Available: true}
	if doc.Backlog.Confirmed {
		b := doc.Backlog.Value
		vm.Backlog = PRBacklogVM{
			Known: true, Open: b.Open, Limit: b.Limit, State: b.State, QueueKnown: b.QueueKnown,
			AwaitingHandoff: b.AwaitingHandoff, RepairBacklog: b.RepairBacklog, OldestWaitSeconds: b.OldestWaitSeconds,
		}
	}
	vm.Intake = buildIntake(doc.Intake)
	for _, r := range doc.Rows {
		row := PRTrackRowVM{
			Issue: r.Issue, Project: r.Project,
			State: gapText(r.State), Stage: gapText(r.Stage), Head: prGap,
			Checks: "not read", ChecksStale: r.ChecksStale,
			PR: prGap, Phase: prGap, Track: "gap", Completed: r.Completed,
			Agent: gapText(r.Agent), SessionID: gapText(r.SessionID),
			SubStage: gapText(r.SubStage),
			Park: detailText(r.Park.Confirmed, r.Park.Value != nil, func() []string {
				return []string{r.Park.Value.Reason, r.Park.Value.Evidence, resumeText(r.Park.Value.Resume)}
			}),
			Gate: detailText(r.Gate.Confirmed, r.Gate.Value != nil, func() []string {
				return []string{r.Gate.Value.Reason, r.Gate.Value.Detail, resumeText(r.Gate.Value.Resume), headText(r.Gate.Value.Head)}
			}),
			Refusal: detailText(r.Refusal.Confirmed, r.Refusal.Value != nil, func() []string {
				return []string{r.Refusal.Value.Reason, r.Refusal.Value.Detail}
			}),
		}
		if r.HeadSHA.Confirmed && r.HeadSHA.Value != "" {
			row.Head = r.HeadSHA.Value[:min(8, len(r.HeadSHA.Value))]
		}
		if r.Checks.Confirmed {
			row.Checks = summarizeChecks(r)
		}
		if r.PR.Confirmed {
			row.PR = "#" + itoa(r.PR.Value.Number)
			row.Phase, row.Track = orGap(r.PR.Value.Phase), prTrack(r.PR.Value.Phase)
		}
		row.Context = buildContext(r.Context)
		row.Ledger = a.buildContextModel(ctx, r.Issue)
		row.Wait = buildWait(r.Wait)
		vm.Rows = append(vm.Rows, row)
	}
	return vm
}

func gapText(f monField[string]) string {
	if !f.Confirmed || f.Value == "" {
		return prGap
	}
	return f.Value
}

// detailText renders a confirmed detail record as its non-empty parts joined
// by " · ". An unconfirmed or absent key is a gap, never "none".
func detailText(confirmed, present bool, parts func() []string) string {
	if !confirmed {
		return prGap
	}
	if !present {
		return "none"
	}
	out := ""
	for _, p := range parts() {
		if p == "" {
			continue
		}
		if out != "" {
			out += " · "
		}
		out += p
	}
	if out == "" {
		return prGap
	}
	return out
}

func resumeText(s string) string {
	if s == "" {
		return ""
	}
	return "resume: " + s
}

func headText(s string) string {
	if s == "" {
		return ""
	}
	return "head " + s[:min(8, len(s))]
}

// prTrack groups the monitor's phases into the five tracked stages. A phase
// this code does not know stays "gap" instead of being guessed into a stage.
func prTrack(phase string) string {
	switch phase {
	case "checks-pending", "checks-failed":
		return "checks"
	case "review-pending":
		return "review"
	case "repair-needed", "repair-dispatched":
		return "repair"
	case "ready-to-merge", "merged-issue-open":
		return "merge"
	case "issue-closed":
		return "closed"
	}
	return "gap"
}

func summarizeChecks(r monRow) string {
	var pass, fail, pend int
	for _, c := range r.Checks.Value {
		switch c.Conclusion {
		case "success":
			pass++
		case "failure", "cancelled", "timed_out":
			fail++
		default:
			pend++
		}
	}
	out := ""
	add := func(n int, label string) {
		if n == 0 {
			return
		}
		if out != "" {
			out += ", "
		}
		out += itoa(n) + " " + label
	}
	add(pass, "passed")
	add(fail, "failed")
	add(pend, "pending")
	if out == "" {
		return "none required"
	}
	return out
}

// prMeta is the panel's one-line summary.
func prMeta(vm PRTrackingVM) string {
	switch {
	case !vm.Available:
		return "no data"
	case !vm.Backlog.Known:
		return "backlog unknown"
	case vm.Backlog.State == "no-limit-set":
		return "open PRs " + itoa(vm.Backlog.Open) + " · no limit set · repair " + queueFigure(vm.Backlog, vm.Backlog.RepairBacklog)
	}
	return "open PRs " + itoa(vm.Backlog.Open) + "/" + itoa(vm.Backlog.Limit) + " · " + vm.Backlog.State +
		" · repair " + queueFigure(vm.Backlog, vm.Backlog.RepairBacklog)
}
