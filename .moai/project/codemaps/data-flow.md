# 데이터 흐름

> `/moai codemaps`로 생성됐습니다. 시스템 동작의 대부분을 실어 나르는 경로를
> 끝에서 끝까지 따라갑니다.

**최초 측정 트리**: worktree `.claude/worktrees/t592`, 브랜치 `WT-home-state-rollout`, HEAD `e7bd89ee3`, 2026-09-10
**재측정 트리**: worktree `.claude/worktrees/t869`, 브랜치 `WT-codemaps-refresh`, HEAD `a851b205c`, 2026-09-18 — § A의 함수 위치(6개 심볼 정의 파일 대조), § C의 임베드 파일 수, § D의 도구 수와 표, 새 § J(자율 미션과 GTD 큐). § B·E~I의 경로는 이번 변경과 무관해 앞 판을 이어받았습니다.
**정기 재측정**: worktree `.claude/worktrees/t999`, 브랜치 `WT-codemaps-remediation`, HEAD `56c64891a`, 2026-09-20 — 새 § K(감사 영수증)만 더했습니다. § A~J의 경로는 이번 변경분과 겹치지 않아 다시 재지 않았고 앞 판을 이어받았습니다.
**정기 재측정**: worktree `.claude/worktrees/t1069`, 브랜치 `WT-graph-restamp`, HEAD `0314801c2`, 2026-09-22 — 새 § L(Jev 호출 경로)을 더하고, § C에 `.mcp.json` 스냅샷 갈래를, § G에 저장 실패 관측성 단락을 더했습니다. § A·B·D~F·H~J의 경로는 이번 변경분과 겹치지 않아 다시 재지 않았고 앞 판을 이어받았습니다. § C의 임베드 파일 수(588)도 이 트리에서 다시 셌습니다.

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
                                          — primary checkout 정규화 → project-key 산출
internal/kanban/state_dir.go            StateDirForRoot
                                          → ~/.moai/db/<project-key>/todo
internal/cli/todo.go                    todoBacklogPath() → kanban.BacklogPathForRootAdopting
internal/cli/todo.go                    newTodoStore() → kanban.NewBacklogStore(path)
internal/cli/todo_analysis.go           appendAnalyzedCard(rec, text, BacklogStateQueued, force)
internal/kanban/backlog_store.go        NewBacklogStore / Mutate(func(*BacklogRecord) error)
internal/kanban/board_lock.go           크로스 프로세스 backlog.lock (+ _unix / _windows)
internal/kanban/backlog_store.go        openEngine (정의 위치) → backlog_sqlite.go 엔진
internal/kanban/backlog_sqlite.go       WAL + busy_timeout ≥ 5000 + BEGIN IMMEDIATE
                                          ↳ <queue-dir>/backlog.db
