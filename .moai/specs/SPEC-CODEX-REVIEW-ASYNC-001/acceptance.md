---
id: SPEC-CODEX-REVIEW-ASYNC-001
title: "수락 기준 — Claude Stop 훅 codex 리뷰 게이트의 asyncRewake 전환"
version: "0.1.0"
created: 2026-10-02
---

# SPEC-CODEX-REVIEW-ASYNC-001 — 수락 기준

**트리 핀(문서 수준, 자체 핀이 없는 모든 기준에 적용): `a0d8014096f894c3a45577482b0eb92892716e04`** — 본 트리 `git rev-parse HEAD` 출력(2026-10-02). 아래 장부의 값은 이 핀에서 이 레인이 실행해 얻었다. 장부의 회귀 칸(R1-R5)만 최초 핀 `3ae43ed8e78ffa673ca238227df6ca7202c1ce70` 에서 측정했고, `git diff --stat 3ae43ed8e a0d801409 -- internal .claude cmd .moai/config` 가 비어 있어(이 개정에서 관측) 인용 코드는 같다.

**개수 규칙.** 수락 기준 = `### AC-` 제목 수 = **15**(하위 ID 없음). 요구 = `### REQ-` 제목 수 = 15. 둘 다 Tier M 상한 16 이하. 묶음 정직 표기: 한 제목 아래 여러 단정이 있다 — 각 AC 본문에 (a)(b)… 로 번호를 매겼고 그 합이 독립 단정 수다(§E).

## §A 관측 규율 [HARD]

AC 는 **종료 코드와 stdout/stderr 내용**(stderr 는 `os.Stderr` 와 `cmd.ErrOrStderr()` 를 도착 순서로 합친 운영 형태), **리뷰어 호출 횟수**(주입 seam: `codexSession`·`codexLookPath`·리뷰 RPC 스텁), **락·재전달 기록·억제 로그의 존재와 내용**(git 디렉터리 아래), **설정 파일의 키·값**을 관측한다. verdict 값 단독은 근거가 못 된다. 동시성 AC 는 장벽으로 겹침을 **강제**한다. 라이브 Claude Code·라이브 codex 를 요구하는 AC 는 없다 — 프로브는 기록되는 관측이다(AC-011).

## §B 픽스처

- **트리 픽스처**: 실제 git 저장소(리뷰 대상 변경 포함, **`.gitignore` 에 `.moai/` 줄이 없다**). 변형: `develop` 브랜치 트리(비카드), 디렉터리 이름 `card-t9999` · 브랜치 `WT-fixture-card` 인 카드 워크트리(디렉터리 이름 ≠ 브랜치 이름), 같은 저장소의 두 번째 연결 워크트리, 하위 디렉터리·심볼릭 링크로 들어가는 경로.
- **리뷰어 스텁**: `pass`·`fail`(findings 3개 이상 / 0건 / 50건×5000자)·`inconclusive`(오류 포함)를 돌려주고, 호출 횟수를 세며, 호출 중 콜백(트리 변경, 장벽 대기)을 실행할 수 있다. **항상 실패하는** 변형은 호출마다 같거나 재서술한 findings 를 준다.
- **seam**: 시계(`now`), 세션 id, 상태 계산(블록 가능), codex 버전 프로브, 설정 다이제스트, 전달 쓰기(실패 주입 가능), `reviewGateBeforeRecord`(리뷰어 반환 뒤 기록 쓰기 전).
- **설정 픽스처**: `enabled: true` + (선택) `tree_scope: skip`.
- **상태 A,B,C,D,E…**: 서로 다른 파일 편집으로 만든 서로 다른 상태 키.

## §C RED/GREEN 두 칸 규율 [HARD]

`verification-completeness.md` §2 를 따른다. 기준마다 **RED-now 칸**(구현 전 트리의 실패와 그 이유)과 **GREEN 경로 칸**(어느 마일스톤이 무엇으로 뒤집는가, 통과 시 출력)을 짝으로 갖는다(§D 표의 두 열). 이미 초록인 기준은 RED 를 가장하지 않고 **회귀 칸(preserve)**으로 분류한다.

- **소스 수준 RED-now** — 아래 증거 장부(읽기 전용 단일 호출, stdout 원문, 종료 코드, 문서 수준 트리 핀). 모든 소스 셀은 계획된 이름·파일로 **뒤집힌다**(GREEN 열).
- **행동 수준 RED-now** — 시험이 아직 없어 plan-time 에 실행할 수 없다. **M1 에서 시험을 먼저 추가해 `-v` 로 `=== RUN`·`--- FAIL` 을 관측하고 `.moai/reports/t1422/red-async/` 에 구현 전 트리 SHA 와 함께 보존할 때까지 릴리스 차단 자격이 없다.** RED 는 §D "RED 이유"가 가리키는 사유로 빨간 것이어야 하며 다른 사유(픽스처 오류)의 RED 는 무효다.
- **사후 조건(post-condition) 기준**: 커밋 순서처럼 구현 뒤에만 관측되는 항은 RED-now 칸을 가질 수 없다. 그 항은 `verification-completeness.md` §2.1 의 미결정 처분에 따라 **릴리스 차단이 아닌 회귀 가드**로 분류하고 통과로 기록하지 않는다(AC-011, AC-012 의 해당 항).
- **회귀 칸**: AC-002 의 종료 코드 0 항·AC-003 의 off 스위치 항·AC-009 전체·AC-012 의 off 스위치 항.

### 증거 장부 (트리 핀 `a0d8014096f894c3a45577482b0eb92892716e04`)

