# SPEC-CODEX-PARTIAL-WIRING-001 — 수용 기준

> 카드 t499 · 판정 단위는 AC-CPW-001 ~ AC-CPW-009.
>
> **측정 기준 트리(SHA pin)**: `ace1c5440fad4d2a3787b0331059d4f3c4149d15` (worktree
> `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t499`, branch `WT-codex-partial-wiring`).
> 모든 RED-now 관측은 **이 트리에서, 구현 착수 전에** 실제로 실행해 얻은 것이다.
> `git status --short -- internal/` 무출력 — 관측 시점 소스 미변경.
>
> **두 셀 규율**: 각 릴리스 게이트 기준은 (1) 오늘 관측된 RED-now와 (2) 그것을 초록으로 뒤집는 경로를
> 함께 싣는다. 한 셀만 있는 기준은 채택되지 않은 것으로 본다.
>
> **종료코드를 읽은 방법(주석)**: 아래 인용은 명령 / stdout / `exit:` 세 필드로 분해해 싣는다. 실제로
> 값을 읽을 때는 `<cmd> > out.txt 2>&1; echo $?` 형태를 썼다 — zsh가 `PIPESTATUS`를 채우지 않아 파이프 뒤
> `$?`가 빈 값을 내고, 그 빈 값이 또 하나의 공허한 초록이 되기 때문이다. 리다이렉션은 **값을 읽는 방법**이지
> 기준이 내거는 명령이 아니므로, 인용되는 명령은 단일 호출 형태로 적는다.
>
> **stdout 인용 범위**: 인용은 판정에 실린 부분(`ok … [no tests to run]`)만 유효하다. `(cached)`와
> `0.939s` 같은 캐시·시간 토큰은 실행마다 달라지므로 바이트 비교 대상이 아니다.

## §D.0 [HARD] RED-now 판독표 — 두 관측의 네 조합

각 RED-now 셀은 **두 관측을 짝지어** 싣는다. 하나는 "러너가 아무것도 고르지 못했다"(`[no tests to run]`),
다른 하나는 "고를 대상이 소스에 없다"(`grep -c 'func <시험 이름>'` → `0`). 하나만으로는 **부재의 원인**이
고정되지 않는다 — `[no tests to run]`은 "아직 안 만들었다"와 "누군가 지웠다"를 구별하지 못하고, 그 셀은
미래에 시험이 삭제돼도 계속 정상으로 읽힌다.

| grep 개수 | 러너 출력 | 읽는 법 |
|---|---|---|
| `0` | `[no tests to run]` | **정당한 RED-now** — 고를 대상이 소스에 없고 러너도 아무것도 고르지 못했다. 원인 고정됨 |
| `≥1` | `--- PASS` + `[no tests to run]` 부재 | **GREEN** — 시험이 있고 실제로 돌았다 |
| `≥1` | `[no tests to run]` | **RED 아님 — 측정 결함(즉시 blocker)**. 시험은 소스에 있는데 셀렉터가 잡지 못한 것: 이름 오타 / 빌드 태그 / 다른 패키지 배치. 이 상태를 RED로 뭉개면 "고쳐도 계속 RED"인 셀이 생겨 같은 종류의 구멍을 하나 더 만든다. **셀렉터를 고치기 전에는 그 AC를 진행시키지 않는다** |
| `0` | `--- PASS` | **grep 범위가 틀렸다(즉시 blocker)** — 위 행의 거울상. 동명 시험이 다른 파일에 있다는 뜻이므로, grep 대상 파일을 바로잡기 전에는 어느 쪽 관측도 근거로 쓰지 않는다 |

아래 두 행은 RED도 GREEN도 아닌 **제3의 상태**다. 두 관측이 어긋난다는 것은 대상이 아니라 **계측 장치**에
대한 정보이고, 그때 나온 값은 어느 쪽으로도 판정 근거가 되지 못한다.

> **`grep -c` 주의**: 0건일 때 종료코드 `1`을 낸다. 판별식은 종료코드가 아니라 **출력된 개수**다.

