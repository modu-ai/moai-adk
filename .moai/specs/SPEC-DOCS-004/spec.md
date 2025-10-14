---
id: DOCS-004
version: 0.0.1
status: draft
created: 2025-10-14
updated: 2025-10-14
author: @Goos
priority: critical
category: docs
labels:
  - readme
  - python-3.13
  - migration
  - v0.3.0
blocks:
  - DOCS-005
scope:
  files:
    - README.md
---

# @SPEC:DOCS-004: README.md Python v0.3.0 업데이트

## HISTORY
### v0.0.1 (2025-10-14)
- **INITIAL**: README.md를 Python v0.3.0에 맞게 재작성
- **AUTHOR**: @Goos
- **REASON**: TypeScript v0.2.x 내용 제거, Python 기준 재작성, CLI 미구현 기능 "Coming Soon" 배지 추가

---

## 개요

README.md 파일을 Python v0.3.0 기준으로 완전히 재작성하여 사용자 혼란을 방지하고 정확한 정보를 제공합니다.

**핵심 문제**:
- 현재 README.md에 TypeScript v0.2.x 내용 포함 (Bun, npm, Node.js)
- CLI 명령어 상세 설명 (실제로는 모두 미구현)
- 프로그래매틱 API 예제 (TypeScript 코드)
- 설치 방법이 TypeScript 기준

**목표**:
- Python v0.3.0 실제 상태 정확 반영
- CLI 미구현 기능 명확히 표시 ("🚧 Coming in v0.4.0" 배지)
- Alfred 커맨드 중심 워크플로우 강조
- Quick Start를 Python 기준으로 재작성

---

## Ubiquitous Requirements (기본 요구사항)

시스템은 다음 기능을 제공해야 한다:
1. README.md 파일을 Python v0.3.0에 맞게 재작성
2. TypeScript 관련 내용 완전 제거
3. CLI 미구현 기능 "🚧 Coming in v0.4.0" 배지 추가
4. Quick Start 섹션을 Python 기준으로 재작성
5. Alfred 커맨드 사용법 강조

---

## Event-driven Requirements (이벤트 기반)

- WHEN TypeScript 관련 섹션을 발견하면, 시스템은 해당 섹션을 제거해야 한다
- WHEN CLI 명령어 설명을 발견하면, 시스템은 "🚧 Coming in v0.4.0" 배지를 추가해야 한다
- WHEN 설치 방법을 작성하면, 시스템은 uv/pip 기준으로 작성해야 한다

---

## State-driven Requirements (상태 기반)

- WHILE README.md를 재작성하는 동안, 시스템은 기존 섹션 구조를 최대한 유지해야 한다
- WHILE 미구현 기능을 표시하는 동안, 시스템은 사용자에게 예상 릴리스 버전을 명시해야 한다

---

## Constraints (제약사항)

- README.md 파일은 7000 줄을 초과하지 않아야 한다
- IF 기존 섹션을 제거하면, 시스템은 제거 이유를 SPEC 문서에 기록해야 한다
- 모든 코드 예제는 Python 3.13+ 문법을 따라야 한다

---

## 제거할 섹션 (TypeScript v0.2.x 관련)

### 1. 시스템 요구사항
```markdown
### 🔴 필수 요구사항
- **Node.js**: 18.0 이상
- **npm**: 8.0.0 이상 (또는 **Bun 1.2.0 이상 강력 추천**)
```

**제거 이유**: Python 프로젝트는 Node.js/npm/Bun이 필요하지 않음

### 2. 설치 옵션 (Option A/B/C)
```markdown
### Option A: Bun 설치 (최적 성능, 강력 추천) 🔥
bun add -g moai-adk

### Option B: npm 설치 (표준 옵션)
npm install -g moai-adk

### Option C: 개발자 설치 (로컬 개발용)
cd moai-adk/moai-adk-ts
bun install
```

**제거 이유**: Python 프로젝트는 uv 또는 pip로 설치

### 3. CLI 명령어 상세 설명
```markdown
### moai init [project]
### moai doctor
### moai status
### moai restore <backup-path>
```

**제거 이유**: CLI 명령어 모두 미구현 (v0.3.0 기준)

### 4. 템플릿 업데이트: 2가지 방법
```markdown
#### 🔧 방법 1: `moai init .` (CLI 직접 실행)
#### 🤖 방법 2: `/alfred:9-update` (Claude Code 전용)
```

**제거 이유**: `moai init .` 미구현, `/alfred:9-update`는 유지

### 5. 프로그래매틱 API
```typescript
import { CLIApp, SystemChecker, TemplateManager } from 'moai-adk';
```

**제거 이유**: Python API 미구현

### 6. 개발 환경 설정
```bash
cd moai-adk/moai-adk-ts
bun install
bun run dev
bun run build
bun test
bun run check
```

**제거 이유**: TypeScript 빌드 환경

---

## 유지할 섹션

### 1. Meet Alfred
- Alfred 소개
- 10개 AI 에이전트 팀
- 4가지 핵심 가치 (일관성, 품질, 추적성, 범용성)

### 2. The Problem
- 바이브 코딩의 한계
- 5가지 문제점 (아름답지만 작동하지 않는 코드, 플랑켄슈타인 코드, 디버깅 지옥, 요구사항 추적성 상실, 팀 협업 붕괴)

### 3. The Solution
- 3단계 워크플로우 (SPEC → BUILD → SYNC)
- TDD 사이클 (RED → GREEN → REFACTOR)

### 4. How Alfred Works
- 10개 AI 에이전트 팀 구조
- 협업 원칙

