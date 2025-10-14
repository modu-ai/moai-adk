---
id: ALFRED-CMD-001
version: 0.1.0
status: completed
created: 2025-10-14
updated: 2025-10-14
author: @Goos
priority: high
category: refactor
labels:
  - alfred
  - commands
  - naming
  - consistency
blocks:
  - DOCS-004
  - DOCS-005
scope:
  files:
    - templates/.claude/commands/alfred/0-project.md
    - .claude/commands/alfred/8-project.md
    - CLAUDE.md
    - docs/**/*.md
---

# @SPEC:ALFRED-CMD-001: Alfred 커맨드 명명 통일

## HISTORY
### v0.1.0 (2025-10-14)
- **CHANGED**: 구현 완료 - 템플릿 본문 "8단계" → "0단계" 수정
- **AUTHOR**: @Goos
- **RELATED**: 커밋 0320ee9 (REFACTOR: Alfred 커맨드 템플릿 명명 통일)

### v0.0.1 (2025-10-14)
- **INITIAL**: /alfred:0-project vs /alfred:8-project 명명 통일
- **AUTHOR**: @Goos
- **DECISION**: 0-project로 통일 (직관성, 문서 일관성)
- **REASON**: 사용자 혼란 방지, 직관적인 단계 표현 (0단계 = 프로젝트 초기화)

---

## 개요

Alfred 커맨드 명명 불일치를 해결하여 사용자 혼란을 방지하고 직관적인 워크플로우를 제공합니다.

**현재 문제**:
- templates/.claude/commands/alfred/0-project.md (파일명)
- YAML frontmatter: `name: alfred:8-project` (내용)
- 실제 프로젝트: `.claude/commands/alfred/8-project.md`
- 문서: `/alfred:0-project` 사용 (CLAUDE.md, README.md, docs/)

**결정**:
- `/alfred:0-project`로 통일 (직관적, 문서 일관성)

---

## Ubiquitous Requirements (기본 요구사항)

시스템은 다음 기능을 제공해야 한다:
1. templates/.claude/commands/alfred/0-project.md YAML frontmatter 수정 (`alfred:8-project` → `alfred:0-project`)
2. .claude/commands/alfred/8-project.md 파일명 변경 (`8-project.md` → `0-project.md`)
3. .claude/commands/alfred/0-project.md YAML frontmatter 수정 (`alfred:8-project` → `alfred:0-project`)
4. 모든 문서에서 `/alfred:0-project` 일관성 확인

---

## Event-driven Requirements (이벤트 기반)

- WHEN templates 파일을 수정하면, 시스템은 YAML frontmatter의 `name` 필드를 `alfred:0-project`로 변경해야 한다
- WHEN .claude/commands/alfred/ 파일명을 변경하면, 시스템은 `8-project.md`를 `0-project.md`로 변경해야 한다
- WHEN YAML frontmatter를 수정하면, 시스템은 `name: alfred:0-project`로 변경해야 한다

---

## State-driven Requirements (상태 기반)

- WHILE 파일명을 변경하는 동안, 시스템은 파일 내용을 유지해야 한다
- WHILE YAML frontmatter를 수정하는 동안, 시스템은 다른 필드를 변경하지 않아야 한다

---

## Constraints (제약사항)

- 파일명 변경은 Git으로 추적 가능해야 한다 (`git mv` 사용)
- IF 파일명을 변경하면, 시스템은 기존 파일을 삭제하지 않아야 한다 (Git이 자동 처리)
- 모든 문서에서 `/alfred:0-project` 일관성을 유지해야 한다

---

## 옵션 비교

### 옵션 A: 0-project로 통일 (권장) ✅

**장점**:
- ✅ 직관적 (0단계 = 프로젝트 초기화)
- ✅ 문서 수정 최소화 (CLAUDE.md, README.md, docs/ 이미 /alfred:0-project 사용)
- ✅ 워크플로우 단계 명확 (0 → 1 → 2 → 3)
- ✅ CLAUDE.md와 일치

**단점**:
- ⚠️ 기존 프로젝트 사용자 혼란 (templates에서 8-project 사용 중)

**작업 범위**:
1. templates/.claude/commands/alfred/0-project.md YAML frontmatter 수정
2. .claude/commands/alfred/8-project.md → 0-project.md 파일명 변경
3. .claude/commands/alfred/0-project.md YAML frontmatter 수정

### 옵션 B: 8-project로 통일 (비권장) ❌

**장점**:
- ✅ templates 파일명 변경 불필요

**단점**:
- ❌ 직관성 감소 (8단계가 프로젝트 초기화?)
- ❌ 대규모 문서 수정 필요 (CLAUDE.md, README.md, docs/ 모두 변경)
- ❌ 워크플로우 단계 혼란 (8 → 1 → 2 → 3)

**작업 범위**:
1. CLAUDE.md 전체 수정 (/alfred:0-project → /alfred:8-project)
2. README.md 전체 수정
3. docs/ 17개 파일 수정

---

## 결정: 옵션 A (0-project로 통일)

**근거**:
1. **직관성**: 0단계 = 프로젝트 초기화 (자연스러운 시작점)
2. **문서 일관성**: CLAUDE.md, README.md, docs/ 이미 /alfred:0-project 사용
3. **작업 범위**: 3개 파일 수정 vs 20개 이상 파일 수정
4. **워크플로우 명확성**: 0 → 1 → 2 → 3 단계적 진행

---

## 구현 계획

