package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/gateway"
	"github.com/modu-ai/moai-adk/internal/gateway/auth"
	"github.com/modu-ai/moai-adk/internal/gateway/conversation"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// Supported four-model default, verified against App Server 0.154.0 metadata:
// context_window=272000, effective_context_window_percent=95 (258400 native).
// max_context_window=872000 is opt-in capacity, not the active thread default.
// Claude reserves its own compact headroom from this declared default window.
const gatewayContextWindow = 272000

// A parser safety ceiling, not a model capability declaration. Larger metadata
// requires an explicit compatibility review before changing Claude's global limit.
const gatewayContextWindowLimit = 2000000
const gatewayRequestBodyLimit = 16 << 20

// gatewayOutputPolicyDisplay names the output policy every launch assembly
// delivers: both the subscription and the explicit API path use the official
// App Server output policy (operator decision, 0.11.0). The display never
// claims Claude generation-token ceilings carry over — byte and cancellation
// limits stay separate budgets.
const gatewayOutputPolicyDisplay = "app-server"

// gatewayAuthDisplay names the authentication surface a launch actually uses,
// derived from the resolved catalog entry's declared auth method — never from
// request-time state — so the displayed method cannot drift from the session
// environment the child receives (AC-MG-026 (a) assembly combination,
// AS-014/AS-021 display consistency).
func gatewayAuthDisplay(method gateway.AuthMethod) string {
	switch method {
	case gateway.AuthPKCE, gateway.AuthOAuthPassthrough:
		return "subscription"
	case gateway.AuthAppServer:
		return "app-server-managed"
	case gateway.AuthAPIKey:
		return "api-key"
	default:
		return "existing-credential"
	}
}

// newGPTGatewayBinding is the production launch seam. It creates only the
// private child configuration; credentials remain in the child-owned store and
// are resolved by the request adapter at send time.
func newGPTGatewayBinding() (*gatewayLaunchBinding, error) {
	return newGPTGatewayBindingWithStart(nil)
}

func newGPTGatewayBindingWithStart(start func(context.Context, gateway.StartOptions) (gatewayStartedChild, error)) (*gatewayLaunchBinding, error) {
	models, err := installedGatewayGPTModels()
	if err != nil {
		return nil, err
	}
	return newProviderGatewayBinding("gpt", models, config.GLMModels{}, "", start)
}

