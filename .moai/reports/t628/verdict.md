# t628 — 작업트리 스킬의 권한 보장 설명과 명령 목록 (SK-07 · SK-11)

- 카드: t628 (Tier S, Class B, plan 생략, run → sync)
- 워크트리: `.claude/worktrees/t628`, 브랜치 `WT-worktree-skill-doc`
- 기반: 로컬 `develop` `a0d8da641`을 `--no-ff`로 병합한 `b04a4ec5c` (`HEAD^2` = `a0d8da641d3e2af00d00583583d7ea04f0ec3beb` 확인)
- 출처: `.moai/reports/instruction-audit-20260910-01a089f6/report.md`, 같은 디렉터리의 `findings.json`
- 감사 기준 트리는 `main 2213871af`이고 재현과 수리는 develop에서 했다. 감사가 인용한 줄번호는 main 기준이라 줄번호가 아니라 문구로 찾았다.

---

## 1. develop에서의 재현

### 1.1 이미 갈라져 있던 두 사본 (SK-07)

```
$ git grep -n -E 'mode: ?plan|read-only|...' a0d8da641 -- <스킬 두 사본>
.claude/skills/moai-workflow-worktree/SKILL.md:284: ... Read-only agents (mode: plan) cannot write. ...
.claude/skills/moai-workflow-worktree/SKILL.md:305: ... (verify mode: plan is sufficient)
internal/template/templates/.claude/skills/moai-workflow-worktree/SKILL.md:285: ... Read-only agents cannot write because their tools list omits Write/Edit (the spawn-time mode parameter is deprecated and ignored). ...
internal/template/templates/.claude/skills/moai-workflow-worktree/SKILL.md:306: ... (verify the tools list omits Write/Edit)
```

전문: `develop-sk07-grep.txt`.

- **로컬 사본**은 `mode: plan`을 쓰기 차단의 근거로 둔다. 결함이 살아 있다.
- **템플릿 사본**은 이미 도구 목록을 근거로 바뀌어 있다. 감사 `findings.json`의 SK-07 근거도 같은 사실을 "canonical worktree SKILL already removes this guarantee"라고 적는다.

**템플릿의 수정본도 보장을 과장한다.** "도구 목록에서 Write/Edit를 빼면 쓸 수 없다"는 문장은 Bash를 빠뜨렸다. Bash는 셸로 파일을 쓴다. 이 세션의 런타임 에이전트 목록은 읽기 전용 탐색 에이전트 `Explore`의 도구를 "Agent, Artifact, ArtifactComments, ArtifactData, ArtifactCheck, ExitPlanMode, Edit, Write, NotebookEdit를 제외한 전부"로 표시한다. Write/Edit는 없고 Bash는 있다. Write/Edit를 빼는 것만으로 쓰기가 막히지 않는다는 실례다.

이 관측의 출처는 이 세션의 런타임 목록뿐이고 Claude Code 버전에 따라 달라질 수 있으므로 템플릿 문구에는 `Explore`라는 이름을 넣지 않고 일반 규칙으로만 썼다.

정확한 정의는 이 저장소의 규칙에 이미 있다. `worktree-integration.md:256`은 "a teammate is read-only only when its tools cannot write"라고 쓴다. 이번 수리는 스킬 문구를 그 정의에 맞췄다.

### 1.2 없는 명령 6개와 빠진 명령 5개 (SK-11)

실제 관리 명령은 **8개**다. 두 경로로 재서 일치를 확인했다.

