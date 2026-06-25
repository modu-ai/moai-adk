package preference

import (
	"strings"
	"testing"
	"time"
)

// TestCorrectInferred_ObservedUpsertAndInferredDemoted verifies the M5
// correction loop (REQ-ADM-016, AC-ADM-016 [S2 Critical]). When the user
// corrects an inferred preference ("이 추론이 틀리다"), the corrected fact is
// upserted with Confidence: ConfidenceObserved AND the prior inferred entry's
// weight is reduced (demoted to archival as an audit record).
//
// Post-state (AC-ADM-016 evidence: "inferred 엔트리 weight 감소 + 새 observed
// 엔트리 존재"):
//   - A new entry for (domain, decisionKey) exists in recall with Confidence
//     ConfidenceObserved and the corrected fact.
//   - The prior inferred entry is superseded — it is NOT the entry Get returns
//     for the same key anymore (Get returns the observed one).
//   - The prior inferred entry's weight was reduced before supersession and the
//     reduced-weight copy is preserved in archival as a correction audit record.
func TestCorrectInferred_ObservedUpsertAndInferredDemoted(t *testing.T) {
	// Cannot run in parallel: mutates a real fileStore under t.TempDir().
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	memDir, err := resolveMemoryDirOverride("")
	if err != nil {
		t.Fatalf("resolveMemoryDirOverride: %v", err)
	}
	store, err := NewFileStore(memDir)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	// Seed an inferred entry with a fact the user will correct.
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	inferred := freshEntry(ScopeStable, now.Add(-2*24*time.Hour), 0.6)
	inferred.Domain = "backend_language"
	inferred.DecisionKey = "preferred_stack"
	inferred.Fact = "prefers Python backend (inferred)"
	inferred.Confidence = ConfidenceInferred
	inferred.SourceCitation = "inference:pattern-match session=abc"
	if err := store.Upsert(inferred.Domain, inferred.DecisionKey, inferred); err != nil {
		t.Fatalf("seed inferred: %v", err)
	}

	// User corrects: "the inference is wrong — I prefer Go".
	correctedFact := "prefers Go backend"
	if err := CorrectInferred(store, inferred.Domain, inferred.DecisionKey, correctedFact, now); err != nil {
		t.Fatalf("CorrectInferred: %v", err)
	}

	// (1) New observed entry exists for the same key.
	got, ok, tier, gErr := store.Get(inferred.Domain, inferred.DecisionKey)
	if gErr != nil {
		t.Fatalf("Get after correction: %v", gErr)
	}
	if !ok {
		t.Fatal("Get after correction: entry missing (ok=false)")
	}
	if got.Confidence != ConfidenceObserved {
		t.Errorf("post-correction confidence = %v, want ConfidenceObserved", got.Confidence)
	}
	if got.Fact != correctedFact {
		t.Errorf("post-correction fact = %q, want %q", got.Fact, correctedFact)
	}
	// Recallable tier (core or recall) — NOT archival only. The corrected
	// observed fact MUST be recallable by the cascade so the orchestrator's
	// next recommendation reflects the correction.
	if tier == TierArchival {
		t.Errorf("post-correction tier = archival; corrected observed entry must be recallable (core/recall)")
	}
	if !got.LastUsed.Equal(now) {
		t.Errorf("post-correction LastUsed = %v, want %v (correction timestamp)", got.LastUsed, now)
	}

	// (2) Source citation records correction provenance (REQ-ADM-018).
	if !strings.Contains(got.SourceCitation, "correction") {
		t.Errorf("post-correction SourceCitation = %q, must record correction provenance (REQ-ADM-018)", got.SourceCitation)
	}
}

