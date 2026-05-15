package spec

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// DriftRecord represents a single SPEC status drift entry
type DriftRecord struct {
	SPECID           string
	FrontmatterStatus string
	GitImpliedStatus  string
	Drifted          bool
}

// DriftReport represents a complete drift detection report
type DriftReport struct {
	Records []DriftRecord
	Count   int
}

// DetectDrift scans all SPECs and compares frontmatter status against git log
// Returns a report with all drift records and the total drift count
func DetectDrift(baseDir string) (*DriftReport, error) {
	specsDir := filepath.Join(baseDir, ".moai", "specs")

	// Check if specs directory exists
	if _, err := os.Stat(specsDir); os.IsNotExist(err) {
		return &DriftReport{Records: []DriftRecord{}, Count: 0}, nil
	}

	// Read all SPEC directories
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read specs directory: %w", err)
	}

	var records []DriftRecord
	driftCount := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		specID := entry.Name()
		specDir := filepath.Join(specsDir, specID)

		// Parse frontmatter status
		frontmatterStatus, err := ParseStatus(specDir)
		if err != nil {
			// Skip SPECs that can't be parsed
			continue
		}

		// Get git-implied status
		gitStatus, err := getGitImpliedStatus(specID)
		if err != nil {
			// If git history is empty or unavailable, skip
			continue
		}

		// Check for drift
		drifted := frontmatterStatus != gitStatus

		record := DriftRecord{
			SPECID:           specID,
			FrontmatterStatus: frontmatterStatus,
			GitImpliedStatus:  gitStatus,
			Drifted:          drifted,
		}

		records = append(records, record)

		if drifted {
			driftCount++
		}
	}

	// Sort records by SPEC-ID for consistent output
	sort.Slice(records, func(i, j int) bool {
		return records[i].SPECID < records[j].SPECID
	})

	return &DriftReport{
		Records: records,
		Count:   driftCount,
	}, nil
}

// gitLogWindowSize는 getGitImpliedStatus 가 git log에서 최대 몇 개의 commit을 조회할지 결정한다.
// @MX:NOTE: [AUTO] N=50 결정 근거: SPEC당 평균 git log 매칭 commit이 5-10건이므로 5-10x 안전 여유.
// @MX:REASON: SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 OQ1 — N 값 변경 시 plan.md §7 OQ1 참조.
const gitLogWindowSize = 50

// getGitImpliedStatus는 SPEC-ID에 대한 git log를 분석하여 lifecycle status를 추론한다.
//
// 본 함수는 SPEC-ID를 언급하는 git commit을 newest-to-oldest 순회하면서
// chore(spec): sweep commit 등 lifecycle 추론에서 의도적으로 제외되는 commit을 건너뛰고,
// 의미 있는 분류(ClassifyPRTitle이 비어있지 않은 status를 반환)를 가진 첫 commit의 status를 채택한다.
//
// 모든 N개 commit이 skip 대상이면 error를 반환하고,
// 상위 lint rule(StatusGitConsistencyRule)은 이를 skip 조건으로 처리한다.
//
// @MX:ANCHOR: [AUTO] getGitImpliedStatus — git-implied status 추론 진입점
// @MX:REASON: StatusGitConsistencyRule.Check + DetectDrift 두 곳에서 호출 (fan_in=2); walker filter 도입으로 SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 핵심 함수
func getGitImpliedStatus(specID string) (string, error) {
	// 기본 브랜치 결정 — main 우선, 없으면 master (현행 동작 유지)
	branch := "main"
	if _, err := exec.Command("git", "rev-parse", "--verify", "main").Output(); err != nil {
		branch = "master"
	}

	// SPEC-ID를 언급하는 최대 N개 commit을 newest-first 순서로 가져옴
	cmd := exec.Command("git", "log", branch, "--oneline", "--no-merges",
		"--grep="+specID, fmt.Sprintf("-%d", gitLogWindowSize))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git log failed: %w", err)
	}

	if len(output) == 0 {
		return "", fmt.Errorf("no git history found for %s", specID)
	}

	// 한 줄씩 순회 — newest first
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		// commit hash 분리 (형식: "<hash> <title>")
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			// 손상된 줄은 건너뜀
			continue
		}
		commitTitle := parts[1]

		// skip pattern 매칭 — chore(spec): 등 lifecycle 추론에서 제외되는 commit
		// @MX:NOTE: [AUTO] chore(spec) commit을 건너뜀 — bootstrapping bug 방지
		// @MX:REASON: sweep commit이 real impl commit을 가리던 SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 핵심 결함 수정
		if shouldSkipCommitTitle(commitTitle) {
			continue
		}

		// commit title 분류
		_, status, err := ClassifyPRTitle(commitTitle)
		if err != nil {
			// 분류 실패는 안전상 skip하고 다음 commit으로 이동
			continue
		}

		if status == "" {
			// unknown prefix 안전망 — 빈 status는 의미 있는 분류로 인정하지 않고 다음 commit 탐색
			continue
		}

		// 의미 있는 분류 발견 → 즉시 반환
		return status, nil
	}

	// N개 commit 모두 소진해도 의미 있는 분류를 못 찾음 → error 반환
	// StatusGitConsistencyRule::Check (lint.go:897-900)는 err != nil이면 finding을 emit하지 않는다
	// @MX:NOTE: [AUTO] walker N=50 소진 시 unknown signal — lint rule이 skip 처리하여 false-positive 방지
	// @MX:REASON: 모든 commit이 skip pattern인 SPEC은 git-consistency 검사 대상에서 제외 (fail-safe)
	return "", fmt.Errorf("no classifiable commit within window of %d for %s", gitLogWindowSize, specID)
}

// shouldSkipCommitTitle은 commit title이 알려진 skip pattern에 매칭되는지 확인한다.
//
// Skip-pattern commit은 lifecycle status 추론에서 제외해야 하는 메타데이터 유지보수 작업
// (frontmatter sweep, lint.skip 등록 등)을 나타낸다.
//
// v3.0.0-rc1 skip pattern: chore(spec): 와 chore(specs): 만 대상.
// 향후 패턴 추가는 별도 SPEC + plan.md §7 OQ2 externalization 결정 시 확장.
//
// @MX:NOTE: [AUTO] shouldSkipCommitTitle — chore(spec) sweep commit 필터
// @MX:REASON: SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 핵심 helper; skip pattern 변경 시 반드시 AC-LSCSK-003 regression guard 재실행
func shouldSkipCommitTitle(title string) bool {
	// 대소문자 무관 prefix 매칭 (plan.md §7 OQ1: strings.HasPrefix + ToLower 선택)
	lower := strings.ToLower(strings.TrimSpace(title))
	return strings.HasPrefix(lower, "chore(spec):") ||
		strings.HasPrefix(lower, "chore(specs):")
}

// DriftCount is a convenience function that returns only the drift count
func DriftCount(baseDir string) (int, error) {
	report, err := DetectDrift(baseDir)
	if err != nil {
		return 0, err
	}
	return report.Count, nil
}
