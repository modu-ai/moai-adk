# M4 보고 — 온디스크 산출물 2종 (SPEC-FEEDBACK-AUTO-SUBMIT-001)

- 대상: `internal/feedback/masklog.go` · `internal/feedback/queue.go` + 테스트 2종 (+ `scrub.go` 배선 1줄)
- 기준 트리: base `3210da7d3`, M3 HEAD `3bcceffc7`, 브랜치 `WT-auto-feedback`
- 대응 AC: AC-F-015 · AC-F-016 · AC-F-017 · AC-F-018

## 1. Claim (주장)

1. 마스킹 로그(`.moai/logs/feedback-mask.log`)가 RFC3339 시각·종류·위치·건수를 기록하고, **원문 값을 담지 않으며**, `0o600` 으로 생성된다. (AC-F-015)
2. 로그를 쓸 수 없는 상태에서도 `Scrub` 이 에러 없이 정상 `Result` 를 반환한다 — 로깅은 fail-open. (AC-F-016)
3. 전송 실패한 보고가 `.moai/state/feedback/queue.json` 에 1건으로 적재되고, 적재된 본문·제목은 **마스킹된** 것이다. (AC-F-017)
4. 이후 성공을 표시하면 항목이 큐에서 **제거**되고 파일은 유효한 JSON 으로 남는다. (AC-F-018)
5. 큐의 읽기 범위는 `queue.json` 뿐이며, 스크럽 이전 원문을 담는 `feedback-draft-<ts>.md` 를 큐 항목으로 읽지 않는다. (D4 경계)
6. 큐는 append-only JSONL 이 아니라 잠금 있는 단일 JSON 이며, 잠금은 read-modify-write 전체를 덮는다. (AP-7)

## 2. Evidence (증거 — 실행한 명령 + 관측 출력)

RED (구현 이전, 테스트만 있는 트리):

```
$ go test ./internal/feedback/ -run 'TestMaskLog|TestQueue' -v
# github.com/modu-ai/moai-adk/internal/feedback [github.com/modu-ai/moai-adk/internal/feedback.test]
internal/feedback/queue_test.go:15:40: undefined: QueueStore
internal/feedback/queue_test.go:22:15: undefined: NewQueueStore
internal/feedback/queue_test.go:22:29: undefined: QueuePathForRoot
internal/feedback/queue_test.go:54:10: undefined: QueuePathForRoot
internal/feedback/queue_test.go:90:41: undefined: queueFilePerm
internal/feedback/queue_test.go:91:46: undefined: queueFilePerm
internal/feedback/masklog_test.go:46:10: undefined: MaskLogPathForRoot
internal/feedback/masklog_test.go:84:39: undefined: maskLogPerm
internal/feedback/masklog_test.go:85:48: undefined: maskLogPerm
internal/feedback/masklog_test.go:127:23: undefined: MaskLogPathForRoot
internal/feedback/queue_test.go:91:46: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/feedback [build failed]
FAIL
```

GREEN (AC별 명령, 각 1회):

```
$ go test ./internal/feedback/ -run 'TestMaskLogRecordsKindAndCountWithoutRawValue' -v
    masklog_test.go:58: mask log entry: 2026-08-23T18:37:28+09:00 | total=1 | kind=secret where=body count=1
--- PASS: TestMaskLogRecordsKindAndCountWithoutRawValue (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.568s

$ go test ./internal/feedback/ -run 'TestMaskLogFailureIsFailOpen' -v
2026/08/23 18:37:09 WARN feedback: cannot create mask log directory dir=/var/folders/.../001/.moai/logs error="mkdir /var/folders/.../001/.moai/logs: not a directory"
--- PASS: TestMaskLogFailureIsFailOpen (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.405s

$ go test ./internal/feedback/ -run 'TestQueueEnqueuesOnSendFailure' -v
--- PASS: TestQueueEnqueuesOnSendFailure (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.411s

$ go test ./internal/feedback/ -run 'TestQueueResolvesOnSuccess' -v
--- PASS: TestQueueResolvesOnSuccess (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.422s

$ go test ./internal/feedback/ -run 'TestMaskLog' -v
--- PASS: TestMaskLogRequiresProjectRoot (0.00s)
--- PASS: TestMaskLogSkipsCleanScrub (0.00s)
--- PASS: TestMaskLogRecordsKindAndCountWithoutRawValue (0.00s)
--- PASS: TestMaskLogFailureIsFailOpen (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.414s
```

경계 3건(AC 외 추가, 같은 실행에서 PASS): `TestQueueRefusesBlockedResult` · `TestQueueNeverReadsPreScrubDraft` · `TestQueueMutateSerializesConcurrentEnqueues`.

