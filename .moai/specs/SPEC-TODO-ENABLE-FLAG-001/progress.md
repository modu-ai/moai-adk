# Progress — SPEC-TODO-ENABLE-FLAG-001

## §E.1 Plan-phase Audit-Ready Signal

- 2026-08-22 (iter2) — plan-audit iter1 **FAIL 0.78**(Tier M 임계 0.80; must-pass 7/7 PASS, 부족분은 Testability 0.65 단일 축) 대응 개정. 블로킹 5건(D1 공허 AC / D5 커밋 후 무조건 통과 / D6 Route 오기 / D2 명시적 호출 미정의 / D4 충돌 규율 단언) + 선택 4건(D3·D7·D8·D9) 처리. `depends_on` 미기재 근거를 "의존 없음"에서 **동시성 대 직렬화 트레이드오프 기록**으로 재작성. version 0.1.0 → 0.2.0, 항목별 처리 내역은 `spec.md` §G. AC 11개 유지(개수 불변, 관측 해상도만 상향).
- 2026-08-22 (iter1) — plan-phase 산출물 3종(Tier M) 작성. 카드 t170에서 AC 예산 초과(32 > 상한)로 분리 신설됐다. 형제 SPEC은 `SPEC-FEEDBACK-AUTO-SUBMIT-001`(Tier L). 근거는 `.moai/reports/t170/lens-web-todo.md`·`lens-init.md`. 카드 전제 P4(전면 억제 불가) 정정과 운영자 결정 D3·D4 반영. AC 11개 / Tier M 상한 16. 남은 결정: M6 템플릿 적재 여부(기본안 = 싣지 않음).

- 2026-08-22 (iter2 PASS 후속) — plan-audit iter2 **PASS 0.87**(Tier M 임계 0.80, must-pass 7/7, 0.78 → 0.87). PASS에 딸린 **run-phase 진입 전 필수 수정 2건** 처리: **N1** 한국어 리터럴 `grep -c '명시적'` 을 AC에서 제거하고 한정 문장 관측을 AC-T-005 왕복 동작에 위임(영어 전용 표면 ↔ 한국어 통과 조건의 양자택일 함정 제거), **N2** `userHomeDirFn` 홈 격리 seam을 이름으로 명명(선례 `internal/cli/todo_queue_root_test.go:122`) — `t.TempDir()` 만으로는 `resolveTodoQueueRoot` 폴백이 개발자 실제 홈에 쓴다. 형제 SPEC §E.1도 이 SPEC의 충돌 해소 규칙과 동일 내용으로 정렬. AC 11 유지, version 0.2.0 → 0.3.0.

## §E.2 Run-phase Evidence

Cycle: TDD (RED → GREEN → REFACTOR). Branch `WT-auto-feedback`, base `70620600d`.

### M1 — 설정 데이터 모델

RED (`go test ./internal/config/ -run 'TestTodoEnabled'`, 편집 전 트리):

```
internal/config/todo_enabled_test.go:69:18: cfg.TodoEnabled undefined (type *Config has no field or method TodoEnabled)
internal/config/todo_enabled_test.go:80:10: cfg.TodoEnabled undefined (type *Config has no field or method TodoEnabled)
internal/config/todo_enabled_test.go:89:6: undefined: TodoEnabledForRoot
internal/config/todo_enabled_test.go:92:6: undefined: TodoEnabledForRoot
internal/config/todo_enabled_test.go:98:5: undefined: TodoEnabledForRoot
FAIL	github.com/modu-ai/moai-adk/internal/config [build failed]
```

GREEN: `ok  github.com/modu-ai/moai-adk/internal/config  3.062s` (패키지 전체).

납품물: `internal/config/types.go` `WorkflowTodoConfig{Enabled *bool}` + `WorkflowConfig.Todo`,
`internal/config/defaults.go` 제로값 유지(nil = 활성, 그 이유를 주석으로 명시),
`internal/config/todo_enabled.go` `(*Config).TodoEnabled()` + `TodoEnabledForRoot(root)`.

### M2 — 런타임 표면 억제 2종

RED (`go test ./internal/hook/ -run 'TestSessionStartKanbanRespectsTodoDisabled'`) — 억제 케이스만 FAIL,
대조 2케이스는 이미 PASS. 대조가 먼저 초록이라는 사실이 "원래 안 나오던 것"을 통과로 오독할 여지를 없앤다:

```
--- FAIL: TestSessionStartKanbanRespectsTodoDisabled/disabled_suppresses_the_backlog_line_in_every_locale
    [en] backlog line present with todo disabled:
    [ko] backlog line present with todo disabled:
    [ja] backlog line present with todo disabled:
    [zh] backlog line present with todo disabled:
```

RED (`go test ./internal/statusline/ -run 'TestRendererBacklogSegmentGating'`):

```
internal/statusline/backlog_gating_test.go:71:6: r.SetTodoEnabled undefined (type *Renderer has no field or method SetTodoEnabled)
FAIL	github.com/modu-ai/moai-adk/internal/statusline [build failed]
```

GREEN: `ok internal/hook 0.633s`, `ok internal/statusline 0.463s`; 두 패키지 트리 전체도 초록.

납품물: `internal/hook/session_start_kanban.go` 백로그 요약 앞 `config.TodoEnabledForRoot(root)` 가드,
`internal/statusline/renderer.go` `todoEnabled *bool` + `SetTodoEnabled`/`isTodoEnabled` + 렌더 판정 합류,
`internal/statusline/builder.go` `Options.TodoEnabled`, `internal/cli/statusline.go` 배선.

statusline 억제 경로 2개는 AND로 합류한다(어느 쪽이든 끄면 꺼진다). 데이터 수집이 아니라 렌더 판정에
합류시킨 이유는 `Backlog.Available == false`(큐를 못 읽음)와 의도적 숨김이 구별 불가능해지는 것을 막기
위해서다 — 테스트 5케이스 중 `the pre-existing statusline.yaml path still suppresses` 가 신규 플래그가
기존 경로를 덮어쓰지 않음을 관측한다.

### M3 — 스킬 라우팅 한정 + CLI 등록 유지

RED 없음 — 정직하게 기록한다. `TestTodoCommandRegisteredRegardlessOfFlag` 와
`TestTodoVerbsUnaffectedByFlag` 는 이 SPEC이 **의도적으로 바꾸지 않는** 동작(REQ-3,
`internal/cli/todo.go:512` 무변경)을 단언하므로, RED를 만들려면 먼저 그 동작을 깨야 한다.
회귀 핀이지 test-first 드라이버가 아니다. 첫 실행부터 PASS:

```
=== RUN   TestTodoCommandRegisteredRegardlessOfFlag
--- PASS: TestTodoCommandRegisteredRegardlessOfFlag (0.05s)
=== RUN   TestTodoVerbsUnaffectedByFlag
--- PASS: TestTodoVerbsUnaffectedByFlag (0.10s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.972s
```

두 테스트 모두 `userHomeDirFn` 을 `t.TempDir()` 로 교체하고(N2 선례
`todo_queue_root_test.go:122`), 교체된 홈 아래에 아무것도 만들어지지 않았음을 함께 단언한다 —
seam은 검사가 따라붙어야 근거가 된다.

한정 문장은 스킬 본문 2사본(소스 + 템플릿 미러)에 영어로 실었다. 한국어 리터럴 grep으로
판정하지 않는다(N1) — 그 표면은 영어 전용이고, 한정이 지켜졌는지는 위 왕복 동작이 판정한다.

AC-T-004 관측:

```
grep -c 'workflow.todo.enabled' .claude/skills/moai/SKILL.md                             → 1
grep -c 'workflow.todo.enabled' internal/template/templates/.claude/skills/moai/SKILL.md → 1
grep -Fxc -e '  feedback, review, clean, codemaps, gate, e2e, harness, goal, todo) to' … → 1
grep -Fc  -e '- **todo** (aliases: backlog): Backlog queue — …'                          → 1
grep -Fc  -e '- Backlog language (add to the backlog, …) routes to **todo**'             → 1
```

### M4 — 마법사 질문 1개 + 실제 파일 기록

RED:

```
internal/cli/wizard/todo_enabled_test.go:38:7: r.TodoEnabled undefined (type *WizardResult has no field or method TodoEnabled)
FAIL	github.com/modu-ai/moai-adk/internal/cli/wizard [build failed]

internal/core/project/initializer_todo_test.go:50:3: unknown field TodoEnabled in struct literal of type InitOptions
FAIL	github.com/modu-ai/moai-adk/internal/core/project [build failed]
```

