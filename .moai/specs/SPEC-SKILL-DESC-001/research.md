# Research — SPEC-SKILL-DESC-001 (Skill Description Optimization)

**SPEC**: SPEC-SKILL-DESC-001
**Wave**: 3 / Tier 2 (검증 통과 — 의존: SPEC-SKILL-TEST-001 Wave 2)
**Created**: 2026-04-30
**Author**: manager-spec

---

## 0. Cross-Worktree Dependency Notice

본 SPEC은 다른 worktree에 위치한 SPEC을 참조합니다. plan-auditor가 현재 worktree (wave-3-tier2)만 보면 dependency 부재로 잘못 판정할 수 있으므로 명시합니다.

| 의존 SPEC | 위치 | Branch | PR | 의존 유형 |
|-----------|------|--------|------|-----------|
| **SPEC-SKILL-TEST-001** | `wave-2-tier1` worktree | `feature/wave-2-tier1` | **#748** | blockedBy (Wave 2 산출물의 regression test framework 활용) |

본 SPEC 구현 전 위 의존 SPEC이 main에 머지되어야 합니다. frontmatter `blockedBy: [SPEC-SKILL-TEST-001]` 그대로 유지하며, 이 의존은 **cross-worktree** 임을 인지하고 plan-auditor false alarm 가능성을 명시합니다.

검증 방법:
- main repo에서 `gh pr view 748 --json mergedAt`로 머지 여부 확인
- 또는 `git fetch origin && git log origin/main --grep "SPEC-SKILL-TEST-001"`

---

## 1. 출처 (Anthropic 공식 자료)

### 1.1 Verbatim 인용

Anthropic blog "Improving Skill Creator: Test, Measure, Refine":

> "Skill-creator's description optimization feature analyzes your current description against sample prompts and suggests edits."

> "A description that is too broad triggers the skill on irrelevant prompts (false positives). A description that is too narrow misses prompts the skill should handle (false negatives). Both degrade routing quality."

> "Measure the false-positive and false-negative rate against a held-out test set, then refine the description until both stay below threshold."

### 1.2 검증 (claude-code-guide 검증 완료)

claude-code-guide 에이전트가 Claude Code skill frontmatter (`description:` field) 기준으로 검증함. 결론:

- **호환성**: ✅ 완전 지원 — `description:` field는 frontmatter 표준이며 skill router가 prompt-skill matching 시 사용
- **선행 의존**: SPEC-SKILL-TEST-001 (Wave 2)이 sample prompt 기반 regression test framework 제공 → 본 SPEC이 활용
- **권고 채택**: ACCEPT — builder-skill에 description optimizer 기능 추가

---

## 2. 현재 상태 (As-Is)

### 2.1 skill catalog 규모

`.claude/skills/moai*/SKILL.md`: 100+ skill (Wave 1 consolidation 후)

각 skill은 frontmatter에 `description:` 보유:

```yaml
---
name: moai-workflow-spec
description: SPEC Workflow Management — EARS format specifications, 4-file structure, Plan-Run-Sync integration
---
```

**관찰**: description은 1-3줄 자연어. router가 user prompt와 cosine similarity / pattern match로 트리거 결정.

### 2.2 description 품질 현황

분포 추정 (수동 sampling):
- broad (false positive 잠재): ~20% (예: "comprehensive workflow management")
- narrow (false negative 잠재): ~10% (예: "Sphinx documentation generator only")
- balanced: ~70%

**관찰**: 자동 측정 인프라 부재 → 추정치만 가능.

### 2.3 SPEC-SKILL-TEST-001 (Wave 2) 가용성

Wave 2에서 완성된 SPEC-SKILL-TEST-001은:
- Sample prompts 기반 skill routing regression test framework
- false-positive/false-negative metric 측정 가능
- 각 skill마다 `tests/` 디렉토리에 sample prompts 보유

**관찰**: 본 SPEC은 SPEC-SKILL-TEST-001의 측정 파이프라인을 입력으로 활용.

### 2.4 builder-skill 현황

`.claude/agents/moai/builder-skill.md`:
- skill 신규 생성 / 업데이트
- frontmatter 검증
- module 분리 가이드

**관찰**: optimization 기능 부재. description은 사람이 작성하고 변경.

---

## 3. 격차 분석 (Gap Analysis)

| 영역 | 현재 (As-Is) | 목표 (To-Be) | 격차 |
|------|-------------|-------------|------|
| description optimization | 없음 | builder-skill에 신규 기능 | 신규 |
| false-positive 임계값 | 없음 | > 15% → tightening 제안 | 임계값 |
| false-negative 임계값 | 없음 | > 10% → broadening 제안 | 임계값 |
| before/after diff | 없음 | optimization 산출물 | 신규 |
| measured improvement | 없음 | metric 비교 보고 | 신규 |
| regression test 자동 실행 | SPEC-SKILL-TEST-001만 (manual) | description 변경 시 자동 트리거 | 통합 |

---

## 4. 코드베이스 분석 (Affected Files)

### 4.1 Primary 수정 대상

