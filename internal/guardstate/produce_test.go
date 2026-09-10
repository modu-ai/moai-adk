package guardstate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/guardliveness"
)

// ---------------------------------------------------------------------------
// The fixture tree. It is a REAL directory with REAL workflow files and a REAL
// manifest, so the manifest loader and the disk enumerator under test are the
// production ones; only the run query is faked, because it is the one that
// would reach a network.
// ---------------------------------------------------------------------------

// fixtureTree writes a manifest and its subjects into a throwaway tree and
// returns the root.
func fixtureTree(t *testing.T, manifest string, workflows ...string) string {
	t.Helper()

	root := t.TempDir()
	wfDir := filepath.Join(root, ".github", "workflows")
	if err := os.MkdirAll(wfDir, 0o755); err != nil {
		t.Fatalf("create the workflow directory: %v", err)
	}
	for _, name := range workflows {
		if err := os.WriteFile(filepath.Join(wfDir, name), []byte("on: push\njobs: {}\n"), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	manifestPath := filepath.Join(root, filepath.FromSlash(ManifestPath))
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		t.Fatalf("create the manifest directory: %v", err)
	}
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o600); err != nil {
		t.Fatalf("write the manifest: %v", err)
	}
	return root
}

// twoSubjectManifest declares two workflow subjects with different expectations,
// so a test can tell which entry an expectation was copied from.
const twoSubjectManifest = `entries:
  - kind: github-workflow
    locator: .github/workflows/stale-one.yml
    events: [push]
    window: 7d
    measure: fired-at-all
  - kind: github-workflow
    locator: .github/workflows/healthy-one.yml
    events: [push]
    window: 30d
    measure: verdict-rendered
`

// testProducer builds a producer over the fixture tree with the given fake
// query behaviour and a fixed clock.
func testProducer(q RunQuerier, now time.Time) *Producer {
	return NewProducer(
		WithQuerierFactory(func(string) RunQuerier { return q }),
		WithClock(func() time.Time { return now }),
	)
}

// entryFor finds the produced entry for a subject.
func entryFor(t *testing.T, res guardliveness.Result, subject string) guardliveness.Entry {
	t.Helper()
	for _, e := range res.Entries {
		if e.Subject == subject {
			return e
		}
	}
	t.Fatalf("no produced entry for %q; got %d entries", subject, len(res.Entries))
	return guardliveness.Entry{}
}

// ---------------------------------------------------------------------------
// AC-GSM-007 clause (d) — the row-4 entry carries the expectation it missed.
// ---------------------------------------------------------------------------

// TestProducedRow4EntryCarriesTheDeclaredExpectation is clause (d), and it is
// stated as an EQUALITY against the manifest entry's own two fields rather than
// as a presence check.
//
// The mutant clause (d) exists to kill is a classifier that emits the row-4
// classification and names nothing; the mutant clause (d)'s own WORDING exists
// to kill is a producer carrying a constant string, which satisfies any
// presence check. Both are excluded by reading the expectation out of the
// parsed manifest and comparing field-against-field.
func TestProducedRow4EntryCarriesTheDeclaredExpectation(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	root := fixtureTree(t, twoSubjectManifest, "stale-one.yml", "healthy-one.yml")

	q := &fakeQuerier{histories: map[string]RunHistory{
		// Outside its 7d window: a qualifying run, merely aged. Row 4.
		".github/workflows/stale-one.yml": {Runs: []Run{{Conclusion: "success", At: now.Add(-30 * 24 * time.Hour)}}},
		// Inside its 30d window, with a verdict. Row 3.
		".github/workflows/healthy-one.yml": {Runs: []Run{{Conclusion: "success", At: now.Add(-2 * 24 * time.Hour)}}},
	}}

	res, err := testProducer(q, now).Produce(context.Background(), guardliveness.Activation{Root: root})
	if err != nil {
		t.Fatalf("produce: %v", err)
	}

	// The comparison target is the MANIFEST, read back through the same parser
	// production uses. Comparing against string literals here would test the
	// literals in this file, not that the producer copied the entry's fields.
	manifest, err := LoadManifest(filepath.Join(root, filepath.FromSlash(ManifestPath)))
	if err != nil {
		t.Fatalf("load the manifest for comparison: %v", err)
	}
	var declared Entry
	for _, e := range manifest.Entries {
		if e.Locator == ".github/workflows/stale-one.yml" {
			declared = e
		}
	}
	if declared.Locator == "" {
		t.Fatal("the fixture manifest does not declare the row-4 subject")
	}

	got := entryFor(t, res, ".github/workflows/stale-one.yml")
	if got.Classifications[0] != string(ClassStale) {
		t.Fatalf("the fixture's row-4 case classified %q, want %q", got.Classifications[0], ClassStale)
	}
	if got.Expectation == nil {
		t.Fatal("the row-4 entry carries no expectation: it names the classification and not what was missed")
	}
	if got.Expectation.Window != declared.Window {
		t.Fatalf("carried window %q != declared window %q", got.Expectation.Window, declared.Window)
	}
	if got.Expectation.Measure != string(declared.Measure) {
		t.Fatalf("carried measure %q != declared measure %q", got.Expectation.Measure, declared.Measure)
	}
}

