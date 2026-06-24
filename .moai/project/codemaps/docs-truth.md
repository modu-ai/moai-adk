# docs-truth.md — Canonical Facts Checklist for the Docs-v3 Cohort

> Navigation aid, NOT a new SSOT. Every fact below is cross-referenced to
> its ground-truth source file. Later cohort SPECs (README / DOCSITE /
> COVERAGE / i18n) re-verify each fact against the cited source, NOT
> against this file.
> If a fact here has drifted from its source, update this file FIRST
> (separate commit) before proceeding with the rewrite.

---

## §1. Agent Catalog (8 retained)

The MoAI agent catalog consists of exactly **8 retained agents** (7 MoAI-custom + 1 Anthropic built-in `Explore`).

| # | Agent | Class | Phase scope |
|---|-------|-------|-------------|
| 1 | `manager-spec` | core/manager | Plan-phase artifact authoring (spec/plan/acceptance/research/design) |
| 2 | `manager-develop` | core/manager | Run-phase implementation (cycle_type ∈ {ddd, tdd, autofix}) |
| 3 | `manager-docs` | core/manager | Sync-phase documentation (CHANGELOG, README, frontmatter transitions) |
| 4 | `manager-git` | core/manager | PR creation per Tier-based routing + Late-Branch closure |
| 5 | `plan-auditor` | meta/evaluator | Independent plan-phase audit, bias prevention, GEARS compliance |
| 6 | `sync-auditor` | meta/evaluator | Independent skeptical quality assessment, 4-dimension scoring |
| 7 | `builder-harness` | builder | Dynamic project-specific harness specialist generation |
| 8 | `Explore` | Anthropic built-in | Read-only codebase exploration (no MoAI file — invoked directly) |

Class breakdown: Manager ×4 (`manager-spec`, `manager-develop`, `manager-docs`, `manager-git`) · Evaluator ×2 (`plan-auditor`, `sync-auditor`) · Builder ×1 (`builder-harness`) · Anthropic built-in ×1 (`Explore`).

**Archived agents**: 12 legacy agent names are archived and MUST NOT be spawned. The full archived-name list + per-archived-agent migration table lives in `.claude/rules/moai/workflow/archived-agent-rejection.md` (consult that file rather than naming the archived agents here, to keep this checklist free of archived-name leakage).

**Source:** `ls .claude/agents/moai/*.md` (= 7 MoAI-custom files) + CLAUDE.md §4 Retained Agents table + `.claude/rules/moai/workflow/archived-agent-rejection.md` (archived-agent migration table). Verified 2026-06-17: `ls -1 .claude/agents/moai/*.md | wc -l` → 7.

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

**Source:** `internal/spec/status.go` `ValidStatuses` slice (lines 13-22). Verified 2026-06-17: `grep -cE '"draft"|"planned"|"in-progress"|"implemented"|"completed"|"superseded"|"archived"|"rejected"' internal/spec/status.go` → 8. Schema SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Enum.

---

## §3. SPEC Frontmatter Schema (12 required fields)

Every `spec.md` MUST contain exactly these 12 fields in YAML frontmatter:

`id` · `title` · `version` · `status` · `created` · `updated` · `author` · `priority` · `phase` · `module` · `lifecycle` · `tags`

Rejected snake_case aliases (silently dropped by the YAML decoder): `created_at:` → use `created:`; `updated_at:` → use `updated:`; `labels:` → use `tags:`; `spec_id:` → use `id:`.

