package cli

// @MX:NOTE: [AUTO] Pre-push hook validates commit messages against git convention
// @MX:NOTE: [AUTO] Reads commits from stdin, exits code 2 on violation per Claude Code protocol
// @MX:NOTE: [AUTO] Priority: MOAI_GIT_CONVENTION env > config > default "auto"

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/git/convention"
)

func init() {
	hookCmd.AddCommand(prePushCmd)
}

var prePushCmd = &cobra.Command{
	Use:   "pre-push",
	Short: "Validate commit messages against the configured convention",
	Long: `Validate commit messages against the configured git convention.
Reads commit messages from stdin (one per line) and validates each
against the active convention. Exits with code 2 if any violations
are found.`,
	RunE: runPrePush,
}

// prePushAction is the resolved per-mode pre-push severity dial.
//
// @MX:NOTE: [AUTO] Wires the previously-dead git_strategy.<mode>.hooks.pre_push
// config into the runtime. skip = no-op, warn = validate but allow, enforce = block.
type prePushAction int

const (
	// prePushSkip performs no validation (the per-mode opt-out, even when the gate is on).
	prePushSkip prePushAction = iota
	// prePushWarn runs validation and prints violations but allows the push (exit 0).
	prePushWarn
	// prePushEnforce runs validation and blocks the push on any violation (exit 2).
	prePushEnforce
)

// resolvePrePushAction resolves the pre-push severity dial from the active mode
// profile's hooks.pre_push value, with an optional MOAI_PRE_PUSH env override.
//
// This function is PURE: it does NOT call os.Exit and does NOT read stdin. It is
// only consulted by runPrePush AFTER the enforce_on_push gate is confirmed ON, so
// it never opens the gate itself.
//
// Resolution (when the gate is on):
//   - MOAI_PRE_PUSH env (skip / warn / enforce) takes precedence when recognized.
//   - Otherwise read ActiveModeProfile().Hooks.PrePush.
//   - Fail-safe defaults toward enforce: a nil ModeProfile or an unrecognized
//     value normalizes to enforce (the gate was explicitly turned on, so fail closed).
func resolvePrePushAction() prePushAction {
	// Env severity override sits below the gate but above the config value.
	if envVal := os.Getenv(config.EnvPrePushMode); envVal != "" {
		if action, ok := parsePrePushAction(envVal); ok {
			return action
		}
		// Unrecognized env value: fall through to the config value.
	}

	if deps != nil && deps.Config != nil {
		if cfg := deps.Config.Get(); cfg != nil {
			if profile, ok := cfg.GitStrategy.ActiveModeProfile(); ok {
				if action, parsed := parsePrePushAction(profile.Hooks.PrePush); parsed {
					return action
				}
				// Unknown value (checkStringField does not enum-validate it):
				// fail-safe to enforce.
				return prePushEnforce
			}
		}
	}

	// nil ModeProfile / unavailable config: fail-safe to enforce.
	return prePushEnforce
}

// parsePrePushAction maps a config/env string to a prePushAction. The second
// return value reports whether the value was recognized.
func parsePrePushAction(value string) (prePushAction, bool) {
	switch value {
	case "skip":
		return prePushSkip, true
	case "warn":
		return prePushWarn, true
	case "enforce":
		return prePushEnforce, true
	default:
		return prePushEnforce, false
	}
}

// decideExit maps a resolved action and violation count to the intended process
// exit code. It is PURE: it does NOT call os.Exit. Only the runPrePush boundary
// translates a non-zero result into os.Exit.
//
//   - enforce + at least one violation → 2 (block the push)
//   - skip / warn / clean-enforce      → 0 (allow the push)
func decideExit(action prePushAction, violationCount int) int {
	if action == prePushEnforce && violationCount > 0 {
		return 2
	}
	return 0
}

