---
id: SPEC-CODEX-REVIEW-ASYNC-001
title: "수락 기준 — Claude Stop 훅 codex 리뷰 게이트의 asyncRewake 전환"
version: "0.1.0"
created: 2026-10-02
---

# SPEC-CODEX-REVIEW-ASYNC-001 — 수락 기준

**트리 핀(문서 수준, 자체 핀이 없는 모든 기준에 적용): `3ae43ed8e78ffa673ca238227df6ca7202c1ce70`** — 본 트리 `git rev-parse HEAD` 출력(2026-10-02). 아래 장부의 모든 값은 이 핀에서 이 레인이 실행해 얻었다.

**개수 규칙.** 수락 기준 = `### AC-` 제목 수 = **13**(하위 ID 없음). 요구 = `### REQ-` 제목 수 = 13. 둘 다 Tier M 상한 16 이하.

## §A 관측 규율 [HARD]

AC 는 **종료 코드와 stdout/stderr 내용**, **리뷰어 호출 횟수**(주입 seam: `codexSession`·`codexLookPath`·리뷰 RPC 스텁), **락·재전달 상태 파일·억제 로그의 존재와 내용**, **설정 파일의 키·값**을 관측한다. verdict 값 단독은 근거가 못 된다. 동시성 AC 는 장벽으로 겹침을 **강제**한다. 라이브 Claude Code·라이브 codex 를 요구하는 AC 는 없다 — 프로브는 기록되는 관측이다(AC-011).

## §B 픽스처

- **트리 픽스처**: 실제 git 저장소(리뷰 대상 변경 포함). 변형: `develop` 브랜치 트리(비카드), `WT-fixture-card` 카드 워크트리.
- **리뷰어 스텁**: `pass`·`fail`(findings 3개 이상)·`inconclusive`(오류 포함)를 돌려주고, 호출 횟수를 세며, 호출 중 콜백(트리 변경, 장벽 대기)을 실행할 수 있다.
- **설정 픽스처**: `enabled: true` + (선택) `tree_scope: skip`, 상태 파일·락 파일 위치는 `<트리>/.moai/state/`.
- **상태 A,B,C,D,E…**: 서로 다른 파일 편집으로 만든 서로 다른 상태 키.

## §C RED/GREEN 두 칸 규율 [HARD]

`verification-completeness.md` §2 를 따른다. 기준마다 **RED-now 칸**(구현 전 트리의 실패와 그 이유)과 **GREEN 경로 칸**(어느 마일스톤이 뒤집는가)을 짝으로 갖는다. 이미 초록인 기준은 RED 를 가장하지 않고 **회귀 칸(preserve)**으로 분류한다.

- **소스 수준 RED-now** — 아래 증거 장부(읽기 전용 단일 호출, stdout 원문, 종료 코드, 문서 수준 트리 핀).
- **행동 수준 RED-now** — 시험이 아직 없어 plan-time 에 실행할 수 없다. **M1 에서 시험을 먼저 추가해 `-v` 로 `=== RUN`·`--- FAIL` 을 관측하고 `.moai/reports/t1422/red-async/` 에 구현 전 트리 SHA 와 함께 보존할 때까지 릴리스 차단 자격이 없다.** RED 는 §D "RED 이유"가 가리키는 사유로 빨간 것이어야 하며 다른 사유(픽스처 오류)의 RED 는 무효다.
- **사후 조건(post-condition) 기준**: 커밋 순서처럼 구현 뒤에만 관측되는 항은 RED-now 칸을 가질 수 없다. 그 항은 `verification-completeness.md` §2.1 의 미결정 처분에 따라 **릴리스 차단이 아닌 회귀 가드**로 분류하고 통과로 기록하지 않는다(AC-011, AC-012 의 해당 항).
- **회귀 칸**: AC-002 의 종료 코드 항·AC-003 의 off 스위치 항·AC-009 전체·AC-012 의 off 스위치 항.

### 증거 장부 (트리 핀 `3ae43ed8e78ffa673ca238227df6ca7202c1ce70`)

