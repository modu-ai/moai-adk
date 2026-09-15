package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/codexbridge"
	"github.com/modu-ai/moai-adk/internal/codextools"
	"github.com/modu-ai/moai-adk/internal/gateway"
	"github.com/modu-ai/moai-adk/internal/gateway/receipt"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// newGPTAppServerGatewayHandler is the only production GPT assembly. It owns
// the official Codex App Server process and never opens the legacy token store,
// exports a bearer token, or calls an OpenAI model endpoint directly.
func newGPTAppServerGatewayHandler(raw json.RawMessage) (http.Handler, error) {
	p, rows, err := decodeGPTAppServerPayload(raw)
	if err != nil {
		return nil, errGatewayFactory
	}
	profile, err := managedGPTProfile()
	if err != nil {
		return nil, errGatewayFactory
	}
	binary, err := exec.LookPath("codex")
	if err != nil {
		return nil, errGatewayFactory
	}
	binary, err = filepath.Abs(binary)
	if err != nil {
		return nil, errGatewayFactory
	}

	lifetime, cancel := context.WithCancel(context.Background())
	client, err := startSharedGPTAppServer(lifetime, codexapp.Config{Binary: binary, Home: profile})
	if err != nil {
		cancel()
		return nil, errGatewayFactory
	}
	cleanupClient := true
	defer func() {
		if cleanupClient {
			cancel()
			_ = client.Close()
		}
	}()
	if _, err = client.Initialize(lifetime, "moai-gateway", "1"); err != nil {
		return nil, errGatewayFactory
	}
	authority, err := gateway.NewAppServerAuthority(client, "chatgpt", func(context.Context) (string, error) {
		return managedGPTProfileGeneration(profile)
	})
	if err != nil {
		return nil, errGatewayFactory
	}
	// Fail before Claude starts when the private App Server profile is not a
	// managed ChatGPT subscription session.
	if _, err = authority.Authorize(lifetime, rows[0]); err != nil {
		return nil, errGatewayFactory
	}

	bridgeDir := filepath.Join(profile, "moai-bridge")
	if err = os.MkdirAll(bridgeDir, 0o700); err != nil {
		return nil, errGatewayFactory
	}
	if err = os.Chmod(bridgeDir, 0o700); err != nil {
		return nil, errGatewayFactory
	}
	store, err := codexbridge.OpenStore(bridgeDir)
	if err != nil {
		return nil, errGatewayFactory
	}
	engine, err := codexbridge.New(lifetime, client, codexbridge.Config{Store: store, MaxOutputBytes: 8 << 20})
	if err != nil {
		return nil, errGatewayFactory
	}

	var receipts *receipt.Store
	limits := translate.Limits{PolicyProfile: translate.PolicyGPTNative, MaxBodyBytes: gatewayRequestBodyLimit, MaxEventBytes: 1 << 20, MaxOutputBytes: 8 << 20}
	conversationID := gatewayTokenConversationID(p.SessionToken)
	cwd, cwdErr := os.Getwd()
	if cwdErr != nil || !filepath.IsAbs(cwd) {
		engine.Close()
		return nil, errGatewayFactory
	}
	if p.Conversation != nil {
		conversationID = p.Conversation.SessionID
		if p.Conversation.CWD != "" {
			cwd = p.Conversation.CWD
		}
		receipts, err = receipt.OpenStore(lifetime, p.Conversation.ReceiptDir, p.Conversation.SessionID, false)
		if err != nil {
			engine.Close()
			return nil, errGatewayFactory
		}
		limits.NativeReceiptAuthorize = func(ctx context.Context, body []byte, policy translate.NativePolicy) error {
			return authorizeManagedGPTNativeReceipt(ctx, body, policy, *p.Conversation, receipts)
		}
	}
	var conversations []gatewayPrivateConversation
	if p.Conversation != nil {
		conversations = append(conversations, *p.Conversation)
	}
	prepare := newManagedGPTPrepare(conversationID, cwd, limits, store, conversations...)
	adapter, err := gateway.NewAppServerAdapter(gateway.AppServerAdapterConfig{Engine: engine, Authority: authority, Prepare: prepare})
	if err != nil {
		engine.Close()
		if receipts != nil {
			_ = receipts.Close()
		}
		return nil, errGatewayFactory
	}
	catalog, err := gateway.NewCatalog(rows)
	if err != nil {
		engine.Close()
		if receipts != nil {
			_ = receipts.Close()
		}
		return nil, errGatewayFactory
	}
	rejections, err := newManagedGPTRejectionLogger(profile, conversationID)
	if err != nil {
		engine.Close()
		if receipts != nil {
			_ = receipts.Close()
		}
		return nil, errGatewayFactory
	}
	server, err := gateway.NewServer(gateway.ServerConfig{
		RejectionLogger:  rejections,
		ManagedAuthority: authority,
		SessionHeader:    "Authorization",
		SessionToken:     "Bearer " + p.SessionToken,
		MaxBodyBytes:     gatewayRequestBodyLimit,
		Catalog:          catalog,
		Adapters:         map[gateway.ProviderID]gateway.Adapter{gateway.ProviderOpenAI: adapter},
	})
	if err != nil {
		engine.Close()
		if receipts != nil {
			_ = receipts.Close()
		}
		return nil, errGatewayFactory
	}
	requestCtx, requestCancel := context.WithCancel(context.Background())
	cleanupClient = false
	return &gatewayOwnedHandler{Handler: server, receipts: receipts, ctx: requestCtx, cancel: requestCancel, managed: func() error {
		engine.Close()
		cancel()
		return client.Close()
	}}, nil
}

