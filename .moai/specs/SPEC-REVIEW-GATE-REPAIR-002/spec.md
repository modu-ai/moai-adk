---
id: SPEC-REVIEW-GATE-REPAIR-002
title: "Codex 리뷰 게이트 수리 2 — 표식 수명과 스위치 해석, 래퍼 상태 명명을 마친다"
version: "0.1.0"
status: draft
created: 2026-10-21
updated: 2026-10-21
author: manager-spec
priority: P0
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tier: M
tags: "codex, review-gate, marker, run-switch, wrapper, card-t1674"
depends_on: [SPEC-REVIEW-GATE-REPAIR-001]
---

# SPEC-REVIEW-GATE-REPAIR-002 — 표식 수명과 스위치 해석, 래퍼 상태 명명을 마친다

카드: **t1674** (SPEC-REVIEW-GATE-REPAIR-001 v0.2.7·v0.2.8 분할 승계 — M5·M6 몫) · 기준 코드 트리 `8249bff8c3` (이 워크트리의 HEAD, 로컬 main 꼭짓점)

## HISTORY

- 2026-10-21 · v0.1.0 · manager-spec · 최초 작성(draft). SPEC-REVIEW-GATE-REPAIR-001 v0.2.7 이 t1674 로 옮긴 M5·M6 와 v0.2.8 이 옮긴 구 REQ-RG-011, 그리고 v0.2.10·v0.2.11 이 t1674 wrapper scope 로 배정한 래퍼 종료 경로를 이어받는다. Tier M 판정 근거는 plan.md §A. 카드가 인용한 t1620 감사 보고서 두 개는 추적되지 않는 파일이라 인용하지 않았고, 그 내용은 카드 본문 요약과 SPEC-001 HISTORY·decision-index 로 대체했다(plan.md §A 증거 원천 Gap). 상속 항목 여섯과 B1·B2 의 처분은 아래 「상속 항목 처분」 절에, 리더 판정이 필요한 열린 질문은 plan.md 「열린 질문」에 둔다. 이 행은 감사 결과를 주장하지 않는다.

## §0 지배 원칙 [HARD]

> **턴 종료 게이트만 차단하고, 프롬프트 진입은 차단하지 않는다. 검토 없이 턴 종료가 허용될 때마다 화면에 드러난다. 표식은 동시 실행만 막고, 판정이 기록된 커밋의 재검토 여부는 결과 캐시가 정한다.**

앞 문장은 SPEC-REVIEW-GATE-REPAIR-001 §0 을 이어받은 것이고(구 REQ-RG-005~008·011 의 의미를 이 SPEC 이 실행한다), 뒤 문장은 SPEC-001 v0.2.3 (2)·design.md §3 이 t1674 에 남겨둔 표식 규칙이다. 검토가 어떤 판정으로 끝나든 그 순간에 표식은 풀리고, 소유자가 죽은 표식은 회수된다. 스위치는 커밋된 값이 기본이며, 재정의 파일은 그 run 의 member lane 에서만 유효하다. 래퍼는 해석하지 않고 처리기를 부른다.

---

## §A 배경 (측정)

### A.1 증거 원천 — 카드가 인용한 감사 보고서는 읽을 수 없다 (Gap)

카드 t1674 는 `.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter10.md` 와 `plan-audit-split-iter1.md` 를 근거로 인용한다. 두 파일은 main 에 추적되지 않고 t1620 워크트리는 더 이상 없다(이 세션 관측: `git ls-tree main .moai/reports/t1620/` 출력 없음, 디렉터리 부재). 이 SPEC 은 그 경로를 읽을 수 있는 증거로 인용하지 않는다. 두 보고서의 내용은 카드 본문 요약, SPEC-001 HISTORY v0.2.6–v0.2.11 행, SPEC-001 decision-index.md 로 대체했다. iter10 B1 과 B2, iter3 부채(N1·N2·N3·F3·F6·F7), R-1, F5, F2, tree_scope-skip 명명 항목은 그 대체 원천에서 나온다.

### A.2 현재 트리의 관측 (이 세션, HEAD `8249bff8c3`)

