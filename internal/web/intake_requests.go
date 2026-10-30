// intake_requests.go — the background intake request panel (Issue #1948, S51).
// It decodes lane-18's request records (claim.Store.ViewRequests, JSON keys per
// claim/request.go) and only reads. BLOCKED and EFFECT_UNKNOWN are derived from
// the record's blocked block, never from the stage; "started" is the record's
// derived flag, not a stage; a stage this code does not know is shown verbatim.
package web

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
)

type monRequest struct {
	Key        string `json:"key"`
	Priority   int    `json:"priority"`
	Seq        int64  `json:"seq"`
	Stage      string `json:"stage"`
	Parked     string `json:"parked"`
	Issue      int    `json:"issue"`
	Agent      string `json:"agent"`
	AgentName  string `json:"agent_name"`
	Slot       string `json:"slot"`
	AckedBy    string `json:"acked_by"`
	Attempts   int    `json:"attempts"`
	Started    bool   `json:"started"`
	StartedRef string `json:"started_ref"`
	Branch     *struct {
		Ref       string `json:"ref"`
		Remote    bool   `json:"remote"`
		RemoteOID string `json:"remote_oid"`
	} `json:"branch"`
	Ready *struct {
		Host          string `json:"host"`
		Slot          string `json:"slot"`
		Path          string `json:"path"`
		Head          string `json:"head"`
		WriterStopped bool   `json:"writer_stopped"`
	} `json:"ready"`
	Blocked *struct {
		Step      string `json:"step"`
		Detail    string `json:"detail"`
		Retryable bool   `json:"retryable"`
		Unknown   bool   `json:"unknown"`
		Resume    string `json:"resume"`
	} `json:"blocked"`
	Evidence []struct {
		Stage string `json:"stage"`
		Ref   string `json:"ref"`
	} `json:"evidence"`
}

// IntakeRequestVM is one background intake request.
type IntakeRequestVM struct {
	// Redacted is set when refining masked something on this row.
	Redacted    bool
	Key         string
	Status      string // stage, or BLOCKED / EFFECT_UNKNOWN when the record is blocked
	Stage       string // the stored stage
	Parked      string // backpressure | no-slot, empty when running
	Issue       string
	AgentName   string // display name; not an identity
	Agent       string // reserved agent's session id
	Slot        string
	Ready       string
	AckedBy     string
	Attempts    int
	Started     bool // derived from the claim, not stored
	StartedRef  string
	Retryable   string // yes | no | —
	BlockStep   string
	BlockDetail string
	Resume      string
	Branch      string
	Evidence    []string
}

// IntakeRequestsVM is the panel; Known is false when the record is unavailable.
type IntakeRequestsVM struct {
	Known        bool
	Source       string // claim-store | monitor
	Note         string
	Rows         []IntakeRequestVM
	ParkedCount  int
	BlockedCount int
}

// IntakeRequestsFrom builds an IntakeRequests source from the claim store's
// read-only ViewRequests. A nil store yields a source that fails.
func IntakeRequestsFrom(s *claim.Store) func(context.Context) ([]byte, error) {
	return func(context.Context) ([]byte, error) {
		if s == nil {
			return nil, errors.New("intake request source is not configured")
		}
		reqs, err := s.ViewRequests()
		if err != nil {
			return nil, err
		}
		return json.Marshal(reqs)
	}
}

// buildIntakeRequests prefers the claim store's request records, which carry
// attempts, evidence and the ready receipt. Without that source it falls back
// to the monitor's intake items; the panel names which one it shows.
func (a *app) buildIntakeRequests(ctx context.Context) IntakeRequestsVM {
	if a.intakeRequests == nil {
		if m := a.buildPRTracking(ctx).Intake; m.Known {
			return m
		}
		return IntakeRequestsVM{Note: "intake requests are not connected"}
	}
	raw, err := a.intakeRequests(ctx)
	if err != nil {
		return IntakeRequestsVM{Note: "intake requests unreadable: " + refineText(err.Error())}
	}
	var reqs []monRequest
	if err := json.Unmarshal(raw, &reqs); err != nil {
		return IntakeRequestsVM{Note: "intake requests malformed: " + refineText(err.Error())}
	}
	vm := IntakeRequestsVM{Known: true, Source: "claim-store"}
	for _, r := range reqs {
		row := IntakeRequestVM{
			Redacted: anyMasked(r.rawTexts()...),
			Key:      refineText(r.Key), Status: orGap(r.Stage), Stage: orGap(r.Stage), Parked: refineText(r.Parked), Issue: prGap,
			AgentName: orGap(r.AgentName), Agent: orGap(r.Agent), Slot: orGap(r.Slot), Ready: prGap,
			AckedBy: orGap(r.AckedBy), Attempts: r.Attempts, Started: r.Started, StartedRef: refineText(r.StartedRef),
			Retryable: prGap, Branch: prGap,
		}
		if r.Issue > 0 {
			row.Issue = "#" + strconv.Itoa(r.Issue)
		}
		if r.Branch != nil {
			row.Branch = refineText(r.Branch.Ref)
		}
		if r.Ready != nil {
			row.Ready = refineText(r.Ready.Host + "/" + r.Ready.Slot + " " + r.Ready.Path + " @" + r.Ready.Head)
		}
		if b := r.Blocked; b != nil {
			row.Status, row.BlockStep, row.BlockDetail, row.Resume = "BLOCKED", refineText(b.Step), refineText(b.Detail), refineText(b.Resume)
			if b.Unknown {
				row.Status = "EFFECT_UNKNOWN"
			}
			row.Retryable = "no"
			if b.Retryable {
				row.Retryable = "yes"
			}
			vm.BlockedCount++
		}
		if r.Parked != "" {
			vm.ParkedCount++
			if r.Blocked == nil {
				row.Status = "PARKED"
			}
		}
		for _, e := range r.Evidence {
			row.Evidence = append(row.Evidence, refineText(e.Stage+": "+e.Ref))
		}
		vm.Rows = append(vm.Rows, row)
	}
	return vm
}

func requestsMeta(v IntakeRequestsVM) string {
	if !v.Known {
		return "no data"
	}
	return v.Source + " · " + strconv.Itoa(len(v.Rows)) + " requests · parked " + strconv.Itoa(v.ParkedCount) + " · blocked " + strconv.Itoa(v.BlockedCount)
}

// rawTexts lists every free text of a request record that reaches the page,
// before refining, so the row can say whether anything was masked.
func (r monRequest) rawTexts() []string {
	out := []string{r.Key, r.Stage, r.Parked, r.Agent, r.AgentName, r.Slot, r.AckedBy, r.StartedRef}
	if b := r.Branch; b != nil {
		out = append(out, b.Ref)
	}
	if rd := r.Ready; rd != nil {
		out = append(out, rd.Host, rd.Slot, rd.Path, rd.Head)
	}
	if b := r.Blocked; b != nil {
		out = append(out, b.Step, b.Detail, b.Resume)
	}
	for _, e := range r.Evidence {
		out = append(out, e.Stage, e.Ref)
	}
	return out
}
