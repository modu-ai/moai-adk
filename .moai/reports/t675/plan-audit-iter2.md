# SPEC 검토 보고서: SPEC-DOCTOR-TEST-CWD-ISOLATION-001

- Iteration: 2 (iteration 1 FAIL 0.67 → 결함 D1-D5 + MP-9 GAP 재검)
- 대상: `spec.md` v0.3.0 + `plan.md` (Tier S), 워크트리 `.claude/worktrees/t675`, 브랜치 `WT-doctor-red`, HEAD `f67438cbb9f113206c269454285f8bcac299be1c`
- Verdict: **PASS**
- Overall Score: **0.92** (조화평균)
- Tier S PASS threshold: **0.75**
- 감사 방식: Claude 단독(설정 섹션에 `audit_model` 키 없음 — 교차 모델 백엔드 미호출)
- 작성자 추론 맥락은 M1 Context Isolation에 따라 무시했다. 판정은 산출물과 이번 실행의 재측정만 근거로 한다.
- 규칙 인용: `verification-completeness.md §1.1`(관측된 실패), `§2`(두 셀 채택 + mutant probe), `§2.1`(RED 셀 네 요소), `verification-claim-integrity.md §2`(baseline 귀속)

## Must-Pass Results

| 기준 | 판정 | 증거 |
|---|---|---|
| MP-1 REQ 번호 일관성 | PASS | `spec.md:L51,L56,L64,L69,L73,L78` — `REQ-DTC-001`~`006` 순차·중복 없음·3자리 패딩 일관 |
| MP-2 GEARS 형식 (요구사항 층 판정) | PASS | REQ별 `shall` 개수 재측정: 001~006 모두 `shall=1`. 001 When / 002 While / 003 Ubiquitous / 004 `shall not` / 005 When / 006 When. AC는 검증 층이므로 여기서 채점하지 않음 |
| MP-3 YAML frontmatter | PASS | `spec.md:L2-L13` 12필드 전부, `version: "0.3.0"` 따옴표 semver, `created`/`updated` ISO, `phase: "v3.2.0 target"`(생애주기 토큰 아님). `moai spec lint --strict --json` → exit 0, `info` 1건만 |
| MP-4 언어 중립성 | N/A | `internal/cli` Go 테스트 한정 단일 언어 SPEC |
| MP-5 D7 교차 SPEC | PASS | spec/plan 본문의 SPEC-ID 참조는 자기 자신뿐(`spec.md:L2,L17`, `plan.md:L1,L129,L130`) — BLOCKING 없음 |
| MP-6 D8 교차 플랫폼 | PASS | `syscall` 일치 0건 |
| MP-7 clarification gate | PASS | `plan.md` `[NEEDS CLARIFICATION` 0건, Tier S라 `research.md` 없음 |
| MP-8 RED 재실행 | PASS | E-RED-001/002를 HEAD에서 재실행, 둘 다 exit 1·기록과 같은 실패 집합(아래 증거 3·4). HEAD와 기준 `dd235a66b` 사이 차이는 `.moai/` 문서뿐(증거 1) |
| MP-9 교차 산출물 순서(CN-4) | PASS | `plan.md:L105` M1 `Exit: AC-DTC-001, AC-DTC-002`, `plan.md:L114` M2 `Exit: AC-DTC-003, AC-DTC-004, AC-DTC-005`; `plan.md §E` 1~5단계가 AC 번호순; `spec.md:L307-308` → `plan.md §D C3` 실재; `plan.md:L63` C3의 문장 `t.Chdir(t.TempDir())`가 AC-DTC-004 조건 (2)의 문자열과 일치 |

## Category Scores

