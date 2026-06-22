# Research — SPEC-WEB-CONSOLE-010

> Tier L research artifact. Records the established field inventory (ground-truth, verified by file reads), the import-dependency constraint rationale, the current divergence, and the Cleanup Candidates appendix (LIST ONLY — no deletion this SPEC).

## §A Established field inventory (verified)

Three read-only audits + direct file reads (this plan-phase) established the following. Each fact is attributable to a named file.

### A.1 ProfilePreferences — 11 fields, ALL USED

Defined in `internal/profile/preferences.go` (struct `ProfilePreferences`, lines 14-51). The 11 USED fields:

| Field | YAML key | Consumer |
|-------|----------|----------|
| `UserName` | `user_name` | `profile/sync.go` → `user.yaml` (`SyncToProjectConfig` lines 27-33) |
| `ConversationLang` | `conversation_lang` | `profile/sync.go` → `language.yaml` (lines 39-43) |
| `GitCommitLang` | `git_commit_lang` | `profile/sync.go` → `language.yaml` (line 44) |
| `CodeCommentLang` | `code_comment_lang` | `profile/sync.go` → `language.yaml` (line 48) |
| `DocLang` | `doc_lang` | `profile/sync.go` → `language.yaml` (line 52) |
| `ModelPolicy` | `model_policy` | launcher (model-policy env) |
| `Model` | `model` | launcher → `DO_CLAUDE_MODEL` env |
| `EffortLevel` | `effort_level` | launcher → `CLAUDE_CODE_EFFORT_LEVEL` env |
| `PermissionMode` | `permission_mode` | launcher → `settings.local.json` + `--permission-mode` flag |
| `StatuslineSegments` | `statusline_segments` | `profile/sync.go` → `statusline.yaml` (`syncStatusline`) |
| `StatuslineTheme` | `statusline_theme` | `profile/sync.go` → `statusline.yaml` |

(`Bypass` is deprecated/migrated to `PermissionMode`; not a USED UI field.)

### A.2 The canonical 6 sections (USED union, 34 fields)

| # | Section | Fields | Persistence target |
|---|---------|--------|--------------------|
| 1 | Identity | `user_name` | profile store → synced `user.yaml` |
| 2 | Language | `conversation_lang`, `git_commit_lang`, `code_comment_lang`, `doc_lang` | profile store → synced `language.yaml` |
| 3 | Launch | `model`, `model_policy`, `effort_level`, `permission_mode` | profile store → `settings.local.json` + launch env |
| 4 | Statusline | `theme` + 15 segments (claude_version, context, directory, effort_thinking, git_branch, git_status, moai_version, model, output_style, pr, session_time, task, usage_5h, usage_7d, worktree) | profile store → synced `statusline.yaml` |
| 5 | Quality | `development_mode`, `test_coverage_target`, `enforce_quality`, `tdd_settings.min_coverage_per_commit` | `.moai/config/sections/quality.yaml` |
| 6 | Git Convention | `convention`, `auto_detection.{enabled, confidence_threshold, sample_size}`, `validation.enforce_on_push` | `.moai/config/sections/git-convention.yaml` |

The 15 segment keys are sourced from `internal/cli/profile_setup.go` `statuslineAllSegments` (lines 55-60) and the `internal/statusline.Segment*` constants (`profile/sync.go:155-172`). `SegmentRepo` (16th constant) is intentionally outside the schema (`sync.go:154`).

## §B Current divergence (what unification must fix)

| Divergence | TUI state | Web state | Source |
|------------|-----------|-----------|--------|
| Statusline (16 fields) | present (theme select + 15-segment MultiSelect, unconditional) | **absent** — panel removed | `profile_setup.go:369-405`; `fieldsets.templ:120-125`; `validate.go:23-24` |
| Quality nested (3 fields) | **absent** — only `development_mode` scalar | present (`writeProjectNestedConfig`) | `profile_setup.go:407-431` vs `projectconfig.go:242-294` |
| Git nested (4 fields) | **absent** — only `convention` scalar | present | `profile_setup.go:420-431` vs `projectconfig.go:268-286` |
| Empty-option label | lang "(not set)", model "Default (no override)", effort verbose, git "(project default)" | lang "(unset)", model "(project default)", effort "(runtime default)", git "(unchanged)" | translation tables vs `fieldsets.templ` / `permissionOption` (`:104-118`) |
| permission_mode normalization | `acceptEdits → ""` (line 443) | renders empty as "(project default)" | `profile_setup.go:443` |
| Canonical option lists | unexported in `profile_setup.go` | hand-mirrored in `validate.go` (`modelCanonical:35`, `effortLevelCanonical:40`) | `validate.go:17-40` |

## §C Import-dependency constraint rationale (the architectural crux)

Verbatim from `internal/web/validate.go:17-22` (the documented cause):

> "이 목록들은 internal/cli/profile_setup.go 의 wizard SSOT … 와 동일한 정규 값이다. wizard의 정규 리스트가 미노출(unexported, internal/cli 패키지 전용)이고 **internal/cli → internal/web 단방향 의존이므로 역참조가 불가능하여** 같은 값을 여기서 재선언한다 … **wizard 정규 리스트 변경 시 본 목록도 함께 갱신해야 한다.**"