## §D. AC 매트릭스

| AC | 대응 REQ | 성격 | 판정 명령 |
|---|---|---|---|
| AC-CPW-001 | REQ-CPW-001, REQ-CPW-002, REQ-CPW-004 | **릴리스 게이트**(RED-now 있음) | `go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredCodexInstalled -v` |
| AC-CPW-002 | REQ-CPW-001, REQ-CPW-003, REQ-CPW-004, REQ-CPW-009, REQ-CPW-011 | **릴리스 게이트**(RED-now 있음) | `go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredCodexAbsent -v` |
| AC-CPW-003 | REQ-CPW-006 | 회귀 가드(게이트 아님) | §D.2 AC-CPW-003의 명령(표 셀 안에서는 교대(`\|`)가 렌더에서 깨지므로 본문에 둔다) |
| AC-CPW-004 | REQ-CPW-006 | 회귀 가드(게이트 아님) | `go test ./internal/cli/ -run TestDoctorGolden -v` + `git diff --stat internal/cli/testdata/` |
| AC-CPW-005 | REQ-CPW-005 | **릴리스 게이트**(RED-now 있음) | `go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredCountIndependent -v` |
| AC-CPW-006 | REQ-CPW-007 | **릴리스 게이트**(RED-now 있음) | `go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredReadOnly -v` |
| AC-CPW-007 | REQ-CPW-008 | **릴리스 게이트**(RED-now 있음) | §D.4 AC-CPW-007의 명령(같은 이유로 본문에 둔다) |
| AC-CPW-008 | REQ-CPW-010 | 회귀 가드(게이트 아님) | `git diff --stat internal/codexwiring/` + `go test ./internal/codexwiring/` |
| AC-CPW-009 | 전체(반증 관문) | **릴리스 게이트**(뮤턴트 관측이 곧 RED 셀) | §D.3의 뮤턴트 8종 |

> **회귀 가드 3건(AC-CPW-003 / -004 / -008)의 위상**: 이 셋은 구현 전에도 초록이고 구현 후에도 초록이어야
> 하는 보존 기준이라 원리상 RED-now 셀을 가질 수 없다. 따라서 **릴리스 게이트가 아니라 회귀 가드**로
> 분류한다 — 깨지면 즉시 blocking이지만, 이 셋이 초록이라는 사실은 어떤 진전도 증명하지 않는다.
>
> 릴리스 게이트는 **6건**이다: AC-CPW-001 / -002 / -005 / -006 / -007(시험 5종) + AC-CPW-009(뮤턴트 관문).
> §D.4 체크리스트가 앞의 다섯을 한 항목으로, AC-CPW-009를 별도 항목으로 세는 것은 같은 6건이다.

## §D.1 릴리스 게이트 (두 셀 전부 기재)

### AC-CPW-001 — 반쪽 배선 + codex 존재 → 실행 가능한 경고

- **Given** `.codex/agents/moai/` 아래 agent 정의 TOML이 하나 이상 있고 `.codex/hooks.json`과 `.codex/config.toml`이 둘 다 없는 프로젝트, `codexWiringLookPath`가 `codex`를 해석하도록 고정, `stubCodexHome`으로 사용자 계층 home 고정
- **When** `checkCodexWiring(root, false)`를 호출하면
- **Then** (a) `check.Status == uikit.CheckWarn`, (b) `check.Message`가 프로젝트에 Codex agent 정의가 있다는 사실과 배선 파일이 없다는 사실을 함께 진술, (c) `check.Message`가 조치 지시문을 포함, (d) `check.Detail`이 부재 경로 `.codex/hooks.json`과 `.codex/config.toml`을 이름으로 인용, (e) `check.Message + check.Detail`에 `claude-only`가 **0회**(REQ-CPW-004는 "어떤 분기에서도"이므로 이 갈래도 측정한다)
- **RED-now** (트리 `ace1c5440`) — 두 관측(§D.0):
  - 관측 1 — 명령: `go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredCodexInstalled`
    - stdout: `ok  	github.com/modu-ai/moai-adk/internal/cli	[no tests to run]`
    - exit: `0`
  - 관측 2 — 명령: `grep -c 'func TestCheckCodexWiring_HalfWiredCodexInstalled' internal/cli/doctor_codex_test.go`
    - stdout: `0`
    - exit: `1` (0건일 때의 `grep -c` 관례 — 판별식은 개수 `0`이다)
  - 판독: (grep `0` · `[no tests to run]`) = **정당한 RED-now**. 지금 이 기준은 판정 불능이다
