# Progress — SPEC-CLI-STATE-DIR-BOUND-001

## §E.1 Plan-phase Audit-Ready Signal

### iter1 — 2026-08-22

- plan-phase 산출물 3종 작성. Tier M. 권고안: Option 1(보호된 프로젝트-루트 규약으로 위임).
- **plan-auditor 판정: FAIL 0.60** (임계 0.80). 보고서: `.moai/reports/t164/plan-audit.md`. MUST-FIX 8건(D1-D8). 방향은 유지 판정.

### iter2 — 2026-08-22

- D1-D8 반영 (v0.2.0).
- **plan-auditor 판정: FAIL 0.667** (임계 0.80). 보고서: `.moai/reports/t164/plan-audit-iter2.md`. MUST-FIX 5건(E1/E4/E5/E6/E7). iter1 MUST-FIX 8건 중 6건 완전 종료, D5 는 형태를 바꾼 새 오류로 잔존.
- 점수 추이 0.60 → 0.667 (개선, STOP 신호 없음).

### iter3 — 2026-08-22 (현재, v0.3.0)

운영자 결정에 따라 iter3 개정. MUST-FIX 5건 처리 내역:

| 감사 항목 | 처리 |
|---|---|
| **E6** — 네 번째 미선언 동작 변경 (5곳이 env 우선순위를 새로 얻고 그중 하나가 삭제) | REQ-1 을 읽기/파괴 축으로 **분할**. 읽기·추가 5곳은 env 를 따르고, 파괴적 `clean` 은 **REQ-9 신설**로 제외. REQ-6 의 "6곳 전부 동일" 전제를 "하나의 걷기 · 읽기 계열 하나의 해석 · 파괴적 소비자는 대상 명시"로 재작성하고 전제 재개 사실을 본문에 기록. §E 표를 4행으로 확장하고 6곳 배분(신규 4 + 기존 1 + 제외 1)을 재계산. **AC-015 신설**(3개 판정 명령). plan.md **§E D6 신설**(진입점 2개 분리: `findStateDir` / `findStateDirNoEnv`), **§F M3 신설**(clean carve-out 마일스톤), AP-4/AP-5 추가 |
| **E4** — REQ-1 "그대로" vs REQ-8 "무조건 정규화" 모순 | REQ-1 의 "그대로"를 *걷기 없음*으로 확정하고 env 값의 `EvalSymlinks` 정규화를 SHALL 로 명시 — REQ-8 이 양 분기에 성립. **AC-001 재작성**(실제값 raw · 기대값만 정규화 → 구현을 실제로 구분), **AC-010 을 분기 A/B 로 분할**(환경변수 분기를 명시적으로 태움). acceptance.md §D.2 에 "양쪽 정규화 금지" 를 [HARD] 로 명문화, plan.md D4 + AP-10. R7 신규 등록(다른 11곳과 정규화 여부가 갈림) |
| **E5** — `internal/hook/chain_event.go:67` 미명시 | §A 표 아래 "패키지 밖 writer 1곳"으로 명시, §F 크로스레퍼런스 추가, **REQ-7 정본 근거 1순위**로 인용(상수보다 강한 증거). **R4 재서술** — 경우 2("양쪽 존재")는 희귀 경로가 아니라 두 표면을 다 쓴 머신의 **기대 상태**. 훅 리터럴 리팩터는 §B 에 이유 2개와 함께 명시적 범위 밖 |
| **E7** — REQ-7 head clause 자기모순 + 이전 실패 AC 부재 | head clause 에 "경우 4 는 명시적 예외" 를 명문화하고 처분을 **4경우 표**로 재작성. **AC-009 에 경우 4 추가**(이전 실패 → 레거시 반환 + 경고 + 무손실). **AC-008 의 부정 단언을 성공 경로로 한정**(Given 이 이전 개입 상황을 배제). plan.md D2 에 "에러 전파는 오답" 명시 + AP-8 |
| **E1** — AC-003 기대 계수 오류 (8 vs 6) | `grep "findStateDir()"` 가 정의(`state.go:212`)와 주석(`chain.go:61`)까지 잡는다는 사실을 §B 에 기록. **AC-003 을 호출 전용 패턴**(`err := findStateDir()`)으로 교체하고 변경 전/후 파일별 표를 병기(6 → 5 + clean 전용 1). **plan.md §C Pre-Flight 재작성**(기대값을 현재 실측치로, 오용 주의 주석 포함). `chain.go:61` 주석을 **AC-014 3번째 명령**으로 추가. AP-17 신설 |
| E8 (optional) | `clean.go:136` 의 동일 산술을 검토하고 안전함을 spec.md §E 제약에 기록 (`filepath.Dir` 는 REQ-2 하에서도 `<root>/.moai` 를 정확히 준다) |
| E9 (optional) | REQ-7 에서 함수 이름(`resolveChainStore`) 제거 — 요구사항 계층은 동작만 규정 |
| E10 (optional) | E1 수정에 포함 (plan.md Pre-Flight) |
| E11 (optional) | AC-002 의 스케줄 의존성(M1 D3 결정 전 실행 불가) 유지 |

