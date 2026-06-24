// Package spec — one-time §E.5→§E.4 backfill migration.
//
// migrate_3phase.go implements REQ-LR-007 of SPEC-V3R6-LIFECYCLE-REDESIGN-001:
// a one-time migration that folds §E.5 Mx-phase Audit-Ready Signal content into
// §E.4 Sync-phase Audit-Ready Signal for modern-era V3R6 SPECs that still carry
// the legacy 5-section layout. Grandfather-protected SPECs (V2.x / V3R2-R4 / V3R5)
// are SKIPPED per N4 / AP-LR-P-004 (no retroactive rewrite of era-protected SPECs).
//
// The migration is invoked once at run-phase M3. After it completes, every migrated
// SPEC classifies V3R6 via the new H-4 predicate (§E.2 + §E.4 + sync_commit_sha)
// rather than the H-4-legacy fallback, and the §E.5 section is gone. A migration
// log at .moai/state/lifecycle-redesign-migration.json records each folded entry.
package spec

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MigrationLogEntry records a single SPEC's §E.5→§E.4 fold.
type MigrationLogEntry struct {
	SpecID              string    `json:"spec_id"`
	Era                 string    `json:"era"`
	FoldedFrom          string    `json:"folded_from"`            // "§E.5"
	FoldedTo            string    `json:"folded_to"`              // "§E.4"
	MxCommitSha         string    `json:"mx_commit_sha,omitempty"`
	MigratedAt          time.Time `json:"migrated_at"`
	MigrationCommitSHA  string    `json:"migration_commit_sha,omitempty"`
}

// MigrationLog is the on-disk record at .moai/state/lifecycle-redesign-migration.json.
type MigrationLog struct {
	MigrationDate string              `json:"migration_date"`
	Spec          string              `json:"spec"` // SPEC-V3R6-LIFECYCLE-REDESIGN-001
	Entries       []MigrationLogEntry `json:"entries"`
	Skipped       int                 `json:"skipped_grandfathered"`
}

// section5Heading is the canonical §E.5 heading literal the migrator detects.
const section5Heading = "## §E.5"

// section4Heading is the canonical §E.4 heading literal the migrator anchors on.
const section4Heading = "## §E.4"

// migratedSubheading is the sub-heading inserted under §E.4 carrying folded §E.5 content.
const migratedSubheading = "### (Migrated from §E.5)"

// MigrateProgressMD folds the §E.5 section(s) of a progress.md body into §E.4.
//
// Returns (newContent, mxCommitSHA, folded) where folded=true when at least one §E.5
// section was found and folded. When no §E.5 section is present, the content is
// returned unchanged with folded=false (idempotent — re-running on an already-migrated
// file is a no-op). If multiple §E.5 sections exist (an authoring outlier), ALL are
// folded; mxCommitSHA reflects the last one processed.
//
// The fold (per §E.5 section):
//  1. Extracts the §E.5 section body (from the "## §E.5" heading line to the next
//     "---", "## ", or EOF).
//  2. Appends it under §E.4 as a "### (Migrated from §E.5)" sub-heading.
//  3. Removes the "## §E.5" heading + its body block.
//
// The function is pure (string in, string out); disk I/O is the caller's responsibility.
func MigrateProgressMD(content string) (newContent string, mxCommitSHA string, folded bool) {
	if !strings.Contains(content, section5Heading) {
		return content, "", false // no §E.5 section — idempotent no-op
	}

	current := content
	var lastMX string
	didFold := false
	// Loop: some SPECs carry multiple §E.5 sections (authoring outlier); fold all of them.
	for strings.Contains(current, section5Heading) {
		lines := strings.Split(current, "\n")

		// Locate §E.5 section start.
		s5Start := -1
		for i, l := range lines {
			if strings.HasPrefix(l, section5Heading) {
				s5Start = i
				break
			}
		}
		if s5Start == -1 {
			break
		}
		// Section end = first line after s5Start that is a top-level section delimiter
		// ("---" at column 0, or a "## " heading, or EOF).
		s5End := len(lines)
		for i := s5Start + 1; i < len(lines); i++ {
			l := lines[i]
			if l == "---" || strings.HasPrefix(l, "## ") {
				s5End = i
				break
			}
		}
		// Capture the §E.5 body (excluding the heading line itself; trim leading blanks).
		s5BodyLines := lines[s5Start+1 : s5End]
		for len(s5BodyLines) > 0 && strings.TrimSpace(s5BodyLines[0]) == "" {
			s5BodyLines = s5BodyLines[1:]
		}
		for len(s5BodyLines) > 0 && strings.TrimSpace(s5BodyLines[len(s5BodyLines)-1]) == "" {
			s5BodyLines = s5BodyLines[:len(s5BodyLines)-1]
		}
		s5Body := strings.Join(s5BodyLines, "\n")
		if mx := extractProgressField(s5Body, "mx_commit_sha"); mx != "" {
			lastMX = mx
		}

		// Remove the §E.5 section (heading + body + trailing "---" delimiter if present).
		var out []string
		out = append(out, lines[:s5Start]...)
		rest := lines[s5End:]
		if len(rest) > 0 && rest[0] == "---" {
			rest = rest[1:]
		}
		out = append(out, rest...)
		contentWithoutS5 := strings.Join(out, "\n")

		// Insert the folded body under §E.4 as a sub-heading.
		foldedBlock := fmt.Sprintf("\n%s\n\n%s\n", migratedSubheading, s5Body)
		current = insertFoldUnderSection4(contentWithoutS5, foldedBlock)
		didFold = true
	}
	return current, lastMX, didFold
}