```text
[L1] command: grep -c asyncRewake .claude/settings.json
     stdout:  0
     exit:    1
     red because: 저장소 설정에 asyncRewake 키가 없다 — AC-007
     green: M5 에서 ≥1 (codex 항목 하나)
[L2] command: grep -c asyncRewake internal/template/templates/.claude/settings.json.tmpl
     stdout:  0
     exit:    1
     red because: 템플릿에 asyncRewake 키가 없다 — AC-007
     green: M5 에서 ≥1
[L3] command: grep -c 'codex review gate: FAIL' internal/template/templates/.claude/hooks/moai/handle-codex-review-gate.sh
     stdout:  0
     exit:    1
     red because: 래퍼에 판정 sentinel 처리가 없다 (마지막 줄이 무조건 exit 0) — AC-003
     green: M3 에서 ≥1
[L4] command: grep -c AsyncRewake internal/template/hook_entries.go
     stdout:  0
     exit:    1
     red because: 와이어링 비교 키에 asyncRewake 가 없다 — AC-007
     green: M4 에서 ≥1
[L5] command: grep -c exitCodeError internal/cli/codex_review_gate.go
     stdout:  0
     exit:    1
     red because: 실패를 종료 코드 2 로 옮기는 코드가 없다 — AC-001
     green: M2 에서 ≥1 (실행기는 이 파일에 있다, plan §F.1)
[L6] command: grep -rl 'AcquireReviewLock' internal/cli --include='*.go' --exclude='*_test.go'
     stdout:  (출력 없음)
     exit:    1
     red because: 핸들러에 리뷰 락이 없어 Stop N 번이면 리뷰가 N 번 돈다(코드 판독) — AC-004, AC-010 의 양성 대조 (파일 집합 스윕: 정의 파일과 호출 파일 어느 쪽이든 잡는다)
     green: M2 에서 internal/cli/codex_review_lock.go 와 internal/cli/codex_review_gate.go 두 줄
[L7] command: grep -c -i stale internal/cli/codex_review_gate.go
     stdout:  0
     exit:    1
     red because: 오래된 결과 표지가 없다 — AC-006
     green: M2 에서 ≥1 (요약 구성은 이 파일에 있다, plan §F.1)
[L8] command: grep -rl 'codex-review-wake.json' internal/cli --include='*.go' --exclude='*_test.go'
     stdout:  (출력 없음)
     exit:    1
     red because: 재전달 기록 파일을 쓰는 코드가 없다 — AC-005, AC-014
     green: M2 에서 internal/cli/codex_review_lock.go 한 줄
[L9] command: grep -c 'return emitHookOutput(' internal/cli/codex_review_gate.go
     stdout:  3
     exit:    0
     red because: 실행기 세 분기(:197,:205,:210)가 모두 stdout JSON 을 낸다 — AC-002 의 "stdout 비어 있음" 항, AC-001
     green: M2 에서 0 (함수 정의 줄은 `return` 으로 시작하지 않으므로 패턴에 잡히지 않는다)
[L10] command: ls .moai/reports/t1422/async-probe/E1-interactive.txt
     stdout:  (없음; stderr: ls: .moai/reports/t1422/async-probe/E1-interactive.txt: No such file or directory)
     exit:    1
     red because: E-1 을 결정하는 프로브 기록 파일이 없다(디렉터리 async-probe/ 와 kit/ 은 절차용으로 이미 있으므로 기록 파일 하나를 지목한다) — AC-011
     green: M1 에서 파일이 존재하고 sha256 이 progress.md 표에 있다
[L11] command: grep -rl 'codex-review-gate.log' internal/cli --include='*.go' --exclude='*_test.go'
     stdout:  (출력 없음)
     exit:    1
     red because: 억제 로그를 쓰는 코드가 없다 — AC-013
     green: M2 에서 internal/cli/codex_review_lock.go 한 줄
[L12] command: grep -rl 'FINAL NOTICE\|이후 알림 없음' internal/cli --include='*.go' --exclude='*_test.go'
     stdout:  (출력 없음)
     exit:    1
     red because: 상한 도달 시 마지막 알림 문구(`이후 알림 없음, 상태는 미해결`)를 내는 코드가 없다 — AC-015
     green: M2 에서 internal/cli/codex_review_gate.go 한 줄
[P1] command: grep -c '"timeout": 900' .claude/settings.json
     stdout:  2
     exit:    0
     red because (AC-008): 두 래퍼 모두 timeout 900 이고 Go 쪽 리뷰 상한도 900s(config.DefaultCodexReviewGateTimeout, defaults.go:557)라 "등록 값 > 핸들러 총 상한 + 마진"이 거짓이다
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
     positive control: L5, L7, L9 의 값이 빈 스윕이 아님 — 같은 파일은 검색 가능하다
[P5] command: grep -c 'codex-review.lock\|codex-review-wake' internal/cli/codex_stop_chain.go internal/cli/codex_review_receipt.go
     stdout:  internal/cli/codex_review_receipt.go:0
              internal/cli/codex_stop_chain.go:0
     exit:    1
     preserve: Codex 경로 파일은 새 락·상태를 참조하지 않는다 (AC-009) — 구현 뒤에도 0
[P6] command: grep -c reviewScopeResolver internal/cli/codex_stop_chain.go internal/cli/codex_review_receipt.go
     stdout:  internal/cli/codex_stop_chain.go:1
              internal/cli/codex_review_receipt.go:2
     exit:    0
     positive control: P5 의 0 이 빈 스윕이 아님
[P7] command: grep -rl 'reviewScopeResolver' internal/cli --include='*.go' --exclude='*_test.go'
     stdout:  internal/cli/codex_review_scope.go
              internal/cli/codex_stop_chain.go
              internal/cli/codex_review_gate.go
              internal/cli/codex_review_receipt.go
     exit:    0
     positive control: L6, L8, L11, L12 의 "출력 없음"이 빈 스윕이 아님 — 같은 -rl 형태가 파일을 돌려준다
[R1] command: go test ./internal/cli/ -run '^(TestReviewGate_CodexFailBlocks|TestReviewGate_CodexPassAllows|TestReviewGate_FailOpenOnMissingCodex|TestReviewGate_FailOpenOnCodexError|TestReviewGate_InconclusiveAllows)$' -count=1 -v   (환경 세척)
     stdout:  --- PASS ×5, ok github.com/modu-ai/moai-adk/internal/cli 1.856s   (핀 3ae43ed8e 에서 측정)
     exit:    0
     preserve: 핸들러 판정 분기 (AC-001/002 의 판정 방향; AC-005 의 "상태가 바뀐 Stop 은 리뷰")
[R2] command: go test ./internal/template/ -run '^(TestReviewGatesRegisteredInRepoSettings|TestReviewGatesRegisteredInTemplateSettings)$' -count=1 -v
     stdout:  --- PASS ×2, ok github.com/modu-ai/moai-adk/internal/template 0.479s   (핀 3ae43ed8e 에서 측정)
     exit:    0
     preserve: 등록 시험 (AC-007 의 변경 전 초록)
[R3] command: go test ./internal/hook/ -run '^(TestReviewGateWrappers_DisabledCostsZeroColdStarts|TestReviewGateWrappers_EnabledReachesBinary)$' -count=1 -v
     stdout:  --- PASS ×2, ok github.com/modu-ai/moai-adk/internal/hook 1.384s   (핀 3ae43ed8e 에서 측정)
     exit:    0
     preserve: 래퍼 off 스위치·바이너리 도달 (AC-003, AC-012)
[R4] command: go test ./internal/cli/ -run '^(TestProduceCodexReviewReceipt_CardScopeRecordsCardState|TestCodexStopHandlerRunsTheChain)$' -count=1 -v
     stdout:  --- PASS ×2, ok github.com/modu-ai/moai-adk/internal/cli 3.446s   (핀 3ae43ed8e 에서 측정)
     exit:    0
     preserve: Codex 경로 (AC-009)
[R5] command: go test ./internal/cli/ -run '^(TestGateLock_AcquireRecordsHolderIdentityAndReleaseAllowsReacquire|TestGateLock_ContentionReturnsSentinel|TestGateLock_DeadHolderArtifactAcquiresFarBelowBudget|TestGateLock_ReadOnlyStateDirIsUnavailableNotFatal)$' -count=1 -v
     stdout:  --- PASS ×4, ok github.com/modu-ai/moai-adk/internal/cli 1.254s   (핀 3ae43ed8e 에서 측정)
     exit:    0
     preserve: gate-run 락 일반화 전 초록 (AC-004 의 회귀 항)
```

L1-L12·P1-P7 은 이 개정에서 핀 `a0d801409` 로 다시 실행했다. `grep -c` 의 종료 코드 1 은 "0건"이며 하네스가 조용히 표시한다(감사도 같은 관측). R1-R5 는 이 개정에서 다시 실행하지 않았다(핀 3ae43ed8e, 인용 코드 동일) — **미재측정 Gap**, M1 이 변경 전 트리에서 다시 관측한다.

