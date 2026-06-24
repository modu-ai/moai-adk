package preference

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// MaxCoreBytes is the always-loaded core-tier byte budget (NFR-ADM-002,
// AC-ADM-NFR-002). When core.yaml would exceed it, the oldest/lowest-weight
// entry is demoted to recall until core is back under budget.
const MaxCoreBytes = 4 * 1024 // 4 KB

// coreYAMLName / recallJSONLName / archivalDirName are the canonical tier
// filenames under the user_decisions/ namespace (design.md §A.1 layout).
const (
	coreYAMLName     = "core.yaml"
	recallJSONLName  = "recall.jsonl"
	archivalDirName  = "archival"
	userDecisionsDir = "user_decisions"
)

// fileStore is the on-disk Store implementation. It owns three files under
// memDir/user_decisions/: core.yaml (the always-loaded profile), recall.jsonl
// (recent-session facts), and archival/ (full-search targets).
type fileStore struct {
	// memDir is the resolved ~/.claude/projects/{slug}/memory/ root.
	memDir string
	// udDir is memDir + "user_decisions/".
	udDir string
}

// NewFileStore creates the user_decisions/ directory layout under memDir and
// returns a ready Store. memDir is the Claude Code memory root
// (~/.claude/projects/{slug}/memory/); the store creates a user_decisions/
// subdirectory inside it.
//
// The caller is responsible for resolving memDir (the project slug + memory
// path). The hook layer (internal/hook/session_end.go resolveMemoryDir) is the
// canonical resolver; this package stays agnostic of the slug derivation.
func NewFileStore(memDir string) (Store, error) {
	if memDir == "" {
		return nil, errors.New("preference: memory dir is empty")
	}
	// Ensure the memory root itself exists (the caller resolves it, but a fresh
	// machine where ~/.claude/projects/{slug}/memory/ has never been written to
	// must still succeed on first capture).
	if err := os.MkdirAll(memDir, 0o755); err != nil {
		return nil, fmt.Errorf("preference: create memory dir: %w", err)
	}
	udDir := filepath.Join(memDir, userDecisionsDir)
	if err := os.MkdirAll(filepath.Join(udDir, archivalDirName), 0o755); err != nil {
		return nil, fmt.Errorf("preference: create user_decisions layout: %w", err)
	}
	// Touch the tier files so callers that Stat them immediately after
	// NewFileStore (e.g. the directory-layout test) see them present.
	for _, name := range []string{coreYAMLName, recallJSONLName} {
		p := filepath.Join(udDir, name)
		if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
			if err := os.WriteFile(p, nil, 0o644); err != nil {
				return nil, fmt.Errorf("preference: touch %s: %w", name, err)
			}
		}
	}
	return &fileStore{memDir: memDir, udDir: udDir}, nil
}

// Upsert implements Store. It validates the entry, routes stable entries to
// core (subject to the 4KB cap) and transient entries to recall, and writes
// atomically via temp-file-then-rename.
func (s *fileStore) Upsert(domain, decisionKey string, entry Entry) error {
	if domain == "" {
		return fmt.Errorf("%w: domain is empty", ErrInvalidEntry)
	}
	if decisionKey == "" {
		return fmt.Errorf("%w: decision_key is empty", ErrInvalidEntry)
	}
	// Normalize the entry's key fields to the call args so a deserialized entry
	// with a drifted Domain/DecisionKey cannot bypass the upsert contract.
	entry.Domain = domain
	entry.DecisionKey = decisionKey
	if err := entry.Validate(); err != nil {
		return err
	}

	// Always remove any prior copy from recall first so the replace semantic
	// is honored across tiers: a key promoted to core must not also linger in
	// recall (which would make Query double-count it).
	if err := s.removeFromRecall(domain, decisionKey); err != nil {
		return fmt.Errorf("preference: upsert (recall sweep): %w", err)
	}

	// Route by scope. Stable → core (with demotion on overflow); transient →
	// recall. The M4 decay policy will later move transient entries between
	// recall and archival based on age; M1 only needs the cascade to resolve.
	if entry.Scope == ScopeStable {
		if err := s.upsertToCore(domain, decisionKey, entry); err != nil {
			return fmt.Errorf("preference: upsert to core: %w", err)
		}
		return nil
	}
	if err := s.upsertToRecall(domain, decisionKey, entry); err != nil {
		return fmt.Errorf("preference: upsert to recall: %w", err)
	}
	return nil
}

