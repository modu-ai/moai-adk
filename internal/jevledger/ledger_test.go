package jevledger

import (
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// SPEC-JEV-LEDGER-001 (card t1762) M2: the ledger recorder. Portable cases
// only: every FIFO case and every Unix-only flag lives in fifo_unix_test.go
// (audit R1). The tests that swap os.Stdout / os.Stderr are not parallel.

const (
	hangGuard  = 2 * time.Second
	ledgerRel  = ".moai/state/jev/decisions.jsonl"
	ledgerDirR = ".moai/state/jev"
)

func ledgerPath(root string) string { return filepath.Join(root, filepath.FromSlash(ledgerRel)) }
func ledgerDir(root string) string  { return filepath.Join(root, filepath.FromSlash(ledgerDirR)) }

func f64(v float64) *float64 { return &v }

// baseRow is an answered row; callers override fields per case.
func baseRow() Row {
	return Row{
		TS: "2026-10-26T10:00:00Z", Consumer: "c1", Event: "stop", Harness: "claude",
		CallID: "call-1", QuestionID: "q1", RubricVersion: "r1",
		ModelID: "jev-1.13.0", RequestModel: "jev-latest", Availability: "available",
		Probability: f64(0.72), InputTokens: 120,
		Mode: "shadow", ConfiguredMode: "advise", Threshold: f64(0),
		Action: "advise", OutcomeLabel: "correct",
	}
}

// appendWithin runs one Append in a goroutine and fails instead of hanging.
func appendWithin(t *testing.T, r *Recorder, row Row) Result {
	t.Helper()
	done := make(chan Result, 1)
	go func() { done <- r.Append(row) }()
	select {
	case res := <-done:
		return res
	case <-time.After(hangGuard):
		t.Fatalf("Append did not return within %v: the recorder blocked", hangGuard)
		return Result{}
	}
}

// captureStd runs fn and returns what reached stdout and stderr.
func captureStd(t *testing.T, fn func()) (stdout, stderr string) {
	t.Helper()
	oldOut, oldErr := os.Stdout, os.Stderr
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	var wg sync.WaitGroup
	var outBuf, errBuf bytes.Buffer
	wg.Add(2)
	go func() { defer wg.Done(); _, _ = io.Copy(&outBuf, outR) }()
	go func() { defer wg.Done(); _, _ = io.Copy(&errBuf, errR) }()
	os.Stdout, os.Stderr = outW, errW
	restore := func() { os.Stdout, os.Stderr = oldOut, oldErr }
	defer restore()
	fn()
	restore()
	_ = outW.Close()
	_ = errW.Close()
	wg.Wait()
	return outBuf.String(), errBuf.String()
}

// readLines returns the ledger's lines, each decoded; it fails if the file is
// missing or a line is not JSON.
func readRows(t *testing.T, root string) []map[string]any {
	t.Helper()
	data, err := os.ReadFile(ledgerPath(root))
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	var rows []map[string]any
	for i, line := range strings.Split(strings.TrimSuffix(string(data), "\n"), "\n") {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("line %d is not JSON: %v: %q", i, err, line)
		}
		rows = append(rows, m)
	}
	return rows
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustWrite(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// walkFiles returns every non-directory entry under root.
func walkFiles(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			out = append(out, p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return out
}

// AC-JEVL-001 (recorder half): constructing a recorder creates nothing; one
// append creates the ledger holding exactly one line, so the zero before it is
// attributable to the absence of an append and not to a recorder that cannot
// write.
func TestLedger_NothingUntilAppend(t *testing.T) {
	root := t.TempDir()
	r := New(root)
	if _, err := os.Lstat(filepath.Join(root, ".moai")); !os.IsNotExist(err) {
		t.Fatalf("constructing a recorder created something under the root: Lstat err = %v", err)
	}
	if res := appendWithin(t, r, baseRow()); !res.Landed {
		t.Fatalf("positive control failed: the append did not land (%s)", res.Reason)
	}
	if rows := readRows(t, root); len(rows) != 1 {
		t.Fatalf("positive control: ledger holds %d lines, want exactly 1", len(rows))
	}
}

// AC-JEVL-009: the three rows of the acceptance fixture.
func ledgerFixtureRows() []Row {
	choice := baseRow()
	choice.QuestionID, choice.Probability, choice.Threshold = "q-choice", f64(0.72), f64(0)

	noul := baseRow()
	noul.QuestionID, noul.Probability, noul.Threshold = "q-noul", f64(0.31), f64(0.65)
	noul.ConfigWarning = "mode bogus"

	unanswered := baseRow()
	unanswered.QuestionID, unanswered.ModelID, unanswered.Availability = "q-unanswered", "", "unavailable"
	unanswered.Probability, unanswered.Threshold = nil, nil
	return []Row{choice, noul, unanswered}
}

func appendFixture(t *testing.T) []map[string]any {
	t.Helper()
	root := t.TempDir()
	r := New(root)
	for _, row := range ledgerFixtureRows() {
		if res := appendWithin(t, r, row); !res.Landed {
			t.Fatalf("append did not land: %s", res.Reason)
		}
	}
	rows := readRows(t, root)
	if len(rows) != 3 {
		t.Fatalf("ledger holds %d lines, want 3 (the field assertions below would be vacuous)", len(rows))
	}
	return rows
}

func TestLedgerRow_Fields(t *testing.T) {
	rows := appendFixture(t)
	for i, row := range rows {
		for _, key := range []string{
			"model_id", "request_model", "rubric_version", "question_id", "probability", "mode", "action",
			"outcome_label", "ts", "consumer", "event", "harness", "availability", "input_tokens", "call_id",
			"threshold", "configured_mode",
		} {
			if _, ok := row[key]; !ok {
				t.Errorf("row %d lacks the key %q", i, key)
			}
		}
		if row["mode"] != "shadow" || row["configured_mode"] != "advise" {
			t.Errorf("row %d mode/configured_mode = %v/%v, want shadow/advise", i, row["mode"], row["configured_mode"])
		}
		if row["model_id"] == "jev-latest" {
			t.Errorf("row %d model_id equals the request alias", i)
		}
		if row["request_model"] != "jev-latest" {
			t.Errorf("row %d request_model = %v, want jev-latest", i, row["request_model"])
		}
	}
	if rows[0]["probability"] != 0.72 || rows[1]["probability"] != 0.31 {
		t.Errorf("probabilities = %v, %v, want the supplied raw values 0.72, 0.31", rows[0]["probability"], rows[1]["probability"])
	}
	if rows[0]["model_id"] != "jev-1.13.0" || rows[0]["rubric_version"] != "r1" || rows[0]["question_id"] != "q-choice" {
		t.Errorf("row 0 identity fields wrong: %v", rows[0])
	}
	if _, ok := rows[0]["config_warning"]; ok {
		t.Error("row 0 carries config_warning though none was supplied")
	}
	if rows[1]["config_warning"] != "mode bogus" {
		t.Errorf("row 1 config_warning = %v, want %q", rows[1]["config_warning"], "mode bogus")
	}
}

func TestLedgerRow_UnansweredNullProbability(t *testing.T) {
	rows := appendFixture(t)
	un := rows[2]
	// request_model is never copied into model_id: an unanswered row has an
	// empty model_id and a null probability.
	if un["model_id"] != "" {
		t.Errorf("unanswered model_id = %v, want the empty string", un["model_id"])
	}
	if v, ok := un["probability"]; !ok || v != nil {
		t.Errorf("unanswered probability = %v (present=%v), want null", v, ok)
	}
	if un["request_model"] != "jev-latest" {
		t.Errorf("unanswered request_model = %v, want jev-latest", un["request_model"])
	}
}

func TestLedgerRow_RecorderOwnsActionAndOutcome(t *testing.T) {
	rows := appendFixture(t)
	for i, row := range rows {
		// The caller supplied action "advise" and outcome_label "correct".
		if row["action"] != "none" || row["outcome_label"] != "unlabeled" {
			t.Errorf("row %d action/outcome_label = %v/%v, want none/unlabeled whatever the caller supplied", i, row["action"], row["outcome_label"])
		}
	}
	// A row whose mode is not shadow keeps the caller's action, but
	// outcome_label is still the recorder's.
	root := t.TempDir()
	row := baseRow()
	row.Mode, row.Action = "advise", "advise"
	if res := appendWithin(t, New(root), row); !res.Landed {
		t.Fatalf("append did not land: %s", res.Reason)
	}
	got := readRows(t, root)[0]
	if got["action"] != "advise" || got["outcome_label"] != "unlabeled" {
		t.Errorf("non-shadow row action/outcome_label = %v/%v, want advise/unlabeled", got["action"], got["outcome_label"])
	}
}

// AC-JEVL-009 mutants 4-6: threshold is always present, 0 is a number and an
// absent threshold is null.
func TestLedgerRow_Threshold(t *testing.T) {
	rows := appendFixture(t)
	if v, ok := rows[0]["threshold"]; !ok || v != float64(0) {
		t.Errorf("choice row threshold = %v (present=%v), want the number 0", v, ok)
	}
	if v, ok := rows[1]["threshold"]; !ok || v != 0.65 {
		t.Errorf("noul row threshold = %v (present=%v), want 0.65", v, ok)
	}
	if v, ok := rows[2]["threshold"]; !ok || v != nil {
		t.Errorf("unanswered row threshold = %v (present=%v), want null", v, ok)
	}
}

// AC-JEVL-010: 50 concurrent appends never interleave a row.
func TestLedger_ConcurrentAppend(t *testing.T) {
	root := t.TempDir()
	r := New(root)
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			row := baseRow()
			row.CallID = "call-" + strings.Repeat("x", 40) + string(rune('A'+i%26))
			if res := r.Append(row); !res.Landed {
				t.Errorf("append %d dropped: %s", i, res.Reason)
			}
		}()
	}
	wg.Wait()
	if rows := readRows(t, root); len(rows) != 50 {
		t.Fatalf("ledger holds %d parseable lines, want 50", len(rows))
	}
}

// AC-JEVL-010: POSIX modes. Skipped by name on windows, which has no POSIX
// mode bits; the file still compiles there because it uses os.FileMode only.
func TestLedger_Modes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX mode bits are not meaningful on windows")
	}
	root := t.TempDir()
	if res := appendWithin(t, New(root), baseRow()); !res.Landed {
		t.Fatalf("append did not land: %s", res.Reason)
	}
	dirInfo, err := os.Stat(ledgerDir(root))
	if err != nil {
		t.Fatal(err)
	}
	fileInfo, err := os.Stat(ledgerPath(root))
	if err != nil {
		t.Fatal(err)
	}
	if got := dirInfo.Mode().Perm(); got != 0o700 {
		t.Errorf("ledger directory mode = %o, want 700", got)
	}
	if got := fileInfo.Mode().Perm(); got != 0o600 {
		t.Errorf("ledger file mode = %o, want 600", got)
	}
}

