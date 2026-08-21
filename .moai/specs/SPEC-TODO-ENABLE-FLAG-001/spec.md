---
id: SPEC-TODO-ENABLE-FLAG-001
title: "todo 기본 사용 설정 — workflow.todo.enabled 와 런타임 안내 표면 억제"
version: 0.1.0
status: draft
created: 2026-08-22
updated: 2026-08-22
author: manager-spec
priority: High
phase: "v3.1.3"
module: "internal/config"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "todo, kanban, config, wizard, web-console, statusline, template-first"
related_specs: [SPEC-FEEDBACK-AUTO-SUBMIT-001, SPEC-KANBAN-TODO-CLI-001]
---

# SPEC-TODO-ENABLE-FLAG-001

## §A Problem / Motivation

운영자 지시(2026-08-22 2차, 카드 t170)는 백로그 큐(todo) 기능의 **기본 사용 여부**를 `moai init` 마법사와 웹 설정 화면에서 노출하도록 요구한다. 기본값은 사용함(yes)이고, 사용 안 함으로 둔 사용자에게 todo 관련 안내가 뜨지 않게 하는 것까지가 범위다.

이 SPEC은 카드 t170에서 **분리**돼 나왔다. 원래 한 SPEC이었으나 AC가 32개로 Tier M 상한(16)과 Tier L 상한(25)을 모두 넘었고, 이는 예산을 완화할 게 아니라 쪼개라는 신호이기 때문이다(`spec-workflow.md` § SPEC Complexity Tier). 피드백 축은 `SPEC-FEEDBACK-AUTO-SUBMIT-001`이 갖는다.

**Tier M 판단 근거**: 편집 대상이 `internal/config`(3) + `internal/hook`(1) + `internal/statusline`(1~2) + `internal/cli/wizard`(4) + `internal/cli/init.go`(1) + `internal/core/project`(1) + `internal/settings`(1) + `internal/web`(1) + 스킬 2사본 + 템플릿·인벤토리(2) ≈ **13파일**, 신규 LOC는 대부분 테스트라 300~1000 구간. Tier M의 "5-15 files / 300-1000 LOC"에 정확히 든다. Tier S(<5 files)는 아니다.

### §A.1 카드 전제 1건이 조사로 반증됨 — 정정 기록 (이 SPEC 소관분)

읽기 전용 렌즈 조사(`.moai/reports/t170/lens-web-todo.md`)에서 카드 전제 4건이 반증됐고, 그중 넷째가 이 SPEC 소관이다. 나머지 셋(P1~P3)은 `SPEC-FEEDBACK-AUTO-SUBMIT-001` §A.1에 있다.

**정정 P4 — "사용 안 함일 때 todo 안내가 뜨지 않게 한다"는 문자 그대로는 충족 불가능하다.**

todo 표면 9개 중 둘은 설정으로 게이트할 수 없다.

- `.claude/rules/moai/workflow/kanban-dispatch.md` 는 **의도적으로 상시 로드**되며(파일 자신이 "Intentionally always-loaded"라고 명시), `moai todo add` 를 [HARD] 유일 생산자 조항으로 담고 있다.
- `.claude/skills/moai/SKILL.md` 의 스킬 목록 메타데이터(`:6`, `:81`, `:105`, `:166-172`)가 `/moai todo` 를 발견 가능하게 만들고 "백로그에 추가해줘" 류 자연어를 라우팅한다.

Claude Code는 룰 파일을 **경로로** 로드하지 YAML 플래그로 로드하지 않으며, 이 리포에 룰 파일이나 스킬 목록을 설정 플래그로 억제하는 기제는 존재하지 않는다(조사에서 0건). 따라서 이 SPEC은 범위를 **런타임 표면으로 한정**하고, 나머지를 §B에 범위 밖으로 명시한다.

[HARD] **"사용 안 함이면 todo 안내가 전혀 뜨지 않는다"로 서술된 AC를 쓰지 않는다.** 충족 불가능한 문구의 AC는 정직한 경계보다 나쁘다 — 통과시키려면 거짓 보고를 해야 하기 때문이다.

