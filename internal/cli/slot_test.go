package cli

// slot_test.go — the `moai slot` verbs (SPEC-RESOURCE-SLOT-LEASE-001, card
// t607, M1 RED; flipped GREEN by M3).
//
// The command is located through rootCmd by NAME rather than through a
// constructor symbol, so this file compiles before M3 exists and fails on
// behaviour ("no `moai slot` command registered") instead of breaking the whole
// package's build. Because the registered command is a singleton, every run
// resets all of its flags to their defaults first — cobra keeps flag values
// across Execute calls, and a leaked --force would make a later case lie.
//
// Execution goes through rootCmd.Execute (never the package Execute, which
// would build the full dependency graph).

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// slotCLIScrubEnv pins the SPEC's isolation variables. Without the pin an
// acquire would resolve the developer's own repository and write into its
// .moai/state; without the scrub the runtime's real session id and owner pid
// would be recorded.
func slotCLIScrubEnv(t *testing.T, projectDir string) {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", projectDir)
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(projectDir))
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
	t.Setenv("MOAI_SESSION_PID", "")
}

func findSlotCmd(t *testing.T) *cobra.Command {
	t.Helper()
	for _, c := range rootCmd.Commands() {
		if c.Name() == "slot" {
			return c
		}
	}
	t.Fatalf("no `moai slot` command registered on the root command")
	return nil
}

func resetSlotFlags(cmd *cobra.Command) {
	reset := func(fs *pflag.FlagSet) {
		fs.VisitAll(func(f *pflag.Flag) {
			_ = f.Value.Set(f.DefValue)
			f.Changed = false
		})
	}
	reset(cmd.Flags())
	reset(cmd.PersistentFlags())
	for _, c := range cmd.Commands() {
		resetSlotFlags(c)
	}
}

// runSlotCLI executes `moai slot <args...>` with CLAUDE_PROJECT_DIR pinned to
// projectDir and returns combined stdout/stderr.
func runSlotCLI(t *testing.T, projectDir string, args ...string) (string, error) {
	t.Helper()
	slotCLIScrubEnv(t, projectDir)
	slot := findSlotCmd(t)
	resetSlotFlags(slot)
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs(append([]string{"slot"}, args...))
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	})
	err := rootCmd.Execute()
	return out.String(), err
}

func slotCLIRecordPath(root string) string {
	return filepath.Join(root, ".moai", "state", "slot-leases", "demo.json")
}

