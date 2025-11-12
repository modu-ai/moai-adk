# Claude Code Architecture Research Report

**Date**: 2025-11-12
**Research Scope**: Claude Agent SDK Documentation
**Focus Areas**: Commands, Hooks, Tasks/Agents, MCP Integration
**Source**: https://docs.claude.com/en/docs/agent-sdk/

---

## Executive Summary

본 연구는 Claude Agent SDK의 공식 문서를 분석하여 Commands, Hooks, Tasks/Agents, MCP 통합에 대한 체계적인 정보를 수집했습니다. 주요 발견사항은 다음과 같습니다:

1. **Agent SDK는 4가지 핵심 확장 메커니즘을 제공**: Commands, Subagents, Skills, MCP Servers
2. **Hooks는 Plugins의 일부로 제공**되며 독립적인 문서가 없음
3. **Agent 위임은 자동/명시적 두 가지 패턴 지원**
4. **MCP 통합은 3가지 전송 방식 지원**: stdio, HTTP/SSE, SDK MCP Servers

---

## 1. Commands Architecture

### 1.1 정의 및 특징

**Slash Commands**는 특수한 제어 지시문으로 `/` 접두사로 시작하며, SDK의 `query()` 함수를 통해 전송됩니다.

#### Built-in Commands

| Command | Purpose | Return Data |
|---------|---------|-------------|
| `/compact` | 대화 기록 압축 (이전 메시지 요약) | 토큰 수, 트리거 정보 포함 메타데이터 |
| `/clear` | 새로운 세션 시작 (모든 기록 제거) | 새 세션 ID |

#### Custom Commands 생성

**파일 기반 정의**:
- **위치**: `.claude/commands/` (프로젝트) 또는 `~/.claude/commands/` (사용자)
- **파일명**: `{command-name}.md` (확장자 제외한 파일명이 명령어 이름)
- **내용**: Markdown 형식의 명령어 지시사항

**YAML Frontmatter 지원**:

```markdown
---
allowed-tools: Read, Grep, Glob
description: Command purpose description
model: claude-sonnet-4-5-20250929
argument-hint: "arg1 arg2"
---

Command instructions here.

Use $1 for first argument, $2 for second argument.
```

#### 고급 기능

1. **Arguments & Placeholders**: `$1`, `$2` 구문으로 동적 파라미터 전달
2. **Bash Execution**: `` !`command` `` 구문으로 셸 명령 실행 및 출력 포함
3. **File References**: `@filename` 구문으로 파일 내용 포함
4. **Namespacing**: 하위 디렉토리로 구조화 (예: `.claude/commands/frontend/component.md`)

#### SDK 사용 패턴

```typescript
for await (const message of query({
  prompt: "/command-name argument1 argument2",
  options: { maxTurns: 3 }
})) {
  // Handle message responses
}
```

### 1.2 Best Practices

- **단일 책임 원칙**: 명령어는 하나의 명확한 책임만 가져야 함
- **도구 제한**: `allowed-tools`로 보안 강화
- **명확한 네이밍**: 설명적인 이름과 frontmatter description 사용
- **네임스페이스 활용**: 대규모 명령어 컬렉션은 하위 디렉토리로 구조화
- **에러 핸들링**: 명령어 로직에 적절한 에러 처리 포함

---

## 2. Hooks Architecture

### 2.1 정의 및 위치

**Hooks는 Plugins의 일부**로 제공되며, 독립적인 문서가 존재하지 않습니다. SDK Plugins 문서에서 다음과 같이 언급됩니다:

> Plugins can include: Custom slash commands, Specialized agents (subagents), Skills, **Hooks responding to tool use and events**, MCP servers

### 2.2 Plugin 구조

```
my-plugin/
├── .claude-plugin/
│   └── plugin.json
├── commands/
├── agents/
├── skills/
└── hooks/          # Hooks는 여기에 위치
```

### 2.3 특징

- **이벤트 응답**: 도구 사용 및 이벤트에 반응
- **Plugin 로딩**: 프로그래매틱하게 파일시스템 경로 지정
- **Verification**: 초기화 메시지에서 로드된 플러그인 확인 가능

### 2.4 제한사항

공식 문서에서 Hooks의 다음 정보가 **누락**되어 있습니다:
- 구체적인 Hook 타입 목록
- Hook 생성 방법 및 API
- 트리거 메커니즘 상세
- 실행 순서 및 라이프사이클
- 예제 코드

**추천 조치**: Hooks에 대한 상세 정보는 추가 조사 필요 (GitHub 저장소, 샘플 코드 분석)

---

## 3. Tasks/Agents Architecture

### 3.1 Subagents 정의

**Subagents**는 특화된 작업을 수행하는 전문 에이전트로, 메인 에이전트로부터 분리된 컨텍스트를 유지합니다.

#### 생성 방법

**1. Programmatic Definition (권장)**

