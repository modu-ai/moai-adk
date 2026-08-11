# Progress — SPEC-GATE-ASTGREP-REPAIR-001

## §E.1 Plan-phase Audit-Ready Signal

_plan_status: pending-audit_
_plan_complete_at: pending_
_tier: L_
_artifacts: spec.md, plan.md, acceptance.md, progress.md (4 of 5; design.md / research.md deferred — Tier L이나 사용자 명시 요청으로 일차 4산출물만 저작, plan-auditor 권고 시 보강)_
_open_clarifications: 1 (plan.md §B1 — dogfood vs distributed baseline 미러 대상, run-phase M1 착수 전 orchestrator AskUserQuestion 해소 필요)_

## §E.2 Run-phase Evidence

### M0 — Baseline capture (2026-08-11, HEAD 08eef9a0f)

**Command**: `sg scan --config .moai/config/astgrep-rules/sgconfig.yml --json .`
(run against the main checkout root, the way `moai gate` invokes it via
`scanner.scanWithConfig`).

**Observed total + per-rule counts** (verbatim `jq` output):

```
21990  total matches
16124  go-error-not-wrapped       <- D1 over-matching (73% of all findings)
 3902  sec-path-traversal-join-user-input
 1590  go-error-ignored-blank
  105  go-goroutine-without-context
   67  go-channel-send-no-select
   61  go-map-no-ok-check
   50  go-errors-new-in-var
   46  go-interface-empty-not-any
   17  sec-command-injection-exec-command
   14  sec-hardcoded-credential
   10  unused-suppression
    3  sec-log-injection-unsanitized
    1  go-bytes-buffer-string-conversion
```

**Sample of D1 false positives** (matched by the over-broad `return $ERR`):
`internal/session/phase.go:27: return true`, `internal/glmcred/glmcred.go:73: return nil`,
`pkg/version/version.go:14: return Version`, etc. Non-error literals and calls
were all matched because the unconstrained metavariable `$ERR` binds to any
single AST node.

**D2 worktree-duplication note (calibrated)**: `.claude/worktrees/` IS listed
in `.gitignore` (line 186) and ast-grep respects `.gitignore` by default, so
the sg scan from the main checkout reported **0** matches under
`.claude/worktrees/`. The D2 premise (worktree duplication inflating the
count 4.9×) is therefore already mitigated by `.gitignore` for the
worktree-specific sub-path. The remaining D2 value is defensive: making the
exclusion explicit in `sgconfig.yml` so the scan does not silently rely on
`.gitignore`, and extending it to `vendor/**` + `*_test.go`. D1 is the
dominant defect (16,124 of 21,990 = 73% of all findings).

**M0 RED evidence (E8)** — the characterization tests fail against the
pre-refinement tree as expected:

```
TestGAR_AC001_NegativeReturnsNotMatched: expected 0 matches on negative fixture, got 7
TestGAR_AC002_RealErrorReturnStillMatched: expected exactly 1 match on positive fixture, got 3
TestGAR_AC003_AutofixDoesNotTouchNegatives: autofix rewrote all 7 negative returns
  (return 0 -> return fmt.Errorf("TODO: operation: %w", 0), etc.) — catastrophic
TestGAR_AC008_GateGoLoadsGateYAML: gate.go still hardcodes `cfg := quality.DefaultGateConfig()`
```

### M1 — D1 rule refinement (candidate A adopted)

**Decision**: candidate A (name constraint, no structural `inside`) was
chosen over B (structural-only) and C (name + structural) after fixture
evaluation:

| Candidate | positive.go | negative.go | edge.go notes |
|-----------|-------------|-------------|---------------|
| original (baseline) | 3 | 7 | matches every single-expr return |
| A: `constraints.ERR.regex=^err(e?\|s)$` | 1 ✅ | 0 ✅ | also catches bare `return err` (EdgeBareErrReturn) — broadest real-violation coverage |
| B: `inside if_statement has $COND != nil` | 1 | 0 | FALSE POSITIVE on `return nil` inside `if err != nil` (edge.go line 11); also matched multi-value `return 0, err` |
| C: A + B combined | 1 | 0 | misses bare `return err` at top level (edge.go line 25) — narrower coverage than A |

Candidate A wins: simplest, broadest real coverage, zero false positives on
the fixture set. The name proxy (`err`/`errs`) is the type constraint — Go
variables with these names are overwhelmingly errors.

