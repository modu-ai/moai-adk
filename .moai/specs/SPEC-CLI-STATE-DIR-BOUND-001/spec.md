---
id: SPEC-CLI-STATE-DIR-BOUND-001
title: "state 디렉터리 탐색의 무한 상향 걷기 제거 — 보호된 프로젝트 루트 규약으로 수렴"
version: "0.3.1"
status: completed
created: 2026-08-22
updated: 2026-08-22
author: manager-spec
priority: High
phase: "v3.1.2"
module: internal/cli
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "cli, state-dir, project-root, silent-success, convention-convergence"
related_specs: [SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001]
---

# SPEC-CLI-STATE-DIR-BOUND-001

## §A Problem / Motivation

`internal/cli/state.go:231` 의 `findStateDirFrom` 은 시작 디렉터리에서 **파일시스템 루트까지 제한 없이** 위로 올라가며 `.moai/state` 를 가진 첫 조상을 반환한다. 함수 자신의 주석이 이미 위험을 인정하고 있다.

> "A caller whose working directory sits inside the user's home therefore inherits any `~/.moai/state` on the machine"

카드 t161 은 테스트 3곳에 명시적 `StateDir` 을 주입했을 뿐, **프로덕션 걷기는 그대로**다.

### 관측된 증거 (2026-08-22, 이 머신에서 실제 실행)

전제: `~/.moai/state` 가 존재한다 (`drwxr-xr-x 8 goos staff`).

```
명령 A:  cd ~/t164probe/sub && unset CLAUDE_PROJECT_DIR && moai state show-blocker
출력 A:  · No blockers found                       (rc=0 — 성공)

명령 B:  cd /tmp/t164probe/sub && unset CLAUDE_PROJECT_DIR && moai state show-blocker
출력 B:  ERROR  Find state dir: .moai/state/ directory not found from /tmp/t164probe/sub.
```

A 와 B 의 유일한 차이는 프로브 디렉터리가 `$HOME` 아래에 있느냐뿐이다. **A 는 호출자가 지목한 적 없는 state 디렉터리를 상대로 조용히 성공했다.** 이것이 카드가 말한 "조용한 성공"이며, 추론이 아니라 실증이다. (프로브 디렉터리는 실행 후 제거했다.)

### 폭발 반경 — 소비자 6곳 + 패키지 밖 writer 1곳 + 테스트측 의존 1곳

| # | 위치 | 함수 | 성격 | `CLAUDE_PROJECT_DIR` 을 오늘 읽는가 | 조용한 성공의 결과 |
|---|---|---|---|---|---|
| 1 | `internal/cli/clean.go:65` | `runClean` | **파괴적** (`os.RemoveAll`, clean.go:116) | **아니오** | 찾은 state 디렉터리 아래 `runs/` 를 **삭제**한다. 반경 최대 |
| 2 | `internal/cli/tokens.go:377` | `resolveTokensStateDir` | 추가(append) | 아니오 | 걷기가 먼저, `<cwd>/.moai/state` 는 폴백. 홈 적중이 신규 체크아웃 폴백을 **선점**한다. windows 전용 CI flake 의 발원지 |
| 3 | `internal/cli/chain.go:67` | `resolveChainStore` | 추가 + 디렉터리 생성 | **예** (chain.go:64 — 6곳 중 유일) | `<found>/../chain` 을 `MkdirAll` 하므로 `~/.moai/chain` 을 **생성**한다 |
| 4 | `internal/cli/chain.go:353` | `loadRegistryForOverlay` | 읽기 (fail-open) | 아니오 | 엉뚱한 레지스트리를 조용히 읽는다 |
| 5 | `internal/cli/state.go:78` | `runStateDump` | 읽기 | 아니오 | 엉뚱한 state 를 덤프 |
| 6 | `internal/cli/state.go:154` | `runShowBlocker` | 읽기 | 아니오 | 위 프로브가 이 경로 |