### 변이 점검(mutant sweep) — 모든 AC 에 대해 요구를 어기면서 기준을 만족하는 가장 값싼 변이와 그것을 죽이는 칸

"죽임" = 해당 단정이 그 변이에 대해 빨갛다(작성 시점 판독 — 구현이 없으므로 실행 관측이 아니다. M1 이 변이 몇 개를 실제로 심어 빨강을 관측하는 것은 §E 선택 항목). "열림" = plan 단계에서 죽일 수 없는 변이.

| AC | 변이 | 죽이는 칸 | 상태 |
|---|---|---|---|
| 001 | 옛 신호(stdout JSON block)를 병행 | (c) stdout 비어 있음 | 죽임 |
| 001 | 요약에서 reviewed HEAD·findings 줄 누락 | (d) 줄별 내용 단정 | 죽임 |
| 001 | 카드 칸을 브랜치 이름으로 채움 / 비카드에도 채움 | (e) 디렉터리 이름 ≠ 브랜치 이름 픽스처, `develop` 공란 | 죽임 |
| 001 | 스코프 로그 행을 요약 뒤에 쓰거나 sentinel 앞에 다른 텍스트 | (b) sentinel 정확히 한 줄, 그 앞은 스코프 행뿐 | 죽임 |
| 001 | sentinel 오타·접두 | (b) 정확한 줄 일치 | 죽임 |
| 001 | 크기 상한 없음 | (g) 50건×5000자 픽스처 ≤ 8000바이트·10건·`(+40 more)` | 죽임 |
| 001 | 0건 findings fail 을 침묵하거나 충돌 | (h) 0건 fail 변형 | 죽임 |
| 001 | cobra `Error:` 줄 추가 | (f) 합친 스트림에 `Error:` 부재 + `moaiErrorHandler` 출력 없음 | 죽임 |
| 002 | inconclusive·오류를 종료 코드 2 로 | (a) 변형별 종료 코드 0 | 죽임 |
| 002 | 통과에도 알림 문구를 stderr 에 씀 | (b) stderr 허용 목록(스코프 행·`codex-review-gate: error:`)만 | 죽임 |
| 002 | stdin 파싱 실패에 `{}` 를 stdout 에 냄 | (a) 변형 (j) stdout 비어 있음 | 죽임 |
| 002 | 락 경합·알려진 상태·상한 억제에서 요약을 냄 | (a) 변형 (h)(i)(k) sentinel 부재 | 죽임 |
| 003 | 래퍼가 항상 0 | (a) sentinel 2 → 2 | 죽임 |
| 003 | 래퍼가 모든 2 를 전달(패닉 포함) | (b) 패닉 2 → 0, stderr 비어 있음 | 죽임 |
| 003 | 래퍼가 stderr 전체(스코프 행 포함)를 전달 | (a) 전달 stderr 가 sentinel 줄부터와 정확히 같다 | 죽임 |
| 003 | sentinel 을 부분 문자열로 판정 | (b) 줄 중간 sentinel(iii) → 0 | 죽임 |
| 003 | 등록 shim 을 `bash "$0"; exit 0` 로 | (f) 등록 명령 문자열을 그대로 실행 + `exec` 포함 단정 | 죽임 |
| 003 | 비판정 2 의 로그 행 누락 | (b) git 디렉터리 로그 행 | 죽임 |
| 003 | 한쪽 사본(C1/C2)만 수정 | 두 사본 각각 실행 | 죽임 |
| 004 | 경합 시 기다림(직렬화) | (a) 장벽: 나머지가 보유자 해제 전에 반환(기다리면 교착 → 스텁 시간 제한 실패) | 죽임 |
| 004 | 락을 리뷰 호출 뒤에 잡음 | (a) 호출 수 1 | 죽임 |
| 004 | 락을 세션별로 키링 | (a) 서로 다른 session_id | 죽임 |
| 004 | 락을 cwd 로 키링 | (b) 하위 디렉터리·심볼릭 링크 경로 Stop 이 같은 락에 걸림 | 죽임 |
| 004 | 락을 저장소 공통 디렉터리로 키링 | (c) 같은 저장소 두 연결 워크트리는 각자 1 회 | 죽임 |
| 004 | 락을 리뷰어 반환 직후 풀고 기록을 락 밖에서 씀 | (d) 기록 쓰기 전 장벽: 나머지가 리뷰어를 부르지 않아야 함 | 죽임 |
| 004 | 오류 경로에서 해제 누락 | (e) 직후 순차 Stop 이 리뷰를 다시 돌림 | 죽임 |
| 004 | 상태 계산이 막혀도 락을 무기한 보유 | (g) 상태 계산 seam 차단 + 짧은 예산: 한도 안에 반환·`state-unavailable`·락 해제 | 죽임 |
| 004 | 락·기록·로그를 작업 트리 아래에 둠 | (h) 무시 줄 없는 저장소에서 `git status --porcelain` 불변·파일은 git 디렉터리 | 죽임 |
| 005 | 키를 HEAD+digest 만으로 | (d) ToolVersion·ConfigDigest·Command 변형 | 죽임 |
| 005 | fail 을 stderr 쓰기 전에 delivered 로 기록 | (e) 쓰기 실패 주입 → 다음 Stop 이 재전달 | 죽임 |
| 005 | 세션 id 무시 | (c) 다른 세션이 같은 상태의 fail 을 받음 | 죽임 |
| 005 | 전달 창 무시 | (f) 시계 +31분 → 재리뷰 | 죽임 |
| 005 | 건너뜀을 pass 에 적용하지 않음(현행) | (b) 같은 상태 pass 재발생 시 호출 수 불변 | 죽임 |
| 005 | 건너뜀이 상태 변경 뒤에도 적용 | (g) 변경 상태 Stop 은 리뷰(회귀 칸 R1) | 죽임 |
| 005 | inconclusive 를 알려진 상태로 기록 | (h) 같은 상태 inconclusive 재시도 시 재호출 | 죽임 |
| 006 | stale 판정을 HEAD 만 비교 | (b) digest 만 다른 변형 | 죽임 |
| 006 | stale 판정을 digest 만 비교 | (d) HEAD 만 다른 변형(빈 커밋) | 죽임 |
| 006 | 항상/절대 표지 | (a)(c) stale 은 표지 있음, 변경 없음은 표지 없음 | 죽임 |
| 006 | stale 실패를 조용히 버림 | (a)(b)(d) 종료 코드 2 | 죽임 |
| 006 | stale 실패를 현재(미리뷰) 키로 기록 | (f) 이어지는 현재 상태 Stop 이 리뷰어를 부름 | 죽임 |
| 007 | 다른 Stop 항목에도 asyncRewake | (a) 파일 전체 `asyncRewake` 키 수 1 이고 codex 항목 소속 | 죽임 |
| 007 | `async` 로 잘못 씀 | (a) 키 이름·`async` 부재 단정 | 죽임 |
| 007 | multi 래퍼에도 | (a) multi 항목 부재/false | 죽임 |
| 007 | 플래그를 비교 키에서 뺌 | (b) 플래그 제거 변형에서 `templateOnly` 비어 있지 않음 | 죽임 |
| 007 | 정렬 tie-break 가 플래그를 무시 | (c) 플래그만 다른 두 항목의 섞은 입력 20회 결정적 순서 | 죽임 |
| 008 | 등록 timeout = 핸들러 총 상한(마진 0) | (a) timeout > 1020 + M | 열림(M 은 O-B — 값이 정해질 때까지 통과로 기록하지 않음) |
| 008 | multi 래퍼 timeout 변경 | (b) multi 900 유지 | 죽임 |
| 009 | Codex 경로가 락·기록을 읽음 | (b) P5 정적 0 (양성 대조 P6) + (a) R4·parity golden | 죽임 |
| 009 | 런타임 파일이 작업 트리에 생겨 `verify.Key` 변경 | (c) 무시 줄 없는 저장소에서 Claude 게이트 실행 전후 `verify.Key` 동일 | 죽임 |
| 009 | 게이트가 영수증 저장소를 건드림 | (d) 영수증 목록 호출 전후 동일 | 죽임 |
| 010 | 락을 skip 판정 앞에서 잡음 | (a) skip 설정에서 git 디렉터리에 락·기록·로그 부재 | 죽임 |
| 010 | skip 인데 리뷰어 호출 | (a) 호출 0 | 죽임 |
| 010 | 대조(review 설정)에서 락이 안 만들어짐 | (b) 대조 항: 락 생성·호출 1 | 죽임 |
| 011 | 프로브 파일만 있고 결과 토큰 없음 | (a) 토큰·sha256 단정 | 죽임 |
| 011 | 프로브 ID 가 절차 문서와 다름 | (a) 아홉 파일 이름 | 죽임 |
| 011 | 해시 기록 커밋이 등록 커밋보다 뒤 | (b) 커밋 그래프 조상 관계 | 열림(사후 조건 — 회귀 가드, 통과로 기록 안 함) |
| 012 | 비활성 설정에서 moai 프로세스 기동 | (a) 가짜 moai 호출 수 0 | 죽임 |
| 012 | 등록 커밋 뒤에 run-phase 커밋 | (b) 커밋 순서 | 열림(사후 조건 — 회귀 가드) |
| 013 | 억제 사유 행 누락/오류 | (a) 사유별 행 단정 | 죽임 |
| 013 | 억제가 아닌 정상 리뷰·마지막 알림에 행 | (a) 억제 아닌 이벤트는 행 없음 | 죽임 |
| 013 | 로그를 작업 트리에 씀 | (b) 위치 단정 | 죽임 |
| 013 | 래퍼 `non-verdict-exit-2` 행 누락 | AC-003 (b) | 죽임 |
| 014 | 지문(재서술·0건)에 의존해 셈 | (b)(c) 재서술·0건 fail 변형이 일반 변형과 같은 깨움 횟수 | 죽임 |
| 014 | 세션별 카운터 | (a) 두 세션이 번갈아 와도 3 번째가 마지막 알림 | 죽임 |
| 014 | 카운터를 프로세스 메모리에만 | (a) 매 Stop 이 새 실행기 호출(파일 기록) | 죽임 |
| 014 | 상한 뒤 일부 실패를 다시 알림(네 번째 깨움) | (d) 상한 뒤 D·E·G 모두 침묵 | 죽임 |
| 014 | pass 가 카운터를 리셋하지 않음 | (e) pass 뒤 fail 이 1 번째 깨움 | 죽임 |
| 014 | inconclusive·상태 변경이 카운터를 리셋 | (f) 사이에 inconclusive·새 상태가 끼어도 횟수 유지 | 죽임 |
| 014 | 상한 뒤 리뷰어를 호출하지 않음(pass 를 못 알아봄) | (d) 상한 뒤 리뷰어 호출 수 증가 | 죽임 |
| 015 | 고정 한국어 문구 누락/영어만 | (a) 정확한 부분 문자열 | 죽임 |
| 015 | 문구가 2·4 번째 알림에 | (a) 전체 시퀀스에서 출현 수 1, 위치 3 번째 | 죽임 |
| 015 | 크기 상한이 마지막 알림을 자름 | (c) 최대 크기 픽스처에서 문구 온전 | 죽임 |
| 015 | 상한에서 알림 없이 침묵(`cap3_silent`) | (a) 3 번째 알림에 문구 | 죽임 |
| 015 | stale 과 마지막 알림 동시 발생 시 한쪽 누락 | (d) 결합 변형 | 죽임 |

