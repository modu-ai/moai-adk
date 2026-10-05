package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/spf13/cobra"
)

// adoptEvidenceFile is one recorded evidence file named in the briefing.
type adoptEvidenceFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

// adoptResumptionRecord is the resuming lane's own record, appended to the
// card's report directory alongside the previous owner's files. The previous
// owner's progress.md and evidence are never rewritten — the record carries
// their SHA-256 so the append-only guarantee is checkable after the fact
// (REQ-FLA-005 / AC-FLA-005).
type adoptResumptionRecord struct {
	Lane          string              `json:"lane"`
	Card          string              `json:"card"`
	SpecID        string              `json:"spec_id,omitempty"`
	AdoptedAt     string              `json:"adopted_at"`
	RecordedPhase string              `json:"recorded_phase"`
	ProgressSHA   string              `json:"progress_sha256,omitempty"`
	Evidence      []adoptEvidenceFile `json:"evidence,omitempty"`
}

// newFactoryAdoptCommand is the resume surface of REQ-FLA-004 (design.md D6):
// when a lane adopts a picked card whose previous owner is determined stalled
// — the stall determination itself being t1241's interface, not this verb's —
// it reads the card's progress.md and recorded evidence FIRST and continues
// from the recorded phase. The verb is the mechanical form of that
// read-before-work obligation: its briefing output is the recorded read, and
// the resumption record it appends never touches the previous owner's bytes.
// There is no stall detection here and no restart path: a card with nothing
// recorded is refused as a fresh pickup, not adopted.
func newFactoryAdoptCommand() *cobra.Command {
	var card string
	cmd := &cobra.Command{
		Use:   "adopt",
		Short: "Resume a stalled owner's picked card from its recorded progress and evidence",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(card) == "" {
				return fmt.Errorf("handoff adopt: --card is required")
			}
			lane, err := factoryLaneLabelFromEnv("handoff adopt")
			if err != nil {
				return err
			}
			// The queue-root resolution is the one the lane verbs share
			// (factoryCardRoot); every path this verb reads hangs off it.
			root := factoryCardRoot()
			if err := requireQueuePicked(root, card); err != nil {
				return fmt.Errorf("handoff adopt: %w", err)
			}
			specID, err := adoptCardSpecID(root, card)
			if err != nil {
				return err
			}
			brief, err := buildAdoptBrief(root, card, specID)
			if err != nil {
				return err
			}
			if brief.progressSHA == "" && len(brief.evidence) == 0 {
				return fmt.Errorf("handoff adopt: card %s has nothing recorded to resume from (no progress.md, no evidence) — a fresh pickup, not a resumption", card)
			}
			phase := phasePlan
			if brief.progressSHA != "" {
				phase = recordedPhase(brief.progress)
			}
			record := adoptResumptionRecord{
				Lane:          lane,
				Card:          card,
				SpecID:        specID,
				AdoptedAt:     time.Now().UTC().Format(time.RFC3339),
				RecordedPhase: phase,
				ProgressSHA:   brief.progressSHA,
				Evidence:      brief.evidence,
			}
			resumePath, err := appendResumptionRecord(brief.evidenceDir, record)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "resuming card %s (lane %s)\n", card, lane)
			if brief.progressPath != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "progress: %s (sha256 %s)\n", brief.progressPath, brief.progressSHA)
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), strings.TrimRight(string(brief.progress), "\n"))
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "recorded phase: %s\nevidence:\n", phase)
			for _, ev := range brief.evidence {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s (sha256 %s)\n", ev.Path, ev.SHA256)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "resumption record appended: %s\ncontinue from the recorded phase — resume semantics, not a silent restart (REQ-FLA-004)\n", resumePath)
			return nil
		},
	}
	cmd.Flags().StringVar(&card, "card", "", "Card id to adopt (must be in picked state)")
	_ = cmd.MarkFlagRequired("card")
	return cmd
}

const (
	phasePlan = "plan"
	phaseRun  = "run"
	phaseSync = "sync"
)

// adoptBrief carries what one adoption read.
type adoptBrief struct {
	progressPath string
	progress     []byte
	progressSHA  string
	evidenceDir  string
	evidence     []adoptEvidenceFile
}

