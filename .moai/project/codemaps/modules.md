# 패키지 모듈 상세

> `/moai codemaps`로 생성된 패키지 목록입니다. 존재 여부는 작업 트리만을 근거로 판정했고,
> 이전 codemaps 문서를 존재의 근거로 쓰지 않았습니다.

**모듈**: `github.com/modu-ai/moai-adk` · **Go**: 1.26.4
**측정 트리**: worktree `.claude/worktrees/t475`, 브랜치 `WT-codemaps-stale`, HEAD `52f863f36`
**측정**: 2026-09-08

파일 수는 전부 `find <dir> -name '*.go' -not -name '*_test.go' | wc -l`로 센 **비테스트 파일**이며
하위 패키지를 포함합니다.

---

## presentation

| 패키지 | 비테스트 | 책임 | 주요 하위 패키지 |
|---|---|---|---|
| `cmd/moai` | 1 | 바이너리 유일 진입점. `cli.Execute()` 호출 후 `cli.ResolveExitCode`로 종료 코드만 매핑 | — |
| `internal/cli` | 279 | 아래 클러스터 표 참조 | `update`(+`plan`/`deploy`/`merge`/`backup`/`report`), `harness`, `worktree`, `agentlint`, `preference`, `wizard`, `uikit`, `printer`, `specid`, `taskledger`, `pr` |
| `internal/hook` | 135 | Claude Code 26종 훅 이벤트의 핸들러 레지스트리와 개별 핸들러. `registry.Dispatch`가 이벤트별 체인을 돌려 `HookOutput`을 만든다 | `quality`, `security`, `mx`(+`complexity`), `memo`(+`taxonomy`), `handoff`, `perf`, `trace`, `testutil` |
| `internal/web` | 31 | 루프백 전용 브라우저 콘솔. `a-h/templ` 컴파일 뷰(`*_templ.go`) + htmx + SSE(fsnotify)로 프로파일·설정·todo 큐를 편집하고, **codex 탭 하나는 편집이 아니라 읽기 전용 미러**다(§ codex 미러 탭) | `assets` |
| `internal/statusline` | 21 | Claude Code statusLine 렌더러. git·github·model·backlog·goal·usage 세그먼트 조립. 렌더가 읽고 쓰는 상태의 **앵커는 세션의 현재 디렉터리가 아니라** `internal/stateanchor` seam이 정한 프로젝트 루트이며, 그 어댑터가 `internal/statusline/state_anchor.go`다 | — |
| `internal/tui` | 19 | 터미널 UI 디자인 시스템 — 박스·필·테이블·테마(Catppuccin)·i18n 메시지(`//go:embed messages/*.yaml`) | `golden`, `internal` |
| `internal/mcp` | 1 | self-hosted MCP 도구 카탈로그(도구명 + write 여부) 단일 선언 | — |

### `internal/cli` 기능 클러스터

루트 216개 비테스트 파일을 파일명 접두어로 묶은 것입니다.

