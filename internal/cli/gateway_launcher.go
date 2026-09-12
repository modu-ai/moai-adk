package cli

import (
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/gateway"
)

func currentLaunchCWD() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Clean(cwd)
}

func envKeyPresent(key string) bool {
	prefix := key + "="
	for _, item := range os.Environ() {
		if len(item) >= len(prefix) && item[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

type gatewayLaunchRequest struct {
	Mode, ExplicitModel, ClaudeDefault, ProfileName string
	Inherited, Args                                 []string
	// CWD is the actual native process working directory. Project is only the
	// stable index key; it must never be used as a substitute for CWD.
	CWD, Project, SecureStorage, OriginalConfig string
	SecureStorageSet, OriginalConfigSet         bool
	Continue                                    bool
}

// gatewayLaunchBinding is constructed only after verified launch gates. Prepare
// owns supervisor startup and settings merging and returns cleanup for failures
// and continue-return paths. POSIX exec is covered by the supervisor parent watch.
type gatewayLaunchBinding struct {
	Mode     string
	Prepare  func(gatewayLaunchRequest) (gateway.LaunchPlan, func(), error)
	Continue func(string, []string, []string) error
}
