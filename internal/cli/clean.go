package cli

// SPEC-V3R2-RT-004 REQ-031, AC-13: runs/ 디렉토리 보존 일수(retention_days) 기반 정리.
// 기본 동작: dry-run (실제 삭제 없음). --force 플래그로 실제 삭제 실행.

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// newCleanCmd creates the clean subcommand.
func newCleanCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean up stale run artifacts",
		Long: `Clean up run artifacts in .moai/state/runs/ that are older than retention_days.
Default: dry-run mode (no actual deletion). Use --force to actually delete.

retention_days is read from .moai/config/sections/state.yaml.`,
		GroupID: "tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClean(force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "실제로 파일을 삭제합니다 (기본값: dry-run)")

	return cmd
}

// stateYAMLWrapper는 state.yaml의 최상위 키 구조입니다.
type stateYAMLWrapper struct {
	State struct {
		RetentionDays int `yaml:"retention_days"`
	} `yaml:"state"`
}

// runClean은 retention_days를 기준으로 오래된 runs/ 디렉토리를 정리합니다.
func runClean(force bool) error {
	// 상태 디렉토리 탐색
	stateDir, err := findStateDir()
	if err != nil {
		return fmt.Errorf("find state dir: %w", err)
	}

	// retention_days 로드 (state.yaml에서)
	retentionDays, err := loadRetentionDays(stateDir)
	if err != nil {
		return fmt.Errorf("load retention_days: %w", err)
	}

	if retentionDays <= 0 {
		fmt.Println("retention_days not configured or 0; nothing to clean")
		return nil
	}

	// runs/ 디렉토리 스캔
	runsDir := filepath.Join(stateDir, "runs")
	entries, err := os.ReadDir(runsDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("runs/ directory not found at %s; nothing to clean\n", runsDir)
			return nil
		}
		return fmt.Errorf("read runs dir: %w", err)
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	var toDelete []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			toDelete = append(toDelete, filepath.Join(runsDir, entry.Name()))
		}
	}

	if len(toDelete) == 0 {
		fmt.Printf("No runs older than %d days found\n", retentionDays)
		return nil
	}

	// dry-run 또는 실제 삭제
	for _, path := range toDelete {
		if force {
			if err := os.RemoveAll(path); err != nil {
				fmt.Fprintf(os.Stderr, "WARN: failed to remove %s: %v\n", path, err)
			} else {
				fmt.Printf("Deleted: %s\n", path)
			}
		} else {
			fmt.Printf("[dry-run] Would delete: %s\n", path)
		}
	}

	if !force {
		fmt.Printf("\n%d runs eligible for deletion. Run with --force to actually delete.\n", len(toDelete))
	}

	return nil
}

// loadRetentionDays는 .moai/config/sections/state.yaml에서 retention_days를 읽습니다.
func loadRetentionDays(stateDir string) (int, error) {
	// stateDir은 .moai/state/ 이므로 .moai/config/sections/으로 이동
	moaiDir := filepath.Dir(stateDir) // .moai/
	configPath := filepath.Join(moaiDir, "config", "sections", "state.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil // state.yaml 없으면 retention_days = 0 (비활성)
		}
		return 0, fmt.Errorf("read state.yaml: %w", err)
	}

	var wrapper stateYAMLWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return 0, fmt.Errorf("parse state.yaml: %w", err)
	}

	return wrapper.State.RetentionDays, nil
}
