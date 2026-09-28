package cli

// factory_role_refusal_m2_test.go — SPEC-ROLE-NAMING-CODE-001 M2: the CLI
// refusal surface. The legacy spellings (`-f worker`, `-f agent`, any case;
// `worker-<n>` / `agent-<n>` labels on any input path; `lead` /
// `lead-<suffix>` leader names) are refused with one error line naming the
// canonical form, a non-zero exit, and no record written (REQ-RNC-002, -003,
// -004, -005, -007, -020; AC-RNC-001, -002, -003, -005, -007).

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// TestFactoryEntryRefusesLegacyRoleTokens (AC-RNC-002, parse level): every
// letter case of the legacy role tokens is refused with an error naming the
// canonical `-f lane`.
func TestFactoryEntryRefusesLegacyRoleTokens(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		args []string
	}{
		{name: "-f worker", args: []string{"-f", "worker"}},
		{name: "-f=worker", args: []string{"-f=worker"}},
		{name: "--factory agent", args: []string{"--factory", "agent"}},
		{name: "--factory=agent", args: []string{"--factory=agent"}},
		{name: "-f WORKER", args: []string{"-f", "WORKER"}},
		{name: "-f Worker", args: []string{"-f", "Worker"}},
		{name: "-f AGENT", args: []string{"-f", "AGENT"}},
		{name: "-f Agent", args: []string{"-f", "Agent"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, err := parseLauncherEntry(c.args)
			if err == nil {
				t.Fatalf("parseLauncherEntry(%v) = nil error, want the legacy-token refusal", c.args)
			}
			msg := err.Error()
			if !strings.Contains(msg, "-f lane") {
				t.Errorf("refusal %q does not name the canonical -f lane", msg)
			}
			if n := strings.Count(msg, "\n"); n != 0 {
				t.Errorf("refusal is not one error line (%d newlines): %q", n, msg)
			}
		})
	}
}

// TestFactoryEntryRefusesLegacyLaneLabels (AC-RNC-005, parse level): a legacy
// lane label on the -f value path, any letter case, is refused with an error
// naming the canonical lane-<n>.
func TestFactoryEntryRefusesLegacyLaneLabels(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		args []string
		want string
	}{
		{name: "-f worker-2", args: []string{"-f", "worker-2"}, want: "lane-2"},
		{name: "-f Worker-4", args: []string{"-f", "Worker-4"}, want: "lane-4"},
		{name: "-f=agent-5", args: []string{"-f=agent-5"}, want: "lane-5"},
		{name: "--factory AGENT-6", args: []string{"--factory", "AGENT-6"}, want: "lane-6"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, err := parseLauncherEntry(c.args)
			if err == nil {
				t.Fatalf("parseLauncherEntry(%v) = nil error, want the legacy-label refusal", c.args)
			}
			if msg := err.Error(); !strings.Contains(msg, c.want) {
				t.Errorf("refusal %q does not name the canonical %s", msg, c.want)
			}
		})
	}
}

// TestFactoryEntryRefusesLegacyLaneNameTyped (REQ-RNC-005, parse level): a
// legacy label typed as an operator --name on any entry branch is refused
// naming the canonical lane-<n>.
func TestFactoryEntryRefusesLegacyLaneNameTyped(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		args []string
		want string
	}{
		{name: "-f --name agent-5", args: []string{"-f", "--name", "agent-5"}, want: "lane-5"},
		{name: "-f -n=worker-3", args: []string{"-f", "-n=worker-3"}, want: "lane-3"},
		{name: "-k --name worker-3", args: []string{"-k", "--name", "worker-3"}, want: "lane-3"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, err := parseLauncherEntry(c.args)
			if err == nil {
				t.Fatalf("parseLauncherEntry(%v) = nil error, want the legacy-name refusal", c.args)
			}
			if msg := err.Error(); !strings.Contains(msg, c.want) {
				t.Errorf("refusal %q does not name the canonical %s", msg, c.want)
			}
		})
	}
}

