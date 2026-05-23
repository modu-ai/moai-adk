// Package handoff implements SPEC-V3R6-SESSION-HANDOFF-AUTO-001 session-end paste-ready resume persistence.
//
// The package provides a best-effort persistence helper that the SessionEnd hook
// invokes after its existing cleanup steps. When the orchestrator has emitted a
// paste-ready resume message during the session and left a pending file at
//
//	<projectDir>/.moai/state/session-handoff/pending.md
//
// PersistIfPending reads that file, validates frontmatter (sprint/spec/status/
// index_line + optional supersedes), writes the verbatim Markdown body to
// <memoryDir>/project_<sprint>_<spec>_<status>.md via atomic rename, prepends
// the index_line to <memoryDir>/MEMORY.md (with optional supersede marker), and
// removes the pending file on success.
//
// All failures are best-effort: slog.Warn is emitted with the prefix
// "session_end: handoff: " and the function always returns nil. The package
// does not invoke AskUserQuestion or write to stdout/stderr in any
// user-visible way (REQ-SHA-009 / AC-SHA-009).
package handoff

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// fieldFormatRegex enforces REQ-SHA-011: sprint and spec field values must
// match `^[a-z0-9_-]+$`. This prevents path-injection via crafted frontmatter
// (e.g., `spec: ../../etc/passwd`) because the field is concatenated into
// the memory file name via filepath.Join.
var fieldFormatRegex = regexp.MustCompile(`^[a-z0-9_-]+$`)

// memoryMdRetryLimit is the maximum number of read-modify-write retries on
// MEMORY.md update before falling back to a slog.Warn (REQ-SHA-007).
const memoryMdRetryLimit = 3

// pendingEntry is the parsed representation of a pending resume file.
type pendingEntry struct {
	Sprint     string `yaml:"sprint"`
	Spec       string `yaml:"spec"`
	Status     string `yaml:"status"`
	Supersedes string `yaml:"supersedes"`
	IndexLine  string `yaml:"index_line"`
	// Body is the Markdown body after the frontmatter (verbatim, including
	// the `## Next Session Entry Point` heading and fenced text block).
	Body string `yaml:"-"`
}

// errStructuralDefect is returned when frontmatter is valid but the body is
// missing the required heading or fenced text block (REQ-SHA-005 path).
var errStructuralDefect = errors.New("structural defect")