```text
[L1] command: grep -c asyncRewake .claude/settings.json
     stdout:  0
     exit:    1
     red because: 저장소 설정에 asyncRewake 키가 없다 — AC-007
[L2] command: grep -c asyncRewake internal/template/templates/.claude/settings.json.tmpl
     stdout:  0
     exit:    1
     red because: 템플릿에 asyncRewake 키가 없다 — AC-007
[L3] command: grep -c 'exit \$' internal/template/templates/.claude/hooks/moai/handle-codex-review-gate.sh
     stdout:  0
     exit:    1
     red because: 래퍼가 핸들러 종료 코드를 전파하지 않는다 (마지막 줄이 무조건 exit 0) — AC-003
[L4] command: grep -c AsyncRewake internal/template/hook_entries.go
     stdout:  0
     exit:    1
     red because: 와이어링 비교 키에 asyncRewake 가 없다 — AC-007
[L5] command: grep -c exitCodeError internal/cli/codex_review_gate.go
     stdout:  0
     exit:    1
     red because: 실패를 종료 코드 2 로 옮기는 코드가 없다 — AC-001
[L6] command: grep -c 'AcquireGateLock\|reviewLock' internal/cli/codex_review_gate.go
     stdout:  0
     exit:    1
     red because: 핸들러에 락이 없어 Stop N 번이면 리뷰가 N 번 돈다(코드 판독) — AC-004, AC-010 의 양성 대조
[L7] command: grep -c -i stale internal/cli/codex_review_gate.go
     stdout:  0
     exit:    1
     red because: 오래된 결과 표지가 없다 — AC-006
[L8] command: grep -c -i wake internal/cli/codex_review_gate.go
     stdout:  0
     exit:    1
     red because: 재전달 억제·깨움 상한이 없다 — AC-005
[L9] command: grep -c emitHookOutput internal/cli/codex_review_gate.go
     stdout:  5
     exit:    0
     red because: 실행기가 allow 에도 stdout JSON 을 낸다(emitHookOutput 호출) — AC-002 의 "stdout 비어 있음" 항
[L10] command: ls .moai/reports/t1422/async-probe/E1-interactive.txt
     stdout:  (없음; stderr: ls: .moai/reports/t1422/async-probe/E1-interactive.txt: No such file or directory)
     exit:    1
     red because: E-1 을 결정하는 프로브 기록 파일이 없다(디렉터리 `async-probe/` 와 `kit/` 은 절차용으로 이미 있으므로 기록 파일 하나를 지목한다) — AC-011
[L11] command: grep -c 'codex-review-gate.log' internal/cli/codex_review_gate.go
     stdout:  0
     exit:    1
     red because: 억제 로그가 없다 — AC-013
[L12] command: grep -c 'FINAL NOTICE\|이후 알림 없음' internal/cli/codex_review_gate.go
     stdout:  0
     exit:    1
     red because: 상한 도달 시 마지막 알림 문구(`이후 알림 없음, 상태는 미해결`)를 내는 코드가 없다 — AC-005 (양성 대조는 P4)
[P1] command: grep -c '"timeout": 900' .claude/settings.json
     stdout:  2
     exit:    0
     red because (AC-008): 두 래퍼 모두 timeout 900 이고 Go 쪽 리뷰 상한도 900s(config.DefaultCodexReviewGateTimeout, defaults.go:557)라 "등록 값 > Go 상한"이 거짓이다
[P2] command: grep -c '"async": true' .claude/settings.json
     stdout:  11
     exit:    0
     positive control: L1 의 0 이 빈 스윕이 아님 — 같은 파일에 async 키는 있다
[P3] command: grep -c Async internal/template/hook_entries.go
     stdout:  5
     exit:    0
     positive control: L4 의 0 이 빈 스윕이 아님
[P4] command: grep -c -i detector internal/cli/codex_review_gate.go
     stdout:  4
     exit:    0
     positive control: L5-L8 의 0 이 빈 스윕이 아님 — 같은 파일은 검색 가능하다
[P5] command: grep -c 'codex-review.lock\|codex-review-wake' internal/cli/codex_stop_chain.go internal/cli/codex_review_receipt.go
     stdout:  internal/cli/codex_stop_chain.go:0
              internal/cli/codex_review_receipt.go:0
     exit:    1
     preserve: Codex 경로 파일은 새 락·상태를 참조하지 않는다 (AC-009) — 구현 뒤에도 0
[P6] command: grep -c reviewScopeResolver internal/cli/codex_stop_chain.go internal/cli/codex_review_receipt.go
     stdout:  internal/cli/codex_stop_chain.go:1
              internal/cli/codex_review_receipt.go:2
     exit:    0
     positive control: P5 의 0 이 빈 스윕이 아님
[R1] command: go test ./internal/cli/ -run '^(TestReviewGate_CodexFailBlocks|TestReviewGate_CodexPassAllows|TestReviewGate_FailOpenOnMissingCodex|TestReviewGate_FailOpenOnCodexError|TestReviewGate_InconclusiveAllows)$' -count=1 -v   (환경 세척)
     stdout:  --- PASS ×5, ok github.com/modu-ai/moai-adk/internal/cli 1.856s
     exit:    0
     preserve: 핸들러 판정 분기 (AC-001/002 의 판정 방향)
[R2] command: go test ./internal/template/ -run '^(TestReviewGatesRegisteredInRepoSettings|TestReviewGatesRegisteredInTemplateSettings)$' -count=1 -v
     stdout:  --- PASS ×2, ok github.com/modu-ai/moai-adk/internal/template 0.479s
     exit:    0
     preserve: 등록 시험 (AC-007 의 변경 전 초록)
[R3] command: go test ./internal/hook/ -run '^(TestReviewGateWrappers_DisabledCostsZeroColdStarts|TestReviewGateWrappers_EnabledReachesBinary)$' -count=1 -v
     stdout:  --- PASS ×2, ok github.com/modu-ai/moai-adk/internal/hook 1.384s
     exit:    0
     preserve: 래퍼 off 스위치·바이너리 도달 (AC-003, AC-012)
[R4] command: go test ./internal/cli/ -run '^(TestProduceCodexReviewReceipt_CardScopeRecordsCardState|TestCodexStopHandlerRunsTheChain)$' -count=1 -v
     stdout:  --- PASS ×2, ok github.com/modu-ai/moai-adk/internal/cli 3.446s
     exit:    0
     preserve: Codex 경로 (AC-009)
[R5] command: go test ./internal/cli/ -run '^(TestGateLock_AcquireRecordsHolderIdentityAndReleaseAllowsReacquire|TestGateLock_ContentionReturnsSentinel|TestGateLock_DeadHolderArtifactAcquiresFarBelowBudget|TestGateLock_ReadOnlyStateDirIsUnavailableNotFatal)$' -count=1 -v
     stdout:  --- PASS ×4, ok github.com/modu-ai/moai-adk/internal/cli 1.254s
     exit:    0
     preserve: gate-run 락 일반화 전 초록 (AC-004 의 회귀 항)
```

