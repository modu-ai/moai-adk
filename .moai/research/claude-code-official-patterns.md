# Claude Code 공식 패턴: Agent 및 Skill 호출 가이드

**연구 대상**: Claude Code 공식 문서 (Anthropic)
**연구 날짜**: 2025-11-12
**목적**: MoAI-ADK v4.0 Enterprise 업그레이드를 위한 권위 있는 참조 자료

---

## 📋 Executive Summary

### 핵심 발견사항

1. **Agent 정의 방식**: `.claude/agents/*.md` 파일에 YAML frontmatter 사용
2. **Agent 호출 방식**: Claude Code SDK의 `query()` 함수 사용
3. **Skill 정의 방식**: `.claude/skills/skill-name/SKILL.md` 파일 구조
4. **Skill 호출 방식**: 명시적 `Skill("skill-name")` 호출 또는 자동 트리거
5. **MCP 통합**: `mcpServers` 설정으로 외부 도구 통합

### MoAI-ADK에 적용할 핵심 패턴

- **Agent**: YAML frontmatter + Markdown 지침 (현재 구조와 동일)
- **Skill**: Progressive Disclosure (3-level) + Context7 통합
- **호출**: 명시적 `Skill()` 호출 패턴 유지
- **MCP**: Context7, Playwright, Sequential-Thinking 서버 통합

---

## 🎯 Section 1: Agent 정의 및 호출 패턴

### 1.1 Agent 정의 (공식 패턴)

#### 파일 위치
```
.claude/agents/agent-name.md
```

#### Frontmatter 형식 (필수)

```yaml
---
name: backend-expert
description: |
  Backend API development expert. Use proactively when:
  - User mentions: "API", "endpoint", "REST", "GraphQL"
  - Files contain: FastAPI, Express, Django patterns
tools: [Read, Write, Bash, Grep, Glob, Edit]
model: sonnet
---

# Backend Expert Agent

You are a backend development specialist with expertise in:
- RESTful API design and implementation
- Database schema design and optimization
- Authentication and authorization patterns
- Error handling and logging strategies

## When to Activate

Automatically activate when:
- API endpoints need to be designed or modified
- Database integration is required
- Authentication logic is discussed
- Backend service architecture is being planned

## Tools Available

- **Read**: Analyze existing code
- **Write**: Create new files
- **Edit**: Modify existing code
- **Bash**: Run commands (npm, pytest, etc.)
- **Grep**: Search codebase patterns
- **Glob**: Find related files

## Best Practices

1. Always validate input data
2. Implement proper error handling
3. Use async/await for I/O operations
4. Write comprehensive tests
5. Document API endpoints clearly
```

#### 필수 필드

| 필드 | 타입 | 설명 | 예시 |
|------|------|------|------|
| `name` | string | Agent 고유 식별자 | `backend-expert` |
| `description` | string | 언제 사용할지 명확한 설명 | `Backend API development expert...` |
| `tools` | array | 허용된 도구 목록 | `[Read, Write, Bash]` |
| `model` | string | 사용할 모델 | `sonnet` 또는 `haiku` |

#### 선택 필드

- `disallowedTools`: 금지된 도구 목록
- `metadata`: 추가 메타데이터 (버전, 작성자 등)

### 1.2 Agent 호출 패턴 (TypeScript SDK)

#### Pattern 1: 직접 호출

```typescript
import { query } from "@anthropic-ai/claude-agent-sdk";

const response = query({
  prompt: "Design REST API for user management",
  options: {
    model: "claude-sonnet-4-5",
    workingDirectory: "/path/to/project",

    // Agents 정의
    agents: {
      "backend-expert": {
        description: "Backend API development expert",
        prompt: "You are a backend specialist. Focus on API design, database integration, and authentication.",
        tools: ["Read", "Write", "Edit", "Bash", "Grep"],
        model: "sonnet"
      }
    }
  }
});

// 응답 처리
for await (const message of response) {
  if (message.type === 'assistant') {
    console.log('Agent:', message.content);
  } else if (message.type === 'system' && message.subtype === 'subagent_start') {
    console.log(`Starting subagent: ${message.agent_name}`);
  }
}
```

#### Pattern 2: Session Resume (컨텍스트 상속)

```typescript
// 첫 번째 세션
let sessionId: string | undefined;

const initialResponse = query({
  prompt: "Analyze the authentication system",
  options: { model: "claude-sonnet-4-5" }
});

for await (const message of initialResponse) {
  if (message.type === 'system' && message.subtype === 'init') {
    sessionId = message.session_id;
  }
}

// 세션 재개 (컨텍스트 유지)
const resumedResponse = query({
  prompt: "Now add OAuth2 support",
  options: {
    resume: sessionId,  // 이전 컨텍스트 상속
    model: "claude-sonnet-4-5"
  }
});
```

#### Pattern 3: 병렬 실행 (여러 Agent 동시 호출)

```typescript
const response = query({
  prompt: "Review the entire application for security, performance, and test coverage",
  options: {
    model: "claude-sonnet-4-5",
    agents: {
      "security-reviewer": {
        description: "Security expert for vulnerability analysis",
        prompt: "Focus on authentication, authorization, SQL injection, XSS vulnerabilities",
        tools: ["Read", "Grep", "Glob"],
        model: "sonnet"
      },
      "performance-analyst": {
        description: "Performance optimization expert",
        prompt: "Analyze bottlenecks, memory leaks, optimization opportunities",
        tools: ["Read", "Grep", "Bash"],
        model: "sonnet"
      },
      "test-analyst": {
        description: "Testing and QA expert",
        prompt: "Evaluate test coverage, edge cases, integration scenarios",
        tools: ["Read", "Grep", "Write"],
        model: "haiku"
      }
    }
  }
});

// Agent들이 자동으로 병렬 실행됨
for await (const message of response) {
  if (message.type === 'system' && message.subtype === 'subagent_start') {
    console.log(`Starting: ${message.agent_name}`);
  } else if (message.type === 'system' && message.subtype === 'subagent_end') {
    console.log(`Completed: ${message.agent_name}`);
  }
}
```

### 1.3 Agent 권한 제어 (canUseTool)