```typescript
for await (const message of query({
  prompt: "Your task here",
  options: {
    agents: {
      "security-auditor": {
        description: "Analyzes code for security vulnerabilities",
        prompt: "You are a security expert focused on OWASP Top 10...",
        tools: ["Read", "Grep", "Glob"],
        model: "claude-sonnet-4-5-20250929"
      }
    },
    maxTurns: 5
  }
})) {
  // Process responses
}
```

**2. Filesystem-Based Definition**

- **위치**: `.claude/agents/` (프로젝트) 또는 `~/.claude/agents/` (사용자)
- **파일**: Markdown + YAML frontmatter
- **우선순위**: Programmatic 정의가 filesystem 정의보다 우선

#### Configuration 요소

| 필드 | 필수 | 설명 |
|------|------|------|
| `description` | ✅ | 언제 사용할지 자연어 설명 |
| `prompt` | ✅ | 에이전트의 시스템 프롬프트 |
| `tools` | ❌ | 허용된 도구 배열 (생략 시 모든 도구 상속) |
| `model` | ❌ | 모델 오버라이드 |

### 3.2 핵심 기능

#### 1. Context Isolation (컨텍스트 격리)

> "Subagents maintain separate contexts, preventing information overload and keeping interactions focused."

- 전문 작업의 세부사항이 메인 대화를 오염시키지 않음
- 각 에이전트는 독립적인 대화 기록 유지

#### 2. Parallelization (병렬 처리)

> "Multiple subagents can run concurrently, dramatically speeding up complex workflows."

**예제 시나리오**: 코드 리뷰 시
- Security Agent: 보안 취약점 분석
- Test Coverage Agent: 테스트 커버리지 검증
- 두 에이전트가 **동시 실행**

#### 3. Specialized Expertise (전문화)

- 각 에이전트는 특정 도메인에 최적화된 시스템 프롬프트 보유
- Best practices, constraints, 전문 지식 포함

#### 4. Tool Restrictions (도구 제한)

- Read-only Agent: `['Read', 'Grep', 'Glob']`
- Test Runner: `['Bash', 'Read', 'Grep']`
- 최소 권한 원칙(Principle of Least Privilege) 적용

### 3.3 호출 패턴

#### 1. Automatic Invocation (자동 호출)

```typescript
agents: {
  "test-engineer": {
    description: "Invoked when user requests test creation or debugging tests",
    prompt: "You are a testing specialist...",
    tools: ["Bash", "Read", "Write"]
  }
}
```

SDK가 **task context에 기반하여 자동으로 적절한 에이전트 호출**

#### 2. Explicit Invocation (명시적 호출)

사용자가 프롬프트에서 특정 에이전트를 직접 요청:

```typescript
prompt: "Ask the security-auditor agent to review this code"
```

#### 3. Dynamic Creation (동적 생성)

애플리케이션 요구사항에 따라 런타임에 에이전트 생성:

```typescript
const agents = userPreferences.enableSecurityChecks
  ? { "security-auditor": { ... } }
  : {};
```

### 3.4 Best Practices

1. **Clear Descriptions**: 언제 호출되어야 하는지 명확한 설명 작성
2. **Appropriate Tool Combinations**: 각 에이전트의 책임에 맞는 도구 선택
3. **Specialized Prompts**: 도메인별 전문성 강화
4. **Concurrent Execution**: 독립적인 분석 작업은 병렬 실행 고려

---

## 4. Skills Architecture

### 4.1 정의 및 특징

**Agent Skills**는 Claude의 기능을 확장하는 전문 능력으로, `SKILL.md` 파일로 패키징됩니다.

> "Skills are packaged as `SKILL.md` files containing instructions, descriptions, and optional supporting resources."

#### 작동 방식

1. **Filesystem Storage**: 디렉토리 내 `SKILL.md` 파일 + YAML frontmatter + Markdown
2. **Automatic Discovery**: 시작 시 사용자/프로젝트 디렉토리에서 자동 발견
3. **Model-Invoked**: Claude가 컨텍스트 기반으로 자동 호출 결정
4. **Enabled via Configuration**: `allowed_tools`에 `"Skill"` 추가 + `settingSources` 설정 필요

### 4.2 위치 및 구조

**디렉토리 위치**:
- **Project Skills**: `.claude/skills/` (Git으로 공유)
- **User Skills**: `~/.claude/skills/` (개인용, 프로젝트 간 공유)

**파일 구조**:
```
.claude/skills/python-testing/
├── SKILL.md          # 메인 스킬 정의
└── examples/         # 선택적 리소스
    └── test_template.py
```

### 4.3 SDK 설정

**Critical Requirement**:

> "By default, the SDK does not load any filesystem settings. To use Skills, you must explicitly configure `settingSources: ['user', 'project']`"

```typescript
for await (const message of query({
  prompt: "Your task here",
  options: {
    settingSources: ['user', 'project'],  // 필수!
    allowedTools: ["Skill"],              // 필수!
    maxTurns: 5
  }
})) {
  // Process responses
}
```

