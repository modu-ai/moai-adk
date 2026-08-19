# t36 Review Verdict — PASS

- reviewer: review lane session (release-v311 hub), 2026-08-17
- target: WT-t36 @ 3c6f99fd9 (single card commit; base refresh 367fa6374 == release tip 091a42e16, already in hub HEAD)
- diff scope: `internal/timing/{timing.go,timing_test.go}` (new, 404 LOC) + `internal/hook/pre_tool_branch_guard_integration_test.go` + `internal/harness/observer_test.go` rewire + card worktree `.moai/reports/t36/report.md`. No production code touched (test-only diff).
- lenses: 기본 4-관점 + 테스트 인프라 설계 검증 (lead dispatch)
- note: 판정문은 리뷰 세션 격리(worktree guard)상 카드 워크트리 대신 release-v311 허브에 기록 — 기존 리뷰 판정문 15건과 동일 경로 규칙

## Verdict

**PASS** — release/v3.1.1 통합 승인. 라이더 2건(권고, 머지 조건 아님).

## 지정 검증 6건 — 전항 재현 완료

### 1. verify-snapshot 불가 입증 — ✅ 코드 근거 실재·해석 정확

`internal/verify/key.go` `Key()` 구조 재현: HEAD SHA(`rev-parse HEAD`) + `git status --porcelain=v2` digest + `git diff HEAD` 내용 해시 3입력. 리포트의 3가지 근거 전부 코드 정확:
- 커밋 1개 → HEAD SHA 변화 → 새 키. dirty 파일 재편집 → diff 해시 변화 → 새 키(key.go doc 자체가 porcelain/diff 분리 근거 명시 — "porcelain-v2 output is byte-identical across successive edits to an already-dirty tracked file").
- 시점 불일치 = 부하 민감성 재래, CI 휘발성(`.moai/state/verify/` 런타임 상태) — 설계 근거로 타당.

### 2. internal/timing 설계 — ✅ 3팔 + 자가테스트 12 + 4배 trip 재현

- 3팔 확인: `Check()`가 p95≤SteadyCeiling / worst<Budget / median≤MaxUnits×refUnit(활성 조건 `MaxUnits>0 && refUnit>0`) 순수 함수로 평가. Assert는 measure+log+Check 배선.
- percentile nearest-rank `ceil(p·n)` — 기존 구현(`(n*95)/100-1`, n=100 → idx 94)과 동일 인덱스. 흡수 시 의미 변화 없음.
- 자가테스트 12개 전수 열거·전부 재실행 PASS (`go test -C <t36wt> ./internal/timing/ -count=1`, 이번 실행).
- **4배 성장 실측 trip 재현**: `TestMeasureCalibratedRatioTripsAt4x` PASS (cpuUnit 2M ref vs 8M fn → calibrated bound 에러 정확히 1건).

### 3. 적용 2건 산술 + 기준 연산 일치 — ✅ (관찰 노트 1건 → R2)

- **BranchGuard**: `ResolveGitDirs` 주경로(`internal/core/git/checkout.go:54`)는 `git rev-parse --path-format=absolute --git-dir --git-common-dir` spawn 정확히 1회(fallback은 실패 시에만) — 테스트 ref 인자와 **정확히 일치** 확인. checkBranchState = spawn 1 + sub-ms 파싱 → 기준과 같은 비용 계층. MaxUnits 1.5: 건강 실측 1.05x(리포트)/0.94x(이번 재실행), spawn 1개 추가 = ≥2.0x → 포착 산술 성립.
- **RecordEvent**: 구현은 Marshal + MkdirAll + `O_APPEND|O_CREATE|O_WRONLY` open/write/close(observer.go:55 재현) — ref(같은 TempDir 형제 파일, 같은 open 플래그, 160B 라인)와 같은 비용 계층. MaxUnits 2.0: fsync 5-50x → 포착 성립.
- ⚠ 관찰: 이번 재실행에서 건강 ratio **1.49x** 관측(리포트 0.94x) — 건강 변동 상단이 MaxUnits 2.0과 가까움. whole-file rewrite 2-3x의 하단(≈2.0x)은 `ratio > 2.0` 임계와 겹칠 여지. 실제 놓침 사례는 아님(R2 권고).

### 4. 전제 정정 — ✅ 4요소 전부 재현

