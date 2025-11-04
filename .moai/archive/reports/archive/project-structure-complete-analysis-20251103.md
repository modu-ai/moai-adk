# MoAI-ADK 프로젝트 구조 완전 분석 및 재구성

**작성일**: 2025-11-03 추가 분석
**분석 대상**: 패키지 자체 vs 템플릿 구조 재검토
**목표**: 올바른 파일 배치 및 메모리→스킬 마이그레이션 완료

---

## 📊 Executive Summary

이전 분석에서 놓친 핵심 문제를 재발견했습니다:

| 파일/디렉토리 | 위치 | 용도 | 상태 |
|------------|------|------|------|
| `core/tags/` | 2곳 | 태그 검증 | ⚠️ 불일치 |
| `workflows/` | 2곳 | CI/CD | ⚠️ 불일치 |
| `.moai/memory/` | 2곳 | 가이드 | ⚠️ PIL 미완료 |

---

## 🔍 핵심 발견사항

### 1. core/tags 구조 불일치

**Template** (`src/moai_adk/templates/src/moai_adk/core/tags/`):
```
pre_commit_validator.py    (10KB, Oct 30)
reporter.py               (29KB, Nov 2)
```

**Actual** (`src/moai_adk/core/tags/`):
```
__init__.py              (2KB)
ci_validator.py          (15KB)  ❌ 템플릿에 없음
cli.py                   (7KB)   ❌ 템플릿에 없음
generator.py             (3KB)   ❌ 템플릿에 없음
inserter.py              (3KB)   ❌ 템플릿에 없음
mapper.py                (4KB)   ❌ 템플릿에 없음
parser.py                (2KB)   ❌ 템플릿에 없음
pre_commit_validator.py  (10KB)  ✅ 있음
reporter.py              (29KB)  ✅ 있음
tags.py                  (4KB)   ❌ 템플릿에 없음
validator.py             (31KB)  ❌ 템플릿에 없음
```

**분석**:
- 템플릿: 최소 2개 파일 (검증만)
- 실제: 전체 11개 파일 (완전한 TAG 시스템)
- **이유**: 새 프로젝트는 최소 검증 기능만 필요, MoAI-ADK는 완전한 구현 필요

✅ **결론**: 이것은 **올바른 배치**. 템플릿은 최소화되어 있어야 함.

### 2. workflows 구조 비교

**Template** (`src/moai_adk/templates/workflows/`):
```
go-tag-validation.yml              (언어별)
javascript-tag-validation.yml      (언어별)
python-tag-validation.yml          (언어별)
typescript-tag-validation.yml      (언어별)
```

**Actual** (`.github/workflows/`):
```
moai-gitflow.yml                   (패키지용)
moai-release-create.yml            (패키지용)
moai-release-pipeline.yml          (패키지용)
release.yml                        (패키지용)
spec-issue-sync.yml                (패키지용)
tag-report.yml                     (패키지용)
tag-validation.yml                 (패키지용)
```

**분석**:
- 템플릿: 새 프로젝트가 사용할 언어별 TAG 검증 워크플로우
- 실제: MoAI-ADK 패키지 자체의 릴리스/배포 워크플로우
- 서로 다른 용도이므로 분리 정상

✅ **결론**: 이것도 **올바른 배치**. 용도가 다름.

---

## ⚠️ 실제 문제: .moai/memory/ 마이그레이션 미완료

### 현재 상태

**Git 히스토리 분석 결과**:

| 커밋 | 날짜 | 작업 | 상태 |
|------|------|------|------|
| 7aace4f7 | Nov 2 17:47 | `.moai/memory/` → `.claude/skills/` 마이그레이션 계획 | 📝 제안 |
| 2be3f613 | Nov 1 23:54 | development-guide.md 참조 → Skill() 전환 | 🔄 부분 완료 |
| a60fd6b0 | Nov 2 21:44 | CLAUDE-*.md 파일 삭제 (3개) | ✅ 완료 |
| 011e19c9 | Nov 2 22:05 | Persona System Upgrade v1.0.0 (새 메모리 파일 추가) | ⚠️ 역행 |

**문제**:
1. 메모리 → 스킬 마이그레이션이 **부분적으로만** 완료
2. 최신 업그레이드에서 **다시 메모리 파일 추가**됨
3. 로컬 프로젝트에 **과도한 개발 문서** 축적

### 메모리 파일 현황

**템플릿** (`src/moai_adk/templates/.moai/memory/`):
```
✅ development-guide.md                (14KB, 핵심 가이드)
✅ spec-metadata.md                    (SPEC 표준)
✅ gitflow-protection-policy.md        (팀 협업 정책)
✅ issue-label-mapping.md              (GitHub 라벨)
✅ skills-description-policy.md        (스킬 정책)
✅ project-notes.json                  (상태 JSON)
✅ session-hint.json                   (상태 JSON)
✅ user-patterns.json                  (상태 JSON)
```

