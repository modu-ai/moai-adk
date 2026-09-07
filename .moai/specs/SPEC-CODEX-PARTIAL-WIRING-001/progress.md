# SPEC-CODEX-PARTIAL-WIRING-001 — 진행 기록

> 카드 t499 · Tier M · 3-phase(plan → run → sync)

## §E.1 Plan-phase Audit-Ready Signal

- **작성 시점 트리**: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t499`, branch `WT-codex-partial-wiring`, base `ace1c5440`
- **산출물 4종**: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`
- **Tier 판정**: M — 근거는 `plan.md` §B/§E의 반경(소스 1~2본 + 시험 1본), 요구 11건 / AC 9건(릴리스 게이트 6 + 회귀 가드 3) + 뮤턴트 8종, 그리고 사용자에게 보이는 출력 표면 변경이 포함된다는 점
- **측정 전제의 출처**: `.moai/reports/t499/repro.md`(레인 선행 측정). 이 SPEC은 그 값을 **재측정하지 않고 인용**하며, 인용 사실을 여기에 명시한다
- **plan-phase에서 실행한 확인**: SPEC ID 정규식 검사(`SPEC-CODEX-PARTIAL-WIRING-001` → `PASS`), 기존 SPEC 디렉터리 중복 없음(`ls .moai/specs/ | grep -i PARTIAL` → 무출력), 골든 하네스 재독(`internal/cli/doctor_golden_test.go`)
- **미검증 전제**: `plan.md` §I에 3건 기재 (release 일정 대조 없음 / 위저드 경로 미측정 / `moai update` 거동 미측정)

### plan-audit iter1 수리 라운드 (2026-09-07)

- **판정**: FAIL 0.76 (Tier M 임계 0.80) — 판정서 `.moai/reports/t499/plan-audit-iter1.md`. must-pass 7건 전부 PASS, FAIL은 Testability 0.60 한 축이 끌었다
  > 0.3.0 정정(감사 iter2 D3): 이 줄은 원래 "`spec lint` 무결"을 plan-audit-ready 증거로 함께 들었다. **그 증거는 공허했다** — 아래 § lint 증거의 정정 참조. 오기 사실을 지우지 않고 증거 항목만 뺀다
- **수리 대응**: D1 → `acceptance.md` AC-CPW-007에 반쪽 경로를 실제로 밟는 신규 시험 추가 / D2 → 릴리스 게이트 5건에 트리 `ace1c5440` 실측 RED-now 셀 부여 + SHA를 `acceptance.md` 머리에 pin / D3 → REQ-CPW-009 신설 + AC-CPW-002에 `initCodexAdvice` 부재 단언 / D4 → AC-CPW-001에 `claude-only` 0회 단언 / D5 → REQ-CPW-010 신설 + AC-CPW-008 재매핑 / D6 → `spec.md` §A.3 좌표 `:105` → `:92` / D7(optional, 수용) → §D-1에 보호 대상 구분 추가
- **레인 이월 2건**: `acceptance.md` 매트릭스의 축약 REQ 표기를 전체 ID로 폄; `plan.md` §I의 `phase` 갭을 리드 측정으로 격하(값은 관측과 일치, 로드맵 확인은 아님)
- **관측했으나 이 카드 범위 밖**: (1) `plan-auditor`의 추적성 검사가 축약 REQ 표기(`REQ-CPW-001, 002`)를 놓쳤다 — D5로 매핑 오류는 찾아냈으니 검사 자체는 돌았고, 축약형만 통과시켰다. (2) `moai spec lint`도 같은 표기를 잡지 못한다(`CoverageIncomplete` 미발화) — lint 규칙 쪽 후속 후보이며 이 카드에서 손대지 않는다
- **RED-now 실측 5건**(전부 이 트리에서 직접 실행, 리다이렉트 후 `echo $?`로 종료코드 판독): 다섯 셀렉터 모두 `exit=0` + `[no tests to run]` — 아직 없는 시험이라 판정 불능이며, 그 상태를 RED로 읽는다. 원문은 `acceptance.md` §D.1 각 항목

### plan-audit iter2 수리 라운드 (2026-09-07, 리드 override로 iter3 진입)

