# SPEC-FIX-LOOPMAP-001 — Progress

> Plan-phase artifact. Run-phase evidence (§E.2, §E.3) and sync-phase evidence (§E.4) are populated by manager-develop and manager-docs respectively. This file is parser-load-bearing: `internal/spec/era.go` `hasAnyProgressMarker` greps for the literal `§E.2` / `§E.3` / `§E.4` substrings — the headings below MUST be preserved verbatim.

---

## §E.1 Plan-phase Audit-Ready Signal

```
plan_status: audit-ready
plan_complete_at: 2026-07-09
tier: S
relation: follow-up to SPEC-LOOP-VERDICT-CONTRACT-001 §Out of Scope L5 deferral (not an Epic Workflow-Reflex member SPEC)
artifacts: 3 (spec.md + plan.md + progress.md; AC inline in spec.md §3, Tier S)
gears_requirements: 6
acceptance_criteria: 10
out_of_scope_topics: 4
audit_findings_traced: L5 (SPEC-LOOP-VERDICT-CONTRACT-001 provenance)
open_decision_points: D1 (loop.md landing-order gate mechanics), D2 (exit_kind enum extension surface), D3 (Phase 4.7 placement)
spec_id_self_check: PASS (SPEC-FIX-LOOPMAP-001 → ^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$)
```

**Plan-phase self-verification**:
- [x] All GEARS requirements use valid patterns (Ubiquitous / When / Where; no IF/THEN).
- [x] All gap claims cite measured source + observed pattern (vci §2; measured 2026-07-09 — fix.md Phase 4 text, loop.md headings, template mirror diff all re-verified via Bash/Read).
- [x] §Out of Scope has 4 `### Out of Scope — <topic>` H3 sub-headings with `-` bullets.
- [x] 12 canonical frontmatter fields + era + tier + depends_on; no snake_case aliases.
- [x] Agentless fixed-pipeline preservation stated as HARD constraint + dedicated regression AC (AC-FLM-010).
- [x] loop.md shared-surface landing-order dependency on SPEC-LOOP-VERDICT-CONTRACT-001 explicitly recorded (Constraint 3, §D D1).
- [x] Deliverable classification: doctrine/skill-doc only; no Go loader; no new config keys.

---

## §E.2 Run-phase Evidence

All commands run 2026-07-09 in the isolated run-phase worktree (base `083cf52cf`, branch `worktree-agent-a735f06961a00ed77`), after the M1-M3 edits to fix.md / loop.md + template mirrors.

### AC Matrix (spec.md §3, 10 rows)