### 변이 점검(mutant probe) 요약 — 채택 전에 구성, pass/fail/inconclusive 변형 포함

| 기준 | 요구를 어기면서 기준을 만족하는 변이 | 이 변이를 죽이는 칸 |
|---|---|---|
| AC-001 | 요약 머리말만 쓰고 reviewed HEAD·findings 를 빼먹음 | 머리말 아래 줄별 내용 단정 |
| AC-001/002 | 옛 신호를 병행(stdout JSON block 도 출력) | stdout 비어 있음 단정(실패 변형) |
| AC-002 | `inconclusive`·오류를 종료 코드 2 로 보냄 | 변형별 종료 코드 0 단정 |
| AC-002 | 통과에도 "passed" 알림을 stderr 에 씀 | 요약 머리말 부재 단정(통과 변형) |
| AC-003 | 래퍼가 비0 종료를 그대로 전파(`exit $rc`) | 종료 1·127 → 래퍼 0 |
| AC-003 | 래퍼가 2 도 삼킴 | 종료 2 → 래퍼 2 + stderr 보존 |
| AC-004 | 락을 리뷰 호출 뒤에 잡음 | 장벽: 보유자가 리뷰어 안에서 기다리는 동안 나머지가 되돌아와야 함 |
| AC-004 | 경합 시 기다림(대기 후 리뷰) | 나머지 호출이 보유자 해제 전에 반환해야 함 |
| AC-004 | 락을 세션별로 키링 | 서로 다른 session_id 로 N 호출 → 여전히 1 회 |
| AC-004 | 해제 누락(오류 경로) | 직후 순차 Stop(다른 상태)이 리뷰를 다시 돌려야 함 |
| AC-005 | 상태 키를 HEAD 만으로 비교 | digest 만 다른 두 상태가 별개로 리뷰됨 |
| AC-005 | 카운터가 pass 에서 리셋되지 않음 | pass 뒤 fail 이 다시 전달됨 |
| AC-005 | 상한 도달 시 마지막 알림 없이 조용히 멈춤(`cap3_silent`) | 세 번째 깨움에 정확한 문구가 있어야 함 |
| AC-005 | 상한 도달 뒤에도 알림이 계속됨(네 번째 깨움) | 같은 지문의 D·E 가 조용해야 함 |
| AC-005 | 마지막 알림에 영어 줄만 있고 고정 한국어 문구가 없음 | 정확한 부분 문자열 `이후 알림 없음, 상태는 미해결` 단정 |
| AC-005 | 지문을 무시하고 상한 뒤 모든 실패를 침묵시킴 | 다른 지문의 실패가 상한 뒤에도 전달되고 연속 횟수가 1 로 다시 시작해야 함 |
| AC-005 | pass 에도 알려진 상태 건너뜀을 적용하지 않음(옛 동작) | 같은 상태 pass 재발생 시 리뷰어 호출 수 불변 |
| AC-005 | 알려진 상태 건너뜀이 상태 변경 뒤에도 적용됨 | 상태가 바뀐 Stop 은 리뷰어를 호출해야 함(회귀 칸 R1) |
| AC-005 | inconclusive 도 알려진 상태로 기록 | 같은 상태 inconclusive 재시도 시 리뷰어 재호출 |
| AC-006 | HEAD 만 비교해 stale 판정 | HEAD 같고 digest 만 변경된 변형 |
| AC-006 | 항상 표지를 붙임 | 변경 없는 변형은 표지 없음 |
| AC-006 | stale 실패를 조용히 버림 | 변형 (i)(ii)에서 stderr·종료 코드 2 존재 |
| AC-007 | multi 래퍼에도 asyncRewake | multi 항목 부재/false 단정 |
| AC-007 | `async` 로 잘못 씀 | `asyncRewake` 키 이름 단정 |
| AC-001 | 카드 칸을 항상 브랜치 이름으로 채움 / 비카드 트리에서도 채움 | 카드 워크트리는 디렉터리 이름, `develop`·primary 는 빈 칸 단정 |
| AC-011 | 프로브 파일만 있고 결과 토큰 없음 | 토큰·sha256 단정 |
| AC-011 | 프로브 ID 가 절차 문서와 다름 | 아홉 개 프로브 ID 파일 이름 단정 |

