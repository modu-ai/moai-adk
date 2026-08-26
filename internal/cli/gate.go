package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook/quality"
)

// gateCmd exposes the heavy commit-quality gate as a standalone CLI verb so
// the git pre-commit hook (internal/cli/hook_install_precommit.go) can shell
// out to it without being subject to Claude Code's 5s PreToolUse hook budget.
// This is the relocation surface for SPEC-PRETOOL-GATE-MOVE-001: the same
// QualityGate.Run path that used to silently drop its deny verdict after 5s
// now runs in the user's shell via `moai gate`, with the full 210s step
// budget (vet 30s + lint 60s + test 120s) available.
//
// @MX:ANCHOR: [AUTO] heavy commit-quality gate CLI entry point; invoked by the git pre-commit hook
// @MX:REASON: this verb is the relocation target for the PreToolUse 5s budget defect (census C-2); the pre-commit hook shells out to it
var gateCmd = &cobra.Command{
	Use:   "gate",
	Short: "Run the heavy commit-quality gate (vet + lint + test, 16-language detection)",
	Long: `Run the MoAI-ADK quality gate outside the PreToolUse 5s budget.

Invoked by the git pre-commit hook (installed via 'moai init' / 'moai update')
to enforce vet / lint / test at commit time. Runs in the user's shell, so the
full step budget (vet 30s + lint 60s + test 120s per .moai/config/sections/gate.yaml)
is available — unlike the PreToolUse hook path which is silently dropped after 5s.

Exit code 0: gate passed (or no recognized language toolchain detected).
Exit code 1: gate failed; diagnostic output on stderr.

Project directory resolution priority:
  1. $CLAUDE_PROJECT_DIR (when invoked from inside Claude Code)
  2. current working directory

Go build tags are read from .moai/config/build-tags (first non-comment non-blank
line) when present, mirroring the PreToolUse gate path.`,
	RunE: runGate,
}

func init() {
	rootCmd.AddCommand(gateCmd)
}

// runGate executes the quality gate and propagates pass/fail as the process
// exit code. Failure output goes to stderr so git's pre-commit hook surfaces
// it to the caller (verified by M1.c — git pre-commit stderr reaches the Bash
// tool result).
func runGate(cmd *cobra.Command, _ []string) error {
	projectDir := resolveGateProjectDir()
	cfg := loadGateCfgForCLI(projectDir)

	// Overall deadline: vet (30s) + lint (60s) + test (120s) + ast-grep + slack.
	// Per-step timeouts inside the gate enforce the real ceilings; this outer
	// deadline is a safety net against a runaway step hanging indefinitely.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	gate := quality.NewQualityGate(cfg)
	passed, output := gate.Run(ctx)
	if !passed {
		if output != "" {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), output)
		}
		return fmt.Errorf("quality gate failed")
	}
	// Pass path: a non-empty output carries a non-blocking notice (e.g. an
	// ast-grep step that skipped because sg is absent). Surface it on stderr
	// without failing the gate.
	if output != "" {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), output)
	}
	return nil
}

// @MX:NOTE: loadGateCfgForCLI is the SSOT seam that routes the standalone
// `moai gate` CLI through the same config.loadGateSection loader as the
// PreToolUse path (internal/hook/pre_tool.go loadGateConfig), so the CLI and
// the hook share one gate.yaml source for Enabled/BlockOnError/WarnOnlyMode
// (SPEC-GATE-ASTGREP-REPAIR-001 M3 / REQ-GAR-006 / D3 config-behavior-consistency fix).
//
// loadGateCfgForCLI reads the gate configuration from
// <projectDir>/.moai/config/sections/gate.yaml via the shared config.loadGateSection
// loader (SPEC-GATE-ASTGREP-REPAIR-001 M3 / REQ-GAR-006). This is the same
// loader the PreToolUse path uses (internal/hook/pre_tool.go loadGateConfig),
// so the standalone `moai gate` CLI and the PreToolUse hook share one SSOT for
// Enabled/BlockOnError/WarnOnlyMode.
//
// Fall-back policy (REQ-GAR-007 / AP-GAR-004): when the config cannot be
// loaded (file missing, unparseable, or projectDir empty), the function
// returns quality.DefaultGateConfig() with the project dir + build tags
// applied. The default is advisory (WarnOnlyMode would be false here but the
// gate's own block predicate requires BlockOnError && !WarnOnlyMode, and
// DefaultGateConfig sets BlockOnError=true — so the fallback IS blocking).
// This matches the pre-M3 behavior so the transition does not silently weaken
// the gate for projects that relied on the hard default. Projects that want
// advisory mode set it explicitly in gate.yaml, which IS read after M3.
//
// The config.GateConfig → quality.GateConfig mapping mirrors
// pre_tool.go loadGateConfig verbatim, including the disabled_steps verbatim
// copy (issue #1265: the runner reads FALSE as "skip"). An empty RulesDir
// stays empty (t50): gate.yaml ast_grep_gate.rules_dir is the SSOT.
func loadGateCfgForCLI(projectDir string) *quality.GateConfig {
	qcfg := quality.DefaultGateConfig()

	if projectDir != "" {
		moaiDir := filepath.Join(projectDir, ".moai")
		loader := config.NewLoader()
		loaded, err := loader.Load(moaiDir)
		if err != nil {
			slog.Warn("moai gate: config load failed, using hardcoded defaults",
				"moaiDir", moaiDir, "error", err)
		} else if loaded != nil {
			qcfg = mapConfigGateToQuality(loaded.Gate)
		}
	}

	qcfg.ProjectDir = projectDir
	qcfg.GoBuildTags = readGateGoBuildTags(projectDir)
	return qcfg
}

