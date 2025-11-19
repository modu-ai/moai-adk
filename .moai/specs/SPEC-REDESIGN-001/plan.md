---
spec_id: SPEC-REDESIGN-001
title: 프로젝트 설정 스키마 v3.0.0 리디자인 - 구현 계획
version: 1.0.0
created_at: 2025-11-19T00:00:00Z
---

# 구현 계획: SPEC-REDESIGN-001

## 프로젝트 개요

**목표**: /moai:0-project 명령어의 설정 경험을 2-3분(Quick Start) 또는 5-30분(Full Docs)으로 단축하고, 프로젝트 문서를 통합하여 AI 에이전트의 컨텍스트 이해를 향상

**성공 지표**:
- ✅ 설정 시간 단축: 15-20분 → 2-30분 (선택에 따라)
- ✅ 질문 감소: 27개 → 10개 essential (63% 감소)
- ✅ 설정 커버리지: 100% (31개 모두)
- ✅ 문서 자동 생성: product.md, structure.md, tech.md
- ✅ AI 에이전트 컨텍스트 통합: 3개 에이전트에 자동 로딩

---

## Phase 별 구현 계획

### Phase 1: 문서 생성 Skill 개발 (2-3일)

#### 1.1 moai-project-documentation-generator Skill 구조 설계

**목표**: 프로젝트 문서 자동 생성 Skill 개발

**작업 항목**:

| 작업 | 상세 | 담당 | 소요시간 |
|-----|------|------|---------|
| 1.1.1 | Skill 메타데이터 정의 (SKILL.md) | backend-expert | 2시간 |
| 1.1.2 | Brainstorm 질문 세트 작성 (Quick/Standard/Deep) | backend-expert | 4시간 |
| 1.1.3 | AI 분석 엔진 설계 | backend-expert | 3시간 |
| 1.1.4 | 문서 생성 템플릿 정의 | docs-manager | 2시간 |

**핵심 기능**:
- Quick (5-10분): 핵심 질문 5-8개
  - 프로젝트 비전 (1개)
  - 기본 아키텍처 (1개)
  - 필수 기술 (1개)
  - 팀 구성 (1개)
  - 타임라인 (1개)

- Standard (15분): 상세 질문 10-15개
  - 위 5개 + Trade-off 분석 (2개)
  - 성능 요구사항 (1개)
  - 보안 고려사항 (1개)
  - 확장성 전략 (1개)

- Deep (30분): 포괄적 질문 20-25개
  - 위 15개 + 경쟁사 분석 (2개)
  - 시장 조사 (2개)
  - 혁신 전략 (1개)
  - 비용 예측 (1개)
  - Best Practices (1개)

**산출물**:
- `.claude/skills/moai-project-documentation-generator/SKILL.md`
- `.claude/skills/moai-project-documentation-generator/brainstorm_questions.json`

#### 1.2 AI 분석 파이프라인 구현

**작업 항목**:

| 작업 | 상세 | 담당 | 소요시간 |
|-----|------|------|---------|
| 1.2.1 | Context7 MCP 통합 (최신 트렌드) | backend-expert | 3시간 |
| 1.2.2 | WebSearch MCP 통합 (경쟁사 분석) | backend-expert | 2시간 |
| 1.2.3 | 문서 생성 프롬프트 최적화 | backend-expert | 3시간 |

**세부 구현**:
- product.md 생성: 프로젝트 비전, 타겟 사용자, USP, 로드맵
- structure.md 생성: 아키텍처 다이어그램, 컴포넌트 설명, 의존성
- tech.md 생성: 기술 스택 선택 이유, Trade-offs, 초기 설정

**산출물**:
- Brainstorm 자동화 파이프라인
- 3개 문서 자동 생성 템플릿

---

### Phase 2: Tab 스키마 완성 및 검증 (1-2일)

#### 2.1 Tab 스키마 v3.0.0 검증

**목표**: 기존 tab_schema.json v3.0.0 구조 검증 및 완성

**작업 항목**:

| 작업 | 상세 | 담당 | 소요시간 |
|-----|------|------|---------|
| 2.1.1 | JSON 스키마 정합성 검증 | backend-expert | 2시간 |
| 2.1.2 | API 제약 준수 확인 (4개 질문/배치) | backend-expert | 1시간 |
| 2.1.3 | Template 변수 문법 검증 | backend-expert | 1시간 |
| 2.1.4 | Conditional 렌더링 로직 검증 | backend-expert | 2시간 |