| AC | Status | Verification Command | Actual Output |
|----|--------|----------------------|---------------|
| AC-FLM-001 | PASS | `grep -A 15 '^## Phase 4: Verification' .claude/skills/moai/workflows/fix.md \| grep -c "claim/evidence\|Claim/Evidence\|verification-claim-integrity"` | `3` — claim/evidence table language + `verification-claim-integrity.md` §1.1/§3 cross-reference within 15 lines of the heading |
| AC-FLM-002 | PASS | derived check (plan-audit debt): `grep -c "^- Confirm fixes resolved the targeted issues$\|^- Detect any regressions introduced by fixes$" .claude/skills/moai/workflows/fix.md` | `0` — both pre-existing bare-prose bullets removed; replaced by Steps 1-3 evidence-bearing instructions |
| AC-FLM-003 | PASS | `sed -n '/^## Phase 4: Verification/,/^## Phase 4.5/p' fix.md \| grep -ci baseline` / `... grep -ci regression` / `... grep -ci "FULL target scope\|full-rescan\|full re-scan"` | `3` / `4` / `2` — full re-scan instruction + explicit baseline diff co-occur in Phase 4 |
| AC-FLM-004 | PASS | `grep -n "reverted\|reported as failed\|Silent acceptance of a regression is prohibited" fix.md` | line 209: Step 3 names both outcomes — "(a) reverted ... or (b) explicitly reported as failed ... Silent acceptance of a regression is prohibited" |
| AC-FLM-005 | PASS | `grep -n 'loop-verdict\|one-shot-residue' .claude/skills/moai/workflows/fix.md` | 4 matches (lines 268, 270, 272, 352) — Phase 4.7 instructs `.moai/state/loop-verdict-<id>.json` write with `exit_kind: "one-shot-residue"` + `iterations_used: 1` |
| AC-FLM-006 | PASS | `grep -c "Re-entry requires explicit user re-invocation" fix.md` → `1`; `grep -n "SHALL NOT auto-invoke" fix.md` → line 274 | Repeatability clause present exactly once (unchanged, not duplicated); `/moai loop` recommended as suggestion only, auto-invoke prohibited |
| AC-FLM-007 | PASS | `grep -c "^## Loop Taxonomy Position" fix.md` → `1`; awk span → `9` lines; per-quadrant grep → turn-based 1 / goal-based 1 / time-based 1 / proactive 1; 3-axis grep → `3`; sibling path refs → `1`+ | exactly one H2, span 9 ≤ 15, all 4 quadrants + how-it-starts/ends/fits + cadence-bridge.md & moai-workflow-ci-loop path refs |
| AC-FLM-008 | PASS | landing-order gate: `git log --oneline --all -- .claude/skills/moai/workflows/loop.md` → `e0298c230` + `083cf52cf` (loop.md rewrite landed at base, §D D1 OR-branch satisfied); same section checks on loop.md → count `1`, span `9`, all 4 quadrants, 3 axes | pure addition — `git diff loop.md` shows ONLY the new section lines (Steps 1/1.5/9, § Completion Conditions untouched) |
| AC-FLM-009 | PASS | `diff` live vs template for fix.md → exit 0; loop.md → exit 0; `grep -rn "SPEC-FIX-LOOPMAP\|SPEC-LOOP-VERDICT" internal/template/templates/.claude/skills/moai/workflows/{fix,loop}.md` | byte-identical both files; neutrality grep exit 1 (0 matches) |
| AC-FLM-010 | PASS | `go test -run TestAgentlessUtilityNoLLMControlFlow -v ./internal/template/` | `--- PASS: TestAgentlessUtilityNoLLMControlFlow` (fix.md / mx.md / codemaps.md / clean.md subtests all PASS) — AC maps to spec.md Constraint 1 (verified: constraint text names this test + AC-FLM-010) |

### Invariant / preservation checks

| Invariant | Command | Result |
|-----------|---------|--------|
| Phase 3 static dispatch table untouched | `grep -n "Level 1 (import, formatting): manager-develop" fix.md` (+ Level 2/3 rows) | lines 187-189 present verbatim |
| no-op exit + fail-fast wording preserved | `grep -n "status no-op and exit code 0 / Fail-fast" fix.md` | lines 54-55 verbatim |
| No AskUserQuestion additions | `git diff fix.md loop.md \| grep "^+" \| grep -c AskUserQuestion` | `0` |
| `make build` | exit 0 (catalog.yaml unchanged after build — `git status --porcelain internal/template/catalog.yaml` empty) | PASS |
| `go build ./...` | exit 0 | PASS |
| `go test ./internal/template/...` | 1 pre-existing failure: `TestOutputStylesTemplateLiveParity` (`moai-easy.md exists only in template, not in live tree`) — proven pre-existing by stashing this SPEC's 4 file edits and re-running: identical failure at base state; template copy committed at `d7e53fcb8` (pre-base), live copy never tracked. NOT introduced by this SPEC (PRESERVE). All other tests + TestAgentlessUtilityNoLLMControlFlow PASS. | PASS-WITH-PRESERVED-BASE-DEFECT |

## §E.3 Run-phase Audit-Ready Signal

