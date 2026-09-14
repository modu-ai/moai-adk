# SPEC-MOAI-GATEWAY-001 — t654 AS-5 동기화 감사 보고서

- 카드: t654 (AS-5 launcher 통합) · SPEC 버전 0.13.0
- 감사 대상 트리: worktree `.claude/worktrees/t654`, branch `WT-gateway-launchers`, HEAD `ce4a3c6ed` (sync 커밋)
- 측정 범위 병합 기준점: `6de8dd489` (= origin/develop d416f8162 + develop 0f5dbe096)
- 감사자: sync-auditor (독립 판정, 레인 산출물과 분리) · 측정일 2026-09-14
- 판정 방식: 4차원(Functionality/Security/Craft/Consistency) 조화평균, must-pass 방화벽(Functionality·Security) 적용

**종합 판정: PASS (조화평균 0.93 ≥ 0.85) — 차단 finding 0건, 선택 finding 5건.**
SPEC은 이 창에서 닫히지 않는다(in-progress 유지, 라이브 창 카드 t844·t851 잔존) — 본 판정은 "이 창의 착지분이 감사 기준을 충족하는가"에 대한 것이며 창 대기 항목을 PASS로 세지 않는다.

---

## 1. Claim (주장)

1. AC-MG-026 (b)(c)의 기계 판정 셀렉터가 GREEN이고, 그 RED→GREEN 로그가 보관돼 있으며 본문 인용과 일치한다.
2. 사전존재 환경 실패(codexbridge lifecycle_subprocess 4건 + gateway `TestAppServerSubprocessHTTPToolContinuation`)는 이 diff 없는 기준 트리에서 동일 재현되는 환경 의존 실패다.
3. t708 비침범 잠금의 소유 시리즈 면제(`059f4e700`)는 잠금 본래 목적(수리 카드의 무단 약화 차단)을 유지한다 — SPEC 터치 없는 침범 diff는 여전히 적색이다.
4. 보안 3축이 성립한다 — 구독→API 자동전환 구조적 불가(GPT catalog PKCE-only), 바인딩 경로의 구독 토큰 저장소 접근 0, GLM 이미지 입력 명시 400+upstream 0.
5. 동기화 표면(§E 신호, frontmatter, CHANGELOG 유예, 증거 파일 목록, 창 대기 6건)이 자기일치한다.

## 2. Evidence (명령 전문 + 관측 출력, 이번 실행)

**E-1. AC-MG-026 (b) — 셀렉터 재실행 (sweep 수 1 확인 포함)**
```
$ go test ./internal/gateway/receipt/ -run 'TestRebaseLedgerRestoration' -count=1 -v
=== RUN   TestRebaseLedgerRestoration
--- PASS: TestRebaseLedgerRestoration (0.06s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/receipt	0.220s
```
RED 로그 대조: `as5-m2-epoch-red.log` = `--- FAIL: TestRebaseLedgerRestoration (0.06s)` / `restoration_test.go:55: restored ledger applied epoch = 0, want the recorded 3` — §E.2 인용과 일치.

**E-2. AC-MG-026 (c) — 셀렉터 재실행**
```
$ go test ./internal/gateway/conversation/ -run 'TestForkPrefixCrossCheck' -count=1 -v
=== RUN   TestForkPrefixCrossCheck
--- PASS: TestForkPrefixCrossCheck (0.10s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/conversation	0.265s
```
RED 로그 대조: `as5-m2-forkprefix-red.log` = `fork_prefix_test.go:63: tampered prefix accepted: <nil>` — 일치. 대조 본문 검증: `Manager.ForkSession`(family.go:320)이 `receipt.ChainDigest(chain) != claimed`(family.go:339)으로 거절하고 거절 시 자식 상태 잔존 0을 단언(fork_prefix_test.go:66·78). engine 경계는 `TestForkPrefixAuthorityGatesForkAcceptance`(fork_test.go:119)가 권한 배선 상태에서 변조 기각(`ErrScope`, `rpc.next == 0`)과 일치 수용을 양방향 고정.

