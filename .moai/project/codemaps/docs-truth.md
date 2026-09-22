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
> **마지막 손 갱신**: 2026-09-22, 워크트리 `.claude/worktrees/t1069`, 브랜치 `WT-graph-restamp`, HEAD `0314801c2`.
> §1 에이전트 파일 목록(13 retained로 정정 — mission-governor가 카탈로그에 편입됐다), §2 상태 enum 개수,
> §3 필수 필드 12개와 줄 위치, §4.1 `moai --help` 렌더 그룹(HEAD에서 빌드한 바이너리로 다시 렌더 — 그룹 불변)과
> 등록 수치, §4.2 명령 파일 목록, §5 GLM 상수 값을 이 트리에서 다시 검증했다.
> §4.1의 `codex` 런처 서술(verb 라우팅·init-offer 게이트)은 이번에 다시 읽지 않았고, 그 Source 문구의
> 과거 검증일을 따른다.

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
| 12 | `mission-governor` | MoAI-custom — no Selection Decision Tree row | GTD auto-mission decision role: read-only decision from a sealed mission snapshot, dispatched by that workflow rather than selected by the tree |
| 13 | `Explore` | Anthropic built-in | Read-only codebase exploration (no MoAI file — invoked directly) |

Class breakdown: Manager ×6 (`manager-spec`, `manager-develop`, `manager-docs`, `manager-git`, `manager-design`, `manager-lead`) · Evaluator ×2 (`plan-auditor`, `sync-auditor`) · Builder ×1 (`builder-harness`) · Advisor ×1 (`super-advisor`) · Specialist ×1 (`e2e-tester`) · Mission decision ×1 (`mission-governor` — deliberately classless in the Selection Decision Tree; CLAUDE.md §4 states it is dispatched by the GTD auto-mission workflow) · Anthropic built-in ×1 (`Explore`).

**Archived agents**: 12 legacy agent names are archived and MUST NOT be spawned. The full archived-name list + per-archived-agent migration table lives in `.claude/rules/moai/workflow/archived-agent-rejection.md` (consult that file rather than naming the archived agents here, to keep this checklist free of archived-name leakage).

**Source:** `ls -1 .claude/agents/moai/*.md` (= 12 MoAI-custom files) + CLAUDE.md §4 Retained Agents line + `.claude/rules/moai/workflow/archived-agent-rejection.md` (archived-agent migration table).

**Re-verified 2026-09-22 (HEAD `0314801c2`) — exhaustive, no sampling.** `find .claude/agents/moai -maxdepth 1 -name '*.md' | wc -l` → **12**. The tree listing and the §1 table rows 1-12 are compared name-by-name below; row 13 (`Explore`) is an Anthropic built-in with no file, so it is expected to be absent from the tree.

> **[미결 드리프트 해소]** 앞 판(2026-09-18, HEAD `a851b205c`)은 `mission-governor.md`가 트리에는
> 있지만 `CLAUDE.md` §4 retained 목록에 없는 것을 미결 드리프트로 적어 뒀다. 이번 검증에서
> `CLAUDE.md` §4는 "exactly **13 retained agents** (12 MoAI-custom + 1 built-in `Explore`)"로
> 갱신돼 있고 retained 열거에 `mission-governor`를 명시한다 — 카탈로그 소유 문서가 판정을
> 내렸으므로 이 파일이 §1을 그 결정에 맞춰 13으로 정정한다.

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
| `mission-governor.md` | 12 `mission-governor` |
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

Top-level verbs rendered by `moai --help`, in the render's own grouping (re-rendered 2026-09-22 from a binary built at HEAD `0314801c2` with `go build ./cmd/moai` — all four groups are byte-equivalent in verb composition to the 2026-09-18 render below; the only new root command since, `jev-suggest`, is `Hidden: true` and renders nowhere):

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

**Source:** `moai --help` rendered output of a binary built from HEAD `0314801c2` (2026-09-22) + `grep -rn 'AddCommand(' internal/cli --include='*.go' | grep -v _test | wc -l` (**220** non-test calls) + `grep -rn 'rootCmd.AddCommand(' internal/cli --include='*.go' | grep -v _test | wc -l` (**66** root registrations across the package; **31**의 `rootCmd.AddCommand`가 `internal/cli/root.go`의 `init()` 안에 있다 — +1은 숨은 `newJevSuggestCmd()`) + `find internal/cli -name '*.go' ! -name '*_test.go' | wc -l` (**325**; 앞 판 표기 315는 t1066 이전 수치였다).

HOME 상태의 사용자 진입점은 `moai migrate home-state`입니다. 기본 실행은 dry-run이고,
실제 쓰기는 `--apply --verified-live`를 함께 요구합니다. 복구 표면은 하위 명령 `recover`와
`rollback`, Factory 레거시 인계 복구는
`moai factory handoff recover-resume --id <id> --expected-token <token> --decision <fail|requeue>`입니다.
이 명령의 존재는 운영 데이터 이전 완료를 뜻하지 않습니다.

The `codex` launcher: closed-set verb routing `{bare, cli, app}` (launch, `--spawn` optional, `-w <worktree>` optional, `--` passthrough) × `{status}` (readout, rc 0, starts nothing); an unknown token is rejected with a one-line usage diagnostic (rc 1), never routed to a launch. Downstream of routing, an argv-translation table forwards a verb to the child only where it names a real codex subcommand — `app` is forwarded, the bare form and `cli` are moai-side synonyms and are not, so the child receives only the operator's own tail. `-w` is consumed by moai (it points the child's working directory at an EXISTING worktree and never creates one) and is not forwarded. All three launching forms pass through ONE init-offer gate function immediately before launching — the gate takes no `--spawn` parameter, accepts exactly `y`/`yes` at its prompt, exits 130 on decline (cancel, not error) and 1 on failure or a non-interactive session (report only, no prompt issued), and on acceptance delegates to the `moai init --agent codex` wiring generator exactly once, then links `AGENTS.md` ↔ `CLAUDE.md` (connection-only: at most one appended `@AGENTS.md` / `@CLAUDE.local.md` directive per file, path-containment guard runs before any read or write, writes are per-file temp+rename, idempotent on re-run).

**Source (gate):** `internal/cli/codex_launcher.go` (verb routing, single gate call site in `runCodexLaunch`), `internal/cli/codex_init.go` (gate + seams), `internal/cli/codex_contract.go` (link contract). Verified 2026-08-28 by direct read; the routing and argv-translation sentences re-verified 2026-09-01 by direct read of `codexVerbRouting` / `codexChildSubcommand` on the tree that reversed the default.

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