func decodeGPTAppServerPayload(raw json.RawMessage) (gatewayPrivatePayload, []gateway.ModelEntry, error) {
	var p gatewayPrivatePayload
	if len(raw) > 65536 || gateway.ValidateJSONObject(raw) != nil {
		return p, nil, errGatewayFactory
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if dec.Decode(&p) != nil || p.Version != 1 || p.SessionToken == "" || len(p.SessionToken) > 4096 || len(p.ModelIDs) == 0 || len(p.ModelIDs) > 4 {
		return p, nil, errGatewayFactory
	}
	if p.Conversation != nil && !validGatewayConversation(*p.Conversation) {
		return p, nil, errGatewayFactory
	}
	for _, r := range p.SessionToken {
		if r < 33 || r > 126 {
			return p, nil, errGatewayFactory
		}
	}
	approved := make(map[string]gateway.ModelEntry, 4)
	// Legacy private v1 payloads had the fixed 272K launch window. New payloads
	// carry the parent's validated snapshot; never re-read mutable cache here.
	window := p.ContextTokens
	if window == 0 {
		var fields map[string]json.RawMessage
		if json.Unmarshal(raw, &fields) != nil || fields["context_tokens"] != nil {
			return p, nil, errGatewayFactory
		}
		window = gatewayContextWindow
	}
	if window < 0 || window > gatewayContextWindowLimit {
		return p, nil, errGatewayFactory
	}
	for _, row := range gatewayGPTModels() {
		row.Capabilities.ContextTokens = window
		approved[row.RouteID] = row
	}
	seen := make(map[string]bool, len(p.ModelIDs))
	rows := make([]gateway.ModelEntry, 0, len(p.ModelIDs))
	for _, id := range p.ModelIDs {
		row, ok := approved[id]
		if !ok || seen[id] {
			return p, nil, errGatewayFactory
		}
		seen[id] = true
		rows = append(rows, row)
	}
	return p, rows, nil
}

func managedGPTProfile() (string, error) {
	home, err := paths.MoaiHome()
	if err != nil || !filepath.IsAbs(home) {
		return "", errGatewayFactory
	}
	profile := filepath.Join(home, "gpt-appserver")
	if err = os.MkdirAll(profile, 0o700); err != nil {
		return "", err
	}
	if err = os.Chmod(profile, 0o700); err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(profile)
}

func managedGPTProfileGeneration(profile string) (string, error) {
	info, err := os.Lstat(filepath.Join(profile, "auth.json"))
	marker := "managed-appserver-profile-v1\x00" + profile
	if err == nil {
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return "", errors.New("managed App Server authentication state is invalid")
		}
		marker += "\x00" + info.ModTime().UTC().String() + "\x00" + strconv.FormatInt(info.Size(), 10)
	} else if !os.IsNotExist(err) {
		return "", err
	}
	sum := sha256.Sum256([]byte(marker))
	return hex.EncodeToString(sum[:]), nil
}

func gatewayTokenConversationID(token string) string {
	sum := sha256.Sum256([]byte("moai-gateway-conversation\x00" + token))
	return hex.EncodeToString(sum[:])
}

