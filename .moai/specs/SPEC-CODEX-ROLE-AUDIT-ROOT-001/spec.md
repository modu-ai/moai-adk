---
id: SPEC-CODEX-ROLE-AUDIT-ROOT-001
title: "codex_role_audit caller-tree verification repair — presented-root registration check replaces the server-cwd equality, optional CLI role-audit verb, codex_task project_root"
version: "0.1.1"
status: in-progress
created: 2026-09-29
updated: 2026-09-29
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "codex, mcp, role-audit, worktree, project-root, lane"
tier: M
card: t1324
related_specs:
  - SPEC-CODEX-AUDIT-READONLY-001  # owns the read-only role launcher this SPEC repairs the root gate of
  - SPEC-CODEX-AUDIT-GATE-AXES-001
---

# SPEC-CODEX-ROLE-AUDIT-ROOT-001

## HISTORY

- v0.1.0 (2026-09-29): plan-phase authoring by manager-spec (card t1324, Class C). Cause analysis was lead-verified before delegation; this SPEC designs the repair, not the investigation.
- v0.1.1 (2026-09-29): plan-audit iter1 repairs (D1-D6) — D1 marker resolved in place by measured caller sweep (super-advisor is the only shipped caller; its update requirement added to M2), D2 trace map aligned to acceptance.md §D.2, D3 AC-003 gating observable added, D4 dangling related_specs entry removed, D5 live-stdio-session requirement added to M0 pre-flight, D6 anchor citation aligned to 293-338.

## §A 배경과 목적

`codexAuditValidateRoot`(`internal/cli/codex_audit_launch.go:293-338`, HEAD `7ca01df35`)은 요청된 `root`가 심볼릭 링크 해석 뒤 `callerTop` — `callerDir`에서 실행한 `git rev-parse --show-toplevel` — 과 **동일할 때만** 받는다. MCP stdio 호출에서는 `callerDir`가 요청 세션의 cwd가 아니라 MCP 서버 프로세스의 고정 cwd(통상 primary 체크아웃)이므로, 서버가 primary에서 시작된 한 카드 워크트리를 소유한 레인도 자기 트리를 제시하면 거부된다. 관측된 거부문: `<root> is not the caller's own worktree (/Users/goos/MoAI/moai-adk-go)`. 주석이 서술하는 설계 의도("the caller's own worktree; a sibling worktree or the primary checkout is refused even though it is registered")는 세션 신원을 알 수 없는 공유 서버에서 **불가능한 계약**이다.

카드가 추가로 확인한 축소 표면: (1) `codex_task`에는 `project_root` 인수가 없어 primary에 고정돼 있다; (2) CLI에는 role-audit 동사가 없어(`moai codex audit` 만 존재) 레인이 MCP를 거치지 않고 감사를 직접 실행할 경로가 없다.

교체 패턴은 `.claude/rules/moai/core/moai-mcp-tools.md` § The `project_root` input 이 규정하는 `project_root`-family 규약이다: 호출자가 **자기 토플레벨을 명시적으로 제시**하고, 도구는 그것이 **같은 저장소의 등록된 linked worktree**임을 검증한다(저장소 경계 밖 경로는 기각). 심볼릭 링크 정규화와 기존 동일-저장소 검증(git-common-dir 비교 + `worktree list --porcelain` 등록 확인)은 유지한다. 명시적 루트 없는 호출은 절대 조용히 primary에 대해 동작하지 않고 거부된다.

**Tier 판정**: M — 한 함수 수리를 축으로 하되 세 표면(검증기+MCP 스키마, CLI 신규 동사, `codex_task` 인수)에 걸치고 거부 계약이라는 사용자 대면 동작 변화가 있어 별도 acceptance.md가 필요하다. LOC 가이드(<300)로는 Tier S 범위지만 산출물 집합과 표면 수가 Tier M을 가리킨다. 요구 상한 16 / AC 상한 16.

## §C 요구사항 (GEARS)

- REQ-001: **When** the `codex_role_audit` tool receives a request presenting an explicit `worktree_root`, the root validator shall accept the presented root if, after symlink resolution, it is a registered linked worktree of the same repository as the serving checkout — verified by the git-common-dir comparison and the registered `worktree list --porcelain` check — and shall not require the presented root to equal the toplevel derived from the server process's start directory.
- REQ-002: **When** a `codex_role_audit` request presents no explicit `worktree_root` (absent or empty), the tool shall refuse with a structured error naming the missing argument and shall not act on any default tree, including the primary checkout.
- REQ-003: The root validator shall keep the same-repository verification semantics unchanged: a root whose resolved git common directory differs from the serving checkout's shall be refused as belonging to a different repository, and a root absent from the serving checkout's registered worktree list shall be refused as unregistered.
- REQ-004: **Where** the CLI exposes a role-audit verb, a session running from its own linked-worktree working directory shall be able to launch the read-only role audit directly against its own tree without the MCP server, and the CLI shall apply the same root-verification and refusal semantics as the MCP path.
- REQ-005: **Where** the `codex_task` tool accepts a `project_root` argument, the same registered-same-repository verification shall gate that argument; **When** a `codex_task` request omits the argument, the tool shall refuse with a structured error rather than quietly acting on the primary checkout.
- REQ-006: The role-audit launch surface shall keep every existing non-root refusal behavior unchanged: verdict and record destinations stay contained under the accepted root's report tree, the read-only role catalogue and sandbox flags are untouched, and the launch-record format is untouched.

## §D 추적

- REQ-001→AC-001 · REQ-002→AC-006 · REQ-003→AC-004/AC-005 · REQ-004→AC-002 · REQ-005→AC-003 · REQ-006→AC-007 (acceptance.md §D.2와 동일; 플립 마일스톤은 각 AC의 green-path 셀 참조).
- RED 기준선의 측정 귀속은 acceptance.md 각 셀에 명시한다 — 리드 검증 실측(카드 본문)과 본 워크트리 직접 측정(이진 빌드 `@7ca01df35`)을 구분한다.
- 계약 문서 정합: `.claude/rules/moai/core/moai-mcp-tools.md` § The `project_root` input — 본 SPEC은 그 규약을 role-audit 표면으로 확장한다. 규약 문서 갱신은 sync-phase 소관이다.

## §E 범위 밖

### Out of Scope — 세션 신원 전파

- MCP 서버가 요청 세션의 신원이나 cwd를 제시 루트 외의 경로로 학습하는 메커니즘은 만들지 않는다 — 공유 서버 모델에서 호출자가 자기 트리를 명시적으로 제시하는 것으로 충분하다.

### Out of Scope — 감사 내용과 역할 정의

- 역할 카탈로그, 읽기전용 샌드박스 플래그, launch-record/verdict 파일 형식, 타임아웃 정책은 변경하지 않는다.

### Out of Scope — 형제 위임 표면

- `glm_task`와 GLM 감사 표면에는 대응 변경을 하지 않는다(후속 카드 후보).
- `codex_job_control` / `codex_job_status` 계열의 루트 처리는 다루지 않는다.

### Out of Scope — 규약 문서 갱신

- `.claude/rules/moai/core/moai-mcp-tools.md` 본문 갱신은 sync-phase 문서 작업이다.
