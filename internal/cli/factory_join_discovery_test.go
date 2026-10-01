package cli

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/discovery"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// The lane-join discovery tests (SPEC-FACTORY-LANE-JOIN-SOCKET-001 M3): the
// join gate's NO_ACTIVE_FACTORY branch falls back to verified leader
// discovery, resumes the one verified leader's run through the dedicated
// writer, and re-enters the SAME gate. The discovery seam is injected — the
// classifier matrix lives in internal/discovery; these tests pin what the
// gate does with a verdict, under the REQ-012 sandbox (project directory
// outside this repository's worktree set, isolated HOME / MOAI_HOME).

// discoveryTestRoot is the REQ-012 sandbox for launcher-level tests.
func discoveryTestRoot(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "join-project")
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}
	// REQ-SD-005 (card t1240, absorbed from develop): a lane join requires
	// the project to be a git working tree — the fixture satisfies the
	// precondition so the tests reach the join gate's own verdicts.
	if err := exec.Command("git", "init", "-q", root).Run(); err != nil {
		t.Fatalf("git init fixture project: %v", err)
	}
	t.Setenv(config.EnvClaudeProjectDir, root)
	t.Setenv("MOAI_HOME", t.TempDir())
	return root
}

// stageDiscoveredLeaders swaps the discovery seam for a fake returning the
// staged verdict, and hands back the targets it was asked about.
func stageDiscoveredLeaders(t *testing.T, leaders []discovery.VerifiedLeader) *[]string {
	t.Helper()
	orig := discoverFactoryLeader
	asked := &[]string{}
	t.Cleanup(func() { discoverFactoryLeader = orig })
	discoverFactoryLeader = func(_ context.Context, _ string, target string) ([]discovery.VerifiedLeader, error) {
		*asked = append(*asked, target)
		return leaders, nil
	}
	return asked
}

// verifiedTestLeader is a verdict whose owner identity is deliberately NOT
// this test process's: the resume must stamp the VERDICT's identity, so an
// implementation that stamps the calling lane (the AC-007 mutant — reusing
// recordFactoryRunStart) fails the owner assertion instead of passing by
// coincidence.
func verifiedTestLeader(runID string) discovery.VerifiedLeader {
	return discovery.VerifiedLeader{
		RunID:        runID,
		PID:          424242,
		ProcessStart: "1700000000.004242",
		Name:         "leader",
		Basis:        `{"liveness":"pid+fingerprint","label":"leader","membership":"cwd","run_id_source":"env"}`,
	}
}

// runRowOwner reads one run row's status and owner stamp.
func runRowOwner(t *testing.T, root, runID string) (status string, pid int, start string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	err = db.DB.QueryRow(`SELECT status,lead_pid,lead_process_start FROM runs WHERE run_id=?`, runID).Scan(&status, &pid, &start)
	if err != nil {
		t.Fatalf("read run row %s: %v", runID, err)
	}
	return status, pid, start
}

