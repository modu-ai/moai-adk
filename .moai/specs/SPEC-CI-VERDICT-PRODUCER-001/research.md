---
id: SPEC-CI-VERDICT-PRODUCER-001
title: "CI verdict producer — research"
version: "0.1.0"
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
tier: M
---

# research.md — SPEC-CI-VERDICT-PRODUCER-001

All anchors re-verified in this worktree at develop `bf3d5144f`, 2026-09-26.

## P1 — The dead limb, precisely

- `internal/escalation/operational.go:25-28`: `const notObservedCIVerdict = "ci verdict (no recorded CI verdict producer)"` — comment names the orchestrator ruling (plan.md §H Q5) that produced it.
- `internal/escalation/operational.go:325`: `notObs := []string{notObservedCIVerdict}` — unconditionally seeded into every `not_observed` list `classContradictoryEvidence()` returns. No read of any CI-verdict evidence exists.
- Original criterion: `.moai/specs/SPEC-AUTONOMY-ESCALATION-001/acceptance.md:94-96` (AC-AE-012 limbs (a)-(e), "A1 plan-audit 통과본으로 재확인"); requirement text at spec.md:201-205 (REQ-AE-010).
- Consumer-side test today: `internal/escalation/operational_m5_test.go:162` `TestContradictoryEvidenceTrips` — covers (a) flag-true, (b) second-model, (d) flag-null, (e) no-CI-verdict. Limb (c) is absent with a comment citing the missing producer. Extending THIS test is the mechanical re-judgement form (AC-CV-008).

## P2 — Class-5 evidence parse precedent (format decision grounds)

- `convergenceFile` (operational.go:311-317): plain `encoding/json` struct, `filepath.Glob(filepath.Join(r.root, ".moai", "state", "audit-multi", "*.json"))` at :326, `os.ReadFile` + `json.Unmarshal`, unreadable file → appended to `not_observed` and skipped. The CI limb should read `.moai/state/ci-verdicts/*.json` through the identical shape — the detector must not learn `gh`'s native schema.

## P3 — Trip machinery to reuse (do not reinvent)

- `contradiction(source, data, pair, observation, ref)` (operational.go:394-410): `Fingerprint(ClassContradictoryEvidence, source, pair)` + `freshEvidence` content-hash gate + `writeRecord`. The CI limb's source = the verdict filename; pair distinguishes the CI limb (e.g. `local-pass|ci-failure`).
- `freshEvidence` (operational.go:363-371): consumed-evidence key = class+fingerprint+file+SHA256Hex(data). This is what makes an idempotent producer re-run safe (REQ-CV-009) and satisfies REQ-AE-020/AC-AE-023 unchanged.
- Head resolution: `r.headSHA()` (detector.go:408-413, `ReadHead(r.root)`); records already carry `head_sha` (record.go:75).

## P4 — Local-pass source: verify snapshots (reuse judgement)

- `internal/verify/store.go:16` `SnapshotDir = ".moai/state/verify/snapshots"`; `SnapshotPath` names files `<head[:12]>-<digest>.json` (Windows-safe colon stripping, store.go:20-34); `Load` treats key mismatch as absent (collision defense, store.go:52-55); `Save` is atomic temp+rename (store.go:66+).
- `Snapshot` (schema.go:55-59) carries `Key` ("head:digest" shape — cut on `':'` for the full head) and `Checks []CheckEntry`; `CheckEntry` (schema.go:44-53) carries `ExitCode` and `Conditions.TestsPass *bool`.
- Chosen local-pass predicate (REQ-CV-006): a snapshot whose key's head portion equals the checkpoint head AND ≥1 check entry with `ExitCode == 0`. Exit-code-only keeps the detector decoupled from `Conditions` optionality; the attributable diff-check system already guarantees the snapshot is head-pinned.
- Rejected: deriving local pass from the escalation event stream (no test-run event reaches the commit checkpoint reliably); a new `.moai/state/local-pass/` writer (a second producer — the exact defect this card exists to avoid).

## P5 — gh CLI house patterns (no new dependency)

- `internal/cli/doctor.go:444-452`: `exec.LookPath("gh")` → warn-and-continue on absence — the fail-open template for REQ-CV-003.
- Other shelling precedents: `internal/cli/todo_pr.go`, `internal/cli/branch_protection.go`, `internal/cli/session_worktree_prmerge.go:364` (`gh pr view --json state`). All injectable-output tests use fixtures, never network.
- `go.mod` gains nothing: `os/exec` + `encoding/json` are stdlib. Cross-platform safe (no new syscall surface; temp+rename already proven on Windows by verify `Save`).

## P6 — Placement and import direction

- The record type needs one home importable by both `internal/cli` (producer) and `internal/escalation` (detector). `internal/cli` already imports widely; `internal/escalation` currently imports nothing from `internal/cli`. Least-risk: export the type + store helpers from `internal/escalation` (sibling of `convergenceFile`) OR a minimal `internal/civerdict` package both import — implementer's call at M1, one direction only: cli → (escalation | civerdict).

## P7 — Visibility: primary vs worktree (Q2 grounds)

- Records land in the tree where the producer runs. In this repo's git-flow the canonical producer site is the lead's tree at the pushed develop head; the detector reads its own `r.root`. The same-head pin (REQ-CV-007/008) makes cross-tree staleness inert rather than wrong: a worktree at an older head fails the head match and lists the mismatch under `not_observed`. No mirroring layer is proposed (spec.md §E).

## P8 — Out-of-card debts (not absorbed)

- t1235 Q2: nothing invokes `escalation.Checkpoint` in production — pinned decoupled by REQ-CV-005.
- t1235 audit findings (e.g. F3 records agent-writable): that card's ledger.
- Debt record: `/Users/goos/MoAI/moai-adk-go/.moai/reports/t1235/verdict.md` (primary checkout; read-only reference).
