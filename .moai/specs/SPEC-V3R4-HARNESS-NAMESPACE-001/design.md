# SPEC-V3R4-HARNESS-NAMESPACE-001 — Design

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.2.0   | 2026-05-16 | manager-spec | plan-auditor REVISE 0.69 round-1 fix. §2.1 hierarchy adds `README.md`/`main.md` canonical reserved names (D-006); §4.2 dependency graph corrected to show HARNESS-003 → HARNESS-002 sibling dep + reframed independence semantics (D-005); §5.1/§5.2/§5.3 closeout architecture rewritten for 2-commit divergence + 4-option escalation (D-004); §3.6 verb stability remark unchanged but stays aligned with corrected regex (D-001). |
| 0.1.0   | 2026-05-16 | manager-spec | Initial draft. Three diagrams: (1) `.moai/harness/*` namespace hierarchy tree, (2) `/moai harness` 4-verb state machine with AskUserQuestion gating, (3) V3R4-HARNESS family lifecycle dependency matrix. Zero implementation diagrams (pure governance — no code architecture). |

---

## 1. Overview

This design document materializes the namespace governance contracts declared in `spec.md` §5 (REQ-HRN-NS-001 through REQ-HRN-NS-010) as three concrete artifacts:

1. **Namespace hierarchy diagram** — the canonical `.moai/harness/*` tree referenced by REQ-HRN-NS-004 and AC-HRN-NS-004.
2. **Verb state machine** — the `/moai harness` 4-verb surface (REQ-HRN-NS-003), AskUserQuestion gating on `apply` (inherited from V3R4-HARNESS-001 REQ-HRN-FND-004), and rollback/disable safety semantics.
3. **Lifecycle dependency matrix** — the eight-SPEC family (SPEC-V3R4-HARNESS-{001..008}) with explicit blocking vs non-blocking relationships, satisfying REQ-HRN-NS-002 lifecycle independence guarantee.

This SPEC is pure governance — no functional code architecture is in scope. All diagrams below document EXISTING contracts inherited from V3R4-HARNESS-001 or document FUTURE constraints that downstream SPECs MUST satisfy.

---

## 2. Namespace Hierarchy Diagram

### 2.1 Project-Local Harness State Root: `.moai/harness/`

```
.moai/
├── harness/                                  ← namespace root (REQ-HRN-NS-004)
│   ├── README.md                             ← Directory orientation (optional)
│   ├── main.md                               ← Harness instance metadata
│   │                                            (V3R4-HARNESS-001 run-phase artifact)
│   ├── usage-log.jsonl                       ← PostToolUse observation log (REQ-HRN-FND-010)
│   │                                            JSONL, append-only, no rotation in this SPEC scope.
│   │                                            Retention floor: 7-day rolling window (REQ-HRN-NS-010).
│   │
│   ├── proposals/                            ← Tier-X candidate proposal artifacts
│   │   └── <proposal-id>.md                  ← One file per Tier-3/Tier-4 candidate
│   │
│   └── learning-history/
│       ├── snapshots/
│       │   └── <YYYY-MM-DD>/                 ← Pre-application byte-identical snapshot
│       │       ├── manifest.json             ← Target paths + content hashes (REQ-HRN-FND-007)
│       │       └── <original-relative-path>  ← Byte-identical copy of each file modified
│       │
│       ├── applied/
│       │   └── <YYYY-MM-DD>-<proposal-id>.md ← Applied evolution record
│       │
│       ├── frozen-guard-violations.jsonl     ← L1 Frozen Guard rejection audit log
│       │                                        (REQ-HRN-FND-006 + REQ-HRN-FND-014)
│       │                                        JSONL, ISO-8601 timestamp + target path + rationale.
│       │
│       └── tier-promotions.jsonl             ← Tier 1→2→3→4 promotion event log
│                                                Feeds REQ-HRN-FND-016 success-metric exposure.
│
├── specs/                                    ← SPEC documents (out of harness namespace, governed separately)
│   ├── SPEC-V3R4-HARNESS-001/                ← Foundation (completed)
│   ├── SPEC-V3R4-HARNESS-002/                ← Observer expansion (TBD)
│   ├── SPEC-V3R4-HARNESS-003/                ← Embedding-cluster classifier (plan only, OPEN PR #923)
│   ├── SPEC-V3R4-HARNESS-NAMESPACE-001/      ← THIS SPEC (governance)
│   └── SPEC-V3R3-HARNESS-*/                  ← Superseded by V3R4-HARNESS-001
│
└── reports/
    └── governance/                           ← Governance verification audit artifacts (this SPEC)
        ├── SPEC-V3R4-HARNESS-NAMESPACE-001-wave1-<YYYY-MM-DD>.md
        └── SPEC-V3R4-HARNESS-NAMESPACE-001-wave2-<YYYY-MM-DD>.md
```

