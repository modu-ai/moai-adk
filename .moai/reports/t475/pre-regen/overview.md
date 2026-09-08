# 아키텍처 개요

> `/moai codemaps`로 생성된 아키텍처 지도입니다. 모든 수치는 아래 트리에서 직접 잰 것이고,
> 다른 트리·다른 시점에서 옮겨온 값은 없습니다.

**모듈**: `github.com/modu-ai/moai-adk` · **Go**: 1.26.4
**측정 트리**: worktree `.claude/worktrees/t476`, 브랜치 `WT-codemaps-progress`, HEAD `25a3212a9`
**측정**: 2026-09-04

---

## 규모

| 값 | 수치 | 산출 명령 |
|---|---|---|
| 비테스트 Go 파일 | 1096 | `find internal cmd pkg -name '*.go' -not -name '*_test.go' \| wc -l` |
| 테스트 Go 파일 | 1771 | `find internal cmd pkg -name '*_test.go' \| wc -l` |
| Go 패키지 총수 | 137 | `go list ./... \| wc -l` (`scripts/` 하위 main 3개 포함) |
| 최상위 패키지 | 68 | `internal` 65 + `cmd/moai` 1 + `pkg` 2 |
| 내부 import 엣지 (패키지 단위) | 1638 | `go list -f '{{range .Imports}}...'` 후 모듈 경로 필터 |
| 내부 import 엣지 (최상위 집계) | 205 | 위를 `internal/<X>` 수준으로 접고 self-edge 제거 |
| 임베드 템플릿 파일 | 564 | `find internal/template/templates -type f \| wc -l` |

테스트 대 비테스트 비율이 **1.6 : 1**입니다. 테스트 파일이 0인 패키지는 3개뿐이고
셋 다 정당한 사유가 있습니다(§ `modules.md` 참조).

---

## 이 코드베이스가 실제로 따르는 구조

**깔끔하게 들어맞는 이름은 없습니다.** 가장 가까운 것은 Go 표준 프로젝트 레이아웃에 얇은
헥사고날 시도를 얹은 형태지만, 실제로 지배적인 것은 **명령별 수직 슬라이스를 가진 모듈러
모놀리스**입니다.

정직하게 한 줄로 적으면 — **SPEC 단위로 증식한 수평 패키지 위에 `internal/cli`라는 단일
거대 어댑터가 얹힌 구조**입니다.

### 헥사고날에 부합하는 근거

- `cmd/` · `internal/` · `pkg/` 3분할과 `internal/`의 65개 도메인 분해는 표준 레이아웃 그대로입니다.
- 합성 루트가 명시적으로 하나 있습니다 — `internal/cli/deps.go`의 `Dependencies` 구조체와
  `InitDependencies()`. `git.Repository`, `hook.Registry`, `hook.Protocol`, `update.Checker`,
  `update.Orchestrator` 같은 인터페이스 타입으로 조립하므로 포트/어댑터 의도가 보입니다.
- `internal/hook/registry.go`의 `Register` / `Dispatch`는 교과서적인 핸들러 레지스트리 + 체인입니다.
- 안정 의존성 원칙을 만족합니다 — fan-in 상위를 `defs` · `paths` · `execerr` · `atomicfile` ·
  `models` 같은 cross-cutting leaf가 차지하고, 그것들의 fan-out은 0에 가깝습니다.

### 부합하지 않는 근거 — 이쪽이 더 결정적입니다

- **`internal/cli`가 최상위 패키지 68개 중 58개를 import 합니다.** 헥사고날이라면 어댑터 하나가
  전 도메인에 닿을 이유가 없습니다. 실제 모양은 "명령 하나 = 파일 하나 = 그 명령이 필요한 것
  전부 import"에 가깝습니다.
- **도메인 로직이 어댑터 안에 삽니다.** `internal/hook/session_start.go`가 67KB로 트리 전체
  비테스트 Go 파일 중 최대이고, `pre_tool.go`가 48KB, `branch_guard.go`가 27KB입니다. 이들은
  프로토콜 변환이 아니라 정책입니다.
- **레이어 방향이 국소적으로 뒤집힙니다.** presentation인 `internal/hook`이 다른 최상위 패키지
  6개에게, `internal/statusline`이 5개에게 import 당합니다.
