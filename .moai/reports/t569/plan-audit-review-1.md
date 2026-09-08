# SPEC Review Report: SPEC-HARNESS-EVIDENCE-WRITE-001

Iteration: 1/1 (Tier S ceiling per `harness.plan_audit_tier_ceilings` S=1)
Verdict: **FAIL**
Overall Score: **0.875** (Tier S PASS threshold 0.75 — numerically above; FAIL is driven by blocking defects D1-D3 and the required-backend codex convergence FAIL, not by the aggregate)

Reasoning context: the dispatch carried lane-10 ground-truth claims. These were treated as
verifiable input claims, and each was independently re-verified mechanically against this
worktree (HEAD `3ac58b5a1`) before use — see §A premise verification below. No author reasoning
was accepted unverified.

Audit target: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t569/.moai/specs/SPEC-HARNESS-EVIDENCE-WRITE-001/`
(spec.md + plan.md + progress.md; Tier S artifact set complete — AC inline in spec.md §C, no
acceptance.md, correct for Tier S).

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency**: REQ-001..REQ-007 at spec.md:72-97 — sequential, zero-padded, no gaps, no duplicates.
- **[PASS] MP-2 EARS/GEARS format compliance** (judged against the requirement layer, spec.md §B REQ-XXX entries only): REQ-001 Unwanted form (`shall not`, L72-74), REQ-002/003 Event-driven (L75-80), REQ-004 Event+Where compound (L81-84), REQ-005 Unwanted (L85-89), REQ-006 Event-driven (L90-93), REQ-007 Where (L94-97). All seven match GEARS patterns or compounds thereof. Label cosmetics (REQ-001 tagged "Ubiquitous" for an Unwanted form; REQ-006 "Event-detected" vs canonical "Event-driven") filed as D6 — the pattern FORM is what MP-2 binds, and it holds. AC Given-When-Then entries in §C were NOT GEARS-tested (verification layer, per M3 § Scope).
- **[PASS] MP-3 YAML frontmatter validity**: spec.md:1-15 carries all 12 canonical fields with correct types/enums — `id` SPEC-HARNESS-EVIDENCE-WRITE-001, `title` quoted, `version` "1.0.0" quoted semver, `status: draft` (valid enum), `created`/`updated` 2026-09-08 ISO, `author`, `priority: P1`, `phase: "v3.2.0 target"` (release target, not a prohibited stage name), `module: "internal/spec"`, `lifecycle: spec-anchored`, `tags` comma-separated string. Optional `tier: S` present. No rejected snake_case aliases.
- **[N/A] MP-4 Section 22 language neutrality**: single-language scoped SPEC (Go test hygiene, module `internal/spec`) — auto-pass per MP-4.
- **[PASS] MP-5 D7 cross-SPEC reconciliation**: referenced SPECs extracted and checked — `SPEC-COVERAGE-RULE-SCOPE-001` → `status: completed`, `SPEC-SPEC-LINT-BLIND-AXES-001` → `status: completed` (both exist under `.moai/specs/` in this worktree). Neither is retired/superseded/archived → no reconciliation obligation, no BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline**: `grep -c syscall` on spec.md → **0 hits** → D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate**: `grep -rn '\[NEEDS CLARIFICATION'` on plan.md → no markers; research.md absent (Tier S — no marker surface), plan.md checked and clean.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 1.0 | 1.0 band | Every REQ resolves to a single interpretation: REQ-001's "by default" is operationalized by REQ-002/003/004; REQ-005's "deleted or repurposed" is bounded by the decidable clause "so no such path remains constructible" (spec.md:88-89). The defects found are correctness defects, not ambiguity. |
| Completeness | 1.0 | 1.0 band | HISTORY (spec.md:19-23); §A Context carries WHY+WHAT + enumerated premise record + root cause (L27-68); §B REQs (L70-97); §C ACs (L99-122); §D NFR (L124-137); §E Out of Scope with four `### Out of Scope — <topic>` H3 headings each with specific `-` bullets (L139-159); §F cross-refs (L161-170). plan.md carries the HOW (§D constraints, §F milestones M1-M4). Standalone "WHAT/Overview" H2 absent — content carried by §A per the repo's §-letter V3R6 convention (disclosed, not scored down). |
| Testability | 0.75 | 0.75 band | AC-005/006 are strong (AC-006 requires an observed RED on a demonstrated mutant — spec.md:116-119 — plus plan.md §E steps 1-4 require verbatim RED capture side-by-side with GREEN before adoption, satisfying the observed-failure/mutant-probe doctrine). AC-003 pins `T528_PROBE_OUT` unset. But AC-001/002/004 pin no environment scrub and no `-count=1`, leaving a sanctioned vacuous-green vector (stale cached PASS or inherited env exercising the wrong branch) — measurable with minor interpretation, not airtight. |
| Traceability | 0.75 | 0.75 band | One REQ uncovered: REQ-007 has no corresponding AC (AC-001..007 verify TempDir retargeting, t.Logf announcement, guard GREEN/RED, zero-diff on pinned evidence — none asserts the header content or the capture procedure). All other REQs covered: 001←AC-001/002/003/005; 002←AC-001; 003←AC-002; 004←AC-003/004; 005←AC-005/006/007; 006←AC-005/006. ACs carry no explicit REQ tags (repo Tier S convention; mapping inferable) — noted under D6 scope, not scored further. |

