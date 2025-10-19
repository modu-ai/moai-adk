---
id: UPDATE-004
version: 0.0.1
status: draft
created: 2025-10-19
updated: 2025-10-19
author: @Goos
priority: high
category: refactor
labels:
  - agents
  - skills
  - architecture
  - refactoring
related_specs:
  - UPDATE-001
  - UPDATE-002
  - UPDATE-003
scope:
  packages:
    - .claude/agents/alfred
    - .claude/skills/moai-alfred-*
  files:
    - tag-agent.md
    - trust-checker.md
    - moai-alfred-tag-scanning/skill.md
    - moai-alfred-trust-validation/skill.md
---

# @SPEC:UPDATE-004: Sub-agents를 Skills로 통합

## HISTORY
### v0.0.1 (2025-10-19)
- **INITIAL**: Sub-agents와 Skills의 명확한 역할 분리 및 중복 프롬프트 제거 명세 작성
- **AUTHOR**: @Goos
- **REASON**: Agent 프롬프트 비대화 방지, 유지보수성 향상, DRY 원칙 준수

---

## 1. 개요

### 1.1 배경

현재 MoAI-ADK의 Sub-agents와 Skills는 역할이 중복되는 부분이 있습니다:
- **tag-agent**: TAG 스캔 로직과 가이드가 Agent 프롬프트에 포함
- **trust-checker**: TRUST 검증 로직과 설명이 Agent 프롬프트에 포함
- **spec-builder**: EARS 작성법이 Agent 프롬프트에 포함

이로 인해 발생하는 문제:
1. **프롬프트 비대화**: 가이드성 내용으로 인해 Agent 프롬프트가 500+ LOC로 증가
2. **중복 관리**: 같은 내용을 여러 파일에서 반복 (CLAUDE.md, Agent 프롬프트, Skills)
3. **유지보수성 저하**: 한 곳을 수정하면 여러 곳을 동시에 수정해야 함
4. **역할 혼란**: Agent가 "어떻게"와 "왜"를 모두 설명

### 1.2 목표

Sub-agents와 Skills의 명확한 역할 분리를 통해:
- **Agents**: "어떻게 실행하는가" (실행 로직만)
- **Skills**: "왜 필요한가, 어떻게 사용하는가" (가이드, 컨텍스트, 사용법)

### 1.3 기대 효과

- **30% 이상 LOC 감소**: Agent 프롬프트에서 가이드성 내용 제거
- **단일 진실 공급원 (SSOT)**: Skills가 각 도메인의 유일한 지식 저장소
- **JIT 참조 개선**: Alfred가 필요 시 Skills만 참조하도록 최적화
- **유지보수성 향상**: 한 곳만 수정하면 모든 에이전트가 동기화

---

## 2. 요구사항 (EARS)

### 2.1 Ubiquitous Requirements (기본 요구사항)
- 시스템은 Sub-agents와 Skills의 명확한 역할 분리를 제공해야 한다
- 시스템은 중복된 프롬프트 제거를 통한 유지보수성 향상을 제공해야 한다
- 시스템은 Agent 프롬프트에 실행 로직만 포함하도록 강제해야 한다
- 시스템은 Skills에 가이드성 내용(설명, 사용법, 예제)만 포함하도록 강제해야 한다

### 2.2 Event-driven Requirements (이벤트 기반)
- WHEN 사용자가 TAG 스캔을 요청하면, Alfred는 moai-alfred-tag-scanning Skill을 자동 참조해야 한다
- WHEN 사용자가 TRUST 검증을 요청하면, Alfred는 moai-alfred-trust-validation Skill을 자동 참조해야 한다
- WHEN Phase 1 마이그레이션이 완료되면, tag-agent와 trust-checker Agent 파일은 제거되어야 한다
- WHEN Agent가 Skills를 참조하면, JIT (Just-in-Time) 방식으로 필요한 시점에만 로드되어야 한다

### 2.3 State-driven Requirements (상태 기반)
- WHILE 마이그레이션 진행 중일 때, 시스템은 기존 호출 방식과의 호환성을 유지해야 한다
- WHILE Agent 프롬프트가 500 LOC를 초과할 때, 시스템은 자동으로 Skill 분리를 권장해야 한다
- WHILE Skills가 업데이트될 때, 관련된 모든 Agent는 자동으로 최신 정보를 참조해야 한다

