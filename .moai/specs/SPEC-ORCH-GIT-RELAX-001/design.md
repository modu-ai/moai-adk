# design.md — SPEC-ORCH-GIT-RELAX-001

> Tier L design artifact. Records the design decision — the **context-sensitivity inversion principle** — and how the 7 delegation-vs-direct principles gate the delegation split. This is the design rationale; spec.md carries the requirements, plan.md carries the milestone execution.

---

## §1. The design decision — context-sensitivity inversion

### 1.1 Principle statement

> **State-sensitive operations SHALL be routed to the actor with maximum live state; context-insensitive recurring workflows SHALL be routed to isolated specialist agents.**

This is the **inversion** of the implicit prior doctrine, which routed ALL git operations to an isolated specialist (`manager-git`) regardless of the op's context-sensitivity. The prior doctrine was correct for the context-INSENSITIVE class (Tier L release, multi-step merge, Late-Branch 4-Phase — all config-driven and independently verifiable) and structurally wrong for the context-SENSITIVE class (Tier S/M push+PR on a shared primary checkout with concurrent sessions — op correctness depends on live HEAD + worktree registry that an isolated subagent cannot observe).

The inversion does NOT abolish the specialist. It sculpts the specialist's responsibility surface to the op class where isolation is an asset, and routes the op class where isolation is a liability to the full-context orchestrator.

### 1.2 Why "inversion"

The prior doctrine's assignment matrix was:

| Op class | Routed to | Reason (implicit) |
|----------|-----------|-------------------|
| All git ops (Tier S/M/L + Late-Branch + release + merge) | `manager-git` (isolated) | SRP — "each retained agent owns a phase" |

The inverted assignment matrix is:

| Op class | Routed to | Reason (explicit) |
|----------|-----------|-------------------|
| Tier S/M push+PR + state-sensitive recovery | **orchestrator-direct** | P1–P6 fail under delegation; the orchestrator is the only actor with live state |
| Tier L release / multi-step merge / Late-Branch 4-Phase | `manager-git` (isolated) | P1–P7 pass; isolation is an asset (config-driven, CI-verifiable, recurring-stable) |

The matrix is inverted only for the state-sensitive class. The rest is unchanged.

### 1.3 Reachability is not justification (the retention side)

Per the verification-claim-integrity doctrine (`§1.1 surface 4`): observing that `manager-git` is *referenced* (in doctrine, in delegation.yaml, in CLAUDE.md) establishes only that a reference exists; it does NOT establish that the referenced capability is the right owner for a given op class. The prior doctrine confused "manager-git exists and is the git specialist" with "manager-git is the right owner for ALL git ops." The inversion separates these: manager-git's existence justifies its retention for the op class where it earns its overhead (Tier L / Late-Branch), not its blanket assignment to every git op.

This is the symmetric of the "retention-claim hazard" documented in `verification-claim-integrity.md` §6: there, a reference was wrongly treated as evidence of a live feature; here, a reference was wrongly treated as evidence of the right owner. Both failures reduce to "a reference is not a justification."

---

## §2. The 7-principle gate (evidence §5.2)

The inversion is not ad-hoc. It is the output of applying the 7 delegation-vs-direct principles to each op class. The full scorecard is in `research.md` §5; the summary:

### 2.1 Tier S/M push+PR — orchestrator-direct (P1, P2, P3, P4, P5, P6 all FAIL delegation)

- **P1 context-sensitivity**: push+PR correctness depends on the orchestrator's live state (HEAD, active worktrees, concurrent-session registry). An isolated subagent holds a stale snapshot. FAIL → direct.
- **P2 decomposability**: push+PR is a 3-step state-coupled sequence (`git switch -c` → `git push -u` → `gh pr create`); decomposing it across a delegation boundary introduces a state-observation gap. FAIL → direct.
- **P3 45% baseline**: the orchestrator (with full state) is well above the 45% baseline; delegating reduces accuracy by handing off to a context-poor actor. FAIL → direct.
- **P4 20% coordination overhead**: subagent spawn costs ≫ 20% of a 3-line `gh pr create`. FAIL → direct.
- **P5 verification asymmetry**: verifying that the correct branch was pushed requires observing the live HEAD the orchestrator already holds; delegating verification to the subagent that lacks HEAD is circular. FAIL → direct.
- **P6 reversibility-weighted risk**: push is a shared-tree, semi-irreversible op (a bad push affects every concurrent session). The full-context actor must perform it. FAIL → direct.
- **P7 recurring stable pattern**: PARTIAL — Tier S/M push+PR recurs, but P1–P6 override. A static agent is only justified when the delegation-merit principles pass; recurrence alone is not sufficient.

