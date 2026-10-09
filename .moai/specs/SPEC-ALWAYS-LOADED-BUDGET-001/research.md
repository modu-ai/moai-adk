---
id: SPEC-ALWAYS-LOADED-BUDGET-001
title: "research — 배포 표면 실측"
version: "0.1.0"
created: 2026-10-03
updated: 2026-10-04
---

# research.md — SPEC-ALWAYS-LOADED-BUDGET-001

§1~§5 의 수치는 plan 첫 실행에서 템플릿 바이트가 `b5815ca80`(당시 로컬 `develop`)인 워크트리에서 잰 값이다. §1.1 은 로컬 develop `d7112d005` 를, §1.2·§4·§5 는 흡수 트리 `2771626b5`(2026-10-04 재측정·재실행)를 각각 다시 잰 값이다. 계수 단위는 UTF-16 코드 단위(`len(s.encode('utf-16-le'))//2`), 선택 규칙은 `instructions_loaded.go` `ruleFileAlwaysLoaded` 와 같다(첫 줄 `---` 블록 안에 최상위 `paths:` 가 없으면 상시 로드).

## §1. 템플릿 원본 기준 상시 표면

측정 대상은 `internal/template/templates/.claude/rules/moai/**/*.md` 와 `internal/template/templates/CLAUDE.md`. 렌더링 전 원본이므로 `AGENTS.md.tmpl`(20,874)·`user.yaml.tmpl`(224)·`language.yaml.tmpl`(1,457)은 렌더링 결과와 조금 다르다(원본 합 22,555, 리드의 렌더링 실측 22,368).

| 파일 | 크기 | 구속 줄 | 구속 줄 자수 | `[HARD]` 줄 |
|---|---:|---:|---:|---:|
| workflow/kanban-dispatch.md | 27,771 | 38 | 16,182 | 38 |
| core/askuser-protocol.md | 17,872 | 24 | 5,250 | 10 |
| core/agent-common-protocol.md | 17,612 | 20 | 5,189 | 15 |
| workflow/session-handoff.md | 15,833 | 11 | 5,583 | 11 |
| core/verification-claim-integrity.md | 15,766 | 17 | 7,883 | 10 |
| core/moai-constitution.md | 15,529 | 16 | 2,275 | 12 |
| workflow/cross-session-messaging.md | 11,325 | 6 | 1,903 | 6 |
| workflow/context-window-management.md | 7,047 | 7 | 1,009 | 5 |
| workflow/main-checkout-branch-guard.md | 6,195 | 3 | 655 | 2 |
| workflow/goal-directive.md | 6,130 | 0 | 0 | 0 |
| workflow/cache-aware-execution.md | 5,925 | 5 | 1,665 | 5 |
| core/moai-mcp-tools.md | 4,312 | 1 | 81 | 1 |
| core/native-idiom-and-register.md | 2,244 | 2 | 840 | 2 |
| **규칙 13개 합** | **153,561** | | | |
| CLAUDE.md | 15,457 | 4 | 3,492 | 3 |
| AGENTS.md.tmpl (원본) | 20,874 | 5 | 482 | 0 |

- 규칙 13개 합 153,561, `CLAUDE.md` 15,457 — 리드 실측과 정확히 일치한다. 렌더링 import 22,368 을 더한 191,386 이 기준선이다.
- 파일당 40,000 초과: `workflow/kanban-dispatch-detail.md` **43,138**(`paths:` 한정) 하나.
- 구속 줄 정의: `[HARD]` · `MUST` · `shall ` 포함 줄(`SPEC-ALWAYS-LOADED-DIET-002` §B). 구속 줄 자수는 줄 길이 + 줄바꿈 1.

## §1.1 기준선은 이미 움직였다 — 로컬 develop `d7112d005` 재측정

같은 계수 규칙으로 로컬 `develop` 팁 `d7112d005` 을 쟀다(`git archive d7112d005 internal/template/templates/.claude/rules/moai internal/template/templates/CLAUDE.md` 를 스크래치에 풀어 측정).

