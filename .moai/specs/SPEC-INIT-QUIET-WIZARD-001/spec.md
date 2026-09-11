---
id: SPEC-INIT-QUIET-WIZARD-001
title: "moai init 위저드 정온화 — 질문 18개를 4개로"
version: "0.1.4"
status: draft
created: 2026-09-11
updated: 2026-09-11
author: GOOS
priority: P1
phase: "v3.2.0 target"
module: "internal/cli/wizard, internal/cli"
lifecycle: spec-anchored
tags: "init, wizard, quiet-init, defaults, mcp-provision, shell-seam, home-safety, t583"
tier: L
related_specs: [SPEC-CLI-WIZARD-RESTRUCTURE-001, SPEC-V3R5-INIT-WIZARD-EXPANSION-001, SPEC-INIT-WIZARD-REPAIR-001, SPEC-INIT-HARNESS-PROMPT-001, SPEC-MCP-DEFAULT-ON-001, SPEC-MOAI-MCP-SERVER-001, SPEC-TODO-ENABLE-FLAG-001, SPEC-FEEDBACK-AUTO-SUBMIT-001, SPEC-PROJECT-CONTINUATION-KEY-001]
---

## HISTORY

- 2026-09-11 — v0.1.0 초안 (manager-spec, 카드 t583 plan 단계). 근거: `.moai/reports/t583/verdict.md`(F1 재현·F4 반증·§7 리드 결정), `.moai/reports/t583/research-explore.md`(질문 표면 조사), `.moai/reports/t583/audit-excerpt.md`(원 감사 발췌와 운영자 결정). 모든 file:line 은 워크트리 HEAD `120436f58` 에서 다시 확인했다(research.md §0).
- 2026-09-11 — v0.1.1 개정 (manager-spec, 리드 결정 (a) 반영). plan.md §L 의 "기존 테스트 소급 범위" 확인 항목을 해소했다. §4.2 홈 안전 점검표는 이 SPEC 이 새로 쓰거나 본문을 다시 쓰는 init 실행 테스트의 코드에만 요구하고, 제거된 필드 참조만 지우는 기존 테스트는 기존 `t.Setenv("HOME")` 헬퍼를 유지한다. 대신 기존·신규를 가리지 않고 init 실행 테스트를 돌리는 모든 run 슬롯에 실제 홈 지문 절차를 새로 요구했다(REQ-IQW-014, §4.3, AC-IQW-015). 대조 대상 셸 설정 파일을 5개에서 6개로 늘렸다(`~/.bash_profile` 추가). 기존 HOME 헬퍼의 시접 기반 이관은 범위 밖 후속 후보로 적었다(§8).
- 2026-09-11 — v0.1.2 개정 (manager-spec, plan-audit 1회차 FAIL 0.79 반영 — `.moai/reports/t583/plan-audit.md` D1~D8). D1 은 리드 결정으로 방향을 정했다: 셸 설정 단계를 `internal/core/project` 의 교체 가능한 테스트 시접 하나를 거쳐 실행하게 하고, 스파이를 끼운 실제 `runInit` 실행으로 단계 도달을 관측한다. 이 시접 하나가 셸 설정 쓰기 억제까지 맡으므로 `internal/cli` 쪽 억제 시접은 두지 않는다(REQ-IQW-011 개정, AC-IQW-004 재작성). REQ-IQW-012 를 대조 8항목으로 좁혔고(D2), 새 테스트를 겨누는 선택자에 스윕 확인을 넣었다(D3). 새 프로젝트 디스크 모양 동일성을 REQ-IQW-015 로 올리고 AC-IQW-008 을 거기에 매핑했다(D4). §6 의 `audit:` 서술을 `model`·`gates` 하위 키 기준으로 정정했다(D5). REQ-IQW-003 에 파일시스템 루트 예외를 넣었다(D6). REQ-IQW-009 의 절차를 AC 로 옮기고 REQ-IQW-014 를 채집·기록(014)과 멈춤·보고(REQ-IQW-016)로 나눴다(D7). AC-IQW-015 가 슬롯 선언 수·지문 파일 수·기록 수 삼자 일치를 판정하게 했다(D8).
- 2026-09-11 — v0.1.3 개정 (manager-spec, plan-audit 2회차 전 리드 조건 반영). (1) 리드가 내보낸 시접 `project.ConfigureShellEnvFn` 하나(design.md §4.2)를 받아들이며 붙인 조건 — 시접을 바꾸는 테스트는 `t.Parallel` 을 쓰지 않고 `t.Cleanup` 으로 원래 값을 되돌린다 — 을 인수 기준 AC-IQW-016 으로 올렸다. 대상은 고정 파일 목록이 아니라 `internal/cli`·`internal/core/project` 에서 시접에 대입하는 모든 테스트·헬퍼이며, 검증 시점에 다시 뽑은 대입 파일이 계획한 수보다 적으면 실패로 본다. 병렬 금지는 텍스트(`t.Parallel` 부재)와 실행(`t.Setenv` 병렬 충돌 panic)으로, 원복은 텍스트(대입 줄 수·`t.Cleanup`)와 실행(`-count=2` 재진입)으로 나눠 관측한다. REQ 문구는 바꾸지 않았다(REQ-IQW-011·012, §4.2 3·6항에 매핑). (2) AC-IQW-003 과 완료 정의의 "이 카드가 바꾸지 않았다" 판정을 리터럴 핀 `120436f58` 기준 diff 에서, 흡수한 로컬 `develop` 과의 merge-base 기준으로 바꿨다. 2026-09-11 재측정에서 `origin/develop` 이 118커밋 앞섰고 그중 t587 `c4990eea7` 가 `internal/cli/update_wizard.go` 를 고쳐, 통합 창에서 흡수하면 리터럴 핀 판정이 이 카드와 무관한 이유로 빨개지기 때문이다. 기준 내용 추출(`git show 120436f58:…`, `git grep … 120436f58`)은 핀을 유지한다. (3) AC-IQW-005 의 금지 토큰 검사(`exit=1` 기대)가 두 겹으로 공허했던 것을 고쳤다 — `--untracked` 가 없어 run 단계가 커밋 전에 재면 새 파일을 읽지 않았고, 파일이 아예 없어도 `exit=1` 이 나왔다. 모든 `git grep` 에 `--untracked` 를 붙이고, 나열한 파일 수와 `package cli` 도달 수가 목록 길이와 같아야 한다는 대조군을 더했다. plan 시점에는 두 대조군이 0 으로 RED 다. 새로 만들 테스트 파일에 대해 no-match `exit=1` 을 기대하는 다른 AC 는 AC-IQW-016 뿐이며, 이미 `--untracked` 와 대입 파일 수 하한을 갖췄다. AC-IQW-002·012·013 의 no-match 검사는 이미 추적되는 비테스트 소스를 보므로 그대로 두었다.
- 2026-09-11 — v0.1.4 개정 (manager-spec, plan-audit 2회차 FAIL 0.86 반영 — `.moai/reports/t583/plan-audit-iter2.md` D9~D16, 리드 추가 지시 포함). D9: AC-IQW-005 의 대상을 고정 세 파일에서 검증 시점 스윕으로 바꿨다 — 카드가 추가한 `internal/cli/*_test.go`(커밋분+미커밋분) 가운데 init 을 실행하는 호출(`runInit`, `runInit` 으로 시작하는 헬퍼, `prepareSafeInitHome`)을 담은 파일에 계획한 세 파일을 합친다. 목록 수 하한 3 과 도달 대조군(목록 수와 같아야 함)을 두고, 금지 토큰에 기존 HOME 헬퍼 호출을 더했으며, 네 번째 파일 우회 뮤턴트 F 를 run 단계 기록 의무로 두었다. D10: AC-IQW-004 에 본문 보존 관측을 더했다 — `120436f58` 의 `configureShellEnv` 와 변경 뒤 `defaultConfigureShellEnv` 의 `ConfigOptions` 옵션 줄 추출 비교(기준 4줄). 포인터 비교는 보조로 남기고, 필드 누락 뮤턴트 E(`PreferLoginShell` 삭제)를 기록 의무로 두었다. D11: REQ-IQW-011 의 착지 순서 절은 유지하고, 컴파일 순서라는 구조적 근거를 적었다. D12: §4.2 머리말에 코드 점검표의 대상(`internal/cli` 의 `runInit` 실행 테스트)과 `internal/core/project` 게이트 테스트의 의무 범위를 적었다. D13: AC-IQW-016 뮤턴트 C1 의 기대를 "대입 줄 수 1 감소, 판정은 실행 RED" 로 고쳤다. D14: AC-IQW-004 명령 블록에 뮤턴트 B 실행·FAIL 계수와 원복 뒤 PASS·`no tests to run` 계수를 적었다. D15: AC-IQW-015 의 기록 계수를 슬롯 행에 고정했다. D16: AC-IQW-003 에 세 점 형식의 기지 입력 RED(`update_wizard.go` → 1)/GREEN(`questions.go` → 0) 관측을 적었다. REQ 16 개와 AC 17 건은 그대로다.