- **판정**: FAIL 0.84 (임계 0.80 통과) — 판정서 `.moai/reports/t499/plan-audit-iter2.md`. 점수 회귀 없음(0.76 → 0.84), must-pass 7/7, iter1 지적 7건 중 5건 완결·1건 부분·1건 좌표 잔존. FAIL은 점수가 아니라 `verification-completeness.md` §2 **뮤턴트 관문**에서 나왔다 — AC-CPW-002를 통과하면서 요구를 위반하는 뮤턴트가 둘 구성됐다
- **수리 대응**:
  - **D1**(major) → AC-CPW-002 Then (d)를 상수 토큰 `initCodexAdvice` 부재에서 **성질**(부분문자열 `moai init` 금지)로 재작성. 감사가 실증한 패러프레이즈 문구를 **뮤턴트 M6**으로 고정
  - **D2**(major) → **REQ-CPW-011 신설**(부재 갈래는 사용자 계층 훑기에 도달하지 않으며, 낡은 `[[skills.config]]` 항목이 선언된 home 아래에서도 `CheckOK`) + `spec.md` §A.4에 **네 번째 보존 계약**(사용자 계층 훑기의 도달 범위) 추가 + AC-CPW-002 Given (ii)에 `stubCodexHome`(낡은 항목 ≥1) 고정 + Then (e) 소견 부재 단언 + **뮤턴트 M7** 신설
  - **D3**(minor) → 아래 § lint 증거의 정정. **SPEC의 요구 표기는 바꾸지 않았다** — lint를 만족시키려고 문서를 비트는 것이므로
  - **D4**(minor) → AC-CPW-007의 좌표 `:394`/`:441` → `:404`/`:450`, 정정 사실을 §A.3 D6과 같은 방식으로 남김
  - **D5**(minor) → RED-now 5셀을 **명령 / stdout / `exit:`** 세 필드로 분해. `> file 2>&1; echo $?` 형식은 리드 승인대로 유지하되, 그것이 **값을 읽는 방법**이지 기준이 내거는 명령이 아님을 머리 주석으로 분리
  - **D6**(optional) → stdout 인용 범위를 판정 실린 부분으로 한정하고, 캐시·시간 토큰이 실행마다 달라진다는 주석 추가
  - **리드 승인 보강**(감사와 무관하게 적용) → RED-now 셀마다 `grep -c 'func <시험 이름>'` 관측을 짝지어 부재의 **원인**을 고정. 두 관측의 네 조합 판독표를 `acceptance.md` **§D.0**에 한 번 두고 다섯 셀은 참조만 한다. 어긋나는 두 조합(grep ≥1 + `[no tests to run]`, grep 0 + `--- PASS`)은 RED도 GREEN도 아닌 **측정 결함**으로 읽고 즉시 blocker로 세운다
- **lint 증거의 정정(D3)** — 직접 재측정했다(감사·리드 값을 옮기지 않았다):
  - `grep -cE '^\s*[-*]\s+\**\s*REQ-[A-Z0-9]+(-[A-Z0-9]+)*-[0-9]+' spec.md` → `0` (목록형 REQ 정의 줄)
  - `grep -cE '^\| REQ-CPW-[0-9]+' spec.md` → `10` (표 행 REQ 정의 줄)
  - lint의 REQ 수집기는 목록 항목 줄만 잡는다(`internal/spec/lint_req_widen.go:59` `reqLineWidePattern` = `^\s*[-*]\s+…`). 이 SPEC의 요구는 전부 표 행이므로 `doc.REQs`가 비고, **REQ 기반 4규칙(`CoverageIncomplete` / `ModalityMalformed` / `InvalidREQID` / `DuplicateREQID`)은 이 SPEC을 한 번도 방문하지 않았다.**
  - 따라서 **`✓ No findings`는 통과가 아니라 미실행이다.** 이 SPEC의 REQ→AC 커버리지는 lint가 아니라 직접 측정(전체 ID 대조)으로만 성립한다. 이전 라운드에서 lint 무결을 plan-audit-ready 증거로 든 문장은 이 사실을 몰랐고, 그 인용은 철회한다
- **관측했으나 이 카드 범위 밖**(리드가 후속 카드로 발행 — 큐 미접촉):
  1. **lint 수집기 사각지대** — 표 형태로 요구를 쓰는 SPEC 전체가 REQ 기반 4규칙의 사각지대다. 원인은 표기 축약이 아니라 **정의 줄 형태**다(iter1 라운드에서 내가 "축약 표기"로 적은 원인 진단은 감사가 정정했다)
  2. **plan-auditor 추적성 검사의 입도** — iter1은 전체 ID 단위로 대조하지 않아 축약 표기를 통과시켰고, 같은 문장에서 lint의 침묵을 방증으로 인용해 **같은 맹점을 가진 두 관측**을 독립 확인처럼 썼다
- **좌표 재측정**(이 카드에서 좌표 오기가 두 번 났으므로 커밋 직전 전수 재측정): `grep -n 'checkCodexWiring(t.TempDir()' internal/cli/doctor_codex_test.go` → `155 / 404 / 450 / 552 / 567`; `grep -n 'wired := hooksErr' internal/cli/doctor_codex.go` → `92`; `codexStaleSkillFinding` 호출 `:189` · 함수 선언 `:376`; `not wired (claude-only project)` 문자열 `:101`; `TestCheckCodexWiring_MessageWidthStaysInBand` 선언 `:390`, `..._RenderedPanelStaysInBand` 선언 `:437`; `stubCodexHome(` 호출 `29`곳