func newNativeGatewayBinding(mode string) (*gatewayLaunchBinding, error) {
	var models config.GLMModels
	key := ""
	if mode == "glm" {
		root, err := findProjectRootFn()
		if err != nil {
			return nil, err
		}
		cfg, err := loadGLMConfig(root)
		if err != nil {
			return nil, err
		}
		models = config.GLMModels{High: cfg.Models.High, Medium: cfg.Models.Medium, Low: cfg.Models.Low, Fable: cfg.Models.Fable}
		key = loadGLMKey()
		if key == "" {
			return nil, errors.New("GLM credential unavailable; run moai glm setup")
		}
	}
	return newNativeGatewayBindingWithStart(mode, models, key, nil)
}
func newNativeGatewayBindingWithStart(mode string, tiers config.GLMModels, key string, start func(context.Context, gateway.StartOptions) (gatewayStartedChild, error)) (*gatewayLaunchBinding, error) {
	models, err := gatewayNativeModels(mode, tiers)
	if err != nil {
		return nil, err
	}
	return newProviderGatewayBinding(mode, models, tiers, key, start)
}
func gatewayNativeModels(mode string, tiers config.GLMModels) ([]gateway.ModelEntry, error) {
	var ids []string
	provider := gateway.ProviderAnthropic
	method := gateway.AuthOAuthPassthrough
	switch mode {
	case "claude":
		ids = []string{"claude-opus-5", "claude-sonnet-5"}
	case "glm":
		ids = []string{tiers.High, tiers.Medium, tiers.Low, tiers.Fable}
		provider = gateway.ProviderZAI
		method = gateway.AuthExistingGLM
	default:
		return nil, errors.New("unsupported native gateway provider")
	}
	// AS-020 capability re-verdict: the declarations are nominal metadata from
	// the official model tables, never an acceptance guarantee. Claude routes
	// declare image acceptance at the 1M nominal context; GLM routes are
	// text-only (image input is refused explicitly at the request boundary)
	// at the documented 200K nominal context.
	images := true
	nominalContext := 1000000
	if provider == gateway.ProviderZAI {
		images = false
		nominalContext = 200000
	}
	rows := []gateway.ModelEntry{}
	seen := map[string]bool{}
	for _, id := range ids {
		if id == "" {
			return nil, errors.New("provider model ID unavailable")
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		rows = append(rows, gateway.ModelEntry{RouteID: id, UpstreamID: id, Provider: provider, AuthMethod: method, Capabilities: gateway.Capabilities{ContextTokens: nominalContext, Images: images, Tools: true, Streaming: true}})
		if provider == gateway.ProviderAnthropic {
			qualified := rows[len(rows)-1]
			qualified.RouteID = id + "[1m]"
			rows = append(rows, qualified)
		}
	}
	return rows, nil
}
func newProviderGatewayBinding(mode string, models []gateway.ModelEntry, tiers config.GLMModels, key string, start func(context.Context, gateway.StartOptions) (gatewayStartedChild, error)) (*gatewayLaunchBinding, error) {
	catalog, err := gateway.NewCatalog(models)
	if err != nil {
		return nil, errors.New("GPT gateway catalog unavailable")
	}
	token, err := newGatewaySessionToken()
	if err != nil {
		return nil, errors.New("GPT gateway session unavailable")
	}
	payload, err := marshalGatewayPrivatePayload(token, models)
	if err != nil {
		return nil, errors.New("GPT gateway configuration unavailable")
	}
	executable, err := os.Executable()
	if err != nil || executable == "" {
		return nil, errors.New("GPT gateway executable unavailable")
	}
	home, err := paths.MoaiHome()
	if err != nil || home == "" {
		return nil, errors.New("GPT conversation state unavailable")
	}
	namespace := "gateway-conversations"
	if mode != "gpt" {
		namespace += "-" + mode
	}
	families, err := conversation.Open(filepath.Join(home, "state", namespace))
	if err != nil {
		return nil, errors.New("GPT conversation state unavailable")
	}
	sessionEnv := []string{"ANTHROPIC_API_KEY=" + token}
	if mode == "gpt" {
		sessionEnv = []string{"ANTHROPIC_AUTH_TOKEN=" + token}
	}
	if mode == "claude" {
		sessionEnv = []string{"ANTHROPIC_CUSTOM_HEADERS=X-MoAI-Session-Token: " + token}
	}
	return newGatewaySessionBinding(gatewaySessionOptions{
		Mode: mode,
		GLM:  tiers, GLMKey: key,
		Catalog: catalog,
		Family:  families,
		Payload: func(descriptor conversation.Descriptor) (json.RawMessage, error) {
			return marshalGatewayPrivatePayloadWithConversation(token, models, descriptor)
		},
		Start: start,
		Child: gateway.StartOptions{
			Executable:     executable,
			Args:           []string{"internal-gateway"},
			Env:            gatewayChildEnvironment(os.Environ()),
			StartupTimeout: 15 * time.Second,
			Config: gateway.ChildConfig{
				Lifetime:     24 * time.Hour,
				PollInterval: 100 * time.Millisecond,
				Payload:      payload,
			},
		},
		SessionEnv: sessionEnv,
	}), nil
}

// productionGatewayHandlerFactory is installed only in the private child.
// All four GPT routes are explicit and carry measured context capability.
func productionGatewayHandlerFactory(raw json.RawMessage) (http.Handler, error) {
	var selection gatewayPrivatePayload
	if json.Unmarshal(raw, &selection) != nil || len(selection.ModelIDs) == 0 {
		return nil, errGatewayFactory
	}
	var models []gateway.ModelEntry
	var broker auth.Broker
	var policy translate.PolicyProfile
	first := selection.ModelIDs[0]
	if strings.HasPrefix(first, "gpt-") {
		return newGPTAppServerGatewayHandler(raw)
	} else if strings.HasPrefix(first, "claude-") {
		models, _ = gatewayNativeModels("claude", config.GLMModels{})
	} else {
		root, err := findProjectRootFn()
		if err != nil {
			return nil, errGatewayFactory
		}
		cfg, err := loadGLMConfig(root)
		if err != nil {
			return nil, errGatewayFactory
		}
		models, err = gatewayNativeModels("glm", config.GLMModels{High: cfg.Models.High, Medium: cfg.Models.Medium, Low: cfg.Models.Low, Fable: cfg.Models.Fable})
		if err != nil {
			return nil, errGatewayFactory
		}
	}
	factory, err := newGatewayHandlerFactory(gatewayFactoryDependencies{
		Models:           models,
		Transport:        &http.Transport{},
		Limits:           translate.Limits{PolicyProfile: policy, MaxBodyBytes: gatewayRequestBodyLimit, MaxEventBytes: 1 << 20, MaxOutputBytes: 8 << 20},
		MeasureInput:     gateway.EstimateInputTokens,
		AnthropicVersion: "2023-06-01",
		Broker:           broker,
		RefreshVerifier:  verifyGPTCredentialOwnership,
		SendOptions:      auth.SendOptions{WriteTimeout: 30 * time.Second, PollInterval: 100 * time.Millisecond, MaxBodyBytes: gatewayRequestBodyLimit},
	})
	if err != nil {
		return nil, err
	}
	return factory(raw)
}

func verifyGPTCredentialOwnership(ctx context.Context, ref auth.CredentialRef) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if ref == nil || ref.Provider() != auth.ProviderOpenAI {
		return auth.ErrWrongProvider
	}
	// No provider acceptance endpoint is part of the verified contract. An
	// owned generation proves store ownership; SendAuthorized still applies
	// the generation barrier. Unknown/foreign refs fail closed here.
	_, err := ref.Generation()
	return err
}