// TestProducedNonRow4EntriesCarryNoExpectation is the other half: an entry that
// missed nothing must carry nothing, or the field degrades into decoration that
// is present on every entry and therefore says nothing about any of them.
func TestProducedNonRow4EntriesCarryNoExpectation(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	root := fixtureTree(t, twoSubjectManifest, "stale-one.yml", "healthy-one.yml", "undeclared-one.yml")

	q := &fakeQuerier{histories: map[string]RunHistory{
		".github/workflows/stale-one.yml":   {Runs: []Run{{Conclusion: "success", At: now.Add(-30 * 24 * time.Hour)}}},
		".github/workflows/healthy-one.yml": {Runs: []Run{{Conclusion: "success", At: now.Add(-2 * 24 * time.Hour)}}},
	}}

	res, err := testProducer(q, now).Produce(context.Background(), guardliveness.Activation{Root: root})
	if err != nil {
		t.Fatalf("produce: %v", err)
	}

	for _, subject := range []string{
		".github/workflows/healthy-one.yml",    // row 3
		".github/workflows/undeclared-one.yml", // row 7
	} {
		e := entryFor(t, res, subject)
		if e.Expectation != nil {
			t.Fatalf("%s missed nothing yet carries an expectation %+v", subject, *e.Expectation)
		}
	}
}

// ---------------------------------------------------------------------------
// AC-GSM-015 — the published contract holds, in memory AND after the store
// round trip.
// ---------------------------------------------------------------------------

