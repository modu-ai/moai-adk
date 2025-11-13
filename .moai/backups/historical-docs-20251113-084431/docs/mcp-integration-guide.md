# MCP Integration Guide
**MoAI-ADK Agent Orchestration with MCP Tools**
**Version**: 1.0.0
**Date**: 2025-11-12

---

## 개요

MoAI-ADK agents가 MCP (Model Context Protocol) 도구들을 어떻게 활용하는지 설명합니다.

**MCP Servers in MoAI-ADK**:

1. **Context7**: Library documentation lookup
2. **Playwright**: E2E test automation
3. **Sequential Thinking**: Deep analytical reasoning

**Agent Orchestration Pattern**:

- MCP Integrator agents: 전용 agents (mcp-context7-integrator, mcp-playwright-integrator, mcp-sequential-thinking-integrator)
- Direct tool usage: Agents가 직접 MCP tools 사용
- Alfred coordination: 모든 MCP 호출은 Alfred context를 통해 추적

---

## MCP Server 1: Context7

### 용도

최신 라이브러리 공식 문서를 실시간으로 가져옵니다.

**Use Cases**:

- SPEC 작성 시 최신 라이브러리 버전 확인
- 구현 시 API 사용법 참조
- Migration 시 breaking changes 확인

---

### MCP Tools

#### `mcp__context7__resolve-library-id`

**기능**: Library 이름을 Context7 ID로 변환

**입력**:

```json
{
  "libraryName": "FastAPI"
}
```

**출력**:

```json
{
  "library_id": "/tiangolo/fastapi",
  "name": "FastAPI",
  "description": "FastAPI framework for building APIs",
  "trust_score": 10
}
```

---

#### `mcp__context7__get-library-docs`

**기능**: 라이브러리 문서 조회

**입력**:

```json
{
  "context7CompatibleLibraryID": "/tiangolo/fastapi",
  "topic": "authentication",
  "tokens": 5000
}
```

**출력**:

```json
{
  "docs": "# FastAPI Authentication\n\n...",
  "sections": ["OAuth2", "JWT", "API Keys"],
  "version": "0.118.3"
}
```

---

### Agent 통합: mcp-context7-integrator

#### Agent Definition

```yaml
---
name: mcp-context7-integrator
description: "Use PROACTIVELY when Context7 library documentation is needed."
tools: [mcp__context7__resolve-library-id, mcp__context7__get-library-docs]
model: haiku

orchestration:
  can_resume: false
  typical_chain_position: "support"
  depends_on: []
  resume_pattern: "independent_lookup"
  session_strategy: "independent"

coordination:
  returns_to_alfred: true
  spawns_subagents: false
  requires_approval: false
  parallel_safe: true

performance:
  avg_execution_time_ms: 3000
  token_intensive: false
  cache_friendly: true
---
```

---

### 사용 예제 1: SPEC 작성 시 라이브러리 버전 확인

**Workflow**: spec-builder → mcp-context7-integrator → spec-builder (resume)

```python
# In /alfred:1-plan command

# STEP 1: spec-builder가 SPEC 초안 작성
spec_result = Task(
    subagent_type="spec-builder",
    prompt="""Create SPEC for user authentication.

    Tech stack: FastAPI, SQLAlchemy, PostgreSQL

    Specify latest stable versions for libraries.
    """
)

spec_builder_id = spec_result["agent_id"]

# STEP 2: Context7로 최신 버전 확인
context7_result = Task(
    subagent_type="mcp-context7-integrator",
    prompt="""Lookup latest stable versions:

    Libraries:
    - FastAPI
    - SQLAlchemy
    - asyncpg (PostgreSQL driver)

    For each library, provide:
    1. Latest stable version
    2. Key features in latest version
    3. Breaking changes (if upgrading)
    """
)

register_agent(
    agent_name="mcp-context7-integrator",
    agent_id=context7_result["agent_id"],
    result=context7_result,
    chain_id=f"{spec_result['spec_id']}-library-lookup"
)

# STEP 3: spec-builder가 SPEC 업데이트 (resume)
updated_spec = Task(
    subagent_type="spec-builder",
    prompt=f"""Update SPEC with verified library versions:

    Library versions from Context7:
    {json.dumps(context7_result['versions'], indent=2)}

    Update tech stack section with specific versions.
    """,
    resume=spec_builder_id  # 🔑 Resume to maintain SPEC context
)
```

