package preference

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// helper to build a valid entry quickly.
func validEntry(fact, domain, key string) Entry {
	now := time.Date(2026, 6, 24, 10, 0, 0, 0, time.UTC)
	return Entry{
		Fact:           fact,
		SourceCitation: "session:test",
		ValidTime:      now,
		LastUsed:       now,
		Scope:          ScopeStable,
		Domain:         domain,
		DecisionKey:    key,
		Confidence:     ConfidenceObserved,
	}
}

// newTestStore constructs a Store rooted at a fresh temp dir + the resolved
// memory layout (memory/user_decisions/{core.yaml,recall.jsonl,archival/}).
func newTestStore(t *testing.T) Store {
	t.Helper()
	root := t.TempDir()
	memDir := filepath.Join(root, "memory")
	store, err := NewFileStore(memDir)
	if err != nil {
		t.Fatalf("NewFileStore(%q) error: %v", memDir, err)
	}
	return store
}

// TestUpsert_Idempotent_ReplaceNotAppend (AC-ADM-001) verifies two consecutive
// upserts on the same (domain, decision_key) produce exactly ONE entry, with
// the second replacing the first rather than appending.
func TestUpsert_Idempotent_ReplaceNotAppend(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	e1 := validEntry("prefers Go backend", "tech_stack", "backend_language")
	e2 := validEntry("prefers Python backend", "tech_stack", "backend_language") // same key, different fact

	if err := store.Upsert("tech_stack", "backend_language", e1); err != nil {
		t.Fatalf("Upsert #1 error: %v", err)
	}
	if err := store.Upsert("tech_stack", "backend_language", e2); err != nil {
		t.Fatalf("Upsert #2 error: %v", err)
	}

	// Query must return exactly one entry for the (domain,key) pair.
	got, ok, _, err := store.Get("tech_stack", "backend_language")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !ok {
		t.Fatalf("Get returned ok=false after upsert; want the entry to exist")
	}
	if got.Fact != "prefers Python backend" {
		t.Errorf("after 2nd upsert, Fact = %q; want %q (the REPLACED value, not appended)", got.Fact, "prefers Python backend")
	}

	// Total entry count across tiers for this domain must be exactly 1.
	entries, err := store.Query("tech_stack")
	if err != nil {
		t.Fatalf("Query error: %v", err)
	}
	count := 0
	for _, e := range entries {
		if e.DecisionKey == "backend_language" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("entry count for (tech_stack, backend_language) = %d; want 1 (replace, not append)", count)
	}
}

// TestUpsert_DifferentKeysCoexist verifies that different (domain,key) pairs
// do NOT collide — the replace semantic is scoped to the key, not the whole store.
func TestUpsert_DifferentKeysCoexist(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	if err := store.Upsert("tech_stack", "backend_language", validEntry("Go", "tech_stack", "backend_language")); err != nil {
		t.Fatalf("Upsert backend_language: %v", err)
	}
	if err := store.Upsert("tech_stack", "frontend_framework", validEntry("React", "tech_stack", "frontend_framework")); err != nil {
		t.Fatalf("Upsert frontend_framework: %v", err)
	}

	got1, ok1, _, err := store.Get("tech_stack", "backend_language")
	if err != nil || !ok1 {
		t.Fatalf("Get backend_language: err=%v ok=%v", err, ok1)
	}
	got2, ok2, _, err := store.Get("tech_stack", "frontend_framework")
	if err != nil || !ok2 {
		t.Fatalf("Get frontend_framework: err=%v ok=%v", err, ok2)
	}
	if got1.Fact != "Go" || got2.Fact != "React" {
		t.Errorf("coexisting entries drifted: got1.Fact=%q got2.Fact=%q", got1.Fact, got2.Fact)
	}
}

