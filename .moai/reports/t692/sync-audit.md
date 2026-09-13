# Sync-Phase Audit Report: SPEC-DECISION-AUTHORITY-001

Auditor: sync-auditor (independent) — tree `.claude/worktrees/t692`, HEAD `cdc374416`,
branch `WT-judgment-authority` (card t692, Tier L). Scope: independent re-verification of the
§E.2 AC matrix (16 Critical + 3 guard ACs), the mandatory two-sided coverage re-measurement,
sync-commit correctness (`fe1fa0508` + backfills `8f2ab282a`/`cdc374416`), and branch scope
vs plan.md §F. Audit-only: nothing fixed, nothing committed.

**Verdict: PASS**
**Overall Score: 0.92** (harmonic mean of 1.00 / 1.00 / 0.75 / 1.00; must-pass dimensions
Functionality + Security both pass their thresholds)

---

## Dimension Scores (default profile, flat weighted)

| Dimension | Weight | Score | Verdict | Evidence |
|-----------|--------|-------|---------|----------|
| Functionality | 40% | 1.00 | PASS (must-pass) | All 19 ACs re-executed this run — see AC table below; `go test -count=1 -cover ./internal/config/...` green (3.682s) |
| Security | 25% | 1.00 | PASS (must-pass) | 0 secret-shaped tokens / 0 URLs in all added lines; no deny/block paths (CONST-7 held — Step 2.3.3b states "No second human gate is introduced", fail-open); no new trust boundary (pure resolver + doctrine text) |
| Craft | 20% | 0.75 | PASS-WITH-DEBT | Coverage 82.0% < 85% — PRE-EXISTING, two-sided measurement below; new code 100%; 0 new lint issues (1 pre-existing QF1001 in `cg_migration.go:159`, a file this SPEC never touched); go vet exit 0; darwin+windows builds green |
| Consistency | 15% | 1.00 | PASS | Resolver mirrors the landed `ResolvedRecommendationMode` precedent verbatim (CONST-2); mirror token-presence exact in both trees (CONST-8 honored — no `cp`, no `diff -q`); template neutrality clean; CHANGELOG entry shape matches the sibling SPEC entry |

Harmonic mean: 4 / (1/1.00 + 1/1.00 + 1/0.75 + 1/1.00) = **0.92**.

---

## Mandatory Coverage Re-measurement (lead requirement, duty 2)

**Gate judgment: PASS-WITH-DEBT.** The 85% figure (acceptance §D.5 gate 3, plan.md E3, profile
Craft threshold) is NOT met at package level — but the shortfall is pre-existing, proven by a
two-sided measurement I ran myself in this audit, and the SPEC's new code is fully covered.

### Side A — this tree (post-change, HEAD `cdc374416`)

Command: `go test -count=1 -cover ./internal/config/...`

```
ok  	github.com/modu-ai/moai-adk/internal/config          3.682s	coverage: 82.0% of statements
ok  	github.com/modu-ai/moai-adk/internal/config/atomicfile	0.482s	coverage: 81.8% of statements
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.810s	coverage: 89.1% of statements
```

### Side B — pre-change baseline (throwaway clone at `c388db2fa`, created under /tmp, cleaned up after)

Commands: `git clone -q <worktree> /tmp/t692-audit-baseline && git -C /tmp/t692-audit-baseline checkout -q c388db2fa`, then `cd /tmp/t692-audit-baseline && go test -count=1 -cover ./internal/config/...`

```
ok  	github.com/modu-ai/moai-adk/internal/config          5.619s	coverage: 82.0% of statements
ok  	github.com/modu-ai/moai-adk/internal/config/atomicfile	0.506s	coverage: 81.8% of statements
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	1.127s	coverage: 89.1% of statements
```

**The two sides are identical: 82.0 / 81.8 / 89.1.** Zero coverage regression attributable to
this SPEC. The §E.3 claim is confirmed verbatim by independent re-measurement.

### New-code coverage

`go tool cover -func` on the fresh profile:

```
types.go:1370:  ResolvedRecommendationMode  100.0%
types.go:1384:  ResolvedDecisionGate        100.0%
```

### Where the shortfall lives (all pre-existing, none touched by this SPEC)

