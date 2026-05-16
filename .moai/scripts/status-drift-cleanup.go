//go:build ignore
// +build ignore

// status-drift-cleanup.go — 원클릭 bulk cleanup for SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001
//
// 동작:
//   1. operations 슬라이스에서 적용할 패턴과 affected-list 파일 경로를 읽는다.
//   2. 각 SPEC spec.md 에서 frontmatter의 status 필드를 ToStatus로 변경한다.
//   3. version 필드를 patch +1 bump 한다.
//   4. updated_at 또는 updated 필드를 오늘 날짜로 갱신한다.
//   5. ## HISTORY 표에 새 row 1줄을 추가한다 (없으면 섹션 생성).
//   6. 이미 ToStatus인 SPEC은 건너뛴다 (idempotent).
//
// 실행:
//   cd /path/to/moai-adk-go
//   go run .moai/scripts/status-drift-cleanup.go
//
// 주의: gopkg.in/yaml.v3 가 go.mod에 있어야 한다.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	updatedDate      = "2026-05-16"
	historyRowAuthor = "manager-develop (run-phase)"
)

// Operation은 하나의 패턴에 대한 bulk 작업을 정의한다.
type Operation struct {
	// Pattern은 패턴 식별자 (예: "A", "B", "C")
	Pattern string
	// AffectedListPath는 SPEC ID 목록이 담긴 파일 경로 (주석 행은 #으로 시작)
	AffectedListPath string
	// FromStatus는 현재 frontmatter status (일치 검증용)
	FromStatus string
	// ToStatus는 변경 목표 status
	ToStatus string
	// HistoryDescr은 HISTORY 테이블에 추가할 설명
	HistoryDescr string
}

// 처리할 패턴 목록 (순서대로 실행)
var operations = []Operation{
	{
		Pattern:          "A",
		AffectedListPath: ".moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001/affected-list-pattern-A.txt",
		FromStatus:       "completed",
		ToStatus:         "implemented",
		HistoryDescr:     "status downgrade completed → implemented — git-implied status 정합성 복원 (SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 Wave 2).",
	},
	{
		Pattern:          "B",
		AffectedListPath: ".moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001/affected-list-pattern-B.txt",
		FromStatus:       "completed",
		ToStatus:         "in-progress",
		HistoryDescr:     "status downgrade completed → in-progress — git-implied status 정합성 복원 (SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 Wave 3).",
	},
	{
		Pattern:          "C",
		AffectedListPath: ".moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001/affected-list-pattern-C.txt",
		FromStatus:       "implemented",
		ToStatus:         "in-progress",
		HistoryDescr:     "status downgrade implemented → in-progress — git-implied status 정합성 복원 (SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 Wave 3).",
	},
}

func main() {
	totalCleaned := 0
	totalSkipped := 0
	totalErrors := 0

	for _, op := range operations {
		fmt.Printf("\n=== Pattern %s (%s → %s) ===\n", op.Pattern, op.FromStatus, op.ToStatus)

		// affected-list 파일 읽기
		listData, err := os.ReadFile(op.AffectedListPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: cannot read %s: %v\n", op.AffectedListPath, err)
			totalErrors++
			continue
		}

		specIDs := []string{}
		for _, line := range strings.Split(strings.TrimSpace(string(listData)), "\n") {
			line = strings.TrimSpace(line)
			// 주석 행(#으로 시작) 및 빈 행 건너뜀
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			specIDs = append(specIDs, line)
		}

		fmt.Printf("Processing %d SPECs...\n", len(specIDs))

		for _, specID := range specIDs {
			specPath := filepath.Join(".moai/specs", specID, "spec.md")
			action, err := processSpec(specPath, specID, op)
			if err != nil {
				fmt.Printf("ERROR: %s: %v\n", specID, err)
				totalErrors++
				continue
			}
			fmt.Printf("  %s: %s\n", specID, action)
			if strings.HasPrefix(action, "skipped") {
				totalSkipped++
			} else {
				totalCleaned++
			}
		}
	}

	fmt.Printf("\n=== 최종 요약 ===\n")
	fmt.Printf("cleaned=%d skipped=%d errors=%d\n", totalCleaned, totalSkipped, totalErrors)
	if totalErrors > 0 {
		os.Exit(1)
	}
}

