package review

import (
	"fmt"
	"regexp"
)

// State is one of the four delivery states of a review command (REQ-AF-003).
type State string

const (
	StateSent      State = "sent"
	StateReceived  State = "received"
	StateExecuted  State = "executed"
	StateCompleted State = "completed"
)

// Source is the evidence that may set a State.
type Source string

const (
	SourceSendResult  Source = "send-result"
	SourceIdleNotice  Source = "idle-notice"
	SourceAck         Source = "ack"
	SourceExecLog     Source = "exec-log"
	SourceVerdictFile Source = "verdict-file"
)

// sourceFor names the only evidence source each state accepts. A send result
// or an idle notice therefore can never set received, executed, or completed.
var sourceFor = map[State]Source{
	StateSent:      SourceSendResult,
	StateReceived:  SourceAck,
	StateExecuted:  SourceExecLog,
	StateCompleted: SourceVerdictFile,
}

// Producer is the role binding the control service observed at registration.
type Producer struct {
	Role    string
	Session string
	Attempt int
	Epoch   int
}

// Record binds one provider verdict to the full HEAD SHA it judged.
type Record struct {
	HeadSHA  string
	Kind     Kind
	Provider string
	Verdict  Verdict
	Producer Producer

	delivery map[State]Source
}

// Mark sets a delivery state, refusing evidence that does not belong to it.
func (r *Record) Mark(st State, src Source) error {
	want, ok := sourceFor[st]
	if !ok || src != want {
		return fmt.Errorf("review: state %q cannot be set from %q", st, src)
	}
	if r.delivery == nil {
		r.delivery = map[State]Source{}
	}
	r.delivery[st] = src
	return nil
}

// Reached reports whether a state was set from its own evidence.
func (r Record) Reached(st State) bool {
	_, ok := r.delivery[st]
	return ok
}

// Reason says why a record is not admitted; ReasonNone means admitted.
type Reason string

const (
	ReasonNone       Reason = ""
	ReasonBadSHA     Reason = "bad-sha"
	ReasonStale      Reason = "stale"
	ReasonIncomplete Reason = "incomplete"
	ReasonProducer   Reason = "producer-not-allowed"
	ReasonDebt       Reason = "debt-present"
	ReasonNotPass    Reason = "not-pass"
)

var fullSHA = regexp.MustCompile(`^[0-9a-f]{40}$`)

// Admit applies the merge-gate rule (REQ-AF-018, REQ-AF-020): only a
// completed, same-HEAD, debt-free PASS from a session other than the writer
// counts. Skipped, missing, failed, stale, and PASS-WITH-DEBT do not.
func (r Record) Admit(candidateHead, writerSession string) Reason {
	switch {
	case !fullSHA.MatchString(candidateHead) || !fullSHA.MatchString(r.HeadSHA):
		return ReasonBadSHA
	case r.HeadSHA != candidateHead:
		return ReasonStale
	case !r.Reached(StateCompleted):
		return ReasonIncomplete
	case r.Producer.Session == "" || r.Producer.Session == writerSession:
		return ReasonProducer
	case r.Verdict == VerdictPassWithDebt:
		return ReasonDebt
	case r.Verdict != VerdictPass:
		return ReasonNotPass
	}
	return ReasonNone
}

// String is a credential-free one-line form for logs.
func (r Record) String() string {
	return fmt.Sprintf("review{head=%s kind=%s provider=%s verdict=%s role=%s session=%s}",
		r.HeadSHA, r.Kind, r.Provider, r.Verdict, r.Producer.Role, r.Producer.Session)
}
