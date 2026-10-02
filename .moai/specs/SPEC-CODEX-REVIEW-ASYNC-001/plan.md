---
id: SPEC-CODEX-REVIEW-ASYNC-001
title: "구현 계획 — Claude Stop 훅 codex 리뷰 게이트의 asyncRewake 전환"
version: "0.1.0"
created: 2026-10-02
---

# SPEC-CODEX-REVIEW-ASYNC-001 — 구현 계획

## §A 맥락

트리 `.moai/worktrees/t1422`, 브랜치 `WT-codex-review-lane-scope`, 측정 HEAD `a0d8014096f894c3a45577482b0eb92892716e04`(인용 코드는 최초 핀 `3ae43ed8e` 과 동일 — `git diff --stat 3ae43ed8e a0d801409 -- internal .claude cmd .moai/config` 가 비어 있음). 카드 t1422. `quality.yaml` `development_mode: tdd`(`.moai/config/sections/quality.yaml:2`) → cycle_type=tdd.

**개수 규칙과 Tier.** 요구 = `### REQ-` 제목 수 **15**, 수락 기준 = `### AC-` 제목 수 **15**(하위 ID 없음, 요구와 1:1) — 둘 다 Tier M 상한 16 이하(여유 1). LOC: 비테스트 Go ≈245 + 래퍼 셸 ≈35 + 테스트 ≈400-500 = Tier M(300-1000). 파일: 코드+테스트 ≈14(Tier M 대역 5-15 의 안), 미러·문서 ≈5(기계적), 합 ≈19(§H) — **합계만 보면 Tier L(>15) 대역이다.** 상한을 조용히 풀지 않고 O-J 로 올렸고 **Tier M 유지로 해결됐다(Jev 0.70, §G)**. 근거: LOC 가 M 대역이고, 합계를 넘기는 5개는 래퍼·설정·CHANGELOG 의 기계적 미러라 L 의 design.md·research.md 가 새 정보를 더하지 않는다(형제 SPEC 이 같은 논거로 Jev 0.68 에 Tier M 을 유지한 선례). 코드+테스트 "≈14" 는 하한이다(감사 D13: 프로덕션 8·테스트 행 6 중 일부가 파일 둘/글롭을 담아 실제 ≥15) — 정직한 표기이며 O-J 판정은 그대로다.

마일스톤 순서는 §I, 결정 순서는 §B. **관측(프로브)이 설계를 게이트한다**(§D). 형제 SPEC(`SPEC-CODEX-REVIEW-OWNERSHIP-001`)이 먼저 착지한다는 전제(`depends_on`) 위에 선다. 형제는 지금 `status: draft` 이므로 `/moai run` 의 depends_on 사전 점검(완료만 충족)이 **M1 포함 첫 실행**을 막는다 — wait / `--ignore-deps`(근거 기록) / abort 를 리더가 정한다(감사 D14).

---

## §B 설계 결정

### B.1 핸들러·실행기 분할 — 시그니처 불변

`HandleCodexReviewGate(input, enabled, projectDir)` 의 시그니처는 **바꾸지 않는다**(테스트 호출 23곳 — 형제 plan-audit 실측; 형제 SPEC 도 불변). 핸들러가 락·상태·요약 문구(`Reason`)까지 구성하고 `Decision: block` 을 돌려준다. **실행기** `runCodexReviewGate`(`codex_review_gate.go:192-211`)가 그것을 신호로 옮긴다: block → `Reason` 을 `cmd.ErrOrStderr()` 에 쓰고 stdout 은 비우며 `exitCodeError{code: 2}`(`constitution.go:305-313`, `cmd/moai/main.go` 의 ExitCoder 처리; `fang.go` `moaiErrorHandler` 는 ExitCoder 를 추가 출력하지 않음 — 감사 관측)를 반환한다; allow → 아무것도 출력하지 않고 종료 코드 0. 오류 분기(`gateErr != nil`)와 stdin 파싱 실패 분기는 stderr 진단 후 **빈 stdout**·종료 코드 0(현재의 `return emitHookOutput(` 3곳이 모두 사라진다 — `grep -c 'return emitHookOutput(' internal/cli/codex_review_gate.go` 3 → 0). 기존 시험 중 stdout `{}` 를 기대하는 것(`codex_review_gate_wiring_test.go`, `hook_e2e_test.go` 계열)은 M2 에서 새 계약으로 고친다.

**전달 확인 표(ticket).** 핸들러는 block 을 돌려주기 직전에 패키지 변수 `reviewDeliveryTicket{dir, stateKey, sessionID}` 를 채운다(단일 훅 프로세스 안에서 핸들러 → 실행기로 값을 넘기는 최소 장치 — 시그니처를 바꾸지 않는 대가이며 `@MX:NOTE` 로 이유를 적는다). 실행기는 요약을 stderr 에 **성공적으로 쓴 직후** `markReviewDelivered(ticket)` 를 부른다 — 기록을 읽어 `state_key`·`session_id` 가 같고 `delivered` 가 거짓이면 참으로 바꿔 `atomicfile.Replace` 로 쓴다. 락 밖의 읽고-고치는 쓰기이므로 같은 순간의 다른 Stop 의 기록과 갱신을 잃을 수 있고(덜 억제하는 방향 — 한 번 더 리뷰), 그 한계는 REQ-CRA-005 에 적었다.

핸들러 새 흐름(기존 번호 유지, 삽입 지점만 명시):