**미변경(의도적)**: 권고안 Option 1. 두 감사 모두 방향 유지 판정.

**예산 확인**: REQ 9개 / AC 15개 — Tier M 상한(각 16) 이내.

### iter3 감사 결과 + v0.3.1 정정 패스 — 2026-08-22

- **plan-auditor 판정: PASS 0.80** (Tier M 임계 0.80 — **동률, 초과 아님**). 보고서: `.moai/reports/t164/plan-audit-iter3.md`. iter2 MUST-FIX 5건 전부 종료 확인. 점수 추이 0.60 → 0.667 → 0.80.
- 잔여 부채 F1-F5 중 **셋을 접었다** (v0.3.1). 설계는 무변경 — REQ-1 분할 · REQ-7 · REQ-9 · §E 4행 그대로.

| 부채 | 처리 |
|---|---|
| **F1** — REQ-6 "걷기 정확히 하나"가 착지 시점에 **거짓** (`internal/hook/cwd_changed_relocate.go:78` `findRegistryUpward` 가 같은 마커를 홈 가드 없이 루트까지 걷는다) | REQ-6 을 `internal/cli` 로 범위 한정 + 왜 한정이 장식이 아닌지 [HARD] 명시. **R2 재계수** 셋 → **적어도 넷**, 네 번째를 이름으로 지목(전수 조사 안 했으므로 "적어도"). **Out of Scope 항목 신설** + 후속 카드 후보 등재. §F 3개 앵커. plan.md §B 9번 + **AP-13b** |
| **F3** — 한 세션 안에서 `clean` 과 나머지 다섯이 다른 프로젝트를 해석 (읽기는 B, 삭제는 A, 신호 없음) | **REQ-10 신설**(행동 전 해석된 루트 한 줄 출력, 최소 `clean`). §E **"선언된 부수 결과" 절 신설** — 수용 결과로 선언 + 되돌리지 않는 이유(REQ-9 회귀) + R8 과 다른 항목임을 명시. **AC-015 판정 명령 4** + `TestCleanAnnouncesResolvedRoot`. plan.md **M3 확장** + AP-13d |
| **AC-009 경고** — 경우 2·4 가 스트림도 문자열도 지목하지 않아 grep 불가 (REQ-7 의 가장 위험한 두 분기) | **stderr 고정 + 안정 부분문자열 2개 고정**(`"chain: legacy events retained at"` / `"chain: migration failed, using legacy path"`) + 경로 동반 의무. Closure Gate 항목 추가 |
| **F2 (부분)** — §E 4행의 안전성 주장이 두-레지스트리 설계를 모른 채 단언됨 | 4행 안전성을 **일반 논거로 유지하되 `loadRegistryForOverlay` 한정 "미확인"** 표기 + `anchor.go:26-33` 인용. 확대 안 함. plan.md §B 10번 + **AP-13c**(레지스트리 개수 변경 금지) |
| **F4 / F5** | 접지 않았다. iter3 보고서에 기록된 부채로 남긴다(spec.md §F 에 그 사실 명시) |

