package cli

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

// Shared harness for the LIVE acceptance tests of SPEC-DUAL-HARNESS-RECOVERY-001
// (AC-DHR-012, AC-DHR-018, AC-DHR-023): invocation/time budget, process
// tracking with cleanup registered before the first spawn, the evidence-file
// channel of acceptance.md §A, and Codex rollout reading.

const envT1100EvidenceDir = "MOAI_T1100_EVIDENCE_DIR"

// liveDenialPattern recognizes a sandbox refusal in a shell command's output.
var liveDenialPattern = regexp.MustCompile(`(?i)operation not permitted|read-only file system|permission denied|not permitted|sandbox`)

// liveBudget is a model-invocation and wall-clock budget. take is called
// BEFORE each model call; a refused take means the call is never started and
// the run is aborted. A run that ends exactly at the limit is not aborted.
type liveBudget struct {
	limit, used int
	window      time.Duration
	start       time.Time
	now         func() time.Time
	stopped     bool
}

func newLiveBudget(limit int, window time.Duration) *liveBudget {
	return &liveBudget{limit: limit, window: window, start: time.Now(), now: time.Now}
}

func (b *liveBudget) take() (bool, string) {
	if b.used+1 > b.limit {
		b.stopped = true
		return false, fmt.Sprintf("invocation budget: call %d would exceed the limit of %d", b.used+1, b.limit)
	}
	if elapsed := b.now().Sub(b.start); elapsed >= b.window {
		b.stopped = true
		return false, fmt.Sprintf("time budget: elapsed %.0fs reached the window of %.0fs", elapsed.Seconds(), b.window.Seconds())
	}
	b.used++
	return true, ""
}

func (b *liveBudget) aborted() bool { return b.stopped }

func (b *liveBudget) remaining() time.Duration {
	if r := b.window - b.now().Sub(b.start); r > 0 {
		return r
	}
	return 0
}

func (b *liveBudget) elapsedSeconds() float64 { return b.now().Sub(b.start).Seconds() }

// liveEvidenceSkipReason returns the explicit NOT_RUN skip reason when the
// evidence channel is absent, or "" when it is set.
func liveEvidenceSkipReason(dir string) string {
	if strings.TrimSpace(dir) != "" {
		return ""
	}
	return "NOT_RUN " + envT1100EvidenceDir + " is empty: LIVE evidence has no channel, so this test does not run"
}

