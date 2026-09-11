package constitution

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Evolution-log reader (SPEC-CON-AMEND-APPLY-001 REQ-CAA-006 … REQ-CAA-009).
//
// The log is a human-edited markdown file: a HISTORY table with |---| rows,
// --- horizontal rules, prose, fenced yaml blocks written by hand, and
// ---delimited machine entries appended by the writer. A candidate block is
// either a fenced yaml block or a ---delimited segment outside every fence;
// a candidate is an entry when it decodes to a mapping with a non-empty
// top-level id.

// Keys that feed each entry field, in precedence order: snake_case first, then
// the legacy concatenated key of the untagged writer, then the human-format key.
var (
	logKeysRuleID         = []string{"rule_id", "ruleid", "const_registry_entry"}
	logKeysApprovedAt     = []string{"approved_at", "approvedat", "timestamp"}
	logKeysZoneBefore     = []string{"zone_before", "zonebefore", "zone"}
	logKeysZoneAfter      = []string{"zone_after", "zoneafter", "zone"}
	logKeysApprovedBy     = []string{"approved_by", "approvedby", "approver"}
	logKeysCanaryVerdict  = []string{"canary_verdict", "canaryverdict"}
	logKeysClauseBefore   = []string{"clause_before", "clausebefore"}
	logKeysClauseAfter    = []string{"clause_after", "clauseafter"}
	logKeysRolledBack     = []string{"rolled_back", "rolledback"}
	logKeysRollbackReason = []string{"rollback_reason", "rollbackreason"}
	logKeysRollbackAt     = []string{"rollback_at", "rollbackat"}
	logKeysContradictions = []string{"contradictions"}
)

// logBlock is one candidate block of the log.
type logBlock struct {
	// firstLine is the 1-based file line of the block's first content line.
	firstLine int
	text      string
}

// logField is one key/value pair of a decoded candidate mapping.
type logField struct {
	key   *yaml.Node
	value *yaml.Node
}

// logEntryDecoder resolves the fields of one candidate block and converts
// yaml node lines into file lines.
type logEntryDecoder struct {
	path   string
	block  logBlock
	fields map[string]logField
	id     string
	idLine int
}

// parseEvolutionLog returns every entry of the log content in file order.
// A recognized entry without a parseable approval timestamp is an error
// (fail closed, REQ-CAA-009).
func parseEvolutionLog(path string, content string) ([]AmendmentLog, error) {
	type located struct {
		line  int
		entry AmendmentLog
	}
	var found []located

	for _, b := range evolutionLogBlocks(content) {
		entry, ok, err := decodeLogBlock(path, b)
		if err != nil {
			return nil, err
		}
		if ok {
			found = append(found, located{line: b.firstLine, entry: entry})
		}
	}

	sort.SliceStable(found, func(i, j int) bool { return found[i].line < found[j].line })
	logs := make([]AmendmentLog, 0, len(found))
	for _, f := range found {
		logs = append(logs, f.entry)
	}
	return logs, nil
}

// evolutionLogBlocks returns the candidate blocks: every fenced yaml block, and
// every ---delimited segment that lies outside all fences. The --- lines are
// scanned line-anchored; a segment that is not an entry advances the scan by
// one delimiter, never by a pair, so the number of horizontal rules before an
// entry cannot shift the pairing (REQ-CAA-007).
func evolutionLogBlocks(content string) []logBlock {
	lines := strings.Split(content, "\n")
	inFence := make([]bool, len(lines))
	var blocks []logBlock

	// Fences: any ``` block is excluded from the --- scan; yaml fences are
	// candidates.
	for i := 0; i < len(lines); i++ {
		open := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(open, "```") {
			continue
		}
		closeAt := -1
		for j := i + 1; j < len(lines); j++ {
			if strings.HasPrefix(strings.TrimSpace(lines[j]), "```") {
				closeAt = j
				break
			}
		}
		if closeAt < 0 {
			break
		}
		for k := i; k <= closeAt; k++ {
			inFence[k] = true
		}
		if info := strings.TrimSpace(strings.TrimPrefix(open, "```")); info == "yaml" || info == "yml" {
			blocks = append(blocks, logBlock{
				firstLine: i + 2,
				text:      strings.Join(lines[i+1:closeAt], "\n"),
			})
		}
		i = closeAt
	}

	var delims []int
	for i, line := range lines {
		if !inFence[i] && strings.TrimRight(line, " \t\r") == "---" {
			delims = append(delims, i)
		}
	}
	for i := 0; i+1 < len(delims); {
		from, to := delims[i], delims[i+1]
		crossesFence := false
		for k := from + 1; k < to; k++ {
			if inFence[k] {
				crossesFence = true
				break
			}
		}
		if crossesFence {
			i++
			continue
		}
		b := logBlock{firstLine: from + 2, text: strings.Join(lines[from+1:to], "\n")}
		if isLogEntryCandidate(b) {
			blocks = append(blocks, b)
			i += 2
			continue
		}
		i++
	}
	return blocks
}

// isLogEntryCandidate reports whether a segment decodes to a mapping with a
// non-empty top-level id.
func isLogEntryCandidate(b logBlock) bool {
	d, ok := newLogEntryDecoder("", b)
	return ok && d.id != ""
}

// decodeLogBlock decodes one candidate block. ok is false when the block is
// not an entry (not a mapping, or no non-empty id).
func decodeLogBlock(path string, b logBlock) (AmendmentLog, bool, error) {
	d, ok := newLogEntryDecoder(path, b)
	if !ok || d.id == "" {
		return AmendmentLog{}, false, nil
	}
	entry, err := d.entry()
	if err != nil {
		return AmendmentLog{}, false, err
	}
	return entry, true, nil
}

