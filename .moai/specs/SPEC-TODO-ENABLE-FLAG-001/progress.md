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

### M7 — 검증 스윕

| AC | Status | 판정 명령 | 관측 |
|---|---|---|---|
| AC-T-001 | PASS | `go test ./internal/config/ -run 'TestTodoEnabled' -v` | 4 서브테스트 전부 PASS, `ok … 0.258s` |
| AC-T-002 | PASS | `go test ./internal/hook/ -run 'TestSessionStartKanbanRespectsTodoDisabled' -v` | 억제 1 + 대조 2 전부 PASS, `ok … 0.709s` |
| AC-T-003 | PASS | `go test ./internal/statusline/ -run 'TestRendererBacklogSegmentGating' -v` | 5 케이스 전부 PASS(기존 statusline.yaml 경로 생존 포함), `ok … 0.450s` |
| AC-T-004 | PASS | 5개 grep (§M3 절) | `1 / 1 / 1 / 1 / 1` |
| AC-T-005 | PASS | `go test ./internal/cli/ -run 'TestTodoCommandRegisteredRegardlessOfFlag\|TestTodoVerbsUnaffectedByFlag' -v` | 2 PASS, `ok … 0.972s` |
| AC-T-006 | PASS | `go test ./internal/cli/wizard/ -run 'TestTodoEnabledQuestion\|TestQuestionOrder\|TestReconfigureQuestions' -v` | `TestQuestionOrder` PASS, `TestReconfigureQuestionsOrder` PASS, `TestTodoEnabledQuestion` PASS |
| AC-T-007 | PASS | `go test ./internal/cli/wizard/ -run 'TestWizardQuestionTranslationCompleteness' -v` | PASS |
| AC-T-008 | PASS | `go test ./internal/core/project/ -run 'TestWritePhase1ConfigsPersistsTodoEnabled' -v` | PASS, `ok … 0.455s` |
| AC-T-009 | PASS | `go test ./internal/settings/ -run 'TestWorkflowTodoEnabledFieldRegistered' -v` + `go test ./internal/web/ -run 'TestI18nKeySetParity' -v` | 둘 다 PASS |
| AC-T-010 | PASS | §M6 절의 4개 관측 | 템플릿 todo 0건(결정 일치), 인벤토리 PASS, 중립성 0건, `make build` exit 0 |
| AC-T-011 | **PARTIAL** | `-run '…\|TestTodoEnabledQuestion\|TestFeedbackAutoSubmitQuestion' -v` | `TestFeedbackAutoSubmitQuestion` 이 **RUN 으로 찍히지 않는다** — 형제 SPEC 미착지 |

**AC-T-011 은 절반만 판정됐다.** `acceptance.md` 의 "한쪽만 착지했다면 착지한 쪽 기준으로
판정하고, 나중 착지 시 재실행" 조항에 해당한다. 이 트리에는 `feedback_auto_submit` 질문이 없고
(`grep -rn 'feedback_auto_submit' internal/cli/wizard/` → 0건), 따라서 두 질문 공존 상태의 관측은
**수행되지 않았다 — 갭이다**. 이 SPEC은 첫 착지자이므로 §E.1 규칙 1에 따라 충돌 해소 소유자가
아니며, 형제 SPEC이 착지할 때 그쪽이 AC-T-011 전체를 재실행한다. `GOOS=windows go vet` 절반은
아래대로 이미 통과했다.

전 패키지 스윕(로컬 스코프, `go test ./...` 금지 규율 준수):

```
ok  internal/cli               402.814s
ok  internal/config              1.862s      (cover 80.7%)
ok  internal/config/atomicfile   0.572s
ok  internal/config/toolpolicy   cached
ok  internal/hook               22.805s      (cover 84.4%)
ok  internal/statusline         14.138s      (cover 90.5%)
ok  internal/cli/wizard          cached
ok  internal/core/project        cached       (cover 88.3%)
ok  internal/settings            cached       (cover 90.1%)
ok  internal/settings/agentfm    1.024s
ok  internal/settings/yamlpatch  0.682s
ok  internal/web                 cached
ok  internal/template            cached
```

