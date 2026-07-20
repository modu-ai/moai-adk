---
title: 카탈로그 시스템
weight: 80
draft: false
---

토크노믹스는 토큰에만 적용되는 원칙이 아닙니다. 프로젝트에 배포되는 템플릿 파일 하나하나도 결국 세션이 로드할 컨텍스트 후보입니다. 카탈로그 시스템은 "필요한 것만 배포한다"는 원칙으로 이 비용을 초기화 단계에서부터 줄입니다.

## 개요

MoAI-ADK의 카탈로그 시스템은 모든 에이전트, 스킬, 규칙을 **3계층 매니페스트**(`catalog.yaml`)로 관리합니다. 기본 배포는 **slim 모드**로 핵심 템플릿(core)만 배포해 초기화가 빠르고 프로젝트에 남는 파일도 가벼워집니다. 전체 배포가 필요하면 `--all` 플래그를 사용합니다.

## 3계층 매니페스트

모든 배포 대상은 세 계층 중 하나에 속합니다.

| 계층 | catalog.yaml 키 | 설명 | 배포 기준 |
|------|-----------------|------|----------|
| **Core** | `catalog.core` | 핵심 인프라 — 오케스트레이터, 품질 게이트, 기본 스킬/에이전트 | 항상 배포 (slim 모드 기본) |
| **Optional Packs** | `catalog.optional_packs` | 도메인 확장 — backend, frontend, design, devops, deployment, testing 팩 | `--all` 플래그 시 배포 |
| **Harness-generated** | `catalog.harness_generated` | 하네스가 동적 생성한 에이전트/스킬 | `--all` 플래그 시 배포 |

## 카탈로그 파일

카탈로그 매니페스트는 `internal/template/catalog.yaml`에 YAML 형식으로 정의됩니다.

```yaml
catalog:
  core:                        # 항상 배포 (slim 모드 기본)
    skills:
      - name: moai-workflow-tdd
        tier: core
        path: templates/.claude/skills/moai-workflow-tdd/
        hash: 6f89fb72...      # 콘텐츠 해시 (무결성 검증)
        version: 1.0.0
    agents:
      - name: manager-spec
        tier: core
        path: templates/.claude/agents/moai/manager-spec.md
        hash: a1b2c3d4...
        version: 1.0.0
  optional_packs:              # --all 플래그 시 배포
    backend:
      - name: moai-domain-backend
        tier: optional-pack:backend
        path: templates/.claude/skills/moai-domain-backend/
        hash: ...
    frontend:
      - name: moai-domain-frontend
        tier: optional-pack:frontend
        path: templates/.claude/skills/moai-domain-frontend/
        hash: ...
  harness_generated:           # --all 플래그 시 배포
    skills: []
    agents:
      - name: builder-harness
        tier: harness-generated
        path: templates/.claude/agents/moai/builder-harness.md
        hash: ...
```

각 엔트리는 `name`, `tier`, `path`, `hash`, `version` 필드로 이뤄집니다. `hash` 필드가 콘텐츠 해시를 담고 있어 배포된 파일이 손상되거나 임의로 바뀌었는지 로더가 검증할 수 있습니다. 스킬 디렉토리 내부의 진입점 파일은 `SKILL.md`입니다 (소문자 `skill.md`가 아님).

## Slim 모드와 --all 플래그

기본 배포는 **slim 모드**로 `catalog.core`만 배포합니다. 전체 배포가 필요하면 `--all` 플래그 또는 `MOAI_DISTRIBUTE_ALL=1` 환경변수를 사용합니다.

```bash
# Slim 설치 (기본 — core만)
moai init my-project

# 전체 설치 (core + optional_packs + harness_generated)
moai init --all my-project

# 환경변수로 전체 설치
MOAI_DISTRIBUTE_ALL=1 moai init my-project
```

### 배포 로직

배포는 두 단계로 동작합니다.

1. `catalog.core` (skills + agents)는 항상 포함 — slim 모드의 기본
2. `--all` 플래그 또는 `MOAI_DISTRIBUTE_ALL=1` 환경변수가 설정된 경우 `catalog.optional_packs`와 `catalog.harness_generated`를 추가 배포

## Typed Loader

`LoadCatalog()` 함수가 매니페스트를 타입 안전하게 로드합니다. 문자열 파싱에 의존하지 않고 구조체 단위로 검증하기 때문에 매니페스트 오류는 배포 전에 걸러집니다.

- 3계층 분류 검증
- 해시 무결성 검사 (Hash Sentinel)
- 누락 필드 감지
- 100% 테스트 커버리지

## 카탈로그 활용

### 프로젝트 초기화

```bash
# 기본 초기화 — core만 배포 (slim 모드)
moai init my-project

# 전체 초기화 — core + optional_packs + harness_generated
moai init --all my-project
```

### 업데이트

`moai update`는 동일한 카탈로그를 기준으로 동작합니다. slim으로 초기화한 프로젝트는 core만, `--all`로 초기화한 프로젝트는 전체를 업데이트합니다.

```bash
# 카탈로그 기반 업데이트
moai update                  # 초기화 모드에 따라 자동 결정
```

## 관련 문서

- [설치](/ko/getting-started/installation) — 설치 가이드
- [초기 설정](/ko/getting-started/init-wizard) — init 마법사
- [업데이트](/ko/cli-reference/update) — 업데이트 가이드
- [스킬 가이드](/ko/advanced/skill-guide) — 스킬 작성 가이드
