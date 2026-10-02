---
id: SPEC-CODEX-REVIEW-ASYNC-001
title: "구현 계획 — Claude Stop 훅 codex 리뷰 게이트의 asyncRewake 전환"
version: "0.1.0"
created: 2026-10-02
---

# SPEC-CODEX-REVIEW-ASYNC-001 — 구현 계획

## §A 맥락

트리 `.moai/worktrees/t1422`, 브랜치 `WT-codex-review-lane-scope`, 측정 HEAD `3ae43ed8e78ffa673ca238227df6ca7202c1ce70`. 카드 t1422. `quality.yaml` `development_mode: tdd`(`.moai/config/sections/quality.yaml:2`) → cycle_type=tdd.

**개수 규칙과 Tier.** 요구 = `### REQ-` 제목 수 **13**, 수락 기준 = `### AC-` 제목 수 **13**(하위 ID 없음) — 둘 다 Tier M 상한 16 이하. LOC: 비테스트 ≈220 + 테스트 ≈350-450 = Tier M(300-1000). 파일: 코드+테스트 ≈14(Tier M 대역 5-15 의 안), 미러·문서 ≈5(기계적), 합 ≈19(§H) — **합계만 보면 Tier L(>15) 대역이다.** 상한을 조용히 풀지 않고 O-J 로 올렸고 **Tier M 유지로 해결됐다(Jev 0.70, §G)**. 근거: LOC 가 M 대역이고, 합계를 넘기는 5개는 래퍼·설정·CHANGELOG 의 기계적 미러라 L 의 design.md·research.md 가 새 정보를 더하지 않는다(형제 SPEC 이 같은 논거로 Jev 0.68 에 Tier M 을 유지한 선례).

마일스톤 순서는 §I, 결정 순서는 §B. **관측(프로브)이 설계를 게이트한다**(§D). 형제 SPEC(`SPEC-CODEX-REVIEW-OWNERSHIP-001`)이 먼저 착지한다는 전제(`depends_on`) 위에 선다.

---

## §B 설계 결정

### B.1 핸들러·실행기 분할 — 시그니처 불변

`HandleCodexReviewGate(input, enabled, projectDir)` 의 시그니처는 **바꾸지 않는다**(테스트 호출 23곳 — 형제 plan-audit 실측; 형제 SPEC 도 불변). 핸들러가 락·상태·요약 문구(`Reason`)까지 구성하고 `Decision: block` 을 돌려준다. **실행기** `runCodexReviewGate`(`codex_review_gate.go:192-211`)가 그것을 신호로 옮긴다: block → `Reason` 을 `cmd.ErrOrStderr()` 에 쓰고 stdout 은 비우며 `exitCodeError{code: 2}`(`constitution.go:305-313`, `cmd/moai/main.go` 의 ExitCoder 처리)를 반환한다; allow → 아무것도 출력하지 않고 종료 코드 0. 오류 분기(`gateErr != nil`)는 stderr 진단 후 종료 코드 0. cobra 가 `exitCodeError` 메시지에 `Error:` 접두를 붙여 stderr 에 한 번 더 쓰는지는 run 이 관측해 `SilenceErrors` 등으로 막는다(AC-001 이 stderr 첫 줄이 요약 머리말임을 고정). 기존 시험 중 stdout `{}` 를 기대하는 것(`codex_review_gate_wiring_test.go`, `hook_e2e_test.go` 계열)은 M2 에서 새 계약으로 고친다.

핸들러 새 흐름(기존 번호 유지, 삽입 지점만 명시):

1. 비활성 → `stop_hook_active` → 스코프 해상·로그 → *(형제: `tree_scope: skip` 판정 — 락 이전)* → 스코프 셀프게이트 → codex 조회. (불변)
2. **[삽입] 트리 정규화** `auditreceipt.TreeRootFromCWD(scope.Dir)`(`treeroot.go:22`). 빈 값이면 락·상태 없이 진행(REQ-CRA-013 에 기록).
3. **[삽입] 리뷰 락** `AcquireReviewLock(treeRoot)`. 경합이면 REQ-CRA-013 행을 쓰고 allow. 기계 오류면 행을 쓰고 락 없이 진행(fail-open). 성공이면 `defer Release`.
4. **[삽입] 시작 상태** `codexReviewReceiptStateForScope`(`codex_review_receipt.go:86`)로 `S0 = (Head, TreeDigest)`. 오류면 상태 기반 단계(5, 8, 9)를 건너뛰고 진행.
5. **[삽입] 알려진 상태 건너뜀** 재전달 기록의 `state_key == S0.key` 이고 `verdict ∈ {pass, fail}` 이면 행을 쓰고 allow(REQ-CRA-005).
6. 리뷰 실행. (불변: `runCodexReviewRPC`, 900s)
7. verdict 분기. (불변: `isBlockVerdict`)
8. **[삽입] 종료 상태** `S1` 재계산, `stale = S1.key != S0.key`.
9. **[삽입] 기록·상한** findings 지문 `fp` = 정렬한 `severity|file|line|title` 목록의 해시. pass → 기록 `{key:S0.key, verdict:pass, consecutive_fail:0, fail_fingerprint:""}`; inconclusive → 기록 안 함. fail: `streak = (fp == 기록.fail_fingerprint) ? 기록.consecutive_fail + 1 : 1`. **`기록.consecutive_fail >= cap`(= 이미 마지막 알림을 보냈음) 이고 `fp` 가 같으면** 행(`failure-cap`)만 쓰고 allow(침묵). 그 외에는 `{key:S0.key, verdict:fail, consecutive_fail:streak, fail_fingerprint:fp}` 를 기록하고 요약을 구성해 block — `streak == cap` 이면 요약 끝에 마지막 알림 문구(§B.2)를 붙인다.