### 5. Output Styles
- 4가지 대화 스타일 (Professional, Beginner, Pair Collaboration, Study Deep)

### 6. SPEC 메타데이터 구조
- 필수 필드 7개
- 선택 필드 9개

### 7. @TAG 시스템
- TAG 체계 철학
- TAG 사용 규칙
- 언어별 TAG 사용 예시

### 8. Universal Language Support
- 웹/백엔드 언어
- 모바일 언어
- 자동 언어 감지

### 9. TRUST 5원칙
- Test First, Readable, Unified, Secured, Trackable

### 10. 문제 해결
- 자주 발생하는 문제 (7가지)
- 로그 확인
- 긴급 복구

### 11. 문서 및 지원
- GitHub Issues
- GitHub Discussions
- npm Package (→ PyPI Package로 변경)

---

## 추가할 섹션

### 1. Python 3.13+ 설치 방법
```markdown
## 설치

### Option A: uv 설치 (권장)
```bash
# uv 설치
curl -LsSf https://astral.sh/uv/install.sh | sh

# MoAI-ADK 설치
uv tool install moai-adk
```

### Option B: pip 설치
```bash
pip install moai-adk
```

### 설치 확인
```bash
moai --version
# 출력: v0.3.0
```
```

### 2. Quick Start (Python 기준)
```markdown
## Quick Start (3분 실전)

### 1단계: 설치 (30초)
```bash
# uv 권장
uv tool install moai-adk

# 또는 pip
pip install moai-adk

# 설치 확인
moai --version  # v0.3.0
```

### 2단계: Claude Code에서 Alfred 사용 (1분)
```text
# Claude Code에서 실행
/alfred:0-project

# Alfred가 자동으로 수행:
# - .moai/project/ 문서 3종 생성 (product/structure/tech.md)
# - 언어별 최적 도구 체인 설정
# - 프로젝트 컨텍스트 완벽 이해
```

### 3단계: 첫 기능 개발 (1분 30초)
```text
# SPEC 작성
/alfred:1-spec "JWT 기반 사용자 로그인 API"

# TDD 구현
/alfred:2-build AUTH-001

# 문서 동기화
/alfred:3-sync
```
```

### 3. Coming Soon (v0.4.0)
```markdown
## Coming Soon (v0.4.0)

다음 기능들이 v0.4.0에 추가될 예정입니다:

### CLI 명령어
- 🚧 `moai init .` - 프로젝트 초기화
- 🚧 `moai doctor` - 시스템 진단
- 🚧 `moai status` - 프로젝트 상태 확인
- 🚧 `moai restore <backup-path>` - 백업 복원

### Python API
- 🚧 프로그래매틱 API
- 🚧 템플릿 매니저
- 🚧 Git 통합

**현재 (v0.3.0)**: Alfred 커맨드로 모든 기능 사용 가능
```

---

## 새로운 README.md 구조

```markdown
# MoAI-ADK v0.3.0

## Meet Alfred
## Quick Start (Python 기준)
  - uv/pip 설치
  - Claude Code에서 Alfred 사용
  - /alfred:0-project, 1-spec, 2-build, 3-sync
## The Problem
## The Solution
## How Alfred Works (10개 에이전트)
## Output Styles
## Language Support
## SPEC & TAG System
## TRUST 5원칙
## Coming Soon (v0.4.0)
  - CLI 명령어 (moai init, doctor, status)
  - Python API
## 문제 해결
## 문서 및 지원
```

---

## 검증 기준

### 필수 검증 항목
1. ✅ TypeScript 관련 내용 완전 제거 (Node.js, Bun, npm)
2. ✅ Python 3.13+ 설치 방법 추가 (uv, pip)
3. ✅ CLI 미구현 기능 "🚧 Coming in v0.4.0" 배지 추가
4. ✅ Quick Start를 Python 기준으로 재작성
5. ✅ Alfred 커맨드 사용법 강조
6. ✅ PyPI Package 링크로 변경

### 선택 검증 항목
1. ⚠️ 코드 예제가 Python 3.13+ 문법을 따르는지 확인
2. ⚠️ README.md 길이가 7000 줄 이하인지 확인
3. ⚠️ 모든 링크가 유효한지 확인

---

## 구현 전략

### 1단계: 제거 작업 (5분)
- TypeScript 관련 섹션 완전 제거
- CLI 명령어 상세 설명 제거
- 프로그래매틱 API 제거
- 개발 환경 설정 제거

### 2단계: 추가 작업 (5분)
- Python 3.13+ 설치 방법 추가
- Quick Start 재작성
- Coming Soon 섹션 추가

### 3단계: 수정 작업 (5분)
- npm Package → PyPI Package 링크 변경
- 모든 코드 예제 Python 3.13+ 검증
- Alfred 커맨드 사용법 강조

### 4단계: 검증 (3분)
- 전체 README.md 읽으면서 TypeScript 흔적 확인
- CLI 명령어 "🚧" 배지 확인
- 링크 유효성 확인

---

## 관련 SPEC

- **SPEC-DOCS-005**: 온라인 문서 v0.3.0 정합성 확보 (depends on DOCS-004)
- **SPEC-ALFRED-CMD-001**: Alfred 커맨드 명명 통일 (blocks DOCS-004)

---

## 참고 문서

- `.moai/memory/spec-metadata.md`: SPEC 메타데이터 표준
- `CLAUDE.md`: Alfred 커맨드 및 에이전트 설명
- `pyproject.toml`: Python v0.3.0 실제 의존성 목록
