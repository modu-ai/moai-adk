// Package cli provides GitHub status command.
package cli

import (
"context"
"fmt"
"os"
"path/filepath"
"strings"

"github.com/spf13/cobra"

"github.com/modu-ai/moai-adk/internal/github/runner"
)

// newStatusCmd creates the status command.
func newStatusCmd() *cobra.Command {
return &cobra.Command{
Use: "status",
Short: "GitHub Actions integration/integrated state verification/check (Check integration status)",
Long: `Display runner version, auth token, and workflow status.`,
Args: cobra.NoArgs,
RunE: runGitHubStatus,
}
}

// runGitHubStatus executes the status check.
func runGitHubStatus(cmd *cobra.Command, args []string) error {
ctx := cmd.Context()
if ctx == nil {
ctx = context.Background()
}

out := cmd.OutOrStdout()

// 1. Check runner version (T-05)
runnerStatus, err := checkRunnerVersion(ctx)
if err != nil {
return fmt.Errorf("check runner version: %w", err)
}

// 2. Display formatted status card
displayStatusCard(out, runnerStatus)

return nil
}

// checkRunnerVersion checks the runner version.
func checkRunnerVersion(ctx context.Context) (string, error) {
homeDir, err := os.UserHomeDir()
if err != nil {
return "", fmt.Errorf("get home directory: %w", err)
}
ghRunnerDir := filepath.Join(homeDir, "actions-runner")

ghClient := runner.NewFileSystemGitHubClient()
checker := runner.NewVersionChecker(ghRunnerDir, ghClient)
result, err := checker.CheckVersion(ctx)
if err != nil {
return "", err
}

var sb strings.Builder
fmt.Fprintf(&sb, "Installed version: %s\n", result.InstalledVersion)
fmt.Fprintf(&sb, "Latest version: %s\n", result.LatestVersion)
fmt.Fprintf(&sb, "Days old: %d\n", result.DaysOld)
fmt.Fprintf(&sb, "Status: %s - %s", result.Status, result.Message)

return sb.String(), nil
}

// displayStatusCard displays the status card.
func displayStatusCard(out interface{}, runnerStatus string) {
_, _ = fmt.Fprintf(out.(interface{ Write([]byte) (int, error) }),
"=== GitHub Actions Status ===\n\n"+
"[Runner]\n%s\n\n"+
"[Auth]\nToken check: Configure with moai github auth <llm> <token>\n\n"+
"[Workflow]\nTemplate check: To be implemented\n",
runnerStatus)
}