```

**주목할 점**: 큐 경로 해석 함수 `BacklogPathForRoot`는
`~/.moai/db/<project-key>/todo/backlog.json`이라는 논리 경로를 반환하고, 저장 엔진이 같은
디렉터리의 `backlog.db`를 정본으로 엽니다. JSON 이름은 구버전 다운그레이드 계약을 위한
호출 인터페이스일 뿐이며, 이전 프로젝트 로컬 `backlog.json`은 정본이 아닙니다. 검증된
명시적 이전 전에는 프로젝트 로컬 `backlog.db`가 롤백 원본으로 보존됩니다.

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
internal/template/embed.go                            //go:embed all:templates   (588개 파일)
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
internal/cli/update/backup/file_snapshot.go   파일별 base 스냅샷 기계 (settings.json과 .mcp.json이 공유)
internal/cli/update/backup/mcp_snapshot.go    .mcp.json 전용 — 배포가 실제로 쓴 렌더를 base로 기억
internal/merge/*                        3-way 머지 (사용자 편집 보존)
                                          이 판에서 머지 결과가 RetainedKeys를 함께 돌려준다 —
                                          새 템플릿이 더 이상 안 들고 온 키의 dotted 경로
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

**`.mcp.json` 스냅샷이 settings.json의 형제가 된 갈래.** 배포가 자기가 쓴 렌더를
staging→promote로 기억해 두면, 다음 update의 3-way 머지 base가 이전 템플릿 값을 가진다 —
그래서 템플릿이 키 값을 바꿨을 때 사용자가 못 건 키에서 그 변화가 보입니다. 이전의 유래 base에서는
템플릿이 **추가**하는 키는 도착하지만 **변경**하는 값은 묻히는, 전달 종류로 갈라지는 결함이었습니다.
`moai init`과 `moai update` 양쪽이 같은 staging/settle을 부르고, 남는 staging 사본은
advisory 한 줄로 보고됩니다. 이 갈래가 **배달하지 않는 것**도 파일 주석이 적어 둡니다 — 템플릿이
키를 **철회**해도 사용자 파일에 남는 셀은 이미 배포된 settings.json 경로와 똑같이 실패하며, 별도로
추적되는 기존 결함입니다.

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
internal/cli/mcp_server.go              add(name, mcp.NewTool(...), handler) × 30
internal/mcp/catalog.go                 MoaiMCPTools — 이름·WriteCapable 단일 선언 (30개)
                                          (가드 테스트가 일치를 강제)

핸들러 → 내부 코어 (CLI와 같은 코어를 공유한다)
  도구             패키지                심볼 / 파일
  session_list      internal/session      QueryActiveWork()
  goal_arm/status   internal/goal         NewGoal() · SaveGoal() · LoadGoal()
  spec_progress     internal/spec         ListDocs()
  verify_snapshot   internal/verify       Load() · RecordCheck()
  graph_*           internal/graph        query · codequery · shortestpath
  codex_* / glm_*   internal/cli          mcp_codex.go · mcp_glm.go → 외부 에이전트 프로세스
  claude_audit      internal/cli          mcp_claude*.go → claude CLI 서브프로세스 (읽기 전용 리뷰)
  audit_multi       internal/cli          mcp_convergence.go — Claude 백엔드는 claude_audit 수행 함수 재사용
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
       3) core/git 의 공통 디렉터리 해석 → 그 부모
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

그리고 **저장이 실패할 때**, 콘솔의 아홉 persistence seam은 하나의 모양으로 실패합니다.

```
moai web 설정 저장                        handleSave — 9개 persistence seam
  ├ 각 seam의 실패                        logSaveFailure(seam, phrase)
  │                                        ↳ stderr 한 줄 — `moai web: ` 접두어 + seam 이름 + 실패 구문만
  │                                          (원시 에러 값은 자격증명 조각을 품을 수 있어 절대 실리지 않는다)
  └ 응답은 500이 아니라 2xx 재렌더        설정 폼은 hx-boosted라 htmx가 2xx가 아닌 본문을 버린다 —
                                           500이면 인라인 슬롯에 렌더한 실패 이유가 브라우저에 도달하지 않았다
                                           ↳ 페이지 본문의 banner가 이유를 실어 같은 모양으로 성공·실패를 처리
```

성공과 실패가 같은 전송 모양을 쓰는 것은 클라이언트 핸들러가 없어도 된다는 뜻이고, stderr
한 줄이 유일한 기계 관측면입니다(접두어가 grep 표면이다).

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

---

## H. HOME SQLite 이전·Factory 인계·프로필 lease

### 명시적 HOME 상태 이전

```
moai migrate home-state                 기본 dry-run: 경로·census·논리 건수·integrity 출력
  └ internal/cli/migrate_home_state.go  source: <project>/.moai/state/todo/backlog.db
                                         target: ~/.moai/db/<project-key>/todo/backlog.db
                                         search: not-applicable (runtime producer 없음)

