---
id: SPEC-MOAI-SKILL-DOCTRINE-FIX-001
title: "moai Skill Folder Doctrine Drift Remediation — Progress"
version: "0.1.0"
status: completed
created: 2026-07-08
updated: 2026-07-08
author: manager-spec
priority: P1
phase: "v3.0.0 target"
module: ".claude/skills/moai"
lifecycle: spec-anchored
tags: "skill-doctrine, drift-remediation, gears, template-neutrality, harness, tier-l, agent-catalog"
---

# Progress: SPEC-MOAI-SKILL-DOCTRINE-FIX-001

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts (spec.md, plan.md, acceptance.md, progress.md) authored 2026-07-08 from a 9-way parallel read-only audit of `.claude/skills/moai/` (42 files, ~10,043 lines; 70 verified findings — CRITICAL 7 / MAJOR 32 / MINOR 31). 53 REQs consolidated across 6 cross-cutting themes and 9 file-disjoint write-groups (WG-A..WG-I). Ground-truth cross-checks performed during authoring (see plan.md §J Cross-References) against `spec-workflow.md`, `harness.yaml`, `frozen_guard.go`, `team-protocol.md`, `workflow.yaml`, `spec-frontmatter-schema.md`, and `internal_content_leak_test.go`.

`plan_status: audit-ready` — pending plan-auditor independent review (Phase 0.5).

## §E.2 Run-phase Evidence

Mode 4 file-disjoint parallel fanout, 2 waves (WG-A/B/C/D/E, then WG-F/G/H) + WG-I sequential last. Every write-group edited the template source (`internal/template/templates/.claude/skills/moai/...`) first then synced to the local mirror (Template-First per CLAUDE.local.md §2); 3 DIFFER files (review.md, plan/clarity-interview.md, run/task-decomposition.md) had SPEC fixes applied to both sides while preserving unrelated pre-existing local drift. All 53 REQs applied. WG-H additionally corrected the stale comment at `internal/cli/root.go:157-165` (REQ-SKF-007 sub-clause b, scoped Go-comment exception). `make build` regenerated the embedded catalog (SKILL.md hash refresh). Orchestrator removed a §25 template-neutrality leak WG-H had introduced into `harness.md` (SPEC-V3R5 SPEC-ID ×2 + internal/cli/root.go path) before committing — accurate Go-binary framing preserved, internal citations genericized.

## §E.3 Run-phase Audit-Ready Signal

In-scope verification (orchestrator-observed, file-redirect contract `/tmp/moai-verify/skf/`):
- `go test ./internal/template/...` → exit 0 (`TestTemplateNoInternalContentLeak` PASS, `TestManifestHashFormat` PASS, `TestReqSkf053NewLeakClassesDetectShapes` PASS — 4 new leak shapes caught + scope-gate negative assertions).
- `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0; `go vet ./...` → exit 0.
- Cross-WG internal-token leak scan across all edited skill files (both trees) → 0 matches.
- Template↔local byte-identity re-established for all edited files (3 DIFFER files carry only their pre-existing, non-SPEC residual drift).

Out-of-scope pre-existing trunk failures (NOT introduced by this SPEC; attribution confirmed by `git stash` of pre-existing statusline/version changes + re-test on base HEAD d33a8700f — tests still FAIL without our changes): `internal/cli` `TestDoctor_*` / `TestStatus_*` golden-snapshot drift + `TestHookWrapper_TempFileCleanup` (slow/flaky). PRESERVE — B10 scope discipline.

## §E.4 Sync-phase Audit-Ready Signal

Sync-phase orchestrator-direct (manager-docs CHANGELOG-thrashing fallback per `feedback_glm_orchestrator_direct_sync_mx`). 3-phase close per Status Transition Ownership Matrix: spec/plan/acceptance/progress.md frontmatter `status: draft → completed` atomic on the single sync commit + `updated: 2026-07-08` + this §E.4 signal + CHANGELOG `### Fixed` entry. Specific-path commit (this SPEC's change set only — pre-existing dirty statusline/version/PERF-001 files excluded per `feedback_shared_checkout_concurrent_commit_race`). plan-auditor iter-2 PASS 0.90 (Tier L threshold 0.85). `sync_commit_sha` backfilled via follow-up chore commit per established convention.

sync_commit_sha: d0d953894

## §F Phase 0.95 Mode Selection

Input parameters: tier=L, scope=42 files (29 REQ-bearing), domains=multi (skill md + template mirror + 1 Go test), file mix=~98% markdown + 1 Go test, concurrency benefit=HIGH (file-disjoint write-groups), team prereqs=not required.

Mode evaluation:
- Mode 1 trivial: not selected (42-file scope)
- Mode 2 background: not selected (write work)
- Mode 3 agent-team: not selected (team.enabled default false; no capability gate)
- Mode 4 parallel: SELECTED — 8 file-disjoint write-groups (WG-A..WG-H), fanned out in 2 waves of ≤5 concurrent Agent() per the 3-5 ceiling; WG-I sequential last.
- Mode 5 sub-agent: fallback (not needed — disjoint groups make parallel safe)
- Mode 6 workflow: not selected (semantic doc edits, not a single uniform mechanical transform)

Decision: parallel

Justification: The plan defines 9 file-disjoint write-groups; no two touch the same file, so parallel multi-Agent() spawn is write-conflict-safe (Mode 4). Respecting Anthropic's 3-5 concurrent ceiling, WG-A..WG-H run in 2 waves (5+3); WG-I (CI regex test) runs sequentially last after make build + go test confirm M1-M3 landed. make build is serialized to a single orchestrator run after all doc edits to avoid embedded-asset race.
