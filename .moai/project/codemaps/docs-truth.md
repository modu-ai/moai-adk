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
> **마지막 손 갱신**: 2026-09-08, 워크트리 `.claude/worktrees/t475`, HEAD `52f863f36`.
> §1~§5 전 항목을 인용된 원천에 대해 재검증했다(§1은 전수 대조, 표본 추출 없음).

---

## §1. Agent Catalog (12 retained)

The MoAI agent catalog consists of exactly **12 retained agents** (11 MoAI-custom + 1 Anthropic built-in `Explore`).

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
| 12 | `Explore` | Anthropic built-in | Read-only codebase exploration (no MoAI file — invoked directly) |

Class breakdown: Manager ×6 (`manager-spec`, `manager-develop`, `manager-docs`, `manager-git`, `manager-design`, `manager-lead`) · Evaluator ×2 (`plan-auditor`, `sync-auditor`) · Builder ×1 (`builder-harness`) · Advisor ×1 (`super-advisor`) · Specialist ×1 (`e2e-tester`) · Anthropic built-in ×1 (`Explore`).

**Archived agents**: 12 legacy agent names are archived and MUST NOT be spawned. The full archived-name list + per-archived-agent migration table lives in `.claude/rules/moai/workflow/archived-agent-rejection.md` (consult that file rather than naming the archived agents here, to keep this checklist free of archived-name leakage).

**Source:** `ls -1 .claude/agents/moai/*.md` (= 11 MoAI-custom files) + CLAUDE.md §4 Retained Agents table + `.claude/rules/moai/workflow/archived-agent-rejection.md` (archived-agent migration table).

**Verified 2026-09-08 (HEAD `52f863f36`) — exhaustive, no sampling.** `find .claude/agents/moai -maxdepth 1 -name '*.md' | wc -l` → **11**. The tree listing and the §1 table rows 1-11 are compared name-by-name below; row 12 (`Explore`) is an Anthropic built-in with no file, so it is expected to be absent from the tree.

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
| `plan-auditor.md` | 5 `plan-auditor` |
| `super-advisor.md` | 8 `super-advisor` |
| `sync-auditor.md` | 6 `sync-auditor` |

**대조 결과: 11/11 일치, 양방향으로 잉여 없음.** 트리에만 있고 표에 없는 파일 0, 표에만 있고 트리에 없는 행 0(built-in `Explore` 제외). 같은 스캔에서 `.claude/agents/harness/` 아래 10개 파일도 관측되나 이것들은 user-owned harness specialist 이며 retained catalog 의 원소가 아니다 — §1 의 수 11 에 들어가지 않는 것이 정상이다.

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

**Source:** `internal/spec/status.go` `ValidStatuses` slice. Verified 2026-09-08 (HEAD `52f863f36`): `grep -cE '"draft"|"planned"|"in-progress"|"implemented"|"completed"|"superseded"|"archived"|"rejected"' internal/spec/status.go` → 8. Schema SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Enum.

---

## §3. SPEC Frontmatter Schema (12 required fields)

Every `spec.md` MUST contain exactly these 12 fields in YAML frontmatter:

`id` · `title` · `version` · `status` · `created` · `updated` · `author` · `priority` · `phase` · `module` · `lifecycle` · `tags`

Rejected snake_case aliases (silently dropped by the YAML decoder): `created_at:` → use `created:`; `updated_at:` → use `updated:`; `labels:` → use `tags:`; `spec_id:` → use `id:`.

