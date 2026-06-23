---
id: SPEC-SKILL-BODY-NEUTRALITY-001
title: "Skill-Body Neutrality — run-phase progress"
version: "0.1.1"
status: in-progress
created: 2026-06-04
updated: 2026-06-23
author: manager-develop
priority: P1
phase: "v3.0.0"
module: "internal/template/templates/.claude/skills"
lifecycle: spec-anchored
tags: "template-system, neutrality, skills, ci-guard, distribution"
tier: M
---

# Run-phase Progress — SPEC-SKILL-BODY-NEUTRALITY-001

## §E.1 Run-phase context (re-run on current main base)

- Worktree base: this run-phase is a RE-RUN on the current local main HEAD `9c5f0e2b1` (the prior run-phase, ~640 commits ago, is STALE and discarded).
- L1 worktree materialized by runtime at `.claude/worktrees/agent-a6a80c9a8d9750c04`, initially based on `origin/main` (`3629ed232`) which LACKED the SPEC files. The 4 intervening commits between `3629ed232` and local main `9c5f0e2b1` touched ONLY SPEC artifacts + CLAUDE.local.md — NOT skill bodies or the leak test — so the edited files are byte-identical at both bases. The worktree-base divergence is therefore inert for this SPEC's scope; final commits FF/cherry-pick cleanly onto local main.
- cycle_type: tdd (Part B RED guard first → Part A purge → Part B GREEN).
- Plan inventory was authored ~640 commits ago and DRIFTED. Ground-truth re-grep was performed fresh on the current tree (see §E.2 RED evidence). Key divergences from stale plan:
  - CLASS4 `99-release`: plan claimed 6 hits; current tree has **2** (`INDEX.md:143`, `moai/references/reference.md:241`). `commands-reference.md` no longer carries `99-release`.
  - CLASS4 `97-release-update` / `release-update` / "NOT distributed": **0** hits (already cleaned in the 640-commit interval; `moai/SKILL.md` carries none).
  - CLASS4 maintainer-doctrine date `2026-05-26`: **0** hits. Residue is `catch-up SPEC` + `maintainer doctrine` in `moai-meta-harness/SKILL.md` (lines 232, 248).
  - CLASS3 SPEC-V3R: 37 files (matches plan). CLASS3 REQ-token: 27 files (broader than plan's stated 13 — default-tier regex). CLASS2 Go-path: 9 files (incl. `internal/design/dtcg/frozen_guard_test.go`).

## §E.2 Run-phase Evidence

### M1 — Part B RED (extend the neutrality guard) — COMPLETE

Guard file edited: `internal/template/internal_content_leak_test.go`. Additions (all skill-body-scoped via new `skillBodyScoped` field + `skillBodyPrefix = ".claude/skills/"`, so they do NOT fire on agents/rules/hooks/config per EXCL-SBN-002):

- `C1b-spec-id-skill-v3r` — broadens SPEC-ID detection to `SPEC-V3R[0-9]-*` / `CONST-V3R[0-9]-*` + named `SPEC-WF-AUDIT-GATE-001` / `SPEC-MX-001` (REQ-SBN-014 / REQ-SBN-006).
- `C6-agentless-test-ref` — matches the literal `agentless_audit_test.go` (REQ-SBN-012).
- `C7-internal-go-path` — package-restricted `internal/(spec|cli|hook|ciwatch|design)/[a-z0-9_/]*\.go` (REQ-SBN-013 HARD; does NOT match `internal/auth|api|core` illustrative paths).
- `S3-req-ac-token-any-prefix` — PROMOTED from former strict tier into the default tier (skill-body-scoped), as a strict-superset regex `(REQ|AC)-[A-Z][A-Z0-9]*-[0-9]+` (catches both `REQ-BRAIN-001` and the `REQ-WF003-010` form; the original narrow `[A-Z]{2,}` form missed the WF-NNN form). The former `strictLeakClasses` S3 sibling was REMOVED so there is exactly ONE REQ-token regex entry (AC-SBN-018(b) partition guard). The superset still matches the original S3 probe → remains a strict superset.

Belt-and-suspenders allowlist (REQ-SBN-013 / AC-SBN-020(c)): the 3 illustrative Go paths (`internal/auth/login.go`, `internal/api/handler.go` in `pr-review-multi-agent.md`; `internal/core/handler.go` in `mx.md`) added to `pedagogicalAllowlist`.

New structural unit tests added: `TestLeakClassReqTokenPartition` (AC-SBN-018(b)), `TestLeakClassNoDateShaInDefaultTier` (AC-SBN-018(a)), `TestSkillBodyLeakClassRecurrenceBackstop` (AC-SBN-017), `TestC7PackageRestriction` (AC-SBN-020(a)+(b)).

#### M1 RED checkpoint — AC-SBN-012 evidence (test FAILS on current leaks)

Command: `go test ./internal/template/ -run '^TestTemplateNoInternalContentLeak$'`
Result: **FAIL** — `template internal-content leak detected (185 occurrences, mode=narrow)`.

Per-class uncapped counts (grep ground-truth against `internal/template/templates/.claude/skills/`):

| Class | Occurrences | Files | Requirement |
|-------|-------------|-------|-------------|
| C6-agentless-test-ref | 6 | 6 | AC-SBN-012 (≥6 file findings) — PASS |
| C1b-spec-id-skill-v3r | 101 | — | REQ-SBN-006/014 |
| C7-internal-go-path | 12 | 9 | REQ-SBN-005/013 |
| S3-req-ac-token-any-prefix | 131 | — | REQ-SBN-007 |

The scan correctly does NOT flag the 3 illustrative paths (allowlisted) nor `SPEC-AUTH-001`/`SPEC-001` placeholders.

Structural unit tests at M1 (all PASS — they validate the guard structure, not the leaks):
- `TestLeakClassReqTokenPartition` PASS
- `TestLeakClassNoDateShaInDefaultTier` PASS
- `TestSkillBodyLeakClassRecurrenceBackstop` PASS
- `TestC7PackageRestriction` PASS

(M2-M6 evidence appended as milestones complete.)

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: PENDING
run_commit_sha: PENDING
run_status: in-progress
ac_pass_count: PENDING
ac_fail_count: PENDING
m1_red_evidence: captured (185 occurrences, C6=6 files, C1b=101, C7=12, S3=131)
```
