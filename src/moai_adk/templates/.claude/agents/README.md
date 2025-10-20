# 🤖 MoAI-ADK Agents Architecture

> **Alfred SuperAgent와 9개 전문 에이전트의 조율 원칙**

---

## 🎭 Agent 체계

### Core Architecture (3개 레벨)

```
            사용자 요청
                 ↓
        ┌────────────────────┐
        │  Alfred 분석       │ ← 요청 분석 + 라우팅
        └────────────────────┘
                 ↓
    ┌────────────────────────────────┐
    │ 작업 라우팅 전략 결정          │
    ├────────────────────────────────┤
    │ 1. 직접 처리 (간단한 조회)     │
    │ 2. Single Agent (단일 전문가)  │
    │ 3. Sequential (순차 실행)      │
    │ 4. Parallel (병렬 실행)        │
    └────────────────────────────────┘
                 ↓
    ┌────────────────────────────────┐
    │      9개 전문 에이전트         │
    └────────────────────────────────┘
                 ↓
        ┌────────────────────┐
        │  품질 게이트 검증  │
        │ - TRUST 5원칙      │
        │ - @TAG 무결성      │
        └────────────────────┘
                 ↓
        ┌────────────────────┐
        │ Alfred 결과 통합   │
        └────────────────────┘
```

---

## 👥 9개 전문 에이전트

### Tier 1: Core Agents (7개) ⭐

#### 1. **spec-builder** 🏗️ (Sonnet)
**역할**: SPEC 작성, EARS 명세 설계
**전문 영역**: 요구사항 분석, 아키텍처 설계, 복잡한 의사결정
**호출 방법**: `/alfred:1-plan` (스펙 작성 단계)
**사용 Skills**:
- moai-foundation-specs (메타데이터 검증)
- moai-foundation-ears (요구사항 작성)
- moai-foundation-langs (언어 감지)

**사용 시나리오**:
```markdown
사용자: "/alfred:1-plan 사용자 인증 시스템 개발"
→ spec-builder 호출
→ Phase 1: 프로젝트 분석, SPEC 후보 제안
→ 사용자 승인
→ Phase 2: SPEC 문서 작성 (.moai/specs/SPEC-AUTH-001/)
→ Git 브랜치 생성 (feature/SPEC-AUTH-001)
→ Draft PR 생성
```

---

#### 2. **code-builder** (aliased: tdd-implementer) 💎 (Sonnet)
**역할**: TDD 구현, Red-Green-Refactor 주도
**전문 영역**: SPEC 분석, 구현 전략, 코드 품질
**호출 방법**: `/alfred:2-run SPEC-ID`
**사용 Skills**:
- moai-lang-python (Python TDD)
- moai-lang-typescript (TypeScript TDD)
- moai-lang-go (Go TDD)
- 18개 추가 언어 스킬

**사용 시나리오**:
```markdown
사용자: "/alfred:2-run AUTH-001"
→ code-builder 호출
→ Phase 1: SPEC 분석, TDD 계획 수립
→ 사용자 승인
→ Phase 2: TDD 구현
  ├─ 🔴 RED: 테스트 작성 및 실패 확인
  ├─ 🟢 GREEN: 최소 구현 및 테스트 통과
  └─ ♻️ REFACTOR: 코드 품질 개선
→ Git 커밋 (RED/GREEN/REFACTOR 단계별)
```

---

#### 3. **doc-syncer** 📖 (Haiku)
**역할**: 문서 동기화, Living Document 관리
**전문 영역**: @TAG 체인 동기화, PR 업데이트, Living Document 생성
**호출 방법**: `/alfred:3-sync`
**사용 Skills**:
- moai-foundation-tags (@TAG 스캔)
- moai-foundation-trust (TRUST 검증)
- moai-foundation-specs (메타데이터 확인)

**사용 시나리오**:
```markdown
사용자: "/alfred:3-sync"
→ doc-syncer 호출
→ Phase 1: 동기화 범위 분석
→ 사용자 승인
→ Phase 2: 문서 동기화
  ├─ TAG 체인 검증 (@SPEC → @TEST → @CODE → @DOC)
  ├─ Living Document 자동 생성 (docs/)
  ├─ PR 설명 업데이트
  └─ PR 상태 전환 (Draft → Ready)
→ CI/CD 확인 후 자동 머지 (선택)
```

---

