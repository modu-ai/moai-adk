---
title: 카탈로그 시스템
weight: 80
draft: false
---

토크노믹스는 토큰에만 적용되는 원칙이 아닙니다. 프로젝트에 배포되는 템플릿 파일 하나하나도 결국 세션이 로드하게 될 컨텍스트 후보입니다. 카탈로그 시스템은 "필요한 것만 배포한다"는 원칙으로 이 비용을 초기화 단계에서부터 줄입니다.

## 개요

MoAI-ADK v2.15+의 카탈로그 시스템은 모든 에이전트, 스킬, 플러그인, 규칙을 **3계층 매니페스트**로 관리합니다. `moai init --slim`을 사용하면 프로젝트에 필요한 최소 템플릿만 골라 배포하므로 초기화가 빨라지고, 프로젝트에 남는 파일도 가벼워집니다.

## 3계층 매니페스트

모든 배포 대상은 세 계층 중 하나에 속합니다.

| 계층 | 설명 | 배포 기준 |
|------|------|----------|
| **Tier 1 (Core)** | 핵심 인프라 — 오케스트레이터, 품질 게이트, 기본 스킬 | 항상 배포 |
| **Tier 2 (Standard)** | 표준 확장 — 언어별 규칙, 프레임워크 스킬 | 프로젝트 언어/프레임워크 감지 시 |
| **Tier 3 (Optional)** | 선택적 — 도메인 스킬, 플랫폼별 설정 | 명시적 요청 또는 프로젝트 설정 시 |

## 카탈로그 파일

카탈로그 매니페스트는 YAML 형식으로 정의됩니다.

```yaml
# 카탈로그 엔트리 예시
- id: moai-workflow-tdd
  tier: 1                    # 1=Core, 2=Standard, 3=Optional
  type: skill
  path: .claude/skills/moai/workflows/tdd.md
  languages: []              # 빈 배열 = 모든 언어
  frameworks: []
  hash: abc123...             # 콘텐츠 해시 (무결성 검증)
```

`hash` 필드가 콘텐츠 해시를 담고 있어, 배포된 파일이 손상되거나 임의로 바뀌었는지 로더가 검증할 수 있습니다.

## SlimFS 필터

`moai init --slim`은 SlimFS 필터를 통해 배포 파일을 제한합니다.

```bash
# 전체 설치 (모든 계층)
moai init my-project

# Slim 설치 (Tier 1 + 감지된 Tier 2만)
moai init --slim my-project
```

### 필터 로직

필터는 네 단계로 동작합니다.

1. Tier 1은 항상 포함
2. 프로젝트 언어 감지 (Go, Python, TypeScript 등)
3. 감지된 언어에 해당하는 Tier 2 항목만 포함
4. Tier 3은 제외

## Typed Loader

`LoadCatalog()` 함수가 매니페스트를 타입 안전하게 로드합니다. 문자열 파싱에 의존하지 않고 구조체 단위로 검증하기 때문에, 매니페스트 오류는 배포 전에 걸러집니다.

- 3계층 분류 검증
- 해시 무결성 검사 (Hash Sentinel)
- 누락 필드 감지
- 100% 테스트 커버리지

## 카탈로그 활용

### 프로젝트 초기화

```bash
# 일반 초기화 — 모든 템플릿 배포
moai init my-project

# Slim 초기화 — 최소 템플릿만 배포
moai init --slim my-project
```

### 업데이트

업데이트도 같은 카탈로그를 기준으로 동작하므로, slim으로 초기화한 프로젝트는 slim으로 업데이트하면 됩니다.

```bash
# 카탈로그 기반 업데이트
moai update                  # 모든 계층 업데이트
moai update --slim           # slim 모드로 업데이트
```

## 관련 문서

- [설치](/ko/getting-started/installation) — 설치 가이드
- [초기 설정](/ko/getting-started/init-wizard) — init 마법사
- [업데이트](/ko/getting-started/update) — 업데이트 가이드
- [스킬 가이드](/ko/advanced/skill-guide) — 스킬 작성 가이드
