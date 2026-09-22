# research.md — SPEC-FACTORY-WORKER-NAMING-001 (card t1085)

Card-mandated plan-phase research artifact (Tier M set + research.md per card dispatch). Baseline attribution: all measurements below were taken THIS session in the worktree `.claude/worktrees/t1085`, HEAD `3f3ffbb57`, branch `WT-worker-rename` — the same commit as develop head at plan time.

## §1 t1074 serialization evidence (measured)

- SPEC-FACTORY-MIXED-HOOK-001 (card t1074): status `in-progress`, run_status fail (11/15 AC pass), actively repaired.
- Its worktree carries uncommitted `internal/factorymsg/{store.go (29KB), store_test.go, roster_test.go}` using `agent` / `agent-%d` slot vocabulary.
- Its copy of `internal/cli/factory.go` is 20,142 bytes vs develop's 18,473 — it has already modified the rename-target file.
- Consequence: items ② ③ serialize behind its develop landing. Gate (REQ-002): `internal/factorymsg/` exists in develop tree AND its SPEC frontmatter status read from develop ∈ {implemented, completed}.

## §2 Vocabulary site inventory (develop, 3f3ffbb57)

- `internal/cli/factory.go` — 18 `lane-` hits. Key sites: `factoryAgentRoleToken = "agent"` (line 55, with comment block 52–56); help text lines 60–84 (including the removed numeric-lane-count note at line 60 and the `-f lane-2` example at 65); join-label doc lines 148–156 (`-f lane-<n>` and the shared `-k N --name lane-<i>` implementation note); error string line 191; role/label parsing comments 195–232 (`parseFactoryLaneLabel` handles `lane-<n>` OR `agent-<n>`); worker-mode comment 285–296; `resolveFactoryWorkerName` (line 355) → `kanban.ClaimFactoryWorkerName` (`internal/kanban/factory_slots.go:117`).
- `internal/hook/session_start_factory_i18n.go` — 4-locale status strings: en 62/66, ko 91/95, ja 117/121, zh 143/147; each locale carries a "Lanes are named lane-1..lane-%d … `-f agent` … agent-<n>" string plus a `-f lane-<n>` numbering-hint string.
- Tests referencing lane- vocabulary (grep -l confirmed): `internal/cli/factory_test.go` (incl. `-f lane-2` parse cases at ~line 159, `-k 4 --name lane-2` cases at 68–74), `codex_factory_test.go`, `goal_mission_test.go`, `goal_blocked_question_regression_test.go`, `doctor_jev_test.go`, `gtd_compat_test.go`.
- Doc twins: `.claude/rules/moai/workflow/kanban-dispatch.md` (2 hits: §"Verification load is lane-local" heading 211; Factory Mode paragraph 266 `lane-1..lane-N`) and `kanban-dispatch-detail.md` (3 hits: 32, 176, 182) — plus template mirrors under `internal/template/templates/.claude/rules/moai/workflow/` with identical counts (2 + 3). Pairs are currently byte-identical in size but the C1/C2 divergence is ALLOWED — judge per file, never blind-copy.
- Default worker-name construction: `kanban.ClaimFactoryWorkerName` — exact format verified at run time on the landed tree (plan note: "verify exact naming format during run" per card).

## §3 GTD naming-axis state (measured)

- `SPEC-GTD-AUTONOMY-001` — status completed (v0.2.0, 2026-09-15, P0). `SPEC-GTD-CANON-BODY-001` — status completed (v0.3.2, card t867): moved the workflows body into `workflows/gtd.md` and retired the todo body, with compat aliasing.
- Current tree: `.claude/skills/moai/workflows/` contains `gtd.md` (29,751 bytes) and NO `todo.md` — the todo name survives via compat aliasing guarded by `internal/cli/gtd_compat_test.go` (three tests: `TestGTDAllTodoVerbsParity`, `TestGTDFiveStageCLIUsesSameSQLite`, `TestGTDCLIInputRefusalsDoNotFallThrough`; the test rewrites `moai gtd`→`moai queue` / `moai todo`→`moai queue` in help output to hold the alias surface invariant).
- Operator decision 2026-09-22: `moai todo` keeps its name — no GTD-branded rename. The closure record (M1) captures this; the alias surface and its tests are untouched (REQ-009). GTD SPEC status transitions are NOT performed here — disposition of the 8 trees (t855, t867, t899, t939, t940, t941, autonomy ×2) is t1084's.

## §4 Design considerations for the rename

- Join-token axis: the `-f` token accepts a role (`agent`) or a label (`lane-<n>`). The worker axis renames role → `worker` and label prefix → `worker-<n>`. The `--name` join-label form (`-k N --name lane-<i>` / `-f N --name lane-<i>`) shares the label shape — it renames with the same sweep, and is the highest-compat-risk form (operator scripts may pin `--name lane-2`).
- Alias options for M4 (decided AFTER M2 inventory): (a) keep-alias — old forms parse + emit a deprecation hint; (b) hard remove. The inventory must sweep operator-side surfaces (scripts/, `.claude/`, docs, hooks) and distinguish repo-internal from operator-owned usage — operator-owned usage cannot be silently broken.
- i18n: all four locales carry two strings each; lockstep edit + per-locale grep is the parity check. Locale text is user-facing → written as natural native prose per locale (ko/ja/zh wording re-authored, not mechanically substituted).
- Template mirror: `//go:embed all:templates` — every mirror edit requires `make build`; embed freshness checked via `make embed-check` awareness. Template neutrality rules (CLAUDE.local.md §15/§25) apply to the renamed doc text.

## §5 Open items

- None requiring operator clarification — the operator decisions (todo keeps its name; no a-priori removal; twin handling; t1084 boundary; t1074 serialization) are all fixed in the card dispatch. Residual judgment calls (keep-alias vs remove; prose-word "lane" per-line dispositions) are explicitly POST-measurement run-phase decisions per REQ-007/008 and are not [NEEDS CLARIFICATION] items.
