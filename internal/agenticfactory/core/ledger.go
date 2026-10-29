package core

// Binding ties a piece of evidence to the role, session, attempt and epoch
// that produced it (design §4, §11). Identity resolution itself belongs to
// SPEC-AGENTIC-FACTORY-IDENTITY-001; core only carries the observed value.
type Binding struct {
	Role    string `json:"role"`
	Session string `json:"session"`
	Attempt string `json:"attempt"`
	Epoch   int64  `json:"epoch"`
}

// Evidence is one review, audit, QA, or handoff record bound to a HEAD.
type Evidence struct {
	Path    string   `json:"path"`
	Head    HeadSHA  `json:"head"`
	Verdict string   `json:"verdict,omitempty"`
	By      *Binding `json:"by,omitempty"`
}

// Ledger is the machine-readable effect-state record of one Issue (design §4).
// Human-owned fields (task name, acceptance criteria, priority, dependencies)
// live in the Issue body and are intentionally not mirrored here.
type Ledger struct {
	Issue IssueRef `json:"issue"`
	State State    `json:"state"`
	// Stage, Sub, Park, Gate and Refusal are the S46 position; State,
	// ReturnStage and BlockReason stay for readers that have not moved yet.
	Stage         Stage       `json:"stage,omitempty"`
	Sub           SubStage    `json:"sub_stage,omitempty"`
	Park          *Park       `json:"park,omitempty"`
	Gate          *Gate       `json:"gate,omitempty"`
	Refusal       *Refusal    `json:"refusal,omitempty"`
	Wait          *Wait       `json:"wait,omitempty"`
	ReturnStage   State       `json:"return_stage,omitempty"`
	BlockReason   BlockReason `json:"block_reason,omitempty"`
	ResumeCond    string      `json:"resume_condition,omitempty"`
	Owner         string      `json:"owner,omitempty"`
	Writer        string      `json:"writer,omitempty"`
	Run           RunID       `json:"run_id,omitempty"`
	Attempt       AttemptID   `json:"attempt_id,omitempty"`
	Head          HeadSHA     `json:"head_sha,omitempty"`
	CheckLinks    []string    `json:"check_links,omitempty"`
	Review        []Evidence  `json:"review_evidence,omitempty"`
	Audit         []Evidence  `json:"audit_evidence,omitempty"`
	QA            []Evidence  `json:"qa_evidence,omitempty"`
	HandoffWriter *Evidence   `json:"handoff_writer,omitempty"`
	HandoffCoord  *Evidence   `json:"handoff_coordinator,omitempty"`
	NextAction    string      `json:"next_action,omitempty"`
	StoreSeq      int64       `json:"store_seq"`
	Delivery      Delivery    `json:"delivery,omitempty"`
	LastCheckedAt string      `json:"last_checked_at,omitempty"`
}
