---
title: Harness v4 Builder 심화 가이드
weight: 45
draft: false
---

[빌더 에이전트 가이드](/ko/advanced/builder-agents)가 Harness v4 Builder의 개요였다면, 이 문서는 설계도입니다 — 4-phase 워크플로우의 각 단계 산출물, Manifest 스키마 전체, Runner 프리미티브의 동작 규칙을 다룹니다.

{{< callout type="info" >}}
**한 줄 요약**: Harness v4 Builder는 Socratic 인터뷰로 필요한 전문성을 파악하고, manifest 기반 Runner로 동적 팀을 운영합니다. 어떤 팀원이 어떤 모델로 일할지는 코드가 아니라 manifest 선언으로 결정됩니다.
{{< /callout >}}

## 4-Phase Workflow 상세

### Phase 1: ANALYZE (분석)

현재 프로젝트의 기술 스택과 요구사항을 분석합니다. 이 단계의 목표는 "이 프로젝트에 어떤 전문성이 부족한가"를 데이터로 답하는 것입니다.

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

ANALYZE 결과를 바탕으로 팀 구성을 설계합니다. 팀 규모부터 역할별 모델 배정까지, 비용에 영향을 주는 결정은 전부 이 단계에서 내려집니다.

#### 계획 결정사항

| 항목 | 결정 방식 | 예시 |
|------|---------|------|
| **Specialist 수** | 프로젝트 복잡도 × 필요 전문성 (HARD 상한 3~7) | 3개 specialist |
| **실행 원시(primitive)** | specialist별 실행 형태 | sub-agent, adversarial-fan-out |
| **격리(isolation)** | 병렬 specialist 충돌 가능성 | none \| worktree |
| **모델·effort 배정** | specialist별 추론 복잡도 (목적 기반) | content-author: opus/high, translator: sonnet/medium |
| **companion 스킬** | specialist 전문성 필요 스킬 | hns-oss-docs-i18n-rules |

specialist별 모델·effort 선택이 토크노믹스의 핵심입니다 — 깊은 추론이 필요한 저작은 상위 모델·high effort에, 반복적인 파생 작업은 저렴한 모델·medium effort에 배정합니다. 사용자 승인 게이트는 PLAN→GENERATE 경계에서 `AskUserQuestion`으로 이루어집니다.

#### 계획 검증

생성 전에 사용자에게 확인합니다. 승인 게이트 없이 파일이 생성되는 일은 없습니다.

```
계획된 하네스 구성:
- 이름: backend-team
- specialist 3개:
  ① architect (primitive: sub-agent, model: opus, effort: high)
  ② implementer (primitive: sub-agent, model: inherit, effort: high)
  ③ tester (primitive: sub-agent, model: sonnet, effort: medium)
- entry 명령어: /harness:backend-team

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

각 파일은 YAML 프롬프트로 정의됩니다.

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

Phase와 Teammate 정의가 포함된 JSON입니다 (스키마는 § Manifest 스키마 참조).

#### 생성 검증

생성 직후 파일 존재와 정의 정확성을 직접 확인할 수 있습니다.

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
3. **커맨드 등록**: `/harness:backend-team` entry 커맨드 활성화
4. **Runner 초기화**: Manifest 기반 Runner 시작 준비
5. **Worktree 생성** (선택적): specialist 격리 활성화 조건 설정

#### 활성화 확인

```bash
moai harness list
# backend-team 표시 (이름 + 도메인 + entry 명령어)

moai harness doctor
# 참조 무결성 스모크 게이트 (specialist·skill·workflow 참조 검증)
```

## Manifest 스키마

### 최상위 필드

| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `name` | string | 예 | 하네스 이름 (entry 명령어에 사용) |
| `domain` | string | 예 | 하네스 도메인 설명 |
| `patterns` | array | 예 | 실행 패턴 (`Pipeline`, `Fan-out/Fan-in`, `Producer-Reviewer` 등) |
| `specialists` | array | 예 | Specialist 객체 배열 (3~7개 HARD 상한) |
| `sprint_contract` | object | 예 | 품질 차원·임계값·must_pass 게이트 |
| `companion_skills` | array | — | 하네스 전용 companion 스킬 목록 |
| `entry_command` | string | 예 | `/harness:<name>` entry 명령어 |
| `runner_workflow` | string | 예 | Runner 워크플로 스크립트 파일 |
| `schedule` | object | — | (선택) 반복 실행 스케줄 — `mode: discovery-only` 등 |

### Specialist 객체

```json
{
  "role": "content-author",
  "description": "canonical-locale 원문 저작",
  "agent_file": ".claude/agents/harness/hns-oss-docs-content-author-specialist.md",
  "primitive": "sub-agent",
  "isolation": "none",
  "effort": "high",
  "model": "opus"
}
```

| 필드 | 설명 |
|------|------|
| `role` | specialist 역할 (하이픈/영문) |
| `description` | 역할 설명 (자유 텍스트) |
| `agent_file` | specialist 에이전트 파일 경로 (`.claude/agents/harness/`) |
| `primitive` | 실행 원시 (`sub-agent`, `adversarial-fan-out` 등) |
| `isolation` | 격리 수준 (`none`, `worktree`) |
| `effort` | 추론 강도 (`low`, `medium`, `high`, `xhigh`) — 목적 기반 |
| `model` | 모델 티어 (`opus`, `sonnet`, `haiku`, `inherit`) — 목적 기반 |

### Sprint Contract

```json
{
  "dimensions": ["locale-parity", "build-clean", "style-compliance", "content-fidelity"],
  "thresholds": { "locale-parity": 1.0, "build-clean": 1.0, "style-compliance": 0.95 },
  "must_pass": ["locale-parity", "build-clean"]
}
```

`dimensions`는 채점 차원, `thresholds`는 차원별 통과 임계값, `must_pass`는 반드시 통과해야 하는 게이트를 정의합니다.

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

Runner의 동작은 manifest의 필드로 제어됩니다.

| 설정 | 의미 |
|------|------|
| `isolation: "worktree"` | specialist에 worktree 격리 적용 |
| `isolation: "none"` | 격리 비활성화 |
| `model: "inherit"` | 부모 세션 모델 상속 |
| `model: "sonnet"` | 파생/반복 작업용 저비용 티어 |
| `effort: "high"` \| `"medium"` | specialist별 추론 강도 (목적 기반) |
| `companion_skills: ["..."]` | 하네스 전용 companion 스킬 |

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

다음 중 하나라도 참이면 격리가 활성화됩니다.

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

- [Harness v4 Builder 사용 가이드](/ko/workflow-commands/moai-harness) - 커맨드 레퍼런스
- [에이전트 가이드](/ko/advanced/agent-guide) - 에이전트 정의 형식
- [SPEC 기반 개발](/ko/workflow-commands/moai-plan) - Harness와 SPEC 통합

{{< callout type="info" >}}
**팁**: Manifest는 생성 후 `moai harness edit <name>`으로 편집 경로를 확인해 언제든 수정할 수 있습니다. specialist 추가, 스킬 변경, 격리 정책 조정이 모두 가능합니다.
{{< /callout >}}