### §A.2 운영자 결정 (재론 금지)

| ID | 해소한 전제 | 결정 |
|---|---|---|
| D3 | P4 | 플래그는 **런타임 표면만 억제**한다. 범위 안: SessionStart 백로그 알림, statusline TODO 세그먼트, `workflows/todo.md` 로의 스킬 라우팅. 범위 밖으로 명시: 상시 로드 룰, 스킬 목록 메타데이터, 슬래시 명령 스텁. |
| D4 | P3 | 마법사 질문은 **en/ko/ja/zh 4로케일**로 싣는다. 프롬프트는 템플릿이 아니라 Go 소스이므로 16언어 템플릿 중립성 규율이 닿지 않으며, 영어 단독은 `TestWizardQuestionTranslationCompleteness`(`internal/cli/wizard/translations_completeness_test.go:89`)를 실패시킨다. |

## §B Scope

**In Scope**

- `workflow.todo.enabled`(`*bool`, 부재 시 ON) 설정 키·기본값·판독 헬퍼.
- 런타임 표면 3종 억제: SessionStart 백로그 알림, statusline TODO 세그먼트, `workflows/todo.md` 스킬 라우팅.
- CLI 명령 등록 유지 결정의 명문화(REQ-3).
- `moai init` 마법사 확인 질문 **1개**(`todo_enabled`, 기본 `true`)와 그 4로케일 번역.
- 웹 스키마 1줄 등록 + i18n 4로케일.
- 템플릿 미러 + 키 인벤토리 등록 + `make build`.

**Out of Scope**

### Out of Scope — 설정으로 게이트 불가능한 안내 표면
- `.claude/rules/moai/workflow/kanban-dispatch.md`(상시 로드 룰, [HARD] 유일 생산자 조항 포함)를 억제하지 않는다.
- `.claude/skills/moai/SKILL.md` 의 스킬 목록 메타데이터를 억제하지 않는다.
- `.claude/commands/moai/todo.md` 슬래시 명령 스텁의 존재 자체를 건드리지 않는다.
- 사유는 §A.1 정정 P4. 이를 억제하는 것은 하네스 계층의 별도 문제이며, 해법 후보(예: 유일 생산자 조항을 지연 로드 동반 파일로 이관)는 전부 상시 로드 독트린을 건드리므로 이 SPEC의 크기를 넘는다.

### Out of Scope — 사장 코드 3건의 수리
- `applyAutonomyTierFromWizard`(`internal/cli/init_autonomy_wizard.go:34`) — 프로덕션 호출자 없음. **카드가 배선 선례로 지목한 파일이 바로 이것이다.**
- `applyWorkflowBranchGuardFlags`(`internal/cli/init_workflow_flags.go:36`) — 호출자 없음. 그 결과 `--branch-guard` / `--worktree-auto-*` 플래그 4개가 등록만 되고 적용되지 않는다.
- `writeWorkflowAuditYAML`(`internal/core/project/initializer_audit.go:37`) — 호출자 없음.
- 이 SPEC은 셋 중 어느 것도 고치지 않으며 **어느 것도 배선 선례로 따르지 않는다**. 살아 있는 경로는 plan.md M4에 있다. 후속 카드 후보.

### Out of Scope — 피드백 축
- `feedback.auto_submit`, 마스킹 스크러버, 취약점 분류, 확인 게이트는 전부 `SPEC-FEEDBACK-AUTO-SUBMIT-001` 소관이다.

### Out of Scope — 큐 자체의 동작
- 백로그 큐의 저장 위치·동사 집합·워크트리 해석(`internal/cli/todo.go:41-99`)은 변경하지 않는다. 이 SPEC은 안내 표면만 다룬다.

## §C Requirements (GEARS)

### REQ-1 — `workflow.todo.enabled`, 부재 시 활성

