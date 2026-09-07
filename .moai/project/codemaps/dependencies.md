# 의존성 그래프

> `/moai codemaps`로 생성됐습니다. 내부 엣지만 대상이며 stdlib·서드파티는 제거했습니다.

**측정 트리**: worktree `.claude/worktrees/t475`, 브랜치 `WT-codemaps-stale`, HEAD `52f863f36`
**측정**: 2026-09-08

두 가지 해상도로 봅니다 — 패키지 단위 **345 엣지**, 이를 `internal/<X>` 최상위로 접고
self-edge를 제거한 **208 엣지**. 아래 표는 후자 기준입니다.

산출:

```
$ go list -f '{{.ImportPath}} {{join .Imports " "}}' ./... \
  | awk '{src=$1; for(i=2;i<=NF;i++) if ($i ~ /^github\.com\/modu-ai\/moai-adk\//) print src, $i}' \
  | wc -l
345
```

> 직전 판(앵커 `25a3212a9`)은 이 자리에 1638을 적었습니다. 위 명령으로 재현되지 않고 직전 판의
> 명령 인용이 생략형이라 무엇을 셌는지 복원할 수 없으므로, 이 판은 위 명령의 출력을 싣습니다.
> 최상위 집계는 205 → 208로 움직였습니다(이쪽은 정합).

---

## fan-in 상위 — 다른 최상위 패키지에게 import 당한 수

| # | 패키지 | 피import | 레이어 |
|---|---|---|---|
| 1 | `internal/config` | 21 | data |
| 2 | `internal/defs` | 11 | cross-cutting |
| 3 | `internal/paths` | 10 | cross-cutting |
| 4 | `internal/atomicfile` | 9 | cross-cutting |
| 5 | `pkg/models` | 8 | cross-cutting |
| 6 | `internal/execerr` | 7 | cross-cutting |
| 6 | `internal/core` | 7 | domain |
| 8 | `internal/template` | 6 | domain |
| 8 | `internal/hook` | 6 | **presentation** |
| 10 | `internal/statusline` | 5 | **presentation** |
| 10 | `internal/spec` | 5 | domain |
| 10 | `internal/lsp` | 5 | infrastructure |

상위 6개 중 5개가 cross-cutting leaf라는 것은 **건강한 신호**입니다 — 안정 의존성 원칙 그대로입니다.

다만 8위 `internal/hook`(6)과 10위 `internal/statusline`(5)은 **presentation인데 피의존
대상**입니다. 방향이 뒤집혀 있고, 이것이 `overview.md`가 "레이어링이 국소적으로 무너진다"고
적은 근거입니다.

`internal/core`가 6 → 7로 오른 것은 새 seam 하나가 그 밑으로 내려갔기 때문입니다 —
`internal/stateanchor`가 리포지터리 해석 단계에서 `internal/core/git`을 재사용합니다.

### 새로 그래프에 들어온 leaf

| 패키지 | fan-in | 비고 |
|---|---|---|
| `internal/stateanchor` | 2 | 상태 앵커 seam. 소비자는 `internal/statusline`과 `internal/cli` |
| `internal/chain` | 2 | 워크트리 세션 origin-trail 원장 |

두 방출기(`internal/template/agentemit`, `internal/template/commandemit`)는 이 표에 **나타나지
않습니다** — 비테스트 fan-in이 0이기 때문입니다. 고아가 아니라 빌드타임 도구이며, 소비자가
`make agents-emit` / `make commands-emit` 타깃과 골든 테스트입니다(`modules.md` §네거티브 스페이스).

---

## fan-out 상위 — 다른 최상위 패키지를 import 한 수

| # | 패키지 | import |
|---|---|---|
| 1 | `internal/cli` | **59** |
| 2 | `internal/hook` | 30 |
| 3 | `internal/web` | 14 |
| 4 | `internal/core` | 13 |
| 5 | `internal/statusline` | 8 |
| 6 | `internal/settings` | 7 |
| 7 | `internal/feedback` | 6 |
| 8 | `internal/kanban` | 5 |
| 9 | `internal/harness` | 4 |
| 10 | `internal/update` · `template` · `spec` | 3 각 |

