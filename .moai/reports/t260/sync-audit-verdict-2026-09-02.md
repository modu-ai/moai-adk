# Sync-Phase Audit Verdict — SPEC-LEARN-CHANNEL-SCOPE-001 (card t260)

- **Auditor**: sync-auditor (independent, skeptical)
- **Date**: 2026-09-02
- **Tree under audit**: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t260`, branch `WT-learn-channel-gap`, HEAD `2ae50c59f`, fork `d7ce6c6bd` (= origin/develop at dispatch)
- **Chain**: `d1044d9d2` (M1) → `efe39e914` (M2) → `0ea123429` (M3) → `3a6db9f16` (run SHA backfill) → `af526f8c9` (sync close) → `2ae50c59f` (sync SHA backfill)
- **Overall Verdict**: **PASS** (harmonic 92.9 / 100; must-pass firewall Functionality+Security both PASS)
- **Scope**: docs-only SPEC — 11 changed files, all markdown, zero Go / zero `.moai/config/sections` (measured: `git diff --name-status d7ce6c6bd..HEAD`). No test suite applies; substance adjudicated on close integrity, docs accuracy, and neutrality.

---

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim

The SPEC's sync close is integral (single-commit merged 3-phase close, honest §E.4, D3 backfill), its CHANGELOG entry is accurate and count-policy-compliant, the template mirror is neutral and minimally divergent from local, and the run's `feat(...)` commit subjects are ownership-lint-conformant rather than a defect.

### Evidence (all commands run in this audit, this tree, HEAD `2ae50c59f` unless noted)

**Item 1 — `feat(...)` commit subjects on a docs-only SPEC: ACCEPT.**

- `internal/spec/lint_ownership.go:16-21` — the lifecycle matrix comment assigns run-phase M{N} commits the canonical subject pattern `feat|fix|refactor|perf|test(SPEC-...)`. The type prefix encodes the OWNING agent/phase, not the content domain. Row 3 assigns `docs(SPEC-...)` to manager-docs (sync phase).
- `internal/spec/lint_ownership.go:112-139` (`commitOwnerKind`) — `docs(` → `ownerManagerDocs`; `feat(` without `plan-phase`/`supersedes` → `ownerManagerDevelop`. A `docs(...)` M1 subject would have MIS-classified ownership in the fallback path.
- `internal/spec/lint_ownership.go:108-111` — subject-prefix classification is retired from the production `Check()` path (the `Authored-By-Agent:` trailer is the WHO SSOT). Measured: **all 6 branch commits carry no trailer** (`git log --format='%(trailers:key=Authored-By-Agent,valueonly)'` → empty ×6), so the F13 subject-prefix fallback IS the live classification path for this SPEC — the run's reasoning was not cosmetic; it was load-bearing and correct.
- `af526f8c9` subject `docs(SPEC-LEARN-CHANNEL-SCOPE-001): sync-phase artifacts — 3-phase close (card t260)` conforms to the merged-close matrix row and satisfies the close-subject full-ID mandate (full SPEC-ID + canonical "3-phase close" infix; `internal/spec/transitions.go` `closeInfixMatch` accepts both infixes per D4).
- Backfill commits `3a6db9f16` / `2ae50c59f` use `chore(...)` and touch only `progress.md` (no `status:` transition → outside `OwnershipTransitionRule`'s scope entirely).
- Adjudication: **accept as-is; no subject rewrite before develop integration.** There is no lint violation to repair — the current subjects are the conformant choice and `docs(...)` on run-phase M1 would have been the defect. A history rewrite would repair nothing.

**Item 2 — Close-transition integrity: PASS.**

- `git show af526f8c9 -- .moai/specs/SPEC-LEARN-CHANNEL-SCOPE-001/spec.md` → exactly 1 line changed: `-status: in-progress` / `+status: completed`. Nothing else in spec.md. `updated: 2026-09-02` was already current (value unchanged — honestly recorded in §E.4).
- The sync commit carries all 5 sync artifacts in one commit (`--stat`: plan-audit-iter1, plan-audit-iter2, progress.md, spec.md, CHANGELOG.md) — the merged 3-phase close (`implemented` never a separate step), matching the matrix row "this same sync commit carries the completed transition + the 3-phase close".
- D3 backfill: `git show 2ae50c59f` → exactly 1 line: `sync_commit_sha: "pending-backfill-sync"` → `"af526f8c9"`. The placeholder-then-backfill pattern is the schema-documented self-referential-hazard workaround, executed exactly as specified.

**Item 3 — CHANGELOG accuracy + count policy: PASS.**

- (a) Anchor doc `.moai/docs/learning-channel-scope.md` exists (82 lines); carries the bounded claim (§ "The bounded claim"), the `human-mediated loop` section, and the dated baseline (measured **2026-09-02** @ full SHA `d7ce6c6bd8dcc5f48a9ab46555f52d14e68540d9`) with a re-runnable tally command, the 5,958-row file size, the `test_fail` 0-row observation, and the human-channel yield (165 files / 146 since 2026-08-25). REQ-LCS-003's full content list is present. The doc correctly separates the re-verifiability pair (command + family-set predicate — the durable claim) from the drifting row count.
- (b) Doctrine/skill corrections are real: `human-mediated loop` marker = 1 hit each in local `.claude/rules/moai/core/moai-constitution-detail.md`, the template mirror, and `.claude/skills/hns-lsel-curator/SKILL.md`; `grep -c '인간 매개 루프' CLAUDE.local.md` → 1. The stale "(624 stubs)" prose is gone from SKILL.md; SKILL.md:35-36 now defers to `.moai/docs/learning-channel-scope.md` ("never in prose here"). The two remaining `624` literals (SKILL.md:91-92) are fenced `clusters.json` **schema-example fixture values**, not live-count claims (see F1).
- (c) The entry contains NO dev-local row counts: the only numbers are "7/7 ACs PASS", the 2026-09-02 dated baseline, and "8 changed files in the run diff, all markdown" — the last verified as a diff-shape fact (8 files at `3a6db9f16`, all markdown). The `grep '5,9\|5950'` hits at CHANGELOG lines 94/1084 belong to pre-existing entries of other SPECs. Line 12's `624` match is the entry NARRATING the removal ("carried a stale \"(624 stubs)\" count") — description of a fix, not a live count.
- Position: `[Unreleased]` → `### Added` → first bullet (line 12, immediately after the heading) — house convention held. AC count: `grep -o 'AC-LCS-[0-9]*' spec.md | sort -u` → exactly 7 (AC-LCS-001..007), matching "7/7 ACs PASS". Duplicate guard: exactly 1 occurrence of the SPEC-ID in CHANGELOG.md.

**Item 4 — Neutrality re-check: PASS.**

- Local vs mirror diff = exactly one line, `54d53` (the local-only scope-anchor line naming `.moai/docs/learning-channel-scope.md`) — byte-matches the run's reported single-line divergence. Sanctioned local-only removal.
- Mirror greps: `SPEC-LEARN-CHANNEL-SCOPE` / `t260` / `/Users/goos` / `WT-learn` → 0 hits. The mirror's 2-line delta is the principle-level capability/composition rewrite with no dev-local values (the anchor doc referenced generically, path omitted — see F2).
- `git diff d7ce6c6bd..HEAD --stat -- internal/template/catalog.yaml` → empty (no catalog drift from the `make build` cascade).
- Secret scan of the full range diff (`api[_-]?key|secret|password|PRIVATE`) → 0 hits.

**Item 5 — §E.4 completeness + records: PASS.**

- `progress.md §E.4` carries every required signal field: `sync_complete_at: 2026-09-02`, `sync_commit_sha: "af526f8c9"` (backfilled), `sync_status: complete`, `frontmatter_status_transitions` (honest merged-close record: "implemented는 별도 단계로 존재한 적 없음"; plan.md n/a; progress.md no-frontmatter), `b12_self_test` a/b/c, `mx_scan` (0 tag changes — consistent with the all-markdown diff, independently re-measured), `sync_phase_observations` (including the feat-vs-docs deviation record and the evidence-log absence), `sync_session_gaps`, `readme_review`, `known_residual_docs_drift: []`.
- Both plan-audit reports tracked: `.moai/reports/t260/plan-audit-iter1.md` (123 lines) + `plan-audit-iter2.md` (94 lines), both `A` in `af526f8c9`.
- §I two-cell adoption discipline (verification-completeness.md §2): RED-now table with command / verbatim output / exit / flipping milestone; document-level tree pin `d7ce6c6bd` (a commit SHA, not a branch); AC-LCS-005/006 explicitly declared regression-guards with the stated reason (no content claim can be red on a pre-work tree).

**AC substance (Functionality) — independently re-verified:**

- **AC-LCS-001 re-run in this audit**: the anchor doc's tally command executed against the primary checkout's live inbox returned **every bucket `tool_failure:*` — zero rows outside `{tool_failure, test_fail}`** (22 tool_failure buckets, no test_fail bucket — consistent with the 0-row test_fail observation). The re-verifiability claim holds on fresh evidence.
- AC-LCS-002/003/007 markers verified (counts above); AC-LCS-004 — all three §A.5-enumerated claim surfaces carry the bounded claim or pointer; AC-LCS-005/006 — diff-verified (11 files all markdown; no new config sections, hooks, scripts, or record formats).
- Code-coordinate citations in the docs verified: `internal/hook/failure_observer.go` tool_failure appender (:75-79 region) and test_fail appender (:108-113 region) both append `appendLessonsInboxStub` with the documented event keys; `internal/hook/evidence_writer.go` `rec.IsTestFail` branch present at the cited :583-591 region.

### Baseline-attribution

All measurements: this audit's run, worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t260`, HEAD `2ae50c59f` (branch range `d7ce6c6bd..2ae50c59f`), except: the AC-LCS-001 tally re-run reads the primary checkout's live runtime file `/Users/goos/MoAI/moai-adk-go/.moai/lessons-inbox.jsonl` (the inbox is runtime state outside any worktree — the anchor doc states this split explicitly), executed 2026-09-02 in this audit. The lead's spot-checks were reused only where I independently reproduced them; every load-bearing number above was re-measured, not carried.

### Gaps

- **No CI verdict** — the branch is unpushed; the develop-window push is the verdict surface. Inherited limitation, not a SPEC defect.
- **CHANGELOG numbers are dated-pointer policy** — row counts live in the anchor doc only by design; the CHANGELOG entry is judged on the dated-baseline phrasing, not on carrying counts.
- **Consistency sweep intentionally limited to claim surfaces** — the §A.5 list (constitution-detail, SKILL.md, CLAUDE.local.md §28) per AC-LCS-004's own scope clause; navigator mirrors, fixtures, code, and config surfaces are untouched by design.
- **Row-count freshness** — I verified the family-set predicate (the durable claim form per the anchor doc's own doctrine) but did not re-measure the 5,958 row count; the count is declared drift-prone append-only state, so no reader is obliged to.
- **evidence_writer.go line range** — the `rec.IsTestFail` branch was verified by reading the :580-592 region (present, shape matches); the exact :583-591 boundary is approximate by inspection, not byte-diffed.

### Residual-risk

- F1/F2 below (surface ambiguities) could mislead a careful reader of SKILL.md's schema example or the mirror's anchor reference; neither falsifies any AC.
- The ownership-lint conformance of the subjects rests on the F13 fallback path (no trailers on any commit). If a future harness change re-weights subject classification or adds trailers, the classification changes — the subjects remain conventional either way.
- The `test_fail` family remains a wired capability with zero observed rows; any future claim that it "captures" test failures should keep citing it as capability-not-composition (the anchor doc already does).

---

## Findings (structured defect-list)

- **F1** [MINOR] [optional] `.claude/skills/hns-lsel-curator/SKILL.md:91-92` — the fenced `clusters.json` schema example carries `offset_after/total_read: 624` and a `tool_failure`-only candidate example. Illustrative fixture values inside a fenced JSON block, not live-count claims (AC-LCS-003's "no live count in prose" is satisfied — the stale PROSE count is gone); but a reader could mistake the example's shape for a composition claim one day. Required fix: none blocking. Optional: an "(illustrative values)" note beside the example.
- **F2** [MINOR] [optional] `internal/template/templates/.claude/rules/moai/core/moai-constitution-detail.md` (mirror, new line 53) — the principle-level line references "the learning-channel scope anchor document", an artifact distributed users' projects do not carry (the path-bearing pointer is the sanctioned local-only line). The sentence still reads correctly as doctrine (a user has no measured composition to point at), but the reference names a file absent from their tree. Required fix: none blocking. Optional: reword to "a scope anchor document (maintainer repositories)" in the mirror only.
- **F3** [OBS] [optional] `internal/hook/failure_observer.go:110` — the function-header comment still says test_fail events go "to usage-log.jsonl". This is the known stale comment the SPEC's constraint-1/§G explicitly left un-adopted (the kickoff-conditional 1-line exception was not taken). Recorded so later sweeps do not re-flag it as a defect of this SPEC — it is a declared residual, in-scope by design.
- **F4** [OBS] [optional] (dispatch-level, not a SPEC artifact) — the lead's dispatch summary said "~5,950 live rows"; the anchor doc records 5,958. The SPEC artifacts are internally consistent; the drift is in the relayed summary only.

## Recommendations

- None blocking. Optional F1/F2 one-line follow-ups can ride any future touch of those files; do not open a repair round for them (each repair round plants a new defect).
- On develop integration, the window push supplies the CI verdict this audit could not.

---

## Dimension Scores

| Dimension | Weight | Score | Verdict | Evidence |
|-----------|--------|-------|---------|----------|
| Functionality | 40% | 95/100 | PASS | All 7 ACs' substance independently re-verified (markers, mirror parity, diff shape, tally re-run, §E.4 fields); two-cell RED-now adoption followed; plan-audit iter-2 PASS-WITH-DEBT 0.86 with debt documented |
| Security | 25% | 96/100 | PASS | Docs-only diff; secret scan 0 hits; template neutrality clean (no SPEC id / card id / dev-local paths / internal counts in mirror); zero mechanism change (AC-LCS-005/006 diff-verified) |
| Craft | 20% | 88/100 | PASS | §E.4 signal-complete incl. honest gaps; anchor doc's re-verifiability-pair doctrine is strong measurement design; deductions: F1/F2 surface ambiguities, self-declared evidence-log absence |
| Consistency | 15% | 93/100 | PASS | Commit subjects lint-conformant (verified against `lint_ownership.go` incl. the no-trailer fallback path); CHANGELOG house position/style/canary documented; D3 backfill pattern exact; mirror/local divergence = exactly the sanctioned line |

**Harmonic mean: 92.9 / 100.** Must-pass firewall (Functionality + Security): both PASS independently.

## Per-item dispositions

| # | Item | Disposition |
|---|------|-------------|
| 1 | `feat(...)` subjects on docs-only SPEC | **ACCEPT** — lint-conformant (matrix row 2 + `commitOwnerKind` mapping); no-trailer fallback makes the choice load-bearing and correct; rewrite rejected (nothing to repair) |
| 2 | Close-transition integrity | **PASS** — single-commit merged close, spec.md 1-line status change only, D3 backfill exact |
| 3 | CHANGELOG accuracy + count policy | **PASS** — all three claims verified against the tree; dated-baseline phrasing held; position + 7/7 AC count + dedup verified |
| 4 | Neutrality | **PASS** — mirror diff = exactly `54d53`; neutrality greps 0; catalog.yaml unchanged |
| 5 | §E.4 completeness | **PASS** — all signal fields present; both plan-audit reports tracked; MX claim consistent with the diff |

Defects beyond the five: F1-F2 (MINOR/optional), F3-F4 (observations) — none blocking, none AC-affecting.