`workflow.todo.enabled` 키가 설정에 없는 경우, todo 기능은 활성으로 해석돼야 한다. 값이 `false`인 경우에만 비활성으로 해석돼야 한다.

Go 필드는 `*bool`이어야 한다(shall) — 평범한 `bool`은 "부재"와 "명시적 false"를 구별하지 못하고, 배포 템플릿은 이 블록을 담지 않는다. 판독은 `readMCPToolEnablement`(`internal/cli/mcp_server.go:378-437`)의 형태를 따른다.

키의 집은 `workflow.yaml`이다 — `auto_clear.enabled`·`branch_guard.enabled` 라는 형제 선례가 이미 있다. 새 `todo.yaml` 섹션을 만들어서는 안 된다(shall not); 섹션 신설은 등록 지점 6곳을 요구한다.

### REQ-2 — 런타임 안내 표면 3종 억제

`workflow.todo.enabled`가 `false`인 경우, 아래 표면은 todo 안내를 출력해서는 안 된다(shall not).

1. **SessionStart 백로그 알림** — `internal/hook/session_start_kanban.go:180`. 이미 kanban 환경 + `source == "startup"` 조건부이므로 한 줄 가드다. 4로케일 문자열(`session_start_kanban_i18n.go:81/103/125/147`) 어느 것도 나와서는 안 된다.
2. **statusline TODO 세그먼트** — `internal/statusline/renderer.go:188`. 이미 `isSegmentEnabled(SegmentBacklog)`로 게이트돼 있으므로 같은 판정에 플래그를 합류시킨다.
3. **`workflows/todo.md` 스킬 라우팅** — 온디맨드 로드이므로 라우터가 도달하지 않으면 억제된다. 스킬 본문(및 템플릿 미러)에 플래그 조건을 명시한다.

키가 부재하거나 `true`인 경우 세 표면은 오늘과 동일하게 동작해야 한다.

### REQ-3 — CLI 명령 등록은 플래그와 무관하게 유지

`workflow.todo.enabled`가 `false`인 경우에도 `moai todo` CLI 명령은 계속 등록돼 동작해야 한다.

근거: 카드의 요구는 "안내가 뜨지 않는 것"이지 "기능이 사라지는 것"이 아니다. 명령을 숨기면 foreman 스킬의 `allowed-tools`(`.claude/skills/moai-kanban-foreman/SKILL.md:17`)와 이미 존재하는 큐 파일이 남는데 진입점만 사라져, 되레 진단하기 어려운 상태가 된다. 이 결정을 REQ로 두는 이유는 검증 가능하게 만들기 위해서다.

### REQ-4 — `moai init` 마법사 질문 1개

`moai init`이 대화형으로 실행되는 경우, 마법사는 "Quality & Workflow" 그룹에 todo 사용 여부 확인 질문 1개(`todo_enabled`, 기본 `true`)를 제시해야 한다.

- 질문 정의는 `Page3Questions`(`internal/cli/wizard/questions.go`)에 추가한다. `DefaultQuestions`에 넣어서는 안 된다(shall not) — `TestQuestionOrder`(`questions_test.go:101`)가 5개로 고정한다.
- en/ko/ja/zh 번역을 `internal/cli/wizard/translations.go`에 함께 실어야 한다(§A.2 결정 D4).
- 답변은 `saveBoolAnswer`(`wizard.go:459`) → `WizardResult` → `applyWizardPage3ToOpts`(`internal/cli/init.go:185`) 경로로 포착하고, 파일 기록은 `WritePhase1Configs`(`internal/core/project/initializer_expansion.go:30`)에서 `yamlpatch.PatchFile`로 수행해야 한다(주석 보존 + 두 배포 경로 모두에서 실행).
- 비대화형에서는 마법사가 실행되지 않으므로 기본값(활성)이 유지된다.

### REQ-5 — 웹 콘솔 토글