- 상시 규칙 13개 155,040, `CLAUDE.md` 15,553, 렌더 독립 소계 **170,593**(`b5815ca80` 의 169,018 대비 +1,575). 40,000 초과는 여전히 `kanban-dispatch-detail.md` 43,138 하나.
- `git diff --stat HEAD develop -- <규칙·CLAUDE.md·AGENTS.md.tmpl>` 은 8파일(상시 규칙 `agent-common-protocol.md`·`native-idiom-and-register.md`·`cache-aware-execution.md` 포함)의 이동을 보였다.
- 그래서 AC 기대값은 고정 숫자로 두지 않고 원장 앵커에서 다시 잰 값을 쓴다(`plan.md` §C.2 4단계). 감사 2회차가 인용한 codex 측정 169,291 은 다른 develop 팁(`7142018fb`)의 값이며 이 실행에서 재측정하지 않았다.

## §1.2 흡수 뒤 재측정 — 로컬 develop 흡수 커밋 `30ce3a02d` 를 포함한 트리 `2771626b5`

카드 t1399(`SPEC-LAUNCHER-ENTRY-FLAGS-001`, 흡수에 담긴 커밋 `859477bc7` M10·`1d0c64ef4` M5b)가 배포 표면의 이름을 바꿨다: `workflow/kanban-dispatch.md` → `workflow/factory-dispatch.md`, `kanban-dispatch-detail.md` → `factory-dispatch-detail.md`, 새 companion `factory-dispatch-mechanics.md`(최상위 `paths:`, UTF-16 17,107 — 상시 표면 밖). §1·§1.1 의 `kanban-dispatch*` 표기는 측정 당시 파일명이다. 같은 흡수로 `internal/hook/session_start_kanban.go` 는 지워졌고, 스킬 `moai-kanban-foreman` 은 `moai-factory-foreman` 으로 개명됐다. §1 과 같은 UTF-16 계수기를 이 실행에서 다시 돌렸다.

- 상시 규칙 13개 154,652, `CLAUDE.md` 15,541, 렌더 독립 소계 **170,193**(§1.1 `d7112d005` 170,593 대비 −400). `@`-import 원본 합 23,541(`AGENTS.md.tmpl` 21,860·`user.yaml.tmpl` 224·`language.yaml.tmpl` 1,457 — `user`·`language` 템플릿은 `templates/.moai/config/sections/` 에 있다).
- 크기 상위: `workflow/factory-dispatch.md` 27,423 · `core/agent-common-protocol.md` 18,421 · `core/askuser-protocol.md` 17,872 · `workflow/session-handoff.md` 15,833 · `core/verification-claim-integrity.md` 15,766 · `core/moai-constitution.md` 15,510 · `workflow/cross-session-messaging.md` 11,321.
- 40,000 초과는 **둘**이다: `workflow/factory-dispatch-detail.md` **40,659**(t1399 분할로 §1 의 43,138 보다 줄었으나 아직 초과)와 `workflow/spec-workflow.md` **40,052**(§1.1 시점에는 초과가 아니었다 — 흡수 사이에 초과에 들어옴). REQ-ALB-003 은 둘 다 걸며, M3 분할 대상은 원장 앵커 재측정이 확정한다(`plan.md` M3).
- 진입점 표면: 무표지 리더급 진입점은 `manager-lead` 에이전트(deputy 포함)·`moai-factory-foreman` 스킬·`gtd` 워크플로 `--auto` 로 유지된다. 스킬 본문은 `.claude/rules/moai/workflow/factory-dispatch.md` 를 인용한다 — 표지·Read 지시는 이 SPEC 의 M2 몫이다.

## §1.3 축소 뒤 목표 대조 — 이슈 목표 150,000 (2026-10-04 재측정, Q10(a))

