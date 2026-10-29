package monitor

import (
	"context"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// PRInfo is what the remote says about the PR tied to an Issue.
type PRInfo struct {
	Number           int
	Open             bool
	Merged           bool
	IssueClosed      bool
	HeadSHA          string
	Checks           []CheckRun
	ReviewVerdict    string // PASS, FAIL, or empty when no verdict exists for this HEAD
	RepairDispatched bool
}

// PRReader reads PR state. It is read-only.
type PRReader interface {
	ReadPR(ctx context.Context, issue int) (PRInfo, bool, error)
	CountOpenPRs(ctx context.Context) (int, error)
}

// PRPhase says where a PR is on its way to completion.
type PRPhase string

const (
	PhaseNone             PRPhase = "none"
	PhaseChecksPending    PRPhase = "checks-pending"
	PhaseChecksFailed     PRPhase = "checks-failed"
	PhaseReviewPending    PRPhase = "review-pending"
	PhaseRepairNeeded     PRPhase = "repair-needed"
	PhaseRepairDispatched PRPhase = "repair-dispatched"
	PhaseReadyToMerge     PRPhase = "ready-to-merge"
	PhaseMergedIssueOpen  PRPhase = "merged-issue-open"
	PhaseIssueClosed      PRPhase = "issue-closed"
)

// PRStatus is the displayed PR summary.
type PRStatus struct {
	Number int     `json:"number"`
	Phase  PRPhase `json:"phase"`
}

// Phase derives the phase from remote facts. Checks that ran against another HEAD count as not run.
func Phase(p PRInfo) PRPhase {
	if p.Merged {
		if p.IssueClosed {
			return PhaseIssueClosed
		}
		return PhaseMergedIssueOpen
	}
	if !p.Open {
		return PhaseNone
	}
	fresh, failed := 0, false
	for _, k := range p.Checks {
		if k.HeadSHA != p.HeadSHA {
			continue
		}
		fresh++
		if k.Conclusion != "success" {
			failed = true
		}
	}
	switch {
	case fresh == 0:
		return PhaseChecksPending
	case failed:
		return PhaseChecksFailed
	}
	switch p.ReviewVerdict {
	case "PASS":
		return PhaseReadyToMerge
	case "FAIL":
		if p.RepairDispatched {
			return PhaseRepairDispatched
		}
		return PhaseRepairNeeded
	}
	return PhaseReviewPending
}

// BacklogState compares the open-PR count with the configured cap.
type BacklogState string

const (
	BacklogWithin  BacklogState = "within-limit"
	BacklogAtLimit BacklogState = "at-limit"
	BacklogOver    BacklogState = "over-limit"
	BacklogNoLimit BacklogState = "no-limit-set"
)

// Backlog is the unmerged-PR pile-up display.
type Backlog struct {
	Open  int          `json:"open"`
	Limit int          `json:"limit"`
	State BacklogState `json:"state"`
	// Queue figures; zero and QueueKnown=false when the queue reader is absent or failed.
	AwaitingHandoff   int  `json:"awaiting_handoff"`
	RepairBacklog     int  `json:"repair_backlog"`
	OldestWaitSeconds int  `json:"oldest_wait_seconds"`
	QueueKnown        bool `json:"queue_known"`
}

func backlogState(open, limit int) BacklogState {
	switch {
	case limit <= 0:
		return BacklogNoLimit
	case open > limit:
		return BacklogOver
	case open == limit:
		return BacklogAtLimit
	}
	return BacklogWithin
}

// View is the whole monitor output: one row per Issue plus the PR backlog.
type View struct {
	Rows    []Row               `json:"rows"`
	Backlog Field[Backlog]      `json:"backlog"`
	Intake  Field[[]IntakeItem] `json:"intake"`
}

// View collects rows and the backlog.
func (c *Collector) View(ctx context.Context, issues []int) View {
	v := View{Rows: c.Collect(ctx, issues)}
	v.Backlog.Source = SrcRemote
	if c.PRs != nil {
		if n, err := c.PRs.CountOpenPRs(ctx); err == nil {
			b := Backlog{Open: n, Limit: c.BacklogLimit, State: backlogState(n, c.BacklogLimit)}
			c.fillQueue(ctx, &b)
			v.Backlog = Field[Backlog]{b, SrcRemote, c.Now(), true, false}
		}
	}
	c.fillIntake(ctx, &v)
	return v
}

func (c *Collector) fillPR(ctx context.Context, r *Row, at time.Time) {
	r.PR.Source = SrcRemote
	if c.PRs == nil {
		return
	}
	info, found, err := c.PRs.ReadPR(ctx, r.Issue)
	if err != nil {
		return
	}
	r.PR = Field[PRStatus]{PRStatus{Number: info.Number, Phase: PhaseNone}, SrcRemote, at, true, false}
	if found {
		r.PR.Value.Phase = Phase(info)
	}
	r.Completed = found && info.Merged && info.IssueClosed && r.Stage.Value == string(core.StageDone)
}