| 클러스터 | 파일 | 담당 |
|---|---|---|
| `update*` | 23 | 템플릿 재배포 — 계획/분류/네임스페이스 보호, 3-way 머지, 백업·롤백, 클린 인스톨, dry-run. 단계 로직은 `cli/update/{plan,deploy,merge,backup,report}` 하위로 분해돼 있다. **재배포가 일어나지 않는 경로에도 복구 하나가 붙는다** — `internal/cli/update_mirror_heal.go`는 버전 일치 update가 Deploy 앞에서 조기 반환하는 자리 옆에서 `.agents/skills` 미러를 복구하며, 존재 게이트는 프로젝트의 기록된 배포 버전이다 |
| `doctor*` | 15 | 진단 — config, disk, harness, hook wiring, mcp version, permission, sandbox, skills, worktree base, agentemit embed, codex |
| `mcp*` | 14 | 두 갈래. `mcp_server.go`(50KB)는 stdio JSON-RPC 서버, `mcp.go`/`mcp_codex.go`(89KB — CLI 최대 파일)/`mcp_glm.go`/`mcp_convergence.go`는 codex·GLM 위임과 다중 모델 감사 수렴 |
| `todo*` | 11 | 백로그 큐 CLI. 파일 헤더가 스스로를 `kanban.BacklogStore`에 대한 얇은 cobra 배선이라고 밝힌다 |
| `codex*` | 10 | 외부 에이전트 백엔드 런처, 잡 제어, 준비 상태 점검, 리뷰 게이트. **여기에 사용자 HOME 계층에 대한 스킬 노출 제어 두 개가 함께 산다** — `internal/cli/codex_skills_disable.go`는 `~/.codex/config.toml`에 `enabled = false`를 실은 `[[skills.config]]` 항목을 발행하고, `internal/cli/codex_skills_prune.go`는 가리키는 파일이 사라진 유령 등록을 제거한다(부재를 증명할 수 있는 것만 지우는 allowlist 형 판정, 기본 dry-run) |
| `migrate*` | 9 | 프로파일·에이전시·스킬 복원 등 일회성 마이그레이션 verb |
| `spec*` | 8 | SPEC 문서 lifecycle CLI (view/close/audit/drift) |
| `hook*` | 7 | 훅 디스패처 진입점(`hook.go`, 61KB)과 pre-commit/pre-push 설치 |
| `harness*` | 7 | harness route/validate/ledger/mute/delegation/clusters |
| `init*` | 7 | 프로젝트 초기화(`init.go`) — 템플릿 배포 + settings 생성 + MCP 프로비저닝 |
| `glm*` | 5 | GLM 백엔드 런처·잡 제어 |
| `navigator*` | 5 | BAS 파이프라인 CLI 단계 (enrich/sync/tiers/route/fix) |
| `graph*` / `gate*` / `web*` / `mx*` | 4 각 | 신선도·인용 게이트, 품질 게이트, 콘솔 기동, MX 태그 스캔 |
| `session*` | 3 | 세션 레지스트리 조회·메시징 CLI |
| `integration*` | 2 | 병합 창(acquire/status/release)과 **그 선행 조건인 설정 드리프트 단정**. `integration_settings_drift.go`가 `acquire`의 precondition 이자 독립 verb `moai integration preflight`이며, 창을 잡지 않고도 같은 질문을 물을 수 있게 두 표면을 함께 둔다 |
| `kanban*` / `goal*` | 2 각 | 보드 CLI, goal 조건 arm/status/clear |
| `skills*` | 1 | `moai skills` 명령 트리(`internal/cli/skills.go`). 스킬 노출을 **계층별** 관심사로 두고 계층을 verb 가 아니라 플래그로 명명하며, `--codex`를 필수로 만들어 사용자 HOME 쓰기를 호출 시점 opt-in으로 고정한다 |
| 나머지 | 약 84 | `launcher.go`(53KB, cc/cg/glm 런처), `deps.go`(합성 루트), `root.go`, `profile*` 등과 플랫폼 분기(`*_windows.go` / `*_unix.go`) |

---

## business / domain

