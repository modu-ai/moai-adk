---
id: SPEC-LAUNCHER-ARGV-001
title: "Acceptance criteria — launcher worktree-guard argv judgment redesign"
version: "0.1.0"
created: 2026-10-15
updated: 2026-10-15
author: manager-spec (MoAI factory, card t1748)
tier: M
---

# SPEC-LAUNCHER-ARGV-001 수용 기준

판정 입력은 plan.md §F의 구속 선택자(binding selectors)를 따른다. 모든 명령은 본 카드 워크트리 루트에서 단일 호출로 실행한다(파이프·연쇄 없음). 각 셀은 이중 관측을 전제로 한다: 시험의 감지력은 RED 관측(§D.3)이, 구현의 충족은 GREEN 관측이 증명한다.

**공허 스윕 금지 규약(모든 검증 셀)**: 명령은 `-v`로 실행하고 stdout 안에 기명 시험의 `--- PASS: <이름>` 행을 요구하며, 셀마다 기대 실행 시험 수를 함께 적는다. `[no tests to run]`과 bare exit 0은 결코 충족 근거가 아니다 — 스윕 집합이 빈 초록은 판정이 아니다(verification-completeness §1.1).

## §D AC Matrix

### AC-LA-001 — 계열 (a): 소비 옵션 그늘 (REQ-LA-010)

- **Given** `WT-` 카드 브랜치가 체크아웃된 카드 워크트리가 존재하고 **When** `moai cc -f -w -b -- <card-tree>`(및 glm 쌍둥이)을 시험 엔진 시암으로 기동하면 **Then** 런처는 소비 뒤 argv(claude가 받는 `--worktree <card-tree>`)에서 앵커를 읽어 어떤 세션 변이 전에 기동을 거부하고, 거부 사유에 트리와 브랜치를 명명한다.
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorFinalArgv$/consumed_option_shadowing$' -v -count=1` → `--- PASS: TestLauncherAnchorFinalArgv/consumed_option_shadowing` 1행 + 실행 시험 수 1 (bare exit 0은 불충분).

### AC-LA-002 — 계열 (b): 다중 옵션 최후 우선 (REQ-LA-011)

- **Given** 동일한 카드 워크트리와 일반 트리가 존재하고 **When** `-w plain-tree --worktree --branch main <card-tree>` 철자로 리더 기동하면(최종 argv가 worktree 옵션 둘을 지니게 된다) **Then** 런처는 백엔드가 선택하는 최후 옵션의 값을 판정해 카드 트리로의 앵커를 거부한다 — 첫 값(plain-tree)으로 통과시키지 않는다.
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorFinalArgv$/multi_option_last_wins$' -v -count=1` → `--- PASS: TestLauncherAnchorFinalArgv/multi_option_last_wins` 1행 + 실행 시험 수 1 (bare exit 0은 불충분).

### AC-LA-003 — 계열 (c): 묶음 짧은 옵션 (REQ-LA-012)

- **Given** 동일한 카드 워크트리가 존재하고 **When** `moai cc -f -- -pw<card-tree> <프롬프트>`(구분자 뒤 묶음 클러스터)로 기동하면 **Then** 런처는 클러스터를 백엔드 모형으로 분해해 묶인 `-w` 값(카드 트리)을 판정하고 거부한다 — `-w`로 시작하지 않는 토큰이어도 놓치지 않는다.
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorFinalArgv$/bundled_short_cluster$' -v -count=1` → `--- PASS: TestLauncherAnchorFinalArgv/bundled_short_cluster` 1행 + 실행 시험 수 1 (bare exit 0은 불충분).

### AC-LA-004 — 계열 (d): 기록 계수기 실효성 (REQ-LA-013)

- **Given** 거부된 기동의 기록 부재를 단정하는 시험 도구가 있고 **When** 정준 run-store 판독기로 계수를 읽는 긍정 대조(기록이 존재하는 상태)를 실행하면 **Then** 계수가 움직임이 관측된다 — 도구가 항상 0을 읽는 공허한 계수기가 아니라는 것을 그 자체로 증명한다. 부재 단정 셀은 이 긍정 대조와 한 쌍이다.
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorRunCountInstrument$' -v -count=1` → `--- PASS: TestLauncherAnchorRunCountInstrument` + 긍정 대조 하위시험의 `--- PASS:` 행, 실행 시험 수 ≥ 2 (bare exit 0은 불충분).

