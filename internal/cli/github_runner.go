// Package cli provides GitHub Actions runner CLI commands.
package cli

import (
"fmt"
"os"
"path/filepath"

"github.com/spf13/cobra"

"github.com/modu-ai/moai-adk/internal/github/runner"
)

// Factory functions for dependency injection.
// Tests replace these with mocks.

var (
// runnerInstallerFactory creates an Installer.
runnerInstallerFactory = func(ghRunnerDir string) *runner.Installer {
return runner.NewInstaller(ghRunnerDir, nil)
}

// runnerRegistrarFactory creates a Registrar.
runnerRegistrarFactory = func(ghRunnerDir string) *runner.Registrar {
return runner.NewRegistrar(ghRunnerDir, nil)
}

// runnerServiceManagerFactory creates a ServiceManager.
runnerServiceManagerFactory = func(ghRunnerDir string) runner.ServiceManager {
return runner.NewLaunchdManager(ghRunnerDir, nil)
}

// runnerVersionCheckerFactory creates a VersionChecker.
runnerVersionCheckerFactory = func(ghRunnerDir string) *runner.VersionChecker {
return runner.NewVersionChecker(ghRunnerDir, nil)
}
)

// newRunnerInstallCmd creates the install subcommand.
func newRunnerInstallCmd() *cobra.Command {
cmd := &cobra.Command{
Use: "install",
Short: "GitHub Actions runner Download and install GitHub Actions runner",
Long: `Download and extract the runner for the specified OS and architecture.`,
Args: cobra.NoArgs,
RunE: func(cmd *cobra.Command, args []string) error {
// Default runner directory (~/actions-runner)
homeDir, err := os.UserHomeDir()
if err != nil {
return fmt.Errorf("get home directory: %w", err)
}
ghRunnerDir := filepath.Join(homeDir, "actions-runner")

installer := runnerInstallerFactory(ghRunnerDir)

// Auto-detect OS and architecture
goos := "darwin" // TODO: implement actual detection
arch := "arm64" // TODO: implement actual detection

if err := installer.DownloadRunner(cmd.Context(), goos, arch); err != nil {
return fmt.Errorf("download runner: %w", err)
}

_, _ = fmt.Fprintln(cmd.OutOrStdout(), ghSuccessCard(
"Runner installed",
fmt.Sprintf("Location: %s/actions-runner", ghRunnerDir),
))
return nil
},
}

return cmd
}

// newRunnerRegisterCmd creates the register subcommand.
// newRunnerRegisterCmd creates the register subcommand.
func newRunnerRegisterCmd() *cobra.Command {
cmd := &cobra.Command{
Use: "register <repo>",
Short: "Register runner with GitHub",
Long: `Register the runner with the specified repository. Format: owner/repo`,
Args: cobra.ExactArgs(1),
RunE: func(cmd *cobra.Command, args []string) error {
repo := args[0]

homeDir, err := os.UserHomeDir()
if err != nil {
return fmt.Errorf("get home directory: %w", err)
}
ghRunnerDir := filepath.Join(homeDir, "actions-runner")

registrar := runnerRegistrarFactory(ghRunnerDir)

labels := []string{"self-hosted"} // TODO: user input labels

result, err := registrar.RegisterRunner(cmd.Context(), repo, labels)
if err != nil {
return fmt.Errorf("register runner: %w", err)
}

_, _ = fmt.Fprintln(cmd.OutOrStdout(), ghSuccessCard(
"Runner registered",
fmt.Sprintf("name: %s", result.RunnerName),
fmt.Sprintf("configuration: %s", result.SettingsURL),
))
return nil
},
}

return cmd
}

// newRunnerStartCmd creates the start subcommand.
// newRunnerStartCmd creates the start subcommand.
func newRunnerStartCmd() *cobra.Command {
cmd := &cobra.Command{
Use: "start",
Short: "Start runner service",
Long: `Start the runner service.`,
Args: cobra.NoArgs,
RunE: func(cmd *cobra.Command, args []string) error {
homeDir, err := os.UserHomeDir()
if err != nil {
return fmt.Errorf("get home directory: %w", err)
}
ghRunnerDir := filepath.Join(homeDir, "actions-runner")

mgr := runnerServiceManagerFactory(ghRunnerDir)

if err := mgr.Start(cmd.Context()); err != nil {
return fmt.Errorf("start service: %w", err)
}

_, _ = fmt.Fprintln(cmd.OutOrStdout(), ghSuccessCard(
"Service started",
))
return nil
},
}

return cmd
}

