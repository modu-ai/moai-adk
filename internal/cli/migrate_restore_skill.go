// SPEC-V3R3-HARNESS-001 / T-M4-03
// migrate_restore_skill.go — moai migrate restore-skill <skill-id> 서브커맨드.
//
// BC-V3R3-007 로 제거된 16개 레거시 스킬을 아카이브에서 복원한다.
// 복원 경로: .moai/archive/skills/v2.16/<id>/ → .claude/skills/<id>/
//
// 에러 코드 (MigrateError 패턴):
//   - RESTORE_ARCHIVE_MISSING   : 아카이브 엔트리 없음
//   - RESTORE_TARGET_EXISTS     : 복원 대상이 이미 존재 (--force 필요)
//   - RESTORE_INVALID_SKILL_ID  : 경로 순회 또는 빈 skillID

package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// restoreSkill은 .moai/archive/skills/v2.16/<skillID>/ 를
// .claude/skills/<skillID>/ 로 재귀 복사한다.
//
// force=true 이면 복원 대상이 이미 존재해도 덮어쓴다.
//
// @MX:ANCHOR: [AUTO] restoreSkill은 archive → skills 복원 계약의 진입점
// @MX:REASON: [AUTO] runRestoreSkill, TestRestoreSkill_*, 라운드트립 테스트에서 fan_in >= 3
func restoreSkill(projectRoot, skillID string, force bool) error {
	// 경로 순회 방지
	if err := validateSkillID(skillID); err != nil {
		return &MigrateError{
			Code:    "RESTORE_INVALID_SKILL_ID",
			Message: err.Error(),
		}
	}
	if skillID == "" {
		return &MigrateError{
			Code:    "RESTORE_INVALID_SKILL_ID",
			Message: "skillID must not be empty",
		}
	}

	archiveDir := filepath.Join(projectRoot, ".moai", "archive", "skills", archiveVersion, skillID)
	targetDir := filepath.Join(projectRoot, ".claude", "skills", skillID)

	// 아카이브 존재 여부 확인
	if _, err := os.Stat(archiveDir); err != nil {
		if os.IsNotExist(err) {
			return &MigrateError{
				Code:    "RESTORE_ARCHIVE_MISSING",
				Message: fmt.Sprintf("no archive entry for skill %q at %s", skillID, archiveDir),
			}
		}
		return fmt.Errorf("stat archive %s: %w", skillID, err)
	}

	// 복원 대상 존재 여부 확인
	if _, err := os.Stat(targetDir); err == nil {
		if !force {
			return &MigrateError{
				Code: "RESTORE_TARGET_EXISTS",
				Message: fmt.Sprintf(
					".claude/skills/%s already exists; use --force to overwrite", skillID,
				),
			}
		}
		// --force: 기존 대상 제거 후 복원
		if err := os.RemoveAll(targetDir); err != nil {
			return fmt.Errorf("remove existing target %s: %w", skillID, err)
		}
	}

	// 대상 디렉토리 생성
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("create target directory for %s: %w", skillID, err)
	}

	// 재귀 복사 (아카이브 → 스킬 디렉토리)
	if err := copyDirAll(archiveDir, targetDir); err != nil {
		_ = os.RemoveAll(targetDir)
		return fmt.Errorf("restore skill %s: %w", skillID, err)
	}

	return nil
}

// --- Cobra command wiring ---

var restoreSkillCmd = &cobra.Command{
	Use:   "restore-skill <skill-id>",
	Short: "Restore a legacy skill from the v2.16 archive",
	Long: `Restore a skill that was archived during moai update (BC-V3R3-007).

Copies .moai/archive/skills/v2.16/<skill-id>/ back to .claude/skills/<skill-id>/.

Error codes:
  RESTORE_ARCHIVE_MISSING   – no archive entry found for the given skill-id
  RESTORE_TARGET_EXISTS     – target directory already exists (use --force)
  RESTORE_INVALID_SKILL_ID  – skill-id contains path traversal characters

Example:
  moai migrate restore-skill moai-domain-backend
  moai migrate restore-skill moai-domain-backend --force`,
	Args: cobra.ExactArgs(1),
	RunE: runRestoreSkill,
}

func init() {
	migrateCmd.AddCommand(restoreSkillCmd)
	restoreSkillCmd.Flags().Bool("force", false, "Overwrite existing target directory")
}

func runRestoreSkill(cmd *cobra.Command, args []string) error {
	skillID := args[0]
	force, _ := cmd.Flags().GetBool("force")
	out := cmd.OutOrStdout()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	if err := restoreSkill(cwd, skillID, force); err != nil {
		if me, ok := err.(*MigrateError); ok {
			_, _ = fmt.Fprintf(out, "Error [%s]: %s\n", me.Code, me.Message)
			return me
		}
		return err
	}

	dstRel := filepath.Join(".claude", "skills", skillID)
	_, _ = fmt.Fprintf(out, "restore: %s → %s\n", skillID, dstRel)
	return nil
}