`internal/cli`가 최상위 68개 중 **59개**를 import 합니다 — 사실상 전 트리에 닿습니다.
합성 루트(`internal/cli/deps.go`)가 여기 있으므로 일부는 의도된 것이지만, 59 중 상당수는
`deps.go`가 아니라 **개별 verb 파일에서 직접** 들어옵니다. 이것이 "명령 하나 = 파일 하나 =
그 명령이 필요한 것 전부 import"라는 수직 슬라이스 성격을 만듭니다.

`internal/statusline`이 7 → 8로 오른 것도 같은 seam 때문입니다 — 렌더의 상태 앵커가
`internal/stateanchor`로 옮겨가면서 엣지가 하나 늘었습니다.

---

## 순환

**패키지 단위 순환은 존재하지 않습니다.** Go 컴파일러가 금지하므로 구조적으로 불가능하고,
`go list ./...`가 오류 없이 완주하는 것으로 확인됩니다.

**최상위 집계 단위에서는 상호 참조가 3쌍** 있습니다. 엣지 목록과 그 역방향을 교차시켜 얻었습니다.

| 상호 쌍 | 실제 엣지 | 원인 |
|---|---|---|
| `internal/cli` ↔ `internal/hook` | `cli → hook`, `cli → hook/{handoff,memo/taxonomy,perf,quality,security}` / `hook → cli/preference` | `cli/preference`가 CLI 표면이 아닌 공유 leaf인데 `internal/cli` 밑에 있다 |
| `internal/cli` ↔ `internal/kanban` | `cli → kanban` / `kanban → cli/specid` | `cli/specid`(SPEC-ID sanitizer leaf)가 `internal/cli` 밑에 있다 |
| `internal/hook` ↔ `internal/migration` | `hook → migration` / `migration/migrations → hook` | 마이그레이션 스텝이 훅의 은퇴 이벤트 목록을 읽는다 |

**세 쌍 모두 패키지 배치 문제이지 실제 순환이 아닙니다.** 앞의 두 쌍은 `cli/preference`와
`cli/specid`를 최상위로 승격하면 즉시 사라집니다.

---

## 외부 의존성

`go.mod`의 direct require 29개 항목입니다.

| 모듈 | 용도 | 사용처 |
|---|---|---|
| `github.com/spf13/cobra` v1.10.2 | CLI 명령 트리 | `internal/cli` 전역 |
| `github.com/spf13/pflag` v1.0.10 | cobra 플래그 | 동상 |
| `charm.land/fang/v2` v2.0.1 | cobra 위 help/error/version/completion 렌더러 | `internal/cli/fang.go` |
| `charm.land/bubbletea/v2` v2.0.9 | TUI 이벤트 루프 | `internal/cli/wizard` |
| `charm.land/bubbles/v2` v2.1.1 | TUI 컴포넌트 | 동상 |
| `charm.land/huh/v2` v2.0.3 | 폼/프롬프트 | `internal/cli/wizard`, `huh_theme.go` |
| `charm.land/lipgloss/v2` v2.0.6 | 스타일링 | `internal/tui` |
| `github.com/charmbracelet/huh` v1.0.0 | **v2와 병존하는 v1 폼** | `internal/cli` |
| `github.com/charmbracelet/lipgloss` v1.1.1-… | **v2와 병존하는 v1 스타일링** | `internal/statusline`, `internal/cli` |
| `github.com/charmbracelet/glamour` v1.0.0 | 마크다운 터미널 렌더 | `internal/cli/spec_view.go` |
| `github.com/charmbracelet/colorprofile` v0.4.3 | 컬러 프로파일 감지 | tui |
| `github.com/charmbracelet/x/powernap` v0.1.6 | LSP JSON-RPC 전송 | `internal/lsp/transport`, `lsp/core` |
| `github.com/muesli/termenv` v0.16.0 | 터미널 능력 감지 | tui / statusline |
| `github.com/mattn/go-isatty` v0.0.24 | TTY 판별 | 출력 분기 |
| `github.com/mattn/go-runewidth` v0.0.28 | 동아시아 문자폭 계산 | 테이블 / statusline 정렬 |
| `github.com/mark3labs/mcp-go` v0.58.0 | MCP 서버 SDK (stdio 전송) | `internal/cli/mcp_server.go` |
| `github.com/a-h/templ` v0.3.1020 | 타입 세이프 HTML 템플릿 컴파일러 | `internal/web/*.templ` |
| `golang.org/x/net` v0.58.0 | HTML 파싱 | **비테스트 사용처 0 — 테스트 전용** |
| `github.com/smacker/go-tree-sitter` | 16개 언어 AST 심볼 추출 | `internal/navigator/astx`, `internal/hook/mx/complexity` |
| `mvdan.cc/sh/v3` v3.13.1 | 셸 명령 파싱 | `internal/permission/stack.go` — 유일 사용처 |
| `github.com/go-playground/validator/v10` v10.30.3 | 구조체 태그 기반 설정 검증 | `internal/config/validation.go` — 유일 사용처 |
| `github.com/fsnotify/fsnotify` v1.10.1 | 파일 변경 감시 | `internal/web/events.go`, `internal/hook/config_change.go` |
| `golang.org/x/tools` v0.49.0 | Go 패키지/AST 로딩 | `internal/lsp/config` |
| `gopkg.in/yaml.v3` v3.0.1 | 설정·카탈로그·프론트매터 파싱 + **노드 트리 수술**(`internal/settings/yamlpatch`) | 트리 전역 |
| `golang.org/x/sync` v0.22.0 | errgroup 등 동시성 유틸 | 병렬 스캔 경로 |
| `golang.org/x/sys` v0.47.0 | syscall 래퍼 (파일 락, PID 조회) | `*_unix.go` / `*_windows.go` |
| `golang.org/x/text` v0.41.0 | 유니코드 / 인코딩 | 정규화 경로 |
| `github.com/stretchr/testify` v1.12.1 | 테스트 단언 | 테스트 전용 |
| `go.uber.org/goleak` v1.3.0 | 고루틴 누수 검출 | `internal/hook` 등 |

