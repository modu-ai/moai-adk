package linear

// SPECIssueMapping represents the mapping between a SPEC and a Linear issue.
type SPECIssueMapping struct {
	SpecID   string `json:"spec_id"`
	IssueID  string `json:"issue_id"`
	IssueURL string `json:"issue_url"`
	Status   string `json:"status"`
}

// Sync manages SPEC-to-Linear issue synchronization.
type Sync struct {
	client *Client
}

// NewSync creates a new Linear sync manager.
func NewSync(client *Client) *Sync {
	return &Sync{client: client}
}

// SyncStatus updates a Linear issue status based on SPEC state.
func (s *Sync) SyncStatus(specID, status string) error {
	if !s.client.IsEnabled() {
		return nil
	}
	// Status sync implementation deferred until Linear issue tracking is established
	return nil
}
