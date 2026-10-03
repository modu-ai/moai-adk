package cli

// plugin_install.go — SPEC-PLUGIN-MARKETPLACE-001 M3a, the moai plugin install
// step. One function, called by `moai init` (gated by the --llm harness) and by
// the `moai plugin install` verb (no harness: every tool found), runs two
// commands per tool through an injected runner:
//
//	claude plugin marketplace add modu-ai/moai-adk ; claude plugin install moai@moai-adk
//	codex  plugin marketplace add modu-ai/moai-adk ; codex  plugin add     moai@moai-adk
//
// The second command runs only after the first exits 0. The step is fail-open:
// an absent tool prints one skip line, a failure, a timeout or an invalid
// llm.claude_bin pin prints one guidance block naming that tool's two manual
// commands, and the step returns nil in every case. The child inherits the
// invoking environment unchanged; the step never sets or clears
// CLAUDE_CONFIG_DIR or CODEX_HOME.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
)

const (
	// pluginMarketplaceSource is the GitHub marketplace both tools add.
	pluginMarketplaceSource = "modu-ai/moai-adk"
	// pluginRef is the plugin both tools install from that marketplace.
	pluginRef = "moai@moai-adk"
)

// pluginTool names a tool the step can act on.
type pluginTool string

const (
	pluginToolClaude pluginTool = "claude"
	pluginToolCodex  pluginTool = "codex"
)

// pluginToolSpec is the per-tool vocabulary: the commands, the human label and
// the config home the tool resolves.
type pluginToolSpec struct {
	label       string
	add         []string
	install     []string
	homeEnv     string
	homeDefault string
}

func pluginSpec(tool pluginTool) pluginToolSpec {
	if tool == pluginToolCodex {
		return pluginToolSpec{
			label:       "Codex",
			add:         []string{"plugin", "marketplace", "add", pluginMarketplaceSource},
			install:     []string{"plugin", "add", pluginRef},
			homeEnv:     codexHomeEnvVar,
			homeDefault: "~/.codex",
		}
	}
	return pluginToolSpec{
		label:       "Claude Code",
		add:         []string{"plugin", "marketplace", "add", pluginMarketplaceSource},
		install:     []string{"plugin", "install", pluginRef},
		homeEnv:     config.EnvClaudeConfigDir,
		homeDefault: "~/.claude",
	}
}

// errPluginRunnerRefused is what the default runner returns under a test
// binary: nothing was started.
var errPluginRunnerRefused = errors.New("plugin command runner refused: the process is a test binary")

// pluginCommandRunner starts one tool command and returns its combined output.
//
// @MX:NOTE: [AUTO] The seam shared by the install step and (M4) the doctor's
// Codex probe. The default refuses under a test binary (REQ-017), so every test
// that reaches the step through runInit, initCmd.RunE or the doctor registry is
// inert without an edit; a test that wants the step to run injects a runner.
type pluginCommandRunner interface {
	Run(ctx context.Context, bin string, args []string, env []string) ([]byte, error)
}

// execPluginRunner is the production runner: a real process, bounded by ctx.
type execPluginRunner struct{}

func (execPluginRunner) Run(ctx context.Context, bin string, args []string, env []string) ([]byte, error) {
	if isPluginTestBinary() {
		return nil, errPluginRunnerRefused
	}
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Env = env
	cmd.WaitDelay = config.DefaultPluginCommandWaitDelay
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.Bytes(), err
}

// pluginRunner is the package-level seam; tests replace it.
var pluginRunner pluginCommandRunner = execPluginRunner{}

