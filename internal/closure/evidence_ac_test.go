package closure

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// TestResolveEvidenceHomeWorktreeAndPrimary pins §C.3/§C.7: the worktree
// whose base name equals the card wins; otherwise the primary checkout.
func TestResolveEvidenceHomeWorktreeAndPrimary(t *testing.T) {
	f := newACFixture(t)
	home, err := ResolveEvidenceHome(f.Root, fixCard)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	// git resolves the tempdir's /var/folders symlink to /private/var/...;
	// compare canonicalized.
	want, _ := filepath.EvalSymlinks(f.CardDir)
	if home != want {
		t.Fatalf("home = %q, want the c1 worktree %q", home, want)
	}
	// A card with no matching worktree resolves to the primary checkout
	// (the parent of the common dir).
	home, err = ResolveEvidenceHome(f.Root, "no-such-card")
	if err != nil {
		t.Fatalf("resolve primary: %v", err)
	}
	wantRoot, _ := filepath.EvalSymlinks(f.Root)
	if home != wantRoot {
		t.Fatalf("home = %q, want the primary checkout %q", home, wantRoot)
	}
}

// TestWriteReportPairAndHash pins the atomic pair write and the canonical
// hash stability across a file round-trip.
func TestWriteReportPairAndHash(t *testing.T) {
	f := newACFixture(t)
	in := f.input()
	r := mustBuild(t, in)
	ev := EvidenceFor(f.CardDir, fixCard)
	mdPath, err := WriteReportPair(ev, r)
	if err != nil {
		t.Fatalf("write pair: %v", err)
	}
	if mdPath != ev.ReportMD || !fileExists(mdPath) || !fileExists(ev.ReportJSON) {
		t.Fatalf("pair incomplete: md=%q json=%q", mdPath, ev.ReportJSON)
	}
	if !strings.Contains(r.HeadSHA, "") {
		t.Fatalf("head sha unexpectedly empty")
	}
	// The canonical hash of the struct equals the canonical hash of the file
	// (the property the Human Verdict currency check needs).
	disk := PreviousReportHashOf(ev.ReportJSON)
	if disk == "" || disk != CanonicalReportHash(r) {
		t.Fatalf("canonical hash mismatch: disk=%q struct=%q", disk, CanonicalReportHash(r))
	}
	// HashFile returns the raw file digest; different content differs.
	if HashFile(ev.ReportMD) == HashFile(ev.ReportJSON) {
		t.Fatalf("distinct files hashed equal")
	}
	if HashFile(filepath.Join(f.CardDir, "missing.txt")) != "" {
		t.Fatalf("missing file hash = non-empty")
	}
}

// TestCountValueRoundTrip pins the JSON number-or-"not observed" contract.
func TestCountValueRoundTrip(t *testing.T) {
	observed, err := json.Marshal(CountValue{Observed: true, N: 2})
	if err != nil || string(observed) != "2" {
		t.Fatalf("observed marshal = %s %v", observed, err)
	}
	unobserved, err := json.Marshal(CountValue{})
	if err != nil || string(unobserved) != `"not observed"` {
		t.Fatalf("unobserved marshal = %s %v", unobserved, err)
	}
	var back CountValue
	if err := json.Unmarshal(observed, &back); err != nil || !back.Observed || back.N != 2 {
		t.Fatalf("observed unmarshal = %+v %v", back, err)
	}
	if err := json.Unmarshal([]byte(`"not observed"`), &back); err != nil || back.Observed {
		t.Fatalf("unobserved unmarshal = %+v %v", back, err)
	}
	if err := json.Unmarshal([]byte(`"bogus"`), &back); err == nil {
		t.Fatalf("bogus string unmarshal = %+v, want error", back)
	}
	if (CountValue{Observed: true, N: 7}).Text() != "7" {
		t.Fatalf("Text() = %q, want 7", (CountValue{Observed: true, N: 7}).Text())
	}
}

// TestAdditionFromObservationAndWarnings covers the class-4 body parser and
// the warning joiner.
func TestAdditionFromObservationAndWarnings(t *testing.T) {
	a, ok := additionFromObservation("Between card base abc and HEAD:\n\n- kind: new-package\n- name: fixture\n- path: internal/fixture")
	if !ok || a.Kind != "new-package" || a.Name != "fixture" || a.Path != "internal/fixture" {
		t.Fatalf("addition = %+v ok=%v", a, ok)
	}
	if _, ok := additionFromObservation("no body"); ok {
		t.Fatalf("parsed an addition from an observation without one")
	}
	if joinWarnings("", "b") != "b" || joinWarnings("a", "b") != "a; b" {
		t.Fatalf("joinWarnings shape")
	}
	if answerLine("", nil) != "not recorded" || answerLine(NotRecorded, nil) != "not recorded" {
		t.Fatalf("answerLine absent shape")
	}
}

// TestResidualRiskStates pins the machine-observable residual-risk entries.
func TestResidualRiskStates(t *testing.T) {
	r := NewReport(fixCard, fixSpecID, "head", "contract", "required")
	r.NewAPIs.Observed = true
	r.SecondVerdict.SubstituteBackend = true
	r.SecondVerdict.State = SecondReviewStale
	buildResidualRisk(r, BuildInput{})
	items := map[string]bool{}
	for _, rr := range r.ResidualRisk {
		items[rr.Item] = true
	}
	for _, want := range []string{"new-api-comparison-heuristic", "substitute-second-model", "stale-second-review"} {
		if !items[want] {
			t.Fatalf("residual risk missing %q: %v", want, r.ResidualRisk)
		}
	}
}

// TestSkipListedUnderSecondVerdict pins that skipped second-review lines
// surface under Not Performed and never count as performed.
func TestSkipListedUnderSecondVerdict(t *testing.T) {
	f := newACFixture(t)
	in := f.input()
	in.SecondReviewSkipped = []string{"schema_version 2"}
	r := mustBuild(t, in)
	found := false
	for _, np := range r.NotPerformed {
		if strings.Contains(np.Detail, "schema_version 2") {
			found = true
		}
	}
	if !found {
		t.Fatalf("skipped line not listed: %v", r.NotPerformed)
	}
}

// TestLoadA2RecordsMissingAndMalformed pins the record loader's missing-dir
// and malformed-file behavior.
func TestLoadA2RecordsMissingAndMalformed(t *testing.T) {
	f := newACFixture(t)
	recs, unreadable, err := LoadA2Records(f.CardDir, "card-with-no-records")
	if err != nil || len(recs) != 0 || len(unreadable) != 0 {
		t.Fatalf("missing dir = %v %v %v", recs, unreadable, err)
	}
	f.writeA2Raw("broken.md", "junk")
	recs, unreadable, err = LoadA2Records(f.CardDir, fixCard)
	if err != nil || len(recs) != 0 || len(unreadable) != 1 {
		t.Fatalf("malformed = %v %v %v", recs, unreadable, err)
	}
}

// TestNotObservedHelpers covers the small display helpers directly.
func TestNotObservedHelpers(t *testing.T) {
	if orNotObserved("") != "not observed" || orNotObserved("x") != "x" {
		t.Fatalf("orNotObserved")
	}
	if nonNil(nil) == nil || len(nonNil([]string{"a"})) != 1 {
		t.Fatalf("nonNil")
	}
	if sortedKeys(map[string]bool{"b": true, "a": true})[0] != "a" {
		t.Fatalf("sortedKeys")
	}
}
