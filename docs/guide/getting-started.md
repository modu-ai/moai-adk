# @CODE:DOCS-001:UI | SPEC: .moai/specs/SPEC-DOCS-001/spec.md

# Quick Start

3분 안에 MoAI-ADK를 시작하고 첫 번째 SPEC을 작성해보세요.

## 📋 준비물

시작하기 전에 다음 도구들이 설치되어 있는지 확인하세요:

- ✅ **Bun 또는 npm** 설치됨
- ✅ **Claude Code** 실행 중
- ✅ **Git** 설치됨 (필수) - Personal/Team 모드 공통 필수 요구사항

## ⚡ 3단계로 시작하기

### 1️⃣ 설치 (30초)

:::code-group

```bash [Bun (권장)]
# Bun 설치 (5배 빠른 성능)
curl -fsSL https://bun.sh/install | bash

# MoAI-ADK 설치
bun add -g moai-adk

# 설치 확인
moai --version
# 출력: v0.x.x
```

```bash [npm]
# npm으로 설치
npm install -g moai-adk

# 설치 확인
moai --version
# 출력: v0.x.x
```

:::

### 2️⃣ 초기화 (1분)

**터미널에서:**

```bash
# 새 프로젝트 생성
moai init my-project
cd my-project

# 또는 기존 프로젝트에 설치
cd existing-project
moai init .

# Claude Code 실행
claude
```

**Claude Code에서** (필수):

```text
/alfred:8-project
```

Alfred가 자동으로 수행하는 작업 (v2.0.0):

- **프로젝트 문서 3종 자동 생성**:
  - `.moai/project/product.md` - 제품 비전, 타겟 유저, 핵심 기능
  - `.moai/project/structure.md` - 아키텍처, 모듈 구조, 디렉토리 레이아웃
  - `.moai/project/tech.md` - 기술 스택, 개발 도구, 의존성 관리

- **Alfred 브랜딩 경로 자동 감지**:
  - `.claude/alfred/` 디렉토리 구조 생성
  - Claude Code 명령어 파일 최적화 배치

- **언어별 최적 도구 체인 자동 설정**:
  - TypeScript → Vitest + Biome
  - Python → pytest + ruff
  - Go → go test + golint
  - Flutter → flutter test + dart analyze

- **프로젝트 컨텍스트 완벽 이해**:
  - 프로젝트 메타데이터 v2.0.0 구조로 저장
  - MoAI-ADK 철학 (`constitution`, `git_strategy`, `tags`, `pipeline`) 반영

### 3️⃣ 첫 기능 개발 (1분 30초)

**Claude Code에서 3단계 워크플로우 실행:**

```text
# 1. SPEC 작성
/alfred:1-spec "JWT 기반 사용자 로그인 API"

# 2. TDD 구현
/alfred:2-build AUTH-001

# 3. 문서 동기화
/alfred:3-sync
```

축하합니다! 🎉 첫 번째 기능을 SPEC-First TDD 방식으로 완성했습니다.

## 🎯 다음 단계

- [MoAI-ADK란?](/guide/what-is-moai-adk) - MoAI-ADK가 해결하는 문제 이해하기
- [SPEC 우선 TDD](/concepts/spec-first-tdd) - SPEC-First TDD 철학 배우기
- [FAQ](/guide/faq) - 자주 묻는 질문 확인하기

## ❓ 문제가 발생하나요?

- [GitHub Issues](https://github.com/modu-ai/moai-adk/issues) - 버그 리포트
- [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions) - 질문 및 토론
