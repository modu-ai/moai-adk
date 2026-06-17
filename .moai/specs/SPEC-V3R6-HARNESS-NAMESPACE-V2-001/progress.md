# Progress — SPEC-V3R6-HARNESS-NAMESPACE-V2-001

> Run-phase complete: 3 namespace commits cherry-picked onto main + Layer B rename committed + catalog regenerated. RE-VERIFIED with real exit codes on main (correcting the previous inaccurate "green except" claim).

## §E.1 Plan-phase Audit-Ready Signal

plan-auditor verdict: PASS-WITH-DEBT 0.82. Debt D2 (layer5.go:121,192 emission) resolved at M2. D1/D3/D4 → manager-docs sync-phase (framing/label/citation).

## §E.2 Run-phase Evidence

### Previous inaccurate claim — corrected

The earlier §E claimed "AC-HNS-011 PASS — go test ./... green (except 2 pre-existing unrelated failures)" and "go test ./... green except 2 pre-existing unrelated failures (... template TestLateBranchTemplateMirror/SKILL.md byte-drift ...)". This was a **self-contradictory claim against a HARD AC**: "green except failures" is not green, and the "TestLateBranchTemplateMirror/SKILL.md" test name does not exist — `TestLateBranchTemplateMirror` covers `spec-assembly.md` (LATE-BRANCH allowlist), NOT `moai/SKILL.md`. The moai/SKILL.md source-vs-mirror byte delta is INTENTIONAL per CLAUDE.local.md §21 (dev-only `release-update` command stripped from template mirror), explicitly removed from the byte-parity allowlist at `rule_template_mirror_test.go` L86-91. Re-verification was performed against the REAL test suite.

### AC matrix — re-stated against genuine evidence on main

| AC | Status | Evidence |
|----|--------|----------|
| AC-HNS-001 | PASS | All 5 surfaces migrated and landed on main: M1 `660da7410` (Go enforcement + dual recognition), M2 `93e4ff6c4` (generator emission incl layer5.go), M3+M4 `fefc5fb87` (CI sentinel + test fixtures), Layer B `8665a0164` (harness-moaiadk-* rename). Namespace-scoped test packages (`internal/cli`, `internal/harness`, `internal/harness/safety`, `internal/design/pipeline`, `internal/template`) all PASS — `ok` 5/5. |
| AC-HNS-002 | PASS | TestIsUserOwnedNamespace_HarnessV2DualRecognition: harness-foo=true, moai-harness-bar=false (substring separation verified) |
| AC-HNS-003 | PASS | internal/harness/layer5.go emits `harness-*` ONLY (grep `my-harness` layer5.go = 0 matches); 0 new func additions |
| AC-HNS-004 | PASS | TestIsUserOwnedNamespace_HarnessV2DualRecognition: my-harness-legacy-skill=true (REQ-HNS-005 backward-compat) |
| AC-HNS-005 | PASS | TestIsUserOwnedNamespace_HarnessV2DualRecognition: moai-harness-learner=false (no policy reversal) |
| AC-HNS-006 | PASS | grep -rn "my-harness" moai-meta-harness/ meta-harness.md = 0; TestTemplateNoInternalContentLeak PASS |
| AC-HNS-007 | PASS | TestNamespaceLeakHarnessSkills PASS (sentinel NAMESPACE_LEAK_HARNESS_SKILL) |
| AC-HNS-008 | PASS | doctor_harness.go checkLayer1Triggers recognizes harness-* (M1 dual recognition); doctor_skills.go classifySkill returns INFO for harness-* |
| AC-HNS-009 | PASS | prefix_conflict.go DetectPrefixConflicts collects harness-* + my-harness-* (dual); suffix strip handles both |
| AC-HNS-010 | PASS | grep -rn "moai-builder" internal/ .claude/ (excl agent-memory) = 0 |
| AC-HNS-011 | PASS-WITH-DEBT | REAL exit codes on main after all fixes: `go build ./...` → exit 0; `go test ./...` → exit 1; `golangci-lint run` → 0 issues. The test exit=1 is caused SOLELY by `internal/statusline` TestCollectMemory / TestCollectMemory_AutoCompactScaling (hardcoded 200000 TokenBudget expectation vs production 1000000 1M-budget policy) — **pre-existing on bef24877d (main before my commits), NOT caused by namespace migration**. All namespace-scoped + harness-scoped + template-mirror + catalog-hash tests PASS. The 2 catalog-hash failures introduced by my M2 cherry-pick conflict resolution (stale moai-meta-harness hash) were RESOLVED by `gen-catalog-hashes.go --all` (commit `340fe6e46`). Debt: the pre-existing statusline 1M-budget test-vs-policy drift is out of this SPEC's scope (recommend a separate chore SPEC to update memory_test.go expectations to the 1M model-budget reality). |
| AC-HNS-012 | PASS | golangci-lint run --timeout=2m = 0 issues; go vet ./... = exit 0 |

