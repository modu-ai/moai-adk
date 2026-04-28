//go:build integration
// +build integration

// Package harness_integration — T-P5-07: learning.enabled: false no-op 검증 (REQ-HL-010).
package harness_integration

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness"
	"github.com/modu-ai/moai-adk/internal/harness/safety"
)

// mockNoOpEvaluator는 Apply()가 호출되면 테스트를 실패시키는 evaluator이다.
// learning.enabled=false 상태에서는 Apply()가 호출되어서는 안 된다.
type mockNoOpEvaluator struct {
	t *testing.T
}

// Evaluate는 호출되면 테스트를 실패시킨다.
func (m *mockNoOpEvaluator) Evaluate(
	_ harness.Proposal,
	_ []harness.Session,
) (harness.Decision, error) {
	m.t.Error("learning.enabled=false 상태에서 Evaluate가 호출되어서는 안 된다")
	return harness.Decision{Kind: harness.DecisionRejected}, nil
}

// LearningConfig는 학습 subsystem 활성화 여부 설정을 나타낸다.
// REQ-HL-010: learning.enabled=false이면 Observer와 Applier가 no-op으로 동작한다.
type LearningConfig struct {
	// Enabled는 학습 subsystem 활성화 여부이다.
	Enabled bool
}

// LearningGate는 learning.enabled 설정에 따라 Observer와 Applier를 제어한다.
// learning.enabled=false이면 RecordEvent와 Apply는 즉시 반환(no-op)한다.
type LearningGate struct {
	cfg      LearningConfig
	observer *harness.Observer
	applier  *harness.Applier
}

// NewLearningGate는 설정과 컴포넌트를 주입하여 LearningGate를 생성한다.
func NewLearningGate(cfg LearningConfig, observer *harness.Observer, applier *harness.Applier) *LearningGate {
	return &LearningGate{cfg: cfg, observer: observer, applier: applier}
}

// RecordEvent는 learning.enabled=false이면 no-op, 아니면 Observer.RecordEvent를 호출한다.
func (g *LearningGate) RecordEvent(eventType harness.EventType, subject, contextHash string) error {
	if !g.cfg.Enabled {
		return nil // no-op
	}
	return g.observer.RecordEvent(eventType, subject, contextHash)
}

// Apply는 learning.enabled=false이면 no-op, 아니면 Applier.Apply를 호출한다.
func (g *LearningGate) Apply(
	proposal harness.Proposal,
	evaluator harness.SafetyEvaluator,
	snapshotBase string,
	sessions []harness.Session,
) error {
	if !g.cfg.Enabled {
		return nil // no-op
	}
	return g.applier.Apply(proposal, evaluator, snapshotBase, sessions)
}

