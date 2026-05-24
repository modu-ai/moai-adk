// SPEC: SPEC-V3R6-HARNESS-PROPOSAL-GEN-001 REQ-PGN-001..003.
//
// reader.go parses the canonical tier-promotions.jsonl learning log line by
// line into internal/harness.Promotion records. The reader is intentionally
// tolerant: malformed JSONL lines are skipped and surfaced via the
// malformed-lines counter rather than aborting the entire read.
//
// Missing or empty input files return an empty slice with no error to
// preserve the V3R4 graceful no-op contract (REQ-PGN-003).
package proposalgen

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"

	"github.com/modu-ai/moai-adk/internal/harness"
)

// ReadPromotions parses tier-promotions.jsonl at the supplied path.
//
// Return contract per REQ-PGN-001..003:
//   - On missing or empty file: returns (empty slice, 0, nil) for graceful
//     no-op semantics. The mapper interprets the empty slice and emits the
//     "tier-promotions.jsonl absent or empty" reason.
//   - On malformed JSONL lines: skips the line, increments the malformed
//     counter, and continues processing remaining lines.
//   - On unrecoverable IO error (permission denied, corrupted bufio state):
//     returns the wrapped error after preserving any records read so far.
//
// Blank lines are not malformed — they are silently ignored.
func ReadPromotions(path string) (promotions []harness.Promotion, malformed int, err error) {
	f, openErr := os.Open(path)
	if openErr != nil {
		// Missing input is the canonical no-op path (REQ-PGN-003); do not
		// surface ENOENT as an error.
		if errors.Is(openErr, fs.ErrNotExist) {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("proposalgen reader: open %q: %w", path, openErr)
	}
	defer func() { _ = f.Close() }()

	promotions = make([]harness.Promotion, 0)
	scanner := bufio.NewScanner(f)
	// Allow lines up to 1 MiB to accommodate large pattern_key payloads
	// without truncation. Default 64 KiB is generally sufficient but the
	// expanded buffer hedges against future format growth.
	const maxLine = 1 << 20
	scanner.Buffer(make([]byte, 0, 64<<10), maxLine)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		// Skip whitespace-only lines.
		blank := true
		for _, b := range line {
			if b != ' ' && b != '\t' && b != '\r' {
				blank = false
				break
			}
		}
		if blank {
			continue
		}

		p, perr := parseLine(line)
		if perr != nil {
			malformed++
			continue
		}
		promotions = append(promotions, p)
	}
	if scanErr := scanner.Err(); scanErr != nil {
		return promotions, malformed, fmt.Errorf("proposalgen reader: scan %q: %w", path, scanErr)
	}
	return promotions, malformed, nil
}

// parseLine unmarshals a single JSONL line into a Promotion.
func parseLine(line []byte) (harness.Promotion, error) {
	var p harness.Promotion
	if err := json.Unmarshal(line, &p); err != nil {
		return harness.Promotion{}, fmt.Errorf("unmarshal: %w", err)
	}
	// Reject records missing the canonical pattern_key — these are
	// schema-incompatible even if they parse as JSON.
	if p.PatternKey == "" {
		return harness.Promotion{}, errors.New("missing pattern_key")
	}
	return p, nil
}

// uniquePatternKeys returns the deduplicated, alphabetically sorted set of
// pattern_key values present in the input promotions. The deterministic
// ordering supports stable test assertions and downstream JSON output
// stability.
func uniquePatternKeys(promotions []harness.Promotion) []string {
	seen := make(map[string]struct{}, len(promotions))
	for _, p := range promotions {
		seen[p.PatternKey] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
