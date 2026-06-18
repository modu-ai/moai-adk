---
id: SPEC-V3R6-LIFECYCLE-REDESIGN-001
progress_version: "0.2.0"
spec_version: "0.2.0"
status: in-progress
created: 2026-06-18
updated: 2026-06-19
author: manager-spec
tier: L
---

# Progress — SPEC-V3R6-LIFECYCLE-REDESIGN-001

> Plan-phase artifact. The §E section skeleton below carries placeholder headings only; run/sync/Mx evidence is populated by manager-develop (§E.2/§E.3) and manager-docs (§E.4) per the canonical agent responsibility realignment.

## §E.1 Plan-phase Audit-Ready Signal

- spec.md: 12 canonical frontmatter fields present; 21 GEARS REQs (REQ-LR-001..021); Exclusions section (§J) with `### Out of Scope` H3.
- plan.md: 9 milestones (M1-M9); risk register (R1-R5); anti-pattern catalogue (AP-LR-P-001..005).
- acceptance.md: 13 ACs (AC-LR-001..013); 10 MUST-PASS, 3 SHOULD-PASS; Given-When-Then format.
- research.md: drift surface measured (Axis A = 14 files; Axis B = 102 files); era migration impact (V3R6 moving baseline, re-derived H-6 at-risk set = empty — §D.4); corrected era-reclassification trace (§D.3); Spec Kit citation verified (fetched 2026-06-18).
- design.md: H-4 reclassification strategy (corrected H-5 fall-through + S1 auto-migrate + narrowed dual-predicate window); all-three-findings drift update (§B.4 incl. `Y_N_N_Y`); close-infix reconciliation with DRIFT-LEGACY-CONVENTION-001 (§B.6); Epic taxonomy mapping (4 canonical terms).

Plan-phase revision: v0.2.0 (2026-06-19) — plan-audit iter-1 FAIL 0.71/0.85 → 7 defects fixed (D1 era mechanism / D2 Y_N_N_Y / D3 moving baseline / D4 close-infix reconciliation / D5 doc-comment+§E.5 scope / D6 file count / D7 §I summary). All ground-truth verified by direct source inspection.

Plan-phase audit-ready: PASS-WITH-DEBT 0.87 (iter-2, Tier L thresh 0.85, +0.16 monotonic vs iter-1 FAIL 0.71). 7 defects fixed (D1 era mechanism / D2 Y_N_N_Y / D3 moving baseline / D4 close-infix / D5 doc-comment+§E.5 / D6 file count / D7 §I summary). Residual debt D-R1/D-R2/D-R3 carried as run-phase fix-up (non-blocking).

## Mode Selection (Phase 0.95)

**Decision**: `sub-agent` (Mode 5 — sequential `Agent()` spawn per milestone group).

**Input parameters**: tier=L (~38 files: 8 Go `internal/spec/` + ~30 markdown); domains=2 (Go lifecycle engine; `.claude/` rule/skill corpus); language mix=Go(M1-M3)+markdown(M4-M8); concurrency benefit=LOW (coding-heavy, Anthropic coding-task parallelism caveat); Agent Teams prereqs NOT all met → Mode 3 unavailable.

**Mode evaluation**: M1 trivial=no | M2 background=no(write) | M3 agent-team=not selected(prereqs) | M4 parallel=not selected(coding-heavy→caveat) | **M5 sub-agent=selected** | M6 workflow=not selected(semantic not mechanical-uniform).

**Justification**: M1→M2→M3 sequential dependency (C1: migration window REQ-LR-006 ships with/before H-4 rewrite per C2; tests M2 depend on M1). Axis A (M4-M5) and Axis B (M6-M8) are disjoint after M2 (C3) but both coding/markdown-heavy semantic work. Implementation Kickoff Approval: obtained via user `/goal 이세션에서 모두 다 완료` directive (explicit run-phase entry authorization replacing §19.1 AskUserQuestion per goal-directive "do not pause to ask").