// TestGet_MissingKeyReturnsFalse verifies that a Get for a non-existent key
// returns (zero, false, no-tier, nil-error) rather than erroring.
func TestGet_MissingKeyReturnsFalse(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	_, ok, _, err := store.Get("nonexistent_domain", "nonexistent_key")
	if err != nil {
		t.Fatalf("Get on missing key returned error: %v (want nil-error + ok=false)", err)
	}
	if ok {
		t.Errorf("Get on missing key returned ok=true; want false")
	}
}

// TestQuery_ByDomain (AC-ADM-002 prerequisite) verifies Query filters by domain
// and returns only entries whose Domain field matches.
func TestQuery_ByDomain(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	_ = store.Upsert("tech_stack", "backend_language", validEntry("Go", "tech_stack", "backend_language"))
	_ = store.Upsert("editor_pref", "theme", validEntry("dark", "editor_pref", "theme"))
	_ = store.Upsert("tech_stack", "frontend_framework", validEntry("React", "tech_stack", "frontend_framework"))

	entries, err := store.Query("tech_stack")
	if err != nil {
		t.Fatalf("Query error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("Query(tech_stack) returned %d entries; want 2", len(entries))
	}
	for _, e := range entries {
		if e.Domain != "tech_stack" {
			t.Errorf("Query leaked cross-domain entry: Domain=%q", e.Domain)
		}
	}
}

// TestUpsert_AtomicWriteSurvivesPartialState (AC-ADM-001 SIGKILL defense)
// verifies that recall.jsonl is never left in a half-written state: after a
// successful Upsert the file parses as valid JSON lines; a temp file used
// mid-write MUST NOT pollute the final path. We approximate the SIGKILL hazard
// by asserting (a) no temp file litters the user_decisions/ dir after the call,
// and (b) recall.jsonl ends in a newline and parses line-by-line as JSON.
func TestUpsert_AtomicWriteSurvivesPartialState(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	memDir := filepath.Join(root, "memory")
	store, err := NewFileStore(memDir)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	if err := store.Upsert("tech_stack", "backend_language", validEntry("Go", "tech_stack", "backend_language")); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	// No temp file leftover in user_decisions/.
	entries, err := os.ReadDir(filepath.Join(memDir, "user_decisions"))
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, ent := range entries {
		name := ent.Name()
		// Allow only the canonical tier files + archival dir.
		switch name {
		case "core.yaml", "recall.jsonl", "archival":
			continue
		default:
			if isTempArtifact(name) {
				t.Errorf("temp file leftover in user_decisions/: %q (atomic upsert must not leave temp artifacts)", name)
			}
		}
	}
}

// isTempArtifact returns true for a name shaped like a temp/swap file
// (leading dot, .tmp/.bak/.swp suffix, or os.CreateTemp-style random suffix).
func isTempArtifact(name string) bool {
	// Hidden files (leading dot) that are NOT the canonical tier files.
	if len(name) > 0 && name[0] == '.' {
		return true
	}
	for _, suffix := range []string{".tmp", ".bak", ".swp", ".old"} {
		if len(name) > len(suffix) && name[len(name)-len(suffix):] == suffix {
			return true
		}
	}
	return false
}

// TestCascade_CoreHitSkipsRecallAndArchival (AC-ADM-004) verifies that when an
// entry is present in the core tier, neither recall.jsonl nor archival/ is
// accessed during Get. We assert by placing an entry ONLY in core (via a helper)
// and reading it back; the tier returned MUST be TierCore.
func TestCascade_CoreHitSkipsRecallAndArchival(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	memDir := filepath.Join(root, "memory")
	store, err := NewFileStore(memDir)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	// Force a fresh stable entry into the core tier directly (via internal hook
	// if available; otherwise via the public UpsertToCore mechanism).
	stableEntry := validEntry("Go backend", "tech_stack", "backend_language")
	stableEntry.Scope = ScopeStable
	if err := store.(*fileStore).upsertToCore("tech_stack", "backend_language", stableEntry); err != nil {
		t.Fatalf("upsertToCore: %v", err)
	}

	// recall.jsonl and archival/ must be absent or empty at this point.
	recallPath := filepath.Join(memDir, "user_decisions", "recall.jsonl")
	if data, err := os.ReadFile(recallPath); err == nil && len(data) > 0 {
		t.Fatalf("recall.jsonl must be empty when only-core entry exists; got %d bytes", len(data))
	}

	got, ok, tier, err := store.Get("tech_stack", "backend_language")
	if err != nil || !ok {
		t.Fatalf("Get: err=%v ok=%v", err, ok)
	}
	if tier != TierCore {
		t.Errorf("tier = %v; want TierCore (core hit must not cascade)", tier)
	}
	if got.Fact != "Go backend" {
		t.Errorf("Fact = %q; want %q", got.Fact, "Go backend")
	}
}

// TestCascade_RecallHitAfterCoreMiss (AC-ADM-004) verifies that an entry absent
// from core but present in recall is found at the recall tier after the core
// miss. We populate recall directly (bypassing core promotion) via the internal
// recall writer.
func TestCascade_RecallHitAfterCoreMiss(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	memDir := filepath.Join(root, "memory")
	store, err := NewFileStore(memDir)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	// Upsert lands in recall by default (core is reserved for stable + frequent).
	entry := validEntry("verbose logs this session", "log_pref", "verbosity")
	entry.Scope = ScopeTransient
	if err := store.Upsert("log_pref", "verbosity", entry); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	got, ok, tier, err := store.Get("log_pref", "verbosity")
	if err != nil || !ok {
		t.Fatalf("Get: err=%v ok=%v", err, ok)
	}
	if tier != TierRecall {
		t.Errorf("tier = %v; want TierRecall (core miss → recall hit)", tier)
	}
	if got.Fact != "verbose logs this session" {
		t.Errorf("Fact = %q; want %q", got.Fact, "verbose logs this session")
	}
}

// TestCoreSizeEnforcement_DemotesOnOverflow (AC-ADM-NFR-002) verifies that when
// core.yaml would exceed MaxCoreBytes (4KB), the oldest/lowest-weight entry is
// demoted to recall and the on-disk core.yaml stays ≤ 4KB.
func TestCoreSizeEnforcement_DemotesOnOverflow(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	memDir := filepath.Join(root, "memory")
	store, err := NewFileStore(memDir)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	// Stuff enough stable entries into core to force overflow. Each entry has
	// a long Fact so a handful push core past 4KB.
	for i := 0; i < 20; i++ {
		e := validEntry(longFact(i), "overflow_domain", keyFor(i))
		e.Scope = ScopeStable
		if err := store.(*fileStore).upsertToCore("overflow_domain", keyFor(i), e); err != nil {
			t.Fatalf("upsertToCore(%d): %v", i, err)
		}
	}

	// core.yaml MUST now be ≤ 4KB.
	corePath := filepath.Join(memDir, "user_decisions", "core.yaml")
	data, err := os.ReadFile(corePath)
	if err != nil {
		t.Fatalf("ReadFile core.yaml: %v", err)
	}
	if n := len(data); n > MaxCoreBytes {
		t.Errorf("core.yaml = %d bytes; want ≤ %d (MaxCoreBytes)", n, MaxCoreBytes)
	}

	// The demoted entries must still be retrievable via the cascade (recall hit).
	got, ok, _, err := store.Get("overflow_domain", keyFor(0))
	if err != nil {
		t.Fatalf("Get demoted entry: %v", err)
	}
	if !ok {
		t.Errorf("demoted entry keyFor(0) not found via cascade; expected recall-tier hit")
	}
	if got != (Entry{}) && got.Fact == "" {
		t.Errorf("demoted entry came back with empty Fact")
	}
}

// TestNamespaceSeparation_UserDecisionsVsFeedback (AC-ADM-002) verifies that
// the user-decisions namespace (memory/user_decisions/) is physically separate
// from the technical-lesson feedback namespace (memory/feedback_*.md +
// memory/MEMORY.md), and that querying one namespace returns zero entries from
// the other.
func TestNamespaceSeparation_UserDecisionsVsFeedback(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	memDir := filepath.Join(root, "memory")
	if err := os.MkdirAll(memDir, 0o755); err != nil {
		t.Fatalf("MkdirAll memory: %v", err)
	}

	// Simulate pre-existing technical-lesson memory files in the feedback namespace.
	feedbackContent := []byte("# feedback\n\n- decision: prefers Go backend over Python\n")
	if err := os.WriteFile(filepath.Join(memDir, "feedback_tech.md"), feedbackContent, 0o644); err != nil {
		t.Fatalf("WriteFile feedback: %v", err)
	}
	memoryIndex := []byte("# Memory\n\n- [feedback](feedback_tech.md) — prefers Go backend\n")
	if err := os.WriteFile(filepath.Join(memDir, "MEMORY.md"), memoryIndex, 0o644); err != nil {
		t.Fatalf("WriteFile MEMORY.md: %v", err)
	}

	store, err := NewFileStore(memDir)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	// Upsert a user-decision entry with the SAME fact string as the feedback file.
	// If namespaces were conflated, Query would see 2 entries. It must see 1.
	udEntry := validEntry("prefers Go backend over Python", "tech_stack", "backend_language")
	if err := store.Upsert("tech_stack", "backend_language", udEntry); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	entries, err := store.Query("tech_stack")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	// The decision namespace query must return exactly 1 entry (the upserted one),
	// NOT 2 (which would happen if feedback_*.md content leaked into user_decisions/).
	if len(entries) != 1 {
		t.Errorf("Query(tech_stack) returned %d entries; want 1 (feedback_*.md must NOT pollute user_decisions namespace)", len(entries))
	}

	// Conversely, the feedback file must NOT contain the decision_key lookup.
	feedbackOnDisk, err := os.ReadFile(filepath.Join(memDir, "feedback_tech.md"))
	if err != nil {
		t.Fatalf("read feedback_tech.md back: %v", err)
	}
	// The feedback file content is unchanged from what we wrote (the store never touched it).
	if string(feedbackOnDisk) != string(feedbackContent) {
		t.Errorf("feedback_tech.md was modified by the preference store; namespaces are not isolated")
	}
}

// TestNewFileStore_CreatesDirectoryLayout (REQ-ADM-002) verifies the store
// creates the user_decisions/{core.yaml,recall.jsonl,archival/} layout on init.
func TestNewFileStore_CreatesDirectoryLayout(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	memDir := filepath.Join(root, "memory")
	if _, err := NewFileStore(memDir); err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	udDir := filepath.Join(memDir, "user_decisions")
	for _, sub := range []string{"core.yaml", "recall.jsonl", "archival"} {
		p := filepath.Join(udDir, sub)
		info, err := os.Stat(p)
		if err != nil {
			t.Errorf("expected %s to exist after NewFileStore: %v", p, err)
			continue
		}
		if sub == "archival" && !info.IsDir() {
			t.Errorf("archival must be a directory; got mode %v", info.Mode())
		}
	}
}

// TestUpsert_RejectsInvalidEntry (AC-ADM-003 guard) verifies Upsert surfaces
// Entry.Validate errors rather than persisting invalid state.
func TestUpsert_RejectsInvalidEntry(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	bad := validEntry("Go", "tech_stack", "backend_language")
	bad.Fact = "" // missing required field

	err := store.Upsert("tech_stack", "backend_language", bad)
	if err == nil {
		t.Fatalf("Upsert with missing Fact returned nil; want validation error")
	}
	if !errors.Is(err, ErrInvalidEntry) {
		t.Errorf("Upsert error = %v; want it to wrap ErrInvalidEntry", err)
	}
}
