---
id: SPEC-FACTORY-MANAGED-CARD-CHILD-001
title: "Codex 레인 카드 자식 세션을 관리 경로에 연결 — `moai codex -f lane` 의 카드별 자식이 관리 Codex 소유자로 뜬다 (SPEC-FACTORY-MANAGED-SESSION-001 후속)"
version: "0.2.0"
status: draft
created: 2026-10-03
updated: 2026-10-03
author: GOOS (manager-spec)
priority: P1
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tier: M
depends_on: [SPEC-FACTORY-MANAGED-SESSION-001, SPEC-FACTORY-MANAGED-HARDEN-001]
related_specs: [SPEC-FACTORY-SELF-DISPATCH-001, SPEC-CODEX-LANE-SLOTS-001]
tags: "factory, managed-session, codex, lane-loop, card-child, opt-in"
---

# SPEC-FACTORY-MANAGED-CARD-CHILD-001 — Codex 레인 카드 자식의 관리 경로 연결

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.2.0 | 2026-10-03 | manager-spec | plan-audit 1차(FAIL 0.81, `audited_sha d26215f25`) 개정. D1/MP-7: 열린 질문 Q1-Q3 를 리더 결정으로 기록(기본값 채택, `decision-index.md`). D2/MP-6: C.5 에 `cross-platform exemption` 문장. D5: REQ-CC-010 을 보장 가능한 범위(끝낸 줄 이후 입력)로 재서술하고 어댑터를 한 줄 단위·종료 토큰 인식으로 재설계. D6: 입력 어댑터를 레인 루프 전용 이음새로 한정하고 REQ-CC-003 에 플레인 경로 stdin 불변을 명시. D9: REQ-CC-007 을 007/013/014 로 분리(REQ 14). D10: REQ-CC-004/005 의 구현 용어 축소. D11: §F 문구. D12: Known limits(무인 레인 한계)·결정 기록 추가. REQ 12→14. |
| 0.1.0 | 2026-10-03 | manager-spec | Initial SPEC (card t1440, lane-17). 완료된 SPEC-FACTORY-MANAGED-SESSION-001 의 Known debt 4번("Codex lane card children are not managed")이 가리킨 후속 작업을 닫는다. 부모와 HARDEN SPEC 은 본문·frontmatter 모두 수정하지 않는다. |

## §A. 개요

부모 SPEC-FACTORY-MANAGED-SESSION-001(상태 `completed`, 카드 t1375)은 관리 세션 계층을 **명시적 옵트인**(`MOAI_FACTORY_MANAGED` 가 `1` 또는 `true`, 기본 꺼짐) 뒤에 두었다. 그 계층의 Codex 쪽 소유자는 런처가 세션 프로세스의 부모로 남아 헤드리스 App Server 를 직접 소유하고, 브로커 inbox 를 전달하고, 서버가 먼저 보내는 요청에 답하고(HARDEN-001), 턴 단위 실패를 격리한다.

그런데 그 소유자에 닿는 문은 하나뿐이다. 평범한 `moai codex` 가 프로세스 환경에 옵트인과 factory 스탬프를 이미 갖고 있을 때만 `runCodexLaunch` 의 분기가 소유자로 보낸다(`internal/cli/codex_launcher.go:1182`). 반면 `moai codex -f lane` 은 `runCodexFactoryLane` 감독 루프로 가서(`codex_launcher.go:734`) 카드마다 `launchCodexCardSession` 을 거쳐 직접 exec 문 `codexDirectLaunchFn` 으로 자식을 띄운다(`codex_launcher.go:1000`, `:1038`). 코드 주석이 이를 명시한다 — "`moai codex -f lane` card children never reach this code"(`codex_launcher.go:1180-1181`). 운영 문서 `.moai/docs/factory-managed-session.md` 는 같은 공백을 "켜도 닿지 않는 범위"에 적어 두었고(26-27행, 30-31행), 부모 SPEC 의 Known debt 4번과 시험 `TestManagedSwitchDoesNotReachCodexLaneLoop`(`managed_optin_test.go:190`)이 그 공백을 고정한다.

이 SPEC 은 그 공백을 **옵트인을 켠 경우에 한해** 닫는다: 스위치와 레인 스탬프가 함께 있으면 레인 루프가 카드마다 자식 세션을 관리 Codex 소유자로 띄운다. 스위치가 없으면 카드 자식은 지금과 바이트 단위로 같게 직접 exec 문으로 간다.

### 이 SPEC 이 정하는 판단 (설계 근거는 design.md)