- **GREEN 경로**: `plan.md` §E M2가 이 시험을 만든다. 초록의 모양은 **세 가지가 동시에** — `--- PASS: TestCheckCodexWiring_HalfWiredCodexInstalled`, `[no tests to run]` **부재**, 그리고 같은 grep이 `1` 이상

### AC-CPW-002 — 반쪽 배선 + codex 부재 → 잔소리 없이 사실만

- **Given** AC-CPW-001과 같은 프로젝트, `codexWiringLookPath`는 `codex`에 대해 오류를 돌려주도록 고정. 사용자 계층 home은 **두 하위 케이스로 나눠 각각 고정**한다:
  - **(i) 깨끗한 home** — `stubCodexHome(t, t.TempDir())`. `~/.codex/config.toml`이 없으므로 `codexStaleSkillFinding`은 fail-open으로 소견 0건
  - **(ii) 낡은 home** — `stubCodexHome(t, writeCodexHomeConfig(t, …))`로 존재하지 않는 경로를 가리키는 `[[skills.config]]` 항목을 하나 이상 선언(감사 D2)
  > **왜 둘 다 재는가 — 그리고 왜 (ii)를 버리면 안 되는가.** 올바른 구현에서는 이 갈래가 사용자 계층 훑기에 도달하지 않으므로(REQ-CPW-011) `problems`가 비고, **home 내용과 무관하게** 같은 Message가 나온다. 즉 등가 단언은 두 하위 케이스에서 **모두** 성립해야 한다. (ii)를 빼고 (i)만 재면 훑기로 흘려보내는 구현(§D.3 M7)이 깨끗한 CI에서 초록으로 통과하고 실사용 머신에서만 뒤집힌다 — 감사 D2가 지적한 바로 그 형태다.
  >
  > **진단 가능성은 픽스처를 약화시켜서가 아니라 하위 케이스를 갈라서 얻는다**: (i)만 실패 → 리터럴 자체가 달라졌다(M6 / M6′ 계열). (ii)만 실패 → 사용자 계층 훑기가 새어 들어왔다(M7). 둘 다 실패 → 구현이 이 갈래를 아예 다르게 처리한다(M1 계열)
  >
  > 헬퍼 확인(읽음): `stubCodexHome`(`internal/cli/doctor_codex_test.go:76`)은 home 해석기를 주어진 경로로 돌릴 뿐이라 `t.TempDir()`를 주면 깨끗한 home이 되고, 낡은 항목은 `writeCodexHomeConfig`(`:95`) + `absentSkillPath`(`:132`)가 만든다. 두 픽스처가 분리돼 있으므로 M7이 등가 시험을 통째로 오염시키지 않는다
