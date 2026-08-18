# Progress — SPEC-V3R6-MOAI-HOME-PATHS-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_complete_at: 2026-08-16T22:18:29+09:00
- plan_status: audit-ready
- plan_audit_verdict: PASS — iteration 2, score 0.81 (Tier M threshold 0.80; report: `.moai/reports/plan-audit/SPEC-V3R6-MOAI-HOME-PATHS-001-review-2.md`)
- audit_history: iter1 FAIL 0.75 (D1–D6) → remediation 2026-08-16 (orchestrator-direct, user-approved) → iter2 PASS 0.81; residual R1 (M4 milestone stale label) fixed in same session

## Decision Point 2 (2026-08-16, post-iteration-2)

Audit iteration 1 FAIL (0.75) remediated via AskUserQuestion round — 3 decisions, all recommended options selected: (1) 8-accessor surface confirmed for D1, (2) orchestrator-direct remediation edit, (3) full D1-D6 scope. Iteration 2 delta re-audit: PASS 0.81, no blocking defects; residual R1 fixed immediately (plan.md M4 row stale label). An independent plan-auditor verdict now exists — spawn re-measurement post-PR #1574 (context-window fix) succeeded: iteration 1 + iteration 2 both completed with artifacts.

## Decision Point 1 (2026-08-16T22:18+09:00)

User APPROVED via AskUserQuestion (option "승인 (권장)"). Audit basis: orchestrator self-check in lieu of independent plan-auditor (3 subagent spawns died on context-window limits: manager-spec ×2, plan-auditor ×1; no independent verdict exists — no audit_score recorded to the routing ledger for this reason). Self-check caught and fixed 1 AC defect (un-anchored predicate, 218→17) and 1 coverage gap (AC-015). Backlog card t66 registered per same approval round: model-profile simplification (GLM session-model inheritance + web effort-map exposure).

## Authoring Record

- 2026-08-16 research.md persisted (workflow run wf_ebb53184-4f0, 4 lenses + synthesizer, 5 agents, 0 errors).
- 2026-08-16 spec.md / plan.md / acceptance.md authored **orchestrator-direct** per user approval (AskUserQuestion, option "제가 직접 작성") after two `manager-spec` subagent delegations failed with context-window-limit API errors (13:04:24Z, 13:05:59Z; both produced zero artifacts). Disputed-site classification (4 sites) resolved by direct line reads; membership finalized at 17 (predicate-hidden site migrate_agency.go:184 added).
- Pending: plan-audit attempt (spawn; fallback per user-approved degradation: orchestrator self-check + user review) → Decision Point 1.

## §F Phase 4 Mode Selection

Input parameters:
- tier: M
- scope: ~21 files (17 migration sites across 14 files + new `internal/paths` package + envkeys.go registration + homedir.go/preference collapse)
- domain count: 1 (Go home-resolution refactor — mechanical migration spanning 7 packages, single semantic domain)
- file language mix: 100% Go (+ test files)
- concurrency benefit: LOW (coding-heavy; ordered milestone dependencies)
- Agent Teams prereqs: N/A (Mode 3 retired)

Mode evaluation:
1. trivial — not selected: multi-file semantic refactor, no single-line scope
2. background — not selected: write-capable implementation work
3. agent-team — RETIRED (never selected)
4. parallel — not selected: coding-heavy per Anthropic's coding-task parallelism caveat; no ≥3-domain research split
5. sub-agent — **selected**: single manager-develop delegation covering M1→M6 with per-milestone Conventional Commits; orchestrator verifies at commit boundaries
6. workflow — not selected: <~30 files; inter-milestone dependencies (M3-M5 all import the M1 package); not a dependency-free uniform transform

Decision: sub-agent
Justification: Tier M coding-heavy refactor with strictly ordered milestones (package creation → env registration → 17-site migration → helper collapse → full verification). Sequential sub-agent delegation keeps every milestone gate verifiable by the orchestrator and satisfies the Tier M Section A-E delegation-template requirement. Implementation Kickoff Approval passed (user selected "승인 · 표준 진행", 2026-08-16) before this spawn.

## §E.2 Run-phase Evidence

Baseline-attribution for every row below: (this run, this tree) — worktree `worktree-moai-home-paths`, base `d6b80a01c`, run-phase HEAD `d50a54ba9` (M5; M6 = this evidence commit). All commands executed 2026-08-17 in the worktree root.

### AC Matrix (AC-MHP-001..016)

