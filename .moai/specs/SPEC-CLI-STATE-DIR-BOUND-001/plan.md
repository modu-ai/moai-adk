# Plan — SPEC-CLI-STATE-DIR-BOUND-001

> 구현 계획. 경로는 워크트리 상대. 순서는 **되돌리기 어려운/바뀔 확률 높은 결정 우선** — 진입점 분리와 chain 데이터 처분이 먼저, 기계적 호출부 정리가 뒤.
>
> v0.3.1 개정: §B 에 네 번째 걷기(`findRegistryUpward`, 범위 밖)와 두-레지스트리 사실 추가, M3 에 REQ-10(해석된 루트 출력) 편입, AP-13b/c/d 신설.
>
> iter3 개정: 진입점 read/destructive 분리(D6 신설), Pre-Flight 의 계수 오류 수정, M2 를 데이터 영향 마일스톤으로 유지하되 경우 4(이전 실패) 추가, 정규화를 양 분기로 확대.

## §A Context

`internal/cli/state.go:231` 의 무한 상향 걷기를 `internal/core/project` 의 보호된 프로젝트-루트 규약으로 수렴시킨다. 관측 증거(spec.md §A 의 A/B 프로브)는 이 세션에서 실측했다. run-phase 모듈: `internal/cli` (주), `internal/core/project` (읽기/위임 대상, 수정 없음).

**핵심 성격**: 기능 추가가 아니라 **규약 수렴**이다. `state.go` 에서는 새 코드보다 지워지는 코드가 많아야 정상이다. 단, REQ-7(chain 레거시 처분)은 이 SPEC 이 유일하게 **추가하는** 로직이다.

**감사 이력**: iter1 FAIL 0.60 → iter2 FAIL 0.667 → iter3(현재). 방향(Option 1)은 두 차례 모두 유지 판정. MUST-FIX 는 전부 이 문서와 형제 문서에 반영되었다.

## §B Known Issues

1. **`CLAUDE_PROJECT_DIR` 을 오늘 읽는 소비자는 6곳 중 하나뿐이다** (`chain.go:64`, `resolveChainStore`). grep 실측: `clean.go` / `tokens.go` / `state.go` 각 0건. 따라서 "공유 헬퍼로 끌어올리기"는 중립적 개선이 아니라 **다섯 곳의 해석 모드를 바꾸는 변경**이다.
2. **그 다섯 중 하나가 삭제한다.** `clean.go:116` 은 `os.RemoveAll`. 이 리포는 `$CLAUDE_PROJECT_DIR` 이 워크트리를 추적하지 못한 사례를 문서로 갖고 있다(`main-checkout-branch-guard.md` § Mechanical Enforcement). 일괄 적용하면 워크트리에서 실행한 `moai clean` 이 primary 아래를 지운다 — **SPEC 의 발단이 된 위험의 재발이다.** REQ-9 가 이 문을 닫는다.
3. **`chain.go` 의 두 분기는 오늘 다른 경로를 만든다.** `ChainStateDir = ".moai/state/chain"`(chain.go:29) 이므로 env 분기는 `<proj>/.moai/state/chain`, walk 분기는 `filepath.Dir("<root>/.moai/state")+"chain"` = `<root>/.moai/chain`. 끌어올리기는 둘 중 하나를 강제로 고르게 하고, 그 선택은 한 부류 사용자의 기존 이벤트 파일을 재배치한다. **정본은 `.moai/state/chain` 으로 확정됐다**(spec.md REQ-7) — 구현자 재량이 아니다.
4. **훅 writer 가 이미 정본 경로에 써 왔다.** `internal/hook/chain_event.go:67` 이 `.moai/state/chain` 을 하드코딩한다. 결과: **두 표면을 다 쓴 머신에는 두 이벤트 파일이 이미 존재한다** — REQ-7 경우 2 는 예외가 아니라 기대 상태다(spec.md R4).
5. **`FindProjectRoot` 는 `EvalSymlinks` 로 정규화한다**(root.go:27-30). `findStateDirFrom` 은 하지 않는다. darwin 에서 `t.TempDir()`(`/var/...`)과 해석 결과(`/private/var/...`)의 문자열이 다르다. **경로 동등성을 단언하는 테스트를 태울 가능성이 가장 높은 항목이며, 양쪽을 정규화해 비교하면 계약을 검증하지 못한다**(acceptance.md §D.2).
6. **마커 불일치.** `FindProjectRoot` 마커는 `.moai`, `findStateDirFrom` 마커는 `.moai/state`. 유효 루트 아래의 맨 `.moai` 하위에서 오늘은 건너뛰고 위에서 성공하지만, 변경 후에는 그 자리에서 실패한다(REQ-3 부류 B). 이 리포에 `internal/hook/.moai/` 형태 이력이 있다.
7. **`loadRegistryForOverlay` 는 fail-open** 이다(chain.go:352-361). REQ-3 이 실패 경우를 늘리므로 통행량이 증가한다. **이 fail-open 을 검증하는 테스트가 오늘 존재하지 않는다** — AC-012 가 신규 작성을 지시한다.
8. **`findStateDirFrom` 은 테스트 전용 seam** 이다(프로덕션 호출자는 `findStateDir` 하나). 홈 경계를 테스트가 주입할 수 있어야 AC-002 가 실행 가능해진다 — §E D3.