### 4.4 제약사항

- **SDK Tool Control**: `SKILL.md`의 `allowed-tools` 필드는 **Claude Code CLI에서만 작동**, SDK에서는 무시됨
- **Main allowedTools**: SDK 애플리케이션은 메인 `allowedTools` 옵션으로 제어
- **No Programmatic API**: Subagents와 달리 프로그래매틱 등록 API 없음, 반드시 filesystem artifacts

### 4.5 Commands vs Skills vs Subagents 비교

| 특성 | Commands | Skills | Subagents |
|------|----------|--------|-----------|
| **호출 방식** | User-invoked (`/cmd`) | Model-invoked (자동) | User/Model-invoked |
| **정의 방식** | Filesystem only | Filesystem only | Programmatic + Filesystem |
| **컨텍스트** | 메인 컨텍스트 공유 | 메인 컨텍스트 공유 | 독립적 컨텍스트 |
| **용도** | 사용자 워크플로우 제어 | 전문 능력 확장 | 복잡한 작업 위임 |
| **병렬 실행** | ❌ | ❌ | ✅ |

---

## 5. MCP Integration Architecture

### 5.1 MCP 개요

**MCP (Model Context Protocol)**는 Claude Code를 커스텀 도구와 기능으로 확장하는 프로토콜입니다.

#### MCP Server 타입

| 타입 | 설명 | 통신 방식 |
|------|------|----------|
| **stdio Servers** | 외부 프로세스 | stdin/stdout |
| **HTTP/SSE Servers** | 원격 서버 | Network (headers, bearer tokens) |
| **SDK MCP Servers** | In-process | 직접 함수 호출 |

### 5.2 Configuration 패턴

#### 1. Basic Setup (.mcp.json)

프로젝트 루트에 `.mcp.json` 파일 생성:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["@modelcontextprotocol/server-filesystem"],
      "env": {
        "ALLOWED_PATHS": "/Users/me/projects"
      }
    }
  }
}
```

#### 2. SDK Integration

```typescript
import { query } from "@anthropic-ai/claude-agent-sdk";

for await (const message of query({
  prompt: "List files in my project",
  options: {
    mcpServers: {
      "filesystem": {
        command: "npx",
        args: ["@modelcontextprotocol/server-filesystem"],
        env: { "ALLOWED_PATHS": process.cwd() }
      }
    },
    allowedTools: ["mcp__filesystem__list_files"],
    maxTurns: 3
  }
})) {
  // Handle results
}
```

### 5.3 Custom Tools 생성

**SDK MCP Servers (In-Process)**를 사용하여 커스텀 도구 생성:

```typescript
import { createSdkMcpServer, tool } from "@anthropic-ai/claude-agent-sdk";
import { z } from "zod";

const customServer = createSdkMcpServer({
  name: "my-custom-tools",
  version: "1.0.0",
  tools: [
    tool(
      "get_weather",
      "Get current temperature for a location using coordinates",
      {
        latitude: z.number().describe("Latitude coordinate"),
        longitude: z.number().describe("Longitude coordinate")
      },
      async (args) => {
        // API 호출 로직
        const response = await fetch(
          `https://api.weather.com?lat=${args.latitude}&lon=${args.longitude}`
        );
        return await response.json();
      }
    )
  ]
});
```

**Critical Constraint**:

> "Custom MCP tools require streaming input mode. You must use an async generator/iterable for the `prompt` parameter - a simple string will not work."

```typescript
// ❌ 작동하지 않음
prompt: "Get weather for London"

// ✅ 작동함
async function* generateMessages() {
  yield { role: "user", content: "Get weather for London" };
}