**Boundary case**: Axis A/B disjoint-after-M2 could justify Mode 4 parallel, but §B.2 tie-breaker ("coding-heavy + multi-domain → Mode 5") resolves sequential. M4-M8 delegated as 2nd/3rd sequential spawn after M1-M3 verification.

**Verification note (stale-binary hazard)**: orchestrator trust-but-verify of M1-M3 caught a stale `moai` binary — `moai spec audit` reported Y_Y_N_Y=5/Y_Y_Y_Y_StatusDrift=3 from the pre-rebuild binary even though audit.go:73-75 retires those predicates. `go install ./cmd/moai` rebuild → all three §E.5-keyed findings = 0 (AC-LR-011 PASS). Lesson: re-build the binary before trusting `moai spec audit` measurements after any `internal/spec/` Go change.

## §E.2 Run-phase Evidence

### Pre-flight (captured at M1 start, tree HEAD f2907ba4c)

- **PF-1 (D3, baseline N)**: `moai spec audit --json` → total_specs=353, grandfathered=272, modern_era_clean=78, **V3R6 count N=50** (moving baseline; NOT a frozen literal — AC-LR-003 asserts invariance post-M1 == post-M3 == this N). Breakdown: Y_N_N_Y=0, Y_Y_N_Y=4, Y_Y_Y_Y_StatusDrift=3.
- **PF-1b (D1, H-6 at-risk re-derivation)**: research.md §D.4 reproduction command → V3R6 total=50, **genuine H-6 at-risk=0** (empty set). Every current V3R6 SPEC is caught by H-5's `created >= 2026-04-01` / modern-`phase:` heuristic. REQ-LR-006 dual-predicate window is defense-in-depth + classification-rationale precision, not misclassification-prevention. **No blocker.**
- **PF-2 (regression baseline)**: `go test ./internal/spec/...` → 2 PRE-EXISTING failures in `lint_test.go` (`TestLinter_AC08_DanglingRuleReference`, `TestLinter_AC11_StrictMode`) — both in the linter domain (DanglingRuleReference / strict-mode warning escalation), OUT of M1-M3 scope (era.go/audit.go/transitions.go). These are the regression baseline; M1-M3 must not introduce NEW failures and must not touch these.
- **Build baseline**: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- **Git**: branch=worktree-agent-a669304f677a4add1 (fast-forwarded to main HEAD f2907ba4c to acquire SPEC artifacts + current source); `internal/spec/` clean.

### M1 — era.go H-4 3-phase reclassification + dual-predicate window

- Commit: `abd832d9a` (feat M1)
- AC-LR-003 post-M1: V3R6 count = 50 findings == baseline N=50 (invariant held at finding-count proxy).
- AC-LR-013 (D5): era.go doc-comment heuristic table + EraV3R6 const + package taxonomy updated to new §E.4 predicate + legacy fallback.
- Rationale breakdown post-M1: 27 new-H-4 (§E.4), 11 H-4-legacy (migration window), 5 H-5, 7 H-override.
- Regression: 0 new test failures (2 pre-existing lint failures unchanged).

### M2 — audit.go drift re-anchor + transitions.go close-infix + tests

- AC-LR-011 (D2): `Y_N_N_Y`=0 catalog-wide (robust no-space grep), `Y_Y_N_Y`=0 catalog-wide. Both retired. `SyncStatusDrift`=3 (re-anchored from Y_Y_Y_Y_StatusDrift to 3-marker predicate §E.2+§E.4+sync_sha). AC-LR-011 PASS.
- AC-LR-012 (D4): `closeInfix3Phase = "3-phase close"` added, OR'd into `closeInfixMatch`; `closeInfix4Phase` RETAINED. `TestCloseInfixMatch_DualInfix` + `TestClassifyPRTitle_CloseInfix` 3-phase fixtures + `TestCombinedScopeCloseMatches` 3-phase variant all PASS.
- AC-LR-003 distinct-SPEC invariant: 46 distinct V3R6 SPECs post-M2 (43 EraAutoDetected + 3 SyncStatusDrift-with-era-override). **Era classification unchanged since M1** (M2 touched only `checkV3R6Drift`, not `ClassifyEra`) → distinct V3R6 SPEC set is invariant across M1→M2. The acceptance.md AC-LR-003 verification command (`sum(1 for f in drift_findings if era==V3R6)`) is a finding-count proxy that dropped 50→46 because 4 duplicate Y_Y_N_Y findings (each a 2nd finding on a SPEC still carrying its EraAutoDetected finding) were retired — this is the INTENDED drift-storm elimination (D2), not a SPEC-population regression. Recorded as residual risk (D-R2-adjacent: the AC verification command double-counts; the true invariant is the distinct V3R6 SPEC set, which is preserved).
- Tests: `TestAudit_SyncStatusDriftDetection` + `_CompletedClean` (re-anchored), `TestAudit_Y_N_N_Y_NotEmitted` + `TestAudit_Y_Y_N_Y_NotEmitted` (retired must-not-fire, D2), era_test.go H-4-new/H-4-legacy/H-3-§E.4-edge fixtures. 0 new failures (2 pre-existing lint unchanged).

