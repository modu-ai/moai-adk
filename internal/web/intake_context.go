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

// monIntakeItem is the monitor's S51 intake item (lane-20 View.intake).
type monIntakeItem struct {
	Key    string `json:"key"`
	Issue  int    `json:"issue"`
	Stage  string `json:"stage"`
	State  string `json:"state"`
	Parked string `json:"parked"`
	Block  *struct {
		Step      string `json:"step"`
		Detail    string `json:"detail"`
		Retryable bool   `json:"retryable"`
		Unknown   bool   `json:"unknown"`
		Resume    string `json:"resume"`
	} `json:"block"`
	Started bool   `json:"started"`
	Agent   string `json:"agent"`
	Slot    string `json:"slot"`
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

// buildIntake maps the monitor's intake items to the request rows. The monitor
// derives State (EFFECT_UNKNOWN, BLOCKED, PARKED, else the stage); an empty
// State is derived here from the same fields rather than shown blank.
func buildIntake(f monField[[]monIntakeItem]) IntakeRequestsVM {
	if !f.Confirmed {
		return IntakeRequestsVM{}
	}
	vm := IntakeRequestsVM{Known: true, Source: "monitor"}
	for _, it := range f.Value {
		row := IntakeRequestVM{
			Key: it.Key, Status: orGap(it.State), Stage: orGap(it.Stage), Parked: it.Parked, Issue: prGap,
			AgentName: prGap, Agent: orGap(it.Agent), Slot: orGap(it.Slot), Ready: prGap, AckedBy: prGap,
			Started: it.Started, Retryable: prGap, Branch: prGap,
		}
		if it.Issue > 0 {
			row.Issue = "#" + strconv.Itoa(it.Issue)
		}
		if b := it.Block; b != nil {
			row.BlockStep, row.BlockDetail, row.Resume, row.Retryable = b.Step, b.Detail, b.Resume, "no"
			if b.Retryable {
				row.Retryable = "yes"
			}
			vm.BlockedCount++
			if row.Status == prGap {
				row.Status = "BLOCKED"
				if b.Unknown {
					row.Status = "EFFECT_UNKNOWN"
				}
			}
		}
		if it.Parked != "" {
			vm.ParkedCount++
			if row.Status == prGap {
				row.Status = "PARKED"
			}
		}
		vm.Rows = append(vm.Rows, row)
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