## §D AC 매트릭스

| AC | 요구 | 시작 상태 | RED 이유 / 회귀 근거 | GREEN 경로 (마일스톤 → 통과 출력) | 관측 대상 |
|---|---|---|---|---|---|
| AC-001 | REQ-CRA-001 | RED(행동) | 종료 코드 2 매핑 부재 (L5) | M2 → 시험 PASS, L5 ≥1 | 실패 변형: 종료 코드·합친 stderr 스트림·stdout·크기 |
| AC-002 | REQ-CRA-002 | RED(stdout 항) + 회귀(종료 코드 항) | allow 에도 stdout JSON (L9=3); 종료 코드 0 은 R1 초록 | M2 → L9=0, 시험 PASS | 조용한 결과 열한 변형의 종료 코드·stdout·요약 부재 |
| AC-003 | REQ-CRA-003 | RED(전파 항) + 회귀(off 스위치) | sentinel 처리 부재 (L3); off 스위치 R3 초록 | M3 → L3 ≥1, 시험 PASS | 등록 명령 문자열 실행의 종료 코드·stderr 행렬 |
| AC-004 | REQ-CRA-004 | RED(행동) + 회귀(gate-run 락) | 락 부재 (L6); R5 초록 | M2 → L6 두 줄, 시험 PASS, R5 초록 유지 | 장벽 동시성 N≥8 리뷰어 호출 수, 보유 시간, 파일 위치 |
| AC-005 | REQ-CRA-005 | RED(행동) + 회귀(상태가 바뀐 Stop 은 리뷰) | 재전달 기록 부재 (L8); 변경 상태 리뷰는 R1 초록 | M2 → L8 한 줄, 시험 PASS | 상태 시퀀스의 리뷰어 호출 수·종료 코드·delivered 표지 |
| AC-006 | REQ-CRA-006 | RED(행동) | stale 표지 부재 (L7) | M2 → L7 ≥1, 시험 PASS | 트리 변경 변형의 요약·기록 키 |
| AC-007 | REQ-CRA-007 | RED(소스) + 회귀(등록 시험) | L1, L2, L4; R2 초록 | M4(비교)→M5(등록) → L1·L2·L4 ≥1, 시험 PASS | 두 설정 파일의 키·와이어링 diff·정렬 |
| AC-008 | REQ-CRA-008 | RED(소스) | 등록 900 ≤ 핸들러 총 상한 1020 (P1) | M4(마진 확정)→M5 → timeout > 1020 + M | 등록 timeout 대 핸들러 총 상한+마진 |
| AC-009 | REQ-CRA-009 | **회귀(초록)** + 키 안정성 RED | R4, P5; 키 안정성은 행동 RED(작업 트리에 파일을 두는 구현이 빨강) | M2 → R4·parity golden 초록 유지, P5=0 유지, 키 동일 | Codex 경로 시험·정적 참조·영수증·`verify.Key` |
| AC-010 | REQ-CRA-010 | RED(양성 대조 항) | skip 이 아닌 세션이 락 파일을 만드는 대조 항이 오늘 빨강 (L6) | M2 → 대조 항 PASS | skip 세션의 락·기록·로그·리뷰어 부재 |
| AC-011 | REQ-CRA-011 | RED(존재 항) + 사후 조건(순서 항) | 프로브 기록 부재 (L10) | M1 → 아홉 파일·sha256 표 | 프로브 파일·sha256·결과 토큰·커밋 순서 |
| AC-012 | REQ-CRA-012 | **회귀(off 스위치)** + 사후 조건(순서) | R3 | M5 → R3 초록 유지 | off 스위치 모사·커밋 순서 |
| AC-013 | REQ-CRA-013 | RED(행동) | 억제 로그 부재 (L11) | M2/M3 → L11 한 줄, 시험 PASS | 억제 사유별 로그 행·위치 |
| AC-014 | REQ-CRA-014 | RED(행동) | 상한·기록 부재 (L8) | M2 → 시험 PASS | 상태 시퀀스의 깨움 횟수·후속 침묵 |
| AC-015 | REQ-CRA-015 | RED(행동) | 마지막 알림 문구 부재 (L12) | M2 → L12 한 줄, 시험 PASS | 세 번째 알림의 문구·크기 상한과의 공존 |