| 판단 | 결정 | 근거 |
|---|---|---|
| 카드 자식의 cwd | 카드 워크트리를 소유자의 런치 디렉터리로 넘긴다. App Server 프로세스와 스레드 cwd 가 그 디렉터리다. `-C` 인수는 넘기지 않는다 | 소유자는 `-c`/`--config` 와 `-m`/`--model` 외의 인수를 거부한다(`managed_codex_factory.go:573-600`, 거부 줄 `:595-597`). 디렉터리는 `session.dir` 로 `cmd.Dir` 에 실린다(`:805`, `:688`) |
| 앵커 락·레인 claim 스탬프 | 둘 다 하지 않는다. 레인 claim 은 런처 PID 에 그대로 있고, 그 PID 가 브로커 endpoint 소유자다 | 카드 자식의 현행 경로가 둘 다 하지 않는다(`launchCodexCardSession` `:1015-1039` 에 락·스탬프 호출 없음; 락 호출 3곳은 전부 `runCodexLaunch` 안 `:1186`, `:1232`, `:1254`) |
| 레인 라벨 환경 | 현행 카드 자식 환경(`codexCardLaunchEnv`) 그대로, 소유자가 요구하는 run id 한 키만 더한다 | 소유자가 run id 를 환경에서 읽는다(`managed_codex_factory.go:789-791`). 현행 카드 환경은 그 키를 지운다(`codexLaneLaunchEnvKeys` `:325-338`) |
| 관리 세션이 끝나면 | 루프는 지금처럼 다음 카드로 이어진다. 소유자가 돌려준 오류는 stderr 한 줄로 남기고 루프를 끝내지 않는다 | 현행 루프의 처리와 같다(`codex_launcher.go:1000-1005`) |
| 운영자 입력(stdin) | 레인 루프 전용 입력 펌프(한 줄 단위, `/exit`·`/quit` 종료 토큰 인식)로, 세션을 끝낸 줄 **이후** 입력이 끝난 세션에 먹히지 않게 한다. 플레인 경로의 stdin 은 건드리지 않는다 | 소유자의 stdin 읽기 고루틴은 세션이 끝나도 풀리지 않고(`managed_factory_session.go:290-296`, `:314`) 스캐너·버퍼 채널이 줄을 미리 소비한다. 종료 토큰 이후를 어댑터가 막지 않으면 보장이 불가능하다(design.md D-6) |
| 카드 시작·무인 레인 | 카드 시작 프롬프트를 주입하지 않는다. 관리 카드 자식은 우선 턴 뒤 유휴라 무인 레인은 첫 카드 이후로 진행하지 않는다 | 리더 결정 Q1(decision-index.md). 알려진 한계로 §F·문서에 적는다 |

## §B. 범위

1. **런치 분기** — `runCodexFactoryLane` 이 카드 자식을 띄울 때, 옵트인과 레인 스탬프가 함께 있으면 관리 Codex 소유자로, 아니면 현행 직접 exec 문으로 보낸다.
2. **런치 형태** — 관리 소유자에 넘기는 디렉터리·인수·환경을 정한다(위 표).
3. **브로커 바인딩의 연속 카드 처리** — 같은 런처 프로세스가 카드마다 같은 레인 라벨로 소유자를 연속해서 띄울 때 endpoint 등록·바인딩·롤백이 성립한다.
4. **운영자 입력 공유** — 연속하는 관리 세션들이 한 stdin 을 나눠 쓸 때 입력이 끝난 세션에 먹히지 않는다.
5. **문서** — `.moai/docs/factory-managed-session.md` 의 해당 범위·한계 문장을 고친다(sync 단계).

기존 관리 계층의 배달·서버 요청 응답·턴 단위 실패 격리는 **재사용**한다. 새 구현을 만들지 않는다.

## §C. 요구사항 (GEARS)

### C.1 분기

- **REQ-CC-001** — **Where** the explicit managed-session opt-in is active (`MOAI_FACTORY_MANAGED` set to `1` or `true`, case-insensitive and trimmed) **and** the lane loop's process environment carries the factory stamps (`MOAI_KANBAN_ID` together with `MOAI_FACTORY_WORKER` or `MOAI_FACTORY_WORKERS`), **when** the Codex lane loop launches the child session for a leased card, the loop shall start that session through the managed Codex owner exactly once per leased card and shall not start it through the direct launch door.
- **REQ-CC-002** — **While** the managed opt-in is inactive (unset, empty, or any value other than `1` or `true`), the Codex lane loop shall launch each card child through the direct launch door with the same program, argument vector, working directory, environment, and stdio as before this SPEC, and shall not consult the managed owner.
- **REQ-CC-003** — The change shall leave unchanged the plain `moai codex` managed divert (including the standard input it hands the owner, so its end-of-input behavior is the same as before), the tmux `--spawn` door (which never diverts), and the lane loop's lease, worktree-ensure, and stop logic.