### B.2 요약 문구 (REQ-CRA-001, 006)

```text
codex review gate: FAIL
card: <워크트리 디렉터리 이름 — WT- 카드 워크트리일 때만, 아니면 공란>   branch: <브랜치>
reviewed HEAD: <S0.Head>
[stale 일 때]
STALE: this is a review of an earlier state (reviewed HEAD <S0.Head>, current HEAD <S1.Head>)
[연속 실패가 상한(3)에 도달한 마지막 알림일 때]
FINAL NOTICE: further identical failures of this tree are suppressed. No further notifications will follow; the state is UNRESOLVED.
이후 알림 없음, 상태는 미해결
summary: <ReviewOutput.Summary>
findings (<n>):
  - [<severity>] <file>:<line> <title>
  … (최대 <상한>개, 초과분은 "(+k more)")
```

- "카드" 칸(Q15 확정, Jev 1.00): 코드에는 카드 id 의 출처가 없다. 레인 규약상 워크트리 디렉터리 이름이 카드 id 다(`.claude/rules/local/gitflow-lane-protocol.md` §1) — `WT-` 카드 워크트리일 때(`cardScopeFromBranch(scope.Branch)`)만 그 이름을 적고, 카드 워크트리가 아닌 트리(리더가 앉은 primary 포함)에서는 공란이다.
- 마지막 알림의 고정 문구 `이후 알림 없음, 상태는 미해결` 은 시험이 정확히 일치를 보는 문자열이다(AC-005). 영어 문장은 같은 뜻의 병기다.
- findings 상한·연속 실패 상한·타임아웃 마진은 `internal/config` 단일 원천(`defaults.go`).

### B.3 락 — gate-run 락 원시 기능의 일반화 (재사용 우선)

`internal/cli/gate_lock.go` 의 `acquireGateLockImpl(lockPath)`(유닉스 flock `LOCK_EX|LOCK_NB`, 보유자 PID 기록, 커널이 종료 시 해제 — `gate_lock_unix.go:38-61`)는 **이미 경로 매개변수형**이다. 고칠 곳은 이름이 상수에 묶인 윗층뿐이다: `AcquireGateLock`·`ReadGateLockHolder`·`ClearStaleGateLock` 이 `gateLockPath(projectDir)` 고정 경로를 쓴다(`gate_lock.go`, 윈도우 `gate_lock_windows.go:129`). 일반화: `acquireStateLock(projectDir, name)` 류의 이름 매개변수 헬퍼를 두고 기존 gate-run 락 함수들은 그것을 `gate-run.lock` 이름으로 부르게 한다(동작 불변 — 기존 `gate_lock` 시험 초록이 회귀 칸). 새 `codex-review.lock` 은 같은 헬퍼로 `<트리>/.moai/state/codex-review.lock`. 죽은 보유자 인수: 유닉스는 flock 이 자동 해제(시험: 죽은 PID 를 적은 낡은 파일 위에서 획득 성공), 윈도우는 `ClearStaleGateLock` 의 이름 매개변수형(`kanban.FactoryProcessAlive` 확인+changed-hands 중단).

선택하지 않은 안: `internal/lockfile`(보유자 신원 없음 → 죽은 보유자 판정 불가), `internal/verify/claim_lock.go`(5분 TTL — 최대 900s 리뷰에 맞지 않는 만료로 살아 있는 보유자를 인수해 중복 실행을 만든다).

### B.4 재전달 기록과 억제 로그

