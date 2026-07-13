package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initVerifyTestRepo creates a throwaway git repo with one committed file and
// a .gitignore covering .moai/ (mirrors the real project: the snapshot store
// writes under .moai/state/, which must not invalidate its own key).
func initVerifyTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q")
	git("config", "user.email", "test@example.com")
	git("config", "user.name", "test")
	git("config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(".moai/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tracked.txt"), []byte("v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", ".gitignore", "tracked.txt")
	git("commit", "-q", "-m", "init")
	return dir
}

func runVerifyCmd(t *testing.T, args ...string) (stdout string, err error) {
	t.Helper()
	cmd := newVerifyCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return out.String(), err
}

// TestVerifyHelpListsVerbs asserts `moai verify --help` lists the record and
// check verbs (the CLI reachability contract).
func TestVerifyHelpListsVerbs(t *testing.T) {
	out, err := runVerifyCmd(t, "--help")
	if err != nil {
		t.Fatalf("verify --help: %v", err)
	}
	for _, verb := range []string{"record", "check"} {
		if !strings.Contains(out, verb) {
			t.Errorf("verify --help must list %q verb:\n%s", verb, out)
		}
	}
}

// TestVerifyRegisteredOnRoot asserts the verify command is registered in the
// root command tree (cross-file reachability — a defined-but-unregistered verb
// is inert).
func TestVerifyRegisteredOnRoot(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "verify" {
			return
		}
	}
	t.Fatal("verify command must be registered on the root command tree")
}

// TestVerifyRecordThenCheckFresh is the E2E happy path: record a check result,
// then a same-tree in-TTL check reports fresh (exit 0 at the RunE level) and
// prints snapshot path + key + matched entry for citation.
func TestVerifyRecordThenCheckFresh(t *testing.T) {
	dir := initVerifyTestRepo(t)

	out, err := runVerifyCmd(t, "record",
		"--project-root", dir,
		"--check-id", "test",
		"--command", "go test ./...",
		"--exit", "0",
		"--duration-ms", "1200")
	if err != nil {
		t.Fatalf("verify record: %v\n%s", err, out)
	}
	var rec map[string]any
	if jsonErr := json.Unmarshal([]byte(out), &rec); jsonErr != nil {
		t.Fatalf("record output must be JSON: %v\n%s", jsonErr, out)
	}
	for _, field := range []string{"snapshot", "key", "command", "exit_code"} {
		if _, okField := rec[field]; !okField {
			t.Errorf("record output missing %q: %s", field, out)
		}
	}

	out, err = runVerifyCmd(t, "check", "--key-current", "--project-root", dir)
	if err != nil {
		t.Fatalf("verify check on unchanged tree must be fresh (exit 0): %v\n%s", err, out)
	}
	for _, tok := range []string{`"fresh": true`, "snapshot", "key", "go test ./..."} {
		if !strings.Contains(out, tok) {
			t.Errorf("check output must carry %q for citation:\n%s", tok, out)
		}
	}
}

// TestVerifyCheckStaleAfterTreeChange asserts a tracked-file mutation after
// recording makes the check report stale (non-nil error → exit 1).
func TestVerifyCheckStaleAfterTreeChange(t *testing.T) {
	dir := initVerifyTestRepo(t)

	if out, err := runVerifyCmd(t, "record",
		"--project-root", dir,
		"--check-id", "test",
		"--command", "go test ./...",
		"--exit", "0"); err != nil {
		t.Fatalf("verify record: %v\n%s", err, out)
	}
	if err := os.WriteFile(filepath.Join(dir, "tracked.txt"), []byte("mutated\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runVerifyCmd(t, "check", "--key-current", "--project-root", dir)
	if err == nil {
		t.Fatalf("check after tree change must be stale (exit 1):\n%s", out)
	}
	if !strings.Contains(out, `"fresh": false`) {
		t.Errorf("stale check output must carry fresh:false:\n%s", out)
	}
}

// TestVerifyCheckMissingSnapshot asserts the first-run path (no snapshot) is
// stale — consumers fall back to plain re-execution.
func TestVerifyCheckMissingSnapshot(t *testing.T) {
	dir := initVerifyTestRepo(t)
	if _, err := runVerifyCmd(t, "check", "--key-current", "--project-root", dir); err == nil {
		t.Fatal("check with no recorded snapshot must be stale (exit 1)")
	}
}

// TestVerifyCheckByCheckID asserts --check filters to the named check id: a
// recorded id is fresh; an unrecorded id is stale.
func TestVerifyCheckByCheckID(t *testing.T) {
	dir := initVerifyTestRepo(t)
	if out, err := runVerifyCmd(t, "record",
		"--project-root", dir,
		"--check-id", "lint",
		"--command", "golangci-lint run",
		"--exit", "0"); err != nil {
		t.Fatalf("verify record: %v\n%s", err, out)
	}
	if out, err := runVerifyCmd(t, "check", "--key-current", "--check", "lint", "--project-root", dir); err != nil {
		t.Fatalf("recorded check id must be fresh: %v\n%s", err, out)
	}
	if _, err := runVerifyCmd(t, "check", "--key-current", "--check", "test", "--project-root", dir); err == nil {
		t.Fatal("unrecorded check id must be stale (exit 1)")
	}
}

// TestVerifyRecordStdin asserts `record --stdin` accepts a CheckEntry JSON
// document including a loop-verdict-shaped conditions block.
func TestVerifyRecordStdin(t *testing.T) {
	dir := initVerifyTestRepo(t)
	entry := `{"check_id":"test","command":"go test ./...","exit_code":0,"duration_ms":900,
		"conditions":{"zero_errors":true,"error_count":0,"tests_pass":true,"coverage_threshold":85,"coverage_actual":87.0,"zero_warnings":false}}`

	cmd := newVerifyCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetIn(strings.NewReader(entry))
	cmd.SetArgs([]string{"record", "--stdin", "--project-root", dir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("verify record --stdin: %v\n%s", err, out.String())
	}
	if checkOut, err := runVerifyCmd(t, "check", "--key-current", "--check", "test", "--project-root", dir); err != nil {
		t.Fatalf("stdin-recorded check must be fresh: %v\n%s", err, checkOut)
	}
}
