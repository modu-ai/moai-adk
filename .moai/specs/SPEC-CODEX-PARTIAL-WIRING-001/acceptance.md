# SPEC-CODEX-PARTIAL-WIRING-001 — 수용 기준

> 카드 t499 · 판정 단위는 AC-CPW-001 ~ AC-CPW-009.
>
> **측정 기준 트리(SHA pin)**: `ace1c5440fad4d2a3787b0331059d4f3c4149d15` (worktree
> `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t499`, branch `WT-codex-partial-wiring`).
> 아래 모든 RED-now 관측은 **이 트리에서, 구현 착수 전에** 실제로 실행해 얻은 것이다.
> `git status --short -- internal/` 무출력 — 관측 시점에 소스는 미변경 상태였다.
>
> **두 셀 규율**: 각 릴리스 게이트 기준은 (1) 오늘 관측된 RED-now와 (2) 그것을 초록으로 뒤집는 경로를
> 함께 싣는다. 한 셀만 있는 기준은 채택되지 않은 것으로 본다.
>
> **[HARD] 종료코드 읽는 법**: 파이프 뒤 `$?` / `PIPESTATUS`로 읽지 않는다(zsh가 안 채워 빈 값이 나오고,
> 그 빈 값이 또 하나의 공허한 초록이 된다). `<cmd> > out.txt 2>&1; echo $?` 형태로 읽는다.

## §D. AC 매트릭스

| AC | 대응 REQ | 성격 | 판정 명령 |
|---|---|---|---|
| AC-CPW-001 | REQ-CPW-001, REQ-CPW-002, REQ-CPW-004 | **릴리스 게이트**(RED-now 있음) | `go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredCodexInstalled -v` |
| AC-CPW-002 | REQ-CPW-001, REQ-CPW-003, REQ-CPW-004, REQ-CPW-009 | **릴리스 게이트**(RED-now 있음) | `go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredCodexAbsent -v` |
| AC-CPW-003 | REQ-CPW-006 | 회귀 가드(게이트 아님) | §D.2 AC-CPW-003의 명령(표 셀 안에서는 교대(`\|`)가 렌더에서 깨지므로 본문에 둔다) |
| AC-CPW-004 | REQ-CPW-006 | 회귀 가드(게이트 아님) | `go test ./internal/cli/ -run TestDoctorGolden -v` + `git diff --stat internal/cli/testdata/` |
| AC-CPW-005 | REQ-CPW-005 | **릴리스 게이트**(RED-now 있음) | `go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredCountIndependent -v` |
| AC-CPW-006 | REQ-CPW-007 | **릴리스 게이트**(RED-now 있음) | `go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredReadOnly -v` |
| AC-CPW-007 | REQ-CPW-008 | **릴리스 게이트**(RED-now 있음) | §D.4 AC-CPW-007의 명령(같은 이유로 본문에 둔다) |
| AC-CPW-008 | REQ-CPW-010 | 회귀 가드(게이트 아님) | `git diff --stat internal/codexwiring/` + `go test ./internal/codexwiring/` |
| AC-CPW-009 | 전체(반증 관문) | **릴리스 게이트**(뮤턴트 관측이 곧 RED 셀) | §D.3의 뮤턴트 5종 |

> **회귀 가드 3건(AC-CPW-003 / -004 / -008)의 위상**: 이 셋은 구현 전에도 초록이고 구현 후에도 초록이어야
> 하는 보존 기준이라 원리상 RED-now 셀을 가질 수 없다. 따라서 **릴리스 게이트가 아니라 회귀 가드**로
> 분류한다 — 깨지면 즉시 blocking이지만, 이 셋이 초록이라는 사실은 어떤 진전도 증명하지 않는다.

## §D.1 릴리스 게이트 (두 셀 전부 기재)

### AC-CPW-001 — 반쪽 배선 + codex 존재 → 실행 가능한 경고

- **Given** `.codex/agents/moai/` 아래 agent 정의 TOML이 하나 이상 있고 `.codex/hooks.json`과 `.codex/config.toml`이 둘 다 없는 프로젝트, 그리고 `codexWiringLookPath`가 `codex`를 해석하도록 고정된 상태
- **When** `checkCodexWiring(root, false)`를 호출하면
- **Then** (a) `check.Status == uikit.CheckWarn`, (b) `check.Message`가 프로젝트에 Codex agent 정의가 있다는 사실과 배선 파일이 없다는 사실을 함께 진술, (c) `check.Message`가 문자열 `run moai init --agent codex`를 포함, (d) `check.Detail`이 부재 경로 `.codex/hooks.json`과 `.codex/config.toml`을 이름으로 인용, **(e) `check.Message + check.Detail`에 `claude-only`가 0회 등장**(D4 — REQ-CPW-004는 "어떤 분기에서도"이므로 이 갈래도 측정한다)
- **RED-now** (트리 `ace1c5440`, 구현 착수 전 실측):
  ```
  $ go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredCodexInstalled > /tmp/t499-red1.txt 2>&1; echo $?
  0
  $ cat /tmp/t499-red1.txt
  ok  	github.com/modu-ai/moai-adk/internal/cli	(cached) [no tests to run]
  ```
  **이 `ok`는 통과가 아니다.** 지정한 테스트가 아직 없어 셀렉터가 0개를 잡았고, 그때 러너는 종료코드 0과 `ok`를 낸다 — `[no tests to run]` 토큰이 그 증거다. 즉 이 기준은 **지금 판정 불능**이며, 그 상태를 RED로 읽는다.