9. **패키지 밖에 네 번째 걷기가 살아남는다.** `internal/hook/cwd_changed_relocate.go:78` `findRegistryUpward` 는 `.moai/state/active-sessions.json` 을 찾아 루트까지 **무제한**으로 올라가며 **홈 가드가 없다** — 이 SPEC 이 없애는 것과 같은 결함 형태다. **범위 밖이며 고치지 않는다**(spec.md §B). run-phase 가 알아야 할 것 하나: REQ-6 은 `internal/cli` 로 범위 한정되어 있으므로, "걷기가 하나 남았는지" 확인할 때 `internal/hook` 을 대상으로 삼지 말 것. AC-003 판정 명령 3(`grep -c "for {" internal/cli/state.go`)이 검사하는 범위가 정확히 그 한정이다.
10. **`loadRegistryForOverlay` 가 읽는 레지스트리는 설계상 두 곳 중 하나다.** `internal/session/anchor.go:26-33` 이 측정 사실로 기록한다 — `LiveAnchoredSessions` 는 트리-로컬과 호출자 프로젝트 레지스트리를 **둘 다** 읽고, `moai cc -w` 레인은 후자에 등록된다. `loadRegistryForOverlay` 는 하나만 읽고 REQ-1 은 그 하나를 바꾼다. **범위 확대 금지** — 이 SPEC 은 읽는 레지스트리 **개수**를 바꾸지 않는다. spec.md §E 4행이 이 소비자에 한해 안전성을 "미확인"으로 표기한 이유다.

### 앞선 반복이 남긴 거짓 전제 (제거됨)

- **iter1**: `TestFindStateDirFromWalksUp` 서브테스트 1 이 "반대 계약"이라 반전이 필요하다고 적었다. **틀렸다.** 그 트리는 `root` 에 `.moai/state` 가 있고 시작점이 그 아래이므로 권고안에서도 같은 조상을 반환한다 — 무변경 통과한다. 반전 지시를 따르면 올바른 구현이 만들지 않는 계약을 고정한다(AP-1).
- **iter2**: AC-003 의 기대 계수를 6 이라 적었다. **틀렸다.** `grep "findStateDir()"` 은 정의(`state.go:212`)와 주석(`chain.go:61`)까지 잡아 **오늘 8을 반환한다**. 게다가 정의는 변경 후에도 살아남으므로(그것이 공유 헬퍼다) `state.go` 가 2가 되는 일은 없다. §C Pre-Flight 와 acceptance.md AC-003 을 호출 전용 패턴으로 교체했다.
- **iter2**: §E "동작 변경 3건" 완전성 주장. **불완전했다** — 소비자 다섯의 모드 변경(위 1·2번)이 빠져 있었다. spec.md §E 를 4행으로 확장했다.

## §C Pre-Flight (Run-Phase 진입 확인)

M1 전에 실행. 기대값은 **현재 트리 실측치**이며, 어긋나면 SPEC 의 사실 기반이 낡았다는 뜻이므로 진행 전에 보고한다.