**E-3. AC-MG-026 (a) 게이트 리터럴 — 대조군 유지 상태**
```
$ grep -rn 'awaiting transport verification' internal/cli --include='*.go' | grep -v _test
internal/cli/launcher.go:142:		return errors.New("GPT gateway launch is awaiting transport verification; use moai gpt status to inspect login")
internal/cli/gpt.go:54:			return errors.New("GPT gateway launch is awaiting transport verification; use moai gpt status to inspect login")
```
2건 — (a) GREEN(0건)이 아닌 창 대기 상태 그대로이며, §E.3은 이를 PASS가 아닌 window_waiting으로 기록해 정직하다. 대조군 시험 `TestGatewayLaunchTransportGateControl`(gateway_launch_assembly_test.go:170)이 리터럴 수 1×2곳·binding-less 기동 거절·`errCGRetired` 비영향을 고정한다.

**E-4. t708 가드 면제 — 셀렉터 재실행 + 본문 판독**
```
$ go test ./internal/cli/ -run 'TestGatewayRepairCardDiffTouchesNoPreservedFile|TestPreservedDiffDiscriminatorRejectsViolations|TestPreservedDiffExemptionOwningSeriesTouch' -count=1 -v
=== RUN   TestGatewayRepairCardDiffTouchesNoPreservedFile
--- PASS: TestGatewayRepairCardDiffTouchesNoPreservedFile (0.21s)
=== RUN   TestPreservedDiffDiscriminatorRejectsViolations
--- PASS: TestPreservedDiffDiscriminatorRejectsViolations (0.00s)
=== RUN   TestPreservedDiffExemptionOwningSeriesTouch
--- PASS: TestPreservedDiffExemptionOwningSeriesTouch (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.088s
```
본문 판독(gateway_preserve_test.go): SPEC 터치 없는 합성 침범 diff → 위반 3건 적색(118-143), 위반+소유시리즈 → 면제, 위반+타시리즈(`SPEC-OTHER-001`) → 위반 2건 적색(149-168), 빈 diff → "unmeasurable, not clean" 거절(108-110). 면제 한계(위장 면제는 plan-audit/run 게이트 표면에서 걸린다)는 코드 주석 21-25행에 명시돼 있다. **잠금 본래 목적 유지 판정.**

**E-5. 보안 3축 — 시험 본문 판독 + 재실행**
```
$ go test ./internal/gateway/ -run 'TestGatewaySubscriptionFailureNeverSwitchesToAPI' -count=1 -v
=== RUN   TestGatewaySubscriptionFailureNeverSwitchesToAPI/valid_subscription_credential_reaches_the_upstream
=== RUN   TestGatewaySubscriptionFailureNeverSwitchesToAPI/absent_subscription_credential_is_a_refusal_with_zero_upstream_requests
--- PASS: TestGatewaySubscriptionFailureNeverSwitchesToAPI (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.237s

$ go test ./internal/cli/ -run 'TestGatewayAuthModesDualPath' -count=1 -v
=== RUN   TestGatewayAuthModesDualPath
--- PASS: TestGatewayAuthModesDualPath (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.852s
```
- 자동전환 구조적 불가: GPT catalog 전수(4행, 전제 단언 `len(models) != 4` → 비공허)가 `AuthPKCE`만 선언(cli/gateway_auth_modes_test.go:29-37) + 구독 실패가 401·upstream 0(양성 대조군 200·calls 1 포함, gateway/gateway_auth_modes_test.go:31-32).
- 토큰 저장소 접근 0: 격리 `MOAI_HOME`에서 바인딩 생성 뒤 `home/gateway-auth` 부재 단언 + 격리 home 실효성 양성 대조군(`home/state/gateway-conversations` 존재, cli/gateway_auth_modes_test.go:40-58).
- GLM 이미지 400+upstream 0: `requestCarriesImageInput` 차단이 credential 해석·upstream 송신 이전(server.go:110-115)이라 계수 0이 구조적으로 보장되고, 텍스트 양성·Claude 이미지 통과 대조군이 같은 시험에 있다(context_paths_test.go:38-62). 생산 선언도 GLM `Images:false`·200K / Claude `Images:true`·1M으로 재판정(gateway_product_binding.go, 시험 gateway_launch_assembly_test.go:131-162).
- M1 exec 경계: 신설 조립 표면(`gatewayAuthDisplay` 스위치, `gatewayOutputPolicyDisplay` 상수)은 카탈로그 enum에서 파생된 고정 문자열만 운반한다 — 외부 입력이 exec/command 경계로 유입되는 경로 없음. 표시↔child env 일치 교차단언(`authDisplayMatchesSessionEnv`)이 경계 양측을 함께 읽는다.