- **표식은 버킷 이름의 빈 내용물이다.** `codexReviewKickMarkerPath`(`internal/cli/codex_review_delay.go:82`)가 검토 예산 버킷을 이름에 싣고(`.kick-<bucket>`), `kickInFlight`(`:96`)가 배타적 생성 한 번으로 중복을 막는다. 내용은 시각 한 줄이고(`:112–117`), 소유자 pid 는 없다. 완료 시 풀리는 코드가 없다 — 판정이 기록된 뒤에도 버킷이 끝날 때까지 표식이 남는다. 이전 버킷 표식의 정리(`:104–108`)가 유일한 수거다.
- **inconclusive 는 지금 재검토를 시작하지 않는다.** 턴 종료 게이트의 캐시 적중 분기(`internal/cli/codex_review_gate.go:227–259`)는 fresh inconclusive 를 재사용해 `inconclusive` 상태 메시지와 함께 허용하고, 킥 분기에 도달하지 않는다. Q4 = (a)(간격 없음, 다음 턴 종료에서 다시 시작)와 다르다. 이 기대값을 바꾸는 일이 t1674 몫이라는 기록은 SPEC-001 acceptance.md §C 셀 B-2 의 t1674 노트에 있다.
- **차단 횟수의 검사와 기록은 이미 한 단계다.** `codexReviewReserveBlock`(`internal/cli/codex_review_gate.go:380`)이 키별 잠금(`acquireCodexBlockLock`, `:320`) 안에서 검사와 기록을 한다. 잠금 원시는 `acquireGateLockNoFollowImpl` — 유닉스는 `O_NOFOLLOW` 열기, 윈도는 배타적 생성(`internal/cli/gate_lock_unix.go:68`, `internal/cli/gate_lock_windows.go:112`)이다. 경합 가드 `TestReviewGate_ConcurrentBlockCap`(`internal/cli/codex_review_gate_concurrency_test.go:23`)가 있다.
- **블록 잠금 획득은 링크를 따르지 않는다. 그 밖은 따른다.** 게이트 실행 잠금 `AcquireGateLock`(`internal/cli/gate_lock.go:99`)은 여전히 링크를 따르는 열기를 쓰고, 윈도의 고아 블록 잠금 파괴 `codexReviewBreakStaleBlockLock`(`internal/cli/codex_review_gate.go:353`)은 `os.Stat` 과 `os.ReadFile` 로 링크를 따라 읽는다(`:357`, `:365`, `:373`). 블록 잠금 획득 경로의 가드는 `TestReviewGate_BlockLockRefusesSymlink`(`internal/cli/codex_review_gate_symlink_test.go:21`)다.
- **스위치 판독은 작업 트리 파일이다.** `readCodexReviewGateEnabled`(`internal/cli/mcp_codex.go:2602`)가 판독 트리의 작업 트리 `workflow.yaml` 을 읽는다. HEAD 커밋값 판독(D11)은 구현되어 있지 않고, 재정의 파일(`.moai/state/codex-review-gate/run-switch.json`)과 member lane 판정도 없다. 단일 활성 run 해석기 `factorymsg.ResolveActiveRun`(`internal/factorymsg/store.go`, fail-closed)와 카드 쪽 호출 `resolveFactoryCardRun`(`internal/cli/factory_card.go:2875`)은 이미 있다 — 이 SPEC 은 결합을 새로 만들지 않는다(SPEC-001 design.md §5).
- **래퍼의 조용한 종료 다섯 곳이 그대로다.** `.claude/hooks/moai/handle-codex-review-gate.sh` 의 35행(`workflow.yaml` 없음), 36행(`awk` 없음), 60행(게이트 꺼짐), 79–80행(바이너리 없음), 83–86행(처리기 호출 뒤 무조건 `exit 0`). 두 면(저장소 본과 템플릿 미러)은 이 세션 `cmp` 로 바이트가 같음을 확인했다.
- **진입 훅은 차단하지 않는다.** `internal/cli/codex_review_delay.go` 진입 경로는 fail 영수증 상세를 키당 한 번 전달하고 차단하지 않는다(SPEC-001 REQ-RG-001 구현분). SPEC-GATE-BOTTLENECK-001 REQ-GBN-002 의 진입 차단 문언과 어긋나는 표면은 이제 코드가 아니라 그 SPEC 의 본문이다.
- **`tree_scope` 키는 `workflow.codex.review_gate.tree_scope` 다.** 값과 기본값은 `internal/config/defaults.go:840–862`, 타입은 `internal/config/types.go:1047`, 소비는 `treeScopeSkipApplies`(`internal/cli/codex_review_tree_scope.go:132`, 게이트 분기 `internal/cli/codex_review_gate.go:199–201`)다. 기본 구성(값 없음)의 주 체크아웃 스킵이 `tree_scope 건너뜀` 으로 표시되어 운영자를 헷갈리게 한다는 기록은 카드 본문(t1620 lane-12, 커밋 `65c24753a`)에 있고, 이 세션은 스킵 행렬 전체를 재측정하지 않았다(plan.md §F 위험 8).

