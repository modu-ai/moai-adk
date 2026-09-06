# 패키지 모듈 상세

> `/moai codemaps`로 생성된 패키지 목록입니다. 존재 여부는 작업 트리만을 근거로 판정했고,
> 이전 codemaps 문서를 존재의 근거로 쓰지 않았습니다.

**모듈**: `github.com/modu-ai/moai-adk` · **Go**: 1.26.4
**측정 트리**: worktree `.claude/worktrees/t476`, 브랜치 `WT-codemaps-progress`, HEAD `25a3212a9`
**측정**: 2026-09-04

파일 수는 전부 `find <dir> -name '*.go' -not -name '*_test.go' | wc -l`로 센 **비테스트 파일**이며
하위 패키지를 포함합니다.

---

## presentation

| 패키지 | 비테스트 | 책임 | 주요 하위 패키지 |
|---|---|---|---|
| `cmd/moai` | 1 | 바이너리 유일 진입점. `cli.Execute()` 호출 후 `cli.ResolveExitCode`로 종료 코드만 매핑 | — |
| `internal/cli` | 273 | 아래 클러스터 표 참조 | `update`(+`plan`/`deploy`/`merge`/`backup`/`report`), `harness`, `worktree`, `agentlint`, `preference`, `wizard`, `uikit`, `printer`, `specid`, `taskledger`, `pr` |
| `internal/hook` | 134 | Claude Code 26종 훅 이벤트의 핸들러 레지스트리와 개별 핸들러. `registry.Dispatch`가 이벤트별 체인을 돌려 `HookOutput`을 만든다 | `quality`, `security`, `mx`(+`complexity`), `memo`(+`taxonomy`), `handoff`, `perf`, `trace`, `testutil` |
| `internal/web` | 29 | 루프백 전용 브라우저 콘솔. `a-h/templ` 컴파일 뷰(`*_templ.go`) + htmx + SSE(fsnotify)로 프로파일·설정·todo 큐를 편집 | `assets` |
| `internal/statusline` | 20 | Claude Code statusLine 렌더러. git·github·model·backlog·goal·usage 세그먼트 조립 | — |
| `internal/tui` | 19 | 터미널 UI 디자인 시스템 — 박스·필·테이블·테마(Catppuccin)·i18n 메시지(`//go:embed messages/*.yaml`) | `golden`, `internal` |
| `internal/mcp` | 1 | self-hosted MCP 도구 카탈로그(도구명 + write 여부) 단일 선언 | — |

### `internal/cli` 기능 클러스터

루트 210개 비테스트 파일을 파일명 접두어로 묶은 것입니다.

| 클러스터 | 파일 | 담당 |
|---|---|---|
| `update*` | 22 | 템플릿 재배포 — 계획/분류/네임스페이스 보호, 3-way 머지, 백업·롤백, 클린 인스톨, dry-run. 단계 로직은 `cli/update/{plan,deploy,merge,backup,report}` 하위로 분해돼 있다 |
| `mcp*` | 14 | 두 갈래. `mcp_server.go`(50KB)는 stdio JSON-RPC 서버, `mcp.go`/`mcp_codex.go`(91KB — CLI 최대 파일)/`mcp_glm.go`/`mcp_convergence.go`는 codex·GLM 위임과 다중 모델 감사 수렴 |
| `doctor*` | 14 | 진단 — config, disk, harness, hook wiring, mcp version, permission, sandbox, skills, worktree base, agentemit embed, codex |
| `todo*` | 11 | 백로그 큐 CLI. 파일 헤더가 스스로를 `kanban.BacklogStore`에 대한 얇은 cobra 배선이라고 밝힌다 |
| `migrate*` | 9 | 프로파일·에이전시·스킬 복원 등 일회성 마이그레이션 verb |
| `codex*` / `glm*` | 8 / 5 | 외부 에이전트 백엔드 런처, 잡 제어, 준비 상태 점검, 리뷰 게이트 |
| `spec*` | 8 | SPEC 문서 lifecycle CLI (view/close/audit/drift) |
| `hook*` | 7 | 훅 디스패처 진입점(`hook.go`, 62KB)과 pre-commit/pre-push 설치 |
| `harness*` | 7 | harness route/validate/ledger/mute/delegation/clusters |
| `init*` | 7 | 프로젝트 초기화(`init.go`, 46KB) — 템플릿 배포 + settings 생성 + MCP 프로비저닝 |
| `navigator*` | 5 | BAS 파이프라인 CLI 단계 (enrich/sync/tiers/route/fix) |
| 나머지 | 약 93 | `launcher.go`(55KB, cc/cg/glm 런처), `deps.go`(합성 루트), `root.go`, `gate*`, `graph*`, `web*`, `mx*`, `kanban*`, `session*`, `profile*`, `goal*` 등과 플랫폼 분기(`*_windows.go` / `*_unix.go`) |

---

## business / domain

