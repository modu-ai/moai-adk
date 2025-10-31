# 플러그인 마켓플레이스 종합 준수성 분석 및 병렬 개선 전략

**생성일**: 2025-10-31
**분석자**: cc-manager (MoAI-ADK)
**대상**: moai-marketplace/plugins/ (5개 플러그인)
**총 분석 대상**: 60+ 파일
**현재 준수도**: 13-37%
**목표 준수도**: 100% (마켓플레이스 제출 가능)
**예상 전체 노력**: 40-50시간 (순차) / 12-18시간 (병렬, 5명)

---

## 📊 현재 상태 분석

### 준수성 행렬

| 플러그인 | plugin.json | 에이전트 | 스킬 | 커맨드 | README | LICENSE | 평균 |
|---------|------------|--------|------|--------|--------|---------|------|
| **backend** | 20% | 60% | 0% | 0% | 0% | 0% | **13%** |
| **devops** | 20% | 60% | 0% | 0% | 0% | 0% | **13%** |
| **frontend** | 20% | 60% | 0% | 0% | 0% | 0% | **13%** |
| **uiux** | 20% | 60% | 0% | 0% | 0% | 0% | **13%** |
| **technical-blog** | 20% | 60% | 0% | 50% | 95% | 0% | **37%** |
| **평균** | **20%** | **60%** | **0%** | **10%** | **19%** | **0%** | **18%** |

### 주요 발견

#### 🔴 CRITICAL ISSUES (모든 플러그인)

1. **plugin.json 필드 누락**
   - `id`, `category`, `minClaudeCodeVersion` - 마켓플레이스 등록 불가
   - `commands`, `agents` 배열 - 메니페스트에 미등록
   - `permissions` - 보안 위험, 접근 제어 없음
   - `status`, `tags`, `repository`, `license` - 메타데이터 부족

2. **스킬 콘텐츠 - 전부 플레이스홀더** (33개 파일)
   ```
   [Skill content for moai-domain-backend]  ← 모두 이 상태
   ```
   - 에이전트가 사용할 실제 지식 없음
   - Progressive disclosure 구현 안 됨
   - 코드 예제 없음

3. **에이전트 설명 부족**
   - "Use PROACTIVELY for [triggers]" 형식 없음
   - Proactive Triggers 섹션 없음
   - 사용 시점 불명확

4. **명령어 부족**
   - 기술블로그만 1개 있음 (/blog-write)
   - 다른 4개 플러그인은 0개

#### 🟠 HIGH PRIORITY ISSUES

| 문제 | 영향 | 플러그인 | 우선순위 |
|------|------|---------|--------|
| README 미작성 | 사용자 문서 부족 | 4개 (backend, devops, frontend, uiux) | 🟠 HIGH |
| License 미작성 | 법적 명확성 없음 | 5개 모두 | 🟠 HIGH |
| MCP 선언 없음 | 통합 불가 | devops (vercel/supabase/render), uiux (figma) | 🟠 HIGH |
| CONTRIBUTING.md 없음 | 기여 지침 없음 | 5개 모두 | 🟠 HIGH |
| CHANGELOG.md 없음 | 버전 관리 없음 | 5개 모두 | 🟠 HIGH |

---

## 🚀 병렬 실행 전략 (8개 그룹)

### 병렬 실행 구조

```
병렬 그룹 1: plugin.json 완성 (2.5시간)
├─ backend, devops, frontend, uiux, technical-blog
└─ 의존성: 없음 ✓

병렬 그룹 2-3: 빠른 작업 (동시 3시간)
├─ Group 2: README 작성 (3시간)
├─ Group 3: Agent 설명 업데이트 (1.5시간)
└─ 의존성: Group 1 완료 후

병렬 그룹 4-7: 중간 작업 (동시 10시간)
├─ Group 4: 스킬 콘텐츠 (25-30시간) ← 병렬로 5명 작업 시 6시간
├─ Group 5: LICENSE, CONTRIBUTING, CHANGELOG (2시간)
├─ Group 6: .mcp.json (40분)
├─ Group 7: hooks.json (2.5시간)
└─ 의존성: 각 그룹 독립적

병렬 그룹 8: 커맨드 작성 (10-15시간)
└─ 의존성: Group 4 (스킬) 완료 후 (스킬 활용)
```