moai migrate home-state --apply --verified-live
  ├ 현재 HEAD용 검증 증거 확인
  ├ admission lock 획득 + migration marker 설치
  ├ runtime census 1차·2차 모두 zero인지 확인
  ├ ~/.moai/backups/<project-key>/<migration-id>/에 원본+manifest 보존
  ├ SQLite 백업 API로 복사
  ├ integrity, 논리 건수·digest, SHA-256 readback
  └ 성공 뒤에도 프로젝트 로컬 source는 롤백 원본으로 보존
```

SessionStart, MCP 서버, Factory는 상태를 열기 전 같은 admission lock과 marker를 검사합니다.
marker의 소유 PID가 살아 있거나 판정 불명확하면 자동 복구하지 않습니다. 죽은 소유자만
`moai migrate home-state recover`로 정리할 수 있고, `rollback`은 manifest와 해시가 검증된
백업만 받습니다.

### Factory resume 인계

```
producer                              FactoryStore.SaveResume(schema v2)
consumer                              ClaimResume(token, owner PID+fingerprint, TTL)
  ├ pending                           claimed로 전이
  ├ expired + owner dead              새 token으로 재점유
  ├ owner live/indeterminate           거절
  └ 주입 성공                          FinishResume(expected token)
```

token 비교는 ABA를 막는 CAS 경계입니다. 주입과 완료 기록 사이에 프로세스가 죽으면 메시지는
다시 전달될 수 있으므로 이 계약은 exactly-once가 아니라 **at-least-once**입니다. v1 레거시
claim이 자동 판정 불가능할 때만 `moai factory handoff recover-resume`으로 운영자가
`fail` 또는 `requeue`를 명시합니다.

### 전역 프로필 lease

```
launcher                              ~/.moai/run/profile-leases.db에 provisional 생성
child spawn                           child PID+fingerprint로 token CAS transfer
SessionStart                          session ID·project key enrich
SessionEnd                            같은 소유권을 확인하고 release
clean                                 live/indeterminate lease는 보존, provably-dead만 정리
```

SQLite 파일은 `0600`, 상위 HOME 디렉터리는 `0700` 경계를 유지합니다. 이 흐름은 코드상
계약을 설명한 것이며, 실제 운영 backlog 이전이 실행됐다는 완료 표시는 아닙니다.

---

## I. codemaps freshness 게이트 — 비교 가능성을 먼저 판정한다

`moai graph check`는 codemaps 층을 숫자로 판정하지만, 그 숫자는 **두 트리가 같은 이력 창에서
비교 가능할 때만** 존재합니다. 판정은 값 계산 이전에 끝납니다.

```
.moai/project/codemaps/provenance.json   clean 스탬프 commit_sha (dirty면 content fingerprint)
internal/graph/check.go                  checkCodemaps — 아래 순서를 위에서 아래로
  1 verifyStampResolves                    git cat-file -e <sha>^{commit}
      실패 → absent 운반체 + system error   "stamped commit not comparable"
  2 codemapsBodyPresent                    git ls-files (추적 + --others) — provenance.json 제외
      본문 없음 → absent + nil error        C1: 판정된 관측이지 실패한 측정이 아니다 (exit 1)
  3 verifyStampAncestorOfHEAD              git merge-base --is-ancestor <sha> HEAD
      비조상 → absent 운반체 + system error  "unreachable stamp … freshness unmeasured"
  4 resolveContentAnchor                   본문이 마지막으로 실제 바뀐 지점
  5 gitDiffNameList(anchor, described)     described-source-diff, 임계 40
internal/cli/graph_check.go              0 전부 fresh · 1 stale/absent · 2 system error
  writeUnreachableStampRecovery             3에서만 복구 안내를 붙인다
