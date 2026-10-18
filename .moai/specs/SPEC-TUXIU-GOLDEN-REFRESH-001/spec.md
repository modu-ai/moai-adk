---
id: SPEC-TUXIU-GOLDEN-REFRESH-001
title: "tuxiu golden fixture refresh — U-08 data-invariance exception, in-repo regeneration harness, postm4 re-capture"
version: "0.1.0"
status: draft
created: 2026-10-19
updated: 2026-10-19
author: manager-spec (lane-11, card t1681)
priority: P1
phase: "v3.2.0"
module: "internal/cli"
lifecycle: spec-anchored
tier: M
depends_on:
  - SPEC-UPDATE-REPAIR-RC30-001
  - SPEC-CLI-TUX-INIT-UPDATE-001
tags: "golden,testdata,regeneration,characterization,data-invariance,test-infra"
---

# SPEC-TUXIU-GOLDEN-REFRESH-001 — tuxiu golden fixture refresh

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-10-19 | manager-spec (lane-11, card t1681) | Initial plan-phase draft (card t1681, run tn1p4e, lane-11, Class C). Tier M — 3 artifacts. U-08 제거 이후 stale 상태로 남은 `internal/cli/testdata/tuxiu/` 골든의 (1) 데이터 불변 계약 예외 확정, (2) 저장소 내 재생성 명령 정의, (3) 재생성 실행을 다룬다. |

## §A Context

### A.1 문제 — 코드는 갱신되었고 골든은 뒤처졌으며, 테스트는 그 차이를 볼 수 없다

U-08(SPEC-UPDATE-REPAIR-RC30-001, REQ-RC30-011)은 `internal/cli/update_archive.go`의 리터럴 `0 user customizations modified`를 제거하고 요약 라인을 실행 시점 측정값만 출력하도록 바꿨다(현재 코드: `total: %d skills archived` 형태의 measured-count-only Pill 출력). AC-RC30-09의 판정 명령은 `--include='*.go'`로 `*.go` 파일만 심사했고, SPEC-UPDATE-REPAIR-RC30-001 progress.md는 그 한계를 기록으로 남겼다 — "The recorded golden fixtures under `internal/cli/testdata/tuxiu/` (pre-logo and `postm4/`, six files) still carry the old literal ... No step names an owner for regenerating these fixtures; make tui-snapshot covers internal/tui only. The leader decides the owner." 이 카드(t1681)가 그 소유자 결정을 받은 카드이다.

본 작업 트리(base `03faddb705`)에서의 실측:

```
$ grep -rn "0 user customizations modified" internal/cli/testdata/tuxiu/
internal/cli/testdata/tuxiu/postm4/update.nocolor.stdout.golden:60
internal/cli/testdata/tuxiu/postm4/update.notty.stdout.golden:60
internal/cli/testdata/tuxiu/postm4/update.tty.stdout.golden:47
internal/cli/testdata/tuxiu/update.nocolor.stdout.golden:50
internal/cli/testdata/tuxiu/update.notty.stdout.golden:50
internal/cli/testdata/tuxiu/update.tty.stdout.golden:37
(6 files — pre-logo 3 + postm4 3)
```

테스트가 이 stale 상태를 잡지 못하는 기제는 두 겹이다. 첫째, 4개 특성화 테스트(`TestInitUpdateTUXCharacterization`, `TestTUXChannelPartition`, `TestTUXDataValuesPreserved`, `TestTUXPostM4CarriesNewPresentation`)는 바이너리를 실행하지 않고 커밋된 골든 파일 내용만 단정한다 — 골든끼리 서로 일치하므로 stale 골든이라도 서로 간에는 일치한다. 둘째, `internal/cli/tuxiu_characterization_test.go:201`의 기대값 `"3 skills archived"`는 stale 리터럴 `"3 skills archived, 0 user customizations modified"`의 접두사이므로 `strings.Contains`가 골든 상태와 무관하게 통과한다(vacuous pass). base에서의 실측:

```
$ go test ./internal/cli -run '^(TestInitUpdateTUXCharacterization|TestTUXChannelPartition|TestTUXDataValuesPreserved|TestTUXPostM4CarriesNewPresentation)$' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	1.252s
(앵커된 셀렉터의 sweep 실측: === RUN 19행 — 상위 4 + 서브테스트 15, 공집합 아님)
```