### C.2 런치 형태

- **REQ-CC-004** — The managed card-child launch shall give the managed owner the card worktree as its launch directory (so the App Server process and the session thread both start there), and shall pass the worktree's local developer-instruction override — produced by the same producer and passing the same size check as the direct door — in the argument vector without the working-directory flag the owner refuses.
- **REQ-CC-005** — The managed card-child launch environment shall carry the card child's identity stamps (the lane role marker, the lane label under both of its carriers, the Codex backend value, and the leased card's id) together with the run id the owner requires, and shall not carry the factory lane-count stamp.
- **REQ-CC-006** — The managed card-child launch shall not place a worktree anchor lock and shall not rewrite the lane claim to another process id; the launcher process shall remain the lane claim holder and the owner process of the broker endpoint.

### C.3 브로커 바인딩과 재사용

- **REQ-CC-007** — **When** a managed card-child session starts, the owner shall register the lane label's launch-pending endpoint under the launcher's process identity and bind it to the new session thread.
- **REQ-CC-013** (번호는 마지막이지만 의미상 REQ-CC-007 의 연속 카드 변형이라 이 절에 둔다) — **When** an earlier card's session of the same launcher process left a bound endpoint under the same lane label, the owner shall complete the registration of the next card's session and replace that endpoint.
- **REQ-CC-014** — **When** a managed card-child session fails to start, no launch-pending row shall remain for its lane label.
- **REQ-CC-008** — **While** a managed card-child session is up, the session shall deliver broker inbox messages, answer App Server requests, and isolate turn failures through the same owner code the plain managed launch uses (REQ-MS-003, REQ-MS-004, REQ-MH-001 through REQ-MH-009), and the change shall add no second implementation of those behaviors.

### C.4 루프 연속과 운영자 입력

- **REQ-CC-009** — **When** a managed card-child session ends, normally or with an error, the lane loop shall continue to the next lease attempt exactly as it does after a direct-door child; an error shall be written to stderr as one line and shall not stop the loop.
- **REQ-CC-010** — **While** the lane loop runs successive managed card-child sessions in one process, the loop shall hand each session operator input one line at a time, and **when** a session receives the line that ends it (`/exit` or `/quit`), the loop shall deliver no later input to that session and shall hold that input for the next session. Input a session already received before it ended by any other cause (a fatal error, for example) belongs to that session and is not recovered.

### C.5 플랫폼과 문서

