package session

import "time"

// BlockerReport represents an interrupt signal from a subagent.
// @MX:NOTE: [AUTO] SPEC-V3R2-RT-004 interrupt() equivalent — subagents call RecordBlocker;
// orchestrator surfaces via AskUserQuestion. Subagents MUST NOT call AskUserQuestion directly.
// See agent-common-protocol.md §User Interaction Boundary.
type BlockerReport struct {
	Kind            string         `json:"kind"` // "missing_input", "ambiguous", "error", "quality_gate"
	Message         string         `json:"message"`
	Context         map[string]any `json:"context,omitempty"`
	RequestedAction string         `json:"requested_action"`
	Provenance      ProvenanceTag  `json:"provenance"`
	Resolved        bool           `json:"resolved"`
	Resolution      string         `json:"resolution,omitempty"`
	Timestamp       time.Time      `json:"timestamp"`
	// SPEC-V3R2-RT-004 REQ-012: blocker file naming 기준 필드
	Phase  Phase  `json:"phase,omitempty"`   // blocker가 발생한 phase
	SPECID string `json:"spec_id,omitempty"` // 대상 SPEC ID
}

// NewBlockerReport creates a new BlockerReport with the current timestamp.
func NewBlockerReport(kind, message, requestedAction string, provenance ProvenanceTag) *BlockerReport {
	return &BlockerReport{
		Kind:            kind,
		Message:         message,
		RequestedAction: requestedAction,
		Provenance:      provenance,
		Resolved:        false,
		Timestamp:       time.Now(),
	}
}

// Resolve marks the blocker as resolved with the given resolution.
func (br *BlockerReport) Resolve(resolution string) {
	br.Resolved = true
	br.Resolution = resolution
}
