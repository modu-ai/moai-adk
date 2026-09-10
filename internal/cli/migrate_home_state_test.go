package cli

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

func homeStateFixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	home := filepath.Join(t.TempDir(), "home")
	t.Setenv("MOAI_HOME", home)
	source := filepath.Join(root, ".moai", "state", "todo", "backlog.db")
	if err := os.MkdirAll(filepath.Dir(source), 0o700); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", source)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE items(id TEXT PRIMARY KEY, text TEXT NOT NULL); INSERT INTO items VALUES('t1','one'),('t2','two')`)
	if closeErr := db.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		t.Fatal(err)
	}
	return root, home
}

func runHomeState(t *testing.T, root string, apply bool) (string, error) {
	t.Helper()
	var out bytes.Buffer
	r := homeStateRunner{projectRoot: root, apply: apply, stdout: &out}
	err := r.Run(context.Background())
	return out.String(), err
}

func validHomeStateLedgerForTest(head string) *homeStateEvidenceLedger {
	ledger := &homeStateEvidenceLedger{Head: head, Records: map[string]homeStateEvidenceRecord{}, Checks: map[string]homeStateEvidenceRecord{}}
	for _, names := range liveTestGroups {
		for _, name := range names {
			ledger.Records[name] = newHomeStateEvidenceRecord("go test -json "+name, []byte(fmt.Sprintf("{\"Action\":\"pass\",\"Test\":%q}\n", name)), 0, head)
		}
	}
	for _, name := range []string{"coverage", "vet", "native"} {
		ledger.Checks[name] = newHomeStateEvidenceRecord(name, []byte("PASS"), 0, head)
	}
	for pkg := range liveRaceGroups {
		ledger.Checks["race:"+pkg] = newHomeStateEvidenceRecord("race "+pkg, []byte("PASS"), 0, head)
	}
	for _, pkg := range []string{"./internal/homestate", "./internal/hook/handoff", "./internal/hook", "./internal/kanban", "./internal/cli"} {
		ledger.Checks["windows:"+pkg] = newHomeStateEvidenceRecord("windows "+pkg, []byte("PASS"), 0, head)
	}
	return ledger
}

func TestHomeStateDryRunNoMutation(t *testing.T) {
	root, home := homeStateFixture(t)
	before, _ := os.ReadDir(home)
	out, err := runHomeState(t, root, false)
	if err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadDir(home)
	if len(before) != len(after) {
		t.Fatalf("dry-run mutated home: before=%d after=%d", len(before), len(after))
	}
	if !strings.Contains(out, "mode: dry-run") {
		t.Fatalf("output = %q", out)
	}
}

