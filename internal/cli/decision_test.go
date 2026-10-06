package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/decision"
	"github.com/modu-ai/moai-adk/internal/paths"
)

func runDecision(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := newDecisionCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errOut.String(), err
}

func decisionProject(t *testing.T) string {
	t.Helper()
	t.Setenv(paths.EnvHome, t.TempDir())
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o700); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestDecisionCmd_RecordThenReadFromTheProjectBoard(t *testing.T) {
	t.Setenv(config.EnvFactoryRole, "")
	t.Setenv(config.EnvMoaiFactoryWorker, "")
	t.Setenv(config.EnvFactoryBackend, "")
	root := decisionProject(t)
	_, stderr, err := runDecision(t, "record", "--project-root", root, "--scope", "card:t1", "--kind", "ruling",
		"--body", "one delta round", "--evidence", ".moai/reports/t1/x.md", "--ladder", "②", "--decided-by", "claude+leader")
	if err != nil {
		t.Fatalf("record: %v %s", err, stderr)
	}
	out, _, err := runDecision(t, "read", "--project-root", root, "--scope", "card:t1")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(out, "board=ok") || !strings.Contains(out, "decision record: decided_by=claude+leader") {
		t.Fatalf("read output %q lacks the status line or the record line", out)
	}
	board, err := decision.BoardPath(root)
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(board, root) {
		t.Fatalf("board %q inside the tree", board)
	}
}

func TestDecisionCmd_LaneCannotRecordButCanRead(t *testing.T) {
	root := decisionProject(t)
	t.Setenv(config.EnvFactoryRole, "")
	t.Setenv(config.EnvMoaiFactoryWorker, "")
	t.Setenv(config.EnvFactoryBackend, "")
	if _, _, err := runDecision(t, "record", "--project-root", root, "--scope", "standing", "--kind", "standing-rule",
		"--predicate", "always", "--body", "x", "--evidence", "y", "--ladder", "②", "--decided-by", "claude+leader"); err != nil {
		t.Fatalf("leader record: %v", err)
	}
	board, err := decision.BoardPath(root)
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(board)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	t.Setenv(config.EnvFactoryRole, config.FactoryRoleLane)
	_, stderr, err := runDecision(t, "record", "--project-root", root, "--scope", "standing", "--kind", "standing-rule",
		"--predicate", "always", "--body", "x", "--evidence", "y", "--ladder", "②", "--decided-by", "claude+lane")
	if err == nil {
		t.Fatalf("lane record accepted, want refusal")
	}
	if !strings.Contains(err.Error()+stderr, "lane boundary") {
		t.Fatalf("lane refusal %v does not name the lane boundary", err)
	}
	after, err := os.Stat(board)
	if err != nil {
		t.Fatal(err)
	}
	if !after.ModTime().Equal(before.ModTime()) || after.Size() != before.Size() {
		t.Fatalf("lane refusal touched the board")
	}
	if out, _, err := runDecision(t, "read", "--project-root", root); err != nil || !strings.Contains(out, "board=ok") {
		t.Fatalf("lane read: err=%v out=%q", err, out)
	}
}

func TestDecisionCmd_ReadReportsAbsentBoard(t *testing.T) {
	root := decisionProject(t)
	out, _, err := runDecision(t, "read", "--project-root", root)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(out, "board=absent") {
		t.Fatalf("absent board output %q", out)
	}
}