**예산 확인 (v0.3.1)**: REQ 10개 / AC 15개 — Tier M 상한(각 16) 이내. REQ-10 은 AC 를 새로 만들지 않고 AC-015 판정 4로 접어 AC 여유 1칸을 남겼다.

### run-phase 진입 전 남은 결정 1건

- **D3 — 홈 경계 주입 seam** (plan.md §E D3 의 3안). M1 에서 결정하고 근거를 §E.2 에 기록. 판정 기준: "프로덕션과 테스트가 같은 코드 경로를 타는가"(감사 R6).

### run-phase 가 남겨야 할 증거

- **reproduction-first 3건**: AC-002 / AC-008 / AC-011 은 구현 전 트리에서 실패해야 한다. 세 빨간 출력을 §E.2 에 인용.
- **진입점 이름 대응표**: plan.md §E D6 의 이름(`findStateDir` / `findStateDirNoEnv`)과 다른 이름을 택했다면 대응표를 §E.2 에 기록하고 acceptance.md 의 grep 을 치환해 실행.

## §E.2 Run-phase Evidence

### D3 결정 — 홈 경계 주입 seam (M1)

**선택**: plan.md §E D3 의 (a) 와 (b) 를 합친 형태. `internal/core/project` 에 **시작 디렉터리만** 받는 `FindProjectRootFrom(start string)` 을 노출하고, **홈 경계는 주입하지 않는다** — 양쪽 모두 `paths.Home()` 으로 해석하며, 테스트는 `t.Setenv("HOME", …)` 로 그 해석의 입력을 제어한다.

**판정 기준("프로덕션과 테스트가 같은 코드 경로를 타는가") 적용**:

- `FindProjectRoot()` 는 `os.Getwd()` 뒤에 `FindProjectRootFrom(dir)` 을 호출하는 얇은 래퍼가 됐다. 정규화·홈 가드·상향 순회 **전부가 한 함수 안**에 있고, 프로덕션과 테스트가 그 함수를 그대로 탄다. 테스트 전용 분기가 없다.
- 홈 경계를 파라미터로 받지 **않은** 이유가 이 기준이다. 파라미터로 받으면 프로덕션은 `paths.Home()` 을, 테스트는 리터럴을 넘기게 되어 **해석 자체가 두 개**가 된다 — R6 이 경고한 "테스트만 빨간" 실패 모양의 씨앗이다. `HOME` 을 제어하면 테스트도 `paths.Home()` 을 통과한다.
- `t.Setenv` 는 병렬 테스트에서 패닉하므로 비병렬을 **구조적으로 강제한다**(CLAUDE.local.md §13 의 오염 위험이 여기서 닫힌다). 해당 테스트 어디에도 `t.Parallel()` 이 없다.

**§D 제약 예외 근거**: `internal/core/project/root.go` 를 수정했다(§D "읽기 전용"). 수정 성격은 **동작 변경이 아니라 시작점 파라미터 추출**이며, 기존 본문은 한 줄도 바뀌지 않고 새 함수로 이동했을 뿐이다. 대안 (c)(cli 쪽 래퍼가 홈 경계를 파라미터화)는 걷기를 둘로 갈라 D1·AP-3 과 정면 충돌한다.

**진입점 이름**: acceptance.md §D.2 의 규약을 그대로 채택했다 — `findStateDir()` / `findStateDirNoEnv()`. 대응표 불필요.

### reproduction-first 증거 (구현 전 트리, HEAD d4edbbc70)

명령: `go test ./internal/cli/ -run 'TestStateDir|TestResolveTokensStateDirFallsBackToCwd|TestLoadRegistryForOverlayFailsOpen' -v`