## 상속 항목 처분 (t1620 → t1674)

카드가 이 계획에 판정을 요구한 상속 항목의 처분과 근거 위치다. 각 행의 처분은 plan.md §B 의 같은 번호와 짝을 이룬다.

| 항목 | 처분 | 이 SPEC 안의 위치 |
|---|---|---|
| 1. F5 (SPEC-GATE-BOTTLENECK-001 REQ-GBN-002·REQ-GBN-004·AC-GBN-004 가 코드와 주석 없이 어긋남) | 기록 처분 지속(운영자 판정 Q6 = (b) 승계). 코드 쪽은 이미 t1620 이 해결했다(진입 훅 비차단, §A.2). 남는 어긋남은 SPEC 대 SPEC 이며, 대상 SPEC 의 개정은 이 카드 밖이다 | Out of Scope — SPEC-GATE-BOTTLENECK-001 편집; plan.md §B 1 |
| 2. tree_scope-skip 명명 | 상태 이름 분리(서명 변경) 대신 사유 표시. 열두 상태 어휘는 SPEC-001 소유로 v0.2.11 에서 동결되었고, 분리는 SPEC-001 개정을 요구한다. 메시지 본문이 명시 설정 제외와 기본 판정 제외를 구별한다 | REQ-RGR-010; plan.md §B 2 |
| 3. R-1 (차단 횟수 검사+증가 경합) | 이미 상류 수리 완료(SPEC-001 v0.2.11 SF-1, `codexReviewReserveBlock` 단일 단계). 수리가 아니라 회귀 가드 재실행으로 이행한다 | AC-RGR-011; plan.md §B 3 |
| 4. 래퍼 처리기 실패 경로(83–86행) 상태 명명 | 이 카드의 REQ. 래퍼 범위 상태 이름 `처리기 실패` 를 002 가 정의한다 | REQ-RGR-009 |
| 5. F2 (게이트 실행 잠금 `gate_lock.go:99` 의 `O_NOFOLLOW` 부재, 윈도 고아 파괴의 링크 추종) | 부채 행(상한 명시). 블록 잠금 획득은 이미 수리되었고(§A.2), 남는 표면은 M5·M6 밖의 `moai gate` 잠금 파일이다. 수리 자체는 두 플랫폼 모두 이미 존재하는 no-follow 원시로 두 줄이지만, 다른 하위시스템(SPEC-GATE-THREE-AXES-001)의 파일과 테스트를 끌어온다 | Out of Scope — 게이트 실행 잠금의 no-follow 전환; plan.md §B 5 |
| 6. iter3 부채 N2·F3·N3·F6·F7·N1 | N2 는 열린 질문 1 로(리더 판정), F3·F6·N3·N1 은 기록 처분(이 카드가 SPEC-001 파일을 바꾸지 않는다), F7 은 선택 항목으로 연기 | Out of Scope — SPEC-001 문서 정정; plan.md §B 6, 「열린 질문」 |
| B1 (iter10 필수 1–3) | `review_gate_registration_test.go`·`codex_review_gate_wiring_test.go` 를 plan §C M6 행과 기대값 변경 목록에 묶고, `review_gate_config_key_test.go` 기대 항목을 D11 에 맞춰 강화한다 | plan.md §C, §D M5·M6 |
| B2 (덮어쓰기 컴파일 전제) | 이 SPEC 의 덮어쓰기 검사가 나타나는 모든 곳에 전제 문장을 함께 둔다 | plan.md §D, acceptance.md §D·§E |