## §D AC 매트릭스

| AC | 요구 | 시작 상태 | RED 이유 / 회귀 근거 | 관측 대상 |
|---|---|---|---|---|
| AC-001 | REQ-CRA-001 | RED(행동) | 종료 코드 2 매핑 부재 (L5) | 실패 변형: 종료 코드·stderr 내용·stdout |
| AC-002 | REQ-CRA-002 | RED(stdout 항) + 회귀(종료 코드 항) | allow 에도 stdout JSON (L9); 종료 코드 0 은 R1 초록 | 조용한 결과 일곱 변형의 종료 코드·stdout·요약 부재 |
| AC-003 | REQ-CRA-003 | RED(전파 항) + 회귀(off 스위치) | 전파 부재 (L3); off 스위치 R3 초록 | 래퍼 종료 코드 행렬 |
| AC-004 | REQ-CRA-004 | RED(행동) + 회귀(gate-run 락) | 락 부재 (L6); R5 초록 | 장벽 동시성 N≥8 리뷰어 호출 수, 죽은 보유자 |
| AC-005 | REQ-CRA-005 | RED(행동) + 회귀(상태가 바뀐 Stop 은 리뷰) | 재전달 기록 부재 (L8), 상한 도달 문구 부재 (L12); 변경 상태 리뷰는 R1 초록 | 상태 시퀀스의 리뷰어 호출 수·종료 코드·정확한 상한 문구·지문 분기 |
| AC-006 | REQ-CRA-006 | RED(행동) | stale 표지 부재 (L7) | 트리 변경 변형의 요약 |
| AC-007 | REQ-CRA-007 | RED(소스) + 회귀(등록 시험) | L1, L2, L4; R2 초록 | 두 설정 파일의 키·와이어링 diff |
| AC-008 | REQ-CRA-008 | RED(소스) | 등록 900 = Go 상한 900 (P1) | 등록 timeout 대 Go 상한+마진 |
| AC-009 | REQ-CRA-009 | **회귀(초록)** | R4, P5 | Codex 경로 시험·정적 참조·영수증 저장소 불변 |
| AC-010 | REQ-CRA-010 | RED(양성 대조 항) | skip 이 아닌 세션이 락 파일을 만드는 대조 항이 오늘 빨강 (L6) | skip 세션의 락·상태·리뷰어 부재 |
| AC-011 | REQ-CRA-011 | RED(존재 항) + 사후 조건(순서 항) | 프로브 기록 부재 (L10) | 프로브 파일·sha256·결과 토큰·커밋 순서 |
| AC-012 | REQ-CRA-012 | **회귀(off 스위치)** + 사후 조건(순서) | R3 | off 스위치 모사·커밋 순서 |
| AC-013 | REQ-CRA-013 | RED(행동) | 억제 로그 부재 (L11) | 억제 사유별 로그 행 |

REQ→AC 누락 없음: 001→001 · 002→002 · 003→003 · 004→004 · 005→005 · 006→006 · 007→007 · 008→008 · 009→009 · 010→010 · 011→011 · 012→012 · 013→013.

