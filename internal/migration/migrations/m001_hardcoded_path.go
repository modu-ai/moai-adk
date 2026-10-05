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

	// Both remaining outcomes are success: the literal was substituted, or it
	// was never present (REQ-V3R2-RT-007-023 calls that already-migrated). A
	// migration reports success as nil, the way m002 does.
	//
	// These two lines used to return fmt.Errorf carrying the scanned and
	// rewritten counts, under a comment describing them as the details field.
	// They never reached the details field: the runner owns that field and
	// fills it itself — "적용 완료" on success, err.Error() only on FAILURE. So
	// a completed rewrite was written to the log as a failure and surfaced to
	// the user as "마이그레이션 1 적용 실패: m001 적용 완료". Dropping the counts
	// loses no working signal, because the only channel they ever reached was
	// the failure one.
	return nil
}