// newRunnerStopCmd creates the stop subcommand.
// newRunnerStopCmd creates the stop subcommand.
func newRunnerStopCmd() *cobra.Command {
cmd := &cobra.Command{
Use: "stop",
Short: "Stop runner service",
Long: `Stop the runner service.`,
Args: cobra.NoArgs,
RunE: func(cmd *cobra.Command, args []string) error {
homeDir, err := os.UserHomeDir()
if err != nil {
return fmt.Errorf("get home directory: %w", err)
}
ghRunnerDir := filepath.Join(homeDir, "actions-runner")

mgr := runnerServiceManagerFactory(ghRunnerDir)

if err := mgr.Stop(cmd.Context()); err != nil {
return fmt.Errorf("stop service: %w", err)
}

_, _ = fmt.Fprintln(cmd.OutOrStdout(), ghSuccessCard(
"Service stopped",
))
return nil
},
}

return cmd
}

// newRunnerStatusCmd creates the status subcommand.
// newRunnerStatusCmd creates the status subcommand.
func newRunnerStatusCmd() *cobra.Command {
cmd := &cobra.Command{
Use: "status",
Short: "Check runner version and status",
Long: `Check runner version and service status.`,
Args: cobra.NoArgs,
RunE: func(cmd *cobra.Command, args []string) error {
homeDir, err := os.UserHomeDir()
if err != nil {
return fmt.Errorf("get home directory: %w", err)
}
ghRunnerDir := filepath.Join(homeDir, "actions-runner")

checker := runnerVersionCheckerFactory(ghRunnerDir)

result, err := checker.CheckVersion(cmd.Context())
if err != nil {
return fmt.Errorf("check version: %w", err)
}

var details []string
details = append(details, fmt.Sprintf("installed version: %s", result.InstalledVersion))
details = append(details, fmt.Sprintf("latest version: %s", result.LatestVersion))
details = append(details, fmt.Sprintf("days old: %d", result.DaysOld))
details = append(details, fmt.Sprintf("state: %s - %s", result.Status, result.Message))

// Connect details with newlines
content := ""
for i, d := range details {
if i > 0 {
content += "\n"
}
content += d
}

_, _ = fmt.Fprintln(cmd.OutOrStdout(), ghInfoCard(
fmt.Sprintf("Runner Status: %s)", result.Status),
content,
))
return nil
},
}

return cmd
}

// newRunnerUpgradeCmd creates the upgrade subcommand.
// newRunnerUpgradeCmd creates the upgrade subcommand.
func newRunnerUpgradeCmd() *cobra.Command {
cmd := &cobra.Command{
Use: "upgrade <repo>",
Short: "Upgrade runner",
Long: `Download, register, and start the runner service.`,
Args: cobra.ExactArgs(1),
RunE: func(cmd *cobra.Command, args []string) error {
repo := args[0]

homeDir, err := os.UserHomeDir()
if err != nil {
return fmt.Errorf("get home directory: %w", err)
}
ghRunnerDir := filepath.Join(homeDir, "actions-runner")

// 1. Download (installer)
installer := runnerInstallerFactory(ghRunnerDir)
goos := "darwin" // TODO
arch := "arm64" // TODO

if err := installer.DownloadRunner(cmd.Context(), goos, arch); err != nil {
return fmt.Errorf("download runner: %w", err)
}

// 2. Register (registrar)
registrar := runnerRegistrarFactory(ghRunnerDir)
labels := []string{"self-hosted"}

_, err = registrar.RegisterRunner(cmd.Context(), repo, labels)
if err != nil {
return fmt.Errorf("register runner: %w", err)
}

// 3. Start service
mgr := runnerServiceManagerFactory(ghRunnerDir)

if err := mgr.Start(cmd.Context()); err != nil {
return fmt.Errorf("start service: %w", err)
}

_, _ = fmt.Fprintln(cmd.OutOrStdout(), ghSuccessCard(
"Runner upgraded",
))
return nil
},
}

return cmd
}

// newRunnerCmd creates the runner command.
func newRunnerCmd() *cobra.Command {
cmd := &cobra.Command{
Use: "runner",
Short: "Manage runner lifecycle",
Long: `Provides commands for runner installation, registration, start, stop, and status check.`,
}

// Add subcommands
cmd.AddCommand(newRunnerInstallCmd())
cmd.AddCommand(newRunnerRegisterCmd())
cmd.AddCommand(newRunnerStartCmd())
cmd.AddCommand(newRunnerStopCmd())
cmd.AddCommand(newRunnerStatusCmd())
cmd.AddCommand(newRunnerUpgradeCmd())

return cmd
}