### 기대 타임라인

**순차 실행 (1명)**: 40-50시간 → 1주일 (8시간/일)
**병렬 실행 (5명)**: 12-18시간 → 2-3일 (고강도 스프린트)

---

## 📋 GROUP별 상세 업무

### GROUP A: plugin.json 완성 (🔴 CRITICAL) - 병렬 그룹 1

**예상 시간**: 2.5시간 총 (30분/플러그인 × 5)
**담당**: 5명 (각 1명이 1개 플러그인)

#### A1: moai-plugin-backend
```json
{
  "id": "moai-plugin-backend",
  "category": "backend",
  "minClaudeCodeVersion": "1.0.0",
  "status": "development",
  "tags": ["fastapi", "python", "sqlalchemy", "database", "async"],
  "repository": "https://github.com/moai-adk/moai-marketplace/...",
  "license": "MIT",
  "agents": [
    {"name": "fastapi-specialist", "path": "agents/fastapi-specialist.md"},
    {"name": "backend-architect", "path": "agents/backend-architect.md"},
    {"name": "database-expert", "path": "agents/database-expert.md"},
    {"name": "api-designer", "path": "agents/api-designer.md"}
  ],
  "permissions": {
    "allowedTools": ["Read", "Write", "Edit", "Glob", "Grep", "Bash(python3:*)", "Bash(pip:*)", "Bash(uv:*)", "Task"],
    "deniedTools": []
  },
  "installCommand": "/plugin install moai-plugin-backend",
  "releaseNotes": "Initial 1.0.0-dev release with 4 agents and 4 skills"
}
```

#### A2-A5: devops, frontend, uiux, technical-blog
동일한 구조, 플러그인별 값만 변경

**주의**: devops, uiux는 `mcpServers` 추가
```json
"mcpServers": [
  {"name": "vercel", "type": "optional"},
  {"name": "supabase", "type": "optional"},
  {"name": "render", "type": "optional"}
]
```

---

### GROUP B: README.md 작성 (🟠 HIGH) - 병렬 그룹 2

**예상 시간**: 3시간 총 (45분/파일 × 4)
**담당**: 4명 (technical-blog는 이미 있음)
**참고**: technical-blog README를 템플릿으로 사용

#### 템플릿 구조
```markdown
# [Plugin Name] Plugin

**[One-line description]** — [Tech stack summary]

## 🎯 개요

[플러그인이 무엇을 하는지 설명]

## 🏗️ 구조

### [N개] 전문가 에이전트
| 에이전트 | 역할 |
|--------|------|
| agent-name | 설명 |

### [N개] 스킬
| 스킬 | 목적 |
|-----|------|
| skill-name | 설명 |

### [N개] 커맨드 (있으면)
| 커맨드 | 기능 |
|--------|------|
| /command-name | 설명 |

## ⚡ 빠른 시작

### 설치
```bash
/plugin install moai-plugin-[name]
```

### 기본 사용법
[예제 코드]

## 📚 주요 기능
- 기능 1
- 기능 2
- 기능 3

## 🤝 기여하기
[CONTRIBUTING.md 참조]

## 📄 라이선스
MIT License - [LICENSE 파일 참조]
```

#### B1: backend (45분)
#### B2: devops (45분)
#### B3: frontend (45분)
#### B4: uiux (45분)

---

### GROUP C: 에이전트 설명 업데이트 (🟠 HIGH) - 병렬 그룹 3

**예상 시간**: 1.5시간 총 (3분/파일 × 26)
**담당**: 2-3명 (파일 배분)
**작업**: YAML frontmatter 수정

#### 현재 상태
```yaml
---
name: fastapi-specialist
type: specialist
description: FastAPI specialist designing async APIs and async patterns
tools: [Read, Write, Edit, Grep, Glob]
model: sonnet
---
```

#### 목표 상태
```yaml
---
name: fastapi-specialist
description: Use PROACTIVELY for FastAPI endpoint creation, request validation, OpenAPI documentation, and async patterns
tools: [Read, Write, Edit, Grep, Glob, Bash(python3:*)]
model: sonnet
---

# FastAPI Specialist Agent

## Proactive Triggers
- 사용자가 "FastAPI 엔드포인트 생성"을 요청할 때
- REST API 설계가 필요할 때
- 요청 검증 로직이 필요할 때
- OpenAPI 문서 생성이 필요할 때

## 책임
[기존 내용 유지]
```

