auditor-model: claude-opus-5-5
verdict: PASS
audited_sha: 9b5f391f1627991dde24f9eec32eb997a6d92fdb
score: 0.92
blocking_count: 0
scope: card
card: t2073
base_sha: 9b632c1eb3544cef5f275e502922a375e48d20aa
iteration: 1
cross_model: none (by instruction)

# t2073 sync 감사 1회차 (독립 감사, Class B, SPEC 없음)

## 결론

카드의 주장("정상 인계에서 TakeOver가 ErrOwnerUnconfirmed로 막힌다"는 보고는 재현되지 않음)은 이 감사의 측정으로도 반박되지 않았다. 실제 프로세스를 띄워 정상 종료·좀비·연쇄 인계를 직접 돌려 봤고, 정상 인계 경로 가운데 ErrOwnerUnconfirmed를 내는 경로는 찾지 못했다. 변경은 시험 파일 하나뿐이고, 새 시험은 OwnerLiveness의 holder/epoch 대조를 끈 변이를 runner 수준에서 잡아낸다. 판정은 PASS, 점수 0.92, 차단 결함 0건이다.

다만 카드 범위 밖에서 이미 있던 경쟁 조건 하나를 실행으로 재현했다(F2). 이 카드가 만든 결함은 아니고, 카드 질문과는 반대 방향(막히는 쪽이 아니라 잘못 넘어가는 쪽)이라서 차단으로 분류하지 않았다. 리더가 후속 카드 발행을 판단할 사항이다.

## 차원별 점수 (조화평균 0.92)

| 차원 | 점수 | 판정 | 근거 한 줄 |
|---|---|---|---|
| Functionality (40%) | 0.92 | PASS | 반박 주장을 실제 프로세스 실행 4건으로 재확인함. 근거 문서의 문장 하나가 실제 동작보다 좁게 쓰여 있음(F1) |
| Security (25%) | 0.90 | PASS | 운영 코드는 바뀌지 않았고 소유권 검사도 약해지지 않음. 범위 밖 기존 TOCTOU 하나를 재현함(F2) |
| Craft (20%) | 0.92 | PASS | 새 시험이 변이를 잡고, 패키지 커버리지 92.7%, vet·lint·gofmt 모두 깨끗함 |
| Consistency (15%) | 0.95 | PASS | 같은 파일의 기존 도우미(`handedOffOwner`/`beat`/`notGone`/`fullTakeover`)와 이름 규칙을 그대로 따름 |

계산: 4 / (1/0.92 + 1/0.90 + 1/0.92 + 1/0.95) = 4 / 4.3377 = 0.922. must-pass 차원(Functionality, Security)은 둘 다 통과.

## 발견 사항

