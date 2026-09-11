---
id: SPEC-RESOURCE-SLOT-LEASE-001
title: "Resource slot lease: atomic acquire/status/release for a named heavy resource, plus an opt-in PreToolUse guard (card t607)"
version: "0.1.0"
status: draft
created: 2026-09-12
updated: 2026-09-12
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/kanban, internal/cli, internal/hook, internal/config, internal/template"
lifecycle: spec-anchored
tags: "slot-lease, resource-lease, heavy-test, mutual-exclusion, kanban, factory, t607"
tier: M
---

# SPEC-RESOURCE-SLOT-LEASE-001 — 자원 슬롯 임대(heavy-test 실행 슬롯)

## §A 배경과 문제

여러 레인 세션이 한 머신에서 동시에 돌 때, 무거운 작업(대형 테스트 스위트, 전체 빌드, 부하가 큰 검증 명령)이 겹치면 머신 전체가 느려지고 타임아웃·불안정한 테스트가 늘어난다. 지금 이 저장소에서 그 겹침을 막는 장치는 **없다**.

- 무거운 테스트를 직렬화하는 **성문 규칙이 없다.** 직렬화 규칙은 병합 창 하나뿐이다(로컬 `gitflow-lane-protocol.md` §3, `kanban-dispatch.md`의 통합 절). 병합 창은 `git merge` 한 가지 행위를 대상으로 설계됐다.
- 레인이 실제로 쓰는 회피책은 **확인 후 시작(probe-then-start)** 이다. 프로세스 목록을 보고 비어 있으면 시작한다. 이것은 상호 배제가 아니다. 확인과 시작 사이에 다른 세션이 시작할 수 있고, 두 세션 모두 "비어 있었다"는 사실을 정당하게 관측한다.

### 직접 관측한 증거와 그 한계 (정직한 서술)

- **직접 관측한 것은 확인-후-시작 경합 한 건뿐이다.** 한 세션이 `ps`로 무거운 테스트가 없음을 확인한 직후, 다른 세션의 `go test`가 시작됐다(판정 기록 `.moai/reports/t607/verdict.md` §1).
- **무거운 패키지끼리의 실제 충돌은 재현하지 못했다.** 사전 판단 때 충돌 대상으로 본 `internal/cli/ptycaptest`는 `internal/cli`를 import하지 않는다. 따라서 이 SPEC은 "측정된 경합 비용"을 주장하지 않는다.
- 이 SPEC이 세우는 근거는 **메커니즘**이다. 확인-후-시작은 구조적으로 두 시작을 모두 허용하고, 원자적 획득은 하나만 허용한다. 그 차이를 대조군 테스트(AC-RSL-001)가 재현한다. 부하 비용의 크기는 이 SPEC의 주장 범위 밖이다.

### 결정된 방향 (운영자 결정, 리드 전달 — 재론하지 않음)

B안: 범용 자원 임대 명령(`moai slot acquire|status|release --resource <name>`)과 **선택형** PreToolUse 가드(기본 꺼짐). 템플릿과 바이너리로 배포하고(로컬 전용 아님), 템플릿 문서는 최소화한다. C안(`moai integration` 확장)은 기각됐다. 기각 사유는 세 가지다. 통합 가드는 `git merge`만 잡는다. 통합 기록 필드는 병합 대상 전용이다. 창이 하나뿐이면 병합 대기와 테스트 대기가 서로를 막는다.

## §B 검증된 전제 (트리 `c4ce42eca` @ `WT-heavy-test-slot`에서 측정)

모든 줄 번호는 이 트리 기준이다. 다른 트리에서 인용할 때는 다시 잰다.