```
$ moai worktree --help          # 설치 바이너리 v3.2.0-rc.5, list-974-g84fa4ece4, 2026-09-09 빌드
COMMANDS
  sync [branch-name] [--flags]  Sync worktree with base branch
  remove [path] [--flags]       Remove a worktree
  clean [--flags]               Clean stale worktree references
  recover                       Repair worktree registry
  done [branch-name] [--flags]  Complete worktree and cleanup
  snapshot [--flags]            Capture working tree state snapshot for guard verification
  verify [--flags]              Verify working tree state against snapshot + check agent response
  restore [--flags]             Restore working tree to a snapshot's HEAD state
help_exit=0

$ git grep -n -E 'Use:[[:space:]]+"' a0d8da641 -- internal/cli/worktree
clean.go:22 "clean" · done.go:18 "done [branch-name]" · guard.go:68 "snapshot" · guard.go:115 "verify"
guard.go:246 "restore" · recover.go:11 "recover" · remove.go:18 "remove [path]" · sync.go:15 "sync [branch-name]"
```

설치 바이너리는 develop tip보다 오래됐으므로 help만으로 확정하지 않고 develop 소스의 등록(`Use:`)과 설명(`Short:`)을 대조했다. 8개 하위 명령의 `Short:` 문자열이 help 설명과 글자까지 같다. 전문: `worktree-help.txt`, `develop-worktree-cobra-use.txt`.

스킬 `SKILL.md` §2가 안내하던 명령은 `new · list · switch · go · sync · remove · clean · status · config`다.

| 구분 | 명령 |
|---|---|
| 스킬에만 있고 CLI에 없음 | `new`, `list`, `switch`, `go`, `status`, `config` |
| CLI에만 있고 스킬에 없음 | `recover`, `done`, `snapshot`, `verify`, `restore` |

### 1.3 부재로 보고하기 전에 발견한 계기 결함 두 건

**아무것도 잡을 수 없는 검색.** 첫 스윕 `git grep -E 'worktree (new|…)\b'`는 0건(exit 1)이었다. 그런데 방금 읽은 `:123`에 "worktree new command"가 있었다. git grep의 확장 정규식이 `\b`를 단어 경계로 해석하지 않은 탓이다. `([^a-z]|$)`로 바꾸고 `:123`이 두 사본에서 모두 잡히는지(양성 대조 2건) 확인한 뒤 다시 쟀다.

**원리상 한 형태만 보는 검색.** 고친 패턴도 `worktree` 접두가 붙은 경우만 잡았다. `:153`(편집 뒤 `:159`)의 "create the worktree using **the new command** … **the switch command** … **the go command**"는 그 접두가 없어 잡히지 않았다. 편집 전 원본을 기준선 측정에 쓰려고 교체했다가 이 줄을 읽고 알게 됐다. "the X command" 형태를 잡는 두 번째 패턴을 추가했고 편집 전 원본에서 11줄을 잡는 것을 대조로 확인했다(§4.2). 첫 패턴만 믿었다면 `SKILL.md`에 없는 명령 4줄이 남은 채로 "고쳤다"고 보고했을 것이다.

### 1.4 재현되지 않은 "참고문서에는 sync가 없다"

감사 보고서는 "참고문서에는 현재 살아 있는 sync가 없다고 적혀 있다"고 쓴다. `findings.json`의 SK-11 근거는 help와 `root.go`의 8개 등록만 인용하고 그 참고문서의 파일은 인용하지 않는다.

부정 문장 필터를 먼저 합성 입력으로 대조했다(3줄 중 부정 2줄만 적중). 그 필터로 스킬 디렉터리의 `sync` 줄을 훑었다.

| 트리 | `sync` 줄 | 부정 문장 적중 |
|---|---|---|
| develop (`HEAD`)의 스킬 두 사본과 worktree 규칙 | 394 | 무관한 "Sync Check" 1건뿐 |
| 감사 기준 `main 2213871af`의 스킬 디렉터리 | 192 | 0 |

두 트리 모두 스킬 안에 "sync가 없다"는 문장은 없다. 가장 가까운 후보는 스킬 밖의 **템플릿 `AGENTS.md:288`**이다.

```
| `moai worktree` | Worktree lifecycle (list / snapshot / verify / restore) |
```

없는 `list`가 들어 있고 `sync · remove · clean · recover · done`이 빠졌다. 다만 목록에서 **빠진** 것일 뿐 "없다"는 문장은 아니며 감사가 가리킨 문서가 이것인지도 확인할 수 없었다. 이 카드에서는 고치지 않았다(§5).

