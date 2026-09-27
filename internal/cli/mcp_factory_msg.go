package cli

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

var factoryProbeProcessIdentity = homestate.ProbeProcessIdentity

func currentFactoryPeer(ctx context.Context, s *factorymsg.Store) (factorymsg.Peer, error) {
	if id, source, ok := resolveCurrentSessionID(); ok && sessionIDSourceIsAuthoritative(source) {
		return s.Peer(ctx, id)
	}
	pid := os.Getppid()
	if raw := os.Getenv(config.EnvMoaiSessionPID); raw != "" {
		var err error
		pid, err = strconv.Atoi(raw)
		if err != nil || pid < 1 {
			return factorymsg.Peer{}, errors.New("factory endpoint attribution unavailable")
		}
	}
	fp, state := factoryProbeProcessIdentity(pid)
	if state != homestate.ProcessIdentityLive {
		return factorymsg.Peer{}, errors.New("factory endpoint owner is not live")
	}
	return s.PeerByOwner(ctx, pid, fp)
}
func factoryStore(req mcp.CallToolRequest) (*factorymsg.Store, error) {
	run := req.GetString("run_id", "")
	if run == "" {
		return nil, errors.New("run_id is required")
	}
	root := resolveProjectDir()
	if err := factorymsg.ValidateActiveRun(context.Background(), root, run); err != nil {
		return nil, err
	}
	return factorymsg.Open(root, run)
}
func handleFactoryMsgSend(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s, e := factoryStore(req)
	if e != nil {
		return toolErr("factory_msg_send", e), nil
	}
	defer closeFactoryToolStore("factory_msg_send", s)
	from, e := currentFactoryPeer(ctx, s)
	if e != nil {
		return toolErr("factory_msg_send", e), nil
	}
	to, e := s.ResolveLane(ctx, req.GetString("to_slot", ""))
	if e != nil {
		return toolErr("factory_msg_send", e), nil
	}
	env, e := s.Send(ctx, factorymsg.SendRequest{From: from, To: to, Kind: req.GetString("kind", ""), IdempotencyKey: req.GetString("idempotency_key", ""), TaskRef: req.GetString("task_ref", ""), CorrelationID: req.GetString("correlation_id", ""), ExpectedTaskRevision: int64(req.GetInt("expected_task_revision", 0)), CurrentTaskRevision: int64(req.GetInt("current_task_revision", 0)), TTL: time.Duration(req.GetInt("ttl_seconds", 3600)) * time.Second, Payload: []byte(req.GetString("body", ""))})
	if e != nil {
		return toolErr("factory_msg_send", e), nil
	}
	return toolJSON("factory_msg_send", env), nil
}
func handleFactoryMsgList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s, e := factoryStore(req)
	if e != nil {
		return toolErr("factory_msg_list", e), nil
	}
	defer closeFactoryToolStore("factory_msg_list", s)
	p, e := currentFactoryPeer(ctx, s)
	if e != nil {
		return toolErr("factory_msg_list", e), nil
	}
	claims, e := s.Claim(ctx, p, req.GetInt("limit", factorymsg.MaxBatch), time.Duration(req.GetInt("lease_seconds", 30))*time.Second)
	if e != nil {
		return toolErr("factory_msg_list", e), nil
	}
	return toolJSON("factory_msg_list", map[string]any{"messages": claims, "body_lookup": "factory_msg_body"}), nil
}
func handleFactoryMsgBody(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s, e := factoryStore(req)
	if e != nil {
		return toolErr("factory_msg_body", e), nil
	}
	defer closeFactoryToolStore("factory_msg_body", s)
	p, e := currentFactoryPeer(ctx, s)
	if e != nil {
		return toolErr("factory_msg_body", e), nil
	}
	body, e := s.ReadBody(ctx, p, req.GetString("message_id", ""), req.GetString("claim_token", ""))
	if e != nil {
		return toolErr("factory_msg_body", e), nil
	}
	return toolJSON("factory_msg_body", map[string]any{"message_id": req.GetString("message_id", ""), "body": string(body), "trust": "untrusted peer data"}), nil
}
func handleFactoryMsgReceipt(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s, e := factoryStore(req)
	if e != nil {
		return toolErr("factory_msg_receipt", e), nil
	}
	defer closeFactoryToolStore("factory_msg_receipt", s)
	p, e := currentFactoryPeer(ctx, s)
	if e != nil {
		return toolErr("factory_msg_receipt", e), nil
	}
	id, tok := req.GetString("message_id", ""), req.GetString("claim_token", "")
	if e = s.RecordDisposition(ctx, p, id, tok, req.GetString("disposition", "")); e == nil {
		e = s.Receipt(ctx, p, id, tok)
	}
	if e != nil {
		return toolErr("factory_msg_receipt", e), nil
	}
	return toolJSON("factory_msg_receipt", map[string]any{"message_id": id, "acknowledged": true}), nil
}
func handleFactoryMsgStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	run := req.GetString("run_id", "")
	if run == "" {
		return toolErr("factory_msg_status", errors.New("run_id is required")), nil
	}
	root := resolveProjectDir()
	if e := factorymsg.ValidateActiveRun(ctx, root, run); e != nil {
		return toolErr("factory_msg_status", e), nil
	}
	s, e := factorymsg.OpenExistingWithDeadline(root, run, 5*time.Second)
	if e != nil {
		return toolErr("factory_msg_status", e), nil
	}
	defer closeFactoryToolStore("factory_msg_status", s)
	st, e := s.Status(ctx)
	if e != nil {
		return toolErr("factory_msg_status", e), nil
	}
	return toolJSON("factory_msg_status", st), nil
}

// closeFactoryToolStore closes a broker handle after a tool call has produced
// its result. The result already reflects what the store committed, so a close
// failure is logged rather than rewriting a completed call into a tool error.
func closeFactoryToolStore(tool string, s *factorymsg.Store) {
	if err := s.Close(); err != nil {
		slog.Warn("factory message broker close failed", "tool", tool, "error", err)
	}
}