25 of 268 functions in `internal/config` measure 0.0%; the deficit is concentrated in:
`resolver.go` (tier loaders — `loadSkillTier` 10.0%, `loadPolicyTier` 38.5%, `parseSkillFrontmatter`
0%, `loadJSONFile` 0%, `loadUserTier` 52.4%, `loadLocalTier` 55.0%, `detectTier` 57.1%),
`closed_sets.go` (6 zero-coverage `Valid*` functions), `loader_crosssession.go` (4 functions),
`loader_git_mode.go` (`LoadGitMode`), `envkeys.go` (`GLMEnvVarSet`), `log.go`, `merge.go`
(`dumpYAML`), `profile.go`, `provenance.go` (`MarshalJSON`). Repairing any of these is outside
this SPEC's scope (scope discipline, AGENTS.md §5) and would have been a drive-by refactor.

---

## AC Re-verification (duty 1 — all probes re-executed this audit, this tree, HEAD `cdc374416`)

| AC | Re-verdict | My evidence (this run) |
|----|-----------|------------------------|
| 001 | PASS | `grep -c decision_gate internal/config/types.go` → 1; `grep -c DecisionGate internal/config/defaults.go` → 1 (default `"off"` at :1154) |
| 002 | PASS | `go test -run TestDecisionGate -v ./internal/config/` → 6 functions (floor ≥4), all PASS; fresh `-count=1` suite green |
| 003 | PASS | Template `interview.yaml:9 decision_gate: off`; local `:6 decision_gate: on`; `TestDecisionGateAbsentKeyIsOff` passes |
| 004 | PASS | All occurrences block-conditioned in BOTH trees — manager-spec 1/1, clarity-interview 2/2 (Where-on + Where-off halves read verbatim), spec-assembly Step 2.3.3b (Where-on + fail-open Where-off), mirrors carry the same "Where the `interview.decision_gate` setting" blocks |
| 005 | PASS | decision-index in manager-spec.md = 1 local + 1 mirror; block states statelessness + gate-on-only authoring |
| 006 | PASS | POLICY-COVERED: manager-spec 3, clarity-interview 1 (local+mirror); second-vocabulary sweep (`URGENT\|CRITICAL-FLAG\|BLOCKER`, 6 files) → 0 matches |
| 007 | PASS | "authority anchor": manager-spec 1, clarity-interview 2 (local+mirror) |
| 008 | PASS | "authority register": manager-spec 1, clarity-interview 1 (local+mirror) |
| 009 | PASS | "Never downgrade": manager-spec 1, clarity-interview 2 (local+mirror) |
| 010 | PASS | decision-index in spec-assembly.md = 1 local + 1 mirror (Step 2.3.3b: enrichment, fail-open, "gate is unchanged") |
| 011 (guard) | PASS | Exactly 1 `[HARD] The Implementation Kickoff Approval` clause local AND mirror; text verbatim vs base `62fbd6baf` (three canonical options + `(권장)` label clause identical); Step 2.3.3b inserted as a sibling section beside, not inside |
| 012 (guard) | PASS | decision-index in plan-auditor.md = 0; `git diff --stat 62fbd6baf..HEAD -- .claude/agents/moai/plan-auditor.md` → empty |
| 013 | PASS | NEED_ANALYSIS in spec-assembly.md = 1; DECIDE/NEED_ANALYSIS/NEED_EVIDENCE/DEFER vocabulary read at :245-246 |
| 014 | PASS | product.md in spec-assembly.md = 1; manager-docs ownership boundary named at :249-252 |
| 015 | PASS | `grep -ic "zero judgment"` → 1 |
| 016 | PASS | "Detect → Explain": manager-spec 1, clarity-interview 2; "never an embedded preferred answer, in either recommendation mode" read verbatim in both surfaces |
| 017 | PASS | Coupling grep = 0 in types.go AND defaults.go; `TestDecisionGateOrthogonalToRecommendationMode` PASS; resolver read — reads its own field only |
| 018 | PASS | Token-presence in both trees for all 6 distinctive tokens × 2 files + spec-assembly mirror; counts exact |
| 019 (guard) | PASS | `/tmp/t692-repro` inspected: off-mode (`interview.yaml` with no `decision_gate` key) → 0 decision-index.md files (find count 0); on-mode decision-index.md → exactly 4 labels (POLICY-COVERED/DECIDED/EVIDENCE-NEEDED/FOUNDER); FOUNDER row's anchor `docs/pricing.md § Tiers` does not exist → correctly escalated, not downgraded |

## Sync-Commit Correctness (duty 3)

- `fe1fa0508` — subject names the full SPEC-ID + "3-phase close" (close-subject mandate met);
  exactly 3 files: spec.md (1+/1- = `status: in-progress → completed` ONLY — frontmatter-only,
  zero body edits, verified in the diff), CHANGELOG.md (+4), progress.md (§E.4 signal);
  `🗿 MoAI` trailer present.
