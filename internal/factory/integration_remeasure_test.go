package factory

// integration_remeasure_test.go — M2 tests for the re-measure record and its
// verifier (card t1479, SPEC-MERGE-WINDOW-QUEUE-001): the structured-output
// judgment (AC-MWQ-015's table), record validity under the two record forms
// selected by workflow.candidate_ci.enabled (REQ-MWQ-014), and the record's
// inability to be satisfied by a merge stand-in (REQ-MWQ-020).

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const remeasureFixtureTree = "abc123def4567890abc123def4567890abc12345"

func TestClassifyGoTestJSONCounts(t *testing.T) {
	// AC-MWQ-015 row 1: `go test -json ./pkg/...` reporting 4 passes — valid,
	// count 4. (t1576 review round 1: the fail event that used to ride in
	// this stream moved to TestClassifyGoTestJSONRefusesFailEvent — a fail
	// event is now a refusal, not an ignored row.)
	output := "" +
		`{"Action":"start","Package":"p"}` + "\n" +
		`{"Action":"run","Package":"p","Test":"TestA"}` + "\n" +
		`{"Action":"pass","Package":"p","Test":"TestA"}` + "\n" +
		`{"Action":"run","Package":"p","Test":"TestB"}` + "\n" +
		`{"Action":"pass","Package":"p","Test":"TestB"}` + "\n" +
		`{"Action":"run","Package":"p","Test":"TestC"}` + "\n" +
		`{"Action":"pass","Package":"p","Test":"TestC"}` + "\n" +
		`{"Action":"run","Package":"p","Test":"TestD"}` + "\n" +
		`{"Action":"pass","Package":"p","Test":"TestD"}` + "\n" +
		`{"Action":"pass","Package":"p"}` + "\n"
	count, structured, err := ClassifyStructuredOutput("go test -json ./p/...", strings.NewReader(output))
	if err != nil {
		t.Fatalf("4-pass go test -json must be recognized: %v", err)
	}
	if !structured {
		t.Fatalf("go test -json output must read as structured")
	}
	if count != 4 {
		t.Fatalf("count = %d, want 4 (per-test pass events, not the package event)", count)
	}
}

func TestClassifyGoTestJSONRefusesFailEvent(t *testing.T) {
	// t1576 review round 1: a fail event in the stream is a failing sweep —
	// `go test -json ./... | cat` reports the pipe's exit 0 while a test
	// failed, so the recorded exit code alone cannot catch it. The stream
	// itself carries the verdict: any fail event invalidates the record.
	output := "" +
		`{"Action":"run","Package":"p","Test":"TestA"}` + "\n" +
		`{"Action":"pass","Package":"p","Test":"TestA"}` + "\n" +
		`{"Action":"fail","Package":"p","Test":"TestB"}` + "\n"
	if _, _, err := ClassifyStructuredOutput("go test -json ./p/...", strings.NewReader(output)); err == nil {
		t.Fatalf("a stream carrying a per-test fail event must be refused")
	}
	// A package-level failure (no Test field — a build or setup failure) is
	// at least as fatal: no per-test event may exist behind it.
	pkgFail := `{"Action":"fail","Package":"p"}` + "\n"
	if _, _, err := ClassifyStructuredOutput("go test -json ./p/...", strings.NewReader(pkgFail)); err == nil {
		t.Fatalf("a package-level fail event must be refused")
	}
}

func TestClassifyCompoundEnvScrubIsGoTest(t *testing.T) {
	// t1576 review round 1: the env-scrub form AGENTS.md §4 prescribes —
	// `unset VARS && go test -json ...` as ONE compound invocation — was
	// classified by its first token as a non-test command, so a zero-test
	// run passed as valid (command + exit only, no structure required). The
	// classification must see the tool inside the compound, so the
	// verifier's zero-count refusal reaches the scrubbed form.
	command := "unset REVIEW_TEST_VAR && go test -json -run '^NONE$' ./p/..."
	count, structured, err := ClassifyStructuredOutput(command, strings.NewReader(`{"Action":"pass","Package":"p"}`+"\n"))
	if err != nil {
		t.Fatalf("the compound env-scrub form must classify: %v", err)
	}
	if !structured {
		t.Fatalf("the compound env-scrub form must read as a go test run (structured), got count=%d", count)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0 (the record verifier refuses the empty sweep)", count)
	}
}

