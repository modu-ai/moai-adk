package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/oklog/ulid/v2"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestCodexFactoryLeaderFreshIdentity(t *testing.T) {
	for _, flag := range []string{"-f", "--factory"} {
		t.Run(flag, func(t *testing.T) {
			root, _ := fcFixture(t)
			t.Chdir(root)
			netScrubLaneEnv(t)
			cap := withCodexLaunchCapture(t)
			withCodexProjectRoot(t, root)
			if err := recordFactoryRunStart(root, "outer-run", factory.BackendClaude, "", 0); err != nil {
				t.Fatal(err)
			}
			t.Setenv(config.EnvFactoryRunID, "outer-run")
			t.Setenv(config.EnvMoaiFactoryWorker, "lane-9")
			t.Setenv(config.EnvFactoryRole, config.FactoryRoleLane)
			t.Setenv(config.EnvFactoryBackend, factory.BackendClaude)
			before := factoryMarkerEnv()
			_, stderr, err := runCodexCmd(t, flag)
			if err != nil {
				t.Fatalf("leader: %v stderr=%s", err, stderr)
			}
			rec := codexOnlyRecord(t, cap)
			env := sdEnvOf(t, rec.Env)
			id := env[config.EnvFactoryRunID]
			if id == "" || id == "outer-run" {
				t.Fatalf("fresh run id=%q", id)
			}
			if len(id) != 26 || id != strings.ToLower(id) {
				t.Fatalf("run id must be a lowercase ULID: %q", id)
			}
			if _, err := ulid.ParseStrict(id); err != nil {
				t.Fatalf("invalid ULID %q: %v", id, err)
			}
			if _, ok := factory.SplitLeaderLabel(factory.LeaderLabel() + "-" + id); !ok {
				t.Fatalf("run id incompatible with leader grammar: %q", id)
			}
			if env[config.EnvFactoryBackend] != BackendCodex || env[config.EnvFactoryRole] != factory.RoleLeader || env[config.EnvMoaiFactoryWorker] != "" {
				t.Fatalf("leader identity=%v", env)
			}
			if !reflect.DeepEqual(before, factoryMarkerEnv()) {
				t.Fatal("launcher did not restore outer identity")
			}
			db := fcOpen(t, root)
			var backend, status string
			if err := db.DB.QueryRowContext(context.Background(), "SELECT lead_backend,status FROM runs WHERE run_id=?", id).Scan(&backend, &status); err != nil || backend != BackendCodex || status != "active" {
				t.Fatalf("backend=%q status=%q err=%v", backend, status, err)
			}
			var owner int
			if err := db.DB.QueryRow("SELECT lead_backend,lead_pid FROM runs WHERE run_id='outer-run'").Scan(&backend, &owner); err != nil || backend != factory.BackendClaude || owner != os.Getpid() {
				t.Fatalf("outer backend=%q owner=%d err=%v", backend, owner, err)
			}
			values := codexShellEnvSetValues(rec.Argv)
			for _, key := range []string{config.EnvFactoryRunID, config.EnvFactoryBackend, config.EnvFactoryRole, config.EnvMoaiFactoryWorker} {
				if values[key] != `"`+env[key]+`"` {
					t.Errorf("shell %s=%s child=%q", key, values[key], env[key])
				}
			}
		})
	}
}

func TestCodexFactoryLeaderMalformedHasNoEffects(t *testing.T) {
	for _, args := range [][]string{{"-f", "2"}, {"-f=2"}, {"--factory=2"}, {"-f", "worker"}, {"-f", "agent"}, {"-f", "lane-2"}, {"-f", "-l"}, {"-f", "-f"}, {"-f", "--factory-run", "x"}, {"-f", "app"}, {"-f", "status"}} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			root, _ := fcFixture(t)
			t.Chdir(root)
			netScrubLaneEnv(t)
			cap := withCodexLaunchCapture(t)
			withCodexProjectRoot(t, root)
			before := factoryMarkerEnv()
			_, _, err := runCodexCmd(t, args...)
			if err == nil || cap.count() != 0 {
				t.Fatalf("args=%v err=%v launches=%d", args, err, cap.count())
			}
			if launcherRunCount(t, root) != 0 || !reflect.DeepEqual(before, factoryMarkerEnv()) {
				t.Fatal("refusal mutated factory state")
			}
		})
	}
}