### 2.2 Project-Generated Specialist Artifacts (outside `.moai/harness/`)

Inherited from V3R3-HARNESS-001 REQ-HARNESS-008 inviolate boundary (preserved unchanged by V3R4-HARNESS-001):

```
.claude/
├── agents/
│   └── my-harness/                           ← Project-generated specialist agents
│       └── <agent-name>.md                   ← USER-AREA, inviolate to MoAI updates
│
└── skills/
    └── my-harness-*/                         ← Project-generated specialist skills
        └── SKILL.md                          ← USER-AREA, inviolate to MoAI updates
```

### 2.3 FROZEN Zone (Inherited from V3R4-HARNESS-001 REQ-HRN-FND-006)

The L1 Frozen Guard rejects any harness write attempt to:

```
.claude/agents/moai/**                        ← MoAI core agents (FROZEN)
.claude/skills/moai-*/**                      ← MoAI core skills (FROZEN)
.claude/rules/moai/**                         ← MoAI rules (FROZEN, includes design constitution)
.moai/project/brand/**                        ← Brand context (FROZEN per constitution §3.1)
```

This SPEC does NOT modify the FROZEN zone definition; it documents the existing contract for downstream-SPEC author awareness.

### 2.4 Hierarchy Compliance Rule (REQ-HRN-NS-009)

```
+--------------------------------------------------------------+
| Project-local harness state                                  |
+--------------------------------------------------------------+
| ALLOWED (canonical reserved names per REQ-HRN-NS-004)        |
|   .moai/harness/README.md                                    |
|   .moai/harness/main.md                                      |
|   .moai/harness/usage-log.jsonl                              |
|   .moai/harness/proposals/**                                 |
|   .moai/harness/learning-history/snapshots/**                |
|   .moai/harness/learning-history/applied/**                  |
|   .moai/harness/learning-history/frozen-guard-violations.jsonl|
|   .moai/harness/learning-history/tier-promotions.jsonl       |
+--------------------------------------------------------------+
| REJECTED (REQ-HRN-NS-009 — plan-auditor rejects offending SPEC)|
|   .moai/harness/anything-not-listed-above/                   |
|   .moai/experimental-harness/                                |
|   <any path outside .moai/harness/ that stores harness state>|
+--------------------------------------------------------------+
```

---

## 3. `/moai harness` Verb State Machine

### 3.1 Verb Surface (REQ-HRN-NS-003, inherited from V3R4-HARNESS-001 REQ-HRN-FND-003)

```
+----------------------------------------------------------+
| /moai harness <verb>                                     |
+----------------------------------------------------------+
|                                                          |
|  status        — read-only telemetry (no gate)           |
|  apply         — Tier-4 evolution (AskUserQuestion GATE) |
|  rollback DATE — restore snapshot (no gate, idempotent)  |
|  disable       — set learning.enabled = false (no gate)  |
|                                                          |
+----------------------------------------------------------+
```

### 3.2 State Diagram — `apply` verb (AskUserQuestion gating)