| 패키지 | 비테스트 | 책임 | 주요 하위 패키지 |
|---|---|---|---|
| `internal/harness` | 82 | GAN 루프 harness — Socratic 인터뷰 버퍼, 계층적 수락 스코어링, 패턴 학습·티어 분류, FROZEN 가드, lineage 매니페스트, 회귀 게이트 | `curator`, `cluster`, `proposalgen`, `router`, `routing`, `safety`, `seeds`, `throttle`, `tier`, `capture`, `delegationmap`, `v4manifest`, `harnessrun` |
| `internal/navigator` | 53 | BAS(Blueprint-Anchored Synchronization) 파이프라인. 루트에 Go 파일이 없고 전부 단계별 하위 패키지 | `astx`(tree-sitter 16개 언어), `detect`, `sync`, `tiers`, `route`, `fix` |
| `internal/kanban` | 35 | 백로그 큐의 상태 레코드·컬럼·역할 모델, SQLite 저장 엔진, 보드 락, PR 링크, 정합성 조정. **여기에 워킹 트리 검사 하나가 더 있다** — `settings_drift.go`가 병합 전 tracked `.claude/settings.json`의 워킹 사본 드리프트를 단정하고 사본을 보존하며 원장에 남긴다(`--no-optional-locks` 강제 — 평범한 status가 인덱스 쓰기 락을 잡아 병합 직전 경합을 스스로 만들기 때문) | — |
| `internal/spec` | 31 | SPEC 문서 파싱/린트/감사, era 분류, per-SPEC 파일 락, atomic close 오케스트레이터 | — |
| `internal/template` | 30 | `//go:embed all:templates` + `catalog.yaml`. 배포기, 렌더러, settings 생성, 스킬 미러, 카탈로그 트리 해시, 모델 정책·프로파일 매트릭스. **배포 뒤편에 두 개의 기계 방출기와 두 개의 미러 보호·복구 seam이 붙어 있다**(§ 템플릿 방출·미러 계열) | `agentemit`, `commandemit`, `scripts` |
| `internal/core` | 24 | 응집 없는 우산 패키지 (§ `overview.md` 참조) | `git`, `project`, `quality` |
| `internal/mx` | 16 | `@MX:` 코드 주석 태그 스캐너·리졸버 (16개 언어) | — |
| `internal/graph` | 15 | 코드베이스 엣지 리스트를 git-diffable JSONL로 영속화하고 fan-in·최단경로·인용 검증·아키텍처 리포트를 생성 | `symbol` |
| `internal/constitution` | 14 | 규칙 트리의 FROZEN/EVOLVABLE 존 모델과 개정 절차 | — |
| `internal/migration` | 8 | 버전 간 마이그레이션 스텝 레지스트리 | `migrations` |
| `internal/feedback` | 7 | 피드백 리포트 스크러빙(민감정보 제거)과 재시도 큐 | — |
| `internal/goal` | 6 | goal 엔진 — 세션별 조건 선언형 완료 조건 | — |
| `internal/loop` | 6 | Ralph 피드백 루프 상태 기계 | — |
| `internal/verify` | 6 | 공유 진단 스냅샷 계약 | — |
| `internal/merge` | 6 | 템플릿 3-way 머지 엔진 | — |
| `internal/permission` | 6 | 8-tier 권한 스택 (`mvdan.cc/sh`로 셸 명령 파싱) | — |
| `internal/evolution` | 5 | Reflective Learning write phase | — |
| `internal/epic` | 5 | 디스크 기반 epic 진행률 산출 | — |
| `internal/codexwiring` | 5 | Codex 측 배선 파일 생성·갱신 | — |
| `internal/guardliveness` | 4 | 가드 발화 생존성 표면 | — |
| `internal/workflow` | 4 | worktree 전반 워크플로 오케스트레이션 | — |
| `internal/foundation` | 4 | TRUST 등 방법론 타입 정의 | `trust`(빈 디렉터리) |
| `internal/profile` | 3 | 사용자 프로파일·선호 동기화 | — |
| `internal/ciwatch` | 3 | CI watch 루프 분류기 | — |
| `internal/ralph` | 1 | Ralph 결정 엔진 (`engine.go` 단일 파일) | — |

### 템플릿 방출·미러 계열 — 앵커 이후 자란 하위 계층

`internal/template`의 책임 칸 한 줄로는 담기지 않는 네 단위가 하위에 있습니다. 넷 다
"배포기·렌더러"와 다른 축의 일을 합니다.