func TestCodexFactoryLeaderModelDefaultsAndShellIdentity(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		t.Run(map[bool]string{false: "defaults", true: "explicit"}[explicit], func(t *testing.T) {
			root, _ := fcFixture(t)
			t.Chdir(root)
			netScrubLaneEnv(t)
			cap := withCodexLaunchCapture(t)
			withCodexProjectRoot(t, root)
			codexSessionDefaultArgsFn = codexSessionDefaultArgs
			args := []string{"-f"}
			if explicit {
				args = append(args, "--", "--model", "custom-model", "-c", `model_reasoning_effort="low"`, "-c", `shell_environment_policy.set.`+config.EnvFactoryRunID+`="foreign"`, "-c", `shell_environment_policy.set.MOAI_FACTORY_WORKER="lane-9"`)
			}
			if _, _, err := runCodexCmd(t, args...); err != nil {
				t.Fatal(err)
			}
			rec := codexOnlyRecord(t, cap)
			if explicit {
				if !slices.Contains(rec.Argv, "custom-model") || slices.Contains(rec.Argv, config.CodexModelSol) {
					t.Fatalf("explicit model argv=%q", rec.Argv)
				}
			} else {
				if !slices.Contains(rec.Argv, config.CodexModelSol) || !slices.Contains(rec.Argv, `model_reasoning_effort="high"`) {
					t.Fatalf("default argv=%q", rec.Argv)
				}
			}
			values := codexShellEnvSetValues(rec.Argv)
			if values[config.EnvFactoryRunID] != `"`+launchEnvValue(rec.Env, config.EnvFactoryRunID)+`"` || values[config.EnvMoaiFactoryWorker] != `""` {
				t.Fatalf("shell identity=%v", values)
			}
		})
	}
}

func TestCodexFactoryLeaderAnchorPrecedesMutation(t *testing.T) {
	fx := newLeaderFixture(t)
	t.Chdir(fx.card)
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, fx.card)
	before := factoryMarkerEnv()
	if _, _, err := runCodexCmd(t, "-f"); err == nil || !strings.Contains(err.Error(), "card") {
		t.Fatalf("anchor refusal=%v", err)
	}
	if cap.count() != 0 || launcherRunCount(t, fx.card) != 0 || !reflect.DeepEqual(before, factoryMarkerEnv()) {
		t.Fatal("anchor refusal mutated state")
	}
}

func TestCodexFactoryLeaderInheritedManagedDoesNotDivert(t *testing.T) {
	root, _ := fcFixture(t)
	t.Chdir(root)
	netScrubLaneEnv(t)
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, root)
	t.Setenv(config.EnvMoaiFactoryManaged, "1")
	t.Setenv(config.EnvFactoryRunID, "outer")
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-9")
	if _, _, err := runCodexCmd(t, "-f"); err != nil {
		t.Fatal(err)
	}
	if cap.count() != 1 {
		t.Fatalf("leader launch count=%d", cap.count())
	}
}

func TestCodexFactoryLeaderExecFailureClearsOwner(t *testing.T) {
	root, _ := fcFixture(t)
	t.Chdir(root)
	netScrubLaneEnv(t)
	withCodexLaunchCapture(t)
	withCodexProjectRoot(t, root)
	codexDirectLaunchFn = func(c *exec.Cmd) error {
		id := launchEnvValue(c.Env, config.EnvFactoryRunID)
		c.Path = filepath.Join(root, "missing-codex")
		err := defaultCodexDirectLaunch(c)
		if err == nil {
			t.Fatal("missing child succeeded")
		}
		if pid, start := runOwnerStamp(t, root, id); pid != 0 || start != "" {
			t.Fatalf("failed child owner=%d %q", pid, start)
		}
		if n := peerCount(t, root, id, factory.RoleLeader); n != 0 {
			t.Fatalf("pending peers after failure=%d", n)
		}
		return err
	}
	if _, _, err := runCodexCmd(t, "-f"); err == nil {
		t.Fatal("exec failure not propagated")
	}
}

