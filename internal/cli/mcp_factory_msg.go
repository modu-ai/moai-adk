package cli

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

var factoryProbeProcessIdentity = homestate.ProbeProcessIdentity

// The two unknown-attribution sentinels (SPEC-MCP-NEUTRALITY-001 REQ-MN-003):
// nothing carries the caller's identity — no actor argument, no session id,
// no live owner. Every other peer-resolution error (an unregistered id, a
// store failure) keeps the error channel; it is not an attribution class.
var (
	errFactoryAttributionUnavailable = errors.New("factory endpoint attribution unavailable")
	errFactoryOwnerNotLive           = errors.New("factory endpoint owner is not live")
)

// currentFactoryPeer resolves the attributed endpoint (SPEC-MCP-NEUTRALITY-001
// REQ-MN-003): an explicit actor argument — the caller's session id — wins,
// and the spawn-time resolution (authoritative session id, then the owner
// process probe) is the default only.
func currentFactoryPeer(ctx context.Context, s *factorymsg.Store, actor string) (factorymsg.Peer, error) {
	if id := strings.TrimSpace(actor); id != "" {
		return s.Peer(ctx, id)
	}
	if id, source, ok := resolveCurrentSessionID(); ok && sessionIDSourceIsAuthoritative(source) {
		return s.Peer(ctx, id)
	}
	pid := os.Getppid()
	if raw := os.Getenv(config.EnvMoaiSessionPID); raw != "" {
		var err error
		pid, err = strconv.Atoi(raw)
		if err != nil || pid < 1 {
			return factorymsg.Peer{}, errFactoryAttributionUnavailable
		}
	}
	fp, state := factoryProbeProcessIdentity(pid)
	if state != homestate.ProcessIdentityLive {
		return factorymsg.Peer{}, errFactoryOwnerNotLive
	}
	return s.PeerByOwner(ctx, pid, fp)
}

// factoryStore opens the broker for the requested run, anchored at the
// caller's project_root (SPEC-MCP-NEUTRALITY-001 REQ-MN-004). requiredRoot is
// true for the mutating members (send/list/receipt): an absent argument
// refuses rather than falling back to the server's spawn tree; the read-only
// members fall back and the result carries the _root provenance. The two
// refusal classes the run id carries surface as structured results
// (REQ-MN-001): a run that is not the active one refuses stale-run, while a
// malformed id stays an argument error on the error channel. A non-nil store
// implies a nil refusal and vice versa.
func factoryStore(req mcp.CallToolRequest, tool string, requiredRoot bool) (*factorymsg.Store, *mcp.CallToolResult) {
	run := req.GetString("run_id", "")
	if run == "" {
		return nil, toolErr(tool, errors.New("run_id is required"))
	}
	var root string
	var err error
	if requiredRoot {
		root, _, err = resolveRequiredToolProjectRoot(req)
	} else {
		root, _, err = resolveToolProjectRootWithSource(req)
	}
	if err != nil {
		return nil, toolErr(tool, err)
	}
	if err := factorymsg.ValidateActiveRun(context.Background(), root, run); err != nil {
		if errors.Is(err, factorymsg.ErrRunNotActive) {
			return nil, toolRefusal(mcpRefusal{
				Refused: true,
				Code:    mcpRefusalStaleRun,
				Reason:  fmt.Sprintf("run %q is not the currently active factory run", run),
				Hint:    "pass the active run id",
			})
		}
		return nil, toolErr(tool, err)
	}
	s, err := factorymsg.Open(root, run)
	if err != nil {
		return nil, toolErr(tool, err)
	}
	return s, nil
}

// mcpFactoryMsgPeer resolves the attributed endpoint for one handler call and
// shapes the unknown-attribution class as a structured refusal (REQ-MN-003):
// no actor argument, no authoritative session id, and no live owner leaves
// nothing to attribute. Every other peer error keeps the error channel.
func mcpFactoryMsgPeer(ctx context.Context, s *factorymsg.Store, req mcp.CallToolRequest, tool string) (factorymsg.Peer, *mcp.CallToolResult) {
	p, e := currentFactoryPeer(ctx, s, req.GetString("actor", ""))
	if e != nil {
		if errors.Is(e, errFactoryAttributionUnavailable) || errors.Is(e, errFactoryOwnerNotLive) {
			return factorymsg.Peer{}, toolRefusal(mcpIdentityRefusal(e.Error()))
		}
		return factorymsg.Peer{}, toolErr(tool, e)
	}
	return p, nil
}

