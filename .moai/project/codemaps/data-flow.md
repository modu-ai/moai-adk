# 데이터 흐름

## 현재 최종 통합 트리의 흐름 — e65b3b

기준은 `e65b3b469c0ee71195b0b568b4b66d0b364b6f3d`다. 다음은 실제 호출부와 저장 경로의 소스 대조이며 외부 실행·OS 런타임의 성공 판정이 아니다.

1. UserPromptSubmit의 기본 2초 bind context → `registerFactoryHookPeerRun` → 일반 등록 또는 rebound 등록의 `OpenWithContext` → `FactoryDirContext` → `ProjectDirContext` → `CanonicalProjectRootContext` → `internal/core`의 checkout.go에 있는 `ResolveGitDirsContext`로 예산이 전달된다. broker는 caller 잔여 예산과 기본 5초 중 작은 값을 사용하고 경로 탐색 뒤 남은 예산으로 DB busy timeout을 구성한다. inbox의 200ms inspection open은 별도다. 기존 contextless API도 남아 있다.
2. initializer의 template 배포 → 기본 AGENTS.md → validator의 존재 진단으로 이어진다. InstructionsLoaded는 AGENTS 우선·legacy CLAUDE fallback으로 anchor와 import closure를 관측한다. Codex contract는 AGENTS.md·AGENTS.local.md를 보호하며 legacy root 파일을 읽거나 쓰지 않는다. Codex launcher의 local producer는 AGENTS.local.md와 CLAUDE.local.md를 순서대로 둘 다 읽어 하나의 override로 연결한다.
3. Factory는 저장된 after·bundle 순서를 먼저 보존하고 여러 hub의 dependency를 합친다. 순환을 닫는 추론 edge는 넣지 않는다. 같은 wait 판정이 selection과 direct nomination으로 이어지고, 현재 hold/queued 상태의 assigned row는 바로 임대하지 않는다. 기존 hint는 생성 필드로 덮지 않는다.
4. PR 전달은 readiness 전후 tip 대조 → merging 재시도의 holder·expiry 검사 → 원격 변경 직전 state·version·holder·expiry 재확인 → 확인한 SHA push → PR head 대조 → match-head auto-merge 요청으로 이어진다. candidate-tip remeasure와 complete T16의 실제 merge-tree remeasure는 기존의 서로 다른 증거 시점을 유지한다.
5. init/update → `internal/cli/user_asset_phase.go` → 사용자 잠금·기록된 bundle selection·embedded catalog/tree → `internal/userassets` Installer로 이어진다. 네 사용자 루트 설치와 프로젝트 자산 이행 확인이 먼저이며, project deployer는 네 공통 skill/agent 루트를 제외한다. AGENTS 기본 instruction 배포와 사용자 자산 설치는 서로 다른 단계다.
6. protected-zone 판정은 shell quote 문법별 literal 복원 → OS별 native 경로와 lexical 비교형 분리 → 실제 구성요소 walk로 이어진다. POSIX literal backslash와 Windows separator를 같은 것으로 바꾸지 않는다. roster sweep은 dated reports를 live roster에서 제외한다.

7. Stop review gate의 resolved scope → tree key → 공유 영수증 Load·CheckReceipt(24시간 TTL) → cached block/allow 또는 live RPC로 갈라진다. live 결과는 RecordReceipt로 기록하고 fail 상세를 같은 key의 로컬 파일에 보존한다. 일반 fail은 exit1, inconclusive는 exit2로 기록한다. runtime-config-only tree finding의 통과 처리는 게이트의 별도 disposition이며 producer와 모든 매핑이 같다고 보지 않는다.
8. SPEC 없는 카드의 verdict 파일 → `internal/auditverdict`의 Parse → backend·손상 receipt 거부 → 공통 Admit의 PhaseSync 판정으로 이어진다. convergence overall fail도 공통 술어에서 거부하며, HEAD 또는 evidence-directory-only drift의 SHA 증거는 `internal/factorylane`에서 별도로 확인한다.
9. packageManager의 Bun 선언 → repository 경계까지 로컬 node_modules tool 확인 → 해결된 tool은 `bun x` argv(타입 검사는 `tsc --noEmit`)로 실행한다. 미해결 tool은 예상 로컬 경로의 mandatory 실행으로 실패하고, 명시 override와 비Bun npx 흐름은 바꾸지 않는다.

