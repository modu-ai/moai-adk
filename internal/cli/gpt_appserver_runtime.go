package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
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
	client, err := codexapp.Start(lifetime, codexapp.Config{Binary: binary, Home: profile})
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
			return authorizeGatewayNativeReceipt(ctx, body, policy, *p.Conversation, receipts)
		}
	}
	prepare := newManagedGPTPrepare(conversationID, cwd, limits, store)
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
	server, err := gateway.NewServer(gateway.ServerConfig{
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
	for _, row := range gatewayGPTModels() {
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

func newManagedGPTPrepare(conversationID, cwd string, base translate.Limits, store *codexbridge.FileStore) func(context.Context, gateway.RoutedRequest) (codexbridge.Request, error) {
	return func(ctx context.Context, r gateway.RoutedRequest) (codexbridge.Request, error) {
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
			Reasoning    struct {
				Effort string `json:"effort"`
			} `json:"reasoning"`
		}
		if json.Unmarshal(projected, &upstream) != nil {
			return codexbridge.Request{}, codexbridge.ErrProtocol
		}
		tools, input, results, references, err := managedGPTPublicDelta(r.Body)
		if err != nil {
			return codexbridge.Request{}, err
		}
		sum := sha256.Sum256(projected)
		current := hex.EncodeToString(sum[:])
		ownerID := conversationID
		if managedGPTTitleRequest(r.Body) {
			// Claude Code generates the session title through a separate model
			// request with title-only developer instructions. Reusing that App
			// Server thread for the coding session permanently pins the title
			// instructions onto the main conversation.
			ownerID += ":title"
		}
		q := codexbridge.Request{
			Owner:        codextools.Binding{ConversationID: ownerID, AccountScope: r.Managed.Scope()},
			Model:        r.Entry.UpstreamID,
			Effort:       upstream.Reasoning.Effort,
			CWD:          cwd,
			Instructions: upstream.Instructions,
			Tools:        tools,
			References:   references,
			Input:        input,
			Results:      results,
			PrefixDigest: current,
		}
		barrier, found, err := store.Barrier(q.Owner)
		if err != nil {
			return codexbridge.Request{}, err
		}
		if found {
			if barrier.Model != q.Model || barrier.CWD != q.CWD {
				return codexbridge.Request{}, codexbridge.ErrScope
			}
			q.ExpectedPrefix = barrier.Prefix
			q.Resume = barrier.Phase == "idle"
		}
		return q, nil
	}
}

func managedGPTTitleRequest(body []byte) bool {
	var root struct {
		Tools        []json.RawMessage `json:"tools"`
		OutputConfig struct {
			Format struct {
				Type   string `json:"type"`
				Schema struct {
					Required []string `json:"required"`
				} `json:"schema"`
			} `json:"format"`
		} `json:"output_config"`
	}
	if json.Unmarshal(body, &root) != nil || len(root.Tools) != 0 || root.OutputConfig.Format.Type != "json_schema" {
		return false
	}
	return len(root.OutputConfig.Format.Schema.Required) == 1 && root.OutputConfig.Format.Schema.Required[0] == "title"
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
				results[len(results)-1].Content = append(results[len(results)-1].Content, codexbridge.Content{Type: "inputText", Text: value["text"]})
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