**`CLAUDE_PROJECT_DIR` 을 오늘 읽는 소비자는 6곳 중 정확히 1곳(#3)뿐이다** — grep 실측: `clean.go` / `tokens.go` / `state.go` 각 0건. 이 비대칭이 §C REQ-1/REQ-9 분기의 사실 근거다.

**패키지 밖 writer 1곳**: `internal/hook/chain_event.go:67` — `filepath.Join(projectDir, ".moai", "state", "chain")` 을 **하드코딩**해서 chain 이벤트를 쓴다. `ChainStateDir` 상수도, `resolveChainStore` 도 경유하지 않는다. 프로젝트 디렉터리 해석 순서는 `payload.ProjectDir` → `CLAUDE_PROJECT_DIR` → `payload.CWD`(chain_event.go:55-61). REQ-7 정본 선택을 **독립적으로 뒷받침하는 증거**이자, REQ-7 케이스 2("양쪽 존재")가 예외가 아니라 **기대 상태**임을 뜻하는 근거다(§E R4).

**테스트측 의존 1곳 (프로덕션 소비자 아님)**: `internal/cli/state_m2_test.go:37` `m2SetupState` — 임시 트리에 `.moai/state` 를 만들고 테스트를 그리로 `chdir` 한다("so findStateDir resolves it"). **이 변경에 영향받지 않는다**: 경로 문자열을 비교하지 않고 해석된 디렉터리를 통해 동작하며, 심볼릭 링크 정규화 전후의 경로는 같은 물리 디렉터리를 가리키기 때문이다.

### 핵심 재구성 — 답은 "경계"도 "플래그"도 아니라 **수렴**이다

카드는 "상한 vs `--state-dir` 플래그"를 물었다. 그러나 코드베이스는 **이미 두 답을 모두 갖고 있다**. `findStateDir` 은 보호 장치가 있는 기존 규약의 **보호 없는 중복 구현**이다.

**(a) `internal/core/project/root.go:20` `FindProjectRoot()`** — 같은 상향 걷기지만, t164 가 원하는 바로 그 가드를 이미 갖고 있다. `paths.Home()` 으로 홈 디렉터리를 해석하고, 홈을 뚫고 올라가는 대신 에러를 반환한다. 주석: *"~/.moai/ is treated as global state (credentials, cache), not a project root."* 그리고 `internal/cli` 는 **이미** `internal/core/project` 를 임포트한다(비테스트 5개 파일) — 순환 임포트 없음, 신규 의존성 없음.

**(b) `internal/cli/harness.go:101` `resolveProjectRoot(cmd)`** — 플래그 우선, 없으면 cwd, 걷기 없음. harness 커맨드 계열 전체가 `--project-root` 를 명시적 오버라이드로 쓰는 규약이다. 이것이 카드의 "`--state-dir` 노출" 옵션의 선례이며, **두 번째 플래그를 새로 만들 게 아니라 `--project-root` 를 재사용하라**는 논거다.

추가로 `internal/config/token_budget_guard.go:51` `findRepoRoot` 는 `go.mod` 마커 기반의 **세 번째** 걷기다. 수정 대상이 아니라 **규약 파편화의 증거**로 기록한다.

### 수렴이 데려오는 네 가지 부수 결과

수렴은 공짜가 아니다. `FindProjectRoot` 는 `findStateDirFrom` 과 다르게 동작하고, `CLAUDE_PROJECT_DIR` 을 공유 헬퍼로 끌어올리면 오늘 그것을 읽지 않던 소비자들의 **해석 모드**가 바뀐다. 넷 모두 관측 가능한 변화이며 §E 표에 완전히 열거한다.

1. **마커가 다르다** (`.moai` vs `.moai/state`) → REQ-3 + AC-004/AC-011.
2. **경로 문자열을 정규화한다** (`EvalSymlinks`) → REQ-8 + AC-010.
3. **`chain.go` 의 두 분기는 오늘 다른 경로를 만든다** → REQ-7 + AC-008/AC-009.
4. **소비자 4곳이 `CLAUDE_PROJECT_DIR` 우선순위를 새로 얻고, 파괴적 소비자 1곳은 의도적으로 받지 않는다** → REQ-1 + REQ-9 + AC-015.

## §B Scope

**In Scope**:
- `findStateDir` / `findStateDirFrom` 의 해석 규약을 보호된 규약으로 수렴 (REQ-2..REQ-4).
- `CLAUDE_PROJECT_DIR` 우선 확인을 공유 헬퍼로 끌어올리되, **읽기·추가 소비자에게만** 적용 (REQ-1) 하고 **파괴적 소비자는 제외** (REQ-9).
- chain 디렉터리 경로 통일과 레거시 데이터 처분 선언 (REQ-7).
- 반환 경로의 정규화 계약 선언 (REQ-8), `CLAUDE_PROJECT_DIR` 값 포함.
- 기존 테스트 `internal/cli/tokens_state_dir_test.go` 의 서브테스트 3 강화 (REQ-5).
- 관측된 A/B 비대칭을 완전 소유 임시 트리 안에서 재현하는 회귀 테스트 신규 추가.

**Out of Scope**: 아래 항목들.

### Out of Scope — Windows CI 조상 `.moai/state` 의 생성 주체
- Windows CI 러너에서 무엇이 조상 `.moai/state` 를 만드는지는 **아직 미확인**이며 이 SPEC 은 답하지 않는다. 정적 후보 3개(`deps.go:120-126` `InitDependencies` → `paths.StateDir()` lazy `MkdirAll`; `statusline/model_cache.go:52` `WriteModelCache` — 현재 프로덕션 호출자 0; `config/cache.go` `LoadWithCache` 의 `<configDir>/state` 계열)와 제외 1건(`deps.go:124` 의 `<TEMP>/.moai/state` 폴백 — Linux 도 동일하게 깨질 것이므로 제외)은 §E 에 잔여 위험으로만 이월한다.
- 확인용 저비용 제안(`release-pr-multi-os.yml` 테스트 스텝 앞뒤에 `dir %USERPROFILE%\.moai` / `dir %TEMP%\.moai` 진단 2줄, 테스트 변경 0)은 **후속 카드** 소관이다.

### Out of Scope — `internal/hook/chain_event.go` 의 경로 리터럴 중복
- 훅 writer(`chain_event.go:67`)는 정본 경로를 **하드코딩**한다. 그 리터럴을 공유 상수로 끌어오는 리팩터는 이 SPEC 이 하지 않는다. 이유 둘: (1) 훅은 이미 **정본 경로를 쓰고 있어** 이 SPEC 이 고치려는 결함을 갖고 있지 않다 — 고칠 것이 없다. (2) 훅 프로세스는 `internal/cli` 를 임포트하지 않으므로 상수 공유에 새 의존 방향이 필요하고, 그것은 이 카드의 범위를 넘는다. 후속 정리 후보로만 기록한다.

### Out of Scope — `internal/hook` 의 `findRegistryUpward` (같은 결함, 다른 패키지)
- `internal/hook/cwd_changed_relocate.go:78` `findRegistryUpward` 는 `.moai/state/active-sessions.json`(`session.DefaultRegistryPath`)을 찾아 **파일시스템 루트까지 제한 없이** 올라가며 **홈 가드가 없다**. 이 SPEC 이 고치는 것과 **같은 결함 형태이며 같은 마커 계열**이다.
- 그럼에도 여기서 고치지 않는 이유: 이 카드의 범위는 `internal/cli` 의 state 해석이고, `internal/hook` 은 다른 프로세스 진입점(훅)에서 다른 수명주기로 동작하므로 회귀 표면과 검증 방법이 다르다. 한 카드에 두 패키지의 해석 규약을 함께 바꾸면 실패 시 원인 분리가 어려워진다.
- **후속 카드 후보**로 명시한다 — Windows CI 진단(R1), 훅 writer 리터럴 공유와 같은 성격의 이월 항목이다. REQ-6 의 범위 한정과 §E R2 의 재계수가 이 사실을 SPEC 본문에 고정한다.

### Out of Scope — `internal/config` 의 `findRepoRoot` 리팩터
- `token_budget_guard.go:51` 의 `go.mod` 마커 걷기는 이 SPEC 에서 손대지 않는다. 규약 파편화의 증거로 §A 에 기록만 한다.

### Out of Scope — 템플릿 미러
- 이 변경은 `internal/` 아래 Go 코드이며 `internal/template/templates/` 에 대응 미러가 없다. **Template-First 미러 의무는 적용되지 않는다.** run-phase 가 미러를 찾아 헤매지 않도록 명시한다.

### Out of Scope — chain 이벤트 파일의 병합과 범용 마이그레이션 프레임워크
- REQ-7 은 chain 이벤트 파일 하나에 대한 일회성 처분만 정의한다. 순서 있는 JSONL 두 개의 **병합은 하지 않는다** — 타임스탬프 신뢰·중복 제거·순서 결정을 요구하며 이 카드의 범위를 넘는다. 다른 상태 파일의 재배치나 범용 마이그레이션 도구도 만들지 않는다.

### Out of Scope — 소비자 6곳의 시맨틱 재설계
- 각 소비자가 state 디렉터리를 어떻게 **쓰는지**(clean 의 보존 정책, tokens 의 원장 스키마 등)는 불변이다. 바뀌는 것은 **어디를 가리키는가**와 REQ-7 이 정한 chain 경로뿐이다.

## §C Requirements (GEARS)

### REQ-1 — 읽기·추가 소비자는 명시적 지목을 최우선으로 한다

`CLAUDE_PROJECT_DIR` 이 설정되어 있을 때, 읽기 전용 및 추가(append) 성격의 state 해석은 그 값을 프로젝트 루트로 사용하고 어떤 상향 걷기도 수행하지 않아야 한다(SHALL). 대상은 §A 표의 #2 `tokens.go:377`, #3 `chain.go:67`, #4 `chain.go:353`, #5 `state.go:78`, #6 `state.go:154` 다섯 곳이다.

**"그대로"의 의미 — REQ-8 과의 관계 확정.** `CLAUDE_PROJECT_DIR` 값은 **사용 직전 `filepath.EvalSymlinks` 로 정규화된 뒤** 사용되어야 한다(SHALL). 여기서 "그대로"는 *걷기를 하지 않는다*는 뜻이지 *정규화하지 않는다*는 뜻이 아니다. 이 확정이 없으면 REQ-8(무조건 정규화)과 정면으로 충돌한다 — 프로덕션의 `CLAUDE_PROJECT_DIR` 독자 12곳은 전부 정규화 없는 `os.Getenv` 이고, darwin 에서 `/var/...` 형태 값이 실제로 들어온다.

**선언된 동작 변경 4/4 의 절반.** 오늘 이 다섯 중 `CLAUDE_PROJECT_DIR` 을 읽는 곳은 `chain.go:67` **하나뿐**이다(grep 실측). 나머지 넷은 해석 **모드**가 바뀐다 — 환경변수가 걷기 결과와 다른 루트를 가리키면 어제와 다른 곳을 해석한다. 이들에게 안전한 이유: 최악의 결과가 **지목된 프로젝트에서 읽거나 추가하는 것**이며, 삭제가 아니다. `chain.go:67` 의 `MkdirAll` 은 지목된 프로젝트 안에 디렉터리를 만들 뿐이다.

### REQ-9 — 파괴적 소비자는 환경변수를 따르지 않는다

`runClean`(§A 표 #1, `clean.go:65`)이 state 디렉터리를 해석할 때, `CLAUDE_PROJECT_DIR` 을 참조해서는 안 되며(SHALL NOT), cwd 에서 도출한 프로젝트 루트만을 기준으로 삼아야 한다(SHALL).

**근거.** `clean.go:116` 은 `os.RemoveAll` 이다. 이 리포는 `$CLAUDE_PROJECT_DIR` 이 워크트리에 있는 에이전트의 실제 디렉터리를 **추적하지 못한** 사례를 문서로 갖고 있다(`main-checkout-branch-guard.md` § Mechanical Enforcement — 그래서 판별자가 `input.CWD` 로 옮겨졌다). 환경변수를 일괄 적용하면 워크트리 안에서 실행한 `moai clean` 이 primary 체크아웃 아래를 삭제한다 — **이 SPEC 의 발단이 된 위험("호출자가 지목한 적 없는 디렉터리")이 다른 문으로 되돌아온다.**

다른 프로젝트를 청소하려면 **그 디렉터리에서 실행한다.** 삭제 대상은 상속된 환경변수가 아니라 명시적 행위로 정해져야 한다.

### REQ-10 — 해석된 루트를 출력으로 드러낸다

여섯 소비자 각각은 state 디렉터리를 해석한 뒤 **행동하기 전에**, 해석된 프로젝트 루트를 명시하는 한 줄을 사용자가 볼 수 있는 출력으로 남겨야 한다(SHALL). 최소한 `runClean` 은 삭제 후보를 열거하거나 삭제하기 전에 그 줄을 남겨야 한다(SHALL).

**근거 — REQ-1/REQ-9 분할이 만드는 동시 분기.** REQ-9 는 `clean` 을 환경변수에서 떼어냈고, 그것은 옳다. 대가는 **한 세션 안에서 두 해석이 갈릴 수 있다**는 것이다: `CLAUDE_PROJECT_DIR` 이 설정되어 있고 cwd 도출 루트와 다르면, `moai state dump` 는 프로젝트 B 를 읽고 `moai clean` 은 프로젝트 A 에서 지운다 — **같은 세션, 아무 신호 없이.** 점검하고 나서 청소하는 운영자는 한 프로젝트를 보고 다른 프로젝트에 손을 대게 된다.

이 분기는 §E 에 **선언된 수용 결과**로 기록한다(동작 변경 표 아래 "선언된 부수 결과" 항). 분기 자체를 없애는 것은 REQ-9 를 되돌리는 것이므로 선택지가 아니다. 완화는 **분기를 보이게 만드는 것** 하나뿐이며, 그것이 이 요구사항이다. 부수 효과로 acceptance.md §D.6 의 `clean` 사후 확인이 자기증거화된다 — 열거된 후보가 어느 루트 아래인지 출력이 직접 말한다.

§E R8(분류가 낡는 위험)은 이 항목을 덮지 않는다. R8 은 **시간이 지나며 분류가 낡는 것**이고, 이것은 **지금 동시에 갈리는 것**이다.

### REQ-2 — 걷기는 홈 디렉터리를 통과하지 않는다

`CLAUDE_PROJECT_DIR` 이 없을 때(그리고 REQ-9 대상은 항상), state 해석은 `internal/core/project` 의 보호된 프로젝트-루트 규약에 위임하여야 하며(SHALL), 사용자 홈 디렉터리를 조상으로 삼아 계속 올라가서는 안 된다(SHALL NOT).

해석 결과는 `<project-root>/.moai/state` 이다.

### REQ-3 — 프로젝트 루트에서 멈추고 실패한다 (마커 불일치 포함)

`.moai` 는 있으나 `.moai/state` 가 없는 디렉터리에서 해석이 멈출 때, 해석은 **그 자리에서 실패**하여야 한다(SHALL) — 조상으로 계속 올라가 성공해서는 안 된다(SHALL NOT).

**선언된 동작 변경 1/4.** 마커가 `.moai/state` 에서 `.moai` 로 바뀌면서 두 부류의 트리가 영향받는다.

- **부류 A — 프로젝트 루트에 `.moai` 는 있고 `.moai/state` 가 없는 경우.** 오늘은 계속 올라가 조상에서 성공할 수 있다. 변경 후에는 프로젝트 루트에서 실패한다. *엉뚱한 자리에서 성공하는 것보다 옳은 자리에서 실패하는 것이 낫다.*
- **부류 B — 유효한 프로젝트 루트 아래에, `state` 없는 맨 `.moai` 를 가진 하위 디렉터리가 있고 그 안에서 실행하는 경우.** 오늘은 그 하위 디렉터리를 **건너뛰고** 위의 진짜 루트에서 성공한다. 변경 후에는 하위 디렉터리에서 멈추고 실패한다 — **정답이 도달 가능했는데 거부하는 케이스**이므로 부류 A 의 정당화가 그대로 적용되지 않는다. 이 리포는 실제로 그 형태를 가진 이력이 있다(`internal/hook/.moai/`).

부류 B 는 **의도적으로 감수하는 회귀**다. 근거: 하위 디렉터리의 맨 `.moai` 자체가 이상 상태이며, 그것을 조용히 건너뛰는 오늘의 동작은 같은 함수가 `~/.moai` 를 건너뛰지 못하는 것과 같은 뿌리(마커 불일치)에서 나온다. 하나만 고치면 다른 하나가 남는다. 실패 메시지는 어느 디렉터리에서 멈췄는지 지목하여야 한다(SHALL).

### REQ-4 — 신규 체크아웃 폴백은 보존된다

`resolveTokensStateDir` 이 프로젝트 루트를 찾지 못할 때, `<cwd>/.moai/state` 폴백은 그대로 동작하여야 한다(SHALL) — 사전 스캐폴딩 없는 신규 체크아웃에서 원장 기록이 가능해야 한다.

### REQ-5 — 기존 테스트의 유효 계약은 보존하고, 환경 의존 서브테스트만 강화한다

`internal/cli/tokens_state_dir_test.go` 의 `TestFindStateDirFromWalksUp` 세 서브테스트는 아래와 같이 처리되어야 한다(SHALL).

| 서브테스트 | 오늘의 단언 | 변경 후 |
|---|---|---|
| 1 — "an ancestor state dir wins over the starting directory" | 조상 `.moai/state` 가 시작 디렉터리를 이긴다 | **보존.** 단언은 그대로 참이다 |
| 2 — "the starting directory wins over an ancestor" | 시작 디렉터리의 자체 state 가 이긴다 | **보존.** 경로 비교만 REQ-8 에 맞춰 정규화 |
| 3 — "nothing inside the tree is claimed…" | 트리 밖 결과는 단언하지 않음(환경 의존) | **강화.** 결정적 에러를 직접 단언 |

**iter1 정정 (유지).** 서브테스트 1 의 트리는 `root := t.TempDir()` 에 `.moai/state` 가 **루트에** 있고 시작점이 그 아래다. 권고안에서도 해석은 위로 올라가 `root` 에서 `.moai` 를 찾고 같은 조상을 반환한다. 시작점과 `root` 사이에 홈 경계가 없으므로 가드가 발동할 여지가 없다 — **무변경으로 통과한다.** 이 변경이 바꾸는 것은 **홈 경계를 넘는 해석뿐**이며 그것은 서브테스트 3 의 영역이다.

어떤 서브테스트도 `t.Skip` 으로 무력화되어서는 안 된다(SHALL NOT).

### REQ-6 — `internal/cli` 안의 걷기 구현은 하나이고, 파괴적 소비자는 대상을 명시한다

**`internal/cli` 의 state 해석**에서 상향 걷기 구현은 **정확히 하나**여야 하며(SHALL), 그 패키지의 소비자가 각자 재구현해서는 안 된다(SHALL NOT). 읽기·추가 소비자 다섯은 동일한 해석(환경변수 우선 + 보호된 걷기)을 공유하여야 하고(SHALL), 파괴적 소비자 하나는 같은 걷기를 쓰되 환경변수 분기를 갖지 않아야 한다(SHALL — REQ-9).

**[HARD] 범위 한정은 장식이 아니다.** 이 요구사항은 리포 전체를 구속하지 않으며, 구속한다고 읽어서도 안 된다 — 착지 시점에 **패키지 밖에 두 번째 무제한 걷기가 살아남기 때문**이다. `internal/hook/cwd_changed_relocate.go:78` `findRegistryUpward` 는 같은 `.moai/state` 마커 계열(`session.DefaultRegistryPath = ".moai/state/active-sessions.json"`)을 대상으로 파일시스템 루트까지 올라가며, **홈 가드가 없다** — 이 SPEC 이 없애려는 바로 그 결함 형태다. 이 SPEC 은 그것을 고치지 않는다(§B 범위 밖 + §E R2 + 후속 카드). 범위를 붙이지 않은 "정확히 하나"는 착지 즉시 거짓이 되고, 그 문장을 읽은 다음 사람이 실재하지 않는 보증 위에서 작업하게 된다.

**iter2 로부터의 변경 — 전제 재개.** iter2 의 REQ-6 은 "소비자 6곳 **전부**가 동일한 해석을 공유한다"였다. 그 전제는 iter2 감사 E6 이 되열었다: 여섯 중 하나가 삭제하고, 그 하나에 환경변수 우선순위를 주는 것은 안전하지 않다. 따라서 불변식을 **"하나의 걷기 · 읽기 계열은 하나의 해석 · 파괴적 소비자는 대상을 명시"** 로 다시 쓴다. 중복 구현을 없앤다는 원래 목적은 그대로다 — 갈라지는 것은 걷기가 아니라 환경변수 분기 하나뿐이다.

### REQ-7 — chain 디렉터리 경로를 통일하고 레거시 데이터 처분을 선언한다

chain 디렉터리 해석의 결과는 `CLAUDE_PROJECT_DIR` 설정 여부와 무관하게 **`<project-root>/.moai/state/chain`** 이어야 한다(SHALL) — **단, 아래 경우 4(이전 실패)는 이 조항의 명시적 예외다.**

**선언된 동작 변경 3/4.** 오늘 두 분기는 서로 다른 경로를 만든다 — 이 SPEC 이 만든 문제가 아니라 이 SPEC 이 드러낸 기존 불일치다.

```go
const ChainStateDir = ".moai/state/chain"                     // chain.go:29
env  분기: filepath.Join(projDir, ChainStateDir)              // <proj>/.moai/state/chain
walk 분기: filepath.Join(filepath.Dir(stateDir), "chain")     // <root>/.moai/chain
```

`filepath.Dir("<root>/.moai/state")` = `<root>/.moai` 이므로, 같은 프로젝트가 환경변수 설정 여부에 따라 **다른 위치에 이벤트를 쓴다**.

**정본은 `.moai/state/chain` 이다.** 근거 넷:

1. **`internal/hook/chain_event.go:67` 의 훅 writer 가 이 경로를 하드코딩해 늘 써 왔다.** 상수보다 강한 증거다 — 실제로 이벤트가 그 경로에 쌓여 왔고, 정본을 `.moai/chain` 으로 고르면 CLI 와 훅이 영구히 갈린다.
2. `ChainStateDir` 상수(chain.go:29)가 그 값을 명시적으로 선언하고 있다 — walk 분기의 `filepath.Dir(...)+"chain"` 은 선언된 의도가 아니라 산술의 부산물이다.
3. chain 이벤트는 state 이므로 `.moai/state` 아래가 의미상 옳다.
4. `.moai/chain` 을 정본으로 삼으면 명시적으로 선언된 상수를 버리게 된다.

**영향받는 사용자와 처분**: `CLAUDE_PROJECT_DIR` **없이** `moai chain` 을 써 온 사용자의 이벤트 파일이 `<root>/.moai/chain/events.jsonl` 에 있다. 이 파일은 **조용히 고아가 되어서는 안 된다**(SHALL NOT). 네 경우로 규정한다.

| 경우 | 조건 | 처분 |
|---|---|---|
| 1 | 레거시에만 이벤트 파일 존재 | 정본 경로로 **일회 이전**하여야 한다(SHALL) |
| 2 | 양쪽 모두 존재 | 정본이 이기고 레거시는 **손대지 않으며**, 관측 가능한 경고를 남긴다(SHALL). 병합은 하지 않는다 |
| 3 | 정본에만 존재 | 아무 동작도 하지 않고 경고도 남기지 않는다 |
| 4 | **이전 실패** (대상 쓰기 불가 등) | **head clause 의 명시적 예외.** 그 호출은 레거시 경로를 계속 쓰고 관측 가능한 경고를 남겨야 한다(SHALL). 조용한 데이터 손실보다 조용하지 않은 분기가 낫다 |

경우 4 는 head clause 의 "무관하게 `.moai/state/chain`" 에 대한 **선언된 유일한 예외**다. 이 예외를 명문화하지 않으면 두 SHALL 이 서로 모순한다(iter2 감사 E7).

**경우 2 는 예외가 아니라 기대 상태다.** 훅 writer 가 늘 `.moai/state/chain` 에 써 왔고 CLI 가 `CLAUDE_PROJECT_DIR` 없이 `.moai/chain` 에 써 왔으므로, **두 표면을 모두 쓴 머신에서는 두 파일이 이미 존재한다.** 드문 구석이 아니라 데이터를 가진 사용자층의 기본값이다 — §E R4 참조.

(이 체크아웃에는 `.moai/state/chain` 도 `.moai/chain` 도 존재하지 않으므로 이 리포 자체에 위험 데이터는 없다. 그러나 SPEC 은 전체 배포 사용자에게 나간다.)

### REQ-8 — 반환 경로의 정규화 계약

state 해석이 반환하는 경로는 — 환경변수 분기와 걷기 분기 **양쪽 모두에서** — `filepath.EvalSymlinks` 로 정규화된 형태여야 한다(SHALL). 이 계약은 함수 주석에 명시되어야 한다(SHALL).

**선언된 동작 변경 2/4.** `FindProjectRoot` 는 `root.go:27-30` 에서 `EvalSymlinks` 를 호출한다. `findStateDirFrom` 은 하지 않고 시작점의 접두사를 그대로 물고 반환한다. macOS 에서 `/var` 는 `private/var` 로의 심볼릭 링크이며, `t.TempDir()` 은 **미해석** 형태를, `getcwd(3)` 은 **해석된** 형태를 준다. 이 세션에서 실측:

```
$ python3 -c "import os,tempfile;d=tempfile.mkdtemp();os.chdir(d);print(d);print(os.getcwd())"
/var/folders/kt/.../tmp8b0_7lt7
/private/var/folders/kt/.../tmp8b0_7lt7
equal = False
```

환경변수 분기도 이 계약에 **포함**된다(REQ-1). `CLAUDE_PROJECT_DIR` 의 프로덕션 독자 12곳은 전부 정규화 없는 `os.Getenv` 이므로, 정규화하지 않으면 환경변수 분기에서만 계약이 깨진다.

## §D 선택지와 권고

**권고: Option 1.** 두 차례 감사에서 방향 자체는 유지 판정을 받았다.

| 안 | 내용 | 평가 |
|---|---|---|
| **1 (권고)** | `findStateDir` 을 보호된 규약에 위임: (읽기·추가) `CLAUDE_PROJECT_DIR` → `project.FindProjectRoot()` → `<root>/.moai/state` / (파괴적) cwd 기준 루트만. 신규 플래그 없음, 신규 걷기 없음 | 중복 구현이 하나로 줄고, 홈 가드를 공짜로 얻는다. 대신 §E 의 동작 변경 4건을 명시적으로 감수 |
| 2 | 걷기는 유지하고 홈 디렉터리 가드만 추가 | diff 최소이고 동작 변경도 REQ-3 부류 A 하나로 줄어든다. 그러나 걷기 구현이 **둘로 남고** 정규화·chain 경로 불일치도 그대로 남는다 |
| 3 | 깊이 N 상한 | **반대.** N 은 임의값이고 리포 구조를 추적하지 못한다 |
| 4 | 신규 `--state-dir` 플래그 | **주 수정안으로는 반대.** 이미 답을 아는 호출자만 돕고 **기본값을 고치지 못한다**. 관측된 A 케이스는 아무 플래그도 주지 않은 호출이었다. 추가적 탈출구가 필요하면 신규 플래그 대신 harness 계열의 기존 `--project-root` 규약(§A(b))을 재사용할 것 |

## §E Constraints / Non-Goals / 잔여 위험

### 선언된 동작 변경 — 전 4건

> iter1 은 1건, iter2 는 3건이라 적었고 둘 다 불완전했다. 아래가 현재 알려진 **전부**이며, 각 행은 요구사항과 검증 AC 를 가진다.

| # | 변경 | 영향 범위 | 요구사항 | 검증 |
|---|---|---|---|---|
| 1 | 마커가 `.moai/state` → `.moai` 로 바뀌어, 맨 `.moai` 를 가진 디렉터리에서 멈추고 실패한다 (부류 A/B) | 소비자 6곳 전부 | REQ-3 | AC-004, AC-011 |
| 2 | 반환 경로가 `EvalSymlinks` 로 정규화된 형태가 된다 (darwin 에서 문자열이 실제로 달라진다). 환경변수 값도 포함 | 소비자 6곳 전부 | REQ-8, REQ-1 | AC-010 |
| 3 | chain 디렉터리가 `.moai/state/chain` 으로 통일되고, `.moai/chain` 의 기존 이벤트가 일회 이전된다 (이전 실패는 선언된 예외) | `chain.go:67` | REQ-7 | AC-008, AC-009 |
| 4 | **읽기·추가 소비자 4곳(`tokens.go:377`, `chain.go:353`, `state.go:78`, `state.go:154`)이 `CLAUDE_PROJECT_DIR` 우선순위를 새로 얻는다** — 환경변수가 걷기 결과와 다른 루트를 가리키면 해석 위치가 이동한다. `clean.go:65` 는 **제외**되어 오늘과 같이 cwd 기준으로만 해석한다 | 읽기·추가 4곳 (모드 변경) / clean 1곳 (변경 없음) | REQ-1, REQ-9 | AC-015 |

`chain.go:67` 은 오늘 이미 환경변수를 읽으므로 4행의 모드 변경 대상이 아니다 — 여섯 중 넷이 새로 얻고, 하나는 이미 갖고 있었고, 하나는 의도적으로 받지 않는다. 4 + 1 + 1 = 6.

**4행 안전성 주장의 범위 (v0.3.1 한정).** "최악의 결과가 지목된 프로젝트에서 읽거나 추가하는 것"이라는 근거는 **소비자별로 확인된 것이 아니라 성격(읽기/추가 vs 삭제)에서 도출한 일반 논거**다. 소비자 하나에 대해서는 확인이 아직 없다: `loadRegistryForOverlay`(chain.go:353)가 읽는 세션 레지스트리는 **설계상 두 곳에 존재한다** — `internal/session/anchor.go:26-33` 이 측정 사실(2026-08-17)로 기록하듯, `LiveAnchoredSessions` 는 트리-로컬 레지스트리와 호출자 프로젝트 레지스트리를 **둘 다** 읽으며 그 이유는 어느 한쪽도 완전하지 않기 때문이다(`moai cc -w` 레인은 런처가 붙인 체크아웃 쪽에 등록되고, 워크트리 자신은 로컬 레지스트리를 갖지 않는다). `loadRegistryForOverlay` 는 그중 **하나만** 읽고, REQ-1 은 그 하나가 **어느 쪽인지**를 바꾼다.
`cc -w` 형태에서는 환경변수 우선이 레인이 실제로 등록하는 쪽을 가리키므로 이 변경은 오히려 개선이다. 다만 v0.3.0 까지의 서술은 레지스트리가 둘이라는 사실을 **모른 채** 안전을 단언했다. 따라서 4행의 안전성은 **일반 논거로는 유지하되, 이 소비자에 한해 "확인됨"이 아니라 "미확인"으로 표기**한다. 확인은 후속 몫이며, 이 SPEC 은 `loadRegistryForOverlay` 가 읽는 레지스트리 개수를 바꾸지 않는다.

### 선언된 부수 결과 — 한 세션 안의 해석 분기 (v0.3.1 신설)

동작 변경 4행과 별개로, REQ-1/REQ-9 분할은 **수용하기로 선언한 부수 결과** 하나를 만든다.

`CLAUDE_PROJECT_DIR` 이 설정되어 있고 cwd 도출 루트와 다를 때, 같은 세션 안에서 **읽기 계열 다섯은 프로젝트 B 를, `clean` 은 프로젝트 A 를** 해석한다. `moai state dump` 로 점검한 뒤 `moai clean` 을 실행하는 운영자는 한 프로젝트를 보고 다른 프로젝트에 손을 대게 된다.

- **되돌리지 않는다.** 분기를 없애려면 REQ-9 를 되돌려야 하고, 그러면 `os.RemoveAll` 이 다시 환경변수를 따라간다 — 이 SPEC 의 발단이 된 위험이다. 조용한 분기보다 **드러난 분기**가 낫다.
- **완화는 가시화 하나뿐**이며 REQ-10 이 그것이다: 각 소비자가 행동 전에 해석된 루트를 한 줄로 출력한다.
- R8(분류가 시간이 지나며 낡는 위험)과 **다른 항목**이다. R8 은 미래의 오분류, 이것은 현재의 동시 분기다.

### 제약

- **fail-open 보존**: `loadRegistryForOverlay`(chain.go:353)는 오늘 fail-open 이다. 해석 실패는 계속 `nil` 을 반환해야 한다. REQ-3 이 실패 경우를 **늘리므로** 이 경로 통행량이 증가한다 — AC-012 가 회귀 아님을 확인한다.
- **`internal/core/project/root.go` 는 읽기 전용**. 부득이 수정이 필요하면 그 자체가 설계 재검토 신호이며 근거를 남긴다.
- **`clean.go:136` 의 동일 산술은 검토했고 안전하다.** `loadRetentionDays` 가 `filepath.Dir(stateDir)` 로 `<root>/.moai/config/sections/state.yaml` 을 만든다. REQ-2 하에서도 state 디렉터리는 `<root>/.moai/state` 이므로 `Dir()` 는 여전히 `<root>/.moai` 를 준다 — **깨지지 않는다.** REQ-7 이 이 산술을 "부산물"이라 부르므로, 유일한 다른 사용처를 검토했음을 여기 기록한다(iter2 감사 E8).
- **로컬 전체 스위트 금지**: `go test ./...` 를 로컬에서 돌리지 않는다. 패키지 단위로 돌리고 전 패키지 판정은 CI 에 맡긴다.

### 잔여 위험

- **R1**: Windows CI 러너에서 조상 `.moai/state` 를 만드는 주체는 여전히 미확인이다. 이 SPEC 은 **증상을 없애지만 원인을 밝히지 않는다**. 후속 카드.
- **R2 (v0.3.1 재계수)**: 상향 걷기 규약이 **적어도 넷**이다 — `findStateDir`(`internal/cli/state.go`), `FindProjectRoot`(`internal/core/project/root.go`), `findRepoRoot`(`internal/config/token_budget_guard.go:51`, `go.mod` 마커), **`findRegistryUpward`(`internal/hook/cwd_changed_relocate.go:78`)**. 이 SPEC 은 앞의 둘만 수렴시킨다. 네 번째는 iter3 감사가 찾아냈고, 하필 REQ-6 이 통합을 주장하는 바로 그 마커(`.moai/state/active-sessions.json`)를 걸으면서 **홈 가드가 없다** — v0.3.0 까지 이 SPEC 은 "규약이 셋"이라 적었고 그것은 틀렸다. "적어도"라고 쓴 이유: 리포 전체의 상향 걷기를 전수 조사하지는 않았다. 넷을 세었을 뿐 넷이 전부라고 주장하지 않는다.
- **R3**: `chain.go:67` 의 `MkdirAll` 은 해석이 옳아진 뒤에도 여전히 생성 부작용이 있다. 옳은 자리에 생성되므로 무해하나, 생성-부작용 자체는 제거하지 않는다.
- **R4 (iter3 재서술)**: REQ-7 의 일회 이전은 이 SPEC 범위 안에서 가장 되돌리기 어려운 조작이다. 그리고 **경우 2("양쪽 존재")는 드문 경로가 아니라 기대 상태다** — 훅 writer(`chain_event.go:67`)가 `.moai/state/chain` 에, CLI 가 `CLAUDE_PROJECT_DIR` 없이 `.moai/chain` 에 써 왔으므로 **두 표면을 모두 쓴 머신에서는 두 파일이 이미 존재한다.** 그 사용자에게는 병합하지 않는 정책상 이력이 두 갈래로 남고, 레거시 쪽은 어떤 리더도 열지 않는다. iter2 는 이를 희귀 경로의 잔여 위험으로 서술했고 그것은 틀렸다 — 데이터를 가진 사용자층의 **기본 결과**다. 완화는 경고 문구가 이 상황을 정확히 설명하는 것뿐이며, 병합·정리는 후속 카드 소관이다.
- **R5**: REQ-3 부류 B 는 정답이 도달 가능한데 거부하는 회귀다. 이 리포의 `internal/hook/.moai/` 형태가 다시 나타나면 그 하위에서 `moai state`/`chain`/`clean` 이 실패한다. 완화는 명확한 실패 메시지뿐이다.
- **R6**: plan.md §E D3 seam 선택에 따라 프로덕션은 초록인데 테스트만 빨간 상태가 나올 수 있다 — 잘못된 수정을 유도하는 최악의 실패 모양이다. 그 선택을 M1 에서 내리고 근거를 남기도록 요구한다.
- **R7 (iter3 신규)**: REQ-1 이 `CLAUDE_PROJECT_DIR` 값을 정규화하기로 확정했으므로, 같은 변수를 읽는 **다른 11곳**과 정규화 여부가 갈린다. 이 SPEC 은 그 11곳을 바꾸지 않으므로, 같은 환경변수가 헬퍼 안팎에서 다른 문자열로 보이는 구간이 남는다. 물리 디렉터리는 같으므로 기능 영향은 없으나, 문자열 비교를 하는 코드가 훗날 헬퍼를 채택하면 드러날 수 있다.
- **R8 (iter3 신규)**: REQ-9 의 carve-out 은 "삭제하는가"라는 성격으로 소비자를 가른다. 훗날 읽기 계열 소비자가 삭제를 하게 되면 이 분류가 조용히 낡는다. 완화: `clean` 의 해석 진입점을 별도 이름으로 두어 성격이 코드에 드러나게 한다(plan.md §E D6).

## §F Cross-References

- `internal/cli/state.go:210-250` — `findStateDir` / `findStateDirFrom` (수정 대상).
- `internal/core/project/root.go:20-60` — `FindProjectRoot`, 보호된 규약 + `EvalSymlinks`(27-30행) (위임 대상, 읽기 전용).
- `internal/cli/harness.go:101` — `resolveProjectRoot`, 플래그 규약 선례.
- `internal/cli/chain.go:29` — `ChainStateDir = ".moai/state/chain"` (REQ-7 정본 근거 2).
- `internal/hook/chain_event.go:55-68` — 패키지 밖 chain writer, 정본 경로 하드코딩 (REQ-7 정본 근거 1, R4 의 근거).
- `internal/cli/chain.go:62-72` — 두 분기가 갈리는 지점.
- `internal/cli/clean.go:65` (해석), `clean.go:116` (`os.RemoveAll`), `clean.go:136` (동일 산술, 안전 확인됨) — REQ-9 의 대상.
- `internal/cli/tokens.go:377`, `chain.go:353`, `state.go:78`, `state.go:154` — REQ-1 의 모드 변경 대상 4곳.
- `internal/cli/tokens_state_dir_test.go:32` — 서브테스트 3개 (REQ-5).
- `internal/cli/state_m2_test.go:37` — `m2SetupState`, 테스트측 의존 (영향 없음).
- `internal/config/envkeys.go:290` — `EnvClaudeProjectDir` 정의; 프로덕션 독자 12곳 (REQ-1 정규화 확정 + R7).
- `.claude/rules/moai/workflow/main-checkout-branch-guard.md` § Mechanical Enforcement — `$CLAUDE_PROJECT_DIR` 이 워크트리를 추적하지 못한 선례 (REQ-9 의 근거).
- `internal/config/token_budget_guard.go:51` — `findRepoRoot`, 세 번째 걷기 (범위 밖).
- `internal/hook/cwd_changed_relocate.go:78` — `findRegistryUpward`, **네 번째 걷기 · 홈 가드 없음** (범위 밖, 후속 카드 후보; REQ-6 범위 한정 + R2 재계수의 근거).
- `internal/session/registry.go:39` — `DefaultRegistryPath = ".moai/state/active-sessions.json"` (네 번째 걷기가 거는 마커).
- `internal/session/anchor.go:26-33` — 레지스트리가 두 곳이라는 측정 사실 (§E 4행 범위 표기의 근거).
- `.moai/reports/t164/plan-audit.md` — iter1 감사 (FAIL 0.60, D1-D11).
- `.moai/reports/t164/plan-audit-iter2.md` — iter2 감사 (FAIL 0.667, E1/E4/E5/E6/E7).
- `.moai/reports/t164/plan-audit-iter3.md` — iter3 감사 (**PASS 0.80**, 잔여 부채 F1-F5). F4/F5 는 이 SPEC 에 접지 않고 그 보고서에 기록된 부채로 남긴다.
- `.github/workflows/release-pr-multi-os.yml` — 후속 진단 카드의 대상.

## §G HISTORY

- **2026-08-22** v0.3.1 — 정정 패스. plan-auditor iter3 **PASS 0.80**(임계와 동률) 후 잔여 부채 중 셋을 접었다. 설계는 건드리지 않았다 — REQ-1 분할·REQ-7·REQ-9·§E 4행 모두 무변경.
  - **F1 (거짓 요구사항 제거)**: REQ-6 의 "걷기 구현은 정확히 하나"는 착지 시점에 **거짓**이었다. `internal/hook/cwd_changed_relocate.go:78` `findRegistryUpward` 가 같은 마커 계열(`.moai/state/active-sessions.json`)을 **홈 가드 없이** 루트까지 걷는다. REQ-6 을 `internal/cli` 로 범위 한정하고, 왜 범위 한정이 장식이 아닌지를 [HARD] 로 명시. **§E R2 재계수** — "규약이 셋"을 "적어도 넷"으로 고치고 네 번째를 이름으로 지목("적어도"인 이유: 전수 조사를 하지 않았다). **Out of Scope 항목 신설**(같은 결함, 다른 패키지) + 후속 카드 후보로 등재. §F 에 3개 앵커 추가.
  - **F3 (동시 분기)**: **REQ-10 신설** — 여섯 소비자가 행동 전에 해석된 루트를 한 줄로 출력한다(최소 `clean`). §E 에 **"선언된 부수 결과" 절 신설** — `CLAUDE_PROJECT_DIR` 이 cwd 루트와 다를 때 읽기 계열은 B 를, `clean` 은 A 를 해석하는 동시 분기를 **수용 결과로 선언**하고, 되돌리지 않는 이유(REQ-9 회귀)와 완화가 가시화뿐임을 기록. R8 과 다른 항목임을 명시. AC-015 에 판정 명령 4 추가.
  - **AC-009 경고 고정**: 경우 2·4 의 "관측 가능한 경고"가 스트림도 문자열도 지목하지 않아 grep 불가능했다 — REQ-7 의 가장 위험한 두 분기가 그 단언에 걸려 있다. **stderr + 안정 부분문자열**로 고정.
  - **F2 (부분)**: §E 4행의 안전성 주장을 **일반 논거로 유지하되 `loadRegistryForOverlay` 한정으로 "미확인" 표기**. `internal/session/anchor.go:26-33` 의 두-레지스트리 측정 사실을 인용. 확대하지 않았다.
  - **F4/F5**: 접지 않았다. iter3 보고서에 기록된 부채로 남긴다(§F).
- **2026-08-22** v0.3.0 — iter3 개정. plan-auditor iter2 FAIL 0.667 의 MUST-FIX 5건 반영.
  - **E6 (핵심)**: REQ-1 을 읽기/파괴 축으로 분할. 읽기·추가 5곳은 `CLAUDE_PROJECT_DIR` 을 따르고(모드 변경 4곳 + 기존 1곳), 파괴적 `clean` 은 REQ-9 로 제외(`os.RemoveAll` + 워크트리 추적 실패 선례). REQ-6 의 "6곳 전부 동일" 전제를 "하나의 걷기 · 읽기 계열 하나의 해석 · 파괴적 소비자는 대상 명시"로 재작성하고 전제 재개 사실을 본문에 기록. §E 표를 4행으로 확장하고 6곳 배분(4 신규 + 1 기존 + 1 제외)을 재계산. AC-015 신설.
  - **E4**: REQ-1 의 "그대로"를 *걷기 없음*으로 확정하고 환경변수 값의 `EvalSymlinks` 정규화를 명시 — REQ-8 이 양 분기에 성립. AC-001/AC-010 을 환경변수 분기를 실제로 태우도록 재작성.
  - **E7**: REQ-7 head clause 에 "경우 4 는 명시적 예외" 를 명문화하고 처분을 4경우 표로 재작성. AC-009 에 경우 4 추가, AC-008 의 부정 단언을 성공 경로로 한정.
  - **E5**: `internal/hook/chain_event.go:67` 을 §A 표 아래·§F·REQ-7 정본 근거 1순위에 명시. R4 를 "경우 2 는 기대 상태"로 재서술.
  - **E1**: 계수 오류 수정 — `grep "findStateDir()"` 는 정의(state.go:212)와 주석(chain.go:61)까지 잡아 8을 낸다. acceptance.md AC-003 과 plan.md §C Pre-Flight 를 호출 전용 패턴으로 교체. `chain.go:61` 주석을 AC-014 스윕에 추가.
  - 부수: E8(`clean.go:136` 안전 확인)을 §E 제약에 기록, E9 반영해 REQ-7 에서 함수 이름 제거, R7·R8 신규 등록.
  - 방향(Option 1)은 두 감사 모두 유지 판정이라 변경하지 않았다.
- **2026-08-22** v0.2.0 — iter2 개정. plan-auditor iter1 FAIL 0.60 의 MUST-FIX 8건 반영: REQ-5 재도출(서브테스트 1 **보존**, iter1 의 "반전" 지시는 오류), REQ-7 신설(chain 정본 + 레거시 처분), REQ-8 신설(`EvalSymlinks` 정규화), §E 를 3건 표로 교체, REQ-3 마커 불일치 부류 A/B 분석, `m2SetupState` 명시, 프론트매터 정리.
- **2026-08-22** v0.1.0 — 최초 초안 (plan-phase). Author: manager-spec. §A 의 A/B 증거는 이 SPEC 을 낳은 세션에서 직접 관측.