- **GREEN 경로**: `plan.md` §E M2가 `TestCheckCodexWiring_HalfWiredCodexInstalled`를 만든다. 초록의 모양은 `[no tests to run]`이 사라지고 `--- PASS: TestCheckCodexWiring_HalfWiredCodexInstalled`가 나타나는 것이다. **판정 시 `[no tests to run]` 부재를 함께 확인한다** — 확인하지 않으면 같은 공허한 초록을 두 번 읽게 된다.

### AC-CPW-002 — 반쪽 배선 + codex 부재 → 잔소리 없이 사실만

- **Given** AC-CPW-001과 같은 프로젝트, 단 `codexWiringLookPath`가 `codex`에 대해 오류를 돌려주도록 고정된 상태
- **When** `checkCodexWiring(root, false)`를 호출하면
- **Then** (a) `check.Status == uikit.CheckOK`, (b) `check.Message`가 Codex agent 정의의 존재를 진술, (c) `check.Message + check.Detail`에 `claude-only`가 **0회**, **(d) `check.Message`가 `initCodexAdvice`(= `run moai init --agent codex`)를 포함하지 않는다**(D3 — REQ-CPW-009. 이 단언이 없으면 "실행 불가능한 지시를 모든 프로젝트에 거는" 구현이 아홉 기준을 전부 만족한다)
- **RED-now** (트리 `ace1c5440`):
  ```
  $ go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredCodexAbsent > /tmp/t499-red2.txt 2>&1; echo $?
  0
  $ cat /tmp/t499-red2.txt
  ok  	github.com/modu-ai/moai-adk/internal/cli	0.939s [no tests to run]
  ```
  AC-CPW-001과 같은 이유로 판정 불능 = RED.
- **GREEN 경로**: M2가 `TestCheckCodexWiring_HalfWiredCodexAbsent`를 만든다. `--- PASS` + `[no tests to run]` 부재.

### AC-CPW-005 — 개수 비의존

- **Given** agent 정의 TOML이 **1개**인 프로젝트, **12개**인 프로젝트, 그리고 디렉터리는 있으나 정의 파일이 **0개**인 프로젝트
- **When** 각각에 대해 `checkCodexWiring`을 호출하면
- **Then** 1개·12개 프로젝트는 half-wired로 분류되고, 0개(빈 디렉터리) 프로젝트는 half-wired가 **아니다**(기존 `unwired` 경로 그대로). 더해 `grep -nE '\b11\b' internal/cli/doctor_codex.go`가 판별식 안에서 11을 쓰는 줄을 하나도 내지 않는다
- **RED-now** (트리 `ace1c5440`):
  ```
  $ go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredCountIndependent > /tmp/t499-red3.txt 2>&1; echo $?
  0
  $ cat /tmp/t499-red3.txt
  ok  	github.com/modu-ai/moai-adk/internal/cli	0.686s [no tests to run]
  ```
- **GREEN 경로**: M2가 3케이스(1 / 12 / 빈 디렉터리)를 가진 `TestCheckCodexWiring_HalfWiredCountIndependent`를 만든다. **빈 디렉터리 케이스가 실제로 존재해야** AC-CPW-009의 M2 뮤턴트가 RED가 된다 — 이 케이스가 빠지면 그 뮤턴트는 통과한다.

### AC-CPW-006 — 읽기 전용·비차단

- **Given** 반쪽 배선 프로젝트의 파일 목록 스냅샷(`find <root> -type f | sort`)과 `~/.codex/config.toml` 대역(`stubCodexHome`)의 내용 해시
- **When** `checkCodexWiring`을 두 갈래(codex 존재/부재) 모두에서 호출한 뒤 다시 스냅샷과 해시를 뜨면
- **Then** 파일 목록과 해시가 호출 전과 **동일**하다(생성·수정·삭제 0건)
- **RED-now** (트리 `ace1c5440`):
  ```
  $ go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredReadOnly > /tmp/t499-red4.txt 2>&1; echo $?
  0
  $ cat /tmp/t499-red4.txt
  ok  	github.com/modu-ai/moai-adk/internal/cli	0.691s [no tests to run]
  ```