### AC-LA-005 — 뷰-exec 동일성 (REQ-LA-001·002)

- **Given** 리더·비리더를 아우르는 argv 철자표(§D.4 표)가 있고 **When** 각 철자로 해석기를 실행해 뷰의 `finalArgv`와 exec 경계 시암이 포착한 argv를 비교하면 **Then** 두 argv는 토큰 단위로 동일하고, 각 열의 기대 선택 필드(철자·값·해석된 트리 경로·실효 브랜치)와 기대 판정(허용/거부)이 뷰에서 그대로 단언된다 — 바이트 동일만으로는 충분하지 않다(모든 `-w` 값을 틀리게 파싱하는 변이는 §D.4의 기대 선택 단언에서 죽는다).
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorFinalArgv$/view_exec_byte_equality$' -v -count=1` → `--- PASS: TestLauncherAnchorFinalArgv/view_exec_byte_equality` 1행 + 실행 시험 수 1, 하위시험 안에서 §D.4 전 행의 기대 선택·판정 단언 실행 (bare exit 0은 불충분).

### AC-LA-006 — 변이 전 거부 (REQ-LA-004)

- **Given** 카드 워크트리 앵커(기동 디렉터리 또는 `-w` 지명)가 있고 **When** 리더 기동이 거부되면 **Then** 리더 역할 마커가 환경에 새겨지지 않고, run 기록 계수가 불변이며, settings 준비 산물이 없다 — 세션 변이가 전부 판정 뒤에 있음이 세 지표로 관측된다. cc·glm 양 팔.
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorGuardSemantics$/refusal_before_mutation$' -v -count=1` → `--- PASS: TestLauncherAnchorGuardSemantics/refusal_before_mutation` 1행 + 실행 시험 수 1 (bare exit 0은 불충분).

### AC-LA-007 — 레인·비팩토리 면제 (REQ-LA-007)

- **Given** 같은 카드 워크트리 앵커가 있고 **When** 레인 기동(`-l`/lane-shape) 또는 팩토리 진입 없는 기동을 하면 **Then** 가드는 발화하지 않고 기동이 진행된다(시험 엔진 시암 관측) — 카드 트리 앵커가 의도인 경로는 오늘과 같다.
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorGuardSemantics$/lane_and_nonfactory_exempt$' -v -count=1` → `--- PASS: TestLauncherAnchorGuardSemantics/lane_and_nonfactory_exempt` 1행 + 실행 시험 수 1 (bare exit 0은 불충분).

### AC-LA-008 — 재정의 + 감사 행 + 기록 실패 거부 (REQ-LA-005)

- **Given** 카드 워크트리 앵커가 있고 **When** `--allow-card-anchor`와 함께 기동하면 **Then** 경고 후 진행되고 감사 행(백엔드·트리·브랜치)이 기록된다. **When** 감사 로그를 기록할 수 없는 상태(경로가 디렉터리 등)에서 같은 기동을 하면 **Then** 기동은 거부된다 — 기록되지 않는 우회는 없다. 플래그가 다른 옵션의 값 자리에 있을 때는 재정의로 인정되지 않는다.
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorGuardSemantics$/override_audit_row$' -v -count=1` → `--- PASS: TestLauncherAnchorGuardSemantics/override_audit_row` 1행 + 실행 시험 수 1 (bare exit 0은 불충분).

### AC-LA-009 — 감사 기록 위생 (REQ-LA-009)

- **Given** 감사 로그 경로가 있고 **When** 그 경로가 symlink이거나 FIFO인 상태에서 재정의 기동하면 **Then** 개방 전 거부(Lstat)와 서술자 형식 재판독 거부가 각각 발화한다 — 행 손실 없는 기록 위생이 이식 셀로 묶인다.
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorAuditHygiene$' -v -count=1` → `--- PASS: TestLauncherAnchorAuditHygiene` + symlink·FIFO·서술자 재판독 하위시험의 `--- PASS:` 행, 실행 시험 수 ≥ 3 (bare exit 0은 불충분).

### AC-LA-010 — fail-open 방향 (REQ-LA-008)

