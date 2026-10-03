package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// claude_binary.go resolves the Claude Code binary for a launch
// (issue #1697): an operator pins a known-good release when a newer Claude
// Code release breaks compatibility with a third-party endpoint, instead of
// every lane auto-adopting the broken build at the next launch.
//
// Resolution order:
//  1. the MOAI_CLAUDE_BIN environment variable (per-launch override),
//  2. the llm.claude_bin config key in .moai/config/sections/llm.yaml
//     (durable per-project pin),
//  3. exec.LookPath("claude") — the unchanged default.
//
// A configured pin is validated: the path must exist and be executable. An
// invalid pin is a launch ERROR, never a silent fallback — a pin that quietly
// fell back to PATH would re-expose exactly the blast radius the pin exists
// to stop, and the operator would have no signal that they are unprotected.

// resolveLaunchClaudeBinary resolves the binary launchClaudeDefault hands the
// process over to. See the file comment for the resolution order and the
// fail-loud contract. The llm.claude_bin pin is read from the project the
// working directory sits in.
func resolveLaunchClaudeBinary() (string, error) {
	root, err := findProjectRoot()
	if err != nil {
		root = ""
	}
	return resolveClaudeBinaryAt(root)
}

// claudeNotFoundError is the typed "no claude on PATH" class. The plugin
// install step tells it apart from an invalid pin: the first is one skip line
// (REQ-014), the second one guidance block (REQ-013). The message is the one
// the launcher has always printed.
type claudeNotFoundError struct{}

func (*claudeNotFoundError) Error() string {
	return "claude not found in PATH. Install Claude Code first"
}

// resolveClaudeBinaryAt is resolveLaunchClaudeBinary for an explicit project
// root ("" = no project, so no config pin): `moai init` resolves the pin of the
// project it initialises, not of whatever project the working directory is in.
func resolveClaudeBinaryAt(projectRoot string) (string, error) {
	if pin := os.Getenv(config.EnvClaudeBin); pin != "" {
		return validateClaudeBinaryPin(pin, "env var "+config.EnvClaudeBin)
	}
	if projectRoot != "" {
		sectionsDir := filepath.Join(filepath.Clean(projectRoot), defs.MoAIDir, defs.SectionsSubdir)
		if llm, err := loadLLMSectionOnly(sectionsDir); err == nil && llm.ClaudeBin != "" {
			return validateClaudeBinaryPin(llm.ClaudeBin,
				"llm.claude_bin in .moai/config/sections/llm.yaml")
		}
	}
	claudeBin, err := exec.LookPath("claude")
	if err != nil {
		return "", &claudeNotFoundError{}
	}
	return claudeBin, nil
}

// validateClaudeBinaryPin validates an explicit binary pin. The path must
// exist and point at an executable file; source names the configuration
// surface the pin came from so the error tells the operator where to fix it.
// The executable-bit check is POSIX-only: Windows has no executable mode bit
// (executability is extension-based there), so on Windows existence and
// not-a-directory are the check.
func validateClaudeBinaryPin(pin, source string) (string, error) {
	info, err := os.Stat(pin)
	if err != nil {
		return "", fmt.Errorf("claude binary pin %q (set via %s) does not exist: %w — fix or clear the pin to fall back to PATH lookup", pin, source, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("claude binary pin %q (set via %s) is a directory, not an executable file — point the pin at the binary itself", pin, source)
	}
	if runtime.GOOS != "windows" && info.Mode()&0o111 == 0 {
		return "", fmt.Errorf("claude binary pin %q (set via %s) is not executable — chmod +x it or point the pin at an executable file", pin, source)
	}
	return pin, nil
}
