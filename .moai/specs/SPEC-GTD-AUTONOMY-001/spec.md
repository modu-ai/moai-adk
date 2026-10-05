---
id: SPEC-GTD-AUTONOMY-001
title: "GTD 지식 그래프와 사전 위임 기반 자율 목표 운영"
version: "0.2.0"
status: completed
created: 2026-09-15
updated: 2026-09-15
author: MoAI
priority: P0
phase: "v3.2.0 target"
module: "internal/cli, internal/kanban, internal/graph, internal/runtime, internal/template"
lifecycle: spec-anchored
tags: "gtd, autonomy, knowledge-graph, compatibility, mission-runtime, governance"
tier: L
---

# SPEC-GTD-AUTONOMY-001

## HISTORY

| 버전 | 날짜 | 변경 |
|---|---|---|
| 0.2.0 | 2026-09-15 | plan-audit iteration 1 D1~D6 보정: repo-local develop 통합, RED-now/green adoption, private graph privacy, todo 전체 호환, autonomous E2E, MissionRuntime recovery matrix |
| 0.1.0 | 2026-09-15 | GTD 명칭·관리 절차, 비공개 관계 그래프, 사전 위임, `/moai:goal --auto` 자율 목표 모드의 최초 Tier L 계약 작성 |

## 1. 목적과 범위

이 SPEC은 승인된 개발 작업을 처리하는 기존 todo 큐를 폐기하거나 GTD 단계로 대체하지 않는다. 기존 큐 위에 수집·명료화·정리·검토·선택 계층을 더하고, 사용자가 봉인한 사전 위임 범위 안에서만 LLM이 카드 발행·선택·배차와 개발 통합을 지속 수행할 수 있게 한다.

다음 경계를 고정한다.

- GTD의 Capture → Clarify → Organize → Reflect → Engage는 일을 정리하고 선택하는 관리 절차다.
- 개발 진행은 backlog → plan → run → sync → done을 계속 사용한다.
- 기존 SQLite 큐, 카드 ID, 순서, `queued / picked / dropped`, 보관·복원 의미는 유지한다.
- `moai gtd`는 정식 사용자 표면이고 `moai todo`는 동일 동작의 호환 표면이다.
- `/moai:goal --auto`는 별도 단축 별칭이 없는 단일 신규 플래그다.
- 기존 `progression_mode=autonomous|semi-autonomous`는 호환 저장 값이며 새 `mission_mode=auto`와 별개다.
- 자율성은 안전 게이트의 제거가 아니라 사용자가 미리 승인한 좁은 결정권의 위임이다.

## 2. 요구사항

### 2.1 기존 큐와 명칭 호환

**REQ-GTD-001 — 기존 큐 불변성.** The GTD 기능 shall preserve the canonical SQLite queue location, existing card identity, insertion order, queue states, findings, and archive/restore semantics for existing projects.

**REQ-GTD-002 — 정식·호환 명령.** **When** a user invokes any of `add`, `list`, `done`, `undone`, `next`, `unpick`, `edit`, `move`, `drop`, `undrop`, `analyze`, `relate`, `unrelate`, `why`, `pr`, `landed`, `auto-done`, `export-json`, or `history` through `gtd` or `todo`, including every supported flag combination named in acceptance.md, the CLI shall operate on the same queue and produce equivalent exit status, stdout, card identity, order, relations, findings, landing metadata, history, and persisted state.

**REQ-GTD-003 — 생성물 전환.** **Where** user-facing commands, skills, templates, help, completion, or four-language documentation are emitted, the product shall present GTD as the canonical name and shall retain todo only as a thin compatibility path that cannot diverge in behavior.

**REQ-GTD-004 — 개발 흐름 분리.** The GTD workflow shall not introduce Capture, Clarify, Organize, Reflect, or Engage as development-board columns or replacements for backlog, plan, run, sync, or done.

### 2.2 GTD 다섯 절차

**REQ-GTD-005 — Capture.** **When** an item arrives from a user or an allowed source, the GTD manager shall record its source, sensitivity classification, and deduplication identity without creating an execution commitment.

**REQ-GTD-006 — Clarify.** **When** a captured item is clarified, the GTD manager shall classify its desired outcome and disposition and shall withhold card publication while goal, completion evidence, authority, or source trust is unresolved.