func handleFactoryMsgSend(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s, refusal := factoryStore(req, "factory_msg_send", true)
	if refusal != nil {
		return refusal, nil
	}
	defer closeFactoryToolStore("factory_msg_send", s)
	from, refusal := mcpFactoryMsgPeer(ctx, s, req, "factory_msg_send")
	if refusal != nil {
		return refusal, nil
	}
	to, e := s.ResolveLane(ctx, req.GetString("to_slot", ""))
	if e != nil {
		return toolErr("factory_msg_send", e), nil
	}
	// task_ref and correlation_id are optional in the tool schema while the
	// store requires both on every message: an omitted or empty value defaults
	// to the idempotency key (GitHub #1737), so a same-key retry carries the
	// same identifiers and the store's collision check still holds.
	idem := req.GetString("idempotency_key", "")
	env, e := s.Send(ctx, factorymsg.SendRequest{From: from, To: to, Kind: req.GetString("kind", ""), IdempotencyKey: idem, TaskRef: cmp.Or(req.GetString("task_ref", ""), idem), CorrelationID: cmp.Or(req.GetString("correlation_id", ""), idem), ExpectedTaskRevision: int64(req.GetInt("expected_task_revision", 0)), CurrentTaskRevision: int64(req.GetInt("current_task_revision", 0)), TTL: time.Duration(req.GetInt("ttl_seconds", 3600)) * time.Second, Payload: []byte(req.GetString("body", ""))})
	if e != nil {
		return toolErr("factory_msg_send", e), nil
	}
	return toolJSON("factory_msg_send", env), nil
}
func handleFactoryMsgList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s, refusal := factoryStore(req, "factory_msg_list", true)
	if refusal != nil {
		return refusal, nil
	}
	defer closeFactoryToolStore("factory_msg_list", s)
	p, refusal := mcpFactoryMsgPeer(ctx, s, req, "factory_msg_list")
	if refusal != nil {
		return refusal, nil
	}
	claims, e := s.Claim(ctx, p, req.GetInt("limit", factorymsg.MaxBatch), time.Duration(req.GetInt("lease_seconds", 30))*time.Second)
	if e != nil {
		return toolErr("factory_msg_list", e), nil
	}
	return toolJSON("factory_msg_list", map[string]any{"messages": claims, "body_lookup": "factory_msg_body"}), nil
}
func handleFactoryMsgBody(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s, refusal := factoryStore(req, "factory_msg_body", false)
	if refusal != nil {
		return refusal, nil
	}
	defer closeFactoryToolStore("factory_msg_body", s)
	p, refusal := mcpFactoryMsgPeer(ctx, s, req, "factory_msg_body")
	if refusal != nil {
		return refusal, nil
	}
	body, e := s.ReadBody(ctx, p, req.GetString("message_id", ""), req.GetString("claim_token", ""))
	if e != nil {
		return toolErr("factory_msg_body", e), nil
	}
	return toolJSON("factory_msg_body", map[string]any{"message_id": req.GetString("message_id", ""), "body": string(body), "trust": "untrusted peer data"}), nil
}
func handleFactoryMsgReceipt(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s, refusal := factoryStore(req, "factory_msg_receipt", true)
	if refusal != nil {
		return refusal, nil
	}
	defer closeFactoryToolStore("factory_msg_receipt", s)
	p, refusal := mcpFactoryMsgPeer(ctx, s, req, "factory_msg_receipt")
	if refusal != nil {
		return refusal, nil
	}
	id, tok := req.GetString("message_id", ""), req.GetString("claim_token", "")
	var e error
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
	// The read-only family member keeps the fallback resolution (REQ-MN-004)
	// and annotates which tree it read.
	root, source, err := resolveToolProjectRootWithSource(req)
	if err != nil {
		return toolErr("factory_msg_status", err), nil
	}
	if e := factorymsg.ValidateActiveRun(ctx, root, run); e != nil {
		if errors.Is(e, factorymsg.ErrRunNotActive) {
			return toolRefusal(mcpRefusal{
				Refused: true,
				Code:    mcpRefusalStaleRun,
				Reason:  fmt.Sprintf("run %q is not the currently active factory run", run),
				Hint:    "pass the active run id",
			}), nil
		}
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
	return toolJSONWithRoot("factory_msg_status", st, root, source), nil
}

// closeFactoryToolStore closes a broker handle after a tool call has produced
// its result. The result already reflects what the store committed, so a close
// failure is logged rather than rewriting a completed call into a tool error.
func closeFactoryToolStore(tool string, s *factorymsg.Store) {
	if err := s.Close(); err != nil {
		slog.Warn("factory message broker close failed", "tool", tool, "error", err)
	}
}