웹 설정 화면에서 이 플래그를 토글할 수 있어야 한다. `internal/settings/schema_sections.go`에 `s(SectionWorkflow, "workflow", TypeBool, "workflow", "todo", "enabled")` 한 줄을 추가하면 파싱·렌더·검증·영속은 모두 제네릭으로 처리된다(`workflow.branch_guard.enabled`, `:334`가 선례).

i18n 키를 4로케일 모두에 등록해야 한다 — 누락 시 `internal/web/schema_label_test.go:96`이 실패한다.

### REQ-6 — Template-First 미러 + 키 인벤토리

키를 템플릿에 실을지 여부는 `branch_guard` 선례를 따라 **싣지 않는 것**을 기본으로 한다 — 부재가 곧 활성이고 `*bool` 판독이 그것을 처리하며, `internal/settings/schema_sections.go:330-333` 주석이 같은 상황("배포 템플릿은 블록을 담지 않고, seam writer가 첫 편집에서 중첩 매핑을 upsert한다")을 이미 문서화한다.

템플릿에 싣기로 한다면 `enabled: true`를 명시하고 `internal/config/testdata/shipped_key_inventory.yaml`에 항목을 등록해야 한다 — 미등록 시 `TestShippedConfigKeysHaveReaders`(`internal/config/shipped_key_reader_test.go:70`)가 실패한다. 어느 쪽이든 `make build`를 실행하고, 템플릿 주석은 중립을 유지한다(CLAUDE.local.md §2.1).

## §D Evidence (조사에서 관측된 사실)

전부 `.moai/reports/t170/lens-web-todo.md`(및 `lens-init.md`)에서 인용하며 각 항목이 `file:line`을 동반한다.

| 사실 | 근거 |
|---|---|
| todo 표면 9개 중 2개(상시 로드 룰, 스킬 목록)는 설정 게이트 불가 | `lens-web-todo.md` §B.4, Verdict |
| 룰 파일·스킬 목록을 설정으로 억제하는 기제가 리포에 0건 | `lens-web-todo.md` Verdict |
| `todo` 설정 키가 리포에 전무(`mx.yaml:178 todo_per_file`은 무관) | `lens-init.md` §5 |
| `workflow.yaml`에 `auto_clear.enabled`·`branch_guard.enabled` 형제 선례 | `lens-web-todo.md` §B.5 |
| `mcp.tools.<name>.enabled`의 `*bool` 기본-ON 판독 선례 | `lens-web-todo.md` §B.6 |
| statusline `backlog` 세그먼트는 이미 키-부재-활성이며 억제 가능 | `lens-web-todo.md` §B.4 #9 |
| SessionStart 알림은 이미 조건부(kanban env + startup) | `lens-web-todo.md` §B.4 #8 |
| seam+웹 렌더가 동시에 되는 섹션은 workflow·mcp·crosssession·report 4개뿐 | `lens-web-todo.md` §B.5 |
| 사장 writer 3건 확인(테스트 제외 호출자 0건) | `lens-init.md` §1, §3, §4 |
| 영어 단독 신규 질문은 번역 완전성 테스트 실패 | `lens-init.md` §5 |

## §E Constraints / Non-Goals

### §E.1 Hard 제약 — 형제 SPEC과 공유하는 파일

`SPEC-FEEDBACK-AUTO-SUBMIT-001`과 이 SPEC은 **같은 파일을 동시에 건드린다**. run-phase가 알아야 할 실제 충돌 위험이다.

| 공유 파일 | 이 SPEC이 추가하는 항목 | 형제 SPEC이 추가하는 항목 |
|---|---|---|
| `internal/cli/wizard/questions.go` | `todo_enabled` 질문 1개 | `feedback_auto_submit` 질문 1개 |
| `internal/cli/wizard/types.go` | `WizardResult` 필드 1개 | 필드 1개 |
| `internal/cli/wizard/wizard.go` | `saveBoolAnswer` case 1개 | case 1개 |
| `internal/cli/wizard/translations.go` | ko/ja/zh 3블록 × 1쌍 | 3블록 × 1쌍 |
| `internal/cli/init.go` | `applyWizardPage3ToOpts` 대입 1줄 | 대입 1줄 |
| `internal/core/project/initializer_expansion.go` | `workflow.yaml` writer | `feedback.yaml` writer |
| `internal/settings/schema_sections.go` | 필드 1줄 | 필드 1줄 |
| `internal/web/assets/i18n.js` | 4로케일 × 1쌍 | 4로케일 × 1쌍 |
| `internal/config/testdata/shipped_key_inventory.yaml` | (싣는 경우) 항목 1개 | 항목 1개 |