**REQ-GTD-007 — Organize.** **When** a clarified item is organized, the GTD manager shall persist its project or material class, action context, review timing, and proposed or confirmed relationships, and shall reject blocking cycles or missing relation targets.

**REQ-GTD-008 — Reflect.** **When** a queue, evidence, relation, archive, reopen, cancellation, or scheduled-review event changes relevant state, the GTD manager shall reassess stale evidence, blocked successors, projects without next actions, and waiting items without treating cancellation as completion.

**REQ-GTD-009 — Engage.** **When** the GTD manager evaluates actionable candidates, it shall consider current policy, context, evidence freshness, dependency state, lane ownership, and resource limits before producing a recommendation or an authorized queue operation.

### 2.3 데이터, 관계, 최신성

**REQ-GTD-010 — 추가형 저장.** **Where** GTD storage is enabled, the data model shall extend the existing project queue without replacing its current schema-version contract, queue tables, archive tables, or physical database path.

**REQ-GTD-011 — 왕복 보존.** **When** supported old and new binaries, backup and restore, export and import, archive and reopen, or WAL maintenance operate on an extended queue, existing and GTD data shall remain logically recoverable without card re-identification.

**REQ-GTD-012 — 관계 의미.** The relation model shall distinguish blocking `depends_on` from non-blocking `part_of`, `supported_by`, and `related_to`; shall preserve the existing `contains`, `absorbs`, `replaces`, and `conflicts` direction, notes, byte representation, logical meaning, archive, and restore behavior; and shall record provenance, assertion status, source revision, and policy version for new assertions.

**REQ-GTD-013 — 비공개 그래프 최신성.** **When** a GTD graph projection is requested, the projector shall derive a deterministic projection beside the home queue DB, restrict its directory to mode `0700` and projection plus sidecar files to `0600`, admit only the current OS account and explicitly delegated local MoAI process as readers, exclude it by default from repository, template, Git, log, telemetry, export, and backup surfaces, require explicit backup opt-in, apply the same sensitivity to the sidecar, obey documented revoke/delete retention, reject ambiguous privacy classification, and report stale whenever metadata or source revision differs from the current source.

### 2.4 사전 위임과 결정권

**REQ-GTD-014 — 봉인된 위임 계약.** **When** a user approves an autonomous mission, the system shall seal a versioned contract containing goal, completion evidence, scope, allowed actions, merge target, resource limits, prohibited actions, stop and recovery conditions, and revocation behavior.

**REQ-GTD-015 — 제안과 집행 분리.** **When** an LLM proposes a card, selection, dispatch, retry, commit, pull request, or merge operation, a deterministic policy boundary shall reject stale state, missing evidence, scope expansion, unauthorized actions, or invalid idempotency identity before any side effect.

**REQ-GTD-016 — `--auto` 의미 분리.** **Where** `/moai:goal --auto` is selected, the mission shall use a dedicated natural-language mission path and `mission_mode=auto`; it shall not execute mission text, pass it to the existing condition parser, or reinterpret the flag as `progression_mode=autonomous`.

**REQ-GTD-017 — 승인 후 무질문 경계.** **While** an approved auto mission remains within its sealed contract, the orchestrator shall proceed without `AskUserQuestion`; **when** safe progress requires new scope, authority, or unverifiable assumptions, it shall stop further side effects and persist a blocked report instead of broadening approval.

**REQ-GTD-018 — 역할 분리.** The autonomous system shall keep `super-advisor` non-binding, assign bounded decisions to a non-writing `mission-governor`, reserve PASS or FAIL verdicts for independent auditors, and reserve state-changing execution for the existing owning roles and deterministic policy executor.

### 2.5 지속 실행, 복구, 배차와 통합

**REQ-GTD-019 — 런타임 능력 경계.** **When** runtime capabilities are probed as full, partial, or unsupported, the mission runtime shall expose the exact start, reconnect, replace, and credential capabilities; shall use durable mode only when required capabilities are live; shall transition to `active-session-only` on unsupported or unrecoverable reconnect; and shall reconcile credential expiry, replacement, lease loss or takeover, and restart against the same sealed mission snapshot before resuming.

