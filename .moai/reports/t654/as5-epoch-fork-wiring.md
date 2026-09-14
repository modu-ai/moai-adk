# t654 A5-M2 — compaction epoch 생산 복원 + fork prefix 원장 대조 증거 (AC-MG-026 (b)(c))

## Claim

(b) `Store.Rebase`가 마지막 적용 epoch를 같은 manifest 트랜잭션에 기록(diskManifest v2
`applied_epoch`)하고, 재시작 세션이 `Store.RestoreRebaseLedger`로 기록값을 정확히 복원한다 — 기록
없음은 0, 판독 불가·훼손은 명시 오류. (c) `Manager.ForkSession`이 `--fork-session` 자식 배리어를
`Manifest.ChainTo(boundary)` 완료 체인의 `receipt.ChainDigest`와 대조해 일치할 때만 수용하며,
변조·불일치·미지 원본은 자식 상태 생성 전에 명시 거절한다. codexbridge Engine은 gateway가 주입한
`ForkPrefix` 권한을 호출 경계에서 상담한다.

## Evidence

### (b) TestRebaseLedgerRestoration — RED (고정값 주입 현행 구현)

명령: `go test ./internal/gateway/receipt/ -run 'TestRebaseLedgerRestoration' -count=1`
(스텁 단계: 시그니처만 도입, epoch 기록 없음 — 시그니처 도입 전에는 컴파일 불능이 RED의 1차 형태)
기준 트리: `6de8dd489`. 출력:

```
--- FAIL: TestRebaseLedgerRestoration (0.06s)
    restoration_test.go:55: restored ledger applied epoch = 0, want the recorded 3
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway/receipt	0.595s
```

로그: `.moai/reports/t654/as5-m2-epoch-red.log`

### (b) GREEN + 기계 판정 명령

명령: `go test ./internal/gateway/... -run 'TestForkPrefixCrossCheck|TestRebaseLedgerRestoration' -count=1`
기준 트리: 커밋 `aec09a5d6` 워킹 트리. 출력:

```
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.742s [no tests to run]
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	0.364s [no tests to run]
ok  	github.com/modu-ai/moai-adk/internal/gateway/conversation	1.502s
ok  	github.com/modu-ai/moai-adk/internal/gateway/opaque	2.007s [no tests to run]
ok  	github.com/modu-ai/moai-adk/internal/gateway/receipt	1.817s
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	1.390s [no tests to run]
```

receipt 패키지 전량(`go test ./internal/gateway/receipt/ -count=1`)도 `ok`(기존
`TestManifestRebaseResetsPublicHistory`·`TestStoreRebaseResetsDurableHistory`는 시그니처 적응 후
원 단언 유지). 로그: `.moai/reports/t654/as5-m2-epoch-green.log`, `as5-m2-abc-green.log`

### (c) TestForkPrefixCrossCheck — RED (caller-asserted 수용이 변조 변형에서 적색)

명령: `go test ./internal/gateway/conversation/ -run 'TestForkPrefixCrossCheck' -count=1`
(스텁: ForkSession이 claimed를 무시하고 ForkAt 위임 — ChainDigest도 고정값 스텁)
기준 트리: `6de8dd489`. 출력:

```
--- FAIL: TestForkPrefixCrossCheck (0.16s)
    fork_prefix_test.go:63: tampered prefix accepted: <nil>
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway/conversation	0.529s
```

로그: `.moai/reports/t654/as5-m2-forkprefix-red.log`. GREEN 시험은 거절 후
`families/<family>/forks` 디렉터리 계수로 자식 상태 미생성까지 단언한다.

### codexbridge 호출 경계

`TestForkPrefixAuthorityGatesForkAcceptance`(rejects-tampered-claim / accepts-matching-claim) —
명령: `go test ./internal/codexbridge/ -count=1 -skip 'TestAuditRealTransportEOF...4건'` →
`ok  github.com/modu-ai/moai-adk/internal/codexbridge 2.500s`.

## Baseline-attribution

모든 출력은 2026-09-14, worktree `.claude/worktrees/t654`, base `6de8dd489`에서 직접 실행.
구현 커밋: `aec09a5d6`.

## Gaps

- (b)의 생산 호출자(어댑터가 세션을 다시 열 때 RestoreRebaseLedger를 부르는 지점)는 현재 트리에
  존재하지 않는다(`NewRebaseLedger` 비-테스트 호출자 0건이 t653 잔여 위험 그대로) — 본 마일스톤은
  착지 시접(기록·판독 API)과 기계 판정 시험이다. 생산 배선은 compaction 흐름이 실세션(t844)에서
  열릴 때 이어진다.
- 실세션 rebase 뒤 재시작 복원의 실측(AS-012 실제 Claude 압축 수집)은 t844 소관.

## Residual-risk

- manifest v1→v2 전환: v1 manifest는 읽기 지원(applied_epoch 0으로 복원), 쓰기는 항상 v2 — 컴팩션이
  있었던 v1 대화는 재개 시 epoch 0으로 복원된다(재rebase 가능). 이 트리의 대화 상태는 dev 단계라
  실 피해 경로 없음.
- codexbridge 4건(`lifecycle_subprocess_test.go`)은 본 트리 baseline에서도 실패하는 환경 의존
  하위프로세스 시험("app server start failed") — baseline 귀속 확인: server.go가 아닌
  `internal/codexapp` 시작 경로, python shim 실행 환경 문제. 내 변경과 무관임을 baseline 재실행으로
  확인했다.