// insertFoldUnderSection4 inserts foldedBlock at the end of the §E.4 section body.
// If §E.4 is absent, appends foldedBlock as a new top-level section.
func insertFoldUnderSection4(content, foldedBlock string) string {
	lines := strings.Split(content, "\n")
	s4Start := -1
	for i, l := range lines {
		if strings.HasPrefix(l, section4Heading) {
			s4Start = i
			break
		}
	}
	if s4Start == -1 {
		// §E.4 absent — append as a fresh section.
		return strings.TrimRight(content, "\n") + "\n\n" + section4Heading + " Audit-Ready Signal\n" + foldedBlock
	}
	// Find §E.4 section end (next "---" or "## " heading or EOF after s4Start).
	s4End := len(lines)
	for i := s4Start + 1; i < len(lines); i++ {
		l := lines[i]
		if l == "---" || strings.HasPrefix(l, "## ") {
			s4End = i
			break
		}
	}
	// Trim trailing blanks in §E.4 body, then insert the folded block.
	insertAt := s4End
	for insertAt > s4Start+1 && strings.TrimSpace(lines[insertAt-1]) == "" {
		insertAt--
	}
	newLines := make([]string, 0, len(lines)+4)
	newLines = append(newLines, lines[:insertAt]...)
	newLines = append(newLines, strings.Split(strings.TrimRight(foldedBlock, "\n"), "\n")...)
	newLines = append(newLines, lines[insertAt:]...)
	return strings.Join(newLines, "\n")
}

// RunMigration iterates all SPECs under baseDir/.moai/specs/, classifies each by era,
// and folds §E.5→§E.4 for modern-era V3R6 SPECs with the legacy 5-section layout.
// Grandfather-protected SPECs (V2.x / V3R2-R4 / V3R5) are skipped (N4). A migration
// log is written to logPath. Returns the populated MigrationLog.
//
// dryRun=true performs the classification + fold computation + log population WITHOUT
// writing any progress.md changes (the log is still written so the operator can
// preview the affected set).
func RunMigration(baseDir, logPath string, dryRun bool) (*MigrationLog, error) {
	if baseDir == "" {
		baseDir = "."
	}
	specsDir := filepath.Join(baseDir, ".moai", "specs")
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return &MigrationLog{Spec: "SPEC-V3R6-LIFECYCLE-REDESIGN-001"}, nil
		}
		return nil, fmt.Errorf("read specs dir: %w", err)
	}

	log := &MigrationLog{
		MigrationDate: time.Now().UTC().Format("2006-01-02"),
		Spec:          "SPEC-V3R6-LIFECYCLE-REDESIGN-001",
		Entries:       []MigrationLogEntry{},
	}

	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "SPEC-") {
			continue
		}
		specDir := filepath.Join(specsDir, e.Name())
		signals, err := LoadEraSignalsFromDir(specDir)
		if err != nil {
			continue // skip unreadable SPECs
		}
		era, _ := ClassifyEra(signals)
		// N4 / AP-LR-P-004: skip grandfather-protected + non-modern SPECs.
		if !era.IsModern() {
			log.Skipped++
			continue
		}
		// Only migrate SPECs carrying the legacy §E.5 section.
		progressPath := filepath.Join(specDir, "progress.md")
		raw, rerr := os.ReadFile(progressPath)
		if rerr != nil {
			continue
		}
		newContent, mxSHA, folded := MigrateProgressMD(string(raw))
		if !folded {
			continue // already migrated or no §E.5 — idempotent skip
		}
		if !dryRun {
			if werr := os.WriteFile(progressPath, []byte(newContent), 0644); werr != nil {
				return nil, fmt.Errorf("write %s: %w", progressPath, werr)
			}
		}
		log.Entries = append(log.Entries, MigrationLogEntry{
			SpecID:             e.Name(),
			Era:                string(era),
			FoldedFrom:         "§E.5",
			FoldedTo:           "§E.4",
			MxCommitSha:        mxSHA,
			MigratedAt:         time.Now().UTC(),
			MigrationCommitSHA: "", // backfilled by caller after the commit lands
		})
	}

	if logPath != "" {
		if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
			return log, fmt.Errorf("mkdir log dir: %w", err)
		}
		b, _ := json.MarshalIndent(log, "", "  ")
		if err := os.WriteFile(logPath, b, 0644); err != nil {
			return log, fmt.Errorf("write log: %w", err)
		}
	}
	return log, nil
}
