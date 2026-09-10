# Progress — SPEC-GIT-DELIVERY-PROCEDURE-001

> Plan-phase skeleton. §E.2–§E.4 are populated by manager-develop (run-phase) and manager-docs (sync-phase); this agent emits only §E.1.

## §E.1 Plan-phase Audit-Ready Signal

- **Plan-phase artifacts emitted**: spec.md, plan.md, acceptance.md, progress.md (this file). Revision 0.1.3 after the operator decisions on B1-B3 (0.1.2 committed as `880c0c702`).
- **Tier**: M (unchanged — counts exceed the Tier M ceiling, reported as T1). **Era**: V3R6. **Status**: `draft`.
- **Card**: t622. **Tree measured**: HEAD `880c0c702`; scope files, `zone-registry.md` L/T, `internal/constitution`, and `internal/cli/constitution.go` have no diff against `b412f8a33b9f82ec5f85ccb5eeb960ef125dd8c0` (`git diff --stat` empty, exit 0), so plan-time counts on template copies are base counts.
- **Pre-write self-checks executed**:
  - SPEC ID regex `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` on `SPEC-GIT-DELIVERY-PROCEDURE-001` → `PASS` (0.1.0 and 0.1.2).
  - Frontmatter: 12 canonical fields present; optional `tier`, plus `era` and `related_specs`.
- **Decisions applied (2026-09-10, operator via lead)**: OD-1 = option 1; OD-2 = option B (C first, replaced after the class D / deprecate_after v3.1.0 record); B1 = one PR per SPEC on `feat/SPEC-*`; B2 = the primary-checkout feature-branch path moves to the worktree flow; B3 = amend CONST-V3R5-027 and 028 through `moai constitution amend` (operator owns HumanOversight). None of these is Implementation Kickoff Approval.
- **Requirement / criterion status**: REQ 23 (active 21: 001-010, 013-023; withdrawn 2: 011, 012). AC 24 (active 22: 001-010, 013-024; withdrawn 2: 011, 012). Tier M ceilings 16 each — both exceeded (spec.md §C.5 T1).
- **Open items before run-phase (spec.md §C.5)**: B4 amend apply step is a stub (`internal/constitution/pipeline.go:256-266`, pinned by `pipeline_test.go`), so B3 cannot write the amendment on this tree; B6 `worktree-integration.md` 45 and 543-556 contradict B1/B2 (Frozen 545 and 556 unregistered); T1 Tier reclassification.
- **Amend source read** (no amend executed, per lead constraint): `moai constitution amend --help` (installed `v3.2.0-rc.5`, `84fa4ece4`) flags `--rule --before --after --evidence --dry-run`; `internal/cli/constitution.go:528-529` exact `--before` match; `:143-153` registry resolution (env → CLAUDE_PROJECT_DIR → cwd) versus `pipeline.go:66` fixed cwd registry; gates in `frozen_guard.go:22`, `canary.go:16/18/38`, `contradiction.go:87/110/124`, `rate_limiter.go:10/12/92`, `human_oversight.go:32/44-45`. Registry contains no amend section (`grep` 0); its only change procedure is § Retiring an Entry (lines 39-58). Canary-eligible SPEC dirs (`^SPEC-[A-Z0-9]+$`) = 0. `CLAUDE_PROJECT_DIR` and `MOAI_CONSTITUTION_REGISTRY` unset in this session (`printenv` exit 1).
- **Base counts and positive controls (0.1.3, template copies)**:
  - AC-GDP-001: Synchronization section paragraphs with fetch and rev-list 1, auto-fail 1.
  - AC-GDP-002: Pre-Spawn block 9 lines, rev-list 1; interpretation table rows 8.
  - AC-GDP-006: delivery Step 3.4 section 52 lines, hits at 8 and 20; doc-execution subsection 9 lines, hit at 8; `manager-[g]it[.]md` 0 and 0.
  - AC-GDP-007: command detector 20 (10/5/1/4).
  - AC-GDP-009: manager-git section 43 lines (`moai cc -w` 1, `push -u origin` 1, `gh pr create` 1); spec-workflow block 15, spec-assembly section 10, delivery Step 3.3.5 11 — markers 0. Prefix tokens: spec-workflow chore 1 / feat 5 / plan 3 / sync 1; manager-git section feat; spec-assembly section feat 1; delivery chore 1 / sync 1 → union differs from `{feat/SPEC-}`. Fixture with only `feat/SPEC-XXX` → diff exit 0; plus one `sync/SPEC-XXX` line → diff exit 1.
  - AC-GDP-010: `worktree add` control hit at `main-checkout-branch-guard.md:38`.
  - AC-GDP-017 (multi-PR detector): 19 hits (spec-workflow 26, 42, 43, 44, 47, 50, 53, 65, 66, 319, 320, 336, 360, 361, 428, 433, 438, 439; delivery 50). Mutant fixture 6 lines → lines 1, 2, 3, 4, 6 hit; single-PR line 5 not hit.
  - AC-GDP-018: github-flow section 30 lines; `**Feature branch** (any branch other than main)` 1; launcher token 0; `primary checkout` 0; `**Worktree context**` 1; `stop and report` 1. Mutant (PR branch in the primary checkout + `moai cc -w`) → counts 0/1/1 pass, prose detector hits line 3.
  - AC-GDP-019: Step 3.3.5 section `ExitWorktree` 0; command detector 3; `return to (the) base branch` 2; merge-landed marker 0. Mutant (`git switch` + `ExitWorktree` + `MERGED`) → command detector 1 → FAIL.
  - AC-GDP-020/023: registry L/T `diff` exit 0; before-clause lines 1 in each copy (027, 028); proposed after-clauses 0 in registry L/T and spec-workflow L/T.
  - AC-GDP-021 (prose detector): 11 hits (spec-workflow 21, 42, 43, 50, 51, 52, 192, 286, 321, 429; delivery 321). Mutant fixture 6 lines → lines 1-4 hit, line 5 ("repository root") missed as documented limit, line 6 (Route A) not hit.
  - AC-GDP-022 (installed build, scratch project copies): fixture registry with the 027 after-clause and unchanged source → exit 1, `status: drift`, `drift_count: 1`, CONST-V3R5-027 DRIFT; unmodified copy → exit 0, `status: ok`, `drift_count: 0`, `retired_count: 4`; registry outside the project dir → exit 1, "escapes project dir", `status: ""`.
  - AC-GDP-024: stdin-injection detector on a 5-line fixture → lines 2-5 hit (pipe from printf, pipe from yes, here-string, file redirect), dry-run line 1 not hit.
