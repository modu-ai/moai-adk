package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

type factoryLiveCase struct{ name, lead, worker string }

func TestFactoryLiveCodexCodex(t *testing.T) {
	runFactoryLiveCase(t, factoryLiveCase{"codex-codex", "codex", "codex"})
}
func TestFactoryLiveCodexClaude(t *testing.T) {
	runFactoryLiveCase(t, factoryLiveCase{"codex-claude", "codex", "claude"})
}
func TestFactoryLiveClaudeCodex(t *testing.T) {
	runFactoryLiveCase(t, factoryLiveCase{"claude-codex", "claude", "codex"})
}
func TestFactoryLiveClaudeClaudeCompletionSeparation(t *testing.T) {
	runFactoryLiveCase(t, factoryLiveCase{"claude-claude", "claude", "claude"})
}

func TestFactoryLiveHookBoundaryIdleTruth(t *testing.T) {
	requireFactoryLive(t, "idle-boundary")
	fx := newFactoryLiveFixture(t, "claude", "claude")
	msg, err := fx.store.Send(context.Background(), factorymsg.SendRequest{From: fx.lead, To: fx.worker, Kind: factorymsg.KindStatusRequest, IdempotencyKey: "idle-" + fx.nonce, TaskRef: "t1074", CorrelationID: "idle-" + fx.nonce, TTL: time.Minute, Payload: []byte("nonce=" + fx.nonce)})
	if err != nil {
		t.Fatal(err)
	}
	st, err := fx.store.Status(context.Background())
	if err != nil || st.Pending != 1 || st.NextDelivery != "pending-until-next-turn" {
		t.Fatalf("idle status=%+v err=%v", st, err)
	}
	fx.runModel(t, fx.worker, "Read the factory inbox for run "+fx.runID+", read message "+msg.ID+", persist accepted disposition, and receipt it. Do not invent ids.")
	st, err = fx.store.Status(context.Background())
	if err != nil || st.Pending != 0 || st.Acknowledged != 1 {
		t.Fatalf("boundary delivery status=%+v err=%v", st, err)
	}
	t.Logf("idle pending truth and next-turn receipt verified: message=%s", msg.ID)
}

func runFactoryLiveCase(t *testing.T, tc factoryLiveCase) {
	requireFactoryLive(t, tc.name)
	fx := newFactoryLiveFixture(t, tc.lead, tc.worker)
	fx.runModel(t, fx.lead, fmt.Sprintf("Use factory_msg_send exactly once: run_id=%s to_slot=agent-1 kind=status_request idempotency_key=lead-%s task_ref=t1074 correlation_id=req-%s ttl_seconds=120 body=nonce:%s. Report only the returned message id.", fx.runID, fx.nonce, fx.nonce, fx.nonce))
	st, err := fx.store.Status(context.Background())
	if err != nil || st.Pending != 1 {
		t.Fatalf("lead send status=%+v err=%v", st, err)
	}
	fx.runModel(t, fx.worker, fmt.Sprintf("For run_id=%s call factory_msg_list, read every body with factory_msg_body, then factory_msg_receipt disposition=accepted. Send exactly one status_report reply to slot lead with idempotency_key=worker-%s task_ref=t1074 correlation_id=reply-%s ttl_seconds=120 and body=reply:%s.", fx.runID, fx.nonce, fx.nonce, fx.nonce))
	fx.runModel(t, fx.lead, "For run_id="+fx.runID+" list the inbox, read every body, and receipt every claim with disposition=accepted.")
	st, err = fx.store.Status(context.Background())
	if err != nil || st.Pending != 0 || st.Claimed != 0 || st.Acknowledged < 2 {
		t.Fatalf("round-trip status=%+v err=%v", st, err)
	}
	t.Logf("live separate-context nonce/receipt verified: case=%s nonce=%s acknowledged=%d", tc.name, fx.nonce, st.Acknowledged)
}

func requireFactoryLive(t *testing.T, want string) {
	t.Helper()
	if os.Getenv("MOAI_FACTORY_LIVE") != "1" {
		t.Skip("set MOAI_FACTORY_LIVE=1 and select one live case; acceptance gates reject skips")
	}
	if got := os.Getenv("MOAI_FACTORY_LIVE_CASE"); got != want {
		t.Fatalf("MOAI_FACTORY_LIVE_CASE=%q, want %q", got, want)
	}
}