| AC | Status | Verification Command | Actual Output (verbatim, this run) |
|----|--------|---------------------|---------------|
| AC-MHP-001 | PASS | `grep -rn '"\.moai"' internal pkg cmd --include='*.go' \| grep -v _test.go \| grep -E 'Join\((home\|homeDir\|r\.homeDir)'` | (empty — 0 matches; pre-migration baseline was exactly 17, this run post-migration) |
| AC-MHP-002 | PASS | `go test ./internal/paths/ -count=1 -run TestMoaiHome_OverrideAbsolute` | `--- PASS` within `ok github.com/modu-ai/moai-adk/internal/paths 0.370s` (full suite ok) |
| AC-MHP-003 | PASS | `go test ./internal/paths/ -count=1 -run TestMoaiHome_OverrideEmptyEqualsUnset` | PASS within `ok ... internal/paths` (empty and unset both resolve `<home>/.moai`) |
| AC-MHP-004 | PASS | `go test ./internal/paths/ -count=1 -run TestMoaiHome_OverrideRelativeDisregarded` | PASS within `ok ... internal/paths` |
| AC-MHP-005 | PASS | `go test ./internal/paths/ -count=1 -run 'TestMoaiHome_FallbackUnderHOME\|TestHome_HOMEEnvWins'` + `go test ./internal/cli/preference/ -count=1 -run TestUserHomeDir_PrefersHOME` | PASS in both packages (`ok ... internal/paths`, `ok ... internal/cli/preference 1.497s`) |
| AC-MHP-006 | PASS | `grep -rn '"MOAI_HOME"' internal pkg cmd --include='*.go' \| grep -v _test.go \| grep -vE 'envkeys\.go\|internal/paths/'` | (empty — 0 matches; literal lives only in envkeys.go:22 and internal/paths/paths.go) |
| AC-MHP-007 | PASS | `go list -deps ./internal/paths` | stdlib packages only (os, path/filepath + runtime deps); sole non-stdlib line is the target package itself; `internal/paths` imports nothing outside stdlib |
| AC-MHP-008 | PASS | `grep -n 'paths\.' internal/hook/pre_tool.go internal/core/project/root.go` | `pre_tool.go:280: if home, err := paths.Home(); err == nil {` / `pre_tool.go:283: if moaiRoot, err := paths.MoaiHome(); err == nil {` / `root.go:42: homeDir, _ := paths.Home()`; behavioral suites `ok ... internal/hook 45.465s`, `ok ... internal/core/project 2.761s` |
| AC-MHP-009 | PASS | `grep -rn 'func userHomeDir' internal/ --include='*.go'` | exactly one match: `internal/cli/homedir.go:18` (thin delegate to paths.Home; preference/cmd.go duplicate deleted) |
| AC-MHP-010 | PASS | `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` | `DARWIN_BUILD_OK` / `WINDOWS_BUILD_OK` (both exit 0) |
| AC-MHP-011 | PASS (environmental footnote ¹) | `go test ./internal/paths/ ./internal/statusline/ ./internal/config/ ./internal/cli/ ./internal/update/ ./internal/kanban/ ./internal/hook/ ./internal/glmcred/ ./internal/cli/preference/ -count=1` | 8/9 `ok` + `internal/cli` FAIL solely on `TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey`; with `MOAI_SKIP_LIVE_CODEX=1` the same cli package returns `ok github.com/modu-ai/moai-adk/internal/cli 252.212s` → 9/9 ok under the CI-equivalent condition (CI has no codex binary → test auto-skips) |
| AC-MHP-012 | PASS | `go test -cover ./internal/paths/` | `ok github.com/modu-ai/moai-adk/internal/paths 1.103s coverage: 100.0% of statements` |
| AC-MHP-013 | N/A(run) | sync-phase doc check | underlying greps already hold: `REQ-MHP-012` present in spec.md (2 mentions) + `related_specs:` carries `SPEC-V3R6-SESSION-HANDOFF-AUTO-001` (spec.md:16); final verification owned by manager-docs |
| AC-MHP-014 | N/A(run) | sync-phase doc check | `.env.glm` shell/Go split hazard documented at spec.md §4; CHANGELOG/follow-up-card emission owned by manager-docs |
| AC-MHP-015 | PASS | `grep -rn 'sync\.Once\|sync\.Mutex\|sync\.Map\|var moaiHome\|cached\|memo' internal/paths --include='*.go' \| grep -v _test.go` | (empty — 0 matches; per-call resolution, REQ-MHP-005) |
| AC-MHP-016 | PASS | `grep -rn 'os\.UserHomeDir()' internal/cli/update.go internal/cli/migrate_agency.go` | (empty — 0 matches; pre-migration baseline was exactly 2 at :827/:711; both now resolve via `paths.Home()`, and `checkpointPath` honours an absolute `MOAI_HOME` override — pinned by `TestMigrateAgency_Checkpoint_MoaiHomeOverride`) |

