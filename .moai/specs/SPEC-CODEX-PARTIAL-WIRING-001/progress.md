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

_run-phase 진입 전 상태: `status: draft`._

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