type factoryLiveFixture struct {
	root, runID, nonce, moai string
	// codexHome and workers are set only by the card-flow fixture.
	codexHome, workers string
	store              *factorymsg.Store
	lead, worker       factorymsg.Peer
}

var factoryLiveOperatorHomeFn = os.UserHomeDir

func newFactoryLiveFixture(t *testing.T, leadBackend, workerBackend string) *factoryLiveFixture {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o700); err != nil {
		t.Fatal(err)
	}
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	moai := filepath.Join(t.TempDir(), "moai")
	cmd := exec.Command("go", "build", "-o", moai, "./cmd/moai")
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "GOCACHE="+filepath.Join(t.TempDir(), "gocache"))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build live moai: %v: %s", err, out)
	}
	if err := os.WriteFile(filepath.Join(root, ".mcp.json"), []byte(fmt.Sprintf(`{"mcpServers":{"moai":{"command":%q,"args":["mcp-server"]}}}`, moai)), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".codex"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".codex", "config.toml"), []byte(fmt.Sprintf("[mcp_servers.moai]\ncommand = %q\nargs = [\"mcp-server\"]\nenv_vars = [\"MOAI_HOME\", \"MOAI_KANBAN_ID\", \"MOAI_SESSION_PID\", \"MOAI_KANBAN_BACKEND\", \"MOAI_FACTORY_WORKER\", \"MOAI_FACTORY_WORKERS\", \"CLAUDE_PROJECT_DIR\", \"CLAUDE_CODE_SESSION_ID\"]\ndefault_tools_approval_mode = \"writes\"\n", moai)), 0o600); err != nil {
		t.Fatal(err)
	}
	runID := "live-" + factoryLiveID()[:12]
	registry, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.RecordRun(context.Background(), homestate.FactoryRun{RunID: runID, LeadSessionID: "live-lead", Backend: leadBackend, ManifestJSON: "{}"}); err != nil {
		_ = registry.Close()
		t.Fatal(err)
	}
	_ = registry.Close()
	s, err := factorymsg.Open(root, runID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	owner := func(backend string) (int, string) {
		pid := os.Getpid()
		if backend == "codex" {
			ctx, cancel := context.WithCancel(context.Background())
			cmd := exec.CommandContext(ctx, "sleep", "300")
			if err := cmd.Start(); err != nil {
				cancel()
				t.Fatalf("start Codex owner fixture: %v", err)
			}
			t.Cleanup(func() {
				cancel()
				_ = cmd.Wait()
			})
			pid = cmd.Process.Pid
		}
		fp, state := homestate.ProbeProcessIdentity(pid)
		if state != homestate.ProcessIdentityLive || fp == "" {
			t.Fatalf("%s owner fingerprint unavailable: pid=%d state=%s", backend, pid, state)
		}
		return pid, fp
	}
	leadPID, leadFP := owner(leadBackend)
	workerPID, workerFP := owner(workerBackend)
	base := factorymsg.Peer{ProjectKey: homestate.ProjectKey(root), RunID: runID, Generation: 1}
	lead := base
	lead.PID, lead.ProcessStart = leadPID, leadFP
	lead.Backend, lead.Role, lead.Slot, lead.SessionUUID = leadBackend, "lead", "lead", "live-lead-"+factoryLiveID()
	worker := base
	worker.PID, worker.ProcessStart = workerPID, workerFP
	worker.Backend, worker.Role, worker.Slot, worker.SessionUUID = workerBackend, "worker", "agent-1", "live-worker-"+factoryLiveID()
	if lead, err = s.RegisterPeer(context.Background(), lead); err != nil {
		t.Fatal(err)
	}
	if worker, err = s.RegisterPeer(context.Background(), worker); err != nil {
		t.Fatal(err)
	}
	return &factoryLiveFixture{root: root, runID: runID, nonce: factoryLiveID(), moai: moai, store: s, lead: lead, worker: worker}
}

func factoryLiveID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b[:])
}