// PersistIfPending reads <projectDir>/.moai/state/session-handoff/pending.md,
// validates and persists it to <memoryDir>/project_<sprint>_<spec>_<status>.md
// plus prepends <index_line> to <memoryDir>/MEMORY.md, then removes the pending
// file. All failure paths are log-only (slog.Warn) and return nil per the
// best-effort contract documented at the package level.
//
// Absent pending file is a no-op: returns nil without slog records and without
// creating the pending directory (REQ-SHA-002).
//
// The ctx parameter is reserved for future use (e.g., to honor session-end
// timeout). Currently only file I/O is performed.
//
// sessionID is reserved for future use (e.g., logging context tags).
//
// projectDir is the absolute path to the project root (typically input.CWD
// or input.ProjectDir from the SessionEnd hook).
//
// memoryDir is the absolute path to the Claude Code memory directory for
// this project (resolved by the caller; see session_end.go resolveMemoryDir).
// If the directory does not exist the helper logs warn and returns nil; it
// MUST NOT create the directory because the project-hash directory is owned
// by Claude Code (§B.2 "Project-hash directory creation").
//
// @MX:ANCHOR: [AUTO] SessionEnd hook 호출 단일 진입점 (public API). 6단계 파이프라인: 1) pending.md 검출 (REQ-SHA-002 absent=no-op) → 2) parsePending 검증 (REQ-SHA-004/005/011) → 3) memoryDir 존재 확인 (§B.2 directory 생성 금지) → 4) atomicWriteFile (REQ-SHA-006) → 5) prependToMemoryMD (REQ-SHA-007/008) → 6) pending 제거 (REQ-SHA-010).
// @MX:REASON: REQ-SHA-001~010 통합 계약 위치 + SessionEnd hook으로부터의 유일한 진입점. 모든 실패 경로 best-effort (slog.Warn + return nil, 사용자 가시적 출력 0, REQ-SHA-009). 시그니처/단계 순서 변경 시 SessionEnd 통합 + AC-SHA-001~010 전체 회귀 위험. ctx + sessionID는 향후 timeout/로깅 컨텍스트 확장 예약 (현재 미사용).
// @MX:SPEC: SPEC-V3R6-SESSION-HANDOFF-AUTO-001
func PersistIfPending(ctx context.Context, sessionID, projectDir, memoryDir string) error {
	_ = ctx
	_ = sessionID

	pendingPath := pendingFilePath(projectDir)

	// REQ-SHA-002: absent pending file is a no-op.
	pendingBytes, err := os.ReadFile(pendingPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		slog.Warn("session_end: handoff: could not read pending file",
			"path", pendingPath,
			"error", err,
		)
		return nil
	}

	// REQ-SHA-004 / REQ-SHA-005: validate frontmatter + body structure.
	entry, err := parsePending(pendingBytes)
	if err != nil {
		reason := "malformed_frontmatter"
		if errors.Is(err, errStructuralDefect) {
			reason = "structural_defect"
		}
		slog.Warn("session_end: handoff: pending file invalid; preserving for inspection",
			"path", pendingPath,
			"reason", reason,
			"error", err,
		)
		return nil
	}

	// Memory directory must exist; the hook MUST NOT create it (§B.2).
	if info, statErr := os.Stat(memoryDir); statErr != nil || !info.IsDir() {
		slog.Warn("session_end: handoff: memory directory unavailable; skipping persistence",
			"memory_dir", memoryDir,
			"error", statErr,
		)
		return nil
	}

	// REQ-SHA-006: atomic write of memory file via os.CreateTemp + os.Rename.
	memoryFileName := fmt.Sprintf("project_%s_%s_%s.md", entry.Sprint, entry.Spec, entry.Status)
	if err := atomicWriteFile(memoryDir, memoryFileName, []byte(entry.Body), 0o644); err != nil {
		slog.Warn("session_end: handoff: could not write memory file",
			"path", filepath.Join(memoryDir, memoryFileName),
			"error", err,
		)
		return nil
	}

	// REQ-SHA-007 + REQ-SHA-008: prepend MEMORY.md with retry + supersede marker.
	if err := prependToMemoryMD(memoryDir, entry.IndexLine, entry.Supersedes, memoryFileName); err != nil {
		slog.Warn("session_end: handoff: could not update MEMORY.md; memory file kept",
			"memory_dir", memoryDir,
			"partial", true,
			"error", err,
		)
		// Per §E.2 recommendation: partial success leaves pending in place so a
		// subsequent session-end can retry the MEMORY.md update. Return nil to
		// preserve best-effort contract.
		return nil
	}

	// REQ-SHA-010: on full success, remove pending file.
	if err := os.Remove(pendingPath); err != nil {
		slog.Warn("session_end: handoff: persistence succeeded but pending file removal failed",
			"path", pendingPath,
			"error", err,
		)
		return nil
	}

	slog.Info("session_end: handoff: paste-ready resume persisted",
		"memory_file", memoryFileName,
		"supersedes", entry.Supersedes,
	)
	return nil
}

// pendingFilePath returns the canonical pending-resume file location per
// REQ-SHA-001. The path is `<projectDir>/.moai/state/session-handoff/pending.md`
// and is the only path read by PersistIfPending.
func pendingFilePath(projectDir string) string {
	return filepath.Join(projectDir, ".moai", "state", "session-handoff", "pending.md")
}