```
                 +-------------------+
                 | User invokes      |
                 | /moai harness     |
                 | apply             |
                 +---------+---------+
                           |
                           v
                 +-------------------+
                 | Orchestrator loads|
                 | proposal queue    |
                 | from .moai/       |
                 | harness/proposals/|
                 +---------+---------+
                           |
              proposal found OR queue empty?
                  /                  \
        empty                       found
           |                           |
           v                           v
   +---------------+         +-----------------+
   | "No pending   |         | L1 Frozen Guard |
   | proposals"    |         | check on every  |
   | exit          |         | target path     |
   +---------------+         +--------+--------+
                                      |
                          all allowed?  / \  any blocked?
                                       /   \
                                  YES        NO
                                   |          |
                                   v          v
                          +-----------+  +---------------+
                          | L2 Canary |  | Log to        |
                          | Check     |  | frozen-guard- |
                          | (score    |  | violations    |
                          | drop ≤    |  | .jsonl,       |
                          | 0.10?)    |  | abort         |
                          +-----+-----+  +---------------+
                                |
                          PASS  / \  FAIL
                              /     \
                             v       v
                    +----------+   +---------+
                    | L3       |   | Abort,  |
                    | Contradic|   | log     |
                    | -tion    |   +---------+
                    | Detector |
                    +----+-----+
                         |
                    NO   / \  YES
                       /     \
                      v       v
              +----------+   +---------+
              | L4 Rate  |   | Surface |
              | Limiter  |   | conflict|
              | (≤1/7d   |   | to user,|
              | window?) |   | abort   |
              +----+-----+   +---------+
                   |
              PASS / \ THROTTLED
                 /     \
                v       v
        +----------+  +----------+
        | L5 Human |  | Defer,   |
        | Oversight|  | wait for |
        | (Ask     |  | window   |
        | UserQ)   |  | reset    |
        +----+-----+  +----------+
             |
       Apply / Modify / Defer / Reject
             |
        ┌────┼────┬────┬────┐
        v    v    v    v    v
    APPLY MODIFY DEFER REJECT
        |    |    |    |
        |    |    v    v
        |    | re-queue NO-OP
        |    v
        | user edits proposal
        v
+---------------------+
| Create snapshot at  |
| learning-history/   |
| snapshots/<date>/   |
| (REQ-HRN-FND-007)   |
+----------+----------+
           |
           v
+---------------------+
| Apply changes       |
| atomically          |
+----------+----------+
           |
           v
+---------------------+
| Append entry to     |
| learning-history/   |
| applied/<date>-     |
| <proposal-id>.md    |
+---------------------+
```

[HARD] AskUserQuestion at L5 is **orchestrator-only** (V3R4-HARNESS-001 REQ-HRN-FND-004 + REQ-HRN-FND-015). Subagents return blocker reports; never invoke AskUserQuestion directly.

### 3.3 State Diagram — `rollback DATE` verb (no AskUserQuestion gate)

```
+-------------------+
| User invokes      |
| /moai harness     |
| rollback <DATE>   |
+---------+---------+
          |
          v
+---------------------+
| Look up snapshot at |
| learning-history/   |
| snapshots/<DATE>/   |
+----------+----------+
           |
     found / \ not found
         /     \
        v       v
+----------+  +----------+
| Restore  |  | Print    |
| each file|  | error,   |
| from     |  | exit 1   |
| manifest |  +----------+
| atomically|
+----+-----+
     |
     v
+-----------+
| Print     |
| restored  |
| file list,|
| exit 0    |
+-----------+
```

Rollback is idempotent — running it twice on the same date is a no-op on the second invocation (files already match snapshot).

### 3.4 State Diagram — `disable` verb (no AskUserQuestion gate)

```
+-------------------+
| User invokes      |
| /moai harness     |
| disable           |
+---------+---------+
          |
          v
+---------------------+
| Set                 |
| learning.enabled =  |
| false in            |
| .moai/config/       |
| sections/harness.   |
| yaml                |
+----------+----------+
           |
           v
+---------------------+
| PostToolUse observer|
| no-ops on next      |
| invocation          |
| (REQ-HRN-FND-009)   |
| usage-log.jsonl is  |
| NOT deleted         |
+---------------------+
```

### 3.5 State Diagram — `status` verb (read-only)

```
+-------------------+
| User invokes      |
| /moai harness     |
| status            |
+---------+---------+
          |
          v
+---------------------+
| Read:               |
|   usage-log.jsonl   |
|   tier-promotions   |
|     .jsonl          |
|   frozen-guard-     |
|     violations      |
|     .jsonl          |
+----------+----------+
           |
           v
+---------------------+
| Compute:            |
|   Tier-4 weekly app |
|   count             |
|   Tier-4 reach rate |
|   (REQ-HRN-FND-016) |
+----------+----------+
           |
           v
+---------------------+
| Print structured    |
| telemetry to stdout |
| No file modification|
+---------------------+
```

### 3.6 Verb Surface Stability (REQ-HRN-NS-008)

```
+-----------------------------------+
| Allowed verbs (governance-stable) |
+-----------------------------------+
| status, apply, rollback, disable  |
+-----------------------------------+

+--------------------------------------------------+
| Forbidden silent additions (REQ-HRN-NS-008)      |
+--------------------------------------------------+
| Any 5th verb (e.g., "evolve", "audit", "export") |
| introduced by a downstream SPEC WITHOUT a new    |
| governance SPEC amending REQ-HRN-NS-003 →        |
| plan-auditor REJECTS the offending SPEC.         |
+--------------------------------------------------+
```

