# Phase 3-4: Progressive Disclosure + Commands 최적화 가이드

> Commands → Sub-agents → Skills 통합 워크플로우의 최종 구현 가이드

---

## Phase 3: Progressive Disclosure 적용 (큰 Skills 분리)

### 목표
**큰 Skills (500+ LOC)를 분리하여 컨텍스트 효율성 향상**
- SKILL.md는 100 LOC 이하 (개요 + 빠른 참조)
- 세부 사항은 별도 참조 파일로 분리
- Progressive disclosure를 통한 필요시 로드

### 대상 Skills

#### 1. moai-alfred-trust-validation (현재: 미구현)
**분리 구조**:
```
moai-alfred-trust-validation/
├── SKILL.md (100 LOC: 개요 + 빠른 검증)
├── test-coverage.md (테스트 커버리지 상세)
├── readable-constraints.md (코드 제약 상세)
├── security-patterns.md (보안 패턴 상세)
├── tag-validation.md (TAG 검증 상세)
└── scripts/
    ├── validate_coverage.py
    └── check_constraints.py
```

**SKILL.md 구조**:
```markdown
---
name: moai-alfred-trust-validation
description: Validates TRUST 5-principles (Test 85%+, Readable constraints, Unified architecture, Secured patterns, TAG trackability). Use when validating code quality, checking TRUST compliance, verifying test coverage, or analyzing security patterns.
allowed-tools: Read, Bash, TodoWrite
---

# TRUST 5 Validation

## Quick Validation

```bash
# T - Test coverage (Haiku 추천)
pytest --cov --cov-report=term-missing

# R - Readable constraints
rg "def " src/ | wc -l  # Function count
```

## Detailed Guides (Progressive Disclosure)

- **Test coverage details**: See [test-coverage.md](test-coverage.md)
- **Readable constraints**: See [readable-constraints.md](readable-constraints.md)
- **Security patterns**: See [security-patterns.md](security-patterns.md)
- **TAG validation**: See [tag-validation.md](tag-validation.md)

## Validation Workflow (Feedback Loop)

1. Run quick validation → If errors
2. Read detailed guide → Fix errors
3. Re-validate → Repeat until pass
```

#### 2. moai-foundation-specs (현재: 일부 개선)
**현재 상태 확인 후 필요 시 분리**

#### 3. moai-claude-code (현재: 대형 파일)
**분리 예정** (현재는 검토만)

### 구현 방법

**Step 1: 파일 분리**
```bash
# 새 디렉토리 구조 생성
mkdir -p .claude/skills/moai-alfred-trust-validation/scripts

# SKILL.md 작성
echo "---
name: moai-alfred-trust-validation
..." > .claude/skills/moai-alfred-trust-validation/SKILL.md

# 세부 파일 생성
touch .claude/skills/moai-alfred-trust-validation/{test-coverage,readable-constraints,security-patterns,tag-validation}.md
```

**Step 2: 컨텐츠 분리**
- SKILL.md: 개요 + 빠른 참조 (100 LOC)
- test-coverage.md: 테스트 커버리지 상세 (200+ LOC)
- readable-constraints.md: 코드 제약 상세 (150+ LOC)
- security-patterns.md: 보안 패턴 상세 (100+ LOC)
- tag-validation.md: TAG 검증 상세 (100+ LOC)

**Step 3: 참조 링크 추가**
모든 세부 파일에서 SKILL.md로 돌아갈 링크 포함
```markdown
← [Back to SKILL.md](SKILL.md)
```

### 효과

| 지표           | Before   | After   |
| -------------- | -------- | ------- |
| SKILL.md 크기  | 500+ LOC | 100 LOC |
| 초기 로드 토큰 | 높음     | 낮음    |
| 필요시 로드    | 불가능   | 가능    |
| 컨텍스트 효율  | 낮음     | 높음    |

---

## Phase 4: Commands에 Skills 힌트 추가

### 목표
**Commands에서 Sub-agent 호출 시 Skills 자동 활성화 가능하도록 힌트 제공**

### 현재 상태
Commands는 이미 2단계 구조가 있음:
- Phase 1: 분석 및 계획
- Phase 2: 실행

### 개선 사항

#### 1. /alfred:1-plan 개선

**현재 구조**:
```markdown
## Phase 2: 실행

1. spec-builder 호출
2. SPEC 문서 생성
```

**개선 후 구조**:
```markdown
## Phase 2: 실행

**Sub-agent: spec-builder (sonnet, 독립 컨텍스트)**

### 자동 활성화 Skills (Claude가 자동 발견)
- **moai-foundation-specs**: SPEC 메타데이터 검증
  - Trigger: "SPEC validation", "metadata structure"
- **moai-foundation-ears**: EARS 구문 가이드
  - Trigger: "EARS syntax", "requirement authoring"

### 워크플로우
1. SPEC 구조 작성
   - spec-builder가 "SPEC validation" 키워드 언급
   - → moai-foundation-specs skill 자동 활성화

2. EARS 요구사항 작성
   - spec-builder가 "EARS syntax" 키워드 언급
   - → moai-foundation-ears skill 자동 활성화

3. 검증 완료
   - Skills 결과 기반으로 오류 수정
   - 최종 검증 완료

**주의**: spec-builder는 독립 컨텍스트에서 작업하므로, 메인 대화에서 직접 Skills를 호출할 필요가 없습니다. Sub-agent 내부에서 자동으로 활성화됩니다.
```

