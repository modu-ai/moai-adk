---
id: SPEC-SKILL-DESC-001
plan_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Implementation Plan — SPEC-SKILL-DESC-001

## 1. Overview

builder-skill에 description optimization 기능을 추가. SPEC-SKILL-TEST-001 (Wave 2) framework 활용해 FP/FN 측정, 임계값 초과 시 LLM 기반 tightening/broadening 제안. Auto-apply 절대 금지.

## 2. Approach Summary

**전략**: Reuse-Test-Framework, LLM-Suggester, Approval-Gated-Apply.

1. SPEC-SKILL-TEST-001 framework 호출해 FP/FN 측정
2. 임계값 초과 시 LLM call (Opus 또는 advisor)로 description 재작성 제안
3. before/after diff + regression run
4. AskUserQuestion 또는 `--apply` flag로 사용자 승인
5. apply 후 full catalog regression run (cross-effect 보호)

## 3. Milestones (Priority-based, no time estimates)

### M0 — Pre-flight (Priority: Critical)

- [ ] SPEC-SKILL-TEST-001 (Wave 2) framework 가용성 확인 (sample prompts, FP/FN measurement API)
- [ ] 본 프로젝트의 100+ skill 중 sample prompts 보유한 skill 카운트
- [ ] LLM call 가능 경로 확인 (builder-skill의 advisor 또는 Opus 직접)
- [ ] `~/.claude/projects/<hash>/memory/` MEMORY.md 형식 (cross-ref)

**Exit Criteria**: dependency 가용성 + sample 데이터 base 확인

### M1 — Analyzer Implementation (Priority: High)

- [ ] `internal/skill/optimizer/analyzer.go`:
  - `func MeasureMetrics(skillName string) (Metrics, error)`
  - SPEC-SKILL-TEST-001 framework 호출 (sample prompts → FP/FN)
  - struct `Metrics { FalsePositiveRate, FalseNegativeRate, SampleCount }`
  - minimum sample size 5 (REQ-SD-011)
- [ ] sample size 부족 시 "data insufficient" error

**Exit Criteria**: 5 skill에서 metric 측정 검증

### M2 — Suggester Implementation (Priority: High)

- [ ] `internal/skill/optimizer/suggester.go`:
  - `func SuggestTightening(currentDesc string, samplePrompts []Prompt) (string, error)` — LLM call
  - `func SuggestBroadening(currentDesc string, samplePrompts []Prompt) (string, error)` — LLM call
  - LLM prompt: 현재 description + FP/FN 메트릭 + 사례 → 새 description 제안
  - struct `Suggestion { OldDesc, NewDesc, Rationale, ExpectedFP, ExpectedFN }`
- [ ] LLM call 경로: builder-skill의 advisor 또는 직접 Anthropic API (SPEC-ADVISOR-001 패턴 준수)

**Exit Criteria**: 5 skill에서 suggestion 생성 검증

### M3 — Threshold Logic (Priority: High)

- [ ] 임계값 default: FP > 15% (tightening), FN > 10% (broadening)
- [ ] per-skill override: skill frontmatter `optimization_thresholds: {fp: 0.20, fn: 0.15}` 지원
- [ ] both 임계값 미초과 시 "no optimization needed" 메시지
- [ ] both 초과 시 우선순위 결정 (FP 먼저, FN 그 다음)

**Exit Criteria**: 임계값 로직 unit test PASS

### M4 — Approval Gate (Priority: Critical)

- [ ] CLI 호출 시 default: 제안 출력 후 사용자 입력 대기
- [ ] orchestrator flow: AskUserQuestion 라운드 (suggestion accept / reject / modify)
- [ ] `--apply` flag 시: 즉시 적용 + audit log 의무
- [ ] `--dry-run` flag: 적용 없이 제안만 출력

**Exit Criteria**: approval 없이 변경 100% block

### M5 — Apply + Regression (Priority: High)

- [ ] 승인 후:
  1. skill frontmatter `description:` 업데이트
  2. SPEC-SKILL-TEST-001 framework로 재측정 (post-apply FP/FN)
  3. full catalog regression run (cross-effect 측정)
  4. diff + improvement 보고서 작성
- [ ] cross-effect > 5% routing 변동 시 자동 rollback + warning

**Exit Criteria**: 5 skill에서 apply + regression 검증

### M6 — Loop Detection (Priority: Medium)

- [ ] 같은 description 반복 제안 시 loop 감지
- [ ] state file `.moai/state/skill-optimizer/<NAME>.json`에 history 기록
- [ ] convergence 도달 시 "convergence reached" 메시지 + halt

**Exit Criteria**: loop scenario에서 halt 검증

### M7 — CLI Subcommand (Priority: High)

- [ ] `cmd/moai/skill.go`에 `optimize` 서브커맨드:
  - `moai skill optimize <name>`: 측정 + 제안 + (대화) 적용
  - `moai skill optimize <name> --apply`: 자동 적용 (audit 기록 의무)
  - `moai skill optimize <name> --dry-run`: 제안만 출력
