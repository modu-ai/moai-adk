# Plan — SPEC-STATE-ANCHOR-001

## §A 맥락

카드 **t510**. 근거 보고서: `.moai/reports/t510/verdict.md` (plan-phase 판정서 — 두 축 열거 + HEAD 격리 재현 §B-2 + 수리 방향 §"수리 방향 (SPEC 입력)"). 작업 트리 `.claude/worktrees/t510`, 브랜치 `WT-state-write-locus`, base `0b1e27877` (= origin/develop).

### 카드 등급 판정

배차 등급은 **Class B**(결함, 원인 미상)였다. t510 판정이 원인을 멤버 단위로 측정 확정했고, 남은 작업에 설계 결정이 있다 — 상태-앵커 리졸버의 우선순위 체인. 결정을 포함하므로 실질 **Class C**(설계 변경)로 다뤄 SPEC을 세운다. 등급 상향의 근거는 속도가 아니라 결정의 존재다.

### Tier 판정 — M

| 축 | 값 | Tier M 상한 | 여유 |
|---|---|---|---|
| REQ | 11 (REQ-SA-001..011) | 16 | 5 |
| AC | 12 (AC-SA-001..012) | 16 | 4 |

- **S 아님**: 멤버 4곳(statusline ×3 좌표 + config) + 테스트 패밀리(cli todo)가 한 변경 안에 있고, 마일스톤이 5개다. 단일 파일 편집이 아니다.
- **L 아님**: 파일 5~10개(statusline `context_usage.go`/`landed.go`/`builder.go`/`backlog.go` + 신규 anchor 리졸버 1 + `internal/config` 1 + 테스트 3~4), LOC 수백 수준, constitutional 범위 없음. 새 아키텍처가 아니라 기존 시접 2개(`memory.go` 워크업, `todo_root.go` common-dir 해석)의 통합이다.
- 따라서 **Tier M**: `spec.md` + `plan.md` + `acceptance.md` + `progress.md`. `design.md` / `research.md` 는 만들지 않는다.

### 개발 방법론

**TDD (RED-GREEN-REFACTOR)** — `quality.yaml` `constitution.development_mode` 기본값이며 본 SPEC의 성격(기존 동작의 결함 수리)과 정확히 일치한다. RED-first 순서는 §F 마일스톤에 구속 조건으로 박혀 있다: run phase는 **B1 재현 테스트(committed RED)에서 시작**하고, 이후 멤버마다 각자 RED→GREEN을 관측한다.

## §B 알려진 결함 (이미 측정됨 — 재론 금지)

1. **B1**: `resolveProjectDir` 사슬이 `current_dir`에 앵커 — `workspace.project_dir` 후보 누락 + git 해석 없음 (`context_usage.go:278`). HEAD 격리 재현 완료(판정서 §B-2, 트리 `0b1e27877`).
2. **B2**: `resolveBoardRoot`가 `original_cwd` 없으면 B1 사슬로 폴백 (`backlog.go:24`).
3. **B3**: goal 읽기가 같은 리졸버 — cd한 세션이 goal 상태를 못 봄 (`builder.go:286`, **읽기 측 결함**).
4. **B4**: config 캐시 경로가 CLI 사슬 configDir에 좌우 — 정확한 유도 지점 미확정(Gap, M3 RED가 고정).
5. **축 A**: 이미 수리됨(`e7a078970`, t422). 잔여 리스크 2건 — (a) 가드 표면이 `runTodo`/`runTodoWithClosedStdin` 2 헬퍼뿐(`newTodoCmd()` 직접 Execute 우회 시 무가드), (b) 생산 홈 폴백이 설계로 fail-open — 비git 임시 디렉터 호출은 지금도 홈 큐를 만든다(A-4 표본 2건). (b)는 운영자 미결 결정(spec.md §5)이고 본 SPEC 범위 밖이다.

## §C 사전 점검 (run-phase 진입 시 재측정 — 값이 다르면 멈추고 보고)

| 명령 | 기대값 |
|---|---|
| `git rev-parse --short HEAD` + `git branch --show-current` | `0b1e27877` 계열 + `WT-state-write-locus` (다르면 흡수 후 재판정) |
| B1 committed RED probe (판정서 §B-2 형태를 `internal/statusline` 테스트로 커밋) | stdin에 `current_dir`(디렉터 A) + `project_dir`(디렉터 B) 둘 다 → 레코드가 A에 착지, B 무오염 → **FAIL (RED)** |
| `grep -n "ProjectDir" internal/statusline/context_usage.go` | 0 매치 (project_dir이 후보에 없음의 재확인) |
| `grep -n "liveTodoQueueRootReason" internal/cli/todo_test.go` | ≥1 (t422 가드 존재) |
| canary-HOME 스윕 baseline (`internal/cli` todo 서브셋, canary HOME) | canary `HOME/.moai/todo` 신규 디렉터 **0** + swept count > 0 (현재 GREEN — regression-guard baseline) |
| `go test ./internal/statusline/... ./internal/config/...` | ok (기존 테스트 전부 통과 — 수리가 깨뜨리면 안 되는 것들의 baseline) |