---

## 4. Lifecycle Dependency Matrix

### 4.1 V3R4-HARNESS Family Members (as of plan-phase 2026-05-16)

| SPEC ID                              | Status     | Plan PR | Run PR | Sync PR | Notes |
|--------------------------------------|------------|---------|--------|---------|-------|
| `SPEC-V3R4-HARNESS-001`              | completed  | merged  | #910   | merged  | Foundation. Main commit `bb80ea0f4`. |
| `SPEC-V3R4-HARNESS-002`              | TBD        | TBD     | TBD    | TBD     | Spec files present, status not declared. |
| `SPEC-V3R4-HARNESS-003`              | TBD (plan) | #923 OPEN | TBD  | TBD     | 9334 LOC embedding-cluster classifier plan. |
| `SPEC-V3R4-HARNESS-004` (anticipated)| not authored | -     | -      | -       | Reflexion self-critique loop. |
| `SPEC-V3R4-HARNESS-005` (anticipated)| not authored | -     | -      | -       | Principle-based scoring. |
| `SPEC-V3R4-HARNESS-006` (anticipated)| not authored | -     | -      | -       | Multi-objective effectiveness measurement. |
| `SPEC-V3R4-HARNESS-007` (anticipated)| not authored | -     | -      | -       | Voyager-style skill library. |
| `SPEC-V3R4-HARNESS-008` (anticipated)| not authored | -     | -      | -       | Cross-project lesson federation (privacy-sensitive). |
| `SPEC-V3R4-HARNESS-NAMESPACE-001`    | draft      | this PR | future | future  | Governance — THIS SPEC. |

### 4.2 Dependency Graph (REQ-HRN-NS-002 — acyclic + declared-dependency discipline)

Current state as observed in main HEAD `82880cd49`:

```
                  +---------------------------+
                  | SPEC-V3R4-HARNESS-001     |
                  | (foundation, completed)   |
                  +----+----+----+----+-------+
                       |    |    |    |
                       v    v    v    v
              +--------+ +--+ +--+   +-------+
              | -002   | |..| |..|   | NSPC- |
              | (TBD)  | |  | |  |   | 001   |
              +---+----+ +--+ +--+   | (this)|
                  ^                  +-------+
                  | sibling-dep
                  | (registered blocker)
                  |
              +---+----+
              | -003   |  depends_on: [-001, -002]
              | (plan) |  Cannot merge run/sync until -002
              +--------+  reaches status: implemented or completed.
                          Foundation dep (-001) ALWAYS permitted.

(Anticipated -004 through -008 not yet on disk; each MUST follow
 REQ-HRN-NS-001 through REQ-HRN-NS-009 per this governance SPEC,
 declaring dependencies explicitly — sibling deps are allowed as
 registered blockers per REQ-HRN-NS-002.)
```

**Rules** (REQ-HRN-NS-002 reframed semantics):

1. Foundation dependency on `SPEC-V3R4-HARNESS-001` is ALWAYS permitted and is the canonical baseline.
2. `SPEC-V3R4-HARNESS-NAMESPACE-001` (this SPEC) depends on `SPEC-V3R4-HARNESS-001` only.
3. **Sibling dependencies ARE permitted** (NOT prohibited) — they act as **registered merge blockers**. Example: `SPEC-V3R4-HARNESS-003` declares `dependencies: [SPEC-V3R4-HARNESS-001, SPEC-V3R4-HARNESS-002]`, which means -003's run PR and sync PR are blocked until -002 reaches `status: implemented` or `status: completed`. This is by-design discipline, not a violation.
4. Downstream SPECs `-004` through `-008` MUST follow the namespace governance from THIS SPEC. They MAY declare sibling deps (e.g., `-005 → -004` if scoping warrants), and those sibling deps become registered blockers per rule 3.
5. The dependency graph MUST remain acyclic (no `A → B → A` chains). Enforced at plan-auditor.
6. Undeclared blockers are PROHIBITED — every blocker MUST be a `dependencies:` entry. Ad-hoc cross-PR coordination is rejected.
7. Cyclic dependencies (e.g., `-002 → -003 → -002`) are REJECTED at plan-auditor.

### 4.3 Lifecycle Examples (declared-dependency discipline)

**Example A: Foundation precedes; foundation-only deps are parallel**