GREEN: `ok internal/cli/wizard 4.201s`, `ok internal/core/project 13.353s`.

설계 판단 2건:

1. **`WizardResult.TodoEnabled` / `InitOptions.TodoEnabled` 도 `*bool`이다.** 이웃 Page-3 confirm은
   전부 default-OFF 게이트를 비추지만 이 키는 default-ON이다. 평범한 bool이면 비대화형(질문 미제시)의
   제로값 false가 그대로 `enabled: false` 로 기록돼 기본 켜짐 기능이 조용히 꺼진다. nil = "묻지 않았음"
   이고, 묻지 않은 질문은 **아무것도 쓰지 않는다**(`TestWritePhase1ConfigsSkipsTodoWhenUnanswered`).
2. **기록은 `yamlpatch.PatchFile` 이다.** 이웃들이 쓰는 로컬 `patchYAMLPathValue` 는 키 부재 시
   `ok=false` 로 파일을 그대로 두는데, M6 결정에 따라 배포 workflow.yaml 에는 todo 블록이 **없다** —
   즉 정확히 중요한 경로에서 조용한 no-op가 된다. yamlpatch는 중첩 매핑을 upsert하고 주석·키 순서를
   보존한다. 파일 자체가 없는 무배포 폴백만 별도 처리했다(PatchFile 의 원자적 쓰기가 원본 권한 승계를
   위해 stat을 요구해 실패한다 — 테스트가 이 경로를 실제로 지난다).

개수 고정 테스트 4건(`TestStepperTotal_DynamicDenominator` 15→16, `TestPage3QuestionsStructure` 9→10,
`TestTotalVisibleQuestions_Page3AlwaysCounted` 14→15 / 8→9, `TestInitPages_Membership` 목록)을 갱신했다.
이들은 **재고 목록**을 단언하지 질문 개수 불변을 단언하지 않는다 — 의도적으로 질문 1개를 추가했으므로
목록 갱신이 정상 유지보수다. plan.md 가 지목한 `TestQuestionOrder`(DefaultQuestions 5개)와
`TestReconfigureQuestionsOrder`(12개)는 **손대지 않았고 그대로 통과**한다.

### M5 — 웹 스키마 + i18n

RED:

```
--- FAIL: TestWorkflowTodoEnabledFieldRegistered
    schema_todo_test.go:22: workflow.todo.enabled is not registered in AllFields()

--- FAIL: TestI18nKeySetParity
    schema_label_test.go:106: i18n.js missing key "f.workflow.todo.enabled.title" in all 4 locales
    schema_label_test.go:106: i18n.js missing key "f.workflow.todo.enabled.desc" in all 4 locales
```

두 번째 RED는 순서가 근거다 — 필드를 먼저 등록하자 parity 테스트가 **스스로** 붉어졌다. i18n 누락을
잡는 기제가 실재로 동작한다는 관측이지 문서상의 주장이 아니다.

GREEN: `ok internal/settings 3.027s`, `ok internal/web 51.267s`.

### M6 — 템플릿 결정: **블록을 싣지 않는다**

`branch_guard` 선례(`internal/settings/schema_sections.go:330-333` 주석이 같은 상황을 문서화)를 따른다.

근거: 부재가 곧 활성이고 `*bool` 판독이 그것을 처리한다. `enabled: true` 를 실으면 전 배포 프로젝트의
config에 기본값을 다시 적는 줄이 하나 늘 뿐이고, `shipped_key_inventory.yaml` 항목도 함께 요구된다 —
비용만 있고 얻는 것이 없다. 웹 seam writer(yamlpatch)가 첫 편집에서 중첩 매핑을 upsert하므로 사용자가
토글할 때 블록이 생긴다.

AC-T-010 관측:

```
grep -n 'todo' internal/template/templates/.moai/config/sections/workflow.yaml   → 0건 (결정과 일치)
go test ./internal/config/ -run 'TestShippedConfigKeysHaveReaders'               → ok 1.696s
grep -rn 'SPEC-TODO-ENABLE-FLAG\|REQ-' <템플릿 workflow.yaml + SKILL.md>          → 0건
make build                                                                        → exit 0
```

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