1. 비활성 → `stop_hook_active` → 스코프 해상·로그 → *(형제: `tree_scope: skip` 판정 — 락 이전)* → 스코프 셀프게이트 → codex 조회. (불변)
2. **[삽입] 트리 정규화·git 디렉터리** 상태 계산 예산(60s, 이 단계와 4 단계가 공유) 안에서 `auditreceipt.TreeRootFromCWD(scope.Dir)`(`treeroot.go:22`) → `git -C <root> rev-parse --absolute-git-dir` → `<gitdir>/moai/` 생성. 실패·시간 초과면 락·상태 없이 진행한다(REQ-CRA-013 행은 쓸 곳이 없으면 stderr 진단으로 대신).
3. **[삽입] 리뷰 락** `AcquireReviewLock(<gitdir>/moai/codex-review.lock)`. 경합이면 REQ-CRA-013 행(`lock-held`)을 쓰고 allow. 기계 오류면 행(`lock-unavailable`)을 쓰고 락 없이 진행(fail-open). 성공이면 `defer Release`.
4. **[삽입] 시작 상태** `codexReviewReceiptStateForScope`(`codex_review_receipt.go:86`)로 `S0` — **다섯 필드**(`Head`·`TreeDigest`·`ConfigDigest`·`Command`·`ToolVersion`)를 `\x00` 로 이은 문자열이 상태 키 `S0.key` 다. 오류·시간 초과면 `state-unavailable` 행을 쓰고 상태 기반 단계(5, 8, 9)를 건너뛰고 진행(리뷰는 돈다).
5. **[삽입] 알려진 상태 건너뜀** 기록을 읽어(파싱 실패 = 기록 없음) `record.state_key == S0.key` 이고 다음 중 하나이면 행(`state-unchanged`)을 쓰고 allow: (가) `verdict == pass`; (나) `verdict == fail` 이고 `consecutive_fail >= cap`(전달할 것이 더는 없다); (다) `verdict == fail` 이고 `delivered` 이고 `session_id == 이 세션` 이고 `recorded_at` 이 전달 창(30분) 안. 그 밖 — 미전달, 다른 세션, 창 경과 — 은 리뷰한다.
6. 리뷰 실행. (불변: `runCodexReviewRPC`, 900s)
7. verdict 분기. (불변: `isBlockVerdict`)
8. **[삽입] 종료 상태** 새 상태 계산 예산(60s)으로 `S1` 재계산, `stale = S1.key != S0.key`.
9. **[삽입] 기록·상한** (상태 키는 **리뷰한 상태 `S0.key`** 로 기록한다 — 아직 리뷰하지 않은 `S1` 이 "알려진 실패 상태"가 되지 않게.)
   - pass → 기록 `{state_key:S0.key, verdict:pass, consecutive_fail:0, session_id, delivered:false, recorded_at}`; allow.
   - inconclusive → 기록하지 않음; allow.
   - fail, `기록.consecutive_fail >= cap` → 기록 `{state_key:S0.key, verdict:fail, consecutive_fail:cap(그대로), …}` 를 갱신하고 행(`failure-cap`) 후 allow(침묵). 리뷰는 이미 돌았다.
   - fail, 그 밖 → `streak = 기록.consecutive_fail + 1`(기록 없음/pass 이후 0); 기록 `{state_key:S0.key, verdict:fail, consecutive_fail:streak, session_id, delivered:false, recorded_at}` 를 쓰고 요약(§B.2)을 구성해 block 반환 + ticket 설정. `streak == cap` 이면 요약 끝에 마지막 알림(§B.2)을 붙인다.

보유 상한: 락을 쥔 채 도는 단계는 2~9 이고 각각 유한하다 — 리뷰어 900s, 상태 계산 묶음 60s 둘. 상한 = 900s + 120s = 1020s(REQ-CRA-004). 상태 계산 seam 이 막히는 시험이 이 상한을 고정한다(AC-004).

### B.2 요약 문구 (REQ-CRA-001, 006, 015)와 크기 상한

```text
codex review gate: FAIL
card: <워크트리 디렉터리 이름 — WT- 카드 워크트리일 때만, 아니면 공란>   branch: <브랜치>
reviewed HEAD: <S0.Head>
[stale 일 때]
STALE: this is a review of an earlier state (reviewed HEAD <S0.Head>, current HEAD <S1.Head>)
summary: <ReviewOutput.Summary, 최대 1500자>
findings (<n>):
  - [<severity>] <file>:<line> <title, 줄당 최대 300자>
  … (최대 10건, 초과분은 "(+k more)")
[streak == cap(3) 일 때 — 마지막 알림]
FINAL NOTICE: this is the third consecutive failure notification for this tree; further failures are suppressed until a review passes. No further notifications will follow; the state is UNRESOLVED.
이후 알림 없음, 상태는 미해결
```

- **첫 줄 `codex review gate: FAIL` 은 sentinel 이다**(REQ-CRA-001·003). 정확히 이 한 줄이어야 하며 요약 블록의 첫 줄이다. 운영 중 프로세스 stderr 에서는 그 앞에 스코프 로그 JSON 행이 있다 — 래퍼가 sentinel 줄부터 추려 알림에 싣는다.
- 크기 상한(`internal/config` 단일 원천): findings 10건·줄 300자·`summary` 1500자·요약 전체 8000바이트. 전체가 넘으면 findings, 다음에 summary 를 줄이고 `[truncated]` 를 붙인다. sentinel 줄·`STALE` 줄·마지막 알림 두 줄은 **먼저 자리를 잡고 자르지 않는다**. 이 숫자는 보수적 선택이며 훅 출력 크기 상한 자체는 미확인이다(spec.md §A.2 Gap — 프로브 목록에 추가하지 않고 Gap 으로 둔다).
- "카드" 칸(Q15 확정, Jev 1.00): 코드에는 카드 id 의 출처가 없다. 레인 규약상 워크트리 디렉터리 이름이 카드 id 다(`.claude/rules/local/gitflow-lane-protocol.md` §1) — `WT-` 카드 워크트리일 때(`cardScopeFromBranch(scope.Branch)`)만 그 이름을 적고, 카드 워크트리가 아닌 트리(리더가 앉은 primary 포함)에서는 공란이다. 디렉터리 이름은 `filepath.Base(scope.Dir 의 트리 루트)` 다.
- 마지막 알림의 고정 문구 `이후 알림 없음, 상태는 미해결` 은 시험이 정확히 일치를 보는 문자열이다(AC-015). 영어 줄은 같은 뜻의 병기다(요약 나머지가 영어인 알림에 한국어 한 줄이 섞이는 것은 의도된 것이다).

### B.3 락 — gate-run 락 원시 기능의 일반화 (재사용 우선)

`internal/cli/gate_lock.go` 의 `acquireGateLockImpl(lockPath)`(유닉스 flock `LOCK_EX|LOCK_NB`, 보유자 PID 기록, 커널이 종료 시 해제 — `gate_lock_unix.go:38-61`)는 **이미 경로 매개변수형**이다. 고칠 곳은 이름·경로가 상수에 묶인 윗층뿐이다: `AcquireGateLock`·`ReadGateLockHolder`·`ClearStaleGateLock` 이 `gateLockPath(projectDir)` 고정 경로를 쓴다(`gate_lock.go`, 윈도우 `gate_lock_windows.go:129`). 일반화: 경로를 받는 헬퍼(`acquireLockAt(lockPath)` 류)를 두고 기존 gate-run 락 함수들은 그것을 `<트리>/.moai/state/gate-run.lock` 경로로 부르게 한다(동작 불변 — 기존 `gate_lock` 시험 초록이 회귀 칸). 새 락은 `<gitdir>/moai/codex-review.lock`. 죽은 보유자 인수: 유닉스는 flock 이 자동 해제(시험: 죽은 PID 를 적은 낡은 파일 위에서 획득 성공), 윈도우는 `ClearStaleGateLock` 의 경로 매개변수형(`kanban.FactoryProcessAlive` 확인+changed-hands 중단).