- 기록: `<트리>/.moai/state/codex-review-wake.json` — `{state_key, verdict, consecutive_fail, fail_fingerprint, recorded_at}`. 락 안에서만 읽고 쓰며 `atomicfile.Replace`(`internal/atomicfile/replace.go:20`)로 임시 파일을 바꿔 쓴다(반쯤 쓴 파일을 다른 프로세스가 읽지 않게). 읽기 실패·파싱 실패는 "기록 없음"으로 취급(fail-open, 한 번 더 리뷰가 돌 뿐).
- **기록은 Codex 경로와 공유하지 않는다**(검증 영수증 저장소 불사용). 영수증을 공유하면 Claude 훅이 남긴 fail 이 Codex 멤버 6 의 차단 근거가 되어 Codex 경로 동작이 바뀐다(REQ-CRA-009). → Q13 해결: 공유하지 않음(Jev 0.02 로 약하나 SPEC 이 이미 가진 설계이고 리더가 수용; Codex 경로 불변).
- 억제 로그: `<트리>/.moai/logs/codex-review-gate.log` 에 JSON 한 줄(`ts`, `reason ∈ {lock-held, state-unchanged, failure-cap, lock-unavailable, state-unavailable}`, `tree`, `state_key`). `.moai/logs/` 는 런타임 접두라 리뷰 대상이 아니다(`codex_review_gate.go:39`).

### B.5 래퍼 (REQ-CRA-003)

```text
printf '%s' "$INPUT" | "$MOAI_BIN" hook codex-review-gate
rc=$?
[ "$rc" -eq 2 ] && exit 2     # stderr 는 그대로 통과
exit 0
```

순수 셸 off 스위치(`:35-60`)와 바이너리 부재 시 `exit 0`(`:78-81`)은 그대로다. 파이프라인 종료 상태는 `moai` 의 것이다(`printf` 가 앞이므로 `PIPESTATUS` 를 쓰거나 here-string 을 쓰는지는 run 이 셸 호환 범위에서 고른다 — 시험이 종료 코드 행렬로 고정). 두 사본(C1 `.claude/hooks/moai/…`, C2 `internal/template/templates/.claude/hooks/moai/…`)을 같은 변경에 고친다. C2 본문은 중립(SPEC ID·카드 번호 없음).

### B.6 등록과 타임아웃 정책 (REQ-CRA-007, 008)

- 등록: `.claude/settings.json:190-194` 항목과 `settings.json.tmpl:191` 항목에 `"asyncRewake": true` 추가. multi 래퍼 항목은 손대지 않는다. `internal/template/hook_entries.go` 의 `HookEntry`(:17-30)에 `AsyncRewake bool` 을 더하고 `settingsDoc` 파싱(:53-56)·`String()`·정렬 비교(:197)를 맞춘다.
- 시험 고정: `review_gate_registration_test.go` 의 `reviewGateWrappers`(두 래퍼의 `timeout` 900 고정)를 래퍼별 `{timeout, asyncRewake}` 로 확장한다.
- **타임아웃 정책 재검토.** 두 층이 있다: 훅 `timeout`(Claude Code 가 적용) 과 Go 쪽 `DefaultCodexReviewGateTimeout` 900s(`codex_review_gate.go:96`, 프로세스가 스스로 리뷰를 끊고 inconclusive → 종료 코드 0). 공식 표는 `timeout` 을 `async: true` 에 대해서만 "enforce 하지 않는다"고 적는다. `asyncRewake` 는 E-2 미관측이다. 결정표(프로브 뒤 확정):

| E-2 결과 | 훅 `timeout` 의 역할 | 정책 |
|---|---|---|
| 강제하지 않음 | 무의미 | Go 쪽 900s 가 유일한 상한. 등록 값은 Go 상한 + 마진(공통 규칙 유지, 시험 단순) |
| 강제(취소+출력 폐기) | 늦은 실패를 삼킬 위험 | 등록 값 > Go 상한 + 마진 → 프로세스 내부 fail-open(조용한 종료 코드 0)이 항상 먼저 발화 |
| 강제하되 종료 코드 2 는 전달 | 중간 | 위와 동일 규칙으로 안전 쪽 |

마진 값은 프로브 뒤 정한다(제안 시작점 60s → 등록 960 — **미확정**, O-B). 현재 900/900 은 동점이라 취소와 자체 상한 중 먼저 발화하는 쪽이 우연이다.

### B.7 깨움 루프 방지 (hazard i)

REQ-CRA-005 의 두 경계: (1) 같은 상태 키+기록된 pass/fail → 리뷰도 깨움도 없음(현행 변경, Q16), (2) 연속 fail 깨움 3 번째가 마지막 알림 — 고정 문구 `이후 알림 없음, 상태는 미해결` 을 담고 이후 지문이 같은 fail 은 침묵(행은 기록). 침묵이 "통과"로 읽히는 위험을 침묵 대신 마지막 알림 하나로 막는다(O-A 재정). 지문이 다른 fail 은 새 실패라 전달되고 연속 횟수를 1 부터 다시 센다; pass 면 0 으로 돌아간다. 오케스트레이터 문구의 "reviewed state 가 바뀔 때까지"는 "동일한 실패가 달라지거나 pass 할 때까지"로 읽었다(§G 확인 사항). `stop_hook_active` 는 비동기 훅에서 의미가 있는지 미관측(E-3/E-4)이라 그것에 의존하지 않고, 핸들러의 기존 `stop_hook_active` 검사는 그대로 둔다.