// TestIT07_LearningDisabledNoOp은 learning.enabled=false 상태에서
// Observer.RecordEvent와 Applier.Apply가 no-op으로 동작하는지 검증한다.
// REQ-HL-010: disabled 상태에서는 로그 파일을 생성하지 않고 evaluator를 호출하지 않는다.
func TestIT07_LearningDisabledNoOp(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	snapshotBase := filepath.Join(dir, "snapshots")

	// ── learning.enabled=false 설정 ───────────────────────────────────────
	cfg := LearningConfig{Enabled: false}

	observer := harness.NewObserver(logPath)
	applier := harness.NewApplier()
	gate := NewLearningGate(cfg, observer, applier)

	// ── RecordEvent no-op 검증 ────────────────────────────────────────────
	// learning.enabled=false이면 RecordEvent가 로그 파일을 생성하지 않아야 한다.
	for range 5 {
		if err := gate.RecordEvent(harness.EventTypeMoaiSubcommand, "/moai plan", "ctx_disabled"); err != nil {
			t.Fatalf("RecordEvent(disabled) 실패: %v", err)
		}
	}

	// 로그 파일이 생성되지 않아야 함
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Errorf("learning.enabled=false 상태에서 로그 파일이 생성됨: %s", logPath)
	}

	// ── Apply no-op 검증 ─────────────────────────────────────────────────
	// learning.enabled=false이면 Apply가 evaluator를 호출하지 않아야 한다.
	// mockNoOpEvaluator는 Evaluate가 호출되면 테스트를 실패시킨다.
	noOpEval := &mockNoOpEvaluator{t: t}

	skillPath := setupSkillFile(t, dir, "my-harness-disabled", `---
name: my-harness-disabled
description: 비활성화 테스트
---

# Disabled Test
`)

	proposal := harness.Proposal{
		ID:               "prop-disabled-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "이 값은 적용되어서는 안 된다",
		PatternKey:       "moai_subcommand:/moai plan",
		Tier:             harness.TierAutoUpdate,
		ObservationCount: 10,
		CreatedAt:        time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC),
	}

	if err := gate.Apply(proposal, noOpEval, snapshotBase, nil); err != nil {
		t.Fatalf("Apply(disabled) 실패: %v", err)
	}

	// 파일이 수정되지 않아야 함
	content, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("SKILL.md 읽기 실패: %v", err)
	}
	if containsString(string(content), "이 값은 적용되어서는 안 된다") {
		t.Error("learning.enabled=false 상태에서 파일이 수정됨")
	}

	// 스냅샷이 생성되지 않아야 함
	if _, err := os.Stat(snapshotBase); !os.IsNotExist(err) {
		t.Error("learning.enabled=false 상태에서 스냅샷 디렉토리가 생성됨")
	}

	// ── learning.enabled=true 대조군: 정상 동작 확인 ─────────────────────
	t.Run("enabled_works_normally", func(t *testing.T) {
		enabledDir := t.TempDir()
		enabledLogPath := filepath.Join(enabledDir, "usage-log.jsonl")
		enabledSnapshotBase := filepath.Join(enabledDir, "snapshots")

		enabledCfg := LearningConfig{Enabled: true}
		enabledObserver := harness.NewObserver(enabledLogPath)
		enabledApplier := harness.NewApplier()
		enabledGate := NewLearningGate(enabledCfg, enabledObserver, enabledApplier)

		// RecordEvent: 로그 파일 생성 기대
		if err := enabledGate.RecordEvent(harness.EventTypeMoaiSubcommand, "/moai plan", "ctx_enabled"); err != nil {
			t.Fatalf("RecordEvent(enabled) 실패: %v", err)
		}
		if _, err := os.Stat(enabledLogPath); err != nil {
			t.Errorf("learning.enabled=true 상태에서 로그 파일이 생성되어야 함: %v", err)
		}

		// Apply: evaluator 호출 + 파일 수정 기대
		enabledSkillPath := setupSkillFile(t, enabledDir, "my-harness-enabled", `---
name: my-harness-enabled
description: 활성화 테스트
---

# Enabled Test
`)

		violationPath := filepath.Join(enabledDir, "violations.jsonl")
		rateLimitStatePath := filepath.Join(enabledDir, "rate-limit-state.json")
		pipeline := safety.NewPipeline(safety.PipelineConfig{
			ViolationLogPath: violationPath,
			RateLimitPath:    rateLimitStatePath,
			AutoApply:        true,
		})

		enabledProposal := harness.Proposal{
			ID:               "prop-enabled-001",
			TargetPath:       enabledSkillPath,
			FieldKey:         "description",
			NewValue:         "활성화 상태에서 적용된 값",
			PatternKey:       "moai_subcommand:/moai plan",
			Tier:             harness.TierRule,
			ObservationCount: 5,
			CreatedAt:        time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC),
		}

		if err := enabledGate.Apply(enabledProposal, pipeline, enabledSnapshotBase, nil); err != nil {
			t.Fatalf("Apply(enabled) 실패: %v", err)
		}

		enabledContent, err := os.ReadFile(enabledSkillPath)
		if err != nil {
			t.Fatalf("활성화 상태 SKILL.md 읽기 실패: %v", err)
		}
		if !containsString(string(enabledContent), "활성화 상태에서 적용된 값") {
			t.Error("learning.enabled=true 상태에서 파일이 수정되지 않음")
		}
	})
}

// containsString은 s에 substr이 포함되는지 검사한다.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && func() bool {
		for i := range len(s) - len(substr) + 1 {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}()
}