**락·기록·로그가 git 디렉터리에 있는 이유**(spec.md §A.5): 작업 트리 아래에 두면 `.gitignore` 가 숨기는 저장소(이 저장소)에서만 `verify.Key` 가 안정하고, 숨기지 않는 저장소에서는 락 파일(내용 = 보유자 PID, 실행마다 다름)·기록·로그가 비추적 내용으로 키에 섞여 알려진 상태 건너뜀이 발화하지 않고 Codex 경로의 키까지 바뀐다. `verify.Key` 를 고쳐 이 세 경로를 제외하는 안은 Codex 경로의 키 계산을 건드리므로(REQ-CRA-009) 택하지 않았다.

선택하지 않은 안: `internal/lockfile`(보유자 신원 없음 → 죽은 보유자 판정 불가), `internal/verify/claim_lock.go`(5분 TTL — 최대 900s 리뷰에 맞지 않는 만료로 살아 있는 보유자를 인수해 중복 실행을 만든다).

### B.4 재전달 기록과 억제 로그

- 기록: `<gitdir>/moai/codex-review-wake.json` — `{state_key, verdict, consecutive_fail, session_id, delivered, recorded_at}`. 락 안에서 쓰고 `atomicfile.Replace`(`internal/atomicfile/replace.go:20`)로 임시 파일을 바꿔 쓴다(반쯤 쓴 파일을 다른 프로세스가 읽지 않게). `delivered` 표지만 락 밖에서 갱신된다(B.1). 읽기 실패·파싱 실패는 "기록 없음"으로 취급(fail-open, 한 번 더 리뷰가 돌 뿐). 수동 해제: 이 파일을 지우면 연속 횟수가 0 으로 돌아간다(`moai hook codex-review-gate --help` 에 적는다).
- **기록은 Codex 경로와 공유하지 않는다**(검증 영수증 저장소 불사용). 영수증을 공유하면 Claude 훅이 남긴 fail 이 Codex 멤버 6 의 차단 근거가 되어 Codex 경로 동작이 바뀐다(REQ-CRA-009). → Q13 해결: 공유하지 않음(Jev 0.02 로 약하나 SPEC 이 이미 가진 설계이고 리더가 수용; Codex 경로 불변).
- 억제 로그: `<gitdir>/moai/codex-review-gate.log` 에 JSON 한 줄(`ts`, `reason ∈ {lock-held, state-unchanged, failure-cap, lock-unavailable, state-unavailable, non-verdict-exit-2}`, `tree`, `state_key`(알면)). 래퍼도 같은 파일에 `non-verdict-exit-2` 행을 덧붙인다(셸 `printf`).

### B.5 래퍼 (REQ-CRA-003) — 판정 실패에서만 깨운다

```text
STDERR_FILE=$(mktemp)            # 실패하면 캡처 없이 종전처럼 실행하고 exit 0
printf '%s' "$INPUT" | "$MOAI_BIN" hook codex-review-gate 2>"$STDERR_FILE"
rc=$?                            # 파이프라인의 마지막 명령(moai)의 상태
if [ "$rc" -eq 2 ] && grep -qx 'codex review gate: FAIL' "$STDERR_FILE"; then
  sed -n '/^codex review gate: FAIL$/,$p' "$STDERR_FILE" >&2     # sentinel 줄부터만
  rm -f "$STDERR_FILE"; exit 2
fi
[ "$rc" -eq 2 ] && <git 디렉터리의 gate 로그에 non-verdict-exit-2 행을 덧붙임>   # 패닉·시스템 오류: 조용히
[ "$rc" -ne 2 ] && cat "$STDERR_FILE" >&2      # 종료 코드 0/1/127 의 진단은 종전처럼 통과
rm -f "$STDERR_FILE"; exit 0
```

순수 셸 off 스위치(`:35-60`)와 바이너리 부재 시 `exit 0`(`:78-81`)은 그대로다. sentinel 이 없는 종료 코드 2(Go 패닉, `os.Exit(2)` 시스템 오류, usage 오류)는 stderr 를 알림에 싣지 않고 로그 행만 남긴다 — 고루틴 트레이스가 세션을 깨우지 않는다. 두 사본(C1 `.claude/hooks/moai/…`, C2 `internal/template/templates/.claude/hooks/moai/…`)을 같은 변경에 고친다. C2 본문은 중립(SPEC ID·카드 번호 없음). 등록된 명령 형태(`bash -c '[ -f "$0" ] && exec bash "$0"; …'`)는 바꾸지 않는다 — `exec` 가 종료 상태를 전달한다. 시험은 이 등록 문자열을 설정 파일에서 읽어 그대로 실행한다(AC-003).

### B.6 등록과 타임아웃 정책 (REQ-CRA-007, 008)

- 등록: `.claude/settings.json:190-194` 항목과 `settings.json.tmpl:191` 항목에 `"asyncRewake": true` 추가. multi 래퍼 항목은 손대지 않는다. `internal/template/hook_entries.go` 의 `HookEntry`(:17-30)에 `AsyncRewake bool` 을 더하고 `settingsDoc` 파싱(:53-56)·`String()`·정렬 비교(`sortHookEntries` 의 tie-break, :197)를 맞춘다. `doctor_hook_wiring.go` 의 비교 필드 문구에 `asyncRewake` 를 넣는다.
- 시험 고정: `review_gate_registration_test.go` 의 `reviewGateWrappers`(두 래퍼의 `timeout` 900 고정)를 래퍼별 `{timeout, asyncRewake}` 로 확장한다.
- **타임아웃 정책 재검토.** 두 층이 있다: 훅 `timeout`(Claude Code 가 적용) 과 핸들러의 총 상한(리뷰 900s + 상태 계산 2×60s = 1020s; 프로세스가 스스로 끊고 inconclusive → 종료 코드 0). 공식 표는 `timeout` 을 `async: true` 에 대해서만 "enforce 하지 않는다"고 적는다. `asyncRewake` 는 E-2 미관측이다. 결정표(프로브 뒤 확정):

