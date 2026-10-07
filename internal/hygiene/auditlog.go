package hygiene

import (
	"encoding/json"
	"path/filepath"
	"time"
)

// AuditRow is one hygiene audit record in .moai/logs/hygiene-audit.jsonl
// (REQ-HYG-004). Report mode appends exactly one summary row per unit per
// run; apply mode appends one row per action taken or skipped, each with
// its reason and signal evidence.
type AuditRow struct {
	TS      string            `json:"ts"`
	Unit    string            `json:"unit"`
	Mode    string            `json:"mode"`
	Outcome Outcome           `json:"outcome"`
	Path    string            `json:"path,omitempty"`
	Reason  string            `json:"reason,omitempty"`
	Count   int               `json:"count,omitempty"`
	Counts  map[string]int    `json:"counts,omitempty"`
	Signal  map[string]string `json:"signal,omitempty"`
}

// auditSinkName is the hygiene unit's own audit sink under the logs dir.
const auditSinkName = "hygiene-audit.jsonl"

// auditSinkPath returns the audit sink path for a log directory.
func auditSinkPath(logDir string) string {
	return filepath.Join(logDir, auditSinkName)
}

// appendAuditRows writes rows to the audit sink, one JSON object per line.
func appendAuditRows(logDir string, rows []AuditRow) error {
	if len(rows) == 0 {
		return nil
	}
	for _, row := range rows {
		if row.TS == "" {
			row.TS = time.Now().UTC().Format(time.RFC3339)
		}
		blob, err := json.Marshal(row)
		if err != nil {
			return err
		}
		if err := appendText(auditSinkPath(logDir), string(blob)+"\n"); err != nil {
			return err
		}
	}
	return nil
}

// readAuditRows parses the audit sink back into rows (diagnostic helper
// shared with the tests).
func readAuditRows(path string) ([]AuditRow, error) {
	lines, err := readLinesFile(path)
	if err != nil {
		return nil, err
	}
	var rows []AuditRow
	for _, line := range lines {
		if line == "" {
			continue
		}
		var row AuditRow
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}