카드가 인용한 이슈 #1746·#1717 의 목표는 설치 규칙(installed rules) 150,000자 이하다. §1.2 와 같은 계수기(UTF-16 코드 단위, `ruleFileAlwaysLoaded` 선택)를 트리 `2771626b5`(카드 브랜치 팁에서 흡수 뒤 템플릿 바이트 동일 확인)에서 이 실행에 다시 돌렸다(`.moai/state/verify/t1469/measure_always_loaded.py`, 기계-local 스크래치 — 명령·출력 전문은 2026-10-04 manager-spec 실행 기록):

- 상시 규칙 13개 합 **154,652** — 목표 150,000 을 **4,652 초과**다. `CLAUDE.md` 15,541 은 규칙 계수에서 제외(렌더 독립 소계 170,193, §1.2 와 일치). 상시 표면에서 40,000 초과 파일은 0개다.
- 축소가 150,000 메커니즘을 훼손하지 않는다 — 이 합계를 끌어내리는 소관 행이 전부 남아 있다: REQ-ALB-001·002(예산 테스트와 115,000 총합 상한), REQ-ALB-006·007·025(역할 한정 stub·주입 — 최대 단일 감축 레버), REQ-ALB-012·015·017(구속 원장과 의미 보존 압축 재작성), REQ-ALB-022(M0 하한 중단점), plan M0·M3(압축 마일스톤). 제거된 `factory-dispatch-detail.md` 분할은 이 합계에 기여하지 않았다 — 그 파일은 최상위 `paths:` 를 가져 상시 표면 밖이므로 그 분할이 상시 로드 합계를 움직이지 않는다. `spec-workflow.md`(40,052)도 `paths:` companion 으로 상시 합계 밖이며 리드가 별도 처분한다.
- §2 가설을 현재 수치로 환산하면: 역할 한정 2개의 stub 전환만으로 상시 규칙 합은 약 154,652 − 27,423 − 11,321 + stub(§2 추정 약 4,500) ≈ **120,400** — 목표보다 약 29,600 아래다. 이후 압축 재작성은 115,000 총합(규칙+`CLAUDE.md`)을 향해 더 내려간다. §2 의 원 가설(115,000 하한이 닿지 않을 가능성)이 그대로 성립해도, 그때의 Q6 중단 경로는 예산 상수 115,000 에 관한 것이지 규칙 150,000 목표에 관한 것이 아니므로 이 목표는 유지된다. M0 실측이 가설을 대체한다(REQ-ALB-022).
- **결론 한 줄**: Q10(a) 축소 뒤에도 남은 범위가 ≤150,000 목표를 소관한다 — 예. 현재 트리는 154,652 로 4,652 초과이며, 역할 한정 주입(M2)만으로도 목표 아래로 내려간다.

## §2. 도달 가능성 개산 — M0 가 확정할 가설

아래는 **가설**이며 M0 의 측정이 대체한다.

- 역할 한정 2개를 stub 으로 바꾸면 191,386 − 27,771 − 11,325 + stub(약 4,500, cross-session 의 비역할 구속 줄 1,903 포함) ≈ **156,800**.
- 예산 115,000 까지 남는 감축량 ≈ **41,800**. 나머지 상시 파일 11개 + `CLAUDE.md` 의 비구속 본문은 약 96,000(전체 129,800 − 구속 줄 약 33,900).
- 즉 비구속 본문의 약 44% 를 companion 으로 옮겨야 한다. `SPEC-ALWAYS-LOADED-HEADROOM-001` 은 더 엄격한 허용 조건(구속 줄이 기대는 표·제목·범위 한정 문장은 옮기지 않음) 아래에서 18경로 전체의 허용 제거 풀을 약 14,000 으로 쟀다. 그 조건을 그대로 승계하면 115,000 은 닿지 않을 가능성이 높다. 이 위험 때문에 M0 를 실행 첫 마일스톤이자 중단 지점으로 둔다(REQ-ALB-022).

## §3. 예산 상수 115,000 의 근거 — 비다이어트 기간의 표면 증가 실측