```
=== RUN   TestStateDirDoesNotCrossHomeBoundary
    state_dir_bound_test.go:45: resolution succeeded at "/var/folders/kt/.../001/home/.moai/state"; it must stop at the home boundary
--- FAIL: TestStateDirDoesNotCrossHomeBoundary (0.00s)
=== RUN   TestStateDirStopsAtProjectRoot
    state_dir_bound_test.go:66: got "/var/folders/kt/.../001/.moai/state"; resolution must fail at the project root that has no state dir
--- FAIL: TestStateDirStopsAtProjectRoot (0.00s)
=== RUN   TestStateDirStopsAtNestedBareMoai
    state_dir_bound_test.go:89: got "/var/folders/kt/.../001/.moai/state"; the bare .moai in the subdirectory must stop the resolution
--- FAIL: TestStateDirStopsAtNestedBareMoai (0.00s)
=== RUN   TestStateDirHonoursProjectDirEnv
    state_dir_bound_test.go:115: got "/var/folders/kt/.../001/.moai/state", want the named project's state dir "/private/var/folders/kt/.../001/explicit/.moai/state"
--- FAIL: TestStateDirHonoursProjectDirEnv (0.00s)
=== RUN   TestStateDirReturnsNormalizedPath/walk_branch
    state_dir_bound_test.go:143: got "/var/folders/.../001/.moai/state", which is not its own normalized form "/private/var/folders/.../001/.moai/state"
=== RUN   TestStateDirReturnsNormalizedPath/env_branch
    state_dir_bound_test.go:158: findStateDir: .moai/state/ directory not found from /var/folders/.../001/sub
--- FAIL: TestStateDirReturnsNormalizedPath (0.00s)
=== RUN   TestResolveTokensStateDirFallsBackToCwd
--- PASS: TestResolveTokensStateDirFallsBackToCwd (0.00s)
=== RUN   TestLoadRegistryForOverlayFailsOpen
--- PASS: TestLoadRegistryForOverlayFailsOpen (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.964s
```

AC-002 / AC-004 / AC-011 의 요구된 빨간 출력 3건이 위에 있다. AC-001 과 AC-010 양 분기도 함께 빨갛다. AC-006(폴백)과 AC-012(fail-open)는 **불변식 보존**을 확인하는 테스트이므로 구현 전에도 초록인 것이 정상이다 — 이 둘은 reproduction-first 대상이 아니다.

AC-008 의 빨간 출력은 아래 M2 절에 별도로 인용한다.

### M2 — chain 정본 경로 + 레거시 처분

**reproduction-first (AC-008)**: `resolveChainDir` 을 **동작 보존 추출**만 한 상태(두 분기가 오늘처럼 서로 다른 경로를 만드는 상태)에서 테스트를 먼저 돌렸다. 이 순서가 아니면 "테스트가 결함을 잡는다"를 관측할 수 없다.

명령: `go test ./internal/cli/ -run 'TestChainDirIsCanonicalUnderBothBranches' -v`

```
=== RUN   TestChainDirIsCanonicalUnderBothBranches
    chain_state_dir_test.go:57: branches disagree: env ".../001/project/.moai/state/chain" vs discovery ".../001/project/.moai/chain"
    chain_state_dir_test.go:61: got "/var/folders/.../001/project/.moai/state/chain", want the canonical "/private/var/folders/.../001/project/.moai/state/chain"
    chain_state_dir_test.go:61: got ".../001/project/.moai/chain", want the canonical ".../001/project/.moai/state/chain"
    chain_state_dir_test.go:64: got the legacy location ".../001/project/.moai/chain"
--- FAIL: TestChainDirIsCanonicalUnderBothBranches (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.846s
```

세 결함이 한 번에 드러난다 — 두 분기의 경로 불일치, 환경변수 분기의 미정규화, 그리고 walk 분기의 레거시 위치.

**경우 4(이전 실패) 재현 방식**: 퍼미션이 아니라 **주입**이다. `migrateChainEvents` 를 패키지 변수로 두고 테스트가 실패를 강제한다. 퍼미션으로 만들면 Windows 에서 재현되지 않고, 그 플랫폼에서 REQ-7 의 가장 위험한 분기가 검증 없이 배포된다 — `t.Skip` 은 AC-005 의 스킵 금지 정신과도 어긋난다.

