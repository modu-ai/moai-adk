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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