| 패키지 | 비테스트 | 책임 | 주요 하위 패키지 |
|---|---|---|---|
| `internal/harness` | 82 | GAN 루프 harness — Socratic 인터뷰 버퍼, 계층적 수락 스코어링, 패턴 학습·티어 분류, FROZEN 가드, lineage 매니페스트, 회귀 게이트 | `curator`, `cluster`, `proposalgen`, `router`, `routing`, `safety`, `seeds`, `throttle`, `tier`, `capture`, `delegationmap`, `v4manifest`, `harnessrun` |
| `internal/navigator` | 53 | BAS(Blueprint-Anchored Synchronization) 파이프라인. 루트에 Go 파일이 없고 전부 단계별 하위 패키지 | `astx`(tree-sitter 16개 언어), `detect`, `sync`, `tiers`, `route`, `fix` |
| `internal/kanban` | 33 | 백로그 큐의 상태 레코드·컬럼·역할 모델, SQLite 저장 엔진, 보드 락, PR 링크, 정합성 조정 | — |
| `internal/spec` | 31 | SPEC 문서 파싱/린트/감사, era 분류, per-SPEC 파일 락, atomic close 오케스트레이터 | — |
| `internal/template` | 25 | `//go:embed all:templates` + `catalog.yaml`. 배포기, 렌더러, settings 생성, 스킬 미러, 카탈로그 트리 해시, 모델 정책·프로파일 매트릭스 | `agentemit`, `scripts` |
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

---

## data / persistence

| 패키지 | 비테스트 | 책임 | 주요 하위 패키지 |
|---|---|---|---|
| `internal/config` | 50 | 프로젝트 설정의 SSOT. 섹션별 YAML 로딩·캐시·검증(`go-playground/validator`), `envkeys.go`의 환경변수 상수 카탈로그, 기본값. **트리 최대 fan-in (21)** | `atomicfile`, `toolpolicy` |
| `internal/session` | 22 | 세션 레지스트리·체크포인트·페이즈·앵커·태스크 원장. PID 조회를 OS별 파일로 분기 | — |
| `internal/settings` | 10 | `moai web` 콘솔과 `moai profile setup` TUI 두 표면이 공유하는 설정 스키마 | `agentfm`, `yamlpatch` |
| `internal/sessionmsg` | 7 | 단일 머신 세션 간 메시징 브로커 (envelope 스키마) | — |
| `internal/guardstate` | 4 | 가드 생존성의 상태 모델·매니페스트 | — |
| `internal/manifest` | 3 | 파일 provenance 추적과 변경 감지 | — |
| `internal/tokenusage` | 3 | Claude Code 트랜스크립트 JSONL을 파싱해 토큰 사용량을 귀속·기록. **호출자 0 — 아래 §네거티브 스페이스** | — |

---

## infrastructure / platform

| 패키지 | 비테스트 | 책임 | 주요 하위 패키지 |
|---|---|---|---|
| `internal/lsp` | 35 | LSP 클라이언트 스택. `charmbracelet/x/powernap` 전송 위에 gopls 브릿지, 서브프로세스 수명 관리, TTL 진단 캐시, 다중 서버 집계 | `core`, `gopls`, `transport`, `subprocess`, `cache`, `config`, `aggregator`, `hook` |
| `internal/astgrep` | 13 | ast-grep(`sg`) CLI 래핑 기반 AST 분석·룰 시드 | — |
| `internal/github` | 10 | `gh` CLI 기반 PR/이슈 오퍼레이션 | `workflow` |
| `internal/runtime` | 10 | 토큰 서킷 브레이커, 감사 캐시/게이트/리포트, 클록 | `gobin` |
| `internal/git` | 8 | `core/git` 위에 얹은 상위 유틸리티 | `convention` |
| `internal/sandbox` | 8 | 명령 실행을 감싸는 임시 샌드박스 실행 계층 | — |
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
| `internal/lockfile` | 2 | 2 | 크로스 플랫폼 advisory 파일 락 |
| `pkg/version` | 2 | 4 | 빌드타임 버전 정보 (ldflags 주입) |
| `internal/paths` | 1 | 10 | `~/.moai` 디렉터리 해석의 단일 지점 |
| `internal/execerr` | 1 | 7 | 서브프로세스 종료 실패를 안전하게 출력 가능한 형태로 유지 |
| `internal/measure` | 1 | 2 | 의존성 없는 순수 leaf — 프로젝트 헬스 지표 |
| `internal/timing` | 1 | 2 | 테스트용 보정된 지연 상한 |
| `internal/skills` | 0 | 0 | 프로덕션 코드 없음 |

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

테스트 비율이 1.6:1이므로 **테스트 부족은 이 코드베이스의 약점이 아닙니다.**

### 프로덕션 코드 없이 테스트만 있는 패키지 — 2개

- **`internal/skills`** — `workflow_split_test.go` 하나뿐, 비테스트 파일 0개.
- **`internal/tui/golden`** — `doc.go`와 `index_test.go`뿐.