for await (const message of query({
  prompt: generateMessages(),  // Async generator 필수!
  options: {
    mcpServers: { "my-custom-tools": customServer },
    allowedTools: ["mcp__my-custom-tools__get_weather"]
  }
})) { ... }
```

### 5.4 Tool Naming Convention

MCP 도구 이름은 특정 패턴을 따릅니다:

**패턴**: `mcp__{server_name}__{tool_name}`

**예제**:
- Server: `my-custom-tools`
- Tool: `get_weather`
- Full Name: `mcp__my-custom-tools__get_weather`

### 5.5 Resource Management

MCP 서버는 리소스를 제공하며, 에이전트는 다음 도구로 접근합니다:

- **`mcp__list_resources`**: 사용 가능한 리소스 탐색
- **`mcp__read_resource`**: 리소스 내용 접근

### 5.6 Authentication

#### Environment Variables

템플릿 구문과 기본값 지원:

```json
{
  "env": {
    "API_TOKEN": "${API_TOKEN}",
    "API_KEY": "${API_KEY:-default-key}"
  }
}
```

#### OAuth2

**제한사항**:
> "OAuth2 MCP authentication in-client is not currently supported."

### 5.7 Error Handling

#### Connection Status Monitoring

```typescript
for await (const message of query({ ... })) {
  if (message.type === "system") {
    // MCP 서버 연결 상태 확인
    const mcpStatus = message.mcp_servers;
    if (mcpStatus.some(s => s.status === "failed")) {
      console.error("MCP connection failed");
    }
  }
}
```

### 5.8 Best Practices

1. **Validate Connections**: 중요 작업 전 MCP 서버 연결 검증
2. **Specific allowedTools**: 에이전트 능력 제어를 위해 명시적 도구 리스트 사용
3. **Environment Templates**: 보안 자격증명 관리를 위한 환경 변수 템플릿 활용
4. **Type Safety**: Zod 스키마로 파라미터 검증 및 TypeScript 타입 추론
5. **Error Messages**: 예외 발생보다 의미 있는 에러 메시지 반환

---

## 6. Architecture Comparison

### 6.1 Commands vs Hooks vs Tasks 차이점

| 측면 | Commands | Hooks | Subagents (Tasks) |
|------|----------|-------|-------------------|
| **정의 위치** | `.claude/commands/` | `.claude-plugin/hooks/` | Code or `.claude/agents/` |
| **실행 트리거** | User-invoked (`/cmd`) | Event-driven (자동) | User/Model-invoked |
| **컨텍스트** | 메인 세션 공유 | 메인 세션 공유 | 독립적 컨텍스트 |
| **병렬 실행** | ❌ Sequential | ❌ Sequential | ✅ Concurrent |
| **도구 제한** | `allowed-tools` | Plugin-level | Agent-level `tools` |
| **프로그래매틱 생성** | ❌ | ❌ | ✅ |
| **용도** | 워크플로우 제어 | 가드레일, 검증 | 복잡한 작업 위임 |

### 6.2 Skills vs MCP Tools 차이점

| 측면 | Skills | MCP Tools |
|------|--------|-----------|
| **정의 방식** | Markdown files | TypeScript/Python code |
| **호출 방식** | Model-invoked (자동) | Explicit tool call |
| **프로그래매틱 생성** | ❌ Filesystem only | ✅ Programmatic |
| **실행 환경** | Claude's context | External process or in-process |
| **용도** | 전문 지식/패턴 제공 | 외부 시스템/API 통합 |
| **타입 안정성** | ❌ | ✅ (Zod schemas) |

### 6.3 계층 구조

```
User Input
    ↓
Commands (/cmd) ──────┐
    ↓                 │
Main Agent            │
    ↓                 │
    ├─→ Skills ───────┤──→ MCP Tools
    ├─→ Subagents ────┤       ↓
    └─→ Hooks ────────┘   External APIs
         (via Plugins)
```

---

## 7. Key Findings & Insights

### 7.1 Architecture Principles

#### 1. Delegation-First Pattern

Claude Agent SDK는 **명확한 위임 계층 구조**를 제공합니다:

```
Commands (Orchestration)
    ↓
Subagents (Specialized Execution)
    ↓
Skills (Knowledge) + MCP Tools (Integration)
```

**Rationale**:
- 관심사의 분리 (Separation of Concerns)
- 병렬 처리를 통한 성능 최적화
- 컨텍스트 오염 방지

#### 2. Filesystem-First Convention

대부분의 확장 메커니즘이 **filesystem artifacts**를 사용합니다:

- Commands: `.claude/commands/*.md`
- Agents: `.claude/agents/*.md`
- Skills: `.claude/skills/*/SKILL.md`
- MCP Config: `.mcp.json`

**장점**:
- Git을 통한 버전 관리
- 팀 간 공유 용이
- IDE에서 직접 편집 가능
- 프로그래밍 지식 없이도 확장 가능

**단점**:
- 동적 생성 제한 (Skills의 경우)
- 런타임 수정 불가능
- 프로그래매틱 제어 제한

#### 3. Programmatic Override

Critical extension points는 **programmatic definition**을 지원:

- **Subagents**: `agents` parameter (권장)
- **MCP Servers**: `mcpServers` parameter
- **Custom Tools**: `createSdkMcpServer()`

**Use Cases**:
- 사용자 설정 기반 동적 에이전트 생성
- 조건부 도구 활성화
- 런타임 최적화

### 7.2 Script Minimization Strategy

문서에서 **스크립트 사용 최소화**를 간접적으로 권장합니다:

#### 대안 패턴

| 기존 패턴 | Agent SDK 패턴 | 이유 |
|----------|----------------|------|
| Bash scripts | Commands + Bash tool | 컨텍스트 인식 실행 |
| Python scripts | MCP Custom Tools | 타입 안정성, 에러 핸들링 |
| Git hooks | Plugin Hooks | IDE 통합, 동적 결정 |
| Make/Task runners | Commands | 자연어 인터페이스 |

#### 예제: 테스트 자동화

**Traditional Approach**:
```bash
#!/bin/bash
# scripts/test.sh
pytest tests/ --cov=src --cov-report=html
```

**Agent SDK Approach**:
```markdown
---
# .claude/commands/test.md
allowed-tools: Bash, Read, Write
description: Run tests with coverage
---