---

## §C 재사용 지점

| 필요 | 재사용 | 새로 쓰는 것 |
|---|---|---|
| 트리별 락·보유자 신원·죽은 보유자 | `acquireGateLockImpl`(경로형), `GateLockOwner`, `ErrGateLockHeld`, `kanban.FactoryProcessAlive` | 이름 매개변수 헬퍼(≈25 LOC), `AcquireReviewLock` |
| 트리 정규화 | `auditreceipt.TreeRootFromCWD` | 없음 |
| 리뷰 상태 키 | `codexReviewReceiptStateForScope` | 키 문자열 결합 한 줄 |
| 임시 파일 교체 | `atomicfile.Replace` | 기록 구조체(≈10 LOC) |
| 종료 코드 | `exitCodeError`, ExitCoder | 실행기 매핑(≈15 LOC) |
| 리뷰 실행 | `runCodexReviewRPC`(불변) | 없음 |
| 와이어링 비교 | `HookEntry`/`DiffHookEntries` | 필드 하나 |

---

## §D 프로브 계획 — E-1~E-5 (M1, 이후 마일스톤을 게이트)

**절차 문서:** `.moai/reports/t1422/async-probes.md` (로컬 gitignore 증거, 한국어, 명령 원문). 프로브마다 목적과 답하는 E 항목, 준비(일회용 `mktemp -d` 프로젝트 — 이 저장소의 추적 설정과 primary 체크아웃은 절대 편집하지 않음), 명령, 가설별(깨움/비깨움) 기대 관측, E 항목 판정 기준, 기록 위치, 정리를 적었다. **키트:** `.moai/reports/t1422/async-probe/kit/`(타임스탬프 줄을 덧붙이는 훅 스크립트와 설정 조각). **기록:** 프로브당 `.moai/reports/t1422/async-probe/<프로브 ID>.txt` — 결과 칸은 비어 있고 실행한 쪽이 채운다. `.moai/reports/*` 는 gitignore 대상이므로 각 파일의 sha256 을 tracked `progress.md` 의「프로브 기록」표에 기록한다(AC-011).

| 프로브 ID | E 항목 | 내용 | 실행 주체 | 게이트하는 것 |
|---|---|---|---|---|
| `E1-headless` | E-1 | `asyncRewake: true`, 잠 20s, stderr 마커, 종료 코드 2 인 Stop 훅. `claude -p --input-format stream-json --output-format stream-json` 로 한 턴 후 stdin 을 연 채 추가 이벤트 관찰 | **레인(헤드리스, 먼저)** — 이 변형이 대화형과 같은 동작이라는 것은 미검증이므로 대화형 결과가 우선 | M5 (참고) |
| `E1-interactive` | E-1 | 같은 훅. 대화형 세션에서 한 턴을 끝내고 **입력 없이 대기** — 새 턴이 시작되는가 | **리더만** | **M5(등록)** — 부정이면 REQ-CRA-007/008 철회 |
| `E2-timeout` | E-2 | `timeout: 5`, 잠 20s, 종료 코드 2. 대조: 같은 훅 `async: true` | **레인(헤드리스, 먼저)** | **M4(타임아웃 마진)** |
| `E3-c1-concurrent` | E-3 | 훅이 긴 잠을 자는 동안 빠르게 3턴 입력 → 겹쳐 도는 훅 수 | 레인(스트림 입력 헤드리스, 먼저) | M2 (참고) |
| `E3-c2-newturn` | E-3 | 훅 실행 중 새 턴이 시작될 때 마커의 위치 | **리더만** | **M2** |
| `E3-c3-sessionexit` | E-3 | 훅 실행 중 세션 종료 → 훅 프로세스 생존/고아 여부 | **리더만** | **M2(락·보유자 사망 시나리오)** |
| `E4-d1-json` | E-4 | 종료 코드 0 + stdout `{"decision":"block","reason":"PROBE-JSON"}` — 무시되는가 | **레인(헤드리스, 먼저)** | **M2·M3(신호 계약)** |
| `E4-d2-exit0stderr` | E-4 | 종료 코드 0 + stderr 마커 — Claude 에 보이는가 | **레인(헤드리스, 먼저)** | **M2·M3(스코프 로그 행 위치, O-I)** |
| `E5-reload` | E-5 | 세션 도중 일회용 프로젝트의 `.claude/settings.json` 에 훅 추가/제거 → 다음 턴 반영 | **리더만** | **M5(진행 중 세션 영향)** |

