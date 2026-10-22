# docs-truth.md — Canonical Facts Checklist for the Docs-v3 Cohort

> Navigation aid, NOT a new SSOT. Every fact below is cross-referenced to
> its ground-truth source file. Later cohort SPECs (README / DOCSITE /
> COVERAGE / i18n) re-verify each fact against the cited source, NOT
> against this file.
> If a fact here has drifted from its source, update this file FIRST
> (separate commit) before proceeding with the rewrite.

> **[HARD] 이 파일은 `/moai codemaps` 산출물이 아니다 — 손으로 유지된다.**
> 생성기가 내는 것은 `overview.md` · `modules.md` · `dependencies.md` ·
> `entry-points.md` · `data-flow.md` **5개**이며, 스킬 워크플로 문서에 `docs-truth`는
> 한 번도 등장하지 않는다. 그래서 `ls .moai/project/codemaps/`가 7항목을 보이며
> 통과하는 동안 이 파일만 조용히 낡을 수 있다. **재생성 때마다 이 파일을 손으로
> 함께 갱신하고, 그 사실을 재생성 증거와 분리해 기록한다.**
>
> **마지막 손 갱신**: 2026-10-23, 워크트리 `.moai/worktrees/t1909`, 브랜치 `WT-moai-codex-codexcmd`, HEAD `d5c083e86d`; card t1909 재생성 판.
> §4.1의 `codex` 런처 단락을 감사 부모 현행 서술로 갈았고(삭제된 런처 서술은 사료로 보존), `moai --help`
> 렌더 그룹을 이 트리 빌드 바이너리로 다시 렌더해 t1909 판 재렌더 노트를 남겼다. 등록 수치 넷(AddCommand
> 252 · rootCmd 68 · root.go init() 29 · cli 비테스트 481/루트 382)을 같은 트리에서 다시 쟀다.
> §1·§2·§3·§4.2·§5는 이번 카드 변경과 무관해 앞 판(2026-09-22, HEAD `0314801c2`) 검증을 따른다.

---

## §1. Agent Catalog (13 retained)

The MoAI agent catalog consists of exactly **13 retained agents** (12 MoAI-custom + 1 Anthropic built-in `Explore`).

| # | Agent | Class | Phase scope |
|---|-------|-------|-------------|
| 1 | `manager-spec` | core/manager | Plan-phase artifact authoring (spec/plan/acceptance/research/design) |
| 2 | `manager-develop` | core/manager | Run-phase implementation (cycle_type ∈ {ddd, tdd, autofix}) |
| 3 | `manager-docs` | core/manager | Sync-phase documentation (CHANGELOG, README, frontmatter transitions) |
| 4 | `manager-git` | core/manager | PR creation per Tier-based routing + Late-Branch closure |
| 5 | `plan-auditor` | meta/evaluator | Independent plan-phase audit, bias prevention, GEARS compliance |
| 6 | `sync-auditor` | meta/evaluator | Independent skeptical quality assessment, 4-dimension scoring |
| 7 | `builder-harness` | builder | Dynamic project-specific harness specialist generation |
| 8 | `super-advisor` | meta/advisor | On-demand high-reasoning consultation (E1-E4 escalation) |
| 9 | `manager-design` | core/manager | Design-phase collaboration (Claude Design bidirectional sync, D1-D5) |
| 10 | `e2e-tester` | core/specialist | E2E test execution (web/mobile/desktop journey scripting) |
| 11 | `manager-lead` | core/manager (sole Agent-carrier) | Multi-milestone Tier L coordination + kanban/factory lead role (depth-2 sealed) |
| 12 | `manager-todo` | MoAI-custom — no Selection Decision Tree row | Todo-queue management (queue lifecycle, `/moai:todo --auto` serial cycle, dispatch guidance, Jev display-only consultation); the read-only decision from a sealed mission snapshot continues as its judgment sub-role, dispatched by that workflow |
| 13 | `Explore` | Anthropic built-in | Read-only codebase exploration (no MoAI file — invoked directly) |

Class breakdown: Manager ×6 (`manager-spec`, `manager-develop`, `manager-docs`, `manager-git`, `manager-design`, `manager-lead`) · Evaluator ×2 (`plan-auditor`, `sync-auditor`) · Builder ×1 (`builder-harness`) · Advisor ×1 (`super-advisor`) · Specialist ×1 (`e2e-tester`) · Mission decision ×1 (`manager-todo` — deliberately classless in the Selection Decision Tree; CLAUDE.md §4 states its judgment sub-role is dispatched by the GTD auto-mission workflow) · Anthropic built-in ×1 (`Explore`).