// parsePending parses a pending resume file (YAML frontmatter + Markdown body).
// Returns errStructuralDefect when frontmatter is valid but body is missing the
// `## Next Session Entry Point` heading or the fenced ```text block. Returns
// other errors for malformed/missing frontmatter or invalid field formats.
//
// REQ-SHA-004 (malformed-frontmatter path):
//   - Missing or empty required field (sprint/spec/status/index_line)
//   - Unparseable YAML
//   - REQ-SHA-011: sprint/spec failing the `^[a-z0-9_-]+$` regex
//
// REQ-SHA-005 (structural-defect path):
//   - Missing `## Next Session Entry Point` heading
//   - Missing fenced ```text block in the body
//
// @MX:ANCHOR: [AUTO] frontmatter + body 구조 검증 단일 진실 공급원. REQ-SHA-004 (malformed) + REQ-SHA-005 (structural-defect) + REQ-SHA-011 (path-injection guard) 통합 게이트. PersistIfPending에서 유일하게 호출 (fan_in=1, 본 함수가 검증 책임 전체 보유).
// @MX:REASON: sprint/spec/status 필드가 fmt.Sprintf("project_%s_%s_%s.md")를 통해 메모리 파일명에 직접 사용되므로 정규식 `^[a-z0-9_-]+$` 위반 시 경로 조작 공격 가능 (예: `spec: ../../etc/passwd`). 본 함수가 유일한 검증 게이트 — 우회 경로 추가 금지. errStructuralDefect 반환 시 reason="structural_defect"로 분류되어 caller가 동등 처리.
// @MX:SPEC: SPEC-V3R6-SESSION-HANDOFF-AUTO-001
func parsePending(data []byte) (*pendingEntry, error) {
	// Frontmatter delimiter: leading `---\n` ... `\n---\n` (or `\n---\r\n`).
	if !bytes.HasPrefix(data, []byte("---\n")) && !bytes.HasPrefix(data, []byte("---\r\n")) {
		return nil, fmt.Errorf("pending file missing leading frontmatter delimiter")
	}

	// Strip leading delimiter line.
	rest := data[len("---"):]
	for len(rest) > 0 && (rest[0] == '\r' || rest[0] == '\n') {
		rest = rest[1:]
	}

	// Locate the closing delimiter line. Pattern: "\n---\n" or "\n---\r\n" or
	// "\n---" followed by EOF. Use bytes.Index for the canonical "\n---\n" form
	// and fall back to a line scan for CRLF.
	closeIdx := bytes.Index(rest, []byte("\n---\n"))
	closeLen := len("\n---\n")
	if closeIdx < 0 {
		closeIdx = bytes.Index(rest, []byte("\n---\r\n"))
		closeLen = len("\n---\r\n")
	}
	if closeIdx < 0 {
		return nil, fmt.Errorf("pending file missing closing frontmatter delimiter")
	}

	yamlBytes := rest[:closeIdx]
	body := string(rest[closeIdx+closeLen:])

	var entry pendingEntry
	if err := yaml.Unmarshal(yamlBytes, &entry); err != nil {
		return nil, fmt.Errorf("yaml unmarshal: %w", err)
	}

	// REQ-SHA-004: required-field check.
	if strings.TrimSpace(entry.Sprint) == "" {
		return nil, fmt.Errorf("required field missing or empty: sprint")
	}
	if strings.TrimSpace(entry.Spec) == "" {
		return nil, fmt.Errorf("required field missing or empty: spec")
	}
	if strings.TrimSpace(entry.Status) == "" {
		return nil, fmt.Errorf("required field missing or empty: status")
	}
	if strings.TrimSpace(entry.IndexLine) == "" {
		return nil, fmt.Errorf("required field missing or empty: index_line")
	}

	// REQ-SHA-011: sprint/spec field-format check (path-injection guard).
	if !fieldFormatRegex.MatchString(entry.Sprint) {
		return nil, fmt.Errorf("invalid field format: sprint=%q must match %s", entry.Sprint, fieldFormatRegex.String())
	}
	if !fieldFormatRegex.MatchString(entry.Spec) {
		return nil, fmt.Errorf("invalid field format: spec=%q must match %s", entry.Spec, fieldFormatRegex.String())
	}
	// Status also flows into the file name; apply the same path-safety regex.
	if !fieldFormatRegex.MatchString(entry.Status) {
		return nil, fmt.Errorf("invalid field format: status=%q must match %s", entry.Status, fieldFormatRegex.String())
	}

	// REQ-SHA-005: body structural validation.
	headingIdx := strings.Index(body, "## Next Session Entry Point")
	if headingIdx < 0 {
		return nil, fmt.Errorf("%w: missing `## Next Session Entry Point` heading", errStructuralDefect)
	}
	bodyFromHeading := body[headingIdx:]
	if !strings.Contains(bodyFromHeading, "```text") {
		return nil, fmt.Errorf("%w: missing fenced ```text block after heading", errStructuralDefect)
	}

	entry.Body = body
	return &entry, nil
}

