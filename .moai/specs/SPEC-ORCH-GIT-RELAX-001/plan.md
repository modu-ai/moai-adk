# plan.md — SPEC-ORCH-GIT-RELAX-001

> Tier L implementation plan. Milestones ordered by **decision-reversibility** (most likely-to-change decisions first: user-facing doctrine + policy table; then agent-definition + delegation config; then mechanical Go verification; then tests + CI gates last). Per the plan.md ordering rule, this sequencing front-loads the decisions a human reviewer most needs to read.

---

## §A. Context

This SPEC relaxes the "always delegate push+PR to manager-git" doctrine. State-sensitive git ops + Tier S/M push+PR move to the orchestrator directly (full context: live HEAD, active worktree registry, concurrent-session state, transcript). `manager-git` retains Tier L release / multi-step merge / Late-Branch 4-Phase closure.

The relaxation is justified by the **context-sensitivity inversion** principle (see `design.md`): state-sensitive ops require the actor with maximum live state; isolated subagents (with only a stale session-start snapshot) are structurally wrong for that class. PR #1338 is the incident that proved the paradox (evidence report §0).

The change-surface is 12 doctrine locations (evidence §4) + 1 Go verification site. The Go exemption path (`isExemptAgent`, `internal/hook/branch_guard.go:144-155`) ALREADY admits the env-var predicate (`MOAI_BRANCH_GUARD_EXEMPT=1`) before the identity check — so orchestrator-direct Tier S/M is expected to be admitted with NO Go change. M3 verifies this; minimal change only if verification fails.

---

## §B. Known Issues / [NEEDS CLARIFICATION]

- **[NEEDS CLARIFICATION: branch-guard opt-in default]** — the distributed default for `Workflow.BranchGuard.Enabled` is `false` (per SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001). Orchestrator-direct Tier S/M relies on `MOAI_BRANCH_GUARD_EXEMPT=1` being meaningful, which presupposes the guard is enabled at least for the maintainer checkout. Is the opt-in gate expected to be on for this project (so the env path is load-bearing), or is it off (so the env path is dormant and the relaxation needs no guard interaction at all)? **Resolve at Implementation Kickoff Approval.** Either answer is acceptable; the run-phase plan differs:
  - If enabled → M3 verifies the env predicate admits orchestrator-direct Tier S/M.
  - If disabled → M3 documents that the guard is inert for the project and the relaxation's exemption-path clause is defense-in-depth (no runtime effect on this checkout; effective for any downstream maintainer checkout that enables the guard).

- **RESOLVED IN-SPEC (iter-2, D4): orchestrator-held env-var lifecycle = per-invocation inline.** `MOAI_BRANCH_GUARD_EXEMPT=1` SHALL be set per-Bash-invocation inline (e.g., `MOAI_BRANCH_GUARD_EXEMPT=1 git switch -c feat/SPEC-XXX`). Rationale: `design.md` §3.3 commits to per-invocation inline as the narrowest blast radius; `plan.md` §G REJECTS the persistent-settings alternative (which would exempt every orchestrator Bash call and defeat the guard). The per-sequence and persistent alternatives are rejected; no Implementation Kickoff input needed.

---

## §C. Pre-flight (before M1)

- Read `.moai/reports/agent-skill-hook-redesign-evidence-20260804.md` §4 (verbatim change-surface locations) and confirm each file:line still resolves at run-phase start (the report was cut at `c7b61777b`; a newer commit may have shifted lines).
- Confirm `make build` is green and `go test ./internal/hook/...` passes before any doctrine edit (baseline for M4 regression coverage).
- Confirm the Pre-Spawn/Pre-Edit Sync Check section is still present verbatim in `agent-common-protocol.md` (M2 depends on its current text).

---

## §D. Constraints

- **Template-First mirror obligation**: every change to a template-mirrored file (CLAUDE.md, `.claude/agents/moai/manager-git.md`, `.claude/rules/moai/core/agent-common-protocol.md`, `.claude/rules/moai/workflow/main-checkout-branch-guard.md`, `.moai/config/sections/delegation.yaml`) MUST be mirrored to `internal/template/templates/` counterpart. The mirror edit is in the SAME milestone as the primary edit (not deferred).
- **LOCAL-ONLY no-mirror obligation**: edits to `CLAUDE.local.md`, `.moai/docs/git-local-workflow-doctrine.md`, `.moai/docs/repo-local-pr-policy.md` MUST NOT be mirrored (no template counterpart exists by design — §2 local-only-files list).
- **No Late-Branch body rewrite**: Phases A–D in `manager-git.md` § Late-Branch Invocation Pattern stay verbatim (REQ-OGR-008, AC-OGR-007 byte-identity check).
- **No frontmatter change to manager-git**: `model: sonnet`, `effort: low`, `permissionMode: bypassPermissions` all retained (out of scope per §F).
- **Conventional-commits + `🗿 MoAI` trailer** on every doctrine-edit commit. Plan-phase artifacts use `feat(SPEC-ORCH-GIT-RELAX-001): plan-phase artifacts (...)`. Per run-phase milestones, each milestone commit is `feat(SPEC-ORCH-GIT-RELAX-001): M{n} <subject>` or `docs(...)` as appropriate.

