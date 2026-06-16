# Progress — SPEC-CC2178-MODEL-POLICY-REPAIR-001

> Run-phase progress tracker. Owned by manager-develop (run-phase + Mx close) and manager-docs (sync-phase §E.4).

## §E.1 Plan-phase Audit-Ready Signal

- **Artifact set**: spec.md (12-field frontmatter, `era: V3R6`, version 0.2.0 iter-2 → 0.3.0 in-progress), plan.md (M1-M6), acceptance.md (14 ACs).
- **SPEC ID regex**: `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` → PASS.
- **Frontmatter schema**: all 12 canonical fields present; `status: in-progress` (transitioned at M1 commit per Status Transition Ownership Matrix — manager-develop owns `draft → in-progress` on M1).
- **GEARS compliance**: REQs use Ubiquitous / When / While / Where / shall not — 0 residual `IF/THEN`.
- **plan-auditor verdict**: PASS-WITH-DEBT 0.86 (iter-2, threshold 0.80 met). User granted Implementation Kickoff Approval.
- **Spec lint**: `moai spec lint` path-prefixed → exit 0, 1 WARNING (`StatusGitConsistency` — expected for plan-phase draft).

## §E.2 Run-phase Evidence

### M1 — Research: `[1m]` re-verification + Default-model key + task-triage decision

**Status**: COMPLETE (2026-06-16)

**Deliverable**: `.moai/specs/SPEC-CC2178-MODEL-POLICY-REPAIR-001/research.md` (full research note).

**M1 verdict — `[1m]` constraint (AC-MPR-011, REQ-MPR-013/014/015)**: **STILL-ACTIVE (conservative)**

Evidence (fetched 2026-06-16 via GitHub REST API + canonical CC CHANGELOG):

| Issue | State | Labels | Verdict contribution |
|-------|-------|--------|----------------------|
| #45847 | closed (2026-04-13) | `duplicate` | closed-as-duplicate; no explicit "fixed" |
| #51060 | closed (2026-05-26) | `bug, area:model, area:agents, stale` | closed-stale; no changelog fix for spawn-time entitlement mismatch |
| #36670 | **OPEN** (updated 2026-06-02) | `bug, has repro, area:agents, stale` | Team-mode `[1m]` inheritance confirmed UNFIXED at CC 2.1.178 |

CC 2.1.170-2.1.178 CHANGELOG `[1m]`-class entries: 2.1.172 stuck-session recovery + suffix normalization; 2.1.173 Fable-5 suffix; 2.1.174 background-session env-inheritance. **None fix the subagent-spawn entitlement-inheritance root cause.**

Conservative default per acceptance.md EC-01: still-active → EX-01 holds → per-agent pinning forbidden → Default-model routing (`availableModels` + `enforceAvailableModels`) is the only confirmed-safe lever.

**M1 task 4 — Default-model JSON key (AC-MPR-003, REQ-MPR-003)**: confirmed `model` (top-level) + `availableModels` + `enforceAvailableModels`. CC 2.1.175 changelog verbatim: `enforceAvailableModels` constrains the Default model. Verification command pinned: Python JSON-parse of rendered settings (avoids multi-match ambiguity of raw `grep '"model"'`).

**M1 effort-map deferral (REQ-MPR-012, AC-MPR-010 part d)**: PRUNE + RECONCILE only; full retirement DEFERRED to `SPEC-CC2178-EFFORT-MAP-RETIREMENT-001` (2 production callers: `initializer.go:181`, `update.go:2661`).

**M1 task-triage decision (REQ-MPR-016/017, AC-MPR-012)**: DEFERRED to `SPEC-CC2178-TASK-TRIAGE-001` (rationale: 3-axis scope already substantial; integration point absent; metrics unvalidated).

### M2-M6

_<pending — M2 phantom-map cleanup + M3 ResolveCycleType + M4 Default-model lever + M5 doctrine + M6 verification>_

## §E.3 Run-phase Audit-Ready Signal

_<pending — populated at M6 completion>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §E.5 Mx-phase Audit-Ready Signal

_<pending mx-phase>_