```typescript
const response = query({
  prompt: "Deploy the application to production",
  options: {
    model: "claude-sonnet-4-5",
    permissionMode: "default",  // "acceptEdits", "default", "bypassPermissions"

    // 세밀한 권한 제어
    canUseTool: async (toolName, input) => {
      // Read-only 도구는 항상 허용
      if (['Read', 'Grep', 'Glob'].includes(toolName)) {
        return { behavior: "allow" };
      }

      // 위험한 명령어 차단
      if (toolName === 'Bash') {
        const dangerous = ['rm -rf', 'dd if=', 'mkfs', '> /dev/'];
        if (dangerous.some(pattern => input.command.includes(pattern))) {
          return {
            behavior: "deny",
            message: "Destructive command blocked for safety"
          };
        }
      }

      // 배포 명령은 확인 요청
      if (input.command?.includes('deploy') || input.command?.includes('kubectl apply')) {
        return {
          behavior: "ask",
          message: "Confirm deployment to production?"
        };
      }

      return { behavior: "allow" };
    }
  }
});
```

### 1.4 베스트 프랙티스

#### ✅ DO

1. **명확한 설명 작성**: `description`에 언제 사용할지 구체적으로 명시
2. **최소 권한 원칙**: 필요한 도구만 `tools`에 포함
3. **적절한 모델 선택**: 복잡한 작업은 `sonnet`, 간단한 작업은 `haiku`
4. **에러 핸들링**: `message.type === 'error'` 처리
5. **세션 ID 저장**: 컨텍스트 유지가 필요한 경우

#### ❌ DON'T

1. **과도한 권한 부여**: 모든 도구를 허용하지 않기
2. **모호한 설명**: "General purpose agent" 같은 설명 지양
3. **세션 관리 실패**: `sessionId` 없이 재개 시도
4. **무한 루프**: Agent가 서로를 무한히 호출하는 구조
5. **권한 검증 생략**: Production 환경에서 `bypassPermissions` 사용

---

## 🎓 Section 2: Skill 정의 및 호출 패턴

### 2.1 Skill 정의 (공식 패턴)

#### 파일 구조

```
.claude/skills/skill-name/
├── SKILL.md              (필수: 메인 스킬 정의)
├── examples.md           (선택: 실용 예제)
├── reference.md          (선택: 완전한 API 문서)
├── scripts/              (선택: 실행 가능한 스크립트)
│   ├── run_query.py
│   └── process_data.sh
├── references/           (선택: 참조 문서)
│   └── schema.md
└── assets/               (선택: 템플릿, 아이콘 등)
    └── template.html
```

#### SKILL.md Frontmatter 형식

```yaml
---
name: "moai-domain-backend"
version: "4.0.0"
description: |
  Backend architecture expertise. Use when:
  - API design required
  - Database integration needed
  - Authentication patterns required
keywords: [api, backend, database, rest, graphql]
primary_agent: backend-expert
secondary_agents: [database-expert, security-expert]
license: "MIT"
allowed-tools: ["Read", "Write", "Bash", "Grep"]
metadata:
  author: "MoAI Team"
  last_updated: "2025-11-12"
---

# Backend Development Expertise

**Domain Skill with AI-Powered API Design**

> **Primary Agent**: backend-expert
> **Version**: 4.0.0
> **Keywords**: api, backend, database, authentication

## 📖 Progressive Disclosure

### Level 1: Quick Reference (500 words max)

Backend development patterns for REST APIs, databases, and authentication.

#### Core Concepts

1. **RESTful API Design**
   - Resource-based URLs
   - HTTP methods (GET, POST, PUT, DELETE)
   - Status codes (200, 201, 400, 401, 404, 500)

2. **Database Integration**
   - ORM patterns (SQLAlchemy, Prisma)
   - Query optimization
   - Transaction management

3. **Authentication**
   - JWT tokens
   - OAuth2 flows
   - Session management

#### Quick Start

```python
# FastAPI example
from fastapi import FastAPI, Depends
from sqlalchemy.orm import Session

app = FastAPI()

@app.get("/users/{user_id}")
async def get_user(user_id: int, db: Session = Depends(get_db)):
    user = db.query(User).filter(User.id == user_id).first()
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    return user
```

### Level 2: Practical Implementation (1,500 words max)

Detailed patterns with complete code examples.

#### Pattern 1: RESTful CRUD API

```python
# FastAPI CRUD operations
from fastapi import FastAPI, HTTPException, Depends
from sqlalchemy.orm import Session
from pydantic import BaseModel
from typing import List

app = FastAPI()

class UserCreate(BaseModel):
    email: str
    username: str
    password: str

class UserResponse(BaseModel):
    id: int
    email: str
    username: str

    class Config:
        from_attributes = True

# CREATE
@app.post("/users/", response_model=UserResponse, status_code=201)
async def create_user(user: UserCreate, db: Session = Depends(get_db)):
    # Hash password
    hashed_password = hash_password(user.password)

    # Check if user exists
    existing = db.query(User).filter(User.email == user.email).first()
    if existing:
        raise HTTPException(status_code=400, detail="Email already registered")

    # Create user
    db_user = User(
        email=user.email,
        username=user.username,
        hashed_password=hashed_password
    )
    db.add(db_user)
    db.commit()
    db.refresh(db_user)

    return db_user

# READ
@app.get("/users/{user_id}", response_model=UserResponse)
async def get_user(user_id: int, db: Session = Depends(get_db)):
    user = db.query(User).filter(User.id == user_id).first()
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    return user