func TestHomeStateDryRunReport(t *testing.T) {
	root, _ := homeStateFixture(t)
	registry := filepath.Join(root, ".moai", "state", "active-sessions.json")
	if err := os.WriteFile(registry, []byte(`[{"session_id":"live","pid":`+strconv.Itoa(os.Getpid())+`}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := runHomeState(t, root, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"canonical-root:", "project-key:", "source:", "target:", "active-census: sessions=1 factory=0 mcp=0", "logical-count: 2", "integrity: ok", "search: not-applicable (no runtime producer)"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %q", want, out)
		}
	}
}

func TestHomeStateDryRunClassifiesEquivalentAndDivergentTargets(t *testing.T) {
	root, _ := homeStateFixture(t)
	if err := (&homeStateRunner{projectRoot: root}).Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := runHomeState(t, root, true); err != nil {
		t.Fatal(err)
	}
	out, err := runHomeState(t, root, false)
	if err != nil || !strings.Contains(out, "target-status: equivalent") {
		t.Fatalf("equivalent dry-run out=%q err=%v", out, err)
	}
	target, _ := homestate.BacklogDBPath(root)
	db, err := sql.Open("sqlite", target)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO items VALUES('divergent','row')`)
	_ = db.Close()
	if err != nil {
		t.Fatal(err)
	}
	out, err = runHomeState(t, root, false)
	if err != nil || !strings.Contains(out, "target-status: divergent") {
		t.Fatalf("divergent dry-run out=%q err=%v", out, err)
	}
}

func TestHomeStateApplyCensusFailClosed(t *testing.T) {
	root, _ := homeStateFixture(t)
	registry := filepath.Join(root, ".moai", "state", "active-sessions.json")
	if err := os.MkdirAll(filepath.Dir(registry), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(registry, []byte(`[{"session_id":"live","pid":`+strconv.Itoa(os.Getpid())+`}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	census, err := homestate.ReadRuntimeCensus(root)
	if err != nil || census.ActiveSessions != 1 {
		t.Fatalf("census=%+v err=%v", census, err)
	}
	_, err = runHomeState(t, root, true)
	if err == nil || !strings.Contains(err.Error(), "active runtime") {
		t.Fatalf("err = %v", err)
	}
	if err := os.WriteFile(registry, []byte("not-json"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err = runHomeState(t, root, true)
	if err == nil || !strings.Contains(err.Error(), "census") {
		t.Fatalf("err = %v", err)
	}
	if err := os.Remove(registry); err != nil {
		t.Fatal(err)
	}
	calls := 0
	r := &homeStateRunner{projectRoot: root, apply: true, stdout: &bytes.Buffer{}, runtimeCensusReader: func(string) (homestate.RuntimeCensus, error) {
		calls++
		if calls == 1 {
			return homestate.RuntimeCensus{Fingerprint: "0:0:0"}, nil
		}
		return homestate.RuntimeCensus{ActiveMCPServers: 1, Fingerprint: "0:0:1"}, nil
	}}
	if err := r.Run(context.Background()); err == nil || !strings.Contains(err.Error(), "changed") {
		t.Fatalf("second census err=%v", err)
	}
	if calls != 2 {
		t.Fatalf("census calls=%d", calls)
	}
	backups, _ := filepath.Glob(filepath.Join(os.Getenv("MOAI_HOME"), "backups", homestate.ProjectKey(root), "*"))
	if len(backups) != 0 {
		t.Fatalf("backup before second census gate: %v", backups)
	}
	root2, _ := homeStateFixture(t)
	calls = 0
	r = &homeStateRunner{projectRoot: root2, apply: true, stdout: &bytes.Buffer{}, runtimeCensusReader: func(string) (homestate.RuntimeCensus, error) {
		calls++
		if calls == 1 {
			return homestate.RuntimeCensus{Fingerprint: "0:0:0"}, nil
		}
		return homestate.RuntimeCensus{}, context.Canceled
	}}
	if err := r.Run(context.Background()); err == nil || !strings.Contains(err.Error(), "second active") {
		t.Fatalf("second census indeterminate err=%v", err)
	}
}

func TestHomeStateBackupBeforeWrite(t *testing.T) {
	root, _ := homeStateFixture(t)
	_, err := runHomeState(t, root, true)
	if err != nil {
		t.Fatal(err)
	}
	target, _ := homestate.BacklogDBPath(root)
	manifests, _ := filepath.Glob(filepath.Join(os.Getenv("MOAI_HOME"), "backups", homestate.ProjectKey(root), "*", "manifest.json"))
	if len(manifests) != 1 {
		t.Fatalf("manifests = %v", manifests)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatal(err)
	}
}

func TestHomeStateRefusesDivergentTarget(t *testing.T) {
	root, _ := homeStateFixture(t)
	target, _ := homestate.BacklogDBPath(root)
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, ".moai", "state", "todo", "backlog.db")
	if err := copySQLiteConsistent(source, target); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", target)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO items VALUES('different','different')`)
	_ = db.Close()
	if err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(target)
	dryOut, dryErr := runHomeState(t, root, false)
	dryAfter, _ := os.ReadFile(target)
	if dryErr != nil || !strings.Contains(dryOut, "target-status: divergent") || !bytes.Equal(before, dryAfter) {
		t.Fatalf("dry err=%v out=%q changed=%v", dryErr, dryOut, !bytes.Equal(before, dryAfter))
	}
	_, err = runHomeState(t, root, true)
	after, _ := os.ReadFile(target)
	if err == nil || !bytes.Equal(before, after) {
		t.Fatalf("err=%v target changed=%v", err, !bytes.Equal(before, after))
	}
}

func TestHomeStateDryRunReportsUnreadableTargetWithoutMutation(t *testing.T) {
	root, _ := homeStateFixture(t)
	target, _ := homestate.BacklogDBPath(root)
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("not sqlite"), 0o600); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(target)
	out, err := runHomeState(t, root, false)
	after, _ := os.ReadFile(target)
	if err != nil || !strings.Contains(out, "target-status: unreadable") || !bytes.Equal(before, after) {
		t.Fatalf("err=%v out=%q changed=%v", err, out, !bytes.Equal(before, after))
	}
}

func TestHomeStateApplyFaultPreservesSource(t *testing.T) {
	root, _ := homeStateFixture(t)
	source := filepath.Join(root, ".moai", "state", "todo", "backlog.db")
	before, _ := os.ReadFile(source)
	var out bytes.Buffer
	err := (&homeStateRunner{projectRoot: root, apply: true, stdout: &out, failAfterBackup: true}).Run(context.Background())
	after, _ := os.ReadFile(source)
	if err == nil || !bytes.Equal(before, after) {
		t.Fatalf("err=%v source changed", err)
	}
}

func TestHomeStateApplyPreservesSourceAndBackup(t *testing.T) {
	root, _ := homeStateFixture(t)
	source := filepath.Join(root, ".moai", "state", "todo", "backlog.db")
	if _, err := runHomeState(t, root, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(source); err != nil {
		t.Fatal(err)
	}
	backups, _ := filepath.Glob(filepath.Join(os.Getenv("MOAI_HOME"), "backups", homestate.ProjectKey(root), "*", "backlog.db"))
	if len(backups) != 1 {
		t.Fatalf("backups = %v", backups)
	}
}

func TestHomeStateApplyIdempotentNoOp(t *testing.T) {
	root, _ := homeStateFixture(t)
	if _, err := runHomeState(t, root, true); err != nil {
		t.Fatal(err)
	}
	target, _ := homestate.BacklogDBPath(root)
	before, _ := os.ReadFile(target)
	out, err := runHomeState(t, root, true)
	after, _ := os.ReadFile(target)
	if err != nil || !bytes.Equal(before, after) || !strings.Contains(out, "already migrated") {
		t.Fatalf("err=%v out=%q", err, out)
	}
}

func TestHomeStateBarrierAdmissionHaltsAllHosts(t *testing.T) {
	root, _ := homeStateFixture(t)
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	release, err := homestate.AcquireMigrationAdmission(root, "m1")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = release(false) }()
	if err := homestate.CheckRuntimeAdmission(root); err == nil {
		t.Fatal("runtime admitted while marker active")
	}
	out, err := hook.NewSessionStartHandler(nil, hook.WithSynchronousDeferredScans()).Handle(context.Background(), &hook.HookInput{SessionID: "blocked", ProjectDir: root, CWD: root})
	if err != nil || out.Continue == nil || *out.Continue || out.StopReason == "" {
		t.Fatalf("session output=%+v err=%v", out, err)
	}
	if _, err := os.Stat(filepath.Join(root, ".moai", "state", "active-sessions.json")); !os.IsNotExist(err) {
		t.Fatalf("blocked session wrote registry: %v", err)
	}
	if _, err := kanban.ClaimFactoryWorkerName(root, "lane-1", os.Getpid(), func(int) bool { return true }); err == nil {
		t.Fatal("factory worker admitted")
	}
	if err := runMCPServer(); err == nil {
		t.Fatal("MCP server admitted")
	}
}

func TestHomeStateStartVsMigrateSerialized(t *testing.T) {
	root, _ := homeStateFixture(t)
	for i := 0; i < 100; i++ {
		lock, err := homestate.AcquireAdmissionLock(root)
		if err != nil {
			t.Fatal(err)
		}
		release, err := homestate.InstallMigrationMarkerLocked(root, "race")
		if err != nil {
			if releaseErr := lock.Release(); releaseErr != nil {
				t.Errorf("release admission lock after marker failure: %v", releaseErr)
			}
			t.Fatal(err)
		}
		admitted := make(chan error, 1)
		go func() { admitted <- homestate.WithRuntimeAdmission(root, func() error { return nil }) }()
		select {
		case <-admitted:
			t.Fatal("runtime passed while migration held lock")
		case <-time.After(time.Millisecond):
		}
		if err := lock.Release(); err != nil {
			t.Fatal(err)
		}
		if err := <-admitted; err == nil {
			t.Fatal("runtime admitted while marker active")
		}
		_ = release(true)
	}
}

func TestHomeStateCrashMarkerFailsClosed(t *testing.T) {
	root, _ := homeStateFixture(t)
	if err := homestate.WriteMigrationMarker(root, []byte(`{"migration_id":"dead","owner_pid":99999999}`)); err != nil {
		t.Fatal(err)
	}
	if err := homestate.CheckRuntimeAdmission(root); err == nil || !strings.Contains(err.Error(), "recover") {
		t.Fatalf("err=%v", err)
	}
}

func TestHomeStateVerifiedLiveGateCannotBypassOrReplay(t *testing.T) {
	root, home := homeStateFixture(t)
	auth, err := prepareLiveAuthorization(root, "test-head")
	if err != nil {
		t.Fatal(err)
	}
	auth.ledger = validHomeStateLedgerForTest("test-head")
	var out bytes.Buffer
	r := &homeStateRunner{projectRoot: root, apply: true, verifiedLive: true, stdout: &out, liveVerifier: func(context.Context, string) (*liveAuthorization, error) { return auth, nil }, headReader: func(string) (string, error) { return "test-head", nil }}
	if err = r.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	target, _ := homestate.BacklogDBPath(root)
	before, _ := os.ReadFile(target)
	if err = r.Run(context.Background()); err == nil || !strings.Contains(err.Error(), "consumed") {
		t.Fatalf("replay err=%v", err)
	}
	after, _ := os.ReadFile(target)
	if !bytes.Equal(before, after) {
		t.Fatal("replay mutated target")
	}
	bare := newMigrateHomeStateCmd()
	bare.SetArgs([]string{"--apply"})
	bare.SetOut(&bytes.Buffer{})
	if err := bare.Execute(); err == nil {
		t.Fatal("bare apply accepted")
	}
	if _, statErr := os.Stat(filepath.Join(home, "backups")); !os.IsNotExist(statErr) {
		// The first authorized apply creates one backup; replay must not create a second.
		entries, _ := filepath.Glob(filepath.Join(home, "backups", homestate.ProjectKey(root), "*"))
		if len(entries) != 1 {
			t.Fatalf("backup count=%d", len(entries))
		}
	}
}

func TestHomeStateAuthorizationRejectsForgedAndStaleBindings(t *testing.T) {
	census := homestate.RuntimeCensus{Fingerprint: "0:0:0"}
	source := homeStateCensus{Digest: "source"}
	for _, tc := range []struct {
		name   string
		auth   *liveAuthorization
		head   string
		census homestate.RuntimeCensus
		source homeStateCensus
	}{
		{"nil", nil, "head", census, source},
		{"empty-nonce", &liveAuthorization{}, "head", census, source},
		{"head", &liveAuthorization{nonce: "n", head: "other", censusFingerprint: "0:0:0", sourceDigest: "source"}, "head", census, source},
		{"census", &liveAuthorization{nonce: "n", head: "head", censusFingerprint: "changed", sourceDigest: "source"}, "head", census, source},
		{"source", &liveAuthorization{nonce: "n", head: "head", censusFingerprint: "0:0:0", sourceDigest: "changed"}, "head", census, source},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.auth.consume(tc.head, tc.census, tc.source); err == nil {
				t.Fatal("invalid authorization accepted")
			}
		})
	}
}