> `internal/template/templates/CLAUDE.md`와 plugin 운반체는 퇴역했다. 아래 이전 미러·생산자 설명은 현재 제공 흐름이 아닌 이력이다.

## 이전 c572 기준의 흐름

기준은 `c572e7baceaa6fd0cd3c78a9335b4baabd320347`다. 아래는 호출부·저장 경로의 소스 대조이며 외부 프로세스나 원격 병합을 실제 실행했다는 주장은 아니다.

1. `internal/cli/integration.go`의 acquire/wait가 `internal/factory/integration_window_queue.go`의 직렬 RMW와 `internal/factory/integration_window_ops.go`의 FIFO 승격으로 이어진다.
2. `internal/factory/integration_remeasure.go`가 트리에 묶인 실행 기록을 만들고 검증한다. `internal/factory/integration_merge_step.go`는 고정한 SHA와 소유권·임대·작업 트리·충돌을 재확인해 병합한다.
3. `internal/cli/factory_card.go`의 complete가 T14/T16 전이를 만들고, `TransitionRequest.VerifyRemeasure` 콜백을 통해 `internal/homestate/card_evidence_readers.go`가 병합 트리의 재측정 기록을 확인한다. 단순 merge SHA 텍스트 파일을 검증 근거로 대신 읽는 흐름이 아니다.
4. Codex의 대화형 부모가 `moai todo --auto`를 수행하면 `internal/cli/todo_auto_lane.go`가 순위 결정 → 지명 임대 → 카드 워크트리 → 증거 인계를 수행한다. 공유 순환의 quota hold는 Claude 레인에만 적용되고, 이때는 자기 할당 카드만 임대 관문을 거친다. Codex/GPT의 할당량 평가는 false를 반환한다. 이는 T8a 감사 승인에서 사용하는 별도의 `QueueHold`와 구분한다.
5. 워크트리 착지 판정은 `internal/cli/worktree/landing_predicate.go`의 `landingExactChangedPaths`에서 native object/mode를 확인한다. CLI의 `session_worktree.go`도 같은 helper를 호출한다.
6. `internal/graph/card_file.go`가 모든 도달 가능한 부모의 카드 귀속을 모으고 native Git 배치로 각 착지 커밋의 첫 부모와 파일 차이를 계산한다. NUL 구분 경로와 마지막 root sentinel이 완전하지 않으면 개별 diff로 돌아간다. 결과는 `internal/graph/graph.go`의 그래프와 `internal/graph/meta.go`의 fingerprint에 쓰인다.

`internal/cli/init.go`의 MCP provisioning 실패는 collector에 전달되고, `internal/hook/session_start_memory_budget.go`는 goroutine 시작 전에 읽기 의존성을 고정한다. 기존 흐름의 과거 구현 설명은 아래에 당시 기준과 함께 남긴다.

update의 이전 루트 거부 규칙 정규화는 `internal/cli/update_deny_migration.go`의 정확한 9종 치환으로 수행한다. binary/dry-run 반환 뒤 같은 버전의 조기 반환 전에 호출하며, clean-install에서는 보존한 설정에 적용한다. 사용자 자산 설치 → 프로젝트 자산 이행의 기존 흐름은 아래 설명대로 유지된다.

## 이전 ff7722 기준의 판정과 관측 흐름

기준은 `ff7722d2d157dd4e3cffd88ebb644e0f8ead83fa`다. 사용자 폴더 설치 흐름은 이전 PR 기준과 같고, 다음 경로를 소스에서 추가 대조했다.

