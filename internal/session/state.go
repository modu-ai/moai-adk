package session

import (
	"encoding/json"
	"time"
)

// PhaseState holds the typed session state for a workflow phase.
type PhaseState struct {
	Phase      Phase          `json:"phase"`
	SPECID     string         `json:"spec_id"`
	Checkpoint Checkpoint     `json:"checkpoint,omitempty"`
	BlockerRpt *BlockerReport `json:"blocker_report,omitempty"`
	UpdatedAt  time.Time      `json:"updated_at"`
	Provenance ProvenanceTag  `json:"provenance"`
}

// ProvenanceTag identifies the source of a state mutation.
type ProvenanceTag struct {
	Source  string    `json:"source"`  // "user", "project", "local", "session", "hook"
	Origin  string    `json:"origin"`  // file path or "cli"
	Loaded  time.Time `json:"loaded"`
}

// MarshalJSON implements custom JSON marshaling for PhaseState to handle
// the Checkpoint interface polymorphism.
func (ps *PhaseState) MarshalJSON() ([]byte, error) {
	type Alias PhaseState
	aux := &struct {
		CheckpointData map[string]any `json:"checkpoint,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(ps),
	}

	if ps.Checkpoint != nil {
		data, err := json.Marshal(ps.Checkpoint)
		if err != nil {
			return nil, err
		}
		var checkpointMap map[string]any
		if err := json.Unmarshal(data, &checkpointMap); err != nil {
			return nil, err
		}
		aux.CheckpointData = checkpointMap
	}

	return json.Marshal(aux)
}

// UnmarshalJSON implements custom JSON unmarshaling for PhaseState to handle
// the Checkpoint interface polymorphism.
func (ps *PhaseState) UnmarshalJSON(data []byte) error {
	type Alias PhaseState
	aux := &struct {
		CheckpointData map[string]any `json:"checkpoint,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(ps),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Reconstruct the concrete checkpoint type based on phase
	if aux.CheckpointData != nil {
		checkpointData, err := json.Marshal(aux.CheckpointData)
		if err != nil {
			return err
		}

		switch aux.Phase {
		case PhasePlan:
			var pc PlanCheckpoint
			if err := json.Unmarshal(checkpointData, &pc); err != nil {
				return err
			}
			aux.Checkpoint = &pc
		case PhaseRun:
			var rc RunCheckpoint
			if err := json.Unmarshal(checkpointData, &rc); err != nil {
				return err
			}
			aux.Checkpoint = &rc
		case PhaseSync:
			var sc SyncCheckpoint
			if err := json.Unmarshal(checkpointData, &sc); err != nil {
				return err
			}
			aux.Checkpoint = &sc
		}
	}

	return nil
}
