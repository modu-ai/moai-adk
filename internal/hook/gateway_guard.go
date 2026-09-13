package hook

import (
	"github.com/modu-ai/moai-adk/internal/config"
	"os"
)

// isGatewaySession recognizes only launcher-owned initial provider values.
func isGatewaySession() bool {
	switch os.Getenv(config.EnvMoaiLaunchProvider) {
	case "claude", "gpt", "glm":
		return true
	default:
		return false
	}
}