func (f *factoryLiveFixture) modelCommand(ctx context.Context, p factorymsg.Peer, prompt string) (*exec.Cmd, error) {
	switch p.Backend {
	case "codex":
		bin, err := exec.LookPath("codex")
		if err != nil {
			return nil, err
		}
		return exec.CommandContext(ctx, bin,
			"exec",
			"--ignore-user-config",
			"--skip-git-repo-check",
			"--approve-for-me",
			"-c", fmt.Sprintf("mcp_servers.moai.command=%q", f.moai),
			"-c", `mcp_servers.moai.args=["mcp-server"]`,
			"-c", `mcp_servers.moai.env_vars=["MOAI_HOME","MOAI_KANBAN_ID","MOAI_SESSION_PID","MOAI_KANBAN_BACKEND","MOAI_FACTORY_WORKER","MOAI_FACTORY_WORKERS","CLAUDE_PROJECT_DIR","CLAUDE_CODE_SESSION_ID"]`,
			"--json",
			prompt,
		), nil
	case "claude":
		return exec.CommandContext(ctx, f.moai,
			"glm", "--",
			"-p", "--output-format", "json",
			"--mcp-config", filepath.Join(f.root, ".mcp.json"),
			"--strict-mcp-config", "--permission-mode", "bypassPermissions",
			prompt,
		), nil
	default:
		return nil, fmt.Errorf("unknown backend %q", p.Backend)
	}
}

func (f *factoryLiveFixture) writePeerMCPConfig(p factorymsg.Peer) error {
	roleEnv := map[string]string{
		config.EnvHome:                os.Getenv(config.EnvHome),
		config.EnvMoaiKanbanID:        f.runID,
		config.EnvMoaiSessionPID:      fmt.Sprint(p.PID),
		config.EnvMoaiKanbanBackend:   p.Backend,
		config.EnvClaudeProjectDir:    f.root,
		config.EnvClaudeCodeSessionID: p.SessionUUID,
		config.EnvMoaiFactoryWorker:   "",
		config.EnvMoaiFactoryWorkers:  "1",
	}
	if p.Role == "worker" {
		roleEnv[config.EnvMoaiFactoryWorker] = p.Slot
		roleEnv[config.EnvMoaiFactoryWorkers] = ""
	}
	doc := map[string]any{"mcpServers": map[string]any{"moai": map[string]any{
		"command": f.moai,
		"args":    []string{"mcp-server"},
		"env":     roleEnv,
	}}}
	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(f.root, ".mcp.json"), body, 0o600)
}

func factoryLiveWithoutAttribution(env []string) []string {
	drop := map[string]bool{
		config.EnvMoaiKanbanID: true, config.EnvMoaiSessionPID: true,
		config.EnvMoaiKanbanBackend: true, config.EnvMoaiFactoryWorker: true,
		config.EnvMoaiFactoryWorkers: true, config.EnvClaudeCodeSessionID: true,
	}
	out := make([]string, 0, len(env))
	for _, item := range env {
		key, _, _ := strings.Cut(item, "=")
		if !drop[key] {
			out = append(out, item)
		}
	}
	return out
}

// factoryLiveOperatorGLMKey bridges the operator credential into the isolated
// live child without copying it to the fixture MOAI_HOME. The raw OS home
// resolver bypasses TestMain's package-level home seam; only this bounded read
// temporarily points the existing credential loader at the operator home.
func factoryLiveOperatorGLMKey() string {
	if key := os.Getenv(config.EnvTestGLMKey); key != "" {
		return key
	}
	realHome, err := factoryLiveOperatorHomeFn()
	if err != nil || strings.TrimSpace(realHome) == "" {
		return ""
	}
	previous, present := os.LookupEnv(config.EnvHome)
	if err := os.Setenv(config.EnvHome, filepath.Join(realHome, ".moai")); err != nil {
		return ""
	}
	defer func() {
		if present {
			_ = os.Setenv(config.EnvHome, previous)
		} else {
			_ = os.Unsetenv(config.EnvHome)
		}
	}()
	return loadGLMKey()
}

func (f *factoryLiveFixture) runModel(t *testing.T, p factorymsg.Peer, prompt string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
	defer cancel()
	if p.Backend == "claude" {
		if err := f.writePeerMCPConfig(p); err != nil {
			t.Fatal(err)
		}
	}
	cmd, err := f.modelCommand(ctx, p, prompt)
	if err != nil {
		t.Fatal(err)
	}
	cmd.Dir = f.root
	env, err := f.modelEnv(p)
	if err != nil {
		t.Fatal(err)
	}
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s live turn failed: %v: %s", p.Backend, err, strings.TrimSpace(string(out)))
	}
	const liveOutputTail = 8192
	if len(out) > liveOutputTail {
		out = out[len(out)-liveOutputTail:]
	}
	t.Logf("%s live turn output tail: %s", p.Backend, strings.TrimSpace(string(out)))
}