## §1 배경과 문제

`moai init` 대화형 위저드는 질문 18개를 한 화면에 늘어놓는다(`internal/cli/wizard/questions.go:296-303` `InitQuestions` = `DefaultQuestions` 5 + `Page3Questions` 13). 그중 대부분은 기본값을 그대로 받아들여도 되는 설정이고, 같은 값을 `moai web` 설정 화면이나 CLI 플래그로 나중에 바꿀 수 있다. 감사 보고서(`init-tui-audit-20260909`)는 첫 실행 경험에 직결되는 질문만 남기자고 제안했고, 운영자가 남길 질문 4개를 확정했다.

질문 수만의 문제가 아니다. 묻기만 하고 기록하지 않는 질문이 이미 있다.

- **F1 — 재현됨.** 위저드에서 `worktree_auto_create` 에 "예"를 골라도 `workflow.yaml` 에 기록되지 않는다. 위저드 답은 `opts.WorktreeAutoCreate` 에만 들어가고(`internal/cli/init.go:305-307`), 기록기는 추적자 `WorktreeAutoCreateSet` 이 켜진 키만 쓰는데(`internal/core/project/initializer_workflow_toggles.go:38`), 추적자를 켜는 곳은 플래그 경로(`internal/cli/init_workflow_flags.go:40-44`)뿐이다. 기존 회귀 테스트(`internal/cli/init_workflow_wiring_test.go:100`)는 위저드 답을 `false` 로만 넣어, 답이 버려져도 통과했다. 실제 `runInit` 경로 재현 테스트가 `auto_create: false` 배포를 관측했다(verdict.md §2 F1).
- **서로 모순되는 설명 세 곳.** `internal/core/project/initializer.go:57-62` 는 "위저드 답은 참고용이며 템플릿 기본값을 건드리지 않는다"고 적고, `internal/cli/init.go:300-304` 는 "위저드가 돌았고 플래그가 없으면 위저드 답이 적용된다"고 적으며, `internal/cli/wizard/questions.go:372` 의 질문 설명은 init 이 워크트리에 자동 진입하게 된다고 안내한다.

