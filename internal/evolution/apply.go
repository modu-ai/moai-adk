package evolution

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/merge"
)

// ApplyProposal reads the skill file identified by proposal.TargetFile,
// locates the evolvable zone proposal.ZoneID, and appends proposal.Addition
// to the zone content.
//
// A backup of the original file is written to <TargetFile>.bak before any
// modification so that RevertProposal can restore it.
//
// IMPORTANT: This function is provided for use after explicit user approval.
// It MUST NOT be called autonomously.
func ApplyProposal(projectRoot string, proposal *ProposedChange) error {
	if proposal == nil {
		return fmt.Errorf("evolution: nil ProposedChange")
	}
	if err := CheckFrozenGuard(proposal.TargetFile); err != nil {
		return err
	}

	fullPath := filepath.Join(projectRoot, proposal.TargetFile)

	// projectRoot 탈출 시도 검증: 정규화된 경로가 여전히 projectRoot 하위인지 확인
	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return fmt.Errorf("evolution: resolve project root: %w", err)
	}
	absTarget, err := filepath.Abs(fullPath)
	if err != nil {
		return fmt.Errorf("evolution: resolve target path: %w", err)
	}
	rel, err := filepath.Rel(absRoot, absTarget)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("evolution: target file %q escapes project root", proposal.TargetFile)
	}
	original, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("evolution: read target file %s: %w", fullPath, err)
	}

	// Write backup.
	backupPath := fullPath + ".bak"
	if err := os.WriteFile(backupPath, original, 0o644); err != nil {
		return fmt.Errorf("evolution: write backup %s: %w", backupPath, err)
	}

	// 기존 존 내용을 찾아 추가 내용을 붙인다.
	// ParseEvolvableZones로 현재 존 내용을 읽은 후 proposal.Addition을 이어 붙이고,
	// ReplaceEvolvableZone으로 파일 전체 구조(헤더/푸터)를 보존하며 존만 교체한다.
	zones, parseErr := merge.ParseEvolvableZones(string(original))
	if parseErr != nil {
		return fmt.Errorf("evolution: parse evolvable zones: %w", parseErr)
	}

	// 대상 존 존재 확인 및 기존 내용 추출
	var existingContent string
	found := false
	for _, z := range zones {
		if z.ID == proposal.ZoneID {
			existingContent = z.Content
			found = true
			break
		}
	}
	if !found {
		return ErrZoneNotFound
	}

	// 새 존 내용: 기존 내용 + 구분자 + 추가 내용
	newContent := existingContent
	if newContent != "" && newContent[len(newContent)-1] != '\n' {
		newContent += "\n"
	}
	newContent += proposal.Addition

	// ReplaceEvolvableZone은 파일의 헤더/푸터를 보존하면서 특정 존만 교체한다.
	// (이전 코드의 MergeEvolvableZones(original, zoneID, newContent) 오용을 수정)
	updated, err := merge.ReplaceEvolvableZone(string(original), proposal.ZoneID, newContent)
	if err != nil {
		return fmt.Errorf("evolution: replace evolvable zone: %w", err)
	}

	// Write atomically.
	tmp := fullPath + ".tmp"
	if err := os.WriteFile(tmp, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("evolution: write updated file: %w", err)
	}
	if err := os.Rename(tmp, fullPath); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("evolution: rename updated file: %w", err)
	}
	return nil
}

// RevertProposal restores the original file from the .bak backup created by
// ApplyProposal.  The backup is deleted after a successful restore.
func RevertProposal(projectRoot string, proposal *ProposedChange) error {
	if proposal == nil {
		return fmt.Errorf("evolution: nil ProposedChange for revert")
	}

	fullPath := filepath.Join(projectRoot, proposal.TargetFile)
	backupPath := fullPath + ".bak"

	backup, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("evolution: read backup %s: %w", backupPath, err)
	}

	if err := os.WriteFile(fullPath, backup, 0o644); err != nil {
		return fmt.Errorf("evolution: restore from backup: %w", err)
	}

	_ = os.Remove(backupPath)
	return nil
}
