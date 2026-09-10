# SPEC Review Report: SPEC-CODEMAPS-REFRESH-001

Iteration: 1/2 (Tier M ceiling per `harness.plan_audit_tier_ceilings`)
Verdict: **FAIL** (score ≥ threshold, but 1 blocking-class consistency defect open — fix is one line in three places; iter-2 re-audit scoped to defect delta)
Overall Score: **0.81** (arithmetic mean of 4 dimensions; harmonic mean 0.80)
Auditor tree: worktree t432 @ ad272be20 (WT-codemaps-refresh), 2026-09-02. Reasoning context from the SPEC author ignored per M1 Context Isolation.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: REQ-CMR-001..008 at spec.md:L80-L94, strictly sequential list items, no gaps, no duplicates, uniform `- REQ-CMR-` form.
- [PASS] MP-2 EARS/GEARS format compliance (requirement layer in spec.md §C): REQ-CMR-001 Ubiquitous "the regeneration shall produce" (L80); REQ-CMR-002 compound While+When "the accuracy verification shall test" (L82); REQ-CMR-003 Ubiquitous (L84); REQ-CMR-004 Ubiquitous (L86); REQ-CMR-005 While + shall-not (L88); REQ-CMR-006 When "the codemaps layer shall report verdict=fresh" (L90); REQ-CMR-007 Unwanted "The executor shall not modify" (L92); REQ-CMR-008 Where (capability gate, premise verified true — gate.yaml:74 `blocking: false`) + shall-not (L94). No legacy IF/THEN pattern (grep `If .*, then the` = 0 hits). AC layer is Given-When-Then by design (M3 § Scope — verification layer, not graded here).
- [PASS] MP-3 YAML frontmatter validity: 12/12 canonical fields present (spec.md:L2-L16) with correct types — `version: "0.1.0"` quoted semver, `created/updated` ISO dates, `priority: P1`, `phase: "v3.2.0 target"` (not a prohibited stage name), `tags` CSV string, no rejected snake_case aliases. Dedicated tool evidence: `moai spec lint .moai/specs/SPEC-CODEMAPS-REFRESH-001/spec.md` → "No findings". Note: `lifecycle: spec-first` (L12) is outside the SSOT enum `spec-anchored|spec-lite|exploratory` — mechanically unenforced (lint checks presence only) and repo-tolerated (population: 642 spec-anchored vs 3 spec-first), recorded as D5 (MINOR), not an MP-3 failure.
- [N/A] MP-4 Section 22 language neutrality: single-language-scoped SPEC (Go-repo codemaps regeneration; no multi-language tooling surface) — auto-passes.
- [PASS] MP-5 D7 cross-SPEC reconciliation: 5 referenced SPECs (related_specs + body) all exist in `.moai/specs/` and all carry `status: completed` — SPEC-GRAPH-FRESHNESS-CADENCE-001, SPEC-STAMP-REACHABILITY-001, SPEC-V3R6-DOCS-CODEMAPS-V3-001, SPEC-V3R6-GRAPH-FRESHNESS-001, SPEC-V3R6-GRAPH-FRESHNESS-002. None retired/superseded/archived → no BLOCKING finding. No reconciliation clause required.
- [PASS] MP-6 D8 cross-platform discipline: literal `syscall` count in spec.md = 0 → D8-4 auto-PASS.
- [PASS] MP-7 clarification gate: `grep -n '\[NEEDS CLARIFICATION'` on plan.md = 0 hits; research.md absent (Tier M does not require it — N/A on that file, plan.md clean ⇒ PASS).

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | minor ambiguity in 1-2 requirements, consistently resolvable | REQ-CMR-004 "bounded 스팟체크" (spec.md:L86) unquantified and inconsistent with AC-CMR-004's full §1-catalog compare (acceptance.md:L52-L54); AC-CMR-002 "행 수가 추출된 인용 경로 수와 일치" (acceptance.md:L38) leaves raw-occurrences-vs-unique-paths open. All other REQ/AC single-interpretation. |
| Completeness | 1.00 | all sections + frontmatter complete | HISTORY (L21), §A Problem, §B Scope with 3 `### Out of Scope — <topic>` H3s each carrying specific `-` bullets (L62-L76), §C GEARS (L78), §D AC summary (L96), §E Cross-References (L110); acceptance.md carries §D.5 edge cases, §D.6 closure gate, §D.7 trace; plan.md carries §C pre-flight and §E self-verification with expected outputs. Baseline claims independently re-measured true (see Baseline Attribution). |
| Testability | 0.75 | one AC measurable only with minor interpretation | AC-CMR-002 row-count clause needs a dedup ruling (D3b). All 7 ACs otherwise binary with command + expected output; AC-CMR-005 correctly anticipates the live trunk-state edge (merge-base == HEAD == ad272be20 in this worktree, measured). |
| Traceability | 0.75 | one REQ without a direct AC | REQ-CMR-001..007 ↔ AC-CMR-001..007 1:1 (acceptance.md §D.7). REQ-CMR-008 maps indirectly to §A principle + §D.6 closure gate — documented (§D.7) and operational, but no numbered AC carries it (D2). |