**E-6. 사전존재 실패 귀속 — 병합 기준점 재현 (`git archive` 추출, /tmp)**
```
$ git archive 6de8dd489 | tar -x -C /tmp/t654-base-verify
$ cd /tmp/t654-base-verify && go test ./internal/codexbridge/ -run 'TestAuditRealTransportEOFWakesBridge' -count=1
--- FAIL: TestAuditRealTransportEOFWakesBridge (0.00s)
    lifecycle_subprocess_test.go:42: app server start failed
FAIL	github.com/modu-ai/moai-adk/internal/codexbridge	0.371s

$ cd /tmp/t654-base-verify && go test ./internal/gateway/ -run 'TestAppServerSubprocessHTTPToolContinuation' -count=1
    appserver_integration_test.go:71: app server start failed
FAIL	github.com/modu-ai/moai-adk/internal/gateway	0.435s
```
두 실패 계열 모두 이 diff 없는 `6de8dd489` 트리에서 동일 적색 — 사전존재 환경 의존 귀속 성립. 현행 트리에서도 `TestAppServerSubprocessHTTPToolContinuation`이 본 감사 실행에서 적색 재현(환경 flaky — §E.2 기록과 모순 없음, F5).

**E-7. Craft 실측**
```
$ golangci-lint run ./internal/cli/... ./internal/gateway/... ./internal/codexbridge/... --timeout=5m
0 issues.
$ go vet ./internal/cli/ ./internal/gateway/... ./internal/codexbridge/   → exit 0
$ gofmt -l <변경 16 .go 파일>                                            → 빈 출력
$ go test -cover ./internal/gateway/receipt/ ./internal/gateway/conversation/ -count=1
ok .../internal/gateway/receipt	1.099s	coverage: 88.1% of statements
ok .../internal/gateway/conversation	1.843s	coverage: 79.4% of statements
$ go test -coverpkg=./internal/gateway/... ./internal/gateway/... (함수단위 집계)
mean-total 92.1%  ·  신규 경로: ForkSession 88.9%, receipt 신규 함수 대부분 100%(패키지 로컬)
```
DoD 신규 코드 경로 85%+: 신규 착지 경로 기준 충족(conversation 79.4%는 사전존재 표면 포함 패키지 전체 수치). 최종 스위트 로그(`as5-final-suite.log`) 판독: 가드 적색 1건(410행, `059f4e700`으로 수리) + codexbridge 사전존재 4건(1155-1162행) — 배차 지시의 "guard 1 red + 4 pre-existing"과 일치.