func TestHomeStateRunnerRejectsVerifierAndHeadChangeBeforeMutation(t *testing.T) {
	root, home := homeStateFixture(t)
	r := &homeStateRunner{projectRoot: root, apply: true, verifiedLive: true, stdout: &bytes.Buffer{}, liveVerifier: func(context.Context, string) (*liveAuthorization, error) {
		return nil, context.Canceled
	}}
	if err := r.Run(context.Background()); err == nil || !strings.Contains(err.Error(), "verified-live") {
		t.Fatalf("verifier err=%v", err)
	}
	auth, err := prepareLiveAuthorization(root, "old-head")
	if err != nil {
		t.Fatal(err)
	}
	r.liveVerifier = func(context.Context, string) (*liveAuthorization, error) { return auth, nil }
	r.headReader = func(string) (string, error) { return "", context.Canceled }
	if err := r.Run(context.Background()); err == nil {
		t.Fatal("unreadable HEAD accepted")
	}
	auth, err = prepareLiveAuthorization(root, "old-head")
	if err != nil {
		t.Fatal(err)
	}
	r.liveVerifier = func(context.Context, string) (*liveAuthorization, error) { return auth, nil }
	r.headReader = func(string) (string, error) { return "new-head", nil }
	if err := r.Run(context.Background()); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("head-change err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(home, "backups")); !os.IsNotExist(err) {
		t.Fatalf("pre-mutation rejection created backup: %v", err)
	}
}