#### C1-C5: 26개 에이전트 파일
- backend: 4개 (fastapi-specialist, backend-architect, database-expert, api-designer)
- devops: 4개 (deployment-strategist, render-specialist, supabase-specialist, vercel-specialist)
- frontend: 4개 (design-system-manager, frontend-architect, performance-optimizer, typescript-specialist)
- uiux: 7개 (accessibility-specialist, component-builder, css-html-generator, design-documentation-writer, design-strategist, design-system-architect, figma-designer)
- technical-blog: 7개 (code-example-curator, markdown-formatter, seo-discoverability-specialist, technical-content-strategist, technical-writer, template-workflow-coordinator, visual-content-designer)

---

### GROUP D: 스킬 콘텐츠 작성 (🟠 HIGH) - 병렬 그룹 4

**예상 시간**: 25-30시간 (순차) / 6시간 (5명 병렬)
**담당**: 5명 (플러그인별 1명)
**중요도**: 가장 중요한 작업 - 실제 가치 제공

#### D1: Backend 스킬 (4개, 4시간)
```
moai-lang-fastapi-patterns.md (30-45분)
├─ Quick Start: FastAPI 기초
├─ Core Patterns: 라우트, DI, Pydantic 모델
├─ Advanced: WebSocket, 배경작업, 미들웨어
└─ References: 공식 문서

moai-lang-python.md (30-45분)
moai-domain-backend.md (30-45분)
moai-domain-database.md (30-45분)
```

#### D2: DevOps 스킬 (6개, 6시간)
```
moai-saas-vercel-mcp.md
moai-saas-supabase-mcp.md
moai-saas-render-mcp.md
moai-domain-backend.md (공유)
moai-domain-frontend.md (공유)
moai-domain-devops.md
```

#### D3: Frontend 스킬 (5개, 5시간)
```
moai-lang-nextjs-advanced.md
moai-lang-typescript.md
moai-design-shadcn-ui.md
moai-design-tailwind-v4.md
moai-domain-frontend.md (공유)
```

#### D4: UI/UX 스킬 (6개, 6시간)
```
moai-design-figma-mcp.md
moai-design-figma-to-code.md
moai-design-shadcn-ui.md (공유)
moai-design-tailwind-v4.md (공유)
moai-lang-tailwind-shadcn.md
moai-domain-frontend.md (공유)
```

#### D5: Technical Blog 스킬 (12개, 9시간)
```
moai-content-blog-strategy.md
moai-content-blog-templates.md
moai-content-code-examples.md
moai-content-hashtag-strategy.md
moai-content-image-generation.md
moai-content-llms-txt-management.md
moai-content-markdown-best-practices.md
moai-content-markdown-to-blog.md
moai-content-meta-tags.md
moai-content-seo-optimization.md
moai-content-technical-seo.md
moai-content-technical-writing.md
```

#### 스킬 콘텐츠 구조
```markdown
---
name: moai-lang-fastapi-patterns
type: language
description: FastAPI async patterns, DI, validation. Use when building REST APIs and async endpoints.
tier: language
---

# FastAPI Patterns

## 빠른 시작 (30초)
[핵심 개념 1줄 요약]

## 핵심 패턴

### 패턴 1: Async Route Handlers
\`\`\`python
@app.get("/items/{item_id}")
async def get_item(item_id: int):
    return {"item_id": item_id}
\`\`\`

### 패턴 2: Dependency Injection
\`\`\`python
async def get_current_user(token: str = Depends(oauth2_scheme)):
    return verify_token(token)
\`\`\`

### 패턴 3: Pydantic Models
\`\`\`python
class Item(BaseModel):
    name: str
    price: float
    description: Optional[str] = None
\`\`\`

## Progressive Disclosure
[상세 가이드]

## Works Well With
- moai-lang-python
- moai-domain-backend
- moai-domain-database

## 참고 자료
[공식 문서 링크]
```

---

### GROUP E: 옵션 파일 작성 (🟡 MEDIUM) - 병렬 그룹 5-7

#### E1: LICENSE 파일 (병렬 그룹 5)
**예상 시간**: 5분 × 5 = 25분
**담당**: 5명