## §D 구속 조건 (재논의 금지)

- **D1 — 앵커 우선순위 체인은 고정이다.** `project_dir` → `worktree.original_cwd` → git common-dir 해석(`gitcore.ResolveGitDirs` 재사용, `primaryCheckoutRoot` 형태) → 실패 시 `""`. 중간 단계 삽입·순서 변경은 요구사항 변경이다(REQ-SA-002).
- **D2 — 표시는 불변이다.** statusline 표시 세그먼트의 디렉터 이름 유도(basename)는 `current_dir`에서 계속 따로따로 한다. `project_dir`을 표시에까지 퍼뜨리는 수리는 금지(REQ-SA-004). 읽기·쓰기·표시의 관심 분리는 판정서 Residual-risk 3의 요구다.
- **D3 — throttle과 silent-failure는 보존이다.** write-if-changed skip과 best-effort 무소음(REQ-THRESHOLD-009/012) 의미론을 잃는 수리는 금지(REQ-SA-005).
- **D4 — B5/B6 불변.** `internal/hook/**`, `internal/session/**`은 변경 금지. R1 시접이 반드시 그들을 건드려야 한다는 것이 입증되면 blocker 보고로 plan을 확장한 뒤에만 재검토한다(REQ-SA-008).
- **D5 — 테스트 격리는 t.TempDir()뿐이다.** 어떤 테스트도 개발자 실제 HOME에 쓰지 않는다(CLAUDE.local.md §6 HARD). HOME 교체가 필요한 테스트는 비-parallel 서브테스트에서만 `t.Setenv`로 한다(Go가 parallel 조합을 panic으로 차단한다 — §13의 병렬 오염 위험을 기계적으로 회피하는 경로). canary 스윕은 canary HOME에서만 판정하고 **실제 HOME을 측정 대상으로 삼지 않는다** — 측정 행위 자체가 오염을 만들면 안 된다(D12).
- **D6 — 로컬 검증 범위는 패키지 단위다.** `internal/statusline`, `internal/config`, `internal/cli`(todo 서브셋), `internal/kanban`(건드릴 때만). `go test ./...` 로컬 전체 실행 금지(CLAUDE.local.md §6) — 전 패키지 판정은 CI 몫이다.
- **D7 — 오염 디렉터 343개는 증거다.** `~/.moai/todo/` 아래 기존 오염 디렉터를 어떤 AC·테스트·정리 단계도 삭제하지 않는다.
- **D8 — RED-first는 멤버 단위다.** B1의 RED는 판정서 §B-2 형태의 **committed** 테스트로 재현하고(C0), 이후 B2→B3→B4 순서로 **각자** RED→GREEN을 관측한다. RED 없이 GREEN을 주장하는 멤버 수리는 없다.
- **D9 — 부재 가드의 판별은 뮤턴트다.** canary 0-오염만으로는 가드의 물성이 증명되지 않는다(verification-completeness §2: 부재 가드 AC는 RED-now로 채택 불가). 가드를 우회하는 뮤턴트 헬퍼의 오염 관측(AC-SA-011)이 채택 요건이며, **못 잡은 뮤턴트도 보고한다** — 그것이 가드의 경계다.
- **D10 — 축 A에는 코드 수리가 없다.** `e7a078970` 이후의 축 A 코드 변경은 금지다. 축 A 마일스톤(M4)은 검증 AC 실행과 그 증거 보존만 한다.
- **D11 — 홈 폴백 임시-디렉터 가드는 구현하지 않는다.** spec.md §5의 운영자 미결 결정이다. 본 SPEC의 run-phase는 이 결정을 기다리지 않는다 — `internal/kanban` 접촉은 시접 참조(코드 변경 없음)뿐이다.
- **D12 — 검증의 self-contamination 금지.** AC 판정에 쓰는 측정(스윕, probe)은 canary 경로 안에서만 일어난다. 판정서 A-7의 교훈 — read sweep조차 mtime 오염을 만든다 — 을 적용해, 실제 HOME을 향한 읽기·쓰기·mtime 접촉을 전부 배제한다.
- **D13 — 세션 텔레메트리 스키마 불변.** `SessionTelemetryRecord`와 `.moai/state/context-usage/` 경로 체계는 변경 금지 — 고치는 것은 앵커뿐이다(spec.md Out of Scope).

## §E 자가 검증

