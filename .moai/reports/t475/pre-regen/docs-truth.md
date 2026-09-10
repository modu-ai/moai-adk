# docs-truth.md — Canonical Facts Checklist for the Docs-v3 Cohort

> Navigation aid, NOT a new SSOT. Every fact below is cross-referenced to
> its ground-truth source file. Later cohort SPECs (README / DOCSITE /
> COVERAGE / i18n) re-verify each fact against the cited source, NOT
> against this file.
> If a fact here has drifted from its source, update this file FIRST
> (separate commit) before proceeding with the rewrite.

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

**Source:** `ls -1 .claude/agents/moai/*.md` (= 11 MoAI-custom files) + CLAUDE.md §4 Retained Agents table + `.claude/rules/moai/workflow/archived-agent-rejection.md` (archived-agent migration table). Verified 2026-09-02: `find .claude/agents/moai -maxdepth 1 -name '*.md' | wc -l` → 11 (builder-harness, e2e-tester, manager-design, manager-develop, manager-docs, manager-git, manager-lead, manager-spec, plan-auditor, super-advisor, sync-auditor).

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

**Source:** `internal/spec/status.go` `ValidStatuses` slice (line 27, values at lines 28-35). Verified 2026-09-02: `grep -cE '"draft"|"planned"|"in-progress"|"implemented"|"completed"|"superseded"|"archived"|"rejected"' internal/spec/status.go` → 8. Schema SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Enum.

---

## §3. SPEC Frontmatter Schema (12 required fields)

Every `spec.md` MUST contain exactly these 12 fields in YAML frontmatter:

`id` · `title` · `version` · `status` · `created` · `updated` · `author` · `priority` · `phase` · `module` · `lifecycle` · `tags`

Rejected snake_case aliases (silently dropped by the YAML decoder): `created_at:` → use `created:`; `updated_at:` → use `updated:`; `labels:` → use `tags:`; `spec_id:` → use `id:`.