Run pytest with coverage:
1. Execute: `pytest tests/ --cov=src --cov-report=html`
2. Read coverage report
3. Summarize results
4. Suggest improvements if coverage < 85%
```

**Benefits**:
- 테스트 결과 자동 분석
- 커버리지 개선 제안
- 실패한 테스트 디버깅 지원
- 자연어 인터페이스

### 7.3 Best Practices Summary

#### 1. Commands

- ✅ **DO**: 워크플로우 오케스트레이션에 사용
- ✅ **DO**: `allowed-tools`로 도구 제한
- ✅ **DO**: Namespacing으로 구조화
- ❌ **DON'T**: 복잡한 로직 포함 (Subagents로 위임)

#### 2. Subagents

- ✅ **DO**: 독립적인 작업은 병렬 실행
- ✅ **DO**: 명확한 description으로 자동 호출 지원
- ✅ **DO**: 최소 권한 원칙으로 도구 제한
- ❌ **DON'T**: 과도한 에이전트 생성 (3-5개 이하 권장)

#### 3. Skills

- ✅ **DO**: 반복적인 패턴/지식 캡슐화
- ✅ **DO**: Git으로 팀과 공유
- ✅ **DO**: `settingSources` 명시적 설정
- ❌ **DON'T**: 외부 API 호출 (MCP Tools 사용)

#### 4. MCP Tools

- ✅ **DO**: 외부 시스템 통합에 사용
- ✅ **DO**: Zod 스키마로 타입 안정성 확보
- ✅ **DO**: Streaming mode 사용 (async generator)
- ❌ **DON'T**: 간단한 로직 (Skills 사용)

---

## 8. Recommendations for MoAI-ADK

### 8.1 Current Architecture 검증

MoAI-ADK의 현재 아키텍처는 Claude Agent SDK의 **best practices와 높은 일치도**를 보입니다:

#### ✅ 일치하는 패턴

| MoAI-ADK | Claude SDK | 일치도 |
|----------|------------|--------|
| Commands → Agents → Skills | Commands → Subagents → Skills | 100% |
| `.claude/commands/` | `.claude/commands/` | 100% |
| `.claude/agents/` | `.claude/agents/` | 100% |
| `.claude/skills/` | `.claude/skills/` | 100% |
| Task() delegation | Subagents programmatic definition | 95% |

#### ⚠️ 차이점

1. **Hooks 구현**:
   - **MoAI-ADK**: `.claude/hooks/` (독립 디렉토리)
   - **Claude SDK**: `.claude-plugin/hooks/` (Plugin의 일부)
   - **영향**: 구조적 차이이지만 기능적으로 동일

2. **Agent 호출 메커니즘**:
   - **MoAI-ADK**: `Task(subagent_type="agent-name")`
   - **Claude SDK**: `agents` parameter in `query()`
   - **영향**: API 차이이지만 개념적으로 동일

### 8.2 최적화 제안

#### 1. Programmatic Agent Definition 지원

**현재 상태**: MoAI-ADK는 주로 filesystem-based agents 사용

**개선안**: Dynamic agent creation API 추가

```python
# 제안: config.json 기반 동적 에이전트 생성
def create_dynamic_agents(config):
    agents = {}

    if config["security"]["enabled"]:
        agents["security-auditor"] = {
            "description": "Security vulnerability analysis",
            "prompt": load_skill("security-expert"),
            "tools": ["Read", "Grep", "Bash"]
        }

    if config["performance"]["enabled"]:
        agents["performance-optimizer"] = {
            "description": "Performance bottleneck detection",
            "prompt": load_skill("performance-engineer"),
            "tools": ["Read", "Bash"]
        }

    return agents
```

**Benefits**:
- 사용자 설정 기반 에이전트 활성화
- 런타임 최적화 (사용하지 않는 에이전트 로드 안 함)
- 조건부 전문가 활성화

#### 2. MCP Integration 강화

**현재 상태**: MoAI-ADK는 `.mcp.json` 지원 시작 (v0.20.0)

**개선안**: MCP Custom Tools 생성 가이드 추가

```python
# 제안: MoAI-ADK용 MCP Tool 템플릿
# .moai/scripts/mcp/database_query_tool.py

from anthropic_sdk import createSdkMcpServer, tool
from pydantic import BaseModel, Field

class DatabaseQueryArgs(BaseModel):
    query: str = Field(description="SQL query to execute")
    database: str = Field(description="Database name")

@tool(
    name="execute_query",
    description="Execute SQL query on project database",
    args_schema=DatabaseQueryArgs
)
async def execute_query(args: DatabaseQueryArgs):
    # Implementation
    result = await db.execute(args.query, args.database)
    return result.to_dict()

