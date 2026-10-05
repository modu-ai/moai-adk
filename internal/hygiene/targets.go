package hygiene

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// sessionKeyShape matches the 36-character hyphenated UUID shape the
// state writers stamp (REQ-HYG-005). A name outside this shape has no
// session key and is out of scope (spared).
var sessionKeyShape = regexp.MustCompile(
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// lockShape matches the excluded spec-close lock class (REQ-HYG-011): the
// GC never evaluates, probes, acquires, or deletes one — a scan hit is
// reported lock-class-excluded and left untouched.
var lockShape = regexp.MustCompile(`^spec-close-.+\.lock$`)

// TargetClass names one class of the closed target registry.
type TargetClass string

// The closed target-registry classes (REQ-HYG-005). Nothing outside this
// set is evaluated, reported, or deleted.
const (
	ClassContextUsage   TargetClass = "context-usage"
	ClassGoalTriple     TargetClass = "goal"
	ClassCodexStopChain TargetClass = "codex-stop-chain"
	ClassCodexStopCap   TargetClass = "codex-stop-cap"
	ClassAgentStops     TargetClass = "agent-stops"
	ClassRoutingPending TargetClass = "routing-pending"
	ClassVerifyScratch  TargetClass = "verify"
	ClassTodoRecord     TargetClass = "todo"
)

// TargetClasses returns the closed class list (order is the enumeration
// order).
func TargetClasses() []TargetClass {
	return []TargetClass{
		ClassContextUsage,
		ClassGoalTriple,
		ClassCodexStopChain,
		ClassCodexStopCap,
		ClassAgentStops,
		ClassRoutingPending,
		ClassVerifyScratch,
		ClassTodoRecord,
	}
}

// Candidate is one enumerated target-registry entry.
type Candidate struct {
	Class TargetClass
	Key   string   // session key (empty for unresolvable shapes)
	Paths []string // paths relative to the .moai root, in deletion order
	// Dating-member path (relative) — the content-recorded timestamp this
	// candidate's age reads from. Empty ⇒ content-undatable ⇒ spared.
	DatingPath string
	// DatingField names the pinned writer field (D31 table).
	DatingField string
}

// datingDatum is one writer-verified dating rule: the JSON field name the
// class's dating member carries and how to parse it (D31-corrected table).
type datingDatum struct {
	field string
	// rfc3339String parses a JSON string timestamp; jsonTime decodes a
	// time.Time-shaped JSON value.
	rfc3339String, jsonTime bool
}

// classDating pins the per-class dating datum (REQ-HYG-005, D31): each
// field below was read out of its writer source at run phase —
// captured_at (internal/statusline/context_usage.go), created_at
// (internal/goal/schema.go — currently written empty ⇒ undatable in
// practice), recorded_at (internal/cli/codex_stop_chain.go),
// created_at (internal/harness/routing/types.go PendingRow.CreatedAt),
// stopped_at (internal/hook/agent_stop_guard.go, newest entry),
// recorded_at per entry (internal/verify/schema.go), entered_at (launcher
// session-record shape). The codex-stop-cap counter body carries NO
// timestamp field (stopCapState) ⇒ content-undatable ⇒ spared by rule.
var classDating = map[TargetClass]datingDatum{
	ClassContextUsage:   {field: "captured_at", rfc3339String: true},
	ClassGoalTriple:     {field: "created_at", rfc3339String: true},
	ClassCodexStopChain: {field: "recorded_at", jsonTime: true},
	ClassCodexStopCap:   {}, // no field ⇒ undatable ⇒ spared
	ClassAgentStops:     {field: "stopped_at", rfc3339String: true},
	ClassRoutingPending: {field: "created_at", jsonTime: true},
	ClassVerifyScratch:  {field: "recorded_at", jsonTime: true},
	ClassTodoRecord:     {field: "entered_at", rfc3339String: true},
}

// enumerateCandidates walks the closed registry classes under the .moai
// root and returns every candidate found, plus the lock-named scan hits
// (reported lock-class-excluded, REQ-HYG-011) and the unresolvable-key
// scan hits (spared, REQ-HYG-005).
//
// @MX:ANCHOR: [AUTO] enumerateCandidates — the closed-registry enumeration
// @MX:REASON: every deletion decision names a candidate from this
// enumeration; a pattern-sweep outside it would violate REQ-HYG-005's
// closed-registry guarantee, so this function is the only path that
// produces deletion candidates.
func enumerateCandidates(moaiRoot string) (candidates []Candidate, lockHits []string, unresolvable []Candidate, err error) {
	stateDir := filepath.Join(moaiRoot, "state")

	// One-level lock-class scan of the state root: report-only exclusion.
	entries, err := os.ReadDir(stateDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, nil, nil, err
	}
	for _, e := range entries {
		if !e.IsDir() && lockShape.MatchString(e.Name()) {
			lockHits = append(lockHits, filepath.Join("state", e.Name()))
		}
	}

	// Per-class enumeration.
	for _, e := range safeReadDir(filepath.Join(stateDir, "context-usage")) {
		if key, ok := trimSuffixKey(e.Name(), ".json"); ok && sessionKeyShape.MatchString(key) {
			candidates = append(candidates, Candidate{
				Class: ClassContextUsage, Key: key,
				Paths:       []string{filepath.Join("state", "context-usage", e.Name())},
				DatingPath:  filepath.Join("state", "context-usage", e.Name()),
				DatingField: classDating[ClassContextUsage].field,
			})
		}
	}
	for _, e := range safeReadDir(filepath.Join(stateDir, "goal")) {
		if key, ok := trimSuffixKey(e.Name(), ".json"); ok && sessionKeyShape.MatchString(key) {
			c := Candidate{
				Class: ClassGoalTriple, Key: key,
				// D29 deletion order: siblings first, the dating member LAST.
				Paths: []string{
					filepath.Join("state", "goal", key+".verdict.json"),
					filepath.Join("state", "goal", key+".html"),
					filepath.Join("state", "goal", key+".json"),
				},
				DatingPath:  filepath.Join("state", "goal", key+".json"),
				DatingField: classDating[ClassGoalTriple].field,
			}
			if fileExists(filepath.Join(moaiRoot, filepath.FromSlash(c.Paths[2]))) {
				candidates = append(candidates, c)
			}
		}
	}
	for _, class := range []TargetClass{ClassCodexStopChain, ClassCodexStopCap, ClassAgentStops} {
		for _, e := range safeReadDir(filepath.Join(stateDir, string(class))) {
			if key, ok := trimSuffixKey(e.Name(), ".json"); ok && sessionKeyShape.MatchString(key) {
				candidates = append(candidates, Candidate{
					Class: class, Key: key,
					Paths:       []string{filepath.Join("state", string(class), e.Name())},
					DatingPath:  filepath.Join("state", string(class), e.Name()),
					DatingField: classDating[class].field,
				})
			}
		}
	}
	for _, e := range safeReadDir(stateDir) {
		if key, ok := trimPrefixKey(e.Name(), "routing-pending-", ".json"); ok && sessionKeyShape.MatchString(key) {
			candidates = append(candidates, Candidate{
				Class: ClassRoutingPending, Key: key,
				Paths:       []string{filepath.Join("state", e.Name())},
				DatingPath:  filepath.Join("state", e.Name()),
				DatingField: classDating[ClassRoutingPending].field,
			})
		}
	}
	// Verify scratch: session-keyed directories, excluding the snapshot
	// store; entries are removed entry-by-entry (REQ-HYG-005).
	for _, e := range safeReadDir(filepath.Join(stateDir, "verify")) {
		if !e.IsDir() || !sessionKeyShape.MatchString(e.Name()) {
			continue
		}
		c := Candidate{
			Class: ClassVerifyScratch, Key: e.Name(),
			DatingField: classDating[ClassVerifyScratch].field,
		}
		dir := filepath.Join(stateDir, "verify", e.Name())
		for _, entry := range safeReadDir(dir) {
			if entry.IsDir() {
				continue // nested directories are not per-entry datable shapes
			}
			c.Paths = append(c.Paths, filepath.Join("state", "verify", e.Name(), entry.Name()))
		}
		sortRelPaths(c.Paths)
		if len(c.Paths) > 0 {
			candidates = append(candidates, c)
		}
	}
	// Todo session records: session-keyed files directly under state/todo;
	// the shared backlog stores and directory entries are excluded
	// regardless (REQ-HYG-005).
	for _, e := range safeReadDir(filepath.Join(stateDir, "todo")) {
		if e.IsDir() {
			continue
		}
		if key, ok := trimSuffixKey(e.Name(), ".json"); ok && sessionKeyShape.MatchString(key) {
			candidates = append(candidates, Candidate{
				Class: ClassTodoRecord, Key: key,
				Paths:       []string{filepath.Join("state", "todo", e.Name())},
				DatingPath:  filepath.Join("state", "todo", e.Name()),
				DatingField: classDating[ClassTodoRecord].field,
			})
		}
	}

	// Unresolvable keys: files in the class shapes whose name carries no
	// session key are reported spared (out of scope).
	unresolvable = append(unresolvable, unresolvableIn(filepath.Join(stateDir, "context-usage"), ".json", ClassContextUsage)...)
	return candidates, lockHits, unresolvable, nil
}

// candidateDate reads the class's content-recorded dating datum from the
// candidate's dating member and reports whether the age is datable (a
// missing field, an empty value, or an unparseable timestamp is
// undatable — REQ-HYG-009; file mtime is never consulted).
func candidateDate(moaiRoot string, c Candidate) (time.Time, bool) {
	datum, known := classDating[c.Class]
	if !known || datum.field == "" || c.DatingPath == "" {
		return time.Time{}, false
	}
	if c.Class == ClassAgentStops {
		return newestStoppedAt(filepath.Join(moaiRoot, filepath.FromSlash(c.DatingPath)))
	}
	if c.Class == ClassVerifyScratch {
		// The verify class is dated per entry; the group-level date is the
		// dating of its oldest dated entry.
		return oldestEntryDate(moaiRoot, c)
	}
	blob, err := os.ReadFile(filepath.Join(moaiRoot, filepath.FromSlash(c.DatingPath)))
	if err != nil {
		return time.Time{}, false
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(blob, &body); err != nil {
		return time.Time{}, false
	}
	raw, ok := body[datum.field]
	if !ok {
		return time.Time{}, false
	}
	if datum.jsonTime {
		var ts time.Time
		if err := json.Unmarshal(raw, &ts); err != nil || ts.IsZero() {
			return time.Time{}, false
		}
		return ts, true
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil || s == "" {
		return time.Time{}, false
	}
	ts, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, false
	}
	return ts, true
}

// newestStoppedAt parses the newest stopped_at entry from an agent-stops
// record body.
func newestStoppedAt(path string) (time.Time, bool) {
	blob, err := os.ReadFile(path)
	if err != nil {
		return time.Time{}, false
	}
	var body struct {
		Entries []struct {
			StoppedAt string `json:"stopped_at"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(blob, &body); err != nil {
		return time.Time{}, false
	}
	var newest time.Time
	ok := false
	for _, e := range body.Entries {
		if e.StoppedAt == "" {
			continue
		}
		ts, err := time.Parse(time.RFC3339, e.StoppedAt)
		if err != nil {
			continue
		}
		if !ok || ts.After(newest) {
			newest, ok = ts, true
		}
	}
	return newest, ok
}

// oldestEntryDate returns the oldest datable per-entry date under the
// verify candidate (entry-level deletion re-dates each entry at action
// time; the group-level read is for the decision report).
func oldestEntryDate(moaiRoot string, c Candidate) (time.Time, bool) {
	var oldest time.Time
	ok := false
	for _, rel := range c.Paths {
		blob, err := os.ReadFile(filepath.Join(moaiRoot, filepath.FromSlash(rel)))
		if err != nil {
			continue
		}
		var body map[string]json.RawMessage
		if err := json.Unmarshal(blob, &body); err != nil {
			continue
		}
		raw, present := body["recorded_at"]
		if !present {
			continue
		}
		var ts time.Time
		if err := json.Unmarshal(raw, &ts); err != nil || ts.IsZero() {
			continue
		}
		if !ok || ts.Before(oldest) {
			oldest, ok = ts, true
		}
	}
	return oldest, ok
}

// entryDate reads one verify-scratch entry's own recorded_at (the D31
// pinned field); an entry without it is undatable ⇒ spared.
func entryDate(moaiRoot, rel string) (time.Time, bool) {
	blob, err := os.ReadFile(filepath.Join(moaiRoot, filepath.FromSlash(rel)))
	if err != nil {
		return time.Time{}, false
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(blob, &body); err != nil {
		return time.Time{}, false
	}
	raw, present := body["recorded_at"]
	if !present {
		return time.Time{}, false
	}
	var ts time.Time
	if err := json.Unmarshal(raw, &ts); err != nil || ts.IsZero() {
		return time.Time{}, false
	}
	return ts, true
}

// trimSuffixKey splits a trailing suffix and reports the stem.
func trimSuffixKey(name, suffix string) (string, bool) {
	if !strings.HasSuffix(name, suffix) || name == suffix {
		return "", false
	}
	return strings.TrimSuffix(name, suffix), true
}

// trimPrefixKey splits a leading prefix and a trailing suffix.
func trimPrefixKey(name, prefix, suffix string) (string, bool) {
	if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, suffix) {
		return "", false
	}
	return strings.TrimSuffix(strings.TrimPrefix(name, prefix), suffix), true
}

// safeReadDir lists a directory, treating absence as empty.
func safeReadDir(dir string) []os.DirEntry {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	return entries
}

// sortRelPaths orders relative paths deterministically.
func sortRelPaths(paths []string) {
	for i := 1; i < len(paths); i++ {
		for j := i; j > 0 && paths[j] < paths[j-1]; j-- {
			paths[j], paths[j-1] = paths[j-1], paths[j]
		}
	}
}

// unresolvableIn lists class-shaped files whose name carries no session
// key (reported spared, REQ-HYG-005's out-of-scope rule).
func unresolvableIn(dir, suffix string, class TargetClass) []Candidate {
	var out []Candidate
	for _, e := range safeReadDir(dir) {
		if e.IsDir() {
			continue
		}
		if key, ok := trimSuffixKey(e.Name(), suffix); ok || suffix == "" {
			_ = key
			if sessionKeyShape.MatchString(strings.TrimSuffix(e.Name(), suffix)) {
				continue
			}
			out = append(out, Candidate{
				Class: class,
				Paths: []string{filepath.Join(filepath.Base(dir), e.Name())},
			})
		}
	}
	return out
}