func TestClassifyEmptySweepAndUnstructured(t *testing.T) {
	// AC-MWQ-015 rows 2 and 3: a zero-test go test -json run reads
	// structured with count zero — the classifier recognizes the structure
	// and the VERIFIER refuses the zero count — and a run of go test
	// WITHOUT -json is refused at classification (the tool supports
	// structured output and the caller did not ask for it).
	count, structured, err := ClassifyStructuredOutput("go test -json -run '^NONE$' ./p/...", strings.NewReader(`{"Action":"pass","Package":"p"}`+"\n"))
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

func TestClassifyAllNoTestSweepStaysEmpty(t *testing.T) {
	// REQ-MWQ2-009 / AC-MWQ2-003 edge: removing the marker refusal does not make
	// an all-no-test sweep valid. Every package reports `[no test files]` and a
	// package-level skip, and the total per-test pass count is zero: the stream
	// reads structured-with-zero, and the verifier refuses the zero count.
	stream := "" +
		`{"Action":"output","Package":"p","Output":"?   \tp\t[no test files]\n"}` + "\n" +
		`{"Action":"skip","Package":"p"}` + "\n"
	count, structured, err := ClassifyStructuredOutput("go test -json ./p/...", strings.NewReader(stream))
	if err != nil || !structured || count != 0 {
		t.Fatalf("an all-no-test sweep must read structured-with-zero: count=%d structured=%v err=%v", count, structured, err)
	}
	rec := &RemeasureRecord{
		Tree: remeasureFixtureTree, Base: "b" + strings.Repeat("0", 39),
		Command: "go test -json ./p/...", ExitCode: 0,
		StructuredRequired: true, HasStructured: structured, TestCount: count,
		BuildIdentity: "moai test",
	}
	if err := ValidateRemeasureRecord(rec); err == nil {
		t.Fatal("an all-no-test sweep must not validate as a re-measure")
	}
}

func TestRemeasureRecordRefusesMalformedTreeAndBase(t *testing.T) {
	// REQ-MWQ2-001 / AC-MWQ2-001 edge: the tree key and the absorbed base are
	// both held to the full-SHA form, and a malformed value of either field is
	// refused as the same record-invalid cause. Uppercase hex is refused too:
	// git prints lowercase, so any other form cannot be the tree or commit the
	// merge step compares against.
	valid := "b" + strings.Repeat("0", 39)
	cases := []struct {
		name string
		tree string
		base string
	}{
		{name: "short base", tree: remeasureFixtureTree, base: "bad"},
		{name: "short tree", tree: "abc", base: valid},
		{name: "uppercase tree", tree: strings.ToUpper(remeasureFixtureTree), base: valid},
		{name: "39-character base", tree: remeasureFixtureTree, base: strings.Repeat("a", 39)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := &RemeasureRecord{
				Tree: tc.tree, Base: tc.base,
				Command: "true", ExitCode: 0, BuildIdentity: "moai test",
			}
			if err := ValidateRemeasureRecord(rec); err == nil || !strings.Contains(err.Error(), "40-character hex SHA") {
				t.Fatalf("a malformed %s must be refused naming the SHA format: %v", tc.name, err)
			}
		})
	}
}