## §B 요구사항 (GEARS)

새 REQ 번호대 `REQ-RGR-`(REVIEW-GATE-REPAIR)를 쓴다. SPEC-001 의 살아 있는 REQ-RG-001–007 번호는 재사용하지 않는다. 폐기된 식별자 대응은 아래 표와 같다. 표의 구 번호는 SPEC-001 v0.2.2 번호계의 것으로, SPEC-001 v0.2.7·v0.2.8 이 t1674 로 이관하고 그 뒤 재번호 매기기로 없어진 식별자다(SPEC-001 HISTORY v0.2.7·v0.2.8 행).

| 구 ID (SPEC-001 v0.2.2 번호계, 폐기됨) | 의미 | 신 REQ |
|---|---|---|
| REQ-RG-007 | 스위치 기본값 = 판독 트리 HEAD 의 커밋된 값(D11) | REQ-RGR-005 |
| REQ-RG-008 | 재정의 파일과 두 층의 바이너리 호출 | REQ-RGR-006, REQ-RGR-008 |
| REQ-RG-011 | 진행 중 표식의 수명(완료 해제, 고아 회수, 캐시 관할) | REQ-RGR-001, REQ-RGR-002, REQ-RGR-003 |
| (구 대응 없음 — 운영자 판정 Q4 = (a) 의 승계) | inconclusive 뒤 간격 없는 재검토 | REQ-RGR-004 |
| (SPEC-001 design.md §7) | 스위치 출처 기록 | REQ-RGR-007 |
| (SPEC-001 v0.2.10·v0.2.11 배정) | 래퍼 35·36·60행·79–80행·83–86행 | REQ-RGR-008, REQ-RGR-009 |
| (카드 상속 항목 2) | `tree_scope 건너뜀` 사유 표시 | REQ-RGR-010 |

### REQ-RGR-001 — 표식은 소유자 pid 를 싣고 검토가 어떤 판정으로 끝나든 완료 시점에 풀린다 (event-driven, M5)

When a background review records its verdict — pass, fail, or inconclusive — the review shall release the review key's in-progress marker at the step that records the verdict.

표식은 검토가 시작될 때 만들어져 소유자 pid 를 싣는다(SPEC-001 design.md §3 (가)). 소유자는 `moai verify codex-review` 로 분리 실행된 검토 프로세스다. 판정의 종류를 묻지 않고, 판정을 기록하는 단계에서 표식을 푼다. 현재 표식은 버킷 이름의 시각 한 줄이고 완료 시 풀리는 코드가 없다(§A.2). 구 REQ-RG-011 의 (가) 절이다.

### REQ-RGR-002 — 고아 표식은 새 검토가 시작되거나 새 판정이 기록되기 전에 회수된다 (event-detected, M5)

When an in-progress marker's recorded pid is not alive, or its pid cannot be read, the actor about to start a new review or record a new verdict shall reclaim the marker first, and shall treat an unreadable marker as an orphan (fail-open).

소유자 pid 가 죽은 표식, pid 를 쓰기 전에 남은 표식, 이전 형식의 표식(시각 한 줄, 빈 파일)은 모두 고아로 본다(SPEC-001 design.md §3 (다), 레인 기본값 (가)). 판정이 불확실하면 회수하는 쪽을 택한다. 표식이 검토를 영구히 막는 것을 막기 위해서이며, 비용은 드문 중복 검토 한 번이다. 구 REQ-RG-011 의 (다) 절이다.

### REQ-RGR-003 — 표식은 동시 실행만 막는다; 재검토 여부는 결과 캐시가 정한다 (Ubiquitous, M5)

