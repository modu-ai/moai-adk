// journal.go — the pending-install recovery journal (SPEC-USER-ASSET-INSTALL-001,
// final-class item 5 + directed repair R-e + in-round extensions E4/E5 + R-f-①).
//
// The READ → asset-changes → SAVE ordering has a crash window: an install
// interrupted AFTER asset writes but BEFORE the manifest save leaves
// moai-written files the manifest does not track, and the retry would
// classify them as REQ-010 collision-skips forever. Before its asset-write
// phase, a run writes this journal recording the intended delta; the next
// run reconciles an existing journal BEFORE any install/collision judgment.
//
// COMPLETENESS (R-e): per entry AND for the run the journal records
// (1) the bundle-SELECTION delta — the bundles: list change the interrupted
// run intended — so recovery never restores an empty selection;
// (2) the FULL manifest-entry provenance per file (bundle, moai_version,
// installed_at), so a different-binary replay restores the same versions;
// (3) the ownership evidence — the write-completion flag, reconciled
// atomically with the manifest save.
//
// RECOVERY LATTICE (E5 — exactly three cases, never a fourth):
//  1. target ABSENT → install from the journal entry (R-f-①: a retry from a
//     different binary writes the CURRENT binary's bytes and re-stamps the
//     provenance honestly — REQ-006's per-file version names the build that
//     produced the bytes on disk);
//  2. target PRESENT, hash == the staged sha256 → CLAIM as the run's own
//     install (the staging record is the intent-and-content proof; rename(2)
//     is atomic, so the flag is not required — E4);
//  3. target PRESENT, hash MISMATCH → NEVER reinstall: the user may have
//     edited after the interrupted install. Classify per the ownership
//     evidence: flag-complete → REQ-023 divergence (backup + report);
//     unflagged → REQ-010 collision. Both preserve the user's bytes.
package userassets

import (
	"encoding/json"
	"fmt"
	"os"
)

// JournalEntry is one staged install's intent-and-content record.
type JournalEntry struct {
	// Path is the root-slug-relative install target (e.g.
	// "claude-skills/moai-workflow-spec/SKILL.md").
	Path string `json:"path"`
	// ExpectedSHA256 is the sha256 of the bytes the run intended to write.
	ExpectedSHA256 string `json:"expected_sha256"`
	// Full provenance (R-e item 2).
	Bundle      string `json:"bundle"`
	MoaiVersion string `json:"moai_version"`
	InstalledAt string `json:"installed_at"`
	// WriteCompleted records that the run observed its own write complete
	// (the flag set after the atomic rename lands, before the manifest
	// save). The flag separates divergence from collision at case 3 — it
	// is NOT required to claim a hash-matching file at case 2 (E4).
	WriteCompleted bool `json:"write_completed"`
}

// PendingJournal is the run's staged intent, written before the asset
// writes and cleared atomically with the manifest save.
type PendingJournal struct {
	SchemaVersion int `json:"schema_version"`
	// BundlesSelection is the intended bundle-SELECTION delta for the run
	// (R-e item 1) — the state the manifest's bundles: list should carry,
	// never an empty stand-in.
	BundlesSelection []string       `json:"bundles_selection"`
	Entries          []JournalEntry `json:"entries"`
	StartedAt        string         `json:"started_at"`
}

// WriteJournal stages the journal atomically.
func WriteJournal(path string, j *PendingJournal) error {
	if j.SchemaVersion == 0 {
		j.SchemaVersion = SchemaVersion
	}
	data, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return fmt.Errorf("userassets: encode journal: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(dirOf(path), 0o755); err != nil {
		return fmt.Errorf("userassets: mkdir journal home: %w", err)
	}
	return atomicWrite(path, data, 0o644)
}

// LoadJournal reads the journal. Absent → (nil, nil). A journal whose
// schema_version this binary does not write is refused with a diagnostic —
// a silent decode would mis-read an unknown schema's recovery data as if it
// were this binary's (SPEC-USERASSET-DEPLOY-GUARD-001 M1, REQ-JRN-004).
func LoadJournal(path string) (*PendingJournal, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("userassets: read journal: %w", err)
	}
	var j PendingJournal
	if err := json.Unmarshal(data, &j); err != nil {
		return nil, fmt.Errorf("userassets: journal corrupt at %s: %w", path, err)
	}
	if j.SchemaVersion != SchemaVersion {
		return nil, fmt.Errorf("userassets: journal at %s carries schema_version %d, this binary writes %d — refusing to decode (preserve the file; the interrupted install's selection + ownership recovery data must not be silently reinterpreted)", path, j.SchemaVersion, SchemaVersion)
	}
	return &j, nil
}

// ClearJournal removes the journal (atomic with the manifest save — the
// caller writes both under the user lock and clears the journal only after
// the manifest save succeeded).
func ClearJournal(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("userassets: clear journal: %w", err)
	}
	return nil
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == os.PathSeparator {
			return path[:i]
		}
	}
	return "."
}
