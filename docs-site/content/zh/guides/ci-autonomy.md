---
title: 자율 CI/CD 가이드
weight: 10
draft: false
---

MoAI-ADK의 자율 CI/CD 시스템으로 풀리퀘스트 품질을 자동으로 관리합니다.

## 개요

SPEC-V3R3-CI-AUTONOMY-001에서 도입된 자율 CI/CD 시스템은 8개 티어로 구성된
품질 자동화 인프라입니다. pre-push hook부터 auto-fix 루프까지, 개발자가 수동으로
품질을 검증할 필요 없이 CI가 자동으로 품질을 보장합니다.

## 8-Tier 아키텍처

| Tier | 이름 | 우선순위 | 설명 |
|------|------|----------|------|
| T1 | Pre-push Hook | P0 | push 전 자동 품질 검증 |
| T2 | Branch Protection | P0 | main 브랜치 보호 규칙 |
| T3 | Auto-fix Loop | P1 | CI 실패 시 자동 수정 |
| T4 | Auxiliary Workflows | P2 | 보조 워크플로우 정리 |
| T5 | Worktree State Guard | P1 | 워크트리 상태 무결성 보장 |
| T6 | i18n Validator | P2 | 4개국어 문서 일관성 검증 |
| T7 | BODP | P0 | 브랜치 원점 결정 프로토콜 |
| T8 | Release Workflow | P1 | 릴리스 자동화 |

## Pre-push Hook (T1)

push 전에 로컬에서 자동으로 품질 검증을 실행합니다.

```bash
# 자동 설치됨 (moai init / moai update 시)
.git/hooks/pre-push → moai hook pre-push
```

실행되는 검증:

- `go vet` / `golangci-lint` (프로젝트 언어에 따라 자동 감지)
- `go test ./...` (테스트 스위트)
- MX 태그 무결성 검사

## Auto-fix Loop (T3)

CI 실패 시 `/moai loop`를 자동으로 호출하여 에러를 수정합니다.

```yaml
# .github/workflows/ci.yml (자동 생성)
- name: Auto-fix on failure
  if: failure()
  run: |
    claude -p "/moai loop --max-iterations 3"
```

## BODP — Branch Origin Decision Protocol (T7)

새 브랜치/워크트리를 생성할 때 base branch를 자동으로 결정합니다.

### 3-Signal 평가

| 시그널 | 출처 | 의미 |
|--------|------|------|
| Signal A | SPEC `depends_on` + diff path overlap | 코드 의존성 |
| Signal B | `git status`에서 `.moai/specs/<NewSpecID>/` 매칭 | 작업 트리 동위치 |
| Signal C | `gh pr list --head <branch> --state open` ≥ 1 | 현재 브랜치 PR |

### 결정 매트릭스

| 시그널 | 결정 |
|--------|------|
| A만 있음 | `stacked` — 현재 브랜치 기반 |
| B 있음 | `continue` — 현재 컨텍스트에서 계속 |
| C만 있음 | `stacked` — 현재 브랜치 기반 |
| 아무것도 없음 | `main` — origin/main 기반 |

### 감사 추적

모든 BODP 결정은 `.moai/branches/decisions/<branch-name>.md`에 기록됩니다.

## i18n Validator (T6)

4개국어 문서의 일관성을 자동 검증합니다.

```bash
scripts/docs-i18n-check.sh
```

검증 항목:

- 4개 locale 간 파일 개수/경로 일치
- front matter `title` 존재
- H1 heading 존재
- MoAI 용어집 준수

## Worktree State Guard (T5)

워크트리의 상태 무결성을 보장합니다:

- 커밋되지 않은 변경 감지
- 워크트리와 메인 브랜치 동기화 상태 확인
- `moai status`에서 상태 표시

## 관련 문서

- [워크트리 가이드](/ko/worktree/guide) — Git Worktree 완벽 가이드
- [/moai loop](/ko/utility-commands/moai-loop) — 반복 수정 루프
- [/moai fix](/ko/utility-commands/moai-fix) — 자동 에러 수정
- [멀티 LLM CI](/ko/guides/multi-llm-ci) — Multi-LLM CI 통합
