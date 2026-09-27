//go:build darwin || linux

package cli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// The two operational live proofs (TestFactoryLiveOperationalRosterBeforePrompt,
// TestFactoryLiveOperationalLauncherChain) launched `moai codex -f` lanes and
// were removed with that entry (SPEC-CODEX-FACTORY-RETIRE-001). The helpers
// below remain for the production-init fixture tests that still use them.

// operationalHookShellEnv keeps the production bare `moai hook ...` wiring
// while making Codex's login-shell lookup resolve the same binary that launched
// the fixture. A login shell may rebuild PATH from user startup files and pick
// an installed, stale moai even when the fixture prepended its build directory.
//
// The boundary under test is "a login shell's startup files decide which moai
// runs", not zsh itself, so the fixture takes whichever login shell the host
// has and seeds that shell's own profile in a private directory: zsh reads
// $ZDOTDIR/.zprofile; bash, the fallback on hosts without zsh (Linux CI
// runners), reads $HOME/.bash_profile, so HOME is pointed at the private
// directory for the shell processes this env reaches. A host with neither
// shell still fails loudly — the boundary cannot be measured there.
func operationalHookShellEnv(t *testing.T, bin string, env []string) []string {
	t.Helper()
	profile := "export PATH=" + shellQuote(filepath.Dir(bin)) + ":\"$PATH\"\n"
	profileDir := t.TempDir()
	var shell, profileFile, dirKey string
	if zsh, err := exec.LookPath("zsh"); err == nil {
		shell, profileFile, dirKey = zsh, ".zprofile", "ZDOTDIR"
	} else if bash, bashErr := exec.LookPath("bash"); bashErr == nil {
		shell, profileFile, dirKey = bash, ".bash_profile", "HOME"
	} else {
		t.Fatalf("Codex hook-shell fixture requires a login shell (zsh or bash): zsh: %v; bash: %v", err, bashErr)
	}
	if err := os.WriteFile(filepath.Join(profileDir, profileFile), []byte(profile), 0600); err != nil {
		t.Fatal(err)
	}
	filtered := make([]string, 0, len(env)+2)
	for _, item := range env {
		key, _, _ := strings.Cut(item, "=")
		if key != "SHELL" && key != "ZDOTDIR" && key != dirKey {
			filtered = append(filtered, item)
		}
	}
	filtered = append(filtered, "SHELL="+shell, dirKey+"="+profileDir)
	t.Logf("HOOK_LOGIN_SHELL %s profile=%s", shell, profileFile)
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
