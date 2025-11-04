# .moai 디렉토리 구조 분석 및 권장사항

**작성일**: 2025-11-03
**분석 대상**: 패키지 템플릿 vs 로컬 프로젝트
**상태**: 주요 구조적 차이점 및 명확화 필요

---

## 📊 Executive Summary

MoAI-ADK 프로젝트는 두 가지 중요한 `.moai` 디렉토리 구조를 관리합니다:

| 위치 | 용도 | 상태 |
|------|------|------|
| `src/moai_adk/templates/.moai/` | **패키지 템플릿** (새 프로젝트용) | ✅ 정의됨 |
| `/Users/goos/MoAI/MoAI-ADK/.moai/` | **로컬 개발 프로젝트** | ⚠️ 불일치 |

**주요 발견사항**:
- 템플릿과 로컬 프로젝트 간 문서 구조 불일치
- 파일명 대소문자 불일치 (케이스 센시티비티 이슈)
- 로컬 프로젝트 고유 파일과 템플릿 파일의 경계가 불명확

---

## 🏗️ 템플릿 구조 (Source of Truth)

### Template `.moai` 디렉토리 파일 목록

```
src/moai_adk/templates/.moai/
├── config.json (템플릿 버전, 변수 치환 필요)
├── docs/
│   └── guide-alfred-persona-integration.md
├── logs/
│   └── skill-usage.json
├── memory/
│   ├── DEVELOPMENT-GUIDE.md
│   ├── ISSUE-LABEL-MAPPING.md
│   ├── SKILLS-DESCRIPTION-POLICY.md
│   ├── gitflow-protection-policy.md
│   ├── project-notes.json
│   ├── session-hint.json
│   ├── spec-metadata.md
│   └── user-patterns.json
├── project/
│   ├── product.md
│   ├── structure.md
│   └── tech.md
└── reports/
    ├── comprehensive-test-results-2025-11-02.md
    └── persona-system-validation-2025-11-02.md
```

**특징**:
- 최소한의 핵심 문서만 포함
- 새 프로젝트 시작 시 필요한 기본 구조
- 각 디렉토리의 용도가 명확함

---

## 📁 로컬 프로젝트 구조

### Local `.moai` 디렉토리 파일 목록

#### `.moai/memory/` (16개 파일 + archive)

**템플릿에서 제공되는 파일**:
- ✅ SKILLS-DESCRIPTION-POLICY.md
- ✅ gitflow-protection-policy.md
- ✅ project-notes.json
- ✅ session-hint.json
- ✅ spec-metadata.md
- ✅ user-patterns.json

**로컬 개발 중 추가된 파일** (프로젝트 고유):
- ❓ claude-code-features-guide.md
- ❓ command-execution-state.json
- ❓ config-schema.md
- ❓ **development-guide.md** ⚠️ (대소문자 불일치)
- ❓ **issue-label-mapping.md** ⚠️ (대소문자 불일치)
- ❓ language-policy-final.md
- ❓ session-state.md
- ❓ subagent-execution.log

#### `.moai/docs/` (14개 파일 + 1개 디렉토리)

**템플릿에서 제공되는 파일**:
- ✅ guide-alfred-persona-integration.md

**로컬 개발 중 추가된 파일** (모두 프로젝트 고유):
- ❓ README-sync-report.md
- ❓ SPEC-HOOKS-EMERGENCY-001-completion-summary.md
- ❓ alfred-command-completion-guide.md
- ❓ exploration-alfred-architecture-20251102.md
- ❓ exploration-update-cache-fix-001.md
- ❓ feature-integration/ (디렉토리)
- ❓ github-label-guide.md
- ❓ implementation-SPEC-SESSION-CLEANUP-001.md
- ❓ language-detection-guide.md
- ❓ persona-system-skills-summary.md
- ❓ powershell-testing-guide.md
- ❓ shell-testing-index.md
- ❓ workflow-templates.md

---

## 🔍 주요 문제점

### 1. ⚠️ 파일명 대소문자 불일치

**문제**: 템플릿의 파일명과 로컬 파일명이 다른 경우