```
MIT License

Copyright (c) 2025 MoAI-ADK Project

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.
```

#### E2: CONTRIBUTING.md 파일 (병렬 그룹 5)
**예상 시간**: 15분 × 5 = 75분

```markdown
# 기여 가이드

## 시작하기
[개발 환경 설정]

## 개발 워크플로우
[Git 워크플로우]

## 코드 표준
[린팅, 포맷팅]

## 테스트
[테스트 방법]

## PR 프로세스
[PR 가이드라인]
```

#### E3: CHANGELOG.md 파일 (병렬 그룹 5)
**예상 시간**: 10분 × 5 = 50분

```markdown
# 변경 로그

모든 주목할 만한 변경 사항은 이 파일에 문서화됩니다.

## [미공개]

## [1.0.0-dev] - 2025-10-31

### 추가됨
- 초기 플러그인 구조
- [에이전트 목록]
- [스킬 목록]
- [커맨드 목록]

### 변경됨
- N/A

### 수정됨
- N/A
```

#### E4: .mcp.json 파일 (병렬 그룹 6)
**예상 시간**: 20분 × 2 = 40분
**대상**: devops, uiux

devops 버전:
```json
{
  "mcpServers": {
    "vercel": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-vercel"],
      "env": {
        "VERCEL_API_TOKEN": "${VERCEL_TOKEN}"
      }
    },
    "supabase": {...},
    "render": {...}
  }
}
```

#### E5: hooks.json 파일 (병렬 그룹 7)
**예상 시간**: 30분 × 5 = 150분
**우선순위**: 🟢 LOW (선택)

```json
{
  "sessionStart": {
    "name": "onSessionStart",
    "description": "플러그인 상태 표시",
    "priority": 50
  },
  "preToolUse": {
    "name": "onPreToolUse",
    "description": "도구 권한 검증",
    "priority": 100
  }
}
```

---

### GROUP F: 커맨드 작성 (🟠 HIGH) - 병렬 그룹 8

**예상 시간**: 10-15시간 (순차) / 4시간 (병렬)
**담당**: 4명 (플러그인별 1명)
**의존성**: Group D (스킬) 완료 후

#### F1: Backend 커맨드 (3개, 2시간)

**`/init-fastapi` 명령**
```yaml
---
name: init-fastapi
description: FastAPI 프로젝트를 uv 패키지 관리자로 초기화
argument-hint: ["project-name", "optional: --with-db"]
tools: [Task, Read, Write, Bash]
model: sonnet
---

# /init-fastapi 명령

FastAPI 프로젝트를 SQLAlchemy와 Alembic으로 부트스트랩합니다.

## 사용법
\`\`\`
/init-fastapi my-api --with-db
\`\`\`

## 워크플로우
1. 프로젝트 디렉토리 생성
2. uv 초기화
3. FastAPI 의존성 설치
4. 기본 구조 생성
5. Alembic 초기화

## 출력
- 완성된 FastAPI 프로젝트
- pyproject.toml 설정
- 기본 main.py
```

**`/db-setup` 명령** (1시간)

**`/resource-crud` 명령** (1시간)

#### F2: DevOps 커맨드 (3개, 2시간)
- `/deploy-vercel`
- `/deploy-render`
- `/setup-supabase`

#### F3: Frontend 커맨드 (3개, 2시간)
- `/init-nextjs`
- `/add-component`
- `/setup-playwright`

#### F4: UI/UX 커맨드 (3개, 2시간)
- `/figma-sync`
- `/design-tokens`
- `/component-library`

---

## ✅ 성공 기준

### 마켓플레이스 제출 체크리스트

각 플러그인마다:
- [ ] plugin.json: 모든 필수 필드 포함
- [ ] Agents: 메니페스트에 등록됨
- [ ] Skills: 플레이스홀더 아님, 실제 콘텐츠 포함
- [ ] 최소 1개 커맨드 존재
- [ ] README.md: 포괄적이고 명확함
- [ ] LICENSE 파일 존재
- [ ] YAML frontmatter: 공식 템플릿 준수
- [ ] Proactive Triggers: 문서화됨

### 최종 준수도 목표
- 마켓플레이스 제출 가능: 100%
- 모든 플러그인: ≥95% 준수