### A.2 두 기준선 설계와 재기준선(re-baseline)이 막힌 지점

테스트 헤더(`internal/cli/tuxiu_characterization_test.go:1-33`)가 문서화한 설계: `testdata/tuxiu/*.golden`은 M1.0 pre-logo 특성화 기준선("Immutable"로 선언), `testdata/tuxiu/postm4/*.golden`은 post-M4 프레젠테이션 기준선이다. AC-TUXIU-016의 데이터 불변 보장은 "두 커밋된 기준선의 diff"로 기계적으로 검증되며, DATA-line 필터(`tuxDataLines`/`tuxIsPresentation`)는 기대되는 신규 프레젠테이션(로고 아트·아이덴티티 밴드·카드·진행 바)과 타임스탬프·pill-bracket을 정규화에서 떨군다.

여기서 두 기준선을 모두 현재 바이너리로 재캡처하는 재기준선안은 불가능하다: 현재 바이너리는 logo-era 프레젠테이션을 출력하므로, 이것으로 pre-logo 집합을 다시 찍으면 pre-logo 기준선의 존재 이유(로고 이전 프레젠테이션 고정 — `TestTUXPostM4CarriesNewPresentation`이 `iBase`에 `Directories` 표기와 `◆ MoAI-ADK` 부재를 요구)가 소멸한다. 따라서 유일하게 성립하는 경로는 (1) postm4 집합만 현재 바이너리로 재생성하고, (2) pre-logo 집합은 디스크에서 불변으로 유지한 채 U-08 예외를 문서화·정규화 계층에서 처리하는 것이다.

### A.3 측정 근거와 하네스 유실

- 제거 리터럴 보유: 6파일(A.1 실측).
- 4테스트 base PASS(공신력 없는 통과 상태 그 자체): A.1 실측.
- 기존 재생성 경로: `make tui-snapshot` / `make tui-snapshot-verify`(`Makefile:193-198`)는 `internal/tui`만 커버. `tuxiu` 대상 타깃 없음(실측: Makefile 내 `tuxiu-golden-capture` 0회).
- M4 스크래치패드 하네스(`capture.sh` + `pty_capture.py`) 복구 불가: `.moai/worktrees/t1619/` 소멸(실측: 디렉터리 부재), 작업 트리 전수 `find -name pty_capture.py` 0건, `git log --all -- '*pty_capture*' '*capture.sh*'` 0건, `rc30-preview` 워크트리에 scratchpad 없음. → 하네스는 재작성 대상이며, 그 내용을 유추해 기록할 수 없으므로 계획서는 유실을 Gap으로 기재하고 재작성을 요구한다.
- 재작성의 구속 조건은 M4 기록(SPEC-CLI-TUX-INIT-UPDATE-001 progress.md M4)이 남긴다: tty 변형 캡처는 실빌드 바이너리를 강제-트루컬러 pty로 실행해야 한다("`tui.downsample` `sync.Once` 제약은 실바이너리를 요구 — in-process 테스트 불가"). 시딩 환경 규모 토큰(`Updated 26 files`, `3 skills archived`, `Created 2 files`, `13` dirs)은 골든과 테스트에 기록되어 있어 역산 가능하다.
- 골든 파일은 실행 변위 타임스탬프를 내장한다(예: `update-2026-07-24T17-07-46Z`, `20260725_020746`). 따라서 정규화/고정 없이는 재생성이 바이트 안정적이지 않다 — 바이트 안정성 요구(REQ-TGR-005)가 이 처리를 강제한다.
- 코드가 옳고 골든이 stale라는 어느 쪽이 옳은가 판정은 AC-RC30-09의 제거 판정으로 이미 끝났다(카드 지시에 따라 열린 질문이 아닌 확정 전제로 운반한다).

### A.4 기초 질문 Q1-Q4 처리 요약