## Baseline Attribution (auditor's own measurements, this run, this tree @ ad272be20)

| SPEC claim | Verdict | Evidence (command → observed output) |
|---|---|---|
| `moai graph check` codemaps value=60 threshold=40 stale; contribution 13 vs f3e11e113 | TRUE | `moai graph check` → `codemaps metric=described-source-diff value=60 threshold=40 verdict=stale` + `contribution: 13 described-worthy file(s) vs first parent f3e11e113` (rc=1) |
| mx-index/edges verdict=absent in fresh worktree | TRUE | same output → both layers `verdict=absent` |
| `go list ./...` = 137 (AC-CMR-003 baseline) | TRUE | `go list ./... \| wc -l` → `137` |
| phantom 6 dirs absent | TRUE | `test -d` loop → zero EXISTS lines |
| codemaps dir = 7 items (AC-CMR-001 expectation) | TRUE | `ls -1 .moai/project/codemaps/` → 6 docs + provenance.json |
| provenance commit_sha `7fc0af324cf4...`, generated 2026-08-31T06:57:41Z, described_roots [internal cmd pkg], tree_root = t287 worktree | TRUE | `jq` on provenance.json → all four fields match spec §A exactly |
| acceptance.md AC-CMR-001 quotes `7fc0af32cf4...` | **TYPO** | actual sha is `7fc0af324cf47f65...` — acceptance.md dropped the `2` (D4) |
| `--force` regeneration flag exists | TRUE | `.claude/skills/moai/workflows/codemaps.md:36` — `--force (alias --regenerate): Regenerate all codemaps even if they already exist` |
| `moai graph stamp codemaps --commit <rev>` exists | TRUE | `internal/cli/graph_stamp.go` help text documents `--commit` + merge-base recipe (its inline example says `origin/main`; the SPEC's `origin/develop` is the repo-correct variant per local git-flow protocol) |
| gate `graph_freshness.blocking: false`, `codemaps_changed_files: 40` | TRUE | `.moai/config/sections/gate.yaml:72-78` — REQ-CMR-008's Where premise is live |
| merge-base HEAD origin/develop == HEAD (ad272be20) | TRUE | `git merge-base HEAD origin/develop` → `ad272be20abf` — AC-CMR-005's trunk-state exception is the live case and is correctly anticipated |
| described roots clean at audit time | TRUE | `git status --porcelain \| grep -cE '(internal\|cmd\|pkg)/'` → 0 — post-restamp value≈0 expectation is sound |
| SPEC-ID collision-free | TRUE | single `.moai/specs/SPEC-CODEMAPS-REFRESH-001/` directory |

## Defects Found

D1. **Scope-hygiene AC contradicts the SPEC's own mandated progress.md writes** — acceptance.md:L78 (AC-CMR-007), plan.md:L51 (§D), plan.md:L68+L73 (E4) vs plan.md:L86 (M4), acceptance.md:L95 (§D.6) — AC-CMR-007 allowlists only `.moai/project/codemaps/**`, `.moai/reports/t432/**`, `.moai/state/**`, and plan §D says "그 외 경로 변경 금지". But M4 and §D.6 MANDATE writing `progress.md §E.2` (`.moai/specs/SPEC-CODEMAPS-REFRESH-001/progress.md` — outside all three globs), and plan §E's E4 filter `grep -v -E '\.moai/(project/codemaps|reports/t432|state)/'` expects 무출력 — a compliant §E.2 write makes E4 report failure and plan §E loops back ("실패 시 해당 마일스톤으로 되돌아간다"). Second failure direction: if the lane commits before the check, `git status --porcelain` is empty and AC-CMR-007/E4 pass vacuously (empty swept set asserts nothing). Two defect shapes from one root. — Severity: **BLOCKING** — Class: blocking — Required fix: (1) add `.moai/specs/SPEC-CODEMAPS-REFRESH-001/**` to AC-CMR-007's allowed paths and plan §D's 허용 경로; (2) extend E4's filter regex with `specs/SPEC-CODEMAPS-REFRESH-001`; (3) pin the measurement window so it is not vacuous post-commit — judge the lane's change set against base ad272be20 (e.g. `git status --porcelain` pre-commit PLUS `git diff --name-only ad272be20..HEAD`), not a bare clean-tree status.

D2. REQ-CMR-008 has no numbered AC — acceptance.md:§D.7 maps it to §A + §D.6 only — Severity: SHOULD-FIX — Class: blocking (traceability criterion the document's own §D.7 claims to satisfy; one-line fix) — Required fix: reword §D.7's REQ-CMR-008 row to declare AC-CMR-002~004 + §D.6 as its operational verification set explicitly, or add a dedicated AC-CMR-008.

D3. Clarity pair — (a) spec.md:L86 "bounded 스팟체크" is unquantified and inconsistent with AC-CMR-004's full §1-catalog vs `.claude/agents/` listing compare (the AC is stricter than the REQ — which governs?); (b) acceptance.md:L38 "행 수가 추출된 인용 경로 수와 일치" does not state dedup semantics (raw occurrences vs unique paths give different row counts) — Severity: SHOULD-FIX — Class: blocking (internal layer inconsistency + operational-definition gap at verdict time) — Required fix: align REQ-CMR-004 wording with AC-CMR-004 (drop "bounded" or define the bound), and state "경로는 중복 제거한 유니크 기준, 행 수 = 유니크 경로 수".

D4. acceptance.md:L28 SHA quote typo — `7fc0af32cf4...` vs actual `7fc0af324cf47f65...` (provenance.json, jq-verified) — Severity: MINOR — Class: optional — Required fix: correct the truncation to `7fc0af324cf4...`.

D5. spec.md:L12 `lifecycle: spec-first` is outside the SSOT enum (`spec-anchored|spec-lite|exploratory`, spec-frontmatter-schema.md); lint passes it (presence-only enforcement); repo population 642 spec-anchored vs 3 spec-first — Severity: MINOR — Class: optional — Required fix: `lifecycle: spec-anchored`.

D6. progress.md:L18 "Artifact set: ... (Tier M 4종)" — Tier M artifact set is 3 files (spec/plan/acceptance); progress.md is the tracking file, not a tier artifact. Also "플래고" typo (L20) — Severity: MINOR — Class: optional — Required fix: "3종" + 오탈자 수정.

Consistency checklist note: CN-2 (exclusions vs included requirements) FAILS via D1. CN-1 (no contradictory requirements) otherwise clean; CN-3 clean (P1 + tier M consistent with scope).

## Regression Check

Iteration 1 — no prior iteration.

## Recommendation

Verdict is FAIL solely on D1 (with D2/D3 as cheap co-fixes); the score (0.81) already meets the Tier M threshold, so a single revision round closes this. Route to manager-spec:

1. Fix D1 exactly as specified (three spots + measurement-window wording). This is the only blocker.
2. While revising: apply D2 (§D.7 wording) and D3 (REQ-CMR-004 alignment + AC-CMR-002 dedup sentence) — both raise Clarity/Traceability to 1.0 at one-line cost each.
3. D4-D6 are optional cosmetics; fold them in the same pass if convenient, do not iterate on them alone.

Iter-2 re-audit scope (per Retry Loop Contract): the D1-D3 defect delta plus regression check on D1. No from-scratch re-audit needed.

## Skip-eligibility note

Verdict FAIL ⇒ NOT skip-eligible regardless of score (the skip contract requires verdict=PASS AND score ≥ 0.80 AND artifact-hash unchanged). This audit never bypasses Implementation Kickoff Approval either way.
