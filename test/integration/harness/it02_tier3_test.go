//go:build integration
// +build integration

// Package harness_integration — T-P5-02: Tier 3 프로모션 + 스냅샷 생성 검증 (REQ-HL-004).
package harness_integration

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness"
	"github.com/modu-ai/moai-adk/internal/harness/safety"
)

// TestIT02_Tier3Promotion은 Tier 3(Rule) 패턴에 대해 Applier.Apply()를 호출하고
// frontmatter 수정 결과와 스냅샷 생성을 검증한다.
// REQ-HL-004: Tier 3 rule 도달 시 trigger injection 대상.
// REQ-HL-005: Apply() 호출 시 스냅샷이 write보다 먼저 생성되어야 한다.
func TestIT02_Tier3Promotion(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	skillsDir := filepath.Join(dir, "skills", "my-harness-search")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatalf("skills 디렉토리 생성 실패: %v", err)
	}

	// ── 테스트용 SKILL.md 파일 생성 ──────────────────────────────────────
	// description 필드에 heuristicNote가 추가되어야 한다.
	skillContent := `---
name: my-harness-search
description: codebase 검색 스킬
triggers:
  - keyword: "search"
---

# My Harness Search Skill
`
	skillPath := filepath.Join(skillsDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(skillContent), 0o644); err != nil {
		t.Fatalf("SKILL.md 쓰기 실패: %v", err)
	}

	// ── Promotion 기록 ─────────────────────────────────────────────────────
	promotionPath := filepath.Join(dir, "tier-promotions.jsonl")
	learner := harness.NewLearner(promotionPath)

	promotion := harness.Promotion{
		Ts:               time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC),
		PatternKey:       "moai_subcommand:/moai search",
		FromTier:         "heuristic",
		ToTier:           "rule",
		ObservationCount: 5,
		Confidence:       1.0,
	}
	if err := learner.WritePromotion(promotion); err != nil {
		t.Fatalf("WritePromotion 실패: %v", err)
	}

	// promotions.jsonl 파일 존재 검증
	data, err := os.ReadFile(promotionPath)
	if err != nil {
		t.Fatalf("promotion 파일 읽기 실패: %v", err)
	}
	var readBack harness.Promotion
	if err := json.Unmarshal([]byte(strings.TrimRight(string(data), "\n")), &readBack); err != nil {
		t.Fatalf("promotion JSON 파싱 실패: %v", err)
	}
	if readBack.PatternKey != promotion.PatternKey {
		t.Errorf("PatternKey = %q, want %q", readBack.PatternKey, promotion.PatternKey)
	}

	// ── Apply: description 필드 수정 + 스냅샷 생성 ────────────────────────
	snapshotBase := filepath.Join(dir, "snapshots")

	// autoApply=true인 pipeline (테스트에서는 human oversight 건너뜀)
	violationLogPath := filepath.Join(dir, "frozen-guard-violations.jsonl")
	rateLimitPath := filepath.Join(dir, "rate-limit-state.json")
	pipeline := safety.NewPipeline(safety.PipelineConfig{
		ViolationLogPath: violationLogPath,
		RateLimitPath:    rateLimitPath,
		AutoApply:        true, // 테스트: L5 human oversight 생략
	})

	proposal := harness.Proposal{
		ID:               "prop-tier3-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "자주 사용되는 검색 패턴",
		PatternKey:       "moai_subcommand:/moai search",
		Tier:             harness.TierRule,
		ObservationCount: 5,
		CreatedAt:        time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC),
	}

	// 원본 파일 SHA256 기록 (롤백 검증을 위해)
	originalHash := sha256File(t, skillPath)

	applier := harness.NewApplier()
	if err := applier.Apply(proposal, pipeline, snapshotBase, nil); err != nil {
		t.Fatalf("Apply 실패: %v", err)
	}

	// ── 검증 1: frontmatter에 heuristicNote가 추가되었는지 ──────────────
	modifiedContent, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("수정된 SKILL.md 읽기 실패: %v", err)
	}
	if !strings.Contains(string(modifiedContent), "자주 사용되는 검색 패턴") {
		t.Errorf("description 필드에 heuristicNote가 추가되지 않음:\n%s", modifiedContent)
	}

	// ── 검증 2: 스냅샷 디렉토리가 생성되었는지 ──────────────────────────
	entries, err := os.ReadDir(snapshotBase)
	if err != nil {
		t.Fatalf("스냅샷 디렉토리 읽기 실패: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("스냅샷 디렉토리가 비어 있음 — Apply() 후 스냅샷이 생성되어야 한다")
	}

	// ── 검증 3: 스냅샷의 원본 파일이 수정 전 내용과 일치하는지 ─────────
	snapshotDir := filepath.Join(snapshotBase, entries[0].Name())
	backupPath := filepath.Join(snapshotDir, "SKILL.md")
	backupHash := sha256File(t, backupPath)
	if originalHash != backupHash {
		t.Errorf("스냅샷의 파일 해시 = %x, want %x (수정 전 원본과 불일치)", backupHash, originalHash)
	}

	// ── 검증 4: manifest.json 존재 및 필드 확인 ──────────────────────────
	manifestPath := filepath.Join(snapshotDir, "manifest.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("manifest.json 읽기 실패: %v", err)
	}
	var manifest map[string]interface{}
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		t.Fatalf("manifest.json 파싱 실패: %v", err)
	}
	if manifest["proposal_id"] != "prop-tier3-001" {
		t.Errorf("manifest proposal_id = %v, want prop-tier3-001", manifest["proposal_id"])
	}
}

// sha256File은 파일의 SHA256 해시를 반환한다.
func sha256File(t *testing.T, path string) [32]byte {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("파일 열기 실패 %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		t.Fatalf("SHA256 계산 실패 %s: %v", path, err)
	}
	var result [32]byte
	copy(result[:], h.Sum(nil))
	return result
}