## 2. 표적 수정으로 한정한 범위

두 항목 모두 감사의 `removal_safety`가 "Targeted correction only; verify canonical source and preserve foreign edits"여서, 감사가 인용한 `SKILL.md` 안에서 **없는 하위 명령을 부르는 줄**만 고쳤다.

### 2.1 고치지 않은 모듈·참고문서의 제거 명령

`git worktree`(실제 git 명령)를 뺀 `worktree <제거된 명령>` 언급:

| 파일 (템플릿 사본 기준) | 줄 수 |
|---|---|
| `modules/worktree-commands.md` | 41 |
| `references/examples.md` | 23 |
| `modules/troubleshooting.md` | 9 |
| `references/reference.md` | 7 |
| `modules/moai-adk-integration.md` | 2 |
| 그 밖의 모듈 5개 | 각 1 |
| **합계** | **89** (로컬 사본은 90) |

전문: `sk11-subcmd-moai-only.txt`. 이 수는 `worktree` 접두 패턴으로만 쟀으므로 "the X command" 형태는 들어 있지 않다(§1.3). 실제 규모는 이보다 크다. §2가 안내하는 `modules/worktree-commands.md`에는 "이 모듈의 명령 예시는 현행 명령 집합보다 먼저 쓰였으며 다르면 위 표와 `moai worktree --help`가 우선한다"는 문장을 붙였다.

### 2.2 실측만 하고 고치지 않은 있는 명령의 플래그

`SKILL.md`에는 없는 하위 명령 말고도 **있는 명령에 없는 플래그**를 붙인 서술이 있다. "명령 목록"과 다른 축이고 감사도 짚지 않아 고치지 않았다. 대신 참인 문장을 지우지 않도록 실측했다.

설치 바이너리 help (`worktree-subcmd-help.txt`):

| 명령 | 실제 플래그 |
|---|---|
| `sync` | `--base`, `--strategy` |
| `clean` | `--base`, `--json`, `--merged-only`, `--stale`, `--yes` |
| `remove` | `--force` |
| `done` | `--auto`, `--delete-branch`, `--force` |

develop 소스 `internal/cli/worktree`의 문자열 리터럴 대조 (`git grep -F -e '"<flag>"'`):

- 양성 대조 `"merged-only"`(`clean.go:34`)와 `"strategy"`(`sync.go:29`)는 적중했다.
- 스킬이 말하는 `"include"`, `"exclude"`, `"auto-resolve"`, `"interactive"`, `"template"`, `"developer"`, `"all-developers"`, `"team-overview"`, `"sync-check"`, `"shallow"`, `"background"`는 **적중 0**이다.

그래서 `clean`의 `--merged-only`를 말하는 `:163`은 **참이므로 남겼다.** 반면 `sync`의 include/exclude 패턴·auto-resolve·interactive 서술, template 플래그, developer 플래그는 없는 플래그다(§5).

## 3. 수리

Template-First 순서로 템플릿 사본을 먼저 고치고 로컬 사본의 같은 구역을 맞췄다. 두 사본은 이 구역 밖에서 원래 갈라져 있어 파일을 통째로 복사하지 않았다.

### 3.1 세 곳 (SK-07)

| 위치 | 수리 전 | 수리 후 |
|---|---|---|
| 합리화 표 | "(mode: plan) cannot write" / "tools list omits Write/Edit" | 쓸 수 있는 도구가 하나도 없을 때만 격리가 불필요하다. mode는 무시되어 아무것도 막지 않는다. Write·Edit·NotebookEdit를 빼는 것으로는 부족하며 Bash·다른 셸·MCP 도구·쓰기 에이전트를 띄우는 Agent도 쓸 수 있다. 그런 도구가 있는 한 읽기 전용은 보장되지 않고 격리는 낭비가 아니다. |
| 레드 플래그 | "Read-only agent spawned with isolation" | "Agent whose tools list holds no tool that can write, spawned with isolation" |
| 검증 체크리스트 | "verify mode: plan is sufficient" / "verify the tools list omits Write/Edit" | Write·Edit·NotebookEdit·Bash와 그 밖의 파일 변경 도구를 확인하고 하나라도 있으면 읽기 전용이 보장되지 않는다고 적음 |

