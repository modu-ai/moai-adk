---
title: 자율 CI/CD 가이드
weight: 10
draft: false
---

MoAI-ADK의 자율 CI/CD 시스템은 풀 리퀘스트 품질을 자동으로 관리합니다.
로컬 세션에서 `/moai loop`가 돌리던 "진단 → 수정 → 검증" 루프를 CI까지
늘린 것이라, 개발자가 일일이 확인하지 않아도 CI가 알아서 품질을 지킵니다.
에이전틱 루프 엔지니어링을 저장소 단위로 적용한 사례입니다.

## 개요

SPEC-V3R3-CI-AUTONOMY-001에서 도입한 자율 CI/CD 시스템은 8개 티어짜리
품질 자동화 인프라입니다. push 전 로컬 검증(pre-push hook)에서 시작해
CI 실패 시 자동 수정(auto-fix loop)까지 하나의 방어선으로 이어집니다.

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

push 전에 로컬에서 품질 검증을 자동으로 돌립니다. CI까지 갔다가 실패하고
돌아오는 왕복 비용을 로컬에서 미리 끊어 주는 첫 번째 방어선입니다.

```bash
# 자동 설치됨 (moai init / moai update 시)
.git/hooks/pre-push → moai hook pre-push
```

실행되는 검증:

- `go vet` / `golangci-lint` (프로젝트 언어에 따라 자동 감지)
- `go test ./...` (테스트 스위트)
- MX 태그 무결성 검사

## Auto-fix Loop (T3)

`/moai sync`가 PR을 만든 뒤, 오케스트레이터가 실패한 필수(required) 체크를
넘겨주면 `manager-develop`이 `cycle_type=autofix` 사이클로 "진단 → 수정 →
재검증" 루프를 돌립니다. 로컬에서 쓰던 진단형 자가 수정 루프를 PR
파이프라인까지 늘린 구조입니다.

- **진입 조건** — 실패한 필수 체크가 하나 이상 있고, 오케스트레이터가 해당
  PR과 브랜치를 지목해 넘겨줄 때만 루프가 시작됩니다. 오케스트레이터가
  유일한 진입점입니다
- **반복 상한** — PR push 한 번당 최대 3회. 4회째로 넘어가면 자동 패치를
  시도하지 않고 blocking AskUserQuestion으로 사용자에게 넘깁니다
- **의미 수준 실패** — data race, deadlock, panic, 테스트 단언 실패는 자동으로
  고치지 않고 사용자 판단으로 넘깁니다
- **보호 파일** — 비밀·자격 증명 파일과 CI 워크플로 정의는 루프가 절대
  건드리지 않습니다. 실패를 보고하는 계층을 고치면 진짜 실패가 가짜 green이
  되기 때문입니다

반복 상한, 에스컬레이션 계약, 의미 수준 실패 처리, 보호 파일 목록의 SSoT는
`.claude/rules/moai/workflow/ci-autofix-protocol.md`입니다.

## BODP — Branch Origin Decision Protocol (T7)

새 브랜치나 워크트리를 만들 때 base branch를 자동으로 골라 줍니다.

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

모든 BODP 결정은 `.moai/branches/decisions/<branch-name>.md`에 남습니다.
추측이 아니라 기록으로 남기는 셈이며, 증거로 완료를 판정한다는 MoAI
원칙이 브랜치를 고를 때도 그대로 적용됩니다.

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
- [GitHub 연동 가이드](/ko/guides/multi-llm-ci) — 이슈 파싱·SPEC 링크
