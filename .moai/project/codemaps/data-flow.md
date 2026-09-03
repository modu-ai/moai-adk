# 데이터 흐름

> `/moai codemaps`로 생성됐습니다. 시스템 동작의 대부분을 실어 나르는 경로 4개를
> 끝에서 끝까지 따라갑니다.

**측정 트리**: worktree `.claude/worktrees/t476`, 브랜치 `WT-codemaps-progress`, HEAD `25a3212a9`
**측정**: 2026-09-04

---

## A. CLI — `moai todo add "<text>"` : argv에서 SQLite 파일까지

```
cmd/moai/main.go                        main() → cli.Execute()
internal/cli/root.go                    Execute() → InitDependencies()   (todo는 trivial 아님)
internal/cli/deps.go                    합성 루트 조립
internal/cli/root.go → fang.go          runFang → cobra 라우팅
internal/cli/todo.go                    newTodoCmd() → add 서브커맨드 RunE
internal/cli/todo.go                    resolveTodoQueueRoot()
  └ internal/kanban/todo_root.go        ResolveTodoQueueRootAdopting
                                          — primary checkout → ~/.moai → fallback 3단 해석
internal/cli/todo.go                    todoBacklogPath() → kanban.BacklogPathForRootAdopting
internal/cli/todo.go                    newTodoStore() → kanban.NewBacklogStore(path)
internal/cli/todo_analysis.go           appendAnalyzedCard(rec, text, BacklogStateQueued, force)
internal/kanban/backlog_store.go        NewBacklogStore / Mutate(func(*BacklogRecord) error)
internal/kanban/board_lock.go           크로스 프로세스 backlog.lock (+ _unix / _windows)
internal/kanban/backlog_sqlite.go       openEngine → WAL + busy_timeout ≥ 5000 + BEGIN IMMEDIATE
                                          ↳ <queue-dir>/backlog.db
```

**주목할 점**: 큐 경로 해석 함수 `BacklogPathForRoot`는 **여전히 `backlog.json` 이름을
반환합니다.** DB는 그 형제 파일로 파생됩니다 — 다운그레이드 시 구버전 바이너리가 JSON만 읽도록
남긴 의도적 설계이며, `backlog_sqlite.go` 헤더가 그 이유를 적고 있습니다.

---

## B. 훅 — Claude Code `PreToolUse` 이벤트에서 종료 코드까지

```
Claude Code                             settings.json hooks.PreToolUse 엔트리 발화
.claude/hooks/moai/handle-pre-tool.sh   (템플릿 원본: internal/template/templates/.claude/hooks/moai/handle-pre-tool.sh.tmpl)
                                          스크립트 부재 → hook-missing.log + exit 0 (fail-open)
                                          존재 → printf '%s' "$payload" | moai hook pre-tool
cmd/moai/main.go                        main()
internal/cli/root.go                    Execute() → InitDependencies()
internal/cli/hook.go                    runHookEvent(cmd, hook.EventPreToolUse)
                                          ├ HookProtocol.ReadInput(os.Stdin)
                                          │   └ 파싱 실패 → stderr 경고 + 빈 HookOutput + exit 0
                                          │      (파이프라인을 절대 깨지 않는다)
                                          ├ HookEventName 비었으면 서브커맨드 이름 주입
                                          ├ harnessModeIsCodex → codexadapter.Resolve 교차 검증
                                          ├ TaskCreated / Notification은 HOI 마스터 토글로 중앙 게이팅
                                          ├ context.WithTimeout(config.DefaultHookDispatcherTimeout)
                                          └ defer registry.Shutdown()   — 비동기 trace writer 플러시 배리어
internal/hook/registry.go               Dispatch → 등록된 핸들러 체인
internal/hook/pre_tool.go (48KB)        실제 정책 판정
  ├ internal/hook/security/*            ast-grep 기반 보안 스캔
  ├ internal/hook/quality/*             린터·포매터·게이트 요약
  └ internal/permission/stack.go        8-tier 권한 스택 (mvdan.cc/sh 셸 파싱)
internal/cli/hook.go                    writeHookOutputCodex 또는 writeHookOutput → stdout JSON
internal/cli/hook.go                    output.ExitCode == 2 → exitCodeError{2}
cmd/moai/main.go                        cli.ResolveExitCode → os.Exit(2)
```

