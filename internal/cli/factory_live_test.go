package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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
		t.Fatal("MOAI_FACTORY_LIVE=1 is required; an unexecuted live criterion is a failure")
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

func newFactoryLiveFixture(t *testing.T, leadBackend, workerBackend string) *factoryLiveFixture {
	t.Helper()
	root := t.TempDir()
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
	if err := os.WriteFile(filepath.Join(root, ".codex", "config.toml"), []byte(fmt.Sprintf("[mcp_servers.moai]\ncommand = %q\nargs = [\"mcp-server\"]\ndefault_tools_approval_mode = \"writes\"\n", moai)), 0o600); err != nil {
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
	fp, state := homestate.ProbeProcessIdentity(os.Getpid())
	if (leadBackend == "codex" || workerBackend == "codex") && (state != homestate.ProcessIdentityLive || fp == "") {
		t.Fatalf("Codex owner fingerprint unavailable: state=%s", state)
	}
	base := factorymsg.Peer{ProjectKey: homestate.ProjectKey(root), RunID: runID, Generation: 1, PID: os.Getpid(), ProcessStart: fp}
	lead := base
	lead.Backend, lead.Role, lead.Slot, lead.SessionUUID = leadBackend, "lead", "lead", "live-lead-"+factoryLiveID()
	worker := base
	worker.Backend, worker.Role, worker.Slot, worker.SessionUUID = workerBackend, "worker", "agent-1", "live-worker-"+factoryLiveID()
	if lead.ProcessStart == "" {
		lead.ProcessStart = "claude-authoritative"
	}
	if worker.ProcessStart == "" {
		worker.ProcessStart = "claude-authoritative"
	}
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

func (f *factoryLiveFixture) runModel(t *testing.T, p factorymsg.Peer, prompt string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
	defer cancel()
	var cmd *exec.Cmd
	switch p.Backend {
	case "codex":
		bin, err := exec.LookPath("codex")
		if err != nil {
			t.Fatal(err)
		}
		cmd = exec.CommandContext(ctx, bin, "exec", "--skip-git-repo-check", "--json", prompt)
	case "claude":
		bin, err := exec.LookPath("claude")
		if err != nil {
			t.Fatal(err)
		}
		cmd = exec.CommandContext(ctx, bin, "-p", "--output-format", "json", "--mcp-config", filepath.Join(f.root, ".mcp.json"), "--strict-mcp-config", "--permission-mode", "bypassPermissions", prompt)
	default:
		t.Fatalf("unknown backend %q", p.Backend)
	}
	cmd.Dir = f.root
	env := append(os.Environ(), config.EnvMoaiKanbanID+"="+f.runID, config.EnvMoaiSessionPID+"="+fmt.Sprint(os.Getpid()), config.EnvClaudeProjectDir+"="+f.root)
	if p.Backend == "claude" {
		env = append(env, config.EnvClaudeCodeSessionID+"="+p.SessionUUID)
	} else {
		env = append(env, config.EnvClaudeCodeSessionID+"=")
	}
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s live turn failed: %v: %s", p.Backend, err, strings.TrimSpace(string(out)))
	}
}