- **Iteration-2 fixes kept** (N1-N6) as recorded in 0.1.2; AC-GDP-015 token extractor fixture still 5 letter tokens / 2 digit tokens.
- **Post-write verification (0.1.3)**:
  - AC-GDP-009 base union on template copies = `chore/SPEC-`, `feat/SPEC-`, `plan/SPEC-`, `sync/SPEC-` → diff against `{feat/SPEC-}` exit 1 (red).
  - AC-GDP-015 controls on revised spec.md: SPEC-ID lines 11, REQ lines 35, date lines 16; SHA tokens 14 (84fa4ece4 x2, 880c0c702 x1, 980ccdc56 x1, b412f8a33 x6, b7447cb90 x3, c352330d3 x1).
  - AC-GDP-020 mutant (`--before` without " at this step") → exact-before count 0, exit 1.
  - AC-GDP-022 judge on recorded JSON: `"status": "ok"` count 0 for the out-of-project refusal and 0 for the DRIFT fixture, 1 for the unmodified copy.
  - AC-GDP-023 mutant (fixture registry carrying the 027 after-clause vs template registry) → diff exit 1; after-clause count fixture 1, template 0.
  - Run evidence absent: `test -e .moai/reports/t622/run/const-amend-commands.txt` exit 1.
  - REQ trace: 23 REQ IDs in spec.md equal the REQ IDs cited by the AC table (diff exit 0).
- **plan_status**: audit-ready (with open items B4, B6, T1 for the lead)

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