같은 계수 규칙으로 `develop` first-parent 이력의 네 지점을 쟀다(`git archive` 로 스크래치에 풀어 측정).

| 기준일(직전 커밋) | 커밋 | 규칙 수 | 규칙 합 | CLAUDE.md | AGENTS.md.tmpl | 합(원본) |
|---|---|---:|---:|---:|---:|---:|
| 2026-09-20 | `b1ec8602c` | 14 | 207,221 | 19,267 | 17,907 | 244,395 |
| 2026-09-27 | `b69cfe7ec` | 14 | 210,591 | 19,367 | 19,106 | 249,064 |
| 2026-10-01 | `f4aa9bf99` | 13 | 151,000 | 15,372 | 20,893 | 187,265 |
| 2026-10-03 | `b5815ca80` | 13 | 153,561 | 15,457 | 20,874 | 189,892 |

- 다이어트가 없던 두 구간의 증가: 09-20→09-27 **+4,669**(7일), 10-01→10-03 **+2,627**(2일). 09-27→10-01 구간은 다이어트 착지 구간이라 증가율 산정에서 뺀다.
- 120,000(200K 모델 한도, 카드 인용) 아래 여유 5,000 은 관측된 최대 주간 증가(+4,669)를 한 번 덮는다. 가드는 CI 에서 먼저 깨지고, 사용자가 런타임 한도를 넘기 전에 다이어트 카드가 생긴다. 그래서 **115,000**.
- 한계: 증가 표본은 두 구간뿐이다. 120,000·150,000 한도 수치 자체는 카드·이슈에서 인용했고 이 실행에서 공식 문서로 재확인하지 않았다.

## §4. 경로를 고정한 소비자

`factory-dispatch`(구 `kanban-dispatch`) 또는 `cross-session-messaging` 을 이름으로 참조하는 Go 파일(템플릿 트리 제외)이 흡수 뒤에도 확인된다 — `factory-dispatch` 12개, `cross-session-messaging` 6개(2026-10-04 측정). 예: `internal/template/lane_recheck_doctrine_test.go`, `internal/template/workflow_rule_paths_pinned_test.go`(둘 다 `factory-dispatch` 경로를 고정). 본문을 다른 경로로 옮기면 이들이 함께 움직여야 하므로, 전체 본문은 기존 경로에 둔다(`spec.md` §D).

## §5. 기존 SessionStart 주입 표면

- `internal/hook/session_start_factory.go` — 팩토리 리더·레인 공지를 `additionalContext` 로 주입한다. 부트스트랩 공지는 source `startup` 에서만, 레인 규칙은 `startup`·`clear` 에서 나간다. `compact`·`resume` 에는 아무것도 주입하지 않는다.
- `internal/hook/session_start_kanban.go` — t1399 M5b(흡수 커밋 `1d0c64ef4`)에서 제거됐다. 칸반 모드가 폐지되어 SessionStart 공지 생산자는 factory 쪽만 남는다.
- 남은 SessionStart 생산자는 역할 한정 규칙 본문을 주입하지 않는다 — `grep -l -e role-core -e factory-dispatch internal/hook/session_start_factory.go internal/hook/factory_messages.go internal/hook/subagent_start.go` 출력 없음, exit 1(2026-10-04 재실행).
- 역할 표지는 `MOAI_FACTORY_WORKER`·`MOAI_FACTORY_WORKERS`·`MOAI_KANBAN_*` 계열 환경 변수다. 카드 t1399(런처 진입 `-f`/`-l`, 칸반 모드 제거)가 진행 중이라 착지 시점의 표지 집합은 바뀔 수 있다(`decision-index.md` Q8).
- `compact` 뒤에는 앞서 주입한 맥락이 요약으로 대체된다. 상시 규칙은 다시 실리지만 주입 맥락은 그렇지 않으므로, 역할 core 는 `compact` 에서도 다시 주입해야 한다(REQ-ALB-007).