| 템플릿 파일 | 로컬 파일 | 상태 |
|------------|----------|------|
| `DEVELOPMENT-GUIDE.md` | `development-guide.md` | ❌ 대소문자 불일치 |
| `ISSUE-LABEL-MAPPING.md` | `issue-label-mapping.md` | ❌ 대소문자 불일치 |
| `SPEC-METADATA.md` | `spec-metadata.md` | ❌ 혼용 (둘 다 있음) |
| `GITFLOW-PROTECTION-POLICY.md` | `gitflow-protection-policy.md` | ❌ 혼용 (둘 다 있음) |

**영향**:
- 파일 참조가 일관성 없음
- 새 프로젝트 생성 시 어떤 버전이 복사될지 불명확
- 템플릿 동기화 프로세스에서 충돌 가능

### 2. ⚠️ 템플릿 vs 로컬 파일 경계 불명확

**문제**: 다음 파일들이 템플릿에 있어야 하는지, 로컬 고유인지 명확하지 않음

```
필요한 명확화가 필요한 파일:
├── development-guide.md (또는 DEVELOPMENT-GUIDE.md)
│   - Alfred 전체 개발 가이드
│   - 17개 에이전트와 55개 스킬에서 참조됨
│   - ✅ 이것은 **템플릿에 포함되어야 함**
│
├── spec-metadata.md (또는 SPEC-METADATA.md)
│   - SPEC 메타데이터 표준 정의
│   - 모든 SPEC 생성 시 필요
│   - ✅ 이것은 **템플릿에 포함되어야 함**
│
├── gitflow-protection-policy.md (또는 GITFLOW-PROTECTION-POLICY.md)
│   - 팀 모드 GitFlow 정책
│   - 팀 프로젝트에 필수
│   - ✅ 이것은 **템플릿에 포함되어야 함**
│
└── issue-label-mapping.md (또는 ISSUE-LABEL-MAPPING.md)
    - GitHub 라벨 매핑
    - 팀 협업에 필수
    - ✅ 이것은 **템플릿에 포함되어야 함**
```

### 3. 📚 `.moai/docs/`의 프로젝트 고유 파일들

**현재 상태**: 로컬 개발 중 생성된 분석/탐색/구현 문서

```
이 파일들은 로컬 프로젝트 고유이며, 템플릿에 포함되지 않음:

프로세스 문서 (MoAI-ADK 개발용):
- alfred-command-completion-guide.md (Alfred 명령 완료 패턴)
- exploration-alfred-architecture-20251102.md (아키텍처 분석)
- guide-alfred-persona-integration.md ✅ (템플릿에 있음)
- persona-system-skills-summary.md (페르소나 시스템 문서)
- implementation-SPEC-SESSION-CLEANUP-001.md (SPEC 구현 기록)

기술 가이드 (언어/도구별):
- language-detection-guide.md (언어 감지 가이드)
- shell-testing-index.md (셸 테스트 인덱스)
- powershell-testing-guide.md (PowerShell 테스트)
- workflow-templates.md (워크플로우 템플릿)
- github-label-guide.md (라벨 가이드)

기타:
- README-sync-report.md (동기화 리포트)
- SPEC-HOOKS-EMERGENCY-001-completion-summary.md (SPEC 완료 요약)
- exploration-update-cache-fix-001.md (캐시 수정 분석)

⚠️ 이 파일들은 새 프로젝트 템플릿에 포함되면 안 됨 (MoAI-ADK 개발 고유)
```

---

## 📍 파일 사용처 분석

### 가장 자주 참조되는 파일

```bash
development-guide.md (또는 DEVELOPMENT-GUIDE.md)
├─ 참조 위치: 17개 에이전트, 스킬, 훅에서 참조
├─ 참조 패턴: ".moai/memory/development-guide.md"
├─ 사용 목적: Alfred 핵심 지침, TRUST 원칙, TAG 체인, TDD 가이드
├─ 필수 여부: ✅ 필수 (모든 프로젝트)
└─ 템플릿 포함: ⚠️ 불확실 (현재 대소문자 혼용)

spec-metadata.md (또는 SPEC-METADATA.md)
├─ 참조 위치: spec-builder 에이전트, 1-plan 명령에서 참조
├─ 참조 패턴: ".moai/memory/spec-metadata.md"
├─ 사용 목적: SPEC 메타데이터 표준, YAML 필드 정의
├─ 필수 여부: ✅ 필수 (SPEC 생성 시)
└─ 템플릿 포함: ⚠️ 불확실 (현재 대소문자 혼용)

gitflow-protection-policy.md (또는 GITFLOW-PROTECTION-POLICY.md)
├─ 참조 위치: git-manager 에이전트에서 참조
├─ 참조 패턴: ".moai/memory/gitflow-protection-policy.md"
├─ 사용 목적: GitFlow 정책, PR 기본 브랜치 설정
├─ 필수 여부: ✅ 필수 (팀 모드 프로젝트)
└─ 템플릿 포함: ✅ 있음 (lowercase)

issue-label-mapping.md (또는 ISSUE-LABEL-MAPPING.md)
├─ 참조 위치: 여러 GitHub 워크플로우에서 참조
├─ 참조 패턴: ".moai/memory/issue-label-mapping.md"
├─ 사용 목적: GitHub 이슈 라벨 매핑
├─ 필수 여부: ✅ 필수 (GitHub 통합 시)
└─ 템플릿 포함: ⚠️ 불확실 (대소문자 혼용)
```