| 질문 | 처리 | 근거 |
|------|------|------|
| Q1 데이터 불변 예외 방식 | 문서화된 예외 목록 + 정규화 계층 처리 (두 기준선 재기준선 아님) | A.2 — pre-logo 재캡처는 설계상 불가능. 상세는 decision-index Q1(DECIDED) |
| Q2 저장소 내 재생성 명령 | `make tuxiu-golden-capture` — 형식은 tui-snapshot 선례를 기본안으로 | 카드 지시(존재) + decision-index Q2(FOUNDER, 기본안 적용) |
| Q3 골든 수(카드 3 vs 기록 6) | 측정으로 소멸: 리터럴 보유 6파일 중 postm4 3파일이 재생성으로 내용 변경, pre-logo 3파일은 예외 문서화로 유지. 재생성 명령은 postm4 전체 12파일을 대상으로 하고, 나머지 9파일의 바이트 동일성은 run 단계 실측 과제(decision-index Q3) | A.1 실측 + RC30 progress 기록 |
| Q4 어느 쪽이 옳은가 | 코드 옳음, 골든 stale — 확정 전제 | AC-RC30-09 판정(REQ-RC30-011) |

## §B Requirements (GEARS)

### REQ-TGR-001 — AC-016 데이터 불변 계약의 U-08 예외 문서화 (Ubiquitous)

The AC-016 데이터 불변 계약(`internal/cli/tuxiu_characterization_test.go` 헤더 문서) shall — U-08에 따른 archive 요약 라인 변경(`0 user customizations modified` 접미사 제거)이 프레젠테이션 회귀이 아니라 SPEC-UPDATE-REPAIR-RC30-001의 REQ-RC30-011/AC-RC30-09로 재판된 데이터 계약 변경임을 — 예외 항목으로 문서화해야 한다. 예외 문서는 제거 전 리터럴, 제거를 재판한 레코드, 예외가 적용되는 라인을 식별해야 한다.

### REQ-TGR-002 — postm4 골든의 제거 리터럴 부재 (Ubiquitous)

The `internal/cli/testdata/tuxiu/postm4/` 골든 집합 shall — 재생성 이후 — 제거된 리터럴 `0 user customizations modified`를 포함하지 않아야 하며, 현재 바이너리의 measured-count-only 출력(`total: 3 skills archived` 형태)을 기록해야 한다.

### REQ-TGR-003 — M1.0 pre-logo 골든의 불변 유지와 예외 표기 (Ubiquitous)

The M1.0 pre-logo 골든 집합(`internal/cli/testdata/tuxiu/*.golden`, 12파일) shall — 이 SPEC의 실행 전후를 통틀어 — 바이트 단위로 변경되지 않아야 하며, REQ-TGR-001의 예외 문서는 stale 접미사를 여전히 디스크에 지니는 pre-logo 파일 3개(`update.tty.stdout.golden`, `update.notty.stdout.golden`, `update.nocolor.stdout.golden`)를 이름으로 기록해야 한다.

### REQ-TGR-004 — 저장소 내 재생성 명령 (When 재생성 명령이 실행될 때)

**When** 개발자가 `make tuxiu-golden-capture`를 실행할 때, the 저장소 내 재생성 명령 shall — 현재 빌드된 바이너리로부터 `internal/cli/testdata/tuxiu/postm4/` 골든 집합(12파일: init/update × tty/notty/nocolor × stdout/stderr)을 재기록해야 한다. 명령은 자기완결적이어야 한다: 캡처용 throwaway 프로젝트 시딩 레시피를 내장하고, tty 변형은 실빌드 바이너리를 pty로 실행하며(`tui.downsample` `sync.Once` 제약), 저장소의 나머지 부분을 변경하지 않아야 한다(B0 — 저장소 비변경, M4 선례 준수).

### REQ-TGR-005 — 재생성의 바이트 안정성 (While 연속 실행 중)

**While** 개발자가 재생성 명령을 두 번 연속 실행하는 동안, the 재생성 명령 shall — 두 번째 실행에서 `internal/cli/testdata/tuxiu/` 이하에 어떠한 바이트 변화도 만들지 않아야 한다. 실행 변위 타임스탬프 등 비결정 요소는 캡처 시점에 정규화되거나 고정되어야 한다.

### REQ-TGR-006 — 재생성 후 특성화 가드 통과 (When 재생성이 완료되면)

**When** 재생성이 완료되면, the 4개 특성화 테스트(`TestInitUpdateTUXCharacterization`, `TestTUXChannelPartition`, `TestTUXDataValuesPreserved`, `TestTUXPostM4CarriesNewPresentation`) shall — 타이트닝된 기대값으로 — `-count=1` 실행에서 통과해야 한다.

### REQ-TGR-007 — vacuous-pass 재발 차단 (When 제거 리터럴이 골든에 존재할 때)

