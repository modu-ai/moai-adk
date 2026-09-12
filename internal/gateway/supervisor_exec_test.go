//go:build !windows

package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

type execRegressionRecord struct {
	ChildHandoff
	PID int
}

func TestSupervisorExecRegressionExecParentProcess(t *testing.T) {
	if os.Getenv("MOAI_AUDIT_EXEC_PARENT") != "1" {
		return
	}
	env := []string{"PATH=/usr/bin:/bin", "MOAI_TEST_SUPERVISOR_HELPER=normal"}
	child, err := StartChild(context.Background(), StartOptions{Executable: os.Args[0], Args: []string{"-test.run=^TestSupervisorChildHelper$"}, Env: env, StartupTimeout: time.Second, Config: ChildConfig{Lifetime: 5 * time.Second, PollInterval: 20 * time.Millisecond}})
	if err != nil {
		os.Exit(31)
	}
	raw, _ := json.Marshal(execRegressionRecord{ChildHandoff: child.ChildHandoff, PID: child.command.Process.Pid})
	if os.WriteFile(os.Getenv("MOAI_AUDIT_HANDOFF"), raw, 0600) != nil {
		child.Stop(context.Background())
		os.Exit(32)
	}
	if syscall.Exec("/bin/sleep", []string{"sleep", "2"}, []string{"PATH=/usr/bin:/bin"}) != nil {
		child.Stop(context.Background())
		os.Exit(33)
	}
}

func TestSupervisorExecRegressionSupervisorSurvivesActualExecAndStopsAfterParentExit(t *testing.T) {
	recordPath := filepath.Join(t.TempDir(), "handoff.json")
	cmd := exec.Command(os.Args[0], "-test.run=^TestSupervisorExecRegressionExecParentProcess$")
	cmd.Env = []string{"PATH=/usr/bin:/bin", "MOAI_AUDIT_EXEC_PARENT=1", "MOAI_AUDIT_HANDOFF=" + recordPath}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	waited := make(chan error, 1)
	go func() { waited <- cmd.Wait() }()
	var child *os.Process
	t.Cleanup(func() {
		cmd.Process.Kill()
		if child != nil {
			child.Kill()
		}
	})
	var rec execRegressionRecord
	deadline := time.Now().Add(1500 * time.Millisecond)
	for time.Now().Before(deadline) {
		raw, err := os.ReadFile(recordPath)
		if err == nil && json.Unmarshal(raw, &rec) == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if rec.PID == 0 {
		t.Fatal("missing child handoff")
	}
	child, _ = os.FindProcess(rec.PID)
	time.Sleep(150 * time.Millisecond)
	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get("http://" + rec.Address)
	if err != nil {
		t.Fatalf("actual exec killed child: %v", err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	select {
	case err := <-waited:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("parent did not exit")
	}
	fp, identity := homestate.ProbeProcessIdentity(cmd.Process.Pid)
	dead, _ := os.FindProcess(cmd.Process.Pid)
	signalErr := dead.Signal(syscall.Signal(0))
	t.Logf("reaped parent identity=%v fingerprint=%q signal=%v matchesESRCH=%v matchesProcessDone=%v", identity, fp, signalErr, errors.Is(signalErr, syscall.ESRCH), errors.Is(signalErr, os.ErrProcessDone))
	deadline = time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		resp, err = client.Get("http://" + rec.Address)
		if err != nil {
			t.Log("child served after actual POSIX exec; port closed after parent exited")
			return
		}
		resp.Body.Close()
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("child listener survived parent exit")
}