1. 임대 표면은 아직 없다. `SlotLease` / `slot_lease` / `slot-lease` 토큰은 `internal`, `cmd`, `pkg` 어디에도 없고(acceptance.md 증거 원장 EL-1, 대조군 EL-2·EL-3), `slot` cobra 명령도 없다(EL-4).
2. 재사용할 코드 방식 — 통합 창(`internal/kanban/integration_lock.go`):
   - 창은 **기록**이고, 유효성은 기록된 보유자의 생존으로 판정한다(`Stale()`, :152-160). pid ≤ 0은 "판정 불가 = 살아 있음"으로 읽는다.
   - 기록의 읽기→판정→쓰기는 **짧은 수명의 변경 락** 안에서 한 번에 돈다(`AcquireIntegrationLock` :216-262, `withIntegrationLockMutation` `integration_lock_mutation.go`:75-88). 변경 락 기반은 Unix의 flock, Windows의 원자적 생성이다(`acquireBoardLockImpl`). Windows에서 죽은 보유자가 남긴 락 파일은 소유자가 확실히 사라진 경우에만 치운다(`integration_lock_mutation.go`:108-122).
   - 기록 쓰기는 호출마다 고유한 임시 파일에 쓴 뒤 rename 한다(`writeIntegrationLock` :315-360).
   - 소유자 pid는 CLI 프로세스가 아니라 **세션**의 pid다(`session.ResolveOwnerPID`, `internal/session/session_pid.go`:82-90). 풀리지 않으면 0을 기록한다(`internal/cli/integration.go`:247-254).
   - `--force` 인수는 밀어낸 보유자를 보고한다(`internal/cli/integration.go`:282-286). 단, 밀어낸 사실은 **표준 출력에만** 나오고 기록에는 남지 않는다.
3. 기록 위치 해석: `CLAUDE_PROJECT_DIR` → `git rev-parse --git-common-dir`의 부모 → 현재 디렉터리(`internal/cli/integration.go`:46-65). 워크트리에서 호출해도 primary 체크아웃의 `.moai/state`로 모인다. 훅 쪽도 같은 순서로 프로젝트 루트를 푼다(`internal/hook/pre_tool.go`:365-401).
4. 세션 id 해석: `--session` 플래그 → `CLAUDE_CODE_SESSION_ID` → SessionStart 사이드 채널 파일(`internal/cli/integration.go`:80-92). 없으면 값을 지어내지 않고 오류로 보고한다.
5. 통합 가드(`internal/hook/integration_lock_guard.go`)의 성질 — 이 SPEC의 가드가 물려받는다:
   - 기본 꺼짐, 꺼진 경로에서는 기록을 읽지 않는다(`pre_tool.go`:547, :821-830).
   - 거부 사유는 고정 접두어로 시작한다(`INTEGRATION_LOCK_VIOLATION`, :44).
   - 불확실하면 허용한다(fail-open): 읽을 수 없는 기록, 세션 id 없음, 루트 없음, 죽은 보유자(:74-98).
   - 따옴표 구간을 지운 뒤 매칭한다(`substituteQuotedArguments`, :70).
   - 매칭 대상은 고정 정규식 `\bgit\s+merge\b` 하나다(:51). 이것이 C안 기각 사유 중 하나다.
6. 설정 키 방식: `workflow.integration_lock.enabled`는 `internal/config/types.go`:461-466, :669-678에 선언되고, `internal/config/defaults.go`:900-902에서 `Enabled: false`로 고정된다. 배포 템플릿 `workflow.yaml`에는 이 키가 없다. 템플릿에 선택형 게이트를 싣는 선례는 `codex.review_gate` / `multi.review_gate`(`enabled: false` + 설명 주석)다.
7. 템플릿 중립성 가드는 현재 초록이다: `TestTemplateNeutralityAudit`, `TestTemplateNoInternalContentLeak` 두 테스트가 실제로 돌았고 통과했다(증거 원장 EL-6, 스윕 수 2).
8. 문서 편집 보류 조건: 카드 t637 브랜치 `WT-acquire-branch-record`(`f680dab46`)는 아직 `develop`(`eb50af5a8`)의 조상이 아니다(EL-5, 종료 코드 1). 따라서 `kanban-dispatch.md`와 `gitflow-lane-protocol.md` 편집은 이 트리에서 착수하지 않는다(§C REQ-RSL-016, plan.md M6).