---

### AC-001 — 실패는 종료 코드 2 + stderr 요약이고 stdout 은 비어 있다

**Given** 리뷰 대상 변경이 있는 트리(비카드 `develop` 과 `WT-` 카드 워크트리 각각), 리뷰어 스텁이 findings 3개 이상을 담은 `fail` 을 돌려주고, 시작 시점 HEAD 가 `H0`,
**When** `moai hook codex-review-gate` 실행기(cobra 명령, 메모리 입출력)가 Stop 입력을 처리하면,
**Then** 종료 상태는 코드 2 이고 stdout 은 비어 있으며 stderr 는 첫 줄이 요약 머리말이고(`Error:` 접두 줄 없음), 줄별로 카드 칸(`WT-` 카드 워크트리에서는 워크트리 디렉터리 이름, `develop`·primary 같은 비카드 트리에서는 빈 칸 — 브랜치 이름으로 대신 채우지 않는다), 브랜치, `reviewed HEAD: H0`, 판정 요약, `findings (<n>)` 과 각 finding 의 `[severity] file:line title` 을 담는다. 상한(plan.md §B.2)을 넘는 finding 은 "(+k more)" 로 접힌다. 같은 입력에서 stdout JSON `decision: block` 은 나오지 않는다.

### AC-002 — 그 밖의 결과는 종료 코드 0, 빈 stdout, 요약 없음 (일곱 변형 + 두 억제 변형)

**Given** 변형 — (a) 리뷰어 `pass` (b) `inconclusive` (c) 리뷰 호출 오류 (d) codex 바이너리 부재 (e) `enabled: false` (f) `stop_hook_active: true` (g) 리뷰할 변경 없음 (h) 락 경합으로 억제(AC-004) (i) 알려진 상태로 억제(AC-005),
**When** 실행기가 각 Stop 입력을 처리하면,
**Then** 아홉 변형 모두 종료 코드 0, stdout 비어 있음, stderr 에 요약 머리말(`codex review gate:` 로 시작하는 줄)이 없다. (stderr 에 기존 스코프 로그 행이 한 줄 남는 것은 허용 — E-4 d2 가 보이는지 확인한 뒤 위치를 정한다, plan.md O-I.)

### AC-003 — 래퍼는 종료 코드 2 와 stderr 만 전달하고 off 스위치를 유지한다

**Given** `moai` 를 가짜 실행 파일로 바꾼 PATH 와 활성 설정 픽스처. 가짜가 stderr `PROBE-ERR` 를 쓰고 (i) 종료 코드 2 (ii) 종료 코드 0 (iii) 종료 코드 1 (iv) 종료 코드 127 로 끝나는 네 변형. 비활성 설정 픽스처(`review_gate` 키 부재, `enabled: false`), `moai` 부재 변형,
**When** 래퍼 스크립트(C1 과 C2 사본 각각)를 실행하면,
**Then** (i) 래퍼 종료 코드 2 + stderr 에 `PROBE-ERR` 보존 (ii)(iii)(iv) 래퍼 종료 코드 0. 비활성 설정에서는 종료 코드 0 이고 가짜 `moai` 호출 수가 0(순수 셸 off 스위치 — 기존 `TestReviewGateWrappers_DisabledCostsZeroColdStarts` 계열 회귀), `moai` 부재에서는 종료 코드 0.

### AC-004 — 트리당 동시에 리뷰는 하나, 죽은 보유자는 막지 못한다

**Given** 같은 트리에 대한 N=8 개의 동시 Stop(서로 다른 `session_id` 포함), 호출 횟수를 세는 리뷰어 스텁이 첫 호출에서 나머지 N−1 호출이 모두 반환할 때까지 대기하고, 별도로 (a) 죽은 PID 를 적었으나 flock 은 잡혀 있지 않은 낡은 락 파일, (b) 락을 잡은 하위 프로세스를 SIGKILL 한 직후,
**When** 모든 호출이 핸들러(또는 실행기)를 지나면,
**Then** 리뷰어 호출 수는 정확히 1 이고 나머지 7 호출은 보유자가 아직 리뷰 중일 때 반환하며(기다림이 아닌 건너뜀) 종료 코드 0·빈 stdout 이다. (a)(b)에서 다음 Stop 은 락을 얻어 리뷰어를 1 회 호출한다. 보유자가 끝난 직후 다른 상태에서 오는 Stop 은 리뷰어를 다시 호출한다(해제 확인). 다중 프로세스 변형(N 개의 별도 프로세스)에서도 호출 수는 1 이다. **양성 대조**: 락을 끈 시험 설정에서는 호출 수가 N 이다(오늘의 동작이 그렇다는 M1 RED 관측). 락 일반화 뒤에도 gate-run 락 시험(R5)이 초록이다.