| 단위 | 비테스트 | 책임 |
|---|---|---|
| `internal/template/agentemit` | 6 | 보존된 에이전트 정의(`.md`)와 임베드 매니페스트(`agents-codex.yaml`)의 쌍을 **중립 원본**으로 삼아 `.codex/agents/` TOML을 결정적으로 이중 발행한다. `.md`의 발행은 항등(identity)이라 재렌더·재정렬이 없고, Codex 쪽은 (`.md` × 매니페스트)의 결정적 변환이다. **fail-closed** — 알 수 없는 tool 토큰·미매핑 effort·유효하지 않은 sandbox 값이면 어느 파일의 어느 토큰인지 지목하며 실패하고 부분 산출물을 남기지 않는다(codex-cli가 알 수 없는 설정을 조용히 무시하므로 생성기 쪽이 자기 출력을 검증해야 한다) |
| `internal/template/commandemit` | 3 | `/moai` 명령 소스를 codex 스킬 아티팩트(`.agents/skills/moai-<command>/SKILL.md`)로 발행한다. 명령 소스는 읽기 전용으로 소비하며 **본문은 바이트 동일 verbatim** — 본문에 남은 Claude 전용 도구 참조는 여기서 고치지 않고 경계 플래그로만 기록한다(그 수리는 명령 본문 계층 소관). fail-closed: 프론트매터 구분자 누락, 설명 누락, 무조건 분기 없는 로케일 조건부 설명, 기존 정본 스킬 디렉터리와 충돌하는 파생 이름 |
| `internal/template/published_skills.go` | (파일) | 위 발행 스킬 경로에 대한 **배포측 보호**. 발행 스킬은 보통의 템플릿 파일처럼 배포되지만 경로 네임스페이스가 스킬 미러가 쓰는 `.agents/skills` 루트와 겹치고, update 모드(forceUpdate)는 다른 곳에서 provenance 검사를 건너뛴다. 이 검사가 그 경로들에 한해 init 모드의 provenance 동작을 살려 사용자 소유 파일이 update를 살아남게 하고, 건너뜀을 침묵이 아니라 보고로 남긴다 |
| `internal/template/skill_mirror_repair.go` | (파일) | `.agents/skills`의 두 생산자(심볼릭 링크 미러, 발행 SKILL.md) 결과를 **Deploy 없이** 복구하는 패키지 수준 패스. DeployerOption이 아닌 형상을 의도적으로 골랐다 — 옵션이었다면 배포 경로에서도 살아나 수리 기능의 부작용으로 배포 동작이 바뀐다. 항목별 의미는 미러 생산자의 것을 재사용하므로 생산자와 갈라질 수 없다 |

두 방출기는 **비테스트 코드에서 아무도 import 하지 않습니다**(`internal/template/agentemit`,
`internal/template/commandemit` 둘 다 패키지 단위 fan-in 0). 소비자는 빌드 타깃
(`make agents-emit` / `make commands-emit`)과 골든 테스트이며, 이는 고아가 아니라
**빌드타임 도구**라는 뜻입니다 — 아래 §네거티브 스페이스의 "호출자 0"과 구별해야 합니다.

### codex 미러 탭 — `internal/web`의 편집하지 않는 표면

`internal/web/codexmirror.go`는 codex 탭의 **행 모델**이며, Audit·MCP 탭에 사는 codex 설정의
**읽기 전용 미러**입니다. 어떤 필드도 옮기지 않습니다 — 미러된 필드는 각자의 소유 탭에서
그대로 선언·렌더·편집되고, 이 파일은 미러가 무엇을 보여주고 어디를 가리키는지만 정합니다.
행은 `settings.AllFields()`와 공유 MCP 도구 카탈로그에 대한 **술어로 파생**되며 손으로
열거하지 않습니다(손 열거는 codex 필드가 하나 늘어나는 순간 조용히 어긋납니다).

렌더 쪽 패널은 `a-h/templ`이 짝 `.templ` 소스에서 생성한 산물입니다.
이 패널은 `name` 속성을 가진 폼 요소를 하나도 내지 않으며, 그 금지는 숨은 bool 동반자
`<name>__present`에도 그대로 걸립니다 — 모든 패널이 한 폼 안에 살고 탭은 표시 전환일 뿐이라
비활성 패널도 함께 제출되기 때문입니다.

---

## data / persistence