**검증 항목**:
- ✅ Tab 1: 4개 배치, 10개 질문
- ✅ Tab 2: 2개 배치 (1개 필수, 1개 조건부)
- ✅ Tab 3: 2개 배치 (조건부, 모드별)
- ✅ Smart defaults: 16개 설정값
- ✅ Auto-detect: 5개 필드
- ✅ 100% 커버리지 (31개)

**산출물**:
- 검증 리포트
- Final tab_schema.json v3.0.0

#### 2.2 설정값 매핑 정의

**작업 항목**:

| 작업 | 상세 | 담당 | 소요시간 |
|-----|------|------|---------|
| 2.2.1 | config.json 구조 매핑 | backend-expert | 2시간 |
| 2.2.2 | Smart defaults 규칙 정의 | backend-expert | 2시간 |
| 2.2.3 | Auto-detect 로직 정의 | backend-expert | 1시간 |

**산출물**:
- 설정값 매핑 문서
- Smart defaults 규칙 정의서
- Auto-detect 알고리즘 명세

---

### Phase 3: 에이전트 및 명령어 업데이트 (1-2일)

#### 3.1 /moai:0-project 명령어 업데이트

**목표**: /moai:0-project 명령어를 새로운 Tab 스키마 및 문서 생성 기능과 통합

**작업 항목**:

| 작업 | 상세 | 담당 | 소요시간 |
|-----|------|------|---------|
| 3.1.1 | 명령어 로직 업데이트 | project-manager | 3시간 |
| 3.1.2 | Tab 스키마 로드 및 검증 | project-manager | 1시간 |
| 3.1.3 | 문서 생성 워크플로우 통합 | docs-manager | 2시간 |
| 3.1.4 | 에러 처리 및 복구 로직 | backend-expert | 1시간 |

**세부 구현**:
- Tab 순서대로 사용자 질문 진행
- 각 배치 완료 후 다음 배치로 자동 진행
- Tab 2 "Full Documentation" 선택 시 → moai-project-documentation-generator 위임
- Tab 3 조건부 배치 표시 (git_strategy.mode 기반)
- 최종 config.json 저장 (원자성 보장)

**산출물**:
- 업데이트된 0-project.md 명령어 정의

#### 3.2 project-manager 에이전트 업데이트

**목표**: 생성된 프로젝트 문서를 에이전트에 자동 로딩

**작업 항목**:

| 작업 | 상세 | 담당 | 소요시간 |
|-----|------|------|---------|
| 3.2.1 | product.md 로드 메커니즘 | project-manager | 1시간 |
| 3.2.2 | Task prompt 확장 | project-manager | 1시간 |
| 3.2.3 | 컨텍스트 주입 로직 | project-manager | 1시간 |

**통합 방식**:
```
product.md (프로젝트 비전)
    ↓
project-manager.system_context에 추가
    ↓
All tasks에 자동 주입
```

**산출물**:
- 업데이트된 project-manager.md 에이전트 정의

#### 3.3 tdd-implementer 에이전트 업데이트

**목표**: structure.md를 architecture reference로 제공

**작업 항목**:

| 작업 | 상세 | 담당 | 소요시간 |
|-----|------|------|---------|
| 3.3.1 | structure.md 로드 메커니즘 | tdd-implementer | 1시간 |
| 3.3.2 | 아키텍처 컨텍스트 통합 | tdd-implementer | 1시간 |

**산출물**:
- 업데이트된 tdd-implementer.md 에이전트 정의

#### 3.4 Domain Experts 업데이트 (backend-expert, frontend-expert)

**목표**: tech.md를 기술 스택 참고 자료로 제공

**작업 항목**:

| 작업 | 상세 | 담당 | 소요시간 |
|-----|------|------|---------|
| 3.4.1 | tech.md 로드 메커니즘 | backend-expert | 1시간 |
| 3.4.2 | tech.md 로드 메커니즘 | frontend-expert | 1시간 |
| 3.4.3 | 기술 스택 선택 컨텍스트 | backend/frontend experts | 1시간 |

**산출물**:
- 업데이트된 backend-expert.md, frontend-expert.md

---

### Phase 4: 문서 및 CLAUDE.md 업데이트 (2-3일)

#### 4.1 CLAUDE.md 업데이트

**목표**: 새로운 설정 시스템 및 문서 통합 기능 문서화

**작업 항목**:

| 작업 | 상세 | 담당 | 소요시간 |
|-----|------|------|---------|
| 4.1.1 | /moai:0-project 기능 설명 추가 | docs-manager | 2시간 |
| 4.1.2 | 프로젝트 문서 생성 워크플로우 설명 | docs-manager | 2시간 |
| 4.1.3 | AI 에이전트 컨텍스트 로딩 설명 | docs-manager | 1시간 |
| 4.1.4 | 빠른 시작 가이드 (Quick Start vs Full Docs) | docs-manager | 1시간 |