### 2.4 Optional Features (선택적 기능)
- WHERE 기존 Agent 호출 방식이 존재하면, 시스템은 Skill 기반 호출로 자동 리다이렉트할 수 있다
- WHERE Agent 프롬프트에 가이드성 내용이 발견되면, 시스템은 자동으로 Skill로 분리할 수 있다

### 2.5 Constraints (제약사항)
- IF Agent 프롬프트에 가이드성 내용이 포함되면, 시스템은 빌드 시점에 경고해야 한다
- IF Skill이 없는 상태에서 Agent가 호출되면, 시스템은 명확한 에러 메시지를 반환해야 한다
- 전체 LOC는 마이그레이션 후 30% 이상 감소해야 한다
- Agent 프롬프트는 300 LOC를 초과할 수 없다

---

## 3. 아키텍처 설계

### 3.1 역할 분리 원칙

```
┌─────────────────────────────────────────────────────────────┐
│                          Alfred                              │
│                  (Central Orchestrator)                      │
└────────┬─────────────────────────────────────────┬──────────┘
         │                                          │
         │ "TAG 스캔 실행"                         │ "TRUST 검증"
         │                                          │
         ↓                                          ↓
┌─────────────────┐                        ┌──────────────────┐
│   tag-agent     │                        │  trust-checker   │
│  (실행 로직)    │                        │   (실행 로직)    │
└────────┬────────┘                        └────────┬─────────┘
         │                                           │
         │ JIT 참조                                 │ JIT 참조
         │                                           │
         ↓                                           ↓
┌───────────────────────────┐          ┌─────────────────────────┐
│ moai-alfred-tag-scanning  │          │ moai-alfred-trust-      │
│        (Skill)             │          │  validation (Skill)     │
│  - 설명                    │          │  - 설명                 │
│  - 사용법                  │          │  - 사용법               │
│  - 예제                    │          │  - 예제                 │
│  - TAG 규칙                │          │  - TRUST 5원칙          │
└───────────────────────────┘          └─────────────────────────┘
```

**핵심 원칙**:
1. **Agent**: 실행 로직만 (≤300 LOC)
2. **Skill**: 가이드 + 컨텍스트 (제한 없음, 필요한 만큼)
3. **Alfred**: JIT 방식으로 필요한 Skill만 로드

### 3.2 마이그레이션 전략

#### Phase 1: tag-agent, trust-checker 완전 통합
1. **tag-agent 마이그레이션**:
   - 실행 로직: `tag-agent.md`에 남김 (≤200 LOC)
   - 가이드: `moai-alfred-tag-scanning/skill.md`로 이동
   - TAG 규칙, 예제, 검증 방법 → Skill로 완전 이전

2. **trust-checker 마이그레이션**:
   - 실행 로직: `trust-checker.md`에 남김 (≤200 LOC)
   - 가이드: `moai-alfred-trust-validation/skill.md`로 이동
   - TRUST 5원칙, 검증 체크리스트 → Skill로 완전 이전

3. **Agent 파일 제거**:
   - 마이그레이션 완료 후 `tag-agent.md`, `trust-checker.md` 삭제
   - Alfred가 직접 Skills를 참조하도록 변경

#### Phase 2: spec-builder EARS 부분 분리
1. **EARS 가이드 분리**:
   - EARS 작성법 → `moai-alfred-ears-authoring/skill.md`로 이동
   - spec-builder는 EARS Skill 참조
   - 실행 로직만 `spec-builder.md`에 남김

2. **SPEC 메타데이터 분리**:
   - 이미 `spec-metadata.md`로 분리됨 (완료)
   - spec-builder가 JIT 참조

#### Phase 3: 호환성 테스트 및 검증
1. **기능 검증**:
   - TAG 스캔 기능이 Skills 참조 방식으로 정상 동작하는지 확인
   - TRUST 검증 기능이 Skills 참조 방식으로 정상 동작하는지 확인

2. **성능 측정**:
   - 전체 LOC 감소율 확인 (목표: ≥30%)
   - JIT 로딩 시간 측정

3. **문서 업데이트**:
   - CLAUDE.md 업데이트 (Agent 목록 변경)
   - development-guide.md 업데이트 (Skills 참조 가이드 추가)

---

## 4. 구현 세부사항

### 4.1 Phase 1 상세 계획

#### 4.1.1 tag-agent 마이그레이션

**Before (tag-agent.md - 500+ LOC)**:
```markdown
---
name: tag-agent
description: TAG 시스템 관리 전문가
model: haiku
---

# TAG 시스템 관리

## TAG 체계
@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID

## TAG 규칙
- TAG ID는 영구 불변
- 3자리 숫자 사용
... (400+ LOC의 가이드)
```