// TestLauncherEntryRefusesLegacyLeaderName (AC-RNC-007, parse level): `lead`
// and `lead-<suffix>` operator names are refused naming the canonical leader
// forms; canonical names still parse.
func TestLauncherEntryRefusesLegacyLeaderName(t *testing.T) {
	t.Parallel()

	refusals := []struct {
		name string
		args []string
		want string
	}{
		{name: "-k --name lead", args: []string{"-k", "--name", "lead"}, want: "leader"},
		{name: "-k --name lead-7", args: []string{"-k", "--name", "lead-7"}, want: "leader-7"},
		{name: "-f --name lead", args: []string{"-f", "--name", "lead"}, want: "leader"},
		{name: "-n=lead-abc123", args: []string{"-n=lead-abc123"}, want: "leader-abc123"},
	}
	for _, c := range refusals {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, err := parseLauncherEntry(c.args)
			if err == nil {
				t.Fatalf("parseLauncherEntry(%v) = nil error, want the legacy-leader-name refusal", c.args)
			}
			if msg := err.Error(); !strings.Contains(msg, c.want) {
				t.Errorf("refusal %q does not name the canonical %s", msg, c.want)
			}
		})
	}

	valid := []struct {
		name string
		args []string
	}{
		{name: "bare -k (leader, no name)", args: []string{"-k"}},
		{name: "-k --name leader-abc123 (composed run id)", args: []string{"-k", "--name", "leader-abc123"}},
		{name: "-f --name leader-r7", args: []string{"-f", "--name", "leader-r7"}},
		{name: "companion plan unaffected", args: []string{"-k", "--name", "plan"}},
	}
	for _, c := range valid {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if _, err := parseLauncherEntry(c.args); err != nil {
				t.Errorf("parseLauncherEntry(%v) = %v, want nil", c.args, err)
			}
		})
	}
}

// seedCLIWorkerRow inserts a factory workers-table row directly — the shape a
// pre-change binary wrote (run boundary M1 machinery, CLI-path seeding).
func seedCLIWorkerRow(t *testing.T, root, label string, pid int, runID string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.DB.Exec(`INSERT INTO workers(label,pid,registered_at,heartbeat_at,run_id) VALUES(?,?,?,?,?)`,
		label, pid, "2026-09-26T00:00:00Z", "2026-09-26T00:00:00Z", runID); err != nil {
		t.Fatal(err)
	}
}