### Milestone summary (on main, post-cherry-pick)
- M1 (Go enforcement + dual recognition): DONE — cherry-picked as `660da7410` (7 files)
- M2 (generator emission incl layer5.go:121,192): DONE — cherry-picked as `93e4ff6c4` (3 files after catalog.yaml conflict resolved; D2 debt resolved)
- M3+M4 (CI sentinel + test fixtures): DONE — cherry-picked as `fefc5fb87` (28 files)
- Catalog regen (cherry-pick conflict fallout): DONE — `340fe6e46` (moai-meta-harness hash refresh, NOT a SKILL.md drift fix)
- Layer B (harness-moaiadk-* rename): DONE — `8665a0164` (6 files: 2 skill dirs + 4 specialist agent skill: refs)
- M5 (atomic integration verify): DONE — all 5 surfaces verified on main with real exit codes

### Layer B local skill decision
Option (a) chosen and committed: renamed `.claude/skills/my-harness-moaiadk-{best-practices,patterns}` → `harness-moaiadk-*` + updated 4 specialist agents (`.claude/agents/harness/*-specialist.md`) skills: refs (all 4 point to `harness-moaiadk-patterns` + `harness-moaiadk-best-practices`, verified zero residual `my-harness-moaiadk-*` references). Rationale: namespace consistency with canonical harness-* recognition. These are dev-local (NOT template-distributed) so template-neutrality guard is unaffected.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-06-18
run_commit_sha: 8665a0164 (main tip after cherry-pick + catalog regen + Layer B rename; origin/main push pending test_exit=0 confirmation — see Push Decision below)
run_status: PASS-WITH-DEBT (namespace scope genuinely green; 1 pre-existing unrelated statusline test failure documented)
ac_pass_count: 11
ac_pass_with_debt_count: 1 (AC-HNS-011 — pre-existing statusline 1M-budget failure out of SPEC scope)
ac_fail_count: 0 (no AC in this SPEC's scope fails)
preserve_list_post_run_count: 37 (intentional my-harness-* legacy dual-recognition refs in Go enforcement code — REQ-HNS-005 deprecation window; every ref is either a comment explaining the dual-recognition contract or a HasPrefix("my-harness-") guard recognizing BOTH harness-* canonical AND my-harness-* legacy)
l44_pre_commit_fetch: 0 0 (clean — local == origin/main at pre-spawn)
l44_post_push_fetch: pending push decision (see Push Decision)
new_warnings_or_lints_introduced: 0 (golangci-lint clean, go vet clean)
cross_platform_build.linux_darwin: exit 0 (go build ./...)
cross_platform_build.windows: pending explicit GOOS=windows check (go build ./... exit 0 on darwin; windows build not re-run this session)
total_run_phase_files: 39 (7 M1 + 3 M2 + 28 M3/M4) + 6 Layer B + 1 catalog regen
m1_to_mN_commit_strategy: 3 cherry-picked commits (M1 `660da7410` / M2 `93e4ff6c4` / M3+M4 `fefc5fb87`) + 1 catalog-regen chore `340fe6e46` + 1 Layer B rename `8665a0164` = 5 commits on main

### Push Decision (NOT pushed — surfaced to orchestrator)

Hybrid Trunk 1-person OSS (§23): the orchestrator's constraint was "Push to origin/main only AFTER test_exit=0 is confirmed". `test_exit=1` is the current state due to the pre-existing `internal/statusline` TestCollectMemory failure (confirmed present on `bef24877d` before my commits, caused by a 1M-budget policy change that predates this SPEC). Since (a) this failure is genuinely pre-existing and unrelated to namespace migration, (b) all namespace-scoped tests PASS, and (c) `go build`/`golangci-lint`/`go vet` are all green — push is defensible BUT the orchestrator's literal gate (`test_exit=0`) is not met. The 5 commits are NOT pushed; the orchestrator decides whether to (a) push as-is given the failure is pre-existing, (b) fix the statusline test first in a separate chore SPEC then push both, or (c) hold the push.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase (manager-docs)>_

## §E.5 Mx-phase Audit-Ready Signal

_<pending Mx-phase>_
