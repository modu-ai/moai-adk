package cli

// integration_test.go — the `moai integration` verbs (card t194).
//
// The command layer is thin over internal/kanban, so these tests cover the
// parts that live only here: the flag plumbing, the refusal to invent a holder
// identity, and the round trip a lane actually performs.

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/spf13/cobra"
)

// runIntegration drives one subcommand with CLAUDE_PROJECT_DIR pinned to a
// throwaway root, and returns its stdout.
//
// The env pin matters: integrationLockRoot() prefers git-common-dir, and this
// test binary runs inside the repository, so without the pin an acquire would
// write into the developer's own .moai/state.
func runIntegration(t *testing.T, root string, args ...string) (string, error) {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(root))

	cmd := newIntegrationCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	// Cobra resolves subcommands from SetArgs; silence usage noise so a
	// failure surfaces the error rather than the help text.
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	walk(cmd, func(c *cobra.Command) { c.SilenceUsage = true; c.SilenceErrors = true })
	err := cmd.Execute()
	return out.String(), err
}

func walk(cmd *cobra.Command, fn func(*cobra.Command)) {
	fn(cmd)
	for _, c := range cmd.Commands() {
		walk(c, fn)
	}
}

// A free window reports as free rather than erroring.
func TestIntegrationStatus_FreeWindow(t *testing.T) {
	out, err := runIntegration(t, t.TempDir(), "status")
	if err != nil {
		t.Fatalf("status on a free window errored: %v", err)
	}
	if !strings.Contains(out, "free") {
		t.Errorf("status did not report a free window: %s", out)
	}
}

// The round trip a lane performs: acquire, see itself as holder, release.
func TestIntegration_AcquireStatusRelease(t *testing.T) {
	root := t.TempDir()

	if _, err := runIntegration(t, root, "acquire", "--session", "sess-lane8", "--name", "lane-8", "--branch", "release/v9.9.9"); err != nil {
		t.Fatalf("acquire: %v", err)
	}

	out, err := runIntegration(t, root, "status", "--json")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	var status struct {
		Held bool                   `json:"held"`
		Lock kanban.IntegrationLock `json:"lock"`
	}
	if err := json.Unmarshal([]byte(out), &status); err != nil {
		t.Fatalf("status --json is not valid JSON (%v): %s", err, out)
	}
	if !status.Held {
		t.Fatalf("window reads free right after acquire: %s", out)
	}
	if status.Lock.SessionID != "sess-lane8" || status.Lock.SessionName != "lane-8" {
		t.Errorf("holder identity not recorded: %+v", status.Lock)
	}
	if status.Lock.Branch != "release/v9.9.9" {
		t.Errorf("branch = %q, want release/v9.9.9", status.Lock.Branch)
	}

	if _, err := runIntegration(t, root, "release", "--session", "sess-lane8"); err != nil {
		t.Fatalf("release: %v", err)
	}
	out, err = runIntegration(t, root, "status")
	if err != nil {
		t.Fatalf("status after release: %v", err)
	}
	if !strings.Contains(out, "free") {
		t.Errorf("window not free after release: %s", out)
	}
}

// A second lane is refused, and the refusal names the holder so the operator
// knows whom to ask.
func TestIntegrationAcquire_RefusesASecondLane(t *testing.T) {
	root := t.TempDir()
	if _, err := runIntegration(t, root, "acquire", "--session", "sess-lane8", "--name", "lane-8"); err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	_, err := runIntegration(t, root, "acquire", "--session", "sess-lane5")
	if err == nil {
		t.Fatal("second lane acquired a held window")
	}
	if !strings.Contains(err.Error(), "lane-8") {
		t.Errorf("refusal does not name the holder: %v", err)
	}
}

// --force takes the window over and says what it displaced. Reporting the
// displacement is the point: the previous lane may still be running.
func TestIntegrationAcquire_ForceReportsWhatItDisplaced(t *testing.T) {
	root := t.TempDir()
	if _, err := runIntegration(t, root, "acquire", "--session", "sess-lane8", "--name", "lane-8"); err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	out, err := runIntegration(t, root, "acquire", "--session", "sess-lane5", "--force")
	if err != nil {
		t.Fatalf("force acquire: %v", err)
	}
	if !strings.Contains(out, "displaced") || !strings.Contains(out, "sess-lane8") {
		t.Errorf("force did not report the displaced holder: %s", out)
	}
}

// An unresolvable session id is a blocker, not a value to invent: a lock whose
// holder is made up can be neither released by its holder nor recognized by
// the guard.
func TestIntegrationAcquire_RefusesWithoutASessionID(t *testing.T) {
	_, err := runIntegration(t, t.TempDir(), "acquire")
	if err == nil {
		t.Fatal("acquire invented a holder identity")
	}
	if !strings.Contains(err.Error(), "--session") {
		t.Errorf("error does not tell the caller how to proceed: %v", err)
	}
}

// Releasing a window nobody holds is reported rather than silently accepted.
func TestIntegrationRelease_EmptyIsReported(t *testing.T) {
	_, err := runIntegration(t, t.TempDir(), "release", "--session", "sess-lane8")
	if err == nil {
		t.Fatal("releasing an unheld window succeeded")
	}
	if !kanban.IsIntegrationLockNotHeld(err) {
		t.Errorf("error is not the not-held sentinel: %v", err)
	}
}