## §C 요구사항 (GEARS)

요구사항은 관측 가능한 계약만 고정하고 코드 모양은 정하지 않는다. 번호는 REQ-RSL-NNN. 각 요구사항의 정본 문장은 GEARS 영어 형식이고(lint가 양상을 판정할 수 있도록), 바로 아래의 한국어 항목은 그 뜻을 풀어 쓴 설명이다. 둘이 어긋나면 영어 문장이 정본이다.

- **REQ-RSL-001** (Ubiquitous) — The slot lease surface shall provide acquire, status, and release operations keyed by a resource name, and shall use a record and a mutation lock that are separate from the integration window, leaving the `moai integration` verbs, record, and guard unchanged.
  - 슬롯 임대는 통합 창과 기록·변경 락을 공유하지 않는다. `moai integration`은 그대로다.

- **REQ-RSL-002** (Event-driven) — **When** an acquire is requested for a resource that has no live, unexpired holder, the slot lease surface shall record the holder with the resource name, holder session id, session name, owner pid, command, acquisition time, declared maximum duration, and the expiry time derived from it.
  - 획득이 성공하면 위 필드가 모두 디스크 기록에 남는다.

- **REQ-RSL-003** (Ubiquitous / 원자성) — The slot lease surface shall admit exactly one of any set of concurrent acquires for the same resource, by performing the record read, the decision, and the write inside one critical section serialized across processes, so that no acquire can interleave between another acquire's check and its write.
  - 확인-후-행동이 아니라 원자적 획득이다. 이것이 이 SPEC이 닫는 결함이다.

- **REQ-RSL-004** (State-driven) — **While** a different session is recorded as a live, unexpired holder, the slot lease surface shall refuse an acquire without `--force` with a held error that names the holder, and shall leave the record unchanged.
  - 거절은 기록을 한 바이트도 바꾸지 않는다.

- **REQ-RSL-005** (Event-driven / 스테일) — **When** the process named by the recorded owner pid no longer exists, the slot lease surface shall let a different session's acquire take the resource over without `--force`, and shall report and record the displaced holder with reason `stale`.
  - 보유 세션이 죽었으면 강제 없이 넘겨받되, 밀어낸 사실을 남긴다.

- **REQ-RSL-006** (Event-driven / 선언 상한) — **When** the holder's declared maximum duration has elapsed, the slot lease surface shall read the lease as expired even though the owner process is alive, shall report the expiry on status, shall let a different session's acquire take it over without `--force` while reporting and recording the displaced holder with reason `expired`, and shall restart the bound when the holder itself re-acquires.
  - 상한은 보유자 자신의 약속이다. 넘기면 살아 있어도 인수 가능하다(근거는 plan.md §B3).

- **REQ-RSL-007** (Event-driven / 강제 인수) — **When** an acquire with `--force` takes over a live, unexpired holder, the slot lease surface shall record the displaced holder with reason `force` both in the new record and in the audit log.
  - 강제 인수는 표준 출력만이 아니라 기록과 감사 로그에 남는다.

- **REQ-RSL-008** (Event-driven / 보수적 판정) — **When** the owner pid cannot be resolved at acquire time, the slot lease surface shall record pid 0 and shall read that lease as live until an explicit release, a recorded `--force` takeover, or expiry of the declared bound, and shall not record the acquiring process's own pid in its place.
  - 판정할 수 없으면 "살아 있음"으로 기운다. CLI 프로세스 pid를 기록하면 즉시 스테일로 읽히는 결함이 재발한다.

- **REQ-RSL-009** (Event-driven / 해제) — **When** a release is requested, the slot lease surface shall accept it only from the recorded holder session, shall refuse a different session's release without `--force`, and shall report a release of an unheld resource as an error rather than a silent success.
  - 남의 임대를 풀거나, 없는 임대를 푸는 것은 오류다.

