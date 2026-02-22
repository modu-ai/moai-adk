# SPEC-UPDATE-001: Global/Local Installation Mode Redesign

```yaml
---
spec_id: SPEC-UPDATE-001
title: Global/Local Installation Mode Redesign
created: 2026-02-21
status: Planned
priority: High
assigned: manager-ddd
related_specs: []
epic: moai-adk-core
labels:
  - cli
  - installation
  - configuration
  - breaking-change
---
```

## Problem Analysis

### Current State

MoAI-ADK는 모든 에이전트, 스킬, 룰을 프로젝트 로컬 .claude/ 디렉토리에 설치합니다. 이로 인해 다음과 같은 문제가 발생합니다:

1. 중복 설치: 여러 프로젝트에서 동일한 MoAI 시스템 파일이 반복 설치됨
2. 버전 불일치: 프로젝트마다 다른 버전의 MoAI 에이전트/스킬을 사용할 수 있음
3. 업데이트 복잡성: 각 프로젝트에서 개별적으로 moai update를 실행해야 함
4. 저장 공간 낭비: 동일한 파일이 여러 프로젝트에 복제됨

### Desired State

1. Global Mode (권장): 공통 MoAI 시스템 파일을 ~/.claude/에 설치하여 프로젝트 간 공유
2. Local Mode: 프로젝트 격리가 필요한 경우 로컬 설치 유지
3. 명확한 파일 분리: 시스템 파일 vs 프로젝트 설정 파일의 명확한 구분
4. 간소화된 업데이트: Global 모드에서는 단일 위치 업데이트로 모든 프로젝트 반영

---

## EARS Requirements

### Ubiquitous Requirements (항상 적용)

REQ-001: 시스템은 모든 설치 모드에서 .moai/ 디렉토리 구조를 유지해야 한다.

REQ-002: 시스템은 모든 설치 모드에서 .claude/hooks/moai/를 프로젝트 로컬에 배치해야 한다.

REQ-003: 시스템은 모든 설치 모드에서 .claude/settings.json을 프로젝트 로컬에 유지해야 한다.

REQ-004: 시스템은 Global 모드에서도 프로젝트별 독립적인 GLM 설정을 지원해야 한다.

---

### Event-Driven Requirements (이벤트 기반)

REQ-010: WHEN 사용자가 moai init을 실행하면 THEN 시스템은 .moai/ 디렉토리 존재 여부를 확인해야 한다.

REQ-011: WHEN .moai/ 디렉토리가 없으면 THEN 시스템은 "moai init . 으로 초기화하세요" 안내 메시지를 표시해야 한다.

REQ-012: WHEN 사용자가 처음 프로젝트를 초기화하면 THEN 시스템은 AskUserQuestion을 통해 설치 모드를 선택해야 한다.

REQ-013: WHEN 사용자가 Global 모드를 선택하면 THEN 시스템은 ~/.claude/agents/moai/, ~/.claude/skills/moai*/, ~/.claude/rules/moai/에 시스템 파일을 설치해야 한다.

REQ-014: WHEN 사용자가 Local 모드를 선택하면 THEN 시스템은 .claude/ 하위에 모든 파일을 설치해야 한다.

REQ-015: WHEN 사용자가 moai update를 실행하면 THEN 시스템은 "설정을 재설정하시겠습니까?" 질문을 표시해야 한다.

REQ-016: WHEN 사용자가 재설정 질문에 'y'로 응답하면 THEN 시스템은 설정 wizard를 재실행해야 한다.

REQ-017: WHEN 사용자가 재설정 질문에 'n'으로 응답하면 THEN 시스템은 기존 설정을 백업 후 병합해야 한다.

REQ-018: WHEN moai update 실행 시 THEN 시스템은 Global 모드 변경 질문을 추가해야 한다.

REQ-019: WHEN 사용자가 moai glm을 실행하면 THEN 시스템은 .claude/settings.local.json만 수정해야 한다.

REQ-020: WHEN Global 모드에서 Local 모드로 전환하면 THEN 시스템은 프로젝트 .claude/에 시스템 파일을 복사해야 한다.

