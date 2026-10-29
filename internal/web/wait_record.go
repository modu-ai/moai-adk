// wait_record.go — the agent's wait record on the PR tracking row
// (Issue #1948). It decodes the monitor Row's `wait` key (Field shape; value
// keys named as core.Wait's JSON), and only displays: the wait sits beside the
// position and never changes the state, stage, park or gate shown. A key the
// monitor did not send or confirm is a gap; a confirmed null (no wait record)
// is "none".
package web

import (
	"strconv"
	"time"
)

type monWait struct {
	Owner      string    `json:"owner"`
	NextAction string    `json:"next_action"`
	RecheckAt  time.Time `json:"recheck_at"`
	Diagnosis  string    `json:"diagnosis"`
	Retries    int       `json:"retries"`
	MaxRetries int       `json:"max_retries"`
	EscalateTo string    `json:"escalate_to"`
	Escalated  bool      `json:"escalated"`
}

// WaitVM is one row's wait record. State is "recorded", "none" or a gap.
type WaitVM struct {
	State      string
	Owner      string
	NextAction string
	RecheckAt  string
	Diagnosis  string
	Retries    string // "n of max", or a gap when no limit was reported
	EscalateTo string
	Escalated  string // yes | no
}

func buildWait(f monField[*monWait]) WaitVM {
	switch {
	case !f.Confirmed:
		return WaitVM{State: prGap}
	case f.Value == nil:
		return WaitVM{State: "none"}
	}
	w := f.Value
	vm := WaitVM{
		State: "recorded", Owner: orGap(w.Owner), NextAction: orGap(w.NextAction), RecheckAt: timeText(w.RecheckAt),
		Diagnosis: orGap(w.Diagnosis), Retries: prGap, EscalateTo: orGap(w.EscalateTo), Escalated: "no",
	}
	if w.MaxRetries > 0 {
		vm.Retries = strconv.Itoa(w.Retries) + " of " + strconv.Itoa(w.MaxRetries)
	}
	if w.Escalated {
		vm.Escalated = "yes"
	}
	return vm
}