// TestProducedResultHoldsTheContractBothInMemoryAndRoundTripped covers clauses
// (a) and (b) on both forms of the result. The round trip is in the Given
// because a shape that serializes wrongly arrives at the consumer as a contract
// violation that never happened, and because clause (d)'s additive field is the
// newest member of that shape.
func TestProducedResultHoldsTheContractBothInMemoryAndRoundTripped(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	root := fixtureTree(t, twoSubjectManifest, "stale-one.yml", "healthy-one.yml", "undeclared-one.yml")
	q := &fakeQuerier{histories: map[string]RunHistory{
		".github/workflows/stale-one.yml":   {Runs: []Run{{Conclusion: "success", At: now.Add(-30 * 24 * time.Hour)}}},
		".github/workflows/healthy-one.yml": {Runs: []Run{{Conclusion: "success", At: now.Add(-2 * 24 * time.Hour)}}},
	}}

	inMemory, err := testProducer(q, now).Produce(context.Background(), guardliveness.Activation{Root: root})
	if err != nil {
		t.Fatalf("produce: %v", err)
	}

	store := guardliveness.NewStore(t.TempDir())
	if err := store.Save(root, inMemory, now); err != nil {
		t.Fatalf("save: %v", err)
	}
	snapshot, err := store.Load(root)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	for _, form := range []struct {
		name string
		res  guardliveness.Result
	}{
		{"in memory", inMemory},
		{"after the store round trip", snapshot.Result},
	} {
		if len(form.res.Entries) == 0 {
			t.Fatalf("%s: the result carries no entries", form.name)
		}
		// (a) exactly one classification per entry.
		for _, e := range form.res.Entries {
			if len(e.Classifications) != 1 {
				t.Fatalf("%s: %s carries %d classifications, want exactly 1",
					form.name, e.Subject, len(e.Classifications))
			}
		}
		// (b) exactly one value designated clean, asserted on the DESIGNATION
		// rather than by counting how many entries happen to be clean.
		if form.res.Clean == nil {
			t.Fatalf("%s: the result carries no clean-value designation", form.name)
		}
		if len(form.res.Clean.Values) != 1 {
			t.Fatalf("%s: the designation names %d values, want exactly 1",
				form.name, len(form.res.Clean.Values))
		}
		if form.res.Clean.Values[0] != string(ClassOK) {
			t.Fatalf("%s: designated %q clean, want %q", form.name, form.res.Clean.Values[0], ClassOK)
		}
		// The consumer's own refusal path must not fire on a conforming result.
		if _, err := form.res.Partition(); err != nil {
			t.Fatalf("%s: the consumer refused the result: %v", form.name, err)
		}
	}
}

// ---------------------------------------------------------------------------
// AC-GSM-014 — the evaluator mutates nothing.
// ---------------------------------------------------------------------------

// treeDigest is the content-bearing instrument AC-GSM-014 (b) requires: a
// path+hash manifest over the tree excluding .git.
//
// It is content-bearing deliberately and the criterion records why twice: `git
// status` reports path STATUS and never file CONTENT, so no flag combination of
// it can decide "byte-identical" for an ignored path — a cache written inside an
// already-collapsed ignored directory produces no porcelain delta at all.
func treeDigest(t *testing.T, root string) map[string]string {
	t.Helper()

	digest := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		payload, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		sum := sha256.Sum256(payload)
		digest[filepath.ToSlash(rel)] = hex.EncodeToString(sum[:])
		return nil
	})
	if err != nil {
		t.Fatalf("digest the tree: %v", err)
	}
	return digest
}

func digestDelta(before, after map[string]string) []string {
	var out []string
	for path, sum := range after {
		prior, existed := before[path]
		switch {
		case !existed:
			out = append(out, "created: "+path)
		case prior != sum:
			out = append(out, "modified: "+path)
		}
	}
	for path := range before {
		if _, still := after[path]; !still {
			out = append(out, "removed: "+path)
		}
	}
	sort.Strings(out)
	return out
}