- F1 [Low][optional][outside — 근거 문서] `.moai/reports/t2073/progress.md:11`의 "both only when the CURRENT holder recorded a presence file"는 실제 동작보다 좁다. `internal/agenticfactory/claim/presence.go:129-133`은 파일이 손상되었는지를 holder/epoch 대조(`presence.go:135`)보다 먼저 판정한다. 그래서 **이전 holder가 남긴 손상 파일**도 다음 인계를 ErrOwnerUnconfirmed로 막는다. 실행으로 재현함(아래 P4). 손상된 파일은 누구 것인지 읽을 수 없으니 fail-closed로 막는 쪽이 타당하고, 손상은 정상 인계 경로에 들지 않는다. 그래서 카드 결론은 그대로 둔다. Required fix: progress.md 문장을 "presence 파일이 손상되었거나(작성자 무관), 현재 holder가 기록한 presence의 프로세스가 판정 불가일 때"로 고친다. 고치지 않아도 판정은 바뀌지 않는다.
- F2 [Medium][optional — 범위 밖, 재현 증거 있음, 리더 판단 요청] `internal/agenticfactory/runner/runner.go:754`(`OwnerLiveness`는 Issue 잠금 밖에서 읽음)와 `runner.go:779`(`TakeOverUnverified`는 잠금 안에서 holder/epoch만 확인) 사이에 TOCTOU가 있다. presence가 없는 소유자가 그 틈에 Heartbeat를 쓰면, 살아 있는 소유자인데도 호출자 주장만으로 인계가 진행된다. 재현함(P5: `race: err=<nil> holder="sess-x"`). 조건이 붙는다. 호출자가 OwnerGone과 WriterStopped를 거짓으로 주장해야 하고, 이 커밋에는 `Runner.TakeOver`나 `claim.Store.Heartbeat`의 운영 호출자가 없다(아래 grep). 이 카드가 운영 코드를 바꾸지 않았으니 카드에서 생긴 결함이 아니고, 방향도 "막힘"이 아니라 "잘못 넘어감"이다. 그래서 이 카드에서는 차단으로 두지 않았다. 규칙상 재현 증거가 있는 발견은 위치와 상관없이 차단 후보가 될 수 있으므로 이 분류를 리더에게 명시해 넘긴다. Required fix(후속 카드 권고): 잠금이 없는 경로도 safe-stop 경로(`TakeOverOnStop`)처럼 `takeOver` 잠금 구간 안에서 presence를 다시 읽거나, 잠금 안에서 "결정 시점 presence 상태"와 비교한다.
- F3 [Low][optional][outside, 추론만 함] Linux `process_fingerprint_unix.go`: `kill(0)`이 live를 돌려준 뒤 `ps -o lstart=`가 끝나기 전에 프로세스가 죽으면 `ps`가 0이 아닌 값으로 끝나 `ProcessIdentityIndeterminate`가 되고, 그러면 ErrOwnerUnconfirmed가 난다. 종료 순간에만 생기는 일시적 경합이라 재시도하면 dead로 풀린다. darwin 호스트라 재현하지 못함.
- F4 [Low][optional][outside, 추론만 함] Linux `ps -o lstart`의 출력은 프로세스의 TZ/로캘을 따른다. 기록한 쪽과 조회한 쪽의 TZ가 다르면 같은 살아 있는 프로세스의 지문이 서로 달라 `OwnerDead`로 분류되고, 살아 있는 소유자를 넘겨줄 수 있다(카드 질문과 반대 방향). 재현하지 못함.
- F5 [Info][optional] 좀비(종료했지만 회수되지 않은 프로세스) 소유자는 회수될 때까지 ErrHeld로 막힌다. ErrOwnerUnconfirmed는 아니다. 실행으로 확인함(P2). 작성자 주장과 맞음.
- F6 [Info][optional][추론만 함] Windows `pid_state_windows.go`: 종료한 프로세스라도 누군가 핸들을 쥐고 있으면 `OpenProcess`가 성공하고 생성 시각이 같아 Live, 즉 ErrHeld가 된다. 다른 사용자의 프로세스(ACCESS_DENIED)는 Indeterminate, 즉 ErrOwnerUnconfirmed다. 같은 사용자의 정상 인계에서는 해당하지 않는다. Windows에서는 실행하지 않음.

차단(blocking) 결함 0건. 수리에서 생긴 결함(repair-born)도 해당 없음(1회차).

## 주장별 증거 (Claim / Evidence / Baseline-attribution)

### C1. 변경은 시험 파일 하나뿐이고 운영 코드는 바뀌지 않았다

```
$ git rev-parse HEAD
9b5f391f1627991dde24f9eec32eb997a6d92fdb
$ git diff --stat 9b632c1eb3 HEAD
 .../runner/takeover_presence_test.go               | 32 ++++++++++++++++++++++
 1 file changed, 32 insertions(+)
$ git log --oneline 9b632c1eb3..HEAD
9b5f391f16 [t2073] test: pin that an earlier holder's presence does not block the next TakeOver
```

기준점: 이 트리, base `9b632c1eb3544cef5f275e502922a375e48d20aa`, 감사 시작과 파일 작성 직전에 HEAD를 두 번 읽었고 값이 같았다.

### C2. 지정된 시험과 race 시험이 통과한다