Translation of the load-bearing clause: because `internal/cli → internal/web` is the only legal import direction and the wizard's canonical lists are unexported, the web layer cannot reference them and must re-declare them by hand; any change to the wizard lists must be hand-synced here.

**Conclusion:** the SSOT cannot live in either `internal/cli` or `internal/web`. It MUST live in a THIRD neutral package that both import — `internal/settings`. Verified plan-phase: neither `internal/settings` nor `internal/settings/schema` exists today.

Both surfaces already share neutral leaf packages (`internal/profile`, `internal/config`, `pkg/models`, `internal/template`, `internal/statusline`), so `internal/settings` sits at a layer that introduces no new cycle (AC-WC10-003 enforces this).

## §D Cleanup Candidates (deferred to a separate verified SPEC)

> **CRITICAL framing (verification-claim-integrity doctrine, `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3):** "Go has no consumer" is a **HYPOTHESIS, NOT proof that config is dead.** Much MoAI config is consumed by the agent/skill/rule layer (markdown) via the MX protocol and rule-file references, NOT by Go. A text-pattern / grep inference is a hypothesis until verified with the domain's tooling.
>
> **Mandatory pre-deletion gate (binds the future cleanup SPEC):** before removing ANY candidate below, the cleanup SPEC MUST run a full grep over `.claude/{skills,rules,agents,workflows}` (NOT just Go) AND confirm zero references. Absence of a Go consumer is necessary but NOT sufficient. This SPEC deletes nothing.

### Tier A — the 6 UI sections (KEEP, unify)

The 34 fields across the 6 canonical sections. These are kept and unified by THIS SPEC. No cleanup.

### Tier B — Go-unread but agent/skill-consumed (KEEP — do NOT delete)

These have no Go consumer but ARE consumed by the agent/skill/rule (markdown) layer. They are NOT cleanup candidates:

- `workflow.*` — `token_budget`, team `role_profiles` (consumed by team-mode orchestration rules)
- `harness.*` — `levels`, `evaluator` (consumed by harness routing rules / evaluator profiles)
- `llm.glm.*` — GLM backend routing config (consumed by `glm-web-tooling.md` rule + launcher)
- `language.yaml` directives — `code_comments` etc. consumed by agents per the MX protocol (`mx-tag-protocol.md` § Language Settings)
- `design.*` — design pipeline settings (consumed by `design/constitution.md`)
- `interview.*` — Socratic interview config (consumed by discovery rules)
- `context_search.*` — session context-search config (consumed by CLAUDE.md §16 context-search protocol)

### Tier C — true cleanup CANDIDATES (require the §D grep gate before any removal)

Hypothesized unused; each MUST pass the mandatory `.claude/{skills,rules,agents,workflows}` grep gate before the future cleanup SPEC may touch it:

- `quality.ddd_settings`
- `quality.coverage_exemptions`
- `quality.test_quality`
- `quality.lsp_quality_gates`
- `quality.coverage_threshold`
- `quality.tdd_settings.{red_green_refactor, test_first_required, mutation_testing_enabled}`
- `quality.principles`
- `git_convention.auto_detection.fallback`
- `git_convention.validation.max_length`
- loader-less orphan section files: `db.yaml`, `observability.yaml`
- `llm.{mode, team_mode, glm_env_var}`

> The Tier C list is ADVISORY ONLY. This SPEC (WEB-CONSOLE-010) does not delete, rename, or alter any of these. The candidate list exists so the future cleanup track has a verified starting inventory plus the non-negotiable grep gate. (Note: `llm.team_mode` interacts with CG mode detection per CLAUDE.local.md §22.3 — a strong example of why the grep gate is mandatory: it looks Go-orphan but is referenced by `internal/tmux/cg_detect.go` `IsCGMode`, so it is NOT safe to delete. This single counter-example justifies the entire gate.)

## §E Source references

- `internal/profile/preferences.go` — `ProfilePreferences` struct (11 USED fields).
- `internal/profile/sync.go` — `SyncToProjectConfig`, `syncStatusline`, `defaultStatuslineSegments`.
- `internal/cli/profile_setup.go` — TUI wizard (6 sections, `persistProjectConfig` 2-scalar only, `permission_mode` normalization line 443, 15-segment slice).
- `internal/cli/profile_setup_translations.go` — TUI Go translation tables (en/ko/ja/zh).
- `internal/web/fieldsets.templ` — web fieldsets; statusline panel removed (lines 120-125); `fieldsetProject` 9-field nested project section.
- `internal/web/validate.go` — hand-mirrored canonical lists + the documented import-constraint comment (lines 17-22).
- `internal/web/projectconfig.go` — `writeProjectNestedConfig` (the nested write seam, line 242), `readProjectNestedConfig`, `parseProjectNestedForm`.
- `internal/web/handlers.go` — `pageView` view-model, `newPageView` option-list assembly.
- `internal/web/assets/i18n.js` — web `window.MOAI_I18N` dictionary (4 locales).
- `.claude/rules/moai/core/verification-claim-integrity.md` — the doctrine grounding the §D grep gate.