**처리 규칙.** 각 프로브는 결과를 `확인됨/부정/불확정` 중 하나로 progress.md 에 적는다. E-1 은 `E1-interactive` 가 정한다(`E1-headless` 는 보조). 불확정(실행 불가)은 Gap 이고 그 항목을 게이트로 가진 마일스톤은 시작하지 않는다 — 리더가 Gap 을 수용하는 기록이 있으면 예외. 리더 전용 프로브(`E1-interactive`·`E3-c2-newturn`·`E3-c3-sessionexit`·`E5-reload`)는 리더 세션이 절차 문서대로 실행한다. 어떤 프로브도 관측을 미리 주장하지 않는다.

---

## §E 레인 안전 (hazard ii)

- 등록 변경(`.claude/settings.json`, `settings.json.tmpl`)은 마지막 run-phase 마일스톤(M5)에서만 한다. 그 이전 어떤 커밋도 두 파일을 건드리지 않는다(AC-011/012 가 커밋 그래프로 증명).
- 이 저장소의 레인 세션은 등록이 바뀌어도 스스로를 깨우지 않는다: (1) 이 저장소 추적 `workflow.yaml` 에 `review_gate` 키가 없어 순수 셸 off 스위치가 래퍼를 즉시 끝낸다(래퍼 `:35-60`; `grep -c review_gate .moai/config/sections/workflow.yaml` → 0), (2) 깨움은 종료 코드 2 에서만 일어나고 그것은 활성+실패 리뷰에서만 나온다. 진행 중 세션이 설정 편집을 즉시 반영하는지는 E-5 미관측이라 이 레인의 현재 세션이 영향받는지는 단정하지 않는다 — (1)로 안전하다.
- 착지 뒤 primary 의 리더는 형제 SPEC 의 `tree_scope: skip`(리더 행위, 형제 인계 항목)로 제외되고, 그렇지 않은 활성 트리 세션만 이 SPEC 의 비동기 게이트를 받는다.

---

## §F 파일 목록·Template-First 미러·마일스톤 배정

### F.1 Go 프로덕션

| 파일 | 변경 | 마일스톤 |
|---|---|---|
| `internal/cli/gate_lock.go`, `gate_lock_unix.go`, `gate_lock_windows.go` | 이름 매개변수 일반화(동작 불변) | M2 |
| **신규** `internal/cli/codex_review_lock.go` | `AcquireReviewLock`, 재전달 기록 읽기·쓰기, 억제 로그 | M2 |
| `internal/cli/codex_review_gate.go` | 새 흐름 삽입(§B.1)·요약 구성·실행기 매핑 | M2 |
| `internal/config/defaults.go` | 연속 실패 상한·findings 상한·타임아웃 마진 상수 | M2 (마진은 M4) |
| `internal/template/hook_entries.go` | `AsyncRewake` 키 | M4 |
| `internal/cli/hook.go` | 명령 Short/Long 문구(신호 계약) | M5 |

### F.2 테스트

| 파일 | 내용 | 마일스톤 |
|---|---|---|
| **신규** `internal/cli/codex_review_async_test.go` | 신호·침묵 결과·락 동시성(장벽)·알려진 상태·상한·stale·skip 순서·억제 로그 | M1(RED)→M2 |
| `internal/cli/gate_lock_*_test.go` | gate-run 락 회귀(일반화 뒤 초록) | M1(회귀)→M2 |
| `internal/hook/review_gate_selfgate_test.go` | 래퍼 종료 코드 행렬 | M1(RED)→M3 |
| `internal/cli/codex_review_gate_wiring_test.go`, `hook_e2e_test.go` | 신호 계약 변경에 맞춘 기대값 | M2 |
| `internal/template/hook_entries_test.go`, `review_gate_registration_test.go` | `asyncRewake`·래퍼별 timeout | M4→M5 |

### F.3 미러 (C1=로컬, C2=배포)

| C1 | C2 (`internal/template/templates/…`) | 마일스톤 |
|---|---|---|
| `.claude/hooks/moai/handle-codex-review-gate.sh` | `.claude/hooks/moai/handle-codex-review-gate.sh` | M3 |
| `.claude/settings.json`(codex 항목) | `.claude/settings.json.tmpl`(codex 항목) | **M5(마지막)** |
| `CHANGELOG.md` | — | M6 (sync) |

`make build` 로 임베드 템플릿을 재생성한다. 에이전트 정의·규칙 파일은 건드리지 않으므로 `make agents-emit` 은 필요 없다.

---

## §G 위험·Open decisions·해결된 결정

### 위험

