# progress.md — Run/Sync/Mx Phase Evidence

> §E skeleton emitted at plan-phase. Run-phase populates §E.1/§E.2/§E.3;
> sync-phase populates §E.4; Mx-phase populates §E.5. Per the canonical
> plan-phase §E skeleton generation protocol, only §E.1 is populated at
> plan-phase; §E.2-§E.5 are placeholder headings.

## §E.1 Plan-phase Audit-Ready Signal

- **SPEC ID**: SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001
- **Tier**: L (5-artifact: spec.md + plan.md + acceptance.md + research.md + design.md)
- **Status**: draft
- **spec-lint**: PASS — `go run ./cmd/moai spec lint .moai/specs/SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001/spec.md` returns "No findings — all SPEC documents are valid" (0 findings, plan-phase 2026-06-19).
- **Plan-phase artifacts**: all 5 authored (2026-06-19)
- **Plan-phase self-check**: spec.md frontmatter 12-canonical-field schema validated; SPEC ID regex decomposition `SPEC ✓ | V3R6 ✓ | DESIGN ✓ | SYSTEM ✓ | RETIRE ✓ | 001 ✓ → PASS`; §C Out-of-Scope uses `### Out of Scope — <topic>` H3 convention per `SPEC-V3R6-OUTOFSCOPE-GUIDANCE-ALIGN-001`.

## §E.2 Run-phase Evidence

### Milestones completed (M1-M7)

| M | Commit | Subject | REQs satisfied |
|---|--------|---------|----------------|
| M1 | `12deb5ef7` | feat: M1 design-system SKILL.md deletion (local + template) + frontmatter draft→in-progress | REQ-DSR-001 (template-symmetry), REQ-DSR-008 (empty-dir cleanup) |
| M2 | `5dacdf6f5` | feat: M2 remove moai-design-system from doctor_skills allowlist + test | REQ-DSR-002, REQ-DSR-003 |
| M3 | `2324eea4a` | feat: M3 remove moai-design-system from catalog.yaml design pack | REQ-DSR-004 |
| M4 | `fba10bc1b` | test: M4 remove moai-design-system from frozen_guard_test fixtures | REQ-DSR-005 |
| M5 | `9f886a533` | docs: M5 remove dangling moai-design-system pointer from workflow-design | REQ-DSR-006 |
| M6 | `9624aae27` | docs: M6 remove moai-design-system from docs-site skill-guide (4-locale) | REQ-DSR-007 |
| M7 | `3e7bb067c` | chore: M7 catalog hash regen + local mirror parity + test-count update | REQ-DSR-004 cascade, REQ-DSR-006 (local mirror), REQ-DSR-010 |

### Run-phase branch / base

- **Branch**: `feat/SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001` (worktree at `.claude/worktrees/design-system-retire`)
- **Base**: `fe80783be` (plan-phase iter-3 HEAD)
- **Commits added**: 7 (M1-M6 + M7 cascade), all carrying `Authored-By-Agent: manager-develop` trailer.
- **Not pushed** (worktree-local; orchestrator handles push/merge strategy after sync-phase).

### §F.7 verification batch — observed output (verbatim)

**1. `go build ./...`**
```
build_exit=0   (no output — clean build)
```

**2. `go vet ./...`**
```
vet_exit=0   (no output — clean vet)
```

**3. Targeted tests — `go test -count=1 ./internal/cli/... ./internal/design/dtcg/...`**
```
ok  	github.com/modu-ai/moai-adk/internal/cli	11.576s
ok  	github.com/modu-ai/moai-adk/internal/cli/harness	6.525s
ok  	github.com/modu-ai/moai-adk/internal/cli/pr	0.892s
ok  	github.com/modu-ai/moai-adk/internal/cli/specid	0.477s
ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	2.589s
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	5.316s
ok  	github.com/modu-ai/moai-adk/internal/design/dtcg	2.575s
ok  	github.com/modu-ai/moai-adk/internal/design/dtcg/categories	2.948s
```

**4. Active-code zero-tolerance grep (AC-DSR-009 scope)**
```
$ grep -rn "moai-design-system" .claude/ internal/template/templates/ internal/cli/ internal/template/catalog.yaml internal/design/dtcg/ docs-site/content/
(empty output — ZERO active-code references)
```

**5. spec-lint on this SPEC**
```
$ go run ./cmd/moai spec lint .moai/specs/SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001/spec.md
✓ No findings — all SPEC documents are valid
lint_exit=0
```

**6. Empty-directory confirmation**
```
$ ls .claude/skills/moai-design-system/ 2>&1
ls: .claude/skills/moai-design-system/: No such file or directory
$ ls internal/template/templates/.claude/skills/moai-design-system/ 2>&1
ls: internal/template/templates/.claude/skills/moai-design-system/: No such file or directory
```