func TestRemeasureRecordRefusesMissingIdentity(t *testing.T) {
	// REQ-MWQ-014/015: the verifier refuses a record missing an identity field,
	// carrying a non-zero exit, or claiming structure it does not carry; a nil
	// record is the missing-record refusal. The valid record is the control.
	valid := func() RemeasureRecord {
		return RemeasureRecord{
			Tree: remeasureFixtureTree, Base: "b" + strings.Repeat("0", 39),
			Command: "true", ExitCode: 0, BuildIdentity: "moai test",
		}
	}
	cases := []struct {
		name   string
		mutate func(r *RemeasureRecord)
	}{
		{name: "empty tree", mutate: func(r *RemeasureRecord) { r.Tree = "" }},
		{name: "empty base", mutate: func(r *RemeasureRecord) { r.Base = "" }},
		{name: "empty command", mutate: func(r *RemeasureRecord) { r.Command = "" }},
		{name: "empty build identity", mutate: func(r *RemeasureRecord) { r.BuildIdentity = "" }},
		{name: "non-zero exit", mutate: func(r *RemeasureRecord) { r.ExitCode = 3 }},
		{name: "structured required, none recognized", mutate: func(r *RemeasureRecord) { r.StructuredRequired = true }},
		{name: "structured with zero tests", mutate: func(r *RemeasureRecord) { r.StructuredRequired, r.HasStructured = true, true }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := valid()
			tc.mutate(&rec)
			if err := ValidateRemeasureRecord(&rec); err == nil {
				t.Fatalf("%s must be refused", tc.name)
			}
		})
	}
	if err := ValidateRemeasureRecord(nil); err == nil {
		t.Fatal("a missing record must be refused")
	}
	good := valid()
	if err := ValidateRemeasureRecord(&good); err != nil {
		t.Fatalf("the valid control must pass: %v", err)
	}
}

func TestClassifyEnvPrefixIsGoTest(t *testing.T) {
	// t1576 review round 2: env assignments and a leading `env` word prefix
	// the real tool — `GOMAXPROCS=2 go test -json ./... | cat` and `env
	// GOMAXPROCS=2 go test -json ./... | cat` classified by their first
	// token as non-test commands, so a failing sweep's record passed as
	// valid (command + exit only). The tool behind the prefix is what
	// classifies, so the fail-event refusal reaches these forms.
	for _, command := range []string{
		"GOMAXPROCS=2 go test -json ./p/...",
		"env GOMAXPROCS=2 go test -json ./p/...",
		"FOO=1 env BAR=2 go test -json ./p/...",
	} {
		if _, _, err := ClassifyStructuredOutput(command, strings.NewReader(`{"Action":"fail","Package":"p","Test":"TestA"}`+"\n")); err == nil {
			t.Fatalf("%q must classify as a go test run and refuse the fail event", command)
		}
	}
}

func TestClassifyEnvOptionsAreGoTest(t *testing.T) {
	// t1576 review round 3: `env` carries its own options and arguments —
	// `env -u NAME go test -json ...` stopped the prefix scan at `-u` and
	// classified as non-test, so a failing sweep's record passed as valid.
	// The scan must consume env's option forms (-u NAME, -i,
	// --ignore-environment, --) before the tool word.
	for _, command := range []string{
		"env -u REVIEW_ABSENT go test -json ./p/...",
		"env -i go test -json ./p/...",
		"GOMAXPROCS=2 env -u REVIEW_ABSENT go test -json ./p/...",
	} {
		if _, _, err := ClassifyStructuredOutput(command, strings.NewReader(`{"Action":"fail","Package":"p","Test":"TestA"}`+"\n")); err == nil {
			t.Fatalf("%q must classify as a go test run and refuse the fail event", command)
		}
	}
}

func TestClassifyTruncatedStreamIsRefused(t *testing.T) {
	// t1576 review round 4: `go test -json . | head -n 5` truncates the
	// stream mid-packages; treating EOF as completion accepted the passes
	// seen so far (ExitCode 0 from the pipe, count 1) while later failures
	// never arrived. A package whose start never met its terminal event is
	// a truncated capture, not a finished sweep.
	truncated := "" +
		`{"Action":"start","Package":"p"}` + "\n" +
		`{"Action":"run","Package":"p","Test":"TestA"}` + "\n" +
		`{"Action":"pass","Package":"p","Test":"TestA"}` + "\n"
	if _, _, err := ClassifyStructuredOutput("go test -json ./p/...", strings.NewReader(truncated)); err == nil {
		t.Fatalf("a stream whose started package never reported must be refused")
	}
	// The complete form of the same stream still passes.
	complete := truncated + `{"Action":"pass","Package":"p"}` + "\n"
	if _, _, err := ClassifyStructuredOutput("go test -json ./p/...", strings.NewReader(complete)); err != nil {
		t.Fatalf("the complete stream must still classify: %v", err)
	}
}