REQ→AC 누락 없음: 001→001 · 002→002 · … · 015→015 (1:1, 번호 동일).

---

### AC-001 — 실패는 종료 코드 2 + stderr 요약이고 stdout 은 비어 있다 (운영 형태 스트림)

**Given** 리뷰 대상 변경이 있는 트리 — 디렉터리 이름 `card-t9999`·브랜치 `WT-fixture-card` 인 카드 워크트리, 비카드 `develop` 트리 — 와 리뷰어 스텁이 (i) findings 3개 이상 (ii) findings 0건이고 `summary` 만 있는 fail (iii) 50건 findings(제목 5000자)를 담은 `fail` 을 돌려주고, 시작 시점 HEAD 가 `H0`; 프로세스 `os.Stderr` 를 파이프로 바꿔 잡고 `cmd.ErrOrStderr()` 버퍼와 도착 순서로 합친 스트림을 만든다,
**When** `moai hook codex-review-gate` 실행기(cobra 명령)가 Stop 입력을 처리하고, 이어 돌려받은 오류를 `moaiErrorHandler` 에 넘기면,
**Then** (a) 종료 상태는 코드 2(`ExitCoder`). (b) 합친 스트림에서 정확히 `codex review gate: FAIL` 인 줄이 **한 번** 있고, 그 앞의 모든 줄은 스코프 로그 JSON 행(`"gate":"codex-review-gate"`를 담은 유효한 JSON)뿐이다 — 운영 형태이므로 "스트림의 첫 줄"이 아니라 "sentinel 줄이 요약 블록의 첫 줄"이 계약이다. (c) stdout 은 비어 있다(`decision: block` JSON 이 나오지 않는다). (d) sentinel 줄 다음부터 줄별로 `card:`·`branch:`, `reviewed HEAD: H0`, `summary:`, `findings (<n>):` 와 각 finding 의 `[severity] file:line title` 이 있다. (e) 카드 칸은 카드 워크트리에서 정확히 `card-t9999`(브랜치 이름 `WT-fixture-card` 가 아니다), `develop` 트리에서 공란이다. (f) 스트림에 `Error:` 로 시작하는 줄이 없고 `moaiErrorHandler(err)` 는 아무것도 쓰지 않는다. (g) 변형 (iii): 요약(sentinel 줄부터)은 8000바이트 이하이고 finding 줄은 10건 이하이며 각 줄 300자 이하, `(+40 more)` 가 있고 `summary` 줄은 1500자 이하이며 `[truncated]` 표지가 있고 sentinel 줄은 잘리지 않는다. (h) 변형 (ii): 종료 코드 2, `findings (0):`, 내용은 `summary` 줄에 있다. (i) 합친 스트림에 스코프 행이 정확히 한 번 있다(중복 없음).

### AC-002 — 그 밖의 결과는 종료 코드 0, 빈 stdout, 요약 없음 (열한 변형)

**Given** 변형 — (a) 리뷰어 `pass` (b) `inconclusive` (c) 리뷰 호출 오류 (d) codex 바이너리 부재 (e) `enabled: false` (f) `stop_hook_active: true` (g) 리뷰할 변경 없음 (h) 락 경합으로 억제(AC-004) (i) 알려진 상태로 억제(AC-005) (j) **파싱할 수 없는 stdin JSON** (k) 연속 실패 상한 뒤 실패(AC-014),
**When** 실행기가 각 Stop 입력을 처리하면(합친 스트림 관측),
**Then** 열한 변형 모두 (a) 종료 코드 0, (b) stdout 비어 있음, (c) 스트림에 `codex review gate: FAIL` 줄이 없고 스코프 로그 행과 기존 진단 줄(`codex-review-gate:` 로 시작)만 있다 — 통과·억제에 알림 문구가 없다. (스코프 행이 종료 코드 0 에서 Claude 에 보이는지는 E-4 d2 가 관측한다, plan §G 위험 9.)

### AC-003 — 래퍼는 판정 실패의 종료 코드 2 에서만 깨우고 등록된 명령 형태로 검증한다

**Given** 설정 파일(`.claude/settings.json` 과 템플릿 렌더)의 codex 래퍼 항목에서 읽은 **등록 명령**(`command: bash`, `args: [-c, <shim>, <스크립트 경로>]`, `${CLAUDE_PROJECT_DIR}` 치환)을 그대로 실행하는 하네스; PATH 의 `moai` 를 가짜 실행 파일로 바꾼다. 가짜의 동작 — (i) 스코프 로그 JSON 행 한 줄 뒤에 sentinel 요약을 stderr 에 쓰고 종료 코드 2 (ii) 패닉 모양 stderr(`panic: boom` + `goroutine 1 [running]:` 줄들)와 종료 코드 2 (iii) `x codex review gate: FAIL`(줄 중간 sentinel)과 종료 코드 2 (iv) stderr 비어 있는 종료 코드 2 (v) 종료 코드 0·1·127 (vi) `moai` 가 `PATH` 에 없음; 설정이 게이트를 켜지 않은 픽스처(`review_gate` 키 부재, `enabled: false`); `mktemp` 없는 PATH 변형; git 저장소 픽스처(로그 행 관측용); C1 과 C2 사본 각각,
**When** 등록 명령을 실행하면,
**Then** (a) (i): 종료 코드 2 이고 stderr 는 sentinel 줄부터 끝까지와 **정확히 같다**(스코프 행 없음). (b) (ii)(iii)(iv): 종료 코드 0, stderr 에 패닉 텍스트·sentinel 이 없고, git 디렉터리 `moai/codex-review-gate.log` 에 `non-verdict-exit-2` 행이 한 줄 추가된다. (c) (v): 종료 코드 0 (종료 코드 0/1/127 의 기존 진단은 stderr 에 그대로 통과). (d) (vi) `moai` 부재: 종료 코드 0. (e) 비활성 설정: 종료 코드 0, 가짜 `moai` 호출 수 0(순수 셸 off 스위치 — `TestReviewGateWrappers_DisabledCostsZeroColdStarts` 계열 회귀). (f) 등록 shim 문자열에 `exec bash "$0"` 가 들어 있다(종료 상태 전달을 구조로 고정). (g) `mktemp` 없는 PATH: 종료 코드 0, 깨움 없음(안전 쪽).