[HARD] **양쪽 모두 같은 파일의 서로 다른 항목만 추가한다.** 기존 항목의 재배치·재작성·서식 변경을 하지 않으며, 어느 쪽이 먼저 착지하든 나중 것이 텍스트 충돌 없이 얹혀야 한다. 두 번째로 착지하는 쪽은 병합 후 마법사 개수 고정 테스트(`TestQuestionOrder` 5개, `TestReconfigureQuestions` 12개)와 번역 완전성 테스트를 **다시** 돌려 두 질문이 함께 통과함을 확인한다(AC-T-011).

**`depends_on` 미기재 근거**: 두 SPEC 사이에 기능 의존이 없다 — 각자 독립된 설정 키를 읽고 독립된 표면을 바꾸며, 한쪽이 없어도 다른 쪽이 완결된다. 남는 것은 텍스트 충돌 위험뿐이고 그것은 순서 의존이 아니라 위 병합 규율로 다룬다.

### §E.2 Hard 제약 — 그 밖

- `*bool` 필수(REQ-1). 평범한 `bool`은 요구를 표현할 수 없다.
- 상시 로드 룰과 스킬 목록은 건드리지 않는다(§B).
- 로컬 검증은 패키지 스코프로만. 전 패키지 판정은 CI 몫(CLAUDE.local.md §4).
- 건드린 모든 패키지에 `GOOS=windows go vet`.
- 신규 테스트는 `t.TempDir()` 사용.

### §E.3 잔여 위험

**억제 범위가 부분적이라는 사실 자체가 잔여 위험이다.** 플래그를 끈 사용자도 상시 로드 룰을 통해 `moai todo add` 조항을 계속 컨텍스트에 갖고, 스킬 목록에서 `/moai todo` 를 계속 본다. 이 SPEC은 그것을 해결하지 않고 **정직하게 경계로 선언**한다. 완료 보고에서 "todo 안내를 전부 껐다"고 서술하는 것은 미검증 주장이다.

두 번째: statusline 세그먼트는 이미 `statusline.yaml`의 `backlog: false`로도 억제할 수 있다. 두 경로가 같은 표면을 끄게 되므로, 어느 쪽이 꺼도 꺼지도록(OR) 판정해야 하며 한쪽이 다른 쪽을 덮어써서는 안 된다.

## §F Cross-References

- 형제 SPEC: `SPEC-FEEDBACK-AUTO-SUBMIT-001` (카드 t170의 피드백 축)
- 근거 렌즈: `.moai/reports/t170/lens-web-todo.md`, `lens-init.md`
- `.claude/rules/moai/workflow/kanban-dispatch.md` — 억제하지 않는 상시 로드 룰(§B)
- `internal/cli/mcp_server.go:378-437` — `*bool` 기본 ON 판독 선례
- `internal/settings/schema_sections.go:334` — 웹 토글 1줄 선례
- 관련 SPEC: SPEC-KANBAN-TODO-CLI-001(큐 CLI 동사의 유래)

## §G HISTORY

- **2026-08-22** v0.1.0 — 최초 초안(plan-phase). 카드 t170에서 AC 예산 초과(32 > 상한)로 분리 신설. 카드 전제 P4가 반증돼 §A.1에 정정으로 기록하고, 충족 불가능한 문구의 AC를 쓰지 않는다는 [HARD]를 함께 남겼다. Tier M 판단 근거는 §A.