**추가 섹션**:
1. `/moai:0-project 신기능 (v3.0.0)`
   - Quick Start (2-3분)
   - Full Documentation (5-30분)
   - Brainstorming 깊이 선택
   - AI 분석 기반 자동 문서 생성

2. `프로젝트 문서 활용`
   - product.md: project-manager에서 사용
   - structure.md: tdd-implementer에서 사용
   - tech.md: domain experts에서 사용

3. `예시 워크플로우`
   - 신규 프로젝트: 0-project → Quick Start → 1-plan
   - 포괄적 설정: 0-project → Full Docs + Deep → 프로젝트 문서 생성

**산출물**:
- 업데이트된 CLAUDE.md
- 프로젝트 문서 활용 가이드 (.moai/docs/PROJECT_DOCS_USAGE.md)

#### 4.2 /moai:0-project 명령어 문서화

**작업 항목**:

| 작업 | 상세 | 담당 | 소요시간 |
|-----|------|------|---------|
| 4.2.1 | 명령어 기능 설명 | docs-manager | 1시간 |
| 4.2.2 | 각 모드별 가이드 작성 | docs-manager | 2시간 |
| 4.2.3 | 스크린샷/예시 추가 | docs-manager | 1시간 |

**산출물**:
- 0-project.md 명령어 문서
- 모드별 가이드 (.moai/docs/0-PROJECT-MODES.md)

#### 4.3 프로젝트 문서 템플릿 작성

**작업 항목**:

| 작업 | 상세 | 담당 | 소요시간 |
|-----|------|------|---------|
| 4.3.1 | product.md 템플릿 | docs-manager | 1시간 |
| 4.3.2 | structure.md 템플릿 | docs-manager | 1시간 |
| 4.3.3 | tech.md 템플릿 | docs-manager | 1시간 |

**산출물**:
- `.moai/project/product.md.template`
- `.moai/project/structure.md.template`
- `.moai/project/tech.md.template`

---

### Phase 5: 품질 검증 및 테스트 (1-2일)

#### 5.1 기능 검증 (Acceptance Testing)

**목표**: SPEC-REDESIGN-001 수용 기준 검증

**작업 항목**:

| 작업 | 상세 | 담당 | 소요시간 |
|-----|------|------|---------|
| 5.1.1 | Quick Start 모드 테스트 (2-3분) | qa-engineer | 1시간 |
| 5.1.2 | Full Documentation 모드 테스트 | qa-engineer | 2시간 |
| 5.1.3 | 조건부 배치 렌더링 테스트 | qa-engineer | 1시간 |
| 5.1.4 | config.json 저장 및 검증 | qa-engineer | 1시간 |
| 5.1.5 | 프로젝트 문서 생성 테스트 | qa-engineer | 1시간 |
| 5.1.6 | AI 에이전트 컨텍스트 로딩 테스트 | qa-engineer | 1시간 |

**테스트 항목**:
- ✅ Tab 1: 10개 질문 표시 (4개 배치)
- ✅ Tab 2: 1개 필수 질문 + 1개 조건부 질문
- ✅ Tab 3: git_strategy.mode에 따라 조건부 배치
- ✅ Smart defaults: 16개 설정값 자동 적용
- ✅ Auto-detect: 5개 필드 감지 및 설정
- ✅ config.json: 31개 설정값 완전 저장
- ✅ 프로젝트 문서: 3개 파일 생성
- ✅ 에이전트 컨텍스트: 자동 로드 확인

**산출물**:
- 테스트 케이스 작성 (test_cases.md)
- 테스트 실행 리포트

#### 5.2 성능 검증

**목표**: 설정 시간 목표 달성 확인

**작업 항목**:

| 작업 | 상세 | 담당 | 소요시간 |
|-----|------|------|---------|
| 5.2.1 | Quick Start 시간 측정 | qa-engineer | 30분 |
| 5.2.2 | Full Docs + Quick 시간 측정 | qa-engineer | 30분 |
| 5.2.3 | Full Docs + Standard 시간 측정 | qa-engineer | 1시간 |
| 5.2.4 | Full Docs + Deep 시간 측정 | qa-engineer | 1시간 |

**성공 기준**:
- Quick Start: 2-3분 이내 ✅
- Full + Quick: 5-10분 이내 ✅
- Full + Standard: 15-20분 이내 ✅
- Full + Deep: 25-35분 이내 ✅

**산출물**:
- 성능 측정 리포트

#### 5.3 호환성 검증