- **Given** 브랜치를 판독할 수 없는 트리(부재·판독 실패)가 `-w`로 지명되고 **When** 리더 기동하면 **Then** 기동은 허용된다 — 거부는 카드 접두의 긍정적 증거 위에서만 한다.
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorGuardSemantics$/fail_open_on_unverifiable$' -v -count=1` → `--- PASS: TestLauncherAnchorGuardSemantics/fail_open_on_unverifiable` 1행 + 실행 시험 수 1 (bare exit 0은 불충분).

### AC-LA-011 — 판별자 환경 독립 (REQ-LA-006)

- **Given** 세션 환경에 오래된 라벨(카드 앵커를 암시하는 env)이 있고 **When** 주 체크아웃(main 브랜치)에서 리더 기동하면 **Then** 기동은 허용된다 — 판별은 브랜치 접두뿐이다. 대조: 같은 환경에서 `WT-` 앵커는 거부된다.
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorGuardSemantics$/discriminator_ignores_env$' -v -count=1` → `--- PASS: TestLauncherAnchorGuardSemantics/discriminator_ignores_env` 1행 + 실행 시험 수 1 (bare exit 0은 불충분).

### AC-LA-012 — `--branch` 생성/재사용 주석 (REQ-LA-004 실효 브랜치 정의)

- **Given** (i) `WT-` 브랜치가 체크아웃된 기존 카드 트리와 (ii) 존재하지 않는 트리 이름이 있고 **When** (i)에 `-w <기존> --branch <다른 브랜치>`로, (ii)에 `-w <신규> --branch WT-<카드>`로 리더 기동하면 **Then** (i)은 그 트리의 실제 체크아웃 브랜치를 판정해 거부하고(플래그 값을 믿지 않는다 — iter6-F2 보존), (ii)는 생성 주석의 브랜치 값을 판정해 materialize 전에 거부한다.
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorBranchAnnotation$' -v -count=1` → `--- PASS: TestLauncherAnchorBranchAnnotation` + (i) 기존 트리·(ii) 생성 주석 하위시험의 `--- PASS:` 행, 실행 시험 수 ≥ 2 (bare exit 0은 불충분).

### AC-LA-013 — 단일 해석 보장 (REQ-LA-001)

- **Given** 기동 경로 호출 클로저 파일 집합(cc.go, glm.go, launcher.go, codex_hooks_seed.go, session_worktree.go, worktree_branch_flag.go, launch_chain.go — cc/glm 진입 → launch → runLaunchClaude가 도달하는 파일)이 있고 **When** 소스 스캔 시험이 해석기와 그 산책 헬퍼 바깥의 worktree 값 재유도 패턴(`worktreeFlagValue` 호출 포함)을 클로저 전체에서 세면 **Then** 0건이다 — worktree 선택의 판독은 해석기 한 곳뿐이며, 헬퍼 파일 뒤에 숨은 판독도 집합 안에서 센다(plan-repair 1: 스캔 집합을 호출 클로저로 고정 — 파일 역할 나열만으로는 codex_hooks_seed.go 뒤의 판독이 스캔을 통과했다).
- 검증: `go test ./internal/cli/ -run '^TestLaunchViewSingleParse$/no_rederivation_on_launch_path$' -v -count=1` → `--- PASS: TestLaunchViewSingleParse/no_rederivation_on_launch_path` 1행 + 실행 시험 수 1 (bare exit 0은 불충분).

### AC-LA-014 — 비리더 경로 행동 보존 (§2 제약)

- **Given** 기존 런처 워크트리 규약 시험(`normalizeWorktreeFlag` 철자표, L2 접두 검증, `--branch` materialize, 작성자 사전 점검, lane 재개 어셈블러)이 있고 **When** 단일 해석기 리팩터 뒤 그 패밀리를 실행하면 **Then** 전부 GREEN이다 — 비리더 기동의 관측 행동은 바뀌지 않는다.
- 검증: `go test ./internal/cli/ -run 'Worktree|TestLaneJoinChildArgv' -v -count=1` → `--- PASS: TestLaneJoinChildArgv` 존재 + `Worktree` 매칭 패밀리 최상위 `--- PASS:` ≥ 1행, 실행 시험 수 ≥ 2 (bare exit 0·`[no tests to run]`은 불충분). 선택자는 lane 재개 어셈블러 시험 `TestLaneJoinChildArgv`(lane_resume_test.go:19 — 본 트리 실측, plan-repair 1)를 명시적으로 닿게 한다; 세부 선택은 run 증거에 기록.