**로컬** (`.moai/memory/`):
```
✅ 템플릿과 동일한 파일들
⚠️ + claude-code-features-guide.md    (로컬 개발 고유)
⚠️ + command-execution-state.json     (로컬 개발 고유)
⚠️ + config-schema.md                 (로컬 개발 고유)
⚠️ + language-policy-final.md         (로컬 개발 고유)
⚠️ + session-state.md                 (로컬 개발 고유)
⚠️ + subagent-execution.log           (로컬 개발 고유)
```

### 상황 분석

**메모리 파일은 두 가지 용도로 혼용되고 있습니다:**

1. **Static Knowledge** (모든 프로젝트가 필요):
   - development-guide.md → Skill("moai-alfred-dev-guide") 로 변환 필요
   - spec-metadata.md → Skill("moai-alfred-spec-metadata-extended") 로 변환 필요
   - gitflow-protection-policy.md → Skill("moai-alfred-gitflow-policy") 로 변환 필요

2. **Session State** (프로젝트 실행 중 생성):
   - user-patterns.json ✅ JSON 형식 유지 (세션 데이터)
   - session-hint.json ✅ JSON 형식 유지 (세션 데이터)
   - project-notes.json ✅ JSON 형식 유지 (세션 데이터)

**Current Decision** (commit 011e19c9):
- JSON 세션 파일만 남김 ✅
- Markdown 가이드도 남김 ⚠️ (Skill로 변환되어야 함)

---

## 🎯 권장 전략: PIL(Progressive Information Loading) 최적화

### 현재 스킬 현황

**이미 존재하는 관련 스킬** (62개 스킬 중):

| 스킬명 | 위치 | 파일 수 | 용도 |
|--------|------|--------|------|
| moai-foundation-trust | .claude/skills/ | 3 | TRUST 5 원칙 |
| moai-foundation-tags | .claude/skills/ | 3 | TAG 라이프사이클 |
| moai-foundation-specs | .claude/skills/ | 3 | SPEC 작성 |
| moai-foundation-ears | .claude/skills/ | 3 | EARS 요구사항 |
| moai-alfred-reporting | .claude/skills/ | 2 | 리포팅 패턴 |
| moai-alfred-workflow | .claude/skills/ | 1 | 워크플로우 |
| moai-cc-memory | .claude/skills/ | 3 | 메모리 관리 |

**누락된 스킬** (생성 필요):

```
moai-alfred-dev-guide
├── SKILL.md (1000-1500 자)
├── reference.md (2000-3000 자, 명령어 예제)
└── examples.md (실제 사용 사례)

moai-alfred-gitflow-policy
├── SKILL.md (1000-1500 자)
├── reference.md (정책, 규칙)
└── examples.md (팀 협업 사례)

moai-alfred-spec-metadata-extended
├── SKILL.md (1000 자)
├── reference.md (필드 설명)
└── examples.md (SPEC 예제)
```

### PIL 구현 계획

#### Phase 1: 메모리 파일 삭제 및 스킬 생성

**제거할 파일**:
```
src/moai_adk/templates/.moai/memory/
├── ❌ DEVELOPMENT-GUIDE.md → Skill("moai-alfred-dev-guide")
├── ❌ GITFLOW-PROTECTION-POLICY.md → Skill("moai-alfred-gitflow-policy")
└── ✅ user-patterns.json, session-hint.json, project-notes.json (유지)

.moai/memory/ (로컬)
├── ❌ development-guide.md → Skill 호출
├── ❌ gitflow-protection-policy.md → Skill 호출
└── ✅ JSON 파일들 유지
```

**생성할 스킬**:
```
src/moai_adk/templates/.claude/skills/moai-alfred-dev-guide/
├── SKILL.md (개발 가이드 핵심, 1200 자)
├── reference.md (명령어, API, 패턴)
└── examples.md (실제 사례)

src/moai_adk/templates/.claude/skills/moai-alfred-gitflow-policy/
├── SKILL.md (정책 개요, 1000 자)
├── reference.md (규칙, 체크리스트)
└── examples.md (팀 협업 시나리오)

src/moai_adk/templates/.claude/skills/moai-alfred-spec-metadata-extended/
├── SKILL.md (SPEC 메타 개요)
├── reference.md (필드 정의, 유효성 검사)
└── examples.md (SPEC 템플릿)
```

#### Phase 2: 참조 변경

**변경 대상**:
```
.claude/agents/alfred/*.md
├── 현재: "@.moai/memory/development-guide.md"
├── 변경: Skill("moai-alfred-dev-guide") ← 자동 로드

.claude/commands/alfred/*.md
├── 현재: ".moai/memory/spec-metadata.md"
├── 변경: Skill("moai-alfred-spec-metadata-extended") ← JIT 로드

.moai/config.json
├── 현재: "docs_directory": ".moai/docs"
├── 변경: 불필요 (로컬 분석 파일용으로만 사용)
```

#### Phase 3: 로컬 프로젝트 정리