### AC-005 — 같은 상태는 다시 리뷰도 깨움도 없고 연속 실패는 상한에서 마지막 알림 한 번 뒤 침묵한다 (깨움 루프)

**Given** 연속 상한 K=3(트리당 연속 실패 깨움 횟수; 결정 O-A — 아래 PROVISIONAL 표기)과 **항상 실패하는 리뷰어 스텁**(상태마다 같은 findings 지문을 돌려준다)이 다음 시퀀스를 받는다 — 상태 A `fail` → 같은 A 에서 Stop 재발생 → 상태 B `fail` → 상태 C `fail`(세 번째 실패 깨움) → 상태 D `fail` → 상태 E `fail` → 상태 G `fail`(**다른 지문**의 findings) → 상태 H `pass` → 같은 H 에서 Stop → 상태 I `fail`; 별도로 상태 X `inconclusive` → 같은 X 에서 Stop; HEAD 가 같고 digest 만 다른 상태 쌍,
**When** 각 Stop 이 처리되면,
**Then**
- A 첫 Stop: 종료 코드 2(리뷰어 1 회). A 재발생: 종료 코드 0, 누적 리뷰어 호출 수 불변(알려진 상태 건너뜀 — 억제 로그 `state-unchanged`).
- B: 종료 코드 2(두 번째 깨움). **C: 종료 코드 2이고 stderr 에 정확한 부분 문자열 `이후 알림 없음, 상태는 미해결` 과 영어 `FINAL NOTICE` 줄이 정확히 한 번 있으며, 이후 같은 실패가 억제된다는 사실이 적혀 있다.** (이 세 번째 깨움이 마지막 알림 자체다 — 네 번째 알림은 없다.)
- D, E: 리뷰어는 호출되고(호출 수 +1씩) 종료 코드 0·빈 stdout·요약 머리말 없음(억제 로그 `failure-cap` 두 행). 상한 도달 뒤 **알림은 정확히 1회**였음을 전체 시퀀스에서 `이후 알림 없음, 상태는 미해결` 출현 수 1 로 단정한다.
- G(다른 지문): 종료 코드 2로 다시 전달되고 연속 횟수는 1 로 다시 시작한다(상한 뒤 침묵이 서로 다른 새 실패를 삼키지 않는다).
- H `pass`: 종료 코드 0 침묵, 연속 횟수 0 으로 초기화. H 재발생: 리뷰어 호출 수 불변(**pass 에도 알려진 상태 건너뜀 — 현행 동작의 변경, 결정 O-H**). I: 종료 코드 2(연속 횟수가 리셋돼 있었다).
- X `inconclusive` 재발생: 리뷰어가 다시 호출된다(inconclusive 는 알려진 상태가 아님). digest 만 다른 쌍은 별개 상태로 각각 리뷰된다.
- **회귀 칸(R1):** 상태가 바뀐 Stop 은 항상 리뷰어를 호출한다 — 알려진 상태 건너뜀은 상태 키가 같은 Stop 에만 적용된다. 이 항이 건너뜀을 pass 에 확장하는 변경의 안전 경계다.

**PROVISIONAL 표기.** O-A 의 상한 동작은 Jev(`cap3_silent`, 신뢰도 0.21 — 약함)와 달리 오케스트레이터가 `cap3_notify_once_at_cap` 으로 덮어쓴 것이며 리더가 수용했다. O-H(pass 에도 건너뜀)는 Jev 0.41 로 잠정이고 리더가 수용했다. 둘 다 이 AC 의 문구에 반영돼 있고, 결정이 뒤집히면 이 AC 의 해당 항만 바뀐다.

### AC-006 — 오래된 결과는 표지를 붙여 전달된다

**Given** 리뷰 중 콜백이 트리를 바꾸는 세 변형 — (i) 새 커밋(HEAD `H0`→`H1`) (ii) 추적 파일 편집(HEAD 는 `H0` 그대로, digest 만 변경) (iii) 변경 없음; 리뷰어 스텁은 각각 `fail`, 그리고 (i) 에 대해 `pass` 변형도,
**When** 실행기가 처리하면,
**Then** (i) fail: 종료 코드 2, stderr 에 `reviewed HEAD: H0` 와 `current HEAD: H1` 와 "review of an earlier state" 표지가 모두 있다. (ii) fail: 종료 코드 2, 두 HEAD 가 같은 값 `H0` 으로 적히고 표지가 있다. (iii) fail: 표지 없음, 종료 코드 2. (i) pass: 종료 코드 0 침묵이고 재전달 기록의 키는 시작 상태 `S0` 의 키다.