// Get implements the 3-tier cascade (REQ-ADM-004). Core is consulted first; on
// a hit, recall and archival are NOT touched.
func (s *fileStore) Get(domain, decisionKey string) (Entry, bool, Tier, error) {
	if entry, ok, err := s.getFromCore(domain, decisionKey); err != nil {
		return Entry{}, false, TierNone, fmt.Errorf("preference: core read: %w", err)
	} else if ok {
		return entry, true, TierCore, nil
	}

	if entry, ok, err := s.getFromRecall(domain, decisionKey); err != nil {
		return Entry{}, false, TierNone, fmt.Errorf("preference: recall read: %w", err)
	} else if ok {
		return entry, true, TierRecall, nil
	}

	if entry, ok, err := s.getFromArchival(domain, decisionKey); err != nil {
		return Entry{}, false, TierNone, fmt.Errorf("preference: archival read: %w", err)
	} else if ok {
		return entry, true, TierArchival, nil
	}

	return Entry{}, false, TierNone, nil
}

// Query implements Store. It unions entries from all tiers whose Domain field
// matches. The cascade order in the result is core, then recall, then archival;
// core/recall duplicate-suppression ensures an entry demoted from core to
// recall does not appear twice (the demotion removes it from core).
func (s *fileStore) Query(domain string) ([]Entry, error) {
	if domain == "" {
		return nil, nil
	}
	seen := make(map[string]bool) // dedupe by decision_key across tiers
	var out []Entry

	coreEntries, err := s.loadCore()
	if err != nil {
		return nil, fmt.Errorf("preference: query core: %w", err)
	}
	for _, e := range coreEntries {
		if e.Domain == domain && !seen[e.DecisionKey] {
			out = append(out, e)
			seen[e.DecisionKey] = true
		}
	}

	recallEntries, err := s.loadRecall()
	if err != nil {
		return nil, fmt.Errorf("preference: query recall: %w", err)
	}
	for _, e := range recallEntries {
		if e.Domain == domain && !seen[e.DecisionKey] {
			out = append(out, e)
			seen[e.DecisionKey] = true
		}
	}

	archivalEntries, err := s.loadArchival()
	if err != nil {
		return nil, fmt.Errorf("preference: query archival: %w", err)
	}
	for _, e := range archivalEntries {
		if e.Domain == domain && !seen[e.DecisionKey] {
			out = append(out, e)
			seen[e.DecisionKey] = true
		}
	}
	return out, nil
}

// ---- core tier (core.yaml) ----

// coreFile is the on-disk shape of core.yaml.
type coreFile struct {
	Entries []Entry `yaml:"entries"`
}

// upsertToCore replaces (or inserts) the entry in core.yaml, then demotes
// lowest-weight stable entries to recall until core.yaml fits MaxCoreBytes.
func (s *fileStore) upsertToCore(domain, decisionKey string, entry Entry) error {
	cf, err := s.loadCoreFile()
	if err != nil {
		return err
	}

	// Replace-in-place if the key exists; otherwise append.
	replaced := false
	for i, e := range cf.Entries {
		if e.Domain == domain && e.DecisionKey == decisionKey {
			cf.Entries[i] = entry
			replaced = true
			break
		}
	}
	if !replaced {
		cf.Entries = append(cf.Entries, entry)
	}

	// Enforce the 4KB cap by demoting lowest-weight entries to recall until the
	// serialized core.yaml fits. Demotion uses the entry's Weight as the primary
	// sort key and ValidTime (oldest first) as the tie-breaker.
	for {
		data, mErr := yaml.Marshal(cf)
		if mErr != nil {
			return fmt.Errorf("marshal core.yaml: %w", mErr)
		}
		if len(data) <= MaxCoreBytes || len(cf.Entries) == 0 {
			break
		}
		// Pick the demotion candidate: lowest weight, then oldest ValidTime.
		demoteIdx := 0
		for i := 1; i < len(cf.Entries); i++ {
			if lessWeight(cf.Entries[i], cf.Entries[demoteIdx]) {
				demoteIdx = i
			}
		}
		demoted := cf.Entries[demoteIdx]
		cf.Entries = append(cf.Entries[:demoteIdx], cf.Entries[demoteIdx+1:]...)
		if err := s.upsertToRecall(demoted.Domain, demoted.DecisionKey, demoted); err != nil {
			return fmt.Errorf("demote to recall: %w", err)
		}
	}

	return s.writeCoreFile(cf)
}