func TestCodexFactoryLeaderNativeSeparatorKeepsIdentityOptions(t *testing.T) {
	root, _ := fcFixture(t)
	t.Chdir(root)
	netScrubLaneEnv(t)
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, root)
	if _, _, err := runCodexCmd(t, "-f", "--", "--", "prompt with --literal"); err != nil {
		t.Fatal(err)
	}
	rec := codexOnlyRecord(t, cap)
	at := slices.Index(rec.Argv, "--")
	if at < 0 || !slices.Equal(rec.Argv[at:], []string{"--", "prompt with --literal"}) {
		t.Fatalf("native tail changed: %q", rec.Argv)
	}
	values := codexShellEnvSetValues(rec.Argv[:at])
	if values[config.EnvFactoryRole] != `"leader"` {
		t.Fatalf("leader options after native separator: %q", rec.Argv)
	}
}

// The child seam replaces only the model session. Dispatch, lease/stage,
// durable bindings and the leader's status read use their production paths.
// The committed work and audit verdict are explicit lane-evidence fixtures;
// merge-ready is a lane result, not a remote merge or final CardDone.
func TestCodexFactoryLeaderQueueRoundTripWithoutNudge(t *testing.T) {
	root, store := fcFixture(t)
	t.Chdir(root)
	netScrubLaneEnv(t)
	withCodexLaunchCapture(t)
	withCodexProjectRoot(t, root)
	calls := 0
	codexDirectLaunchFn = func(c *exec.Cmd) error {
		calls++
		id := launchEnvValue(c.Env, config.EnvFactoryRunID)
		if id == "" {
			t.Fatal("leader child has no run")
		}
		broker, err := factorymsg.BrokerPath(root, id)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(broker); !os.IsNotExist(err) {
			t.Fatalf("messaging must be unavailable: %v", err)
		}
		sdRegisterLane(t, root, "lane-1")
		restoreLane := enterFactoryLaneMode("lane-1", 0, "", config.FactoryDispatchAuto)
		// No card is an intentional refusal, never a dispatch success.
		_, _, err = runFactory(t, "next", "--run", id)
		sdExit3(t, "empty leader queue", err)
		restoreLane()
		_, stderr, cardID := fcGTDDispatch(t, store, "Codex leader durable dispatch", "codex-leader-roundtrip", "lane-1", id)
		if strings.Contains(stderr, factoryRecordUnavailableTag) {
			t.Fatalf("dispatch not durable: %s", stderr)
		}
		// Reopen the queue and factory stores; no fixture dispatch is pre-created.
		reopened := factory.NewBacklogStore(factory.BacklogPathForRoot(root))
		fcWantCurrent(t, reopened, cardID, id, "lane-1")
		record, err := reopened.LoadPure()
		if err != nil {
			t.Fatal(err)
		}
		assignmentFound := false
		for _, assignment := range record.Runtime.Assignments {
			if assignment.RunID == id && assignment.CardID == cardID && assignment.OwnerLabel == "lane-1" {
				assignmentFound = true
				break
			}
		}
		if !assignmentFound {
			t.Fatalf("durable queue assignment missing run=%s card=%s lane=lane-1: %+v", id, cardID, record.Runtime.Assignments)
		}
		db, err := homestate.OpenFactory(root)
		if err != nil {
			t.Fatal(err)
		}
		row, linked, err := db.RecordedCardRowReadonly(context.Background(), cardID)
		if err != nil || !linked || row.RunID != id || row.OwnerLabel != "lane-1" || row.State != homestate.CardAssigned {
			t.Fatalf("durable dispatch row=%+v linked=%v err=%v", row, linked, err)
		}
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
		restoreLane = enterFactoryLaneMode("lane-1", 0, "", config.FactoryDispatchAuto)
		defer restoreLane()
		if _, _, err = runFactory(t, "next", "--run", id); err != nil {
			t.Fatal(err)
		}
		db, err = homestate.OpenFactory(root)
		if err != nil {
			t.Fatal(err)
		}
		card, err := db.LoadCard(context.Background(), id, cardID)
		_ = db.Close()
		if err != nil {
			t.Fatal(err)
		}
		wt := card.WorktreePath
		if wt == "" {
			t.Fatal("lease created no worktree")
		}
		if err := os.WriteFile(filepath.Join(wt, "done.txt"), []byte("lane work fixture\n"), 0600); err != nil {
			t.Fatal(err)
		}
		fcGit(t, wt, "add", "done.txt")
		fcGit(t, wt, "commit", "-q", "-m", "lane work fixture")
		sha := fcGit(t, wt, "rev-parse", "HEAD")
		for _, stage := range []struct{ state, evidence string }{{homestate.CardRun, ""}, {homestate.CardSync, sha}, {homestate.CardSyncAudit, sha + ":done.txt"}} {
			args := []string{"stage", cardID, stage.state}
			if stage.evidence != "" {
				args = append(args, stage.evidence)
			}
			args = append(args, "--run", id)
			if _, _, err := runFactory(t, args...); err != nil {
				t.Fatalf("stage %s: %v", stage.state, err)
			}
		}
		verdict := filepath.Join(wt, ".moai", "reports", cardID, "sync-audit.md")
		if err := os.MkdirAll(filepath.Dir(verdict), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(verdict, []byte("verdict: PASS\naudited_sha: "+sha+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
		if _, _, err := runFactory(t, "stage", cardID, "merge-ready", "--run", id); err != nil {
			t.Fatal(err)
		}
		restoreLane()
		out, _, err := runFactory(t, "status", "--run", id, "--json")
		if err != nil {
			t.Fatal(err)
		}
		var report factoryStatusReport
		if err := json.Unmarshal([]byte(out), &report); err != nil {
			t.Fatal(err)
		}
		if report.Run != id || len(report.Cards) != 1 || report.Cards[0].State != homestate.CardMergeReady || report.Cards[0].CardID != cardID || report.Cards[0].RunID != id || report.Cards[0].Owner != "lane-1" {
			t.Fatalf("saved lane completion status=%s", out)
		}
		if _, err := os.Stat(broker); !os.IsNotExist(err) {
			t.Fatalf("queue roundtrip unexpectedly created messaging broker: %v", err)
		}
		t.Logf("persisted dispatch run=%s card=%s lane=lane-1; factory status state=%s; messaging nudge calls=0", id, cardID, report.Cards[0].State)
		return nil
	}
	if _, _, err := runCodexCmd(t, "-f"); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("leader launches=%d", calls)
	}
}

func TestCodexFactoryLeaderSpawnRegistersOwner(t *testing.T) {
	root, _ := fcFixture(t)
	t.Chdir(root)
	netScrubLaneEnv(t)
	withCodexLaunchCapture(t)
	withCodexProjectRoot(t, root)
	priorTmux, priorLookup, priorSpawn, priorIdentity, priorAnchor := inTmuxFn, spawnLookPath, tmuxSpawnFn, codexSpawnPaneIdentityFn, codexSpawnAnchorFn
	t.Cleanup(func() {
		inTmuxFn, spawnLookPath, tmuxSpawnFn, codexSpawnPaneIdentityFn, codexSpawnAnchorFn = priorTmux, priorLookup, priorSpawn, priorIdentity, priorAnchor
	})
	inTmuxFn = func() bool { return true }
	spawnLookPath = func(s string) (string, error) { return "/sentinel/" + s, nil }
	tmuxSpawnFn = func(string, string) (string, error) { return "%leader", nil }
	codexSpawnPaneIdentityFn = func(string) (int, string, error) { return os.Getpid(), homestate.CurrentProcessFingerprint(), nil }
	codexSpawnAnchorFn = nil
	var id string
	codexSpawnLaunchFn = func(dir, program string, args, env []string) error {
		id = launchEnvValue(env, config.EnvFactoryRunID)
		if launchEnvValue(env, config.EnvFactoryRole) != factory.RoleLeader {
			t.Fatalf("spawn identity=%q", env)
		}
		return defaultCodexSpawnLaunch(dir, program, args, env)
	}
	if _, err := withStdoutCapture(t, func() error { _, _, err := runCodexCmd(t, "-f", "--spawn"); return err }); err != nil {
		t.Fatal(err)
	}
	if pid, start := runOwnerStamp(t, root, id); pid != os.Getpid() || start != homestate.CurrentProcessFingerprint() {
		t.Fatalf("spawn owner=%d %q", pid, start)
	}
	if n := peerCount(t, root, id, factory.RoleLeader); n != 1 {
		t.Fatalf("leader pending peers=%d", n)
	}
}

func TestCodexFactoryLeaderWriterRefusalPreservesAnchor(t *testing.T) {
	for _, reason := range []string{"claude session other (pid " + fmt.Sprint(os.Getppid()) + ")", "held by an operator"} {
		t.Run(reason, func(t *testing.T) {
			cap := withCodexLaunchCapture(t)
			r := newAnchorRepo(t)
			withRealCodexWorktreeAnchor(t)
			tree := r.tree(t, "plain")
			lockAnchorTree(t, tree, reason)
			before := snapshotAnchorTree(t, tree)
			_, _, err := runCodexCmd(t, "-f", "-w", "plain")
			if err == nil || cap.count() != 0 {
				t.Fatalf("writer refusal=%v launches=%d", err, cap.count())
			}
			if !reflect.DeepEqual(before, snapshotAnchorTree(t, tree)) {
				t.Fatal("leader refusal changed existing anchor")
			}
		})
	}
}

func TestCodexFactoryLeaderSameSecondDoesNotOverwriteRun(t *testing.T) {
	root, _ := fcFixture(t)
	t.Chdir(root)
	netScrubLaneEnv(t)
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, root)
	oldID := factory.NewRunID()
	if err := recordFactoryRunStart(root, oldID, factory.BackendClaude, "", 0); err != nil {
		t.Fatal(err)
	}
	t.Setenv(config.EnvFactoryRunID, oldID)
	if _, _, err := runCodexCmd(t, "-f"); err != nil {
		t.Fatal(err)
	}
	id := launchEnvValue(codexOnlyRecord(t, cap).Env, config.EnvFactoryRunID)
	if id == oldID {
		t.Fatalf("fresh leader reused same-second existing run %q", id)
	}
	db := fcOpen(t, root)
	var backend string
	if err := db.DB.QueryRow("SELECT lead_backend FROM runs WHERE run_id=?", oldID).Scan(&backend); err != nil || backend != factory.BackendClaude {
		t.Fatalf("old run backend=%q err=%v", backend, err)
	}
}

func TestCodexFactoryLeaderMixedLanesJoinULIDRun(t *testing.T) {
	root, _ := fcFixture(t)
	t.Chdir(root)
	netScrubLaneEnv(t)
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, root)
	if _, _, err := runCodexCmd(t, "-f"); err != nil {
		t.Fatal(err)
	}
	id := launchEnvValue(codexOnlyRecord(t, cap).Env, config.EnvFactoryRunID)
	for _, launcher := range []string{"cc", "glm"} {
		n, err := driveFactoryEntry(t, launcher, "-l", "--factory-run", id)
		if err != nil || n != 1 {
			t.Fatalf("%s ULID run join launches=%d err=%v", launcher, n, err)
		}
	}
	if _, _, err := runCodexCmd(t, "-l", "--factory-run", id); err != nil {
		t.Fatalf("Codex ULID run join: %v", err)
	}
	if cap.count() != 2 {
		t.Fatalf("Codex launches=%d want leader+lane", cap.count())
	}
	lane := cap.records[1]
	if launchEnvValue(lane.Env, config.EnvFactoryRunID) != id || launchEnvValue(lane.Env, config.EnvFactoryRole) != config.FactoryRoleLane {
		t.Fatalf("Codex lane identity=%q", lane.Env)
	}
	if workerRows(t, root) != 3 {
		t.Fatalf("mixed lane claims=%d", workerRows(t, root))
	}
	for _, launcher := range []string{"cc", "glm"} {
		if n, err := driveFactoryEntry(t, launcher, "-f", "--factory-run", id); err == nil || n != 0 {
			t.Fatalf("%s cross-leader takeover launches=%d err=%v", launcher, n, err)
		}
	}
	_, backend, pid := runRow(t, root, id)
	if backend != BackendCodex || pid != os.Getpid() {
		t.Fatalf("leader ownership after joins=%q pid=%d", backend, pid)
	}
}