### AC-007 — 레지스트리는 codex 래퍼에만 asyncRewake 를 켜고 와이어링 비교가 그것을 본다

**Given** `.claude/settings.json` 과 `internal/template/templates/.claude/settings.json.tmpl`(렌더 후),
**When** Stop 훅 항목을 파싱하면,
**Then** codex 래퍼 항목은 `"asyncRewake": true` 이고 multi 래퍼 항목은 `asyncRewake` 가 없거나 false, 두 래퍼 모두 `async` 키는 없다(`async` 와 `asyncRewake` 혼동 변이 차단). `HookEntry` 에 `AsyncRewake` 가 있고 템플릿 렌더 결과와 저장소 설정의 `DiffHookEntries` 는 양방향 빈 결과이며, 항목에서 `asyncRewake` 를 뺀 변형에 대해서는 `templateOnly` 가 비어 있지 않다(와이어링 비교가 플래그 누락을 본다). 변경 전에는 `TestReviewGatesRegisteredInRepoSettings`·`…InTemplateSettings`(R2)가 초록이고 변경 뒤에도 래퍼별 기대값을 갱신해 초록이다. 이 기준은 E-1 이 확인됨일 때만 효력이 있다(AC-011).

### AC-008 — 등록 타임아웃은 Go 쪽 리뷰 상한보다 마진만큼 크다

**Given** 두 설정 파일과 `config.DefaultCodexReviewGateTimeout`, 마진 상수 `M`(> 0, `internal/config` 단일 원천),
**When** codex 래퍼 항목의 `timeout` 을 읽으면,
**Then** `timeout` > `int(DefaultCodexReviewGateTimeout.Seconds())` + `M`(현재는 900 = 900 이라 거짓 — P1). multi 래퍼의 `timeout` 은 900 그대로. Go 쪽 리뷰 상한 값은 900s 그대로다. `M` 의 값은 E-2 결과 뒤 확정된다(plan.md §B.6, O-B) — 값 확정 전에는 이 AC 를 통과로 기록하지 않는다.

### AC-009 — Codex 경로는 그대로이고 새 락·상태를 모른다 (회귀 칸)

**Given** 구현 뒤 트리,
**When** Codex 경로 시험과 정적 참조를 검사하면,
**Then** R4 의 시험(`TestProduceCodexReviewReceipt_CardScopeRecordsCardState`·`TestCodexStopHandlerRunsTheChain`)과 M1 이 열거해 고정한 Codex 멤버 6·영수증 시험이 전부 초록이다. `grep -c 'codex-review.lock\|codex-review-wake' internal/cli/codex_stop_chain.go internal/cli/codex_review_receipt.go` 는 두 파일 모두 0 이다(P5, 양성 대조 P6). Claude 훅을 픽스처 트리에서 실행해도 검증 영수증 저장소(`.moai/state/verify/`)의 목록은 호출 전후 동일하다.

### AC-010 — 형제의 skip 은 락·상태보다 앞서 일어난다

**Given** 형제 SPEC 이 착지한 트리 — `enabled: true`+`tree_scope: skip` 설정과 비 `WT-` 브랜치의 변경이 있는 트리에 8 개의 동시 Stop, 그리고 **양성 대조**로 같은 트리의 `tree_scope: review` 설정,
**When** 실행기가 처리하면,
**Then** skip 설정에서는 리뷰어 호출 0, `codex-review.lock`·`codex-review-wake.json`·억제 로그가 생성되지 않는다. 대조(`review` 설정)에서는 락 파일이 생성되고 리뷰어가 1 회 호출된다 — 이 대조 항은 오늘 락이 없어 빨강이다(L6).

### AC-011 — 프로브는 기록되고 결과가 등록을 게이트한다

