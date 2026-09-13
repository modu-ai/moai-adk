package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/modu-ai/moai-adk/internal/gateway"
	"github.com/modu-ai/moai-adk/internal/gateway/auth"
	"github.com/modu-ai/moai-adk/internal/gateway/receipt"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
)

// Dependencies are caller-verified policy, not assertions supplied by stdin.
// A production factory must not be registered until these observations exist.
type gatewayFactoryDependencies struct {
	Models           []gateway.ModelEntry
	Transport        *http.Transport
	Limits           translate.Limits
	MeasureInput     func(gateway.ModelEntry, []byte) (int64, error)
	AnthropicVersion string
	AllowedBetas     []string
	Broker           auth.Broker
	RefreshVerifier  func(context.Context, auth.CredentialRef) error
	OpenStore        func() (*auth.Store, error)
	SendOptions      auth.SendOptions
}

type gatewayPrivatePayload struct {
	Version      int                         `json:"version"`
	SessionToken string                      `json:"session_token"`
	ModelIDs     []string                    `json:"model_ids"`
	Conversation *gatewayPrivateConversation `json:"conversation,omitempty"`
}

type gatewayPrivateConversation struct {
	FamilyID   string `json:"family_id"`
	SessionID  string `json:"session_id"`
	ReceiptDir string `json:"receipt_dir"`
}

var errGatewayFactory = errors.New("gateway private configuration or verified dependencies unavailable")

func validGatewayConversation(c gatewayPrivateConversation) bool {
	if _, err := receipt.New(c.FamilyID); err != nil {
		return false
	}
	if _, err := receipt.New(c.SessionID); err != nil {
		return false
	}
	return filepath.IsAbs(c.ReceiptDir) && filepath.Clean(c.ReceiptDir) == c.ReceiptDir && !strings.ContainsAny(c.ReceiptDir, "\x00\r\n")
}

func authorizeGatewayNativeReceipt(ctx context.Context, _ []byte, policy translate.NativePolicy, c gatewayPrivateConversation, store *receipt.Store) error {
	if store == nil || ctx == nil {
		return errGatewayFactory
	}
	if policy.UserID != "" {
		var identity struct {
			DeviceID    string `json:"device_id"`
			AccountUUID string `json:"account_uuid"`
			SessionID   string `json:"session_id"`
		}
		if json.Unmarshal([]byte(policy.UserID), &identity) != nil || identity.SessionID != c.SessionID || identity.DeviceID == "" {
			return errGatewayFactory
		}
	}
	manifest, err := store.Snapshot(ctx)
	if err != nil {
		return err
	}
	return manifest.Check(c.SessionID, nil)
}