func newManagedGPTPrepare(conversationID, cwd string, base translate.Limits, store *codexbridge.FileStore, conversations ...gatewayPrivateConversation) func(context.Context, gateway.RoutedRequest) (codexbridge.Request, error) {
	return func(ctx context.Context, r gateway.RoutedRequest) (codexbridge.Request, error) {
		diagnostic := r.SummaryDiagnostic
		if diagnostic == nil {
			diagnostic = &gateway.AppServerSummaryDiagnostic{}
		}
		diagnostic.AgentHeaderPresent = r.Headers.Get("X-Claude-Code-Agent-Id") != ""
		diagnostic.PromptMatch = managedGPTAgentSummaryShape(r.Body, diagnostic)
		agentSummary := diagnostic.AgentHeaderPresent && diagnostic.PromptMatch
		hookAgent := managedGPTHookAgentDigest(r.Body, r.Headers)
		if !diagnostic.AgentHeaderPresent && hookAgent == "" && managedGPTHookVerifierText(r.Body) != "" {
			return codexbridge.Request{}, fmt.Errorf("%w: unsupported hook agent shape", codexbridge.ErrProtocol)
		}
		diagnostic.Classified = agentSummary
		if store == nil {
			return codexbridge.Request{}, codexbridge.ErrScope
		}
		limits := base
		limits.CredentialScope = r.Managed.Scope()
		projected, _, err := translate.RequestContext(ctx, r.Entry.UpstreamID, r.Body, limits)
		if err != nil {
			return codexbridge.Request{}, err
		}
		var upstream struct {
			Input        []any  `json:"input"`
			Instructions string `json:"instructions"`
			Text         struct {
				Format struct {
					Schema json.RawMessage `json:"schema"`
				} `json:"format"`
			} `json:"text"`
			Reasoning struct {
				Effort string `json:"effort"`
			} `json:"reasoning"`
		}
		if json.Unmarshal(projected, &upstream) != nil {
			return codexbridge.Request{}, codexbridge.ErrProtocol
		}
		tools, input, results, references, err := managedGPTPublicDelta(r.Body)
		if err != nil {
			if errors.Is(err, codexbridge.ErrScope) {
				return codexbridge.Request{}, fmt.Errorf("%w: prepare public delta", err)
			}
			return codexbridge.Request{}, err
		}
		sum := sha256.Sum256(projected)
		current := hex.EncodeToString(sum[:])
		effectiveConversation := conversationID
		if len(conversations) == 1 {
			effectiveConversation, err = managedGPTConversationForRequest(ctx, conversations[0], r.Headers, r.Body)
			if err != nil {
				return codexbridge.Request{}, err
			}
		}
		ownerID, err := managedGPTRequestOwner(effectiveConversation, r.Headers)
		if err != nil {
			return codexbridge.Request{}, fmt.Errorf("%w: prepare owner binding", err)
		}
		contextDigest := managedGPTNativeContextDigest(r.Body)
		if contextDigest != "" {
			ownerID += ":context:" + contextDigest
		}
		ephemeral := false
		if hookAgent != "" {
			ownerID += ":hook-agent:" + hookAgent
			ephemeral = true
		} else if managedGPTTitleRequest(r.Body) {
			// Claude Code generates the session title through a separate model
			// request with title-only developer instructions. Reusing that App
			// Server thread for the coding session permanently pins the title
			// instructions onto the main conversation.
			ownerID += ":title"
			ephemeral = true
		} else if agentSummary {
			// Observed in Claude Code 2.1.270: summary requests retain the
			// child's agent header AND system prompt, but append this separate
			// progress request to public history. They must never advance the
			// child's pending tool turn. Snapshot identity makes retries stable.
			ownerID += ":summary:" + current
			ephemeral = true
		} else if managedGPTCompactRequest(r.Body) {
			ownerID += ":compact:" + current
			ephemeral = true
		} else if len(upstream.Text.Format.Schema) != 0 {
			// Structured utility calls (including Claude prompt hooks) evaluate
			// a snapshot; they must not repin or fail the working conversation.
			ownerID += ":structured:" + current
			ephemeral = true
		}
		if len(ownerID) > 256 {
			return codexbridge.Request{}, fmt.Errorf("%w: prepare owner length", codexbridge.ErrScope)
		}
		q := codexbridge.Request{
			Owner:                 codextools.Binding{ConversationID: ownerID, AccountScope: r.Managed.Scope()},
			Model:                 r.Entry.UpstreamID,
			Effort:                upstream.Reasoning.Effort,
			CWD:                   cwd,
			Instructions:          upstream.Instructions,
			Tools:                 tools,
			References:            references,
			Input:                 input,
			Results:               results,
			PrefixDigest:          current,
			OutputSchema:          upstream.Text.Format.Schema,
			Ephemeral:             ephemeral,
			HookAgentTerminalTool: hookAgent != "",
		}
		if agentSummary || (ephemeral && len(upstream.Text.Format.Schema) != 0) {
			// Claude retains the child's tool catalog in summary requests. This
			// snapshot may read historical calls, but cannot execute or resolve any.
			q.Tools, q.References, q.Results = nil, nil, nil
			if history, images := managedGPTInheritedHistory(upstream.Input); history != "" {
				q.Input = append([]any{map[string]string{"type": "text", "text": history}}, images...)
			}
		}
		q.Instructions = managedGPTExecutionInstructions(q.Instructions, q.Tools)
		barrier, found, err := store.Barrier(q.Owner)
		if err != nil {
			return codexbridge.Request{}, err
		}
		// A rejected first thread/start retains a conservative durable barrier,
		// but no public input was accepted. Rebuild its initial context exactly.
		// Only the live Engine can permit retry: failed owners and a restarted
		// Engine still reject this existing barrier before sending another RPC.
		if found && barrier.Phase == "starting" && barrier.Prefix == "" {
			found = false
		}
		if found {
			if barrier.Phase == "idle" && len(q.Results) != 0 && !ephemeral {
				// Cancellation can complete after tool results were accepted but
				// before Claude receives an assistant message. Only an exact match
				// to the durable accepted prefix permits removing those old inputs.
				if delta := managedGPTAfterAcceptedPrefix(ctx, r, limits, barrier.Prefix); delta != nil {
					_, q.Input, q.Results, q.References, err = managedGPTPublicDelta(delta)
					if err != nil {
						return codexbridge.Request{}, err
					}
				}
			}
			if contextDigest != "" && len(q.Results) == 0 && barrier.Phase != "idle" {
				return codexbridge.Request{}, codexbridge.ErrRecovery
			}
			// A pending tool RPC still belongs to the model that issued it. Finish
			// that exact continuation before applying the user's new selection on
			// the next idle turn. The engine still checks prefix and pending IDs.
			if barrier.Model != q.Model && barrier.Phase == "waiting" && len(q.Input) == 0 && len(q.Results) != 0 && !ephemeral && managedGPTPureToolResultRequest(r.Body) {
				q.Model = barrier.Model
			}
			if (barrier.Model != q.Model && barrier.Phase != "idle") || barrier.CWD != q.CWD {
				return codexbridge.Request{}, fmt.Errorf("%w: prepare active model or cwd", codexbridge.ErrScope)
			}
			q.ExpectedPrefix = barrier.Prefix
			q.Resume = barrier.Phase == "idle"
		} else if r.Headers.Get("X-Claude-Code-Agent-Id") != "" || contextDigest != "" || ephemeral {
			// An agent may inherit a public conversation without inheriting an
			// App Server thread. The delta alone would discard its context, or
			// incorrectly answer a tool RPC owned by its parent. Import the
			// already-validated public history as data, never as executable RPCs.
			if history, images := managedGPTInheritedHistory(upstream.Input); history != "" {
				q.Input = []any{map[string]string{"type": "text", "text": history}}
				q.Input = append(q.Input, images...)
				q.Results = nil
			}
		}
		return q, nil
	}
}