**결과**: SPEC에 최신 검증된 버전 명시

---

### 사용 예제 2: 구현 시 API 사용법 참조

**Workflow**: tdd-implementer → mcp-context7-integrator (직접 tool 사용)

```python
# In tdd-implementer agent

# Option A: Agent가 직접 MCP tool 사용
# (tdd-implementer의 tools 목록에 MCP tools 포함 필요)

# STEP 1: FastAPI OAuth2 문서 조회
library_id = mcp__context7__resolve-library-id(libraryName="FastAPI")

docs = mcp__context7__get-library-docs(
    context7CompatibleLibraryID=library_id["library_id"],
    topic="OAuth2 authentication",
    tokens=5000
)

# STEP 2: 문서 기반 코드 작성
# Write tests based on official docs
Write("tests/test_auth.py", f"""
# Based on FastAPI OAuth2 docs: {docs['version']}

{generate_test_code(docs)}
""")
```

**또는 Integrator Agent 사용**:

```python
# Option B: Integrator agent에게 위임

# In /alfred:2-run command
fastapi_docs = Task(
    subagent_type="mcp-context7-integrator",
    prompt="""Get FastAPI OAuth2 authentication documentation.

    Focus on:
    - OAuth2PasswordBearer usage
    - JWT token creation
    - Protected routes
    """
)

# Pass to tdd-implementer
impl_result = Task(
    subagent_type="tdd-implementer",
    prompt=f"""Implement OAuth2 authentication.

    Official FastAPI docs:
    {fastapi_docs['docs']}

    Follow official patterns exactly.
    """
)
```

---

## MCP Server 2: Playwright

### 용도

E2E (End-to-End) 테스트 자동화, 브라우저 자동화

**Use Cases**:

- UI 컴포넌트 E2E 테스트
- User flow 검증
- 스크린샷 캡처 및 비교

---

### MCP Tools

#### `mcp__playwright__*`

**주요 Tools**:

- `navigate`: URL 이동
- `click`: 요소 클릭
- `fill`: 입력 필드 채우기
- `screenshot`: 스크린샷 캡처
- `evaluate`: JavaScript 실행

---

### Agent 통합: mcp-playwright-integrator

#### Agent Definition

```yaml
---
name: mcp-playwright-integrator
description: "Use PROACTIVELY when E2E test automation with Playwright is needed."
tools: [Read, Write, Bash, mcp__playwright__*]
model: haiku

orchestration:
  can_resume: true
  typical_chain_position: "support"
  depends_on: []
  resume_pattern: "test_scenario"
  session_strategy: "resumable"

coordination:
  returns_to_alfred: true
  spawns_subagents: false
  requires_approval: false
  parallel_safe: false  # Browser interaction

performance:
  avg_execution_time_ms: 25000
  token_intensive: false
  cache_friendly: false
---
```

---

### 사용 예제 1: 로그인 Flow E2E 테스트

**Workflow**: mcp-playwright-integrator (resume로 여러 시나리오)