**After (tag-agent.md - 200 LOC)**:
```markdown
---
name: tag-agent
description: TAG 시스템 관리 전문가
model: haiku
skills:
  - moai-alfred-tag-scanning
---

# TAG 시스템 관리

## 실행 로직
- TAG 스캔: rg '@(SPEC|TEST|CODE|DOC):' -n
- 고아 TAG 탐지: 의존성 체인 검증
- TAG 무결성 검증

상세 가이드: @moai-alfred-tag-scanning Skill 참조
```

**Skills (moai-alfred-tag-scanning/skill.md - 300+ LOC)**:
```markdown
---
name: moai-alfred-tag-scanning
description: TAG 시스템 스캔 및 검증 가이드
category: guidance
---

# TAG 시스템 스캔 가이드

## TAG 체계
@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID

## TAG 규칙
... (전체 가이드)
```

#### 4.1.2 trust-checker 마이그레이션

**Before (trust-checker.md - 600+ LOC)**:
```markdown
---
name: trust-checker
description: TRUST 5원칙 검증 전문가
model: haiku
---

# TRUST 5원칙

## T - Test First
...
## R - Readable
...
... (500+ LOC의 가이드)
```

**After (trust-checker.md - 200 LOC)**:
```markdown
---
name: trust-checker
description: TRUST 5원칙 검증 전문가
model: haiku
skills:
  - moai-alfred-trust-validation
---

# TRUST 5원칙 검증

## 실행 로직
- 테스트 커버리지 확인
- 린터 실행
- 타입 검증
- 보안 스캔
- TAG 추적성 확인

상세 가이드: @moai-alfred-trust-validation Skill 참조
```

**Skills (moai-alfred-trust-validation/skill.md - 400+ LOC)**:
```markdown
---
name: moai-alfred-trust-validation
description: TRUST 5원칙 검증 가이드
category: guidance
---

# TRUST 5원칙 가이드

## T - Test First
...
## R - Readable
...
... (전체 가이드)
```

### 4.2 Phase 2 상세 계획

#### 4.2.1 spec-builder EARS 부분 분리

**Before (spec-builder.md - 800+ LOC)**:
```markdown
# EARS 작성법
## Ubiquitous Requirements
...
## Event-driven Requirements
...
... (300+ LOC의 EARS 가이드)
```

**After (spec-builder.md - 500 LOC)**:
```markdown
---
skills:
  - moai-alfred-ears-authoring
---

# SPEC 작성 전문가

## SPEC 작성 프로세스
1. EARS 구문 적용 (상세: @moai-alfred-ears-authoring)
2. SPEC 메타데이터 작성 (상세: spec-metadata.md)
3. 검증

EARS 작성법: @moai-alfred-ears-authoring Skill 참조
```

**Skills (moai-alfred-ears-authoring/skill.md - 300+ LOC)**:
```markdown
---
name: moai-alfred-ears-authoring
description: EARS 요구사항 작성 가이드
category: guidance
---

# EARS (Easy Approach to Requirements Syntax)

## Ubiquitous Requirements
...
## Event-driven Requirements
...
... (전체 가이드)
```

### 4.3 Phase 3 상세 계획

#### 4.3.1 호환성 테스트

**테스트 시나리오**:
1. **TAG 스캔 테스트**:
   ```bash
   # 사용자 호출
   @agent-tag-agent "AUTH 도메인 TAG 목록 조회"

   # 예상 동작
   1. tag-agent 실행
   2. moai-alfred-tag-scanning Skill JIT 로드
   3. TAG 스캔 수행
   4. 결과 반환
   ```

2. **TRUST 검증 테스트**:
   ```bash
   # 사용자 호출
   @agent-trust-checker "현재 프로젝트 TRUST 원칙 준수도 확인"

   # 예상 동작
   1. trust-checker 실행
   2. moai-alfred-trust-validation Skill JIT 로드
   3. TRUST 5원칙 검증
   4. 보고서 생성
   ```

3. **SPEC 작성 테스트**:
   ```bash
   # 사용자 호출
   /alfred:1-spec "새 기능"

   # 예상 동작
   1. spec-builder 실행
   2. moai-alfred-ears-authoring Skill JIT 로드
   3. EARS 구문 적용
   4. SPEC 문서 생성
   ```

#### 4.3.2 LOC 감소율 측정