둘 다 "테스트가 다른 곳(템플릿 트리, 골든 파일)을 검증하는데 담을 자리가 없어 만들어진 빈
패키지"로 보입니다. 필요한 것은 패키지가 아니라 테스트 파일을 둘 자리입니다.

### 호출자가 0인 프로덕션 패키지 — 4개

| 패키지 | 상태 |
|---|---|
| **`internal/tokenusage`** | 완전 고아. 유일한 언급이 `internal/spec/audit.go`의 주석인데, 상수를 공유할 수 있지만 파서를 self-contained로 두려고 로컬 상수를 쓴다는 내용이다 — **공유 의도가 있었으나 거부된 뒤 아무도 쓰지 않게 된** 패키지다. `moai tokens` 서브커맨드조차 이것을 쓰지 않는다 |
| `internal/github/workflow` | GitHub Actions 워크플로 검증기. import 하는 코드가 없다 |
| `internal/harness/harnessrun` | harness 제안 실행 타입. 형제 패키지 어느 것도 참조하지 않는다 |
| `internal/migration/migrations` | `internal/cli/migration_m3_test.go`가 명시한다 — `internal/cli`가 이 패키지를 import 하지 않으므로 m001/m002는 `Register()`를 호출하지 않는다. blank import로 등록되는 패턴인데 그 blank import가 어디에도 없다. 테스트가 이 사실을 *기술*할 뿐 *거부*하지 않는 것이 문제다 |

### 빈 디렉터리 — 3개

`.gitkeep`만 있고 Go 코드가 없습니다.

- `internal/core/integration/`
- `internal/core/migration/` — 실제 마이그레이션은 `internal/migration`에 있습니다
- `internal/foundation/trust/` — TRUST 5 구현은 `internal/core/quality/trust.go`에 있습니다

뒤의 둘이 특히 위험합니다. **실제 구현이 다른 곳에 있는데 자리표시 디렉터리가 그 이름을
선점**하고 있어서, 다음 사람이 여기에 코드를 넣으면 두 벌이 생깁니다.

### 중복된 능력

- **`internal/atomicfile`(17개 파일에서 사용) vs `internal/config/atomicfile`(10개 파일에서 사용).**
  둘 다 살아 있고 둘 다 원자적 파일 쓰기를 합니다. 전자는 `Replace` + unix/windows 분기 + 읽기
  헬퍼, 후자는 `Write` + guard입니다. `internal/cli`의 6개 파일이 후자를, 나머지 트리가 전자를
  씁니다. **한 프로세스 안에서 서로 다른 fsync/rename 전략의 두 구현이 같은 디렉터리를 건드릴
  수 있습니다.**
- `internal/merge`(3-way 머지 엔진) vs `internal/cli/update/merge`(업데이트 오케스트레이션 머지) —
  층위가 다르지만 grep 시 매번 함께 걸립니다. `internal/report/planhtml` vs
  `internal/cli/update/report`도 같은 모양입니다.

### 경계가 잘못 그어진 것

- **`internal/cli/preference`와 `internal/cli/specid`는 CLI 표면이 아닙니다.** `preference`는
  AskUserQuestion 결정 메모리 레이어(8개 파일), `specid`는 SPEC-ID sanitizer leaf입니다. 둘 다
  `internal/cli` 밖에서 import 당하고, 그것이 `dependencies.md`의 집계 상호 참조 두 쌍을
  만듭니다. 최상위로 승격하면 두 쌍이 사라집니다.
- **`internal/hook/session_start.go`가 67KB**로 트리 최대 비테스트 파일입니다. 옆에
  `session_start_compact.go` · `_factory.go` · `_kanban.go` · `_guard_liveness.go` ·
  `_binary_lag.go` 등이 이미 따로 있는데도 그렇습니다. 세션 시작은 이미 자기 패키지가 되기에
  충분한 크기입니다.
- **`internal/cli/mcp_codex.go`가 91KB**로 CLI 최대 파일입니다. `internal/codexadapter`와
  `internal/codexwiring`이 이미 있는데도 로직 대부분이 CLI 파일에 남아 있습니다.
- **`internal/goal` / `loop` / `ralph`** — 6 / 6 / 1 파일이며 셋이 한 루프 서브시스템입니다.
  SPEC이 셋이었다는 것 외에 경계가 셋인 근거가 보이지 않습니다.
  **`guardliveness`(4) / `guardstate`(4)** 도 8개 파일을 둘로 나눌 분량이 아닙니다.

### 폐기 표식이 코드로 남은 것

- `internal/cli/root.go` — `newHarnessCmd()`가 은퇴 마커로서 컴파일 가능 상태로만 남아 있고,
  트리에 등록되지 않습니다. `TestHarnessFactoryStillCompiles`가 이를 고정합니다.
  죽은 코드를 테스트가 살려두는 구조입니다.
- `internal/hook/retired_events.go` — 은퇴한 훅 이벤트 목록. 소비자가 테스트 스위트와
  (등록되지 않는) `migration/migrations` m002뿐입니다.