`git log --all -S "TestBuilderNormalizesMode" -- builder_test.go` → 커밋 2개(초기, #466)만. `-S "time.Since"` → **0건**. 현재 파일 wall-clock 단언 부재(grep 0건). #1552 = `6ff7c3aaf` 존재("bound usage fetch phase against unbounded Retry-After (#1467)"). TestBuilderNormalizesMode 제외 판정 정당.

### 5. 부하 규율 — ✅

전 diff 독독: 무한 루프·백그라운드 고루틴·time.Sleep·스핀 대기 0건. 모든 측정 유한(cpuUnit 2M-8M 반복, ref spawn 13회, ref append 23회, 측정 100회). t.TempDir 자동 정리. 이번 리뷰 재실행도 타깃 패키지 3개만(풀 스위트 미실행 — 규율 준수).

### 6. 부수 발견 2건 귀속 — ✅ 재현, 판정 제외 확인

`dd060a191`(tokens.go)/`557877c49`(mx spec_loader)는 `origin/main..3c6f99fd9`에 존재(release 선행), 카드 diff `091a42e16..3c6f99fd9`에는 tokens.go/mx **0건**. 카드 무관 입증 — t116 별도 카드로 분리됐으므로 본 판정에서 제외.

## 4-관점 평가

| 관점 | 판정 | 근거 |
|---|---|---|
| Functionality | PASS | 리포트 Claim 1-3 전부 이번 재실행으로 재현 초록(아래 Evidence). Claim 4는 gofmt/build/vet 재현 ✅, lint는 갭. Claim 5 귀속 git log 재현 ✅ |
| Security | PASS | production 코드 0변경(test-only). 신규 외부 입력·비밀·권한 경로 없음. ref git spawn은 기존 hook과 동일 클래스 명령 |
| Craft | PASS | 패키지 doc이 설계 근거(왜 in-run, 왜 persisted baseline 불가, reference 3규칙)를 밀도있게 문서화. Check 순수함수 분리로 측정 없는 로직 테스트. percentileDuration 중복 흡수. ratio 로그 출력으로 기계적 관찰 가능. panic 계약(doc 명시)은 니트 수준 |
| Consistency | PASS | 코드·주석 전부 영어(프로젝트 규칙), observer 기존 한국어 에러 메시지 형식 보존, 5섹션 리포트·doc 주석 스타일이 #1541 계보 타이밍 테스트와 연속 |

## 재현 Evidence (이번 리뷰 실행, t36 트리)

```
$ go test -C <t36wt> ./internal/timing/ ./internal/hook/ ./internal/harness/ \
    -run 'TestBranchGuard_Latency|TestRecordEvent100Sequential|Test…' -count=1
ok  github.com/modu-ai/moai-adk/internal/timing    1.464s   (12/12 PASS, 4x trip 포함)
    timing.go:144: checkBranchState: n=100 median=48.857ms p95=79.238ms worst=254.37ms | refUnit=51.759ms ratio=0.94x (maxUnits=1.50x)
    --- PASS: TestBranchGuard_Latency (6.84s)
$ go test -C <t36wt> ./internal/harness/ -run TestRecordEvent100Sequential -count=1
    timing.go:144: RecordEvent: n=100 median=45µs p95=68µs | refUnit=31µs ratio=1.49x (maxUnits=2.00x)
    --- PASS: TestRecordEvent100Sequential
$ gofmt -l … → 빈 출력 / go build + go vet → OK
```

## Gaps (이번 리뷰 미재현)

- `golangci-lint 0 issues`(리포트 Claim 4 일부) 재실행 불가 — 리뷰 세션 worktree guard가 타 워크트리 lint 실행 차단. **CI 매트릭스 판정 대기**.
- Windows/Linux 러너 캘리브레이션 ratio 미실측(리포트 Gaps와 동일) — CI 대기.
- hook 패키지 전체 재실행 안 함(부하 규율) — 사전 존재 실패 2건 귀속은 git log로 재현, 실행 재현은 CI 몫.

## 라이더 (권고 — 통합 조건 아님)

- **R1 (문서)**: report §2의 "`checkout.go:59` 단일 rev-parse" 인용은 실제 위치 `internal/core/git/checkout.go:54`로 정정 권고. 주경로 spawn 1회·인자 일치는 본 리뷰가 코드로 확인했으므로 판정 영향 없음.
- **R2 (관찰 범위 문서화)**: RecordEvent 건강 ratio 관측 범위 0.94x(작성자)~1.49x(이번 리뷰) — MaxUnits 2.0 마진의 실측 하한이 얇음. (a) 관측 범위를 report/package doc에 기록, (b) CI 플레이크 시 ref 샘플 증가 또는 MaxUnits 조정 여지 명시. fsync(5-50x) 포착·건강 통과에는 영향 없음.