### §1.1 카드 전제 정정 — F4 는 신규 init 의 결함이 아니다

감사 보고서의 F4 ("`--non-interactive` init 은 MCP 를 설치하지 않는다")는 실행으로 재 보니 성립하지 않았다. 비대화형 신규 init 이 배포한 `.mcp.json` 의 `mcpServers` 키는 `[moai context7]` 였다(verdict.md §2 F4). 템플릿 `.mcp.json` 에 moai 항목이 이미 들어 있기 때문이다. 비대화형에서 실제로 생략되는 것은 `internal/cli/init.go:1004-1011` 의 항목 보장 호출과 그 안내 문구뿐이며, 이 동작은 `internal/cli/init_agent_wizard_test.go:172-205` 가 의도된 동작으로 고정하고 있다. 따라서 이 SPEC 은 F4 수리를 요구하지 않는다. 기존 `.mcp.json` 이 있는 디렉터리(`--force` 재초기화, moai 항목이 없는 사용자 파일)는 측정하지 않았으며 plan.md §K 미측정 목록에 남긴다. `--no-mcp` 플래그는 운영자 결정으로 범위에서 뺐다.

## §2 확정된 범위

### §2.1 남기는 질문 4개 — [EXISTING]

| 순서 | ID | 기본값 | 그룹 라벨(현행 유지) |
|---|---|---|---|
| 1 | `conversation_language` | `en`, 프로필 로케일로 미리 채움 | Basic |
| 2 | `user_name` | 빈 값, 프로필로 미리 채움 | Basic |
| 3 | `agent_wiring` | `claude` | Quality & Workflow |
| 4 | `autonomy_tier` | `semi-auto` | Autonomy |

### §2.2 init 경로에서 없애는 질문 14개 — [REMOVE]

| ID | 없앤 뒤 해석되는 값 | 해석 경로 |
|---|---|---|
| `project_name` | 프로젝트 디렉터리 이름 | `applyDetectedDefaults` 가 `filepath.Base` 로 채움(`internal/core/project/phase.go:195-199`). 파일시스템 루트 예외는 REQ-IQW-003 |
| `model_policy` | `llm.profile: medium`, `llm.performance_tier` 는 템플릿 값 `"medium"` | `template.ApplyProfile` 가 빈 값을 medium 으로 정규화(`internal/cli/init.go:940-953`), 성능 등급 기록은 값이 있을 때만(`init.go:925-933`) |
| `report_format` | `html+md` | `writeReportConfig` 의 빈 값 기본값(`internal/core/project/initializer.go:579-593`) |
| `project_mode` | `personal` | 템플릿 `project.yaml.tmpl` 의 `mode: personal` |
| `worktree_auto_create` | `false` | 템플릿 `workflow.yaml` 의 `auto_create: false` |
| `todo_enabled` | 해석값 `true` (디스크에는 키 없음) | 로더가 키 부재를 활성으로 해석(`internal/config/todo_enabled.go` `Config.TodoEnabled`) |
| `feedback_auto_submit` | `false` | 템플릿 `feedback.yaml` 의 `auto_submit: false` |
| `project_continuation` | `card` | 템플릿 `workflow.yaml` 의 `continuation: card` |
| `audit_model` | `claude` | 컴파일 기본값(`internal/config/defaults.go:966-973`) |
| `audit_gate_claude` / `audit_gate_codex` / `audit_gate_glm` | `required` / `required` / `advisory` | 같은 컴파일 기본값 |
| `codex_audit_enabled` | `false` | 템플릿 `workflow.yaml` 의 `review_gate` 기본값 |
| `mcp_provision` | 대화형: 항목 보장 호출 실행 / 비대화형: 현행 유지 | REQ-IQW-005, REQ-IQW-006 |

`project_name`·`model_policy`·`report_format` 세 문항은 `DefaultQuestions` 에 있어 reconfigure 와 공유된다. 이 SPEC 은 이 셋을 **init 경로에서만** 없앤다(§2.3 D1).

### §2.3 리드 결정