// PublicDelta preserves auxiliary text as tool-result context. That compatibility
// does not authorize deferring a new user input onto a different selected model.
func managedGPTPureToolResultRequest(body []byte) bool {
	var root struct {
		Messages []struct {
			Role    string          `json:"role"`
			Content json.RawMessage `json:"content"`
		} `json:"messages"`
	}
	if json.Unmarshal(body, &root) != nil {
		return false
	}
	start := 0
	for i := len(root.Messages) - 1; i >= 0; i-- {
		if root.Messages[i].Role == "assistant" {
			start = i + 1
			break
		}
	}
	found := false
	for _, message := range root.Messages[start:] {
		if message.Role != "user" {
			return false
		}
		var blocks []struct {
			Type string `json:"type"`
		}
		if json.Unmarshal(message.Content, &blocks) != nil || len(blocks) == 0 {
			return false
		}
		for _, block := range blocks {
			if block.Type != "tool_result" {
				return false
			}
			found = true
		}
	}
	return found
}

// managedGPTAfterAcceptedPrefix proves the exact previously accepted projection
// before trimming anything. Unknown, edited, and cross-owner results retain the
// ordinary rejection path. This cold path runs only for an idle tool replay.
func managedGPTAfterAcceptedPrefix(ctx context.Context, r gateway.RoutedRequest, limits translate.Limits, accepted string) []byte {
	var root map[string]json.RawMessage
	var messages []json.RawMessage
	if accepted == "" || json.Unmarshal(r.Body, &root) != nil || json.Unmarshal(root["messages"], &messages) != nil {
		return nil
	}
	encode := func(items []json.RawMessage) []byte {
		root["messages"], _ = json.Marshal(items)
		body, _ := json.Marshal(root)
		return body
	}
	attempts := 0
	matches := func(items []json.RawMessage) bool {
		attempts++
		projected, _, err := translate.RequestContext(ctx, r.Entry.UpstreamID, encode(items), limits)
		if err != nil {
			return false
		}
		digest := sha256.Sum256(projected)
		return hex.EncodeToString(digest[:]) == accepted
	}
	for i := len(messages) - 1; i >= 0; i-- {
		if attempts >= 64 || ctx.Err() != nil {
			return nil
		}
		var message struct {
			Role    string          `json:"role"`
			Content json.RawMessage `json:"content"`
		}
		if json.Unmarshal(messages[i], &message) != nil || message.Role == "assistant" {
			break
		}
		if i+1 < len(messages) && matches(messages[:i+1]) {
			return encode(messages[i+1:])
		}
		var blocks []json.RawMessage
		if json.Unmarshal(message.Content, &blocks) != nil {
			continue
		}
		for j := len(blocks) - 1; j > 0; j-- {
			if attempts >= 64 || ctx.Err() != nil {
				return nil
			}
			message.Content, _ = json.Marshal(blocks[:j])
			prefix := append([]json.RawMessage(nil), messages[:i]...)
			last, _ := json.Marshal(message)
			prefix = append(prefix, last)
			if matches(prefix) {
				message.Content, _ = json.Marshal(blocks[j:])
				last, _ = json.Marshal(message)
				return encode(append([]json.RawMessage{last}, messages[i+1:]...))
			}
		}
	}
	return nil
}