### AC-004 — 트리당 동시에 리뷰는 하나, 보유 시간은 유한, 파일은 git 디렉터리에 있다

**Given** 같은 트리에 대한 N=8 개의 동시 Stop(서로 다른 `session_id` 포함), 호출 횟수를 세는 리뷰어 스텁이 첫 호출에서 나머지 N−1 호출이 모두 반환할 때까지 대기(스텁 쪽 시간 제한으로 교착 대신 실패); 별도로 (a') 죽은 PID 를 적었으나 flock 은 잡혀 있지 않은 낡은 락 파일, (b') 락을 잡은 하위 프로세스를 SIGKILL 한 직후; 하위 디렉터리·심볼릭 링크 경로로 들어오는 Stop 변형; 같은 저장소의 두 연결 워크트리; `reviewGateBeforeRecord` seam 에서 홀더가 멈추는 변형; 상태 계산 seam 이 취소될 때까지 막는 변형(예산 200ms 주입); 리뷰어 호출이 예산(300ms 주입)까지 막는 변형; 무시 줄이 없는 저장소,
**When** 모든 호출이 핸들러(또는 실행기)를 지나면,
**Then** (a) 리뷰어 호출 수는 정확히 1 이고 나머지 7 호출은 보유자가 아직 리뷰 중일 때 반환하며(기다림이 아닌 건너뜀) 종료 코드 0·빈 stdout 이다; 다중 프로세스 변형(N 개의 별도 프로세스)에서도 1. (b) 하위 디렉터리·심볼릭 링크 경로로 들어온 Stop 도 같은 락에 걸린다(호출 수 1). (c) 같은 저장소의 서로 다른 두 연결 워크트리는 각자 정확히 1 회씩 리뷰한다(서로 막지 않는다). (d) 홀더가 기록 쓰기 직전에 멈춰 있는 동안 도착한 Stop 7 건은 리뷰어를 부르지 않고 반환한다(락이 기록 쓰기까지 유지됨). (e) (a')(b') 에서 다음 Stop 은 락을 얻어 리뷰어를 1 회 호출한다; 보유자가 오류로 끝난 직후 다른 상태에서 오는 Stop 도 리뷰어를 다시 호출한다(해제 확인). (f) 양성 대조: 락을 끈 시험 설정에서는 호출 수가 N 이다(오늘의 동작 — M1 RED 관측). (g) 상태 계산이 막혀도 핸들러는 예산의 5 배 안에 반환하고 `state-unavailable` 행을 쓰며 락을 해제한다(이어지는 Stop 이 락을 얻는다); 리뷰어가 예산까지 막히면 inconclusive 로 종료 코드 0 이고 락이 해제된다. (h) 무시 줄이 없는 저장소에서 게이트 실행 뒤 `git status --porcelain` 은 실행 전과 같고, 락·기록·로그는 `git rev-parse --absolute-git-dir` 아래 `moai/` 에만 있으며 작업 트리의 `.moai/` 아래에는 새 파일이 없다. (i) 락 일반화 뒤에도 gate-run 락 시험(R5)이 초록이다.

### AC-005 — 같은 상태는 다시 리뷰도 깨움도 없다, 전달 확인 뒤에만 (다섯 필드 키)

**Given** 시험마다 새 트리 픽스처(연속 실패 상한에 닿지 않도록), 세션 `s1`·`s2`, 주입 시계, 호출 횟수를 세는 리뷰어 스텁, 전달 쓰기 seam(실패 주입 가능), codex 버전 프로브·설정 다이제스트·명령 값을 바꿀 수 있는 seam,
**When** 아래 시퀀스를 처리하면,
**Then** (a) 상태 A `fail`(s1): 종료 코드 2, 기록 `delivered` 가 stderr 쓰기 성공 **뒤** 참이 된다. (b) 같은 A, s1, 창(30분) 안: 리뷰어 호출 수 불변·종료 코드 0·행 `state-unchanged`; 상태 B `pass` 뒤 같은 B(어느 세션이든): 호출 수 불변(**pass 에도 건너뜀 — 현행 동작 변경, O-H**). (c) 같은 A, **s2**: 리뷰어 호출됨·종료 코드 2(다른 세션은 한 번 받는다). (d) B 가 `pass` 로 기록된 뒤 **`ToolVersion` 만 바꾸면**, **`ConfigDigest` 만 바꾸면**, **`Command` 만 바꾸면** 각각 리뷰어가 다시 호출된다(다섯 필드 키). (e) **전달 전에 죽은 fail**: 전달 쓰기 seam 이 실패하는 시퀀스 — 기록은 `delivered` 가 아닌 채 남고 같은 A·같은 s1 의 다음 Stop 이 리뷰어를 다시 호출해 종료 코드 2 로 전달한다. (f) 같은 A, s1, 시계 +31분: 리뷰어 호출됨. (g) **회귀 칸(R1)**: 상태가 바뀐 Stop(C)은 항상 리뷰어를 호출한다 — 건너뜀은 상태 키가 같은 Stop 에만 적용된다; HEAD 가 같고 digest 만 다른 상태 쌍은 별개 상태로 각각 리뷰된다. (h) 상태 X `inconclusive` 는 기록되지 않아 같은 X 에서 리뷰어가 다시 호출된다. (i) 기록 파일을 지우면 다음 Stop 은 리뷰어를 호출한다.

### AC-006 — 오래된 결과는 표지를 붙여 전달되고 기록 키는 리뷰한 상태다

**Given** 리뷰 중 콜백이 트리를 바꾸는 변형 — (i) 새 커밋(HEAD `H0`→`H1`, 작업 트리 변경 포함) (ii) 추적 파일 편집(HEAD 는 `H0` 그대로, digest 만 변경) (iii) 변경 없음 (iv) **빈 커밋**(`git commit --allow-empty` — HEAD `H0`→`H1`, digest 동일); 리뷰어 스텁은 각각 `fail`, 그리고 (i) 에 대해 `pass` 변형도,
**When** 실행기가 처리하면,
**Then** (a) (i) fail: 종료 코드 2, 요약에 `reviewed HEAD: H0`·`current HEAD: H1`·`STALE: this is a review of an earlier state` 가 모두 있다. (b) (ii) fail: 종료 코드 2, 두 HEAD 가 같은 값 `H0` 으로 적히고 표지가 있다. (c) (iii) fail: 표지 없음, 종료 코드 2. (d) (iv) fail: 종료 코드 2, 표지가 있고 `reviewed HEAD: H0`·`current HEAD: H1` 이 다르다(digest 만 비교하는 구현이 놓친다). (e) (i) pass: 종료 코드 0 침묵이고 기록 키는 시작 상태 `S0` 의 키다. (f) (i)·(ii) fail 직후 현재 상태(`S1`, 한 번도 리뷰하지 않은 상태)에서 Stop 이 오면 리뷰어가 호출된다 — 기록 키는 `S0` 이지 `S1` 이 아니다.

### AC-007 — 레지스트리는 codex 래퍼에만 asyncRewake 를 켜고 와이어링 비교가 그것을 본다