### AC-LA-015 — RED 관측 기록 (§D.3 폐쇄 게이트)

- **Given** 네 계열 셀과 변형 프로토콜이 있고 **When** run 단계가 종료할 때 **Then** progress.md §E.2에 계열 (a)(b)(c)(d) 각각의 RED 기록이 네 요소(명령·원문 stdout·exit code 자체 필드·고정 tree SHA)를 모두 지니고 셀마다 green-flip 마일스톤 레이블이 적혀 있으며, 돌연변이가 되돌려진 뒤 최종 GREEN 관측과 한 쌍을 이룬다.
- 검증: run 종료 감사가 §E.2의 기록을 판독한다 — 기계 열: 계열 레이블마다 ① 명령 존재 ② 원문 stdout 존재 ③ exit code 필드 존재 ④ 고정 tree SHA 존재(커밋 SHA 형태, 브랜치 이름 아님) ⑤ green-flip 마일스톤 레이블 존재 — 다섯 중 하나라도 결시면 그 계열의 RED 기록은 미채택으로 계수한다(셋만으로는 채택 아님 — §D.3·verification-completeness §2.1).

## §D.1 심각도 분포

- Blocker(P1): AC-LA-001..005 (판정면 본령 + 동일성), AC-LA-006 (변이 전 거부).
- Major: AC-LA-007..013 (보존 의미 + 단일 해석).
- Gate: AC-LA-014 (행동 보존), AC-LA-015 (증거 프로토콜).
- Blocker 셀이 하나라도 FAIL이면 run 단계는 종료하지 않는다.

## §D.2 요구사항 대응표

| REQ | AC |
|---|---|
| REQ-LA-001 | AC-LA-005, AC-LA-013 |
| REQ-LA-002 | AC-LA-005 |
| REQ-LA-003 | AC-LA-001..003, AC-LA-005 |
| REQ-LA-004 | AC-LA-006, AC-LA-012 |
| REQ-LA-005 | AC-LA-008 |
| REQ-LA-006 | AC-LA-011 |
| REQ-LA-007 | AC-LA-007 |
| REQ-LA-008 | AC-LA-010 |
| REQ-LA-009 | AC-LA-009 |
| REQ-LA-010 | AC-LA-001 |
| REQ-LA-011 | AC-LA-002 |
| REQ-LA-012 | AC-LA-003 |
| REQ-LA-013 | AC-LA-004 |
| §2 비리더 보존 | AC-LA-014 |
| §D.3 프로토콜 | AC-LA-015 |

## §D.3 이중 관측(RED/GREEN) 채택 프로토콜

각 계열 셀은 한 쌍으로 채택된다.

**RED-now 셀**(교체 전 상태에서의 실패 관측, progress.md §E.2에 기록)은 네 요소를 **모두** 지닌다 — 하나라도 빠진 셀은 미채택이다(verification-completeness §2.1):

1. **명령** — 단일 호출 형태(파이프·리다이렉션·`&&`·`;`·서브셸 밖).
2. **그 명령의 원문 stdout** — 가공 없는 raw 출력.
3. **exit code** — 자체 필드로 기록한다(단일 호출 형태는 `; echo $?`를 금하므로 명령 안에 실릴 수 없다 — 이 분리가 존재하는 이유다).
4. **고정 tree SHA** — 측정 시점의 커밋 SHA(브랜치 이름 아님).

셀마다 **green-flip 마일스톤**(어느 마일스톤이 무엇을 GREEN으로 만드는지, 기대되는 통과 출력)을 함께 적는다. RED의 원인은 명명된다: (a)-(c)는 소비 전 argv 판독 변형(mutant), (d)는 레거시 경로 계수기 변형. 변형은 관측 뒤 되돌린다 — RED가 "올바른 이유"로 적색임을 변형의 정의가 보증한다. AC-LA-015의 기계 열이 이 다섯 항목(네 요소 + 마일스톤 레이블)을 계열별로 점검한다.

## §D.4 간접 검증 — argv 철자표 (AC-LA-005의 최소 열)

