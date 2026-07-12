package routing

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"time"
)

// Filter composes optional, ANDed ledger read filters (REQ-HEV-008). A zero
// Filter matches every row.
type Filter struct {
	Subcommand string     // exact match on matched_subcommand when non-empty
	Outcome    string     // exact match on outcome when non-empty
	Since      *time.Time // inclusive lower bound on ts when non-nil
	Until      *time.Time // inclusive upper bound on ts when non-nil
}

// Reader streams ledger rows applying composable filters. Malformed lines are
// skipped fail-open (the skipped count is returned, never a panic) so a single
// corrupt line never denies the whole ledger to Loop 1 consumers (REQ-HEV-008).
type Reader struct{ path string }

// NewReader constructs a Reader over the ledger at path.
func NewReader(path string) *Reader { return &Reader{path: path} }

// Read returns the filtered rows and the count of malformed lines skipped. An
// absent ledger file yields an empty result and a nil error (not-yet-observed
// is not an error condition).
func (r *Reader) Read(f Filter) ([]Row, int, error) {
	file, err := os.Open(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, 0, nil
		}
		return nil, 0, err
	}
	defer func() { _ = file.Close() }()

	var rows []Row
	skipped := 0
	sc := bufio.NewScanner(file)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var row Row
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			skipped++
			continue
		}
		if !f.match(row) {
			continue
		}
		rows = append(rows, row)
	}
	if err := sc.Err(); err != nil {
		return rows, skipped, err
	}
	return rows, skipped, nil
}

// match reports whether row satisfies every active filter clause.
func (f Filter) match(row Row) bool {
	if f.Subcommand != "" && row.MatchedSubcommand != f.Subcommand {
		return false
	}
	if f.Outcome != "" && string(row.Outcome) != f.Outcome {
		return false
	}
	if f.Since != nil || f.Until != nil {
		ts, err := time.Parse(time.RFC3339, row.TS)
		if err != nil {
			return false // an unparseable ts is excluded when a time window is active
		}
		if f.Since != nil && ts.Before(*f.Since) {
			return false
		}
		if f.Until != nil && ts.After(*f.Until) {
			return false
		}
	}
	return true
}