// TestProducerMutatesNothing is AC-GSM-014, and BOTH clauses are asserted
// because each alone admits the other's mutant: (a) alone is satisfied by an
// evaluator writing its cache into the working tree, and (b) alone by one that
// opens an issue and touches no file.
//
// The fixture carries the two classifications most likely to tempt an action —
// a STALE entry and an UNDECLARED file — per the criterion's own Given.
func TestProducerMutatesNothing(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	root := fixtureTree(t, twoSubjectManifest, "stale-one.yml", "healthy-one.yml", "undeclared-one.yml")
	q := &fakeQuerier{histories: map[string]RunHistory{
		".github/workflows/stale-one.yml":   {Runs: []Run{{Conclusion: "success", At: now.Add(-30 * 24 * time.Hour)}}},
		".github/workflows/healthy-one.yml": {Runs: []Run{{Conclusion: "success", At: now.Add(-2 * 24 * time.Hour)}}},
	}}

	before := treeDigest(t, root)

	res, err := testProducer(q, now).Produce(context.Background(), guardliveness.Activation{Root: root})
	if err != nil {
		t.Fatalf("produce: %v", err)
	}

	// The fixture actually produced the tempting classifications, or the
	// criterion's Given was not met and the run proves nothing.
	sawStale, sawUndeclared := false, false
	for _, e := range res.Entries {
		switch e.Classifications[0] {
		case string(ClassStale):
			sawStale = true
		case string(ClassUndeclared):
			sawUndeclared = true
		}
	}
	if !sawStale || !sawUndeclared {
		t.Fatalf("the fixture did not produce the Given's classifications (stale=%v undeclared=%v)", sawStale, sawUndeclared)
	}

	// (b) the tree is byte-identical.
	if delta := digestDelta(before, treeDigest(t, root)); len(delta) != 0 {
		t.Fatalf("the evaluator wrote to the tree it evaluated: %v", delta)
	}

	// (a) zero mutating forge calls, counted the way AC-GSM-006 counts
	// queries: the fake records every call it received, and the read method is
	// the only one that was reached.
	if q.globalCalls != 0 {
		t.Fatalf("the repository-global listing was called %d times; REQ-GSM-005 forbids it", q.globalCalls)
	}
	if len(q.subjectCalls) == 0 {
		t.Fatal("no per-subject query was issued: a run that asked nothing cannot demonstrate it mutated nothing")
	}
}

// TestForgeQueryIsPerSubjectAndReadOnly measures the argument vector the real
// gh-backed querier builds, which is the one surface a fake cannot speak for.
//
// Two properties, and the second is the mutation clause at the only layer where
// this deliverable could ever mutate a forge: the subcommand is a listing, and
// the subject is named in the request rather than filtered out of a global one.
func TestForgeQueryIsPerSubjectAndReadOnly(t *testing.T) {
	t.Parallel()

	args := runListArgs(".github/workflows/auto-merge.yml")
	joined := strings.Join(args, " ")

	if len(args) < 2 || args[0] != "run" || args[1] != "list" {
		t.Fatalf("the per-subject query is not a listing: %q", joined)
	}
	if !strings.Contains(joined, "--workflow") {
		t.Fatalf("the query does not name its subject, so it is a global listing filtered afterwards: %q", joined)
	}
	if !strings.Contains(joined, "auto-merge.yml") {
		t.Fatalf("the query does not carry the subject's locator: %q", joined)
	}

	// Every gh verb and flag that changes forge state. The comparison is
	// token-EXACT rather than a substring scan over the joined vector: `--json
	// conclusion,createdAt` contains "create" as a substring of a field name,
	// and a substring scan would report a read query as a mutating one — a
	// false positive that would push an implementer to weaken the check rather
	// than to fix it.
	mutating := map[string]bool{
		"create": true, "delete": true, "edit": true, "comment": true,
		"close": true, "merge": true, "rerun": true, "cancel": true,
		"dispatch": true, "watch": true, "--method": true, "-X": true, "-f": true,
	}
	for i, a := range args {
		// The value of --json is a field list, not a verb.
		if i > 0 && args[i-1] == "--json" {
			continue
		}
		if mutating[a] {
			t.Fatalf("the query carries the mutating token %q: %q", a, joined)
		}
	}

	if global := allRunsArgs(); strings.Contains(strings.Join(global, " "), "--workflow") {
		t.Fatal("the global listing names a workflow; the two queries must stay distinguishable")
	}
}

// ---------------------------------------------------------------------------
// Degradation: what the producer does when it cannot read its own inputs.
// ---------------------------------------------------------------------------

// TestProduceWithoutAManifestFailsRatherThanReportingClean is the failure
// direction that matters. A producer that cannot find its manifest has
// evaluated nothing; returning an empty result would partition to nothing
// non-clean, which the consumer would render as silence — an all-clear about a
// set nobody read, this SPEC's own subject at its own layer.
func TestProduceWithoutAManifestFailsRatherThanReportingClean(t *testing.T) {
	t.Parallel()

	_, err := testProducer(&fakeQuerier{}, time.Now()).
		Produce(context.Background(), guardliveness.Activation{Root: t.TempDir()})
	if err == nil {
		t.Fatal("a producer with no manifest returned a result rather than an error")
	}
	if errors.Is(err, guardliveness.ErrProducerUnwired) {
		t.Fatal("a wired producer reported itself unwired")
	}
}

