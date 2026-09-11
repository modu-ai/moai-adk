# t652 — moai worktree 스킬·문서가 없는 명령과 플래그를 안내하는 표류: 측정과 정정

- card: t652 (문서 카드) · worktree `.claude/worktrees/t652` · branch `WT-worktree-doc-drift`
- base: 로컬 develop = origin/develop `ac6c42c2d`
- 상태: **정정 완료(R1~R5, 리드 승인).** 코드 변경 없음. 범위가 카드의 89줄을 넘어(141줄·8파일) 먼저 수리 방향을 보고했고, 리드가 R1~R5 를 승인했다(§5).

## 1. 정본 — 소스 리터럴 (`ac6c42c2d`)

설치 바이너리(rc.7)는 develop 보다 오래돼 소스를 기준으로 삼았다.

- 하위 명령(`grep -n "Use:" internal/cli/worktree/*.go`, 테스트 제외): `sync [branch-name]`, `remove [path]`, `clean`, `recover`, `done [branch-name]`, `snapshot`, `verify`, `restore` — 8개. 별칭 `wt`.
- 플래그 이름 리터럴(`Flags().String|Bool|Int|…` 등록 호출에서 추출): `agent-name agent-response auto base delete-branch dry-run force json main merge merged-only out snapshot stale strategy yes` — 16개.
- 진입은 명령이 아니라 런처 몫: `moai cc -w <name>`, `moai cc -w <name> --spawn` (`root.go` Long).
- `moai-worktree` 라는 실행 파일·별칭은 없다. `cmd/`·`internal/`·`pkg/` 의 적중 3건은 모두 `~/.moai/worktrees` 경로를 슬러그로 바꾼 문자열(`-Users-…--moai-worktrees-foo`, 테스트 디렉터리명)이다.

## 2. 측정 (템플릿 사본 `internal/template/templates/.claude/skills/moai-workflow-worktree/`)

### 2.1 없는 실행 파일 이름으로 된 명령 호출 — 접두형

`moai-worktree <단어>` 가 든 줄: **141줄 / 8파일**

| 파일 | 줄 |
|---|---|
| `modules/worktree-commands.md` | 66 |
| `references/examples.md` | 35 |
| `modules/troubleshooting.md` | 21 |
| `references/reference.md` | 13 |
| `modules/integration-patterns.md` | 2 |
| `modules/tools-integration.md` | 2 |
| `modules/parallel-development.md` | 1 |
| `modules/moai-adk-integration.md` | 1 |

토큰 분포: `sync` 20 · `new` 19 · `status` 17 · `clean` 15 · `remove` 13 · `go` 13 · `config` 12 · `switch` 10 · `list` 9 · (산문 명사: `registry` 6 · `with` 3 · `skill` 2 · `commands` 2 · `integration` 1 · `configuration` 1) · `optimize` 1 · `join` 1.

- **없는 명령**(`new|status|go|config|switch|list|optimize|join`)이 든 줄: **81**
- **있는 명령을 없는 실행 파일 이름으로** 부르는 줄(`moai-worktree sync|clean|remove|done|…`): **47**
- 올바른 실행 파일 이름(`moai worktree`)으로 없는 명령을 부르는 줄: 1 (`moai worktree new`)

카드의 "89줄"은 다른 계수 패턴(`worktree X` 형태)으로 잰 값이다. 이번 계수는 실행 파일 이름까지 틀린 형태를 잡는 `moai-worktree <단어>` 패턴이며, 두 수는 같은 대상을 다른 경계로 잰 것이다.

### 2.2 산문형 명령 언급

`the (worktree )?<cmd> command` 형태: 5건 — `the worktree sync command` · `the worktree clean command` · `the sync command` · `the clean command` · `The sync command`. 모두 **있는** 명령을 가리키므로 산문형 자체는 표류가 아니다. 다만 이 문장들이 설명하는 기능(선택 동기화 등)이 없는 경우가 있다(§2.3).

### 2.3 있는 명령의 없는 플래그

| 플래그 | 스킬 문서 적중 | 소스 리터럴 | 위치 |
|---|---|---|---|
| `--include` | 5 | 0 | modules·references |
| `--exclude` | 4 | 0 | 〃 |
| `--auto-resolve` | 2 | 0 | 〃 |
| `--interactive` | 6 | 0 | 〃 |
| `--template` | 4 | 0 | 〃 |
| `--developer` | 0 | 0 | 플래그 형태로는 없음. SKILL.md:216-220·268 이 "developer prefixes / team mode registry" 를 산문으로 설명 |
| **양성 대조** `--merged-only` | 6 | 14 | 소스·문서 모두 있음 |
| **양성 대조** `--strategy` | 0 | 3 | 소스에 있음(문서 미기재 — 표류 아님) |