**정본 경로의 출처**: `filepath.Join(projectRootOf(stateDir), ChainStateDir)` 로 만든다. 선언된 상수(`ChainStateDir = ".moai/state/chain"`)를 다시 하중을 받는 자리에 두기 위해서다 — 레거시 경로는 누가 선언한 값이 아니라 `filepath.Dir(stateDir)+"chain"` 산술의 부산물이었다(spec.md REQ-7 근거 2).

### M3 — clean carve-out + REQ-10 가시화

REQ-10 의 한 줄은 여섯 소비자 **전부**에 있다(AP-13d). `clean` 만 AC 로 이진 판정되지만 그것은 최소 요구다.

| 소비자 | 위치 | 출력 경로 |
|---|---|---|
| `clean.go` `runClean` | 삭제 후보 열거 **이전** | `printResolvedRoot` → `Printer.Info` → stderr |
| `state.go` `runStateDump` | store 생성 이전 | 같음 |
| `state.go` `runShowBlocker` | blocker 검색 이전 | 같음 |
| `chain.go` `resolveChainDirWith` | 레거시 처분 이전 | `announceResolvedRoot` → stderr |
| `chain.go` `loadRegistryForOverlay` | 레지스트리 읽기 이전 | 같음 |
| `tokens.go` `resolveTokensStateDir` | 반환 이전 (폴백 분기 포함) | 같음 |

**선언된 부수 결과 — 기존 테스트 2건의 단언 수정**: `TestStateM2_JSON_StdoutByteIdentical` / `TestStateM2_Human_StdoutByteIdentical` 은 `stderr.Len() != 0` 으로 **stderr 완전 공백**을 단언하고 있었다. REQ-10 의 한 줄이 그 단언과 정면으로 충돌한다.

- **stdout 바이트 동일 단언은 한 글자도 건드리지 않았다** — M2 SPEC 의 하중은 그쪽이다.
- stderr 단언은 그 테스트가 지키려던 성질(**데이터 페이로드가 stderr 로 새지 않는다**)로 좁혔다: `assertStderrCarriesOnlyResolvedRoot` 가 해석-루트 줄 외의 어떤 줄도 허용하지 않는다. 완전 공백보다 약하지만, 채널 분리라는 원래 계약은 그대로다.
- 이것은 AP-1 이 경고한 "올바른 구현이 만들지 않는 계약을 테스트에 고정하는 것"의 반대 방향이다 — 새 요구사항이 옛 단언을 무효화한 경우이며, 조용히 넘기지 않고 여기 기록한다.

**AC-015 판정 2 를 위한 주석 조정**: `runClean` 의 주석에 `CLAUDE_PROJECT_DIR` 리터럴을 썼더니 `grep -c "EnvClaudeProjectDir\|CLAUDE_PROJECT_DIR" internal/cli/clean.go` 가 3을 반환했다(기대 0). 그 grep 은 "환경변수가 clean 의 해석 경로에 아예 없다"를 이진 판정하는 장치이므로, 주석이 그것을 무디게 만들면 안 된다. 변수 이름은 `findStateDirNoEnv` 의 주석이 갖고 있고, `clean.go` 는 그리로 가리킨다.

### §D.4 삭제량 확인 — 임계 초과, 검토 결과 복제 아님

`git diff --numstat -- internal/cli/state.go` → **+76 / −25 (순증 +51)**. §D.4 의 임계(순증 20줄)를 넘으므로 D1(위임 vs 복제) 재검토를 실행했다.

**판정: 복제가 아니다.** 임계의 목적은 "홈 가드를 복제했는가"를 잡는 것이고, 그 질문은 순증 줄 수가 아니라 코드로 답한다:

- `grep -c "for {" internal/cli/state.go` → **0**. 상향 순회가 없다.
- `state.go` 에 `paths.Home` 참조가 **0건**이다. 홈 경계 판정 로직이 복제되지 않았다.
- `EvalSymlinks` 2건은 주석 1 + `normalizeDir` 1 이며, 걷기와 무관한 REQ-8 계약이다.

