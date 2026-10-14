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

## §D AC Matrix

### AC-LA-001 — 계열 (a): 소비 옵션 그늘 (REQ-LA-010)

- **Given** `WT-` 카드 브랜치가 체크아웃된 카드 워크트리가 존재하고 **When** `moai cc -f -w -b -- <card-tree>`(및 glm 쌍둥이)을 시험 엔진 시암으로 기동하면 **Then** 런처는 소비 뒤 argv(claude가 받는 `--worktree <card-tree>`)에서 앵커를 읽어 어떤 세션 변이 전에 기동을 거부하고, 거부 사유에 트리와 브랜치를 명명한다.
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorFinalArgv$/consumed_option_shadowing$' -count=1` → exit 0.

### AC-LA-002 — 계열 (b): 다중 옵션 최후 우선 (REQ-LA-011)

- **Given** 동일한 카드 워크트리와 일반 트리가 존재하고 **When** `-w plain-tree --worktree --branch main <card-tree>` 철자로 리더 기동하면(최종 argv가 worktree 옵션 둘을 지니게 된다) **Then** 런처는 백엔드가 선택하는 최후 옵션의 값을 판정해 카드 트리로의 앵커를 거부한다 — 첫 값(plain-tree)으로 통과시키지 않는다.
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorFinalArgv$/multi_option_last_wins$' -count=1` → exit 0.

### AC-LA-003 — 계열 (c): 묶음 짧은 옵션 (REQ-LA-012)

- **Given** 동일한 카드 워크트리가 존재하고 **When** `moai cc -f -- -pw<card-tree> <프롬프트>`(구분자 뒤 묶음 클러스터)로 기동하면 **Then** 런처는 클러스터를 백엔드 모형으로 분해해 묶인 `-w` 값(카드 트리)을 판정하고 거부한다 — `-w`로 시작하지 않는 토큰이어도 놓치지 않는다.
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorFinalArgv$/bundled_short_cluster$' -count=1` → exit 0.

### AC-LA-004 — 계열 (d): 기록 계수기 실효성 (REQ-LA-013)

- **Given** 거부된 기동의 기록 부재를 단정하는 시험 도구가 있고 **When** 정준 run-store 판독기로 계수를 읽는 긍정 대조(기록이 존재하는 상태)를 실행하면 **Then** 계수가 움직임이 관측된다 — 도구가 항상 0을 읽는 공허한 계수기가 아니라는 것을 그 자체로 증명한다. 부재 단정 셀은 이 긍정 대조와 한 쌍이다.
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorRunCountInstrument$' -count=1` → exit 0 (긍정 대조 하위시험 포함).

### AC-LA-005 — 뷰-exec 동일성 (REQ-LA-001·002)

- **Given** 리더·비리더를 아우르는 argv 철자표(§D.4 표)가 있고 **When** 각 철자로 해석기를 실행해 뷰의 `finalArgv`와 exec 경계 시암이 포착한 argv를 비교하면 **Then** 두 argv는 토큰 단위로 동일하다 — 가드가 읽은 것과 백엔드가 받은 것의 어긋남은 존재하지 않는다.
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorFinalArgv$/view_exec_byte_equality$' -count=1` → exit 0.

### AC-LA-006 — 변이 전 거부 (REQ-LA-004)

- **Given** 카드 워크트리 앵커(기동 디렉터리 또는 `-w` 지명)가 있고 **When** 리더 기동이 거부되면 **Then** 리더 역할 마커가 환경에 새겨지지 않고, run 기록 계수가 불변이며, settings 준비 산물이 없다 — 세션 변이가 전부 판정 뒤에 있음이 세 지표로 관측된다. cc·glm 양 팔.
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorGuardSemantics$/refusal_before_mutation$' -count=1` → exit 0.

### AC-LA-007 — 레인·비팩토리 면제 (REQ-LA-007)

- **Given** 같은 카드 워크트리 앵커가 있고 **When** 레인 기동(`-l`/lane-shape) 또는 팩토리 진입 없는 기동을 하면 **Then** 가드는 발화하지 않고 기동이 진행된다(시험 엔진 시암 관측) — 카드 트리 앵커가 의도인 경로는 오늘과 같다.
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorGuardSemantics$/lane_and_nonfactory_exempt$' -count=1` → exit 0.

### AC-LA-008 — 재정의 + 감사 행 + 기록 실패 거부 (REQ-LA-005)

- **Given** 카드 워크트리 앵커가 있고 **When** `--allow-card-anchor`와 함께 기동하면 **Then** 경고 후 진행되고 감사 행(백엔드·트리·브랜치)이 기록된다. **When** 감사 로그를 기록할 수 없는 상태(경로가 디렉터리 등)에서 같은 기동을 하면 **Then** 기동은 거부된다 — 기록되지 않는 우회는 없다. 플래그가 다른 옵션의 값 자리에 있을 때는 재정의로 인정되지 않는다.
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorGuardSemantics$/override_audit_row$' -count=1` → exit 0.

### AC-LA-009 — 감사 기록 위생 (REQ-LA-009)

- **Given** 감사 로그 경로가 있고 **When** 그 경로가 symlink이거나 FIFO인 상태에서 재정의 기동하면 **Then** 개방 전 거부(Lstat)와 서술자 형식 재판독 거부가 각각 발화한다 — 행 손실 없는 기록 위생이 이식 셀로 묶인다.
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorAuditHygiene$' -count=1` → exit 0.

### AC-LA-010 — fail-open 방향 (REQ-LA-008)

