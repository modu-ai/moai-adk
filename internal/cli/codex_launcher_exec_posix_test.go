//go:build !windows

package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// TestCodexDirectPOSIXExecRegistersNoFactoryPeer is AC-CFR-007's direct half
// and keeps the exec-replacement property covered: the launcher replaces
// itself with the child (same pid in MOAI_SESSION_PID, the launch cwd, the
// argv verbatim), and — although the child env carries a Claude worker lane's
// identity — no launch-pending peer is registered and the run owner stays as
// recorded (REQ-CFR-008).
func TestCodexDirectPOSIXExecRegistersNoFactoryPeer(t *testing.T) {
	switch os.Getenv("T1074_CODEX_EXEC_ROLE") {
	case "launcher":
		cmd := exec.Command(os.Args[0], "-test.run=^TestCodexDirectPOSIXExecRegistersNoFactoryPeer$", "--", "arg-one", "two words")
		cmd.Dir = os.Getenv("T1074_CODEX_EXEC_ROOT")
		cmd.Env = make([]string, 0, len(os.Environ())+1)
		for _, item := range os.Environ() {
			if !strings.HasPrefix(item, "T1074_CODEX_EXEC_ROLE=") {
				cmd.Env = append(cmd.Env, item)
			}
		}
		cmd.Env = append(cmd.Env, "T1074_CODEX_EXEC_ROLE=helper", config.EnvMoaiSessionPID+"=999999")
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		if err := defaultCodexDirectLaunch(cmd); err != nil {
			t.Fatal(err)
		}
		t.Fatal("default Codex POSIX launch returned after successful exec")
	case "helper":
		pid := os.Getpid()
		if got := os.Getenv(config.EnvMoaiSessionPID); got != strconv.Itoa(pid) {
			t.Fatalf("%s=%q want %d", config.EnvMoaiSessionPID, got, pid)
		}
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		if want := os.Getenv("T1074_CODEX_EXEC_ROOT"); cwd != want {
			t.Fatalf("cwd=%q want %q", cwd, want)
		}
		at := -1
		for i, arg := range os.Args {
			if arg == "--" {
				at = i
				break
			}
		}
		if at < 0 || !reflect.DeepEqual(os.Args[at+1:], []string{"arg-one", "two words"}) {
			t.Fatalf("argv=%q", os.Args)
		}
		run := os.Getenv(config.EnvMoaiKanbanID)
		s, err := factorymsg.Open(cwd, run)
		if err != nil {
			t.Fatal(err)
		}
		closeOnCleanup(t, "factory message broker", s)
		status, err := s.Status(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(status.Lanes) != 0 {
			t.Fatalf("codex direct launch under a lane env registered peers: %+v", status.Lanes)
		}
		if pid, start := runOwnerStamp(t, cwd, run); pid != 424242 || start != "t1242-lead" {
			t.Fatalf("run owner = (%d, %q), want unchanged (424242, t1242-lead)", pid, start)
		}
		return
	}

	home := t.TempDir()
	root := t.TempDir()
	root, _ = filepath.EvalSymlinks(root)
	run := "r1"
	t.Setenv("MOAI_HOME", home)
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.RecordRun(context.Background(), homestate.FactoryRun{RunID: run, Backend: "claude", ManifestJSON: "{}", LeadPID: 424242, LeadProcessStart: "t1242-lead"}); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	env := make([]string, 0, len(os.Environ())+8)
	for _, item := range os.Environ() {
		key, _, _ := strings.Cut(item, "=")
		if key == config.EnvMoaiSessionPID || key == "T1074_CODEX_EXEC_ROLE" {
			continue
		}
		env = append(env, item)
	}
	env = append(env,
		"MOAI_HOME="+home,
		config.EnvMoaiKanbanID+"="+run,
		config.EnvMoaiKanbanBackend+"=claude",
		config.EnvMoaiFactoryWorker+"=worker-1",
		config.EnvClaudeProjectDir+"="+root,
		"T1074_CODEX_EXEC_ROOT="+root,
		"T1074_CODEX_EXEC_ROLE=launcher",
	)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCodexDirectPOSIXExecRegistersNoFactoryPeer$")
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("subprocess: %v: %s", err, out)
	}
	if ctx.Err() != nil {
		t.Fatal(ctx.Err())
	}
}

func codexPOSIXFailureFixture(t *testing.T, run string) (string, []string) {
	t.Helper()
	home := t.TempDir()
	root := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	if err := recordFactoryRunStart(root, run, "codex", ""); err != nil {
		t.Fatal(err)
	}
	env := make([]string, 0, len(os.Environ())+6)
	for _, item := range os.Environ() {
		key, _, _ := strings.Cut(item, "=")
		switch key {
		case "MOAI_HOME", config.EnvMoaiSessionPID, config.EnvMoaiKanbanID, config.EnvMoaiKanbanBackend, config.EnvMoaiFactoryWorkers, config.EnvMoaiFactoryWorker, config.EnvClaudeProjectDir:
			continue
		}
		env = append(env, item)
	}
	env = append(env,
		"MOAI_HOME="+home,
		config.EnvMoaiKanbanID+"="+run,
		config.EnvMoaiKanbanBackend+"=codex",
		config.EnvMoaiFactoryWorkers+"=1",
		config.EnvMoaiFactoryWorker+"=",
		config.EnvClaudeProjectDir+"="+root,
	)
	return root, env
}

func TestCodexDirectPOSIXRejectsIncompleteCommand(t *testing.T) {
	for _, cmd := range []*exec.Cmd{nil, {}, {Path: "/bin/true"}} {
		if err := defaultCodexDirectLaunch(cmd); err == nil || !strings.Contains(err.Error(), "command unavailable") {
			t.Fatalf("cmd %+v: error = %v, want the unavailable-command refusal", cmd, err)
		}
	}
}

func TestCodexDirectPOSIXChdirFailureDoesNotRegisterPending(t *testing.T) {
	root, env := codexPOSIXFailureFixture(t, "run-chdir-failure")
	cmd := exec.Command("/definitely/not/executed")
	cmd.Dir, cmd.Env = filepath.Join(root, "missing"), env
	if err := defaultCodexDirectLaunch(cmd); err == nil || !strings.Contains(err.Error(), "enter Codex launch directory") {
		t.Fatalf("error=%v", err)
	}
	path, err := factorymsg.BrokerPath(root, "run-chdir-failure")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("chdir failure created broker state: %v", err)
	}
}

func TestCodexDirectPOSIXExecFailureRollsBackExactPending(t *testing.T) {
	root, env := codexPOSIXFailureFixture(t, "run-exec-failure")
	before, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(before) }()
	cmd := exec.Command(filepath.Join(root, "missing-codex"))
	cmd.Dir, cmd.Env = root, env
	if err := defaultCodexDirectLaunch(cmd); err == nil {
		t.Fatal("exec failure returned nil")
	}
	s, err := factorymsg.Open(root, "run-exec-failure")
	if err != nil {
		t.Fatal(err)
	}
	closeOnCleanup(t, "factory message broker", s)
	status, err := s.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Lanes) != 0 {
		t.Fatalf("exec failure left pending rows: %+v", status.Lanes)
	}
}