func processSpec(specPath, specID string, op Operation) (string, error) {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return "", fmt.Errorf("read: %w", err)
	}

	content := string(data)

	// frontmatter 분리
	if !strings.HasPrefix(content, "---\n") {
		return "", fmt.Errorf("no frontmatter start")
	}

	rest := content[4:] // "---\n" 건너뜀
	fmEnd := strings.Index(rest, "\n---\n")
	var fmText, bodyText string
	if fmEnd >= 0 {
		fmText = rest[:fmEnd]
		bodyText = rest[fmEnd+5:]
	} else {
		fmEnd2 := strings.Index(rest, "\n---")
		if fmEnd2 >= 0 && fmEnd2 == len(rest)-4 {
			fmText = rest[:fmEnd2]
			bodyText = ""
		} else {
			return "", fmt.Errorf("no frontmatter end")
		}
	}

	// yaml.Node API로 파싱 (key 순서 보존)
	var docNode yaml.Node
	if err := yaml.Unmarshal([]byte(fmText), &docNode); err != nil {
		return "", fmt.Errorf("yaml parse: %w", err)
	}

	if docNode.Kind != yaml.DocumentNode || len(docNode.Content) == 0 {
		return "", fmt.Errorf("yaml: unexpected document structure")
	}

	mapping := docNode.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return "", fmt.Errorf("yaml: root is not mapping")
	}

	// status 필드 찾기
	statusIdx := findKey(mapping.Content, "status")
	if statusIdx < 0 {
		return "", fmt.Errorf("no status field")
	}
	currentStatus := strings.Trim(mapping.Content[statusIdx+1].Value, `"'`)

	// Idempotency 체크: 이미 목표 status라면 스킵
	if currentStatus == op.ToStatus {
		return fmt.Sprintf("skipped (already %s)", op.ToStatus), nil
	}

	// FromStatus 불일치 시 경고 (스킵하지 않고 강제 변환은 하지 않음)
	if currentStatus != op.FromStatus {
		return fmt.Sprintf("skipped (status=%s, expected=%s, would-set=%s)", currentStatus, op.FromStatus, op.ToStatus), nil
	}

	// version 필드
	versionIdx := findKey(mapping.Content, "version")
	if versionIdx < 0 {
		return "", fmt.Errorf("no version field")
	}
	currentVersion := strings.Trim(mapping.Content[versionIdx+1].Value, `"'`)
	origValNode := mapping.Content[versionIdx+1]
	wasQuoted := origValNode.Style == yaml.DoubleQuotedStyle || origValNode.Style == yaml.SingleQuotedStyle

	newVersion := patchBump(currentVersion)
	mapping.Content[versionIdx+1].Value = newVersion
	if wasQuoted {
		mapping.Content[versionIdx+1].Style = yaml.DoubleQuotedStyle
	} else {
		mapping.Content[versionIdx+1].Style = 0
	}

	// status 변경
	mapping.Content[statusIdx+1].Value = op.ToStatus
	mapping.Content[statusIdx+1].Style = 0

	// updated_at 또는 updated 필드 갱신
	updatedAtIdx := findKey(mapping.Content, "updated_at")
	if updatedAtIdx >= 0 {
		mapping.Content[updatedAtIdx+1].Value = updatedDate
		mapping.Content[updatedAtIdx+1].Style = 0
	} else {
		updatedIdx := findKey(mapping.Content, "updated")
		if updatedIdx >= 0 {
			mapping.Content[updatedIdx+1].Value = updatedDate
			mapping.Content[updatedIdx+1].Style = 0
		}
	}

	// frontmatter 재직렬화
	newFMBytes, err := yaml.Marshal(&docNode)
	if err != nil {
		return "", fmt.Errorf("yaml marshal: %w", err)
	}
	newFM := strings.TrimSuffix(string(newFMBytes), "\n")

	// HISTORY row 삽입
	newBody := insertHistoryRow(bodyText, newVersion, updatedDate, op.HistoryDescr)

	// 파일 재조립
	result := "---\n" + newFM + "\n---\n" + newBody

	if err := os.WriteFile(specPath, []byte(result), 0644); err != nil {
		return "", fmt.Errorf("write: %w", err)
	}

	return fmt.Sprintf("cleaned (status %s → %s, version %s → %s)", op.FromStatus, op.ToStatus, currentVersion, newVersion), nil
}

