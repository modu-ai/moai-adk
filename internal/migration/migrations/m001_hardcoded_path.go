package migrations

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/migration"
)

// @MX:NOTE - SPEC-V3R2-RT-007 REQ-024 m001 is intentionally NON-rollback-able.
// As a CRITICAL bug-fix migration, rollback attempts return a MigrationNotRollbackable error.

const (
	// hardcodedLiteral is the absolute path hardcoded in v2.x wrappers.
	hardcodedLiteral = "/Users/goos/go/bin/moai"
	// replacement is the portable path that replaces the literal.
	replacement = "$HOME/go/bin/moai"
)

// init registers m001 in the registry.
func init() {
	migration.Register(migration.Migration{
		Version:  1,
		Name:     "remove_hardcoded_gobin_path",
		Apply:    m001Apply,
		Rollback: nil, // REQ-V3R2-RT-007-024: non-rollback-able
	})
}

// m001Apply rewrites shell wrappers that contain a hardcoded path.
// REQ-V3R2-RT-007-022: substitutes /Users/goos/go/bin/moai with $HOME/go/bin/moai.
// REQ-V3R2-RT-007-023: no-op on a project that is already clean.
func m001Apply(projectRoot string) error {
	// File pattern for .claude/hooks/moai/handle-*.sh.
	pattern := filepath.Join(projectRoot, ".claude", "hooks", "moai", "handle-*.sh")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("wrapper glob 실패: %w", err)
	}

	rewrittenCount := 0
	scannedCount := 0

	for _, wrapperPath := range matches {
		scannedCount++

		// Read the file.
		content, err := os.ReadFile(wrapperPath)
		if err != nil {
			return fmt.Errorf("wrapper 읽기 실패 %s: %w", wrapperPath, err)
		}

		// Check whether the hardcoded literal is present.
		if !bytes.Contains(content, []byte(hardcodedLiteral)) {
			// Already clean (REQ-V3R2-RT-007-023).
			continue
		}

		// Preserve file mode (REQ-V3R2-RT-007-022).
		info, err := os.Stat(wrapperPath)
		if err != nil {
			return fmt.Errorf("wrapper stat 실패 %s: %w", wrapperPath, err)
		}
		mode := info.Mode()

		// Substitute the literal (bytes.ReplaceAll - exact match only).
		newContent := bytes.ReplaceAll(content, []byte(hardcodedLiteral), []byte(replacement))

		// Atomic write (temporary file + rename).
		tmpPath := wrapperPath + ".tmp"
		if err := os.WriteFile(tmpPath, newContent, mode); err != nil {
			return fmt.Errorf("wrapper 쓰기 실패 %s: %w", wrapperPath, err)
		}

		// Atomic rename.
		if err := os.Rename(tmpPath, wrapperPath); err != nil {
			return fmt.Errorf("wrapper rename 실패 %s: %w", wrapperPath, err)
		}

		rewrittenCount++
	}

	// Log message (for the details field).
	if rewrittenCount > 0 {
		return fmt.Errorf("m001 적용 완료: %d 파일 재작성됨 (scanned %d 파일)", rewrittenCount, scannedCount)
	}

	// No-op (already clean).
	return fmt.Errorf("이미 migrated됨 (scanned %d 파일, 0 재작성됨)", scannedCount)
}
