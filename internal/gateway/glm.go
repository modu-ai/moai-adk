package gateway

import "github.com/modu-ai/moai-adk/internal/gateway/auth"

// NewGLMAdapter relays ordinary Messages only to the fixed Z.AI endpoint. The
// shared sender never forwards Anthropic-Beta on this provider's requests.
func NewGLMAdapter(c MessagesConfig) (*MessagesAdapter, error) {
	return newMessagesAdapter(c, ProviderZAI, AuthExistingGLM, auth.GLMEndpoint)
}