| 파일 | 수정 유형 | 변경 사유 |
|------|----------|----------|
| `.claude/agents/moai/builder-skill.md` | 수정 (확장) | description optimizer protocol 절 추가 |
| `internal/skill/optimizer/analyzer.go` | 신규 | sample prompt 입력 → metric 측정 |
| `internal/skill/optimizer/suggester.go` | 신규 | tightening/broadening 제안 생성 |
| `cmd/moai/skill.go` | 수정 | `moai skill optimize <name>` 서브커맨드 |
| `.claude/rules/moai/development/skill-description.md` | 신규 | description 작성 가이드 (post-optimization) |

### 4.2 Secondary 의존

| 파일 | 의존 사유 |
|------|----------|
| SPEC-SKILL-TEST-001 framework | sample prompt + metric 측정 재사용 |
| `.claude/skills/moai/builder-skill/` | optimizer 기능을 skill body로 노출 |

### 4.3 Templates (Template-First Rule 준수)

- `internal/template/templates/.claude/agents/moai/builder-skill.md` 동기화
- `internal/template/templates/.claude/rules/moai/development/skill-description.md` 동기화

### 4.4 optimization workflow

```
사용자 호출: moai skill optimize moai-workflow-spec
  ↓
builder-skill 활성화
  ↓
Step 1: SPEC-SKILL-TEST-001 framework로 현재 description 측정
  - false_positive_rate (FP) = X
  - false_negative_rate (FN) = Y
  ↓
Step 2: 임계값 평가
  - if FP > 15%: 'tightening 필요' → narrower description 제안
  - if FN > 10%: 'broadening 필요' → wider description 제안
  - else: '최적화됨' 메시지
  ↓
Step 3: before/after diff 생성
  ↓
Step 4: 사용자 승인 → frontmatter 업데이트
  ↓
Step 5: regression test 재실행 → improvement 측정
  ↓
Step 6: optimization report 작성 (`.moai/reports/skill-optimization-<NAME>-<DATE>.md`)
```

---

## 5. 위험 및 가정

### 5.1 Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| sample prompts 부족 → metric 신뢰도 낮음 | High | High | minimum sample size (e.g., 10 prompts per skill) 미달 시 optimization 보류 |
| LLM 기반 description 생성이 부정확 | Medium | High | suggested description은 사람 검토 필수, auto-apply 금지 |
| optimization 후 다른 skill의 routing 영향 (cross-effect) | Medium | High | full skill catalog regression run 의무 |
| 임계값 15%/10%가 모든 skill에 부적합 | Medium | Medium | per-skill override 가능 (frontmatter `optimization_thresholds:`) |
| Wave 2 의존성 미완성 시 차단 | Low | Critical | blockedBy 명시 + 의존 SPEC 완료 확인 게이트 |

### 5.2 Assumptions

- A1: SPEC-SKILL-TEST-001 framework 안정 작동
- A2: 각 skill에 적어도 5-10개 sample prompts 존재
- A3: builder-skill agent가 description rewriting LLM call 가능 (Opus 또는 advisor 활용)
- A4: 임계값 15%/10%는 baseline 출발점, 측정 후 조정
- A5: optimization은 read-then-write, 사용자 승인 게이트 필수

---

## 6. 측정 계획 (Baseline + Validation)

| Metric | Baseline 측정 방법 | 목표 |
|--------|-------------------|------|
| optimization 정확도 (suggested 수용률) | 5 skill 시범 적용 | >= 60% acceptance |
| FP 감소 | tightening 적용 후 재측정 | -50% 이상 감소 |
| FN 감소 | broadening 적용 후 재측정 | -50% 이상 감소 |
| cross-effect (다른 skill) | full regression run | <= 5% routing 변동 |
| optimization 시간 | 단일 skill | < 60s |

---

## 7. 대안 검토 (Alternatives Considered)

| 대안 | 채택 여부 | 이유 |
|------|----------|------|
| auto-apply (사람 승인 없이) | ❌ | description은 user-facing, hallucination 위험 |
| skill body까지 동시 최적화 | ❌ | scope 폭증, 별도 SPEC |
| LLM 호출 없이 rule-based suggestion | ❌ | 제안 품질 한계, LLM이 contextual rewriting 우월 |
| Wave 2 의존 제거하고 자체 framework 신설 | ❌ | 중복 인프라 — 의존성 유지가 정도 |
| description 외 metadata 최적화 (allowed-tools, etc.) | ❌ | 별도 SPEC 후보, scope 분리 |

---

## 8. 참고 SPEC

- SPEC-SKILL-TEST-001 (Wave 2, blockedBy): regression test framework
- SPEC-SKILL-002: skill 시스템 기반
- SPEC-SKILL-ENHANCE-001: skill 향상 패턴
- SPEC-SKILL-GATE-001: skill 품질 게이트

---

## 9. Open Questions (Plan 단계 해결 대상)

- OQ1: optimization 호출 빈도 (분기별 / 변경 시 / 명시 호출만)?
- OQ2: 임계값을 글로벌 default + per-skill override 중 어떤 구조로?
- OQ3: optimization report의 보존 정책?
- OQ4: 같은 skill에 대한 반복 optimization 시 무한 루프 방지 (e.g., 직전 변경과 동일 제안)?

---

End of research.md (SPEC-SKILL-DESC-001).