// runPrePush reads commit messages from stdin and validates them.
func runPrePush(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	// Master gate (D3 precedence: env > enforce_on_push > pre_push). When the
	// gate is OFF, short-circuit BEFORE consulting pre_push (REQ-PMW-007). This
	// preserves SPEC-PREPUSH-WIRING-001 OFF behavior exactly.
	if !isEnforceOnPushEnabled() {
		return nil
	}

	// Gate is ON: resolve the per-mode severity dial.
	action := resolvePrePushAction()

	// skip is a per-mode opt-out: no convention load, no validation, no output.
	if action == prePushSkip {
		return nil
	}

	// Determine repository root from CLAUDE_PROJECT_DIR or current directory.
	repoPath := os.Getenv(config.EnvClaudeProjectDir)
	if repoPath == "" {
		var err error
		repoPath, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("pre-push: determine working directory: %w", err)
		}
	}

	// Load convention configuration, honoring the auto-detection knobs (Fix A)
	// and forwarding the configured max_length (Fix B).
	convName := resolveConventionName()
	opts, maxLength := resolveAutoDetectOptions()
	mgr := convention.NewManager(repoPath)
	if err := mgr.LoadConvention(convName, opts); err != nil {
		return fmt.Errorf("pre-push: load convention: %w", err)
	}
	mgr.SetMaxLength(maxLength)

	// Read commit messages from stdin (one per line).
	input, err := readStdinLines()
	if err != nil {
		return fmt.Errorf("pre-push: read stdin: %w", err)
	}

	if len(input) == 0 {
		_, _ = fmt.Fprintln(out, "No commit messages to validate.")
		return nil
	}

	// Validate each message.
	results := mgr.ValidateMessages(input)
	conv := mgr.Convention()

	violations := 0
	for _, r := range results {
		if !r.Valid {
			violations++
		}
	}

	if violations == 0 {
		_, _ = fmt.Fprintf(out, "All %d commit(s) follow %s convention.\n", len(input), conv.Name)
		return nil
	}

	// Print violations (both warn and enforce print; only enforce blocks).
	_, _ = fmt.Fprintf(out, "%d of %d commit(s) violate %s convention:\n\n",
		violations, len(input), conv.Name)

	for _, r := range results {
		if !r.Valid {
			errMsg := convention.FormatError(r, conv)
			_, _ = fmt.Fprint(out, errMsg)
			_, _ = fmt.Fprintln(out)
		}
	}

	// Thin boundary: translate the pure decision into the process exit code.
	// warn → exit 0 (non-blocking); enforce + violations → exit 2 (deny per
	// Claude Code protocol). This is the ONLY os.Exit site.
	if decideExit(action, violations) == 2 {
		os.Exit(2)
	}
	return nil
}

// resolveConventionName determines the convention name from configuration.
// Priority: MOAI_GIT_CONVENTION env var > config > default "auto".
func resolveConventionName() string {
	if envVal := os.Getenv(config.EnvGitConvention); envVal != "" {
		return envVal
	}

	if deps != nil && deps.Config != nil {
		cfg := deps.Config.Get()
		if cfg != nil && cfg.GitConvention.Convention != "" {
			return cfg.GitConvention.Convention
		}
	}

	return "auto"
}

// resolveAutoDetectOptions builds the convention.AutoDetectOptions (Fix A) and the
// forwarded max_length (Fix B) from the loaded git_convention config. When config
// is unavailable it returns the compiled-in defaults (auto-detection enabled,
// sample_size 100, threshold 0.5, fallback conventional-commits, max_length 100) so
// the pre-push path behaves identically to the prior hardcoded behavior.
func resolveAutoDetectOptions() (convention.AutoDetectOptions, int) {
	opts := convention.AutoDetectOptions{
		Enabled:             true,
		SampleSize:          config.DefaultGitConventionSampleSize,
		ConfidenceThreshold: config.DefaultGitConventionConfidenceThreshold,
		Fallback:            config.DefaultGitConventionFallback,
	}
	maxLength := config.DefaultGitConventionMaxLength

	if deps != nil && deps.Config != nil {
		if cfg := deps.Config.Get(); cfg != nil {
			ad := cfg.GitConvention.AutoDetection
			opts = convention.AutoDetectOptions{
				Enabled:             ad.Enabled,
				SampleSize:          ad.SampleSize,
				ConfidenceThreshold: ad.ConfidenceThreshold,
				Fallback:            ad.Fallback,
			}
			maxLength = cfg.GitConvention.Validation.MaxLength
		}
	}

	return opts, maxLength
}

// isEnforceOnPushEnabled checks whether convention enforcement is enabled.
// Priority: MOAI_ENFORCE_ON_PUSH env var > config > default false.
func isEnforceOnPushEnabled() bool {
	if envVal := os.Getenv(config.EnvEnforceOnPush); envVal != "" {
		return envVal == "true" || envVal == "1"
	}

	if deps != nil && deps.Config != nil {
		cfg := deps.Config.Get()
		if cfg != nil {
			return cfg.GitConvention.Validation.EnforceOnPush
		}
	}

	return false
}

// readStdinLines reads all non-empty lines from stdin.
func readStdinLines() ([]string, error) {
	data, err := os.ReadFile("/dev/stdin")
	if err != nil {
		// stdin may not be available; return empty.
		return nil, nil //nolint:nilerr // empty stdin is not an error
	}

	raw := strings.Split(strings.TrimSpace(string(data)), "\n")
	var lines []string
	for _, line := range raw {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return lines, nil
}