**계약상 중요한 지점**: `hook.go`의 주석이 명시하듯 **deny 결정에는 exit-2 분기가 없습니다.**
JSON deny는 `hookSpecificOutput` 안에 살고 exit 0으로 나갑니다 — exit 2에서는 stdout JSON이
무시되어 deny가 유실되기 때문입니다.

---

## C. 템플릿 — 임베드에서 사용자 디스크까지

### 빌드 타임

```
internal/template/embed.go                            //go:embed all:templates   (564개 파일)
internal/template/embed.go                            //go:embed catalog.yaml
internal/template/scripts/gen-catalog-hashes.go       별도 main — 카탈로그 해시 사전 생성
```

### 런타임 — `moai init`

```
internal/cli/init.go                    template.NewDeployerWithRenderer(embeddedFS, renderer)
internal/template/deployer.go           NewDeployer(fs.FS, opts...)
                                          프로덕션은 go:embed, 테스트는 fstest.MapFS
internal/template/catalog_loader.go     catalog.yaml 로드
internal/template/catalog_tree_hash.go  스킬 디렉터리 트리 해시 (변경 감지)
internal/template/renderer.go           .tmpl → 플랫폼별 렌더 (Platform=windows 분기 포함)
internal/template/settings.go           .claude/settings.json 생성
internal/template/hook_entries.go       SettingsTemplateName = ".claude/settings.json.tmpl"
                                        ParseHookEntries — (event, matcher, script, if, timeout, async) 튜플 비교
internal/template/skill_mirror.go       심볼릭 링크 우회 미러링
                                          (go:embed가 symlink를 조용히 버리는 문제 대응)
internal/mirrornotice/notice.go         미러 결과를 사용자 알림으로 전환
internal/config/atomicfile/write.go     원자적 쓰기
   ↳ 프로젝트 디스크: .claude/ · .moai/ · .codex/ · .github/ · CLAUDE.md · AGENTS.md …
```

### 런타임 — `moai update` (재배포)

```
internal/cli/update_template_sync.go    NewDeployerWithRendererAndForceUpdate(embedded, renderer, true)
internal/cli/update/plan/plan.go        분석·분류·네임스페이스 보호
internal/cli/update/backup/backup.go    백업 + 로테이션
internal/merge/*                        3-way 머지 (사용자 편집 보존)
internal/cli/update/deploy/deploy.go    배포 + 레거시 마이그레이션
internal/cli/update/report/report.go    사용자 대상 advisory 출력
internal/manifest/*                     provenance 기록
```

---

## D. MCP — 외부 클라이언트에서 내부 코어까지

```
MCP 클라이언트                          .mcp.json 엔트리 (기본 비활성, `moai mcp add`로 활성화)
cmd/moai/main.go → cli.Execute()
internal/cli/root.go                    rootCmd.AddCommand(newMCPServerCmd())
internal/cli/mcp_server.go              newMCPServerCmd — stdio JSON-RPC 루프
internal/cli/mcp_server.go              .moai/config/sections/mcp.yaml 로드 (파일 없으면 전 도구 등록)
internal/cli/mcp_server.go              add(name, mcp.NewTool(...), handler) × 28
internal/mcp/catalog.go                 MoaiMCPTools — 이름·WriteCapable 단일 선언
                                          (가드 테스트가 일치를 강제)

핸들러 → 내부 코어 (CLI와 같은 코어를 공유한다)
  도구             패키지                심볼 / 파일
  session_list      internal/session      QueryActiveWork()
  goal_arm/status   internal/goal         NewGoal() · SaveGoal() · LoadGoal()
  spec_progress     internal/spec         ListDocs()
  verify_snapshot   internal/verify       Load() · RecordCheck()
  graph_*           internal/graph        query · codequery · shortestpath
  codex_* / glm_*   internal/cli          mcp_codex.go · mcp_glm.go → 외부 에이전트 프로세스
  session_msg_*     internal/sessionmsg   —
```

MCP 표면이 CLI와 **같은 코어를 공유**한다는 점이 이 경로의 핵심입니다 — MCP는 별도 구현이 아니라
같은 도메인 패키지에 대한 두 번째 어댑터입니다.
