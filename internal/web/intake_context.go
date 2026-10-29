// intake_context.go — background intake stages and the per-Issue context
// ledger on the monitor (Issue #1948). Both are decoded from the monitor JSON
// (keys confirmed in lane-20's monitor package, commit 12c06332e8) and only
// read. An unconfirmed wrapper renders as "not provided", a stale item and an
// inaccessible item stay distinct, and a link opens only when the monitor
// itself verified it remotely.
package web

import (
	"strconv"
	"strings"
	"time"
)

type monIntakeItem struct {
	Issue  int       `json:"issue"`
	Stage  string    `json:"stage"`
	State  string    `json:"state"`
	Reason string    `json:"reason"`
	At     time.Time `json:"at"`
}

type monCtxItem struct {
	Section        string    `json:"section"`
	Link           string    `json:"link"`
	TargetHead     string    `json:"target_head"`
	Hash           string    `json:"hash"`
	Owner          string    `json:"owner"`
	State          string    `json:"state"`
	Stage          string    `json:"stage"`
	RemoteVerified bool      `json:"remote_verified"`
	Availability   string    `json:"availability"`
	At             time.Time `json:"at"`
}

// IntakeRowVM is one Issue's acceptance record.
type IntakeRowVM struct {
	Issue  int
	Stage  string // as reported; empty when the monitor gave none
	State  string
	Reason string // already masked by the monitor
	At     string
}

// IntakeVM is the intake panel; Known is false when the monitor did not
// provide the record at all.
type IntakeVM struct {
	Known bool
	Rows  []IntakeRowVM
}

// ContextItemVM is one context-ledger entry.
type ContextItemVM struct {
	Section    string
	Link       string
	LinkOpen   bool // true only for a remotely verified, https link
	LinkStatus string
	TargetHead string
	Hash       string
	Owner      string
	State      string // current | stale | inaccessible, as reported
	Stage      string // how far the evidence travelled, as reported
	At         string
}

// ContextVM is one Issue's context ledger.
type ContextVM struct {
	Known bool
	Items []ContextItemVM
}

func buildIntake(f monField[[]monIntakeItem]) IntakeVM {
	if !f.Confirmed {
		return IntakeVM{}
	}
	vm := IntakeVM{Known: true}
	for _, it := range f.Value {
		vm.Rows = append(vm.Rows, IntakeRowVM{
			Issue: it.Issue, Stage: it.Stage, State: it.State, Reason: it.Reason, At: timeText(it.At),
		})
	}
	return vm
}

func buildContext(f monField[[]monCtxItem]) ContextVM {
	if !f.Confirmed {
		return ContextVM{}
	}
	vm := ContextVM{Known: true}
	for _, it := range f.Value {
		item := ContextItemVM{
			Section: it.Section, Link: it.Link, TargetHead: orGap(it.TargetHead), Hash: orGap(it.Hash),
			Owner: orGap(it.Owner), State: orGap(it.State), Stage: orGap(it.Stage), At: timeText(it.At),
		}
		switch {
		case it.Availability == "available" && it.RemoteVerified && it.Stage == "remote-verified" && strings.HasPrefix(it.Link, "https://"):
			item.LinkOpen, item.LinkStatus = true, "available"
		case it.Availability == "inaccessible", it.Availability == "restricted":
			item.LinkStatus = it.Availability
		default:
			// unverified, an unknown value, or "available" without the remote
			// check and the remote-verified stage: never an open link.
			item.LinkStatus = "unverified"
		}
		vm.Items = append(vm.Items, item)
	}
	return vm
}

func orGap(s string) string {
	if s == "" {
		return prGap
	}
	return s
}

func timeText(t time.Time) string {
	if t.IsZero() {
		return prGap
	}
	return t.UTC().Format("2006-01-02 15:04Z")
}

// queueFigure renders one PRQueue figure, or "not provided" when the monitor
// did not know the queue.
func queueFigure(b PRBacklogVM, n int) string {
	if !b.QueueKnown {
		return "not provided"
	}
	return strconv.Itoa(n)
}

// oldestWait renders the longest handoff wait.
func oldestWait(b PRBacklogVM) string {
	if !b.QueueKnown {
		return "not provided"
	}
	return (time.Duration(b.OldestWaitSeconds) * time.Second).String()
}

func intakeMeta(v IntakeVM) string {
	if !v.Known {
		return "no data"
	}
	return strconv.Itoa(len(v.Rows)) + " issues"
}

func contextMeta(vm PRTrackingVM) string {
	if !vm.Available {
		return "no data"
	}
	n := 0
	for _, r := range vm.Rows {
		n += len(r.Context.Items)
	}
	return strconv.Itoa(n) + " items"
}
