---
id: SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001
title: "Hook cwd leak audit + resolveProjectRoot consistency — Progress Tracker"
version: "0.2.0"
status: implemented
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
tier: S
---

# SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001 — Progress Tracker

## Status Overview

| Phase | Status | Started | Completed | Commit |
|-------|--------|---------|-----------|--------|
| Plan | COMPLETE | 2026-05-23 | 2026-05-23 | `d553d585b` |
| Run M1 | COMPLETE | 2026-05-23 | 2026-05-23 | `e629afc9d` |
| Run M2 | COMPLETE | 2026-05-23 | 2026-05-23 | `8a8ad3051` |
| Run M3 | COMPLETE | 2026-05-23 | 2026-05-23 | `0f4860e17` |
| Sync | NOT STARTED | — | — | — |

---

## Plan-Phase Deliverables (completed 2026-05-23)

- [x] `spec.md` — 4 EARS requirement categories, §A Pre-existing State Survey (5 facts), Exclusions, Risks, Cross-references
- [x] `plan.md` — Section A (Pre-flight) + Section B (Implementation Approach, 9 sub-sections) + Section C (M1/M2/M3) + Section D (Risk Mitigation) + Section E (Implementation Decisions placeholder) + Section F (Cross-references)
- [x] `acceptance.md` — 7 binary ACs with single verification commands each, Summary Matrix, Definition of Done
- [x] `progress.md` — this file

**Plan-phase scope verification**:
- 9 cwd leak patterns confirmed via `grep -rn "os\.Getwd" internal/hook/ | grep -v "_test.go"` (count = 9, matches §A.1 enumeration)
- Reference pattern `resolveProjectRoot(input *HookInput)` verified in `post_tool_metrics.go:98-113`
- Prior fix commit `a9b3e8cd8` verified in git log
- PRESERVE list (3 files) identified and listed in Exclusions

---

## Run-Phase Milestones

### M1 — `subagent_start.go` refactor (2 sites)

**Status**: NOT STARTED

**Entry criteria** (verify before starting):
- [ ] `git branch --show-current` returns the working branch (main or feat/SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001)
- [ ] `grep -rn "os\.Getwd" internal/hook/ | grep -v "_test.go" | wc -l` returns 9
- [ ] M0 lint baseline captured to `/tmp/m0-lint-baseline.txt`
- [ ] M0 race baseline captured (test suite passes with `-race`)
- [ ] PRESERVE file hashes captured to `/tmp/preserve-baseline.sha256`

**Deliverables checklist**:
- [ ] Add `resolveProjectRootFromEnv(caller string) string` helper to `internal/hook/post_tool_metrics.go`
- [ ] Add `resolveProjectRootFromInputOrEnv(input *HookInput, caller string) string` helper to same file
- [ ] Refactor `subagent_start.go:37` → `resolveProjectRootFromEnv("NewSubagentStartHandlerWithConfig")`
- [ ] Refactor `subagent_start.go:211` → `resolveProjectRootFromInputOrEnv(input, "subagentStartHandler.Handle")`
- [ ] Run `go test -race -count=1 ./internal/hook/...` → PASS
- [ ] Verify `grep -n "os\.Getwd" internal/hook/subagent_start.go` → 0 matches
- [ ] Verify PRESERVE hashes unchanged via `sha256sum -c /tmp/preserve-baseline.sha256`
- [ ] Commit M1 with the structured message in plan.md § C.M1

**Exit criteria**:
- [ ] M1 commit pushed (orchestrator action; manager-develop does NOT push per B9)
- [ ] Local test suite green
- [ ] 7 cwd leaks remaining (M2/M3 scope)

### M2 — `pre_tool.go` + `observability_master.go` refactor (3 sites)

**Status**: NOT STARTED

**Entry criteria**:
- [ ] M1 commit landed (or local M1 work green)
- [ ] M1 helpers (`resolveProjectRootFromEnv`) reusable from M2

**Deliverables checklist**:
- [ ] Refactor `pre_tool.go:326` (NewPreToolHandler) → `resolveProjectRootFromEnv("NewPreToolHandler")`
- [ ] Refactor `pre_tool.go:336` (NewPreToolHandlerWithScanner) → `resolveProjectRootFromEnv("NewPreToolHandlerWithScanner")`
- [ ] Refactor `observability_master.go:82` (loadObservabilityMaster) → `resolveProjectRootFromEnv("loadObservabilityMaster")`
- [ ] Run `go test -race -count=1 ./internal/hook/...` → PASS
- [ ] Verify `grep -n "os\.Getwd" internal/hook/pre_tool.go internal/hook/observability_master.go` → 0 matches
- [ ] Verify PRESERVE hashes unchanged
- [ ] Commit M2 with the structured message in plan.md § C.M2

**Exit criteria**:
- [ ] M2 commit pushed (orchestrator action)
- [ ] Local test suite green
- [ ] 4 cwd leaks remaining (M3 scope — quality/gate.go)

### M3 — `quality/gate.go` refactor (4 sites) + status:implemented

**Status**: NOT STARTED

