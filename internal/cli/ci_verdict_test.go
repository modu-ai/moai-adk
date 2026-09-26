package cli

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/civerdict"
)

const testHead = "0123456789abcdef0123456789abcdef01234567"

// verdictFiles lists the record files written under root's ci-verdicts dir.
func verdictFiles(t *testing.T, root string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, ".moai", "state", "ci-verdicts"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

// runCIVerdict executes the producer verb with args and returns its stdout.
func runCIVerdict(t *testing.T, runner ghRunner, root string, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd := newCIVerdictCmd(runner)
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(append([]string{"--project-root", root}, args...))
	err := cmd.Execute()
	return out.String(), err
}

// AC-CV-001 (REQ-CV-001, REQ-CV-002, REQ-CV-004): the offline --from-json
// path records the input verdict; the same invocation re-run is a byte-
// identical rewrite; a changed conclusion overwrites (last-writer-wins).
func TestCIVerdictFromJSON(t *testing.T) {
	root := t.TempDir()
	in := filepath.Join(t.TempDir(), "verdict.json")
	body := `{"head_sha":"` + testHead + `","conclusion":"failure","run_id":"run-42","observed_at":"2026-09-26T12:00:00Z","producer":"test"}`
	if err := os.WriteFile(in, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runCIVerdict(t, defaultGhRunner, root, "--from-json", in)
	if err != nil {
		t.Fatalf("first run: %v\n%s", err, out)
	}
	rec, err := civerdict.Load(root, testHead)
	if err != nil || rec == nil {
		t.Fatalf("Load: %v, %v", rec, err)
	}
	if rec.Conclusion != "failure" || rec.RunID != "run-42" || rec.Producer != "test" {
		t.Errorf("recorded = %+v", rec)
	}
	first, err := os.ReadFile(civerdict.Path(root, testHead))
	if err != nil {
		t.Fatal(err)
	}
	// Idempotent re-run.
	if _, err := runCIVerdict(t, defaultGhRunner, root, "--from-json", in); err != nil {
		t.Fatalf("re-run: %v", err)
	}
	second, err := os.ReadFile(civerdict.Path(root, testHead))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Errorf("idempotent rewrite is not byte-identical")
	}
	// Changed conclusion overwrites.
	over := strings.Replace(body, `"failure"`, `"success"`, 1)
	if err := os.WriteFile(in, []byte(over), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runCIVerdict(t, defaultGhRunner, root, "--from-json", in); err != nil {
		t.Fatalf("overwrite run: %v", err)
	}
	rec, _ = civerdict.Load(root, testHead)
	if rec == nil || rec.Conclusion != "success" {
		t.Errorf("last-writer-wins record = %+v", rec)
	}
}

// AC-CV-002 (REQ-CV-001): fetch mode records the fabricated gh result for
// the head — no real gh, no network.
func TestCIVerdictFetch(t *testing.T) {
	root := t.TempDir()
	var gotArgs []string
	runner := func(args ...string) ([]byte, error) {
		gotArgs = args
		return []byte(`[{"conclusion":"failure","databaseId":4242}]`), nil
	}
	out, err := runCIVerdict(t, runner, root, "--head", testHead)
	if err != nil {
		t.Fatalf("fetch run: %v\n%s", err, out)
	}
	if gotArgs == nil || !strings.Contains(strings.Join(gotArgs, " "), testHead) {
		t.Errorf("gh invoked without the head: %v", gotArgs)
	}
	rec, err := civerdict.Load(root, testHead)
	if err != nil || rec == nil {
		t.Fatalf("Load: %v, %v", rec, err)
	}
	if rec.Conclusion != "failure" || rec.RunID != "4242" {
		t.Errorf("recorded = %+v, want the fabricated CI result", rec)
	}
	if rec.ObservedAt == "" {
		t.Errorf("observed_at not stamped: %+v", rec)
	}
	if rec.Producer != civerdict.ProducerID {
		t.Errorf("producer = %q, want %q", rec.Producer, civerdict.ProducerID)
	}
}

// AC-CV-003 (REQ-CV-003): gh absent, failing, and unparseable each print one
// clear line naming the fault, exit 0, and write no record — a failed
// observation never writes a fabricated verdict.
func TestCIVerdictDegradation(t *testing.T) {
	cases := []struct {
		name   string
		runner ghRunner
		want   string
	}{
		{"gh-absent", func(args ...string) ([]byte, error) {
			return nil, exec.ErrNotFound
		}, "gh not found"},
		{"gh-fails", func(args ...string) ([]byte, error) {
			return nil, errors.New("exit status 4: auth required")
		}, "gh query failed"},
		{"gh-unparseable", func(args ...string) ([]byte, error) {
			return []byte("not json at all"), nil
		}, "unparseable"},
		{"no-run-for-head", func(args ...string) ([]byte, error) {
			return []byte(`[]`), nil
		}, "no CI run"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			out, err := runCIVerdict(t, tc.runner, root, "--head", testHead)
			if err != nil {
				t.Fatalf("exit = %v, want 0\n%s", err, out)
			}
			if !strings.Contains(out, tc.want) {
				t.Errorf("message %q does not name the fault %q", out, tc.want)
			}
			if n := len(strings.Split(strings.TrimSpace(out), "\n")); out != "" && n != 1 {
				t.Errorf("message is not one line (%d lines): %q", n, out)
			}
			if files := verdictFiles(t, root); len(files) != 0 {
				t.Errorf("degraded run wrote records: %v", files)
			}
		})
	}
}

// REQ-CV-005: the producer's source path contains no escalation checkpoint,
// hook, or detector reference — it writes evidence files and nothing else.
func TestCIVerdictNoDetectorReference(t *testing.T) {
	src, err := os.ReadFile("ci_verdict.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"escalation.", "Checkpoint", "Observe("} {
		if strings.Contains(string(src), forbidden) {
			t.Errorf("producer source references %q (REQ-CV-005)", forbidden)
		}
	}
}