| E-2 결과 | 훅 `timeout` 의 역할 | 정책 |
|---|---|---|
| 강제하지 않음 | 무의미 | Go 쪽 상한이 유일한 상한. 등록 값은 1020 + 마진(공통 규칙 유지, 시험 단순) |
| 강제(취소+출력 폐기) | 늦은 실패를 삼킬 위험 | 등록 값 > 1020 + 마진 → 프로세스 내부 fail-open(조용한 종료 코드 0)이 항상 먼저 발화 |
| 강제하되 종료 코드 2 는 전달 | 중간 | 위와 동일 규칙으로 안전 쪽 |

마진 값은 프로브 뒤 정한다(제안 시작점 60s → 등록 1080 — **미확정**, O-B). 현재 900/900 은 동점이라 취소와 자체 상한 중 먼저 발화하는 쪽이 우연이다.

### B.7 깨움 루프 방지 (hazard i)

REQ-CRA-005·014 의 두 경계: (1) 같은 상태 키(다섯 필드)+기록된 pass, 또는 전달 확인된 같은 세션·창 안의 fail → 리뷰도 깨움도 없음(현행 변경, Q16); (2) 트리의 **연속 실패 깨움 횟수** 3 — 세 번째가 마지막 알림(고정 문구 `이후 알림 없음, 상태는 미해결`)이고 그 뒤 모든 fail 은 침묵(행은 기록), pass 만 0 으로 되돌린다. 지문을 쓰지 않는다(재서술·0건 fail 충돌, spec.md §A.5). 침묵이 "통과"로 읽히는 위험을 침묵 대신 마지막 알림 하나로 막는다(O-A 재정). `stop_hook_active` 는 비동기 훅에서 의미가 있는지 미관측(E-3/E-4)이라 그것에 의존하지 않고, 핸들러의 기존 `stop_hook_active` 검사는 그대로 둔다. 오케스트레이터 문구의 "동일한 실패는 reviewed state 가 바뀔 때까지 억제"는 이 규칙으로 읽었다(§G 확인 사항).

---

## §C 재사용 지점

| 필요 | 재사용 | 새로 쓰는 것 |
|---|---|---|
| 트리별 락·보유자 신원·죽은 보유자 | `acquireGateLockImpl`(경로형), `GateLockOwner`, `ErrGateLockHeld`, `kanban.FactoryProcessAlive` | 경로 매개변수 헬퍼(≈25 LOC), `AcquireReviewLock` |
| 트리 정규화 | `auditreceipt.TreeRootFromCWD` | git 디렉터리 해상 한 줄 |
| 리뷰 상태 키 | `codexReviewReceiptStateForScope`(다섯 필드) | 키 문자열 결합 한 줄 |
| 임시 파일 교체 | `atomicfile.Replace` | 기록 구조체(≈10 LOC)·`markReviewDelivered`(≈15 LOC) |
| 종료 코드 | `exitCodeError`, ExitCoder | 실행기 매핑(≈15 LOC) |
| 리뷰 실행 | `runCodexReviewRPC`(불변) | 없음 |
| 와이어링 비교 | `HookEntry`/`DiffHookEntries` | 필드 하나 |

---

## §D 프로브 계획 — E-1~E-5 (M1, 이후 마일스톤을 게이트)

**절차 문서:** `.moai/reports/t1422/async-probes.md` (로컬 gitignore 증거, 한국어, 명령 원문). 프로브마다 목적과 답하는 E 항목, 준비(일회용 `mktemp -d` 프로젝트 — 이 저장소의 추적 설정과 primary 체크아웃은 절대 편집하지 않음), 명령, 가설별(깨움/비깨움) 기대 관측, E 항목 판정 기준, 기록 위치, 정리를 적었다. **키트:** `.moai/reports/t1422/async-probe/kit/`(타임스탬프 줄을 덧붙이는 훅 스크립트와 설정 조각). **기록:** 프로브당 `.moai/reports/t1422/async-probe/<프로브 ID>.txt` — 결과 칸은 비어 있고 실행한 쪽이 채운다. `.moai/reports/*` 는 gitignore 대상이므로 각 파일의 sha256 을 tracked `progress.md` 의「프로브 기록」표에 기록한다(AC-011).

**밀폐(hermetic) 실행(감사 D12).** 프로브가 띄우는 중첩 `claude` 가 레인의 환경(`MOAI_KANBAN*`·`MOAI_FACTORY_WORKER*`·`MOAI_PROFILE_LEASE_TOKEN`·`CLAUDE_PROJECT_DIR` 등)과 사용자 수준 설정·훅을 물려받으면 칸 레인처럼 행동하거나 프로필 임대를 두고 다투거나 사용자 훅이 프로브 기록을 오염한다. 키트의 `run-clean.sh` 가 `MOAI_*`·`CLAUDE_PROJECT_DIR`·`CLAUDE_CODE_SESSION_ID`·`CLAUDE_CODE_ENTRYPOINT` 를 환경에서 지운 채 명령을 실행하고, `claude` 는 `--setting-sources project` 로 일회용 프로젝트의 설정만 읽으며(사용자 수준 훅 제외 — 플래그는 `claude --help` 로 존재 확인, 동작은 프로브 기록에 `--debug-file` 로 남길 수 있음), 플래그와 환경 목록을 기록 파일에 적는다. 인증 백엔드가 Anthropic 구독이어야 한다(`moai glm` 세션의 `ANTHROPIC_BASE_URL` 도 지운다).

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

**처리 규칙.** 각 프로브는 결과를 `확인됨/부정/불확정` 중 하나로 progress.md 에 적는다. E-1 은 `E1-interactive` 가 정한다(`E1-headless` 는 보조). 불확정(실행 불가)은 Gap 이고 그 항목을 게이트로 가진 마일스톤은 시작하지 않는다 — **리더가 Gap 을 수용하는 기록이 있으면 예외**(acceptance.md §E 도 같은 문장). 리더 전용 프로브(`E1-interactive`·`E3-c2-newturn`·`E3-c3-sessionexit`·`E5-reload`)는 리더 세션이 절차 문서대로 실행한다. 어떤 프로브도 관측을 미리 주장하지 않는다.

---

## §E 레인 안전 (hazard ii)