REQ-021: WHEN Local 모드에서 Global 모드로 전환하면 THEN 시스템은 .claude/agents/moai/, .claude/skills/moai*/, .claude/rules/moai/를 삭제해야 한다.

---

### State-Driven Requirements (상태 기반)

REQ-030: IF 프로젝트가 Global 모드로 설정되어 있으면 THEN .claude/ 디렉토리에는 hooks, settings.json, settings.local.json만 존재해야 한다.

REQ-031: IF Global 모드에서 ~/.claude/에 시스템 파일이 없으면 THEN 시스템은 자동으로 Global 설치를 수행해야 한다.

REQ-032: IF settings.local.json이 존재하면 THEN 시스템은 이를 settings.json보다 우선적으로 로드해야 한다.

---

### Unwanted Behavior Requirements (금지 사항)

REQ-040: 시스템은 Global 모드에서 ~/.claude/settings.json을 생성하거나 수정해서는 안 된다.

REQ-041: 시스템은 moai glm 실행 시 .claude/settings.json을 수정해서는 안 된다.

REQ-042: 시스템은 Global 모드 전환 시 .claude/hooks/, .claude/settings.json, .claude/settings.local.json, .moai/를 삭제해서는 안 된다.

---

### Optional Requirements (선택 사항)

REQ-050: WHERE 가능하면 Global 모드를 기본값으로 제안해야 한다.

---

## File Placement Strategy

### Global Mode (~/.claude/)

~/.claude/
  agents/
    moai/              # MoAI 시스템 에이전트
      manager-spec.md
      manager-ddd.md
      expert-backend.md
      ...
  skills/
    moai/              # MoAI 코어 스킬
    moai-foundation-core/
    moai-foundation-claude/
    moai-workflow-project/
    moai*/             # 모든 moai-* 스킬
  rules/
    moai/              # MoAI 시스템 룰
      core/
      workflow/
      development/

### Project (모든 모드)

project/
  .claude/
    hooks/
      moai/          # 항상 프로젝트 로컬
    settings.json      # 항상 프로젝트 로컬 (팀 공유)
    settings.local.json # 항상 프로젝트 로컬 (개인 설정)
  .moai/                 # 항상 프로젝트 로컬
    config/
    specs/
    project/
    ...

### Local Mode (.claude/)

project/.claude/
  agents/
    moai/              # 로컬 복사본
  skills/
    moai/              # 로컬 복사본
    moai*/             # 로컬 복사본
  rules/
    moai/              # 로컬 복사본
  hooks/
    moai/
  settings.json
  settings.local.json

---

## Configuration Schema

### .moai/config/sections/system.yaml 추가 필드

moai:
  version: "v2.5.0"
  template_version: "v2.5.0"
  installation_mode: "global"  # "global" | "local"
  global_path: "~/.claude"     # Global 설치 경로 (읽기 전용)

---

## Technical Constraints

### Constraint-001: Backward Compatibility

- 기존 Local 모드 프로젝트는 자동으로 Local 모드로 유지됨
- 마이그레이션은 명시적 사용자 선택 시에만 수행

### Constraint-002: Multi-Platform Support

- macOS, Linux, Windows에서 동일한 동작 보장
- 경로 처리는 filepath 패키지 사용

### Constraint-003: Git Integration

- settings.local.json은 .gitignore에 추가 권장
- Global 설치 파일은 Git에서 제외

---

## Security Considerations

1. API Key 격리: GLM API 키는 프로젝트별 settings.local.json에만 저장
2. Global 파일 보호: ~/.claude/ 하위 파일은 읽기 전용으로 권장
3. 설정 병합 안전성: 백업 후 병합으로 데이터 손실 방지

---

## Traceability

| Requirement ID | Source | Implementation Target |
|----------------|--------|----------------------|
| REQ-001~004 | Problem Analysis #1 | deployer.go, init.go |
| REQ-010~021 | Feature Description | init.go, update.go, glm.go |
| REQ-030~032 | State Management | config/types.go, deployer.go |
| REQ-040~042 | Safety Constraints | update.go, glm.go |
| REQ-050 | UX Improvement | wizard/run.go |

---

Version: 1.0.0
Last Updated: 2026-02-21