```
Time →
1. -001 plan merged
2. -001 run merged
3. -001 sync merged → status: completed
4. -002 (deps: [-001]) plan, NAMESPACE-001 (deps: [-001]) plan open simultaneously
5. -002 / NAMESPACE-001 merge in any order (no sibling dependency between them)
6. -003 (deps: [-001, -002]) plan can merge anytime (plan-phase has no blocker)
7. -003 run PR opens — BLOCKED until -002 reaches status: implemented (registered blocker)
```

**Example B: NAMESPACE-001 ships before downstream SPECs (RECOMMENDED)**

```
Time →
1. -001 completed
2. NAMESPACE-001 plan/run/sync merged → status: completed
3. -002 lifecycle continues with explicit governance contract reference
4. -003 plan merged (foundation+sibling deps declared)
5. -002 run merged → status: implemented
6. -003 run PR unblocks (sibling dep satisfied) → can merge
7. -004 through -008 plan-phase MUST cite NAMESPACE-001 REQs as binding constraints
```

**Example C: NAMESPACE-001 ships AFTER -003 plan begins (PERMITTED)**

```
Time →
1. -001 completed
2. -003 plan PR opens (deps: [-001, -002])
3. NAMESPACE-001 plan PR opens in parallel (deps: [-001])
4. -003 plan PR merges (plan-auditor verifies governance-compatible structure)
5. NAMESPACE-001 plan PR merges (no retroactive impact on -003)
6. -003 run PR remains BLOCKED on -002 (sibling dep is a registered blocker per REQ-HRN-NS-002)
7. -002 implements → -003 run unblocks
```

Lifecycle discipline ENSURES that NAMESPACE-001 is not a release blocker for `-002`, `-003`, or any other family member, AND that declared sibling deps create predictable, auditable blockers instead of ad-hoc coordination.

### 4.4 Supersede Chain (Inherited from V3R4-HARNESS-001)

```
+----------------------------+      +-----------------------------+
| SPEC-V3R3-HARNESS-001      |      | SPEC-V3R3-HARNESS-LEARNING- |
| (Meta-Harness Skill)       |      | 001 (5-Layer Safety + 4-Tier|
|                            |      | Evolution Ladder)           |
+--------------+-------------+      +--------------+--------------+
               |                                   |
               +----------------+----------+-------+
                                |          |
                                v          v
                  +-----------------------------+
                  | SPEC-V3R3-PROJECT-HARNESS-  |
                  | 001 (16Q Interview)         |
                  +--------------+--------------+
                                 |
                                 v
                  +-----------------------------+
                  | superseded by               |
                  | SPEC-V3R4-HARNESS-001       |
                  | (foundation, completed)     |
                  +-----------------------------+
```

[HARD] This SPEC does NOT modify the supersede chain. It documents the existing relationship for governance audit purposes (AC-HRN-NS-005).

---

## 5. Closeout Architecture for PR #908

### 5.1 Current State (plan-phase reference — 2-commit divergence CONFIRMED)

```
+--------------------------------------------------+
| GitHub state (plan-phase 2026-05-16 verified)    |
+--------------------------------------------------+
| PR #908                                          |
|   title: feat(cmd): /moai harness 슬래시 명령    |
|          thin-wrapper 추가                        |
|   state: OPEN                                    |
|   created: 2026-05-13                            |
|   head: feat/cmd-harness-slash-wrapper           |
|   head SHA: a41d6d139c8c769bf395a25a055d59c14e180191 |
|   merged: false                                  |
+--------------------------------------------------+

+--------------------------------------------------+
| PR #908 branch commits (newest first)            |
+--------------------------------------------------+
|  a41d6d139  docs(audit): harness 서브시스템      |
|             end-to-end 진단 보고서                |
|             ← UNABSORBED into main               |
|  452aa638f  feat(cmd): /moai harness slash wrap  |
|             ← ABSORBED into main via PR #910     |
|               commit bb80ea0f4                   |
+--------------------------------------------------+

+--------------------------------------------------+
| Main branch state                                |
+--------------------------------------------------+
| commit 82880cd49 (HEAD as of 2026-05-16)         |
|   ↑                                              |
|   ... (intermediate commits)                     |
|   ↑                                              |
| commit bb80ea0f4 (PR #910 squash merge)          |
|   ← absorbed PR #908 commit 452aa638f content    |
|     (Commit 1); a41d6d139 (Commit 2) NOT here    |
+--------------------------------------------------+
```