// runRowCount counts the project's run rows.
func runRowCount(t *testing.T, root string) int {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	var n int
	if err := db.DB.QueryRow(`SELECT count(*) FROM runs`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// AC-001 through the cc launcher: zero active runs, one verified leader — the
// join succeeds on the leader's own run identity, the lead name is exported
// for the child, and the run row is active carrying the VERIFIED leader
// identity as its owner stamp (AC-006's owner-stamp half at the gate level).
func TestCCFactoryLaneJoinsDiscoveredLeader(t *testing.T) {
	root := discoveryTestRoot(t)
	clearFactoryTestEnv(t)
	c := installFactoryLaunchSeam(t)
	asked := stageDiscoveredLeaders(t, []discovery.VerifiedLeader{verifiedTestLeader("runlead01")})

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)
	if err := runCC(ccCmd, []string{"-f", "lane"}); err != nil {
		t.Fatalf("runCC(-f lane) with one verified leader: %v", err)
	}
	if len(*asked) != 1 || (*asked)[0] != "leader" {
		t.Errorf("discovery asked = %v, want exactly [leader]", *asked)
	}
	if c.runID != "runlead01" {
		t.Errorf("%s at launch = %q, want the leader's own run id", config.EnvMoaiKanbanID, c.runID)
	}
	if c.leadName != "leader" {
		t.Errorf("%s at launch = %q, want the verified leader's name (AC-012)", config.EnvMoaiKanbanLeadName, c.leadName)
	}
	status, pid, start := runRowOwner(t, root, "runlead01")
	if status != "active" {
		t.Errorf("resumed row status = %q, want active", status)
	}
	if pid != 424242 || start != "1700000000.004242" {
		t.Errorf("resumed row owner = (%d,%q), want the verdict's verified identity (424242,1700000000.004242) — the calling process's identity here is the AC-007 caller-stamp mutant", pid, start)
	}
}

// The same join through the glm launcher (AC-013's cc/glm legs share the
// behavior; the codex twin is pinned at source level by the parity assert).
func TestGLMFactoryLaneJoinsDiscoveredLeader(t *testing.T) {
	root := discoveryTestRoot(t)
	clearFactoryTestEnv(t)
	c := installFactoryLaunchSeam(t)
	stageDiscoveredLeaders(t, []discovery.VerifiedLeader{verifiedTestLeader("runlead02")})

	buf := new(bytes.Buffer)
	glmCmd.SetOut(buf)
	glmCmd.SetErr(buf)
	if err := runGLM(glmCmd, []string{"-f", "lane"}); err != nil {
		t.Fatalf("runGLM(-f lane) with one verified leader: %v", err)
	}
	if c.runID != "runlead02" || c.leadName != "leader" {
		t.Errorf("launch env = (run %q, lead %q), want (runlead02, leader)", c.runID, c.leadName)
	}
	if status, _, _ := runRowOwner(t, root, "runlead02"); status != "active" {
		t.Errorf("resumed row status = %q, want active", status)
	}
}

// AC-002 — zero verified leaders: the original refusal stands and no run row
// is created or modified.
func TestFactoryLaneJoinZeroLeadersRefuses(t *testing.T) {
	root := discoveryTestRoot(t)
	clearFactoryTestEnv(t)
	installFactoryLaunchSeam(t)
	asked := stageDiscoveredLeaders(t, nil)

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)
	err := runCC(ccCmd, []string{"-f", "lane"})
	if err == nil || !strings.Contains(err.Error(), "NO_ACTIVE_FACTORY") {
		t.Fatalf("runCC(-f lane) with zero leaders = %v, want NO_ACTIVE_FACTORY", err)
	}
	if len(*asked) != 1 {
		t.Errorf("discovery asked %d time(s), want exactly one probe before refusing (REQ-001)", len(*asked))
	}
	if n := runRowCount(t, root); n != 0 {
		t.Errorf("refused join left %d run row(s); nothing is created or modified on refusal (AC-002)", n)
	}
}

// AC-003 — an explicit --factory-run names a decision, not an absence:
// discovery is not attempted and the run is not resumed.
func TestFactoryLaneJoinExplicitRunSkipsDiscovery(t *testing.T) {
	root := discoveryTestRoot(t)
	clearFactoryTestEnv(t)
	installFactoryLaunchSeam(t)

	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.RecordRun(context.Background(), homestate.FactoryRun{RunID: "runold01", Backend: "glm", LeadPID: 111, LeadProcessStart: "111.000001"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB.Exec(`UPDATE runs SET status='retired' WHERE run_id='runold01'`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	asked := stageDiscoveredLeaders(t, []discovery.VerifiedLeader{verifiedTestLeader("runlead03")})

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)
	err = runCC(ccCmd, []string{"-f", "lane", "--factory-run", "runold01"})
	if err == nil || !strings.Contains(err.Error(), "NO_ACTIVE_FACTORY") {
		t.Fatalf("explicit non-active run join = %v, want NO_ACTIVE_FACTORY (REQ-006)", err)
	}
	if len(*asked) != 0 {
		t.Errorf("discovery asked %d time(s) under an explicit selection; an explicit run id skips discovery (AC-003)", len(*asked))
	}
	if n := runRowCount(t, root); n != 1 {
		t.Errorf("run rows = %d, want only the retired fixture row (the leader's run was not resumed)", n)
	}
}

// AC-004 — two verified leaders fail closed naming both, and nothing is
// mutated.
func TestFactoryLaneJoinMultiLeaderFailsClosed(t *testing.T) {
	root := discoveryTestRoot(t)
	clearFactoryTestEnv(t)
	installFactoryLaunchSeam(t)
	both := []discovery.VerifiedLeader{
		{RunID: "runaaaa1", PID: 4242, ProcessStart: "1700000000.4242", Name: "leader", Basis: "{}"},
		{RunID: "runbbbb2", PID: 4243, ProcessStart: "1700000000.4243", Name: "leader", Basis: "{}"},
	}
	stageDiscoveredLeaders(t, both)

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)
	err := runCC(ccCmd, []string{"-f", "lane"})
	if err == nil || !strings.Contains(err.Error(), "AMBIGUOUS_FACTORY_LEADER") {
		t.Fatalf("two-verified-leader join = %v, want the fail-closed leader ambiguity refusal (REQ-005)", err)
	}
	for _, want := range []string{"runaaaa1", "runbbbb2"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal %q does not name candidate %s", err, want)
		}
	}
	if n := runRowCount(t, root); n != 0 {
		t.Errorf("fail-closed refusal left %d run row(s); no run state is mutated on that refusal", n)
	}
}