The in-progress marker shall block only a concurrent review of the same review key. While a fresh pass or fail receipt backs the current review key, the gate shall decide reuse from the result cache, not from the marker; the marker shall never extend a block past the verdict that released it.

같은 커밋의 재검토를 막는 것은 결과 캐시(검증 영수증 저장소)다. pass 와 fail 은 재사용되어 재검토를 막고, inconclusive 는 pass 가 아니므로 재검토를 막지 않는다(SPEC-001 design.md §3 결과 캐시 항목). 표식이 캐시의 판단을 덧씌워 재검토를 막는 일은 없다. 구 REQ-RG-011 의 (라) 절이다.

### REQ-RGR-004 — fresh inconclusive 턴 종료는 허용하고 상태 메시지를 내며 간격 없이 재검토를 시작한다 (event-driven, M5)

When a turn ends with a fresh inconclusive receipt for the current review key, the turn-end gate shall allow the stop, shall name `inconclusive` in its system message, and shall start a new review for that key without an added interval — unless a review for that key is already in flight.

운영자 판정 Q4 = (a)(간격 없음 — 다음 턴 종료에서 다시 시작하고, 진행 중 표식으로만 중복을 막는다)의 구현이다. 현재 코드는 fresh inconclusive 를 캐시에서 재사용해 킥에 도달하지 않는다(§A.2). 재시작은 REQ-RGR-001 이 풀어 둔 표식과 REQ-RGR-002 의 고아 회수 덕분에 다음 턴 종료에서 가능하다. 재검토가 또 inconclusive 로 끝나면 그 다음 턴 종료가 또 시작한다 — 상한은 두지 않는다(Q4 문언). 메시지는 REQ-RG-007(SPEC-001)이 정한 `inconclusive` 상태 그대로다. design.md §2·§6 의 t1674 몫이다.

### REQ-RGR-005 — 스위치 기본값은 판독 트리 HEAD 의 커밋된 값이다 (Where, M5)

Where the override file does not apply, the gate shall read `workflow.codex.review_gate.enabled` from the HEAD commit of the reading tree; When the committed value cannot be read, the gate shall allow without starting a review and shall emit the `커밋 값 읽기 실패` state message.

작업 트리의 미커밋 값은 쓰지 않는다(보드 판정 D11). 판독은 `git show HEAD:` 같은 읽기 전용 조회다. 판독에 실패하면(파일이 없거나, 키가 없거나, 값이 불리언이 아니면) 검토를 시작하지 않고 허용하며 fail-open 한다(리더 판정 d-20261010T083549Z-d5ad). 현재 판독기는 작업 트리 파일을 읽는다(§A.2). 구 REQ-RG-007 이다. multi 게이트 판독기는 바꾸지 않는다(§E).

### REQ-RGR-006 — 재정의 파일은 규칙을 만족할 때만 적용된다 (When, M5)

When a schema-valid override file exists at the primary checkout's `.moai/state/codex-review-gate/run-switch.json`, its `run_id` equals the single active run, and the session's tree is a member lane of that run, the gate shall apply the file's `codex_review_gate` value; in every other case the gate shall treat the file as absent and shall follow the committed value.

형식(JSON 객체, 필드 `run_id`·`codex_review_gate`), 무시 조건, 단일 활성 run 해석(0개·2개 이상은 fail-closed), member lane 판정(세션 트리 루트의 정규화 경로가 활성 run 의 `cards.worktree_path` 행과 같을 때만 member; 정규화 불가는 member 아님 — 레인 기본값 (나))은 SPEC-001 design.md §5 가 고정한 대로다. 코드 근거: `factorymsg.ResolveActiveRun`(`internal/factorymsg/store.go`), `resolveFactoryCardRun`(`internal/cli/factory_card.go:2875`). 경로에서 run 으로 가는 조회(`run_id` 와 `worktree_path` 로 읽기 한 번)는 run 단계가 추가하며, 결합 자체는 새로 만들지 않는다. 구 REQ-RG-008 의 재정의 절반이다.

### REQ-RGR-007 — 스위치 출처를 관측 행에 기록한다 (event-driven, M5)

