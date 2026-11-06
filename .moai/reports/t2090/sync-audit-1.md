auditor-model: claude-opus-5-5
verdict: PASS
audited_sha: bc80f85b3fc4a8b7849f265e655702126fa21293
overall_score: 0.92
must_pass_failed: 0
blocking_count: 0
blocking_defects: none
receipts: rcpt-2c16553b8507e6e5fc84a857
spec: t2090 (Class B 카드라 SPEC이 없다. 판정 줄의 spec= 값은 `internal/auditreceipt/store.go:142`의 cardIDBody 문법 `(?:card-)?t[0-9]+`에 맞춰 카드 id를 쓴다. 규약 문서 `.moai/docs/audit-artifact-convention.md`에는 SPEC 없는 카드의 spec 필드 규칙이 따로 없어서 판정 줄 파서 문법을 따랐다)

# t2090 sync 감사 1회차 — TakeOver가 Issue 잠금 안에서 소유자 생존을 다시 확인하는가

- 카드: t2090 (Class B, SPEC 없음). 대상: `Runner.TakeOver`가 `OwnerLiveness`를 Issue 잠금 밖에서만 읽던 TOCTOU (t2073 sync 감사 F2, 재현 P5)
- 감사 커밋: HEAD `bc80f85b3fc4a8b7849f265e655702126fa21293` (브랜치 WT-agenticfactory-runner-ru), 기준 `9b632c1eb3544cef5f275e502922a375e48d20aa` (= HEAD~1). 변경은 이 커밋 하나, `internal/agenticfactory/` 아래 5개 파일
- 판정: **PASS** (조화평균 0.92). must-pass(Functionality, Security) 모두 통과. 차단 결함 0건
- codex 단일 감사 1회 (운영자 결정 d-20261106T055816Z-449d, 이후 d-20261106T061529Z-5d7e로 대체됨. 대체 규칙은 카드당 최대 2회를 허용하고 리더 지시는 1회였으므로 두 규칙 모두를 만족하도록 1회만 호출함)

## 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 95 | PASS | 기준 커밋 export에서 새 시험이 RED(live·indeterminate가 `err=<nil>`), 수리 커밋에서 GREEN. 잠금 안 `Refusal` 호출을 지운 변이에서 다시 RED |
| Security (25%) | 93 | PASS | 소유권 검사(Verify holder/epoch, CheckSafeStop, 중복 dispatch) 순서·내용 그대로. 잘못 넘어가던 방향(live 소유자 인계)이 막힘. 약해진 검사 없음 |
| Craft (20%) | 90 | PASS | claim 85.8%, runner 92.6%. vet·gofmt·golangci-lint 깨끗, `-race` 통과 |
| Consistency (15%) | 92 | PASS | 오류 문구를 `OwnerState.Refusal` 하나로 모아 runner와 store가 같은 문장을 냄. 기존 시험 도우미(`handedOffOwner`/`beat`/`fullTakeover`)와 `beforeClaimMoves` seam 재사용 |

계산: 4 / (1/0.95 + 1/0.93 + 1/0.90 + 1/0.92) = 4 / 4.3279 = 0.924

## 백엔드별 결과

- codex: verdict=pass, findings=0, audit_receipt=`rcpt-2c16553b8507e6e5fc84a857`, resolved_model=`gpt-6.1-sol`, requested_effort=high, mode=adversarial, target=baseBranch, review_base=`main (merge base 9b632c1eb3544cef5f275e502922a375e48d20aa)`, build_commit=`646955327e`, transport=`codex app-server`(MCP 서버 내부 경로이며 감사관이 app-server를 직접 부르지 않음)
- 감사 계획(`moai verify audit-plan --project-root <toplevel>`): config_status=ok, model=codex, gates claude=off / codex=required / glm=off, cross_model_active=true, enforced_required=[codex]. 계획상 audit_multi 경로지만 운영자 결정이 audit_multi를 금지하므로 codex_audit 단일 호출로 대신함. codex가 유일한 required 게이트이고 pass를 냈으므로 required 게이트 미충족 없음 (Gaps G1 참고)

## 발견 사항

차단(blocking) 0건. 수리에서 생긴 결함(repair-born) 해당 없음(1회차).