- **When** `checkCodexWiring(root, false)`를 호출하면
- **Then**
  - (a) `check.Status == uikit.CheckOK` — **두 하위 케이스 모두**, 즉 낡은 항목이 선언된 home 아래에서도 그렇다(REQ-CPW-011). 근거(읽음): `internal/cli/doctor_codex.go:193-201` — `problems`가 하나라도 차면 Status가 `CheckWarn`이 되고 Message는 `joinCodexSummaries(problems)`로 `"; "` 이어붙은 문자열이 된다. 따라서 (a)와 (d)는 같은 사실의 두 얼굴이다
  - (b) `check.Message`가 Codex agent 정의의 존재를 진술
  - (c) `check.Message + check.Detail`에 `claude-only`가 **0회**
  - (d) **`check.Message`가 시험 파일에 적힌 리터럴 한 개와 정확히 같다 — 두 하위 케이스 모두에서**(등가 단언, allowlist):
    ```go
    const wantHalfWiredAbsentMessage = "<run-phase에서 확정한 문구>"   // 시험 파일 쪽 리터럴
    if check.Message != wantHalfWiredAbsentMessage { t.Errorf(...) }
    ```
    금지 목록(denylist)이 아니라 **허용 목록**이다. 어떤 패러프레이즈를 새로 발명하든 리터럴과 다르므로 **원리상 하나도 통과하지 못한다**(REQ-CPW-009). 이 파일의 기존 관용구와도 맞는다 — `internal/cli/doctor_codex.go`는 프로젝트별 가변부가 없는 갈래의 Message를 이미 평문 리터럴로 못박고(`:101` unwired-skip, `:195` wired-OK) 경로 같은 가변부는 Detail로 내린다. 조치 문구를 이름 붙인 상수로 두는 관례도 이미 있다(`reTrustAdvice` `:46`, `initCodexAdvice` `:52`)
  - (d′) **그 리터럴 자체가 부분문자열 `moai init` / `run ` / `install`을 담지 않는다**
    > **(d)와 (d′)는 약한 검사와 강한 검사의 중복이 아니라 — 서로 다른 대상을 잰다.** (d)는 *런타임 Message가 리터럴에서 벗어났는가*를 재고, (d′)는 *그 리터럴로 무엇을 골랐는가*를 잰다. (d)가 아무리 강해도 (d′)의 대상에는 닿지 못한다: 저자가 지시문을 리터럴로 고르고 시험 리터럴도 같게 적으면 (d)는 초록이다.
    >
    > **남는 잔여분.** (d′)의 세 토큰 열거는 알려진 형태만 잡으므로, 그것을 모두 피한 새 지시문을 리터럴로 고르는 경우는 여전히 열려 있다. 그 선택은 run-phase 저자의 한 번의 판단이며, 이제 **한 곳에, 한 줄로, diff에 보이게** 모인다(구현 상수 + 시험 리터럴이 함께 바뀌어야 초록). 검토 지점이 반증하기 어려운 체크리스트 항목에서 diff에 보이는 코드 한 줄로 옮겨간 것이 이 전환의 실질이다 — 구멍이 사라진 것이 아니라
  - (e) `check.Message + check.Detail`에 `[[skills.config]]` 유래 소견이 등장하지 않는다 — 사용자 계층 훑기가 이 갈래에 도달하지 않았다는 관측(REQ-CPW-011)
- **RED-now** (트리 `ace1c5440`) — 두 관측(§D.0):
  - 관측 1 — 명령: `go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredCodexAbsent`
    - stdout: `ok  	github.com/modu-ai/moai-adk/internal/cli	[no tests to run]`
    - exit: `0`
  - 관측 2 — 명령: `grep -c 'func TestCheckCodexWiring_HalfWiredCodexAbsent' internal/cli/doctor_codex_test.go`
    - stdout: `0`
    - exit: `1`
  - 판독: **정당한 RED-now**
- **GREEN 경로**: M2가 이 시험을 만든다. `--- PASS` + `[no tests to run]` 부재 + grep `≥1`

### AC-CPW-005 — 개수 비의존

- **Given** agent 정의 TOML이 **1개**인 프로젝트, **12개**인 프로젝트, 그리고 디렉터리는 있으나 정의 파일이 **0개**인 프로젝트. 세 경우 모두 `stubCodexHome` 고정
- **When** 각각에 대해 `checkCodexWiring`을 호출하면
- **Then** 1개·12개 프로젝트는 half-wired로 분류되고, 0개(빈 디렉터리) 프로젝트는 half-wired가 **아니다**(기존 `unwired` 경로 그대로). 더해 `grep -nE '\b11\b' internal/cli/doctor_codex.go`가 판별식 안에서 11을 쓰는 줄을 하나도 내지 않는다
- **RED-now** (트리 `ace1c5440`) — 두 관측:
  - 관측 1 — 명령: `go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredCountIndependent`
    - stdout: `ok  	github.com/modu-ai/moai-adk/internal/cli	[no tests to run]`
    - exit: `0`
  - 관측 2 — 명령: `grep -c 'func TestCheckCodexWiring_HalfWiredCountIndependent' internal/cli/doctor_codex_test.go`
    - stdout: `0` · exit: `1`
  - 판독: **정당한 RED-now**