### 허용 목록 전환 (2026-09-07, 리드 질의 → 채택)

- **전환**: AC-CPW-002 (d)를 금지 목록(denylist) → **등가 단언(allowlist)**으로 바꿨다. 부재 갈래 Message는 시험 파일의 리터럴 한 개와 정확히 같아야 한다. 어떤 패러프레이즈를 새로 발명하든 리터럴과 다르므로 **원리상 통과하지 못한다**
- **리드가 물은 세 가지에 대한 답**:
  1. **상수로 표현 가능한가 — 가능하다.** Message에 프로젝트별 가변부가 필요 없다: 경로는 Detail로 내리는 것이 이 파일의 기존 관용구이고(`internal/cli/doctor_codex.go` `:101` unwired-skip, `:195` wired-OK가 이미 평문 리터럴), 개수는 REQ-CPW-005가 판별식에서 금지한다. 조치 문구를 이름 붙인 상수로 두는 관례도 이미 있다(`reTrustAdvice` `:46`, `initCodexAdvice` `:52`)
  2. **다른 단언과 충돌하는가 — 충돌 없음, 부분 중복만.** 폭 상한(AC-CPW-007)은 등가 단언이 덮는 부재 갈래 **밖**(codex 존재 갈래는 `joinCodexSummaries` 합성을 거친다)에서 여전히 필요하다. `claude-only` 0회(c)는 Message에 대해서는 등가에 포섭되지만 **Detail까지** 재므로 남긴다. 중복은 남기는 쪽이 싸다
  3. **M6·M7이 사소해지는가 — M6은 사소해지고 M7은 그대로 의미 있다.** M6은 이제 "리터럴과 다르다" 한 줄로 잡히고(관문이 세진 것이므로 좋다), M7은 Status가 Warn으로 뒤집히는 축이라 등가 단언과 별개로 남는다. 더해 **M6′**(세 금지 토큰을 전부 피한 지시문)을 신설했다 — 종전 처리로는 통과했고 등가 단언으로는 RED가 되는, 이 전환의 근거 그 자체인 뮤턴트다
- **닫히지 않은 잔여분(정직하게 기록)**: 등가 단언은 *구현이 리터럴에서 벗어나는 것*을 완전히 막지만 *리터럴로 지시문을 고르는 것*은 막지 못한다. 그 선택은 run-phase 저자의 한 번의 판단이며, 이제 **diff에 보이는 코드 한 줄**에 모인다. (d′)의 세 토큰 모양 검사가 그 한 줄에 대한 자동 1차 검사다. 리드의 "완전히 닫힌다"는 표현은 *패러프레이즈 뮤턴트에 대해서는* 참이고, *리터럴 선택 자체*에 대해서는 참이 아니다 — 이 구분을 남긴다
- **제거**: 사람 판정 DoD 항목("문구를 사람이 읽고 지시문인지 판정")은 지웠다. 반증하기 어려운 체크리스트 항목이었고, 그 판단 지점이 코드 한 줄로 옮겨갔다

### 리드 어셈블리 판독에 대한 확인과 반박 1건 (2026-09-07)

- **확인**: `internal/cli/doctor_codex.go:193-201`을 직접 읽었고 리드 판독과 일치한다 — `problems`가 비면 `CheckOK` + 고정 문자열, 하나라도 차면 `CheckWarn` + `joinCodexSummaries(problems)`. **등가 단언 성립**
- **반박 1건**: "등가는 `stubCodexHome`이 **깨끗한** home일 때만 성립한다"는 **정정이 필요하다.** 그것은 *결함이 있는 구현*에서만 참이다. REQ-CPW-011대로 이 갈래가 훑기에 도달하지 않으면 `problems`는 home 내용과 무관하게 비므로, **낡은 home에서도 등가가 성립해야 한다.** 깨끗한 home으로 바꾸면 M7(훑기 흘려보내기)이 fail-open으로 초록이 되어 감사 D2가 지적한 "CI에서만 초록"이 되살아난다 — 즉 낡은 home 쪽이 더 강한 픽스처다
- **처리**: 픽스처를 약화시키는 대신 **하위 케이스를 갈랐다** — (i) 깨끗한 home / (ii) 낡은 home, (a)·(d)를 둘 다에서 단언. 리드가 걱정한 진단 가능성은 이 분리가 준다: (i)만 실패 → 리터럴이 달라짐(M6/M6′), (ii)만 실패 → 훑기 누수(M7), 둘 다 실패 → 갈래 처리 자체가 다름(M1). 헬퍼도 읽어 확인했다 — `stubCodexHome`(`:76`)은 해석기를 돌릴 뿐이고 낡은 항목은 `writeCodexHomeConfig`(`:95`)가 만들므로 픽스처가 서로 오염되지 않는다
- **리드 질의 2(폭 상한)**: AC-CPW-007을 등가에 **흡수시키지 않는다.** 부재 갈래에서는 폭이 리터럴의 성질이 되어 덧붙는 값이 작지만, **codex 존재 갈래는 `joinCodexSummaries` 합성이라 흡수 자체가 불가능**하다 — 폭이 하중을 받는 곳이 거기다. 두 갈래를 함께 재는 하나의 기준으로 유지하고 근거를 AC 본문에 남겼다
- **(d′) 유지 결정(리드 권고와 다름)**: 리드는 금지 목록 세 토큰과 "충분조건 아님" 문구를 잉여로 보고 삭제를 권했으나, **(d)와 (d′)는 서로 다른 대상을 잰다** — (d)는 런타임 Message가 리터럴에서 벗어났는지, (d′)는 그 리터럴로 무엇을 골랐는지. (d)가 아무리 강해도 후자에 닿지 못하므로 삭제하면 그 축이 무방비가 된다. 대신 "약한 검사가 함께 있다"는 인상을 없애도록 대상 차이를 명시하는 서술로 바꿨다. 사람 판정 DoD 항목은 지시대로 삭제했다