// The two-active-runs ambiguity is the RESOLVER's answer and stays untouched:
// discovery never fires on it (REQ-001 names NO_ACTIVE_FACTORY as the only
// discovery branch).
func TestFactoryLaneJoinAmbiguousRunsSkipDiscovery(t *testing.T) {
	root := discoveryTestRoot(t)
	clearFactoryTestEnv(t)
	installFactoryLaunchSeam(t)
	// Two active rows whose owners are live (this process) survive
	// reconciliation, so the resolver answers AMBIGUOUS_FACTORY.
	if err := recordFactoryRunStart(root, "runamb01", "glm", "", homestate.LaneCapacityDerived); err != nil {
		t.Fatal(err)
	}
	if err := recordFactoryRunStart(root, "runamb02", "glm", "", homestate.LaneCapacityDerived); err != nil {
		t.Fatal(err)
	}
	asked := stageDiscoveredLeaders(t, []discovery.VerifiedLeader{verifiedTestLeader("runlead04")})

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)
	err := runCC(ccCmd, []string{"-f", "lane"})
	if err == nil || !strings.Contains(err.Error(), "AMBIGUOUS_FACTORY") {
		t.Fatalf("two-active-runs join = %v, want AMBIGUOUS_FACTORY (unchanged contract)", err)
	}
	if len(*asked) != 0 {
		t.Errorf("discovery asked %d time(s) on AMBIGUOUS_FACTORY; only NO_ACTIVE_FACTORY discovers", len(*asked))
	}
}

// AC-009 — the --leader target reaches discovery: the default targets the
// canonical leader label, an explicit value targets that leader, and the
// verified leader's own name is what the child receives.
func TestFactoryLaneJoinLeadTargeting(t *testing.T) {
	t.Run("default targets the canonical leader label", func(t *testing.T) {
		discoveryTestRoot(t)
		clearFactoryTestEnv(t)
		c := installFactoryLaunchSeam(t)
		asked := stageDiscoveredLeaders(t, []discovery.VerifiedLeader{verifiedTestLeader("runlead05")})

		if err := runCC(ccCmd, []string{"-f", "lane"}); err != nil {
			t.Fatalf("default-target join: %v", err)
		}
		if len(*asked) != 1 || (*asked)[0] != "leader" {
			t.Errorf("discovery asked = %v, want [leader] (kanban.LeaderLabel default)", *asked)
		}
		if c.leadName != "leader" {
			t.Errorf("lead name at launch = %q, want leader", c.leadName)
		}
	})

	t.Run("explicit --leader targets that leader", func(t *testing.T) {
		discoveryTestRoot(t)
		clearFactoryTestEnv(t)
		c := installFactoryLaunchSeam(t)
		asked := stageDiscoveredLeaders(t, []discovery.VerifiedLeader{{
			RunID: "runlead06", PID: os.Getpid(), ProcessStart: homestate.CurrentProcessFingerprint(),
			Name: "leader-2", Basis: "{}",
		}})

		if err := runCC(ccCmd, []string{"-f", "lane", "-l", "leader-2"}); err != nil {
			t.Fatalf("--leader leader-2 join: %v", err)
		}
		if len(*asked) != 1 || (*asked)[0] != "leader-2" {
			t.Errorf("discovery asked = %v, want [leader-2]", *asked)
		}
		if c.leadName != "leader-2" {
			t.Errorf("lead name at launch = %q, want leader-2 (the verified leader's own name)", c.leadName)
		}
	})

	t.Run("unmatched label refuses (AC-011)", func(t *testing.T) {
		discoveryTestRoot(t)
		clearFactoryTestEnv(t)
		installFactoryLaunchSeam(t)
		stageDiscoveredLeaders(t, nil)

		if err := runCC(ccCmd, []string{"-f", "lane", "--leader", "nobody"}); err == nil || !strings.Contains(err.Error(), "NO_ACTIVE_FACTORY") {
			t.Fatalf("--leader nobody join = %v, want NO_ACTIVE_FACTORY", err)
		}
	})
}