- **D1 = (a)** — 질문 생성자를 나눠 `moai update -c` 의 reconfigure 질문 집합(`ReconfigureQuestions` = `DefaultQuestions` + `GitQuestions`, `questions.go:268-289`)은 지금 그대로 둔다. 이 카드는 init 만 바꾼다.
- **D2 = (a)** — `mcp_provision` 을 없앤 뒤에도 대화형 경로는 코드 기본값으로 프로비저닝을 켠다. 오늘 기본값을 받아들인 사용자의 동작이 보존된다. 비대화형은 `init_agent_wizard_test.go:172-205` 가 고정한 현행 동작을 유지한다.
- **안전** — init 흐름의 셸 설정 단계를 교체 가능한 테스트 시접을 거쳐 실행하게 하고, 그 시접을 run 단계의 선행 조건으로 둔다(REQ-IQW-011). 테스트는 이 시접에 스파이를 끼워 실제 셸 설정 파일에 쓰지 않은 채 단계 도달을 관측한다. 소스 문자열 검사는 실행 전달을 증명하지 못하므로 증거로 쓰지 않는다(리드 결정, 2026-09-11).
- **기존 테스트 소급 범위 = (a)** (2026-09-11) — §4.2 점검표는 이 SPEC 이 새로 쓰거나 본문을 다시 쓰는 init 실행 테스트의 **코드**에만 넣는다. 제거된 필드 참조만 지우는 기존 테스트(예: `internal/cli/init_agent_wizard_test.go:64-160`, `internal/cli/doctor_codex_e2e_test.go`, `internal/cli/init_workflow_wiring_test.go:100-134`)는 기존 `t.Setenv("HOME")` 헬퍼를 그대로 쓴다. 그 대신 기존·신규를 가리지 않고 init 실행 테스트를 돌리는 **모든 run 슬롯**은 §4.3 실제 홈 지문 절차를 따른다(REQ-IQW-014, REQ-IQW-016).

## §3 요구사항 (GEARS)