| 평가 항목 | 점수 | 밴드 | 근거 |
|---|---:|---|---|
| Clarity | 0.75 | 0.75 | REQ/AC는 단일 해석. 다만 사후 병합 시 "공허하게 통과"라는 서술(`plan.md:L36-38`)이 실제 동작과 반대이고, `spec.md:L37-38`이 인용한 자연 상태 근거가 비정확 selector에 기대고 있음(O2, O3) |
| Completeness | 1.00 | 1.0 | HISTORY/문제/요구사항/inline AC/RED ledger/제약/`### Out of Scope — …` 4개/교차참조 모두 존재. HISTORY의 0.2.0 행 누락은 선택 결함(O1) |
| Testability | 1.00 | 1.0 | 5개 AC 모두 이진 판정 가능·weasel word 없음. GREEN 경로 실측 가능, mutant 5종 중 요구사항 위반 변이는 AC 집합이 전부 거부(증거 6·7·8) |
| Traceability | 1.00 | 1.0 | REQ-001←AC1,2,5 / 002←AC1,2 / 003←AC2,4 / 004←AC3 / 005←AC5 / 006←AC5. 고아 AC·미커버 REQ 없음 |

```text
4 / (1/0.75 + 1/1.00 + 1/1.00 + 1/1.00) = 4 / 4.3333 = 0.923 ≈ 0.92
```

## Regression Check (iteration 1 결함)

- **D1 MP2-GEARS-MULTICLAUSE — RESOLVED.** 이전 REQ-002/003의 이중 규범이 REQ-002/003/004로 분리됨. 재측정: 모든 REQ `shall=1`(증거 5). 문구 치환이 아니라 구조 분리임을 확인 — REQ-004 "shall not modify any non-test Go file"은 이전 REQ-003 둘째 문장을 독립 REQ로 옮긴 것이고 AC-DTC-003이 새로 커버함.
- **D2 AC-BASELINE-CONTRADICTION — RESOLVED.** `spec.md:L85-89`가 모든 AC의 Given을 "implementation descendant"로 고정하고, RED 기준 SHA는 §3.1 ledger에만 둔다. AC-DTC-001 `spec.md:L95`의 Given에 더 이상 기준 SHA가 없다.
- **D3 RED-LEDGER-NONVERBATIM — RESOLVED.** `spec.md:L159-166`, `L179-274`의 fenced 블록을 원본 파일과 바이트 비교: 둘 다 `identical=True`(6행/94행). sha256도 `spec.md:L157,L177` 기재값과 일치(증거 2). exit 코드는 별도 `.exit` 파일(`exit=1`)로 분리 기록.
- **D4 AC-MUTANT-COVERAGE-GAP — RESOLVED.** AC-DTC-004(제거 줄 0 + 추가 줄은 정확히 9개의 격리 문장)와 AC-DTC-005(`os.Chdir(` 0 + hunk 헤더가 9개 테스트를 각 1회) 추가. 이번 감사가 독립 mutant로 재검(증거 6~8). 단, 작성자 probe(`red/mutant-probe.txt`)는 hunk 헤더 조건을 실행하지 않았다 — 선택 결함 O4, 이번 감사가 공백을 메웠다.
- **D5 REQ-CLEANUP-WORDING-AND-HOW — RESOLVED.** REQ-005(임시 디렉터리 부재)와 REQ-006(원래 CWD와 같음)으로 관찰 가능한 결과만 서술하고, "Go test framework" 수단은 `plan.md:L58-63` C3로 이동. 요구사항 층에 구현 수단 없음.
- **MP-9 GAP — RESOLVED.** 이번 iteration에서 명시 판정(PASS, 위 표).

## Defects Found

치명(critical)·차단(blocking) 결함 없음. 아래는 모두 선택 결함이다.

O1. HISTORY-VERSION-GAP — `spec.md:L21-24` — iteration 1이 감사한 v0.2.0(커밋 `49bf74a82`의 `version: "0.2.0"`)이 HISTORY에 행이 없어 0.1.0 → 0.3.0으로 건너뛴다. — Severity: minor — Class: optional — Required fix: 0.2.0 행(2026-09-13, iteration 1 감사 대상)을 추가하거나 0.3.0 행에 0.2.0을 대체했다고 적는다.

