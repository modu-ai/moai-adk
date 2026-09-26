package escalation

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// maxRetripOrdinal bounds the search for the next free re-trip ordinal.
const maxRetripOrdinal = 1000

// WriteRecord writes one trip as an escalation record and returns its path
// (REQ-AE-018 to REQ-AE-020). The record's class and fingerprint select the
// file:
//   - no record yet: a new open record, occurrences 1;
//   - an open record: its occurrences is incremented and updated_at
//     refreshed, every other byte of it kept;
//   - a resolved record: left byte for byte, the trip going to the next
//     ordinal <class>-<fingerprint>-<n>.md (n >= 2) under the same rules.
//
// The detector never writes decider; a new record starts with decider "".
//
// @MX:ANCHOR: [AUTO] the single escalation record writer — dedup, occurrence counting, and re-trip naming
// @MX:REASON: every class detector (1-10) reports through it; REQ-AE-020's dedup and never-overwrite-a-resolved-record guarantees hold only while all writes go through this function
func WriteRecord(worktreeRoot string, r Record, now time.Time) (string, error) {
	if r.Card == "" || r.Class == "" || r.Fingerprint == "" {
		return "", errors.New("escalation: record needs card, class, and fingerprint")
	}
	stamp := now.UTC().Format(time.RFC3339)
	for n := 1; n <= maxRetripOrdinal; n++ {
		path := RecordPath(worktreeRoot, r.Card, r.Class, r.Fingerprint, n)
		data, err := os.ReadFile(path)
		if errors.Is(err, fs.ErrNotExist) {
			return path, createRecord(path, r, stamp)
		}
		if err != nil {
			return "", fmt.Errorf("escalation: read record %s: %w", path, err)
		}
		existing, err := ParseRecord(data)
		if err != nil {
			return "", fmt.Errorf("escalation: record %s: %w", path, err)
		}
		if existing.Status == StatusResolved {
			continue
		}
		existing.Occurrences++
		existing.UpdatedAt = stamp
		return path, rewriteFrontmatter(path, data, existing)
	}
	return "", fmt.Errorf("escalation: no free re-trip ordinal for %s-%s", r.Class, r.Fingerprint)
}

// createRecord writes a new open record at path.
func createRecord(path string, r Record, stamp string) error {
	r.SchemaVersion = RecordSchemaVersion
	r.Status = StatusOpen
	r.Decider = ""
	r.Occurrences = 1
	r.DetectedAt, r.UpdatedAt = stamp, stamp
	data, err := r.Marshal()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("escalation: create record dir: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("escalation: write record: %w", err)
	}
	return nil
}

// rewriteFrontmatter replaces the frontmatter of an existing record with r's
// and keeps the body bytes unchanged.
func rewriteFrontmatter(path string, data []byte, r Record) error {
	text := string(data)
	end := strings.Index(text[4:], "\n---\n")
	if !strings.HasPrefix(text, "---\n") || end < 0 {
		return errNoFrontmatter
	}
	body := text[4+end+len("\n---\n"):]
	if r.NotObserved == nil {
		r.NotObserved = []string{}
	}
	fm, err := yaml.Marshal(r)
	if err != nil {
		return fmt.Errorf("escalation: encode record frontmatter: %w", err)
	}
	out := "---\n" + string(fm) + "---\n" + body
	if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
		return fmt.Errorf("escalation: rewrite record: %w", err)
	}
	return nil
}

// ContractLine returns the 1-based line of the mapping key at keyPath in a
// contract (for example "budget", "operations"), or 0 when absent. Verify
// output carries no line numbers, so contract_ref lines are mapped from the
// file text (spec.md §F O6).
func ContractLine(data []byte, keyPath ...string) int {
	node := contractNode(data, keyPath...)
	if node == nil {
		return 0
	}
	return node.key.Line
}

// ContractItemLine returns the line of the sequence item equal to value under
// keyPath (for example a glob in ownership.never), or 0 when absent.
func ContractItemLine(data []byte, value string, keyPath ...string) int {
	node := contractNode(data, keyPath...)
	if node == nil || node.value.Kind != yaml.SequenceNode {
		return 0
	}
	for _, item := range node.value.Content {
		if item.Value == value {
			return item.Line
		}
	}
	return 0
}

// keyValue is one mapping entry.
type keyValue struct{ key, value *yaml.Node }

// contractNode walks keyPath through nested mappings.
func contractNode(data []byte, keyPath ...string) *keyValue {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil || len(doc.Content) == 0 {
		return nil
	}
	cur := doc.Content[0]
	var found *keyValue
	for _, k := range keyPath {
		if cur == nil || cur.Kind != yaml.MappingNode {
			return nil
		}
		found = nil
		for i := 0; i+1 < len(cur.Content); i += 2 {
			if cur.Content[i].Value == k {
				found = &keyValue{key: cur.Content[i], value: cur.Content[i+1]}
				break
			}
		}
		if found == nil {
			return nil
		}
		cur = found.value
	}
	return found
}