```bash
go test ./internal/cli/... ./internal/core/project/...   # 초록 baseline

# 호출 지점 6곳 (호출 전용 패턴 — 정의/주석 제외)
grep -c "err := findStateDir()" internal/cli/clean.go internal/cli/tokens.go internal/cli/chain.go internal/cli/state.go
#   기대: clean.go:1  tokens.go:1  chain.go:2  state.go:2   (합 6)
#   참고: grep -c "findStateDir()" 는 정의(state.go:212)+주석(chain.go:61)까지 잡아 8을 낸다 — 계수용으로 쓰지 말 것

grep -c "EnvClaudeProjectDir" internal/cli/chain.go       # 기대: 2  (line 64 범위 안 / line 83 범위 밖)
grep -c "EnvClaudeProjectDir" internal/cli/clean.go internal/cli/tokens.go internal/cli/state.go
#   기대: 각 0 — 이 셋은 환경변수를 읽지 않는다 (REQ-1 모드 변경의 사실 근거)

grep -n "ChainStateDir" internal/cli/chain.go             # 기대: 상수 = .moai/state/chain
grep -n "state\", \"chain\"" internal/hook/chain_event.go # 기대: 훅 writer 의 하드코딩 경로
ls -d .moai/chain .moai/state/chain 2>&1                  # 기대: 둘 다 없음 (이 리포엔 위험 데이터 없음)
grep -c "for {" internal/cli/state.go                     # 기대: 1  (그 하나가 문제의 걷기)
```

## §D Constraints (Hard)

- 템플릿 미러 의무 **없음** — `internal/` 아래 Go 코드, `internal/template/templates/` 대응 파일 없음. 찾지 말 것.
- `internal/core/project/root.go` 는 **읽기 전용**. 부득이 수정이 필요하면 설계 재검토 신호이며 progress.md 에 근거를 남긴다.
- `internal/hook/chain_event.go` 는 **범위 밖**. 훅은 이미 정본 경로를 쓰므로 고칠 것이 없다(spec.md §B).
- `resolveCWD`(chain.go:83)의 `EnvClaudeProjectDir` 사용은 **범위 밖 — 손대지 않는다.** AC-007 이 잔여 1건으로 이진 판정한다.
- `clean.go` 에 `CLAUDE_PROJECT_DIR` 참조를 **추가하지 않는다**(REQ-9). AC-015 판정 명령 2 가 0건을 요구한다.
- `loadRegistryForOverlay` 의 fail-open 은 불변.
- 경로 동등성 단언은 **기대값만 정규화**한다(acceptance.md §D.2). 양쪽 정규화는 REQ-8 을 검증하지 못한다.
- 로컬 `go test ./...` 금지. 패키지 단위만.
- 크로스 플랫폼 컴파일 게이트: `GOOS=windows go vet ./internal/cli/...`.

## §E Self-Verification (설계 결정)

### D1 — 해석 계약: 위임인가 복제인가

**선택**: 보호된 걷기는 `project.FindProjectRoot()` 에 **위임**한다. 홈 가드를 복제하지 않는다.

**근거**: 홈 가드는 이미 `root.go:38-52` 에 있고 `paths.Home()` + `EvalSymlinks` 정규화까지 갖췄다. 복제하면 두 구현이 갈라진다 — 지금 고치는 문제가 정확히 그 형태다. 임포트는 이미 존재하므로 비용 0.

**기각안**: `state.go` 안에 홈 가드를 인라인 추가(spec.md §D Option 2). diff 는 작지만 걷기 구현이 둘로 남고 정규화 규약도 갈린 채 남는다.

### D6 — 진입점을 성격별로 둘로 나눈다 (iter3 신설, 가장 중요한 구조 결정)

**선택**: 걷기 구현은 하나로 두되, **진입점을 두 개** 노출한다.

| 진입점 | 동작 | 소비자 |
|---|---|---|
| `findStateDir()` | `CLAUDE_PROJECT_DIR`(정규화) → 보호된 걷기 → `<root>/.moai/state` | 읽기·추가 5곳 |
| `findStateDirNoEnv()` | 보호된 걷기만. 환경변수를 **참조하지 않음** | `clean.go` 1곳 |

