# Claude Code 실전 예제 모음

> **실제 동작하는 예제 코드**
>
> 복사하여 바로 사용 가능한 검증된 예제

---

## 📋 목차

1. [Agent 예제](#1-agent-예제)
2. [Command 예제](#2-command-예제)
3. [Skill 예제](#3-skill-예제)
4. [Plugin 예제](#4-plugin-예제)
5. [Settings 예제](#5-settings-예제)

---

## 1. Agent 예제

### 예제 1: spec-builder (MoAI-ADK)

**파일**: `.claude/agents/alfred/spec-builder.md`

```markdown
---
name: spec-builder
description: "Use when: SPEC 작성, EARS 명세, 요구사항 분석이 필요할 때"
tools: Read, Write, Edit, MultiEdit, Grep, Glob, Bash, TodoWrite
model: sonnet
---

# SPEC Builder 🏗️ - 시스템 아키텍트

**MoAI-ADK SPEC 작성 전문가**

## 🎭 에이전트 페르소나

**아이콘**: 🏗️
**직무**: 시스템 아키텍트
**전문 영역**: SPEC 작성, EARS 명세, 요구사항 분석
**역할**: 비즈니스 요구사항을 체계적인 SPEC으로 변환
**목표**: 명확하고 테스트 가능한 SPEC 문서 작성

## 🎯 핵심 역할

### 1. SPEC 문서 작성
- EARS 5가지 구문 적용
- YAML Front Matter 7개 필수 필드
- HISTORY 섹션 관리

### 2. 자동 실행 조건
- `/alfred:1-spec` 커맨드 실행 시
- 새로운 기능 요구사항 발생 시
- 기존 SPEC 개선 요청 시

## 📐 워크플로우

### STEP 1: 프로젝트 문서 분석
\`\`\`bash
# product.md 읽기
Read .moai/project/product.md

# 기존 SPEC 확인
ls .moai/specs/SPEC-*/spec.md
\`\`\`

### STEP 2: SPEC 초안 작성
\`\`\`bash
Write .moai/specs/SPEC-{ID}/spec.md
\`\`\`

## 💡 사용 가이드
\`\`\`bash
@agent-spec-builder "사용자 인증 기능 SPEC 작성"
\`\`\`
```

### 예제 2: code-analyzer (간단한 분석 에이전트)

**파일**: `.claude/agents/code-analyzer.md`

```markdown
---
name: code-analyzer
description: Use when: code quality analysis is needed
tools: Read, Grep, Bash(rg:*)
model: haiku
---

# Code Analyzer 🔍 - 품질 관리자

코드 품질 지표를 분석하고 개선 제안을 제공합니다.

## 🎭 에이전트 페르소나

**아이콘**: 🔍
**직무**: 품질 관리자
**전문 영역**: 코드 품질 분석, 린팅, 테스트 커버리지
**역할**: 코드 품질 지표 측정 및 개선 제안
**목표**: 높은 코드 품질 유지

## 🎯 핵심 역할

### 1. 품질 지표 수집
- 테스트 커버리지 확인
- 린팅 오류 검출
- 복잡도 분석

### 2. 자동 실행 조건
- 커밋 전 품질 검증 요청 시
- PR 생성 전 자동 검증

## 📐 워크플로우

### STEP 1: 코드 스캔
\`\`\`bash
rg "@TEST:" -n tests/
pytest --cov
\`\`\`

### STEP 2: 보고서 생성
- 커버리지: {percentage}%
- 린팅 오류: {count}개

## 💡 사용 가이드
\`\`\`bash
@agent-code-analyzer "현재 프로젝트 품질 분석"
\`\`\`
```

---

## 2. Command 예제

### 예제 1: /alfred:1-spec (MoAI-ADK)

**파일**: `.claude/commands/alfred/1-spec.md`

```markdown
---
name: alfred:1-spec
description: 프로젝트 문서 분석 및 SPEC 작성 (EARS 방식)
argument-hint: [feature-description]
allowed-tools:
  - Read
  - Write
  - Edit
  - Task
  - Grep
  - Glob
  - TodoWrite
  - Bash(git:*)
---

# 📋 MoAI-ADK 1단계: SPEC 작성

## 🎯 커맨드 목적

프로젝트 문서를 분석하여 EARS 방식으로 SPEC 문서를 작성하고 Git 브랜치를 생성합니다.

## 💡 사용 예시
\`\`\`bash
/alfred:1-spec "사용자 인증 기능"
\`\`\`

## 📋 실행 흐름

1. **Phase 1**: 프로젝트 문서 분석
   - product.md 읽기
   - 기존 SPEC 확인

2. **Phase 2**: SPEC 작성
   - EARS 방식 명세
   - Git 브랜치 생성

## 🔗 연관 에이전트

- **Primary**: spec-builder (🏗️ 시스템 아키텍트)
- **Secondary**: git-manager (🚀 릴리스 엔지니어)

## ⚠️ 주의사항

- SPEC ID 중복 확인 필수
- EARS 5가지 구문 준수

## 📋 다음 단계

- `/alfred:2-build SPEC-{ID}` - TDD 구현 시작
```

### 예제 2: /deploy-api (배포 커맨드)

**파일**: `.claude/commands/deploy-api.md`

```markdown
---
name: deploy-api
description: API 서버를 프로덕션 환경에 배포
argument-hint: [environment]
allowed-tools:
  - Read
  - Bash(git:*)
  - Bash(npm:*)
---

# 🚀 API 배포 커맨드

API 서버를 지정된 환경에 배포합니다.

## 🎯 커맨드 목적

Git 태그 생성 및 환경별 배포 자동화

## 💡 사용 예시
\`\`\`bash
/deploy-api production
/deploy-api staging
\`\`\`

## 📋 실행 흐름

1. **Phase 1**: Git 상태 확인
   - 현재 브랜치 확인 (main 필수)
   - 버전 태그 생성

2. **Phase 2**: 배포 실행
   - npm run build
   - 환경별 배포 스크립트 실행

## ⚠️ 주의사항

- main 브랜치에서만 실행 가능
- 모든 테스트 통과 필수
```

---

## 3. Skill 예제

### 예제 1: moai-alfred-tag-scanning

**파일**: `.claude/skills/moai-alfred-tag-scanning/SKILL.md`

```markdown
---
name: moai-alfred-tag-scanning
description: TAG 마커 직접 스캔 및 인벤토리 생성 (CODE-FIRST 원칙)
model: haiku
allowed-tools:
  - Grep
  - Read
---

# TAG 스캐너

> CODE-FIRST 원칙: 중간 캐시 없이 코드를 직접 스캔

## 🎯 목적

`@SPEC`, `@TEST`, `@CODE`, `@DOC` TAG를 코드에서 직접 스캔합니다.

## 💡 사용법

"AUTH 도메인 TAG 목록 조회"

## 📋 스캔 방법

\`\`\`bash
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
\`\`\`

## ✅ 검증

- 모든 `@CODE` TAG는 대응하는 `@SPEC`이 있는가?
- 고아 TAG 없음
```

### 예제 2: moai-alfred-feature-selector

**파일**: `.claude/skills/moai-alfred-feature-selector/SKILL.md`

```markdown
---
name: moai-alfred-feature-selector
description: 프로젝트 유형별 최적 기능 선택 (37개 스킬 → 3~5개 자동 필터링)
model: haiku
allowed-tools:
  - Read
---

# MoAI Alfred Feature Selector

> 프로젝트 특성에 맞는 MoAI-ADK 기능 자동 선택

## 🎯 목적

프로젝트 유형을 분석하여 필요한 기능만 선택합니다.

## 📋 프로젝트 분류

### 언어별
- **Python**: pytest, mypy, ruff
- **TypeScript**: Vitest, Biome

### 도메인별
- **CLI Tool**: 인자 파싱, POSIX 준수
- **Web API**: REST/GraphQL, 인증

## 💡 사용법

"/alfred:0-project 실행 시 자동 호출"
```

---

## 4. Plugin 예제

### 예제 1: GitHub + Filesystem (기본)

**파일**: `.claude/settings.json` (mcpServers 섹션)

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    },
    "filesystem": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-filesystem",
        "${CLAUDE_PROJECT_DIR}/src",
        "${CLAUDE_PROJECT_DIR}/tests"
      ]
    }
  }
}
```

### 예제 2: MoAI-ADK 완전 구성

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    },
    "moai-filesystem": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-filesystem",
        "${CLAUDE_PROJECT_DIR}/.moai",
        "${CLAUDE_PROJECT_DIR}/src",
        "${CLAUDE_PROJECT_DIR}/tests",
        "${CLAUDE_PROJECT_DIR}/docs"
      ]
    },
    "brave-search": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-brave-search"],
      "env": {
        "BRAVE_API_KEY": "${BRAVE_API_KEY}"
      }
    }
  }
}
```