- CHANGELOG entry accuracy: every path it cites exists and carries the claimed change (all
  verified above); the coverage disclosure inside the entry (82.0% pre-existing, new resolver
  100%) matches my independent measurement; `make agents-emit` provenance claim confirmed
  (`.codex/agents/moai/manager-spec.toml` carries the new flow tokens).
- `8f2ab282a` — `run_commit_sha: pending-backfill-m5 → 25d04f292` (1-line backfill, sanctioned
  by the SHA-placeholder exemption).
- `cdc374416` — `sync_commit_sha: pending-backfill-sync → fe1fa0508` (same exemption).
- `25d04f292` (M5) touched only progress.md (86+/2-) — so §E.2's claim that the closing-sweep
  measurement tree `13486b852` differs from M5 only by the progress record holds.

## Scope Check (duty 4)

`git diff --name-status 62fbd6baf..HEAD` = 23 files. All accounted for:

- Plan-declared (plan.md §F): 3 local flow surfaces, 3 mirrors, both interview.yaml files,
  types.go, defaults.go, new test file. `loader_interview.go` declared "no change expected" —
  confirmed unchanged.
- Mechanical consequences (justified, not violations): `internal/config/testdata/shipped_key_inventory.yaml`
  (the config key-inventory test REQUIRES the new `interview.decision_gate` entry — class W,
  evidence: reader); `internal/template/catalog.yaml` (template hash regeneration on `make build`).
- Phase deliverables: 6 SPEC artifacts, CHANGELOG.md, 3 evidence files under `.moai/reports/t692/`
  (card evidence is commit-eligible per the §A.3 D4 measurement; committed at plan-phase `6732d1461`).

No unexplained file. Scope PASS.

---

## Findings

- **F1** [info] [optional] `internal/config` (package-level) — coverage 82.0% sits below the
  85% gate figure. Confidence: high (two-sided measurement above). Pre-existing and
  baseline-identical; new code 100%; 0 regression. Non-blocking. Required fix: none for this
  SPEC — the debt belongs to a package-repair card targeting `resolver.go` tier loaders /
  `closed_sets.go` / `loader_crosssession.go` if the maintainer wants the gate figure met.
- **F2** [info] [optional] plan.md §F file lists — two changed files were not anticipated by
  name (`shipped_key_inventory.yaml`, `catalog.yaml`). Confidence: high. Both are mechanical
  consequences of files the plan DID declare. Non-blocking. Required fix: none (future plans
  touching config keys or template content can name the inventory/catalog as expected
  by-products).
- **F3** [low] [optional] spec.md:439 (§G) — describes the proposal as "untracked evidence",
  but the file was committed in the same plan-phase commit `6732d1461` that authored the
  reference; the descriptor is stale relative to the closed tree. §A.3's "currently untracked"
  is a point-in-time measurement accurate at its pin and needs no change. Confidence: high.
  Non-blocking, cosmetic.
- **F4** [info] [optional] acceptance.md §D.5 gate 3 — the "coverage ≥ 85%" closure gate
  over-scopes for a SPEC that adds ~20 lines to a 268-function package with an inherited
  82.0%. The SPEC disclosed the shortfall honestly in §E.3 and CHANGELOG rather than gaming
  it. Confidence: high. Non-blocking. Required fix (future): scope the gate to changed code,
  or carry a standing package-coverage card.

## Gaps (explicitly unobserved)

- `make build` / `make agents-embed` regeneration was verified indirectly (catalog.yaml hashes
  updated; `.codex/manager-spec.toml` carries the new tokens; `make build`'s `agents-emit-check`
  is wired into CI) — I did not re-run `make build` itself (audit-only constraint; it rewrites
  generated files).
- The `go test ./internal/template/...` suites referenced in §E.3's template-neutrality note
  were not re-run by me; neutrality was verified directly on the added template lines instead
  (0 forbidden tokens).
- The M5 on-mode repro was verified against its persisted artifacts under `/tmp/t692-repro`
  (still present at audit time, 22:03) — I did not re-execute the authoring flow itself.

## Residual-risk

- The decision-gate flow text (Step 2.3.3b, clarity-interview, manager-spec) is doctrine
  executed by agents, not code — its behavioral correctness under `on` rests on the M5 repro
  plus the block-conditioning structure, not on a mechanical test. The off-mode regression
  property (the load-bearing one) is mechanically anchored by the config tests and the repro.
- The two coverage figures (82.0%) will drift as `internal/config` evolves; the pre-existing
  attribution in this report is pinned to trees `cdc374416` / `c388db2fa` measured 2026-09-13.