```
$ go test ./internal/agenticfactory/runner -run 'TestAnEarlierHoldersPresence|TestTakeover|TestADamagedPresence|TestARunningOwner' -count=1 -v
exit=0 (38 lines)
--- PASS: TestTakeoverReadsTheRecordedProcessInsteadOfTheCallersWord (0.86s)
--- PASS: TestAnEarlierHoldersPresenceDoesNotBlockTheNextHandover (0.22s)
--- PASS: TestARunningOwnerIsNotTakenOverOnAnInventedStopReference (0.17s)
--- PASS: TestADamagedPresenceBlocksATakeoverAndNamesTheRecovery (0.15s)
--- PASS: TestARunningOwnerThatBeatsAfterItsStopIsNotTakenOver (0.17s)
--- PASS: TestTakeoverNeedsTheOwnerGoneTheWriterStoppedAndTheWorkPreserved (0.24s)
--- PASS: TestTakeoverHandsTheIssueToTheNewOwnerAndKeepsTheWork (0.23s)
--- PASS: TestTakeoverRefusesAStaleDecision (0.16s)
--- PASS: TestTakeoverStartsANewAttemptAndKeepsTheOldOneInTheHistory (0.15s)
--- PASS: TestTakeoverAcceptsASafeStopInsteadOfOwnerGone (0.13s)
--- PASS: TestTakeoverRefusesASafeStopThatIsNotARecordOfThisOwner (0.44s)
--- PASS: TestTakeoverRefusesAStopTheOwnerHasActedAfter (0.22s)
--- PASS: TestTakeoverRefusesAStopTheOwnerActsAfterBeforeTheClaimMoves (0.18s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/runner	3.726s

$ go test -race -count=1 ./internal/agenticfactory/runner
exit=0
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/runner	28.925s
```

### C3. 새 시험이 holder/epoch 대조를 끈 변이를 잡는다

방법: `git archive --format=tar -o <scratchpad>/t2073-head.tar HEAD`로 HEAD를 떠서 세션 스크래치 하위 디렉터리에 풀고, 그 사본의 `internal/agenticfactory/claim/presence.go:135`를 `if false && (p.Holder != rec.Holder || p.Epoch != rec.Epoch) { // MUTANT`로 바꿨다. 이 트리의 파일은 건드리지 않았다.

```
$ go test ./internal/agenticfactory/runner -run 'TestAnEarlierHoldersPresenceDoesNotBlockTheNextHandover' -count=1   (변이 사본)
--- FAIL: TestAnEarlierHoldersPresenceDoesNotBlockTheNextHandover (0.13s)
    takeover_presence_test.go:112: second handover blocked by the earlier holder's presence: agenticfactory: the Issue's current owner is not confirmed handed over or absent: the owner's process 4242 cannot be told running or gone
FAIL
```

새 시험만 빼고 돌렸을 때(같은 변이):

```
$ go test ./internal/agenticfactory/runner ./internal/agenticfactory/claim -skip 'TestAnEarlierHoldersPresenceDoesNotBlockTheNextHandover' -count=1
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/runner	40.856s
--- FAIL: TestAnOlderHoldersHeartbeatDoesNotDescribeTheNewHolder (0.07s)
--- FAIL: TestAnotherHoldersPresenceFileIsAbsentNotDamaged (0.01s)
FAIL	github.com/modu-ai/moai-adk/internal/agenticfactory/claim	74.044s
```

해석: store 수준에서는 claim 패키지 시험 두 개가 이미 이 변이를 잡는다. runner 패키지 안에서는 새 시험이 이 변이를 잡는 유일한 시험이다. 그러니 새 시험은 중복이 아니라 TakeOver 경로를 고정하는 시험이다. 확인이 끝난 뒤 변이를 되돌렸고, 스크래치 사본·tar·커버리지 파일은 모두 지웠다.

### C4. 정상 인계 경로에서 ErrOwnerUnconfirmed가 나는지 직접 돌려 봤다 (실행 재현)

변이 없는 스크래치 사본에 임시 시험 파일 `zz_audit_probe_test.go`를 넣고, 기본 플랫폼 탐침(`homestate.ProbeProcessIdentity`)과 실제 `sleep 30` 자식 프로세스로 실행했다(darwin).