// countCLIWorkerRows returns the factory workers-table row count.
func countCLIWorkerRows(t *testing.T, root string) int {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	var n int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM workers`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// TestRunCCRefusesLegacySpellingsNothingWritten (AC-RNC-002 / -005 / -007 at
// the command level): each legacy spelling is refused through runCC, the
// launch seam never fires, and no claim row is written.
func TestRunCCRefusesLegacySpellingsNothingWritten(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want []string
	}{
		{name: "-f worker", args: []string{"-f", "worker"}, want: []string{"-f lane"}},
		{name: "-f agent", args: []string{"-f", "agent"}, want: []string{"-f lane"}},
		{name: "-f WORKER", args: []string{"-f", "WORKER"}, want: []string{"-f lane"}},
		{name: "-f worker-2", args: []string{"-f", "worker-2"}, want: []string{"lane-2"}},
		{name: "-f Worker-4", args: []string{"-f", "Worker-4"}, want: []string{"lane-4"}},
		{name: "-f --name agent-5", args: []string{"-f", "--name", "agent-5"}, want: []string{"lane-5"}},
		{name: "-k --name lead", args: []string{"-k", "--name", "lead"}, want: []string{"leader"}},
		{name: "-k --name lead-7", args: []string{"-k", "--name", "lead-7"}, want: []string{"leader-7"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv(config.EnvClaudeProjectDir, root)
			t.Setenv("MOAI_HOME", t.TempDir())
			clearFactoryTestEnv(t)
			cap := installFactoryLaunchSeam(t)

			leadsPath := leaderRegistryPath(root)
			_ = os.MkdirAll(filepath.Dir(leadsPath), 0o755)
			before, _ := os.ReadFile(leadsPath)

			buf := new(bytes.Buffer)
			ccCmd.SetOut(buf)
			ccCmd.SetErr(buf)
			err := runCC(ccCmd, c.args)
			if err == nil {
				t.Fatalf("runCC(%v) = nil error, want the refusal", c.args)
			}
			msg := err.Error()
			for _, want := range c.want {
				if !strings.Contains(msg, want) {
					t.Errorf("refusal %q missing %q", msg, want)
				}
			}
			if cap.worker != "" || cap.runID != "" || len(cap.args) > 0 {
				t.Errorf("a refused launch must not reach the launcher (capture %+v)", cap)
			}
			if n := countCLIWorkerRows(t, root); n != 0 {
				t.Errorf("workers table holds %d rows after a refused launch, want 0", n)
			}
			after, _ := os.ReadFile(leadsPath)
			if !bytes.Equal(before, after) {
				t.Errorf("leads.json changed across a refused launch: %q → %q", before, after)
			}
		})
	}
}

// TestRunCCRefusesLegacyLeaderNameLeadsJSONSeeded (AC-RNC-007 byte-identity
// clause): with a live `lead` entry already in leads.json, the refusal leaves
// it byte-identical.
func TestRunCCRefusesLegacyLeaderNameLeadsJSONSeeded(t *testing.T) {
	root := t.TempDir()
	t.Setenv(config.EnvClaudeProjectDir, root)
	t.Setenv("MOAI_HOME", t.TempDir())
	clearFactoryTestEnv(t)
	cap := installFactoryLaunchSeam(t)

	leadsPath := leaderRegistryPath(root)
	seed := []byte(`{"lead":{"pid":1,"registered_at":"2026-09-01T00:00:00Z"}}`)
	if err := os.MkdirAll(filepath.Dir(leadsPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(leadsPath, seed, 0o644); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)
	if err := runCC(ccCmd, []string{"-k", "--name", "lead"}); err == nil {
		t.Fatal("runCC(-k --name lead) = nil error, want the refusal")
	} else if !strings.Contains(err.Error(), "leader") {
		t.Errorf("refusal %q does not name leader", err.Error())
	}
	after, err := os.ReadFile(leadsPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(seed, after) {
		t.Errorf("leads.json not byte-identical: %q → %q", seed, after)
	}
	if cap.worker != "" || len(cap.args) > 0 {
		t.Errorf("a refused launch must not reach the launcher (capture %+v)", cap)
	}
}

// TestRunCCFactoriesEntryWritesLane1 (AC-RNC-001): `-f lane` joins as the
// next free lane, the session is named lane-1, the factory registry row's
// label is lane-1, and the command succeeds.
func TestRunCCFactoriesEntryWritesLane1(t *testing.T) {
	root := t.TempDir()
	// SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-005: the lane join needs a git
	// working tree, so this fixture is one.
	initGitRepo(t, root)
	t.Setenv(config.EnvClaudeProjectDir, root)
	t.Setenv("MOAI_HOME", t.TempDir())
	clearFactoryTestEnv(t)
	cap := installFactoryLaunchSeam(t)

	const run = "run-m2-lane-join"
	if err := recordFactoryRunStart(root, run, kanban.BackendClaude, ""); err != nil {
		t.Fatalf("record factory run: %v", err)
	}

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)
	if err := runCC(ccCmd, []string{"-f", "lane"}); err != nil {
		t.Fatalf("runCC(-f lane): %v", err)
	}
	if cap.worker != "lane-1" {
		t.Errorf("%s at launch = %q, want lane-1", config.EnvMoaiFactoryWorker, cap.worker)
	}
	if !containsArg(cap.args, "--name", "lane-1") {
		t.Errorf("the desugared --name lane-1 must reach the launcher, got %v", cap.args)
	}
	if n := countCLIWorkerRows(t, root); n != 1 {
		t.Fatalf("workers table holds %d rows, want 1", n)
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	var label string
	if err := db.DB.QueryRow(`SELECT label FROM workers LIMIT 1`).Scan(&label); err != nil {
		t.Fatal(err)
	}
	if label != "lane-1" {
		t.Errorf("registry row label = %q, want lane-1", label)
	}
}

func containsArg(args []string, flag, value string) bool {
	for i, a := range args {
		if a == flag && i+1 < len(args) && args[i+1] == value {
			return true
		}
	}
	return false
}

// TestRunCCLiveLegacyClaimRefusedThroughCLI (AC-RNC-003, CLI path): a live
// worker-3 claim from a pre-change binary refuses the `-f lane` join with the
// retire-and-relaunch message and no lane-<n> claim is written; a dead one is
// stale and the join proceeds exactly as with no claim.
func TestRunCCLiveLegacyClaimRefusedThroughCLI(t *testing.T) {
	root := t.TempDir()
	// SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-005: the lane join needs a git
	// working tree, so this fixture is one.
	initGitRepo(t, root)
	t.Setenv(config.EnvClaudeProjectDir, root)
	t.Setenv("MOAI_HOME", t.TempDir())
	clearFactoryTestEnv(t)
	cap := installFactoryLaunchSeam(t)

	const run = "run-legacy-live"
	if err := recordFactoryRunStart(root, run, kanban.BackendClaude, ""); err != nil {
		t.Fatalf("record factory run: %v", err)
	}
	seedCLIWorkerRow(t, root, "worker-3", os.Getpid(), run)

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)
	err := runCC(ccCmd, []string{"-f", "lane"})
	if err == nil {
		t.Fatal("runCC(-f lane) with a live legacy claim = nil error, want the refusal")
	}
	msg := err.Error()
	for _, want := range []string{"worker-3", run, "moai factory runs --retire " + run} {
		if !strings.Contains(msg, want) {
			t.Errorf("refusal %q missing %q", msg, want)
		}
	}
	if cap.worker != "" || len(cap.args) > 0 {
		t.Errorf("a refused join must not reach the launcher (capture %+v)", cap)
	}
	if n := countCLIWorkerRows(t, root); n != 1 {
		t.Errorf("workers table holds %d rows after refusal, want 1 (the legacy row, untouched)", n)
	}
}

func TestRunCCDeadLegacyClaimProceedsThroughCLI(t *testing.T) {
	root := t.TempDir()
	// SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-005: the lane join needs a git
	// working tree, so this fixture is one.
	initGitRepo(t, root)
	t.Setenv(config.EnvClaudeProjectDir, root)
	t.Setenv("MOAI_HOME", t.TempDir())
	clearFactoryTestEnv(t)
	cap := installFactoryLaunchSeam(t)

	const run = "run-legacy-dead"
	if err := recordFactoryRunStart(root, run, kanban.BackendClaude, ""); err != nil {
		t.Fatalf("record factory run: %v", err)
	}
	seedCLIWorkerRow(t, root, "worker-3", 999999999, run) // dead pid

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)
	if err := runCC(ccCmd, []string{"-f", "lane"}); err != nil {
		t.Fatalf("runCC(-f lane) with a dead legacy claim: %v", err)
	}
	if cap.worker != "lane-1" {
		t.Errorf("%s at launch = %q, want lane-1 (dead legacy pruned as stale)", config.EnvMoaiFactoryWorker, cap.worker)
	}
}

// TestFactoryFlagUsageErrorVocabulary (AC-RNC-016, usage-error text): the -f
// usage error contains lane and leader, no worker, and no \blead\b match.
func TestFactoryFlagUsageErrorVocabulary(t *testing.T) {
	t.Parallel()

	_, err := parseFactoryFlag([]string{"-f", "bogus-value"})
	if err == nil {
		t.Fatal("want the usage error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "lane") || !strings.Contains(msg, "leader") {
		t.Errorf("usage error %q must contain lane and leader", msg)
	}
	if strings.Contains(msg, "worker") {
		t.Errorf("usage error %q still mentions worker", msg)
	}
	if re := regexp.MustCompile(`(?i)\blead\b`); re.MatchString(msg) {
		t.Errorf("usage error %q has a \\blead\\b match", msg)
	}
}

// TestLauncherHelpLaneVocabulary (AC-RNC-016, help text): the cc and glm help
// surfaces teach -f lane / -f lane-<n> and neither -f worker nor -f agent.
func TestLauncherHelpLaneVocabulary(t *testing.T) {
	t.Parallel()

	for name, text := range map[string]string{
		"cc":  ccCmd.Use + "\n" + ccCmd.Long,
		"glm": glmCmd.Use + "\n" + glmCmd.Long,
	} {
		for _, want := range []string{"-f lane", "-f lane-<n>"} {
			if !strings.Contains(text, want) {
				t.Errorf("%s help missing %q", name, want)
			}
		}
		for _, banned := range []string{"-f worker", "-f agent"} {
			if strings.Contains(text, banned) {
				t.Errorf("%s help still advertises %q", name, banned)
			}
		}
	}
}

// TestTodoNextHelpLeaderAndLanePromotion (AC-RNC-021, help text): the pick
// help names both the operator's pick through the leader and a lane's
// self-dispatch, and contains no `lead session`.
func TestTodoNextHelpLeaderAndLanePromotion(t *testing.T) {
	t.Parallel()

	text := newTodoNextCmd().Long + "\n" + newTodoNextCmd().Short
	if !strings.Contains(text, "leader") {
		t.Errorf("todo next help %q does not name the leader pick path", text)
	}
	if !strings.Contains(text, "self-dispatch") {
		t.Errorf("todo next help %q does not name the lane self-dispatch path", text)
	}
	if strings.Contains(text, "lead session") {
		t.Errorf("todo next help %q still says `lead session`", text)
	}
}
