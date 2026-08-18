# wscfg-timing VCI Evidence — windows·ubuntu 타이밍 보정 클러스터 ③

Branch: WT-wscfg-timing @ base a39646a91 (origin/release/v3.1.1)
Date: 2026-08-18
Session: db221a6c-e73f-4806-b60e-bc00af9ab6fa (run lane)

## 1. Claim (주장)

1. **ubuntu 오작동(TestRecordEvent100Sequential ratio 2.56x~3.61x 허위 발화)의 근본 원인은
   기준 연산의 비용-믹스 불일치**다: 현행 refUnit은 append 사이클(syscall)뿐이나 RecordEvent는
   타임스탬프+MkdirAll+json.Marshal(CPU)+append(syscall)의 혼합이다. VM 러너는 CPU 계층과
   syscall 계층을 서로 다른 배율로 인플레이션시켜(실측: ref 24µs 그대로 vs 측정 60µs — 
   darwin에서는 양쪽 29/31µs) 비율이 건강 코드에서 2.56~3.61x로 무너진다. 이는 패키지 문서
   스스로의 규칙("기준은 같은 비용 계층이어야 한다 — 불일치면 부하에서 분리된다")이 µs 규모에서
   위반된 형태다.
2. **처방 = 카드 설계 방향 ①(기준측정 refUnit이 측정 대상과 같은 비용 프로파일)**: refUnit을
   RecordEvent 건강 경로 전체 믹스(동일 타임스탬프 호출 + 동일 MkdirAll + 동일 Event shape의
   json.Marshal + marshaled 라인 1 append 사이클)를 미러링하게 재구성했다. 클래스별 인플레이션
   배율 C는 양변에 동일 가중치로 곱해져 상쇄된다(ratio = (C·mix)/(C·mix)). **덧셈형 회귀 — 가드
   본연의 목적인 fsync 추가(5-50x)·서브프로세스 추가 — 는 ref에 없는 새 비용 항이므로 그대로
   배수로 발화**한다. MaxUnits 2.0은 변경하지 않았다(임계치를 건드리지 않고 기준을 바로잡는
   최소 처방 — 카드의 "가드 목적 약화 금지" 준수).
3. **windows 오작동(TestMedianRunsWarmupPlusSamples "Median of a real call = 0s")의 근본
   원인은 Windows 단조 클록 해상도**: GitHub Windows 러너의 클록은 인터럽트 주기로 굵어서
   서브틱 크기의 cpuUnit(200k) 호출이 0으로 측정된다. 처방: 자가테스트의 "real call"을
   **클록 1틱 진행을 보장하는 단위**(틱이 진행될 때까지 스핀, 반복 상한으로 비정상 클록
   안전장치)로 교체 — 어떤 클록 해상도에서도 측정값 > 0이 보장된다.
4. 패키지 문서(internal/timing package doc)에 반증 교훈을 반영: "부하는 양변을 동등히
   부풀린다"는 주장을 "같은 비용 **믹스**일 때만 성립"으로 정밀화 + 근거 잡(job 95500006280) +
   Reference rules에 믹스 규칙·클록-틱 규칙 신설.
5. 정직한 한계 고지: 순수 곱셈형 회귀(전체 재작성 2-3x)는 VM에서 노이즈 플로어 아래로
   희석된다 — 단, 구형(쓰기 전용 ref) 설계에서도 VM의 건강 노이즈(2.56-3.61x)가 신호(2-3x)를
   이미 삼키고 있었으므로 **감지력의 순손실은 없다**. dev 머신(CPU 점유율 ≈ 0)에서는 여전히
   2-3x로 보인다. 테스트 doc 주석에 이 매트릭스를 명시.

## 2. Evidence (증거) — command + verbatim output

### darwin 6회 관찰 (재구성된 ref, 건강 코드)

```
$ go test ./internal/harness/ -run TestRecordEvent100Sequential -count=5 -v (+최초 1회)
RecordEvent: n=100 median=46µs ... | refUnit=47µs ratio=0.97x
RecordEvent: n=100 median=46µs ... | refUnit=41µs ratio=1.11x
RecordEvent: n=100 median=47µs ... | refUnit=31µs ratio=1.48x
RecordEvent: n=100 median=44µs ... | refUnit=44µs ratio=0.99x
RecordEvent: n=100 median=49µs ... | refUnit=43µs ratio=1.15x
RecordEvent: n=100 median=46µs ... | refUnit=43µs ratio=1.09x
--- PASS ×6
```