`--*` 형태 적중 파일: `references/examples.md`, `references/reference.md`, `modules/moai-adk-integration.md`, `modules/worktree-commands.md`. SKILL.md 에는 `--` 형태가 없고, 대신 산문으로 같은 기능을 설명한다: `:226` (sync 의 include/exclude 패턴, auto-resolve, interactive), `:228-232` (template 플래그로 만드는 워크트리 템플릿).

### 2.4 템플릿 `AGENTS.md:288`

```
| `moai worktree` | Worktree lifecycle (list / snapshot / verify / restore) |
```

`list` 는 없는 명령이고 `sync / remove / clean / recover / done` 이 빠졌다. 로컬 루트 `AGENTS.md` 에는 이 줄이 없다(`grep -n "list / snapshot"` 적중은 템플릿 사본뿐).

### 2.5 `internal/cli/worktree/root.go` Long — 코드, 이 카드 범위 밖

`Long: "… Supports creating, syncing, removing, and cleaning worktrees."` — 생성 명령은 없다(바로 아래 줄이 "Entering a worktree is the launchers' job" 이라 자기모순). 리드 지시(문서만, 코드 변경 없음)에 따라 고치지 않고 후속 후보로 기록한다.

### 2.6 로컬 사본은 템플릿보다 오래됐다

`diff -r .claude/skills/moai-workflow-worktree internal/template/templates/.claude/skills/moai-workflow-worktree` → 11개 파일에서 다르다. 차이 63줄 중 20줄은 로컬에만 있는 `Last Updated: <날짜>` 줄(템플릿은 날짜를 싣지 않는 규칙)이고, 나머지 43줄은 로컬이 옛 판이다 — 예: 로컬 `/moai:1-plan`, `/moai:2-run`, `/moai:3-sync` ↔ 템플릿 `/moai plan`, `/moai run`, `/moai sync`; frontmatter `updated: "2026-01-08"` ↔ `"2026-07-10"`. 로컬 `.claude/skills/moai*` 는 `moai update` 가 템플릿으로 통째 덮는 관리 대상이다(CLAUDE.local.md §2.3).

## 3. 수리 방향 (결정 요청)

| # | 대상 | 제안 |
|---|---|---|
| R1 | `modules/worktree-commands.md` (66줄이 사라진 CLI 설명서) | **재작성**: 8개 하위 명령 + 런처 진입 + 소스 리터럴 플래그만 담은 짧은 참조로 교체. `new/list/switch/go/status/config` 절은 삭제하고, 대체 경로(생성·진입 = `moai cc -w <name>`, 목록·상태 = `git worktree list`)를 한 절로 안내 |
| R2 | `references/examples.md` 35 · `modules/troubleshooting.md` 21 · `references/reference.md` 13 · 나머지 4파일 6줄 | **표적 치환**: 있는 명령은 `moai-worktree X` → `moai worktree X`; 없는 명령 줄은 대체 경로로 바꾸거나, 해당 예시·절이 없는 명령에만 기대면 삭제. 참인 문장은 유지 |
| R3 | 없는 플래그(`--include/--exclude/--auto-resolve/--interactive/--template`)와 SKILL.md `:226`·`:228-232`·`:216-220`·`:268` 산문 | 해당 플래그 문장·예시 삭제. SKILL.md 의 없는 기능 절(선택 동기화 패턴, 워크트리 템플릿, 팀 레지스트리 모드)은 삭제하거나 "지원하지 않음 — git 로 직접" 한 줄로 축소 |
| R4 | 템플릿 `AGENTS.md:288` | `(sync / remove / clean / recover / done / snapshot / verify / restore)` 로 교체. 꼬리 절단 예산 문서라 바이트 변화 기록 |
| R5 | 로컬 스킬 사본 | 템플릿 수정 후 **로컬을 템플릿과 같게** 맞춘다(날짜 줄 포함 옛 판 43줄도 함께 사라짐). `moai update` 가 결국 같은 결과를 내므로 배포 결과와 일치시키는 방향. 대안: 로컬은 이번 표류 줄만 고치고 옛 판 차이는 유지 |
| — | `root.go` Long "Supports creating" | 코드라 이 카드에서 제외 → 후속 카드 후보 |