---

## §E. Self-Verification (run-phase §E binding)

- **§E.1 Plan-phase Audit-Ready Signal** — populated at plan-phase close (this file + spec.md + acceptance.md + research.md + design.md + progress.md all authored).
- **§E.2 Run-phase Evidence** — manager-develop populates (per-milestone command + verbatim output).
- **§E.3 Run-phase Audit-Ready Signal** — manager-develop populates.
- **§E.4 Sync-phase Audit-Ready Signal** — manager-docs populates at sync commit.

---

## §F. Milestones

### M1 — Doctrine core (user-facing policy + HARD rule + Tier table)

**Decision-reversibility: HIGHEST.** These are the user-facing policy statements and the HARD rule (§23.7/§23.9 of git-local-workflow-doctrine.md). A human reviewer must read these first.

**Files edited (6 edits across 4 files)**:
1. `CLAUDE.md` §4 "Selection Decision Tree" row 6 — relax "Tier L OR `--pr` → manager-git" to "Tier L OR `--pr` → manager-git; Tier S/M push+PR → orchestrator-direct".
2. `CLAUDE.md` §4 catalog table row for `manager-git` — refresh the "Phase scope" cell.
3. `.moai/docs/git-local-workflow-doctrine.md` §23.9 tier table (rows S/M) — owner cell flips from `manager-git` to `orchestrator-direct (MOAI_BRANCH_GUARD_EXEMPT=1 + Pre-Spawn Sync Check)`. Tier L row unchanged.
4. `.moai/docs/git-local-workflow-doctrine.md` §23.9 prose around REQ-ATR-020 — add the Tier S/M carve-out clause. Cite SPEC-ORCH-GIT-RELAX-001 as the amending SPEC.
5. `.moai/docs/git-local-workflow-doctrine.md` §23.7 [HARD] bullet — soften "모든 tier PR 경유 manager-git" to "Tier L / `--pr` PR 경유 manager-git; Tier S/M orchestrator 직접".
6. `.claude/rules/moai/workflow/spec-workflow.md:26` Route B description (**change-surface location #13 — iter-2 addition**) — qualify the Route B actor: the current "`manager-git` creates a feature branch and opens a PR per phase" becomes "Tier L / `--pr`: `manager-git` creates a feature branch and opens a PR per phase; Tier S/M Route B (per `repo-local-pr-policy.md` ALL-tier Route B override): orchestrator-direct per SPEC-ORCH-GIT-RELAX-001". WITHOUT this edit, the repo-local ALL-tier Route B override would textually assign Tier S/M Route B PRs to manager-git even post-relaxation.

**Mirror obligation**: CLAUDE.md → `internal/template/templates/CLAUDE.md` mirrored edit. `.claude/rules/moai/workflow/spec-workflow.md` → `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` mirrored edit (**template-mirrored — location #13 carries mirror obligation**). git-local-workflow-doctrine.md is LOCAL-ONLY — no mirror.

**Verify**:
```bash
grep -n "manager-git" CLAUDE.md internal/template/templates/CLAUDE.md
grep -n "manager-git" .moai/docs/git-local-workflow-doctrine.md
grep -n "manager-git\|orchestrator-direct" .claude/rules/moai/workflow/spec-workflow.md internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md
make build
```

**Commit**: `feat(SPEC-ORCH-GIT-RELAX-001): M1 doctrine core — relax Tier S/M push+PR to orchestrator-direct`

---

### M2 — Agent definition + delegation config + sync-check context

**Decision-reversibility: HIGH.** The `manager-git` invocation gate (description + body) and delegation.yaml note define when the agent fires. Wrong text here breaks both directions (over-firing manager-git OR under-firing it for Tier L).

**Files edited (4 edits across 4 files)**:
1. `.claude/agents/moai/manager-git.md` frontmatter `description:` — rewrite the invocation gate clause: "Tier L / `--pr` delegated; Tier S/M orchestrator-direct per SPEC-ORCH-GIT-RELAX-001". Late-Branch body (§ Late-Branch Invocation Pattern, Phases A–D) UNCHANGED — byte-identity preserved (AC-OGR-007).
2. `.moai/config/sections/delegation.yaml:74` — refresh the inline comment to "manager-git is Tier-L / `--pr` only; Tier S/M is orchestrator-direct (SPEC-ORCH-GIT-RELAX-001)".
3. `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn Sync Check + § Pre-Edit Sync Check — add a note that these checks are the gate for orchestrator-direct Tier S/M push+PR (the env-path exemption is predicated on the sync check passing).
4. `.claude/rules/moai/workflow/main-checkout-branch-guard.md` § Mechanical Enforcement + § Pattern refinement — document that orchestrator-direct Tier S/M uses the `MOAI_BRANCH_GUARD_EXEMPT=1` env path, which `isExemptAgent` already admits before the `AgentType == "manager-git"` identity branch (so no Go change expected).

**Mirror obligation**:
- `.claude/agents/moai/manager-git.md` → `internal/template/templates/.claude/agents/moai/manager-git.md`
- `.claude/rules/moai/core/agent-common-protocol.md` → `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md`
- `.claude/rules/moai/workflow/main-checkout-branch-guard.md` → `internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard.md`
- `.moai/config/sections/delegation.yaml` → `internal/template/templates/.moai/config/sections/delegation.yaml`

**Verify**:
```bash
# Late-Branch body byte-identity (AC-OGR-007)
diff <(sed -n '/### Late-Branch Invocation Pattern/,/^## Mode-Specific/p' .claude/agents/moai/manager-git.md) \
     <(sed -n '/### Late-Branch Invocation Pattern/,/^## Mode-Specific/p' internal/template/templates/.claude/agents/moai/manager-git.md)
# Mirror presence
for f in .claude/agents/moai/manager-git.md .claude/rules/moai/core/agent-common-protocol.md .claude/rules/moai/workflow/main-checkout-branch-guard.md .moai/config/sections/delegation.yaml; do
  test -f "internal/template/templates/$f" && echo "MIRROR OK: $f" || echo "MIRROR MISSING: $f"
done
make build
```

**Commit**: `feat(SPEC-ORCH-GIT-RELAX-001): M2 agent def + delegation + sync-check context`

---

### M3 — Go branch-guard exemption verification (minimal change)

**Decision-reversibility: MEDIUM.** The hypothesis (env path already admits orchestrator-direct Tier S/M, no Go change) is evidence-based (the code is already in context). M3 confirms by running the existing branch-guard test suite + targeted predicate test.

**Procedure**:
1. Run `go test ./internal/hook/... -run BranchGuard` — confirm the `isExemptAgent` env-var branch is covered and green.
2. Add a focused test (if not already present) that exercises `isExemptAgent(&HookInput{AgentType: ""})` with `MOAI_BRANCH_GUARD_EXEMPT=1` set — assert it returns `true` via the env branch, NOT via the identity branch. This is the mechanical proof of REQ-OGR-011.
3. If the test fails (env path NOT admitting) — minimal Go change: ensure `os.Getenv(branchGuardExemptEnv) == "1"` returns true unconditionally (it already does per `branch_guard.go:145-147`). Do NOT add a new identity branch for "orchestrator" — the relaxation's invariant is that the env path is the sole admission mechanism for orchestrator-direct ops.
4. If the test passes — no Go change. Document the verification result in `progress.md` §E.2.

**Mirror obligation**: none (Go code is not template-distributed).

**Verify**:
```bash
go test ./internal/hook/... -run BranchGuard -count=1 -v
go test ./internal/hook/... -run TestIsExemptAgent -count=1 -v
```

**Commit**: `test(SPEC-ORCH-GIT-RELAX-001): M3 verify branch-guard env exemption admits orchestrator-direct Tier S/M`
—or— `fix(SPEC-ORCH-GIT-RELAX-001): M3 minimal Go change to admit orchestrator-direct Tier S/M via env path` (only if verification fails).

---

### M4 — Regression test (PR-#1338 incident class) + full gate

**Decision-reversibility: LOWEST.** Mechanical: add the regression test, run the full gate, confirm all 8 ACs.

**M4.1 — Regression test**: add a Go test that reproduces the PR-#1338 incident class signature: a multi-worktree state op where (a) an isolated snapshot-holder (simulating the old manager-git behavior) would cross-swap, and (b) the orchestrator-with-context path (reading live HEAD + worktree registry) restores primary→main correctly. The test MUST demonstrate the orchestrator-direct path resolves the incident.

The test belongs in `internal/hook/branch_guard_test.go` or a sibling file (the branch-guard is the closest mechanical analog). The test is a unit test (no real git operations against the primary checkout — use `t.TempDir()`).

**M4.2 — Doctrine cross-reference audit (AC-OGR-003)**:
```bash
# Zero residual "all tiers route push+PR through manager-git" outside the Tier L / --pr carve-out
grep -rn "all tiers\|모든 tier.*PR.*manager-git\|push + PR is always delegated\|manager-git.*creates a feature branch and opens a PR per phase" \
  CLAUDE.md .claude/agents/moai/manager-git.md \
  .claude/rules/moai/core/agent-common-protocol.md \
  .claude/rules/moai/workflow/main-checkout-branch-guard.md \
  .claude/rules/moai/workflow/spec-workflow.md \
  .moai/config/sections/delegation.yaml \
  .moai/docs/git-local-workflow-doctrine.md \
  internal/template/templates/CLAUDE.md \
  internal/template/templates/.claude/agents/moai/manager-git.md \
  internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md \
  | grep -v "Tier L\|--pr\|carve-out\|RELAX-001"
# Expected: 0 matches (the grep -v excludes the legitimate Tier L / --pr mentions)
```

**M4.3 — LOCAL-ONLY no-mirror audit (AC-OGR-008)**:
```bash
# Confirm CLAUDE.local.md, git-local-workflow-doctrine.md, repo-local-pr-policy.md have NO template counterpart
for f in CLAUDE.local.md .moai/docs/git-local-workflow-doctrine.md .moai/docs/repo-local-pr-policy.md; do
  test ! -e "internal/template/templates/$f" && echo "NO-MIRROR OK: $f" || echo "NO-MIRROR VIOLATED: $f"
done
```

**M4.4 — Full gate (AC-OGR-005)**:
```bash
go test ./...                                 # all tests
go test ./internal/hook/... -count=1          # branch-guard suite
go test ./internal/template/... -count=1      # template-neutrality + mirror-parity + commands_audit + output_styles_audit
moai agent lint                              # agent catalog lint
moai workflow lint                           # workflow lint
make build                                   # embed regeneration
golangci-lint run                            # Go lint
```

**Commit**: `test(SPEC-ORCH-GIT-RELAX-001): M4 regression coverage + full gate green`

---

## §G. Anti-Patterns (rejected approaches)

- **REJECTED — abolish manager-git entirely.** The relaxation sculpts the responsibility surface; it does not remove the agent. manager-git demonstrably performs standard push+PR well (PRs #1319/#1321/#1326/#1330) and owns the Late-Branch 4-Phase closure (Tier L release). Abolition would lose canonical owners for genuine complex flows. Evidence §0 names the paradox, not the agent.
- **REJECTED — add an "orchestrator" identity branch to `isExemptAgent`.** The env path (`MOAI_BRANCH_GUARD_EXEMPT=1`) already admits orchestrator-direct ops unconditionally. Adding an identity branch would widen the exemption surface and couple the guard to orchestrator identity (which is not an `AgentType` — the orchestrator is the main thread). REQ-OGR-012 forbids this.
- **REJECTED — set `MOAI_BRANCH_GUARD_EXEMPT=1` persistently in settings.** This would exempt every orchestrator Bash call from the branch-state guard, defeating the guard entirely. The exemption is per-invocation (inline `MOAI_BRANCH_GUARD_EXEMPT=1 git switch -c ...` — the narrowest blast radius per `design.md` §3.3). *(iter-2: this rejection resolves plan.md §B marker #2 IN-SPEC — per-invocation inline is mandated; the `[NEEDS CLARIFICATION]` marker is removed.)*
- **REJECTED — rewrite the Late-Branch Phases A–D for the new world.** Out of scope (§F). The 4-Phase closure is Tier L only and stays verbatim.
- **REJECTED — bundle skills/hooks cleanup into this SPEC.** Out of scope (§F). Skills and hooks cleanup are separate follow-up SPECs ("순차 분할" — the user's phased-split directive).
- **REJECTED — flip the distributed default of `Workflow.BranchGuard.Enabled` to `true`.** Out of scope (§F). The relaxation uses the env exemption path; it does not change the opt-in gate.

---

## §H. Cross-References

- `spec.md` §D Table 1 — the 12 change-surface locations.
- `acceptance.md` — Given-When-Then scenarios for the 8 ACs.
- `research.md` — evidence report §4 + §5 incorporation + 12 verified URLs.
- `design.md` — the context-sensitivity inversion principle + 7-principle gating.
- Evidence base: `.moai/reports/agent-skill-hook-redesign-evidence-20260804.md`.
- Precedent: `SPEC-WORKTREE-BRANCH-GUARD-001`, `SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001`, `SPEC-V3R6-AGENT-TEAM-REBUILD-001` (REQ-ATR-020).