// liveEvidenceDir resolves the evidence directory or skips with NOT_RUN.
func liveEvidenceDir(t *testing.T) string {
	t.Helper()
	raw := os.Getenv(envT1100EvidenceDir)
	if reason := liveEvidenceSkipReason(raw); reason != "" {
		t.Skip(reason)
	}
	dir, err := filepath.Abs(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// writeLiveEvidenceFile writes v as JSON and returns the tag line
// "<TAG>_EVIDENCE_SHA256 [<case> ]<hex>" carrying the sha256 of the bytes.
func writeLiveEvidenceFile(dir, name, tag, caseName string, v any) (string, error) {
	body, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	body = append(body, '\n')
	if err := os.WriteFile(filepath.Join(dir, name), body, 0o644); err != nil {
		return "", err
	}
	sum := sha256.Sum256(body)
	line := tag + "_EVIDENCE_SHA256 "
	if caseName != "" {
		line += caseName + " "
	}
	return line + hex.EncodeToString(sum[:]), nil
}

// emitLiveEvidence writes the file and prints the short tag line on stdout.
func emitLiveEvidence(t *testing.T, dir, name, tag, caseName string, v any) {
	t.Helper()
	line, err := writeLiveEvidenceFile(dir, name, tag, caseName, v)
	if err != nil {
		t.Fatalf("write evidence %s: %v", name, err)
	}
	fmt.Println(line)
}

// liveProcs tracks every process a LIVE test spawns. newLiveProcs registers
// reap with t.Cleanup, so constructing it before the first spawn guarantees
// cleanup on every exit path. Each process runs in its own process group and
// reap kills the whole group, reaching grandchildren (MCP servers, shells).
type liveProcs struct {
	mu      sync.Mutex
	cmds    []*exec.Cmd
	waited  map[*exec.Cmd]bool
	cleaned []int
	done    map[int]bool
}

func newLiveProcs(t *testing.T) *liveProcs {
	p := &liveProcs{waited: map[*exec.Cmd]bool{}, done: map[int]bool{}}
	t.Cleanup(func() { p.reap() })
	return p
}

// liveCommand builds a context-bounded command whose cancellation kills its
// whole process group rather than only the direct child.
func liveCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	setLiveProcGroup(cmd)
	cmd.Cancel = func() error { return killLiveProcGroup(cmd.Process.Pid) }
	cmd.WaitDelay = 10 * time.Second
	return cmd
}

func (p *liveProcs) start(cmd *exec.Cmd) error {
	if cmd.SysProcAttr == nil {
		setLiveProcGroup(cmd)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	p.mu.Lock()
	p.cmds = append(p.cmds, cmd)
	p.mu.Unlock()
	return nil
}

// run starts cmd tracked, waits for it, and kills any group leftovers.
func (p *liveProcs) run(cmd *exec.Cmd) ([]byte, error) {
	var buf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &buf, &buf
	if err := p.start(cmd); err != nil {
		return nil, err
	}
	err := cmd.Wait()
	p.mu.Lock()
	p.waited[cmd] = true
	p.mu.Unlock()
	_ = killLiveProcGroup(cmd.Process.Pid)
	return buf.Bytes(), err
}

// reap kills every tracked process group, waits for unwaited processes, and
// returns the pids whose group is confirmed gone. It is idempotent.
func (p *liveProcs) reap() []int {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, cmd := range p.cmds {
		pid := cmd.Process.Pid
		if p.done[pid] {
			continue
		}
		_ = killLiveProcGroup(pid)
		if !p.waited[cmd] {
			_ = cmd.Wait()
			p.waited[cmd] = true
		}
		if liveProcGroupGone(pid) {
			p.done[pid] = true
			p.cleaned = append(p.cleaned, pid)
		}
	}
	return append([]int(nil), p.cleaned...)
}

// codexRollout is the part of one Codex session record the LIVE tests read.
type codexRollout struct {
	Path                 string
	Subagent             bool
	Role, Sandbox, Final string
	tools                []liveToolIO
}

type liveToolIO struct{ Input, Output string }

type rolloutLine struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// parseCodexRollout reads a rollout JSONL tolerantly: unknown line shapes
// are ignored, never fatal.
func parseCodexRollout(data []byte) codexRollout {
	var r codexRollout
	inputs := map[string]string{}
	outputs := map[string]string{}
	var order []string
	lastAssistant := ""
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 1<<20), 64<<20)
	for sc.Scan() {
		var ln rolloutLine
		if json.Unmarshal(sc.Bytes(), &ln) != nil {
			continue
		}
		var p map[string]any
		if json.Unmarshal(ln.Payload, &p) != nil {
			continue
		}
		switch ln.Type {
		case "session_meta":
			if src, ok := p["source"].(map[string]any); ok {
				if sub, ok := src["subagent"]; ok {
					r.Subagent = true
					if m, ok := sub.(map[string]any); ok {
						if ts, ok := m["thread_spawn"].(map[string]any); ok {
							r.Role, _ = ts["agent_role"].(string)
						}
					}
				}
			}
			if role, ok := p["agent_role"].(string); ok && role != "" && r.Subagent {
				r.Role = role
			}
		case "turn_context":
			switch sp := p["sandbox_policy"].(type) {
			case map[string]any:
				r.Sandbox, _ = sp["type"].(string)
			case string:
				r.Sandbox = sp
			}
		case "response_item":
			id, _ := p["call_id"].(string)
			switch p["type"] {
			case "function_call":
				inputs[id] = liveText(p["arguments"])
				order = append(order, id)
			case "custom_tool_call":
				inputs[id] = liveText(p["input"])
				order = append(order, id)
			case "local_shell_call":
				inputs[id] = liveText(p["action"])
				order = append(order, id)
			case "function_call_output", "custom_tool_call_output":
				outputs[id] += liveText(p["output"])
			case "message":
				if p["role"] == "assistant" {
					if t := liveText(p["content"]); t != "" {
						lastAssistant = t
					}
				}
			}
		case "event_msg":
			switch p["type"] {
			case "task_complete":
				if m, ok := p["last_agent_message"].(string); ok && m != "" {
					r.Final = m
				}
			case "exec_command_end":
				r.tools = append(r.tools, liveToolIO{Input: liveText(p["command"]), Output: liveText(p["aggregated_output"]) + liveText(p["stderr"])})
			}
		}
	}
	for _, id := range order {
		r.tools = append(r.tools, liveToolIO{Input: inputs[id], Output: outputs[id]})
	}
	if r.Final == "" {
		r.Final = lastAssistant
	}
	return r
}