.github/workflows/graph-freshness.yml    같은 조상성을 CI에서 이벤트별 대상으로 검사
  push                 TARGET=HEAD
  pull_request         TARGET=origin/<base_ref>
  pull_request release/*  TARGET=HEAD (merge preview)
```

핵심은 **1과 3이 서로 다른 조건**이라는 점입니다. `git cat-file -e`는 커밋이 해석되는지만
답하고, squash와 rebase는 내용을 보존하면서 원래 커밋을 HEAD 이력 밖으로 밀어냅니다. 객체
존재를 조상성으로 대신 읽으면 그 간극을 가로지르는 diff가 숫자를 만들어내고, 성립한 적 없는
창에 대한 값이 stale로 보고됩니다. 그래서 3에서 실패한 경로는 `Value`·`ContentAnchor`·
`Contribution`·`DrivingPaths`를 하나도 채우지 않습니다.

2가 3보다 앞에 있는 것도 의도된 순서입니다. 본문이 아예 없는 트리는 stale일 대상 자체가 없는
**판정된 관측**이라 nil error/exit 1로 남고, 그 처분은 이 게이트가 바꾸지 않는 기존 계약입니다.
반대로 1은 2보다 앞입니다 — 해석되지 않는 스탬프는 본문 상태와 무관하게 실패한 측정입니다.

복구도 갈립니다. 1의 답은 이력(더 깊은 fetch)이고, 3의 답은 **본문 재생성 뒤 도달 가능한
커밋으로 재스탬핑**입니다. 본문을 그대로 둔 맨손 재스탬프는 3을 통과시키지만 content anchor가
움직이지 않아 값도 그대로입니다 — 그것이 anti-false-green 계약입니다.

---

## J. 자율 미션 — `moai goal --auto`에서 GTD 큐 operation까지

**이 판에서 새로 생긴 경로입니다.** 자연어 미션을 곧바로 실행하지 않고, 봉인된 계약과 정책
대조를 거친 operation만 큐에 적용합니다.

```
internal/cli/goal.go                    goal --auto "<mission>" → 승인 대기 draft
                                          하위 verb: approve · run · status · revoke · resume
  approve
    internal/mission/contract.go        SealMissionContract — 범위·행동·증거·한도를 한 번 봉인
    internal/mission/auto_state.go      SaveAutoMission → 미션 상태 파일 (internal/atomicfile 경유)
  run (한 operation씩)
    internal/mission/policy.go          ValidateMissionDecision(sealed, snapshot, decision, now)
    internal/kanban/gtd_engage.go       EngageGTDItem — 권한·증거 신선도·의존성·레인 조건 판정
    internal/kanban/gtd_operation.go    ExecuteGTDOperation — 준비된 operation 실행 후 readback
                                          ↳ ~/.moai/db/<project-key>/todo/backlog.db (GTD 확장 테이블)
  supervise
    internal/mission/supervisor.go      SuperviseAutoMission — 증거 적재 → 정책 검증 → 실행 → readback
    internal/mission/git_owner.go       커밋과 git merge --no-ff 효과를 소유하고 상태 재판독으로 확인
    internal/mission/delivery_owner.go  원격 전달 provider 부재 → ErrDeliveryUnsupported ("provider_unsupported")
    internal/mission/completion_receipt.go  LoadCompletionReceipt — 완료 판정이 읽는 receipt
```

**`internal/mission`의 비테스트 소비자는 `internal/cli/goal.go` 하나**이며, 같은 파일이
`internal/kanban`의 GTD 함수도 직접 부릅니다. 미션 텍스트는 셸 명령이나 goal 조건으로 해석되지
않습니다(`--auto` 플래그 도움말: "without condition or shell parsing"). 거버넌스 receipt의
발행자는 `mission-governor`이고 상태는 권고(`GovernanceRecommended`)이며, 효과는 위 소유자
어댑터가 적용합니다. 원격 push·PR·보호 브랜치 병합은 설정된 provider가 없으면 일어나지
않습니다 — 이 절의 존재는 원격 전달이 동작한다는 뜻이 아닙니다.

GTD의 사람 쪽 표면은 `moai gtd`입니다(`internal/cli/gtd.go`). `moai todo` 명령 트리를 감싼
두 번째 이름이라 같은 큐·같은 카드 id를 보며, `capture` · `clarify` · `organize` · `reflect` ·
`engage`가 다섯 단계를 따로 기록합니다. `internal/graph/gtd_private.go`는 같은 저장소에서 비공개
관계 투영을 만들고 권한을 검사합니다.

---

## K. 감사 영수증 — 도구 호출이 실제로 있었는가

**이 판에서 새로 생긴 경로입니다.** 앞의 모든 절이 「무엇이 어디로 흐르는가」를 따라간다면,
이 절은 **「흐름이 실제로 일어났는가」를 나중에 확인할 수 있게 만드는 기록**을 따라갑니다.

존재 이유를 `internal/auditreceipt/store.go`의 패키지 주석이 직접 적습니다 — **PASS 판정은
에이전트가 쓴 텍스트이고, 텍스트는 도구가 불렸음을 보일 수 없습니다.** 그것을 기록할 수
있는 것은 런타임뿐입니다.

```
생산 쪽 — MCP 도구 호출
  internal/cli/mcp_codex.go            codex_audit 실행
  internal/cli/mcp_convergence.go      audit_multi 수렴 실행
    internal/cli/mcp_audit_receipt.go  WriteReceipt — 호출 1건당 영수증 1건
      internal/auditreceipt/treeroot.go  TreeRootFromCWD — git rev-parse --show-toplevel (2초)
                                          ↳ CLAUDE_PROJECT_DIR는 의도적으로 무시
      ↳ <treeRoot>/.moai/state/audit-receipts/receipts/<id>.json  (임시 파일 + rename)

소비 쪽 — 훅 가드
  internal/hook/audit_receipt_guard.go
    SubagentStart  WriteStartMarker      ↳ .../starts/<id>.json
    SubagentStop   ParseVerdictLine      판정 줄에서 PASS 여부와 인용을 읽고
                   CheckCitedReceipts    인용된 영수증이 실재하는지 대조
                   WriteRejection        불일치면 ↳ .../rejections/<id>.json
    PreToolUse     ListRejections        거부가 미해소면 manager-develop ·
                                          manager-docs · manager-git spawn을 거절

게이트 스위치
  .moai/config/sections/workflow.yaml → CodexGateRequired
    문자열이 정확히 "required"일 때만 켜진다 (그 밖의 값·부재는 전부 off)
```

**세 이벤트가 하나의 판정을 이룹니다.** `SubagentStart`가 없으면 `SubagentStop`이 대조할
기준이 없고, `PreToolUse`가 없으면 거부가 아무것도 막지 않습니다. 이벤트별로 나눠 읽으면
각각은 멀쩡해 보이므로 한 자리에 적습니다.

**트리 루트 판정이 이 경로의 조용한 실패 지점입니다.** 워크트리 세션에서 `CLAUDE_PROJECT_DIR`는
primary 체크아웃을 가리키므로, 그 값을 썼다면 영수증은 카드 브랜치가 아닌 다른 트리에 쌓이고
가드는 빈 디렉터리를 보며 「영수증 없음」이라 판정했을 것입니다. `treeroot.go`가 그 변수를
쓰지 않고 `git rev-parse`로 직접 묻는 것은 이 실패를 피하기 위한 선택입니다.

`internal/auditreceipt`는 다른 `internal/...` 패키지를 하나도 import 하지 않습니다(표준
라이브러리와 `gopkg.in/yaml.v3`뿐). 기록 형식은 JSONL이 아니라 **기록 1건 = 파일 1개**이며,
파일명은 런타임이 준 id를 sanitize해 만듭니다.

---

## L. Jev — 게이트에서 표시까지

**이 판에서 새로 생긴 경로입니다.** TypeSafe System One 판단 능력이 설정 게이트에서 모델 답,
그리고 그 답이 사람에게 닿는 자리까지 어떻게 흐르는지입니다. 이 경로의 모든 소비자는 현재
**게이트 미실행 상태**입니다 — 측정 게이트(`jevmeasure`)가 아직 실행되지 않았으므로 코드는
있되 배송 기본값에서 도달할 수 없고, 그래야 게이트가 판정을 내리기 전까지 「있음」과 「쓸 수
있음」이 같아지지 않습니다.

```
게이트 — workflow.jev.enabled (기본 false)
  internal/config/defaults.go            코드 기본값의 원천 (템플릿은 false를 문서화할 뿐)
  internal/cli/wizard/questions.go       다섯 번째 init 질문 jev_enabled — init 전용,
                                           --reconfigure에는 도달하지 않는다
  internal/settings/jev.go               SetJevEnabled — 위자드의 진입. 같은 ApplySchemaEdits
                                           seam의 네이밍 진입이지 두 번째 쓰기 경로가 아니다
  internal/web (jev 패널)                같은 키를 콘솔이 렌더·수정

자격증명 — ~/.moai/.env.typesafe
  internal/jevcred                       쓰기·읽기 단일 구현 (glmcred의 형제)
                                           스키마 AllFields() 밖 — 어떤 스키마 순회도 못 읽는다
  internal/defs + internal/paths         파일명 상수와 HOME 경로 (stdlib-only를 지키는 전부)

호출 — internal/jev (표준 라이브러리만)
  게이트 확인                             비활성이면 요청을 조립하지 않는다 (호출자 아래의 내부 게이트)
  ScreenPayload                          payload가 자격증명 형태 토큰을 실으면 보내기 전에 중단
  Authorization Bearer                    자격증명은 시도마다 다시 읽는다 — 교체가 반영된다
  응답                                    불가능은 에러가 아니라 Availability 값
                                           (no-credential · disabled · secret-detected …)
                                           「답 없음」은 「아니오」가 아니라 신호 자체가 아니다

소비자 — 전부 게이트 미실행, 표시 전용
  moai doctor (Jev check)                활성·credential·도달성을 읽기 전용 확인 — 도달성은
                                           TCP 접속·종료뿐, 판정 요청을 보내지 않는다
  moai todo analyze (admission 경로만)   근접 중복 카드에 세 번째 finding 출처(jev)로 기록
                                           재분석 재스윕은 이 seam을 부르지 않는다
  moai jev-suggest (Hidden)              게이트가 꺼진 채로는 안내 한 줄 — 스킬 제안 순위 신호
  moai web Jev 패널                      스위치와 자격증명 필드 (§ `entry-points.md` 웹 콘솔)

측정 장치 — internal/jevmeasure (소비자 0)
  두 언어 팔 (한국어 원문 · 영역 번역)    라벨 붙은 표본을 돌리고
  ConstantBaseline                       상수 응답의 정확도가 기준선이다 — 날정확도가 아니다
  Run → Report → Verdict                 이 보고서가 소비자의 존재 허가를 판정한다
                                           접점 연락은 없다 — Answerer 주입, 살아있는 구현은 jev.Client
```

이 경로를 지배하는 세 성질이 설계의 이유를 말합니다. **불가능한 답은 값이다** — 에러 반환은
전파되고 전파는 어딘가의 비종료가 되므로, 저하가 기본 경로다. **게이트는 패키지 안쪽에 있다** —
호출점 N곳의 게이트는 N곳이 잊힐 수 있지만, 비활성 상태에서 요청 자체가 조립되지 않는 성질은
한 곳에서 검증된다. **표시 전용이다** — 이 패키지들이 파일을 쓰거나 큐를 바꾸거나 git을 만지는
경로가 없어서, 표시가 판정으로 오독되더라도 그 오독이 상태를 바꿀 수 없습니다.