- **GREEN 경로**: M2가 3케이스(1 / 12 / 빈 디렉터리)를 가진 시험을 만든다. **빈 디렉터리 케이스가 실제로 존재해야** §D.3 M2 뮤턴트가 RED가 된다 — 빠지면 그 뮤턴트는 통과한다

### AC-CPW-006 — 읽기 전용·비차단

- **Given** 반쪽 배선 프로젝트의 파일 목록 스냅샷(`find <root> -type f | sort`)과 `stubCodexHome`으로 고정한 사용자 계층 config의 내용 해시
- **When** `checkCodexWiring`을 두 갈래(codex 존재/부재) 모두에서 호출한 뒤 다시 스냅샷과 해시를 뜨면
- **Then** 파일 목록과 해시가 호출 전과 **동일**하다(생성·수정·삭제 0건)
- **RED-now** (트리 `ace1c5440`) — 두 관측:
  - 관측 1 — 명령: `go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredReadOnly`
    - stdout: `ok  	github.com/modu-ai/moai-adk/internal/cli	[no tests to run]`
    - exit: `0`
  - 관측 2 — 명령: `grep -c 'func TestCheckCodexWiring_HalfWiredReadOnly' internal/cli/doctor_codex_test.go`
    - stdout: `0` · exit: `1`
  - 판독: **정당한 RED-now**
- **GREEN 경로**: M2가 이 시험을 만든다. 더해 실제 반쪽 프로젝트에서 `moai doctor`의 종료코드가 `0`이다(종료코드는 `CheckFail` 수만 세고 `CheckWarn`에는 둔감하다 — `internal/cli/doctor.go`의 `countFailedChecks` / `doctorExitStatus`. REQ-CPW-007의 종료코드 절에 대해 그 민감도가 옳다)

### AC-CPW-007 — 폭 상한 (감사 iter1 D1 수리: 반쪽 경로를 실제로 밟는 시험)

- **Given** 반쪽 배선 프로젝트를, **두 PATH 갈래 모두**에 대해(양쪽 다 `stubCodexHome` 고정)
- **When** `checkCodexWiring`을 호출하면
- **Then** 두 갈래 각각에서 `utf8.RuneCountInString(check.Message) <= codexMessageWidthCeiling`(113)이며, `--verbose` 렌더에서도 패널 폭이 밴드를 벗어나지 않는다
  > **등가 단언에 흡수시키지 않고 별도로 두는 이유(리드 질의 2).** 부재 갈래에서는 폭이 리터럴 하나의 성질이 되므로 이 기준이 **덧붙는 값**은 작다(등가가 이미 그 문자열을 고정한다). 그러나 **codex 존재 갈래는 흡수할 수 없다** — 그 갈래는 `joinCodexSummaries`로 여러 소견이 `"; "`로 이어붙어 Message가 만들어지므로(`internal/cli/doctor_codex.go:193-201`), 반쪽 소견이 사용자 계층 소견과 합쳐지는 순간 길이가 리터럴의 성질이 아니게 된다. 폭 상한이 실제로 하중을 받는 곳이 바로 거기다. 따라서 AC-CPW-007은 **두 갈래를 함께 재는 하나의 기준으로 유지**하고, 부재 갈래분은 값싼 중복으로 남긴다
- **판정 명령**(신규 1 = 게이트, 기존 2 = 회귀 가드):
  - 신규(게이트): `go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredMessageWidthStaysInBand -v`
  - 기존(가드): `go test ./internal/cli/ -run 'TestCheckCodexWiring_(MessageWidthStaysInBand|RenderedPanelStaysInBand)' -v`