- **REQ-IQW-001** (Ubiquitous) [MODIFY] — init 위저드 질문 집합은 정확히 네 문항 — `conversation_language`, `user_name`, `agent_wiring`, `autonomy_tier` — 을 이 순서로 담아야 한다(shall).
- **REQ-IQW-002** (Unwanted) [REMOVE] — init 위저드는 §2.2 의 제거 대상 14문항 중 어느 것도 제시해서는 안 된다(shall not).
- **REQ-IQW-003** (Event-driven) — **When** 대화형 init 이 제거된 키를 묻지 않고 끝나면, init 흐름은 그 키가 §2.2 표의 값으로 해석되도록 남겨야 한다(shall). 해석된 값은 오늘 모든 질문에 기본값을 받아들인 사용자가 얻는 값과 같아야 한다. 단, 프로젝트 루트가 파일시스템 루트인 경우의 `project_name` 은 이 동일성에서 제외한다 — 오늘 위저드는 이름이 `.`·`/`·`\` 이면 `my-project` 로 바꾸고(`internal/cli/wizard/questions.go:50-53`), 정리 뒤 기본값 경로는 `filepath.Base` 결과를 그대로 쓰며(`internal/core/project/phase.go:195-199`), 이 경우는 측정하지 않는다.
- **REQ-IQW-004** (Ubiquitous) [EXISTING] — `moai update -c` 의 reconfigure 질문 집합은 현재의 12문항 ID 집합과 순서(`conversation_language`, `user_name`, `project_name`, `model_policy`, `report_format`, 뒤이어 Git 7문항)를 그대로 유지해야 한다(shall).
- **REQ-IQW-005** (State-driven) [MODIFY] — **While** init 이 대화형으로 실행되는 동안, init 흐름은 `.mcp.json` moai 항목 보장 호출을 기본으로 실행해야 한다(shall). 하네스 선택이 `codex` 면 호출을 건너뛰고 `both` 면 실행한다는 기존 규칙은 그대로 적용된다.
- **REQ-IQW-006** (State-driven) [EXISTING] — **While** init 이 비대화형(`--non-interactive` 또는 TTY 부재)으로 실행되는 동안, init 흐름은 현재 동작을 유지해야 한다(shall). 항목 보장 호출과 안내 문구는 생략되고, 템플릿 `.mcp.json` 은 moai 항목을 담은 채 배포되며, 플래그 없는 실행의 `workflow.yaml` 은 템플릿과 바이트 동일하다.
- **REQ-IQW-007** (Capability gate) [EXISTING] — **Where** `--project-mode`, `--worktree-auto-create`, `--autonomy-tier`, `--llm` 플래그가 명시되면, init 흐름은 질문 제거와 무관하게 그 값을 지금처럼 기록해야 한다(shall).
- **REQ-IQW-008** (Ubiquitous) [MODIFY] — 워크트리 자동 생성의 init 시점 기록 규칙을 설명하는 코드 주석은 정리 뒤 하나의 규칙만 서술해야 한다(shall): init 에서는 명시한 `--worktree-auto-create` 플래그만 기록하고, 대화형으로 바꾸는 경로는 reconfigure 단계와 web 콘솔이다. 위저드 답을 근거로 드는 서술은 남아서는 안 된다(shall not).
- **REQ-IQW-009** (Ubiquitous) [NEW] — "미설정 = 기본값" 성질은 실제 init 실행 경로를 거친 결과, 곧 디스크에 남은 설정 파일과 로더가 해석한 설정을 관측하는 테스트로 입증되어야 한다(shall). 질문 목록만 검사하는 테스트는 이 요구를 충족하지 못한다. 실행 절차는 acceptance.md AC-IQW-006 이 정한다.
- **REQ-IQW-010** (Event-driven) [NEW] — **When** 실행 테스트의 관측기가 제거된 키 가운데 하나라도 기본값이 아닌 값을 만나면, 그 테스트는 실패해야 한다(shall). 이 실패 가능성은 같은 테스트 안의 음성 대조군과, 기록된 코드 뮤턴트 실행으로 입증되어야 한다.
- **REQ-IQW-011** (Ubiquitous) [NEW] — init 흐름의 셸 설정 단계는 교체 가능한 테스트 시접을 거쳐 실행되어야 한다(shall). 시접의 운영 기본값은 현재의 셸 설정 기록 동작이어서 운영 동작은 바뀌지 않으며, 테스트는 이 시접을 바꿔 끼워 실제 셸 설정 파일에 쓰지 않은 채 단계가 호출됐는지 관측할 수 있어야 한다. 이 시접은 어떤 init 실행 테스트보다 먼저 착지해야 한다. 스파이를 끼우는 테스트는 시접 변수가 없으면 컴파일되지 않으므로 이 순서는 구조적으로 강제되며, 그래서 인수 기준에서 따로 재지 않는다. 시접의 형태와 위치는 design.md §4 가 정한다.
- **REQ-IQW-012** (Unwanted) [NEW] — 이 SPEC 이 새로 쓰거나 본문을 다시 쓰는 init 실행 테스트는 실행 전후로 §4.2 5항의 대조 8항목 — `~/.claude/settings.json` sha256, `~/.claude/hooks/moai` 존재 여부, `~/.zshenv`·`~/.zshrc`·`~/.zprofile`·`~/.profile`·`~/.bashrc`·`~/.bash_profile` 의 mtime 과 sha256 — 을 바꿔서는 안 된다(shall not). 각 테스트는 §4.2 점검표를 코드에 갖춘다. 8항목 밖의 홈 경로는 이 요구의 대상이 아니며 plan.md §K 미측정으로 남는다. 제거된 필드 참조만 지우는 기존 테스트는 이 요구의 코드 대상이 아니며 기존 HOME 헬퍼를 유지한다.
- **REQ-IQW-013** (Unwanted) [REMOVE] — 위저드 패키지와 init 경로는 제거된 질문에만 딸린 번역 항목·답 저장 분기·결과 필드·매핑·기록기를, CLI 플래그나 reconfigure 같은 다른 호출자가 닿지 않는 한 남겨서는 안 된다(shall not). 다른 호출자가 닿는 항목은 유지한다.
- **REQ-IQW-014** (Event-driven) [NEW] — **When** run 단계의 실행 슬롯이 기존 것이든 새 것이든 init 실행 테스트를 하나라도 돌리면, 그 슬롯은 실행 직전과 직후에 §4.3 의 실제 홈 지문을 글자 그대로 같은 명령으로, 표준 오류를 표준 출력과 분리해 채집하고, 슬롯 선언·두 지문·명령·종료 코드를 progress.md §E.2 에 남겨야 한다(shall).
- **REQ-IQW-015** (Event-driven) [NEW] — **When** 대화형 init 이 유지 4문항의 답만 받고 끝나면, init 흐름이 남기는 `.moai/config/sections/` 의 `workflow.yaml`·`project.yaml`·`report.yaml`·`feedback.yaml`·`llm.yaml` 은 같은 이름의 디렉터리에서 플래그 없이 실행한 비대화형 init 이 남기는 파일과 바이트 동일해야 한다(shall). §6 에 적은 디스크 모양 변화가 이 요구의 결과다.
- **REQ-IQW-016** (Event-driven) [NEW] — **When** 한 슬롯의 두 지문이 다르거나 어느 쪽 표준 오류가 비어 있지 않으면, 그 슬롯의 작업자는 이후 작업을 멈추고 리드에게 보고해야 하며(shall), 실제 홈 파일을 되돌려서는 안 된다(shall not).

## §4 제약

### §4.1 일반 제약

- SPEC 은 무엇을·왜를 정한다. 함수 이름·배치는 design.md 가 제안하고 run 단계가 확정한다.
- 남는 4문항의 그룹 라벨·렌더링·huh 그룹 구성은 바꾸지 않는다. 렌더링과 i18n 정리는 카드 t586 소관이며, 같은 위저드 파일을 lane-2 가 다룬다. 위저드 변경 구간은 plan.md §F.1 에 명시한다.
- `moai init`·`moai update` 바이너리 실행으로 검증하지 않는다. init 꼬리의 전역 설정 정리가 실제 홈에 쓰기 때문이다(verdict.md §1).
- 셸 설정 단계 시접의 도달은 실행으로 관측한다. 소스 문자열 검사는 텍스트 패턴 추론일 뿐이라, 배선을 우회하는 변형도 통과할 수 있으므로 증거로 쓰지 않는다.

### §4.2 홈 안전 점검표 — 코드 의무 (이 SPEC 이 새로 쓰거나 본문을 다시 쓰는 init 실행 테스트)

아래 점검표는 이 SPEC 이 새로 쓰거나 본문을 다시 쓰는 init 실행 테스트의 **코드에** 들어간다. 대상은 `internal/cli` 의 `runInit` 실행 테스트이며, `internal/core/project` 게이트 테스트는 `MOAI_HOME` 우회·스파이·원복만 갖추고 실제 홈은 §4.3 슬롯 지문이 덮는다. 제거된 필드 참조만 지우는 기존 테스트(예: `internal/cli/init_agent_wizard_test.go:64-160`, `internal/cli/doctor_codex_e2e_test.go`, `internal/cli/init_workflow_wiring_test.go:100-134`)는 이 점검표의 대상이 아니며 기존 `t.Setenv("HOME")` 헬퍼를 그대로 쓴다. 두 부류 모두 실행될 때는 §4.3 절차가 적용된다.

1. `t.Setenv("HOME", …)` 을 쓰지 않는다.
2. `userHomeDirFn`, `profile.BaseDirOverride`, `MOAI_HOME` 을 모두 `t.TempDir()` 아래로 돌린다.
3. 셸 설정 단계 시접(REQ-IQW-011)에 기록용 스파이를 끼우고 테스트가 끝나면 `t.Cleanup` 으로 원래 값으로 되돌린다. 스파이는 호출을 기록할 뿐 실제 셸 설정 파일에 쓰지 않는다. 시접을 바꾸는 모든 테스트·헬퍼에 대한 판정은 acceptance.md AC-IQW-016 이다.
4. `runInit` 호출 전에, 돌린 홈 세 곳이 각각 `os.UserHomeDir()` 밖에 있는지 단언한다. 안에 있으면 `t.Fatal`.
5. 실행 전후 대조 목록(8항목) — `~/.claude/settings.json` sha256, `~/.claude/hooks/moai` 존재 여부, `~/.zshenv`·`~/.zshrc`·`~/.zprofile`·`~/.profile`·`~/.bashrc`·`~/.bash_profile` 의 mtime 과 sha256. 하나라도 바뀌면 테스트 실패.
6. 환경과 패키지 변수를 바꾸므로 `t.Parallel` 을 쓰지 않는다. 시접을 바꾸는 모든 테스트·헬퍼에 대한 판정은 acceptance.md AC-IQW-016 이다.

### §4.3 실제 홈 지문 절차 — run 슬롯 의무 (기존·신규 init 실행 테스트 모두)

REQ-IQW-014 와 REQ-IQW-016 의 실행 절차다. 테스트 코드 안의 점검(§4.2)과 별개로, 테스트 프로세스 밖에서 실제 홈을 잰다.

- **슬롯의 범위**: run 단계에서 `go test` 를 `./internal/cli` 패키지 또는 `./internal/core/project/...` 에 대해 실행하는 모든 호출이 슬롯이다. `-run` 선택자가 있든 없든, 뮤턴트 실행(AC-IQW-004·007b)이든 마감 검증(AC-IQW-014)이든 같다. 두 패키지 모두 init 을 실제로 도는 기존 테스트를 담고 있기 때문이다.
- **슬롯 선언**: 작업자는 슬롯을 실행하기 전에 progress.md §E.2 의 슬롯 표에 `SLOT-<번호>` 행으로 선언한다. 선언 수·지문 파일 수·기록 수는 AC-IQW-015 가 대조한다.
- **지문 항목(8)**: `~/.claude/settings.json` sha256, `~/.claude/hooks/moai` 존재 여부, `~/.zshenv`·`~/.zshrc`·`~/.zprofile`·`~/.profile`·`~/.bashrc`·`~/.bash_profile` 의 mtime 과 sha256. 없는 파일은 "absent" 로 기록한다.
- **같은 명령**: 실행 전 지문과 실행 후 지문은 글자 그대로 같은 명령 본문으로 잰다. 달라지는 것은 출력 파일 이름의 before/after 뿐이다. 앞뒤를 다른 도구로 재면 내용이 같아도 형식 차이가 거짓 차이로 나온다 — 카드 t661 에서 실제로 일어났다. 기준 명령은 acceptance.md AC-IQW-015 에 있다.
- **표준 오류 분리**: 표준 출력과 표준 오류를 서로 다른 파일로 받는다. 표준 오류가 비어 있지 않으면 그 지문은 믿을 수 없으므로 차이가 난 것과 똑같이 다룬다.
- **증거**: 두 지문의 내용, 실행한 명령, 각 종료 코드, 두 지문의 diff 종료 코드를 progress.md §E.2 의 해당 슬롯 행에 남긴다.
- **멈춤 규칙 (REQ-IQW-016)**: 차이가 하나라도 있으면 작업자는 멈추고 리드에게 보고한다. 실제 홈 파일을 되돌리지 않는다(백업 복원·편집·삭제 모두 금지). 다른 세션이 같은 시각에 파일을 고쳐 생긴 차이도 같은 절차를 따르며, 원인 판정은 리드가 한다.
- **판독의 실행 확인**: "기존 HOME 헬퍼를 쓰면 셸 감지도 임시 홈을 따라가므로 셸 설정 누출이 없다" 는 `internal/shell/detect.go:128` 의 코드 판독일 뿐 실행으로 확인한 사실이 아니다. 기존 헬퍼를 쓰는 테스트가 도는 첫 run 슬롯의 지문 대조가 이 판독의 첫 실행 확인이 된다.

## §5 기존 SPEC 과의 관계 (본문은 수정하지 않음)

아래 SPEC 들의 HISTORY 개정은 sync 단계에서 다룬다. 이 표는 어떤 요구가 대체되는지만 기록한다.

| SPEC (상태) | 대체되는 요구 | 유지되는 요구 |
|---|---|---|
| SPEC-CLI-WIZARD-RESTRUCTURE-001 (completed) | REQ-WIZ-001(init 3페이지 구성), REQ-WIZ-003 중 init 의 `project_name` 문항, REQ-WIZ-004(init 의 Model & Report 페이지), REQ-WIZ-015 중 `project.mode` 위저드 답 기록 | REQ-WIZ-016(reconfigure 순서). REQ-WIZ-014·017 은 이번 제거의 이행 의무로 적용 |
| SPEC-V3R5-INIT-WIZARD-EXPANSION-001 (implemented) | REQ-IWE-001(`project.mode` 질문) | REQ-IWE-008 의 `--project-mode` 플래그 |
| SPEC-INIT-WIZARD-REPAIR-001 (completed) | REQ-005 의 위저드 쪽(플래그와 겨룰 위저드 답이 사라짐), REQ-008(init 에서 audit 블록 기록 — 도달 불가가 되어 기록기 제거) | REQ-006(추적자 기반 기록), REQ-007(reconfigure 워크플로 단계) |
| SPEC-INIT-HARNESS-PROMPT-001 (completed) | REQ-IHP-009 의 "`claude` 는 `mcp_provision` 답을 유지" 절(대상 질문 소멸 — 대화형 기본값 유지로 바뀜), plan.md 결정 B1("`mcp_provision` 무조건 질문") | REQ-IHP-001(하네스 선택 질문) |
| SPEC-MCP-DEFAULT-ON-001 (completed) | REQ-A-3 의 기본 true 확인 질문과 거절 경로, REQ-A-5 중 위저드 로케일 문자열 | REQ-A-3 의 "명시적 거절이 없으면 기본 프로비저닝" 방향 |
| SPEC-MOAI-MCP-SERVER-001 (completed) | REQ-MCP-015 의 init 위저드 절반(audit·게이트·codex 검토·MCP 프로비저닝 질문) | REQ-MCP-015 의 web 콘솔 절반 |
| SPEC-TODO-ENABLE-FLAG-001 (completed) | REQ-4(init 위저드 질문) | REQ-1(키 부재 시 활성), REQ-5(web 토글) |
| SPEC-FEEDBACK-AUTO-SUBMIT-001 (completed) | REQ-11(init 위저드 동의 질문) | REQ-1(기본 OFF), REQ-12(web 토글) |
| SPEC-PROJECT-CONTINUATION-KEY-001 (completed) | REQ-PCK-010(위저드의 continuation 선택 질문) | 설정 키와 기본값 `card` |

## §6 알려진 동작 변화 — 새 프로젝트의 디스크 모양

해석된 설정값은 바뀌지 않지만, 대화형으로 만든 **새** 프로젝트의 파일 모양은 달라진다. 이는 받아들인 변화이며 결함으로 보지 않는다. 바뀐 뒤의 모양은 REQ-IQW-015 가 요구한다.

- 오늘 대화형 init 은 `AuditConfigSet = true`(`internal/cli/init.go:338`) 때문에 `workflow.yaml` 의 `workflow.audit` 아래에 `model` 과 `gates`(claude·codex·glm) 키를 명시적으로 넣는다. 정리 뒤에는 이 하위 키들이 없고 컴파일 기본값이 같은 값을 공급한다. `audit:` 키 자체는 남는다 — 템플릿이 `internal/template/templates/.moai/config/sections/workflow.yaml:85-91` 에서 `audit.codex`·`audit.glm` 의 model·effort 핀을 싣고 배포하기 때문이다.
- 오늘 대화형 init 은 `workflow.todo.enabled: true` 를 넣는다. 정리 뒤에는 키가 없고 로더가 활성으로 해석한다.
- 오늘 대화형 init 은 `llm.performance_tier` 를 따옴표 없는 `medium` 으로 다시 쓴다. 정리 뒤에는 템플릿의 `"medium"` 이 남는다.
- 결과적으로 대화형 init 이 남기는 이 섹션 파일들은 비대화형 init 의 것과 같은 모양이 된다(REQ-IQW-015, acceptance.md AC-IQW-008).
- 위험: `moai update` 의 3-way 병합은 init 시점 스냅숏을 기준으로 삼는다. 모양이 달라진 새 프로젝트와 기존 프로젝트가 병합에서 다르게 다뤄질 수 있다(plan.md §J R3).

## §7 Brownfield 변경 요약

| 표지 | 대상 | 내용 |
|---|---|---|
| [MODIFY] | `internal/cli/wizard/questions.go` `InitQuestions`, `Page3Questions` | init 전용 4문항 구성. `DefaultQuestions`·`GitQuestions`·`ReconfigureQuestions` 는 [EXISTING] |
| [REMOVE] | `internal/cli/wizard/questions.go`, `wizard.go`, `translations.go`, `types.go` | 제거 질문 11개(페이지 3 전용)의 정의·답 저장 분기·ko/ja/zh 번역·결과 필드 |
| [EXISTING] | `project_name`·`model_policy`·`report_format` 의 정의·번역·답 저장 분기 | reconfigure 가 계속 사용 |
| [MODIFY] | `internal/cli/init.go` 위저드 결과 적용부, MCP 기본값 | 제거 질문의 매핑 삭제, 대화형 프로비저닝 기본 켜기 |
| [REMOVE] | `internal/core/project` 의 audit·todo·feedback·continuation 기록기와 해당 `InitOptions` 필드 | 호출자가 사라지는 항목만. 판정 기준과 목록은 plan.md §G |
| [EXISTING] | `--project-mode`·`--worktree-auto-create` 플래그 경로와 그 기록기, `runWorkflowConfigStep` | 플래그와 reconfigure 가 계속 도달 |
| [EXISTING] | 기존 테스트 헬퍼 `runInitForAutonomyAtHomeCapturingOut`(`internal/cli/init_autonomy_wiring_test.go:40`) | 필드 참조만 지우는 기존 테스트가 계속 사용(§2.3) |
| [MODIFY] | `internal/core/project/initializer.go` 셸 설정 단계(Step 6 `:333-347`, `configureShellEnv` `:677`) | 셸 설정 기록을 교체 가능한 테스트 시접을 거쳐 호출. 운영 기본값은 현재 동작 그대로(REQ-IQW-011) |
| [NEW] | `internal/cli` init 실행 테스트, 홈 안전 헬퍼, 셸 설정 단계 스파이 관측 테스트 | REQ-IQW-009~012 |

## §8 Exclusions (What NOT to Build)

이 SPEC 에서 만들지 않는 것은 아래와 같다 (out of scope).

### Out of Scope — `--no-mcp` 플래그

- `--no-mcp` 신설은 운영자 결정으로 범위에서 뺐다. F4 가 신규 init 에서 재현되지 않았으므로 수리 대상도 없다.

### Out of Scope — reconfigure 의 버려지는 답

- reconfigure 가 묻는 `report_format`·`project_name` 답을 `applyWizardConfig`(`internal/cli/update_wizard.go:133`)가 저장하지 않는 기존 결함은 카드 t588 로 넘긴다. 이 SPEC 은 reconfigure 질문 집합을 바꾸지 않는다.

### Out of Scope — 다른 카드 소관

- 자율 등급 선택지 재정의(기본값을 accept edits 로 바꾸는 일 포함)는 카드 t584 소관이다.
- 하네스 3-way 배포 의미(both·codex 단독 배포 범위)는 카드 t585 소관이다.
- 위저드 TUX 렌더링, i18n 과 huh v1 통합, v1 프로필 위저드 흡수는 카드 t586 소관이다. 그룹 라벨 재구성·스타일 변경을 하지 않는다.

### Out of Scope — web 콘솔 표면 추가

- `project.name`·`project.mode`·`.mcp.json` 프로비저닝을 web 콘솔에서 편집하는 표면은 추가하지 않는다. 이 값들은 CLI 플래그(`--project-mode`, `--llm`)와 위치 인자로만 바꿀 수 있게 남는다.

### Out of Scope — 다른 SPEC 본문 수정

- §5 에 적은 SPEC 들의 본문과 HISTORY 는 이 plan 단계에서 고치지 않는다. 개정은 sync 단계의 일이다.

### Out of Scope — 기존 `.mcp.json` 이 있는 재초기화

- `--force` 재초기화와 moai 항목이 없는 기존 `.mcp.json` 에서의 동작은 측정하지 않았다. 요구사항으로 두지 않고 plan.md §K 미측정 목록에만 남긴다.

### Out of Scope — 기존 HOME 헬퍼의 시접 기반 이관

- `runInitForAutonomyAtHomeCapturingOut`(`internal/cli/init_autonomy_wiring_test.go:40`)와 그 호출 테스트를 `t.Setenv("HOME")` 에서 §4.2 의 시접 기반 홈 안전 헬퍼로 옮기는 일은 이 SPEC 에서 하지 않는다. 후속 후보이며, 카드는 리드가 큐 복구 뒤 발행한다(plan.md §N).

### Out of Scope — `moai update` 의 셸 설정 경로

- `internal/cli/update.go:824` 가 따로 부르는 셸 설정 기록은 이 SPEC 의 시접 대상이 아니다. init 흐름의 셸 설정 단계만 시접을 거친다.