```python
# In /alfred:2-run command

# STEP 1: E2E 테스트 시나리오 1 - Happy Path
e2e_result_1 = Task(
    subagent_type="mcp-playwright-integrator",
    prompt="""Create E2E test for login flow (happy path).

    Scenario:
    1. Navigate to /login
    2. Fill username and password
    3. Click "Login" button
    4. Verify redirect to /dashboard
    5. Check "Welcome, {username}" message

    Generate Playwright test code.
    """
)

playwright_id = e2e_result_1["agent_id"]
register_agent("mcp-playwright-integrator", playwright_id, e2e_result_1, chain_id="login-e2e")

# STEP 2: E2E 테스트 시나리오 2 - Invalid Credentials (resume)
e2e_result_2 = Task(
    subagent_type="mcp-playwright-integrator",
    prompt="""Continue E2E tests for login flow (error case).

    Scenario:
    1. Navigate to /login
    2. Fill invalid username/password
    3. Click "Login" button
    4. Verify error message appears
    5. Check still on /login page

    Add to existing test file.
    """,
    resume=playwright_id  # 🔑 Resume to add to same test file
)

session_mgr.increment_resume_count(playwright_id)

# STEP 3: E2E 테스트 시나리오 3 - Logout (resume)
e2e_result_3 = Task(
    subagent_type="mcp-playwright-integrator",
    prompt="""Continue E2E tests for complete auth flow (logout).

    Scenario:
    1. Login (use happy path)
    2. Click "Logout" button
    3. Verify redirect to /login
    4. Check session cleared

    Complete test suite.
    """,
    resume=playwright_id  # 🔑 Resume for complete test suite
)

session_mgr.increment_resume_count(playwright_id)
```

**Resume의 이점**:

- ✅ 모든 시나리오가 하나의 test file에 추가됨
- ✅ 일관된 test 구조 유지
- ✅ 공통 setup/teardown 공유

---

### 사용 예제 2: 직접 Playwright Tool 사용

**Agent가 직접 MCP tools 사용** (Agent에 tools 권한 필요):

```python
# In mcp-playwright-integrator agent

# STEP 1: Browser 시작 및 페이지 이동
page = mcp__playwright__navigate(url="http://localhost:3000/login")

# STEP 2: 폼 채우기
mcp__playwright__fill(selector="#username", value="testuser")
mcp__playwright__fill(selector="#password", value="password123")

# STEP 3: 로그인 버튼 클릭
mcp__playwright__click(selector="button[type='submit']")

# STEP 4: 결과 확인
current_url = mcp__playwright__evaluate(expression="window.location.href")
assert "/dashboard" in current_url

# STEP 5: 스크린샷 캡처
screenshot_path = mcp__playwright__screenshot(path=".moai/temp/login-success.png")

# STEP 6: 테스트 코드 생성
Write("tests/e2e/test_login.py", f"""
import pytest
from playwright.sync_api import Page

def test_login_success(page: Page):
    # Navigate
    page.goto("http://localhost:3000/login")

    # Fill form
    page.fill("#username", "testuser")
    page.fill("#password", "password123")

    # Submit
    page.click("button[type='submit']")

    # Verify
    assert "/dashboard" in page.url
    assert page.locator("text=Welcome, testuser").is_visible()
""")
```

---

## MCP Server 3: Sequential Thinking

### 용도

복잡한 문제에 대한 단계별 심층 분석

**Use Cases**:

- 아키텍처 설계 의사결정
- 복잡한 버그 원인 분석
- Migration 전략 수립
- Performance 병목 분석

---

### MCP Tools

#### `mcp__sequential_thinking_think`

**기능**: 복잡한 문제를 단계별로 사고

**입력**:

```json
{
  "prompt": "Should we use microservices or monolith for this project?",
  "context": {
    "team_size": 5,
    "expected_scale": "medium",
    "timeline": "3 months"
  }
}
```

**출력**:

```json
{
  "thinking_process": [
    "Step 1: Analyze team size and expertise...",
    "Step 2: Consider timeline constraints...",
    "Step 3: Evaluate scalability needs...",
    "Step 4: Compare trade-offs..."
  ],
  "conclusion": "Monolith with modular architecture",
  "reasoning": "...",
  "alternatives": ["Microservices", "Hybrid"]
}
```

---

### Agent 통합: mcp-sequential-thinking-integrator

#### Agent Definition

