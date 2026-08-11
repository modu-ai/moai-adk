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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

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
