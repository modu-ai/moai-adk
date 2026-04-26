// SPEC-V3R3-HARNESS-001 / T-M4-01
// update_archive.go — .claude/skills/<id>/ 를 .moai/archive/skills/v2.16/<id>/ 에
// 재귀적으로 복사하는 archiveSkill 함수를 제공한다.
//
// 멱등성 보장:
//   - 소스 absent → 즉시 nil 반환
//   - 아카이브 이미 존재 + 내용 동일 → nil 반환
//   - 아카이브 이미 존재 + 내용 불일치(drift) → ARCHIVE_DRIFT 에러
//
// 경로 순회 보호:
//   - skillID에 ".." 또는 "/" 포함 시 에러 반환

package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// archiveVersion은 아카이브 디렉토리에 사용되는 버전 태그이다.
const archiveVersion = "v2.16"

// legacySkillIDs는 BC-V3R3-007에서 제거된 16개 스킬 ID 목록이다.
// moai update 실행 시 이 스킬들은 .moai/archive/skills/v2.16/ 으로 이동된다.
var legacySkillIDs = []string{
	"moai-domain-backend",
	"moai-domain-frontend",
	"moai-domain-database",
	"moai-domain-db-docs",
	"moai-domain-mobile",
	"moai-framework-electron",
	"moai-library-shadcn",
	"moai-library-mermaid",
	"moai-library-nextra",
	"moai-tool-ast-grep",
	"moai-platform-auth",
	"moai-platform-deployment",
	"moai-platform-chrome-extension",
	"moai-workflow-research",
	"moai-workflow-pencil-integration",
	"moai-formats-data",
}

// archiveSkill은 projectRoot/.claude/skills/<skillID>/ 를
// projectRoot/.moai/archive/skills/v2.16/<skillID>/ 에 복사한다.
//
// @MX:ANCHOR: [AUTO] archiveSkill은 레거시 스킬 아카이브 계약의 진입점
// @MX:REASON: [AUTO] runUpdate 흐름, restore-skill, 멱등성 테스트 등 fan_in >= 3
func archiveSkill(projectRoot, skillID string) error {
	// 경로 순회 공격 방지: skillID에 ".." 또는 "/"(구분자) 포함 금지
	if err := validateSkillID(skillID); err != nil {
		return err
	}

	srcDir := filepath.Join(projectRoot, ".claude", "skills", skillID)

	// 소스가 없으면 멱등하게 nil 반환 (이미 제거된 경우 처리)
	if _, err := os.Stat(srcDir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat source skill %s: %w", skillID, err)
	}

	dstDir := filepath.Join(projectRoot, ".moai", "archive", "skills", archiveVersion, skillID)

	// 아카이브가 이미 존재하는 경우 drift 검사
	if _, err := os.Stat(dstDir); err == nil {
		if err := checkArchiveDrift(srcDir, dstDir); err != nil {
			return err
		}
		// 내용 동일 → 멱등하게 성공
		return nil
	}

	// 아카이브 대상 디렉토리 생성
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return fmt.Errorf("create archive directory for %s: %w", skillID, err)
	}

	// 재귀 복사
	if err := copyDirAll(srcDir, dstDir); err != nil {
		// 실패 시 불완전한 아카이브 정리
		_ = os.RemoveAll(dstDir)
		return fmt.Errorf("copy skill %s to archive: %w", skillID, err)
	}

	return nil
}

// validateSkillID는 skillID에 경로 순회 문자("..", "/", 절대 경로)가
// 포함되어 있는지 검사한다.
func validateSkillID(skillID string) error {
	// 절대 경로 거부
	if filepath.IsAbs(skillID) {
		return &MigrateError{
			Code:    "ARCHIVE_INVALID_ID",
			Message: fmt.Sprintf("skillID must be a simple name, not an absolute path: %q", skillID),
		}
	}
	// ".." 포함 거부
	if strings.Contains(skillID, "..") {
		return &MigrateError{
			Code:    "ARCHIVE_INVALID_ID",
			Message: fmt.Sprintf("skillID must not contain '..': %q", skillID),
		}
	}
	// "/" 포함 거부 (플랫폼 구분자)
	if strings.ContainsAny(skillID, "/\\") {
		return &MigrateError{
			Code:    "ARCHIVE_INVALID_ID",
			Message: fmt.Sprintf("skillID must not contain path separators: %q", skillID),
		}
	}
	return nil
}