func readSlotCLIRecord(t *testing.T, root string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(slotCLIRecordPath(root))
	if err != nil {
		t.Fatalf("record not at %s: %v", slotCLIRecordPath(root), err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("record is not JSON: %v", err)
	}
	return raw
}

// AC-RSL-003a (CLI half) — --json carries every field, and the recorded pid is
// the owning session's, never this (acquiring) process's own.
func TestSlotCLI_AcquireJSONFields(t *testing.T) {
	root := t.TempDir()
	out, err := runSlotCLI(t, root, "acquire", "--resource", "demo", "--session", "s-1",
		"--name", "lane-a", "--command", "heavy-suite", "--max-duration", "5m", "--json")
	if err != nil {
		t.Fatalf("acquire: %v\n%s", err, out)
	}
	var got struct {
		Acquired bool           `json:"acquired"`
		Lease    map[string]any `json:"lease"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("acquire --json is not JSON (%v): %s", err, out)
	}
	if !got.Acquired {
		t.Errorf("acquired = false: %s", out)
	}
	for _, key := range []string{"resource", "session_id", "session_name", "pid", "pid_source", "command", "acquired_at", "max_duration", "expires_at"} {
		if _, ok := got.Lease[key]; !ok {
			t.Errorf("lease JSON has no %q: %s", key, out)
		}
	}
	if got.Lease["session_name"] != "lane-a" || got.Lease["command"] != "heavy-suite" || got.Lease["max_duration"] == "" {
		t.Errorf("lease JSON = %v, want name lane-a, command heavy-suite, a max_duration", got.Lease)
	}
	if pid, _ := got.Lease["pid"].(float64); int(pid) == os.Getpid() {
		t.Errorf("recorded pid %d is the acquiring process's own — it is dead the moment the verb returns; record the session owner (or 0)", os.Getpid())
	}
	if rec := readSlotCLIRecord(t, root); rec["session_id"] != "s-1" {
		t.Errorf("record session_id = %v, want s-1", rec["session_id"])
	}
}

// AC-RSL-003b — omitted --name/--command are recorded empty and shown as
// "(not given)".
func TestSlotCLI_OmittedNameAndCommand(t *testing.T) {
	root := t.TempDir()
	if out, err := runSlotCLI(t, root, "acquire", "--resource", "demo", "--session", "s-1"); err != nil {
		t.Fatalf("acquire without --name/--command must succeed: %v\n%s", err, out)
	}
	rec := readSlotCLIRecord(t, root)
	for _, key := range []string{"session_name", "command"} {
		if v, ok := rec[key]; !ok || v != "" {
			t.Errorf("record[%q] = %v (present=%v), want an empty string", key, v, ok)
		}
	}
	out, err := runSlotCLI(t, root, "status", "--resource", "demo")
	if err != nil {
		t.Fatalf("status: %v\n%s", err, out)
	}
	if n := strings.Count(out, "(not given)"); n < 2 {
		t.Errorf("status shows %d \"(not given)\" markers, want one each for the name and the command:\n%s", n, out)
	}
}

// AC-RSL-003c — status of a free resource as JSON.
func TestSlotCLI_StatusJSONFree(t *testing.T) {
	root := t.TempDir()
	out, err := runSlotCLI(t, root, "status", "--resource", "demo", "--json")
	if err != nil {
		t.Fatalf("status on a free resource: %v\n%s", err, out)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("status --json is not JSON (%v): %s", err, out)
	}
	if got["resource"] != "demo" || got["held"] != false {
		t.Errorf("status JSON = %v, want resource demo and held false", got)
	}
}

// AC-RSL-003c — the help lists the three verbs.
func TestSlotCLI_HelpListsVerbs(t *testing.T) {
	out, err := runSlotCLI(t, t.TempDir(), "--help")
	if err != nil {
		t.Fatalf("slot --help: %v", err)
	}
	for _, verb := range []string{"acquire", "status", "release"} {
		if !strings.Contains(out, verb) {
			t.Errorf("slot --help does not list %q:\n%s", verb, out)
		}
	}
}

// AC-RSL-007 (CLI half) — a forced takeover reports the displaced holder.
func TestSlotCLI_ForceReportsDisplaced(t *testing.T) {
	root := t.TempDir()
	if out, err := runSlotCLI(t, root, "acquire", "--resource", "demo", "--session", "s-1", "--name", "lane-1"); err != nil {
		t.Fatalf("first acquire: %v\n%s", err, out)
	}
	out, err := runSlotCLI(t, root, "acquire", "--resource", "demo", "--session", "s-2", "--force")
	if err != nil {
		t.Fatalf("forced acquire: %v\n%s", err, out)
	}
	if !strings.Contains(out, "displaced") || !strings.Contains(out, "s-1") {
		t.Errorf("forced acquire did not report the displaced holder:\n%s", out)
	}
	displaced, _ := readSlotCLIRecord(t, root)["displaced"].(map[string]any)
	if displaced == nil || displaced["session_id"] != "s-1" || displaced["reason"] != "force" {
		t.Errorf("record displaced = %v, want s-1 with reason force", displaced)
	}
}

// An unresolvable session id is a blocker, never a value to invent.
func TestSlotCLI_RefusesWithoutSessionID(t *testing.T) {
	root := t.TempDir()
	_, err := runSlotCLI(t, root, "acquire", "--resource", "demo")
	if err == nil {
		t.Fatal("acquire invented a holder identity")
	}
	if !strings.Contains(err.Error(), "--session") {
		t.Errorf("error does not tell the caller how to proceed: %v", err)
	}
}

// plan-audit N1 — the CLI resolves its root with the SAME function the guard
// uses (kanban.ResolveSlotLeaseRoot): a CLAUDE_PROJECT_DIR pointing into a
// linked worktree still writes the record into the primary checkout, where the
// guard reads it.
func TestSlotCLI_RootNormalizesWorktreeProjectDir(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("git not in PATH: %v", err)
	}
	parent := t.TempDir()
	primary := filepath.Join(parent, "primary")
	wt := filepath.Join(parent, "wt")
	if err := os.MkdirAll(primary, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	git := func(args ...string) {
		t.Helper()
		out, err := exec.Command("git", append([]string{"-C", primary}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	git("init", "-q")
	git("config", "user.email", "slot-cli-test@example.com")
	git("config", "user.name", "Slot CLI Test")
	git("config", "core.hooksPath", "/dev/null")
	git("commit", "-q", "--allow-empty", "-m", "seed")
	git("worktree", "add", "-q", "--detach", wt)

	if out, err := runSlotCLI(t, wt, "acquire", "--resource", "demo", "--session", "s-1"); err != nil {
		t.Fatalf("acquire with CLAUDE_PROJECT_DIR in a worktree: %v\n%s", err, out)
	}
	if _, err := os.Stat(slotCLIRecordPath(primary)); err != nil {
		t.Errorf("record not in the primary checkout (%v) — the CLI did not normalize the worktree root", err)
	}
	if _, err := os.Stat(slotCLIRecordPath(wt)); err == nil {
		t.Errorf("record written inside the worktree — the guard, which normalizes, would never read it")
	}
}