mcp_server = createSdkMcpServer({
    "name": "moai-database-tools",
    "version": "1.0.0",
    "tools": [execute_query]
})
```

**Benefits**:
- 프로젝트별 커스텀 도구 생성
- 외부 시스템 통합 간소화
- 타입 안정성 확보

#### 3. Hooks 문서화 개선

**현재 상태**: MoAI-ADK Hooks는 잘 작동하지만 공식 SDK 문서 부족

**개선안**: Hooks 사용 예제 및 패턴 문서 추가

**예제: PreToolUse Hook Pattern**

```python
# .claude/hooks/pre_tool_use/validate_git_operations.py

def pre_tool_use(tool_name: str, tool_args: dict) -> bool:
    """Validate git operations before execution"""

    if tool_name == "Bash" and "git push" in tool_args.get("command", ""):
        # Check if on main/master branch
        current_branch = get_current_branch()
        if current_branch in ["main", "master"]:
            return confirm_user("Pushing to main. Continue?")

    return True  # Allow execution
```

**Benefits**:
- 안전한 Git 워크플로우
- 런타임 검증
- 사용자 확인 프롬프트

#### 4. Skills Progressive Disclosure 최적화

**현재 상태**: 55개 Skills, 토큰 사용량 높음

**개선안**: Skill 메타데이터 기반 필터링

```python
# 제안: Skill 선택적 로드
# .moai/config/config.json

{
  "skills": {
    "auto_load": false,
    "load_on_demand": true,
    "categories": {
      "foundation": ["enabled"],
      "core": ["enabled"],
      "workflow": ["on_demand"],
      "domain": ["on_demand"],
      "integration": ["disabled"],
      "advanced": ["disabled"]
    }
  }
}
```

**Benefits**:
- 초기 토큰 사용량 감소
- 필요한 Skills만 로드
- 성능 최적화

### 8.3 Migration Path

#### Phase 1: Documentation Alignment (즉시)

- [ ] Hooks 문서화 (공식 SDK 패턴과 비교)
- [ ] MCP Integration 가이드 추가
- [ ] Programmatic API 문서화

#### Phase 2: API Enhancement (Short-term)

- [ ] Dynamic agent creation API 구현
- [ ] MCP Custom Tools 템플릿 추가
- [ ] Skills 선택적 로드 구현

#### Phase 3: Optimization (Mid-term)

- [ ] Agent 병렬 실행 최적화
- [ ] Token usage monitoring
- [ ] Performance benchmarking

---

## 9. Code Examples

### 9.1 Complete Command Example

```markdown
---
# .claude/commands/alfred/1-plan.md
allowed-tools: Read, Write, Grep, Glob, Task
description: Create SPEC document using plan-agent
model: claude-sonnet-4-5-20250929
argument-hint: "feature description"
---

# Alfred Plan Command

Create a comprehensive SPEC document for the requested feature.

## Arguments
- $1: Feature description (natural language)

## Workflow
1. Analyze feature requirements
2. Delegate to plan-agent:
   Task(
     prompt="Create SPEC for: $1",
     subagent_type="plan-agent"
   )
3. Review SPEC structure
4. Ask user for confirmation

## Output
- SPEC document in .moai/specs/SPEC-XXX/spec.md
- Initial TodoWrite initialization
```

### 9.2 Complete Subagent Example

```typescript
// Programmatic subagent definition
const agents = {
  "plan-agent": {
    description: `Invoked when creating SPEC documents or planning features.
                  Analyzes requirements, identifies dependencies, estimates work.`,
    prompt: `You are a planning specialist focused on SPEC-first development.

             Your responsibilities:
             1. Analyze user requirements thoroughly
             2. Break down features into tasks
             3. Identify dependencies and risks
             4. Create structured SPEC documents

             SPEC Structure:
             - Title and description
             - Acceptance criteria
             - Technical requirements
             - @TAG assignments

             Follow TRUST 5 principles:
             - Test-first mindset
             - Readable specifications
             - Unified patterns
             - Secured by design
             - Trackable via @TAGs`,
    tools: ["Read", "Write", "Grep", "Glob"],
    model: "claude-sonnet-4-5-20250929"
  },

  "tdd-implementer": {
    description: `Invoked when implementing features following TDD methodology.
                  Writes tests first (RED), then implementation (GREEN), then refactors.`,
    prompt: `You are a TDD specialist following strict RED-GREEN-REFACTOR cycle.

             RED Phase:
             1. Write failing tests based on SPEC
             2. Verify tests fail for the right reason
             3. Include edge cases and error scenarios

             GREEN Phase:
             1. Write minimal code to pass tests
             2. No premature optimization
             3. Focus on functionality first

             REFACTOR Phase:
             1. Improve code structure
             2. Remove duplication
             3. Enhance readability
             4. Ensure tests still pass

             Always include @TEST and @CODE TAGs for traceability.`,
    tools: ["Read", "Write", "Edit", "Bash", "Grep"],
    model: "claude-sonnet-4-5-20250929"
  }
};