// mapConfigGateToQuality converts the config-layer GateConfig into the
// quality-layer GateConfig shape. It mirrors pre_tool.go loadGateConfig's
// mapping so both paths agree on timeout/disabled_steps/ast-grep field
// interpretation.
func mapConfigGateToQuality(g config.GateConfig) *quality.GateConfig {
	qcfg := &quality.GateConfig{
		Enabled:     g.Enabled,
		SkipTests:   g.SkipTests,
		VetTimeout:  g.VetTimeoutDuration(),
		LintTimeout: g.LintTimeoutDuration(),
		TestTimeout: g.TestTimeoutDuration(),

		TypecheckEnabled: g.Typecheck.Enabled,
		TypecheckCommand: g.Typecheck.Command,
		TypecheckTimeout: g.TypecheckTimeoutDuration(),
	}
	// Map gate.disabled_steps through verbatim (issue #1265): the runner reads
	// FALSE as "skip this step", so normalising values here would silently stop
	// the skip from applying.
	if len(g.DisabledSteps) > 0 {
		qcfg.DisabledSteps = make(map[string]bool, len(g.DisabledSteps))
		for k, v := range g.DisabledSteps {
			qcfg.DisabledSteps[k] = v
		}
	}
	ag := g.AstGrepGate
	qcfg.AstGrepGate = &quality.AstGrepGateConfig{
		Enabled:      ag.Enabled,
		RulesDir:     ag.RulesDir,
		BlockOnError: ag.BlockOnError,
		WarnOnlyMode: ag.WarnOnlyMode,
	}
	// Graph-freshness step (gate.yaml graph_freshness): zero-value thresholds
	// mean "not configured" and fall back to the quality-layer defaults, so a
	// partial gate.yaml cannot silently zero a red line.
	gf := g.GraphFreshness
	qgf := quality.DefaultGraphFreshnessConfig()
	qgf.Enabled = gf.Enabled
	qgf.Blocking = gf.Blocking
	if gf.CodemapsChangedFiles > 0 {
		qgf.CodemapsChangedFiles = gf.CodemapsChangedFiles
	}
	if gf.MXIndexChangedFiles > 0 {
		qgf.MXIndexChangedFiles = gf.MXIndexChangedFiles
	}
	qcfg.GraphFreshness = qgf
	// t50: an empty gate.yaml rules_dir maps through as empty — no hardcoded
	// path fallback. gate.yaml ast_grep_gate.rules_dir is the SSOT for where a
	// project keeps its ast-grep rules; the template ships the key with an
	// explicit value (.moai/config/astgrep-rules, where it deploys the rules),
	// so template users are covered by config rather than by a code guess.
	return qcfg
}

// resolveGateProjectDir returns the project directory the gate should run
// against, honouring the same $CLAUDE_PROJECT_DIR priority the PreToolUse
// hook path uses (CLAUDE.local.md B7). Falls back to os.Getwd() when the env
// var is unset or empty.
func resolveGateProjectDir() string {
	if dir := strings.TrimSpace(os.Getenv("CLAUDE_PROJECT_DIR")); dir != "" {
		if abs, err := filepath.Abs(dir); err == nil {
			return abs
		}
		return dir
	}
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return ""
}

// readGateGoBuildTags reads the project Go build tags from
// <dir>/.moai/config/build-tags (first non-empty, non-comment line). Mirrors
// the unexported readGoBuildTags helper in internal/hook/pre_tool.go so the
// standalone gate CLI honours the same build-tag file a project configures
// for the PreToolUse path. Returns "" when the file is absent or unreadable.
//
// The duplication is deliberate: pre_tool.go's helper is an unexported method
// on a handler struct, unreachable from the cli package without a wider
// refactor. Consolidating into a shared internal/hook/quality helper is a
// follow-up; both readers are leaf file reads with no side effects.
func readGateGoBuildTags(dir string) string {
	if dir == "" {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(dir, ".moai", "config", "build-tags"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return line
	}
	return ""
}