**Archived agents**: 12 legacy agent names are archived and MUST NOT be spawned. The full archived-name list + per-archived-agent migration table lives in `.claude/rules/moai/workflow/archived-agent-rejection.md` (consult that file rather than naming the archived agents here, to keep this checklist free of archived-name leakage).

**Source:** `ls -1 .claude/agents/moai/*.md` (= 12 MoAI-custom files) + CLAUDE.md §4 Retained Agents line + `.claude/rules/moai/workflow/archived-agent-rejection.md` (archived-agent migration table).

**Re-verified 2026-09-22 (HEAD `0314801c2`) — exhaustive, no sampling.** `find .claude/agents/moai -maxdepth 1 -name '*.md' | wc -l` → **12**. The tree listing and the §1 table rows 1-12 are compared name-by-name below; row 13 (`Explore`) is an Anthropic built-in with no file, so it is expected to be absent from the tree.

> **[미결 드리프트 해소]** 앞 판(2026-09-18, HEAD `a851b205c`)은 자동 임무 판정 에이전트 파일이 트리에는
> 있지만 `CLAUDE.md` §4 retained 목록에 없는 것을 미결 드리프트로 적어 뒀다. 카탈로그 소유 문서가
> 13 retained 판정을 내려 §1을 13으로 정정했고, 이후 그 역할은 manager-todo로 개명·재정의됐다
> (todo-queue 관리 주 임무 + 읽기 전용 봉인 스냅숏 판정 하위 역할) — 이 표는 그 현행 이름을 따른다.

| Tree file (`.claude/agents/moai/`) | §1 table row |
|---|---|
| `builder-harness.md` | 7 `builder-harness` |
| `e2e-tester.md` | 10 `e2e-tester` |
| `manager-design.md` | 9 `manager-design` |
| `manager-develop.md` | 2 `manager-develop` |
| `manager-docs.md` | 3 `manager-docs` |
| `manager-git.md` | 4 `manager-git` |
| `manager-lead.md` | 11 `manager-lead` |
| `manager-spec.md` | 1 `manager-spec` |
| `manager-todo.md` | 12 `manager-todo` |
| `plan-auditor.md` | 5 `plan-auditor` |
| `super-advisor.md` | 8 `super-advisor` |
| `sync-auditor.md` | 6 `sync-auditor` |

**대조 결과: 표의 12행과 트리의 12파일이 일치한다.** 표에만 있고 트리에 없는 행은 0(built-in `Explore` 제외), 트리에만 있고 표에 없는 파일도 0. 같은 스캔에서 `.claude/agents/harness/` 아래 파일들도 관측되나 이것들은 user-owned harness specialist 이며 retained catalog 의 원소가 아니다 — §1 의 수에 들어가지 않는 것이 정상이다.

---

## §2. SPEC Status Enum (8 lowercase values)

The SPEC status enum is the exact lowercase 8-value set:

`draft` · `planned` · `in-progress` · `implemented` · `completed` · `superseded` · `archived` · `rejected`

Lifecycle flow:
```
draft → planned → in-progress → implemented → completed
                                         ↓
                               superseded | archived | rejected
```

**Source:** `internal/spec/status.go` `ValidStatuses` slice. Re-verified 2026-09-22 (HEAD `0314801c2`) with the same command, same result: `grep -cE '"draft"|"planned"|"in-progress"|"implemented"|"completed"|"superseded"|"archived"|"rejected"' internal/spec/status.go` → 8. Schema SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Enum.

---

## §3. SPEC Frontmatter Schema (12 required fields)

Every `spec.md` MUST contain exactly these 12 fields in YAML frontmatter:

`id` · `title` · `version` · `status` · `created` · `updated` · `author` · `priority` · `phase` · `module` · `lifecycle` · `tags`

Rejected snake_case aliases (silently dropped by the YAML decoder): `created_at:` → use `created:`; `updated_at:` → use `updated:`; `labels:` → use `tags:`; `spec_id:` → use `id:`.