규모 추정: R1 은 모듈 1개 재작성, R2 는 약 75줄 치환·삭제, R3 은 SKILL.md 4곳과 3개 파일의 플래그 줄, R4 는 1줄. 결과적으로 템플릿 8~9개 파일과 로컬 사본 전체가 바뀐다.

## 5. 정정 (리드 승인: R1~R5, R5 = 로컬을 템플릿과 같게)

정정 중 소스를 더 읽어 확인한 인자 규칙(새 거짓 문장을 만들지 않으려고):
- `sync [branch-name]` — 인자가 있으면 그 브랜치의 워크트리, 없으면 **현재 디렉터리**의 워크트리(`sync.go` `cobra.MaximumNArgs(1)` + Long).
- `remove [path]`, `done [branch-name]` — 도움말 표기와 달리 인자 **필수**(`cobra.ExactArgs(1)`). `done` 은 디렉터리 이름이 아니라 **브랜치 이름**을 받는다(`done.go` 가 `wt.Branch == branchName` 으로 찾음).
- `recover` — `git worktree repair` 후 prune 하고 결과를 나열(`recover.go`).

### 5.1 파일별 변경과 삭제 근거 (삭제·교체한 절마다 "소스에 없음" 근거 한 줄)

| 파일 | 변경 | 없는 기능 근거 |
|---|---|---|
| `modules/worktree-commands.md` | **재작성**(R1): 8개 하위 명령·런처 진입·소스 리터럴 플래그만 담은 참조 + "Not Provided by This Command" 대체 표 1개 | `new`/`list`/`switch`/`go`/`status`/`config` 절 삭제 — `Use:` 는 8개뿐(§1). 절에 적힌 플래그(`--branch --template --shallow --depth --format --status --sort --reverse --verbose --auto-sync --new-terminal --absolute --relative --export --auto-resolve --interactive --include --exclude --keep-branch --backup --days`) — 소스 플래그 리터럴 16개 목록에 없음 |
| `references/examples.md` | **재작성**(R2): 예시 10개 → 6개(런처·`moai worktree`·git 만 사용) | 삭제: Ex4·5 Python `moai_worktree` 모듈 — 저장소 전체에 `moai_worktree` 0건 / Ex7 팀 설정(`config set`, `registry_type team`, `join`) — `config`·`join` 하위 명령 없음, `team`·`shared registry` 리터럴 0건(`internal/cli/worktree`) / Ex8 워크트리 템플릿 — `template` 플래그 리터럴 0건 / Ex9 자동화(`new`·`list`·`status`·`optimize`) — 하위 명령 없음 |
| `modules/troubleshooting.md` | **표적 치환**(R2) 13곳 | `registry validate/prune/export/rebuild` — `registry` 하위 명령 없음, 대체 `moai worktree recover` + `git worktree prune`. `status` — 없음, 대체 `git worktree list` / `moai worktree clean --json`. `remove --keep-branch` — 플래그 없음(`remove` 는 원래 브랜치를 남김) |
| `references/reference.md` | 절 3개 삭제 + 제목·꼬리말 이름 정정(R2) | Click·Rich·Python 자료("used by moai-worktree") — CLI 는 Go/cobra(`go.mod: github.com/spf13/cobra`). 최적화 기법(`--shallow --depth --include --exclude --all --background`) — 플래그 리터럴 0건. CI/CD 예시(GitHub Actions·Jenkins: `new`·`go`·`config`) — 하위 명령 없음 |
| `modules/integration-patterns.md` | 제목 + 핵심 흐름 문단 | `moai-worktree go` — 없음. "/moai plan 이 워크트리를 자동 생성" → 런처(`moai cc -w`)가 생성·진입 |
| `modules/parallel-development.md` | 핵심 흐름 문단 | `new`·`go`·`sync --all` — 하위 명령·플래그 없음 |
| `modules/tools-integration.md` | 제목 + 1줄 | 프로그램 이름만 정정 |
| `modules/moai-adk-integration.md` | 제목 + 충돌 해결 2곳 | `auto-resolve`/`interactive`/`abort`, `--include/--exclude`, sync `--force` — sync 가 등록하는 플래그는 `base`·`strategy` 뿐(`sync.go:28-29`) |
| `SKILL.md` | 절 4개 → 2개로 교체(R3) + Quick Decision Guide 2문장 | 팀 공유 레지스트리·developer 접두(`team`·`shared registry` 0건) / 선택 동기화·auto-resolve·interactive(리터럴 0건) / 워크트리 템플릿(`template` 0건) / 성능 최적화 `shallow·background·parallel·cache`(리터럴 0건). 교체 절: "Synchronization Strategies"(실제 `--strategy`·`--base`) + "Capabilities Not Provided" |
| 템플릿 `AGENTS.md:288` | `(list / snapshot / verify / restore)` → `(sync / remove / clean / recover / done / snapshot / verify / restore)`(R4) | `list` 하위 명령 없음, 5개 누락 |
| 로컬 `.claude/skills/moai-workflow-worktree/` | 템플릿과 같게(R5) | 옛 판 43줄·로컬 전용 날짜 줄 20줄도 함께 사라짐 |
| `internal/template/catalog.yaml` | `moai-workflow-worktree` 해시 1줄 재생성 | `ea5542b8…` → `2be92941…` |

