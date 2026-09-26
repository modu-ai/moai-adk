---
id: SPEC-CI-VERDICT-PRODUCER-001
title: "CI verdict producer — implementation plan"
version: "0.1.0"
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
tier: M
---

# plan.md — SPEC-CI-VERDICT-PRODUCER-001

## §A — Context

- Card t1268, worktree `.claude/worktrees/t1268`, branch `WT-ci-verdict-producer` @ develop `bf3d5144f`. cycle_type: tdd (quality.yaml `development_mode: tdd`).
- Debt: `internal/escalation/operational.go:25-28` + `:325` hardcode the class-5 CI limb as never-observed; no producer of CI verdicts exists (t1235 verdict.md, AC-AE-012(c) UNVERIFIED).
- Artifact set (Tier M + research): spec.md, plan.md, acceptance.md, progress.md, research.md under `.moai/specs/SPEC-CI-VERDICT-PRODUCER-001/`.
- Existing infrastructure to REUSE: `internal/escalation` `contradiction()` / `freshEvidence()` (operational.go:363-410); `convergenceFile` parse shape (operational.go:311-317); `internal/verify` `SnapshotDir`/`Snapshot`/`Load` (store.go:16, schema.go:55); house `gh` shelling (`internal/cli/doctor.go:444-452` LookPath + warn-and-continue); atomic temp+rename write discipline (`internal/verify/store.go` `Save`).

## §B — Known Issues (relevant subset)

- **B1 cross-platform**: no syscall surface is introduced; verify with `GOOS=windows GOARCH=amd64 go build ./...`.
- **B4 frontmatter**: canonical 12 fields, no snake_case aliases (done in spec.md).
- **B5 CI tiers**: pre-existing `internal/cli` and `internal/escalation` baselines are measured in §C before work; only deltas are this card's.
- **B8 tree hygiene**: writes go to the SPEC dir + `internal/` + tests only; `.moai/state/` is runtime data — tests use `t.TempDir()`, never the real tree.
- **B11 boundary**: any blocker returns as a structured report; no AskUserQuestion.
- **Test isolation**: all new tests fabricate gh output via an injected runner (interface or PATH-fixture); no test touches the network.

## §C — Pre-flight

```bash
git branch --show-current && git rev-parse --short HEAD
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run ./internal/escalation/... ./internal/cli/... --timeout=2m 2>&1 | tail -5
go test ./internal/escalation/... -run TestContradictoryEvidenceTrips   # limb (c) is not exercised today — RED-now for the limb
grep -n "notObservedCIVerdict" internal/escalation/operational.go       # the two sites this card removes/rewires
git diff --name-only "$(git merge-base develop HEAD)"..HEAD -- go.mod    # MUST stay empty through close (merge-base form: a develop absorption must not drag sibling cards' go.mod changes into this guard — t543 precedent)
```

RED-now note: limb (c) has no test today (the existing `TestContradictoryEvidenceTrips` covers (a),(b),(d),(e) only). The trip test written in M2 must be observed failing before M2's implementation flips it — it is red for the right stated reason: `classContradictoryEvidence()` never reads `.moai/state/ci-verdicts/`.

## §D — Constraints

- PRESERVE: `internal/escalation` behavior for classes other than 5's CI limb; the existing limb (a),(b),(d),(e) test expectations byte-for-byte (the "ci verdict" not-observed string survives for the no-record case, wording may change only with the test updated in the same commit); SPEC-AUTONOMY-ESCALATION-001 files (read-only).
- FORBIDDEN: `go.mod` edits; new gate machinery; invoking `escalation.Checkpoint` from the producer; `--no-verify`; sweeping `git add`.
- The producer never fabricates a verdict on gh failure (REQ-CV-003): message + exit 0 + no file.

## §E — Self-Verification (planned §E items for run phase)

1. AC PASS/FAIL matrix over AC-CV-001..008 with verbatim `go test` outputs (`./internal/escalation/...`, `./internal/cli/...` change-scoped runs).
2. Cross-platform build output (`go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`).
3. Coverage for touched packages ≥ 85%.
4. `go.mod` untouched: `git diff --name-only "$(git merge-base develop HEAD)"..HEAD -- go.mod` → empty.
5. Lint status (new-vs-baseline separated).
6. RED evidence for the new trip test captured before GREEN.

## §F — Milestones (priority-ordered; decisions first, mechanics last)

- **M1 — Verdict record type + store (data-model decision; most likely to change under review).** Shared record struct (five fields per REQ-CV-004) + atomic save/load, placed so both the CLI producer and the escalation detector import it (sibling-package or exported-from-escalation — implementer's call, one location, one import direction: cli → escalation, never the reverse). Offline `--from-json` path lands here (REQ-CV-002) so M1 is testable with zero network surface. Tests: schema round-trip, idempotent rewrite, atomicity under `t.TempDir()`.
- **M2 — Detector CI limb (behavior decision).** Extend `classContradictoryEvidence()`: drop the unconditional `notObservedCIVerdict` seed; glob `.moai/state/ci-verdicts/*.json` under `r.root`; head-match against `r.headSHA()`; local-pass lookup via the verify snapshot store (REQ-CV-006); trip via the existing `contradiction()` path (REQ-CV-007); not-observed labeling per REQ-CV-008. Extend `TestContradictoryEvidenceTrips` with limb (c): same-head local-pass + CI-failure fabricated records → exactly one record; different-head CI failure → none + listed; CI success → none, not listed; no record → existing limb-(e) assertions unchanged (REQ-CV-009 idempotent-retrip case included). RED-first per §C.
- **M3 — Producer CLI verb (user-facing surface).** New cobra command wired in `internal/cli/root.go`, in the `newStateCmd()`-house style; fetches the CI conclusion for `--head <sha>` via the house `gh` pattern with the gh runner injectable for tests (fabricated output — AC-CV-002 runs without network); writes via M1's store.
- **M4 — Degradation + polish (mechanical; lowest change likelihood).** gh-absent/unauth/failure path: one-line clear message, exit 0, no file (REQ-CV-003); cross-platform build; lint deltas; coverage measurement; §E evidence capture.

Ordering rationale: M1 and M2 carry the two decisions a reviewer is most likely to move (record schema; same-head semantics) and land first; M3's verb surface is stable once M1's store exists; M4 is mechanical.

## §G — Anti-Patterns

- Inventing a second local-pass source beside the verify snapshot store (REQ-CV-006 forbids it).
- Producer silently writing a "neutral" record on gh failure — that is a fabricated verdict, not a degradation.
- Re-parsing `gh`'s native JSON inside `internal/escalation` — the detector knows only the five-field schema (Q3).
- Cross-tree mirroring of `.moai/state/ci-verdicts/` — rejected in spec.md §E.
- Touching SPEC-AUTONOMY-ESCALATION-001 files.

## §H — Cross-References

- spec.md §G (the four settled design questions, with rejected alternatives), acceptance.md (AC-CV-001..008), research.md (anchors + gh-degradation precedents).
- t1235 verdict (primary checkout): `/Users/goos/MoAI/moai-adk-go/.moai/reports/t1235/verdict.md`.