**Source:** `internal/spec/lint.go` `FrontmatterSchemaRule.Check()` required slice (12 entries of the form `{"<field>", fm.<Field>}`). Re-verified 2026-09-22 (HEAD `0314801c2`): the slice now starts at `internal/spec/lint.go:1192` (`required := []struct {`) and carries exactly 12 `fm.<Field>` entries, in the order listed above. Schema SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md` § Canonical 12 Required Fields. Lint rule code: `FrontmatterInvalid` (Warning severity).

---

## §4. CLI Subcommand Surface

### §4.1 `moai` terminal verbs (human-facing)

Top-level verbs rendered by `moai --help`, in the render's own grouping (re-rendered 2026-09-22 from a binary built at HEAD `0314801c2` with `go build ./cmd/moai` — all four groups are byte-equivalent in verb composition to the 2026-09-18 render below. The one hidden root command that existed at that HEAD (a skill-suggestion anchor) has since been withdrawn by SPEC-JEV-GUARD-001 — it was never rendered by `--help` and is absent from every group):

| Render group | Verbs |
|------------------|-------|
| COMMANDS | `factory`, `gate`, `goal`, `integration`, `config`, `ast-grep`, `ast-edit`, `migration`, `harness`, `mcp-server`, `mcp`, `plan`, `feedback`, `slot`, `help`, `completion` |
| LAUNCH COMMANDS | `cc`, `glm`, `codex` |
| PROJECT COMMANDS | `init`, `status`, `doctor`, `update`, `migrate`, `pr` |
| TOOLS | `hook`, `spec`, `session`, `mx`, `loop`, `handoff`, `model`, `constitution`, `state`, `epic`, `github`, `graph`, `gtd`, `lsp`, `memory`, `profile`, `research`, `worktree`, `agent`, `workflow`, `telemetry`, `tokens`, `clean`, `skills`, `chain`, `tool-policy`, `tool`, `inventory`, `preference`, `inbox`, `todo`, `verify`, `version`, `web` |

> **2026-09-18 변화.** 앞 판(2026-09-10 렌더) 대비 `slot`(COMMANDS)과 `gtd`(TOOLS)가 더해졌고,
> `cg`가 LAUNCH COMMANDS에서 사라졌다. `cg`는 `internal/cli/root.go`의 `trivialCommands`에 은퇴
> 토큰으로만 남아 있다. `gtd`는 `todo` 명령 트리를 감싼 두 번째 이름으로, 같은 큐를 본다.

> **정정(2026-09-08).** 직전 판은 이 자리에 렌더의 그룹이 아니라 손으로 묶은 5분류
> (Project / Launchers / Autonomous-Dev / Governance / Tools-Infra)를 실었고, 그 분류는
> `moai --help` 출력에 존재하지 않습니다. 그리고 그 표는 `skills`를 빠뜨렸습니다 —
> `skills`는 `root.go:177`의 `newSkillsCmd()`로 등록된 라이브 루트 명령입니다.
> 이후 판은 렌더가 실제로 내는 4개 그룹을 그대로 옮깁니다.

Additional note: `statusline` is a root-registered command but `Hidden: true` in `moai --help`; `help` and `completion` are cobra-generated. The `run` verb exists ONLY as a `moai migration` subcommand (`internal/cli/migration.go`) — there is NO standalone `moai run` or `moai sync` root command; the plan/run/sync workflow lives in the `/moai` Claude Code skill set (§4.2).

> **t1909 판 재렌더(2026-10-23, HEAD `d5c083e86d` 트리에서 빌드한 바이너리).** 위 표의 그룹 조성이
> 바뀌었다 — COMMANDS에 `bundle`·`contract`·`decision`이 더했고, TOOLS에서 `model`(t1246 은퇴 잔존
> 표기 정정)·`skills`·`tool`(둘 다 card t1909 철수)이 빠지고 `jev`·`ci-verdict`·`paths`가 더했다.
> LAUNCH COMMANDS는 `cc`·`glm`·`codex` 셋이지만 `codex`는 감사 부모다(§ 아래 런처 단락).
> 등록 수치는 아래 Source 문단의 t1909 판 재측정을 따른다.

**Source:** `moai --help` rendered output of a binary built from HEAD `0314801c2` (2026-09-22) + `grep -rn 'AddCommand(' internal/cli --include='*.go' | grep -v _test | wc -l` (**220** non-test calls) + `grep -rn 'rootCmd.AddCommand(' internal/cli --include='*.go' | grep -v _test | wc -l` (**66** root registrations across the package; **31**의 `rootCmd.AddCommand`가 `internal/cli/root.go`의 `init()` 안에 있다 — +1은 숨은 `newJevSuggestCmd()`) + `find internal/cli -name '*.go' ! -name '*_test.go' | wc -l` (**325**; 앞 판 표기 315는 t1066 이전 수치였다). **t1443 판 재측정(2026-10-02, HEAD `4bf547bca` 트리)**: AddCommand **239** · rootCmd.AddCommand **66** 불변 · root.go init() **30**(숨은 제안 명령 철수 반영 — 위 정정 문단과 일치) · cli 비테스트 파일 **402**(루트 321). 등록 줄 diff는 앵커 `c2703f698` 이후 0 — 동사 조성 변화 없음(위 표 유효). **t1456 판 재측정(2026-10-03, HEAD `5501c06af` 트리)**: AddCommand **239** · rootCmd.AddCommand **66** · root.go init() **30** 전부 불변(`moai factory relaunch`는 기존 등록 줄 인자 행, `moai verify audit-plan`은 `verifyExtraCommands` 레지스트리 합류 — 동사 조성의 하위 동사 확장일 뿐 위 표의 루트 동사 유효) · cli 비테스트 파일 **406**(루트 325). **t1909 판 재측정(2026-10-23, HEAD `d5c083e86d` 트리)**: AddCommand **252** · rootCmd.AddCommand **68** · root.go init() **29**(skills·tool 두 등록 철수 반영) · cli 비테스트 파일 **481**(루트 **382** — card t1909가 24개를 빼고 2개를 더한 뒤). 위 표의 그룹 조성은 앞 문단의 t1909 판 재렌더 노트를 따른다.

HOME 상태의 사용자 진입점은 `moai migrate home-state`입니다. 기본 실행은 dry-run이고,
실제 쓰기는 `--apply --verified-live`를 함께 요구합니다. 복구 표면은 하위 명령 `recover`와
`rollback`, Factory 레거시 인계 복구는
`moai factory handoff recover-resume --id <id> --expected-token <token> --decision <fail|requeue>`입니다.
이 명령의 존재는 운영 데이터 이전 완료를 뜻하지 않습니다.

The `codex` command (t1909 판 — **런처가 아니라 감사 부모다**): `codexCmd`는 `rootCmd`에 등록되지만 RunE가 없어 어느 verb도 일반 Codex 실행을 시작하지 않는다. 남는 서브커맨드는 `codex audit <role> [--out <path>]`와 `codex role-audit` 둘이며(`codex_audit_launch.go`의 `init()`이 등록 — 읽기 전용 `codex exec -s read-only` 감사 런처), 둘 다 `-c`로 MCP 서버를 끄는 감사 인자와 `.moai/reports/` 리포트 경로 제한을 공유한다. `moai codex`를 동사 없이 부르면 아무것도 시작하지 않는다. **아래 단락은 삭제된 런처의 기록이다.** 런처는 closed-set verb routing `{bare, cli, app}` (launch, `--spawn` optional, `-w <worktree>` optional, `--` passthrough) × `{status}` (readout, rc 0, starts nothing)였고, 알 수 없는 토큰은 한 줄 사용 진단(rc 1)으로 거부했다. argv 번역 표는 실제 codex 서브커맨드 이름만 자식에게 전달했고(`app`은 전달, bare·`cli`는 moai 쪽 동의어라 미전달), `-w`는 moai가 소비해 존재하는 워크트리만 열었다(새 트리를 만들지 않는다). 세 실행 형태는 모두 실행 직전 하나의 init-offer 게이트 함수를 통과했다 — `y`/`yes`만 받고, 거절은 130(cancel), 실패·비대화형은 1(보고만, 프롬프트 없음), 수락 시 `moai init --agent codex` 배선 생성기를 한 번만 돌린 뒤 `AGENTS.md` ↔ `CLAUDE.md`를 연결했다(connection-only: 파일당 `@AGENTS.md` / `@CLAUDE.local.md` 지시문 최대 한 줄, 경로 포함 가드가 모든 읽기·쓰기에 선행, 파일별 temp+rename 쓰기, 재실행 멱등).

**Source (gate):** `the retired codex launcher (removed card t1909)` (verb routing, single gate call site in `runCodexLaunch`), `the retired codex init surface (removed card t1909)` (gate + seams), `the retired codex link contract (removed card t1909)` (link contract). Verified 2026-08-28 by direct read; the routing and argv-translation sentences re-verified 2026-09-01 by direct read of `codexVerbRouting` / `codexChildSubcommand` on the tree that reversed the default. **t1909 판: 위 세 소스는 트리에서 삭제됐다** — 현행 감사 부모의 Source는 `internal/cli/codex_cmd.go`(등록)와 `internal/cli/codex_audit_launch.go`(서브커맨드·런처)다.

### §4.2 `/moai` Claude Code skill set (17 commands)

The complete `/moai` slash-command set in `.claude/commands/moai/`:

`clean` · `codemaps` · `e2e` · `feedback` · `fix` · `gate` · `goal` · `gtd` · `harness` · `loop` · `mx` · `plan` · `project` · `review` · `run` · `sync` · `todo`

(17 files total)

**Source:** `find .claude/commands/moai -maxdepth 1 -name '*.md' | wc -l` → 17. Re-verified 2026-09-22 (HEAD `0314801c2`): the listing returns exactly the 17 names above and nothing else (`brain`/`coverage`/`design` are NOT present as standalone command files). `gtd.md` is the one file added since the 2026-09-08 verification.

이 명령 소스들은 codex 쪽으로도 발행됩니다 — `internal/template/commandemit`이 각각을
`.agents/skills/moai-<command>/SKILL.md`로 내보내며 본문은 바이트 동일 verbatim 입니다.

---

## §5. GLM → Claude Tier Mapping (full tier-models table)

The GLM→Claude-tier model mapping reflects the glm-5.3-flash default activation (SPEC-GLM-FLASH-DEFAULT-001). The full tier-models table:

| Constant | Value | Claude tier |
|----------|-------|-------------|
| `DefaultGLMBaseURL` | `https://api.z.ai/api/anthropic` | (z.ai gateway) |
| `DefaultGLMHigh` | `glm-5.3-flash` | High |
| `DefaultGLMMedium` | `glm-5.3-flash` | Medium |
| `DefaultGLMLow` | `glm-5.3-flash` | Low |
| `DefaultGLMSonnet` | `glm-5.3-flash` | Sonnet |
| `DefaultGLMHaiku` | `glm-5.3-flash` | Haiku |
| `DefaultGLMOpus` | `glm-5.3-flash` | Opus |
| `DefaultGLMFable` | `glm-5.3` | Fable |

