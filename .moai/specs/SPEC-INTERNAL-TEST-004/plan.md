---
id: SPEC-INTERNAL-TEST-004
title: "Regenerate stale doctor/status golden testdata for version bump rc7→rc10 (whole-repo green)"
version: "0.1.0"
status: completed
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0-rc10"
module: "internal/cli/testdata"
lifecycle: spec-anchored
tags: "golden-test, test-fix, debt-cleanup, version-bump"
tier: S
depends_on: []
related_specs: [SPEC-INTERNAL-TEST-003, SPEC-INTERNAL-ARCH-001, SPEC-WEB-CONSOLE-011]
---

# plan.md — SPEC-INTERNAL-TEST-004

## §A. Context

| Field | Value |
|-------|-------|
| Working directory | `/Users/goos/MoAI/moai-adk-go` |
| Branch | `main` |
| HEAD | `e35c85662 feat(SPEC-AGENT-ARCH-V2-001): plan-audit iter-2 revision (D1-D4 fix)` (see `git log` for current HEAD at run-time) |
| `origin/main...HEAD` | local-ahead state (see `git rev-list --count --left-right origin/main...HEAD` at run-time) |
| Route | Route A — Hybrid Trunk main-direct (Tier S default) |
| SPEC artifacts | `.moai/specs/SPEC-INTERNAL-TEST-004/{spec,plan,acceptance,progress,research}.md` |
| Debt provenance | SPEC-INTERNAL-TEST-003 AC-006 (external debt transfer, TEST-003 completed) |
| Unblock target | SPEC-INTERNAL-ARCH-001 plan-audit M0 (whole-repo-green precondition) |
| Root cause | `pkg/version/version.go` bumped rc7→rc10 (reached rc10 via `ce2a509dc`); golden testdata was stuck at rc7 — NOW FIXED by external commit `ce2a509dc` (see spec.md §G) |
| Approach | Golden regenerate only (`UPDATE_GOLDEN=1`); no Go source changes |

### §A.1 Tier classification rationale

Tier S — single concern (stale golden testdata regeneration due to version bump), < 5 affected files in run-phase scope (6 golden testdata files, each a 1-line mechanical change), no code logic change, no API change, no architectural impact. The `UPDATE_GOLDEN=1` + `git diff` evidence (`research.md` §2.3) confirms the change is mechanical and uniform.

### §A.2 Plan-phase research verdict

The drift cause is FULLY resolved by `research.md` — no blocker report is needed. The cache-hit emoji scope boundary (research Q2/Q4) is decided: the emoji change belongs to WEB-CONSOLE-011 and is PRESERVE for TEST-004 (research §3). The approach is golden-regenerate-only (research §5).

## §B. Known Issues

| ID | Category | Applicability | Note |
|----|----------|---------------|------|
| B1 | Cross-platform build tags | N/A | No Go source changes; testdata-only |
| B2 | Cross-SPEC policy conflict | Checked | No retired/superseded SPEC conflict in `internal/cli/testdata/` |
| B3 | Subagent boundary (C-HRA-008) | N/A | No harness/hook code touched |
| B4 | Frontmatter canonical schema | Applied | 12 fields, canonical names, `created`/`updated` (not snake_case) |
| B5 | CI 3-tier awareness | Applies | spec-lint + golangci-lint + Test; testdata change is Test-tier only |
| B6 | spec-lint heading rules | Applied | `### Out of Scope — <topic>` H3 sub-headings present in spec.md §E |
| B7 | observer.go capture path | N/A | No hook code |
| B8 | Working tree hygiene | **CRITICAL** | PRESERVE ~15 unrelated paths; `git add` specific path only |
| B9 | Git commit + push (Hybrid Trunk) | Applies | manager-develop commits + pushes to main directly (Tier S Route A) |
| B10 | Untouched paths PRESERVE | **CRITICAL** | Only `internal/cli/testdata/*.golden` (6 files) in run-phase |
| B11 | AskUserQuestion prohibition | Applies | Subagent returns blocker report, never prompts |
| B12 | CHANGELOG emission | N/A | manager-docs owns sync-phase; not a plan-phase concern |