1. **E-1 부정(깨우지 않음).** 등록 요구 철회, 핸들러 개선은 동기 등록 아래에서도 유효(종료 코드 2 = 동기 Stop 에서 중단 방지, A.2). 이 경우 리뷰 대기 문제는 그대로이고 별도 대안이 필요하다.
2. **형제 SPEC 의 핸들러 변경.** 형제는 이 SPEC 보다 먼저 착지하고 시그니처를 바꾸지 않는다. 이 SPEC 은 같은 핸들러에 삽입 지점만 더한다(§B.1) — 형제가 착지한 뒤 소스를 다시 읽어 삽입 위치를 확정한다.
3. **t1399** — `MOAI_KANBAN*` 삭제는 `reviewGateEnvContext`(`codex_review_scope.go:328-334`)에 닿는다. 이 SPEC 은 env 를 읽지 않는다.
4. **중복 실행 부류(t1425).** 이 SPEC 은 codex 리뷰 게이트만 막는다. 관측 로그 폭주는 별개 카드.
5. **명시 실행 경로는 락을 잡지 않는다.** `moai verify codex-review` 와 Claude 훅이 같은 트리에서 동시에 돌 수 있다(범위 밖).
6. **오래된 Claude Code.** `asyncRewake` 필드를 모르는 버전의 동작은 미관측(Gap). 동기 등록처럼 돌면 종료 코드 2 는 중단 방지다.
7. **지속적 침묵의 오독.** 상한 도달 뒤 침묵이 "통과"로 읽힐 수 있다 — 그래서 상한은 침묵 대신 마지막 알림 하나(`이후 알림 없음, 상태는 미해결`)로 바꿨고(O-A 재정), 억제 로그가 보조 완화책이다.
8. **래퍼 종료 상태 포착.** 파이프라인의 종료 상태가 `moai` 의 것인지 셸 호환 범위에서 확인해야 한다(M3 시험 행렬이 고정).
9. **스코프 로그 행의 stderr.** 종료 코드 0 훅의 stderr 가 Claude 에 보이면(E-4 d2) 조용하지 않다 — 그때는 로그 파일로 옮긴다(REQ-CGS-010 의 sink 변경이므로 별도 개정 필요).

### 해결된 결정 (Jev `jev-1.13.0`, 운영자 위임, 2026-10-02 — 리더 수용 반영)

| 결정 | 처분 | 신뢰도 / 상태 |
|---|---|---|
| 중복 실행 방지 | 트리별 락, 보유 중이면 건너뜀, 죽은 보유자 인수 | 0.83 |
| 오래된 결과 | 표지를 붙여 전달(`label_stale`), 조용히 버리지도 표지 없이 전달하지도 않음 | 0.99 |
| 구조 | 형제 SPEC 분리(이 SPEC) | 0.93 |
| **O-A** (Q9) 연속 실패 상한 | Jev 선택 `cap3_silent` **0.21(약함)** → 오케스트레이터 재정 **`cap3_notify_once_at_cap`**: 트리당 연속 실패 깨움 3 번째가 마지막 알림 — "동일한 실패는 억제" + 고정 문구 `이후 알림 없음, 상태는 미해결` — 이후 침묵. 재정 사유: 침묵이 "통과"로 오독되는 위험(작성자가 표시한 위험 7). 리더 통지·재정 수용 | **잠정(PROVISIONAL)**, 리더 수용 |
| **O-B** (Q10) 타임아웃 마진 | **미결 — E-2 프로브 전까지 열려 있음.** 60s 는 시작값 제안이지 결정이 아니다 | OPEN |
| **O-C** (Q11) 오래됨의 정의 | `full_state_key` — 상태 키 전체(HEAD+작업 트리 digest)의 어떤 차이든 오래됨, 표지가 두 HEAD 를 적는다. 운영자 문구("HEAD")의 확장 — 리더 수용 | 0.56 |
| **O-D** (Q12) 크기 | 수용(§H) | 0.98 |
| **O-E** (Q13) 영수증 공유 | `do_not_share` — Codex 경로 불변. Jev 신뢰도는 약하나(0.02) SPEC 의 기존 설계이고 리더 수용 | 0.02, **잠정**, 리더 수용 |
| **O-F** (Q14) 프로브 실행 | 레인이 헤드리스를 먼저 시도하고 대화형만 리더가 실행. 확정: 리더가 리더 세션에서 E-1 대화형·E-3(c2,c3)·E-5 를 실행하며 절차는 `.moai/reports/t1422/async-probes.md` | 0.60 |
| **O-G** (Q15) 카드 칸 | 워크트리 디렉터리 이름, `WT-` 카드 워크트리가 아니면 공란 | 1.00 |
| **O-H** (Q16) 알려진 상태 건너뜀(pass 포함) | `skip_known_state_for_pass` — 현행 동작 변경(오늘은 상태가 같아도 Stop 마다 재리뷰). SPEC 본문에만 명시하며 CHANGELOG 항목이 아니다(리더). 상태가 바뀐 Stop 의 리뷰는 회귀 칸(AC-005) | 0.41 (< 0.5), **잠정**, 리더 수용 |
| **O-I** (Q17) 스코프 로그 행 | stderr 유지, E-4 프로브 전까지(`E4-d2-exit0stderr` 가 정한다) | 0.97 |
| **O-J** (Q18) Tier | M 유지. 코드+테스트 ≈14 / 미러·문서 ≈5 | 0.70 |

