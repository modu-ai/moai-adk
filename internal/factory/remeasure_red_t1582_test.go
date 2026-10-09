package factory

// remeasure_red_t1582_test.go — plan-phase RED reproduction tests for card
// t1582 (SPEC-MERGE-WINDOW-QUEUE-002). Each test is RED on the pre-repair
// tree (fails for the defect it names) and flips GREEN when the run phase
// repairs it. The probe test at the bottom is an observation record, not a
// RED test: it pins a documented residual, not a defect.

import (
	"strings"
	"testing"
	"time"
)

// R7-2 (card t1582 item ④): the empty-sweep marker ANYWHERE in the stream
// refuses the whole re-measure, so a normal `go test -json ./...` sweep over
// a real package plus a no-test auxiliary package is wrongly refused. The
// marker in a -json stream arrives as an output event of the no-test
// package — the per-test pass count of the real package is one, and the
// package-level skip is that package's terminal event, not a truncation.
//
// RED on the pre-repair tree: countGoTestJSONTests returns the marker
// refusal. GREEN after the repair: one per-test pass counted, no error.
func TestRedT1582MixedSweepWithNoTestPackageCounts(t *testing.T) {
	stream := strings.Join([]string{
		`{"Action":"start","Package":"example.com/a"}`,
		`{"Action":"run","Package":"example.com/a","Test":"TestA"}`,
		`{"Action":"pass","Package":"example.com/a","Test":"TestA"}`,
		`{"Action":"pass","Package":"example.com/a"}`,
		`{"Action":"output","Package":"example.com/b","Output":"?   \tb\t[no test files]\n"}`,
		`{"Action":"skip","Package":"example.com/b"}`,
	}, "\n") + "\n"
	count, structured, err := countGoTestJSONTests(stream)
	if err != nil {
		t.Fatalf("RED t1582-R7-2: the [no test files] marker anywhere in the stream refuses a mixed sweep that measured 1 passing test: %v", err)
	}
	if !structured {
		t.Fatal("the mixed sweep must read as a recognized structured report")
	}
	if count != 1 {
		t.Fatalf("the mixed sweep must count the real package's 1 per-test pass, got %d", count)
	}
}

// R7-3 unit component (card t1582 item ④, TOP priority): the ONE verifier
// admits a record whose Base is not a SHA — 3 bytes pass every gate — and
// the cause-2 message render downstream slices it to 12. RED on the
// pre-repair tree: the verifier returns nil. GREEN after the repair: a
// malformed Base is refused as record-invalid before any message render.
func TestRedT1582VerifierAdmitsAShortBaseSHA(t *testing.T) {
	rec := &RemeasureRecord{
		Tree: "0123456789abcdef0123456789abcdef01234567",
		Base: "bad",
		Command: "true", ExitCode: 0, StructuredRequired: false,
		BuildIdentity: "moai test",
	}
	if err := ValidateRemeasureRecord(rec); err == nil {
		t.Fatal("RED t1582-R7-3: the verifier admits a record whose Base is 3 bytes — it reaches the [:12] message renders downstream")
	}
}

// R7-3 end-to-end (card t1582 item ④, TOP priority): a record with a short
// base makes the cause-2 message render panic (`record.Base[:12]` on 3
// bytes) and the panic kills the step BEFORE releaseWindow — the integration
// window stays held for nobody until the lease lapses.
//
// RED on the pre-repair tree: the step panics. GREEN after the repair: the
// step refuses with the record-invalid code, releases the window, and
// promotes the waiting ticket.
func TestRedT1582ShortBaseSHAPanicsBeforeWindowRelease(t *testing.T) {
	f := newMergeFixture(t)
	// Overwrite the fixture's valid record with the malformed one — a
	// hand-written record is exactly what the verifier's trust model already
	// admits (integration_remeasure.go § package comment: actors are
	// cooperative, forgery is a residual risk).
	rec := *f.record
	rec.Base = "bad"
	if err := WriteRemeasureRecord(f.root, rec.Tree, rec); err != nil {
		t.Fatal(err)
	}
	card := f.withCardTree(readyCardPtr())
	var panicked any
	var stepErr error
	func() {
		defer func() { panicked = recover() }()
		_, stepErr = RunMergeStep(f.input(), f.seams(card))
	}()
	if panicked != nil {
		t.Fatalf("RED t1582-R7-3: the merge step panicked on the short base %q (%v) and died before releasing the window — the window stays held", rec.Base, panicked)
	}
	if stepErr == nil {
		t.Fatal("the step must refuse the malformed record, not succeed")
	}
	requireCode(t, stepErr, MergeExitRecordInvalid)
	// The refusal must have released the window and promoted the waiter —
	// the exact thing the panic skipped.
	requireWindowReleasedAndCPromoted(t, f)
}

