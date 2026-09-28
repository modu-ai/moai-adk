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
		config.EnvMoaiFactoryWorker+"=lane-1",
		config.EnvClaudeProjectDir+"="+root,
		"T1074_CODEX_EXEC_ROOT="+root,
		"T1074_CODEX_EXEC_ROLE=launcher",
		// The child re-executes this binary, so its TestMain runs again: pin
		// the composed factory family or the t1252 ambient clear empties
		// MOAI_KANBAN_ID before the helper role reads it (factory_test.go).
		factoryEnvPinnedEnv+"=1",
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

// codexPOSIXFailureFixture returns a project root and a child env under an
// isolated MOAI_HOME. It carries no lane identity: the codex direct launch
// writes no factory state on any path (SPEC-CODEX-FACTORY-RETIRE-001), so the
// failure cells below measure the launch itself.
func codexPOSIXFailureFixture(t *testing.T) (string, []string) {
	t.Helper()
	home := t.TempDir()
	root := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	env := make([]string, 0, len(os.Environ())+1)
	for _, item := range os.Environ() {
		key, _, _ := strings.Cut(item, "=")
		if key == "MOAI_HOME" || key == config.EnvMoaiSessionPID {
			continue
		}
		env = append(env, item)
	}
	return root, append(env, "MOAI_HOME="+home)
}

func TestCodexDirectPOSIXRejectsIncompleteCommand(t *testing.T) {
	for _, cmd := range []*exec.Cmd{nil, {}, {Path: "/bin/true"}} {
		if err := defaultCodexDirectLaunch(cmd); err == nil || !strings.Contains(err.Error(), "command unavailable") {
			t.Fatalf("cmd %+v: error = %v, want the unavailable-command refusal", cmd, err)
		}
	}
}

func TestCodexDirectPOSIXChdirFailureIsReported(t *testing.T) {
	root, env := codexPOSIXFailureFixture(t)
	before, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("/definitely/not/executed")
	cmd.Dir, cmd.Env = filepath.Join(root, "missing"), env
	if err := defaultCodexDirectLaunch(cmd); err == nil || !strings.Contains(err.Error(), "enter Codex launch directory") {
		t.Fatalf("error=%v", err)
	}
	if after, _ := os.Getwd(); after != before {
		t.Fatalf("a failed chdir moved the process to %q", after)
	}
}

func TestCodexDirectPOSIXExecFailureIsReported(t *testing.T) {
	root, env := codexPOSIXFailureFixture(t)
	before, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(before) }()
	cmd := exec.Command(filepath.Join(root, "missing-codex"))
	cmd.Dir, cmd.Env = root, env
	if err := defaultCodexDirectLaunch(cmd); err == nil || !strings.Contains(err.Error(), "exec Codex") {
		t.Fatalf("error=%v, want the exec failure", err)
	}
}