- **DI가 전역 변수 하나로 전달됩니다.** `var deps *Dependencies`는 컨테이너가 아니라 전역
  상태이고, 그래서 `if deps == nil` 형태의 nil 방어가 곳곳에 필요해졌습니다.
- **패키지 경계가 응집도가 아니라 SPEC 단위로 그어졌습니다.** 대부분의 패키지 doc 코멘트가
  `SPEC-XXX-NNN` 형태로 시작합니다. `goal` / `loop` / `ralph`가 셋으로,
  `guardliveness` / `guardstate`가 둘로 쪼개진 것이 그 결과입니다.

---

## 레이어

패키지의 **주된 대화 상대**로 판정했습니다.

| 레이어 | 판정 규칙 | 대표 패키지 |
|---|---|---|
| presentation | 프로세스 경계 바깥의 표면(터미널·HTTP·훅 프로토콜)과 직접 말한다 | `cmd/moai`, `internal/cli`, `internal/hook`, `internal/tui`, `internal/web`, `internal/statusline`, `internal/mcp` |
| business/domain | MoAI 고유 규칙·정책만 담고 자체 I/O 프리미티브를 소유하지 않는다 | `internal/spec`, `internal/harness`, `internal/navigator`, `internal/kanban`, `internal/graph`, `internal/mx` … |
| data/persistence | 디스크상 named artifact 하나의 스키마와 읽기·쓰기 계약을 소유한다 | `internal/config`, `internal/session`, `internal/settings`, `internal/manifest` … |
| infrastructure/platform | 외부 프로세스·OS·네트워크 설비를 감싼다 | `internal/lsp`, `internal/git`, `internal/github`, `internal/astgrep`, `internal/tmux` … |
| cross-cutting | 정책이 없고 무관한 다수 패키지가 쓰는 leaf (fan-in ≥ 5, 도메인 지식 없음) | `internal/defs`, `internal/paths`, `internal/atomicfile`, `pkg/models` … |

전체 배치는 `modules.md`에 있습니다.

### 도식에 들어맞지 않는 패키지

분류가 어긋나는 자리는 반올림이 아니라 **발견**이므로 그대로 적습니다.

- **`internal/hook` (134 파일)** — 가장 큰 불일치입니다. 겉으로는 Claude Code 훅 JSON을
  stdin에서 읽어 stdout으로 내보내는 인바운드 어댑터지만, 안에 브랜치 가드·세션 시작
  오케스트레이션·증거 기록기 같은 순수 정책이 함께 삽니다. presentation으로 부르면 정책이
  감춰지고 domain으로 부르면 stdin/stdout 계약이 감춰집니다. 어느 쪽이든 손실이 있습니다.
- **`internal/kanban` (33 파일)** — 카드 도메인 규칙(`role.go` · `column.go` · `reconcile.go`)과
  SQLite 스토리지 엔진(`backlog_sqlite.go` — WAL · busy_timeout · IMMEDIATE 트랜잭션을 직접
  소유)이 한 패키지에 있습니다. domain과 data가 분리돼 있지 않습니다.
- **`internal/template` (25 파일)** — 도메인(카탈로그·모델 정책), 데이터(564개 파일의
  `//go:embed all:templates` 트리), 인프라(배포기)를 동시에 수행합니다.
- **`internal/core`** — 이름과 달리 응집된 core가 아닙니다. `core/git`은 인프라,
  `core/project` · `core/quality`는 도메인이며, `core/integration`과 `core/migration`은
  `.gitkeep` 하나뿐인 빈 디렉터리입니다.
- **`internal/mcp` (1 파일)** — `catalog.go`의 도구 이름 + write 여부 선언 리스트뿐입니다.
  프로토콜 계약 선언이라 presentation에 두었으나 실질은 `internal/defs`와 같은 상수 leaf입니다.
- **`internal/skills`** — 비테스트 Go 파일이 0개이고 `workflow_split_test.go` 하나만 있습니다.
  어떤 레이어에도 속하지 않습니다.

---

## 관련 문서

- `modules.md` — 패키지별 책임과 파일 수
- `dependencies.md` — fan-in / fan-out 상위와 상호 참조 3쌍
- `entry-points.md` — `main()`, Cobra 트리, 훅, MCP 표면
- `data-flow.md` — 계층을 관통하는 경로 4개