// modelEnv is the child environment of one live model turn: Claude turns get
// the operator GLM credential and no inherited attribution; Codex turns get
// the peer's attribution variables for the moai MCP server.
func (f *factoryLiveFixture) modelEnv(p factorymsg.Peer) ([]string, error) {
	env := os.Environ()
	if p.Backend == "claude" {
		env = factoryLiveWithoutAttribution(env)
		key := factoryLiveOperatorGLMKey()
		if key == "" {
			return nil, fmt.Errorf("operator GLM credential unavailable")
		}
		env = replaceEnvValue(env, config.EnvTestGLMKey, key)
		env = replaceEnvValue(env, config.EnvClaudeProjectDir, f.root)
	} else {
		env = replaceEnvValue(env, config.EnvMoaiKanbanID, f.runID)
		env = replaceEnvValue(env, config.EnvMoaiSessionPID, fmt.Sprint(p.PID))
		env = replaceEnvValue(env, config.EnvMoaiKanbanBackend, p.Backend)
		env = replaceEnvValue(env, config.EnvClaudeProjectDir, f.root)
		env = replaceEnvValue(env, config.EnvClaudeCodeSessionID, "")
		if p.Role == "worker" {
			env = replaceEnvValue(env, config.EnvMoaiFactoryWorker, p.Slot)
			env = replaceEnvValue(env, config.EnvMoaiFactoryWorkers, "")
		} else {
			env = replaceEnvValue(env, config.EnvMoaiFactoryWorker, "")
			env = replaceEnvValue(env, config.EnvMoaiFactoryWorkers, "1")
		}
		if f.codexHome != "" {
			env = replaceEnvValue(env, "CODEX_HOME", f.codexHome)
		}
		if f.workers != "" && p.Role != "worker" {
			env = replaceEnvValue(env, config.EnvMoaiFactoryWorkers, f.workers)
		}
	}
	return env, nil
}

// Card-flow LIVE budget per combination (REQ-DHR-023, lead decision 4).
const (
	factoryCardLiveInvocations = 8
	factoryCardLiveWindow      = 900 * time.Second
	factoryCardLiveCallBound   = 180 * time.Second
)

var factoryCardAttemptPattern = regexp.MustCompile(`attempt=(\d+)`)

// TestFactoryLiveCardFlow* are AC-DHR-018 (REQ-DHR-023). Each drives the
// REQ-DHR-022 card flow with real, separate model contexts: the lead model
// sends both assignments, worker models receive them and send the results,
// and the interrupted first worker sends its result late. The harness stands
// in for the lane runtime that owns the dispatch record (there is no dispatch
// MCP surface): it performs the explicit assignee transitions after the
// observed receipt, reassigns with an explicit revoke, applies each result
// before receipting it (REQ-DHR-019), and integrates.
func TestFactoryLiveCardFlowClaudeClaude(t *testing.T) {
	runFactoryCardFlow(t, factoryLiveCase{"claude-claude", "claude", "claude"})
}
func TestFactoryLiveCardFlowCodexCodex(t *testing.T) {
	runFactoryCardFlow(t, factoryLiveCase{"codex-codex", "codex", "codex"})
}
func TestFactoryLiveCardFlowClaudeCodex(t *testing.T) {
	runFactoryCardFlow(t, factoryLiveCase{"claude-codex", "claude", "codex"})
}
func TestFactoryLiveCardFlowCodexClaude(t *testing.T) {
	runFactoryCardFlow(t, factoryLiveCase{"codex-claude", "codex", "claude"})
}

// requireFactoryCardLive is the card-flow gate: an unset gate or an absent
// evidence channel is an explicit NOT_RUN skip; a mismatched case is fatal,
// like requireFactoryLive, so a combination only runs when selected.
func requireFactoryCardLive(t *testing.T, name string) string {
	t.Helper()
	if os.Getenv("MOAI_FACTORY_LIVE") != "1" {
		t.Skip("NOT_RUN MOAI_FACTORY_LIVE is not 1: the live card flow is gated off")
	}
	if got, want := os.Getenv("MOAI_FACTORY_LIVE_CASE"), "cardflow-"+name; got != want {
		t.Fatalf("MOAI_FACTORY_LIVE_CASE=%q, want %q", got, want)
	}
	return liveEvidenceDir(t)
}