O2. POST-MERGE-VACUITY-MISSTATED — `plan.md:L36-38`, `spec.md:L88-89` — "병합 뒤 범위가 비어 공허하게 통과한다"고 쓰지만, 빈 범위에서 AC-DTC-003은 "정확히 세 경로" 미충족, AC-DTC-004 (2)와 AC-DTC-005는 "정확히 9개" 미충족으로 **적색**이 된다(증거 9: 현 HEAD의 빈 출력이 바로 그 상태이며 §3.1 E-RED-003도 이를 RED로 기록). 판정식은 비공허하게 설계돼 있고 서술만 틀렸다. — Severity: minor — Class: optional — Required fix: "병합 뒤에는 범위가 비어 조건이 적색이 되므로 병합 전 평가 전용"으로 정정.

O3. NATURAL-STATE-EVIDENCE-SELECTOR — `spec.md:L37-38` → `progress.md:L24` — 인용된 "자연 상태 9건 전부 PASS"는 selector `'TestRunDoctor_|TestDoctorCmd_'`로 측정됐는데, 이 selector는 이번 실측에서 27개 테스트를 고른다(증거 10). "9건"은 그 출력의 부분 판독이다. 규범 문장이 아닌 배경 서술이라 AC에는 영향 없음. — Severity: minor — Class: optional — Required fix: E-RED-002의 정확한 selector로 무주입 실행을 재측정해 원본 파일로 보존하거나, 인용을 "27개 선택 중 9개 포함 전부 PASS"로 정정.

O4. MUTANT-PROBE-HEADER-UNEXERCISED — `.moai/reports/t675/red/mutant-probe.txt:L2-4,L11-22` — probe가 평문 `diff -U0`를 써서 hunk 헤더에 함수명이 없다. AC-DTC-005의 "hunk 헤더가 9개 테스트를 명명" 조건은 작성자 증거로는 한 번도 관측되지 않았다(`§1.1`). 이번 감사가 `git diff --no-index -U0`로 헤더가 9개 이름을 정확히 내는 것을 관측해 실현 가능성은 확인됨(증거 6). — Severity: minor — Class: optional — Required fix: run 단계 `progress.md §E.2`에 실제 `git diff -U0 develop...HEAD` 출력을 원문 보존(이미 계획됨, `spec.md:L295-296`).

O5. REQ-002-SINGLE-MEMBER-WITNESS — `spec.md:L56-62`, `L91-107` — REQ-002는 "모든 repository-only 검사"를 말하지만 AC는 Agent Emit Embed 한 구성원만 오염시켜 확인한다. 격리가 CWD 단위라 다른 구성원에도 같은 기제로 작동하므로 차단 사유는 아니다. — Severity: minor — Class: optional — Required fix: 필요 시 rationale에 "현재 repository-only 검사는 Agent Emit Embed 하나"라는 실측 근거를 붙인다.

## 사용자 지정 검증 항목 판정

### 1) iteration 1 결함이 실제로 해결됐는가
위 Regression Check 참조. 다섯 건 모두 구조적으로 해결, 문구 치환 아님.

### 2) RED ledger 재실행
HEAD `f67438cbb`에서 E-RED-001 exit 1, E-RED-002 exit 1. 실패 테스트 집합·실패 줄 번호(`doctor_test.go:76`, `coverage_improvement_test.go:715/737/777/4930/5754/5804`, `integration_test.go:176/202`)·유일한 ✗가 Agent Emit Embed라는 점이 기록과 동일. 소요 시간만 다름(부하 차이). `-list` selector는 9개 이름을 정확히 출력.

### 3) Mutant probe
원본 워크트리는 건드리지 않고, 스크래치패드 사본 + `go test -overlay`로 실행했다.