```
run_complete_at: 2026-07-09
run_commit_sha: <M1-M3 commit on branch worktree-agent-a735f06961a00ed77 — orchestrator integrates into main; backfill on integration>
run_status: audit-ready
ac_pass_count: 10
ac_fail_count: 0
preserve_list_post_run_count: 0 violations (moai.md / ralph.yaml / workflow.yaml / gate.md / review.md untouched; loop.md Steps 1/1.5/9 + § Completion Conditions untouched)
l44_pre_commit_fetch: origin/main fetched at run entry — 0 0 (synced at base 083cf52cf); push deferred to orchestrator integration per delegation contract
l44_post_push_fetch: n/a — no push from run-phase worktree (orchestrator integrates)
new_warnings_or_lints_introduced: 0 (doc-only; TestOutputStylesTemplateLiveParity failure pre-exists at base, proven via stash re-run)
cross_platform_build:
  go_build: exit 0
  make_build: exit 0
total_run_phase_files: 6 (fix.md + loop.md + 2 template mirrors + spec.md frontmatter + progress.md)
m1_to_mN_commit_strategy: commit 1 = plan-phase artifacts (draft, as authored); commit 2 = M1-M3 combined (Tier S single run commit) + draft-to-in-progress transition + §E.2/§E.3 population
```

## §E.4 Sync-phase Audit-Ready Signal

```
sync_complete_at: 2026-07-09
sync_commit_sha: <to be backfilled in follow-up chore commit>
sync_status: audit-ready (3-phase close on single sync commit; completed transition rides this commit)
changelog_entry_position: [Unreleased] > Added (SPEC-FIX-LOOPMAP-001, top of Added — latest sync, above LOOP-VERDICT)
frontmatter_status_transitions:
  spec.md: in-progress -> completed (single sync commit, 3-phase close)
  plan.md: n/a (no YAML frontmatter — "# " header; Tier S 2-artifact plan-phase)
  progress.md: n/a (no YAML frontmatter — "# " header; era.go parses §E.* headings + sync_commit_sha)
  acceptance.md: n/a (absent — Tier S SPEC; acceptance criteria embedded in spec.md body)
b12_self_test_a_pre_emission_grep: PASS (grep 'SPEC-FIX-LOOPMAP-001' CHANGELOG.md -> 0 pre-emit)
b12_self_test_b_ac_count_match: PASS (progress.md §E.3 ac_pass_count: 10; CHANGELOG references 10/10)
b12_self_test_c_file_path_verify: PASS (fix.md + template mirror + spec/plan/progress exist)
deployment_pattern: temp-worktree isolation at origin/main c4bd0bbe7 (shared-checkout live-race avoidance; per feedback_glm_orchestrator_direct_sync_mx + feedback_shared_checkout_concurrent_commit_race)
```

### Sync-phase verification (measured 2026-07-09 by orchestrator-direct, GLM backend)

**SCOPING NOTE**: doctrine/skill-doc primary SPEC (zero Go source, Tier S). Sync-phase verification is doc-scope: spec.md frontmatter transition, era.go parser-load-bearing §E.* heading preservation, live↔template byte-parity of fix.md + loop.md, spec-lint, cross-platform build. Full `go test ./...` deferred to race-free window (shared checkout diverged 2/5 — parallel-session commits; out of scope for doc-primary SPEC).

```
V1 spec.md frontmatter: status in-progress -> completed (single artifact with YAML frontmatter; plan/progress use "# " header, no frontmatter — 3-phase close scope = spec.md only)
V2 era.go §E.* headings: 3 preserved (§E.2, §E.3, §E.4) — parser-load-bearing substrings intact
V3 live↔template byte-parity: fix.md PARITY, loop.md PARITY (both core doctrine pairs at parity — run-phase claim faithful, unlike LOOP-VERDICT's config divergence)
V4 CHANGELOG ordering: SPEC-FIX-LOOPMAP-001 at Added top (above LOOP-VERDICT) — latest-sync-first
V5 moai spec lint spec.md: No findings — all SPEC documents are valid
V6 go build ./...: exit 0 (doc-only, sanity)
```

**artifact-structure note**: this SPEC carries 3 artifacts (spec.md + plan.md + progress.md) — acceptance.md is absent (Tier S 2-artifact plan-phase convention; acceptance criteria embedded in spec.md body). plan.md and progress.md use `# ` markdown headers with no YAML frontmatter, so the 3-phase-close frontmatter flip applies to spec.md alone; era.go lifecycle classification reads spec.md frontmatter status + progress.md §E.* headings + sync_commit_sha.
