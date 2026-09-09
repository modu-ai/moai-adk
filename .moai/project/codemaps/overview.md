# 아키텍처 개요

> `/moai codemaps`로 생성된 아키텍처 지도입니다. 모든 수치는 아래 트리에서 직접 잰 것이고,
> 다른 트리·다른 시점에서 옮겨온 값은 없습니다.

**모듈**: `github.com/modu-ai/moai-adk` · **Go**: 1.26.4
**측정 트리**: worktree `.claude/worktrees/t592`, 브랜치 `WT-home-state-rollout`, HEAD `e7bd89ee3`
**측정**: 2026-09-10

---

## 규모

| 값 | 수치 | 산출 명령 |
|---|---|---|
| 비테스트 Go 파일 | 1140 | `find internal cmd pkg -name '*.go' -not -name '*_test.go' \| wc -l` |
| 테스트 Go 파일 | 1865 | `find internal cmd pkg -name '*_test.go' \| wc -l` |
| Go 패키지 총수 | 140 | `go list ./... \| wc -l` (`scripts/` 하위 main 3개 포함) |
| 최상위 패키지 | 69 | `internal` 66 + `cmd/moai` 1 + `pkg` 2 |
| 내부 import 엣지 (패키지 단위) | 351 | `go list -f '{{.ImportPath}} {{join .Imports " "}}' ./...` 후 모듈 경로 필터 |
| 내부 import 엣지 (최상위 집계) | 214 | 위를 `internal/<X>` 수준으로 접고 self-edge 제거 |
| 임베드 템플릿 파일 | 581 | `find internal/template/templates -type f \| wc -l` |

테스트 대 비테스트 비율이 **1.64 : 1**입니다. 테스트 파일이 0인 패키지는 3개뿐이고
셋 다 정당한 사유가 있습니다(§ `modules.md` 참조).

> **엣지 수 정정 이력.** 앵커 `25a3212a9` 판은 패키지 단위 엣지를 1638로 적었지만 당시
> 명령 문자열이 생략형이라 재현할 수 없었습니다. 이후 `52f863f36` 판은 완전한 명령으로
> 345 / 208을 측정했고, 이번 판은 같은 명령으로 351 / 214를 측정했습니다.

---

## 이 코드베이스가 실제로 따르는 구조

**깔끔하게 들어맞는 이름은 없습니다.** 가장 가까운 것은 Go 표준 프로젝트 레이아웃에 얇은
헥사고날 시도를 얹은 형태지만, 실제로 지배적인 것은 **명령별 수직 슬라이스를 가진 모듈러
모놀리스**입니다.

정직하게 한 줄로 적으면 — **SPEC 단위로 증식한 수평 패키지 위에 `internal/cli`라는 단일
거대 어댑터가 얹힌 구조**입니다.

### 헥사고날에 부합하는 근거

- `cmd/` · `internal/` · `pkg/` 3분할과 `internal/`의 66개 도메인 분해는 표준 레이아웃 그대로입니다.
- 합성 루트가 명시적으로 하나 있습니다 — `internal/cli/deps.go`의 `Dependencies` 구조체와
  `InitDependencies()`. `git.Repository`, `hook.Registry`, `hook.Protocol`, `update.Checker`,
  `update.Orchestrator` 같은 인터페이스 타입으로 조립하므로 포트/어댑터 의도가 보입니다.
- `internal/hook/registry.go`의 `Register` / `Dispatch`는 교과서적인 핸들러 레지스트리 + 체인입니다.
- 안정 의존성 원칙을 만족합니다 — fan-in 상위를 `defs` · `paths` · `execerr` · `atomicfile` ·
  `models` 같은 cross-cutting leaf가 차지하고, 그것들의 fan-out은 0에 가깝습니다.

### 부합하지 않는 근거 — 이쪽이 더 결정적입니다

- **`internal/cli`가 최상위 패키지 69개 중 60개를 import 합니다.** 헥사고날이라면 어댑터 하나가
  전 도메인에 닿을 이유가 없습니다. 실제 모양은 "명령 하나 = 파일 하나 = 그 명령이 필요한 것
  전부 import"에 가깝습니다.
- **도메인 로직이 어댑터 안에 삽니다.** `internal/hook/session_start.go`가 61KB,
  `internal/cli/hook.go`가 61KB, `internal/hook/quality/gate.go`가 53KB입니다. 이들은
  프로토콜 변환이 아니라 정책입니다.
- **레이어 방향이 국소적으로 뒤집힙니다.** presentation인 `internal/hook`이 다른 최상위 패키지
  6개에게, `internal/statusline`이 5개에게 import 당합니다.
- **DI가 전역 변수 하나로 전달됩니다.** `var deps *Dependencies`는 컨테이너가 아니라 전역
  상태이고, 그래서 `if deps == nil` 형태의 nil 방어가 곳곳에 필요해졌습니다.
- **패키지 경계가 응집도가 아니라 SPEC 단위로 그어졌습니다.** 대부분의 패키지 doc 코멘트가
  `SPEC-XXX-NNN` 형태로 시작합니다. `goal` / `loop` / `ralph`가 셋으로,
  `guardliveness` / `guardstate`가 둘로 쪼개진 것이 그 결과입니다. 이 판에서 새로 잡힌
  `internal/stateanchor`(1 파일) · `internal/chain`(4 파일)도 같은 증식의 사례입니다.

---

## 레이어

패키지의 **주된 대화 상대**로 판정했습니다.