| 변이 | AC-003 | AC-004 | AC-005 | AC-001/002 (실행) | 결과 |
|---|---|---|---|---|---|
| good: 9개 테스트 첫 문장에 `t.Chdir(t.TempDir())` | 충족 | 충족 | 충족(헤더 9개 이름) | 9 PASS, exit 0 | 수용 — GREEN 경로 실현 가능 |
| mutA: good + 단언 3줄 삭제 | 충족 | **거부**(제거 3줄) | 충족 | — | 거부 |
| mutB: bare `_ = os.Chdir(t.TempDir())` | 충족 | **거부** | **거부** | — | 거부 |
| mutC: 헬퍼 호출 `isolateCWD(t)` | 충족 | **거부** | **거부** | — | 거부 |
| mutD: 정확한 문장을 함수 **끝**에 배치(격리 무효) | 충족 | 충족 | 충족 | **거부**(exit 1, 여전히 FAIL) | 거부 — 구조 AC는 통과하지만 실행 AC가 잡음 |

사용자가 제시한 변이 판정:
- 단언 약화(삭제가 아닌 수정): 기존 줄 수정은 git diff에서 제거 줄 1개를 만들어 AC-004 (1)이 거부. 새 줄로 약화하면 (2)의 "정확히 격리 문장만"이 거부.
- 임시가 아닌 디렉터리로 `t.Chdir`(예: `t.Chdir("/")`): 추가 줄이 `t.Chdir(t.TempDir())`와 다르므로 AC-004 (2)가 거부.
- 테스트별이 아닌 헬퍼에서 격리: mutC로 실측 거부.
- 격리 문장을 비실행 분기/defer 안에 두기: 추가 줄은 AC-004를 통과할 수 있으나 격리가 doctor 실행 전에 일어나지 않아 AC-001/002가 오염 입력에서 적색 — mutD로 같은 기제를 실측.

AC 집합 전체를 통과하면서 REQ를 위반하는 변이는 작성하지 못했다. 선택된 9개 테스트의 단언은 모두 절대 경로(`t.TempDir()` 기반) 또는 출력 문자열에만 의존하므로, CWD 이동이 기존 단언을 의미상 공허하게 만들지도 않는다(각 테스트 본문 확인).

### 4) `develop...HEAD` 범위의 흡수 후 건전성
- 로컬 `develop`은 지시문의 `9aee76589`에서 다시 `27220fb94`로 이동해 있다(증거 9). merge-base는 여전히 `f67d2193f`.
- 3점 범위는 흡수되지 않은 develop 변경을 배제한다: 같은 시점 2점 `develop HEAD`는 Go 파일 21개를 내지만 3점은 0개(증거 9).
- 흡수 시뮬레이션: `git merge-tree --write-tree develop HEAD` → `e9a8126a8…`(참조 불변), `git diff --name-only develop e9a8126a8 -- '*.go'` → 빈 출력. 즉 깨끗한 흡수 뒤 merge-base가 흡수한 develop 끝이 되어, 범위에는 카드 자기 기여만 남는다.
- `develop`이 흡수한 커밋들은 세 테스트 파일도 doctor 소스도 건드리지 않는다(`f67d2193f..develop -- internal/cli/`는 init/update 파일만).
- 잔여 위험: 흡수 중 세 테스트 파일에 충돌이 나서 해결 편집이 생기면 AC-004 (1)이 적색이 된다. 이는 오탐이 아니라 실제 편집을 드러내는 올바른 적색이다. 병합 뒤 평가는 적색(O2)이므로 공허 통과 위험 없음.

### 5) 공허·불가능 AC 여부
- 공허: 없음. AC-001/002는 기준에서 적색이고(RED-now 셀 실측), AC-003~005는 빈 범위에서 적색(정확한 개수 조건).
- 불가능: 없음. good 변이로 AC-001~005 전부 충족을 실측(증거 6·7). `go.mod`의 `go 1.26.8`은 `testing.T.Chdir`를 제공하며 overlay 컴파일·실행으로 확인됨.
- 잘못된 이유의 적색: 없음. RED 원인은 오직 Agent Emit Embed(✗ 1개)이며 격리만으로 뒤집힘.