When the gate resolves the review-gate switch, the gate shall record on its stderr observation row whether the value came from the override file or from the committed value.

스위치 출처 기록은 design.md §7 이 t1674 와 함께 오도록 남겨둔 항목이다. 운영자가 `스위치 꺼짐` 메시지를 보았을 때 그 값이 어느 층에서 왔는지 추적할 수 있게 하는 것이 목적이다. 기록은 기존 관측 행(표준 오류의 JSON 행)에 필드를 더는 방식을 따른다.

### REQ-RGR-008 — 턴 종료 래퍼는 해석하지 않고 모든 턴 종료에서 바이너리를 호출한다 (Ubiquitous, M6)

The turn-end gate wrapper shall not interpret the switch in the shell and shall invoke the Go handler on every turn end; the switch-off state message (`스위치 꺼짐`) is emitted by the handler.

v0.2.6 판정(운영자 결정 51a4, design.md §5 「두 층」)의 문언대로다. 셸 자체 게이트(`handle-codex-review-gate.sh` 32–60행의 `workflow.yaml` 판독과 `awk` 해석)를 래퍼에서 없애면 35행·36행·60행의 조용한 종료는 경로 자체가 사라지고, 게이트가 꺼진 턴 종료에서도 처리기가 `스위치 꺼짐` 메시지를 낸다(보드 판정 D4). 대가 — 스위치가 꺼진 배포 기본 상태의 프로젝트는 턴 종료마다 바이너리 기동 1회를 낸다 — 는 같은 판정이 이미 수용했다. 이전의 「꺼짐이면 무출력 종료」는 폐기된다. 진입 래퍼(`handle-codex-review-entry.sh`)는 이 요구 대상이 아니다.

### REQ-RGR-009 — 래퍼의 조용한 종료와 처리기 실패는 래퍼 범위 상태 이름을 낸다 (event-detected, M6)

When the wrapper ends without the handler's output — the moai binary absent in all three resolution tiers, or the handler invocation exiting non-zero — the wrapper shall emit an allow carrying a system message that names the wrapper-scope state: `바이너리 없음` for the absent binary, `처리기 실패` for the failed handler invocation.

Go 처리기가 도달할 수 없는 경로이므로 메시지를 내는 층은 래퍼뿐이다. 두 이름은 이 SPEC 이 정의하는 래퍼 범위 상태 이름이며, SPEC-001 design.md §7 의 열두 상태(처리기가 내는 어휘, v0.2.11 에서 동결)에는 더하지 않는다 — 어휘가 두 층에 갈라지는 것을 막기보다, 처리기가 도달 불가능한 경로만 래퍼가 이름 붙인다고 경계를 나눈다. 79–80행과 83–86행이 대상이다. 35·36·60행은 REQ-RGR-008 이 경로를 없앤다. 래퍼는 이 경우에도 0 으로 끝나 Stop 파이프를 깨지 않는다.

### REQ-RGR-010 — `tree_scope 건너뜀` 메시지는 사유를 구별해 담는다 (event-driven, M6)

When the tree-scope skip allows a turn end, the system message shall keep the state name `tree_scope 건너뜀` and shall additionally name whether the skip comes from an explicitly configured `tree_scope` value or from the default policy without configuration.

상속 항목 2 의 처분(사유 표시)이다. 기본 구성의 주 체크아웃 스킵이 `tree_scope 건너뜀` 으로 표시되면 운영자는 자기가 설정한 적 없는 정책이 동작했다고 읽는다(카드 기록, §A.2). 열두 상태 어휘는 SPEC-001 소유로 동결되어 있으므로 상태 이름을 쪼개지 않고 메시지 본문이 구별한다. 허용 결정 자체는 바뀌지 않는다(SPEC-001 REQ-RG-006).

---

## §C AC 와 검증 구속