순증 +51 의 귀속: 주석 +39(옛 걷기 서술을 새 계약 서술로 교체 — AC-014 가 요구), REQ-10 헬퍼 3개와 호출 2곳 약 +14, `normalizeDir` +6, 진입점 분리(`findStateDirNoEnv`) +8. 걷기 본체는 **순감**했다. 즉 증가분은 전부 이 SPEC 이 **새로 요구한 것**(REQ-8·REQ-9·REQ-10)이며 위임 대상의 복제가 아니다.

### AC PASS/FAIL 매트릭스

모든 판정 명령은 `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && …` 로 환경을 걷어낸 뒤 이 트리에서 실행했다.

| AC | REQ | 상태 | 판정 명령 | 관측된 출력 |
|---|---|---|---|---|
| AC-001 | REQ-1 | **PASS** | `go test ./internal/cli/ -run 'TestStateDirHonoursProjectDirEnv' -v` | `--- PASS: TestStateDirHonoursProjectDirEnv (0.00s)` — SKIP 0건, `no tests to run` 없음 |
| AC-002 | REQ-2 | **PASS** | `go test ./internal/cli/ -run 'TestStateDirDoesNotCrossHomeBoundary' -v` | `--- PASS: TestStateDirDoesNotCrossHomeBoundary (0.00s)` |
| AC-003 | REQ-6 | **PASS** | ① `grep -c "err := findStateDir()" internal/cli/{clean,tokens,chain,state}.go` ② `grep -c "err := findStateDirNoEnv()" internal/cli/clean.go` ③ `grep -c "for {" internal/cli/state.go` | ① `clean.go:0 tokens.go:1 chain.go:2 state.go:2` (합 5, 기대 5) ② `1` ③ `0` |
| AC-004 | REQ-3 (부류 A) | **PASS** | `go test ./internal/cli/ -run 'TestStateDirStopsAtProjectRoot' -v` | `--- PASS: TestStateDirStopsAtProjectRoot (0.00s)` |
| AC-005 | REQ-5 | **PASS** | ① `go test ./internal/cli/ -run 'TestFindStateDirFromWalksUp' -v` ② `grep -c "t.Skip" internal/cli/tokens_state_dir_test.go` ③ `grep -c "err == nil && strings.HasPrefix" …` | ① 세 서브테스트 모두 `--- PASS`, `--- SKIP` 0건 ② `0` ③ `0` |
| AC-006 | REQ-4 | **PASS** | `go test ./internal/cli/ -run 'TestResolveTokensStateDirFallsBackToCwd' -v` | `--- PASS: TestResolveTokensStateDirFallsBackToCwd (0.00s)` |
| AC-007 | REQ-6 | **PASS** | ① `grep -c "EnvClaudeProjectDir" internal/cli/chain.go` ② `awk '/^func resolveChainStore/,/^}/' … \| grep -c` ③ `awk '/^func resolveCWD/,/^}/' … \| grep -c` | ① `1` ② `0` ③ `1` — 범위 안은 비었고 범위 밖(`resolveCWD`)은 불변 |
| AC-008 | REQ-7 | **PASS** | `go test ./internal/cli/ -run 'TestChainDirIsCanonicalUnderBothBranches' -v` | `--- PASS: TestChainDirIsCanonicalUnderBothBranches (0.00s)` (구현 전 FAIL 출력은 위 M2 절) |
| AC-009 | REQ-7 | **PASS** | `go test ./internal/cli/ -run 'TestChainLegacyEventsRelocation' -v` | 부모 `--- PASS` + 네 서브테스트 각각 `--- PASS`: `legacy_only_is_relocated_once` / `both_present_keeps_the_legacy_file_and_warns` / `canonical_only_is_silent` / `failed_relocation_keeps_using_the_legacy_path` |
| AC-010 | REQ-8 | **PASS** | ① `go test ./internal/cli/ -run 'TestStateDirReturnsNormalizedPath' -v` ② `grep -c "EvalSymlinks" internal/cli/state.go` | ① 부모 + `walk_branch` + `env_branch` 모두 `--- PASS` ② `2` (≥1) |
| AC-011 | REQ-3 (부류 B) | **PASS** | `go test ./internal/cli/ -run 'TestStateDirStopsAtNestedBareMoai' -v` | `--- PASS: TestStateDirStopsAtNestedBareMoai (0.00s)` |
| AC-012 | 횡단 (fail-open) | **PASS** | `go test ./internal/cli/ -run 'TestLoadRegistryForOverlayFailsOpen' -v` | `--- PASS: TestLoadRegistryForOverlayFailsOpen (0.00s)` — 이름이 실재하므로 `no tests to run` 침묵 통과(AP-16)가 아니다 |
| AC-013 | 횡단 (플랫폼) | **PASS** | `GOOS=windows go vet ./internal/cli/...` | 출력 없음, `vet exit: 0` |
| AC-014 | 횡단 (문서 정합) | **PASS** | ① `grep -c "unbounded\|inherits any" internal/cli/state.go` ② `grep -c "existing walk-up convention" internal/cli/tokens.go` ③ `grep -c "falls back to findStateDir() directory walk" internal/cli/chain.go` | `0` / `0` / `0` |
| AC-015 | REQ-9 + REQ-10 | **PASS** | ① `go test ./internal/cli/ -run 'TestCleanIgnoresProjectDirEnv' -v` ② `grep -c "EnvClaudeProjectDir\|CLAUDE_PROJECT_DIR" internal/cli/clean.go` ③ `grep -c "err := findStateDir()" internal/cli/clean.go` ④ `go test ./internal/cli/ -run 'TestCleanAnnouncesResolvedRoot' -v` | ① `--- PASS` ② `0` ③ `0` ④ `--- PASS: TestCleanAnnouncesResolvedRoot (0.00s)` |