- 착지: ancestry 성공이면 즉시 반영된 것으로 판정한다. ancestry가 아닌 경우 누적 patch-id 후보 일치 → `internal/cli/worktree/landing_predicate.go`의 landingExactChangedPaths → base·tip·ref의 NUL 구분 ls-tree → 변경 경로의 native object/type/mode·삭제 상태 비교로 이어진다. done/sweep은 이 경로로 확인되지 않으면 gh의 merged PR 확인도 사용한다. 세션 종료는 gh를 호출하지 않고 확인되지 않은 트리를 보존한다.
- 카드·파일: `internal/graph/card_file.go`의 walkCardCommits가 HEAD에서 도달 가능한 모든 부모 경로의 커밋 subject를 attribution → merge·squash landing의 first parent와 NUL 구분 파일 diff → CardFileEdges → `internal/graph/graph.go`의 edge 계층. 같은 landing 목록은 CardAttributedMergeSHAs → CardMergeFingerprint → `internal/graph/meta.go`의 freshness 값으로 흐른다. root 또는 비교 실패 커밋은 파일 edge를 만들지 않는다.
- init MCP 오류: `internal/cli/init.go`의 provisionMCPEntryUnlessDeclined 오류 반환 → runInit의 p.Collect → deferred emitSummary 한 번. 이 오류는 요약으로 전달하고 init 실패로 승격하지 않는다.
- update 표시: 진입 시 registry reset → 실행 중 Require/Reference 행 수집 → defer로 `internal/cli/update_action_block.go` 종료 블록 출력. severity glyph는 `internal/cli/severity_line.go`가 tui.StatusIcon에서 가져온다.

## PR #1772의 사용자 폴더 설치 흐름

`internal/cli/user_asset_phase.go`는 카탈로그와 임베드 템플릿을 `internal/userassets/install.go`의 Installer에 전달한다. `paths.go`의 설치 루트는 `~/.claude/skills`, `~/.claude/agents`, `~/.agents/skills`, `~/.codex/agents` 네 곳이다. 사용자 manifest는 `~/.moai/user-assets.json`에 두며, 설치 중단 기록과 백업 경로도 같은 사용자 상태 영역에서 관리한다.

update는 사용자 잠금 획득 → 기록된 번들 선택으로 Install → 새 manifest의 PruneUnselected → manifest 저장 순서다. 그 뒤 `internal/cli/migrate_project_assets.go`가 프로젝트 파일의 현재 hash·provenance·사용자 대응 파일을 확인하고 프로젝트 템플릿 동기화로 넘어간다. 수정되거나 대응 파일을 확인하지 못한 항목은 프로젝트에 남긴다. 프로젝트 배포기는 공통 자산 루트를 제외하며 버전 일치 조기 반환에서 프로젝트 스킬 미러를 다시 만들지 않는다.

> 아래의 과거 미러 생산자 설명 중 `internal/cli/update_mirror_heal.go`와 `internal/template/skill_mirror_repair.go` 경로는 폐기됐다. 현재 사용자 폴더 설치와 별개의 이전 구현 이력이다.

## 이전 081899 기준의 흐름 보충

기준은 `081899adb825935d5263b1699fe730373deaa4fd`다. 다음은 호출부와 저장 경로의 소스 대조이며 외부 프로세스·원격 병합을 실제 실행했다는 뜻은 아니다.

- `internal/hook/post_tool_scope.go`가 프로젝트와 CWD의 가장 가까운 Git 루트를 물리 경로로 비교한다. 외부임이 확인된 Write/Edit 대상은 `post_tool.go`·`post_tool_guardian.go`의 LSP·AST·guardian 스캔을 생략하고, 판정이 불확실하면 스캔을 유지한다.
- Factory 병합 관측은 `factoryCloseLaneCard`와 `BacklogStore.ArchiveOnRuntimeCompletion`을 거쳐 호출자 루트의 큐 카드 보관과 runtime `completed`로 이어진다. 이미 보관된 카드도 닫힌 상태로 재관측한다.
- 훅 이벤트 append 뒤 `MaybeSpawnRetentionPruner`가 stamp를 읽고, 오래됐거나 없을 때 같은 바이너리의 `hook retention-prune`를 분리 실행한다. 플랫폼별 실행은 `retention_spawn_unix.go`·`retention_spawn_windows.go`가 맡는다.
- 감사 시작 키는 `auditreceipt.StartMarkerKey`가 agent ID 우선으로 만들고, 없으면 session/role에서 파생한다. 파생 키는 가장 이른 시작을 유지하며 단일 stop으로 삭제하지 않는다.
- `internal/graph/card_file.go`의 카드·파일 연결은 `graph.go`의 그래프 생성과 `meta.go`의 병합 출처 fingerprint에 쓰인다. 웹 `TodoGraph`는 별도로 `internal/web/todo_queue_read.go`의 `readTodoGraph`를 통해 큐·보관 카드와 findings의 카드 간 관계를 읽는다.