- F1 [Low][optional][touched-file] `internal/agenticfactory/runner/runner.go:773-774` 주석 "the store checks holder and epoch only"가 이제 사실과 다르다. 이 커밋 뒤로 store는 unverified 경로에서 잠금 안에서 presence도 다시 읽는다. 동작에는 영향 없음. Required fix(선택): 주석을 "the store re-checks holder, epoch and the owner's recorded process under the lock"처럼 고친다. 신뢰도 높음
- F2 [Info][optional][in-diff] 첫 확인에서 presence가 없어 note에 "owner gone (caller-asserted, no process on record)"를 붙였는데, 그 사이 기록된 presence가 잠금 안에서 dead로 읽히면 인계는 진행되고 note 문구는 실제와 조금 어긋난다. 안전 방향(진행해도 되는 경우)이라 판정 영향 없음. 추론만 함, 재현하지 않음. 신뢰도 중간
- F3 [Low][optional][in-diff] 새 시험은 잠금 안 재확인의 live·indeterminate·dead만 다루고, 잠금 안에서 presence 파일이 손상된(Damaged) 경우는 다루지 않는다. 같은 `Refusal` 함수의 Damaged 분기는 첫 확인 경로 시험 `TestADamagedPresenceBlocksATakeoverAndNamesTheRecovery`(takeover_presence_test.go:106)가 실행하므로 분기 자체는 시험된다. Required fix(선택): 필요하면 `beforeClaimMoves`에서 손상 파일을 쓰는 하위 시험 하나를 더한다
- F4 [Info][optional][outside] 다른 잠금 밖 생존 판정 지점을 찾아봤다. `claim/lock.go:94,132`(잠금 재발급)는 rename 뒤 다시 분류하는 자체 재확인이 이미 있다. `claim/stoprecord.go:109`의 `readPresence`는 `CheckSafeStop` 안이고 `takeOver` 잠금 구간에서 불린다. `monitor/flow/flow.go:34`, `monitor/adapt/prober.go:61`, `identity/identity.go:199`, `factorymsg/factory_run_retire.go:99`는 Issue 인계와 무관한 관측·은퇴 경로다. runner.go 안의 `OwnerLiveness` 호출은 753 한 곳뿐이고, 이번 수리로 잠금 안 재확인이 뒤를 받친다. adopt/reopen 경로(runner.go)에는 생존 판정 호출이 없다. 남은 F2 류 결함 없음

## 확인한 것 (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)

### Claim 1 — 잠금 안 재확인은 unverified 경로에서만 돌고 교착이 없다

Evidence: `git show HEAD`의 claim/next.go 변경은 `if strings.TrimSpace(safeStop) != "" { CheckSafeStop } else { OwnerLiveness → Refusal }` 구조로, safe-stop 경로는 손대지 않았다. `withLock`(claim/claim.go:423-460)은 `O_CREATE|O_EXCL` 잠금 파일이라 재진입하지 않는다. `OwnerLiveness`(presence.go:131-161)는 `s.read`(claim.go:463, `os.ReadFile`만)와 `s.readPresence`(presence.go:109, `os.ReadFile`만)와 `ClassifyOwnerWith`(probe 호출)만 부르고 `withLock`을 부르지 않는다. 같은 잠금 안에서 `Verify`를 부르는 `Heartbeat`(presence.go:88-89)가 이미 있어 `Verify`도 잠금을 잡지 않음이 확인된다. `-race` 실행:

```
$ go test -count=1 -race ./internal/agenticfactory/claim/ ./internal/agenticfactory/runner/
exit=0
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/claim	52.827s
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/runner	29.274s
```

### Claim 2 — 오류 문구와 sentinel이 그대로이고 검사 순서도 그대로다

Evidence: diff에서 runner.go에서 지운 세 case와 presence.go `Refusal`의 세 case가 문자열 단위로 같다(ErrHeld "the owner's process %d is still running", ErrOwnerUnconfirmed "the presence file %s cannot be read: once the owner's process is confirmed gone, remove that file (the owner's next Heartbeat writes a new one)", ErrOwnerUnconfirmed "the owner's process %d cannot be told running or gone"). 순서 Live → Damaged → Indeterminate도 같다(Damaged는 Class=Indeterminate를 같이 가지므로 Damaged가 먼저여야 하고, 그대로다). `st.HasPresence`(gone) → note, `ev.OwnerGone` → caller-asserted note, 둘 다 없으면 ErrEvidenceIncomplete 순서도 남아 있다. `takeOver`의 `Verify` → (CheckSafeStop | 재확인) → 중복 dispatch → 갱신 순서에서 빠지거나 옮겨진 검사 없음. 잠금 안 `OwnerLiveness`는 `Verify` 직후라 `!ok`(ErrNotHeld) 분기에 닿지 않는다.