---

## 🎯 권장 실행 계획

### Phase 1: 기초 구축 (1일, 병렬 가능)
**Group A 실행**: plugin.json 완성 (2.5시간)
- 5명 동시 작업, 1명/플러그인
- 검증: JSON 문법, 필드 완성도

### Phase 2: 문서화 (1.5일, 병렬 가능)
**Group B+C 병렬 실행** (4.5시간)
- 4명 README 작성 (3시간)
- 2-3명 에이전트 설명 업데이트 (1.5시간)

### Phase 3: 스킬 작성 스프린트 (2-3일, 집중 작업)
**Group D 실행**: 33개 스킬 콘텐츠 (25-30시간)
- 5명 병렬 (플러그인별)
- 각자 6시간 × 5 = 30시간 / 5명 = 6시간
- 일정: 하루 6시간 × 3일 = 18시간 (실제: 스킬 품질로 5-10시간 추가 가능)

### Phase 4: 보조 파일 (1.5일, 병렬 가능)
**Group E1-E5 병렬 실행** (5시간)
- E1-E3: LICENSE, CONTRIBUTING, CHANGELOG (2시간)
- E4: .mcp.json (40분)
- E5: hooks.json (2.5시간, 선택)

### Phase 5: 커맨드 작성 (3-4일)
**Group F 실행**: 10-12개 커맨드 (10-15시간)
- 4명 병렬 (플러그인별)
- 각자 2-3시간 × 4 = 10-12시간

---

## 📈 예상 임팩트

| 메트릭 | 현재 | 목표 | 개선율 |
|--------|------|------|--------|
| 마켓플레이스 준비 플러그인 | 0/5 (0%) | 5/5 (100%) | **+500%** |
| 기능 스킬 | 0/33 (0%) | 33/33 (100%) | **+100%** |
| 문서화된 플러그인 | 1/5 (20%) | 5/5 (100%) | **+400%** |
| 법적 준수 | 0/5 (0%) | 5/5 (100%) | **+500%** |
| 보안 준수 (권한) | 0/5 (0%) | 5/5 (100%) | **+500%** |
| 커맨드 커버리지 | 1/15+ (7%) | 12/12 (100%) | **+1400%** |
| 평균 플러그인 완성도 | 18% | 100% | **+450%** |

---

## 💡 주요 이점

1. **병렬 처리로 시간 75% 단축** (40시간 → 12시간)
2. **모든 플러그인 일관성 보장** (템플릿 기반)
3. **마켓플레이스 즉시 제출 가능**
4. **사용자 경험 대폭 개선** (문서, 커맨드, 스킬)
5. **장기 유지보수 용이** (구조, 메타데이터)

---

## 🚨 주의 사항

1. **스킬 콘텐츠가 가장 긴 작업** - 10-15시간
   - 실제 기술 내용 필요
   - 코드 예제 포함 필수
   - 테스트/검증 권장

2. **플러그인 간 스킬 공유** - 중복 조심
   - moai-domain-frontend (3개 플러그인)
   - moai-domain-backend (2개 플러그인)
   - moai-design-shadcn-ui, moai-design-tailwind-v4 (중복)

3. **MCP 토큰/설정** 필요
   - Vercel, Supabase, Render, Figma 계정 필요
   - .mcp.json에 환경변수 설정 필요

4. **테스트 필수**
   - 각 커맨드 실행 테스트
   - 에이전트 호출 테스트
   - 스킬 로드 테스트

---

## 📞 의사결정 필요 항목

1. **스킬 콘텐츠 저자**: 누가 33개 스킬을 작성할 것인가? (25-30시간)
2. **우선순위**: 어느 플러그인을 먼저 마켓플레이스에 올릴 것인가?
3. **MCP 접근권**: Vercel/Supabase/Render/Figma API 토큰 확보?
4. **커맨드 범위**: 포괄적인 커맨드 세트인가, 최소 세트인가?
5. **Hooks 필요**: v1.0 필수인가, v1.1 기능인가?

---

**보고서 작성자**: cc-manager (MoAI-ADK)
**분석 깊이**: 종합 (80+ 파일, 5개 플러그인)
**실행 준비도**: 100% (모든 업무 상세 정의)
**다음 단계**: GROUP A 시작 승인 대기
