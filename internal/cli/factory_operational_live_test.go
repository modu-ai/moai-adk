//go:build darwin || linux

package cli

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
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
	runOperationalLauncherProof(t, false)
}

func TestFactoryLiveOperationalLauncherChain(t *testing.T) {
	requireFactoryLive(t, "operational-launcher-chain")
	runOperationalLauncherProof(t, true)
}

// This fixture never writes broker/registry rows or stamps an owner PID.
// script supplies only a terminal; every session starts via the built launcher.
func runOperationalLauncherProof(t *testing.T, proveBoundChain bool) {
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
	env := make([]string, 0, len(os.Environ()))
	for _, item := range os.Environ() {
		key, _, _ := strings.Cut(item, "=")
		if strings.HasPrefix(key, "MOAI_") || key == "CLAUDE_CODE_SESSION_ID" || key == "CLAUDE_PROJECT_DIR" || key == "CLAUDE_CONFIG_DIR" || key == "PATH" {
			continue
		}
		env = append(env, item)
	}
	env = append(env, "MOAI_HOME="+t.TempDir(), "CLAUDE_CONFIG_DIR="+t.TempDir(), "PATH="+filepath.Dir(bin)+string(os.PathListSeparator)+os.Getenv("PATH"), "CLAUDE_PROJECT_DIR="+root, "TERM=xterm-256color")
	env = operationalHookShellEnv(t, bin, env)
	if err := prepareOperationalProject(root, bin, env); err != nil {
		t.Fatal(err)
	}
	// Read-only broker helpers resolve the same isolated namespace as children.
	for _, item := range env {
		if strings.HasPrefix(item, "MOAI_HOME=") {
			t.Setenv("MOAI_HOME", strings.TrimPrefix(item, "MOAI_HOME="))
		}
	}
	var terminals []*operationalTerminal
	bootstrapMCP := bootstrapOperationalTrust(t, root, bin, env)
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
			if err := term.acceptFixtureDirectoryTrust(); err != nil {
				t.Fatal(err)
			}
			if err := term.acceptFixtureHookTrust(); err != nil {
				t.Fatal(err)
			}
			time.Sleep(100 * time.Millisecond)
		}
		_, lanes, err := operationalRegisteredLanes(root)
		if err != nil || len(lanes) != want {
			t.Fatalf("GAP: production launch-pending before prompt: want=%d got=%d err=%v last=%v terminal=%q", want, len(lanes), err, lastErr, term.output())
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
		if lane.Slot != slot || lane.Backend != "codex" || lane.BindingState != factorymsg.BindingLaunchPending || lane.SessionUUID != "" || lane.Generation < 1 || state != homestate.ProcessIdentityLive || fp != lane.ProcessStart || lane.TaskState != "unknown" {
			t.Fatalf("owner mismatch: lane=%+v probe=%s/%s", lane, state, fp)
		}
		out, err := exec.Command("ps", "-p", strconv.Itoa(lane.PID), "-o", "pid=,ppid=,command=").Output()
		if err != nil || !strings.Contains(string(out), "codex") {
			t.Fatalf("not a Codex owner: %s %v", out, err)
		}
		t.Logf("OWNER %s registered=%s measured=%s tree=%s", slot, lane.ProcessStart, fp, out)
	}
	t.Log("PRODUCTION_ARGV_OK moai codex -f | moai codex -f agent | moai codex -f agent")
	t.Log("LAUNCH_PENDING_ROSTER_OK lead,agent-1,agent-2")
	t.Log("NO_SESSION_UUID_BEFORE_TURN_OK")
	t.Log("NO_BYPASS_OK")
	if !proveBoundChain {
		return
	}
	if err := waitOperationalPromptReady(terminals, 45*time.Second); err != nil {
		t.Fatalf("Codex prompt composer not ready: %v", err)
	}
	t.Log("PROMPT_COMPOSER_READY_OK lead,agent-1,agent-2")
	for i, terminal := range terminals {
		prompt := fmt.Sprintf("Reply with exactly FACTORY_READY_%d and do not call any tool.", i+1)
		if err := submitOperationalPrompt(terminal, prompt, 10*time.Second); err != nil {
			t.Fatalf("submit first prompt to terminal %d: %v", i, err)
		}
		deadline := time.Now().Add(60 * time.Second)
		for time.Now().Before(deadline) {
			_, current, readErr := operationalRegisteredLanes(root)
			if readErr == nil && len(current) == 3 {
				slot := []string{"lead", "agent-1", "agent-2"}[i]
				for _, lane := range current {
					if lane.Slot == slot && lane.BindingState == factorymsg.BindingBound && lane.SessionUUID != "" {
						goto rebound
					}
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
		t.Fatalf("UserPromptSubmit did not rebind terminal %d: %q", i, terminal.output())
	rebound:
	}
	_, lanes, err = operationalRegisteredLanes(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, lane := range lanes {
		if lane.BindingState != factorymsg.BindingBound || lane.SessionUUID == "" || lane.Generation < 2 {
			t.Fatalf("lane not rebound: %+v", lane)
		}
	}
	t.Log("USERPROMPT_REBIND_OK lead,agent-1,agent-2")
	t.Log("BOUND_ROSTER_OK lead,agent-1,agent-2")
	stableBefore := append([]factorymsg.LaneStatus(nil), lanes...)
	for i := range stableBefore {
		stableBefore[i].ObservedAt = time.Time{}
	}
	for i, terminal := range terminals {
		answer := fmt.Sprintf("BOUND_STABLE_%d", i+1)
		prompt := fmt.Sprintf("Reply with exactly %s and do not call any tool.", answer)
		if err := submitOperationalPrompt(terminal, prompt, 10*time.Second); err != nil {
			t.Fatalf("submit bound prompt to terminal %d: %v", i, err)
		}
		slot := []string{"lead", "agent-1", "agent-2"}[i]
		var sessionID string
		for _, lane := range stableBefore {
			if lane.Slot == slot {
				sessionID = lane.SessionUUID
				break
			}
		}
		if sessionID == "" {
			t.Fatalf("bound session absent for %s", slot)
		}
		waitOperationalAssistantOutput(t, sessionID, answer)
	}
	_, stableAfter, err := operationalRegisteredLanes(root)
	if err != nil {
		t.Fatal(err)
	}
	for i := range stableAfter {
		stableAfter[i].ObservedAt = time.Time{}
	}
	if !reflect.DeepEqual(stableBefore, stableAfter) {
		t.Fatalf("bound follow-up prompt rewrote peers: before=%+v after=%+v", stableBefore, stableAfter)
	}
	t.Log("BOUND_PROMPT_NO_REWRITE_OK lead,agent-1,agent-2")
	lead := lanes[2]
	leadMCP := operationalOwnedMCP(t, lead.PID, bin)
	if bootstrapMCP == leadMCP {
		t.Fatalf("production lead reused bootstrap MCP pid=%d", leadMCP)
	}
	// The production lead's freshly spawned MCP remains live for its session.
	// The bootstrap MCP was terminated before any measured production launch.
	before := factoryBrokerSnapshot(t, root, run)
	prompt := fmt.Sprintf("Call only the registered factory_msg_status tool twice with run_id=%q. Do not call list/send/receipt or any other tool. Report the returned operational lanes unchanged.", run)
	if err := submitOperationalPrompt(terminals[0], prompt, 10*time.Second); err != nil {
		t.Fatal(err)
	}
	got := waitOperationalMCPResult(t, lead.SessionUUID, run, 2)
	assertOperationalRoster(t, lanes, got)
	if currentMCP := operationalOwnedMCP(t, lead.PID, bin); currentMCP != leadMCP {
		t.Fatalf("lead MCP changed during status calls: before=%d after=%d", leadMCP, currentMCP)
	}
	if before != factoryBrokerSnapshot(t, root, run) {
		t.Fatal("MCP status mutated broker rows")
	}
	t.Logf("MCP_RESTART_OK old=%d new=%d binary=%s sha256=%x", bootstrapMCP, leadMCP, bin, sha256.Sum256(data))
	t.Log("LEAD_MCP_ROSTER_OK lead,agent-1,agent-2")
	// Phase 2 is a separate project; its peer is registered by production
	// SessionStart, not by RegisterPeer or SQL seeding in this fixture.
	other := t.TempDir()
	other, _ = filepath.EvalSymlinks(other)
	otherEnv := append([]string(nil), env...)
	for i, v := range otherEnv {
		if strings.HasPrefix(v, "CLAUDE_PROJECT_DIR=") {
			otherEnv[i] = "CLAUDE_PROJECT_DIR=" + other
		}
	}
	if err := prepareOperationalProject(other, bin, otherEnv); err != nil {
		t.Fatal(err)
	}
	_ = bootstrapOperationalTrust(t, other, bin, otherEnv)
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
		if err := secondary.acceptFixtureDirectoryTrust(); err != nil {
			t.Fatal(err)
		}
		if err := secondary.acceptFixtureHookTrust(); err != nil {
			t.Fatal(err)
		}
		time.Sleep(100 * time.Millisecond)
	}
	if len(foreign) != 1 {
		t.Fatalf("secondary production peer absent: %v %s", err, secondary.output())
	}
	if foreign[0].BindingState != factorymsg.BindingLaunchPending || foreign[0].SessionUUID != "" {
		t.Fatalf("secondary pre-turn endpoint was not provisional: %+v", foreign[0])
	}
	if err := submitOperationalPrompt(terminals[0], prompt, 10*time.Second); err != nil {
		t.Fatal(err)
	}
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
	t.Log("NO_SYNTHETIC_TURN_OR_BYPASS_OK")
	// AC-OPS-002 stale evidence lives in TestFactoryLaneRosterStateTruth:
	// a real live OS owner plus a deliberately mismatched auxiliary fingerprint.
	// Such auxiliary rows never participate in this production-chain proof.
}

// operationalHookShellEnv keeps the production bare `moai hook ...` wiring
// while making Codex's login-shell lookup resolve the same binary that launched
// the fixture. A login shell may rebuild PATH from user startup files and pick
// an installed, stale moai even when the fixture prepended its build directory.
func operationalHookShellEnv(t *testing.T, bin string, env []string) []string {
	t.Helper()
	shell, err := exec.LookPath("zsh")
	if err != nil {
		t.Fatalf("Codex hook-shell fixture requires zsh: %v", err)
	}
	zdotdir := t.TempDir()
	profile := "export PATH=" + shellQuote(filepath.Dir(bin)) + ":\"$PATH\"\n"
	if err := os.WriteFile(filepath.Join(zdotdir, ".zprofile"), []byte(profile), 0600); err != nil {
		t.Fatal(err)
	}
	filtered := make([]string, 0, len(env)+2)
	for _, item := range env {
		key, _, _ := strings.Cut(item, "=")
		if key != "SHELL" && key != "ZDOTDIR" {
			filtered = append(filtered, item)
		}
	}
	filtered = append(filtered, "SHELL="+shell, "ZDOTDIR="+zdotdir)
	resolve := exec.Command(shell, "-lc", "command -v moai")
	resolve.Env = filtered
	out, err := resolve.Output()
	if err != nil {
		t.Fatalf("resolve Codex hook binary: %v", err)
	}
	got, err := filepath.EvalSymlinks(strings.TrimSpace(string(out)))
	if err != nil {
		t.Fatalf("canonicalize resolved Codex hook binary: %v", err)
	}
	want, err := filepath.EvalSymlinks(bin)
	if err != nil {
		t.Fatalf("canonicalize built launcher: %v", err)
	}
	if got != want {
		t.Fatalf("Codex login-shell hook resolved %q want built launcher %q", got, want)
	}
	t.Logf("HOOK_BINARY_IDENTITY_OK %s", want)
	return filtered
}

func prepareOperationalProject(root, bin string, env []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "init", "--non-interactive", "--llm", "gpt", "--no-hooks", "--root", root)
	cmd.Dir, cmd.Env = root, env
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("production init: %w: %s", err, out)
	}
	return nil
}