- 등록 변경(`.claude/settings.json`, `settings.json.tmpl`)은 마지막 run-phase 마일스톤(M5)에서만 한다. 그 이전 어떤 커밋도 두 파일을 건드리지 않는다(AC-011/012 가 커밋 그래프로 증명).
- 이 저장소의 레인 세션은 등록이 바뀌어도 스스로를 깨우지 않는다: (1) 래퍼의 순수 셸 off 스위치(`:35-60`)가 설정이 게이트를 켜지 않은 세션에서 래퍼를 즉시 끝낸다. 이 저장소의 **추적** `workflow.yaml` 에는 `review_gate` 키가 없다(`grep -c review_gate .moai/config/sections/workflow.yaml` → 0, 본 트리, 2026-10-02). 래퍼가 읽는 파일은 `${CLAUDE_PROJECT_DIR:-$PWD}/.moai/config/sections/workflow.yaml` 이라(래퍼 `:26-27`) 연결 워크트리 세션에서 `CLAUDE_PROJECT_DIR` 이 primary 를 가리키면 **primary 의 로컬 수정본**을 읽는다. 그 로컬 파일은 이동하는 운영자 상태다(2026-10-02 16:39 KST 관측: `codex.review_gate.enabled: false`, 수정 15:32) — 이 SPEC 은 그 값에 **의존하지 않는다**: 값이 `false` 면 게이트가 꺼져 래퍼가 즉시 끝나고, `true` 로 켜져 있어도 (2) 가 안전선이다. (2) 깨움은 종료 코드 2 + sentinel 줄에서만 일어나고 그것은 활성+실패 리뷰에서만 나온다. 진행 중 세션이 설정 편집을 즉시 반영하는지는 E-5 미관측이라 이 레인의 현재 세션이 영향받는지는 단정하지 않는다.
- 착지 뒤 primary 의 리더는 형제 SPEC 의 `tree_scope: skip`(리더 행위, 형제 인계 항목)로 제외되고, 그렇지 않은 활성 트리 세션만 이 SPEC 의 비동기 게이트를 받는다.

---

## §F 파일 목록·Template-First 미러·마일스톤 배정

### F.1 Go 프로덕션

| 파일 | 변경 | 마일스톤 |
|---|---|---|
| `internal/cli/gate_lock.go`, `gate_lock_unix.go`, `gate_lock_windows.go` | 경로 매개변수 일반화(동작 불변) | M2 |
| **신규** `internal/cli/codex_review_lock.go` | `AcquireReviewLock`, git 디렉터리 해상, 재전달 기록 읽기·쓰기·`markReviewDelivered`, 억제 로그 | M2 |
| `internal/cli/codex_review_gate.go` | 새 흐름 삽입(§B.1)·요약 구성(stale·마지막 알림 포함)·실행기 매핑·ticket | M2 |
| `internal/config/defaults.go` | 연속 실패 상한·전달 창·상태 계산 예산·findings/크기 상한 상수 | M2 (마진은 M4) |
| `internal/template/hook_entries.go`, `internal/cli/doctor_hook_wiring.go` | `AsyncRewake` 키·정렬·비교 문구 | M4 |
| `internal/cli/hook.go` | 명령 Short/Long 문구(신호 계약, 로그·기록 위치, 수동 해제) | M5 |

### F.2 테스트

| 파일 | 내용 | 마일스톤 |
|---|---|---|
| **신규** `internal/cli/codex_review_async_test.go` | 신호(운영 stderr 스트림)·침묵 결과·락 동시성(장벽·보유 시간)·알려진 상태(전달 확인)·상한·마지막 알림·stale·skip 순서·억제 로그·키 안정성 | M1(RED)→M2 |
| `internal/cli/gate_lock_*_test.go` | gate-run 락 회귀(일반화 뒤 초록) | M1(회귀)→M2 |
| `internal/hook/review_gate_selfgate_test.go` (또는 신규 `review_gate_wake_test.go`) | 래퍼 종료 코드 행렬·등록 명령 문자열 실행·패닉·sentinel | M1(RED)→M3 |
| `internal/cli/codex_review_gate_wiring_test.go`, `hook_e2e_test.go` | 신호 계약 변경에 맞춘 기대값 | M2 |
| `internal/cli/codex_stop_chain_golden_test.go`(`TestStopChainEffectParityGolden`) | Codex 멤버 6 의 효과 동치 — AC-009 회귀 집합에 포함(감사 D13) | M1(회귀) |
| `internal/hook/review_gate_directory_test.go` | 래퍼 종료 코드를 실패로 읽는 헬퍼가 있다 — 새 계약에 맞춰 확인(감사 D13) | M3 |
| `internal/template/hook_entries_test.go`, `review_gate_registration_test.go` | `asyncRewake`·래퍼별 timeout·정렬 | M4→M5 |

### F.3 미러 (C1=로컬, C2=배포)

| C1 | C2 (`internal/template/templates/…`) | 마일스톤 |
|---|---|---|
| `.claude/hooks/moai/handle-codex-review-gate.sh` | `.claude/hooks/moai/handle-codex-review-gate.sh` | M3 |
| `.claude/settings.json`(codex 항목) | `.claude/settings.json.tmpl`(codex 항목) | **M5(마지막)** |
| `CHANGELOG.md` | — | M6 (sync) |

`make build` 로 임베드 템플릿을 재생성한다. 에이전트 정의·규칙 파일은 건드리지 않으므로 `make agents-emit` 은 필요 없다. 새 `moai update` 가 사용자 `settings.json` 에 플래그를 어떻게 전달하는지(D17-i)는 이 SPEC 의 AC 밖이다 — 템플릿 렌더 결과가 `asyncRewake` 를 담는 것까지가 범위다(**Gap**).

---

## §G 위험·Open decisions·해결된 결정

### 위험