// newLogEntryDecoder decodes the block into a yaml node tree. ok is false
// when the block is not valid yaml or not a mapping.
func newLogEntryDecoder(path string, b logBlock) (*logEntryDecoder, bool) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(b.text), &doc); err != nil {
		return nil, false
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return nil, false
	}
	m := doc.Content[0]
	d := &logEntryDecoder{path: path, block: b, fields: make(map[string]logField, len(m.Content)/2)}
	for i := 0; i+1 < len(m.Content); i += 2 {
		k, v := m.Content[i], m.Content[i+1]
		if v.Kind == yaml.AliasNode && v.Alias != nil {
			v = v.Alias
		}
		if k.Kind != yaml.ScalarNode {
			continue
		}
		if _, seen := d.fields[k.Value]; !seen {
			d.fields[k.Value] = logField{key: k, value: v}
		}
	}
	if f, ok := d.fields["id"]; ok && f.value.Kind == yaml.ScalarNode {
		d.id = strings.TrimSpace(f.value.Value)
		d.idLine = d.fileLine(f.key)
	}
	return d, true
}

// fileLine converts a node's block-relative line into the 1-based file line.
func (d *logEntryDecoder) fileLine(n *yaml.Node) int {
	return d.block.firstLine + n.Line - 1
}

// fail builds the located error of REQ-CAA-009: log path, file line, key, id.
func (d *logEntryDecoder) fail(line int, key, format string, args ...any) error {
	return fmt.Errorf("evolution log %s: line %d: key %s: entry %s: %s",
		d.path, line, key, d.id, fmt.Sprintf(format, args...))
}

// first returns the first present key of keys and its field.
func (d *logEntryDecoder) first(keys []string) (string, logField, bool) {
	for _, k := range keys {
		if f, ok := d.fields[k]; ok {
			return k, f, true
		}
	}
	return "", logField{}, false
}

// str resolves a string field; an absent key yields "".
func (d *logEntryDecoder) str(keys []string) (string, error) {
	key, f, ok := d.first(keys)
	if !ok || f.value.Tag == "!!null" {
		return "", nil
	}
	var s string
	if err := f.value.Decode(&s); err != nil {
		return "", d.fail(d.fileLine(f.key), key, "value is not a string")
	}
	return s, nil
}

// zone resolves a zone field; an absent key yields Frozen (the zero value).
func (d *logEntryDecoder) zone(keys []string) (Zone, error) {
	key, f, ok := d.first(keys)
	if !ok {
		return ZoneFrozen, nil
	}
	z, err := parseZoneValue(f.value.Value)
	if err != nil {
		return 0, d.fail(d.fileLine(f.key), key, "%v", err)
	}
	return z, nil
}

// parseLogTime parses an approval or rollback timestamp: RFC 3339, or a
// date-only YYYY-MM-DD value meaning 00:00:00 UTC of that date.
func parseLogTime(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, true
	}
	if t, err := time.Parse(time.DateOnly, s); err == nil {
		return t.UTC(), true
	}
	return time.Time{}, false
}

// entry maps the decoded fields onto an AmendmentLog (REQ-CAA-006, REQ-CAA-008).
func (d *logEntryDecoder) entry() (AmendmentLog, error) {
	e := AmendmentLog{ID: d.id}
	var err error

	// ApprovedAt: fail closed when absent or unparseable (REQ-CAA-009).
	key, f, ok := d.first(logKeysApprovedAt)
	if !ok {
		return AmendmentLog{}, d.fail(d.idLine, "approved_at", "no approval time recorded")
	}
	t, parsed := parseLogTime(f.value.Value)
	if f.value.Kind != yaml.ScalarNode || !parsed {
		return AmendmentLog{}, d.fail(d.fileLine(f.key), key,
			"approval time %q does not parse (want RFC 3339 or YYYY-MM-DD)", f.value.Value)
	}
	e.ApprovedAt = t

	for _, s := range []struct {
		keys []string
		dst  *string
	}{
		{logKeysRuleID, &e.RuleID},
		{logKeysApprovedBy, &e.ApprovedBy},
		{logKeysCanaryVerdict, &e.CanaryVerdict},
		{logKeysClauseBefore, &e.ClauseBefore},
		{logKeysClauseAfter, &e.ClauseAfter},
		{logKeysRollbackReason, &e.RollbackReason},
	} {
		if *s.dst, err = d.str(s.keys); err != nil {
			return AmendmentLog{}, err
		}
	}

	if e.ZoneBefore, err = d.zone(logKeysZoneBefore); err != nil {
		return AmendmentLog{}, err
	}
	if e.ZoneAfter, err = d.zone(logKeysZoneAfter); err != nil {
		return AmendmentLog{}, err
	}

	if key, f, ok := d.first(logKeysRolledBack); ok {
		if err := f.value.Decode(&e.RolledBack); err != nil {
			return AmendmentLog{}, d.fail(d.fileLine(f.key), key, "value is not a boolean")
		}
	}

	if key, f, ok := d.first(logKeysRollbackAt); ok && f.value.Tag != "!!null" && strings.TrimSpace(f.value.Value) != "" {
		rt, parsed := parseLogTime(f.value.Value)
		if !parsed {
			return AmendmentLog{}, d.fail(d.fileLine(f.key), key, "rollback timestamp %q does not parse", f.value.Value)
		}
		e.RollbackAt = &rt
	}

	if key, f, ok := d.first(logKeysContradictions); ok && f.value.Tag != "!!null" {
		if err := f.value.Decode(&e.Contradictions); err != nil {
			return AmendmentLog{}, d.fail(d.fileLine(f.key), key, "value is not a list of strings")
		}
	}

	return e, nil
}