해석기-동일성 시험은 최소한 다음 철자를 표로 지니며, 각 열에서 ① 뷰의 `finalArgv` == exec 시암 포착 argv(토큰 단위), ② 기대 worktree 선택(철자·값·해석된 트리·실효 브랜치), ③ 기대 판정(허용/거부)을 **모두** 단언한다 — 바이트 동일만으로 충분하지 않다(plan-repair 1: 모든 `-w` 값을 틀리게 파싱하는 변이가 동일성만으로 통과하는 눈금을 막는다). 표의 트리 관례: `WT-card`는 `WT-` 브랜치가 체크아웃된 카드 트리, `plain`은 비카드 일반 트리.

| # | 진입 철자 | 기대 선택 (철자·값·해석 트리·실효 브랜치) | 기대 판정 |
|---|---|---|---|
| 1 | `-w WT-card` | 짧은 철자, 값 `WT-card`, 트리 `.claude/worktrees/WT-card`, 실효 브랜치 `WT-card` | 거부 |
| 2 | `--worktree WT-card` | 2토큰 철자, 값 `WT-card`, 같은 트리, `WT-card` | 거부 |
| 3 | `--worktree=WT-card` | 등호 철자, 값 `WT-card`, 같은 트리, `WT-card` | 거부 |
| 4 | `-wWT-card` | 붙임 철자, 값 `WT-card`, 같은 트리, `WT-card` | 거부 |
| 5 | `-w -b -- WT-card`(계열 a) | 소비 뒤 claude argv의 `--worktree WT-card`, 값 `WT-card`, 같은 트리, `WT-card` | 거부 |
| 6 | `-w plain --worktree --branch main WT-card`(계열 b) | `--branch` 제거 뒤 옵션 둘, 최후 우선 → 값 `WT-card`, 같은 트리, `WT-card` — 첫 값 `plain`으로 통과 아님 | 거부 |
| 7 | `-- -pwWT-card <프롬프트>`(계열 c) | 클러스터 분해(불리언 `p` + 값 옵션 `w`), 값 `WT-card`, 같은 트리, `WT-card` | 거부 |
| 8 | 값 없는 `-w` | 선택 없음(auto-name) — 해석 트리 없음, 실효 브랜치는 기동 디렉터리 브랜치 | 기동 디렉터리가 비카드 브랜치면 허용 |
| 9 | `-- --worktree WT-card <프롬프트>` | 꼬리 통과 뒤 백엔드 모형이 읽는 옵션, 값 `WT-card`, 같은 트리, `WT-card` | 거부 |
| 10 | `--worktree=WT-card -- <꼬리>` | 값 `WT-card`, 같은 트리, `WT-card`; 구분자 뒤 꼬리는 프롬프트 텍스트 | 거부 |
| 11 | `-w plain -- "<-w WT-card 토큰이 본문에 섞인 프롬프트>"` | 프롬프트 값 경계 — 구분자 뒤 본문은 옵션이 아니라 프롬프트 텍스트 → 값 `plain`, 일반 트리, 비카드 브랜치 | 허용 |
| 12 | `-w plain -- <p1> -- --worktree WT-card` | 두 번째 `--` 경계 — 런처는 첫 구분자만 소비하고, 꼬리의 두 번째 `--`는 백엔드 모형의 옵션 종결자(그 뒤 토큰은 값 아님) → 값 `plain` | 허용 |

11열은 프롬프트 본문을 스캔에 넣는 판독기가, 12열은 꼬리의 `--`를 구분자로 오해하는 판독기가 각각 죽는 경계다. 12열의 백엔드 측 `--` 규약은 C-1 재측정으로 확정하며 어긋나면 REQ-LA-003 경유로 정정한다(모형 추측 고정 금지).

## §D.5 폐쇄 게이트

1. 네 계열 셀 GREEN + RED 기록 한 쌍(AC-LA-015).
2. 보존 의미 여덟 GREEN(AC-LA-006..012, 009 포함).
3. 단일 해석 + 비리더 보존 GREEN(AC-LA-013·014).
4. 크로스 플랫폼 빌드 + 영향 패밀리 회귀 0 + 커버리지 85%+ + lint 신규 0.
5. C-1 백엔드 모형 재측정 원문이 run 증거에 존재.

## §D.6 선행 관측(미래 지향 검사)

- 판정면이 뷰만 음을 재확인하는 소스 스캔(AC-LA-013)은 이후 argv 관련 변경에서도 회귀 감지기로 남는다.
- 계열 셀의 철자표는 새 철자 발견 시 표에 추가하는 확장점이다 — 표가 곧 회귀 감시 목록이다.
