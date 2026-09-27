package escalation

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/contract"
)

// stateSchemaVersion is the card state file's schema version.
const stateSchemaVersion = 1

// CardState is the card state file: a cache of the arming snapshot, the verify
// cache, and the operational counters. Its content never arms or disarms a
// card by itself; the card audit log decides that (design.md §C.6).
type CardState struct {
	SchemaVersion int    `json:"schema_version"`
	Card          string `json:"card"`
	// Armed is the snapshot of the current arming, nil while unarmed.
	Armed *Arming `json:"armed,omitempty"`
	// LastArming is the snapshot of the most recent ended arming; after a
	// disarm the operational classes count against its budget (REQ-AE-014).
	LastArming *Arming `json:"last_arming,omitempty"`
	// VerifyCache is the last verify result for the armed contract, keyed by
	// the contract file's byte digest.
	VerifyCache *VerifyCache `json:"verify_cache,omitempty"`
	Counters    Counters     `json:"counters"`
	// ConsumedEvidence lists the evidence files classes 5 and 9 have already
	// tripped on, as class:fingerprint:file:content-digest. Evidence files
	// outlive an arming, so this is kept across re-arming (unlike Counters).
	ConsumedEvidence []string `json:"consumed_evidence,omitempty"`
}

// Arming is the snapshot recorded when a card arms (REQ-AE-023).
type Arming struct {
	SpecID       string `json:"spec_id"`
	ContractPath string `json:"contract_path"`
	// ContractSHA256 is the contract's signature.contract_sha256.
	ContractSHA256 string `json:"contract_sha256"`
	// ContractDigest is the SHA-256 of the contract file bytes at arming.
	ContractDigest string          `json:"contract_digest"`
	Card           string          `json:"card"`
	FrozenFiles    []string        `json:"frozen_files"`
	EffectiveNever []string        `json:"effective_never"`
	Scratch        []string        `json:"scratch"`
	Write          []string        `json:"write"`
	Invariants     []string        `json:"invariants"`
	Actions        []string        `json:"actions"`
	SecondModel    string          `json:"second_model"`
	Budget         contract.Budget `json:"budget"`
	ArmedAt        string          `json:"armed_at"`
	// DisarmReasons are the disarm reasons observed for this arming once it
	// ended; a new reason increments its record, a repeated one does not.
	DisarmReasons []string `json:"disarm_reasons,omitempty"`
}

// VerifyCache is one cached verify result.
type VerifyCache struct {
	ContractDigest string   `json:"contract_digest"`
	State          string   `json:"state"`
	Reasons        []string `json:"reasons"`
}

// Counters are the class 7 counters (turns and audit retries arrive with
// their classes in a later milestone).
type Counters struct {
	Operations int `json:"operations"`
	Turns      int `json:"turns"`
	// Streaks tracks class 8: per normalized command, the current failure
	// fingerprint and how many consecutive failures carried it.
	Streaks map[string]Streak `json:"streaks,omitempty"`
	// ExecutedInvariants are the command-kind invariants a tool call executed
	// since the previous commit checkpoint (REQ-AE-022).
	ExecutedInvariants []string `json:"executed_invariants,omitempty"`
}

// Streak is one command's consecutive-failure run (class 8).
type Streak struct {
	Fingerprint string `json:"fingerprint"`
	Count       int    `json:"count"`
}

// readCardStateRaw reads the card state file and its bytes; ok is false when
// it is absent.
func readCardStateRaw(path string) (CardState, []byte, bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return CardState{}, nil, false, nil
	}
	if err != nil {
		return CardState{}, nil, false, fmt.Errorf("escalation: read card state: %w", err)
	}
	var st CardState
	if err := json.Unmarshal(data, &st); err != nil {
		// Undecodable bytes are still present: the digest check judges them.
		return CardState{}, data, true, nil
	}
	return st, data, true, nil
}

// ReadCardState reads the card state file; ok is false when it is absent.
func ReadCardState(path string) (CardState, bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return CardState{}, false, nil
	}
	if err != nil {
		return CardState{}, false, fmt.Errorf("escalation: read card state: %w", err)
	}
	var st CardState
	if err := json.Unmarshal(data, &st); err != nil {
		return CardState{}, false, fmt.Errorf("escalation: decode card state: %w", err)
	}
	return st, true, nil
}

// writeCardState replaces the card state file atomically and returns the
// bytes written, whose digest the caller records in a state log entry.
func writeCardState(path string, st CardState) ([]byte, error) {
	st.SchemaVersion = stateSchemaVersion
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("escalation: encode card state: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("escalation: create store: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".state-*.json")
	if err != nil {
		return nil, fmt.Errorf("escalation: write card state: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return nil, fmt.Errorf("escalation: write card state: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return nil, fmt.Errorf("escalation: write card state: %w", err)
	}
	if err := atomicfile.Replace(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return nil, fmt.Errorf("escalation: replace card state: %w", err)
	}
	return data, nil
}