// AC-010 — the legacy spelling is refused at parse time with the
// canonical-form error, exactly the --name refusal's shape.
func TestFactoryLeadFlagLegacyRefused(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"-f", "lane", "--leader", "lead"}, "legacy leader spelling"},
		{[]string{"-f", "lane", "-l", "lead-2"}, "legacy leader spelling"},
		{[]string{"-f", "lane", "--leader=lead"}, "legacy leader spelling"},
	} {
		_, err := parseLauncherEntry(tc.args)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("parseLauncherEntry(%v) = %v, want the canonical-form legacy refusal", tc.args, err)
			continue
		}
		if !strings.Contains(err.Error(), "--leader") {
			t.Errorf("refusal for %v = %q, should name the --leader flag", tc.args, err)
		}
	}
}

// REQ-008's surface gates: --leader is lane-only, and --leader + --factory-run is
// a conflict (the two name different selectors).
func TestFactoryLeadFlagSurfaceGates(t *testing.T) {
	t.Run("leader entry carrying --leader is an error", func(t *testing.T) {
		_, err := parseLauncherEntry([]string{"-f", "--leader", "leader"})
		if err == nil || !strings.Contains(err.Error(), "--leader") {
			t.Errorf("bare -f with --leader = %v, want an error naming the lane-only surface", err)
		}
	})
	t.Run("--leader with --factory-run is an error", func(t *testing.T) {
		_, err := parseLauncherEntry([]string{"-f", "lane", "--factory-run", "runx0001", "--leader", "leader"})
		if err == nil || !strings.Contains(err.Error(), "--factory-run") {
			t.Errorf("--leader with --factory-run = %v, want a selector-conflict error", err)
		}
	})
	t.Run("short and joined forms parse", func(t *testing.T) {
		for _, tc := range []struct {
			args []string
			want string
		}{
			{[]string{"-f", "lane", "-l", "leader"}, "leader"},
			{[]string{"-f", "lane", "--leader", "leader-2"}, "leader-2"},
			{[]string{"-f", "lane-3", "--leader=leader"}, "leader"},
		} {
			entry, err := parseLauncherEntry(tc.args)
			if err != nil {
				t.Errorf("parseLauncherEntry(%v): %v", tc.args, err)
				continue
			}
			if entry.FactoryLead != tc.want {
				t.Errorf("parseLauncherEntry(%v).FactoryLead = %q, want %q", tc.args, entry.FactoryLead, tc.want)
			}
		}
	})
	t.Run("absent --leader leaves the field empty", func(t *testing.T) {
		entry, err := parseLauncherEntry([]string{"-f", "lane"})
		if err != nil {
			t.Fatal(err)
		}
		if entry.FactoryLead != "" {
			t.Errorf("FactoryLead = %q, want empty (the gate resolves the default)", entry.FactoryLead)
		}
	})
}