### 1단계: templates 파일 수정 (1분)
**파일**: `/Users/goos/MoAI/MoAI-ADK/templates/.claude/commands/alfred/0-project.md`

**수정 내용**:
```yaml
---
name: alfred:0-project  # 기존: alfred:8-project
description: Use PROACTIVELY for 프로젝트 문서 초기화 - product/structure/tech.md 생성 및 언어별 최적화 설정
allowed-tools:
  - Read
  - Write
  - Edit
  - MultiEdit
  - Grep
  - Glob
  - TodoWrite
  - Bash(ls:*)
  - Bash(find:*)
  - Bash(cat:*)
  - Task
---
```

### 2단계: .claude/commands/alfred/ 파일명 변경 (1분)
**기존 파일**: `/Users/goos/MoAI/MoAI-ADK/.claude/commands/alfred/8-project.md`
**새 파일**: `/Users/goos/MoAI/MoAI-ADK/.claude/commands/alfred/0-project.md`

**명령어**:
```bash
cd /Users/goos/MoAI/MoAI-ADK/.claude/commands/alfred/
mv 8-project.md 0-project.md
```

### 3단계: .claude/commands/alfred/0-project.md YAML frontmatter 수정 (1분)
**파일**: `/Users/goos/MoAI/MoAI-ADK/.claude/commands/alfred/0-project.md`

**수정 내용**:
```yaml
---
name: alfred:0-project  # 기존: alfred:8-project
description: Use PROACTIVELY for 프로젝트 문서 초기화 - product/structure/tech.md 생성 및 언어별 최적화 설정
allowed-tools:
  - Read
  - Write
  - Edit
  - MultiEdit
  - Grep
  - Glob
  - TodoWrite
  - Bash(ls:*)
  - Bash(find:*)
  - Bash(cat:*)
  - Task
---
```

### 4단계: 문서 일관성 확인 (2분)
**검증 명령어**:
```bash
# CLAUDE.md 확인
rg "/alfred:0-project" CLAUDE.md -n

# README.md 확인
rg "/alfred:0-project" README.md -n

# docs/ 확인
rg "/alfred:0-project" docs/ -n

# 8-project 흔적 확인 (없어야 함)
rg "/alfred:8-project" . -n
```

---

## 검증 기준

### 필수 검증 항목
1. ✅ templates/.claude/commands/alfred/0-project.md YAML frontmatter: `name: alfred:0-project`
2. ✅ .claude/commands/alfred/0-project.md 파일 존재 확인
3. ✅ .claude/commands/alfred/8-project.md 파일 부재 확인
4. ✅ .claude/commands/alfred/0-project.md YAML frontmatter: `name: alfred:0-project`
5. ✅ 모든 문서에서 `/alfred:0-project` 일관성 확인

### 선택 검증 항목
1. ⚠️ Git 히스토리에 파일명 변경 기록 확인 (`git log --follow`)
2. ⚠️ 다른 Alfred 커맨드 파일 확인 (1-spec, 2-build, 3-sync, 9-update)

---

## 영향 분석

### 긍정적 영향
1. ✅ 사용자 혼란 감소 (명확한 0단계)
2. ✅ 문서 일관성 향상 (모든 문서에서 /alfred:0-project 사용)
3. ✅ 워크플로우 직관성 향상 (0 → 1 → 2 → 3)

### 부정적 영향
1. ⚠️ 기존 프로젝트 사용자 혼란 (templates에서 8-project 사용 중)
2. ⚠️ 캐시된 Claude Code 설정 업데이트 필요

### 완화 방안
1. CHANGELOG.md에 명명 변경 기록
2. 마이그레이션 가이드 제공
3. v0.4.0 릴리스 노트에 명시

---

## 롤백 계획

만약 0-project로 통일이 문제를 일으킨다면:

### 롤백 단계 (5분)
1. templates/.claude/commands/alfred/0-project.md YAML frontmatter: `name: alfred:8-project`
2. .claude/commands/alfred/0-project.md → 8-project.md 파일명 변경
3. .claude/commands/alfred/8-project.md YAML frontmatter: `name: alfred:8-project`

### 롤백 트리거
- 사용자 피드백: 명명 변경으로 인한 치명적 혼란
- CI/CD 실패: 파일명 변경으로 인한 빌드 오류
- 테스트 실패: Alfred 커맨드 인식 실패

---

## 다음 단계

1. ✅ SPEC-ALFRED-CMD-001 승인
2. ✅ templates/.claude/commands/alfred/0-project.md 수정
3. ✅ .claude/commands/alfred/8-project.md → 0-project.md 파일명 변경
4. ✅ .claude/commands/alfred/0-project.md YAML frontmatter 수정
5. ✅ 문서 일관성 검증
6. ✅ Git 커밋: "🔧 REFACTOR: Alfred 커맨드 명명 통일 (8-project → 0-project)"

---

## 관련 SPEC

- **SPEC-DOCS-004**: README.md Python v0.3.0 업데이트 (blocked by ALFRED-CMD-001)
- **SPEC-DOCS-005**: 온라인 문서 v0.3.0 정합성 확보 (blocked by ALFRED-CMD-001)

---

## 참고 문서

- `CLAUDE.md`: Alfred 커맨드 전체 목록
- `templates/CLAUDE.md`: 템플릿 버전 CLAUDE.md
- `.claude/commands/alfred/`: Alfred 커맨드 파일 디렉토리
- `docs/guides/alfred-superagent.md`: Alfred 사용법 가이드