```
--- PASS: TestAnEarlierHoldersPresenceDoesNotBlockTheNextHandover (0.37s)
=== RUN   TestAuditProbeRealOwnerExitedCleanly
    zz_audit_probe_test.go:52: note="from sess-a, preserved at refs/heads/WT-agentic-x@preserve-1; owner process 16736 is gone"
--- PASS: TestAuditProbeRealOwnerExitedCleanly (0.29s)
=== RUN   TestAuditProbeZombieOwner
    zz_audit_probe_test.go:66: zombie takeover err=agenticfactory: issue is owned by another session: the owner's process 16787 is still running unconfirmed=false held=true
--- PASS: TestAuditProbeZombieOwner (0.64s)
=== RUN   TestAuditProbeChainedRealHandover
--- PASS: TestAuditProbeChainedRealHandover (0.38s)
=== RUN   TestAuditProbeDamagedEarlierHolderPresence
    zz_audit_probe_test.go:116: damaged earlier-holder presence: err=agenticfactory: the Issue's current owner is not confirmed handed over or absent: the presence file /var/folders/.../issue-1948.presence cannot be read: once the owner's process is confirmed gone, remove that file (the owner's next Heartbeat writes a new one) unconfirmed=true
--- PASS: TestAuditProbeDamagedEarlierHolderPresence (0.30s)
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/runner	2.346s
```

| 후보 | 방법 | 관측 |
|---|---|---|
| P1 실제 소유자가 Heartbeat 뒤 정상 종료·회수됨 | 실행 | 살아 있을 때 ErrHeld, 종료 뒤 인계 성공("owner process … is gone") |
| P2 좀비 소유자 | 실행 | ErrHeld (ErrOwnerUnconfirmed 아님) |
| P3 연쇄 인계(두 번째 holder도 실제 Heartbeat 뒤 종료) + 인계 뒤 이전 holder의 Heartbeat | 실행 | 두 인계 모두 성공, 이전 holder의 Heartbeat는 거부됨(epoch 불일치로 다른 holder 이름의 presence를 쓸 수 없음) |
| P4 이전 holder의 presence 파일이 손상됨 | 실행 | ErrOwnerUnconfirmed (F1, 비정상 경로, fail-closed) |
| P5 Heartbeat가 TakeOver와 경합 | 실행(`beforeClaimMoves` 시험 seam) | `race: err=<nil> holder="sess-x"` (F2, 막힘이 아니라 넘어감) |
| Linux 종료 순간 ps 경합 | 추론만 | 일시적 ErrOwnerUnconfirmed 가능(F3) |
| Windows | 추론만 | 같은 사용자의 정상 인계에서는 해당 없음(F6) |
| Start 지문이 비어 있음 | 코드 읽기 | `Heartbeat`(`presence.go:68`)가 빈 start를 거부하므로 그런 presence는 정상 경로로 생기지 않음. 외부에서 쓴 경우 `readPresence`가 Damaged로 처리 |

### C5. 운영 호출자가 없다 (F2의 노출 범위)

```
$ grep -rn "\.Heartbeat(\|OwnerLiveness(\|\.TakeOver(" internal cmd --include='*.go' | grep -v "_test.go"
internal/agenticfactory/runner/runner.go:754:		st, err := r.Store.OwnerLiveness(i)
internal/agenticfactory/claim/presence.go:114:func (s *Store) OwnerLiveness(i core.IssueRef) (OwnerState, error) {
internal/cli/session.go:99:			if err := session.Heartbeat(sessionID); err != nil {
internal/hook/session_heartbeat.go:50:	if err := session.NewRegistry(path, nil).Heartbeat(input.SessionID); err != nil {
internal/hook/ledger_heartbeat.go:45:		return st.Heartbeat(ctx, cardID, epoch, nil)
internal/session/registry.go:250:	return defaultRegistry().Heartbeat(sessionID)
```

나머지 Heartbeat 호출은 다른 저장소(session registry, ledger)의 것이다. `Runner.TakeOver`와 `claim.Store.Heartbeat`의 운영 호출자는 이 커밋에 없다.

### C6. 품질 도구 (Craft / Consistency)

```
$ go test -count=1 -coverprofile=<scratch>/t2073-runner.cov ./internal/agenticfactory/runner
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/runner	44.160s	coverage: 92.7% of statements
$ go tool cover -func=<scratch>/t2073-runner.cov | grep -E "TakeOver|total"
github.com/modu-ai/moai-adk/internal/agenticfactory/runner/runner.go:733:	TakeOver		91.7%
total:										(statements)		92.7%
$ gofmt -l internal/agenticfactory/runner/        (출력 없음)
$ go vet ./internal/agenticfactory/runner          vet=0
$ golangci-lint run ./internal/agenticfactory/runner/...
0 issues.
$ grep -c "^func Test[A-Z][A-Za-z]*(t \*testing.T)" internal/agenticfactory/runner/takeover_presence_test.go
5
```