카드의 [HARD]대로 문구를 지우고 끝내지 않았다. 실제로 쓰기를 막는 것(도구 목록)으로 대체했고 막는 것이 없는 경우는 "보장되지 않는다"고 명시했다.

### 3.2 없는 하위 명령을 부르는 줄 (SK-11)

| 위치 | 수리 |
|---|---|
| §2 전체 (명령 안내) | 진입은 launcher(`moai cc -w <name>`, `--spawn`), 조회는 `git worktree list`, 관리 명령 8개는 help 출력에서 옮긴 표. 표 아래에 help가 정본이라고 명시 |
| §3 `:123` | "worktree new command" → launcher가 격리된 워크트리로 들어간다 |
| §4 `:159` | "the new command … the switch command … the go command" → launcher 안내 |
| §4 `:167` | "The status command with sync-check option …" 문장 삭제 |
| 고급 절 팀 레지스트리 (템플릿 `:220`, 로컬 `:218`) | "The list command … and the status command …" 문장 삭제 |
| 고급 절 템플릿 (템플릿 `:232`, 로컬 `:230`) | "Configure custom templates through the config command …" 문장 삭제 |

로컬 사본의 옛 표기(`/moai:1-plan`, `/moai:2-run`, `/moai:3-sync`)는 범위 밖이라 그대로 두었다. 같은 문단 안의 플래그·설정 서술(§2.2)도 건드리지 않았다.

## 4. 검증

### 4.1 help 출력과 표의 글자 대조

help 출력의 COMMANDS 구역에서 설명 8개를 기계적으로 뽑아, 각각이 템플릿 표의 한 칸(`| <설명> |`)으로 들어 있는지 셌다.

```
help_desc_count=8
1 :: Sync worktree with base branch
1 :: Remove a worktree
1 :: Clean stale worktree references
1 :: Repair worktree registry
1 :: Complete worktree and cleanup
1 :: Capture working tree state snapshot for guard verification
1 :: Verify working tree state against snapshot + check agent response
1 :: Restore working tree to a snapshot's HEAD state
verbatim_found=8 (expect 8)
```

### 4.2 두 패턴과 같은 엔진 대조로 확인한 결함 문구 제거

두 패턴을 썼다.

- `P1`("the X command" 형태): `(^|[^a-z-])(new|list|switch|go|status|config) (command|subcommand)`
- `P2`("worktree X" 형태): `worktree (new|go|list|switch|status|config)([^a-z]|$)`, `git worktree` 줄 제외

앞선 양성 대조는 git grep이었고 이 검사는 BSD grep이다. 엔진이 달라 앞의 대조로는 이번 "0"을 보증하지 못하므로 같은 명령을 편집 전 `HEAD` 원본에 먼저 돌렸다.

**템플릿 사본**

```
CONTROL pre-edit: p1=11 (expect 11)  p2=2 (expect 2)
NOW template:     p1=0  (expect 0)   p2=0 (expect 0)
merged_only_kept=1 (expect 1)
mode_plan_claims=0  omits_write_edit_only=0  holds?_no_tool_that_can_write=3
```

**로컬 사본**

```
CONTROL pre-edit HEAD local: p1=11  p2=3  mode_plan=2
NOW local: p1=0 (expect 0)  p2=1 (expect 1: historical)  mode_plan=0  holds?_no_tool_that_can_write=3
208:Relationship to the retired `--team` flag: `moai worktree new --team` previously created a worktree ...
merged_only_kept=1 (expect 1)
```