### Claim 3 — 결함 재현(RED)과 수리(GREEN), 변이 검출

Baseline: `git archive HEAD~1`을 새 빈 임시 디렉터리에 풀고 HEAD의 `runner/takeover_presence_test.go`만 덮어씀.

```
$ go -C <base export> test -count=1 -run TestTakeoverRechecksTheOwnerUnderTheLockWhenItBeatsAfterTheFirstLook -v ./internal/agenticfactory/runner/
exit=1
    takeover_presence_test.go:167: an owner that beat after the first look: err=<nil> holder="sess-new"   (a_live_process)
    takeover_presence_test.go:167: an owner that beat after the first look: err=<nil> holder="sess-new"   (an_indeterminate_process)
    --- FAIL: .../a_live_process (0.16s)
    --- FAIL: .../an_indeterminate_process (0.18s)
    --- PASS: .../a_process_that_is_gone (0.16s)
```

변이: `git archive HEAD`를 다른 빈 임시 디렉터리에 풀고 next.go의 잠금 안 `if err := st.Refusal(); err != nil { return err }`를 `_ = st`로 바꿈.

```
$ go -C <mutant export> test -count=1 -run TestTakeoverRechecksTheOwnerUnderTheLockWhenItBeatsAfterTheFirstLook -v ./internal/agenticfactory/runner/
exit=1
    --- FAIL: .../a_live_process (0.18s)
    --- FAIL: .../an_indeterminate_process (0.17s)
    --- PASS: .../a_process_that_is_gone (0.15s)
```

수리 트리(HEAD)에서는 같은 시험이 전체 패키지 실행 안에서 통과(Claim 5).

### Claim 4 — 바뀐 기존 시험은 회귀를 숨기지 않는다

`TestAnEarlierHoldersBeatDoesNotStaleTheNextHoldersStop`(claim/stoprecord_test.go:119)은 sess-a가 live probe 상태로 Heartbeat한 뒤 `TakeOverUnverified`로 인계했다. 즉 살아 있는 소유자를 넘기는, 이번 카드가 막으려는 동작 위에 서 있던 fixture다. 확인 실행: HEAD export에 기준 커밋의 stoprecord_test.go를 덮어씀.

```
$ go -C <fixture export> test -count=1 -run "TestAnEarlierHoldersBeatDoesNotStaleTheNextHoldersStop|TestAnOlderHoldersHeartbeatDoesNotDescribeTheNewHolder" -v ./internal/agenticfactory/claim/
exit=1
--- PASS: TestAnOlderHoldersHeartbeatDoesNotDescribeTheNewHolder (0.19s)
    stoprecord_test.go:132: agenticfactory: issue is owned by another session: the owner's process 4242 is still running
--- FAIL: TestAnEarlierHoldersBeatDoesNotStaleTheNextHoldersStop (0.03s)
```

수정은 인계 직전에만 probe를 dead로 두고 직후 live로 되돌린다. 시험이 보려는 단언("앞선 소유자의 presence 파일이 다음 소유자의 stop을 낡게 만들지 않는다")은 live probe 아래에서 그대로 검사되고, presence 파일도 남아 있다. fixture 정정이지 회귀 은폐가 아니다.

호출자 전수(`grep -rn -e TakeOverUnverified -e "\.TakeOver(" -e TakeOverOnStop --include="*.go" internal cmd pkg`): 운영(비시험) 호출자는 `internal/agenticfactory/runner/runner.go:775`(`TakeOverUnverified`)과 `:777`(`TakeOverOnStop`) 두 곳, 모두 `Runner.TakeOver` 안이다. `Runner.TakeOver`의 운영 호출자는 없다(나머지는 모두 `_test.go`). 행동이 달라지는 운영 호출자는 `Runner.TakeOver` 자신뿐이고, 달라지는 것은 첫 확인과 이동 사이에 live/판정 불가 presence가 생긴 경우 거부하는 것 하나다. 다른 시험 호출자(stage_test.go:304,308, presence_test.go:107, takeover_stop_test.go:81)는 전체 실행에서 통과한다.

