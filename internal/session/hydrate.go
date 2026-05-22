package session

import "fmt"

// HydrateForPrompt returns the three context fragments needed to assemble a
// prompt in a fixed order.
//
// cache-prefix: DO NOT REORDER
// The (systemPrompt, userContext, systemContext) order is a component of the
// Anthropic cache key. Reordering causes cache misses and increases inference
// cost. Never change this order without a new SPEC.
//
// @MX:ANCHOR: [AUTO] SPEC-V3R2-RT-004 cache-prefix discipline (P-C05 closure)
// @MX:REASON: (systemPrompt, userContext, systemContext) assembly order is frozen here.
// DO NOT REORDER without a new SPEC — reordering causes Anthropic cache misses.
func HydrateForPrompt(phase Phase, specID string, store SessionStore) (systemPrompt, userContext, systemContext string, err error) {
	// cache-prefix: DO NOT REORDER — changing the return order changes the cache key.
	state, err := store.Hydrate(phase, specID)
	if err != nil {
		return "", "", "", fmt.Errorf("hydrate phase state: %w", err)
	}
	if state == nil {
		// Return empty strings when there is no checkpoint (the orchestrator treats it as a new session)
		return "", "", "", nil
	}

	// 1. systemPrompt: checkpoint metadata (phase, specID, status summary)
	systemPrompt = fmt.Sprintf("phase=%s spec=%s status=%s", state.Phase, state.SPECID, checkpointStatus(state))

	// 2. userContext: user context based on SPECID
	userContext = fmt.Sprintf("spec_id=%s", state.SPECID)

	// 3. systemContext: provenance info
	systemContext = fmt.Sprintf("provenance_source=%s origin=%s", state.Provenance.Source, state.Provenance.Origin)

	return systemPrompt, userContext, systemContext, nil
}

// checkpointStatus extracts a simple status string from the checkpoint.
func checkpointStatus(state *PhaseState) string {
	if state.Checkpoint == nil {
		return "no_checkpoint"
	}
	switch cp := state.Checkpoint.(type) {
	case *PlanCheckpoint:
		return cp.Status
	case *RunCheckpoint:
		return cp.Status
	case *SyncCheckpoint:
		// SyncCheckpoint has no Status field — return status based on docs_synced
		if cp.DocsSynced {
			return "synced"
		}
		return "pending"
	default:
		return "unknown"
	}
}
