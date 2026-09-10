# SPEC Review Report: SPEC-SPECLINT-ARTIFACT-STATUS-001

Iteration: 1/3 (Tier S — harness.yaml `plan_audit_tier_ceilings.S = 1` is the SSOT ceiling: single-pass audit; the dispatch's stated ceiling of 3 is superseded by the SSOT)
Verdict: FAIL
Overall Score: 0.81 (mean of dimensions; harmonic 0.80 — both clear the Tier S 0.75 threshold, but the M5 must-pass firewall overrides: MP-3 FAILED)

Auditor: plan-auditor (independent, fresh-judgment). Reasoning context ignored per M1 Context Isolation; only the SPEC artifacts + this tree were read.
Tree: `.claude/worktrees/t490` @ `615d18c1f`, branch `WT-speclint-status-transition`. SPEC documents unmodified by this audit; no commits made.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: REQ-001..REQ-006 at spec.md:L38-L43 — sequential, uniformly zero-padded, no gaps, no duplicates.
- [PASS] MP-2 EARS/GEARS format compliance (judged against the REQUIREMENT layer, spec.md §2 only; the §3 Given-When-Then ACs are the verification layer and are graded under Group 4, not here): all six REQ entries match GEARS patterns — REQ-001 (L38) and REQ-004 (L41) Event-driven ("When …, the lint engine shall …"); REQ-002 (L39) and REQ-005 (L42) Ubiquitous ("The … shall …"); REQ-003 (L40) Unwanted ("shall not modify"); REQ-006 (L43) Ubiquitous. No informal modality, no test scenario posing as a REQ.
- [FAIL] MP-3 YAML frontmatter validity: 12/12 canonical fields present (spec.md:L2-L14) with correct types for eleven of them (`version: "0.1.0"` quoted semver; `created`/`updated` ISO dates; `priority: P1`; `phase: "v3.2.0 target"` quoted release-target, not a prohibited stage name; `tags` CSV string; optional `tier: S` is a valid optional field). **`lifecycle: spec-first` (spec.md:L12) is not a member of the SSOT enum** `spec-anchored|spec-lite|exploratory` (`.claude/rules/moai/development/spec-frontmatter-schema.md` § Field Reference — the MP-3 SSOT). The Go lint is presence-only (`internal/spec/lint.go:996` iterates `{"lifecycle", fm.Lifecycle}` for missing/empty), so the lint does not flag it — re-verified this run: scoped lint on this spec.md prints "✓ No findings". MP-3 binds to the SSOT, not the lint's detection surface; an enum value outside the enum is a type mismatch = FAIL.
- [N/A] MP-4 Section 22 language neutrality: single-purpose repair SPEC scoped to this repo's own spec-lint CI and two markdown artifacts; no multi-language tooling covered. Auto-pass.
- [PASS] MP-5 D7 cross-SPEC reconciliation: references extracted — `SPEC-CODEX-E2E-MEASURE-001` (status: completed), `SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001` (status: completed), self. Both referenced SPECs exist; neither is retired/superseded/archived; no BLOCKING finding.
- [PASS] MP-6 D8 cross-platform discipline: `syscall` count in spec.md = 0 → auto-PASS.
- [PASS] MP-7 clarification gate: `grep -rn 'NEEDS CLARIFICATION'` over the SPEC directory = 0 matches. No `research.md` at Tier S (N/A for that file per the MP-4 precedent).

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 — minor ambiguity in one or two items, resolvable consistently | AC-SCOPE's second clause (spec.md:L50) "the full-branch diff touches no other file with the fix" is ambiguous — the full-branch diff will legitimately carry other files (baseline, evidence, SPEC artifacts); REQ-002's field enumeration (L39) omits `tier` (see D3/D4) |
| Completeness | 0.75 | 0.75 — one non-critical section missing; frontmatter complete | Sections present: Background (WHY), Requirements, Acceptance Criteria, Constraints, Out of Scope with two `### Out of Scope — <topic>` H3 headings + specific bullets (spec.md:L61, L66). HISTORY absent (low-harm on a fresh v0.1.0 draft); HOW legitimately lives in plan.md at Tier S (spec.md:L55) |
| Testability | 1.0 | 1.0 — every AC binary-testable, zero weasel words | All five ACs judge counts/diffs/logs: AC-GREEN (grep `^ERROR` = 0, L47), AC-MUTANT (exactly 2 `^ERROR` lines, both `ArtifactStatusFieldForbidden`, L48), AC-ORDERING (git log ordering, L49), AC-SCOPE (exact two-line diff, L50), AC-EVIDENCE (files + commit-message grep, L51). Weasel-word scan = 0 matches. The exit-code trap is defused on FOUR surfaces: REQ-006 (L43), AC-GREEN (L47), plan.md §E (L45-46), plan.md §G (L75) |
| Traceability | 0.75 | 0.75 — one AC with no parent requirement; mapping indirect | Every REQ has ≥1 AC (REQ-001→AC-GREEN, REQ-002/003→AC-SCOPE, REQ-004→AC-MUTANT, REQ-005→AC-ORDERING, REQ-006→AC-GREEN + plan §E). But AC-EVIDENCE (L51) carries two obligations (evidence export; `t490` commit-message carrier) with NO parent REQ in §2 — they live only in plan.md §D (L41) / §E (L54-55). No AC cites its REQ by ID (semantic naming only) |

## Defects Found (structured defect-list)

D1. MP3-LIFECYCLE-ENUM — spec.md:L12 — `lifecycle: spec-first` violates the SSOT enum (`spec-anchored|spec-lite|exploratory`, spec-frontmatter-schema.md § Field Reference). The lint cannot see it (presence-only check, internal/spec/lint.go:996) — the defect is invisible to the exact tool this SPEC repairs, which is precisely why the audit layer must catch it. Material irony: a SPEC whose subject is frontmatter schema discipline carries its own frontmatter schema violation of the same class the SPEC's cited doctrine names ("a value the schema does not permit", § Non-transition frontmatter corrections). — Severity: critical — Class: blocking — Required fix: change to `lifecycle: spec-anchored` (tree-dominant value 671/776; template default). One word.

D2. TRACE-ORPHAN-AC-EVIDENCE — spec.md:L51 — AC-EVIDENCE's two clauses (GREEN+mutant outputs exported under `.moai/reports/t490/`; every branch commit message carries `t490`) have no parent REQ; the requirement layer under-specifies what the verification layer demands. — Severity: minor — Class: blocking — Required fix: add REQ-007 covering evidence export + commit-message traceability (or fold both clauses into REQ-005).

D3. REQ2-ENUM-INCOMPLETE — spec.md:L39 — the parenthetical enumerates surviving fields as (id, title, version, created, updated, author) but both target files also carry `tier: M` (verified this run in the files' frontmatter) — enumeration not exhaustive. Low harm: the binding clause "every other frontmatter field … shall remain byte-identical" covers `tier`, and AC-SCOPE's diff test would catch any change. — Severity: minor — Class: optional — Required fix: append `tier` to the enumeration or delete the parenthetical.

D4. ACSCOPE-CLAUSE2-WORDING — spec.md:L50 — "the full-branch diff touches no other file with the fix" reads literally as a property of the whole branch diff, which will legitimately contain other files; a literal tester could fail a correct implementation. — Severity: minor — Class: optional — Required fix: reword to "the commits carrying the two-line repair touch no file other than the two named artifacts".

D5. PLAN-SPEC-DIR-COMMIT-GAP — plan.md §F (L63-L71) — no milestone commits the SPEC's own plan-phase artifacts; `.moai/specs/SPEC-SPECLINT-ARTIFACT-STATUS-001/` is untracked (git status, this run), and on this card's git-flow protocol the WT branch must carry the work before the develop merge — untracked artifacts never reach develop and are lost at worktree disposal. M1/M4 commit only red-baseline.md and evidence outputs. — Severity: minor — Class: optional — Required fix: extend M4 (or add M5) to commit the SPEC directory with explicit pathspec + `t490` in the message.

No other defects found. The mutant direction, ordering attribution, exit-code discipline, and Tier S proportionality are all soundly specified (see verification annex).

## Regression Check (Iteration 2+ only)

N/A — iteration 1.

## Recommendation

1. Fix D1 (one word): spec.md:L12 → `lifecycle: spec-anchored`.
2. Fix D2: add REQ-007 to §2 so AC-EVIDENCE has a parent requirement (evidence export under `.moai/reports/t490/` + `t490` in every branch commit message).
3. Same pass, one line each: D3 (append `tier` to REQ-002's enumeration or drop the parenthetical) and D4 (reword AC-SCOPE clause 2).
4. Fix D5: add the SPEC-directory commit step to plan.md §F (explicit pathspec, `t490` in the message).
5. After the frontmatter fix, re-run `go run ./cmd/moai spec lint .moai/specs/SPEC-SPECLINT-ARTIFACT-STATUS-001/spec.md` — it must remain "✓ No findings" (it will; the change moves INTO the enum).
6. Verdict routing: Tier S audit ceiling is 1 (harness.yaml SSOT), so this iteration-1 FAIL routes to the orchestrator — apply the fixes above, then obtain an explicit user-approved re-audit or acceptance decision per the retry contract. The defect set is four one-line edits; nothing structural.

## Measured Premise Re-Verification (this audit run — VCI §3 annex)

**Claim.** The SPEC's factual premises are true; its verification design is executable as written; its own artifacts comply with the statelessness rule it enforces (except D1).

**Evidence.** Commands run and verbatim outputs, this session:
- `go run ./cmd/moai spec lint --strict` (full tree) → exit 1; `grep -c '^ERROR'` = **2**; `grep -c '^WARNING'` = **4,333**; both ERROR lines are `ArtifactStatusFieldForbidden` at `SPEC-CODEX-E2E-MEASURE-001/plan.md` line 5 and `acceptance.md` line 5; StatusTransitionInvalid = **99**; MissingExclusions = **25**. Full output: `/tmp/t490-audit-full.txt` (4,340 lines).
- The lint message prescribes the repair verbatim: "Remove the `status:` line" — confirming spec.md:L32's claim.
- `git rev-parse HEAD:.moai/specs/SPEC-CODEX-E2E-MEASURE-001/plan.md` = `15d5ef6ae478b97fc301f9c778e472aa7bd6c8d8`; `…/acceptance.md` = `13d830137cd34b4fcb3616a92f0099e4880f2d54` — blob pins (spec.md:L56, plan.md:L33) exact.
- `git log -1 185ce7d57` → "docs(SPEC-CODEX-E2E-MEASURE-001): sync-phase artifacts — 3-phase close (card t462)"; its diff shows `-status: in-progress` / `+status: completed` in both target files — root-cause attribution (spec.md:L28) verified.
- `git log -1 ce779f9ee` → 2026-05-16, "feat(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 — 77건 status drift 일괄 해소 + terminal-state exemption (#939)" — historical warning source verified.
- `.github/workflows/spec-lint.yml` "Run SPEC lint" step: `run: go run ./cmd/moai spec lint --strict` — claimed CI invocation verified.
- `internal/cli/spec_lint.go` → `if report.HasErrors() { return &exitCodeError{code: 1, …} }` — fails on ERROR findings only; the SPEC's exit-code claim and the REQ-006/AC-GREEN discipline rest on verified code.
- `go run ./cmd/moai spec lint .moai/specs/SPEC-SPECLINT-ARTIFACT-STATUS-001/spec.md` → "✓ No findings" — lane's scoped measurement reproduced.
- Full-tree lint contains **0** occurrences of `SPECLINT-ARTIFACT-STATUS` — the new SPEC directory is clean, including progress.md (its §E.1 body yaml `status: draft` is body content, not frontmatter; progress.md sits outside the statelessness sub-section by schema design).
- Self-compliance with the enforced rule: this SPEC's plan.md frontmatter carries NO `status:` field; progress.md has no frontmatter; spec.md alone carries `status: draft`. Compliant on the status axis — the sole violation is D1 (a different field, invisible to the lint).

**Baseline-attribution.** All measurements this audit run, tree `615d18c1f` (worktree `.claude/worktrees/t490`), tool built from tree source via `go run` (same basis as CI, VCI §2.2). Independent of the lane's baseline — the lane's figures (2/4,333/99/25/48-ce779f9ee-citations) reproduced exactly.

**Gaps.** CI remote re-judgment not observed (lead's batch-push owns that read, matching red-baseline.md Gaps). The 4,333 warnings not exhaustively classified — out of scope by SPEC design and immaterial to this verdict. Go toolchain minor-version parity with CI not separately confirmed (residual carried in red-baseline.md).

**Residual-risk.** After the 2 errors are removed, future SPEC intake can reintroduce ERROR findings — the verdict and the GREEN state are tree-scoped, a boundary the SPEC's own baseline already declares.

## Verdict Rationale

The SPEC is well-built for its size: the measured defect is real (independently reproduced), every premise traces to a verified artifact, the anti-vacuous mutant is correctly specified with a tree-ish SHA and its exactness is guaranteed by AC-SCOPE, the exit-code trap is defused on four surfaces, and the two-file scope is proportionate to Tier S (6 REQ / 5 AC, within the 8/8 budget; no inflation, no load-bearing under-specification beyond D2/D5). It fails on exactly one must-pass criterion: its own frontmatter carries `lifecycle: spec-first`, a value outside the SSOT enum — a one-word repair, but a hard firewall failure, and one the SPEC's own subject domain makes material rather than pedantic.

---

# Iteration 2 — DELTA CONFIRMATION (exception record)

> **Tier S 상한 1 소진 후 운영자 승인으로 1회 예외, 델타 확인 한정** — one operator-approved exception after the Tier S ceiling (harness.yaml `plan_audit_tier_ceilings.S = 1`) was exhausted; scope bound to delta confirmation of the five iteration-1 findings plus regression over the touched regions. NOT a fresh full audit: the iteration-1 independent verifications (red-baseline reproduction, blob pins, root-cause commit, exit-code mechanism) stand unopened.

Iteration: 2 (exception — no third spawn exists; if this re-audit had FAILED, the card would go to an acceptance decision. It did not — the bound is recorded but moot.)
Verdict: PASS
Overall Score: 0.94 (mean of dimensions 1.0/0.75/1.0/1.0; harmonic 0.92 — clears the Tier S 0.75 threshold; iteration-2 score > iteration-1 score 0.81, no STOP signal)

## Per-Finding Resolution (verified in the documents, not the claims)

- **D1 RESOLVED** — spec.md:L12 now `lifecycle: spec-anchored`, a member of the canonical enum `spec-anchored|spec-lite|exploratory` (SSOT Field Reference; operator reference constant). MP-3 now PASSES: 12/12 fields present, types correct, enum conformant, optional `tier: S` valid.
- **D2 RESOLVED** — new REQ-007 at spec.md:L44 ("The repair shall export the GREEN and mutant lint outputs under `.moai/reports/t490/`, and every branch commit message on `WT-speclint-status-transition` shall carry `t490`") — Ubiquitous GEARS form, compliant with MP-2. AC-EVIDENCE (spec.md:L52) now carries "(parent: REQ-007)". Every REQ now has ≥1 AC and every AC has a parent; Traceability 0.75 → 1.0.
- **D3 RESOLVED** — REQ-002 (spec.md:L39) enumeration now exhaustive: (`id`, `title`, `version`, `created`, `updated`, `author`, `tier`) — matches the actual surviving frontmatter of both target files (verified in iteration 1).
- **D4 RESOLVED** — AC-SCOPE clause 2 (spec.md:L51) now reads "the commits carrying the two-line repair modify no file other than the two named artifacts" — a decidable, binary-testable predicate. Bonus consistency: the reword composes with the new M5 (the M5 SPEC-directory commit is not "the commit carrying the two-line repair", so no contradiction — under the old full-branch-diff wording, M5 would have contradicted AC-SCOPE).
- **D5 RESOLVED** — plan.md §F M5 (L72-L75) commits the SPEC directory with explicit pathspec and `t490` in the message, with the git-flow rationale stated.

## Regression Check (touched regions)

- REQ numbering: REQ-001..REQ-007 sequential, no gaps, no duplicates; 7 ≤ Tier S ceiling of 8. MP-1 PASS.
- Frontmatter: all other fields unchanged from the iteration-1 read; only L12 changed. MP-3 PASS (see D1).
- Scoped lint re-run on the edited spec.md (this iteration, this tree): "✓ No findings" — the edits broke nothing.
- REQ-007 and M5 introduce no new SPEC-ID references (D7 surface unchanged), no `syscall` (D8 unchanged), no clarification markers (MP-7 unchanged). MP-4/5/6/7 stand from iteration 1.
- Score regression check: 0.94 > 0.81 — no STOP signal; tier threshold cleared.

## New Findings

D6. UPDATED-FIELD-STALE — spec.md:L7 — `updated: 2026-09-06` is one day stale: file mtimes are 2026-09-07 00:13 (both artifacts, this run) and the SSOT's Non-transition frontmatter corrections procedure has the repair touch `updated:` alongside the corrected field. — Severity: minor — Class: optional — Required fix: bump to `2026-09-07`. Does not affect any AC, REQ, or must-pass criterion; does not affect the verdict.

## Verdict Rationale (Iteration 2)

All five iteration-1 findings are resolved in the documents exactly as specified; the touched regions carry no regression (numbering, GEARS form, lint cleanliness, cross-SPEC/platform/clarification surfaces all re-verified or standing); one new minor optional finding (D6) is surfaced for the orchestrator's discretion. The SPEC now passes every must-pass criterion and scores 0.94 against the Tier S 0.75 threshold. This PASS closes the plan-phase audit for card t490; Implementation Kickoff Approval (plan→run HUMAN GATE) remains mandatory and is never auto-bypassed by this verdict.