type factoryCardFixture struct {
	*factoryLiveFixture
	procs         *liveProcs
	first, second factorymsg.Peer
	dispatchID    string
}

// newFactoryCardFixture builds an isolated repository, MOAI_HOME and
// CODEX_HOME with a lead and two worker lanes. Every process it spawns goes
// through procs, whose cleanup the caller registered before this call.
func newFactoryCardFixture(t *testing.T, procs *liveProcs, leadBackend, workerBackend string) *factoryCardFixture {
	t.Helper()
	t.Setenv(config.EnvHome, t.TempDir())
	root := canonicalDir(t, t.TempDir())
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o700); err != nil {
		t.Fatal(err)
	}
	codexHome := ""
	if leadBackend == "codex" || workerBackend == "codex" {
		codexHome, _ = isolatedCodexHome(t, root)
	}
	moai := buildLiveMoai(t, procs)
	runID := "live-" + factoryLiveID()[:12]
	registry, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.RecordRun(context.Background(), homestate.FactoryRun{RunID: runID, LeadSessionID: "live-lead", Backend: leadBackend, ManifestJSON: "{}"}); err != nil {
		_ = registry.Close()
		t.Fatal(err)
	}
	_ = registry.Close()
	s, err := factorymsg.Open(root, runID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	owner := func(backend string) (int, string) {
		pid := os.Getpid()
		if backend == "codex" {
			cmd := liveCommand(context.Background(), "sleep", "1200")
			if err := procs.start(cmd); err != nil {
				t.Fatalf("start Codex owner fixture: %v", err)
			}
			pid = cmd.Process.Pid
		}
		fp, state := homestate.ProbeProcessIdentity(pid)
		if state != homestate.ProcessIdentityLive || fp == "" {
			t.Fatalf("%s owner fingerprint unavailable: pid=%d state=%s", backend, pid, state)
		}
		return pid, fp
	}
	register := func(backend, role, slot string) factorymsg.Peer {
		pid, fp := owner(backend)
		p := factorymsg.Peer{ProjectKey: homestate.ProjectKey(root), RunID: runID, Generation: 1, PID: pid, ProcessStart: fp,
			Backend: backend, Role: role, Slot: slot, SessionUUID: "live-" + slot + "-" + factoryLiveID()}
		got, err := s.RegisterPeer(context.Background(), p)
		if err != nil {
			t.Fatal(err)
		}
		return got
	}
	lead := register(leadBackend, "lead", "lead")
	first := register(workerBackend, "worker", "agent-1")
	second := register(workerBackend, "worker", "agent-2")
	nonce := factoryLiveID()
	base := &factoryLiveFixture{root: root, runID: runID, nonce: nonce, moai: moai, store: s, lead: lead, worker: first, codexHome: codexHome, workers: "2"}
	return &factoryCardFixture{factoryLiveFixture: base, procs: procs, first: first, second: second, dispatchID: "d-" + nonce[:12]}
}

// turn runs one model turn tracked and bounded by the remaining budget.
func (f *factoryCardFixture) turn(t *testing.T, budget *liveBudget, p factorymsg.Peer, prompt string) error {
	t.Helper()
	bound := factoryCardLiveCallBound
	if rem := budget.remaining(); rem < bound {
		bound = rem
	}
	ctx, cancel := context.WithTimeout(context.Background(), bound)
	defer cancel()
	if p.Backend == "claude" {
		if err := f.writePeerMCPConfig(p); err != nil {
			return err
		}
	}
	proto, err := f.modelCommand(ctx, p, prompt)
	if err != nil {
		return err
	}
	cmd := liveCommand(ctx, proto.Path, proto.Args[1:]...)
	cmd.Dir = f.root
	if cmd.Env, err = f.modelEnv(p); err != nil {
		return err
	}
	out, err := f.procs.run(cmd)
	t.Logf("%s %s turn output tail: %s", p.Backend, p.Slot, tailString(out, 2000))
	if ctx.Err() != nil && budget.remaining() == 0 {
		return fmt.Errorf("%w: time budget ran out during the %s turn", errLiveAborted, p.Slot)
	}
	return err
}

