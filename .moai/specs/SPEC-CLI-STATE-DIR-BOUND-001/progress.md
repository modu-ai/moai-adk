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

AC-008 의 빨간 출력은 M2 절에 별도로 인용한다.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