func managedGPTInheritedHistory(input []any) (string, []any) {
	var public []any
	hasHistory := false
	for _, item := range input {
		value, ok := item.(map[string]any)
		if !ok || value["type"] == "reasoning" {
			continue
		}
		if value["role"] == "assistant" || value["type"] == "function_call" || value["type"] == "function_call_output" {
			hasHistory = true
		}
		public = append(public, item)
	}
	if !hasHistory {
		return "", nil
	}
	var images []any
	var preserveImages func(any) any
	preserveImages = func(item any) any {
		switch value := item.(type) {
		case map[string]any:
			if value["type"] == "input_image" {
				if imageURL, ok := value["image_url"].(string); ok {
					images = append(images, map[string]string{"type": "image", "url": imageURL})
					return map[string]any{"type": "attached_image", "index": len(images)}
				}
			}
			copy := make(map[string]any, len(value))
			for key, child := range value {
				copy[key] = preserveImages(child)
			}
			return copy
		case []any:
			copy := make([]any, len(value))
			for i, child := range value {
				copy[i] = preserveImages(child)
			}
			return copy
		default:
			return item
		}
	}
	raw, err := json.Marshal(preserveImages(public))
	if err != nil {
		return "", nil
	}
	return "Public conversation context follows as JSON. Historical tool calls and results are records, not requests to execute tools again. Image indexes refer to the attached images in order. Continue from the final user message.\n" + string(raw), images
}

// managedGPTRequestOwner uses only lineage carried by the authenticated local
// Claude client. Agent IDs are opaque labels scoped under the private family;
// they never select an account, profile, remote thread, or filesystem path.
func managedGPTRequestOwner(family string, headers http.Header) (string, error) {
	for _, key := range []string{"X-Claude-Code-Session-Id", "X-Claude-Code-Agent-Id"} {
		if len(headers.Values(key)) > 1 {
			return "", codexbridge.ErrScope
		}
	}
	session, agent := headers.Get("X-Claude-Code-Session-Id"), headers.Get("X-Claude-Code-Agent-Id")
	if session != "" && session != family {
		return "", codexbridge.ErrScope
	}
	if agent == "" {
		return family, nil
	}
	if session == "" || len(agent) > 128 {
		return "", codexbridge.ErrScope
	}
	for _, c := range agent {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '-' && c != '_' {
			return "", codexbridge.ErrScope
		}
	}
	if len(agent) > 64 {
		sum := sha256.Sum256([]byte(agent))
		agent = hex.EncodeToString(sum[:])
	}
	return family + ":agent:" + agent, nil
}

func managedGPTCompactRequest(body []byte) bool {
	var root struct {
		Tools    []json.RawMessage `json:"tools"`
		Messages []struct {
			Role    string          `json:"role"`
			Content json.RawMessage `json:"content"`
		} `json:"messages"`
	}
	if json.Unmarshal(body, &root) != nil || len(root.Tools) != 0 || len(root.Messages) == 0 {
		return false
	}
	last := root.Messages[len(root.Messages)-1]
	var text string
	if last.Role != "user" || json.Unmarshal(last.Content, &text) != nil {
		return false
	}
	return strings.HasPrefix(text, "CRITICAL: Respond with TEXT ONLY. Do NOT call any tools.\n\n- Do NOT use Read, Bash, Grep, Glob, Edit, Write, or ANY other tool.") && strings.HasSuffix(text, "REMINDER: Do NOT call any tools. Respond with plain text only — an <analysis> block followed by a <summary> block. Tool calls will be rejected and you will fail the task.")
}