- **Given** 브랜치를 판독할 수 없는 트리(부재·판독 실패)가 `-w`로 지명되고 **When** 리더 기동하면 **Then** 기동은 허용된다 — 거부는 카드 접두의 긍정적 증거 위에서만 한다.
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorGuardSemantics$/fail_open_on_unverifiable$' -count=1` → exit 0.

### AC-LA-011 — 판별자 환경 독립 (REQ-LA-006)

- **Given** 세션 환경에 오래된 라벨(카드 앵커를 암시하는 env)이 있고 **When** 주 체크아웃(main 브랜치)에서 리더 기동하면 **Then** 기동은 허용된다 — 판별은 브랜치 접두뿐이다. 대조: 같은 환경에서 `WT-` 앵커는 거부된다.
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorGuardSemantics$/discriminator_ignores_env$' -count=1` → exit 0.

### AC-LA-012 — `--branch` 생성/재사용 주석 (REQ-LA-004 실효 브랜치 정의)

- **Given** (i) `WT-` 브랜치가 체크아웃된 기존 카드 트리와 (ii) 존재하지 않는 트리 이름이 있고 **When** (i)에 `-w <기존> --branch <다른 브랜치>`로, (ii)에 `-w <신규> --branch WT-<카드>`로 리더 기동하면 **Then** (i)은 그 트리의 실제 체크아웃 브랜치를 판정해 거부하고(플래그 값을 믿지 않는다 — iter6-F2 보존), (ii)는 생성 주석의 브랜치 값을 판정해 materialize 전에 거부한다.
- 검증: `go test ./internal/cli/ -run '^TestLauncherAnchorBranchAnnotation$' -count=1` → exit 0.

### AC-LA-013 — 단일 해석 보장 (REQ-LA-001)

- **Given** 기동 경로 파일 집합(cc/glm 진입, launch, runLaunchClaude)이 있고 **When** 소스 스캔 시험이 해석기 바깥의 worktree 값 재유도 패턴을 세면 **Then** 0건이다 — worktree 선택의 판독은 해석기 한 곳뿐이다.
- 검증: `go test ./internal/cli/ -run '^TestLaunchViewSingleParse$/no_rederivation_on_launch_path$' -count=1` → exit 0.

### AC-LA-014 — 비리더 경로 행동 보존 (§2 제약)

- **Given** 기존 런처 워크트리 규약 시험(`normalizeWorktreeFlag` 철자표, L2 접두 검증, `--branch` materialize, 작성자 사전 점검, lane 재개 어셈블러)이 있고 **When** 단일 해석기 리팩터 뒤 그 패밀리를 실행하면 **Then** 전부 GREEN이다 — 비리더 기동의 관측 행동은 바뀌지 않는다.
- 검증: `go test ./internal/cli/ -run 'Worktree' -count=1` → exit 0 (리팩터가 닿는 패밀리; 세부 선택은 run 증거에 기록).

### AC-LA-015 — RED 관측 기록 (§D.3 폐쇄 게이트)

- **Given** 네 계열 셀과 변형 프로토콜이 있고 **When** run 단계가 종료할 때 **Then** progress.md §E.2에 계열 (a)(b)(c)(d) 각각의 RED 관측(명령·원문 출력·고정 tree SHA)이 존재하고, 돌연변이가 되돌려진 뒤 최종 GREEN 관측과 한 쌍을 이룬다.
- 검증: run 종료 감사가 §E.2의 기록을 판독(기계 열: 네 계열 레이블 + SHA 존재).

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

각 계열 셀은 한 쌍으로 채택된다 — RED-now 셀(교체 전 상태에서의 실패 관측, 명령·원문 출력·고정 tree SHA를 progress.md §E.2에 기록)과 green path 셀(어느 마일스톤이 무엇을 GREEN으로 만드는지). RED의 원인은 명명된다: (a)-(c)는 소비 전 argv 판독 변형(mutant), (d)는 레거시 경로 계수기 변형. 변형은 관측 뒤 되돌린다 — RED가 "올바른 이유"로 적색임을 변형의 정의가 보증한다. `.claude/rules/moai/development/verification-completeness.md` §2 준수.

## §D.4 간접 검증 — argv 철자표 (AC-LA-005의 최소 열)

해석기-동일성 시험은 최소한 다음 철자를 표로 지닌다: `-w <name>`, `--worktree <name>`, `--worktree=<name>`, `-w<name>`(붙임), `-w -b -- <name>`(계열 a), `-w a --worktree --branch b c`(계열 b), `-- -pw<name>`(계열 c), 값 없는 `-w`, 구분자 뒤 `--worktree <name>`, `--worktree=<name>` 뒤 구분자·꼬리. 각 열에서 뷰의 `finalArgv` == exec 시암 포착 argv.

## §D.5 폐쇄 게이트

1. 네 계열 셀 GREEN + RED 기록 한 쌍(AC-LA-015).
2. 보존 의미 여덟 GREEN(AC-LA-006..012, 009 포함).
3. 단일 해석 + 비리더 보존 GREEN(AC-LA-013·014).
4. 크로스 플랫폼 빌드 + 영향 패밀리 회귀 0 + 커버리지 85%+ + lint 신규 0.
5. C-1 백엔드 모형 재측정 원문이 run 증거에 존재.

## §D.6 선행 관측(미래 지향 검사)

- 판정면이 뷰만 음을 재확인하는 소스 스캔(AC-LA-013)은 이후 argv 관련 변경에서도 회귀 감지기로 남는다.
- 계열 셀의 철자표는 새 철자 발견 시 표에 추가하는 확장점이다 — 표가 곧 회귀 감시 목록이다.
