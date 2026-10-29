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

func (a *app) buildIntakeRequests(ctx context.Context) IntakeRequestsVM {
	if a.intakeRequests == nil {
		return IntakeRequestsVM{Note: "intake requests are not connected"}
	}
	raw, err := a.intakeRequests(ctx)
	if err != nil {
		return IntakeRequestsVM{Note: "intake requests unreadable: " + err.Error()}
	}
	var reqs []monRequest
	if err := json.Unmarshal(raw, &reqs); err != nil {
		return IntakeRequestsVM{Note: "intake requests malformed: " + err.Error()}
	}
	vm := IntakeRequestsVM{Known: true}
	for _, r := range reqs {
		row := IntakeRequestVM{
			Key: r.Key, Status: orGap(r.Stage), Stage: orGap(r.Stage), Parked: r.Parked, Issue: prGap,
			AgentName: orGap(r.AgentName), Agent: orGap(r.Agent), Slot: orGap(r.Slot), Ready: prGap,
			AckedBy: orGap(r.AckedBy), Attempts: r.Attempts, Started: r.Started, StartedRef: r.StartedRef,
			Retryable: prGap, Branch: prGap,
		}
		if r.Issue > 0 {
			row.Issue = "#" + strconv.Itoa(r.Issue)
		}
		if r.Branch != nil {
			row.Branch = r.Branch.Ref
		}
		if r.Ready != nil {
			row.Ready = r.Ready.Host + "/" + r.Ready.Slot + " " + r.Ready.Path + " @" + r.Ready.Head
		}
		if b := r.Blocked; b != nil {
			row.Status, row.BlockStep, row.BlockDetail, row.Resume = "BLOCKED", b.Step, b.Detail, b.Resume
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
		}
		for _, e := range r.Evidence {
			row.Evidence = append(row.Evidence, e.Stage+": "+e.Ref)
		}
		vm.Rows = append(vm.Rows, row)
	}
	return vm
}

func requestsMeta(v IntakeRequestsVM) string {
	if !v.Known {
		return "no data"
	}
	return strconv.Itoa(len(v.Rows)) + " requests · parked " + strconv.Itoa(v.ParkedCount) + " · blocked " + strconv.Itoa(v.BlockedCount)
}