**Given** M1 이후의 트리,
**When** `.moai/reports/t1422/async-probe/` 와 tracked `progress.md` 의「프로브 기록」표를 검사하면,
**Then** 절차 문서(`.moai/reports/t1422/async-probes.md`)가 정한 아홉 개 프로브 ID — `E1-headless`·`E1-interactive`·`E2-timeout`·`E3-c1-concurrent`·`E3-c2-newturn`·`E3-c3-sessionexit`·`E4-d1-json`·`E4-d2-exit0stderr`·`E5-reload` — 각각에 대해 `.moai/reports/t1422/async-probe/<프로브 ID>.txt` 가 있고 그 안에 `claude --version` 출력, 사용한 설정 조각 이름, 실행한 명령, 관측 원문, 결과 토큰 `확인됨`·`부정`·`불확정` 중 하나가 있으며, `progress.md`「프로브 기록」표에 각 파일의 sha256 이 적혀 있고 실제 파일 해시와 일치한다. E-1 은 `E1-interactive` 의 토큰이 정하고 `E1-headless` 는 보조 증거다. E-1 의 결과가 `확인됨` 이 아니면 두 설정 파일에 `asyncRewake` 가 없다. **순서 항(사후 조건, 회귀 가드):** 해시를 기록한 커밋은 `.claude/settings.json` 에 `asyncRewake` 를 도입한 커밋의 조상이다(커밋 그래프로 관측 — `verification-claim-integrity.md` §2.3). 실행 주체는 progress.md 표에 적는다 — 레인이 헤드리스로 먼저 돌리는 것은 `E1-headless`·`E2-timeout`·`E3-c1-concurrent`·`E4-d1-json`·`E4-d2-exit0stderr`, 리더만 돌리는 것은 `E1-interactive`·`E3-c2-newturn`·`E3-c3-sessionexit`·`E5-reload` 이며, 결과가 올 때까지 후자는 `불확정` 이다.

### AC-012 — 비활성 트리의 세션은 프로세스도 깨움도 만들지 않고 등록은 마지막이다

**Given** `asyncRewake` 가 등록된 설정과 `review_gate` 키가 없는 `workflow.yaml`(이 저장소의 추적 설정과 같은 형태 — `grep -c review_gate .moai/config/sections/workflow.yaml` → 0, 본 트리),
**When** 래퍼를 실행하면,
**Then** 종료 코드 0, 가짜 `moai` 호출 수 0, stdout·stderr 비어 있음(R3 계열 회귀). **순서 항(사후 조건, 회귀 가드):** 두 설정 파일을 바꾸는 커밋 뒤에는 M6 문서·progress 기록 외의 run-phase 커밋이 없다.

### AC-013 — 억제는 사유와 함께 기록된다

**Given** AC-004·AC-005 의 시퀀스,
**When** 각 억제(락 경합, 알려진 상태, 연속 실패 상한 뒤 같은 지문의 실패)가 일어나면,
**Then** `<트리>/.moai/logs/codex-review-gate.log` 에 억제 한 건당 JSON 한 줄이 추가되고 `reason` 은 각각 `lock-held`·`state-unchanged`·`failure-cap`, `tree` 와 `state_key` 가 채워져 있다(AC-005 시퀀스에서 `failure-cap` 은 정확히 두 행 — D, E). 상한에 닿은 세 번째 깨움(C)은 억제가 아니라 마지막 알림이므로 억제 행을 쓰지 않는다. 억제가 아닌 정상 리뷰(전달 포함)는 억제 행을 쓰지 않는다. 락 기계 오류(상태 디렉터리 쓰기 불가) 변형은 `lock-unavailable` 행을 쓰고 리뷰는 진행된다(fail-open).

---

## §E 품질 게이트와 DoD

- **DoD**: §D 전 AC 최종 상태 PASS(사후 조건 항은 회귀 가드로 별도 표기) · 행동 수준 RED 관측 기록이 `.moai/reports/t1422/red-async/` 에 구현 전 트리 SHA 와 함께 존재 · 프로브 E-1~E-5 기록과 sha256 (AC-011) · `go vet`(darwin)+`GOOS=windows GOARCH=amd64 go build ./...` 신규 0 · 수정 파일 커버리지 `quality.yaml` `test_coverage_target`(85) 이상 · 회귀 칸 R1-R5 가 변경 뒤에도 초록.
- **간접 검증**: AC-004 는 장벽으로 "기다림이 아닌 건너뜀"을, AC-005 는 상태 시퀀스로 "깨움 루프"를 양성 관측한다. 부재 주장(정적 0)은 양성 대조(P2-P4, P6)와 함께만 읽는다.
- **선행 폐쇄 게이트**: M1 회귀 칸(R1-R5)이 변경 전 트리에서 초록으로 관측되지 않으면 이후 마일스톤을 시작하지 않는다. 프로브 게이트(spec.md §D)를 건너뛰고 M2 이후를 시작하지 않는다.
- **측정 규율**: 소관 패키지 단위로 재측정한다. `-run` 이름 패턴 단독 재측정 금지(0개를 골라도 `ok`). 전체 스위트 로컬 금지.
