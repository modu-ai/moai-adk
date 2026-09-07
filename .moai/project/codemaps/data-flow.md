# 데이터 흐름

> `/moai codemaps`로 생성됐습니다. 시스템 동작의 대부분을 실어 나르는 경로를
> 끝에서 끝까지 따라갑니다.

**측정 트리**: worktree `.claude/worktrees/t475`, 브랜치 `WT-codemaps-stale`, HEAD `52f863f36`
**측정**: 2026-09-08

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
internal/hook/pre_tool.go               실제 정책 판정
  ├ internal/hook/security/*            ast-grep 기반 보안 스캔
  ├ internal/hook/quality/*             린터·포매터·게이트 요약
  │   └ internal/hook/quality/step_git_env.go
  │                                     자식 프로세스에서 리포지터리 **위치** 환경변수만 제거
  │                                     (GIT_DIR·GIT_INDEX_FILE 등 — 훅 환경이 cmd.Dir를 이긴다)
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
internal/template/embed.go                            //go:embed all:templates   (581개 파일)
internal/template/embed.go                            //go:embed catalog.yaml
internal/template/scripts/gen-catalog-hashes.go       별도 main — 카탈로그 해시 사전 생성
internal/template/agentemit                           make agents-emit
                                                        .claude/agents/moai/*.md × agents-codex.yaml
                                                        → .codex/agents/moai/*.toml (fail-closed)
internal/template/commandemit                         make commands-emit
                                                        .claude/commands/moai/* (읽기 전용)
                                                        → .agents/skills/moai-<command>/SKILL.md
                                                        (본문 바이트 동일 verbatim)
```

두 방출기는 **빌드 타임에만** 돕니다. 런타임 배포 경로는 그 산물을 다른 템플릿 파일과
구분하지 않고 나릅니다 — 그래서 발행 스킬 경로에 별도 보호가 필요해집니다(아래).

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
internal/template/published_skills.go   발행 스킬 경로(.agents/skills/moai-<command>/SKILL.md)에
                                          한해 update 모드에서도 init 모드 provenance 유지 →
                                          사용자 소유 파일이 살아남고, 건너뜀은 보고된다
internal/template/skill_mirror.go       심볼릭 링크 우회 미러링 (Deploy 마지막 단계)
                                          (go:embed가 symlink를 조용히 버리는 문제 대응)
internal/mirrornotice/notice.go         미러 결과를 사용자 알림으로 전환
internal/config/atomicfile/write.go     원자적 쓰기
   ↳ 프로젝트 디스크: .claude/ · .moai/ · .codex/ · .agents/ · .github/ · CLAUDE.md · AGENTS.md …
```

### 런타임 — `moai update` (재배포, 그리고 재배포가 없는 경로)

```
internal/cli/update_template_sync.go    NewDeployerWithRendererAndForceUpdate(embedded, renderer, true)
internal/cli/update/plan/plan.go        분석·분류·네임스페이스 보호
internal/cli/update/backup/backup.go    백업 + 로테이션
internal/merge/*                        3-way 머지 (사용자 편집 보존)
internal/cli/update/deploy/deploy.go    배포 + 레거시 마이그레이션
internal/cli/update/report/report.go    사용자 대상 advisory 출력
internal/manifest/*                     provenance 기록

── 버전이 일치해 Deploy 앞에서 조기 반환하는 경로 ──
internal/cli/update_mirror_heal.go      그 조기 반환 자리 옆에서 실행
  └ internal/template/skill_mirror_repair.go
                                        Deploy 없이 .agents/skills 두 생산자 결과를 복구
                                        (패키지 수준 함수 — DeployerOption 이었다면 배포 경로에서도
                                         살아나 수리 기능의 부작용으로 배포 동작이 바뀐다)
```

**이 갈래가 존재하는 이유**: `.agents/skills`의 두 생산자가 **모두 Deploy 안에** 삽니다.
버전 일치 update는 Deploy 앞에서 반환하므로, 그것만으로는 지워진 미러가 그 프로젝트에서
영구히 복구되지 않습니다. 존재 게이트는 프로젝트의 **기록된 배포 버전**이며, `.agents/` 의
존재(필요할 때 정확히 사라져 있다)도 `.codex/` 배선의 존재(claude 전용 프로젝트에는 기본
부재)도 게이트로 쓰지 않습니다.

---

## D. MCP — 외부 클라이언트에서 내부 코어까지

```
MCP 클라이언트                          .mcp.json 엔트리 (기본 비활성, `moai mcp add`로 활성화)
cmd/moai/main.go → cli.Execute()
internal/cli/root.go                    rootCmd.AddCommand(newMCPServerCmd())
internal/cli/mcp_server.go              newMCPServerCmd — stdio JSON-RPC 루프
internal/cli/mcp_server.go              .moai/config/sections/mcp.yaml 로드 (파일 없으면 전 도구 등록)
internal/cli/mcp_server.go              add(name, mcp.NewTool(...), handler) × 29
internal/mcp/catalog.go                 MoaiMCPTools — 이름·WriteCapable 단일 선언 (29개)
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

---

## E. 상태 앵커 — 렌더가 어느 프로젝트 루트에 쓰는가

statusline 렌더 한 번은 텔레메트리를 쓰고, 보드 루트를 잡고, goal 상태를 읽습니다. 이 셋이
**같은 앵커**를 써야 하며, 그 앵커를 정하는 자리가 하나로 모여 있습니다.

```
Claude Code                             statusLine 훅 → stdin JSON (workspace.*, worktree.*)
internal/cli/statusline.go              렌더 진입
internal/statusline/state_anchor.go     resolveStateAnchor(...)  ← statusline 쪽 어댑터
  └ internal/stateanchor/stateanchor.go 우선순위 사슬 (요건으로 고정 — 삽입·재정렬은 요건 변경)
       1) stdin workspace.project_dir
       2) worktree.original_cwd
       3) internal/core/git 의 공통 디렉터리 해석 → 그 부모
          (리포지터리의 모든 체크아웃·워크트리에 대해 하나인 루트)
       ↳ 셋 다 실패하면 ""  — 호출자는 상태 쓰기·읽기를 건너뛰고 렌더는 정상 완료
소비자
  B1  텔레메트리 쓰기            .moai/state/… (앵커 아래)
  B2  보드 루트 + landed·github-counts 소비자
  B3  goal 상태 읽기
  B4  CLI 설정 캐시 사슬          internal/cli
```

**세션의 현재 디렉터리는 앵커가 아닙니다** — 3단계의 walk-up 입력으로만 쓰입니다. 이것이
GH #1694의 수리 지점입니다: 이전에는 텔레메트리 쓰기가 `workspace.current_dir`에 앵커돼
**cd 한 디렉터리마다 `.moai` 디렉터리가 하나씩** 남았습니다. 표시 이름 유도는 별개
관심사로 statusline의 `extractProjectDirectory`에 그대로 남아 있습니다.

---

## F. 워크트리 계보 — spawn 에서 재개까지

```
세션 시작 / worktree 진입              internal/hook 세션 훅
internal/chain/node.go                 EventType — spawn 경계 / session_id 백필 / 완료 엣지
internal/chain/populate.go             이벤트 구성
internal/chain/store.go                Store.Append — 매 쓰기마다 O_APPEND 로 열고 닫는다
                                         (커널이 동시 append 를 직렬화, read-modify-write 없음)
   ↳ .moai/state/chain/events.jsonl    append-only JSONL 계보 트리
internal/chain/prune.go                보존 정책에 따른 정리
internal/cli (moai chain)              조회 표면
```

읽기는 **깨진 줄에 관대**합니다 — 손상된 한 줄이 스트림 전체를 무효화하지 않습니다.
이 원장이 존재하는 이유는 depth-N 워크트리에 `/clear` 이후 재진입한 사람이 grep·스크롤백
고고학 없이 origin·완료·재개 지점을 즉시 복원할 수 있게 하는 것입니다.

---

## G. 설정 쓰기 — 두 경로, 그리고 병합 전 단정

콘솔·TUI가 설정을 저장할 때 경로가 둘로 갈립니다.

```
internal/settings/*                     두 표면(moai web 콘솔 / moai profile setup TUI)이
                                          공유하는 설정 스키마
  ├ Save() 경로가 있는 섹션              ConfigManager.Save() — typed struct 재직렬화
  │                                       (주석과 미모델링 키를 파괴한다)
  └ Save() 경로가 없는 8개 섹션          internal/settings/yamlpatch
      workflow · harness · ralph ·         gopkg.in/yaml.v3 노드 트리 수술로 대상 스칼라만 upsert
      research · feedback ·                → 주석 · 미모델링 키(team.patterns, role-profile effort) ·
      observability · security · db          키 순서 보존. 노드 삭제는 지원하지 않는다
                                          byte-stability 는 보증이 아니라 검증 대상 —
                                          섹션별 골든 round-trip 테스트가 그 범위를 고정한다
```

그리고 **병합 직전**, 워킹 트리의 tracked `.claude/settings.json` 이 손대진 채로 창에
들어가는 것을 막는 단정이 따로 돕니다.

```
moai integration acquire                창을 기록하기 전에 (precondition)
moai integration preflight [경로]        창을 잡지 않고 같은 질문만
  └ internal/cli/integration_settings_drift.go   CLI 절반 — 두 표면의 배선
      └ internal/kanban/settings_drift.go        도메인 절반 — 검출 · 보존 · 원장
            git --no-optional-locks status --porcelain -- .claude/settings.json
              (--no-optional-locks 는 필수다: 평범한 status 는 인덱스 WRITE 락을 수십 ms 잡아,
               레인 여럿이 도는 머신에서 병합 직전 검사가 스스로 경합을 만든다.
               이 플래그는 출력도 반환값도 바꾸지 않고 그 부작용만 없애므로,
               동작이 아니라 실행자가 실제로 받은 argv 로 고정된다)
       ↳ 적중 시: 사본을 .moai/state/settings-drift/ 아래 보존 + ledger.jsonl 한 줄
```

**검출·보존·원장은 설정과 무관하게 매번 돕니다.** 거절만 opt-in이며, 어떤 경우에도
자동 복원하지 않습니다 — 그 파일은 런타임이 쓰고 토큰·절대경로를 담을 수 있어 자동 복원
자체가 데이터 파괴이기 때문입니다.