**Given** `.claude/settings.json` 과 `internal/template/templates/.claude/settings.json.tmpl`(렌더 후),
**When** 모든 훅 항목을 파싱하면,
**Then** (a) 파일 전체에서 `asyncRewake` 키는 **정확히 한 번**이고 그것은 codex 래퍼 항목의 `true` 다; multi 래퍼 항목은 `asyncRewake` 가 없거나 false; 두 래퍼 모두 `async` 키는 없다(혼동 변이 차단). (b) `HookEntry` 에 `AsyncRewake` 가 있고 템플릿 렌더 결과와 저장소 설정의 `DiffHookEntries` 는 양방향 빈 결과이며, 항목에서 `asyncRewake` 를 뺀 변형에 대해서는 `templateOnly` 가 비어 있지 않다(와이어링 비교가 플래그 누락을 본다). (c) `asyncRewake` 만 다른 두 항목을 섞어 20회 정렬해도 순서가 같다(`sortHookEntries` tie-break 포함). (d) `moai doctor` 의 비교 필드 문구에 `asyncRewake` 가 들어 있다. (e) 변경 전에는 `TestReviewGatesRegisteredInRepoSettings`·`…InTemplateSettings`(R2)가 초록이고 변경 뒤에도 래퍼별 기대값을 갱신해 초록이다. 이 기준은 E-1 이 확인됨일 때만 효력이 있다(AC-011).

### AC-008 — 등록 타임아웃은 핸들러 총 상한보다 마진만큼 크다

**Given** 두 설정 파일과 `config.DefaultCodexReviewGateTimeout`·상태 계산 예산, 마진 상수 `M`(> 0, `internal/config` 단일 원천),
**When** codex 래퍼 항목의 `timeout` 을 읽으면,
**Then** (a) `timeout` > `int(DefaultCodexReviewGateTimeout.Seconds())` + 2×`int(상태 계산 예산.Seconds())` + `M` (현재 900 ≤ 1020 이라 거짓 — P1). (b) multi 래퍼의 `timeout` 은 900 그대로, Go 쪽 리뷰 상한 값은 900s 그대로다. `M` 의 값은 E-2 결과 뒤 확정된다(plan §B.6, O-B) — 값 확정 전에는 이 AC 를 통과로 기록하지 않는다(**열림**).

### AC-009 — Codex 경로는 그대로이고 새 락·상태를 모르며 작업 트리 키를 바꾸지 않는다 (회귀 칸 + 키 안정성)

**Given** 구현 뒤 트리, 그리고 `.gitignore` 에 `.moai/` 줄이 없는 저장소 픽스처,
**When** Codex 경로 시험과 정적 참조를 검사하고, 그 저장소에서 Claude 게이트를 (락 경합·알려진 상태 억제를 포함해) 여러 번 실행하면,
**Then** (a) R4 의 시험(`TestProduceCodexReviewReceipt_CardScopeRecordsCardState`·`TestCodexStopHandlerRunsTheChain`)과 `TestStopChainEffectParityGolden` 과 M1 이 열거해 고정한 Codex 멤버 6·영수증 시험이 전부 초록이다. (b) `grep -c 'codex-review.lock\|codex-review-wake' internal/cli/codex_stop_chain.go internal/cli/codex_review_receipt.go` 는 두 파일 모두 0 이다(P5, 양성 대조 P6). (c) 그 저장소에서 `verify.Key` 값이 Claude 게이트 실행 전후로 동일하고 `produceCodexReviewReceipt` 가 기록하는 상태 키도 실행 전후 동일하다(런타임 파일이 작업 트리에 없다). (d) Claude 훅을 픽스처 트리에서 실행해도 검증 영수증 저장소(`.moai/state/verify/`)의 목록은 호출 전후 동일하다.

### AC-010 — 형제의 skip 은 락·상태보다 앞서 일어난다

**Given** 형제 SPEC 이 착지한 트리 — `enabled: true`+`tree_scope: skip` 설정과 비 `WT-` 브랜치의 변경이 있는 트리에 8 개의 동시 Stop, 그리고 **양성 대조**로 같은 트리의 `tree_scope: review` 설정,
**When** 실행기가 처리하면,
**Then** (a) skip 설정에서는 리뷰어 호출 0, git 디렉터리 `moai/` 아래 `codex-review.lock`·`codex-review-wake.json`·`codex-review-gate.log` 가 생성되지 않는다. (b) 대조(`review` 설정)에서는 락 파일이 생성되고 리뷰어가 1 회 호출된다 — 이 대조 항은 오늘 락이 없어 빨강이다(L6).

### AC-011 — 프로브는 기록되고 결과가 등록을 게이트한다

**Given** M1 이후의 트리,
**When** `.moai/reports/t1422/async-probe/` 와 tracked `progress.md` 의「프로브 기록」표를 검사하면,
**Then** (a) 절차 문서(`.moai/reports/t1422/async-probes.md`)가 정한 아홉 개 프로브 ID — `E1-headless`·`E1-interactive`·`E2-timeout`·`E3-c1-concurrent`·`E3-c2-newturn`·`E3-c3-sessionexit`·`E4-d1-json`·`E4-d2-exit0stderr`·`E5-reload` — 각각에 대해 `.moai/reports/t1422/async-probe/<프로브 ID>.txt` 가 있고 그 안에 `claude --version` 출력, 사용한 설정 조각 이름, 사용한 플래그·환경 목록, 실행한 명령, 관측 원문, 결과 토큰 `확인됨`·`부정`·`불확정` 중 하나가 있으며, `progress.md`「프로브 기록」표에 각 파일의 sha256 이 적혀 있고 실제 파일 해시와 일치한다. E-1 은 `E1-interactive` 의 토큰이 정하고 `E1-headless` 는 보조 증거다. E-1 의 결과가 `확인됨` 이 아니면 두 설정 파일에 `asyncRewake` 가 없다. (b) **순서 항(사후 조건, 회귀 가드):** 해시를 기록한 커밋은 `.claude/settings.json` 에 `asyncRewake` 를 도입한 커밋의 조상이다(커밋 그래프로 관측 — `verification-claim-integrity.md` §2.3). 실행 주체는 progress.md 표에 적는다 — 레인이 헤드리스로 먼저 돌리는 것은 `E1-headless`·`E2-timeout`·`E3-c1-concurrent`·`E4-d1-json`·`E4-d2-exit0stderr`, 리더만 돌리는 것은 `E1-interactive`·`E3-c2-newturn`·`E3-c3-sessionexit`·`E5-reload` 이며, 결과가 올 때까지 후자는 `불확정` 이다.

### AC-012 — 비활성 트리의 세션은 프로세스도 깨움도 만들지 않고 등록은 마지막이다

**Given** `asyncRewake` 가 등록된 설정과 게이트를 켜지 않은 `workflow.yaml`(`review_gate` 키 부재 또는 `enabled: false` — 어느 쪽이든 같은 결과이며 이 AC 는 로컬 수정본의 값에 의존하지 않는다),
**When** 래퍼를 실행하면,
**Then** (a) 종료 코드 0, 가짜 `moai` 호출 수 0, stdout·stderr 비어 있음(R3 계열 회귀). (b) **순서 항(사후 조건, 회귀 가드):** 두 설정 파일을 바꾸는 커밋 뒤에는 M6 문서·progress 기록 외의 run-phase 커밋이 없다.

### AC-013 — 억제는 사유와 함께 기록된다 (git 디렉터리)