// AC-013's codex leg: the codex lane twin inherits the discovery behavior
// through the SAME join function — zero active runs, one verified leader, the
// join lands on the leader's run with the lead name exported.
func TestCodexFactoryLaneJoinsDiscoveredLeader(t *testing.T) {
	root := discoveryTestRoot(t)
	clearFactoryTestEnv(t)
	stageDiscoveredLeaders(t, []discovery.VerifiedLeader{verifiedTestLeader("runcodex1")})

	entry := factoryFlagParse{Enabled: true, LaneRole: true}
	restore, err := enterCodexFactory(root, entry, nil)
	if err != nil {
		t.Fatalf("enterCodexFactory lane with one verified leader: %v", err)
	}
	if got := os.Getenv(config.EnvMoaiKanbanID); got != "runcodex1" {
		t.Errorf("%s after codex lane join = %q, want the leader's run runcodex1", config.EnvMoaiKanbanID, got)
	}
	if got := os.Getenv(config.EnvMoaiKanbanLeadName); got != "leader" {
		t.Errorf("%s after codex lane join = %q, want leader (REQ-009 on the codex twin)", config.EnvMoaiKanbanLeadName, got)
	}
	restore()

	if status, _, _ := runRowOwner(t, root, "runcodex1"); status != "active" {
		t.Errorf("resumed row status = %q, want active", status)
	}
}

// The codex parse carries the same --leader surface as cc/glm (REQ-008 mirror
// parity).
func TestCodexFactoryLeadFlagSurface(t *testing.T) {
	rest, entry, err := parseCodexFactoryEntry([]string{"-f", "lane", "--leader", "leader"})
	if err != nil {
		t.Fatalf("parseCodexFactoryEntry(-f lane --leader leader): %v", err)
	}
	if entry.Lead != "leader" {
		t.Errorf("entry.Lead = %q, want leader", entry.Lead)
	}
	if len(rest) != 0 {
		t.Errorf("rest = %v, want the flags consumed", rest)
	}
	if _, _, err := parseCodexFactoryEntry([]string{"-f", "lane", "--leader", "lead"}); err == nil || !strings.Contains(err.Error(), "legacy leader spelling") {
		t.Errorf("codex legacy --leader = %v, want the canonical-form refusal", err)
	}
	if _, _, err := parseCodexFactoryEntry([]string{"-f", "--leader", "leader"}); err == nil || !strings.Contains(err.Error(), "lane") {
		t.Errorf("codex leader-entry --leader = %v, want the lane-only surface error", err)
	}
	if _, _, err := parseCodexFactoryEntry([]string{"-f", "lane", "--factory-run", "runx0001", "--leader", "leader"}); err == nil || !strings.Contains(err.Error(), "--factory-run") {
		t.Errorf("codex --leader with --factory-run = %v, want the selector-conflict error", err)
	}
}

// AC-013's source-level assert: the discovery+resume path lives in the shared
// join point and NO launcher carries a private copy. The test binary's cwd is
// the package directory, so the launcher sources read directly.
func TestFactoryLaneJoinMirrorParity(t *testing.T) {
	gate, err := os.ReadFile("factory.go")
	if err != nil {
		t.Fatalf("read factory.go: %v", err)
	}
	gateSrc := string(gate)
	for _, want := range []string{
		"func enterFactoryLaneRun(",
		"discoverFactoryLeader(",
		"ResumeRun(",
		"enterSelectedFactoryRun(root, \"\", true, timing)", // the re-entry is the SAME gate
	} {
		if !strings.Contains(gateSrc, want) {
			t.Errorf("factory.go (the shared join point) does not carry %q", want)
		}
	}
	for _, f := range []string{"cc.go", "glm.go", "codex_factory.go"} {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		src := string(raw)
		if !strings.Contains(src, "enterFactoryLaneRun(") {
			t.Errorf("%s does not route its lane join through the shared enterFactoryLaneRun (AC-013)", f)
		}
		for _, forbidden := range []string{"DiscoverLeader(", "ResumeRun(", "discoverFactoryLeader"} {
			if strings.Contains(src, forbidden) {
				t.Errorf("%s carries a launcher-private discovery copy (%s) — REQ-010 forbids it", f, forbidden)
			}
		}
	}
}