// TestProduceRejectsAnEmptyRoot records that the producer, not its caller,
// judges an empty root. A caller-side skip is a condition wearing a guard
// clause's clothes, which is what the wiring site's own comment refuses.
func TestProduceRejectsAnEmptyRoot(t *testing.T) {
	t.Parallel()

	_, err := testProducer(&fakeQuerier{}, time.Now()).
		Produce(context.Background(), guardliveness.Activation{})
	if err == nil {
		t.Fatal("an empty root produced a result")
	}
}

// ---------------------------------------------------------------------------
// The real disk enumerator: two INDEPENDENT observations of the same fact.
// ---------------------------------------------------------------------------

// TestDiskEnumeratorPointTestDoesNotConsultTheEnumeration is the property the
// integrity gate rests on. Exists must be a point test of one named locator; an
// implementation that answered from the enumeration would inherit the
// enumeration's own defect and the corroboration would agree with whatever the
// glob got wrong.
//
// Measured by giving the enumerator a pattern that cannot match, then asking it
// about a file that is genuinely there.
func TestDiskEnumeratorPointTestDoesNotConsultTheEnumeration(t *testing.T) {
	t.Parallel()

	root := fixtureTree(t, twoSubjectManifest, "stale-one.yml")
	enum := newDiskEnumerator(root)

	present, err := enum.Exists(".github/workflows/stale-one.yml")
	if err != nil {
		t.Fatalf("point test: %v", err)
	}
	if !present {
		t.Fatal("the point test denies a file that is on disk")
	}

	absent, err := enum.Exists(".github/workflows/never-existed.yml")
	if err != nil {
		t.Fatalf("point test: %v", err)
	}
	if absent {
		t.Fatal("the point test affirms a file that is not on disk")
	}

	// And the enumeration returns repository-relative locators, which is the
	// form the manifest declares — a mismatch here would make every entry read
	// as absent from the enumeration.
	files, err := enum.Enumerate()
	if err != nil {
		t.Fatalf("enumerate: %v", err)
	}
	found := false
	for _, f := range files {
		if f == ".github/workflows/stale-one.yml" {
			found = true
		}
		if filepath.IsAbs(f) {
			t.Fatalf("the enumeration returned an absolute path %q; the manifest declares relative locators", f)
		}
	}
	if !found {
		t.Fatalf("the enumeration did not return the subject: %v", files)
	}
}

// TestDiskEnumeratorPointTestRefusesAnEscapingLocator keeps the point test
// inside the tree it was given. A locator is data from a file, and a point test
// that resolves one outside the root would answer about a subject the
// enumeration could never have seen.
func TestDiskEnumeratorPointTestRefusesAnEscapingLocator(t *testing.T) {
	t.Parallel()

	root := fixtureTree(t, twoSubjectManifest, "stale-one.yml")
	enum := newDiskEnumerator(root)

	// A locator that climbs OUT of the tree. `../<basename(root)>/...` is not
	// this case and was the first thing written here: it climbs one level and
	// descends straight back in, so it resolves inside the root and the test
	// asserted the opposite of what it measured.
	for _, escaping := range []string{
		"../escaped.yml",
		"../../escaped.yml",
		".github/workflows/../../../escaped.yml",
	} {
		if _, err := enum.Exists(escaping); err == nil {
			t.Fatalf("the point test resolved %q, which leaves the evaluated tree", escaping)
		}
	}

	// And the sibling case stays admissible, so the guard is a containment
	// check rather than a ban on any `..` at all.
	if _, err := enum.Exists(".github/workflows/../workflows/stale-one.yml"); err != nil {
		t.Fatalf("a locator that stays inside the tree was refused: %v", err)
	}
}