_run-phase 진입 전 상태: `status: draft`._

## §E.2 Run-phase Evidence

> 측정 트리: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t499`, branch
> `WT-codex-partial-wiring`. 착수 시 HEAD `5a600abe4`(base `ace1c5440`), M1+M2 커밋 후
> `fa352752d`. 아래 모든 값은 **이 실행에서, 이 트리에 대해** 직접 측정했다.
> 종료코드는 전부 `<cmd> > out.txt 2>&1; echo $?` 로 읽었다(파이프 뒤 `$?` 금지).

### 기준선 (착수 전, 트리 `5a600abe4`)

| 명령 | 종료코드 | 판독 |
|---|---|---|
| `go build ./...` | `0` | — |
| `GOOS=windows GOARCH=amd64 go build ./...` | `0` | — |
| `go test ./internal/cli/...` | `0` | 17개 패키지 전부 `ok` (`internal/cli` 355.207s). 이후의 적색은 전부 내게 귀속된다 |

### 문구 결정 (plan.md §B-3 — run-phase 저자 결정, 결정 즉시 고정)

- 부재 갈래 Message 리터럴: **`codex agent definitions present, no wiring files — codex not on PATH`** (68 runes)
- 구현 상수 `halfWiredAbsentMessage`(`internal/cli/doctor_codex.go:71`) = `halfWiredSummary`(`:59`) + `" — codex not on PATH"`
- 시험 리터럴 `wantHalfWiredAbsentMessage`(`internal/cli/doctor_codex_test.go:954`) — 등가 단언으로 못박힘(AC-CPW-002 (d))
- 존재 갈래 Message: `halfWiredSummary + " — " + initCodexAdvice` (78 runes)
- (d′) 모양 검사: 리터럴에 `moai init` / `run ` / `install` 부분문자열 **0회** — `TestCheckCodexWiringHalfWiredAbsentLiteralCarriesNoDirective`(`:1076`)가 기계적으로 잰다

### AC 판정 매트릭스

세 관측을 함께 싣는다(§D.4): `--- PASS` · `[no tests to run]` **부재** · `grep -c 'func <이름>'` ≥ 1.
전체 시험군 재실행(`-count=1`, 뮤턴트 전수 원복 후) 결과는 `--- PASS` **44건**, `[no tests to run]` **0건**,
`ok github.com/modu-ai/moai-adk/internal/cli 6.628s`, exit `0`.

| AC | 성격 | 판정 | 명령 / 관측 |
|---|---|---|---|
| AC-CPW-001 | 게이트 | **PASS** | `go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredCodexInstalled -v` → `--- PASS`, `[no tests to run]` 부재, exit `0`; `grep -c` → `1` |
| AC-CPW-002 | 게이트 | **PASS** | `go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredCodexAbsent -v` → 하위 케이스 (i) `clean_home` · (ii) `stale_home` 둘 다 `--- PASS`, exit `0`; `grep -c` → `1` |
| AC-CPW-003 | 회귀 가드 | **PASS** | `go test ./internal/cli/ -run 'TestCheckCodexWiring_(ClaudeOnlyMachineStaysSilent\|InactiveProjectInformationalSkip\|UnwiredWithCodexInstalledWarns)' -v` → 세 줄 모두 `--- PASS`, exit `0`. **세 시험 본문 미수정** — `git diff` 무출력 |
| AC-CPW-004 | 회귀 가드 | **PASS** | `go test ./internal/cli/ -run TestDoctorGolden -v` → `Light`/`Dark`/`NoColor` 세 줄 `--- PASS`, exit `0`; `git diff --stat internal/cli/testdata/` **무출력** |
| AC-CPW-005 | 게이트 | **PASS** | `go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredCountIndependent -v` → `1_definitions` · `12_definitions` · `empty_agents_directory` 세 하위 케이스 `--- PASS`, exit `0`; `grep -c` → `1`. 더해 `grep -nE '\b11\b' internal/cli/doctor_codex.go` → **무출력**(exit `1`) |
| AC-CPW-006 | 게이트 | **PASS** | `go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredReadOnly -v` → `--- PASS`, exit `0`; `grep -c` → `1`. 실물 확인: 두 갈래 `moai doctor` 모두 `rc=0`(아래 M4) |
| AC-CPW-007 | 게이트 | **PASS** | `go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredMessageWidthStaysInBand -v` → `--- PASS`, exit `0`; `grep -c` → `1`. 기존 가드 2종(`MessageWidthStaysInBand` / `RenderedPanelStaysInBand`)도 `--- PASS` |
| AC-CPW-008 | 회귀 가드 | **PASS** | `git diff -U0 internal/codexwiring/wire.go` **무출력** — `RefreshWiring` / `wireProject` / `wiringFilesExist` 무접촉. `internal/codexwiring/` 변경은 `codexwiring.go` 상수 6줄 추가뿐(`git diff --stat` → `1 file changed, 6 insertions(+)`); 추가된 줄은 `AgentsRelPath`(`:37`)와 그 주석으로, 세 함수 어디서도 참조되지 않는다. `go test ./internal/codexwiring/` → `ok … 1.207s`, exit `0` |
| AC-CPW-009 | 게이트(반증 관문) | **PASS** | 뮤턴트 8종 전부 RED 관측 — 아래 표 |

### RED 증거 (GREEN 이전에 포착 — TDD 반증 가능성)

`internal/codexwiring`에 `AgentsRelPath` 상수만 먼저 넣어(컴파일 가능하게) 시험을 돌렸다. 판별식은 아직 손대지 않았다.

```
$ go test ./internal/cli/ -run 'TestCheckCodexWiring_HalfWired|TestCheckCodexWiringHalfWired' -v   ; exit: 1
=== RUN   TestCheckCodexWiring_HalfWiredCodexInstalled
    doctor_codex_test.go:999: Message does not state "agent definitions": "codex installed, project not wired — run moai init --agent codex"
    doctor_codex_test.go:999: Message does not state "no wiring files": "codex installed, project not wired — run moai init --agent codex"