### 6) `OwnershipTransitionUnmeasured` 공개 여부
`progress.md:L16`이 커밋 `49bf74a82`, `info` 등급, trailer 부재, 이력 재작성 없이는 해소 불가를 정확히 기재. 이번 lint 출력(증거 11)과 커밋 SHA·메시지·등급이 일치. `--strict`에서 exit 0 유지. 올바르게 공개됨.

## Evidence

### Claim
- 이전 결함 D1-D5와 MP-9 GAP이 모두 해소됐다.
- RED ledger는 HEAD에서 재현되고 원본 파일과 바이트 동일하다.
- GREEN 경로는 실제로 달성 가능하며, 요구사항을 위반하는 변이는 AC 집합이 거부한다.
- 흡수 전후 모두 diff 범위가 카드 기여만 측정한다.

### 증거 1 — 트리 귀속
명령: `git -C <wt> diff --stat dd235a66b1145922565841d33acafef0d1ded6a8 f67438cbb`
```text
 .moai/reports/t675/red/e-red-001.exit              |   1 +
 .moai/reports/t675/red/e-red-001.txt               |   6 +
 .moai/reports/t675/red/e-red-002.exit              |   1 +
 .moai/reports/t675/red/e-red-002.txt               |  94 ++++++
 .moai/reports/t675/red/mutant-probe.txt            |  30 ++
 .../SPEC-DOCTOR-TEST-CWD-ISOLATION-001/plan.md     | 154 +++++-----
 .../SPEC-DOCTOR-TEST-CWD-ISOLATION-001/progress.md |  10 +-
 .../SPEC-DOCTOR-TEST-CWD-ISOLATION-001/spec.md     | 316 ++++++++++++++++-----
 8 files changed, 455 insertions(+), 157 deletions(-)
```

### 증거 2 — 해시와 ledger 원문 동일성
명령: `shasum -a 256 spec.md plan.md red/e-red-001.txt red/e-red-002.txt`
```text
23fef05b32fb4a0dae34443c25cbf92261f6162cfd63f65c561b94f44769111b  spec.md
0fd2f2514a1383185e6e6548b08c62b8fb88f7d4a24820e6f67ee634bacb2cfa  plan.md
2ff331b759f1764a778f22559dbd43e15678e3ab3e923f18b0354f56807d3739  e-red-001.txt
7d1e24166573f4caec68a378ee7b4e9e8789e93f69c9ac92ece8ccb25a361b7c  e-red-002.txt
```
`progress.md:L11-12`, `spec.md:L157,L177` 기재값과 일치. 스크래치 probe 출력:
```text
ledger block 1 vs e-red-001.txt: identical=True (block_lines=6, raw_lines=6)
ledger block 2 vs e-red-002.txt: identical=True (block_lines=94, raw_lines=94)
```

### 증거 3 — E-RED-001 재실행 (HEAD f67438cbb)
명령: `MOAI_EMBED_CHECK_BIN=/usr/bin/false go test -count=1 -v -run '^TestDoctorCmd_Execution$' ./internal/cli/`
```text
=== RUN   TestDoctorCmd_Execution
    doctor_test.go:76: doctor command RunE error: doctor: 1 check(s) failed
--- FAIL: TestDoctorCmd_Execution (12.61s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	13.365s
FAIL
```
exit 1

