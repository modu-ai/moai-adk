# t36 — 절대 시각 임계값 테스트 3건, 머신 부하가 아니라 코드를 측정하게 하기 (+t2 흡수)

- worktree: `.claude/worktrees/t36` (branch `WT-t36`, base = origin/main + `origin/release/v3.1.1` tip `091a42e16` merge, base commit `367fa6374`)
- date: 2026-08-17
- lane-local: push 금지 — 리드 허브 리뷰 PASS 후 자가 통합 예정

## 0. 카드 전제 검증 (처방보다 먼저)

| 카드 명시 테스트 | 실측 결과 | 처방 대상 |
|---|---|---|
| `internal/hook TestBranchGuard_Latency` (t2) | 존재. #1541 예산 비율 형태(p95 ≤ 예산 20%, worst < 100%) | ✅ 적용 |
| `internal/harness TestRecordEvent100Sequential` | 존재. 같은 예산 비율 형태 | ✅ 적용 |
| `internal/statusline TestBuilderNormalizesMode` | **시간 임계 단언이 한 번도 존재한 적 없음** | ❌ 제외(전제 정정) |

전제 정정 근거(TestBuilderNormalizesMode):
- `git log --all -S "TestBuilderNormalizesMode" -- internal/statusline/builder_test.go` → 생성 커밋 2개(#466 추가, 초기)뿐.
- `git log -S "time.Since" -- internal/statusline/builder_test.go` → **0건**. 현재 파일에도 wall-clock 단언 부재(패키지 전체 grep: TTL/backoff/timeout *의미* 테스트만 존재).
- 이 테스트의 과거 풀 스위트 플레이크는 #1467 — `New()`가 실제 usage collector를 만들어 keychain+OAuth를 deadline 없이 호출하는 행. 현재 코드의 주석(builder_test.go:422-427)과 mockUsageProvider 주입 + usage 3s 바운드(#1552, 6ff7c3aaf)로 이미 해결. 타이밍 처방 대상 아님.

## 1. 카드 핵심 질문 — verify snapshot에 얹을 수 있는가? → **불가 (코드로 입증)**

`internal/verify/key.go:32` `Key()`는 HEAD SHA + `git status --porcelain=v2` digest + `git diff HEAD` 내용 해시의 3입력으로 키를 만든다. 따라서:

1. **정확한 트리 상태에 바인딩** — 커밋 1개가 올라가면(심지어 dirty 파일 1개 편집만으로도) 키가 바뀌어 이전 baseline에 도달 불가. `observer_test.go` 구주석도 같은 결론("a previous-commit baseline is unreachable by construction").
2. **시점 불일치 = 부하 민감성 재래** — 다른 시점에 기록된 baseline은 다른 부하 상태를 측정한 것. idle 머신 baseline vs 부하 걸린 실행 비교는 이 카드가 제거하려는 결함(머신 부하 측정)을 원래대로 되돌린다.
3. **CI 휘발성** — 스냅샷은 projectRoot `.moai/state/verify/snapshots/`(gitignore 런타임 상태). ephemeral 러너에선 매번 first-run 폴백이라 캘리브레이션 팔이 작동하지 않는다.

**대안 처방(채택)**: "기록된 baseline 대비 diff"의 정직한 형태는 **같은 실행 안(in-run)에서 같은 비용 계층의 기준 연산을 측정해 비교**하는 것 — 기계·부하·CI 환경이 동일한 조건에서 잡은 baseline끼리만 비교가 성립한다. 비율(측정/기준)은 코드 속성(기계 비용 유닛을 몇 배 소모하는가)이 된다.

## 2. 처방 — 공통 헬퍼 `internal/timing` (신규 패키지, 3개 팔)

`timing.Assert(t, Bound{...}, refUnit, fn)` — fn을 warmup 제외 N회 측정 후 3개 단언:

| 팔 | 단언 | 잡는 것 |
|---|---|---|
| 분포 | `p95 ≤ SteadyCeiling`(예산의 20%) | 분포 전체를 늦추는 회귀(네트워크 호출, 리포지터리 스캔) |
| 계약 | `worst < Budget`(예산 100%) | 단일 호출이 예산 전체를 소모하는 계약 위반 |
| **캘리브레이션(신규)** | `median ≤ MaxUnits × refUnit` | **예산 분율 안에 숨는 2배 저하** — 기준 연산(spawn 1회 / append 1사이클) 대비 배수가 코드 회귀를 직접 가림 |

설계 규칙(패키지 doc에 문서화): 기준 연산은 측정 대상과 **같은 비용 계층**(같은 syscall/스폰), 같은 프로세스·같은 파일시스템, 측정 직전에 측정. 비율은 median-vs-median(부하에서 함께 이동).

적용:
- `TestBranchGuard_Latency`: 기준 = ResolveGitDirs 주경로와 동일 인자의 `git rev-parse` spawn 1회(3 warmup + 10측정 중앙값). `MaxUnits=1.5`(건강 시 ~1.0배 — checkBranchState는 spawn 1회 + sub-ms 파싱, `checkout.go:59` 단일 rev-parse). 서브프로세스 1개 추가 회귀 = ≥2.0배 → 포착. 캘리브레이션 팔은 OS 균일(t2 원문의 Windows 헤드룸 우려 해소 — p95 팔의 Windows 50% 분율은 유지).
- `TestRecordEvent100Sequential`: 기준 = 같은 TempDir 파일시스템에서 `open(O_APPEND|O_CREATE|O_WRONLY) + write(160B) + close` 1사이클(3 warmup + 20측정 중앙값). `MaxUnits=2.0`(건강 시 ~1.0배; fsync 회귀 5-50배, 전체 재작성 2-3배 → 포착). Windows skip 유지.
- t2 원문의 POSIX 500ms 무지터 천장은 #1541에서 이미 제거됨(확인), 본 카드로 캘리브레이션 팔이 그 빈 자리의 회귀 검출을 가져옴.

## 3. 다섯 섹션 증거

### Claim (주장)
1. `internal/timing` 신규 패키지(헬퍼+자가테스트 12개)가 초록이다.
2. 재배선된 두 타이밍 테스트가 통과하며, 캘리브레이션 비율이 건강 코드에서 ~1.0배로 관측된다.
3. 4배 비용 성장이 캘리브레이션 팔에 실측으로 걸린다(예산 분율이 전부 관대해도).
4. build/vet/lint/gofmt 통과.
5. hook 패키지 전체 실행의 실패 2건은 release 사이드 커밋이 가져온 사전 존재 실패다(내 diff 무관).

### Evidence (증거) — verbatim
```
$ go test ./internal/timing/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/timing	0.515s

$ go test ./internal/hook/ -run TestBranchGuard_Latency -count=1 -v
    timing.go:144: checkBranchState: n=100 median=38.961ms p95=49.825ms worst=58.712ms avg=39.759ms | refUnit=37.062ms ratio=1.05x (maxUnits=1.50x, steadyCeiling=1s, budget=5s)
--- PASS: TestBranchGuard_Latency (5.01s)

$ go test ./internal/harness/ -run TestRecordEvent100Sequential -count=1 -v
    timing.go:144: RecordEvent: n=100 median=29µs p95=40µs worst=78µs avg=31µs | refUnit=31µs ratio=0.94x (maxUnits=2.00x, steadyCeiling=1s, budget=5s)
--- PASS: TestRecordEvent100Sequential (0.01s)

$ go test ./internal/timing/ ./internal/harness/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/timing	0.242s
ok  	github.com/modu-ai/moai-adk/internal/harness	0.554s

$ go build ./internal/timing/ ./internal/hook/ ./internal/harness/ && go vet ./internal/timing/ ./internal/hook/ ./internal/harness/
BUILD_VET_OK

$ gofmt -l internal/timing/ internal/hook/pre_tool_branch_guard_integration_test.go internal/harness/observer_test.go
(빈 출력 = 정상)

$ golangci-lint run ./internal/timing/... ./internal/hook/... ./internal/harness/...
0 issues.
```
4배 포착 자가테스트(실측 기반): `TestMeasureCalibratedRatioTripsAt4x` PASS — ref=cpuUnit(2M) 대비 4배 fn이 MaxUnits 1.5를 침범해 "calibrated bound" 에러 1건 반환 확인.

### Baseline-attribution (baseline 귀속)
- 위 전부 본 워크트리(`.claude/worktrees/t36`), HEAD `367fa6374` + 작업 diff(2 test 파일 수정, internal/timing/ 신규)에서 이번 실행으로 측정. `-count=1`로 캐시 무효화.
- 사전 존재 실패 귀속(내 diff와 무관함의 기계적 증거):
  - `TestHomeJoinSiteCountIsPinned` 5≠4: 5번째 사이트 `internal/cli/tokens.go`는 release 전용 커밋 `dd060a191`("feat(cli): add moai tokens record")이 추가. `git log --oneline origin/main..HEAD -- internal/cli/tokens.go internal/hook/session_end.go`로 확인. → **release 팁 자체의 미해결 결함**(wantSites 미인상). 별도 카드 권장(wantSites 4→5 + tokens.go 누수 가드 추가).
  - `TestConsumerOnly_M0AndMxByteUnchanged`: mx 경로 hit(`internal/mx/spec_loader{,_test}.go`)은 release 전용 커밋 `557877c49` 소유. `git log --oneline origin/main..HEAD -- internal/mx/spec_loader.go`로 확인. 이 가드는 `origin/main...HEAD` diff를 검사하므로 release 팁을 머지한 어떤 레인 브랜치에서도 동일하게 발화.
  - hook 패키지 전체 실행(`go test ./internal/hook/ -count=1`)의 실패 함수는 정확히 위 2개뿐(`--- FAIL` 라인 2건).

### Gaps (미검증)
- hook 패키지의 나머지 전체 테스트가 위 2건 제외 시 통과임은 실패 라인 목록으로만 확인(2건 실패 외 전부 PASS는 러너 요약에서 함의). 원천 봉쇄 재실행은 안 함(직렬 검증 규율 + CI가 판정).
- Windows에서의 캘리브레이션 비율 미실측(로컬 darwin만). Windows git.exe spawn 비율도 ~1.0배일 것으로 예상하나 CI 매트릭스 판정 대기.
- 풀 스위트 부하 상태에서의 재측정 없음 — 부하 재현 스핀루프 금지 규율 준수. 부하 불변성은 설계(양변 동일 부하)와 4배 자가테스트로 뒷받침.
- TestBuilderNormalizesMode 관련: #1467 해결(#1552)이 '풀 스위트 부하에서만 FAIL'이었던 원인이라는 해석은 주석+커밋 정황에 기반한 추정(해당 실패 로그 원문은 미확인).

### Residual-risk (잔여 위험)
- 극단적 스케줄러 경합에서 짧은 기준 연산의 중앙값이 흔들릴 수 있음 — median-vs-median + 넉넉한 MaxUnits(1.5/2.0 vs 건강 1.0)로 완충. CI에서 마진이 빠듯하게 보이면 상수 조정 여지.
- MaxUnits 상수는 darwin 실측(1.05x/0.94x) 기반 — Linux/Windows 러너에서 마진 검증은 CI 몫.
- 캘리브레이션 팔은 "기준과 같은 비용 계층의 회귀"만 잡는다(예: spawn 수 증가, fsync 추가). 기준 계층 밖의 미세 회귀(파싱 로직 2배 등 sub-ms 영역)는 여전히 분포 팔·예산 팔의 영역.

## 4. 변경 파일

- 신규 `internal/timing/timing.go` — 패키지 doc(교리: 왜 in-run 캘리브레이션인지, verify snapshot 불가 근거), `Bound`/`Stats`/`Median`/`Assert`/`Check`.
- 신규 `internal/timing/timing_test.go` — 자가테스트 12개(순수 Check 로직 + 실측 1배/4배).
- `internal/hook/pre_tool_branch_guard_integration_test.go` — TestBranchGuard_Latency 재배선(+doc 갱신).
- `internal/harness/observer_test.go` — TestRecordEvent100Sequential 재배선, `percentileDuration` 헬퍼로 흡수(중복 제거).
- 본 보고서.
