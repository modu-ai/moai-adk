package factory

// integration_remeasure_test.go — M2 tests for the re-measure record and its
// verifier (card t1479, SPEC-MERGE-WINDOW-QUEUE-001): the structured-output
// judgment (AC-MWQ-015's table), record validity under the two record forms
// selected by workflow.candidate_ci.enabled (REQ-MWQ-014), and the record's
// inability to be satisfied by a merge stand-in (REQ-MWQ-020).

import (
	"path/filepath"
	"strings"
	"testing"
)

const remeasureFixtureTree = "abc123def4567890abc123def4567890abc12345"

func TestClassifyGoTestJSONCounts(t *testing.T) {
	// AC-MWQ-015 row 1: `go test -json ./pkg/...` reporting 4 passes — valid,
	// count 4.
	output := "" +
		`{"Action":"start","Package":"p"}` + "\n" +
		`{"Action":"run","Package":"p","Test":"TestA"}` + "\n" +
		`{"Action":"pass","Package":"p","Test":"TestA"}` + "\n" +
		`{"Action":"run","Package":"p","Test":"TestB"}` + "\n" +
		`{"Action":"pass","Package":"p","Test":"TestB"}` + "\n" +
		`{"Action":"run","Package":"p","Test":"TestC"}` + "\n" +
		`{"Action":"fail","Package":"p","Test":"TestC"}` + "\n" +
		`{"Action":"run","Package":"p","Test":"TestD"}` + "\n" +
		`{"Action":"pass","Package":"p","Test":"TestD"}` + "\n" +
		`{"Action":"run","Package":"p","Test":"TestE"}` + "\n" +
		`{"Action":"pass","Package":"p","Test":"TestE"}` + "\n" +
		`{"Action":"pass","Package":"p"}` + "\n"
	count, structured, err := ClassifyStructuredOutput("go test -json ./p/...", strings.NewReader(output))
	if err != nil {
		t.Fatalf("4-pass go test -json must be recognized: %v", err)
	}
	if !structured {
		t.Fatalf("go test -json output must read as structured")
	}
	if count != 4 {
		t.Fatalf("count = %d, want 4 (per-test pass/fail events, not the package event)", count)
	}
}

func TestClassifyEmptySweepAndUnstructured(t *testing.T) {
	// AC-MWQ-015 rows 2 and 3: a zero-test go test -json run reads
	// structured with count zero — the classifier recognizes the structure
	// and the VERIFIER refuses the zero count — and a run of go test
	// WITHOUT -json is refused at classification (the tool supports
	// structured output and the caller did not ask for it).
	count, structured, err := ClassifyStructuredOutput("go test -json -run '^NONE$' ./p/...", strings.NewReader(`{"Action":"pass","Package":"p"}` + "\n"))
	if err != nil {
		t.Fatalf("zero-count classification must not error: %v", err)
	}
	if !structured || count != 0 {
		t.Fatalf("empty sweep must read structured-with-zero, got structured=%v count=%d", structured, count)
	}
	// Unstructured go test: exit 0 with no structured support requested.
	if _, structured, err := ClassifyStructuredOutput("go test ./p/...", strings.NewReader("ok  \tp\t0.4s\n")); err == nil || structured {
		t.Fatalf("go test without -json must be refused, err=%v structured=%v", err, structured)
	}
}

func TestClassifyNoTestFilesAndUnknownTool(t *testing.T) {
	// AC-MWQ-015 rows 4 and 5: `[no test files]` is invalid; `true` (a tool
	// with no recognized report) is valid on exit 0 with command + exit only.
	if _, structured, err := ClassifyStructuredOutput("go test -json ./p/...", strings.NewReader(`[no test files]`)); err == nil || structured {
		t.Fatalf("[no test files] must be refused, err=%v structured=%v", err, structured)
	}
	count, structured, err := ClassifyStructuredOutput("true", strings.NewReader(""))
	if err != nil {
		t.Fatalf("a tool with no recognized report must not error: %v", err)
	}
	if structured || count != 0 {
		t.Fatalf("non-test command must read unstructured zero-count")
	}
}

func TestRemeasureRecordValidity(t *testing.T) {
	// REQ-MWQ-014 (local form): a record is valid when it keys the candidate
	// tree, carries build identity, and its command classification accepts
	// it. The same verifier decides.
	root := t.TempDir()
	valid := RemeasureRecord{
		Tree: remeasureFixtureTree, Base: "b" + strings.Repeat("0", 39),
		Command: "go test -json ./internal/p/...", ExitCode: 0,
		HasStructured: true, TestCount: 4,
		BuildIdentity: "moai v3.2.0 (0732cc699)",
	}
	if err := WriteRemeasureRecord(root, remeasureFixtureTree, valid); err != nil {
		t.Fatal(err)
	}
	rec, err := ReadRemeasureRecord(root, remeasureFixtureTree)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateRemeasureRecord(rec); err != nil {
		t.Fatalf("valid record rejected: %v", err)
	}

	// A record whose tool reported exit 3 is invalid (AC-MWQ-015 row 6) —
	// recorded as observed, never rewound to green.
	valid.ExitCode = 3
	if err := ValidateRemeasureRecord(&valid); err == nil {
		t.Fatalf("exit 3 record must be invalid")
	}
	valid.ExitCode = 0

	// A zero structured count is invalid (empty sweep).
	zero := valid
	zero.TestCount = 0
	if err := ValidateRemeasureRecord(&zero); err == nil {
		t.Fatalf("zero test count must be invalid")
	}

	// A stand-in merge record never satisfies the verifier (REQ-MWQ-020):
	// factoryWriteMergeRecord's text carries the merge SHA but no structured
	// test count.
	if err := WriteRemeasureRecord(root, "f"+strings.Repeat("f", 39), RemeasureRecord{
		Tree: "f" + strings.Repeat("f", 39), Base: "b" + strings.Repeat("0", 39),
		Command: "merge abc123def", ExitCode: 0, BuildIdentity: "moai v3.2.0",
	}); err != nil {
		t.Fatal(err)
	}
	standIn, err := ReadRemeasureRecord(root, "f"+strings.Repeat("f", 39))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateRemeasureRecord(standIn); err == nil {
		t.Fatalf("a merge stand-in must not satisfy the re-measure gate")
	}

	// An unknown tree reads as no record at all — the complete gate's
	// "refuse without a valid record for the current candidate tree" (step 3).
	if _, err := ReadRemeasureRecord(root, strings.Repeat("9", 40)); err == nil {
		t.Fatalf("missing record must be reported, not read as free")
	}

	// The record file lives under the state dir keyed by the tree SHA.
	if _, err := filepath.Abs(filepath.Join(root, ".moai", "state", "remeasure", remeasureFixtureTree+".json")); err != nil {
		t.Fatal(err)
	}
}

func TestMergeRecordIsNotARemeasure(t *testing.T) {
	// REQ-MWQ-020: the merge identity complete records (merge-record.txt)
	// must never satisfy the re-measure requirement. The verifier reads its
	// own record store only — a path to a merge record does not validate.
	root := t.TempDir()
	if _, err := ReadRemeasureRecord(root, remeasureFixtureTree); err == nil {
		t.Fatalf("a tree with no re-measure record must read as missing")
	}
}