## §C. Pre-flight Check List (manager-develop executes before code change)

```bash
# 1. Branch + baseline
git branch --show-current                          # expect: main
git rev-parse HEAD                                 # expect: bd97306af (or descendant)
git rev-list --count --left-right origin/main...HEAD  # expect: 0 0 (or 0 N if local ahead)

# 2. Build sanity (no code change expected, but verify baseline builds)
go build ./...                                     # expect: exit 0

# 3. Current golden FAIL baseline (confirm the 6 FAIL before fixing)
go test ./internal/cli/ -run 'TestDoctor_Current_Light|TestStatus_Current_Light' -count=1 2>&1 | tail -5
# expect: 2+ FAIL (confirming the drift is present)

# 4. version.go value (confirm rc9 — the target golden value)
grep 'Version = ' pkg/version/version.go           # expect: Version = "v3.0.0-rc9"

# 5. PRESERVE list snapshot (confirm working-tree state before touching testdata)
git status --short internal/statusline/            # expect: 2 modified (renderer.go + cache_hit_test.go) — PRESERVE
```

## §D. Constraints / PRESERVE List

### §D.1 Run-phase MAY modify (exhaustive)

| Path | Change |
|------|--------|
| `internal/cli/testdata/doctor-light.golden` | version string rc7→rc9 (1 line) |
| `internal/cli/testdata/doctor-dark.golden` | version string rc7→rc9 (1 line) |
| `internal/cli/testdata/doctor-nocolor.golden` | version string rc7→rc9 (1 line) |
| `internal/cli/testdata/status-light.golden` | version string rc7→rc9 (1 line) |
| `internal/cli/testdata/status-dark.golden` | version string rc7→rc9 (1 line) |
| `internal/cli/testdata/status-nocolor.golden` | version string rc7→rc9 (1 line) |

### §D.2 PRESERVE (MUST NOT modify — exhaustive working-tree snapshot)

| Path | Owner / reason |
|------|----------------|
| `internal/statusline/renderer.go` | WEB-CONSOLE-011 (💾→♻️, ancestor `22220186c`) |
| `internal/statusline/cache_hit_test.go` | WEB-CONSOLE-011 (test sync) |
| `pkg/version/version.go` | Release process (rc9 is frozen) |
| `.claude/agents/moai/manager-docs.md` | Unrelated work |
| `.claude/agents/moai/manager-git.md` | Unrelated work |
| `.moai/config/sections/llm.yaml` | Unrelated work |
| `.moai/config/sections/workflow.yaml` | Unrelated work |
| `.moai/specs/SPEC-OBSERVE-HYGIENE-001/progress.md` | Unrelated work |
| `internal/template/catalog.yaml` | Unrelated work |
| `internal/template/templates/.claude/agents/moai/manager-docs.md` | Unrelated work |
| `internal/template/templates/.claude/agents/moai/manager-git.md` | Unrelated work |
| `internal/template/templates/.moai/config/sections/llm.yaml` | Unrelated work |
| All `??` untracked paths | Various owners — do NOT `git add -A` |

### §D.3 Forbidden commands

- `git add -A` / `git add .` (use `git add internal/cli/testdata/` specific path only)
- `--no-verify` (pre-commit hook warn-only is normal)
- `--amend` / force-push to main (per CONST-V3R5-007)
- Modifying `pkg/version/version.go` (frozen at rc9)
- Modifying any `internal/statusline/` file

### §D.4 Required conventions

- Conventional Commits: `fix(SPEC-INTERNAL-TEST-004): M1 regenerate doctor/status goldens rc7→rc9`
- `🗿 MoAI` commit trailer
- Tier S Route A: commit + push directly to main (no PR)

## §E. Self-Verification (manager-develop reports at run-phase completion)