// findKey는 yaml mapping node content 슬라이스에서 key의 인덱스를 반환한다.
// 없으면 -1 반환.
func findKey(content []*yaml.Node, key string) int {
	for i := 0; i < len(content)-1; i += 2 {
		if content[i].Value == key {
			return i
		}
	}
	return -1
}

// patchBump는 버전 문자열의 patch 번호를 1 증가시킨다.
func patchBump(version string) string {
	v := strings.Trim(version, `"'`)
	parts := strings.Split(v, ".")
	if len(parts) < 3 {
		parts = append(parts, "0")
	}
	patch, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return v + ".1"
	}
	parts[len(parts)-1] = strconv.Itoa(patch + 1)
	return strings.Join(parts, ".")
}

// insertHistoryRow는 body의 ## HISTORY 섹션에 새 행을 추가한다.
func insertHistoryRow(body, newVersion, date, desc string) string {
	newRow := fmt.Sprintf("| %-7s | %-10s | %s | %s |",
		newVersion, date, historyRowAuthor, desc)

	lines := strings.Split(body, "\n")

	// ## HISTORY 섹션 찾기
	historyIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "## HISTORY" {
			historyIdx = i
			break
		}
	}

	if historyIdx < 0 {
		return insertHistorySection(body, newRow)
	}

	tableHeaderIdx := -1
	tableSepIdx := -1
	firstDataRowIdx := -1
	lastDataRowIdx := -1
	bulletStartIdx := -1

	for i := historyIdx + 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "---") {
			break
		}

		if strings.HasPrefix(trimmed, "| ") || strings.HasPrefix(trimmed, "|---") || strings.HasPrefix(trimmed, "|:--") {
			if tableHeaderIdx < 0 && strings.Contains(trimmed, "|") && !strings.Contains(trimmed, "---") {
				tableHeaderIdx = i
			} else if tableHeaderIdx >= 0 && tableSepIdx < 0 && strings.Contains(trimmed, "---") {
				tableSepIdx = i
			} else if tableSepIdx >= 0 && strings.HasPrefix(trimmed, "|") {
				if firstDataRowIdx < 0 {
					firstDataRowIdx = i
				}
				lastDataRowIdx = i
			}
		} else if strings.HasPrefix(trimmed, "- ") {
			if bulletStartIdx < 0 {
				bulletStartIdx = i
			}
		}
	}

	if bulletStartIdx >= 0 && tableHeaderIdx < 0 {
		bulletRow := fmt.Sprintf("- %s v%s: %s", date, newVersion, desc)
		return insertAtBulletEnd(lines, historyIdx, bulletRow)
	}

	if tableHeaderIdx < 0 {
		return insertTableAfterHistoryHeader(lines, historyIdx, newRow)
	}

	if firstDataRowIdx < 0 {
		insertPos := tableSepIdx + 1
		if tableSepIdx < 0 {
			insertPos = tableHeaderIdx + 1
		}
		newLines := make([]string, 0, len(lines)+1)
		newLines = append(newLines, lines[:insertPos]...)
		newLines = append(newLines, newRow)
		newLines = append(newLines, lines[insertPos:]...)
		return strings.Join(newLines, "\n")
	}

	// 신구 ordering 판별: top-newest vs bottom-newest
	firstRowVersion := extractVersionFromTableRow(lines[firstDataRowIdx])
	lastRowVersion := extractVersionFromTableRow(lines[lastDataRowIdx])
	isTopNewest := compareVersions(firstRowVersion, lastRowVersion) > 0

	var insertPos int
	if isTopNewest || firstDataRowIdx == lastDataRowIdx {
		if tableSepIdx >= 0 {
			insertPos = tableSepIdx + 1
		} else {
			insertPos = tableHeaderIdx + 1
		}
	} else {
		insertPos = lastDataRowIdx + 1
	}

	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertPos]...)
	newLines = append(newLines, newRow)
	newLines = append(newLines, lines[insertPos:]...)
	return strings.Join(newLines, "\n")
}