**유지할 파일** (`.moai/`):
```
.moai/memory/
├── user-patterns.json          ✅ 세션 데이터
├── session-hint.json           ✅ 세션 데이터
└── project-notes.json          ✅ 세션 데이터

.moai/docs/
├── guide-*.md                  ✅ 프로젝트 고유 가이드
├── exploration-*.md            ✅ 분석 문서 (아카이브 가능)
└── implementation-*.md         ✅ 구현 기록 (아카이브 가능)

.moai/specs/                    ✅ SPEC 문서 (유지)
.moai/reports/                  ✅ 동기화 리포트 (유지)
```

**정리할 파일** (아카이브 또는 삭제):
```
.moai/memory/
├── ❌ claude-code-features-guide.md     → .moai/archive/
├── ❌ command-execution-state.json      → 불필요 (최신 시스템에서 생성)
├── ❌ config-schema.md                  → .moai/archive/
├── ❌ language-policy-final.md          → .moai/archive/
├── ❌ session-state.md                  → .moai/archive/
└── ❌ subagent-execution.log            → 불필요 (로그 파일)

.moai/docs/
├── ⚠️ exploration-*.md                  → .moai/archive/exploration/
├── ⚠️ implementation-*.md               → .moai/archive/implementation/
└── ⚠️ shell-testing-index.md            → Skill 호출로 변경 권장
```

---

## 📋 실행 계획 (3단계)

### Step 1: 스킬 생성 (Priority: HIGH)

```bash
# 1. moai-alfred-dev-guide 스킬 생성
mkdir -p src/moai_adk/templates/.claude/skills/moai-alfred-dev-guide
# development-guide.md 내용을 SKILL.md (1200자)로 압축
# 명령어, 예제를 reference.md로 정리
# 실제 사용 사례를 examples.md로 작성

# 2. moai-alfred-gitflow-policy 스킬 생성
mkdir -p src/moai_adk/templates/.claude/skills/moai-alfred-gitflow-policy
# gitflow-protection-policy.md를 SKILL.md로 압축
# 정책, 규칙을 reference.md로 정리
# 팀 시나리오를 examples.md로 작성

# 3. moai-alfred-spec-metadata-extended 스킬 생성
# 또는 기존 moai-alfred-spec-metadata-validation 확장
```

### Step 2: 참조 변경 (Priority: HIGH)

```bash
# 1. 모든 에이전트 파일에서 Skill() 호출로 변경
.claude/agents/alfred/*.md
  ".moai/memory/development-guide.md" → Skill("moai-alfred-dev-guide")
  ".moai/memory/gitflow-protection-policy.md" → Skill("moai-alfred-gitflow-policy")

# 2. 커맨드 파일 업데이트
.claude/commands/alfred/*.md
  ".moai/memory/spec-metadata.md" → Skill("moai-alfred-spec-metadata-extended")

# 3. 훅 파일 검증
.claude/hooks/alfred/*/context.py
  필요시 Skill 호출로 변경
```

### Step 3: 파일 정리 및 동기화 (Priority: MEDIUM)

```bash
# 1. 템플릿 메모리 파일 삭제
rm src/moai_adk/templates/.moai/memory/DEVELOPMENT-GUIDE.md
rm src/moai_adk/templates/.moai/memory/GITFLOW-PROTECTION-POLICY.md

# 2. 로컬 메모리 파일 아카이브 (유지하되 로드하지 않음)
mkdir -p .moai/archive/memory
mv .moai/memory/claude-code-features-guide.md .moai/archive/memory/
mv .moai/memory/config-schema.md .moai/archive/memory/
mv .moai/memory/language-policy-final.md .moai/archive/memory/

# 3. JSON 세션 파일만 .moai/memory/에 유지
ls -la .moai/memory/ | grep "\.json"

# 4. .moai/docs 파일도 필요시 아카이브
mkdir -p .moai/archive/docs
# 로컬 분석 문서만 아카이브
```

---

## ✅ 결론: 프로젝트 구조 현황

### 올바른 배치

✅ **`src/moai_adk/templates/src/moai_adk/core/tags/`**
- 새 프로젝트용 최소 TAG 검증 파일 포함
- 실제 프로젝트와 분리 정상

✅ **`src/moai_adk/templates/workflows/`**
- 새 프로젝트용 언어별 TAG 검증 워크플로우
- 패키지 워크플로우와 분리 정상

### 해결 필요

⚠️ **`.moai/memory/` 파일들**
- Markdown 가이드 → Skill 마이그레이션 필요
- JSON 세션 파일 → 유지 필요
- PIL(Progressive Information Loading) 완성 필요

⚠️ **로컬 개발 문서 축적**
- `.moai/docs/`: 13개 분석/탐색 문서
- 아카이브 또는 정리 필요
- 일부는 Wiki로 이동 권장

---

## 🚀 다음 단계

1. **모든 3개 스킬 생성**
2. **모든 참조 Skill() 호출로 변경**
3. **메모리 파일 정리 및 동기화**
4. **git 커밋 및 패키지 릴리스**

예상 효과:
- 메모리 파일 크기: 100+ KB → 5 KB (JSON만)
- 스킬 로드: JIT (Just-In-Time) 방식
- 컨텍스트 절감: ~15-20%