| 패키지 | 비테스트 | 책임 | 주요 하위 패키지 |
|---|---|---|---|
| `internal/config` | 50 | 프로젝트 설정의 SSOT. 섹션별 YAML 로딩·캐시·검증(`go-playground/validator`), `envkeys.go`의 환경변수 상수 카탈로그, 기본값. **트리 최대 fan-in (21)** 이며 `types.go`(75KB)·`defaults.go`(52KB)가 트리에서 가장 큰 손 저작 파일 축에 든다 | `atomicfile`, `toolpolicy` |
| `internal/session` | 22 | 세션 레지스트리·체크포인트·페이즈·앵커·태스크 원장. PID 조회를 OS별 파일로 분기 | — |
| `internal/settings` | 10 | `moai web` 콘솔과 `moai profile setup` TUI 두 표면이 공유하는 설정 스키마 | `agentfm`, `yamlpatch` |
| `internal/sessionmsg` | 7 | 단일 머신 세션 간 메시징 브로커 (envelope 스키마) | — |
| `internal/chain` | 4 | **워크트리 세션 origin-trail 체인** — `.moai/state/chain/events.jsonl`에 spawn 경계·`session_id` 백필·완료 엣지를 append-only JSONL 계보 트리로 적는다. 쓰기는 매번 `O_APPEND`로 열어 커널이 동시 append를 직렬화하게 두며, 읽고-고치고-쓰는 주기가 없다(전체 파일을 올려 변형하지 않는다). 깨진 줄은 스트림을 중단시키지 않고 건너뛴다. 목적은 depth-N 워크트리에 `/clear` 이후 재진입한 사람이 grep·스크롤백 고고학 없이 origin·완료·재개 지점을 바로 복원하는 것이다 | — |
| `internal/guardstate` | 4 | 가드 생존성의 상태 모델·매니페스트 | — |
| `internal/manifest` | 3 | 파일 provenance 추적과 변경 감지 | — |
| `internal/tokenusage` | 3 | Claude Code 트랜스크립트 JSONL을 파싱해 토큰 사용량을 귀속·기록. **호출자 0 — 아래 §네거티브 스페이스** | — |

### `internal/settings/yamlpatch` — 보존 쓰기 경로

`internal/settings`의 책임은 "두 표면이 공유하는 **스키마**"지만, `yamlpatch`가 지는 것은
스키마가 아니라 **쓰기 방식**입니다. `ConfigManager.Save()`의 typed struct 재직렬화는
YAML 주석 전량과 미모델링 키(예: `workflow.yaml`의 `team.patterns`, role-profile의 `effort`)를
파괴합니다. `yamlpatch`는 `gopkg.in/yaml.v3` 노드 트리를 수술해 대상 스칼라만 upsert 하고
나머지 문서 구조를 보존하며, **`Save()` 경로가 없는 8개 섹션**(workflow, harness, ralph,
research, feedback, observability, security, db)의 **유일한 쓰기 경로**입니다.

한계도 요건의 일부입니다 — yaml.v3 Encoder는 재직렬화 시 일부 포매팅(빈 줄, 긴 스칼라
줄바꿈)을 정규화할 수 있어 byte-stability는 보증이 아니라 **검증 대상**이고, 섹션별 골든
round-trip 테스트가 그 범위를 고정합니다. 노드 삭제는 지원하지 않습니다.

---

## infrastructure / platform

| 패키지 | 비테스트 | 책임 | 주요 하위 패키지 |
|---|---|---|---|
| `internal/lsp` | 35 | LSP 클라이언트 스택. `charmbracelet/x/powernap` 전송 위에 gopls 브릿지, 서브프로세스 수명 관리, TTL 진단 캐시, 다중 서버 집계 | `core`, `gopls`, `transport`, `subprocess`, `cache`, `config`, `aggregator`, `hook` |
| `internal/astgrep` | 13 | ast-grep(`sg`) CLI 래핑 기반 AST 분석·룰 시드 | — |
| `internal/github` | 10 | `gh` CLI 기반 PR/이슈 오퍼레이션 | `workflow` |
| `internal/runtime` | 10 | 토큰 서킷 브레이커, 감사 캐시/게이트/리포트, 클록 | `gobin` |
| `internal/sandbox` | 8 | 명령 실행을 감싸는 임시 샌드박스 실행 계층 | — |
| `internal/git` | 8 | `core/git` 위에 얹은 상위 유틸리티. **루트 패키지 자체의 비테스트 import 는 0이고, 최상위 집계 fan-in 1은 전부 하위 `convention`이 받은 것**이다 — 아래 §네거티브 스페이스 | `convention` |
| `internal/update` | 7 | 바이너리 셀프 업데이트 (체커 + 오케스트레이터) | — |
| `internal/shell` | 6 | 셸 탐지와 환경 구성 | — |
| `internal/telemetry` | 5 | 텔레메트리 수집·전송 | — |
| `internal/codexadapter` | 5 | Codex 훅 표면 ↔ MoAI 훅 스키마 번역 | — |
| `internal/tmux` | 4 | 병렬 SPEC용 tmux 세션 관리 | — |
| `internal/worktree` | 4 | 워킹 트리 상태 가드 프리미티브 | — |
| `internal/resilience` | 3 | 외부 서비스 연동용 서킷 브레이커 | — |
| `internal/glmcred` | 1 | GLM API 자격증명 단일 구현 | — |
| `internal/binlag` | 1 | 설치된 바이너리 지연 판정 | — |
| `internal/mirrornotice` | 1 | 스킬 미러 결과를 사용자 알림으로 전환 | — |
| `internal/report` | 1 | 루트에 Go 파일 없음 — 하위 `planhtml`만 존재 | `planhtml` |