func TestHomeStateAuthorizationPreparationFailsClosed(t *testing.T) {
	root, _ := homeStateFixture(t)
	registry := filepath.Join(root, ".moai", "state", "active-sessions.json")
	if err := os.WriteFile(registry, []byte(`[{"pid":`+strconv.Itoa(os.Getpid())+`}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := prepareLiveAuthorization(root, "head"); err == nil || !strings.Contains(err.Error(), "not zero") {
		t.Fatalf("active authorization err=%v", err)
	}
	if err := os.WriteFile(registry, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := prepareLiveAuthorization(root, "head"); err == nil {
		t.Fatal("corrupt census authorization accepted")
	}
	root2, _ := homeStateFixture(t)
	source := filepath.Join(root2, ".moai", "state", "todo", "backlog.db")
	if err := os.WriteFile(source, []byte("corrupt"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := prepareLiveAuthorization(root2, "head"); err == nil {
		t.Fatal("corrupt source authorization accepted")
	}
}

func TestHomeStateLiveValidatorRequiresExecutedTestsAndAllTooling(t *testing.T) {
	root, _ := homeStateFixture(t)
	commands := 0
	runner := func(_ context.Context, _ string, _ []string, args ...string) ([]byte, error) {
		commands++
		if len(args) > 0 && args[0] == "run" {
			return []byte("[]\n"), nil
		}
		var out strings.Builder
		seen := map[string]bool{}
		for _, groups := range []map[string][]string{liveTestGroups, liveRaceGroups} {
			for _, names := range groups {
				for _, name := range names {
					if seen[name] {
						continue
					}
					seen[name] = true
					fmt.Fprintf(&out, "{\"Action\":\"pass\",\"Test\":%q}\n", name)
				}
			}
		}
		return []byte(out.String()), nil
	}
	auth, err := validateLivePreApplyWith(context.Background(), root, func(string) (string, error) { return "abcdef", nil }, "abc", runner, func(context.Context, string) (float64, error) { return 85, nil })
	if err != nil || auth == nil || commands < 10 {
		t.Fatalf("auth=%v commands=%d err=%v", auth != nil, commands, err)
	}
	if err := runNamedTests(context.Background(), root, "./internal/cli", []string{"never-ran"}, func(context.Context, string, []string, ...string) ([]byte, error) { return []byte("PASS"), nil }); err == nil {
		t.Fatal("zero executed tests accepted")
	}
}

func TestHomeStateRunNamedTestsRejectsSkipFailAndZeroJSONEvents(t *testing.T) {
	for _, action := range []string{"skip", "fail", "", "duplicate-pass"} {
		t.Run(action, func(t *testing.T) {
			out := []byte(`{"Action":"` + action + `","Test":"required"}` + "\n")
			switch action {
			case "":
				out = []byte(`{"Action":"output","Test":"required"}` + "\n")
			case "duplicate-pass":
				out = []byte("{\"Action\":\"pass\",\"Test\":\"required\"}\n{\"Action\":\"pass\",\"Test\":\"required\"}\n")
			}
			runner := func(context.Context, string, []string, ...string) ([]byte, error) { return out, nil }
			if err := runNamedTests(context.Background(), t.TempDir(), "./internal/cli", []string{"required"}, runner); err == nil {
				t.Fatalf("%q event accepted", action)
			}
		})
	}
}

func TestHomeStateLiveValidatorFailsClosedAtEveryToolGate(t *testing.T) {
	root, _ := homeStateFixture(t)
	allRuns := func() []byte {
		var out strings.Builder
		seen := map[string]bool{}
		for _, groups := range []map[string][]string{liveTestGroups, liveRaceGroups} {
			for _, names := range groups {
				for _, name := range names {
					if seen[name] {
						continue
					}
					seen[name] = true
					fmt.Fprintf(&out, "{\"Action\":\"pass\",\"Test\":%q}\n", name)
				}
			}
		}
		return []byte(out.String())
	}
	for _, gate := range []string{"head", "build", "tests", "vet", "lint", "windows", "coverage-low", "coverage-zero", "coverage-error"} {
		t.Run(gate, func(t *testing.T) {
			headReader := func(string) (string, error) { return "abcdef", nil }
			build := "abc"
			if gate == "head" {
				headReader = func(string) (string, error) { return "", context.Canceled }
			}
			if gate == "build" {
				build = "dev"
			}
			runner := func(_ context.Context, _ string, env []string, args ...string) ([]byte, error) {
				if gate == "tests" && len(args) > 0 && args[0] == "test" && env == nil {
					return nil, context.Canceled
				}
				if gate == "vet" && len(args) > 0 && args[0] == "vet" {
					return nil, context.Canceled
				}
				if len(args) > 0 && args[0] == "run" {
					if gate == "lint" {
						return []byte(`[{}]`), nil
					}
					return []byte("[]"), nil
				}
				if gate == "windows" && env != nil {
					return nil, context.Canceled
				}
				return allRuns(), nil
			}
			coverage := func(context.Context, string) (float64, error) { return 85, nil }
			if gate == "coverage-low" {
				coverage = func(context.Context, string) (float64, error) { return 84.9, nil }
			}
			if gate == "coverage-zero" {
				coverage = func(context.Context, string) (float64, error) { return 0, nil }
			}
			if gate == "coverage-error" {
				coverage = func(context.Context, string) (float64, error) { return 0, context.Canceled }
			}
			if _, err := validateLivePreApplyWith(context.Background(), root, headReader, build, runner, coverage); err == nil {
				t.Fatalf("%s gate accepted", gate)
			}
		})
	}
}

func TestHomeStateValidationCommandWrappersAndHelperFailures(t *testing.T) {
	repo, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if head, err := readGitHead(repo); err != nil || head == "" {
		t.Fatalf("head=%q err=%v", head, err)
	}
	if out, err := execLiveCommand(context.Background(), repo, nil, "version"); err != nil || !strings.Contains(string(out), "go version") {
		t.Fatalf("go version=%q err=%v", out, err)
	}
	if _, err := validateLivePreApply(context.Background(), repo); err == nil {
		t.Fatal("development binary unexpectedly passed live validation")
	}
	missing := filepath.Join(t.TempDir(), "missing.db")
	if _, err := sqliteCensus(missing); err == nil {
		t.Fatal("missing sqlite accepted")
	}
	if err := copySQLiteConsistent(missing, filepath.Join(t.TempDir(), "target.db")); err == nil {
		t.Fatal("missing sqlite copied")
	}
	if _, err := homeStateFileSHA256(missing); err == nil {
		t.Fatal("missing file hashed")
	}
	if _, err := readGitHead(t.TempDir()); err == nil {
		t.Fatal("non-git directory returned a head")
	}
	if _, err := execLiveCommand(context.Background(), filepath.Join(t.TempDir(), "missing"), nil, "version"); err == nil {
		t.Fatal("missing command directory executed")
	}
	corrupt := filepath.Join(t.TempDir(), "corrupt.db")
	if err := os.WriteFile(corrupt, []byte("not sqlite"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := sqliteCensus(corrupt); err == nil {
		t.Fatal("corrupt sqlite accepted")
	}
}

func TestHomeStateRecoveryAndRollbackCoverSafeExistingAndMissingTargets(t *testing.T) {
	t.Run("recovery keeps a matching target", func(t *testing.T) {
		root, home := homeStateFixture(t)
		if err := (&homeStateRunner{projectRoot: root, apply: true, stdout: &bytes.Buffer{}, failAfterBackup: true}).Run(context.Background()); err == nil {
			t.Fatal("fault did not leave recovery state")
		}
		marker, err := homestate.ReadMigrationMarker(root)
		if err != nil {
			t.Fatal(err)
		}
		marker.OwnerPID = os.Getpid()
		marker.OwnerFingerprint = "reused-pid"
		raw, _ := json.Marshal(marker)
		if err := homestate.WriteMigrationMarker(root, raw); err != nil {
			t.Fatal(err)
		}
		backup := filepath.Join(home, "backups", homestate.ProjectKey(root), marker.MigrationID, "backlog.db")
		target, _ := homestate.BacklogDBPath(root)
		if err := copySQLiteConsistent(backup, target); err != nil {
			t.Fatal(err)
		}
		if err := recoverHomeState(context.Background(), root, marker.MigrationID, marker.MigrationID); err != nil {
			t.Fatal(err)
		}
		if matches, _ := filepath.Glob(target + ".quarantine-*"); len(matches) != 0 {
			t.Fatalf("matching target quarantined: %v", matches)
		}
	})
	t.Run("recovery quarantines a valid divergent target", func(t *testing.T) {
		root, _ := homeStateFixture(t)
		if err := (&homeStateRunner{projectRoot: root, apply: true, stdout: &bytes.Buffer{}, failAfterBackup: true}).Run(context.Background()); err == nil {
			t.Fatal("fault did not leave recovery state")
		}
		marker, err := homestate.ReadMigrationMarker(root)
		if err != nil {
			t.Fatal(err)
		}
		marker.OwnerPID = os.Getpid()
		marker.OwnerFingerprint = "reused-pid"
		raw, _ := json.Marshal(marker)
		if err := homestate.WriteMigrationMarker(root, raw); err != nil {
			t.Fatal(err)
		}
		target, _ := homestate.BacklogDBPath(root)
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			t.Fatal(err)
		}
		db, err := sql.Open("sqlite", target)
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(`CREATE TABLE items(id TEXT PRIMARY KEY, body TEXT); INSERT INTO items VALUES('different','row')`)
		_ = db.Close()
		if err != nil {
			t.Fatal(err)
		}
		if err := recoverHomeState(context.Background(), root, marker.MigrationID, marker.MigrationID); err != nil {
			t.Fatal(err)
		}
		if matches, _ := filepath.Glob(target + ".quarantine-*"); len(matches) != 1 {
			t.Fatalf("divergent target quarantine=%v", matches)
		}
	})
	t.Run("rollback restores a missing target", func(t *testing.T) {
		root, _ := homeStateFixture(t)
		if _, err := runHomeState(t, root, true); err != nil {
			t.Fatal(err)
		}
		target, _ := homestate.BacklogDBPath(root)
		if err := os.Remove(target); err != nil {
			t.Fatal(err)
		}
		if err := rollbackHomeState(context.Background(), root, "latest"); err != nil {
			t.Fatal(err)
		}
		if _, err := sqliteCensus(target); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("rollback rejects an empty backup directory", func(t *testing.T) {
		root, home := homeStateFixture(t)
		base := filepath.Join(home, "backups", homestate.ProjectKey(root))
		if err := os.MkdirAll(base, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := rollbackHomeState(context.Background(), root, "latest"); err == nil || !strings.Contains(err.Error(), "no backup") {
			t.Fatalf("empty backup rollback err=%v", err)
		}
	})
}

func TestHomeStateRecoverCrashMarkerSafely(t *testing.T) {
	root, _ := homeStateFixture(t)
	err := (&homeStateRunner{projectRoot: root, apply: true, stdout: &bytes.Buffer{}, failAfterBackup: true}).Run(context.Background())
	if err == nil {
		t.Fatal("fault did not stop apply")
	}
	marker, err := homestate.ReadMigrationMarker(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := recoverHomeState(context.Background(), root, "wrong", marker.MigrationID); err == nil || !strings.Contains(err.Error(), "migration id") {
		t.Fatalf("wrong migration id err=%v", err)
	}
	originalFingerprint := marker.OwnerFingerprint
	marker.OwnerFingerprint = ""
	raw, _ := json.Marshal(marker)
	if err := homestate.WriteMigrationMarker(root, raw); err != nil {
		t.Fatal(err)
	}
	if err := recoverHomeState(context.Background(), root, marker.MigrationID, marker.MigrationID); err == nil || !strings.Contains(err.Error(), "fingerprint") {
		t.Fatalf("missing fingerprint err=%v", err)
	}
	marker.OwnerFingerprint = originalFingerprint
	raw, _ = json.Marshal(marker)
	if err := homestate.WriteMigrationMarker(root, raw); err != nil {
		t.Fatal(err)
	}
	if err := recoverHomeState(context.Background(), root, marker.MigrationID, marker.MigrationID); err == nil || !strings.Contains(err.Error(), "live") {
		t.Fatalf("live owner recovery err=%v", err)
	}
	// A live PID with a different process-start fingerprint is a provable PID
	// reuse, so recovery may continue.
	marker.OwnerPID = os.Getpid()
	marker.OwnerFingerprint = "different-process-start"
	raw, _ = json.Marshal(marker)
	if err := homestate.WriteMigrationMarker(root, raw); err != nil {
		t.Fatal(err)
	}
	if err := recoverHomeState(context.Background(), root, marker.MigrationID, ""); err == nil || !strings.Contains(err.Error(), "backup id") {
		t.Fatalf("missing backup id err=%v", err)
	}
	if err := recoverHomeState(context.Background(), root, marker.MigrationID, "missing"); err == nil {
		t.Fatal("missing backup accepted")
	}
	target, _ := homestate.BacklogDBPath(root)
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("incomplete"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := recoverHomeState(context.Background(), root, marker.MigrationID, marker.MigrationID); err != nil {
		t.Fatal(err)
	}
	if err := recoverHomeState(context.Background(), root, marker.MigrationID, marker.MigrationID); err != nil {
		t.Fatalf("idempotent recovery: %v", err)
	}
	if err := homestate.CheckRuntimeAdmission(root); err != nil {
		t.Fatal(err)
	}
	if _, err := sqliteCensus(target); err != nil {
		t.Fatal(err)
	}
	quarantined, _ := filepath.Glob(target + ".quarantine-*")
	if len(quarantined) != 1 {
		t.Fatalf("quarantined=%v", quarantined)
	}
}

func TestHomeStateRollbackVerifiedBackup(t *testing.T) {
	root, _ := homeStateFixture(t)
	if _, err := runHomeState(t, root, true); err != nil {
		t.Fatal(err)
	}
	target, _ := homestate.BacklogDBPath(root)
	beforeTarget, _ := os.ReadFile(target)
	beforeEntries, _ := os.ReadDir(filepath.Dir(target))
	// Rollback to the verified pre-apply source backup is intentionally explicit.
	if err := rollbackHomeState(context.Background(), root, "latest"); err != nil {
		t.Fatal(err)
	}
	if err := rollbackHomeState(context.Background(), root, "latest"); err != nil {
		t.Fatalf("idempotent rollback: %v", err)
	}
	afterTarget, _ := os.ReadFile(target)
	afterEntries, _ := os.ReadDir(filepath.Dir(target))
	if !bytes.Equal(beforeTarget, afterTarget) || len(beforeEntries) != len(afterEntries) {
		t.Fatalf("parity rollback mutated target: entries %d -> %d", len(beforeEntries), len(afterEntries))
	}
	if err := homestate.CheckRuntimeAdmission(root); err != nil {
		t.Fatal(err)
	}
	root2, _ := homeStateFixture(t)
	if _, err := runHomeState(t, root2, true); err != nil {
		t.Fatal(err)
	}
	target2, _ := homestate.BacklogDBPath(root2)
	before, _ := os.ReadFile(target2)
	backups, _ := filepath.Glob(filepath.Join(os.Getenv("MOAI_HOME"), "backups", homestate.ProjectKey(root2), "*", "backlog.db"))
	if len(backups) != 1 {
		t.Fatalf("backups=%v", backups)
	}
	if err := os.WriteFile(backups[0], []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := rollbackHomeState(context.Background(), root2, "latest"); err == nil || !strings.Contains(err.Error(), "hash") {
		t.Fatalf("tampered rollback err=%v", err)
	}
	after, _ := os.ReadFile(target2)
	if !bytes.Equal(before, after) {
		t.Fatal("tampered rollback changed target")
	}
}

func TestHomeStateRollbackQuarantinesDivergentTarget(t *testing.T) {
	root, _ := homeStateFixture(t)
	if _, err := runHomeState(t, root, true); err != nil {
		t.Fatal(err)
	}
	target, _ := homestate.BacklogDBPath(root)
	db, err := sql.Open("sqlite", target)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO items VALUES('divergent','row')`)
	_ = db.Close()
	if err != nil {
		t.Fatal(err)
	}
	if err := rollbackHomeState(context.Background(), root, "latest"); err != nil {
		t.Fatal(err)
	}
	entries, _ := filepath.Glob(target + ".rollback-*")
	if len(entries) != 1 {
		t.Fatalf("rollback quarantine=%v", entries)
	}
}

func TestHomeStateBackupIdentityMismatchFailsClosed(t *testing.T) {
	root, home := homeStateFixture(t)
	if _, err := runHomeState(t, root, true); err != nil {
		t.Fatal(err)
	}
	dirs, _ := filepath.Glob(filepath.Join(home, "backups", homestate.ProjectKey(root), "*"))
	if len(dirs) != 1 {
		t.Fatalf("backup dirs=%v", dirs)
	}
	raw, err := os.ReadFile(filepath.Join(dirs[0], "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest homeStateBackupManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.ProjectRoot = t.TempDir()
	raw, _ = json.Marshal(manifest)
	if err := os.WriteFile(filepath.Join(dirs[0], "manifest.json"), raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := verifiedBackup(root, filepath.Base(dirs[0])); err == nil || !strings.Contains(err.Error(), "identity") {
		t.Fatalf("identity mismatch err=%v", err)
	}
}

func TestHomeStateRollbackRequiresMigrationIdentity(t *testing.T) {
	cmd := newMigrateHomeStateCmd()
	cmd.SetArgs([]string{"rollback", "--backup-id", "backup-1"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "--migration-id") {
		t.Fatalf("rollback without migration identity err=%v", err)
	}
}

func TestHomeStateCommandSurfacesDryRunRecoverAndRollback(t *testing.T) {
	root, home := homeStateFixture(t)
	t.Chdir(root)
	if err := rollbackHomeState(context.Background(), root, "latest"); err == nil {
		t.Fatal("rollback without backup accepted")
	}
	dry := newMigrateHomeStateCmd()
	dry.SetOut(&bytes.Buffer{})
	if err := dry.Execute(); err != nil {
		t.Fatal(err)
	}
	recoverCmd := newMigrateHomeStateCmd()
	recoverCmd.SetArgs([]string{"recover"})
	if err := recoverCmd.Execute(); err == nil {
		t.Fatal("recover without identities accepted")
	}
	mismatch := newMigrateHomeStateCmd()
	mismatch.SetArgs([]string{"rollback", "--migration-id", "one", "--backup-id", "two"})
	if err := mismatch.Execute(); err == nil || !strings.Contains(err.Error(), "must match") {
		t.Fatalf("rollback mismatch err=%v", err)
	}
	if _, err := runHomeState(t, root, true); err != nil {
		t.Fatal(err)
	}
	dirs, _ := filepath.Glob(filepath.Join(home, "backups", homestate.ProjectKey(root), "*"))
	id := filepath.Base(dirs[0])
	rollback := newMigrateHomeStateCmd()
	rollback.SetArgs([]string{"rollback", "--migration-id", id, "--backup-id", id})
	if err := rollback.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestHomeStateRecoverCommandRestoresDeadOwner(t *testing.T) {
	root, _ := homeStateFixture(t)
	if err := (&homeStateRunner{projectRoot: root, apply: true, stdout: &bytes.Buffer{}, failAfterBackup: true}).Run(context.Background()); err == nil {
		t.Fatal("fault did not leave recovery state")
	}
	marker, err := homestate.ReadMigrationMarker(root)
	if err != nil {
		t.Fatal(err)
	}
	marker.OwnerPID = os.Getpid()
	marker.OwnerFingerprint = "reused-pid"
	raw, _ := json.Marshal(marker)
	if err := homestate.WriteMigrationMarker(root, raw); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	cmd := newMigrateHomeStateCmd()
	cmd.SetArgs([]string{"recover", "--migration-id", marker.MigrationID, "--backup-id", marker.MigrationID})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestHomeStateVerdictEvidenceValidator(t *testing.T) {
	root, _ := homeStateFixture(t)
	if ids := preApplyACIDs(); len(ids) != 24 || ids[0] != "AC-HSR-001" || ids[len(ids)-1] != "AC-HSR-025" || slices.Contains(ids, "AC-HSR-022") {
		t.Fatalf("pre-apply AC IDs=%v", ids)
	}
	if err := validateHomeStateEvidenceLedger(homeStateEvidenceLedger{}); err == nil || !strings.Contains(err.Error(), "HEAD") {
		t.Fatalf("empty evidence ledger err=%v", err)
	}
	if validHomeStateEvidenceRecord(homeStateEvidenceRecord{}, "abc") {
		t.Fatal("empty evidence record accepted")
	}
	ledger := homeStateEvidenceLedger{Head: "abc", Records: map[string]homeStateEvidenceRecord{}, Readback: homeStateReadback{SourceIntegrity: "ok", TargetIntegrity: "ok", SourceDigest: "same", TargetDigest: "same", SourceExists: true, BackupExists: true, Coverage: 85}}
	if err := validateHomeStateEvidenceLedger(ledger); err == nil {
		t.Fatal("missing records accepted")
	}
	for _, names := range liveTestGroups {
		for _, id := range names {
			out := fmt.Sprintf("{\"Action\":\"pass\",\"Test\":%q}\n", id)
			ledger.Records[id] = newHomeStateEvidenceRecord("go test -json "+id, []byte(out), 0, "abc")
		}
	}
	if _, exists := ledger.Records["AC-HSR-022"]; exists {
		t.Fatal("AC-022 became its own input")
	}
	if err := validateHomeStateEvidenceLedger(ledger); err == nil || !strings.Contains(err.Error(), "evidence incomplete") {
		t.Fatalf("missing quality checks accepted: %v", err)
	}
	ledger.Checks = map[string]homeStateEvidenceRecord{}
	for _, name := range []string{"coverage", "vet", "native"} {
		ledger.Checks[name] = newHomeStateEvidenceRecord(name, []byte("PASS"), 0, "abc")
	}
	for pkg := range liveRaceGroups {
		ledger.Checks["race:"+pkg] = newHomeStateEvidenceRecord("race "+pkg, []byte("PASS"), 0, "abc")
	}
	for _, pkg := range []string{"./internal/homestate", "./internal/hook/handoff", "./internal/hook", "./internal/kanban", "./internal/cli"} {
		ledger.Checks["windows:"+pkg] = newHomeStateEvidenceRecord("windows "+pkg, []byte("PASS"), 0, "abc")
	}
	badCheck := ledger.Checks["vet"]
	badCheck.Output += "tamper"
	ledger.Checks["vet"] = badCheck
	if err := validateHomeStateEvidenceLedger(ledger); err == nil || !strings.Contains(err.Error(), "vet evidence") {
		t.Fatalf("bad check accepted: %v", err)
	}
	ledger.Checks["vet"] = newHomeStateEvidenceRecord("vet", []byte("PASS"), 0, "abc")
	delete(ledger.Checks, "race:./internal/cli")
	if err := validateHomeStateEvidenceLedger(ledger); err == nil || !strings.Contains(err.Error(), "race evidence") {
		t.Fatalf("missing race accepted: %v", err)
	}
	ledger.Checks["race:./internal/cli"] = newHomeStateEvidenceRecord("race", []byte("PASS"), 0, "abc")
	delete(ledger.Checks, "windows:./internal/cli")
	if err := validateHomeStateEvidenceLedger(ledger); err == nil || !strings.Contains(err.Error(), "windows evidence") {
		t.Fatalf("missing windows accepted: %v", err)
	}
	ledger.Checks["windows:./internal/cli"] = newHomeStateEvidenceRecord("windows", []byte("PASS"), 0, "abc")
	ledger.SkipCount = 1
	if err := validateHomeStateEvidenceLedger(ledger); err == nil || !strings.Contains(err.Error(), "skip") {
		t.Fatalf("skip accepted: %v", err)
	}
	ledger.SkipCount = 0
	ledger.Readback.Coverage = 84.9
	if err := validateHomeStateEvidenceLedger(ledger); err == nil || !strings.Contains(err.Error(), "readback") {
		t.Fatalf("low coverage accepted: %v", err)
	}
	ledger.Readback.Coverage = 85
	if err := validateHomeStateEvidenceLedger(ledger); err != nil {
		t.Fatal(err)
	}
	first := ledger.Records[liveTestGroups["./internal/cli"][0]]
	first.Output += "tampered"
	ledger.Records[liveTestGroups["./internal/cli"][0]] = first
	if err := validateHomeStateEvidenceLedger(ledger); err == nil {
		t.Fatal("tampered output/hash accepted")
	}
	_ = root
}

func TestPersistedHomeStateEvidenceIsImmutableAndReadsRealDatabases(t *testing.T) {
	root, home := homeStateFixture(t)
	if _, err := runHomeState(t, root, true); err != nil {
		t.Fatal(err)
	}
	backupDirs, err := filepath.Glob(filepath.Join(home, "backups", homestate.ProjectKey(root), "*"))
	if err != nil || len(backupDirs) != 1 {
		t.Fatalf("backup dirs=%v err=%v", backupDirs, err)
	}
	id := filepath.Base(backupDirs[0])
	source, _ := sqliteCensus(filepath.Join(root, ".moai", "state", "todo", "backlog.db"))
	targetPath, _ := homestate.BacklogDBPath(root)
	target, _ := sqliteCensus(targetPath)
	ledger := &homeStateEvidenceLedger{Head: "abc", Records: map[string]homeStateEvidenceRecord{}, Checks: map[string]homeStateEvidenceRecord{}, Readback: homeStateReadback{SourceIntegrity: source.Integrity, TargetIntegrity: target.Integrity, SourceDigest: source.Digest, TargetDigest: target.Digest, SourceCount: source.Count, TargetCount: target.Count, SourceExists: true, BackupExists: true, Coverage: 85}}
	for _, names := range liveTestGroups {
		for _, name := range names {
			ledger.Records[name] = newHomeStateEvidenceRecord("go test -json "+name, []byte(fmt.Sprintf("{\"Action\":\"pass\",\"Test\":%q}\n", name)), 0, "abc")
		}
	}
	for _, name := range []string{"coverage", "vet", "native"} {
		ledger.Checks[name] = newHomeStateEvidenceRecord(name, []byte("PASS"), 0, "abc")
	}
	for pkg := range liveRaceGroups {
		ledger.Checks["race:"+pkg] = newHomeStateEvidenceRecord("race "+pkg, []byte("PASS"), 0, "abc")
	}
	for _, pkg := range []string{"./internal/homestate", "./internal/hook/handoff", "./internal/hook", "./internal/kanban", "./internal/cli"} {
		ledger.Checks["windows:"+pkg] = newHomeStateEvidenceRecord("windows "+pkg, []byte("PASS"), 0, "abc")
	}
	path, err := persistHomeStateEvidence(root, id, ledger)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := persistHomeStateEvidence(root, id, ledger); err == nil {
		t.Fatal("immutable ledger was overwritten")
	}
	if err := validatePersistedHomeStateEvidence(path, root, id, func(string) (string, error) { return "abc", nil }); err != nil {
		t.Fatal(err)
	}
	if _, err := persistHomeStateEvidence(root, "nil-ledger", nil); err == nil {
		t.Fatal("nil ledger accepted")
	}
	if err := validatePersistedHomeStateEvidence(filepath.Join(home, "missing.json"), root, id, func(string) (string, error) { return "abc", nil }); err == nil {
		t.Fatal("missing ledger accepted")
	}
	if err := validatePersistedHomeStateEvidence(path, root, id, func(string) (string, error) { return "different", nil }); err == nil {
		t.Fatal("stale HEAD accepted")
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	if err := os.WriteFile(path, append(raw, 'x'), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := validatePersistedHomeStateEvidence(path, root, id, func(string) (string, error) { return "abc", nil }); err == nil {
		t.Fatal("tampered ledger accepted")
	}
	if err := os.WriteFile(path, []byte("{"), 0o400); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o400); err != nil {
		t.Fatal(err)
	}
	if err := validatePersistedHomeStateEvidence(path, root, id, func(string) (string, error) { return "abc", nil }); err == nil {
		t.Fatal("malformed ledger accepted")
	}
}

func TestHomeStatePersistedEvidenceReadbackFailureBranches(t *testing.T) {
	root, home := homeStateFixture(t)
	if _, err := runHomeState(t, root, true); err != nil {
		t.Fatal(err)
	}
	dirs, _ := filepath.Glob(filepath.Join(home, "backups", homestate.ProjectKey(root), "*"))
	id := filepath.Base(dirs[0])
	sourcePath := filepath.Join(root, ".moai", "state", "todo", "backlog.db")
	targetPath, _ := homestate.BacklogDBPath(root)
	source, _ := sqliteCensus(sourcePath)
	target, _ := sqliteCensus(targetPath)
	ledger := validHomeStateLedgerForTest("abc")
	ledger.Readback = homeStateReadback{SourceIntegrity: "ok", TargetIntegrity: "ok", SourceDigest: source.Digest, TargetDigest: target.Digest, SourceCount: source.Count, TargetCount: target.Count, SourceExists: true, BackupExists: true, Coverage: 85}
	path, err := persistHomeStateEvidence(root, id+"-checks", ledger)
	if err != nil {
		t.Fatal(err)
	}
	if err := validatePersistedHomeStateEvidence(path, root, "missing-backup", func(string) (string, error) { return "abc", nil }); err == nil {
		t.Fatal("missing backup accepted")
	}
	sourceHold := sourcePath + ".hold"
	if err := os.Rename(sourcePath, sourceHold); err != nil {
		t.Fatal(err)
	}
	if err := validatePersistedHomeStateEvidence(path, root, id, func(string) (string, error) { return "abc", nil }); err == nil {
		t.Fatal("missing source accepted")
	}
	if err := os.Rename(sourceHold, sourcePath); err != nil {
		t.Fatal(err)
	}
	targetHold := targetPath + ".hold"
	if err := os.Rename(targetPath, targetHold); err != nil {
		t.Fatal(err)
	}
	if err := validatePersistedHomeStateEvidence(path, root, id, func(string) (string, error) { return "abc", nil }); err == nil {
		t.Fatal("missing target accepted")
	}
	if err := os.Rename(targetHold, targetPath); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", targetPath)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO items VALUES('extra','extra')`)
	_ = db.Close()
	if err != nil {
		t.Fatal(err)
	}
	if err := validatePersistedHomeStateEvidence(path, root, id, func(string) (string, error) { return "abc", nil }); err == nil {
		t.Fatal("divergent readback accepted")
	}
}

func TestHomeStateRecoveryAndEvidenceStorageFailureBranches(t *testing.T) {
	root, home := homeStateFixture(t)
	if err := recoverHomeState(context.Background(), root, "none", "none"); err != nil {
		t.Fatal(err)
	}
	writeMarker := func(pid int, fingerprint string) {
		t.Helper()
		raw, _ := json.Marshal(map[string]any{"migration_id": "m", "project_key": homestate.ProjectKey(root), "project_root": homestate.CanonicalProjectRoot(root), "owner_pid": pid, "owner_fingerprint": fingerprint})
		if err := homestate.WriteMigrationMarker(root, raw); err != nil {
			t.Fatal(err)
		}
	}
	writeMarker(99999999, "")
	if err := recoverHomeState(context.Background(), root, "m", "b"); err == nil || !strings.Contains(err.Error(), "fingerprint") {
		t.Fatalf("empty fingerprint err=%v", err)
	}
	_ = homestate.ClearMigrationMarker(root)
	writeMarker(os.Getpid(), homestate.CurrentProcessFingerprint())
	if err := recoverHomeState(context.Background(), root, "m", "b"); err == nil || !strings.Contains(err.Error(), "live") {
		t.Fatalf("live owner err=%v", err)
	}
	_ = homestate.ClearMigrationMarker(root)
	writeMarker(-1, "dead")
	if err := recoverHomeState(context.Background(), root, "m", ""); err == nil || !strings.Contains(err.Error(), "backup id") {
		t.Fatalf("empty backup err=%v", err)
	}
	_ = homestate.ClearMigrationMarker(root)
	if err := rollbackHomeState(context.Background(), root, "latest"); err == nil {
		t.Fatal("missing latest backup accepted")
	}
	if _, err := quarantineTarget(filepath.Join(root, "missing.db")); err == nil {
		t.Fatal("missing quarantine target accepted")
	}
	blocked := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(blocked, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOAI_HOME", blocked)
	if _, err := persistHomeStateEvidence(root, "m", validHomeStateLedgerForTest("abc")); err == nil {
		t.Fatal("blocked evidence storage accepted")
	}
	t.Setenv("MOAI_HOME", home)
}

func TestHomeStateBackupRecoveryAndEvidenceRejectUnsafeInputs(t *testing.T) {
	t.Run("malformed backup manifest", func(t *testing.T) {
		root, home := homeStateFixture(t)
		if _, err := runHomeState(t, root, true); err != nil {
			t.Fatal(err)
		}
		dirs, _ := filepath.Glob(filepath.Join(home, "backups", homestate.ProjectKey(root), "*"))
		id := filepath.Base(dirs[0])
		if err := os.WriteFile(filepath.Join(dirs[0], "manifest.json"), []byte("{"), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, _, err := verifiedBackup(root, id); err == nil {
			t.Fatal("malformed backup manifest accepted")
		}
	})

	t.Run("backup manifest count mismatch", func(t *testing.T) {
		root, home := homeStateFixture(t)
		if _, err := runHomeState(t, root, true); err != nil {
			t.Fatal(err)
		}
		dirs, _ := filepath.Glob(filepath.Join(home, "backups", homestate.ProjectKey(root), "*"))
		id := filepath.Base(dirs[0])
		manifestPath := filepath.Join(dirs[0], "manifest.json")
		raw, err := os.ReadFile(manifestPath)
		if err != nil {
			t.Fatal(err)
		}
		var manifest homeStateBackupManifest
		if err := json.Unmarshal(raw, &manifest); err != nil {
			t.Fatal(err)
		}
		manifest.LogicalCount++
		raw, _ = json.Marshal(manifest)
		if err := os.WriteFile(manifestPath, raw, 0o600); err != nil {
			t.Fatal(err)
		}
		if _, _, err := verifiedBackup(root, id); err == nil || !strings.Contains(err.Error(), "restore probe") {
			t.Fatalf("count mismatch err=%v", err)
		}
	})

	t.Run("corrupt recovery marker", func(t *testing.T) {
		root, _ := homeStateFixture(t)
		if err := homestate.WriteMigrationMarker(root, []byte("{")); err != nil {
			t.Fatal(err)
		}
		if err := recoverHomeState(context.Background(), root, "m", "b"); err == nil {
			t.Fatal("corrupt recovery marker accepted")
		}
	})

	t.Run("rollback refuses an existing migration marker", func(t *testing.T) {
		root, _ := homeStateFixture(t)
		if _, err := runHomeState(t, root, true); err != nil {
			t.Fatal(err)
		}
		target, _ := homestate.BacklogDBPath(root)
		db, err := sql.Open("sqlite", target)
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(`INSERT INTO items VALUES('changed','changed')`)
		_ = db.Close()
		if err != nil {
			t.Fatal(err)
		}
		release, err := homestate.AcquireMigrationAdmission(root, "conflict")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := release(true); err != nil {
				t.Errorf("release migration admission: %v", err)
			}
		})
		if err := rollbackHomeState(context.Background(), root, "latest"); err == nil {
			t.Fatal("rollback replaced an active migration marker")
		}
	})

	t.Run("immutable evidence path must be a file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "ledger.json")
		if err := os.Mkdir(path, 0o400); err != nil {
			t.Fatal(err)
		}
		if err := validatePersistedHomeStateEvidence(path, t.TempDir(), "m", func(string) (string, error) { return "abc", nil }); err == nil {
			t.Fatal("evidence directory accepted as a ledger")
		}
	})
}