**Entry criteria**:
- [ ] M2 commit landed (or local M2 work green)
- [ ] Decision logged: use literal `"CLAUDE_PROJECT_DIR"` string in `package quality` (avoids `internal/config` import dependency)

**Deliverables checklist**:
- [ ] Add `resolveQualityProjectDir(cfg GateConfig, caller string) string` helper to `internal/hook/quality/gate.go`
- [ ] Refactor `gate.go:276` → `resolveQualityProjectDir(g.config, "QualityGate.executeStep.astgrep")`
- [ ] Refactor `gate.go:297` → `resolveQualityProjectDir(g.config, "QualityGate.detectToolchain")`
- [ ] Refactor `gate.go:399` → `resolveQualityProjectDir(g.config, "QualityGate.executeStep.extfilter")`
- [ ] Refactor `gate.go:480` → `resolveQualityProjectDir(g.config, "QualityGate.anyConfigFileExists")`
- [ ] Run REQ-HCWA-011 inventory: `grep -rn "os\.Getwd" internal/hook/ | grep "_test.go" > .moai/specs/SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001/test-getwd-inventory.txt`
- [ ] Update `spec.md` frontmatter: `status: draft → implemented`, `version: "0.1.0" → "0.2.0"`, `updated: <today>`
- [ ] Fill `plan.md` Section E "Notable Implementation Decisions"
- [ ] Update this `progress.md` to mark M1/M2/M3 as COMPLETE with commit hashes

**Exit criteria — all 7 ACs verified**:
- [ ] AC-HCWA-001 — `grep -rn "os\.Getwd" internal/hook/ | grep -v "_test.go" | wc -l` returns 0
- [ ] AC-HCWA-002 — Helper call sites: ≥5 env helpers in `package hook`, ≥4 in `package quality`
- [ ] AC-HCWA-003 — `cohabitation_guard_test.go` byte-identical AND test PASS
- [ ] AC-HCWA-004 — `go test -race -count=1 ./internal/hook/...` exit 0, 0 race warnings
- [ ] AC-HCWA-005 — golangci-lint count <= M0 baseline (`/tmp/m0-lint-baseline.txt`)
- [ ] AC-HCWA-006 — `subagent_stop.go` SHA-256 unchanged vs baseline
- [ ] AC-HCWA-007 — `resolveProjectRoot` function body 15-17 lines, env-var lookup + .moai/ Stat guard intact

**Final commit**:
- [ ] M3 final commit per plan.md § C.M3 commit message (includes SPEC ID, all 3 milestones, AC summary)

---

## Sync Phase (post-M3)

**Status**: NOT STARTED

**Triggers**: All 7 ACs pass, M3 commit landed on main (or PR merged).

**Sync deliverables** (manager-docs scope, NOT this SPEC):
- [ ] `CHANGELOG.md` `[Unreleased]` section append: `feat(SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001): close 9 cwd leak sites in internal/hook/`
- [ ] Optional: codemaps regen for `internal/hook/` if helper-extraction shifts function locations
- [ ] No README/docs-site update needed (internal refactor, no user-facing surface change)

---

## Verification Snapshot (updated at M3)

```
M0 baseline (plan-phase, HEAD fc9ce6822):
- cwd leak count: 9
- lint findings: 0 issues (golangci-lint ./internal/hook/...)
- race detector: PASS (all 13 packages)
- preserve hashes:
  - subagent_stop.go            = 69cbeaa66d9026392e96dafa31d53a340bcc99fd2e3e59fb6e1de09cf3122b7e
  - post_tool_metrics.go        = 17e46ee9fc717215d61175cff669031a61fcaf7d9f4e7b9604d76c91874c2790
  - cohabitation_guard_test.go  = e1afdee26f7693419f16e401dfc9a7b161128c1367f88f880b6a0989346487f2

M3 final (post-implementation):
- cwd leak count: 0 (helper definitions + comments excluded per AC-HCWA-001 refinement)
- lint findings: 0 issues (same as M0)
- race detector: PASS (all 13 packages)
- preserve hashes: ALL UNCHANGED (verified by `shasum -a 256 -c /tmp/preserve-baseline.sha256`)

AC Matrix (all 7 PASS):
- AC-HCWA-001 PASS — 0 leak sites in production hook code (helpers + comments excluded)
- AC-HCWA-002 PASS — 6 env-helper call sites (≥5 required) + 4 quality-helper call sites (≥4 required)
- AC-HCWA-003 PASS — cohabitation_guard_test.go byte-identical AND TestCohabitationGuard test PASS
- AC-HCWA-004 PASS — go test -race exit 0, 0 race warnings
- AC-HCWA-005 PASS — golangci-lint 0 NEW issues vs M0 baseline
- AC-HCWA-006 PASS — subagent_stop.go SHA-256 unchanged
- AC-HCWA-007 PASS — resolveProjectRoot body 16 lines, env-var + .moai/ Stat both present
```

---

## Notes & Blockers

_None at plan-phase. To be appended during run-phase by manager-develop or orchestrator._

---

Version: 0.2.0
Status: implemented
Last Updated: 2026-05-23
