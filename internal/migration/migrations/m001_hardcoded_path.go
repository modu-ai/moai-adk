package migrations

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/migration"
)

// @MX:NOTE - SPEC-V3R2-RT-007 REQ-024 m001은 의도적으로 NON-rollback-able입니다.
// CRITICAL bug-fix 마이그레이션이므로 rollback 시도 시 MigrationNotRollbackable 에러를 반환합니다.

const (
	// hardcodedLiteral은 v2.x wrapper에 하드코딩된 절대 경로입니다.
	hardcodedLiteral = "/Users/goos/go/bin/moai"
	// replacement는 리터럴을 교체할 portable 경로입니다.
	replacement = "$HOME/go/bin/moai"
)

// init는 m001을 registry에 등록합니다.
func init() {
	migration.Register(migration.Migration{
		Version: 1,
		Name:    "remove_hardcoded_gobin_path",
		Apply:   m001Apply,
		Rollback: nil, // REQ-V3R2-RT-007-024: non-rollback-able
	})
}

// m001Apply는 hardcoded path를 가진 shell wrapper를 재작성합니다.
// REQ-V3R2-RT-007-022: /Users/goos/go/bin/moai를 $HOME/go/bin/moai로 치환합니다.
// REQ-V3R2-RT-007-023: 이미 깨끗한 프로젝트에서는 no-op입니다.
func m001Apply(projectRoot string) error {
	// .claude/hooks/moai/handle-*.sh 파일 패턴
	pattern := filepath.Join(projectRoot, ".claude", "hooks", "moai", "handle-*.sh")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("wrapper glob 실패: %w", err)
	}

	rewrittenCount := 0
	scannedCount := 0

	for _, wrapperPath := range matches {
		scannedCount++

		// 파일 읽기
		content, err := os.ReadFile(wrapperPath)
		if err != nil {
			return fmt.Errorf("wrapper 읽기 실패 %s: %w", wrapperPath, err)
		}

		// Hardcoded literal 존재 확인
		if !bytes.Contains(content, []byte(hardcodedLiteral)) {
			// 이미 깨끗함 (REQ-V3R2-RT-007-023)
			continue
		}

		// 파일 mode 보존 (REQ-V3R2-RT-007-022)
		info, err := os.Stat(wrapperPath)
		if err != nil {
			return fmt.Errorf("wrapper stat 실패 %s: %w", wrapperPath, err)
		}
		mode := info.Mode()

		// Literal 치환 (bytes.ReplaceAll - 정확한 일치만)
		newContent := bytes.ReplaceAll(content, []byte(hardcodedLiteral), []byte(replacement))

		// Atomic write (임시 파일 + rename)
		tmpPath := wrapperPath + ".tmp"
		if err := os.WriteFile(tmpPath, newContent, mode); err != nil {
			return fmt.Errorf("wrapper 쓰기 실패 %s: %w", wrapperPath, err)
		}

		// Atomic rename
		if err := os.Rename(tmpPath, wrapperPath); err != nil {
			return fmt.Errorf("wrapper rename 실패 %s: %w", wrapperPath, err)
		}

		rewrittenCount++
	}

	// 로그 메시지 (details 필드용)
	if rewrittenCount > 0 {
		return fmt.Errorf("m001 적용 완료: %d 파일 재작성됨 (scanned %d 파일)", rewrittenCount, scannedCount)
	}

	// No-op (이미 깨끗함)
	return fmt.Errorf("이미 migrated됨 (scanned %d 파일, 0 재작성됨)", scannedCount)
}