// TestCorrectInferred_PriorInferredWeightReducedAndArchived verifies the
// weight-reduction + archival half of AC-ADM-016. The prior inferred entry's
// weight MUST be reduced before supersession, and the reduced-weight copy MUST
// be preserved in archival as an audit record (NOT hard-deleted — the audit
// trail must survive).
func TestCorrectInferred_PriorInferredWeightReducedAndArchived(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	memDir, err := resolveMemoryDirOverride("")
	if err != nil {
		t.Fatalf("resolveMemoryDirOverride: %v", err)
	}
	store, err := NewFileStore(memDir)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	fs, ok := store.(*fileStore)
	if !ok {
		t.Fatalf("store is not *fileStore: %T", store)
	}

	// Seed an inferred entry with a known weight (0.7).
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	inferred := freshEntry(ScopeStable, now.Add(-1*24*time.Hour), 0.7)
	inferred.Domain = "log_level"
	inferred.DecisionKey = "verbose"
	inferred.Fact = "prefers verbose (inferred)"
	inferred.Confidence = ConfidenceInferred
	inferred.SourceCitation = "inference:heuristic"
	if err := store.Upsert(inferred.Domain, inferred.DecisionKey, inferred); err != nil {
		t.Fatalf("seed inferred: %v", err)
	}
	priorWeight := 0.7

	// Correct.
	if err := CorrectInferred(store, inferred.Domain, inferred.DecisionKey, "prefers quiet", now); err != nil {
		t.Fatalf("CorrectInferred: %v", err)
	}

	// The prior inferred entry MUST now be in archival (audit record), with
	// reduced weight.
	archived, aOk, aErr := fs.getFromArchival(inferred.Domain, inferred.DecisionKey)
	if aErr != nil {
		t.Fatalf("getFromArchival after correction: %v", aErr)
	}
	if !aOk {
		t.Fatal("prior inferred entry was hard-deleted; must be preserved in archival as audit record (AC-ADM-016)")
	}
	if archived.Confidence != ConfidenceInferred {
		t.Errorf("archived prior entry Confidence = %v, want ConfidenceInferred (provenance preserved)", archived.Confidence)
	}
	if archived.Weight >= priorWeight {
		t.Errorf("archived prior entry weight = %v, want STRICTLY less than prior %v (AC-ADM-016 weight reduction)", archived.Weight, priorWeight)
	}
	if archived.Fact != inferred.Fact {
		t.Errorf("archived prior entry Fact = %q, want original inferred %q (audit record must preserve original)", archived.Fact, inferred.Fact)
	}
}

// TestCorrectInferred_NoPriorEntryStillUpserts verifies the correction path
// when there is NO prior inferred entry — the corrected fact is still upserted
// as observed. This is the "user volunteers a fact without correcting an
// inference" case; CorrectInferred treats it as a clean observed upsert.
func TestCorrectInferred_NoPriorEntryStillUpserts(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	memDir, err := resolveMemoryDirOverride("")
	if err != nil {
		t.Fatalf("resolveMemoryDirOverride: %v", err)
	}
	store, err := NewFileStore(memDir)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	err = CorrectInferred(store, "editor", "theme", "dark theme", now)
	if err != nil {
		t.Fatalf("CorrectInferred (no prior): %v", err)
	}
	got, ok, _, gErr := store.Get("editor", "theme")
	if gErr != nil {
		t.Fatalf("Get: %v", gErr)
	}
	if !ok {
		t.Fatal("CorrectInferred (no prior) did not persist an entry")
	}
	if got.Confidence != ConfidenceObserved {
		t.Errorf("CorrectInferred (no prior) Confidence = %v, want ConfidenceObserved", got.Confidence)
	}
}

// TestCorrectInferred_ObservedPriorIsIdempotent verifies that correcting a key
// whose prior entry is ALREADY observed does not demote it (there is no
// inference to correct — the user is just re-asserting). The corrected fact
// replaces the prior observed entry; no audit-record archival write happens
// because the prior was not inferred.
func TestCorrectInferred_ObservedPriorIsIdempotent(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	memDir, err := resolveMemoryDirOverride("")
	if err != nil {
		t.Fatalf("resolveMemoryDirOverride: %v", err)
	}
	store, err := NewFileStore(memDir)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	fs, ok := store.(*fileStore)
	if !ok {
		t.Fatalf("store is not *fileStore: %T", store)
	}

	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	prior := freshEntry(ScopeStable, now.Add(-1*24*time.Hour), 0.9)
	prior.Domain = "ci"
	prior.DecisionKey = "strategy"
	prior.Fact = "prefers squash (observed)"
	prior.Confidence = ConfidenceObserved
	prior.SourceCitation = "session=abc tool=AskUserQuestion"
	if err := store.Upsert(prior.Domain, prior.DecisionKey, prior); err != nil {
		t.Fatalf("seed observed: %v", err)
	}

	if err := CorrectInferred(store, prior.Domain, prior.DecisionKey, "prefers rebase", now); err != nil {
		t.Fatalf("CorrectInferred: %v", err)
	}
	// Get returns the corrected observed entry.
	got, ok, _, gErr := store.Get(prior.Domain, prior.DecisionKey)
	if gErr != nil || !ok {
		t.Fatalf("Get: ok=%v err=%v", ok, gErr)
	}
	if got.Confidence != ConfidenceObserved {
		t.Errorf("post-correction Confidence = %v, want ConfidenceObserved", got.Confidence)
	}
	// No audit-record archival entry for an observed→observed transition (there
	// was no inference to demote).
	if archived, aOk, _ := fs.getFromArchival(prior.Domain, prior.DecisionKey); aOk {
		t.Errorf("observed→observed correction archived a record %v; no inference to demote, expected no archival audit", archived)
	}
}