### 증거 4 — E-RED-002 재실행 (HEAD f67438cbb)
명령: E-RED-002 원문 그대로. exit 1. 요지(원문 전체는 이 감사 세션 출력에서 관측, 기록 파일과 구조 동일):
```text
--- FAIL: TestRunDoctor_WithExport (10.17s)      coverage_improvement_test.go:715
--- FAIL: TestRunDoctor_WithFix (10.21s)         coverage_improvement_test.go:737
--- FAIL: TestRunDoctor_Verbose (10.81s)         coverage_improvement_test.go:777
  ✗ Error: could not extract embedded artifacts from /usr/bin/false: false init: exit status 1 ()
--- FAIL: TestRunDoctor_AllFlags (10.46s)        coverage_improvement_test.go:4930
--- FAIL: TestRunDoctor_VerboseAndDetail (10.20s) coverage_improvement_test.go:5754
--- FAIL: TestRunDoctor_ExportMode (10.15s)      coverage_improvement_test.go:5804
--- FAIL: TestDoctorCmd_Execution (10.13s)       doctor_test.go:76
--- FAIL: TestDoctorCmd_ExportFlag (11.45s)      integration_test.go:176
--- FAIL: TestDoctorCmd_VerboseExecution (10.83s) integration_test.go:202
FAIL	github.com/modu-ai/moai-adk/internal/cli	95.006s
```
(위 블록은 줄 번호를 한 줄로 합친 요약이다. 원문 stdout은 기록 파일 `e-red-002.txt`와 소요 시간 외 동일한 줄 구성이었다.)

Selector 명령: `go test ./internal/cli -list '<E-RED-002 정규식>'`
```text
TestRunDoctor_WithExport
TestRunDoctor_WithFix
TestRunDoctor_Verbose
TestRunDoctor_AllFlags
TestRunDoctor_VerboseAndDetail
TestRunDoctor_ExportMode
TestDoctorCmd_Execution
TestDoctorCmd_ExportFlag
TestDoctorCmd_VerboseExecution
ok  	github.com/modu-ai/moai-adk/internal/cli	0.574s
```

### 증거 5 — REQ 절 개수와 추적
```text
REQ-DTC-001 shall= 1 shall_not= 0
REQ-DTC-002 shall= 1 shall_not= 0
REQ-DTC-003 shall= 1 shall_not= 0
REQ-DTC-004 shall= 1 shall_not= 1
REQ-DTC-005 shall= 1 shall_not= 0
REQ-DTC-006 shall= 1 shall_not= 0
ACs ['AC-DTC-001', 'AC-DTC-002', 'AC-DTC-003', 'AC-DTC-004', 'AC-DTC-005']
covers ['REQ-DTC-001, REQ-DTC-002', 'REQ-DTC-001, REQ-DTC-002, REQ-DTC-003', 'REQ-DTC-004', 'REQ-DTC-003', 'REQ-DTC-001, REQ-DTC-005, REQ-DTC-006']
```

### 증거 6 — 구조 AC mutant 판정 (`git diff --no-index -U0 --no-color`, 스크래치 사본)
```text
good: removed=0 nonblank_added=9 AC-004(c1=True,c2=True)=True AC-005(no_os=True, headers=[9개 이름])=True
mutA: removed=3 nonblank_added=9 AC-004(c1=False,c2=True)=False AC-005(...)=True
mutB: removed=0 nonblank_added=9 AC-004(c1=True,c2=False)=False AC-005(no_os=False, headers=[])=False
mutC: removed=0 nonblank_added=9 AC-004(c1=True,c2=False)=False AC-005(no_os=True, headers=[])=False
mutD: removed=0 nonblank_added=9 AC-004(c1=True,c2=True)=True AC-005(no_os=True, headers=[9개 이름])=True
```
good 변이의 hunk 헤더(AC-DTC-005 조건 실관측):
```text
@@ -699,0 +700 @@ func TestRunDoctor_WithExport(t *testing.T) {
@@ -724,0 +726 @@ func TestRunDoctor_WithFix(t *testing.T) {
@@ -764,0 +767 @@ func TestRunDoctor_Verbose(t *testing.T) {
@@ -4911,0 +4915 @@ func TestRunDoctor_AllFlags(t *testing.T) {
@@ -5740,0 +5745 @@ func TestRunDoctor_VerboseAndDetail(t *testing.T) {
@@ -5787,0 +5793 @@ func TestRunDoctor_ExportMode(t *testing.T) {
@@ -69,0 +70 @@ func TestDoctorCmd_Execution(t *testing.T) {
@@ -156,0 +157 @@ func TestDoctorCmd_ExportFlag(t *testing.T) {
@@ -185,0 +187 @@ func TestDoctorCmd_VerboseExecution(t *testing.T) {
```
(각 헤더 다음 줄은 `+	t.Chdir(t.TempDir())`.)