**목표**: v2.1.0 설정과의 호환성 확인

**작업 항목**:

| 작업 | 상세 | 담당 | 소요시간 |
|-----|------|------|---------|
| 5.3.1 | v2.1.0 config.json 마이그레이션 테스트 | backend-expert | 1시간 |
| 5.3.2 | 설정값 충돌 확인 | backend-expert | 30분 |
| 5.3.3 | 롤백 테스트 | backend-expert | 30분 |

**산출물**:
- 호환성 테스트 리포트

#### 5.4 사용자 수용 테스트 (UAT)

**목표**: 실제 사용자 경험 검증

**작업 항목**:

| 작업 | 상세 | 담당 | 소요시간 |
|-----|------|------|---------|
| 5.4.1 | 신규 사용자 테스트 (3명) | qa-engineer | 2시간 |
| 5.4.2 | 기존 사용자 테스트 (3명) | qa-engineer | 1시간 |
| 5.4.3 | 피드백 수집 및 개선 | qa-engineer | 1시간 |

**피드백 항목**:
- 질문 이해도
- 진행도 표시 명확성
- 예상 시간 정확성
- 문서 품질
- 설정 완료 후 개발 시작 용이성

**산출물**:
- UAT 리포트
- 개선 사항 목록

---

## 리소스 및 역할 할당

### 주요 역할

| 역할 | 담당자 | 책임 |
|------|--------|------|
| **Project Manager** | project-manager | 전체 일정 및 조율 |
| **Backend Expert** | backend-expert | Skill 개발, 에이전트 통합 |
| **Docs Manager** | docs-manager | 문서 생성, CLAUDE.md 업데이트 |
| **QA Engineer** | qa-engineer | 테스트 및 검증 |
| **TDD Implementer** | tdd-implementer | 에이전트 업데이트 |

### 도구 및 리소스

| 도구 | 용도 |
|-----|------|
| **tab_schema.json** | Tab 스키마 정의 |
| **moai-project-batch-questions** Skill | 질문 표시 |
| **moai-project-documentation-generator** Skill | 문서 생성 (신규) |
| **Context7 MCP** | 최신 정보 조회 |
| **WebSearch MCP** | 경쟁사 분석 |
| **pytest** | 단위 테스트 |

---

## 마일스톤 및 체크포인트

### Milestone 1: Skill 개발 완료 (2-3일)
- ✅ moai-project-documentation-generator Skill 완성
- ✅ Brainstorm 질문 세트 작성
- ✅ AI 분석 엔진 통합
- **Go/No-Go**: Brainstorm 테스트 통과

### Milestone 2: 스키마 검증 완료 (1-2일)
- ✅ tab_schema.json v3.0.0 검증
- ✅ 설정값 매핑 완성
- ✅ Smart defaults 규칙 정의
- **Go/No-Go**: 모든 스키마 검증 통과

### Milestone 3: 에이전트 업데이트 완료 (1-2일)
- ✅ /moai:0-project 명령어 업데이트
- ✅ 4개 에이전트 컨텍스트 통합
- ✅ 명령어 테스트 통과
- **Go/No-Go**: 통합 테스트 성공

### Milestone 4: 문서 완성 (2-3일)
- ✅ CLAUDE.md 업데이트
- ✅ 프로젝트 문서 템플릿 작성
- ✅ 사용자 가이드 작성
- **Go/No-Go**: 문서 검토 통과

### Milestone 5: QA 및 배포 (1-2일)
- ✅ 기능 검증 완료
- ✅ 성능 검증 완료
- ✅ 사용자 수용 테스트 통과
- **Go/No-Go**: 모든 테스트 통과

---

## 위험 및 대응 계획

### 위험 1: Brainstorm 질문 답변 품질 저하
**위험도**: 중간
**영향**: 문서 품질 저하, 사용자 만족도 감소
**대응 계획**:
- Brainstorm 질문 반복 개선 (pilot 사용자 피드백)
- AI 프롬프트 최적화
- 질문-응답 예시 제공

### 위험 2: 문서 생성 시간 초과
**위험도**: 중간
**영향**: 사용자 기대치 미충족
**대응 계획**:
- 타임아웃 설정 (30분)
- 부분 생성 시 즉시 저장
- 비동기 백그라운드 완성

### 위험 3: 조건부 배치 렌더링 오류
**위험도**: 높음
**영향**: 사용자가 불필요한 질문 표시
**대응 계획**:
- 철저한 유닛 테스트
- 모든 조건 조합 테스트 (personal, team, hybrid)
- 사용자 피드백 조기 수집

