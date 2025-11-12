
> **Sub-agents를 Skills로 통합**

---

## 1. 구현 전략

### 1.1 핵심 접근 방법

**DRY (Don't Repeat Yourself) 원칙 적용**:
- Agent 프롬프트: 실행 로직만 (≤300 LOC)
- Skills: 가이드 + 컨텍스트 (제한 없음)
- 단일 진실 공급원 (SSOT): Skills가 각 도메인의 유일한 지식 저장소

**JIT (Just-in-Time) 참조**:
- Alfred가 필요한 시점에만 Skills 로드
- 초기 컨텍스트 부담 최소화
- 성능 영향 최소화 (목표: ≤100ms)

### 1.2 단계별 우선순위

```
Phase 1 (High Priority)
├─ tag-agent 마이그레이션
├─ trust-checker 마이그레이션
└─ Agent 파일 제거

Phase 2 (Medium Priority)
├─ spec-builder EARS 부분 분리
└─ SPEC 메타데이터 참조 최적화

Phase 3 (High Priority)
├─ 기능 검증
├─ 성능 측정
└─ 문서 업데이트
```

---

## 2. Phase 1: tag-agent, trust-checker 완전 통합

### 2.1 tag-agent 마이그레이션

#### 작업 1: 실행 로직과 가이드 분리

**현재 상태 분석**:
```bash
# tag-agent.md LOC 측정
wc -l .claude/agents/alfred/tag-agent.md

# 가이드 부분 식별
rg "## TAG 체계|## TAG 규칙|## 예제" .claude/agents/alfred/tag-agent.md
```

**실행 로직 (tag-agent.md에 남김)**:
- TAG 스캔: `rg '@(SPEC|TEST|CODE|DOC):' -n`
- 고아 TAG 탐지: 의존성 체인 검증
- TAG 무결성 검증: SPEC ↔ TEST ↔ CODE ↔ DOC 연결 확인

**가이드 (moai-alfred-tag-scanning/skill.md로 이동)**:
- TAG 체계 설명
- TAG 규칙 상세
- TAG 검증 방법
- 예제 및 사용법
- 트러블슈팅 가이드

#### 작업 2: Skills 파일 생성

**파일 구조**:
```
.claude/skills/moai-alfred-tag-scanning/
├── skill.md          # TAG 스캔 가이드
└── README.md         # Skill 개요 (선택)
```

**skill.md 구조**:
```markdown
---
name: moai-alfred-tag-scanning
description: TAG 시스템 스캔 및 검증 가이드
category: guidance
tags:
  - tag
  - traceability
  - validation
---

# TAG 시스템 스캔 가이드

## Overview
TAG 시스템의 전체 구조, 규칙, 검증 방법을 설명합니다.

## TAG 체계
...

## TAG 규칙
...

## 검증 방법
...

## 예제
...
```

#### 작업 3: tag-agent.md 업데이트

**업데이트 내용**:
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

### TAG 스캔
\`\`\`bash
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
\`\`\`

### 고아 TAG 탐지
\`\`\`python
def detect_orphan_tags():
    spec_tags = scan_tags("SPEC")
    test_tags = scan_tags("TEST")
    code_tags = scan_tags("CODE")

    orphans = []
    for tag in code_tags:
        if tag not in spec_tags:
            orphans.append(tag)

    return orphans
\`\`\`

### TAG 무결성 검증
\`\`\`bash
# SPEC → TEST 연결 확인
    if [ -z "$test_exists" ]; then
        echo "Missing test for $spec"
    fi
done
\`\`\`

## 상세 가이드
TAG 규칙, 체계, 검증 방법: @moai-alfred-tag-scanning Skill 참조
```

**목표 LOC**: ≤200 LOC (현재 500+ LOC에서 60% 감소)

### 2.2 trust-checker 마이그레이션

#### 작업 1: 실행 로직과 가이드 분리

**실행 로직 (trust-checker.md에 남김)**:
- 테스트 커버리지 확인: `pytest --cov`
- 린터 실행: `ruff check`, `mypy`
- 타입 검증: 언어별 타입 체커 실행
- 보안 스캔: `bandit`, `safety`
- TAG 추적성 확인: TAG 체인 검증

**가이드 (moai-alfred-trust-validation/skill.md로 이동)**:
- TRUST 5원칙 상세 설명
- 언어별 도구 매핑
- 검증 체크리스트
- 예제 및 사용법
- 트러블슈팅 가이드

#### 작업 2: Skills 파일 생성

**파일 구조**:
```
.claude/skills/moai-alfred-trust-validation/
├── skill.md          # TRUST 검증 가이드
└── README.md         # Skill 개요 (선택)
```

**skill.md 구조**:
```markdown
---
name: moai-alfred-trust-validation
description: TRUST 5원칙 검증 가이드
category: guidance
tags:
  - trust
  - quality
  - validation
---

# TRUST 5원칙 가이드

## T - Test First
...

## R - Readable
...

## U - Unified
...

## S - Secured
...

## T - Trackable
...

## 검증 체크리스트
...

## 언어별 도구 매핑
...
```

#### 작업 3: trust-checker.md 업데이트

**업데이트 내용**:
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

### T - Test First
\`\`\`bash
pytest --cov --cov-report=term
\`\`\`

### R - Readable
\`\`\`bash
ruff check .
mypy .
\`\`\`

### U - Unified
\`\`\`bash
# 타입 안전성 검증
mypy --strict .
\`\`\`

### S - Secured
\`\`\`bash
bandit -r src/
safety check
\`\`\`

### T - Trackable
\`\`\`bash
rg '@(SPEC|TEST|CODE|DOC):' -n
\`\`\`

## 상세 가이드
TRUST 5원칙, 검증 방법: @moai-alfred-trust-validation Skill 참조
```

**목표 LOC**: ≤200 LOC (현재 600+ LOC에서 66% 감소)

### 2.3 Agent 파일 제거 (마이그레이션 완료 후)

**제거 대상**:
- `.claude/agents/alfred/tag-agent.md` → Alfred가 직접 Skills 참조
- `.claude/agents/alfred/trust-checker.md` → Alfred가 직접 Skills 참조

**Alfred 업데이트**:
```markdown
# Alfred에서 직접 Skills 참조
WHEN 사용자가 TAG 스캔을 요청하면:
1. @moai-alfred-tag-scanning Skill JIT 로드
2. TAG 스캔 실행
3. 결과 반환

WHEN 사용자가 TRUST 검증을 요청하면:
1. @moai-alfred-trust-validation Skill JIT 로드
2. TRUST 5원칙 검증
3. 보고서 생성
```

---

## 3. Phase 2: spec-builder EARS 부분 분리

### 3.1 EARS 가이드 분리

#### 작업 1: 실행 로직과 가이드 분리

**실행 로직 (spec-builder.md에 남김)**:
- SPEC 파일 생성 (spec.md, plan.md, acceptance.md)
- YAML Front Matter 생성
- EARS 구문 적용 (상세는 Skill 참조)
- HISTORY 섹션 생성

**가이드 (moai-alfred-ears-authoring/skill.md - 이미 생성됨)**:
- EARS 5가지 구문 설명
- 작성 예제
- 베스트 프랙티스
- 안티패턴

#### 작업 2: spec-builder.md 업데이트

**업데이트 내용**:
```markdown
---
name: spec-builder
description: SPEC 작성 전문가
model: sonnet
skills:
  - moai-alfred-ears-authoring
---

# SPEC 작성 전문가

## SPEC 작성 프로세스

### 1. SPEC 메타데이터 작성
\`\`\`yaml
---
id: {ID}
version: 0.0.1
status: draft
created: {YYYY-MM-DD}
updated: {YYYY-MM-DD}
author: @{GitHub ID}
priority: {low|medium|high|critical}
---
\`\`\`

### 2. EARS 구문 적용
상세 가이드: @moai-alfred-ears-authoring Skill 참조

### 3. HISTORY 섹션 작성
\`\`\`markdown
## HISTORY
### v0.0.1 ({YYYY-MM-DD})
- **INITIAL**: {설명}
- **AUTHOR**: @{GitHub ID}
\`\`\`

## 상세 가이드
- EARS 작성법: @moai-alfred-ears-authoring
- SPEC 메타데이터: spec-metadata.md
```

**목표 LOC**: ≤500 LOC (현재 800+ LOC에서 37% 감소)

---

## 4. Phase 3: 호환성 테스트 및 검증

### 4.1 기능 검증

#### 테스트 시나리오 1: TAG 스캔
```bash
# 사용자 호출
@agent-tag-agent "AUTH 도메인 TAG 목록 조회"

# 예상 동작
1. Alfred가 @moai-alfred-tag-scanning Skill JIT 로드
2. TAG 스캔 실행
3. 결과 반환

# 검증 방법
- TAG 목록이 정확하게 반환되는지 확인
- JIT 로딩 시간 측정 (≤100ms)
- 기존 호출 방식과 결과 동일성 확인
```

#### 테스트 시나리오 2: TRUST 검증
```bash
# 사용자 호출
@agent-trust-checker "현재 프로젝트 TRUST 원칙 준수도 확인"

# 예상 동작
1. Alfred가 @moai-alfred-trust-validation Skill JIT 로드
2. TRUST 5원칙 검증
3. 보고서 생성

# 검증 방법
- TRUST 5원칙 모두 검증되는지 확인
- 보고서 형식이 일관되는지 확인
- 기존 호출 방식과 결과 동일성 확인
```

#### 테스트 시나리오 3: SPEC 작성
```bash
# 사용자 호출
/alfred:1-plan "새 기능"

# 예상 동작
1. spec-builder가 @moai-alfred-ears-authoring Skill JIT 로드
2. EARS 구문 적용
3. SPEC 문서 생성

# 검증 방법
- EARS 5가지 구문 모두 적용되는지 확인
- SPEC 메타데이터가 올바르게 생성되는지 확인
- 기존 SPEC 품질과 동일 또는 향상되는지 확인
```

### 4.2 성능 측정

#### LOC 감소율 측정
```bash
# Before (마이그레이션 전)
echo "Before:"
wc -l .claude/agents/alfred/tag-agent.md
wc -l .claude/agents/alfred/trust-checker.md
wc -l .claude/agents/alfred/spec-builder.md
echo "Total: $(cat .claude/agents/alfred/{tag-agent,trust-checker,spec-builder}.md | wc -l)"

# After (마이그레이션 후)
echo "After (Agents only):"
wc -l .claude/agents/alfred/tag-agent.md
wc -l .claude/agents/alfred/trust-checker.md
wc -l .claude/agents/alfred/spec-builder.md
echo "Total: $(cat .claude/agents/alfred/{tag-agent,trust-checker,spec-builder}.md | wc -l)"

# 감소율 계산
echo "Reduction: $((100 - (After * 100 / Before)))%"
```

**목표**: ≥30% LOC 감소

#### JIT 로딩 시간 측정
```bash
# Skill 로딩 시간 측정
time rg "@moai-alfred-tag-scanning" .claude/skills/

# 목표: ≤100ms
```

### 4.3 문서 업데이트

#### CLAUDE.md 업데이트
```markdown
# 9개 전문 에이전트 생태계 (수정)

| 에이전트          | 페르소나           | 전문 영역      | 호출                     |
| ----------------- | ------------------ | -------------- | ------------------------ |
| ~~tag-agent~~     | ~~지식 관리자~~    | ~~TAG 시스템~~ | ~~@agent-tag-agent~~     |
| ~~trust-checker~~ | ~~품질 보증 리드~~ | ~~TRUST 검증~~ | ~~@agent-trust-checker~~ |

# Skills 참조 가이드 (추가)
Alfred는 필요 시 다음 Skills를 JIT 방식으로 로드합니다:
- moai-alfred-tag-scanning: TAG 스캔 및 검증
- moai-alfred-trust-validation: TRUST 5원칙 검증
- moai-alfred-ears-authoring: EARS 요구사항 작성
```

#### development-guide.md 업데이트
```markdown
# Skills 참조 가이드 (추가)

## JIT 참조 방법
Alfred는 필요한 시점에만 Skills를 로드합니다:
1. 사용자 요청 분석
2. 필요한 Skill 식별
3. JIT 로드
4. 작업 실행

## Skills 목록
- moai-alfred-tag-scanning: TAG 스캔 및 검증
- moai-alfred-trust-validation: TRUST 5원칙 검증
- moai-alfred-ears-authoring: EARS 요구사항 작성
```

---

## 5. 기술적 접근 방법

### 5.1 JIT 로딩 구현

**Alfred의 JIT 참조 전략**:
```python
def handle_user_request(request: str):
    # 1. 요청 분석
    intent = analyze_intent(request)

    # 2. 필요한 Skill 식별
    required_skills = identify_required_skills(intent)

    # 3. JIT 로드 (필요한 것만)
    for skill in required_skills:
        load_skill(skill)  # ≤100ms

    # 4. 작업 실행
    execute_task(intent, required_skills)
```

### 5.2 Skills 파일 구조

**표준 Skill 구조**:
```markdown
---
name: moai-alfred-{domain}
description: {간단한 설명}
category: guidance
tags:
  - {tag1}
  - {tag2}
---

# {Skill 제목}

## Overview
{Skill 개요}

## How it works
{동작 원리}

## Usage
{사용법}

## Examples
{예제}

## Best Practices
{베스트 프랙티스}

## Troubleshooting
{트러블슈팅}
```

---

## 6. 리스크 관리

### 6.1 높은 리스크

**리스크**: 기존 호출 방식 호환성 깨짐
**완화 방안**:
- Phase 1 완료 후 철저한 기능 테스트
- 롤백 계획 수립 (Agent 파일 복원)
- 단계별 마이그레이션 (한 번에 하나씩)

### 6.2 중간 리스크

**리스크**: JIT 로딩 성능 저하
**완화 방안**:
- 벤치마크 테스트 수행 (목표: ≤100ms)
- Skill 파일 크기 제한 (≤500KB)
- 캐싱 전략 고려 (필요 시)

### 6.3 낮은 리스크

**리스크**: 문서 불일치
**완화 방안**:
- 자동화 스크립트 활용
- 문서 검증 체크리스트 작성
- Phase 3에서 최종 검증

---

## 7. 우선순위별 마일스톤

### 마일스톤 1: Phase 1 완료 (High Priority)
- [ ] tag-agent 마이그레이션
- [ ] trust-checker 마이그레이션
- [ ] moai-alfred-tag-scanning Skill 생성
- [ ] moai-alfred-trust-validation Skill 생성
- [ ] Agent 파일 제거 (선택: 기능 검증 후)

### 마일스톤 2: Phase 2 완료 (Medium Priority)
- [ ] spec-builder EARS 부분 분리
- [ ] spec-builder.md 업데이트
- [ ] SPEC 메타데이터 참조 최적화

### 마일스톤 3: Phase 3 완료 (High Priority)
- [ ] 기능 검증 (3개 테스트 시나리오)
- [ ] 성능 측정 (LOC 감소율, JIT 로딩 시간)
- [ ] 문서 업데이트 (CLAUDE.md, development-guide.md)
- [ ] 최종 검증 및 승인

---

## 8. 다음 단계 가이드

### 8.1 구현 시작
```bash
# Phase 1 시작
/alfred:2-run UPDATE-004

# 예상 작업
1. tag-agent.md 분석
2. moai-alfred-tag-scanning/skill.md 생성
3. tag-agent.md 업데이트 (실행 로직만)
4. trust-checker.md 분석
5. moai-alfred-trust-validation/skill.md 생성
6. trust-checker.md 업데이트 (실행 로직만)
```

### 8.2 검증
```bash
# 기능 테스트
@agent-tag-agent "AUTH 도메인 TAG 목록 조회"
@agent-trust-checker "현재 프로젝트 TRUST 원칙 준수도 확인"
/alfred:1-plan "테스트 기능"

# LOC 측정
wc -l .claude/agents/alfred/{tag-agent,trust-checker,spec-builder}.md
```

### 8.3 문서 동기화
```bash
# Phase 3 완료 후
/alfred:3-sync

# CLAUDE.md, development-guide.md 자동 업데이트
```

---

**최초 작성일**: 2025-10-19