**Source:** `internal/spec/lint.go` `FrontmatterSchemaRule.Check()` required slice (12 entries of the form `{"<field>", fm.<Field>}`). Verified 2026-09-08 (HEAD `52f863f36`): the slice is at `internal/spec/lint.go:982-998` and carries exactly 12 entries, in the order listed above. Schema SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md` § Canonical 12 Required Fields. Lint rule code: `FrontmatterInvalid` (Warning severity).

---

## §4. CLI Subcommand Surface

### §4.1 `moai` terminal verbs (human-facing)

Top-level verbs rendered by `moai --help`, in the render's own grouping (2026-09-08 actual render, `./bin/moai --help`):

| Render group | Verbs |
|------------------|-------|
| COMMANDS | `gate`, `goal`, `integration`, `config`, `ast-grep`, `ast-edit`, `migration`, `harness`, `mcp-server`, `mcp`, `plan`, `feedback`, `help`, `completion` |
| LAUNCH COMMANDS | `cc`, `glm`, `cg`, `codex` |
| PROJECT COMMANDS | `init`, `status`, `doctor`, `update`, `migrate`, `pr` |
| TOOLS | `hook`, `spec`, `session`, `mx`, `loop`, `handoff`, `model`, `constitution`, `state`, `epic`, `github`, `graph`, `lsp`, `memory`, `profile`, `research`, `worktree`, `agent`, `workflow`, `telemetry`, `tokens`, `clean`, `skills`, `chain`, `tool-policy`, `inventory`, `preference`, `inbox`, `todo`, `verify`, `version`, `web` |

> **정정(2026-09-08).** 직전 판은 이 자리에 렌더의 그룹이 아니라 손으로 묶은 5분류
> (Project / Launchers / Autonomous-Dev / Governance / Tools-Infra)를 실었고, 그 분류는
> `moai --help` 출력에 존재하지 않습니다. 그리고 그 표는 `skills`를 빠뜨렸습니다 —
> `skills`는 `root.go:177`의 `newSkillsCmd()`로 등록된 라이브 루트 명령입니다.
> 이 판은 렌더가 실제로 내는 4개 그룹을 그대로 옮깁니다.

Additional note: `statusline` is a root-registered command but `Hidden: true` in `moai --help`; `help` and `completion` are cobra-generated. The `run` verb exists ONLY as a `moai migration` subcommand (`internal/cli/migration.go`) — there is NO standalone `moai run` or `moai sync` root command; the plan/run/sync workflow lives in the `/moai` Claude Code skill set (§4.2).

**Source:** `./bin/moai --help` rendered output (2026-09-08, built from this tree at HEAD `52f863f36`) + `grep -rn '\.AddCommand(' internal/cli/ --include='*.go' | grep -v _test | wc -l` (**207** non-test calls) + `grep -rn 'rootCmd\.AddCommand(' internal/cli --include='*.go' | grep -v _test | wc -l` (**62** root registrations across the package; **29**의 `rootCmd.AddCommand`가 `internal/cli/root.go:143-269`의 `init()` 안에 있다) + `find internal/cli -name '*.go' ! -name '*_test.go' | wc -l` (**279**).

The `codex` launcher: closed-set verb routing `{bare, cli, app}` (launch, `--spawn` optional, `-w <worktree>` optional, `--` passthrough) × `{status}` (readout, rc 0, starts nothing); an unknown token is rejected with a one-line usage diagnostic (rc 1), never routed to a launch. Downstream of routing, an argv-translation table forwards a verb to the child only where it names a real codex subcommand — `app` is forwarded, the bare form and `cli` are moai-side synonyms and are not, so the child receives only the operator's own tail. `-w` is consumed by moai (it points the child's working directory at an EXISTING worktree and never creates one) and is not forwarded. All three launching forms pass through ONE init-offer gate function immediately before launching — the gate takes no `--spawn` parameter, accepts exactly `y`/`yes` at its prompt, exits 130 on decline (cancel, not error) and 1 on failure or a non-interactive session (report only, no prompt issued), and on acceptance delegates to the `moai init --agent codex` wiring generator exactly once, then links `AGENTS.md` ↔ `CLAUDE.md` (connection-only: at most one appended `@AGENTS.md` / `@CLAUDE.local.md` directive per file, path-containment guard runs before any read or write, writes are per-file temp+rename, idempotent on re-run).

**Source (gate):** `internal/cli/codex_launcher.go` (verb routing, single gate call site in `runCodexLaunch`), `internal/cli/codex_init.go` (gate + seams), `internal/cli/codex_contract.go` (link contract). Verified 2026-08-28 by direct read; the routing and argv-translation sentences re-verified 2026-09-01 by direct read of `codexVerbRouting` / `codexChildSubcommand` on the tree that reversed the default.

### §4.2 `/moai` Claude Code skill set (16 commands)

The complete `/moai` slash-command set in `.claude/commands/moai/`:

`clean` · `codemaps` · `e2e` · `feedback` · `fix` · `gate` · `goal` · `harness` · `loop` · `mx` · `plan` · `project` · `review` · `run` · `sync` · `todo`

(16 files total)

**Source:** `find .claude/commands/moai -maxdepth 1 -name '*.md' | wc -l` → 16. Verified 2026-09-08 (HEAD `52f863f36`): the listing returns exactly the 16 names above and nothing else (`brain`/`coverage`/`design` are NOT present as standalone command files).

이 16개 소스는 codex 쪽으로도 발행됩니다 — `internal/template/commandemit`이 각각을
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

**Source:** `internal/config/defaults.go` — `DefaultGLMBaseURL` line 124, the `DefaultGLM53Flash`/`DefaultGLM53` values lines 157-158, the tier mapping lines 159-162 and 181-183 (DefaultGLM* constants block). Verified 2026-09-08 (HEAD `52f863f36`): `grep -nE 'DefaultGLM(High|Medium|Low|Fable|Opus|Sonnet|Haiku|BaseURL) ' internal/config/defaults.go` → `DefaultGLMHigh`(159) · `Medium`(160) · `Low`(161) · `Haiku`(181) · `Sonnet`(182) · `Opus`(183) 여섯이 `DefaultGLM53Flash`로, `DefaultGLMFable`(162)만 `DefaultGLM53`으로 해석된다. Base URL SSOT: `DefaultGLMBaseURL = "https://api.z.ai/api/anthropic"` (line 124). Launcher suffix expansion: `expandModelString` (launcher.go:1112) / `splitModelSuffix`.

---

## How to use this checklist

1. **Before rewriting any docs fact** (agent count, status value, frontmatter field, CLI verb, GLM model), re-verify it against the cited source file in THIS commit's tree.
2. **If a fact has drifted** from its source, open a SEPARATE commit updating this file FIRST, then proceed with the rewrite.
3. **Do NOT cite this file as ground truth** in user-facing docs — cite the primary source (`internal/spec/status.go`, etc.). This file is a navigation aid that points AT the sources.
4. **Neutrality**: this file lives under `.moai/project/codemaps/` which is NOT covered by the template-neutrality CI guard. Keep it free of internal SPEC-ID / REQ / Audit-citation leakage except for the self-reference to `SPEC-V3R6-DOCS-CODEMAPS-V3-001` (this file's authoring SPEC).