**확인이 필요한 읽기(새 불일치, §G 보고).** O-A 재정 문구의 "동일한 실패는 reviewed state 가 바뀔 때까지 억제"를 이 SPEC 은 "동일한 실패(findings 지문이 같음)는 실패가 달라지거나 pass 할 때까지 억제"로 읽었다. 같은 상태 재리뷰는 REQ-CRA-005 첫 절이 이미 막고 있어 문구를 문자 그대로 읽으면 상한이 아무것도 제한하지 못하고(상태가 바뀌면 곧 풀림), 마지막 알림의 "이후 알림 없음"과도 모순되기 때문이다. 그리고 "3 번째 깨움이 마지막 알림"으로 읽었다(4 번째 별도 알림 아님). 오케스트레이터가 다르게 읽는다면 REQ-CRA-005·AC-005·§B.1 step 9 가 바뀐다.

---

## §H 규모 추정과 3배 점검

(전부 **추정** — 이 레인은 구현하지 않았다. LOC 는 비테스트 Go 기준.)

| 구성 | 추정 |
|---|---|
| 락 일반화(유닉스·윈도우·윗층) | ≈ 25 LOC |
| 리뷰 락·재전달 기록·억제 로그 | ≈ 90 LOC |
| 핸들러 흐름 삽입·요약 구성 | ≈ 70 LOC |
| 실행기 매핑 | ≈ 15 LOC |
| 상수·`HookEntry`·help 문구 | ≈ 20 LOC |
| 합계 | **≈ 220 LOC** |
| 테스트 | ≈ 350-450 LOC |

**파일 수(코드+테스트와 미러·문서를 분리).** 코드+테스트 ≈ 14(프로덕션 8·신규 1, 테스트 6·신규 1), 미러·문서 ≈ 5(래퍼 C1+C2, settings C1+tmpl, CHANGELOG). 합 ≈ 19. 프로브 키트·기록은 gitignore 대상 로컬 증거라 세지 않는다.

**3배 점검 기준선.** (i) 가장 단순한 기준선 = `exit 0`+JSON 을 종료 코드 2+stderr 로 바꾸고 레지스트리에 플래그를 더하는 것(≈30 LOC): ≈220/30 ≈ **7배, 3배 초과** — 이 기준선은 운영자 요구 2(트리당 하나·재전달 억제)와 3(오래된 결과 표지)을 충족하지 못해 유효한 비교가 아니다 — 크기 수용으로 해결(O-D, Jev 0.98). (ii) 요구 2·3 을 만족시키는 최소 구현 ≈ 락(자체 구현 ≈ 90)+상태·표지(≈ 90)+신호(≈ 30) ≈ 210 → 제안 ≈220 은 ≈ **1.0배**; 락 원시 기능 재사용이 이미 그 안에 반영돼 있다.

---

## §I 마일스톤 (우선순위 순서, 시간 추정 없음)

### M1 — 프로브와 회귀선, RED (우선순위 High)

- §D 프로브(ID: `E1-headless`·`E1-interactive`·`E2-timeout`·`E3-c1-concurrent`·`E3-c2-newturn`·`E3-c3-sessionexit`·`E4-d1-json`·`E4-d2-exit0stderr`·`E5-reload`) 수행·기록. 키트와 절차 문서(`.moai/reports/t1422/async-probes.md`)는 이미 작성돼 있고(결과 칸 공란), 레인은 헤드리스 프로브만, 리더는 리더 전용 프로브를 리더 세션에서 실행한다. 결과를 progress.md 에 `확인됨/부정/불확정` 으로.
- 변경 전 초록 관측: 아래 §J 회귀 칸 시험, gate-run 락 시험.
- RED: 신규 AC 시험을 구현 없이 추가해 `-v` 로 `=== RUN`·`--- FAIL` 을 관측하고 `.moai/reports/t1422/red-async/` 에 구현 전 트리 SHA 와 함께 보존.
- M1 종료: 프로브 결과가 이후 게이트를 연다(§D). E-1 부정이면 M5 철회 기록.

### M2 — 핸들러·락·재전달·stale (우선순위 High) — 게이트: E-3, E-4

- 락 일반화(회귀 초록 유지) → `AcquireReviewLock` → 흐름 삽입 → 요약 구성 → 실행기 매핑. REQ-CRA-001, 002, 004, 005, 006, 010, 013.
- M2 종료: `./internal/cli` 소관 시험 초록(락 동시성·알려진 상태·상한·stale·skip 순서).