**측정 방법**:
```bash
# Before (마이그레이션 전)
wc -l .claude/agents/alfred/tag-agent.md
wc -l .claude/agents/alfred/trust-checker.md
wc -l .claude/agents/alfred/spec-builder.md

# After (마이그레이션 후)
wc -l .claude/agents/alfred/tag-agent.md
wc -l .claude/agents/alfred/trust-checker.md
wc -l .claude/agents/alfred/spec-builder.md
wc -l .claude/skills/moai-alfred-tag-scanning/skill.md
wc -l .claude/skills/moai-alfred-trust-validation/skill.md
wc -l .claude/skills/moai-alfred-ears-authoring/skill.md

# 감소율 계산
총 감소율 = (Before - After_Agents) / Before * 100
```

**목표**: ≥30% LOC 감소

---

## 5. 제약사항 및 고려사항

### 5.1 기술적 제약사항
- Agent 프롬프트는 300 LOC를 초과할 수 없다
- Skills는 JIT 방식으로 로드되어야 한다 (성능 영향 최소화)
- 기존 호출 방식과의 호환성을 유지해야 한다

### 5.2 보안 고려사항
- Skills에 민감한 정보가 포함되지 않도록 검증
- Agent 실행 로직에만 민감한 작업 포함

### 5.3 성능 고려사항
- JIT 로딩 시간: ≤100ms
- Skill 파일 크기: ≤500KB
- 전체 컨텍스트 비용: 마이그레이션 전과 동일 또는 감소

---

## 6. 검증 기준

### 6.1 기능 검증
- [ ] TAG 스캔 기능이 Skills 참조 방식으로 정상 동작
- [ ] TRUST 검증 기능이 Skills 참조 방식으로 정상 동작
- [ ] EARS 작성이 Skills 참조 방식으로 정상 동작
- [ ] 기존 호출 방식과의 호환성 유지

### 6.2 품질 검증
- [ ] 전체 LOC 30% 이상 감소
- [ ] Agent 프롬프트 ≤300 LOC
- [ ] JIT 로딩 시간 ≤100ms
- [ ] Skill 파일 크기 ≤500KB

### 6.3 문서 검증
- [ ] CLAUDE.md 업데이트 완료
- [ ] development-guide.md 업데이트 완료
- [ ] 각 Agent 프롬프트에 Skills 참조 명시
- [ ] 각 Skill에 사용 예제 포함

---

## 7. 롤백 계획

### 7.1 롤백 조건
- 기능 손실 발생
- 성능 저하 (JIT 로딩 시간 >200ms)
- 호환성 문제 발생

### 7.2 롤백 절차
1. 이전 Agent 프롬프트 파일 복원
2. Skills 참조 제거
3. CLAUDE.md 이전 버전으로 복원
4. 기능 테스트 재실행

---

## 8. 다음 단계

### 8.1 구현 순서
1. Phase 1: tag-agent, trust-checker 마이그레이션 (우선순위: high)
2. Phase 2: spec-builder EARS 부분 분리 (우선순위: medium)
3. Phase 3: 호환성 테스트 및 검증 (우선순위: high)

### 8.2 예상 작업량
- Phase 1: 2-3시간
- Phase 2: 1-2시간
- Phase 3: 1시간

### 8.3 리스크 관리
- **높은 리스크**: 기존 호출 방식 호환성 → 철저한 테스트 필요
- **중간 리스크**: JIT 로딩 성능 → 벤치마크 테스트 필요
- **낮은 리스크**: 문서 업데이트 → 자동화 스크립트 활용

---

## 9. 참고 문서

- `.moai/memory/spec-metadata.md` - SPEC 메타데이터 표준
- `.moai/memory/development-guide.md` - EARS 작성법, TRUST 5원칙
- `CLAUDE.md` - Agent 목록, Skills 참조 가이드
- `.claude/skills/moai-alfred-*/skill.md` - 각 Skill 상세 가이드

---

## 10. 메타데이터

**TAG**: @SPEC:UPDATE-004
**관련 TAG**:
- @SPEC:UPDATE-001 (Skills Flat 구조 변경)
- @SPEC:UPDATE-002 (Skills name 필드 prefix 제거)
- @SPEC:UPDATE-003 (moai-alfred-ears-authoring Skill 추가)

**의존성**:
- UPDATE-001 완료 (Skills Flat 구조)
- UPDATE-003 완료 (EARS Skill 생성)

**차단**:
- 없음 (독립적으로 진행 가능)

---

**작성자**: @Goos
**최초 작성일**: 2025-10-19
**최종 수정일**: 2025-10-19
