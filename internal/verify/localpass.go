package verify

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// HasLocalPass reports whether any snapshot under projectRoot whose key's
// head portion equals head contains at least one check entry recorded with
// exit code 0 (SPEC-CI-VERDICT-PRODUCER-001 REQ-CV-006): the local-pass
// evidence source of the escalation detector's contradictory-evidence CI
// limb. Absence of snapshots, an unreadable file, or a key mismatch is a
// plain no — never an error, never a second local-pass source.
//
// @MX:NOTE: [AUTO] exit-code-only predicate keeps the detector decoupled from Conditions optionality; the snapshot key pins the head
func HasLocalPass(projectRoot, head string) bool {
	paths, _ := filepath.Glob(filepath.Join(projectRoot, SnapshotDir, "*.json"))
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var s Snapshot
		if err := json.Unmarshal(data, &s); err != nil {
			continue
		}
		if h, _, ok := strings.Cut(s.Key, ":"); !ok || h != head {
			continue
		}
		for _, c := range s.Checks {
			if c.ExitCode == 0 {
				return true
			}
		}
	}
	return false
}
