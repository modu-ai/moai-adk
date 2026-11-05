package core

import (
	"fmt"
	"strings"
	"time"
)

// Wait is the record an agent keeps while its work waits on something outside
// its hands (CI, a review, a host answer). It says who is responsible, what
// happens next on its own, when it is looked at again, why it waits, how many
// retries are spent of how many are allowed, and who is told when they run out.
// MaxRetries comes from configuration: there is no default and a missing value
// is refused. A Wait sits beside Gate and Park and never replaces them.
type Wait struct {
	Owner      string    `json:"owner"`
	NextAction string    `json:"next_action"`
	RecheckAt  time.Time `json:"recheck_at"`
	Diagnosis  string    `json:"diagnosis"`
	Retries    int       `json:"retries"`
	MaxRetries int       `json:"max_retries"`
	Escalate   string    `json:"escalate_to"`
	// Escalated is set when the retries ran out: the escalation target must now
	// be told, and the agent does not retry again by itself.
	Escalated bool `json:"escalated,omitempty"`
}

// Validate refuses a wait that leaves any of the six answers open.
func (w Wait) Validate() error {
	switch {
	case strings.TrimSpace(w.Owner) == "":
		return fmt.Errorf("%w: a wait needs its responsible owner", ErrEvidenceIncomplete)
	case strings.TrimSpace(w.NextAction) == "":
		return fmt.Errorf("%w: a wait needs the next automatic action", ErrEvidenceIncomplete)
	case w.RecheckAt.IsZero():
		return fmt.Errorf("%w: a wait needs a recheck time", ErrEvidenceIncomplete)
	case strings.TrimSpace(w.Diagnosis) == "":
		return fmt.Errorf("%w: a wait needs a diagnosis", ErrEvidenceIncomplete)
	case w.MaxRetries < 1:
		return fmt.Errorf("%w: a wait needs a configured retry limit of at least 1", ErrEvidenceIncomplete)
	case w.Retries < 0 || w.Retries > w.MaxRetries:
		return fmt.Errorf("%w: retries %d outside 0..%d", ErrEvidenceIncomplete, w.Retries, w.MaxRetries)
	case strings.TrimSpace(w.Escalate) == "":
		return fmt.Errorf("%w: a wait needs an escalation target", ErrEvidenceIncomplete)
	}
	return nil
}

// Retry spends one retry with a fresh diagnosis, next action and recheck time.
// The retry that reaches the limit returns the updated, escalated wait together
// with ErrRetryLimit; nothing is retried past the limit.
func (w Wait) Retry(diagnosis, next string, recheck time.Time) (Wait, error) {
	if w.Escalated || w.Retries >= w.MaxRetries {
		return w, fmt.Errorf("%w: %d of %d retries are spent; escalate to %s", ErrRetryLimit, w.Retries, w.MaxRetries, w.Escalate)
	}
	w.Retries++
	w.Diagnosis, w.NextAction, w.RecheckAt = diagnosis, next, recheck
	if err := w.Validate(); err != nil {
		return Wait{}, err
	}
	if w.Retries >= w.MaxRetries {
		w.Escalated = true
		return w, fmt.Errorf("%w: %d of %d retries are spent; escalate to %s", ErrRetryLimit, w.Retries, w.MaxRetries, w.Escalate)
	}
	return w, nil
}
