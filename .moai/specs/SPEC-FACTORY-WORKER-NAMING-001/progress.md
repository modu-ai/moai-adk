# progress.md — SPEC-FACTORY-WORKER-NAMING-001 (card t1085)

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-22
- plan_audit: iter-1 PASS 0.9375 (commit 5a96704e2) → D1/D2/D6 fix (commit 08f22109c) → iter-2 confirm PASS
- artifacts: spec.md, plan.md, acceptance.md, research.md (card-mandated Tier M + research addendum), this skeleton
- baseline: HEAD 3f3ffbb57, branch WT-worker-rename

## §E.2 Run-phase Evidence

### M1 — GTD todo naming-axis closure record (2026-09-22, HEAD cb74ebd58 tree)

- AC-001 PASS — the closure note `gtd-todo-naming-closure.md` (committed 08f22109c) was read this run; all four required elements present: (1) operator decision dated 2026-09-22, `moai todo` keeps its name (lines 3-4); (2) 8-tree GTD family closed as investigation-only (line 8); (3) t855 zero-work-commits note (line 9); (4) t1084 named as disposal owner (line 10). No SPEC status transitions performed. Baseline: this run, this tree, HEAD cb74ebd58.
- AC-008 (post-M1 leg) PASS — `$ go test ./internal/cli/ -run TestGTD -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 4.934s` (this run, this tree, HEAD cb74ebd58).
- M2 gate re-measurement at M1 close: `git ls-tree develop --name-only internal/factorymsg/ | wc -l` → `0` (factorymsg NOT in develop); t1074 worktree files unchanged (mtime 17:17). Gate CLOSED — M2-M4 halt per REQ-003; resume on gate re-measurement after t1074 lands.

## §F Phase 4 Mode Selection

- Input parameters: tier M; scope ~10 files (factory.go, factory_slots.go, i18n, tests, doc twins); domains 3 (go source, hook i18n, rule-doc twins + templates); file mix go+md; concurrency benefit LOW (coding-heavy rename).
- Mode evaluation: direct — no (multi-file code change); fanout — no (coding-heavy, Anthropic caveat); sweep — no (semantic rename across coupled surfaces, not uniform-mechanical ≥30 files); agent-team — not requested.
- Decision: **serial** (one manager-develop spawn per milestone batch: M2-M3 inventory+rename, then M4 decision).
- Justification: coding-heavy work on interdependent surfaces (token, help text, i18n 4-locale lockstep, tests, doc twins) — sequential edits in one tree by one writer; the t1074 gate already serializes the whole rename. Boundary case: none.

### M2 gate re-measurement (2026-09-22, post-18:18 — this session, branch WT-worker-rename, HEAD eace7849f)

Both gate observations re-measured fresh from this session against the develop ref (attributed baseline: `git rev-parse --short develop` → **a08972a28**, measured 2026-09-22 after 18:18):

1. develop SHA (the tree the gate read):
   - `$ git rev-parse --short develop` → `a08972a28`
2. factorymsg presence in develop:
   - `$ git ls-tree develop --name-only internal/factorymsg/ | wc -l` → `0`
3. t1074 SPEC status read from develop:
   - `$ git show develop:.moai/specs/SPEC-FACTORY-MIXED-HOOK-001/spec.md` → `fatal: path '.moai/specs/SPEC-FACTORY-MIXED-HOOK-001/spec.md' does not exist in 'develop'` (exit 128 — path absent from the develop tree)

**Conclusion**: M2 gate remains **CLOSED** (factorymsg absent from develop AND SPEC-FACTORY-MIXED-HOOK-001 SPEC absent from develop). M2-M4 **halt per REQ-003**. Resume condition: t1074 (SPEC-FACTORY-MIXED-HOOK-001) lands in the local `develop` ref — observable as `git ls-tree develop --name-only internal/factorymsg/` returning non-empty AND `git show develop:.moai/specs/SPEC-FACTORY-MIXED-HOOK-001/spec.md` succeeding with frontmatter `status` ∈ {implemented, completed}.

### M2 — gate OPEN + old-token inventory (2026-09-23, branch WT-worker-rename, HEAD 861510fb6)

Kickoff: operator Implementation Kickoff Approval relayed by the lead 2026-09-23. All commands below ran in this session against this tree; outputs verbatim.

**Step 1 — gate (REQ-002 / AC-002)**

```
$ git rev-parse --short HEAD                                   → 861510fb6
$ git rev-parse --short develop                                → 5d2d2b780   (first read; later reads: 08113ff0f — develop is moving under concurrent lane merges)
$ git rev-list --count --left-right develop...HEAD             → 3	0         (first read; later 18	0)
$ git ls-tree develop --name-only internal/factorymsg/ | wc -l → 4
$ git ls-tree develop --name-only internal/factorymsg/
internal/factorymsg/launch_pending_rollback_test.go
internal/factorymsg/roster_test.go
internal/factorymsg/store.go
internal/factorymsg/store_test.go
$ git show develop:.moai/specs/SPEC-FACTORY-MIXED-HOOK-001/spec.md | grep '^status:' → status: completed
```

