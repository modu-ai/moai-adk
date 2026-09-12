package cli

import (
	"errors"
	"net"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/gateway"
)

type gatewayPrepareInput struct {
	Mode, ExplicitModel, ClaudeDefault string
	GLM                                config.GLMModels
	GLMKey                             string
	Catalog                            gateway.CatalogSnapshot
	Inherited                          []string
	Address                            string
	// SessionEnv is supplied only by a verified transport binding. There is no default auth key.
	SessionEnv []string
}

// gatewayScrubKeys returns fresh storage: callers cannot change future launches.
func gatewayScrubKeys() []string {
	return []string{config.EnvAnthropicAuthToken, config.EnvAnthropicBaseURL, config.EnvAnthropicDefaultOpusModel, config.EnvAnthropicDefaultSonnetModel, config.EnvAnthropicDefaultHaikuModel, config.EnvAnthropicDefaultFableModel, "MOAI_BACKUP_AUTH_TOKEN", config.EnvClaudeCodeDisableExperimentalBetas, "API_TIMEOUT_MS", "CLAUDE_CODE_AUTO_COMPACT_WINDOW", "CLAUDE_CODE_MAX_CONTEXT_TOKENS", config.EnvClaudeCodeDisableNonessentialTraffic, config.EnvClaudeCodeTeammateDisplay, "MOAI_STATUSLINE_CONTEXT_SIZE"}
}

// @MX:WARN: [AUTO] Launch mode, exact model and inherited environment are separate trust boundaries.
// @MX:REASON: GLM tier aliases must not reintroduce inherited routing keys or change the initial provider signal.
func prepareGatewayLaunch(in gatewayPrepareInput) (gateway.LaunchPlan, error) {
	host, port, err := net.SplitHostPort(in.Address)
	n, _ := strconv.Atoi(port)
	if err != nil || host != "127.0.0.1" || n < 1 || n > 65535 {
		return gateway.LaunchPlan{}, errors.New("gateway handoff requires a bound loopback address")
	}
	model := in.ExplicitModel
	if model == "" {
		switch in.Mode {
		case "claude":
			model = in.ClaudeDefault
			if model == "" {
				model = "claude-opus-5"
			}
		case "gpt":
			model = "gpt-5.6-sol"
		case "glm":
			model = in.GLM.High
		default:
			return gateway.LaunchPlan{}, errors.New("unsupported gateway launcher mode")
		}
	}
	if in.Mode != "claude" && in.Mode != "gpt" && in.Mode != "glm" {
		return gateway.LaunchPlan{}, errors.New("unsupported gateway launcher mode")
	}
	if in.Mode == "glm" {
		switch model {
		case "opus":
			model = in.GLM.High
		case "sonnet":
			model = in.GLM.Medium
		case "haiku":
			model = in.GLM.Low
		case "fable":
			model = in.GLM.Fable
		}
	}
	model = expandModelString(model)
	entry, err := in.Catalog.Resolve(model)
	if err != nil {
		return gateway.LaunchPlan{}, err
	}
	expected := map[string]gateway.ProviderID{"claude": gateway.ProviderAnthropic, "gpt": gateway.ProviderOpenAI, "glm": gateway.ProviderZAI}[in.Mode]
	if entry.Provider != expected {
		return gateway.LaunchPlan{}, errors.New("model belongs to another launcher provider")
	}
	var rows []gateway.ModelEntry
	for _, row := range in.Catalog.Entries() {
		if row.Provider == expected {
			rows = append(rows, row)
		}
	}
	catalog, err := gateway.NewCatalog(rows)
	if err != nil {
		return gateway.LaunchPlan{}, err
	}
	scrub := map[string]bool{"Z_AI_API_KEY": true, config.EnvAnthropicAPIKey: true, "ANTHROPIC_CUSTOM_HEADERS": true, config.EnvAnthropicModel: true, "CLAUDE_CODE_SUBAGENT_MODEL": true, "CLAUDE_CODE_DISABLE_1M_CONTEXT": true, config.EnvMoaiLaunchProvider: true}
	for _, k := range gatewayScrubKeys() {
		scrub[k] = true
	}
	if in.Mode == "gpt" {
		scrub["ENABLE_TOOL_SEARCH"] = true
	}
	env := make([]string, 0, len(in.Inherited)+8)
	for _, item := range in.Inherited {
		k, _, ok := strings.Cut(item, "=")
		if ok && !scrub[k] {
			env = append(env, item)
		}
	}
	if in.Mode == "glm" {
		if in.GLMKey == "" {
			return gateway.LaunchPlan{}, errors.New("GLM credential unavailable")
		}
		for _, id := range []string{in.GLM.High, in.GLM.Medium, in.GLM.Low, in.GLM.Fable} {
			e, err := in.Catalog.Resolve(id)
			if err != nil || e.Provider != gateway.ProviderZAI {
				return gateway.LaunchPlan{}, errors.New("GLM tier is not registered with Z.AI")
			}
		}
		env = append(env, "Z_AI_API_KEY="+in.GLMKey, "ANTHROPIC_DEFAULT_OPUS_MODEL="+in.GLM.High, "ANTHROPIC_DEFAULT_SONNET_MODEL="+in.GLM.Medium, "ANTHROPIC_DEFAULT_HAIKU_MODEL="+in.GLM.Low, "ANTHROPIC_DEFAULT_FABLE_MODEL="+in.GLM.Fable)
	}
	if in.Mode != "glm" {
		for _, tier := range []string{"OPUS", "SONNET", "HAIKU", "FABLE"} {
			env = append(env, "ANTHROPIC_DEFAULT_"+tier+"_MODEL="+model)
		}
	}
	provider := "claude"
	switch entry.Provider {
	case gateway.ProviderOpenAI:
		provider = "gpt"
	case gateway.ProviderZAI:
		provider = "glm"
	}
	// GPT uses an explicit picker allowlist and declared custom-model window.
	// The legacy GLM gateway seam still suppresses implicit Claude variants.
	if in.Mode == "glm" {
		env = append(env, "CLAUDE_CODE_DISABLE_1M_CONTEXT=1")
	}
	// Native GPT translation has no tool_reference mapping; load full MCP schemas.
	if in.Mode == "gpt" {
		// Claude reads this once at startup. Use the smallest selectable window
		// so an in-session /model switch cannot exceed another route's limit.
		window := entry.Capabilities.ContextTokens
		for _, row := range rows {
			if row.Capabilities.ContextTokens < window {
				window = row.Capabilities.ContextTokens
			}
		}
		if window > 0 {
			env = append(env, "CLAUDE_CODE_MAX_CONTEXT_TOKENS="+strconv.Itoa(window))
		}
		env = append(env, "ENABLE_TOOL_SEARCH=false")
	}
	env = append(env, "ANTHROPIC_BASE_URL=http://"+in.Address, config.EnvMoaiLaunchProvider+"="+provider)
	env = append(env, in.SessionEnv...)
	return gateway.LaunchPlan{InitialModel: model, InitialProvider: entry.Provider, Catalog: catalog, ChildEnv: env}, nil
}