// step runs a turn and checks its postcondition; it retries once only while
// the budget allows, and reports errLiveAborted when a needed call would
// exceed it.
func (f *factoryCardFixture) step(t *testing.T, budget *liveBudget, label string, p factorymsg.Peer, prompt string, check func() error) error {
	t.Helper()
	var last error
	for try := 0; try < 2; try++ {
		if ok, reason := budget.take(); !ok {
			return fmt.Errorf("%w: %s needed another call: %s", errLiveAborted, label, reason)
		}
		if err := f.turn(t, budget, p, prompt); err != nil {
			if errors.Is(err, errLiveAborted) {
				return err
			}
			last = fmt.Errorf("%s turn: %v", label, err)
			continue
		}
		if last = check(); last == nil {
			return nil
		}
		t.Logf("%s postcondition not met (try %d): %v", label, try+1, last)
	}
	return last
}

func (f *factoryCardFixture) expectStatus(pending, claimed, acked int) func() error {
	return func() error {
		st, err := f.store.Status(context.Background())
		if err != nil {
			return err
		}
		if st.Pending != pending || st.Claimed != claimed || st.Acknowledged != acked {
			return fmt.Errorf("status pending=%d claimed=%d acknowledged=%d, want %d/%d/%d", st.Pending, st.Claimed, st.Acknowledged, pending, claimed, acked)
		}
		return nil
	}
}