로컬에 남은 한 줄은 제거된 플래그의 역사를 과거형으로 설명하는 문단이라 결함이 아니다.

새 문구 개수는 처음에 3을 기대했는데 2가 나왔다. 체크박스 문장의 주어가 복수("Teammates … hold")라 `holds`가 아니었기 때문이다. 틀린 것은 편집이 아니라 기대값의 문법이었으며 `holds?`로 다시 재서 3을 확인했다.

### 4.3 템플릿 중립성

```
synthetic_control=3 (expect 3)      # SPEC ID · 날짜 · 절대 경로가 담긴 합성 입력
added_lines=26
neutrality_hits=0
```

검사 패턴: SPEC ID, REQ 토큰, 날짜, `/Users/`, `CLAUDE.local`, 커밋 SHA 형태. 대상은 템플릿 `SKILL.md`의 추가 줄이다.

### 4.4 두 사본의 차이

```
differing_lines_final=43 (before edits 47)
edited phrases present in diff: 2
  22:< During Plan Phase Integration with /moai:1-plan, after SPEC creation, enter the worktree through the launcher: ...
  24:> During Plan Phase Integration with /moai plan, after SPEC creation, enter the worktree through the launcher: ...
  hunk: 159c159
```

적중 2줄은 한 덩어리(`159c159`)의 `<`·`>` 쌍이다. 두 줄의 차이는 원래부터 있던 `/moai:1-plan` 대 `/moai plan` 표기뿐이며 고친 문구는 두 사본에서 같다. 새로 만든 차이는 없다. 남은 43줄은 이 카드 이전부터 있던 무관한 차이다(§5).

### 4.5 영향 패키지 테스트

```
$ go test ./internal/template/... -count=1 -timeout 20m > .moai/reports/t628/gotest-template-final.txt 2>&1
TEMPLATE_FINAL_EXIT=1        # 파이프 없이 받은 명령의 종료 코드 (래퍼는 0을 보고)
FAIL  github.com/modu-ai/moai-adk/internal/template  35.270s
ok    github.com/modu-ai/moai-adk/internal/template/agentemit
ok    github.com/modu-ai/moai-adk/internal/template/commandemit
?     github.com/modu-ai/moai-adk/internal/template/scripts  [no test files]
panic_timeout_build=0

고유 실패 이름:
  1 --- FAIL: TestCatalogHashCoversSkillSubfiles
  1 --- FAIL: TestManifestHashFormat

CATALOG_HASH_UNSTABLE: moai-workflow-worktree stored hash=05920664…, computed hash=ea5542b8…
CATALOG_HASH_SKINNY:   moai-workflow-worktree hash=05920664… does not cover the deployed directory tree
```

두 실패 모두 이 카드가 고친 스킬 하나만 지목한다. 카탈로그는 스킬 **디렉터리 전체**에 해시를 걸어 두므로(`catalog.yaml:91-94`) 그 안의 파일을 바꾸면 저장된 해시가 맞지 않게 된다.

"편집이 원인"이라는 주장은 이름만으로 세우지 않고 **편집 전 원본으로 되돌려** 쟀다.

```
# 전제: 템플릿 스킬 디렉터리에서 바뀐 파일은 SKILL.md 하나뿐
$ git status --short -- internal/template/templates/.claude/skills/moai-workflow-worktree
 M internal/template/templates/.claude/skills/moai-workflow-worktree/SKILL.md

# 편집 전 원본으로 교체 (sha256 6d1f41f3…) 뒤 두 테스트만 실행
$ go test ./internal/template/ -run 'TestCatalogHashCoversSkillSubfiles|TestManifestHashFormat' -v -count=1
BASELINE_EXIT=0
--- PASS: TestCatalogHashCoversSkillSubfiles (0.02s)
--- PASS: TestManifestHashFormat (0.02s)
ok  	github.com/modu-ai/moai-adk/internal/template	0.444s

# 편집본 복원 (sha256 584e48d3…, 1차 편집 시점) — 이후 2차 편집을 이어서 적용
```