### 2.2 Tier L release / multi-step merge / Late-Branch 4-Phase — manager-git retained (P1–P7 PASS)

- **P1**: release ops are config-driven (merge_method from git-strategy.yaml, release PR template, Late-Branch phases). State-light. PASS → delegate.
- **P2**: Late-Branch 4-Phase is decomposable (Phase A → B → C → D, each independently verifiable). PASS → delegate.
- **P3**: manager-git has the release playbook (checkpoint, merge, tag, cleanup); orchestrator does not retain it inline. PASS → delegate.
- **P4**: release is multi-hour; subagent spawn overhead is negligible against the task size. PASS → delegate.
- **P5**: CI gates verify mechanically (Test ubuntu-latest, Lint, Build linux/amd64, CodeQL — the 4 required checks). Orchestrator context not needed for verification. PASS → delegate.
- **P6**: release has a rollback playbook (tag revert, branch restore). Reversibility-weighted risk is bounded. PASS → delegate.
- **P7**: Late-Branch 4-Phase is stable and recurring across SPEC sessions — meets the "keep spawning the same kind of worker with the same instructions" criterion. PASS → delegate.

### 2.3 Why the 7-principle gate is load-bearing

Without the gate, the inversion would be a judgment call ("Tier S/M feels direct, Tier L feels delegated"). With the gate, every op class has a falsifiable test: P1–P7 each pass or fail, and the routing follows. Future op classes (a new git workflow, a new tier) apply the same gate. The gate is the defense against both (a) over-delegation (the PR-#1338 failure mode) and (b) under-delegation (the temptation to abolish manager-git on the back of one incident).

---

## §3. Why `MOAI_BRANCH_GUARD_EXEMPT=1` is the right exemption mechanism

### 3.1 The existing branch-guard is a correct safety net

`internal/hook/branch_guard.go` already encodes the right invariant: a branch-state mutation on the primary checkout is denied unless the actor is trusted. The trust predicate (`isExemptAgent`, lines 144–155) has two branches:

```go
func isExemptAgent(input *HookInput) bool {
    if os.Getenv(branchGuardExemptEnv) == "1" {   // env branch — unconditional
        return true
    }
    // ...
    return input.AgentType == "manager-git"        // identity branch — scoped to manager-git
}
```

The env branch fires FIRST and is unconditional. The identity branch is a secondary, narrower path.

### 3.2 The orchestrator is not an `AgentType`

The orchestrator is the main thread (`claude` / `claude --agent`); it is NOT a subagent and does not carry an `AgentType`. The identity branch (`AgentType == "manager-git"`) is therefore structurally unable to admit orchestrator-direct ops. Adding an "orchestrator" identity branch would be wrong: it would couple the guard to a concept (`AgentType`) the orchestrator does not populate, and it would widen the exemption to a blanket trust-the-orchestrator path.

### 3.3 The env path is the narrowest correct mechanism

The env var `MOAI_BRANCH_GUARD_EXEMPT=1` is set per-invocation (inline: `MOAI_BRANCH_GUARD_EXEMPT=1 git switch -c ...`) or per-sequence. It does NOT persist in settings (which would defeat the guard for every orchestrator Bash call). This is the narrowest blast radius: the exemption is scoped to the specific git-op sequence the orchestrator is performing, and the guard continues to protect every other Bash invocation.

### 3.4 No Go change expected

The env branch already returns true before the identity branch is reached. Orchestrator-direct Tier S/M with `MOAI_BRANCH_GUARD_EXEMPT=1` set is admitted by the existing code. M3 verifies this mechanically (the `TestIsExemptAgent` test with `HookInput{AgentType: ""}` + env=1 returns true via the env branch). REQ-OGR-012 forbids widening the exemption beyond the env predicate (no new identity branch).

---

## §4. Why the Late-Branch 4-Phase body stays verbatim

The Late-Branch Invocation Pattern (`manager-git.md` § Late-Branch, Phases A–D) is the canonical Tier L closure. The relaxation does NOT touch it, for three reasons:

1. **Tier L is the retention surface.** Late-Branch is explicitly retained for manager-git (REQ-OGR-008). The 7-principle gate passes (§2.2) — there is no justification for changing the body.
2. **Byte-identity is a regression guard.** AC-OGR-007 enforces byte-identical Late-Branch body between the local file and the template mirror. The regression test mechanically detects any drift.
3. **Symptom 7 (doctrine bloat compensating for delegation ambiguity) is resolved by the relaxation, not by rewriting Late-Branch.** Once Tier S/M no longer routes through manager-git, the Late-Branch body is under less doctrinal pressure — it can remain the dense Tier L closure it was designed to be.

---

## §5. Why the regression test reproduces the incident class, not the literal incident

AC-OGR-006 requires a regression test that reproduces the PR-#1338 incident *class* (multi-worktree state op: isolated-snapshot cross-swap vs full-context correct restore), not the literal incident (which involved a specific concurrent session's worktree state that cannot be replayed deterministically).

The test shape:
- **Setup**: construct a multi-worktree fixture in `t.TempDir()` (primary + 1 foreign worktree).
- **Isolated-snapshot simulation**: an "actor" that reads only a cached snapshot (taken at setup time, before the foreign worktree's branch changes) and decides a branch-state op. The snapshot is stale relative to the live state.
- **Orchestrator-direct simulation**: an "actor" that reads live state (`git rev-parse` + `git worktree list` at decision time).
- **Assertion**: the isolated-snapshot actor's decision would cross-swap (reproduces the failure mode); the orchestrator-direct actor's decision restores correctly (resolves it).

This shape is deterministic, fast (no real git server, no real concurrent process), and covers the signature that matters: state-snapshot staleness under delegation vs live-state observation under direct action.

---

## §6. Alternatives considered and rejected

(Also documented in `plan.md` §G Anti-Patterns; restated here as design alternatives for the Tier L design artifact.)

### 6.1 Abolish manager-git entirely

**Rejected.** The 7-principle gate passes for Tier L / Late-Branch (§2.2). Abolishing manager-git would leave no canonical owner for release PRs and multi-step merges, and would contradict the "keep spawning the same kind of worker" Anthropic criterion (the Late-Branch 4-Phase is exactly that recurring worker). PRs #1319/#1321/#1326/#1330 are positive evidence that manager-git performs standard push+PR well for the retained class.

### 6.2 Add an "orchestrator" identity branch to `isExemptAgent`

**Rejected.** The orchestrator is not an `AgentType`. Coupling the guard to a concept the orchestrator does not populate is structurally wrong. The env path is the narrowest correct mechanism (§3.3).

### 6.3 Set `MOAI_BRANCH_GUARD_EXEMPT=1` persistently in settings

**Rejected.** This would exempt every orchestrator Bash call from the branch-state guard, defeating the guard entirely. The exemption is per-invocation (§3.3).

### 6.4 Rewrite the Late-Branch Phases A–D for the new world

**Rejected.** Out of scope (§F) and unjustified (§4). The 4-Phase closure is Tier L only and stays verbatim.

### 6.5 Bundle skills/hooks cleanup into this SPEC

**Rejected.** The user's "순차 분할" (phased split) directive explicitly separates the manager-git relaxation from skills/hooks cleanup. Bundling would violate scope discipline (CLAUDE.md §7 Rule 2, Agent Core Behaviors #5) and complicate the regression audit.

### 6.6 Flip the distributed `Workflow.BranchGuard.Enabled` default to `true`

**Rejected.** Out of scope (§F). The relaxation uses the env exemption path; the opt-in gate default is a separate decision owned by `SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001`.

---

## §7. Cross-references

- **Evidence base**: `.moai/reports/agent-skill-hook-redesign-evidence-20260804.md` (§0 paradox, §1.3 classification, §4 change-surface, §5 principles + quantitative grounding).
- **Principle doctrine**: `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 4 (reachability ≠ justification) — the symmetric that the prior doctrine violated.
- **Branch-guard SSOT**: `.claude/rules/moai/workflow/main-checkout-branch-guard.md` + `internal/hook/branch_guard.go`.
- **Precedent SPECs**: `SPEC-WORKTREE-BRANCH-GUARD-001`, `SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001`, `SPEC-V3R6-AGENT-TEAM-REBUILD-001` (REQ-ATR-020 — the doctrine being carved out).
- **Companion artifacts**: `spec.md` (requirements), `plan.md` (milestones), `acceptance.md` (Given-When-Then), `research.md` (evidence + URLs), `progress.md` (§E lifecycle skeleton).