#### 4. **tag-agent** 🏷️ (Haiku)
**역할**: @TAG 시스템 관리, 추적성 보장
**전문 영역**: TAG 스캔, 중복 방지, 고아 TAG 탐지
**호출 방법**: `@agent-tag-agent` (온디맨드)
**사용 명령어**:
```bash
@agent-tag-agent "AUTH 도메인 TAG 목록"
@agent-tag-agent "고아 TAG 찾아줘"
@agent-tag-agent "TAG 체인 검증"
```

**사용 시나리오**:
```markdown
사용자: "@agent-tag-agent 전체 TAG 무결성 검증"
→ tag-agent 호출
→ CODE-FIRST 스캔 실행: rg '@(SPEC|TEST|CODE|DOC):' -n
→ 결과:
  ✅ @SPEC:AUTH-001 (PASS)
  ✅ @TEST:AUTH-001 (PASS)
  ✅ @CODE:AUTH-001 (PASS)
  ⚠️ @DOC:AUTH-001 (MISSING)
→ 권장사항 제시 (동기화 필요)
```

---

#### 5. **git-manager** 🚀 (Haiku)
**역할**: Git 워크플로우 자동화, 릴리즈 관리
**전문 영역**: 브랜치 생성, PR 관리, 커밋 자동화, 배포
**호출 방법**: `@agent-git-manager` (온디맨드)
**사용 명령어**:
```bash
@agent-git-manager "SPEC-AUTH-001 브랜치 생성"
@agent-git-manager "PR-123 상태 확인"
@agent-git-manager "develop → main 머지 준비"
```

**사용 시나리오**:
```markdown
사용자: "develop을 main으로 머지하고 릴리즈해줘"
→ git-manager 호출
→ Phase 1: 머지 준비 확인
  ├─ CI/CD 통과 여부
  ├─ 브랜치 충돌 확인
  └─ 버전 업데이트 필요 여부
→ 사용자 승인
→ Phase 2: 머지 및 릴리즈
  ├─ Git 커밋 (Squash merge)
  ├─ 버전 태그 생성
  └─ GitHub Release 생성 (Draft)
```

---

#### 6. **debug-helper** 🔬 (Sonnet)
**역할**: 오류 진단, 문제 해결
**전문 영역**: 스택 트레이스 분석, 근본 원인 파악, 해결 방법 도출
**호출 방법**: `@agent-debug-helper` (온디맨드 또는 자동 호출)
**사용 명령어**:
```bash
@agent-debug-helper "TypeError: 'NoneType' object has no attribute 'name'"
@agent-debug-helper "SPEC과의 불일치 확인"
```

**사용 시나리오**:
```markdown
사용자: "JWT 만료 오류 해결해줘"
→ debug-helper 호출
→ 스택 트레이스 분석
→ SPEC 문서 확인 (AUTH-001)
→ 근본 원인: SPEC 요구사항 vs 구현 불일치
→ 3단계 해결방안 제시
  ├─ Level 1: 즉시 코드 수정
  ├─ Level 2: SPEC 업데이트 필요
  └─ Level 3: 아키텍처 개선 제안
→ 권장사항 제시
```

---

#### 7. **trust-checker** ✅ (Haiku)
**역할**: TRUST 5원칙 검증, 품질 보증
**전문 영역**: 테스트 커버리지, 코드 가독성, 복잡도, 보안, 추적성 검증
**호출 방법**: `@agent-trust-checker` (온디맨드)
**사용 명령어**:
```bash
@agent-trust-checker "TRUST 5원칙 검증"
@agent-trust-checker "AUTH-001 코드 품질 확인"
```

**사용 시나리오**:
```markdown
사용자: "@agent-trust-checker 전체 프로젝트 품질 검증"
→ trust-checker 호출
→ TRUST 검증 실행
  ├─ T (Test): 커버리지 87% (목표 85%) ✓
  ├─ R (Readable): 45/45 파일 준수 ✓
  ├─ U (Unified): SPEC 기반 아키텍처 ✓
  ├─ S (Secured): 입력 검증, SQL Injection 방어 ✓
  └─ T (Trackable): @TAG 체인 99% 완성도 ✓
→ 최종 평가: 🟢 양호 (90/100)
```

---

### Tier 2: Support Agents (10개) 🔧

#### 8. **cc-manager** 🛠️ (Sonnet)
**역할**: Claude Code 설정 최적화
**전문 영역**: .claude/ 디렉토리 관리, settings.json 최적화, Hooks 설정
**호출 방법**: `@agent-cc-manager` (온디맨드)

#### 9. **project-manager** 📋 (Sonnet)
**역할**: 프로젝트 초기화, 전체 조율
**전문 영역**: .moai/ 구조 생성, 언어 감지, 템플릿 최적화
**호출 방법**: `/alfred:0-project`