### 증거 7 — GREEN 실현 가능성 (good 변이, overlay)
명령: `MOAI_EMBED_CHECK_BIN=/usr/bin/false go test -count=1 -v -overlay <scratch>/good.overlay.json -run '<E-RED-002 정규식>' ./internal/cli/`
```text
--- PASS: TestRunDoctor_WithExport (10.40s)
--- PASS: TestRunDoctor_WithFix (10.63s)
--- PASS: TestRunDoctor_Verbose (10.26s)
  ✓ Agent Emit Embed
--- PASS: TestRunDoctor_AllFlags (10.01s)
--- PASS: TestRunDoctor_VerboseAndDetail (10.14s)
--- PASS: TestRunDoctor_ExportMode (10.73s)
--- PASS: TestDoctorCmd_Execution (10.88s)
--- PASS: TestDoctorCmd_ExportFlag (10.13s)
--- PASS: TestDoctorCmd_VerboseExecution (10.06s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	94.047s
```
exit 0, `--- PASS` 9줄, `--- FAIL` 0줄. (AllFlags 진행 출력 중 해당 줄만 발췌.)

### 증거 8 — mutD 실행 거부
명령: `MOAI_EMBED_CHECK_BIN=/usr/bin/false go test -count=1 -v -overlay <scratch>/mutD.overlay.json -run '^TestDoctorCmd_Execution$' ./internal/cli/`
```text
=== RUN   TestDoctorCmd_Execution
    doctor_test.go:76: doctor command RunE error: doctor: 1 check(s) failed
--- FAIL: TestDoctorCmd_Execution (10.12s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	10.888s
FAIL
```
exit 1

### 증거 9 — diff 범위와 흡수 시뮬레이션
```text
$ git rev-parse develop HEAD
27220fb949e8105f1ffd3bb60652482beaba6109
f67438cbb9f113206c269454285f8bcac299be1c
$ git merge-base develop HEAD
f67d2193f22cbc80921e42539c3462a4100f31e9
$ git diff --name-only develop...HEAD -- '*.go'
(빈 출력, exit 0)
$ git diff -U0 --no-color develop...HEAD -- <세 테스트 파일>
(빈 출력, exit 0)
$ git diff --name-only develop HEAD -- '*.go'
internal/cli/init.go … internal/web/schema_parity_guard_test.go   (21개)
$ git merge-tree --write-tree develop HEAD
e9a8126a817d6d873a34e221a9d39de4a881a4c0
$ git diff --name-only develop e9a8126a817d6d873a34e221a9d39de4a881a4c0 -- '*.go'
(빈 출력)
$ git diff --name-only f67d2193f… develop -- internal/cli/
internal/cli/init.go, init_flag_precedence_test.go, init_quiet_wizard_test.go, init_test.go,
init_update_notice_test.go, update_outcome_roots_test.go, update_pair_dedupe_test.go,
update_template_sync.go, update_tux.go
```

### 증거 10 — progress §E.1a selector 범위
명령: `go test ./internal/cli -list 'TestRunDoctor_|TestDoctorCmd_'` → 27개 이름 출력(`TestRunDoctor_WithCheckFilter`, `TestRunDoctor_FixMode`, `TestDoctorCmd_HelpOutput`, `TestRunDoctor_FixWithFailures` 등 포함), `ok … 0.581s`.