func insertHistorySection(body, newRow string) string {
	lines := strings.Split(body, "\n")
	h1Idx := -1
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "# ") {
			h1Idx = i
			break
		}
	}

	tableSection := []string{
		"",
		"## HISTORY",
		"",
		"| Version | Date       | Author                     | Description |",
		"|---------|------------|----------------------------|-------------|",
		newRow,
		"",
	}

	var insertPos int
	if h1Idx >= 0 {
		insertPos = h1Idx + 1
		for insertPos < len(lines) && strings.TrimSpace(lines[insertPos]) == "" {
			insertPos++
		}
	} else {
		insertPos = 0
	}

	newLines := make([]string, 0, len(lines)+len(tableSection))
	newLines = append(newLines, lines[:insertPos]...)
	newLines = append(newLines, tableSection...)
	newLines = append(newLines, lines[insertPos:]...)
	return strings.Join(newLines, "\n")
}

func insertAtBulletEnd(lines []string, historyIdx int, bulletRow string) string {
	lastBulletIdx := -1
	for i := historyIdx + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "---") {
			break
		}
		if strings.HasPrefix(trimmed, "- ") {
			lastBulletIdx = i
		}
	}

	insertPos := lastBulletIdx + 1
	if lastBulletIdx < 0 {
		insertPos = historyIdx + 1
	}

	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertPos]...)
	newLines = append(newLines, bulletRow)
	newLines = append(newLines, lines[insertPos:]...)
	return strings.Join(newLines, "\n")
}

func insertTableAfterHistoryHeader(lines []string, historyIdx int, newRow string) string {
	insertPos := historyIdx + 1
	for insertPos < len(lines) && strings.TrimSpace(lines[insertPos]) == "" {
		insertPos++
	}

	tableLines := []string{
		"| Version | Date       | Author                     | Description |",
		"|---------|------------|----------------------------|-------------|",
		newRow,
	}

	newLines := make([]string, 0, len(lines)+len(tableLines))
	newLines = append(newLines, lines[:insertPos]...)
	newLines = append(newLines, tableLines...)
	newLines = append(newLines, lines[insertPos:]...)
	return strings.Join(newLines, "\n")
}

func extractVersionFromTableRow(row string) string {
	parts := strings.SplitN(row, "|", 3)
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func compareVersions(a, b string) int {
	aParts := strings.Split(strings.Trim(a, `"' `), ".")
	bParts := strings.Split(strings.Trim(b, `"' `), ".")

	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}

	for i := 0; i < maxLen; i++ {
		var aVal, bVal int
		if i < len(aParts) {
			aVal, _ = strconv.Atoi(strings.TrimSpace(aParts[i]))
		}
		if i < len(bParts) {
			bVal, _ = strconv.Atoi(strings.TrimSpace(bParts[i]))
		}
		if aVal != bVal {
			return aVal - bVal
		}
	}
	return 0
}
