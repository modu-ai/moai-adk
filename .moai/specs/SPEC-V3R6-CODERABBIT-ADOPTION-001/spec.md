---
id: SPEC-V3R6-CODERABBIT-ADOPTION-001
title: "CodeRabbit PR-review adoption — disable claude-pr-review, retain @claude interactive"
version: "0.1.0"
status: draft
created: 2026-07-29
updated: 2026-07-29
author: MoAI orchestrator
priority: P2
phase: "v3.0.2"
module: ".github/workflows"
lifecycle: spec-anchored
tags: "coderabbit, code-review, ci, github-actions"
tier: S
---

## §A. 배경 (Context)

moai-adk는 퍼블릭 OSS 저장소(`visibility: PUBLIC`)로 CodeRabbit 무료 플랜 대상이다. `.coderabbit.yaml`이 이미 존재하며 Go 모범사례·path별 지침·`auto_review`(main/moai-go-v2/release/*)·라벨링·knowledge_base로 잘 세팅되어 있다(`language: en-US` — 유지).

그러나 CodeRabbit이 최근 PR(#1203, #1207)을 리뷰하지 않았다(리뷰/코멘트 0건). 과거 13건의 상호작용 이력이 있어 App은 한때 활성이었으나, 현재는 App 토큰/권한 만료 또는 비활성로 추정된다. 대신 `.github/workflows/claude.yml`의 Job A(`claude-pr-review`, claude-code-action)가 자문 리뷰를 담당해 왔으나, #1207에서 action 에러(OIDC/인증)로 실패했다(비필수 체크라 머지는 무관했지만 리뷰 미제공).

## §B. 목표 (Goal)

CodeRabbit을 체계적인 PR 리뷰 도구로 단일화하고, 중복/고장난 `claude-pr-review`(Job A)를 비활성화한다. 단, 이슈/PR `@claude` 트러블슈팅(Job B, `claude-interactive`)은 CodeRabbit이 커버하지 않는 기능이므로 유지한다.

## §C. 범위 (Scope)

### 대상 변경 (repo)
- `.github/workflows/claude.yml` — `on:` 블록에서 `pull_request:` 트리거 제거 → Job A(`claude-pr-review`)가 더 이상 트리거되지 않음. Job B(`claude-interactive`)는 `issue_comment` + `issues`로 유지(PR 코멘트도 `issue_comment` 이벤트이므로 `@claude` 동작 무영향).

### 비대상 (변경 없음)
- `.coderabbit.yaml` — 이미 잘 세팅됨, `language: en-US` 유지(GOOS 결정). CodeRabbit `language`는 저장소 단위 고정값(작성자 자동 감지 미지원)이므로 변경하지 않는다.
- `claude-interactive`(Job B) — 유지.
- CodeRabbit App 설치/재활성화 — repo 변경이 아닌 **GOOS 액션**(GitHub App 설정/대시보드).

### CodeRabbit App 재활성화 절차 (GOOS 액션, repo 외)
1. https://github.com/apps/coderabbitai → Configure (또는 Install)
2. `modu-ai/moai-adk`(또는 modu-ai 조직) 접근 권한 부여/확인
3. 이미 설치된 경우 https://app.coderabbit.ai 대시보드에서 리포 활성 + 토큰/권한 만료 확인
4. 활성 후 신규 PR에 CodeRabbit이 `.coderabbit.yaml` 기반 자동 리뷰

## §D. 요구사항 / AC (GEARS)

- **REQ-CRA-001** (Unwanted) 본 변경은 `claude-pr-review` 잡이 `pull_request` 이벤트로 트리거되게 해서는 안 된다(shall not) — `on:` 블록에서 `pull_request:` 트리거가 제거되어야 한다.
- **REQ-CRA-002** (Ubiquitous) `claude-interactive` 잡(Job B)은 `issue_comment` + `issues` 트리거로 보존되어야 한다(shall) — `@claude` 멘션 트러블슈팅이 PR 코멘트에서도 동작한다.
- **REQ-CRA-003** (Ubiquitous) `.coderabbit.yaml`은 본 변경에서 수정되지 않는다(shall) — 기존 `en-US` 설정 + Go path 지침이 그대로 유지된다.
- **AC-CRA-001** `grep -A3 '^on:' .github/workflows/claude.yml` 출력에 `pull_request:` 가 없다 → exit 0.
- **AC-CRA-002** `grep -n 'claude-interactive' .github/workflows/claude.yml` 이 Job B 정의를 반환 + `issue_comment:` / `issues:` 트리거가 `on:` 에 존재.
- **AC-CRA-003** `git diff` 가 `.coderabbit.yaml`을 포함하지 않는다(변경 없음).
- **AC-CRA-004** (GOOS 액션, 관측 지연) CodeRabbit App 재활성화 후 최초 신규 PR에서 CodeRabbit(coderabbitai) 리뷰가 게시된다 — run-phase에서는 관측 불가, sync/운영 단계에서 확인.

## §E. Gaps (미검증)
1. CodeRabbit App의 현재 활성 상태는 본 SPEC에서 제어하지 못한다(GOOS 대시보드 액션). App이 재활성화되지 않으면 AC-CRA-004는 관측되지 않는다.
2. CodeRabbit `language` 자동 감지(작성자 기반)는 CodeRabbit이 네이티브 미지원 — 저장소 고정값(`en-US`)으로 운영.