### 위험 4: Backward Compatibility 깨짐
**위험도**: 높음
**영향**: 기존 사용자 config 손상
**대응 계획**:
- v2.1.0 config.json 마이그레이션 테스트
- 롤백 메커니즘 구현
- 백업 자동 생성

### 위험 5: 에이전트 컨텍스트 로딩 실패
**위험도**: 중간
**영향**: AI 에이전트 성능 저하
**대응 계획**:
- Fallback 컨텍스트 제공
- 에러 로깅 및 모니터링
- Graceful degradation

---

## 타임라인 및 예상 소요시간

### 전체 소요시간: 7-12일

| Phase | 예상 소요시간 | 시작일 | 종료일 |
|-------|-------------|--------|--------|
| Phase 1: Skill 개발 | 2-3일 | Day 1 | Day 3 |
| Phase 2: 스키마 검증 | 1-2일 | Day 4 | Day 5 |
| Phase 3: 에이전트/명령어 업데이트 | 1-2일 | Day 6 | Day 7 |
| Phase 4: 문서 및 CLAUDE.md 업데이트 | 2-3일 | Day 8 | Day 10 |
| Phase 5: QA 및 배포 | 1-2일 | Day 11 | Day 12 |

### 병렬 처리 가능 작업
- Phase 1과 Phase 2: 독립적 (순차 필수 아님)
- Phase 3과 Phase 4: 부분적 병렬 처리 가능
- Phase 5: 모든 작업 완료 후

**최적화된 스케줄**: 7-9일 (병렬 처리 활용)

---

## 성공 기준 및 정의

### Definition of Done (DoD)

#### Phase 1 완료
- [ ] moai-project-documentation-generator Skill 코드 완성
- [ ] Brainstorm 질문 세트 (Quick/Standard/Deep) 정의
- [ ] AI 분석 엔진 구현 및 테스트 통과
- [ ] 3개 문서 생성 예시 확인
- [ ] Code review 통과

#### Phase 2 완료
- [ ] tab_schema.json v3.0.0 JSON 검증 통과
- [ ] 31개 설정값 매핑 완료
- [ ] 16개 smart defaults 규칙 정의
- [ ] 5개 auto-detect 로직 정의
- [ ] 모든 설정값 매핑 문서 검증

#### Phase 3 완료
- [ ] /moai:0-project 명령어 구현 완료
- [ ] 4개 에이전트 (project-manager, tdd-implementer, backend-expert, frontend-expert) 업데이트 완료
- [ ] 모든 에이전트 통합 테스트 통과
- [ ] 명령어 구동 테스트 성공

#### Phase 4 완료
- [ ] CLAUDE.md 업데이트 완료
- [ ] 프로젝트 문서 템플릿 (product.md, structure.md, tech.md) 작성
- [ ] 사용자 가이드 문서 작성
- [ ] 모든 문서 검토 통과

#### Phase 5 완료
- [ ] 기능 검증: 9/9 테스트 케이스 통과
- [ ] 성능 검증: 모든 시간 목표 달성
- [ ] 호환성 검증: v2.1.0 마이그레이션 성공
- [ ] 사용자 수용 테스트: 만족도 4/5 이상
- [ ] 모든 버그 fix 및 개선 사항 적용

---

## 배포 계획

### Pre-Deployment
- [ ] 모든 Phase 완료 및 QA 통과
- [ ] 롤백 계획 수립
- [ ] 릴리스 노트 작성
- [ ] 사용자 알림 메시지 준비

### Deployment
- [ ] GitHub에 새 Branch 푸시
- [ ] Release notes 게시
- [ ] CLAUDE.md 및 documentation 업데이트
- [ ] 사용자에게 알림

### Post-Deployment
- [ ] 초기 피드백 수집 (3-7일)
- [ ] 버그 패치 준비
- [ ] 사용자 만족도 모니터링
- [ ] 마이너 개선 사항 적용

---

## 결론

SPEC-REDESIGN-001의 구현으로 /moai:0-project 명령어의 사용자 경험이 획기적으로 개선되어:

1. **설정 시간 단축**: 15-20분 → 2-30분 (선택에 따라)
2. **질문 감소**: 27개 → 10개 (63% 감소)
3. **프로젝트 문서 자동 생성**: 신규 기능
4. **AI 에이전트 컨텍스트 향상**: 3개 에이전트에 자동 로딩
5. **100% 설정 커버리지**: 31개 설정값 완전 관리

이를 통해 신규 사용자의 진입 장벽을 낮추고, 기존 사용자는 포괄적인 프로젝트 문서를 통해 더 나은 개발 경험을 제공할 수 있습니다.

