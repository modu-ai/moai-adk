package config

import (
	"bufio"
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// @MX:NOTE: [AUTO] SPEC-TOKEN-EFFICIENCY-001 P0-1 — always-loaded 토큰 예산 가드.
// Claude Code가 매 턴 재주입하는 always-loaded 컨텍스트 표면(CLAUDE.md + no-paths: 규칙 +
// output-style + MEMORY.md head)의 토큰 총량이 예산을 넘으면 회귀로 판정한다. CC 네이티브
// 압축/캐싱은 재구현하지 않는다(over-engineering guard, plan.md §G).

// AlwaysLoadedTokenBudget는 always-loaded 컨텍스트 표면에 허용되는 추정 토큰 상한이다.
// 회귀 트립와이어로서, 측정 표면이 이 예산을 초과하면 가드가 실패한다 — Epic Steering-Align
// 다이어트가 조용히 되돌아가는 것을 잡는다.
//
// 도출 근거: 측정 baseline(2026-07-02) ≈ 64,624 토큰(char/4 추정: CLAUDE.md + no-paths:
// .claude/rules/moai/** 규칙 파일 + moai.md + MEMORY.md head 합계 258,498 bytes / 4) +
// 약 15% 여유(≈ 74,317)를 클린 상수로 올림. 여유분은 통상적 규칙 편집을 흡수하되 의미 있는
// 증가에는 발화한다.
const AlwaysLoadedTokenBudget = 75000

// memoryHeadLineCap / memoryHeadByteCap는 가드가 측정하는 MEMORY.md head 범위를 제한한다.
// Claude Code auto-memory 로더 상한(첫 200줄 또는 25KB 중 먼저 도달하는 쪽)과 일치한다.
const (
	memoryHeadLineCap = 200
	memoryHeadByteCap = 25 * 1024
)

// estimateTokens는 char/4 rule-of-thumb(len(bytes)/4)로 b의 근사 토큰 수를 반환한다.
// 실제 tokenizer 대비 ±약 15% 오차가 있는 의도적 무의존 근사다. 이 가드는 상대적 증가를
// 감시하는 트립와이어이지 회계 원장이 아니므로 정확도가 요구되지 않는다(simplicity ladder —
// 가드를 위해 tokenizer 의존성을 추가하지 않는다).
func estimateTokens(b []byte) int {
	return len(b) / 4
}

// findRepoRoot는 start에서 위로 올라가며 go.mod(리포 루트 마커)를 담은 디렉터리를 찾는다.
// go.mod 조상을 찾지 못하면 ("", false)를 반환한다(트리 밖에서 실행 시 가드는 graceful skip).
func findRepoRoot(start string) (string, bool) {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// frontmatterHasPaths는 마크다운 바이트의 YAML frontmatter에 top-level `paths:` 키가
// 있는지 판정한다. frontmatter는 첫 줄 `---`로 시작해 닫는 `---`로 끝난다. frontmatter가
// 아예 없으면(첫 줄이 `---`가 아니면) false. 닫는 `---` 이후의 `paths:`는 본문이므로 무시한다.
func frontmatterHasPaths(data []byte) bool {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			first = false
			if strings.TrimRight(line, " \t\r") != "---" {
				return false // frontmatter 없음
			}
			continue
		}
		if strings.TrimRight(line, " \t\r") == "---" {
			return false // frontmatter 종료, paths: 미발견
		}
		if strings.HasPrefix(line, "paths:") {
			return true
		}
	}
	return false // 닫히지 않은 frontmatter → 보수적으로 미제한(always-loaded로 계수)
}

// hasPathsRestriction은 path의 마크다운 파일이 frontmatter에 `paths:` 제한을 갖는지(즉
// 조건부 로드 규칙인지) 판정한다. 읽을 수 없거나 frontmatter가 없는 파일은 제한 없음으로
// 취급한다(보수적: always-loaded 표면에 계수).
func hasPathsRestriction(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false // 보수적: 계수
	}
	return frontmatterHasPaths(data)
}

// alwaysLoadedSurface는 repoRoot 기준 always-loaded 컨텍스트 표면을 나열한다: frontmatter에
// `paths:` 제한이 없는 모든 .claude/rules/moai/**/*.md 파일(정렬), 이어서 3개의 고정 표면
// 슬롯(CLAUDE.md, .claude/output-styles/moai/moai.md, MEMORY.md). 3개 고정 슬롯은 디스크에
// 파일이 없어도 항상 목록에 포함된다 — 없는 파일은 측정 시 0 토큰으로 계산한다(hermetic:
// machine-specific auto-memory 사본이 아니라 repo-relative MEMORY.md만 측정).
func alwaysLoadedSurface(repoRoot string) ([]string, error) {
	rulesDir := filepath.Join(repoRoot, ".claude", "rules", "moai")
	var ruleFiles []string
	err := filepath.WalkDir(rulesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		if !hasPathsRestriction(path) {
			ruleFiles = append(ruleFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(ruleFiles)

	// 3개 고정 표면 슬롯을 항상 고정 순서로 추가한다.
	fixed := []string{
		filepath.Join(repoRoot, "CLAUDE.md"),
		filepath.Join(repoRoot, ".claude", "output-styles", "moai", "moai.md"),
		filepath.Join(repoRoot, "MEMORY.md"),
	}
	return append(ruleFiles, fixed...), nil
}

// memoryHead는 MEMORY.md 내용의 로드 head를 반환한다: memoryHeadLineCap번째 개행까지 또는
// memoryHeadByteCap 바이트까지 중 먼저 도달하는 쪽 — Claude Code auto-memory 로더 상한과 일치.
func memoryHead(data []byte) []byte {
	if len(data) > memoryHeadByteCap {
		data = data[:memoryHeadByteCap]
	}
	lines := 0
	for i, c := range data {
		if c == '\n' {
			lines++
			if lines == memoryHeadLineCap {
				return data[:i+1]
			}
		}
	}
	return data
}

// measureAlwaysLoaded는 repoRoot 기준 always-loaded 표면의 추정 토큰을 합산한다. 총 토큰
// 추정치와 나열된 표면(카운트 assertion용)을 반환한다. MEMORY.md 슬롯은 head만 측정한다
// (memoryHeadLineCap 줄 또는 memoryHeadByteCap 바이트 중 먼저). 없는 파일은 0 토큰이다.
func measureAlwaysLoaded(repoRoot string) (total int, surface []string, err error) {
	surface, err = alwaysLoadedSurface(repoRoot)
	if err != nil {
		return 0, nil, err
	}
	memoryPath := filepath.Join(repoRoot, "MEMORY.md")
	for _, path := range surface {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			continue // 없는 파일 → 0 토큰(hermetic)
		}
		if path == memoryPath {
			data = memoryHead(data)
		}
		total += estimateTokens(data)
	}
	return total, surface, nil
}