**E-8. 동기화 표면 일관성**
```
$ grep -c 'SPEC-MOAI-GATEWAY-001' CHANGELOG.md
0   (exit 1)  → B12 유예 전제(grep 0) 성립
$ head spec.md frontmatter → version "0.13.0" / status in-progress / updated 2026-09-14
→ §E.4 "frontmatter_status_transitions: NONE this window"와 일치
$ ls .moai/reports/t654/as5-*.md *.log → §E.2가 인용한 as5-* 경로 전수 존재
$ grep -ln "@MX:" (주장된 5개 .go) → 5파일 전수 적중, diff 추가 @MX 행 6건
```
창 대기 6건(§E.3 5건 + §E.4의 M2b 생산 배선 보강 1건) 전수 확인. plan A5-M1~M7 대비: M1~M4 자동화 착지, M5 실증 준비(`as5-pty-automation.md`), M6 로컬 준비(`as5-windows-ci-prep.md`), M7 미실행 — M7은 강제 전제 {(a),(b),(c),AS-017,AS-019,AS-021} 미충족이라 **실행하지 않은 것이 옳은 처분**이다.

## 3. Baseline-attribution (귀속)

- 모든 E-1~E-8 측정은 이번 감사 실행에서, worktree `.claude/worktrees/t654` HEAD `ce4a3c6ed` 트리에 대해 직접 수행했다. 인용한 보관 로그(RED/GREEN, 최종 스위트)는 파일 판독으로 대조했고 재측정으로 대체하지 않았다.
- 병합 기준점 재현(E-6)은 `git archive 6de8dd489` 추출 트리(/tmp/t654-base-verify)에서 수행 — stash·브랜치 전환·워킹 트리 변경 없음.
- 레인이 기록한 cli 전체 스위트 1차 `ok 1306.951s`와 최종 스위트는 재실행하지 않았다(레인 로컬 규율 + 배차 지시 6번) — FAIL 목록은 로그 직접 판독으로만 검증.

## 4. Gaps (이번 감사에서 관측하지 않은 것)

- 실계정·실PTY 실증(AS-014·017·018·019·021), AS-022 GitHub CI 실행 증거, AC-MG-026 (a) GREEN(리터럴 제거), (d) 배포 게이트 — 창 대기 설계상 이번 창에서 관측 불가. 본 감사는 이들을 판정하지 않는다.
- codexbridge lifecycle_subprocess 4건 중 병합기준점 직접 재현은 1건(`TestAuditRealTransportEOFWakesBridge`) — 나머지 3건은 final-suite 로그 판독으로만 확인.
- `GOOS=windows` 빌드·`-race`·cli 전체 스위트는 재실행하지 않았다(M6 기록과 final-suite 로그 인용).
- §E.2 M3 GREEN 행의 "사전존재 2하위시험 skip"은 비verbose 보관 로그에서 관측 불가(F2).
- `as5-m4-authmodes-green.log`는 이름된 2테스트 중 gateway 측 1개만 담는다(F3).

## 5. Residual-risk (관측에도 남는 위험)

- subprocess 기반 테스트 5계열은 부하·환경 의존 flaky — 같은 트리에서 ok와 FAIL이 혼재 관측됐다(§4 F5). 통합 브랜치 CI에서 재현 가능성이 남는다.
- "토큰 저장소 접근 0" 단언은 생성 부재(Stat) 관측이지 읽기 계수 0이 아니며, 실제 보장은 구조 설계(child-owned store + send-time adapter 해석)가 운반한다.
- 이미지 탐지는 Messages 블록 형태 스캔이라 파싱 불가 형태에 fail-open — §E.2가 인정한 잔여위험 그대로.
- engine `ForkPrefix` 훅은 현재 배선 0(F4) — 생산 codexbridge engine 생성지가 처음 생기는 카드에서 배선이 누락되면 engine 레벨 caller-asserted 수용이 부활한다.
- manifest v1 경험본의 v2 복원 epoch 0 — dev 단계 무피해(§E.2 기록과 동일), 실세션 흐름(t844)에서 재판정.

---

