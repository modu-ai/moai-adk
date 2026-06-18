# Progress — SPEC-V3R6-RULES-VERSION-FORMAT-001

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored: spec.md + plan.md + acceptance.md + progress.md (status: draft).

- SPEC ID self-check: `SPEC ✓ | V3R6 ✓ | RULES ✓ | VERSION ✓ | FORMAT ✓ | 001 ✓ → PASS` (regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`).
- 12 canonical frontmatter fields present + valid (no snake_case aliases).
- Out of Scope: 7 `### Out of Scope —` H3 sub-headings with `-` bullets (incl. CONST-RULEID-001 Go-fix exclusion).
- Milestones A-G grouped into M1-M7 (priority-ordered, no time estimates).
- Decisions recorded: language-skeleton DEFERRED (→ SPEC-V3R6-RULES-LANG-SKELETON-001); footer policy = consistent-by-absence (option b); team-pattern-cookbook time-fix OWNED here with SSOT-DEDUP sequencing dependency.
- Risks registered: R1 zone-registry clause-sync (HIGH), R5 Korean-example over-translation (HIGH), R7 mirror-vacuity (HIGH), R9 keystone validator baseline-RED (HIGH), R10 L845 non-substring, R11 ruby/Rails.
- iter-2 (plan-auditor FAIL 0.73 → fixes; all 4 defects confirmed deterministically before acting):
  - D1 (CRITICAL): `moai constitution validate` baseline-RED confirmed (CONST-V3R6-001 rejected by `ruleIDPattern`, registry load aborts). AC-VFM-001a re-scoped to `grep -Fq` substring proof; AC-VFM-001b deferred SHOULD blocked on sibling `SPEC-V3R6-RULES-CONST-RULEID-001`; plan §C false "baseline clean" claim corrected with observed RED output (acceptance.md §D.6).
  - D2: L845 confirmed non-substring (`grep -cF '1M context (Opus 4.7)' == 0`); excluded from edit via AC-VFM-001e carve-out (option i).
  - D3: ruby Rails 7.2 → 8.0 bump paired with Ruby 4.0 (AC-VFM-002c delta-grep both).
  - D4: emoji matcher aligned on `'✅\|❌'` across AC matrix + plan §F.5.

## §E.2 Run-phase Evidence

Run-phase executed by manager-develop (cycle_type=doc-only; TDD-agnostic — grep ACs + Go-test trio are the verification surface, not new behavior tests). 7 milestones (M1-M7) completed as a single scrub pass across 16 in-scope rule files (deployed) + 16 template mirrors = 32 files modified, 0 Go source changed.

**M1 — Opus model identity (AC-VFM-001a/c/d/e):**
- `core/moai-constitution.md` L56-57: "Opus 4.7" → "Opus 4.7+ / 4.8" (Principle 4/5). Atomic pair with zone-registry clauses.
- `core/zone-registry.md` L291 (CONST-V3R2-028) + L299 (CONST-V3R2-029): clauses updated to identical substring. Verified `grep -Fq` PASS (validator-independent).
- `core/zone-registry.md` L845 (CONST-V3R5-022): LEFT UNEDITED (D2 carve-out — pre-existing non-substring, source is TABLE). `grep -cF '1M context (Opus 4.7)' context-window-management.md == 0`.
- `development/agent-authoring.md` L83: effort caveat annotated "(current substrate: Opus 4.7+ / 4.8)".
- skill-authoring.md L25 + worktree-integration.md L351/355: SHOULD-scoped back-compat floors, left as-is per plan §F.1.

**M2 — Language toolchain versions (AC-VFM-002a/b/c/d/e):**
- `languages/csharp.md`: `.NET 8 / C# 12` → `.NET 10 (LTS) / C# 14` (title, features, body refs, net10.0 framework flag). Delta-grep: `.NET 10` ×3, `.NET 8` == 0.
- `languages/kotlin.md`: `Kotlin 2.0 / Ktor 3.0` → `Kotlin 2.2+ (current 2.4) / Ktor 3.x (current 3.5)` (title, features, headers, Context7 refs, L129 body). Delta-grep: `Kotlin 2.2` ×5, `Kotlin 2.0` == 0.
- `languages/ruby.md`: `Ruby 3.3+ / Rails 7.2` → `Ruby 4.0 / Rails 8.0` (paired bump — R11 consistency). Title, features, headers, body, version-check commands. Delta-grep: `Ruby 4.0` ×4, `Rails 8.0` ×3, `Rails 7.2` == 0.
- `languages/go.md` L101: Docker tag `golang:1.23-alpine` → `1.26-alpine`. `Go 1.23+` floor (L7) retained. Delta-grep: `1.26-alpine` ×1, `1.23-alpine` == 0.
- verified-correct pins UNCHANGED: python (3.13+/Django 5.2), swift (6+), php (8.3+), elixir (1.17+), rust (1.92), typescript React 19/Next.js 16.

**M3 — Footer consistency policy (AC-VFM-003):**
- `development/coding-standards.md`: new `## Footer Convention` section documenting consistent-by-absence policy (option b). No bulk footer-adding performed.

**M4 — Korean → English instruction prose (AC-VFM-004a/b/c):**
- `core/settings-management.md` L40-44: `alwaysLoad` Korean notes → English. Residual Korean: 0.
- `workflow/session-handoff.md` § Diet Constraints + § V0 Abort Gate Doctrine: Korean instruction prose → English. PRESERVED: localization table (en/ko/ja/zh), cut-line markers (`✂────`), paste-ready skeleton examples (`전제 검증:` ×5, `실행:` ×5, `머지 후:` ×4).
- `development/manager-develop-prompt-template.md` Section E (Self-Verification Deliverables): Korean → English. Folds with M6b wall-time removal.

**M5 — Emoji removal (AC-VFM-005):**
- `development/skill-writing-craft.md` L217: `✅` → "Required". `development/skill-ab-testing.md` L210-217: `✅` → "PASS"/"Met". Delta-grep `✅\|❌` == 0 both files.

**M6 — Time-estimation removal (AC-VFM-006a/b/c):**
- `development/orchestrator-templates.md` L199: "Parallel development for 2 days" → "across two phases".
- `development/manager-develop-prompt-template.md` §5 L234-239: wall-time targets ("≤30분", "91분") → phase/priority ordering (folded with M4c).
- `workflow/team-pattern-cookbook.md` L94/99/104: "Day 1/2/3" → "Phase 1/2/3"; L308-310: "First/Middle/Final day" → "Initial/Middle/Final phase".
- Literal values left: `10s` hook-timeout (orchestrator-templates L56), `0.5s` example test-output (manager-develop L174) — NOT planning estimates.

**M7 — Precision framing (AC-VFM-007a/b):**
- `workflow/session-handoff.md` Block 1 L77: "`ultrathink.` triggers Adaptive Thinking xhigh effort" → "`ultrathink.` sets `effort: xhigh` ... Adaptive Thinking is a DISTINCT axis ... explicitly enabled via `thinking: {type: "adaptive"}`".
- `core/hooks-system.md` L12 + L62: "Setup | REMOVED" → "Upstream CC event is CURRENT ... only moai-adk Go EventSetup constant + moai hook setup subcommand retired (internal)".

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-19T20:05:00Z
run_commit_sha: 3803a2686
run_status: complete
ac_pass_count: 21
ac_fail_count: 0
preserve_list_post_run_count: 5
l44_pre_commit_fetch: "0 0 (clean — worktree isolated, origin/main synced)"
l44_post_push_fetch: "n/a — worktree commit, orchestrator cherry-picks to shared main"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  go_build: pass
  goos_windows_goarch_amd64: n/a (doc-only SPEC, 0 Go source change)
total_run_phase_files: 32
m1_to_mN_commit_strategy: "single-run-commit (M1-M7 as one logical scrub — 16 deployed + 16 mirror files)"
```

**Go-test trio result:**
```
$ go test ./internal/template/... -run 'TestRuleTemplateMirrorDrift|TestTemplateNoInternalContentLeak|TestTemplateNeutralityAudit'
ok  github.com/modu-ai/moai-adk/internal/template  0.845s
```

**Mirror parity (pre-commit TOCTOU re-run):**
- 18 in-scope files: IDENTICAL (byte-identical deployed↔template)
- `core/zone-registry.md`: DIVERGENT (expected — deployed-only CONST-V3R6-001 block L1006-1018, §25-stripped from neutral mirror). Scoped diff of L285-300 Opus-clause region: IDENTICAL. AC-VFM-008c mechanism (leak-test + scoped clause-line diff, NOT whole-file diff -q).
- Enrolled files (`core/hooks-system.md`, `workflow/session-handoff.md`): `TestRuleTemplateMirrorDrift` PASS (non-vacuous) + `diff -q` belt-and-suspenders.

**Pre-existing session-handoff mirror divergence resolved:** L227/L236 example block had a pre-existing naming divergence (`Epic`/`Milestone` deployed vs `Wave`/`Round`/`Sprint` mirror — a pre-existing naming-rewrite residue from a sibling SPEC lane, NOT this SPEC's version/format lane). Mirror synchronized to deployed's canonical `Epic`/`Milestone` naming to satisfy the enrolled byte-parity AC (TestRuleTemplateMirrorDrift). This sync is naming-only (2 example lines) and does not touch this SPEC's version/format scope.

**Pre-existing baseline test failures (NOT caused by this SPEC):** `go test ./...` shows 3 pre-existing failures (TestSpecAudit_JSONSchema_DriftFindings, TestAuditLoaderCompleteness, TestCollectMemory) that fail identically on the clean baseline (verified via `git stash` + test + `git stash pop`). These are unrelated to this SPEC's doc-only scope (spec audit drift detection, config loader, memory collection — not rule-file content). Go-test trio (the SPEC-relevant gate) is GREEN.

**R9 keystone validator note:** `moai constitution validate` now loads the registry successfully (CONST-RULEID-001 `ruleIDPattern = ^CONST-V3R[256]-\d{3,}$` landed in source + `make install` applied). However 74 pre-existing DRIFT findings remain (unrelated to this SPEC — various other clauses). AC-VFM-001a (clause substring grep -Fq) verified directly for CONST-V3R2-028/029. AC-VFM-001b (full validator green-gate) remains SHOULD, blocked on the broader 74-finding cleanup (not this SPEC's lane).

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_

## §E.5 Mx-phase Audit-Ready Signal

_<pending Mx-phase — owned by manager-docs / orchestrator>_