Gate **OPEN** (factorymsg present on develop AND SPEC-FACTORY-MIXED-HOOK-001 `completed`). The t1074 merge `861510fb6` is HEAD itself. develop drift since kickoff (`861510fb6..08113ff0f`, 18 commits: t1086/t1087/t1090/t1091) touches `internal/web/*_test.go`, `internal/mission/governance_receipt_jev_test.go`, `internal/harness/rosterguard/registry.go`, `.claude/rules/moai/workflow/worktree-integration.md`, `.claude/rules/local/*`, and other SPEC/report dirs — `git diff --name-only HEAD develop` shares **zero** paths with this SPEC's edit set. Absorption is deferred to the integration window (lane protocol), not done here.

**Step 2 — inventory (REQ-007 input)**. Raw grep outputs: `.moai/reports/t1085/run-verdict.md` § M2 inventory. Per surface class:

| # | Surface class | Files (count of `lane-` / agent-shape¹ lines) | Disposition input |
|---|---|---|---|
| 1 | Factory CLI production (help, usage error, parse, desugar) | cc.go 10/4 · glm.go 9/0 · factory.go 18/11 · codex_factory.go 2/5 · kanban.go 5/0 | user-facing → rename to worker; old forms parse as aliases |
| 2 | Label vocabulary (kanban) | bootstrap.go 8/7 · factory_slots.go 2/0 · record.go 2/0 · role.go 1/0 | label producer → `worker-<n>`; parser accepts legacy `lane-<n>`/`agent-<n>` |
| 3 | Hook factory surfaces + i18n | session_start_factory.go 4/0 · session_start_factory_i18n.go 8/8 (4 locales × 2 fields) · session_start_kanban_i18n.go 4/0 (`nameChoices`, 4 locales) · session_start_record.go / factory_messages.go 0/0 (consume the label via the kanban parsers) | 4-locale lockstep rename of notation |
| 4 | **factorymsg (persisted, per-run broker.db)** | store.go 0/2 — `if p.Slot == "agent"` allocates `fmt.Sprintf("agent-%d", n)` (store.go:320-322); `peers.slot` is an opaque TEXT key; production callers (hook `registerFactoryHookPeer`, cli `registerFactoryLaunchPending`) pass the already-claimed env label (`MOAI_FACTORY_WORKER`), never the bare sentinel — only tests (store_test.go:459, factory_mixed_test.go:116) reach it | **schema unchanged**; rename changes only newly-written values. Compat: allocator treats an existing `agent-<n>`/`lane-<n>`/`worker-<n>` row as taking number n; old rows stay readable (opaque strings) |
| 5 | **factory.db `workers.label` (persisted, project-scoped)** | factory_slots.go `ClaimFactoryWorkerName` / `FactoryFreeSlots`; rows hold `lane-<n>` / `agent-<n>` from pre-rename launchers, reaped by pid liveness | schema unchanged; compat: a live legacy-labelled row blocks its number for the canonical `worker-<n>` claim |
| 6 | **kanban session record (persisted JSON)** | `Record.Role` value `"lane"` (`kanban.RoleLane`, role.go:42) + `Record.Lane` int, read by web | **KEEP** — internal persisted role key, not user notation; renaming would change on-disk record format |
| 7 | web console | factory_lanes.go 2 (`SplitFactoryLaneLabel` ×2) | consumes the kanban parser; widening it to worker+legacy keeps legacy rows visible |
| 8 | Tests (six named) | factory_test.go 86 lane-/2 · codex_factory_test.go 2/2 · goal_mission_test.go 13 (`lane-10` owner label) · gtd_compat_test.go 1 (`--lane lane-10`) · goal_blocked_question_regression_test.go 1 + doctor_jev_test.go 1 (prose "lane-question routing") | rename notation; prose compounds kept |
| 9 | Tests (other factory) | session_start_factory_test.go 27 · session_start_record_test.go 7 · factory_slots_test.go 15 · factory_label_test.go 4 · web factory_lanes_test.go 12 · factory_lane_section_test.go 9 · factorymsg/*_test + hook/cli factorymsg tests (`agent-1`, `agent-2` slot fixtures) | update where they assert produced labels; legacy-input cases retained as alias coverage |
| 10 | Rule-doc twins | local kanban-dispatch.md 2 (L211 "lane-local" adj., L266 notation) · detail 3 (L32/L182 notation, L176 "lane-local" adj.); template mirrors identical (sha `ea596b32d163` / `6607f4238a55` pairwise) | rename notation lines; keep "lane-local" (prose adjective, per-line) |
| 11 | Out of this SPEC's write scope (public/user docs, agents) | docs-site 4 locales × (factory-mode 8 + launchers 2 + kanban-mode 2) = 48 · README ×4 = 12 · CHANGELOG 2 · manager-lead.md ×2 + `.codex` toml 1 — all advertise `-f lane-<n>` / `-f agent` | sync-phase / separate card; **published docs teach the old forms → strong keep-alias evidence** |
| 12 | Unrelated homographs | integration-lock / slot-lease tests using `lane-2` as an arbitrary holder name (kanban 26+9+7+6+5+3+2, cli 6+6+5+7+4…), `agent-memory`, `--agent-name`, `agent-model` | not factory labels — untouched |
| 13 | Operator-side scripts outside the repo | not measurable from this tree | informs keep-alias; out of write reach |

¹ agent-shape = lines matching `-f agent` · `agent-<` · `agent-%d` · `agent lane` · `factoryAgentRole` · `Slot == "agent"`.

`-f agent` repo-wide: 45 lines (8 i18n, 6 factory.go, 3 cc.go, 2 codex_factory.go, 1 bootstrap.go, tests 7, SPEC dirs 18). Doc/README/CHANGELOG/agent sweep (`-f lane` · `--name lane` · `lane-<` · `-f agent`): 65 lines across 20 files.

Persisted-format verdict (constraint 2): **no blocker** — neither SQLite schema nor the record JSON shape changes; only newly-written label values change, and every reader of those values gains a legacy-accepting parse (compat read path + tests in M3).

## §E.3 Run-phase Audit-Ready Signal

- run_status: audit-ready (M1 scope)
- run_complete_at: 2026-09-22
- run_commit_sha: eace7849f (M1 evidence baseline; this close commit is docs-only on top — see below)
- scope note: M2-M4 are gated behind card t1074 landing in local develop (REQ-003); this signal covers the M1 scope ONLY. The SPEC stays `in-progress` until the gate opens — sync does NOT close it, §E.4 remains pending.
- new_warnings_or_lints_introduced: 0 (docs-only changes; no code touched)
- m1_to_mN_commit_strategy: M1 evidence commit (eace7849f) + one M1-close docs commit (this); M2-M4 commits deferred to gate open.

### AC status matrix (M1 scope)

| AC | Status | Verification | Actual output (verbatim) |
|----|--------|--------------|--------------------------|
| AC-001 | PASS | Read `gtd-todo-naming-closure.md` (this run, this tree, HEAD eace7849f) | All four required elements confirmed present: (1) "**Date of decision**: 2026-09-22 (operator decision, relayed via card t1085 dispatch)" + "**Decision**: the `moai todo` command KEEPS its name. It is NOT renamed to a GTD-branded name."; (2) "The GTD card family — t855 (naming), t867 (canon body), t899 (landed store), t939, t940, t941, autonomy ×2 (8 trees total) — is closed as **investigation-only scope**."; (3) "**t855 has zero work commits** (the naming card was never executed), so a closure record suffices for it"; (4) "the physical DISPOSAL of the 8 old GTD work trees is **card t1084's** concern." |
| AC-002 | PASS | Gate observations recorded with command + verbatim output | See the new §E.2 entry "M2 gate re-measurement (2026-09-22, post-18:18)" above — both observations (factorymsg presence = 0; t1074 SPEC path absent in develop) carry exact commands + verbatim outputs, attributed to develop SHA a08972a28. |
| AC-003 | PASS (gate-fail branch) | Blocker-path exercise per §D.1: "When the gate fails, Then a blocker report is returned and zero rename edits exist in the tree" | Gate FAILED (CLOSED) → halt per REQ-003 recorded in §E.2 above (the halt record stands in for the blocker return to the orchestrator, per this close delegation's §A scope). Zero rename edits: `$ git status --porcelain` → empty (pre-edit measurement, this run, HEAD eace7849f); no `-f agent`/`lane-N`/`worker-N` rename edit exists anywhere in the tree — M3 never started. |
| AC-008 | PASS (M1 leg) | `$ go test ./internal/cli/ -run TestGTD -count=1` | `ok  	github.com/modu-ai/moai-adk/internal/cli	4.849s` (this run, this tree, HEAD eace7849f — pre-commit measurement; the close commit on top is docs-only touching `.moai/specs/*.md`, so the `internal/cli` tree it verifies is byte-identical at the final HEAD). The gtd↔todo alias surface is unchanged by the closure record. |
| AC-004 | BLOCKED | M3 (vocabulary rename) — gated | M2-M4 halt per REQ-003; resume condition = §E.2 entry above (t1074 lands in develop). |
| AC-005 | BLOCKED | M3 (4-locale i18n parity) — gated | Same resume condition. |
| AC-006 | BLOCKED | M3 (doc twins + embed check) — gated | Same resume condition. |
| AC-007 | BLOCKED | M4 (alias decision record) — gated | Same resume condition. |
| AC-009 | BLOCKED | M3 (affected-package tests post-rename) — gated | Same resume condition; AC-008's M1-leg green is NOT a substitute. |
| AC-010 | BLOCKED | M3/M4 (closing sweep) — gated | Same resume condition. |

- resume: after t1074 lands in local develop, re-measure the gate per the §E.2 resume condition; on OPEN, resume M2 (gate + inventory), then M3 (rename), then M4 (decision + verification).

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