```yaml
---
name: mcp-sequential-thinking-integrator
description: "Use PROACTIVELY when deep analytical thinking is needed."
tools: [mcp__sequential_thinking_think]
model: sonnet  # Sonnet for better reasoning

orchestration:
  can_resume: true
  typical_chain_position: "support"
  depends_on: []
  resume_pattern: "deep_analysis"
  session_strategy: "resumable"

coordination:
  returns_to_alfred: true
  spawns_subagents: false
  requires_approval: false
  parallel_safe: true

performance:
  avg_execution_time_ms: 18000
  token_intensive: true
  cache_friendly: false
---
```

---

### 사용 예제 1: 아키텍처 설계 의사결정

**Workflow**: backend-expert → mcp-sequential-thinking-integrator → backend-expert (resume)

```python
# In /alfred:1-plan command

# STEP 1: backend-expert가 아키텍처 옵션 제시
backend_analysis = Task(
    subagent_type="backend-expert",
    prompt="""Analyze backend architecture options for SPEC-AUTH-001.

    Provide:
    1. Architecture options (Monolith, Microservices, Serverless)
    2. Trade-offs for each
    3. Initial recommendation
    """
)

backend_expert_id = backend_analysis["agent_id"]

# STEP 2: Sequential Thinking으로 심층 분석
thinking_result = Task(
    subagent_type="mcp-sequential-thinking-integrator",
    prompt=f"""Deep analysis of architecture decision.

    Context:
    {json.dumps(backend_analysis['context'], indent=2)}

    Options:
    {json.dumps(backend_analysis['options'], indent=2)}

    Analyze:
    1. Long-term maintainability
    2. Team skill requirements
    3. Deployment complexity
    4. Cost implications
    5. Migration path (if needed later)

    Provide step-by-step reasoning and final recommendation.
    """
)

register_agent(
    agent_name="mcp-sequential-thinking-integrator",
    agent_id=thinking_result["agent_id"],
    result=thinking_result,
    chain_id="SPEC-AUTH-001-architecture-decision"
)

# STEP 3: backend-expert가 최종 결정 (resume)
final_decision = Task(
    subagent_type="backend-expert",
    prompt=f"""Finalize architecture decision.

    Sequential thinking analysis:
    {json.dumps(thinking_result['conclusion'], indent=2)}

    Reasoning:
    {thinking_result['reasoning']}

    Make final recommendation and update SPEC.
    """,
    resume=backend_expert_id  # 🔑 Resume to maintain architecture context
)
```

**결과**: 심층 분석 기반 아키텍처 결정, SPEC에 반영

---

### 사용 예제 2: 복잡한 버그 분석

**Workflow**: debug-helper → mcp-sequential-thinking-integrator → debug-helper (resume)

```python
# In /alfred:2-run command (quality gate failed)

# STEP 1: debug-helper가 초기 분석
debug_analysis = Task(
    subagent_type="debug-helper",
    prompt=f"""Analyze quality gate failures.

    Issues:
    {json.dumps(qa_result['issues'], indent=2)}

    Error logs:
    {error_logs}

    Provide initial diagnosis.
    """
)

debug_helper_id = debug_analysis["agent_id"]

# STEP 2: 복잡한 이슈에 대해 Sequential Thinking 사용
if debug_analysis.get("complexity") == "high":
    thinking_result = Task(
        subagent_type="mcp-sequential-thinking-integrator",
        prompt=f"""Deep root cause analysis.

        Error: {debug_analysis['main_error']}

        Context:
        - Code structure: {debug_analysis['code_context']}
        - Recent changes: {debug_analysis['recent_changes']}
        - Dependencies: {debug_analysis['dependencies']}

        Analyze:
        1. What are all possible root causes?
        2. Which is most likely based on evidence?
        3. What additional information is needed?
        4. What is the fix strategy?

        Provide step-by-step analysis.
        """
    )

    register_agent(
        agent_name="mcp-sequential-thinking-integrator",
        agent_id=thinking_result["agent_id"],
        result=thinking_result,
        chain_id="debug-session"
    )

    # STEP 3: debug-helper가 수정 제안 (resume)
    fix_recommendation = Task(
        subagent_type="debug-helper",
        prompt=f"""Generate fix based on deep analysis.

        Root cause: {thinking_result['conclusion']}

        Reasoning:
        {thinking_result['reasoning']}

        Provide:
        1. Specific code changes
        2. Test cases to prevent regression
        3. Verification steps
        """,
        resume=debug_helper_id  # 🔑 Resume to maintain debug context
    )
```

