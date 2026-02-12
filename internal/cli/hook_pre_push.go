package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

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

// runPrePush reads commit messages from stdin and validates them.
func runPrePush(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	// Check if enforcement is enabled via configuration.
	if !isEnforceOnPushEnabled() {
		return nil
	}

	// Determine repository root from CLAUDE_PROJECT_DIR or current directory.
	repoPath := os.Getenv("CLAUDE_PROJECT_DIR")
	if repoPath == "" {
		var err error
		repoPath, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("pre-push: determine working directory: %w", err)
		}
	}

	// Load convention configuration.
	convName := resolveConventionName()
	mgr := convention.NewManager(repoPath)
	if err := mgr.LoadConvention(convName); err != nil {
		return fmt.Errorf("pre-push: load convention: %w", err)
	}

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

	// Print violations.
	_, _ = fmt.Fprintf(out, "%d of %d commit(s) violate %s convention:\n\n",
		violations, len(input), conv.Name)

	for _, r := range results {
		if !r.Valid {
			errMsg := convention.FormatError(r, conv)
			_, _ = fmt.Fprint(out, errMsg)
			_, _ = fmt.Fprintln(out)
		}
	}

	// Exit with code 2 to signal deny per Claude Code protocol.
	os.Exit(2)
	return nil // unreachable
}

// resolveConventionName determines the convention name from configuration.
// Priority: MOAI_GIT_CONVENTION env var > config > default "auto".
func resolveConventionName() string {
	if envVal := os.Getenv("MOAI_GIT_CONVENTION"); envVal != "" {
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

// isEnforceOnPushEnabled checks whether convention enforcement is enabled.
// Priority: MOAI_ENFORCE_ON_PUSH env var > config > default false.
func isEnforceOnPushEnabled() bool {
	if envVal := os.Getenv("MOAI_ENFORCE_ON_PUSH"); envVal != "" {
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