| 레이어 | 판정 규칙 | 대표 패키지 |
|---|---|---|
| presentation | 프로세스 경계 바깥의 표면(터미널·HTTP·훅 프로토콜)과 직접 말한다 | `cmd/moai`, `internal/cli`, `internal/hook`, `internal/tui`, `internal/web`, `internal/statusline`, `internal/mcp` |
| business/domain | MoAI 고유 규칙·정책만 담고 자체 I/O 프리미티브를 소유하지 않는다 | `internal/spec`, `internal/harness`, `internal/navigator`, `internal/kanban`, `internal/graph`, `internal/mx` … |
| data/persistence | 디스크상 named artifact 하나의 스키마와 읽기·쓰기 계약을 소유한다 | `internal/config`, `internal/session`, `internal/settings`, `internal/manifest`, `internal/chain`, `internal/homestate` … |
| infrastructure/platform | 외부 프로세스·OS·네트워크 설비를 감싼다 | `internal/lsp`, `internal/git`, `internal/github`, `internal/astgrep`, `internal/tmux` … |
| cross-cutting | 정책이 없고 무관한 다수 패키지가 쓰는 leaf (fan-in ≥ 5, 도메인 지식 없음) | `internal/defs`, `internal/paths`, `internal/atomicfile`, `pkg/models`, `internal/stateanchor` … |

전체 배치는 `modules.md`에 있습니다.

### 도식에 들어맞지 않는 패키지

분류가 어긋나는 자리는 반올림이 아니라 **발견**이므로 그대로 적습니다.

- **`internal/hook` (135 파일)** — 가장 큰 불일치입니다. 겉으로는 Claude Code 훅 JSON을
  stdin에서 읽어 stdout으로 내보내는 인바운드 어댑터지만, 안에 브랜치 가드·세션 시작
  오케스트레이션·증거 기록기 같은 순수 정책이 함께 삽니다. presentation으로 부르면 정책이
  감춰지고 domain으로 부르면 stdin/stdout 계약이 감춰집니다. 어느 쪽이든 손실이 있습니다.
- **`internal/kanban` (38 파일)** — 카드 도메인 규칙(`role.go` · `column.go` · `reconcile.go`)과
  SQLite 스토리지 엔진(`backlog_sqlite.go` — WAL · busy_timeout · IMMEDIATE 트랜잭션을 직접
  소유)이 한 패키지에 있습니다. 여기에 워킹 트리 설정 파일을 검사하는
  `settings_drift.go`까지 들어와, 이제 domain · data · 워킹 트리 검사 셋이 한 자리에 있습니다.
- **`internal/template` (30 파일)** — 도메인(카탈로그·모델 정책), 데이터(581개 파일의
  `//go:embed all:templates` 트리), 인프라(배포기)를 동시에 수행하고, 하위에 두 개의
  **기계 방출기**(`agentemit` · `commandemit`)를 품습니다.
- **`internal/core`** — 이름과 달리 응집된 core가 아닙니다. `core/git`은 인프라,
  `core/project` · `core/quality`는 도메인이며, `core/integration`과 `core/migration`은
  `.gitkeep` 하나뿐인 빈 디렉터리입니다.
- **`internal/mcp` (1 파일)** — `catalog.go`의 도구 이름 + write 여부 선언 리스트뿐입니다.
  프로토콜 계약 선언이라 presentation에 두었으나 실질은 `internal/defs`와 같은 상수 leaf입니다.
- **`internal/skills`** — 비테스트 Go 파일이 0개이고 `workflow_split_test.go` 하나만 있습니다.
  어떤 레이어에도 속하지 않습니다.
- **`internal/stateanchor` (1 파일)** — 정책이 없는 leaf처럼 보이지만 담는 것은 **결정 규칙**
  입니다(어느 프로젝트 루트가 상태의 앵커인가). cross-cutting에 두되, 우선순위 사슬 자체가
  요건으로 고정돼 있다는 점에서 순수 leaf와 다릅니다.
- **`internal/homestate` (14 파일)** — `~/.moai` 아래 프로젝트별 SQLite 경로와 스키마를
  소유하는 data/persistence 패키지이면서, Unix `flock`·Windows `LockFileEx`, PID 지문,
  런타임 진입 차단까지 함께 다룹니다. 저장 계약과 플랫폼 동시성 경계가 한 패키지에 만나는
  의도적인 seam이며 소비자는 `cli`, `hook`, `hook/handoff`, `kanban`, `web` 다섯 곳입니다.

## HOME 상태 전환 경계

이번 판은 프로젝트 로컬 상태를 즉시 지우는 전환이 아닙니다. `internal/homestate`가
`~/.moai/db/<project-key>/{todo,factory}/`와 전역 `~/.moai/run/profile-leases.db`의 계약을
제공하고, `moai migrate home-state`가 기본 dry-run으로 이전 가능성을 검사합니다. 실제 적용은
`--apply --verified-live`와 두 차례의 zero-active census, admission lock, 백업·해시·논리 동등성
검증을 모두 통과해야 합니다. 현재 코드와 문서의 존재는 **운영 데이터 이전 완료를 뜻하지
않습니다**. 프로젝트의 `.moai/state/todo/backlog.db`는 검증된 apply 전까지 보존됩니다.

---

## 관련 문서

- `modules.md` — 패키지별 책임과 파일 수
- `dependencies.md` — fan-in / fan-out 상위와 상호 참조 3쌍
- `entry-points.md` — `main()`, Cobra 트리, 훅, MCP 표면
- `data-flow.md` — 계층을 관통하는 경로 4개
