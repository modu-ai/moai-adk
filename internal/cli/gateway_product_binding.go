package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/exec"
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

// Current subscription model metadata permits this maximum configuration.
// Official App Server long-history probes are recorded under t649.
const gatewayContextWindow = 872000
const gatewayRequestBodyLimit = 16 << 20

// newGPTGatewayBinding is the production launch seam. It creates only the
// private child configuration; credentials remain in the child-owned store and
// are resolved by the request adapter at send time.
func newGPTGatewayBinding() (*gatewayLaunchBinding, error) {
	return newGPTGatewayBindingWithStart(nil)
}

func newGPTGatewayBindingWithStart(start func(context.Context, gateway.StartOptions) (gatewayStartedChild, error)) (*gatewayLaunchBinding, error) {
	return newProviderGatewayBinding("gpt", gatewayGPTModels(), config.GLMModels{}, "", start)
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
		rows = append(rows, gateway.ModelEntry{RouteID: id, UpstreamID: id, Provider: provider, AuthMethod: method, Capabilities: gateway.Capabilities{ContextTokens: 1000000, Images: true, Tools: true, Streaming: true}})
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
	models := gatewayGPTModels()
	var broker auth.Broker
	var policy translate.PolicyProfile
	first := selection.ModelIDs[0]
	if strings.HasPrefix(first, "gpt-") {
		var err error
		broker, err = installedGPTBrokerForGateway()
		if err != nil {
			return nil, errGatewayFactory
		}
		policy = translate.PolicyGPTNative
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

func installedGPTBrokerForGateway() (auth.Broker, error) {
	executable, err := exec.LookPath("codex")
	if err != nil {
		return nil, errors.New("GPT authentication broker unavailable")
	}
	executable, err = filepath.Abs(executable)
	if err != nil {
		return nil, errors.New("GPT authentication broker unavailable")
	}
	return auth.CodexBroker{Executable: executable, Timeout: 10 * time.Minute}, nil
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
		models = append(models, gateway.ModelEntry{RouteID: id, UpstreamID: id, Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthPKCE, Capabilities: gateway.Capabilities{ContextTokens: gatewayContextWindow, Images: true, Tools: true, Streaming: true}})
	}
	return models
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
	return json.Marshal(gatewayPrivatePayload{Version: 1, SessionToken: token, ModelIDs: ids})
}

func marshalGatewayPrivatePayloadWithConversation(token string, models []gateway.ModelEntry, descriptor conversation.Descriptor) ([]byte, error) {
	ids := make([]string, 0, len(models))
	for _, model := range models {
		ids = append(ids, model.RouteID)
	}
	conversation := &gatewayPrivateConversation{FamilyID: descriptor.FamilyID, SessionID: descriptor.UUID, ReceiptDir: descriptor.ReceiptDir}
	return json.Marshal(gatewayPrivatePayload{Version: 1, SessionToken: token, ModelIDs: ids, Conversation: conversation})
}

func gatewayChildEnvironment(inherited []string) []string {
	scrub := map[string]bool{"Z_AI_API_KEY": true, "OPENAI_API_KEY": true, "ANTHROPIC_API_KEY": true, "CODEX_HOME": true}
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