---

## 🔄 Agent 호출 전략

### Sequential Pattern
```
요청 → Agent A → 결과 → Agent B → 결과 → 최종 보고
```
**사용 시나리오**: `/alfred:1-plan` → `/alfred:2-run` → `/alfred:3-sync`

### Parallel Pattern
```
요청 → ┬─ Agent A → 결과
       ├─ Agent B → 결과
       └─ Agent C → 결과
       ↓
       통합 보고
```
**사용 시나리오**: 테스트 실행 + 린트 + 타입 체크 동시 실행

### Conditional Pattern
```
요청 → 조건 확인
       ├─ IF SPEC 없음 → spec-builder
       ├─ ELIF 테스트 실패 → debug-helper
       └─ ELSE → code-builder
```

---

## 📊 Agent 모델 선택 가이드

### Sonnet 4.5 (복잡한 판단)
```
사용: spec-builder, code-builder, debug-helper, cc-manager, project-manager
이유: 복잡한 분석, 창의적 설계, 다단계 문제 해결
비용: 높음 → 중요한 작업에만 사용
```

### Haiku 4.5 (반복 작업)
```
사용: doc-syncer, tag-agent, git-manager, trust-checker
이유: 정형화된 작업, 빠른 응답, 높은 처리량
비용: 낮음 (67% 절감) → 많은 작업에 사용
```

**최적 조합**:
- Sonnet: 7개 (39%)
- Haiku: 10개 (61%)
- **평균 비용**: Haiku 사용으로 40% 절감 + 속도 3배

---

## 🎯 각 Skill별 사용 Agent

### Foundation Skills
| Skill | 사용 Agent | 호출 시점 |
|-------|----------|---------|
| moai-foundation-specs | spec-builder, doc-syncer | SPEC 작성/검증 시 |
| moai-foundation-ears | spec-builder | SPEC 작성 시 |
| moai-foundation-tags | doc-syncer, tag-agent | TAG 체인 검증 시 |
| moai-foundation-trust | trust-checker, doc-syncer | 품질 검증 시 |
| moai-foundation-langs | language-detector, project-manager | 프로젝트 초기화 시 |
| moai-foundation-git | git-manager | Git 작업 시 |

### Language Skills
| Skill | 사용 Agent | 호출 시점 |
|-------|----------|---------|
| moai-lang-python | code-builder | Python SPEC 구현 시 |
| moai-lang-typescript | code-builder | TypeScript SPEC 구현 시 |
| moai-lang-go | code-builder | Go SPEC 구현 시 |
| (18개 추가) | code-builder | 언어별 TDD 구현 시 |

### Domain Skills
| Skill | 사용 Agent | 호출 시점 |
|-------|----------|---------|
| moai-domain-backend | code-builder | 백엔드 구현 시 |
| moai-domain-frontend | code-builder | 프론트엔드 구현 시 |
| moai-domain-ml | code-builder | ML/AI 구현 시 |
| (7개 추가) | code-builder | 도메인별 구현 시 |

---

## 📝 Agent 협업 원칙

1. **커맨드 우선순위**: 커맨드 지침 > Agent 지침 > Skill 지침
2. **단일 책임 원칙**: 각 Agent는 자신의 영역만 담당
3. **중앙 조율**: Alfred만이 Agent 간 작업 조율
4. **품질 게이트**: 각 단계 완료 시 검증 (TRUST + TAG)
5. **에러 처리**: 예외 발생 시 debug-helper 자동 호출

---

## 🚀 다음 단계

### Immediate (지금)
- [x] Agent 아키텍처 문서화
- [x] Skill별 사용 Agent 매핑
- [ ] Agent 호출 시나리오 집합

### Short-term (1주)
- [ ] Agent 성능 모니터링 (로깅)
- [ ] Agent 간 의존성 그래프 생성
- [ ] Agent 실패 처리 정책 수립

### Long-term (1달)
- [ ] Agent 활용도 분석
- [ ] 미사용 Agent 식별
- [ ] Agent 최적화 및 통합

---

**관련 문서**:
- [SKILL_INTEGRATION_TEST_REPORT.md](../../SKILL_INTEGRATION_TEST_REPORT.md)
- [.claude/skills/README.md](./../skills/README.md)
- [.moai/memory/development-guide.md](../../.moai/memory/development-guide.md)

**작성자**: @agent-cc-manager
**최종 업데이트**: 2025-10-20