**Source:** `internal/spec/lint.go` `FrontmatterSchemaRule.Check()` required slice (~lines 586-602, 12 entries of the form `{"<field>", fm.<Field>}`). Verified 2026-06-17: the required slice carries exactly 12 entries. Schema SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md` § Canonical 12 Required Fields. Lint rule code: `FrontmatterInvalid` (Warning severity).

---

## §4. CLI Subcommand Surface

### §4.1 `moai` terminal verbs (human-facing)

Top-level verbs rendered by `moai --help`, grouped by capability:

| Capability group | Verbs |
|------------------|-------|
| Project | `init`, `doctor`, `status`, `update`, `version` |
| Launchers | `cc`, `cg`, `glm`, `web`, `statusline` |
| Autonomous Development | `loop`, `spec`, `worktree`, `brain` |
| Governance | `constitution`, `mx`, `telemetry` |

Additional verbs registered across the `internal/cli/` tree (40 files, 119 `rootCmd.AddCommand` calls): `hook`, `agent`, `research`, `workflow`, `migrate`, `profile`, `lsp`, `github`, `clean`, `feedback`, `review`, `coverage`, `e2e`, `codemaps`, `design`, `project`, `plan`, `run`, `sync`, `harness`, `session`, `fix`, `gate`, `mx`.

The checklist lists human-facing verb names (e.g. `init`, `update`, `glm`, `cc`, `cg`, `web`, `session`, `spec`, `harness`, `worktree`, `hook`, `agent`, `research`, `workflow`), NOT internal Go identifiers like `worktree.WorktreeCmd`.

**Source:** `moai --help` rendered output + `grep -rnE '\.AddCommand\(' internal/cli/` (spans 40 files, 119 non-test calls). Verified 2026-06-17.

### §4.2 `/moai` Claude Code skill set (17 commands)

The complete `/moai` slash-command set in `.claude/commands/moai/`:

`brain` · `clean` · `codemaps` · `coverage` · `design` · `e2e` · `feedback` · `fix` · `gate` · `harness` · `loop` · `mx` · `plan` · `project` · `review` · `run` · `sync`

(17 files total)

**Source:** `ls -1 .claude/commands/moai/*.md | wc -l` → 17. Verified 2026-06-17: `ls -1 .claude/commands/moai/*.md | xargs -n1 basename | sed 's/\.md$//'` yields the 17 names above.

---

## §5. GLM → Claude Tier Mapping (full tier-models table)

The GLM→Claude-tier model mapping reflects the glm-5.2[1m] activation. The full tier-models table:

| Constant | Value | Claude tier |
|----------|-------|-------------|
| `DefaultGLMBaseURL` | `https://api.z.ai/api/anthropic` | (z.ai gateway) |
| `DefaultGLMHigh` | `glm-5.2[1m]` | High (1M context) |
| `DefaultGLMMedium` | `glm-4.7` | Medium |
| `DefaultGLMLow` | `glm-4.5-air` | Low |
| `DefaultGLMSonnet` | `glm-4.7` | Sonnet |
| `DefaultGLMHaiku` | `glm-4.5-air` | Haiku |
| `DefaultGLMOpus` | `glm-5.2[1m]` | Opus (1M context) |

The `[1m]` suffix on `glm-5.2[1m]` activates Claude Code's 1M context mode (the suffix is parsed and stripped by Claude Code before the upstream API call; z.ai never sees it).

Additional GLM models available but not default-mapped: `glm-4.5`, `glm-4.6`, `glm-5.1`, `glm-5-turbo`.

**Source:** `internal/config/defaults.go` lines 40-57 (DefaultGLM* constants block). Verified 2026-06-17: `grep -q 'DefaultGLMHigh.*=.*"glm-5\.2\[1m\]"' internal/config/defaults.go` → exit 0. Base URL SSOT: `DefaultGLMBaseURL = "https://api.z.ai/api/anthropic"`.

---

## How to use this checklist

1. **Before rewriting any docs fact** (agent count, status value, frontmatter field, CLI verb, GLM model), re-verify it against the cited source file in THIS commit's tree.
2. **If a fact has drifted** from its source, open a SEPARATE commit updating this file FIRST, then proceed with the rewrite.
3. **Do NOT cite this file as ground truth** in user-facing docs — cite the primary source (`internal/spec/status.go`, etc.). This file is a navigation aid that points AT the sources.
4. **Neutrality**: this file lives under `.moai/project/codemaps/` which is NOT covered by the template-neutrality CI guard. Keep it free of internal SPEC-ID / REQ / Audit-citation leakage except for the self-reference to `SPEC-V3R6-DOCS-CODEMAPS-V3-001` (this file's authoring SPEC).