편집 전 원본에서는 두 테스트가 통과한다. 두 빨간 줄은 이 카드의 편집에서 비롯했으며 해소 방법은 해시 재생성(`gen-catalog-hashes.go --all`, `make build`의 한 단계)이다. 카드 지시에 따라 레인은 `make build`를 돌리지 않았다. 리드가 배치 끝에 빌드하면 녹색이 되리라 예상하지만 **그 녹색은 관측하지 않았다.** 기준선 입력 파일: `head-template-SKILL.md`, 결과: `baseline-catalog-tests.txt`.

## 5. 이 카드에서 고치지 않고 남긴 것

1. **카탈로그 해시 재생성.** 리드의 배치 끝 `make build` 소관이다(§4.5).
2. **모듈·참고문서의 제거 명령.** `worktree` 접두 형태만 89줄이고 "the X command" 형태는 세지 않았다(§2.1). 모듈 재작성은 별도 카드감이다.
3. **있는 명령의 없는 플래그**(§2.2). `SKILL.md`의 `sync` include/exclude·auto-resolve·interactive 서술(고급 절 동기화), template 플래그(고급 절 템플릿), developer 플래그와 팀 레지스트리 모드(고급 절 팀), 설정 키 `auto_create`·`auto_sync`·`cleanup_merged`·`worktree_root`(§4 설정 통합, 미측정).
4. **템플릿 `AGENTS.md:288`.** `moai worktree`를 "(list / snapshot / verify / restore)"로 적는다. 없는 `list`가 있고 5개가 빠졌다(§1.4).
5. **같은 과장이 규칙에도 있다.** `agent-authoring.md:235`, `worktree-integration.md:256`, `manager-lead.md:181`(로컬)·`:183`(템플릿)이 "읽기 전용은 `Explore` 또는 Write/Edit를 뺀 `tools:` 목록에 기댄다"고 쓴다. 특히 `worktree-integration.md:256`은 한 문장 안에서 "쓸 수 없을 때만 읽기 전용"이라는 정의와 Bash를 가진 예시를 함께 들어 스스로 어긋난다.
6. **`internal/cli/worktree/root.go`의 Long 설명.** "Supports creating, syncing, removing, and cleaning worktrees"라고 하지만 생성 명령은 없다.
7. **두 사본의 무관한 차이 43줄.** 로컬의 옛 슬래시 표기, 참고문서 경로(`examples.md` 대 `references/examples.md`), 로컬에만 있는 역사 문단 등이다.

## 6. Gaps (관측하지 않은 것)

- **런타임 권한 우회는 시험하지 않았다.** Bash를 가진 에이전트가 실제로 파일을 쓸 수 있는지, 샌드박스나 권한 프롬프트가 막는지는 재지 않았다. 문구는 "보장되지 않는다"는 보수적 표현에 머문다.
- **`Explore`가 Bash를 가진다는 근거는 이 세션의 런타임 에이전트 목록뿐이다.** 배포 문서에는 `Explore`의 도구 목록이 없으며 Claude Code 버전에 따라 달라질 수 있다.
- **`make build`·임베드 검사·해시 재생성 뒤의 녹색은 관측하지 않았다**(§4.5).
- **설치 바이너리는 develop tip보다 오래됐다.** help를 develop 소스의 `Use:`·`Short:`·플래그 리터럴과 대조해 일치를 확인했으나 develop tip으로 빌드한 바이너리의 help를 직접 보지는 않았다.
- **플래그 부재 대조는 `internal/cli/worktree` 패키지 안의 문자열 리터럴로 한정된다.** 다른 패키지에 같은 이름의 플래그가 정의돼 있을 가능성은 재지 않았다.
- **§2.1의 89줄에는 "the X command" 형태가 들어 있지 않다.** 모듈·참고문서의 실제 규모는 재지 않았다.
- **감사가 말한 "참고문서"는 특정하지 못했다**(§1.4).