### 5.2 AGENTS.md 바이트 (항상 로드되는 계약, 꼬리 절단 예산)

`git cat-file -s HEAD:internal/template/templates/AGENTS.md` = **16936** → 수정 후 `wc -c` = **16970** (+34B). 로컬 루트 `AGENTS.md` 에는 해당 줄이 없어 변경 없음. 같은 문구의 다른 사본: `grep -rln "list / snapshot / verify / restore"` → 템플릿 `AGENTS.md` 와 과거 보고서(`.moai/reports/t628/`, 이 verdict)뿐.

## 6. 검증

### 6.1 재측정 — 템플릿 스킬 폴더 (정정 후)

```
moai-worktree 적중: 3
  modules/worktree-commands.md:103   "Earlier documentation described a separate `moai-worktree` program…" (대체 표 도입문 — 의도)
  modules/worktree-management.md:143, modules/registry-architecture.md:12   `.moai-worktree-registry` 파일 이름 (§Residual 참조)
`moai worktree (new|list|switch|go|status|config|optimize|join|registry)` 적중: 1
  modules/moai-adk-integration.md:28   "The retired `/moai plan --worktree` flag and the retired `moai worktree new`…" (폐지 사실을 적은 참인 문장)
없는 플래그(카드·측정 목록) 적중: 모두 modules/worktree-commands.md:110-113 대체 표 안
`moai worktree …` 줄에 붙은 플래그: --base --delete-branch --json --merged-only --stale --strategy --yes (모두 소스 16개 안)
  + --spawn (`moai cc -w <name> --spawn` 이 같은 줄에 있는 경우), --worktree (폐지 문장)
```

### 6.2 사본·카탈로그·테스트

```
diff -rq .claude/skills/moai-workflow-worktree internal/template/templates/.claude/skills/moai-workflow-worktree   → 출력 없음 (local == template)
go run ./internal/template/scripts/gen-catalog-hashes.go --entry moai-workflow-worktree --dry-run → 2be92941… (catalog.yaml not modified)
go run ./internal/template/scripts/gen-catalog-hashes.go --entry moai-workflow-worktree         → catalog.yaml updated successfully (12899 bytes)
git diff internal/template/catalog.yaml → hash 한 줄만 변경
go test ./internal/template/... -count=1 → exit 0   (template-tests.txt)
  ok  internal/template 29.033s · ok agentemit · ok commandemit
```

## Gaps

- 설치 바이너리의 `--help` 는 읽지 않았다(리드 지시: 소스 리터럴 기준). 설치본(rc.7)이 develop 보다 오래돼 도움말이 다를 수 있다.
- 새로 쓴 예시 명령은 실행해 보지 않았다. 인자·플래그 규칙은 소스(`Use:`, `cobra.*Args`, `Flags()` 등록)로만 대조했다.
- 정정 뒤 재측정은 템플릿 사본에서 했고, 로컬은 `diff -rq` 동일로 대신했다.

## Residual-risk

- **범위 밖으로 남긴 표류(후속 후보)**: `modules/registry-architecture.md`·`worktree-management.md` 가 설명하는 `.moai-worktree-registry` JSON 레지스트리 — 소스에 그 파일 이름 0건(`grep -rn "worktree-registry" internal cmd pkg --include='*.go'`). `references/reference.md` 의 가짜 커뮤니티 URL(`@username`), 가상의 로드맵(v1.1~1.3), `pip install`·`pytest` 개발 환경. `modules/parallel-advanced.md` "Shell Integration" 절 등 셸 함수 설명. 이 카드는 명령·플래그 표류만 다뤘다.
- `internal/cli/worktree/root.go` Long 의 "Supports creating…" — 코드라 제외(후속 후보).
- `remove`·`done` 의 `Use:` 가 대괄호(`[path]`, `[branch-name]`)로 인자를 선택처럼 보이게 하지만 실제로는 필수다 — 코드 쪽 표기 불일치, 이 카드 범위 밖.