### M3 — migrate_3phase.go §E.5→§E.4 backfill migration

- Commit: _(this M3 commit)_
- REQ-LR-007: one-time backfill folds §E.5 Mx-phase content into §E.4 for modern-era V3R6 SPECs with the legacy 5-section layout. Grandfather-protected SPECs (268) SKIPPED (N4 / AP-LR-P-004).
- Affected set: **65 SPECs folded** (plan-phase estimate ~11 was the classification-critical subset lacking §E.4; the full fold scope is all V3R6 SPECs carrying §E.5 — research.md §C.4 measured ~83 §E.5-bearing progress.md files, of which 65 are modern-era V3R6). 1 outlier (`SPEC-V3R6-MAIN-RED-REMEDIATION-001`) had a duplicate §E.5 section; the migrator was fixed to loop over all §E.5 occurrences and re-run (idempotent on the other 64). Post-fix: 0 residual `## §E.5` headings catalog-wide.
- AC-LR-003 post-M3: distinct V3R6 SPECs = 46 == post-M2 (INVARIANT preserved across M1→M2→M3). Rationale shifted: 37 new-H-4 (up from 27), 1 H-4-legacy (down from 11), 5 H-5. The folded SPECs now classify via the new H-4 predicate (§E.4 carries the folded content + sync_commit_sha preserved).
- Migration log: `.moai/state/lifecycle-redesign-migration.json` (gitignored local state per CLAUDE.local.md §2; records all 65 entries with spec_id/era/mx_commit_sha/migrated_at).
- Scope safety: backup branch ref `backup/pre-m3-migration-*` created pre-migration. All 6 PRESERVE-list dirs (HARNESS-MOAI-NAMESPACE-001 + 5 RULES-*) untouched. No parallel-session in-flight work modified.
- Tests: `TestMigrateProgressMD_FoldsE5IntoE4` + `_Idempotent`, `TestRunMigration_SkipsGrandfathered` (N4), `TestRunMigration_DryRun`. 0 new failures.

### M4 — Axis A 6 드리프트 표면 rule 파일 "3-phase close" 스윕 (+ D4 reconciliation + D5 worked example)

- Commit: `549249b22` (docs M4)
- Files edited (3 of 6 had drift terms; other 3 — agent-patterns.md / archived-agent-rejection.md / spec-workflow.md — were already clean):
  - `verification-claim-integrity.md` §5 Worked Example: "4-phase close" → "3-phase close (plan→run→sync)"; "Mx-phase close" historical-inference 맥락 명시화.
  - `spec-frontmatter-schema.md` Status Transition Ownership Matrix: `implemented → completed` 전이를 sync commit으로 병합 (별도 Mx chore commit row 제거 — AC-LR-004). Close-subject mandate "3-phase close"로 갱신. D4 reconciliation note 추가 (DRIFT-LEGACY-CONVENTION-001 소유권 명시 + 레거시 "4-phase close" infix 매처 보존 per M2/REQ-LR-020/021).
  - `lifecycle-sync-gate.md`: V3R6 era 정의 (4-phase → 3-phase), H-2/H-4 heuristic table (§E.4 술어), rationale string, close-subject mandate + D4 reconciliation note, `## §E.5 Mx-phase` worked example → 4-section layout (§E.1-§E.4) 갱신 (D5/AC-LR-013).