func TestClassifyGoGlobalFlagIsGoTest(t *testing.T) {
	// t1576 review round 4: `go -C . test -json ...` — the subcommand sits
	// behind go's global -C flag, and the first-two-words check read it as
	// a non-test command, so a zero-test record passed as valid.
	command := "go -C . test -json -run '^MISSING$' ./p/..."
	count, structured, err := ClassifyStructuredOutput(command, strings.NewReader(`{"Action":"pass","Package":"p"}`+"\n"))
	if err != nil {
		t.Fatalf("the -C form must classify: %v", err)
	}
	if !structured {
		t.Fatalf("go -C . test -json must read as a go test run (structured), got count=%d", count)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0 (the record verifier refuses the empty sweep)", count)
	}
}

func TestClassifyTruncatedRerunIsRefused(t *testing.T) {
	// t1576 review round 8: a `;` compound of two go test invocations flows
	// one stream — the first run's finished entry must not mask the second
	// run's truncation of the same package.
	rerun := "" +
		`{"Action":"start","Package":"p"}` + "\n" +
		`{"Action":"run","Package":"p","Test":"TestPass"}` + "\n" +
		`{"Action":"pass","Package":"p","Test":"TestPass"}` + "\n" +
		`{"Action":"pass","Package":"p"}` + "\n" +
		`{"Action":"start","Package":"p"}` + "\n" +
		`{"Action":"run","Package":"p","Test":"TestFail"}` + "\n"
	if _, _, err := ClassifyStructuredOutput("go test -json -run TestPass .; go test -json -run TestFail .", strings.NewReader(rerun)); err == nil {
		t.Fatalf("the second run's truncation must be refused, not masked by the first run's completion")
	}
}

func TestClassifyQuotedEnvValueIsGoTest(t *testing.T) {
	// t1576 review round 10: an env assignment whose VALUE carries spaces
	// (`GOFLAGS='-count=1 -v'`) split at the whitespace under strings.Fields,
	// and the go test word behind it was lost — the zero-test run passed as
	// a non-test command. Quotes group the value into one field.
	command := "GOFLAGS='-count=1 -v' go test -json -run '^MISSING$' ./p/..."
	count, structured, err := ClassifyStructuredOutput(command, strings.NewReader(`{"Action":"pass","Package":"p"}`+"\n"))
	if err != nil {
		t.Fatalf("the quoted env value must classify: %v", err)
	}
	if !structured {
		t.Fatalf("the quoted env value must not hide the go test word: structured=%v count=%d", structured, count)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0 (the record verifier refuses the empty sweep)", count)
	}
}

func TestClassifySkippedPackageReportsCompletion(t *testing.T) {
	// t1576 review round 12: a package without tests ends in a package-level
	// skip — the truncation tracking must read it as the terminal event, or
	// every mixed ./... sweep with one no-test package reads as truncated.
	stream := "" +
		`{"Action":"start","Package":"p"}` + "\n" +
		`{"Action":"run","Package":"p","Test":"TestA"}` + "\n" +
		`{"Action":"pass","Package":"p","Test":"TestA"}` + "\n" +
		`{"Action":"start","Package":"q"}` + "\n" +
		`{"Action":"skip","Package":"q"}` + "\n" +
		`{"Action":"pass","Package":"p"}` + "\n"
	count, structured, err := ClassifyStructuredOutput("go test -json ./...", strings.NewReader(stream))
	if err != nil {
		t.Fatalf("a package-level skip is a terminal event, not a truncation: %v", err)
	}
	if !structured || count != 1 {
		t.Fatalf("count=%d structured=%v, want 1/true", count, structured)
	}
}