**When** 어떤 postm4 골든이 제거된 리터럴 `0 user customizations modified`를 포함하도록 조작되면, the 타이트닝된 단정(`internal/cli/tuxiu_characterization_test.go:201`의 접두사 취약한 `strings.Contains`를 대체하는 정확-라인 단정) shall — 해당 골든에서 실패해야 한다. 현재의 접두사 단정은 stale 골든과 fresh 골든을 구별하지 못하며, 이 판별력 상실이 본 결함의 재발 경로이다.

### REQ-TGR-008 — DATA 필터의 예외 목록 처리 (Ubiquitous)

The DATA-line 필터(`tuxDataLines`/`tuxNormLine` 계층) shall — archive 요약 라인의 제거된 접미사(`, N user customizations modified` 형태)를 타임스탬프·pill-bracket 정규화와 같은 계층에서 처리해야 한다. 이 처리의 목적은 REQ-TGR-003의 pre-logo 불변 유지와 양립하면서 REQ-TGR-001의 예외 하에서 pre-logo/postm4 DATA 비교(AC-TUXIU-016)가 성립하게 하는 것이다. archived 스킬 수 자체의 데이터 내용은 예외 대상이 아니며 정확-라인 단정(REQ-TGR-007)으로 계속 보호된다.

## §C Out of Scope

### Out of Scope — 프레젠테이션 코드 변경

- update/init 출력 코드(`internal/cli/update_archive.go`, `internal/cli/update_tux.go` 등)의 변경. AC-RC30-09가 제거를 재판했고 현재 코드가 옳다(Q4 확정 전제). 이 SPEC은 코드를 골든에 맞추지 않는다 — 골든을 코드에 맞춘다.

### Out of Scope — internal/tui 골든

- `internal/tui` 스냅샷 골든. `make tui-snapshot`/`make tui-snapshot-verify`가 이미 소관이다. 이 SPEC은 `internal/cli/testdata/tuxiu/`만 다룬다.

### Out of Scope — 코드-대-골든 상시 검증 게이트

- 재생성 외에 "커밋된 골든이 현재 바이너리 출력과 일치하는가"를 상시 검사하는 verify 모드/CI 게이트 신설. 기본안은 capture-only이다(decision-index Q5 — 기본안 적용). 현행 4테스트는 골든-대-골든 단정이라는 성격을 유지한다.

### Out of Scope — M1.0 pre-logo 집합 재캡처

- pre-logo 골든 12파일의 재캡처. A.2에서 기술했듯 프레젠테이션을 되돌리지 않는 한 불가능하며, 예외 메커니즘(REQ-TGR-001/003/008)이 그 자리를 대신한다.

## §D Success Criteria

성공 기준은 `acceptance.md`의 AC-TGR-001 ~ AC-TGR-008이다. 요지: (1) postm4에서 제거 리터럴 0매치, (2) pre-logo 12파일 바이트 무변경, (3) `make tuxiu-golden-capture` 존재·실행, (4) 재생성 두 번 연속 실행 바이트 안정, (5) 재생성 후 4테스트 PASS, (6) mutant 조작 골든에 대한 타이트닝 단정 FAIL, (7) 재생성 diff가 `internal/cli/testdata/tuxiu/postm4/` 이하로 한정, (8) 예외 문서의 근거 레코드 인용.

## §E Dependencies, Adjacent Work, and Prior Art

- **SPEC-UPDATE-REPAIR-RC30-001** (선행, depends_on): U-08/REQ-RC30-011/AC-RC30-09 — 제거 판정과 "소유자 미지정" 기록(progress.md). 본 SPEC은 그 소유자 결정을 수행한다.
- **SPEC-CLI-TUX-INIT-UPDATE-001** (선행, depends_on): AC-TUXIU-016 가드와 두 기준선 설계, M4 스크래치패드 하네스 기록(실빌드+pty 제약). 하네스 자체는 유실되어 재작성 대상(A.3).
- **SPEC-AC-BASELINE-REFRESH-001** (인접 선례): 다른 도메인(AC-count corpus)에서 "in-tree regeneration mode"를 정한 선례. 형식 참조이며 의존 아님.
- **make tui-snapshot** (선례): 환경변수 게이트형 go test + make 타깃 골든 재생성의 기존 패턴(decision-index Q2 기본안의 근거).