- [ ] verbose mode (-v)

**Exit Criteria**: CLI functional

### M8 — builder-skill Body 확장 (Priority: High)

- [ ] `.claude/agents/moai/builder-skill.md`에 새 절 추가:
  - "Description Optimization Protocol" 절
  - 5 단계 명시 (measure → suggest → diff → approve → apply + regression)
  - 임계값 / cross-effect 보호 규칙
  - LLM call 가이드 (advisor 활용)
- [ ] Template-First 동기화

**Exit Criteria**: builder-skill body에 새 절 작성

### M9 — 정책 문서 (Priority: Medium)

- [ ] `.claude/rules/moai/development/skill-description.md` 신규:
  - description 작성 가이드 (post-optimization 학습 반영)
  - broad / narrow / balanced 사례
  - 임계값 의미
  - sample prompt 작성 권장
- [ ] Template-First 동기화

**Exit Criteria**: 가이드 문서 작성

### M10 — Optimization Report (Priority: Medium)

- [ ] `.moai/reports/skill-optimization-<NAME>-<DATE>.md` 형식 정의:
  - before/after FP/FN
  - description before/after diff
  - LLM rationale
  - cross-effect measurement
  - approval decision (auto-apply / user / reject)

**Exit Criteria**: 보고서 정상 생성

### M11 — Validation + Acceptance Sign-off (Priority: High)

- [ ] acceptance.md 시나리오 모두 PASS
- [ ] auto-apply block 강제 검증
- [ ] cross-effect 보호 검증
- [ ] dependency (SPEC-SKILL-TEST-001) 정상 호출
- [ ] plan-auditor PASS

**Exit Criteria**: 모든 acceptance + plan-auditor PASS

## 4. Technical Approach

### 4.1 Optimization workflow (의사코드)

```go
func Optimize(skillName string, opts Options) error {
    // M1: measure
    metrics, err := analyzer.MeasureMetrics(skillName)
    if err != nil { return err }
    if metrics.SampleCount < 5 {
        return ErrDataInsufficient
    }
    
    // M3: threshold check
    if metrics.FP <= 0.15 && metrics.FN <= 0.10 {
        return logInfo("no optimization needed")
    }
    
    // M2: suggest
    var suggestion *Suggestion
    if metrics.FP > 0.15 {
        suggestion, err = suggester.SuggestTightening(...)
    } else {
        suggestion, err = suggester.SuggestBroadening(...)
    }
    
    // M6: loop detection
    if convergenceReached(skillName, suggestion.NewDesc) {
        return logInfo("convergence reached")
    }
    
    // M4: approval
    approved, err := getApproval(suggestion, opts)
    if !approved { return ErrUserRejected }
    
    // M5: apply + regression
    if err := applyDescription(skillName, suggestion.NewDesc); err != nil { return err }
    if crossEffect := runCatalogRegression(); crossEffect > 0.05 {
        rollback(skillName)
        return ErrCrossEffectExceeded
    }
    
    // M10: report
    return writeReport(skillName, metrics, suggestion, opts)
}
```

### 4.2 임계값 + override 적용

```yaml
# skill frontmatter (optional)
---
name: moai-workflow-spec
description: ...
optimization_thresholds:
  fp: 0.20  # override default 0.15
  fn: 0.12  # override default 0.10
---
```

## 5. Risks and Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| sample prompts 부족 | High | High | min 5 hard requirement, "data insufficient" halt |
| LLM 제안 hallucination | Medium | High | approval gate strict, dry-run default 권장 |
| cross-effect 폭증 | Medium | High | M5 자동 rollback + 5% 임계값 |
| convergence 미감지 | Low | Medium | M6 history 기반 detection |
| Wave 2 의존성 미완성 | Low | Critical | blockedBy 명시 + dependency 게이트 |
| 임계값 default 부적합 | Medium | Low | per-skill override 지원 |

## 6. Dependencies

- 선행 SPEC: **SPEC-SKILL-TEST-001** (Wave 2, blockedBy) — sample prompts + metric 측정 framework
- 의존: SPEC-ADVISOR-001 advisor pattern (LLM call 효율화 권장)
- sibling SPEC: SPEC-SKILL-002, SPEC-SKILL-ENHANCE-001
- 도구: `make build`, plan-auditor

## 7. Open Questions Resolution

- **OQ1** (호출 빈도): 사용자 명시 호출만 (자동 schedule 금지)
- **OQ2** (per-skill override 구조): skill frontmatter `optimization_thresholds:` (M3에 명시)
- **OQ3** (보고서 보존): 1년 보존, 이후 archive
- **OQ4** (반복 무한 루프): convergence detection (M6)

## 8. Rollout Plan

1. M1-M5 구현 후 dogfooding: 5 skill에 시범 적용
2. acceptance rate >= 60% 검증
3. CHANGELOG + v2.x.0 minor release
4. SPEC-SKILL-TEST-001과 통합 회귀 테스트

End of plan.md (SPEC-SKILL-DESC-001).