- **GREEN 경로**: M2가 `TestCheckCodexWiring_HalfWiredReadOnly`를 만든다. 더해 실제 반쪽 프로젝트에서 `moai doctor > out.txt 2>&1; echo $?` → `0`(종료코드는 `CheckFail` 수만 센다 — `CheckWarn`에는 둔감하며, REQ-CPW-007의 종료코드 절에 대해서는 그 민감도가 옳다).

### AC-CPW-007 — 폭 상한 (D1 수리: 반쪽 경로를 실제로 밟는 시험을 추가)

- **Given** 반쪽 배선 프로젝트를, **두 PATH 갈래 모두**에 대해
- **When** `checkCodexWiring`을 호출하면
- **Then** 두 갈래 각각에서 `utf8.RuneCountInString(check.Message) <= codexMessageWidthCeiling`(113)이며, `--verbose` 렌더에서도 패널 폭이 밴드를 벗어나지 않는다
- **판정 명령**(신규 1 + 기존 2):
  - 신규(게이트): `go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredMessageWidthStaysInBand -v`
  - 기존(회귀 가드로만 유지): `go test ./internal/cli/ -run 'TestCheckCodexWiring_(MessageWidthStaysInBand|RenderedPanelStaysInBand)' -v`
- **RED-now** (트리 `ace1c5440`) — 신규 시험 부재:
  ```
  $ go test ./internal/cli/ -run TestCheckCodexWiring_HalfWiredMessageWidthStaysInBand > /tmp/t499-red5.txt 2>&1; echo $?
  0
  $ cat /tmp/t499-red5.txt
  ok  	github.com/modu-ai/moai-adk/internal/cli	0.683s [no tests to run]
  ```
- **왜 기존 두 시험만으로는 안 되는가**(감사 D1, 실측): 두 시험 모두 `checkCodexWiring(t.TempDir(), …)`를 호출한다(`internal/cli/doctor_codex_test.go:394`, `:441`) — 빈 임시 디렉터리, 즉 REQ-CPW-006이 얼려 놓은 `unwired` 경로다. **반쪽 경로에 결코 도달하지 못하므로 새 문구의 폭을 관측할 수 없다.** 반쪽 갈래에 200 rune짜리 Message를 주는 뮤턴트가 이 둘을 통과한다. 따라서 이 둘은 남기되 **회귀 가드**로만 세고, 게이트는 신규 시험이 진다.
- **GREEN 경로**: M2가 `TestCheckCodexWiring_HalfWiredMessageWidthStaysInBand`를 만든다.

## §D.2 회귀 가드

### AC-CPW-003 — 기존 세 시험이 수정 없이 통과

- **Given** run-phase 종료 시점의 트리
- **When** `git diff -- internal/cli/doctor_codex_test.go`에서 기존 세 시험 본문(`TestCheckCodexWiring_ClaudeOnlyMachineStaysSilent`, `TestCheckCodexWiring_InactiveProjectInformationalSkip`, `TestCheckCodexWiring_UnwiredWithCodexInstalledWarns`)이 **변경되지 않았음**을 확인하고 세 시험을 실행하면
- **Then** 세 시험 모두 `--- PASS`이며, 이 셋의 단언 문구는 어느 것도 완화되지 않았다
- **판정**: `go test ./internal/cli/ -run 'TestCheckCodexWiring_(ClaudeOnlyMachineStaysSilent|InactiveProjectInformationalSkip|UnwiredWithCodexInstalledWarns)' -v` → 세 줄 모두 `--- PASS`. 시험 본문이 수정되었다면 이 AC는 FAIL이며, 수정 사유를 `progress.md` §E.2에 적고 리드 판단을 받는다
- **RED-now 없음**: 구현 전에도 초록이다(보존 기준의 성질). 그래서 게이트가 아니라 가드다.

### AC-CPW-004 — 골든 3본이 재생성 없이 통과

- **Given** `internal/cli/testdata/doctor-{light,dark,nocolor}.golden`
- **When** `UPDATE_GOLDEN` 없이 골든 시험을 실행하면
- **Then** 세 시험 모두 `--- PASS`이고, `git diff --stat internal/cli/testdata/`가 **빈 출력**이다(골든 파일 0바이트 변경)
- **판정**: `go test ./internal/cli/ -run TestDoctorGolden -v > out.txt 2>&1; echo $?` → `0`, `--- PASS` ×3; 이어서 `git diff --stat internal/cli/testdata/ | wc -l` → `0`
- **RED-now 없음**: 위와 같은 이유.

