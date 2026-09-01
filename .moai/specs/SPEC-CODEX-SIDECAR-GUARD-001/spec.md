---
id: SPEC-CODEX-SIDECAR-GUARD-001
title: "init --agent 경로 테스트의 trust sidecar 단언 결손 봉합 (존재·부재 양방향)"
version: "0.1.3"
status: completed
created: 2026-09-01
updated: 2026-09-01
author: manager-spec
priority: P2
phase: "v3.1.5 target"
module: internal/cli
lifecycle: spec-anchored
tier: S
tags: "codex, wiring, sidecar, test-assertion, isolating-mutant, init-agent-flag, t405"
related_specs: [SPEC-CODEX-WIRING-001]
---

# SPEC-CODEX-SIDECAR-GUARD-001 — init `--agent` 경로 테스트의 sidecar 단언 결손 봉합

> 카드: **t405** · 선행 SPEC: `SPEC-CODEX-WIRING-001`(REQ-CW-001 / AC-CW-004 / AC-CW-005 소유)

## HISTORY

| 버전 | 날짜 | 변경 |
|---|---|---|
| 0.1.0 | 2026-09-01 | 최초 작성 (plan-phase, 카드 t405). 카드가 지목한 `:109` 결손에 더해, 재측정 중 발견된 거울상 `:82` 결손을 별개 요구로 분리 기재 |
| 0.1.1 | 2026-09-01 | 리드 수정 라운드 D1-D4 반영. D1: AC-CSG-003의 grep 패턴에서 `\.codex/agents` 대안 제거(그 대안은 `:95`의 산문 주석에만 매치돼 기준 트리에서 히트 1건 → *Then*의 "0건"과 모순) + 이 AC가 결함 탐지가 아닌 금지 제약임을 명시. D3: §F 수용 기준 범위 오프바이원 정정(AC-CSG-007 → AC-CSG-008). D4: `phase` 값의 출처를 plan.md §I에 미검증으로 기록. D2(progress.md 신설)는 별도 파일 |
| 0.1.2 | 2026-09-01 | 리드 최종 라운드. (1) `progress.md`의 `§F.1` 지시 철회 확인 — `§E.1` 유지, 변경 없음(era.go는 `§E.2`/`§E.4`/`§E.5`만 grep하고 스키마 § progress.md Section Map이 `§E.1`을 manager-spec 소유로 명시). (2) Tier S 산출물 규정(2종) 대비 `acceptance.md` 1종 초과가 의도적 오케스트레이터 결정임을 plan.md §J에 기록 — 접지 않음, 커밋 주체는 3종으로 표기 |
| 0.1.3 | 2026-09-01 | plan-audit 수리 라운드(PASS-WITH-DEBT 0.82, blocking 2건). D1: AC-CSG-008에 시험별 판정 명령(`go test ./internal/cli/ -run 'TestRunInit_Agent' -v`) 추가 — 비상세 실행의 `ok` 한 줄로는 네 시험 중 하나의 삭제를 구별 못 하는(§A.4 공허 초록) 자기 판정 불능 결함 봉합. D2: AC-CSG-004 부재 방향 뮤턴트를 컴파일 가능한 임시 2편집 형태로 재작성 — `writeSidecar`가 타 패키지 비exported라 직접 형태는 컴파일 불가, 임시 export 래퍼(`WriteSidecarOnly`) + `claude` 분기 호출로 심고 AC-CSG-007 검사로 원복 검증 |

## §A. 측정 전제 (Verified baseline)

> 아래 사실은 전부 **본 워크트리의 커밋 `64bba61aa`에서 실측**되었다. run-phase에서 이를
> 재확립하기 위해 다시 잴 필요는 없으나, 인용을 위해 해당 파일을 다시 읽는 것은 허용된다.

### §A.1 배선 산출물 3종

`internal/codexwiring/codexwiring.go`는 프로젝트 루트 상대 경로 3개를 선언한다.