// Item ② round-trip half (card t1582, the classification side): the lines
// shellJoinArgs builds on the cli side (internal/cli/shelljoin_red_t1582_test.go
// pins their exact quoting) must re-parse in the classifiers WITH the
// boundaries the caller passed. A boundary that survives the join but
// collapses here is the divergence the card hunts. The literals mirror the
// cli side's wantJoin values verbatim; if the join changes, the cli test
// fails first and both halves get updated together.
func TestRedT1582ClassifiersReadBackTheJoinedBoundaries(t *testing.T) {
	cases := []struct {
		name       string
		line       string
		wantFields []string
	}{
		{
			name:       "regex with spaces and pipe",
			line:       "go test -json -run 'Test A|Test B' ./pkg",
			wantFields: []string{"go", "test", "-json", "-run", "Test A|Test B", "./pkg"},
		},
		{
			name:       "regex with semicolon",
			line:       "go test -json -run 'Test A;Test B' ./pkg",
			wantFields: []string{"go", "test", "-json", "-run", "Test A;Test B", "./pkg"},
		},
		{
			name:       "regex with ampersand",
			line:       "go test -json -run 'Test A&Test B' ./pkg",
			wantFields: []string{"go", "test", "-json", "-run", "Test A&Test B", "./pkg"},
		},
		{
			name:       "env assignment value with spaces",
			line:       "GOFLAGS='-json -v' go test ./pkg",
			wantFields: []string{"GOFLAGS=-json -v", "go", "test", "./pkg"},
		},
		{
			// Observed on the pre-repair tree (plan-phase probe): the
			// classifier keeps the escaped quote as the two literal bytes
			// `\'` instead of folding them — the boundary itself survives
			// (one field, never split), so the tool and -json judgments are
			// unaffected. The current behavior is SEALED here, not repaired:
			// the card's divergence question (a boundary lost at
			// classification time) is NOT reproduced.
			name:       "apostrophe inside argument",
			line:       "go test -json -run 'it'\\''s fine' ./pkg",
			wantFields: []string{"go", "test", "-json", "-run", `it\'s fine`, "./pkg"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got []string
			for _, seg := range shellSegments(tc.line) {
				got = append(got, shellFields(seg)...)
			}
			t.Logf("classifier fields: %q", got)
			if strings.Join(got, "\x00") != strings.Join(tc.wantFields, "\x00") {
				t.Fatalf("RED t1582-item2 (classification half): a boundary the join preserved collapses at classification — got %q, want %q", got, tc.wantFields)
			}
		})
	}
}

// Probe (card item ①, NOT a RED test — an observation record): a
// non-go-test command whose output carries failure semantics but whose exit
// is 0 stands as a valid record. This is the residual the SPEC's §D trust
// model names ("a tool with no recognized report ... valid on exit code
// zero") and the candidate-CI form owns its removal. Recorded here so the
// plan-phase sweep documents WHY it is not a repair target.
func TestRedT1582ProbeNonTestCommandFailureSemanticsRideExitZero(t *testing.T) {
	output := "FAIL: something happened\n"
	count, structured, err := ClassifyStructuredOutput("printf 'FAIL: something happened\\n'", strings.NewReader(output))
	if err != nil {
		t.Fatalf("the non-test command must classify without error: %v", err)
	}
	if structured || count != 0 {
		t.Fatalf("a non-test command must not read as structured: structured=%t count=%d", structured, count)
	}
	rec := &RemeasureRecord{
		Tree: "0123456789abcdef0123456789abcdef01234567",
		Base: "0123456789abcdef0123456789abcdef01234567",
		Command: "printf 'FAIL: something happened\\n'", ExitCode: 0,
		StructuredRequired: false, BuildIdentity: "moai test",
		RecordedAt: time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
	}
	if err := ValidateRemeasureRecord(rec); err != nil {
		t.Fatalf("expected the documented residual to hold (valid on exit zero): %v", err)
	}
	t.Log("observed: failure semantics in a non-test command's output ride exit 0 to a valid record — spec.md §D residual, not repaired here")
}