// AC-JEVL-018: a row is appended after malformed lines, which stay byte for
// byte; the recorder does not drop the row on account of them.
func TestLedger_AppendAfterMalformedLines(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, ledgerDir(root))
	valid, _ := json.Marshal(baseRow())
	prefix := append(append([]byte{}, valid...), []byte("\nnot json\n{\"truncated\":\n")...)
	mustWrite(t, ledgerPath(root), prefix)

	if res := appendWithin(t, New(root), baseRow()); !res.Landed {
		t.Fatalf("the row was dropped on account of the malformed lines: %s", res.Reason)
	}
	got, err := os.ReadFile(ledgerPath(root))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(got, prefix) {
		t.Fatalf("the existing bytes changed:\nwant prefix %q\ngot        %q", prefix, got)
	}
	rest := string(got[len(prefix):])
	if strings.Count(rest, "\n") != 1 || !strings.HasSuffix(rest, "\n") {
		t.Fatalf("expected exactly one new line, got %q", rest)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(rest), &m); err != nil || m["outcome_label"] != "unlabeled" {
		t.Errorf("the appended line is not a full row: err=%v row=%v", err, m)
	}
}

// A row that cannot be encoded (a NaN probability has no JSON form) is dropped
// before anything touches the disk: no panic, and no directory is created for a
// row that was never going to land.
func TestLedger_UnencodableRowDropped(t *testing.T) {
	root := t.TempDir()
	row := baseRow()
	row.Probability = f64(math.NaN())

	res := appendWithin(t, New(root), row)
	if res.Landed || res.Reason == "" {
		t.Errorf("result = %+v, want a dropped result that names a reason", res)
	}
	if _, err := os.Lstat(filepath.Join(root, ".moai")); !os.IsNotExist(err) {
		t.Errorf("a row that could not be encoded still created %s: %v", filepath.Join(root, ".moai"), err)
	}
}