| 좌표 | 상수 | 값 |
|---|---|---|
| `internal/codexwiring/codexwiring.go:29` | `HooksRelPath` | `.codex/hooks.json` |
| `internal/codexwiring/codexwiring.go:31` | `ConfigRelPath` | `.codex/config.toml` |
| `internal/codexwiring/codexwiring.go:34` | `SidecarPath` | `.moai/state/codex-wiring.json` |

`internal/codexwiring/wire.go:187`의 `writeSidecar`가 `filepath.Join(projectRoot, SidecarPath)`를
쓰며, `Wire`에서 도달한다. 즉 sidecar는 hooks.json·config.toml과 **동일한 조건에서 함께 생성되는
세 번째 산출물**이지 부수적 캐시가 아니다.

### §A.2 게이팅 지점

`internal/cli/init.go:167-170`의 `wireCodexUnlessClaude`는 해소된 배선 값이 `claude`이면 조기
반환하고, 그렇지 않으면 `codexwiring.Wire`를 호출한다. 따라서:

- `--agent codex` · `--agent both` → 산출물 **3종 전부** 기록
- `--agent claude` · 플래그 부재 → 산출물 **3종 전부** 미기록

### §A.3 시험 4종의 단언 강도 — 양방향으로 어긋나 있다

`internal/cli/init_agent_flag_test.go`의 `runInit` 기반 경로-집합 시험 4종:

| 좌표 | 시험 | 모드 | 단언 경로 수 | 방향 |
|---|---|---|---|---|
| `:70` | `TestRunInit_AgentCodexWiresAndSkipsMCPProvisioning` | codex | 3 (sidecar 포함) | 존재 |
| `:82` | `TestRunInit_AgentBothWiresBothSides` | both | **2 — sidecar 누락** | 존재 |
| `:97` | `TestRunInit_AgentAbsentLeavesNoCodexFiles` | 부재 | 3 (sidecar 포함) | 부재 |
| `:109` | `TestRunInit_AgentClaudeLeavesNoCodexFiles` | claude | **2 — sidecar 누락** | 부재 |

같은 방향의 형제끼리 강도가 다르다: `:97`은 sidecar 부재까지 지키는데 `:109`는 지키지 않고,
`:70`은 sidecar 존재까지 지키는데 `:82`는 지키지 않는다. **카드 t405가 지목한 것은 `:109`
하나뿐이며, `:82`는 본 카드의 재측정 과정에서 발견된 거울상이다.**

### §A.4 결손이 조용한 이유

두 결손 모두 **적색을 만들지 않는다**. 빠진 단언은 실패하지 않고 단지 아무것도 보지 않으므로,
`--agent claude`가 sidecar를 흘리기 시작하거나 `--agent both`가 sidecar를 빠뜨리기 시작해도
시험 4종은 전부 초록으로 통과한다. 결손의 비용은 오검출이 아니라 **미검출**이다.

## §B. 요구 (GEARS)

### REQ-CSG-001 — claude 경로의 sidecar 부재 단언 (핵심 범위)

**When** 시험 `TestRunInit_AgentClaudeLeavesNoCodexFiles`(`internal/cli/init_agent_flag_test.go:109`)가
`--agent claude`로 초기화한 프로젝트를 검사할 때, 해당 시험은 `.moai/state/codex-wiring.json`의
부재를 개별 경로 단언으로 검사해야 한다(shall). 검사 후 이 시험의 경로 집합은 형제 시험
`:97`(플래그 부재)의 경로 집합과 동일해야 한다.

### REQ-CSG-002 — both 경로의 sidecar 존재 단언 (거울상 범위)

**When** 시험 `TestRunInit_AgentBothWiresBothSides`(`internal/cli/init_agent_flag_test.go:82`)가
`--agent both`로 초기화한 프로젝트를 검사할 때, 해당 시험은 `.moai/state/codex-wiring.json`의
존재를 개별 경로 단언으로 검사해야 한다(shall). 검사 후 이 시험의 경로 집합은 형제 시험
`:70`(codex)의 경로 집합과 동일해야 한다.