> **t1510 판(card t1510)** — § N이 하나 더 붙었다: **보호 구역 거부 흐름**. PreToolUse(`moai hook pre-tool`)가 `harness-learner` identity의 Write·Edit·Bash를 받으면 `protected_zone_guard.go`(파일 도구)와 `protected_zone_shell.go`(Bash — mvdan/sh AST 위의 가능-디렉터리-집합 워커)가 대상 경로를 `protected_zone_path.go`의 물리 해석(끊긴 심링크 목적지 추종 포함)으로 정규화해 shipped+overlay 매니페스트와 compiled baseline에 대조하고, 적중 시 `HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION` sentinel과 사람 회송 이유로 거부하며 감사 행을 `.moai/logs/protected-zone-audit.jsonl`에 남긴다. 매니페스트가 invalid면 변이 명령을 읽지 않고 거부한다(fail-closed).
> **t1524 판(card t1518)** — § N이 새로 붙었다: `.moai` 위생 엔진의 두 흐름(감사 로그 회전 한 판, 끝난 세션 상태 GC의 생존 판정→연대→재판정→삭제). 신규 패키지 `internal/hygiene`(card t1518, SPEC-MOAI-HYGIENE-001) — SessionStart 자동 경로는 `internal/hook/session_start_hygiene.go`가 best-effort로 운반하고(기동을 막지 않는다), 수동 표면은 `moai clean --audit-logs|--session-state`다. 두 경로 다 기본은 report — 파일을 하나도 바꾸지 않는다. 같은 창의 세션 비상 재개 경로(`lane_resume.go` · card t1465)는 § F 계열 런처의 예비 경로고, 감사 상한(card t1500)은 § I의 판정 위에 반복 상한을 얹는다.

> **t1443 판(card t1347·t1375)** — Factory 흐름에 두 관문이 더했다: ① **할당량 게이트** — `factory next` 임대 전 `internal/cli/factory_quota.go`가 상태 디렉터리의 사용량 창 원장으로 보유 창·압력을 평가한다(원장은 `internal/statusline/context_usage.go`가 스키마 v3로 쓰고 `internal/statusline/quota.go`의 `AggregateQuota`가 읽는 것과 같은 것). 판정은 배차 스티어링 행(`factory_quota_lanes.go`)과 `--auto` 사이클 안내 줄로 흘러 임대 보유를 설명한다. 게이트 설정은 `workflow.quota_gate.*`(`loader_quota_gate.go` — 모든 실패에서 꺼짐 기본). ② **관리 세션 divert** — `MOAI_FACTORY_MANAGED`가 명시적으로 켜진 런치는 `launcher.go`·`codex_launcher.go`의 게이트에서 관리 소유자로 갈라진다(`managed_factory_session.go`의 stream-json Claude 자식 · `managed_codex_factory.go`의 websocket Codex App-Server 자식 — 런처가 대화를 소유한다). 기본은 꺼짐이다. ③ **llm.yaml 셋째 쓰기 경로**(card t1411) — § G의 두 갈래 밖에 `internal/settings/llmoverrides.go`가 더했다: 웹 에이전트 설정 탭의 저장이 llm.yaml 프로파일·에이전트별 model/effort를 원자 쓰기+스냅샷/복원으로 쓴다. ④ **할당량 판독 확장**(card t1442 — t1347 부채 F1) — 게이트 판독이 primary 하나가 아니라 primary+링크된 워크트리 상태 디렉터리 전부로 넓었다(`internal/statusline/quota_dirs.go`의 `QuotaStateDirs` — git 메타데이터 파일 읽기만으로 열거, `workflow.quota_gate.max_scan_dirs` 바운드; 다른 형태의 상태 디렉터리는 단독 판독 — fail-open). ⑤ **스테일 런 리바인딩**(card t1345) — 프롬프트마다 env 네임 런의 활성을 재측정해 비활성이면 레인을 살아 있는 런으로 재결합하거나(`internal/hook/factory_rebind.go`) `moai factory relaunch` 실행 명령줄을 운영자에게 안내한다.

