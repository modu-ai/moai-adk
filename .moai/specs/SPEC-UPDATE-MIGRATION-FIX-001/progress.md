# SPEC-UPDATE-MIGRATION-FIX-001 — Progress

SPEC ID: SPEC-UPDATE-MIGRATION-FIX-001
Card: t1578
Status: draft (plan phase)
Tier: M

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored: spec.md, plan.md, acceptance.md, research.md,
  decision-index.md, progress.md (this file) — all under
  `.moai/specs/SPEC-UPDATE-MIGRATION-FIX-001/`.
- Authored on tree: 2aab5f797 (branch WT-update-migration-fixes, card
  worktree `.moai/worktrees/t1578`); the six artifacts landed as lane
  commit f569be5d8 — the plan-artifact baseline is f569be5d8, and the
  anchor-path delta 2aab5f797..f569be5d8 is empty (acceptance.md EV-7),
  so all plan-phase measurements carry over. BASELINE RE-CUT NOTICE: the
  worktree fast-forwarded from 81786284e to 2aab5f797 (main absorbed, 155
  commits) mid-research. All research anchors re-verified on 2aab5f797
  (acceptance.md EV-6): the deny-migration map, its test file, and
  internal/template/deployer_mode.go are UNCHANGED between the two bases;
  internal/cli/update.go carries a 57-line change (reconciliation preview
  rename, migrateProjectCommonAssets removal-arm relocation into the sync
  flow gated on userAssetsInstalled, participation step added to the skip
  block) — none of it touches the SPEC's conclusions; the skip-path block
  was re-read byte-identical and still carries NO integrity probe.
- Plan-phase measurements recorded: acceptance.md evidence ledger EV-1
  through EV-7 (all read-only, all run in this worktree; EV-6 is the
  re-verification batch on the re-cut baseline, including the
  normalization guard re-run — `ok ... 2.866s`, exit 0 on 2aab5f797; EV-7
  extends the anchors to f569be5d8 and records the EV-5 grep's actual
  execution). Authoring-discipline correction recorded at EV-5: one
  ledger row was initially written before its command ran; it has been
  re-measured and the correction is stated in the entry itself.
- Scope decisions: card item (3) out-of-scope with rationale (spec.md C.1);
  card item (4) in-scope as M3 (spec.md C.2); both recorded in
  decision-index.md (Q3 evidence-needed; Q1/Q2 implementation-level
  defaults applied at plan close).
- Known-issue classification: K1 (card P1) and K2 (card P2) verified as
  already repaired upstream on this tree — K1 by t1569 M2 / PR #1792
  (measured: normalization guard green), K2 by SPEC-USER-ASSET-INSTALL-001
  mechanism retirement. The only new implementation is the version-match
  integrity probe (K3). Branch contingency: if run-phase M1-b measures a
  live empty-directory producer, M2 escalates to repair per spec.md R1.
- Plan status: audit-ready.

## §E.2 Run-phase Evidence

_pending run-phase_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_

## §F Phase 4 Mode Selection

Plan→run Kickoff decision record (autonomous transition, auto-semantics §9.1):

```text
decision record: decided_by=lane-27 orchestrator (card t1578) evidence_refs=.moai/reports/t1578/plan-audit-iter3.md (verdict PASS, overall 0.94 >= Tier M 0.80, blocking 0, convergence pass, codex pass, audited_sha 817b5b86f) + plan-audit-iter2.md (rcpt-5bb5f28c03f4602d0879cd46) ladder_path=autonomous-kickoff §9.1 (verdict PASS + score >= threshold + artifact-hash unchanged on the Go ComputeHash subject set + no blocker open)
```

Mode selection inputs: tier=M; scope≈6 files (probe + guard tests + install.go handler); domain count=1 (Go CLI internal/cli + internal/userassets); file language mix=Go + SPEC artifacts; concurrency benefit=LOW (coding-heavy); Agent Teams prereqs=not requested.

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | no | semantic multi-file change, not a typo fix |
| serial | **yes** | coding-heavy Go implementation (Anthropic coding-task caveat) |
| fanout | no | research-heavy work only; single domain, single writer |
| sweep | no | not a mechanical-uniform ≥30-file transform |

Decision: **serial** (one manager-develop spawn, milestones M1→M4 in sequence).

Justification: the implementation is coding-heavy Go work in one subsystem family (update path + userassets installer); per Anthropic's coding-task parallelism caveat the sequential single-agent path is the safe default, and the write contract is one writer per tree (this worktree). sweep/fanout offer no concurrency benefit here; direct is below the semantic-change bar.

Boundary cases: none — all four mode criteria resolved unambiguously.