---

## 🔧 권장 사항

### Phase 1: 템플릿 정규화 (Priority: HIGH)

#### 1.1 파일명 대소문자 통일

**현재 상태**:
```
src/moai_adk/templates/.moai/memory/
├── DEVELOPMENT-GUIDE.md (대문자)
├── ISSUE-LABEL-MAPPING.md (대문자)
├── SKILLS-DESCRIPTION-POLICY.md (대문자)
├── gitflow-protection-policy.md (소문자)
├── spec-metadata.md (소문자)
└── (혼용)
```

**권장**:
모든 파일을 **소문자 하이픈** 형식으로 통일

```
src/moai_adk/templates/.moai/memory/
├── development-guide.md ✅
├── issue-label-mapping.md ✅
├── skills-description-policy.md ✅
├── gitflow-protection-policy.md ✅
└── spec-metadata.md ✅
```

**이유**:
- 리눅스/맥 파일시스템의 관례
- 참조 일관성
- 새 프로젝트 복사 시 확실성
- Git 이력 추적 단순화

#### 1.2 누락된 핵심 문서 추가

템플릿에 다음 파일 추가 필요:

```
src/moai_adk/templates/.moai/memory/
├── development-guide.md (현재 없음 - 추가 필요) ⚠️
├── issue-label-mapping.md (현재 대문자 - 정규화)
├── skills-description-policy.md (현재 대문자 - 정규화)
├── gitflow-protection-policy.md ✅ (이미 있음)
├── spec-metadata.md ✅ (이미 있음)
└── (기타 필수 파일들)
```

**검증**: 각 에이전트의 문서 참조 확인

```bash
# development-guide.md 참조 확인
rg "\.moai/memory/development-guide\.md" /Users/goos/MoAI/MoAI-ADK/.claude/ | wc -l
# 결과: 12+ 파일에서 참조

# 현재 템플릿에 있는지 확인
ls /Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/.moai/memory/ | grep -i development
# 결과: DEVELOPMENT-GUIDE.md (존재하지만 대문자)
```

### Phase 2: 로컬 프로젝트 정리 (Priority: MEDIUM)

#### 2.1 `.moai/docs/` 정책 수립

**추천 구조**:

```
.moai/docs/                                    (모든 문서)
├── guide-*.md                                 (사용 가이드)
│   └── guide-alfred-persona-integration.md   (템플릿)
├── exploration-*.md                          (분석/탐색 - 로컬 프로젝트 고유)
├── implementation-SPEC-*.md                  (구현 기록)
├── language-*.md                             (기술 가이드)
└── workflow-*.md                             (워크플로우 문서)
```

**분류**:

| 카테고리 | 파일 | 용도 | 템플릿 포함 | 설명 |
|---------|------|------|-----------|------|
| 가이드 | `guide-*.md` | 사용자 가이드 | ✅ 선택 | 필요한 경우만 |
| 분석 | `exploration-*.md` | 아키텍처/분석 | ❌ 아니오 | 로컬 개발 고유 |
| 구현 | `implementation-*.md` | SPEC 구현 기록 | ❌ 아니오 | 로컬 개발 고유 |
| 기술 | `language-*.md`, `workflow-*.md` | 기술 가이드 | ❓ 검토 필요 | 패키지 문서와 중복 가능 |

#### 2.2 프로젝트 고유 파일 정리

**권장**: 로컬 전용 파일은 별도 위치로 이동

