package preference

import (
	"os"
	"path/filepath"
	"testing"
)

// TestTier_String covers the Tier.String diagnostic renderer including the
// TierNone default branch.
func TestTier_String(t *testing.T) {
	t.Parallel()
	cases := []struct {
		t    Tier
		want string
	}{
		{TierNone, "none"},
		{TierCore, "core"},
		{TierRecall, "recall"},
		{TierArchival, "archival"},
		{Tier(999), "none"}, // unknown falls to default
	}
	for _, c := range cases {
		if got := c.t.String(); got != c.want {
			t.Errorf("Tier(%d).String() = %q; want %q", c.t, got, c.want)
		}
	}
}

// TestArchivalTier_ReadWrite exercises the archival tier end-to-end: an entry
// written directly to archival is retrievable via Get at TierArchival and via
// Query, after both core and recall miss.
func TestArchivalTier_ReadWrite(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	store, err := NewFileStore(filepath.Join(root, "memory"))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	fs := store.(*fileStore)

	// Write directly to archival via the internal helper.
	entry := validEntry("ancient decision", "old_domain", "old_key")
	if err := fs.writeArchivalEntry("old_domain", "old_key", entry); err != nil {
		t.Fatalf("writeArchivalEntry: %v", err)
	}

	got, ok, tier, err := store.Get("old_domain", "old_key")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok {
		t.Fatalf("archival entry not found via cascade")
	}
	if tier != TierArchival {
		t.Errorf("tier = %v; want TierArchival", tier)
	}
	if got.Fact != "ancient decision" {
		t.Errorf("Fact = %q; want %q", got.Fact, "ancient decision")
	}

	entries, err := store.Query("old_domain")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Query(old_domain) = %d entries; want 1", len(entries))
	}
}

// TestArchivalTier_FileMissing exercises the os.ErrNotExist path in
// getFromArchival.
func TestArchivalTier_FileMissing(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	store, err := NewFileStore(filepath.Join(root, "memory"))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	_, ok, _, err := store.Get("never_existed", "never_existed")
	if err != nil {
		t.Errorf("Get on fully-missing key returned error: %v", err)
	}
	if ok {
		t.Errorf("Get on fully-missing key returned ok=true")
	}
}

// TestUpsert_ReplacesAcrossTiers_NoDuplicate verifies that promoting a key from
// recall (transient) to core (stable) removes the recall copy so Query does not
// double-count. This exercises the removeFromRecall path during Upsert.
func TestUpsert_ReplacesAcrossTiers_NoDuplicate(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	// First store as transient (lands in recall).
	tr := validEntry("first preference", "tech_stack", "backend_language")
	tr.Scope = ScopeTransient
	if err := store.Upsert("tech_stack", "backend_language", tr); err != nil {
		t.Fatalf("Upsert transient: %v", err)
	}

	// Now upsert as stable with the SAME key (should land in core and remove the
	// recall copy).
	st := validEntry("updated stable preference", "tech_stack", "backend_language")
	st.Scope = ScopeStable
	if err := store.Upsert("tech_stack", "backend_language", st); err != nil {
		t.Fatalf("Upsert stable: %v", err)
	}

	// Query must return exactly one entry — no recall/core duplicate.
	entries, err := store.Query("tech_stack")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Query returned %d entries for the same key across tiers; want 1 (no duplicate)", len(entries))
	}
	if len(entries) > 0 && entries[0].Fact != "updated stable preference" {
		t.Errorf("Fact = %q; want %q", entries[0].Fact, "updated stable preference")
	}
}

// TestUpsert_RecallReplaceInPlace exercises the recall-tier replace path
// (upsertToRecall with an existing key).
func TestUpsert_RecallReplaceInPlace(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	t1 := validEntry("transient v1", "log_pref", "verbosity")
	t1.Scope = ScopeTransient
	if err := store.Upsert("log_pref", "verbosity", t1); err != nil {
		t.Fatalf("Upsert v1: %v", err)
	}
	t2 := validEntry("transient v2", "log_pref", "verbosity")
	t2.Scope = ScopeTransient
	if err := store.Upsert("log_pref", "verbosity", t2); err != nil {
		t.Fatalf("Upsert v2: %v", err)
	}

	entries, err := store.Query("log_pref")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Query returned %d; want 1 (recall replace)", len(entries))
	}
	if entries[0].Fact != "transient v2" {
		t.Errorf("Fact = %q; want %q", entries[0].Fact, "transient v2")
	}
}

