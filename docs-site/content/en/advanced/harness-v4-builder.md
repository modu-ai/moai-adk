---
title: Harness v4 Builder Advanced Guide
weight: 45
draft: false
---

If the [Builder Agents Guide](/en/advanced/builder-agents) was the overview of the Harness v4 Builder, this document is the blueprint — it covers the deliverables of each stage of the 4-phase workflow, the full Manifest schema, and the operating rules of the Runner primitive.

{{< callout type="info" >}}
**One-line summary**: The Harness v4 Builder identifies the expertise you need through a Socratic interview and operates a dynamic team through a manifest-based Runner. Which teammate works with which model is decided by manifest declaration, not by code.
{{< /callout >}}

## 4-Phase Workflow in Detail

### Phase 1: ANALYZE

Analyzes the current project's tech stack and requirements. The goal of this phase is to answer "what expertise is this project missing" with data.

#### What Is Analyzed

- **Project structure**: directory hierarchy, identification of core packages
- **Languages**: detection of Go, Python, TypeScript, Java, etc.
- **Frameworks**: recognition of REST API, gRPC, FastAPI, Django, etc.
- **Existing agents**: catalog of existing definitions in `.claude/agents/`
- **Project scale**: estimation based on file count and lines of code
- **Dependencies**: analysis of `go.mod`, `package.json`, `pyproject.toml`

#### Deliverable

```yaml
analysis_result:
  languages:
    - go (primary)
    - shell (build scripts)
  frameworks:
    - REST API (net/http)
    - PostgreSQL ORM (sqlc)
  scale: "100~300 files, ~50K LOC"
  existing_agents: 0
  expertise_gaps:
    - Database schema design
    - API error handling patterns
    - Test coverage automation
```

### Phase 2: PLAN

Designs the team composition based on the ANALYZE results. Every decision that affects cost — from team size to per-role model assignment — is made in this phase.

#### Planning Decisions

| Item | How decided | Example |
|------|---------|------|
| **Team size** | Project complexity × required expertise | 3-5 members |
| **Role profiles** | Anthropic role_profiles (researcher/architect/implementer/tester/designer/reviewer) | architect, implementer, tester |
| **Worktree isolation** | Likelihood of parallel-teammate conflicts | L1_optional (optional isolation) |
| **Model selection** | Reasoning complexity per role | architect: inherit, tester: haiku |
| **Skill preload** | Skills needed for role expertise | moai-foundation-core, moai-domain-backend |

Per-role model selection is the heart of tokenomics — design goes to a model capable of deep reasoning, while repetitive test writing is assigned to a cheaper model.

#### Plan Confirmation

The plan is confirmed with the user before generation. No files are ever created without an approval gate.

```
계획된 팀 구성:
- 팀명: Backend Development Team
- 팀원 3명:
  ① architect (model: inherit)
  ② implementer (model: inherit)
  ③ tester (model: haiku)
- Worktree 격리: L1_optional
- Manifest: .moai/harness/manifest.json

이 구성으로 진행할까요?
```

### Phase 3: GENERATE

After PLAN approval, the actual agent files and manifest are generated.

#### Generated Artifacts

**1. Agent definition files**

```
.claude/agents/harness/
├── architect.md
├── implementer.md
└── tester.md
```

Each file is defined with a YAML prompt.

```yaml
---
name: architect
description: API 아키텍처 설계 전문가
tools: Read, Write, Edit, Grep, Glob, Bash
model: inherit
---

당신은 이 프로젝트의 API 아키텍처 전문가입니다.
[역할별 상세 지침]
```

**2. Manifest file**

```
.moai/harness/manifest.json
```

A JSON containing the Phase and Teammate definitions (see § Manifest Schema for the schema).

#### Generation Verification

Right after generation, you can directly verify file existence and definition correctness.

```bash
ls .claude/agents/harness/
# architect.md, implementer.md, tester.md 확인

ls .moai/harness/
# manifest.json 확인

grep -c "\"name\": \"architect\"" .moai/harness/manifest.json
# phase 정의가 정확한지 확인
```

### Phase 4: ACTIVATE

Registers the generated harness and makes it immediately usable.

#### Activation Steps

1. **Agent validation**: syntax check on each agent file
2. **Manifest validation**: JSON schema and field validation
3. **Command registration**: the `/harness:backend-team` command is activated
4. **Runner initialization**: the manifest-based Runner is prepared to start
5. **Worktree creation** (optional): L1 isolation activation conditions configured

#### Activation Check

```bash
/harness list
# backend-team 표시

/harness:backend-team status
# 팀원 3명, 모델, 상태 확인
```

## Manifest Schema

### Top-Level Fields

| Field | Type | Required | Description |
|------|------|------|------|
| `spec_id` | string | Yes | `HARNESS-{DOMAIN}-{NUM}` format |
| `name` | string | Yes | Team display name |
| `version` | string | Yes | Semantic versioning `X.Y.Z` |
| `created_at` | string | Yes | ISO 8601 timestamp |
| `worktree_isolation` | enum | Yes | `L1_optional` \| `none` |
| `phases` | array | Yes | Array of Phase objects |