// atomicWriteFile writes data to <dir>/<baseName> using os.CreateTemp +
// os.Rename so that concurrent readers either observe the file as absent or
// as complete (REQ-SHA-006). The temp file is created in the same directory
// as the target to keep the rename atomic within a single filesystem.
//
// On any error after temp creation the temp file is best-effort removed.
//
// @MX:ANCHOR: [AUTO] atomic 쓰기 계약 (CreateTemp → Write → Chmod → Sync → Close → Rename). 동시 읽기자에게 부분 쓰기를 노출하지 않음 (REQ-SHA-006). PersistIfPending + prependToMemoryMD 양쪽 호출 (fan_in=2 + 잠재 외부 호출자).
// @MX:REASON: tmp 파일은 target과 동일 디렉토리에 생성되어야 rename이 atomic (단일 파일시스템 경계). 다른 디렉토리로 변경 시 cross-device link 에러 + atomicity 보장 깨짐. perm 인자 명시로 umask 회피. Sync 호출은 crash safety (POSIX fsync 등가).
// @MX:SPEC: SPEC-V3R6-SESSION-HANDOFF-AUTO-001
func atomicWriteFile(dir, baseName string, data []byte, perm os.FileMode) error {
	tmp, err := os.CreateTemp(dir, baseName+".tmp.*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpPath := tmp.Name()
	cleanup := func() {
		_ = os.Remove(tmpPath)
	}

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("chmod temp: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("sync temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temp: %w", err)
	}

	targetPath := filepath.Join(dir, baseName)
	if err := os.Rename(tmpPath, targetPath); err != nil {
		cleanup()
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

// prependToMemoryMD reads <memoryDir>/MEMORY.md, optionally applies the
// supersede marker per REQ-SHA-008, and prepends `indexLine` followed by a
// newline. The read-modify-write loop retries up to memoryMdRetryLimit times
// when the file changes between read and write (mtime or content drift,
// REQ-SHA-007). When MEMORY.md does not exist the function creates it with the
// index line as the sole content (still via atomic rename for crash safety).
//
// supersedesFileName references a prior memory entry whose MEMORY.md line
// should be prefixed with `[SUPERSEDED by <newFileName>] `. When the supersedes
// field is empty or the matching line is not found the supersede step is
// silently skipped (per §A.3 spec — supersede is opportunistic).
//
// @MX:NOTE: [AUTO] read-modify-write 루프 최대 3회 재시도 (memoryMdRetryLimit). mtime + size 양쪽 비교로 read와 write 사이 동시 수정 감지. drift 감지 시 재시도, 한계 초과 시 contention 에러 반환 (slog.Warn 처리는 caller PersistIfPending). MEMORY.md 부재 시 신규 생성 (atomic rename 보장).
// @MX:SPEC: SPEC-V3R6-SESSION-HANDOFF-AUTO-001
func prependToMemoryMD(memoryDir, indexLine, supersedesFileName, newFileName string) error {
	memoryPath := filepath.Join(memoryDir, "MEMORY.md")

	for attempt := 0; attempt < memoryMdRetryLimit; attempt++ {
		// Snapshot mtime for change detection.
		var preInfo os.FileInfo
		if info, statErr := os.Stat(memoryPath); statErr == nil {
			preInfo = info
		}

		var existing []byte
		if preInfo != nil {
			var readErr error
			existing, readErr = os.ReadFile(memoryPath)
			if readErr != nil {
				return fmt.Errorf("read MEMORY.md: %w", readErr)
			}
		}

		// REQ-SHA-008: apply supersede marker on the FIRST matching line.
		if supersedesFileName != "" && len(existing) > 0 {
			existing = applySupersedeMarker(existing, supersedesFileName, newFileName)
		}

		// Prepend the new index line with a trailing newline.
		var buf bytes.Buffer
		buf.WriteString(indexLine)
		if !strings.HasSuffix(indexLine, "\n") {
			buf.WriteByte('\n')
		}
		buf.Write(existing)

		// Re-check mtime to detect concurrent modification between read and write.
		if preInfo != nil {
			postInfo, statErr := os.Stat(memoryPath)
			if statErr != nil {
				return fmt.Errorf("re-stat MEMORY.md: %w", statErr)
			}
			if !postInfo.ModTime().Equal(preInfo.ModTime()) || postInfo.Size() != preInfo.Size() {
				// Drift detected; retry.
				continue
			}
		}

		if err := atomicWriteFile(memoryDir, "MEMORY.md", buf.Bytes(), 0o644); err != nil {
			return fmt.Errorf("atomic write MEMORY.md: %w", err)
		}
		return nil
	}

	return fmt.Errorf("MEMORY.md contention: %d retries exhausted", memoryMdRetryLimit)
}

// applySupersedeMarker rewrites the FIRST line of `content` that references
// `supersedesFileName` to be prefixed with `[SUPERSEDED by <newFileName>] `.
// When no match is found the content is returned unchanged.
//
// @MX:NOTE: [AUTO] FIRST 매치 라인만 supersede marker 적용 후 즉시 반환 (§A.3 opportunistic 정책 — 다중 매치 시에도 첫 줄만 처리). 중복 marker 방지를 위해 라인 prefix `[SUPERSEDED ` 검사 후 skip. 미매치 시 content 변경 없이 그대로 반환.
// @MX:SPEC: SPEC-V3R6-SESSION-HANDOFF-AUTO-001
func applySupersedeMarker(content []byte, supersedesFileName, newFileName string) []byte {
	lines := bytes.Split(content, []byte("\n"))
	marker := []byte(fmt.Sprintf("[SUPERSEDED by %s] ", newFileName))
	for i, line := range lines {
		if bytes.Contains(line, []byte(supersedesFileName)) {
			// Avoid double-marking: if the line already starts with [SUPERSEDED, skip.
			if bytes.HasPrefix(bytes.TrimLeft(line, " \t-*"), []byte("[SUPERSEDED ")) {
				return content
			}
			lines[i] = append(marker, line...)
			return bytes.Join(lines, []byte("\n"))
		}
	}
	return content
}
