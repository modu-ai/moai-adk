package cli

// plugin_install.go — SPEC-PLUGIN-MARKETPLACE-001 M3a, the moai plugin install
// step (RED stub: compiles, performs nothing).

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/spf13/cobra"
)

const (
	pluginMarketplaceSource = "modu-ai/moai-adk"
	pluginRef               = "moai@moai-adk"
)

type pluginTool string

const (
	pluginToolClaude pluginTool = "claude"
	pluginToolCodex  pluginTool = "codex"
)

var errPluginRunnerRefused = errors.New("plugin command runner refused under a test binary")

type pluginCommandRunner interface {
	Run(ctx context.Context, bin string, args []string, env []string) ([]byte, error)
}

type execPluginRunner struct{}

func (execPluginRunner) Run(context.Context, string, []string, []string) ([]byte, error) {
	return nil, nil
}

var pluginRunner pluginCommandRunner = execPluginRunner{}

type pluginInstallOptions struct {
	Tools       []pluginTool
	ProjectRoot string
	NoPlugin    bool
	Timeout     time.Duration
}

func newPluginInstallOptions(tools []pluginTool, projectRoot string, noPlugin bool) pluginInstallOptions {
	return pluginInstallOptions{Tools: tools, ProjectRoot: projectRoot, NoPlugin: noPlugin}
}

func isPluginTestProgram(string) bool { return false }

func isPluginTestBinary() bool { return false }

func runPluginInstallStep(io.Writer, pluginInstallOptions) error { return nil }

func runPluginInstallStepForInit(*cobra.Command, agentWiring, string) {}
