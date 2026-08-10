package delegationmap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness"
	"github.com/modu-ai/moai-adk/internal/harness/proposalgen"
)

// TestPatternKey_NamespaceIsolated is AC-HLA-011.
//
// The rejection is tested by feeding each emitted key to the REAL mapper as a
// maximally-actionable promotion (actionable tier, confidence above the
// threshold) and requiring zero candidates. Reconstructing the mapper's regex
// in the test would assert against a copy; running the mapper asserts against
// the gate that actually exists.
func TestPatternKey_NamespaceIsolated(t *testing.T) {
	t.Parallel()

	res, err := Analyze(opts("two_kinds.jsonl"))
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	candidates := BuildCandidates(res)
	if len(candidates) == 0 {
		t.Fatal("expected candidates from the two_kinds fixture")
	}

	for _, c := range candidates {
		if !strings.HasPrefix(c.PatternKey, "delegation_map:") {
			t.Errorf("pattern key %q is outside the reserved namespace", c.PatternKey)
		}
		promoted := proposalgen.MapPromotions([]harness.Promotion{{
			PatternKey:       c.PatternKey,
			ObservationCount: 100,
			Confidence:       1.0,
			ToTier:           harness.TierAutoUpdate.String(),
			Ts:               time.Now().UTC(),
		}})
		if len(promoted) != 0 {
			t.Errorf("the existing mapper accepted %q; the namespace is not isolated", c.PatternKey)
		}
	}

	// The event-type SSOT is untouched: delegation_map is not a member, so the
	// mapper's format regex rejects it by construction rather than by a
	// hand-maintained exclusion that could be edited away.
	for _, et := range harness.PatternBearingEventTypes() {
		if string(et) == "delegation_map" {
			t.Error("delegation_map was added to PatternBearingEventTypes(); REQ-HLA-011 forbids widening the SSOT")
		}
	}
}

// TestProposal_CarriesEvidence is AC-HLA-012.
func TestProposal_CarriesEvidence(t *testing.T) {
	t.Parallel()

	outDir := t.TempDir()
	res, err := Analyze(opts("two_kinds.jsonl"))
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	candidates := BuildCandidates(res)
	written, err := proposalgen.WriteProposals(outDir, candidates)
	if err != nil {
		t.Fatalf("WriteProposals: %v", err)
	}
	if len(written) != len(candidates) {
		t.Fatalf("wrote %d drafts, want %d", len(written), len(candidates))
	}

	for _, id := range written {
		specBody, err := os.ReadFile(filepath.Join(outDir, id, "spec.md"))
		if err != nil {
			t.Fatal(err)
		}
		propBody, err := os.ReadFile(filepath.Join(outDir, id, "proposal.json"))
		if err != nil {
			t.Fatal(err)
		}

		var prop map[string]any
		if err := json.Unmarshal(propBody, &prop); err != nil {
			t.Fatalf("proposal.json is not valid JSON: %v", err)
		}
		evidence, ok := prop["evidence"].(map[string]any)
		if !ok {
			t.Fatalf("proposal.json for %s carries no evidence block: %s", id, propBody)
		}
		for _, key := range []string{
			"kind", "observation_count", "support_ratio",
			"qualifying_rows", "unattributed_share", "subcommand", "agent",
		} {
			if _, present := evidence[key]; !present {
				t.Errorf("proposal.json for %s is missing evidence key %q", id, key)
			}
			if !strings.Contains(string(specBody), key) {
				t.Errorf("spec.md for %s is missing evidence key %q", id, key)
			}
		}

		// The body must say plainly that application is gated. A proposal that
		// reads as an instruction rather than a request is the failure mode the
		// Tier-4 gate exists to prevent.
		if !strings.Contains(string(specBody), "Tier-4") {
			t.Errorf("spec.md for %s does not state that application requires the Tier-4 approval gate", id)
		}
	}
}

// TestAnalyze_DeterministicIdempotent is AC-HLA-013.
func TestAnalyze_DeterministicIdempotent(t *testing.T) {
	t.Parallel()

	outDir := t.TempDir()

	first, err := Analyze(opts("two_kinds.jsonl"))
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	firstCandidates := BuildCandidates(first)
	if _, err := proposalgen.WriteProposals(outDir, firstCandidates); err != nil {
		t.Fatalf("WriteProposals: %v", err)
	}
	firstBytes := snapshotDir(t, outDir)

	second, err := Analyze(opts("two_kinds.jsonl"))
	if err != nil {
		t.Fatalf("Analyze (second run): %v", err)
	}
	secondCandidates := BuildCandidates(second)

	if len(firstCandidates) != len(secondCandidates) {
		t.Fatalf("candidate counts differ: %d vs %d", len(firstCandidates), len(secondCandidates))
	}
	for i := range firstCandidates {
		if !reflect.DeepEqual(firstCandidates[i], secondCandidates[i]) {
			t.Errorf("candidate %d differs between runs:\n%+v\n%+v", i, firstCandidates[i], secondCandidates[i])
		}
	}

	if _, err := proposalgen.WriteProposals(outDir, secondCandidates); err != nil {
		t.Fatalf("WriteProposals (second run): %v", err)
	}
	secondBytes := snapshotDir(t, outDir)

	if len(firstBytes) != len(secondBytes) {
		t.Fatalf("the second run changed the file set: %d vs %d files", len(firstBytes), len(secondBytes))
	}
	for path, body := range firstBytes {
		if secondBytes[path] != body {
			t.Errorf("re-running churned %s", path)
		}
	}
}

// snapshotDir reads every file under dir into a path-keyed map.
func snapshotDir(t *testing.T, dir string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		body, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			return relErr
		}
		out[rel] = string(body)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}