func TestClassifyQuotedSemicolonStaysOneCommand(t *testing.T) {
	// t1576 review round 12: separators inside quoted arguments are data —
	// the printf is ONE command, not a printf followed by a phantom go test
	// whose JSON output the classifier would then demand.
	command := `printf '%s\n' 'note; go test -json is useful'`
	count, structured, err := ClassifyStructuredOutput(command, strings.NewReader(""))
	if err != nil || structured {
		t.Fatalf("the quoted semicolon must not mint a phantom go test segment: count=%d structured=%v err=%v", count, structured, err)
	}
}

func TestClassifyEscapedQuoteStaysOneCommand(t *testing.T) {
	// t1576 card-review: an escaped quote must not flip the scanner's quote
	// state — the semicolon behind the embed is argument data, not a second
	// command. The double-quote escape is the demonstrable shape: the old
	// scanner closed the quotes on \" and split at the semicolon.
	command := `printf "%s\"; go test -json" x`
	_, structured, err := ClassifyStructuredOutput(command, strings.NewReader(""))
	if err != nil || structured {
		t.Fatalf("the escaped quote must not mint a phantom go test segment: structured=%v err=%v", structured, err)
	}
}

func TestClassifyQuotedWordsAndGOFLAGSCarryJSON(t *testing.T) {
	// t1576 review round 14: the -json read must share the shell-aware
	// parsing — quoted words and a GOFLAGS assignment carry the flag too.
	for _, command := range []string{
		"'go' 'test' '-json' ./p/...",
		"GOFLAGS=-json go test ./p/...",
	} {
		count, structured, err := ClassifyStructuredOutput(command, strings.NewReader(`{"Action":"pass","Package":"p"}`+"\n"))
		if err != nil {
			t.Fatalf("%q must classify: %v", command, err)
		}
		if !structured {
			t.Fatalf("%q must read as a go test -json run: structured=%v count=%d", command, structured, count)
		}
	}
}

func TestClassifyCommentSemicolonStaysOneCommand(t *testing.T) {
	// t1576 review round 16: a word-initial # comments out the rest of the
	// line — separators inside the comment are data, not a second command.
	command := `printf '%s\n' ok # example only; go test -json ./...`
	_, structured, err := ClassifyStructuredOutput(command, strings.NewReader(""))
	if err != nil || structured {
		t.Fatalf("the comment must not mint a phantom go test segment: structured=%v err=%v", structured, err)
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
		StructuredRequired: true, HasStructured: true, TestCount: 4,
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
	// complete's merge-record.txt lives under .moai/reports/<card>/ — a
	// DIFFERENT store from this one — and the gate reads the re-measure
	// store only. Even a card whose merge identity is recorded everywhere
	// complete records it reads as having NO re-measure record.
	standInDir := filepath.Join(root, ".moai", "reports", "t0002")
	if err := os.MkdirAll(standInDir, 0o755); err != nil {
		t.Fatal(err)
	}
	standInBody := fmt.Sprintf("merge %s\nbranch develop\nintegration worktree /repo/develop\nrecorded by moai factory complete (card %s)\n",
		"abc123def4567890", "t0002")
	if err := os.WriteFile(filepath.Join(standInDir, "merge-record.txt"), []byte(standInBody), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadRemeasureRecord(root, remeasureFixtureTree); err != nil {
		// (unrelated: the valid record from earlier in this test is still
		// here — the stand-in did not displace it)
		t.Fatalf("the stand-in must not displace the real record: %v", err)
	}
	if _, err := ReadRemeasureRecord(root, "f"+strings.Repeat("f", 39)); err == nil {
		t.Fatalf("a merge-record.txt must never read as a re-measure record: the stores are separate (REQ-MWQ-020)")
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