**근거**: REQ-9. 환경변수를 일괄 적용하면 `os.RemoveAll` 이 그것을 따라간다. 성격이 다른 두 소비자를 같은 함수로 서비스하면 **분기 조건이 호출자 문서에만 존재**하게 되고, 다음 사람이 `clean` 에서 `findStateDir()` 을 호출해도 아무 신호가 나지 않는다. 이름을 갈라두면 잘못된 선택이 **코드 리뷰에서 보인다**.

**기각안 A**: `findStateDir(honorEnv bool)` 불리언 파라미터. 호출부에서 `false` 의 의미가 보이지 않고, 실수로 `true` 를 넘겨도 컴파일된다.
**기각안 B**: 단일 진입점 + `clean` 쪽에서 환경변수를 미리 `unset`. 프로세스 전역 상태를 건드리는 것이라 다른 코드에 영향이 번진다.

**대응 의무**: 다른 이름을 택하면 progress.md 에 대응표를 남기고 acceptance.md 의 grep 을 치환한다.

**잔여**: 이 분류는 "삭제하는가"라는 성격에 기댄다. 훗날 읽기 계열이 삭제를 하게 되면 조용히 낡는다(spec.md R8). 이름이 성격을 드러내는 것이 유일한 완화다.

### D2 — chain 디렉터리 정본 경로와 레거시 처분

**선택**: 정본은 **`<project-root>/.moai/state/chain`**. `resolveChainStore` 는 env 설정 여부와 무관하게 이 경로를 만들며, 레거시 `<root>/.moai/chain` 은 spec.md REQ-7 의 **4경우** 규칙대로 처분한다(레거시만 → 일회 이전 / 양쪽 → 정본 승리 + 경고 + 레거시 불변 / 정본만 → 무동작·무경고 / **이전 실패 → 레거시 계속 사용 + 경고**).

**근거**: (1) 훅 writer(`chain_event.go:67`)가 이 경로를 하드코딩해 늘 써 왔다 — 상수보다 강한 증거이며, 정본을 `.moai/chain` 으로 고르면 CLI 와 훅이 영구히 갈린다. (2) `ChainStateDir` 상수가 그 값을 명시 선언한다 — walk 분기의 `filepath.Dir(...)+"chain"` 은 산술의 부산물이다. (3) chain 이벤트는 state 다.

**경우 4 의 위치**: head clause 의 **선언된 예외**다(spec.md REQ-7). 이 예외를 코드에 반영할 때 "이전 실패 시 에러로 중단"은 **오답**이다 — 그러면 chain 명령 전체가 죽는다. 반환은 레거시 경로, 경고는 관측 가능하게.

**병합하지 않는 이유**: 순서 있는 JSONL 두 개의 병합은 타임스탬프 신뢰·중복 제거·순서 결정을 요구하며 범위를 넘는다. 경우 2 는 경고로 사용자에게 넘긴다 — 그리고 그 경우가 **드물지 않다**(§B 4번).

### D3 — 홈 경계 주입 seam (M1 에서 확정, 근거를 progress.md 에 기록)

**문제**: AC-002 는 완전 소유 임시 트리 안의 미끼 디렉터리가 `$HOME` 역할을 해야 한다. `project.FindProjectRoot()` 는 `paths.Home()` 을 내부에서 읽으므로 주입 지점이 없다.

| 안 | 내용 | 위험 |
|---|---|---|
| (a) | `internal/core/project` 에 시작점+홈경계를 받는 내부 함수를 노출하고 `FindProjectRoot` 가 그것을 호출 | §D 제약("읽기 전용")과 충돌 — 채택 시 근거 필수 |
| (b) | 테스트에서 `HOME` 환경변수를 제어 | 병렬 테스트 오염(CLAUDE.local.md §13). 비병렬 강제 필요 |
| (c) | `internal/cli` 쪽에 얇은 래퍼를 두고 홈 경계를 파라미터화 | 걷기 로직이 다시 둘로 갈릴 위험 — D1 의 취지와 충돌 |

**결정 시점**: M1. **AC-002 는 이 결정 전까지 실행 불가**이므로 M3 이전으로 일정을 잡을 수 없다.

**판정 기준 (감사 R6)**: "프로덕션과 테스트가 같은 코드 경로를 타는가". `os.Getwd()` 경유로만 해석하면 프로덕션에서는 분기가 숨고 테스트만 깨진다 — **잘못된 수정을 유도하는 최악의 실패 모양**이다.