--- FAIL: TestCheckCodexWiring_HalfWiredCodexInstalled (0.00s)
=== RUN   TestCheckCodexWiring_HalfWiredCodexAbsent/clean_home
    doctor_codex_test.go:1055: Message = "not wired (claude-only project) — skipped", want exactly "codex agent definitions present, no wiring files — codex not on PATH"
    doctor_codex_test.go:1059: half-wired project described as claude-only: {… Message:not wired (claude-only project) — skipped Detail:}
=== RUN   TestCheckCodexWiring_HalfWiredCodexAbsent/stale_home
    doctor_codex_test.go:1055: Message = "not wired (claude-only project) — skipped", want exactly "codex agent definitions present, no wiring files — codex not on PATH"
--- FAIL: TestCheckCodexWiring_HalfWiredCodexAbsent (0.01s)
=== RUN   TestCheckCodexWiring_HalfWiredCountIndependent/1_definitions
    doctor_codex_test.go:1095: 1 definitions did not classify as half-wired: "not wired (claude-only project) — skipped"
=== RUN   TestCheckCodexWiring_HalfWiredCountIndependent/12_definitions
    doctor_codex_test.go:1095: 12 definitions did not classify as half-wired: "not wired (claude-only project) — skipped"
--- FAIL: TestCheckCodexWiring_HalfWiredCountIndependent (0.01s)
    --- PASS: TestCheckCodexWiring_HalfWiredCountIndependent/empty_agents_directory (0.00s)
