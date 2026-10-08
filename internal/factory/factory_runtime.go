package factory

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/specid"
)

// factoryProvenance is captured at card assignment/state-change time. A
// factory run can start before a SPEC is selected, so the card event is the
// SSOT for the SPEC revision actually assigned to that card.
type factoryProvenance struct {
	SpecID     string `json:"spec_id"`
	SpecPath   string `json:"spec_path"`
	SpecSHA256 string `json:"spec_sha256"`
	GitCommit  string `json:"git_commit"`
	CapturedAt string `json:"captured_at"`
}

func RecordFactoryRunStart(root, runID, backend, specID string) error {
	_ = specID // compatibility: run creation does not claim a card-level SPEC snapshot.
	manifest := captureFactoryProvenance(root, "")
	raw, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	return NewBacklogStore(BacklogPathForRoot(root)).recordRuntime(TodoRuntimeRun{RunID: runID, Backend: backend, ManifestJSON: string(raw)}, nil)
}

func captureFactoryProvenance(root, specID string) factoryProvenance {
	manifest := factoryProvenance{SpecID: specID, CapturedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	if specID != "" && specid.ValidateSpecID(specID) == nil {
		candidate, err := filepath.Abs(filepath.Join(root, ".moai", "specs", specID, "spec.md"))
		if err == nil {
			manifest.SpecPath = filepath.Clean(candidate)
			if resolved, resolveErr := filepath.EvalSymlinks(candidate); resolveErr == nil {
				manifest.SpecPath = resolved
			}
			if raw, readErr := os.ReadFile(candidate); readErr == nil {
				sum := sha256.Sum256(raw)
				manifest.SpecSHA256 = hex.EncodeToString(sum[:])
			}
		}
	}
	if out, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output(); err == nil {
		manifest.GitCommit = strings.TrimSpace(string(out))
	}
	return manifest
}

func RecordFactoryCardAssignment(root, runID, cardID, owner, specID string) error {
	provenance := captureFactoryProvenance(root, specID)
	payload, err := json.Marshal(provenance)
	if err != nil {
		return err
	}
	return NewBacklogStore(BacklogPathForRoot(root)).recordRuntimeHook(TodoRuntimeRun{RunID: runID, ManifestJSON: "{}"}, &TodoRuntimeAssignment{
		RunID: runID, CardID: cardID, OwnerLabel: owner, ReportedState: "picked", EventKind: "card.assigned", ProvenanceJSON: string(payload),
	}, func() error {
		// The dispatch binding follows every successful assignment, in the
		// same queue-lock critical section (review round-20 P1,
		// SPEC-FACTORY-COMPLETION-RECOVERY-001): a completion path cannot
		// interleave between the assignment save and the binding re-point,
		// and no caller can forget the re-point. REQ-FCR-002's scope
		// sentence holds inside — a card with no factory row in ANY run is
		// an ordinary card and the write skips silently.
		// A same-run reassignment moves the row's owner in the same
		// factory transaction as the binding (review round-24 P1-2/P1-4,
		// V1): a refusal commits neither.
		return RecordDispatchEngagementIfEngaged(root, cardID, runID, owner)
	})
}

func RecordFactoryCardState(root, runID, cardID, owner, specID, state, eventKind string) error {
	provenance := captureFactoryProvenance(root, specID)
	payload, err := json.Marshal(provenance)
	if err != nil {
		return err
	}
	return NewBacklogStore(BacklogPathForRoot(root)).recordRuntime(TodoRuntimeRun{RunID: runID, ManifestJSON: "{}"}, &TodoRuntimeAssignment{
		RunID: runID, CardID: cardID, OwnerLabel: owner, ReportedState: state, EventKind: eventKind, ProvenanceJSON: string(payload),
	})
}