## 차원 점수

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 95/100 | PASS | E-1·E-2 (b)(c) GREEN 재현+sweep 1 확인, E-3 (a) 대조군 정직한 창 대기, E-6 사전존재 귀속, M1~M4 plan 대응 착지 |
| Security (25%) | 94/100 | PASS | E-5 3축 전부 양성 대조군 동반 비공허 단언, M1 exec 경계 외부 입력 0, upstream-0 구조 보장(차단이 credential 해석 전) |
| Craft (20%) | 90/100 | PASS | E-7 lint 0·vet 0·gofmt clean·신규 경로 ≥85%(ForkSession 88.9%), 단 증거-본문 수치 표류 F1~F3 감점 |
| Consistency (15%) | 93/100 | PASS | E-8 frontmatter 무전이·CHANGELOG grep 0·증거 전수 존재·창 대기 6건·MX 5파일, 단 F1·F4 표류 감점 |

**조화평균: 4/(1/95 + 1/94 + 1/90 + 1/93) = 0.93 → PASS (기준 0.85).**
Must-pass 방화벽: Functionality PASS · Security PASS — 강제 FAIL 없음.

## Findings (structured defect-list)

- F1 [Low] [optional] `.moai/specs/SPEC-MOAI-GATEWAY-001/progress.md` §E.2 t654 M3 GREEN 행 - 본문이 `ok 7.440s`를 인용하나 보관 로그 `as5-m3-context-green.log`는 `ok 6.914s` — 인용 수치와 보관 증거의 표류(VCI §2). - Required fix: 본문 수치를 보관 로그 값으로 정정하거나 해당 실행의 로그를 추가 보관.
- F2 [Low] [optional] progress.md §E.2 같은 행 - "사전존재 2하위시험 skip" 주장이 비verbose 로그에서 관측 불가(skip은 무음). - Required fix: `-v` 실행 로그를 보관하거나 관측 불가 문구를 삭제.
- F3 [Low] [optional] `.moai/reports/t654/as5-m4-authmodes-green.log` - §E.2가 이름한 2테스트 중 gateway 측 1개만 담고 cli 측 `TestGatewayAuthModesDualPath` 로그가 없다(본 감사 재실행으로 양측 PASS 확인). - Required fix: cli 실행 로그를 보관.
- F4 [Low] [optional] `internal/codexbridge/engine.go:39` - "The gateway wires the receipt-backed authority here" 주석이 아직 존재하지 않는 배선을 현재형으로 서술(생산 `codexbridge.Config{` 생성지 0, `ForkPrefix` 세터는 테스트뿐). AC (c) 본체는 `Manager.ForkSession` 대조로 성립. - Required fix: 주석을 의도 표현으로 정정하고, engine 생산 생성지가 처음 생기는 카드에서 `Config.ForkPrefix` 배선을 AC 항목으로 명시.
- F5 [Info] [optional] `internal/gateway/appserver_integration_test.go` `TestAppServerSubprocessHTTPToolContinuation` - 같은 트리에서 ok(M3 GREEN·최종 스위트)와 FAIL(본 감사)이 혼재하는 부하 의존 flakiness. 귀속 자체는 E-6으로 성립. - Required fix: lifecycle_subprocess 계열 환경 감지·부하 격리 정리 카드에서 일괄 처분.

차단(blocking) finding: **0건**. 전부 optional — finding-소비 규율에 따라 자동 수리 라우팅 대상이 아니며 다음 sync(라이브 창 카드) 때 일괄 정정으로 충분하다.

## Recommendations

- F1~F3의 증거 위생 정정은 다음 sync 커밋에 progress.md 행 정정 + 로그 보충 한 번으로 묶어 처리할 것.
- t844(실세션 흐름) 카드에 codexbridge engine 생성 시 `ForkPrefix` 배선 의무를 넘길 것 — 그렇지 않으면 AC (c)의 engine층 방어가 테스트 전용으로 남는다.
- 창 대기 6건의 해제 순서는 §E.4 기록대로 — 배포 게이트(M7)는 강제 전제 충족 전 절대 선실행하지 말 것(이번 창의 미실행이 옳은 처분이었다).