# UPDATE
@app.put("/users/{user_id}", response_model=UserResponse)
async def update_user(
    user_id: int,
    user_update: UserCreate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    # Authorization check
    if current_user.id != user_id:
        raise HTTPException(status_code=403, detail="Not authorized")

    user = db.query(User).filter(User.id == user_id).first()
    if not user:
        raise HTTPException(status_code=404, detail="User not found")

    # Update fields
    user.email = user_update.email
    user.username = user_update.username
    db.commit()
    db.refresh(user)

    return user

# DELETE
@app.delete("/users/{user_id}", status_code=204)
async def delete_user(
    user_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    # Authorization check
    if current_user.id != user_id:
        raise HTTPException(status_code=403, detail="Not authorized")

    user = db.query(User).filter(User.id == user_id).first()
    if not user:
        raise HTTPException(status_code=404, detail="User not found")

    db.delete(user)
    db.commit()
```

#### Pattern 2: JWT Authentication

```python
from fastapi import Depends, HTTPException, status
from fastapi.security import OAuth2PasswordBearer, OAuth2PasswordRequestForm
from jose import JWTError, jwt
from passlib.context import CryptContext
from datetime import datetime, timedelta

# Configuration
SECRET_KEY = "your-secret-key-here"
ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE_MINUTES = 30

pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")
oauth2_scheme = OAuth2PasswordBearer(tokenUrl="token")

# Password hashing
def hash_password(password: str) -> str:
    return pwd_context.hash(password)

def verify_password(plain_password: str, hashed_password: str) -> bool:
    return pwd_context.verify(plain_password, hashed_password)

# Token creation
def create_access_token(data: dict, expires_delta: timedelta = None):
    to_encode = data.copy()
    if expires_delta:
        expire = datetime.utcnow() + expires_delta
    else:
        expire = datetime.utcnow() + timedelta(minutes=15)

    to_encode.update({"exp": expire})
    encoded_jwt = jwt.encode(to_encode, SECRET_KEY, algorithm=ALGORITHM)
    return encoded_jwt

# Token validation
async def get_current_user(
    token: str = Depends(oauth2_scheme),
    db: Session = Depends(get_db)
):
    credentials_exception = HTTPException(
        status_code=status.HTTP_401_UNAUTHORIZED,
        detail="Could not validate credentials",
        headers={"WWW-Authenticate": "Bearer"},
    )

    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        user_id: int = payload.get("sub")
        if user_id is None:
            raise credentials_exception
    except JWTError:
        raise credentials_exception

    user = db.query(User).filter(User.id == user_id).first()
    if user is None:
        raise credentials_exception

    return user

# Login endpoint
@app.post("/token")
async def login(
    form_data: OAuth2PasswordRequestForm = Depends(),
    db: Session = Depends(get_db)
):
    user = db.query(User).filter(User.email == form_data.username).first()
    if not user or not verify_password(form_data.password, user.hashed_password):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Incorrect email or password",
            headers={"WWW-Authenticate": "Bearer"},
        )

    access_token_expires = timedelta(minutes=ACCESS_TOKEN_EXPIRE_MINUTES)
    access_token = create_access_token(
        data={"sub": str(user.id)},
        expires_delta=access_token_expires
    )

    return {"access_token": access_token, "token_type": "bearer"}

# Protected endpoint
@app.get("/users/me", response_model=UserResponse)
async def read_users_me(current_user: User = Depends(get_current_user)):
    return current_user
```

### Level 3: Advanced Patterns (3,000 words max)

Complete API reference with advanced use cases.

#### Advanced Pattern 1: Database Query Optimization

```python
from sqlalchemy import select, func
from sqlalchemy.orm import selectinload, joinedload

# N+1 query problem
# BAD: Triggers N additional queries
@app.get("/users/with-posts")
async def get_users_bad(db: Session = Depends(get_db)):
    users = db.query(User).all()
    # Each user.posts access triggers a new query
    return [{"user": user, "post_count": len(user.posts)} for user in users]

# GOOD: Single query with join
@app.get("/users/with-posts-optimized")
async def get_users_good(db: Session = Depends(get_db)):
    users = db.query(User)\
        .options(selectinload(User.posts))\
        .all()
    return [{"user": user, "post_count": len(user.posts)} for user in users]

# Complex query with aggregation
@app.get("/user-stats")
async def get_user_stats(db: Session = Depends(get_db)):
    stats = db.query(
        User.id,
        User.username,
        func.count(Post.id).label("post_count"),
        func.avg(Post.views).label("avg_views")
    )\
    .join(Post, User.id == Post.user_id)\
    .group_by(User.id, User.username)\
    .having(func.count(Post.id) > 5)\
    .all()

    return [
        {
            "user_id": stat.id,
            "username": stat.username,
            "total_posts": stat.post_count,
            "average_views": round(stat.avg_views, 2)
        }
        for stat in stats
    ]
```

## 🎯 Context7 MCP Integration

### Setup

```json
// .claude/mcp.json
{
  "mcpServers": {
    "context7": {
      "command": "npx",
      "args": ["-y", "@context7/mcp-server"]
    }
  }
}
```

### Usage in Agent

```typescript
const response = query({
  prompt: "Find best practices for FastAPI authentication",
  options: {
    mcpServers: {
      "context7": {
        command: "npx",
        args: ["-y", "@context7/mcp-server"]
      }
    },
    allowedTools: [
      "mcp__context7__resolve-library-id",
      "mcp__context7__get-library-docs"
    ],
    model: "claude-sonnet-4-5"
  }
});
```

### Agent가 Skill 로드 시

```markdown
# Agent 내부 로직
1. Detect need: Task mentions "FastAPI authentication"
2. Load skill: Skill("moai-domain-backend")
3. Context7 lookup:
   - resolve_library_id("fastapi")
   - get_library_docs(libraryID, topic="authentication", tokens=10000)
4. Apply guidance: Follow retrieved best practices
5. Generate code: Implement with official patterns
```

## 📚 Official References

- [FastAPI Documentation](https://fastapi.tiangolo.com/)
- [SQLAlchemy ORM](https://docs.sqlalchemy.org/en/14/orm/)
- [JWT.io](https://jwt.io/)
- [OAuth2 RFC 6749](https://datatracker.ietf.org/doc/html/rfc6749)
```

### 2.2 Skill 호출 패턴

#### Pattern 1: 명시적 호출 (Agent에서)

```markdown
# Agent 내부 지침
When you need backend expertise:

1. **Detect requirement**: User mentions API, database, or authentication
2. **Load skill**: Skill("moai-domain-backend")
3. **Read content**: Parse SKILL.md frontmatter and content
4. **Apply guidance**: Follow patterns and best practices
5. **Reference examples**: If needed, load examples.md or reference.md
6. **Generate code**: Implement based on skill guidance

Example:
```python
# Agent detects: "Create a REST API for user management"
# Agent executes: Skill("moai-domain-backend")
# Agent reads: SKILL.md content
# Agent applies: RESTful patterns from Level 2
# Agent generates: FastAPI code with proper error handling
```
```

#### Pattern 2: 자동 트리거 (Keyword 기반)

```yaml
# SKILL.md frontmatter
keywords: [api, backend, database, rest, graphql, fastapi, express]

# Claude Code가 자동으로 트리거하는 경우:
# - User prompt에 keywords 포함
# - 파일 경로에 keywords 관련 패턴 (src/api/, server.py)
# - 이전 대화 컨텍스트에서 관련 주제 언급
```

#### Pattern 3: Progressive Disclosure

```markdown
# Level 1: Quick Reference (항상 로드)
- 500 words 이하
- 핵심 개념만
- 빠른 참조용

# Level 2: Practical Implementation (필요 시 로드)
- 1,500 words 이하
- 완전한 코드 예제
- 실용적인 패턴

# Level 3: Advanced Patterns (명시적 요청 시만 로드)
- 3,000 words 이하
- 고급 사용 사례
- 완전한 API 레퍼런스
```

### 2.3 Skill 초기화 및 패키징

#### 새 Skill 생성

```bash
# Skill 초기화 스크립트 사용
python scripts/init_skill.py bigquery-helper --path ./skills

# 생성되는 구조:
# skills/bigquery-helper/
#   ├── SKILL.md (TODO 플레이스홀더 포함)
#   ├── scripts/ (예제 스크립트)
#   ├── references/ (예제 레퍼런스)
#   └── assets/ (예제 에셋)
```

#### Skill 검증 및 패키징

```bash
# 구조 검증
python scripts/quick_validate.py path/to/my-skill

# 배포 가능한 zip 생성
python scripts/package_skill.py path/to/my-skill ./dist
# 출력: dist/my-skill.zip
```

### 2.4 베스트 프랙티스

#### ✅ DO

1. **명확한 Frontmatter**: 모든 필수 필드 작성
2. **Progressive Disclosure**: 3단계 구조 준수
3. **실행 가능한 예제**: 복사-붙여넣기 가능한 코드
4. **공식 문서 링크**: 권위 있는 출처 참조
5. **Context7 통합**: AI-powered documentation lookup

#### ❌ DON'T

1. **과도한 내용**: Level 1에 3,000 words 작성
2. **추상적 설명**: 구체적인 코드 없이 이론만
3. **깨진 예제**: 테스트하지 않은 코드
4. **오래된 정보**: 현재 버전과 맞지 않는 내용
5. **주관적 의견**: 공식 패턴 없이 개인 선호도 강요

---

## 🔧 Section 3: MCP 통합 패턴

### 3.1 MCP Server 설정

#### .claude/mcp.json 구조

```json
{
  "mcpServers": {
    "context7": {
      "command": "npx",
      "args": ["-y", "@context7/mcp-server"]
    },
    "playwright": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-playwright"]
    },
    "sequential-thinking": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-sequential-thinking"]
    },
    "github": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-server-github"],
      "oauth": {
        "clientId": "your-client-id",
        "clientSecret": "your-client-secret",
        "scopes": ["repo", "issues"]
      }
    },
    "filesystem": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-filesystem",
        "/path/to/allowed/files"
      ]
    }
  }
}
```

### 3.2 Agent에서 MCP Tools 사용

#### Agent 정의에 MCP Tools 포함

```yaml
---
name: research-agent
description: Research and documentation expert using Context7
tools:
  - Read
  - Write
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
model: sonnet
---

# Research Agent

You use Context7 MCP to research official documentation.

## Workflow

1. **Resolve Library ID**:
   ```typescript
   const libId = await resolve_library_id("fastapi");
   ```

2. **Get Documentation**:
   ```typescript
   const docs = await get_library_docs({
     context7CompatibleLibraryID: libId,
     topic: "authentication",
     tokens: 10000
   });
   ```

3. **Apply to Task**: Use retrieved docs to solve user's problem
```

#### TypeScript SDK에서 MCP 사용

```typescript
import { query } from "@anthropic-ai/claude-agent-sdk";

const response = query({
  prompt: "Research FastAPI best practices and create authentication system",
  options: {
    model: "claude-sonnet-4-5",

    // MCP 서버 설정
    mcpServers: {
      "context7": {
        command: "npx",
        args: ["-y", "@context7/mcp-server"]
      }
    },

    // 허용된 MCP 도구
    allowedTools: [
      "Read",
      "Write",
      "Edit",
      "mcp__context7__resolve-library-id",
      "mcp__context7__get-library-docs"
    ]
  }
});

for await (const message of response) {
  if (message.type === 'tool_call' && message.tool_name.startsWith('mcp__')) {
    console.log(`MCP Tool: ${message.tool_name}`);
    console.log(`Input:`, message.input);
  } else if (message.type === 'assistant') {
    console.log('Agent:', message.content);
  }
}
```

### 3.3 Custom MCP Tools 생성

```typescript
import { createSdkMcpServer, tool } from "@anthropic-ai/claude-agent-sdk";
import { z } from "zod";

// Custom notification tool
const appTools = createSdkMcpServer({
  name: "app-services",
  version: "1.0.0",
  tools: [
    tool(
      "send_notification",
      "Send notification to users",
      {
        userId: z.string(),
        message: z.string(),
        priority: z.enum(["low", "medium", "high"]).default("medium")
      },
      async (args) => {
        // Integration logic
        await notificationService.send(args);
        return {
          content: [{ type: "text", text: "Notification sent" }]
        };
      }
    ),

    tool(
      "log_event",
      "Log application events",
      {
        event: z.string(),
        data: z.record(z.any()).optional(),
        severity: z.enum(["info", "warning", "error"]).default("info")
      },
      async (args) => {
        logger.log(args.severity, args.event, args.data);
        return {
          content: [{ type: "text", text: "Event logged" }]
        };
      }
    )
  ]
});

// Agent에서 사용
const response = query({
  prompt: "Monitor system and send alerts",
  options: {
    mcpServers: {
      "app-services": appTools
    },
    allowedTools: [
      "mcp__app-services__send_notification",
      "mcp__app-services__log_event"
    ]
  }
});
```

### 3.4 Context7 MCP 사용 패턴

#### Pattern 1: Library 리서치

```typescript
// Step 1: Resolve library ID
const libId = await resolve_library_id("fastapi");
// Result: "/tiangolo/fastapi"

// Step 2: Get specific documentation
const docs = await get_library_docs({
  context7CompatibleLibraryID: "/tiangolo/fastapi",
  topic: "authentication security middleware",
  tokens: 15000
});

// Step 3: Apply to implementation
// Agent uses docs to generate secure authentication code
```

#### Pattern 2: 캐싱 전략

```typescript
const cache: Record<string, any> = {};

async function getCachedDocs(libraryName: string, topic: string) {
  const cacheKey = `${libraryName}:${topic}`;

  if (cache[cacheKey]) {
    console.log('Cache hit:', cacheKey);
    return cache[cacheKey];
  }

  console.log('Cache miss:', cacheKey);
  const libId = await resolve_library_id(libraryName);
  const docs = await get_library_docs({
    context7CompatibleLibraryID: libId,
    topic,
    tokens: 10000
  });

  cache[cacheKey] = docs;
  return docs;
}
```

#### Pattern 3: 에러 핸들링

```typescript
async function safeResearchLibrary(libraryName: string) {
  try {
    const libId = await resolve_library_id(libraryName);

    if (!libId) {
      console.log(`No library found for: ${libraryName}`);
      return null;
    }

    const docs = await get_library_docs({
      context7CompatibleLibraryID: libId,
      tokens: 5000
    });

    return docs;

  } catch (error) {
    if (error.code === 'RATE_LIMIT_EXCEEDED') {
      console.log('Rate limit exceeded, retry after delay');
      await delay(5000);
      return safeResearchLibrary(libraryName);
    } else if (error.code === 'LIBRARY_NOT_FOUND') {
      console.log(`Library not in Context7: ${libraryName}`);
      return null;
    } else {
      console.error('Research failed:', error);
      throw error;
    }
  }
}
```

### 3.5 베스트 프랙티스

#### ✅ DO

1. **캐싱 구현**: 동일한 문서를 반복 요청하지 않기
2. **에러 핸들링**: Rate limit, timeout 처리
3. **적절한 토큰 수**: 필요한 만큼만 요청 (5,000-15,000)
4. **명확한 topic**: 구체적인 주제로 검색 정확도 향상
5. **Fallback 전략**: Context7 실패 시 대체 방법

#### ❌ DON'T

1. **과도한 요청**: 매 쿼리마다 resolve_library_id 호출
2. **토큰 낭비**: 50,000 tokens로 간단한 주제 검색
3. **모호한 topic**: "general programming" 같은 광범위한 주제
4. **에러 무시**: 실패 시 재시도 없이 진행
5. **캐시 없음**: 동일 문서를 매번 새로 가져오기

---

## 🚀 Section 4: v4.0 Enterprise 업그레이드 가이드

### 4.1 v2.0 → v4.0 Enterprise 변환 체크리스트

#### 필수 추가 섹션

- [ ] **Context7 MCP Integration**: 해당 도메인의 주요 라이브러리 리서치 방법
- [ ] **AI-Powered Features**: Context7을 활용한 동적 문서 검색
- [ ] **Predictive Optimization**: 일반적인 실수 패턴과 예방 방법
- [ ] **Advanced Code Examples**: 10+ 실행 가능한 코드 예제

#### 구조 요구사항

- [ ] Progressive Disclosure 3단계 구현
- [ ] Frontmatter에 `version: "4.0.0"` 명시
- [ ] `primary_agent` 및 `secondary_agents` 정의
- [ ] `keywords` 배열로 자동 트리거 설정
- [ ] Official references 링크 포함

### 4.2 v4.0 Enterprise 템플릿

```yaml
---
name: "moai-skill-name"
version: "4.0.0"
description: |
  [Domain] expertise with AI-powered [feature].

  Use when:
  - [Trigger condition 1]
  - [Trigger condition 2]
  - [Trigger condition 3]
keywords: [keyword1, keyword2, keyword3, keyword4]
primary_agent: agent-name
secondary_agents: [agent2, agent3]
license: "MIT"
allowed-tools: ["Read", "Write", "Edit", "Bash", "Grep"]
metadata:
  author: "MoAI Team"
  category: "domain"  # domain, language, baas, essentials
  last_updated: "2025-11-12"
---

# [Skill Name]

**[Domain] with AI-Powered [Feature]**

> **Primary Agent**: agent-name
> **Version**: 4.0.0
> **Keywords**: keyword1, keyword2, keyword3

## 📖 Progressive Disclosure

### Level 1: Quick Reference (500 words max)

Brief overview of core concepts and quick start guide.

#### Core Concepts

1. **Concept 1**: Brief explanation
2. **Concept 2**: Brief explanation
3. **Concept 3**: Brief explanation

#### Quick Start

```[language]
# Minimal working example
[code snippet]
```

### Level 2: Practical Implementation (1,500 words max)

Complete, production-ready code examples.

#### Pattern 1: [Pattern Name]

```[language]
# Complete implementation
[full code example with comments]
```

#### Pattern 2: [Pattern Name]

```[language]
# Alternative approach
[full code example with comments]
```

### Level 3: Advanced Patterns (3,000 words max)

Advanced use cases and optimization techniques.

#### Advanced Pattern 1: [Pattern Name]

```[language]
# Complex scenario
[advanced code example]
```

## 🎯 Context7 MCP Integration

### Official Libraries

```typescript
// Primary library
const libId = await resolve_library_id("[library-name]");
const docs = await get_library_docs({
  context7CompatibleLibraryID: libId,
  topic: "[specific-topic]",
  tokens: 10000
});
```

### Research Workflow

1. **Identify need**: Detect [library-name] usage
2. **Resolve ID**: Get Context7 library identifier
3. **Fetch docs**: Retrieve relevant documentation
4. **Apply patterns**: Implement based on official guidance
5. **Validate**: Compare with retrieved best practices

## 🤖 AI-Powered Features

### Predictive Optimization

Common mistakes and prevention:

1. **Anti-pattern 1**: [Description]
   - **Problem**: [Why it's bad]
   - **Solution**: [Correct approach]
   - **Example**:
     ```[language]
     # BAD
     [bad code]

     # GOOD
     [good code]
     ```

2. **Anti-pattern 2**: [Description]
   - **Problem**: [Why it's bad]
   - **Solution**: [Correct approach]
   - **Example**:
     ```[language]
     # BAD
     [bad code]

     # GOOD
     [good code]
     ```

### Intelligent Analysis

Agent automatically checks for:
- [ ] Error handling completeness
- [ ] Security vulnerability patterns
- [ ] Performance bottlenecks
- [ ] Code quality issues
- [ ] Best practice adherence

## 📚 Official References

- [Primary Documentation]([url])
- [API Reference]([url])
- [Tutorial]([url])
- [Community Resources]([url])

## 🏷️ Tags

`[tag1]` `[tag2]` `[tag3]` `[tag4]`

---

**Last Updated**: 2025-11-12
**Version**: 4.0.0 (Enterprise)
**Maintainer**: MoAI Team
```

### 4.3 Quality Checklist

#### Documentation Quality

- [ ] All code examples tested and working
- [ ] Progressive Disclosure properly implemented
- [ ] Context7 integration examples included
- [ ] Official references linked
- [ ] Anti-patterns documented with solutions

#### Completeness

- [ ] 10+ code examples across 3 levels
- [ ] Primary and secondary agents defined
- [ ] Keywords trigger conditions specified
- [ ] Allowed tools list complete
- [ ] License specified

#### AI Integration

- [ ] Context7 MCP usage pattern documented
- [ ] Caching strategy implemented
- [ ] Error handling for MCP failures
- [ ] Fallback documentation sources
- [ ] Research workflow clearly defined

---

## 📊 Section 5: 실전 예제

### 5.1 Complete Agent + Skill 통합 예제

#### Agent 정의 (.claude/agents/backend-expert.md)

```yaml
---
name: backend-expert
description: |
  Backend API development expert using FastAPI.

  Activate when:
  - User mentions: "API", "endpoint", "REST", "authentication"
  - Files: server.py, main.py, routers/*.py
  - Tasks: API design, database integration, auth implementation
tools: [Read, Write, Edit, Bash, Grep, Glob, mcp__context7__resolve-library-id, mcp__context7__get-library-docs]
model: sonnet
---

# Backend Development Expert

You are a backend API specialist using FastAPI and SQLAlchemy.

## Workflow

1. **Understand requirement**: Parse user's API specification
2. **Research best practices**: Use Context7 to fetch FastAPI patterns
   ```typescript
   const libId = await resolve_library_id("fastapi");
   const docs = await get_library_docs({
     context7CompatibleLibraryID: libId,
     topic: user_specified_topic,
     tokens: 10000
   });
   ```
3. **Load skill**: Skill("moai-domain-backend")
4. **Implement**: Generate production-ready code
5. **Validate**: Check against official patterns

## Code Standards

- Use Pydantic for validation
- Implement proper error handling
- Add type hints
- Write docstrings
- Follow REST conventions

## Tools Usage

- **Read**: Analyze existing code
- **Write**: Create new modules
- **Edit**: Modify existing files
- **Bash**: Run tests (pytest)
- **Grep**: Find patterns
- **Context7**: Research official docs
```

#### Skill 정의 (.claude/skills/moai-domain-backend/SKILL.md)

[Section 2.1의 전체 예제 참조]

#### TypeScript 호출 코드

```typescript
import { query } from "@anthropic-ai/claude-agent-sdk";

async function buildBackendAPI() {
  const response = query({
    prompt: "Create a FastAPI REST API for user management with JWT authentication",
    options: {
      model: "claude-sonnet-4-5",
      workingDirectory: "/Users/developer/projects/my-api",

      // Settings 로드 (CLAUDE.md, skills, etc.)
      settingSources: ["user", "project", "local"],

      // MCP 서버 설정
      mcpServers: {
        "context7": {
          command: "npx",
          args: ["-y", "@context7/mcp-server"]
        }
      },

      // Agents 정의
      agents: {
        "backend-expert": {
          description: "Backend API development expert",
          prompt: `You are a FastAPI specialist. Use Context7 to research official patterns.

          Workflow:
          1. Research FastAPI patterns using Context7
          2. Load Skill("moai-domain-backend")
          3. Implement following official best practices
          4. Generate tests`,
          tools: [
            "Read", "Write", "Edit", "Bash", "Grep",
            "mcp__context7__resolve-library-id",
            "mcp__context7__get-library-docs"
          ],
          model: "sonnet"
        }
      },

      // 권한 설정
      permissionMode: "acceptEdits",

      // 예산 제한
      maxBudgetUsd: 5.0
    }
  });

  let sessionId: string | undefined;

  for await (const message of response) {
    switch (message.type) {
      case 'system':
        if (message.subtype === 'init') {
          sessionId = message.session_id;
          console.log(`Session: ${sessionId}`);
          console.log(`Skills: ${message.skills?.join(', ')}`);
        }
        break;

      case 'assistant':
        console.log('Agent:', message.content);
        break;

      case 'tool_call':
        console.log(`Executing: ${message.tool_name}`);
        if (message.tool_name.startsWith('mcp__context7__')) {
          console.log('Context7 research:', message.input);
        }
        break;

      case 'error':
        console.error('Error:', message.error);
        break;
    }
  }

  return sessionId;
}

// 실행
buildBackendAPI().catch(console.error);
```

### 5.2 실행 결과 예상 플로우

```
1. Session started: cc-session-abc123
   Skills: moai-domain-backend, moai-foundation-tdd, moai-alfred-best-practices

2. Agent: I'll create a FastAPI REST API with JWT authentication.
   First, let me research official FastAPI patterns...

3. Executing: mcp__context7__resolve-library-id
   Input: { libraryName: "fastapi" }

4. Tool result: /tiangolo/fastapi

5. Executing: mcp__context7__get-library-docs
   Input: {
     context7CompatibleLibraryID: "/tiangolo/fastapi",
     topic: "authentication jwt security",
     tokens: 10000
   }

6. Tool result: [FastAPI authentication documentation]

7. Agent: Loading backend expertise...
   Executing: Skill("moai-domain-backend")

8. Agent: Based on official patterns, I'll implement:
   - User model with SQLAlchemy
   - JWT token generation and validation
   - OAuth2 password bearer authentication
   - Protected endpoints with dependency injection

9. Executing: Write
   File: src/main.py
   [Creates FastAPI app with configuration]

10. Executing: Write
    File: src/models.py
    [Creates User model with SQLAlchemy]

11. Executing: Write
    File: src/auth.py
    [Creates JWT authentication logic]

12. Executing: Write
    File: src/routers/users.py
    [Creates user CRUD endpoints]

13. Executing: Write
    File: tests/test_auth.py
    [Creates authentication tests]

14. Executing: Bash
    Command: pytest tests/
    [Runs tests to validate implementation]

15. Agent: Implementation complete!
    - ✅ User model with password hashing
    - ✅ JWT token authentication
    - ✅ CRUD endpoints with authorization
    - ✅ Comprehensive tests (95% coverage)

    All tests passing. Ready for deployment.
```

---

## 🎓 Section 6: MoAI-ADK 적용 가이드

### 6.1 현재 MoAI-ADK 구조 분석

#### 현재 상태

```
.claude/
├── agents/           # 29 agents (✅ 공식 패턴 일치)
├── commands/         # 명령어 계층
├── skills/           # 119 skills
│   ├── v4.0.0/      # 38 skills (Domain, Language, BaaS, Essentials)
│   ├── v2.0.0/      # 17 skills (Alfred Core - 업그레이드 필요)
│   └── v1.0.0/      # Multiple skills (업그레이드 필요)
├── hooks/            # 라이프사이클 훅
└── mcp.json          # MCP 서버 설정 (✅ 공식 패턴 일치)
```

#### 공식 패턴과의 비교

| 항목 | MoAI-ADK 현재 | Claude Code 공식 | 호환성 |
|------|---------------|-----------------|--------|
| Agent 정의 | YAML frontmatter + Markdown | YAML frontmatter + Markdown | ✅ 완전 일치 |
| Skill 정의 | YAML frontmatter + Markdown | YAML frontmatter + Markdown | ✅ 완전 일치 |
| Agent 호출 | `Task(subagent_type="...")` | `query({ agents: {...} })` | ⚠️ 구문 다름 |
| Skill 호출 | `Skill("skill-name")` | `Skill("skill-name")` | ✅ 완전 일치 |
| MCP 통합 | `.claude/mcp.json` | `.claude/mcp.json` | ✅ 완전 일치 |
| Progressive Disclosure | 일부만 구현 | 3-level 표준 | ⚠️ 업그레이드 필요 |

### 6.2 업그레이드 전략

#### Phase 1: v2.0 Alfred Core Skills → v4.0 Enterprise

**대상**: 17 skills (moai-alfred-* 시리즈)

**작업 항목**:
1. Frontmatter에 `version: "4.0.0"` 추가
2. Progressive Disclosure 3-level 구조 구현
3. Context7 MCP 통합 섹션 추가
4. AI-Powered Features 섹션 추가
5. 10+ 코드 예제 추가
6. Official References 링크 추가

**우선순위**:
- High: moai-alfred-agent-guide, moai-alfred-best-practices
- Medium: moai-alfred-ask-user-questions, moai-alfred-personas
- Low: 나머지 Alfred 지원 skills

#### Phase 2: v1.0 Skills → v4.0 Enterprise

**대상**: Multiple NEW skills

**작업 항목**:
1. 모든 v4.0 Enterprise 요구사항 적용
2. Primary/Secondary agents 정의
3. Keywords 자동 트리거 설정
4. Context7 통합 (해당 도메인 라이브러리)

#### Phase 3: Agent 호출 패턴 표준화

**현재 문제점**:
```python
# MoAI-ADK 현재 방식 (Claude Code SDK와 다름)
Task(subagent_type="backend-expert", prompt="...", ...)
```

**해결 방안**:
1. **옵션 A**: MoAI-ADK의 `Task()` 함수를 Claude Code SDK 호환 래퍼로 유지
2. **옵션 B**: Claude Code SDK의 `query()` 패턴으로 전환
3. **옵션 C**: 하이브리드 접근 (내부적으로 SDK 사용, 외부 API는 유지)

**권장**: **옵션 A** - 기존 코드 호환성 유지하면서 내부적으로 Claude Code SDK 패턴 준수

### 6.3 구현 로드맵

#### Week 1-2: v4.0 Enterprise 템플릿 확정

- [ ] 템플릿 파일 생성 (`.claude/templates/skill-v4.0.0.md`)
- [ ] 검증 스크립트 업데이트 (`scripts/validate_skill.py`)
- [ ] 패키징 스크립트 업데이트 (`scripts/package_skill.py`)
- [ ] 문서화 (CONTRIBUTING.md 업데이트)

#### Week 3-4: Alfred Core Skills 업그레이드

- [ ] moai-alfred-agent-guide → v4.0
- [ ] moai-alfred-best-practices → v4.0
- [ ] moai-alfred-ask-user-questions → v4.0
- [ ] 나머지 14 skills → v4.0

#### Week 5-6: v1.0 Skills 업그레이드

- [ ] 모든 NEW skills 리스트업
- [ ] 우선순위 결정 (사용 빈도 기준)
- [ ] 순차 업그레이드 (주당 5-7 skills)

#### Week 7-8: Agent 호출 패턴 표준화

- [ ] `Task()` 래퍼 함수 재설계
- [ ] 모든 agents에 호환성 테스트
- [ ] 문서 업데이트 (moai-alfred-agent-guide)
- [ ] 예제 코드 업데이트

---

## 📝 Section 7: 체크리스트 및 검증 도구

### 7.1 v4.0 Enterprise Skill 검증 체크리스트

```yaml
# skill-v4.0-validation.yml

required_frontmatter:
  - name: string (kebab-case)
  - version: "4.0.0"
  - description: string (multi-line with triggers)
  - keywords: array (min 3, max 10)
  - primary_agent: string
  - license: "MIT"

optional_frontmatter:
  - secondary_agents: array
  - allowed-tools: array
  - metadata: object

required_sections:
  - "Progressive Disclosure"
  - "Level 1: Quick Reference"
  - "Level 2: Practical Implementation"
  - "Level 3: Advanced Patterns"
  - "Context7 MCP Integration"
  - "AI-Powered Features"
  - "Official References"

content_requirements:
  level_1_max_words: 500
  level_2_max_words: 1500
  level_3_max_words: 3000
  min_code_examples: 10
  min_official_references: 3

quality_checks:
  - all_code_examples_tested: true
  - official_references_valid: true
  - context7_integration_present: true
  - anti_patterns_documented: true
  - primary_agent_defined: true
```

### 7.2 자동 검증 스크립트

```python
#!/usr/bin/env python3
"""
Validate v4.0 Enterprise Skill compliance
"""

import yaml
import re
from pathlib import Path
from typing import Dict, List, Tuple

def validate_skill_v4(skill_path: Path) -> Tuple[bool, List[str]]:
    """
    Validate skill against v4.0 Enterprise requirements.

    Returns:
        (is_valid, error_messages)
    """
    errors = []

    # Read SKILL.md
    skill_md = skill_path / "SKILL.md"
    if not skill_md.exists():
        return False, ["SKILL.md not found"]

    content = skill_md.read_text()

    # Parse frontmatter
    match = re.match(r'^---\n(.*?)\n---', content, re.DOTALL)
    if not match:
        errors.append("Invalid YAML frontmatter")
        return False, errors

    try:
        frontmatter = yaml.safe_load(match.group(1))
    except yaml.YAMLError as e:
        errors.append(f"YAML parse error: {e}")
        return False, errors

    # Validate required fields
    required = ['name', 'version', 'description', 'keywords', 'primary_agent', 'license']
    for field in required:
        if field not in frontmatter:
            errors.append(f"Missing required field: {field}")

    # Validate version
    if frontmatter.get('version') != '4.0.0':
        errors.append(f"Version must be 4.0.0, got {frontmatter.get('version')}")

    # Validate keywords
    keywords = frontmatter.get('keywords', [])
    if not isinstance(keywords, list):
        errors.append("keywords must be an array")
    elif len(keywords) < 3:
        errors.append("keywords must have at least 3 items")
    elif len(keywords) > 10:
        errors.append("keywords must have at most 10 items")

    # Validate sections
    required_sections = [
        "Progressive Disclosure",
        "Level 1: Quick Reference",
        "Level 2: Practical Implementation",
        "Level 3: Advanced Patterns",
        "Context7 MCP Integration",
        "AI-Powered Features",
        "Official References"
    ]

    for section in required_sections:
        if section not in content:
            errors.append(f"Missing required section: {section}")

    # Count code examples
    code_blocks = re.findall(r'```[\w]*\n.*?```', content, re.DOTALL)
    if len(code_blocks) < 10:
        errors.append(f"Must have at least 10 code examples, found {len(code_blocks)}")

    # Validate word counts
    levels = {
        'Level 1': (500, re.search(r'### Level 1:.*?(?=###|$)', content, re.DOTALL)),
        'Level 2': (1500, re.search(r'### Level 2:.*?(?=###|$)', content, re.DOTALL)),
        'Level 3': (3000, re.search(r'### Level 3:.*?(?=###|$)', content, re.DOTALL))
    }

    for level_name, (max_words, match) in levels.items():
        if match:
            text = match.group(0)
            word_count = len(text.split())
            if word_count > max_words:
                errors.append(f"{level_name} exceeds {max_words} words ({word_count} found)")

    return len(errors) == 0, errors


def main():
    import sys

    if len(sys.argv) != 2:
        print("Usage: validate_skill_v4.py <skill-directory>")
        sys.exit(1)

    skill_path = Path(sys.argv[1])
    is_valid, errors = validate_skill_v4(skill_path)

    if is_valid:
        print(f"✅ {skill_path.name} is v4.0 Enterprise compliant")
        sys.exit(0)
    else:
        print(f"❌ {skill_path.name} validation failed:")
        for error in errors:
            print(f"  - {error}")
        sys.exit(1)


if __name__ == '__main__':
    main()
```

### 7.3 사용 예시

```bash
# Single skill 검증
python scripts/validate_skill_v4.py .claude/skills/moai-domain-backend

# 모든 skills 검증
find .claude/skills -name "SKILL.md" -exec dirname {} \; | while read skill; do
  python scripts/validate_skill_v4.py "$skill"
done

# v4.0만 검증 (version 필터링)
find .claude/skills -name "SKILL.md" -exec grep -l "version: \"4.0.0\"" {} \; | \
  xargs dirname | while read skill; do
    python scripts/validate_skill_v4.py "$skill"
  done
```

---

## 🎯 결론 및 권장사항

### 핵심 발견

1. **MoAI-ADK 구조는 Claude Code 공식 패턴과 대부분 일치**: Agent/Skill 정의 방식, MCP 통합 모두 호환
2. **Progressive Disclosure 표준화 필요**: 3-level 구조를 모든 skills에 적용해야 함
3. **Context7 MCP 통합 패턴 확립**: 공식 문서 기반 권장사항 수립 완료
4. **v4.0 Enterprise 업그레이드 경로 명확**: 17 v2.0 skills + multiple v1.0 skills 대상

### 즉시 적용 가능한 패턴

1. **Agent 정의**: 현재 MoAI-ADK 방식 유지 (✅ 공식 패턴 일치)
2. **Skill 정의**: Progressive Disclosure 추가 (⚠️ 업그레이드 필요)
3. **MCP 통합**: 현재 `.claude/mcp.json` 방식 유지 (✅ 공식 패턴 일치)
4. **Context7 사용**: 이 문서의 Section 3.4 패턴 적용

### 다음 단계

1. **cc-manager Agent**: 이 문서를 참조하여 MCP 설정 및 Agent 표준화
2. **skill-factory Agent**: Section 4.2 템플릿 사용하여 v4.0 Enterprise 업그레이드
3. **검증 자동화**: Section 7.2 스크립트를 CI/CD에 통합

---

**문서 버전**: 1.0.0
**연구 완료일**: 2025-11-12
**다음 업데이트**: 2025-12-12 (또는 Claude Code 공식 변경 시)