### 이례적인 것

1. **charm 계열 v1 / v2 동시 사용.** `huh`와 `lipgloss`가 각각 두 메이저 버전을 동시에 require
   하고, `internal/cli/huh_theme.go` **한 파일 안에서 v1과 v2를 모두 import** 합니다.
   절반쯤 진행된 마이그레이션의 흔적이고, 테마 정의를 두 벌 유지하는 비용을 낳습니다.
2. **`modernc.org/sqlite`가 `// indirect` 블록에 있는데 실제로는 직접 import 됩니다.**
   `internal/kanban/backlog_sqlite.go`가 드라이버(`_ "modernc.org/sqlite"`)와 상수
   (`sqlite3 "modernc.org/sqlite/lib"`)를 직접 씁니다 — `go mod tidy`가 아직 반영되지 않은
   상태이며 직전 판 이후에도 그대로입니다. 순수 Go 구현이라 `CGO_ENABLED=0`에서 동작하는
   것은 CLI 배포에 맞는 선택이지만 `modernc.org/libc` 등이 딸려 와 바이너리가 커집니다.
3. **`golang.org/x/net`이 direct require인데 비테스트 사용처가 없습니다.** 사용처 5곳이 전부
   `*_test.go`입니다. direct 블록에 있을 이유가 없습니다.
4. **CLI 치고 의외인 조합** — tree-sitter(cgo), sqlite, LSP 클라이언트, HTML 템플릿 컴파일러,
   로컬 HTTP 서버가 한 바이너리에 다 들어 있습니다. 이 도구는 CLI라기보다 개발 환경
   런타임에 가깝습니다.
5. **`github.com/a-h/templ`이 `tool` 지시어로 등록**돼 있습니다(`go.mod:106`
   `tool github.com/a-h/templ/cmd/templ`). `.templ` → `_templ.go` 생성이 빌드 전제이며
   생성물이 트리에 커밋돼 있습니다 — `internal/web/fieldsets_codex_templ.go`(codex 미러 패널)와
   트리 최대 비테스트 파일 두 개(`fieldsets_templ.go` 168KB, `screens_templ.go` 121KB)가
   그 산물입니다.
6. **`gopkg.in/yaml.v3`의 쓰임이 두 축입니다.** 대부분은 마샬/언마샬이지만
   `internal/settings/yamlpatch`는 같은 라이브러리의 **노드 트리**를 직접 수술해 주석과
   미모델링 키를 보존합니다. 이 두 번째 쓰임이 typed struct 재직렬화가 파괴하는 것을
   보존하는 유일한 경로입니다.