- **REQ-RSL-010** (Unwanted / 입력 검증) — The slot lease surface shall not accept a resource name outside the permitted character set, including names containing path separators, `..`, an empty name, or a name longer than the permitted length, and shall write no file anywhere for such a request.
  - 자원 이름은 파일 경로의 일부가 되는 신뢰 경계다.

- **REQ-RSL-011** (Capability gate / 가드 거부) — **Where** `workflow.slot_lease.enabled` is true, **When** a Bash command matches a configured command pattern of resource R and R's holder is live, unexpired, and a session other than the caller, the PreToolUse guard shall deny the command with a reason that starts with the fixed prefix `SLOT_LEASE_VIOLATION:` and names the resource, the holder, and how to wait, release, or take over.
  - 거부는 "다른 세션이 살아 있고 만료 전 보유 중"이라는 긍정적 증거가 있을 때만 일어난다.

- **REQ-RSL-012** (Capability gate / 기본 꺼짐) — **Where** `workflow.slot_lease.enabled` is unset or false, the PreToolUse guard shall read no lease record and shall deny no command, while the acquire, status, and release operations shall keep working regardless of the value.
  - 기본 꺼짐. 꺼진 경로는 비용이 없다.

- **REQ-RSL-013** (Ubiquitous / 가드 허용 경로) — The PreToolUse guard shall allow a command when the holder is the calling session, when the command matches no configured pattern outside quoted spans, when the resource has no holder, or when the holder is stale or expired, and shall match only against per-resource patterns taken from configuration rather than a built-in list of commands of any programming language.
  - 패턴은 사용자가 자기 생태계의 명령으로 적는다. 내장 목록은 템플릿 중립성과 충돌한다.

- **REQ-RSL-014** (Event-driven / fail-open) — **When** the PreToolUse guard meets uncertainty, namely an unreadable record, a missing caller session id, a missing project root, a pattern that does not compile, or unavailable configuration, the guard shall allow the command, write an advisory to standard error, and append one line to the audit log where a project root is available, and shall never deny on uncertainty.
  - 손상된 JSON 하나로 모든 무거운 명령을 막지 않는다.

- **REQ-RSL-015** (Ubiquitous / 배포) — The distributed template shall carry the `workflow.slot_lease` key with `enabled: false` and with resource patterns either empty or shown only as a placeholder that names no programming language, and every template file this SPEC adds shall remain neutral across the supported programming languages and shall contain no SPEC ID, date, commit SHA, or internal card id.
  - 템플릿에는 꺼진 키와 언어 중립 자리표시자만 싣는다.

- **REQ-RSL-016** (Event-driven / 문서 게이트) — **When** the entry condition `git merge-base --is-ancestor WT-acquire-branch-record develop` exits 0, the lane documents shall describe the slot lease in the local and template `kanban-dispatch.md` and in the local `gitflow-lane-protocol.md`; until that condition holds, no commit of this SPEC shall change those three files.
  - 카드 t637이 develop에 병합되기 전에는 세 문서를 건드리지 않는다.

## §D 제약