---

## cross-cutting

| 패키지 | 비테스트 | fan-in | 책임 |
|---|---|---|---|
| `internal/defs` | 5 | 11 | 디렉터리명·파일명 등 프로젝트 전역 상수 |
| `internal/atomicfile` | 5 | 9 | 크로스 플랫폼 원자적 파일 교체 (unix/windows 분기) |
| `pkg/models` | 4 | 8 | 공유 데이터 모델. 외부 공개 2개 패키지 중 하나 |
| `pkg/version` | 2 | 4 | 빌드타임 버전 정보 (ldflags 주입) |
| `internal/lockfile` | 2 | 1 | 크로스 플랫폼 advisory 파일 락 |
| `internal/paths` | 1 | 10 | `~/.moai` 디렉터리 해석의 단일 지점 |
| `internal/execerr` | 1 | 7 | 서브프로세스 종료 실패를 안전하게 출력 가능한 형태로 유지 |
| `internal/stateanchor` | 1 | 2 | **상태 앵커 seam** — 아래 상세 |
| `internal/measure` | 1 | 2 | 의존성 없는 순수 leaf — 프로젝트 헬스 지표 |
| `internal/timing` | 1 | 0 | 테스트용 보정된 지연 상한 (비테스트 fan-in 0) |
| `internal/skills` | 0 | 0 | 프로덕션 코드 없음 |

### `internal/stateanchor` — 상태를 어디에 쓸지 정하는 단일 seam