### 증거 11 — lint와 lifecycle audit
명령: `moai spec lint SPEC-DOCTOR-TEST-CWD-ISOLATION-001 --strict --json` → exit 0
```json
[{"severity":"info","code":"OwnershipTransitionUnmeasured","message":"SPEC SPEC-DOCTOR-TEST-CWD-ISOLATION-001 transition \"(none)\" → \"draft\" expected owner \"manager-spec\" but commit 49bf74a8283b9f26d445f6162115691a993b7b06 (chore(t675): preserve partial plan and gateway blocker) has no Authored-By-Agent trailer — ownership transition unmeasured"}]
```
`mcp__moai__spec_audit(filter_spec, include_grandfathered=true, project_root=<wt>)`:
```json
{"total_specs":1,"grandfathered":0,"modern_era_clean":1,"drift_findings":[{"spec_id":"SPEC-DOCTOR-TEST-CWD-ISOLATION-001","era":"V3R6","finding_type":"EraAutoDetected","severity":"INFO","details":{"heuristic_matched":"H-5 (modern phase or created date)"}}]}
```

### 증거 12 — 워크트리 무변경
명령: `git -C <wt> status --short` (probe와 테스트 실행 이후) → 빈 출력. mutant는 스크래치패드 사본과 `-overlay`로만 실행했다. `merge-tree --write-tree`는 객체 DB에 트리 객체만 쓰고 ref는 바꾸지 않는다.

## Baseline-attribution

모든 측정은 2026-09-18 이 감사 세션에서 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t675`, HEAD `f67438cbb9f113206c269454285f8bcac299be1c`, 로컬 `develop` `27220fb949e8105f1ffd3bb60652482beaba6109`을 대상으로 했다. HEAD와 RED 기준 `dd235a66b`의 차이는 `.moai/` 문서뿐이므로 Go 트리는 동일하다. 측정 시점 부하: `uptime` load averages 8.21 / 12.83 / 14.12.

## Gaps

- `go vet`, `gofmt`, `golangci-lint`, 커버리지는 run 단계 항목이라 실행하지 않았다.
- 무주입 상태(`MOAI_EMBED_CHECK_BIN` 미설정)의 9개 selector 실행은 이번에 재측정하지 않았다(O3의 교정 대상).
- REQ-005/006(임시 디렉터리 제거, CWD 복원)의 런타임 결과는 직접 관측하지 않았다. AC-DTC-005가 구조로 판정하며, 그 보장은 Go `testing` 프레임워크의 cleanup 의미론에 기댄다.
- 흡수는 `merge-tree` 시뮬레이션으로만 확인했다. 실제 흡수 커밋과 충돌 해결 경로는 관측하지 않았다.
- darwin 외 OS(`/usr/bin/false` 경로 포함)는 검증 범위 밖이다.
- 교차 모델 백엔드(codex/GLM) 의견은 호출하지 않았다.

## Residual-risk

- 흡수 시 세 테스트 파일에 충돌이 생기면 AC-DTC-004가 적색이 되어 재작업이 필요할 수 있다(올바른 적색이지만 비용).
- `TMPDIR`이 저장소 안을 가리키는 비표준 환경에서는 REQ-001의 "저장소 밖" 조건이 AC로 잡히지 않는다.
- 향후 새 repository-only doctor 검사가 추가되면 REQ-002의 AC 증인이 여전히 Agent Emit Embed 하나뿐이다(O5).

## Recommendation

**PASS (0.92 ≥ 0.75).** 차단 결함 없음. must-pass 근거 요약:
- MP-1/2: REQ 6개 순차, 각 단일 `shall`의 GEARS 형식(증거 5).
- MP-3: 12필드 완비, strict lint exit 0(증거 11).
- MP-5/6/7: 교차 SPEC 참조·`syscall`·clarification 표식 모두 없음.
- MP-8: RED 재현(증거 3·4), ledger 원문 동일(증거 2).
- MP-9: 마일스톤 `Exit:`이 AC를 번호순으로 묶음.

선택 결함 O1~O5는 orchestrator 재량이다. run 진입 전에 비용 대비 가치가 큰 것은 O2(사후 병합 서술 정정)와 O3(자연 상태 근거 재측정) 두 건이다. Implementation Kickoff Approval은 이 판정과 무관하게 여전히 필요하다.