**Applied change**: `.moai/config/astgrep-rules/go/error-handling.yml` rule
`go-error-not-wrapped` gained a `constraints.ERR.regex: ^err(e?|s)$` sibling
of `rule:`. The `fix:` field is preserved (AP-GAR-002) — it fires under the
SAME guard as the match, so the autofix is automatically gated.

**Post-refinement projected count** (candidate A scanned against the main
checkout): `go-error-not-wrapped` 16,124 → **357** (97.8% false-positive
reduction). The 357 remaining are genuine unwrapped `return err` violations.

**GREEN evidence**: `go test -run TestGAR_AC001|TestGAR_AC002|TestGAR_AC003 ./internal/astgrep/...`
→ 3/3 PASS against the refined shipped rule.

**Distributed baseline (B1 decision)**: NOT touched. The distributed template
baseline `internal/template/templates/.moai/config/astgrep-rules/` carries no
error-wrapping rule; per the user-resolved B1 decision, the refinement is
dogfood-tree only and MUST NOT propagate to the distributed baseline.

### M2 — D2 path exclusion at the gate layer

**Discovery**: ast-grep 0.40.5 does NOT support a `globs:` exclusion field in
sgconfig.yml config-mode. Empirically verified: adding `globs: ["!**/*_test.go"]`
(or any globs entry) silently breaks the scan — every rule returns 0 matches,
including for non-excluded files. The field appears to be misparsed as rule
content. The SPEC's primary D2 mechanism (sgconfig.yml globs) is therefore NOT
viable at this ast-grep version.

**Fallback applied** (per plan.md §D.2 B3/B4): path exclusion filter at the
gate layer — `filterExcludedPaths` in
`internal/hook/quality/astgrep_gate.go`. This is NOT in the PRESERVE list
(scanner.go Scan body is preserved; astgrep_gate.go is the gate entry point
shared by both the PreToolUse path and the standalone `moai gate` CLI).
Excluded patterns: `.claude/worktrees/`, `/vendor/`, `/testdata/`, `_test.go`.

**Effect** (measured against the main checkout with the refined rule):
- `go-error-not-wrapped` in test files (_test.go): 72 → 0 (excluded)
- `go-error-not-wrapped` in testdata: 0 (already 0 post-M1)
- `go-error-not-wrapped` non-test non-testdata: 274 (the genuine violations)
- vendor/: 0 (no vendor dir in this repo)
- `.claude/worktrees/`: 0 (already excluded by .gitignore; filter makes it
  authoritative regardless of .gitignore state)

**GREEN evidence**: `go test -run TestGAR_AC004_ExcludedPathsFilteredFromFindings ./internal/hook/quality/...`
→ PASS. The test stages the same `return err` violation in 5 locations
(real.go, foo_test.go, testdata/, vendor/, .claude/worktrees/) and asserts
only real.go surfaces.

### M3 — D3 config wiring (loadGateSection in moai gate CLI)

