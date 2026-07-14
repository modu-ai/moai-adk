---
title: 자율 CI/CD 가이드
weight: 10
draft: false
---

MoAI-ADK의 자율 CI/CD 시스템은 풀리퀘스트 품질을 자동으로 관리합니다.
로컬 세션에서 `/moai loop`가 하던 "진단 → 수정 → 검증" 루프를 CI까지
연장한 것으로, 개발자가 수동으로 품질을 검증하지 않아도 CI가 스스로
품질을 보장합니다 — 에이전틱 루프 엔지니어링을 저장소 수준에 적용한
사례입니다.

## 개요

SPEC-V3R3-CI-AUTONOMY-001에서 도입된 자율 CI/CD 시스템은 8개 티어로 구성된
품질 자동화 인프라입니다. push 전 로컬 검증(pre-push hook)부터 CI 실패 시
자동 수정(auto-fix loop)까지 하나의 방어선으로 이어집니다.

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

push 전에 로컬에서 자동으로 품질 검증을 실행합니다. CI까지 갔다가 실패하고
돌아오는 왕복 비용을 로컬에서 미리 끊는 첫 번째 방어선입니다.

```bash
# 자동 설치됨 (moai init / moai update 시)
.git/hooks/pre-push → moai hook pre-push
```

실행되는 검증:

- `go vet` / `golangci-lint` (프로젝트 언어에 따라 자동 감지)
- `go test ./...` (테스트 스위트)
- MX 태그 무결성 검사

## Auto-fix Loop (T3)

`/moai sync`가 PR을 생성한 뒤, CI 감시 스크립트와 CI 루프 스킬이 함께
"진단 → 수정 → 재검증" 루프를 돌립니다. 로컬의 진단형 자가 수정 루프를
PR 파이프라인 위로 연장한 구조입니다.

**CI 감시 스크립트 (`scripts/ci-watch/run.sh`)**

```bash
sh scripts/ci-watch/run.sh <PR_NUMBER> [BRANCH]
```

- 30초 간격으로 `gh pr checks`를 폴링하며 필수(required) 체크와
  보조(auxiliary) 체크를 분류합니다
- 종료 코드: `0` 전체 통과 · `2` 필수 체크 실패(구조화 JSON 핸드오프를
  stdout으로 출력) · `3` 30분 하드 타임아웃 · `1` 오류
- required 체크 목록은 SSoT 파일에서 읽으며, 테스트용 환경변수
  오버라이드(`MOAI_CIWATCH_GH`, `CIWATCH_TIMEOUT_SECONDS` 등)를 지원합니다

**CI 루프 스킬 (`moai-workflow-ci-loop`)**

감시 스크립트가 필수 실패를 핸드오프하면, `moai-workflow-ci-loop` 스킬이
실패를 분류하고 안전한 자동 패치를 최대 3회까지 시도합니다. 의미 수준의
실패(자동 수정이 위험한 경우)는 사용자에게 에스컬레이션합니다.

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
결정을 추측이 아닌 기록으로 남기는 것 — 증거 기반 완료 판정이라는 MoAI
원칙이 브랜치 결정에도 적용됩니다.

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