- **REQ-CC-011** — Cross-platform exemption: the change shall add no operating-system-specific call (the system-call package is not referenced) in newly added files, and the `GOOS=windows GOARCH=amd64` cross build shall pass.
- **REQ-CC-012** — The managed-session operator guide shall state the card-child reach (opt-in on: the lane loop's card children run under the managed owner; default off: direct exec as before), shall keep the inherited limits (headless output, no signal handling) and name the follow-up cards t1408 and t1459 for them, shall state the unattended-lane limit (the managed card child is idle after its priming turn, so an unattended lane does not advance past its first card), and shall no longer state that card children are never managed.

## §D. 제약

- **기본 꺼짐 불변**: 스위치가 꺼져 있으면 모든 경로가 이 SPEC 이전과 바이트 단위로 같다. 이 SPEC 은 스위치 해석(`factoryManagedRequested`, `factoryLaunchEnabled`; `factory_launch_pending.go:29-41`)을 바꾸지 않는다.
- **소유자 불변**: 관리 소유자·드라이버·브로커 저장소의 동작은 바꾸지 않는다(`managed_codex_factory.go`, `managed_factory_session.go`, `internal/factorymsg/store.go` 무수정이 원칙). stdin 공유는 소유자 바깥(레인 루프 전용 진입 이음새)에서 푼다. 플레인 관리 경로의 이음새와 그 `os.Stdin` 은 바꾸지 않는다. 소유자가 부족하면 run 단계는 blocker 를 돌려준다.
- **어휘**: `lane-<n>` 정규형만. 레거시 토큰을 되살리지 않는다.
- **문서 언어**: 문서·안내는 한국어, 코드 주석·커밋은 영어.
- **진행 조건**: plan-audit 통과 뒤 run 에 들어간다. 카드 절차는 레인 내 연속(plan → run → sync)이다.

## §E. 성공 기준 요약

`acceptance.md` 의 AC-CC-001..013 전부 통과. 특히 (1) 스위치 켬 + 스탬프에서 두 장의 카드가 두 번 소유자로 가고 직접 exec 문은 0번 불린다, (2) 스위치 끔(미설정·`0`·`yes` 등)에서 직접 exec 문이 카드마다 한 번, `*exec.Cmd` 전체가 현행 형태로 불리고 소유자와 입력 펌프는 시작되지 않는다, (3) 같은 런처 PID 로 두 번째 카드의 소유자가 실제 브로커 저장소와 가짜 App Server 로 등록·바인딩에 성공한다, (4) 세션을 끝낸 줄 이후 입력이 끝난 세션에 먹히지 않는다(여러 줄 덩어리 포함), (5) 플레인 관리 경로가 `os.Stdin` 을 그대로 쓴다, (6) 문서의 옛 한계 문장이 사라지고 무인 레인 한계가 적힌다.

## 알려진 한계 — Known limits (리더 결정 Q1·Q2·Q3 의 귀결 — decision-index.md)

- **무인 레인 한계.** 관리 카드 자식은 우선 턴(`managedPrimingPrompt`) 뒤 유휴이고 레인 루프는 카드 시작 프롬프트를 주입하지 않는다. 그래서 운영자가 입력하거나 브로커 메시지가 오기 전에는 스스로 카드 작업을 시작하지 않으며, **무인 레인은 첫 카드 이후로 진행하지 않는다.** 이 SPEC 의 이득(루프가 소유자 아래에서 카드를 이어 감)은 운영자가 붙어 있는 레인에서만 닿는다. 후속 카드 후보: 카드 시작 프롬프트 주입(progress.md).
- **입력 종료 없음(Q2).** 드라이버 끝 조건은 그대로다(`/exit`, `/quit`, 치명 오류). stdin 이 EOF 인 레인은 이 셋 없이는 세션이 끝나지 않는다.
- **연속 시작 실패 상한 없음(Q3).** 현행 직접 문과 같이 루프에 연속 실패 상한이 없다. 관리 시작이 계속 실패하면 큐의 카드가 연속으로 lease 되어 큐가 소진될 수 있다(카드는 lease 만료까지 묶인다).

## §F. Exclusions

### Out of Scope — tmux `--spawn` door

- The tmux `--spawn` launch never diverts to the managed owner, with or without the opt-in, and this SPEC leaves that door as it is (`managed_launcher_wiring_test.go:385` `TestManagedCodexLaunchKeepsSpawnDoor` pins it).

### Out of Scope — TUI attach (card t1408)

- A managed Codex card child owns a headless App Server and shows no model output on the operator's terminal. Attaching a TUI to the owned server is card t1408's (branch `WT-managed-codex-tui-attach`); this SPEC does not add it and records the headless limit as inherited.

### Out of Scope — Start/Close and signal lifecycle (card t1459)

- Signal handling, the `Start`/`Close` lifecycle, and early-failure cleanup of the owner are card t1459's (held) and t1410's; this SPEC does not change them. (It only observes, in AC-CC-007, that the owner's existing launch-pending rollback leaves no row when a card-child start fails.)

### Out of Scope — card start and completion judgment

- The lane loop injects no card-start prompt into the managed session (leader decision Q1, decision-index.md) and judges no card completion. Consequence stated plainly: after its priming turn the managed card child is idle, so **an unattended lane does not advance past its first card**; a follow-up card owns the prompt-injection gap. Receipts remain delivery evidence only (parent REQ-MS-014, REQ-MS-015); card disposition stays the model's own `factory` verb work and the controller's.

### Out of Scope — owner and driver behavior

- No change to the managed owner, the delivery driver (end conditions `/exit`, `/quit`, fatal error), the broker store, the App Server protocol handling, or the opt-in predicates.

### Out of Scope — other backends and the loop's failure policy

- Claude and GLM lanes (`moai cc -f`, `moai glm -f`) are untouched. The lane loop's behavior when sessions fail repeatedly (no consecutive-failure ceiling exists today) is not changed here.

## §G. 교차 참조

- `plan.md` — 마일스톤 M0..M4, 사전 측정, 기록된 결정.
- `acceptance.md` — AC-CC-001..013, RED-now 원장, 에지 케이스, DoD.
- `design.md` — D-1..D-6 설계 결정과 기각한 대안.
- `decision-index.md` — 리더가 결정한 Q1·Q2·Q3 기록.
- `.moai/specs/SPEC-FACTORY-MANAGED-SESSION-001/spec.md` — 부모(Known debt 4번이 이 SPEC 의 출발점).
- `.moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/spec.md` — 서버 요청 응답과 턴 격리(재사용 대상).
- `.moai/docs/factory-managed-session.md` — 운영 안내(sync 단계에서 갱신).