### D4 — 정규화 계약을 양 분기에 적용한다

**선택**: 반환 경로가 `EvalSymlinks` 정규화 형태임을 계약으로 선언하고 주석에 적는다(REQ-8). **`CLAUDE_PROJECT_DIR` 값도 사용 직전 정규화한다**(REQ-1).

**근거**: 정규화는 `FindProjectRoot` 가 이미 하는 일이고, 벗겨내려면 위임의 이점을 잃는다. 그리고 정규화된 경로가 **더 옳다** — 같은 물리 디렉터리에 하나의 문자열을 준다. 환경변수 분기를 예외로 두면 "무조건 정규화"가 거짓이 되고, 같은 헬퍼가 분기에 따라 다른 형태를 반환하게 된다.

**iter2 정정**: iter2 는 REQ-1 의 "그대로"와 REQ-8 의 "무조건"을 나란히 두어 모순을 남겼다. "그대로" = *걷기 없음*으로 확정한다.

**기각안**: 반환 직전 원래 접두사를 복원. 두 표현을 오가는 코드가 그 자체로 다음 결함이다.

**잔여**: 같은 환경변수를 읽는 다른 11곳은 정규화하지 않는다 — 헬퍼 안팎에서 문자열이 갈린다(spec.md R7). 물리 디렉터리는 같으므로 기능 영향은 없다.

### D5 — 신규 플래그를 만들지 않는다

**선택**: `--state-dir` 신규 플래그를 만들지 않는다. 명시적 지목이 필요하면 `CLAUDE_PROJECT_DIR`(이미 존재) 또는 harness 계열의 `--project-root` 규약(harness.go:101)을 재사용한다.

**근거**: 플래그는 이미 답을 아는 호출자만 돕고 기본값을 고치지 못한다. 관측된 A 케이스는 아무 플래그도 주지 않은 호출이었다. 규약이 셋인 상황에서 네 번째를 추가하는 것은 문제의 재생산.

## §F Milestones (되돌리기 어려운 순)

### M1 — 해석 계약 + 진입점 분리 + 홈 경계 seam (가장 바뀔 확률 높은 결정)

**파일**:
- `internal/cli/state.go:210-250` — 걷기를 `project.FindProjectRoot()` 위임으로 교체하고 진입점 둘을 노출(D6): `findStateDir()`(env 정규화 우선 → 보호된 걷기), `findStateDirNoEnv()`(걷기만). `findStateDirFrom` seam 의 형태는 D3 결정에 따른다.
- (D3-(a) 채택 시에 한해) `internal/core/project/root.go` — 홈 경계 주입 함수 노출 + progress.md 에 §D 제약 예외 근거 기록.

**결정 산출물**: D3 3안 중 선택 + 근거를 progress.md 에 기록. 판정 기준은 "프로덕션과 테스트가 같은 코드 경로를 타는가".

**Exit**: `go build ./internal/cli/...` 초록. 이 시점에 일부 테스트가 빨간 것은 정상(M4 가 정리).

**왜 첫 번째**: 진입점 구조가 가장 바뀔 확률이 높다. 리드/사용자가 다른 모양을 원하면 되돌리는 비용이 가장 싸다.

### M2 — chain 경로 통일 + 레거시 4경우 처분 (데이터 영향 결정)

**파일**:
- `internal/cli/chain.go:62-72` — `resolveChainStore` 를 공유 헬퍼 경유로 바꾸고 두 분기 모두 `<root>/.moai/state/chain` 을 만들게 한다(REQ-7). 인라인 `EnvClaudeProjectDir` 확인 제거. `resolveCWD`(line 83)는 **손대지 않는다**.
- `internal/cli/chain.go` — 레거시 처분 로직 신설, **4경우 전부**(D2). 이전 실패는 에러 전파가 아니라 레거시 경로 반환 + 경고.
- 테스트: `TestChainDirIsCanonicalUnderBothBranches`(AC-008), `TestChainLegacyEventsRelocation`(AC-009, 서브테스트 4개) 신규.

**reproduction-first**: AC-008 은 M2 이전 코드에서 반드시 실패한다(두 분기가 다른 경로를 만드므로). 빨간 출력을 progress.md 에 인용.