- **RED-now** (트리 `ace1c5440`) — 두 관측:
  - 관측 1 — 명령: `go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredMessageWidthStaysInBand`
    - stdout: `ok  	github.com/modu-ai/moai-adk/internal/cli	[no tests to run]`
    - exit: `0`
  - 관측 2 — 명령: `grep -c 'func TestCheckCodexWiring_HalfWiredMessageWidthStaysInBand' internal/cli/doctor_codex_test.go`
    - stdout: `0` · exit: `1`
  - 판독: **정당한 RED-now**
- **왜 기존 두 시험만으로는 안 되는가**(실측): 두 시험 모두 `checkCodexWiring(t.TempDir(), …)`를 호출한다 — `internal/cli/doctor_codex_test.go:404`(`TestCheckCodexWiring_MessageWidthStaysInBand`, 함수 선언 `:390`)와 `:450`(`TestCheckCodexWiring_RenderedPanelStaysInBand`, 함수 선언 `:437`). 빈 임시 디렉터리, 즉 REQ-CPW-006이 얼려 놓은 `unwired` 경로다. **반쪽 경로에 결코 도달하지 못하므로 새 문구의 폭을 관측할 수 없다.** 반쪽 갈래에 200 rune Message를 주는 뮤턴트가 이 둘을 통과한다
  > 0.3.0 정정(감사 iter2 D4): 0.2.0은 이 좌표를 `:394` / `:441`로 적었다. 논거는 실측으로 참이고 **좌표만** 틀렸다 — 실측 `grep -n 'checkCodexWiring(t.TempDir()' internal/cli/doctor_codex_test.go` → `155: 404: 450: 552: 567:`(트리 `ace1c5440`). §A.3의 D6 정정과 같은 계열의 재발이라 원문 오기를 지우지 않고 남긴다
- **GREEN 경로**: M2가 신규 시험을 만든다. `--- PASS` + `[no tests to run]` 부재 + grep `≥1`

## §D.2 회귀 가드

### AC-CPW-003 — 기존 세 시험이 수정 없이 통과

- **Given** run-phase 종료 시점의 트리
- **When** `git diff -- internal/cli/doctor_codex_test.go`에서 기존 세 시험 본문(`TestCheckCodexWiring_ClaudeOnlyMachineStaysSilent`, `TestCheckCodexWiring_InactiveProjectInformationalSkip`, `TestCheckCodexWiring_UnwiredWithCodexInstalledWarns`)이 **변경되지 않았음**을 확인하고 세 시험을 실행하면
- **Then** 세 시험 모두 `--- PASS`이며, 이 셋의 단언 문구는 어느 것도 완화되지 않았다
- **판정**: `go test ./internal/cli/ -run 'TestCheckCodexWiring_(ClaudeOnlyMachineStaysSilent|InactiveProjectInformationalSkip|UnwiredWithCodexInstalledWarns)' -v` → 세 줄 모두 `--- PASS`. 시험 본문이 수정되었다면 이 AC는 FAIL이며, 수정 사유를 `progress.md` §E.2에 적고 리드 판단을 받는다
- **RED-now 없음**: 구현 전에도 초록이다(보존 기준의 성질). 그래서 게이트가 아니라 가드다

### AC-CPW-004 — 골든 3본이 재생성 없이 통과

- **Given** `internal/cli/testdata/doctor-{light,dark,nocolor}.golden`
- **When** `UPDATE_GOLDEN` 없이 골든 시험을 실행하면
- **Then** 세 시험 모두 `--- PASS`이고, `git diff --stat internal/cli/testdata/`가 **빈 출력**이다(골든 파일 0바이트 변경)
- **판정**: `go test ./internal/cli/ -run TestDoctorGolden -v`(exit `0`, `--- PASS` ×3) 후 `git diff --stat internal/cli/testdata/`(빈 출력)
- **RED-now 없음**: 위와 같은 이유

### AC-CPW-008 — 존재-게이트 무접촉