// SDK usage
for await (const message of query({
  prompt: "Implement authentication feature from SPEC-001",
  options: {
    agents: agents,
    maxTurns: 10,
    allowedTools: ["Read", "Write", "Edit", "Bash", "Grep", "Glob"]
  }
})) {
  if (message.type === "text") {
    console.log(message.content);
  }
}
```

### 9.3 Complete Skill Example

```markdown
---
# .claude/skills/moai-foundation-tags/SKILL.md
name: moai-foundation-tags
description: Invoked when working with @TAG system for traceability
version: 1.0.0
---

# MoAI Foundation: @TAG System

## Purpose
@TAGs provide complete traceability across SPEC → TEST → CODE → DOC lifecycle.

## TAG Types

### @SPEC-XXX
- **Location**: SPEC documents
- **Format**: @SPEC-001, @SPEC-002
- **Usage**: Unique identifier for each specification

### @TEST-XXX
- **Location**: Test files
- **Format**: @TEST-001 → @SPEC-001
- **Usage**: Links tests to specifications
- **Required**: Must reference parent @SPEC

### @CODE-XXX
- **Location**: Implementation files
- **Format**: @CODE-001 → @TEST-001 → @SPEC-001
- **Usage**: Links implementation to tests and specs
- **Required**: Must reference parent @TEST and @SPEC

### @DOC-XXX
- **Location**: Documentation files
- **Format**: @DOC-001 → @CODE-001 → @TEST-001 → @SPEC-001
- **Usage**: Links documentation to implementation
- **Optional**: References parent @CODE when applicable

## Usage Examples

### Test File
```python
# tests/test_authentication.py
# @TEST-001 → @SPEC-001

import pytest
from src.auth import login  # @CODE-001

def test_successful_login():  # @TEST-001
    """Test user can login with valid credentials"""
    result = login("user@example.com", "password123")
    assert result.success is True
```

### Implementation File
```python
# src/auth.py
# @CODE-001 → @TEST-001 → @SPEC-001

def login(email: str, password: str):  # @CODE-001
    """Authenticate user with email and password

    Related: @TEST-001, @SPEC-001
    """
    # Implementation
    pass
```

## Traceability Chain

```
@SPEC-001 (Requirement)
    ↓
@TEST-001 (Verification)
    ↓
@CODE-001 (Implementation)
    ↓
@DOC-001 (Documentation)
```

## Best Practices

1. **Always start with @SPEC**: No @TEST or @CODE without @SPEC
2. **Link backward**: Each TAG must reference its parent
3. **Be consistent**: Use the same TAG across commits, PRs, reviews
4. **Update docs**: Sync @DOC TAGs when @CODE changes
```

### 9.4 Complete MCP Custom Tool Example

```typescript
// .moai/scripts/mcp/project-analysis-tools.ts

import { createSdkMcpServer, tool } from "@anthropic-ai/claude-agent-sdk";
import { z } from "zod";
import { exec } from "child_process";
import { promisify } from "util";

const execAsync = promisify(exec);

// Tool 1: Code Quality Analysis
const analyzeCodeQuality = tool(
  "analyze_code_quality",
  "Analyze code quality metrics including coverage, linting, type checking",
  {
    directory: z.string().describe("Directory to analyze"),
    include_tests: z.boolean().default(true).describe("Include test coverage")
  },
  async (args) => {
    const results = {
      coverage: null,
      linting: null,
      typecheck: null
    };

    // Run coverage
    if (args.include_tests) {
      const { stdout } = await execAsync(`pytest ${args.directory} --cov --cov-report=json`);
      const coverageData = JSON.parse(stdout);
      results.coverage = {
        percentage: coverageData.totals.percent_covered,
        lines_covered: coverageData.totals.covered_lines,
        lines_total: coverageData.totals.num_statements
      };
    }

    // Run linting
    try {
      await execAsync(`ruff check ${args.directory} --output-format=json`);
      results.linting = { status: "passed", issues: [] };
    } catch (error) {
      const issues = JSON.parse(error.stdout);
      results.linting = { status: "failed", issues };
    }

    // Run type checking
    try {
      await execAsync(`mypy ${args.directory} --json`);
      results.typecheck = { status: "passed", errors: [] };
    } catch (error) {
      const errors = JSON.parse(error.stdout);
      results.typecheck = { status: "failed", errors };
    }

    return results;
  }
);

