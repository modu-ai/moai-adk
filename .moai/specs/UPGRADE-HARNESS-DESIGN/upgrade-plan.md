# MoAI-ADK v3.0 업그레이드 계획서: Harness Design + SDD 베스트 프랙티스

> **문서 유형**: 전략적 업그레이드 계획서
> **기반 자료**: [Anthropic Harness Design](https://www.anthropic.com/engineering/harness-design-long-running-apps) + [2026 SDD 업계 분석](https://medium.com/@visrow/spec-driven-development-is-eating-software-engineering-a-map-of-30-agentic-coding-frameworks-6ac0b5e2b484)
> **작성일**: 2026-04-01
> **버전**: 3.0.0
> **대상**: MoAI-ADK v2.8.4 → v3.0.0

---

## Executive Summary

Anthropic 엔지니어링 블로그 Harness Design 패턴과 2026 SDD(Spec-Driven Development) 업계 베스트 프랙티스를 통합 분석하여, MoAI-ADK의 에이전트 오케스트레이션과 SPEC 워크플로우를 동시에 업그레이드하는 계획서입니다.

### 현재 상태 평가

| 구분 | Harness 패턴 (10개) | SDD 베스트 프랙티스 (8개) |
|------|-------------------|------------------------|
| 완전 구현 | 6개 (60%) | 5개 (62%) |
| 부분 구현 | 3개 (30%) | 3개 (38%) |
| 초과 달성 | 1개 (10%) | — |

### 핵심 Gap 5가지

**Harness Design 기반:**
1. **독립적 회의적 Evaluator 부재** — 자가 평가 함정(self-evaluation trap)에 취약
2. **Sprint Contract 사전 협상 없음** — Evaluator가 사후 검증만 수행
3. **Harness Depth 자동 판단 없음** — 복잡도 기반 자동 조절 부재

**SDD 베스트 프랙티스 기반:**
4. **Negative Constraints 부재** — "무엇을 만들지 않을 것인가"가 구조적으로 강제되지 않음
5. **Spec Drift 실시간 감지 없음** — 사후 보고만 존재, 구현 중 이탈 실시간 감지 부재

### 예상 효과

| 지표 | 현재 | 목표 | 개선폭 |
|------|------|------|--------|
| 병합 전 이슈 감지율 | ~60% | 95% | +35% |
| Scope Creep 발생률 | 측정 불가 | < 10% | 구조적 방지 |
| 비용 최적화 (단순 작업) | 1.0x | 0.3x | -70% |
| SPEC ↔ 구현 일치율 | ~80% | > 95% | +15% |

---

## 1. Gap 분석: Harness Design 패턴

### 1.1 GAN형 Generator-Evaluator 분리

**블로그 핵심**: 평가자를 회의적(skeptical)으로 튜닝하면 자가 비판보다 훨씬 효과적.

| 항목 | MoAI-ADK 현재 | Gap |
|------|-------------|-----|
| Generator 분리 | manager-ddd/tdd ✅ | 없음 |
| Evaluator 분리 | manager-quality (haiku) | **모델 레벨 부족, 회의적 튜닝 부재** |
| 독립적 평가 컨텍스트 | 동일 오케스트레이터에 보고 | **평가 독립성 미확보** |

**Gap 심각도**: 🔴 높음

### 1.2 Three-Agent Pipeline + Planner What/Why

**블로그 핵심**: Planner는 What/Why에 집중, How는 Generator 영역.

| 항목 | MoAI-ADK 현재 | Gap |
|------|-------------|-----|
| Planner → Generator → Evaluator | manager-spec → manager-ddd/tdd → manager-quality ✅ | 파이프라인 존재 |
| Planner What/Why 집중 | EARS 형식 (부분적) | **How 과명세 방지 규칙 없음** |

**Gap 심각도**: 🟡 중간

### 1.3 Sprint Contracts

**블로그 핵심**: 구현 시작 '이전(Before)'에 완료 기준을 상호 협상.

| 항목 | MoAI-ADK 현재 | Gap |
|------|-------------|-----|
| Acceptance Criteria | SPEC EARS ✅ | 없음 |
| 사전 협상(Negotiation) | 없음 | **Evaluator의 사전 개입 없음** |

**Gap 심각도**: 🟡 중간

### 1.4 Context Management

**Opus 4.6**: 1M 토큰 + auto-compaction으로 context reset 필요성 대폭 감소.

**Gap 심각도**: 🟢 낮음 — 세션 복구용 progress.md 보완만 필요

### 1.5 에이전트 간 통신

**MoAI-ADK**: Sub-agent (Agent() 반환) + Team (SendMessage + TaskList) + CG (tmux 격리) 하이브리드.

**Gap 심각도**: 🔵 초과 달성

### 1.6 Cost-Quality Tradeoff

| 항목 | MoAI-ADK 현재 | Gap |
|------|-------------|-----|
| CG Mode | moai cc/glm/cg ✅ | 없음 |
| 하네스 깊이 자동 판단 | 없음 | **복잡도 기반 자동 조절 불가** |

**Gap 심각도**: 🔴 높음

---

## 2. Gap 분석: SDD 베스트 프랙티스

### 2.1 Negative Constraints (What NOT to Build)

> "무엇을 만들지 않을 것인가가 무엇을 만들 것인가만큼 중요하다" — 2026 SDD 합의

| 항목 | MoAI-ADK 현재 | Gap |
|------|-------------|-----|
| EARS Unwanted 패턴 | 정의는 있음 | **구조적 강제 없음** |
| Exclusions 섹션 | 없음 | **spec.md에 필수 섹션 부재** |
| Quality Gate 검증 | 없음 | **Unwanted 0개여도 통과** |

**Gap 심각도**: 🔴 높음 — scope creep의 구조적 원인

### 2.2 Persistent TASKS.md

**업계 표준**: SPEC.md / ARCHITECTURE.md / TASKS.md 3파일 아티팩트를 Git에 버전 관리.

| 항목 | MoAI-ADK 현재 | Gap |
|------|-------------|-----|
| spec.md + plan.md + acceptance.md | ✅ 3파일 구조 | 없음 |
| 태스크 영속화 | TaskCreate (인메모리, ephemeral) | **완료 후 감사 추적 불가** |

**Gap 심각도**: 🟡 중간

### 2.3 Spec Drift 실시간 감지

| 항목 | MoAI-ADK 현재 | Gap |
|------|-------------|-----|
| Re-planning Gate | Phase 2.7 (acceptance criteria 기반) ✅ | 없음 |
| Divergence Tracking | Phase 2A/2B 사후 보고 ✅ | 없음 |
| 실시간 Drift Guard | 없음 | **DDD/TDD cycle 중 planned vs actual 실시간 비교 없음** |

**Gap 심각도**: 🟡 중간

### 2.4 Constitution 자동 검증

| 항목 | MoAI-ADK 현재 | Gap |
|------|-------------|-----|
| tech.md 문서화 | ✅ | 없음 |
| 기계 판독 가능 형식 | 없음 | **constitution.yaml 부재** |
| SPEC ↔ Constitution 자동 검증 | 없음 | **위반 자동 감지 불가** |

**Gap 심각도**: 🟡 중간

### 2.5 Delta Markers for Brownfield

| 항목 | MoAI-ADK 현재 | Gap |
|------|-------------|-----|
| research.md | 기존 코드 분석 ✅ | 없음 |
| @MX tags | 코드 레벨 컨텍스트 ✅ | 없음 |
| SPEC 레벨 delta 구분 | 없음 | **기존 동작 vs 새 동작 미구분** |

**Gap 심각도**: 🟢 낮음 — DDD ANALYZE에서 부분 커버

### 2.6 Token-Efficient SPEC

| 항목 | MoAI-ADK 현재 | Gap |
|------|-------------|-----|
| 3-Phase 토큰 예산 | 30K/180K/40K ✅ | 없음 |
| Progressive Disclosure | Level 1/2/3 ✅ | 없음 |
| Run Phase SPEC 로딩 | 전체 로딩 | **compact 버전 미지원** |

**Gap 심각도**: 🟢 낮음

---

## 3. 아키텍처 개선 제안

### 3.1 제안 A: 독립적 회의적 Evaluator (evaluator-active)

**새 에이전트**: `.claude/agents/moai/evaluator-active.md`

```yaml
name: evaluator-active
description: >
  Skeptical code evaluator. Tuned toward finding defects, not rationalizing acceptance.
tools: Read, Grep, Glob, Bash, mcp__sequential-thinking__sequentialthinking
model: sonnet
permissionMode: plan
```

**핵심 프롬프트**: 수용 합리화 금지, concrete evidence 없이 PASS 불가, UNVERIFIED 구분.

**평가 차원**: Functionality(40%) / Security(25%) / Craft(20%) / Consistency(15%). Security FAIL = 전체 FAIL.

**모드별 배치**: Sub-agent → Agent() 호출 | Team → reviewer teammate | CG → Leader가 직접 수행

**개입 모드**: final-pass (standard, Opus 4.6+ 기본) | per-sprint (thorough)

**Phase 2.8 변경**: Phase 2.8a (evaluator-active 능동 테스트) → Phase 2.8b (manager-quality TRUST 5 정적 검증)

### 3.2 제안 B: Harness Depth 자동 판단

**새 파일**: `.moai/config/sections/harness.yaml`

```yaml
harness:
  mode_defaults:
    solo: auto
    team: auto
    cg: thorough          # CG = Generator-Evaluator 자연 분리

  auto_detection:
    minimal_when: "file_count <= 3 AND single_domain"
    standard_when: "file_count > 3 OR multi_domain"
    thorough_when: "security/payment keywords OR priority == critical"

  escalation:
    enabled: true
    max_escalations: 2    # minimal → standard → thorough
```

**3개 레벨**: minimal (skip 불필요 phases) / standard (evaluator final-pass) / thorough (sprint contract + per-sprint evaluator)

### 3.3 제안 C: Sprint Contract Negotiation (Phase 2.0)

구현 시작 전 evaluator-active가 manager-ddd의 구현 계획을 사전 리뷰. `contract.md`로 합의 기록. thorough 전용.

### 3.4 제안 D: CG Mode Generator-Evaluator 매핑

Leader(Claude) = Evaluator, Teammates(GLM) = Generators. 모델 분리로 자가 평가 함정 원천 차단. thorough 품질을 standard 비용으로 달성.

### 3.5 제안 E: Planner What/Why 제약 강화

manager-spec 프롬프트에 "How 유보" 규칙 추가. SPEC은 What + Why 집중, How는 manager-ddd 영역.

### 3.6 제안 F: Negative Constraints 필수화

**spec.md 템플릿 변경**: `## Exclusions (What NOT to Build)` 섹션 필수 추가.

```markdown
## Exclusions (What NOT to Build)
- Shall NOT support [feature X] (reason: [A])
- Shall NOT implement [pattern Y] (reason: [B])
- Will NOT be optimized for [use case Z]
```

**Phase 3.6 Quality Gate**: Exclusions 최소 1개 검증. 구현 에이전트 프롬프트에 exclusion list를 `[HARD] DO NOT implement` 형태로 주입.

### 3.7 제안 G: Persistent TASKS.md

Phase 1.5 output으로 `.moai/specs/SPEC-{ID}/tasks.md` 자동 생성.

```markdown
## Task Decomposition
SPEC: {SPEC-ID}
Generated: {timestamp}

| Task ID | Description | Requirement | Dependencies | Status |
|---------|-------------|-------------|--------------|--------|
| TASK-001 | {description} | REQ-001 | - | pending |
| TASK-002 | {description} | REQ-002 | TASK-001 | pending |
```

### 3.8 제안 H: Spec Drift Guard

DDD/TDD 각 cycle 완료 시 planned_files vs actual_files 자동 비교. Plan 대비 +20% 초과 시 경고. Run Phase 2A/2B에 drift check 단계 추가.

### 3.9 제안 I: Constitution 자동 검증

**새 파일**: `.moai/project/constitution.yaml`

```yaml
constitution:
  approved_languages: [go]
  approved_frameworks: [cobra, viper]
  forbidden_patterns:
    - "global mutable state"
    - "init() with side effects"
  naming_conventions:
    packages: "lowercase, single word"
    exported: "PascalCase"
```

Phase 3.6 SPEC Quality Gate에서 tech.md/constitution.yaml 대비 자동 검증.

### 3.10 제안 J: Delta Markers for Brownfield

spec.md에서 brownfield 요구사항에 delta marker 사용:

```markdown
### [DELTA] Authentication Module
- [EXISTING] JWT token generation (unchanged)
- [MODIFY] Token expiration: 1h → 24h
- [NEW] Refresh token endpoint
- [REMOVE] Basic auth fallback
```

DDD ANALYZE 단계에서 `[EXISTING]`은 characterization test만, `[NEW]`은 새 구현.

### 3.11 제안 K: Token-Efficient SPEC (Compact)

Run phase 전용 `spec-compact.md` 자동 생성 (EARS 요구사항 + acceptance criteria만 추출, ~30% 토큰 절약). Phase별 선택적 SPEC 로딩.

### 3.12 제안 L: 평가자 프롬프트 라이브러리

`.moai/config/evaluator-profiles/` 디렉토리에 default.md / strict.md / lenient.md / frontend.md 프로필. 반복 개선 워크플로우로 False Positive/Negative 분석 후 프로필 업데이트.

---

## 4. 구현 로드맵

### Phase 0 (즉시 적용) — 설정 변경만으로 가능

| 항목 | 변경 대상 | 변경 내용 | 기대 효과 |
|------|----------|----------|----------|
| Quality 모델 업그레이드 | `manager-quality.md` | Phase 2.8 전용 `model: sonnet` | 리뷰 품질 즉시 향상 |
| 회의적 프롬프트 추가 | `manager-quality.md` | "수용 합리화 금지" 지시문 | QA 과다 통과 방지 |
| Planner What/Why 제약 | `manager-spec.md` | "How 유보" 규칙 추가 (제안 E) | 다운스트림 오류 방지 |
| **Negative Constraints** | `plan.md` Phase 2 템플릿 | **Exclusions 섹션 필수화 (제안 F)** | scope creep 구조적 방지 |
| **Constitution YAML** | `.moai/project/constitution.yaml` | **기계 판독 가능 tech 제약 (제안 I)** | 아키텍처 일관성 보장 |

**예상 소요**: 4-6시간

### Phase 1 (단기, SPEC 3개)

| SPEC ID | 제목 | 제안 | 핵심 산출물 | 의존성 |
|---------|------|------|-----------|--------|
| SPEC-EVAL-001 | evaluator-active 에이전트 | A, C, D | `evaluator-active.md`, Phase 2.0 + 2.8a, 모드별 배치 | P0 완료 |
| SPEC-DRIFT-001 | Spec Drift Guard + TASKS.md | G, H | `tasks.md` 템플릿, drift check 로직, Phase 2A/2B 수정 | 없음 |
| SPEC-SDD-001 | SDD 통합 (Delta + Compact) | B, J, K | `harness.yaml`, delta marker, `spec-compact.md` 생성 | 없음 |

**핵심 검증 포인트**:
- SPEC-EVAL-001: evaluator-active가 manager-quality 대비 실제로 더 많은 이슈 발견하는지 A/B 비교
- SPEC-DRIFT-001: drift guard가 scope creep을 Phase 2 중간에 감지하는지 확인
- SPEC-SDD-001: delta marker가 brownfield 구현 정확도를 향상시키는지 측정

### Phase 2 (중기, SPEC 3개)

| SPEC ID | 제목 | 제안 | 의존성 |
|---------|------|------|--------|
| SPEC-EVALLIB-001 | 평가자 프롬프트 라이브러리 | L | SPEC-EVAL-001 |
| SPEC-PLAYWRIGHT-001 | Playwright 능동 테스트 통합 | — | SPEC-EVAL-001 |
| SPEC-SEMAP-001 | SEMAP Behavioral Contracts | — | SPEC-EVAL-001 |

**SPEC-SEMAP-001 상세**: 각 agent 정의에 `contract` 섹션 추가 (preconditions, postconditions, invariants, forbidden). Phase 완료 시 postcondition 자동 검증. 먼저 manager-ddd와 manager-quality에만 적용 후 효과 검증 → 확대.

---

## 5. 리스크 평가

### 기술적 리스크

| 리스크 | 영향 | 확률 | 대응 전략 |
|--------|------|------|----------|
| Evaluator 토큰 비용 폭증 | 중 | 높음 | auto 모드에서 복잡도 기반 활성화, CG Mode에서는 추가 비용 0 |
| Evaluator-Generator 무한 루프 | 높 | 중 | 최대 3회 수정 사이클 제한 |
| Sprint Contract 협상 지연 | 중 | 중 | 협상 최대 2회전 제한, 합의 실패 시 사용자 판단 위임 |
| Drift Guard False Positive | 중 | 중 | 임계값 +20% 이상만 경고, 경고 즉시 차단이 아닌 알림 |
| Negative Constraints 과도 | 낮 | 중 | 최소 1개 ~ 최대 10개 권장 가이드라인 |

### 하위 호환성

| 항목 | 보장 방법 |
|------|----------|
| 기존 quality.yaml | harness.yaml는 독립 파일, quality.yaml 변경 없음 |
| 기존 SPEC 파이프라인 | 모든 새 기능은 opt-in, 기본값은 현재 동작 유지 |
| 기존 spec.md 템플릿 | Exclusions 섹션은 신규 SPEC부터 적용, 기존 SPEC 소급 불필요 |
| CG Mode / GLM 호환 | evaluator-active는 `model: sonnet` → GLM 모드에서 GLM sonnet 사용 |

### 마이그레이션 전략

```
1. Zero-break: 모든 새 파일/설정은 선택적. 기존 프로젝트 영향 제로
2. 점진적 채택: P0 → P1 → P2 순서로 검증 후 적용
3. moai update 연동: 템플릿 동기화 시 harness.yaml, constitution.yaml 자동 생성
4. 롤백: 모든 기능 비활성화 가능 (evaluator: false, sprint_contract: false)
```

---

## 6. 아키텍처 변경 요약

### 변경 대상 파일

| 파일 | 변경 유형 | Phase | 근거 |
|------|----------|-------|------|
| `.claude/agents/moai/evaluator-active.md` | 신규 생성 | P1 | 제안 A |
| `.claude/agents/moai/manager-quality.md` | 프롬프트 강화 (sonnet) | P0 | 제안 A |
| `.claude/agents/moai/manager-spec.md` | What/Why 제약 + Exclusions | P0 | 제안 E, F |
| `.claude/skills/moai/workflows/run.md` | Phase 2.0 + 2.8a/b + drift guard | P1 | 제안 A, C, H |
| `.claude/skills/moai/workflows/moai.md` | Complexity Estimator | P1 | 제안 B |
| `.claude/skills/moai/workflows/plan.md` | Exclusions 필수 + delta marker | P0/P1 | 제안 F, J |
| `.moai/config/sections/harness.yaml` | 신규 생성 | P1 | 제안 B |
| `.moai/project/constitution.yaml` | 신규 생성 | P0 | 제안 I |
| `.moai/config/evaluator-profiles/` | 신규 디렉토리 | P2 | 제안 L |

### 신규 에이전트

| 에이전트 | 역할 | 모델 | 모드 |
|---------|------|------|------|
| evaluator-active | 회의적 독립 평가자 | sonnet | plan (읽기 전용) |

### 신규 SPEC 산출물

| 파일 | 용도 | Phase |
|------|------|-------|
| contract.md | Sprint Contract 합의 (thorough) | P1 |
| tasks.md | 영속 태스크 목록 (Git 버전 관리) | P1 |
| spec-compact.md | Run phase 토큰 최적화 버전 | P1 |

---

## 7. 성공 지표

### 정량 지표

| 지표 | 현재 기준 | P1 목표 | P2 목표 | 측정 방법 |
|------|----------|---------|---------|----------|
| 병합 전 이슈 감지율 | ~60% | 80% | 95% | evaluator FAIL 대비 실제 이슈 비율 |
| QA False Negative | 측정 불가 | < 15% | < 5% | 프로덕션 버그 역추적 |
| Scope Creep 발생률 | 측정 불가 | < 15% | < 10% | Exclusions 위반 + drift guard 경고 횟수 |
| SPEC ↔ 구현 일치율 | ~80% | 90% | > 95% | tasks.md 완료율 + drift report |
| 단순 작업 비용 절감 | 1.0x | 0.5x | 0.3x | auto minimal 판단 시 토큰 사용량 |

### 정성 지표

- evaluator-active가 구체적이고 실행 가능한 피드백 제공
- Negative Constraints가 불필요한 기능 구현을 사전 방지
- Delta Markers가 brownfield 프로젝트에서 구현 정확도 향상
- tasks.md가 SPEC 완료 후 감사 추적 및 회고에 활용

---

## 8. 참고 자료

| 자료 | 링크/경로 |
|------|----------|
| Anthropic Harness Design | https://www.anthropic.com/engineering/harness-design-long-running-apps |
| SDD 30+ Framework Map | https://medium.com/@visrow/spec-driven-development-is-eating-software-engineering-a-map-of-30-agentic-coding-frameworks-6ac0b5e2b484 |
| Thoughtworks SDD | https://www.thoughtworks.com/en-us/insights/blog/agile-engineering-practices/spec-driven-development-unpacking-2025-new-engineering-practices |
| SEMAP Protocol | https://arxiv.org/html/2510.12120 |
| Claude Code Best Practices | https://code.claude.com/docs/en/best-practices |
| MIT Missing Semester: Agentic Coding | https://missing.csail.mit.edu/2026/agentic-coding/ |
| MoAI-ADK SPEC 워크플로우 | `.claude/rules/moai/workflow/spec-workflow.md` |
| MoAI-ADK 실행 지침 | `CLAUDE.md` |
| TRUST 5 프레임워크 | `.claude/rules/moai/core/moai-constitution.md` |

---

## Appendix A: 전체 개선 우선순위 매트릭스

| 순위 | 개선 항목 | 출처 | 영향도 | 실행 가능성 | ROI | Phase |
|------|----------|------|--------|-----------|-----|-------|
| 1 | Negative Constraints 필수화 | SDD | 10 | 10 | 100 | **P0** |
| 2 | Constitution 자동 검증 | SDD | 7 | 9 | 63 | **P0** |
| 3 | GAN형 Evaluator 분리 | Harness | 10 | 7 | 70 | P1 |
| 4 | Sprint Contract 사전 협상 | Harness | 8 | 8 | 64 | P1 |
| 5 | Persistent TASKS.md | SDD | 8 | 9 | 72 | P1 |
| 6 | Spec Drift Guard | SDD | 8 | 7 | 56 | P1 |
| 7 | Delta Markers (Brownfield) | SDD | 7 | 7 | 49 | P1 |
| 8 | Token-Efficient SPEC | SDD | 6 | 8 | 48 | P1 |
| 9 | Harness Depth 자동 판단 | Harness | 7 | 8 | 56 | P1 |
| 10 | CG Mode Evaluator 매핑 | Harness | 8 | 9 | 72 | P1 |
| 11 | Planner What/Why 제약 | Harness+SDD | 6 | 10 | 60 | **P0** |
| 12 | 평가자 프롬프트 라이브러리 | Harness | 7 | 7 | 49 | P2 |
| 13 | Playwright 능동 테스트 | Harness | 9 | 5 | 45 | P2 |
| 14 | SEMAP Behavioral Contracts | SDD | 9 | 4 | 36 | P2 |

---

## Appendix B: 버전 변경 이력

### v2.1 → v3.0 변경사항

| 변경 | 이유 |
|------|------|
| SPEC-HARNESS-001 제거 | Harness depth는 자동 판단으로 SPEC-SDD-001에 통합 |
| SPEC-HANDOFF-001 제거 | 1M 토큰 시대에 handoff 전용 SPEC 불필요. progress.md 보완으로 충분 |
| SPEC-METRICS-001 제거 | 독립 메트릭 SPEC 불필요. evaluator-active 로그로 대체 |
| Phase 3 (장기 로드맵) 전체 제거 | Harness Evolution 등 장기 항목은 모델 진화 시 자연 해결 |
| SDD Gap 분석 추가 (Section 2) | 2026 SDD 베스트 프랙티스 6개 항목 gap 분석 |
| 제안 F-L 추가 (7개) | Negative Constraints, TASKS.md, Drift Guard, Constitution, Delta Markers, Compact SPEC, Evaluator Profiles |
| 로드맵 재구성 | P0 5개 + P1 SPEC 3개 + P2 SPEC 3개로 통합 재배치 |

### v1.0 → v2.1 변경사항

| 변경 | 이유 |
|------|------|
| CLI → Claude Code Skill 명령 | `/moai`는 터미널 CLI가 아닌 Skill 명령 |
| 통신 모델 재작성 | 하이브리드 통신 (SendMessage + TaskList + 파일) |
| Context Reset → Compaction-First | Opus 4.6 1M 토큰 + auto-compaction |
| CG Mode 제안 신설 | Generator-Evaluator 자연 분리 |
| Sprint Contract 신설 | Evaluator 사전 협상 |
| `--harness` 플래그 제거 | 시스템 자동 판단으로 전환 |

---

*이 계획서는 Anthropic Harness Design + 2026 SDD 업계 베스트 프랙티스 + 사용자 피드백을 통합하여 MoAI-ADK의 전략적 업그레이드 방향을 제시합니다. 모든 변경사항은 하위 호환성을 보장하며, 점진적으로 적용 가능합니다.*