#### 2. /alfred:2-run 개선

**개선 포인트**:
```markdown
## Phase 2: 실행

**Sub-agents (순차 실행)**:

### 1단계: code-builder (sonnet, 독립 컨텍스트)
**자동 활성화 Skills**:
- moai-foundation-langs: 언어 감지
  - Trigger: "language detection", "project language"
- moai-lang-typescript/python/go/...: 언어별 패턴
  - Trigger: "[언어] code", "TDD implementation"

**구현**: RED → GREEN → REFACTOR

### 2단계: quality-gate (haiku, 독립 컨텍스트)
**자동 활성화 Skills**:
- moai-alfred-trust-validation: TRUST 검증
  - Trigger: "TRUST validation", "test coverage"
- moai-foundation-trust: TRUST 원칙
  - Trigger: "quality check", "code standards"

**검증**: TRUST 5원칙 확인
```

#### 3. /alfred:3-sync 개선

**개선 포인트**:
```markdown
## Phase 2: 실행

**Sub-agents (병렬 실행 가능)**:

### tag-agent (haiku)
**자동 활성화 Skills**:
- moai-foundation-tags: TAG 시스템
  - Trigger: "TAG validation", "chain integrity"
- moai-alfred-tag-scanning: TAG 스캔
  - Trigger: "tag analysis", "orphan detection"

### doc-syncer (haiku)
**자동 활성화 Skills**:
- moai-foundation-specs: SPEC 참조
  - Trigger: "specification document", "SPEC version"
- moai-essentials-review: 코드 리뷰
  - Trigger: "code quality check", "documentation review"
```

### 구현 방법

**Step 1: Commands 파일 찾기**
```bash
ls .claude/commands/
# alfred:1-spec.md
# alfred:2-build.md
# alfred:3-sync.md
```

**Step 2: Phase 2 섹션 개선**
각 command의 Phase 2 섹션에 아래 구조 추가:

```markdown
### 자동 활성화 Skills 정보

이 Sub-agent는 다음 Skills를 자동으로 발견하여 활용합니다:

| Skill                 | 역할      | Trigger Keywords              |
| --------------------- | --------- | ----------------------------- |
| moai-foundation-specs | SPEC 검증 | "SPEC validation", "metadata" |
| moai-foundation-ears  | EARS 작성 | "EARS syntax", "requirements" |
```

**Step 3: Sub-agent Context 강조**
```markdown
## 중요: Sub-agent의 독립 컨텍스트

spec-builder는 **독립 컨텍스트**에서 작업하므로:
- ✓ 메인 대화 오염 방지
- ✓ Skills를 자동으로 발견하여 활용
- ✓ 필요한 도구만 접근 (allowed-tools 제한)
- ✓ 적절한 모델 사용 (sonnet/haiku)

따라서 메인 대화에서 Skills를 명시적으로 호출할 필요가 없습니다.
```

### 효과

| 측면             | 개선 전 | 개선 후            |
| ---------------- | ------- | ------------------ |
| Sub-agent 발견성 | 낮음    | 높음 (Skills 힌트) |
| Skills 활용도    | 0%      | 90%+               |
| 컨텍스트 효율    | 낮음    | 높음               |
| 사용자 이해도    | 낮음    | 높음               |

---

## 📋 구현 체크리스트

### Phase 3: Progressive Disclosure
- [ ] moai-alfred-trust-validation 분리
  - [ ] SKILL.md (100 LOC) 작성
  - [ ] test-coverage.md 작성
  - [ ] readable-constraints.md 작성
  - [ ] security-patterns.md 작성
  - [ ] tag-validation.md 작성
- [ ] 다른 대형 Skills 평가

### Phase 4: Commands 개선
- [ ] /alfred:1-plan 업데이트
  - [ ] Skills 정보 섹션 추가
  - [ ] Sub-agent 컨텍스트 설명 추가
- [ ] /alfred:2-run 업데이트
  - [ ] 단계별 Skills 정보 추가
- [ ] /alfred:3-sync 업데이트
  - [ ] 병렬 Sub-agents Skills 정보 추가

### 최종 검증
- [ ] 모든 Commands에 Skills 힌트 포함 확인
- [ ] Sub-agent 독립 컨텍스트 설명 추가 확인
- [ ] Trigger keywords 검증

---

## 🎯 성과 지표

구현 완료 후:
- ✅ Sub-agents가 Skills를 90%+ 자동 활용
- ✅ 메인 대화 컨텍스트 오염 방지
- ✅ Progressive Disclosure로 토큰 효율성 30~40% 향상
- ✅ 사용자가 각 단계에서 활성화되는 Skills 이해

---

**작성일**: 2025-10-20