### AC-CPW-008 — 존재-게이트 무접촉 (D5 수리: REQ-CPW-010으로 재매핑)

- **Given** run-phase 종료 시점의 트리
- **When** `git diff -U0 internal/codexwiring/wire.go`를 읽으면
- **Then** `RefreshWiring` / `wireProject` / `wiringFilesExist`의 **동작 변경이 없다**(REQ-CPW-010). 배선 상태 판별을 위한 **상수 추가**는 허용되며, 그 경우 추가된 줄을 인용해 세 함수의 동작과 무관함을 보인다
- **판정**: `git diff -U0 internal/codexwiring/wire.go > out.txt 2>&1; echo $?` 후 변경 줄을 눈으로 확인. 더해 `go test ./internal/codexwiring/ > out.txt 2>&1; echo $?` → `0`
- **매핑 정정 이력**: 초판은 이 기준을 REQ-CPW-007로 매핑했으나 그 요구의 주체는 doctor 검사이지 `codexwiring`이 아니다(감사 D5). 0.2.0에서 REQ-CPW-010을 신설해 그쪽으로 옮겼다.
- **RED-now 없음**: 보존 기준.

## §D.3 반증 관문 — 뮤턴트 5종

### AC-CPW-009 — 다섯 뮤턴트가 모두 RED를 만든다

각 뮤턴트는 **심고 → 해당 시험이 FAIL 하는 것을 관측하고 → 원복**한다. 원복 확인은 `git diff --stat internal/cli/` 빈 출력.
여기서는 뮤턴트 관측 자체가 RED 셀이다.

| # | 심는 변형 | RED가 되어야 하는 시험 | 어떤 요구를 지키는가 |
|---|---|---|---|
| M1 | 반쪽 판별을 다시 배선 파일 2종만 보는 한 줄로 되돌린다(agent 조회 제거) | `..._HalfWiredCodexAbsent`, `..._HalfWiredCodexInstalled` | REQ-CPW-001 |
| M2 | agent 정의 존재 판별을 "디렉터리가 존재하는가"로 바꾼다(파일 유무를 보지 않음) | `..._HalfWiredCountIndependent`의 빈 디렉터리 케이스 | REQ-CPW-005 |
| M3 | codex 부재 갈래의 Message를 기존 `not wired (claude-only project) — skipped`로 되돌린다 | `..._HalfWiredCodexAbsent`(`claude-only` 0회 단언) | REQ-CPW-004, REQ-CPW-006 |
| M4 | codex **부재** 갈래 Message 끝에 `— run moai init --agent codex`를 붙인다 | `..._HalfWiredCodexAbsent`(`initCodexAdvice` 부재 단언) | REQ-CPW-009 |
| M5 | 반쪽 갈래 Message를 200 rune로 늘린다 | `..._HalfWiredMessageWidthStaysInBand` | REQ-CPW-008 |

> M4·M5는 감사가 "아홉 기준을 전부 통과하면서 요구를 위반하는 뮤턴트"로 실제 지목한 둘이다(D3 / D1).
> 이 둘이 RED가 되지 않으면 D3·D1 수리가 이름만 들어간 것이므로, 그 상태에서는 어떤 AC도 PASS로 적지 않는다.
> M1이 RED가 되지 않으면 신규 시험이 판별식을 잡고 있지 않다는 뜻이므로 마찬가지다.

## §D.4 Definition of Done

- [ ] 릴리스 게이트 5건(AC-CPW-001 / -002 / -005 / -006 / -007) 전부: 판정 명령을 실행하고, `--- PASS`와 **`[no tests to run]` 부재**를 함께 확인해 `progress.md` §E.2에 출력 그대로 인용
- [ ] 회귀 가드 3건(AC-CPW-003 / -004 / -008) 전부: 판정 명령 실행 + 출력 인용
- [ ] AC-CPW-009 뮤턴트 5종 각각 RED 관측 → 원복 → `git diff --stat internal/cli/` 빈 출력 확인
- [ ] `go test ./internal/cli/... ./internal/codexwiring/...` 통과 (전체 스위트는 CI 몫 — CLAUDE.local.md §4)
- [ ] `go vet ./internal/cli/... ./internal/codexwiring/...` 통과
- [ ] `gofmt -l internal/cli internal/codexwiring` 빈 출력
- [ ] 실제 반쪽 프로젝트에서 `moai doctor` 두 갈래를 눈으로 확인하고 출력을 증거로 남김 (절차는 `.moai/reports/t499/repro.md`와 동일)
- [ ] 범위 밖(`spec.md` §E) 항목을 건드리지 않았음을 `git diff --name-only`로 확인
- [ ] 모든 종료코드는 리다이렉트 후 `echo $?`로 읽는다(파이프 뒤 `$?` / `PIPESTATUS` 금지)
