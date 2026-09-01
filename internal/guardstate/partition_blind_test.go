package guardstate

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/guardliveness"
)

// AC-GSM-015 clause (c) — a consumer given ONLY the result, and no knowledge of
// any value's name, can partition its entries.
//
// THIS FILE IS THE HARNESS THE CLAUSE SPEAKS OF, and its blindness is MEASURED
// rather than asserted: TestThisHarnessNamesNoValue greps this file for every
// token of the producer's vocabulary and requires no match. Do not add one —
// not in an assertion, not in a fixture name, not in a comment. A harness that
// names a value has hardcoded the literal the clause exists to make
// unnecessary, and would keep passing after the vocabulary changed underneath
// it.
//
// The mutant the clause kills: a result that designates its clean value in the
// SPEC's prose and emits nothing. The consumer is then forced either to
// hardcode the literal — which its own criteria forbid — or to read the surface
// fold, which under-reports because two classifications fold to the settled
// surface value while only one is clean.

// blindFixture builds a tree whose subjects are named for what they DID, never
// for what they will be decided to be.
func blindFixture(t *testing.T) (root string, now time.Time, q *fakeQuerier) {
	t.Helper()

	now = time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	root = t.TempDir()

	wfDir := filepath.Join(root, ".github", "workflows")
	if err := os.MkdirAll(wfDir, 0o755); err != nil {
		t.Fatalf("create the workflow directory: %v", err)
	}
	for _, name := range []string{"recent.yml", "silent.yml", "unnamed.yml"} {
		if err := os.WriteFile(filepath.Join(wfDir, name), []byte("on: push\njobs: {}\n"), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	manifest := `entries:
  - kind: github-workflow
    locator: .github/workflows/recent.yml
    events: [push]
    window: 30d
    measure: fired-at-all
  - kind: github-workflow
    locator: .github/workflows/silent.yml
    events: [push]
    window: 7d
    measure: fired-at-all
`
	manifestPath := filepath.Join(root, filepath.FromSlash(ManifestPath))
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		t.Fatalf("create the manifest directory: %v", err)
	}
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatalf("write the manifest: %v", err)
	}

	q = &fakeQuerier{histories: map[string]RunHistory{
		// Ran two hours ago, inside a thirty-day window.
		".github/workflows/recent.yml": {Runs: []Run{{Conclusion: "success", At: now.Add(-2 * time.Hour)}}},
		// No run records at all.
		".github/workflows/silent.yml": {},
	}}
	return root, now, q
}

// TestAValueBlindConsumerCanPartitionTheResult is clause (c) itself.
//
// Everything below is expressed in terms of SUBJECTS and of the result's own
// carried designation. Nothing reads a value name, and nothing reads the
// surface fold — reading the fold is the specific under-reporting the seam was
// built to prevent.
func TestAValueBlindConsumerCanPartitionTheResult(t *testing.T) {
	t.Parallel()

	root, now, q := blindFixture(t)
	producer := NewProducer(
		WithQuerierFactory(func(string) RunQuerier { return q }),
		WithClock(func() time.Time { return now }),
	)

	result, err := producer.Produce(context.Background(), guardliveness.Activation{Root: root})
	if err != nil {
		t.Fatalf("produce: %v", err)
	}

	// The designation is carried in machine-readable form, so a consumer that
	// has never heard of the vocabulary has a referent to partition against.
	if result.Clean == nil {
		t.Fatal("the result carries no designation, so a consumer must hardcode a literal to partition it")
	}
	if len(result.Clean.Values) != 1 {
		t.Fatalf("the designation names %d values; a partition has no single referent",
			len(result.Clean.Values))
	}

	fired, err := result.Partition()
	if err != nil {
		t.Fatalf("the consumer could not partition the result: %v", err)
	}

	firedOn := map[string]bool{}
	for _, e := range fired {
		firedOn[e.Subject] = true
	}

	// The subject that ran two hours ago is on the settled side.
	if firedOn[".github/workflows/recent.yml"] {
		t.Fatal("a subject that ran inside its window was partitioned as something to report")
	}
	// The subject with no runs, and the file no entry declares, are on the
	// other side.
	for _, subject := range []string{
		".github/workflows/silent.yml",
		".github/workflows/unnamed.yml",
	} {
		if !firedOn[subject] {
			t.Fatalf("%s was partitioned as nothing to report", subject)
		}
	}

	// And the same holds after the store round trip, which is the form the
	// consumer actually reads.
	store := guardliveness.NewStore(t.TempDir())
	if err := store.Save(root, result, now); err != nil {
		t.Fatalf("save: %v", err)
	}
	snapshot, err := store.Load(root)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	roundTripped, err := snapshot.Result.Partition()
	if err != nil {
		t.Fatalf("the consumer could not partition the round-tripped result: %v", err)
	}
	if len(roundTripped) != len(fired) {
		t.Fatalf("the partition changed across the round trip: %d before, %d after",
			len(fired), len(roundTripped))
	}
}