**이전 갱신 — t1295, `develop` `cee197917` (2026-09-28).**
MoAI가 만드는 L1 워크트리 경로와 기존 트리의 이전 경로를 § M에 추가했다.
Factory 런 은퇴의 `lead` 표기는 이전 런의 저장 역할값이다. 현재 런은 `leader/lane`을 쓰며,
옛 역할을 가진 세션은 `internal/hook/session_stale_run.go`에서 재등록하지 않고 안내한다.

**이전 갱신 — t1187, `origin/develop` `a8a9b9376` (2026-09-25).**
앵커 `bd71c59e4` 뒤 끝점 변경 41개 비테스트 Go 파일을 확인했다.
이번 판에서 아래 두 경로를 보강했다.

### Codex 읽기 전용 감사

`moai codex audit`와 MCP `codex_role_audit`는
`internal/cli/codex_audit_launch.go`의 `prepareCodexAudit`으로 합류한다.
호출자의 워크트리·역할·보고서 경로·지시문 길이를 검사한 뒤
`codex exec -s read-only` 별도 프로세스를 시작한다. MCP 경로는
`codex_audit_mcp.go`의 서버 메모리 job 표에 ID를 두고 즉시 반환한다.
`codex_role_audit_status`·`codex_role_audit_result`가 실행 결과와
기록 경로를 읽는다. 런처가 결과 원문과 시작 기록을 쓴다.

### Factory 런 은퇴

`internal/factorymsg/factory_run_retire.go`가 등록된 이전 런의 `lead` peer PID와
프로세스 시작 지문을 이전 런의 fallback 신원으로 건넨다.
`internal/homestate/factory_run_retire.go`는 소유자를 분류하고,
정상적인 신원이 없는 런에 대해서는 부팅 시각보다 모든 기록 활동이
이전인지 확인한다. `OwnerDead`가 양성으로 확인된 런만 `retired`로
바꾸고 `run.retired` 이벤트에 증거 근거를 남긴다. 살아 있음 또는
판정 불명확 상태는 은퇴시키지 않는다.

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
  └ internal/factory/todo_root.go        ResolveTodoQueueRootAdopting
                                          — primary checkout 정규화 → project-key 산출
internal/factory/state_dir.go            StateDirForRoot
                                          → ~/.moai/db/<project-key>/todo
internal/cli/todo.go                    todoBacklogPath() → factory.BacklogPathForRootAdopting
internal/cli/todo.go                    newTodoStore() → factory.NewBacklogStore(path)
internal/cli/todo_analysis.go           appendAnalyzedCard(rec, text, BacklogStateQueued, force)
internal/factory/backlog_store.go        NewBacklogStore / Mutate(func(*BacklogRecord) error)
internal/factory/state_lock.go           크로스 프로세스 backlog.lock (+ _unix / _windows)
internal/factory/backlog_store.go        openEngine (정의 위치) → backlog_sqlite.go 엔진
internal/factory/backlog_sqlite.go       WAL + busy_timeout ≥ 5000 + BEGIN IMMEDIATE
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

**t1274 판 추가 — 계약 모드 이탈 관측 옆길(card t1235, SPEC-AUTONOMY-ESCALATION-001).**
같은 훅 디스패치에서 갈라져 나가는 관측 전용 옆길이 하나 더 있다. `internal/hook/escalation_observe.go`가
internal/escalation 패키지의 `Active` 게이트(contract 모드인가?)를 지나 계약 이벤트를 조립해 같은 패키지의 `Observe`로
넘긴다. 이 옆길은 **위의 종료 코드 계약에 영향을 주지 않는다** — 감지기는 아무 값도 반환하지 않고(결정 불개입),
고장(panic·입력 불가·락 시간 초과)은 `not-checked` 로그 한 줄로만 남는다. 산출물은 두 갈래:
에스컬레이션 기록 `<worktree>/.moai/reports/<card>/escalation/<class>-<fingerprint>.md`(재발생은 새 ordinal),
해시 체인 카드 로그와 상태 캐시 `$MOAI_HOME/db/<project-key>/contract/escalation/<card>.{log.jsonl,json}`.
기록·해시에 들어가는 명령 텍스트는 `MaskCommand`로 자격증이 마스킹된 뒤다.

---

## C. 템플릿 — 임베드에서 사용자 디스크까지

### 빌드 타임