- **Given** run-phase 종료 시점의 트리
- **When** `git diff -U0 internal/codexwiring/wire.go`를 읽으면
- **Then** `RefreshWiring` / `wireProject` / `wiringFilesExist`의 **동작 변경이 없다**(REQ-CPW-010). 배선 상태 판별을 위한 **상수 추가**는 허용되며, 그 경우 추가된 줄을 인용해 세 함수의 동작과 무관함을 보인다
- **판정**: `git diff -U0 internal/codexwiring/wire.go`의 변경 줄을 눈으로 확인 + `go test ./internal/codexwiring/`(exit `0`)
- **매핑 정정 이력**: 0.1.0은 이 기준을 REQ-CPW-007로 매핑했으나 그 요구의 주체는 doctor 검사이지 `codexwiring`이 아니다(감사 iter1 D5). 0.2.0에서 REQ-CPW-010을 신설해 옮겼다
- **RED-now 없음**: 보존 기준

## §D.3 반증 관문 — 뮤턴트 8종

### AC-CPW-009 — 여덟 뮤턴트가 모두 RED를 만든다

각 뮤턴트는 **심고 → 해당 시험이 FAIL 하는 것을 관측하고 → 원복**한다. 원복 확인은 `git diff --stat internal/cli/` 빈 출력.
여기서는 뮤턴트 관측 자체가 RED 셀이다.

| # | 심는 변형 | RED가 되어야 하는 시험 | 지키는 요구 |
|---|---|---|---|
| M1 | 반쪽 판별을 다시 배선 파일 2종만 보는 한 줄로 되돌린다(agent 조회 제거) | `..._HalfWiredCodexAbsent`, `..._HalfWiredCodexInstalled` | REQ-CPW-001 |
| M2 | agent 정의 존재 판별을 "디렉터리가 존재하는가"로 바꾼다(파일 유무를 보지 않음) | `..._HalfWiredCountIndependent`의 빈 디렉터리 케이스 | REQ-CPW-005 |
| M3 | codex 부재 갈래 Message를 기존 `not wired (claude-only project) — skipped`로 되돌린다 | `..._HalfWiredCodexAbsent`(`claude-only` 0회) | REQ-CPW-004, REQ-CPW-006 |
| M4 | codex **부재** 갈래 Message 끝에 `— run moai init --agent codex`를 붙인다 | `..._HalfWiredCodexAbsent`(`moai init` 부재) | REQ-CPW-009 |
| M5 | 반쪽 갈래 Message를 200 rune로 늘린다 | `..._HalfWiredMessageWidthStaysInBand` | REQ-CPW-008 |
| **M6** | codex **부재** 갈래 Message를 감사가 실증한 **패러프레이즈 지시문**으로 심는다 — 정확히: `` codex agent definitions present, no wiring files — `moai init --agent codex` wires them `` | `..._HalfWiredCodexAbsent`(Then (d) 등가 단언 — 리터럴과 다르므로 즉시 FAIL) | REQ-CPW-009 |
| **M6′** | 세 금지 토큰을 **모두 피한** 지시문을 심는다 — 예: `codex agent definitions present, no wiring files — wire them once codex is available` | `..._HalfWiredCodexAbsent` 두 하위 케이스 모두(Then (d) 등가 단언 — 리터럴과 다르므로 즉시 FAIL) | REQ-CPW-009 |
| **M7** | codex 부재 갈래를 조기 반환시키지 않고 본 경로로 흘려보내 `codexStaleSkillFinding`(`internal/cli/doctor_codex.go:189` 호출, 함수 `:376`)이 돌게 한다 | `..._HalfWiredCodexAbsent`의 **하위 케이스 (ii)만** — Then (a) `CheckOK` · (d) 등가 · (e) 소견 부재가 함께 깨진다. 하위 케이스 (i)는 초록으로 남는데, **그 비대칭이 곧 진단**이다(깨끗한 home에서는 fail-open이라 증상이 없다 = 이 결함이 CI에서 안 보이는 이유 그 자체) | REQ-CPW-003, REQ-CPW-011 |