func operationalBootstrapArgs() []string { return []string{"codex"} }
func operationalBootstrapReady(output string, directoryTrusted, hooksTrusted bool) bool {
	if !directoryTrusted || !hooksTrusted {
		return false
	}
	text := strings.Join(strings.Fields(operationalANSI.ReplaceAllString(output, " ")), " ")
	at := strings.LastIndex(text, "Trusting hooks...")
	return at >= 0 && strings.Contains(text[at+len("Trusting hooks..."):], "Ask Codex to do anything")
}

func operationalPromptReady(output string) bool {
	text := strings.Join(strings.Fields(operationalANSI.ReplaceAllString(output, " ")), " ")
	ready := strings.LastIndex(text, "Ask Codex to do anything")
	if ready < 0 {
		return false
	}
	return ready > strings.LastIndex(text, "model: loading") &&
		ready > strings.LastIndex(text, "directory: loading") &&
		ready > strings.LastIndex(text, "Trusting hooks...")
}

func waitOperationalPromptReady(terminals []*operationalTerminal, limit time.Duration) error {
	deadline := time.Now().Add(limit)
	stableSince := make([]time.Time, len(terminals))
	for time.Now().Before(deadline) {
		allReady := true
		for i, terminal := range terminals {
			if strings.Contains(terminal.output(), "\x1b[6n") {
				if _, err := terminal.stdin.Write([]byte("\x1b[1;1R")); err != nil {
					return err
				}
			}
			if err := terminal.acceptFixtureDirectoryTrust(); err != nil {
				return err
			}
			if err := terminal.acceptFixtureHookTrust(); err != nil {
				return err
			}
			if !operationalPromptReady(terminal.output()) {
				stableSince[i] = time.Time{}
				allReady = false
				continue
			}
			if stableSince[i].IsZero() {
				stableSince[i] = time.Now()
			}
			if time.Since(stableSince[i]) < 500*time.Millisecond {
				allReady = false
			}
		}
		if allReady {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return errors.New("prompt composer readiness deadline exceeded")
}

func submitOperationalPrompt(terminal *operationalTerminal, prompt string, limit time.Duration) error {
	if terminal == nil || terminal.stdin == nil {
		return errors.New("terminal input unavailable")
	}
	before := terminal.output()
	if _, err := terminal.stdin.Write([]byte(prompt)); err != nil {
		return fmt.Errorf("write prompt text: %w", err)
	}
	deadline := time.Now().Add(limit)
	settledAt := time.Now().Add(500 * time.Millisecond)
	activity := false
	for time.Now().Before(deadline) {
		if terminal.output() != before {
			activity = true
		}
		if terminal.pid > 0 {
			_, state := homestate.ProbeProcessIdentity(terminal.pid)
			if state != homestate.ProcessIdentityLive {
				return fmt.Errorf("terminal process is not live during prompt handoff: pid=%d state=%s", terminal.pid, state)
			}
		}
		if activity && !time.Now().Before(settledAt) {
			if _, err := terminal.stdin.Write([]byte("\r")); err != nil {
				return fmt.Errorf("submit settled prompt: %w", err)
			}
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	tail := terminal.output()
	if len(tail) > 1200 {
		tail = tail[len(tail)-1200:]
	}
	return fmt.Errorf("no terminal output activity after prompt text write; terminal_tail=%q", tail)
}

func finishOperationalBootstrap(poll func() (bool, error), stop, check func() error, limit time.Duration) error {
	deadline := time.Now().Add(limit)
	for {
		ready, err := poll()
		if err != nil {
			return errors.Join(err, stop())
		}
		if ready {
			if err := stop(); err != nil {
				return err
			}
			return check()
		}
		if !time.Now().Before(deadline) {
			return errors.Join(errors.New("trust bootstrap did not reach post-trust idle"), stop())
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func finishOperationalBootstrapCapture(poll func() (bool, error), capture func() (int, error), stop func() error, check func(int) error, limit time.Duration) (int, error) {
	deadline := time.Now().Add(limit)
	for {
		ready, err := poll()
		if err != nil {
			return 0, errors.Join(err, stop())
		}
		if ready {
			pid, err := capture()
			if err != nil {
				return 0, errors.Join(err, stop())
			}
			if err := stop(); err != nil {
				return 0, err
			}
			if err := check(pid); err != nil {
				return 0, err
			}
			return pid, nil
		}
		if !time.Now().Before(deadline) {
			return 0, errors.Join(errors.New("trust bootstrap did not reach post-trust idle"), stop())
		}
		time.Sleep(100 * time.Millisecond)
	}
}
func operationalNoBroker(root string) error {
	dir, err := homestate.FactoryDir(root)
	if err != nil {
		return err
	}
	for _, path := range []string{filepath.Join(dir, "factory.db"), filepath.Join(dir, "messages")} {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("factory state exists before measured launch: %s", path)
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func bootstrapOperationalTrust(t *testing.T, root, bin string, env []string) int {
	t.Helper()
	// The observed 0.155.1 run reached post-trust idle without a factory
	// registration during its 30s window. This preparation does not assert a
	// universal hook-replay rule: fresh measured launches must prove registration.
	if err := operationalNoBroker(root); err != nil {
		t.Fatal(err)
	}
	term := startOperationalTerminal(t, root, bin, operationalBootstrapArgs(), env)
	var capturedFingerprint string
	pid, err := finishOperationalBootstrapCapture(func() (bool, error) {
		if strings.Contains(term.output(), "\x1b[6n") {
			if _, err := term.stdin.Write([]byte("\x1b[1;1R")); err != nil {
				return false, err
			}
		}
		if err := term.acceptFixtureDirectoryTrust(); err != nil {
			return false, err
		}
		if err := term.acceptFixtureHookTrust(); err != nil {
			return false, err
		}
		term.mu.Lock()
		defer term.mu.Unlock()
		return operationalBootstrapReady(term.text.String(), term.trust.sent, term.hookTrust.sent), nil
	}, func() (int, error) {
		pid := operationalTerminalOwnedMCP(t, term.pid, bin)
		capturedFingerprint, _ = homestate.ProbeProcessIdentity(pid)
		return pid, nil
	}, term.stop, func(pid int) error {
		if actual, state := homestate.ProbeProcessIdentity(pid); state == homestate.ProcessIdentityLive && actual == capturedFingerprint {
			return fmt.Errorf("bootstrap MCP remains live after cleanup: pid=%d", pid)
		}
		return operationalNoBroker(root)
	}, 45*time.Second)
	if err != nil {
		t.Fatalf("trust bootstrap: %v terminal=%q", err, term.output())
	}
	t.Logf("TRUST_BOOTSTRAP_OK binary=%s root=%s mcp_pid=%d mcp_dead=true no_factory_state=true", bin, root, pid)
	return pid
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

func operationalExecutableIdentityEqual(want, actual string) bool {
	wantCanonical, err := filepath.EvalSymlinks(filepath.Clean(want))
	if err != nil {
		return false
	}
	actualCanonical, err := filepath.EvalSymlinks(filepath.Clean(actual))
	return err == nil && wantCanonical == actualCanonical
}

func operationalLoadedExecutable(output []byte, want string) bool {
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "n") && operationalExecutableIdentityEqual(want, strings.TrimPrefix(line, "n")) {
			return true
		}
	}
	return false
}

func operationalSelectTerminalOwnedMCP(output []byte, descendants map[int]string) int {
	type process struct {
		parent  int
		command string
		arg1    string
	}
	processes := make(map[int]process)
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		pid, _ := strconv.Atoi(fields[0])
		ppid, _ := strconv.Atoi(fields[1])
		entry := process{parent: ppid, command: filepath.Base(fields[2])}
		if len(fields) > 3 {
			entry.arg1 = fields[3]
		}
		processes[pid] = entry
	}
	for pid, entry := range processes {
		if _, ok := descendants[pid]; !ok {
			continue
		}
		if _, ok := descendants[entry.parent]; !ok {
			continue
		}
		owner, ok := processes[entry.parent]
		if !ok || owner.command != "codex" {
			continue
		}
		if entry.command == "moai" && entry.arg1 == "mcp-server" {
			return pid
		}
	}
	return 0
}

func operationalVerifyMCPBinary(t *testing.T, pid int, bin string) {
	t.Helper()
	if runtime.GOOS == "linux" {
		actual, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid))
		if err != nil || !operationalExecutableIdentityEqual(bin, actual) {
			t.Fatalf("MCP binary identity %q: %v", actual, err)
		}
		return
	}
	loaded, err := exec.Command("lsof", "-a", "-p", strconv.Itoa(pid), "-d", "txt", "-Fn").Output()
	if err != nil || !operationalLoadedExecutable(loaded, bin) {
		t.Fatalf("MCP loaded binary identity unavailable: %v %s", err, loaded)
	}
}

func operationalTerminalOwnedMCP(t *testing.T, terminalPID int, bin string) int {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		descendants, err := operationalDescendants(terminalPID)
		if err != nil {
			t.Fatalf("bootstrap process tree unavailable: %v", err)
		}
		out, err := exec.Command("ps", "-axo", "pid=,ppid=,command=").Output()
		if err != nil {
			t.Fatalf("bootstrap process commands unavailable: %v", err)
		}
		if pid := operationalSelectTerminalOwnedMCP(out, descendants); pid != 0 {
			operationalVerifyMCPBinary(t, pid, bin)
			t.Logf("BOOTSTRAP_OWNED_MCP terminal_pid=%d mcp_pid=%d", terminalPID, pid)
			return pid
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("no built-tree MCP owned by the bootstrap terminal tree")
	return 0
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
				operationalVerifyMCPBinary(t, pid, bin)
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

func waitOperationalAssistantOutput(t *testing.T, sessionID, exact string) {
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
			if err == nil && operationalAssistantOutput(data, exact) {
				return
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("session %s has no attributable assistant output %q", sessionID, exact)
}

func operationalAssistantOutput(data []byte, exact string) bool {
	scan := bufio.NewScanner(bytes.NewReader(data))
	scan.Buffer(make([]byte, 4096), 4<<20)
	for scan.Scan() {
		var event struct {
			Type    string `json:"type"`
			Payload struct {
				Type    string `json:"type"`
				Role    string `json:"role"`
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
			} `json:"payload"`
		}
		if json.Unmarshal(scan.Bytes(), &event) != nil || event.Type != "response_item" || event.Payload.Type != "message" || event.Payload.Role != "assistant" {
			continue
		}
		for _, content := range event.Payload.Content {
			if content.Type == "output_text" && strings.TrimSpace(content.Text) == exact {
				return true
			}
		}
	}
	return false
}

func operationalMCPResult(data []byte, run string) ([]factorymsg.LaneStatus, bool) {
	return operationalMCPResultCount(data, run, 1)
}

func operationalMCPResultCount(data []byte, run string, minimum int) ([]factorymsg.LaneStatus, bool) {
	calls := map[string]bool{}
	seen := map[string]bool{}
	type currentTurn struct {
		calls     map[string]bool
		completed map[string]bool
		outputs   map[string]struct {
			lanes []factorymsg.LaneStatus
			count int
		}
	}
	current := map[string]*currentTurn{}
	var turnOrder []string
	completedIDs := map[string]bool{}
	turn := func(id string) *currentTurn {
		if current[id] == nil {
			current[id] = &currentTurn{calls: map[string]bool{}, completed: map[string]bool{}, outputs: map[string]struct {
				lanes []factorymsg.LaneStatus
				count int
			}{}}
			turnOrder = append(turnOrder, id)
		}
		return current[id]
	}
	scan := bufio.NewScanner(strings.NewReader(string(data)))
	scan.Buffer(make([]byte, 4096), 4<<20)
	var latest []factorymsg.LaneStatus
	for scan.Scan() {
		var event struct {
			Type    string `json:"type"`
			Payload struct {
				Type, Name, Arguments string
				Output                json.RawMessage `json:"output"`
				CallID                string          `json:"call_id"`
				TurnID                string          `json:"turn_id"`
				Metadata              struct {
					TurnID string `json:"turn_id"`
				} `json:"internal_chat_message_metadata_passthrough"`
				Item struct {
					ID, Type, Server, Tool, Status string
					Arguments                      json.RawMessage `json:"arguments"`
				} `json:"item"`
			} `json:"payload"`
		}
		if json.Unmarshal(scan.Bytes(), &event) != nil {
			continue
		}
		p := event.Payload
		if event.Type == "response_item" && p.Type == "function_call" && p.CallID != "" && (p.Name == "mcp__moai__factory_msg_status" || p.Name == "functions.mcp__moai__factory_msg_status") {
			var args struct {
				Run string `json:"run_id"`
			}
			if json.Unmarshal([]byte(p.Arguments), &args) == nil && args.Run == run {
				calls[p.CallID] = true
			}
		}
		if event.Type == "response_item" && p.Type == "function_call_output" && p.CallID != "" && calls[p.CallID] {
			var output string
			if json.Unmarshal(p.Output, &output) == nil && output != "" {
				if lanes, ok := operationalOutputLanes([]byte(output)); ok {
					latest = lanes
					seen[p.CallID] = true
				}
			}
		}

		if event.Type == "response_item" && p.Type == "custom_tool_call" && p.Name == "exec" && p.CallID != "" && p.Metadata.TurnID != "" {
			turn(p.Metadata.TurnID).calls[p.CallID] = true
		}
		if event.Type == "event_msg" && p.Type == "item_completed" && p.TurnID != "" && p.Item.ID != "" && p.Item.Type == "McpToolCall" && p.Item.Server == "moai" && p.Item.Tool == "factory_msg_status" && p.Item.Status == "completed" {
			var args struct {
				Run string `json:"run_id"`
			}
			if json.Unmarshal(p.Item.Arguments, &args) == nil && args.Run == run {
				if !completedIDs[p.Item.ID] {
					completedIDs[p.Item.ID] = true
					turn(p.TurnID).completed[p.Item.ID] = true
				}
			}
		}
		if event.Type == "response_item" && p.Type == "custom_tool_call_output" && p.CallID != "" && p.Metadata.TurnID != "" {
			raw := bytes.TrimSpace(p.Output)
			if len(raw) > 0 && raw[0] == '[' {
				if lanes, count := operationalCurrentMCPOutput(raw); count > 0 {
					turn(p.Metadata.TurnID).outputs[p.CallID] = struct {
						lanes []factorymsg.LaneStatus
						count int
					}{lanes: lanes, count: count}
				}
			}
		}
	}
	currentCount := 0
	var currentLatest []factorymsg.LaneStatus
	for _, turnID := range turnOrder {
		evidence := current[turnID]
		outputCount := 0
		var turnLatest []factorymsg.LaneStatus
		for callID := range evidence.calls {
			output, ok := evidence.outputs[callID]
			if ok && output.count > 0 && output.lanes != nil {
				outputCount += output.count
				turnLatest = output.lanes
			}
		}
		if outputCount > len(evidence.completed) {
			outputCount = len(evidence.completed)
		}
		if outputCount > 0 && turnLatest != nil {
			currentCount += outputCount
			currentLatest = turnLatest
		}
	}
	if currentCount >= minimum && currentLatest != nil {
		return currentLatest, true
	}
	return latest, latest != nil && len(seen) >= minimum
}

func operationalCurrentMCPOutput(data []byte) ([]factorymsg.LaneStatus, int) {
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(data, &blocks) == nil && blocks != nil {
		for i, block := range blocks {
			if block.Type != "input_text" && block.Type != "text" {
				continue
			}
			marker := "Output:\n"
			at := strings.LastIndex(block.Text, marker)
			if at < 0 || (at > 0 && block.Text[at-1] != '\n') {
				continue
			}
			payload := strings.TrimSpace(block.Text[at+len(marker):])
			if payload == "" {
				payloadIndex := -1
				for j := i + 1; j < len(blocks); j++ {
					if strings.TrimSpace(blocks[j].Text) == "" {
						continue
					}
					if payloadIndex >= 0 || (blocks[j].Type != "input_text" && blocks[j].Type != "text") {
						return nil, 0
					}
					payloadIndex = j
					payload = strings.TrimSpace(blocks[j].Text)
				}
				if payloadIndex < 0 {
					return nil, 0
				}
			} else {
				for _, extra := range blocks[i+1:] {
					if strings.TrimSpace(extra.Text) != "" {
						return nil, 0
					}
				}
			}
			if lanes, count := operationalCurrentMCPOutput([]byte(payload)); count > 0 {
				return lanes, count
			}
			return nil, 0
		}
		return nil, 0
	}
	var wrapped struct {
		First  json.RawMessage `json:"first"`
		Second json.RawMessage `json:"second"`
	}
	if json.Unmarshal(data, &wrapped) == nil && (len(wrapped.First) > 0 || len(wrapped.Second) > 0) {
		var latest []factorymsg.LaneStatus
		count := 0
		for _, result := range []json.RawMessage{wrapped.First, wrapped.Second} {
			if len(result) == 0 {
				continue
			}
			if lanes, ok := operationalOutputLanes(result); ok {
				latest, count = lanes, count+1
			}
		}
		return latest, count
	}
	if lanes, ok := operationalOutputLanes(data); ok {
		return lanes, 1
	}
	return nil, 0
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
	mu        sync.Mutex
	text      strings.Builder
	stdin     io.Writer
	pid       int
	trust     operationalDirectoryTrust
	hookTrust operationalHookTrust
	stop      func() error
}

type operationalDirectoryTrust struct {
	root, raw string
	sent      bool
}

type operationalHookTrust struct{ operationalDirectoryTrust }

func (s *operationalHookTrust) takeSequence(approvedRoot string) string {
	if s.sent || s.root == "" || approvedRoot != s.root {
		return ""
	}
	text := strings.Join(strings.Fields(operationalANSI.ReplaceAllString(s.raw, " ")), " ")
	if strings.Contains(text, "› 2.") || strings.Contains(text, "› 3.") {
		return ""
	}
	for _, signature := range []string{"Hooks need review", "8 hooks are new or changed.", "Hooks can run outside the sandbox after you trust them.", "› 1. Review hooks", "2. Trust all and continue", "3. Continue without trusting (hooks won't run)", "Press enter to confirm or esc to go back"} {
		if !strings.Contains(text, signature) {
			return ""
		}
	}
	expected, err := codexwiring.RenderHooks(nil)
	if err != nil {
		return ""
	}
	actual, err := os.ReadFile(filepath.Join(s.root, ".codex/hooks.json"))
	if err != nil || !bytes.Equal(actual, expected) {
		return ""
	}
	s.sent = true
	s.raw = ""
	// Codex 0.155.1 displays option 1 selected. CSI B moves down to option 2;
	// CR confirms it. Never confirm the initially selected Review hooks entry.
	return "\x1b[B\r"
}

var operationalANSI = regexp.MustCompile("\x1b\\[[0-?]*[ -/]*[@-~]")

func (s *operationalDirectoryTrust) observe(chunk string) {
	if s.sent {
		return
	}
	s.raw += chunk
	// Discard previous screens, including a signature no longer displayed.
	if at := strings.LastIndex(s.raw, "\x1b[J"); at >= 0 {
		s.raw = s.raw[at+3:]
	}
	if len(s.raw) > 32768 {
		s.raw = s.raw[len(s.raw)-32768:]
	}
}
func (s *operationalDirectoryTrust) takeEnter() bool {
	if s.sent || s.root == "" {
		return false
	}
	text := strings.Join(strings.Fields(operationalANSI.ReplaceAllString(s.raw, " ")), " ")
	for _, signature := range []string{"You are in " + s.root + " ", "Do you trust the contents of this directory?", "› 1. Yes, continue", "2. No, quit", "Press enter to continue"} {
		if !strings.Contains(text, signature) {
			return false
		}
	}
	// This handler is confined to test-created directories, never user projects.
	base, err := filepath.EvalSymlinks(os.TempDir())
	if err != nil {
		return false
	}
	root, err := filepath.EvalSymlinks(s.root)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(base, root)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return false
	}
	s.sent = true
	s.raw = ""
	return true
}

func (b *operationalTerminal) acceptFixtureDirectoryTrust() error {
	b.mu.Lock()
	send := b.trust.takeEnter()
	b.mu.Unlock()
	if !send {
		return nil
	}
	_, err := b.stdin.Write([]byte("\r"))
	return err
}

func (b *operationalTerminal) acceptFixtureHookTrust() error {
	b.mu.Lock()
	approvedRoot := ""
	if b.trust.sent {
		approvedRoot = b.trust.root
	}
	sequence := b.hookTrust.takeSequence(approvedRoot)
	b.mu.Unlock()
	if sequence == "" {
		return nil
	}
	_, err := b.stdin.Write([]byte(sequence))
	return err
}

func (b *operationalTerminal) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.trust.observe(string(p))
	b.hookTrust.observe(string(p))
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

func operationalScriptArgs(bin string, args []string) []string {
	// script inherits a zero-sized terminal from go test's pipes. Set geometry
	// before exec; positional args preserve the exact production launch argv.
	argv := append([]string{"-q", "/dev/null", "/bin/sh", "-c", `/bin/stty rows 40 cols 160 && exec "$@"`, "moai-terminal", bin}, args...)
	if runtime.GOOS == "linux" {
		quoted := make([]string, 0, len(args)+1)
		for _, arg := range append([]string{bin}, args...) {
			quoted = append(quoted, "'"+strings.ReplaceAll(arg, "'", "'\\''")+"'")
		}
		argv = []string{"-q", "-c", "/bin/stty rows 40 cols 160 && exec " + strings.Join(quoted, " "), "/dev/null"}
	}
	return argv
}

func startOperationalTerminal(t *testing.T, root, bin string, args, env []string) *operationalTerminal {
	t.Helper()
	argv := operationalScriptArgs(bin, args)
	cmd := exec.Command("script", argv...)
	cmd.Dir, cmd.Env = root, env
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	b := &operationalTerminal{stdin: w, trust: operationalDirectoryTrust{root: root}, hookTrust: operationalHookTrust{operationalDirectoryTrust: operationalDirectoryTrust{root: root}}}
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
	var stopOnce sync.Once
	var stopErr error
	b.stop = func() error {
		stopOnce.Do(func() {
			record := func(format string, args ...any) { stopErr = errors.Join(stopErr, fmt.Errorf(format, args...)) }
			close(stopInventory)
			<-inventoryDone
			if inventoryErr != nil {
				record("cleanup continuous inventory unavailable: %v", inventoryErr)
			}
			owned, err := operationalDescendants(cmd.Process.Pid)
			if err != nil {
				record("cleanup process inventory unavailable: %v", err)
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
				record("cleanup timeout pid=%d", cmd.Process.Pid)
			}
			if err := waitOperationalOwnedExit(owned, homestate.ProbeProcessIdentity, 5*time.Second); err != nil {
				record("%v", err)
			}
			if err := syscall.Kill(-cmd.Process.Pid, 0); err != syscall.ESRCH {
				record("process group remains: %d %v", cmd.Process.Pid, err)
			}
			if stopErr == nil {
				t.Logf("CLEANUP_OK terminal_pid=%d descendants=%d", cmd.Process.Pid, len(owned))
			}
		})
		return stopErr
	}
	t.Cleanup(func() {
		if err := b.stop(); err != nil {
			t.Error(err)
		}
	})
	label := "PRODUCTION_ARGV"
	if reflect.DeepEqual(args, operationalBootstrapArgs()) {
		label = "TRUST_BOOTSTRAP_ARGV"
	}
	t.Logf("%s %s %s terminal_pid=%d", label, bin, strings.Join(args, " "), cmd.Process.Pid)
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

func waitOperationalOwnedExit(owned map[int]string, probe func(int) (string, homestate.ProcessIdentityState), limit time.Duration) error {
	deadline := time.Now().Add(limit)
	for {
		var remaining error
		for pid, fp := range owned {
			actual, state := probe(pid)
			if state == homestate.ProcessIdentityDead || (state == homestate.ProcessIdentityLive && actual != "" && actual != fp) {
				continue
			}
			remaining = fmt.Errorf("owned child remains: %d %s", pid, state)
			break
		}
		if remaining == nil {
			return nil
		}
		if !time.Now().Before(deadline) {
			return remaining
		}
		time.Sleep(10 * time.Millisecond)
	}
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