// AC-JEVL-012 (i): .moai/state/jev is a regular file, so the directory cannot
// be created. The row is dropped, the file keeps its bytes, and no ledger
// appears anywhere under the root.
func TestLedger_WriteFailureDropped(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Dir(ledgerDir(root)))
	recorded := []byte("a regular file where the directory should be\n")
	mustWrite(t, ledgerDir(root), recorded)

	var res Result
	stdout, stderr := captureStd(t, func() { res = appendWithin(t, New(root), baseRow()) })
	if res.Landed || res.Reason == "" {
		t.Errorf("result = %+v, want a dropped result that names a reason", res)
	}
	if stdout != "" || stderr != "" {
		t.Errorf("a drop must be silent: stdout=%q stderr=%q", stdout, stderr)
	}
	if got, _ := os.ReadFile(ledgerDir(root)); !bytes.Equal(got, recorded) {
		t.Errorf("the file at .moai/state/jev changed: %q", got)
	}
	for _, p := range walkFiles(t, root) {
		if filepath.Base(p) == "decisions.jsonl" {
			t.Errorf("a ledger appeared at %s", p)
		}
	}
}

// AC-JEVL-012 (iii) and (v): a symbolic link and a directory at the ledger
// path. Case (v) is portable (os.Mkdir only) and runs behind the same hang
// guard as the FIFO cases.
func TestLedger_NonRegularPathDropped(t *testing.T) {
	t.Run("symlink", func(t *testing.T) {
		root := t.TempDir()
		outside := filepath.Join(t.TempDir(), "outside.txt")
		recorded := []byte("bytes outside the project root\n")
		mustWrite(t, outside, recorded)
		mustMkdirAll(t, ledgerDir(root))
		if err := os.Symlink(outside, ledgerPath(root)); err != nil {
			t.Skipf("os.Symlink unavailable here: %v", err)
		}

		var res Result
		stdout, stderr := captureStd(t, func() { res = appendWithin(t, New(root), baseRow()) })
		if res.Landed || res.Reason == "" {
			t.Errorf("result = %+v, want a dropped result that names a reason", res)
		}
		if stdout != "" || stderr != "" {
			t.Errorf("a drop must be silent: stdout=%q stderr=%q", stdout, stderr)
		}
		if info, err := os.Lstat(ledgerPath(root)); err != nil || info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("the link was replaced: err=%v mode=%v", err, info)
		}
		if got, _ := os.ReadFile(outside); !bytes.Equal(got, recorded) {
			t.Errorf("the link target changed: %q", got)
		}
	})

	t.Run("directory", func(t *testing.T) {
		root := t.TempDir()
		mustMkdirAll(t, ledgerPath(root)) // an empty directory AT the ledger path

		var res Result
		stdout, stderr := captureStd(t, func() { res = appendWithin(t, New(root), baseRow()) })
		if res.Landed || res.Reason == "" {
			t.Errorf("result = %+v, want a dropped result that names a reason", res)
		}
		if stdout != "" || stderr != "" {
			t.Errorf("a drop must be silent: stdout=%q stderr=%q", stdout, stderr)
		}
		info, err := os.Lstat(ledgerPath(root))
		if err != nil || !info.IsDir() {
			t.Fatalf("the directory at the ledger path is gone: err=%v", err)
		}
		if entries, _ := os.ReadDir(ledgerPath(root)); len(entries) != 0 {
			t.Errorf("the directory holds %d entries after the append, want none", len(entries))
		}
		if files := walkFiles(t, ledgerDir(root)); len(files) != 0 {
			t.Errorf("a row landed in %v", files)
		}
	})
}