> **뮤턴트 이름 대조**(혼선 방지): 감사 iter2 판정서는 패러프레이즈 뮤턴트를 `M4′`, 사용자 계층 훑기 뮤턴트를
> `M6`으로 불렀다. 이 문서는 번호를 이어 붙여 각각 **M6 / M7**로 쓴다. 대상은 동일하다.
>
> **[HARD] M6·M6′·M7이 RED가 되지 않으면 이번 수리는 이름만 들어간 것이다.** M6은 iter2 D1(토큰 단언의
> 패러프레이즈 우회)을, **M6′는 금지 토큰을 전부 피한 새 형태**(등가 단언으로 전환한 이유 그 자체)를,
> M7은 iter2 D2(CI에서만 초록인 구현)를 각각 잡으라고 있는 것이므로, 셋 중 하나라도 초록이면
> 어떤 AC도 PASS로 적지 않는다. M1이 RED가 되지 않는 경우도 마찬가지다 — 그때는 신규 시험이 판별식을
> 잡고 있지 않다는 뜻이다.
>
> **M7의 특이점**: 이 뮤턴트는 `stubCodexHome`이 **낡은 항목을 선언한** home으로 고정돼 있어야만 RED가 된다.
> 깨끗한 home으로 고정하면 `codexStaleSkillFinding`이 fail-open으로 소견 0건을 내 초록이 되고, 그 초록은
> 실사용 머신에서 뒤집히는 구현을 그대로 통과시킨다. AC-CPW-002 Given (ii)가 이 조건이다.

## §D.4 Definition of Done

- [ ] 릴리스 게이트 시험 5건(AC-CPW-001 / -002 / -005 / -006 / -007): 판정 명령 실행 후 **세 가지 동시 확인** — `--- PASS`, `[no tests to run]` 부재, 그리고 §D.0 관측 2의 grep이 `1` 이상. 출력 그대로 `progress.md` §E.2에 인용
- [ ] 회귀 가드 3건(AC-CPW-003 / -004 / -008): 판정 명령 실행 + 출력 인용
- [ ] AC-CPW-009 뮤턴트 **8종** 각각 RED 관측 → 원복 → `git diff --stat internal/cli/` 빈 출력 확인. **M6·M6′·M7의 RED는 이번 라운드 수리의 합격 조건이므로 생략 불가**
- [ ] `go test ./internal/cli/... ./internal/codexwiring/...` 통과 (전체 스위트는 CI 몫 — CLAUDE.local.md §4)
- [ ] `go vet ./internal/cli/... ./internal/codexwiring/...` 통과
- [ ] `gofmt -l internal/cli internal/codexwiring` 빈 출력
- [ ] 실제 반쪽 프로젝트에서 `moai doctor` 두 갈래를 눈으로 확인하고 출력을 증거로 남김 (절차는 `.moai/reports/t499/repro.md`와 동일)
- [ ] **부재 갈래 Message가 시험 파일의 리터럴과 등가로 고정**되어 있고(AC-CPW-002 (d)), 그 리터럴이 (d′)의 모양 검사를 통과한다. 문구를 다듬으면 시험이 함께 깨지는 것이 정상이며 — 사용자에게 보이는 문구가 바뀌는데 시험이 안 깨지는 쪽이 나쁘다 — 그때 구현 상수와 시험 리터럴을 **함께** 고친 diff를 증거로 남긴다
- [ ] 범위 밖(`spec.md` §E) 항목을 건드리지 않았음을 `git diff --name-only`로 확인
- [ ] **인용한 모든 `file:line` 좌표를 커밋 직전에 재측정**했음을 보고에 명시(이 카드에서 좌표 오기가 두 번 났다 — §A.3 D6, AC-CPW-007 D4)
- [ ] 종료코드는 `<cmd> > out.txt 2>&1; echo $?`로 읽는다(파이프 뒤 `$?` / `PIPESTATUS` 금지). 인용은 명령 / stdout / `exit:` 세 필드로 분해