### C7. 보안 (Security)

```
$ grep -rnE "(api[_-]?key|secret|token|password)\s*[:=]" internal/agenticfactory/runner/takeover_presence_test.go
(일치 없음)
```

운영 코드가 바뀌지 않았으니 입력 검증·주입 경계도 그대로다. `TakeOver`의 Indeterminate/Damaged 분기(`runner.go:761-764`)도 그대로 남아 있다. 의존성 manifest(go.mod/go.sum)도 바뀌지 않았다(C1).

## 작성자 근거 파일 대조

- `progress.md:11` — F1. 손상 파일은 작성자와 상관없이 막는다. 실측(P4)과 문장이 다르다. Low.
- `progress.md:13` 좀비 → live/ErrHeld — 실측(P2)과 맞음.
- `progress.md:15` 변이 결과 문구 — 이 감사가 따로 돌린 변이(C3)에서 같은 오류 문구가 나옴. 맞음.
- `progress.md:22` race `ok 27.963s` — 이 감사의 실행은 `ok 28.925s`. 결과가 같음.
- `progress.md:40` "start HEAD 69b10035f" — `git merge-base --is-ancestor 69b10035f 9b632c1eb3`가 exit 0으로 끝나므로 base의 조상이다. 모순 아님.
- `card-review.md` — codex_review는 advisory이고 receipt가 아니라고 스스로 적어 두었다. 이 감사도 그렇게 본다.

## 반복 이력

1회차. 이전 감사 없음.

## Gaps (관측하지 않은 것)

- 교차 모델 감사를 돌리지 않았다(지시에 따름). `audit_multi`·`codex_audit`·`glm_audit`·`claude_audit`을 호출하지 않았으니 audit receipt가 없다. 이 트리의 `.moai/config/sections/workflow.yaml:38`은 `codex: required`이므로 그 게이트는 이 감사로 채워지지 않는다. 리더가 따로 처리해야 한다. 같은 이유로 `moai verify audit-plan`도 돌리지 않았다.
- Linux와 Windows에서는 실행하지 않았다. F3·F4·F6은 코드를 읽고 추론한 것이다.
- 원래 보고(961e5b0c41 리뷰)의 실제 재현 절차를 받지 못했다. 그래서 "보고자가 본 바로 그 경로"를 재현했다고는 말할 수 없다. 반박이 성립하는 범위는 위에서 나열한 후보들까지다.
- 전체 저장소 시험(`go test ./...`)은 지시에 따라 돌리지 않았다.
- 절차 기록: 4번 단계에서 시험 출력을 `.moai/reports/t2073/audit-targeted.log`·`audit-race.log`로 리디렉션했다(38줄·소량). 그런데 "트리 안에 파일은 정확히 하나" 제약과 겹쳐서, 필요한 출력을 이 파일로 옮긴 뒤 두 로그를 지웠다. `git status --short`에는 아무것도 나오지 않았다. 원래 출력은 위 C2에 그대로 옮겨 적었다.
- 처음 쓴 스크래치 명령 일부(복합 명령, `rm -rf`)는 worktree 격리 가드나 권한 거부로 실행되지 않았다. 단순 명령으로 나눠 다시 실행했고, 스크래치 정리는 `rm -r` 단일 명령으로 마쳤다.

## 잔여 위험

- F2의 TOCTOU는 운영 호출자가 연결되는 순간 실제로 노출된다. 호출자가 OwnerGone을 관측 없이 주장하면 살아 있는 소유자를 넘겨줄 수 있다. 연결 전에 후속 카드로 다루기를 권한다.
- Linux 지문이 TZ/로캘에 따라 달라지는 문제(F4)가 실제로 생기면 살아 있는 소유자가 dead로 분류된다. 이 카드의 시험은 가짜 탐침을 쓰므로 이 경로를 잡지 못한다.
- 손상된 presence 파일의 복구는 운영자가 파일을 지우는 수동 절차에 기대고 있다. 오류 문구에 그 안내가 들어 있다(P4 출력에서 확인).