Aggregate: (1.0 + 1.0 + 0.75 + 0.75) / 4 = **0.875**

## Defects Found (structured defect-list)

D1. **D1** — spec.md:94-97 (REQ-007) — **Unimplementable as written**: REQ-007 directs the operator to copy the report "from the `t.Logf`-announced `t.TempDir()` path" as an act performed OUTSIDE the test run, but `t.TempDir()` directories are removed when the test and its subtests complete (Go stdlib contract) — the announced path no longer exists post-run. The decompose harness logs only the path (`t.Logf("decomposition written to %s", out)`), not content, so nothing durable survives the run for it; the corpus harness does log full content (`t.Logf("measurement written to %s\n%s", out, b.String())`). Joint effect of REQ-005 (TempDir retarget) + REQ-007 (post-run copy from TempDir): durable capture capability for the t362 pair is eliminated — an internal contradiction (CN-1). Found independently by this audit's convergence round (codex required-gate FAIL, finding 1) after the initial in-session pass missed it. — Severity: **critical** — Class: **blocking** — Required fix: reword REQ-007 so durable capture selects an explicit output location BEFORE the run (the single env override already contemplated in §D L134-135 / Out of Scope §E L156-159 — with no-clobber, NEW-filename semantics per REQ-007's own "never overwriting a pinned file"), or require full-content emission through test output for external capture. Keep the TempDir default for ordinary runs unchanged.

D2. **D2** — plan.md:51-58 (§E) + spec.md:101-108, 112-113 (AC-001/002/004) — **Vacuous-green vectors in the verification batch**: §E commands omit `-count=1` (Go caches successful package results — a stale cached PASS satisfies AC-001/002 without executing the current write path), and the "without gate" run (plan.md:53) does not scrub inherited `MOAI_T362_CORPUS_SCAN` / `T528_PROBE_OUT` (an inherited value silently exercises the gated/default-override branch instead of the branch under test). For a SPEC whose subject is verification integrity, this is an internal inconsistency with the observed-failure doctrine. — Severity: **major** — Class: **blocking** — Required fix: state environment-scrubbed single-invocation form in §E and in the AC preconditions, e.g. `unset MOAI_T362_CORPUS_SCAN T528_PROBE_OUT && go test ./internal/spec/ -run '...' -count=1 -v`; AC-003 already pins the env precondition — extend the same pinning to AC-001/002/004.

D3. **D3** — spec.md:94-97 / §C (L99-122) — **REQ-007 has no covering AC** (traceability gap; the header-update work is planned at plan.md:20-22 and plan.md:83 but nothing verifies it). — Severity: **major** — Class: **blocking** — Required fix: after D1's rewording, add AC-008: Given the fixed harness sources, When the in-report `# produced by:` header is read, Then it describes the (workable) durable-capture procedure and names no repo-relative default path.