// lessWeight returns true if a should be demoted before b. Primary key: Weight
// (lower first). Tie-breaker: ValidTime (older first). Final tie-breaker:
// DecisionKey lexicographic (deterministic regardless of map iteration order).
func lessWeight(a, b Entry) bool {
	if a.Weight != b.Weight {
		return a.Weight < b.Weight
	}
	if !a.ValidTime.Equal(b.ValidTime) {
		return a.ValidTime.Before(b.ValidTime)
	}
	return a.DecisionKey < b.DecisionKey
}

func (s *fileStore) getFromCore(domain, decisionKey string) (Entry, bool, error) {
	cf, err := s.loadCoreFile()
	if err != nil {
		return Entry{}, false, err
	}
	for _, e := range cf.Entries {
		if e.Domain == domain && e.DecisionKey == decisionKey {
			return e, true, nil
		}
	}
	return Entry{}, false, nil
}

func (s *fileStore) loadCore() ([]Entry, error) {
	cf, err := s.loadCoreFile()
	if err != nil {
		return nil, err
	}
	return cf.Entries, nil
}

func (s *fileStore) loadCoreFile() (coreFile, error) {
	p := filepath.Join(s.udDir, coreYAMLName)
	data, err := os.ReadFile(p)
	if err != nil {
		return coreFile{}, fmt.Errorf("read core.yaml: %w", err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return coreFile{}, nil
	}
	var cf coreFile
	if err := yaml.Unmarshal(data, &cf); err != nil {
		return coreFile{}, fmt.Errorf("parse core.yaml: %w", err)
	}
	return cf, nil
}

func (s *fileStore) writeCoreFile(cf coreFile) error {
	data, err := yaml.Marshal(cf)
	if err != nil {
		return fmt.Errorf("marshal core.yaml: %w", err)
	}
	return atomicWrite(filepath.Join(s.udDir, coreYAMLName), data)
}

// ---- recall tier (recall.jsonl) ----

// upsertToRecall replaces (or appends) the entry in recall.jsonl. The file is
// JSON Lines: one Entry per line, keyed by (Domain, DecisionKey).
func (s *fileStore) upsertToRecall(domain, decisionKey string, entry Entry) error {
	entries, err := s.loadRecall()
	if err != nil {
		return err
	}
	replaced := false
	for i, e := range entries {
		if e.Domain == domain && e.DecisionKey == decisionKey {
			entries[i] = entry
			replaced = true
			break
		}
	}
	if !replaced {
		entries = append(entries, entry)
	}
	return s.writeRecall(entries)
}

func (s *fileStore) removeFromRecall(domain, decisionKey string) error {
	entries, err := s.loadRecall()
	if err != nil {
		return err
	}
	out := entries[:0]
	removed := false
	for _, e := range entries {
		if e.Domain == domain && e.DecisionKey == decisionKey {
			removed = true
			continue
		}
		out = append(out, e)
	}
	if !removed {
		return nil
	}
	return s.writeRecall(out)
}

func (s *fileStore) getFromRecall(domain, decisionKey string) (Entry, bool, error) {
	entries, err := s.loadRecall()
	if err != nil {
		return Entry{}, false, err
	}
	// Last-writer-wins: scan in reverse so the most recent upsert wins on the
	// (very unlikely) case of duplicate keys in the file.
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		if e.Domain == domain && e.DecisionKey == decisionKey {
			return e, true, nil
		}
	}
	return Entry{}, false, nil
}