**주의**: AC-009 경우 4 를 퍼미션으로 재현하면 Windows 에서 안 걸릴 수 있다. `t.Skip` 대신 **주입 가능한 실패 seam** 으로 구성한다.

**Exit**: `go test ./internal/cli/ -run 'Chain' -v` 초록. AC-007/AC-008/AC-009 판정 가능.

**왜 두 번째**: 사용자 데이터를 옮기는 유일한 마일스톤이다. 되돌리기가 가장 어렵고, 정본 선택이 뒤집히면 통째로 다시 쓰인다.

### M3 — `clean` carve-out (파괴 경로) + 해석된 루트 가시화

**파일**:
- `internal/cli/clean.go:65` — `findStateDir()` → `findStateDirNoEnv()` 로 교체(REQ-9). `clean.go` 에 `CLAUDE_PROJECT_DIR` 참조를 추가하지 않는다.
- `internal/cli/clean.go` — **해석된 프로젝트 루트를 한 줄로 출력**(REQ-10). 위치는 삭제 후보 열거 **이전**이어야 한다 — 후보를 다 보여준 뒤에 루트를 말하면 운영자는 이미 잘못된 목록을 읽은 뒤다.
- 다섯 읽기·추가 소비자에도 같은 한 줄을 추가한다(REQ-10 은 여섯 전부를 구속한다). `clean` 만 AC 로 이진 판정하는 것은 **최소 요구**이지 면제가 아니다.
- 테스트: `TestCleanIgnoresProjectDirEnv`(AC-015 판정 1-3), `TestCleanAnnouncesResolvedRoot`(AC-015 판정 4) 신규.

**REQ-10 이 이 마일스톤에 있는 이유**: 이 줄은 REQ-1/REQ-9 분할이 만든 **동시 분기**(같은 세션에서 읽기는 B, `clean` 은 A)에 대한 유일한 완화다(spec.md §E "선언된 부수 결과"). 분할과 완화를 같은 마일스톤에 두어, 분할만 착지하고 완화가 뒤로 밀리는 상태를 만들지 않는다.
- `clean.go:136` 은 **변경하지 않는다** — `filepath.Dir(stateDir)` 산술은 REQ-2 하에서도 `<root>/.moai` 를 정확히 준다(spec.md §E 제약).

**Exit**: `go test ./internal/cli/ -run 'Clean' -v` 초록. AC-015 **네** 판정 명령 통과.

**왜 세 번째**: M1 의 진입점이 존재해야 교체할 수 있고, 삭제 경로이므로 chain 이전보다 뒤에 두어 한 번에 하나의 위험만 다룬다.

### M4 — 테스트 계약 정리 (기존 보존 + 신규 회귀)

**파일**:
- `internal/cli/tokens_state_dir_test.go` — 서브테스트 1 **보존**(단언 의미 무변경). 서브테스트 2 보존하되 경로 비교를 §D.2 규약으로. 서브테스트 3 을 결정적 `err != nil` 단언으로 강화. `t.Skip` 추가 금지. 개명은 SHOULD.
- 공통 헬퍼 `normPath` 도입 + **기대값만 정규화**하는 비교 규약 준수.
- 신규 테스트: `TestStateDirHonoursProjectDirEnv`(AC-001), `TestStateDirDoesNotCrossHomeBoundary`(AC-002), `TestStateDirStopsAtProjectRoot`(AC-004), `TestStateDirStopsAtNestedBareMoai`(AC-011), `TestStateDirReturnsNormalizedPath`(AC-010, 서브테스트 2개), `TestResolveTokensStateDirFallsBackToCwd`(AC-006), `TestLoadRegistryForOverlayFailsOpen`(AC-012).

**reproduction-first**: AC-002 와 AC-011 은 M1 이전 코드에서 실패해야 한다. 빨간 출력 인용.

**Exit**: `go test ./internal/cli/ -run 'StateDir|Tokens|Overlay|WalksUp' -v` 초록. `--- SKIP` 0건, `no tests to run` 0건.

### M5 — 나머지 소비자 확인