### M3 — 래퍼 (우선순위 High) — 게이트: E-4

- 두 사본 동시 수정, 종료 코드 행렬 시험. REQ-CRA-003.

### M4 — 와이어링 비교와 타임아웃 정책 (우선순위 Medium) — 게이트: E-2

- `HookEntry` 에 `AsyncRewake`, 시험 확장, 마진 상수. 설정 파일은 **아직 건드리지 않는다.** REQ-CRA-007(비교 부분), 008(값 결정).

### M5 — 등록 (우선순위 Medium, **마지막 run-phase 변경**) — 게이트: E-1(확인됨), E-5

- `.claude/settings.json` + `settings.json.tmpl` 에 `asyncRewake`·타임아웃 반영, `hook.go` 문구, 등록 시험. 직전 커밋에 프로브 산출물 sha256 기록이 있어야 한다. REQ-CRA-007, 008, 011, 012.

### M6 — 문서 (우선순위 Low, sync 단계)

- CHANGELOG `[Unreleased]`. 인계 항목은 완료 보고에 옮긴다.

---

## §J 자기 검증

| 항목 | 명령 |
|---|---|
| 대상 테스트 | M1 에서 이름을 확정한다. `go test ./internal/cli/ -list 'ReviewAsync\|ReviewLock\|ReviewGate\|GateLock'` 로 열거한 정확한 이름들을 `-run '^(<이름1>\|<이름2>\|…)$' -count=1 -v` 로 돌려 `=== RUN` 수를 열거 수와 대조 |
| 래퍼·등록 | `go test ./internal/hook/ -list 'ReviewGateWrappers'`, `go test ./internal/template/ -list 'ReviewGates\|HookEntries'` 를 같은 정확-이름 형태로 |
| 회귀 칸(변경 전·후 초록) | `TestReviewGate_CodexFailBlocks`·`…CodexPassAllows`·`…FailOpenOnMissingCodex`·`…FailOpenOnCodexError`·`…InconclusiveAllows`, `TestReviewGatesRegisteredInRepoSettings`·`…InTemplateSettings`, `TestReviewGateWrappers_DisabledCostsZeroColdStarts`·`…EnabledReachesBinary` |
| 정적 | `go vet ./internal/cli/... ./internal/hook/... ./internal/template/... ./internal/config/...`, `GOOS=windows GOARCH=amd64 go build ./...` |
| 환경 세척 | `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test …` (한 번의 복합 호출) |
| 범위 침범 | 변경 파일에 `codex_stop_chain.go`, `codex_review_receipt.go`, `multi_review_gate.go`, `internal/verify/` 가 없어야 한다 |

전체 스위트는 로컬에서 돌리지 않는다.

## §K 안티패턴

1. 프로브 전에 "깨운다"를 사실로 쓰거나 등록을 먼저 바꾸기.
2. 락을 세션 단위로 잡기(같은 트리의 여러 세션이 리뷰를 겹친다) 또는 리뷰 호출 뒤에 잡기.
3. 경합 시 기다리기(훅 프로세스가 쌓인다) — 건너뛰어야 한다.
4. 5분 TTL 락(`verify/claim_lock.go`) 재사용 — 살아 있는 900s 보유자를 인수한다.
5. 오래된 결과를 조용히 버리거나 표지 없이 전달하기. 또는 HEAD 만 비교하기.
6. 옛 신호(stdout JSON block)를 병행 유지하기 — E-4 미관측이라 이중 신호의 의미를 모른다.
7. 래퍼가 비0 종료를 그대로 전파하기(2 이외의 오류가 깨움 신호로 읽힌다).
8. 억제를 기록하지 않기(조용한 게이트와 죽은 게이트를 가를 수 없다).
9. 재전달 기록을 Codex 영수증 저장소에 쓰기.
10. 이 SPEC 에서 형제의 `tree_scope`·자기 리뷰 도구 결정을 다시 건드리기.

## §L 참조

- spec.md §A(측정)·§B(요구)·§E(범위 밖)
- acceptance.md — AC·RED/GREEN 두 칸·변이 점검
- decision-index.md — 결정 권위 표기와 EVIDENCE-NEEDED 행
- progress.md — Decision Log·프로브 기록 칸
- `SPEC-CODEX-REVIEW-OWNERSHIP-001`(형제) plan.md §B.3·§G — skip 순서, 핸들러 시그니처 불변
- `.moai/reports/t1422/plan-audit.md` — 형제 plan-audit 1회차(적용한 교훈 D1-D17)
- `.claude/rules/moai/development/verification-completeness.md` §1.3(연속 발화)·§2(두 칸 채택)