**Given** AC-004·AC-005·AC-014 의 시퀀스와 AC-003 의 비판정 종료 코드 2,
**When** 각 억제(락 경합, 알려진 상태, 연속 실패 상한 뒤 실패)와 래퍼의 비판정 2 가 일어나면,
**Then** (a) `<git 디렉터리>/moai/codex-review-gate.log` 에 이벤트 한 건당 JSON 한 줄이 추가되고 `reason` 은 각각 `lock-held`·`state-unchanged`·`failure-cap`·`non-verdict-exit-2` 이며 `ts` 와 `tree` 가 채워져 있고 상태 기반 사유에는 `state_key` 가 있다. 억제가 아닌 정상 리뷰(전달 포함, 세 번째 마지막 알림 포함)는 행을 쓰지 않는다. (b) 로그 파일은 작업 트리 아래에 생기지 않는다. (c) 락 기계 오류(git 디렉터리 쓰기 불가) 변형은 `lock-unavailable` 행을 쓰려 시도하고 리뷰는 진행된다(fail-open). (d) 상태 계산 오류·시간 초과 변형은 `state-unavailable` 행을 쓰고 리뷰는 진행된다.

### AC-014 — 연속 실패 깨움은 트리당 세 번까지이고 지문에 의존하지 않는다 (깨움 루프)

**Given** 상한 3과 **항상 실패하는** 리뷰어 스텁(호출마다 같거나 재서술한 findings, 또는 findings 0건에 서로 다른 `summary`), 세션 `s1`·`s2` 가 번갈아 오는 Stop, 매 Stop 이 새 실행기 호출(기록은 파일),
**When** 세 시퀀스를 처리하면 — **S-main**: 상태 A `fail`(s1) → B `fail`(s2) → C `fail`(s1) → D `fail` → E `fail`(재서술된 findings) → G `fail`(findings 0건, 다른 summary) → H `pass` → 같은 H → I `fail`; **S-reword**: 같은 이슈를 매번 다른 제목으로 재서술한 fail 세 번(서로 다른 새 상태); **S-empty**: findings 0건에 서로 다른 `summary` 인 fail 세 번(서로 다른 새 상태); **S-interleave**: A `fail` → X `inconclusive` → B' (다른 새 상태) `fail` → C' `fail`,
**Then** (a) S-main 의 A·B·C: 각각 종료 코드 2 (깨움 1·2·3 — 세션이 번갈아도 하나의 카운터). (b) S-reword·S-empty 의 세 번도 각각 깨움 1·2·3 이다 — 재서술·0건 findings 가 카운터를 리셋하지도 충돌시켜 앞당기지도 않는다. (c) 어떤 변형도 findings 문자열·`summary` 를 카운터에 쓰지 않는다(정적: 카운터 코드 경로에 findings 접근 없음은 요구하지 않고, (b) 의 행동 관측으로 고정한다). (d) D·E·G: 리뷰어는 호출되고(누적 호출 수 +1 씩) 종료 코드 0·빈 stdout·sentinel 없음·행 `failure-cap` — 상태가 달라도, findings 가 달라도, 0건이어도 침묵. (e) H `pass`: 종료 코드 0 침묵, 카운터 0. 같은 H: 호출 수 불변. I: 종료 코드 2 (깨움 1 — 카운터가 리셋됐다). (f) S-interleave: A 깨움 1 → X `inconclusive`(기록·카운터 불변) → B' 깨움 2 → C' 깨움 3 = 마지막 알림 — inconclusive 와 상태 변경이 카운터를 리셋하지 않는다.

### AC-015 — 세 번째 깨움이 마지막 알림이고 미해결을 밝힌다

**Given** AC-014 의 시퀀스와 최대 크기 findings(50건×5000자) 변형, stale 이면서 세 번째인 변형,
**When** 세 번째 연속 실패 깨움을 처리하면,
**Then** (a) 그 알림의 stderr 에 정확한 부분 문자열 `이후 알림 없음, 상태는 미해결` 이 **한 번** 있고 `FINAL NOTICE` 줄이 영어로 (i) 더는 알림이 없다 (ii) 상태가 UNRESOLVED 다 (iii) pass 가 알림을 다시 켠다 를 말한다; 전체 시퀀스에서 이 한국어 문구의 출현 수는 1 이고 위치는 세 번째 알림이다(첫·둘째·다섯째 알림에는 없다). (b) 알림은 여전히 sentinel 줄로 시작한다. (c) 최대 크기 변형에서도 요약은 8000바이트 이하이며 마지막 알림 두 줄은 잘리지 않는다. (d) stale+세 번째 결합 변형에서 `STALE` 줄과 마지막 알림 두 줄이 모두 있다.

---

## §E 품질 게이트와 DoD

- **DoD**: §D 전 AC 최종 상태 PASS(사후 조건 항은 회귀 가드로 별도 표기, AC-008 은 O-B 확정 뒤) · 행동 수준 RED 관측 기록이 `.moai/reports/t1422/red-async/` 에 구현 전 트리 SHA 와 함께 존재 · 프로브 아홉 개 기록과 sha256 (AC-011) · `go vet`(darwin)+`GOOS=windows GOARCH=amd64 go build ./...` 신규 0 · 수정 파일 커버리지 `quality.yaml` `test_coverage_target`(85) 이상 · 회귀 칸 R1-R5 가 변경 전에 초록으로 관측되고 변경 뒤에도 초록.
- **독립 단정 수(정직 표기)**: 헤딩 수 15 이지만 본문에 번호를 매긴 단정은 AC-001 9 · AC-002 3(열한 변형 × 3) · AC-003 7 · AC-004 9 · AC-005 9 · AC-006 6 · AC-007 5 · AC-008 2 · AC-009 4 · AC-010 2 · AC-011 2 · AC-012 2 · AC-013 4 · AC-014 6 · AC-015 4, 합 **74 개 번호 단정**(변형 수를 곱하면 훨씬 더 크다). "AC 15 개 통과"는 74 개 이상의 독립 점검을 뜻한다.
- **간접 검증**: AC-004 는 장벽으로 "기다림이 아닌 건너뜀"을, AC-005·014 는 상태 시퀀스로 "깨움 루프"를 양성 관측한다. 부재 주장(정적 0)은 양성 대조(P2-P4, P6, P7)와 함께만 읽는다.
- **선행 폐쇄 게이트**: M1 회귀 칸(R1-R5)이 변경 전 트리에서 초록으로 관측되지 않으면 이후 마일스톤을 시작하지 않는다. 프로브 게이트(spec.md §D)를 건너뛰고 M2 이후를 시작하지 않는다 — 단 **리더가 Gap 수용을 기록한 항목은 예외**(plan §D 처리 규칙과 같은 문장).
- **측정 규율**: 소관 패키지 단위로 재측정한다. `-run` 이름 패턴 단독 재측정 금지(0개를 골라도 `ok`). 전체 스위트 로컬 금지.
- **선택(변이 실측)**: M1 이 §C 변이 점검의 몇 개(락 cwd 키링, 기록 쓰기 전 해제, delivered 선기록, 래퍼의 모든 2 전달)를 실제로 심어 해당 단정이 빨갛게 되는 것을 `red-async/` 에 남기면 "죽임" 칸이 작성 시점 판독에서 실행 관측으로 격상된다.
