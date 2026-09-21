# SPEC-APPJS-IIFE-GUARD-001 — 진행 기록

- SPEC: SPEC-APPJS-IIFE-GUARD-001
- card: t1048
- worktree: `.claude/worktrees/t1048` (branch `WT-iife-guard`)
- plan-phase base: `d8304b49a`

---

## §E.1 Plan-phase Audit-Ready Signal

**Claim** — plan-phase 산출물 4종(spec.md / plan.md / acceptance.md / progress.md)이 작성됐고, SPEC 이 딛고 선 실측값은 이 트리에서 이번 실행으로 측정됐다.

**Evidence**

- SPEC ID 정규식 자가검사: `[[ "SPEC-APPJS-IIFE-GUARD-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS`
- IIFE 스팬: `grep -n '^[^[:space:]]' internal/web/assets/app.js` → 컬럼 0 IIFE 둘, `[(39,660),(673,796)]`
- 핸들러 인벤토리: `grep -n 'addEventListener' internal/web/assets/app.js` → 20건, 그중 bare identifier 4건 (530·548·551·727)
- 선언 위치: `grep -n -E '(function|var|let|const)[[:space:]]+(syncSegmentsVisibility|initConsole|stampRefreshed)\b'` → 45 / 526 / 711
- 최상위 문장 형태 전수조사 → 그 밖의 형태 **0** (상세는 spec.md §B.2)
- 프로토타입 2방향: 정방향 `checked=4, violations=0, EXIT=0` / 돌연변이 `checked=4, violations=1, EXIT=1` (`line 552: stampRefreshed — declared in IIFE#[1] not #0`)
- 의존성: 직접 30 / 전체 require 89, JS 엔진 모듈 0건
- 수리 커밋 조상 확인: `git merge-base --is-ancestor 18fa0d384 HEAD` → `YES`