// isPluginTestProgram reports whether argv0 names a Go test binary. It looks at
// the program name only, never at an argument: `moai init my-app.test` is a
// project name, not a test run.
//
// @MX:NOTE: [AUTO] Deliberately not isTestEnvironment() (glm.go), which returns
// true when ANY argument ends in ".test" and would silently skip the step in
// production for a project named app.test (design.md section 3.4).
func isPluginTestProgram(argv0 string) bool {
	base := argv0
	if i := strings.LastIndexAny(base, `/\`); i >= 0 {
		base = base[i+1:]
	}
	return strings.HasSuffix(base, ".test") || strings.HasSuffix(base, ".test.exe")
}

// isPluginTestBinary reports whether this process is a Go test binary.
func isPluginTestBinary() bool {
	return testing.Testing() || isPluginTestProgram(os.Args[0])
}

// pluginInstallOptions is one run of the step.
type pluginInstallOptions struct {
	Tools       []pluginTool  // in order of action
	ProjectRoot string        // project whose llm.claude_bin pin applies ("" = none)
	NoPlugin    bool          // the --no-plugin flag
	Timeout     time.Duration // bound of ONE command
}

// newPluginInstallOptions is the production wiring: the bound is the config
// constant.
func newPluginInstallOptions(tools []pluginTool, projectRoot string, noPlugin bool) pluginInstallOptions {
	return pluginInstallOptions{
		Tools:       tools,
		ProjectRoot: projectRoot,
		NoPlugin:    noPlugin,
		Timeout:     config.DefaultPluginInstallCommandTimeout,
	}
}

// pluginToolsForHarness is the --llm gate, the same rule wireCodexUnlessClaude
// applies: claude selects Claude, gpt Codex, both Claude then Codex.
func pluginToolsForHarness(wiring agentWiring) []pluginTool {
	switch wiring {
	case agentWiringGPT:
		return []pluginTool{pluginToolCodex}
	case agentWiringBoth:
		return []pluginTool{pluginToolClaude, pluginToolCodex}
	default:
		return []pluginTool{pluginToolClaude}
	}
}

// pluginOptOutFromEnv reads MOAI_SKIP_PLUGIN_INSTALL: "1" or "true" opts out; an
// empty value or "0" does not.
func pluginOptOutFromEnv() bool {
	v := strings.TrimSpace(os.Getenv(config.EnvSkipPluginInstall))
	return v == "1" || strings.EqualFold(v, "true")
}

// runPluginInstallStepForInit is the runInit call: the harness selects the
// tools, the guidance goes to stderr.
func runPluginInstallStepForInit(cmd *cobra.Command, wiring agentWiring, projectRoot string) {
	opts := newPluginInstallOptions(pluginToolsForHarness(wiring), projectRoot, getBoolFlag(cmd, "no-plugin"))
	_ = runPluginInstallStep(cmd.ErrOrStderr(), opts)
}

// runPluginInstallStep acts on opts.Tools in order and returns nil in every
// case (REQ-013 to REQ-015); everything it has to say goes to out.
func runPluginInstallStep(out io.Writer, opts pluginInstallOptions) error {
	if opts.NoPlugin || pluginOptOutFromEnv() {
		return nil
	}
	// A test binary reaching the default runner is inert by construction; stay
	// silent so the many tests that call runInit see no new output.
	if _, isDefault := pluginRunner.(execPluginRunner); isDefault && isPluginTestBinary() {
		return nil
	}
	if opts.Timeout <= 0 {
		opts.Timeout = config.DefaultPluginInstallCommandTimeout
	}
	for _, tool := range opts.Tools {
		installPluginFor(out, opts, tool)
	}
	return nil
}

// installPluginFor runs the two-command sequence for one tool.
func installPluginFor(out io.Writer, opts pluginInstallOptions, tool pluginTool) {
	spec := pluginSpec(tool)
	name := string(tool)

	bin, err := resolvePluginBinary(tool, opts.ProjectRoot)
	if err != nil {
		var notFound *claudeNotFoundError
		if tool == pluginToolCodex || errors.As(err, &notFound) {
			_, _ = fmt.Fprintf(out, "note: %s not found on PATH; skipping the moai plugin install\n", name)
			return
		}
		printPluginGuidance(out, spec, name, err.Error())
		return
	}

	home := os.Getenv(spec.homeEnv)
	if home == "" {
		home = spec.homeDefault
	}
	_, _ = fmt.Fprintf(out, "moai plugin: %s config home: %s\n", spec.label, home)

	for _, args := range [][]string{spec.add, spec.install} {
		if reason := runPluginCommand(opts.Timeout, bin, args); reason != "" {
			printPluginGuidance(out, spec, name, reason)
			return
		}
	}
	_, _ = fmt.Fprintf(out, "moai plugin: installed the moai plugin for %s\n", spec.label)
}

// resolvePluginBinary finds the tool binary: Claude through the pinned-binary
// resolution of the given project, Codex through the PATH seam.
func resolvePluginBinary(tool pluginTool, projectRoot string) (string, error) {
	if tool == pluginToolCodex {
		return codexWiringLookPath("codex")
	}
	return resolveClaudeBinaryAt(projectRoot)
}

// runPluginCommand runs one command under the per-command bound with the
// invoking environment unchanged. It returns "" on exit 0 and a one-line
// reason otherwise.
func runPluginCommand(timeout time.Duration, bin string, args []string) string {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	output, err := pluginRunner.Run(ctx, bin, args, os.Environ())
	if err == nil {
		return ""
	}
	reason := err.Error()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		reason = fmt.Sprintf("timed out after %s", timeout)
	}
	if tail := strings.Join(strings.Fields(boundedTail(output)), " "); tail != "" {
		reason += ": " + tail
	}
	return reason
}

// printPluginGuidance prints the one guidance block for a tool: the reason and
// the two manual commands.
func printPluginGuidance(out io.Writer, spec pluginToolSpec, name, reason string) {
	_, _ = fmt.Fprintf(out, "note: could not install the moai plugin for %s (%s). Install it yourself:\n", spec.label, reason)
	_, _ = fmt.Fprintf(out, "        %s %s\n", name, strings.Join(spec.add, " "))
	_, _ = fmt.Fprintf(out, "        %s %s\n", name, strings.Join(spec.install, " "))
}
