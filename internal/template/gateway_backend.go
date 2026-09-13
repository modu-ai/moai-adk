package template

import "github.com/modu-ai/moai-adk/internal/config"

// IsGatewayBackend reads explicit gateway (gpt) intent — team_mode gpt, or the
// dormant mode field. Under a gateway backend the launcher pins every Claude
// alias slot (ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU,FABLE}_MODEL) to ONE gpt
// model id (internal/cli/gateway_prepare.go), so every sub-agent inherits the
// session model and the per-agent model axis carries no differentiation; the
// web matrix and `moai model profile` state that inheritance instead of
// offering a selection. Disjoint from IsGLMBackend by construction: the two
// read different closed-set values of the same fields.
//
// At runtime the signal arrives through config.LLMConfig.WithLaunchProvider
// (a gateway launch persists nothing into llm.yaml); an explicit team_mode in
// llm.yaml is honored as-is.
func IsGatewayBackend(cfg config.LLMConfig) bool {
	return cfg.TeamMode == config.TeamModeGPT || cfg.Mode == config.LLMModeGPT
}