```
GOOS=windows go vet ./internal/config/... ./internal/hook/... ./internal/statusline/... \
  ./internal/cli/... ./internal/core/project/... ./internal/settings/... ./internal/web/...   → exit 0
golangci-lint run --timeout=3m                                                                → 0 issues.
make build                                                                                     → exit 0
```

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-22
run_commit_sha: 73af5b73c  # last run-phase commit; this backfill follows it
run_status: complete-with-gap
ac_pass_count: 10
ac_fail_count: 0
ac_partial_count: 1        # AC-T-011 — 형제 SPEC 미착지, 재실행은 두 번째 착지자 소관
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-performed   # 워크트리 로컬 스택, push 금지 지시
l44_post_push_fetch: not-performed    # push 하지 않음 (리드가 통합)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  goos_windows_vet: pass
  go_build: pass
total_run_phase_files: 20
m1_to_mN_commit_strategy: 3 commits (M1-M2 / M3 / M4-M6), 각 커밋 빌드 통과
```

억제 범위에 대한 정직한 서술(§D.4 마감 게이트 항목): 이 SPEC이 끄는 것은 **런타임 표면 3종**
(SessionStart 백로그 요약 · statusline TODO 세그먼트 · `workflows/todo.md` 로의 자동 라우팅)이다.
상시 로드 룰 `kanban-dispatch.md` 와 `SKILL.md` 의 스킬 목록 메타데이터는 **범위 밖이며 여전히
로드된다** — 플래그를 끈 사용자도 컨텍스트에 `moai todo add` 조항을 계속 갖고 스킬 목록에서
`/moai todo` 를 계속 본다. "todo 안내를 전부 껐다"는 서술은 이 SPEC에 대해 거짓이다.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-23
sync_commit_sha: pending-backfill-sync   # a commit cannot cite its own hash; the lead backfills
sync_status: complete-with-carried-gap
b12_self_test_a: pass    # grep -c 'SPEC-TODO-ENABLE-FLAG-001' CHANGELOG.md → 0 (no prior entry, emission proceeds)
b12_self_test_b: pass    # 11 distinct AC ids in acceptance.md; §E.2 M7 matrix carries 11 rows
b12_self_test_c: pass    # no file paths claimed in the CHANGELOG entry — it names config keys and surfaces only
changelog_entry_position: "[Unreleased]"   # created this section; it was empty in this tree
frontmatter_status_transitions:
  spec_md: in-progress → completed (updated: 2026-08-22 → 2026-08-23)
  plan_md: n/a         # no YAML frontmatter — markdown-header convention
  acceptance_md: n/a   # no YAML frontmatter
  progress_md: n/a     # no YAML frontmatter
canary_compliance_check: n/a   # this SPEC defines no forward-looking policy that its own sync tests
```

### Carried-forward gap — AC-T-011 remains partial

`AC-T-011` closed **partial** in run-phase and closes partial here. The sibling
`SPEC-FEEDBACK-AUTO-SUBMIT-001` has not landed, so the two-questions-coexisting
tree does not exist and the shared-file conflict re-run cannot be observed. This
is the correct state under §E.1 resolution rule 1: this SPEC landed first and is
therefore not the conflict-resolution owner. The second SPEC to land re-runs
AC-T-011 in full.

### Sync-phase verification observed

```
git merge-base --is-ancestor 73af5b73c HEAD; echo $?   → 0
grep -c 'SPEC-TODO-ENABLE-FLAG-001' CHANGELOG.md        → 0 (before emission)
frontmatter completed-status grep over the 4 artifacts        → 1 hit (spec.md only)
go build ./...                                          → exit 0
```

Not observed (gaps): the test suite was not re-run this phase — there are no code
changes in the sync commit, and `go test ./internal/hook/...` is specifically
withheld because it rewrites perf fixtures belonging to another SPEC. The
run-phase evidence in §E.2 stands as the functional attribution; nothing in this
commit could invalidate it.

### Scope of what shipped, restated

The flag suppresses three runtime surfaces (SessionStart backlog summary,
statusline TODO segment, inferred skill routing). The always-loaded kanban rule
and the skill-listing metadata are out of scope and still load. "todo guidance is
fully off" is false for this change, and the CHANGELOG entry says so in the body
rather than implying otherwise.