- **통합 창과 분리 (결정 사항).** 기록 파일, 변경 락 파일, 설정 키, 거부 접두어, CLI 명령 모두 통합 창과 다른 이름을 쓴다. `moai integration`에 동작이나 필드를 추가하지 않는다.
- **판정 방향.** 소유자 생존 판정의 불확실성은 "살아 있음" 쪽으로 기운다(통합 창과 같은 비대칭). 단, 선언 상한은 보유자 자신의 선언이므로, 상한이 지나면 살아 있는 소유자라도 인수 가능하다(REQ-RSL-006). 이 선택의 근거와 비용은 plan.md §B에 적는다.
- **가드의 판정 방향.** 가드의 불확실성은 "허용" 쪽으로 기운다(branch guard, integration lock guard와 같은 fail-open 원칙).
- **입력 검증은 단순화 대상이 아니다.** 자원 이름은 파일 경로의 일부가 되므로 신뢰 경계다(REQ-RSL-010).
- **템플릿 우선 규칙.** 새 템플릿 파일은 `internal/template/templates/` 아래에 먼저 만들고 `make build`로 바이너리에 다시 임베드한다. 로컬 사본은 그 뒤에 맞춘다.
- **템플릿 중립성.** 템플릿 내용은 16개 프로그래밍 언어에 중립이다. 특정 언어의 테스트 명령을 "무거운 명령"의 대표로 싣지 않는다. SPEC ID, 날짜, 커밋 SHA, 카드 id를 넣지 않는다(`template-neutrality-check.yaml`, `internal_content_leak_test.go`).
- **테스트 격리.** 모든 테스트는 `t.TempDir()` 아래에서 돌고 `CLAUDE_PROJECT_DIR` + `GIT_CEILING_DIRECTORIES`를 고정한다. 실제 primary 체크아웃의 `.moai/state`나 `.moai/logs`를 건드리지 않는다.
- **로컬 검증 범위.** 영향받는 패키지 테스트만 돌린다. `go test ./...` 전체 스위트는 로컬에서 돌리지 않는다.
- **플랫폼.** Unix와 Windows 모두 빌드돼야 한다. Windows에서 죽은 보유자가 남긴 변경 락 파일이 이후 획득을 영구히 막아서는 안 된다.

## §E 인수 기준 요약

정본 목록은 `acceptance.md`(AC-RSL-001..016)에 있다. 핵심은 세 가지다.

- **AC-RSL-001 대조군** — 같은 테스트 안에 두 갈래가 있다. 표면 없이 확인-후-시작을 하면 두 세션이 모두 시작한다. 표면으로 획득하면 한 세션이 거절된다. 확인과 행동 사이의 끼어들기는 기다려서 우연히 얻지 않고 **구성**한다.
- **AC-RSL-002 뮤턴트** — 임계 구역 직렬화를 한 줄로 끄면 AC-RSL-001의 표면 갈래가 이중 보유로 실패해야 한다. 통과의 원인이 락임을 보인다.
- **AC-RSL-011 가드 복합 조건 표** — 거부 조건(설정 켜짐 ∧ 패턴 일치 ∧ 다른 세션 ∧ 살아 있음 ∧ 만료 전)의 각 항마다 그 항만 거짓인 입력을 두고, 그 입력이 허용되는지 확인한다.

## §F 의존성과 관련 SPEC

- 코드 방식 출처: SPEC-INTEGRATION-LOCK-LIVENESS-001(세션 소유자 pid 앵커), SPEC-INTEGRATION-LOCK-ATOMIC-001(변경 락 직렬화). 둘 다 `completed`이며 이 SPEC은 그 **코드 방식만** 빌려 온다. 두 SPEC이 만든 통합 창에는 손대지 않는다.
- 문서 편집 게이트: 카드 t637(`WT-acquire-branch-record`)의 develop 병합. SPEC 의존성(`depends_on`)이 아니라 plan.md M6의 진입 조건으로만 다룬다. M1-M5는 이 조건과 무관하게 진행한다.
- `depends_on` 항목 없음.

## §G 위험