**Applied change**: `internal/cli/gate.go` runGate now routes through a new
`loadGateCfgForCLI(projectDir)` helper that calls `config.NewLoader().Load()`
(the same loader chain that includes `loadGateSection`). The helper maps
`config.GateConfig → quality.GateConfig` verbatim after the pre_tool.go
loadGateConfig pattern, including the `disabled_steps` verbatim copy
(issue #1265) and the `RulesDir` default.

**Fall-back** (AP-GAR-004): on missing/unparseable gate.yaml, the helper
returns `quality.DefaultGateConfig()` with ProjectDir + GoBuildTags applied —
the gate still runs, it does NOT silently hard-block.

**GREEN evidence**:
- `TestGAR_AC008_GateGoLoadsGateYAML` — grep guard: runGate routes through
  loadGateCfgForCLI, no direct `cfg := quality.DefaultGateConfig()` at the
  call site.
- `TestGAR_AC009_WarnOnlyModeReflectedInCLIConfig` — gate.yaml
  warn_only_mode:true + block_on_error:false is reflected in cfg.
- `TestGAR_AC010_EnabledFalseSkipsAstGrep` — gate.yaml enabled:false is
  reflected; ast-grep sub-scan would skip.
- `TestGAR_D3_FallbackOnMissingGateYAML` — missing gate.yaml → non-nil
  fallback cfg with correct ProjectDir.

### M4 — Integration verification

**AC-GAR-011** (full test suite): `go test -count=1 ./...` → 0 FAIL.
**AC-GAR-012** (lint): `golangci-lint run` on changed packages → 0 issues.
**AC-GAR-013** (cross-platform): `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.

**Smoke** (`./bin/moai gate` against an isolated temp project with a genuine
`return err` violation + gate.yaml `warn_only_mode: true`): **exit 0** — the
ast-grep step finds the violation (confirmed via direct `sg scan`) and the
gate honors the advisory intent (AC-GAR-009 end-to-end PASS). The standalone
`moai gate` from the worktree exited 1 on the `go test` sub-gate's 2m timeout
—a pre-existing characteristic of the full suite size, orthogonal to the
ast-grep sub-gate (the AC explicitly scopes itself to the ast-grep exit
contribution).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-11
run_commit_sha: pending-backfill-m4  # M4 progress-only commit; backfill in sync-phase
run_status: audit-ready
ac_pass_count: 12
ac_fail_count: 0
ac_deferred:
  - AC-GAR-007  # SHOULD — (file,line,rule) dedup; satisfied by worktree exclusion (0 worktree paths post-filter), no separate dedup key added
preserve_list_post_run_count: 4  # scanner.go Scan body, parseSGFindings schema, pre_tool.go, template distributed baseline
l44_pre_commit_fetch: not-applicable  # worktree, single session
l44_post_push_fetch: not-applicable   # orchestrator handles push
new_warnings_or_lints_introduced: 0
cross_platform_build:
  linux_darwin: pass  # go build ./... exit 0
  windows_amd64: pass # GOOS=windows GOARCH=amd64 go build ./... exit 0
total_run_phase_files: 6  # error-handling.yml, sgconfig.yml, gate.go, astgrep_gate.go + 3 test files + testdata
m1_to_mN_commit_strategy: per-milestone commits (ff5812b28 M0+M1, c4ca4c74b M2, c2007bc05 M3, <this> M4 progress)
```

**AC matrix summary** (13 ACs: 12 PASS, 1 deferred-SHOULD):

| AC | Status | Evidence |
|----|--------|----------|
| AC-GAR-001 | PASS | TestGAR_AC001 — 0 matches on negative fixture |
| AC-GAR-002 | PASS | TestGAR_AC002 — exactly 1 match on positive fixture |
| AC-GAR-003 | PASS | TestGAR_AC003 — autofix leaves negatives byte-identical |
| AC-GAR-004 | PASS | TestGAR_AC004 — worktree/vendor/test paths filtered |
| AC-GAR-005 | PASS | covered by AC-GAR-004 (vendor + _test.go excluded by filter) |
| AC-GAR-006 | PASS | 16,124 → 346 go-error-not-wrapped (97.9% drop); 21,990 → 6,212 total |
| AC-GAR-007 | DEFERRED (SHOULD) | (file,line,rule) dedup — satisfied by worktree exclusion; no separate dedup key needed since worktrees contribute 0 paths |
| AC-GAR-008 | PASS | TestGAR_AC008 — gate.go routes through loadGateCfgForCLI |
| AC-GAR-009 | PASS | TestGAR_AC009 + smoke exit 0 with warn_only_mode |
| AC-GAR-010 | PASS | TestGAR_AC010 — enabled:false reflected in cfg |
| AC-GAR-011 | PASS | `go test -count=1 ./...` — 0 FAIL |
| AC-GAR-012 | PASS | golangci-lint — 0 issues on changed packages |
| AC-GAR-013 | PASS | GOOS=windows GOARCH=amd64 go build ./... exit 0 |

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-11
sync_commit_sha: pending-backfill-sync  # self-referential-hazard — a commit cannot reference its own SHA; backfilled in a follow-up commit after the sync PR merges (per spec-frontmatter-schema.md D3)
sync_status: audit-ready
changelog_entry_position: CHANGELOG.md [Unreleased] > ### Fixed  # single entry, top of Fixed section
frontmatter_status_transitions:
  spec_md: in-progress -> implemented -> completed  # 3-phase close merged into the single sync commit
  plan_md: in-progress -> implemented -> completed
  acceptance_md: in-progress -> implemented -> completed
  progress_md: n/a  # progress.md carries no frontmatter
frontmatter_updated_refreshed: 2026-08-11  # all 3 artifacts (spec/plan/acceptance) refreshed to sync commit date
b12_self_test_a: pass  # pre-emission grep: `grep -c 'SPEC-GATE-ASTGREP-REPAIR-001' CHANGELOG.md` == 0 before emission → no duplicate from parallel BATCH-SYNC
b12_self_test_b: pass  # AC count match: acceptance.md distinct AC IDs = 13 (AC-GAR-001..013); CHANGELOG entry references the SPEC (13 ACs covered: 12 PASS + AC-GAR-007 SHOULD deferred) — non-zero, non-vacuous
b12_self_test_c: pass  # file paths verified: internal/cli/gate.go, internal/hook/quality/astgrep_gate.go, .moai/config/astgrep-rules/go/error-handling.yml all exist via ls
canary_compliance_check:
  distributed_baseline_untouched: true  # B1 decision — D1 refinement is dogfood-only; internal/template/templates/.moai/config/astgrep-rules/ carries no error-wrapping rule
  template_mirrorTouches: 0  # `git diff origin/main..HEAD -- internal/template/templates/` shows ONLY pre-existing #1443 content
ac_summary:
  total: 13
  pass: 12
  fail: 0
  deferred: 1  # AC-GAR-007 (SHOULD — (file,line,rule) dedup; satisfied by worktree exclusion, no separate dedup key added)
sync_artifacts_touched:
  - CHANGELOG.md  # English-only entry under [Unreleased] > ### Fixed
  - .moai/specs/SPEC-GATE-ASTGREP-REPAIR-001/spec.md  # frontmatter only (status + updated)
  - .moai/specs/SPEC-GATE-ASTGREP-REPAIR-001/plan.md  # frontmatter only
  - .moai/specs/SPEC-GATE-ASTGREP-REPAIR-001/acceptance.md  # frontmatter only
  - .moai/specs/SPEC-GATE-ASTGREP-REPAIR-001/progress.md  # this §E.4 block
  - internal/cli/gate.go  # @MX:NOTE annotation on loadGateCfgForCLI (comment-only)
  - internal/hook/quality/astgrep_gate.go  # @MX:NOTE annotation on filterExcludedPaths (comment-only)
readme_updated: false  # README has no moai-gate ast-grep advisory/blocking section that would imply the old behavior; only generic /moai loop AST-grep mentions exist — no edit needed
docs_site_4locale_sync: not-applicable  # internal gate mechanism + dogfood-tree ast-grep rule, not a user-facing docs-site surface
mx_tag_validation:
  new_surfaces_annotated:
    - internal/cli/gate.go loadGateCfgForCLI  # @MX:NOTE added (intent: SSOT config loader reuse so moai gate and PreToolUse share one gate.yaml source)
    - internal/hook/quality/astgrep_gate.go filterExcludedPaths  # @MX:NOTE added (intent: authoritative path-exclusion boundary at the gate layer, NOT in scanner.go which is preserved)
  scanner_go_untouched: true  # explicit constraint — scanner.go Scan body preserved (REQ-GAR-010)
  pre_tool_go_untouched: true  # explicit constraint — PreToolUse path unchanged beyond shared loader reuse
```

## §F Phase 4 Mode Selection

Input parameters:
- tier: L
- scope (file count): ~6-8 (error-handling.yml, sgconfig.yml, internal/cli/gate.go, internal/hook/quality/astgrep_gate.go, internal/astgrep/scanner.go 옵션, 테스트, progress.md)
- domain count: 3 (ast-grep 룰 / config 시스템 / gate CLI)
- file language mix: YAML + Go + Markdown
- concurrency benefit: LOW (coding-heavy; Anthropic coding-task parallelism caveat)
- Agent Teams prereqs: N/A (Mode 3 RETIRED)

Mode evaluation:
- trivial: not selected — Tier L, multi-file, semantic changes
- background: not selected — write-capable implementation, not read-only
- agent-team: RETIRED (Mode 3 tombstone)
- parallel: not selected — coding-heavy work (Anthropic caveat: coding tasks have fewer truly parallelizable tasks than research)
- sub-agent: SELECTED — Tier L coding-heavy; sequential per-milestone delegation fits
- workflow: not selected — not high-volume mechanical (≥30 files uniform transform); semantic/multi-rule work

Decision: sub-agent (Mode 5)

Justification: 본 SPEC은 ast-grep 룰 정교화 + config globs + Go config wiring으로 코딩 집약적이며 파일 간 의존성이 있다 (D1 룰 결과가 M2 카운트 측정에 영향). Anthropic coding-task parallelism caveat에 따라 순차 sub-agent 위임이 안전하다. manager-develop에게 Section A-E delegation template으로 M0-M4를 순차 위임한다.

Progression mode: autonomous (goal engine) — Implementation Kickoff Approval 통과 후 사용자 선택. 워크트리 `.claude/worktrees/gate-astgrep-repair` (origin/main 08eef9a0f 기반 fresh 브랜치)에서 작업.