---

## Agent Tool Permissions

### Frontend Matter에 MCP Tools 명시

```yaml
# Example: spec-builder.md
---
name: spec-builder
tools:
  - Read
  - Write
  - Edit
  - MultiEdit
  - Bash
  - Grep
  - Glob
  - TodoWrite
  - WebFetch
  - AskUserQuestion
  - mcp__sequential_thinking_think  # 🔑 MCP tool
  - mcp__context7__resolve-library-id  # 🔑 MCP tool
  - mcp__context7__get-library-docs  # 🔑 MCP tool
model: inherit
---
```

**Rule**: Agent가 MCP tool을 직접 사용하려면 tools 목록에 명시 필요

---

## MCP Integrator vs Direct Usage

### When to Use Integrator Agents

✅ **Use mcp-xxx-integrator when**:

- 복잡한 MCP workflow 필요
- 여러 MCP 호출을 조합
- MCP 결과를 가공/분석 필요
- Resume으로 연속 작업 (예: Playwright test scenarios)

**Example**:

```python
# Integrator agent handles complex Playwright workflow
playwright_result = Task(
    subagent_type="mcp-playwright-integrator",
    prompt="Create complete E2E test suite with multiple scenarios"
)
```

---

### When to Use Direct Tool Call

✅ **Use direct MCP tools when**:

- 단순 조회/실행
- Agent 내부에서 즉시 사용
- Workflow 중단 없이 진행

**Example**:

```python
# In tdd-implementer agent
# Direct call for quick library lookup
library_id = mcp__context7__resolve-library-id(libraryName="FastAPI")
docs = mcp__context7__get-library-docs(context7CompatibleLibraryID=library_id)

# Use docs immediately in code generation
Write("src/auth.py", generate_code_from_docs(docs))
```

---

## MCP Integration Patterns

### Pattern 1: Library Lookup Chain

```
spec-builder → mcp-context7-integrator → spec-builder (resume)
```

**용도**: SPEC 작성 시 최신 라이브러리 버전 확인 및 반영

---

### Pattern 2: Deep Analysis Chain

```
backend-expert → mcp-sequential-thinking-integrator → backend-expert (resume)
```

**용도**: 복잡한 설계 의사결정에 심층 분석 추가

---

### Pattern 3: E2E Test Generation

```
mcp-playwright-integrator (scenario 1) → resume → (scenario 2) → resume → (scenario 3)
```

**용도**: 여러 E2E test scenario를 하나의 test suite로 생성

---

### Pattern 4: Parallel Expert + Sequential Thinking

```
                  ┌─ backend-expert
Alfred (parallel) ┼─ frontend-expert → Alfred (merge) → mcp-sequential-thinking-integrator → Final Decision
                  └─ devops-expert
```

**용도**: 여러 전문가 의견을 심층 분석으로 통합

---

## Configuration

### .claude/mcp.json

```json
{
  "mcpServers": {
    "context7": {
      "command": "npx",
      "args": ["-y", "@upleveled/mcp-context7"],
      "env": {}
    },
    "playwright": {
      "command": "npx",
      "args": ["-y", "@upleveled/mcp-playwright"],
      "env": {
        "BROWSER": "chromium"
      }
    },
    "sequential-thinking": {
      "command": "npx",
      "args": ["-y", "@upleveled/mcp-sequential-thinking"],
      "env": {}
    }
  }
}
```

---

### Auto-Setup (moai-adk init)