func runFactoryCardFlow(t *testing.T, tc factoryLiveCase) {
	evDir := requireFactoryCardLive(t, tc.name)
	budget := newLiveBudget(factoryCardLiveInvocations, factoryCardLiveWindow)
	// Cleanup is registered here, before the first spawned process.
	procs := newLiveProcs(t)
	f := newFactoryCardFixture(t, procs, tc.lead, tc.worker)
	ctx := context.Background()
	id, n := f.dispatchID, f.nonce
	ev := map[string]any{"case": tc.name, "nonce": n, "dispatch_id": id, "run_id": f.runID, "budget_invocations": factoryCardLiveInvocations, "budget_seconds": factoryCardLiveWindow.Seconds(), "harness_role": "lane runtime: explicit assignee transitions, reassignment with revoke, result application before receipt"}
	outcomes := map[string]string{}
	applied := 0

	flow := func() error {
		if _, err := f.store.CreateDispatch(ctx, id, "t1100", f.first); err != nil {
			return fmt.Errorf("create dispatch: %v", err)
		}
		send := func(slot string, attempt int64) string {
			return fmt.Sprintf("Use factory_msg_send exactly once: run_id=%s to_slot=%s kind=dispatch_notice idempotency_key=%s task_ref=t1100 correlation_id=assign-%d-%s ttl_seconds=1200 body=assign dispatch=%s attempt=%d nonce=%s. Report only the returned message id.", f.runID, slot, factorymsg.AssignmentKey(id, attempt), attempt, n, id, attempt, n)
		}
		receive := "For run_id=" + f.runID + " call factory_msg_list, read every listed body with factory_msg_body, then call factory_msg_receipt with disposition=accepted for each. Do not send any message."
		result := func(attempt int64, preface string) string {
			return fmt.Sprintf("%sUse factory_msg_send exactly once: run_id=%s to_slot=lead kind=status_report idempotency_key=%s task_ref=t1100 correlation_id=result-%d-%s ttl_seconds=1200 body=result dispatch=%s attempt=%d nonce=%s. Report only the returned message id.", preface, f.runID, factorymsg.ResultKey(id, attempt), attempt, n, id, attempt, n)
		}
		if err := f.step(t, budget, "lead assigns attempt 1", f.lead, send("agent-1", 1), f.expectStatus(1, 0, 0)); err != nil {
			return err
		}
		if err := f.step(t, budget, "worker 1 receives", f.first, receive, f.expectStatus(0, 0, 1)); err != nil {
			return err
		}
		if _, err := f.store.MarkDispatchDelivered(ctx, f.first, id, 1); err != nil {
			return fmt.Errorf("delivered 1: %v", err)
		}
		if _, err := f.store.StartDispatch(ctx, f.first, id, 1); err != nil {
			return fmt.Errorf("start 1: %v", err)
		}
		// Worker 1 is interrupted after started: the lead revokes and reassigns.
		if d, err := f.store.ReassignDispatch(ctx, id, 1, f.second, true); err != nil || d.Attempt != 2 {
			return fmt.Errorf("reassign: %+v %v", d, err)
		}
		if err := f.step(t, budget, "lead assigns attempt 2", f.lead, send("agent-2", 2), f.expectStatus(1, 0, 1)); err != nil {
			return err
		}
		if err := f.step(t, budget, "worker 2 receives", f.second, receive, f.expectStatus(0, 0, 2)); err != nil {
			return err
		}
		if _, err := f.store.MarkDispatchDelivered(ctx, f.second, id, 2); err != nil {
			return fmt.Errorf("delivered 2: %v", err)
		}
		if _, err := f.store.StartDispatch(ctx, f.second, id, 2); err != nil {
			return fmt.Errorf("start 2: %v", err)
		}
		if err := f.step(t, budget, "worker 2 reports", f.second, result(2, ""), f.expectStatus(1, 0, 2)); err != nil {
			return err
		}
		if err := f.step(t, budget, "worker 1 reports late", f.first, result(1, "You are worker agent-1 resuming after an interruption. "), f.expectStatus(2, 0, 2)); err != nil {
			return err
		}
		claims, err := f.store.Claim(ctx, f.lead, factorymsg.MaxBatch, time.Minute)
		if err != nil || len(claims) != 2 {
			return fmt.Errorf("lead claim=%d err=%v, want 2 results", len(claims), err)
		}
		for _, cl := range claims {
			body, err := f.store.ReadBody(ctx, f.lead, cl.ID, cl.ClaimToken)
			if err != nil {
				return err
			}
			var attempt int64
			if m := factoryCardAttemptPattern.FindSubmatch(body); m != nil {
				_, _ = fmt.Sscan(string(m[1]), &attempt)
			}
			if !strings.Contains(string(body), n) {
				return fmt.Errorf("result %s body lacks the run nonce: %q", cl.ID, body)
			}
			outcome, err := f.store.ApplyResult(ctx, factorymsg.ResultReport{DispatchID: id, Attempt: attempt, Slot: cl.SenderSlot, Generation: cl.SenderGeneration, Digest: factorymsg.ResultDigest(body), Ref: cl.ID})
			if err != nil {
				return fmt.Errorf("apply %s: %v", cl.ID, err)
			}
			outcomes[cl.SenderSlot] = outcome
			if outcome == factorymsg.ApplyAccepted {
				applied++
			}
			if err := f.store.RecordDisposition(ctx, f.lead, cl.ID, cl.ClaimToken, factorymsg.DispositionAccepted); err != nil {
				return err
			}
			if err := f.store.Receipt(ctx, f.lead, cl.ID, cl.ClaimToken); err != nil {
				return err
			}
		}
		d, err := f.store.IntegrateDispatch(ctx, id)
		if err != nil {
			return fmt.Errorf("integrate: %v", err)
		}
		ev["final_state"], ev["result_attempt"], ev["result_lane"] = d.State, d.ResultAttempt, d.LaneSlot
		return nil
	}
	flowErr := flow()
	ev["applied_results"] = applied
	ev["late_result_outcome"] = outcomes["agent-1"]
	ev["current_result_outcome"] = outcomes["agent-2"]
	ev["invocations"] = budget.used
	ev["elapsed_seconds"] = budget.elapsedSeconds()
	ev["aborted"] = errors.Is(flowErr, errLiveAborted)
	if flowErr != nil {
		ev["error"] = flowErr.Error()
	}
	ev["cleaned_pids"] = procs.reap()
	if f.codexHome != "" {
		codexRoleExportSessions(t, f.codexHome, filepath.Join(evDir, "ac018-sessions-"+tc.name))
	}
	emitLiveEvidence(t, evDir, "ac018-evidence-"+tc.name+".json", "AC018", tc.name, ev)
	if errors.Is(flowErr, errLiveAborted) {
		t.Fatalf("ABORTED %v", flowErr)
	}
	if flowErr != nil {
		t.Fatal(flowErr)
	}
	if applied != 1 || outcomes["agent-1"] != factorymsg.ApplyStale || ev["final_state"] != factorymsg.DispatchIntegrated || ev["result_attempt"] != int64(2) {
		t.Fatalf("card flow result: applied=%d outcomes=%v state=%v attempt=%v", applied, outcomes, ev["final_state"], ev["result_attempt"])
	}
}