15/15 MUST-PASS. FAIL 0건, PASS-WITH-DEBT 0건.

### 횡단 검증

| 항목 | 명령 | 결과 |
|---|---|---|
| 영향 패키지 스위트 | `go test ./internal/cli/... ./internal/core/project/...` | 전부 `ok` — `internal/cli` 377.953s, `internal/core/project` 2.014s |
| lint | `golangci-lint run --timeout=2m ./internal/cli/... ./internal/core/project/...` | `0 issues.` |
| 크로스 플랫폼 | `GOOS=windows go vet ./internal/cli/...` | 출력 없음, exit 0 |
| race (영향 테스트) | `go test -race ./internal/cli/ -run 'TestStateDir\|TestChain\|TestClean\|TestFindStateDirFromWalksUp\|TestStateM2\|TestResolveTokens\|TestLoadRegistry'` | `ok github.com/modu-ai/moai-adk/internal/cli 4.910s` |
| `m2SetupState` 무영향 | `go test ./internal/cli/ -run 'M2' -v` | `TestStateM2_*` 6건 전부 `--- PASS`. `state_m2_test.go` 의 **수정은 2건의 stderr 단언뿐**이며 `m2SetupState`(line 37)는 무변경 — spec.md §A 의 무영향 주장은 관측으로 확인됐다 |

**로컬 `go test ./...` 는 실행하지 않았다.** 전 패키지 판정은 PR CI 몫이다(CLAUDE.local.md §4).

### 관측된 A 프로브 재실행 (§D.6 첫 항목, 이 트리에서 선행 확인)

spec.md §A 의 "조용한 성공"을 낳은 바로 그 호출을, 이 트리에서 빌드한 바이너리로 재현했다. 전제는 그대로다 — `~/.moai/state` 존재(`drwxr-xr-x 8 goos staff`).

```
$ cd ~/t164probe2/sub && unset CLAUDE_PROJECT_DIR && <this-tree>/moai state show-blocker
  ERROR
  Find state dir: not in a MoAI project (no .moai directory found in project directories).
rc=1
```

이전: `· No blockers found` (rc=0 — 호출자가 지목한 적 없는 `~/.moai/state` 를 상대로 성공). 이후: 에러. **유닛 테스트가 대신할 수 없는 확인이며, 결함이 실제로 사라졌다.** (프로브 디렉터리는 실행 후 제거했다.) §D.6 의 나머지 세 항목은 머지 후 몫으로 남는다.