// liveText flattens a payload value: strings verbatim, content arrays by
// their text fields, anything else as JSON.
func liveText(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case []any:
		var parts []string
		for _, e := range x {
			if m, ok := e.(map[string]any); ok {
				if s, ok := m["text"].(string); ok {
					parts = append(parts, s)
					continue
				}
			}
			parts = append(parts, liveText(e))
		}
		return strings.Join(parts, "")
	default:
		b, _ := json.Marshal(x)
		return string(b)
	}
}

// outputsMentioning joins the outputs of tool calls whose input names needle.
func (r codexRollout) outputsMentioning(needle string) string {
	var parts []string
	for _, io := range r.tools {
		if strings.Contains(io.Input, needle) && io.Output != "" {
			parts = append(parts, io.Output)
		}
	}
	return strings.Join(parts, "\n")
}

// listCodexRollouts returns the rollout files under a CODEX_HOME.
func listCodexRollouts(codexHome string) map[string]bool {
	out := map[string]bool{}
	_ = filepath.Walk(filepath.Join(codexHome, "sessions"), func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasPrefix(info.Name(), "rollout-") && strings.HasSuffix(info.Name(), ".jsonl") {
			out[path] = true
		}
		return nil
	})
	return out
}

// newCodexRollouts parses the rollouts present in after but not in before.
func newCodexRollouts(before, after map[string]bool) []codexRollout {
	var paths []string
	for p := range after {
		if !before[p] {
			paths = append(paths, p)
		}
	}
	sort.Strings(paths)
	var out []codexRollout
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		r := parseCodexRollout(data)
		r.Path = p
		out = append(out, r)
	}
	return out
}

// isolatedCodexHome builds a throwaway CODEX_HOME holding a copy of the
// operator's existing Codex login (no other credential is introduced) and a
// config that trusts the test repository. It skips with NOT_RUN when no login
// exists, or when the login is due for a token refresh: a refresh from the
// copy would rotate the operator's token out from under ~/.codex.
func isolatedCodexHome(t *testing.T, trusted ...string) (home string, authHash string) {
	t.Helper()
	real, err := factoryLiveOperatorHomeFn()
	if err != nil {
		t.Skip("NOT_RUN operator home unavailable: " + err.Error())
	}
	src := filepath.Join(real, ".codex", "auth.json")
	auth, err := os.ReadFile(src)
	if err != nil {
		t.Skip("NOT_RUN no Codex login at ~/.codex/auth.json")
	}
	var meta struct {
		LastRefresh time.Time `json:"last_refresh"`
	}
	if json.Unmarshal(auth, &meta) == nil && !meta.LastRefresh.IsZero() && time.Since(meta.LastRefresh) > 7*24*time.Hour {
		t.Skip("NOT_RUN Codex login is due for a token refresh; refreshing a copy would rotate the operator token")
	}
	home = t.TempDir()
	if err := os.WriteFile(filepath.Join(home, "auth.json"), auth, 0o600); err != nil {
		t.Fatal(err)
	}
	var cfg strings.Builder
	for _, dir := range trusted {
		fmt.Fprintf(&cfg, "[projects.%q]\ntrust_level = \"trusted\"\n", dir)
	}
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte(cfg.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	return home, sha256Hex(auth)
}

// codexAuthChanged reports whether Codex rewrote the copied login.
func codexAuthChanged(home, before string) bool {
	b, err := os.ReadFile(filepath.Join(home, "auth.json"))
	return err != nil || sha256Hex(b) != before
}

// buildLiveMoai builds the moai binary from this tree through the tracker.
func buildLiveMoai(t *testing.T, procs *liveProcs) string {
	t.Helper()
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(t.TempDir(), "moai")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := liveCommand(ctx, "go", "build", "-o", bin, "./cmd/moai")
	cmd.Dir = repoRoot
	if out, err := procs.run(cmd); err != nil {
		t.Fatalf("build live moai: %v: %s", err, out)
	}
	return bin
}

// canonicalDir resolves symlinks (macOS /var -> /private/var).
func canonicalDir(t *testing.T, dir string) string {
	t.Helper()
	c, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// tailString bounds logged output.
func tailString(b []byte, n int) string {
	if len(b) > n {
		b = b[len(b)-n:]
	}
	return strings.TrimSpace(string(b))
}

var errLiveAborted = errors.New("live budget exhausted")