각 마일스톤 종료 시 §C 표의 해당 행과 acceptance.md §D.1의 판정 명령을 재측정하고 **실제 출력을 그대로** `progress.md` §E.2에 인용한다. 요약 문장("전부 통과")은 증거가 아니다. §E 항목별 귀속 삼인조(명령 + 관측 출력 + baseline attribution `(this run, this tree)` + HEAD SHA)를 채운다 — `manager-develop-prompt-template.md` §E attribution discipline.

## §F 마일스톤

### 순서 구속 — 권고가 아니라 의존성

[HARD] **시접이 먼저고 멤버 흡수가 나중이다.** 리졸버가 없는 상태에서 멤버부터 고치면 멤버마다 애드혹 앵커가 생기고, 시접 착지 시 그것들을 다시 걷어낸다. 결과적으로 **M1 시접+B1 → M2 B2/B3 → M3 B4 → M4 축 A 검증 → M5 배치** 순서는 의존성이다.

가장 바뀔 가능성이 큰 결정 — 앵커 우선순위 체인 — 은 M1 착수 전에 §D1과 spec.md REQ-SA-002로 **지금 확정**한다. 결정을 앞에 두고 착지를 뒤에 두는 형태다.

### C0 — B1 committed RED (M1 첫 단계)

판정서 §B-2의 throwaway 프로브를 **커밋되는 RED 테스트**로 재현한다:

- `internal/statusline` 패키지 테스트: stdin에 `workspace.current_dir`(임의 디렉터 A, `t.TempDir()`)과 `workspace.project_dir`(임의 디렉터 B, `t.TempDir()`)을 **둘 다** 주고 렌더를 수행한다.
- 단언: 레코드는 `B/.moai/state/context-usage/<sid>.json`에 존재하고, A 아래에는 `.moai`가 **존재하지 않는다**.
- HEAD(`0b1e27877`)에서 이 테스트는 **FAIL (RED)** 이다 — 판정서가 같은 형태의 프로브로 A에 착지하는 것을 이미 관측했다. RED 출력 전문을 progress.md §E.2에 남긴다(manager-develop E8 항목).
- 이 테스트는 수리 후 GREEN이 되며, 이후에도 **회귀 가드로 계속 산다** — #1694의 재발을 잡는 최저 비용 신호다.

### M1 — 상태-앵커 리졸버 시접 + B1 GREEN

- `internal/statusline`(또는 그보다 낮은 공유 위치로의 이동이 자연스러우면 그렇게 — 단, 이동은 M1 내 최소로)에 단일 상태-앵커 리졸버를 둔다. 우선순위: `project_dir` → `worktree.original_cwd` → `gitcore.ResolveGitDirs` common-dir 부모 → `""`(REQ-SA-002).
- B1 `resolveProjectDir`를 리졸버로 교체한다. `current_dir`/`input.CWD`/`os.Getwd()`는 상태 쓰기 앵커 후보에서 **제거**되고, 표시 이름 유도용으로만 남는다(D2).
- C0 테스트 GREEN. 무프로젝트 케이스(AC-SA-005)의 RED→GREEN도 같이 관측한다.
- 기존 statusline 테스트 전수 통과 유지(§C 마지막 행 baseline).

**M1 종료 조건**: C0 GREEN + AC-SA-001·005·006·007 GREEN + 기존 테스트 무파괴. RED 출력 전문 §E.2 보존.

### M2 — B2/B3 흡수

- B2: `resolveBoardRoot`(`backlog.go:24`)가 리졸버를 통해 board root를 얻는다. `worktree.original_cwd` 우선은 리졸버 체인의 2단계로 **흡수**된다(별도 유산 제거). RED: `original_cwd` 없이 `current_dir` = 서브디렉터인 입력에서 현재 board root가 서브디렉터인 것을 테스트로 관측(RED) → 수리 후 앵커에서 해석(GREEN).
- B3: `builder.go:286`의 goal 읽기가 리졸버의 앵커에서 읽는다. RED: goal 상태를 프로젝트 루트에 시드하고 `current_dir` = 서브디렉터로 렌더 → 현재는 못 봄(RED) → 수리 후 GoalArmed=true(GREEN).
- **가시성 변화 주의**(판정서 Residual-risk 3): B3 수리는 "cd한 세션이 goal을 보게 된다"는 사용자 가시성 변화를 수반한다 — 의도된 수리이지만, 기존에 cwd 기반 읽기에 의존하던 테스트·소비자가 있는지 M2에서 확인한다.

**M2 종료 조건**: AC-SA-002·003 GREEN, 기존 테스트 무파괴.

### M3 — B4 (config 캐시)

