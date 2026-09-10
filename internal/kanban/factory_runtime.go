package kanban

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/specid"
	"github.com/modu-ai/moai-adk/internal/homestate"
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
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	return db.RecordRun(context.Background(), homestate.FactoryRun{RunID: runID, Backend: backend, ManifestJSON: string(raw)})
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
	return RecordFactoryCardState(root, runID, cardID, owner, specID, "picked", "card.assigned")
}

func RecordFactoryCardState(root, runID, cardID, owner, specID, state, eventKind string) error {
	provenance := captureFactoryProvenance(root, specID)
	payload, err := json.Marshal(map[string]string{
		"card_id": cardID, "owner": owner, "state": state,
		"spec_id": provenance.SpecID, "spec_path": provenance.SpecPath,
		"spec_sha256": provenance.SpecSHA256, "git_commit": provenance.GitCommit,
		"captured_at": provenance.CapturedAt,
	})
	if err != nil {
		return err
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	return db.RecordCard(context.Background(), homestate.FactoryCard{RunID: runID, CardID: cardID, OwnerLabel: owner,
		State: state, EventKind: eventKind, PayloadJSON: string(payload)})
}