--- PASS: TestCheckCodexWiring_HalfWiredReadOnly (0.00s)
--- PASS: TestCheckCodexWiring_HalfWiredMessageWidthStaysInBand (0.01s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.920s
```

**정직하게 기록할 두 가지.** `HalfWiredReadOnly`와 `HalfWiredMessageWidthStaysInBand`는 **구현 전에도 초록이었다**.
`acceptance.md`가 이 둘에 부여한 RED-now 셀은 *"시험이 아직 존재하지 않는다"*였고, 시험을 만든 순간 그 RED는 소진된다 —
둘 다 보존 성질(읽기 전용·폭 상한)을 재는 기준이라 원리상 행위 적색을 가질 수 없다. 이 둘이 실제로 판별식을
잡고 있다는 증거는 **뮤턴트 M5(폭)와 M2/M1(경로 도달)** 이지, 이 실행의 초록이 아니다. 그 초록만으로
"진전을 증명했다"고 읽어서는 안 된다.

### 뮤턴트 8종 (AC-CPW-009 — 이 SPEC에서 처음 집행된 반증 관문)

각각 심고 → RED 관측 → 원복. 전수 원복 후 `git diff --stat internal/cli/` **무출력** 확인.

| # | 심은 변형 | 관측 | RED가 된 시험 |
|---|---|---|---|
| M1 | `halfWired := false`(agent 조회 제거 — 배선 파일 2종만 보는 형태로 회귀) | exit `1` | `..._HalfWiredCodexInstalled`(문구 2건), `..._HalfWiredCodexAbsent`(두 하위 케이스), `..._HalfWiredCountIndependent`(1·12) |
| M2 | `codexAgentDefinitionsExist`를 "디렉터리가 존재하는가"로 교체 | exit `1` | `..._HalfWiredCountIndependent/empty_agents_directory` — `an empty agents directory classified as half-wired` |
| M3 | 부재 갈래 Message를 `not wired (claude-only project) — skipped`로 되돌림 | exit `1` | `..._HalfWiredCodexAbsent` 두 하위 케이스 — 등가 단언 + `claude-only` 0회 단언 동시 파괴 |
| M4 | 부재 갈래 Message 끝에 `— run moai init --agent codex` 부착 | exit `1` | `..._HalfWiredCodexAbsent` 두 하위 케이스 — 등가 단언 |
| M5 | 반쪽 갈래 Message를 200 rune로 확대 | exit `1` | `..._HalfWiredMessageWidthStaysInBand` — `codexFound=true` 253 runes / `codexFound=false` 218 runes, 둘 다 113 초과 |
| **M6** | 감사가 실증한 패러프레이즈 지시문(`` `moai init --agent codex` wires them ``) | exit `1` | `..._HalfWiredCodexAbsent` 두 하위 케이스 — 등가 단언 |
| **M6′** | 세 금지 토큰을 **전부 피한** 지시문(`wire them once codex is available`) | exit `1` | `..._HalfWiredCodexAbsent` 두 하위 케이스 — 등가 단언 |
| **M7** | 부재 갈래를 사용자 계층 훑기로 흘려보냄 | exit `1` | `..._HalfWiredCodexAbsent` **하위 케이스 (ii) `stale_home`만** — (a) `status = warn` · (d) 등가 · (e) `skills.config`/`stale skill` 등장, 셋이 함께 파괴. **(i) `clean_home`은 PASS로 남았다** |

**M5에 딸린 부수 관측(감사 iter1 D1의 실증).** M5를 심은 채 기존 폭 가드 2종을 돌리면
`go test ./internal/cli/ -run 'TestCheckCodexWiring_(MessageWidthStaysInBand|RenderedPanelStaysInBand)' -v` → 둘 다 `--- PASS`, exit `0`.
빈 `t.TempDir()`로 `unwired` 경로만 밟기 때문이다. 신규 폭 시험이 없었다면 200-rune 뮤턴트는 통과했다.

**M6′와 M7이 이번 라운드 수리의 합격 조건이었다는 사실의 확인.**
M6′의 문구 `codex agent definitions present, no wiring files — wire them once codex is available` 은
`moai init` · `run ` · `install` 세 금지 토큰을 하나도 담지 않는다 — 즉 **종전의 금지 목록 처리로는 통과했을**
지시문이며, 등가 단언(allowlist)으로 전환한 덕분에만 RED가 된다. 0.3.1 전환이 산 것이 무엇인지의 직접 증거다.
M7의 비대칭 — (ii)만 적색, (i)는 초록 — 은 감사 D2가 지적한 *"깨끗한 CI에서만 초록"* 결함 그 자체이며,
`acceptance.md` Given (ii)의 낡은 home 픽스처를 빼면 이 뮤턴트는 전부 초록으로 통과한다.

### M4 — 실물 바이너리 확인

`make build`(트리 `fa352752d`, `bin/moai` = `list-284-gfa352752d`, built 2026-09-07T05:29:53Z) 후,
`moai init --non-interactive`로 만든 실제 반쪽 프로젝트에 대해 두 PATH 갈래를 실행했다.
`moai doctor`는 프로세스 cwd에서 프로젝트 루트를 읽으므로(`internal/cli/doctor.go:244`) cwd 고정을 위해
`.moai/reports/t499/run-doctor-m4.sh`를 경유했다.

- 반쪽 전제 실측: `find <proj>/.codex -type f | wc -l` → `11`(전부 `.codex/agents/moai/*.toml`);
  `ls .codex/hooks.json .codex/config.toml` → 둘 다 `No such file or directory`
- **codex 존재** — `rc=0`, 증거 `.moai/reports/t499/doctor-halfwired-codex-present.txt:103`
  ```
  │    warn    Codex Wiring          codex agent definitions present, no wiring files — run moai init --agent codex (+1 more, see --verbose)  │
  ```
- **codex 부재**(`PATH=/usr/bin:/bin`) — `rc=0`, 증거 `.moai/reports/t499/doctor-halfwired-codex-absent.txt:103`
  ```
  │    ok      Codex Wiring          codex agent definitions present, no wiring files — codex not on PATH  │
  ```

두 갈래 모두 `rc=0` — REQ-CPW-007의 종료코드 절 확인(`CheckWarn`은 종료코드를 움직이지 않는다).
`.moai/reports/t499/repro.md`가 기록한 두 거짓 문장은 사라졌다: `claude-only` 단언도, agent 정의의 존재를
지우는 문구도 더 이상 나오지 않는다.

> **부수 관측(내 변경 소관 아님).** codex-부재 실행의 패널이 더 넓다. 최광폭 행을 실측하면
> `ast-grep CLI  sg not found — …` 178열이며, `PATH`를 깎아 `sg`가 사라진 탓이다. `Codex Wiring` 행이 아니다.

### 좌표 재측정 (커밋 직전 전수 — 이 카드에서 좌표 오기가 두 번 났다)

| 좌표 | 측정 명령 | 값 |
|---|---|---|
| `internal/cli/doctor_codex.go` 판별식 | `grep -n 'wired := hooksErr'` | `111` |
| 반쪽 판별 | `grep -n 'halfWired := '` | `120` |
| `halfWiredSummary` | `grep -n 'halfWiredSummary = '` | `59` |
| `halfWiredAbsentMessage` | `grep -n 'halfWiredAbsentMessage = '` | `71` |
| `codexAgentDefinitionsExist` 선언 | `grep -n 'func codexAgentDefinitionsExist'` | `268` |
| `halfWiredDetail` 선언 | `grep -n 'func halfWiredDetail'` | `290` |
| `AgentsRelPath` (`internal/codexwiring/codexwiring.go`) | `grep -n 'AgentsRelPath'` | `32`(주석) / `37`(상수) |
| 시험 리터럴 (`internal/cli/doctor_codex_test.go`) | `grep -n 'wantHalfWiredAbsentMessage = '` | `954` |
| 픽스처 헬퍼 | `grep -n 'func halfWiredProject'` | `961` |
| 신규 시험 6종 선언 | `grep -n 'func TestCheckCodexWiring_HalfWired\|func TestCheckCodexWiringHalfWired'` | `987` / `1029` / `1076` / `1088` / `1115` / `1167` |
| doctor 호출부 | `grep -n 'checkCodexWiring' internal/cli/doctor.go` | `244` |

### 범위 확인

`git diff --name-only 5a600abe4..HEAD`(run-phase 착수 시점 대비 — 즉 **내 반경만**) 출력 4건:

```
.moai/specs/SPEC-CODEX-PARTIAL-WIRING-001/spec.md
internal/cli/doctor_codex.go
internal/cli/doctor_codex_test.go
internal/codexwiring/codexwiring.go
```

여기에 이 커밋의 `progress.md` + `.moai/reports/t499/` 증거 3본이 더해진다.
`spec.md`는 `status:` 한 필드만(draft → in-progress) 바뀌었고 본문은 미접촉.
`plan.md` / `acceptance.md`는 **무접촉**. `.moai/state/` · `.moai/cache/` · `.moai/logs/` ·
다른 SPEC 디렉터리는 어느 것도 건드리지 않았다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-07
run_commit_sha: 5b436ff5d   # M1+M2 = fa352752d, M3+M4 evidence = 5b436ff5d
run_status: complete
ac_pass_count: 9
ac_fail_count: 0
release_gate_pass: 6      # AC-CPW-001 / -002 / -005 / -006 / -007 / -009
regression_guard_pass: 3  # AC-CPW-003 / -004 / -008
mutants_planted: 8
mutants_red: 8
mutants_reverted: 8       # git diff --stat internal/cli/ 무출력으로 확인
preserve_list_post_run_count: 4   # 존재-게이트 / 읽기 전용·비차단 / un-nagging / 사용자 계층 훑기 도달 범위
tests_added: 6
tests_modified: 0
red_evidence_captured: true       # GREEN 이전 포착, §E.2에 verbatim
cross_platform_build:
  darwin_arm64: pass              # go build ./... exit 0
  windows_amd64: pass             # GOOS=windows GOARCH=amd64 go build ./... exit 0
lint:
  command: golangci-lint run --timeout=5m ./internal/cli/... ./internal/codexwiring/...
  result: "0 issues."
  exit: 0
  new_warnings_or_lints_introduced: 0
vet:
  command: go vet ./internal/cli/... ./internal/codexwiring/...
  exit: 0
gofmt:
  command: gofmt -l internal/cli internal/codexwiring
  output: ""                      # 무출력
coverage:
  internal/cli: "80.7%"           # 패키지 기준선 — 이 카드가 낮추지 않았다(측정만)
  internal/codexwiring: "88.2%"
full_touched_package_suite:
  command: go test ./internal/cli/... ./internal/codexwiring/...
  exit: 0
golden_fixtures_changed: 0        # git diff --stat internal/cli/testdata/ 무출력
existence_gate_touched: false     # git diff -U0 internal/codexwiring/wire.go 무출력
real_binary_confirmation:
  binary: "list-284-gfa352752d"
  doctor_codex_present_rc: 0
  doctor_codex_absent_rc: 0
total_run_phase_files: 3          # 소스 2 + 상수 1 (SPEC/증거 산출물 제외)
m1_to_mN_commit_strategy: "M1+M2 한 커밋(fa352752d) + M3/M4 증거 커밋"
pushed: false                     # 레인은 push하지 않는다 — 리드 일괄
```

**미검증분(Gap) — 정직하게 남긴다.**

- 전체 스위트(`go test ./...`)는 **로컬에서 돌리지 않았다**(CLAUDE.local.md §4/§6). 전 패키지 판정은 CI 몫이며, 이 카드가 건드린 두 패키지 밖의 영향은 관측되지 않았다.
- `GOOS=windows` 는 **빌드만** 통과했다. windows에서 시험을 컴파일·실행하지는 않았다.
- 대화형 위저드 경로와 `moai update`의 반쪽 프로젝트 거동은 여전히 미측정(`plan.md` §I 그대로).
- 실물 확인은 darwin/arm64 한 대에서만 했다.

**잔여 위험.**

- 등가 단언은 *구현이 리터럴에서 벗어나는 것*을 막지만 *리터럴로 무엇을 고르는가*는 (d′)의 세 토큰 모양 검사까지만 잡는다. 그 세 토큰을 모두 피한 새 지시문을 리터럴로 고르는 경우는 여전히 열려 있고, 이제 diff에 보이는 코드 한 줄에 모인다 — 0.3.1이 기록한 잔여분 그대로이며 이번 라운드가 닫은 것이 아니다.
- `codexAgentDefinitionsExist`는 fail-open이라 권한 오류로 읽히지 않는 `.codex/agents/`는 "정의 없음"으로 읽힌다. 이 검사가 조언용이라 의도한 방향이지만, 그런 프로젝트는 반쪽 상태여도 종전 문구를 받는다.
- 폭 상한은 `Message`에 대해서만 신규 시험이 있다. 반쪽 갈래의 `--verbose` 렌더 폭은 기존 `RenderedPanelStaysInBand`가 `unwired` 경로만 밟으므로 여전히 미관측이다.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-07
sync_commit_sha: pending-backfill-sync   # this commit cannot cite its own hash; backfilled in the following commit
sync_status: complete
changelog_entry_position: "[Unreleased] > ### Fixed (first entry)"
changelog_duplicate_check:
  command: "grep -c 'SPEC-CODEX-PARTIAL-WIRING-001' CHANGELOG.md"
  pre_write_output: 0
docs_surfaces_checked:
  - surface: "README.md (+ .ko/.ja/.zh)"
    result: "no mention of the Codex Wiring check anywhere; nothing to update"
  - surface: "docs-site/content/{en,ko,ja,zh}/cli-reference/doctor.md"
    result: "documents only two checks by name (Home Disk Usage, Hook Delivery); Codex Wiring was never documented on this surface either before or after SPEC-CODEX-WIRING-001 — nothing to update, no 4-locale obligation triggered"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (this commit)"
  updated_field: "unchanged (2026-09-07, already current)"
b12_self_test:
  a_duplicate_grep: "PASS (0 before write)"
  b_ac_count_match: "PASS (9 distinct AC-CPW-* in acceptance.md == 9 cited in progress.md §E.3 ac_pass_count)"
  c_file_path_verification: "PASS (ls internal/cli/doctor_codex.go internal/cli/doctor_codex_test.go internal/codexwiring/codexwiring.go all resolve)"
```

**Sync-phase scope.** Modified: `CHANGELOG.md` (Unreleased > Fixed), this file's §E.4, `spec.md` frontmatter `status:` only (`in-progress → completed`). Untouched: `plan.md`, `acceptance.md`, all `spec.md` body content, implementation source (already closed at run-phase), `.moai/state/`, `.moai/cache/`, `.moai/logs/`, other SPEC directories.