1. **E-1 부정(깨우지 않음).** 등록 요구 철회, 핸들러 개선은 동기 등록 아래에서도 유효(종료 코드 2 = 동기 Stop 에서 중단 방지, A.2). 이 경우 리뷰 대기 문제는 그대로이고 별도 대안이 필요하다.
2. **형제 SPEC 의 핸들러 변경.** 형제는 이 SPEC 보다 먼저 착지하고 시그니처를 바꾸지 않는다. 형제의 삽입 지점은 스코프 해상 직후(skip)이고 이 SPEC 의 삽입 지점은 codex 조회 뒤(트리 루트·락·상태 키)다 — 겹치지 않는다(교차 확인: 형제 plan.md §B.3). 형제가 착지한 뒤 소스를 다시 읽어 삽입 위치를 확정한다.
3. **t1399** — `MOAI_KANBAN*` 삭제는 `reviewGateEnvContext`(`codex_review_scope.go:328-334`)에 닿는다. 이 SPEC 은 env 를 읽지 않는다.
4. **중복 실행 부류(t1425).** 이 SPEC 은 codex 리뷰 게이트만 막는다. 관측 로그 폭주는 별개 카드. 건너뛰는 Stop 도 트리 루트·git 디렉터리·상태 키(git 하위 프로세스 4회+비추적 파일 읽기)·codex 버전 프로브를 락 앞뒤에서 치른다 — 락 전 값싼 점검(감사 D20)은 도입하지 않고 비용을 공시한다.
5. **명시 실행 경로는 락을 잡지 않는다.** `moai verify codex-review` 와 Claude 훅이 같은 트리에서 동시에 돌 수 있다(범위 밖; 감사 D21).
6. **오래된 Claude Code.** `asyncRewake` 필드를 모르는 버전의 동작은 미관측(Gap). 동기 등록처럼 돌면 종료 코드 2 는 중단 방지다.
7. **지속적 침묵의 오독.** 상한 도달 뒤 침묵이 "통과"로 읽힐 수 있다 — 그래서 상한은 침묵 대신 마지막 알림 하나(`이후 알림 없음, 상태는 미해결`)로 바꿨고(O-A 재정), 억제 로그가 보조 완화책이다. 상한에서 pass 가 오지 않는 장기 미해결 트리는 영구 침묵이므로 기록 파일 삭제가 수동 해제다.
8. **래퍼 종료 상태 포착.** 파이프라인의 종료 상태가 `moai` 의 것인지 셸 호환 범위에서 확인해야 한다(M3 시험 행렬이 고정). `mktemp` 가 없는 환경에서는 캡처 없이 종전처럼 exit 0 한다(깨움 없음 — 안전 쪽).
9. **스코프 로그 행의 stderr.** 종료 코드 0 훅의 stderr 가 Claude 에 보이면(E-4 d2) 조용하지 않다 — 그때는 로그 파일로 옮긴다(REQ-CGS-010 의 sink 변경이므로 별도 개정 필요). 종료 코드 2 알림에는 래퍼가 sentinel 줄부터 추리므로 실리지 않는다.
10. **delivered 표지 구간.** 실행기가 stderr 를 쓴 뒤 표지를 쓰기 전에 들어온 Stop 은 미전달로 읽혀 한 번 더 리뷰·전달할 수 있다(REQ-CRA-005에 명시) — 연속 실패 상한이 총량을 묶는다.
11. **락 보유자가 죽어도 codex 자식 프로세스가 남을 수 있다**(`exec.CommandContext` 는 직접 프로세스만 죽인다 — 감사 D18). E-3 c3 가 훅 프로세스 생존을 관측하지만 리뷰어 자식의 생존은 프로브하지 않는다 — Gap.
12. **상태 계산이 코드베이스에서 가장 비싼 부분이다.** 키 계산은 비추적 파일 전부를 읽는다(`verify.Key`). 거대한 비추적 디렉터리가 있는 트리에서는 60s 예산이 걸릴 수 있고, 그때 리뷰는 상태 없이 돈다(`state-unavailable`).

### 해결된 결정 (Jev `jev-1.13.0`, 운영자 위임, 2026-10-02 — 리더 수용 반영)

| 결정 | 처분 | 신뢰도 / 상태 |
|---|---|---|
| 중복 실행 방지 | 트리별 락, 보유 중이면 건너뜀, 죽은 보유자 인수 | 0.83 |
| 오래된 결과 | 표지를 붙여 전달(`label_stale`), 조용히 버리지도 표지 없이 전달하지도 않음 | 0.99 |
| 구조 | 형제 SPEC 분리(이 SPEC) | 0.93 |
| **O-A** (Q9) 연속 실패 상한 | Jev 선택 `cap3_silent` **0.21(약함)** → 오케스트레이터 재정 **`cap3_notify_once_at_cap`**: 트리당 연속 실패 깨움 3 번째가 마지막 알림 — 고정 문구 `이후 알림 없음, 상태는 미해결` — 이후 침묵. 재정 사유: 침묵이 "통과"로 오독되는 위험(위험 7). 리더 통지·재정 수용. 이 개정이 "동일한 실패"의 정의를 **버렸다**(지문 없음 — 연속 횟수만) | **잠정(PROVISIONAL)**, 리더 수용 |
| **O-B** (Q10) 타임아웃 마진 | **미결 — E-2 프로브 전까지 열려 있음.** 60s 는 시작값 제안이지 결정이 아니다 | OPEN |
| **O-C** (Q11) 오래됨의 정의 | `full_state_key` — 상태 키 전체의 어떤 차이든 오래됨, 표지가 두 HEAD 를 적는다. 운영자 문구("HEAD")의 확장 — 리더 수용 | 0.56 |
| **O-D** (Q12) 크기 | 수용(§H) | 0.98 |
| **O-E** (Q13) 영수증 공유 | `do_not_share` — Codex 경로 불변. Jev 신뢰도는 약하나(0.02) SPEC 의 기존 설계이고 리더 수용 | 0.02, **잠정**, 리더 수용 |
| **O-F** (Q14) 프로브 실행 | 레인이 헤드리스를 먼저 시도하고 대화형만 리더가 실행. 확정: 리더가 리더 세션에서 E-1 대화형·E-3(c2,c3)·E-5 를 실행하며 절차는 `.moai/reports/t1422/async-probes.md` | 0.60 |
| **O-G** (Q15) 카드 칸 | 워크트리 디렉터리 이름, `WT-` 카드 워크트리가 아니면 공란 | 1.00 |
| **O-H** (Q16) 알려진 상태 건너뜀(pass 포함) | `skip_known_state_for_pass` — 현행 동작 변경(오늘은 상태가 같아도 Stop 마다 재리뷰). SPEC 본문에만 명시하며 CHANGELOG 항목이 아니다(리더). 상태가 바뀐 Stop 의 리뷰는 회귀 칸(AC-005) | 0.41 (< 0.5), **잠정**, 리더 수용 |
| **O-I** (Q17) 스코프 로그 행 | stderr 유지, E-4 프로브 전까지(`E4-d2-exit0stderr` 가 정한다) | 0.97 |
| **O-J** (Q18) Tier | M 유지. 코드+테스트 ≈14(하한) / 미러·문서 ≈5 | 0.70 |