// adoptCardSpecID reads the card's spec id from the queue record; an absent
// or nil spec id yields "" (the card is then resumed from its evidence
// alone).
func adoptCardSpecID(root, cardID string) (string, error) {
	record, err := todoReadStoreAt(root).LoadPure()
	if err != nil {
		return "", fmt.Errorf("handoff adopt: read queue: %w", err)
	}
	for _, item := range record.Items {
		if item.ID == cardID {
			return adoptDerefSpecID(item), nil
		}
	}
	for _, entry := range record.Archived {
		if entry.Item.ID == cardID {
			return adoptDerefSpecID(entry.Item), nil
		}
	}
	return "", nil
}

func adoptDerefSpecID(item factory.BacklogItem) string {
	if item.SpecID == nil {
		return ""
	}
	return *item.SpecID
}

// buildAdoptBrief reads the card's recorded surfaces: the SPEC progress.md
// (when the card carries a spec id) and every evidence file under the card's
// report directory. Reads only — nothing here writes.
func buildAdoptBrief(root, cardID, specID string) (adoptBrief, error) {
	brief := adoptBrief{evidenceDir: filepath.Dir(autoEvidencePath(root, cardID))}
	if specID != "" {
		brief.progressPath = filepath.Join(root, ".moai", "specs", specID, "progress.md")
		data, err := os.ReadFile(brief.progressPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return brief, fmt.Errorf("handoff adopt: read %s: %w", brief.progressPath, err)
			}
		} else if len(strings.TrimSpace(string(data))) > 0 {
			brief.progress = data
			brief.progressSHA = adoptFileSHA(data)
		}
	}
	entries, err := os.ReadDir(brief.evidenceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return brief, nil
		}
		return brief, fmt.Errorf("handoff adopt: read %s: %w", brief.evidenceDir, err)
	}
	for _, e := range entries {
		if e.IsDir() || e.Name() == "resumption.jsonl" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(brief.evidenceDir, e.Name()))
		if err != nil {
			return brief, fmt.Errorf("handoff adopt: read evidence %s: %w", e.Name(), err)
		}
		brief.evidence = append(brief.evidence, adoptEvidenceFile{
			Path:   filepath.Join(brief.evidenceDir, e.Name()),
			SHA256: adoptFileSHA(data),
		})
	}
	return brief, nil
}

// recordedPhase derives the card's recorded phase from its progress.md
// markers: a sync close (§E.4 carrying a real sync_commit_sha) reads sync,
// the run-phase evidence markers (§E.2/§E.3) read run, anything earlier reads
// plan. The derived phase is what the resuming lane continues from.
func recordedPhase(progress []byte) string {
	if strings.Contains(string(progress), "§E.4") {
		for _, line := range strings.Split(string(progress), "\n") {
			if !strings.Contains(line, "sync_commit_sha:") {
				continue
			}
			val := strings.TrimSpace(strings.Trim(strings.TrimSpace(strings.SplitN(line, ":", 2)[1]), `"`))
			if val != "" && !strings.HasPrefix(val, "pending-backfill") {
				return phaseSync
			}
		}
	}
	if strings.Contains(string(progress), "§E.2") || strings.Contains(string(progress), "§E.3") {
		return phaseRun
	}
	return phasePlan
}

// appendResumptionRecord appends the resuming lane's entry as one JSON line
// alongside the previous owner's files. The append is the only write this
// verb performs (REQ-FLA-005).
func appendResumptionRecord(evidenceDir string, record adoptResumptionRecord) (string, error) {
	if err := os.MkdirAll(evidenceDir, 0o755); err != nil {
		return "", fmt.Errorf("handoff adopt: mkdir %s: %w", evidenceDir, err)
	}
	path := filepath.Join(evidenceDir, "resumption.jsonl")
	data, err := json.Marshal(record)
	if err != nil {
		return "", fmt.Errorf("handoff adopt: marshal resumption record: %w", err)
	}
	data = append(data, '\n')
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return "", fmt.Errorf("handoff adopt: open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write(data); err != nil {
		return "", fmt.Errorf("handoff adopt: append %s: %w", path, err)
	}
	return path, nil
}

// adoptFileSHA is the SHA-256 hex of one recorded file's bytes.
func adoptFileSHA(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