### Claim 5 — 트리 전체 검증

```
$ go test -count=1 ./internal/agenticfactory/...
exit=0   (ok 21개 패키지: browser, checkpoint, claim 64.757s, core, ctxmodel, e2e 26.949s, effect, effect/internal/procrun, identity, intake, ledger, ledger/importer, measure, mergeq, monitor, monitor/adapt, monitor/flow, reconcile, retirement, review, runner 41.599s)
빈 실행 토큰("no test files"/"no tests to run") 0건, FAIL/SKIP 0건

$ go test -count=1 -cover ./internal/agenticfactory/claim/ ./internal/agenticfactory/runner/
exit=0
ok  	.../claim	74.565s	coverage: 85.8% of statements
ok  	.../runner	18.454s	coverage: 92.6% of statements

$ go vet ./internal/agenticfactory/...          exit=0, 출력 없음
$ gofmt -l internal/agenticfactory              exit=0, 출력 없음
$ golangci-lint run ./internal/agenticfactory/claim/... ./internal/agenticfactory/runner/...   exit=0, "0 issues."
$ GOOS=windows go build ./...                   exit=0, 출력 없음
```

Baseline-attribution: 모든 명령은 이 실행에서 트리 HEAD `bc80f85b3f`(dirty 없음, `git status --short` 빈 출력을 감사 시작·codex 호출 전·후에 확인)에서 돌렸다. RED·변이·fixture 확인은 `git archive`로 만든 별도 export 세 개에서만 돌렸고 추적 트리는 바꾸지 않았다. Go 시험은 트리 소스를 컴파일하므로 판정 빌드는 트리 자체다. codex 백엔드를 실은 moai MCP 서버 빌드는 `646955327e`이며 `git merge-base --is-ancestor 646955327e HEAD`는 exit=1(HEAD의 조상이 아님)이다.

## Gaps

- G1: 감사 계획은 cross_model_active=true라 `audit_multi` 경로를 가리키지만, 운영자 결정이 audit_multi·claude_audit·glm을 금지해서 `codex_audit` adversarial 1회로 대신했다. claude·glm 게이트는 설정상 off라 빠진 required 백엔드는 없다. `audit_multi` 결과 다이제스트(overall_verdict/gate_unmet/plan_source)는 만들어지지 않았으므로 `moai verify audit-plan --result-file` 확인 대상도 없다
- G2: codex 결과의 summary는 "Verdict: pass" 한 줄뿐이고 근거 서술이 없다. codex가 무엇을 읽었는지는 review_base와 receipt로만 확인된다
- G3: 실제 프로세스(가짜 probe가 아닌 실제 PID)로 하는 경합 재현은 하지 않았다. 새 시험은 `beforeClaimMoves` seam과 가짜 probe를 쓴다
- G4: Linux·Windows에서 시험을 실행하지 않았다(darwin에서만 실행, Windows는 cross-build만)
- G5: 레인 progress.md(`.moai/reports/t2090/progress.md`)의 주장은 증거로 쓰지 않았다
- 거부된 명령 없음

## Residual-risk

- 잠금 안 재확인은 `Runner.TakeOver`의 첫 확인이 presence 없음 또는 gone을 본 경우를 막아 준다. 그러나 presence를 아예 쓰지 않는 소유자(Heartbeat를 하지 않는 소유자)에 대해서는 여전히 호출자의 `OwnerGone` 주장만으로 인계된다. 설계상 의도된 동작(문서 주석에 명시)이며 카드 범위 밖이다
- t2073 F3/F4(Linux `ps` 지문 경합·TZ 차이)는 그대로다. 지문이 틀리게 dead로 나오면 잠금 안 재확인도 dead로 읽는다
- 모든 Go 시험 출력과 임시 export는 세션 scratch(`/private/tmp/claude-501/.../scratchpad/t2090a/`)에 남아 있고 다른 기계로 가지 않는다. 판정에 쓴 줄은 이 파일에 옮겼다

## 반복 이력

- 1회차(이 파일): HEAD bc80f85b3f, PASS 0.92, codex pass `rcpt-2c16553b8507e6e5fc84a857`