```
internal/template/embedemit                          Git 추적 경로 ApprovedPaths → Render
internal/template/embed_manifest_gen.go               개별 템플릿 지시문 604개 + catalog.yaml
                                                      → embeddedRaw
internal/template/embed.go                            → EmbeddedTemplates
internal/template/scripts/gen-catalog-hashes.go       별도 main — 카탈로그 해시 사전 생성
internal/template/agentemit                           make agents-emit
                                                        .claude/agents/moai/*.md × agents-codex.yaml
                                                        → .codex/agents/moai/*.toml (fail-closed)
internal/template/commandemit                         make commands-emit
                                                        .claude/commands/moai/* (읽기 전용)
                                                        → .agents/skills/moai-<command>/SKILL.md
                                                        (본문 바이트 동일 verbatim)
```

방출기는 **빌드 타임에만** 돕니다. 런타임 배포 경로는 그 산물을 다른 템플릿 파일과
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
internal/cli/update/plan/plan.go        분석·분류·네임스페이스 보호 (t1547 판부터 관리 뿌리 파일도 분석 대상)
internal/cli/update/reconcile_classify.go   소유권 분류기 — 4부류 (template-owned/user-modified/user-owned/stale)
internal/cli/update/reconcile.go        조정 파이프라인 — 갱신·3-way 병합·무접촉 보존·아카이브 후 삭제
                                          (t1547 판 — 기본 경로의 wipe-first 대체,
                                           충돌은 `<경로>.moai-new[.N]` 사이드카로 보고)