**파일**:
- `internal/cli/tokens.go:376-385` — `<cwd>/.moai/state` 폴백 보존 확인(REQ-4). 순서는 헬퍼 → 폴백 그대로.
- `internal/cli/state.go:78`, `state.go:154`, `chain.go:353` — 헬퍼 시그니처 유지 시 대부분 무변경. `chain.go:353` 의 fail-open 유지 확인.
- `internal/cli/state_m2_test.go` — **수정 예정 없음.** `m2SetupState`(line 37)는 경로 문자열을 비교하지 않고 해석된 디렉터리를 통해 동작하므로 정규화 변경에 무영향이다. `go test ./internal/cli/ -run 'M2' -v` 로 그 주장을 관측 확인만 한다.

**Exit**: `go test ./internal/cli/...` 초록. AC-003/AC-006 판정 가능.

### M6 — 검증 스윕

```bash
go test ./internal/cli/... ./internal/core/project/...
go test -race ./internal/cli/...
GOOS=windows go vet ./internal/cli/...
golangci-lint run --timeout=2m
```

`go test ./...` 는 **로컬에서 돌리지 않는다** — push 후 CI 판정. AC-013 판정.

### M7 — 문서 반영

**파일**:
- `internal/cli/state.go` 함수 주석 — "unbounded", "inherits any `~/.moai/state`" 를 새 동작으로 교체하고 **정규화 계약을 명시**(REQ-8 SHALL). 두 진입점의 성격 차이도 주석에 남긴다.
- `internal/cli/tokens.go:372-375` 주석 — "existing walk-up convention (findStateDir, state.go)" 갱신.
- `internal/cli/chain.go:61` 주석 — "Prefers CLAUDE_PROJECT_DIR; falls back to findStateDir() directory walk." 는 끌어올리기 이후 **사실과 어긋난다.** 새 동작 + 정본 경로 + 레거시 처분으로 교체.

**Exit**: AC-010 명령 2, AC-014 세 명령 모두 기대값 일치.

## §G Anti-Patterns (피할 것)

- **AP-1**: `TestFindStateDirFromWalksUp` 서브테스트 1 을 **반전시키는 것**. iter1 의 지시였고 틀렸다(§B). 보존이 정답이다.
- **AP-2**: 서브테스트를 `t.Skip` 으로 무력화하는 것. AC-005 가 0건을 이진 판정한다. AC-009 경우 4 의 플랫폼 차이도 스킵이 아니라 주입 seam 으로 해결한다.
- **AP-3**: 홈 가드를 `state.go` 에 복제하는 것. 지금 고치는 문제의 재생산.
- **AP-4**: **`clean` 에 `CLAUDE_PROJECT_DIR` 우선순위를 주는 것.** `os.RemoveAll` 이 환경변수를 따라가면 SPEC 의 발단이 된 위험이 되돌아온다(REQ-9). AC-015 가 실패한다.
- **AP-5**: 진입점을 하나로 두고 불리언 파라미터로 가르는 것. 호출부에서 의미가 보이지 않는다(D6 기각안 A).
- **AP-6**: chain 정본을 `.moai/chain` 으로 고르는 것. 훅 writer 와 영구히 갈린다(D2). AC-008 이 실패한다.
- **AP-7**: 레거시 chain 이벤트를 **조용히** 버리는 것. REQ-7 은 이전 또는 경고를 SHALL 로 요구한다.
- **AP-8**: **이전 실패 시 에러를 전파해 chain 명령을 죽이는 것.** 경우 4 의 정의된 동작은 레거시 경로 반환 + 경고다(D2).
- **AP-9**: 두 chain 이벤트 파일을 병합하려 시도하는 것. 범위 밖(D2).
- **AP-10**: 경로를 **양쪽 정규화**해 비교하는 테스트. 어느 구현에서나 통과해 REQ-8 을 검증하지 못한다(acceptance.md §D.2). 기대값만 정규화한다.
- **AP-11**: 정규화를 되돌려 원래 접두사를 복원하는 것. 두 표현을 오가는 코드가 다음 결함이다(D4 기각안).
- **AP-12**: `resolveCWD`(chain.go:83)를 건드리는 것. 범위 밖이며 AC-007 이 잔여 1건을 요구한다.
- **AP-13**: `internal/hook/chain_event.go` 를 수정하는 것. 훅은 이미 정본 경로를 쓴다 — 범위 밖(spec.md §B).
- **AP-13b**: `internal/hook/cwd_changed_relocate.go` 의 `findRegistryUpward` 를 "겸사겸사" 고치는 것. 같은 결함 형태지만 **범위 밖**이며 후속 카드 소관이다(spec.md §B). 한 카드에 두 패키지의 해석 규약을 함께 바꾸면 실패 시 원인 분리가 안 된다. REQ-6 의 `internal/cli` 한정을 리포 전체 보증으로 확대 해석하지 말 것.
- **AP-13c**: `loadRegistryForOverlay` 가 읽는 레지스트리를 **둘로 늘리는** 것. `anchor.go` 의 두-레지스트리 설계는 이 SPEC 의 범위 밖이며, 개수를 바꾸면 §E 4행의 영향 분석이 무효가 된다(§B 10번).
- **AP-13d**: REQ-10 의 한 줄을 `clean` 에만 넣고 나머지 다섯을 빠뜨리는 것. AC 가 `clean` 만 이진 판정하는 것은 **최소 요구**이지 면제가 아니다(REQ-10 은 여섯 전부를 구속한다).
- **AP-14**: 깊이 N 상한 도입, 또는 신규 `--state-dir` 플래그. spec.md §D Option 3/4 기각.
- **AP-15**: `loadRegistryForOverlay` 의 fail-open 을 에러 전파로 바꾸는 것. 범위 밖이고 회귀다.
- **AP-16**: 실제로 존재하지 않는 테스트 이름을 `-run` 패턴으로 검증하는 것. `no tests to run` 은 exit 0 이라 **아무것도 검증하지 않고 PASS 로 읽힌다**(iter1 AC-008 의 실패 형태).
- **AP-17**: `grep "findStateDir()"` 로 호출 수를 세는 것. 정의와 주석까지 잡아 8을 낸다(§B, iter2 의 오류). 호출 전용 패턴을 쓴다.
- **AP-18**: 옛 주석 3곳(`state.go`, `tokens.go`, `chain.go:61`)을 남기는 것.
- **AP-19**: Windows CI 진단(`dir %USERPROFILE%\.moai`)을 이 SPEC 에 접는 것. 후속 카드 소관.
- **AP-20**: `internal/template/templates/` 에서 미러를 찾는 것. 없다(§D).