Per `verification-claim-integrity.md` 5-section format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk).

| Item | Verification |
|------|--------------|
| E1 | AC binary PASS/FAIL matrix (acceptance.md §D) — 6 golden + whole-repo + git-diff + PRESERVE |
| E2 | `go build ./...` exit 0 (baseline sanity, no code change) |
| E3 | Coverage N/A (testdata-only change, no source coverage delta) |
| E4 | Subagent boundary N/A (no harness/hook code) |
| E5 | `golangci-lint run` baseline unchanged (no Go source modified) |
| E6 | Commit SHA + `git push origin main` result |
| E7 | Blocker report if any (none expected — scope fully resolved) |

## §F. Milestones

> **External-resolution note (2026-07-09).** M1/M2 golden regeneration was performed by external commit `ce2a509dc chore(version): bump to v3.0.0-rc10 and realign tag with source` (the rc10 version-bump commit), NOT by a TEST-004 run-phase commit. TEST-004 therefore has **NO own run-phase commits** — the SPEC proceeds directly from plan to a consolidated sync close. AC-007 (whole-repo green) is debt-transferred to SPEC-AGENT-ARCH-V2-001 (the 3 remaining `internal/template` FAILs belong to the super-advisor integration). See `spec.md` §G for the full resolution record. The M1/M2 milestones below are retained as the plan-phase design record of what the golden regeneration would have looked like if TEST-004 had executed its own run-phase.

### M1 — Regenerate 6 goldens + verify golden diff

**Action**: `UPDATE_GOLDEN=1 go test ./internal/cli/ -run 'TestDoctor_|TestStatus_' -count=1`
**Verify**:
- `git diff --stat internal/cli/testdata/` → 6 files, 6 insertions, 6 deletions
- `git diff internal/cli/testdata/ | grep -E '^[-+].*rc[0-9]' | sort -u` → exactly 4 lines (2 `-` rc7 + 2 `+` rc9, the 2 unique line patterns)
- `go test ./internal/cli/ -run 'TestDoctor_|TestStatus_' -count=1` → exit 0 (all 6 PASS without UPDATE_GOLDEN)
**Commit**: `fix(SPEC-INTERNAL-TEST-004): M1 regenerate doctor/status goldens rc7→rc9`

### M2 — Whole-repo exit 0 verification + push

**Action**: `go test ./...`
**Verify**:
- exit 0
- 0 FAIL packages (93+ ok packages expected)
- PRESERVE check: `git status --short` shows only the 6 golden files as modified (plus pre-existing PRESERVE paths unchanged)
**Push**: `git push origin main`
**Evidence**: persist verbatim output to `.moai/state/verify/<session>/test-004-m2-wholerepo.log`

**No M3** — no code fix needed (golden regenerate is necessary and sufficient per research §5).

## §G. Anti-Patterns

- `git add -A` instead of specific-path `git add internal/cli/testdata/` — would sweep PRESERVE paths into the commit
- Modifying `pkg/version/version.go` to match goldens (reversing the dependency) — version.go is the source of truth; goldens follow
- Touching `internal/statusline/renderer.go` / `cache_hit_test.go` — WEB-CONSOLE-011 scope
- Characterizing this as a "statusline drift" (TEST-003's imprecise label) — it is a version-bump golden drift
- Running UPDATE_GOLDEN=1 outside the targeted `-run` filter — would regenerate unrelated goldens if any exist

## §H. Cross-References

- `spec.md` — requirements (GEARS), Out of Scope
- `acceptance.md` — AC matrix, Given-When-Then
- `research.md` — drift-cause evidence, scope decision, approach rationale
- `progress.md` — §E.1 plan-phase audit-ready signal
- SPEC-INTERNAL-TEST-003 (debt transfer source, completed)
- SPEC-INTERNAL-ARCH-001 (whole-repo-green M0 consumer)
- SPEC-WEB-CONSOLE-011 (cache-hit emoji scope owner)
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability — Tier S minimal delegation form permitted