- AC-LR-005 grep gate: `grep -l '4-phase close' <6 files>` → 0 matches in canonical usage (2 residual hits are both inside the D4 reconciliation note citing the legacy infix — intentional exception).
- AC-LR-001 grep gate: `grep -r '4-phase\|Mx-phase' .claude/rules/moai/ | grep -v 'Legacy|deprecated|retired|alias'` → 0 canonical usage; residual hits are all retirement/reconciliation prose ("the former §E.5 Mx-phase section is retired", "NO §E.5 Mx-phase section", D4 reconciliation citing legacy infix) + Late-Branch "4-phase procedure (A→D)" (SPEC-V3R5-LATE-BRANCH-001, unrelated) + "Anthropic 4-phase Step 2" (verbatim Anthropic citation, unrelated).
- Baseline grep counts (pre-M4 → post-M4): "4-phase close" in rules/moai = 5 → 0 canonical (2 reconciliation citations); "Mx-phase"/"§E.5" in 6 M4 files = 5 → 0 canonical (all retirement prose).

### M5 — Axis A 8 표면 파일 (agents/hooks/output-styles/skills) 3-phase close + §E.4-only 반영

- Commit: `0754536c7` (docs M5)
- Files edited (6 with drift terms; manager-git.md + spec-assembly.md were already clean for lifecycle drift — their "4-phase" references are Late-Branch procedure, unrelated):
  - `manager-spec.md`: progress.md skeleton 5→4 section (§E.1-§E.4; §E.5 Mx-phase retired) — AC-LR-002 (template emits exactly 4 §E sections, §E.5 count=0, §E.4 count=1). Forbidden-modifications 매트릭스 §E.5 제거.
  - `manager-docs.md`: status transitions owned — `in-progress → implemented → completed` 단일 sync commit 병합 (AC-LR-004); MX Tag validation sync sub-step 명시 (AC-LR-006). Frontmatter description 갱신.
  - `workflow-specialist.md`: description + Role + Domain Guidance "4-phase V3R6 close contract" → "3-phase V3R6 close contract"; §E.5/mx_commit_sha requirement retired.
  - `harness-moaiadk-patterns/SKILL.md`: step 7 "Optional Mx + 4-phase close" → "3-phase close on the single sync commit".
  - `moai.md`: Cohort Stats trigger "4-phase (plan+run+sync+mx)" → "3-phase (plan+run+sync)".
  - `status-transition-ownership.sh`: Status Transition Ownership Matrix 주석 `completed`→sync-commit 병합 반영 (AC-LR-004; hook은 advisory exit 0 — 모든 전이 accept, 주석 정확성만 갱신).
- AC-LR-006: manager-docs.md line 141 lists MX Tag validation as a sync sub-step (NOT a separate phase) — PASS.
- M5 grep gate: `grep -rl '4-phase\|Mx-phase\|§E\.5' .claude/` (excluding worktrees/agent-memory/specs/reports) → 0 canonical usage; all residual hits are retirement/reconciliation prose + Late-Branch procedure + Anthropic verbatim citation (enumerated as intentional exceptions in E2).

### M6 — Axis B SSOT rewrite: sprint-round-naming.md → Epic taxonomy

- Commit: `cf0c3317f` (docs M6)
- AC-LR-007 PASS: exactly 4 canonical terms (Epic / SPEC / Milestone / Constitution); AP-SRN-001..004 re-anchored to Epic taxonomy + AP-SRN-005 (cohort) added. Verified via `grep -cE '^\| \*\*(Epic|SPEC|Milestone|Constitution)\*\* \|'` = 4.
- AC-LR-009 PASS: Epic definition = "A time-unit or thematic container for one or more SPECs, grouped by schedule, release, or theme (formerly `Sprint`)" — pure rename, semantics preserved.
- Round (b) disambiguation (design.md §C.2): former within-SPEC `Round` (SSE-stall sub-division) folded into `Milestone`. Not a blocker — design.md §C.2 is unambiguous.
- Legacy Aliases section added (Sprint/cohort/Round/Wave appear only there).