// TestLessWeight_TieBreakers exercises the Weight tie-breaker and the
// ValidTime secondary tie-breaker so demotion is deterministic.
func TestLessWeight_TieBreakers(t *testing.T) {
	t.Parallel()

	t0 := validEntry("a", "d", "k")
	t1 := validEntry("b", "d", "k")

	// Lower weight wins.
	t0.Weight = 0.1
	t1.Weight = 0.9
	if !lessWeight(t0, t1) {
		t.Errorf("lessWeight(lower-weight) = false; want true")
	}

	// Equal weight → older ValidTime wins.
	t0.Weight = 0.5
	t1.Weight = 0.5
	t0.ValidTime = t0.ValidTime.AddDate(0, 0, -1) // t0 older
	if !lessWeight(t0, t1) {
		t.Errorf("lessWeight(older-time) = false; want true (tie-breaker)")
	}

	// Equal weight + equal time → lexicographic key wins.
	t0.ValidTime = t1.ValidTime
	t0.DecisionKey = "aaa"
	t1.DecisionKey = "zzz"
	if !lessWeight(t0, t1) {
		t.Errorf("lessWeight(lower-key) = false; want true (final tie-breaker)")
	}
}

// TestNewFileStore_EmptyMemDir exercises the empty-memory-dir guard.
func TestNewFileStore_EmptyMemDir(t *testing.T) {
	t.Parallel()
	_, err := NewFileStore("")
	if err == nil {
		t.Fatalf("NewFileStore(\"\") returned nil; want error")
	}
}

// TestSlugify_EdgeCases exercises the slugifier's collapse + trim + empty
// fallback behaviors.
func TestSlugify_EdgeCases(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"tech_stack", "tech-stack"},
		{"tech__stack!!", "tech-stack"},
		{"  leading", "leading"},
		{"trailing  ", "trailing"},
		{"", "x"},
		{"!!!", "x"},
		{"a--b", "a-b"},
	}
	for _, c := range cases {
		if got := slugify(c.in); got != c.want {
			t.Errorf("slugify(%q) = %q; want %q", c.in, got, c.want)
		}
	}
}

// TestLoadCoreFile_ParseError exercises the malformed-YAML error path.
func TestLoadCoreFile_ParseError(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	memDir := filepath.Join(root, "memory")
	if err := os.MkdirAll(filepath.Join(memDir, userDecisionsDir), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// Write invalid YAML.
	if err := os.WriteFile(filepath.Join(memDir, userDecisionsDir, coreYAMLName), []byte("entries: [this: is: broken"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	store, err := NewFileStore(memDir)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	if _, _, _, err := store.Get("x", "y"); err == nil {
		t.Errorf("Get with malformed core.yaml returned nil; want parse error")
	}
}

// TestLoadRecall_ParseError exercises the malformed-JSONL error path.
func TestLoadRecall_ParseError(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	memDir := filepath.Join(root, "memory")
	store, err := NewFileStore(memDir)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	// Seed a valid core entry first so Get reaches the recall tier.
	coreEntry := validEntry("stable", "d", "k")
	coreEntry.Scope = ScopeStable
	if err := store.(*fileStore).upsertToCore("d", "k", coreEntry); err != nil {
		t.Fatalf("upsertToCore: %v", err)
	}
	// Now corrupt recall.jsonl.
	if err := os.WriteFile(filepath.Join(memDir, userDecisionsDir, recallJSONLName), []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	// Query must surface the recall parse error (it reads recall regardless of
	// core hit, because Query unions all tiers).
	if _, err := store.Query("d"); err == nil {
		t.Errorf("Query with malformed recall.jsonl returned nil; want parse error")
	}
}
