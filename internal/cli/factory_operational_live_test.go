//go:build darwin || linux

package cli

import (
	"bufio"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestFactoryLiveOperationalRosterBeforePrompt(t *testing.T) {
	requireFactoryLive(t, "operational-roster-before-prompt")
	runOperationalLauncherProof(t)
}

func TestFactoryLiveOperationalLauncherChain(t *testing.T) {
	requireFactoryLive(t, "operational-launcher-chain")
	runOperationalLauncherProof(t)
}

// This fixture never writes broker/registry rows or stamps an owner PID.
// script supplies only a terminal; every session starts via the built launcher.
func runOperationalLauncherProof(t *testing.T) {
	root := t.TempDir()
	root, _ = filepath.EvalSymlinks(root)
	bin := filepath.Join(t.TempDir(), "moai")
	build := exec.Command("go", "build", "-o", bin, "./cmd/moai")
	build.Dir = "../.."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v: %s", err, out)
	}
	data, err := os.ReadFile(bin)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("BUILT_BINARY %s sha256=%x", bin, sha256.Sum256(data))
	for _, dir := range []string{".moai", ".codex"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	hooks, err := codexwiring.RenderHooks(nil)
	if err != nil {
		t.Fatal(err)
	}
	for path, body := range map[string][]byte{
		".codex/hooks.json":  hooks,
		".codex/config.toml": codexwiring.EnsureMCPTable(nil),
		"AGENTS.md":          []byte("Synthetic factory operational acceptance fixture. Do not inspect anything outside this fixture.\n"),
	} {
		if err := os.WriteFile(filepath.Join(root, path), body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	env := make([]string, 0, len(os.Environ()))
	for _, item := range os.Environ() {
		key, _, _ := strings.Cut(item, "=")
		if strings.HasPrefix(key, "MOAI_") || key == "CLAUDE_CODE_SESSION_ID" || key == "CLAUDE_PROJECT_DIR" || key == "PATH" {
			continue
		}
		env = append(env, item)
	}
	env = append(env, "MOAI_HOME="+t.TempDir(), "PATH="+filepath.Dir(bin)+string(os.PathListSeparator)+os.Getenv("PATH"), "CLAUDE_PROJECT_DIR="+root, "TERM=xterm-256color")
	// Read-only broker helpers resolve the same isolated namespace as children.
	for _, item := range env {
		if strings.HasPrefix(item, "MOAI_HOME=") {
			t.Setenv("MOAI_HOME", strings.TrimPrefix(item, "MOAI_HOME="))
		}
	}
	var terminals []*operationalTerminal
	for _, args := range [][]string{{"codex", "-f"}, {"codex", "-f", "agent"}, {"codex", "-f", "agent"}} {
		term := startOperationalTerminal(t, root, bin, args, env)
		terminals = append(terminals, term)
		want := len(terminals)
		deadline := time.Now().Add(30 * time.Second)
		var lastErr error
		for time.Now().Before(deadline) {
			run, lanes, err := operationalRegisteredLanes(root)
			lastErr = err
			if err == nil && len(lanes) == want {
				t.Logf("PRE_PROMPT run=%s peers=%d", run, len(lanes))
				break
			}
			// Terminal status query is terminal protocol, not a model prompt.
			if strings.Contains(term.output(), "\x1b[6n") {
				_, _ = term.stdin.Write([]byte("\x1b[1;1R"))
			}
			time.Sleep(100 * time.Millisecond)
		}
		_, lanes, err := operationalRegisteredLanes(root)
		if err != nil || len(lanes) != want {
			t.Fatalf("GAP: production SessionStart before prompt: want=%d got=%d err=%v last=%v terminal=%q", want, len(lanes), err, lastErr, term.output())
		}
	}
	run, lanes, err := operationalRegisteredLanes(root)
	if err != nil {
		t.Fatal(err)
	}
	for i, slot := range []string{"agent-1", "agent-2", "lead"} {
		lane := lanes[i]
		terminalIndex := map[string]int{"lead": 0, "agent-1": 1, "agent-2": 2}[slot]
		children, err := operationalDescendants(terminals[terminalIndex].pid)
		if err != nil || children[lane.PID] == "" {
			t.Fatalf("owner is not a descendant of its exact production launcher: slot=%s pid=%d err=%v", slot, lane.PID, err)
		}
		fp, state := homestate.ProbeProcessIdentity(lane.PID)
		if lane.Slot != slot || lane.Backend != "codex" || lane.SessionUUID == "" || lane.Generation < 1 || state != homestate.ProcessIdentityLive || fp != lane.ProcessStart || lane.TaskState != "unknown" {
			t.Fatalf("owner mismatch: lane=%+v probe=%s/%s", lane, state, fp)
		}
		out, err := exec.Command("ps", "-p", strconv.Itoa(lane.PID), "-o", "pid=,ppid=,command=").Output()
		if err != nil || !strings.Contains(string(out), "codex") {
			t.Fatalf("not a Codex owner: %s %v", out, err)
		}
		t.Logf("OWNER %s registered=%s measured=%s tree=%s", slot, lane.ProcessStart, fp, out)
	}
	t.Log("PRODUCTION_ARGV_OK moai codex -f | moai codex -f agent | moai codex -f agent")
	t.Log("SESSIONSTART_OWNER_MATCH_OK lead,agent-1,agent-2")
	lead := lanes[2]
	oldMCP := operationalOwnedMCP(t, lead.PID, bin)
	if err := syscall.Kill(oldMCP, syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	// The real client must restart its server; the test never launches an MCP
	// substitute. If Codex cannot reconnect, this is a failed LIVE gate.
	before := factoryBrokerSnapshot(t, root, run)
	prompt := fmt.Sprintf("Call only the registered factory_msg_status tool twice with run_id=%q. If the server disconnected, reconnect your moai MCP and retry. Do not call list/send/receipt or any other tool. Report the returned operational lanes unchanged.", run)
	if _, err := terminals[0].stdin.Write([]byte(prompt + "\r")); err != nil {
		t.Fatal(err)
	}
	got := waitOperationalMCPResult(t, lead.SessionUUID, run, 2)
	assertOperationalRoster(t, lanes, got)
	newMCP := operationalOwnedMCP(t, lead.PID, bin)
	if oldMCP == newMCP {
		t.Fatal("MCP PID did not change")
	}
	if stateFP, state := homestate.ProbeProcessIdentity(oldMCP); state != homestate.ProcessIdentityDead {
		t.Fatalf("old MCP not dead: %d %s %s", oldMCP, state, stateFP)
	}
	if before != factoryBrokerSnapshot(t, root, run) {
		t.Fatal("MCP status mutated broker rows")
	}
	t.Logf("MCP_RESTART_OK old=%d new=%d binary=%s sha256=%x", oldMCP, newMCP, bin, sha256.Sum256(data))
	t.Log("LEAD_MCP_ROSTER_OK lead,agent-1,agent-2")
	t.Log("REAL_SESSION_ROSTER_OK lead,agent-1,agent-2")
	// Phase 2 is a separate project; its peer is registered by production
	// SessionStart, not by RegisterPeer or SQL seeding in this fixture.
	other := t.TempDir()
	other, _ = filepath.EvalSymlinks(other)
	for _, dir := range []string{".moai", ".codex"} {
		if err := os.MkdirAll(filepath.Join(other, dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{".codex/hooks.json", ".codex/config.toml", "AGENTS.md"} {
		body, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(other, name), body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	otherEnv := append([]string(nil), env...)
	for i, v := range otherEnv {
		if strings.HasPrefix(v, "CLAUDE_PROJECT_DIR=") {
			otherEnv[i] = "CLAUDE_PROJECT_DIR=" + other
		}
	}
	t.Log("PHASE_2 secondary production fixture; phase-1 primary exact 1+2 proof is complete")
	secondary := startOperationalTerminal(t, other, bin, []string{"codex", "-f"}, otherEnv)
	var foreign []factorymsg.LaneStatus
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		_, foreign, err = operationalRegisteredLanes(other)
		if err == nil && len(foreign) == 1 {
			break
		}
		if strings.Contains(secondary.output(), "\x1b[6n") {
			_, _ = secondary.stdin.Write([]byte("\x1b[1;1R"))
		}
		time.Sleep(100 * time.Millisecond)
	}
	if len(foreign) != 1 {
		t.Fatalf("secondary production peer absent: %v %s", err, secondary.output())
	}
	_, _ = terminals[0].stdin.Write([]byte(prompt + "\r"))
	got = waitOperationalMCPResult(t, lead.SessionUUID, run, 4)
	assertOperationalRoster(t, lanes, got)
	for _, l := range got {
		if l.SessionUUID == foreign[0].SessionUUID || l.PID == foreign[0].PID {
			t.Fatal("foreign endpoint leaked")
		}
	}
	t.Logf("PROJECT_ISOLATION_OK primary=%s secondary=%s foreign_session=%s", homestate.ProjectKey(root), homestate.ProjectKey(other), foreign[0].SessionUUID)
	if err := syscall.Kill(lanes[0].PID, syscall.SIGKILL); err != nil {
		t.Fatal(err)
	}
	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		_, state := homestate.ProbeProcessIdentity(lanes[0].PID)
		if state == homestate.ProcessIdentityDead {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	_, dead, err := operationalRegisteredLanes(root)
	if err != nil || dead[0].EndpointState != "dead" {
		t.Fatalf("real terminated owner not dead: %+v %v", dead, err)
	}
	t.Log("DEAD_OWNER_OK")
	t.Log("NO_BYPASS_OK")
	// AC-OPS-002 stale evidence lives in TestFactoryLaneRosterStateTruth:
	// a real live OS owner plus a deliberately mismatched auxiliary fingerprint.
	// Such auxiliary rows never participate in this production-chain proof.
}

func assertOperationalRoster(t *testing.T, want, got []factorymsg.LaneStatus) {
	t.Helper()
	if len(want) != len(got) {
		t.Fatalf("MCP roster size=%d want=%d", len(got), len(want))
	}
	for i := range want {
		a, b := want[i], got[i]
		a.ObservedAt = time.Time{}
		b.ObservedAt = time.Time{}
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("MCP endpoint mismatch want=%+v got=%+v", a, b)
		}
	}
}

func operationalOwnedMCP(t *testing.T, owner int, bin string) int {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		out, err := exec.Command("ps", "-axo", "pid=,ppid=,command=").Output()
		if err != nil {
			t.Fatalf("process tree unavailable: %v", err)
		}
		for _, line := range strings.Split(string(out), "\n") {
			fields := strings.Fields(line)
			if len(fields) < 4 {
				continue
			}
			pid, _ := strconv.Atoi(fields[0])
			ppid, _ := strconv.Atoi(fields[1])
			if ppid == owner && filepath.Base(fields[2]) == "moai" && fields[3] == "mcp-server" {
				if runtime.GOOS == "linux" {
					actual, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid))
					if err != nil || actual != bin {
						t.Fatalf("MCP binary identity %q: %v", actual, err)
					}
				} else {
					loaded, err := exec.Command("lsof", "-a", "-p", strconv.Itoa(pid), "-d", "txt", "-Fn").Output()
					if err != nil || !strings.Contains(string(loaded), "n"+bin+"\n") {
						t.Fatalf("MCP loaded binary identity unavailable: %v %s", err, loaded)
					}
				}
				t.Logf("LEAD_OWNED_MCP %s", line)
				return pid
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("no new built-tree MCP directly owned by the lead")
	return 0
}

func waitOperationalMCPResult(t *testing.T, sessionID, run string, minimum int) []factorymsg.LaneStatus {
	t.Helper()
	home := os.Getenv("CODEX_HOME")
	if home == "" {
		base, err := os.UserHomeDir()
		if err != nil {
			t.Fatal(err)
		}
		home = filepath.Join(base, ".codex")
	}
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		paths, err := filepath.Glob(filepath.Join(home, "sessions", "*", "*", "*", "*"+sessionID+"*.jsonl"))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range paths {
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			if lanes, ok := operationalMCPResultCount(data, run, minimum); ok {
				return lanes
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatal("lead session rollout has no attributable factory_msg_status call/result")
	return nil
}

func operationalMCPResult(data []byte, run string) ([]factorymsg.LaneStatus, bool) {
	return operationalMCPResultCount(data, run, 1)
}

func operationalMCPResultCount(data []byte, run string, minimum int) ([]factorymsg.LaneStatus, bool) {
	calls := map[string]bool{}
	seen := map[string]bool{}
	scan := bufio.NewScanner(strings.NewReader(string(data)))
	scan.Buffer(make([]byte, 4096), 4<<20)
	var latest []factorymsg.LaneStatus
	for scan.Scan() {
		var event struct {
			Type    string `json:"type"`
			Payload struct {
				Type, Name, Arguments, Output string
				CallID                        string `json:"call_id"`
			} `json:"payload"`
		}
		if json.Unmarshal(scan.Bytes(), &event) != nil || event.Type != "response_item" {
			continue
		}
		p := event.Payload
		if p.Type == "function_call" && p.CallID != "" && (p.Name == "mcp__moai__factory_msg_status" || p.Name == "functions.mcp__moai__factory_msg_status") {
			var args struct {
				Run string `json:"run_id"`
			}
			if json.Unmarshal([]byte(p.Arguments), &args) == nil && args.Run == run {
				calls[p.CallID] = true
			}
		}
		if p.Type == "function_call_output" && p.CallID != "" && calls[p.CallID] {
			if lanes, ok := operationalOutputLanes([]byte(p.Output)); ok {
				latest = lanes
				seen[p.CallID] = true
			}
		}
	}
	return latest, latest != nil && len(seen) >= minimum
}

func operationalOutputLanes(data []byte) ([]factorymsg.LaneStatus, bool) {
	var v struct {
		IsError    bool                    `json:"isError"`
		Lanes      []factorymsg.LaneStatus `json:"lanes"`
		Structured json.RawMessage         `json:"structuredContent"`
		Content    []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if json.Unmarshal(data, &v) != nil || v.IsError {
		return nil, false
	}
	if v.Lanes != nil {
		return v.Lanes, true
	}
	if len(v.Structured) > 0 {
		if lanes, ok := operationalOutputLanes(v.Structured); ok {
			return lanes, true
		}
	}
	for _, c := range v.Content {
		if lanes, ok := operationalOutputLanes([]byte(c.Text)); ok {
			return lanes, true
		}
	}
	return nil, false
}

type operationalTerminal struct {
	mu    sync.Mutex
	text  strings.Builder
	stdin *os.File
	pid   int
}

func (b *operationalTerminal) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.text.Write(p)
}
func (b *operationalTerminal) output() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	s := b.text.String()
	if len(s) > 6000 {
		return s[len(s)-6000:]
	}
	return s
}

func startOperationalTerminal(t *testing.T, root, bin string, args, env []string) *operationalTerminal {
	t.Helper()
	argv := append([]string{"-q", "/dev/null", bin}, args...)
	if runtime.GOOS == "linux" {
		// Quote generated absolute binary path, never interpolate model/user data.
		argv = []string{"-q", "-c", "'" + strings.ReplaceAll(bin, "'", "'\\''") + "' " + strings.Join(args, " "), "/dev/null"}
	}
	cmd := exec.Command("script", argv...)
	cmd.Dir, cmd.Env = root, env
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	b := &operationalTerminal{stdin: w}
	cmd.Stdin, cmd.Stdout, cmd.Stderr = r, b, b
	if err := cmd.Start(); err != nil {
		r.Close()
		w.Close()
		t.Fatal(err)
	}
	b.pid = cmd.Process.Pid
	r.Close()
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	stopInventory := make(chan struct{})
	inventoryDone := make(chan struct{})
	tracked := map[int]string{}
	var inventoryErr error
	go func() {
		defer close(inventoryDone)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stopInventory:
				return
			case <-ticker.C:
				owned, err := operationalDescendants(cmd.Process.Pid)
				if err != nil {
					inventoryErr = err
					continue
				}
				for pid, fp := range owned {
					tracked[pid] = fp
				}
			}
		}
	}()
	t.Cleanup(func() {
		close(stopInventory)
		<-inventoryDone
		if inventoryErr != nil {
			t.Errorf("cleanup continuous inventory unavailable: %v", inventoryErr)
		}
		owned, err := operationalDescendants(cmd.Process.Pid)
		if err != nil {
			t.Errorf("cleanup process inventory unavailable: %v", err)
		}
		if owned == nil {
			owned = make(map[int]string)
		}
		for pid, fp := range tracked {
			if _, ok := owned[pid]; !ok {
				owned[pid] = fp
			}
		}
		w.Close()
		for pid, fp := range owned {
			actual, state := homestate.ProbeProcessIdentity(pid)
			if state == homestate.ProcessIdentityLive && fp == actual {
				_ = syscall.Kill(pid, syscall.SIGKILL)
			}
		}
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Errorf("cleanup timeout pid=%d", cmd.Process.Pid)
		}
		if err := syscall.Kill(-cmd.Process.Pid, 0); err != syscall.ESRCH {
			t.Errorf("process group remains: %d %v", cmd.Process.Pid, err)
		}
		for pid, fp := range owned {
			actual, state := homestate.ProbeProcessIdentity(pid)
			if state != homestate.ProcessIdentityDead && actual == fp {
				t.Errorf("owned child remains: %d %s", pid, state)
			}
		}
		if !t.Failed() {
			t.Logf("CLEANUP_OK terminal_pid=%d descendants=%d", cmd.Process.Pid, len(owned))
		}
	})
	t.Logf("PRODUCTION_ARGV %s %s terminal_pid=%d", bin, strings.Join(args, " "), cmd.Process.Pid)
	return b
}

func operationalDescendants(root int) (map[int]string, error) {
	out, err := exec.Command("ps", "-axo", "pid=,ppid=").Output()
	if err != nil {
		return nil, err
	}
	parents := map[int]int{}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		pid, _ := strconv.Atoi(fields[0])
		ppid, _ := strconv.Atoi(fields[1])
		parents[pid] = ppid
	}
	owned := map[int]string{}
	for pid := range parents {
		for at, steps := pid, 0; at > 1 && steps < len(parents); steps++ {
			if at == root {
				fp, state := homestate.ProbeProcessIdentity(pid)
				if state == homestate.ProcessIdentityLive {
					owned[pid] = fp
				}
				break
			}
			at = parents[at]
		}
	}
	return owned, nil
}

func operationalRegisteredLanes(root string) (string, []factorymsg.LaneStatus, error) {
	path, err := homestate.FactoryDBPath(root)
	if err != nil {
		return "", nil, err
	}
	if _, err = os.Stat(path); err != nil {
		return "", nil, err
	}
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?mode=ro")
	if err != nil {
		return "", nil, err
	}
	defer db.Close()
	var run string
	if err = db.QueryRow(`SELECT run_id FROM runs WHERE status='active'`).Scan(&run); err != nil {
		return "", nil, err
	}
	s, err := factorymsg.OpenExistingWithDeadline(root, run, 200*time.Millisecond)
	if err != nil {
		return run, nil, err
	}
	defer s.Close()
	st, err := s.Status(context.Background())
	return run, st.Lanes, err
}