func gatewayGPTModels() []gateway.ModelEntry {
	ids := []string{"gpt-6-astra", "gpt-5.6-sol", "gpt-5.6-terra", "gpt-5.6-luna"}
	models := make([]gateway.ModelEntry, 0, len(ids))
	for _, id := range ids {
		models = append(models, gateway.ModelEntry{RouteID: id, UpstreamID: id, Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer, Capabilities: gateway.Capabilities{ContextTokens: gatewayContextWindow, Images: true, Tools: true, Streaming: true}})
	}
	return models
}

func installedGatewayGPTModels() ([]gateway.ModelEntry, error) {
	home, err := paths.MoaiHome()
	if err != nil || !filepath.IsAbs(home) {
		return nil, errors.New("GPT model metadata profile unavailable")
	}
	return gatewayGPTModelsForProfile(filepath.Join(home, "gpt-appserver"))
}

// Only an absent cache uses the observed conservative default. Invalid present
// metadata must not silently widen or guess the single Claude session window.
func gatewayGPTModelsForProfile(profile string) ([]gateway.ModelEntry, error) {
	models := gatewayGPTModels()
	f, err := os.Open(filepath.Join(profile, "models_cache.json"))
	if errors.Is(err, os.ErrNotExist) {
		return models, nil
	}
	invalid := errors.New("GPT context metadata invalid; refresh the managed Codex models cache")
	if err != nil {
		return nil, invalid
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > 4<<20 {
		return nil, invalid
	}
	raw, err := io.ReadAll(io.LimitReader(f, (4<<20)+1))
	if err != nil || len(raw) > 4<<20 || gateway.ValidateJSONObject(raw) != nil {
		return nil, invalid
	}
	var cache struct {
		Models []struct {
			Slug   string `json:"slug"`
			Window int    `json:"context_window"`
		} `json:"models"`
	}
	if json.Unmarshal(raw, &cache) != nil {
		return nil, invalid
	}
	windows := make(map[string]int, len(models))
	for _, row := range models {
		windows[row.RouteID] = 0
	}
	window := 0
	for _, row := range cache.Models {
		previous, supported := windows[row.Slug]
		if !supported {
			continue
		}
		if previous != 0 || row.Window <= 0 || row.Window > gatewayContextWindowLimit || (window != 0 && window != row.Window) {
			return nil, invalid
		}
		windows[row.Slug], window = row.Window, row.Window
	}
	for i := range models {
		if windows[models[i].RouteID] == 0 {
			return nil, invalid
		}
		models[i].Capabilities.ContextTokens = window
	}
	return models, nil
}

func newGatewaySessionToken() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw[:]), nil
}

func marshalGatewayPrivatePayload(token string, models []gateway.ModelEntry) ([]byte, error) {
	ids := make([]string, 0, len(models))
	for _, model := range models {
		ids = append(ids, model.RouteID)
	}
	return json.Marshal(gatewayPrivatePayload{Version: 1, SessionToken: token, ModelIDs: ids, ContextTokens: gatewayPayloadContext(models)})
}

func marshalGatewayPrivatePayloadWithConversation(token string, models []gateway.ModelEntry, descriptor conversation.Descriptor) ([]byte, error) {
	ids := make([]string, 0, len(models))
	for _, model := range models {
		ids = append(ids, model.RouteID)
	}
	conversation := &gatewayPrivateConversation{FamilyID: descriptor.FamilyID, SessionID: descriptor.UUID, ReceiptDir: descriptor.ReceiptDir, CWD: descriptor.CWD}
	return json.Marshal(gatewayPrivatePayload{Version: 1, SessionToken: token, ModelIDs: ids, Conversation: conversation, ContextTokens: gatewayPayloadContext(models)})
}

func gatewayPayloadContext(models []gateway.ModelEntry) int {
	if len(models) > 0 && models[0].Provider == gateway.ProviderOpenAI {
		return models[0].Capabilities.ContextTokens
	}
	return 0
}

func gatewayChildEnvironment(inherited []string) []string {
	scrub := map[string]bool{"Z_AI_API_KEY": true, "OPENAI_API_KEY": true, config.EnvAnthropicAPIKey: true, "CODEX_HOME": true}
	for _, key := range gatewayScrubKeys() {
		scrub[key] = true
	}
	out := make([]string, 0, len(inherited))
	for _, item := range inherited {
		for i := 0; i < len(item); i++ {
			if item[i] == '=' {
				if !scrub[item[:i]] {
					out = append(out, item)
				}
				break
			}
		}
	}
	return out
}