### 이번 개정(plan-audit 1회차 반응)에서 새로 표면화된 결정 — 확인이 필요하다 (decision-index Q19-Q23)

| ID | 결정 | 이 SPEC 의 선택(권고 아님 — 감사 지적에 대한 가장 작은 응답) | 상태 |
|---|---|---|---|
| Q19 | 상한의 계수 규칙 | 지문 폐기, 트리의 연속 실패 깨움 횟수만 센다. 해제는 pass 하나. 오케스트레이터 문구("동일한 실패는 reviewed state 가 바뀔 때까지")를 좁혀 읽었다 — 같은 상태 재리뷰는 REQ-CRA-005 가 이미 막아 문자 그대로는 상한이 무의미하고, 고정 문구 "이후 알림 없음"과도 모순되기 때문 | **확인 필요** |
| Q20 | 전달 확인 설계 | `session_id`+`delivered`(stderr 쓰기 성공 뒤 표지)+전달 창 30분. 대안(미채택): fail 에는 건너뜀을 적용하지 않음(연속 실패 상한만으로 묶기) | **확인 필요** |
| Q21 | 락·기록·로그 위치 | 작업 트리가 아니라 `<git 디렉터리>/moai/`. 대안(미채택): `.moai/state`·`.moai/logs` 유지+`verify.Key` 에서 제외(Codex 경로 키 변경) | **확인 필요** |
| Q22 | 래퍼의 깨움 판정 | 종료 코드 2 **그리고** sentinel 줄. 래퍼는 stderr 를 캡처해 sentinel 줄부터만 전달 | **확인 필요** |
| Q23 | 크기·시간 숫자 | findings 10·줄 300자·summary 1500자·전체 8000바이트·전달 창 30분·상태 계산 예산 60s×2. 모두 근거 없는 보수적 시작값이며 훅 출력 상한은 미확인 | **확인 필요** |

**확인이 필요한 읽기(O-A 와 Q19).** 오케스트레이터 재정 문구의 "동일한 실패는 reviewed state 가 바뀔 때까지 억제"를 이 SPEC 은 "연속 실패 깨움 3 번째 뒤 모든 실패는 pass 할 때까지 침묵"으로 읽었다. 같은 상태 재리뷰는 REQ-CRA-005 가 이미 막고 있어 문구를 문자 그대로 읽으면 상한이 아무것도 제한하지 못하고(상태가 바뀌면 곧 풀림), 마지막 알림의 "이후 알림 없음"과도 모순된다. 지문으로 "동일"을 정의하는 안은 감사가 지적한 두 결함(재서술 회피·0건 fail 충돌) 때문에 폐기했다. 그리고 "3 번째 깨움이 마지막 알림"으로 읽었다(4 번째 별도 알림 아님). 오케스트레이터가 다르게 읽는다면 REQ-CRA-014·015·AC-014·015·§B.1 step 9 가 바뀐다.

---

## §H 규모 추정과 3배 점검

(전부 **추정** — 이 레인은 구현하지 않았다. LOC 는 비테스트 Go 기준.)

| 구성 | 추정 |
|---|---|
| 락 일반화(유닉스·윈도우·윗층) | ≈ 25 LOC |
| 리뷰 락·git 디렉터리·재전달 기록·전달 확인·억제 로그 | ≈ 115 LOC |
| 핸들러 흐름 삽입·요약 구성(stale·마지막 알림·크기 상한) | ≈ 80 LOC |
| 실행기 매핑 | ≈ 15 LOC |
| 상수·`HookEntry`·help 문구 | ≈ 25 LOC |
| 합계(Go) | **≈ 260 LOC** (지문 제거 −15, 전달 확인·시간 상한·크기 상한 +55 가 초안 ≈220 에서의 변화) |
| 래퍼 셸 | ≈ 35 LOC (C1+C2 두 사본) |
| 테스트 | ≈ 400-500 LOC |

**파일 수(코드+테스트와 미러·문서를 분리).** 코드+테스트 ≈ 14(프로덕션 8·신규 1, 테스트 행 8·신규 1-2; 하한), 미러·문서 ≈ 5(래퍼 C1+C2, settings C1+tmpl, CHANGELOG). 합 ≈ 19. 프로브 키트·기록은 gitignore 대상 로컬 증거라 세지 않는다.

**3배 점검 기준선.** (i) 가장 단순한 기준선 = `exit 0`+JSON 을 종료 코드 2+stderr 로 바꾸고 레지스트리에 플래그를 더하는 것(≈30 LOC): ≈260/30 ≈ **8.7배, 3배 초과** — 이 기준선은 운영자 요구 2(트리당 하나·재전달 억제)와 3(오래된 결과 표지)을 충족하지 못해 유효한 비교가 아니다 — 크기 수용으로 해결(O-D, Jev 0.98; 크기가 초안보다 커졌으므로 리더에게 재통지한다). (ii) 요구 2·3 을 만족시키는 최소 구현 ≈ 락(자체 구현 ≈ 90)+상태·표지(≈ 90)+신호(≈ 30) ≈ 210 + 감사가 요구한 안전 장치(전달 확인·보유 상한·sentinel 래퍼 ≈ 50) ≈ 260 → 제안 ≈260 은 ≈ **1.0배**; 락 원시 기능 재사용이 이미 그 안에 반영돼 있다.

---

## §I 마일스톤 (우선순위 순서, 시간 추정 없음)

### M1 — 프로브와 회귀선, RED (우선순위 High)

- §D 프로브(ID: `E1-headless`·`E1-interactive`·`E2-timeout`·`E3-c1-concurrent`·`E3-c2-newturn`·`E3-c3-sessionexit`·`E4-d1-json`·`E4-d2-exit0stderr`·`E5-reload`) 수행·기록. 키트와 절차 문서(`.moai/reports/t1422/async-probes.md`)는 이미 작성돼 있고(결과 칸 공란, 밀폐 실행 포함), 레인은 헤드리스 프로브만, 리더는 리더 전용 프로브를 리더 세션에서 실행한다. 결과를 progress.md 에 `확인됨/부정/불확정` 으로.
- 변경 전 초록 관측: 아래 §J 회귀 칸 시험, gate-run 락 시험, `TestStopChainEffectParityGolden`.
- RED: 신규 AC 시험을 구현 없이 추가해 `-v` 로 `=== RUN`·`--- FAIL` 을 관측하고 `.moai/reports/t1422/red-async/` 에 구현 전 트리 SHA 와 함께 보존. 시험 하네스 항목: 운영 stderr 캡처(`os.Stderr` 파이프+`cmd.ErrOrStderr`), 등록 명령 문자열 실행기, 시계·세션 id·상태 계산·전달 쓰기 seam.
- M1 종료: 프로브 결과가 이후 게이트를 연다(§D). E-1 부정이면 M5 철회 기록.