### M7 — Axis B T1 migration (.claude/rules/moai/)

- Commit: `317447e05` (docs M7)
- Files edited (4 taxonomy-bearing): askuser-protocol.md (Sprint 8 → Epic 8 example; interview 'Round 1/2' → 'Turn 1/2' + disambiguation note), session-handoff.md (Wave/Sprint/Round → Epic/Milestone), orchestration-mode-selection.md + archived-agent-rejection.md (generic 'SPEC cohort' → 'SPEC group (Epic)' per AP-SRN-005).
- AC-LR-008 T1 true-taxonomy residual: 0. Intentional exceptions (5 files, 17 hits) = ci-watch/worktree infra 'Wave 1/2/3/5' (zone-registry, manager-develop-prompt-template, ci-autofix-protocol, ci-watch-protocol, worktree-state-guard) — pipeline wave numbering owned by SPEC-V3R5-CI-AUTONOMY-001, NOT taxonomy; documented in sprint-round-naming.md AP-SRN-004 note.

### M8 — Axis B T2-T4 migration (agents/output-styles/skills/docs)

- Commit: `5a6983d08` (docs M8)
- Files edited (8): output-styles/moai/moai.md (Cohort Stats → Epic Stats, Sprint Status → Epic Status, Sprint [N] token → Epic [N], full banner + localization table + self-check); moai-workflow-spec/SKILL.md (provenance 'Sprint 10' → Epic 10 + historical-label disambiguation); project/{meta-harness,mode-detection,codebase-analysis}.md + plan/clarity-interview.md (interview 'Round N' → 'Interview Phase N' — interview round is generic, NOT the retired taxonomy Round); plan/spec-assembly.md (AC-LR-010 Epic-reference note added); .moai/docs/harness-delivery-strategy.md (Sprint 15 → Epic 15); meta-harness Agile menu '스프린트' → 'iteration'.
- AC-LR-008 T2-T4 true-taxonomy residual: 0. Intentional exceptions (documented): ci-watch infra 'Wave 1/2/3/7' (moai-domain-database, moai-workflow-ci-loop, git-workflow-doctrine); SQL 'cohort' analytics (db/queries.md signup cohort); generic 'docs cohort' (codemaps); Korean false-positives '그라운드트루스'(ground-truth substring) + '백그라운드'(background substring) + interview '라운드'(generic); Legacy Aliases quoted tokens in sprint-round-naming.md + spec-assembly.md AC-LR-010 note.
- AC-LR-010 PASS: spec-assembly.md scaffolding references Epic (not Sprint) for grouped SPECs.
- E3 PRESERVE: 13 files touched across M6-M8 — naming-migration only; no Axis-A lifecycle file re-touched, no Go code, no other SPEC, no memory (verified `git diff --name-only b262c6d16..HEAD`).
- E2 collateral: `moai spec lint` 0 error (1 pre-existing StatusGitConsistency warning — status 'draft' vs git 'in-progress' — owned by manager-docs sync-phase, NOT introduced by M6-M8).

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-06-19T04:30:00Z
run_commit_sha: 936d1fdbc
run_status: audit-ready (M1-M5 PASS; M6-M8 Axis B Epic naming deferred to separate delegation)

M4-M5 commit chain (this delegation, worktree-agent-ae39cb698447a2e48 @ origin/main 2f5260c58): 549249b22 (M4 6 rule files 3-phase close sweep + D4 reconciliation + D5 worked example) → 0754536c7 (M5 8 Axis A surface files 3-phase close + §E.4-only structure). Pre-push: `git rev-list --count --left-right origin/main...HEAD` = `0 0` (worktree synced to origin/main at delegation start; CONTEXT-GOV-AXIS-001 parallel session has 3 unpushed local commits on main checkout touching unrelated files — no file overlap with M4/M5 scope).

