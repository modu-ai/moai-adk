---
title: Harness v4 Builder 심화 가이드
weight: 45
draft: false
---

Harness v4 Builder의 4-phase 워크플로우, Manifest 스키마, Runner 프리미티브를 상세히 안내합니다.

{{< callout type="info" >}}
**한 줄 요약**: Harness v4 Builder는 Socratic 인터뷰로 필요한 전문성을 파악하고, manifest 기반 Runner로 동적 팀을 운영합니다.
{{< /callout >}}

## 4-Phase Workflow 상세

### Phase 1: ANALYZE (분석)

현재 프로젝트의 기술 스택과 요구사항을 분석합니다.

#### 분석 대상

- **프로젝트 구조**: 디렉토리 계층, 핵심 패키지 식별
- **사용 언어**: Go, Python, TypeScript, Java 등 감지
- **프레임워크**: REST API, gRPC, FastAPI, Django 등 인식
- **기존 에이전트**: `.claude/agents/` 기존 정의 카탈로그
- **프로젝트 규모**: 파일 수, 코드 라인 수 기반 추정
- **의존성**: `go.mod`, `package.json`, `pyproject.toml` 분석

#### 산출물

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

### Phase 2: PLAN (계획)

ANALYZE 결과를 바탕으로 팀 구성을 설계합니다.

#### 계획 결정사항

| 항목 | 결정 방식 | 예시 |
|------|---------|------|
| **팀 규모** | 프로젝트 복잡도 × 필요 전문성 | 3~5명 |
| **역할 프로필** | Anthropic role_profiles (researcher/architect/implementer/tester/designer/reviewer) | architect, implementer, tester |
| **Worktree 격리** | 병렬 팀원 충돌 가능성 | L1_optional (선택적 격리) |
| **모델 선택** | 역할별 추론 복잡도 | architect: inherit, tester: haiku |
| **스킬 사전 로드** | 역할 전문성 필요 스킬 | moai-foundation-core, moai-domain-backend |

#### 계획 검증

생성 전에 사용자에게 확인:

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

### Phase 3: GENERATE (생성)

PLAN 승인 후 실제 에이전트 파일과 manifest를 생성합니다.

#### 생성 결과물

**1. 에이전트 정의 파일**

```
.claude/agents/harness/
├── architect.md
├── implementer.md
└── tester.md
```

각 파일은 YAML 프롬프트로 정의:

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

**2. Manifest 파일**

```
.moai/harness/manifest.json
```

Phase와 Teammate 정의가 포함된 JSON (스키마는 § Manifest 스키마 참조).

#### 생성 검증

```bash
ls .claude/agents/harness/
# architect.md, implementer.md, tester.md 확인

ls .moai/harness/
# manifest.json 확인

grep -c "\"name\": \"architect\"" .moai/harness/manifest.json
# phase 정의가 정확한지 확인
```

### Phase 4: ACTIVATE (활성화)

생성된 하네스를 등록하고 즉시 사용 가능하게 합니다.

#### 활성화 단계

1. **에이전트 검증**: 각 에이전트 파일 문법 확인
2. **Manifest 검증**: JSON 스키마 및 필드 검증
3. **커맨드 등록**: `/harness:backend-team` 커맨드 활성화
4. **Runner 초기화**: Manifest 기반 Runner 시작 준비
5. **Worktree 생성** (선택적): L1 격리 활성화 조건 설정

#### 활성화 확인

```bash
/harness list
# backend-team 표시

/harness:backend-team status
# 팀원 3명, 모델, 상태 확인
```

## Manifest 스키마

### 최상위 필드

| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `spec_id` | string | 예 | `HARNESS-{DOMAIN}-{NUM}` 형식 |
| `name` | string | 예 | 팀 표시 이름 |
| `version` | string | 예 | Semantic versioning `X.Y.Z` |
| `created_at` | string | 예 | ISO 8601 타임스탬프 |
| `worktree_isolation` | enum | 예 | `L1_optional` \| `none` |
| `phases` | array | 예 | Phase 객체 배열 |

### Phase 객체

```json
{
  "name": "run",
  "description": "구현 단계",
  "teammates": [...]
}
```

| 필드 | 타입 | 설명 |
|------|------|------|
| `name` | string | `plan` \| `run` \| `sync` |
| `description` | string | Phase 목표 설명 |
| `teammates` | array | Teammate 객체 배열 |

### Teammate 객체

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

| 필드 | 기본값 | 설명 |
|------|--------|------|
| `name` | 필수 | 팀원 ID (하이픈 사용, 공백 없음) |
| `role` | 필수 | 역할 설명 (자유 텍스트) |
| `model` | `inherit` | `inherit`, `haiku`, `sonnet`, `opus` |
| `mode` | `acceptEdits` | 권한 모드 (`acceptEdits`, `default`, `bypassPermissions`) |
| `skills` | `[]` | 사전 로드 스킬 배열 (예: `["moai-foundation-core"]`) |
| `isolation` | 없음 | `worktree_optional` (worktree 격리 조건부 활성화) |

### 전체 예시

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

## Runner 프리미티브

Manifest 기반 Runner는 생성된 팀을 실행합니다.

### Runner 생명 사이클

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

### Runner 설정

Runner의 동작은 manifest의 필드로 제어됩니다:

| 설정 | 의미 |
|------|------|
| `worktree_isolation: "L1_optional"` | 충돌 감지 시 자동 격리 적용 |
| `worktree_isolation: "none"` | 격리 비활성화 |
| `model: "inherit"` | 부모 세션 모델 상속 |
| `model: "haiku"` | Haiku 모델 강제 (비용 최적) |
| `skills: ["..."]` | 사전 로드 스킬 |

## Worktree 격리 규칙

### L1_optional 동작

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

### 격리 조건

다음 중 하나라도 참이면 격리 활성화:

1. **동일 파일 병렬 편집**: 두 팀원이 같은 파일을 동시에 수정
2. **재귀적 디렉토리 쓰기**: 팀원들이 같은 디렉토리에 여러 파일 생성
3. **의존성 경합**: 팀원 A의 출력이 팀원 B의 입력 (순서 중요)

### 비격리 (none) 선택 시

```
모든 팀원이 메인 프로젝트에서 작업
장점: 최소 메모리, 빠른 병렬
단점: 충돌 가능성
```

## 관련 문서

- [Harness v4 Builder 사용 가이드](/workflow-commands/moai-harness) - 커맨드 레퍼런스
- [에이전트 가이드](/advanced/agent-guide) - 에이전트 정의 형식
- [SPEC 기반 개발](/workflow-commands/moai-plan) - Harness와 SPEC 통합

{{< callout type="info" >}}
**팁**: Manifest는 생성 후 `/harness:team-name edit`으로 언제든 수정할 수 있습니다. 팀원 추가, 스킬 변경, 격리 정책 조정이 모두 가능합니다.
{{< /callout >}}