internal/cli/update/backup/backup.go    백업 + 로테이션
internal/cli/update/backup/file_snapshot.go   파일별 base 스냅샷 기계 (settings.json과 .mcp.json이 공유)
internal/cli/update/backup/mcp_snapshot.go    .mcp.json 전용 — 배포가 실제로 쓴 렌더를 base로 기억
internal/merge/*                        3-way 머지 (사용자 편집 보존)
                                          이 판에서 머지 결과가 RetainedKeys를 함께 돌려준다 —
                                          새 템플릿이 더 이상 안 들고 온 키의 dotted 경로
internal/cli/update/deploy/deploy.go    배포 + 링크 처분(DisposeSymlinks) + 레거시 마이그레이션
internal/cli/update/report/report.go    사용자 대상 advisory 출력
internal/cli/update/report/outcome.go   조정 요약 — 5범주(refresh/merge/conflict/preserve/archive-removed)
                                          집계 + 경로 목록 (t1547 판 — 삭제가 요약에 숨지 않는다)
internal/manifest/*                     provenance 기록

── 버전이 일치해 Deploy 앞에서 조기 반환하는 경로 ──
프로젝트 스킬 미러 복구 없음            사용자 폴더 설치가 공통 자산을 담당
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
  ├ Save() 경로가 있는 6개 섹션         ConfigManager.Save() — 단, 섹션별 dirty 게이트
  │   git-strategy · git-convention       (t1351 판, SPEC-WEB-SAVE-LOSSLESS-001): 이번 세션에
  │   user · language · quality · llm      SetSection 으로 변형된 섹션 파일(또는 애초에 없는
  │                                        파일)만 다시 쓴다 — 미변형 파일은 통째로
  │                                        재마샬되지 않아 주석·미모델링 키가 살아 남는다
  ├ 스칼라 편집(devMode·convention)      internal/settings/projectscalars.go — 웹·TUI 공유
  │                                        yamlpatch 라인 스플라이스 seam(빈 제출값은 기존
  │                                        영속값을 덮지 않는다)
  └ Save() 경로가 없는 8개 섹션          internal/settings/yamlpatch
      workflow · harness · ralph ·         gopkg.in/yaml.v3 노드 트리 수술로 대상 스칼라만 upsert
      research · feedback ·                → 주석 · 미모델링 키(team.patterns, role-profile effort) ·
      observability · security · db          키 순서 보존. 노드 삭제는 지원하지 않는다
                                          byte-stability 는 보증이 아니라 검증 대상 —
                                          섹션별 골든 round-trip 테스트가 그 범위를 고정한다
```

이 그림의 첫 갈래는 역사가 있다. `Save()`의 typed struct 재직렬화는 원래 주석과 미모델링 키를
파괴했고(GitHub issue #1731 — quality.yaml 의 `constitution.session_effort_default` 같은 키가
devMode 편집 한 번에 사라졌다), 그 결함의 수리가 지금의 모양이다: git-strategy 선례였던 섹션별
dirty 게이트가 여섯 Save() 섹션 전부로 확장됐고(`config/manager.go` — 변형 추적은 `Load`·`Reload`·
성공한 `Save()` 에서 리셋된다), 재마샬이 남아 있던 마지막 자리들도 스플라이스로 갈렸다 — 프로필
동기화(`internal/profile/sync.go`)는 user·language 를 **행 치환**으로, 웹·TUI의 devMode·convention
편집은 `projectscalars.go` seam 한 곳으로 모인다. 설정 파일의 바이트 대부분은 이제 어느 편집도
만지지 않는다.

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
      └ internal/factory/settings_drift.go        도메인 절반 — 검출 · 보존 · 원장
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

### F1 자가 배차 루프 (t1351 판, SPEC-FACTORY-SELF-DISPATCH-001)

lane 세션이 운영자 개입 없이 다음 카드를 집는 순환입니다. 모든 진입은 같은 승인·거부 술어
(`factoryLaneAdmission`·`factoryLaneRefusal` — CLI·MCP·todo 가드가 공유)를 통과합니다.

```
moai cc|glm -f lane [--clear-policy each|when-full|relaunch]
  └ 런처가 lane 스탬프를 환경에 심는다 (역할·라벨·백엔드·정책)
moai factory next (CLI 또는 MCP factory_next)
  ├ factoryNextLeaseOnce          카드 임대 + 큐 승격 (merge-ready 스킵·백엔드별 스킵)
  ├ factoryEnsureCardWorktree     카드 워크트리 보장 — 만드는 유일한 경로
  └ homestate RecordCardWorktree  카드 기록에 트리 경로 남김 (다른 트리 이동은 거절)
  → 레인 세션이 그 트리에서 일한다 (SessionStart 룰이 다음 카드 절차를 싣는다 — REQ-SD-019)
moai factory stage                엣지 적용 + 임대 갱신 (병합 준비→병합 중 엣지는 거부)
moai factory complete             통합 브랜치 해석 → merge --no-ff → 병합 기록 → 종료 문장
  └ clear-policy 별로: each = /clear 요청, when-full = 문턱 도달 시에만, relaunch = 세션 종료
     후 런처(factory_lane_relaunch.go)가 다음 카드로 새 세션을 기동
```

정책 기본값은 `clear-each`다. Codex 레인은 정책을 받지 않고 소유 카드 룰로 움직이며,
`moai codex` 쪽 공장 진입은 `-f lane` 하나뿐이다. 운영자 결정이 필요한 카드는
`factory decide`(CLI·MCP `factory_decide`)가 기록층에 남긴다 — 임대는 결정 대기 상태에서
잡지 않는다.

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
    internal/factory/gtd_engage.go       EngageGTDItem — 권한·증거 신선도·의존성·레인 조건 판정
    internal/factory/gtd_operation.go    ExecuteGTDOperation — 준비된 operation 실행 후 readback
                                          ↳ ~/.moai/db/<project-key>/todo/backlog.db (GTD 확장 테이블)
  supervise
    internal/mission/supervisor.go      SuperviseAutoMission — 증거 적재 → 정책 검증 → 실행 → readback
    internal/mission/git_owner.go       커밋과 git merge --no-ff 효과를 소유하고 상태 재판독으로 확인
    internal/mission/delivery_owner.go  원격 전달 provider 부재 → ErrDeliveryUnsupported ("provider_unsupported")
    internal/mission/completion_receipt.go  LoadCompletionReceipt — 완료 판정이 읽는 receipt
```

**`internal/mission`의 비테스트 소비자는 `internal/cli/goal.go` 하나**이며, 같은 파일이
`internal/factory`의 GTD 함수도 직접 부릅니다. 미션 텍스트는 셸 명령이나 goal 조건으로 해석되지
않습니다(`--auto` 플래그 도움말: "without condition or shell parsing"). 거버넌스 receipt의
발행자는 `manager-todo`(구 판정 역할의 개명 후 이름)이고 상태는 권고(`GovernanceRecommended`)이며, 효과는 위 소유자
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

---

## M. MoAI 워크트리 생성·기존 트리 이전

`moai worktree new <name>`은 `internal/cli/worktree/new.go`에서
세션 워크트리 L1 생성기로 내려가 프로젝트의
`.moai/worktrees/<name>`을 만든다. Claude Code가 자체 생성한
`.claude/worktrees/<name>`은 별도 위치로 남는다. `moai codex -w <name>`은
`internal/cli/codex_launcher.go`에서 두 위치의 **기존** 트리를 조회한다.
둘 다 있으면 이름만으로 고르지 않고 절대 경로를 요구한다. 없는 이름은 만들지 않는다.

`moai update`는 템플릿 배포 전에 `internal/cli/update_worktree_migration.go`를 호출한다.
이 경로는 Git에 등록된 기존 `.claude/worktrees` 항목만 계획하고,
잠금·활성 세션·사용 중인 프로세스·대상 충돌을 확인해 이동 가능 항목에만
`git worktree move`를 실행한다. 이동 뒤 HEAD·브랜치·상태를 재확인한다.
건너뛴 항목은 원래 위치에 남아 다음 update에서 재시도할 수 있다.

---

## N. `.moai` 위생 — 회전 한 판과 GC 판정

두 단위 모두 `internal/hygiene`이 소유하고(자동 경로 배선은
`internal/hook/session_start_hygiene.go`, 수동 표면은 `moai clean`),
**report가 출하 기본**이다 — report 모드는 후보 파일을 하나도 바꾸지 않고,
회전도 락파일도 만들지 않으며, 단위마다 요약 행 하나를
`.moai/logs/hygiene-audit.jsonl`에 덧붙이는 것이 전부다. 변이는 apply 모드에서만
일어나고, apply 여부를 정하는 것은 경로마다 하나씩이다 — 자동 경로는
`workflow.hygiene.mode` 설정, 수동 경로는 그 호출의 `--apply` 한 길(REQ-HYG-013 —
CLI는 설정을 자기 변이 결정에 쓰지 않는다).

### 감사 로그 회전 한 판

```
SessionStart(자동) · moai clean --audit-logs(수동)
  → Settings 판독(workflow.hygiene 6키 — audit_log_max_bytes 기본 10 MiB,
    kept_rotations 1)
  → ScanSourceSinks — 닫힌 20항목 싱크 레지스트리(sinks.go)와 대조
      미등록 append-only 쓰기 발견 → 적색, 회전 생략
      (새 로거의 수용 경로는 회전이 아니라 레지스트리 갱신이다)
  → lockfile 획득 — 단일 통과 직렬화
      (Windows는 LockFileEx 사이드카 — rotate_lock_unix.go·rotate_lock_windows.go)
  → 싱크별 크기 판정 → 초과 싱크만 회전(keep-1)
      스테일 판정 — 이미 회전된 뒤의 오래된 결정은 건너뛴다
  → report: hygiene-audit.jsonl 요약 행만 (자기 싱크도 apply에서만 회전한다)
```

### 끝난 세션 상태 GC — 생존 판정 → 연대 → 재판정 → 삭제

```
SessionStart(자동, 기본 report) · moai clean --session-state
  → 후보 열거(targets.go) — 클래스별, 최소 연령 7일 바닥
  → 생존 판정(liveness.go) — 세 신호, fail-closed:
      pid 프로브(probe_pid_unix.go·probe_pid_windows.go —
        시작시각 지문을 못 읽는 플랫폼은 미측정)
      48시간 전사본 활동 · 24시간 심박
      긍정 하나 = LIVE · 전부 음성 = DEAD · 미측정 하나 = INDETERMINATE(보존·사유 동행)
  → DEAD만 다음으로. 연대는 내용 연대표(targets.go)에서 얻는다 —
      mtime은 삭제 데이터가 아니고, 연대 불능 클래스(codex-stop-cap)는 보존된다
  → 삭제 직전 재판정 — 판정 시각과 실행 시각 사이에 세션이 살아나면 그른다(D28)
  → 해석된 .moai 루트 아래 심볼릭 링크 성분은 거부하고,
      삭제는 디렉터리 fd에 고정한 루트 핸들로 실행한다
      SPEC 종결 락 클래스는 대상 열거에서 처음부터 뺀다
```