// checkArchiveDrift는 소스 디렉토리와 기존 아카이브 디렉토리의 내용이
// 동일한지 SHA-256으로 비교한다.
// 동일하면 nil, 다르면 ARCHIVE_DRIFT 에러를 반환한다.
func checkArchiveDrift(srcDir, dstDir string) error {
	srcHashes, err := computeDirHashes(srcDir)
	if err != nil {
		return fmt.Errorf("compute source hashes: %w", err)
	}
	dstHashes, err := computeDirHashes(dstDir)
	if err != nil {
		return fmt.Errorf("compute archive hashes: %w", err)
	}

	if len(srcHashes) != len(dstHashes) {
		return &MigrateError{
			Code: "ARCHIVE_DRIFT",
			Message: fmt.Sprintf(
				"archive already exists but file count differs (src=%d, dst=%d). "+
					"Use --force to overwrite.",
				len(srcHashes), len(dstHashes),
			),
		}
	}

	for rel, srcHash := range srcHashes {
		dstHash, ok := dstHashes[rel]
		if !ok || srcHash != dstHash {
			return &MigrateError{
				Code: "ARCHIVE_DRIFT",
				Message: fmt.Sprintf(
					"archive already exists but content differs for %s. "+
						"Use --force to overwrite.",
					rel,
				),
			}
		}
	}

	return nil
}

// computeDirHashes는 디렉토리 내 모든 파일의 SHA-256 해시를
// 경로→해시 맵으로 반환한다.
// 기존 hashFile(design_folder.go) 함수를 재사용한다.
func computeDirHashes(dir string) (map[string]string, error) {
	hashes := make(map[string]string)
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		// hashFile은 design_folder.go 에서 선언, []byte 반환
		rawHash, err := hashFile(path)
		if err != nil {
			return err
		}
		hashes[rel] = fmt.Sprintf("%x", rawHash)
		return nil
	})
	return hashes, err
}

// copyDirAll은 srcDir의 모든 내용을 dstDir에 재귀적으로 복사한다.
// 각 파일의 Unix 권한을 보존한다.
func copyDirAll(srcDir, dstDir string) error {
	return filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dstDir, rel)

		if d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			return os.MkdirAll(dstPath, info.Mode().Perm())
		}

		return copyFile(path, dstPath)
	})
}

// archiveLegacySkills는 프로젝트 루트에서 legacySkillIDs 목록의 스킬을
// 순회하며 archiveSkill을 호출한다. 새로 아카이브된 스킬 수를 반환한다.
//
// 멱등성: 아카이브가 이미 존재하고 내용이 동일한 경우 카운트하지 않는다.
//
// 출력 형식:
//
//	archive: <id> → .moai/archive/skills/v2.16/<id>
//	total: N skills archived, 0 user customizations modified
//
// @MX:ANCHOR: [AUTO] archiveLegacySkills는 update 흐름의 레거시 스킬 아카이브 진입점
// @MX:REASON: [AUTO] runUpdate, dry-run, idempotency 테스트에서 fan_in >= 3
func archiveLegacySkills(projectRoot string, out io.Writer) (int, error) {
	archived := 0
	for _, id := range legacySkillIDs {
		srcDir := filepath.Join(projectRoot, ".claude", "skills", id)
		if _, err := os.Stat(srcDir); err != nil {
			// 소스 없음 → 스킵
			continue
		}

		// 아카이브가 이미 존재하는지 먼저 확인 (멱등성 검사)
		dstDir := filepath.Join(projectRoot, ".moai", "archive", "skills", archiveVersion, id)
		alreadyArchived := false
		if _, err := os.Stat(dstDir); err == nil {
			alreadyArchived = true
		}

		if err := archiveSkill(projectRoot, id); err != nil {
			return archived, fmt.Errorf("archive %s: %w", id, err)
		}

		// 이미 존재했던 경우는 카운트하지 않음 (멱등 실행)
		if alreadyArchived {
			continue
		}

		archiveDst := filepath.Join(".moai", "archive", "skills", archiveVersion, id)
		_, _ = fmt.Fprintf(out, "archive: %s → %s\n", id, archiveDst)
		archived++
	}

	_, _ = fmt.Fprintf(out, "total: %d skills archived, 0 user customizations modified\n", archived)
	return archived, nil
}

// dryRunArchiveLegacySkills는 --dry-run 모드로 실행되며
// 실제 파일시스템 변경 없이 계획된 작업을 출력한다.
func dryRunArchiveLegacySkills(projectRoot string, out io.Writer) error {
	planned := 0
	for _, id := range legacySkillIDs {
		srcDir := filepath.Join(projectRoot, ".claude", "skills", id)
		if _, err := os.Stat(srcDir); err != nil {
			continue
		}
		archiveDst := filepath.Join(".moai", "archive", "skills", archiveVersion, id)
		_, _ = fmt.Fprintf(out, "[dry-run] archive: %s → %s\n", id, archiveDst)
		planned++
	}
	_, _ = fmt.Fprintf(out, "[dry-run] total: %d skills archived, 0 user customizations modified\n", planned)
	return nil
}

// copyFile은 단일 파일을 소스에서 대상으로 복사한다.
// 소스의 권한 비트를 보존한다.
func copyFile(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat %s: %w", src, err)
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode().Perm())
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy %s → %s: %w", src, dst, err)
	}

	return nil
}