### Phase Object

```json
{
  "name": "run",
  "description": "구현 단계",
  "teammates": [...]
}
```

| Field | Type | Description |
|------|------|------|
| `name` | string | `plan` \| `run` \| `sync` |
| `description` | string | Description of the phase goal |
| `teammates` | array | Array of Teammate objects |

### Teammate Object

```json
{
  "name": "api-developer",
  "role": "REST API 엔드포인트 개발",
  "model": "inherit",
  "mode": "acceptEdits",
  "skills": ["moai-foundation-core"],
  "isolation": "worktree_optional"
}
```

| Field | Default | Description |
|------|--------|------|
| `name` | required | Teammate ID (hyphens, no spaces) |
| `role` | required | Role description (free text) |
| `model` | `inherit` | `inherit`, `haiku`, `sonnet`, `opus` |
| `mode` | `acceptEdits` | Permission mode (`acceptEdits`, `default`, `bypassPermissions`) |
| `skills` | `[]` | Preloaded skill array (e.g. `["moai-foundation-core"]`) |
| `isolation` | none | `worktree_optional` (conditional worktree isolation) |

### Full Example

```json
{
  "spec_id": "HARNESS-BACKEND-001",
  "name": "Backend Development Team",
  "version": "1.0.0",
  "created_at": "2026-07-01T10:00:00Z",
  "worktree_isolation": "L1_optional",
  
  "phases": [
    {
      "name": "plan",
      "description": "아키텍처 설계 및 SPEC 작성",
      "teammates": [
        {
          "name": "architect",
          "role": "API 아키텍처 전문가",
          "model": "inherit",
          "mode": "acceptEdits",
          "skills": ["moai-foundation-core"]
        }
      ]
    },
    {
      "name": "run",
      "description": "실제 구현",
      "teammates": [
        {
          "name": "db-engineer",
          "role": "DB 설계 및 마이그레이션",
          "model": "inherit",
          "mode": "acceptEdits",
          "isolation": "worktree_optional"
        },
        {
          "name": "api-developer",
          "role": "REST API 엔드포인트 구현",
          "model": "inherit",
          "mode": "acceptEdits",
          "isolation": "worktree_optional"
        },
        {
          "name": "test-engineer",
          "role": "단위 테스트 및 통합 테스트",
          "model": "haiku",
          "mode": "acceptEdits"
        }
      ]
    }
  ]
}
```

## Runner Primitive

The manifest-based Runner executes the generated team.

### Runner Lifecycle

```
Team Spawn
  ↓
[Phase 1: plan]
  → Teammate(architect) 생성 및 위임
  → 결과 수집
  ↓
[Phase 2: run]
  → Teammate(db-engineer) 병렬 생성
  → Teammate(api-developer) 병렬 생성
  → Teammate(test-engineer) 순차 생성
  → 결과 수집 및 통합
  ↓
[Phase 3: sync]
  → 기본 manager-docs 실행
  ↓
Team Teardown
```

### Runner Configuration

The Runner's behavior is controlled by manifest fields.

| Setting | Meaning |
|------|------|
| `worktree_isolation: "L1_optional"` | Automatic isolation applied on conflict detection |
| `worktree_isolation: "none"` | Isolation disabled |
| `model: "inherit"` | Inherit the parent session's model |
| `model: "haiku"` | Force the Haiku model (cost-optimal) |
| `skills: ["..."]` | Preloaded skills |

## Worktree Isolation Rules

### L1_optional Behavior

```
Runner 생성 시:
├── 팀원 1: 메인 프로젝트 루트
├── 팀원 2: 메인 프로젝트 루트
└── 충돌 감지 시
    ├── 팀원 2 → L1 워크트리로 전환
    └── 팀원 1은 메인 유지 (또는 팀원 1도 전환)

결과:
└── 파일 충돌 회피 ✓
```

### Isolation Conditions

Isolation activates when any of the following is true.

1. **Parallel edits to the same file**: two teammates modify the same file simultaneously
2. **Recursive directory writes**: teammates create multiple files in the same directory
3. **Dependency contention**: teammate A's output is teammate B's input (ordering matters)

### When Choosing No Isolation (none)

```
모든 팀원이 메인 프로젝트에서 작업
장점: 최소 메모리, 빠른 병렬
단점: 충돌 가능성
```

## Related Documents

- [Harness v4 Builder Usage Guide](/en/workflow-commands/moai-harness) - command reference
- [Agent Guide](/en/advanced/agent-guide) - agent definition format
- [SPEC-Based Development](/en/workflow-commands/moai-plan) - Harness and SPEC integration

{{< callout type="info" >}}
**Tip**: After creation, the manifest can be edited anytime with `/harness:team-name edit`. Adding teammates, changing skills, and adjusting the isolation policy are all possible.
{{< /callout >}}