- **서브에이전트의 세션 id 귀속이 측정되지 않았다.** 보유 세션이 띄운 서브에이전트의 Bash 호출에서 훅 입력의 `session_id`가 부모 세션 id와 같은지 이 트리에서 관측하지 않았다. 다르다면 보유자 자신의 서브에이전트가 "다른 세션"으로 읽혀 거부된다. 이 경우 거부는 기록된 보유자가 살아 있다는 긍정적 증거 위에서만 일어나므로 조용한 오작동은 아니지만, 보유자에게 불필요한 마찰이 된다. plan.md §E OQ-1에서 run 단계 M4가 실제 훅 입력으로 잰다.
- **확인하지 않고 건너뛰는 두 세션은 여전히 경합한다.** 가드는 "다른 세션이 이미 보유 중"일 때만 거부한다. 두 세션이 모두 획득 없이 무거운 명령을 시작하면 가드는 둘 다 허용한다(보유자가 없으므로). 이 SPEC이 닫는 것은 획득하는 세션 사이의 경합과 "보유 중인데 건너뛴" 경로다. 보유 없이 매칭 명령을 거부하는 강제 모드는 범위 밖이다(Exclusions).
- **잊힌 해제.** 소유자 pid는 세션의 pid이므로, 해제를 잊은 살아 있는 세션은 선언 상한까지 자원을 붙잡는다. 선언 상한이 그 출구다(REQ-RSL-006). 상한 기본값이 너무 길면 대기가 늘고, 너무 짧으면 느린 정상 실행이 넘겨질 수 있다.
- **pid 재사용.** 세션이 죽은 뒤 운영체제가 그 pid를 다른 프로세스에 주면 스테일 판정이 "살아 있음"으로 잘못 기운다. 통합 창에서 물려받은 성질이며, 이 SPEC에서는 선언 상한이 이 경우의 출구도 된다.
- **단일 호스트 전제.** 기록에는 호스트 필드가 없다. 다른 호스트에서 잰 pid는 의미가 없다. 통합 창과 같은 전제다.
- **Windows 변경 락 잔재.** Windows 기반은 파일 자체가 락이므로 임계 구역 안에서 죽은 프로세스가 파일을 남긴다. 통합 창의 잔재 정리 방식을 자원별 락 경로에 적용해야 하며, 그 동작을 로컬에서 행동으로 관측할 수단은 없다(Windows 로컬 증거는 컴파일 검사까지다).
- **템플릿 규칙 인벤토리 테스트.** 새 템플릿 규칙 파일이나 설정 키가 기존 인벤토리·예산 테스트(규칙 목록, 설정 키 선언 대응)에 걸릴 수 있다. plan.md 사전 점검에서 확인한다.

## §H History

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-09-12 | 최초 초안(plan 단계, 카드 t607). 운영자 B안 결정을 전제로 작성. 네 개 plan 단계 산출물 모두 `draft`로 생성. |

## Exclusions

### Out of Scope — 통합 창의 변경
- `moai integration` 명령, 통합 기록 필드, `workflow.integration_lock` 키, `INTEGRATION_LOCK_VIOLATION` 가드 동작의 변경. 슬롯 임대를 통합 창의 하위 기능으로 합치는 것(C안)은 기각됐다.

### Out of Scope — 보유 강제 모드
- 보유자가 없을 때 매칭 명령을 거부하는 모드(모든 무거운 명령에 획득을 요구). 단일 세션 작업을 막는 의무 절차가 되므로 이 SPEC에서 만들지 않는다. 필요성이 측정되면 후속 카드로 다룬다.

### Out of Scope — 명령 래퍼
- 획득→명령 실행→해제를 한 번에 하는 `moai slot run -- <cmd>` 류의 래퍼. 잊힌 해제를 줄이는 데 유용하지만 운영자가 정한 표면(획득·상태·해제)을 넘는다.

### Out of Scope — 부하 비용 측정
- 무거운 테스트 겹침이 만드는 부하·지연·불안정성의 정량 측정. 이 SPEC은 메커니즘(원자적 획득 대 확인-후-시작)만 재현한다.

### Out of Scope — 다중 호스트와 대기열
- 여러 머신 사이의 조율, 대기열(순번 예약), 공정성 보장. 획득은 거절 아니면 성공이며, 기다리는 방법은 호출자가 정한다.

### Out of Scope — 게이트 전 문서 편집
- 진입 조건(REQ-RSL-016)이 참이 되기 전의 `kanban-dispatch.md`(로컬·템플릿)와 `gitflow-lane-protocol.md` 편집. 게이트가 닫힌 채 run 단계가 끝나면 후속 카드로 넘긴다.

### Out of Scope — 가드 기본값 켜기
- 어떤 설정에서든 `workflow.slot_lease.enabled`를 참으로 배포하는 것. 이 저장소의 로컬 설정에서 켤지는 운영자의 별도 결정이다.