### 커버리지 (변경된 함수 단위)

`go tool cover -func` 로 측정. 해석 경로의 핵심은 전부 100%:

```
findStateDir            100.0%
findStateDirFrom        100.0%
resolveChainDir         100.0%
disposeLegacyChainDir   100.0%
isRegularFile           100.0%
announceResolvedRoot    100.0%
printResolvedRoot       100.0%
projectRootOf           100.0%
resolveChainDirWith      85.7%
runShowBlocker           83.9%
findStateDirNoEnv        75.0%
runStateDump             72.0%
runClean                 66.7%
resolveTokensStateDir    66.7%
normalizeDir             66.7%
```

85% 미만 5건의 미커버 구간은 전부 **주입 없이는 도달할 수 없는 오류 분기**(`os.Getwd()` 실패, `filepath.EvalSymlinks` 실패)이거나 이 SPEC 이 건드리지 않은 기존 커맨드 본문이다. §D.7 의 "state 해석 경로 커버리지 ≥ 85%" 는 위 100% 항목들이 충족한다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-22
run_commit_sha: pending-backfill-run-final
run_status: complete
ac_pass_count: 15
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-applicable
l44_post_push_fetch: not-applicable
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_amd64: pass
  windows_vet: pass
total_run_phase_files: 9
m1_to_mN_commit_strategy: per-milestone
```

**커밋** (worktree branch `WT-state-walkup`, 미푸시 — 통합은 리드 몫):

| 커밋 | 마일스톤 |
|---|---|
| `977f124d4` | M1 — 해석 계약 + 진입점 분리 + 홈 경계 seam |
| `9de04cb4d` | M2 — chain 정본 경로 + 레거시 4경우 처분 |
| `c4a3b90c1` | M3 — clean carve-out + REQ-10 가시화 |
| (M4-M7) | 검증·문서·progress 기록 커밋 |

**변경 파일 9개**: `internal/cli/{state,chain,clean,tokens}.go`, `internal/core/project/root.go`, `internal/cli/{state_dir_bound_test,chain_state_dir_test,clean_state_dir_test}.go`(신규 3), `internal/cli/{tokens_state_dir_test,state_m2_test}.go`(기존 2 수정). `internal/template/templates/` 무변경(미러 대상 없음), `internal/hook/chain_event.go` 무변경(범위 밖), `internal/hook/cwd_changed_relocate.go` 무변경(범위 밖 — AP-13b).

### 잔여 위험 / 미검증

- **R1 (불변)**: Windows CI 러너의 조상 `.moai/state` 생성 주체는 여전히 미확인이다. 이 변경은 증상을 없애지만 원인을 밝히지 않는다.
- **§D.6 사후 확인 4건은 실행하지 않았다** — 머지 후 실제 머신 확인 항목이며 run-phase 범위가 아니다. 특히 `~` 하위에서의 A 프로브 재실행은 유닛 테스트가 대신할 수 없는 최종 확인으로 남는다.
- **REQ-3 부류 B(R5)**: 이 리포의 `internal/hook/.moai/` 형태가 다시 나타나면 그 하위에서 `moai state`/`chain`/`clean` 이 실패한다. 완화는 실패 메시지가 멈춘 디렉터리를 지목하는 것뿐이며 AC-011 이 그것을 고정한다.
- **새로 관측된 사실(범위 밖)**: `FindProjectRoot` 는 프로젝트 루트가 **$HOME 자신일 때** 거부한다 — 홈 가드가 `.moai` 확인보다 먼저 오기 때문이다. 이 SPEC 이 만든 동작이 아니라 위임 대상의 기존 성질이며 바꾸지 않았다. 테스트 픽스처는 홈 경계를 프로젝트의 부모에 두는 방식으로 이 성질을 피해 간다.
- **§E 4행의 `loadRegistryForOverlay` 안전성은 여전히 "미확인"**이다(spec.md §E). 이 SPEC 은 그 소비자가 읽는 레지스트리 **개수**를 바꾸지 않았다(AP-13c 준수).

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