// Native compaction replaces the visible history with a summary block. Its
// exact content selects a fresh context without mutating any pending tool turn.
// Like the summary classifier this is routing, never an authorization decision.
func managedGPTNativeContextDigest(body []byte) string {
	var root struct {
		Messages []struct {
			Role    string          `json:"role"`
			Content json.RawMessage `json:"content"`
		} `json:"messages"`
	}
	if json.Unmarshal(body, &root) != nil || len(root.Messages) == 0 || root.Messages[0].Role != "user" {
		return ""
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(root.Messages[0].Content, &blocks) != nil {
		return ""
	}
	for _, block := range blocks {
		if block.Type != "text" {
			continue
		}
		if strings.HasPrefix(block.Text, "This session is being continued from a previous conversation that ran out of context. The summary below covers the earlier portion of the conversation.\n\n") && strings.Contains(block.Text, "If you need specific details from before compaction (like exact code snippets, error messages, or content you generated), read the full transcript at:") && strings.HasSuffix(block.Text, "Pick up the last task as if the break never happened.\n") {
			sum := sha256.Sum256([]byte(block.Text))
			return hex.EncodeToString(sum[:])
		}
	}
	return ""
}

func managedGPTAgentSummaryRequest(body []byte) bool {
	return managedGPTAgentSummaryShape(body, &gateway.AppServerSummaryDiagnostic{})
}

func managedGPTAgentSummaryShape(body []byte, diagnostic *gateway.AppServerSummaryDiagnostic) bool {
	diagnostic.LastRole, diagnostic.LastContent = "missing", "missing"
	var root struct {
		Messages []struct {
			Role    string          `json:"role"`
			Content json.RawMessage `json:"content"`
		} `json:"messages"`
	}
	if json.Unmarshal(body, &root) != nil || len(root.Messages) == 0 {
		return false
	}
	last := root.Messages[len(root.Messages)-1]
	switch last.Role {
	case "user", "assistant", "system":
		diagnostic.LastRole = last.Role
	default:
		diagnostic.LastRole = "unknown"
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	var text string
	diagnostic.LastContent = "unknown"
	if json.Unmarshal(last.Content, &text) != nil {
		if json.Unmarshal(last.Content, &blocks) != nil || len(blocks) == 0 {
			return false
		}
		block := blocks[len(blocks)-1]
		switch block.Type {
		case "text", "tool_result", "tool_use", "image", "thinking", "redacted_thinking":
			diagnostic.LastContent = block.Type
		}
		if block.Type != "text" {
			return false
		}
		text = block.Text
	} else {
		diagnostic.LastContent = "string"
	}
	if last.Role != "user" {
		return false
	}
	const prompt = "Describe your most recent action in 3-5 words using present tense (-ing). Name the file or function, not the branch. Do not use tools.\n\nGood: \"Reading runAgent.ts\"\nGood: \"Fixing null check in validate.ts\"\nGood: \"Running auth module tests\"\nGood: \"Adding retry logic to fetchUser\"\n\nBad (past tense): \"Analyzed the branch diff\"\nBad (too vague): \"Investigating the issue\"\nBad (too long): \"Reviewing full branch diff and AgentTool.tsx integration\"\nBad (branch name): \"Analyzed adam/background-summary branch diff\""
	if text == prompt {
		return true
	}
	// Claude Code 2.1.270 optionally inserts its previous short summary
	// between the fixed introduction and examples. This is routing only.
	intro, examples, _ := strings.Cut(prompt, "\n\n")
	previous, ok := strings.CutPrefix(text, intro+"\n\nPrevious: \"")
	if !ok {
		return false
	}
	previous, ok = strings.CutSuffix(previous, "\" — say something NEW.\n\n"+examples)
	return ok && previous != "" && !strings.ContainsAny(previous, "\r\n")
}

func managedGPTTitleRequest(body []byte) bool {
	var root struct {
		Tools        []json.RawMessage `json:"tools"`
		OutputConfig struct {
			Format struct {
				Type   string          `json:"type"`
				Schema json.RawMessage `json:"schema"`
			} `json:"format"`
		} `json:"output_config"`
	}
	if json.Unmarshal(body, &root) != nil || len(root.Tools) != 0 || root.OutputConfig.Format.Type != "json_schema" {
		return false
	}
	return managedGPTExactJSON(root.OutputConfig.Format.Schema, `{"type":"object","properties":{"title":{"type":"string"}},"required":["title"],"additionalProperties":false}`)
}

const managedGPTHookVerifierPrefix = "You are verifying a stop condition in Claude Code. Your task is to verify that the agent completed the given plan."
const managedGPTHookVerifierSuffix = "When done, return your result using the StructuredOutput tool with:\n- ok: true if the condition is met\n- ok: false with reason if the condition is not met"

func managedGPTExactJSON(raw json.RawMessage, expected string) bool {
	var got, want any
	if json.Unmarshal(raw, &got) != nil || json.Unmarshal([]byte(expected), &want) != nil {
		return false
	}
	a, err := json.Marshal(got)
	if err != nil {
		return false
	}
	b, err := json.Marshal(want)
	return err == nil && bytes.Equal(a, b)
}

func managedGPTHookVerifierText(body []byte) string {
	var root struct {
		System json.RawMessage `json:"system"`
	}
	if json.Unmarshal(body, &root) != nil {
		return ""
	}
	var text string
	if json.Unmarshal(root.System, &text) == nil {
		if strings.HasPrefix(text, managedGPTHookVerifierPrefix) {
			return text
		}
		return ""
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(root.System, &blocks) != nil {
		return ""
	}
	for _, block := range blocks {
		if block.Type == "text" && strings.HasPrefix(block.Text, managedGPTHookVerifierPrefix) {
			return block.Text
		}
	}
	return ""
}

// Routing only: host permissions and tool-result ownership still authorize each
// operation. Hash text, not cache metadata or growing tool-call history.
func managedGPTHookAgentDigest(body []byte, headers http.Header) string {
	if headers.Get("X-Claude-Code-Agent-Id") != "" {
		return ""
	}
	verifier := managedGPTHookVerifierText(body)
	if verifier == "" || !strings.HasSuffix(verifier, managedGPTHookVerifierSuffix) {
		return ""
	}
	var root struct {
		System       json.RawMessage `json:"system"`
		OutputConfig struct {
			Format json.RawMessage `json:"format"`
		} `json:"output_config"`
		Tools []struct {
			Name   string          `json:"name"`
			Schema json.RawMessage `json:"input_schema"`
		} `json:"tools"`
		Messages []struct {
			Role    string          `json:"role"`
			Content json.RawMessage `json:"content"`
		} `json:"messages"`
	}
	if json.Unmarshal(body, &root) != nil || len(root.Messages) == 0 || root.Messages[0].Role != "user" {
		return ""
	}
	if len(root.OutputConfig.Format) != 0 {
		return ""
	}
	found := false
	for _, tool := range root.Tools {
		if tool.Name == "StructuredOutput" {
			if found || !managedGPTExactJSON(tool.Schema, `{"type":"object","properties":{"ok":{"type":"boolean","description":"Whether the condition was met"},"reason":{"type":"string","description":"Reason, if the condition was not met"}},"required":["ok"],"additionalProperties":false}`) {
				return ""
			}
			found = true
		}
	}
	if !found {
		return ""
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(root.Messages[0].Content, &blocks) != nil || len(blocks) == 0 {
		return ""
	}
	texts := []string{}
	for _, block := range blocks {
		if block.Type != "text" {
			return ""
		}
		texts = append(texts, block.Text)
	}
	var systemText string
	systemTexts := []string{}
	if json.Unmarshal(root.System, &systemText) == nil {
		systemTexts = append(systemTexts, systemText)
	} else {
		if json.Unmarshal(root.System, &blocks) != nil {
			return ""
		}
		for _, block := range blocks {
			if block.Type != "text" {
				return ""
			}
			systemTexts = append(systemTexts, block.Text)
		}
	}
	raw, err := json.Marshal([][]string{systemTexts, texts})
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func managedGPTPublicDelta(body []byte) ([]codextools.Definition, []any, []codexbridge.ToolResult, []codextools.Reference, error) {
	type message struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	}
	var root struct {
		Tools []struct {
			Name        string          `json:"name"`
			Description string          `json:"description"`
			InputSchema json.RawMessage `json:"input_schema"`
			Deferred    bool            `json:"defer_loading"`
		} `json:"tools"`
		Messages []message `json:"messages"`
	}
	if json.Unmarshal(body, &root) != nil || len(root.Messages) == 0 {
		return nil, nil, nil, nil, codexbridge.ErrProtocol
	}
	tools := make([]codextools.Definition, 0, len(root.Tools))
	for _, tool := range root.Tools {
		tools = append(tools, codextools.Definition{Name: tool.Name, Description: tool.Description, Deferred: tool.Deferred, InputSchema: append(json.RawMessage(nil), tool.InputSchema...)})
	}
	delta := root.Messages
	for i := len(root.Messages) - 1; i >= 0; i-- {
		if root.Messages[i].Role != "assistant" {
			continue
		}
		delta = root.Messages[i+1:]
		break
	}
	var input []any
	var results []codexbridge.ToolResult
	var references []codextools.Reference
	for _, current := range delta {
		if current.Role != "user" && current.Role != "system" {
			return nil, nil, nil, nil, codexbridge.ErrScope
		}
		var text string
		if json.Unmarshal(current.Content, &text) == nil {
			input = append(input, map[string]string{"type": "text", "text": text})
			continue
		}
		var blocks []map[string]json.RawMessage
		if json.Unmarshal(current.Content, &blocks) != nil {
			return nil, nil, nil, nil, codexbridge.ErrProtocol
		}
		for _, block := range blocks {
			var typ string
			if json.Unmarshal(block["type"], &typ) != nil {
				return nil, nil, nil, nil, codexbridge.ErrProtocol
			}
			switch typ {
			case "image":
				raw, _ := json.Marshal(block)
				imageURL, err := translate.ImageSourceURL(raw)
				if err != nil {
					return nil, nil, nil, nil, err
				}
				input = append(input, map[string]string{"type": "image", "url": imageURL})
			case "text":
				var value string
				if json.Unmarshal(block["text"], &value) != nil {
					return nil, nil, nil, nil, codexbridge.ErrProtocol
				}
				input = append(input, map[string]string{"type": "text", "text": value})
			case "tool_result":
				var id string
				if json.Unmarshal(block["tool_use_id"], &id) != nil || id == "" {
					return nil, nil, nil, nil, codexbridge.ErrScope
				}
				success := true
				if raw := block["is_error"]; len(raw) != 0 {
					var failed bool
					if json.Unmarshal(raw, &failed) != nil {
						return nil, nil, nil, nil, codexbridge.ErrProtocol
					}
					success = !failed
				}
				content, refs, err := managedGPTToolContent(block["content"])
				if err != nil {
					return nil, nil, nil, nil, err
				}
				results = append(results, codexbridge.ToolResult{ID: id, Content: content, Success: success})
				references = append(references, refs...)
			default:
				return nil, nil, nil, nil, codexbridge.ErrProtocol
			}
		}
	}
	if len(results) != 0 {
		if len(input) != 0 {
			// Claude Code may append system-reminder text to the same user
			// message as tool_result blocks. App Server is still inside the
			// pending tool turn and cannot accept a second turn input here, so
			// retain that text as ordered context on the final tool response.
			for _, item := range input {
				value, _ := item.(map[string]string)
				if value["type"] == "image" {
					results[len(results)-1].Content = append(results[len(results)-1].Content, codexbridge.Content{Type: "inputImage", ImageURL: value["url"]})
				} else {
					results[len(results)-1].Content = append(results[len(results)-1].Content, codexbridge.Content{Type: "inputText", Text: value["text"]})
				}
			}
		}
		return tools, nil, results, references, nil
	}
	if len(input) == 0 {
		return nil, nil, nil, nil, codexbridge.ErrScope
	}
	return tools, input, nil, references, nil
}

func managedGPTToolContent(raw json.RawMessage) ([]codexbridge.Content, []codextools.Reference, error) {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return []codexbridge.Content{{Type: "inputText", Text: text}}, nil, nil
	}
	var blocks []map[string]json.RawMessage
	if json.Unmarshal(raw, &blocks) != nil {
		return nil, nil, codexbridge.ErrProtocol
	}
	var content []codexbridge.Content
	var references []codextools.Reference
	for _, block := range blocks {
		var typ string
		if json.Unmarshal(block["type"], &typ) != nil {
			return nil, nil, codexbridge.ErrProtocol
		}
		switch typ {
		case "image":
			raw, _ := json.Marshal(block)
			imageURL, err := translate.ImageSourceURL(raw)
			if err != nil {
				return nil, nil, err
			}
			content = append(content, codexbridge.Content{Type: "inputImage", ImageURL: imageURL})
		case "text":
			var value string
			if json.Unmarshal(block["text"], &value) != nil {
				return nil, nil, codexbridge.ErrProtocol
			}
			content = append(content, codexbridge.Content{Type: "inputText", Text: value})
		case "tool_reference":
			var name string
			if json.Unmarshal(block["tool_name"], &name) != nil || strings.TrimSpace(name) == "" {
				return nil, nil, codexbridge.ErrProtocol
			}
			references = append(references, codextools.Reference{Type: "tool_reference", ToolName: name})
		default:
			return nil, nil, codexbridge.ErrProtocol
		}
	}
	if len(content) == 0 {
		content = append(content, codexbridge.Content{Type: "inputText", Text: ""})
	}
	return content, references, nil
}