- **먼저 유도 지점을 RED 테스트로 고정한다**(판정서 Gap 해소). `internal/config`의 configDir 사슬(env → Getwd 패턴)에서 캐시 경로 유입점을 관측하는 테스트 — cwd에 좌우되는 현재 형태를 RED로 남긴다.
- 이후 `cacheFilePath`의 소유자(`LoadWithCache` 호출 사슬)가 리졸버의 앵커(또는 그와 동일한 우선순위 체인)를 통해 configDir를 얻게 한다. `statusline` → `config` 역방향 의존이 생기지 않게 해석 위치는 호출 사슬 쪽에서 정한다 — 시접의 **규칙**은 REQ-SA-002 하나고, 그 적용 지점은 M3에서 결정한다.
- lane-1 t507 관측(`fixtures/.moai/state/config-cache.json`)이 재현 형태의 대조군이다.

**M3 종료 조건**: AC-SA-004 GREEN(RED 관측 포함), `go test ./internal/config/...` 무파괴.

### M4 — 축 A 검증 (코드 수리 없음 — D10)

- **가드 활성 실측**(AC-SA-009): `runTodo` 게이트(`liveTodoQueueRootReason`)의 존재와 활성을 관측하는 검증. 가드 판별식의 정확한 표면은 기존 구현을 읽고 테스트로 고정한다 — 이 SPEC이 가드를 다시 쓰지 않는다.
- **canary-HOME 스윕**(AC-SA-010): `internal/cli` todo 패밀리 전체를 canary HOME(`t.TempDir()` 기반)에서 실행 → canary `HOME/.moai/todo` 신규 디렉터 **0** + swept count N 명시. **선택자 0매치는 판정 아님** — N=0이면 테스트가 스스로 실패한다.
- **뮤턴트 증명**(AC-SA-011): 가드를 우회하는 뮤턴트 헬퍼(`newTodoCmd()` 직접 Execute)가 canary HOME 오염을 만들어내는 것을 관측 — 가드가 실제로 잡는 것의 증거. **못 잡은 뮤턴트도 보고**한다(D9).

**M4 종료 조건**: AC-SA-009·010·011 관측 완료, 실제 HOME 오염 0(D5/D12), 축 A 코드 diff 0.

### M5 — 배치 검증 + sync 인계

- §C 전 행 재측정 + AC 매트릭스 전수 판정, 출력 전문 §E.2.
- `git diff 0b1e27877..HEAD -- internal/hook/ internal/session/` 무변경 확인(AC-SA-008).
- **GH #1694 회신은 sync-phase 태스크다**: 수리 착지 후 제보자(binsworld)에게 — (a) 원인(statusline이 상태를 current_dir에 기록), (b) 수리 내용, (c) 제보자 환경의 226개 stray `.moai` 정리 안내(제보자 자신의 정리 대상 — Out of Scope) — 를 안내한다. sync 단계에서 초안 확정.

**M5 종료 조건**: AC 전수 판정 완료 + 회신 초안 재료가 sync-phase에 인계됨.

## §G Anti-Patterns (이 SPEC이 금지하는 것)

- **셀렉터 0매치 초록** — canary 스윕이 0개 테스트를 돌려도 "0 오염"으로 통과시키는 것. swept count N>0 강제(AC-SA-010).
- **실제 HOME 측정** — AC 판정을 개발자 실제 HOME에서 수행해 오염을 재생산하거나 mtime을 더럽히는 것(D5/D12). 판정서 A-7의 read-sweep 오염 전례.
- **표시까지의 앵커 전파** — `project_dir`을 렌더 표시에 쓰는 것(D2). 표시는 `current_dir` 기반 유지.
- **부분 수리** — B1만 고치는 것. 멤버 4곳 전부가 시접을 거쳐야 REQ-SA-001 충족이다.
- **전체 스위트 로컬 실행** — `go test ./...` 금지(D6). 패키지 범위만.
- **축 A 코드 접촉** — 가드를 "개선"하거나 폴백을 "고치는" 것(D10/D11). 둘 다 본 SPEC 밖이다.

## §H 상호 참조

- 근거: `.moai/reports/t510/verdict.md` (§B-1 멤버 표 = 본 SPEC의 멤버 경계, §B-2 = C0 재현 원형, §A-5/A-6 = 축 A 근거, §수리 방향 = REQ 원형)
- 관련 SPEC: `SPEC-SESSION-TELEMETRY-001` (세션 텔레메트리 스키마 — 불변 유지 대상, D13)
- 규칙: `.claude/rules/moai/development/verification-completeness.md` (두 셀 규율 + 뮤턴트 판별 + swept count), `.claude/rules/moai/core/verification-claim-integrity.md` (§E 귀속 삼인조)
- 외부: GH #1694 (제보 binsworld — sync-phase 회신 대상)