```bash
# Interactive selection
moai-adk init

# CLI selection
moai-adk init --with-mcp context7 --with-mcp playwright

# Auto-install all
moai-adk init --mcp-auto
```

---

## SessionManager Integration

### MCP Integrator Agent 등록

```python
# MCP agent 실행 및 등록
context7_result = Task(
    subagent_type="mcp-context7-integrator",
    prompt="Lookup FastAPI latest version"
)

register_agent(
    agent_name="mcp-context7-integrator",
    agent_id=context7_result["agent_id"],
    result=context7_result,
    chain_id="library-lookup"
)
```

---

### Resume Pattern for Playwright

```python
# Multiple scenarios with resume
playwright_id = None

for scenario in ["happy_path", "error_case", "edge_case"]:
    result = Task(
        subagent_type="mcp-playwright-integrator",
        prompt=f"E2E test for {scenario}",
        resume=playwright_id if playwright_id else None
    )

    if playwright_id is None:
        playwright_id = result["agent_id"]
    else:
        session_mgr.increment_resume_count(playwright_id)

    register_agent(
        agent_name="mcp-playwright-integrator",
        agent_id=result["agent_id"],
        result=result,
        chain_id="e2e-test-suite"
    )
```

---

## Best Practices

### ✅ DO

1. **MCP Integrator agents에 위임**
   ```python
   # 복잡한 작업은 integrator에게
   result = Task(subagent_type="mcp-context7-integrator", ...)
   ```

2. **Agent frontmatter에 MCP tools 명시**
   ```yaml
   tools: [..., mcp__context7__resolve-library-id, mcp__context7__get-library-docs]
   ```

3. **MCP 결과를 Alfred context에 저장**
   ```python
   alfred_context["library_docs"] = context7_result
   ```

4. **Resume로 연속 MCP 작업**
   ```python
   # Playwright multiple scenarios
   result = Task(..., resume=playwright_id)
   ```

---

### ❌ DON'T

1. **Agent가 MCP tool 없이 호출 시도**
   ```python
   # ❌ Agent tools 목록에 mcp__* 없는데 사용
   mcp__context7__resolve-library-id(...)  # Permission denied
   ```

2. **MCP 결과를 파일로 공유**
   ```python
   # ❌ MCP 결과를 파일에 저장하고 다른 agent가 읽기
   Write(".moai/temp/library-docs.json", context7_result)
   ```

3. **Sequential Thinking 과다 사용**
   ```python
   # ❌ 단순 조회에 Sequential Thinking 사용 (비효율)
   # Sequential Thinking은 복잡한 의사결정에만 사용
   ```

---

## Troubleshooting

### MCP Server Not Found

**증상**: `mcp__context7__*` tool이 작동하지 않음

**해결**:

```bash
# MCP server 설정 확인
cat .claude/mcp.json

# MCP server 재설치
npx -y @upleveled/mcp-context7 --version

# Claude Code 재시작
```

---

### Permission Denied

**증상**: Agent가 MCP tool 호출 시 permission denied

**해결**:

```yaml
# Agent frontmatter에 MCP tools 추가
---
tools: [..., mcp__context7__resolve-library-id]
---
```

---

### Resume Not Working for MCP Integrator

**증상**: Playwright integrator resume 시 context 손실

**해결**:

```python
# SessionManager로 agentId 추적 확인
resume_id = get_resume_id("mcp-playwright-integrator", chain_id="e2e-tests")

# Resume count 증가 확인
session_mgr.increment_resume_count(resume_id)
```

---

## 참고 자료

- **Alfred Orchestration**: `.moai/config/alfred-orchestration.yaml`
- **Agent Invocation**: `.moai/guidelines/agent-invocation.md`
- **SessionManager**: `src/moai_adk/core/session_manager.py`
- **MCP Servers**: https://modelcontextprotocol.io/servers
- **Context7**: https://github.com/upleveled/mcp-context7
- **Playwright MCP**: https://github.com/upleveled/mcp-playwright

---

**Last Updated**: 2025-11-12
**Version**: 1.0.0