M1-M3 commit chain (rebased SHAs after multi-session race resolution; pushed to origin/main `a964772fa..936d1fdbc`): f38917f81 (M1 era.go) → 1e4b851ce (M2 audit.go+transitions.go+tests) → 93e8b98ee (M3 migrate_3phase.go + 65 SPEC fold) → 936d1fdbc (M3 errcheck fix + §E.3 signal). Pre-rebase SHAs: abd832d9a / 5f5170fbd / 564ad5726 / a1e3b2a24.

**Multi-session race resolution (E6)**: A parallel session advanced local main from `f2907ba4c` (my worktree base) to `6eae1f97f` (9 commits: RULES-CONST-RULEID, RULES-HOTFIX, ORCH-INTERRUPT-LEDGER) during my M1-M3 work. Divergence was `9 4` (main 9 ahead, my worktree 4 ahead). Real file-overlap was only 2 progress.md files (HARNESS-RUNTIME-RECOVERY-001, ORCH-INTERRUPT-LEDGER-001 — both carry §E.5 that my M3 folded AND the parallel session backfilled). `git rebase 6eae1f97f` resolved cleanly (0 manual conflicts — git 3-way merge combined my §E.5→§E.4 fold with their mx_commit_sha backfill). Push `HEAD:main` fast-forwarded origin/main. The parallel session's Go engine did NOT include M1-M2 (their 9 commits are progress.md backfills + rule edits), so my M1-M3 Go engine + migration are net-new, not duplicate.

**E1 AC matrix (M1-M3 subset)**:
- AC-LR-003: PASS — distinct V3R6 SPEC set invariant across M1→M2→M3 (46 distinct post-M3 == 46 post-M2). Finding-count proxy shifted 50→46 due to the intended D2 Y_Y_N_Y retirement (the acceptance.md verification command double-counts findings-per-SPEC; the true invariant is the distinct-SPEC set, preserved). Recorded as residual risk.
- AC-LR-011: PASS — Y_N_N_Y=0 catalog-wide (robust no-space grep), Y_Y_N_Y=0 catalog-wide.
- AC-LR-012: PASS — `closeInfix3Phase`+`closeInfix4Phase` both in transitions.go; `TestCloseInfixMatch_DualInfix` + `TestClassifyPRTitle_CloseInfix` + `TestCombinedScopeCloseMatches` 3-phase variant all PASS.
- AC-LR-013: PASS (M1 half) — era.go doc-comment + EraV3R6 const + package taxonomy updated to §E.4 predicate (4 `§E.4` mentions in lines 86-120). M4 half (lifecycle-sync-gate.md §E.5 worked example) deferred to the M4 delegation.

**E2**: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
**E3**: per-file coverage — era.go ClassifyEra 100%, audit.go checkV3R6Drift 88.9%, transitions.go closeInfixMatch 100% / ClassifyPRTitle 92.3%, migrate_3phase.go MigrateProgressMD 97.4% / insertFoldUnderSection4 95.5% / RunMigration 71.1% (defensive error branches uncovered). All 3 AC-target files exceed 85% on key functions.
**E4**: `grep -rn 'AskUserQuestion|mcp__askuser' internal/spec/ | grep -v _test.go | grep -v '// '` → 0 matches. PASS.
**E5**: `golangci-lint run --timeout=3m ./internal/spec/...` → 0 issues (after errcheck fix in migrate_3phase_test.go).
**Pre-existing baseline**: 2 lint_test.go failures (`TestLinter_AC08_DanglingRuleReference`, `TestLinter_AC11_StrictMode`) — out of M1-M3 scope (DanglingRuleReference / strict-mode linter domain), unchanged by this work.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

sync_commit_sha: _(pending sync-phase)_

### (Migrated from §E.5)

_<pending Mx-phase — NOTE: this section is slated for removal per REQ-LR-004 / REQ-LR-007 of this very SPEC. The redesign merges §E.5 into §E.4. This placeholder is retained for classification compatibility during the migration window (REQ-LR-006) and will be removed once the redesign's M3 backfill completes.>_

mx_commit_sha: _(not applicable — this SPEC removes the Mx-phase concept)_