`glm-5.3-flash` is the sparse-attention GLM-5.3-Flash variant (1M context). The `[1m]` 1M-context suffix is NOT hardcoded in `defaults.go` — the constants carry the bare model ids. The `[1m]` suffix is expanded at the launcher layer (`internal/cli/launcher.go` — `expandModelString` / `splitModelSuffix`) only when the 1M-context variant is explicitly requested; Claude Code parses and strips the suffix before the upstream API call because z.ai rejects a verbatim `[1m]`.

Additional GLM models available but not default-mapped: `glm-4.5`, `glm-4.6`, `glm-4.7`, `glm-4.5-air`, `glm-5.1`, `glm-5.2`, `glm-5-turbo`.

**Source:** `internal/config/defaults.go` — `DefaultGLMBaseURL` line 166, the `DefaultGLM53Flash`/`DefaultGLM53` values lines 199-200, the tier mapping lines 201-204 and 223-225 (DefaultGLM* constants block). Re-verified 2026-09-22 (HEAD `0314801c2`): `grep -nE 'DefaultGLM(High|Medium|Low|Fable|Opus|Sonnet|Haiku|BaseURL|53Flash|53) ' internal/config/defaults.go` → same line numbers and same values as the 2026-09-18 verification (`DefaultGLMHigh`(201) · `Medium`(202) · `Low`(203) · `Haiku`(223) · `Sonnet`(224) · `Opus`(225) resolve to `DefaultGLM53Flash`; `DefaultGLMFable`(204) to `DefaultGLM53`). Launcher suffix expansion: `expandModelString` (launcher.go:1139) / `splitModelSuffix` (launcher.go:1157).

---

## How to use this checklist

1. **Before rewriting any docs fact** (agent count, status value, frontmatter field, CLI verb, GLM model), re-verify it against the cited source file in THIS commit's tree.
2. **If a fact has drifted** from its source, open a SEPARATE commit updating this file FIRST, then proceed with the rewrite.
3. **Do NOT cite this file as ground truth** in user-facing docs — cite the primary source (`internal/spec/status.go`, etc.). This file is a navigation aid that points AT the sources.
4. **Neutrality**: this file lives under `.moai/project/codemaps/` which is NOT covered by the template-neutrality CI guard. Keep it free of internal SPEC-ID / REQ / Audit-citation leakage except for the self-reference to `SPEC-V3R6-DOCS-CODEMAPS-V3-001` (this file's authoring SPEC).
