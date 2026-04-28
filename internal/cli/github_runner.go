// Package cli는 GitHub Actions runner CLI 명령을 제공합니다.
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
	// runnerInstallerFactory는 Installer를 생성합니다.
	runnerInstallerFactory = func(ghRunnerDir string) *runner.Installer {
		return runner.NewInstaller(ghRunnerDir, nil)
	}

	// runnerRegistrarFactory는 Registrar를 생성합니다.
	runnerRegistrarFactory = func(ghRunnerDir string) *runner.Registrar {
		return runner.NewRegistrar(ghRunnerDir, nil)
	}

	// runnerServiceManagerFactory는 ServiceManager를 생성합니다.
	runnerServiceManagerFactory = func(ghRunnerDir string) runner.ServiceManager {
		return runner.NewLaunchdManager(ghRunnerDir, nil)
	}

	// runnerVersionCheckerFactory는 VersionChecker를 생성합니다.
	runnerVersionCheckerFactory = func(ghRunnerDir string) *runner.VersionChecker {
		return runner.NewVersionChecker(ghRunnerDir, nil)
	}
)

// newRunnerInstallCmd는 install 서브커맨드를 생성합니다.
// newRunnerInstallCmd creates the install subcommand.
func newRunnerInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "GitHub Actions runner 다운로드 및 설치 (Download and install GitHub Actions runner)",
		Long:  `지정된 OS 및 아키텍처용 runner를 다운로드하고 압축을 해제합니다.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 기본 runner 디렉토리 (~/actions-runner)
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("get home directory: %w", err)
			}
			ghRunnerDir := filepath.Join(homeDir, "actions-runner")

			installer := runnerInstallerFactory(ghRunnerDir)

			// OS 및 아키텍처 자동 감지
			goos := "darwin" // TODO: 실제 감지 로직
			arch := "arm64"  // TODO: 실제 감지 로직

			if err := installer.DownloadRunner(cmd.Context(), goos, arch); err != nil {
				return fmt.Errorf("download runner: %w", err)
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), ghSuccessCard(
				"Runner 설치 완료 (Runner installed)",
				fmt.Sprintf("위치: %s/actions-runner", ghRunnerDir),
			))
			return nil
		},
	}

	return cmd
}

// newRunnerRegisterCmd는 register 서브커맨드를 생성합니다.
// newRunnerRegisterCmd creates the register subcommand.
func newRunnerRegisterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register <repo>",
		Short: "Runner를 GitHub에 등록 (Register runner with GitHub)",
		Long:  `지정된 리포지토리에 runner를 등록합니다. 형식: owner/repo`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo := args[0]

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("get home directory: %w", err)
			}
			ghRunnerDir := filepath.Join(homeDir, "actions-runner")

			registrar := runnerRegistrarFactory(ghRunnerDir)

			labels := []string{"self-hosted"} // TODO: 사용자 입력 라벨

			result, err := registrar.RegisterRunner(cmd.Context(), repo, labels)
			if err != nil {
				return fmt.Errorf("register runner: %w", err)
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), ghSuccessCard(
				"Runner 등록 완료 (Runner registered)",
				fmt.Sprintf("이름: %s", result.RunnerName),
				fmt.Sprintf("설정: %s", result.SettingsURL),
			))
			return nil
		},
	}

	return cmd
}

// newRunnerStartCmd는 start 서브커맨드를 생성합니다.
// newRunnerStartCmd creates the start subcommand.
func newRunnerStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Runner 서비스 시작 (Start runner service)",
		Long:  `runner 서비스를 시작합니다.`,
		Args:  cobra.NoArgs,
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
				"Runner 서비스 시작 완료 (Service started)",
			))
			return nil
		},
	}

	return cmd
}

// newRunnerStopCmd는 stop 서브커맨드를 생성합니다.
// newRunnerStopCmd creates the stop subcommand.
func newRunnerStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Runner 서비스 중지 (Stop runner service)",
		Long:  `runner 서비스를 중지합니다.`,
		Args:  cobra.NoArgs,
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
				"Runner 서비스 중지 완료 (Service stopped)",
			))
			return nil
		},
	}

	return cmd
}

// newRunnerStatusCmd는 status 서브커맨드를 생성합니다.
// newRunnerStatusCmd creates the status subcommand.
func newRunnerStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Runner 버전 및 상태 확인 (Check runner version and status)",
		Long:  `runner 버전과 서비스 상태를 확인합니다.`,
		Args:  cobra.NoArgs,
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
			details = append(details, fmt.Sprintf("설치 버전: %s", result.InstalledVersion))
			details = append(details, fmt.Sprintf("최신 버전: %s", result.LatestVersion))
			details = append(details, fmt.Sprintf("경과 일수: %d일", result.DaysOld))
			details = append(details, fmt.Sprintf("상태: %s - %s", result.Status, result.Message))

				// details를 개행으로 연결
				content := ""
				for i, d := range details {
					if i > 0 {
						content += "\n"
					}
					content += d
				}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), ghInfoCard(
				fmt.Sprintf("Runner 상태 (Runner Status: %s)", result.Status),
				content,
			))
			return nil
		},
	}

	return cmd
}

// newRunnerUpgradeCmd는 upgrade 서브커맨드를 생성합니다.
// newRunnerUpgradeCmd creates the upgrade subcommand.
func newRunnerUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade <repo>",
		Short: "Runner 업그레이드 (Upgrade runner)",
		Long:  `runner를 다운로드하고, 등록하고, 서비스를 시작합니다.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo := args[0]

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("get home directory: %w", err)
			}
			ghRunnerDir := filepath.Join(homeDir, "actions-runner")

			// 1. 다운로드 (installer)
			installer := runnerInstallerFactory(ghRunnerDir)
			goos := "darwin" // TODO
			arch := "arm64"  // TODO

			if err := installer.DownloadRunner(cmd.Context(), goos, arch); err != nil {
				return fmt.Errorf("download runner: %w", err)
			}

			// 2. 등록 (registrar)
			registrar := runnerRegistrarFactory(ghRunnerDir)
			labels := []string{"self-hosted"}

			_, err = registrar.RegisterRunner(cmd.Context(), repo, labels)
			if err != nil {
				return fmt.Errorf("register runner: %w", err)
			}

			// 3. 서비스 시작
			mgr := runnerServiceManagerFactory(ghRunnerDir)

			if err := mgr.Start(cmd.Context()); err != nil {
				return fmt.Errorf("start service: %w", err)
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), ghSuccessCard(
				"Runner 업그레이드 완료 (Runner upgraded)",
			))
			return nil
		},
	}

	return cmd
}

// newRunnerCmd는 runner 명령을 생성합니다.
func newRunnerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runner",
		Short: "Runner 관리 명령 (Manage runner lifecycle)",
		Long:  `Runner 설치, 등록, 시작, 중지, 상태 확인 등 라이프사이클 관리 명령을 제공합니다.`,
	}

	// 서브커맨드 추가
	cmd.AddCommand(newRunnerInstallCmd())
	cmd.AddCommand(newRunnerRegisterCmd())
	cmd.AddCommand(newRunnerStartCmd())
	cmd.AddCommand(newRunnerStopCmd())
	cmd.AddCommand(newRunnerStatusCmd())
	cmd.AddCommand(newRunnerUpgradeCmd())

	return cmd
}
