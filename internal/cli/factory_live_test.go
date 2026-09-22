package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	store                    *factorymsg.Store
	lead, worker             factorymsg.Peer
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
	env := os.Environ()
	if p.Backend == "claude" {
		env = factoryLiveWithoutAttribution(env)
		key := factoryLiveOperatorGLMKey()
		if key == "" {
			t.Fatal("operator GLM credential unavailable")
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