func (s *fileStore) loadRecall() ([]Entry, error) {
	p := filepath.Join(s.udDir, recallJSONLName)
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("read recall.jsonl: %w", err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, nil
	}
	var out []Entry
	dec := json.NewDecoder(bytes.NewReader(data))
	for dec.More() {
		var e Entry
		if err := dec.Decode(&e); err != nil {
			return nil, fmt.Errorf("parse recall.jsonl line: %w", err)
		}
		out = append(out, e)
	}
	return out, nil
}

func (s *fileStore) writeRecall(entries []Entry) error {
	// Sort for deterministic output (does not change semantics; Query dedupes
	// by key). This keeps recall.jsonl diffs reviewable.
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Domain != entries[j].Domain {
			return entries[i].Domain < entries[j].Domain
		}
		return entries[i].DecisionKey < entries[j].DecisionKey
	})
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for _, e := range entries {
		if err := enc.Encode(&e); err != nil {
			return fmt.Errorf("encode recall entry: %w", err)
		}
	}
	return atomicWrite(filepath.Join(s.udDir, recallJSONLName), buf.Bytes())
}

// ---- archival tier (archival/) ----

// archivalEntryPath returns the file path for an archival entry. Archival is
// one JSON file per (domain, decision_key) pair under archival/.
func (s *fileStore) archivalEntryPath(domain, decisionKey string) string {
	safe := slugify(domain) + "__" + slugify(decisionKey) + ".json"
	return filepath.Join(s.udDir, archivalDirName, safe)
}

// writeArchivalEntry persists a single entry to the archival tier. Archival is
// write-once-per-key in M1 (the M4 decay scan is what promotes recall entries
// to archival); this helper exists so the cascade test can seed the archival
// tier directly without going through Upsert (which routes by Scope).
func (s *fileStore) writeArchivalEntry(domain, decisionKey string, entry Entry) error {
	entry.Domain = domain
	entry.DecisionKey = decisionKey
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal archival entry: %w", err)
	}
	return atomicWrite(s.archivalEntryPath(domain, decisionKey), data)
}

func (s *fileStore) getFromArchival(domain, decisionKey string) (Entry, bool, error) {
	p := s.archivalEntryPath(domain, decisionKey)
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Entry{}, false, nil
		}
		return Entry{}, false, fmt.Errorf("read archival: %w", err)
	}
	var e Entry
	if err := json.Unmarshal(data, &e); err != nil {
		return Entry{}, false, fmt.Errorf("parse archival: %w", err)
	}
	return e, true, nil
}

func (s *fileStore) loadArchival() ([]Entry, error) {
	dir := filepath.Join(s.udDir, archivalDirName)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read archival dir: %w", err)
	}
	var out []Entry
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".json") {
			continue
		}
		data, rErr := os.ReadFile(filepath.Join(dir, ent.Name()))
		if rErr != nil {
			return nil, fmt.Errorf("read archival %s: %w", ent.Name(), rErr)
		}
		var e Entry
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, fmt.Errorf("parse archival %s: %w", ent.Name(), err)
		}
		out = append(out, e)
	}
	return out, nil
}

// ---- shared helpers ----

// atomicWrite writes data to path via a temp file in the same directory
// followed by os.Rename. This guarantees the target path is never observed in
// a half-written state, even under SIGKILL mid-write (AC-ADM-001 edge case,
// design.md §C/KI-3). The temp file is cleaned up on any error path.
//
// Reference: internal/update/updater.go uses the same temp-in-same-dir +
// os.Rename pattern for binary downloads.
func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir for atomic write: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".pref-write-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpName) }

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		cleanup()
		return fmt.Errorf("rename temp → final: %w", err)
	}
	return nil
}

// slugify turns a domain or decision_key into a filesystem-safe slug for the
// archival tier filename. It keeps alphanumerics and dashes; everything else
// becomes a dash. Multiple consecutive dashes collapse to one.
func slugify(s string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash && b.Len() > 0 {
				b.WriteRune('-')
				prevDash = true
			}
		}
	}
	out := strings.TrimRight(b.String(), "-")
	if out == "" {
		return "x"
	}
	return out
}
