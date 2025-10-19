# Claude Code 컴포넌트 작성 가이드

> **5가지 컴포넌트의 완벽한 작성 가이드**
>
> Agent, Command, Skill, Plugin, Settings

---

## 📋 목차

1. [Agent 작성 가이드](#1-agent-작성-가이드)
2. [Command 작성 가이드](#2-command-작성-가이드)
3. [Skill 작성 가이드](#3-skill-작성-가이드)
4. [Plugin 설정 가이드](#4-plugin-설정-가이드)
5. [Settings 설정 가이드](#5-settings-설정-가이드)

---

## 1. Agent 작성 가이드

### 📐 파일 구조

**위치**: `.claude/agents/{agent-name}.md`

**YAML Frontmatter** (필수):
```yaml
---
name: {agent-name}              # kebab-case
description: "Use when: {trigger}"  # "Use when:" 패턴 필수
tools: Read, Write, Edit        # 필요한 도구만
model: sonnet                   # sonnet|haiku
---
```

### 🎭 에이전트 페르소나

**필수 요소**:
- **아이콘**: 시각적 식별자 (emoji)
- **직무**: IT 전문 직무 (System Architect, QA Lead 등)
- **전문 영역**: 구체적 전문 분야
- **역할**: 에이전트 책임
- **목표**: 달성하려는 목표

**예시**:
```markdown
## 🎭 에이전트 페르소나

**아이콘**: 🏗️
**직무**: 시스템 아키텍트 (System Architect)
**전문 영역**: SPEC 작성, EARS 명세, 요구사항 분석
**역할**: 비즈니스 요구사항을 체계적인 SPEC으로 변환
**목표**: 명확하고 테스트 가능한 SPEC 문서 작성
```

### ⚙️ 모델 선택 가이드

| 모델 | 사용 시점 | 예시 |
|------|----------|------|
| **sonnet** | 복잡한 판단, 설계, 창의성 | SPEC 작성, TDD 전략, 디버깅 |
| **haiku** | 빠른 처리, 패턴 기반 작업 | 문서 동기화, TAG 스캔, 린팅 |

### 🛠️ 도구 선택 가이드

| 작업 유형 | 필수 도구 |
|----------|----------|
| **분석** | Read, Grep, Glob |
| **문서 작성** | Read, Write, Edit |
| **코드 구현** | Read, Write, Edit, MultiEdit |
| **Git 작업** | Read, Bash(git:*) |
| **검증** | Read, Grep, Bash |

### ✅ 검증 체크리스트

- [ ] YAML frontmatter 존재
- [ ] `name`: kebab-case
- [ ] `description`: "Use when:" 패턴 포함
- [ ] `tools`: 필요한 도구만
- [ ] `model`: sonnet 또는 haiku
- [ ] 에이전트 페르소나 섹션 포함
- [ ] 워크플로우 구체적 단계 포함

---

## 2. Command 작성 가이드

### 📐 파일 구조

**위치**: `.claude/commands/{command-name}.md`

**YAML Frontmatter** (필수):
```yaml
---
name: {command-name}            # kebab-case
description: {한 줄 설명}        # 명확한 목적
argument-hint: [{param}]        # 선택적
allowed-tools:                  # 필요한 도구만
  - Read
  - Write
  - Task
---
```

### 🔧 명명 규칙

- **kebab-case** 사용
- **동사로 시작** (run, check, deploy, create)
- **명확하고 구체적**

**올바른 예시**:
- ✅ `deploy-production`
- ✅ `run-tests`
- ✅ `alfred:1-spec`

**잘못된 예시**:
- ❌ `doSomething` (camelCase)
- ❌ `cmd1` (불명확)

### 📋 표준 섹션 구조

```markdown
# {Command Title}

{Brief description}

## 🎯 커맨드 목적
{Detailed purpose}

## 💡 사용 예시
\`\`\`bash
/{command-name} {example-args}
\`\`\`

## 📋 실행 흐름
1. **Phase 1**: {Planning}
2. **Phase 2**: {Execution}

## 🔗 연관 에이전트
- **Primary**: {agent-name} - {role}

## ⚠️ 주의사항
- {Warning 1}

## 📋 다음 단계
- {Next step}
```

### ✅ 검증 체크리스트

- [ ] YAML frontmatter 존재
- [ ] `name`: kebab-case
- [ ] `description`: 한 줄 설명
- [ ] `allowed-tools`: 배열 형식
- [ ] Bash 도구 사용 시 구체적 패턴 (`Bash(git:*)`)
- [ ] 사용 예시 포함
- [ ] 실행 흐름 명시

---

## 3. Skill 작성 가이드

### 📐 파일 구조

**위치**: `.claude/skills/{skill-name}/SKILL.md`

**YAML Frontmatter** (필수):
```yaml
---
name: {skill-name}              # kebab-case
description: {한 줄 설명}        # 동사로 시작, 200자 이하
model: haiku                    # haiku|sonnet
allowed-tools:                  # 최소 권한
  - Read
  - Write
---
```

### 🎯 description 작성법

**중요**: Claude가 언제 스킬을 호출할지 결정하는 핵심 필드 (200자 이하)

**좋은 예시**:
- ✅ "TAG 마커 직접 스캔 및 인벤토리 생성 (CODE-FIRST 원칙)"
- ✅ "프로젝트 유형별 최적 기능 선택 (37개 스킬 → 3~5개 자동 필터링)"

**나쁜 예시**:
- ❌ "스킬입니다" (너무 모호)
- ❌ "This skill does something" (비구체적)

### 📁 선택적 파일 구조

```
.claude/skills/{skill-name}/
├── SKILL.md (required)
├── reference.md (optional)
├── examples.md (optional)
├── scripts/ (optional)
│   └── helper.py
└── templates/ (optional)
    └── template.txt
```

### ✅ 검증 체크리스트

- [ ] YAML frontmatter 존재
- [ ] `name`: kebab-case
- [ ] `description`: 200자 이하, 구체적
- [ ] `model`: haiku 또는 sonnet
- [ ] `allowed-tools`: 최소 권한 원칙
- [ ] 제목 (# {Skill Title}) 존재
- [ ] 목적 섹션 포함

---

## 4. Plugin 설정 가이드

### 📐 파일 구조

**위치**: `.claude/settings.json` (mcpServers 섹션)

**기본 구조**:
```json
{
  "mcpServers": {
    "{plugin-name}": {
      "command": "npx",
      "args": ["-y", "{plugin-package}"],
      "env": {
        "API_KEY": "${API_KEY}"
      }
    }
  }
}
```

### 🌟 추천 Plugin

| Plugin | 용도 | MoAI-ADK 활용 |
|--------|------|--------------|
| **@modelcontextprotocol/server-github** | GitHub API | PR/Issue 자동 생성 |
| **@modelcontextprotocol/server-filesystem** | 파일 시스템 | `.moai/` 안전 접근 |
| **@modelcontextprotocol/server-brave-search** | 웹 검색 | 기술 문서 참조 |

### 🔒 보안 원칙

#### 필수 체크리스트
- [ ] 환경변수 사용 (하드코딩 금지)
- [ ] 경로 제한 (Filesystem MCP)
- [ ] 최소 권한
- [ ] 민감 정보 차단 (`.env`, `secrets/`)
- [ ] 출처 신뢰성 (공식 또는 검증된 Plugin)

#### 안전한 설정
```json
{
  "mcpServers": {
    "github": {
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"  // ✅ 환경변수
      }
    },
    "filesystem": {
      "args": [
        "-y",
        "@modelcontextprotocol/server-filesystem",
        "${CLAUDE_PROJECT_DIR}/.moai",  // ✅ 제한된 경로
        "${CLAUDE_PROJECT_DIR}/src"
      ]
    }
  }
}
```

### ✅ 검증 체크리스트

- [ ] JSON 구문 오류 없음
- [ ] 환경변수 사용 (하드코딩 금지)
- [ ] 파일 시스템 경로 제한
- [ ] 필요한 환경변수 설정 완료

---

## 5. Settings 설정 가이드

### 📐 파일 구조

**위치**: `.claude/settings.json`

**주요 섹션**:
```json
{
  "permissions": {
    "defaultMode": "default",
    "allow": [...],
    "ask": [...],
    "deny": [...]
  },
  "hooks": {
    "SessionStart": [...],
    "PreToolUse": [...]
  },
  "mcpServers": {...}
}
```

### 🔒 3단계 권한 관리

#### 1. `allow` - 자동 승인
안전하고 필수적인 도구만 허용:
```json
"allow": [
  "Read",
  "Write",
  "Edit",
  "Grep",
  "Glob",
  "Bash(git:*)",
  "Bash(pytest:*)"
]
```

#### 2. `ask` - 사용자 확인
중요하거나 변경 가능성이 있는 작업:
```json
"ask": [
  "Bash(git push:*)",
  "Bash(pip install:*)",
  "Bash(rm:*)"
]
```

#### 3. `deny` - 절대 금지
위험하거나 민감한 작업 차단:
```json
"deny": [
  "Read(./.env)",
  "Read(./secrets/**)",
  "Bash(sudo:*)",
  "Bash(rm -rf:*)"
]
```

### 🪝 훅 시스템

#### SessionStart 훅
세션 시작 시 프로젝트 정보 표시:
```json
{
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
    ]
  }
}
```

#### PreToolUse 훅
도구 실행 전 검증 및 차단:
```json
{
  "hooks": {
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

### ✅ 검증 체크리스트

- [ ] JSON 구문 오류 없음
- [ ] `allow`: 필수 도구만
- [ ] `ask`: 중요한 작업
- [ ] `deny`: 민감한 파일/명령
- [ ] Bash 패턴 구체화 (`Bash(git:*)`)
- [ ] 훅 파일 존재 및 실행 권한

---

## 📊 베스트 프랙티스

### 공통 원칙

1. **최소 권한 원칙**
   - 필요한 도구만 명시
   - Bash 사용 시 구체적 패턴

2. **공식 표준 준수**
   - YAML frontmatter 필수 필드
   - 파일명 규칙 (kebab-case)

3. **보안 우선**
   - 민감 정보 환경변수로 관리
   - 위험한 작업 차단

4. **문서화**
   - 명확한 설명
   - 구체적 예시
   - 검증 방법

---

**최종 업데이트**: 2025-10-19
**작성자**: @Alfred
