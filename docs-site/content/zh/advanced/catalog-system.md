---
title: 카탈로그 시스템
weight: 80
draft: false
---

3계층 카탈로그 매니페스트와 slim init으로 프로젝트 초기화를 최적화합니다.

## 개요

MoAI-ADK v2.15+의 카탈로그 시스템은 모든 에이전트, 스킬, 플러그인, 규칙을 **3계층
매니페스트**로 관리합니다. `moai init --slim`을 통해 프로젝트에 필요한 최소 템플릿만
배포하여 초기화 시간을 단축합니다.

## 3계층 매니페스트

| 계층 | 설명 | 배포 기준 |
|------|------|----------|
| **Tier 1 (Core)** | 핵심 인프라 — 오케스트레이터, 품질 게이트, 기본 스킬 | 항상 배포 |
| **Tier 2 (Standard)** | 표준 확장 — 언어별 규칙, 프레임워크 스킬 | 프로젝트 언어/프레임워크 감지 시 |
| **Tier 3 (Optional)** | 선택적 — 도메인 스킬, 플랫폼별 설정 | 명시적 요청 또는 프로젝트 설정 시 |

## 카탈로그 파일

카탈로그 매니페스트는 YAML 형식으로 정의됩니다:

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

## SlimFS 필터

`moai init --slim`은 SlimFS 필터를 통해 배포 파일을 제한합니다:

```bash
# 전체 설치 (모든 계층)
moai init my-project

# Slim 설치 (Tier 1 + 감지된 Tier 2만)
moai init --slim my-project
```

### 필터 로직

1. Tier 1은 항상 포함
2. 프로젝트 언어 감지 (Go, Python, TypeScript 등)
3. 감지된 언어에 해당하는 Tier 2 항목만 포함
4. Tier 3은 제외

## Typed Loader

`LoadCatalog()` 함수가 매니페스트를 타입 안전하게 로드합니다:

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