**Source:** `internal/spec/lint.go` `FrontmatterSchemaRule.Check()` required slice (lines 956-971, 12 entries of the form `{"<field>", fm.<Field>}`). Verified 2026-09-02: the required slice carries exactly 12 entries. Schema SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md` § Canonical 12 Required Fields. Lint rule code: `FrontmatterInvalid` (Warning severity).

---

## §4. CLI Subcommand Surface

### §4.1 `moai` terminal verbs (human-facing)

Top-level verbs rendered by `moai --help`, grouped by capability (2026-09-02 actual render):

| Capability group | Verbs |
|------------------|-------|
| Project | `init`, `status`, `doctor`, `update`, `migrate`, `pr` |
| Launchers | `cc`, `cg`, `glm`, `codex` |
| Autonomous/Dev | `loop`, `spec`, `plan`, `goal`, `gate` |
| Governance | `constitution`, `mx`, `telemetry` |
| Tools/Infra | `hook`, `session`, `worktree`, `migration`, `integration`, `graph`, `chain`, `handoff`, `verify`, `todo`, `epic`, `memory`, `model`, `tokens`, `clean`, `inventory`, `ast-grep`, `ast-edit`, `mcp`, `mcp-server`, `config`, `tool-policy`, `preference`, `github`, `lsp`, `research`, `agent`, `workflow`, `web`, `version` |

Additional note: `statusline` is a root-registered command but `Hidden: true` in `moai --help`; `help` and `completion` are cobra-generated. The `run` verb exists ONLY as a `moai migration` subcommand (`internal/cli/migration.go`) — there is NO standalone `moai run` or `moai sync` root command; the plan/run/sync workflow lives in the `/moai` Claude Code skill set (§4.2).

**Source:** `moai --help` rendered output (2026-09-02, built from this tree) + `grep -rn '\.AddCommand(' internal/cli/ --include='*.go' | grep -v _test` (202 non-test calls) + `grep -rn 'rootCmd\.AddCommand(' internal/cli --include='*.go' | grep -v _test | wc -l` (60 root registrations, 33 files) + `find internal/cli -name '*.go' ! -name '*_test.go' | wc -l` (264).

The `codex` launcher: closed-set verb routing `{bare, cli, app}` (launch, `--spawn` optional, `-w <worktree>` optional, `--` passthrough) × `{status}` (readout, rc 0, starts nothing); an unknown token is rejected with a one-line usage diagnostic (rc 1), never routed to a launch. Downstream of routing, an argv-translation table forwards a verb to the child only where it names a real codex subcommand — `app` is forwarded, the bare form and `cli` are moai-side synonyms and are not, so the child receives only the operator's own tail. `-w` is consumed by moai (it points the child's working directory at an EXISTING worktree and never creates one) and is not forwarded. All three launching forms pass through ONE init-offer gate function immediately before launching — the gate takes no `--spawn` parameter, accepts exactly `y`/`yes` at its prompt, exits 130 on decline (cancel, not error) and 1 on failure or a non-interactive session (report only, no prompt issued), and on acceptance delegates to the `moai init --agent codex` wiring generator exactly once, then links `AGENTS.md` ↔ `CLAUDE.md` (connection-only: at most one appended `@AGENTS.md` / `@CLAUDE.local.md` directive per file, path-containment guard runs before any read or write, writes are per-file temp+rename, idempotent on re-run).

**Source (gate):** `internal/cli/codex_launcher.go` (verb routing, single gate call site in `runCodexLaunch`), `internal/cli/codex_init.go` (gate + seams), `internal/cli/codex_contract.go` (link contract). Verified 2026-08-28 by direct read; the routing and argv-translation sentences re-verified 2026-09-01 by direct read of `codexVerbRouting` / `codexChildSubcommand` on the tree that reversed the default.

### §4.2 `/moai` Claude Code skill set (16 commands)

The complete `/moai` slash-command set in `.claude/commands/moai/`:

`clean` · `codemaps` · `e2e` · `feedback` · `fix` · `gate` · `goal` · `harness` · `loop` · `mx` · `plan` · `project` · `review` · `run` · `sync` · `todo`

(16 files total)

**Source:** `find .claude/commands/moai -maxdepth 1 -name '*.md' | wc -l` → 16. Verified 2026-09-02: the 16 names above (`todo` added since the prior census; `brain`/`coverage`/`design` are NOT present as standalone command files).

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

**Source:** `internal/config/defaults.go` — `DefaultGLMBaseURL` line 124, the `DefaultGLM53Flash`/`DefaultGLM53` values lines 157-158, the tier mapping lines 159-162 and 181-183 (DefaultGLM* constants block). Verified 2026-09-02: `grep -E 'DefaultGLM(High|Medium|Low|Fable|Opus|Sonnet|Haiku)' internal/config/defaults.go` → all seven resolve to `DefaultGLM53Flash` except `DefaultGLMFable = DefaultGLM53`. Base URL SSOT: `DefaultGLMBaseURL = "https://api.z.ai/api/anthropic"` (line 124). Launcher suffix expansion: `expandModelString` (launcher.go:1112) / `splitModelSuffix`.

---

## How to use this checklist

1. **Before rewriting any docs fact** (agent count, status value, frontmatter field, CLI verb, GLM model), re-verify it against the cited source file in THIS commit's tree.
2. **If a fact has drifted** from its source, open a SEPARATE commit updating this file FIRST, then proceed with the rewrite.
3. **Do NOT cite this file as ground truth** in user-facing docs — cite the primary source (`internal/spec/status.go`, etc.). This file is a navigation aid that points AT the sources.
4. **Neutrality**: this file lives under `.moai/project/codemaps/` which is NOT covered by the template-neutrality CI guard. Keep it free of internal SPEC-ID / REQ / Audit-citation leakage except for the self-reference to `SPEC-V3R6-DOCS-CODEMAPS-V3-001` (this file's authoring SPEC).