### M2 — 핸들러·락·재전달·stale (우선순위 High) — 게이트: E-3, E-4

- 락 일반화(회귀 초록 유지) → `AcquireReviewLock`/git 디렉터리 → 흐름 삽입 → 요약 구성 → 실행기 매핑·ticket. REQ-CRA-001, 002, 004, 005, 006, 010, 013, 014, 015.
- M2 종료: `./internal/cli` 소관 시험 초록(락 동시성·보유 상한·알려진 상태·상한·마지막 알림·stale·skip 순서·키 안정성).

### M3 — 래퍼 (우선순위 High) — 게이트: E-4

- 두 사본 동시 수정, 종료 코드 행렬·패닉·등록 명령 문자열 시험. REQ-CRA-003.

### M4 — 와이어링 비교와 타임아웃 정책 (우선순위 Medium) — 게이트: E-2

- `HookEntry` 에 `AsyncRewake`(정렬·doctor 문구 포함), 시험 확장, 마진 상수. 설정 파일은 **아직 건드리지 않는다.** REQ-CRA-007(비교 부분), 008(값 결정).

### M5 — 등록 (우선순위 Medium, **마지막 run-phase 변경**) — 게이트: E-1(확인됨), E-5

- `.claude/settings.json` + `settings.json.tmpl` 에 `asyncRewake`·타임아웃 반영, `hook.go` 문구, 등록 시험. 직전 커밋에 프로브 산출물 sha256 기록이 있어야 한다. REQ-CRA-007, 008, 011, 012.

### M6 — 문서 (우선순위 Low, sync 단계)

- CHANGELOG `[Unreleased]` (알려진 상태 건너뜀의 pass 확장은 SPEC 본문에만 기술하고 CHANGELOG 항목으로 올리지 않는다 — 리더 지시). 인계 항목은 완료 보고에 옮긴다.

---

## §J 자기 검증

| 항목 | 명령 |
|---|---|
| 대상 테스트 | M1 에서 이름을 확정한다. `go test ./internal/cli/ -list 'ReviewAsync\|ReviewLock\|ReviewGate\|GateLock'` 로 열거한 정확한 이름들을 `-run '^(<이름1>\|<이름2>\|…)$' -count=1 -v` 로 돌려 `=== RUN` 수를 열거 수와 대조 |
| 래퍼·등록 | `go test ./internal/hook/ -list 'ReviewGateWrappers\|ReviewGateWake'`, `go test ./internal/template/ -list 'ReviewGates\|HookEntries'` 를 같은 정확-이름 형태로 |
| 회귀 칸(변경 전·후 초록) | `TestReviewGate_CodexFailBlocks`·`…CodexPassAllows`·`…FailOpenOnMissingCodex`·`…FailOpenOnCodexError`·`…InconclusiveAllows`, `TestReviewGatesRegisteredInRepoSettings`·`…InTemplateSettings`, `TestReviewGateWrappers_DisabledCostsZeroColdStarts`·`…EnabledReachesBinary`, `TestStopChainEffectParityGolden`, `TestGateLock_*` |
| 정적 | `go vet ./internal/cli/... ./internal/hook/... ./internal/template/... ./internal/config/...`, `GOOS=windows GOARCH=amd64 go build ./...` |
| 환경 세척 | `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test …` (한 번의 복합 호출) |
| 범위 침범 | 변경 파일에 `codex_stop_chain.go`, `codex_review_receipt.go`, `multi_review_gate.go`, `internal/verify/` 가 없어야 한다 |

전체 스위트는 로컬에서 돌리지 않는다.

## §K 안티패턴

1. 프로브 전에 "깨운다"를 사실로 쓰거나 등록을 먼저 바꾸기.
2. 락을 세션 단위나 cwd 단위로 잡기(같은 트리의 여러 세션·하위 디렉터리 Stop 이 리뷰를 겹친다) 또는 리뷰 호출 뒤에 잡기, 기록 쓰기 전에 풀기.
3. 경합 시 기다리기(훅 프로세스가 쌓인다) — 건너뛰어야 한다.
4. 5분 TTL 락(`verify/claim_lock.go`) 재사용 — 살아 있는 900s 보유자를 인수한다.
5. 오래된 결과를 조용히 버리거나 표지 없이 전달하기. 또는 HEAD 만(또는 digest 만) 비교하기. 또는 stale 실패를 현재 상태 키로 기록하기.
6. 옛 신호(stdout JSON block)를 병행 유지하기 — E-4 미관측이라 이중 신호의 의미를 모른다.
7. 래퍼가 비0 종료(또는 sentinel 없는 2)를 깨움으로 전달하기.
8. 억제를 기록하지 않기(조용한 게이트와 죽은 게이트를 가를 수 없다).
9. 재전달 기록을 Codex 영수증 저장소에 쓰기. 락·기록·로그를 작업 트리에 두기.
10. 이 SPEC 에서 형제의 `tree_scope`·자기 리뷰 도구 결정을 다시 건드리기.
11. findings 문자열로 "동일한 실패"를 판정하기(LLM 산문 — 재서술·0건 충돌).
12. 기록을 전달 전에 "전달됨"으로 쓰기.

## §L 참조

- spec.md §A(측정)·§B(요구)·§E(범위 밖)
- acceptance.md — AC·RED/GREEN 두 칸·변이 점검
- decision-index.md — 결정 권위 표기와 EVIDENCE-NEEDED 행
- progress.md — Decision Log·프로브 기록 칸·변이 점검 결과
- `SPEC-CODEX-REVIEW-OWNERSHIP-001`(형제) plan.md §B.2·§B.3·§G — skip 순서, 핸들러 시그니처 불변
- `.moai/reports/t1422/plan-audit.md`·`async-plan-audit.md` — 형제·이 SPEC plan-audit 보고서
- `.claude/rules/moai/development/verification-completeness.md` §1.3(연속 발화)·§2(두 칸 채택)