---

## 5. Settings 예제

### 예제 1: Python 프로젝트

**파일**: `.claude/settings.json`

```json
{
  "permissions": {
    "defaultMode": "default",
    "allow": [
      "Read",
      "Write",
      "Edit",
      "Grep",
      "Glob",
      "Bash(git:*)",
      "Bash(python:*)",
      "Bash(pytest:*)",
      "Bash(mypy:*)",
      "Bash(ruff:*)"
    ],
    "ask": [
      "Bash(pip install:*)",
      "Bash(git push:*)"
    ],
    "deny": [
      "Read(./.env)",
      "Bash(sudo:*)",
      "Bash(rm -rf:*)"
    ]
  }
}
```

### 예제 2: TypeScript 프로젝트

```json
{
  "permissions": {
    "defaultMode": "default",
    "allow": [
      "Read",
      "Write",
      "Edit",
      "Grep",
      "Glob",
      "Bash(git:*)",
      "Bash(node:*)",
      "Bash(npm:*)",
      "Bash(pnpm:*)"
    ],
    "ask": [
      "Bash(npm install:*)",
      "Bash(pnpm install:*)",
      "Bash(git push:*)"
    ],
    "deny": [
      "Read(./.env)",
      "Read(./secrets/**)",
      "Bash(sudo:*)",
      "Bash(rm -rf:*)"
    ]
  }
}
```