## §H Cross-References

- SPEC: `.moai/specs/SPEC-CLI-STATE-DIR-BOUND-001/spec.md`
- Acceptance: `.moai/specs/SPEC-CLI-STATE-DIR-BOUND-001/acceptance.md`
- 감사 보고서: `.moai/reports/t164/plan-audit.md` (iter1 FAIL 0.60), `.moai/reports/t164/plan-audit-iter2.md` (iter2 FAIL 0.667)
- `internal/cli/state.go:210-250` — 수정 대상.
- `internal/core/project/root.go:20-60` — 위임 대상(읽기 전용), `EvalSymlinks` 27-30행.
- `internal/cli/chain.go:29` — `ChainStateDir` 상수 (D2 정본 근거 2).
- `internal/hook/chain_event.go:55-68` — 훅 writer (D2 정본 근거 1, 범위 밖).
- `internal/cli/clean.go:65` / `:116` / `:136` — REQ-9 대상 / `os.RemoveAll` / 안전 확인된 동일 산술.
- `internal/cli/harness.go:101` — 플래그 규약 선례.
- `internal/cli/tokens_state_dir_test.go:32` — 서브테스트 3개.
- `internal/cli/state_m2_test.go:37` — `m2SetupState`, 무영향 확인 대상.
- `internal/config/envkeys.go:290` — `EnvClaudeProjectDir`; 프로덕션 독자 12곳.
- `internal/hook/cwd_changed_relocate.go:78` — `findRegistryUpward`, 네 번째 걷기 (범위 밖 — AP-13b).
- `internal/session/anchor.go:26-33` — 레지스트리 두 곳 측정 사실 (범위 밖 — AP-13c).
- `.claude/rules/moai/workflow/main-checkout-branch-guard.md` § Mechanical Enforcement — REQ-9 의 선례 근거.
- CLAUDE.local.md §4 (로컬 전체 스위트 금지), §6 (테스트 격리 `t.TempDir()`), §13 (`t.Setenv("HOME")` 병렬 오염 주의).
