package hook

import (
	"context"
	"log/slog"

	"github.com/modu-ai/moai-adk/internal/github/runner"
)

// checkRunnerVersion checks the GitHub Actions runner version and returns
// a warning message if the runner is outdated or not found.
// This implements T-27 (Doctor Integration) and T-28 (SessionStart Hook Integration).
//
// The check is non-blocking: any errors are logged and an empty string is returned.
func checkRunnerVersion(projectDir string) string {
	ghRunnerDir := runner.DefaultRunnerDir()
	ghClient := runner.NewFileSystemGitHubClient()
	checker := runner.NewVersionChecker(ghRunnerDir, ghClient)

	result, err := checker.CheckVersion(context.Background())
	if err != nil {
		slog.Debug("runner version check failed", "error", err)
		return "" // Non-blocking: suppress errors
	}

	switch result.Status {
	case runner.VersionCheckOK:
		return "" // Runner is up-to-date
	case runner.VersionCheckWarn:
		return result.Message // Return warning message
	case runner.VersionCheckFail:
		// Critical failure: recommend runner installation
		return "GitHub Actions runner not found. Run 'moai github init' to install."
	case runner.VersionCheckSkip:
		return "" // Check was skipped (e.g., unsupported platform)
	default:
		return ""
	}
}