회귀·게이트:

```
$ go test -count=1 ./internal/feedback/... ./internal/sandbox/...
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.693s
ok  	github.com/modu-ai/moai-adk/internal/sandbox	0.267s

$ go test -race -count=1 ./internal/feedback/
ok  	github.com/modu-ai/moai-adk/internal/feedback	1.890s

$ go vet ./internal/feedback/... ./internal/sandbox/...                 # 무출력, exit 0
$ GOOS=windows go vet ./internal/feedback/... ./internal/sandbox/...    # 무출력, exit 0
$ golangci-lint run --timeout=3m ./internal/feedback/... ./internal/sandbox/...
0 issues.
$ gofmt -l internal/feedback/                                          # 무출력
```

## 3. Baseline-attribution (baseline 귀속)

- 측정 트리: 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t170`, 브랜치 `WT-auto-feedback`, M3 HEAD `3bcceffc7` 위의 미커밋 작업본. 위 모든 출력은 **이 트리, 이 회차**에서 관측했다.
- 부재 baseline: `git ls-tree -r 3210da7d3 --name-only -- internal/feedback/masklog.go internal/feedback/queue.go | wc -l` → `0`. 두 파일은 base 에 없었고, `git ls-tree -r 3bcceffc7 --name-only -- internal/feedback/` 출력 8줄에도 없다 — RED 의 `undefined:` 목록은 실재하는 부재의 결과이지 공허한 검사가 아니다.
- 트리 오염 baseline: `git status --short` → `M internal/feedback/scrub.go` + 신규 4파일. 실제 리포의 `.moai/logs/feedback-mask.log` 는 생성되지 않았다(테스트가 전부 `t.TempDir()` 루트를 명시 전달).

## 4. Gaps (미검증)

- **다중 프로세스 잠금 경합을 관측하지 않았다.** 직렬화는 한 프로세스 안의 4 goroutine × 3 회로만 관측했다. 경합 상한(25ms × 40 ≈ 1s)이 실제 다중 세션에서 충분한지는 미측정.
- **Windows 에서 아무 테스트도 실행하지 않았다.** `GOOS=windows go vet` 은 **컴파일만** 증명한다 — `atomicfile.Claim`/`Replace` 의 Windows 경로와 권한 skip 분기는 CI 매트릭스가 판정한다.
- **`Resolve` 에 프로덕션 호출자가 없다.** M5 의 큐 동사가 붙기 전까지 테스트가 유일한 소비자다. 같은 의미에서 `EnqueueMasked` 도 아직 CLI 에서 호출되지 않는다 — M4 는 형식 계약만 고정한다.
- **로그 회전·상한을 두지 않았다.** 무한 append 이며 크기 상한이 없다. 범위 밖이나 후속 카드 후보.
- **stale lock 정리 경로가 없다.** 프로세스가 비정상 종료하면 `queue.lock` 이 남고(POSIX flock 과 달리 자동 해제되지 않는다), 이후 모든 mutation 이 ~1s 후 에러로 끝난다. 수동 삭제가 유일한 복구다.
- `make build`·전체 스위트는 돌리지 않았다(M9 소관, CLAUDE.local.md §4 로컬 전 스위트 금지).

## 5. Residual-risk (잔여 위험)

- **로그 인터리빙.** 잠금이 없으므로 두 스크럽이 동시에 쓰면 한 줄이 섞일 수 있다. 의도된 트레이드오프(잠금이 스크럽을 막는 쪽이 더 나쁘다)이며, 훼손 결과는 판독 불가 한 줄이지 값 유출이 아니다.
- **fail-open 의 이면.** 로그가 조용히 비어 있어도 스크럽은 성공한다 — "로그가 없다"를 "마스킹이 없었다"로 읽으면 안 된다. `slog.Warn` 이 유일한 신호다.
- **D4 오인의 대가는 여전히 크다.** 코드·주석·테스트로 경계를 고정했으나, 재전송을 구현할 M5/M6 가 초안 파일을 글롭하면 스크럽 이전 원문이 공개 이슈로 나간다. 그 지점의 리뷰가 필요하다.
- **`Options.ProjectRoot` 계약 변경.** "빈 값 = 상향 탐색"에서 "빈 값 = 산출물 없음"으로 바꿨다. M5 가 `--root`/`ResolveProjectRoot` 폴백을 배선하지 않으면 **로그도 큐도 조용히 쓰이지 않는다** — 기능 부재가 에러로 드러나지 않는 형태이므로 M5 에 스모크 관측이 필요하다.
