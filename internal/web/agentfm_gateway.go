package web

import (
	"os"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
)

// agentFMLiveLLM is the ONE place the console folds the launcher-owned
// gateway signal (MOAI_LAUNCH_PROVIDER, set on the child env of a `moai gpt`
// launch) into the loaded llm.yaml config. A gateway launch persists nothing
// into llm.yaml, so without the fold a console started inside a gpt session
// would render editable model selects the launcher makes meaningless. Both
// the GET view seed and the POST parse path call this so render and save agree.
func agentFMLiveLLM(llm config.LLMConfig) config.LLMConfig {
	return llm.WithLaunchProvider(os.Getenv(config.EnvMoaiLaunchProvider))
}

// agentFMIsGatewayBackend reports whether the live llm config marks a gateway
// (gpt) session backend. Thin reuse of the config-level intent signal so the
// panel gate and the CLI resolver (`moai model profile`) read the same
// predicate (t840 web half) — the mirror of agentFMIsGLMBackend.
//
// Under a gateway backend the launcher pins every alias slot to one gpt model
// id, so the model cell renders as a fixed "inherit" display (no select) and
// effort stays the only editable per-agent axis. The save path is unaffected:
// an absent model field is backfilled by parseAgentFMForm with the resolved
// value, exactly as an unsubmitted (disabled) effort already is.
func agentFMIsGatewayBackend(llm config.LLMConfig) bool {
	return template.IsGatewayBackend(llm)
}