> 이 요구는 카드 t405의 본문이 아니라 **리드 판정**으로 범위에 들어왔다. 경위와 근거는 §E.1.

### REQ-CSG-003 — 디렉터리 단언 금지 (금지 요구)

부재 방향의 어떤 시험도 `.codex/` **디렉터리**의 부재를 단언해서는 안 된다(shall not). 템플릿
트리가 `.codex/agents/**`를 배포하므로 배선이 전혀 기록되지 않은 초기화에서도 그 디렉터리는
정당하게 존재한다. 부재는 **파일 경로 단위로만** 단언한다.

### REQ-CSG-004 — 격리 뮤턴트로 적색 능력 입증

**When** 본 SPEC이 새 단언을 추가할 때, 각 새 단언은 **격리 뮤턴트**로 적색이 될 수 있음을
입증해야 한다(shall). 격리 뮤턴트란 그 시험에 **기존부터 있던 두 단언을 초록으로 남기는**
변형이며, 따라서 적색을 만드는 것이 새 단언 하나임이 드러난다. 기존 단언까지 함께 깨뜨리는
변형은 새 단언이 제 몫을 한다는 것을 입증하지 못하므로 이 요구를 충족하지 않는다.

**Where** 뮤턴트가 트리에 심어진 상태, 해당 뮤턴트는 관측이 끝나는 즉시 되돌려져야 한다
(shall) — 뮤턴트는 관측 도구이지 산출물이 아니다.

### REQ-CSG-005 — 검증 범위는 건드린 패키지로 한정

본 SPEC의 지역 검증은 `go test ./internal/cli/... ./internal/codexwiring/...`로 한정해야 한다
(shall). 기준별 판정 행(per-test verdict)을 얻기 위한 상세 형태 — 같은 `internal/cli` 패키지
범위를 `-run` 필터와 `-v`로 좁힌 `go test ./internal/cli/ -run 'TestRunInit_Agent' -v` — 는
이 한정 안의 동일 범위 실행으로 본다. 전체 스위트 지역 실행은 금지한다(shall not) — 부하가 걸린 개발 머신에서의 전량 실행은
코드가 아니라 기계를 재는 것이며, 전 패키지 판정은 PR head에 대한 CI의 몫이다.

## §C. 제약

| # | 제약 | 근거 |
|---|---|---|
| C1 | 프로덕션 코드(`internal/cli/init.go`, `internal/codexwiring/**`)의 **영구 변경 0줄**. 변경은 시험 파일 한 개에 국한된다 | 본 SPEC은 계약을 바꾸지 않고 이미 존재하는 계약의 관측 결손만 메운다 |
| C2 | 새 단언은 기존 단언의 표현 방식(경로 슬라이스 순회 + `os.Stat`)을 따른다 | 파일 내부 일관성 |
| C3 | 뮤턴트는 커밋되지 않는다 | REQ-CSG-004 |

## §D. 범위 밖 (Exclusions)

### Out of Scope — 프로덕션 동작 변경

- `wireCodexUnlessClaude`의 게이팅 논리 변경
- `writeSidecar` 호출 위치·조건 변경
- sidecar 스키마(`sidecarDoc`) 변경

### Out of Scope — 다른 검사 표면

- `moai doctor`의 sidecar 판독 경로에 대한 시험 추가
- `moai update` 재실행 경로(멱등)의 sidecar 단언 — SPEC-CODEX-WIRING-001 AC-CW-006 소관
- wizard(대화형) 경로의 `--agent` 해소 — 카드 t393 소관(§E)

### Out of Scope — 세 번째 이후 표면

- `internal/cli/init_agent_flag_test.go`의 `:82`·`:109` 외 어떤 시험의 단언 강도 조정
- 작업 중 발견되는 추가 거울상의 흡수 — 새 카드로 발행하고 리드에게 올린다(§E.2)

### Out of Scope — 도구·프로세스

