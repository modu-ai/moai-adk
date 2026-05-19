# Research — SPEC-V3R5-LINT-CLEAN-001 — Lint Baseline Analysis

Captured: 2026-05-19 on `main` HEAD `02b2bb0a3`
Author: GOOS Kim (via manager-spec plan phase)

## 0. Baseline Source Disambiguation

The user-facing instruction referenced "golangci-lint baseline ~321 findings (237 ERROR + 84 WARN)". Empirical investigation reveals this count refers to **`moai agent lint --strict`** (the agent-definition linter), NOT `golangci-lint` (the Go source linter). Both are run during plan-phase to confirm:

| Linter | Command | Findings | Notes |
|---|---|---|---|
| `moai agent lint --strict` | `./bin/moai agent lint --strict --format=json` | **176** (93 errors + 83 warnings) | Post-W0 state. Memory claimed 321 pre-W0; W0 reduced by 145 (mostly LR-07 -141 via PR #1005 hotfix). |
| `golangci-lint run ./...` (default config, 6 linters) | `golangci-lint run ./... --output.json.path=...` | **0** issues | No `.golangci.yml` config file in repo; only default linters active (errcheck, govet, ineffassign, staticcheck, typecheck, unused). |
| `golangci-lint run ./... --default=all` (all linters) | Same with `--default=all` | **1773** (capped 50/linter) | Reference-only; not the project's intended baseline. Indicates Go source has style/convention drift if all linters were activated. |
| `moai spec lint --strict` | `./bin/moai spec lint --strict` | **0** | SPEC frontmatter is clean. |
| `moai workflow lint` | `./bin/moai workflow lint` | **0** | workflow.yaml is clean. |
| `go vet ./...` | `go vet ./...` | **0** | No vet issues. |

This SPEC scopes the cleanup to `moai agent lint --strict` because:

1. It is the only non-zero baseline relevant to v3.5.0 release readiness (per memory `project_v3r5_w0_lifecycle_complete` Goal Stop hook).
2. It directly affects agent runtime quality (skill preload drift, effort matrix drift, AskUserQuestion enforcement, worktree isolation enforcement).
3. The Go source path is already clean under the project's chosen lint policy (default linter set).

Reference to `golangci-lint` work is **out of scope** for this SPEC; if Go source lint expansion is desired in the future, a separate SPEC (e.g., `SPEC-V3R5-GO-LINT-EXPAND-001`) should propose `.golangci.yml` adoption with curated linter selection.

## 1. Current Baseline Snapshot (2026-05-19, main HEAD 02b2bb0a3)

### 1.1 Headline numbers

```
Summary: 176 total (93 errors, 83 warnings)
```

Captured to `/tmp/agent-lint-baseline.json` (plan-phase snapshot, NOT committed — runtime artifact).

### 1.2 Rule distribution

| Rule | Severity | Count | Description |
|---|---|---|---|
| LR-08 | warning | 83 | Skill preload drift within same category (manager/expert) |
| LR-06 | error | 29 | Boilerplate `--deepthink flag:` text in description field |
| LR-01 | error | 28 | Literal `AskUserQuestion` token in body text (non-orchestrator agent) |
| LR-03 | error | 25 | Missing `effort:` field in frontmatter |
| LR-12 | error | 6 | Effort drift from SPEC-V3R2-ORC-003 canonical matrix |
| LR-02 | error | 3 | `Agent` tool in CSV (subagents cannot spawn sub-subagents) |
| LR-05 | error | 2 | Write-heavy agent missing `isolation: worktree` |

Total errors: 93 = 29 (LR-06) + 28 (LR-01) + 25 (LR-03) + 6 (LR-12) + 3 (LR-02) + 2 (LR-05)
Total warnings: 83 = 83 (LR-08)
Total: 176

### 1.3 File distribution (template vs live)

Per the template-first rule (CLAUDE.local.md §2), `internal/template/templates/.claude/agents/moai/*.md` is the canonical source. The runtime `.claude/agents/moai/*.md` is a sync copy. The current state shows each finding mirrored across both surfaces (with exceptions noted below).

| Surface | Findings |
|---|---|
| Live (`.claude/agents/moai/*`) | 90 |
| Template (`internal/template/templates/.claude/agents/moai/*`) | 86 |
| **Total** | **176** |

**Asymmetry** (90 - 86 = 4): the live `.claude/agents/moai/expert-mobile.md` file persists (4 findings: LR-02, LR-03, LR-06, LR-08). The template counterpart was hard-deleted by W0 (SPEC-V3R5-CLAUDE-REFRESH-001 REQ-CLR-004). This asymmetry is the residual cleanup blocker — fixing requires either `moai update` synchronization OR explicit deletion of the live copy.

### 1.4 Top-10 files by finding count (live surface, sorted desc)

| Rank | File (live surface) | Findings |
|---|---|---|
| 1 | `.claude/agents/moai/manager-strategy.md` | 7 |
| 1 | `.claude/agents/moai/manager-project.md` | 7 |
| 1 | `.claude/agents/moai/manager-develop.md` | 7 |
| 1 | `.claude/agents/moai/expert-frontend.md` | 7 |
| 1 | `.claude/agents/moai/expert-backend.md` | 7 |
| 6 | `.claude/agents/moai/manager-spec.md` | 6 |
| 6 | `.claude/agents/moai/manager-quality.md` | 6 |
| 6 | `.claude/agents/moai/expert-devops.md` | 6 |
| 9 | `.claude/agents/moai/manager-git.md` | 5 |
| 9 | `.claude/agents/moai/manager-docs.md` | 5 |

The template mirror duplicates these counts (90/2 = ~45 distinct logical units, mostly mirrored 1:1). Cleanup applied to the template will resolve both via `make build` + `moai update`.

### 1.5 LR-08 drift skills (most-missing across category)

| Category | Skill name | Drift count |
|---|---|---|
| manager | moai-foundation-core | 14 (= 7 agents × 2 surfaces) |
| manager | moai-workflow-project | 12 |
| expert | moai-workflow-testing | 9 |
| manager | moai-foundation-thinking | 8 |
| manager | moai-workflow-worktree | 4 |
| manager | moai-workflow-spec | 4 |
| expert | moai-foundation-quality | 4 |

`moai-foundation-core` not being preloaded on every `manager-*` agent is the highest-impact drift; resolving it alone reduces LR-08 warnings by 14 (16.9% of total warnings). The drift typically resolves by adding the skill to the `skills:` YAML array in agent frontmatter, NOT by removing references from outlier agents (the linter favors superset alignment).

## 2. D-Category Classification

Per the user's instruction "D1..D7", the actual baseline maps to the following structurally-distinct categories. The mapping diverges from generic Go lint D1..D7 because the rules under analysis are agent-definition rules, not Go source rules.

| D-Category | Rules | Count | Risk | Mechanical? |
|---|---|---|---|---|
| **D1: Frontmatter Hygiene** | LR-03 (missing effort) + LR-12 (effort drift) | 31 | Low | Yes — add/correct `effort:` field |
| **D2: Description Hygiene** | LR-06 (`--deepthink` boilerplate) | 29 | Low | Yes — string-level removal |
| **D3: Tool Boundary** | LR-01 (literal AskUserQuestion in body) + LR-02 (Agent in tools) | 31 | Medium | Mostly — requires careful body rewrite to avoid losing intent |
| **D4: Worktree Discipline** | LR-05 (missing `isolation: worktree`) | 2 | Low | Yes — add `isolation: worktree` to frontmatter |
| **D5: Preload Drift** | LR-08 (skill preload drift) | 83 | Medium | Partial — adding skills is mechanical, but each addition must be justified for token budget impact |
| **D6: Delta Semantics (meta)** | — | 0 | — | Verification mechanism: `lint-baseline-pre-LCLN-P<N>.json` / `lint-baseline-post-LCLN-P<N>.json` snapshots + `NEW_COUNT=0` invariant per LCLN-Phase |
| **D7: Residual Live-Surface Drift** | All rules on `expert-mobile.md` (live-only) | 4 | Low | Yes — delete file + run sync to confirm template parity |

### 2.1 Risk classification rationale

- **Low-risk (mechanical)**: D1, D2, D4, D7 — pure frontmatter or single-line edits with no semantic impact. ~66 findings (37.5%).
- **Medium-risk (semantic)**: D3, D5 — require understanding the agent's purpose to avoid information loss (D3) or token-budget impact analysis (D5). ~114 findings (64.8%).
- **High-risk (refactor)**: None at this baseline. All 176 findings are addressable without architectural change.

### 2.2 Mega-Sprint W2 Entanglement and Canonical W2-Deferred Set

**Canonical W2-deferred set computation** (per plan-auditor iter-1 Finding 1 + Finding 7): The W2-deferred set is defined by **union** semantics:

```
W2-deferred = {LR-08 finding f | f.drift_skill ∈ {moai-domain-backend, moai-domain-frontend, moai-domain-database}
                                 OR f.affected_agent ∈ {expert-backend, expert-frontend, expert-mobile}}
```

Empirical recount on `main` HEAD `02b2bb0a3` (2026-05-19, captured to `/tmp/agent-lint-baseline.json`):

```bash
jq '[.violations[] | select(.rule == "LR-08") |
     select(
       (.message | test("moai-domain-(backend|frontend|database)")) or
       (.file | test("expert-(backend|frontend|mobile)\\.md$"))
     )] | length' /tmp/agent-lint-baseline.json
# Output: 13
```

**Canonical W2-deferred set size: 13.** The union of (skill-based: 6) + (agent-based: 13) — by set inclusion, every skill-based finding is also agent-based (because the 6 skill-based findings all live on expert-backend or expert-frontend, which are themselves in the agent-retirement set). Therefore |union| = |agent-based| = 13.

| Finding subset | Empirical count (2026-05-19) | Entanglement | Resolution path |
|---|---|---|---|
| LR-08 W2-deferred set (union semantics above) | **13** | Mega-Sprint W2 (SPEC-V3R5-CORE-SLIM-001) retires skills + agents | Defer ~12 of the 13 to Mega-Sprint W2 (1 finding on live expert-mobile.md is cleared by this SPEC's LCLN-Phase 2) |
| LR-08 W4-resolvable subset | **70** (= 83 LR-08 total − 13 W2-deferred) | None | Phase 4 |
| Findings on live `expert-mobile.md` (LR-02 ×1, LR-03 ×1, LR-06 ×1, LR-08 ×1) | **4** | Live file delete-only (template was Mega-Sprint W0-retired) | LCLN-Phase 2 deletes the file. Of these 4, exactly 1 (the LR-08) is also a member of the W2-deferred set; this is the in-flight overlap. |
| LR-12 effort drift on `evaluator-active`/`expert-refactoring`/`plan-auditor` | **6** | Pre-existing SPEC-V3R2-ORC-003 canonical matrix; no Mega-Sprint W-series dependency | LCLN-Phase 1 (T1.2) |
| LR-05 missing `isolation: worktree` on `manager-develop` | **2** | Already fixed in some agent definitions; only `manager-develop` outstanding | LCLN-Phase 1 (T1.4) |
| LR-01 / LR-02 / LR-03 / LR-06 (non-mobile) | **84** = 28 LR-01 + 2 LR-02 (excl. mobile) + 24 LR-03 (excl. mobile) + 28 LR-06 (excl. mobile) + 2 LR-05 | No Mega-Sprint W-series entanglement | LCLN-Phases 1, 3 |

**Net resolvable in this SPEC**: 60 (Phase 1) + 4 (Phase 2) + 30 (Phase 3) + 70 (Phase 4) = **164 findings**.

**W2-deferred residual after this SPEC**: 176 − 164 = **12 findings**.

Numeric reconciliation between "set size = 13" and "residual = 12":
- The W2-deferred SET (computed at baseline, before any LCLN-Phase runs) has **13 elements**.
- Of those 13 elements, exactly **1** is the LR-08 finding on live `.claude/agents/moai/expert-mobile.md` (line 16 message: "Skill preload drift in category 'expert': moai-workflow-testing is not preloaded by all agents").
- LCLN-Phase 2 deletes the entire `expert-mobile.md` file, clearing that 1 finding in-flight.
- Therefore the **post-Phase-2 residual** is 12, and remains 12 through Phase 3, Phase 4, and the final post-this-SPEC state.
- Mega-Sprint W2 (CORE-SLIM-001) dissolves the remaining 12 when it retires `expert-backend.md` and `expert-frontend.md` agents.

**AC-LCLN-005.2 bound derivation** (per plan-auditor iter-1 Finding 7): The bound `[11, 16]` for the W2-deferred set size is derived as `13 ± 3 upstream drift tolerance`. The ±3 covers:

1. Mega-Sprint W1 (CONSTITUTION-DUAL) may introduce or remove rule-document references that the linter's drift detection treats as `expert-backend.md` / `expert-frontend.md` modifications, perturbing the set count by ±1.
2. Mega-Sprint W2's interim run-phase commits (before CORE-SLIM lifecycle COMPLETE) may add or remove the affected agent files temporarily, perturbing the count by up to ±2.
3. The 12-element residual (post-Phase-2 in this SPEC's run-phase) is the lower bound case (= 13 − 1 mobile = 12, but ±1 tolerance pushes to 11).
4. Upper bound 16 accommodates concurrent perturbation without forcing a re-plan.

The iter-1 bound `[6, 18]` is replaced because:
- Lower 6 corresponded to the intersection-only (skill-based) interpretation, which contradicts the union semantics actually used.
- Upper 18 was unjustified — no derivation in iter-1 docs.

Tighter bound `[11, 16]` is operationally correct under the union semantics and the 3-finding upstream-drift tolerance.

## 3. Methodology

### 3.1 Capture command

```bash
./bin/moai agent lint --strict --format=json > /tmp/agent-lint-baseline.json
```

The `--format=json` flag emits a structured envelope:

```json
{
  "version": "1.0",
  "summary": { "total": 176, "errors": 93, "warnings": 83 },
  "violations": [
    { "rule": "LR-XX", "severity": "error|warning", "file": "...", "line": N, "message": "..." }
  ]
}
```

### 3.2 Cross-validation commands

```bash
# Go source: should remain 0 throughout SPEC lifecycle
golangci-lint run ./... --timeout=10m

# SPEC frontmatter: should remain 0 throughout (this SPEC must not introduce SPEC drift)
./bin/moai spec lint --strict

# Workflow YAML: orthogonal, should remain 0
./bin/moai workflow lint
```

All three above MUST remain at 0 after every LCLN-Phase merges, to ensure no collateral damage.

### 3.3 Delta verification (D6 semantics)

Per `project_v3r5_w0_lifecycle_complete` § AC-CLR-008 precedent:

```bash
# Pre-phase: capture baseline
./bin/moai agent lint --strict --format=json > .moai/state/lint-baseline-pre-LCLN-P<N>.json

# Post-phase: capture post-merge state
./bin/moai agent lint --strict --format=json > .moai/state/lint-baseline-post-LCLN-P<N>.json

# Delta computation
PRE_COUNT=$(jq '.summary.total' .moai/state/lint-baseline-pre-LCLN-P<N>.json)
POST_COUNT=$(jq '.summary.total' .moai/state/lint-baseline-post-LCLN-P<N>.json)
NEW_FINDINGS=$(jq -s '
  (.[1].violations | map(.file + ":" + (.line|tostring) + ":" + .rule)) -
  (.[0].violations | map(.file + ":" + (.line|tostring) + ":" + .rule)) | length
' .moai/state/lint-baseline-pre-LCLN-P<N>.json .moai/state/lint-baseline-post-LCLN-P<N>.json)

# LCLN-Phase acceptance:
# - POST_COUNT must equal PRE_COUNT − phase reduction target (60/4/30/70)
# - NEW_FINDINGS must equal 0 (no new findings introduced)
echo "Pre: $PRE_COUNT, Post: $POST_COUNT, New: $NEW_FINDINGS"
```

This preserves Mega-Sprint W0's delta-only D6 NEW=0 semantics while also enforcing forward progress (POST < PRE).

## 4. Discrepancies vs Memory

| Source | Claim | Reality | Delta | Explanation |
|---|---|---|---|---|
| Memory `project_v3r5_w0_lifecycle_complete` | "321 findings (237 ERROR + 84 WARN)" | Currently 176 (93+83) | -145 | W0 PR #1005 hotfix retired LR-07 v2 fingerprint (-141 errors). Other deltas: -3 LR-12 + -1 LR-03 + -2 LR-08 = -5 finer. Net: -145. **Memory was true pre-W0 baseline; post-W0 is current.** |
| User instruction | "reduce golangci-lint baseline" | golangci-lint baseline is 0 under default config | N/A | Likely terminology confusion: the relevant linter is `moai agent lint`, not `golangci-lint`. This SPEC corrects the framing. |
| User instruction | "D6 delta-only NEW=0 semantics" | Confirmed — W0 used same mechanism (AC-CLR-008) | 0 | Adopted verbatim; see §3.3 above. |

## 5. Why this SPEC Now (Strategic Rationale)

The v3.5.0 Mega-Sprint roadmap defines W0..W4 in `.moai/research/harness-autonomy-vision-2026-05-18.md` §5. Mega-Sprint W0 closed at 176 findings (down from 321). The roadmap reads:

> Mega-Sprint W2 SPEC-V3R5-CORE-SLIM-001 ... retires expert-{backend,frontend,mobile} causing the LR-08 preload drift findings.

Implication: **13** findings (canonical W2-deferred set per §2.2) will dissolve organically when Mega-Sprint W2 lifecycle COMPLETEs. Of those 13, this SPEC's LCLN-Phase 2 resolves 1 in-flight (LR-08 on live `expert-mobile.md`); the remaining 12 await Mega-Sprint W2.

The other **163** findings (= 176 − 13) are **independent of Mega-Sprint W2** — they are agent-definition hygiene items that should not block on architectural retirement decisions. This SPEC's LCLN-Phases 1-4 collectively resolve **164** findings (163 W2-independent + 1 W2-overlap cleared by Phase 2 deletion).

This SPEC's purpose is to:

1. Resolve the 164 findings (163 Mega-Sprint W2-independent + 1 in-flight overlap via Phase 2) via mechanical/semi-mechanical cleanup phases.
2. Preserve Mega-Sprint W2 entanglement: explicitly defer the 12 W2-overlapping residual findings to avoid duplicate work.
3. Establish a delta-only verification baseline (`lint-baseline-pre-LCLN-P<N>.json`) that future SPECs can adopt for cross-SPEC quality continuity.

Without this SPEC, the 176 baseline will persist across Mega-Sprint W1 and W2 plan/run cycles, polluting their delta checks and obscuring whether their own changes introduce regressions. The cleanup is a **prerequisite for high-fidelity delta verification** in the remainder of v3.5.0 Mega-Sprint.

## 6. Boundaries (What This SPEC Does NOT Touch)

- **FROZEN zones** per `.claude/rules/moai/core/zone-registry.md`:
  - `.claude/rules/moai/core/moai-constitution.md` § Agent Core Behaviors (CONST-V3R2-030..035)
  - CONST-V3R2-001..010 (Phase Overview, TRUST 5, MX TAG, language neutrality, Template-First, AskUserQuestion monopoly, Claude Code substrate)
  - `.claude/rules/moai/design/constitution.md` § FROZEN Zone (CONST-V3R2-051..072)
  - Any file under `.claude/rules/moai/core/` requires explicit FROZEN check before edit
- **Go source code under `internal/`, `pkg/`, `cmd/`** (where memory's framing pointed; baseline is 0, no action needed)
- **SPEC documents under `.moai/specs/`** (`moai spec lint` is 0; do not touch)
- **Mega-Sprint W2-deferred set** (canonical 13 elements; 12 post-Phase-2 residual) — documented in §2.2 above to avoid duplicate work with SPEC-V3R5-CORE-SLIM-001

## 7. References

- `.moai/specs/SPEC-V3R5-CLAUDE-REFRESH-001/spec.md` — W0 SPEC (precedent for AC-CLR-008 delta-only semantics)
- `.moai/research/harness-autonomy-vision-2026-05-18.md` §5 W0-W4 — Mega-Sprint roadmap (W2 entanglement)
- `.moai/research/architecture-audit-2026-05-18.md` — defect catalog (F-001..F-105)
- `.claude/rules/moai/core/zone-registry.md` — FROZEN zone enumeration (CONST-V3R2-001..152)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — 12-field canonical schema (this SPEC complies)
- `internal/cli/agent_lint.go` — LR-01..LR-14 rule implementations (source-of-truth for finding semantics)
- Memory entry `project_v3r5_w0_lifecycle_complete` — provides the 321 baseline reference
- `CLAUDE.local.md` §2 (Template-First), §18.2 (branch naming), §22 (dev settings intent)

---

Captured baseline retained for run-phase delta verification:

```
/tmp/agent-lint-baseline.json  (176 findings, plan-phase snapshot)
/tmp/lint-baseline-w0.json     (321 findings, pre-W0, retained for historical reference)
```