`.moai/state/`를 읽고 쓰는 **모든** 표면이 앵커(프로젝트 루트)를 여기서 받습니다. 이 seam
이전에는 각 표면이 세션이 서 있던 자리에서 제 앵커를 유도했고, statusline의 텔레메트리
쓰기가 `workspace.current_dir`에 앵커돼 **cd 한 디렉터리마다 `.moai` 디렉터리가 하나씩
남았습니다**(GH #1694). 지금 이 seam을 통과하는 소비자는 statusline 텔레메트리 쓰기,
보드 루트와 그 landed·github-counts 소비자, goal 상태 읽기, CLI 설정 캐시 사슬입니다.

우선순위 사슬은 요건으로 **고정**돼 있습니다(단계 삽입·재정렬은 요건 변경입니다):

1. stdin `workspace.project_dir`
2. `worktree.original_cwd`
3. 세션 디렉터리에 대한 리포지터리 해석 — `core/git`의 공통 디렉터리 해석을 재사용하며,
   그 부모는 리포지터리의 모든 체크아웃과 워크트리에 대해 하나다

셋 다 실패하면 앵커는 빈 문자열이고, 호출자는 상태 쓰기·읽기를 건너뛴 채 렌더를 정상
완료합니다(프로젝트가 없으면 상태도 없다). **세션의 현재 디렉터리는 앵커가 아니며**
리포지터리 walk-up의 입력으로만 쓰입니다. 표시 이름 유도는 별개 관심사로
statusline의 `extractProjectDirectory`에 남아 있습니다. statusline 쪽 어댑터가
`internal/statusline/state_anchor.go`입니다.

---

## 네거티브 스페이스

결정적 도구가 만들 수 없는 관찰입니다. 이 절이 이 문서에서 가장 값이 나가는 부분입니다.

### 테스트가 없는 패키지 — 실질 3개

`go list -f '{{.ImportPath}} {{len .TestGoFiles}} {{len .XTestGoFiles}}' ./...` 기준입니다.

| 패키지 | 판단 |
|---|---|
| `cmd/moai` | 정당. 20줄 위임 로직이고 `internal/cli`에 통합 테스트가 있다 |
| `internal/template/scripts` | 정당. 빌드타임 생성 도구 main |
| `scripts/convert-nextra-to-hextra` | 일회성 문서 변환 스크립트 |

테스트 비율이 1.63:1이므로 **테스트 부족은 이 코드베이스의 약점이 아닙니다.**

### 프로덕션 코드 없이 테스트만 있는 패키지 — 2개

- **`internal/skills`** — `workflow_split_test.go` 하나뿐, 비테스트 파일 0개.
- **`internal/tui/golden`** — `doc.go`와 `index_test.go`뿐.

둘 다 "테스트가 다른 곳(템플릿 트리, 골든 파일)을 검증하는데 담을 자리가 없어 만들어진 빈
패키지"로 보입니다. 필요한 것은 패키지가 아니라 테스트 파일을 둘 자리입니다.

### 비테스트 코드에서 아무도 import 하지 않는 패키지

`.Imports`(테스트 import 제외) 기준입니다. **셋으로 갈립니다** — 빌드타임 도구, 의도된
고아, 그리고 새로 생긴 것.

| 패키지 | 상태 |
|---|---|
| `internal/template/agentemit` · `internal/template/commandemit` | **고아가 아니다.** 소비자가 `make agents-emit` / `make commands-emit` 빌드 타깃과 골든 테스트다. 방출기는 빌드타임 도구이므로 런타임 fan-in 0이 정상 상태다 |
| **`internal/git`** (루트 패키지) | **이 판에서 새로 잡혔다.** `core/git` 위의 상위 유틸리티 8 파일인데 루트 패키지를 import 하는 비테스트 코드가 0이다. 실제로 import 되는 것은 하위 `internal/git/convention` 하나뿐이며(`internal/cli` → `internal/git/convention`), 최상위 집계 fan-in 1은 그것이다. 소비자가 `core/git`로 직접 내려가면서 중간 계층만 남은 모양으로 읽힌다 — 확인이 필요한 관찰이며, 이 문서가 답을 주지는 않는다 |
| **`internal/tokenusage`** | 완전 고아. 유일한 언급이 `internal/spec/audit.go`의 주석인데, 상수를 공유할 수 있지만 파서를 self-contained로 두려고 로컬 상수를 쓴다는 내용이다 — **공유 의도가 있었으나 거부된 뒤 아무도 쓰지 않게 된** 패키지다. `moai tokens` 서브커맨드조차 이것을 쓰지 않는다 |
| `internal/github/workflow` | GitHub Actions 워크플로 검증기. import 하는 코드가 없다 |
| `internal/harness/harnessrun` · `seeds` · `throttle` | harness 하위인데 형제 패키지 어느 것도 참조하지 않는다 |
| `internal/migration/migrations` | `internal/cli/migration_m3_test.go`가 명시한다 — `internal/cli`가 이 패키지를 import 하지 않으므로 m001/m002는 `Register()`를 호출하지 않는다. blank import로 등록되는 패턴인데 그 blank import가 어디에도 없다. 테스트가 이 사실을 *기술*할 뿐 *거부*하지 않는 것이 문제다 |
| `internal/cli/taskledger` · `internal/lsp/aggregator` · `internal/hook/testutil` · `internal/timing` · `internal/tui/golden` | 테스트 전용 소비자만 갖는 leaf. 앞의 셋은 의도로 보이고, `timing`은 이름이 그것을 말한다 |

### 빈 디렉터리 — 3개

`.gitkeep`만 있고 Go 코드가 없습니다.

- `internal/core/integration/`
- `internal/core/migration/` — 실제 마이그레이션은 `internal/migration`에 있습니다
- `internal/foundation/trust/` — TRUST 5 구현은 `internal/core/quality/trust.go`에 있습니다

뒤의 둘이 특히 위험합니다. **실제 구현이 다른 곳에 있는데 자리표시 디렉터리가 그 이름을
선점**하고 있어서, 다음 사람이 여기에 코드를 넣으면 두 벌이 생깁니다.

### 중복된 능력

- **`internal/atomicfile`(18개 파일에서 사용) vs `internal/config/atomicfile`(10개 파일에서 사용).**
  둘 다 살아 있고 둘 다 원자적 파일 쓰기를 합니다. 전자는 `Replace` + unix/windows 분기 + 읽기
  헬퍼, 후자는 `Write` + guard입니다. **한 프로세스 안에서 서로 다른 fsync/rename 전략의 두
  구현이 같은 디렉터리를 건드릴 수 있습니다.**
- `internal/merge`(3-way 머지 엔진) vs `internal/cli/update/merge`(업데이트 오케스트레이션 머지) —
  층위가 다르지만 grep 시 매번 함께 걸립니다. `internal/report/planhtml` vs
  `internal/cli/update/report`도 같은 모양입니다.
- **설정 드리프트 판정이 두 곳에 있습니다** — `internal/kanban/settings_drift.go`(도메인 절반)와
  `internal/cli/integration_settings_drift.go`(CLI 절반). 이것은 중복이 아니라 의도된 분할이며,
  두 표면(`acquire` precondition / `preflight` 독립 verb) 중 어느 쪽도 뺄 수 없다는 것이 그
  파일들의 주석이 적어 둔 설계입니다.

### 경계가 잘못 그어진 것

- **`internal/cli/preference`와 `internal/cli/specid`는 CLI 표면이 아닙니다.** `preference`는
  AskUserQuestion 결정 메모리 레이어, `specid`는 SPEC-ID sanitizer leaf입니다. 둘 다
  `internal/cli` 밖에서 import 당하고, 그것이 `dependencies.md`의 집계 상호 참조 두 쌍을
  만듭니다. 최상위로 승격하면 두 쌍이 사라집니다.
- **`internal/hook/session_start.go`가 61KB**입니다. 옆에
  `session_start_compact.go` · `_factory.go` · `_kanban.go` · `_guard_liveness.go` ·
  `_binary_lag.go` 등이 이미 따로 있는데도 그렇습니다. 세션 시작은 이미 자기 패키지가 되기에
  충분한 크기입니다.
- **`internal/cli/mcp_codex.go`가 89KB**로 CLI 최대 파일입니다. `internal/codexadapter`와
  `internal/codexwiring`이 이미 있는데도 로직 대부분이 CLI 파일에 남아 있습니다.
- **`internal/goal` / `loop` / `ralph`** — 6 / 6 / 1 파일이며 셋이 한 루프 서브시스템입니다.
  SPEC이 셋이었다는 것 외에 경계가 셋인 근거가 보이지 않습니다.
  **`guardliveness`(4) / `guardstate`(4)** 도 8개 파일을 둘로 나눌 분량이 아닙니다.
  **`internal/stateanchor`(1)** 도 같은 계열이지만 이쪽은 정당화가 있습니다 — 여러 표면이
  공유해야 하는 단일 결정 규칙이라 어느 소비자 밑에도 둘 수 없습니다.

### 트리에서 가장 큰 비테스트 파일은 손으로 쓴 것이 아닙니다

```
$ find internal cmd pkg -name '*.go' -not -name '*_test.go' -exec ls -l {} + | sort -k5 -rn | head -4
168KB internal/web/fieldsets_templ.go      (생성)
121KB internal/web/screens_templ.go        (생성)
 89KB internal/cli/mcp_codex.go            (손 저작 — CLI 최대)
 75KB internal/config/types.go             (손 저작)
```

직전 판은 `internal/hook/session_start.go`(당시 67KB)를 "트리 최대 비테스트 Go 파일"이라고
적었습니다. **지금은 사실이 아닙니다** — 상위 둘이 `a-h/templ` 생성 산물이고, `session_start.go`는
61KB로 5위권입니다. 크기 순위를 읽을 때는 생성 파일과 손 저작 파일을 갈라 세어야 합니다.

### 폐기 표식이 코드로 남은 것

- `internal/cli/root.go` — `newHarnessCmd()`가 은퇴 마커로서 컴파일 가능 상태로만 남아 있고,
  트리에 등록되지 않습니다. `TestHarnessFactoryStillCompiles`가 이를 고정합니다.
  죽은 코드를 테스트가 살려두는 구조입니다.
- `internal/hook/retired_events.go` — 은퇴한 훅 이벤트 목록. 소비자가 테스트 스위트와
  (등록되지 않는) `migration/migrations` m002뿐입니다.