### 5.2 Target State (post run-phase Wave 2, depends on EC-001 option chosen)

```
Option 1 (Recommended): rollback-tip-then-close
+--------------------------------------------------+
| PR #908                                          |
|   state: CLOSED                                  |
|   branch HEAD reset to 452aa638f then deleted    |
|   audit doc (a41d6d139): discarded               |
+--------------------------------------------------+

Option 2: cherry-pick-to-new-PR-then-close
+--------------------------------------------------+
| PR #908                                          |
|   state: CLOSED                                  |
|   branch: deleted                                |
| New PR #XXX                                      |
|   contains: a41d6d139 cherry-pick                |
|   base: main, branch: docs/harness-audit-*       |
|   audit doc: preserved in proper scope           |
+--------------------------------------------------+

Option 3: abandon-new-commits
+--------------------------------------------------+
| PR #908                                          |
|   state: CLOSED                                  |
|   branch: deleted                                |
|   audit doc (a41d6d139): discarded (reflog only) |
+--------------------------------------------------+

Option 4: leave-open
+--------------------------------------------------+
| PR #908                                          |
|   state: OPEN (unchanged)                        |
|   branch: feat/cmd-harness-slash-wrapper exists  |
|   audit doc: still on branch                     |
|   deferral rationale recorded in Wave 2 audit    |
+--------------------------------------------------+
```

### 5.3 Closeout Decision Tree (Wave 2 T-Wave2-001 + T-Wave2-002, EC-001 EXPECTED)

```
                  +------------------+
                  | orchestrator     |
                  | runs T-Wave2-001 |
                  +--------+---------+
                           |
                           v
              gh pr view 908 --json headRefOid
                           |
                           v
              [[ "$head" == 452aa638f* ]] ?
                       /        \
                    YES          NO  ← EXPECTED at plan-phase
                  (rare)          (HEAD = a41d6d139...)
                     |            |
                     v            v
            +----------------+   +-------------------------+
            | EC_001_PATH    |   | EC_001_PATH=true        |
            | =false         |   |                         |
            |                |   | orchestrator invokes    |
            | simple abandon |   | AskUserQuestion with    |
            | (close+delete) |   | 4 options per EC-001:   |
            |                |   |  1. rollback-tip (권장) |
            |                |   |  2. cherry-pick-to-new  |
            |                |   |  3. abandon-new         |
            |                |   |  4. leave-open          |
            +-------+--------+   +-----------+-------------+
                    |                        |
                    v                        v
                +----------------+   +-------------------------+
                | manager-git    |   | manager-git executes    |
                | closes PR      |   | user-selected option    |
                | (no comment    |   | (rollback / cherry-pick |
                |  needed)       |   |  / close / leave-open)  |
                +-------+--------+   +-----------+-------------+
                        \                       /
                         \                     /
                          v                   v
                       +-------------------------+
                       | T-Wave2-005             |
                       | audit artifact persisted|
                       | (records option chosen) |
                       +-------------------------+
```

---

## 6. Out of Scope

This design document explicitly EXCLUDES:

1. Implementation diagrams (no code architecture — this is a governance SPEC).
2. Sequence diagrams for `/moai harness` verb invocation (owned by V3R4-HARNESS-001 plan.md).
3. Class / module structure of `internal/cli/harness.go` (deprecation marker — no changes).
4. UI/UX mockups for AskUserQuestion gating (inherited from V3R4-HARNESS-001 contract).
5. Data flow diagrams for `usage-log.jsonl` → tier classifier → proposals (owned by future SPEC-V3R4-HARNESS-003 + -004).
6. Performance / latency diagrams (no SLOs in governance SPEC).
7. Diagrams for cross-project federation (deferred to SPEC-V3R4-HARNESS-008).

---

## 7. References

- spec.md §5 (REQs), §6 (Acceptance Coverage Map), §10 (Glossary).
- plan.md §2 (Wave 1), §3 (Wave 2), §5 (Technical Approach).
- acceptance.md §2 (AC scenarios), §3 (EC cases).
- tasks.md — run-phase task list.
- SPEC-V3R4-HARNESS-001 plan.md §2 (Architecture diagrams) — visual reference for inherited contracts.
- `.claude/rules/moai/design/constitution.md` §2 (FROZEN zones), §5 (5-Layer Safety) — inherited diagrams (do NOT duplicate).
- `.claude/skills/moai/workflows/harness.md` — verb implementation surface (DO NOT MODIFY in this SPEC).