¹ Environmental footnote (AC-MHP-011): `TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey` is a LIVE integration test driving the real codex binary; its failure mode `decision="" err=<nil>` is an inconclusive live-verdict result the test's own comment anticipates. Verified pre-existing: identical failure reproduced at the base state with all run-phase working-tree changes stashed (tagged stash `mhp-m4-wip-baseline-verify`, SHA `67626b1b4`, applied back and dropped). No code path is shared with this SPEC's diff (the gate's only `.moai` references are path-string filters). CI skips the test (no codex on PATH).

### E8 — RED evidence (verbatim pre-GREEN, M1)

Command: `go test ./internal/paths/ -count=1` — executed against the stub scaffold (signatures returning zero values) BEFORE the GREEN implementation. All 10 test functions failed on assertions:

```
--- FAIL: TestHome_HOMEEnvWins (0.00s)
    paths_test.go:38: Home() = "", want "/var/folders/.../TestHome_HOMEEnvWins482180198/001" (HOME must win)
--- FAIL: TestHome_EmptyHOMEFallsBackToUserHomeDir (0.00s)
    paths_test.go:56: Home() error = <nil>, want os.UserHomeDir()'s error state ($HOME is not defined)
--- FAIL: TestMoaiHome_OverrideAbsolute (0.00s)
    paths_test.go:69: MoaiHome() = "", want "/tmp/moai-home-override" (absolute MOAI_HOME honoured verbatim)
--- FAIL: TestMoaiHome_OverrideEmptyEqualsUnset (0.00s)
    paths_test.go:91: MoaiHome() with MOAI_HOME="" = "", want "<tmp>/.moai" (empty == unset)
    paths_test.go:94: MoaiHome() with MOAI_HOME unset = "", want "<tmp>/.moai"
--- FAIL: TestMoaiHome_OverrideRelativeDisregarded (0.00s)
    paths_test.go:109: MoaiHome() with relative MOAI_HOME = "", want "<tmp>/.moai" (relative value disregarded)
--- FAIL: TestMoaiHome_FallbackUnderHOME (0.00s)
    paths_test.go:126: MoaiHome() = "", want "<tmp>/.moai" (HOME-first fallback)
--- FAIL: TestMoaiHome_ErrorNeverDots (0.00s)
    paths_test.go:146: MoaiHome() = "", want an absolute path, never "."
--- FAIL: TestSubAccessors_OverrideAndFallback (0.00s)
    (16 subtests: StateDir/CacheDir/ReleasesDir/WorktreesDir/ProfilesDir/GlmEnvFile/
     UserSettingsFile/UserConfigSectionsDir × {fallback, override} — all `= "", want <path>`)
--- FAIL: TestSubAccessors_ErrorPassthrough (0.00s)
    paths_test.go:224: <Accessor>() = "", want absolute path (8 accessors)
--- FAIL: TestResolution_PerCall (0.00s)
    paths_test.go:246: MoaiHome() before/after HOME change = "" / "", want different values
```

(`(0.00s)` timings, temp-dir paths, and file:line cites are from the captured run; `<tmp>` abbreviates `/var/folders/...` fixture paths whose random suffixes differ per run. The two subtest-block rows condense 16+8 verbatim assertion lines of identical shape; every listed function failed before GREEN.)

### E5 — Lint (NEW vs baseline)

Command: `golangci-lint run --timeout=5m` → `0 issues.` — identical to the pre-flight baseline (`golangci-lint run --timeout=2m` → `0 issues.` at `d6b80a01c`). NEW issues introduced: 0.

### E4 — Subagent boundary

Command: `grep -rn 'AskUserQuestion' internal/paths internal/defs | grep -v "_test.go" | grep -v "// "` → (empty — 0 matches).

### Run-phase design decisions (within SPEC latitude, recorded for audit)

1. **Segment constants as defs-owned local aliases in internal/paths** (not an import): AC-007 requires `go list -deps ./internal/paths` to list stdlib only; importing internal/defs (module-path-prefixed) would add a non-stdlib entry. The glmcred `EnvTestGLMKey` alias precedent (glmcred.go:22-27) is applied to the segments; internal/defs/dirs.go gained the previously missing documented sub-segments so it remains the complete canonical owner (REQ-MHP-007).
2. **`paths.EnvHome` exported**: AC-006 forbids the `"MOAI_HOME"` literal outside envkeys.go/internal/paths, so seam-carrying sibling packages (glmcred, migrate_agency) reference the name through the exported alias instead of re-declaring it. A constant, not an extra path accessor — the sketched 10-function accessor set is unchanged (pinned by compile-time signature vars in paths_test.go).
3. **MOAI_HOME override branch at seam-carrying sites** (glmcred.Path, migrateAgencyRunner.checkpointPath): REQ-MHP-010 preserves the HomeDirFn / injected-homeDir test seams (fixtures build `<tmp>/.moai/...` and redirect the seam, not the env), so those sites honour an absolute MOAI_HOME explicitly and fall back to the seam — the same result `Join(MoaiHome(), segment)` yields, without breaking parallel-safe seams. The 1-line override predicate mirrors MoaiHome's internal check and is pinned by tests at all three sites.
4. **Seam-site joins consume defs constants** (model_cache ×2, usage, migrate_profiles, glmcred fallback, checkpointPath fallback): kills the home-anchored `".moai"` literal (P1) while the seam parameter keeps raw-home semantics. Seam CALLERS resolve via paths (builder.go:110 → paths.Home(), glmcred default HomeDirFn → paths.Home) per spec.md §3.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-17T00:00:00+09:00
run_commit_sha: d50725c85
run_status: audit-ready
ac_pass_count: 14
ac_fail_count: 0
ac_na_run_count: 2
preserve_list_post_run_count: 0
preserve_list_note: "no PRESERVE target modified (update.go touched only at :827; migrate_profiles.go join-only; todo.go/memory.go/linkage.go/defaults.go/cache.yaml/.mcp.json/migrate_profiles_test.go absent from d6b80a01c..HEAD diff)"
l44_pre_commit_fetch: false
l44_post_push_fetch: false
l44_note: "no push performed — repo-local PR policy (branch protection enforce_admins); push + PR belong to the orchestrator's manager-git delegation, so the post-push fetch check is not yet applicable"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: pass
  windows_amd64: pass
total_run_phase_files: 31
m1_to_mN_commit_strategy: "per-milestone Conventional Commits on worktree branch worktree-moai-home-paths (M1 6600bfd8e incl. spec.md draft->in-progress flip + plan artifacts, M2 da85d179c, M3 51a587bb0, M4 2a598a877, M5 d50a54ba9, M6 = evidence commit backfilling run_commit_sha)"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-17T01:20:00+09:00
sync_commit_sha: 796394d7d
sync_status: audit-ready
b12_self_test_a: "pre-emission grep -c 'SPEC-V3R6-MOAI-HOME-PATHS-001' CHANGELOG.md → 0 (duplicate guard PASS; [Unreleased] section created fresh — release-prep renames it per house lifecycle)"
b12_self_test_b: "AC count match: acceptance.md SSOT = 16 (AC-MHP-001..016, at the Tier M ceiling); canonical grep returns 19 raw tokens of which AC-001/AC-002/AC-013 are in-file prose shorthand aliases of the AC-MHP family — CHANGELOG entry cites 16"
b12_self_test_c: "all 7 implementation paths named in the entry verified via ls (paths.go, envkeys.go, homedir.go, defs/dirs.go, glmcred.go, pre_tool.go, model_cache.go)"
changelog_entry_position: "CHANGELOG.md [Unreleased] > Added — first entry (section created by this sync; top of file above [3.1.0])"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (this sync commit; updated: already 2026-08-17)"
  acceptance_plan: "markdown-header convention — no frontmatter status to transition"
mx_debt_check: "moai mx query --kind DEBT --json → empty output, exit 0 (no dangling DEBT markers; paths.go carries @MX:ANCHOR ×2 + @MX:NOTE ×1 — fine)"
lane_routing: "release/v3.1.1 integration lane per user directive — run merged at e518fc897; sync commits on this branch, no per-phase PR, no push (lane push owned by another session)"
close_summary: "3-phase close complete — CHANGELOG [Unreleased] entry emitted (incl. AC-MHP-014 §4 .env.glm shell/Go split hazard note + AC-MHP-013 supersession linkage), spec.md frontmatter in-progress → completed, §E.4 recorded; AC 16/16 discharged (14 run-phase PASS + 2 doc-check PASS in this commit)"
```

