# progress.md — SPEC-ORCH-GIT-RELAX-001

> Tier L progress artifact. §E section skeleton populated at plan-phase close; §E.2–§E.4 are placeholder-only per the manager-spec §E skeleton generation protocol.

---

## §A. SPEC summary

- **ID**: SPEC-ORCH-GIT-RELAX-001
- **Title**: Orchestrator-direct Tier S/M git ops + state-sensitive worktree recovery (manager-git relaxation)
- **Tier**: L (6-artifact set: spec.md + plan.md + acceptance.md + research.md + design.md + progress.md)
- **Phase**: plan (artifacts authored; awaiting plan-auditor + Implementation Kickoff Approval)
- **Scope**: relax "always delegate push+PR to manager-git" — Tier S/M push+PR + state-sensitive git ops move to orchestrator-direct; manager-git RETAINED for Tier L / `--pr` / Late-Branch 4-Phase closure.
- **Change-surface**: 14 enumerated locations (13 edit + 1 verified no-change; evidence §4's original 12 + iter-2 additions: `spec-workflow.md` Route B edit + `CLAUDE.md:131` no-change).
- **Triggering incident**: PR #1338 handling (manager-git primary→main restore failure + concurrent-session worktree cross-swap).

---

## §B. Plan-phase artifact manifest

- [x] `spec.md` — 16 REQ-OGR requirements (GEARS notation), 10 AC tokens (iter-2: +AC-OGR-009/010), §D Table 1 (14 enumerated: 13 edit + 1 no-change), §F Out of Scope (6 H3 sub-headings).
- [x] `plan.md` — 4 milestones (M1 doctrine core / M2 agent def + delegation / M3 Go verification / M4 regression + full gate), ordered by decision-reversibility. 1 `[NEEDS CLARIFICATION]` marker carried to Implementation Kickoff Approval (iter-2: marker #2 resolved IN-SPEC).
- [x] `acceptance.md` — 10 AC-OGR tokens (all MUST severity, iter-2: +009/010), Given-When-Then scenarios, §D Traceability (16/16 REQ-mapped, iter-2 closed 5-gap deficit), §E Definition of Done, §F edge cases.
- [x] `research.md` — evidence §4 + §5 incorporation + iter-2 audit (locations #13/#14), 12 verified URLs, §10 open questions (iter-2: item #2 resolved IN-SPEC).
- [x] `design.md` — context-sensitivity inversion principle, 7-principle gate scorecard, 6 rejected alternatives.
- [x] `progress.md` — this file (§E skeleton).

---

## §C. Pre-plan-phase audit checks (run by manager-spec before §E.1 sign-off)

- [x] SPEC ID regex check: `SPEC-ORCH-GIT-RELAX-001` → `PASS` (executed via Bash, verbatim output cited).
- [x] ID uniqueness: no existing SPEC under `.moai/specs/` collides with `SPEC-ORCH-GIT-RELAX-001` (closest neighbors: `SPEC-V3R6-ORCH-IGGDA-001`, `SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001`, `SPEC-WORKTREE-BRANCH-GUARD-001` — different domain tokens, no collision).
- [x] Frontmatter 12 canonical fields present (id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags). `phase: "v14.4.0 target"` — release target, not a lifecycle token; passes the prohibited-phase-value check.
- [x] Requirements in GEARS notation (Ubiquitous / When / While / Where / shall not) — no residual IF/THEN modality.
- [x] Out of Scope section satisfies `OutOfScopeRule`: 6 `### Out of Scope — <topic>` H3 sub-headings, each with ≥1 `-` bullet.
- [x] Artifact set matches Tier L (6 artifacts: spec + plan + acceptance + research + design + progress).
- [x] spec.md carries no implementation detail (no function names, no API schemas — REQ tokens only).

---

## §D. Open items (resolved at Implementation Kickoff Approval)

- [ ] **branch-guard opt-in default state** — is `Workflow.BranchGuard.Enabled` on or off for the maintainer checkout? (plan.md §B, research.md §10 item 1)
- [x] **`MOAI_BRANCH_GUARD_EXEMPT=1` lifecycle** — RESOLVED IN-SPEC (iter-2, D4): per-invocation inline mandated per design.md §3.3 + plan.md §G.

One operational detail remains, carried to Implementation Kickoff Approval. Neither blocks plan-phase audit-readiness.

---

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-04
spec_id: SPEC-ORCH-GIT-RELAX-001
tier: L
artifact_count: 6
req_count: 16
ac_count: 10   # iter-2: +AC-OGR-009 (catalog+frontmatter preservation), +AC-OGR-010 (foreign-session auto-isolation)
open_clarifications: 1   # iter-2: marker #2 (env-var lifecycle) resolved IN-SPEC per design.md §3.3; marker #1 (branch-guard opt-in) remains
evidence_base: .moai/reports/agent-skill-hook-redesign-evidence-20260804.md
notes: >
  Tier L redesign SPEC, first of a phased split ("순차 분할").
  manager-git relaxation only; skills/hooks cleanup in follow-up SPECs.
  Change-surface is 14 enumerated locations (13 edit + 1 verified no-change):
  evidence §4's original 12 + iter-2 additions (spec-workflow.md Route B edit;
  CLAUDE.md:131 no-change). Env exemption path expected to admit
  orchestrator-direct Tier S/M with no Go change.
  Iter-2 plan-auditor FAIL 0.79 → 5 defects D1-D5 addressed (no commit/push).
```

---

## §E.2 Run-phase Evidence

Run-phase executed 2026-08-05 on `feat/SPEC-ORCH-GIT-RELAX-001` (HEAD stacked on plan-phase commit `7fba13b3d` off origin/main `1579687e6`). 4 milestone commits, no push (Route B Tier L — manager-git owns push+PR at sync).

### Milestone commits

| M | SHA | Subject | Cycle |
|---|-----|---------|-------|
| M1 | `7cca3be25` | `feat(SPEC-ORCH-GIT-RELAX-001): M1 doctrine core — relax Tier S/M push+PR to orchestrator-direct` | doctrine edits (6 across 4 files) + 2 template mirrors + spec.md `draft→in-progress` |
| M2 | `d9362a4d0` | `feat(SPEC-ORCH-GIT-RELAX-001): M2 agent def + delegation + sync-check context` | 4 edits + 4 template mirrors; Late-Branch body byte-identity preserved |
| M3 | `6e3f0662c` | `test(SPEC-ORCH-GIT-RELAX-001): M3 verify branch-guard env exemption admits orchestrator-direct Tier S/M` | characterization test only — no Go source change (`git diff internal/hook/branch_guard.go` empty) |
| M4 | `898d2b854` | `test(SPEC-ORCH-GIT-RELAX-001): M4 regression coverage + full gate green` | PR-#1338 regression test + spec-workflow byte-parity genericization |

### AC PASS/FAIL matrix (acceptance.md SSOT)

| AC | Status | Verifying command | Result |
|----|--------|-------------------|--------|
| AC-OGR-001 | PASS (doctrine) | `grep -n "orchestrator-direct\|MOAI_BRANCH_GUARD_EXEMPT" CLAUDE.md .moai/docs/git-local-workflow-doctrine.md .claude/rules/moai/workflow/spec-workflow.md` | Tier S/M orchestrator-direct path documented at all 3 user-facing surfaces; the live push+PR execution itself is sync-phase (manager-git / orchestrator runtime behavior, not mechanically reproducible in run-phase — doctrine wiring is the run-phase deliverable) |
| AC-OGR-002 | PASS | `grep -rn "manager-git" CLAUDE.md .claude/agents/moai/manager-git.md .moai/docs/git-local-workflow-doctrine.md .moai/config/sections/delegation.yaml` | manager-git retained ONLY for Tier L / `--pr` / multi-step merge / Late-Branch 4-Phase; Tier S/M carve-out explicit |
| AC-OGR-003 | PASS | M4.2 grep audit (below) | 0 matches outside Tier L / `--pr` carve-out |
| AC-OGR-004 | PASS | `go test ./internal/hook/ -run TestIsExemptAgent -count=1 -v` | `--- PASS: TestIsExemptAgent_OrchestratorDirectEnvPath` (2 subtests green); `git diff HEAD~3 -- internal/hook/branch_guard.go` empty (no Go change) |
| AC-OGR-005 | PASS | full gate (M4.4 below) | every command exit 0 |
| AC-OGR-006 | PASS | `go test ./internal/hook/ -run TestPR1338Regression -count=1 -v` | `--- PASS: TestPR1338Regression` — stale path leaves primary broken (RED evidence), live path restores primary + leaves wt-A undisturbed (GREEN resolution) |
| AC-OGR-007 | PASS | `diff <(sed -n '/### Late-Branch Invocation Pattern/,/^## Mode-Specific/p' .claude/agents/moai/manager-git.md) <(...)` | empty diff (byte-identical local vs template); `git diff HEAD~3 -- .claude/agents/moai/manager-git.md` shows only the frontmatter `description:` block changed |
| AC-OGR-008 | PASS | M4.3 no-mirror audit (below) | 3/3 `NO-MIRROR OK` |
| AC-OGR-009 | PASS | `grep -c "manager-git" CLAUDE.md` → ≥1; frontmatter diff | `manager-git` retained in CLAUDE.md §4 catalog (count ≥1); `model: sonnet` / `effort: low` / `permissionMode: bypassPermissions` unchanged (only `description:` edited) |
| AC-OGR-010 | PASS (doctrine) | `grep -n "Pre-Spawn Sync Check\|auto-isolate" .claude/rules/moai/core/agent-common-protocol.md` | the M2 note wires the Pre-Spawn/Pre-Edit Sync Check as the gate for orchestrator-direct Tier S/M push+PR; halt+auto-isolate route documented (live foreign-session detection is orchestrator runtime behavior) |

### M4.2 doctrine cross-reference audit (AC-OGR-003)

```bash
$ grep -rn "all tiers\|모든 tier.*PR.*manager-git\|push + PR is always delegated\|manager-git.*creates a feature branch and opens a PR per phase" \
    CLAUDE.md .claude/agents/moai/manager-git.md \
    .claude/rules/moai/core/agent-common-protocol.md \
    .claude/rules/moai/workflow/main-checkout-branch-guard.md \
    .claude/rules/moai/workflow/spec-workflow.md \
    .moai/config/sections/delegation.yaml \
    .moai/docs/git-local-workflow-doctrine.md \
    internal/template/templates/CLAUDE.md \
    internal/template/templates/.claude/agents/moai/manager-git.md \
    internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md \
  | grep -v "Tier L\|--pr\|carve-out\|RELAX-001\|RETIRED"
# exit 1 (grep -v produced 0 matches — PASS)
```

### M4.3 LOCAL-ONLY no-mirror audit (AC-OGR-008)

```
NO-MIRROR OK: CLAUDE.local.md
NO-MIRROR OK: .moai/docs/git-local-workflow-doctrine.md
NO-MIRROR OK: .moai/docs/repo-local-pr-policy.md
```

### M4.4 full gate (AC-OGR-005) — verbatim exit codes

```
$ go test ./... -count=1            → 38 ok packages, 0 FAIL  (exit 0)
$ go test ./internal/hook/...       → all ok                  (exit 0)
$ go test ./internal/template/...   → ok                      (exit 0)
$ moai agent lint                   → 0 errors, 24 warnings   (exit 0; all warnings pre-existing LR-05/LR-08, none introduced by this SPEC)
$ moai workflow lint                → No violations found     (exit 0)
$ golangci-lint run --timeout=3m    → 0 issues                (exit 0)
$ go build ./...                    → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
$ make build                        → exit 0
$ moai spec lint spec.md            → 0 errors, 1 warning (StatusGitConsistency — see Gaps)
```

### M3 verification result (REQ-OGR-011 / REQ-OGR-012)

`isExemptAgent` env branch (`internal/hook/branch_guard.go:145`, `os.Getenv(branchGuardExemptEnv) == "1"`) returns true BEFORE the `AgentType == "manager-git"` identity branch (line 154). Orchestrator-direct Tier S/M (AgentType="") admitted via env sentinel with NO Go change. REQ-OGR-012 (no widening) confirmed: env-unset + AgentType="" → denied (no identity branch for "orchestrator" added). The guard is dormant on this checkout (`Workflow.BranchGuard.Enabled=false` distributed default per `internal/config/defaults.go:598`), so the env path is defense-in-depth here; the test + doctrine cover any downstream maintainer checkout that enables the guard.

### E8 TDD RED evidence (M4 regression test)

The PR-#1338 regression test (`TestPR1338Regression`) carries both the RED and GREEN in one run: the stale-snapshot subtest asserts the primary is NOT restored (`staleRestored == false` — the cross-swap / left-broken failure mode reproduces), and the live-state subtest asserts the primary IS restored (`liveRestored == true` — orchestrator-direct resolves it). Verbatim:

```
=== RUN   TestPR1338Regression
--- PASS: TestPR1338Regression (0.96s)
PASS
```

The RED is structural (the stale path's `staleRestored` assertion is the failing condition that the incident class forces; the live path's `liveRestored` is the resolution). There is no separate pre-GREEN RED run because the "implementation" being verified is the orchestrator-direct DATA-READ pattern itself (live `git worktree list --porcelain` vs frozen snapshot), not a new Go code path — the test passes against real git behavior on the first run.

---

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-05
run_commit_sha: "898d2b854"   # M4 (final run-phase commit); M1=7cca3be25, M2=d9362a4d0, M3=6e3f0662c
run_status: audit-ready
spec_id: SPEC-ORCH-GIT-RELAX-001
tier: L
ac_pass_count: 10
ac_fail_count: 0
preserve_list_post_run_count: 0   # no PRESERVE-list items outstanding
l44_pre_commit_fetch: n/a (Route B Tier L worktree; no pre-commit push — manager-git owns push+PR)
l44_post_push_fetch: n/a (no push performed; Route B)
new_warnings_or_lints_introduced: 0   # moai agent lint 24 warnings all pre-existing (LR-05/LR-08); golangci-lint 0 issues; moai workflow lint clean
cross_platform_build:
  darwin_arm64: ok
  linux_amd64: ok (make build default target)
  windows_amd64: ok (GOOS=windows GOARCH=amd64 go build ./... exit 0)
total_run_phase_files: 16   # 6 doctrine local (CLAUDE.md, git-local-workflow-doctrine.md x4 edits, spec.md frontmatter) + 6 template mirrors + 1 delegation.yaml local + 1 delegation.yaml mirror + 2 Go test files + spec.md frontmatter; counted as distinct file paths touched
m1_to_mN_commit_strategy: per-milestone Conventional Commit (feat/test), each with 🗿 MoAI trailer; no --amend, no force-push, no --no-verify
notes: >
  Run-phase complete. 4 milestone commits on feat/SPEC-ORCH-GIT-RELAX-001
  (no push — Route B Tier L). M3 confirmed env exemption path admits
  orchestrator-direct with zero Go change. M4 regression test mechanically
  reproduces PR-#1338 incident class and demonstrates the orchestrator-direct
  resolution. All 10 AC PASS. Sync-phase (manager-docs) owns §E.4 + the
  in-progress→implemented→completed transition + push+PR via manager-git
  (Tier L). Open items for sync: (a) catalog.yaml sync-auditor hash drift
  observed pre-run (unrelated to this SPEC — left unstaged per B10 scope
  discipline); (b) moai spec lint StatusGitConsistency warning (in-progress
  vs git-implied draft — resolves at sync status finalization).
```

---

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-05
sync_commit_sha: "53572e4b6"   # backfilled post-merge — sync PR #1351 squash on main (2026-08-04)
run_commit_sha: "898d2b854"                # M4 run-phase tip; squash-landed on main as 876e9936142a (PR #1347)
run_pr: 1347
sync_status: audit-ready
spec_id: SPEC-ORCH-GIT-RELAX-001
tier: L
ac_pass_count: 10
ac_fail_count: 0
frontmatter_transition: "in-progress → completed (single sync commit, 3-phase close)"
notes: >
  Sync-phase closed the 3-phase lifecycle (plan→run→sync). CHANGELOG entry
  added under [Unreleased] ### Changed. spec.md frontmatter transitioned
  in-progress → completed (version 0.1.1 → 0.2.0, updated 2026-08-05). §E.4
  populated. catalog.yaml was regenerated in run-phase (commit 2f54dd58f,
  included in the PR #1347 squash 876e9936142a) and is unchanged in sync —
  no template edits, no `make build` in this phase. The run-phase doctrine
  + Go test edits are already on main via PR #1347; this sync commit
  carries only frontmatter + progress.md §E.4 + CHANGELOG. Open post-merge
  item: `sync_commit_sha` backfill (D3 placeholder → real merged SHA) by
  the orchestrator after the sync PR merges.
```

---

## §F. (reserved — Phase 4 Mode Selection is logged here by the orchestrator before the first run-phase Agent() spawn)

**Decision: sub-agent (Mode 5)** — selected 2026-08-04, before first run-phase Agent() spawn.

Input parameters:
- tier: L
- scope: ~14 change-surface locations (13 edit + 1 verified no-change) across ~12 files + 4 template mirrors + 1 Go test
- domain count: 5 (doctrine/rules, agent-definition, config, local-docs, Go-test) — ≥3 domains threshold MET
- file language mix: markdown-heavy + YAML + 1 Go test
- concurrency benefit: LOW (coding-heavy; sequential M1→M2→M3→M4 dependencies with cross-references between doctrine edits)
- Agent Teams prereqs: N/A (Mode 3 retired)

Mode evaluation:
| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | no | 14-location doctrinal change, not trivial |
| 2 background | no | write work, not read-only async |
| 3 agent-team | no | RETIRED |
| 4 parallel | no | ≥3 domains met BUT coding-heavy + sequential milestone deps (Anthropic coding-task parallelism caveat → Mode 5) |
| 5 sub-agent | YES | coding-heavy doctrinal change, sequential M1→M4, default fallback |
| 6 workflow | no | ~14 semantic edits, not ≥30-file uniform-mechanical-transform |

Justification: Tier L coding-heavy doctrinal relaxation with sequential milestone dependencies (M1 doctrine core → M2 agent def + delegation → M3 Go branch-guard verify → M4 regression + full gate). Per Anthropic's coding-task parallelism caveat, sequential sub-agent is the safe default. Mode 4 excluded despite ≥3 domains: the work is coding-heavy with ordered cross-references, not parallelizable research. Mode 6 excluded: ~14 semantic edits do not meet the ≥30-file uniform-mechanical-transform bar. Execution: sequential per-milestone manager-develop delegations (M1→M2→M3→M4), orchestrator verifies between; worktree-isolated (feat/SPEC-ORCH-GIT-RELAX-001 off origin/main 1579687e6 + plan cherry-pick 7fba13b3d).

---

## §G. Risk register

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Env exemption path does NOT admit orchestrator-direct (M3 fails) | Low (code already in context confirms env branch fires first) | Medium — minimal Go change required | M3 verification + REQ-OGR-012 minimal-change clause |
| Doctrine cross-reference drifts post-edit (M4.2 grep audit fails) | Medium (12 locations, easy to miss one) | High — inconsistent doctrine is worse than the prior doctrine | M4.2 grep audit + AC-OGR-003 binary check |
| Late-Branch body accidentally edited (AC-OGR-007 fails) | Low (milestone explicitly excludes it) | Medium — regression in canonical Tier L closure | AC-OGR-007 byte-identity diff |
| Template-mirror forgotten for a template-mirrored file | Medium (4 mirrored files across M1+M2) | High — CI template-neutrality / mirror-parity fails | M2 verify block lists all 4 mirrors; AC-OGR-005 full gate |
| LOCAL-ONLY file accidentally mirrored | Low (explicit no-mirror list) | Medium — violates §24/§25 | AC-OGR-008 no-mirror audit |
| Over-relaxation (manager-git abolished instead of sculpted) | Low (design.md §6.1 rejects) | High — loses Tier L owner | REQ-OGR-008 + REQ-OGR-016 forbid abolition |

---

## §H. Recursive Self-Diagnosis Log

_<pending run-phase — manager-develop / orchestrator (DIAGNOSE-PATCH-VERIFY mechanical failures)>_

---

## §I. Token Accounting

_<pending sync-close — manager-docs invokes the token-accounting writer>_