// TestCorrectInferred_EmptyArgsRejected verifies the validation gate.
// Cannot use t.Parallel on the parent because t.Setenv is incompatible with
// parallel tests; the subtests themselves are safe (they do not touch env).
func TestCorrectInferred_EmptyArgsRejected(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	memDir, err := resolveMemoryDirOverride("")
	if err != nil {
		t.Fatalf("resolveMemoryDirOverride: %v", err)
	}
	store, err := NewFileStore(memDir)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name, domain, key, fact string
	}{
		{"empty domain", "", "k", "fact"},
		{"empty key", "d", "", "fact"},
		{"empty fact", "d", "k", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := CorrectInferred(store, tc.domain, tc.key, tc.fact, now)
			if err == nil {
				t.Errorf("CorrectInferred(%q,%q,%q) returned nil; want validation error", tc.domain, tc.key, tc.fact)
			}
		})
	}
}

// TestCorrectInferred_NilStoreRejected verifies the nil-store defensive path.
func TestCorrectInferred_NilStoreRejected(t *testing.T) {
	t.Parallel()
	err := CorrectInferred(nil, "d", "k", "fact", time.Now())
	if err == nil {
		t.Error("CorrectInferred(nil store) returned nil; want error")
	}
}

// TestCorrectInferred_NonFileStoreSkipsArchival verifies the branch where the
// store is NOT a *fileStore (e.g. a future in-memory test fake): the demotion
// step is skipped, but the observed upsert still runs. This guards the
// type-assertion fallback.
func TestCorrectInferred_NonFileStoreSkipsArchival(t *testing.T) {
	t.Parallel()
	store := &fakeStore{entries: map[string]Entry{}}
	// Seed an inferred entry via the fake's own map (bypass Upsert).
	key := "d/k"
	store.entries[key] = Entry{
		Fact: "inferred", Domain: "d", DecisionKey: "k",
		Confidence: ConfidenceInferred, Scope: ScopeStable,
	}
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	if err := CorrectInferred(store, "d", "k", "corrected", now); err != nil {
		t.Fatalf("CorrectInferred (fake store): %v", err)
	}
	got := store.entries[key]
	if got.Confidence != ConfidenceObserved {
		t.Errorf("fake-store post-correction Confidence = %v, want ConfidenceObserved", got.Confidence)
	}
	if got.Fact != "corrected" {
		t.Errorf("fake-store post-correction Fact = %q, want 'corrected'", got.Fact)
	}
}

// fakeStore is a minimal Store implementation that is NOT a *fileStore. It is
// used to exercise the CorrectInferred type-assertion fallback (non-fileStore
// path skips the archival demotion). It does NOT implement the archival tier;
// Get always returns the in-memory map entry with TierRecall.
type fakeStore struct {
	entries map[string]Entry
}

func (f *fakeStore) Upsert(domain, decisionKey string, entry Entry) error {
	f.entries[domain+"/"+decisionKey] = entry
	return nil
}

func (f *fakeStore) Get(domain, decisionKey string) (Entry, bool, Tier, error) {
	e, ok := f.entries[domain+"/"+decisionKey]
	if !ok {
		return Entry{}, false, TierNone, nil
	}
	return e, true, TierRecall, nil
}

func (f *fakeStore) Query(domain string) ([]Entry, error) {
	var out []Entry
	for _, e := range f.entries {
		if e.Domain == domain {
			out = append(out, e)
		}
	}
	return out, nil
}
