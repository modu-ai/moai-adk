# docs-truth.md — Canonical Facts Checklist for the Docs-v3 Cohort

> Navigation aid, NOT a new SSOT. Every fact below is cross-referenced to
> its ground-truth source file. Later cohort SPECs (README / DOCSITE /
> COVERAGE / i18n) re-verify each fact against the cited source, NOT
> against this file.
> If a fact here has drifted from its source, update this file FIRST
> (separate commit) before proceeding with the rewrite.

---

## §1. Agent Catalog (11 retained)

The MoAI agent catalog consists of exactly **11 retained agents** (10 MoAI-custom + 1 Anthropic built-in `Explore`).

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
| 11 | `Explore` | Anthropic built-in | Read-only codebase exploration (no MoAI file — invoked directly) |

Class breakdown: Manager ×5 (`manager-spec`, `manager-develop`, `manager-docs`, `manager-git`, `manager-design`) · Evaluator ×2 (`plan-auditor`, `sync-auditor`) · Builder ×1 (`builder-harness`) · Advisor ×1 (`super-advisor`) · Specialist ×1 (`e2e-tester`) · Anthropic built-in ×1 (`Explore`).

**Archived agents**: 12 legacy agent names are archived and MUST NOT be spawned. The full archived-name list + per-archived-agent migration table lives in `.claude/rules/moai/workflow/archived-agent-rejection.md` (consult that file rather than naming the archived agents here, to keep this checklist free of archived-name leakage).

**Source:** `ls .claude/agents/moai/*.md` (= 10 MoAI-custom files) + CLAUDE.md §4 Retained Agents table + `.claude/rules/moai/workflow/archived-agent-rejection.md` (archived-agent migration table). Verified 2026-07-29: `ls -1 .claude/agents/moai/*.md | wc -l` → 10 (manager-spec, manager-develop, manager-docs, manager-git, plan-auditor, sync-auditor, builder-harness, super-advisor, manager-design, e2e-tester).

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
| Autonomous Development | `loop`, `spec`, `worktree`, `goal` |
| Governance | `constitution`, `mx`, `telemetry` |

Additional verbs registered across the `internal/cli/` tree (109 non-test `.go` files, 152 non-test `.AddCommand(` calls): `hook`, `agent`, `research`, `workflow`, `migrate`, `profile`, `lsp`, `github`, `clean`, `feedback`, `review`, `coverage`, `e2e`, `codemaps`, `design`, `project`, `plan`, `run`, `sync`, `harness`, `session`, `fix`, `gate`, `mx`.

The checklist lists human-facing verb names (e.g. `init`, `update`, `glm`, `cc`, `cg`, `web`, `session`, `spec`, `harness`, `worktree`, `hook`, `agent`, `research`, `workflow`), NOT internal Go identifiers like `worktree.WorktreeCmd`.

**Source:** `moai --help` rendered output + `grep -rn '\.AddCommand(' internal/cli/ --include='*.go' | grep -v _test` (109 non-test files, 152 non-test calls). Verified 2026-07-29.

### §4.2 `/moai` Claude Code skill set (15 commands)

The complete `/moai` slash-command set in `.claude/commands/moai/`:

`clean` · `codemaps` · `e2e` · `feedback` · `fix` · `gate` · `goal` · `harness` · `loop` · `mx` · `plan` · `project` · `review` · `run` · `sync`

(15 files total)

**Source:** `ls -1 .claude/commands/moai/*.md | wc -l` → 15. Verified 2026-07-29: `ls -1 .claude/commands/moai/*.md | xargs -n1 basename | sed 's/\.md$//'` yields the 15 names above (note: `goal` added since v3.0; `brain`/`coverage`/`design` are NOT present as standalone command files).

---

## §5. GLM → Claude Tier Mapping (full tier-models table)

The GLM→Claude-tier model mapping reflects the glm-5.2[1m] activation. The full tier-models table:

| Constant | Value | Claude tier |
|----------|-------|-------------|
| `DefaultGLMBaseURL` | `https://api.z.ai/api/anthropic` | (z.ai gateway) |
| `DefaultGLMHigh` | `glm-5.2` | High |
| `DefaultGLMMedium` | `glm-4.7` | Medium |
| `DefaultGLMLow` | `glm-4.5-air` | Low |
| `DefaultGLMSonnet` | `glm-4.7` | Sonnet |
| `DefaultGLMHaiku` | `glm-4.5-air` | Haiku |
| `DefaultGLMOpus` | `glm-5.2` | Opus |

The `[1m]` 1M-context suffix is NOT hardcoded in `defaults.go` — `DefaultGLMHigh`/`DefaultGLMOpus` are the bare value `glm-5.2`. The `[1m]` suffix is expanded at the launcher layer (`internal/cli/launcher.go` — `expandModelString` / `splitModelSuffix`) only when the 1M-context variant is explicitly requested; Claude Code parses and strips the suffix before the upstream API call because z.ai rejects a verbatim `[1m]`.

Additional GLM models available but not default-mapped: `glm-4.5`, `glm-4.6`, `glm-5.1`, `glm-5-turbo`.

**Source:** `internal/config/defaults.go` lines 59-79 (DefaultGLM* constants block). Verified 2026-07-29: `grep -E 'DefaultGLM(High|Medium|Low|Fable|Opus)' internal/config/defaults.go` → `DefaultGLMHigh = "glm-5.2"`, `DefaultGLMMedium = "glm-4.7"`, `DefaultGLMLow = "glm-4.5-air"`, `DefaultGLMFable = "glm-5.2"`, `DefaultGLMOpus = "glm-5.2"` (NO `[1m]` suffix on any default). Base URL SSOT: `DefaultGLMBaseURL = "https://api.z.ai/api/anthropic"`.

---

## How to use this checklist

1. **Before rewriting any docs fact** (agent count, status value, frontmatter field, CLI verb, GLM model), re-verify it against the cited source file in THIS commit's tree.
2. **If a fact has drifted** from its source, open a SEPARATE commit updating this file FIRST, then proceed with the rewrite.
3. **Do NOT cite this file as ground truth** in user-facing docs — cite the primary source (`internal/spec/status.go`, etc.). This file is a navigation aid that points AT the sources.
4. **Neutrality**: this file lives under `.moai/project/codemaps/` which is NOT covered by the template-neutrality CI guard. Keep it free of internal SPEC-ID / REQ / Audit-citation leakage except for the self-reference to `SPEC-V3R6-DOCS-CODEMAPS-V3-001` (this file's authoring SPEC).