D4. **D4** — spec.md:85-93 (REQ-005/006) vs plan.md:24-26 (§B) vs spec.md:114-115 (AC-005) — **Guard detection predicate stated at three breadths**: REQ-005 (repo-root join + `.moai/**` destination), REQ-006 (`.moai/`-anchored path + write primitive), plan §B (write primitive regardless of path string), AC-005 ("repo-anchored write targets"). All variants catch the demonstrated mutant class; they differ for non-`.moai` repo-tree writes (covered by REQ-001's stated invariant, enforced by none) and a naive path-string detector would false-positive on legitimate TempDir-rooted `.moai/` fixtures (e.g. `writeFixtureSpecs` joins `root/.moai/specs/<id>` with root=t.TempDir()). — Severity: minor — Class: optional — Required fix: harmonize on one predicate and define repo-root provenance (AST-level or root-provenance-aware, reads exempted); add a false-positive control (TempDir-rooted `.moai` fixture must stay GREEN) alongside the RED mutant.

D5. **D5** — spec.md:81-84 (REQ-004) / spec.md:112-113 (AC-004) — the preserved `T528_PROBE_OUT` override, if an operator points it at the pinned evidence directory, overwrites tracked before-images (`probe/after/*` — the 8 names are tracked there per `git ls-files`). Status-quo behavior the SPEC deliberately preserves byte-for-byte (plan.md:44); it introduces no new hazard, and the operator-approved direction is "never into the repo **by default**". — Severity: minor — Class: optional — Required fix (future hardening, likely a follow-up SPEC): reject repo-internal targets after symlink resolution, or no-clobber file creation.

D6. **D6** — spec.md:72, 90 — GEARS label cosmetics: REQ-001 labeled "(Ubiquitous)" but written in Unwanted form; REQ-006 labeled "(Event-detected)" vs canonical "(Event-driven)". — Severity: minor — Class: optional — Required fix: relabel.

## §A Premise Verification (ground truth re-checked against worktree HEAD `3ac58b5a1`, this run)

Every §A claim was mechanically re-verified and all held:

1. `decomposeReportRelPath` at lint_req_widen_decompose_test.go:15, value `.moai/reports/t362/m2-gate0-decomposition.txt`; write site `os.MkdirAll` + `os.WriteFile` at ~541-548. ✓
2. `corpusMeasurementRelPath` at lint_req_widen_corpus_test.go:19; write site ~239-247. ✓
3. `git ls-files .moai/reports/t362/` → exactly 12 tracked files including both report files. ✓
4. `git merge-base --is-ancestor 130846ab2 fc02d2542` → yes (t362 commit exists, subject confirms t362 provenance). ✓
5. `t528ProbeOutDir` (zz_t528_anchor_probe_test.go:55-65): ungated default `../../.moai/reports/t528/probe/out`, env override `T528_PROBE_OUT`; exactly 8 `t528Write` call sites; `git ls-files .moai/reports/t528/probe/out/` → 0 tracked; header states "run-scoped directory, NOT the plan-phase artifact directory". ✓
6. `ac_count_clause_test.go` reads tracked `.moai/reports/t338/ac-count-baseline.txt` (its two `os.WriteFile` sites write MUTANT COPIES into `t.TempDir()` — the READ-ONLY claim holds); `zz_t528_overacceptance_test.go` reads `../../.moai/reports/t528/probe/nondecl-bullets.txt`. `findRepoRoot` defined at drift_doctrine_test.go:13, used by readers. ✓
7. **Defect census complete**: swept every write primitive (`os.WriteFile`/`os.MkdirAll`/`os.Create`) across `internal/spec/*_test.go` (25 files, ~120 sites) — every site other than the three named writers resolves to a `t.TempDir()`-rooted fixture (catalog_hash_test.go:240 writes a TempDir copy; haiku/req_table/drift_cache/wiring/listdocs/lint_phase/drift_characterization/lint_status_unreachable fixtures all TempDir-rooted; progress_section_letter_guard is read-only). The SPEC names all writers and no others.

## Cross-Model Convergence (mcp__moai__audit_multi, project_root = this worktree)

- claude (required): pass (initial in-session analysis)
- **codex (required): FAIL** — 4 findings; findings 1 (TempDir evaporation vs REQ-007) and 4 (-count=1 / env-scrub) verified valid on the merits and folded in as D1 and D2; finding 3 folded into D4; finding 2 folded into D5 with a path correction (codex cited `probe/before/`; tracked before-images live under `probe/after/` per `git ls-files` — substance unaffected).
- glm (advisory): inconclusive (fail-open — `uncommittedChanges` produced an empty diff; the SPEC directory is untracked). `fail_open_backends: ["glm"]`.

Residual-risk note (verbatim): `cross-model disagreement (advisory, NOT a block): pass=[claude(required)] fail=[codex(required)]`

Convergence-policy note: the engine's conservative required-split rule yields `overall_verdict: fail`. Verdict authority remains with this agent; the FAIL below is independently grounded in D1-D3 (verified blocking defects), and codex's required FAIL corroborates rather than dictates it. Side note from the engine: the moai MCP server binary lags this tree's HEAD — it did not affect this audit (all evidence read directly from the tree).

## Regression Check

N/A — iteration 1 (no prior-iteration defects exist).

## Recommendation

**FAIL — revision round required before run-phase entry.** The defects are narrow and cheap; this is a revision FAIL, not a structural one. Numbered fixes for manager-spec (reference lines are spec.md unless noted):

1. **Reword REQ-007 (L94-97)** so durable capture happens BEFORE the run: select an explicit output location via the single env override already contemplated in §D L134-135 / Out of Scope §E L156-159, with no-clobber NEW-filename semantics; alternative or complement: require full-content emission via test output (the corpus harness already does this). Do not weaken the TempDir default for ordinary runs.
2. **Pin the verification environment**: add `-count=1` and the compound `unset MOAI_T362_CORPUS_SCAN T528_PROBE_OUT && ...` form to plan.md §E (L51-58), and add the scrubbed-environment precondition to AC-001/002/004 (AC-003 already has it).
3. **Add AC-008** covering the (now workable) `# produced by:` header content and the durable-capture procedure, closing the REQ-007 traceability gap.
4. Optional (orchestrator discretion): harmonize the guard predicate (D4) and relabel REQ-001/REQ-006 (D6). D5 is follow-up-SPEC material.

A confirming re-audit after these fixes is delta-scoped to D1-D3 per the Retry Loop Contract. Note: Tier S ceiling is 1 iteration, so after the fix lands the orchestrator should route the delta re-audit explicitly rather than counting on a further in-loop iteration.
