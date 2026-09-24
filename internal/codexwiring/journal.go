package codexwiring

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// JournalRelPath is the wiring write journal. Every change to a wiring file is
// appended here before its temporary file exists, so an interrupted change is
// always recoverable from the journal alone.
const JournalRelPath = ".moai/state/codex-wiring-journal.json"

// tempPrefix names the temporary files a wiring write stages in the target
// directory. The same prefix was used before the journal existed, so a temp
// file of an older binary is recognised as an orphan.
const tempPrefix = ".codexwiring-"

// Journal operations.
const (
	OpWrite  = "write"
	OpDelete = "delete"
)

// Journal entry states.
const (
	// JournalStaged is an incomplete change: recovery classifies it.
	JournalStaged = "staged"
	// JournalComplete is a change whose read-back and provenance landed.
	JournalComplete = "complete"
	// JournalConflict is a change refused because the target moved under it.
	JournalConflict = "conflict"
	// JournalDiverged is a change whose target matched neither side.
	JournalDiverged = "diverged"
	// JournalDiscarded is a change recovery found not applied and dropped.
	JournalDiscarded = "discarded"
)

// JournalEntry is one wiring file change. PreHash and PostHash are the hex
// sha256 of the target before and after the change; an empty hash means the
// file is absent.
type JournalEntry struct {
	ID           string              `json:"id"`
	Op           string              `json:"op"`
	Path         string              `json:"path"`
	PreHash      string              `json:"pre_hash"`
	PostHash     string              `json:"post_hash"`
	Temp         string              `json:"temp,omitempty"`
	Provenance   *manifest.FileEntry `json:"provenance,omitempty"`
	State        string              `json:"state"`
	ObservedHash string              `json:"observed_hash,omitempty"`
}

type journalDoc struct {
	Entries []JournalEntry `json:"entries"`
}

// LoadJournal reads the wiring journal. A missing journal is empty.
func LoadJournal(projectRoot string) ([]JournalEntry, error) {
	raw, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(JournalRelPath)))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read wiring journal: %w", err)
	}
	var doc journalDoc
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse wiring journal: %w", err)
	}
	return doc.Entries, nil
}

// saveJournal replaces the journal atomically. The caller holds the wiring lock.
func saveJournal(projectRoot string, entries []JournalEntry) error {
	raw, err := json.MarshalIndent(journalDoc{Entries: entries}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal wiring journal: %w", err)
	}
	return writeAtomic(filepath.Join(projectRoot, filepath.FromSlash(JournalRelPath)), append(raw, '\n'))
}

// appendJournal adds e to the journal.
func appendJournal(projectRoot string, e JournalEntry) error {
	entries, err := LoadJournal(projectRoot)
	if err != nil {
		return err
	}
	return saveJournal(projectRoot, append(entries, e))
}

// setJournalState records the outcome of the entry with the given ID.
func setJournalState(projectRoot, id, state, observed string) error {
	entries, err := LoadJournal(projectRoot)
	if err != nil {
		return err
	}
	for i := range entries {
		if entries[i].ID == id {
			entries[i].State = state
			if observed != "" {
				entries[i].ObservedHash = observed
			}
		}
	}
	return saveJournal(projectRoot, entries)
}

// randomToken returns 16 hex characters for entry IDs and temp file names.
func randomToken() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