- 모든 인수 기준은 acceptance.md 에 있다(AC-RGR-001–012, 요구와 1:1 과 회귀 가드 둘, 고아 기준 없음). 이 SPEC 은 plan 시점에 릴리스 게이트를 하나도 갖지 않는다. 이유: 행위 요구를 단언하는 테스트 실행을 plan 단계에서 하지 않았고(카드 제약), 신규 RED 파일 세 개는 run 단계에서 만들어진다. 기존 테스트 가운데 목표 행위를 단언하는 것은 없으며, `TestReviewGate_InconclusiveReceiptIsNotPassEvidence` 와 `TestReviewGateWrappersSelfGateBeforeBinaryResolution` 은 결함(목표와의 차이)을 현재 기대값으로 고정하고 있다(acceptance.md §C).
- 검증 명령은 항상 작업 트리 루트에서, 단일 호출로 실행한다. 실행된 테스트 수를 함께 기록하며, 0개는 측정 불가로 보고한다.

## §D 실행 순서 구속

1. 열린 질문 1(처리기 오류 경로의 상태 명명)은 리더 판정 대상이다. 판정 전에 M6 의 GREEN 은 `runCodexReviewGate` 의 처리기 오류 분기를 바꾸지 않고, `codex_review_gate_wiring_test.go` 의 기대값 변경도 하지 않는다.
2. 각 마일스톤의 RED 선행은 plan.md §D 의 커밋 그래프 검사로 확인하고, 그 원문은 progress.md §E.2 에 남긴다. 덮어쓰기 단위의 테스트는 첫 부모의 비테스트 코드만으로 컴파일되어야 한다(덮어쓰기 컴파일 전제). 컴파일 실패는 `tdd-result-contract.md` 의 TOOL_FAILURE 이며 RED 로 세지 않는다.
3. **변경 세트(change set)는 커밋 하나다.** 래퍼의 두 면(`.claude/hooks/moai/handle-codex-review-gate.sh` 와 `internal/template/templates/.claude/hooks/moai/handle-codex-review-gate.sh`)은 같은 커밋에서 갱신한다. 두 면이 서로 다른 커밋에 갈라지는 상태는 허용하지 않는다.
4. M5 가 M6 보다 먼저다. 래퍼가 셸 해석을 벗는 순간 스위치 판독은 전적으로 Go 층(커밋값 + 재정의 파일)에 의존하므로, Go 층 해석(REQ-RGR-005·006)이 먼저 자리해야 한다.

---

## §E 범위 밖 (Out of Scope)

이 SPEC 은 SPEC-REVIEW-GATE-REPAIR-001 이 t1674 로 이관한 M5·M6 만 다룬다. 구 REQ-RG-005(pass 재사용)·REQ-RG-006(건너뜀 결정 불변)의 살아 있는 문장은 SPEC-001 이 여전히 소유하며(M7 회귀), 이 SPEC 은 그 문장을 다시 요구하지 않고 회귀 가드로 지킨다(AC-RGR-012).

### Out of Scope — SPEC-GATE-BOTTLENECK-001 의 편집 (상속 항목 1)

- 대상 SPEC 의 REQ-GBN-002(진입 차단 절), REQ-GBN-004, AC-GBN-004 는 편집하지 않는다. 운영자 판정 Q6 = (b)(보드 기록 d-20261010T054047Z-3fc2)의 기록 처분을 승계한다. 코드 쪽은 이미 t1620 이 REQ-RG-001 로 해결했다. 대상 SPEC 개정은 별도 카드로 추적한다.
- 대상 SPEC 의 편집을 이 카드의 run 단계 승인 파일로 넣지 않는다.

### Out of Scope — 게이트 실행 잠금의 no-follow 전환 (상속 항목 5, F2 잔여)

- `internal/cli/gate_lock.go:99`(`AcquireGateLock`)의 링크 추종 열기와 윈도 `codexReviewBreakStaleBlockLock`(`internal/cli/codex_review_gate.go:353`)의 `os.Stat`·`os.ReadFile` 링크 추종은 이 SPEC 에서 고치지 않는다. 부채로 기록한다: 상한은 같은 디렉터리에 쓰기 권한을 가진 로컬 행위자가 `.moai/state/gate-run.lock` 자리에 심볼릭 링크를 놓는 경우이며, 블록 잠금 획득 경로는 이미 no-follow 로 닫혀 있다. 승격 조건은 `gate_lock*.go` 를 다시 만지는 다음 변경이다. 수리는 두 플랫폼 모두 이미 존재하는 `acquireGateLockNoFollowImpl` 로의 전환이다.
- 블록 잠금 획득 경로(`acquireCodexBlockLock`)는 이 SPEC 에서 다시 고치지 않는다(이미 수리됨, AC-RGR-011 이 지킨다).