**Baseline-attribution** — 전부 worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1048`, branch `WT-iife-guard`, HEAD `d8304b49a` 에서 이번 plan-phase 실행 중에 측정.

**Gaps**

- `go test ./internal/web/` 를 **돌리지 않았다** — plan-phase 는 저작만 하며, 기존 패키지 테스트의 현재 상태는 이 실행에서 측정되지 않았다.
- 돌연변이는 스크래치패드의 사본에 대해서만 합성했다. 작업 트리의 `app.js` 는 읽기만 했고 바이트 하나 바뀌지 않았다.
- 규칙이 spec.md §F 의 세 한계를 못 잡는다는 것은 **선언**이며, 그 방향으로 돌연변이를 돌려 확인하지는 않았다.

**Residual-risk**

- 최상위 문장 형태 전수조사는 파서가 아니라 줄 형태 기준이다. 여러 줄 문장의 이어지는 줄이 다른 형태를 숨기고 있을 가능성은 배제되지 않았다.
- 프로토타입은 Python 이고 구현은 Go 다. 정규식 방언 차이(특히 `$` 의 단어 문자 취급)가 이식 과정에서 결과를 바꿀 수 있다 — run-phase 에서 2방향을 **Go 로 다시** 재야 한다.

---

## §E.1b Run-phase 진입 블로커 (2026-09-21)

> 이 절은 run-phase **증거가 아니라 run-phase 가 시작되지 않은 이유**다. §E.2 / §E.3 은
> manager-develop 소유이며 run 이 실제로 시작될 때까지 비워 둔다.

**Claim** — Kickoff 승인은 났으나 run-phase 에 진입하지 않았다. 배차문의 `cmd` 와 운영자
결정이 서로 다른 기전을 가리켜, 그대로 실행하면 자기 인수조건에 불합격하는 구현이 나온다.

**Evidence**

배차문(2026-09-21): `cmd: /moai run SPEC-APPJS-IIFE-GUARD-001` · `결정: 정적 참조 금지 방식은
채택 안 함 — 버튼 핸들러가 실제로 발화하는지를 런타임으로 잰다`.

이 SPEC 의 요구사항 10개는 전부 정적 규칙이며, 런타임 탐침 구현과 최소 6개가 충돌한다
(`grep -oE 'REQ-AIG-[0-9]+' spec.md | sort -u` → 001..010):

| 요구사항 | 조항 | 런타임 구현 시 |
|---|---|---|
| REQ-AIG-002 | bare identifier 핸들러의 IIFE 스코프 검사 | 기전 자체가 다름 |
| REQ-AIG-003 | 실패 메시지에 식별자명 + 등록 줄번호 | 런타임은 그 좌표를 못 냄 |
| REQ-AIG-004 | 검사한 핸들러 **개수** 단언 | 해당 개념 부재 |
| REQ-AIG-005 | 함수 표현식 핸들러 검사 제외 | 해당 개념 부재 |
| REQ-AIG-006 | `readEmbeddedAsset` 로 소스 읽기 | 런타임은 소스를 읽지 않고 실행함 |
| REQ-AIG-007 | 신규 `go.mod` 의존성 금지, 표준 라이브러리만 | 브라우저 구동이 이 조항 위반 |

CI 비용 측정(운영자 결정의 실제 입력):

- 워크플로 전수: `ls .github/workflows/` → 18개. 브라우저 grep 적중 1건은
  `test-install.yml:259` 의 주석 문구 `release browser endpoints` 로 **브라우저가 아니다**.
  즉 CI 브라우저 인프라는 **0 에서 시작**한다.
- 테스트 잡: `ci.yml:125` → `os: [ubuntu-latest]` 단독(SPEC-V3R6-CI-PR-SPEEDUP-001 로 좁혀짐),
  본체는 `ci.yml:210` 의 `go test ./...` 한 줄, race 는 `:292`.
- `e2e/` 트리: `e2e/cli/tux3_journeys.sh` 하나뿐 — 셸 CLI 여정, 브라우저 없음.
- t1041 탐침 의존성: 표준 `asyncio`/`json`/`sys`/`urllib.request` + **`websockets`**(서드파티,
  로컬 `15.0.1` 설치 확인) + CDP 가 `127.0.0.1:9333` 에 열린 Chrome + 라이브 서버.
- 따라서 CI 편입 비용 = ubuntu 잡에 Chrome 설치 · python 패키지 설치 · 서버 기동/종료 ·
  포트 할당이 붙고, `go test` 한 줄이던 잡이 별도 실행 계통을 갖는다.

**Baseline-attribution** — 전부 이 트리(`.claude/worktrees/t1048`, `WT-iife-guard`,
HEAD `8628672d8`)에서 2026-09-21 이 실행 중에 측정. 배차문 문면은 수신 메시지 그대로.

**Gaps**

- 런타임 방식을 **실제로 구현해 보지 않았다.** 충돌 판정은 SPEC 요구사항 문면과 탐침
  의존성 대조에서 나온 것이지, 구현해서 인수조건에 걸리는 것을 본 것이 아니다.
- CI 편입의 **소요 시간·요금을 재지 않았다.** 늘어나는 항목을 셌을 뿐이다.
- `t.Skip` 대안이 실제로 CI 에서 조용히 건너뛰는지 **돌려서 확인하지 않았다.**

**Residual-risk**

- 경로 A/B/C 중 무엇을 고르든 SPEC 재작성이 선행되어야 하며, 재작성 없이 구현하면
  인수조건이 구현을 검증하지 못한다 — 통과가 아무것도 보장하지 않는 상태가 된다.
- 이 기록 시점에 구현 코드는 0줄이다. 판별식은 **커밋 열거**다 —
  `git log develop..HEAD --name-only` → 이 브랜치의 커밋 2개가 건드린 파일은
  `.moai/specs/SPEC-APPJS-IIFE-GUARD-001/` 4개뿐, Go 파일 0개. 되돌리는 비용이 가장 싼
  지점이며, 결정이 늦어질수록 싸지지는 않는다.

  > **쓰면 안 되는 판별식**: `git diff --stat develop..HEAD -- 'internal/**/*.go'`.
  > `git diff` 는 `A..B` 철자를 포함해 전부 **트리 비교**라 내 변경과 남의 변경을 가르지
  > 못한다. 이 기록을 쓰는 동안 실제로 걸렸다 — 그 명령이 8행을 냈고, 확인해 보니 develop 이
  > `d8304b49a` → `db2270127` 로 11커밋 움직이며 들어온 **다른 레인의 변경**이었다.
  > 내 기반과 develop 이 같은 순간에만 우연히 0을 낸다.

**배차문 결함 기록 (리드 요청)** — 리드가 운영자 결정을 전달하면서 그 결정이 기존 SPEC 과
양립하는지 대조하지 않았고, `cmd` 에 기존 SPEC ID 를 그대로 실었다. 리드가 결함을 인정했다.
같은 계열: 배차문의 수치·전제를 열어 보지 않고 옮긴 건. 이 카드에서만 두 번째다(첫 번째는
새 워크트리 기반을 `d8304b49a` 로 적었으나 실제로는 `origin/develop` = `116820f40` 에서
태어난 건, `git rev-list --count HEAD..develop` → 9 로 확인 후 ff 로 교정).

---

## §E.2 Run-phase Evidence

_<pending run-phase>_

---

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