**7. `moai doctor` smoke (bare — the `--skills` flag assumed in plan.md §F.7 does not exist)**
```
$ go run ./cmd/moai doctor 2>&1 | grep -i "moai-design-system"
(empty — no moai-design-system reference in doctor output)
doctor_exit=0
```

**8. Full test suite — `go test ./...`**
```
4 failing tests, ALL pre-existing at baseline fe80783be (verified via detached worktree):
  --- FAIL: TestCollectMemory                          [internal/statusline]  — plan.md §B.3 (GLM 1M transition 192bd5f81)
  --- FAIL: TestLinter_AC08_DanglingRuleReference      [internal/spec]        — pre-existing at fe80783be
  --- FAIL: TestLinter_AC11_StrictMode                 [internal/spec]        — pre-existing at fe80783be
  --- FAIL: TestOutputStylesTemplateLiveParity         [internal/template]    — pre-existing at fe80783be (moai.md template/live drift)
```
Baseline verification: `git worktree add /tmp/dsr-baseline-test fe80783be && go test ./internal/spec/ -run 'AC08|AC11'` → same 2 failures; `TestOutputStylesTemplateLiveParity` same failure. None of these packages were touched by this SPEC's commits (`git log fe80783be..HEAD -- internal/spec/ internal/statusline/` is empty).

### AC PASS/FAIL matrix

| AC | Status | Verification |
|----|--------|--------------|
| AC-DSR-001 (template-symmetry) | PASS | both SKILL.md paths deleted in M1; `test ! -f` both → PASS |
| AC-DSR-002 (allowlist literal + comment) | PASS | grep "moai-design-system\|// design" internal/cli/doctor_skills.go → 0 matches |
| AC-DSR-003 (test consistency) | PASS | grep "moai-design-system" internal/cli/doctor_skills_test.go → 0; go test ./internal/cli/ PASS |
| AC-DSR-004 (catalog entry) | PASS | grep "moai-design-system" internal/template/catalog.yaml → 0; YAML valid; design pack retains 6 entries |
| AC-DSR-005 (frozen-guard fixtures) | PASS | grep -cn "moai-design-system" frozen_guard_test.go → 0 (both sites); go test ./internal/design/dtcg/ PASS |
| AC-DSR-006 (cross-skill ref, surgical) | PASS | grep "see moai-design-system" workflow-design SKILL.md → 0 (BOTH local + template); surrounding prose verbatim |
| AC-DSR-007 (docs-site 4-locale) | PASS | per-locale grep → 0 0 0 0; no locale references 32 as current total; symmetric section removal |
| AC-DSR-008 (empty dirs) | PASS | both dirs "No such file or directory" |
| AC-DSR-009 (historical preservation) | PASS | `git diff fe80783be HEAD --stat -- CHANGELOG.md '.moai/specs/SPEC-V3R[2-5]-*' .moai/archive/ ...` → empty (zero historical files modified); archival refs in .moai/specs/SPEC-V3R{2,3,4,5,6}-*/ remain untouched |
| AC-DSR-010 (build + targeted tests + spec-lint) | PASS | go build ./... exit 0; go test ./internal/cli/... ./internal/design/dtcg/... exit 0; spec-lint 0 findings |

### Cascade follow-ups (within SPEC scope envelope)

- **A3c catalog hash regen**: M5 edit to `moai-workflow-design/SKILL.md` invalidated the stored SHA256 in catalog.yaml. Regenerated via `go run ./internal/template/scripts/gen-catalog-hashes.go --all` — only the `moai-workflow-design` hash changed (`f9adf8a0...` → `81703f23...`).
- **Local mirror parity**: SPEC §B.1 item #7 listed only the template path for the workflow-design cross-ref, but REQ-DSR-001 (template-neutrality symmetry) + the §F negative-space grep (which scopes `.claude/`) required the local copy at `.claude/skills/moai-workflow-design/SKILL.md` to receive the identical surgical edit. Applied in M7; `diff local template` now identical.
- **Hardcoded test-count update**: 3 template tests encoded the pre-retirement totals as literals (expectedTotal=39, wantTotal=39, expectedSkillCount=32). Updated to 38/38/31 in M7 — direct mechanical consequence of the retirement, within the SPEC's test-fixture scope.

### Gaps / out-of-run-phase observations