### Out of Scope — SPEC-001 문서의 정정 (상속 항목 6 — N1·N3·F6)

- SPEC-001 `spec.md:53` 의 문장이 사실과 다르다는 카드 기록(N1), 잔여 행 표기의 문서 간 불일치(N3), decision-index.md 129·158행의 폐기 식별자 인용(F6)은 이 SPEC 에서 고치지 않는다. `.moai/specs/SPEC-REVIEW-GATE-REPAIR-001/` 아래 어떤 파일도 편집하지 않는다. N1 의 무엇이 어떻게 다른지는 이 세션에서 검증하지 않았다(카드 보도, 미검증 — plan.md §A Gap).
- Go 765–769행(빈·형식 오류 입력, SPEC-001 표기; 본 트리 `runCodexReviewGate` `:767–773`)의 상태 이름 부재는 SPEC-001 이 문서화한 예외로 남는다. 이 SPEC 은 그 경로에 이름을 더하지 않는다.

### Out of Scope — multi 게이트와 진입 래퍼

- `readMultiReviewGateEnabled`(`internal/cli/multi_review_gate.go:214`)의 D11 전환은 하지 않는다. 커밋값 판독은 codex 게이트에만 적용한다(구 REQ-RG-007 문언).
- 진입 래퍼 `handle-codex-review-entry.sh` 와 진입 훅의 전달 기록·상세 전달 로직은 바꾸지 않는다(SPEC-001 M2 구현분).

### Out of Scope — 열두 상태 어휘의 변경

- SPEC-001 design.md §7 의 처리기 상태 어휘(열두 상태)에 이름을 더하거나 빼지 않는다. REQ-RGR-009 의 `바이너리 없음`·`처리기 실패` 는 래퍼 범위의 별도 이름이며, 처리기 어휘와 섞이지 않는다고 REQ 본문이 경계를 나눈다.

### Out of Scope — F7 과 운영 정리 항목

- SF-5 테스트 이름과 SF-8 부재 기본값의 정리(F7, 선택), 차단 사유 중복 제거·길이 상한, kick 중복 제거의 확장, 마커 일괄 정리(t1637 몫), 진입 훅 timeout 정렬은 이 SPEC 밖이다.

---

## §F 제약

- 이 SPEC 의 산출물은 `.moai/specs/SPEC-REVIEW-GATE-REPAIR-002/` 안에만 둔다.
- 작업 트리의 `.moai/config/sections/workflow.yaml` 은 건드리지 않는다. `.moai/specs/SPEC-REVIEW-GATE-REPAIR-001/` 을 건드리지 않는다.
- 기준 코드 SHA 는 `8249bff8c3`(이 워크트리 HEAD, 로컬 main 꼭짓점)다. 측정은 트리의 소스에서 만든 빌드로 하고, 설치된 빌드의 결과를 트리의 증거로 인용하지 않는다.

## §G 참조

- plan.md — 마일스톤과 기술 접근, 상속 처분의 판정 근거, 열린 질문
- acceptance.md — 인수 기준 매트릭스와 기준선
- progress.md — §E 골격(run 단계 증거는 §E.2 에 쌓인다)
- `.moai/specs/SPEC-REVIEW-GATE-REPAIR-001/` — spec.md(HISTORY v0.2.3·v0.2.6–v0.2.11, REQ-RG-007 범위 한계 표), design.md §3·§5·§7(t1674 몫 규칙의 원문), plan.md §C(파일 지도 패턴), acceptance.md §C(셀 B-2 t1674 노트), decision-index.md(Q2–Q6, D3·D4·D11, 레인 기본값 (가)/(나))