**REQ-GTD-020 — 정확히 한 번의 논리 효과.** **When** a process exits or observation fails around an external side effect, the system shall reconcile a stable operation identity against authoritative queue, dispatch, commit, pull-request, and landed state before deciding to retry or advance.

**REQ-GTD-021 — 안전한 자율 수명주기와 배차.** **When** one captured item proceeds through Clarify, Organize, sealed delegation, card publication, pick, and dispatch, the system shall preserve one item/card/operation identity and authorized order across normal, denied, crash-restart, and concurrent-manager paths; the lead shall re-read current lane availability, ownership, lease, card revision, evidence, and existing dispatch state and shall refuse duplicate, out-of-order, or conflicting assignment.

**REQ-GTD-022 — 저장소 로컬 카드 통합.** **Where** commit and merge actions are authorized in this repository, card work shall start from local `develop` in a launcher-entered `WT-gtd-autonomy`, commit only explicit paths, forbid lane push and card-level pull requests, and merge with `--no-ff` only inside the single local develop integration worktree under its integration lease, recording distinct base, card, and local develop merge SHAs.

**REQ-GTD-023 — develop 흡수와 release 게이트.** **When** autonomous delivery advances after local develop integration, the lead shall batch-push local develop to `origin/develop`, require CI on that exact develop SHA, create a release branch from the verified develop SHA, and use a release pull request to protected `main`; **when** any owner, lease, base/card/local-merge/origin-develop-CI/release-head/main-landed SHA, audit, review, CI, or protection condition is absent or stale, automatic delivery shall stop without lane push, card PR, or gate override.

**REQ-GTD-024 — 비신뢰 입력 격리.** The autonomous system shall not allow external content, mission text, shell metacharacters, tool output, or advisor prose to modify the sealed policy, gain executable interpretation, or grant an action that the deterministic validator did not authorize.

**REQ-GTD-025 — 철회와 종료.** **When** a user revokes a mission or all completion evidence is satisfied, the supervisor shall prevent new work, reconcile in-flight effects, preserve an audit trail, and report either a safe recovery point or verified completion without inferring success from agent idleness, lease expiry, or command exit alone.

## 3. 품질 및 운영 제약

- 저장 마이그레이션은 실제 사용자 큐가 아닌 격리된 `MOAI_HOME`에서 먼저 검증한다.
- 동시성 검증은 반복 이벤트, 복수 관리자, WAL 사용, archive/reopen, 각 부작용 직후의 crash cut을 포함한다.
- 신규 권한 경계는 fail-closed이며 정책·근거·snapshot 버전 불일치는 자동 허용하지 않는다.
- 영향 범위 테스트를 로컬에서 수행하고 전체 회귀는 push 뒤 CI의 현재 SHA에서 확인한다.
- 사용자 표면은 한국어·영어·일본어·중국어와 template source, emitted artifact 간 동등성을 가진다.
- 운영 기록은 임무·카드·결정·작업·SHA·감사 결과를 상호 추적할 수 있어야 한다.

## 4. 추적성

REQ-GTD-001부터 REQ-GTD-025까지는 숫자 suffix가 같은 AC-GTD-001부터 AC-GTD-025까지와 각각 일대일로 대응한다. Milestone 묶음과 실행 순서는 plan.md §F가, 상세 검증 fixture와 판정 기준은 acceptance.md §B가 소유한다.

## 5. 제외 사항

### Out of Scope — 기존 저장소의 일괄 개명

- `internal/kanban`, `BacklogStore`, 기존 SQLite 표·열, `todo/backlog.db` 물리 경로의 일괄 개명은 하지 않는다.

### Out of Scope — 안전 게이트 우회

- force push, 보호 브랜치 우회, 감사 결과 수정, 검증되지 않은 충돌 자동 해결, 금지된 외부 행위의 묵시적 승인은 구현하지 않는다.

### Out of Scope — GTD 단계의 개발 열 전환

- Capture, Clarify, Organize, Reflect, Engage를 backlog, plan, run, sync, done의 대체 열로 만들지 않는다.

### Out of Scope — 별도 원본 그래프와 실증 효과 주장

- 두 번째 카드 큐, 별도 GTD 원본 DB, repo 지식 그래프로의 개인 메모 자동 export, GTD 또는 개인 지식 그래프의 생산성 효과 측정은 포함하지 않는다.