func newGatewayHandlerFactory(d gatewayFactoryDependencies) (gatewayHandlerFactory, error) {
	if len(d.Models) == 0 || len(d.Models) > 64 || d.MeasureInput == nil || d.Transport == nil || d.Transport.Proxy != nil || d.Transport.TLSClientConfig != nil && d.Transport.TLSClientConfig.InsecureSkipVerify || d.Limits.MaxBodyBytes <= 0 || d.Limits.MaxEventBytes <= 0 || d.Limits.MaxOutputBytes <= 0 {
		return nil, errGatewayFactory
	}
	d.Models = append([]gateway.ModelEntry(nil), d.Models...)
	d.AllowedBetas = append([]string(nil), d.AllowedBetas...)
	d.Transport = d.Transport.Clone()
	// Limits describes budgets only; per-request measurements must not be shared.
	d.Limits.InputTokens = nil
	d.Limits.ContextTokens = 0
	var tiers []string
	for _, e := range d.Models {
		if e.Provider == gateway.ProviderZAI {
			tiers = append(tiers, e.RouteID)
		}
	}
	canonical, err := gateway.NewSessionCatalog(tiers)
	if err != nil {
		return nil, errGatewayFactory
	}
	approved := make(map[string]gateway.ModelEntry, len(d.Models))
	for _, e := range d.Models {
		expected, err := canonical.Resolve(e.RouteID)
		if err != nil || e.Provider != expected.Provider || e.UpstreamID != expected.UpstreamID || e.AuthMethod != expected.AuthMethod || e.Capabilities.ContextTokens <= 0 {
			return nil, errGatewayFactory
		}
		if _, exists := approved[e.RouteID]; exists {
			return nil, errGatewayFactory
		}
		approved[e.RouteID] = e
		if e.Provider == gateway.ProviderOpenAI && (d.Broker == nil || d.RefreshVerifier == nil) {
			return nil, errGatewayFactory
		}
		if e.Provider != gateway.ProviderOpenAI && (d.AnthropicVersion == "" || strings.ContainsAny(d.AnthropicVersion, "\r\n")) {
			return nil, errGatewayFactory
		}
	}
	if d.OpenStore == nil {
		d.OpenStore = openGPTAuthStore
	}
	return func(raw json.RawMessage) (http.Handler, error) {
		if len(raw) > 65536 || gateway.ValidateJSONObject(raw) != nil {
			return nil, errGatewayFactory
		}
		// encoding/json matches struct fields case-insensitively. Validate exact
		// private protocol keys before decoding; ValidateJSONObject rejects duplicates.
		var fields map[string]json.RawMessage
		if json.Unmarshal(raw, &fields) != nil {
			return nil, errGatewayFactory
		}
		for key := range fields {
			switch key {
			case "version", "session_token", "model_ids", "conversation":
			default:
				return nil, errGatewayFactory
			}
		}
		var p gatewayPrivatePayload
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		if dec.Decode(&p) != nil || p.Version != 1 || p.SessionToken == "" || len(p.SessionToken) > 4096 || len(p.ModelIDs) == 0 || len(p.ModelIDs) > len(approved) {
			return nil, errGatewayFactory
		}
		if p.Conversation != nil && !validGatewayConversation(*p.Conversation) {
			return nil, errGatewayFactory
		}
		for _, c := range p.SessionToken {
			if c < 33 || c > 126 {
				return nil, errGatewayFactory
			}
		}
		var rows []gateway.ModelEntry
		providers := map[gateway.ProviderID]bool{}
		seen := map[string]bool{}
		for _, id := range p.ModelIDs {
			e, ok := approved[id]
			if !ok || seen[id] {
				return nil, errGatewayFactory
			}
			seen[id] = true
			rows = append(rows, e)
			providers[e.Provider] = true
		}
		catalog, err := gateway.NewCatalog(rows)
		if err != nil {
			return nil, errGatewayFactory
		}
		var store *auth.Store
		var receipts *receipt.Store
		if providers[gateway.ProviderOpenAI] {
			store, err = d.OpenStore()
			if err != nil || store == nil {
				if store != nil {
					_ = store.Close()
				}
				return nil, errGatewayFactory
			}
			if p.Conversation != nil {
				receipts, err = receipt.OpenStore(context.Background(), p.Conversation.ReceiptDir, p.Conversation.SessionID, false)
				if err != nil || receipts == nil {
					if receipts != nil {
						_ = receipts.Close()
					}
					_ = store.Close()
					return nil, errGatewayFactory
				}
			}
		}
		success := false
		defer func() {
			if !success {
				if receipts != nil {
					_ = receipts.Close()
				}
				if store != nil {
					_ = store.Close()
				}
			}
		}()
		limits := d.Limits
		if receipts != nil && p.Conversation != nil {
			limits.History = translate.NewGPTSubscriptionReceiptHistory(receipts, p.Conversation.SessionID, p.Conversation.FamilyID)
			limits.NativeReceiptAuthorize = func(ctx context.Context, body []byte, policy translate.NativePolicy) error {
				return authorizeGatewayNativeReceipt(ctx, body, policy, *p.Conversation, receipts)
			}
		}
		adapters := make(map[gateway.ProviderID]gateway.Adapter)
		for provider := range providers {
			config := gateway.MessagesConfig{Transport: d.Transport, AnthropicVersion: d.AnthropicVersion, AllowedBetas: d.AllowedBetas, Limits: limits, MeasureInput: d.MeasureInput}
			switch provider {
			case gateway.ProviderAnthropic:
				adapters[provider], err = gateway.NewAnthropicOAuthAdapter(config)
			case gateway.ProviderZAI:
				adapters[provider], err = gateway.NewGLMAdapter(config)
			case gateway.ProviderOpenAI:
				adapters[provider], err = gateway.NewOpenAIAdapter(gateway.OpenAIConfig{Subscription: store, Transport: d.Transport, Limits: limits, MeasureInput: d.MeasureInput, SendOptions: d.SendOptions})
			}
			if err != nil {
				return nil, errGatewayFactory
			}
		}
		resolver := func(ctx context.Context, e gateway.ModelEntry) (gateway.CredentialRef, error) {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			switch e.Provider {
			case gateway.ProviderOpenAI:
				if store != nil {
					return store.ResolveFresh(ctx, d.Broker, d.RefreshVerifier)
				}
			case gateway.ProviderZAI:
				return auth.NewGLMCredential()
			}
			return nil, auth.ErrCredentialAbsent
		}
		sessionHeader := "X-Api-Key"
		sessionToken := p.SessionToken
		for _, entry := range rows {
			if entry.Provider == gateway.ProviderOpenAI {
				sessionHeader = "Authorization"
				sessionToken = "Bearer " + p.SessionToken
				break
			}
			if entry.AuthMethod == gateway.AuthOAuthPassthrough {
				sessionHeader = "X-MoAI-Session-Token"
				break
			}
		}
		server, err := gateway.NewServer(gateway.ServerConfig{SessionHeader: sessionHeader, SessionToken: sessionToken, MaxBodyBytes: int64(d.Limits.MaxBodyBytes), Catalog: catalog, ResolveCredential: resolver, Adapters: adapters})
		if err != nil {
			return nil, errGatewayFactory
		}
		ctx, cancel := context.WithCancel(context.Background())
		success = true
		return &gatewayOwnedHandler{Handler: server, store: store, receipts: receipts, ctx: ctx, cancel: cancel}, nil
	}, nil
}

type gatewayOwnedHandler struct {
	http.Handler
	store    *auth.Store
	receipts *receipt.Store
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
	closed   bool
	requests sync.WaitGroup
	once     sync.Once
	closeErr error
}

func (h *gatewayOwnedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		http.Error(w, "gateway closed", http.StatusServiceUnavailable)
		return
	}
	h.requests.Add(1)
	h.mu.Unlock()
	defer h.requests.Done()
	ctx, cancel := context.WithCancel(r.Context())
	stop := context.AfterFunc(h.ctx, cancel)
	defer stop()
	defer cancel()
	h.Handler.ServeHTTP(w, r.WithContext(ctx))
}
func (h *gatewayOwnedHandler) Close() error {
	h.once.Do(func() {
		h.mu.Lock()
		h.closed = true
		h.cancel()
		h.mu.Unlock()
		h.requests.Wait()
		if h.store != nil {
			h.closeErr = h.store.Close()
		}
		if h.receipts != nil {
			if err := h.receipts.Close(); h.closeErr == nil {
				h.closeErr = err
			}
		}
	})
	return h.closeErr
}