```
.moai/project-analysis/                       (새 디렉토리)
├── exploration-alfred-architecture-20251102.md
├── exploration-update-cache-fix-001.md
├── claude-code-features-guide.md
├── language-detection-guide.md
└── shell-testing-index.md

.moai/docs/                                    (공식 가이드만)
├── guide-alfred-persona-integration.md       (템플릿)
├── language-detection-guide.md                (필요시)
└── workflow-templates.md                      (필요시)
```

### Phase 3: 동기화 프로세스 정립 (Priority: MEDIUM)

#### 3.1 템플릿 동기화 스크립트

```bash
# 추천: 패키지 업그레이드 후 자동 동기화
uv tool upgrade moai-adk

# 그 다음: 템플릿 메모리 파일 동기화 (settings 제외)
rsync -av --exclude="settings*.json" --exclude="*.local.json" \
  src/moai_adk/templates/.moai/ .moai/
```

**주의**: 로컬 개발 파일은 덮어씌우지 않도록 주의

#### 3.2 파일명 일관성 검증

```bash
# 템플릿 파일명 확인
find src/moai_adk/templates/.moai -type f -name "*.md" | \
  xargs -I {} basename {} | \
  sort | uniq

# 로컬 파일명 확인
find .moai -type f -name "*.md" | \
  xargs -I {} basename {} | \
  sort | uniq

# 차이점 확인
comm -3 <(find src/moai_adk/templates/.moai -name "*.md" | xargs -I {} basename {} | sort -u) \
        <(find .moai -name "*.md" | xargs -I {} basename {} | sort -u)
```

---

## 📋 Action Items (우선순위)

### Immediate (이번 주)

- [ ] Phase 1.1 실행: 템플릿 파일명 대소문자 정규화
  - `DEVELOPMENT-GUIDE.md` → `development-guide.md`
  - `ISSUE-LABEL-MAPPING.md` → `issue-label-mapping.md`
  - `SKILLS-DESCRIPTION-POLICY.md` → `skills-description-policy.md`
  - `GITFLOW-PROTECTION-POLICY.md` → `gitflow-protection-policy.md` (이미 있음)

- [ ] Phase 1.2 실행: 누락된 파일 추가 확인
  - `development-guide.md` 템플릿에 확인/추가
  - Git에 커밋

- [ ] 로컬 프로젝트 파일명 동기화

### Short-term (이번 달)

- [ ] Phase 2.1 실행: `.moai/docs/` 정책 명확화
  - 가이드 vs 로컬 분석 문서 분류
  - 필요시 별도 디렉토리 생성

- [ ] Phase 3 실행: 동기화 스크립트 작성/검증

### Long-term (향후)

- [ ] 새 프로젝트 생성 시 템플릿 정합성 자동 검증
- [ ] GitHub Wiki에 `.moai` 디렉토리 구조 문서화
- [ ] 템플릿 변경 시 로컬 프로젝트 자동 감지 및 동기화

---

## 📚 참고 문서

| 문서 | 위치 | 용도 |
|------|------|------|
| 프로젝트 CLAUDE.md | `/CLAUDE.md` | 로컬 개발 가이드 |
| 템플릿 CLAUDE.md | `src/moai_adk/templates/CLAUDE.md` | 새 프로젝트용 |
| 개발 가이드 | `.moai/memory/development-guide.md` | Alfred 핵심 지침 |
| SPEC 메타 | `.moai/memory/spec-metadata.md` | SPEC 표준 |
| GitFlow 정책 | `.moai/memory/gitflow-protection-policy.md` | 팀 협업 정책 |

---

## 🎯 결론

MoAI-ADK의 `.moai` 디렉토리는 다음과 같이 구조화되어야 합니다:

1. **템플릿 (src/moai_adk/templates/.moai/)**:
   - 새 프로젝트를 위한 최소한의 핵심 파일
   - 모든 필수 가이드 포함
   - 파일명 대소문자 일관성 유지

2. **로컬 프로젝트 (.moai/)**:
   - 템플릿 파일의 사본 (동기화 관리)
   - 프로젝트 고유 분석/탐색 문서
   - 로컬 개발 메모리 파일

3. **동기화 프로세스**:
   - 패키지 업그레이드 후 자동 동기화
   - 파일명 일관성 검증
   - 로컬 파일 보호

이를 통해 새 프로젝트 생성 시 일관되고 완전한 기본 구조를 제공할 수 있습니다.