- 전체 테스트 스위트의 단언 강도 감사
- sidecar 단언 결손을 기계적으로 잡는 린트/가드 신설

## §E. 범위 경위와 의존

### §E.1 왜 이 SPEC은 카드 t405보다 넓은가 — 리드 판정 기록

**카드 t405의 본문이 지목한 결손은 `:109` 하나뿐이다.** `:82`는 카드에 없었고, 본 카드의
plan-phase 재측정 과정에서 발견된 뒤 **리드 판정으로 범위에 편입**되었다. 즉 이 SPEC이 카드보다
넓은 것은 작성자의 재량 확장이 아니라 기록된 판정의 결과다.

**판정을 정당화한 근거 — 추론이 아니라 측정.**

1. **4종 단언 강도 분포**(트리 `64bba61aa` 실측, §A.3 표): `:70`/`:97`은 sidecar를 포함한 3경로,
   `:82`/`:109`는 2경로. 결손은 한 방향의 사고가 아니라 **존재·부재 양방향에 각각 하나씩** 있다.
2. **`--agent both`가 실제로 sidecar를 쓴다는 코드 경로**: `internal/cli/init.go:167-170`의
   `wireCodexUnlessClaude`는 해소값이 `claude`일 때만 조기 반환하고, 그 외에는 `codexwiring.Wire`를
   호출한다. `Wire`는 `internal/codexwiring/wire.go:187`의 `writeSidecar`에 도달한다. 따라서
   `:82`가 sidecar를 단언하지 않는 것은 "both에서는 sidecar가 안 나오니 단언할 게 없다"가 아니라
   **나오는 산출물을 보지 않고 있는 것**이다.

**판정의 논거**: 같은 파일, 같은 결함 계열이며, 한쪽만 닫으면 남은 거울상이 다음 감사에서 다시
떠오른다. 두 줄을 함께 닫는 편이 두 번 여는 것보다 싸다.

### §E.2 확장의 상한 — 여기서 멈춘다

본 SPEC의 확장은 `internal/cli/init_agent_flag_test.go`의 **실측된 두 줄(`:82`, `:109`)에서
끝난다**. 이는 범위 서술이 아니라 제약이다:

- **프로덕션 코드 변경 없음.** `internal/codexwiring/**`와 `internal/cli/init.go`는 본 SPEC에서
  읽기 전용이다. 유일한 예외는 REQ-CSG-004의 일시적 뮤턴트이며, 그것은 관측 후 반드시 되돌린다.
- **세 번째·네 번째 표면으로 넓히지 않는다.** 작업 중 또 다른 거울상을 발견하더라도 본 SPEC에
  흡수하지 않는다 — 새 카드로 발행하고 리드에게 올린다.

### §E.3 의존 (기록만, 행동하지 않음)

카드 **t393**은 아직 착지하지 않았다. 착지하면 그 `AC-IHP-006a`가 wizard 경로를 덮으므로,
남는 결손 범위는 **그 시점에 다시 측정**해야 한다. 본 SPEC은 t393에 게이팅되지 않으며,
plan-phase에서 그 잔여 범위를 지금 측정하지 않는다(후속 메모: plan.md §F).

## §F. 추적

| 항목 | 값 |
|---|---|
| 카드 | t405 |
| 선행 SPEC | SPEC-CODEX-WIRING-001 |
| 소비하는 선행 요구 | REQ-CW-001 (플래그 폐쇄집합 + claude≡플래그 부재 등가) |
| 소비하는 선행 AC | AC-CW-004 (플래그 부재 시 배선 미생성) · AC-CW-005 (both 시 배선 생성) |
| 수용 기준 | `acceptance.md` (AC-CSG-001 … AC-CSG-008) |

`:109`와 `:97`은 REQ-CW-001의 "claude ≡ 플래그 부재" 등가를 지키는 한 쌍이고, `:82`와 `:70`은
AC-CW-005의 both 의미론을 지키는 한 쌍이다. 본 SPEC은 두 쌍 각각의 강도를 맞춘다.