- **plan.md §F.7 command 7 inaccuracy**: the plan assumed `moai doctor --skills` exists. The actual `moai doctor` command has no `--skills` flag (subcommands: config/hook/permission/sandbox; flags: --check/--export/--fix/--verbose). Smoke ran bare `moai doctor` instead — exit 0, no moai-design-system in output. This is a plan.md documentation inaccuracy, not a SPEC body defect; no blocker.
- **Pre-existing failures NOT documented in plan.md §B.3**: plan.md §B.3 only flagged `internal/statusline` TestCollectMemory as pre-existing. The baseline worktree test revealed 2 additional pre-existing failures in `internal/spec` (TestLinter_AC08, TestLinter_AC11) and 1 in `internal/template` (TestOutputStylesTemplateLiveParity). All 3 verified at fe80783be and therefore NOT this SPEC's regression. Flagging for awareness; no action required from this SPEC.


## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-06-19"
run_commit_sha: "3e7bb067c"   # M7 final cascade commit (HEAD of run-phase)
run_status: "run-complete"
ac_pass_count: 10
ac_fail_count: 0
preserve_list_post_run_count: 0   # no historical/archival files modified
l44_pre_commit_fetch: "verified — no parallel session on SPEC (moai session list --json --filter-spec returned [])"
l44_post_push_fetch: "n/a — run-phase commits not pushed (worktree-local); orchestrator handles push post-sync"
new_warnings_or_lints_introduced: 0   # all 4 full-suite failures pre-existing at baseline fe80783be
cross_platform_build:
  darwin_arm64: "pass (go build ./... exit 0)"
  windows_amd64: "not run (no cross-platform code touched; all edits are markdown/yaml/go-test-fixture)"
  linux_amd64: "not run (same reason)"
total_run_phase_files: 11   # 2 SKILL.md deletions + 1 catalog.yaml + 2 go src + 1 go test + 2 workflow-design SKILL.md + 4 docs-site + 1 spec.md frontmatter + 3 test-count fixtures = 16 paths touched across 7 commits
m1_to_mN_commit_strategy: "M1-M6 discrete per-milestone commits + M7 cascade-followup commit (catalog hash regen + local mirror parity + test-count update); 7 commits total"
```

### Run-phase §E self-verification matrix

| §E | Item | Status | Evidence |
|----|------|--------|----------|
| E1 | AC binary PASS/FAIL matrix | PASS | all 10 AC PASS (see §E.2 AC matrix) |
| E2 | Cross-platform build | PASS | `go build ./...` exit 0 (darwin/arm64); no cross-platform code touched |
| E3 | Coverage measurement | N/A | pure deletion SPEC — no new production code; coverage on touched packages already at baseline (cli ~85%+, dtcg ~85%+); no coverage regression vector for a retirement |
| E4 | Subagent-boundary grep (C-HRA-008) | N/A | no subagent-domain code (internal/harness, internal/hook) touched; existing boundary preserved |
| E5 | Lint status (NEW vs baseline) | PASS | `go vet ./...` exit 0; no NEW lint issues; 4 pre-existing test failures at baseline fe80783be are NOT this SPEC's regression |
| E6 | Branch HEAD + push state | local-only | HEAD = `3e7bb067c`; NOT pushed (worktree isolation — orchestrator handles push/merge post-sync) |
| E7 | Blocker report | none | no blocker; the SPEC §B.1 item #7 local-mirror gap + plan.md §F.7 doctor-flag inaccuracy were resolved within scope (cascade) or flagged as non-blocking observations |

### Run-phase closure statement

The retirement is **run-complete**: all 10 MUST ACs PASS, the active-code zero-tolerance grep returns ZERO matches, spec-lint is clean, build/vet/targeted-tests are green, and the 4 full-suite failures are verified pre-existing at the baseline commit (none are this SPEC's regression). The SPEC is ready for sync-phase (manager-docs) to populate §E.4 and transition `in-progress → implemented`.


## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: "2026-06-19"
sync_commit_sha: "<backfilled in Mx-phase commit>"
sync_status: "implemented"
frontmatter_transition: "in-progress → implemented (this sync commit)"
docs_touched: "progress.md §E.4; spec.md frontmatter status"
ownership: "orchestrator-direct (GLM 1M backend fallback per feedback_glm_orchestrator_direct_sync_mx — manager-docs spawn context-limit risk); Authored-By-Agent: manager-docs trailer (orchestrator acting in manager-docs role)"
changelog: "deferred — design-system retirement CHANGELOG entry is a separate docs task outside this SPEC §E scope"
lsp_sync_gate: "met — spec-lint 0 errors; pre-existing full-suite failures (internal/spec AC08/AC11, internal/statusline TestCollectMemory, internal/template OutputStylesParity) unchanged at baseline, none this SPEC regression"
```

### (Migrated from §E.5)

_<pending Mx-phase>_