// Tool 2: SPEC Validation
const validateSpec = tool(
  "validate_spec",
  "Validate SPEC document structure and completeness",
  {
    spec_path: z.string().describe("Path to SPEC document"),
    strict: z.boolean().default(false).describe("Enable strict validation")
  },
  async (args) => {
    const fs = require("fs").promises;
    const content = await fs.readFile(args.spec_path, "utf-8");

    const validations = {
      has_title: /^#\s+.+/m.test(content),
      has_description: /##\s+Description/i.test(content),
      has_acceptance_criteria: /##\s+Acceptance Criteria/i.test(content),
      has_tags: /@SPEC-\d{3,}/.test(content),
      has_technical_requirements: /##\s+Technical Requirements/i.test(content)
    };

    const issues = [];
    if (!validations.has_title) issues.push("Missing title");
    if (!validations.has_description) issues.push("Missing description section");
    if (!validations.has_acceptance_criteria) issues.push("Missing acceptance criteria");
    if (!validations.has_tags) issues.push("Missing @SPEC TAG");

    if (args.strict && !validations.has_technical_requirements) {
      issues.push("Missing technical requirements (strict mode)");
    }

    return {
      valid: issues.length === 0,
      validations,
      issues,
      score: Object.values(validations).filter(v => v).length / Object.keys(validations).length
    };
  }
);

// Tool 3: Git Branch Analysis
const analyzeGitBranch = tool(
  "analyze_git_branch",
  "Analyze current git branch status, commits, and diff",
  {
    include_diff: z.boolean().default(true).describe("Include diff analysis")
  },
  async (args) => {
    const branch = (await execAsync("git branch --show-current")).stdout.trim();
    const commits = (await execAsync("git log origin/develop..HEAD --oneline")).stdout.trim();
    const status = (await execAsync("git status --porcelain")).stdout.trim();

    const result = {
      current_branch: branch,
      commits_ahead: commits.split("\n").filter(l => l).length,
      uncommitted_changes: status.split("\n").filter(l => l).length,
      commits: commits.split("\n").map(line => {
        const [hash, ...message] = line.split(" ");
        return { hash, message: message.join(" ") };
      })
    };

    if (args.include_diff) {
      const diff = (await execAsync("git diff origin/develop...HEAD --stat")).stdout;
      result.diff_summary = diff;
    }

    return result;
  }
);

// Create MCP Server
export const projectAnalysisServer = createSdkMcpServer({
  name: "moai-project-analysis",
  version: "1.0.0",
  tools: [
    analyzeCodeQuality,
    validateSpec,
    analyzeGitBranch
  ]
});

// SDK Usage Example
async function* generateMessages() {
  yield {
    role: "user",
    content: "Analyze the current project quality and create a report"
  };
}

for await (const message of query({
  prompt: generateMessages(),  // Must be async generator!
  options: {
    mcpServers: {
      "moai-project-analysis": projectAnalysisServer
    },
    allowedTools: [
      "mcp__moai-project-analysis__analyze_code_quality",
      "mcp__moai-project-analysis__validate_spec",
      "mcp__moai-project-analysis__analyze_git_branch",
      "Read",
      "Write"
    ],
    maxTurns: 5
  }
})) {
  if (message.type === "text") {
    console.log(message.content);
  }
}
```

---

## 10. Conclusion

### 10.1 Summary

Claude Agent SDK는 **강력한 아키텍처**를 제공하며 MoAI-ADK의 현재 설계와 **높은 일치도**를 보입니다:

1. **Commands → Agents → Skills 계층**: 명확한 관심사 분리
2. **Filesystem-First + Programmatic Override**: 유연한 확장성
3. **Delegation Pattern**: 병렬 처리 및 컨텍스트 격리
4. **MCP Integration**: 외부 시스템 통합 표준화

### 10.2 Next Steps

#### Immediate Actions (이번 주)

1. **Hooks 문서화**: 공식 SDK 패턴과 비교하여 상세 문서 작성
2. **MCP 가이드 작성**: Custom Tools 생성 튜토리얼 추가
3. **Architecture 문서 업데이트**: SDK 공식 패턴 반영

#### Short-term Goals (이번 달)

1. **Dynamic Agent API**: Programmatic agent creation 구현
2. **Skills 최적화**: Progressive disclosure 개선
3. **MCP Templates**: 프로젝트별 커스텀 도구 템플릿

#### Long-term Vision (다음 분기)

1. **Performance Optimization**: Agent 병렬 실행 최적화
2. **Token Usage Monitoring**: 실시간 토큰 사용량 추적
3. **Plugin Ecosystem**: 커뮤니티 플러그인 지원

---

## References

- [Claude Agent SDK Overview](https://docs.claude.com/en/docs/agent-sdk/overview)
- [Slash Commands Documentation](https://docs.claude.com/en/docs/agent-sdk/slash-commands)
- [Subagents Guide](https://docs.claude.com/en/docs/agent-sdk/subagents)
- [Skills Documentation](https://docs.claude.com/en/docs/agent-sdk/skills)
- [Custom Tools Guide](https://docs.claude.com/en/docs/agent-sdk/custom-tools)
- [MCP Integration](https://docs.claude.com/en/docs/agent-sdk/mcp)
- [SDK Plugins](https://docs.claude.com/en/docs/agent-sdk/plugins)

---

**Report Generated**: 2025-11-12
**Author**: Alfred (MoAI-ADK SuperAgent)
**Version**: 1.0.0
**Status**: Completed