관측 범위 0.97x~1.48x — t36의 구형 ref darwin 관측(0.94~1.49x)과 동일 대역. darwin에서는
CPU 믹스 점유가 ≈0이므로 두 설계가 같게 동작하는 것이 이론적 예상과 일치. 최악 1.48x 대비
MaxUnits 2.0 마진 ~35% 유지.

### windows 수정 대상 자가테스트 (darwin에서의 회귀 없음 확인)

```
$ go test ./internal/timing/ -count=1
ok  github.com/modu-ai/moai-adk/internal/timing 2.826s
```

### 전체 검증

```
$ go build ./internal/timing/ ./internal/harness/ && go vet ./internal/timing/ ./internal/harness/
BUILD_VET_OK
$ gofmt -l internal/timing/timing.go internal/timing/timing_test.go internal/harness/observer_test.go
(빈 출력)
$ golangci-lint run ./internal/timing/... ./internal/harness/...
0 issues.
$ go test ./internal/timing/ ./internal/harness/ -count=1
ok  github.com/modu-ai/moai-adk/internal/timing  0.316s
ok  github.com/modu-ai/moai-adk/internal/harness  0.811s
$ go test -race ./internal/timing/ ./internal/harness/ -count=1
ok  github.com/modu-ai/moai-adk/internal/timing  1.853s
ok  github.com/modu-ai/moai-adk/internal/harness  3.443s
```

### CI 실패 원문 (근거, 본 카드 수정 대상)

- ubuntu job 95500006280: `RecordEvent: n=100 median=60µs ... | refUnit=24µs ratio=2.56x (maxUnits=2.00x)` — 캘리브레이션 팔만 발화 (p95=141µs « steadyCeiling 1s, worst=7.592ms « budget 5s — 분포·계약 팔 정상).
- windows job 95500006316: `TestMedianRunsWarmupPlusSamples ... timing_test.go:154: Median of a real call = 0s, want > 0`.

## 3. Baseline-attribution (baseline 귀속)

위 전부 본 워크트리(`.claude/worktrees/wscfg-timing`), HEAD a39646a91 + 작업 diff(3파일)에서
이번 실행으로 측정. `-count=1`/`-count=5`로 캐시 무효화. darwin (Apple Silicon) — ubuntu/windows
거동은 본地在 확인 불가.

## 4. Gaps (미검증)

- **ubuntu/windows 실거동 미검증** — 본 처방의 최종 판정은 카드 지정대로 ubuntu·windows
  Release Verify 재실행. darwin은 두 설계가 구분되지 않는 영역이므로 로컬 초록은 "회귀 없음"
  증명일 뿐 "VM 수정" 증명이 아니다.
- 곱셈형 회귀(전체 재작성)의 harness-수준 시뮬레이션 테스트는 추가하지 않음(t36도 timing
  패키지 자가테스트 4x 사례에 의존 — 스코프 유지).
- TestBranchGuard_Latency(spawn 계층, ms 규모)는 CI에서 이미 통과 — 미변경(스코프 외).

## 5. Residual-risk (잔여 위험)

- 만약 ubuntu 재실행에서 매칭 ref로도 ratio > 2.0x가 나오면(인플레이션이 클래스-곱셈형이
  아니라 op-특이형이라는 뜻), 남은 카드 옵션: ② CI 환경 감지 별도 한도(MaxUnits만 CI 한정
  상향 — 덧셈형 감지력 유지 범위 내 ≤4x), 또는 ref-측정 interleaving(측정 전·후 2윈도우
  median의 max). 설계 노트를 본 증거에 남겨 다음 레인의 판단 근거로 쓴다.
- darwin 최악 관측 1.48x — MaxUnits 2.0 대비 마진 35%. darwin에서 경계 플레이크가 보이면
  t36이 남긴 조정 레버(MaxUnits 2.5 또는 ref 표본 증가)는 여전히 유효.
- Windows 클록 틱 보장 단위의 상한(5,000,000 반복)은 비정상 클록 안전장치 — 정상 클록에서는
  1틱(≤15.6ms) 내 루프 탈출. 틱이 15.6ms인 최악 환경에서도 자가테스트 총 시간 ≤ ~110ms.

## 변경 파일

- `internal/timing/timing.go` — package doc 정밀화(믹스 규칙·반증 근거·클록-틱 규칙). 코드 변경 없음.
- `internal/timing/timing_test.go` — TestMedianRunsWarmupPlusSamples의 real call을 틱-보장 단위로 교체.
- `internal/harness/observer_test.go` — TestRecordEvent100Sequential의 refUnit을 RecordEvent 전체 믹스 미러링으로 재구성 + doc 주석 갱신(bytes import 제거, path/filepath 추가).
- 본 보고서.