### 예제 3: MoAI-ADK 프로젝트 (훅 포함)

```json
{
  "permissions": {
    "defaultMode": "default",
    "allow": [
      "Task",
      "Read",
      "Write",
      "Edit",
      "MultiEdit",
      "Grep",
      "Glob",
      "TodoWrite",
      "Bash(git:*)",
      "Bash(python:*)",
      "Bash(pytest:*)",
      "Bash(mypy:*)",
      "Bash(ruff:*)",
      "Bash(moai-adk:*)",
      "Bash(alfred:*)"
    ],
    "ask": [
      "Bash(git push:*)",
      "Bash(pip install:*)"
    ],
    "deny": [
      "Read(./.env)",
      "Read(./secrets/**)",
      "Bash(sudo:*)",
      "Bash(rm -rf:*)"
    ]
  },
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "command": "node $CLAUDE_PROJECT_DIR/.claude/hooks/alfred/session-notice.cjs",
            "type": "command"
          }
        ],
        "matcher": "*"
      }
    ],
    "PreToolUse": [
      {
        "hooks": [
          {
            "command": "node $CLAUDE_PROJECT_DIR/.claude/hooks/alfred/pre-write-guard.cjs",
            "type": "command"
          }
        ],
        "matcher": "Edit|Write|MultiEdit"
      }
    ]
  }
}
```

---

**최종 업데이트**: 2025-10-19
**작성자**: @Alfred
